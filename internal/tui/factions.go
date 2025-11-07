// File: internal/tui/factions.go
// Project: Terminal Velocity
// Description: Faction management UI
// Version: 1.0.0

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type factionsModel struct {
	viewMode    string // "list", "my_faction", "create"
	cursor      int
	createName  string
	createTag   string
	createAlign string
	inputField  int // 0=name, 1=tag, 2=alignment
}

func newFactionsModel() factionsModel {
	return factionsModel{
		viewMode:    "list",
		cursor:      0,
		createAlign: models.AlignmentTrader,
	}
}

func (m Model) updateFactions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.factionsModel.viewMode == "create" {
			return m.updateFactionsCreate(msg)
		}

		switch msg.String() {
		case "esc", "backspace", "q":
			if m.factionsModel.viewMode == "my_faction" {
				m.factionsModel.viewMode = "list"
				return m, nil
			}
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.factionsModel.cursor > 0 {
				m.factionsModel.cursor--
			}

		case "down", "j":
			factions := m.factionManager.GetAllFactions()
			if m.factionsModel.cursor < len(factions)-1 {
				m.factionsModel.cursor++
			}

		case "c":
			// Create new faction
			m.factionsModel.viewMode = "create"
			m.factionsModel.createName = ""
			m.factionsModel.createTag = ""
			m.factionsModel.inputField = 0

		case "v":
			// View my faction
			m.factionsModel.viewMode = "my_faction"
		}
	}

	return m, nil
}

func (m Model) updateFactionsCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.factionsModel.viewMode = "list"
		return m, nil

	case "tab":
		m.factionsModel.inputField = (m.factionsModel.inputField + 1) % 3

	case "enter":
		if m.factionsModel.inputField == 2 && len(m.factionsModel.createName) > 0 && len(m.factionsModel.createTag) > 0 {
			// Create faction
			_, err := m.factionManager.CreateFaction(
				m.factionsModel.createName,
				m.factionsModel.createTag,
				m.playerID,
				m.factionsModel.createAlign,
			)
			if err == nil {
				m.factionsModel.viewMode = "my_faction"
			}
		}

	case "backspace":
		if m.factionsModel.inputField == 0 && len(m.factionsModel.createName) > 0 {
			m.factionsModel.createName = m.factionsModel.createName[:len(m.factionsModel.createName)-1]
		} else if m.factionsModel.inputField == 1 && len(m.factionsModel.createTag) > 0 {
			m.factionsModel.createTag = m.factionsModel.createTag[:len(m.factionsModel.createTag)-1]
		}

	default:
		if len(msg.String()) == 1 {
			if m.factionsModel.inputField == 0 && len(m.factionsModel.createName) < 30 {
				m.factionsModel.createName += msg.String()
			} else if m.factionsModel.inputField == 1 && len(m.factionsModel.createTag) < 5 {
				m.factionsModel.createTag += strings.ToUpper(msg.String())
			}
		}
	}

	return m, nil
}

func (m Model) viewFactions() string {
	if m.factionsModel.viewMode == "create" {
		return m.viewFactionsCreate()
	}

	if m.factionsModel.viewMode == "my_faction" {
		return m.viewMyFaction()
	}

	s := titleStyle.Render("🏛️  FACTIONS") + "\n\n"

	// Stats
	stats := m.factionManager.GetStats()
	s += fmt.Sprintf("Total Factions: %d | Total Members: %d | Recruiting: %d\n",
		stats.TotalFactions, stats.TotalMembers, stats.RecruitingFactions)
	s += strings.Repeat("─", 80) + "\n\n"

	// Player's faction status
	myFaction, err := m.factionManager.GetPlayerFaction(m.playerID)
	if err == nil {
		s += fmt.Sprintf("Your Faction: %s\n\n", myFaction.GetFullName())
	} else {
		s += "You are not in a faction\n\n"
	}

	// Faction list
	factions := m.factionManager.GetAllFactions()
	if len(factions) == 0 {
		s += helpStyle.Render("No factions exist yet. Create one!")
		s += "\n\n" + renderFooter("C: Create Faction | ESC: Back")
		return s
	}

	s += "Available Factions:\n\n"
	for i, faction := range factions {
		cursor := "  "
		if i == m.factionsModel.cursor {
			cursor = "> "
		}

		recruiting := ""
		if faction.IsRecruiting {
			recruiting = successStyle.Render(" [Recruiting]")
		}

		s += fmt.Sprintf("%s%s - %d members | Level %d%s\n",
			cursor, faction.GetFullName(), len(faction.Members), faction.Level, recruiting)
	}

	s += "\n" + renderFooter("C: Create | V: View My Faction | ESC: Back")
	return s
}

func (m Model) viewFactionsCreate() string {
	s := titleStyle.Render("🏛️  CREATE FACTION") + "\n\n"

	// Name field
	namePrompt := "Name: "
	if m.factionsModel.inputField == 0 {
		namePrompt = highlightStyle.Render("Name: ")
	}
	s += namePrompt + m.factionsModel.createName
	if m.factionsModel.inputField == 0 {
		s += "█"
	}
	s += "\n\n"

	// Tag field
	tagPrompt := "Tag (3-5 chars): "
	if m.factionsModel.inputField == 1 {
		tagPrompt = highlightStyle.Render("Tag (3-5 chars): ")
	}
	s += tagPrompt + m.factionsModel.createTag
	if m.factionsModel.inputField == 1 {
		s += "█"
	}
	s += "\n\n"

	// Alignment
	alignPrompt := "Alignment: " + m.factionsModel.createAlign
	if m.factionsModel.inputField == 2 {
		alignPrompt = highlightStyle.Render(alignPrompt)
	}
	s += alignPrompt + "\n\n"

	s += helpStyle.Render("Available: trader, mercenary, explorer, pirate, corporate")
	s += "\n\n" + renderFooter("Tab: Next Field | Enter: Create | ESC: Cancel")

	return s
}

func (m Model) viewMyFaction() string {
	faction, err := m.factionManager.GetPlayerFaction(m.playerID)
	if err != nil {
		return "You are not in a faction\n\n" + renderFooter("ESC: Back")
	}

	s := titleStyle.Render(fmt.Sprintf("🏛️  %s", faction.GetFullName())) + "\n\n"

	// Faction info
	s += fmt.Sprintf("Leader: %s\n", faction.LeaderID)
	s += fmt.Sprintf("Founded: %s\n", faction.CreatedAt.Format("2006-01-02"))
	s += fmt.Sprintf("Members: %d/%d\n", len(faction.Members), faction.MemberLimit)
	s += fmt.Sprintf("Level: %d (XP: %d)\n", faction.Level, faction.Experience)
	s += fmt.Sprintf("Treasury: %d CR\n", faction.Treasury)
	s += fmt.Sprintf("Alignment: %s\n", faction.Alignment)
	s += strings.Repeat("─", 80) + "\n\n"

	// Member list
	s += "Members:\n"
	s += fmt.Sprintf("  Leader: %s\n", faction.LeaderID)
	for _, officerID := range faction.Officers {
		s += fmt.Sprintf("  ⭐ Officer: %s\n", officerID)
	}
	for _, memberID := range faction.Members {
		if memberID != faction.LeaderID && !contains(faction.Officers, memberID) {
			s += fmt.Sprintf("  👤 Member: %s\n", memberID)
		}
	}

	s += "\n" + renderFooter("ESC: Back to List")
	return s
}

func contains(slice []uuid.UUID, item uuid.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
