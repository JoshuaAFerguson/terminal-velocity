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

// PadRight pads a string to the right
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadLeft pads a string to the left
func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// Center centers a string within a given width
func Center(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// TruncateString truncates a string to a maximum length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
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
