// File: internal/tui/login.go
// Project: Terminal Velocity
// Description: Functional login screen with authentication and ASCII logo
// Version: 2.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// ASCII logo for Terminal Velocity
const asciiLogo = `
                        ████████╗███████╗██████╗ ███╗   ███╗
                        ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║
                           ██║   █████╗  ██████╔╝██╔████╔██║
                           ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║
                           ██║   ███████╗██║  ██║██║ ╚═╝ ██║
                           ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝

                       ██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
                       ██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
                       ██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝
                       ╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝
                        ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║
                         ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝

                          A Multiplayer Space Trading Game
`

type loginModel struct {
	focusedField  int // 0: username, 1: password, 2: login button, 3: register button
	username      string
	password      string
	showPassword  bool
	error         string
	isAuthenticating bool
}

func newLoginModel() loginModel {
	return loginModel{
		focusedField: 0,
	}
}

// loginSuccessMsg is sent when login succeeds
type loginSuccessMsg struct {
	playerID uuid.UUID
	username string
}

// loginFailureMsg is sent when login fails
type loginFailureMsg struct {
	error string
}

func (m Model) viewLogin() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxTopRight + "\n")

	// Empty space
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// ASCII Logo (centered)
	logoLines := strings.Split(asciiLogo, "\n")
	for _, line := range logoLines {
		if line == "" {
			continue
		}
		sb.WriteString(BoxVertical)
		sb.WriteString(Center(line, width-2))
		sb.WriteString(BoxVertical + "\n")
	}

	// Empty space
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Login panel (centered)
	panelWidth := 50
	panelLeft := (width - panelWidth) / 2

	// Login panel top
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, panelWidth-2))
	sb.WriteString(BoxTopRight)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Title
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center("LOGIN TO YOUR ACCOUNT", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Separator
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, panelWidth-2))
	sb.WriteString(BoxCross)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Username field
	usernameLabel := "Username: "
	usernameFocused := m.loginModel.focusedField == 0
	usernameField := m.loginModel.username
	if usernameFocused {
		usernameField = HighlightStyle.Render("[" + PadRight(usernameField+"_", 27) + "]")
	} else {
		usernameField = "[" + PadRight(usernameField, 27) + "]"
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString("  " + usernameLabel + usernameField + "  ")
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Password field
	passwordLabel := "Password: "
	passwordFocused := m.loginModel.focusedField == 1
	passwordField := strings.Repeat("*", len(m.loginModel.password))
	if passwordFocused {
		passwordField = HighlightStyle.Render("[" + PadRight(passwordField+"_", 27) + "]")
	} else {
		passwordField = "[" + PadRight(passwordField, 27) + "]"
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString("  " + passwordLabel + passwordField + "  ")
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Login Button
	btnLogin := "[ Login ]"
	if m.loginModel.focusedField == 2 {
		btnLogin = HighlightStyle.Render(btnLogin)
	}
	if m.loginModel.isAuthenticating {
		btnLogin = "[ Authenticating... ]"
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center(btnLogin, panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Separator OR
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center("─────────────── OR ───────────────────", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Register button
	btnRegister := "[ Create New Account ]"
	if m.loginModel.focusedField == 3 {
		btnRegister = HighlightStyle.Render(btnRegister)
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center(btnRegister, panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty line
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Panel bottom
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, panelWidth-2))
	sb.WriteString(BoxBottomRight)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	// Empty space
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Error message if present
	if m.loginModel.error != "" {
		sb.WriteString(BoxVertical)
		sb.WriteString(Center(ErrorStyle.Render(m.loginModel.error), width-2))
		sb.WriteString(BoxVertical + "\n")
		sb.WriteString(BoxVertical)
		sb.WriteString(strings.Repeat(" ", width-2))
		sb.WriteString(BoxVertical + "\n")
	} else {
		// Empty space
		sb.WriteString(BoxVertical)
		sb.WriteString(strings.Repeat(" ", width-2))
		sb.WriteString(BoxVertical + "\n")
	}

	// Footer
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(" [Tab/↑/↓] Navigate  [Enter] Select  [Ctrl+C] Quit")
	sb.WriteString(strings.Repeat(" ", width-len(" [Tab/↑/↓] Navigate  [Enter] Select  [Ctrl+C] Quit")-3))
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

func (m Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Don't process input while authenticating
	if m.loginModel.isAuthenticating {
		switch msg := msg.(type) {
		case loginSuccessMsg:
			// Login successful - load player and transition to game
			m.loginModel.isAuthenticating = false
			m.playerID = msg.playerID
			m.username = msg.username
			return m, m.loadPlayer()
		case loginFailureMsg:
			// Login failed - show error
			m.loginModel.isAuthenticating = false
			m.loginModel.error = msg.error
			m.loginModel.password = "" // Clear password on failure
			return m, nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "tab", "down":
			// Cycle forward through fields
			m.loginModel.focusedField = (m.loginModel.focusedField + 1) % 4
			m.loginModel.error = ""
			return m, nil

		case "up":
			// Cycle backward through fields
			m.loginModel.focusedField = (m.loginModel.focusedField - 1 + 4) % 4
			m.loginModel.error = ""
			return m, nil

		case "enter":
			// Handle field submission
			if m.loginModel.focusedField == 2 {
				// Login button
				return m.handleLogin()
			} else if m.loginModel.focusedField == 3 {
				// Register button
				m.screen = ScreenRegistration
				return m, nil
			}
			return m, nil

		case "backspace":
			// Handle backspace for text input fields
			if m.loginModel.focusedField == 0 {
				// Username field
				if len(m.loginModel.username) > 0 {
					m.loginModel.username = m.loginModel.username[:len(m.loginModel.username)-1]
				}
			} else if m.loginModel.focusedField == 1 {
				// Password field
				if len(m.loginModel.password) > 0 {
					m.loginModel.password = m.loginModel.password[:len(m.loginModel.password)-1]
				}
			}
			m.loginModel.error = ""
			return m, nil

		default:
			// Handle text input for username/password fields
			if len(msg.String()) == 1 {
				if m.loginModel.focusedField == 0 {
					// Username field
					m.loginModel.username += msg.String()
					m.loginModel.error = ""
				} else if m.loginModel.focusedField == 1 {
					// Password field
					m.loginModel.password += msg.String()
					m.loginModel.error = ""
				}
			}
			return m, nil
		}

	case playerLoadedMsg:
		// Player data loaded after successful login
		m.player = msg.player
		m.currentShip = msg.ship
		m.err = msg.err

		if m.err != nil {
			// Error loading player data
			m.loginModel.error = fmt.Sprintf("Error loading player: %v", m.err)
			m.playerID = uuid.Nil
			m.username = ""
			return m, nil
		}

		// Initialize presence when player loads
		if m.player != nil {
			m.InitializePresence()
		}

		// Transition to main menu
		m.screen = ScreenMainMenu
		return m, nil
	}

	return m, nil
}

// handleLogin authenticates the user
func (m Model) handleLogin() (Model, tea.Cmd) {
	// Validate input
	if m.loginModel.username == "" {
		m.loginModel.error = "Please enter your username"
		return m, nil
	}
	if m.loginModel.password == "" {
		m.loginModel.error = "Please enter your password"
		return m, nil
	}

	// Start authentication
	m.loginModel.isAuthenticating = true
	m.loginModel.error = ""

	return m, m.authenticateUser()
}

// authenticateUser performs the authentication
func (m Model) authenticateUser() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Authenticate user
		player, err := m.playerRepo.Authenticate(ctx, m.loginModel.username, m.loginModel.password)
		if err != nil {
			if err == database.ErrInvalidCredentials {
				return loginFailureMsg{error: "Invalid username or password"}
			}
			return loginFailureMsg{error: fmt.Sprintf("Authentication error: %v", err)}
		}

		return loginSuccessMsg{
			playerID: player.ID,
			username: player.Username,
		}
	}
}
