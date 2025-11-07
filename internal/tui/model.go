package tui

import (
	"context"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/achievements"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/leaderboards"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/news"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/presence"
	"github.com/charmbracelet/bubbletea"
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
	ScreenSettings
	ScreenRegistration
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

	// Screen dimensions
	width  int
	height int

	// Sub-models for different screens
	mainMenu       mainMenuModel
	gameView       gameViewModel
	registration   registrationModel
	navigation     navigationModel
	trading        tradingModel
	cargo          cargoModel
	shipyard       shipyardModel
	outfitter      outfitterModel
	shipManagement shipManagementModel
	combat         combatModel
	missions       missionsModel
	achievementsUI achievementsModel
	encounterModel encounterModel
	newsModel      newsModel
	leaderboardsModel leaderboardsModel
	playersModel   playersModel

	// Achievement tracking
	achievementManager *achievements.Manager
	pendingAchievements []*models.Achievement // Newly unlocked, pending display

	// News system
	newsManager *news.Manager

	// Leaderboards system
	leaderboardManager *leaderboards.Manager

	// Presence system
	presenceManager *presence.Manager

	// Error message
	err error
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
	return m.loadPlayer()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Only quit from main menu
			if m.screen == ScreenMainMenu {
				return m, tea.Quit
			}
			// From other screens, go back to main menu
			m.screen = ScreenMainMenu
			return m, nil
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
	default:
		return m, nil
	}
}

// View renders the model
func (m Model) View() string {
	// Show error if present
	if m.err != nil {
		return errorView(m.err.Error())
	}

	// Loading state
	if m.player == nil {
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
	default:
		return "Unknown screen"
	}
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
	return nil
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
	return achievement.Icon + " Achievement Unlocked: " + achievement.Title + " (" + string(achievement.Rarity) + ", " + string(achievement.Points) + " pts)"
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
