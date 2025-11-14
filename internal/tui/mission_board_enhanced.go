// File: internal/tui/mission_board_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced mission board screen with mission listings
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type missionBoardEnhancedModel struct {
	selectedMission int
	missions        []missionListing
	mode            string // "browse", "confirm"
}

type missionListing struct {
	title        string
	missionType  string // "DELIVERY", "BOUNTY", "ESCORT", "COMBAT"
	reward       int64
	deadline     string // "3 days", "None", etc.
	difficulty   int    // 0-10 for progress bar
	cargo        int    // tons required
	ships        int    // for escort missions
	description  string
	employer     string
	destination  string
	timeLimit    string
	reputation   string
	isUrgent     bool
}

func newMissionBoardEnhancedModel() missionBoardEnhancedModel {
	// Sample missions
	missions := []missionListing{
		{
			title: "Rush Shipment to Mars", missionType: "DELIVERY",
			reward: 8500, deadline: "3 days", difficulty: 3, cargo: 15,
			description: "A shipment of industrial components needs to reach Mars Colony before the next construction cycle begins. Time is of the essence! Deliver 15 tons of components within 3 days.",
			employer:    "Mars Construction Guild",
			destination: "Mars - Olympus Mons Spaceport",
			timeLimit:   "3 days (72 hours)",
			reputation:  "None required",
			isUrgent:    false,
		},
		{
			title: "Eliminate Pirate Lord Zaxon", missionType: "BOUNTY",
			reward: 45000, deadline: "None", difficulty: 8,
			description: "The notorious pirate lord Zaxon has been terrorizing trade routes in the outer systems. Eliminate him and collect the bounty. Extreme danger - recommended for experienced pilots only.",
			employer:    "United Earth Navy",
			destination: "Outer Systems - Various",
			timeLimit:   "No deadline",
			reputation:  "Combat Rating: 60+",
			isUrgent:    false,
		},
		{
			title: "Protect Convoy to Alpha Centauri", missionType: "ESCORT",
			reward: 22000, deadline: "7 days", difficulty: 6, ships: 3,
			description: "Escort a convoy of 3 trade ships through pirate-infested space to Alpha Centauri. Protect the convoy from all threats. Bonus payment if all ships arrive intact.",
			employer:    "Merchant Guild",
			destination: "Alpha Centauri - Proxima Station",
			timeLimit:   "7 days",
			reputation:  "Combat Rating: 40+",
			isUrgent:    false,
		},
		{
			title: "Medical Supplies Needed", missionType: "DELIVERY",
			reward: 12000, deadline: "2 days", difficulty: 4, cargo: 8,
			description: "Critical medical supplies are urgently needed at a remote colony suffering from an outbreak. Fast delivery required. Lives depend on your speed.",
			employer:    "Colonial Medical Corps",
			destination: "Epsilon Eridani - Colony 7",
			timeLimit:   "2 days (48 hours)",
			reputation:  "None required",
			isUrgent:    true,
		},
		{
			title: "Clear Pirate Nest", missionType: "COMBAT",
			reward: 35000, deadline: "None", difficulty: 6,
			description: "A pirate base has been discovered in an asteroid field. Eliminate all hostile forces and destroy the base. Multiple enemy ships expected. Salvage rights included.",
			employer:    "System Defense Force",
			destination: "Asteroid Belt - Sector 7G",
			timeLimit:   "No deadline",
			reputation:  "Combat Rating: 50+",
			isUrgent:    false,
		},
	}

	return missionBoardEnhancedModel{
		selectedMission: 0,
		missions:        missions,
		mode:            "browse",
	}
}

func (m Model) viewMissionBoardEnhanced() string {
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
	header := DrawHeader("MISSION BBS - Earth Station", "", credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Initialize if needed
	if len(m.missionBoardEnhanced.missions) == 0 {
		m.missionBoardEnhanced = newMissionBoardEnhancedModel()
	}

	// Mission list panel
	listWidth := width - 4
	var listContent strings.Builder
	listContent.WriteString(" AVAILABLE MISSIONS                                                   \n")
	listContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	listContent.WriteString("                                                                      \n")

	for i, mission := range m.missionBoardEnhanced.missions {
		prefix := "   "
		if i == m.missionBoardEnhanced.selectedMission {
			prefix = " " + IconArrow + " "
		}

		urgentTag := ""
		if mission.isUrgent {
			urgentTag = "  [URGENT]"
		}

		// Mission title line
		titleLine := fmt.Sprintf("%s[%s] %s%s",
			prefix, mission.missionType, mission.title, urgentTag)
		listContent.WriteString(PadRight(titleLine, listWidth-2) + "\n")

		// Details line
		var detailsLine string
		if mission.missionType == "DELIVERY" {
			detailsLine = fmt.Sprintf("   Reward: %s   Deadline: %s   Cargo: %d tons",
				FormatCredits(mission.reward), mission.deadline, mission.cargo)
		} else if mission.missionType == "ESCORT" {
			detailsLine = fmt.Sprintf("   Reward: %s   Deadline: %s   Ships: %d",
				FormatCredits(mission.reward), mission.deadline, mission.ships)
		} else if mission.missionType == "BOUNTY" || mission.missionType == "COMBAT" {
			diffBar := DrawProgressBar(mission.difficulty, 10, 8)
			detailsLine = fmt.Sprintf("   Reward: %s   Deadline: %s   Difficulty: %s",
				FormatCredits(mission.reward), mission.deadline, diffBar)
		}
		listContent.WriteString(PadRight(detailsLine, listWidth-2) + "\n")
		listContent.WriteString("                                                                      \n")
	}

	missionList := DrawPanel("", listContent.String(), listWidth, 18, false)
	missionLines := strings.Split(missionList, "\n")
	for _, line := range missionLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Mission details panel
	detailsWidth := width - 4
	var detailsContent strings.Builder

	if m.missionBoardEnhanced.selectedMission < len(m.missionBoardEnhanced.missions) {
		mission := m.missionBoardEnhanced.missions[m.missionBoardEnhanced.selectedMission]

		detailsContent.WriteString(fmt.Sprintf(" MISSION DETAILS: %-51s\n", mission.title))
		detailsContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		detailsContent.WriteString("                                                                      \n")

		// Wrap description to fit width
		descWords := strings.Fields(mission.description)
		var currentLine string
		for _, word := range descWords {
			if len(currentLine)+len(word)+1 > 67 {
				detailsContent.WriteString(fmt.Sprintf("  %-68s\n", currentLine))
				currentLine = word
			} else {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			}
		}
		if currentLine != "" {
			detailsContent.WriteString(fmt.Sprintf("  %-68s\n", currentLine))
		}
		detailsContent.WriteString("                                                                      \n")

		detailsContent.WriteString(fmt.Sprintf("  Employer: %-61s\n", mission.employer))
		detailsContent.WriteString(fmt.Sprintf("  Destination: %-58s\n", mission.destination))
		detailsContent.WriteString(fmt.Sprintf("  Payment: %-60s\n", FormatCredits(mission.reward)))
		if mission.cargo > 0 {
			detailsContent.WriteString(fmt.Sprintf("  Cargo Space Required: %d tons%-41s\n", mission.cargo, ""))
		}
		detailsContent.WriteString(fmt.Sprintf("  Time Limit: %-57s\n", mission.timeLimit))
		detailsContent.WriteString(fmt.Sprintf("  Reputation: %-57s\n", mission.reputation))
		detailsContent.WriteString("                                                                      \n")
		detailsContent.WriteString("  [ Accept Mission ]  [ Decline ]                                     \n")
		detailsContent.WriteString("                                                                      \n")
	}

	details := DrawPanel("", detailsContent.String(), detailsWidth, 15, false)
	detailLines := strings.Split(details, "\n")
	for _, line := range detailLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[↑↓] Select  [Enter] View Details  [A]ccept  [D]ecline  [ESC] Back", width)
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) updateMissionBoardEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.missionBoardEnhanced.selectedMission > 0 {
				m.missionBoardEnhanced.selectedMission--
			}
			return m, nil

		case "down", "j":
			if m.missionBoardEnhanced.selectedMission < len(m.missionBoardEnhanced.missions)-1 {
				m.missionBoardEnhanced.selectedMission++
			}
			return m, nil

		case "a", "A":
			// Accept mission
			// TODO: Implement mission acceptance via API
			return m, nil

		case "d", "D":
			// Decline mission
			// Just return to browse mode
			return m, nil

		case "esc":
			// Back to landing
			m.screen = ScreenLanding
			return m, nil
		}
	}

	return m, nil
}

// Add ScreenMissionBoardEnhanced constant to Screen enum when integrating
