// File: internal/game/trading/pricing.go
// Project: Terminal Velocity
// Description: Trading and pricing system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package trading implements a dynamic commodity pricing and market simulation system.
//
// The economy is based on realistic supply/demand mechanics with additional factors:
//
// Price Calculation:
//
//	finalPrice = basePrice × techModifier × supplyDemandModifier
//	buybackPrice = sellPrice × (0.60-0.80)  // 60-80% of sell price
//
// Tech Level Effects:
//   - Systems can only produce commodities at or below their tech level
//   - Higher tech systems have better production (more stock)
//   - Price modifiers handled by models.GetPriceModifier()
//
// Supply/Demand Mechanics:
//   - Stock decreases when players buy, increases when players sell
//   - Demand increases when players buy, decreases when players sell
//   - Markets naturally recover toward equilibrium over time (5% per hour)
//   - Random market events (5% chance per hour): supply shocks, demand surges, etc.
//
// Trading Effects:
//   - Player purchases: Stock -N, Demand +N/2 (demand increases slower)
//   - Player sales: Stock +N, Demand -N/3 (demand decreases even slower)
//   - Asymmetric response rates prevent market manipulation
//
// This creates an economy where:
//   - Profitable trade routes exist (buy low tech, sell high tech)
//   - Markets respond to player activity
//   - Excessive trading impacts prices (prevents infinite money exploits)
//   - Markets gradually return to equilibrium when left alone
package trading

import (
	"math"
	"math/rand"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// PricingEngine handles dynamic commodity price calculation and market simulation.
//
// The engine maintains a random source for:
//   - Buyback price variance (60-80% of sell price)
//   - Initial stock/demand generation (±30-40% variance)
//   - Random market events (supply shocks, demand surges)
//
// Thread Safety: NOT thread-safe. The internal rand.Rand must not be shared across
// goroutines. Create separate engines for concurrent use or synchronize access.
type PricingEngine struct {
	rand *rand.Rand  // Random source for price variance and market events
}

// NewPricingEngine creates a new pricing engine with its own random source.
//
// Each engine gets a time-seeded random source, so creating multiple engines
// simultaneously will produce different results (good for market variety).
//
// Returns:
//
//	Initialized pricing engine ready for price calculations
func NewPricingEngine() *PricingEngine {
	return &PricingEngine{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CalculateMarketPrice calculates buy and sell prices for a commodity at a planet's market.
//
// This is the core pricing function that combines all economic factors:
//
// Price Formula:
//
//	sellPrice = basePrice × techModifier × supplyDemandModifier
//	buyPrice = sellPrice × randomPercent(0.60-0.80)
//
// Tech Modifier:
//   - Determined by models.GetPriceModifier(commodityTech, planetTech, false)
//   - Higher planet tech = better prices for manufactured goods
//   - Lower planet tech = worse prices (scarcity/poor infrastructure)
//
// Supply/Demand Modifier:
//   - Oversupply (stock > demand): Price decreases up to -70%
//   - Undersupply (stock < demand): Price increases up to +150%
//   - No supply + demand: Price up to +200% + (demand × 10%)
//   - No demand + supply: Price down to 30%
//
// Buyback Mechanics:
//   - Players sell to planets at 60-80% of the sell price
//   - Randomized to prevent perfect arbitrage
//   - Creates natural profit margins for trading
//
// Parameters:
//   - commodity: The commodity being priced (provides basePrice)
//   - planet: The planet market (provides techLevel)
//   - stock: Current inventory of this commodity at the planet
//   - demand: Current demand for this commodity at the planet
//
// Returns:
//   - buyPrice: What the planet will pay when player sells to it
//   - sellPrice: What the planet charges when player buys from it
//
// Both prices are clamped to minimum of 1 credit to prevent zero/negative prices.
func (e *PricingEngine) CalculateMarketPrice(
	commodity *models.Commodity,
	planet *models.Planet,
	stock int,
	demand int,
) (buyPrice, sellPrice int64) {
	basePrice := commodity.BasePrice

	// Apply tech level modifier
	techModifier := models.GetPriceModifier(commodity.TechLevel, planet.TechLevel, false)

	// Apply supply/demand modifier
	supplyDemandModifier := e.calculateSupplyDemandModifier(stock, demand)

	// Calculate sell price (what planet sells for)
	sellPrice = int64(float64(basePrice) * techModifier * supplyDemandModifier)

	// Buy price is typically 60-80% of sell price
	buybackPercent := 0.60 + (e.rand.Float64() * 0.20) // 60-80%
	buyPrice = int64(float64(sellPrice) * buybackPercent)

	// Ensure minimum prices
	if sellPrice < 1 {
		sellPrice = 1
	}
	if buyPrice < 1 {
		buyPrice = 1
	}

	return buyPrice, sellPrice
}

// calculateSupplyDemandModifier calculates price modifier based on supply and demand.
//
// This implements realistic economic scarcity/surplus pricing:
//
// Algorithm:
//  1. Handle edge cases:
//     - Zero stock + positive demand = scarce (2.0+ multiplier)
//     - Zero demand + positive stock = worthless (0.3 multiplier)
//  2. Calculate supply/demand ratio
//  3. If oversupply (ratio > 1.0):
//     - Price decreases: 1.0 - min(ratio-1.0, 1.0) × 0.4
//     - Floor at 0.3 (30% of base price)
//  4. If undersupply (ratio < 1.0):
//     - Price increases: 1.0 + (1.0 - ratio) × 0.8
//     - Cap at 2.5 (250% of base price)
//
// Price Ranges:
//   - Extreme oversupply: 0.3× (70% discount)
//   - Balanced market: 1.0× (no modifier)
//   - Extreme scarcity: 2.5×+ (150%+ markup)
//
// Examples:
//   - stock=100, demand=50: ratio=2.0, modifier=0.6 (40% discount due to surplus)
//   - stock=50, demand=100: ratio=0.5, modifier=1.4 (40% markup due to scarcity)
//   - stock=0, demand=100: modifier=2.0+ (extreme scarcity)
//
// Parameters:
//   - stock: Current inventory
//   - demand: Current demand
//
// Returns:
//
//	Price multiplier (0.3 to 2.5+)
func (e *PricingEngine) calculateSupplyDemandModifier(stock, demand int) float64 {
	if stock == 0 && demand > 0 {
		// No supply, high demand = very expensive
		return 2.0 + (float64(demand) * 0.1)
	}

	if demand == 0 && stock > 0 {
		// No demand, high supply = very cheap
		return 0.3
	}

	// Calculate supply/demand ratio
	ratio := float64(stock) / float64(demand+1) // +1 to avoid division by zero

	if ratio > 1.0 {
		// Oversupply - prices drop
		modifier := 1.0 - (math.Min(ratio-1.0, 1.0) * 0.4) // Max 40% reduction
		return math.Max(modifier, 0.3)                     // Floor at 30%
	}

	// Under-supply - prices rise
	modifier := 1.0 + ((1.0 - ratio) * 0.8) // Up to 80% increase
	return math.Min(modifier, 2.5)          // Cap at 250%
}

// GenerateInitialStock generates initial stock levels for a commodity at a planet
func (e *PricingEngine) GenerateInitialStock(commodity *models.Commodity, planet *models.Planet) int {
	// Tech level difference affects production
	techDiff := planet.TechLevel - commodity.TechLevel

	if techDiff < 0 {
		// Cannot produce this commodity
		return 0
	}

	// Base stock increases with tech advantage
	baseStock := 100 + (techDiff * 20)

	// Add randomness
	variance := int(float64(baseStock) * 0.3) // ±30%
	stock := baseStock + e.rand.Intn(variance*2) - variance

	if stock < 0 {
		stock = 0
	}

	return stock
}

// GenerateInitialDemand generates initial demand levels for a commodity at a planet
func (e *PricingEngine) GenerateInitialDemand(commodity *models.Commodity, planet *models.Planet) int {
	// Base demand on population
	populationFactor := float64(planet.Population) / 1000000.0 // Per million people

	// Different categories have different demand patterns
	categoryMultiplier := e.getCategoryDemandMultiplier(commodity.Category)

	baseDemand := int(populationFactor * categoryMultiplier)

	// Tech level affects demand
	techDiff := planet.TechLevel - commodity.TechLevel
	if techDiff < -2 {
		// Much lower tech - reduced demand for advanced goods
		baseDemand = int(float64(baseDemand) * 0.5)
	} else if techDiff > 2 {
		// Much higher tech - increased demand for basic goods
		baseDemand = int(float64(baseDemand) * 1.3)
	}

	// Add randomness
	variance := int(float64(baseDemand) * 0.4) // ±40%
	demand := baseDemand + e.rand.Intn(variance*2) - variance

	if demand < 10 {
		demand = 10 // Minimum demand
	}

	return demand
}

// getCategoryDemandMultiplier returns demand multiplier for a commodity category
func (e *PricingEngine) getCategoryDemandMultiplier(category string) float64 {
	switch category {
	case models.CategoryFood:
		return 150.0 // High demand for essentials
	case models.CategoryMedical:
		return 100.0 // Steady demand
	case models.CategoryElectronics:
		return 80.0
	case models.CategoryIndustrial:
		return 70.0
	case models.CategoryOre:
		return 60.0 // Lower demand for raw materials
	case models.CategoryWeapons:
		return 50.0 // Niche demand
	case models.CategoryLuxuries:
		return 30.0 // Low demand, high margin
	case models.CategoryContraband:
		return 20.0 // Very low legal demand
	default:
		return 50.0
	}
}

// UpdateMarketPrice updates market prices based on trading activity
func (e *PricingEngine) UpdateMarketPrice(
	price *models.MarketPrice,
	commodity *models.Commodity,
	planet *models.Planet,
	quantityTraded int,
	wasBuying bool, // True if player was buying (planet selling)
) {
	if wasBuying {
		// Player bought from planet - decrease stock, increase demand
		price.Stock -= quantityTraded
		if price.Stock < 0 {
			price.Stock = 0
		}
		price.Demand += quantityTraded / 2 // Demand increases slower
	} else {
		// Player sold to planet - increase stock, decrease demand
		price.Stock += quantityTraded
		price.Demand -= quantityTraded / 3 // Demand decreases slower
		if price.Demand < 10 {
			price.Demand = 10 // Minimum demand
		}
	}

	// Recalculate prices
	price.BuyPrice, price.SellPrice = e.CalculateMarketPrice(
		commodity,
		planet,
		price.Stock,
		price.Demand,
	)

	// Update timestamp
	price.LastUpdate = time.Now().Unix()
}

// SimulateMarketTick simulates market evolution over time without player interaction.
//
// This creates a living economy where markets naturally recover from player trading
// and experience random events. Called when players return to a system after time away.
//
// Market Recovery Algorithm (per hour):
//  1. Calculate target stock/demand (equilibrium values)
//  2. Move current values 5% toward target:
//     - stock += (target - stock) × 0.05
//     - demand += (target - demand) × 0.05
//  3. 5% chance of random market event:
//     - Supply shock: Stock × 0.7 (production failure)
//     - Demand surge: Demand × 1.3 (increased consumption)
//     - Production boom: Stock × 1.4 (good harvest/production)
//     - Demand drop: Demand × 0.8 (reduced consumption)
//  4. Recalculate prices
//
// Recovery Rate:
//   - 5% per hour means ~14 hours to recover 50% from disturbance
//   - Full recovery takes days, giving players time to exploit opportunities
//   - Gradual recovery prevents instant market resets
//
// Random Events:
//   - Keep markets dynamic even without player interaction
//   - Create trading opportunities for returning players
//   - Prevent markets from becoming perfectly stable
//
// Parameters:
//   - price: Market price data (modified in-place)
//   - commodity: Commodity being traded
//   - planet: Planet market location
//   - deltaHours: Hours elapsed since last update
//
// Thread Safety: NOT thread-safe. Caller must synchronize access to price.
func (e *PricingEngine) SimulateMarketTick(
	price *models.MarketPrice,
	commodity *models.Commodity,
	planet *models.Planet,
	deltaHours int,
) {
	// Natural market recovery over time
	for i := 0; i < deltaHours; i++ {
		// Stock slowly regenerates
		targetStock := e.GenerateInitialStock(commodity, planet)
		if price.Stock < targetStock {
			price.Stock += int(float64(targetStock-price.Stock) * 0.05) // 5% recovery per hour
		} else if price.Stock > targetStock {
			price.Stock -= int(float64(price.Stock-targetStock) * 0.05)
		}

		// Demand slowly normalizes
		targetDemand := e.GenerateInitialDemand(commodity, planet)
		if price.Demand < targetDemand {
			price.Demand += int(float64(targetDemand-price.Demand) * 0.05)
		} else if price.Demand > targetDemand {
			price.Demand -= int(float64(price.Demand-targetDemand) * 0.05)
		}

		// Add random market events (5% chance per hour)
		if e.rand.Float64() < 0.05 {
			e.applyRandomMarketEvent(price, commodity)
		}
	}

	// Recalculate prices
	price.BuyPrice, price.SellPrice = e.CalculateMarketPrice(
		commodity,
		planet,
		price.Stock,
		price.Demand,
	)

	price.LastUpdate = time.Now().Unix()
}

// applyRandomMarketEvent applies a random market event
func (e *PricingEngine) applyRandomMarketEvent(price *models.MarketPrice, commodity *models.Commodity) {
	eventType := e.rand.Intn(4)

	switch eventType {
	case 0: // Supply shock - stock drops
		price.Stock = int(float64(price.Stock) * 0.7)
	case 1: // Demand surge
		price.Demand = int(float64(price.Demand) * 1.3)
	case 2: // Production boom - stock increases
		price.Stock = int(float64(price.Stock) * 1.4)
	case 3: // Demand drop
		price.Demand = int(float64(price.Demand) * 0.8)
	}

	// Ensure minimums
	if price.Stock < 0 {
		price.Stock = 0
	}
	if price.Demand < 10 {
		price.Demand = 10
	}
}

// CalculateProfitMargin calculates the profit margin for a commodity at two locations
func CalculateProfitMargin(buyPrice, sellPrice int64, quantity int) (profit int64, margin float64) {
	totalCost := buyPrice * int64(quantity)
	totalRevenue := sellPrice * int64(quantity)
	profit = totalRevenue - totalCost

	if totalCost > 0 {
		margin = float64(profit) / float64(totalCost) * 100.0
	}

	return profit, margin
}
