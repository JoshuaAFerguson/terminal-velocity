// File: internal/tui/ui_components.go
// Project: Terminal Velocity
// Description: Reusable UI components and box-drawing utilities for consistent terminal display
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14
//
// This file provides reusable UI components and drawing functions for creating
// a consistent Escape Velocity-style interface throughout the game.
//
// Components include:
//   - Box drawing with single and double-line borders
//   - Progress bars for health, shields, fuel, etc.
//   - Headers and footers for screens
//   - Text formatting utilities
//   - Styled text using lipgloss
//
// Usage:
//   - Use DrawBox() for basic bordered content
//   - Use DrawHeader() for screen headers with stats
//   - Use DrawFooter() for command hints
//   - Use DrawProgressBar() for visual meters
//   - Apply styles with TitleStyle, HighlightStyle, etc.
//
// Thread Safety:
//   - All functions are pure and thread-safe
//   - Styles are immutable lipgloss.Style values

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ===== Box-Drawing Characters =====
// Unicode characters for drawing boxes, borders, and UI elements.
// These provide a retro terminal aesthetic consistent with classic space games.

const (
	// Single-line box borders (primary UI borders)
	BoxTopLeft     = "┏" // Top-left corner
	BoxTopRight    = "┓" // Top-right corner
	BoxBottomLeft  = "┗" // Bottom-left corner
	BoxBottomRight = "┛" // Bottom-right corner
	BoxHorizontal  = "━" // Horizontal line
	BoxVertical    = "┃" // Vertical line
	BoxCross       = "┫" // T-junction (right side)
	BoxCrossLeft   = "┣" // T-junction (left side)

	// Double-line box borders (for emphasis or inner panels)
	BoxTopLeftDouble     = "╔" // Double-line top-left corner
	BoxTopRightDouble    = "╗" // Double-line top-right corner
	BoxBottomLeftDouble  = "╚" // Double-line bottom-left corner
	BoxBottomRightDouble = "╝" // Double-line bottom-right corner
	BoxHorizontalDouble  = "═" // Double-line horizontal
	BoxVerticalDouble    = "║" // Double-line vertical

	// Progress bar characters (for health, shields, fuel meters)
	ProgressFull  = "█" // Filled section of progress bar
	ProgressEmpty = "░" // Empty section of progress bar

	// Icons (used throughout the UI for visual indicators)
	IconShip   = "△" // Generic ship icon
	IconPlanet = "⊕" // Planet icon
	IconEnemy  = "◆" // Enemy ship icon
	IconStar   = "*" // Star icon
	IconSystem = "◉" // Star system icon
	IconPlayer = "▲" // Player ship icon
	IconCheck  = "✓" // Checkmark (for completed items)
	IconBullet = "▪" // Bullet point for lists
	IconArrow  = "▶" // Arrow for selections/navigation
)

// DrawBox draws a box with single-line borders containing the given title and content.
//
// The box is drawn using box-drawing characters (BoxTopLeft, BoxHorizontal, etc.)
// and sized to the specified width and height. The title appears in the top border.
//
// Parameters:
//   - title: Optional title to display in top border (empty string for no title)
//   - content: Content to display inside the box (newline-separated for multiple lines)
//   - width: Total width of the box including borders (minimum 10)
//   - height: Total height of the box including borders (minimum 3)
//
// Returns:
//   - Formatted string representing the box
//   - If width < 10 or height < 3, returns content unmodified
//
// Behavior:
//   - Content lines longer than width-2 are truncated
//   - Content lines shorter than width-2 are padded with spaces
//   - If content has fewer lines than height-2, empty lines are added
//
// Example:
//   box := DrawBox("Player Info", "Name: Alice\nCredits: 1000", 30, 5)
//   // ┏━━━ Player Info ━━━━━━━━━━━┓
//   // ┃ Name: Alice               ┃
//   // ┃ Credits: 1000             ┃
//   // ┃                           ┃
//   // ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
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

// DrawBoxDouble draws a box with double-line borders for emphasis.
//
// Similar to DrawBox() but uses double-line box-drawing characters
// (BoxTopLeftDouble, BoxHorizontalDouble, etc.) for a more prominent appearance.
//
// Use this for:
//   - Important notifications
//   - Error messages
//   - Achievement unlocks
//   - Inner panels within single-line boxes
//
// Parameters and behavior are identical to DrawBox().
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

// DrawProgressBar creates a visual progress bar for meters and gauges.
//
// Progress bars are used throughout the game for:
//   - Hull integrity
//   - Shield strength
//   - Fuel levels
//   - Quest/mission progress
//   - Experience bars
//
// The bar is filled proportionally to current/max ratio using ProgressFull (█)
// and ProgressEmpty (░) characters.
//
// Parameters:
//   - current: Current value (e.g., current hull points)
//   - max: Maximum value (e.g., max hull points)
//   - width: Width of the progress bar in characters
//
// Returns:
//   - String of width characters showing the progress bar
//
// Behavior:
//   - If max == 0 or width < 3: returns all empty bar
//   - Percentage is clamped to 0.0-1.0 range
//   - Fractional fill is rounded down
//
// Example:
//   bar := DrawProgressBar(75, 100, 20)
//   // Returns: "███████████████░░░░░" (15 filled, 5 empty)
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

// DrawHeader creates a standard screen header with title, subtitle, and player stats.
//
// The header displays:
//   - Left side: Screen title and optional subtitle
//   - Right side: Shield bar and/or credits
//
// Used at the top of most game screens for consistent navigation and status display.
//
// Parameters:
//   - title: Main screen title (e.g., "Trading", "Combat", "Navigation")
//   - subtitle: Optional context (e.g., planet name, system name) - empty string for none
//   - credits: Player's credit balance (use -1 to hide)
//   - shield: Player's shield percentage 0-100 (use -1 to hide)
//   - width: Total width of the header
//
// Returns:
//   - Formatted header string with borders and content
//
// Example:
//   header := DrawHeader("Trading", "Rigel IV", 10000, 85, 80)
//   // ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
//   // ┃ Trading  [Rigel IV]                  Shields: █████████░ 85% 10,000 cr ┃
//   // ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
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

// DrawFooter creates a standard screen footer with command hints.
//
// The footer displays available keyboard commands for the current screen,
// providing consistent help text across all screens.
//
// Parameters:
//   - commands: Command hint string (e.g., "[Q] Quit  [Enter] Select  [↑/↓] Navigate")
//   - width: Total width of the footer
//
// Returns:
//   - Formatted footer string with separator and bottom border
//
// Example:
//   footer := DrawFooter("[Q] Quit  [Enter] Select", 80)
//   // ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
//   // ┃ [Q] Quit  [Enter] Select                                                 ┃
//   // ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
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

// FormatCredits formats a credit amount with thousand separators and unit label.
//
// Used throughout the game to display credit amounts in a consistent, readable format.
//
// Parameters:
//   - credits: Amount of credits (positive or negative)
//
// Returns:
//   - Formatted string with commas and "credits" label
//   - Negative amounts prefixed with "-"
//
// Examples:
//   FormatCredits(1000)      // "1,000 credits"
//   FormatCredits(1000000)   // "1,000,000 credits"
//   FormatCredits(-500)      // "-500 cr"
func FormatCredits(credits int64) string {
	if credits < 0 {
		return fmt.Sprintf("-%s cr", formatNumber(-credits))
	}
	return fmt.Sprintf("%s credits", formatNumber(credits))
}

// formatNumber adds thousand separators to a number for readability.
//
// Helper function used by FormatCredits to format large numbers.
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

// PadRight pads a string to the right with spaces.
//
// If the string is longer than width, it is truncated.
//
// Example:
//   PadRight("Hello", 10)  // "Hello     "
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadLeft pads a string to the left with spaces.
//
// If the string is longer than width, it is truncated.
//
// Example:
//   PadLeft("Hello", 10)  // "     Hello"
func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// Center centers a string within a given width.
//
// If the string is longer than width, it is truncated.
// If centering creates an odd number of spaces, the extra space goes on the right.
//
// Example:
//   Center("Hi", 10)  // "    Hi    "
func Center(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// TruncateString truncates a string to a maximum length with ellipsis.
//
// Similar to truncate() in utils.go but exported for use in other packages.
//
// Example:
//   TruncateString("Very long string", 10)  // "Very lo..."
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ===== Lipgloss Styles =====
// Pre-configured lipgloss styles for consistent theming across all screens.
//
// These styles use ANSI color codes and should work in most terminal emulators.
// Color codes reference the 256-color palette:
//   - 8: Gray (muted text)
//   - 9: Red (errors, danger)
//   - 10: Green (success, positive actions)
//   - 11: Yellow (selection, warnings)
//   - 14: Cyan (highlights, links)
//
// Usage:
//   text := TitleStyle.Render("Screen Title")
//   error := ErrorStyle.Render("Error: Something went wrong")

var (
	// TitleStyle is used for screen titles and headings (bold green)
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	// HighlightStyle is used for important information and links (bold cyan)
	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	// ErrorStyle is used for error messages and warnings (bold red)
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	// SuccessStyle is used for success messages and confirmations (bold green)
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	// MutedStyle is used for secondary text and hints (gray)
	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// SelectedStyle is used for selected menu items and list entries (bold yellow on black)
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")).
			Background(lipgloss.Color("0"))
)

// DrawPanel creates a simple panel with optional title bar.
//
// Similar to DrawBox but with more control over the title display.
// If titleBar is true, a separator line is drawn below the title.
//
// Parameters:
//   - title: Optional title to display
//   - content: Content to display inside the panel
//   - width: Total width of the panel including borders
//   - height: Total height of the panel including borders
//   - titleBar: If true, draws a separator after the title
//
// Returns:
//   - Formatted panel string
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
