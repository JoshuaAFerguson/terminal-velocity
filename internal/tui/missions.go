// File: internal/tui/missions.go
// Project: Terminal Velocity
// Description: Missions screen - Mission board and progress tracking interface
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The missions screen provides access to the mission system:
// - Browse available missions (generated dynamically)
// - Accept missions (max 5 active missions)
// - View mission details with objectives and rewards
// - Track mission progress (delivery, combat, exploration, bounty)
// - Complete missions for credits and reputation
// - Abandon missions (with penalties)
// - Generate new missions (testing/debugging feature)
//
// Mission Types:
// - Delivery: Transport cargo to destination
// - Combat: Destroy specific targets
// - Exploration: Visit specific systems
// - Bounty: Hunt down specific ships
// - Trading: Trade specific commodities
// - Escort: Protect ships during travel
//
// Mission System:
// - Missions generated based on player level and location
// - Requirements: Combat rating, reputation with factions
// - Progress tracked automatically during gameplay
// - Rewards: Credits, reputation changes, items
// - Time limits enforced with deadlines
// - Failure penalties for expired or abandoned missions
// - Icons differentiate mission types (📦, ⚔️, 💀, 💰, 🛡️, 🔭)

package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/missions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// missionsModel contains the state for the missions screen.
// Manages mission board display, acceptance, and progress tracking.
type missionsModel struct {
	mode            string             // Current mode: "board", "active", "details"
	cursor          int                // Current cursor position in mission list
	selectedMission *models.Mission    // Mission selected for viewing
	manager         *missions.Manager  // Mission management system
	message         string             // Status or error message to display
	tab             int                // Current tab: 0 = available, 1 = active
}

// newMissionsModel creates and initializes a new missions screen model.
// Initializes the mission manager for generating and tracking missions.
func newMissionsModel() missionsModel {
	return missionsModel{
		mode:    "board",
		cursor:  0,
		manager: missions.NewManager(),
		tab:     0,
	}
}

// updateMissions handles input and state updates for the missions screen.
//
// Key Bindings (Board Mode):
//   - esc/q: Return to main menu
//   - up/k, down/j: Navigate mission list
//   - tab: Switch between available/active tabs
//   - enter: View mission details
//   - g: Generate new missions (testing feature)
//   - c: Check mission progress
//   - a: Abandon selected mission (active missions only)
//
// Key Bindings (Details Mode):
//   - esc/q: Return to mission board
//   - a: Accept mission (if available)
//   - d: Decline mission (if available)
//
// Mission Workflow:
//   1. Browse available missions on board
//   2. Select mission and view details
//   3. Accept mission (validates requirements)
//   4. Mission added to active missions (max 5)
//   5. Progress tracked automatically during gameplay
//   6. Check progress with 'c' key
//   7. Complete mission when objectives met
//   8. Receive rewards (credits, reputation, items)
//
// Message Handling:
//   - All updates happen synchronously through manager
func (m Model) updateMissions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.missions.mode {
		case "board":
			return m.updateMissionsBoard(msg)
		case "active":
			return m.updateActiveMissions(msg)
		case "details":
			return m.updateMissionDetails(msg)
		}
	}

	return m, nil
}

// updateMissionsBoard handles mission board input
func (m Model) updateMissionsBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	missions := m.missions.manager.GetAvailableMissions()

	switch msg.String() {
	case "up", "k":
		if m.missions.cursor > 0 {
			m.missions.cursor--
		}

	case "down", "j":
		if m.missions.cursor < len(missions)-1 {
			m.missions.cursor++
		}

	case "tab":
		// Switch between available and active tabs
		m.missions.tab = (m.missions.tab + 1) % 2
		m.missions.cursor = 0

	case "enter":
		// View mission details
		if len(missions) > 0 && m.missions.cursor < len(missions) {
			m.missions.selectedMission = missions[m.missions.cursor]
			m.missions.mode = "details"
		}

	case "g":
		// Generate new missions (for testing)
		if m.player.CurrentSystem != uuid.Nil {
			newMissions := m.missions.manager.GenerateMissions(nil, m.player.CurrentSystem, "test_faction", 5)
			m.missions.message = fmt.Sprintf("Generated %d new missions", len(newMissions))
		}

	case "esc", "q":
		// Return to main menu
		m.screen = ScreenMainMenu
	}

	return m, nil
}

// updateActiveMissions handles active missions input
func (m Model) updateActiveMissions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	activeMissions := m.missions.manager.GetActiveMissions()

	switch msg.String() {
	case "up", "k":
		if m.missions.cursor > 0 {
			m.missions.cursor--
		}

	case "down", "j":
		if m.missions.cursor < len(activeMissions)-1 {
			m.missions.cursor++
		}

	case "enter":
		// View mission details
		if len(activeMissions) > 0 && m.missions.cursor < len(activeMissions) {
			m.missions.selectedMission = activeMissions[m.missions.cursor]
			m.missions.mode = "details"
		}

	case "a":
		// Abandon mission
		if len(activeMissions) > 0 && m.missions.cursor < len(activeMissions) {
			mission := activeMissions[m.missions.cursor]
			err := m.missions.manager.FailMission(mission.ID, "abandoned by player", m.player)
			if err == nil {
				m.missions.message = fmt.Sprintf("Abandoned mission: %s", mission.Title)
				m.missions.cursor = 0
			} else {
				m.missions.message = fmt.Sprintf("Error: %s", err.Error())
			}
		}

	case "c":
		// Check mission progress
		progressMsgs := m.missions.manager.CheckMissionProgress(m.player, m.currentShip)
		if len(progressMsgs) > 0 {
			m.missions.message = "Mission progress checked:"
			for _, msg := range progressMsgs {
				m.missions.message += "\n" + msg
			}
		} else {
			m.missions.message = "No missions completed"
		}

	case "esc", "q":
		// Return to missions board
		m.missions.mode = "board"
	}

	return m, nil
}

// updateMissionDetails handles mission details input
func (m Model) updateMissionDetails(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "a":
		// Accept mission
		if m.missions.selectedMission != nil && m.missions.selectedMission.Status == models.MissionStatusAvailable {
			// Get ship type
			var shipType *models.ShipType
			if m.currentShip != nil {
				shipType = models.GetShipTypeByID(m.currentShip.TypeID)
			}

			err := m.missions.manager.AcceptMission(m.missions.selectedMission.ID, m.player, m.currentShip, shipType)
			if err == nil {
				m.missions.message = fmt.Sprintf("Accepted mission: %s", m.missions.selectedMission.Title)

				// Check for immediate mission progress
				progressMsgs := m.missions.manager.CheckMissionProgress(m.player, m.currentShip)
				for _, msg := range progressMsgs {
					m.missions.message += "\n" + msg
				}

				m.missions.mode = "board"
				m.missions.cursor = 0
			} else {
				m.missions.message = fmt.Sprintf("Cannot accept: %s", err.Error())
			}
		}

	case "d":
		// Decline mission
		if m.missions.selectedMission != nil && m.missions.selectedMission.Status == models.MissionStatusAvailable {
			err := m.missions.manager.DeclineMission(m.missions.selectedMission.ID)
			if err == nil {
				m.missions.message = "Mission declined"
				m.missions.mode = "board"
				m.missions.cursor = 0
			}
		}

	case "esc", "q":
		// Return to board
		m.missions.mode = "board"
		m.missions.selectedMission = nil
	}

	return m, nil
}

// viewMissions renders the missions screen.
//
// Layout (Mission Board):
//   - ASCII border box with title "MISSION BOARD"
//   - Tab indicators: [▶AVAILABLE◀] or [▶ACTIVE◀]
//   - Mission list (max 10 displayed): Type icon, title, reward
//   - Cursor indicator (▶) for selected mission
//   - Message area for status/error messages
//   - Footer: Key bindings help
//
// Layout (Mission Details):
//   - ASCII border box with title "MISSION DETAILS"
//   - Mission title
//   - Type and status indicators
//   - Description with word wrapping
//   - Objectives: Destination, target, quantity/progress
//   - Rewards: Credits, reputation changes
//   - Requirements: Combat rating, faction reputation (with ✓/✗ indicators)
//   - Deadline: Time remaining formatted
//   - Footer: Accept/decline/back options
//
// Visual Features:
//   - Mission type icons: 📦 (delivery), ⚔️ (combat), 💀 (bounty), 💰 (trading), 🛡️ (escort), 🔭 (exploration)
//   - Credits formatted: 1.5K, 2.3M for large amounts
//   - Requirements checked with ✓ (met) or ✗ (not met)
//   - Duration formatting: days, hours, minutes
//   - Word wrapping for long descriptions (68 char width)
//   - Active tab highlighted with arrows
func (m Model) viewMissions() string {
	switch m.missions.mode {
	case "board":
		return m.viewMissionsBoard()
	case "active":
		return m.viewActiveMissions()
	case "details":
		return m.viewMissionDetails()
	default:
		return "Unknown missions view"
	}
}

// viewMissionsBoard renders the mission board
func (m Model) viewMissionsBoard() string {
	var s strings.Builder

	s.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	s.WriteString("║                         MISSION BOARD                                  ║\n")
	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")

	// Tabs
	availableTab := "[ AVAILABLE ]"
	activeTab := "[ ACTIVE ]"
	if m.missions.tab == 0 {
		availableTab = "[▶AVAILABLE◀]"
	} else {
		activeTab = "[▶ACTIVE◀]"
	}

	s.WriteString(fmt.Sprintf("║ %s  %s%s║\n",
		availableTab, activeTab, strings.Repeat(" ", 44-len(availableTab)-len(activeTab))))

	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")

	// Show appropriate list
	var missionList []*models.Mission
	if m.missions.tab == 0 {
		missionList = m.missions.manager.GetAvailableMissions()
	} else {
		missionList = m.missions.manager.GetActiveMissions()
	}

	if len(missionList) == 0 {
		s.WriteString("║                      No missions available                             ║\n")
	} else {
		// Display missions (max 10)
		displayCount := len(missionList)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			mission := missionList[i]
			cursor := "  "
			if i == m.missions.cursor {
				cursor = "▶"
			}

			// Format mission line
			typeIcon := getMissionTypeIcon(mission.Type)
			title := mission.Title
			if len(title) > 35 {
				title = title[:32] + "..."
			}

			reward := formatCredits(mission.Reward)

			line := fmt.Sprintf("║ %s %s %-35s  %8s cr   ║\n", cursor, typeIcon, title, reward)
			s.WriteString(line)
		}

		// Pad remaining lines
		for i := displayCount; i < 10; i++ {
			s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")
		}
	}

	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")

	// Message area
	if m.missions.message != "" {
		msg := m.missions.message
		if len(msg) > 68 {
			msg = msg[:65] + "..."
		}
		s.WriteString(fmt.Sprintf("║ %s%s║\n", msg, strings.Repeat(" ", 70-len(msg))))
	} else {
		s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")
	}

	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")
	s.WriteString("║ [↑/↓] Nav [TAB] Switch [Enter] View [G] Gen [C] Check [Q] Quit        ║\n")
	s.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n")

	return s.String()
}

// viewActiveMissions renders active missions list
func (m Model) viewActiveMissions() string {
	// Similar to board, but show active missions with progress
	return m.viewMissionsBoard() // For now, reuse board view
}

// viewMissionDetails renders detailed mission view
func (m Model) viewMissionDetails() string {
	if m.missions.selectedMission == nil {
		return "No mission selected"
	}

	mission := m.missions.selectedMission
	var s strings.Builder

	s.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	s.WriteString("║                        MISSION DETAILS                                 ║\n")
	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")

	// Title
	s.WriteString(fmt.Sprintf("║ %s%s║\n",
		mission.Title, strings.Repeat(" ", 70-len(mission.Title))))

	s.WriteString("║" + strings.Repeat("─", 72) + "║\n")

	// Type and Status
	typeIcon := getMissionTypeIcon(mission.Type)
	statusText := getStatusText(mission.Status)
	s.WriteString(fmt.Sprintf("║ Type: %s %-20s  Status: %-20s          ║\n",
		typeIcon, mission.Type, statusText))

	s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")

	// Description (word wrap)
	descLines := wordWrap(mission.Description, 68)
	for _, line := range descLines {
		s.WriteString(fmt.Sprintf("║ %s%s║\n", line, strings.Repeat(" ", 70-len(line))))
	}

	s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")

	// Objectives
	s.WriteString("║ Objectives:                                                            ║\n")
	if mission.Destination != nil {
		s.WriteString("║   • Destination: [System/Planet ID]                                   ║\n")
	}
	if mission.Target != nil {
		s.WriteString(fmt.Sprintf("║   • Target: %-59s║\n", *mission.Target))
	}
	if mission.Quantity > 0 {
		s.WriteString(fmt.Sprintf("║   • Quantity: %d / %d%s║\n",
			mission.Progress, mission.Quantity,
			strings.Repeat(" ", 55-len(fmt.Sprintf("%d / %d", mission.Progress, mission.Quantity)))))
	}

	s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")

	// Rewards
	s.WriteString("║ Rewards:                                                               ║\n")
	s.WriteString(fmt.Sprintf("║   • Credits: %s cr%s║\n",
		formatCredits(mission.Reward),
		strings.Repeat(" ", 55-len(formatCredits(mission.Reward)))))

	if len(mission.ReputationChange) > 0 {
		for factionID, repChange := range mission.ReputationChange {
			faction := models.GetFactionByID(factionID)
			factionName := factionID
			if faction != nil {
				factionName = faction.ShortName
			}
			s.WriteString(fmt.Sprintf("║   • Reputation: %s %+d%s║\n",
				factionName, repChange,
				strings.Repeat(" ", 52-len(factionName)-len(fmt.Sprintf("%+d", repChange)))))
		}
	}

	s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")

	// Requirements
	if mission.MinCombatRating > 0 || len(mission.RequiredRep) > 0 {
		s.WriteString("║ Requirements:                                                          ║\n")
		if mission.MinCombatRating > 0 {
			meetsReq := "✓"
			if m.player.CombatRating < mission.MinCombatRating {
				meetsReq = "✗"
			}
			s.WriteString(fmt.Sprintf("║   %s Combat Rating: %d%s║\n",
				meetsReq, mission.MinCombatRating,
				strings.Repeat(" ", 55-len(fmt.Sprintf("%d", mission.MinCombatRating)))))
		}
		for factionID, requiredRep := range mission.RequiredRep {
			faction := models.GetFactionByID(factionID)
			factionName := factionID
			if faction != nil {
				factionName = faction.ShortName
			}
			playerRep := m.player.GetReputation(factionID)
			meetsReq := "✓"
			if playerRep < requiredRep {
				meetsReq = "✗"
			}
			s.WriteString(fmt.Sprintf("║   %s %s Reputation: %+d%s║\n",
				meetsReq, factionName, requiredRep,
				strings.Repeat(" ", 52-len(factionName)-len(fmt.Sprintf("%+d", requiredRep)))))
		}
		s.WriteString("║" + strings.Repeat(" ", 72) + "║\n")
	}

	// Deadline
	timeRemaining := time.Until(mission.Deadline)
	deadlineText := formatDuration(timeRemaining)
	s.WriteString(fmt.Sprintf("║ Deadline: %s%s║\n",
		deadlineText, strings.Repeat(" ", 61-len(deadlineText))))

	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")

	// Actions
	if mission.Status == models.MissionStatusAvailable {
		s.WriteString("║ [A] Accept Mission  [D] Decline  [Q] Back                              ║\n")
	} else if mission.Status == models.MissionStatusActive {
		s.WriteString("║ [Q] Back                                                               ║\n")
	} else {
		s.WriteString("║ [Q] Back                                                               ║\n")
	}

	s.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n")

	return s.String()
}

// Helper functions

func getMissionTypeIcon(missionType string) string {
	switch missionType {
	case models.MissionTypeDelivery:
		return "📦"
	case models.MissionTypeCombat:
		return "⚔️"
	case models.MissionTypeBounty:
		return "💀"
	case models.MissionTypeTrading:
		return "💰"
	case models.MissionTypeEscort:
		return "🛡️"
	case models.MissionTypeExploration:
		return "🔭"
	default:
		return "❓"
	}
}

func getStatusText(status string) string {
	switch status {
	case models.MissionStatusAvailable:
		return "Available"
	case models.MissionStatusActive:
		return "Active"
	case models.MissionStatusCompleted:
		return "Completed"
	case models.MissionStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

func wordWrap(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}

func formatCredits(amount int64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(amount)/1000000.0)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1fK", float64(amount)/1000.0)
	}
	return strconv.FormatInt(amount, 10)
}
