// File: internal/tui/shipyard_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced shipyard screen with ship browser and trade-in
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
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

// purchaseShipCmd purchases a new ship without trading in the old one
func (m Model) purchaseShipCmd(shipName string, shipPrice int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate credits
		if m.player.Credits < shipPrice {
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("insufficient credits (need %d, have %d)",
					shipPrice, m.player.Credits),
			}
		}

		// Create new ship instance
		shipTypeID := strings.ToLower(shipName) // Map ship name to type ID
		newShip := &models.Ship{
			ID:      uuid.New(),
			OwnerID: m.playerID,
			TypeID:  shipTypeID,
			Name:    shipName, // Default name, player can rename later
			Hull:    getShipMaxHull(shipName),
			Shields: getShipMaxShields(shipName),
			Fuel:    getShipMaxFuel(shipName),
			Cargo:   []models.CargoItem{},
			Crew:    1,
			Weapons: []string{},
			Outfits: []string{},
		}

		// Create ship in database
		err := m.shipRepo.Create(ctx, newShip)
		if err != nil {
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("failed to create ship: %w", err),
			}
		}

		// Deduct credits
		m.player.Credits -= shipPrice
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Rollback ship creation
			_ = m.shipRepo.Delete(ctx, newShip.ID)
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Set as current ship
		m.currentShip = newShip

		return shipPurchaseCompleteMsg{ship: newShip}
	}
}

// tradeInPurchaseShipCmd purchases a new ship and trades in the current ship
func (m Model) tradeInPurchaseShipCmd(shipName string, shipPrice int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if m.currentShip == nil {
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("no ship to trade in"),
			}
		}

		// Calculate trade-in value (70% of original price)
		oldShipPrice := getShipPrice(m.currentShip.TypeID)
		tradeInValue := int64(float64(oldShipPrice) * 0.7)

		// Calculate net cost
		netCost := shipPrice - tradeInValue

		// Validate credits for net cost
		if netCost > 0 && m.player.Credits < netCost {
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("insufficient credits (need %d after trade-in, have %d)",
					netCost, m.player.Credits),
			}
		}

		// Save old ship ID for deletion
		oldShipID := m.currentShip.ID

		// Create new ship instance
		shipTypeID := strings.ToLower(shipName)
		newShip := &models.Ship{
			ID:      uuid.New(),
			OwnerID: m.playerID,
			TypeID:  shipTypeID,
			Name:    shipName,
			Hull:    getShipMaxHull(shipName),
			Shields: getShipMaxShields(shipName),
			Fuel:    getShipMaxFuel(shipName),
			Cargo:   []models.CargoItem{},
			Crew:    1,
			Weapons: []string{},
			Outfits: []string{},
		}

		// Create new ship in database
		err := m.shipRepo.Create(ctx, newShip)
		if err != nil {
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("failed to create ship: %w", err),
			}
		}

		// Update credits (add trade-in value, subtract new ship price)
		if netCost > 0 {
			m.player.Credits -= netCost
		} else {
			m.player.Credits += (-netCost) // netCost is negative, so we add the absolute value
		}

		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Rollback ship creation
			_ = m.shipRepo.Delete(ctx, newShip.ID)
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("failed to update credits: %w", err),
			}
		}

		// Delete old ship
		err = m.shipRepo.Delete(ctx, oldShipID)
		if err != nil {
			// Rollback credit update
			if netCost > 0 {
				m.player.Credits += netCost
			} else {
				m.player.Credits -= (-netCost)
			}
			_ = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
			_ = m.shipRepo.Delete(ctx, newShip.ID)
			return shipPurchaseCompleteMsg{
				err: fmt.Errorf("failed to delete old ship: %w", err),
			}
		}

		// Set as current ship
		m.currentShip = newShip

		return shipPurchaseCompleteMsg{ship: newShip}
	}
}

// Helper functions to get ship specifications
func getShipMaxHull(shipName string) int {
	shipSpecs := map[string]int{
		"Shuttle": 40, "Lightning": 80, "Courier": 100,
		"Corvette": 200, "Destroyer": 400, "Freighter": 300,
		"Cruiser": 600, "Battleship": 1000,
	}
	if hull, ok := shipSpecs[shipName]; ok {
		return hull
	}
	return 100 // Default
}

func getShipMaxShields(shipName string) int {
	shipSpecs := map[string]int{
		"Shuttle": 20, "Lightning": 60, "Courier": 80,
		"Corvette": 150, "Destroyer": 300, "Freighter": 100,
		"Cruiser": 500, "Battleship": 800,
	}
	if shields, ok := shipSpecs[shipName]; ok {
		return shields
	}
	return 50 // Default
}

func getShipMaxFuel(shipName string) int {
	shipSpecs := map[string]int{
		"Shuttle": 400, "Lightning": 300, "Courier": 500,
		"Corvette": 600, "Destroyer": 800, "Freighter": 1000,
		"Cruiser": 1200, "Battleship": 1500,
	}
	if fuel, ok := shipSpecs[shipName]; ok {
		return fuel
	}
	return 300 // Default
}

func getShipPrice(shipTypeID string) int64 {
	// Convert ship type ID back to name and get price
	shipName := strings.Title(shipTypeID)
	shipPrices := map[string]int64{
		"Shuttle": 12000, "Lightning": 45000, "Courier": 75000,
		"Corvette": 180000, "Destroyer": 450000, "Freighter": 220000,
		"Cruiser": 780000, "Battleship": 1500000,
	}
	if price, ok := shipPrices[shipName]; ok {
		return price
	}
	return 10000 // Default
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
			// Purchase ship without trade-in
			if m.shipyardEnhanced.selectedShip < len(m.shipyardEnhanced.ships) {
				ship := m.shipyardEnhanced.ships[m.shipyardEnhanced.selectedShip]
				return m, m.purchaseShipCmd(ship.name, ship.price)
			}
			return m, nil

		case "t", "T":
			// Trade-in purchase
			if m.shipyardEnhanced.selectedShip < len(m.shipyardEnhanced.ships) {
				ship := m.shipyardEnhanced.ships[m.shipyardEnhanced.selectedShip]
				return m, m.tradeInPurchaseShipCmd(ship.name, ship.price)
			}
			return m, nil

		case "esc":
			// Back to landing
			m.screen = ScreenLanding
			return m, nil
		}

	case shipPurchaseCompleteMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
		} else {
			// Success - ship purchased
			m.currentShip = msg.ship
			// Optionally show success message or return to landing
			m.screen = ScreenLanding
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenShipyardEnhanced constant to Screen enum when integrating
