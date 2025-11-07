// Package models - Random encounter definitions
//
// This file defines the random encounter system for dynamic space events.
// Encounters occur during travel and can include pirates, traders, distress calls,
// and faction patrols.
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

// EncounterType represents the type of random encounter
type EncounterType string

const (
	EncounterTypePirate      EncounterType = "pirate"       // Hostile pirates attack
	EncounterTypeTrader      EncounterType = "trader"       // Friendly trader offers goods
	EncounterTypeDistress    EncounterType = "distress"     // Ship in distress needs help
	EncounterTypePolice      EncounterType = "police"       // Police patrol (checks bounties)
	EncounterTypeFaction     EncounterType = "faction"      // Faction patrol
	EncounterTypeDerelict    EncounterType = "derelict"     // Abandoned ship to salvage
	EncounterTypeMerchant    EncounterType = "merchant"     // Merchant convoy
	EncounterTypeAsteroid    EncounterType = "asteroid"     // Asteroid field with minerals
)

// EncounterStatus represents the current state of an encounter
type EncounterStatus string

const (
	EncounterStatusActive    EncounterStatus = "active"     // Encounter in progress
	EncounterStatusResolved  EncounterStatus = "resolved"   // Encounter completed
	EncounterStatusFled      EncounterStatus = "fled"       // Player fled
	EncounterStatusIgnored   EncounterStatus = "ignored"    // Player ignored
)

// Encounter represents a random space encounter
type Encounter struct {
	ID          uuid.UUID       `json:"id"`
	Type        EncounterType   `json:"type"`
	Status      EncounterStatus `json:"status"`
	Title       string          `json:"title"`
	Description string          `json:"description"`

	// Ships involved in encounter
	Ships      []*Ship        `json:"ships,omitempty"`
	ShipTypes  []string       `json:"ship_types,omitempty"` // Ship type IDs
	ShipCount  int            `json:"ship_count"`

	// Faction involvement
	FactionID  string         `json:"faction_id,omitempty"`
	Hostile    bool           `json:"hostile"`

	// Rewards/Consequences
	CreditReward    int64  `json:"credit_reward,omitempty"`
	ReputationGain  int    `json:"reputation_gain,omitempty"`
	CargoReward     string `json:"cargo_reward,omitempty"` // Commodity ID
	CargoQuantity   int    `json:"cargo_quantity,omitempty"`

	// System context
	SystemID       uuid.UUID `json:"system_id"`
	DangerLevel    int       `json:"danger_level"` // 1-10

	// Metadata
	CreatedAt time.Time `json:"created_at"`
}

// EncounterOption represents a choice the player can make
type EncounterOption struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`

	// Requirements
	RequireCombatRating int   `json:"require_combat_rating,omitempty"`
	RequireCredits      int64 `json:"require_credits,omitempty"`
	RequireReputation   int   `json:"require_reputation,omitempty"` // With encounter faction

	// Outcomes
	StartsConflict bool   `json:"starts_conflict,omitempty"`
	CostCredits    int64  `json:"cost_credits,omitempty"`
	GrantsReward   bool   `json:"grants_reward,omitempty"`
	EndsEncounter  bool   `json:"ends_encounter,omitempty"`
	FleeAttempt    bool   `json:"flee_attempt,omitempty"`
}

// NewEncounter creates a new random encounter
//
// Parameters:
//   - encounterType: Type of encounter to create
//   - systemID: UUID of the system where encounter occurs
//   - dangerLevel: Danger level of the system (1-10)
//
// Returns:
//   - Pointer to new Encounter
func NewEncounter(encounterType EncounterType, systemID uuid.UUID, dangerLevel int) *Encounter {
	encounter := &Encounter{
		ID:          uuid.New(),
		Type:        encounterType,
		Status:      EncounterStatusActive,
		SystemID:    systemID,
		DangerLevel: dangerLevel,
		CreatedAt:   time.Now(),
	}

	// Generate encounter details based on type
	switch encounterType {
	case EncounterTypePirate:
		encounter.generatePirateEncounter(dangerLevel)
	case EncounterTypeTrader:
		encounter.generateTraderEncounter()
	case EncounterTypeDistress:
		encounter.generateDistressEncounter()
	case EncounterTypePolice:
		encounter.generatePoliceEncounter()
	case EncounterTypeFaction:
		encounter.generateFactionEncounter(dangerLevel)
	case EncounterTypeDerelict:
		encounter.generateDerelictEncounter()
	case EncounterTypeMerchant:
		encounter.generateMerchantEncounter()
	case EncounterTypeAsteroid:
		encounter.generateAsteroidEncounter()
	}

	return encounter
}

// generatePirateEncounter creates a hostile pirate encounter
func (e *Encounter) generatePirateEncounter(dangerLevel int) {
	e.Title = "Pirate Attack!"
	e.Description = "A group of pirates has locked onto your ship and is moving to attack!"
	e.Hostile = true
	e.FactionID = "crimson_syndicate"

	// More pirates in dangerous systems
	e.ShipCount = 1 + (dangerLevel / 3)
	if e.ShipCount > 4 {
		e.ShipCount = 4
	}

	// Ship types based on danger level
	if dangerLevel <= 3 {
		e.ShipTypes = []string{"fighter", "fighter"}
	} else if dangerLevel <= 6 {
		e.ShipTypes = []string{"fighter", "corvette"}
	} else {
		e.ShipTypes = []string{"corvette", "corvette", "frigate"}
	}

	// Pirate bounty reward
	e.CreditReward = int64(500 * dangerLevel * e.ShipCount)
	e.ReputationGain = 5 * e.ShipCount
}

// generateTraderEncounter creates a friendly trader encounter
func (e *Encounter) generateTraderEncounter() {
	e.Title = "Independent Trader"
	e.Description = "A trader is hailing you. They might have goods to sell or information to share."
	e.Hostile = false
	e.FactionID = "free_traders_guild"
	e.ShipCount = 1
	e.ShipTypes = []string{"light_freighter"}

	// Random commodity offer
	commodities := []string{"food", "water", "electronics", "fuel", "luxury_goods"}
	e.CargoReward = commodities[rand.Intn(len(commodities))]
	e.CargoQuantity = 5 + rand.Intn(15)

	// Discount price (30% off market)
	e.CreditReward = -int64(1000 + rand.Intn(3000)) // Negative = player pays
}

// generateDistressEncounter creates a distress call encounter
func (e *Encounter) generateDistressEncounter() {
	e.Title = "Distress Signal"
	e.Description = "You've picked up a distress signal from a damaged ship. They're requesting assistance."
	e.Hostile = false
	e.ShipCount = 1
	e.ShipTypes = []string{"shuttle", "light_freighter"}

	// Rescue reward
	e.CreditReward = int64(2000 + rand.Intn(5000))
	e.ReputationGain = 10
}

// generatePoliceEncounter creates a police patrol encounter
func (e *Encounter) generatePoliceEncounter() {
	e.Title = "Police Patrol"
	e.Description = "A police patrol is scanning your ship. They're checking for contraband and outstanding bounties."
	e.Hostile = false // Initially non-hostile
	e.FactionID = "united_earth_federation"
	e.ShipCount = 2
	e.ShipTypes = []string{"corvette", "corvette"}
}

// generateFactionEncounter creates a faction patrol encounter
func (e *Encounter) generateFactionEncounter(dangerLevel int) {
	// Random faction
	factions := []string{
		"united_earth_federation",
		"rigel_outer_marches",
		"free_worlds_alliance",
		"auroran_empire",
	}
	e.FactionID = factions[rand.Intn(len(factions))]

	e.Title = fmt.Sprintf("%s Patrol", e.FactionID)
	e.Description = "A faction patrol is in the area. Their response depends on your reputation with them."
	e.Hostile = false // Depends on reputation
	e.ShipCount = 2 + (dangerLevel / 4)
	e.ShipTypes = []string{"corvette", "frigate"}
}

// generateDerelictEncounter creates a derelict ship encounter
func (e *Encounter) generateDerelictEncounter() {
	e.Title = "Derelict Ship"
	e.Description = "You've found an abandoned ship floating in space. It might contain salvageable cargo or equipment."
	e.Hostile = false
	e.ShipCount = 1
	e.ShipTypes = []string{"light_freighter", "corvette"}

	// Salvage rewards
	e.CreditReward = int64(1000 + rand.Intn(10000))
	e.CargoQuantity = 5 + rand.Intn(20)
}

// generateMerchantEncounter creates a merchant convoy encounter
func (e *Encounter) generateMerchantEncounter() {
	e.Title = "Merchant Convoy"
	e.Description = "A well-protected merchant convoy is passing through. They might have rare goods for sale."
	e.Hostile = false
	e.FactionID = "free_traders_guild"
	e.ShipCount = 3
	e.ShipTypes = []string{"heavy_freighter", "corvette", "corvette"}

	// Expensive rare goods
	e.CargoReward = "luxury_goods"
	e.CargoQuantity = 10 + rand.Intn(20)
	e.CreditReward = -int64(5000 + rand.Intn(10000))
}

// generateAsteroidEncounter creates an asteroid field encounter
func (e *Encounter) generateAsteroidEncounter() {
	e.Title = "Asteroid Field"
	e.Description = "You've entered an asteroid field. With the right equipment, you could mine valuable minerals."
	e.Hostile = false
	e.ShipCount = 0

	// Mining rewards
	e.CargoReward = "ore"
	e.CargoQuantity = 10 + rand.Intn(30)
	e.CreditReward = int64(500 + rand.Intn(2000))
}

// GetOptions returns available options for this encounter
//
// Parameters:
//   - player: Player to check requirements against
//
// Returns:
//   - Slice of available EncounterOptions
func (e *Encounter) GetOptions(player *Player) []*EncounterOption {
	options := []*EncounterOption{}

	switch e.Type {
	case EncounterTypePirate:
		options = append(options, &EncounterOption{
			ID:             "engage",
			Label:          "Engage Pirates",
			Description:    fmt.Sprintf("Fight the pirates. Reward: %d cr", e.CreditReward),
			StartsConflict: true,
			GrantsReward:   true,
		})
		options = append(options, &EncounterOption{
			ID:            "flee",
			Label:         "Attempt to Flee",
			Description:   "Try to escape before they can attack",
			FleeAttempt:   true,
			EndsEncounter: true,
		})

	case EncounterTypeTrader:
		options = append(options, &EncounterOption{
			ID:            "trade",
			Label:         "Trade with Merchant",
			Description:   fmt.Sprintf("Buy %d tons of %s for %d cr", e.CargoQuantity, e.CargoReward, -e.CreditReward),
			CostCredits:   -e.CreditReward,
			GrantsReward:  true,
			EndsEncounter: true,
		})
		options = append(options, &EncounterOption{
			ID:            "ignore",
			Label:         "Ignore",
			Description:   "Decline the offer and move on",
			EndsEncounter: true,
		})

	case EncounterTypeDistress:
		options = append(options, &EncounterOption{
			ID:            "rescue",
			Label:         "Provide Assistance",
			Description:   fmt.Sprintf("Help the stranded ship. Reward: %d cr, +%d reputation", e.CreditReward, e.ReputationGain),
			GrantsReward:  true,
			EndsEncounter: true,
		})
		options = append(options, &EncounterOption{
			ID:            "ignore",
			Label:         "Ignore Distress Call",
			Description:   "Continue on your way",
			EndsEncounter: true,
		})

	case EncounterTypePolice:
		options = append(options, &EncounterOption{
			ID:            "cooperate",
			Label:         "Cooperate with Scan",
			Description:   "Allow the police to scan your ship",
			EndsEncounter: true,
		})
		if player.IsCriminal {
			options = append(options, &EncounterOption{
				ID:                  "bribe",
				Label:               "Attempt Bribe",
				Description:         "Try to bribe the officers (10,000 cr)",
				RequireCombatRating: 0,
				CostCredits:         10000,
				EndsEncounter:       true,
			})
		}
		options = append(options, &EncounterOption{
			ID:             "attack",
			Label:          "Attack Patrol",
			Description:    "Engage the police (will make you a criminal!)",
			StartsConflict: true,
		})

	case EncounterTypeDerelict:
		options = append(options, &EncounterOption{
			ID:            "salvage",
			Label:         "Salvage Ship",
			Description:   fmt.Sprintf("Search for cargo and equipment. Potential reward: %d cr", e.CreditReward),
			GrantsReward:  true,
			EndsEncounter: true,
		})
		options = append(options, &EncounterOption{
			ID:            "ignore",
			Label:         "Leave It",
			Description:   "Don't risk it, move on",
			EndsEncounter: true,
		})

	case EncounterTypeFaction:
		options = append(options, &EncounterOption{
			ID:            "hail",
			Label:         "Hail Patrol",
			Description:   "Greet the patrol. Response depends on your reputation",
			EndsEncounter: true,
		})
		options = append(options, &EncounterOption{
			ID:            "ignore",
			Label:         "Ignore and Move On",
			Description:   "Don't interact with them",
			EndsEncounter: true,
		})
	}

	return options
}

// CanAffordOption checks if player can afford an encounter option
//
// Parameters:
//   - option: Option to check
//   - player: Player to check requirements against
//
// Returns:
//   - true if player meets all requirements
func (e *Encounter) CanAffordOption(option *EncounterOption, player *Player) bool {
	if option.RequireCombatRating > 0 && player.CombatRating < option.RequireCombatRating {
		return false
	}

	if option.CostCredits > 0 && player.Credits < option.CostCredits {
		return false
	}

	if option.RequireReputation > 0 && e.FactionID != "" {
		if player.GetReputation(e.FactionID) < option.RequireReputation {
			return false
		}
	}

	return true
}

// Resolve marks the encounter as resolved
func (e *Encounter) Resolve() {
	e.Status = EncounterStatusResolved
}

// Flee marks the encounter as fled
func (e *Encounter) Flee() {
	e.Status = EncounterStatusFled
}

// Ignore marks the encounter as ignored
func (e *Encounter) Ignore() {
	e.Status = EncounterStatusIgnored
}
