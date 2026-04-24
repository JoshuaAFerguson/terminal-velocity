// File: internal/tui/leaderboards.go
// Project: Terminal Velocity
// Description: Leaderboards screen - Competitive rankings across multiple categories with player stats
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The leaderboards screen provides:
// - 7 competitive ranking categories (Overall, Combat, Trading, Exploration, Wealth, Missions, Reputation)
// - Global top rankings view (top 15 players)
// - Near-player view (±7 positions around current player)
// - Live player ranking and score display
// - Real-time leaderboard refresh
// - Detailed stats for each category
// - Medal indicators for top 3 positions (🥇🥈🥉)
//
// Categories:
//   - Overall: Combined performance across all areas
//   - Combat: Combat rating, kills, rank title
//   - Trading: Trade volume, profit, rating
//   - Exploration: Systems visited, jumps made
//   - Wealth: Total credits and assets
//   - Missions: Completed and failed missions
//   - Reputation: Total reputation across all factions
//
// View Modes:
//   - Global: Shows top players across entire server
//   - Near Player: Shows rankings around player's position

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// leaderboardCategoryTab is a single tab entry. The tab bar derives its
// order, digit shortcuts, icons, and the 1-7 keymap from this slice so
// the view and the update handler stay in sync without a second source
// of truth. Ordering here defines tab index for arrow-key navigation.
type leaderboardCategoryTab struct {
	key      string
	label    string
	category models.LeaderboardCategory
}

var leaderboardCategoryTabs = []leaderboardCategoryTab{
	{"1", "Overall", models.LeaderboardOverall},
	{"2", "Combat", models.LeaderboardCombat},
	{"3", "Trading", models.LeaderboardTrading},
	{"4", "Exploration", models.LeaderboardExploration},
	{"5", "Wealth", models.LeaderboardWealth},
	{"6", "Missions", models.LeaderboardMissions},
	{"7", "Reputation", models.LeaderboardReputation},
}

// leaderboardCategoryIndex returns the slice index for a category, or
// -1 if the category isn't in the tab list.
func leaderboardCategoryIndex(cat models.LeaderboardCategory) int {
	for i, t := range leaderboardCategoryTabs {
		if t.category == cat {
			return i
		}
	}
	return -1
}

// leaderboardCategoryByDigit resolves a digit keystring (e.g. "3") to a
// category. Returns ok=false on any non-matching key.
func leaderboardCategoryByDigit(digit string) (models.LeaderboardCategory, bool) {
	for _, t := range leaderboardCategoryTabs {
		if t.key == digit {
			return t.category, true
		}
	}
	return "", false
}

// cycleLeaderboardCategory returns the next/previous category with
// wrap-around. delta is typically +1 (right) or -1 (left). Invalid
// current category falls back to the first entry.
func cycleLeaderboardCategory(cur models.LeaderboardCategory, delta int) models.LeaderboardCategory {
	n := len(leaderboardCategoryTabs)
	if n == 0 {
		return cur
	}
	idx := leaderboardCategoryIndex(cur)
	if idx < 0 {
		return leaderboardCategoryTabs[0].category
	}
	// Go modulo can return negative for negative operands; normalize.
	next := ((idx+delta)%n + n) % n
	return leaderboardCategoryTabs[next].category
}

// leaderboardsModel contains the state for the leaderboards screen.
// Manages category selection, view mode, and cursor position.
type leaderboardsModel struct {
	cursor           int                        // Current cursor position in leaderboard list
	selectedCategory models.LeaderboardCategory // Currently displayed category
	viewMode         string                     // Display mode: "global" or "near_player"
	displayCount     int                        // Number of entries to show (default: 15)
}

// newLeaderboardsModel creates and initializes a new leaderboards screen model.
// Starts with Overall category in global view mode.
func newLeaderboardsModel() leaderboardsModel {
	return leaderboardsModel{
		cursor:           0,
		selectedCategory: models.LeaderboardOverall,
		viewMode:         "global",
		displayCount:     15,
	}
}

// updateLeaderboards handles input and state updates for the leaderboards screen.
//
// Key Bindings:
//   - esc/backspace/q: Return to main menu
//   - up/k: Move cursor up in leaderboard list
//   - down/j: Move cursor down in leaderboard list
//   - v: Toggle between global and near-player view
//   - r: Refresh leaderboards (reload from manager)
//   - 1-7: Switch to specific category
//     - 1: Overall rankings
//     - 2: Combat rankings
//     - 3: Trading rankings
//     - 4: Exploration rankings
//     - 5: Wealth rankings
//     - 6: Missions rankings
//     - 7: Reputation rankings
//
// View Modes:
//   - Global: Shows top `displayCount` players (default: 15)
//   - Near Player: Shows players within ±7 ranks of current player
//
// Features:
//   - Automatic cursor clamping to available entries
//   - Cursor reset when changing categories
//   - Live rank and score display
//   - Last update timestamp
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

		case "left":
			m.leaderboardsModel.selectedCategory = cycleLeaderboardCategory(m.leaderboardsModel.selectedCategory, -1)
			m.leaderboardsModel.cursor = 0

		case "right":
			m.leaderboardsModel.selectedCategory = cycleLeaderboardCategory(m.leaderboardsModel.selectedCategory, 1)
			m.leaderboardsModel.cursor = 0

		default:
			// Digit shortcut: 1-7 jumps directly to that category.
			// Up/down are handled above; 'h'/'l' remain free (vim-nav
			// on the leaderboard row cursor, not tab switching), so
			// tab motion is arrow-keys only.
			if cat, ok := leaderboardCategoryByDigit(msg.String()); ok {
				m.leaderboardsModel.selectedCategory = cat
				m.leaderboardsModel.cursor = 0
			}
		}
	}

	return m, nil
}

// viewLeaderboards renders the leaderboards screen.
//
// Layout:
//   - Title: Icon + "LEADERBOARDS - [Category Name]"
//   - Stats Header: Total players, player's rank and score
//   - Last Updated: Timestamp of leaderboard snapshot
//   - Category Tabs: 7 categories with active indicator
//   - View Mode Indicator: Global vs Near Player
//   - Leaderboard Table: Rank, Player, Score, Details
//   - Footer: Navigation controls
//
// Visual Features:
//   - Medal emojis for top 3 (🥇🥈🥉)
//   - Player's own entry highlighted
//   - Category-specific details formatting
//   - Active tab highlighted
//   - Cursor selection indicator
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

	// Category tabs. Two-line underline style: line 1 shows the icon
	// + label + "(n)" digit shortcut for every tab, line 2 draws a
	// heavy ━ under only the active tab. Much easier to scan at a
	// glance than the prior bracketed inline strip.
	s += renderLeaderboardTabs(m.leaderboardsModel.selectedCategory) + "\n\n"

	// View mode toggle
	if m.leaderboardsModel.viewMode == "near_player" && playerRank > 0 {
		s += helpStyle.Render("📍 Showing rankings near you (Press V for global view)\n\n")
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
		s += renderFooter("ESC: Back | ←/→ or 1-7: Category | V: Toggle View | R: Refresh")
		return s
	}

	// Display leaderboard entries
	s += m.renderLeaderboardEntries(entries, category)

	// Footer
	s += "\n" + renderFooter("↑/↓: Navigate | ←/→ or 1-7: Category | V: Toggle View | R: Refresh | ESC: Back")

	return s
}

// renderLeaderboardEntries renders the leaderboard table with player entries.
// Formats each entry with rank, medals, player name, score, and category-specific details.
// Highlights the current player's entry with special styling.
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

// formatScore formats a leaderboard score based on category.
// Adds appropriate units (CR for wealth, rep for reputation, plain numbers for others).
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

// formatLeaderboardDetails formats category-specific details for a leaderboard entry.
// Each category shows different relevant stats (kills for combat, trades for trading, etc.).
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

// tabActiveStyle and tabInactiveStyle are local (no margins) because
// the global helpStyle has MarginTop(1) — fine for inline help lines
// but it injects a leading newline into every rendered cell, which
// would break a single-line tab strip. Same colors as helpStyle /
// highlightStyle so the visual language stays consistent.
var (
	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Cyan
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")) // Gray
)

// renderLeaderboardTabs draws the two-line tab strip. Line 1 is the
// icon + label + "(N)" digit shortcut for every tab. Line 2 draws a
// heavy ━ run under ONLY the active tab, in the same highlight color,
// which reads as a Chrome-style underline without needing a full box.
//
// Widths are measured with lipgloss.Width so emoji/variation-selector
// runes that occupy 2 cells (🏆, ⚔️, 🧭 …) underline correctly. Byte
// len() on these strings would under-count the active width and leave
// a visible gap between the label and its underline.
func renderLeaderboardTabs(active models.LeaderboardCategory) string {
	var labels, underlines strings.Builder
	sep := "  " // two spaces between tabs

	for i, t := range leaderboardCategoryTabs {
		icon := models.GetCategoryIcon(t.category)
		// "( N )" trailing digit hint lives inside each tab's cell so
		// the underline covers both the label and its shortcut.
		cell := fmt.Sprintf(" %s %s (%s) ", icon, t.label, t.key)
		width := lipgloss.Width(cell)

		isActive := t.category == active
		if isActive {
			labels.WriteString(tabActiveStyle.Render(cell))
			underlines.WriteString(tabActiveStyle.Render(strings.Repeat("━", width)))
		} else {
			labels.WriteString(tabInactiveStyle.Render(cell))
			underlines.WriteString(strings.Repeat(" ", width))
		}

		if i < len(leaderboardCategoryTabs)-1 {
			labels.WriteString(sep)
			underlines.WriteString(strings.Repeat(" ", lipgloss.Width(sep)))
		}
	}

	return labels.String() + "\n" + underlines.String()
}
