// File: internal/tui/shipyard_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced shipyard screen with ship browser and trade-in
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type shipyardEnhancedModel struct {
	selectedShip int
	mode         string // "browse", "confirm"
	ships        []shipListing
}

type shipListing struct {
	name         string
	class        string
	price        int64
	hull         int
	shields      int
	speed        int
	accel        int
	maneuver     int
	cargo        int
	fuel         int
	weaponSlots  int
	outfitSlots  int
	description  string
}

func newShipyardEnhancedModel() shipyardEnhancedModel {
	// Sample ships
	ships := []shipListing{
		{
			name: "Shuttle", class: "Utility", price: 12000,
			hull: 40, shields: 20, speed: 200, accel: 150, maneuver: 180,
			cargo: 50, fuel: 400, weaponSlots: 1, outfitSlots: 2,
			description: "Reliable civilian transport",
		},
		{
			name: "Lightning", class: "Light Fighter", price: 45000,
			hull: 80, shields: 60, speed: 450, accel: 380, maneuver: 420,
			cargo: 15, fuel: 300, weaponSlots: 2, outfitSlots: 3,
			description: "Fast and maneuverable fighter",
		},
		{
			name: "Courier", class: "Fast Courier", price: 75000,
			hull: 100, shields: 80, speed: 420, accel: 350, maneuver: 380,
			cargo: 40, fuel: 500, weaponSlots: 2, outfitSlots: 4,
			description: "Swift cargo transport",
		},
		{
			name: "Corvette", class: "Combat Corvette", price: 180000,
			hull: 200, shields: 150, speed: 320, accel: 280, maneuver: 300,
			cargo: 50, fuel: 600, weaponSlots: 4, outfitSlots: 6,
			description: "Balanced combat and cargo",
		},
		{
			name: "Destroyer", class: "Heavy Destroyer", price: 450000,
			hull: 400, shields: 300, speed: 250, accel: 200, maneuver: 220,
			cargo: 80, fuel: 800, weaponSlots: 6, outfitSlots: 8,
			description: "Powerful warship",
		},
		{
			name: "Freighter", class: "Bulk Freighter", price: 220000,
			hull: 300, shields: 100, speed: 180, accel: 120, maneuver: 140,
			cargo: 200, fuel: 1000, weaponSlots: 2, outfitSlots: 4,
			description: "Maximum cargo capacity",
		},
		{
			name: "Cruiser", class: "Heavy Cruiser", price: 780000,
			hull: 600, shields: 500, speed: 280, accel: 220, maneuver: 240,
			cargo: 100, fuel: 1200, weaponSlots: 8, outfitSlots: 12,
			description: "Elite warship",
		},
		{
			name: "Battleship", class: "Capital Battleship", price: 1500000,
			hull: 1000, shields: 800, speed: 200, accel: 150, maneuver: 180,
			cargo: 150, fuel: 1500, weaponSlots: 12, outfitSlots: 16,
			description: "Ultimate firepower",
		},
	}

	return shipyardEnhancedModel{
		selectedShip: 0,
		mode:         "browse",
		ships:        ships,
	}
}

func (m Model) viewShipyardEnhanced() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Header
	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	header := DrawHeader("SHIPYARD - Earth Station", "", credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Initialize if needed
	if len(m.shipyardEnhanced.ships) == 0 {
		m.shipyardEnhanced = newShipyardEnhancedModel()
	}

	// Two-column layout: Ships list (left) + Ship details (right)
	leftWidth := 30
	rightWidth := width - leftWidth - 6

	// Ship list
	var shipsContent strings.Builder
	shipsContent.WriteString(" AVAILABLE SHIPS:           \n")
	shipsContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	shipsContent.WriteString("                            \n")

	for i, ship := range m.shipyardEnhanced.ships {
		prefix := "   "
		if i == m.shipyardEnhanced.selectedShip {
			prefix = " " + IconArrow + " "
		}
		line := fmt.Sprintf("%s%-12s %9s cr", prefix, ship.name, formatNumber(ship.price))
		shipsContent.WriteString(PadRight(line, leftWidth-2) + "\n")
	}

	// Pad to height
	for i := len(m.shipyardEnhanced.ships); i < 10; i++ {
		shipsContent.WriteString(strings.Repeat(" ", leftWidth-2) + "\n")
	}

	shipsPanel := DrawPanel("", shipsContent.String(), leftWidth, 16, false)

	// Ship details
	var detailsContent strings.Builder
	if m.shipyardEnhanced.selectedShip < len(m.shipyardEnhanced.ships) {
		ship := m.shipyardEnhanced.ships[m.shipyardEnhanced.selectedShip]

		detailsContent.WriteString(fmt.Sprintf(" SHIP: %s                      \n", strings.ToUpper(ship.name)))
		detailsContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		detailsContent.WriteString("                                      \n")

		// ASCII ship art (simplified)
		detailsContent.WriteString("         ___                          \n")
		detailsContent.WriteString("        /   \\___                      \n")
		detailsContent.WriteString("       |  " + IconShip + "  ___>                     \n")
		detailsContent.WriteString("        \\___/                         \n")
		detailsContent.WriteString("                                      \n")

		detailsContent.WriteString(fmt.Sprintf(" Class: %-29s\n", ship.class))
		detailsContent.WriteString(fmt.Sprintf(" Price: %-29s\n", FormatCredits(ship.price)))
		detailsContent.WriteString("                                      \n")

		// Stats with progress bars
		detailsContent.WriteString(fmt.Sprintf(" Hull: %s %d\n",
			DrawProgressBar(ship.hull, 1000, 6), ship.hull))
		detailsContent.WriteString(fmt.Sprintf(" Shields: %s %d\n",
			DrawProgressBar(ship.shields, 800, 6), ship.shields))
		detailsContent.WriteString(fmt.Sprintf(" Speed: %s %d\n",
			DrawProgressBar(ship.speed, 500, 6), ship.speed))
		detailsContent.WriteString(fmt.Sprintf(" Accel: %s %d\n",
			DrawProgressBar(ship.accel, 400, 6), ship.accel))
		detailsContent.WriteString(fmt.Sprintf(" Maneuver: %s %d\n",
			DrawProgressBar(ship.maneuver, 500, 6), ship.maneuver))
		detailsContent.WriteString("                                      \n")

		detailsContent.WriteString(fmt.Sprintf(" Cargo: %d tons                      \n", ship.cargo))
		detailsContent.WriteString(fmt.Sprintf(" Fuel: %d units                     \n", ship.fuel))
		detailsContent.WriteString(fmt.Sprintf(" Weapon Slots: %d                    \n", ship.weaponSlots))
		detailsContent.WriteString(fmt.Sprintf(" Outfit Slots: %d                    \n", ship.outfitSlots))
	}

	detailsPanel := DrawPanel("", detailsContent.String(), rightWidth, 16, false)

	// Combine panels (simplified - actual side-by-side rendering would be better)
	sb.WriteString(BoxVertical + "  ")
	sb.WriteString(shipsPanel)
	sb.WriteString("  ")
	sb.WriteString(BoxVertical + "\n")
	sb.WriteString(BoxVertical + "  ")
	sb.WriteString(detailsPanel)
	sb.WriteString("  ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Purchase/Trade-in panel
	purchaseWidth := width - 4
	var purchaseContent strings.Builder

	if m.shipyardEnhanced.selectedShip < len(m.shipyardEnhanced.ships) {
		ship := m.shipyardEnhanced.ships[m.shipyardEnhanced.selectedShip]
		currentShip := "Corvette \"Starhawk\""
		tradeInValue := int64(126000) // TODO: Calculate from current ship

		purchaseContent.WriteString(fmt.Sprintf(" YOUR SHIP: %-35s Trade-in: %s\n",
			currentShip, FormatCredits(tradeInValue)))
		purchaseContent.WriteString("                                                                      \n")
		purchaseContent.WriteString(fmt.Sprintf(" Purchase %s for %s?\n", ship.name, FormatCredits(ship.price)))

		diff := tradeInValue - ship.price
		if diff > 0 {
			purchaseContent.WriteString(fmt.Sprintf(" With trade-in credit: You will GAIN %s\n", FormatCredits(diff)))
		} else {
			purchaseContent.WriteString(fmt.Sprintf(" With trade-in credit: You will PAY %s\n", FormatCredits(-diff)))
		}
		purchaseContent.WriteString("                                                                      \n")
		purchaseContent.WriteString(" [ Purchase ] [ Trade-In Purchase ] [ Cancel ]                       \n")
	}

	purchase := DrawPanel("", purchaseContent.String(), purchaseWidth, 8, false)
	purchaseLines := strings.Split(purchase, "\n")
	for _, line := range purchaseLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[↑↓] Select Ship  [Enter] Details  [P]urchase  [T]rade-In  [ESC] Back", width)
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) updateShipyardEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.shipyardEnhanced.selectedShip > 0 {
				m.shipyardEnhanced.selectedShip--
			}
			return m, nil

		case "down", "j":
			if m.shipyardEnhanced.selectedShip < len(m.shipyardEnhanced.ships)-1 {
				m.shipyardEnhanced.selectedShip++
			}
			return m, nil

		case "p", "P":
			// Purchase ship
			// TODO: Implement purchase logic via API
			return m, nil

		case "t", "T":
			// Trade-in purchase
			// TODO: Implement trade-in logic via API
			return m, nil

		case "esc":
			// Back to landing
			m.screen = ScreenLanding
			return m, nil
		}
	}

	return m, nil
}

// Add ScreenShipyardEnhanced constant to Screen enum when integrating
