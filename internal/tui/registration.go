// File: internal/tui/registration.go
// Project: Terminal Velocity
// Description: Registration screen - New player account creation with validation
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The registration screen guides new players through account creation with:
// - Multi-step wizard interface (welcome, email, password, confirm, creating, success)
// - Email validation (required or optional based on server config)
// - Password strength validation and visual feedback
// - Password confirmation to prevent typos
// - Async account creation with loading state
// - Automatic transition back to login screen on success
//
// Security Features:
// - ANSI escape code stripping to prevent injection attacks
// - Input length limits (email: 254 chars, password: 128 chars)
// - Password complexity requirements enforced
// - Visual password strength indicator

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/validation"
	tea "github.com/charmbracelet/bubbletea"
)

// registrationModel contains the state for the registration wizard.
// It manages the multi-step flow, input values, and creation state.
type registrationModel struct {
	step         int    // Current step: 0=welcome, 1=email, 2=password, 3=confirm, 4=creating, 5=success
	email        string // Email input value
	password     string // Password input value (displayed as bullets)
	confirmPass  string // Password confirmation input value
	error        string // Error message to display (empty if no error)
	cursor       int    // Cursor position (currently unused)
	requireEmail bool   // Whether email is required (from server config)
	sshKeyData   []byte // SSH public key data (deprecated - SSH key auth removed)
}

// registrationCompleteMsg is a BubbleTea message sent when account creation completes.
// Contains success status and any error that occurred.
type registrationCompleteMsg struct {
	success bool  // True if account was created successfully
	err     error // Error if creation failed, nil on success
}

// switchToLoginMsg is a BubbleTea message that triggers transition to login screen.
// Sent after successful registration and brief success message display.
type switchToLoginMsg struct{}

// newRegistrationModel creates and initializes a new registration screen model.
// Parameters:
//   - requireEmail: If true, email field is required; if false, email is optional
//   - sshKeyData: SSH public key (deprecated, should be nil)
func newRegistrationModel(requireEmail bool, sshKeyData []byte) registrationModel {
	return registrationModel{
		step:         0,
		requireEmail: requireEmail,
		sshKeyData:   sshKeyData,
	}
}

// updateRegistration handles input and state updates for the registration screen.
//
// Key Bindings:
//   - ctrl+c: Quit application
//   - esc: Cancel registration and quit
//   - enter: Proceed to next step or submit
//   - backspace: Delete character from current input field
//   - Any printable character: Add to current input field
//
// Registration Flow:
//   Step 0 (Welcome): Shows username and auth method, press enter to continue
//   Step 1 (Email): Enter email address (validated)
//   Step 2 (Password): Enter password (shows strength indicator)
//   Step 3 (Confirm): Re-enter password (must match)
//   Step 4 (Creating): Async account creation in progress
//   Step 5 (Success): Account created, auto-transitions to login
//
// Message Handling:
//   - registrationCompleteMsg: Account creation result
//   - switchToLoginMsg: Triggers return to login screen
func (m Model) updateRegistration(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Cancel registration and quit
			return m, tea.Quit

		case "enter":
			// Proceed to next step
			return m.handleRegistrationStep()

		case "backspace":
			// Delete last character from current field
			return m.handleRegistrationBackspace()

		default:
			// Add character to current input field
			return m.handleRegistrationInput(msg.String())
		}

	case registrationCompleteMsg:
		if msg.success {
			// Registration successful - go back to login screen
			m.registration.step = 5 // Success screen
			// Wait a moment then go to login
			return m, func() tea.Msg {
				return switchToLoginMsg{}
			}
		} else {
			m.registration.error = fmt.Sprintf("Registration failed: %v", msg.err)
			return m, nil
		}

	case switchToLoginMsg:
		// Switch to login screen after successful registration
		m.screen = ScreenLogin
		m.registration = newRegistrationModel(false, nil) // Reset registration state
		return m, tea.ClearScreen
	}

	return m, nil
}

// handleRegistrationStep processes the current step and advances to the next.
// Each step performs validation before proceeding:
//
// Step 0 (Welcome): No validation, just advances
// Step 1 (Email): Validates email format (required or optional per config)
// Step 2 (Password): Validates password complexity requirements
// Step 3 (Confirm): Validates passwords match, then creates account
// Step 4 (Creating): Waiting for async account creation
// Step 5 (Success): Quits to return to login
func (m Model) handleRegistrationStep() (Model, tea.Cmd) {
	reg := &m.registration

	switch reg.step {
	case 0: // Welcome screen
		// Advance to email entry
		reg.step = 1
		reg.error = ""

	case 1: // Email input
		// Validate email (required or optional based on config)
		var emailErr error
		if reg.requireEmail {
			emailErr = validation.ValidateEmail(reg.email)
		} else {
			emailErr = validation.ValidateEmailOptional(reg.email)
		}

		if emailErr != nil {
			reg.error = emailErr.Error()
			return m, nil
		}
		reg.error = ""

		// SSH key authentication has been removed - always go to password
		reg.step = 2

	case 2: // Password input
		// Validate password complexity (min 8 chars, uppercase, lowercase, number)
		if err := validation.ValidatePassword(reg.password); err != nil {
			reg.error = err.Error()
			return m, nil
		}
		reg.error = ""
		reg.step = 3

	case 3: // Confirm password
		// Ensure passwords match
		if reg.password != reg.confirmPass {
			reg.error = "Passwords do not match"
			reg.confirmPass = "" // Clear confirmation to retry
			return m, nil
		}
		reg.error = ""
		reg.step = 4
		return m, m.createAccount() // Start async account creation

	case 4: // Creating account (waiting)
		// Do nothing, wait for registrationCompleteMsg

	case 5: // Success
		// Account created, quit to return to login
		return m, tea.Quit
	}

	return m, nil
}

// handleRegistrationBackspace deletes the last character from the current input field.
// Works on steps 1 (email), 2 (password), and 3 (confirm password).
// Clears any error message when editing.
func (m Model) handleRegistrationBackspace() (Model, tea.Cmd) {
	reg := &m.registration

	switch reg.step {
	case 1: // Email field
		if len(reg.email) > 0 {
			reg.email = reg.email[:len(reg.email)-1]
		}
	case 2: // Password field
		if len(reg.password) > 0 {
			reg.password = reg.password[:len(reg.password)-1]
		}
	case 3: // Confirm password field
		if len(reg.confirmPass) > 0 {
			reg.confirmPass = reg.confirmPass[:len(reg.confirmPass)-1]
		}
	}

	// Clear error when user edits input
	reg.error = ""
	return m, nil
}

// handleRegistrationInput adds a character to the current input field.
// Applies security measures:
//   - Only accepts single printable characters
//   - Filters dangerous control characters (except escape for ANSI stripping)
//   - Enforces length limits (email: 254, password: 128)
//   - Strips ANSI escape codes to prevent injection attacks
//
// RFC 5321 specifies max email length of 254 characters.
// Password max of 128 is reasonable for bcrypt hashing.
func (m Model) handleRegistrationInput(input string) (Model, tea.Cmd) {
	reg := &m.registration

	// Only handle single characters
	if len(input) != 1 {
		return m, nil
	}

	// Filter out dangerous control characters but allow escape sequences for ANSI stripping
	// We allow \x1b (escape) because StripANSI will remove the full ANSI sequence
	char := input[0]
	if (char < 32 && char != 27) || char == 127 {
		// Filter: null byte, backspace, DEL, etc. but NOT escape (\x1b = 27)
		return m, nil
	}

	switch reg.step {
	case 1: // Email input
		// Limit email length to prevent memory exhaustion (RFC 5321 max is 254)
		if len(reg.email) < 254 {
			reg.email += input
			// Strip ANSI escape codes to prevent injection attacks
			reg.email = validation.StripANSI(reg.email)
			reg.error = ""
		}
	case 2: // Password input
		// Limit password length to prevent memory exhaustion (reasonable max is 128)
		if len(reg.password) < 128 {
			reg.password += input
			// Strip ANSI escape codes to prevent injection attacks
			reg.password = validation.StripANSI(reg.password)
			reg.error = ""
		}
	case 3: // Confirm password input
		// Limit confirm password length to match password max
		if len(reg.confirmPass) < 128 {
			reg.confirmPass += input
			// Strip ANSI escape codes to prevent injection attacks
			reg.confirmPass = validation.StripANSI(reg.confirmPass)
			reg.error = ""
		}
	}

	return m, nil
}

// createAccount performs async account creation in the database.
// Returns a tea.Cmd that executes the creation in the background.
//
// Two Creation Paths:
//  1. SSH Key (deprecated): Creates account without password, then adds SSH key
//  2. Password (current): Creates account with hashed password and email
//
// Success: Returns registrationCompleteMsg with success=true
// Failure: Returns registrationCompleteMsg with success=false and error
//
// Database Operations:
//   - PlayerRepository.CreateWithEmail: Creates player with password
//   - PlayerRepository.CreateWithSSHKey: Creates player without password (deprecated)
//   - SSHKeyRepository.AddKey: Adds SSH key to player (deprecated)
func (m Model) createAccount() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var err error
		if len(m.registration.sshKeyData) > 0 {
			// Create account with SSH key (deprecated path)
			_, err = m.playerRepo.CreateWithSSHKey(ctx, m.username, m.registration.email)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}

			// Add the SSH key to the new account
			player, err := m.playerRepo.GetByUsername(ctx, m.username)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}

			_, err = m.sshKeyRepo.AddKey(ctx, player.ID, string(m.registration.sshKeyData))
			if err != nil {
				return registrationCompleteMsg{success: false, err: fmt.Errorf("account created but failed to add SSH key: %w", err)}
			}
		} else {
			// Create account with password (current path)
			_, err = m.playerRepo.CreateWithEmail(ctx, m.username, m.registration.password, m.registration.email)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}
		}

		return registrationCompleteMsg{success: true}
	}
}

// viewRegistration renders the registration wizard screen.
//
// Display varies by current step:
//   Step 0 (Welcome): Shows username, auth method, continue prompt
//   Step 1 (Email): Email input field with cursor
//   Step 2 (Password): Password input (bullets), requirements, strength indicator
//   Step 3 (Confirm): Password confirmation input (bullets)
//   Step 4 (Creating): Loading message
//   Step 5 (Success): Success message
//
// All steps show:
//   - Title banner
//   - Error message (if present)
//   - Context-appropriate help text
//
// Password Strength Colors:
//   - Red (0-29): Weak
//   - Yellow (30-49): Fair
//   - Cyan (50-69): Good
//   - Green (70+): Strong
func (m Model) viewRegistration() string {
	reg := m.registration

	s := titleStyle.Render("=== New Account Registration ===") + "\n\n"

	if reg.error != "" {
		s += errorStyle.Render("⚠ "+reg.error) + "\n\n"
	}

	switch reg.step {
	case 0: // Welcome
		authMethod := "password"
		if len(reg.sshKeyData) > 0 {
			authMethod = "SSH key"
		}

		s += "Welcome to Terminal Velocity!\n\n"
		s += fmt.Sprintf("Username: %s\n", statsStyle.Render(m.username))
		s += fmt.Sprintf("Auth Method: %s\n\n", authMethod)
		s += "Let's set up your account.\n\n"
		s += helpStyle.Render("Press Enter to continue  •  ESC to cancel")

	case 1: // Email
		if reg.requireEmail {
			s += "Email address (required):\n"
		} else {
			s += "Email address (optional, press Enter to skip):\n"
		}
		s += "> " + reg.email + "█\n\n"
		s += helpStyle.Render("Type your email  •  Enter to continue  •  ESC to cancel")

	case 2: // Password
		s += "Create a secure password:\n"
		s += "Requirements:\n"
		s += "  • At least 8 characters\n"
		s += "  • At least one uppercase letter\n"
		s += "  • At least one lowercase letter\n"
		s += "  • At least one number\n\n"
		s += "> " + strings.Repeat("•", len(reg.password)) + "█\n\n"

		// Show password strength if there's input
		if len(reg.password) > 0 {
			score, strength := validation.GetPasswordStrength(reg.password)
			var strengthColor string
			switch {
			case score < 30:
				strengthColor = "\x1b[31m" // Red
			case score < 50:
				strengthColor = "\x1b[33m" // Yellow
			case score < 70:
				strengthColor = "\x1b[36m" // Cyan
			default:
				strengthColor = "\x1b[32m" // Green
			}
			s += fmt.Sprintf("Strength: %s%s\x1b[0m (%d/100)\n\n", strengthColor, strength, score)
		}

		s += helpStyle.Render("Type your password  •  Enter to continue  •  ESC to cancel")

	case 3: // Confirm password
		s += "Confirm your password:\n"
		s += "> " + strings.Repeat("•", len(reg.confirmPass)) + "█\n\n"
		s += helpStyle.Render("Retype your password  •  Enter to continue  •  ESC to cancel")

	case 4: // Creating
		s += "Creating your account...\n\n"
		s += "Please wait..."

	case 5: // Success
		s += "✓ Account created successfully!\n\n"
		s += "Returning to login screen..."
	}

	return s
}
