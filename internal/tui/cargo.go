// File: internal/tui/cargo.go
// Project: Terminal Velocity
// Description: Cargo hold screen - Inventory management and jettison interface
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The cargo hold screen provides:
// - Complete cargo inventory display with quantities
// - Cargo capacity tracking (current/max tons)
// - Jettison (drop) functionality to free up space
// - Quantity selection for partial jettison
// - Sorted cargo list for easy navigation
//
// Key Features:
// - View all cargo items with commodities and quantities
// - Press 'd' to enter jettison mode
// - Adjust jettison quantity with +/- keys
// - Confirm jettison to remove items from cargo
// - Automatic cargo weight calculation

package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// cargoModel contains the state for the cargo hold screen.
// Manages inventory display, jettison operations, and error states.
type cargoModel struct {
	cursor       int               // Current cursor position in cargo list
	mode         string            // Current mode: "view" or "jettison"
	selectedItem *models.CargoItem // Item selected for jettison
	jettisonQty  int               // Quantity to jettison (1 to item quantity)
	error        string            // Error or status message to display
}

// cargoJettisonedMsg is sent when cargo jettison operation completes.
// Contains success status and any error that occurred.
type cargoJettisonedMsg struct {
	success bool  // True if jettison succeeded
	err     error // Error if jettison failed
}

// newCargoModel creates and initializes a new cargo hold screen model.
// Starts in view mode with cursor at top of list.
func newCargoModel() cargoModel {
	return cargoModel{
		cursor:      0,
		mode:        "view",
		jettisonQty: 1,
	}
}

// updateCargo handles input and state updates for the cargo hold screen.
//
// Key Bindings (View Mode):
//   - esc/backspace: Return to main menu
//   - up/k: Move cursor up in cargo list
//   - down/j: Move cursor down in cargo list
//   - d: Enter jettison mode for selected item
//
// Key Bindings (Jettison Mode):
//   - esc: Cancel jettison, return to view mode
//   - +/=: Increase jettison quantity
//   - -/_: Decrease jettison quantity
//   - enter/space: Confirm and execute jettison
//
// Jettison Flow:
//   1. Select item with cursor, press 'd'
//   2. Adjust quantity with +/- (1 to item.Quantity)
//   3. Press enter to jettison
//   4. Cargo updated in database, ship reloaded
//
// Message Handling:
//   - cargoJettisonedMsg: Jettison complete, reload player/ship data
func (m Model) updateCargo(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.cargo.mode == "view" {
				// Go back to main menu
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel jettison and return to view mode
			m.cargo.mode = "view"
			m.cargo.jettisonQty = 1
			m.cargo.error = ""
			return m, nil

		case "backspace":
			// Quick return to main menu
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			// Move cursor up (vi-style navigation supported with k)
			if m.cargo.cursor > 0 {
				m.cargo.cursor--
			}

		case "down":
			// Move cursor down
			if m.currentShip != nil {
				maxCursor := len(m.currentShip.Cargo) - 1
				if m.cargo.cursor < maxCursor {
					m.cargo.cursor++
				}
			}

		case "d": // Jettison (drop) cargo
			if m.cargo.mode == "view" && m.currentShip != nil && m.cargo.cursor < len(m.currentShip.Cargo) {
				m.cargo.selectedItem = &m.currentShip.Cargo[m.cargo.cursor]
				m.cargo.mode = "jettison"
				m.cargo.jettisonQty = 1
				m.cargo.error = ""
			}

		case "+", "=":
			if m.cargo.mode == "jettison" && m.cargo.selectedItem != nil {
				if m.cargo.jettisonQty < m.cargo.selectedItem.Quantity {
					m.cargo.jettisonQty++
				}
			}

		case "-", "_":
			if m.cargo.mode == "jettison" {
				if m.cargo.jettisonQty > 1 {
					m.cargo.jettisonQty--
				}
			}

		case "enter", " ":
			if m.cargo.mode == "jettison" {
				return m, m.executeJettison()
			}
		}

	case cargoJettisonedMsg:
		if msg.success {
			// Reload ship data
			m.cargo.mode = "view"
			m.cargo.jettisonQty = 1
			m.cargo.selectedItem = nil
			m.cargo.error = "Cargo jettisoned successfully"
			// Reload player/ship to refresh cargo
			return m, m.loadPlayer()
		} else {
			m.cargo.error = fmt.Sprintf("Failed to jettison: %v", msg.err)
			m.cargo.mode = "view"
		}
	}

	return m, nil
}

// viewCargo renders the cargo hold screen.
//
// Layout (View Mode):
//   - Header: Player stats (name, credits, location)
//   - Title: "=== Cargo Hold ==="
//   - Capacity: Current/Max cargo weight
//   - Cargo List: Items with quantities and weights
//   - Footer: Key bindings help
//
// Layout (Jettison Mode):
//   - Same as view mode
//   - Plus: Jettison confirmation panel
//   - Shows: Item, quantity selector, total weight to jettison
//
// Visual Features:
//   - Cargo items sorted alphabetically
//   - Selected item highlighted
//   - Color-coded capacity (green < 75%, yellow < 95%, red >= 95%)
//   - Empty cargo message when no items
func (m Model) viewCargo() string {
	// Header with player stats
	locationName := "Space"
	if m.player != nil && m.player.CurrentSystem.String() != "00000000-0000-0000-0000-000000000000" {
		locationName = "Space"
	}
	s := renderHeader(m.username, m.player.Credits, locationName)
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Cargo Hold ===") + "\n\n"

	// Error or status message
	if m.cargo.error != "" {
		s += helpStyle.Render(m.cargo.error) + "\n\n"
	}

	// Validate ship availability
	if m.currentShip == nil {
		s += errorStyle.Render("No ship available") + "\n"
		return s
	}

	// Get ship type for cargo capacity
	shipType := models.GetShipTypeByID(m.currentShip.TypeID)
	if shipType == nil {
		s += errorStyle.Render("Unknown ship type") + "\n"
		return s
	}

	// Mode-specific view
	switch m.cargo.mode {
	case "view":
		s += m.viewCargoList(shipType)
	case "jettison":
		s += m.viewJettisonInterface()
	default:
		s += "Unknown mode\n"
	}

	return s
}

func (m Model) viewCargoList(shipType *models.ShipType) string {
	s := ""

	// Ship and cargo info
	s += fmt.Sprintf("Ship: %s (%s)\n", statsStyle.Render(m.currentShip.Name), shipType.Name)

	cargoUsed := m.currentShip.GetCargoUsed()
	cargoSpace := shipType.CargoSpace
	cargoPercent := 0
	if cargoSpace > 0 {
		cargoPercent = (cargoUsed * 100) / cargoSpace
	}

	s += fmt.Sprintf("Cargo: %s / %d (%d%% full)\n\n",
		statsStyle.Render(fmt.Sprintf("%d", cargoUsed)),
		cargoSpace,
		cargoPercent)

	// Check if cargo is empty
	if len(m.currentShip.Cargo) == 0 {
		s += helpStyle.Render("Cargo hold is empty") + "\n\n"
		s += renderFooter("ESC: Main Menu")
		return s
	}

	// Sort cargo by commodity name
	sortedCargo := make([]models.CargoItem, len(m.currentShip.Cargo))
	copy(sortedCargo, m.currentShip.Cargo)
	sort.Slice(sortedCargo, func(i, j int) bool {
		commI := models.GetCommodityByID(sortedCargo[i].CommodityID)
		commJ := models.GetCommodityByID(sortedCargo[j].CommodityID)
		if commI != nil && commJ != nil {
			return commI.Name < commJ.Name
		}
		return sortedCargo[i].CommodityID < sortedCargo[j].CommodityID
	})

	// Cargo table header
	s += "Commodity                 Category        Quantity    Space\n"
	s += strings.Repeat("─", 78) + "\n"

	// List cargo items
	for i, item := range sortedCargo {
		commodity := models.GetCommodityByID(item.CommodityID)
		if commodity == nil {
			continue
		}

		line := fmt.Sprintf("%-25s %-15s %-11d %-7d",
			commodity.Name,
			commodity.Category,
			item.Quantity,
			item.Quantity, // For now, 1 unit = 1 cargo space
		)

		if i == m.cargo.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	// Help text
	s += renderFooter("↑/↓: Select  •  D: Drop/Jettison  •  ESC: Main Menu")

	return s
}

func (m Model) viewJettisonInterface() string {
	if m.cargo.selectedItem == nil {
		return "No item selected\n"
	}

	s := ""

	commodity := models.GetCommodityByID(m.cargo.selectedItem.CommodityID)
	if commodity == nil {
		return "Unknown commodity\n"
	}

	// Commodity info
	s += fmt.Sprintf("Jettisoning: %s\n", statsStyle.Render(commodity.Name))
	s += fmt.Sprintf("Description: %s\n\n", commodity.Description)

	// Quantity
	s += fmt.Sprintf("In cargo: %s units\n", statsStyle.Render(fmt.Sprintf("%d", m.cargo.selectedItem.Quantity)))
	s += fmt.Sprintf("Jettison quantity: %s\n\n", statsStyle.Render(fmt.Sprintf("%d", m.cargo.jettisonQty)))

	// Warning
	s += errorStyle.Render("⚠ Warning: Jettisoned cargo will be lost!") + "\n\n"

	// Help text
	s += helpStyle.Render("+/-: Adjust quantity  •  Enter: Confirm  •  ESC: Cancel")

	return s
}

// executeJettison jettisons selected cargo
func (m Model) executeJettison() tea.Cmd {
	return func() tea.Msg {
		// Validate
		if m.cargo.selectedItem == nil {
			return cargoJettisonedMsg{
				success: false,
				err:     fmt.Errorf("no item selected"),
			}
		}

		if m.currentShip == nil {
			return cargoJettisonedMsg{
				success: false,
				err:     fmt.Errorf("no ship available"),
			}
		}

		if m.cargo.jettisonQty > m.cargo.selectedItem.Quantity {
			return cargoJettisonedMsg{
				success: false,
				err:     fmt.Errorf("invalid quantity"),
			}
		}

		// Remove cargo from ship in database
		ctx := context.Background()
		err := m.shipRepo.RemoveCargo(ctx, m.currentShip.ID, m.cargo.selectedItem.CommodityID, m.cargo.jettisonQty)
		if err != nil {
			return cargoJettisonedMsg{
				success: false,
				err:     fmt.Errorf("failed to remove cargo: %w", err),
			}
		}

		return cargoJettisonedMsg{
			success: true,
			err:     nil,
		}
	}
}
