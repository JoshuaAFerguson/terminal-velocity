// File: internal/tui/cargo.go
// Project: Terminal Velocity
// Description: Terminal UI component for cargo
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type cargoModel struct {
	cursor       int
	mode         string // "view", "jettison"
	selectedItem *models.CargoItem
	jettisonQty  int
	error        string
}

type cargoJettisonedMsg struct {
	success bool
	err     error
}

func newCargoModel() cargoModel {
	return cargoModel{
		cursor:      0,
		mode:        "view",
		jettisonQty: 1,
	}
}

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
			// Cancel jettison
			m.cargo.mode = "view"
			m.cargo.jettisonQty = 1
			m.cargo.error = ""
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.cargo.cursor > 0 {
				m.cargo.cursor--
			}

		case "down":
			if m.currentShip != nil {
				maxCursor := len(m.currentShip.Cargo) - 1
				if m.cargo.cursor < maxCursor {
					m.cargo.cursor++
				}
			}

		case "d": // Jettison (drop)
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

	// Error display
	if m.cargo.error != "" {
		s += helpStyle.Render(m.cargo.error) + "\n\n"
	}

	// Ship info
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
