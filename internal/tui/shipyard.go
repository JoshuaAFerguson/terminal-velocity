package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type shipyardModel struct {
	cursor         int
	mode           string // "list", "details", "confirm_buy", "confirm_trade"
	selectedShip   *models.ShipType
	availableShips []models.ShipType
	currentPlanet  *models.Planet
	tradeInValue   int64
	loading        bool
	error          string
}

type shipyardLoadedMsg struct {
	ships  []models.ShipType
	planet *models.Planet
	err    error
}

type shipPurchasedMsg struct {
	success bool
	newShip *models.Ship
	err     error
}

func newShipyardModel() shipyardModel {
	return shipyardModel{
		cursor:  0,
		mode:    "list",
		loading: true,
	}
}

func (m Model) updateShipyard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.shipyard.mode == "list" {
				// Go back to main menu
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel current operation
			m.shipyard.mode = "list"
			m.shipyard.error = ""
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.shipyard.cursor > 0 {
				m.shipyard.cursor--
			}

		case "down", "j":
			maxCursor := len(m.shipyard.availableShips) - 1
			if m.shipyard.cursor < maxCursor {
				m.shipyard.cursor++
			}

		case "enter", " ":
			if m.shipyard.mode == "list" && m.shipyard.cursor < len(m.shipyard.availableShips) {
				// View ship details
				m.shipyard.selectedShip = &m.shipyard.availableShips[m.shipyard.cursor]
				m.shipyard.mode = "details"
				m.shipyard.error = ""
			} else if m.shipyard.mode == "details" {
				// Show purchase confirmation
				m.shipyard.mode = "confirm_buy"
			} else if m.shipyard.mode == "confirm_buy" {
				// Execute purchase
				return m, m.executePurchase(false)
			} else if m.shipyard.mode == "confirm_trade" {
				// Execute trade-in
				return m, m.executePurchase(true)
			}

		case "t": // Trade-in
			if m.shipyard.mode == "details" && m.currentShip != nil {
				// Show trade-in confirmation
				m.shipyard.mode = "confirm_trade"
				// Calculate trade-in value (70% of ship value)
				currentShipType := models.GetShipTypeByID(m.currentShip.TypeID)
				if currentShipType != nil {
					m.shipyard.tradeInValue = int64(float64(currentShipType.Price) * 0.70)
				}
			}
		}

	case shipyardLoadedMsg:
		m.shipyard.loading = false
		if msg.err != nil {
			m.shipyard.error = fmt.Sprintf("Failed to load shipyard: %v", msg.err)
		} else {
			m.shipyard.availableShips = msg.ships
			m.shipyard.currentPlanet = msg.planet
			m.shipyard.error = ""
		}

	case shipPurchasedMsg:
		if msg.success {
			// Purchase successful
			m.shipyard.mode = "list"
			m.shipyard.selectedShip = nil
			m.shipyard.error = "Ship purchased successfully!"
			// Reload player and ship data
			return m, m.loadPlayer()
		} else {
			m.shipyard.error = fmt.Sprintf("Purchase failed: %v", msg.err)
			m.shipyard.mode = "details"
		}
	}

	return m, nil
}

func (m Model) viewShipyard() string {
	// Header with player stats
	locationName := "Space"
	if m.shipyard.currentPlanet != nil {
		locationName = m.shipyard.currentPlanet.Name
	}
	s := renderHeader(m.username, m.player.Credits, locationName)
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Shipyard ===") + "\n\n"

	// Error display
	if m.shipyard.error != "" {
		s += helpStyle.Render(m.shipyard.error) + "\n\n"
	}

	// Loading state
	if m.shipyard.loading {
		s += "Loading shipyard data...\n"
		return s
	}

	// Mode-specific view
	switch m.shipyard.mode {
	case "list":
		s += m.viewShipList()
	case "details":
		s += m.viewShipDetails()
	case "confirm_buy":
		s += m.viewPurchaseConfirmation(false)
	case "confirm_trade":
		s += m.viewPurchaseConfirmation(true)
	default:
		s += "Unknown mode\n"
	}

	return s
}

func (m Model) viewShipList() string {
	s := ""

	// Current ship info
	if m.currentShip != nil {
		currentType := models.GetShipTypeByID(m.currentShip.TypeID)
		if currentType != nil {
			s += fmt.Sprintf("Current Ship: %s (%s)\n",
				statsStyle.Render(m.currentShip.Name),
				currentType.Name)
			s += fmt.Sprintf("Value: %s cr\n\n",
				statsStyle.Render(fmt.Sprintf("%d", currentType.Price)))
		}
	}

	// Check if any ships available
	if len(m.shipyard.availableShips) == 0 {
		s += helpStyle.Render("No ships available at this location") + "\n\n"
		s += renderFooter("ESC: Main Menu")
		return s
	}

	// Ship table header
	s += "Ship                      Class         Price        Cargo  Speed  Weapons\n"
	s += strings.Repeat("─", 78) + "\n"

	// List ships
	for i, ship := range m.shipyard.availableShips {
		// Check affordability
		affordable := ship.Price <= m.player.Credits

		line := fmt.Sprintf("%-25s %-13s %-12s %-6d %-6d %-7d",
			ship.Name,
			ship.Class,
			fmt.Sprintf("%d cr", ship.Price),
			ship.CargoSpace,
			ship.Speed,
			ship.WeaponSlots,
		)

		if !affordable {
			line = helpStyle.Render(line) // Dim unaffordable ships
		}

		if i == m.shipyard.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	// Help text
	s += renderFooter("↑/↓: Select  •  Enter: View Details  •  ESC: Main Menu")

	return s
}

func (m Model) viewShipDetails() string {
	if m.shipyard.selectedShip == nil {
		return "No ship selected\n"
	}

	ship := m.shipyard.selectedShip
	s := ""

	// Ship name and description
	s += fmt.Sprintf("%s\n", titleStyle.Render(ship.Name))
	s += fmt.Sprintf("%s\n\n", ship.Description)

	// Price and affordability
	affordable := ship.Price <= m.player.Credits
	s += fmt.Sprintf("Price: %s cr", statsStyle.Render(fmt.Sprintf("%d", ship.Price)))
	if !affordable {
		needed := ship.Price - m.player.Credits
		s += errorStyle.Render(fmt.Sprintf(" (need %d more)", needed))
	}
	s += "\n\n"

	// Combat stats
	s += "Combat Stats:\n"
	s += fmt.Sprintf("  Hull:         %s HP\n", statsStyle.Render(fmt.Sprintf("%d", ship.MaxHull)))
	s += fmt.Sprintf("  Shields:      %s HP (regen: %d/turn)\n",
		statsStyle.Render(fmt.Sprintf("%d", ship.MaxShields)), ship.ShieldRegen)
	s += fmt.Sprintf("  Weapon Slots: %s\n", statsStyle.Render(fmt.Sprintf("%d", ship.WeaponSlots)))
	s += "\n"

	// Performance stats
	s += "Performance:\n"
	s += fmt.Sprintf("  Speed:        %s\n", statsStyle.Render(fmt.Sprintf("%d", ship.Speed)))
	s += fmt.Sprintf("  Maneuver:     %s\n", statsStyle.Render(fmt.Sprintf("%d", ship.Maneuverability)))
	s += "\n"

	// Capacity stats
	s += "Capacity:\n"
	s += fmt.Sprintf("  Cargo Space:  %s units\n", statsStyle.Render(fmt.Sprintf("%d", ship.CargoSpace)))
	s += fmt.Sprintf("  Fuel Tank:    %s units\n", statsStyle.Render(fmt.Sprintf("%d", ship.MaxFuel)))
	s += fmt.Sprintf("  Crew:         %s\n", statsStyle.Render(fmt.Sprintf("%d", ship.MaxCrew)))
	s += fmt.Sprintf("  Outfit Space: %s\n", statsStyle.Render(fmt.Sprintf("%d", ship.OutfitSpace)))
	s += "\n"

	// Requirements
	if ship.MinCombatRating > 0 {
		s += fmt.Sprintf("Requires: Combat Rating %d\n\n", ship.MinCombatRating)
		if m.player.CombatRating < ship.MinCombatRating {
			s += errorStyle.Render("⚠ You don't meet the combat rating requirement!\n\n")
		}
	}

	// Trade-in option
	if m.currentShip != nil {
		currentType := models.GetShipTypeByID(m.currentShip.TypeID)
		if currentType != nil {
			tradeInValue := int64(float64(currentType.Price) * 0.70)
			s += fmt.Sprintf("Trade-in your %s: %s cr (70%% value)\n",
				currentType.Name,
				statsStyle.Render(fmt.Sprintf("%d", tradeInValue)))
			s += fmt.Sprintf("Net cost with trade-in: %s cr\n\n",
				statsStyle.Render(fmt.Sprintf("%d", ship.Price-tradeInValue)))
		}
	}

	// Help text
	helpText := "Enter: Purchase"
	if m.currentShip != nil {
		helpText += "  •  T: Trade-In"
	}
	helpText += "  •  ESC: Back"
	s += helpStyle.Render(helpText)

	return s
}

func (m Model) viewPurchaseConfirmation(isTradeIn bool) string {
	if m.shipyard.selectedShip == nil {
		return "No ship selected\n"
	}

	ship := m.shipyard.selectedShip
	s := ""

	if isTradeIn {
		s += errorStyle.Render("=== Trade-In Confirmation ===") + "\n\n"

		currentType := models.GetShipTypeByID(m.currentShip.TypeID)
		if currentType != nil {
			s += fmt.Sprintf("Trade in: %s (%s)\n", m.currentShip.Name, currentType.Name)
			s += fmt.Sprintf("Trade-in value: %s cr\n\n",
				statsStyle.Render(fmt.Sprintf("%d", m.shipyard.tradeInValue)))
		}

		s += fmt.Sprintf("New ship: %s\n", ship.Name)
		netCost := ship.Price - m.shipyard.tradeInValue
		s += fmt.Sprintf("Net cost: %s cr\n\n",
			statsStyle.Render(fmt.Sprintf("%d", netCost)))

		if netCost > m.player.Credits {
			s += errorStyle.Render("⚠ Insufficient credits!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else {
			s += errorStyle.Render("⚠ Warning: Your current ship and all cargo will be lost!\n\n")
			s += helpStyle.Render("Enter: Confirm Trade-In  •  ESC: Cancel")
		}
	} else {
		s += errorStyle.Render("=== Purchase Confirmation ===") + "\n\n"
		s += fmt.Sprintf("Ship: %s\n", ship.Name)
		s += fmt.Sprintf("Price: %s cr\n\n",
			statsStyle.Render(fmt.Sprintf("%d", ship.Price)))

		if ship.Price > m.player.Credits {
			s += errorStyle.Render("⚠ Insufficient credits!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else {
			if m.currentShip != nil {
				s += helpStyle.Render("Note: Your current ship will be kept. You can manage ships from the main menu.\n\n")
			}
			s += helpStyle.Render("Enter: Confirm Purchase  •  ESC: Cancel")
		}
	}

	return s
}

// loadShipyard loads available ships at the current location
func (m Model) loadShipyard() tea.Cmd {
	return func() tea.Msg {
		// TODO: Determine current planet and tech level
		// For now, show all ships

		// Filter ships by player's combat rating
		var availableShips []models.ShipType
		for _, ship := range models.StandardShipTypes {
			if ship.MinCombatRating <= m.player.CombatRating {
				availableShips = append(availableShips, ship)
			}
		}

		// Sort by price
		sort.Slice(availableShips, func(i, j int) bool {
			return availableShips[i].Price < availableShips[j].Price
		})

		return shipyardLoadedMsg{
			ships:  availableShips,
			planet: nil, // TODO: Get current planet
			err:    nil,
		}
	}
}

// executePurchase executes a ship purchase
func (m Model) executePurchase(isTradeIn bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if m.shipyard.selectedShip == nil {
			return shipPurchasedMsg{
				success: false,
				err:     fmt.Errorf("no ship selected"),
			}
		}

		// Check combat rating requirement
		if m.player.CombatRating < m.shipyard.selectedShip.MinCombatRating {
			return shipPurchasedMsg{
				success: false,
				err:     fmt.Errorf("insufficient combat rating (need %d)", m.shipyard.selectedShip.MinCombatRating),
			}
		}

		var finalCost int64
		if isTradeIn {
			// Trade-in current ship
			if m.currentShip == nil {
				return shipPurchasedMsg{
					success: false,
					err:     fmt.Errorf("no ship to trade in"),
				}
			}

			finalCost = m.shipyard.selectedShip.Price - m.shipyard.tradeInValue

			// Delete old ship (this will cascade delete cargo)
			err := m.shipRepo.Delete(ctx, m.currentShip.ID)
			if err != nil {
				return shipPurchasedMsg{
					success: false,
					err:     fmt.Errorf("failed to remove old ship: %w", err),
				}
			}
		} else {
			finalCost = m.shipyard.selectedShip.Price
		}

		// Check credits
		if finalCost > m.player.Credits {
			return shipPurchasedMsg{
				success: false,
				err:     fmt.Errorf("insufficient credits"),
			}
		}

		// Create new ship
		newShip := &models.Ship{
			ID:      uuid.New(),
			OwnerID: m.player.ID,
			TypeID:  m.shipyard.selectedShip.ID,
			Name:    m.shipyard.selectedShip.Name, // Default name
			Hull:    m.shipyard.selectedShip.MaxHull,
			Shields: m.shipyard.selectedShip.MaxShields,
			Fuel:    m.shipyard.selectedShip.MaxFuel,
			Crew:    m.shipyard.selectedShip.MaxCrew,
			Cargo:   []models.CargoItem{},
			Weapons: []string{},
			Outfits: []string{},
		}

		// Save new ship to database
		err := m.shipRepo.Create(ctx, newShip)
		if err != nil {
			return shipPurchasedMsg{
				success: false,
				err:     fmt.Errorf("failed to create ship: %w", err),
			}
		}

		// Update player credits
		newCredits := m.player.Credits - finalCost
		err = m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
		if err != nil {
			// Rollback: delete the ship we just created
			m.shipRepo.Delete(ctx, newShip.ID)
			return shipPurchasedMsg{
				success: false,
				err:     fmt.Errorf("failed to update credits: %w", err),
			}
		}

		// Update player's current ship
		err = m.playerRepo.UpdateShip(ctx, m.player.ID, newShip.ID)
		if err != nil {
			// Continue anyway - player has the ship, just not set as current
		}

		return shipPurchasedMsg{
			success: true,
			newShip: newShip,
			err:     nil,
		}
	}
}
