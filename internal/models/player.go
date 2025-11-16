// File: internal/models/player.go
// Project: Terminal Velocity
// Description: Data models for player with comprehensive field documentation
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package models defines core game data structures.
//
// This package provides data models for all game entities including:
// - Player characters with progression tracking
// - Ships, weapons, and equipment
// - Universe structures (systems, planets, routes)
// - Trading commodities and markets
// - Missions and objectives
// - Factions and reputation
//
// Version: 1.1.0
// Last Updated: 2025-01-07
package models

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// Player represents a player character in the game.
//
// This is the central model for all player data, tracking their progress,
// statistics, reputation, and current state in the game world.
type Player struct {
	// ID is the unique identifier for this player (primary key)
	ID uuid.UUID `json:"id"`

	// Username is the player's display name (unique, 3-20 characters)
	Username string `json:"username"`

	// Email is the player's email address (unique, optional for SSH-only auth)
	Email string `json:"email,omitempty"`

	// EmailVerified indicates whether the player has verified their email address
	EmailVerified bool `json:"email_verified"`

	// PasswordHash stores the bcrypt hash of the player's password
	// Never serialized to JSON for security
	PasswordHash string `json:"-"`

	// CreatedAt is when the player account was created
	CreatedAt time.Time `json:"created_at"`

	// LastLogin is when the player last logged in
	LastLogin time.Time `json:"last_login"`

	// Game state

	// Credits is the player's current credit balance
	Credits int64 `json:"credits"`

	// CurrentSystem is the star system where the player is currently located
	CurrentSystem uuid.UUID `json:"current_system"`

	// CurrentPlanet is the planet where the player is docked (nil if in space)
	CurrentPlanet *uuid.UUID `json:"current_planet,omitempty"`

	// ShipID is the UUID of the ship the player is currently piloting
	ShipID uuid.UUID `json:"ship_id"`

	// Position (coordinates within system)

	// X is the X coordinate in the current system (in-system position)
	X float64 `json:"x"`

	// Y is the Y coordinate in the current system (in-system position)
	Y float64 `json:"y"`

	// Progression - Combat

	// CombatRating is the player's combat skill rating (0-100)
	CombatRating int `json:"combat_rating"`

	// TotalKills is the total number of enemy ships destroyed
	TotalKills int `json:"total_kills"`

	// PlayTime is the total time played in seconds
	PlayTime int64 `json:"play_time"`

	// Progression - Trading

	// TradingRating is the player's trading skill rating (0-100)
	TradingRating int `json:"trading_rating"`

	// TotalTrades is the total number of commodity trades completed
	TotalTrades int `json:"total_trades"`

	// TradeProfit is the cumulative profit from all trades (can be negative)
	TradeProfit int64 `json:"trade_profit"`

	// HighestProfit is the largest single trade profit achieved
	HighestProfit int64 `json:"highest_profit"`

	// Progression - Exploration

	// ExplorationRating is the player's exploration skill rating (0-100)
	ExplorationRating int `json:"exploration_rating"`

	// SystemsVisited is the number of unique star systems visited
	SystemsVisited int `json:"systems_visited"`

	// TotalJumps is the total number of hyperspace jumps made
	TotalJumps int `json:"total_jumps"`

	// Progression - Missions

	// MissionsCompleted is the number of missions successfully completed
	MissionsCompleted int `json:"missions_completed"`

	// MissionsFailed is the number of missions failed or abandoned
	MissionsFailed int `json:"missions_failed"`

	// Progression - Quests

	// QuestsCompleted is the number of story quests completed
	QuestsCompleted int `json:"quests_completed"`

	// Progression - Capture

	// TotalCaptureAttempts is the number of ship boarding attempts made
	TotalCaptureAttempts int `json:"total_capture_attempts"`

	// SuccessfulBoards is the number of successful boarding actions
	SuccessfulBoards int `json:"successful_boards"`

	// SuccessfulCaptures is the number of ships successfully captured
	SuccessfulCaptures int `json:"successful_captures"`

	// Progression - Mining

	// TotalMiningOps is the number of mining operations performed
	TotalMiningOps int `json:"total_mining_ops"`

	// TotalYield is the total quantity of resources mined (all types)
	TotalYield int64 `json:"total_yield"`

	// ResourcesMined tracks quantities mined by resource type
	ResourcesMined map[string]int64 `json:"resources_mined"`

	// Progression - Manufacturing/Crafting

	// CraftingSkill is the crafting skill level (0-100)
	CraftingSkill int `json:"crafting_skill"`

	// TotalCrafts is the total number of items crafted
	TotalCrafts int `json:"total_crafts"`

	// Progression - Research

	// ResearchPoints are available points for unlocking technologies
	ResearchPoints int `json:"research_points"`

	// Progression - Overall

	// Level is the overall player level (1-100)
	Level int `json:"level"`

	// Experience is the accumulated experience points for leveling
	Experience int64 `json:"experience"`

	// Reputation with NPC factions

	// Reputation tracks standing with each NPC faction (range: -100 to +100)
	// Map key is faction ID, value is reputation score
	Reputation map[string]int `json:"reputation"`

	// Faction membership

	// FactionID is the UUID of the player faction this player belongs to (nil if none)
	FactionID *uuid.UUID `json:"faction_id,omitempty"`

	// FactionRank is the player's rank within their faction (e.g., "Leader", "Officer", "Member")
	FactionRank string `json:"faction_rank,omitempty"`

	// Legal status

	// LegalStatus is the player's legal standing ("citizen", "outlaw", "pirate", "wanted", "hostile")
	LegalStatus string `json:"legal_status"`

	// Bounty is the credit reward for destroying/capturing this player
	Bounty int64 `json:"bounty"`

	// Status

	// IsOnline indicates whether the player is currently connected
	IsOnline bool `json:"is_online"`

	// IsCriminal indicates whether the player has a criminal record
	// Deprecated: use LegalStatus instead
	IsCriminal bool `json:"is_criminal"`

	// UpdatedAt is the timestamp of the last update to this player record
	UpdatedAt time.Time `json:"updated_at"`
}

// SSHKey represents an SSH public key for player authentication.
//
// Players can authenticate using SSH public keys instead of passwords.
// Each player can have multiple SSH keys associated with their account.
type SSHKey struct {
	// ID is the unique identifier for this SSH key
	ID uuid.UUID `json:"id"`

	// PlayerID is the UUID of the player who owns this key
	PlayerID uuid.UUID `json:"player_id"`

	// KeyType is the algorithm used for the key (rsa, ed25519, ecdsa)
	KeyType string `json:"key_type"`

	// PublicKey is the actual SSH public key data
	PublicKey string `json:"public_key"`

	// Fingerprint is the SHA256 fingerprint of the public key (used for matching)
	Fingerprint string `json:"fingerprint"`

	// Comment is an optional description or identifier for the key
	Comment string `json:"comment,omitempty"`

	// AddedAt is when the key was added to the account
	AddedAt time.Time `json:"added_at"`

	// LastUsed is when the key was last used for authentication (nil if never used)
	LastUsed *time.Time `json:"last_used,omitempty"`

	// IsActive indicates whether the key is currently enabled for authentication
	IsActive bool `json:"is_active"`
}

// NewPlayer creates a new player with default starting values.
// All progression statistics start at zero.
//
// Parameters:
//   - username: Player's chosen username
//   - passwordHash: Bcrypt hash of player's password
//
// Returns:
//   - Pointer to new Player struct with initialized values
func NewPlayer(username, passwordHash string) *Player {
	now := time.Now()
	return &Player{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		LastLogin:    now,
		Credits:      10000, // Starting credits

		// Position
		X: 0,
		Y: 0,

		// Combat progression
		CombatRating: 0,
		TotalKills:   0,
		PlayTime:     0,

		// Trading progression
		TradingRating: 0,
		TotalTrades:   0,
		TradeProfit:   0,
		HighestProfit: 0,

		// Exploration progression
		ExplorationRating: 0,
		SystemsVisited:    0,
		TotalJumps:        0,

		// Mission progression
		MissionsCompleted: 0,
		MissionsFailed:    0,

		// Quest progression
		QuestsCompleted: 0,

		// Capture progression
		TotalCaptureAttempts: 0,
		SuccessfulBoards:     0,
		SuccessfulCaptures:   0,

		// Mining progression
		TotalMiningOps: 0,
		TotalYield:     0,
		ResourcesMined: make(map[string]int64),

		// Manufacturing/Crafting progression
		CraftingSkill: 0,
		TotalCrafts:   0,

		// Research progression
		ResearchPoints: 100, // Starting research points

		// Overall progression
		Level:      1,  // Start at level 1
		Experience: 0,  // No experience yet

		Reputation: make(map[string]int),

		// Legal status
		LegalStatus: "citizen",  // Start as citizen
		Bounty:      0,           // No bounty

		IsOnline:   false,
		IsCriminal: false,
		UpdatedAt:  now,
	}
}

// CanAfford checks if player has enough credits
func (p *Player) CanAfford(amount int64) bool {
	return p.Credits >= amount
}

// AddCredits adds credits to player (can be negative)
func (p *Player) AddCredits(amount int64) {
	p.Credits += amount
	if p.Credits < 0 {
		p.Credits = 0
	}
}

// GetReputation returns reputation with a faction (0 if not set)
func (p *Player) GetReputation(factionID string) int {
	if rep, ok := p.Reputation[factionID]; ok {
		return rep
	}
	return 0
}

// ModifyReputation changes reputation with a faction
func (p *Player) ModifyReputation(factionID string, delta int) {
	current := p.GetReputation(factionID)
	newRep := current + delta

	// Clamp to -100 to +100
	if newRep < -100 {
		newRep = -100
	} else if newRep > 100 {
		newRep = 100
	}

	p.Reputation[factionID] = newRep
}

// IsInFaction checks if player is in any faction
func (p *Player) IsInFaction() bool {
	return p.FactionID != nil
}

// IsDocked checks if player is currently docked at a planet
func (p *Player) IsDocked() bool {
	return p.CurrentPlanet != nil
}

// ==================== Progression System ====================

// CalculateCombatRating calculates the player's combat rating based on performance.
// Rating is based on total kills with diminishing returns for higher kill counts.
// Rating scale: 0 (Harmless) to 100 (Elite)
//
// Formula: rating = sqrt(kills) * 10, capped at 100
//
// Returns:
//   - Combat rating from 0-100
func (p *Player) CalculateCombatRating() int {
	if p.TotalKills <= 0 {
		return 0
	}

	// Square root scaling for diminishing returns
	// 1 kill = 10 rating, 4 kills = 20 rating, 16 kills = 40 rating, 100 kills = 100 rating
	rating := int(math.Sqrt(float64(p.TotalKills)) * 10)

	// Cap at 100
	if rating > 100 {
		rating = 100
	}

	return rating
}

// CalculateTradingRating calculates the player's trading rating based on performance.
// Rating is based on total profit with volume multiplier.
// Rating scale: 0 (Peddler) to 100 (Tycoon)
//
// Formula: rating = (profit / 100000) * (1 + trades/100), capped at 100
//
// Returns:
//   - Trading rating from 0-100
func (p *Player) CalculateTradingRating() int {
	if p.TradeProfit <= 0 {
		return 0
	}

	// Base rating from profit (100K = 1 point)
	baseRating := float64(p.TradeProfit) / 100000.0

	// Volume multiplier (more trades = higher rating)
	volumeMultiplier := 1.0 + (float64(p.TotalTrades) / 100.0)

	rating := int(baseRating * volumeMultiplier)

	// Cap at 100
	if rating > 100 {
		rating = 100
	}

	return rating
}

// CalculateExplorationRating calculates the player's exploration rating.
// Rating is based on systems visited and jump activity.
// Rating scale: 0 (Tourist) to 100 (Pathfinder)
//
// Formula: rating = (systems * 2) + (jumps / 10), capped at 100
//
// Returns:
//   - Exploration rating from 0-100
func (p *Player) CalculateExplorationRating() int {
	// Systems visited is weighted more heavily than jump count
	systemPoints := p.SystemsVisited * 2
	jumpPoints := p.TotalJumps / 10

	rating := systemPoints + jumpPoints

	// Cap at 100
	if rating > 100 {
		rating = 100
	}

	return rating
}

// GetCombatRankTitle returns the player's combat rank title based on rating.
//
// Rank thresholds:
//   - 0-9: Harmless
//   - 10-19: Mostly Harmless
//   - 20-39: Poor
//   - 40-59: Average
//   - 60-79: Competent
//   - 80-89: Dangerous
//   - 90-99: Deadly
//   - 100: Elite
//
// Returns:
//   - Combat rank title string
func (p *Player) GetCombatRankTitle() string {
	rating := p.CalculateCombatRating()

	switch {
	case rating >= 100:
		return "Elite"
	case rating >= 90:
		return "Deadly"
	case rating >= 80:
		return "Dangerous"
	case rating >= 60:
		return "Competent"
	case rating >= 40:
		return "Average"
	case rating >= 20:
		return "Poor"
	case rating >= 10:
		return "Mostly Harmless"
	default:
		return "Harmless"
	}
}

// GetTradingRankTitle returns the player's trading rank title based on rating.
//
// Rank thresholds:
//   - 0-9: Peddler
//   - 10-19: Trader
//   - 20-39: Merchant
//   - 40-59: Dealer
//   - 60-79: Wholesaler
//   - 80-89: Magnate
//   - 90-99: Mogul
//   - 100: Tycoon
//
// Returns:
//   - Trading rank title string
func (p *Player) GetTradingRankTitle() string {
	rating := p.CalculateTradingRating()

	switch {
	case rating >= 100:
		return "Tycoon"
	case rating >= 90:
		return "Mogul"
	case rating >= 80:
		return "Magnate"
	case rating >= 60:
		return "Wholesaler"
	case rating >= 40:
		return "Dealer"
	case rating >= 20:
		return "Merchant"
	case rating >= 10:
		return "Trader"
	default:
		return "Peddler"
	}
}

// GetExplorationRankTitle returns the player's exploration rank title based on rating.
//
// Rank thresholds:
//   - 0-9: Tourist
//   - 10-19: Traveler
//   - 20-39: Voyager
//   - 40-59: Surveyor
//   - 60-79: Navigator
//   - 80-89: Pioneer
//   - 90-99: Trailblazer
//   - 100: Pathfinder
//
// Returns:
//   - Exploration rank title string
func (p *Player) GetExplorationRankTitle() string {
	rating := p.CalculateExplorationRating()

	switch {
	case rating >= 100:
		return "Pathfinder"
	case rating >= 90:
		return "Trailblazer"
	case rating >= 80:
		return "Pioneer"
	case rating >= 60:
		return "Navigator"
	case rating >= 40:
		return "Surveyor"
	case rating >= 20:
		return "Voyager"
	case rating >= 10:
		return "Traveler"
	default:
		return "Tourist"
	}
}

// RecordTrade updates trading statistics after a trade transaction.
//
// Parameters:
//   - profit: The profit (or loss if negative) from the trade
func (p *Player) RecordTrade(profit int64) {
	p.TotalTrades++
	p.TradeProfit += profit

	// Update highest profit if this trade is a new record
	if profit > p.HighestProfit {
		p.HighestProfit = profit
	}

	// Recalculate trading rating
	p.TradingRating = p.CalculateTradingRating()
}

// RecordKill updates combat statistics after defeating an enemy.
func (p *Player) RecordKill() {
	p.TotalKills++

	// Recalculate combat rating
	p.CombatRating = p.CalculateCombatRating()
}

// RecordSystemVisit updates exploration statistics when entering a new system.
//
// Parameters:
//   - systemID: The UUID of the system being visited
func (p *Player) RecordSystemVisit(systemID uuid.UUID) {
	// Note: In a full implementation, we'd track which systems have been visited
	// to avoid double-counting. For now, this increments on every system change.
	p.SystemsVisited++

	// Recalculate exploration rating
	p.ExplorationRating = p.CalculateExplorationRating()
}

// RecordJump updates exploration statistics when making a hyperspace jump.
func (p *Player) RecordJump() {
	p.TotalJumps++

	// Recalculate exploration rating
	p.ExplorationRating = p.CalculateExplorationRating()
}

// RecordMissionCompletion updates mission statistics after completing a mission.
func (p *Player) RecordMissionCompletion() {
	p.MissionsCompleted++
}

// RecordMissionFailure updates mission statistics after failing a mission.
func (p *Player) RecordMissionFailure() {
	p.MissionsFailed++
}

// RecordCaptureAttempt updates capture statistics when attempting to board a ship.
func (p *Player) RecordCaptureAttempt() {
	p.TotalCaptureAttempts++
}

// RecordSuccessfulBoard updates capture statistics after successfully boarding a ship.
func (p *Player) RecordSuccessfulBoard() {
	p.SuccessfulBoards++
}

// RecordSuccessfulCapture updates capture statistics after successfully capturing a ship.
func (p *Player) RecordSuccessfulCapture() {
	p.SuccessfulCaptures++
}

// RecordMiningOperation updates mining statistics after completing a mining operation.
//
// Parameters:
//   - yield: The amount of resources mined
//   - resources: Map of resource types to quantities mined
func (p *Player) RecordMiningOperation(yield int64, resources map[string]int) {
	p.TotalMiningOps++
	p.TotalYield += yield

	// Initialize map if nil
	if p.ResourcesMined == nil {
		p.ResourcesMined = make(map[string]int64)
	}

	// Update resources mined by type
	for resourceType, quantity := range resources {
		p.ResourcesMined[resourceType] += int64(quantity)
	}
}

// GetMostCommonResource returns the most commonly mined resource type.
// Returns empty string if no resources have been mined.
func (p *Player) GetMostCommonResource() string {
	if p.ResourcesMined == nil || len(p.ResourcesMined) == 0 {
		return ""
	}

	var mostCommon string
	var maxQuantity int64

	for resourceType, quantity := range p.ResourcesMined {
		if quantity > maxQuantity {
			maxQuantity = quantity
			mostCommon = resourceType
		}
	}

	return mostCommon
}

// GetOverallRank returns a combined rank based on all progression categories.
// The overall rank is a weighted average of all rating categories.
//
// Weighting:
//   - Combat: 30%
//   - Trading: 30%
//   - Exploration: 20%
//   - Missions: 20%
//
// Returns:
//   - Overall rank title string
func (p *Player) GetOverallRank() string {
	// Calculate individual ratings
	combatRating := float64(p.CalculateCombatRating())
	tradingRating := float64(p.CalculateTradingRating())
	explorationRating := float64(p.CalculateExplorationRating())

	// Mission rating based on completion rate
	missionRating := 0.0
	totalMissions := p.MissionsCompleted + p.MissionsFailed
	if totalMissions > 0 {
		successRate := float64(p.MissionsCompleted) / float64(totalMissions)
		missionRating = successRate * 100
	}

	// Weighted average
	overall := (combatRating * 0.3) + (tradingRating * 0.3) + (explorationRating * 0.2) + (missionRating * 0.2)

	switch {
	case overall >= 90:
		return "Legendary"
	case overall >= 80:
		return "Master"
	case overall >= 70:
		return "Expert"
	case overall >= 60:
		return "Veteran"
	case overall >= 50:
		return "Experienced"
	case overall >= 40:
		return "Proficient"
	case overall >= 30:
		return "Competent"
	case overall >= 20:
		return "Novice"
	case overall >= 10:
		return "Rookie"
	default:
		return "Beginner"
	}
}
