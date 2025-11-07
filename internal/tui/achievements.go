// File: internal/tui/achievements.go
// Project: Terminal Velocity
// Description: Terminal UI component for achievements
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/achievements"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type achievementsModel struct {
	cursor int
	tab    string // "all", "unlocked", "locked", or category name
	filter models.AchievementCategory
}

func newAchievementsModel() achievementsModel {
	return achievementsModel{
		cursor: 0,
		tab:    "all",
	}
}

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
			s += highlightStyle.Render("["+tabName+"]")
		} else {
			s += helpStyle.Render(" "+tabName+" ")
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
