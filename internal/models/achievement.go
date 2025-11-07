// Package models - Achievement system definitions
//
// This file defines the achievement system for tracking player accomplishments.
// Achievements are unlocked based on specific criteria across all gameplay categories.
//
// Version: 1.0.0
// Last Updated: 2025-01-07
package models

import (
	"time"

	"github.com/google/uuid"
)

// AchievementCategory represents the type of achievement
type AchievementCategory string

const (
	AchievementCategoryCombat      AchievementCategory = "combat"
	AchievementCategoryTrading     AchievementCategory = "trading"
	AchievementCategoryExploration AchievementCategory = "exploration"
	AchievementCategoryMissions    AchievementCategory = "missions"
	AchievementCategoryWealth      AchievementCategory = "wealth"
	AchievementCategoryShips       AchievementCategory = "ships"
	AchievementCategorySpecial     AchievementCategory = "special"
)

// AchievementRarity represents how difficult an achievement is to obtain
type AchievementRarity string

const (
	AchievementRarityCommon    AchievementRarity = "common"    // Most players will earn
	AchievementRarityUncommon  AchievementRarity = "uncommon"  // Requires some effort
	AchievementRarityRare      AchievementRarity = "rare"      // Significant effort
	AchievementRarityEpic      AchievementRarity = "epic"      // Challenging goals
	AchievementRarityLegendary AchievementRarity = "legendary" // Extremely difficult
)

// Achievement represents a single achievement definition
type Achievement struct {
	ID          string              `json:"id"`          // Unique identifier (e.g., "first_kill")
	Title       string              `json:"title"`       // Display name
	Description string              `json:"description"` // What the player must do
	Category    AchievementCategory `json:"category"`    // Category of achievement
	Rarity      AchievementRarity   `json:"rarity"`      // How rare/difficult
	Icon        string              `json:"icon"`        // Unicode icon/emoji
	Points      int                 `json:"points"`      // Achievement points value
	Hidden      bool                `json:"hidden"`      // Hidden until unlocked

	// Criteria for unlocking (one of these will be checked)
	RequireKills           int   `json:"require_kills,omitempty"`            // Total kills needed
	RequireTrades          int   `json:"require_trades,omitempty"`           // Total trades needed
	RequireCredits         int64 `json:"require_credits,omitempty"`          // Credits needed
	RequireProfit          int64 `json:"require_profit,omitempty"`           // Total profit needed
	RequireSystemsVisited  int   `json:"require_systems_visited,omitempty"`  // Systems to visit
	RequireJumps           int   `json:"require_jumps,omitempty"`            // Jumps needed
	RequireMissions        int   `json:"require_missions,omitempty"`         // Missions to complete
	RequireCombatRating    int   `json:"require_combat_rating,omitempty"`    // Combat rating needed
	RequireTradingRating   int   `json:"require_trading_rating,omitempty"`   // Trading rating needed
	RequireShipsPurchased  int   `json:"require_ships_purchased,omitempty"`  // Ships purchased
	RequireSpecificShipID  string `json:"require_specific_ship_id,omitempty"` // Specific ship type
}

// PlayerAchievement represents a player's unlock of an achievement
type PlayerAchievement struct {
	ID            uuid.UUID `json:"id"`
	PlayerID      uuid.UUID `json:"player_id"`
	AchievementID string    `json:"achievement_id"` // References Achievement.ID
	UnlockedAt    time.Time `json:"unlocked_at"`
	Progress      int       `json:"progress"` // For tracking partial progress (0-100)
}

// GetAllAchievements returns the complete list of defined achievements
func GetAllAchievements() []*Achievement {
	return []*Achievement{
		// ==================== Combat Achievements ====================
		{
			ID:           "first_blood",
			Title:        "First Blood",
			Description:  "Destroy your first enemy ship",
			Category:     AchievementCategoryCombat,
			Rarity:       AchievementRarityCommon,
			Icon:         "⚔️",
			Points:       10,
			RequireKills: 1,
		},
		{
			ID:           "ace_pilot",
			Title:        "Ace Pilot",
			Description:  "Destroy 10 enemy ships",
			Category:     AchievementCategoryCombat,
			Rarity:       AchievementRarityUncommon,
			Icon:         "✈️",
			Points:       25,
			RequireKills: 10,
		},
		{
			ID:           "veteran_warrior",
			Title:        "Veteran Warrior",
			Description:  "Destroy 50 enemy ships",
			Category:     AchievementCategoryCombat,
			Rarity:       AchievementRarityRare,
			Icon:         "🎖️",
			Points:       50,
			RequireKills: 50,
		},
		{
			ID:           "death_dealer",
			Title:        "Death Dealer",
			Description:  "Destroy 100 enemy ships",
			Category:     AchievementCategoryCombat,
			Rarity:       AchievementRarityEpic,
			Icon:         "💀",
			Points:       100,
			RequireKills: 100,
		},
		{
			ID:           "elite_combat",
			Title:        "Elite Combat Rating",
			Description:  "Achieve Elite combat rating (100)",
			Category:     AchievementCategoryCombat,
			Rarity:       AchievementRarityEpic,
			Icon:         "👑",
			Points:       100,
			RequireCombatRating: 100,
		},

		// ==================== Trading Achievements ====================
		{
			ID:           "first_trade",
			Title:        "First Trade",
			Description:  "Complete your first trade transaction",
			Category:     AchievementCategoryTrading,
			Rarity:       AchievementRarityCommon,
			Icon:         "💰",
			Points:       10,
			RequireTrades: 1,
		},
		{
			ID:           "merchant",
			Title:        "Merchant",
			Description:  "Complete 50 trade transactions",
			Category:     AchievementCategoryTrading,
			Rarity:       AchievementRarityUncommon,
			Icon:         "🏪",
			Points:       25,
			RequireTrades: 50,
		},
		{
			ID:           "trade_magnate",
			Title:        "Trade Magnate",
			Description:  "Complete 200 trade transactions",
			Category:     AchievementCategoryTrading,
			Rarity:       AchievementRarityRare,
			Icon:         "📈",
			Points:       50,
			RequireTrades: 200,
		},
		{
			ID:           "millionaire",
			Title:        "Millionaire",
			Description:  "Accumulate 1,000,000 credits",
			Category:     AchievementCategoryWealth,
			Rarity:       AchievementRarityRare,
			Icon:         "💎",
			Points:       50,
			RequireCredits: 1000000,
		},
		{
			ID:           "profit_master",
			Title:        "Profit Master",
			Description:  "Earn 5,000,000 credits in total trading profit",
			Category:     AchievementCategoryTrading,
			Rarity:       AchievementRarityEpic,
			Icon:         "🤑",
			Points:       100,
			RequireProfit: 5000000,
		},
		{
			ID:           "tycoon",
			Title:        "Tycoon",
			Description:  "Achieve Tycoon trading rating (100)",
			Category:     AchievementCategoryTrading,
			Rarity:       AchievementRarityEpic,
			Icon:         "🏆",
			Points:       100,
			RequireTradingRating: 100,
		},

		// ==================== Exploration Achievements ====================
		{
			ID:           "first_jump",
			Title:        "First Jump",
			Description:  "Make your first hyperspace jump",
			Category:     AchievementCategoryExploration,
			Rarity:       AchievementRarityCommon,
			Icon:         "🚀",
			Points:       10,
			RequireJumps: 1,
		},
		{
			ID:           "explorer",
			Title:        "Explorer",
			Description:  "Visit 25 different star systems",
			Category:     AchievementCategoryExploration,
			Rarity:       AchievementRarityUncommon,
			Icon:         "🗺️",
			Points:       25,
			RequireSystemsVisited: 25,
		},
		{
			ID:           "pathfinder",
			Title:        "Pathfinder",
			Description:  "Achieve Pathfinder exploration rating (100)",
			Category:     AchievementCategoryExploration,
			Rarity:       AchievementRarityEpic,
			Icon:         "🧭",
			Points:       100,
			RequireSystemsVisited: 50,
		},
		{
			ID:           "space_nomad",
			Title:        "Space Nomad",
			Description:  "Make 500 hyperspace jumps",
			Category:     AchievementCategoryExploration,
			Rarity:       AchievementRarityRare,
			Icon:         "🌌",
			Points:       50,
			RequireJumps: 500,
		},

		// ==================== Mission Achievements ====================
		{
			ID:           "first_mission",
			Title:        "First Mission",
			Description:  "Complete your first mission",
			Category:     AchievementCategoryMissions,
			Rarity:       AchievementRarityCommon,
			Icon:         "📋",
			Points:       10,
			RequireMissions: 1,
		},
		{
			ID:           "mission_runner",
			Title:        "Mission Runner",
			Description:  "Complete 25 missions",
			Category:     AchievementCategoryMissions,
			Rarity:       AchievementRarityUncommon,
			Icon:         "📦",
			Points:       25,
			RequireMissions: 25,
		},
		{
			ID:           "contractor",
			Title:        "Contractor",
			Description:  "Complete 100 missions",
			Category:     AchievementCategoryMissions,
			Rarity:       AchievementRarityRare,
			Icon:         "📜",
			Points:       50,
			RequireMissions: 100,
		},
		{
			ID:           "legendary_agent",
			Title:        "Legendary Agent",
			Description:  "Complete 500 missions",
			Category:     AchievementCategoryMissions,
			Rarity:       AchievementRarityLegendary,
			Icon:         "🌟",
			Points:       200,
			RequireMissions: 500,
		},

		// ==================== Ship Achievements ====================
		{
			ID:           "ship_collector",
			Title:        "Ship Collector",
			Description:  "Purchase 5 different ships",
			Category:     AchievementCategoryShips,
			Rarity:       AchievementRarityUncommon,
			Icon:         "🚢",
			Points:       25,
			RequireShipsPurchased: 5,
		},
		{
			ID:           "battleship_captain",
			Title:        "Battleship Captain",
			Description:  "Purchase a Battleship",
			Category:     AchievementCategoryShips,
			Rarity:       AchievementRarityEpic,
			Icon:         "⚓",
			Points:       100,
			RequireSpecificShipID: "battleship",
			Hidden:       true,
		},

		// ==================== Special Achievements ====================
		{
			ID:          "early_adopter",
			Title:       "Early Adopter",
			Description: "Play during the beta period",
			Category:    AchievementCategorySpecial,
			Rarity:      AchievementRarityRare,
			Icon:        "🎮",
			Points:      50,
			Hidden:      true,
		},
		{
			ID:          "jack_of_all_trades",
			Title:       "Jack of All Trades",
			Description: "Reach rating 50 in combat, trading, and exploration",
			Category:    AchievementCategorySpecial,
			Rarity:      AchievementRarityEpic,
			Icon:        "⭐",
			Points:      100,
			Hidden:      true,
		},
	}
}

// IsUnlocked checks if the achievement criteria are met by the player
//
// Parameters:
//   - player: Player to check against
//
// Returns:
//   - true if achievement should be unlocked
func (a *Achievement) IsUnlocked(player *Player) bool {
	// Check each criterion (only one needs to be set per achievement)
	if a.RequireKills > 0 && player.TotalKills >= a.RequireKills {
		return true
	}
	if a.RequireTrades > 0 && player.TotalTrades >= a.RequireTrades {
		return true
	}
	if a.RequireCredits > 0 && player.Credits >= a.RequireCredits {
		return true
	}
	if a.RequireProfit > 0 && player.TradeProfit >= a.RequireProfit {
		return true
	}
	if a.RequireSystemsVisited > 0 && player.SystemsVisited >= a.RequireSystemsVisited {
		return true
	}
	if a.RequireJumps > 0 && player.TotalJumps >= a.RequireJumps {
		return true
	}
	if a.RequireMissions > 0 && player.MissionsCompleted >= a.RequireMissions {
		return true
	}
	if a.RequireCombatRating > 0 && player.CombatRating >= a.RequireCombatRating {
		return true
	}
	if a.RequireTradingRating > 0 && player.TradingRating >= a.RequireTradingRating {
		return true
	}

	// Special multi-criteria achievements
	if a.ID == "jack_of_all_trades" {
		return player.CombatRating >= 50 && player.TradingRating >= 50 && player.ExplorationRating >= 50
	}

	return false
}

// CalculateProgress returns percentage progress toward unlocking (0-100)
//
// Parameters:
//   - player: Player to check progress for
//
// Returns:
//   - Progress percentage (0-100)
func (a *Achievement) CalculateProgress(player *Player) int {
	var current, required int

	// Determine which criterion to check
	if a.RequireKills > 0 {
		current = player.TotalKills
		required = a.RequireKills
	} else if a.RequireTrades > 0 {
		current = player.TotalTrades
		required = a.RequireTrades
	} else if a.RequireCredits > 0 {
		current = int(player.Credits)
		required = int(a.RequireCredits)
	} else if a.RequireProfit > 0 {
		current = int(player.TradeProfit)
		required = int(a.RequireProfit)
	} else if a.RequireSystemsVisited > 0 {
		current = player.SystemsVisited
		required = a.RequireSystemsVisited
	} else if a.RequireJumps > 0 {
		current = player.TotalJumps
		required = a.RequireJumps
	} else if a.RequireMissions > 0 {
		current = player.MissionsCompleted
		required = a.RequireMissions
	} else if a.RequireCombatRating > 0 {
		current = player.CombatRating
		required = a.RequireCombatRating
	} else if a.RequireTradingRating > 0 {
		current = player.TradingRating
		required = a.RequireTradingRating
	} else {
		return 0
	}

	if required == 0 {
		return 0
	}

	progress := (current * 100) / required
	if progress > 100 {
		progress = 100
	}

	return progress
}

// GetRarityColor returns a color code for the achievement rarity
func (a *Achievement) GetRarityColor() string {
	switch a.Rarity {
	case AchievementRarityCommon:
		return "#CCCCCC" // Gray
	case AchievementRarityUncommon:
		return "#00FF00" // Green
	case AchievementRarityRare:
		return "#0080FF" // Blue
	case AchievementRarityEpic:
		return "#9933FF" // Purple
	case AchievementRarityLegendary:
		return "#FFD700" // Gold
	default:
		return "#FFFFFF" // White
	}
}
