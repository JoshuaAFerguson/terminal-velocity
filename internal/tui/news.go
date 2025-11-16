// File: internal/tui/news.go
// Project: Terminal Velocity
// Description: News screen - Galactic news feed with dynamic articles and filtering
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The news screen provides:
// - Dynamic news article generation based on game events
// - Category filtering (Combat, Economic, Political, General, Achievement)
// - Priority-based article display (Critical, High, Normal, Low)
// - Breaking news notifications
// - Player-specific news highlighting (articles triggered by player actions)
// - Article detail view with full content
// - Age-based sorting (newest first)
//
// News Generation:
// - Articles generated from player actions (trades, combat, exploration)
// - Universe events (faction wars, economic shifts, system events)
// - Achievements and milestones
// - Player-based articles highlighted with special indicator
//
// Visual Features:
// - Priority icons and color coding
// - Category badges with emoji icons
// - Age display (e.g., "2m ago", "5h ago")
// - Word wrapping for article bodies
// - Breaking news counter in header

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// newsModel contains the state for the news screen.
// Manages article list, filtering, and detail view navigation.
type newsModel struct {
	cursor          int                 // Current cursor position in article list
	selectedFilter  models.NewsCategory // Current category filter ("" for all news)
	viewMode        string              // Current view: "list" or "detail"
	selectedArticle *models.NewsArticle // Article being viewed in detail mode
}

// newNewsModel creates and initializes a new news screen model.
// Starts in list view with all categories filter and cursor at top.
func newNewsModel() newsModel {
	return newsModel{
		cursor:         0,
		selectedFilter: "",
		viewMode:       "list",
	}
}

// updateNews handles input and state updates for the news screen.
//
// Key Bindings (List Mode):
//   - esc/backspace/q: Return to main menu
//   - up/k: Move cursor up in article list
//   - down/j: Move cursor down in article list
//   - enter/space: View selected article in detail
//   - 1: Filter by all news
//   - 2: Filter by combat news
//   - 3: Filter by economic news
//   - 4: Filter by political news
//   - 5: Filter by general news
//   - 6: Filter by achievement news
//
// Key Bindings (Detail Mode):
//   - esc/backspace: Return to article list
//
// View Modes:
//   - list: Display articles in compact list with filters
//   - detail: Show full article with word-wrapped body
//
// Filtering:
//   - Category filters apply to article list
//   - Cursor resets to top when filter changes
//   - Breaking news indicator shows critical articles
func (m Model) updateNews(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.newsModel.viewMode == "detail" {
				// Go back to article list from detail view
				m.newsModel.viewMode = "list"
				m.newsModel.selectedArticle = nil
				return m, nil
			}
			// Return to main menu from list view
			m.screen = ScreenMainMenu
			return m, nil

		case "q":
			// Quick exit to main menu
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			// Move cursor up (vi-style navigation supported with k)
			if m.newsModel.viewMode == "list" && m.newsModel.cursor > 0 {
				m.newsModel.cursor--
			}

		case "down", "j":
			// Move cursor down (vi-style navigation supported with j)
			if m.newsModel.viewMode == "list" {
				articles := m.newsManager.GetRecentArticles(50, m.newsModel.selectedFilter)
				if m.newsModel.cursor < len(articles)-1 {
					m.newsModel.cursor++
				}
			}

		case "enter", " ":
			// View article detail
			if m.newsModel.viewMode == "list" {
				articles := m.newsManager.GetRecentArticles(50, m.newsModel.selectedFilter)
				if m.newsModel.cursor < len(articles) {
					m.newsModel.selectedArticle = articles[m.newsModel.cursor]
					m.newsModel.viewMode = "detail"
				}
			}

		case "1":
			// Show all news (no filter)
			m.newsModel.selectedFilter = ""
			m.newsModel.cursor = 0

		case "2":
			// Filter by combat news
			m.newsModel.selectedFilter = models.NewsCategoryCombat
			m.newsModel.cursor = 0

		case "3":
			// Filter by economic news
			m.newsModel.selectedFilter = models.NewsCategoryEconomic
			m.newsModel.cursor = 0

		case "4":
			// Filter by political news
			m.newsModel.selectedFilter = models.NewsCategoryPolitical
			m.newsModel.cursor = 0

		case "5":
			// Filter by general news
			m.newsModel.selectedFilter = models.NewsCategoryGeneral
			m.newsModel.cursor = 0

		case "6":
			// Filter by achievement news
			m.newsModel.selectedFilter = models.NewsCategoryAchievement
			m.newsModel.cursor = 0
		}
	}

	return m, nil
}

// viewNews renders the news screen (dispatches to list or detail view).
//
// Layout:
//   - Routes to viewNewsDetail() if in detail mode
//   - Otherwise renders article list view
//
// The news screen provides two main views:
//   1. List view: Compact article list with category filters
//   2. Detail view: Full article content with metadata
func (m Model) viewNews() string {
	if m.newsModel.viewMode == "detail" && m.newsModel.selectedArticle != nil {
		return m.viewNewsDetail()
	}

	// Render article list view
	s := titleStyle.Render("📰 GALACTIC NEWS") + "\n\n"

	// Stats header with article count and breaking news indicator
	totalArticles := m.newsManager.GetArticleCount()
	breakingNews := len(m.newsManager.GetBreakingNews())

	s += fmt.Sprintf("Active Articles: %d", totalArticles)
	if breakingNews > 0 {
		s += " | " + errorStyle.Render(fmt.Sprintf("⚠️  %d Breaking News", breakingNews))
	}
	s += "\n"
	s += strings.Repeat("─", 80) + "\n\n"

	// Category filters
	tabs := []struct {
		key      string
		label    string
		category models.NewsCategory
	}{
		{"1", "All News", ""},
		{"2", "Combat", models.NewsCategoryCombat},
		{"3", "Economic", models.NewsCategoryEconomic},
		{"4", "Political", models.NewsCategoryPolitical},
		{"5", "General", models.NewsCategoryGeneral},
		{"6", "Achievements", models.NewsCategoryAchievement},
	}

	s += "Categories: "
	for i, tab := range tabs {
		isActive := m.newsModel.selectedFilter == tab.category

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

	// Get articles
	articles := m.newsManager.GetRecentArticles(50, m.newsModel.selectedFilter)

	if len(articles) == 0 {
		s += helpStyle.Render("No news articles available.\n\n")
		s += helpStyle.Render("News will be generated based on your actions and universe events.")
		s += "\n\n" + renderFooter("ESC: Back | 1-6: Filter")
		return s
	}

	// Display articles (compact list view)
	displayCount := 15
	if len(articles) > displayCount {
		articles = articles[:displayCount]
	}

	for i, article := range articles {
		cursor := "  "
		if i == m.newsModel.cursor {
			cursor = "> "
		}

		// Priority icon
		priorityIcon := article.GetPriorityString()

		// Headline with styling based on priority
		headline := article.Headline
		if article.Priority == models.NewsPriorityCritical {
			headline = errorStyle.Render(headline)
		} else if article.Priority == models.NewsPriorityHigh {
			headline = highlightStyle.Render(headline)
		}

		// Category badge
		categoryBadge := m.getCategoryBadge(article.Category)

		// Age
		age := helpStyle.Render(article.GetAgeString())

		s += cursor + priorityIcon + " " + headline + "\n"
		s += "     " + categoryBadge + " • " + age

		// Player-based indicator
		if article.PlayerBased {
			s += " • " + successStyle.Render("⭐ Your Action")
		}

		s += "\n\n"
	}

	// Footer
	s += renderFooter("↑/↓: Navigate | Enter: Read Article | 1-6: Filter | ESC: Back")

	return s
}

// viewNewsDetail renders the detailed view of a single news article.
//
// Layout:
//   - Title: "📰 ARTICLE DETAIL"
//   - Metadata: Priority icon, category badge, age
//   - Headline: Color-coded by priority (critical=red, high=yellow, normal=default)
//   - Separator line
//   - Body: Word-wrapped article content (75 char width)
//   - Separator line
//   - Additional metadata: Related faction, player-based indicator
//   - Footer: Navigation help
//
// Visual Features:
//   - Priority affects headline color (critical/high get special styling)
//   - Category badge with emoji icon
//   - Age string (e.g., "2m ago", "1h ago", "3d ago")
//   - Player-based articles show star indicator
//   - Word wrapping prevents horizontal overflow
func (m Model) viewNewsDetail() string {
	article := m.newsModel.selectedArticle
	if article == nil {
		return "No article selected\n"
	}

	s := titleStyle.Render("📰 ARTICLE DETAIL") + "\n\n"

	// Priority and category metadata line
	priorityIcon := article.GetPriorityString()
	categoryBadge := m.getCategoryBadge(article.Category)
	s += priorityIcon + " " + categoryBadge + " • " + helpStyle.Render(article.GetAgeString()) + "\n\n"

	// Headline with priority-based styling
	headline := article.Headline
	if article.Priority == models.NewsPriorityCritical {
		headline = errorStyle.Render(headline)
	} else if article.Priority == models.NewsPriorityHigh {
		headline = highlightStyle.Render(headline)
	} else {
		headline = subtitleStyle.Render(headline)
	}
	s += headline + "\n\n"

	s += strings.Repeat("─", 80) + "\n\n"

	// Body text with word wrap (75 character width)
	wrappedBody := m.wordWrap(article.Body, 75)
	s += wrappedBody + "\n\n"

	s += strings.Repeat("─", 80) + "\n\n"

	// Additional metadata section
	if article.FactionID != "" {
		s += "Related Faction: " + statsStyle.Render(article.FactionID) + "\n"
	}

	if article.PlayerBased {
		s += successStyle.Render("⭐ This event was caused by your actions") + "\n"
	}

	s += "\n" + renderFooter("ESC: Back to List")

	return s
}

// getCategoryBadge returns a formatted badge string for a news category.
// Each category has an emoji icon and display name.
//
// Categories:
//   - Combat: ⚔️  Combat
//   - Economic: 💰 Economic
//   - Political: 🏛️  Political
//   - Exploration: 🚀 Exploration
//   - Military: 🎖️  Military
//   - Criminal: 🏴‍☠️ Criminal
//   - General: 📰 General
//   - Achievement: 🏆 Achievement
//
// Returns "📰 News" for unknown categories.
// Badge is styled with helpStyle for consistent appearance.
func (m Model) getCategoryBadge(category models.NewsCategory) string {
	badges := map[models.NewsCategory]string{
		models.NewsCategoryCombat:      "⚔️  Combat",
		models.NewsCategoryEconomic:    "💰 Economic",
		models.NewsCategoryPolitical:   "🏛️  Political",
		models.NewsCategoryExploration: "🚀 Exploration",
		models.NewsCategoryMilitary:    "🎖️  Military",
		models.NewsCategoryCriminal:    "🏴‍☠️ Criminal",
		models.NewsCategoryGeneral:     "📰 General",
		models.NewsCategoryAchievement: "🏆 Achievement",
	}

	badge, exists := badges[category]
	if !exists {
		return "📰 News"
	}

	return helpStyle.Render(badge)
}

// wordWrap wraps text to fit within a maximum line width.
// Breaks text on word boundaries to prevent mid-word splits.
//
// Algorithm:
//   1. Split text into words using whitespace
//   2. Build lines by adding words until maxWidth exceeded
//   3. When line would exceed width, start new line
//   4. Join lines with newlines
//
// Parameters:
//   - text: Input text to wrap
//   - maxWidth: Maximum characters per line
//
// Returns:
//   - Word-wrapped text with newlines at appropriate positions
//
// Note: Very long words that exceed maxWidth will be placed on their own line.
func (m Model) wordWrap(text string, maxWidth int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
