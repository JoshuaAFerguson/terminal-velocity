// File: internal/tui/messages.go
// Project: Terminal Velocity
// Description: Custom message type definitions for async BubbleTea operations
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14
//
// This file defines all custom message types used in the TUI for communicating
// async operation results. BubbleTea uses message passing to handle async operations:
//
//   1. Screen calls an async operation (e.g., m.loadMarketData())
//   2. Operation returns a tea.Cmd function
//   3. BubbleTea executes the function in a goroutine
//   4. Function sends a message (e.g., marketLoadedMsg) back to Update()
//   5. Update() handles the message and updates model state
//
// Message Naming Convention:
//   - Operation completed: <noun>LoadedMsg (e.g., playerLoadedMsg)
//   - Action completed: <noun>ActionMsg (e.g., missionActionMsg)
//   - Event occurred: <noun>Msg (e.g., errorMsg)
//
// Message Categories:
//   - Market/Trading: Trading and commodity operations
//   - Shipyard: Ship purchasing operations
//   - Missions/Quests: Mission and quest lifecycle
//   - Navigation: System jumping and navigation
//   - Combat: Combat actions and results
//   - Equipment: Equipment installation and loadouts
//   - Space/Planets: Space view and planet services
//   - Player/Account: Player data and credits
//   - Social: Chat, mail, friends, notifications
//   - Meta: News, leaderboards, achievements
//   - Session: Save/load operations
//   - Errors/Loading: Error handling and loading states
//
// Thread Safety:
//   - Messages are immutable and safe to pass between goroutines
//   - Message handlers in Update() are called sequentially
//   - Avoid sharing mutable state between messages
//
// Error Handling:
//   - Most messages include an err field for error reporting
//   - Check err != nil in Update() before processing success case
//   - Display errors to user via m.err or screen-specific error messages

package tui

import (
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// ===== Market/Trading Messages =====
// Messages related to commodity trading and market operations

// tradingEnhancedDataMsg is sent when market data has been loaded for the enhanced trading screen.
//
// This message contains all the data needed to display the trading interface:
//   - Available commodities with current prices
//   - Player's current cargo for comparison
//   - Any errors that occurred during loading
//
// Note: The legacy trading screen uses marketLoadedMsg defined in trading.go.
// Use tradingEnhancedDataMsg for the enhanced trading interface.
type tradingEnhancedDataMsg struct {
	commodities []*models.Commodity // All commodities available at current planet
	cargo       []*models.CargoItem // Player's current cargo items
	err         error                // Error if loading failed
}

// transactionCompleteMsg is sent when a buy/sell transaction has completed.
//
// This message contains the result of a commodity transaction:
//   - Action type (buy or sell)
//   - Which commodity was traded
//   - Quantity traded
//   - Player's new credit balance
//   - Any errors that occurred
//
// The receiving screen should update the player's credits and cargo display.
type transactionCompleteMsg struct {
	action      string // "buy" or "sell"
	commodityID string // ID of the commodity traded
	quantity    int    // Amount bought or sold
	newBalance  int64  // Player's credits after transaction
	err         error  // Error if transaction failed
}

// ===== Shipyard Messages =====
// Messages related to ship purchasing and management

// shipyardEnhancedDataMsg is sent when shipyard data has been loaded.
//
// Note: The legacy shipyard screen uses shipyardLoadedMsg defined in shipyard.go.
// Use shipyardEnhancedDataMsg for the enhanced shipyard interface.
type shipyardEnhancedDataMsg struct {
	ships       []*models.ShipType // Available ship types for purchase
	currentShip *models.Ship       // Player's current ship for comparison
	err         error              // Error if loading failed
}

// shipPurchaseCompleteMsg is sent when a ship purchase has completed.
//
// The new ship replaces the player's current ship, and cargo is transferred.
type shipPurchaseCompleteMsg struct {
	ship *models.Ship // The newly purchased ship
	err  error        // Error if purchase failed
}

// ===== Mission Messages =====
// Messages related to the mission system

// missionsLoadedMsg is sent when missions have been loaded from the database.
//
// Missions are divided into:
//   - available: Missions the player can accept (limited by max active missions)
//   - active: Missions currently in progress
type missionsLoadedMsg struct {
	available []*models.Mission // Missions available to accept
	active    []*models.Mission // Missions currently active
	err       error             // Error if loading failed
}

// missionActionMsg is sent when a mission action has completed.
//
// Actions include:
//   - "accept": Player accepted a mission (moves to active)
//   - "decline": Player declined a mission (removes from available)
//   - "abandon": Player abandoned an active mission (fails mission)
//   - "complete": Player completed a mission (awards rewards)
type missionActionMsg struct {
	action    string    // Type of action performed
	missionID uuid.UUID // ID of the affected mission
	err       error     // Error if action failed
}

// ===== Quest Messages =====
// Messages related to the quest and storyline system

// questsLoadedMsg is sent when quests have been loaded from the quest manager.
//
// Quests are divided into:
//   - active: Quests currently in progress
//   - available: Quests the player can start
type questsLoadedMsg struct {
	active    []*models.Quest // Quests currently in progress
	available []*models.Quest // Quests available to start
	err       error           // Error if loading failed
}

// questActionMsg is sent when a quest action has completed.
//
// Actions include:
//   - "accept": Player started a quest
//   - "abandon": Player abandoned a quest
//   - "complete": Player completed a quest (awards rewards)
type questActionMsg struct {
	action  string // Type of action performed
	questID string // Quest ID as string (matches quest manager API)
	err     error  // Error if action failed
}

// questProgressMsg is sent when quest progress has been updated.
//
// This message tracks individual objective completion within a quest.
type questProgressMsg struct {
	questID     uuid.UUID // ID of the quest
	objectiveID string    // ID of the objective
	completed   bool      // Whether objective is now complete
	err         error     // Error if update failed
}

// ===== Navigation Messages =====
// Messages related to system navigation and jumping

// navigationEnhancedDataMsg is sent when navigation data has been loaded.
//
// Note: The legacy navigation screen uses systemsLoadedMsg defined in navigation.go.
// Use navigationEnhancedDataMsg for the enhanced navigation interface.
type navigationEnhancedDataMsg struct {
	current *models.StarSystem   // Player's current system
	nearby  []*models.StarSystem // Connected systems (jump destinations)
	err     error                // Error if loading failed
}

// jumpExecutedMsg is sent when a system jump has completed.
//
// The player's location is updated to the destination system.
// Fuel is consumed based on jump distance.
type jumpExecutedMsg struct {
	destination *models.StarSystem // System jumped to
	fuelUsed    int                // Fuel consumed by jump
	err         error              // Error if jump failed (insufficient fuel, etc.)
}

// ===== Combat Messages =====
// Messages related to turn-based combat

// combatInitMsg is sent when combat has been initiated.
//
// This message contains the initial state for a combat encounter:
//   - Player's ship stats and equipment
//   - Enemy ship stats and equipment
//   - Encounter details (context for the combat)
type combatInitMsg struct {
	playerShip *models.Ship      // Player's ship
	enemyShip  *models.Ship      // Enemy ship
	encounter  *models.Encounter // Encounter context
	err        error             // Error if combat init failed
}

// combatActionMsg is sent when a player combat action has been executed.
//
// Actions include:
//   - "fire": Attack with a weapon
//   - "evade": Attempt to dodge enemy attacks
//   - "defend": Defensive posture (shield boost)
//   - "hail": Attempt communication (may end combat)
type combatActionMsg struct {
	actionType string // Type of action performed
	weaponSlot int    // Weapon slot used (for "fire" action)
	hit        bool   // Whether attack hit
	damage     int    // Damage dealt (if hit)
	logMessage string // Message to display in combat log
	combatOver bool   // Whether combat has ended
	victory    bool   // Whether player won (if combatOver)
	err        error  // Error if action failed
}

// enemyTurnMsg is sent when the enemy's combat turn has been executed.
//
// The enemy AI selects an action based on difficulty level and ship state.
type enemyTurnMsg struct {
	action     string // Enemy action ("fire", "evade", etc.)
	hit        bool   // Whether enemy attack hit
	damage     int    // Damage dealt to player
	logMessage string // Message to display in combat log
	combatOver bool   // Whether combat has ended (player defeated)
	err        error  // Error if enemy turn failed
}

// combatEndMsg is sent when combat has concluded.
//
// This message contains the results of the combat:
//   - Victory status
//   - Rewards (credits, loot, experience)
type combatEndMsg struct {
	victory       bool               // Whether player won
	creditsEarned int64              // Credits earned from victory
	loot          []*models.Equipment // Equipment dropped by enemy
	experience    int                // Experience points earned
	err           error              // Error if combat end failed
}

// combatLootGeneratedMsg is sent when loot has been generated from combat.
//
// Loot generation is based on:
//   - Enemy ship type
//   - Player's luck stat
//   - Rarity rolls (common, uncommon, rare, legendary)
type combatLootGeneratedMsg struct {
	loot          interface{}      // *combat.LootDrop (avoiding circular import)
	enemyShipType *models.ShipType // Type of ship defeated
	err           error            // Error if loot generation failed
}

// combatLootCollectedMsg is sent when loot has been collected after combat.
//
// Loot is added to the player's cargo if there's space,
// or the player can choose to jettison items.
type combatLootCollectedMsg struct {
	success       bool   // Whether loot was successfully collected
	creditsEarned int64  // Credits earned from selling auto-sell items
	message       string // Message to display to player
	err           error  // Error if collection failed
}

// ===== Equipment Messages =====
// Messages related to equipment and ship outfitting

// equipmentLoadedMsg is sent when equipment data has been loaded.
type equipmentLoadedMsg struct {
	available []*models.Equipment    // Equipment available for purchase
	installed []*models.Equipment    // Equipment currently installed on ship
	loadouts  []*models.ShipLoadout  // Saved loadout configurations
	err       error                  // Error if loading failed
}

// equipmentActionMsg is sent when an equipment action has completed.
//
// Actions: "install", "uninstall", "buy", "sell"
type equipmentActionMsg struct {
	action      string    // Type of action performed
	equipmentID uuid.UUID // ID of the equipment
	slotIndex   int       // Equipment slot index (for install/uninstall)
	err         error     // Error if action failed
}

// loadoutActionMsg is sent when a loadout action has completed.
//
// Actions: "save", "load", "delete", "clone"
type loadoutActionMsg struct {
	action      string    // Type of action performed
	loadoutID   uuid.UUID // ID of the loadout
	loadoutName string    // Name of the loadout
	err         error     // Error if action failed
}

// ===== Space View Messages =====
// Messages related to the 3D space visualization

// spaceViewLoadedMsg is sent when space view data has been loaded.
type spaceViewLoadedMsg struct {
	system      *models.StarSystem // Current star system
	planets     []*models.Planet   // Planets in system
	nearbyShips []*models.Ship     // Other ships in vicinity
	playerShip  *models.Ship       // Player's ship
	err         error              // Error if loading failed
}

// targetSelectedMsg is sent when a target has been selected in space view.
type targetSelectedMsg struct {
	target      interface{} // Target object (Ship, Planet, etc.)
	targetType  string      // Type: "ship", "planet", "station"
	targetIndex int         // Index of selected target
	err         error       // Error if selection failed
}

// ===== Planet Messages =====
// Messages related to planet landing and services

// planetLoadedMsg is sent when planet data has been loaded.
type planetLoadedMsg struct {
	planet   *models.Planet // Planet data
	services []string       // Available services
	err      error          // Error if loading failed
}

// serviceCompleteMsg is sent when a planet service has been completed.
//
// Services: "refuel", "repair", etc.
type serviceCompleteMsg struct {
	service string // Service performed
	cost    int64  // Cost of service
	err     error  // Error if service failed
}

// ===== Player Messages =====
// Messages related to player data and account

// playerDataLoadedMsg is sent when player data has been loaded.
type playerDataLoadedMsg struct {
	player *models.Player // Player data
	ship   *models.Ship   // Player's ship
	err    error          // Error if loading failed
}

// creditsUpdatedMsg is sent when player's credits have been updated.
type creditsUpdatedMsg struct {
	newBalance int64  // New credit balance
	change     int64  // Amount changed (positive or negative)
	reason     string // Reason for change (for logging)
	err        error  // Error if update failed
}

// ===== Chat Messages =====
// Messages related to the chat system

// chatMessageReceivedMsg is sent when a chat message has been received.
type chatMessageReceivedMsg struct {
	channel string // Channel: "global", "system", "faction", "dm"
	from    string // Sender username
	message string // Message content
	err     error  // Error if receive failed
}

// chatMessageSentMsg is sent when a chat message has been sent.
type chatMessageSentMsg struct {
	success bool  // Whether send succeeded
	err     error // Error if send failed
}

// ===== News Messages =====
// Messages related to the news system

// newsLoadedMsg is sent when news articles have been loaded.
type newsLoadedMsg struct {
	articles []*models.NewsArticle // News articles
	err      error                 // Error if loading failed
}

// ===== Leaderboard Messages =====
// Messages related to player rankings

// leaderboardsLoadedMsg is sent when leaderboards have been loaded.
type leaderboardsLoadedMsg struct {
	credits     []*models.LeaderboardEntry // Credits leaderboard
	combat      []*models.LeaderboardEntry // Combat leaderboard
	trade       []*models.LeaderboardEntry // Trade leaderboard
	exploration []*models.LeaderboardEntry // Exploration leaderboard
	err         error                      // Error if loading failed
}

// ===== Achievement Messages =====
// Messages related to achievement tracking

// achievementsLoadedMsg is sent when achievements have been loaded.
type achievementsLoadedMsg struct {
	unlocked []*models.Achievement // Unlocked achievements
	all      []*models.Achievement // All achievements
	err      error                 // Error if loading failed
}

// achievementUnlockedMsg is sent when an achievement has been unlocked.
type achievementUnlockedMsg struct {
	achievement *models.Achievement // Newly unlocked achievement
	err         error               // Error if unlock failed
}

// ===== Session Messages =====
// Messages related to session save/load

// sessionSavedMsg is sent when session has been saved.
type sessionSavedMsg struct {
	timestamp int64 // Timestamp of save
	err       error // Error if save failed
}

// sessionLoadedMsg is sent when session has been loaded.
type sessionLoadedMsg struct {
	success bool  // Whether load succeeded
	err     error // Error if load failed
}

// ===== Error and Loading Messages =====
// Generic messages for error handling and loading states

// errorMsg is sent when a generic error occurs.
type errorMsg struct {
	context string // Context of the error
	err     error  // The error itself
}

// loadingStartMsg is sent when a loading operation starts.
type loadingStartMsg struct {
	operation string // Description of operation
}

// loadingCompleteMsg is sent when a loading operation completes.
type loadingCompleteMsg struct {
	operation string // Description of operation
	success   bool   // Whether operation succeeded
	err       error  // Error if operation failed
}

// operationCompleteMsg is sent when a generic operation completes.
//
// This is a catch-all message for operations that don't have
// a specific message type.
type operationCompleteMsg struct {
	operation string // Description of operation
	success   bool   // Whether operation succeeded
	message   string // Message to display to user
	err       error  // Error if operation failed
}
