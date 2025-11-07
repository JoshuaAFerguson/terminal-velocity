package tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type registrationModel struct {
	step         int // 0: welcome, 1: email, 2: password, 3: confirm password, 4: creating
	email        string
	password     string
	confirmPass  string
	error        string
	cursor       int
	requireEmail bool
	sshKeyData   []byte // If registering with SSH key
}

type registrationCompleteMsg struct {
	success bool
	err     error
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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
			// Registration successful - need to reconnect
			m.registration.step = 5 // Success screen
			return m, tea.Quit
		} else {
			m.registration.error = fmt.Sprintf("Registration failed: %v", msg.err)
			return m, nil
		}
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
		if reg.requireEmail && reg.email == "" {
			reg.error = "Email is required"
			return m, nil
		}
		if reg.email != "" && !emailRegex.MatchString(reg.email) {
			reg.error = "Invalid email format"
			return m, nil
		}
		reg.error = ""

		// If using SSH key, skip password steps
		if len(reg.sshKeyData) > 0 {
			reg.step = 4 // Go to creating account
			return m, m.createAccount()
		}

		reg.step = 2

	case 2: // Password input
		if len(reg.password) < 8 {
			reg.error = "Password must be at least 8 characters"
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

	switch reg.step {
	case 1: // Email
		reg.email += input
		reg.error = ""
	case 2: // Password
		reg.password += input
		reg.error = ""
	case 3: // Confirm password
		reg.confirmPass += input
		reg.error = ""
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
		s += errorStyle.Render("⚠ " + reg.error) + "\n\n"
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
		s += "Create a password (minimum 8 characters):\n"
		s += "> " + strings.Repeat("•", len(reg.password)) + "█\n\n"
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
		s += "You can now log in with your credentials.\n\n"
		s += "Disconnecting..."
	}

	return s
}
