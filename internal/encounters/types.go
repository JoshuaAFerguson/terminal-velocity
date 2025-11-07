// File: internal/encounters/types.go
// Project: Terminal Velocity
// Description: Random encounter types and event system
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package encounters

import (
	"time"

	"github.com/google/uuid"
)

// EncounterType represents different types of random encounters
type EncounterType string

const (
	// Common encounters
	EncounterPirate       EncounterType = "pirate"         // Hostile pirate attack
	EncounterTrader       EncounterType = "trader"         // Friendly trader
	EncounterPatrol       EncounterType = "patrol"         // System patrol
	EncounterDistress     EncounterType = "distress"       // Ship in distress

	// Uncommon encounters
	EncounterConvoy       EncounterType = "convoy"         // Trading convoy
	EncounterMercenary    EncounterType = "mercenary"      // Mercenary offer
	EncounterScavenger    EncounterType = "scavenger"      // Scavenger ship

	// Rare encounters
	EncounterDerelict     EncounterType = "derelict"       // Abandoned ship
	EncounterAnomaly      EncounterType = "anomaly"        // Space anomaly
	EncounterBountyTarget EncounterType = "bounty_target"  // Wanted criminal
	EncounterMystery      EncounterType = "mystery"        // Unknown signal

	// Very rare encounters
	EncounterAncient      EncounterType = "ancient"        // Ancient artifact
	EncounterLeviathan    EncounterType = "leviathan"      // Massive creature
	EncounterGhostShip    EncounterType = "ghost_ship"     // Mysterious vessel
)

// EncounterRarity determines how often an encounter appears
type EncounterRarity string

const (
	RarityCommon   EncounterRarity = "common"    // 50% chance
	RarityUncommon EncounterRarity = "uncommon"  // 30% chance
	RarityRare     EncounterRarity = "rare"      // 15% chance
	RarityVeryRare EncounterRarity = "very_rare" // 5% chance
	RarityLegendary EncounterRarity = "legendary" // 1% chance
)

// EncounterOutcome represents what happened in an encounter
type EncounterOutcome string

const (
	OutcomeEngaged  EncounterOutcome = "engaged"  // Player engaged
	OutcomeAvoided  EncounterOutcome = "avoided"  // Player avoided
	OutcomeFled     EncounterOutcome = "fled"     // Player fled
	OutcomeHelped   EncounterOutcome = "helped"   // Player helped
	OutcomeIgnored  EncounterOutcome = "ignored"  // Player ignored
	OutcomeDestroyed EncounterOutcome = "destroyed" // Player destroyed enemy
)

// Encounter represents a random event
type Encounter struct {
	ID          uuid.UUID       `json:"id"`
	Type        EncounterType   `json:"type"`
	Rarity      EncounterRarity `json:"rarity"`

	// Location
	SystemID    uuid.UUID       `json:"system_id"`
	SystemName  string          `json:"system_name"`

	// Timing
	OccurredAt  time.Time       `json:"occurred_at"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty"`

	// Details
	Title       string          `json:"title"`
	Description string          `json:"description"`

	// NPCs involved
	NPCName     string          `json:"npc_name,omitempty"`
	NPCShipType string          `json:"npc_ship_type,omitempty"`
	NPCLevel    int             `json:"npc_level"` // Difficulty

	// Rewards/penalties
	Credits     int64           `json:"credits"`
	Cargo       map[string]int  `json:"cargo,omitempty"`
	Reputation  int             `json:"reputation"`

	// Outcome
	Outcome     EncounterOutcome `json:"outcome,omitempty"`
	PlayerID    uuid.UUID        `json:"player_id"`
}

// EncounterTemplate defines a type of encounter
type EncounterTemplate struct {
	Type        EncounterType
	Rarity      EncounterRarity
	Title       string
	Description string

	// Spawn conditions
	MinTechLevel    int
	MaxTechLevel    int
	RequiredGovType string // Empty = any

	// NPC details
	ShipTypes       []string
	MinLevel        int
	MaxLevel        int

	// Rewards
	MinCredits      int64
	MaxCredits      int64
	PossibleCargo   []string
	ReputationRange [2]int // Min, max

	// Options
	CanAvoid        bool
	CanFlee         bool
	RequiresCombat  bool
	IsHostile       bool
}

// NewEncounter creates a new encounter from a template
func NewEncounter(template EncounterTemplate, systemID uuid.UUID, systemName string, playerID uuid.UUID) *Encounter {
	return &Encounter{
		ID:          uuid.New(),
		Type:        template.Type,
		Rarity:      template.Rarity,
		SystemID:    systemID,
		SystemName:  systemName,
		OccurredAt:  time.Now(),
		Title:       template.Title,
		Description: template.Description,
		NPCLevel:    template.MinLevel, // Will be randomized by manager
		Credits:     template.MinCredits,
		Cargo:       make(map[string]int),
		Reputation:  template.ReputationRange[0],
		PlayerID:    playerID,
	}
}

// Resolve marks the encounter as resolved
func (e *Encounter) Resolve(outcome EncounterOutcome) {
	e.Outcome = outcome
	now := time.Now()
	e.ResolvedAt = &now
}

// IsResolved checks if the encounter has been resolved
func (e *Encounter) IsResolved() bool {
	return e.ResolvedAt != nil
}

// GetRarityIcon returns an icon for the rarity
func (r EncounterRarity) GetIcon() string {
	icons := map[EncounterRarity]string{
		RarityCommon:    "⚪",
		RarityUncommon:  "🟢",
		RarityRare:      "🔵",
		RarityVeryRare:  "🟣",
		RarityLegendary: "🟡",
	}

	if icon, exists := icons[r]; exists {
		return icon
	}
	return "⚪"
}

// GetTypeIcon returns an icon for the encounter type
func (t EncounterType) GetIcon() string {
	icons := map[EncounterType]string{
		EncounterPirate:       "🏴‍☠️",
		EncounterTrader:       "🚚",
		EncounterPatrol:       "🚔",
		EncounterDistress:     "🆘",
		EncounterConvoy:       "🚛",
		EncounterMercenary:    "⚔️",
		EncounterScavenger:    "🔧",
		EncounterDerelict:     "👻",
		EncounterAnomaly:      "🌀",
		EncounterBountyTarget: "🎯",
		EncounterMystery:      "❓",
		EncounterAncient:      "🏛️",
		EncounterLeviathan:    "🐋",
		EncounterGhostShip:    "⛴️",
	}

	if icon, exists := icons[t]; exists {
		return icon
	}
	return "❓"
}

// EncounterHistory tracks player's encounter history
type EncounterHistory struct {
	PlayerID           uuid.UUID                    `json:"player_id"`
	TotalEncounters    int                          `json:"total_encounters"`
	ByType             map[EncounterType]int        `json:"by_type"`
	ByRarity           map[EncounterRarity]int      `json:"by_rarity"`
	RarestFound        EncounterRarity              `json:"rarest_found"`
	LastEncounterAt    time.Time                    `json:"last_encounter_at"`

	// Statistics
	Engaged            int                          `json:"engaged"`
	Avoided            int                          `json:"avoided"`
	Fled               int                          `json:"fled"`
	Helped             int                          `json:"helped"`

	// Rewards
	TotalCreditsEarned int64                        `json:"total_credits_earned"`
	TotalCargoFound    int                          `json:"total_cargo_found"`
}

// NewEncounterHistory creates a new encounter history
func NewEncounterHistory(playerID uuid.UUID) *EncounterHistory {
	return &EncounterHistory{
		PlayerID:           playerID,
		TotalEncounters:    0,
		ByType:             make(map[EncounterType]int),
		ByRarity:           make(map[EncounterRarity]int),
		RarestFound:        RarityCommon,
		Engaged:            0,
		Avoided:            0,
		Fled:               0,
		Helped:             0,
		TotalCreditsEarned: 0,
		TotalCargoFound:    0,
	}
}

// RecordEncounter records an encounter in history
func (h *EncounterHistory) RecordEncounter(encounter *Encounter) {
	h.TotalEncounters++
	h.ByType[encounter.Type]++
	h.ByRarity[encounter.Rarity]++
	h.LastEncounterAt = encounter.OccurredAt

	// Update rarest found
	if h.isRarer(encounter.Rarity, h.RarestFound) {
		h.RarestFound = encounter.Rarity
	}

	// Record outcome
	switch encounter.Outcome {
	case OutcomeEngaged, OutcomeDestroyed:
		h.Engaged++
	case OutcomeAvoided:
		h.Avoided++
	case OutcomeFled:
		h.Fled++
	case OutcomeHelped:
		h.Helped++
	}

	// Record rewards
	if encounter.Credits > 0 {
		h.TotalCreditsEarned += encounter.Credits
	}

	for _, qty := range encounter.Cargo {
		h.TotalCargoFound += qty
	}
}

// isRarer checks if a rarity is rarer than another
func (h *EncounterHistory) isRarer(a, b EncounterRarity) bool {
	rarityOrder := map[EncounterRarity]int{
		RarityCommon:    1,
		RarityUncommon:  2,
		RarityRare:      3,
		RarityVeryRare:  4,
		RarityLegendary: 5,
	}

	return rarityOrder[a] > rarityOrder[b]
}

// GetEngagementRate returns the percentage of encounters engaged
func (h *EncounterHistory) GetEngagementRate() float64 {
	if h.TotalEncounters == 0 {
		return 0.0
	}
	return (float64(h.Engaged) / float64(h.TotalEncounters)) * 100.0
}

// GetAvoidanceRate returns the percentage of encounters avoided
func (h *EncounterHistory) GetAvoidanceRate() float64 {
	if h.TotalEncounters == 0 {
		return 0.0
	}
	return (float64(h.Avoided + h.Fled) / float64(h.TotalEncounters)) * 100.0
}
