// File: internal/database/market_repository.go
// Project: Terminal Velocity
// Description: Repository for market prices and commodity trading economy
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/errors"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// MarketRepository handles all database operations for market prices.
//
// Manages the dynamic trading economy:
//   - Market prices for commodities at each planet
//   - Stock and demand levels
//   - Price updates based on trading activity
//   - Stale market detection for price refresh
//
// Data model:
//   - Market prices stored per (planet_id, commodity_id) pair
//   - Prices have buy/sell values, stock, and demand
//   - Last update timestamp for staleness tracking
//
// Thread-safety:
//   - All methods are thread-safe
//   - Uses UPSERT for atomic price updates
type MarketRepository struct {
	db *DB // Database connection pool
}

// NewMarketRepository creates a new market repository
func NewMarketRepository(db *DB) *MarketRepository {
	return &MarketRepository{db: db}
}

// GetMarketPrice retrieves market price for a commodity at a planet
func (r *MarketRepository) GetMarketPrice(ctx context.Context, planetID uuid.UUID, commodityID string) (*models.MarketPrice, error) {
	query := `
		SELECT planet_id, commodity_id, buy_price, sell_price, stock, demand, last_update
		FROM market_prices
		WHERE planet_id = $1 AND commodity_id = $2
	`

	var price models.MarketPrice
	err := r.db.QueryRowContext(ctx, query, planetID, commodityID).Scan(
		&price.PlanetID,
		&price.CommodityID,
		&price.BuyPrice,
		&price.SellPrice,
		&price.Stock,
		&price.Demand,
		&price.LastUpdate,
	)

	if err == sql.ErrNoRows {
		log.Debug("Market price not found: planet_id=%s, commodity_id=%s", planetID, commodityID)
		return nil, ErrMarketPriceNotFound
	}
	if err != nil {
		errors.RecordGlobalError("market_repository", "query_price", err)
		log.Error("Failed to query market price: planet_id=%s, commodity_id=%s, error=%v", planetID, commodityID, err)
		return nil, fmt.Errorf("failed to query market price: %w", err)
	}

	log.Debug("Retrieved market price: planet_id=%s, commodity_id=%s", planetID, commodityID)
	return &price, nil
}

// GetMarketPricesForPlanet retrieves all market prices for a planet
func (r *MarketRepository) GetMarketPricesForPlanet(ctx context.Context, planetID uuid.UUID) ([]*models.MarketPrice, error) {
	query := `
		SELECT planet_id, commodity_id, buy_price, sell_price, stock, demand, last_update
		FROM market_prices
		WHERE planet_id = $1
		ORDER BY commodity_id
	`

	rows, err := r.db.QueryContext(ctx, query, planetID)
	if err != nil {
		errors.RecordGlobalError("market_repository", "query_planet_prices", err)
		log.Error("Failed to query market prices for planet: planet_id=%s, error=%v", planetID, err)
		return nil, fmt.Errorf("failed to query market prices: %w", err)
	}
	defer rows.Close()

	var prices []*models.MarketPrice
	for rows.Next() {
		var price models.MarketPrice
		err := rows.Scan(
			&price.PlanetID,
			&price.CommodityID,
			&price.BuyPrice,
			&price.SellPrice,
			&price.Stock,
			&price.Demand,
			&price.LastUpdate,
		)
		if err != nil {
			log.Error("Failed to scan market price row: planet_id=%s, error=%v", planetID, err)
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		log.Error("Error iterating market prices: planet_id=%s, error=%v", planetID, err)
		return nil, fmt.Errorf("error iterating market prices: %w", err)
	}

	log.Debug("Retrieved %d market prices for planet: planet_id=%s", len(prices), planetID)
	return prices, nil
}

// UpsertMarketPrice inserts or updates a market price atomically.
//
// Uses PostgreSQL UPSERT (INSERT ... ON CONFLICT DO UPDATE) to create or
// update market prices. This is the primary method for updating the economy
// based on trading activity and market fluctuations.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - price: Market price with all fields populated
//
// Returns:
//   - error: Database error
//
// Thread-safety:
//   - Safe for concurrent price updates
//   - UPSERT ensures atomicity
func (r *MarketRepository) UpsertMarketPrice(ctx context.Context, price *models.MarketPrice) error {
	query := `
		INSERT INTO market_prices (planet_id, commodity_id, buy_price, sell_price, stock, demand, last_update)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (planet_id, commodity_id)
		DO UPDATE SET
			buy_price = EXCLUDED.buy_price,
			sell_price = EXCLUDED.sell_price,
			stock = EXCLUDED.stock,
			demand = EXCLUDED.demand,
			last_update = EXCLUDED.last_update
	`

	_, err := r.db.ExecContext(ctx, query,
		price.PlanetID,
		price.CommodityID,
		price.BuyPrice,
		price.SellPrice,
		price.Stock,
		price.Demand,
		price.LastUpdate,
	)

	if err != nil {
		errors.RecordGlobalError("market_repository", "upsert_price", err)
		log.Error("Failed to upsert market price: planet_id=%s, commodity_id=%s, error=%v", price.PlanetID, price.CommodityID, err)
		return fmt.Errorf("failed to upsert market price: %w", err)
	}

	log.Debug("Upserted market price: planet_id=%s, commodity_id=%s", price.PlanetID, price.CommodityID)
	return nil
}

// UpdateMarketPrice updates an existing market price
func (r *MarketRepository) UpdateMarketPrice(ctx context.Context, price *models.MarketPrice) error {
	query := `
		UPDATE market_prices
		SET buy_price = $1, sell_price = $2, stock = $3, demand = $4, last_update = $5
		WHERE planet_id = $6 AND commodity_id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		price.BuyPrice,
		price.SellPrice,
		price.Stock,
		price.Demand,
		price.LastUpdate,
		price.PlanetID,
		price.CommodityID,
	)

	if err != nil {
		errors.RecordGlobalError("market_repository", "update_price", err)
		log.Error("Failed to update market price: planet_id=%s, commodity_id=%s, error=%v", price.PlanetID, price.CommodityID, err)
		return fmt.Errorf("failed to update market price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("Failed to get rows affected: planet_id=%s, commodity_id=%s, error=%v", price.PlanetID, price.CommodityID, err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Debug("Market price not found for update: planet_id=%s, commodity_id=%s", price.PlanetID, price.CommodityID)
		return ErrMarketPriceNotFound
	}

	log.Debug("Updated market price: planet_id=%s, commodity_id=%s", price.PlanetID, price.CommodityID)
	return nil
}

// DeleteMarketPrice deletes a market price
func (r *MarketRepository) DeleteMarketPrice(ctx context.Context, planetID uuid.UUID, commodityID string) error {
	query := `DELETE FROM market_prices WHERE planet_id = $1 AND commodity_id = $2`

	result, err := r.db.ExecContext(ctx, query, planetID, commodityID)
	if err != nil {
		errors.RecordGlobalError("market_repository", "delete_price", err)
		log.Error("Failed to delete market price: planet_id=%s, commodity_id=%s, error=%v", planetID, commodityID, err)
		return fmt.Errorf("failed to delete market price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("Failed to get rows affected: planet_id=%s, commodity_id=%s, error=%v", planetID, commodityID, err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Debug("Market price not found for deletion: planet_id=%s, commodity_id=%s", planetID, commodityID)
		return ErrMarketPriceNotFound
	}

	log.Debug("Deleted market price: planet_id=%s, commodity_id=%s", planetID, commodityID)
	return nil
}

// InitializePlanetMarket initializes market prices for all commodities at a planet
func (r *MarketRepository) InitializePlanetMarket(ctx context.Context, planetID uuid.UUID) error {
	// This would typically be called by the pricing engine
	// For now, just a placeholder
	return nil
}

// GetStaleMarkets returns markets that haven't been updated in a while
func (r *MarketRepository) GetStaleMarkets(ctx context.Context, olderThanSeconds int64) ([]*models.MarketPrice, error) {
	query := `
		SELECT planet_id, commodity_id, buy_price, sell_price, stock, demand, last_update
		FROM market_prices
		WHERE last_update < $1
		ORDER BY last_update ASC
		LIMIT 1000
	`

	currentTime := sql.NullInt64{Int64: olderThanSeconds, Valid: true}
	rows, err := r.db.QueryContext(ctx, query, currentTime)
	if err != nil {
		log.Error("Failed to query stale markets: olderThan=%d, error=%v", olderThanSeconds, err)
		return nil, fmt.Errorf("failed to query stale markets: %w", err)
	}
	defer rows.Close()

	var prices []*models.MarketPrice
	for rows.Next() {
		var price models.MarketPrice
		err := rows.Scan(
			&price.PlanetID,
			&price.CommodityID,
			&price.BuyPrice,
			&price.SellPrice,
			&price.Stock,
			&price.Demand,
			&price.LastUpdate,
		)
		if err != nil {
			log.Error("Failed to scan stale market price row: error=%v", err)
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		log.Error("Error iterating stale markets: error=%v", err)
		return nil, fmt.Errorf("error iterating stale markets: %w", err)
	}

	log.Debug("Found %d stale markets older than %d seconds", len(prices), olderThanSeconds)
	return prices, nil
}

// GetCommoditiesBySystemID retrieves all market prices for all planets in a system
func (r *MarketRepository) GetCommoditiesBySystemID(ctx context.Context, systemID uuid.UUID) ([]models.MarketPrice, error) {
	query := `
		SELECT mp.planet_id, mp.commodity_id, mp.buy_price, mp.sell_price, mp.stock, mp.demand, mp.last_update
		FROM market_prices mp
		INNER JOIN planets p ON mp.planet_id = p.id
		WHERE p.system_id = $1
		ORDER BY mp.commodity_id, mp.planet_id
	`

	rows, err := r.db.QueryContext(ctx, query, systemID)
	if err != nil {
		errors.RecordGlobalError("market_repository", "query_system_prices", err)
		log.Error("Failed to query market prices for system: system_id=%s, error=%v", systemID, err)
		return nil, fmt.Errorf("failed to query market prices: %w", err)
	}
	defer rows.Close()

	var prices []models.MarketPrice
	for rows.Next() {
		var price models.MarketPrice
		err := rows.Scan(
			&price.PlanetID,
			&price.CommodityID,
			&price.BuyPrice,
			&price.SellPrice,
			&price.Stock,
			&price.Demand,
			&price.LastUpdate,
		)
		if err != nil {
			log.Error("Failed to scan market price row: system_id=%s, error=%v", systemID, err)
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}
		prices = append(prices, price)
	}

	if err := rows.Err(); err != nil {
		log.Error("Error iterating market prices: system_id=%s, error=%v", systemID, err)
		return nil, fmt.Errorf("error iterating market prices: %w", err)
	}

	log.Debug("Retrieved %d market prices for system: system_id=%s", len(prices), systemID)
	return prices, nil
}

// UpdateStock updates the stock of a commodity at a planet
func (r *MarketRepository) UpdateStock(ctx context.Context, planetID uuid.UUID, commodityID string, delta int) error {
	query := `
		UPDATE market_prices
		SET stock = stock + $1, last_update = $2
		WHERE planet_id = $3 AND commodity_id = $4
	`

	result, err := r.db.ExecContext(ctx, query, delta, sql.NullInt64{Int64: 0, Valid: true}, planetID, commodityID)
	if err != nil {
		errors.RecordGlobalError("market_repository", "update_stock", err)
		log.Error("Failed to update stock: planet_id=%s, commodity_id=%s, delta=%d, error=%v", planetID, commodityID, delta, err)
		return fmt.Errorf("failed to update stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("Failed to get rows affected: planet_id=%s, commodity_id=%s, error=%v", planetID, commodityID, err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Debug("Market price not found for stock update: planet_id=%s, commodity_id=%s", planetID, commodityID)
		return ErrMarketPriceNotFound
	}

	log.Debug("Updated stock: planet_id=%s, commodity_id=%s, delta=%d", planetID, commodityID, delta)
	return nil
}

// ErrMarketPriceNotFound is returned when a market price is not found
var ErrMarketPriceNotFound = fmt.Errorf("market price not found")
