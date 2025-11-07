package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type newsModel struct {
	cursor         int
	selectedFilter models.NewsCategory // Empty string for all news
	viewMode       string               // "list" or "detail"
	selectedArticle *models.NewsArticle
}

func newNewsModel() newsModel {
	return newsModel{
		cursor:         0,
		selectedFilter: "",
		viewMode:       "list",
	}
}

func (m Model) updateNews(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.newsModel.viewMode == "detail" {
				// Go back to list
				m.newsModel.viewMode = "list"
				m.newsModel.selectedArticle = nil
				return m, nil
			}
			// Go back to main menu
			m.screen = ScreenMainMenu
			return m, nil

		case "q":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.newsModel.viewMode == "list" && m.newsModel.cursor > 0 {
				m.newsModel.cursor--
			}

		case "down", "j":
			if m.newsModel.viewMode == "list" {
				articles := m.newsManager.GetRecentArticles(50, m.newsModel.selectedFilter)
				if m.newsModel.cursor < len(articles)-1 {
					m.newsModel.cursor++
				}
			}

		case "enter", " ":
			if m.newsModel.viewMode == "list" {
				// View article detail
				articles := m.newsManager.GetRecentArticles(50, m.newsModel.selectedFilter)
				if m.newsModel.cursor < len(articles) {
					m.newsModel.selectedArticle = articles[m.newsModel.cursor]
					m.newsModel.viewMode = "detail"
				}
			}

		case "1":
			m.newsModel.selectedFilter = ""
			m.newsModel.cursor = 0

		case "2":
			m.newsModel.selectedFilter = models.NewsCategoryCombat
			m.newsModel.cursor = 0

		case "3":
			m.newsModel.selectedFilter = models.NewsCategoryEconomic
			m.newsModel.cursor = 0

		case "4":
			m.newsModel.selectedFilter = models.NewsCategoryPolitical
			m.newsModel.cursor = 0

		case "5":
			m.newsModel.selectedFilter = models.NewsCategoryGeneral
			m.newsModel.cursor = 0

		case "6":
			m.newsModel.selectedFilter = models.NewsCategoryAchievement
			m.newsModel.cursor = 0
		}
	}

	return m, nil
}

func (m Model) viewNews() string {
	if m.newsModel.viewMode == "detail" && m.newsModel.selectedArticle != nil {
		return m.viewNewsDetail()
	}

	s := titleStyle.Render("📰 GALACTIC NEWS") + "\n\n"

	// Stats header
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
			s += highlightStyle.Render("["+tab.label+"]")
		} else {
			s += helpStyle.Render(" "+tab.label+" ")
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

func (m Model) viewNewsDetail() string {
	article := m.newsModel.selectedArticle
	if article == nil {
		return "No article selected\n"
	}

	s := titleStyle.Render("📰 ARTICLE DETAIL") + "\n\n"

	// Priority and category
	priorityIcon := article.GetPriorityString()
	categoryBadge := m.getCategoryBadge(article.Category)
	s += priorityIcon + " " + categoryBadge + " • " + helpStyle.Render(article.GetAgeString()) + "\n\n"

	// Headline
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

	// Body text with word wrap
	wrappedBody := m.wordWrap(article.Body, 75)
	s += wrappedBody + "\n\n"

	s += strings.Repeat("─", 80) + "\n\n"

	// Metadata
	if article.FactionID != "" {
		s += "Related Faction: " + statsStyle.Render(article.FactionID) + "\n"
	}

	if article.PlayerBased {
		s += successStyle.Render("⭐ This event was caused by your actions") + "\n"
	}

	s += "\n" + renderFooter("ESC: Back to List")

	return s
}

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
