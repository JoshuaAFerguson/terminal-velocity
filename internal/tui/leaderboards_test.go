// File: internal/tui/leaderboards_test.go
// Project: Terminal Velocity
// Description: Unit tests for leaderboard category tab helpers and
//   the underline-style tab renderer.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tui

import (
	"strings"
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/lipgloss"
)

func TestLeaderboardCategoryIndex(t *testing.T) {
	tests := []struct {
		name string
		cat  models.LeaderboardCategory
		want int
	}{
		{"overall is first", models.LeaderboardOverall, 0},
		{"combat is second", models.LeaderboardCombat, 1},
		{"reputation is last", models.LeaderboardReputation, 6},
		{"unknown category returns -1", models.LeaderboardCategory("nonexistent"), -1},
		{"empty string returns -1", models.LeaderboardCategory(""), -1},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := leaderboardCategoryIndex(tc.cat); got != tc.want {
				t.Fatalf("leaderboardCategoryIndex(%q) = %d, want %d", tc.cat, got, tc.want)
			}
		})
	}
}

func TestLeaderboardCategoryByDigit(t *testing.T) {
	tests := []struct {
		name    string
		digit   string
		wantCat models.LeaderboardCategory
		wantOK  bool
	}{
		{"1 → Overall", "1", models.LeaderboardOverall, true},
		{"2 → Combat", "2", models.LeaderboardCombat, true},
		{"3 → Trading", "3", models.LeaderboardTrading, true},
		{"4 → Exploration", "4", models.LeaderboardExploration, true},
		{"5 → Wealth", "5", models.LeaderboardWealth, true},
		{"6 → Missions", "6", models.LeaderboardMissions, true},
		{"7 → Reputation", "7", models.LeaderboardReputation, true},
		{"8 is out of range", "8", "", false},
		{"0 is out of range", "0", "", false},
		{"letter is ignored", "a", "", false},
		{"empty string is ignored", "", "", false},
		{"shift+1 (!) is ignored", "!", "", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := leaderboardCategoryByDigit(tc.digit)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.wantCat {
				t.Fatalf("category = %q, want %q", got, tc.wantCat)
			}
		})
	}
}

func TestCycleLeaderboardCategory(t *testing.T) {
	tests := []struct {
		name  string
		cur   models.LeaderboardCategory
		delta int
		want  models.LeaderboardCategory
	}{
		// Right-cycle.
		{"right from Overall → Combat", models.LeaderboardOverall, 1, models.LeaderboardCombat},
		{"right from Combat → Trading", models.LeaderboardCombat, 1, models.LeaderboardTrading},
		{"right from Reputation wraps → Overall", models.LeaderboardReputation, 1, models.LeaderboardOverall},

		// Left-cycle.
		{"left from Combat → Overall", models.LeaderboardCombat, -1, models.LeaderboardOverall},
		{"left from Overall wraps → Reputation", models.LeaderboardOverall, -1, models.LeaderboardReputation},
		{"left from Trading → Combat", models.LeaderboardTrading, -1, models.LeaderboardCombat},

		// Unknown starting category falls back to first.
		{"unknown category falls back to first", models.LeaderboardCategory("garbage"), 1, models.LeaderboardOverall},

		// Zero delta returns current (or falls back if unknown).
		{"delta 0 is identity for known", models.LeaderboardWealth, 0, models.LeaderboardWealth},

		// Large deltas modulate correctly.
		{"delta 7 (full lap) returns current", models.LeaderboardCombat, 7, models.LeaderboardCombat},
		{"delta 8 is one step forward", models.LeaderboardCombat, 8, models.LeaderboardTrading},
		{"delta -7 (reverse full lap) returns current", models.LeaderboardCombat, -7, models.LeaderboardCombat},
		{"delta -8 is one step back", models.LeaderboardCombat, -8, models.LeaderboardOverall},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := cycleLeaderboardCategory(tc.cur, tc.delta)
			if got != tc.want {
				t.Fatalf("cycle(%q, %d) = %q, want %q", tc.cur, tc.delta, got, tc.want)
			}
		})
	}
}

// TestRenderLeaderboardTabsShape checks the rendered tab bar has two
// lines, and the underline line under the active tab has non-zero
// visual width matching the active tab cell.
func TestRenderLeaderboardTabsShape(t *testing.T) {
	out := renderLeaderboardTabs(models.LeaderboardCombat)

	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 lines (labels + underline), got %d: %q", len(lines), out)
	}

	labelsWidth := lipgloss.Width(lines[0])
	underlineWidth := lipgloss.Width(lines[1])

	// Both lines must have equal visual width so the underline aligns
	// with the labels it marks.
	if labelsWidth != underlineWidth {
		t.Fatalf("labels width %d != underline width %d; underline cannot align", labelsWidth, underlineWidth)
	}

	// Underline must contain at least one ━ (the active-tab marker).
	if !strings.ContainsRune(lines[1], '━') {
		t.Fatalf("underline line missing heavy-underline char: %q", lines[1])
	}

	// Labels line must mention every tab by its digit shortcut.
	for _, tab := range leaderboardCategoryTabs {
		if !strings.Contains(lines[0], "("+tab.key+")") {
			t.Fatalf("labels line missing digit shortcut (%s) for %s: %q", tab.key, tab.label, lines[0])
		}
	}
}

// TestRenderLeaderboardTabsActiveSwitches checks that rendering with a
// different active category moves the underline to a different column.
// This catches regressions where the renderer always underlines the
// same tab (e.g. hardcoded index 0).
func TestRenderLeaderboardTabsActiveSwitches(t *testing.T) {
	out0 := renderLeaderboardTabs(models.LeaderboardOverall)
	out1 := renderLeaderboardTabs(models.LeaderboardCombat)

	lines0 := strings.Split(out0, "\n")
	lines1 := strings.Split(out1, "\n")
	if len(lines0) != 2 || len(lines1) != 2 {
		t.Fatalf("expected 2 lines each")
	}

	// The underline rows must differ when the active tab changes.
	// (The label rows also differ because of style codes, but the
	// underline row is the invariant the player sees.)
	if lines0[1] == lines1[1] {
		t.Fatalf("underline row did not change between active tabs; renderer may be hardcoding active index")
	}
}

// TestRenderLeaderboardTabsPreview is a visual smoke test — it prints
// the rendered output when run with `go test -v`, so a maintainer can
// eyeball the tab strip without spinning up the SSH server. Makes no
// assertions beyond the others; purely for reviewability.
func TestRenderLeaderboardTabsPreview(t *testing.T) {
	for _, cat := range []models.LeaderboardCategory{
		models.LeaderboardOverall,
		models.LeaderboardCombat,
		models.LeaderboardReputation,
	} {
		t.Logf("active=%s:\n%s\n\n", cat, renderLeaderboardTabs(cat))
	}
}
