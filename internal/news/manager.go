// Package news provides news feed management and generation.
//
// This package handles:
// - News article management and storage
// - Event-based news generation
// - News filtering and sorting
// - Integration with player actions
// - Automatic news expiration
//
// Version: 1.0.0
// Last Updated: 2025-01-07
package news

import (
	"math/rand"
	"sort"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// Manager handles the news feed
type Manager struct {
	articles           []*models.NewsArticle
	lastRandomNewsTime time.Time
	randomNewsInterval time.Duration
}

// NewManager creates a new news manager
//
// Returns:
//   - Pointer to new Manager
func NewManager() *Manager {
	return &Manager{
		articles:           []*models.NewsArticle{},
		lastRandomNewsTime: time.Now(),
		randomNewsInterval: 30 * time.Minute, // Generate random news every 30 minutes
	}
}

// AddArticle adds a news article to the feed
//
// Parameters:
//   - article: Article to add
func (m *Manager) AddArticle(article *models.NewsArticle) {
	if article == nil {
		return
	}
	m.articles = append(m.articles, article)
	m.pruneExpiredArticles()
}

// GetRecentArticles returns recent news articles
//
// Parameters:
//   - count: Maximum number of articles to return
//   - category: Optional category filter (empty string for all)
//
// Returns:
//   - Slice of recent articles, sorted by creation time (newest first)
func (m *Manager) GetRecentArticles(count int, category models.NewsCategory) []*models.NewsArticle {
	m.pruneExpiredArticles()

	// Filter by category if specified
	filtered := []*models.NewsArticle{}
	for _, article := range m.articles {
		if category == "" || article.Category == category {
			filtered = append(filtered, article)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	// Return up to count articles
	if len(filtered) > count {
		filtered = filtered[:count]
	}

	return filtered
}

// GetArticlesByPriority returns articles filtered by minimum priority
//
// Parameters:
//   - minPriority: Minimum priority level
//
// Returns:
//   - Slice of articles meeting priority threshold
func (m *Manager) GetArticlesByPriority(minPriority models.NewsPriority) []*models.NewsArticle {
	m.pruneExpiredArticles()

	filtered := []*models.NewsArticle{}
	for _, article := range m.articles {
		if article.Priority >= minPriority {
			filtered = append(filtered, article)
		}
	}

	// Sort by priority (highest first), then by time
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority > filtered[j].Priority
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return filtered
}

// GetArticleCount returns the current number of active articles
//
// Returns:
//   - Count of non-expired articles
func (m *Manager) GetArticleCount() int {
	m.pruneExpiredArticles()
	return len(m.articles)
}

// pruneExpiredArticles removes expired articles from the feed
func (m *Manager) pruneExpiredArticles() {
	active := []*models.NewsArticle{}
	for _, article := range m.articles {
		if !article.IsExpired() {
			active = append(active, article)
		}
	}
	m.articles = active
}

// Update checks if random news should be generated
//
// Returns:
//   - New random article if generated, nil otherwise
func (m *Manager) Update() *models.NewsArticle {
	// Check if it's time for random news
	if time.Since(m.lastRandomNewsTime) >= m.randomNewsInterval {
		m.lastRandomNewsTime = time.Now()

		// 50% chance to generate random news
		if rand.Float64() < 0.5 {
			article := models.GenerateRandomNews()
			m.AddArticle(article)
			return article
		}
	}

	return nil
}

// OnPlayerCombat handles combat-related news generation
//
// Parameters:
//   - playerName: Name of player
//   - enemyCount: Number of enemies defeated
//   - systemName: System where combat occurred
//   - factionID: Faction of enemies
//
// Returns:
//   - Generated article if significant enough, nil otherwise
func (m *Manager) OnPlayerCombat(playerName string, enemyCount int, systemName, factionID string) *models.NewsArticle {
	// Only generate news for significant battles
	if enemyCount < 3 {
		return nil
	}

	article := models.GenerateCombatNews(playerName, enemyCount, systemName, factionID)
	m.AddArticle(article)
	return article
}

// OnPlayerTrade handles trade-related news generation
//
// Parameters:
//   - playerName: Name of player
//   - profit: Profit from trade
//   - commodity: Commodity traded
//   - systemName: System where trade occurred
//
// Returns:
//   - Generated article if significant enough, nil otherwise
func (m *Manager) OnPlayerTrade(playerName string, profit int64, commodity, systemName string) *models.NewsArticle {
	article := models.GenerateTradeNews(playerName, profit, commodity, systemName)
	if article != nil {
		m.AddArticle(article)
	}
	return article
}

// OnPlayerAchievement handles achievement-related news generation
//
// Parameters:
//   - playerName: Name of player
//   - achievement: Achievement that was unlocked
//
// Returns:
//   - Generated article if significant enough, nil otherwise
func (m *Manager) OnPlayerAchievement(playerName string, achievement *models.Achievement) *models.NewsArticle {
	if achievement == nil {
		return nil
	}

	article := models.GenerateAchievementNews(playerName, achievement.Title, achievement.Rarity)
	if article != nil {
		m.AddArticle(article)
	}
	return article
}

// OnFactionRelationChange handles faction relation news generation
//
// Parameters:
//   - faction1: First faction name
//   - faction2: Second faction name
//   - isPositive: Whether the change is positive
//
// Returns:
//   - Generated article
func (m *Manager) OnFactionRelationChange(faction1, faction2 string, isPositive bool) *models.NewsArticle {
	article := models.GenerateFactionNews(faction1, faction2, isPositive)
	m.AddArticle(article)
	return article
}

// GetBreakingNews returns only critical priority news
//
// Returns:
//   - Slice of critical priority articles
func (m *Manager) GetBreakingNews() []*models.NewsArticle {
	return m.GetArticlesByPriority(models.NewsPriorityCritical)
}

// GetPlayerNews returns news articles related to player actions
//
// Parameters:
//   - count: Maximum number of articles to return
//
// Returns:
//   - Slice of player-generated articles
func (m *Manager) GetPlayerNews(count int) []*models.NewsArticle {
	m.pruneExpiredArticles()

	playerArticles := []*models.NewsArticle{}
	for _, article := range m.articles {
		if article.PlayerBased {
			playerArticles = append(playerArticles, article)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(playerArticles, func(i, j int) bool {
		return playerArticles[i].CreatedAt.After(playerArticles[j].CreatedAt)
	})

	if len(playerArticles) > count {
		playerArticles = playerArticles[:count]
	}

	return playerArticles
}

// ClearOldNews removes all articles older than a specified duration
//
// Parameters:
//   - maxAge: Maximum age for articles to keep
func (m *Manager) ClearOldNews(maxAge time.Duration) {
	cutoffTime := time.Now().Add(-maxAge)
	active := []*models.NewsArticle{}

	for _, article := range m.articles {
		if article.CreatedAt.After(cutoffTime) {
			active = append(active, article)
		}
	}

	m.articles = active
}

// GetCategoryCount returns the number of articles in each category
//
// Returns:
//   - Map of category to article count
func (m *Manager) GetCategoryCount() map[models.NewsCategory]int {
	m.pruneExpiredArticles()

	counts := make(map[models.NewsCategory]int)
	for _, article := range m.articles {
		counts[article.Category]++
	}

	return counts
}

// GenerateInitialNews creates starting news articles for a new game
func (m *Manager) GenerateInitialNews() {
	// Generate some initial background news
	for i := 0; i < 5; i++ {
		article := models.GenerateRandomNews()
		// Backdate articles slightly
		article.CreatedAt = time.Now().Add(-time.Duration(i) * time.Hour)
		m.AddArticle(article)
	}

	// Add a faction relations article
	factions := []string{
		"United Earth Federation",
		"Rigel Outer Marches",
		"Free Worlds Alliance",
		"Auroran Empire",
	}

	if len(factions) >= 2 {
		article := models.GenerateFactionNews(factions[0], factions[1], rand.Float64() < 0.5)
		article.CreatedAt = time.Now().Add(-2 * time.Hour)
		m.AddArticle(article)
	}
}
