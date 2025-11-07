package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/s0v3r1gn/terminal-velocity/internal/database"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
)

// Screen represents different game screens
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenGame
	ScreenNavigation
	ScreenTrading
	ScreenShipyard
	ScreenMissions
	ScreenSettings
	ScreenRegistration
)

// Model is the main TUI model
type Model struct {
	// Current screen
	screen Screen

	// Player data
	player       *models.Player
	playerID     uuid.UUID
	username     string
	currentShip  *models.Ship

	// Database repositories
	playerRepo   *database.PlayerRepository
	systemRepo   *database.SystemRepository
	sshKeyRepo   *database.SSHKeyRepository

	// Screen dimensions
	width        int
	height       int

	// Sub-models for different screens
	mainMenu     mainMenuModel
	gameView     gameViewModel
	registration registrationModel
	navigation   navigationModel

	// Error message
	err          error
}

// NewModel creates a new TUI model
func NewModel(
	playerID uuid.UUID,
	username string,
	playerRepo *database.PlayerRepository,
	systemRepo *database.SystemRepository,
	sshKeyRepo *database.SSHKeyRepository,
) Model {
	return Model{
		screen:     ScreenMainMenu,
		playerID:   playerID,
		username:   username,
		playerRepo: playerRepo,
		systemRepo: systemRepo,
		sshKeyRepo: sshKeyRepo,
		width:      80,
		height:     24,
		mainMenu:   newMainMenuModel(),
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
) Model {
	return Model{
		screen:       ScreenRegistration,
		playerID:     uuid.Nil,
		username:     username,
		playerRepo:   playerRepo,
		systemRepo:   systemRepo,
		sshKeyRepo:   sshKeyRepo,
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
		m.err = msg.err
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
	default:
		return "Unknown screen"
	}
}

// playerLoadedMsg is sent when player data is loaded
type playerLoadedMsg struct {
	player *models.Player
	err    error
}

// loadPlayer loads player data from the database
func (m Model) loadPlayer() tea.Cmd {
	return func() tea.Msg {
		player, err := m.playerRepo.GetByID(nil, m.playerID)
		return playerLoadedMsg{player: player, err: err}
	}
}

// changeScreen changes the current screen
func (m *Model) changeScreen(screen Screen) tea.Cmd {
	m.screen = screen
	return nil
}
