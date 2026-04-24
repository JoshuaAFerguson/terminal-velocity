// File: internal/factionwar/manager_test.go
// Project: Terminal Velocity
// Description: Tests for the faction war manager. Covers declare /
//   resolve / ceasefire lifecycles, war zone index correctness,
//   sentinel errors, and news emission.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package factionwar

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// fakeNewsBus records every article for inspection — no real news
// manager needed to assert the manager's news emission.
type fakeNewsBus struct {
	mu       sync.Mutex
	articles []*models.NewsArticle
}

func (f *fakeNewsBus) AddArticle(a *models.NewsArticle) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.articles = append(f.articles, a)
}

func (f *fakeNewsBus) snapshot() []*models.NewsArticle {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*models.NewsArticle, len(f.articles))
	copy(out, f.articles)
	return out
}

// testFactions returns two fixture NPC factions that share no systems
// initially, plus a third sharing one system with the first. Used to
// keep test setup short without depending on the live
// StandardNPCFactions data (which can evolve).
func testFactions() (a, b, c *models.NPCFaction) {
	a = &models.NPCFaction{
		ID:          "fac_a",
		Name:        "Alpha Republic",
		ShortName:   "ALR",
		CoreSystems: []string{"Sol", "Alpha Centauri"},
		Influence:   []string{"Procyon"},
	}
	b = &models.NPCFaction{
		ID:          "fac_b",
		Name:        "Crimson Pact",
		ShortName:   "CRP",
		CoreSystems: []string{"Wolf 359"},
		Influence:   []string{"Barnard"},
	}
	c = &models.NPCFaction{
		ID:          "fac_c",
		Name:        "Eastern Coalition",
		ShortName:   "ECO",
		CoreSystems: []string{"Sol"}, // shares with A for overlap tests
		Influence:   []string{},
	}
	return
}

// newTestManager gives a manager with a controllable clock and a
// fake news bus so tests can assert both state and side effects.
func newTestManager() (*Manager, *fakeNewsBus, *time.Time) {
	bus := &fakeNewsBus{}
	m := NewManager(bus)
	clock := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return clock }
	return m, bus, &clock
}

func TestDeclareWarHappyPath(t *testing.T) {
	m, bus, _ := newTestManager()
	a, b, _ := testFactions()

	war, err := m.DeclareWar(a, b, "border dispute")
	if err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}
	if war.Status != models.FactionWarActive {
		t.Errorf("status: got %q, want active", war.Status)
	}
	if war.AggressorID != a.ID || war.DefenderID != b.ID {
		t.Errorf("belligerents swapped: got %s→%s, want %s→%s",
			war.AggressorID, war.DefenderID, a.ID, b.ID)
	}
	// Zones: union of a.Core(Sol, Alpha Centauri) + a.Influence(Procyon)
	//        + b.Core(Wolf 359) + b.Influence(Barnard)  = 5 systems
	if len(war.WarZoneSystems) != 5 {
		t.Errorf("zones count: got %d, want 5", len(war.WarZoneSystems))
	}
	// CasusBelli flows through to the news body.
	arts := bus.snapshot()
	if len(arts) != 1 {
		t.Fatalf("expected 1 declaration article, got %d", len(arts))
	}
	if arts[0].Priority != models.NewsPriorityCritical {
		t.Errorf("declaration priority: got %v, want critical", arts[0].Priority)
	}
}

func TestDeclareWarRejectsSameFaction(t *testing.T) {
	m, _, _ := newTestManager()
	a, _, _ := testFactions()
	_, err := m.DeclareWar(a, a, "self-attack")
	if !errors.Is(err, ErrSameFaction) {
		t.Errorf("self-war: got %v, want ErrSameFaction", err)
	}
}

func TestDeclareWarRejectsNilFaction(t *testing.T) {
	m, _, _ := newTestManager()
	a, _, _ := testFactions()
	if _, err := m.DeclareWar(nil, a, ""); !errors.Is(err, ErrNilFaction) {
		t.Errorf("nil aggressor: got %v, want ErrNilFaction", err)
	}
	if _, err := m.DeclareWar(a, nil, ""); !errors.Is(err, ErrNilFaction) {
		t.Errorf("nil defender: got %v, want ErrNilFaction", err)
	}
}

func TestDeclareWarRejectsDuplicate(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	if _, err := m.DeclareWar(a, b, ""); err != nil {
		t.Fatalf("first declare: %v", err)
	}
	// Second declare in either direction → ErrAlreadyAtWar.
	if _, err := m.DeclareWar(a, b, ""); !errors.Is(err, ErrAlreadyAtWar) {
		t.Errorf("same-direction redeclare: got %v, want ErrAlreadyAtWar", err)
	}
	if _, err := m.DeclareWar(b, a, ""); !errors.Is(err, ErrAlreadyAtWar) {
		t.Errorf("reverse-direction redeclare: got %v, want ErrAlreadyAtWar", err)
	}
}

func TestResolveWarHappyPath(t *testing.T) {
	m, bus, clock := newTestManager()
	a, b, _ := testFactions()

	war, _ := m.DeclareWar(a, b, "")
	*clock = clock.Add(24 * time.Hour)

	if err := m.ResolveWar(war.ID, a.ID); err != nil {
		t.Fatalf("ResolveWar: %v", err)
	}
	if war.Status != models.FactionWarResolved {
		t.Errorf("status: got %q, want resolved", war.Status)
	}
	if war.WinnerFactionID != a.ID {
		t.Errorf("winner: got %q, want %q", war.WinnerFactionID, a.ID)
	}
	if war.ResolvedAt == nil {
		t.Fatal("ResolvedAt should be set")
	}
	// War zones cleared.
	if m.IsSystemWarZone("Sol") {
		t.Error("Sol should no longer be a war zone after resolve")
	}
	// IsAtWar flips to false for both sides.
	if m.IsAtWar(a.ID) || m.IsAtWar(b.ID) {
		t.Error("neither faction should be at war after resolve")
	}
	// Duration pulled from resolved timestamp (24h), not clock.
	if d := war.Duration(clock.Add(72 * time.Hour)); d != 24*time.Hour {
		t.Errorf("resolved-war duration should use ResolvedAt: got %v, want 24h", d)
	}
	// News article emitted.
	arts := bus.snapshot()
	if len(arts) != 2 { // declare + resolve
		t.Fatalf("expected 2 articles, got %d", len(arts))
	}
}

func TestResolveWarRejectsInvalidWinner(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	// A third faction can't win a war they weren't in.
	if err := m.ResolveWar(war.ID, "fac_c"); !errors.Is(err, ErrInvalidWinner) {
		t.Errorf("foreign winner: got %v, want ErrInvalidWinner", err)
	}
}

func TestResolveWarRejectsUnknownWar(t *testing.T) {
	m, _, _ := newTestManager()
	randomID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	if err := m.ResolveWar(randomID, "anything"); !errors.Is(err, ErrWarNotFound) {
		t.Errorf("unknown war: got %v, want ErrWarNotFound", err)
	}
}

func TestResolveWarRejectsTerminalWar(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")
	_ = m.ResolveWar(war.ID, a.ID)

	// Can't resolve again.
	if err := m.ResolveWar(war.ID, a.ID); !errors.Is(err, ErrWarNotActive) {
		t.Errorf("double-resolve: got %v, want ErrWarNotActive", err)
	}
	// Can't ceasefire either.
	if err := m.CeaseFire(war.ID); !errors.Is(err, ErrWarNotActive) {
		t.Errorf("ceasefire after resolve: got %v, want ErrWarNotActive", err)
	}
}

func TestCeaseFireHappyPath(t *testing.T) {
	m, bus, _ := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	if err := m.CeaseFire(war.ID); err != nil {
		t.Fatalf("CeaseFire: %v", err)
	}
	if war.Status != models.FactionWarCeased {
		t.Errorf("status: got %q, want ceased", war.Status)
	}
	if war.WinnerFactionID != "" {
		t.Errorf("ceasefire should have no winner, got %q", war.WinnerFactionID)
	}
	if m.IsSystemWarZone("Sol") {
		t.Error("war zones should clear on ceasefire")
	}
	arts := bus.snapshot()
	if len(arts) != 2 {
		t.Fatalf("expected declare + ceasefire articles, got %d", len(arts))
	}
}

func TestIsAtWar(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, c := testFactions()
	if m.IsAtWar(a.ID) {
		t.Error("peace-time: should not be at war")
	}
	_, _ = m.DeclareWar(a, b, "")
	if !m.IsAtWar(a.ID) {
		t.Error("A is aggressor: should be at war")
	}
	if !m.IsAtWar(b.ID) {
		t.Error("B is defender: should be at war")
	}
	if m.IsAtWar(c.ID) {
		t.Error("C is neutral: should not be at war")
	}
}

func TestGetWarBetweenIsDirectionAgnostic(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	if got := m.GetWarBetween(a.ID, b.ID); got == nil || got.ID != war.ID {
		t.Errorf("A→B lookup: got %v, want %v", got, war.ID)
	}
	if got := m.GetWarBetween(b.ID, a.ID); got == nil || got.ID != war.ID {
		t.Errorf("B→A lookup: got %v, want %v", got, war.ID)
	}
	if got := m.GetWarBetween("nope", a.ID); got != nil {
		t.Errorf("unknown-faction lookup should be nil, got %v", got)
	}
}

func TestIsSystemWarZoneIsCaseInsensitive(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	_, _ = m.DeclareWar(a, b, "")
	for _, variant := range []string{"Sol", "sol", "SOL", "sOl"} {
		if !m.IsSystemWarZone(variant) {
			t.Errorf("war zone %q should match (case-insensitive)", variant)
		}
	}
	if m.IsSystemWarZone("   Sol   ") {
		// IsSystemWarZone does NOT auto-trim lookups on purpose —
		// the manager trims on insert only; callers are expected
		// to pass clean system names. Documenting via test.
		t.Logf("note: whitespace lookups are not auto-trimmed (caller's responsibility)")
	}
}

func TestWarsInSystemOverlap(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, c := testFactions()
	_, _ = m.DeclareWar(a, b, "") // Sol is in A's core
	_, _ = m.DeclareWar(c, b, "") // Sol is in C's core too → two wars in Sol

	wars := m.WarsInSystem("Sol")
	if len(wars) != 2 {
		t.Fatalf("expected 2 wars covering Sol, got %d", len(wars))
	}
	// Ordering is by declaration time; both share the same clock
	// instant here, but the older one is still at index 0 per
	// map iteration + sort.
	wars = m.WarsInSystem("Barnard") // only in B's influence → 2 wars (both vs B)
	if len(wars) != 2 {
		t.Fatalf("expected 2 wars covering Barnard, got %d", len(wars))
	}
	if wars := m.WarsInSystem("Unknown"); wars != nil {
		t.Errorf("peaceful system should return nil, got %v", wars)
	}
}

func TestWarZoneCleanupOnResolve(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, c := testFactions()
	_, _ = m.DeclareWar(a, b, "")     // war1 covers Sol
	war2, _ := m.DeclareWar(c, b, "") // war2 covers Sol
	if len(m.WarsInSystem("Sol")) != 2 {
		t.Fatal("setup: expected 2 wars covering Sol")
	}
	_ = m.ResolveWar(war2.ID, c.ID)
	// One war removed — Sol should still be a zone (war1 active).
	if !m.IsSystemWarZone("Sol") {
		t.Error("Sol should still be a war zone while war1 is active")
	}
	if got := len(m.WarsInSystem("Sol")); got != 1 {
		t.Errorf("wars in Sol after war2 resolve: got %d, want 1", got)
	}
}

func TestGetActiveWarsSortedByDeclaration(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, c := testFactions()

	// Declare war1 first (earliest)
	war1, _ := m.DeclareWar(a, b, "")
	*clock = clock.Add(1 * time.Hour)
	war2, _ := m.DeclareWar(c, b, "")

	active := m.GetActiveWars()
	if len(active) != 2 {
		t.Fatalf("expected 2 active wars, got %d", len(active))
	}
	if active[0].ID != war1.ID {
		t.Errorf("expected oldest first: got %v, want %v", active[0].ID, war1.ID)
	}
	if active[1].ID != war2.ID {
		t.Errorf("expected newest second: got %v, want %v", active[1].ID, war2.ID)
	}
}

func TestGetAllWarsSortedNewestFirst(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, c := testFactions()

	war1, _ := m.DeclareWar(a, b, "")
	_ = m.ResolveWar(war1.ID, a.ID)
	*clock = clock.Add(1 * time.Hour)
	war2, _ := m.DeclareWar(c, b, "")

	all := m.GetAllWars()
	if len(all) != 2 {
		t.Fatalf("expected 2 wars total, got %d", len(all))
	}
	// Newest first — war2 declared later.
	if all[0].ID != war2.ID {
		t.Errorf("expected newest first: got %v, want %v", all[0].ID, war2.ID)
	}
}

func TestManagerHandlesNilNewsBus(t *testing.T) {
	// Construct without news — no panics on declare/resolve/cease.
	m := NewManager(nil)
	m.now = func() time.Time { return time.Now() }
	a, b, _ := testFactions()

	war, err := m.DeclareWar(a, b, "")
	if err != nil {
		t.Fatalf("DeclareWar with nil news bus: %v", err)
	}
	if err := m.ResolveWar(war.ID, a.ID); err != nil {
		t.Fatalf("ResolveWar with nil news bus: %v", err)
	}
}

func TestWarZoneSystemsDedupesOverlap(t *testing.T) {
	a, b, c := testFactions()
	_ = b
	// A and C both have Sol in CoreSystems; union should contain
	// Sol exactly once.
	zones := warZoneSystems(a, c)
	solCount := 0
	for _, z := range zones {
		if z == "Sol" {
			solCount++
		}
	}
	if solCount != 1 {
		t.Errorf("Sol should appear once, got %d", solCount)
	}
}

func TestWarZoneSystemsNilGuard(t *testing.T) {
	if zones := warZoneSystems(nil, nil); zones != nil {
		t.Errorf("nil factions: got %v, want nil", zones)
	}
	a, _, _ := testFactions()
	if zones := warZoneSystems(a, nil); zones != nil {
		t.Errorf("nil defender: got %v, want nil", zones)
	}
}

func TestPairKeyIsStable(t *testing.T) {
	if pairKey("a", "b") != pairKey("b", "a") {
		t.Error("pairKey should be direction-agnostic")
	}
	if pairKey("a", "b") != "a|b" {
		t.Errorf("pairKey: got %q, want a|b", pairKey("a", "b"))
	}
}

// ============================================================================
// P5C-3 gameplay integration
// ============================================================================

func TestIsWarMaterial(t *testing.T) {
	tests := map[string]bool{
		"weapons":    true,
		"medical":    true,
		"industrial": true,
		"ore":        true,
		// Non-war-material categories stay unaffected:
		"food":        false,
		"electronics": false,
		"luxuries":    false,
		"contraband":  false,
		"":            false,
		"WEAPONS":     false, // case-sensitive per models constants
	}
	for category, want := range tests {
		if got := IsWarMaterial(category); got != want {
			t.Errorf("IsWarMaterial(%q) = %v, want %v", category, got, want)
		}
	}
}

func TestWarZoneReputationScale(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, c := testFactions()

	// No wars → baseline 1.0 for anyone.
	if got := m.WarZoneReputationScale("Sol", a.ID); got != 1.0 {
		t.Errorf("peacetime Sol: got %v, want 1.0", got)
	}

	// Declare A vs B. Sol is in A's core → war zone.
	_, _ = m.DeclareWar(a, b, "")
	if got := m.WarZoneReputationScale("Sol", a.ID); got != WarZoneReputationMultiplier {
		t.Errorf("belligerent A in war zone Sol: got %v, want %v", got, WarZoneReputationMultiplier)
	}
	if got := m.WarZoneReputationScale("Sol", b.ID); got != WarZoneReputationMultiplier {
		t.Errorf("belligerent B in war zone Sol: got %v, want %v", got, WarZoneReputationMultiplier)
	}
	// Third faction not at war — no amplification even in the zone.
	if got := m.WarZoneReputationScale("Sol", c.ID); got != 1.0 {
		t.Errorf("neutral faction in war zone: got %v, want 1.0", got)
	}
	// Peaceful system — no amplification for anyone.
	if got := m.WarZoneReputationScale("Vega", a.ID); got != 1.0 {
		t.Errorf("belligerent in peaceful system: got %v, want 1.0", got)
	}
}

func TestWarZoneReputationScaleNilManager(t *testing.T) {
	var m *Manager // nil
	if got := m.WarZoneReputationScale("Sol", "any"); got != 1.0 {
		t.Errorf("nil manager: got %v, want 1.0", got)
	}
}

func TestWarEconomyPriceMultiplier(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()

	// Peacetime — baseline even for war materials.
	if got := m.WarEconomyPriceMultiplier("Sol", "weapons"); got != 1.0 {
		t.Errorf("peacetime weapons: got %v, want 1.0", got)
	}

	_, _ = m.DeclareWar(a, b, "")

	// War material in a war zone → multiplied.
	if got := m.WarEconomyPriceMultiplier("Sol", "weapons"); got != WarEconomyMultiplier {
		t.Errorf("wartime weapons in Sol: got %v, want %v", got, WarEconomyMultiplier)
	}
	if got := m.WarEconomyPriceMultiplier("Sol", "medical"); got != WarEconomyMultiplier {
		t.Errorf("wartime medical in Sol: got %v, want %v", got, WarEconomyMultiplier)
	}

	// Non-war-material in a war zone → baseline.
	if got := m.WarEconomyPriceMultiplier("Sol", "food"); got != 1.0 {
		t.Errorf("wartime food: got %v, want 1.0", got)
	}
	if got := m.WarEconomyPriceMultiplier("Sol", "luxuries"); got != 1.0 {
		t.Errorf("wartime luxuries: got %v, want 1.0", got)
	}

	// War material outside the zone → baseline.
	if got := m.WarEconomyPriceMultiplier("Vega", "weapons"); got != 1.0 {
		t.Errorf("peaceful-system weapons: got %v, want 1.0", got)
	}
}

func TestWarEconomyPriceMultiplierNilManager(t *testing.T) {
	var m *Manager
	if got := m.WarEconomyPriceMultiplier("Sol", "weapons"); got != 1.0 {
		t.Errorf("nil manager: got %v, want 1.0", got)
	}
}

// ============================================================================
// P5C-4 lifecycle automation
// ============================================================================

// scriptedRNG returns the configured floats / ints in order, one per
// Float64() / Intn() call respectively. Runs off the end → panics so
// tests catch unexpected extra RNG consumption. Used instead of a
// fixed-value RNG because TickWars makes multiple distinct rolls per
// invocation (declaration chance, aggressor coin flip, winner coin
// flip) and we need to script each independently.
type scriptedRNG struct {
	floats []float64
	ints   []int
}

func (r *scriptedRNG) Float64() float64 {
	if len(r.floats) == 0 {
		panic("scriptedRNG: out of floats")
	}
	v := r.floats[0]
	r.floats = r.floats[1:]
	return v
}

func (r *scriptedRNG) Intn(n int) int {
	if len(r.ints) == 0 {
		panic("scriptedRNG: out of ints")
	}
	v := r.ints[0]
	r.ints = r.ints[1:]
	return v
}

// hostileFactionPair returns two factions that list each other as
// mutual enemies. Used to build a predictable StandardNPCFactions
// slice for TickWars tests that doesn't depend on the live data.
func hostileFactionPair() (a, b models.NPCFaction) {
	a = models.NPCFaction{
		ID:          "hostile_a",
		Name:        "A",
		ShortName:   "A",
		CoreSystems: []string{"Sol"},
		Enemies:     []string{"hostile_b"},
	}
	b = models.NPCFaction{
		ID:          "hostile_b",
		Name:        "B",
		ShortName:   "B",
		CoreSystems: []string{"Wolf 359"},
		Enemies:     []string{"hostile_a"},
	}
	return
}

func TestDefaultLifecycleConfig(t *testing.T) {
	c := DefaultLifecycleConfig()
	if c.MaxWarDuration != 7*24*time.Hour {
		t.Errorf("default max war duration: got %v, want 7d", c.MaxWarDuration)
	}
	if c.DeclarationProbability <= 0 || c.DeclarationProbability >= 1 {
		t.Errorf("declaration probability out of (0,1): got %v", c.DeclarationProbability)
	}
	if c.EmergentCasusBelli == "" {
		t.Error("default casus belli should be non-empty")
	}
}

func TestBuildHostilePairsSkipsOneWayAnimosity(t *testing.T) {
	// A lists B as enemy, but B doesn't reciprocate → no pair.
	one := []models.NPCFaction{
		{ID: "x", Enemies: []string{"y"}},
		{ID: "y", Enemies: []string{}},
	}
	if got := buildHostilePairs(one); len(got) != 0 {
		t.Errorf("one-way animosity should produce 0 pairs, got %d", len(got))
	}
}

func TestBuildHostilePairsDedupesReciprocal(t *testing.T) {
	a, b := hostileFactionPair()
	got := buildHostilePairs([]models.NPCFaction{a, b})
	if len(got) != 1 {
		t.Fatalf("mutual hostility should produce 1 pair, got %d", len(got))
	}
	// Canonical ordering: alphabetical by ID.
	if got[0].a.ID != "hostile_a" || got[0].b.ID != "hostile_b" {
		t.Errorf("pair ordering: got (%s, %s), want (hostile_a, hostile_b)",
			got[0].a.ID, got[0].b.ID)
	}
}

func TestBuildHostilePairsIgnoresUnknownEnemy(t *testing.T) {
	factions := []models.NPCFaction{
		{ID: "x", Enemies: []string{"ghost"}}, // "ghost" doesn't exist
	}
	if got := buildHostilePairs(factions); len(got) != 0 {
		t.Errorf("unknown enemy should be skipped, got %d pairs", len(got))
	}
}

func TestTickWarsResolvesExpiredWars(t *testing.T) {
	m, bus, clock := newTestManager()
	a, b, _ := testFactions()

	war, _ := m.DeclareWar(a, b, "test")

	// RNG script: Intn(2) for winner coin flip → 1 (aggressor
	// wins per the code's convention: 0 picks defender, non-zero
	// stays on aggressor). Float64 is still consumed in the
	// second pass: after the expired war resolves, the same pair
	// becomes eligible again in-tick, so TickWars rolls
	// declaration for it. DeclarationProbability 0.0 means any
	// float ≥ 0.0 (i.e. all of them) skips — we just have to
	// supply one to feed the RNG.
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{1}})
	m.SetLifecycleConfig(LifecycleConfig{
		MaxWarDuration:         1 * time.Hour,
		DeclarationProbability: 0.0,
	})

	// Advance past max duration.
	*clock = clock.Add(2 * time.Hour)

	// TickWars fixture: mutual-hostile pair matching war belligerents.
	tickFactions := []models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	}
	m.TickWars(tickFactions)

	if war.Status != models.FactionWarResolved {
		t.Errorf("expected war to be resolved by TickWars, status %q", war.Status)
	}
	if war.WinnerFactionID != a.ID {
		t.Errorf("expected aggressor winner (coin flip 0), got %q", war.WinnerFactionID)
	}
	// Declaration + resolution news articles emitted.
	if len(bus.snapshot()) < 2 {
		t.Errorf("expected declare + resolve articles, got %d", len(bus.snapshot()))
	}
}

func TestTickWarsExpiredDefenderWinsOnCoinFlip1(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	// Ints[0]=0 → defender wins coin flip (code convention: 0
	// picks defender, non-zero stays on aggressor). Float64 is
	// consumed in the post-resolve declaration pass.
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{0}})
	m.SetLifecycleConfig(LifecycleConfig{MaxWarDuration: time.Hour, DeclarationProbability: 0.0})

	*clock = clock.Add(2 * time.Hour)
	m.TickWars([]models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	})

	if war.WinnerFactionID != b.ID {
		t.Errorf("coin flip 1 should make defender win, got %q", war.WinnerFactionID)
	}
}

func TestTickWarsRollsEmergentDeclaration(t *testing.T) {
	m, _, _ := newTestManager()
	// Float64 roll: 0.0 is below threshold → declaration fires.
	// Intn(2) for aggressor pick → 0 (first pair member).
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.0}, ints: []int{0}})
	m.SetLifecycleConfig(LifecycleConfig{
		MaxWarDuration:         time.Hour,
		DeclarationProbability: 0.5,
		EmergentCasusBelli:     "test incident",
	})

	a, b := hostileFactionPair()
	m.TickWars([]models.NPCFaction{a, b})

	active := m.GetActiveWars()
	if len(active) != 1 {
		t.Fatalf("expected 1 emergent war, got %d", len(active))
	}
	if active[0].AggressorID != a.ID {
		t.Errorf("Intn(2)=0 should make canonical-first the aggressor, got %q", active[0].AggressorID)
	}
	if active[0].CasusBelli != "test incident" {
		t.Errorf("casus belli: got %q, want %q", active[0].CasusBelli, "test incident")
	}
}

func TestTickWarsRollsAboveThresholdDoesNothing(t *testing.T) {
	m, _, _ := newTestManager()
	// Float64 above the 0.1 threshold → no declaration. No Intn
	// should be consumed (panics scriptedRNG if it is).
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.99}, ints: nil})
	m.SetLifecycleConfig(LifecycleConfig{
		MaxWarDuration:         time.Hour,
		DeclarationProbability: 0.1,
	})
	a, b := hostileFactionPair()
	m.TickWars([]models.NPCFaction{a, b})
	if got := len(m.GetActiveWars()); got != 0 {
		t.Errorf("expected 0 wars, got %d", got)
	}
}

func TestTickWarsSkipsPairAlreadyAtWar(t *testing.T) {
	m, _, _ := newTestManager()
	a := models.NPCFaction{ID: "a", Name: "A", CoreSystems: []string{"Sol"}, Enemies: []string{"b"}}
	b := models.NPCFaction{ID: "b", Name: "B", CoreSystems: []string{"Wolf 359"}, Enemies: []string{"a"}}

	// Declare war first (uses the real DeclareWar).
	_, err := m.DeclareWar(&a, &b, "")
	if err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}
	// RNG intentionally empty — we expect no rolls because the
	// only pair is already at war.
	m.SetLifecycleRNG(&scriptedRNG{})
	m.SetLifecycleConfig(LifecycleConfig{
		MaxWarDuration:         time.Hour,
		DeclarationProbability: 1.0, // would otherwise always fire
	})
	m.TickWars([]models.NPCFaction{a, b})
	if got := len(m.GetActiveWars()); got != 1 {
		t.Errorf("should still have 1 war (no new declaration), got %d", got)
	}
}

func TestReportIncidentFiresWarWhenIntensityRollsLow(t *testing.T) {
	m, _, _ := newTestManager()
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.1}, ints: []int{0}})

	a, b := hostileFactionPair()
	war := m.ReportIncident(&a, []models.NPCFaction{a, b}, 0.5)
	if war == nil {
		t.Fatalf("intensity 0.5 with roll 0.1 should fire, got nil")
	}
	if war.AggressorID != a.ID || war.DefenderID != b.ID {
		t.Errorf("belligerents: got %s→%s, want %s→%s", war.AggressorID, war.DefenderID, a.ID, b.ID)
	}
}

func TestReportIncidentNoopsWhenRollAboveIntensity(t *testing.T) {
	m, _, _ := newTestManager()
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.9}})
	a, b := hostileFactionPair()
	if war := m.ReportIncident(&a, []models.NPCFaction{a, b}, 0.5); war != nil {
		t.Errorf("roll 0.9 vs intensity 0.5 should no-op, got war %v", war)
	}
}

func TestReportIncidentClampsIntensity(t *testing.T) {
	m, _, _ := newTestManager()
	// Intensity >1 is clamped to 1, so roll 0.99 fires.
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.99}, ints: []int{0}})
	a, b := hostileFactionPair()
	if war := m.ReportIncident(&a, []models.NPCFaction{a, b}, 5.0); war == nil {
		t.Errorf("clamped intensity should fire with roll 0.99, got nil")
	}
}

func TestReportIncidentZeroIntensityIsNoop(t *testing.T) {
	m, _, _ := newTestManager()
	m.SetLifecycleRNG(&scriptedRNG{})
	a, b := hostileFactionPair()
	if war := m.ReportIncident(&a, []models.NPCFaction{a, b}, 0); war != nil {
		t.Errorf("zero intensity should no-op, got war %v", war)
	}
	// Also: negative intensity.
	if war := m.ReportIncident(&a, []models.NPCFaction{a, b}, -1); war != nil {
		t.Errorf("negative intensity should no-op, got war %v", war)
	}
}

func TestReportIncidentSkipsAlreadyAtWarEnemies(t *testing.T) {
	m, _, _ := newTestManager()
	a, b := hostileFactionPair()

	// Existing war between a and b → no enemies left to declare on.
	_, _ = m.DeclareWar(&a, &b, "")
	m.SetLifecycleRNG(&scriptedRNG{})
	if war := m.ReportIncident(&a, []models.NPCFaction{a, b}, 1.0); war != nil {
		t.Errorf("no-available-enemy path should return nil, got %v", war)
	}
}

func TestReportIncidentNilSafety(t *testing.T) {
	var nilMgr *Manager
	if got := nilMgr.ReportIncident(nil, nil, 1.0); got != nil {
		t.Errorf("nil manager should return nil, got %v", got)
	}
	m, _, _ := newTestManager()
	if got := m.ReportIncident(nil, nil, 1.0); got != nil {
		t.Errorf("nil faction should return nil, got %v", got)
	}
}

// ============================================================================
// P5D-1 territory hook integration
// ============================================================================

func TestTerritoryHookFiresOnResolve(t *testing.T) {
	m, _, _ := newTestManager()
	a, b, _ := testFactions()

	// Record every hook invocation so we can assert parameters.
	type hookCall struct {
		zones    []string
		loserID  string
		winnerID string
	}
	var calls []hookCall
	m.SetTerritoryHook(func(zones []string, loserID, winnerID string) {
		calls = append(calls, hookCall{zones: zones, loserID: loserID, winnerID: winnerID})
	})

	war, _ := m.DeclareWar(a, b, "border")
	if err := m.ResolveWar(war.ID, a.ID); err != nil {
		t.Fatalf("ResolveWar: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 hook call, got %d", len(calls))
	}
	if calls[0].winnerID != a.ID {
		t.Errorf("winner: got %q, want %q", calls[0].winnerID, a.ID)
	}
	if calls[0].loserID != b.ID {
		t.Errorf("loser: got %q, want %q", calls[0].loserID, b.ID)
	}
	if len(calls[0].zones) != len(war.WarZoneSystems) {
		t.Errorf("zones count: got %d, want %d", len(calls[0].zones), len(war.WarZoneSystems))
	}
}

func TestTerritoryHookFiresOnAutoResolve(t *testing.T) {
	// TickWars' auto-resolution path must also call the hook — this
	// is how emergent wars flip territory without admin action.
	m, _, clock := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	called := false
	m.SetTerritoryHook(func(zones []string, loserID, winnerID string) {
		called = true
	})
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{1}}) // aggressor wins
	m.SetLifecycleConfig(LifecycleConfig{MaxWarDuration: time.Hour, DeclarationProbability: 0.0})

	*clock = clock.Add(2 * time.Hour)
	m.TickWars([]models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	})

	if !called {
		t.Error("territory hook should fire on auto-resolution too")
	}
	_ = war
}

func TestTerritoryHookNotFiredOnCeaseFire(t *testing.T) {
	// Ceasefire means neither side "won," so no territory flips.
	// The hook is specifically a resolve-path concern.
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	called := false
	m.SetTerritoryHook(func(zones []string, loserID, winnerID string) {
		called = true
	})
	war, _ := m.DeclareWar(a, b, "")
	if err := m.CeaseFire(war.ID); err != nil {
		t.Fatalf("CeaseFire: %v", err)
	}
	if called {
		t.Error("ceasefire should not flip territory")
	}
}

// ============================================================================
// P5D-2 contribution-aware resolution
// ============================================================================

func TestTickWarsUsesContributionLeaderOverCoinFlip(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	// Resolver says defender wins regardless of RNG. The scripted
	// RNG's single int would pick aggressor (Intn=1), so if the
	// resolver wasn't consulted the test would fail.
	m.SetWinnerResolver(func(zones []string, aggID, defID string) (string, int64) {
		return defID, 42
	})
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{1}})
	m.SetLifecycleConfig(LifecycleConfig{MaxWarDuration: time.Hour, DeclarationProbability: 0.0})

	*clock = clock.Add(2 * time.Hour)
	m.TickWars([]models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	})

	if war.WinnerFactionID != b.ID {
		t.Errorf("resolver should override RNG: winner %q, want %q", war.WinnerFactionID, b.ID)
	}
}

func TestTickWarsFallsBackToCoinFlipOnNoLeader(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")

	// Resolver returns "" (tie / no data) → RNG path runs.
	m.SetWinnerResolver(func(zones []string, aggID, defID string) (string, int64) {
		return "", 0
	})
	// Int=1 → aggressor wins per coin-flip convention.
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{1}})
	m.SetLifecycleConfig(LifecycleConfig{MaxWarDuration: time.Hour, DeclarationProbability: 0.0})

	*clock = clock.Add(2 * time.Hour)
	m.TickWars([]models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	})

	if war.WinnerFactionID != a.ID {
		t.Errorf("empty resolver should defer to RNG (aggressor wins on Int=1), got %q", war.WinnerFactionID)
	}
}

func TestTickWarsNilResolverIsSafe(t *testing.T) {
	m, _, clock := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")
	// No resolver set — pure RNG path.
	m.SetLifecycleRNG(&scriptedRNG{floats: []float64{0.5}, ints: []int{0}})
	m.SetLifecycleConfig(LifecycleConfig{MaxWarDuration: time.Hour, DeclarationProbability: 0.0})

	*clock = clock.Add(2 * time.Hour)
	m.TickWars([]models.NPCFaction{
		{ID: a.ID, Enemies: []string{b.ID}, CoreSystems: a.CoreSystems},
		{ID: b.ID, Enemies: []string{a.ID}, CoreSystems: b.CoreSystems},
	})

	// Int=0 → defender wins.
	if war.WinnerFactionID != b.ID {
		t.Errorf("nil resolver + Int=0 coin flip: winner %q, want %q", war.WinnerFactionID, b.ID)
	}
}

func TestTerritoryHookNilSafe(t *testing.T) {
	// nil hook → ResolveWar still succeeds.
	m, _, _ := newTestManager()
	a, b, _ := testFactions()
	war, _ := m.DeclareWar(a, b, "")
	// Hook was nil by default; explicit SetTerritoryHook(nil) is redundant
	// but documents intent.
	m.SetTerritoryHook(nil)
	if err := m.ResolveWar(war.ID, a.ID); err != nil {
		t.Errorf("ResolveWar with nil hook: %v", err)
	}
}

func TestConcurrentDeclareAndQuery(t *testing.T) {
	// Race test — run with -race to catch lock misuse. Spawns one
	// declarer and many readers; all queries should return
	// consistent state without panics.
	m, _, _ := newTestManager()
	a, b, _ := testFactions()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = m.DeclareWar(a, b, "")
	}()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.IsAtWar(a.ID)
			_ = m.GetActiveWars()
			_ = m.IsSystemWarZone("Sol")
		}()
	}
	wg.Wait()
}
