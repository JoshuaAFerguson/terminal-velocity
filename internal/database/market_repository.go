// File: internal/database/market_repository.go
// Project: Terminal Velocity
// Description: Database repository for market_repository
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// MarketRepository handles market price data persistence
type MarketRepository struct {
	db *DB
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
		return nil, ErrMarketPriceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query market price: %w", err)
	}

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
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating market prices: %w", err)
	}

	return prices, nil
}

// UpsertMarketPrice inserts or updates a market price
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
		return fmt.Errorf("failed to upsert market price: %w", err)
	}

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
		return fmt.Errorf("failed to update market price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMarketPriceNotFound
	}

	return nil
}

// DeleteMarketPrice deletes a market price
func (r *MarketRepository) DeleteMarketPrice(ctx context.Context, planetID uuid.UUID, commodityID string) error {
	query := `DELETE FROM market_prices WHERE planet_id = $1 AND commodity_id = $2`

	result, err := r.db.ExecContext(ctx, query, planetID, commodityID)
	if err != nil {
		return fmt.Errorf("failed to delete market price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMarketPriceNotFound
	}

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
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stale markets: %w", err)
	}

	return prices, nil
}

// ErrMarketPriceNotFound is returned when a market price is not found
var ErrMarketPriceNotFound = fmt.Errorf("market price not found")
