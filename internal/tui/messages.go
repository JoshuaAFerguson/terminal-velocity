// File: internal/tui/messages.go
// Project: Terminal Velocity
// Description: Message type definitions for async BubbleTea operations
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Market/Trading messages
// Note: marketLoadedMsg already defined in trading.go (old screen)
// Use tradingEnhancedDataMsg for enhanced screens
type tradingEnhancedDataMsg struct {
	commodities []*models.Commodity
	cargo       []*models.CargoItem
	err         error
}

type transactionCompleteMsg struct {
	action       string // "buy" or "sell"
	commodityID  string
	quantity     int
	newBalance   int64
	err          error
}

// Shipyard messages
// Note: shipyardLoadedMsg/shipPurchasedMsg already defined in shipyard.go (old screen)
// Use shipyardEnhancedDataMsg for enhanced screens
type shipyardEnhancedDataMsg struct{
	ships       []*models.ShipType
	currentShip *models.Ship
	err         error
}

type shipPurchaseCompleteMsg struct {
	ship *models.Ship
	err  error
}

// Mission messages
type missionsLoadedMsg struct {
	available []*models.Mission
	active    []*models.Mission
	err       error
}

type missionActionMsg struct {
	action    string // "accept", "decline", "abandon", "complete"
	missionID uuid.UUID
	err       error
}

// Quest messages
type questsLoadedMsg struct {
	active    []*models.Quest
	available []*models.Quest
	err       error
}

type questActionMsg struct {
	action  string // "accept", "abandon", "complete"
	questID string // Quest ID as string (matches quest manager API)
	err     error
}

type questProgressMsg struct {
	questID      uuid.UUID
	objectiveID  string
	completed    bool
	err          error
}

// Navigation/System messages
// Note: systemsLoadedMsg/jumpCompleteMsg already defined in navigation.go (old screen)
// Use navigationEnhancedDataMsg for enhanced screens
type navigationEnhancedDataMsg struct {
	current  *models.StarSystem
	nearby   []*models.StarSystem
	err      error
}

type jumpExecutedMsg struct {
	destination *models.StarSystem
	fuelUsed    int
	err         error
}

// Combat messages
type combatInitMsg struct {
	playerShip *models.Ship
	enemyShip  *models.Ship
	encounter  *models.Encounter
	err        error
}

type combatActionMsg struct {
	actionType  string // "fire", "evade", "defend", "hail"
	weaponSlot  int
	hit         bool
	damage      int
	logMessage  string
	combatOver  bool
	victory     bool
	err         error
}

type enemyTurnMsg struct {
	action     string
	hit        bool
	damage     int
	logMessage string
	combatOver bool
	err        error
}

type combatEndMsg struct {
	victory       bool
	creditsEarned int64
	loot          []*models.Equipment
	experience    int
	err           error
}

// Combat loot messages
type combatLootGeneratedMsg struct {
	loot          interface{} // *combat.LootDrop (avoiding circular import)
	enemyShipType *models.ShipType
	err           error
}

type combatLootCollectedMsg struct {
	success       bool
	creditsEarned int64
	message       string
	err           error
}

// Equipment/Outfitter messages
type equipmentLoadedMsg struct {
	available []*models.Equipment
	installed []*models.Equipment
	// loadouts  []*models.Loadout  // TODO: models.Loadout not yet implemented
	err       error
}

type equipmentActionMsg struct {
	action      string // "install", "uninstall", "buy", "sell"
	equipmentID uuid.UUID
	slotIndex   int
	err         error
}

type loadoutActionMsg struct {
	action     string // "save", "load", "delete", "clone"
	loadoutID  uuid.UUID
	loadoutName string
	err        error
}

// Space View messages
type spaceViewLoadedMsg struct {
	system      *models.StarSystem
	planets     []*models.Planet
	nearbyShips []*models.Ship
	playerShip  *models.Ship
	err         error
}

type targetSelectedMsg struct {
	target      interface{} // could be Ship, Planet, etc.
	targetType  string      // "ship", "planet", "station"
	targetIndex int         // index of the selected target
	err         error
}

// Landing/Planet messages
type planetLoadedMsg struct {
	planet   *models.Planet
	services []string
	err      error
}

type serviceCompleteMsg struct {
	service string // "refuel", "repair"
	cost    int64
	err     error
}

// Player/Account messages
type playerDataLoadedMsg struct {
	player *models.Player
	ship   *models.Ship
	err    error
}

type creditsUpdatedMsg struct {
	newBalance int64
	change     int64
	reason     string
	err        error
}

// Chat messages
type chatMessageReceivedMsg struct {
	channel string
	from    string
	message string
	err     error
}

type chatMessageSentMsg struct {
	success bool
	err     error
}

// News messages
type newsLoadedMsg struct {
	articles []*models.NewsArticle
	err      error
}

// Leaderboard messages
type leaderboardsLoadedMsg struct {
	credits     []*models.LeaderboardEntry
	combat      []*models.LeaderboardEntry
	trade       []*models.LeaderboardEntry
	exploration []*models.LeaderboardEntry
	err         error
}

// Achievement messages
type achievementsLoadedMsg struct {
	unlocked []*models.Achievement
	all      []*models.Achievement
	err      error
}

type achievementUnlockedMsg struct {
	achievement *models.Achievement
	err         error
}

// Session messages
type sessionSavedMsg struct {
	timestamp int64
	err       error
}

type sessionLoadedMsg struct {
	success bool
	err     error
}

// Error messages
type errorMsg struct {
	context string
	err     error
}

// Loading state messages
type loadingStartMsg struct {
	operation string
}

type loadingCompleteMsg struct {
	operation string
	success   bool
	err       error
}

// Generic success message
type operationCompleteMsg struct {
	operation string
	success   bool
	message   string
	err       error
}
