// File: internal/tui/landing.go
// Project: Terminal Velocity
// Description: Planetary landing screen with services menu
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
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
		{"Q", "Quest Terminal", ""},
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

// refuelShipCmd refuels the player's ship to full
func (m Model) refuelShipCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if player has a ship
		if m.currentShip == nil {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    0,
				err:     fmt.Errorf("no ship equipped"),
			}
		}

		// Get ship type to determine max fuel
		// For now, assume max fuel is 300 (we'll need to load ship type from DB later)
		// TODO: Load ship type from database to get actual MaxFuel value
		maxFuel := 300
		currentFuel := m.currentShip.Fuel

		// Calculate fuel needed
		fuelNeeded := maxFuel - currentFuel
		if fuelNeeded <= 0 {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    0,
				err:     fmt.Errorf("ship is already fully fueled"),
			}
		}

		// Calculate cost (10 credits per unit of fuel)
		costPerUnit := int64(10)
		totalCost := costPerUnit * int64(fuelNeeded)

		// Check if player has enough credits
		if m.player.Credits < totalCost {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("insufficient credits (need %d, have %d)", totalCost, m.player.Credits),
			}
		}

		// Update ship fuel in database
		err := m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, maxFuel)
		if err != nil {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("failed to refuel ship: %w", err),
			}
		}

		// Deduct credits from player
		m.player.Credits -= totalCost
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Try to rollback fuel update
			_ = m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, currentFuel)
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Update local ship state
		m.currentShip.Fuel = maxFuel

		return serviceCompleteMsg{
			service: "refuel",
			cost:    totalCost,
			err:     nil,
		}
	}
}

// repairShipCmd repairs the player's ship to full hull and shields
func (m Model) repairShipCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if player has a ship
		if m.currentShip == nil {
			return serviceCompleteMsg{
				service: "repair",
				cost:    0,
				err:     fmt.Errorf("no ship equipped"),
			}
		}

		// Get ship type to determine max hull/shields
		// For now, assume max values (we'll need to load ship type from DB later)
		// TODO: Load ship type from database to get actual MaxHull and MaxShields
		maxHull := 100
		maxShields := 100
		currentHull := m.currentShip.Hull
		currentShields := m.currentShip.Shields

		// Calculate damage
		hullDamage := maxHull - currentHull
		shieldDamage := maxShields - currentShields
		totalDamage := hullDamage + shieldDamage

		if totalDamage <= 0 {
			return serviceCompleteMsg{
				service: "repair",
				cost:    0,
				err:     fmt.Errorf("ship is already fully repaired"),
			}
		}

		// Calculate cost (50 credits per point of hull damage, 10 per shield)
		hullCostPerPoint := int64(50)
		shieldCostPerPoint := int64(10)
		totalCost := (hullCostPerPoint * int64(hullDamage)) + (shieldCostPerPoint * int64(shieldDamage))

		// Check if player has enough credits
		if m.player.Credits < totalCost {
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("insufficient credits (need %d, have %d)", totalCost, m.player.Credits),
			}
		}

		// Update ship hull and shields in database
		err := m.shipRepo.UpdateHullAndShields(ctx, m.currentShip.ID, maxHull, maxShields)
		if err != nil {
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("failed to repair ship: %w", err),
			}
		}

		// Deduct credits from player
		m.player.Credits -= totalCost
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Try to rollback repair
			_ = m.shipRepo.UpdateHullAndShields(ctx, m.currentShip.ID, currentHull, currentShields)
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Update local ship state
		m.currentShip.Hull = maxHull
		m.currentShip.Shields = maxShields

		return serviceCompleteMsg{
			service: "repair",
			cost:    totalCost,
			err:     nil,
		}
	}
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
			// Max 8 services
			if m.navigation.cursor < 7 {
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
			m.screen = ScreenMissionBoardEnhanced
			return m, nil

		case "q", "Q":
			// Quest Terminal
			m.screen = ScreenQuestBoardEnhanced
			return m, nil

		case "b", "B":
			// Bar & News
			m.screen = ScreenNews
			return m, nil

		case "r", "R":
			// Refuel
			return m, m.refuelShipCmd()

		case "h", "H":
			// Repairs
			return m, m.repairShipCmd()

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
				m.screen = ScreenMissionBoardEnhanced
			case 4: // Quest Terminal
				m.screen = ScreenQuestBoardEnhanced
			case 5: // Bar & News
				m.screen = ScreenNews
			case 6: // Refuel
				return m, m.refuelShipCmd()
			case 7: // Repairs
				return m, m.repairShipCmd()
			}
			return m, nil
		}

	case serviceCompleteMsg:
		// Handle refuel/repair completion
		if msg.err != nil {
			// Show error message
			m.errorMessage = fmt.Sprintf("%s failed: %v", msg.service, msg.err)
			m.showErrorDialog = true
		} else {
			// Show success message
			m.errorMessage = fmt.Sprintf("%s completed! Cost: %d credits",
				strings.Title(msg.service), msg.cost)
			m.showErrorDialog = true
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenLanding and ScreenTradingEnhanced constants to Screen enum when integrating
