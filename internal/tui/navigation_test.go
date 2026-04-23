// File: internal/tui/navigation_test.go
// Project: Terminal Velocity
// Description: Tests for screen navigation and transitions
// Version: 1.0.1
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// TestScreenTransitions verifies basic screen navigation paths
func TestScreenTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialScreen Screen
		keyPress     string
		expectedScreen Screen
	}{
		// Space View navigation
		{"SpaceView to Landing", ScreenSpaceView, "l", ScreenLanding},
		{"SpaceView to Combat", ScreenSpaceView, "f", ScreenCombatEnhanced},
		{"SpaceView to Navigation", ScreenSpaceView, "m", ScreenNavigation},
		{"SpaceView to NavigationJ", ScreenSpaceView, "j", ScreenNavigation},
		{"SpaceView to MainMenu", ScreenSpaceView, "esc", ScreenMainMenu},

		// Landing navigation
		{"Landing to Trading", ScreenLanding, "c", ScreenTradingEnhanced},
		{"Landing to Outfitter", ScreenLanding, "o", ScreenOutfitterEnhanced},
		{"Landing to Shipyard", ScreenLanding, "s", ScreenShipyardEnhanced},
		{"Landing to Missions", ScreenLanding, "m", ScreenMissionBoardEnhanced},
		{"Landing to Quests", ScreenLanding, "q", ScreenQuestBoardEnhanced},
		{"Landing to News", ScreenLanding, "b", ScreenNews},
		{"Landing to SpaceView", ScreenLanding, "t", ScreenSpaceView},
		{"Landing to SpaceView ESC", ScreenLanding, "esc", ScreenSpaceView},

		// Return paths
		{"Trading to Landing", ScreenTradingEnhanced, "esc", ScreenLanding},
		{"Shipyard to Landing", ScreenShipyardEnhanced, "esc", ScreenLanding},
		{"Missions to Landing", ScreenMissionBoardEnhanced, "esc", ScreenLanding},
		{"Quests to Landing", ScreenQuestBoardEnhanced, "esc", ScreenLanding},
		{"Navigation to SpaceView", ScreenNavigation, "esc", ScreenSpaceView},
		{"Combat to MainMenu", ScreenCombatEnhanced, "esc", ScreenMainMenu},
		{"Combat Retreat", ScreenCombatEnhanced, "r", ScreenSpaceView},

		// Login flow
		{"Login to Registration", ScreenLogin, "enter", ScreenRegistration}, // When cursor on register button
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal model
			m := Model{
				screen: tt.initialScreen,
				width:  80,
				height: 24,
			}

			// Initialize sub-models to prevent nil panics
			m.spaceView = newSpaceViewModel()
			m.landing = newLandingModel()
			m.navigation = newNavigationModel()
			m.mainMenu = newMainMenuModel()
			m.tradingEnhanced = newTradingEnhancedModel()
			m.shipyardEnhanced = newShipyardEnhancedModel()
			m.missionBoardEnhanced = newMissionBoardEnhancedModel()
			m.combatEnhanced = newCombatEnhancedModel()
			m.questBoardEnhanced = newQuestBoardEnhancedModel()
			m.loginModel = newLoginModel()

			// Special setup for SpaceView combat test - requires target
			if tt.initialScreen == ScreenSpaceView && tt.keyPress == "f" {
				m.spaceView.hasTarget = true
				m.spaceView.targetIndex = 0
			}

			// Screens that honor previousScreen on Esc need it wired for
			// the test to exercise the SpaceView-return path.
			if tt.initialScreen == ScreenNavigation && tt.expectedScreen == ScreenSpaceView {
				m.previousScreen = ScreenSpaceView
				m.hasPreviousScreen = true
			}

			// SpaceView's "l" (land) only transitions when a planet is
			// reachable — the handler intentionally ignores keypresses
			// in empty systems. Seed the viewport with one planet so the
			// transition fires.
			if tt.initialScreen == ScreenSpaceView && tt.keyPress == "l" {
				m.spaceView.planets = []*models.Planet{{Name: "Testworld"}}
			}

			// Simulate key press
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyPress)}
			if tt.keyPress == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			} else if tt.keyPress == "enter" {
				msg = tea.KeyMsg{Type: tea.KeyEnter}
				// For login test, set focused field to register button (index 3)
				if tt.initialScreen == ScreenLogin {
					m.loginModel.focusedField = 3
				}
			}

			// Update model
			updated, _ := m.Update(msg)
			updatedModel := updated.(Model)

			// Verify screen transition
			if updatedModel.screen != tt.expectedScreen {
				t.Errorf("Expected screen %v, got %v", tt.expectedScreen, updatedModel.screen)
			}
		})
	}
}

// TestListNavigation verifies up/down navigation in list-based screens
func TestListNavigation(t *testing.T) {
	tests := []struct {
		name          string
		screen        Screen
		initialIndex  int
		keyPress      string
		expectedIndex int
	}{
		// Trading navigation
		{"Trading down", ScreenTradingEnhanced, 0, "down", 1},
		{"Trading up from 1", ScreenTradingEnhanced, 1, "up", 0},
		{"Trading up at top", ScreenTradingEnhanced, 0, "up", 0}, // Should stay at 0

		// Missions navigation
		{"Missions down", ScreenMissionBoardEnhanced, 0, "down", 1},
		{"Missions up", ScreenMissionBoardEnhanced, 1, "up", 0},

		// Quests navigation
		{"Quests down", ScreenQuestBoardEnhanced, 0, "down", 1},
		{"Quests up", ScreenQuestBoardEnhanced, 1, "up", 0},

		// NOTE: Navigation (ScreenNavigation) list nav is covered by
		// TestNavigationModel in the real screen's test scope — this
		// screen loads systems from the DB rather than from a static
		// slice, so it's not exercisable by this table-driven test.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				screen: tt.screen,
				width:  80,
				height: 24,
			}

			// Initialize models
			m.tradingEnhanced = newTradingEnhancedModel()
			m.missionBoardEnhanced = newMissionBoardEnhancedModel()
			m.questBoardEnhanced = newQuestBoardEnhancedModel()

			// Set initial index
			switch tt.screen {
			case ScreenTradingEnhanced:
				m.tradingEnhanced.selectedCommodity = tt.initialIndex
			case ScreenMissionBoardEnhanced:
				m.missionBoardEnhanced.selectedMission = tt.initialIndex
			case ScreenQuestBoardEnhanced:
				m.questBoardEnhanced.selectedQuest = tt.initialIndex
			}

			// Simulate key press
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyPress)}
			if tt.keyPress == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.keyPress == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			}

			// Update model
			updated, _ := m.Update(msg)
			updatedModel := updated.(Model)

			// Verify index changed correctly
			var actualIndex int
			switch tt.screen {
			case ScreenTradingEnhanced:
				actualIndex = updatedModel.tradingEnhanced.selectedCommodity
			case ScreenMissionBoardEnhanced:
				actualIndex = updatedModel.missionBoardEnhanced.selectedMission
			case ScreenQuestBoardEnhanced:
				actualIndex = updatedModel.questBoardEnhanced.selectedQuest
			}

			if actualIndex != tt.expectedIndex {
				t.Errorf("Expected index %d, got %d", tt.expectedIndex, actualIndex)
			}
		})
	}
}

// TestVimKeyBindings verifies vim-style navigation works
func TestVimKeyBindings(t *testing.T) {
	m := Model{
		screen: ScreenTradingEnhanced,
		width:  80,
		height: 24,
	}
	m.tradingEnhanced = newTradingEnhancedModel()
	m.tradingEnhanced.selectedCommodity = 1

	// Test 'k' (up)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	updated, _ := m.Update(msg)
	updatedModel := updated.(Model)

	if updatedModel.tradingEnhanced.selectedCommodity != 0 {
		t.Errorf("Vim 'k' should move up, expected 0, got %d",
			updatedModel.tradingEnhanced.selectedCommodity)
	}

	// Test 'j' (down)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = updatedModel.Update(msg)
	updatedModel = updated.(Model)

	if updatedModel.tradingEnhanced.selectedCommodity != 1 {
		t.Errorf("Vim 'j' should move down, expected 1, got %d",
			updatedModel.tradingEnhanced.selectedCommodity)
	}
}

// TestChatToggle verifies chat expand/collapse in Space View
func TestChatToggle(t *testing.T) {
	m := Model{
		screen: ScreenSpaceView,
		width:  80,
		height: 24,
	}
	m.spaceView = newSpaceViewModel()

	// Initial state should be collapsed
	if m.spaceView.chatExpanded {
		t.Error("Chat should start collapsed")
	}

	// Press 'c' to expand
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	updated, _ := m.Update(msg)
	updatedModel := updated.(Model)

	if !updatedModel.spaceView.chatExpanded {
		t.Error("Chat should be expanded after pressing 'c'")
	}

	// Press 'c' again to collapse
	updated, _ = updatedModel.Update(msg)
	updatedModel = updated.(Model)

	if updatedModel.spaceView.chatExpanded {
		t.Error("Chat should be collapsed after pressing 'c' again")
	}
}

// TestDataInitialization verifies all screens initialize with sample data
func TestDataInitialization(t *testing.T) {
	tests := []struct {
		name     string
		checkFunc func(Model) bool
		errorMsg string
	}{
		{
			"Trading has commodities",
			func(m Model) bool {
				return len(m.tradingEnhanced.commodities) > 0
			},
			"Trading should initialize with commodities",
		},
		{
			"Shipyard has ships",
			func(m Model) bool {
				return len(m.shipyardEnhanced.ships) > 0
			},
			"Shipyard should initialize with ships",
		},
		{
			"Missions has missions",
			func(m Model) bool {
				return len(m.missionBoardEnhanced.missions) > 0
			},
			"Mission board should initialize with missions",
		},
		{
			"Quests has quests",
			func(m Model) bool {
				return len(m.questBoardEnhanced.activeQuests) > 0 ||
					len(m.questBoardEnhanced.availableQuests) > 0
			},
			"Quest board should initialize with quests",
		},
		// Navigation (ScreenNavigation) loads its systems asynchronously
		// from the DB, so there is no static fixture to assert on at
		// construction time — the async path is covered elsewhere.
		{
			"Combat has ships",
			func(m Model) bool {
				return len(m.combatEnhanced.playerShip.weapons) > 0 &&
					len(m.combatEnhanced.enemyShip.weapons) > 0
			},
			"Combat should initialize with ship weapons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				width:  80,
				height: 24,
			}

			// Initialize all models
			m.tradingEnhanced = newTradingEnhancedModel()
			m.shipyardEnhanced = newShipyardEnhancedModel()
			m.missionBoardEnhanced = newMissionBoardEnhancedModel()
			m.questBoardEnhanced = newQuestBoardEnhancedModel()
			m.combatEnhanced = newCombatEnhancedModel()

			if !tt.checkFunc(m) {
				t.Error(tt.errorMsg)
			}
		})
	}
}
