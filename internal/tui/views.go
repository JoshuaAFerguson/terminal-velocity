package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Style definitions
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // Cyan
			MarginTop(1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			MarginTop(1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")) // White

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")). // Cyan
				Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")). // Cyan
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")). // Yellow
			Bold(true)
)

// loadingView shows a loading screen
func loadingView() string {
	return titleStyle.Render("Terminal Velocity") + "\n\n" +
		"Loading player data...\n"
}

// errorView shows an error message
func errorView(message string) string {
	return titleStyle.Render("Terminal Velocity") + "\n\n" +
		errorStyle.Render("Error: "+message) + "\n\n" +
		helpStyle.Render("Press Ctrl+C to exit")
}

// renderHeader renders the game header with player stats
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

// renderFooter renders the help text footer
func renderFooter(helpText string) string {
	return "\n" + helpStyle.Render(helpText)
}

// renderBox renders content in a box
func renderBox(title, content string) string {
	titleLine := titleStyle.Render(title)
	return titleLine + "\n" + boxStyle.Render(content)
}
