// File: internal/tui/ui_components.go
// Project: Terminal Velocity
// Description: UI components and helpers for Escape Velocity-style interface
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Box-drawing characters for borders
const (
	BoxTopLeft     = "┏"
	BoxTopRight    = "┓"
	BoxBottomLeft  = "┗"
	BoxBottomRight = "┛"
	BoxHorizontal  = "━"
	BoxVertical    = "┃"
	BoxCross       = "┫"
	BoxCrossLeft   = "┣"

	// Double-line box for inner panels
	BoxTopLeftDouble     = "╔"
	BoxTopRightDouble    = "╗"
	BoxBottomLeftDouble  = "╚"
	BoxBottomRightDouble = "╝"
	BoxHorizontalDouble  = "═"
	BoxVerticalDouble    = "║"

	// Progress bar characters
	ProgressFull  = "█"
	ProgressEmpty = "░"

	// Icons
	IconShip   = "△"
	IconPlanet = "⊕"
	IconEnemy  = "◆"
	IconStar   = "*"
	IconSystem = "◉"
	IconPlayer = "▲"
	IconCheck  = "✓"
	IconBullet = "▪"
	IconArrow  = "▶"
)

// DrawBox draws a box with the given title and content
func DrawBox(title, content string, width, height int) string {
	if width < 10 || height < 3 {
		return content
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeft)
	if title != "" {
		titleWidth := len(title) + 2
		if titleWidth < width-2 {
			sb.WriteString(" " + title + " ")
			sb.WriteString(strings.Repeat(BoxHorizontal, width-titleWidth-2))
		} else {
			sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
		}
	} else {
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	}
	sb.WriteString(BoxTopRight + "\n")

	// Content lines
	lines := strings.Split(content, "\n")
	contentHeight := height - 2

	for i := 0; i < contentHeight; i++ {
		sb.WriteString(BoxVertical)
		if i < len(lines) {
			line := lines[i]
			// Pad or trim line to fit width
			if len(line) > width-2 {
				sb.WriteString(line[:width-2])
			} else {
				sb.WriteString(line + strings.Repeat(" ", width-2-len(line)))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", width-2))
		}
		sb.WriteString(BoxVertical + "\n")
	}

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// DrawBoxDouble draws a box with double-line borders
func DrawBoxDouble(title, content string, width, height int) string {
	if width < 10 || height < 3 {
		return content
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeftDouble)
	if title != "" {
		titleWidth := len(title) + 2
		if titleWidth < width-2 {
			sb.WriteString(" " + title + " ")
			sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-titleWidth-2))
		} else {
			sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-2))
		}
	} else {
		sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-2))
	}
	sb.WriteString(BoxTopRightDouble + "\n")

	// Content lines
	lines := strings.Split(content, "\n")
	contentHeight := height - 2

	for i := 0; i < contentHeight; i++ {
		sb.WriteString(BoxVerticalDouble)
		if i < len(lines) {
			line := lines[i]
			// Pad or trim line to fit width
			if len(line) > width-2 {
				sb.WriteString(line[:width-2])
			} else {
				sb.WriteString(line + strings.Repeat(" ", width-2-len(line)))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", width-2))
		}
		sb.WriteString(BoxVerticalDouble + "\n")
	}

	// Bottom border
	sb.WriteString(BoxBottomLeftDouble)
	sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-2))
	sb.WriteString(BoxBottomRightDouble)

	return sb.String()
}

// DrawProgressBar creates a progress bar
func DrawProgressBar(current, max, width int) string {
	if max == 0 || width < 3 {
		return strings.Repeat(ProgressEmpty, width)
	}

	percentage := float64(current) / float64(max)
	if percentage > 1.0 {
		percentage = 1.0
	}
	if percentage < 0.0 {
		percentage = 0.0
	}

	filled := int(float64(width) * percentage)
	empty := width - filled

	return strings.Repeat(ProgressFull, filled) + strings.Repeat(ProgressEmpty, empty)
}

// DrawHeader creates a screen header with title and credits
func DrawHeader(title, subtitle string, credits int64, shield int, width int) string {
	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxTopRight + "\n")

	// Title line
	sb.WriteString(BoxVertical)

	// Left: Title
	leftText := " " + title
	if subtitle != "" {
		leftText += "  [" + subtitle + "]"
	}

	// Right: Shields and/or credits
	var rightText string
	if shield >= 0 {
		shieldBar := DrawProgressBar(shield, 100, 10)
		rightText = fmt.Sprintf("Shields: %s %d%%", shieldBar, shield)
	}
	if credits >= 0 {
		if rightText != "" {
			rightText += " "
		}
		rightText += fmt.Sprintf("%s", FormatCredits(credits))
	}
	rightText += " "

	// Calculate spacing
	totalText := len(leftText) + len(rightText)
	spacing := width - 2 - totalText
	if spacing < 1 {
		spacing = 1
	}

	sb.WriteString(leftText)
	sb.WriteString(strings.Repeat(" ", spacing))
	sb.WriteString(rightText)
	sb.WriteString(BoxVertical + "\n")

	// Separator
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross)

	return sb.String()
}

// DrawFooter creates a screen footer with command hints
func DrawFooter(commands string, width int) string {
	var sb strings.Builder

	// Separator
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Commands line
	sb.WriteString(BoxVertical)
	sb.WriteString(" " + commands)
	sb.WriteString(strings.Repeat(" ", width-len(commands)-3))
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// FormatCredits formats credits with thousand separators
func FormatCredits(credits int64) string {
	if credits < 0 {
		return fmt.Sprintf("-%s cr", formatNumber(-credits))
	}
	return fmt.Sprintf("%s credits", formatNumber(credits))
}

// formatNumber adds thousand separators
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	var result []rune
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}

// Pad/truncate helpers work in terminal cells, not bytes.
//
// All TUI layout uses multi-byte Unicode (box-drawing chars are 3 bytes / 1 cell)
// and some styled content contains ANSI escape sequences (invisible, 0 cells).
// Measuring with Go's len() or slicing with s[:n] is a bug trap: it counts
// bytes, so pads are too short and truncations land in the middle of a
// codepoint — which is why renders showed replacement chars and broken borders.
//
// cellWidth returns the on-screen width of s in terminal cells, stripping ANSI
// escapes (via lipgloss.Width). truncateCells returns the longest prefix of s
// that fits within maxCells, never splitting a rune.

// cellWidth reports the visible cell width of s. ANSI escape sequences count
// as zero cells; East Asian wide runes count as two.
func cellWidth(s string) int {
	return lipgloss.Width(s)
}

// truncateCells returns the longest prefix of s whose cell width is <= maxCells,
// preserving rune boundaries. ANSI escape sequences encountered in the prefix
// pass through unchanged. If s has no ANSI and fits entirely, s is returned
// without copying.
func truncateCells(s string, maxCells int) string {
	if maxCells <= 0 {
		return ""
	}
	if cellWidth(s) <= maxCells {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	used := 0
	inEscape := false
	for _, r := range s {
		if inEscape {
			b.WriteRune(r)
			// terminator of CSI/SGR: any letter @A-Za-z
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '~' {
				inEscape = false
			}
			continue
		}
		if r == 0x1b {
			inEscape = true
			b.WriteRune(r)
			continue
		}
		w := runewidth.RuneWidth(r)
		if used+w > maxCells {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String()
}

// PadRight right-pads s with spaces to occupy exactly width terminal cells.
// If s is wider than width, it is truncated on rune boundaries.
func PadRight(s string, width int) string {
	w := cellWidth(s)
	if w >= width {
		return truncateCells(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

// PadLeft left-pads s with spaces to occupy exactly width terminal cells.
// If s is wider than width, it is truncated on rune boundaries.
func PadLeft(s string, width int) string {
	w := cellWidth(s)
	if w >= width {
		return truncateCells(s, width)
	}
	return strings.Repeat(" ", width-w) + s
}

// Center centers s within width terminal cells, padding with spaces on both
// sides. If s is wider than width, it is truncated on rune boundaries.
func Center(s string, width int) string {
	w := cellWidth(s)
	if w >= width {
		return truncateCells(s, width)
	}
	leftPad := (width - w) / 2
	rightPad := width - w - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// TruncateString shortens s to at most maxLen terminal cells, appending "..."
// when truncation occurred. Operates on runes, never splitting a codepoint.
func TruncateString(s string, maxLen int) string {
	if cellWidth(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return truncateCells(s, maxLen)
	}
	return truncateCells(s, maxLen-3) + "..."
}

// Color styles using lipgloss
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")).
			Background(lipgloss.Color("0"))
)

// DrawPanel creates a simple panel with optional title
func DrawPanel(title, content string, width, height int, titleBar bool) string {
	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeft)
	if titleBar && title != "" {
		titleWidth := len(title) + 2
		if titleWidth < width-2 {
			sb.WriteString(" " + title + " ")
			sb.WriteString(strings.Repeat(BoxHorizontal, width-titleWidth-2))
		} else {
			sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
		}
	} else {
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	}
	sb.WriteString(BoxTopRight + "\n")

	// Optional title separator
	if titleBar && title != "" {
		sb.WriteString(BoxCrossLeft)
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
		sb.WriteString(BoxCross + "\n")
	}

	// Content
	lines := strings.Split(content, "\n")
	contentHeight := height - 2
	if titleBar && title != "" {
		contentHeight-- // Account for title separator
	}

	for i := 0; i < contentHeight; i++ {
		sb.WriteString(BoxVertical)
		if i < len(lines) {
			line := lines[i]
			if len(line) > width-2 {
				sb.WriteString(line[:width-2])
			} else {
				sb.WriteString(line + strings.Repeat(" ", width-2-len(line)))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", width-2))
		}
		sb.WriteString(BoxVertical + "\n")
	}

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}
