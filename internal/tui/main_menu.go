// File: internal/tui/main_menu.go
// Project: Terminal Velocity
// Description: Terminal UI component for main_menu
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

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
				playerSettings, _ := m.settingsManager.LoadSettings(m.playerID)
				m.settingsModel.settings = playerSettings
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
