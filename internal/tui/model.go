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
	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/friends"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/leaderboards"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/mail"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/marketplace"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/missions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/news"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/notifications"
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

	// ===== Database Repositories =====
	// Repositories provide typed CRUD operations for database access
	// All database operations should go through repositories, never direct SQL

	playerRepo *database.PlayerRepository // Player accounts and stats
	systemRepo *database.SystemRepository // Star systems and connections
	sshKeyRepo *database.SSHKeyRepository // SSH public keys for authentication
	shipRepo   *database.ShipRepository   // Player ships and equipment
	marketRepo *database.MarketRepository // Market prices and commodities
	mailRepo   *database.MailRepository   // Player mail system
	socialRepo *database.SocialRepository // Friends, blocks, etc.
	itemRepo   *database.ItemRepository   // Items and equipment

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
	achievementsUI       achievementsModel         // Achievement tracking
	encounterModel       encounterModel            // Random encounters
	newsModel            newsModel                 // News articles
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
	navigationEnhanced   navigationEnhancedModel   // Enhanced navigation
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
	fleetManager *fleet.Manager,
	mailManager *mail.Manager,
	notificationsManager *notifications.Manager,
	friendsManager *friends.Manager,
	marketplaceManager *marketplace.Manager,
) Model {
	return Model{
		screen:              ScreenMainMenu,
		playerID:            playerID,
		username:            username,
		playerRepo:          playerRepo,
		systemRepo:          systemRepo,
		sshKeyRepo:          sshKeyRepo,
		shipRepo:            shipRepo,
		marketRepo:          marketRepo,
		mailRepo:            mailRepo,
		socialRepo:          socialRepo,
		itemRepo:            itemRepo,
		width:               80,
		height:              24,
		mainMenu:            newMainMenuModel(),
		trading:             newTradingModel(),
		cargo:               newCargoModel(),
		shipyard:            newShipyardModel(),
		outfitter:           newOutfitterModel(),
		shipManagement:      newShipManagementModel(),
		combat:              newCombatModel(),
		missions:            newMissionsModel(),
		achievementsUI:      newAchievementsModel(),
		achievementManager:  achievements.NewManager(),
		pendingAchievements: []*models.Achievement{},
		encounterModel:      newEncounterModel(),
		newsModel:           newNewsModel(),
		newsManager:         news.NewManager(),
		leaderboardsModel:   newLeaderboardsModel(),
		leaderboardManager:  leaderboards.NewManager(),
		playersModel:        newPlayersModel(),
		presenceManager:     presence.NewManager(),
		chatModel:           newChatModel(),
		chatManager:         chat.NewManager(),
		fleetManager:        fleetManager,
		mailManager:         mailManager,
		notificationsManager: notificationsManager,
		friendsManager:      friendsManager,
		marketplaceManager:  marketplaceManager,
		factionsModel:       newFactionsModel(),
		factionManager:      factions.NewManager(),
		territoryManager:    territory.NewManager(),
		tradeModel:          newTradeModel(),
		tradeManager:        trade.NewManager(),
		pvpModel:            newPvPModel(),
		pvpManager:          pvp.NewManager(),
		helpModel:           newHelpModel(),
		encounterManager:    encounters.NewManager(),
		outfitterEnhanced:   newOutfitterEnhancedModel(),
		outfittingManager:   outfitting.NewManager(),
		settingsModel:       newSettingsModel(),
		settingsManager:     settings.NewManager(".config/terminal-velocity"),
		adminModel:          newAdminModel(),
		adminManager:        admin.NewManager(playerRepo),
		tutorialModel:       newTutorialModel(),
		tutorialManager:     tutorial.NewManager(),
		questsModel:         newQuestsModel(),
		questManager:        quests.NewManager(),
		missionManager:      missions.NewManager(),
		loginModel:          newLoginModel(),
		spaceView:           newSpaceViewModel(),
		landing:             newLandingModel(),
		tradingEnhanced:      newTradingEnhancedModel(),
		shipyardEnhanced:     newShipyardEnhancedModel(),
		missionBoardEnhanced: newMissionBoardEnhancedModel(),
		navigationEnhanced:   newNavigationEnhancedModel(),
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

// InitializeNews generates initial news articles
func (m *Model) InitializeNews() {
	if m.newsManager != nil {
		m.newsManager.GenerateInitialNews()
	}
}

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
) Model {
	return Model{
		screen:              ScreenLogin,
		playerID:            uuid.Nil,
		username:            "",
		playerRepo:          playerRepo,
		systemRepo:          systemRepo,
		sshKeyRepo:          sshKeyRepo,
		shipRepo:            shipRepo,
		marketRepo:          marketRepo,
		mailRepo:            mailRepo,
		socialRepo:          socialRepo,
		width:               80,
		height:              24,
		loginModel:          newLoginModel(),
		mainMenu:            newMainMenuModel(),
		trading:             newTradingModel(),
		cargo:               newCargoModel(),
		shipyard:            newShipyardModel(),
		outfitter:           newOutfitterModel(),
		shipManagement:      newShipManagementModel(),
		combat:              newCombatModel(),
		missions:            newMissionsModel(),
		achievementsUI:      newAchievementsModel(),
		achievementManager:  achievements.NewManager(),
		pendingAchievements: []*models.Achievement{},
		encounterModel:      newEncounterModel(),
		newsModel:           newNewsModel(),
		newsManager:         news.NewManager(),
		leaderboardsModel:   newLeaderboardsModel(),
		leaderboardManager:  leaderboards.NewManager(),
		playersModel:        newPlayersModel(),
		presenceManager:     presence.NewManager(),
		chatModel:           newChatModel(),
		chatManager:         chat.NewManager(),
		mailManager:         mail.NewManager(socialRepo),
		factionsModel:       newFactionsModel(),
		factionManager:      factions.NewManager(),
		territoryManager:    territory.NewManager(),
		tradeModel:          newTradeModel(),
		tradeManager:        trade.NewManager(),
		pvpModel:            newPvPModel(),
		pvpManager:          pvp.NewManager(),
		helpModel:           newHelpModel(),
		encounterManager:    encounters.NewManager(),
		outfitterEnhanced:   newOutfitterEnhancedModel(),
		outfittingManager:   outfitting.NewManager(),
		settingsModel:       newSettingsModel(),
		settingsManager:     settings.NewManager(".config/terminal-velocity"),
		adminModel:          newAdminModel(),
		adminManager:        admin.NewManager(playerRepo),
		tutorialModel:       newTutorialModel(),
		tutorialManager:     tutorial.NewManager(),
		questsModel:         newQuestsModel(),
		questManager:        quests.NewManager(),
		missionManager:      missions.NewManager(),
		registration:        newRegistrationModel(false, nil),
		spaceView:           newSpaceViewModel(),
		landing:             newLandingModel(),
		tradingEnhanced:     newTradingEnhancedModel(),
		shipyardEnhanced:    newShipyardEnhancedModel(),
		missionBoardEnhanced: newMissionBoardEnhancedModel(),
		navigationEnhanced:  newNavigationEnhancedModel(),
		combatEnhanced:      newCombatEnhancedModel(),
		questBoardEnhanced:  newQuestBoardEnhancedModel(),
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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case playerLoadedMsg:
		m.player = msg.player
		m.currentShip = msg.ship
		m.err = msg.err

		// Initialize presence when player loads
		if m.player != nil && m.err == nil {
			m.InitializePresence()
		}

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
	case ScreenNavigationEnhanced:
		return m.updateNavigationEnhanced(msg)
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
	case ScreenNavigationEnhanced:
		return m.viewNavigationEnhanced()
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
	player *models.Player
	ship   *models.Ship
	err    error
}

// loadPlayer loads player data from the database asynchronously.
//
// This function returns a tea.Cmd that will:
//   1. Query the database for player data via playerRepo
//   2. Load the player's current ship via shipRepo (if they have one)
//   3. Send a playerLoadedMsg with the results
//
// The actual database operations happen in a goroutine managed by BubbleTea,
// ensuring the UI remains responsive during the load.
//
// Error Handling:
//   - If player not found: returns playerLoadedMsg with err set
//   - If ship load fails: logs error but doesn't fail (ship may be nil)
//   - Errors are handled in Update() when playerLoadedMsg is received
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

		// Load player's ship if they have one
		var ship *models.Ship
		if player.ShipID != uuid.Nil {
			ship, err = m.shipRepo.GetByID(ctx, player.ShipID)
			if err != nil {
				// Log error but don't fail - player might not have a ship yet
				// In future, we should handle this better
			}
		}

		return playerLoadedMsg{player: player, ship: ship, err: nil}
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

		// Generate news for notable achievements
		if m.newsManager != nil {
			for _, achievement := range newUnlocks {
				m.newsManager.OnPlayerAchievement(m.username, achievement)
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
