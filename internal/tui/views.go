// File: internal/tui/views.go
// Project: Terminal Velocity
// Description: View helper functions and style definitions for common TUI patterns
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// This file provides helper functions for rendering common view patterns:
//   - Loading screens
//   - Error screens
//   - Headers with player stats
//   - Footers with help text
//   - Boxed content
//
// These functions use lipgloss styles defined in this file for consistent appearance.
//
// Usage:
//   - Call from screen view functions (e.g., viewMainMenu, viewCombat)
//   - Use for common UI elements that appear across multiple screens
//   - Styles are configured with margins and colors for terminal display
//
// Note: These are older view helpers. Newer screens use the functions in
// ui_components.go which provide more flexibility and better box-drawing.

package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ===== Lipgloss Style Definitions =====
// Pre-configured styles for common view elements.
//
// These styles use ANSI 256-color palette:
//   - 39: Cyan (titles, highlights, selections)
//   - 46: Green (success messages)
//   - 196: Red (errors)
//   - 228: Yellow (stats)
//   - 241: Gray (help text, subtitles)
//   - 255: White (normal text)

var (
	// titleStyle is used for main titles (bold cyan with margins)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // Cyan
			MarginTop(1).
			MarginBottom(1)

	// subtitleStyle is used for secondary headings (gray with bottom margin)
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			MarginBottom(1)

	// errorStyle is used for error messages (bold red with margins)
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	// helpStyle is used for help text and hints (gray with top margin)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			MarginTop(1)

	// menuItemStyle is used for unselected menu items (white)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")) // White

	// selectedMenuItemStyle is used for selected menu items (bold cyan)
	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")). // Cyan
				Bold(true)

	// boxStyle is used for content boxes (rounded border with padding)
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")). // Cyan
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	// statsStyle is used for player stats (bold yellow)
	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")). // Yellow
			Bold(true)

	// highlightStyle is used for highlighted text (bold cyan)
	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Cyan
			Bold(true)

	// successStyle is used for success messages (green)
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green

	// normalStyle is used for normal text (white)
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")) // White
)

// ===== View Helper Functions =====

// loadingView renders a loading screen.
//
// This is displayed when player data is being loaded asynchronously
// after Init() but before the player data is available.
//
// Returns:
//   - Formatted loading screen string
func loadingView() string {
	return titleStyle.Render("Terminal Velocity") + "\n\n" +
		"Loading player data...\n"
}

// errorView renders an error screen with the given message.
//
// This is displayed when a critical error occurs (e.g., database connection failure,
// player not found, etc.) that prevents the game from continuing.
//
// Parameters:
//   - message: Error message to display to the user
//
// Returns:
//   - Formatted error screen string with quit instructions
func errorView(message string) string {
	return titleStyle.Render("Terminal Velocity") + "\n\n" +
		errorStyle.Render("Error: "+message) + "\n\n" +
		helpStyle.Render("Press Ctrl+C to exit")
}

// renderHeader renders a game header with player stats.
//
// This is used by some screens to show player information at the top.
// Newer screens tend to use DrawHeader from ui_components.go instead.
//
// Parameters:
//   - username: Player's username
//   - credits: Player's credit balance
//   - system: Current star system name
//
// Returns:
//   - Formatted header string with stats
func renderHeader(username string, credits int64, system string) string {
	header := titleStyle.Render("=== Terminal Velocity ===")

	stats := fmt.Sprintf(
		"\nPilot: %s  |  Credits: %s  |  Location: %s",
		statsStyle.Render(username),
		statsStyle.Render(fmt.Sprintf("%d cr", credits)),
		statsStyle.Render(system),
	)

	return header + stats + "\n"
}

// renderFooter renders a help text footer.
//
// This is used to display command hints at the bottom of screens.
// Newer screens tend to use DrawFooter from ui_components.go instead.
//
// Parameters:
//   - helpText: Help text to display (e.g., "Press Q to quit")
//
// Returns:
//   - Formatted footer string
func renderFooter(helpText string) string {
	return "\n" + helpStyle.Render(helpText)
}

// renderBox renders content in a styled box with rounded borders.
//
// This uses lipgloss's built-in rounded border style.
// For more control over box appearance, use DrawBox from ui_components.go.
//
// Parameters:
//   - title: Title to display above the box
//   - content: Content to display inside the box
//
// Returns:
//   - Formatted box string with title
func renderBox(title, content string) string {
	titleLine := titleStyle.Render(title)
	return titleLine + "\n" + boxStyle.Render(content)
}
