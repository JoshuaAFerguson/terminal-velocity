// File: internal/models/player.go
// Project: Terminal Velocity
// Description: Data models for player
// Version: 1.0.0
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

// Player represents a player character in the game

type Player struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	PasswordHash  string    `json:"-"` // Never serialize password
	CreatedAt     time.Time `json:"created_at"`
	LastLogin     time.Time `json:"last_login"`

	// Game state
	Credits       int64      `json:"credits"`
	CurrentSystem uuid.UUID  `json:"current_system"`
	CurrentPlanet *uuid.UUID `json:"current_planet,omitempty"` // nil if in space
	ShipID        uuid.UUID  `json:"ship_id"`

	// Position (coordinates within system)
	X float64 `json:"x"` // X coordinate in current system
	Y float64 `json:"y"` // Y coordinate in current system

	// Progression - Combat
	CombatRating int   `json:"combat_rating"`
	TotalKills   int   `json:"total_kills"`
	PlayTime     int64 `json:"play_time"` // seconds

	// Progression - Trading
	TradingRating int   `json:"trading_rating"`
	TotalTrades   int   `json:"total_trades"`
	TradeProfit   int64 `json:"trade_profit"`   // Total profit from trading
	HighestProfit int64 `json:"highest_profit"` // Largest single trade profit

	// Progression - Exploration
	ExplorationRating int `json:"exploration_rating"`
	SystemsVisited    int `json:"systems_visited"`
	TotalJumps        int `json:"total_jumps"`

	// Progression - Missions
	MissionsCompleted int `json:"missions_completed"`
	MissionsFailed    int `json:"missions_failed"`

	// Progression - Quests
	QuestsCompleted int `json:"quests_completed"`

	// Progression - Capture
	TotalCaptureAttempts int `json:"total_capture_attempts"`
	SuccessfulBoards     int `json:"successful_boards"`
	SuccessfulCaptures   int `json:"successful_captures"`

	// Progression - Mining
	TotalMiningOps int   `json:"total_mining_ops"`
	TotalYield     int64 `json:"total_yield"` // Total resources mined

	// Progression - Overall
	Level      int   `json:"level"`       // Overall player level (1-100)
	Experience int64 `json:"experience"`  // Experience points for leveling

	// Reputation with NPC factions (-100 to +100)
	Reputation map[string]int `json:"reputation"`

	// Faction membership
	FactionID   *uuid.UUID `json:"faction_id,omitempty"`
	FactionRank string     `json:"faction_rank,omitempty"`

	// Legal status
	LegalStatus string `json:"legal_status"` // "citizen", "outlaw", "pirate", "wanted", "hostile"
	Bounty      int64  `json:"bounty"`       // Bounty on player's head (credits)

	// Status
	IsOnline   bool      `json:"is_online"`
	IsCriminal bool      `json:"is_criminal"` // Deprecated: use LegalStatus instead
	UpdatedAt  time.Time `json:"updated_at"`  // Last update timestamp
}

// SSHKey represents an SSH public key for player authentication
type SSHKey struct {
	ID          uuid.UUID  `json:"id"`
	PlayerID    uuid.UUID  `json:"player_id"`
	KeyType     string     `json:"key_type"`    // rsa, ed25519, ecdsa
	PublicKey   string     `json:"public_key"`  // The actual public key
	Fingerprint string     `json:"fingerprint"` // SHA256 fingerprint
	Comment     string     `json:"comment,omitempty"`
	AddedAt     time.Time  `json:"added_at"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	IsActive    bool       `json:"is_active"`
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
func (p *Player) RecordMiningOperation(yield int64) {
	p.TotalMiningOps++
	p.TotalYield += yield
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
