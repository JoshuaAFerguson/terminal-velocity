// File: internal/models/npc_faction.go
// Project: Terminal Velocity
// Description: Data models for npc_faction
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

// NPCFaction represents a non-player faction/government in the universe
type NPCFaction struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"short_name"` // Abbreviation
	Description string `json:"description"`
	Lore        string `json:"lore"`

	// Visual identity
	Color string `json:"color"` // Hex color for UI
	Flag  string `json:"flag"`  // ASCII art flag/symbol

	// Territory
	HomeSystem  string   `json:"home_system"`  // System name (Earth, Alpha Centauri, etc.)
	CoreSystems []string `json:"core_systems"` // Fully controlled systems
	Influence   []string `json:"influence"`    // Systems with presence

	// Characteristics
	Government string `json:"government"` // democracy, autocracy, corporate, hive_mind, etc.
	Economy    string `json:"economy"`    // capitalist, socialist, mixed, post_scarcity
	Military   string `json:"military"`   // peaceful, defensive, aggressive, expansionist
	Technology int    `json:"technology"` // Average tech level 1-10

	// Behavior
	Stance string   `json:"stance"` // isolationist, diplomatic, aggressive, expansionist
	Traits []string `json:"traits"` // wealthy, militaristic, scientific, etc.

	// Relations
	Allies  []string `json:"allies"`  // Allied faction IDs
	Enemies []string `json:"enemies"` // Enemy faction IDs
	Neutral []string `json:"neutral"` // Neutral faction IDs

	// Gameplay
	StartingRep    int    `json:"starting_rep"`    // Default reputation with player (-100 to 100)
	ShipPrefix     string `json:"ship_prefix"`     // Ship name prefix (e.g., "UEF")
	AllowsRecruit  bool   `json:"allows_recruit"`  // Can players join missions for them
	PatrolStrength int    `json:"patrol_strength"` // 1-10, how strong patrols are

	// Trading
	PrimaryExport []string `json:"primary_export"` // Commodities they export
	PrimaryImport []string `json:"primary_import"` // Commodities they import
	IllegalGoods  []string `json:"illegal_goods"`  // Contraband in their space
	TradingBonus  float64  `json:"trading_bonus"`  // Price modifier for trading

	// Special
	IsAlien       bool `json:"is_alien"`        // Alien species
	IsPlayerStart bool `json:"is_player_start"` // Players can start here
}

// Standard NPC Factions
var StandardNPCFactions = []NPCFaction{
	{
		ID:             "united_earth_federation",
		Name:           "United Earth Federation",
		ShortName:      "UEF",
		Description:    "The primary human government, controlling Earth and the core systems",
		Lore:           "Formed in 2187 after the Unification Wars, the UEF represents humanity's first truly unified government. Bureaucratic but stable, the Federation maintains order in the core systems through a democratic council and a powerful navy.",
		Color:          "#0066CC",
		Flag:           "⊕",
		HomeSystem:     "Sol",
		CoreSystems:    []string{"Sol", "Alpha Centauri", "Tau Ceti", "Epsilon Eridani"},
		Influence:      []string{"Sirius", "Procyon", "Altair", "Vega"},
		Government:     "federal_democracy",
		Economy:        "mixed",
		Military:       "defensive",
		Technology:     8,
		Stance:         "diplomatic",
		Traits:         []string{"bureaucratic", "lawful", "stable", "populous"},
		Allies:         []string{"republic_of_mars"},
		Enemies:        []string{"crimson_collective"},
		Neutral:        []string{"free_traders_guild", "frontier_worlds", "auroran_empire"},
		StartingRep:    10,
		ShipPrefix:     "UEF",
		AllowsRecruit:  true,
		PatrolStrength: 7,
		PrimaryExport:  []string{"electronics", "machinery", "medicine"},
		PrimaryImport:  []string{"ore", "food", "luxuries"},
		IllegalGoods:   []string{"narcotics", "slaves", "weapons"},
		TradingBonus:   1.0,
		IsAlien:        false,
		IsPlayerStart:  true,
	},
	{
		ID:             "republic_of_mars",
		Name:           "Republic of Mars",
		ShortName:      "ROM",
		Description:    "The independent Martian government, industrial powerhouse of the core",
		Lore:           "Mars declared independence from Earth in 2215, sparking a brief but intense conflict. Now a close ally of the UEF, the Martian Republic focuses on industrial production and terraforming technology. Their shipyards are the finest in human space.",
		Color:          "#CC3300",
		Flag:           "♂",
		HomeSystem:     "Sol",
		CoreSystems:    []string{"Sol"},
		Influence:      []string{"Alpha Centauri", "Epsilon Eridani", "Tau Ceti"},
		Government:     "republic",
		Economy:        "capitalist",
		Military:       "defensive",
		Technology:     9,
		Stance:         "diplomatic",
		Traits:         []string{"industrial", "innovative", "proud", "wealthy"},
		Allies:         []string{"united_earth_federation"},
		Enemies:        []string{},
		Neutral:        []string{"free_traders_guild", "frontier_worlds", "crimson_collective", "auroran_empire"},
		StartingRep:    5,
		ShipPrefix:     "RMS",
		AllowsRecruit:  true,
		PatrolStrength: 8,
		PrimaryExport:  []string{"machinery", "weapons", "ships"},
		PrimaryImport:  []string{"ore", "food", "water"},
		IllegalGoods:   []string{"narcotics", "slaves"},
		TradingBonus:   1.1,
		IsAlien:        false,
		IsPlayerStart:  true,
	},
	{
		ID:             "free_traders_guild",
		Name:           "Free Traders Guild",
		ShortName:      "FTG",
		Description:    "A loose confederation of independent traders and merchant stations",
		Lore:           "Neither government nor corporation, the Free Traders Guild is a cooperative of independent merchants who control key trade routes. They maintain neutrality in conflicts, valuing profit and freedom above all. Their stations are safe havens for traders of all allegiances.",
		Color:          "#FFB300",
		Flag:           "¤",
		HomeSystem:     "Sirius",
		CoreSystems:    []string{"Sirius", "Procyon"},
		Influence:      []string{"Alpha Centauri", "Tau Ceti", "Altair", "many outer systems"},
		Government:     "guild_council",
		Economy:        "capitalist",
		Military:       "defensive",
		Technology:     7,
		Stance:         "neutral",
		Traits:         []string{"mercantile", "neutral", "opportunistic", "wealthy"},
		Allies:         []string{},
		Enemies:        []string{"crimson_collective"},
		Neutral:        []string{"united_earth_federation", "republic_of_mars", "frontier_worlds", "auroran_empire"},
		StartingRep:    0,
		ShipPrefix:     "FTG",
		AllowsRecruit:  true,
		PatrolStrength: 5,
		PrimaryExport:  []string{"all commodities"},
		PrimaryImport:  []string{"all commodities"},
		IllegalGoods:   []string{}, // They trade almost anything
		TradingBonus:   0.9,        // Best prices
		IsAlien:        false,
		IsPlayerStart:  true,
	},
	{
		ID:             "frontier_worlds",
		Name:           "Frontier Worlds Alliance",
		ShortName:      "FWA",
		Description:    "Independent frontier colonies, loosely organized for mutual defense",
		Lore:           "The Frontier Worlds are colonies that rejected both Earth and Martian authority. Scattered across the outer systems, they're rugged individualists who value freedom above all. Their military is weak individually, but they support each other when threatened.",
		Color:          "#00AA00",
		Flag:           "⚑",
		HomeSystem:     "Epsilon Eridani",
		CoreSystems:    []string{"various outer systems"},
		Influence:      []string{"many outer systems"},
		Government:     "loose_confederation",
		Economy:        "frontier",
		Military:       "militia",
		Technology:     5,
		Stance:         "isolationist",
		Traits:         []string{"independent", "tough", "poor", "resourceful"},
		Allies:         []string{},
		Enemies:        []string{"crimson_collective"},
		Neutral:        []string{"united_earth_federation", "republic_of_mars", "free_traders_guild", "auroran_empire"},
		StartingRep:    0,
		ShipPrefix:     "FWS",
		AllowsRecruit:  true,
		PatrolStrength: 3,
		PrimaryExport:  []string{"ore", "food", "water"},
		PrimaryImport:  []string{"machinery", "medicine", "weapons"},
		IllegalGoods:   []string{"slaves"},
		TradingBonus:   1.0,
		IsAlien:        false,
		IsPlayerStart:  true,
	},
	{
		ID:             "crimson_collective",
		Name:           "Crimson Collective",
		ShortName:      "Crimson",
		Description:    "Pirate confederation and black marketeers operating in lawless space",
		Lore:           "Born from the chaos of the Unification Wars, the Crimson Collective is a loose alliance of pirates, smugglers, and outlaws. They control several asteroid bases in the outer systems and prey on merchant shipping. The UEF has declared them terrorists, but they're too dispersed to eliminate.",
		Color:          "#990000",
		Flag:           "☠",
		HomeSystem:     "unknown",
		CoreSystems:    []string{},
		Influence:      []string{"outer systems", "asteroid belts", "lawless regions"},
		Government:     "pirate_confederation",
		Economy:        "black_market",
		Military:       "raiders",
		Technology:     6,
		Stance:         "aggressive",
		Traits:         []string{"lawless", "dangerous", "opportunistic", "ruthless"},
		Allies:         []string{},
		Enemies:        []string{"united_earth_federation", "republic_of_mars", "free_traders_guild", "frontier_worlds"},
		Neutral:        []string{"auroran_empire"},
		StartingRep:    -50,
		ShipPrefix:     "Crimson",
		AllowsRecruit:  false, // They attack players unless reputation is high
		PatrolStrength: 6,
		PrimaryExport:  []string{"narcotics", "stolen goods", "weapons"},
		PrimaryImport:  []string{"weapons", "fuel", "supplies"},
		IllegalGoods:   []string{}, // Nothing is illegal to them
		TradingBonus:   0.8,        // Cheap black market prices
		IsAlien:        false,
		IsPlayerStart:  false,
	},
	{
		ID:             "auroran_empire",
		Name:           "Auroran Empire",
		ShortName:      "Auroran",
		Description:    "Mysterious alien civilization at the edge of known space",
		Lore:           "First contact with the Aurorans occurred in 2245 at the edge of explored space. Their technology is advanced but different, their motives unclear. They maintain strict borders around their territory and rarely communicate with humans. Some believe they're observers, others fear they're waiting for the right moment to strike.",
		Color:          "#AA00AA",
		Flag:           "⧈",
		HomeSystem:     "Auroran Prime",
		CoreSystems:    []string{"Auroran Prime", "several unknown systems"},
		Influence:      []string{"edge systems", "beyond the frontier"},
		Government:     "empire",
		Economy:        "unknown",
		Military:       "advanced",
		Technology:     10, // Superior tech
		Stance:         "isolationist",
		Traits:         []string{"alien", "mysterious", "advanced", "inscrutable"},
		Allies:         []string{},
		Enemies:        []string{},
		Neutral:        []string{"united_earth_federation", "republic_of_mars", "free_traders_guild", "frontier_worlds", "crimson_collective"},
		StartingRep:    -10,   // Slightly suspicious of humans
		ShipPrefix:     "AVS", // Auroran Vessel
		AllowsRecruit:  false, // Very rare missions
		PatrolStrength: 9,
		PrimaryExport:  []string{"exotic_technology"},
		PrimaryImport:  []string{"cultural_artifacts", "information"},
		IllegalGoods:   []string{"weapons", "narcotics"},
		TradingBonus:   1.5, // Expensive but unique goods
		IsAlien:        true,
		IsPlayerStart:  false,
	},
}

// GetFactionByID retrieves an NPC faction by ID
func GetFactionByID(id string) *NPCFaction {
	for i := range StandardNPCFactions {
		if StandardNPCFactions[i].ID == id {
			return &StandardNPCFactions[i]
		}
	}
	return nil
}

// GetStarterFactions returns factions where players can start
func GetStarterFactions() []NPCFaction {
	var starters []NPCFaction
	for _, faction := range StandardNPCFactions {
		if faction.IsPlayerStart {
			starters = append(starters, faction)
		}
	}
	return starters
}

// IsHostileTo checks if this faction is hostile to another
func (f *NPCFaction) IsHostileTo(otherFactionID string) bool {
	for _, enemy := range f.Enemies {
		if enemy == otherFactionID {
			return true
		}
	}
	return false
}

// IsAlliedWith checks if this faction is allied with another
func (f *NPCFaction) IsAlliedWith(otherFactionID string) bool {
	for _, ally := range f.Allies {
		if ally == otherFactionID {
			return true
		}
	}
	return false
}

// GetStanding returns relationship status with another faction
func (f *NPCFaction) GetStanding(otherFactionID string) string {
	if f.IsAlliedWith(otherFactionID) {
		return "allied"
	}
	if f.IsHostileTo(otherFactionID) {
		return "hostile"
	}
	return "neutral"
}

// IsGoodLegal checks if a commodity is legal in this faction's space
func (f *NPCFaction) IsGoodLegal(commodityID string) bool {
	for _, illegal := range f.IllegalGoods {
		if illegal == commodityID {
			return false
		}
	}
	return true
}

// ReputationTier returns reputation tier name
func GetReputationTier(rep int) string {
	switch {
	case rep >= 75:
		return "Beloved"
	case rep >= 50:
		return "Respected"
	case rep >= 25:
		return "Friendly"
	case rep >= 10:
		return "Liked"
	case rep > -10:
		return "Neutral"
	case rep > -25:
		return "Disliked"
	case rep > -50:
		return "Unfriendly"
	case rep > -75:
		return "Hostile"
	default:
		return "Hated"
	}
}

// CanDockAt checks if a reputation level allows docking
func CanDockAt(rep int) bool {
	return rep > -50 // Can dock if not openly hostile
}

// WillAttackOnSight checks if faction attacks player on sight
func WillAttackOnSight(rep int) bool {
	return rep <= -75
}
