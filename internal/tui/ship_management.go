// File: internal/tui/ship_management.go
// Project: Terminal Velocity
// Description: Ship management screen - Multi-ship inventory and switching interface
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The ship management screen allows players to:
// - View all owned ships with their current status
// - Switch active ship for flying
// - Rename ships with custom names
// - View detailed ship information including cargo and equipment
// - Monitor ship health (hull/shields), fuel, and crew
// - See cargo contents and equipment loadouts
//
// Ship Management Features:
// - List all owned ships with key stats
// - Active ship clearly marked with asterisk (*)
// - Switch between ships (updates player's current ship)
// - Rename ships with live input (max 30 characters)
// - Detailed view shows full ship status and equipment
// - Color-coded damage warnings (hull < 50% shown in red)
// - Cargo contents expanded with commodity names
// - Equipment lists (weapons and outfits) with names and stats

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// shipManagementModel contains the state for the ship management screen.
// Manages multiple ship inventory, switching, and renaming operations.
type shipManagementModel struct {
	cursor       int            // Current cursor position in ship list
	mode         string         // Current mode: "list", "details", "rename", "confirm_switch"
	selectedShip *models.Ship   // Ship selected for viewing or operations
	ownedShips   []*models.Ship // All ships owned by the player
	renameInput  string         // Input buffer for ship renaming
	loading      bool           // True while loading ship data
	error        string         // Error or status message to display
}

// shipsLoadedMsg is sent when owned ships have been loaded from database.
// Contains all ships owned by the player.
type shipsLoadedMsg struct {
	ships []*models.Ship // Player's owned ships
	err   error          // Error if loading failed
}

// shipSwitchedMsg is sent when player switches active ship.
// Contains the newly active ship and success status.
type shipSwitchedMsg struct {
	success bool         // True if switch succeeded
	ship    *models.Ship // The newly active ship
	err     error        // Error if switch failed
}

// shipRenamedMsg is sent when a ship rename operation completes.
// Contains success status and any error that occurred.
type shipRenamedMsg struct {
	success bool  // True if rename succeeded
	err     error // Error if rename failed
}

// newShipManagementModel creates and initializes a new ship management screen model.
// Sets loading flag to true to trigger ship list load.
func newShipManagementModel() shipManagementModel {
	return shipManagementModel{
		cursor:  0,
		mode:    "list",
		loading: true,
	}
}

// updateShipManagement handles input and state updates for the ship management screen.
//
// Key Bindings (List Mode):
//   - esc/backspace: Return to main menu
//   - up/k: Move cursor up in ship list
//   - down/j: Move cursor down in ship list
//   - enter/space: View detailed ship information
//
// Key Bindings (Details Mode):
//   - esc: Return to ship list
//   - s: Switch to this ship (if not already active)
//   - r: Rename this ship
//
// Key Bindings (Rename Mode):
//   - esc: Cancel rename
//   - enter: Save new name
//   - backspace: Delete character
//   - Any character: Add to name (max 30 chars)
//
// Key Bindings (Confirm Switch Mode):
//   - esc: Cancel switch
//   - enter/space: Confirm switch
//
// Ship Switch Workflow:
//   1. Select ship from list
//   2. View ship details
//   3. Press 's' to initiate switch
//   4. Confirm switch operation
//   5. Player's active ship updated in database
//   6. Reload player data
//   7. Return to list with success message
//
// Message Handling:
//   - shipsLoadedMsg: Display owned ships
//   - shipSwitchedMsg: Update player state, show success
//   - shipRenamedMsg: Reload ships, show success
func (m Model) updateShipManagement(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle rename input mode
		if m.shipManagement.mode == "rename" {
			switch msg.String() {
			case "enter":
				// Save new name
				if m.shipManagement.renameInput != "" {
					return m, m.executeRename()
				}
				m.shipManagement.mode = "details"
				m.shipManagement.renameInput = ""
				return m, nil

			case "esc":
				// Cancel rename
				m.shipManagement.mode = "details"
				m.shipManagement.renameInput = ""
				return m, nil

			case "backspace":
				if len(m.shipManagement.renameInput) > 0 {
					m.shipManagement.renameInput = m.shipManagement.renameInput[:len(m.shipManagement.renameInput)-1]
				}
				return m, nil

			default:
				// Add character to input
				if len(msg.String()) == 1 && len(m.shipManagement.renameInput) < 30 {
					m.shipManagement.renameInput += msg.String()
				}
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "esc":
			if m.shipManagement.mode == "list" {
				// Go back to main menu
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel current operation
			m.shipManagement.mode = "list"
			m.shipManagement.error = ""
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.shipManagement.cursor > 0 {
				m.shipManagement.cursor--
			}

		case "down", "j":
			maxCursor := len(m.shipManagement.ownedShips) - 1
			if m.shipManagement.cursor < maxCursor {
				m.shipManagement.cursor++
			}

		case "enter", " ":
			if m.shipManagement.mode == "list" && m.shipManagement.cursor < len(m.shipManagement.ownedShips) {
				// View ship details
				m.shipManagement.selectedShip = m.shipManagement.ownedShips[m.shipManagement.cursor]
				m.shipManagement.mode = "details"
				m.shipManagement.error = ""
			} else if m.shipManagement.mode == "confirm_switch" {
				// Execute ship switch
				return m, m.executeSwitchShip()
			}

		case "s": // Switch active ship
			if m.shipManagement.mode == "details" {
				// Check if this is already the active ship
				if m.shipManagement.selectedShip.ID == m.currentShip.ID {
					m.shipManagement.error = "This is already your active ship"
				} else {
					m.shipManagement.mode = "confirm_switch"
				}
			}

		case "r": // Rename ship
			if m.shipManagement.mode == "details" {
				m.shipManagement.mode = "rename"
				m.shipManagement.renameInput = m.shipManagement.selectedShip.Name
			}
		}

	case shipsLoadedMsg:
		m.shipManagement.loading = false
		if msg.err != nil {
			m.shipManagement.error = fmt.Sprintf("Failed to load ships: %v", msg.err)
		} else {
			m.shipManagement.ownedShips = msg.ships
			m.shipManagement.error = ""
		}

	case shipSwitchedMsg:
		if msg.success {
			m.shipManagement.mode = "list"
			m.shipManagement.selectedShip = nil
			m.shipManagement.error = "Active ship changed successfully!"
			// Reload player and ship data
			return m, m.loadPlayer()
		} else {
			m.shipManagement.error = fmt.Sprintf("Switch failed: %v", msg.err)
			m.shipManagement.mode = "details"
		}

	case shipRenamedMsg:
		if msg.success {
			m.shipManagement.mode = "details"
			m.shipManagement.error = "Ship renamed successfully!"
			// Reload ships
			return m, m.loadOwnedShips()
		} else {
			m.shipManagement.error = fmt.Sprintf("Rename failed: %v", msg.err)
			m.shipManagement.mode = "details"
		}
	}

	return m, nil
}

// viewShipManagement renders the ship management screen.
//
// Layout (List Mode):
//   - Header: Player stats (name, credits, "Ship Management")
//   - Title: "=== Ship Management ==="
//   - Active Ship: Current active ship name
//   - Ship Count: Total ships owned
//   - Ship Table: Name, type, hull, shields, cargo, fuel (active marked with *)
//   - Footer: Key bindings help
//
// Layout (Details Mode):
//   - Ship name and type (active status if applicable)
//   - Current Status: Hull, shields, fuel, crew (with health warnings)
//   - Cargo Hold: Space used/max, contents list
//   - Equipment: Weapons and outfits lists
//   - Footer: Switch/rename options
//
// Layout (Rename Mode):
//   - Current name display
//   - Live input field with cursor
//   - Character count and instructions
//   - Footer: Confirm or cancel
//
// Layout (Confirm Switch Mode):
//   - From/to ship names
//   - New ship stats preview
//   - Confirmation prompt
//   - Footer: Confirm or cancel
//
// Visual Features:
//   - Active ship marked with asterisk (*)
//   - Damaged ships shown with red health warnings (hull < 50%)
//   - Selected ship highlighted with cursor
//   - Cargo percentage displayed with current/max
//   - Equipment lists expanded with item names
func (m Model) viewShipManagement() string {
	// Header with player stats
	s := renderHeader(m.username, m.player.Credits, "Ship Management")
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Ship Management ===") + "\n\n"

	// Error display
	if m.shipManagement.error != "" {
		s += helpStyle.Render(m.shipManagement.error) + "\n\n"
	}

	// Loading state
	if m.shipManagement.loading {
		s += "Loading ships...\n"
		return s
	}

	// Mode-specific view
	switch m.shipManagement.mode {
	case "list":
		s += m.viewShipInventory()
	case "details":
		s += m.viewShipManagementDetails()
	case "rename":
		s += m.viewRenamePrompt()
	case "confirm_switch":
		s += m.viewSwitchConfirmation()
	default:
		s += "Unknown mode\n"
	}

	return s
}

func (m Model) viewShipInventory() string {
	s := ""

	// Check if any ships owned
	if len(m.shipManagement.ownedShips) == 0 {
		s += helpStyle.Render("You don't own any ships yet!") + "\n\n"
		s += "Visit the shipyard to purchase your first ship.\n\n"
		s += renderFooter("ESC: Main Menu")
		return s
	}

	// Current active ship info
	s += fmt.Sprintf("Active Ship: %s\n", statsStyle.Render(m.currentShip.Name))
	s += fmt.Sprintf("Total Ships: %d\n\n", len(m.shipManagement.ownedShips))

	// Ship table header
	s += "Ship Name                 Type           Hull      Shields   Cargo     Fuel\n"
	s += strings.Repeat("─", 78) + "\n"

	// List ships
	for i, ship := range m.shipManagement.ownedShips {
		shipType := models.GetShipTypeByID(ship.TypeID)
		if shipType == nil {
			continue
		}

		// Mark active ship
		active := ""
		if ship.ID == m.currentShip.ID {
			active = "* "
		}

		// Get cargo used
		cargoUsed := ship.GetCargoUsed()

		line := fmt.Sprintf("%s%-23s %-14s %-9s %-9s %-9s %-6s",
			active,
			ship.Name,
			shipType.Name,
			fmt.Sprintf("%d/%d", ship.Hull, shipType.MaxHull),
			fmt.Sprintf("%d/%d", ship.Shields, shipType.MaxShields),
			fmt.Sprintf("%d/%d", cargoUsed, shipType.CargoSpace),
			fmt.Sprintf("%d/%d", ship.Fuel, shipType.MaxFuel),
		)

		if i == m.shipManagement.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	// Help text
	s += "\n" + renderFooter("↑/↓: Select  •  Enter: View Details  •  ESC: Main Menu")

	return s
}

func (m Model) viewShipManagementDetails() string {
	if m.shipManagement.selectedShip == nil {
		return "No ship selected\n"
	}

	ship := m.shipManagement.selectedShip
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return "Unknown ship type\n"
	}

	s := ""

	// Ship name and type
	s += titleStyle.Render(ship.Name)
	if ship.ID == m.currentShip.ID {
		s += statsStyle.Render(" (Active)")
	}
	s += "\n"
	s += fmt.Sprintf("%s (%s)\n\n", shipType.Name, shipType.Class)

	// Current status
	s += "Current Status:\n"
	s += fmt.Sprintf("  Hull:    %s / %d HP",
		statsStyle.Render(fmt.Sprintf("%d", ship.Hull)), shipType.MaxHull)
	if ship.Hull < shipType.MaxHull/2 {
		s += errorStyle.Render(" (Damaged)")
	}
	s += "\n"
	s += fmt.Sprintf("  Shields: %s / %d HP\n",
		statsStyle.Render(fmt.Sprintf("%d", ship.Shields)), shipType.MaxShields)
	s += fmt.Sprintf("  Fuel:    %s / %d units\n",
		statsStyle.Render(fmt.Sprintf("%d", ship.Fuel)), shipType.MaxFuel)
	s += fmt.Sprintf("  Crew:    %s / %d\n\n",
		statsStyle.Render(fmt.Sprintf("%d", ship.Crew)), shipType.MaxCrew)

	// Cargo
	cargoUsed := ship.GetCargoUsed()
	cargoPercent := 0
	if shipType.CargoSpace > 0 {
		cargoPercent = (cargoUsed * 100) / shipType.CargoSpace
	}
	s += "Cargo Hold:\n"
	s += fmt.Sprintf("  Space:   %s / %d tons (%d%% full)\n",
		statsStyle.Render(fmt.Sprintf("%d", cargoUsed)), shipType.CargoSpace, cargoPercent)
	if len(ship.Cargo) > 0 {
		s += "  Contents:\n"
		for _, item := range ship.Cargo {
			commodity := models.GetCommodityByID(item.CommodityID)
			if commodity != nil {
				s += fmt.Sprintf("    - %s (%d tons)\n", commodity.Name, item.Quantity)
			}
		}
	}
	s += "\n"

	// Equipment
	s += "Equipment:\n"
	s += fmt.Sprintf("  Weapons: %d installed (%d/%d slots)\n",
		len(ship.Weapons), len(ship.Weapons), shipType.WeaponSlots)
	for _, weaponID := range ship.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			s += fmt.Sprintf("    - %s\n", weapon.Name)
		}
	}
	s += "\n"

	// Calculate outfit space used
	outfitSpaceUsed := 0
	for _, weaponID := range ship.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			outfitSpaceUsed += weapon.OutfitSpace
		}
	}
	for _, outfitID := range ship.Outfits {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			outfitSpaceUsed += outfit.OutfitSpace
		}
	}

	s += fmt.Sprintf("  Outfits: %d installed (%d/%d space)\n",
		len(ship.Outfits), outfitSpaceUsed, shipType.OutfitSpace)
	for _, outfitID := range ship.Outfits {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			s += fmt.Sprintf("    - %s\n", outfit.Name)
		}
	}
	s += "\n"

	// Actions
	helpText := ""
	if ship.ID != m.currentShip.ID {
		helpText = "S: Switch to this ship  •  "
	}
	helpText += "R: Rename  •  ESC: Back"
	s += helpStyle.Render(helpText)

	return s
}

func (m Model) viewRenamePrompt() string {
	s := ""

	s += errorStyle.Render("=== Rename Ship ===") + "\n\n"
	s += fmt.Sprintf("Current name: %s\n", m.shipManagement.selectedShip.Name)
	s += fmt.Sprintf("New name:     %s_\n\n", m.shipManagement.renameInput)
	s += helpStyle.Render("Type new name (max 30 chars)  •  Enter: Confirm  •  ESC: Cancel")

	return s
}

func (m Model) viewSwitchConfirmation() string {
	if m.shipManagement.selectedShip == nil {
		return "No ship selected\n"
	}

	s := ""

	s += errorStyle.Render("=== Switch Active Ship ===") + "\n\n"
	s += fmt.Sprintf("Switch from: %s\n", m.currentShip.Name)
	s += fmt.Sprintf("Switch to:   %s\n\n", m.shipManagement.selectedShip.Name)

	shipType := models.GetShipTypeByID(m.shipManagement.selectedShip.TypeID)
	if shipType != nil {
		s += "New ship stats:\n"
		s += fmt.Sprintf("  Class:   %s\n", shipType.Class)
		s += fmt.Sprintf("  Cargo:   %d tons\n", shipType.CargoSpace)
		s += fmt.Sprintf("  Weapons: %d slots\n", shipType.WeaponSlots)
		s += "\n"
	}

	s += helpStyle.Render("Enter: Confirm Switch  •  ESC: Cancel")

	return s
}

// loadOwnedShips loads all ships owned by the player
func (m Model) loadOwnedShips() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ships, err := m.shipRepo.GetByOwner(ctx, m.player.ID)
		if err != nil {
			return shipsLoadedMsg{err: err}
		}
		return shipsLoadedMsg{ships: ships, err: nil}
	}
}

// executeSwitchShip switches the active ship
func (m Model) executeSwitchShip() tea.Cmd {
	return func() tea.Msg {
		if m.shipManagement.selectedShip == nil {
			return shipSwitchedMsg{
				success: false,
				err:     fmt.Errorf("no ship selected"),
			}
		}

		ctx := context.Background()

		// Update player's ship_id
		err := m.playerRepo.UpdateShip(ctx, m.player.ID, m.shipManagement.selectedShip.ID)
		if err != nil {
			return shipSwitchedMsg{
				success: false,
				err:     fmt.Errorf("failed to update active ship: %w", err),
			}
		}

		return shipSwitchedMsg{
			success: true,
			ship:    m.shipManagement.selectedShip,
			err:     nil,
		}
	}
}

// executeRename renames the selected ship
func (m Model) executeRename() tea.Cmd {
	return func() tea.Msg {
		if m.shipManagement.selectedShip == nil {
			return shipRenamedMsg{
				success: false,
				err:     fmt.Errorf("no ship selected"),
			}
		}

		if m.shipManagement.renameInput == "" {
			return shipRenamedMsg{
				success: false,
				err:     fmt.Errorf("name cannot be empty"),
			}
		}

		ctx := context.Background()

		// Update ship name
		m.shipManagement.selectedShip.Name = m.shipManagement.renameInput
		err := m.shipRepo.Update(ctx, m.shipManagement.selectedShip)
		if err != nil {
			return shipRenamedMsg{
				success: false,
				err:     fmt.Errorf("failed to rename ship: %w", err),
			}
		}

		return shipRenamedMsg{
			success: true,
			err:     nil,
		}
	}
}
