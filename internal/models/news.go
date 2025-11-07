// Package models - News and events system definitions
//
// This file defines the dynamic news and events system that generates
// content based on player actions, faction relations, and universe state.
//
// Version: 1.0.0
// Last Updated: 2025-01-07
package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// NewsCategory represents the type of news article
type NewsCategory string

const (
	NewsCategoryCombat      NewsCategory = "combat"       // Combat-related news
	NewsCategoryEconomic    NewsCategory = "economic"     // Trade and market news
	NewsCategoryPolitical   NewsCategory = "political"    // Faction relations
	NewsCategoryExploration NewsCategory = "exploration"  // Discovery news
	NewsCategoryMilitary    NewsCategory = "military"     // Military actions
	NewsCategoryCriminal    NewsCategory = "criminal"     // Crime and piracy
	NewsCategoryGeneral     NewsCategory = "general"      // General announcements
	NewsCategoryAchievement NewsCategory = "achievement"  // Player achievements
)

// NewsPriority represents the importance of a news item
type NewsPriority int

const (
	NewsPriorityLow      NewsPriority = 1  // Minor events
	NewsPriorityMedium   NewsPriority = 2  // Noteworthy events
	NewsPriorityHigh     NewsPriority = 3  // Important events
	NewsPriorityCritical NewsPriority = 4  // Major universe events
)

// NewsArticle represents a single news item
type NewsArticle struct {
	ID          uuid.UUID    `json:"id"`
	Category    NewsCategory `json:"category"`
	Priority    NewsPriority `json:"priority"`
	Headline    string       `json:"headline"`
	Body        string       `json:"body"`
	SystemID    *uuid.UUID   `json:"system_id,omitempty"`    // Location of event
	FactionID   string       `json:"faction_id,omitempty"`   // Related faction
	CreatedAt   time.Time    `json:"created_at"`
	ExpiresAt   time.Time    `json:"expires_at"`             // When news becomes old
	PlayerBased bool         `json:"player_based"`           // Generated from player actions
}

// NewsEvent represents a universe event that can generate news
type NewsEvent struct {
	Type        string                 `json:"type"`
	SystemID    uuid.UUID              `json:"system_id,omitempty"`
	FactionID   string                 `json:"faction_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewNewsArticle creates a new news article
//
// Parameters:
//   - category: Category of the news
//   - priority: Priority level
//   - headline: News headline
//   - body: Full news text
//
// Returns:
//   - Pointer to new NewsArticle
func NewNewsArticle(category NewsCategory, priority NewsPriority, headline, body string) *NewsArticle {
	now := time.Now()

	// News expires after different durations based on priority
	expirationHours := 24
	switch priority {
	case NewsPriorityLow:
		expirationHours = 6
	case NewsPriorityMedium:
		expirationHours = 12
	case NewsPriorityHigh:
		expirationHours = 24
	case NewsPriorityCritical:
		expirationHours = 48
	}

	return &NewsArticle{
		ID:        uuid.New(),
		Category:  category,
		Priority:  priority,
		Headline:  headline,
		Body:      body,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(expirationHours) * time.Hour),
	}
}

// IsExpired checks if the news article has expired
//
// Returns:
//   - true if article is expired
func (n *NewsArticle) IsExpired() bool {
	return time.Now().After(n.ExpiresAt)
}

// GetAgeString returns a human-readable age string
//
// Returns:
//   - Age string like "2 hours ago"
func (n *NewsArticle) GetAgeString() string {
	duration := time.Since(n.CreatedAt)

	if duration.Hours() < 1 {
		minutes := int(duration.Minutes())
		if minutes <= 1 {
			return "Just now"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration.Hours() < 24 {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// GetPriorityString returns a display string for priority
//
// Returns:
//   - Priority string with icon
func (n *NewsArticle) GetPriorityString() string {
	switch n.Priority {
	case NewsPriorityLow:
		return "📰"
	case NewsPriorityMedium:
		return "📢"
	case NewsPriorityHigh:
		return "⚠️"
	case NewsPriorityCritical:
		return "🚨"
	default:
		return "📰"
	}
}

// GenerateCombatNews generates news from combat events
//
// Parameters:
//   - playerName: Name of player involved
//   - enemyCount: Number of enemies defeated
//   - systemName: Name of system where combat occurred
//   - factionID: Faction of defeated enemies
//
// Returns:
//   - NewsArticle about the combat
func GenerateCombatNews(playerName string, enemyCount int, systemName, factionID string) *NewsArticle {
	headlines := []string{
		fmt.Sprintf("Pilot %s Defeats %d Hostiles in %s", playerName, enemyCount, systemName),
		fmt.Sprintf("Space Battle in %s: %d Ships Destroyed", systemName, enemyCount),
		fmt.Sprintf("%s System Sees Increased Combat Activity", systemName),
	}

	bodies := []string{
		fmt.Sprintf("Independent pilot %s successfully engaged and destroyed %d hostile vessels in the %s system. Local authorities have expressed gratitude for the pilot's actions in keeping space lanes safe.", playerName, enemyCount, systemName),
		fmt.Sprintf("A fierce space battle erupted in %s today, resulting in %d confirmed ship destructions. Pilot %s was instrumental in the engagement, which lasted several hours.", systemName, enemyCount, playerName),
		fmt.Sprintf("The %s system has become a hotbed of military activity following today's engagement where %d vessels were destroyed. Pilot %s has been credited with defending civilian traffic in the area.", systemName, enemyCount, playerName),
	}

	idx := rand.Intn(len(headlines))
	priority := NewsPriorityLow
	if enemyCount >= 5 {
		priority = NewsPriorityMedium
	}
	if enemyCount >= 10 {
		priority = NewsPriorityHigh
	}

	article := NewNewsArticle(NewsCategoryCombat, priority, headlines[idx], bodies[idx])
	article.FactionID = factionID
	article.PlayerBased = true
	return article
}

// GenerateTradeNews generates news from significant trade events
//
// Parameters:
//   - playerName: Name of player involved
//   - profit: Profit from trade
//   - commodity: Commodity traded
//   - systemName: System where trade occurred
//
// Returns:
//   - NewsArticle about the trade
func GenerateTradeNews(playerName string, profit int64, commodity, systemName string) *NewsArticle {
	if profit < 10000 {
		return nil // Only generate news for significant trades
	}

	headlines := []string{
		fmt.Sprintf("Major %s Trade in %s", commodity, systemName),
		fmt.Sprintf("%s Market Sees Record Transaction", systemName),
		fmt.Sprintf("Independent Trader Makes Fortune in %s", systemName),
	}

	bodies := []string{
		fmt.Sprintf("The %s market in %s experienced a significant transaction today involving large quantities of %s. Market analysts project this could influence regional prices. Independent trader %s is reported to have made considerable profit on the deal.", systemName, systemName, commodity, playerName),
		fmt.Sprintf("Local merchants in %s report unusual market activity involving %s today. The transaction, valued at over %d credits, was conducted by trader %s and has caught the attention of commodity speculators.", systemName, commodity, profit, playerName),
	}

	idx := rand.Intn(len(headlines))
	priority := NewsPriorityLow
	if profit >= 50000 {
		priority = NewsPriorityMedium
	}

	article := NewNewsArticle(NewsCategoryEconomic, priority, headlines[idx], bodies[idx])
	article.PlayerBased = true
	return article
}

// GenerateAchievementNews generates news from major player achievements
//
// Parameters:
//   - playerName: Name of player
//   - achievementTitle: Title of achievement unlocked
//   - achievementRarity: Rarity of achievement
//
// Returns:
//   - NewsArticle about the achievement
func GenerateAchievementNews(playerName string, achievementTitle string, achievementRarity AchievementRarity) *NewsArticle {
	// Only generate news for rare+ achievements
	if achievementRarity != AchievementRarityRare &&
	   achievementRarity != AchievementRarityEpic &&
	   achievementRarity != AchievementRarityLegendary {
		return nil
	}

	headlines := []string{
		fmt.Sprintf("Pilot %s Achieves Rare Feat: %s", playerName, achievementTitle),
		fmt.Sprintf("%s: Notable Achievement by Independent Pilot", achievementTitle),
		fmt.Sprintf("Pilot %s Makes Headlines with Impressive Achievement", playerName),
	}

	bodies := []string{
		fmt.Sprintf("Independent pilot %s has achieved a rare distinction in space exploration: %s. This achievement is recognized by only a small percentage of pilots in known space.", playerName, achievementTitle),
		fmt.Sprintf("The galactic community takes note as pilot %s accomplishes %s. This notable achievement demonstrates exceptional skill and dedication to their craft.", playerName, achievementTitle),
	}

	idx := rand.Intn(len(headlines))

	priority := NewsPriorityMedium
	if achievementRarity == AchievementRarityEpic {
		priority = NewsPriorityHigh
	}
	if achievementRarity == AchievementRarityLegendary {
		priority = NewsPriorityCritical
	}

	article := NewNewsArticle(NewsCategoryAchievement, priority, headlines[idx], bodies[idx])
	article.PlayerBased = true
	return article
}

// GenerateFactionNews generates news about faction relations
//
// Parameters:
//   - factionName1: First faction
//   - factionName2: Second faction
//   - isPositive: Whether the news is positive
//
// Returns:
//   - NewsArticle about faction relations
func GenerateFactionNews(factionName1, factionName2 string, isPositive bool) *NewsArticle {
	var headlines []string
	var bodies []string

	if isPositive {
		headlines = []string{
			fmt.Sprintf("%s and %s Announce Trade Agreement", factionName1, factionName2),
			fmt.Sprintf("Diplomatic Breakthrough Between %s and %s", factionName1, factionName2),
			fmt.Sprintf("%s Welcomes %s Delegation", factionName1, factionName2),
		}
		bodies = []string{
			fmt.Sprintf("In a surprising turn of events, %s and %s have announced a new trade agreement that promises to benefit both regions. Analysts predict increased stability in affiliated systems.", factionName1, factionName2),
			fmt.Sprintf("After weeks of negotiations, %s and %s have reached a diplomatic accord. The agreement is expected to ease tensions that have existed between the factions for years.", factionName1, factionName2),
		}
	} else {
		headlines = []string{
			fmt.Sprintf("Tensions Rise Between %s and %s", factionName1, factionName2),
			fmt.Sprintf("%s Accuses %s of Border Violations", factionName1, factionName2),
			fmt.Sprintf("Diplomatic Crisis: %s and %s Relations Deteriorate", factionName1, factionName2),
		}
		bodies = []string{
			fmt.Sprintf("Relations between %s and %s have reached a new low following recent incidents. Military analysts warn of potential conflict if diplomatic channels cannot resolve the dispute.", factionName1, factionName2),
			fmt.Sprintf("A diplomatic crisis is unfolding as %s formally accuses %s of aggressive actions. Independent pilots are advised to exercise caution when traveling through systems controlled by either faction.", factionName1, factionName2),
		}
	}

	idx := rand.Intn(len(headlines))
	article := NewNewsArticle(NewsCategoryPolitical, NewsPriorityMedium, headlines[idx], bodies[idx])
	return article
}

// GenerateRandomNews generates random universe events
//
// Returns:
//   - NewsArticle with random event
func GenerateRandomNews() *NewsArticle {
	events := []struct {
		category NewsCategory
		headline string
		body     string
		priority NewsPriority
	}{
		{
			NewsCategoryExploration,
			"New Jump Route Discovered",
			"Explorers have charted a previously unknown hyperspace corridor connecting two remote systems. The discovery is expected to open new trade opportunities and reduce travel time for independent pilots.",
			NewsPriorityMedium,
		},
		{
			NewsCategoryEconomic,
			"Commodity Prices Fluctuate Across Core Systems",
			"Market analysts report unusual volatility in commodity prices this week. Traders are advised to monitor market conditions closely for optimal trading opportunities.",
			NewsPriorityLow,
		},
		{
			NewsCategoryGeneral,
			"Pirate Activity on the Rise",
			"Security forces report increased pirate activity in outer rim systems. Independent pilots are advised to travel in convoys when possible and to avoid carrying valuable cargo through unpatrolled space.",
			NewsPriorityMedium,
		},
		{
			NewsCategoryMilitary,
			"Naval Exercises Scheduled in Multiple Systems",
			"Military officials announce large-scale naval exercises will be conducted over the next week. Civilian traffic may experience delays in affected systems.",
			NewsPriorityLow,
		},
		{
			NewsCategoryCriminal,
			"Notorious Pirate Gang Spotted in Outer Systems",
			"Law enforcement agencies warn that a well-organized pirate group has been operating in several outer systems. A bounty has been placed on the gang's leaders.",
			NewsPriorityHigh,
		},
		{
			NewsCategoryGeneral,
			"Merchant Guild Announces Safety Initiative",
			"The Free Traders Guild has announced new safety measures for merchant vessels, including improved convoy coordination and enhanced communication protocols.",
			NewsPriorityLow,
		},
		{
			NewsCategoryEconomic,
			"Rare Minerals Discovered in Asteroid Belt",
			"Prospectors have located significant deposits of rare minerals in a previously unexplored asteroid belt. Mining rights are expected to be auctioned soon.",
			NewsPriorityMedium,
		},
		{
			NewsCategoryGeneral,
			"Shipyard Announces New Vessel Class",
			"A major shipyard has revealed plans for a new class of multi-role vessel designed for independent pilots. Specifications and pricing are expected to be released next quarter.",
			NewsPriorityLow,
		},
	}

	event := events[rand.Intn(len(events))]
	return NewNewsArticle(event.category, event.priority, event.headline, event.body)
}
