// File: internal/tui/factions.go
// Project: Terminal Velocity
// Description: Factions screen - Player faction management with creation and membership
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The factions screen provides:
// - Faction listing with member counts and levels
// - Faction creation interface with name, tag, and alignment
// - Current faction viewing with member list and details
// - Faction treasury information
// - Member role display (Leader, Officers, Members)
// - Recruitment status indicators
// - Faction statistics dashboard
//
// View Modes:
//   - list: Browse all factions on server
//   - my_faction: View current faction details (if member)
//   - create: Create new faction form
//
// Faction Creation:
//   - Name: 1-30 characters
//   - Tag: 3-5 characters (auto-uppercase)
//   - Alignment: trader, mercenary, explorer, pirate, corporate
//
// Faction Details:
//   - Leader and officers display
//   - Member count and limit
//   - Faction level and experience
//   - Treasury balance
//   - Founded date
//   - Recruitment status
//
// Visual Features:
//   - [Recruiting] badge for open factions
//   - Role indicators (⭐ for officers, 👤 for members)
//   - Full faction name display with tag: "Name [TAG]"
//   - Level and experience progress

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// factionsModel contains the state for the factions screen.
// Manages faction browsing, creation, and membership viewing.
type factionsModel struct {
	viewMode    string // Current view: "list", "my_faction", "create"
	cursor      int    // Current cursor position in faction list
	createName  string // Faction name input (creation mode)
	createTag   string // Faction tag input (creation mode)
	createAlign string // Faction alignment input (creation mode)
	inputField  int    // Active input field in creation: 0=name, 1=tag, 2=alignment
}

// newFactionsModel creates and initializes a new factions screen model.
// Starts in list view with trader alignment default for creation.
func newFactionsModel() factionsModel {
	return factionsModel{
		viewMode:    "list",
		cursor:      0,
		createAlign: models.AlignmentTrader,
	}
}

// updateFactions handles input and state updates for the factions screen.
//
// Key Bindings (List/My Faction Mode):
//   - esc/backspace/q: Return to main menu (or list from my_faction)
//   - up/k: Move cursor up in faction list
//   - down/j: Move cursor down in faction list
//   - c: Enter faction creation mode
//   - v: View current faction details
//
// Key Bindings (Create Mode):
//   - esc: Cancel creation, return to list
//   - tab: Cycle through input fields (name → tag → alignment)
//   - enter: Submit faction creation (if valid)
//   - backspace: Delete character from text fields
//   - Any char: Add to active text field
//
// Faction Creation Flow:
//   1. Press 'c' from list view
//   2. Fill name (1-30 chars)
//   3. Fill tag (3-5 chars, auto-uppercase)
//   4. Select alignment (shown on field 2)
//   5. Press Enter to create
//
// Validation:
//   - Name and tag required
//   - Alignment from predefined list
func (m Model) updateFactions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.factionsModel.viewMode == "create" {
			return m.updateFactionsCreate(msg)
		}

		switch msg.String() {
		case "esc", "backspace", "q":
			if m.factionsModel.viewMode == "my_faction" {
				// Return to faction list from details view
				m.factionsModel.viewMode = "list"
				return m, nil
			}
			// Return to main menu from list view
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

// updateFactionsCreate handles input in faction creation mode.
// Manages text input for name/tag fields and faction creation submission.
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

// viewFactions renders the factions screen (dispatches to appropriate view).
//
// Layout:
//   - Routes to viewFactionsCreate() if in create mode
//   - Routes to viewMyFaction() if viewing own faction
//   - Otherwise renders faction list view
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

// viewFactionsCreate renders the faction creation form.
//
// Layout:
//   - Title: "🏛️  CREATE FACTION"
//   - Name field with cursor if active
//   - Tag field with cursor if active
//   - Alignment field (displays current selection)
//   - Available alignments help text
//   - Footer with controls
//
// Visual Features:
//   - Active field highlighted
//   - Input cursor (█) on focused field
//   - Tag auto-uppercased
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

// viewMyFaction renders the player's current faction details.
//
// Layout:
//   - Title: Faction name with tag
//   - Faction info: Leader, Founded, Members, Level, Treasury, Alignment
//   - Member list: Leader, Officers (⭐), Members (👤)
//   - Footer with controls
//
// Returns error message if player not in a faction.
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

// contains checks if a UUID exists in a slice.
// Helper function for member list filtering.
func contains(slice []uuid.UUID, item uuid.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
