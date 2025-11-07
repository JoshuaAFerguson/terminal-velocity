// File: internal/tui/leaderboards.go
// Project: Terminal Velocity
// Description: Leaderboard UI displaying competitive rankings across multiple categories
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type leaderboardsModel struct {
	cursor           int
	selectedCategory models.LeaderboardCategory
	viewMode         string // "global" or "near_player"
	displayCount     int    // Number of entries to show
}

func newLeaderboardsModel() leaderboardsModel {
	return leaderboardsModel{
		cursor:           0,
		selectedCategory: models.LeaderboardOverall,
		viewMode:         "global",
		displayCount:     15,
	}
}

func (m Model) updateLeaderboards(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			// Go back to main menu
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.leaderboardsModel.cursor > 0 {
				m.leaderboardsModel.cursor--
			}

		case "down", "j":
			// Limit based on current leaderboard size
			snapshot := m.leaderboardManager.GetLeaderboard(m.leaderboardsModel.selectedCategory)
			if snapshot != nil {
				maxCursor := len(snapshot.Entries) - 1
				if maxCursor > m.leaderboardsModel.displayCount-1 {
					maxCursor = m.leaderboardsModel.displayCount - 1
				}
				if m.leaderboardsModel.cursor < maxCursor {
					m.leaderboardsModel.cursor++
				}
			}

		case "v":
			// Toggle view mode between global and near player
			if m.leaderboardsModel.viewMode == "global" {
				m.leaderboardsModel.viewMode = "near_player"
			} else {
				m.leaderboardsModel.viewMode = "global"
			}
			m.leaderboardsModel.cursor = 0

		case "r":
			// Refresh leaderboards
			return m, m.refreshLeaderboards()

		// Category selection shortcuts
		case "1":
			m.leaderboardsModel.selectedCategory = models.LeaderboardOverall
			m.leaderboardsModel.cursor = 0

		case "2":
			m.leaderboardsModel.selectedCategory = models.LeaderboardCombat
			m.leaderboardsModel.cursor = 0

		case "3":
			m.leaderboardsModel.selectedCategory = models.LeaderboardTrading
			m.leaderboardsModel.cursor = 0

		case "4":
			m.leaderboardsModel.selectedCategory = models.LeaderboardExploration
			m.leaderboardsModel.cursor = 0

		case "5":
			m.leaderboardsModel.selectedCategory = models.LeaderboardWealth
			m.leaderboardsModel.cursor = 0

		case "6":
			m.leaderboardsModel.selectedCategory = models.LeaderboardMissions
			m.leaderboardsModel.cursor = 0

		case "7":
			m.leaderboardsModel.selectedCategory = models.LeaderboardReputation
			m.leaderboardsModel.cursor = 0
		}
	}

	return m, nil
}

func (m Model) viewLeaderboards() string {
	category := m.leaderboardsModel.selectedCategory
	icon := models.GetCategoryIcon(category)
	displayName := models.GetCategoryDisplayName(category)

	s := titleStyle.Render(icon+" LEADERBOARDS - "+displayName) + "\n\n"

	// Stats header
	snapshot := m.leaderboardManager.GetLeaderboard(category)
	if snapshot == nil {
		s += helpStyle.Render("No leaderboard data available yet.\n")
		s += helpStyle.Render("Leaderboards will be generated as players compete.\n\n")
		s += renderFooter("ESC: Back | 1-7: Change Category | R: Refresh")
		return s
	}

	// Player's current rank
	playerRank := m.leaderboardManager.GetPlayerRank(m.playerID, category)
	playerEntry := m.leaderboardManager.GetPlayerEntry(m.playerID, category)

	s += fmt.Sprintf("Total Players: %d | ", snapshot.TotalPlayers)
	if playerRank > 0 {
		s += fmt.Sprintf("Your Rank: %s #%d", models.GetRankMedal(playerRank), playerRank)
		if playerEntry != nil {
			s += fmt.Sprintf(" (Score: %s)", m.formatScore(category, playerEntry.Score))
		}
	} else {
		s += "Your Rank: Unranked"
	}
	s += "\n"

	lastUpdate := snapshot.UpdatedAt.Format("15:04:05")
	s += helpStyle.Render(fmt.Sprintf("Last Updated: %s", lastUpdate))
	s += "\n"
	s += strings.Repeat("─", 80) + "\n\n"

	// Category tabs
	tabs := []struct {
		key      string
		label    string
		category models.LeaderboardCategory
	}{
		{"1", "Overall", models.LeaderboardOverall},
		{"2", "Combat", models.LeaderboardCombat},
		{"3", "Trading", models.LeaderboardTrading},
		{"4", "Exploration", models.LeaderboardExploration},
		{"5", "Wealth", models.LeaderboardWealth},
		{"6", "Missions", models.LeaderboardMissions},
		{"7", "Reputation", models.LeaderboardReputation},
	}

	s += "Categories: "
	for i, tab := range tabs {
		isActive := m.leaderboardsModel.selectedCategory == tab.category

		if isActive {
			s += highlightStyle.Render("[" + tab.label + "]")
		} else {
			s += helpStyle.Render(" " + tab.label + " ")
		}

		if i < len(tabs)-1 {
			s += " "
		}
	}
	s += "\n\n"

	// View mode toggle
	if m.leaderboardsModel.viewMode == "near_player" && playerRank > 0 {
		s += helpStyle.Render(fmt.Sprintf("📍 Showing rankings near you (Press V for global view)\n\n"))
	} else {
		s += helpStyle.Render("🌍 Showing global top rankings (Press V for near-you view)\n\n")
	}

	// Get entries to display
	var entries []*models.LeaderboardEntry
	if m.leaderboardsModel.viewMode == "near_player" && playerRank > 0 {
		// Show entries around the player
		entries = m.leaderboardManager.GetLeaderboardsAroundPlayer(m.playerID, category, 7, 7)
	} else {
		// Show top entries
		entries = m.leaderboardManager.GetTopEntries(category, m.leaderboardsModel.displayCount)
	}

	if len(entries) == 0 {
		s += helpStyle.Render("No entries to display.\n\n")
		s += renderFooter("ESC: Back | 1-7: Change Category | V: Toggle View | R: Refresh")
		return s
	}

	// Display leaderboard entries
	s += m.renderLeaderboardEntries(entries, category)

	// Footer
	s += "\n" + renderFooter("↑/↓: Navigate | 1-7: Category | V: Toggle View | R: Refresh | ESC: Back")

	return s
}

func (m Model) renderLeaderboardEntries(entries []*models.LeaderboardEntry, category models.LeaderboardCategory) string {
	var s strings.Builder

	// Header row
	s.WriteString(statsStyle.Render("Rank") + "  ")
	s.WriteString(statsStyle.Render("Player") + strings.Repeat(" ", 20-len("Player")))
	s.WriteString(statsStyle.Render("Score") + strings.Repeat(" ", 15-len("Score")))
	s.WriteString(statsStyle.Render("Details"))
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", 80) + "\n")

	for i, entry := range entries {
		cursor := "  "
		if i == m.leaderboardsModel.cursor {
			cursor = "> "
		}

		// Rank with medal
		rankStr := fmt.Sprintf("#%d", entry.Rank)
		medal := models.GetRankMedal(entry.Rank)
		if medal != "" {
			rankStr = medal + " " + rankStr
		}

		// Highlight player's own entry
		isPlayerEntry := entry.PlayerID == m.playerID
		nameStyle := normalStyle
		if isPlayerEntry {
			nameStyle = successStyle
		}

		// Player name (truncated if needed)
		playerName := entry.PlayerName
		if len(playerName) > 18 {
			playerName = playerName[:15] + "..."
		}

		// Score formatted
		scoreStr := m.formatScore(category, entry.Score)

		// Details based on category
		detailsStr := m.formatLeaderboardDetails(category, entry)

		// Build the line
		line := cursor
		line += fmt.Sprintf("%-6s", rankStr)
		line += nameStyle.Render(fmt.Sprintf("%-20s", playerName))
		line += fmt.Sprintf("%-15s", scoreStr)
		line += detailsStr

		s.WriteString(line + "\n")
	}

	return s.String()
}

func (m Model) formatScore(category models.LeaderboardCategory, score int64) string {
	switch category {
	case models.LeaderboardWealth:
		return fmt.Sprintf("%d CR", score)
	case models.LeaderboardReputation:
		return fmt.Sprintf("%d rep", score)
	default:
		return fmt.Sprintf("%d", score)
	}
}

func (m Model) formatLeaderboardDetails(category models.LeaderboardCategory, entry *models.LeaderboardEntry) string {
	switch category {
	case models.LeaderboardCombat:
		kills := entry.Details["kills"]
		rating := entry.Details["rating"]
		rankTitle := entry.Details["rank_title"]
		return fmt.Sprintf("%v kills • Rating: %v (%v)", kills, rating, rankTitle)

	case models.LeaderboardTrading:
		trades := entry.Details["trades"]
		profit := entry.Details["profit"]
		rating := entry.Details["rating"]
		return fmt.Sprintf("%v trades • Profit: %v CR • Rating: %v", trades, profit, rating)

	case models.LeaderboardExploration:
		systems := entry.Details["systems"]
		jumps := entry.Details["jumps"]
		rating := entry.Details["rating"]
		return fmt.Sprintf("%v systems • %v jumps • Rating: %v", systems, jumps, rating)

	case models.LeaderboardWealth:
		credits := entry.Details["credits"]
		return fmt.Sprintf("%v CR in assets", credits)

	case models.LeaderboardReputation:
		totalRep := entry.Details["total_reputation"]
		factionCount := entry.Details["faction_count"]
		return fmt.Sprintf("%v total • %v factions", totalRep, factionCount)

	case models.LeaderboardMissions:
		completed := entry.Details["completed"]
		failed := entry.Details["failed"]
		return fmt.Sprintf("%v completed • %v failed", completed, failed)

	case models.LeaderboardOverall:
		rankTitle := entry.Details["rank_title"]
		combatRating := entry.Details["combat_rating"]
		tradingRating := entry.Details["trading_rating"]
		return fmt.Sprintf("%v • C:%v T:%v", rankTitle, combatRating, tradingRating)

	default:
		return ""
	}
}
