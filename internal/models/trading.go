// File: internal/models/trading.go
// Project: Terminal Velocity
// Description: Data models for trading
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import "github.com/google/uuid"

// Commodity represents a tradeable good
type Commodity struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BasePrice   int64    `json:"base_price"`
	Category    string   `json:"category"`   // food, electronics, weapons, luxuries, etc.
	IllegalIn   []string `json:"illegal_in"` // Government IDs where this is contraband
	TechLevel   int      `json:"tech_level"` // Minimum tech level to trade
}

// MarketPrice represents commodity pricing at a specific planet
type MarketPrice struct {
	PlanetID    uuid.UUID `json:"planet_id"`
	CommodityID string    `json:"commodity_id"`
	BuyPrice    int64     `json:"buy_price"`   // What planet pays for commodity
	SellPrice   int64     `json:"sell_price"`  // What planet sells commodity for
	Stock       int       `json:"stock"`       // Available quantity
	Demand      int       `json:"demand"`      // How much they want to buy
	LastUpdate  int64     `json:"last_update"` // Unix timestamp
}

// Commodity categories
const (
	CategoryFood        = "food"
	CategoryElectronics = "electronics"
	CategoryWeapons     = "weapons"
	CategoryLuxuries    = "luxuries"
	CategoryIndustrial  = "industrial"
	CategoryMedical     = "medical"
	CategoryOre         = "ore"
	CategoryContraband  = "contraband"
)

// Standard commodities
var StandardCommodities = []Commodity{
	// Food & Basic Resources
	{
		ID:          "food",
		Name:        "Food",
		Description: "Basic foodstuffs and provisions",
		BasePrice:   50,
		Category:    CategoryFood,
		TechLevel:   1,
		IllegalIn:   []string{},
	},
	{
		ID:          "water",
		Name:        "Water",
		Description: "Purified water for consumption and industry",
		BasePrice:   30,
		Category:    CategoryFood,
		TechLevel:   1,
		IllegalIn:   []string{},
	},
	{
		ID:          "textiles",
		Name:        "Textiles",
		Description: "Fabrics and clothing materials",
		BasePrice:   60,
		Category:    CategoryFood,
		TechLevel:   2,
		IllegalIn:   []string{},
	},
	{
		ID:          "livestock",
		Name:        "Livestock",
		Description: "Live animals for food production",
		BasePrice:   120,
		Category:    CategoryFood,
		TechLevel:   1,
		IllegalIn:   []string{},
	},

	// Electronics & Technology
	{
		ID:          "electronics",
		Name:        "Electronics",
		Description: "Consumer electronics and components",
		BasePrice:   200,
		Category:    CategoryElectronics,
		TechLevel:   4,
		IllegalIn:   []string{},
	},
	{
		ID:          "computers",
		Name:        "Computers",
		Description: "Advanced computing systems",
		BasePrice:   350,
		Category:    CategoryElectronics,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
	{
		ID:          "robotics",
		Name:        "Robotics",
		Description: "Automated systems and robotic equipment",
		BasePrice:   600,
		Category:    CategoryElectronics,
		TechLevel:   7,
		IllegalIn:   []string{},
	},
	{
		ID:          "ai_cores",
		Name:        "AI Cores",
		Description: "Artificial intelligence processor cores",
		BasePrice:   1200,
		Category:    CategoryElectronics,
		TechLevel:   9,
		IllegalIn:   []string{},
	},

	// Weapons & Military
	{
		ID:          "weapons",
		Name:        "Weapons",
		Description: "Small arms and personal defense systems",
		BasePrice:   500,
		Category:    CategoryWeapons,
		TechLevel:   3,
		IllegalIn:   []string{"pacifist_union"},
	},
	{
		ID:          "ammunition",
		Name:        "Ammunition",
		Description: "Various ammunition types",
		BasePrice:   180,
		Category:    CategoryWeapons,
		TechLevel:   2,
		IllegalIn:   []string{"pacifist_union"},
	},
	{
		ID:          "explosives",
		Name:        "Explosives",
		Description: "Industrial and military explosives",
		BasePrice:   400,
		Category:    CategoryWeapons,
		TechLevel:   3,
		IllegalIn:   []string{"pacifist_union"},
	},
	{
		ID:          "military_hardware",
		Name:        "Military Hardware",
		Description: "Advanced military equipment",
		BasePrice:   900,
		Category:    CategoryWeapons,
		TechLevel:   6,
		IllegalIn:   []string{"pacifist_union", "independent"},
	},

	// Medical
	{
		ID:          "medicine",
		Name:        "Medicine",
		Description: "Pharmaceuticals and medical supplies",
		BasePrice:   150,
		Category:    CategoryMedical,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
	{
		ID:          "medical_equipment",
		Name:        "Medical Equipment",
		Description: "Surgical tools and diagnostic devices",
		BasePrice:   380,
		Category:    CategoryMedical,
		TechLevel:   6,
		IllegalIn:   []string{},
	},
	{
		ID:          "vaccines",
		Name:        "Vaccines",
		Description: "Disease prevention and treatment",
		BasePrice:   280,
		Category:    CategoryMedical,
		TechLevel:   7,
		IllegalIn:   []string{},
	},
	{
		ID:          "organs",
		Name:        "Bio-Organs",
		Description: "Synthetic replacement organs",
		BasePrice:   2500,
		Category:    CategoryMedical,
		TechLevel:   8,
		IllegalIn:   []string{},
	},

	// Luxury Goods
	{
		ID:          "luxuries",
		Name:        "Luxury Goods",
		Description: "Fine wines, art, and entertainment",
		BasePrice:   400,
		Category:    CategoryLuxuries,
		TechLevel:   6,
		IllegalIn:   []string{},
	},
	{
		ID:          "jewelry",
		Name:        "Jewelry",
		Description: "Precious gems and metals",
		BasePrice:   800,
		Category:    CategoryLuxuries,
		TechLevel:   4,
		IllegalIn:   []string{},
	},
	{
		ID:          "art",
		Name:        "Art",
		Description: "Fine art and cultural artifacts",
		BasePrice:   1000,
		Category:    CategoryLuxuries,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
	{
		ID:          "exotic_animals",
		Name:        "Exotic Animals",
		Description: "Rare creatures from distant worlds",
		BasePrice:   1500,
		Category:    CategoryLuxuries,
		TechLevel:   3,
		IllegalIn:   []string{},
	},

	// Industrial
	{
		ID:          "machinery",
		Name:        "Machinery",
		Description: "Industrial equipment and tools",
		BasePrice:   250,
		Category:    CategoryIndustrial,
		TechLevel:   4,
		IllegalIn:   []string{},
	},
	{
		ID:          "construction_materials",
		Name:        "Construction Materials",
		Description: "Building supplies and materials",
		BasePrice:   90,
		Category:    CategoryIndustrial,
		TechLevel:   2,
		IllegalIn:   []string{},
	},
	{
		ID:          "power_cells",
		Name:        "Power Cells",
		Description: "High-capacity energy storage",
		BasePrice:   320,
		Category:    CategoryIndustrial,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
	{
		ID:          "industrial_chemicals",
		Name:        "Industrial Chemicals",
		Description: "Chemical compounds for manufacturing",
		BasePrice:   150,
		Category:    CategoryIndustrial,
		TechLevel:   4,
		IllegalIn:   []string{},
	},

	// Raw Materials (Ore)
	{
		ID:          "ore",
		Name:        "Metal Ore",
		Description: "Unprocessed mineral ores",
		BasePrice:   80,
		Category:    CategoryOre,
		TechLevel:   2,
		IllegalIn:   []string{},
	},
	{
		ID:          "precious_metals",
		Name:        "Precious Metals",
		Description: "Gold, platinum, and rare earth metals",
		BasePrice:   450,
		Category:    CategoryOre,
		TechLevel:   3,
		IllegalIn:   []string{},
	},
	{
		ID:          "crystals",
		Name:        "Crystals",
		Description: "Rare crystals for advanced technology",
		BasePrice:   350,
		Category:    CategoryOre,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
	{
		ID:          "radioactives",
		Name:        "Radioactive Materials",
		Description: "Uranium and other radioactive elements",
		BasePrice:   600,
		Category:    CategoryOre,
		TechLevel:   6,
		IllegalIn:   []string{"pacifist_union"},
	},

	// Contraband (High Risk, High Reward - 20-25% price increase)
	{
		ID:          "narcotics",
		Name:        "Narcotics",
		Description: "Illegal drugs and stimulants",
		BasePrice:   1000,
		Category:    CategoryContraband,
		TechLevel:   4,
		IllegalIn:   []string{"federation", "republic", "corporate"},
	},
	{
		ID:          "slaves",
		Name:        "Slaves",
		Description: "Indentured laborers (highly illegal)",
		BasePrice:   2000,
		Category:    CategoryContraband,
		TechLevel:   1,
		IllegalIn:   []string{"federation", "republic", "pacifist_union", "independent"},
	},
	{
		ID:          "stolen_goods",
		Name:        "Stolen Goods",
		Description: "Hot merchandise from piracy",
		BasePrice:   800,
		Category:    CategoryContraband,
		TechLevel:   1,
		IllegalIn:   []string{"federation", "republic", "corporate"},
	},
	{
		ID:          "alien_artifacts",
		Name:        "Alien Artifacts",
		Description: "Mysterious items of unknown origin",
		BasePrice:   3750,
		Category:    CategoryContraband,
		TechLevel:   1,
		IllegalIn:   []string{"federation"},
	},
	{
		ID:          "military_intel",
		Name:        "Military Intelligence",
		Description: "Classified military data (extremely illegal)",
		BasePrice:   7500,
		Category:    CategoryContraband,
		TechLevel:   7,
		IllegalIn:   []string{"federation", "republic", "corporate", "independent"},
	},
}

// CalculateProfit calculates profit from trading
func CalculateProfit(buyPrice, sellPrice int64, quantity int) int64 {
	return (sellPrice - buyPrice) * int64(quantity)
}

// IsIllegal checks if commodity is illegal in a government
func (c *Commodity) IsIllegal(governmentID string) bool {
	for _, id := range c.IllegalIn {
		if id == governmentID {
			return true
		}
	}
	return false
}

// GetPriceModifier calculates price modifier based on tech level difference
func GetPriceModifier(commodityTechLevel, planetTechLevel int, isBuying bool) float64 {
	diff := planetTechLevel - commodityTechLevel

	if isBuying {
		// High tech planets pay less for low tech goods
		if diff > 0 {
			return 1.0 - (float64(diff) * 0.1)
		}
		// Low tech planets pay more for high tech goods
		return 1.0 + (float64(-diff) * 0.15)
	}
	// High tech planets sell high tech goods cheaper (improved from 0.05 to 0.07)
	if diff >= 0 {
		return 1.0 - (float64(diff) * 0.07)
	}
	// Low tech planets sell high tech goods at premium
	return 1.0 + (float64(-diff) * 0.2)
}

// GetCommodityByID finds a commodity by its ID
func GetCommodityByID(id string) *Commodity {
	for i := range StandardCommodities {
		if StandardCommodities[i].ID == id {
			return &StandardCommodities[i]
		}
	}
	return nil
}

// GetCommoditiesByCategory returns all commodities in a category
func GetCommoditiesByCategory(category string) []Commodity {
	var result []Commodity
	for _, commodity := range StandardCommodities {
		if commodity.Category == category {
			result = append(result, commodity)
		}
	}
	return result
}

// GetLegalCommoditiesForSystem returns commodities legal in a system
func GetLegalCommoditiesForSystem(governmentID string) []Commodity {
	var result []Commodity
	for _, commodity := range StandardCommodities {
		if !commodity.IsIllegal(governmentID) {
			result = append(result, commodity)
		}
	}
	return result
}

// GetAvailableCommoditiesAtTechLevel returns commodities available at tech level
func GetAvailableCommoditiesAtTechLevel(techLevel int) []Commodity {
	var result []Commodity
	for _, commodity := range StandardCommodities {
		if commodity.TechLevel <= techLevel {
			result = append(result, commodity)
		}
	}
	return result
}
