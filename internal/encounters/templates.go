// File: internal/encounters/templates.go
// Project: Terminal Velocity
// Description: Pre-defined encounter templates
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package encounters

// GetAllTemplates returns all encounter templates
func GetAllTemplates() []EncounterTemplate {
	return []EncounterTemplate{
		// Common Encounters
		GetPirateTemplate(),
		GetTraderTemplate(),
		GetPatrolTemplate(),
		GetDistressTemplate(),

		// Uncommon Encounters
		GetConvoyTemplate(),
		GetMercenaryTemplate(),
		GetScavengerTemplate(),

		// Rare Encounters
		GetDerelictTemplate(),
		GetAnomalyTemplate(),
		GetBountyTargetTemplate(),
		GetMysteryTemplate(),

		// Very Rare Encounters
		GetAncientTemplate(),
		GetLeviathanTemplate(),
		GetGhostShipTemplate(),
	}
}

// GetPirateTemplate returns a pirate encounter
func GetPirateTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterPirate,
		Rarity:          RarityCommon,
		Title:           "Pirate Attack!",
		Description:     "A pirate ship appears on your scanners, weapons hot. They're demanding your cargo or your life!",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Scout", "Corvette", "Frigate"},
		MinLevel:        1,
		MaxLevel:        5,
		MinCredits:      500,
		MaxCredits:      3000,
		PossibleCargo:   []string{"Weapons", "Drugs", "Stolen Goods"},
		ReputationRange: [2]int{-5, 10}, // Loss if you pay, gain if you win
		CanAvoid:        true,
		CanFlee:         true,
		RequiresCombat:  false,
		IsHostile:       true,
	}
}

// GetTraderTemplate returns a friendly trader encounter
func GetTraderTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterTrader,
		Rarity:          RarityCommon,
		Title:           "Friendly Trader",
		Description:     "An independent trader hails you. They're willing to trade goods at fair prices.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Light Freighter", "Freighter", "Courier"},
		MinLevel:        1,
		MaxLevel:        3,
		MinCredits:      0,
		MaxCredits:      0,
		PossibleCargo:   []string{"Food", "Water", "Textiles", "Machinery"},
		ReputationRange: [2]int{0, 5},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetPatrolTemplate returns a system patrol encounter
func GetPatrolTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterPatrol,
		Rarity:          RarityCommon,
		Title:           "System Patrol",
		Description:     "A system patrol ship is conducting routine scans. They request permission to scan your cargo for contraband.",
		MinTechLevel:    2,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Corvette", "Frigate"},
		MinLevel:        2,
		MaxLevel:        6,
		MinCredits:      -5000, // Fine if carrying contraband
		MaxCredits:      1000,  // Reward for reporting pirates
		PossibleCargo:   []string{},
		ReputationRange: [2]int{-10, 5},
		CanAvoid:        false,
		CanFlee:         true,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetDistressTemplate returns a distress call encounter
func GetDistressTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterDistress,
		Rarity:          RarityCommon,
		Title:           "Distress Call",
		Description:     "You receive a distress signal from a damaged ship. They're requesting assistance and offering a reward.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Shuttle", "Scout", "Courier"},
		MinLevel:        1,
		MaxLevel:        3,
		MinCredits:      1000,
		MaxCredits:      5000,
		PossibleCargo:   []string{"Fuel", "Equipment"},
		ReputationRange: [2]int{5, 15},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetConvoyTemplate returns a trading convoy encounter
func GetConvoyTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterConvoy,
		Rarity:          RarityUncommon,
		Title:           "Trading Convoy",
		Description:     "A large trading convoy passes through. They're well-guarded but offer bulk trading opportunities.",
		MinTechLevel:    2,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Freighter", "Heavy Freighter"},
		MinLevel:        3,
		MaxLevel:        5,
		MinCredits:      0,
		MaxCredits:      0,
		PossibleCargo:   []string{"Luxury Goods", "Electronics", "Machinery", "Gems"},
		ReputationRange: [2]int{0, 10},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetMercenaryTemplate returns a mercenary encounter
func GetMercenaryTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterMercenary,
		Rarity:          RarityUncommon,
		Title:           "Mercenary Offer",
		Description:     "A skilled mercenary pilot offers their services. They can help with combat missions or escort duties for a fee.",
		MinTechLevel:    2,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Corvette", "Frigate", "Destroyer"},
		MinLevel:        4,
		MaxLevel:        8,
		MinCredits:      -10000, // Cost to hire
		MaxCredits:      0,
		PossibleCargo:   []string{},
		ReputationRange: [2]int{0, 5},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetScavengerTemplate returns a scavenger encounter
func GetScavengerTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterScavenger,
		Rarity:          RarityUncommon,
		Title:           "Scavenger Ship",
		Description:     "A scavenger has found salvage from a recent battle. They're willing to sell equipment at discounted prices.",
		MinTechLevel:    1,
		MaxTechLevel:    5,
		RequiredGovType: "",
		ShipTypes:       []string{"Courier", "Scout"},
		MinLevel:        2,
		MaxLevel:        4,
		MinCredits:      0,
		MaxCredits:      0,
		PossibleCargo:   []string{"Weapons", "Equipment", "Scrap Metal"},
		ReputationRange: [2]int{0, 5},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetDerelictTemplate returns a derelict ship encounter
func GetDerelictTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterDerelict,
		Rarity:          RarityRare,
		Title:           "Derelict Vessel",
		Description:     "You discover an abandoned ship drifting in space. Scans show no life signs, but valuable cargo remains aboard.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Light Freighter", "Freighter", "Corvette"},
		MinLevel:        0,
		MaxLevel:        0,
		MinCredits:      5000,
		MaxCredits:      20000,
		PossibleCargo:   []string{"Luxury Goods", "Equipment", "Weapons", "Fuel"},
		ReputationRange: [2]int{0, 0},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetAnomalyTemplate returns a space anomaly encounter
func GetAnomalyTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterAnomaly,
		Rarity:          RarityRare,
		Title:           "Space Anomaly",
		Description:     "Your instruments detect a spatial anomaly. Investigation could yield valuable scientific data... or danger.",
		MinTechLevel:    3,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{},
		MinLevel:        0,
		MaxLevel:        0,
		MinCredits:      10000,
		MaxCredits:      50000,
		PossibleCargo:   []string{"Exotic Matter", "Ancient Artifacts"},
		ReputationRange: [2]int{10, 25},
		CanAvoid:        true,
		CanFlee:         true,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetBountyTargetTemplate returns a bounty target encounter
func GetBountyTargetTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterBountyTarget,
		Rarity:          RarityRare,
		Title:           "Wanted Criminal",
		Description:     "Your scanners identify a ship with an active bounty. This could be a lucrative opportunity... if you can handle them.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Corvette", "Frigate", "Destroyer"},
		MinLevel:        5,
		MaxLevel:        9,
		MinCredits:      15000,
		MaxCredits:      50000,
		PossibleCargo:   []string{"Stolen Goods", "Contraband", "Weapons"},
		ReputationRange: [2]int{15, 30},
		CanAvoid:        true,
		CanFlee:         true,
		RequiresCombat:  true,
		IsHostile:       true,
	}
}

// GetMysteryTemplate returns a mystery signal encounter
func GetMysteryTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterMystery,
		Rarity:          RarityRare,
		Title:           "Unknown Signal",
		Description:     "You detect an unusual signal that doesn't match any known ship or station. The source is unclear.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{},
		MinLevel:        0,
		MaxLevel:        10,
		MinCredits:      -10000, // Could be a trap
		MaxCredits:      100000, // Could be treasure
		PossibleCargo:   []string{"Unknown"},
		ReputationRange: [2]int{-20, 50},
		CanAvoid:        true,
		CanFlee:         true,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetAncientTemplate returns an ancient artifact encounter
func GetAncientTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterAncient,
		Rarity:          RarityVeryRare,
		Title:           "Ancient Artifact",
		Description:     "Your long-range scanners detect energy signatures consistent with ancient alien technology. This could be a historic discovery!",
		MinTechLevel:    4,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{},
		MinLevel:        0,
		MaxLevel:        0,
		MinCredits:      50000,
		MaxCredits:      250000,
		PossibleCargo:   []string{"Ancient Technology", "Alien Artifacts"},
		ReputationRange: [2]int{50, 100},
		CanAvoid:        true,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}

// GetLeviathanTemplate returns a space leviathan encounter
func GetLeviathanTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterLeviathan,
		Rarity:          RarityVeryRare,
		Title:           "Space Leviathan",
		Description:     "A massive creature of unknown origin appears. Legends speak of space-dwelling beasts, but few believed they were real.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{},
		MinLevel:        10,
		MaxLevel:        15,
		MinCredits:      100000,
		MaxCredits:      500000,
		PossibleCargo:   []string{"Bio-matter", "Rare Compounds"},
		ReputationRange: [2]int{100, 200},
		CanAvoid:        false,
		CanFlee:         true,
		RequiresCombat:  true,
		IsHostile:       true,
	}
}

// GetGhostShipTemplate returns a ghost ship encounter
func GetGhostShipTemplate() EncounterTemplate {
	return EncounterTemplate{
		Type:            EncounterGhostShip,
		Rarity:          RarityLegendary,
		Title:           "The Ghost Ship",
		Description:     "A vessel materializes from nowhere, matching descriptions of the legendary \"Phantom Wanderer\" - a ship that appears only to the worthy or the doomed.",
		MinTechLevel:    1,
		MaxTechLevel:    7,
		RequiredGovType: "",
		ShipTypes:       []string{"Ancient Vessel"},
		MinLevel:        12,
		MaxLevel:        20,
		MinCredits:      250000,
		MaxCredits:      1000000,
		PossibleCargo:   []string{"Legendary Equipment", "Ancient Treasures", "Cursed Artifacts"},
		ReputationRange: [2]int{200, 500},
		CanAvoid:        false,
		CanFlee:         false,
		RequiresCombat:  false,
		IsHostile:       false,
	}
}
