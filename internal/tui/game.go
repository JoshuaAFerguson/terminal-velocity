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
		case "n":
			// Navigation — must init the navigation model and kick off a
			// systems load, matching what main_menu does when selecting it.
			m.navigation = newNavigationModel()
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenNavigation
			return m, m.loadConnectedSystems()
		case "t":
			m.trading = newTradingModel()
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenTrading
			return m, m.loadTradingMarket()
		case "s":
			m.shipyard = newShipyardModel()
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenShipyard
			return m, m.loadShipyard()
		case "m":
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenMissions
			return m, nil
		case "r":
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenTradeRoutes
			return m, nil
		case "M":
			// Load inbox on entering mail screen
			m.mail.mode = mailModeInbox
			m.mail.selectedIndex = 0
			m.mail.loading = true
			m.previousScreen = ScreenGame
			m.hasPreviousScreen = true
			m.screen = ScreenMail
			return m, m.loadInbox()
		}
	}

	return m, nil
}

func (m Model) viewGame() string {
	location := "Space"
	if m.currentSystem != nil {
		location = m.currentSystem.Name + " System"
	}
	s := renderHeader(m.username, m.player.Credits, location)
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
