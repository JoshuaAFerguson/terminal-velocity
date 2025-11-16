// File: internal/tui/main_menu.go
// Project: Terminal Velocity
// Description: Main menu screen - Central navigation hub for accessing all game features
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The main menu serves as the primary navigation interface, providing access to all
// major game systems including navigation, trading, combat, missions, quests, and
// multiplayer features. It displays player stats in the header and presents a
// scrollable list of menu options.

package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// mainMenuModel contains the state for the main menu screen.
// It manages cursor position and the list of available menu items.
type mainMenuModel struct {
	cursor int        // Current cursor position (0-indexed)
	items  []menuItem // List of menu items to display
}

// menuItem represents a single selectable option in the main menu.
// Each item can either navigate to a screen or execute a custom action.
type menuItem struct {
	label  string              // Display text for the menu item
	screen Screen              // Target screen to navigate to (if action is nil)
	action func(*Model) tea.Cmd // Optional custom action (e.g., quit game)
}

// newMainMenuModel creates and initializes a new main menu model.
// Sets up the complete menu with all available screens and actions.
// Returns a mainMenuModel with cursor at position 0.
func newMainMenuModel() mainMenuModel {
	return mainMenuModel{
		cursor: 0,
		items: []menuItem{
			{label: "Launch", screen: ScreenGame},
			{label: "Navigation", screen: ScreenNavigation},
			{label: "Trading", screen: ScreenTrading},
			{label: "Cargo Hold", screen: ScreenCargo},
			{label: "Shipyard", screen: ScreenShipyard},
			{label: "Outfitter", screen: ScreenOutfitter},
			{label: "Advanced Outfitting", screen: ScreenOutfitterEnhanced},
			{label: "Ship Management", screen: ScreenShipManagement},
			{label: "Missions", screen: ScreenMissions},
			{label: "Quests", screen: ScreenQuests},
			{label: "Achievements", screen: ScreenAchievements},
			{label: "Leaderboards", screen: ScreenLeaderboards},
			{label: "Players", screen: ScreenPlayers},
			{label: "Chat", screen: ScreenChat},
			{label: "Factions", screen: ScreenFactions},
			{label: "Trade", screen: ScreenTrade},
			{label: "PvP Combat", screen: ScreenPvP},
			{label: "News", screen: ScreenNews},
			{label: "Help", screen: ScreenHelp},
			{label: "Settings", screen: ScreenSettings},
			{label: "Tutorials", screen: ScreenTutorial},
			{label: "Admin Panel", screen: ScreenAdmin},
			{label: "Quit", action: func(m *Model) tea.Cmd { return tea.Quit }},
		},
	}
}

// updateMainMenu handles input and state updates for the main menu screen.
//
// Key Bindings:
//   - q: Quit the game
//   - up/k: Move cursor up
//   - down/j: Move cursor down
//   - enter/space: Select current menu item
//
// This function routes the player to different screens based on their selection
// and initializes the appropriate screen state (loading data, setting up models).
func (m Model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Quit from main menu
			return m, tea.Quit
		case "up", "k":
			// Move cursor up (vi-style navigation supported with k)
			if m.mainMenu.cursor > 0 {
				m.mainMenu.cursor--
			}
		case "down", "j":
			// Move cursor down (vi-style navigation supported with j)
			if m.mainMenu.cursor < len(m.mainMenu.items)-1 {
				m.mainMenu.cursor++
			}
		case "enter", " ":
			// Select current menu item
			selected := m.mainMenu.items[m.mainMenu.cursor]

			// If item has a custom action (like quit), execute it
			if selected.action != nil {
				return m, selected.action(&m)
			}

			// Otherwise, navigate to the target screen
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
			if selected.screen == ScreenOutfitter {
				m.outfitter = newOutfitterModel()
				return m, m.loadOutfitter()
			}
			if selected.screen == ScreenOutfitterEnhanced {
				m.outfitterEnhanced = newOutfitterEnhancedModel()
				// Load player inventory and loadouts
				m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
				m.outfitterEnhanced.loadouts = m.outfittingManager.GetPlayerLoadouts(m.playerID)
				return m, nil
			}
			if selected.screen == ScreenShipManagement {
				m.shipManagement = newShipManagementModel()
				return m, m.loadOwnedShips()
			}
			if selected.screen == ScreenLeaderboards {
				m.leaderboardsModel = newLeaderboardsModel()
				return m, m.refreshLeaderboards()
			}
			if selected.screen == ScreenSettings {
				m.settingsModel = newSettingsModel()
				// Load player settings
				if playerSettings, err := m.settingsManager.LoadSettings(m.playerID); err == nil {
					m.settingsModel.settings = playerSettings
				}
				return m, nil
			}
			if selected.screen == ScreenAdmin {
				m.adminModel = newAdminModel()
				// Check if player is admin
				m.adminModel.isAdmin = m.adminManager.IsAdmin(m.playerID)
				if m.adminModel.isAdmin {
					// Get admin role from manager
					// For now, default to moderator
					m.adminModel.role = "moderator"
				}
				return m, nil
			}
			if selected.screen == ScreenTutorial {
				m.tutorialModel = newTutorialModel()
				m.tutorialModel.viewMode = tutorialViewList
				m.tutorialModel.allTutorials = m.tutorialManager.GetAllTutorials()
				return m, nil
			}
			if selected.screen == ScreenQuests {
				m.questsModel = newQuestsModel()
				m.questsModel.viewMode = questViewActive
				m.questsModel.activeQuests = m.questManager.GetActiveQuests(m.playerID)
				m.questsModel.availableQuests = m.questManager.GetAvailableQuests(m.playerID)
				m.questsModel.completedQuests = m.questManager.GetCompletedQuests(m.playerID)
				return m, nil
			}

			return m, nil
		}
	}

	return m, nil
}

// viewMainMenu renders the main menu screen.
//
// Layout:
//   - Header: Player name, credits, and current location
//   - Welcome message: Personalized greeting
//   - Menu items: Scrollable list with cursor highlight
//   - Footer: Key binding help text
//
// Visual Styling:
//   - Selected item: Highlighted with ">" prefix and special styling
//   - Unselected items: Normal styling with spacing indent
func (m Model) viewMainMenu() string {
	// Get current system name for header display
	systemName := "Unknown"
	if m.player != nil && m.player.CurrentSystem.String() != "00000000-0000-0000-0000-000000000000" {
		// Try to load system name
		// For now, just show "Space"
		systemName = "Space"
	}

	// Render header with player stats (name, credits, location)
	s := renderHeader(m.username, m.player.Credits, systemName)
	s += "\n"

	// Welcome message with player name
	welcome := fmt.Sprintf("Welcome, Commander %s!", m.username)
	s += subtitleStyle.Render(welcome) + "\n\n"

	// Main menu items list
	s += "Main Menu:\n\n"
	for i, item := range m.mainMenu.items {
		if i == m.mainMenu.cursor {
			// Highlight selected item with cursor indicator
			s += "> " + selectedMenuItemStyle.Render(item.label) + "\n"
		} else {
			// Normal item with spacing to align with selected items
			s += "  " + menuItemStyle.Render(item.label) + "\n"
		}
	}

	// Help text footer with key bindings
	s += renderFooter("↑/↓ or j/k: Navigate  •  Enter: Select  •  q: Quit")

	return s
}
