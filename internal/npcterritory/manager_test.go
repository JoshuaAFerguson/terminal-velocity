// File: internal/npcterritory/manager_test.go
// Project: Terminal Velocity
// Description: Tests for NPC system ownership. Covers seed,
//   lookup, transfer/no-op/unknown, war resolution, news emission.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package npcterritory

import (
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
