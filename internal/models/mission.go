package models

import (
	"time"

	"github.com/google/uuid"
)

// Mission represents a mission/quest
type Mission struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // delivery, combat, escort, bounty, exploration
	Title       string    `json:"title"`
	Description string    `json:"description"`

	// Giver
	GiverID      string    `json:"giver_id"` // NPC or faction ID
	OriginPlanet uuid.UUID `json:"origin_planet"`

	// Objectives
	Destination *uuid.UUID `json:"destination,omitempty"` // Destination system or planet
	Target      *string    `json:"target,omitempty"`      // Enemy ship type or commodity
	Quantity    int        `json:"quantity,omitempty"`    // Cargo quantity or kill count

	// Rewards
	Reward           int64          `json:"reward"`
	ReputationChange map[string]int `json:"reputation_change"` // Faction rep changes

	// Timing
	Deadline   time.Time `json:"deadline"`
	AcceptedAt time.Time `json:"accepted_at"`

	// State
	Status   string `json:"status"`   // available, active, completed, failed
	Progress int    `json:"progress"` // For multi-part missions

	// Requirements
	MinCombatRating int            `json:"min_combat_rating"`
	RequiredRep     map[string]int `json:"required_rep"` // Minimum reputation needed

	// Associated cargo (for delivery missions)
	Cargo *CargoItem `json:"cargo,omitempty"`
}

// Mission types
const (
	MissionTypeDelivery    = "delivery"
	MissionTypeCombat      = "combat"
	MissionTypeEscort      = "escort"
	MissionTypeBounty      = "bounty"
	MissionTypeExploration = "exploration"
	MissionTypeTrading     = "trading"
)

// Mission status
const (
	MissionStatusAvailable = "available"
	MissionStatusActive    = "active"
	MissionStatusCompleted = "completed"
	MissionStatusFailed    = "failed"
)

// NewDeliveryMission creates a delivery mission
func NewDeliveryMission(giver string, origin, destination uuid.UUID, commodity string, quantity int, reward int64, deadline time.Time) *Mission {
	return &Mission{
		ID:           uuid.New(),
		Type:         MissionTypeDelivery,
		Title:        "Cargo Delivery",
		Description:  "Deliver cargo to the destination",
		GiverID:      giver,
		OriginPlanet: origin,
		Destination:  &destination,
		Quantity:     quantity,
		Reward:       reward,
		Deadline:     deadline,
		AcceptedAt:   time.Now(),
		Status:       MissionStatusAvailable,
		Progress:     0,
		Cargo: &CargoItem{
			CommodityID: commodity,
			Quantity:    quantity,
		},
		ReputationChange: make(map[string]int),
		RequiredRep:      make(map[string]int),
	}
}

// NewCombatMission creates a combat mission
func NewCombatMission(giver string, origin uuid.UUID, target string, kills int, reward int64, minCombatRating int) *Mission {
	return &Mission{
		ID:               uuid.New(),
		Type:             MissionTypeCombat,
		Title:            "Combat Patrol",
		Description:      "Eliminate hostile targets",
		GiverID:          giver,
		OriginPlanet:     origin,
		Target:           &target,
		Quantity:         kills,
		Reward:           reward,
		Deadline:         time.Now().Add(24 * time.Hour),
		AcceptedAt:       time.Now(),
		Status:           MissionStatusAvailable,
		Progress:         0,
		MinCombatRating:  minCombatRating,
		ReputationChange: make(map[string]int),
		RequiredRep:      make(map[string]int),
	}
}

// IsExpired checks if mission has passed deadline
func (m *Mission) IsExpired() bool {
	return time.Now().After(m.Deadline)
}

// IsCompleted checks if mission objectives are met
func (m *Mission) IsCompleted() bool {
	return m.Status == MissionStatusCompleted || m.Progress >= m.Quantity
}

// CanAccept checks if player meets mission requirements
func (m *Mission) CanAccept(player *Player) bool {
	// Check combat rating
	if player.CombatRating < m.MinCombatRating {
		return false
	}

	// Check reputation requirements
	for factionID, requiredRep := range m.RequiredRep {
		if player.GetReputation(factionID) < requiredRep {
			return false
		}
	}

	return true
}

// UpdateProgress increments mission progress
func (m *Mission) UpdateProgress(amount int) {
	m.Progress += amount
	if m.Progress >= m.Quantity {
		m.Status = MissionStatusCompleted
	}
}

// Fail marks mission as failed
func (m *Mission) Fail() {
	m.Status = MissionStatusFailed
}

// Complete marks mission as completed
func (m *Mission) Complete() {
	m.Status = MissionStatusCompleted
	m.Progress = m.Quantity
}
