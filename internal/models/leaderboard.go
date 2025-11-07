// File: internal/models/leaderboard.go
// Project: Terminal Velocity
// Description: Leaderboard models and ranking calculations for competitive player statistics
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// LeaderboardCategory represents different types of leaderboards
type LeaderboardCategory string

const (
	LeaderboardCombat      LeaderboardCategory = "combat"       // Combat prowess rankings
	LeaderboardTrading     LeaderboardCategory = "trading"      // Trading success rankings
	LeaderboardExploration LeaderboardCategory = "exploration"  // Exploration rankings
	LeaderboardWealth      LeaderboardCategory = "wealth"       // Total credits rankings
	LeaderboardReputation  LeaderboardCategory = "reputation"   // Overall reputation rankings
	LeaderboardMissions    LeaderboardCategory = "missions"     // Mission completion rankings
	LeaderboardOverall     LeaderboardCategory = "overall"      // Overall ranking (composite)
)

// LeaderboardEntry represents a single entry in a leaderboard
//
// Each entry tracks a player's performance in a specific category along with
// their current rank and the timestamp of when this ranking was calculated.
type LeaderboardEntry struct {
	ID         uuid.UUID           `json:"id"`          // Unique entry identifier
	PlayerID   uuid.UUID           `json:"player_id"`   // Player this entry belongs to
	PlayerName string              `json:"player_name"` // Player's display name
	Category   LeaderboardCategory `json:"category"`    // Which leaderboard this entry is for
	Rank       int                 `json:"rank"`        // Current rank (1 = best)
	Score      int64               `json:"score"`       // Numeric score for this category
	UpdatedAt  time.Time           `json:"updated_at"`  // When this entry was last updated

	// Category-specific details for display
	Details map[string]interface{} `json:"details,omitempty"` // Additional context (kills, trades, etc.)
}

// LeaderboardSnapshot represents a full leaderboard at a point in time
type LeaderboardSnapshot struct {
	Category  LeaderboardCategory `json:"category"`   // Which leaderboard this is
	Entries   []*LeaderboardEntry `json:"entries"`    // Ranked list of entries
	UpdatedAt time.Time           `json:"updated_at"` // When this snapshot was taken
	TotalPlayers int              `json:"total_players"` // Total number of ranked players
}

// NewLeaderboardEntry creates a new leaderboard entry for a player
func NewLeaderboardEntry(playerID uuid.UUID, playerName string, category LeaderboardCategory, score int64) *LeaderboardEntry {
	return &LeaderboardEntry{
		ID:         uuid.New(),
		PlayerID:   playerID,
		PlayerName: playerName,
		Category:   category,
		Score:      score,
		UpdatedAt:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// CalculateCombatScore calculates a player's combat leaderboard score
//
// Combat score is based primarily on kills but also considers combat rating
// to reward skill progression.
//
// Formula: (kills * 100) + (combat_rating * 10)
func CalculateCombatScore(player *Player) int64 {
	return (int64(player.TotalKills) * 100) + (int64(player.CombatRating) * 10)
}

// CalculateTradingScore calculates a player's trading leaderboard score
//
// Trading score is based on total profit and number of trades to reward
// both volume and profitability.
//
// Formula: trade_profit + (total_trades * 100)
func CalculateTradingScore(player *Player) int64 {
	return player.TradeProfit + (int64(player.TotalTrades) * 100)
}

// CalculateExplorationScore calculates a player's exploration leaderboard score
//
// Exploration score rewards discovering new systems and making jumps,
// with systems weighted more heavily than jumps.
//
// Formula: (systems_visited * 1000) + (total_jumps * 10)
func CalculateExplorationScore(player *Player) int64 {
	return (int64(player.SystemsVisited) * 1000) + (int64(player.TotalJumps) * 10)
}

// CalculateWealthScore calculates a player's wealth leaderboard score
//
// Wealth score is simply the player's current credits.
func CalculateWealthScore(player *Player) int64 {
	return player.Credits
}

// CalculateReputationScore calculates a player's reputation leaderboard score
//
// Reputation score is the sum of all faction reputations (positive values only)
func CalculateReputationScore(player *Player) int64 {
	totalRep := int64(0)
	for _, rep := range player.Reputation {
		if rep > 0 {
			totalRep += int64(rep)
		}
	}
	return totalRep
}

// CalculateMissionScore calculates a player's mission leaderboard score
//
// Mission score rewards completions heavily while penalizing failures moderately.
//
// Formula: (missions_completed * 1000) - (missions_failed * 100)
func CalculateMissionScore(player *Player) int64 {
	score := (int64(player.MissionsCompleted) * 1000) - (int64(player.MissionsFailed) * 100)
	if score < 0 {
		score = 0
	}
	return score
}

// CalculateOverallScore calculates a player's overall leaderboard score
//
// Overall score is a weighted composite of all other scores to determine
// the best all-around players.
//
// Weights:
// - Combat: 20%
// - Trading: 25%
// - Exploration: 15%
// - Wealth: 20%
// - Reputation: 10%
// - Missions: 10%
func CalculateOverallScore(player *Player) int64 {
	combat := float64(CalculateCombatScore(player)) * 0.20
	trading := float64(CalculateTradingScore(player)) * 0.25
	exploration := float64(CalculateExplorationScore(player)) * 0.15
	wealth := float64(CalculateWealthScore(player)) * 0.20
	reputation := float64(CalculateReputationScore(player)) * 0.10
	missions := float64(CalculateMissionScore(player)) * 0.10

	return int64(combat + trading + exploration + wealth + reputation + missions)
}

// GetScoreForCategory calculates the appropriate score for a given leaderboard category
func GetScoreForCategory(player *Player, category LeaderboardCategory) int64 {
	switch category {
	case LeaderboardCombat:
		return CalculateCombatScore(player)
	case LeaderboardTrading:
		return CalculateTradingScore(player)
	case LeaderboardExploration:
		return CalculateExplorationScore(player)
	case LeaderboardWealth:
		return CalculateWealthScore(player)
	case LeaderboardReputation:
		return CalculateReputationScore(player)
	case LeaderboardMissions:
		return CalculateMissionScore(player)
	case LeaderboardOverall:
		return CalculateOverallScore(player)
	default:
		return 0
	}
}

// PopulateDetails fills the Details map with category-specific information
func (e *LeaderboardEntry) PopulateDetails(player *Player) {
	switch e.Category {
	case LeaderboardCombat:
		e.Details["kills"] = player.TotalKills
		e.Details["rating"] = player.CombatRating
		e.Details["rank_title"] = player.GetCombatRankTitle()

	case LeaderboardTrading:
		e.Details["trades"] = player.TotalTrades
		e.Details["profit"] = player.TradeProfit
		e.Details["highest_profit"] = player.HighestProfit
		e.Details["rating"] = player.TradingRating

	case LeaderboardExploration:
		e.Details["systems"] = player.SystemsVisited
		e.Details["jumps"] = player.TotalJumps
		e.Details["rating"] = player.ExplorationRating

	case LeaderboardWealth:
		e.Details["credits"] = player.Credits

	case LeaderboardReputation:
		e.Details["total_reputation"] = e.Score
		e.Details["faction_count"] = len(player.Reputation)

	case LeaderboardMissions:
		e.Details["completed"] = player.MissionsCompleted
		e.Details["failed"] = player.MissionsFailed

	case LeaderboardOverall:
		e.Details["combat_rating"] = player.CombatRating
		e.Details["trading_rating"] = player.TradingRating
		e.Details["exploration_rating"] = player.ExplorationRating
		e.Details["rank_title"] = player.GetOverallRank()
	}
}

// GetCategoryDisplayName returns a human-readable name for a leaderboard category
func GetCategoryDisplayName(category LeaderboardCategory) string {
	names := map[LeaderboardCategory]string{
		LeaderboardCombat:      "Combat Masters",
		LeaderboardTrading:     "Trade Moguls",
		LeaderboardExploration: "Explorers",
		LeaderboardWealth:      "Wealthiest",
		LeaderboardReputation:  "Most Reputable",
		LeaderboardMissions:    "Mission Experts",
		LeaderboardOverall:     "Overall Rankings",
	}
	return names[category]
}

// GetCategoryIcon returns an emoji icon for a leaderboard category
func GetCategoryIcon(category LeaderboardCategory) string {
	icons := map[LeaderboardCategory]string{
		LeaderboardCombat:      "⚔️",
		LeaderboardTrading:     "💰",
		LeaderboardExploration: "🚀",
		LeaderboardWealth:      "💎",
		LeaderboardReputation:  "🏆",
		LeaderboardMissions:    "📋",
		LeaderboardOverall:     "👑",
	}
	return icons[category]
}

// GetRankMedal returns a medal emoji for top 3 ranks
func GetRankMedal(rank int) string {
	switch rank {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return ""
	}
}
