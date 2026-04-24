// File: internal/tui/achievements.go
// Project: Terminal Velocity
// Description: Achievements screen - Achievement tracking and rewards interface
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The achievements screen provides a comprehensive achievement system:
// - Browse all achievements with filtering options
// - View unlocked achievements with timestamps
// - Track progress on locked achievements with progress bars
// - Filter by category: Combat, Trading, Exploration, Missions
// - Filter by status: All, Unlocked, Locked
// - View achievement details, descriptions, and point values
// - Track total achievement points earned
// - View hidden achievements (revealed only when unlocked)
//
// Achievement System:
// - Multiple categories: Combat, Trading, Exploration, Missions, Social, Story
// - Rarity tiers: Common, Rare, Epic, Legendary
// - Point values based on difficulty and rarity
// - Progress tracking for incremental achievements
// - Hidden achievements for spoiler-sensitive content
// - Unlock notifications during gameplay
// - Total points and completion percentage displayed
//
// Filtering System:
// - Tab 1: All achievements (complete list)
// - Tab 2: Unlocked only (earned achievements)
// - Tab 3: Locked only (not yet earned)
// - Tab 4-7: Category filters (Combat, Trading, Exploration, Missions)
// - Progress bars show completion for locked achievements
// - Hidden achievements show as "🔒 Hidden Achievement" until unlocked

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/achievements"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

var log = logger.WithComponent("Tui")

// achievementsModel contains the state for the achievements screen.
// Manages achievement display, filtering, and progress tracking.
type achievementsModel struct {
	cursor int                          // Current cursor position in achievement list
	tab    string                       // Current tab/filter: "all", "unlocked", "locked", or category name
	filter models.AchievementCategory   // Category filter for achievement display
}

// newAchievementsModel creates and initializes a new achievements screen model.
// Starts with "all" achievements view and cursor at top.
func newAchievementsModel() achievementsModel {
	return achievementsModel{
		cursor: 0,
		tab:    "all",
	}
}

// updateAchievements handles input and state updates for the achievements screen.
//
// Key Bindings:
//   - esc/backspace/q: Return to main menu
//   - up/k: Move cursor up in achievement list
//   - down/j: Move cursor down in achievement list
//   - 1: Show all achievements
//   - 2: Show unlocked achievements only
//   - 3: Show locked achievements only
//   - 4: Filter by Combat category
//   - 5: Filter by Trading category
//   - 6: Filter by Exploration category
//   - 7: Filter by Missions category
//
// Filtering Workflow:
//   1. Press number key (1-7) to activate filter
//   2. Achievement list updates to show filtered items
//   3. Cursor resets to top of filtered list
//   4. Progress updates for visible locked achievements
//
// Progress Display:
//   - Unlocked achievements show "✓ Unlocked" status
//   - Locked achievements show progress bar and percentage
//   - Hidden achievements show as locked until unlocked
//   - Progress bars color-coded: green (75%+), yellow (50-74%), normal (25-49%), dim (<25%)
//
// Message Handling:
//   - All updates happen synchronously through achievement manager
func (m Model) updateAchievements(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.achievementsUI.cursor > 0 {
				m.achievementsUI.cursor--
			}

		case "down", "j":
			// Get current list to determine max cursor
			var achievements []*models.Achievement
			if m.achievementsUI.tab == "all" {
				achievements = models.GetAllAchievements()
			} else if m.achievementsUI.tab == "unlocked" {
				achievements = m.achievementManager.GetUnlockedAchievements()
			} else if m.achievementsUI.tab == "locked" {
				achievements = m.achievementManager.GetLockedAchievements(false)
			}

			if m.achievementsUI.cursor < len(achievements)-1 {
				m.achievementsUI.cursor++
			}

		case "1":
			m.achievementsUI.tab = "all"
			m.achievementsUI.cursor = 0

		case "2":
			m.achievementsUI.tab = "unlocked"
			m.achievementsUI.cursor = 0

		case "3":
			m.achievementsUI.tab = "locked"
			m.achievementsUI.cursor = 0

		case "4":
			m.achievementsUI.tab = "combat"
			m.achievementsUI.filter = models.AchievementCategoryCombat
			m.achievementsUI.cursor = 0

		case "5":
			m.achievementsUI.tab = "trading"
			m.achievementsUI.filter = models.AchievementCategoryTrading
			m.achievementsUI.cursor = 0

		case "6":
			m.achievementsUI.tab = "exploration"
			m.achievementsUI.filter = models.AchievementCategoryExploration
			m.achievementsUI.cursor = 0

		case "7":
			m.achievementsUI.tab = "missions"
			m.achievementsUI.filter = models.AchievementCategoryMissions
			m.achievementsUI.cursor = 0
		}
	}

	return m, nil
}

// viewAchievements renders the achievements screen.
//
// Layout:
//   - Title: "🏆 ACHIEVEMENTS"
//   - Stats Header: Progress (unlocked/total), Points (earned/max)
//   - Separator line
//   - Filter Tabs: 1:All, 2:Unlocked, 3:Locked, 4:Combat, 5:Trading, 6:Exploration, 7:Missions
//   - Achievement List: Icon, title, rarity, points
//   - Achievement Details: Description, progress bar (locked), unlock status
//   - Footer: Key bindings help
//
// Achievement Display:
//   - Icon and title on first line
//   - Rarity badge [Common/Rare/Epic/Legendary] and point value
//   - Description (unless hidden and locked)
//   - Progress bar for locked achievements (30 chars width)
//   - Percentage display for progress
//   - "✓ Unlocked" indicator for earned achievements
//   - Cursor highlights selected achievement
//
// Visual Features:
//   - Active filter tab highlighted in accent color
//   - Unlocked achievements in success color (green)
//   - Hidden locked achievements show "🔒 Hidden Achievement"
//   - Progress bars color-coded by completion:
//     * Green: 75-100% complete
//     * Yellow: 50-74% complete
//     * Normal: 25-49% complete
//     * Dim: 0-24% complete
//   - Achievement icons displayed (from achievement definition)
//   - Rarity and points shown in help text style
//
// Stats Display:
//   - Total unlocked vs total achievements
//   - Total points earned vs maximum possible points
//   - Formatted as "Progress: X/Y unlocked | Points: X/Y"
func (m Model) viewAchievements() string {
	s := titleStyle.Render("🏆 ACHIEVEMENTS") + "\n\n"

	// Stats header
	totalUnlocked := m.achievementManager.GetUnlockCount()
	totalAchievements := achievements.GetTotalCount()
	totalPoints := m.achievementManager.GetTotalPoints()
	maxPoints := achievements.GetMaxPoints()

	s += fmt.Sprintf("Progress: %d/%d unlocked | Points: %d/%d\n",
		totalUnlocked, totalAchievements, totalPoints, maxPoints)
	s += strings.Repeat("─", 80) + "\n\n"

	// Tabs
	tabs := []string{"1:All", "2:Unlocked", "3:Locked", "4:Combat", "5:Trading", "6:Exploration", "7:Missions"}
	s += "Filters: "
	for i, tab := range tabs {
		tabName := strings.Split(tab, ":")[1]
		tabKey := strings.Split(tab, ":")[0]

		isActive := false
		switch m.achievementsUI.tab {
		case "all":
			isActive = tabKey == "1"
		case "unlocked":
			isActive = tabKey == "2"
		case "locked":
			isActive = tabKey == "3"
		case "combat":
			isActive = tabKey == "4"
		case "trading":
			isActive = tabKey == "5"
		case "exploration":
			isActive = tabKey == "6"
		case "missions":
			isActive = tabKey == "7"
		}

		if isActive {
			s += highlightStyle.Render("[" + tabName + "]")
		} else {
			s += helpStyle.Render(" " + tabName + " ")
		}

		if i < len(tabs)-1 {
			s += " "
		}
	}
	s += "\n\n"

	// Get achievements based on current tab
	var achievementsList []*models.Achievement
	if m.achievementsUI.tab == "all" {
		achievementsList = models.GetAllAchievements()
	} else if m.achievementsUI.tab == "unlocked" {
		achievementsList = m.achievementManager.GetUnlockedAchievements()
	} else if m.achievementsUI.tab == "locked" {
		achievementsList = m.achievementManager.GetLockedAchievements(false)
	} else {
		// Category filter
		achievementsList = m.achievementManager.GetAchievementsByCategory(m.achievementsUI.filter, false)
	}

	// Display achievements
	for i, achievement := range achievementsList {
		isUnlocked := m.achievementManager.IsUnlocked(achievement.ID)
		progress := m.achievementManager.GetProgress(achievement.ID, m.player)

		// Cursor indicator
		cursor := "  "
		if i == m.achievementsUI.cursor {
			cursor = "> "
		}

		// Icon and title
		title := achievement.Icon + " " + achievement.Title
		if isUnlocked {
			title = successStyle.Render(title)
		} else {
			if achievement.Hidden {
				title = helpStyle.Render("🔒 Hidden Achievement")
			} else {
				title = normalStyle.Render(title)
			}
		}

		s += cursor + title

		// Rarity and points
		rarityStr := fmt.Sprintf("[%s]", achievement.Rarity)
		pointsStr := fmt.Sprintf("%dpts", achievement.Points)
		s += " " + helpStyle.Render(rarityStr+" "+pointsStr)
		s += "\n"

		// Description
		if !achievement.Hidden || isUnlocked {
			desc := "   " + achievement.Description
			s += helpStyle.Render(desc) + "\n"
		}

		// Progress bar for locked achievements
		if !isUnlocked && !achievement.Hidden {
			progressBar := m.renderProgressBar(progress, 30)
			s += "   " + progressBar + fmt.Sprintf(" %d%%", progress) + "\n"
		}

		// Unlock timestamp for unlocked achievements
		if isUnlocked {
			// Note: We'd need to store unlock time to show this properly
			s += "   " + successStyle.Render("✓ Unlocked") + "\n"
		}

		s += "\n"
	}

	// Footer
	s += "\n" + renderFooter("↑/↓: Navigate | 1-7: Filter | ESC: Back")

	return s
}

func (m Model) renderProgressBar(progress int, width int) string {
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}

	filled := (progress * width) / 100
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	if progress >= 75 {
		return successStyle.Render(bar)
	} else if progress >= 50 {
		return highlightStyle.Render(bar)
	} else if progress >= 25 {
		return normalStyle.Render(bar)
	}
	return helpStyle.Render(bar)
}
