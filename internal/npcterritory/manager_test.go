// File: internal/npcterritory/manager_test.go
// Project: Terminal Velocity
// Description: Tests for NPC system ownership. Covers seed,
//   lookup, transfer/no-op/unknown, war resolution, news emission.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package npcterritory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// recordingBus captures emitted news articles for inspection.
type recordingBus struct {
	mu       sync.Mutex
	articles []*models.NewsArticle
}

func (r *recordingBus) AddArticle(a *models.NewsArticle) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.articles = append(r.articles, a)
}

func (r *recordingBus) snapshot() []*models.NewsArticle {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*models.NewsArticle, len(r.articles))
	copy(out, r.articles)
	return out
}

func testFactions() []models.NPCFaction {
	return []models.NPCFaction{
		{
			ID:          "uef",
			Name:        "United Earth Federation",
			ShortName:   "UEF",
			CoreSystems: []string{"Sol", "Alpha Centauri"},
		},
		{
			ID:          "crimson",
			Name:        "Crimson Collective",
			ShortName:   "CRM",
			CoreSystems: []string{"Wolf 359", "Barnard"},
		},
		{
			ID:          "rom",
			Name:        "Republic of Mars",
			ShortName:   "ROM",
			CoreSystems: []string{"Procyon"},
		},
	}
}

func TestSeedPopulatesOwnership(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())

	if owner, err := m.GetOwner("Sol"); err != nil || owner != "uef" {
		t.Errorf("Sol owner: got (%q, %v), want (uef, nil)", owner, err)
	}
	if owner, err := m.GetOwner("Wolf 359"); err != nil || owner != "crimson" {
		t.Errorf("Wolf 359 owner: got (%q, %v), want (crimson, nil)", owner, err)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	// Re-seed with reduced data — state fully replaced.
	m.Seed([]models.NPCFaction{
		{ID: "uef", Name: "UEF", CoreSystems: []string{"Sol"}},
	})
	if _, err := m.GetOwner("Wolf 359"); !errors.Is(err, ErrUnknownSystem) {
		t.Errorf("re-seed should drop Wolf 359, got err %v", err)
	}
	if owner, _ := m.GetOwner("Sol"); owner != "uef" {
		t.Errorf("re-seed should keep Sol as uef, got %q", owner)
	}
}

func TestGetOwnerCaseInsensitive(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	for _, variant := range []string{"Sol", "sol", "SOL", "  Sol  "} {
		owner, err := m.GetOwner(variant)
		if err != nil || owner != "uef" {
			t.Errorf("case variant %q: got (%q, %v), want (uef, nil)", variant, owner, err)
		}
	}
}

func TestGetOwnerUnknownSystem(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if _, err := m.GetOwner("Nowheresville"); !errors.Is(err, ErrUnknownSystem) {
		t.Errorf("unknown system: got %v, want ErrUnknownSystem", err)
	}
}

func TestGetOwnerNameAndShortName(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if got := m.GetOwnerName("Sol"); got != "United Earth Federation" {
		t.Errorf("full name: got %q", got)
	}
	if got := m.GetOwnerShortName("Sol"); got != "UEF" {
		t.Errorf("short name: got %q", got)
	}
	if got := m.GetOwnerName("Mystery"); got != "" {
		t.Errorf("unknown system: got %q, want empty", got)
	}
}

func TestTransferSystemFlipsOwner(t *testing.T) {
	bus := &recordingBus{}
	m := NewManager(bus)
	m.Seed(testFactions())

	rec, err := m.TransferSystem("Sol", "crimson")
	if err != nil {
		t.Fatalf("TransferSystem: %v", err)
	}
	if rec == nil {
		t.Fatal("expected non-nil flip record")
	}
	if rec.FromID != "uef" || rec.ToID != "crimson" {
		t.Errorf("flip record: from=%q to=%q, want uef→crimson", rec.FromID, rec.ToID)
	}
	if rec.SystemName != "Sol" {
		t.Errorf("flip preserves original casing, got %q", rec.SystemName)
	}
	if got, _ := m.GetOwner("Sol"); got != "crimson" {
		t.Errorf("post-transfer owner: got %q, want crimson", got)
	}
	arts := bus.snapshot()
	if len(arts) != 1 {
		t.Fatalf("expected 1 news article, got %d", len(arts))
	}
	if arts[0].Priority != models.NewsPriorityHigh {
		t.Errorf("flip news priority: got %v, want high", arts[0].Priority)
	}
}

func TestTransferSystemNoopWhenAlreadyOwned(t *testing.T) {
	bus := &recordingBus{}
	m := NewManager(bus)
	m.Seed(testFactions())

	// Sol already owned by UEF — transferring to UEF again should
	// return nil/nil and emit NO news.
	rec, err := m.TransferSystem("Sol", "uef")
	if rec != nil || err != nil {
		t.Errorf("idempotent transfer: got (%v, %v), want (nil, nil)", rec, err)
	}
	if len(bus.snapshot()) != 0 {
		t.Errorf("no news should fire for idempotent transfer")
	}
}

func TestTransferSystemUnknownFaction(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if _, err := m.TransferSystem("Sol", "ghost"); !errors.Is(err, ErrUnknownFaction) {
		t.Errorf("unknown faction: got %v, want ErrUnknownFaction", err)
	}
}

func TestTransferSystemUnknownSystem(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if _, err := m.TransferSystem("Nowhere", "uef"); !errors.Is(err, ErrUnknownSystem) {
		t.Errorf("unknown system: got %v, want ErrUnknownSystem", err)
	}
}

func TestResolveWarTerritoryFlipsLoserOwnedZones(t *testing.T) {
	bus := &recordingBus{}
	m := NewManager(bus)
	m.Seed(testFactions())

	// Zone list includes UEF-owned (Sol), Crimson-owned (Wolf 359),
	// and ROM-owned (Procyon). UEF loses to Crimson.
	zones := []string{"Sol", "Alpha Centauri", "Wolf 359", "Procyon"}
	flips := m.ResolveWarTerritory(zones, "uef", "crimson")

	// Expected: Sol and Alpha Centauri flip (UEF-owned);
	// Wolf 359 stays with Crimson (they're already winner);
	// Procyon stays with ROM (third party).
	if len(flips) != 2 {
		t.Fatalf("expected 2 flips, got %d: %v", len(flips), flips)
	}
	for _, f := range flips {
		if f.FromID != "uef" || f.ToID != "crimson" {
			t.Errorf("flip %+v should be uef→crimson", f)
		}
	}
	if got, _ := m.GetOwner("Procyon"); got != "rom" {
		t.Errorf("third-party Procyon preserved, got %q", got)
	}
	if got, _ := m.GetOwner("Wolf 359"); got != "crimson" {
		t.Errorf("winner-owned Wolf 359 preserved, got %q", got)
	}
	if got, _ := m.GetOwner("Sol"); got != "crimson" {
		t.Errorf("Sol flipped to winner, got %q", got)
	}

	// News: 2 articles (one per flip).
	if len(bus.snapshot()) != 2 {
		t.Errorf("expected 2 flip articles, got %d", len(bus.snapshot()))
	}
}

func TestResolveWarTerritorySameWinnerLoserNoop(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if flips := m.ResolveWarTerritory([]string{"Sol"}, "uef", "uef"); len(flips) != 0 {
		t.Errorf("same-side resolve should be no-op, got %d flips", len(flips))
	}
}

func TestResolveWarTerritoryUnknownWinner(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if flips := m.ResolveWarTerritory([]string{"Sol"}, "uef", "ghost"); len(flips) != 0 {
		t.Errorf("unknown winner should be no-op, got %d flips", len(flips))
	}
}

func TestResolveWarTerritoryNilManager(t *testing.T) {
	var m *Manager
	if flips := m.ResolveWarTerritory([]string{"Sol"}, "a", "b"); flips != nil {
		t.Errorf("nil manager: got %v, want nil", flips)
	}
}

func TestGetFactionSystemsFiltersAndSorts(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	uef := m.GetFactionSystems("uef")
	// Alphabetical: Alpha Centauri, Sol.
	if len(uef) != 2 || uef[0] != "Alpha Centauri" || uef[1] != "Sol" {
		t.Errorf("UEF systems: got %v, want [Alpha Centauri, Sol]", uef)
	}

	// Known faction with no holdings returns empty slice (not nil).
	empty := NewManager(nil)
	empty.Seed([]models.NPCFaction{{ID: "x", Name: "X", CoreSystems: nil}})
	if got := empty.GetFactionSystems("x"); got == nil {
		t.Error("known faction with no systems should return empty slice, not nil")
	}

	// Unknown faction returns nil.
	if got := m.GetFactionSystems("unknown"); got != nil {
		t.Errorf("unknown faction: got %v, want nil", got)
	}
}

func TestAllOwnershipSnapshot(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	snap := m.AllOwnership()
	if len(snap) != 5 { // Sol + AC + W359 + Barnard + Procyon
		t.Fatalf("expected 5 tracked systems, got %d", len(snap))
	}
	if snap["Sol"] != "uef" || snap["Wolf 359"] != "crimson" {
		t.Errorf("snapshot mismatch: %v", snap)
	}
}

func TestFlipRecordTimestamp(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	fixed := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	rec, _ := m.TransferSystem("Sol", "crimson")
	if rec == nil || !rec.At.Equal(fixed) {
		t.Errorf("flip timestamp: got %v, want %v", rec.At, fixed)
	}
}

func TestManagerNilNewsBus(t *testing.T) {
	// Should not panic when no news bus is wired.
	m := NewManager(nil)
	m.Seed(testFactions())
	if _, err := m.TransferSystem("Sol", "crimson"); err != nil {
		t.Errorf("transfer with nil bus: %v", err)
	}
}

// ============================================================================
// P5D-2 contribution tracking
// ============================================================================

func TestAddContributionAccumulates(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 10)
	m.AddContribution("Sol", "uef", 15)
	m.AddContribution("Sol", "crimson", 5)

	if got := m.ContributionFor("Sol", "uef"); got != 25 {
		t.Errorf("uef: got %d, want 25", got)
	}
	if got := m.ContributionFor("Sol", "crimson"); got != 5 {
		t.Errorf("crimson: got %d, want 5", got)
	}
}

func TestAddContributionIgnoresUnknownSystem(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Nowhere", "uef", 100)
	// Unknown system → no-op. Verify via SystemContributions nil
	// return (unknown system).
	if snap := m.SystemContributions("Nowhere"); snap != nil {
		t.Errorf("unknown system contributions: got %v, want nil", snap)
	}
}

func TestAddContributionIgnoresUnknownFaction(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "ghost", 50)
	// Known system but unknown faction → no contribution stored.
	if got := m.ContributionFor("Sol", "ghost"); got != 0 {
		t.Errorf("unknown faction: got %d, want 0", got)
	}
	// Sol's contributions map should be absent since we never
	// stored anything (an empty map would still be a leak).
	if snap := m.SystemContributions("Sol"); len(snap) != 0 {
		t.Errorf("no-ops should not create an empty map, got %v", snap)
	}
}

func TestAddContributionIgnoresZero(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 0)
	if snap := m.SystemContributions("Sol"); len(snap) != 0 {
		t.Errorf("zero-amount should not create entries, got %v", snap)
	}
}

func TestAddContributionAllowsNegative(t *testing.T) {
	// Negative amounts are valid: sabotage / rep penalty hooks use
	// them. Running total must reflect the signed sum.
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 20)
	m.AddContribution("Sol", "uef", -5)
	if got := m.ContributionFor("Sol", "uef"); got != 15 {
		t.Errorf("signed sum: got %d, want 15", got)
	}
}

func TestAddContributionNilManagerSafe(t *testing.T) {
	var m *Manager
	m.AddContribution("Sol", "uef", 10) // must not panic
}

func TestContributionLeaderAcrossSystems(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	// uef leads in Sol; crimson leads in Wolf 359; uef aggregate
	// wins because Sol contributions are larger.
	m.AddContribution("Sol", "uef", 50)
	m.AddContribution("Sol", "crimson", 10)
	m.AddContribution("Wolf 359", "uef", 5)
	m.AddContribution("Wolf 359", "crimson", 20)

	leader, margin := m.ContributionLeader(
		[]string{"Sol", "Wolf 359"}, "uef", "crimson",
	)
	if leader != "uef" {
		t.Errorf("leader: got %q, want uef", leader)
	}
	// uef total = 55, crimson total = 30 → margin 25.
	if margin != 25 {
		t.Errorf("margin: got %d, want 25", margin)
	}
}

func TestContributionLeaderTieReturnsEmpty(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 10)
	m.AddContribution("Sol", "crimson", 10)
	leader, margin := m.ContributionLeader([]string{"Sol"}, "uef", "crimson")
	if leader != "" {
		t.Errorf("tie should return empty leader, got %q", leader)
	}
	if margin != 0 {
		t.Errorf("tie margin: got %d, want 0", margin)
	}
}

func TestContributionLeaderNoDataReturnsEmpty(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	if leader, _ := m.ContributionLeader([]string{"Sol"}, "uef", "crimson"); leader != "" {
		t.Errorf("empty contributions: leader %q, want empty", leader)
	}
}

func TestContributionLeaderNilManagerSafe(t *testing.T) {
	var m *Manager
	leader, margin := m.ContributionLeader([]string{"Sol"}, "uef", "crimson")
	if leader != "" || margin != 0 {
		t.Errorf("nil manager: got (%q, %d), want empty", leader, margin)
	}
}

func TestFlipClearsContributions(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 30)
	m.AddContribution("Sol", "crimson", 20)

	// Transfer flips ownership — contributions should reset for
	// the flipped system so the new owner starts fresh.
	_, err := m.TransferSystem("Sol", "crimson")
	if err != nil {
		t.Fatalf("TransferSystem: %v", err)
	}
	if got := m.ContributionFor("Sol", "uef"); got != 0 {
		t.Errorf("post-flip uef contribution: got %d, want 0", got)
	}
	if got := m.ContributionFor("Sol", "crimson"); got != 0 {
		t.Errorf("post-flip crimson contribution: got %d, want 0", got)
	}
}

func TestResolveWarTerritoryClearsContributionsOnFlip(t *testing.T) {
	// Same clearing behavior, but via the war-resolution path.
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 100)
	m.ResolveWarTerritory([]string{"Sol"}, "uef", "crimson")
	if got := m.ContributionFor("Sol", "uef"); got != 0 {
		t.Errorf("post-war-resolve contribution: got %d, want 0", got)
	}
}

func TestSystemContributionsSnapshot(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.AddContribution("Sol", "uef", 10)
	m.AddContribution("Sol", "crimson", 5)

	snap := m.SystemContributions("Sol")
	if snap == nil || snap["uef"] != 10 || snap["crimson"] != 5 {
		t.Errorf("snapshot: got %v, want uef:10 crimson:5", snap)
	}
	// Mutating the snapshot must not bleed into manager state.
	snap["uef"] = 9999
	if got := m.ContributionFor("Sol", "uef"); got != 10 {
		t.Errorf("snapshot leaked mutation: got %d, want 10", got)
	}
}

// ============================================================================
// P5D-3 persistence
// ============================================================================

// fakePersister records every upsert call. Order matters for the
// transfer vs. war-resolve tests, so we keep the full sequence
// rather than a map.
type persistCall struct {
	system  string
	faction string
}

type fakePersister struct {
	mu    sync.Mutex
	calls []persistCall
	err   error // set non-nil to simulate DB failure
}

func (p *fakePersister) hook(ctx context.Context, systemName, factionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, persistCall{system: systemName, faction: factionID})
	return p.err
}

func (p *fakePersister) snapshot() []persistCall {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]persistCall, len(p.calls))
	copy(out, p.calls)
	return out
}

func TestPersisterFiresOnTransfer(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	p := &fakePersister{}
	m.SetPersister(p.hook)

	if _, err := m.TransferSystem("Sol", "crimson"); err != nil {
		t.Fatalf("TransferSystem: %v", err)
	}
	calls := p.snapshot()
	if len(calls) != 1 {
		t.Fatalf("expected 1 persist call, got %d", len(calls))
	}
	if calls[0].system != "Sol" || calls[0].faction != "crimson" {
		t.Errorf("call: got %+v, want {Sol, crimson}", calls[0])
	}
}

func TestPersisterDoesNotFireOnIdempotentTransfer(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	p := &fakePersister{}
	m.SetPersister(p.hook)

	// Transfer Sol to its current owner (uef) — no-op path should
	// NOT hit the DB.
	_, _ = m.TransferSystem("Sol", "uef")
	if got := len(p.snapshot()); got != 0 {
		t.Errorf("idempotent transfer should skip persister, got %d calls", got)
	}
}

func TestPersisterFiresPerFlipOnResolveWarTerritory(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	p := &fakePersister{}
	m.SetPersister(p.hook)

	// UEF owns Sol + Alpha Centauri; war zone includes both plus
	// Procyon (ROM) and Wolf 359 (Crimson). Only UEF-owned
	// systems should persist.
	zones := []string{"Sol", "Alpha Centauri", "Wolf 359", "Procyon"}
	m.ResolveWarTerritory(zones, "uef", "crimson")

	calls := p.snapshot()
	if len(calls) != 2 {
		t.Fatalf("expected 2 persist calls (UEF-owned flips), got %d", len(calls))
	}
	for _, c := range calls {
		if c.faction != "crimson" {
			t.Errorf("flip should target winner crimson, got %+v", c)
		}
	}
}

func TestPersisterErrorDoesNotRollbackFlip(t *testing.T) {
	// A DB hiccup must not undo the in-memory flip. Better to show
	// the player the new owner and re-sync on next restart than to
	// resurrect the old one and create ghost states.
	m := NewManager(nil)
	m.Seed(testFactions())
	p := &fakePersister{err: errors.New("simulated DB failure")}
	m.SetPersister(p.hook)

	rec, err := m.TransferSystem("Sol", "crimson")
	if err != nil {
		t.Fatalf("persister error should not surface as transfer error: %v", err)
	}
	if rec == nil {
		t.Fatal("flip should still succeed despite persister error")
	}
	if owner, _ := m.GetOwner("Sol"); owner != "crimson" {
		t.Errorf("in-memory flip rolled back on persister error: got %q", owner)
	}
}

func TestRestoreOwnershipOverridesSeed(t *testing.T) {
	// Startup sequence: Seed(static), then RestoreOwnership(DB).
	// The persisted row should win for any system it covers.
	m := NewManager(nil)
	m.Seed(testFactions())
	// Persisted history says Crimson captured Sol on a previous run.
	m.RestoreOwnership(map[string]string{"Sol": "crimson"})

	if owner, _ := m.GetOwner("Sol"); owner != "crimson" {
		t.Errorf("restore should override seed: got %q, want crimson", owner)
	}
	// Non-restored systems keep their seed values.
	if owner, _ := m.GetOwner("Alpha Centauri"); owner != "uef" {
		t.Errorf("seed preserved for non-restored: got %q", owner)
	}
}

func TestRestoreOwnershipSkipsUnknownFaction(t *testing.T) {
	// If the persisted row references a faction the current build
	// no longer knows (faction removed / renamed), skip rather
	// than crash.
	m := NewManager(nil)
	m.Seed(testFactions())
	m.RestoreOwnership(map[string]string{"Sol": "ghost_faction"})
	// Sol keeps its seed value.
	if owner, _ := m.GetOwner("Sol"); owner != "uef" {
		t.Errorf("unknown faction should be skipped: got %q, want uef", owner)
	}
}

func TestRestoreOwnershipEmptyIsNoop(t *testing.T) {
	m := NewManager(nil)
	m.Seed(testFactions())
	m.RestoreOwnership(nil)                 // must not panic
	m.RestoreOwnership(map[string]string{}) // also must not panic
	if owner, _ := m.GetOwner("Sol"); owner != "uef" {
		t.Errorf("empty restore should leave seed intact, got %q", owner)
	}
}

func TestRestoreOwnershipNilManagerSafe(t *testing.T) {
	var m *Manager
	m.RestoreOwnership(map[string]string{"Sol": "uef"}) // must not panic
}

func TestRestoreOwnershipAddsNewSystem(t *testing.T) {
	// A system captured in a previous run but not in anyone's
	// CoreSystems at seed time should still appear as owned after
	// restore. (Happens if the system list evolves.)
	m := NewManager(nil)
	m.Seed(testFactions())
	m.RestoreOwnership(map[string]string{"Vega": "uef"})
	if owner, _ := m.GetOwner("Vega"); owner != "uef" {
		t.Errorf("unseeded system should be addable via restore: got %q", owner)
	}
}

func TestConcurrentReadsDuringTransfer(t *testing.T) {
	// Race test — run with -race to catch lock misuse.
	m := NewManager(nil)
	m.Seed(testFactions())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = m.TransferSystem("Sol", "crimson")
	}()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.GetOwner("Sol")
			_ = m.AllOwnership()
			_ = m.GetFactionSystems("uef")
		}()
	}
	wg.Wait()
}
