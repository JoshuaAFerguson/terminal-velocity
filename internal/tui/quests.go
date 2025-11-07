// File: internal/tui/quests.go
// Project: Terminal Velocity
// Description: Quest and storyline UI
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

// Quest view modes
const (
	questViewActive    = "active"
	questViewAvailable = "available"
	questViewCompleted = "completed"
	questViewDetail    = "detail"
)

type questsModel struct {
	viewMode       string
	cursor         int
	activeQuests   []*models.PlayerQuest
	availableQuests []*models.Quest
	completedQuests []*models.PlayerQuest
	selectedQuest  *models.Quest
	selectedPlayer *models.PlayerQuest
}

func newQuestsModel() questsModel {
	return questsModel{
		viewMode:        questViewActive,
		cursor:          0,
		activeQuests:    []*models.PlayerQuest{},
		availableQuests: []*models.Quest{},
		completedQuests: []*models.PlayerQuest{},
	}
}

func (m Model) updateQuests(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.questsModel.viewMode == questViewDetail {
				m.questsModel.viewMode = questViewActive
				m.questsModel.selectedQuest = nil
				m.questsModel.selectedPlayer = nil
			} else {
				m.screen = ScreenMainMenu
			}
			return m, nil

		case "tab":
			// Cycle through view modes
			switch m.questsModel.viewMode {
			case questViewActive:
				m.questsModel.viewMode = questViewAvailable
			case questViewAvailable:
				m.questsModel.viewMode = questViewCompleted
			case questViewCompleted:
				m.questsModel.viewMode = questViewActive
			}
			m.questsModel.cursor = 0
			return m, nil

		case "up", "k":
			if m.questsModel.cursor > 0 {
				m.questsModel.cursor--
			}
			return m, nil

		case "down", "j":
			maxCursor := m.getQuestsMaxCursor()
			if m.questsModel.cursor < maxCursor {
				m.questsModel.cursor++
			}
			return m, nil

		case "enter", " ":
			return m.handleQuestSelect()

		case "a":
			// Abandon quest (if viewing active quest detail)
			if m.questsModel.viewMode == questViewDetail && m.questsModel.selectedPlayer != nil {
				if m.questManager != nil {
					m.questManager.AbandonQuest(m.playerID, m.questsModel.selectedPlayer.QuestID)
					m.questsModel.viewMode = questViewActive
					m.questsModel.selectedQuest = nil
					m.questsModel.selectedPlayer = nil
					// Reload quests
					m.questsModel.activeQuests = m.questManager.GetActiveQuests(m.playerID)
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) getQuestsMaxCursor() int {
	switch m.questsModel.viewMode {
	case questViewActive:
		return len(m.questsModel.activeQuests) - 1
	case questViewAvailable:
		return len(m.questsModel.availableQuests) - 1
	case questViewCompleted:
		return len(m.questsModel.completedQuests) - 1
	case questViewDetail:
		return 0
	default:
		return 0
	}
}

func (m Model) handleQuestSelect() (tea.Model, tea.Cmd) {
	if m.questManager == nil {
		return m, nil
	}

	switch m.questsModel.viewMode {
	case questViewActive:
		if m.questsModel.cursor < len(m.questsModel.activeQuests) {
			pq := m.questsModel.activeQuests[m.questsModel.cursor]
			m.questsModel.selectedQuest = m.questManager.GetQuest(pq.QuestID)
			m.questsModel.selectedPlayer = pq
			m.questsModel.viewMode = questViewDetail
		}

	case questViewAvailable:
		if m.questsModel.cursor < len(m.questsModel.availableQuests) {
			quest := m.questsModel.availableQuests[m.questsModel.cursor]
			// Start quest
			m.questManager.StartQuest(m.playerID, quest.ID)
			// Reload quests
			m.questsModel.activeQuests = m.questManager.GetActiveQuests(m.playerID)
			m.questsModel.availableQuests = m.questManager.GetAvailableQuests(m.playerID)
			m.questsModel.viewMode = questViewActive
			m.questsModel.cursor = 0
		}

	case questViewCompleted:
		if m.questsModel.cursor < len(m.questsModel.completedQuests) {
			pq := m.questsModel.completedQuests[m.questsModel.cursor]
			m.questsModel.selectedQuest = m.questManager.GetQuest(pq.QuestID)
			m.questsModel.selectedPlayer = pq
			m.questsModel.viewMode = questViewDetail
		}
	}

	return m, nil
}

func (m Model) viewQuests() string {
	if m.questManager == nil {
		return errorView("Quest system not initialized")
	}

	s := renderHeader(m.username, m.player.Credits, "Quests")
	s += "\n"

	s += subtitleStyle.Render("=== Quest Journal ===") + "\n"

	// Tab navigation
	tabs := fmt.Sprintf("[%s] Active  [%s] Available  [%s] Completed",
		m.getTabIndicator(questViewActive),
		m.getTabIndicator(questViewAvailable),
		m.getTabIndicator(questViewCompleted),
	)
	s += helpStyle.Render(tabs) + "\n\n"

	switch m.questsModel.viewMode {
	case questViewActive:
		s += m.viewActiveQuests()
	case questViewAvailable:
		s += m.viewAvailableQuests()
	case questViewCompleted:
		s += m.viewCompletedQuests()
	case questViewDetail:
		s += m.viewQuestDetail()
	}

	return s
}

func (m Model) getTabIndicator(mode string) string {
	if m.questsModel.viewMode == mode {
		return "•"
	}
	return " "
}

func (m Model) viewActiveQuests() string {
	if len(m.questsModel.activeQuests) == 0 {
		s := helpStyle.Render("No active quests") + "\n\n"
		s += "Check the 'Available' tab to find new quests!\n\n"
		s += renderFooter("Tab: Switch View  •  ESC: Back")
		return s
	}

	s := fmt.Sprintf("Active Quests (%d):\n\n", len(m.questsModel.activeQuests))

	for i, pq := range m.questsModel.activeQuests {
		quest := m.questManager.GetQuest(pq.QuestID)
		if quest == nil {
			continue
		}

		progress := pq.GetProgress(quest)
		typeIcon := m.getQuestTypeIcon(quest.Type)

		line := fmt.Sprintf("%s %s (%.0f%% complete)", typeIcon, quest.Title, progress*100)

		if i == m.questsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: View Details  •  Tab: Switch View  •  ESC: Back")
	return s
}

func (m Model) viewAvailableQuests() string {
	if len(m.questsModel.availableQuests) == 0 {
		s := helpStyle.Render("No available quests at this time") + "\n\n"
		s += "Complete active quests or explore new systems to unlock more!\n\n"
		s += renderFooter("Tab: Switch View  •  ESC: Back")
		return s
	}

	s := fmt.Sprintf("Available Quests (%d):\n\n", len(m.questsModel.availableQuests))

	for i, quest := range m.questsModel.availableQuests {
		typeIcon := m.getQuestTypeIcon(quest.Type)
		line := fmt.Sprintf("%s %s (Level %d)", typeIcon, quest.Title, quest.Level)

		if i == m.questsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
			s += "  " + helpStyle.Render(quest.Description) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Accept Quest  •  Tab: Switch View  •  ESC: Back")
	return s
}

func (m Model) viewCompletedQuests() string {
	if len(m.questsModel.completedQuests) == 0 {
		s := helpStyle.Render("No completed quests yet") + "\n\n"
		s += "Complete quests to see your achievements here!\n\n"
		s += renderFooter("Tab: Switch View  •  ESC: Back")
		return s
	}

	s := fmt.Sprintf("Completed Quests (%d):\n\n", len(m.questsModel.completedQuests))

	for i, pq := range m.questsModel.completedQuests {
		quest := m.questManager.GetQuest(pq.QuestID)
		if quest == nil {
			continue
		}

		typeIcon := m.getQuestTypeIcon(quest.Type)
		line := fmt.Sprintf("✓ %s %s", typeIcon, quest.Title)

		if i == m.questsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + successStyle.Render(line) + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: View Details  •  Tab: Switch View  •  ESC: Back")
	return s
}

func (m Model) viewQuestDetail() string {
	if m.questsModel.selectedQuest == nil {
		return helpStyle.Render("No quest selected") + "\n"
	}

	quest := m.questsModel.selectedQuest
	pq := m.questsModel.selectedPlayer

	s := highlightStyle.Render(quest.Title) + "\n"
	s += helpStyle.Render(fmt.Sprintf("Type: %s  •  Level: %d", quest.Type, quest.Level)) + "\n\n"

	s += quest.Description + "\n\n"

	if quest.Giver != "" {
		s += fmt.Sprintf("Quest Giver: %s\n\n", statsStyle.Render(quest.Giver))
	}

	s += "Objectives:\n"
	for _, obj := range quest.Objectives {
		status := "○"
		progress := ""

		if pq != nil {
			if pq.IsObjectiveComplete(obj.ID) {
				status = "✓"
			} else {
				current := pq.Objectives[obj.ID]
				progress = fmt.Sprintf(" (%d/%d)", current, obj.Required)
			}
		}

		optional := ""
		if obj.Optional {
			optional = " (Optional)"
		}

		s += fmt.Sprintf("  %s %s%s%s\n", status, obj.Description, progress, optional)
	}

	s += "\n" + "Rewards:\n"
	if quest.Rewards.Credits > 0 {
		s += fmt.Sprintf("  • Credits: %s\n", statsStyle.Render(fmt.Sprintf("%d cr", quest.Rewards.Credits)))
	}
	if quest.Rewards.Experience > 0 {
		s += fmt.Sprintf("  • Experience: %s\n", statsStyle.Render(fmt.Sprintf("%d XP", quest.Rewards.Experience)))
	}
	if len(quest.Rewards.Items) > 0 {
		s += "  • Items:\n"
		for itemID, qty := range quest.Rewards.Items {
			s += fmt.Sprintf("    - %s x%d\n", itemID, qty)
		}
	}
	if quest.Rewards.Special != "" {
		s += fmt.Sprintf("  • %s\n", highlightStyle.Render(quest.Rewards.Special))
	}

	s += "\n"
	if pq != nil && pq.Status == models.QuestStatusActive {
		progress := pq.GetProgress(quest)
		s += fmt.Sprintf("Progress: %.0f%%\n\n", progress*100)
		s += renderFooter("A: Abandon Quest  •  ESC: Back")
	} else {
		s += renderFooter("ESC: Back")
	}

	return s
}

func (m Model) getQuestTypeIcon(questType models.QuestType) string {
	switch questType {
	case models.QuestTypeMain:
		return "★"
	case models.QuestTypeSide:
		return "○"
	case models.QuestTypeFaction:
		return "⚑"
	case models.QuestTypeDaily:
		return "◎"
	case models.QuestTypeChain:
		return "⚬"
	case models.QuestTypeHidden:
		return "◆"
	case models.QuestTypeEvent:
		return "◈"
	default:
		return "•"
	}
}
