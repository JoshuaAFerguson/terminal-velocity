// File: internal/tui/model.go
// Project: Terminal Velocity
// Description: Core TUI model with BubbleTea integration, screen routing, and state management
// Version: 1.3.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// This file implements the main TUI model for Terminal Velocity using the BubbleTea framework.
// It follows the Model-View-Update (MVU) architecture pattern where:
//   - Model: Holds all application state (player data, screen models, managers)
//   - Update: Handles messages and returns updated model + commands
//   - View: Renders the current state to the terminal
//
// Key architectural patterns:
//   - Screen-based routing: Each screen has its own model and update/view functions
//   - Async operations: Long-running operations return tea.Cmd for non-blocking execution
//   - Message passing: Custom message types communicate async operation results
//   - Repository pattern: All database access goes through typed repositories
//   - Manager pattern: Game systems (chat, factions, etc.) are managed by dedicated managers
//
// Thread Safety:
//   - The BubbleTea Update() function is called sequentially, so no locking is needed in TUI code
//   - However, managers and repositories may be accessed concurrently and use their own locking
//   - Use context.Background() for database operations in tea.Cmd functions
//
// Screen Transitions:
//   - Screens change via m.screen = ScreenName in Update()
//   - Always return tea.ClearScreen when changing screens to prevent artifacts
//   - Screen-specific state is preserved in sub-models (e.g., m.trading, m.combat)

package tui

import (
	"context"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/achievements"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/admin"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/chat"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/encounters"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/factions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/factionwar"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/friends"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/leaderboards"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/mail"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/marketplace"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/missions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/news"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/notifications"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/npcterritory"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/outfitting"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/presence"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/pvp"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/quests"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/settings"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/territory"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/trade"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/tutorial"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// Screen represents different game screens in the TUI.
//
// The screen enum is used for routing in the Update() and View() functions.
// Each screen has a corresponding sub-model (e.g., mainMenuModel, combatModel)
// and update/view functions (e.g., updateCombat, viewCombat).
//
// Screen Categories:
//   - Authentication: Login, Registration
//   - Core Game: MainMenu, Game, Help, Tutorial, Settings
//   - Navigation: Navigation, NavigationEnhanced, SpaceView, Landing
//   - Commerce: Trading, TradingEnhanced, Cargo, TradeRoutes, Marketplace
//   - Ships: Shipyard, ShipyardEnhanced, Outfitter, OutfitterEnhanced, ShipManagement, Fleet
//   - Combat: Combat, CombatEnhanced, PvP, Encounter
//   - Missions & Quests: Missions, MissionBoardEnhanced, Quests, QuestBoardEnhanced
//   - Social: Chat, Players, Friends, Mail, Notifications
//   - Organizations: Factions, Trade
//   - Progression: Achievements, Leaderboards, News
//   - Administration: Admin
type Screen int

const (
	// ScreenMainMenu is the main menu screen shown after login
	ScreenMainMenu Screen = iota

	// ScreenGame is the primary game view (currently minimal, mostly redirects to other screens)
	ScreenGame

	// ScreenNavigation is the legacy system navigation and jump interface
	ScreenNavigation

	// ScreenTrading is the legacy commodity trading interface
	ScreenTrading

	// ScreenCargo displays the player's cargo hold and allows jettisoning items
	ScreenCargo

	// ScreenShipyard is the legacy ship purchase interface
	ScreenShipyard

	// ScreenOutfitter is the legacy equipment purchase and installation interface
	ScreenOutfitter

	// ScreenShipManagement provides ship repair, refuel, and maintenance services
	ScreenShipManagement

	// ScreenCombat is the turn-based combat interface with tactical options
	ScreenCombat

	// ScreenMissions shows available and active missions
	ScreenMissions

	// ScreenAchievements displays unlocked and locked achievements with progress
	ScreenAchievements

	// ScreenEncounter handles random encounters (pirates, traders, distress calls)
	ScreenEncounter

	// ScreenNews shows recent news articles generated from game events
	ScreenNews

	// ScreenLeaderboards displays player rankings across multiple categories
	ScreenLeaderboards

	// ScreenPlayers shows online players and their locations
	ScreenPlayers

	// ScreenChat provides multi-channel chat (global, system, faction, DM)
	ScreenChat

	// ScreenFactions manages player factions, membership, and treasury
	ScreenFactions

	// ScreenTrade handles player-to-player trading with escrow
	ScreenTrade

	// ScreenPvP manages PvP challenges and faction wars
	ScreenPvP

	// ScreenHelp displays context-sensitive help content
	ScreenHelp

	// ScreenOutfitterEnhanced is an enhanced equipment browser with filtering
	ScreenOutfitterEnhanced

	// ScreenSettings manages player preferences and color schemes
	ScreenSettings

	// ScreenAdmin provides server administration tools (RBAC-protected)
	ScreenAdmin

	// ScreenTutorial displays interactive tutorials for new players
	ScreenTutorial

	// ScreenQuests shows quest progression and branching narratives
	ScreenQuests

	// ScreenRegistration handles new player account creation
	ScreenRegistration

	// ScreenLogin is the initial login screen for unauthenticated users
	ScreenLogin

	// ScreenSpaceView is the 3D space visualization with targeting
	ScreenSpaceView

	// ScreenLanding handles planet landing and service selection
	ScreenLanding

	// ScreenTradingEnhanced is the enhanced commodity trading interface with analytics
	ScreenTradingEnhanced

	// ScreenShipyardEnhanced is the enhanced ship browser with detailed comparisons
	ScreenShipyardEnhanced

	// ScreenMissionBoardEnhanced is the enhanced mission browser with filtering
	ScreenMissionBoardEnhanced

	// ScreenNavigationEnhanced is the enhanced navigation interface with route planning
	ScreenNavigationEnhanced

	// ScreenCombatEnhanced is the enhanced combat interface with advanced tactics
	ScreenCombatEnhanced

	// ScreenQuestBoardEnhanced is the enhanced quest browser with storyline tracking
	ScreenQuestBoardEnhanced

	// ScreenTradeRoutes displays profitable trade routes and market analysis
	ScreenTradeRoutes

	// ScreenMail manages player-to-player mail and messages
	ScreenMail

	// ScreenFleet manages multi-ship ownership, escorts, and formations
	ScreenFleet

	// ScreenFriends manages friend lists and social connections
	ScreenFriends

	// ScreenMarketplace is the player-to-player item marketplace
	ScreenMarketplace

	// ScreenNotifications displays game notifications and alerts
	ScreenNotifications
	ScreenPilotRecord
	ScreenFactionWars
	ScreenTerritoryMap
)

// Model is the main TUI model that holds all application state.
//
// The Model follows the BubbleTea MVU (Model-View-Update) pattern and contains:
//   - Current screen and routing information
//   - Player data (loaded from database)
//   - Database repositories for data access
//   - Sub-models for each screen (preserves screen-specific state)
//   - Game system managers (chat, factions, quests, etc.)
//   - Error and loading state
//
// State Lifecycle:
//   1. Model is initialized via NewModel() or NewLoginModel()
//   2. Init() loads player data asynchronously
//   3. Update() handles messages and state changes
//   4. View() renders the current screen
//   5. Sub-models are updated/viewed based on current screen
//
// Screen State Preservation:
//   - Each screen has a dedicated sub-model (e.g., m.trading for ScreenTrading)
//   - Screen state is preserved when switching between screens
//   - This allows users to return to screens with their previous state intact
//
// Database Access:
//   - All database operations go through repositories (playerRepo, systemRepo, etc.)
//   - Async operations use tea.Cmd to avoid blocking the UI
//   - Results are communicated via custom message types (e.g., playerLoadedMsg)
//
// Manager Integration:
//   - Managers handle game systems (achievements, chat, factions, etc.)
//   - Managers are thread-safe and can be called from any screen
//   - Managers often run background goroutines for periodic tasks
type Model struct {
	// ===== Screen Routing =====

	// screen is the currently active screen (determines which view is rendered)
	screen Screen
	// previousScreen/hasPreviousScreen record who opened the current screen,
	// so that screens reachable from multiple places (e.g. dock-at-station
	// and the main menu both lead to ScreenOutfitterEnhanced) know where to
	// send the user on Esc. We can't just use a Screen field with a zero
	// sentinel because Screen(0) == ScreenMainMenu (iota).
	previousScreen    Screen
	hasPreviousScreen bool

	// ===== Player State =====

	// player contains the current player's data (loaded from database)
	// nil during login/registration or if loading failed
	player *models.Player

	// playerID is the UUID of the current player
	// Set during authentication, used to load player data
	playerID uuid.UUID

	// username is the player's display name
	// Set during authentication for display purposes
	username string

	// currentShip is the player's active ship (loaded from database)
	// nil if player doesn't have a ship yet
	currentShip *models.Ship

	// currentSystem caches the last-loaded star system (from player.CurrentSystem)
	currentSystem *models.StarSystem

	// currentPlanet caches the last-docked planet (nil when in space)
	currentPlanet *models.Planet

	// ===== Database Repositories =====
	// Repositories provide typed CRUD operations for database access
	// All database operations should go through repositories, never direct SQL

	playerRepo      *database.PlayerRepository      // Player accounts and stats
	systemRepo      *database.SystemRepository      // Star systems and connections
	sshKeyRepo      *database.SSHKeyRepository      // SSH public keys for authentication
	shipRepo        *database.ShipRepository        // Player ships and equipment
	marketRepo      *database.MarketRepository      // Market prices and commodities
	mailRepo        *database.MailRepository        // Player mail system
	socialRepo      *database.SocialRepository      // Friends, blocks, etc.
	itemRepo        *database.ItemRepository        // Items and equipment
	achievementRepo *database.AchievementRepository // Achievement progress

	// ===== Terminal Dimensions =====

	// width is the terminal width in characters (updated on WindowSizeMsg)
	width int

	// height is the terminal height in characters (updated on WindowSizeMsg)
	height int

	// ===== Screen Sub-Models =====
	// Each screen has a dedicated model to preserve state between screen switches
	// Sub-models are initialized in NewModel() and persist for the session

	mainMenu             mainMenuModel             // Main menu after login
	gameView             gameViewModel             // Primary game view (minimal, mostly redirects)
	registration         registrationModel         // New account registration
	navigation           navigationModel           // Legacy system navigation
	trading              tradingModel              // Legacy commodity trading
	cargo                cargoModel                // Cargo hold management
	shipyard             shipyardModel             // Legacy ship purchasing
	outfitter            outfitterModel            // Legacy equipment management
	shipManagement       shipManagementModel       // Ship services (repair, refuel)
	combat               combatModel               // Turn-based combat
	missions             missionsModel             // Mission board
	pilotRecord          pilotRecordModel          // Pilot record and stats
	achievementsUI       achievementsModel         // Achievement tracking
	encounterModel       encounterModel            // Random encounters
	newsModel            newsModel                 // News articles
	newsTicker           newsTickerState           // Main-menu newsreel ticker state
	factionWarsModel     factionWarsModel          // Faction wars screen state
	territoryMap         territoryMapModel         // Territory map screen state
	leaderboardsModel    leaderboardsModel         // Player rankings
	playersModel         playersModel              // Online players list
	chatModel            chatModel                 // Multi-channel chat
	factionsModel        factionsModel             // Faction management
	tradeModel           tradeModel                // Player trading
	pvpModel             pvpModel                  // PvP challenges
	helpModel            helpModel                 // Context-sensitive help
	outfitterEnhanced    outfitterEnhancedModel    // Enhanced equipment browser
	settingsModel        settingsModel             // Player preferences
	adminModel           adminModel                // Server administration
	tutorialModel        tutorialModel             // Interactive tutorials
	questsModel          questsModel               // Quest progression
	loginModel           loginModel                // Login screen
	spaceView            spaceViewModel            // 3D space visualization
	landing              landingModel              // Planet landing
	tradingEnhanced      tradingEnhancedModel      // Enhanced trading interface
	shipyardEnhanced     shipyardEnhancedModel     // Enhanced ship browser
	missionBoardEnhanced missionBoardEnhancedModel // Enhanced mission board
	combatEnhanced       combatEnhancedModel       // Enhanced combat
	questBoardEnhanced   questBoardEnhancedModel   // Enhanced quest board
	tradeRoutes          tradeRoutesState          // Trade route analysis
	mail                 mailState                 // Mail system
	fleet                fleetState                // Fleet management
	friends              friendsState              // Friends list
	marketplace          marketplaceState          // Player marketplace
	notifications        notificationsState        // Notifications

	// ===== Game System Managers =====
	// Managers encapsulate game systems and often run background workers
	// Managers are thread-safe and can be accessed from any screen

	achievementManager   *achievements.Manager   // Achievement tracking and unlocks
	newsManager          *news.Manager           // News generation from events
	leaderboardManager   *leaderboards.Manager   // Player ranking calculations
	presenceManager      *presence.Manager       // Online player tracking
	chatManager          *chat.Manager           // Multi-channel chat system
	mailManager          *mail.Manager           // Player mail system
	fleetManager         *fleet.Manager          // Multi-ship management
	friendsManager       *friends.Manager        // Social connections
	notificationsManager *notifications.Manager  // Game notifications
	marketplaceManager   *marketplace.Manager    // Player marketplace
	factionManager       *factions.Manager       // Player factions
	territoryManager     *territory.Manager      // Territory control
	tradeManager         *trade.Manager          // Player trading
	pvpManager           *pvp.Manager            // PvP combat
	encounterManager     *encounters.Manager     // Random encounters
	outfittingManager    *outfitting.Manager     // Equipment management
	settingsManager      *settings.Manager       // Player settings
	adminManager         *admin.Manager          // Server administration
	tutorialManager      *tutorial.Manager       // Tutorial system
	questManager         *quests.Manager         // Quest system
	missionManager       *missions.Manager       // Mission system

	// ===== Achievement Display Queue =====

	// pendingAchievements holds newly unlocked achievements waiting to be displayed
	// Achievements are added via checkAchievements() and displayed via getAchievementNotification()
	// The queue allows multiple achievements to be shown sequentially without blocking gameplay
	pendingAchievements []*models.Achievement

	// ===== Error Handling and Loading State =====

	// err holds any error that occurred during async operations
	// Checked in View() to display error screens when non-nil
	err error

	// errorMessage is a user-friendly error message for display
	errorMessage string

	// showErrorDialog controls whether to show an error dialog overlay
	showErrorDialog bool

	// loadingOperation describes the current loading operation (e.g., "Loading player data...")
	loadingOperation string

	// isLoading indicates whether an async operation is in progress
	isLoading bool

	// ===== P5 Additions =====

	// factionWarManager is the faction war system (P5C). Set from the server so every
	// session sees the same active-war list; nil-tolerant for standalone/test model construction.
	factionWarManager *factionwar.Manager

	// npcTerritoryManager is the NPC territory system (P5D). Shared across sessions so a
	// system captured mid-war is reflected on every player's space-view banner immediately.
	npcTerritoryManager *npcterritory.Manager
}

// NewModel creates a new TUI model for an authenticated player.
//
// This constructor is used when the player has already been authenticated
// (e.g., via SSH public key or password). It initializes:
//   - All screen sub-models
//   - All game system managers
//   - Database repositories
//   - Player state (playerID and username)
//
// The model starts on ScreenMainMenu and will load player data in Init().
//
// Parameters:
//   - playerID: UUID of the authenticated player
//   - username: Display name of the player
//   - Various repositories and managers for game systems
//
// Returns:
//   - Initialized Model ready for use with BubbleTea
//
// Usage:
//   model := NewModel(playerID, username, playerRepo, systemRepo, ...)
//   program := tea.NewProgram(model)
//   program.Run()
func NewModel(
	playerID uuid.UUID,
	username string,
	playerRepo *database.PlayerRepository,
	systemRepo *database.SystemRepository,
	sshKeyRepo *database.SSHKeyRepository,
	shipRepo *database.ShipRepository,
	marketRepo *database.MarketRepository,
	mailRepo *database.MailRepository,
	socialRepo *database.SocialRepository,
	itemRepo *database.ItemRepository,
	achievementRepo *database.AchievementRepository,
	fleetManager *fleet.Manager,
	mailManager *mail.Manager,
	notificationsManager *notifications.Manager,
	friendsManager *friends.Manager,
	marketplaceManager *marketplace.Manager,
	newsManager *news.Manager,
	chatManager *chat.Manager,
	presenceManager *presence.Manager,
	pvpManager *pvp.Manager,
	territoryManager *territory.Manager,
	tutorialManager *tutorial.Manager,
	factionWarManager *factionwar.Manager,
	npcTerritoryManager *npcterritory.Manager,
) Model {
	// All server-wide feeds fall back to standalone managers so tests
	// injecting nil still run. Production call sites always pass a
	// shared instance from the server so cross-session semantics
	// (chat fan-out, bounty board visibility, presence, etc.) work.
	if newsManager == nil {
		newsManager = news.NewManager()
	}
	if chatManager == nil {
		chatManager = chat.NewManager()
	}
	if presenceManager == nil {
		presenceManager = presence.NewManager()
	}
	if pvpManager == nil {
		pvpManager = pvp.NewManager()
	}
	if territoryManager == nil {
		territoryManager = territory.NewManager()
	}
	if tutorialManager == nil {
		tutorialManager = tutorial.NewManager()
	}
	// factionWarManager can stay nil — the TUI treats nil as "no
	// wars known" and renders a placeholder screen, so offline
	// tests don't need to spin up a manager.
	return Model{
		screen:               ScreenMainMenu,
		playerID:             playerID,
		username:             username,
		playerRepo:           playerRepo,
		systemRepo:           systemRepo,
		sshKeyRepo:           sshKeyRepo,
		shipRepo:             shipRepo,
		marketRepo:           marketRepo,
		mailRepo:             mailRepo,
		socialRepo:           socialRepo,
		itemRepo:             itemRepo,
		achievementRepo:      achievementRepo,
		width:                80,
		height:               24,
		mainMenu:             newMainMenuModel(),
		trading:              newTradingModel(),
		cargo:                newCargoModel(),
		shipyard:             newShipyardModel(),
		outfitter:            newOutfitterModel(),
		shipManagement:       newShipManagementModel(),
		combat:               newCombatModel(),
		missions:             newMissionsModel(),
		pilotRecord:          newPilotRecordModel(),
		achievementsUI:       newAchievementsModel(),
		achievementManager:   achievements.NewManager(),
		pendingAchievements:  []*models.Achievement{},
		encounterModel:       newEncounterModel(),
		newsModel:            newNewsModel(),
		newsTicker:           newNewsTickerState(),
		newsManager:          newsManager,
		leaderboardsModel:    newLeaderboardsModel(),
		leaderboardManager:   leaderboards.NewManager(),
		playersModel:         newPlayersModel(),
		presenceManager:      presenceManager,
		chatModel:            newChatModel(),
		chatManager:          chatManager,
		fleetManager:         fleetManager,
		mailManager:          mailManager,
		notificationsManager: notificationsManager,
		friendsManager:       friendsManager,
		marketplaceManager:   marketplaceManager,
		factionsModel:        newFactionsModel(),
		factionManager:       factions.NewManager(),
		territoryManager:     territoryManager,
		tradeModel:           newTradeModel(),
		tradeManager:         trade.NewManager(),
		pvpModel:             newPvPModel(),
		pvpManager:           pvpManager,
		helpModel:            newHelpModel(),
		encounterManager:     encounters.NewManager(),
		outfitterEnhanced:    newOutfitterEnhancedModel(),
		outfittingManager:    outfitting.NewManager(),
		settingsModel:        newSettingsModel(),
		settingsManager:      settings.NewManager(".config/terminal-velocity"),
		adminModel:           newAdminModel(),
		adminManager:         admin.NewManager(playerRepo),
		tutorialModel:        newTutorialModel(),
		tutorialManager:      tutorialManager,
		questsModel:          newQuestsModel(),
		questManager:         quests.NewManager(),
		missionManager:       missions.NewManager(),
		loginModel:           newLoginModel(),
		spaceView:            newSpaceViewModel(),
		landing:              newLandingModel(),
		tradingEnhanced:      newTradingEnhancedModel(),
		shipyardEnhanced:     newShipyardEnhancedModel(),
		missionBoardEnhanced: newMissionBoardEnhancedModel(),
		combatEnhanced:       newCombatEnhancedModel(),
		questBoardEnhanced:   newQuestBoardEnhancedModel(),
		fleet:                newFleetState(),
		friends:              newFriendsState(),
		notifications:        newNotificationsState(),
	}
}

// InitializeTutorials initializes tutorial progress for the player
func (m *Model) InitializeTutorials() {
	if m.tutorialManager != nil && m.playerID != uuid.Nil {
		m.tutorialManager.InitializePlayer(m.playerID)
		// Trigger first login tutorial
		m.tutorialManager.HandleTrigger(m.playerID, models.TriggerFirstLogin)
	}
}

// InitializeNews used to seed the session's news manager. Now that the
// manager is server-wide and seeded once in initDatabase, this is a
// no-op retained for callers that might still exist in tests — calling
// GenerateInitialNews on every login would flood the shared feed.
func (m *Model) InitializeNews() {}

// InitializePresence registers the player as online
func (m *Model) InitializePresence() {
	if m.presenceManager != nil && m.player != nil {
		m.presenceManager.Connect(m.player, m.currentShip)
	}
}

// UpdatePresenceActivity updates the player's current activity
func (m *Model) UpdatePresenceActivity(activity models.ActivityType) {
	if m.presenceManager != nil {
		m.presenceManager.UpdateActivity(m.playerID, activity)
	}
}

// UpdatePresenceLocation updates the player's location
func (m *Model) UpdatePresenceLocation(systemID uuid.UUID, planetID *uuid.UUID) {
	if m.presenceManager != nil {
		m.presenceManager.UpdateLocation(m.playerID, systemID, planetID)
	}
}

// NewLoginModel creates a new TUI model starting with the login screen
func NewLoginModel(
	playerRepo *database.PlayerRepository,
	systemRepo *database.SystemRepository,
	sshKeyRepo *database.SSHKeyRepository,
	shipRepo *database.ShipRepository,
	marketRepo *database.MarketRepository,
	mailRepo *database.MailRepository,
	socialRepo *database.SocialRepository,
	achievementRepo *database.AchievementRepository,
	newsManager *news.Manager,
	chatManager *chat.Manager,
	presenceManager *presence.Manager,
	pvpManager *pvp.Manager,
	territoryManager *territory.Manager,
	tutorialManager *tutorial.Manager,
	factionWarManager *factionwar.Manager,
	npcTerritoryManager *npcterritory.Manager,
) Model {
	if newsManager == nil {
		newsManager = news.NewManager()
	}
	if chatManager == nil {
		chatManager = chat.NewManager()
	}
	if presenceManager == nil {
		presenceManager = presence.NewManager()
	}
	if pvpManager == nil {
		pvpManager = pvp.NewManager()
	}
	if territoryManager == nil {
		territoryManager = territory.NewManager()
	}
	if tutorialManager == nil {
		tutorialManager = tutorial.NewManager()
	}
	// factionWarManager can stay nil — the TUI treats nil as "no
	// wars known" and renders a placeholder screen, so offline
	// tests don't need to spin up a manager.
	return Model{
		screen:               ScreenLogin,
		playerID:             uuid.Nil,
		username:             "",
		playerRepo:           playerRepo,
		systemRepo:           systemRepo,
		sshKeyRepo:           sshKeyRepo,
		shipRepo:             shipRepo,
		marketRepo:           marketRepo,
		mailRepo:             mailRepo,
		socialRepo:           socialRepo,
		achievementRepo:      achievementRepo,
		width:                80,
		height:               24,
		loginModel:           newLoginModel(),
		mainMenu:             newMainMenuModel(),
		trading:              newTradingModel(),
		cargo:                newCargoModel(),
		shipyard:             newShipyardModel(),
		outfitter:            newOutfitterModel(),
		shipManagement:       newShipManagementModel(),
		combat:               newCombatModel(),
		missions:             newMissionsModel(),
		pilotRecord:          newPilotRecordModel(),
		achievementsUI:       newAchievementsModel(),
		achievementManager:   achievements.NewManager(),
		pendingAchievements:  []*models.Achievement{},
		encounterModel:       newEncounterModel(),
		newsModel:            newNewsModel(),
		newsTicker:           newNewsTickerState(),
		newsManager:          newsManager,
		leaderboardsModel:    newLeaderboardsModel(),
		leaderboardManager:   leaderboards.NewManager(),
		playersModel:         newPlayersModel(),
		presenceManager:      presenceManager,
		chatModel:            newChatModel(),
		chatManager:          chatManager,
		mailManager:          mail.NewManager(socialRepo),
		factionsModel:        newFactionsModel(),
		factionManager:       factions.NewManager(),
		territoryManager:     territoryManager,
		factionWarManager:    factionWarManager,
		npcTerritoryManager:  npcTerritoryManager,
		factionWarsModel:     newFactionWarsModel(),
		territoryMap:         newTerritoryMapModel(),
		tradeModel:           newTradeModel(),
		tradeManager:         trade.NewManager(),
		pvpModel:             newPvPModel(),
		pvpManager:           pvpManager,
		helpModel:            newHelpModel(),
		encounterManager:     encounters.NewManager(),
		outfitterEnhanced:    newOutfitterEnhancedModel(),
		outfittingManager:    outfitting.NewManager(),
		settingsModel:        newSettingsModel(),
		settingsManager:      settings.NewManager(".config/terminal-velocity"),
		adminModel:           newAdminModel(),
		adminManager:         admin.NewManager(playerRepo),
		tutorialModel:        newTutorialModel(),
		tutorialManager:      tutorialManager,
		questsModel:          newQuestsModel(),
		questManager:         quests.NewManager(),
		missionManager:       missions.NewManager(),
		registration:         newRegistrationModel(false, nil),
		spaceView:            newSpaceViewModel(),
		landing:              newLandingModel(),
		tradingEnhanced:      newTradingEnhancedModel(),
		shipyardEnhanced:     newShipyardEnhancedModel(),
		missionBoardEnhanced: newMissionBoardEnhancedModel(),
		combatEnhanced:       newCombatEnhancedModel(),
		questBoardEnhanced:   newQuestBoardEnhancedModel(),
	}
}

// NewRegistrationModel creates a new TUI model for registration
func NewRegistrationModel(
	username string,
	requireEmail bool,
	sshKeyData []byte,
	playerRepo *database.PlayerRepository,
	systemRepo *database.SystemRepository,
	sshKeyRepo *database.SSHKeyRepository,
	shipRepo *database.ShipRepository,
	marketRepo *database.MarketRepository,
) Model {
	return Model{
		screen:       ScreenRegistration,
		playerID:     uuid.Nil,
		username:     username,
		playerRepo:   playerRepo,
		systemRepo:   systemRepo,
		sshKeyRepo:   sshKeyRepo,
		shipRepo:     shipRepo,
		marketRepo:   marketRepo,
		width:        80,
		height:       24,
		registration: newRegistrationModel(requireEmail, sshKeyData),
	}
}

// changeReputation adjusts the player's reputation with a faction by `delta`
// and persists the new value to the DB. Updates the in-memory player map
// first so subsequent render passes see the change without a round-trip.
// Silently skips when player or repo isn't ready (login/bootstrap paths).
//
// Used by the encounter system when the player attacks a patrol or rescues
// a distressed ship; returns no error because reputation drift isn't a
// critical-path failure — the hit is visible on the next Reputation read.
func (m *Model) changeReputation(factionID string, delta int) {
	if m.player == nil || factionID == "" || delta == 0 {
		return
	}
	m.player.ModifyReputation(factionID, delta)
	if m.playerRepo == nil {
		return
	}
	ctx := context.Background()
	if err := m.playerRepo.UpdateReputation(ctx, m.player.ID, factionID, delta); err != nil {
		// Swallow — the in-memory change survives the session and the
		// DB will converge on the next successful write.
		_ = err
	}
}

// currentLocationLabel returns the most specific place the player is in right
// now: cached star system name when we have it, "In transit" when the player
// has a system assigned but it hasn't loaded yet, and "Unknown" otherwise.
// Screens that render the shared header should route through this instead of
// inventing their own "Space" / screen-name fallback, otherwise the location
// field ends up lying to the user ("Location: Settings", "Location: Ship
// Management").
//
// Screens that track a docked planet (trading, shipyard) should prefer that
// planet's name when set and fall through to this helper when not docked.
func (m Model) currentLocationLabel() string {
	if m.currentSystem != nil && m.currentSystem.Name != "" {
		return m.currentSystem.Name
	}
	if m.player != nil && m.player.CurrentSystem != uuid.Nil {
		return "In transit"
	}
	return "Unknown"
}

// Init initializes the model and returns initial commands.
//
// This is the first method called by BubbleTea after creating the model.
// It performs initial setup and kicks off any async operations needed.
//
// Behavior:
//   - Always clears the screen to prevent terminal artifacts
//   - If on login/registration screen: returns only tea.ClearScreen
//   - If authenticated: loads player data via m.loadPlayer()
//
// The player data loading happens asynchronously via tea.Cmd.
// The result is communicated back via playerLoadedMsg in Update().
//
// Returns:
//   - tea.Cmd to execute (clear screen and optionally load player)
func (m Model) Init() tea.Cmd {
	// Clear screen on initialization to prevent artifacts
	// If we're on the login screen, don't load player data yet
	if m.screen == ScreenLogin || m.screen == ScreenRegistration {
		return tea.ClearScreen
	}
	return tea.Batch(tea.ClearScreen, m.loadPlayer())
}

// Update handles messages and updates the model.
//
// This is the core of the BubbleTea MVU pattern. It receives messages,
// updates the model state, and returns optional commands to execute.
//
// Message Flow:
//   1. User input (tea.KeyMsg) triggers actions
//   2. Async operations complete and send custom messages (e.g., playerLoadedMsg)
//   3. Update() processes the message and updates state
//   4. Update() may return tea.Cmd for further async operations
//   5. Cycle repeats
//
// Message Handling Order:
//   1. Global messages (Ctrl+C for quit, WindowSize for resize)
//   2. Common async messages (playerLoadedMsg, etc.)
//   3. Screen-specific messages (delegated to screen update functions)
//
// Screen Routing:
//   - Each screen has its own update function (e.g., updateCombat, updateTrading)
//   - Update() delegates to the appropriate function based on m.screen
//   - Screen updates may change m.screen to transition to other screens
//
// Thread Safety:
//   - Update() is called sequentially by BubbleTea, so no locking needed
//   - However, managers and repositories may be accessed concurrently
//
// Returns:
//   - Updated tea.Model (always return m, not Model)
//   - Optional tea.Cmd for async operations
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Quit from any screen
			return m, tea.Quit

		// Global tutorial shortcuts. Ctrl+T toggles the overlay;
		// Ctrl+N marks the current step complete and advances to the
		// next; Ctrl+K cycles hint verbosity. Ctrl-based so we don't
		// conflict with any screen's own hot keys (h, s, n, etc. are
		// all already in use somewhere).
		case "ctrl+t":
			m.tutorialModel.showOverlay = !m.tutorialModel.showOverlay
			return m, nil
		case "ctrl+n":
			if m.tutorialManager != nil && m.playerID != uuid.Nil {
				if step := m.tutorialManager.GetTutorialForScreen(m.playerID, m.getScreenName()); step != nil {
					m.tutorialManager.CompleteStep(m.playerID, step.ID)
				}
			}
			return m, nil
		case "ctrl+k":
			if m.tutorialModel.hintLevel < models.HintFull {
				m.tutorialModel.hintLevel++
			} else {
				m.tutorialModel.hintLevel = models.HintNone
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case playerLoadedMsg:
		// The login screen has its own playerLoadedMsg handler that ALSO
		// transitions to ScreenMainMenu on success — if we consume the
		// message here, login gets stuck. Skip when on login/registration
		// screens and let updateLogin / updateRegistration handle it below.
		if m.screen != ScreenLogin && m.screen != ScreenRegistration {
			m.player = msg.player
			m.currentShip = msg.ship
			m.currentSystem = msg.system
			m.err = msg.err

			// Initialize presence when player loads.
			if m.player != nil && m.err == nil {
				m.InitializePresence()
			}

			// Register this session with the shared chat manager so
			// SendGlobalMessage's fanout reaches us. Without this, we'd
			// send messages into the void (history-wise) because the
			// manager's iteration over m.histories would skip an
			// unregistered player.
			if m.player != nil && m.chatManager != nil {
				m.chatManager.GetOrCreateHistory(m.player.ID)
			}

			// Seed the in-memory achievement manager with what's already
			// persisted so checkAchievements doesn't re-fire unlock
			// notifications for anything the player has done before.
			if m.achievementManager != nil && len(msg.achievements) > 0 {
				m.achievementManager.LoadUnlocked(msg.achievements)
			}

			return m, nil
		}

	case dockedMsg:
		// dockCmd / takeoffCmd result. A non-nil planet means the player
		// just docked; nil means they took off. Either way the DB is
		// already updated — this branch just keeps the local cache in
		// sync so Landing, Trading, and Shipyard render correctly.
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
			return m, nil
		}
		m.currentPlanet = msg.planet
		if m.player != nil {
			if msg.planet != nil {
				planetID := msg.planet.ID
				m.player.CurrentPlanet = &planetID
				// Let the shared presence manager know we're planetside
				// so other players' space viewports stop rendering us as
				// a free-flying ship in the system.
				m.UpdatePresenceLocation(m.player.CurrentSystem, &planetID)
			} else {
				m.player.CurrentPlanet = nil
				m.UpdatePresenceLocation(m.player.CurrentSystem, nil)
			}
		}
		// Generate a small mission board on dock so the Missions screen
		// has content when the player visits. Populate both the top-level
		// missionManager (used by ScreenMissionBoardEnhanced) and
		// m.missions.manager (used by the main-menu Missions screen) so
		// either entry point sees the same board. In-memory for now;
		// Phase 2 of GAME_COMPLETENESS_PLAN persists + expires them.
		if msg.planet != nil {
			govID := ""
			if m.currentSystem != nil {
				govID = m.currentSystem.GovernmentID
			}
			// Populate both mission managers — the main-menu Missions
			// screen reads m.missions.manager while the landing-board
			// enhanced variant reads m.missionManager. Keeping them in
			// sync lets either entry point surface the same fresh board.
			if m.missionManager != nil {
				m.missionManager.GenerateMissions(context.Background(), msg.planet.ID, govID, 3)
			}
			if m.missions.manager != nil {
				m.missions.manager.GenerateMissions(context.Background(), msg.planet.ID, govID, 3)
			}
		}
		return m, nil
	}

	// Drop news-ticker ticks that arrive on any screen other than the
	// main menu. Resetting `active` here lets the ticker restart on
	// the next main-menu entry (see ensureNewsTickerTick). Without
	// this guard, a tick scheduled right before the user navigated
	// away would be silently re-routed to that screen's updater and
	// the ticker would never restart when they came back.
	if _, ok := msg.(newsTickerMsg); ok && m.screen != ScreenMainMenu {
		m = m.stopNewsTicker()
		return m, nil
	}

	// Delegate to screen-specific update
	switch m.screen {
	case ScreenMainMenu:
		return m.updateMainMenu(msg)
	case ScreenGame:
		return m.updateGame(msg)
	case ScreenRegistration:
		return m.updateRegistration(msg)
	case ScreenNavigation:
		return m.updateNavigation(msg)
	case ScreenTrading:
		return m.updateTrading(msg)
	case ScreenCargo:
		return m.updateCargo(msg)
	case ScreenShipyard:
		return m.updateShipyard(msg)
	case ScreenOutfitter:
		return m.updateOutfitter(msg)
	case ScreenShipManagement:
		return m.updateShipManagement(msg)
	case ScreenCombat:
		return m.updateCombat(msg)
	case ScreenMissions:
		return m.updateMissions(msg)
	case ScreenAchievements:
		return m.updateAchievements(msg)
	case ScreenEncounter:
		return m.updateEncounter(msg)
	case ScreenNews:
		return m.updateNews(msg)
	case ScreenLeaderboards:
		return m.updateLeaderboards(msg)
	case ScreenPlayers:
		return m.updatePlayers(msg)
	case ScreenChat:
		return m.updateChat(msg)
	case ScreenFactions:
		return m.updateFactions(msg)
	case ScreenTrade:
		return m.updateTrade(msg)
	case ScreenPvP:
		return m.updatePvP(msg)
	case ScreenHelp:
		return m.updateHelp(msg)
	case ScreenOutfitterEnhanced:
		return m.updateOutfitterEnhanced(msg)
	case ScreenSettings:
		return m.updateSettings(msg)
	case ScreenAdmin:
		return m.updateAdmin(msg)
	case ScreenTutorial:
		return m.updateTutorial(msg)
	case ScreenQuests:
		return m.updateQuests(msg)
	case ScreenLogin:
		return m.updateLogin(msg)
	case ScreenSpaceView:
		return m.updateSpaceView(msg)
	case ScreenLanding:
		return m.updateLanding(msg)
	case ScreenTradingEnhanced:
		return m.updateTradingEnhanced(msg)
	case ScreenShipyardEnhanced:
		return m.updateShipyardEnhanced(msg)
	case ScreenMissionBoardEnhanced:
		return m.updateMissionBoardEnhanced(msg)
	case ScreenCombatEnhanced:
		return m.updateCombatEnhanced(msg)
	case ScreenQuestBoardEnhanced:
		return m.updateQuestBoardEnhanced(msg)
	case ScreenTradeRoutes:
		return m.updateTradeRoutes(msg)
	case ScreenMail:
		return m.updateMail(msg)
	case ScreenFleet:
		return m.updateFleet(msg)
	case ScreenFriends:
		return m.updateFriends(msg)
	case ScreenMarketplace:
		return m.updateMarketplace(msg)
	case ScreenNotifications:
		return m.updateNotifications(msg)
	case ScreenPilotRecord:
		return m.updatePilotRecord(msg)
	case ScreenFactionWars:
		return m.updateFactionWars(msg)
	case ScreenTerritoryMap:
		return m.updateTerritoryMap(msg)
	default:
		return m, nil
	}
}

// View renders the model to a string for display.
//
// This is the "View" part of the BubbleTea MVU pattern. It converts
// the current model state into a string that will be displayed in the terminal.
//
// Rendering Priority:
//   1. Error screens (if m.err is set and not on login/registration)
//   2. Loading screens (if player data not loaded and not on login/registration)
//   3. Screen-specific views (based on m.screen)
//
// Screen Routing:
//   - Each screen has its own view function (e.g., viewCombat, viewTrading)
//   - View() delegates to the appropriate function based on m.screen
//   - Screen views can access all model state (player, managers, etc.)
//
// Performance Notes:
//   - View() is called frequently (on every Update() and periodically)
//   - Avoid expensive operations in view functions
//   - Pre-calculate and cache expensive computations in Update()
//
// Styling:
//   - Use lipgloss for terminal styling (colors, borders, etc.)
//   - Use ui_components.go helpers for common UI elements
//   - Respect terminal dimensions (m.width, m.height)
//
// Returns:
//   - String to display in the terminal
func (m Model) View() string {
	// Render the current screen, then layer the tutorial overlay (when
	// the player has an active step that matches the screen) via
	// ViewWithTutorial. renderTutorialOverlay no-ops when tutorials are
	// disabled or no step matches, so we can do this unconditionally.
	return m.ViewWithTutorial(m.viewScreen())
}

func (m Model) viewScreen() string {
	// Show error if present (but not on login screen)
	if m.err != nil && m.screen != ScreenLogin && m.screen != ScreenRegistration {
		return errorView(m.err.Error())
	}

	// Loading state (but not on login or registration screen)
	if m.player == nil && m.screen != ScreenLogin && m.screen != ScreenRegistration {
		return loadingView()
	}

	// Delegate to screen-specific view
	switch m.screen {
	case ScreenMainMenu:
		return m.viewMainMenu()
	case ScreenGame:
		return m.viewGame()
	case ScreenRegistration:
		return m.viewRegistration()
	case ScreenNavigation:
		return m.viewNavigation()
	case ScreenTrading:
		return m.viewTrading()
	case ScreenCargo:
		return m.viewCargo()
	case ScreenShipyard:
		return m.viewShipyard()
	case ScreenOutfitter:
		return m.viewOutfitter()
	case ScreenShipManagement:
		return m.viewShipManagement()
	case ScreenCombat:
		return m.viewCombat()
	case ScreenMissions:
		return m.viewMissions()
	case ScreenAchievements:
		return m.viewAchievements()
	case ScreenEncounter:
		return m.viewEncounter()
	case ScreenNews:
		return m.viewNews()
	case ScreenLeaderboards:
		return m.viewLeaderboards()
	case ScreenPlayers:
		return m.viewPlayers()
	case ScreenChat:
		return m.viewChat()
	case ScreenFactions:
		return m.viewFactions()
	case ScreenTrade:
		return m.viewTrade()
	case ScreenPvP:
		return m.viewPvP()
	case ScreenHelp:
		return m.viewHelp()
	case ScreenOutfitterEnhanced:
		return m.viewOutfitterEnhanced()
	case ScreenSettings:
		return m.viewSettings()
	case ScreenAdmin:
		return m.viewAdmin()
	case ScreenTutorial:
		return m.viewTutorial()
	case ScreenQuests:
		return m.viewQuests()
	case ScreenLogin:
		return m.viewLogin()
	case ScreenSpaceView:
		return m.viewSpaceView()
	case ScreenLanding:
		return m.viewLanding()
	case ScreenTradingEnhanced:
		return m.viewTradingEnhanced()
	case ScreenShipyardEnhanced:
		return m.viewShipyardEnhanced()
	case ScreenMissionBoardEnhanced:
		return m.viewMissionBoardEnhanced()
	case ScreenCombatEnhanced:
		return m.viewCombatEnhanced()
	case ScreenQuestBoardEnhanced:
		return m.viewQuestBoardEnhanced()
	case ScreenTradeRoutes:
		return m.viewTradeRoutes()
	case ScreenMail:
		return m.viewMail()
	case ScreenFleet:
		return m.viewFleet()
	case ScreenFriends:
		return m.viewFriends()
	case ScreenMarketplace:
		return m.viewMarketplace()
	case ScreenNotifications:
		return m.viewNotifications()
	case ScreenPilotRecord:
		return m.viewPilotRecord()
	case ScreenFactionWars:
		return m.viewFactionWars()
	case ScreenTerritoryMap:
		return m.viewTerritoryMap()
	default:
		return "Unknown screen"
	}
}

// ViewWithTutorial wraps screen content with tutorial overlay if active
func (m Model) ViewWithTutorial(content string) string {
	return m.renderTutorialOverlay(content)
}

// playerLoadedMsg is sent when player data has been loaded from the database.
//
// This message is the result of the async m.loadPlayer() command.
// It's processed in Update() to populate m.player and m.currentShip.
//
// Fields:
//   - player: The loaded player data (nil if error)
//   - ship: The player's current ship (nil if no ship or error)
//   - err: Any error that occurred during loading (nil if successful)
//
// Message Flow:
//   1. Init() or screen transition calls m.loadPlayer()
//   2. loadPlayer() returns a tea.Cmd that runs asynchronously
//   3. When complete, tea.Cmd sends playerLoadedMsg back to Update()
//   4. Update() handles playerLoadedMsg and updates m.player/m.currentShip
type playerLoadedMsg struct {
	player       *models.Player
	ship         *models.Ship
	system       *models.StarSystem
	achievements []*models.PlayerAchievement
	err          error
}

// loadPlayer loads player data from the database asynchronously,
// along with the player's current ship and star system, if any.
//
// This function returns a tea.Cmd that will:
//   1. Query the database for player data via playerRepo
//   2. Load the player's current ship via shipRepo (if they have one)
//   3. Load the player's current star system via systemRepo (if any)
//   4. Send a playerLoadedMsg with the results
//
// Ship and system lookups are best-effort — a newly-registered account may not
// have either set yet, and downstream screens are expected to handle nil.
//
// Thread Safety:
//   - Uses context.Background() for database operations
//   - Repositories are thread-safe and can be called from goroutines
//
// Returns:
//   - tea.Cmd that will send playerLoadedMsg when complete
func (m Model) loadPlayer() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		player, err := m.playerRepo.GetByID(ctx, m.playerID)
		if err != nil {
			return playerLoadedMsg{err: err}
		}

		var ship *models.Ship
		if player != nil && player.ShipID != uuid.Nil {
			ship, _ = m.shipRepo.GetByID(ctx, player.ShipID)
		}

		var system *models.StarSystem
		if player != nil && player.CurrentSystem != uuid.Nil {
			system, _ = m.systemRepo.GetSystemByID(ctx, player.CurrentSystem)
		}

		// Pre-load the player's achievement unlocks so checkAchievements
		// doesn't re-fire OnPlayerAchievement news articles for every
		// kill/trade/jump after reconnect. Best-effort: a failure here
		// just means the news feed sees duplicate unlocks for this
		// session, not a crash.
		var achievements []*models.PlayerAchievement
		if player != nil && m.achievementRepo != nil {
			achievements, _ = m.achievementRepo.LoadForPlayer(ctx, player.ID)
		}

		return playerLoadedMsg{player: player, ship: ship, system: system, achievements: achievements, err: nil}
	}
}

// changeScreen changes the current screen and returns a clear screen command.
//
// This is the standard way to transition between screens in the TUI.
// It updates m.screen and returns tea.ClearScreen to prevent terminal artifacts.
//
// Usage:
//   return m.changeScreen(ScreenCombat)
//
// Note: This is a helper method used by screen update functions. The actual
// screen transition logic is handled in the Update() function's switch statement.
//
// Parameters:
//   - screen: The new screen to display
//
// Returns:
//   - tea.ClearScreen command to clear the terminal before rendering new screen
func (m *Model) changeScreen(screen Screen) tea.Cmd {
	m.screen = screen
	// Clear screen to prevent artifacts when transitioning
	return tea.ClearScreen
}

// checkAchievements checks for newly unlocked achievements and queues them for display.
//
// This should be called after any player action that might unlock achievements:
//   - Enemy kills (combat victories)
//   - Trade transactions (buy/sell commodities)
//   - Mission completions
//   - Quest completions
//   - System jumps (exploration)
//   - Credits earned/spent
//   - Faction reputation changes
//
// Achievement Flow:
//   1. checkAchievements() queries achievementManager for new unlocks
//   2. New achievements are appended to m.pendingAchievements queue
//   3. getAchievementNotification() displays them one at a time
//   4. News articles are generated for notable achievements
//
// Thread Safety:
//   - Safe to call from any screen update function
//   - Achievement manager is thread-safe
//   - News manager is thread-safe
//
// Performance:
//   - Runs synchronously, but achievement checks are fast (in-memory)
//   - Only checks achievements that match the player's current stats
func (m *Model) checkAchievements() {
	if m.player == nil || m.achievementManager == nil {
		return
	}

	newUnlocks := m.achievementManager.CheckNewUnlocks(m.player)
	if len(newUnlocks) > 0 {
		m.pendingAchievements = append(m.pendingAchievements, newUnlocks...)

		// Generate news for notable achievements + persist the unlock
		// so it survives reconnect. Both calls are best-effort — the
		// in-memory manager has already flagged the unlock as consumed,
		// so a DB failure at worst loses the persistence for this
		// unlock until the next occurrence.
		ctx := context.Background()
		for _, achievement := range newUnlocks {
			if m.newsManager != nil {
				m.newsManager.OnPlayerAchievement(m.username, achievement)
			}
			if m.achievementRepo != nil {
				_ = m.achievementRepo.Unlock(ctx, m.player.ID, achievement.ID)
			}
		}
	}
}

// getAchievementNotification returns a notification message for pending achievements
//
// Returns empty string if no pending achievements
func (m *Model) getAchievementNotification() string {
	if len(m.pendingAchievements) == 0 {
		return ""
	}

	achievement := m.pendingAchievements[0]
	return fmt.Sprintf("%s Achievement Unlocked: %s (%s, %d pts)", achievement.Icon, achievement.Title, achievement.Rarity, achievement.Points)
}

// clearAchievementNotification removes the first pending achievement from the queue
func (m *Model) clearAchievementNotification() {
	if len(m.pendingAchievements) > 0 {
		m.pendingAchievements = m.pendingAchievements[1:]
	}
}

// leaderboardsRefreshedMsg is sent when leaderboards have been refreshed
type leaderboardsRefreshedMsg struct {
	success bool
}

// refreshLeaderboards updates all leaderboard rankings
//
// This fetches all players from the database and recalculates rankings
// across all categories. In a production system, this would be optimized
// with caching and incremental updates.
func (m Model) refreshLeaderboards() tea.Cmd {
	return func() tea.Msg {
		// For now, we'll simulate with just the current player
		// In a full implementation, we would fetch all players from the database
		// ctx := context.Background()
		// players, err := m.playerRepo.GetAll(ctx)
		// if err != nil {
		//     return leaderboardsRefreshedMsg{success: false}
		// }

		// For this demo, create a simulated player list with just the current player
		players := []*models.Player{}
		if m.player != nil {
			players = append(players, m.player)
		}

		// Update all leaderboards
		m.leaderboardManager.UpdateAllLeaderboards(players)

		return leaderboardsRefreshedMsg{success: true}
	}
}
