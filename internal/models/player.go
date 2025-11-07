package models

import (
	"time"

	"github.com/google/uuid"
)

// Player represents a player character in the game
type Player struct {
	ID             uuid.UUID          `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email,omitempty"`
	EmailVerified  bool               `json:"email_verified"`
	PasswordHash   string             `json:"-"` // Never serialize password
	CreatedAt      time.Time          `json:"created_at"`
	LastLogin      time.Time          `json:"last_login"`

	// Game state
	Credits        int64              `json:"credits"`
	CurrentSystem  uuid.UUID          `json:"current_system"`
	CurrentPlanet  *uuid.UUID         `json:"current_planet,omitempty"` // nil if in space
	ShipID         uuid.UUID          `json:"ship_id"`

	// Progression
	CombatRating   int                `json:"combat_rating"`
	TotalKills     int                `json:"total_kills"`
	PlayTime       int64              `json:"play_time"` // seconds

	// Reputation with NPC factions (-100 to +100)
	Reputation     map[string]int     `json:"reputation"`

	// Faction membership
	FactionID      *uuid.UUID         `json:"faction_id,omitempty"`
	FactionRank    string             `json:"faction_rank,omitempty"`

	// Status
	IsOnline       bool               `json:"is_online"`
	IsCriminal     bool               `json:"is_criminal"`
}

// SSHKey represents an SSH public key for player authentication
type SSHKey struct {
	ID          uuid.UUID  `json:"id"`
	PlayerID    uuid.UUID  `json:"player_id"`
	KeyType     string     `json:"key_type"`     // rsa, ed25519, ecdsa
	PublicKey   string     `json:"public_key"`   // The actual public key
	Fingerprint string     `json:"fingerprint"`  // SHA256 fingerprint
	Comment     string     `json:"comment,omitempty"`
	AddedAt     time.Time  `json:"added_at"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// NewPlayer creates a new player with default starting values
func NewPlayer(username, passwordHash string) *Player {
	now := time.Now()
	return &Player{
		ID:             uuid.New(),
		Username:       username,
		PasswordHash:   passwordHash,
		CreatedAt:      now,
		LastLogin:      now,
		Credits:        10000, // Starting credits
		CombatRating:   0,
		TotalKills:     0,
		PlayTime:       0,
		Reputation:     make(map[string]int),
		IsOnline:       false,
		IsCriminal:     false,
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
