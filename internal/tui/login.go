// File: internal/tui/login.go
// Project: Terminal Velocity
// Description: Login screen - Authenticates players via username/password with ASCII branding
// Version: 2.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14
//
// The login screen is the entry point for returning players. It features:
// - Large ASCII art logo for Terminal Velocity
// - Username and password input fields with visual focus indicators
// - Login button to authenticate
// - Registration button to create new account
// - Async authentication with loading state
// - Error messages for failed login attempts

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// asciiLogo is the large ASCII art branding displayed at the top of the login screen.
// Uses box-drawing characters to create the Terminal Velocity title.
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

// loginModel contains the state for the login screen.
// It manages field focus, input values, and authentication state.
type loginModel struct {
	focusedField     int    // Current focused field: 0=username, 1=password, 2=login button, 3=register button
	username         string // Username input value
	password         string // Password input value (displayed as asterisks)
	showPassword     bool   // Whether to show password in plaintext (currently unused)
	error            string // Error message to display (empty if no error)
	isAuthenticating bool   // True while authentication request is in progress
}

// newLoginModel creates and initializes a new login screen model.
// Sets focus to the username field by default.
func newLoginModel() loginModel {
	return loginModel{
		focusedField: 0,
	}
}

// loginSuccessMsg is a BubbleTea message sent when authentication succeeds.
// Contains the authenticated player's ID and username.
type loginSuccessMsg struct {
	playerID uuid.UUID // Authenticated player's unique identifier
	username string    // Authenticated player's username
}

// loginFailureMsg is a BubbleTea message sent when authentication fails.
// Contains the error message to display to the user.
type loginFailureMsg struct {
	error string // Human-readable error message
}

// viewLogin renders the login screen with ASCII logo and input form.
//
// Layout:
//   - Top border
//   - ASCII logo (centered)
//   - Login panel (centered):
//     - Username field with focus indicator
//     - Password field (masked with asterisks)
//     - Login button
//     - OR separator
//     - Register button
//   - Error message area (if error present)
//   - Footer with key bindings
//
// Visual Features:
//   - Focused fields highlighted with special styling and cursor
//   - Password displayed as asterisks for security
//   - "Authenticating..." text shown during login process
func (m Model) viewLogin() string {
	// Set minimum width for proper layout
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

	// Username field. The content inside the inner panel must span exactly
	// panelWidth-2 cells, otherwise the inner right border lands off-column and
	// BubbleTea wraps the remainder onto the next row (visible as a ghost
	// duplicate of the row below).
	usernameLabel := "Username: "
	usernameFocused := m.loginModel.focusedField == 0
	usernameField := m.loginModel.username
	if usernameFocused {
		usernameField = HighlightStyle.Render("[" + PadRight(usernameField+"_", 27) + "]")
	} else {
		usernameField = "[" + PadRight(usernameField, 27) + "]"
	}
	usernameRow := PadRight("  "+usernameLabel+usernameField, panelWidth-2)

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(usernameRow)
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

	// Password field. Same padding rule as Username.
	passwordLabel := "Password: "
	passwordFocused := m.loginModel.focusedField == 1
	passwordField := strings.Repeat("*", len([]rune(m.loginModel.password)))
	if passwordFocused {
		passwordField = HighlightStyle.Render("[" + PadRight(passwordField+"_", 27) + "]")
	} else {
		passwordField = "[" + PadRight(passwordField, 27) + "]"
	}
	passwordRow := PadRight("  "+passwordLabel+passwordField, panelWidth-2)

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(passwordRow)
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

	footerText := " [Tab/↑/↓] Navigate  [Enter] Select  [Ctrl+C] Quit"
	sb.WriteString(BoxVertical)
	sb.WriteString(PadRight(footerText, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// updateLogin handles input and state updates for the login screen.
//
// Key Bindings:
//   - ctrl+c: Quit application
//   - tab/down: Cycle focus to next field
//   - up: Cycle focus to previous field
//   - enter: Submit login or switch to registration
//   - backspace: Delete character from current input field
//   - Any printable character: Add to current input field
//
// Authentication Flow:
//   1. User fills username and password
//   2. User presses enter on login button
//   3. isAuthenticating flag set, async authentication starts
//   4. On success: loginSuccessMsg received, loads player data
//   5. On failure: loginFailureMsg received, shows error
//
// State Transitions:
//   - Login success -> Loads player -> ScreenMainMenu
//   - Register button -> ScreenRegistration
func (m Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Don't process input while authenticating (wait for response)
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
				return m, tea.ClearScreen
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
			// Append printable text into the focused field.
			//
			// We used to only accept msg.String() of length 1, which silently
			// dropped pastes (BubbleTea delivers paste events as multi-rune
			// KeyMsg strings) and composed characters. Now we accept any msg
			// whose Runes slice is non-empty and contains only printable
			// runes — keeps us robust to pastes while still rejecting named
			// control keys like "esc", "ctrl+x", "f1", etc. which have
			// multi-character String() but empty/absent Runes.
			if s, ok := printableRuneString(msg); ok {
				switch m.loginModel.focusedField {
				case 0:
					m.loginModel.username += s
					m.loginModel.error = ""
				case 1:
					m.loginModel.password += s
					m.loginModel.error = ""
				}
			}
			return m, nil
		}

	case playerLoadedMsg:
		// Player data loaded after successful login
		m.player = msg.player
		m.currentShip = msg.ship
		m.currentSystem = msg.system
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

		// Register this session with the shared chat manager so
		// SendGlobalMessage's fanout reaches us. The top-level
		// playerLoadedMsg handler also does this, but logins path
		// through THIS handler first and short-circuits before the
		// top-level one runs — without this line, a just-logged-in
		// session sends into a manager that doesn't know about it.
		if m.player != nil && m.chatManager != nil {
			m.chatManager.GetOrCreateHistory(m.player.ID)
		}

		// Seed achievement manager from persisted unlocks so reconnect
		// doesn't re-fire notifications (mirror of the top-level path).
		if m.achievementManager != nil && len(msg.achievements) > 0 {
			m.achievementManager.LoadUnlocked(msg.achievements)
		}

		// Register the tutorial progress + fire the first-login trigger.
		// The tutorial manager is currently per-session (pre-5A.1) so
		// this effectively restarts the tutorial on every login for
		// beta testers — fine, because the in-memory manager has no
		// persisted progress to carry between sessions yet.
		m.InitializeTutorials()

		// Transition to main menu
		m.screen = ScreenMainMenu
		return m, nil
	}

	return m, nil
}

// handleLogin validates login inputs and initiates authentication.
// Performs client-side validation before sending credentials to database:
//   - Ensures username is not empty
//   - Ensures password is not empty
//
// Sets isAuthenticating flag and calls authenticateUser for async processing.
func (m Model) handleLogin() (Model, tea.Cmd) {
	// Validate username is present
	if m.loginModel.username == "" {
		m.loginModel.error = "Please enter your username"
		return m, nil
	}

	// Validate password is present
	if m.loginModel.password == "" {
		m.loginModel.error = "Please enter your password"
		return m, nil
	}

	// Start authentication process
	m.loginModel.isAuthenticating = true
	m.loginModel.error = ""

	return m, m.authenticateUser()
}

// authenticateUser performs async authentication against the database.
// Returns a tea.Cmd that executes the authentication in the background.
//
// Success Flow:
//   - Returns loginSuccessMsg with player ID and username
//   - Triggers player data loading
//
// Failure Flow:
//   - Returns loginFailureMsg with error message
//   - Error displayed on screen, password cleared
//
// Database Errors:
//   - ErrInvalidCredentials: User-friendly message
//   - Other errors: Technical error message
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
