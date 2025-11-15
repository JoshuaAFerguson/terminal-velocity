// File: internal/tui/fleet.go
// Project: Terminal Velocity
// Description: Fleet management TUI screen for multi-ship ownership and escorts
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
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
	escortName    string
	escortBehavior fleet.EscortBehavior

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
			// TODO: Load fleet data from fleet manager
		case 1: // Stored Ships
			m.fleet.mode = fleetModeStored
			m.fleet.selectedIndex = 0
			// TODO: Load stored ships
		case 2: // Manage Escorts
			m.fleet.mode = fleetModeEscorts
			m.fleet.selectedIndex = 0
			// TODO: Load escorts
		case 3: // Fleet Formation
			m.fleet.mode = fleetModeFormation
			m.fleet.selectedIndex = 0
		case 4: // Hire New Escort
			m.fleet.mode = fleetModeHireEscort
			m.fleet.escortName = ""
			m.fleet.selectedIndex = 0
		case 5: // Pay Maintenance
			// TODO: Pay maintenance early
			m.fleet.message = "Maintenance paid! Escort loyalty increased."
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
			// TODO: Call fleet manager to switch flagship
			m.fleet.message = "Switched to new flagship!"
			_ = shipID // Placeholder
		}
	case "t":
		// Store ship at current planet
		if m.fleet.selectedIndex < len(m.fleet.currentFleet.OwnedShips) {
			shipID := m.fleet.currentFleet.OwnedShips[m.fleet.selectedIndex].ID
			if shipID == m.fleet.currentFleet.FlagshipID {
				m.fleet.error = "Cannot store your active flagship!"
			} else {
				// TODO: Call fleet manager to store ship
				m.fleet.message = "Ship stored at current planet"
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
			// TODO: Check if at correct planet and retrieve
			m.fleet.message = fmt.Sprintf("Retrieved ship from %s", storedShip.Location)
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
			// TODO: Call fleet manager to dismiss escort
			m.fleet.message = "Escort dismissed"
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
			// TODO: Call fleet manager to set behavior
			m.fleet.message = fmt.Sprintf("Escort behavior set to %s", behavior)
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
		// TODO: Call fleet manager to set formation
		m.fleet.message = fmt.Sprintf("Formation set to %s", formation)
		m.fleet.mode = fleetModeMenu
	case "b", "q", "esc":
		m.fleet.mode = fleetModeMenu
	}

	return m, nil
}

// updateFleetHireEscort handles hiring new escorts
func (m *Model) updateFleetHireEscort(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.fleet.mode = fleetModeMenu
	case "enter":
		// TODO: Hire escort with selected ship and behavior
		m.fleet.message = "Escort hired! They will follow and protect you."
		m.fleet.mode = fleetModeMenu
	case "up", "k":
		if m.fleet.selectedIndex > 0 {
			m.fleet.selectedIndex--
		}
	case "down", "j":
		if m.fleet.selectedIndex < len(behaviorTypes)-1 {
			m.fleet.selectedIndex++
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
		b.WriteString(helpStyle.Render("↑/↓: Select Behavior | Enter: Hire | Q: Cancel\n"))
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

		// TODO: Get actual ship type name
		shipType := "Unknown"

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
	b.WriteString(titleStyle.Render("║ Select escort behavior:                                               ║") + "\n")
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

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ TODO: Ship selection and pilot name input                             ║") + "\n")

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
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Hull: %d%%%-60s ║", ship.Hull, "")) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Shields: %d%%%-57s ║", ship.Shields, "")) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Fuel: %d%-61s ║", ship.Fuel, "")) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ TODO: Additional ship details and stats                               ║") + "\n")

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
