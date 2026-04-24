// File: internal/npcterritory/manager.go
// Project: Terminal Velocity
// Description: NPC system ownership — which NPC faction currently
//   controls a given star system. Seeded from the static
//   StandardNPCFactions.CoreSystems data at server start, then
//   mutated as wars resolve and systems flip between factions.
//   Separate from internal/territory/, which tracks *player-founded*
//   guild claims keyed by uuid.UUID — NPC control uses string slugs
//   and is driven by the faction-war lifecycle, not player actions
//   against a repo.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package npcterritory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

var log = logger.WithComponent("NPCTerritory")

// Sentinel errors — callers `errors.Is` against these to branch on
// specific failure modes rather than string-matching.
var (
	ErrUnknownSystem  = errors.New("system not tracked")
	ErrUnknownFaction = errors.New("faction not tracked")
)

// NewsBus is the minimum news.Manager surface this package needs.
// Kept as an interface so npcterritory doesn't pull the full news
// package, and so tests can record emissions without a real manager.
type NewsBus interface {
	AddArticle(*models.NewsArticle)
}

// OwnershipPersister is the write-through hook for persisting
// ownership changes. Called on every flip so a server restart
// preserves the political map. Production wires a simple adapter
// around database.NPCTerritoryRepository.UpsertOwnership; tests
// pass nil for no persistence or an in-memory recorder for
// assertions.
//
// Errors are logged but do not roll back the in-memory flip — a
// transient DB hiccup shouldn't undo a war resolution, it should
// just mean the next restart recovers to the pre-flip state and a
// later war can re-flip it.
type OwnershipPersister func(ctx context.Context, systemName, factionID string) error

// FlipRecord captures one ownership change. Returned from
// ResolveWarTerritory so callers (the factionwar manager, primarily)
// can surface the news, update caches, and trigger downstream
// systems. Separate from news emission because the manager is
// decoupled from the news package at this layer — the caller wires
// the news bus.
type FlipRecord struct {
	SystemName string
	FromID     string
	FromName   string
	ToID       string
	ToName     string
	At         time.Time
}

// Manager tracks system → NPC-faction ownership. Thread-safe for
// concurrent reads during TUI rendering alongside war-resolution
// writes via TickWars.
//
// Ownership is keyed by lowercased system name so lookups match the
// case-insensitive convention established in the factionwar
// war-zone index. Faction metadata (display name, short name) is
// cached alongside the ID so the news + banner paths don't have to
// re-query the static faction list.
type Manager struct {
	mu sync.RWMutex

	// ownerBySystem: lowercased system name → NPCFaction.ID slug.
	ownerBySystem map[string]string

	// originalCase preserves the system name's casing at first
	// sight so the TUI can display "Alpha Centauri" rather than
	// "alpha centauri" after a lookup round-trip.
	originalCase map[string]string

	// factionName / factionShortName: ID → display name / tag, so
	// news emission and UI code don't need to re-scan the static
	// faction list. Populated at Seed() time; the list is
	// small (~8) so a full copy is fine.
	factionName      map[string]string
	factionShortName map[string]string

	// P5D-2: player-contribution tracking. Nested map keyed by
	// lowercased system name → factionID → points accumulated.
	// Cleared for a system when its ownership flips, so a freshly-
	// captured system starts fresh — past contributions to the
	// losing side don't transfer with the flag.
	contributions map[string]map[string]int64

	newsBus NewsBus

	// P5D-3: write-through DB persistence. nil when no
	// persistence is wired (tests; standalone servers with
	// no DB). Called with the ORIGINAL-CASE system name so the
	// row's user-facing display stays stable across restarts.
	persister OwnershipPersister

	now func() time.Time // seam for deterministic tests
}

// NewManager constructs a Manager. newsBus may be nil — callers that
// don't need territory-flip headlines (tests, one-off migrations)
// can pass nil and the flip-news emission becomes a no-op.
func NewManager(newsBus NewsBus) *Manager {
	return &Manager{
		ownerBySystem:    make(map[string]string),
		originalCase:     make(map[string]string),
		factionName:      make(map[string]string),
		factionShortName: make(map[string]string),
		contributions:    make(map[string]map[string]int64),
		newsBus:          newsBus,
		now:              time.Now,
	}
}

// SetPersister wires the write-through persistence hook. Call at
// server startup, before the first flip — it's not synchronized
// against concurrent TransferSystem/ResolveWarTerritory.
func (m *Manager) SetPersister(p OwnershipPersister) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persister = p
}

// RestoreOwnership applies a list of persisted overrides on top of
// the Seed()-populated defaults. Call this AFTER Seed() so that a
// war-captured system recorded in the DB overrides its static
// CoreSystems entry.
//
// Unknown factions are logged and skipped (can happen if the static
// faction list was trimmed between server versions — we don't want
// a bad DB row to crash startup). The system_name original casing
// is whatever the DB had at write time; if that system isn't in
// the seed list, it's added with that casing.
func (m *Manager) RestoreOwnership(overrides map[string]string) {
	if m == nil || len(overrides) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for sysName, factionID := range overrides {
		key := normalize(sysName)
		if key == "" {
			continue
		}
		if _, known := m.factionName[factionID]; !known {
			log.Warn("RestoreOwnership: skipping unknown faction %q for system %q", factionID, sysName)
			continue
		}
		m.ownerBySystem[key] = factionID
		// Persist the original casing from the DB row (preserves
		// whatever the seed had, or the casing from a later flip).
		if _, kept := m.originalCase[key]; !kept {
			m.originalCase[key] = sysName
		}
	}
}

// Seed populates the manager from the static NPC faction list,
// using each faction's CoreSystems as the initial owned set.
// Influence systems are *not* seeded as owned — influence means
// "some presence," not control. Later wars will transfer core
// systems between factions.
//
// Idempotent: re-seeding overwrites prior state (ownership AND
// contributions). Useful for hot-reload and test resets.
func (m *Manager) Seed(factions []models.NPCFaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ownerBySystem = make(map[string]string, 64)
	m.originalCase = make(map[string]string, 64)
	m.factionName = make(map[string]string, len(factions))
	m.factionShortName = make(map[string]string, len(factions))
	m.contributions = make(map[string]map[string]int64)

	for _, f := range factions {
		m.factionName[f.ID] = f.Name
		m.factionShortName[f.ID] = f.ShortName

		for _, sys := range f.CoreSystems {
			key := normalize(sys)
			if key == "" {
				continue
			}
			// Last faction in the list wins if two factions
			// claim the same core system in the fixture data —
			// fine for tests; production data is unambiguous.
			m.ownerBySystem[key] = f.ID
			if _, kept := m.originalCase[key]; !kept {
				m.originalCase[key] = sys
			}
		}
	}
}

// GetOwner returns the NPCFaction.ID that currently controls the
// named system. Returns "", ErrUnknownSystem if the system isn't
// tracked (wasn't in anyone's CoreSystems at seed time).
// Case-insensitive lookup.
func (m *Manager) GetOwner(systemName string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	owner, ok := m.ownerBySystem[normalize(systemName)]
	if !ok {
		return "", ErrUnknownSystem
	}
	return owner, nil
}

// GetOwnerName returns the owning faction's full display name, or
// empty string if the system or faction isn't tracked. Convenience
// for TUI code that doesn't want to stitch GetOwner + factionName
// together manually.
func (m *Manager) GetOwnerName(systemName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ownerID, ok := m.ownerBySystem[normalize(systemName)]
	if !ok {
		return ""
	}
	return m.factionName[ownerID]
}

// GetOwnerShortName returns the owning faction's tag (e.g., "UEF").
// Used by the space-view territory banner where full names would
// overflow a one-line strip.
func (m *Manager) GetOwnerShortName(systemName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ownerID, ok := m.ownerBySystem[normalize(systemName)]
	if !ok {
		return ""
	}
	return m.factionShortName[ownerID]
}

// GetFactionSystems returns the system names currently owned by a
// faction, in alphabetical order for stable iteration. Returns nil
// if the faction isn't tracked or owns nothing (note: empty slice vs
// nil is semantically distinct — nil means "we don't know this
// faction"; empty means "known faction, no holdings").
func (m *Manager) GetFactionSystems(factionID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, tracked := m.factionName[factionID]; !tracked {
		return nil
	}
	out := []string{}
	for key, owner := range m.ownerBySystem {
		if owner == factionID {
			if orig, ok := m.originalCase[key]; ok {
				out = append(out, orig)
			} else {
				out = append(out, key)
			}
		}
	}
	sort.Strings(out)
	return out
}

// TransferSystem moves a named system from its current owner to
// newFactionID and (if newsBus is configured and the flip actually
// happened) emits a political news article. Returns the FlipRecord
// so callers can log, audit, or chain further side-effects.
//
// Returns nil FlipRecord + nil error when newFactionID already
// controls the system (no-op — not an error; the factionwar
// integration may call this unconditionally on every war-zone
// system and we don't want to spam the news feed on a trivial
// re-transfer).
func (m *Manager) TransferSystem(systemName, newFactionID string) (*FlipRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, known := m.factionName[newFactionID]; !known {
		return nil, fmt.Errorf("%w: %s", ErrUnknownFaction, newFactionID)
	}
	key := normalize(systemName)
	if key == "" {
		return nil, ErrUnknownSystem
	}
	prev, tracked := m.ownerBySystem[key]
	if !tracked {
		return nil, ErrUnknownSystem
	}
	if prev == newFactionID {
		return nil, nil // already-owned, idempotent no-op
	}

	rec := &FlipRecord{
		SystemName: m.originalCase[key],
		FromID:     prev,
		FromName:   m.factionName[prev],
		ToID:       newFactionID,
		ToName:     m.factionName[newFactionID],
		At:         m.now(),
	}
	m.ownerBySystem[key] = newFactionID
	// Flip resets per-system contributions: the captured system
	// starts fresh under its new owner. Leaving old contributions
	// intact would let a faction "bank" effort from an earlier
	// conflict across ownership changes, which makes contested
	// systems frustrating to re-contest.
	delete(m.contributions, key)

	m.persistOwnershipLocked(rec.SystemName, newFactionID)
	m.emitFlipNews(rec)
	return rec, nil
}

// ResolveWarTerritory is the factionwar-driven capture path: called
// when a war resolves, it flips every war-zone system currently
// owned by the loser over to the winner. Systems owned by third
// parties (or the winner already) are untouched — a UEF vs Crimson
// war doesn't redistribute ROM-owned border systems even if they
// appear in the war-zone list because both belligerents had
// influence there.
//
// Returns the ordered list of flips so the caller can append them
// to a combat/news log. Empty slice = war had no territorial
// impact (resolution still happens, just no systems changed hands).
func (m *Manager) ResolveWarTerritory(zoneSystems []string, loserID, winnerID string) []*FlipRecord {
	if m == nil || loserID == winnerID {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, known := m.factionName[winnerID]; !known {
		return nil
	}
	var flips []*FlipRecord
	for _, sys := range zoneSystems {
		key := normalize(sys)
		if key == "" {
			continue
		}
		if m.ownerBySystem[key] != loserID {
			continue // not owned by loser → not captured
		}
		rec := &FlipRecord{
			SystemName: m.originalCase[key],
			FromID:     loserID,
			FromName:   m.factionName[loserID],
			ToID:       winnerID,
			ToName:     m.factionName[winnerID],
			At:         m.now(),
		}
		m.ownerBySystem[key] = winnerID
		delete(m.contributions, key) // fresh slate under new owner
		m.persistOwnershipLocked(rec.SystemName, winnerID)
		m.emitFlipNews(rec)
		flips = append(flips, rec)
	}
	return flips
}

// AllOwnership returns a snapshot of every tracked system → owner
// mapping, using original-case system names. Primarily for admin
// tools and territory-map rendering. O(n) copy; pin the result
// rather than calling this in a hot loop.
func (m *Manager) AllOwnership() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.ownerBySystem))
	for key, owner := range m.ownerBySystem {
		name := key
		if orig, ok := m.originalCase[key]; ok {
			name = orig
		}
		out[name] = owner
	}
	return out
}

// ============================================================================
// P5D-2 player contribution tracking
// ============================================================================

// AddContribution credits `amount` points toward a faction's control
// of the named system. Intended callers: combat (player kill in a
// war zone credits the opposite side), mission handlers (completion
// of faction war mission), and admin tools.
//
// Amount is allowed to be negative (future use: faction-friendly
// act sabotages your account with other belligerents). Zero and
// non-tracked-system inputs are silently ignored — this is not a
// validation path, it's a hot-path hook called from every combat
// resolution, so we don't want it failing the caller on an unknown
// system name.
func (m *Manager) AddContribution(systemName, factionID string, amount int64) {
	if m == nil || amount == 0 {
		return
	}
	key := normalize(systemName)
	if key == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, tracked := m.ownerBySystem[key]; !tracked {
		return
	}
	if _, known := m.factionName[factionID]; !known {
		return
	}
	if m.contributions[key] == nil {
		m.contributions[key] = make(map[string]int64, 2)
	}
	m.contributions[key][factionID] += amount
}

// ContributionFor returns the points a faction has accumulated in a
// specific system. Zero for systems/factions with no activity —
// unknown inputs aren't distinguished from "tracked but zero"
// because the caller can always check GetOwner first if they need
// to disambiguate.
func (m *Manager) ContributionFor(systemName, factionID string) int64 {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	bySystem := m.contributions[normalize(systemName)]
	if bySystem == nil {
		return 0
	}
	return bySystem[factionID]
}

// ContributionLeader aggregates contributions across a set of
// systems (typically a war's WarZoneSystems list) for each of the
// two candidate factions and reports who's in the lead. Return
// values:
//
//   - leaderID: factionID with the highest total, or "" on tie /
//     no-contributions
//   - margin: leader's total minus runner-up's total (non-negative).
//     A margin of 0 indicates a tie.
//
// Used by factionwar.TickWars auto-resolution: when a war expires
// and there's a clear contribution leader, they win; otherwise
// the RNG coin flip decides.
func (m *Manager) ContributionLeader(systems []string, aggressorID, defenderID string) (leaderID string, margin int64) {
	if m == nil {
		return "", 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var aggressorTotal, defenderTotal int64
	for _, sys := range systems {
		bySystem := m.contributions[normalize(sys)]
		if bySystem == nil {
			continue
		}
		aggressorTotal += bySystem[aggressorID]
		defenderTotal += bySystem[defenderID]
	}

	switch {
	case aggressorTotal > defenderTotal:
		return aggressorID, aggressorTotal - defenderTotal
	case defenderTotal > aggressorTotal:
		return defenderID, defenderTotal - aggressorTotal
	default:
		return "", 0
	}
}

// SystemContributions returns a snapshot of the contribution map
// for a single system: factionID → points. Empty map when the
// system is tracked but has no activity, nil when the system is
// unknown. Primarily for TUI rendering of a contested-system
// status panel.
func (m *Manager) SystemContributions(systemName string) map[string]int64 {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := normalize(systemName)
	if _, tracked := m.ownerBySystem[key]; !tracked {
		return nil
	}
	bySystem := m.contributions[key]
	out := make(map[string]int64, len(bySystem))
	for k, v := range bySystem {
		out[k] = v
	}
	return out
}

// ============================================================================
// Internal helpers
// ============================================================================

func normalize(systemName string) string {
	return strings.ToLower(strings.TrimSpace(systemName))
}

// persistOwnershipLocked fires the persister hook if configured.
// Called while m.mu is held in write mode (by the caller); errors
// are logged, not returned, because a transient DB blip shouldn't
// roll back an in-memory flip — the worst case is the next
// restart recovers to the pre-flip state, which is still consistent.
//
// 5-second timeout keeps the write bounded even on a degraded DB;
// beyond that the flip is orphaned in memory (acceptable given
// wars are infrequent and this handler runs under lock).
func (m *Manager) persistOwnershipLocked(systemName, factionID string) {
	if m.persister == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.persister(ctx, systemName, factionID); err != nil {
		log.Warn("persist ownership %s → %s: %v", systemName, factionID, err)
	}
}

func (m *Manager) emitFlipNews(rec *FlipRecord) {
	if m.newsBus == nil || rec == nil {
		return
	}
	fromShort := m.factionShortName[rec.FromID]
	toShort := m.factionShortName[rec.ToID]
	if fromShort == "" {
		fromShort = rec.FromName
	}
	if toShort == "" {
		toShort = rec.ToName
	}
	headline := fmt.Sprintf("%s captures %s from %s", toShort, rec.SystemName, fromShort)
	body := fmt.Sprintf("%s forces have wrested control of %s from %s. New patrol patterns and station allegiances are expected within the week.",
		rec.ToName, rec.SystemName, rec.FromName)
	m.newsBus.AddArticle(models.NewNewsArticle(
		models.NewsCategoryPolitical,
		models.NewsPriorityHigh,
		headline,
		body,
	))
}
