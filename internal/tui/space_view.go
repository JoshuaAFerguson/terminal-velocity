// File: internal/tui/space_view.go
// Project: Terminal Velocity
// Description: Main space view with 2D viewport, HUD, radar, and status
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type spaceViewModel struct {
	// Space objects visible in current system
	planets  []*models.Planet
	ships    []spaceObject
	player   playerPosition

	// Target selection
	targetIndex int
	hasTarget   bool

	// Chat state
	chatExpanded bool
	chatInput    string
	chatChannel  int // 0: Global, 1: System, 2: Faction, 3: DM
}

type spaceObject struct {
	name     string
	icon     string
	x, y     float64
	distance float64
	hostile  bool
	objType  string // "planet", "ship", "enemy", "player"
}

type playerPosition struct {
	x, y float64
}

func newSpaceViewModel() spaceViewModel {
	return spaceViewModel{
		player:       playerPosition{x: 0, y: 0},
		chatExpanded: false,
		chatChannel:  0,
	}
}

func (m Model) viewSpaceView() string {
	width := 80
	height := 24
	if m.width > 80 {
		width = m.width
	}
	if m.height > 24 {
		height = m.height
	}

	var sb strings.Builder

	// Calculate shield percentage
	// TODO: Get max values from ShipType when API integration is complete
	maxShields := 100
	shieldPercent := 80
	if m.currentShip != nil {
		if maxShields > 0 {
			shieldPercent = (m.currentShip.Shields * 100) / maxShields
		}
	}

	// Header
	systemName := "Unknown System"
	if m.player != nil {
		// Would load system name from database
		systemName = "Sol System"
	}
	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}

	header := DrawHeader("TERMINAL VELOCITY v1.0", systemName, credits, shieldPercent, width)
	sb.WriteString(header + "\n")

	// Main content area
	contentHeight := height - 6 // Header + footer + chat
	if m.spaceView.chatExpanded {
		contentHeight -= 8 // More space for expanded chat
	}

	// Left side: Space viewport + target/cargo panels
	viewportWidth := width - 17 // Leave room for right sidebar
	viewportHeight := contentHeight - 8 // Leave room for bottom panels

	// Draw space viewport
	sb.WriteString(m.drawSpaceViewport(viewportWidth, viewportHeight))

	// Right sidebar: Radar + Status
	// TODO: Implement proper side-by-side rendering
	// rightSidebar := m.drawRightSidebar(15, viewportHeight)

	// For now, the sidebar is rendered inline below the viewport
	sb.WriteString("\n")

	// Bottom panels: Target info + Cargo
	sb.WriteString(m.drawBottomPanels(viewportWidth, 6))

	// Chat window
	if m.spaceView.chatExpanded {
		sb.WriteString(m.drawChatExpanded(width))
	} else {
		sb.WriteString(m.drawChatCollapsed(width))
	}

	// Footer
	footer := DrawFooter("[L]and  [J]ump  [T]arget  [F]ire  [H]ail  [M]ap  [C]hat  [I]nfo  [ESC] Menu", width)
	sb.WriteString("\n" + footer)

	return sb.String()
}

func (m Model) drawSpaceViewport(width, height int) string {
	var sb strings.Builder

	// Top border
	sb.WriteString(BoxVertical + "    ")
	sb.WriteString(BoxTopLeftDouble)
	sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-8))
	sb.WriteString(BoxTopRightDouble + "\n")

	// Space content
	for i := 0; i < height; i++ {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(BoxVerticalDouble)

		// Draw space objects based on y position
		line := ""
		switch i {
		case 2:
			// Stars scattered
			line = "                          " + IconStar + "                                    "
		case 4:
			// Planet (Earth)
			line = "             " + IconStar + "                    " + IconPlanet + " Earth                      "
		case height / 2:
			// Player ship in center
			line = Center(IconShip, width-8)
			line += "\n" + BoxVertical + "    " + BoxVerticalDouble
			line += Center("You", width-8)
		case height/2 + 3:
			// Enemy ship
			line = "                                             " + IconEnemy + " Pirate          "
		case height - 3:
			// Another planet (Mars)
			line = "           " + IconPlanet + " Mars                                              "
		case 1, 6, height - 2:
			// Stars
			line = "        " + IconStar + "                                                      " + IconStar + "       "
		default:
			line = strings.Repeat(" ", width-8)
		}

		if len(line) < width-8 {
			line = PadRight(line, width-8)
		}
		sb.WriteString(line[:width-8])
		sb.WriteString(BoxVerticalDouble + "\n")
	}

	// Bottom border
	sb.WriteString(BoxVertical + "    ")
	sb.WriteString(BoxBottomLeftDouble)
	sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-8))
	sb.WriteString(BoxBottomRightDouble)

	return sb.String()
}

func (m Model) drawRightSidebar(width, height int) string {
	var sb strings.Builder

	// Radar panel
	radarHeight := 13
	var radarContent strings.Builder
	radarContent.WriteString("   RADAR     \n")
	radarContent.WriteString("             \n")
	radarContent.WriteString("      " + IconStar + "      \n")
	radarContent.WriteString("             \n")
	radarContent.WriteString("   " + IconPlanet + "    " + IconEnemy + "    \n")
	radarContent.WriteString("        " + IconPlayer + "    \n")
	radarContent.WriteString("      " + IconStar + "      \n")
	radarContent.WriteString("             \n")

	radar := DrawPanel("", radarContent.String(), width, radarHeight, false)
	sb.WriteString(radar + "\n")

	// Status panel
	var statusContent strings.Builder
	// TODO: Get max values from ShipType when API integration is complete
	maxHull := 100
	maxFuel := 100
	hullPercent := 100
	fuelPercent := 67
	if m.currentShip != nil {
		if maxHull > 0 {
			hullPercent = (m.currentShip.Hull * 100) / maxHull
		}
		if maxFuel > 0 {
			fuelPercent = (m.currentShip.Fuel * 100) / maxFuel
		}
	}

	statusContent.WriteString("   STATUS    \n")
	statusContent.WriteString("━━━━━━━━━━━━━\n")
	statusContent.WriteString(fmt.Sprintf(" Hull: %s\n", DrawProgressBar(hullPercent, 100, 6)))
	statusContent.WriteString(fmt.Sprintf("       %d%%  \n", hullPercent))
	statusContent.WriteString(fmt.Sprintf(" Fuel: %s\n", DrawProgressBar(fuelPercent, 100, 6)))
	statusContent.WriteString(fmt.Sprintf("       %d%%   \n", fuelPercent))
	statusContent.WriteString(" Speed: 340  \n")

	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	statusContent.WriteString(" Credits:    \n")
	statusContent.WriteString(fmt.Sprintf("  %s\n", FormatCredits(credits)))

	status := DrawPanel("", statusContent.String(), width, height-radarHeight-1, false)
	sb.WriteString(status)

	return sb.String()
}

func (m Model) drawBottomPanels(width, height int) string {
	var sb strings.Builder

	// Target panel (left)
	targetWidth := 25
	var targetContent strings.Builder
	targetContent.WriteString(" TARGET: Pirate Viper    \n")
	targetContent.WriteString(" Distance: 2,340 km      \n")
	targetContent.WriteString(" Shields: 45%            \n")
	targetContent.WriteString(" Attitude: Hostile       \n")

	// Cargo panel (right)
	cargoWidth := 38
	var cargoContent strings.Builder
	cargoContent.WriteString(" CARGO: 15/50 tons                \n")
	cargoContent.WriteString(" " + IconBullet + " Food (10t)  " + IconBullet + " Electronics (5t) \n")

	sb.WriteString(BoxVertical + "  ")

	// Draw both panels inline (simplified)
	target := DrawPanel("", targetContent.String(), targetWidth, height, false)
	cargo := DrawPanel("", cargoContent.String(), cargoWidth, height, false)

	// This is simplified - actual implementation would render side-by-side
	sb.WriteString(target)
	sb.WriteString("  ")
	sb.WriteString(cargo)

	return sb.String()
}

func (m Model) drawChatCollapsed(width int) string {
	var sb strings.Builder

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxTopRight + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" CHAT [Global] " + IconArrow + "                                               [C] to expand ")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" SpaceCadet: Anyone near Sol system?                                  3m ago ")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxBottomRight + " ")
	sb.WriteString(BoxVertical)

	return sb.String()
}

func (m Model) drawChatExpanded(width int) string {
	var sb strings.Builder

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxTopRight + " ")
	sb.WriteString(BoxVertical + "\n")

	// Chat header with channels
	channels := []string{"Global", "System", "Faction", "DM"}
	channelText := " CHAT: "
	for i, ch := range channels {
		if i == m.spaceView.chatChannel {
			channelText += "[" + ch + " " + IconArrow + "] "
		} else {
			channelText += "[" + ch + "] "
		}
	}
	channelText = PadRight(channelText, width-29) + "[C] to collapse "

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(channelText)
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Separator
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxCross + " ")
	sb.WriteString(BoxVertical + "\n")

	// Chat messages
	messages := []string{
		" [SpaceCadet] Anyone near Sol system?                     3m ago ",
		" [TraderJoe] Yeah I'm docked at Earth. Need anything?     2m ago ",
		" [SpaceCadet] Looking for escort to Alpha Centauri        2m ago ",
		" [PirateKing] I'll escort you... to your doom! Arr!       1m ago ",
		" [TraderJoe] Ignore him. I can escort for 5k credits      1m ago ",
		" [YOU] I'm at Earth too, what's the pirate situation?     now    ",
	}

	for _, msg := range messages {
		sb.WriteString(BoxVertical + " ")
		sb.WriteString(BoxVertical)
		sb.WriteString(PadRight(msg, width-4))
		sb.WriteString(BoxVertical + " ")
		sb.WriteString(BoxVertical + "\n")
	}

	// Empty line
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-4))
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Message input
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" Message: [" + PadRight(m.spaceView.chatInput+"_", width-16) + "]")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxBottomRight + " ")
	sb.WriteString(BoxVertical)

	return sb.String()
}

func (m Model) updateSpaceView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c", "C":
			// Toggle chat
			m.spaceView.chatExpanded = !m.spaceView.chatExpanded
			return m, nil

		case "l", "L":
			// Land on planet (if near one)
			m.screen = ScreenLanding
			return m, nil

		case "j", "J":
			// Jump (navigation)
			m.screen = ScreenNavigation
			return m, nil

		case "t", "T":
			// Target next object
			// TODO: Cycle through targetable objects
			return m, nil

		case "m", "M":
			// System map
			m.screen = ScreenNavigation
			return m, nil

		case "i", "I":
			// Player info
			// TODO: Implement ScreenPlayerInfo
			// m.screen = ScreenPlayerInfo
			return m, nil

		case "esc":
			// Menu
			m.screen = ScreenMainMenu
			return m, nil

		default:
			// Handle chat input if expanded
			if m.spaceView.chatExpanded {
				if msg.String() == "enter" {
					// Send chat message
					// TODO: Send to chat manager
					m.spaceView.chatInput = ""
					return m, nil
				} else if msg.String() == "backspace" {
					if len(m.spaceView.chatInput) > 0 {
						m.spaceView.chatInput = m.spaceView.chatInput[:len(m.spaceView.chatInput)-1]
					}
					return m, nil
				} else if len(msg.String()) == 1 {
					// Add character to chat input
					m.spaceView.chatInput += msg.String()
					return m, nil
				}
			}
		}
	}

	return m, nil
}

// Add ScreenSpaceView and ScreenLanding constants to Screen enum when integrating
