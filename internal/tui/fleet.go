// File: internal/tui/fleet.go
// Project: Terminal Velocity
// Description: Fleet management TUI screen for multi-ship ownership and escorts
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// Fleet screen modes
const (
	fleetModeMenu      = "menu"
	fleetModeShips     = "ships"
	fleetModeStored    = "stored"
	fleetModeEscorts   = "escorts"
	fleetModeFormation = "formation"
	fleetModeHireEscort = "hire_escort"
	fleetModeViewShip   = "view_ship"
	fleetModeViewEscort = "view_escort"
)

type fleetState struct {
	mode          string
	menuIndex     int
	selectedIndex int
	viewOffset    int

	// Fleet data
	currentFleet  *fleet.Fleet
	selectedShip  uuid.UUID
	selectedEscort uuid.UUID

	// Forms
	escortName     string
	escortBehavior fleet.EscortBehavior
	inputMode      string // "name", "ship", "behavior", "confirm"
	nameInput      string // Text input buffer for pilot name

	loading       bool
	error         string
	message       string
}

func newFleetState() fleetState {
	return fleetState{
		mode:      fleetModeMenu,
		menuIndex: 0,
	}
}

var (
	fleetMenuItems = []string{
		"View Fleet Ships",
		"Stored Ships",
		"Manage Escorts",
		"Fleet Formation",
		"Hire New Escort",
		"Pay Maintenance",
		"Back to Main Menu",
	}

	formationTypes = []fleet.FormationType{
		fleet.FormationLine,
		fleet.FormationWedge,
		fleet.FormationBox,
		fleet.FormationCircular,
	}

	behaviorTypes = []fleet.EscortBehavior{
		fleet.BehaviorDefensive,
		fleet.BehaviorAggressive,
		fleet.BehaviorPassive,
		fleet.BehaviorSupport,
	}
)

// updateFleet handles all fleet screen updates
func (m *Model) updateFleet(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.fleet.mode {
		case fleetModeMenu:
			return m.updateFleetMenu(msg)
		case fleetModeShips:
			return m.updateFleetShips(msg)
		case fleetModeStored:
			return m.updateFleetStored(msg)
		case fleetModeEscorts:
			return m.updateFleetEscorts(msg)
		case fleetModeFormation:
			return m.updateFleetFormation(msg)
		case fleetModeHireEscort:
			return m.updateFleetHireEscort(msg)
		case fleetModeViewShip:
			return m.updateFleetViewShip(msg)
		case fleetModeViewEscort:
			return m.updateFleetViewEscort(msg)
		}
	}

	return m, nil
}

// updateFleetMenu handles main fleet menu
func (m *Model) updateFleetMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.fleet.menuIndex > 0 {
			m.fleet.menuIndex--
		}
	case "down", "j":
		if m.fleet.menuIndex < len(fleetMenuItems)-1 {
			m.fleet.menuIndex++
		}
	case "enter":
		switch m.fleet.menuIndex {
		case 0: // View Fleet Ships
			m.fleet.mode = fleetModeShips
			m.fleet.selectedIndex = 0
			// Load fleet data from fleet manager
			if m.fleetManager != nil {
				m.fleet.currentFleet = m.fleetManager.GetOrCreateFleet(m.playerID)
			}
		case 1: // Stored Ships
			m.fleet.mode = fleetModeStored
			m.fleet.selectedIndex = 0
			// Fleet data already includes stored ships
			if m.fleetManager != nil {
				m.fleet.currentFleet = m.fleetManager.GetOrCreateFleet(m.playerID)
			}
		case 2: // Manage Escorts
			m.fleet.mode = fleetModeEscorts
			m.fleet.selectedIndex = 0
			// Fleet data already includes escorts
			if m.fleetManager != nil {
				m.fleet.currentFleet = m.fleetManager.GetOrCreateFleet(m.playerID)
			}
		case 3: // Fleet Formation
			m.fleet.mode = fleetModeFormation
			m.fleet.selectedIndex = 0
		case 4: // Hire New Escort
			m.fleet.mode = fleetModeHireEscort
			m.fleet.escortName = ""
			m.fleet.nameInput = ""
			m.fleet.inputMode = "name" // Start with name input
			m.fleet.selectedIndex = 0
		case 5: // Pay Maintenance
			// Pay maintenance early
			if m.fleetManager != nil {
				err := m.fleetManager.PayMaintenanceEarly(context.Background(), m.playerID)
				if err != nil {
					m.fleet.error = fmt.Sprintf("Failed to pay maintenance: %v", err)
				} else {
					m.fleet.message = "Maintenance paid! Escort loyalty increased."
				}
			}
		case 6: // Back
			m.screen = ScreenMainMenu
		}
	case "q", "esc":
		m.screen = ScreenMainMenu
	}

	return m, nil
}

// updateFleetShips handles fleet ships view
func (m *Model) updateFleetShips(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.fleet.currentFleet == nil {
		m.fleet.mode = fleetModeMenu
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.fleet.selectedIndex > 0 {
			m.fleet.selectedIndex--
		}
	case "down", "j":
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips)-1 {
			m.fleet.selectedIndex++
		}
	case "enter":
		// View ship details
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips) {
			m.fleet.selectedShip = m.fleet.currentFleet.OwnedShips[m.fleet.selectedIndex].ID
			m.fleet.mode = fleetModeViewShip
		}
	case "s":
		// Switch to this ship as flagship
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips) {
			shipID := m.fleet.currentFleet.OwnedShips[m.fleet.selectedIndex].ID
			// Call fleet manager to switch flagship
			if m.fleetManager != nil {
				err := m.fleetManager.SwitchFlagship(context.Background(), m.playerID, shipID)
				if err != nil {
					m.fleet.error = fmt.Sprintf("Failed to switch flagship: %v", err)
				} else {
					m.fleet.message = "Switched to new flagship!"
					// Reload current ship
					m.currentShip, _ = m.shipRepo.GetByID(context.Background(), shipID)
				}
			}
		}
	case "t":
		// Store ship at current planet
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips) {
			shipID := m.fleet.currentFleet.OwnedShips[m.fleet.selectedIndex].ID
			if shipID == m.fleet.currentFleet.FlagshipID {
				m.fleet.error = "Cannot store your active flagship!"
			} else {
				// Call fleet manager to store ship
				if m.fleetManager != nil && m.player.CurrentPlanet != nil {
					err := m.fleetManager.StoreShip(context.Background(), m.playerID, shipID, *m.player.CurrentPlanet, "Current Planet")
					if err != nil {
						m.fleet.error = fmt.Sprintf("Failed to store ship: %v", err)
					} else {
						m.fleet.message = "Ship stored at current planet"
					}
				} else {
					m.fleet.error = "Must be docked at a planet to store ships"
				}
			}
		}
	case "b", "q", "esc":
		m.fleet.mode = fleetModeMenu
	}

	return m, nil
}

// updateFleetStored handles stored ships view
func (m *Model) updateFleetStored(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.fleet.currentFleet == nil {
		m.fleet.mode = fleetModeMenu
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.fleet.selectedIndex > 0 {
			m.fleet.selectedIndex--
		}
	case "down", "j":
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.StoredShips)-1 {
			m.fleet.selectedIndex++
		}
	case "enter", "r":
		// Retrieve ship (must be at same planet)
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.StoredShips) {
			storedShip := m.fleet.currentFleet.StoredShips[m.fleet.selectedIndex]
			// Check if at correct planet and retrieve
			if m.fleetManager != nil && m.player.CurrentPlanet != nil {
				if storedShip.LocationID == *m.player.CurrentPlanet {
					err := m.fleetManager.RetrieveShip(context.Background(), m.playerID, storedShip.Ship.ID, *m.player.CurrentPlanet)
					if err != nil {
						m.fleet.error = fmt.Sprintf("Failed to retrieve ship: %v", err)
					} else {
						m.fleet.message = fmt.Sprintf("Retrieved ship from %s", storedShip.Location)
					}
				} else {
					m.fleet.error = "You must be at the planet where the ship is stored"
				}
			} else {
				m.fleet.error = "Must be docked at a planet to retrieve ships"
			}
		}
	case "b", "q", "esc":
		m.fleet.mode = fleetModeMenu
	}

	return m, nil
}

// updateFleetEscorts handles escorts view
func (m *Model) updateFleetEscorts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.fleet.currentFleet == nil {
		m.fleet.mode = fleetModeMenu
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.fleet.selectedIndex > 0 {
			m.fleet.selectedIndex--
		}
	case "down", "j":
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.Escorts)-1 {
			m.fleet.selectedIndex++
		}
	case "enter":
		// View escort details
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.Escorts) {
			m.fleet.selectedEscort = m.fleet.currentFleet.Escorts[m.fleet.selectedIndex].ID
			m.fleet.mode = fleetModeViewEscort
		}
	case "d":
		// Dismiss escort
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.Escorts) {
			// Call fleet manager to dismiss escort
			escort := m.fleet.currentFleet.Escorts[m.fleet.selectedIndex]
			if m.fleetManager != nil {
				err := m.fleetManager.DismissEscort(context.Background(), m.playerID, escort.ID)
				if err != nil {
					m.fleet.error = fmt.Sprintf("Failed to dismiss escort: %v", err)
				} else {
					m.fleet.message = "Escort dismissed"
					// Reload fleet data
					m.fleet.currentFleet = m.fleetManager.GetOrCreateFleet(m.playerID)
				}
			}
		}
	case "1", "2", "3", "4":
		// Change behavior: 1=Defensive, 2=Aggressive, 3=Passive, 4=Support
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.Escorts) {
			behaviors := map[string]fleet.EscortBehavior{
				"1": fleet.BehaviorDefensive,
				"2": fleet.BehaviorAggressive,
				"3": fleet.BehaviorPassive,
				"4": fleet.BehaviorSupport,
			}
			behavior := behaviors[msg.String()]
			// Call fleet manager to set behavior
			escort := m.fleet.currentFleet.Escorts[m.fleet.selectedIndex]
			if m.fleetManager != nil {
				err := m.fleetManager.SetEscortBehavior(m.playerID, escort.ID, behavior)
				if err != nil {
					m.fleet.error = fmt.Sprintf("Failed to set behavior: %v", err)
				} else {
					m.fleet.message = fmt.Sprintf("Escort behavior set to %s", behavior)
				}
			}
		}
	case "b", "q", "esc":
		m.fleet.mode = fleetModeMenu
	}

	return m, nil
}

// updateFleetFormation handles formation selection
func (m *Model) updateFleetFormation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.fleet.selectedIndex > 0 {
			m.fleet.selectedIndex--
		}
	case "down", "j":
		if m.fleet.selectedIndex < len(formationTypes)-1 {
			m.fleet.selectedIndex++
		}
	case "enter":
		formation := formationTypes[m.fleet.selectedIndex]
		// Call fleet manager to set formation
		if m.fleetManager != nil {
			err := m.fleetManager.SetFormation(m.playerID, formation)
			if err != nil {
				m.fleet.error = fmt.Sprintf("Failed to set formation: %v", err)
			} else {
				m.fleet.message = fmt.Sprintf("Formation set to %s", formation)
			}
		}
		m.fleet.mode = fleetModeMenu
	case "b", "q", "esc":
		m.fleet.mode = fleetModeMenu
	}

	return m, nil
}

// updateFleetHireEscort handles hiring new escorts
func (m *Model) updateFleetHireEscort(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.fleet.inputMode {
	case "name":
		// Handle name input
		switch msg.String() {
		case "esc", "q":
			m.fleet.mode = fleetModeMenu
		case "enter":
			if m.fleet.nameInput != "" {
				m.fleet.escortName = m.fleet.nameInput
				m.fleet.inputMode = "ship"
				m.fleet.selectedIndex = 0
			}
		case "backspace":
			if len(m.fleet.nameInput) > 0 {
				m.fleet.nameInput = m.fleet.nameInput[:len(m.fleet.nameInput)-1]
			}
		default:
			// Add character to name (limit to 20 chars)
			if len(msg.String()) == 1 && len(m.fleet.nameInput) < 20 {
				m.fleet.nameInput += msg.String()
			}
		}

	case "ship":
		// Handle ship selection from owned ships
		switch msg.String() {
		case "esc", "q":
			m.fleet.inputMode = "name"
		case "enter":
			// Select current ship
			if m.fleet.currentFleet != nil && len(m.fleet.currentFleet.OwnedShips) > 0 {
				if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips) {
					m.fleet.selectedShip = m.fleet.currentFleet.OwnedShips[m.fleet.selectedIndex].ID
					m.fleet.inputMode = "behavior"
					m.fleet.selectedIndex = 0
				}
			}
		case "up", "k":
			if m.fleet.selectedIndex > 0 {
				m.fleet.selectedIndex--
			}
		case "down", "j":
			if m.fleet.currentFleet != nil && m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips)-1 {
				m.fleet.selectedIndex++
			}
		}

	case "behavior":
		// Handle behavior selection
		switch msg.String() {
		case "esc", "q":
			m.fleet.inputMode = "ship"
			m.fleet.selectedIndex = 0
		case "enter":
			if m.fleet.selectedIndex < len(behaviorTypes) {
				m.fleet.escortBehavior = behaviorTypes[m.fleet.selectedIndex]
				m.fleet.inputMode = "confirm"
			}
		case "up", "k":
			if m.fleet.selectedIndex > 0 {
				m.fleet.selectedIndex--
			}
		case "down", "j":
			if m.fleet.selectedIndex < len(behaviorTypes)-1 {
				m.fleet.selectedIndex++
			}
		}

	case "confirm":
		// Confirmation screen
		switch msg.String() {
		case "esc", "q", "n":
			m.fleet.inputMode = "behavior"
			m.fleet.selectedIndex = 0
		case "enter", "y":
			// Hire escort with selected ship and behavior
			if m.fleetManager != nil && m.fleet.escortName != "" {
				// Find the selected ship
				var escortShip *models.Ship
				if m.fleet.currentFleet != nil {
					for _, ship := range m.fleet.currentFleet.OwnedShips {
						if ship.ID == m.fleet.selectedShip {
							escortShip = ship
							break
						}
					}
				}

				if escortShip != nil {
					_, err := m.fleetManager.HireEscort(context.Background(), m.playerID, escortShip, m.fleet.escortName, m.fleet.escortBehavior)
					if err != nil {
						m.fleet.error = fmt.Sprintf("Failed to hire escort: %v", err)
					} else {
						m.fleet.message = fmt.Sprintf("Escort pilot %s hired! They will follow and protect you.", m.fleet.escortName)
					}
				} else {
					m.fleet.error = "Selected ship not found"
				}
			} else {
				m.fleet.error = "Missing escort information"
			}
			m.fleet.mode = fleetModeMenu
		}
	}

	return m, nil
}

// updateFleetViewShip handles viewing ship details
func (m *Model) updateFleetViewShip(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "b":
		m.fleet.mode = fleetModeShips
		m.fleet.selectedShip = uuid.Nil
	}

	return m, nil
}

// updateFleetViewEscort handles viewing escort details
func (m *Model) updateFleetViewEscort(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "b":
		m.fleet.mode = fleetModeEscorts
		m.fleet.selectedEscort = uuid.Nil
	}

	return m, nil
}

// viewFleet renders the fleet screen
func (m *Model) viewFleet() string {
	var b strings.Builder

	// Header
	header := titleStyle.Render("╔═══════════════════════════════════════════════════════════════════════╗")
	title := titleStyle.Render("║                         FLEET MANAGEMENT                              ║")
	divider := titleStyle.Render("╠═══════════════════════════════════════════════════════════════════════╣")

	b.WriteString(header + "\n")
	b.WriteString(title + "\n")
	b.WriteString(divider + "\n")

	// Player info
	info := fmt.Sprintf("║ Credits: %s%-47d%s ║\n",
		fleetCreditStyle.Render(""), m.player.Credits, resetStyle.Render(""))
	b.WriteString(titleStyle.Render(info))

	// Fleet stats
	if m.fleet.currentFleet != nil {
		stats := fmt.Sprintf("║ Ships: %d | Stored: %d | Escorts: %d%37s ║\n",
			len(m.fleet.currentFleet.OwnedShips),
			len(m.fleet.currentFleet.StoredShips),
			len(m.fleet.currentFleet.Escorts),
			"")
		b.WriteString(titleStyle.Render(stats))
	}

	b.WriteString(titleStyle.Render("╠═══════════════════════════════════════════════════════════════════════╣") + "\n")

	// Content based on mode
	switch m.fleet.mode {
	case fleetModeMenu:
		b.WriteString(m.viewFleetMenu())
	case fleetModeShips:
		b.WriteString(m.viewFleetShips())
	case fleetModeStored:
		b.WriteString(m.viewFleetStored())
	case fleetModeEscorts:
		b.WriteString(m.viewFleetEscorts())
	case fleetModeFormation:
		b.WriteString(m.viewFleetFormation())
	case fleetModeHireEscort:
		b.WriteString(m.viewFleetHireEscort())
	case fleetModeViewShip:
		b.WriteString(m.viewFleetViewShip())
	case fleetModeViewEscort:
		b.WriteString(m.viewFleetViewEscort())
	}

	// Footer
	b.WriteString(titleStyle.Render("╚═══════════════════════════════════════════════════════════════════════╝") + "\n")

	// Help text
	switch m.fleet.mode {
	case fleetModeMenu:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select | Q: Back\n"))
	case fleetModeShips:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: View | S: Switch Flagship | T: Store | Q: Back\n"))
	case fleetModeStored:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | R: Retrieve (if at location) | Q: Back\n"))
	case fleetModeEscorts:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: View | 1-4: Change Behavior | D: Dismiss | Q: Back\n"))
	case fleetModeFormation:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select Formation | Q: Back\n"))
	case fleetModeHireEscort:
		switch m.fleet.inputMode {
		case "name":
			b.WriteString(helpStyle.Render("Type: Enter Name | Enter: Next | Q: Cancel\n"))
		case "ship":
			b.WriteString(helpStyle.Render("↑/↓: Select Ship | Enter: Next | Q: Back\n"))
		case "behavior":
			b.WriteString(helpStyle.Render("↑/↓: Select Behavior | Enter: Next | Q: Back\n"))
		case "confirm":
			b.WriteString(helpStyle.Render("Enter/Y: Confirm | Q/N: Cancel\n"))
		}
	default:
		b.WriteString(helpStyle.Render("Q: Back\n"))
	}

	// Messages
	if m.fleet.message != "" {
		b.WriteString("\n" + successStyle.Render(m.fleet.message) + "\n")
	}
	if m.fleet.error != "" {
		b.WriteString("\n" + errorStyle.Render(m.fleet.error) + "\n")
	}

	return b.String()
}

// viewFleetMenu renders the main menu
func (m *Model) viewFleetMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	for i, item := range fleetMenuItems {
		if i == m.fleet.menuIndex {
			line := fmt.Sprintf("║   ► %-65s ║", item)
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			line := fmt.Sprintf("║     %-65s ║", item)
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewFleetShips renders owned ships list
func (m *Model) viewFleetShips() string {
	var b strings.Builder

	if m.fleet.currentFleet == nil || len(m.fleet.currentFleet.OwnedShips) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No ships in fleet                                  ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-25s %-15s %-10s %-10s %-10s ║", "Ship", "Type", "Hull", "Shields", "Status")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	for i, ship := range m.fleet.currentFleet.OwnedShips {
		status := "Ready"
		if ship.ID == m.fleet.currentFleet.FlagshipID {
			status = "★ FLAGSHIP"
		}

		// Get ship type name from TypeID
		shipType := ship.TypeID
		if shipType == "" {
			shipType = "Unknown"
		}

		line := fmt.Sprintf("║ %-25s %-15s %7d%% %7d%% %-10s ║",
			truncateString(ship.Name, 25),
			shipType,
			ship.Hull,
			ship.Shields,
			status)

		if i == m.fleet.selectedIndex {
			b.WriteString(fleetSelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

// viewFleetStored renders stored ships list
func (m *Model) viewFleetStored() string {
	var b strings.Builder

	if m.fleet.currentFleet == nil || len(m.fleet.currentFleet.StoredShips) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No stored ships                                    ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-25s %-25s %-15s ║", "Ship", "Location", "Stored Date")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	for i, stored := range m.fleet.currentFleet.StoredShips {
		storedDate := stored.StoredAt.Format("2006-01-02")

		line := fmt.Sprintf("║ %-25s %-25s %-15s ║",
			truncateString(stored.Ship.Name, 25),
			truncateString(stored.Location, 25),
			storedDate)

		if i == m.fleet.selectedIndex {
			b.WriteString(fleetSelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

// viewFleetEscorts renders escorts list
func (m *Model) viewFleetEscorts() string {
	var b strings.Builder

	if m.fleet.currentFleet == nil || len(m.fleet.currentFleet.Escorts) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No active escorts                                  ║") + "\n")
		b.WriteString(titleStyle.Render("║         Press H to hire an escort from the main menu                  ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-20s %-15s %-12s %-10s %-10s ║", "Pilot", "Behavior", "Loyalty", "Level", "Status")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	for i, escort := range m.fleet.currentFleet.Escorts {
		loyaltyPercent := int(escort.Loyalty * 100)

		line := fmt.Sprintf("║ %-20s %-15s %10d%% %10d %-10s ║",
			truncateString(escort.Pilot, 20),
			escort.Behavior,
			loyaltyPercent,
			escort.Level,
			escort.Status)

		if i == m.fleet.selectedIndex {
			b.WriteString(fleetSelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ Behaviors: 1=Defensive 2=Aggressive 3=Passive 4=Support               ║") + "\n")

	return b.String()
}

// viewFleetFormation renders formation selection
func (m *Model) viewFleetFormation() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                   SELECT FLEET FORMATION                              ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	formations := map[fleet.FormationType]string{
		fleet.FormationLine:     "Line Formation    - Escorts follow in single file",
		fleet.FormationWedge:    "Wedge Formation   - V-shaped attack formation",
		fleet.FormationBox:      "Box Formation     - Square defensive formation",
		fleet.FormationCircular: "Circular Formation - Escorts orbit around flagship",
	}

	for i, formationType := range formationTypes {
		desc := formations[formationType]
		if i == m.fleet.selectedIndex {
			line := fmt.Sprintf("║   ► %-65s ║", desc)
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			line := fmt.Sprintf("║     %-65s ║", desc)
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewFleetHireEscort renders escort hiring screen
func (m *Model) viewFleetHireEscort() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                   HIRE ESCORT PILOT                                   ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ Hiring Cost: 50,000 credits                                           ║") + "\n")
	b.WriteString(titleStyle.Render("║ Daily Maintenance: 5,000 credits per escort                           ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	switch m.fleet.inputMode {
	case "name":
		b.WriteString(titleStyle.Render("║ Step 1/3: Enter Pilot Name                                            ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		nameDisplay := m.fleet.nameInput
		if nameDisplay == "" {
			nameDisplay = "_"
		}
		line := fmt.Sprintf("║ Name: %-63s ║", nameDisplay)
		b.WriteString(selectedStyle.Render(line) + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║ Press Enter to continue                                               ║") + "\n")

	case "ship":
		b.WriteString(titleStyle.Render("║ Step 2/3: Select Ship for Escort                                     ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		if m.fleet.currentFleet != nil && len(m.fleet.currentFleet.OwnedShips) > 0 {
			for i, ship := range m.fleet.currentFleet.OwnedShips {
				desc := fmt.Sprintf("%s (%s) - Hull: %d%%", ship.Name, ship.TypeID, ship.Hull)
				if i == m.fleet.selectedIndex {
					line := fmt.Sprintf("║   ► %-65s ║", desc)
					b.WriteString(selectedStyle.Render(line) + "\n")
				} else {
					line := fmt.Sprintf("║     %-65s ║", desc)
					b.WriteString(titleStyle.Render(line) + "\n")
				}
			}
		} else {
			b.WriteString(titleStyle.Render("║ No ships available for escort duty                                    ║") + "\n")
		}

	case "behavior":
		b.WriteString(titleStyle.Render("║ Step 3/3: Select Escort Behavior                                      ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		behaviors := map[fleet.EscortBehavior]string{
			fleet.BehaviorDefensive:  "Defensive  - Only engages when you're attacked",
			fleet.BehaviorAggressive: "Aggressive - Attacks all hostile targets",
			fleet.BehaviorPassive:    "Passive    - Never engages in combat",
			fleet.BehaviorSupport:    "Support    - Provides healing and buffs",
		}

		for i, behaviorType := range behaviorTypes {
			desc := behaviors[behaviorType]
			if i == m.fleet.selectedIndex {
				line := fmt.Sprintf("║   ► %-65s ║", desc)
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				line := fmt.Sprintf("║     %-65s ║", desc)
				b.WriteString(titleStyle.Render(line) + "\n")
			}
		}

	case "confirm":
		b.WriteString(titleStyle.Render("║ Confirm Escort Hire                                                   ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Pilot Name: %-58s ║", m.fleet.escortName)) + "\n")

		// Find selected ship name
		var shipName string
		if m.fleet.currentFleet != nil {
			for _, ship := range m.fleet.currentFleet.OwnedShips {
				if ship.ID == m.fleet.selectedShip {
					shipName = fmt.Sprintf("%s (%s)", ship.Name, ship.TypeID)
					break
				}
			}
		}
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Ship: %-63s ║", shipName)) + "\n")
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Behavior: %-59s ║", m.fleet.escortBehavior)) + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║ Press Enter to confirm, Q to cancel                                   ║") + "\n")
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewFleetViewShip renders ship details
func (m *Model) viewFleetViewShip() string {
	var b strings.Builder

	// Find selected ship
	var ship *models.Ship
	if m.fleet.currentFleet != nil {
		for _, s := range m.fleet.currentFleet.OwnedShips {
			if s.ID == m.fleet.selectedShip {
				ship = s
				break
			}
		}
	}

	if ship == nil {
		b.WriteString(titleStyle.Render("║ Ship not found                                                        ║") + "\n")
		return b.String()
	}

	b.WriteString(titleStyle.Render("║                   SHIP DETAILS                                        ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Name: %-63s ║", ship.Name)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Type: %-63s ║", ship.TypeID)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Status
	b.WriteString(titleStyle.Render("║ STATUS                                                                ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Hull: %d%%%-60s ║", ship.Hull, "")) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Shields: %d%%%-57s ║", ship.Shields, "")) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Fuel: %d%-61s ║", ship.Fuel, "")) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Cargo
	b.WriteString(titleStyle.Render("║ CARGO                                                                 ║") + "\n")
	cargoUsed := 0
	if ship.Cargo != nil {
		for _, item := range ship.Cargo {
			cargoUsed += item.Quantity
		}
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Items in hold: %d%-53s ║", cargoUsed, "")) + "\n")

	// List cargo items
	if ship.Cargo != nil && len(ship.Cargo) > 0 {
		b.WriteString(titleStyle.Render("║ Items:                                                                ║") + "\n")
		for i, item := range ship.Cargo {
			if i >= 5 { // Limit to 5 items shown
				b.WriteString(titleStyle.Render(fmt.Sprintf("║   ... and %d more%-51s ║", len(ship.Cargo)-5, "")) + "\n")
				break
			}
			itemLine := fmt.Sprintf("%-40s x%d", item.CommodityID, item.Quantity)
			b.WriteString(titleStyle.Render(fmt.Sprintf("║   %-67s ║", itemLine)) + "\n")
		}
	} else {
		b.WriteString(titleStyle.Render("║ No cargo                                                              ║") + "\n")
	}
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Equipment/Outfits
	b.WriteString(titleStyle.Render("║ EQUIPMENT                                                             ║") + "\n")
	if ship.Outfits != nil && len(ship.Outfits) > 0 {
		for i, outfit := range ship.Outfits {
			if i >= 5 { // Limit to 5 outfits shown
				b.WriteString(titleStyle.Render(fmt.Sprintf("║   ... and %d more%-51s ║", len(ship.Outfits)-5, "")) + "\n")
				break
			}
			b.WriteString(titleStyle.Render(fmt.Sprintf("║   %-67s ║", outfit)) + "\n")
		}
	} else {
		b.WriteString(titleStyle.Render("║ No outfits installed                                                  ║") + "\n")
	}
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Weapons
	b.WriteString(titleStyle.Render("║ WEAPONS                                                               ║") + "\n")
	if ship.Weapons != nil && len(ship.Weapons) > 0 {
		for i, weapon := range ship.Weapons {
			if i >= 5 { // Limit to 5 weapons shown
				b.WriteString(titleStyle.Render(fmt.Sprintf("║   ... and %d more%-51s ║", len(ship.Weapons)-5, "")) + "\n")
				break
			}
			// Check for ammo
			ammoStr := ""
			if ship.WeaponAmmo != nil {
				if ammo, ok := ship.WeaponAmmo[i]; ok {
					ammoStr = fmt.Sprintf(" [%d rounds]", ammo)
				}
			}
			weaponLine := fmt.Sprintf("%s%s", weapon, ammoStr)
			b.WriteString(titleStyle.Render(fmt.Sprintf("║   %-67s ║", weaponLine)) + "\n")
		}
	} else {
		b.WriteString(titleStyle.Render("║ No weapons installed                                                  ║") + "\n")
	}
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Crew
	b.WriteString(titleStyle.Render("║ CREW                                                                  ║") + "\n")
	crewDisplay := "Unknown"
	if ship.Crew > 0 {
		crewDisplay = fmt.Sprintf("%d crew members", ship.Crew)
	} else {
		crewDisplay = "Skeleton crew"
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ %s%-s ║", crewDisplay, strings.Repeat(" ", 69-len(crewDisplay)))) + "\n")

	return b.String()
}

// viewFleetViewEscort renders escort details
func (m *Model) viewFleetViewEscort() string {
	var b strings.Builder

	// Find selected escort
	var escort *fleet.Escort
	if m.fleet.currentFleet != nil {
		for _, e := range m.fleet.currentFleet.Escorts {
			if e.ID == m.fleet.selectedEscort {
				escort = e
				break
			}
		}
	}

	if escort == nil {
		b.WriteString(titleStyle.Render("║ Escort not found                                                      ║") + "\n")
		return b.String()
	}

	b.WriteString(titleStyle.Render("║                   ESCORT DETAILS                                      ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Pilot: %-62s ║", escort.Pilot)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Level: %-62d ║", escort.Level)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Loyalty: %d%%%-57s ║", int(escort.Loyalty*100), "")) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Behavior: %-59s ║", escort.Behavior)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Status: %-61s ║", escort.Status)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Hired: %-62s ║", escort.HiredAt.Format("2006-01-02"))) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	if escort.Ship != nil {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Ship: %-63s ║", escort.Ship.Name)) + "\n")
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Hull: %d%%%-60s ║", escort.Ship.Hull, "")) + "\n")
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Shields: %d%%%-57s ║", escort.Ship.Shields, "")) + "\n")
	}

	return b.String()
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

var (
	fleetCreditStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	fleetSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // Cyan
	resetStyle         = lipgloss.NewStyle()                                  // Reset to default
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // Cyan (alias for fleet selected)
)
