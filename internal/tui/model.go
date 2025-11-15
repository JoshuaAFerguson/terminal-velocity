// File: internal/tui/model.go
// Project: Terminal Velocity
// Description: Terminal UI component for model with login screen and unauthenticated state support
// Version: 1.2.2
// Author: Joshua Ferguson
// Created: 2025-01-07

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

// Screen represents different game screens

type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenGame
	ScreenNavigation
	ScreenTrading
	ScreenCargo
	ScreenShipyard
	ScreenOutfitter
	ScreenShipManagement
	ScreenCombat
	ScreenMissions
	ScreenAchievements
	ScreenEncounter
	ScreenNews
	ScreenLeaderboards
	ScreenPlayers
	ScreenChat
	ScreenFactions
	ScreenTrade
	ScreenPvP
	ScreenHelp
	ScreenOutfitterEnhanced
	ScreenSettings
	ScreenAdmin
	ScreenTutorial
	ScreenQuests
	ScreenRegistration
	ScreenLogin
	ScreenSpaceView
	ScreenLanding
	ScreenTradingEnhanced
	ScreenShipyardEnhanced
	ScreenMissionBoardEnhanced
	ScreenNavigationEnhanced
	ScreenCombatEnhanced
	ScreenQuestBoardEnhanced
	ScreenTradeRoutes
	ScreenMail
	ScreenFleet
	ScreenFriends
	ScreenMarketplace
	ScreenNotifications
)

// Model is the main TUI model
type Model struct {
	// Current screen
	screen Screen

	// Player data
	player      *models.Player
	playerID    uuid.UUID
	username    string
	currentShip *models.Ship

	// Database repositories
	playerRepo *database.PlayerRepository
	systemRepo *database.SystemRepository
	sshKeyRepo *database.SSHKeyRepository
	shipRepo   *database.ShipRepository
	marketRepo *database.MarketRepository
	mailRepo   *database.MailRepository
	socialRepo *database.SocialRepository
	itemRepo   *database.ItemRepository

	// Screen dimensions
	width  int
	height int

	// Sub-models for different screens
	mainMenu          mainMenuModel
	gameView          gameViewModel
	registration      registrationModel
	navigation        navigationModel
	trading           tradingModel
	cargo             cargoModel
	shipyard          shipyardModel
	outfitter         outfitterModel
	shipManagement    shipManagementModel
	combat            combatModel
	missions          missionsModel
	achievementsUI    achievementsModel
	encounterModel    encounterModel
	newsModel         newsModel
	leaderboardsModel leaderboardsModel
	playersModel      playersModel
	chatModel         chatModel
	factionsModel     factionsModel
	tradeModel        tradeModel
	pvpModel          pvpModel
	helpModel         helpModel
	outfitterEnhanced outfitterEnhancedModel
	settingsModel     settingsModel
	adminModel        adminModel
	tutorialModel     tutorialModel
	questsModel       questsModel
	loginModel        loginModel
	spaceView         spaceViewModel
	landing               landingModel
	tradingEnhanced       tradingEnhancedModel
	shipyardEnhanced      shipyardEnhancedModel
	missionBoardEnhanced  missionBoardEnhancedModel
	navigationEnhanced    navigationEnhancedModel
	combatEnhanced        combatEnhancedModel
	questBoardEnhanced    questBoardEnhancedModel
	tradeRoutes           tradeRoutesState
	mail                  mailState
	fleet                 fleetState
	friends               friendsState
	marketplace           marketplaceState
	notifications         notificationsState

	// Achievement tracking
	achievementManager  *achievements.Manager
	pendingAchievements []*models.Achievement // Newly unlocked, pending display

	// News system
	newsManager *news.Manager

	// Leaderboards system
	leaderboardManager *leaderboards.Manager

	// Presence system
	presenceManager *presence.Manager

	// Chat system
	chatManager *chat.Manager

	// Mail system
	mailManager *mail.Manager

	// Fleet system
	fleetManager *fleet.Manager

	// Friends system
	friendsManager *friends.Manager

	// Notifications system
	notificationsManager *notifications.Manager

	// Marketplace system
	marketplaceManager *marketplace.Manager

	// Faction system
	factionManager *factions.Manager

	// Territory system
	territoryManager *territory.Manager

	// Trade system
	tradeManager *trade.Manager

	// PvP system
	pvpManager *pvp.Manager

	// Encounter system
	encounterManager *encounters.Manager

	// Outfitting system
	outfittingManager *outfitting.Manager

	// Settings system
	settingsManager *settings.Manager

	// Admin system
	adminManager *admin.Manager

	// Tutorial system
	tutorialManager *tutorial.Manager

	// Quest system
	questManager *quests.Manager

	// Mission system
	missionManager *missions.Manager

	// Error handling
	err               error
	errorMessage      string
	showErrorDialog   bool
	loadingOperation  string
	isLoading         bool
}

// NewModel creates a new TUI model
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

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Clear screen on initialization to prevent artifacts
	// If we're on the login screen, don't load player data yet
	if m.screen == ScreenLogin || m.screen == ScreenRegistration {
		return tea.ClearScreen
	}
	return tea.Batch(tea.ClearScreen, m.loadPlayer())
}

// Update handles messages and updates the model
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

// View renders the model
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

// playerLoadedMsg is sent when player data is loaded
type playerLoadedMsg struct {
	player *models.Player
	ship   *models.Ship
	err    error
}

// loadPlayer loads player data from the database
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

// changeScreen changes the current screen
func (m *Model) changeScreen(screen Screen) tea.Cmd {
	m.screen = screen
	// Clear screen to prevent artifacts when transitioning
	return tea.ClearScreen
}

// checkAchievements checks for newly unlocked achievements and queues them for display
//
// This should be called after any player action that might unlock achievements
// (kills, trades, mission completions, etc.)
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
