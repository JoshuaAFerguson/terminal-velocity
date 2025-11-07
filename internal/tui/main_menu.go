package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
)

type mainMenuModel struct {
	cursor int
	items  []menuItem
}

type menuItem struct {
	label  string
	screen Screen
	action func(*Model) tea.Cmd
}

func newMainMenuModel() mainMenuModel {
	return mainMenuModel{
		cursor: 0,
		items: []menuItem{
			{label: "Launch", screen: ScreenGame},
			{label: "Navigation", screen: ScreenNavigation},
			{label: "Trading", screen: ScreenTrading},
			{label: "Cargo Hold", screen: ScreenCargo},
			{label: "Shipyard", screen: ScreenShipyard},
			{label: "Missions", screen: ScreenMissions},
			{label: "Settings", screen: ScreenSettings},
			{label: "Quit", action: func(m *Model) tea.Cmd { return tea.Quit }},
		},
	}
}

func (m Model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.mainMenu.cursor > 0 {
				m.mainMenu.cursor--
			}
		case "down", "j":
			if m.mainMenu.cursor < len(m.mainMenu.items)-1 {
				m.mainMenu.cursor++
			}
		case "enter", " ":
			selected := m.mainMenu.items[m.mainMenu.cursor]
			if selected.action != nil {
				return m, selected.action(&m)
			}
			m.screen = selected.screen

			// Initialize screen-specific data
			if selected.screen == ScreenNavigation {
				m.navigation = newNavigationModel()
				return m, m.loadConnectedSystems()
			}
			if selected.screen == ScreenTrading {
				m.trading = newTradingModel()
				return m, m.loadTradingMarket()
			}
			if selected.screen == ScreenCargo {
				m.cargo = newCargoModel()
				return m, nil
			}
			if selected.screen == ScreenShipyard {
				m.shipyard = newShipyardModel()
				return m, m.loadShipyard()
			}

			return m, nil
		}
	}

	return m, nil
}

func (m Model) viewMainMenu() string {
	// Get current system name
	systemName := "Unknown"
	if m.player != nil && m.player.CurrentSystem.String() != "00000000-0000-0000-0000-000000000000" {
		// Try to load system name
		// For now, just show "Space"
		systemName = "Space"
	}

	// Render header with player stats
	s := renderHeader(m.username, m.player.Credits, systemName)
	s += "\n"

	// Welcome message
	welcome := fmt.Sprintf("Welcome, Commander %s!", m.username)
	s += subtitleStyle.Render(welcome) + "\n\n"

	// Main menu items
	s += "Main Menu:\n\n"
	for i, item := range m.mainMenu.items {
		if i == m.mainMenu.cursor {
			s += "> " + selectedMenuItemStyle.Render(item.label) + "\n"
		} else {
			s += "  " + menuItemStyle.Render(item.label) + "\n"
		}
	}

	// Help text
	s += renderFooter("↑/↓ or j/k: Navigate  •  Enter: Select  •  q: Quit")

	return s
}
