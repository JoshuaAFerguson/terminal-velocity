// File: internal/encounters/generator.go
// Project: Terminal Velocity
// Description: Random encounter system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package encounters provides random encounter generation and management.
//
// This package handles:
// - Encounter probability calculation
// - Encounter type selection based on context
// - Encounter generation based on system danger level
// - Integration with player status and reputation
//
// Version: 1.0.0
// Last Updated: 2025-01-07
package encounters

import (
	"math/rand"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Generator handles encounter generation

var log = logger.WithComponent("Encounters")

type Generator struct {
	baseEncounterChance float64 // Base 10% chance per jump
}

// NewGenerator creates a new encounter generator
//
// Returns:
//   - Pointer to new Generator
func NewGenerator() *Generator {
	return &Generator{
		baseEncounterChance: 0.10, // 10% base chance
	}
}

// ShouldGenerateEncounter determines if an encounter should occur
//
// Parameters:
//   - dangerLevel: System danger level (1-10)
//   - player: Player for context
//
// Returns:
//   - true if encounter should be generated
func (g *Generator) ShouldGenerateEncounter(dangerLevel int, player *models.Player) bool {
	// Adjust chance based on danger level
	// Danger 1 = 5%, Danger 10 = 25%
	chance := g.baseEncounterChance * (0.5 + (float64(dangerLevel) * 0.15))

	// Increase chance if player is criminal
	if player.IsCriminal {
		chance += 0.10
	}

	return rand.Float64() < chance
}

// GenerateEncounter creates a random encounter appropriate for the context
//
// Parameters:
//   - systemID: UUID of current system
//   - dangerLevel: System danger level (1-10)
//   - player: Player for context
//
// Returns:
//   - Pointer to generated Encounter
func (g *Generator) GenerateEncounter(systemID uuid.UUID, dangerLevel int, player *models.Player) *models.Encounter {
	encounterType := g.selectEncounterType(dangerLevel, player)
	return models.NewEncounter(encounterType, systemID, dangerLevel)
}

// selectEncounterType chooses an appropriate encounter type
//
// Parameters:
//   - dangerLevel: System danger level (1-10)
//   - player: Player for context
//
// Returns:
//   - Selected EncounterType
func (g *Generator) selectEncounterType(dangerLevel int, player *models.Player) models.EncounterType {
	// Weight encounters based on danger level and player status
	weights := make(map[models.EncounterType]int)

	// Pirate encounters more common in dangerous systems
	weights[models.EncounterTypePirate] = 10 + (dangerLevel * 5)

	// Trader encounters less common in dangerous systems
	weights[models.EncounterTypeTrader] = 20 - (dangerLevel * 2)

	// Distress calls moderate in all systems
	weights[models.EncounterTypeDistress] = 15

	// Police more common in safe systems
	weights[models.EncounterTypePolice] = 25 - (dangerLevel * 2)

	// Faction patrols in mid-danger systems
	weights[models.EncounterTypeFaction] = 10 + (dangerLevel / 2)

	// Derelicts in dangerous systems
	weights[models.EncounterTypeDerelict] = dangerLevel * 3

	// Merchants in safe systems
	weights[models.EncounterTypeMerchant] = 15 - dangerLevel

	// Asteroids rare but consistent
	weights[models.EncounterTypeAsteroid] = 5

	// Increase police encounters if player is criminal
	if player.IsCriminal {
		weights[models.EncounterTypePolice] += 20
		weights[models.EncounterTypeTrader] -= 10 // Traders avoid criminals
	}

	// Calculate total weight
	totalWeight := 0
	for _, weight := range weights {
		if weight > 0 {
			totalWeight += weight
		}
	}

	// Select random encounter based on weights
	roll := rand.Intn(totalWeight)
	currentWeight := 0

	for encounterType, weight := range weights {
		if weight <= 0 {
			continue
		}
		currentWeight += weight
		if roll < currentWeight {
			return encounterType
		}
	}

	// Fallback (should never reach here)
	return models.EncounterTypePirate
}

// GenerateEncounterShips creates ships for an encounter
//
// Parameters:
//   - encounter: Encounter to generate ships for
//
// Returns:
//   - Slice of generated ships
func (g *Generator) GenerateEncounterShips(encounter *models.Encounter) []*models.Ship {
	ships := []*models.Ship{}

	for i := 0; i < encounter.ShipCount && i < len(encounter.ShipTypes); i++ {
		shipTypeID := encounter.ShipTypes[i]
		shipType := models.GetShipTypeByID(shipTypeID)
		if shipType == nil {
			continue
		}

		// Create ship
		ship := &models.Ship{
			ID:      uuid.New(),
			TypeID:  shipTypeID,
			Name:    g.generateShipName(encounter.Type, i+1),
			Hull:    shipType.MaxHull,
			Shields: shipType.MaxShields,
			Fuel:    shipType.MaxFuel,
			Cargo:   []models.CargoItem{},
			Weapons: []string{},
			Outfits: []string{},
		}

		// Equip with basic weapons based on ship class
		ship.Weapons = g.generateShipWeapons(shipType)

		ships = append(ships, ship)
	}

	return ships
}

// generateShipName creates a name for an encounter ship
//
// Parameters:
//   - encounterType: Type of encounter
//   - index: Ship number in encounter
//
// Returns:
//   - Ship name string
func (g *Generator) generateShipName(encounterType models.EncounterType, index int) string {
	prefixes := map[models.EncounterType][]string{
		models.EncounterTypePirate:   {"Crimson", "Blood", "Shadow", "Raven", "Viper"},
		models.EncounterTypeTrader:   {"Free", "Independent", "Merchant"},
		models.EncounterTypeDistress: {"Distressed", "Damaged", "Stranded"},
		models.EncounterTypePolice:   {"UEF Patrol", "UEF Guardian", "UEF Enforcer"},
		models.EncounterTypeFaction:  {"Patrol", "Scout", "Sentinel"},
		models.EncounterTypeDerelict: {"Abandoned", "Derelict", "Ghost"},
		models.EncounterTypeMerchant: {"Caravan", "Trader", "Convoy"},
	}

	suffixes := []string{"Alpha", "Beta", "Gamma", "Delta", "One", "Two", "Three"}

	prefixList, exists := prefixes[encounterType]
	if !exists || len(prefixList) == 0 {
		return "Unknown Ship"
	}

	prefix := prefixList[rand.Intn(len(prefixList))]
	suffix := ""
	if index > 0 && index <= len(suffixes) {
		suffix = " " + suffixes[index-1]
	}

	return prefix + suffix
}

// generateShipWeapons equips a ship with appropriate weapons
//
// Parameters:
//   - shipType: Type of ship to equip
//
// Returns:
//   - Slice of weapon IDs
func (g *Generator) generateShipWeapons(shipType *models.ShipType) []string {
	weapons := []string{}

	// Determine weapon count based on ship class
	weaponCount := 0
	switch shipType.Class {
	case "shuttle":
		weaponCount = 1
	case "fighter":
		weaponCount = 2
	case "light_freighter", "heavy_freighter":
		weaponCount = 1
	case "corvette":
		weaponCount = 3
	case "frigate":
		weaponCount = 4
	case "destroyer", "heavy_destroyer":
		weaponCount = 5
	case "cruiser", "heavy_cruiser":
		weaponCount = 6
	case "battleship":
		weaponCount = 8
	}

	// Select appropriate weapons
	lightWeapons := []string{"pulse_laser", "gatling_laser"}
	mediumWeapons := []string{"beam_laser", "light_missile"}
	heavyWeapons := []string{"heavy_laser", "heavy_missile", "plasma_cannon"}

	for i := 0; i < weaponCount; i++ {
		var weaponID string
		if shipType.Class == "shuttle" || shipType.Class == "fighter" {
			weaponID = lightWeapons[rand.Intn(len(lightWeapons))]
		} else if shipType.Class == "corvette" || shipType.Class == "frigate" {
			weaponID = mediumWeapons[rand.Intn(len(mediumWeapons))]
		} else {
			weaponID = heavyWeapons[rand.Intn(len(heavyWeapons))]
		}
		weapons = append(weapons, weaponID)
	}

	return weapons
}

// CalculateFleeSuccess determines if a flee attempt succeeds
//
// Parameters:
//   - playerShip: Player's ship
//   - playerShipType: Player's ship type
//   - encounterShips: Enemy ships
//
// Returns:
//   - true if flee succeeds
func (g *Generator) CalculateFleeSuccess(playerShip *models.Ship, playerShipType *models.ShipType, encounterShips []*models.Ship) bool {
	// Base 50% chance
	baseChance := 0.5

	// Bonus for faster ships
	if playerShipType != nil {
		speedBonus := float64(playerShipType.Speed) / 200.0 // Max speed ~100-200
		baseChance += speedBonus * 0.3
	}

	// Penalty for damaged ship
	if playerShipType != nil && playerShip.Hull < playerShipType.MaxHull {
		hullPercent := float64(playerShip.Hull) / float64(playerShipType.MaxHull)
		if hullPercent < 0.5 {
			baseChance -= 0.2
		}
	}

	// Penalty for multiple enemies
	if len(encounterShips) > 1 {
		baseChance -= float64(len(encounterShips)-1) * 0.1
	}

	// Cap between 10% and 90%
	if baseChance < 0.1 {
		baseChance = 0.1
	}
	if baseChance > 0.9 {
		baseChance = 0.9
	}

	return rand.Float64() < baseChance
}
