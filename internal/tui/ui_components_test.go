// File: internal/tui/ui_components_test.go
// Project: Terminal Velocity
// Description: Tests for cell-aware layout helpers (PadRight, PadLeft, Center, TruncateString)
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-22

package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// The regressions these tests pin down:
//
//   - Box-drawing chars ("┃", "━", etc.) are 3 UTF-8 bytes but occupy 1 terminal
//     cell. The old helpers used len(s) which is byte length, so layouts were
//     systematically too short and content was truncated mid-codepoint.
//
//   - ANSI escape sequences (added by lipgloss styles) have zero visual width
//     but consume many bytes. The new helpers must see through them via
//     lipgloss.Width.

func TestPadRightAscii(t *testing.T) {
	got := PadRight("abc", 5)
	if got != "abc  " {
		t.Errorf("PadRight(abc,5) = %q, want %q", got, "abc  ")
	}
}

func TestPadRightUnicodeBoxDrawing(t *testing.T) {
	// "━" is 3 bytes, 1 cell. Three of them = 3 cells, pad to 5.
	in := "━━━"
	got := PadRight(in, 5)
	if w := cellWidth(got); w != 5 {
		t.Errorf("PadRight(%q, 5) produced width %d, want 5 (got %q)", in, w, got)
	}
}

func TestPadRightTruncatesOnRuneBoundary(t *testing.T) {
	in := "━━━━━"
	got := PadRight(in, 3)
	if got != "━━━" {
		t.Errorf("PadRight(%q, 3) = %q, want %q (must not split runes)", in, got, "━━━")
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("result %q contains U+FFFD — truncated mid-codepoint", got)
	}
}

func TestCenterBoxDrawingWithinPanel(t *testing.T) {
	// Matches the real login "OR" separator content.
	in := "─────────────── OR ───────────────────"
	got := Center(in, 48)
	if w := cellWidth(got); w != 48 {
		t.Errorf("Center produced width %d, want 48", w)
	}
	if !strings.Contains(got, "OR") {
		t.Errorf("Center dropped content: %q", got)
	}
}

func TestCenterExactWidthReturnsOriginal(t *testing.T) {
	in := "abcdef"
	got := Center(in, 6)
	if got != in {
		t.Errorf("Center at exact width: got %q, want %q", got, in)
	}
}

func TestCenterSmallerThanContentTruncatesByRune(t *testing.T) {
	in := "━━━━━━━━━━"
	got := Center(in, 4)
	if cellWidth(got) != 4 {
		t.Errorf("Center truncation produced %q with width %d, want 4", got, cellWidth(got))
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("truncation split a rune: %q", got)
	}
}

func TestPadRightWithANSIEscapes(t *testing.T) {
	// Styled text has zero-width ANSI escapes. The layout still needs to know
	// the visible width is 4 cells, not however many bytes the escapes took.
	styled := lipgloss.NewStyle().Bold(true).Render("abcd")
	got := PadRight(styled, 8)
	if w := lipgloss.Width(got); w != 8 {
		t.Errorf("PadRight of styled text: width %d, want 8", w)
	}
}

func TestPadLeftBoxDrawing(t *testing.T) {
	got := PadLeft("┃", 5)
	if w := cellWidth(got); w != 5 {
		t.Errorf("PadLeft width = %d, want 5", w)
	}
	if !strings.HasSuffix(got, "┃") {
		t.Errorf("PadLeft should push content right: %q", got)
	}
}

func TestTruncateStringAppendsEllipsisOnRuneBoundary(t *testing.T) {
	in := "━━━━━━━━━━"
	got := TruncateString(in, 5)
	if cellWidth(got) != 5 {
		t.Errorf("TruncateString width %d, want 5", cellWidth(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ... suffix, got %q", got)
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("truncation broke a rune: %q", got)
	}
}

func TestTruncateStringShortBudget(t *testing.T) {
	in := "abcdef"
	got := TruncateString(in, 3)
	// When budget is too small for ellipsis we take the first 3 cells.
	if cellWidth(got) != 3 {
		t.Errorf("width = %d, want 3", cellWidth(got))
	}
}

func TestTruncateStringShorterThanMaxIsUnchanged(t *testing.T) {
	in := "hi"
	if got := TruncateString(in, 10); got != in {
		t.Errorf("got %q, want %q", got, in)
	}
}

func TestCellWidthIgnoresANSI(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("hello")
	if cellWidth(styled) != 5 {
		t.Errorf("cellWidth of %q = %d, want 5", styled, cellWidth(styled))
	}
}
