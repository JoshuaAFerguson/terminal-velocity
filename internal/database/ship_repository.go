// File: internal/database/ship_repository.go
// Project: Terminal Velocity
// Description: Repository for ship management including cargo, weapons, outfits,
//              and combat damage tracking
// Version: 1.2.0
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

// ShipRepository handles all database operations for ships.
//
// Manages:
//   - Ship creation, retrieval, updates, deletion
//   - Cargo management (add/remove commodities)
//   - Weapons and ammunition tracking
//   - Outfits (equipment) management
//   - Combat damage (hull/shields)
//   - Fuel levels
//
// Data model:
//   - Ships stored in 'ships' table
//   - Cargo in 'ship_cargo' table (many-to-many with commodities)
//   - Weapons in 'ship_weapons' table with ammo counts
//   - Outfits in 'ship_outfits' table
//
// Thread-safety:
//   - All methods are thread-safe
//   - Cargo operations check for sufficient quantity before removal
type ShipRepository struct {
	db *DB // Database connection pool
}

// NewShipRepository creates a new ship repository
func NewShipRepository(db *DB) *ShipRepository {
	return &ShipRepository{db: db}
}

// GetByID retrieves a ship by ID
func (r *ShipRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Ship, error) {
	query := `
		SELECT id, owner_id, type_id, name, hull, shields, fuel, crew
		FROM ships
		WHERE id = $1
	`

	var ship models.Ship
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ship.ID,
		&ship.OwnerID,
		&ship.TypeID,
		&ship.Name,
		&ship.Hull,
		&ship.Shields,
		&ship.Fuel,
		&ship.Crew,
	)

	if err == sql.ErrNoRows {
		return nil, ErrShipNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query ship: %w", err)
	}

	// Load cargo
	ship.Cargo, err = r.loadCargo(ctx, ship.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cargo: %w", err)
	}

	// Load weapons
	ship.Weapons, err = r.loadWeapons(ctx, ship.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapons: %w", err)
	}

	// Load weapon ammo
	ship.WeaponAmmo, err = r.loadWeaponAmmo(ctx, ship.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapon ammo: %w", err)
	}

	// Load outfits
	ship.Outfits, err = r.loadOutfits(ctx, ship.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load outfits: %w", err)
	}

	return &ship, nil
}

// GetByOwner retrieves all ships owned by a player
func (r *ShipRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*models.Ship, error) {
	query := `
		SELECT id, owner_id, type_id, name, hull, shields, fuel, crew
		FROM ships
		WHERE owner_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ships: %w", err)
	}
	defer rows.Close()

	var ships []*models.Ship
	for rows.Next() {
		var ship models.Ship
		err := rows.Scan(
			&ship.ID,
			&ship.OwnerID,
			&ship.TypeID,
			&ship.Name,
			&ship.Hull,
			&ship.Shields,
			&ship.Fuel,
			&ship.Crew,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ship: %w", err)
		}

		// Load cargo
		ship.Cargo, err = r.loadCargo(ctx, ship.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load cargo: %w", err)
		}

		// Load weapons
		ship.Weapons, err = r.loadWeapons(ctx, ship.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load weapons: %w", err)
		}

		// Load outfits
		ship.Outfits, err = r.loadOutfits(ctx, ship.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load outfits: %w", err)
		}

		ships = append(ships, &ship)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ships: %w", err)
	}

	return ships, nil
}

// Create creates a new ship
func (r *ShipRepository) Create(ctx context.Context, ship *models.Ship) error {
	query := `
		INSERT INTO ships (id, owner_id, type_id, name, hull, shields, fuel, crew)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		ship.ID,
		ship.OwnerID,
		ship.TypeID,
		ship.Name,
		ship.Hull,
		ship.Shields,
		ship.Fuel,
		ship.Crew,
	)

	if err != nil {
		return fmt.Errorf("failed to create ship: %w", err)
	}

	return nil
}

// Update updates a ship's basic properties
func (r *ShipRepository) Update(ctx context.Context, ship *models.Ship) error {
	query := `
		UPDATE ships
		SET type_id = $1, name = $2, hull = $3, shields = $4, fuel = $5, crew = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		ship.TypeID,
		ship.Name,
		ship.Hull,
		ship.Shields,
		ship.Fuel,
		ship.Crew,
		ship.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrShipNotFound
	}

	return nil
}

// UpdateFuel updates a ship's fuel level
func (r *ShipRepository) UpdateFuel(ctx context.Context, shipID uuid.UUID, fuel int) error {
	query := `
		UPDATE ships
		SET fuel = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, fuel, shipID)
	if err != nil {
		return fmt.Errorf("failed to update fuel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrShipNotFound
	}

	return nil
}

// UpdateWeaponAmmo updates the ammo count for a specific weapon slot
func (r *ShipRepository) UpdateWeaponAmmo(ctx context.Context, shipID uuid.UUID, slotIndex, ammo int) error {
	query := `
		UPDATE ship_weapons
		SET current_ammo = $1
		WHERE ship_id = $2 AND slot_index = $3
	`

	result, err := r.db.ExecContext(ctx, query, ammo, shipID, slotIndex)
	if err != nil {
		return fmt.Errorf("failed to update weapon ammo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("weapon slot not found")
	}

	return nil
}

// UpdateAllWeaponAmmo updates ammo for all weapons on a ship
func (r *ShipRepository) UpdateAllWeaponAmmo(ctx context.Context, shipID uuid.UUID, weaponAmmo map[int]int) error {
	for slotIndex, ammo := range weaponAmmo {
		if err := r.UpdateWeaponAmmo(ctx, shipID, slotIndex, ammo); err != nil {
			return err
		}
	}
	return nil
}

// UpdateHullAndShields updates a ship's hull and shields (combat damage)
func (r *ShipRepository) UpdateHullAndShields(ctx context.Context, shipID uuid.UUID, hull, shields int) error {
	query := `
		UPDATE ships
		SET hull = $1, shields = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, hull, shields, shipID)
	if err != nil {
		return fmt.Errorf("failed to update hull/shields: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrShipNotFound
	}

	return nil
}

// Delete deletes a ship
func (r *ShipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ships WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrShipNotFound
	}

	return nil
}

// AddCargo adds cargo to a ship using UPSERT for quantity accumulation.
//
// If the commodity already exists in cargo, quantities are added together.
// Uses ON CONFLICT to handle concurrent additions safely.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - shipID: Ship UUID
//   - commodityID: Commodity identifier (e.g., "food", "minerals")
//   - quantity: Amount to add (must be positive)
//
// Returns:
//   - error: Database error
//
// Note: Does NOT check cargo capacity - caller must verify before calling.
func (r *ShipRepository) AddCargo(ctx context.Context, shipID uuid.UUID, commodityID string, quantity int) error {
	query := `
		INSERT INTO ship_cargo (ship_id, commodity_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (ship_id, commodity_id)
		DO UPDATE SET quantity = ship_cargo.quantity + $3
	`

	_, err := r.db.ExecContext(ctx, query, shipID, commodityID, quantity)
	if err != nil {
		return fmt.Errorf("failed to add cargo: %w", err)
	}

	return nil
}

// RemoveCargo removes cargo from a ship with quantity validation.
//
// Checks that sufficient cargo exists before removal to prevent negative quantities.
// Deletes the row entirely if quantity reaches zero.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - shipID: Ship UUID
//   - commodityID: Commodity identifier
//   - quantity: Amount to remove (must be positive)
//
// Returns:
//   - error: "insufficient cargo" if not enough quantity, "commodity not in cargo" if not found
//
// Thread-safety:
//   - Safe for concurrent cargo operations
func (r *ShipRepository) RemoveCargo(ctx context.Context, shipID uuid.UUID, commodityID string, quantity int) error {
	// First check if we have enough
	var current int
	checkQuery := `SELECT quantity FROM ship_cargo WHERE ship_id = $1 AND commodity_id = $2`
	err := r.db.QueryRowContext(ctx, checkQuery, shipID, commodityID).Scan(&current)
	if err == sql.ErrNoRows {
		return fmt.Errorf("commodity not in cargo")
	}
	if err != nil {
		return fmt.Errorf("failed to check cargo: %w", err)
	}

	if current < quantity {
		return fmt.Errorf("insufficient cargo (have %d, need %d)", current, quantity)
	}

	// Remove cargo
	if current == quantity {
		// Delete the row entirely
		query := `DELETE FROM ship_cargo WHERE ship_id = $1 AND commodity_id = $2`
		_, err = r.db.ExecContext(ctx, query, shipID, commodityID)
	} else {
		// Decrease quantity
		query := `UPDATE ship_cargo SET quantity = quantity - $3 WHERE ship_id = $1 AND commodity_id = $2`
		_, err = r.db.ExecContext(ctx, query, shipID, commodityID, quantity)
	}

	if err != nil {
		return fmt.Errorf("failed to remove cargo: %w", err)
	}

	return nil
}

// loadCargo loads cargo for a ship
func (r *ShipRepository) loadCargo(ctx context.Context, shipID uuid.UUID) ([]models.CargoItem, error) {
	query := `
		SELECT commodity_id, quantity
		FROM ship_cargo
		WHERE ship_id = $1
		ORDER BY commodity_id
	`

	rows, err := r.db.QueryContext(ctx, query, shipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cargo []models.CargoItem
	for rows.Next() {
		var item models.CargoItem
		if err := rows.Scan(&item.CommodityID, &item.Quantity); err != nil {
			return nil, err
		}
		cargo = append(cargo, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cargo, nil
}

// loadWeapons loads weapons for a ship
func (r *ShipRepository) loadWeapons(ctx context.Context, shipID uuid.UUID) ([]string, error) {
	query := `
		SELECT weapon_id
		FROM ship_weapons
		WHERE ship_id = $1
		ORDER BY slot_index
	`

	rows, err := r.db.QueryContext(ctx, query, shipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weapons []string
	for rows.Next() {
		var weaponID string
		if err := rows.Scan(&weaponID); err != nil {
			return nil, err
		}
		weapons = append(weapons, weaponID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return weapons, nil
}

// loadWeaponAmmo loads ammo counts for a ship's weapons
func (r *ShipRepository) loadWeaponAmmo(ctx context.Context, shipID uuid.UUID) (map[int]int, error) {
	query := `
		SELECT slot_index, current_ammo
		FROM ship_weapons
		WHERE ship_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, shipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ammo := make(map[int]int)
	for rows.Next() {
		var slotIndex, currentAmmo int
		if err := rows.Scan(&slotIndex, &currentAmmo); err != nil {
			return nil, err
		}
		ammo[slotIndex] = currentAmmo
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ammo, nil
}

// loadOutfits loads outfits for a ship
func (r *ShipRepository) loadOutfits(ctx context.Context, shipID uuid.UUID) ([]string, error) {
	query := `
		SELECT outfit_id
		FROM ship_outfits
		WHERE ship_id = $1
		ORDER BY outfit_id
	`

	rows, err := r.db.QueryContext(ctx, query, shipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outfits []string
	for rows.Next() {
		var outfitID string
		if err := rows.Scan(&outfitID); err != nil {
			return nil, err
		}
		outfits = append(outfits, outfitID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return outfits, nil
}

// ErrShipNotFound is returned when a ship is not found
var ErrShipNotFound = fmt.Errorf("ship not found")
