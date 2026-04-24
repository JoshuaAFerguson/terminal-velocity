// File: internal/tui/territory_map_test.go
// Project: Terminal Velocity
// Description: Unit tests for the territory-map pure helpers.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"strings"
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/npcterritory"
)

// territoryFixtureModel returns a Model with a fully seeded
// npcterritory manager. testFixtureFactions is separate from
// production StandardNPCFactions so unit tests don't drift when
// game data rebalances.
func territoryFixtureModel() Model {
	mgr := npcterritory.NewManager(nil)
	mgr.Seed([]models.NPCFaction{
		{
			ID:          "uef",
			Name:        "United Earth Federation",
			ShortName:   "UEF",
			CoreSystems: []string{"Sol", "Alpha Centauri", "Tau Ceti"},
		},
		{
			ID:          "crimson",
			Name:        "Crimson Collective",
			ShortName:   "CRM",
			CoreSystems: []string{"Wolf 359", "Barnard"},
		},
	})
	return Model{npcTerritoryManager: mgr}
}

func TestTerritoryGroupsSortedAlphabetical(t *testing.T) {
	m := territoryFixtureModel()
	groups := m.territoryGroupsSorted()

	if len(groups) != 2 {
		t.Fatalf("expected 2 factions, got %d", len(groups))
	}
	// Alphabetical by display name: Crimson Collective < UEF.
	if groups[0].name != "Crimson Collective" {
		t.Errorf("first group: got %q, want Crimson Collective", groups[0].name)
	}
	if groups[1].name != "United Earth Federation" {
		t.Errorf("second group: got %q, want UEF", groups[1].name)
	}
	// Each group's systems list is sorted alphabetically.
	want := []string{"Alpha Centauri", "Sol", "Tau Ceti"}
	if len(groups[1].systems) != len(want) {
		t.Fatalf("UEF system count: got %d, want %d", len(groups[1].systems), len(want))
	}
	for i, s := range want {
		if groups[1].systems[i] != s {
			t.Errorf("UEF[%d]: got %q, want %q", i, groups[1].systems[i], s)
		}
	}
}

func TestTerritoryGroupsNilManager(t *testing.T) {
	m := Model{}
	if got := m.territoryGroupsSorted(); got != nil {
		t.Errorf("nil manager: got %v, want nil", got)
	}
}

func TestRenderTerritoryListCursor(t *testing.T) {
	groups := []territoryGroup{
		{factionID: "a", name: "Alpha", systems: []string{"Sol"}},
		{factionID: "b", name: "Beta", systems: []string{"Wolf 359", "Barnard"}},
	}
	got := renderTerritoryList(groups, 1, 30)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(lines))
	}
	if strings.HasPrefix(lines[0], "> ") {
		t.Errorf("row 0 should not have cursor marker: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "> ") {
		t.Errorf("row 1 should have cursor marker: %q", lines[1])
	}
	// Each row shows system count in parens.
	if !strings.Contains(lines[0], "(1)") {
		t.Errorf("row 0 should show Alpha (1): %q", lines[0])
	}
	if !strings.Contains(lines[1], "(2)") {
		t.Errorf("row 1 should show Beta (2): %q", lines[1])
	}
}

func TestRenderTerritoryDetailNilGroup(t *testing.T) {
	got := renderTerritoryDetail(nil, 40)
	if !strings.Contains(got, "No faction selected") {
		t.Errorf("nil group: got %q, want placeholder", got)
	}
}

func TestRenderTerritoryDetailShowsSystems(t *testing.T) {
	g := &territoryGroup{
		factionID: "uef",
		name:      "United Earth Federation",
		systems:   []string{"Alpha Centauri", "Sol", "Tau Ceti"},
	}
	got := renderTerritoryDetail(g, 40)
	for _, want := range []string{"United Earth Federation", "3 systems", "Sol", "Tau Ceti", "Alpha Centauri"} {
		if !strings.Contains(got, want) {
			t.Errorf("detail missing %q: %q", want, got)
		}
	}
}

func TestRenderTerritoryDetailEmptyHoldings(t *testing.T) {
	g := &territoryGroup{
		factionID: "collapsed",
		name:      "Lost Faction",
		systems:   nil,
	}
	got := renderTerritoryDetail(g, 40)
	if !strings.Contains(got, "no holdings") {
		t.Errorf("empty-holdings placeholder: %q", got)
	}
	if !strings.Contains(got, "0 systems") {
		t.Errorf("zero count label: %q", got)
	}
}

// TestTerritoryMapPreview dumps a rendered view so `go test -v`
// shows the layout without needing a running server.
func TestTerritoryMapPreview(t *testing.T) {
	m := territoryFixtureModel()
	m.width = 80
	t.Logf("territory map (cursor=0):\n%s\n", m.viewTerritoryMap())
	m.territoryMap.cursor = 1
	t.Logf("territory map (cursor=1):\n%s\n", m.viewTerritoryMap())
}
