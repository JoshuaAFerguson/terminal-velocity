// File: internal/tui/login.go
// Project: Terminal Velocity
// Description: Login screen with Escape Velocity-style ASCII logo
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	focusedField int // 0: username, 1: password, 2: buttons
	buttonIndex  int // 0: password login, 1: SSH key login, 2: register
	username     string
	password     string
	showPassword bool
	error        string
}

func newLoginModel() loginModel {
	return loginModel{
		focusedField: 0,
		buttonIndex:  0,
	}
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
	usernameFocused := m.mainMenu.cursor == 0 // Using mainMenu.cursor as login cursor for now
	usernameField := m.registration.email // Using email field for username
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
	passwordFocused := m.mainMenu.cursor == 1
	passwordField := strings.Repeat("*", len(m.registration.password))
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

	// Buttons
	btnPasswordLogin := "[ Login with Password ]"
	btnSSHLogin := "[ Login with SSH Key  ]"
	if m.mainMenu.cursor == 2 {
		btnPasswordLogin = HighlightStyle.Render(btnPasswordLogin)
	}
	if m.mainMenu.cursor == 3 {
		btnSSHLogin = HighlightStyle.Render(btnSSHLogin)
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center(btnPasswordLogin, panelWidth-2))
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-panelLeft-panelWidth-1))
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", panelLeft-1))
	sb.WriteString(BoxVertical)
	sb.WriteString(Center(btnSSHLogin, panelWidth-2))
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
	if m.mainMenu.cursor == 4 {
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

	// SSH connection info
	sshInfo := "Connect via SSH: ssh username@terminal-velocity.io:2222"
	sb.WriteString(BoxVertical)
	sb.WriteString(Center(sshInfo, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Empty space
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Error message if present
	if m.registration.error != "" {
		sb.WriteString(BoxVertical)
		sb.WriteString(Center(ErrorStyle.Render(m.registration.error), width-2))
		sb.WriteString(BoxVertical + "\n")
		sb.WriteString(BoxVertical)
		sb.WriteString(strings.Repeat(" ", width-2))
		sb.WriteString(BoxVertical + "\n")
	}

	// Footer
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(" [Tab] Next Field  [Enter] Submit  [R]egister  [Q]uit")
	sb.WriteString(strings.Repeat(" ", width-len(" [Tab] Next Field  [Enter] Submit  [R]egister  [Q]uit")-3))
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// Add a ScreenLogin constant to the Screen enum in model.go when integrating
