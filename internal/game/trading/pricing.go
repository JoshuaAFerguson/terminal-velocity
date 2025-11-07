// File: internal/game/trading/pricing.go
// Project: Terminal Velocity
// Description: Trading and pricing system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package trading

import (
	"math"
	"math/rand"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// PricingEngine handles dynamic price calculation

var log = logger.WithComponent("Trading")

type PricingEngine struct {
	rand *rand.Rand
}

// NewPricingEngine creates a new pricing engine
func NewPricingEngine() *PricingEngine {
	return &PricingEngine{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CalculateMarketPrice calculates buy and sell prices for a commodity at a planet
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

// calculateSupplyDemandModifier calculates price modifier based on supply and demand
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

// SimulateMarketTick simulates market changes over time
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
