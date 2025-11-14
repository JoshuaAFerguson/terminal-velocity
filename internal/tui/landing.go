// File: internal/tui/landing.go
// Project: Terminal Velocity
// Description: Planetary landing screen with services menu
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type landingModel struct {
	selectedService int
	planetName      string
	government      string
	techLevel       int
	population      string
}

func newLandingModel() landingModel {
	return landingModel{
		selectedService: 0,
		planetName:      "Earth Station",
		government:      "United Earth",
		techLevel:       9,
		population:      "8.2B",
	}
}

func (m Model) viewLanding() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Get planet info
	planetName := "Earth Station"
	government := "United Earth"
	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}

	// Header
	header := DrawHeader(planetName, government, credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Main content area with ASCII planet art
	planetArtWidth := 65
	planetArtLeft := (width - planetArtWidth) / 2

	// Planet art box
	var planetArt strings.Builder
	planetArt.WriteString("                                                               \n")
	planetArt.WriteString("            Welcome to Earth Station, Commander.               \n")
	planetArt.WriteString("                                                               \n")
	planetArt.WriteString("         [ASCII art of planet/station could go here]           \n")
	planetArt.WriteString("                       _______________                         \n")
	planetArt.WriteString("                      /               \\                        \n")
	planetArt.WriteString("                     /    " + IconPlanet + "  EARTH     \\                       \n")
	planetArt.WriteString("                    |  (Terran Alliance)|                      \n")
	planetArt.WriteString("                     \\     Pop: 8.2B    /                      \n")
	planetArt.WriteString("                      \\_____    _______/                       \n")
	planetArt.WriteString("                        /   \\__/   \\                           \n")
	planetArt.WriteString("                       /  Station   \\                          \n")
	planetArt.WriteString("                       \\____________/                          \n")
	planetArt.WriteString("                                                               \n")

	// Draw planet art (centered)
	artLines := strings.Split(planetArt.String(), "\n")
	for _, line := range artLines {
		if line == "" {
			continue
		}
		sb.WriteString(BoxVertical)
		sb.WriteString(strings.Repeat(" ", planetArtLeft-1))
		sb.WriteString(line)
		sb.WriteString(strings.Repeat(" ", width-planetArtLeft-len(line)-2))
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Services and Ship Status panels (side by side)
	servicesWidth := 30
	statusWidth := 39
	panelHeight := 12

	// Services panel content
	services := []struct {
		key   string
		label string
		price string
	}{
		{"C", "Commodity Exchange", ""},
		{"O", "Outfitters", ""},
		{"S", "Shipyard", ""},
		{"M", "Mission BBS", ""},
		{"B", "Bar & News", ""},
		{"R", "Refuel", "(1,200 cr)"},
		{"H", "Repairs", "(Free)"},
	}

	var servicesContent strings.Builder
	servicesContent.WriteString("  AVAILABLE SERVICES:       \n")
	servicesContent.WriteString("                            \n")
	for i, svc := range services {
		prefix := "  "
		if i == m.navigation.cursor {
			prefix = IconArrow + " "
		}
		line := fmt.Sprintf("%s[%s] %-18s %s", prefix, svc.key, svc.label, svc.price)
		servicesContent.WriteString(PadRight(line, servicesWidth-2) + "\n")
	}
	servicesContent.WriteString("                            \n")

	// Ship status panel content
	var statusContent strings.Builder
	statusContent.WriteString("  SHIP STATUS:                   \n")
	statusContent.WriteString("                                 \n")
	statusContent.WriteString("  Ship: Corvette \"Starhawk\"      \n")
	statusContent.WriteString("  Hull: 100%  Shields: 80%       \n")
	statusContent.WriteString("  Fuel: 67%   Cargo: 15/50t      \n")
	statusContent.WriteString("                                 \n")
	statusContent.WriteString("  Current System: Sol            \n")
	statusContent.WriteString("  Government: United Earth       \n")
	statusContent.WriteString("  Tech Level: 9                  \n")
	statusContent.WriteString("                                 \n")

	// Draw panels (simplified - actual implementation would render side-by-side)
	servicesPanel := DrawPanel("", servicesContent.String(), servicesWidth, panelHeight, false)
	statusPanel := DrawPanel("", statusContent.String(), statusWidth, panelHeight, false)

	// Draw both panels (this is simplified)
	servicesLines := strings.Split(servicesPanel, "\n")
	statusLines := strings.Split(statusPanel, "\n")

	for i := 0; i < len(servicesLines) && i < len(statusLines); i++ {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(servicesLines[i])
		sb.WriteString("  ")
		sb.WriteString(statusLines[i])
		sb.WriteString("    ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// News ticker
	newsWidth := width - 8
	newsPanel := DrawPanel("", " NEWS: Pirate activity reported in nearby systems...             ", newsWidth, 3, false)
	newsLines := strings.Split(newsPanel, "\n")
	for _, line := range newsLines {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(line)
		sb.WriteString("    ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[T]akeoff  [Tab] Next Service  [ESC] Exit", width)
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) updateLanding(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.navigation.cursor > 0 {
				m.navigation.cursor--
			}
			return m, nil

		case "down", "j":
			// Max 7 services
			if m.navigation.cursor < 6 {
				m.navigation.cursor++
			}
			return m, nil

		case "c", "C":
			// Commodity Exchange
			m.screen = ScreenTradingEnhanced
			return m, nil

		case "o", "O":
			// Outfitters
			m.screen = ScreenOutfitterEnhanced
			return m, nil

		case "s", "S":
			// Shipyard
			m.screen = ScreenShipyardEnhanced
			return m, nil

		case "m", "M":
			// Missions
			m.screen = ScreenMissions
			return m, nil

		case "b", "B":
			// Bar & News
			m.screen = ScreenNews
			return m, nil

		case "r", "R":
			// Refuel
			// TODO: Implement refuel logic
			return m, nil

		case "h", "H":
			// Repairs
			// TODO: Implement repair logic
			return m, nil

		case "t", "T":
			// Takeoff
			m.screen = ScreenSpaceView
			return m, nil

		case "esc":
			// Exit (takeoff)
			m.screen = ScreenSpaceView
			return m, nil

		case "enter":
			// Select current service
			switch m.navigation.cursor {
			case 0: // Commodity Exchange
				m.screen = ScreenTradingEnhanced
			case 1: // Outfitters
				m.screen = ScreenOutfitterEnhanced
			case 2: // Shipyard
				m.screen = ScreenShipyardEnhanced
			case 3: // Missions
				m.screen = ScreenMissions
			case 4: // Bar & News
				m.screen = ScreenNews
			case 5: // Refuel
				// TODO: Implement refuel
			case 6: // Repairs
				// TODO: Implement repairs
			}
			return m, nil
		}
	}

	return m, nil
}

// Add ScreenLanding and ScreenTradingEnhanced constants to Screen enum when integrating
