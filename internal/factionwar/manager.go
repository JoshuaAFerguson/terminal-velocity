// File: internal/factionwar/manager.go
// Project: Terminal Velocity
// Description: In-memory manager for FactionWars between NPC
//   factions. Provides declaration, resolution, cease-fire, and
//   query operations; emits news articles on state transitions via
//   an injected NewsBus. P5C-1 backend; TUI surface deferred to
//   P5C-2.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package factionwar

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Sentinel errors — callers can errors.Is against these to branch on
// specific failure modes without string-matching.
var (
	ErrSameFaction   = errors.New("faction cannot declare war on itself")
	ErrAlreadyAtWar  = errors.New("these factions are already at war")
	ErrWarNotFound   = errors.New("war not found")
	ErrWarNotActive  = errors.New("war is not in active state")
	ErrInvalidWinner = errors.New("winner must be one of the belligerents")
	ErrNilFaction    = errors.New("faction cannot be nil")
)

// NewsBus is the minimum news.Manager surface this package needs.
// Kept as an interface (rather than importing news directly) to
// dodge a circular-import risk: news → factionwar → news would land
// eventually once news wants to react to wars. The production wire-up
// passes *news.Manager at construction time.
type NewsBus interface {
	AddArticle(*models.NewsArticle)
}

// TerritoryHook is the callback factionwar invokes when a war
// resolves so the territory layer can flip ownership of war-zone
// systems from loser to winner. A callback (vs. interface) keeps
// factionwar decoupled from any particular territory implementation
// — tests pass an in-line closure, production passes a one-line
// adapter that calls (*npcterritory.Manager).ResolveWarTerritory.
//
// The hook is nil-safe on the call site (ResolveWar / TickWars
// check before invoking); passing nil means "no territory
// integration" which is the default for tests.
type TerritoryHook func(zoneSystems []string, loserID, winnerID string)

// Manager holds all faction-war state. All mutating methods take
// the write lock; queries take the read lock. Thread-safe for
// concurrent use by SSH session goroutines.
type Manager struct {
	mu sync.RWMutex

	// Primary storage: war ID → war. pointer-valued so lookups
	// return aliased entries that match the one at wars[id] — saves
	// a second map lookup in the common "get and mutate" path.
	wars map[uuid.UUID]*models.FactionWar

	// Active-war index keyed by direction-agnostic pair ("a|b" with
	// alphabetical ordering). Lets IsAtWar/GetWarBetween avoid a
	// full scan of every historical war.
	activeByPair map[string]*models.FactionWar

	// System-level war-zone index: lowercased system name → set of
	// active war IDs covering it. Multiple wars can flag the same
	// system when two factions both fight a third in the same
	// border space.
	warZones map[string]map[uuid.UUID]struct{}

	newsBus NewsBus
	now     func() time.Time // seam for deterministic tests

	// P5C-4 lifecycle automation. config/rng are set at
	// construction to defaults; SetLifecycleConfig/RNG at server
	// startup override them. Tick-callers all go through TickWars,
	// which holds the write lock, so no additional synchronization
	// is needed on these fields once set.
	config LifecycleConfig
	rng    LifecycleRNG

	// P5D-1: territory integration hook. Invoked from ResolveWar /
	// CeaseFire with the zone list and belligerent IDs so the
	// territory layer can flip system ownership. nil when no
	// integration is wired (tests default; production sets via
	// SetTerritoryHook at startup).
	territoryHook TerritoryHook
}

// NewManager constructs a Manager with optional news integration.
// newsBus may be nil (useful for tests that don't care about news
// side effects). The now func defaults to time.Now but can be
// overridden via a constructor option if needed later.
func NewManager(newsBus NewsBus) *Manager {
	return &Manager{
		wars:         make(map[uuid.UUID]*models.FactionWar),
		activeByPair: make(map[string]*models.FactionWar),
		warZones:     make(map[string]map[uuid.UUID]struct{}),
		newsBus:      newsBus,
		now:          time.Now,
		config:       DefaultLifecycleConfig(),
		rng:          &defaultLifecycleRNG{},
	}
}

// SetLifecycleConfig overrides the tick-driven lifecycle parameters
// (max war duration, emergent-declaration probability). Intended for
// server startup — not thread-safe vs. running TickWars calls.
func (m *Manager) SetLifecycleConfig(c LifecycleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

// SetLifecycleRNG swaps the random source used by TickWars for
// declaration rolls and coin-flip resolutions. Test-only seam;
// production callers should leave the default in place.
func (m *Manager) SetLifecycleRNG(r LifecycleRNG) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rng = r
}

// SetTerritoryHook wires the territory-flip callback invoked on
// war resolution. Pass nil to disable. Not thread-safe against
// concurrent ResolveWar — call at server startup, before the first
// war is declared.
func (m *Manager) SetTerritoryHook(hook TerritoryHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.territoryHook = hook
}

// DeclareWar creates a new active war between two NPC factions and
// emits a critical-priority news article. The war-zone systems are
// snapshotted from the union of both factions' core + influence
// systems at declaration time — later territory changes don't
// rewrite the declared zones.
//
// Returns ErrAlreadyAtWar if these two factions already have an
// active war. Use GetWarBetween first if you want to continue an
// existing war rather than declare a new one.
func (m *Manager) DeclareWar(aggressor, defender *models.NPCFaction, casusBelli string) (*models.FactionWar, error) {
	if aggressor == nil || defender == nil {
		return nil, ErrNilFaction
	}
	if aggressor.ID == defender.ID {
		return nil, ErrSameFaction
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := pairKey(aggressor.ID, defender.ID)
	if existing, ok := m.activeByPair[key]; ok && existing.IsActive() {
		return nil, fmt.Errorf("%w: %s vs %s (war %s)",
			ErrAlreadyAtWar, aggressor.ShortName, defender.ShortName, existing.ID)
	}

	zones := warZoneSystems(aggressor, defender)

	war := &models.FactionWar{
		ID:             uuid.New(),
		AggressorID:    aggressor.ID,
		AggressorName:  aggressor.Name,
		DefenderID:     defender.ID,
		DefenderName:   defender.Name,
		Status:         models.FactionWarActive,
		DeclaredAt:     m.now(),
		WarZoneSystems: zones,
		CasusBelli:     casusBelli,
	}

	m.wars[war.ID] = war
	m.activeByPair[key] = war
	for _, sys := range zones {
		m.addWarZone(sys, war.ID)
	}

	m.emitDeclarationNews(war)
	return war, nil
}

// ResolveWar marks a war as resolved with the given winner. The
// winner's factionID must be one of the belligerents — anything else
// returns ErrInvalidWinner. Resolving a non-active war is
// ErrWarNotActive (ceased and already-resolved wars can't be
// re-resolved).
func (m *Manager) ResolveWar(warID uuid.UUID, winnerFactionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, ok := m.wars[warID]
	if !ok {
		return ErrWarNotFound
	}
	if !war.IsActive() {
		return ErrWarNotActive
	}
	if winnerFactionID != war.AggressorID && winnerFactionID != war.DefenderID {
		return ErrInvalidWinner
	}

	resolvedAt := m.now()
	war.Status = models.FactionWarResolved
	war.ResolvedAt = &resolvedAt
	war.WinnerFactionID = winnerFactionID

	delete(m.activeByPair, pairKey(war.AggressorID, war.DefenderID))
	for _, sys := range war.WarZoneSystems {
		m.removeWarZone(sys, war.ID)
	}

	// P5D-1: flip territorial control from loser to winner. Loser
	// is whichever belligerent isn't the winner. Hook is called
	// while holding the write lock so the territory manager's
	// lock ordering must be {factionwar → territory} to avoid
	// deadlocks with any future cross-lookup.
	m.invokeTerritoryHookLocked(war, winnerFactionID)

	m.emitResolutionNews(war)
	return nil
}

// invokeTerritoryHookLocked fires territoryHook with the loser's
// faction ID (aggressor or defender, whichever isn't the winner).
// Must be called with m.mu held in write mode; no-op when no hook
// is wired.
func (m *Manager) invokeTerritoryHookLocked(war *models.FactionWar, winnerID string) {
	if m.territoryHook == nil {
		return
	}
	loserID := war.AggressorID
	if winnerID == war.AggressorID {
		loserID = war.DefenderID
	}
	m.territoryHook(war.WarZoneSystems, loserID, winnerID)
}

// CeaseFire ends a war with no winner. Shares most of its body with
// ResolveWar — the distinction is semantic (news copy, no winner
// recorded). Kept as a separate method so callers don't have to
// pass a sentinel "ceasefire" value.
func (m *Manager) CeaseFire(warID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, ok := m.wars[warID]
	if !ok {
		return ErrWarNotFound
	}
	if !war.IsActive() {
		return ErrWarNotActive
	}

	resolvedAt := m.now()
	war.Status = models.FactionWarCeased
	war.ResolvedAt = &resolvedAt

	delete(m.activeByPair, pairKey(war.AggressorID, war.DefenderID))
	for _, sys := range war.WarZoneSystems {
		m.removeWarZone(sys, war.ID)
	}

	m.emitCeaseFireNews(war)
	return nil
}

// GetWar returns the war with the given ID, or (nil, false) if not
// found. Returns the live pointer; callers must not mutate.
func (m *Manager) GetWar(warID uuid.UUID) (*models.FactionWar, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.wars[warID]
	return w, ok
}

// GetActiveWars returns a snapshot slice of currently-active wars,
// sorted by declaration time (oldest first) for stable iteration.
// Returned pointers aliases internal state; don't mutate.
func (m *Manager) GetActiveWars() []*models.FactionWar {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*models.FactionWar, 0, len(m.activeByPair))
	for _, w := range m.activeByPair {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DeclaredAt.Before(out[j].DeclaredAt)
	})
	return out
}

// GetAllWars returns every war — active, resolved, and ceased —
// sorted newest-declared first. For the war-history screen / admin
// audits. Pointers alias internal state; don't mutate.
func (m *Manager) GetAllWars() []*models.FactionWar {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*models.FactionWar, 0, len(m.wars))
	for _, w := range m.wars {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DeclaredAt.After(out[j].DeclaredAt)
	})
	return out
}

// IsAtWar returns true if factionID has any active war. Direction-
// agnostic — works for both aggressor and defender sides.
func (m *Manager) IsAtWar(factionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.activeByPair {
		if w.AggressorID == factionID || w.DefenderID == factionID {
			return true
		}
	}
	return false
}

// GetWarBetween returns the active war involving factions a and b,
// or nil if they're at peace. Order of arguments doesn't matter.
func (m *Manager) GetWarBetween(a, b string) *models.FactionWar {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeByPair[pairKey(a, b)]
}

// IsSystemWarZone returns true if the named system (case-insensitive)
// is covered by any active war. Used by combat and mission systems
// to amplify reputation, unlock war missions, etc.
func (m *Manager) IsSystemWarZone(systemName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.warZones[strings.ToLower(systemName)]
	return ok
}

// WarsInSystem returns the set of active wars whose war zones
// include the named system. Used by the territory/combat UI to show
// "Contested: UEF vs Crimson" banners. Empty slice on peaceful
// systems.
func (m *Manager) WarsInSystem(systemName string) []*models.FactionWar {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, ok := m.warZones[strings.ToLower(systemName)]
	if !ok {
		return nil
	}
	out := make([]*models.FactionWar, 0, len(ids))
	for id := range ids {
		if w, ok := m.wars[id]; ok {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DeclaredAt.Before(out[j].DeclaredAt)
	})
	return out
}

// ============================================================================
// P5C-3 gameplay integration
// ============================================================================

// Gameplay multipliers — tuned per docs/FACTION_RELATIONS.md spec
// (§"War Zones" amplified rep, §"War Economy" price spikes).
// Exposed as constants so test assertions don't go stale when these
// are rebalanced.
const (
	// WarZoneReputationMultiplier scales reputation deltas earned or
	// lost toward belligerent factions when the combat happens
	// inside a war zone. 1.5× is chosen so kills feel more
	// impactful without letting a player grind allegiance in a
	// single session.
	WarZoneReputationMultiplier = 1.5

	// WarEconomyMultiplier scales the sell price of war-material
	// commodities (weapons, medical, fuel/ore proxies) inside a
	// war-zone planet's market. Players bringing supplies into hot
	// systems earn a premium; enemies of the belligerents take the
	// hit on import costs.
	WarEconomyMultiplier = 1.40
)

// warMaterialCategories lists commodity categories that spike in
// war economies per the spec. Weapons and medical are direct
// consumables; industrial covers repair parts; ore is the upstream
// feedstock for both.
//
// Kept as a set (map for O(1) lookup) so the check in
// WarEconomyPriceMultiplier stays allocation-free.
var warMaterialCategories = map[string]struct{}{
	"weapons":    {},
	"medical":    {},
	"industrial": {},
	"ore":        {},
}

// IsWarMaterial reports whether a commodity category counts as war
// material for pricing purposes. Exposed so callers (trading UI,
// mission generators, news copy) can share the classification
// without each redefining their own list.
func IsWarMaterial(category string) bool {
	_, ok := warMaterialCategories[category]
	return ok
}

// WarZoneReputationScale returns the reputation multiplier for a
// player's combat action inside a war zone against a belligerent
// faction. Returns 1.0 on any miss:
//   - System is peaceful (no active wars cover it)
//   - factionID is not a belligerent in any war covering this system
//   - manager is nil (callers should use the zero-value path)
//
// The multiplier is applied to absolute reputation deltas, so both
// gains and losses are amplified — fighting for the UEF in UEF-vs-
// Crimson zone gets you more rep with UEF allies, but drops you
// faster with Crimson allies too.
func (m *Manager) WarZoneReputationScale(systemName, factionID string) float64 {
	if m == nil {
		return 1.0
	}
	wars := m.WarsInSystem(systemName)
	for _, w := range wars {
		if w.AggressorID == factionID || w.DefenderID == factionID {
			return WarZoneReputationMultiplier
		}
	}
	return 1.0
}

// WarEconomyPriceMultiplier returns the sell-price multiplier for a
// commodity category at a planet in the given system. War-material
// categories (IsWarMaterial) spike by WarEconomyMultiplier when the
// system is any active war's zone; all other commodities are
// unaffected. Peaceful systems and nil manager return 1.0.
func (m *Manager) WarEconomyPriceMultiplier(systemName, category string) float64 {
	if m == nil {
		return 1.0
	}
	if !IsWarMaterial(category) {
		return 1.0
	}
	if !m.IsSystemWarZone(systemName) {
		return 1.0
	}
	return WarEconomyMultiplier
}

// ============================================================================
// P5C-4 lifecycle automation
// ============================================================================

// LifecycleConfig tunes TickWars behavior. Exposed so the server can
// override defaults (e.g., tighten the cadence for a live event) and
// so tests can set deterministic thresholds.
type LifecycleConfig struct {
	// MaxWarDuration is how long an active war can run before
	// TickWars force-resolves it by coin flip. Defaults to 7 days
	// of real time — faction wars aren't supposed to grind
	// indefinitely, and auto-resolution creates natural story
	// beats ("the UEF-Crimson conflict ended yesterday...").
	MaxWarDuration time.Duration

	// DeclarationProbability is the per-tick chance each pair of
	// mutually-hostile NPC factions rolls for an emergent war
	// declaration. At the default tick cadence (5 minutes) and
	// default probability (0.0005), a single pair has ~50% odds
	// of declaring over ~8 hours of continuous uptime. With 4
	// hostile pairs in StandardNPCFactions, something spins up
	// every few hours of server time on average.
	DeclarationProbability float64

	// EmergentCasusBelli is the copy attached to auto-declared
	// wars. Kept configurable so events / seasons can customize
	// the narrative (e.g., "Piracy along the trade corridor").
	EmergentCasusBelli string
}

// DefaultLifecycleConfig returns sensible starting values for the
// tick-driven war lifecycle. Callers can override individual fields
// before passing to SetLifecycleConfig.
func DefaultLifecycleConfig() LifecycleConfig {
	return LifecycleConfig{
		MaxWarDuration:         7 * 24 * time.Hour,
		DeclarationProbability: 0.0005,
		EmergentCasusBelli:     "Border tensions escalated into open conflict.",
	}
}

// LifecycleRNG is the RNG seam TickWars uses for declaration rolls
// and coin-flip war resolution. Tests inject deterministic
// implementations; production uses defaultLifecycleRNG (wraps
// math/rand's global).
type LifecycleRNG interface {
	Float64() float64
	Intn(n int) int
}

type defaultLifecycleRNG struct{}

func (defaultLifecycleRNG) Float64() float64 { return rand.Float64() }
func (defaultLifecycleRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// tickCandidate is a NPCFaction pair ordered alphabetically by ID
// (matches pairKey) so declaration rolls are deterministic across
// runs with the same RNG seed and faction list.
type tickCandidate struct {
	a, b *models.NPCFaction
}

// TickWars runs one iteration of the war lifecycle: resolves any
// active war that has exceeded MaxWarDuration, then rolls
// declaration probabilities for each pair of mutually-hostile NPC
// factions that aren't already at war. Called from the server's
// tick service (see server.initTick).
//
// factions is the current list of NPC factions — normally
// models.StandardNPCFactions, but plumbed in so tests can use a
// controlled fixture. The function takes the write lock once and
// holds it for the whole operation; TickWars isn't expected to run
// concurrently with itself but is safe against concurrent queries.
func (m *Manager) TickWars(factions []models.NPCFaction) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// First pass: auto-resolve long-running wars. Work on a
	// snapshot because we mutate m.activeByPair during iteration.
	now := m.now()
	var expired []*models.FactionWar
	for _, w := range m.activeByPair {
		if now.Sub(w.DeclaredAt) >= m.config.MaxWarDuration {
			expired = append(expired, w)
		}
	}
	for _, w := range expired {
		// Coin flip for winner — a proper territory-weighted
		// outcome lands in P5D once territory exists.
		winnerID := w.AggressorID
		if m.rng.Intn(2) == 0 {
			winnerID = w.DefenderID
		}
		m.resolveLocked(w, winnerID)
	}

	// Second pass: roll emergent declarations. Build candidate
	// pairs deterministically so a seeded RNG produces the same
	// outcomes across runs.
	candidates := buildHostilePairs(factions)
	for _, pc := range candidates {
		key := pairKey(pc.a.ID, pc.b.ID)
		if _, already := m.activeByPair[key]; already {
			continue
		}
		if m.rng.Float64() >= m.config.DeclarationProbability {
			continue
		}
		// RNG picks which side is aggressor — no lore reason the
		// first-listed enemy is always the instigator.
		aggressor, defender := pc.a, pc.b
		if m.rng.Intn(2) == 1 {
			aggressor, defender = pc.b, pc.a
		}
		m.declareLocked(aggressor, defender, m.config.EmergentCasusBelli)
	}
}

// buildHostilePairs walks the faction list and returns each
// mutually-hostile pair exactly once, with stable ordering so
// seeded RNGs produce deterministic TickWars outputs.
//
// A pair is "mutually hostile" if each faction lists the other in
// its Enemies slice. One-way animosities (e.g., pirates hate UEF
// but UEF's Enemies doesn't name pirates) don't roll for war — the
// feeling has to be reciprocal.
func buildHostilePairs(factions []models.NPCFaction) []tickCandidate {
	// Index enemies sets for O(1) mutual-hostility checks.
	enemiesOf := make(map[string]map[string]struct{}, len(factions))
	byID := make(map[string]*models.NPCFaction, len(factions))
	for i := range factions {
		f := &factions[i]
		byID[f.ID] = f
		set := make(map[string]struct{}, len(f.Enemies))
		for _, e := range f.Enemies {
			set[e] = struct{}{}
		}
		enemiesOf[f.ID] = set
	}

	seen := make(map[string]struct{}, len(factions))
	var out []tickCandidate
	for i := range factions {
		a := &factions[i]
		for _, enemyID := range a.Enemies {
			b, ok := byID[enemyID]
			if !ok {
				continue // unknown faction in Enemies list — skip
			}
			// Reciprocal check.
			if _, mutual := enemiesOf[b.ID][a.ID]; !mutual {
				continue
			}
			key := pairKey(a.ID, b.ID)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			// Canonical (alphabetical) ordering so the RNG sees
			// the same pair regardless of which faction iterated
			// first.
			if a.ID <= b.ID {
				out = append(out, tickCandidate{a: a, b: b})
			} else {
				out = append(out, tickCandidate{a: b, b: a})
			}
		}
	}
	return out
}

// declareLocked is the core of DeclareWar minus the entry checks —
// callers that already hold the write lock and have validated
// inputs (e.g., TickWars picking from hostile-pair candidates) use
// this path to avoid re-checking invariants. Emits the declaration
// news article, just like the public path.
func (m *Manager) declareLocked(aggressor, defender *models.NPCFaction, casusBelli string) {
	zones := warZoneSystems(aggressor, defender)
	war := &models.FactionWar{
		ID:             uuid.New(),
		AggressorID:    aggressor.ID,
		AggressorName:  aggressor.Name,
		DefenderID:     defender.ID,
		DefenderName:   defender.Name,
		Status:         models.FactionWarActive,
		DeclaredAt:     m.now(),
		WarZoneSystems: zones,
		CasusBelli:     casusBelli,
	}
	m.wars[war.ID] = war
	m.activeByPair[pairKey(aggressor.ID, defender.ID)] = war
	for _, sys := range zones {
		m.addWarZone(sys, war.ID)
	}
	m.emitDeclarationNews(war)
}

// resolveLocked is the core of ResolveWar minus the entry checks.
// Called by TickWars' auto-resolution path.
func (m *Manager) resolveLocked(war *models.FactionWar, winnerFactionID string) {
	resolvedAt := m.now()
	war.Status = models.FactionWarResolved
	war.ResolvedAt = &resolvedAt
	war.WinnerFactionID = winnerFactionID

	delete(m.activeByPair, pairKey(war.AggressorID, war.DefenderID))
	for _, sys := range war.WarZoneSystems {
		m.removeWarZone(sys, war.ID)
	}
	m.invokeTerritoryHookLocked(war, winnerFactionID)
	m.emitResolutionNews(war)
}

// ReportIncident is the admin / combat-system hook for incident-
// triggered wars. When the combat system detects a major
// provocation (e.g., a player destroys a patrol flagship of a
// belligerent's ally), it calls this with the offended faction and
// the system context. If the offended faction has any mutually-
// hostile enemies AND the incident intensity rolls above threshold,
// a war is declared between the offended faction and a random
// hostile enemy.
//
// intensity is 0.0..1.0 — higher values make war more likely. Use
// 1.0 for "unambiguously hostile act against a high-profile target".
//
// Returns the declared war (if any) or nil if no declaration
// happened this call. nil factions / nil manager / no enemies → nil.
func (m *Manager) ReportIncident(offended *models.NPCFaction, factions []models.NPCFaction, intensity float64) *models.FactionWar {
	if m == nil || offended == nil {
		return nil
	}
	if intensity <= 0 {
		return nil
	}
	if intensity > 1 {
		intensity = 1
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Pick a mutually-hostile enemy not already at war with us.
	candidates := make([]*models.NPCFaction, 0, len(offended.Enemies))
	byID := make(map[string]*models.NPCFaction, len(factions))
	for i := range factions {
		byID[factions[i].ID] = &factions[i]
	}
	for _, enemyID := range offended.Enemies {
		enemy, ok := byID[enemyID]
		if !ok {
			continue
		}
		// Reciprocal hostility.
		reciprocal := false
		for _, e := range enemy.Enemies {
			if e == offended.ID {
				reciprocal = true
				break
			}
		}
		if !reciprocal {
			continue
		}
		if _, exists := m.activeByPair[pairKey(offended.ID, enemy.ID)]; exists {
			continue
		}
		candidates = append(candidates, enemy)
	}
	if len(candidates) == 0 {
		return nil
	}

	if m.rng.Float64() >= intensity {
		return nil
	}

	// Pick uniformly among eligible enemies.
	chosen := candidates[m.rng.Intn(len(candidates))]

	casus := fmt.Sprintf("Incident in %s-aligned space escalated beyond diplomatic repair.", offended.ShortName)
	war := &models.FactionWar{
		ID:             uuid.New(),
		AggressorID:    offended.ID,
		AggressorName:  offended.Name,
		DefenderID:     chosen.ID,
		DefenderName:   chosen.Name,
		Status:         models.FactionWarActive,
		DeclaredAt:     m.now(),
		WarZoneSystems: warZoneSystems(offended, chosen),
		CasusBelli:     casus,
	}
	m.wars[war.ID] = war
	m.activeByPair[pairKey(offended.ID, chosen.ID)] = war
	for _, sys := range war.WarZoneSystems {
		m.addWarZone(sys, war.ID)
	}
	m.emitDeclarationNews(war)
	return war
}

// ============================================================================
// Internal helpers
// ============================================================================

// pairKey returns a direction-agnostic key for two faction IDs by
// sorting them alphabetically. Keeping wars keyed this way means
// DeclareWar(A, B) and a later GetWarBetween(B, A) find the same
// record without duplicating state.
func pairKey(a, b string) string {
	if a <= b {
		return a + "|" + b
	}
	return b + "|" + a
}

// warZoneSystems returns the deduplicated union of both factions'
// core + influence system names. Snapshot — the caller stores this
// on the war record so later territory shifts don't change what was
// once a hot zone.
func warZoneSystems(a, b *models.NPCFaction) []string {
	if a == nil || b == nil {
		return nil
	}
	seen := make(map[string]struct{}, len(a.CoreSystems)+len(b.CoreSystems))
	var out []string
	add := func(systems []string) {
		for _, s := range systems {
			key := strings.ToLower(strings.TrimSpace(s))
			if key == "" {
				continue
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, s) // preserve original casing for display
		}
	}
	add(a.CoreSystems)
	add(a.Influence)
	add(b.CoreSystems)
	add(b.Influence)
	sort.Strings(out) // stable ordering for snapshots & tests
	return out
}

func (m *Manager) addWarZone(system string, warID uuid.UUID) {
	key := strings.ToLower(strings.TrimSpace(system))
	if key == "" {
		return
	}
	if _, ok := m.warZones[key]; !ok {
		m.warZones[key] = make(map[uuid.UUID]struct{})
	}
	m.warZones[key][warID] = struct{}{}
}

func (m *Manager) removeWarZone(system string, warID uuid.UUID) {
	key := strings.ToLower(strings.TrimSpace(system))
	if key == "" {
		return
	}
	if set, ok := m.warZones[key]; ok {
		delete(set, warID)
		if len(set) == 0 {
			delete(m.warZones, key)
		}
	}
}

// ============================================================================
// News emission
// ============================================================================

func (m *Manager) emitDeclarationNews(war *models.FactionWar) {
	if m.newsBus == nil {
		return
	}
	headline := fmt.Sprintf("%s declares war on %s", war.AggressorName, war.DefenderName)
	body := fmt.Sprintf("Hostilities have begun between %s and %s. %d systems are now active war zones.",
		war.AggressorName, war.DefenderName, len(war.WarZoneSystems))
	if war.CasusBelli != "" {
		body = fmt.Sprintf("%s Casus belli: %s.", body, war.CasusBelli)
	}
	m.newsBus.AddArticle(models.NewNewsArticle(
		models.NewsCategoryPolitical,
		models.NewsPriorityCritical,
		headline,
		body,
	))
}

func (m *Manager) emitResolutionNews(war *models.FactionWar) {
	if m.newsBus == nil {
		return
	}
	winnerName := war.AggressorName
	loserName := war.DefenderName
	if war.WinnerFactionID == war.DefenderID {
		winnerName = war.DefenderName
		loserName = war.AggressorName
	}
	headline := fmt.Sprintf("War ends: %s claims victory over %s", winnerName, loserName)
	body := fmt.Sprintf("The war between %s and %s has ended. %s has emerged victorious; territorial and reputation adjustments will follow in the coming days.",
		war.AggressorName, war.DefenderName, winnerName)
	m.newsBus.AddArticle(models.NewNewsArticle(
		models.NewsCategoryPolitical,
		models.NewsPriorityHigh,
		headline,
		body,
	))
}

func (m *Manager) emitCeaseFireNews(war *models.FactionWar) {
	if m.newsBus == nil {
		return
	}
	headline := fmt.Sprintf("Ceasefire: %s and %s halt hostilities", war.AggressorName, war.DefenderName)
	body := fmt.Sprintf("%s and %s have signed a mutual ceasefire. War zones are no longer active; diplomatic relations will need time to recover.",
		war.AggressorName, war.DefenderName)
	m.newsBus.AddArticle(models.NewNewsArticle(
		models.NewsCategoryPolitical,
		models.NewsPriorityHigh,
		headline,
		body,
	))
}
