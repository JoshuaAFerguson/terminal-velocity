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
	mode           string // "list", "details", "confirm_buy", "confirm_trade", "compare", "compare_select"
	selectedShip   *models.ShipType
	compareShip1   *models.ShipType
	compareShip2   *models.ShipType
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
			} else if m.shipyard.mode == "compare_select" && m.shipyard.cursor < len(m.shipyard.availableShips) {
				// Select ship for comparison
				selectedShip := &m.shipyard.availableShips[m.shipyard.cursor]
				if m.shipyard.compareShip1 == nil {
					m.shipyard.compareShip1 = selectedShip
					m.shipyard.error = "Select second ship to compare..."
				} else if m.shipyard.compareShip2 == nil {
					m.shipyard.compareShip2 = selectedShip
					m.shipyard.mode = "compare"
					m.shipyard.error = ""
				}
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

		case "c": // Compare ships
			if m.shipyard.mode == "list" {
				m.shipyard.mode = "compare_select"
				m.shipyard.compareShip1 = nil
				m.shipyard.compareShip2 = nil
				m.shipyard.cursor = 0
				m.shipyard.error = "Select first ship to compare..."
			} else if m.shipyard.mode == "compare" {
				// Start new comparison
				m.shipyard.compareShip1 = nil
				m.shipyard.compareShip2 = nil
				m.shipyard.mode = "compare_select"
				m.shipyard.cursor = 0
				m.shipyard.error = "Select first ship to compare..."
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
	case "compare_select":
		s += m.viewShipListCompare()
	case "compare":
		s += m.viewShipComparison()
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
	s += renderFooter("↑/↓: Select  •  Enter: View Details  •  C: Compare Ships  •  ESC: Main Menu")

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

func (m Model) viewShipListCompare() string {
	s := ""

	// Show selected ships
	if m.shipyard.compareShip1 != nil {
		s += fmt.Sprintf("First ship: %s\n", statsStyle.Render(m.shipyard.compareShip1.Name))
	}
	if m.shipyard.compareShip2 != nil {
		s += fmt.Sprintf("Second ship: %s\n", statsStyle.Render(m.shipyard.compareShip2.Name))
	}
	s += "\n"

	// Check if any ships available
	if len(m.shipyard.availableShips) == 0 {
		s += helpStyle.Render("No ships available at this location") + "\n\n"
		s += renderFooter("ESC: Cancel")
		return s
	}

	// Ship table header
	s += "Ship                      Class         Price        Cargo  Speed  Weapons\n"
	s += strings.Repeat("─", 78) + "\n"

	// List ships
	for i, ship := range m.shipyard.availableShips {
		line := fmt.Sprintf("%-25s %-13s %-12s %-6d %-6d %-7d",
			ship.Name,
			ship.Class,
			fmt.Sprintf("%d cr", ship.Price),
			ship.CargoSpace,
			ship.Speed,
			ship.WeaponSlots,
		)

		if i == m.shipyard.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	// Help text
	s += renderFooter("↑/↓: Select  •  Enter: Choose Ship  •  ESC: Cancel")

	return s
}

func (m Model) viewShipComparison() string {
	if m.shipyard.compareShip1 == nil || m.shipyard.compareShip2 == nil {
		return "Error: Both ships must be selected\n"
	}

	ship1 := m.shipyard.compareShip1
	ship2 := m.shipyard.compareShip2
	s := ""

	// Title
	s += titleStyle.Render("=== Ship Comparison ===") + "\n\n"

	// Ship names side-by-side
	s += fmt.Sprintf("%-40s %s\n",
		statsStyle.Render(ship1.Name),
		statsStyle.Render(ship2.Name))
	s += fmt.Sprintf("%-40s %s\n",
		ship1.Class,
		ship2.Class)
	s += "\n"

	// Price comparison
	s += m.renderComparisonLine("Price:", ship1.Price, ship2.Price, "cr", false)

	// Cost-benefit analysis
	if m.currentShip != nil {
		currentType := models.GetShipTypeByID(m.currentShip.TypeID)
		if currentType != nil {
			tradeIn := int64(float64(currentType.Price) * 0.70)
			netCost1 := ship1.Price - tradeIn
			netCost2 := ship2.Price - tradeIn
			s += fmt.Sprintf("\nWith trade-in (%s):\n", currentType.Name)
			s += m.renderComparisonLine("Net cost:", netCost1, netCost2, "cr", false)
		}
	}
	s += "\n"

	// Combat stats
	s += subtitleStyle.Render("Combat Stats:") + "\n"
	s += m.renderComparisonLine("Hull:", int64(ship1.MaxHull), int64(ship2.MaxHull), "HP", true)
	s += m.renderComparisonLine("Shields:", int64(ship1.MaxShields), int64(ship2.MaxShields), "HP", true)
	s += m.renderComparisonLine("Shield Regen:", int64(ship1.ShieldRegen), int64(ship2.ShieldRegen), "/turn", true)
	s += m.renderComparisonLine("Weapon Slots:", int64(ship1.WeaponSlots), int64(ship2.WeaponSlots), "", true)
	s += "\n"

	// Performance stats
	s += subtitleStyle.Render("Performance:") + "\n"
	s += m.renderComparisonLine("Speed:", int64(ship1.Speed), int64(ship2.Speed), "", true)
	s += m.renderComparisonLine("Maneuverability:", int64(ship1.Maneuverability), int64(ship2.Maneuverability), "", true)
	s += "\n"

	// Capacity stats
	s += subtitleStyle.Render("Capacity:") + "\n"
	s += m.renderComparisonLine("Cargo Space:", int64(ship1.CargoSpace), int64(ship2.CargoSpace), "tons", true)
	s += m.renderComparisonLine("Fuel Tank:", int64(ship1.MaxFuel), int64(ship2.MaxFuel), "units", true)
	s += m.renderComparisonLine("Max Crew:", int64(ship1.MaxCrew), int64(ship2.MaxCrew), "", true)
	s += m.renderComparisonLine("Outfit Space:", int64(ship1.OutfitSpace), int64(ship2.OutfitSpace), "", true)
	s += "\n"

	// Performance ratings
	s += subtitleStyle.Render("Performance Ratings:") + "\n"
	s += m.renderRatingComparison("Combat:", m.calculateCombatRating(ship1), m.calculateCombatRating(ship2))
	s += m.renderRatingComparison("Trading:", m.calculateTradingRating(ship1), m.calculateTradingRating(ship2))
	s += m.renderRatingComparison("Speed:", m.calculateSpeedRating(ship1), m.calculateSpeedRating(ship2))
	s += m.renderRatingComparison("Overall:", m.calculateOverallRating(ship1), m.calculateOverallRating(ship2))
	s += "\n"

	// Upgrade recommendations
	if m.currentShip != nil {
		s += m.renderUpgradeRecommendation(ship1, ship2)
	}

	// Help text
	s += renderFooter("C: New Comparison  •  ESC: Back to List")

	return s
}

func (m Model) renderComparisonLine(label string, val1, val2 int64, unit string, higherBetter bool) string {
	// Format values
	str1 := fmt.Sprintf("%d", val1)
	str2 := fmt.Sprintf("%d", val2)
	if unit != "" {
		str1 += " " + unit
		str2 += " " + unit
	}

	// Calculate difference
	diff := val2 - val1
	diffStr := ""
	if diff > 0 {
		diffStr = fmt.Sprintf("+%d", diff)
		if higherBetter {
			diffStr = statsStyle.Render(diffStr)
		} else {
			diffStr = errorStyle.Render(diffStr)
		}
	} else if diff < 0 {
		diffStr = fmt.Sprintf("%d", diff)
		if higherBetter {
			diffStr = errorStyle.Render(diffStr)
		} else {
			diffStr = statsStyle.Render(diffStr)
		}
	} else {
		diffStr = helpStyle.Render("=")
	}

	// Visual bars
	maxVal := val1
	if val2 > val1 {
		maxVal = val2
	}
	bar1 := m.renderStatBar(int(val1), int(maxVal), 20)
	bar2 := m.renderStatBar(int(val2), int(maxVal), 20)

	return fmt.Sprintf("  %-18s %-25s %s   %s\n                      %-25s %s\n",
		label,
		str1+" "+bar1,
		diffStr,
		"",
		str2+" "+bar2,
		"")
}

func (m Model) renderStatBar(value, maxValue, width int) string {
	if maxValue == 0 {
		return strings.Repeat("░", width)
	}
	filled := (value * width) / maxValue
	if filled > width {
		filled = width
	}
	return statsStyle.Render(strings.Repeat("█", filled)) + helpStyle.Render(strings.Repeat("░", width-filled))
}

func (m Model) renderRatingComparison(label string, rating1, rating2 float64) string {
	str1 := fmt.Sprintf("%.1f/10", rating1)
	str2 := fmt.Sprintf("%.1f/10", rating2)

	diff := rating2 - rating1
	diffStr := ""
	if diff > 0.1 {
		diffStr = statsStyle.Render(fmt.Sprintf("+%.1f", diff))
	} else if diff < -0.1 {
		diffStr = errorStyle.Render(fmt.Sprintf("%.1f", diff))
	} else {
		diffStr = helpStyle.Render("≈")
	}

	stars1 := m.renderStars(rating1)
	stars2 := m.renderStars(rating2)

	return fmt.Sprintf("  %-18s %-15s %s %s\n                      %-15s %s\n",
		label,
		str1+" "+stars1,
		diffStr,
		"",
		str2+" "+stars2,
		"")
}

func (m Model) renderStars(rating float64) string {
	stars := int(rating)
	halfStar := (rating - float64(stars)) >= 0.5
	result := strings.Repeat("★", stars)
	if halfStar && stars < 10 {
		result += "½"
		stars++
	}
	if stars < 10 {
		result += strings.Repeat("☆", 10-stars)
	}
	return result
}

func (m Model) calculateCombatRating(ship *models.ShipType) float64 {
	// Weighted combination of combat stats
	hullScore := float64(ship.MaxHull) / 1000.0
	shieldScore := float64(ship.MaxShields) / 500.0
	weaponScore := float64(ship.WeaponSlots) * 1.5
	maneuverScore := float64(ship.Maneuverability) / 10.0

	rating := (hullScore*2 + shieldScore*2 + weaponScore*3 + maneuverScore) / 2.0
	if rating > 10.0 {
		rating = 10.0
	}
	return rating
}

func (m Model) calculateTradingRating(ship *models.ShipType) float64 {
	// Weighted combination of trading stats
	cargoScore := float64(ship.CargoSpace) / 20.0
	fuelScore := float64(ship.MaxFuel) / 100.0
	speedScore := float64(ship.Speed) / 5.0

	rating := (cargoScore*4 + fuelScore*2 + speedScore*2) / 2.0
	if rating > 10.0 {
		rating = 10.0
	}
	return rating
}

func (m Model) calculateSpeedRating(ship *models.ShipType) float64 {
	speedScore := float64(ship.Speed) * 2.0
	maneuverScore := float64(ship.Maneuverability)

	rating := (speedScore + maneuverScore) / 2.0
	if rating > 10.0 {
		rating = 10.0
	}
	return rating
}

func (m Model) calculateOverallRating(ship *models.ShipType) float64 {
	combat := m.calculateCombatRating(ship)
	trading := m.calculateTradingRating(ship)
	speed := m.calculateSpeedRating(ship)

	return (combat + trading + speed) / 3.0
}

func (m Model) renderUpgradeRecommendation(ship1, ship2 *models.ShipType) string {
	currentType := models.GetShipTypeByID(m.currentShip.TypeID)
	if currentType == nil {
		return ""
	}

	s := subtitleStyle.Render("Upgrade Recommendation:") + "\n"

	// Calculate value improvements
	combat1 := m.calculateCombatRating(ship1) - m.calculateCombatRating(currentType)
	combat2 := m.calculateCombatRating(ship2) - m.calculateCombatRating(currentType)
	trading1 := m.calculateTradingRating(ship1) - m.calculateTradingRating(currentType)
	trading2 := m.calculateTradingRating(ship2) - m.calculateTradingRating(currentType)

	tradeIn := int64(float64(currentType.Price) * 0.70)
	cost1 := ship1.Price - tradeIn
	cost2 := ship2.Price - tradeIn

	// Calculate value per credit
	value1 := (combat1*2 + trading1*3) / float64(cost1)
	value2 := (combat2*2 + trading2*3) / float64(cost2)

	if value1 > value2 {
		s += fmt.Sprintf("  → %s offers better value for credits\n", statsStyle.Render(ship1.Name))
		s += fmt.Sprintf("    Value score: %.4f vs %.4f\n", value1, value2)
	} else if value2 > value1 {
		s += fmt.Sprintf("  → %s offers better value for credits\n", statsStyle.Render(ship2.Name))
		s += fmt.Sprintf("    Value score: %.4f vs %.4f\n", value2, value1)
	} else {
		s += "  → Both ships offer similar value\n"
	}

	// Specific recommendations
	if combat1 > 2.0 || combat2 > 2.0 {
		better := ship1.Name
		if combat2 > combat1 {
			better = ship2.Name
		}
		s += fmt.Sprintf("  • For combat: %s\n", statsStyle.Render(better))
	}
	if trading1 > 2.0 || trading2 > 2.0 {
		better := ship1.Name
		if trading2 > trading1 {
			better = ship2.Name
		}
		s += fmt.Sprintf("  • For trading: %s\n", statsStyle.Render(better))
	}

	s += "\n"
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
