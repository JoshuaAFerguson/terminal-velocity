// File: internal/tui/quest_board_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced quest board screen with progress tracking and storylines
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type questBoardEnhancedModel struct {
	selectedQuest int
	activeQuests  []questEntry
	availableQuests []questEntry
	viewMode      string // "active", "details", "available"
}

type questEntry struct {
	id          string
	title       string
	questType   string // "MAIN", "SIDE", "FACTION", "EXPLORATION", "COMBAT"
	chapter     string // For main quests
	progress    int    // 0-100
	description string
	objectives  []questObjective
	rewards     []string
	nextHint    string // Short hint for next objective
	isActive    bool
}

type questObjective struct {
	description string
	completed   bool
	inProgress  bool
}

func newQuestBoardEnhancedModel() questBoardEnhancedModel {
	// Sample active quests
	activeQuests := []questEntry{
		{
			id:        "quest_pirate_menace",
			title:     "The Pirate Menace",
			questType: "MAIN",
			chapter:   "Chapter 1",
			progress:  40,
			description: "A mysterious increase in pirate activity has been reported across human space. United Earth Intelligence suspects a larger organization is coordinating these attacks. Your mission is to investigate and neutralize the threat.",
			objectives: []questObjective{
				{description: "Speak with Admiral Chen at Earth Station", completed: true},
				{description: "Eliminate 5 pirate ships in Sol system", completed: true},
				{description: "Investigate pirate base in Sirius system", inProgress: true},
				{description: "Recover pirate communications logs", completed: false},
				{description: "Return to Admiral Chen with findings", completed: false},
			},
			rewards: []string{
				"50,000 credits",
				"Reputation: +20 United Earth",
				"Unlock: Advanced Weapons Access",
				"Unlock: Chapter 2 - \"The Shadow Syndicate\"",
			},
			nextHint: "Investigate Sirius system",
			isActive: true,
		},
		{
			id:        "quest_traders_gambit",
			title:     "Trader's Gambit",
			questType: "SIDE",
			chapter:   "",
			progress:  60,
			description: "A merchant guild contact has offered you a lucrative trading opportunity. Deliver specialized goods to Mars Colony before the deadline to earn a substantial profit and guild reputation.",
			objectives: []questObjective{
				{description: "Accept contract from Merchant Guild", completed: true},
				{description: "Purchase 20 tons of Electronics", completed: true},
				{description: "Deliver goods to Mars Colony", inProgress: true},
				{description: "Collect payment from guild representative", completed: false},
			},
			rewards: []string{
				"25,000 credits",
				"Reputation: +15 Merchant Guild",
				"Unlock: Premium Trading Routes",
			},
			nextHint: "Deliver goods to Mars",
			isActive: true,
		},
	}

	// Sample available quests
	availableQuests := []questEntry{
		{
			id:        "quest_lost_cargo",
			title:     "Lost Cargo",
			questType: "SIDE",
			description: "A trader's ship was destroyed by pirates. Recover the cargo from the wreckage.",
			isActive: false,
		},
		{
			id:        "quest_pirate_hunters",
			title:     "Pirate Hunters United",
			questType: "FACTION",
			description: "Join the Pirate Hunters faction in their campaign against organized crime.",
			isActive: false,
		},
		{
			id:        "quest_outer_reaches",
			title:     "The Outer Reaches",
			questType: "EXPLORATION",
			description: "Explore uncharted systems beyond known space and report your findings.",
			isActive: false,
		},
	}

	return questBoardEnhancedModel{
		selectedQuest:   0,
		activeQuests:    activeQuests,
		availableQuests: availableQuests,
		viewMode:        "active",
	}
}

func (m Model) viewQuestBoardEnhanced() string {
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
	header := DrawHeader("QUEST TERMINAL - Earth Station", "", credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Initialize if needed
	if len(m.questBoardEnhanced.activeQuests) == 0 {
		m.questBoardEnhanced = newQuestBoardEnhancedModel()
	}

	// Active quests panel
	panelWidth := width - 4
	var activeContent strings.Builder
	activeContent.WriteString(fmt.Sprintf(" ACTIVE QUESTS                                          [%d/5 Active]   \n",
		len(m.questBoardEnhanced.activeQuests)))
	activeContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	activeContent.WriteString("                                                                      \n")

	for i, quest := range m.questBoardEnhanced.activeQuests {
		prefix := "   "
		if i == m.questBoardEnhanced.selectedQuest {
			prefix = " ▶ "
		}

		// Quest title line
		titleLine := fmt.Sprintf("%s[%s] %-43s %s",
			prefix, quest.questType, quest.title, quest.chapter)
		activeContent.WriteString(PadRight(titleLine, panelWidth-2) + "\n")

		// Progress line
		progressBar := DrawProgressBar(quest.progress, 100, 10)
		progressLine := fmt.Sprintf("   Progress: %s %d%%   Next: %s",
			progressBar, quest.progress, quest.nextHint)
		activeContent.WriteString(PadRight(progressLine, panelWidth-2) + "\n")
		activeContent.WriteString("                                                                      \n")
	}

	activePanel := DrawPanel("", activeContent.String(), panelWidth, 10, false)
	activeLines := strings.Split(activePanel, "\n")
	for _, line := range activeLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Quest details panel - show selected quest
	var detailsContent strings.Builder
	if m.questBoardEnhanced.selectedQuest < len(m.questBoardEnhanced.activeQuests) {
		quest := m.questBoardEnhanced.activeQuests[m.questBoardEnhanced.selectedQuest]

		// Quest header
		questHeader := fmt.Sprintf(" QUEST: %s", quest.title)
		if quest.questType == "MAIN" && quest.chapter != "" {
			questHeader += fmt.Sprintf(" (%s Quest - %s)", quest.questType, quest.chapter)
		} else if quest.questType != "MAIN" {
			questHeader += fmt.Sprintf(" (%s Quest)", quest.questType)
		}
		detailsContent.WriteString(PadRight(questHeader, panelWidth-2) + "\n")
		detailsContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		detailsContent.WriteString("                                                                      \n")

		// Description - word wrap
		descWords := strings.Fields(quest.description)
		var currentLine string
		for _, word := range descWords {
			if len(currentLine)+len(word)+1 > 68 {
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

		// Objectives
		detailsContent.WriteString("  OBJECTIVES:                                                         \n")
		for _, obj := range quest.objectives {
			var marker string
			var suffix string
			if obj.completed {
				marker = "✓"
			} else if obj.inProgress {
				marker = "▪"
				suffix = "           [IN PROGRESS]"
			} else {
				marker = "▪"
			}

			objLine := fmt.Sprintf("  %s %s%s", marker, obj.description, suffix)
			detailsContent.WriteString(PadRight(objLine, panelWidth-2) + "\n")
		}
		detailsContent.WriteString("                                                                      \n")

		// Rewards
		detailsContent.WriteString("  REWARDS:                                                            \n")
		for _, reward := range quest.rewards {
			rewardLine := fmt.Sprintf("  • %s", reward)
			detailsContent.WriteString(PadRight(rewardLine, panelWidth-2) + "\n")
		}
		detailsContent.WriteString("                                                                      \n")
	}

	detailsPanel := DrawPanel("", detailsContent.String(), panelWidth, 18, false)
	detailsLines := strings.Split(detailsPanel, "\n")
	for _, line := range detailsLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Available quests panel
	var availableContent strings.Builder
	availableContent.WriteString(fmt.Sprintf(" AVAILABLE QUESTS (%d)                                                 \n",
		len(m.questBoardEnhanced.availableQuests)))

	var availableLine string
	for i, quest := range m.questBoardEnhanced.availableQuests {
		questText := fmt.Sprintf("[%s] %s", quest.questType, quest.title)
		if i > 0 && i%2 == 0 {
			availableContent.WriteString(PadRight(" "+availableLine, panelWidth-2) + "\n")
			availableLine = questText
		} else {
			if availableLine != "" {
				availableLine += "    "
			}
			availableLine += questText
		}
	}
	if availableLine != "" {
		availableContent.WriteString(PadRight(" "+availableLine, panelWidth-2) + "\n")
	}

	availablePanel := DrawPanel("", availableContent.String(), panelWidth, 4, false)
	availableLines := strings.Split(availablePanel, "\n")
	for _, line := range availableLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[↑↓] Select Quest  [Enter] Details  [A]bandon Quest  [ESC] Back", width)
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) updateQuestBoardEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.questBoardEnhanced.selectedQuest > 0 {
				m.questBoardEnhanced.selectedQuest--
			}
			return m, nil

		case "down", "j":
			if m.questBoardEnhanced.selectedQuest < len(m.questBoardEnhanced.activeQuests)-1 {
				m.questBoardEnhanced.selectedQuest++
			}
			return m, nil

		case "enter":
			// Show full quest details
			// TODO: Implement detailed quest view or accept quest
			return m, nil

		case "a", "A":
			// Abandon quest
			// TODO: Implement quest abandonment via API
			if m.questBoardEnhanced.selectedQuest < len(m.questBoardEnhanced.activeQuests) {
				// Would call API to abandon quest
				// quest := m.questBoardEnhanced.activeQuests[m.questBoardEnhanced.selectedQuest]
			}
			return m, nil

		case "esc":
			// Back to landing or main menu
			m.screen = ScreenLanding
			return m, nil
		}
	}

	return m, nil
}

// Add ScreenQuestBoardEnhanced constant to Screen enum when integrating
