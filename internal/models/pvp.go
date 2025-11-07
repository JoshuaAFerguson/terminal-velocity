// File: internal/models/pvp.go
// Project: Terminal Velocity
// Description: PvP combat models with consent system and bounties
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PvPChallengeStatus represents the state of a combat challenge
type PvPChallengeStatus string

const (
	ChallengePending  PvPChallengeStatus = "pending"  // Awaiting response
	ChallengeAccepted PvPChallengeStatus = "accepted" // Both parties ready
	ChallengeActive   PvPChallengeStatus = "active"   // Combat in progress
	ChallengeDeclined PvPChallengeStatus = "declined" // Refused
	ChallengeExpired  PvPChallengeStatus = "expired"  // Timed out
	ChallengeComplete PvPChallengeStatus = "complete" // Finished
)

// PvPChallengeType represents different combat scenarios
type PvPChallengeType string

const (
	ChallengeDuel       PvPChallengeType = "duel"        // Honorable 1v1, no penalty
	ChallengeAggression PvPChallengeType = "aggression"  // Unprovoked attack, bounty risk
	ChallengeBountyHunt PvPChallengeType = "bounty_hunt" // Hunting wanted player
	ChallengeFactionWar PvPChallengeType = "faction_war" // Faction vs faction
	ChallengeDefense    PvPChallengeType = "defense"     // Defending territory
)

// PvPChallenge represents a combat proposal between players
type PvPChallenge struct {
	ID          uuid.UUID          `json:"id"`
	ChallengerID uuid.UUID         `json:"challenger_id"`
	ChallengerName string          `json:"challenger_name"`
	DefenderID  uuid.UUID          `json:"defender_id"`
	DefenderName string             `json:"defender_name"`

	Type        PvPChallengeType   `json:"type"`
	Status      PvPChallengeStatus `json:"status"`

	// Location
	SystemID uuid.UUID `json:"system_id"`
	PlanetID *uuid.UUID `json:"planet_id,omitempty"`

	// Timing
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`

	// Wager (optional)
	Wager int64 `json:"wager"` // Credits on the line

	// Result
	WinnerID *uuid.UUID `json:"winner_id,omitempty"`
	LoserID  *uuid.UUID `json:"loser_id,omitempty"`

	// Reason/message
	Message string `json:"message,omitempty"`

	// Consent flags
	RequiresConsent bool `json:"requires_consent"` // If false, immediate combat (faction war, etc.)
	ConsentGiven    bool `json:"consent_given"`
}

// PvPCombatResult represents the outcome of a combat
type PvPCombatResult struct {
	ChallengeID uuid.UUID `json:"challenge_id"`
	WinnerID    uuid.UUID `json:"winner_id"`
	LoserID     uuid.UUID `json:"loser_id"`

	// Damage dealt
	WinnerDamage int `json:"winner_damage"`
	LoserDamage  int `json:"loser_damage"`

	// Rewards/penalties
	CreditsWon  int64 `json:"credits_won"`
	CreditsLost int64 `json:"credits_lost"`

	// Cargo loot
	CargoLooted map[string]int `json:"cargo_looted"` // commodity -> quantity

	// Reputation changes
	WinnerRepChange int `json:"winner_rep_change"`
	LoserRepChange  int `json:"loser_rep_change"`

	// Bounty effects
	BountyAdded   int64 `json:"bounty_added"`   // If aggressor
	BountyClaimed int64 `json:"bounty_claimed"` // If bounty hunter

	Duration time.Duration `json:"duration"`
	Timestamp time.Time    `json:"timestamp"`
}

// Bounty represents a wanted status on a player
type Bounty struct {
	ID       uuid.UUID `json:"id"`
	TargetID uuid.UUID `json:"target_id"`
	TargetName string  `json:"target_name"`

	// Bounty details
	Amount      int64     `json:"amount"`       // Credits for capture/kill
	Reason      string    `json:"reason"`       // Why they're wanted
	IssuedBy    string    `json:"issued_by"`    // "System" or faction name
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`

	// Status
	Active      bool       `json:"active"`
	ClaimedBy   *uuid.UUID `json:"claimed_by,omitempty"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`

	// Crime tracking
	Kills       int   `json:"kills"`        // Number of unjustified kills
	Thefts      int   `json:"thefts"`       // Cargo piracy
	Attacks     int   `json:"attacks"`      // Unprovoked attacks
	CrimeValue  int64 `json:"crime_value"`  // Total value of crimes
}

// PvPStats tracks a player's combat history
type PvPStats struct {
	PlayerID uuid.UUID `json:"player_id"`

	// Win/loss record
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`

	// Combat types
	DuelsWon       int `json:"duels_won"`
	BountiesHunted int `json:"bounties_hunted"`
	PirateKills    int `json:"pirate_kills"`

	// Negative actions
	AggressionCount int `json:"aggression_count"` // Unprovoked attacks
	PlayersKilled   int `json:"players_killed"`

	// Financial
	TotalCreditsWon  int64 `json:"total_credits_won"`
	TotalCreditsLost int64 `json:"total_credits_lost"`

	// Reputation
	CombatRating    int     `json:"combat_rating"`    // 0-1000
	HonorRating     float64 `json:"honor_rating"`     // 0.0-1.0
	NotorietyLevel  int     `json:"notoriety_level"`  // 0-5 (wanted level)

	// Statistics
	TotalDamageDealt    int64 `json:"total_damage_dealt"`
	TotalDamageTaken    int64 `json:"total_damage_taken"`
	LongestWinStreak    int   `json:"longest_win_streak"`
	CurrentWinStreak    int   `json:"current_win_streak"`

	LastCombatAt time.Time `json:"last_combat_at"`
}

// NewPvPChallenge creates a new combat challenge
func NewPvPChallenge(
	challengerID uuid.UUID,
	challengerName string,
	defenderID uuid.UUID,
	defenderName string,
	challengeType PvPChallengeType,
	systemID uuid.UUID,
) *PvPChallenge {
	now := time.Now()

	requiresConsent := true
	if challengeType == ChallengeFactionWar || challengeType == ChallengeDefense {
		requiresConsent = false
	}

	return &PvPChallenge{
		ID:              uuid.New(),
		ChallengerID:    challengerID,
		ChallengerName:  challengerName,
		DefenderID:      defenderID,
		DefenderName:    defenderName,
		Type:            challengeType,
		Status:          ChallengePending,
		SystemID:        systemID,
		CreatedAt:       now,
		ExpiresAt:       now.Add(5 * time.Minute), // 5 minute expiry
		Wager:           0,
		RequiresConsent: requiresConsent,
		ConsentGiven:    !requiresConsent, // Auto-consent for faction wars
	}
}

// IsExpired checks if the challenge has expired
func (c *PvPChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt) && c.Status == ChallengePending
}

// Accept marks the challenge as accepted
func (c *PvPChallenge) Accept() {
	c.Status = ChallengeAccepted
	now := time.Now()
	c.AcceptedAt = &now
	c.ConsentGiven = true
}

// Decline marks the challenge as declined
func (c *PvPChallenge) Decline() {
	c.Status = ChallengeDeclined
}

// Start marks the combat as active
func (c *PvPChallenge) Start() {
	c.Status = ChallengeActive
	now := time.Now()
	c.StartedAt = &now
}

// Complete marks the combat as finished
func (c *PvPChallenge) Complete(winnerID uuid.UUID) {
	c.Status = ChallengeComplete
	now := time.Now()
	c.EndedAt = &now
	c.WinnerID = &winnerID

	if winnerID == c.ChallengerID {
		c.LoserID = &c.DefenderID
	} else {
		c.LoserID = &c.ChallengerID
	}
}

// GetTimeRemaining returns time until expiry
func (c *PvPChallenge) GetTimeRemaining() string {
	if c.Status != ChallengePending {
		return "-"
	}

	duration := time.Until(c.ExpiresAt)
	if duration < 0 {
		return "Expired"
	}

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}

	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// GetTypeIcon returns an icon for the challenge type
func (t PvPChallengeType) GetIcon() string {
	icons := map[PvPChallengeType]string{
		ChallengeDuel:       "⚔️",
		ChallengeAggression: "💢",
		ChallengeBountyHunt: "🎯",
		ChallengeFactionWar: "⚔️",
		ChallengeDefense:    "🛡️",
	}

	if icon, exists := icons[t]; exists {
		return icon
	}
	return "⚔️"
}

// GetStatusIcon returns an icon for the challenge status
func (s PvPChallengeStatus) GetIcon() string {
	icons := map[PvPChallengeStatus]string{
		ChallengePending:  "⏳",
		ChallengeAccepted: "✅",
		ChallengeActive:   "⚡",
		ChallengeDeclined: "❌",
		ChallengeExpired:  "⌛",
		ChallengeComplete: "✔️",
	}

	if icon, exists := icons[s]; exists {
		return icon
	}
	return "❓"
}

// NewBounty creates a new bounty
func NewBounty(targetID uuid.UUID, targetName string, amount int64, reason string, issuedBy string) *Bounty {
	now := time.Now()

	return &Bounty{
		ID:         uuid.New(),
		TargetID:   targetID,
		TargetName: targetName,
		Amount:     amount,
		Reason:     reason,
		IssuedBy:   issuedBy,
		IssuedAt:   now,
		ExpiresAt:  now.Add(7 * 24 * time.Hour), // 7 days
		Active:     true,
		Kills:      0,
		Thefts:     0,
		Attacks:    0,
		CrimeValue: 0,
	}
}

// AddCrime increases the bounty for a crime
func (b *Bounty) AddCrime(crimeType string, value int64) {
	switch crimeType {
	case "kill":
		b.Kills++
		b.Amount += 10000
	case "theft":
		b.Thefts++
		b.Amount += value / 2 // Half the stolen value
	case "attack":
		b.Attacks++
		b.Amount += 5000
	}

	b.CrimeValue += value
}

// Claim marks the bounty as claimed
func (b *Bounty) Claim(hunterID uuid.UUID) {
	b.Active = false
	b.ClaimedBy = &hunterID
	now := time.Now()
	b.ClaimedAt = &now
}

// IsExpired checks if the bounty has expired
func (b *Bounty) IsExpired() bool {
	return time.Now().After(b.ExpiresAt)
}

// GetWantedLevel returns a display string for the wanted level
func (b *Bounty) GetWantedLevel() string {
	totalCrimes := b.Kills + b.Thefts + b.Attacks

	switch {
	case totalCrimes >= 10:
		return "⭐⭐⭐⭐⭐ MOST WANTED"
	case totalCrimes >= 7:
		return "⭐⭐⭐⭐ DANGEROUS"
	case totalCrimes >= 5:
		return "⭐⭐⭐ WANTED"
	case totalCrimes >= 3:
		return "⭐⭐ SUSPECT"
	default:
		return "⭐ MINOR"
	}
}

// NewPvPStats creates new PvP stats for a player
func NewPvPStats(playerID uuid.UUID) *PvPStats {
	return &PvPStats{
		PlayerID:           playerID,
		Wins:               0,
		Losses:             0,
		Draws:              0,
		DuelsWon:           0,
		BountiesHunted:     0,
		PirateKills:        0,
		AggressionCount:    0,
		PlayersKilled:      0,
		TotalCreditsWon:    0,
		TotalCreditsLost:   0,
		CombatRating:       100, // Start at 100
		HonorRating:        1.0, // Start with perfect honor
		NotorietyLevel:     0,
		TotalDamageDealt:   0,
		TotalDamageTaken:   0,
		LongestWinStreak:   0,
		CurrentWinStreak:   0,
	}
}

// RecordWin records a combat victory
func (s *PvPStats) RecordWin(combatType PvPChallengeType, creditsWon int64, damageDealt int64) {
	s.Wins++
	s.CurrentWinStreak++
	s.TotalCreditsWon += creditsWon
	s.TotalDamageDealt += damageDealt
	s.LastCombatAt = time.Now()

	if s.CurrentWinStreak > s.LongestWinStreak {
		s.LongestWinStreak = s.CurrentWinStreak
	}

	// Type-specific tracking
	switch combatType {
	case ChallengeDuel:
		s.DuelsWon++
		s.CombatRating += 10
	case ChallengeBountyHunt:
		s.BountiesHunted++
		s.CombatRating += 25
		s.HonorRating = min(1.0, s.HonorRating+0.05) // Bounty hunting increases honor
	case ChallengeAggression:
		s.PlayersKilled++
		s.AggressionCount++
		s.CombatRating += 5
		s.HonorRating = maxFloat(0.0, s.HonorRating-0.1) // Aggression decreases honor
		s.NotorietyLevel++
	}

	s.updateCombatRating()
}

// RecordLoss records a combat defeat
func (s *PvPStats) RecordLoss(creditsLost int64, damageTaken int64) {
	s.Losses++
	s.CurrentWinStreak = 0
	s.TotalCreditsLost += creditsLost
	s.TotalDamageTaken += damageTaken
	s.LastCombatAt = time.Now()

	s.CombatRating = maxInt(0, s.CombatRating-5)
	s.updateCombatRating()
}

// RecordDraw records a draw
func (s *PvPStats) RecordDraw() {
	s.Draws++
	s.CurrentWinStreak = 0
	s.LastCombatAt = time.Now()
}

// updateCombatRating ensures combat rating stays within bounds
func (s *PvPStats) updateCombatRating() {
	if s.CombatRating > 1000 {
		s.CombatRating = 1000
	}
	if s.CombatRating < 0 {
		s.CombatRating = 0
	}
}

// GetWinRate returns win percentage
func (s *PvPStats) GetWinRate() float64 {
	total := s.Wins + s.Losses + s.Draws
	if total == 0 {
		return 0.0
	}
	return (float64(s.Wins) / float64(total)) * 100.0
}

// GetKDRatio returns kill/death ratio
func (s *PvPStats) GetKDRatio() float64 {
	if s.Losses == 0 {
		return float64(s.Wins)
	}
	return float64(s.Wins) / float64(s.Losses)
}

// GetRatingClass returns a human-readable rating class
func (s *PvPStats) GetRatingClass() string {
	switch {
	case s.CombatRating >= 900:
		return "⭐ Elite"
	case s.CombatRating >= 750:
		return "🥇 Master"
	case s.CombatRating >= 600:
		return "🥈 Expert"
	case s.CombatRating >= 450:
		return "🥉 Veteran"
	case s.CombatRating >= 300:
		return "💪 Skilled"
	case s.CombatRating >= 150:
		return "⚔️ Competent"
	default:
		return "🔰 Novice"
	}
}

// GetHonorRank returns a human-readable honor rank
func (s *PvPStats) GetHonorRank() string {
	switch {
	case s.HonorRating >= 0.9:
		return "😇 Honorable"
	case s.HonorRating >= 0.7:
		return "👍 Reputable"
	case s.HonorRating >= 0.5:
		return "😐 Neutral"
	case s.HonorRating >= 0.3:
		return "😠 Dishonorable"
	default:
		return "👿 Villain"
	}
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
