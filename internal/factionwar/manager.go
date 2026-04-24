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
	}
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

	m.emitResolutionNews(war)
	return nil
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
