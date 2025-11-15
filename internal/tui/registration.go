// File: internal/tui/registration.go
// Project: Terminal Velocity
// Description: Terminal UI component for registration with login screen integration
// Version: 1.1.1
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/validation"
	tea "github.com/charmbracelet/bubbletea"
)

type registrationModel struct {
	step         int // 0: welcome, 1: email, 2: password, 3: confirm password, 4: creating
	email        string
	password     string
	confirmPass  string
	error        string
	cursor       int
	requireEmail bool
	sshKeyData   []byte // If registering with SSH key (deprecated - SSH key auth removed)
}

type registrationCompleteMsg struct {
	success bool
	err     error
}

type switchToLoginMsg struct{}

func newRegistrationModel(requireEmail bool, sshKeyData []byte) registrationModel {
	return registrationModel{
		step:         0,
		requireEmail: requireEmail,
		sshKeyData:   sshKeyData,
	}
}

func (m Model) updateRegistration(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Cancel registration
			return m, tea.Quit

		case "enter":
			return m.handleRegistrationStep()

		case "backspace":
			return m.handleRegistrationBackspace()

		default:
			// Add character to current field
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

func (m Model) handleRegistrationStep() (Model, tea.Cmd) {
	reg := &m.registration

	switch reg.step {
	case 0: // Welcome screen
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
		// Validate password complexity
		if err := validation.ValidatePassword(reg.password); err != nil {
			reg.error = err.Error()
			return m, nil
		}
		reg.error = ""
		reg.step = 3

	case 3: // Confirm password
		if reg.password != reg.confirmPass {
			reg.error = "Passwords do not match"
			reg.confirmPass = ""
			return m, nil
		}
		reg.error = ""
		reg.step = 4
		return m, m.createAccount()

	case 4: // Creating account (waiting)
		// Do nothing, wait for result

	case 5: // Success
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleRegistrationBackspace() (Model, tea.Cmd) {
	reg := &m.registration

	switch reg.step {
	case 1: // Email
		if len(reg.email) > 0 {
			reg.email = reg.email[:len(reg.email)-1]
		}
	case 2: // Password
		if len(reg.password) > 0 {
			reg.password = reg.password[:len(reg.password)-1]
		}
	case 3: // Confirm password
		if len(reg.confirmPass) > 0 {
			reg.confirmPass = reg.confirmPass[:len(reg.confirmPass)-1]
		}
	}

	reg.error = ""
	return m, nil
}

func (m Model) handleRegistrationInput(input string) (Model, tea.Cmd) {
	reg := &m.registration

	// Only handle printable characters
	if len(input) != 1 {
		return m, nil
	}

	// Filter out control characters and non-printable characters
	char := input[0]
	if char < 32 || char == 127 {
		return m, nil
	}

	switch reg.step {
	case 1: // Email
		// Limit email length to prevent memory exhaustion (RFC 5321 max is 254)
		if len(reg.email) < 254 {
			reg.email += input
			reg.error = ""
		}
	case 2: // Password
		// Limit password length to prevent memory exhaustion (reasonable max is 128)
		if len(reg.password) < 128 {
			reg.password += input
			reg.error = ""
		}
	case 3: // Confirm password
		// Limit confirm password length to match password max
		if len(reg.confirmPass) < 128 {
			reg.confirmPass += input
			reg.error = ""
		}
	}

	return m, nil
}

func (m Model) createAccount() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var err error
		if len(m.registration.sshKeyData) > 0 {
			// Create account with SSH key
			_, err = m.playerRepo.CreateWithSSHKey(ctx, m.username, m.registration.email)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}

			// Add the SSH key
			player, err := m.playerRepo.GetByUsername(ctx, m.username)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}

			_, err = m.sshKeyRepo.AddKey(ctx, player.ID, string(m.registration.sshKeyData))
			if err != nil {
				return registrationCompleteMsg{success: false, err: fmt.Errorf("account created but failed to add SSH key: %w", err)}
			}
		} else {
			// Create account with password
			_, err = m.playerRepo.CreateWithEmail(ctx, m.username, m.registration.password, m.registration.email)
			if err != nil {
				return registrationCompleteMsg{success: false, err: err}
			}
		}

		return registrationCompleteMsg{success: true}
	}
}

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
