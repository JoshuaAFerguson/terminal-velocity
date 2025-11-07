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
		Description: "Purified water",
		BasePrice:   30,
		Category:    CategoryFood,
		TechLevel:   1,
		IllegalIn:   []string{},
	},
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
		ID:          "weapons",
		Name:        "Weapons",
		Description: "Small arms and ammunition",
		BasePrice:   500,
		Category:    CategoryWeapons,
		TechLevel:   3,
		IllegalIn:   []string{"pacifist_union"},
	},
	{
		ID:          "medicine",
		Name:        "Medicine",
		Description: "Medical supplies and pharmaceuticals",
		BasePrice:   150,
		Category:    CategoryMedical,
		TechLevel:   5,
		IllegalIn:   []string{},
	},
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
		ID:          "machinery",
		Name:        "Machinery",
		Description: "Industrial equipment and tools",
		BasePrice:   250,
		Category:    CategoryIndustrial,
		TechLevel:   4,
		IllegalIn:   []string{},
	},
	{
		ID:          "ore",
		Name:        "Raw Ore",
		Description: "Unprocessed mineral ores",
		BasePrice:   80,
		Category:    CategoryOre,
		TechLevel:   2,
		IllegalIn:   []string{},
	},
	{
		ID:          "narcotics",
		Name:        "Narcotics",
		Description: "Illegal drugs and stimulants",
		BasePrice:   800,
		Category:    CategoryContraband,
		TechLevel:   4,
		IllegalIn:   []string{"federation", "republic"},
	},
	{
		ID:          "slaves",
		Name:        "Slaves",
		Description: "Indentured laborers (highly illegal)",
		BasePrice:   1500,
		Category:    CategoryContraband,
		TechLevel:   1,
		IllegalIn:   []string{"federation", "republic", "pacifist_union"},
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
	} else {
		// High tech planets sell high tech goods cheaper
		if diff >= 0 {
			return 1.0 - (float64(diff) * 0.05)
		}
		// Low tech planets sell high tech goods at premium
		return 1.0 + (float64(-diff) * 0.2)
	}
}
