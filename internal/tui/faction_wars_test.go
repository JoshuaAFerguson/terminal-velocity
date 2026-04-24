// File: internal/tui/faction_wars_test.go
// Project: Terminal Velocity
// Description: Unit tests for faction war TUI pure helpers —
//   status markers, duration formatting, short-name truncation,
//   and the two-column zone layout.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

func TestWarStatusMarker(t *testing.T) {
	tests := map[models.FactionWarStatus]string{
		models.FactionWarActive:                "⚔",
		models.FactionWarResolved:              "✓",
		models.FactionWarCeased:                "⏸",
		models.FactionWarStatus("nonexistent"): "?",
	}
	for status, want := range tests {
		if got := warStatusMarker(status); got != want {
			t.Errorf("warStatusMarker(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestWarStatusLabel(t *testing.T) {
	tests := []struct {
		status models.FactionWarStatus
		want   string
	}{
		{models.FactionWarActive, "ACTIVE"},
		{models.FactionWarResolved, "RESOLVED"},
		{models.FactionWarCeased, "CEASEFIRE"},
		{models.FactionWarStatus("unknown"), "unknown"},
	}
	for _, tc := range tests {
		w := &models.FactionWar{Status: tc.status}
		if got := warStatusLabel(w); got != tc.want {
			t.Errorf("warStatusLabel(%q) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestFormatWarDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero is just now", 0, "just now"},
		{"negative is just now", -5 * time.Minute, "just now"},
		{"sub-minute is <1m", 30 * time.Second, "<1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"59 minutes", 59 * time.Minute, "59m"},
		{"1 hour exactly shows 0m suffix", time.Hour, "1h 0m"},
		{"3h 20m", 3*time.Hour + 20*time.Minute, "3h 20m"},
		{"1 day exactly", 24 * time.Hour, "1d 0h"},
		{"3d 4h", 3*24*time.Hour + 4*time.Hour, "3d 4h"},
		// Edge: 23h 59m should stay in hours format, not flip to 0d.
		{"23h 59m stays hours", 23*time.Hour + 59*time.Minute, "23h 59m"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := formatWarDuration(tc.d); got != tc.want {
				t.Errorf("formatWarDuration(%v) = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

func TestShortName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"UEF", "UEF"},
		{"Up to twenty char!!!", "Up to twenty char!!!"},  // exactly 20
		{"Twenty-one characters", "Twenty-one charac..."}, // truncated to 17+...
		{"A much longer faction name that definitely overflows", "A much longer fac..."},
	}
	for _, tc := range tests {
		if got := shortName(tc.in); got != tc.want {
			t.Errorf("shortName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTwoColumnList(t *testing.T) {
	t.Run("empty input returns nil", func(t *testing.T) {
		if got := twoColumnList(nil, 40); got != nil {
			t.Fatalf("nil input should return nil, got %v", got)
		}
	})

	t.Run("narrow width falls back to single column", func(t *testing.T) {
		got := twoColumnList([]string{"one", "two", "three"}, 10)
		if len(got) != 3 {
			t.Fatalf("narrow layout: expected 3 rows (single col), got %d", len(got))
		}
	})

	t.Run("wide width produces half-sized rows", func(t *testing.T) {
		items := []string{"one", "two", "three", "four", "five"}
		got := twoColumnList(items, 40)
		// 5 items / 2 = 3 rows (ceil).
		if len(got) != 3 {
			t.Fatalf("expected 3 rows, got %d: %v", len(got), got)
		}
		// Row 0 should contain both "one" and "four" (left col first).
		if !strings.Contains(got[0], "one") || !strings.Contains(got[0], "four") {
			t.Errorf("row 0 should pair one + four, got %q", got[0])
		}
		// Row 2 has no right-column entry — should still render.
		if !strings.Contains(got[2], "three") {
			t.Errorf("row 2 should include three, got %q", got[2])
		}
	})

	t.Run("odd-count last row leaves right column empty", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		got := twoColumnList(items, 40)
		// 3 items / 2 = 2 rows.
		if len(got) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(got))
		}
	})
}

func TestSideBySide(t *testing.T) {
	t.Run("equal-height panels", func(t *testing.T) {
		left := "A\nB\nC"
		right := "1\n2\n3"
		got := sideBySide(left, right, 5, 5)
		lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
		}
		for i, line := range lines {
			if !strings.Contains(line, string(rune('A'+i))) {
				t.Errorf("line %d missing left content: %q", i, line)
			}
			if !strings.Contains(line, string(rune('1'+i))) {
				t.Errorf("line %d missing right content: %q", i, line)
			}
		}
	})

	t.Run("unequal heights pad the shorter side", func(t *testing.T) {
		left := "A\nB\nC\nD"
		right := "1\n2"
		got := sideBySide(left, right, 5, 5)
		lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
		if len(lines) != 4 {
			t.Fatalf("should match taller panel height (4), got %d", len(lines))
		}
		// Row 3 has left content but no right content — should still render.
		if !strings.Contains(lines[3], "D") {
			t.Errorf("row 3 missing left-only content: %q", lines[3])
		}
	})
}

func TestMaxZero(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{5, 5},
		{0, 0},
		{-1, 0},
		{-100, 0},
	}
	for _, tc := range tests {
		if got := maxZero(tc.in); got != tc.want {
			t.Errorf("maxZero(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestRenderWarDetailNilSafe(t *testing.T) {
	// nil war should produce a placeholder, not panic.
	got := renderWarDetail(nil, time.Now(), 40)
	if !strings.Contains(got, "No war selected") {
		t.Errorf("nil war: got %q, want placeholder", got)
	}
}

func TestRenderWarDetailActiveWar(t *testing.T) {
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	w := &models.FactionWar{
		AggressorName:  "Alpha Republic",
		DefenderName:   "Crimson Pact",
		Status:         models.FactionWarActive,
		DeclaredAt:     now.Add(-48 * time.Hour),
		CasusBelli:     "border incident in Sol",
		WarZoneSystems: []string{"Sol", "Alpha Centauri", "Wolf 359"},
	}
	got := renderWarDetail(w, now, 40)
	for _, want := range []string{"Alpha Republic", "Crimson Pact", "ACTIVE", "border incident", "Sol", "Alpha Centauri", "Wolf 359"} {
		if !strings.Contains(got, want) {
			t.Errorf("detail missing %q: %q", want, got)
		}
	}
	// Duration should be 2 days.
	if !strings.Contains(got, "2d") {
		t.Errorf("duration should show 2d: %q", got)
	}
	// No "Ended" or "Winner" line for active wars.
	if strings.Contains(got, "Ended") {
		t.Errorf("active war should not show Ended line: %q", got)
	}
	if strings.Contains(got, "Winner") {
		t.Errorf("active war should not show Winner line: %q", got)
	}
}

func TestRenderWarDetailResolvedWarShowsWinner(t *testing.T) {
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	resolved := now.Add(-1 * time.Hour)
	w := &models.FactionWar{
		AggressorID:     "alpha",
		AggressorName:   "Alpha Republic",
		DefenderID:      "crimson",
		DefenderName:    "Crimson Pact",
		Status:          models.FactionWarResolved,
		DeclaredAt:      now.Add(-48 * time.Hour),
		ResolvedAt:      &resolved,
		WinnerFactionID: "crimson", // defender wins
		WarZoneSystems:  []string{"Sol"},
	}
	got := renderWarDetail(w, now, 40)
	if !strings.Contains(got, "RESOLVED") {
		t.Errorf("status label missing: %q", got)
	}
	if !strings.Contains(got, "Winner:") || !strings.Contains(got, "Crimson Pact") {
		t.Errorf("defender should be winner: %q", got)
	}
	if !strings.Contains(got, "Ended:") {
		t.Errorf("resolved war should show Ended: %q", got)
	}
}

func TestRenderWarListCursor(t *testing.T) {
	wars := []*models.FactionWar{
		{AggressorName: "A Republic", DefenderName: "B Pact", Status: models.FactionWarActive},
		{AggressorName: "C Alliance", DefenderName: "D Empire", Status: models.FactionWarResolved},
	}
	// Cursor at 1 — second row gets the > marker.
	got := renderWarList(wars, 1, 40)
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
	// Active war shows ⚔; resolved shows ✓.
	if !strings.Contains(lines[0], "⚔") {
		t.Errorf("row 0 (active) missing marker: %q", lines[0])
	}
	if !strings.Contains(lines[1], "✓") {
		t.Errorf("row 1 (resolved) missing marker: %q", lines[1])
	}
}

// TestWarZoneBannerNilSafe verifies the banner returns empty on the
// various nil paths (no manager, no player, no system) rather than
// panicking. The space-view integration relies on the empty string
// to suppress the banner row.
func TestWarZoneBannerNilSafe(t *testing.T) {
	// Empty Model — everything is nil.
	m := Model{}
	if got := m.warZoneBanner(); got != "" {
		t.Errorf("nil-everything should return empty, got %q", got)
	}
}

// TestRenderWarDetailPreview dumps a full rendered detail panel so
// reviewers can eyeball layout with `go test -v`.
func TestRenderWarDetailPreview(t *testing.T) {
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	resolved := now.Add(-2 * time.Hour)

	active := &models.FactionWar{
		AggressorName:  "United Earth Federation",
		DefenderName:   "Crimson Collective",
		Status:         models.FactionWarActive,
		DeclaredAt:     now.Add(-3*24*time.Hour - 4*time.Hour),
		CasusBelli:     "Piracy escalation along the Sol-Barnard trade corridor",
		WarZoneSystems: []string{"Sol", "Alpha Centauri", "Wolf 359", "Barnard", "Procyon", "Altair"},
	}
	resolvedWar := &models.FactionWar{
		AggressorID:     "uef",
		AggressorName:   "United Earth Federation",
		DefenderID:      "crimson",
		DefenderName:    "Crimson Collective",
		Status:          models.FactionWarResolved,
		DeclaredAt:      now.Add(-14*24*time.Hour - 5*time.Hour),
		ResolvedAt:      &resolved,
		WinnerFactionID: "uef",
		CasusBelli:      "Border incident",
		WarZoneSystems:  []string{"Sol", "Barnard"},
	}

	t.Logf("ACTIVE war detail (width=40):\n%s\n", renderWarDetail(active, now, 40))
	t.Logf("RESOLVED war detail (width=40):\n%s\n", renderWarDetail(resolvedWar, now, 40))
}
