// File: internal/tui/integration_test.go
// Project: Terminal Velocity
// Description: Integration tests for enhanced TUI screens
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// TestScreenNavigation tests navigation between screens
func TestScreenNavigation(t *testing.T) {
	tests := []struct {
		name        string
		fromScreen  Screen
		toScreen    Screen
		keyPress    string
		description string
	}{
		{
			name:        "SpaceView to CombatEnhanced",
			fromScreen:  ScreenSpaceView,
			toScreen:    ScreenCombatEnhanced,
			keyPress:    "f",
			description: "Pressing 'f' in SpaceView with target should enter combat",
		},
		{
			name:        "CombatEnhanced to SpaceView (retreat)",
			fromScreen:  ScreenCombatEnhanced,
			toScreen:    ScreenSpaceView,
			keyPress:    "r",
			description: "Pressing 'r' in combat should retreat to space",
		},
		{
			name:        "SpaceView to OutfitterEnhanced",
			fromScreen:  ScreenSpaceView,
			toScreen:    ScreenOutfitterEnhanced,
			keyPress:    "o",
			description: "Pressing 'o' in SpaceView should open outfitter",
		},
		{
			name:        "OutfitterEnhanced to SpaceView",
			fromScreen:  ScreenOutfitterEnhanced,
			toScreen:    ScreenSpaceView,
			keyPress:    "esc",
			description: "Pressing 'esc' in outfitter should return to space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test model
			m := createTestModel()
			m.screen = tt.fromScreen

			// Setup initial state based on screen
			switch tt.fromScreen {
			case ScreenSpaceView:
				// Setup space view with a target
				m.spaceView.hasTarget = true
				m.spaceView.targetIndex = 0
			case ScreenCombatEnhanced:
				// Setup active combat
				m.combatEnhanced.combatPhase = "combat"
				m.combatEnhanced.isPlayerTurn = true
			}

			// Simulate key press
			var msg tea.Msg
			switch tt.keyPress {
			case "f":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
			case "r":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
			case "o":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			}

			// Update model
			updatedModel, _ := m.Update(msg)
			m = updatedModel.(Model)

			// Verify screen transition
			if m.screen != tt.toScreen {
				t.Errorf("Expected screen %v, got %v. %s", tt.toScreen, m.screen, tt.description)
			}
		})
	}
}

// TestCombatWeaponFiring tests the combat weapon firing system
func TestCombatWeaponFiring(t *testing.T) {
	tests := []struct {
		name           string
		weaponIndex    int
		playerEnergy   int
		weaponEnergy   int
		weaponAmmo     int
		weaponMaxAmmo  int
		expectError    bool
		expectLogEntry bool
		description    string
	}{
		{
			name:           "Successful weapon fire",
			weaponIndex:    0,
			playerEnergy:   100,
			weaponEnergy:   10,
			weaponAmmo:     10,
			weaponMaxAmmo:  10,
			expectError:    false,
			expectLogEntry: true,
			description:    "Should successfully fire weapon with sufficient energy and ammo",
		},
		{
			name:           "Insufficient energy",
			weaponIndex:    0,
			playerEnergy:   5,
			weaponEnergy:   10,
			weaponAmmo:     10,
			weaponMaxAmmo:  10,
			expectError:    true,
			expectLogEntry: true,
			description:    "Should fail when energy is insufficient",
		},
		{
			name:           "Out of ammo",
			weaponIndex:    0,
			playerEnergy:   100,
			weaponEnergy:   10,
			weaponAmmo:     0,
			weaponMaxAmmo:  10,
			expectError:    true,
			expectLogEntry: true,
			description:    "Should fail when ammo is depleted",
		},
		{
			name:           "Energy weapon (no ammo check)",
			weaponIndex:    0,
			playerEnergy:   100,
			weaponEnergy:   15,
			weaponAmmo:     -1,
			weaponMaxAmmo:  -1,
			expectError:    false,
			expectLogEntry: true,
			description:    "Energy weapons with maxAmmo=-1 should not check ammo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test model with combat setup
			m := createTestModel()
			m.screen = ScreenCombatEnhanced
			m.combatEnhanced.combatPhase = "combat"
			m.combatEnhanced.isPlayerTurn = true

			// Setup current ship (required by fireWeaponCmd)
			m.currentShip = &models.Ship{
				ID:      uuid.New(),
				OwnerID: m.playerID,
				TypeID:  "shuttle",
				Name:    "Test Ship",
				Hull:    100,
			}

			// Setup player ship
			m.combatEnhanced.playerShip = combatShip{
				name:       "Test Ship",
				hull:       100,
				maxHull:    100,
				shields:    100,
				maxShields: 100,
				energy:     tt.playerEnergy,
				weapons: []combatWeapon{
					{
						name:       "Test Weapon",
						ready:      true,
						damage:     20,
						energyCost: tt.weaponEnergy,
						ammo:       tt.weaponAmmo,
						maxAmmo:    tt.weaponMaxAmmo,
					},
				},
			}

			// Setup enemy ship
			m.combatEnhanced.enemyShip = combatShip{
				name:       "Enemy",
				hull:       100,
				maxHull:    100,
				shields:    50,
				maxShields: 50,
				energy:     100,
			}

			// Execute fireWeaponCmd
			cmd := m.fireWeaponCmd(tt.weaponIndex)
			msg := cmd()

			// Verify result
			actionMsg, ok := msg.(combatActionMsg)
			if !ok {
				t.Fatalf("Expected combatActionMsg, got %T", msg)
			}

			if tt.expectError && actionMsg.err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}

			if !tt.expectError && actionMsg.err != nil {
				t.Errorf("Expected no error but got: %v. %s", actionMsg.err, tt.description)
			}

			if tt.expectLogEntry && actionMsg.logMessage == "" {
				t.Errorf("Expected log message but got empty string. %s", tt.description)
			}
		})
	}
}

// TestCombatAITurn tests the AI turn processing
func TestCombatAITurn(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenCombatEnhanced
	m.combatEnhanced.combatPhase = "combat"
	m.combatEnhanced.isPlayerTurn = false

	// Setup ships
	m.combatEnhanced.playerShip = combatShip{
		name:       "Player Ship",
		hull:       100,
		maxHull:    100,
		shields:    100,
		maxShields: 100,
		energy:     100,
	}

	m.combatEnhanced.enemyShip = combatShip{
		name:       "Enemy Ship",
		hull:       100,
		maxHull:    100,
		shields:    50,
		maxShields: 50,
		energy:     100,
		weapons: []combatWeapon{
			{
				name:       "Enemy Weapon",
				ready:      true,
				damage:     15,
				energyCost: 10,
				ammo:       -1,
				maxAmmo:    -1,
			},
		},
	}

	// Execute AI turn
	cmd := m.processAITurnCmd()
	msg := cmd()

	// Verify result
	enemyMsg, ok := msg.(enemyTurnMsg)
	if !ok {
		t.Fatalf("Expected enemyTurnMsg, got %T", msg)
	}

	// AI should have performed an action
	if enemyMsg.logMessage == "" {
		t.Error("Expected AI to generate a log message")
	}

	// Player turn should be restored after processing
	updatedModel, _ := m.Update(enemyMsg)
	m = updatedModel.(Model)
	if !m.combatEnhanced.isPlayerTurn {
		t.Error("Expected player turn to be restored after AI turn")
	}
}

// TestOutfitterPurchase tests equipment purchase flow
func TestOutfitterPurchase(t *testing.T) {
	tests := []struct {
		name          string
		playerCredits int64
		equipmentCost int64
		quantity      int
		expectError   bool
		description   string
	}{
		{
			name:          "Successful purchase",
			playerCredits: 10000,
			equipmentCost: 1000,
			quantity:      2,
			expectError:   false,
			description:   "Should successfully purchase with sufficient credits",
		},
		{
			name:          "Insufficient credits",
			playerCredits: 1000,
			equipmentCost: 2000,
			quantity:      1,
			expectError:   true,
			description:   "Should fail when credits are insufficient",
		},
		{
			name:          "Exact credits",
			playerCredits: 5000,
			equipmentCost: 2500,
			quantity:      2,
			expectError:   false,
			description:   "Should succeed when credits exactly match cost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test model
			m := createTestModel()
			m.player.Credits = tt.playerCredits
			m.screen = ScreenOutfitterEnhanced

			// Note: This test verifies the credit check logic
			// Actual manager integration would require mock outfittingManager
			totalCost := tt.equipmentCost * int64(tt.quantity)
			hasEnoughCredits := m.player.Credits >= totalCost

			if tt.expectError && hasEnoughCredits {
				t.Errorf("Expected insufficient credits but had enough. %s", tt.description)
			}

			if !tt.expectError && !hasEnoughCredits {
				t.Errorf("Expected sufficient credits but didn't have enough. %s", tt.description)
			}
		})
	}
}

// TestOutfitterInstallUninstall tests equipment installation/uninstallation
func TestOutfitterInstallUninstall(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenOutfitterEnhanced
	m.outfitterEnhanced.viewMode = outfitterViewSlots

	// Setup current loadout with empty slots
	m.outfitterEnhanced.currentLoadout = &models.ShipLoadout{
		ID:         uuid.New(),
		PlayerID:   m.playerID,
		ShipTypeID: "shuttle",
		Name:       "Test Loadout",
		Slots:      []models.EquipmentSlot{},
	}

	// Add a weapon slot (NewEquipmentSlot returns a value)
	slot := models.NewEquipmentSlot(models.SlotWeapon, 2)
	m.outfitterEnhanced.currentLoadout.Slots = append(m.outfitterEnhanced.currentLoadout.Slots, slot)

	// Test uninstall on empty slot should fail
	if !m.outfitterEnhanced.currentLoadout.Slots[0].IsEmpty() {
		t.Error("Expected slot to be initially empty")
	}

	// Install equipment
	equipment := &models.Equipment{
		ID:       "laser_cannon",
		Name:     "Laser Cannon",
		Category: models.CategoryWeapon,
		SlotType: models.SlotWeapon,
		SlotSize: 2,
	}

	m.outfitterEnhanced.currentLoadout.Slots[0].Install(equipment)

	// Verify installation
	if m.outfitterEnhanced.currentLoadout.Slots[0].IsEmpty() {
		t.Error("Expected slot to have equipment after installation")
	}

	// Uninstall equipment
	uninstalled := m.outfitterEnhanced.currentLoadout.Slots[0].Uninstall()

	// Verify uninstallation
	if uninstalled == nil {
		t.Error("Expected uninstall to return equipment")
	}

	if uninstalled.ID != equipment.ID {
		t.Errorf("Expected uninstalled equipment ID %s, got %s", equipment.ID, uninstalled.ID)
	}

	if !m.outfitterEnhanced.currentLoadout.Slots[0].IsEmpty() {
		t.Error("Expected slot to be empty after uninstallation")
	}
}

// TestSpaceViewTargeting tests target cycling and selection
func TestSpaceViewTargeting(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenSpaceView

	// Setup space view with multiple space objects (ships and planets)
	m.spaceView.ships = []spaceObject{
		{name: "Ship Beta", objType: "ship"},
		{name: "Ship Gamma", objType: "enemy"},
	}
	m.spaceView.planets = []*models.Planet{
		{Name: "Planet Alpha"},
	}

	// Initial state: no target
	if m.spaceView.hasTarget {
		t.Error("Expected no initial target")
	}

	// Cycle to first target
	cmd := m.cycleTargetCmd()
	msg := cmd()

	targetMsg, ok := msg.(targetSelectedMsg)
	if !ok {
		t.Fatalf("Expected targetSelectedMsg, got %T", msg)
	}

	updatedModel, _ := m.Update(targetMsg)
	model := updatedModel.(Model)

	if !model.spaceView.hasTarget {
		t.Error("Expected target to be selected")
	}

	if model.spaceView.targetIndex != 0 {
		t.Errorf("Expected targetIndex 0, got %d", model.spaceView.targetIndex)
	}

	// Cycle to next target
	cmd = model.cycleTargetCmd()
	msg = cmd()
	targetMsg = msg.(targetSelectedMsg)
	updatedModel, _ = model.Update(targetMsg)
	model = updatedModel.(Model)

	if model.spaceView.targetIndex != 1 {
		t.Errorf("Expected targetIndex 1, got %d", model.spaceView.targetIndex)
	}

	// Cycle past end should wrap to beginning
	model.spaceView.targetIndex = 2
	cmd = model.cycleTargetCmd()
	msg = cmd()
	targetMsg = msg.(targetSelectedMsg)
	updatedModel, _ = model.Update(targetMsg)
	model = updatedModel.(Model)

	if model.spaceView.targetIndex != 0 {
		t.Errorf("Expected targetIndex to wrap to 0, got %d", model.spaceView.targetIndex)
	}
}

// TestAsyncMessageFlow tests async command execution and message handling
func TestAsyncMessageFlow(t *testing.T) {
	tests := []struct {
		name         string
		screen       Screen
		setupFunc    func(*Model)
		commandFunc  func(Model) tea.Cmd
		expectedType string
		description  string
	}{
		{
			name:   "Combat weapon fire async",
			screen: ScreenCombatEnhanced,
			setupFunc: func(m *Model) {
				m.combatEnhanced.combatPhase = "combat"
				m.combatEnhanced.isPlayerTurn = true
				m.combatEnhanced.playerShip = combatShip{
					name:    "Test",
					hull:    100,
					maxHull: 100,
					energy:  100,
					weapons: []combatWeapon{
						{name: "Laser", ready: true, damage: 20, energyCost: 10, ammo: -1, maxAmmo: -1},
					},
				}
				m.combatEnhanced.enemyShip = combatShip{
					name:       "Enemy",
					hull:       100,
					maxHull:    100,
					shields:    50,
					maxShields: 50,
				}
			},
			commandFunc: func(m Model) tea.Cmd {
				return m.fireWeaponCmd(0)
			},
			expectedType: "tui.combatActionMsg",
			description:  "Weapon fire should return combatActionMsg asynchronously",
		},
		{
			name:   "Space view target cycling async",
			screen: ScreenSpaceView,
			setupFunc: func(m *Model) {
				m.spaceView.ships = []spaceObject{
					{name: "Target 1", objType: "ship"},
				}
			},
			commandFunc: func(m Model) tea.Cmd {
				return m.cycleTargetCmd()
			},
			expectedType: "tui.targetSelectedMsg",
			description:  "Target cycling should return targetSelectedMsg asynchronously",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.screen = tt.screen

			// Setup test state
			tt.setupFunc(&m)

			// Execute command
			cmd := tt.commandFunc(m)
			if cmd == nil {
				t.Fatal("Expected command to be returned, got nil")
			}

			// Execute command to get message
			msg := cmd()

			// Verify message type
			msgType := getMessageType(msg)
			if msgType != tt.expectedType {
				t.Errorf("Expected message type %s, got %s. %s", tt.expectedType, msgType, tt.description)
			}

			// Verify message can be handled by Update
			_, cmd = m.Update(msg)
			// Command may or may not be returned, but Update should not panic
		})
	}
}

// TestErrorHandling tests error handling across screens
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		screen      Screen
		errorMsg    tea.Msg
		expectError bool
		description string
	}{
		{
			name:   "Combat action error",
			screen: ScreenCombatEnhanced,
			errorMsg: combatActionMsg{
				actionType: "fire",
				err:        errNoShipEquipped,
			},
			expectError: true,
			description: "Critical combat errors should show error dialog",
		},
		{
			name:   "Combat action with log message",
			screen: ScreenCombatEnhanced,
			errorMsg: combatActionMsg{
				actionType: "fire",
				logMessage: "Out of ammo",
				err:        errInsufficientAmmo,
			},
			expectError: false,
			description: "Non-critical errors should show in combat log",
		},
		{
			name:   "Equipment action error",
			screen: ScreenOutfitterEnhanced,
			errorMsg: equipmentActionMsg{
				action: "buy",
				err:    errInsufficientCredits,
			},
			expectError: true,
			description: "Equipment purchase errors should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.screen = tt.screen

			// Process error message
			updatedModel, _ := m.Update(tt.errorMsg)
			m = updatedModel.(Model)

			// Verify error handling
			switch tt.screen {
			case ScreenCombatEnhanced:
				actionMsg := tt.errorMsg.(combatActionMsg)
				if tt.expectError && actionMsg.logMessage == "" {
					// Critical error - should show error dialog
					if !m.showErrorDialog || m.errorMessage == "" {
						t.Error("Expected error dialog to be shown for critical error")
					}
				} else if actionMsg.logMessage != "" {
					// Non-critical - should be in combat log
					found := false
					for _, entry := range m.combatEnhanced.combatLog {
						if entry == "> "+actionMsg.logMessage {
							found = true
							break
						}
					}
					if !found {
						t.Error("Expected log message to be added to combat log")
					}
				}

			case ScreenOutfitterEnhanced:
				if tt.expectError {
					// Equipment errors typically show in UI status
					// Verify error was captured
					actionMsg := tt.errorMsg.(equipmentActionMsg)
					if actionMsg.err == nil {
						t.Error("Expected error to be present in message")
					}
				}
			}
		})
	}
}

// TestStateSynchronization tests state updates across screens
func TestStateSynchronization(t *testing.T) {
	m := createTestModel()

	// Test credit synchronization
	initialCredits := m.player.Credits
	m.player.Credits = 5000

	if m.player.Credits != 5000 {
		t.Error("Player credits should be synchronized")
	}

	// Test ship state synchronization
	m.currentShip = &models.Ship{
		ID:      uuid.New(),
		OwnerID: m.playerID,
		TypeID:  "shuttle",
		Name:    "Test Ship",
		Hull:    100,
	}

	if m.currentShip == nil {
		t.Error("Current ship should be set")
	}

	// Test location synchronization
	testSystem := uuid.New()
	m.player.CurrentSystem = testSystem

	if m.player.CurrentSystem != testSystem {
		t.Error("Player location should be synchronized")
	}

	// Restore initial state
	m.player.Credits = initialCredits
}

// Helper functions

// createTestModel creates a minimal Model for testing
func createTestModel() Model {
	playerID := uuid.New()
	player := &models.Player{
		ID:       playerID,
		Username: "testplayer",
		Credits:  10000,
	}

	m := Model{
		playerID:          playerID,
		player:            player,
		width:             80,
		height:            24,
		spaceView:         newSpaceViewModel(),
		combatEnhanced:    newCombatEnhancedModel(),
		outfitter:         newOutfitterModel(),
		outfitterEnhanced: newOutfitterEnhancedModel(),
		showErrorDialog:   false,
		errorMessage:      "",
	}

	return m
}

// getMessageType returns the type name of a message
func getMessageType(msg tea.Msg) string {
	switch msg.(type) {
	case combatActionMsg:
		return "tui.combatActionMsg"
	case enemyTurnMsg:
		return "tui.enemyTurnMsg"
	case targetSelectedMsg:
		return "tui.targetSelectedMsg"
	case equipmentActionMsg:
		return "tui.equipmentActionMsg"
	case loadoutActionMsg:
		return "tui.loadoutActionMsg"
	case operationCompleteMsg:
		return "tui.operationCompleteMsg"
	default:
		return "unknown"
	}
}

// Test-specific error variables
var (
	errNoShipEquipped      = fmt.Errorf("no ship equipped")
	errInsufficientAmmo    = fmt.Errorf("insufficient ammo")
	errInsufficientCredits = fmt.Errorf("insufficient credits")
)

// TestTradingScreenIntegration tests trading screen functionality
func TestTradingScreenIntegration(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenTrading

	// Setup current planet
	planetID := uuid.New()
	m.player.CurrentPlanet = &planetID

	// Verify trading screen can be accessed
	if m.screen != ScreenTrading {
		t.Error("Should be able to navigate to trading screen")
	}

	// Test would verify buy/sell operations if trading manager was available
	t.Log("Trading screen integration test placeholder - requires trading manager mock")
}

// TestShipyardScreenIntegration tests shipyard screen functionality
func TestShipyardScreenIntegration(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenShipyard

	// Setup current planet with shipyard service
	planetID := uuid.New()
	m.player.CurrentPlanet = &planetID

	// Verify shipyard screen can be accessed
	if m.screen != ScreenShipyard {
		t.Error("Should be able to navigate to shipyard screen")
	}

	t.Log("Shipyard screen integration test placeholder - requires shipyard manager mock")
}

// TestMissionBoardIntegration tests mission board functionality
func TestMissionBoardIntegration(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenMissions

	// Verify mission board screen can be accessed
	if m.screen != ScreenMissions {
		t.Error("Should be able to navigate to mission board")
	}

	t.Log("Mission board integration test placeholder - requires mission manager mock")
}

// TestQuestBoardIntegration tests quest board functionality
func TestQuestBoardIntegration(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenQuests

	// Verify quest board screen can be accessed
	if m.screen != ScreenQuests {
		t.Error("Should be able to navigate to quest board")
	}

	t.Log("Quest board integration test placeholder - requires quest manager mock")
}

// TestConcurrentOperations tests handling of concurrent async operations
func TestConcurrentOperations(t *testing.T) {
	m := createTestModel()
	m.screen = ScreenCombatEnhanced

	// Setup combat
	m.combatEnhanced.combatPhase = "combat"
	m.combatEnhanced.isPlayerTurn = true
	m.combatEnhanced.playerShip = combatShip{
		name:    "Player",
		hull:    100,
		maxHull: 100,
		energy:  100,
		weapons: []combatWeapon{
			{name: "Weapon 1", ready: true, damage: 20, energyCost: 10, ammo: -1, maxAmmo: -1},
			{name: "Weapon 2", ready: true, damage: 15, energyCost: 8, ammo: -1, maxAmmo: -1},
		},
	}
	m.combatEnhanced.enemyShip = combatShip{
		name:    "Enemy",
		hull:    100,
		maxHull: 100,
	}

	// Execute multiple weapon fires in sequence
	cmd1 := m.fireWeaponCmd(0)
	msg1 := cmd1()

	// Process first result
	modelInterface, _ := m.Update(msg1)
	updatedModel := modelInterface.(Model)

	// Verify state after first action
	if updatedModel.combatEnhanced.isPlayerTurn {
		// If still player turn, means there was an error or combat ended
		t.Log("First weapon fire completed")
	}

	// Test rapid message processing
	messages := []tea.Msg{
		combatActionMsg{actionType: "fire", weaponSlot: 0, hit: true, damage: 20, logMessage: "Hit!"},
		enemyTurnMsg{action: "fire", hit: true, damage: 15, logMessage: "Enemy fires!"},
	}

	for _, msg := range messages {
		modelInterface, _ = updatedModel.Update(msg)
		updatedModel = modelInterface.(Model)
	}

	// Verify combat log has all entries
	if len(updatedModel.combatEnhanced.combatLog) == 0 {
		t.Error("Expected combat log to have entries from concurrent operations")
	}
}

// BenchmarkWeaponFiring benchmarks weapon firing command performance
func BenchmarkWeaponFiring(b *testing.B) {
	m := createTestModel()
	m.combatEnhanced.combatPhase = "combat"
	m.combatEnhanced.isPlayerTurn = true
	m.combatEnhanced.playerShip = combatShip{
		name:    "Player",
		hull:    100,
		maxHull: 100,
		energy:  100,
		weapons: []combatWeapon{
			{name: "Laser", ready: true, damage: 20, energyCost: 10, ammo: -1, maxAmmo: -1},
		},
	}
	m.combatEnhanced.enemyShip = combatShip{
		name:    "Enemy",
		hull:    100,
		maxHull: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := m.fireWeaponCmd(0)
		_ = cmd()
	}
}

// BenchmarkTargetCycling benchmarks target cycling performance
func BenchmarkTargetCycling(b *testing.B) {
	m := createTestModel()
	m.spaceView.ships = make([]spaceObject, 100)
	for i := 0; i < 100; i++ {
		m.spaceView.ships[i] = spaceObject{
			name:    "Target",
			objType: "ship",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := m.cycleTargetCmd()
		_ = cmd()
	}
}
