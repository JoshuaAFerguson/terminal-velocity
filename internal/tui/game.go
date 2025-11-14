// File: internal/tui/game.go
// Project: Terminal Velocity
// Description: Terminal UI component for game
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import tea "github.com/charmbracelet/bubbletea"

type gameViewModel struct {
	// Game state will go here
}

func (m Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.screen = ScreenMainMenu
			return m, nil
		case "r":
			m.screen = ScreenTradeRoutes
			return m, nil
		case "M":
			// Load inbox on entering mail screen
			m.mail.mode = mailModeInbox
			m.mail.selectedIndex = 0
			m.mail.loading = true
			m.screen = ScreenMail
			return m, m.loadInbox()
		}
	}

	return m, nil
}

func (m Model) viewGame() string {
	s := renderHeader(m.username, m.player.Credits, "Space")
	s += "\n"

	content := `You are floating in space.

Your ship's systems hum quietly as you gaze out at the stars.

Commands:
  n - Navigation
  t - Trading
  r - Trade Routes & Nav Planner
  s - Shipyard
  m - Missions
  M - Mail

Press ESC to return to main menu.`

	s += boxStyle.Render(content)

	s += renderFooter("ESC: Main Menu  •  n/t/r/s/m/M: Quick Access")

	return s
}
