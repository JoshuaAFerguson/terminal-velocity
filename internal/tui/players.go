// File: internal/tui/players.go
// Project: Terminal Velocity
// Description: Player list UI displaying online players and their status
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type playersModel struct {
	cursor       int
	filterMode   string // "all", "same_system", "nearby", "combat"
	sortMode     string // "name", "rating", "online_time", "activity"
	selectedPlayer *models.PlayerPresence
}

func newPlayersModel() playersModel {
	return playersModel{
		cursor:     0,
		filterMode: "all",
		sortMode:   "name",
	}
}

func (m Model) updatePlayers(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			// Go back to main menu
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.playersModel.cursor > 0 {
				m.playersModel.cursor--
			}

		case "down", "j":
			players := m.getFilteredPlayers()
			if m.playersModel.cursor < len(players)-1 {
				m.playersModel.cursor++
			}

		case "r":
			// Refresh player list
			return m, nil

		// Filter mode shortcuts
		case "1":
			m.playersModel.filterMode = "all"
			m.playersModel.cursor = 0

		case "2":
			m.playersModel.filterMode = "same_system"
			m.playersModel.cursor = 0

		case "3":
			m.playersModel.filterMode = "nearby"
			m.playersModel.cursor = 0

		case "4":
			m.playersModel.filterMode = "combat"
			m.playersModel.cursor = 0

		// Sort mode shortcuts
		case "s":
			// Cycle through sort modes
			switch m.playersModel.sortMode {
			case "name":
				m.playersModel.sortMode = "rating"
			case "rating":
				m.playersModel.sortMode = "online_time"
			case "online_time":
				m.playersModel.sortMode = "activity"
			case "activity":
				m.playersModel.sortMode = "name"
			}
			m.playersModel.cursor = 0
		}
	}

	return m, nil
}

func (m Model) viewPlayers() string {
	s := titleStyle.Render("👥 ONLINE PLAYERS") + "\n\n"

	// Stats header
	stats := m.presenceManager.GetStats()
	s += fmt.Sprintf("Online: %d", stats.TotalOnline)
	if stats.InCombat > 0 {
		s += " | " + errorStyle.Render(fmt.Sprintf("⚔️  Combat: %d", stats.InCombat))
	}
	if stats.Trading > 0 {
		s += fmt.Sprintf(" | 💰 Trading: %d", stats.Trading)
	}
	if stats.Docked > 0 {
		s += fmt.Sprintf(" | 🛬 Docked: %d", stats.Docked)
	}
	if stats.InSpace > 0 {
		s += fmt.Sprintf(" | 🚀 In Space: %d", stats.InSpace)
	}
	s += "\n"
	s += strings.Repeat("─", 80) + "\n\n"

	// Filter tabs
	tabs := []struct {
		key    string
		label  string
		mode   string
	}{
		{"1", "All Players", "all"},
		{"2", "Same System", "same_system"},
		{"3", "Nearby", "nearby"},
		{"4", "In Combat", "combat"},
	}

	s += "Filter: "
	for i, tab := range tabs {
		isActive := m.playersModel.filterMode == tab.mode

		if isActive {
			s += highlightStyle.Render("[" + tab.label + "]")
		} else {
			s += helpStyle.Render(" " + tab.label + " ")
		}

		if i < len(tabs)-1 {
			s += " "
		}
	}
	s += "\n"

	// Sort indicator
	sortLabels := map[string]string{
		"name":        "Name",
		"rating":      "Combat Rating",
		"online_time": "Online Time",
		"activity":    "Activity",
	}
	s += helpStyle.Render(fmt.Sprintf("Sort: %s (S to change)", sortLabels[m.playersModel.sortMode]))
	s += "\n\n"

	// Get filtered and sorted players
	players := m.getFilteredPlayers()

	if len(players) == 0 {
		emptyMsg := "No players online."
		if m.playersModel.filterMode == "same_system" {
			emptyMsg = "No other players in your system."
		} else if m.playersModel.filterMode == "nearby" {
			emptyMsg = "No nearby players available for interaction."
		} else if m.playersModel.filterMode == "combat" {
			emptyMsg = "No players currently in combat."
		}

		s += helpStyle.Render(emptyMsg) + "\n\n"
		s += renderFooter("ESC: Back | 1-4: Filter | S: Sort | R: Refresh")
		return s
	}

	// Display player list
	s += m.renderPlayerList(players)

	// Footer
	s += "\n" + renderFooter("↑/↓: Navigate | 1-4: Filter | S: Sort | R: Refresh | ESC: Back")

	return s
}

func (m Model) getFilteredPlayers() []*models.PlayerPresence {
	var players []*models.PlayerPresence

	switch m.playersModel.filterMode {
	case "all":
		players = m.presenceManager.GetAllOnline()

	case "same_system":
		if m.player != nil {
			players = m.presenceManager.GetPlayersInSystem(m.player.CurrentSystem)
			// Filter out self
			filtered := []*models.PlayerPresence{}
			for _, p := range players {
				if p.PlayerID != m.playerID {
					filtered = append(filtered, p)
				}
			}
			players = filtered
		}

	case "nearby":
		players = m.presenceManager.GetNearbyPlayers(m.playerID)

	case "combat":
		players = m.presenceManager.GetPlayersInCombat()
	}

	// Sort players
	m.sortPlayers(players)

	return players
}

func (m Model) sortPlayers(players []*models.PlayerPresence) {
	switch m.playersModel.sortMode {
	case "name":
		sort.Slice(players, func(i, j int) bool {
			return players[i].Username < players[j].Username
		})

	case "rating":
		sort.Slice(players, func(i, j int) bool {
			return players[i].CombatRating > players[j].CombatRating
		})

	case "online_time":
		sort.Slice(players, func(i, j int) bool {
			return players[i].GetOnlineDuration() > players[j].GetOnlineDuration()
		})

	case "activity":
		sort.Slice(players, func(i, j int) bool {
			return players[i].CurrentActivity < players[j].CurrentActivity
		})
	}
}

func (m Model) renderPlayerList(players []*models.PlayerPresence) string {
	var s strings.Builder

	// Header row
	s.WriteString(statsStyle.Render("Player") + strings.Repeat(" ", 20-len("Player")))
	s.WriteString(statsStyle.Render("Ship") + strings.Repeat(" ", 18-len("Ship")))
	s.WriteString(statsStyle.Render("Rating") + strings.Repeat(" ", 8-len("Rating")))
	s.WriteString(statsStyle.Render("Status") + strings.Repeat(" ", 16-len("Status")))
	s.WriteString(statsStyle.Render("Online"))
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", 80) + "\n")

	// Display players (limit to 12 visible)
	displayPlayers := players
	if len(displayPlayers) > 12 {
		displayPlayers = displayPlayers[:12]
	}

	for i, player := range displayPlayers {
		cursor := "  "
		if i == m.playersModel.cursor {
			cursor = "> "
		}

		// Player name with criminal indicator
		playerName := player.Username
		if player.IsCriminal {
			playerName = errorStyle.Render(playerName + " ⚠️")
		}
		if len(playerName) > 18 {
			playerName = playerName[:15] + "..."
		}

		// Ship info (truncated)
		shipInfo := player.ShipName
		if len(shipInfo) > 16 {
			shipInfo = shipInfo[:13] + "..."
		}

		// Combat rating
		ratingStr := fmt.Sprintf("%d", player.CombatRating)

		// Status
		statusStr := player.GetStatusString()

		// Online duration
		onlineStr := player.GetOnlineDurationString()

		// Build the line
		line := cursor
		line += fmt.Sprintf("%-20s", playerName)
		line += fmt.Sprintf("%-18s", shipInfo)
		line += fmt.Sprintf("%-8s", ratingStr)
		line += fmt.Sprintf("%-16s", statusStr)
		line += onlineStr

		s.WriteString(line + "\n")
	}

	if len(players) > 12 {
		s.WriteString("\n" + helpStyle.Render(fmt.Sprintf("... and %d more", len(players)-12)))
	}

	return s.String()
}
