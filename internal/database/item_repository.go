// File: internal/database/item_repository.go
// Project: Terminal Velocity
// Description: Repository for player inventory items with CRUD operations,
//              location tracking, and transfer audit logging
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// ItemRepository handles all database operations for player items.
//
// Manages player inventory and item ownership:
//   - Item creation, retrieval, updates, deletion
//   - Location tracking (ship, station, equipped)
//   - Item properties (JSONB for flexible data)
//   - Item transfers between players with audit logging
//   - Batch operations for efficiency
//
// Data model:
//   - Items have type, equipment_id, location, and properties
//   - Properties stored as JSONB for flexibility
//   - Transfer history logged in item_transfers table
//
// Thread-safety:
//   - All methods are thread-safe
//   - Transfers use transactions for atomicity
type ItemRepository struct {
	db *DB // Database connection pool
}

// NewItemRepository creates a new ItemRepository
func NewItemRepository(db *DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// GetPlayerItems returns all items owned by a player
func (r *ItemRepository) GetPlayerItems(ctx context.Context, playerID uuid.UUID) ([]*models.PlayerItem, error) {
	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE player_id = $1
		ORDER BY acquired_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query player items: %w", err)
	}
	defer rows.Close()

	var items []*models.PlayerItem
	for rows.Next() {
		item, err := scanPlayerItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player items: %w", err)
	}

	return items, nil
}

// GetAvailableItems returns items available for use (ship or station storage)
func (r *ItemRepository) GetAvailableItems(ctx context.Context, playerID uuid.UUID) ([]*models.PlayerItem, error) {
	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE player_id = $1 AND location IN ('ship', 'station_storage')
		ORDER BY item_type, equipment_id
	`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available items: %w", err)
	}
	defer rows.Close()

	var items []*models.PlayerItem
	for rows.Next() {
		item, err := scanPlayerItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating available items: %w", err)
	}

	return items, nil
}

// GetItemsByLocation returns items at a specific location
func (r *ItemRepository) GetItemsByLocation(ctx context.Context, playerID uuid.UUID, location models.ItemLocation, locationID uuid.UUID) ([]*models.PlayerItem, error) {
	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE player_id = $1 AND location = $2 AND location_id = $3
		ORDER BY item_type, equipment_id
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, location, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items by location: %w", err)
	}
	defer rows.Close()

	var items []*models.PlayerItem
	for rows.Next() {
		item, err := scanPlayerItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items by location: %w", err)
	}

	return items, nil
}

// GetItemsByType returns all items of a specific type for a player
func (r *ItemRepository) GetItemsByType(ctx context.Context, playerID uuid.UUID, itemType models.ItemType) ([]*models.PlayerItem, error) {
	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE player_id = $1 AND item_type = $2
		ORDER BY equipment_id
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, itemType)
	if err != nil {
		return nil, fmt.Errorf("failed to query items by type: %w", err)
	}
	defer rows.Close()

	var items []*models.PlayerItem
	for rows.Next() {
		item, err := scanPlayerItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items by type: %w", err)
	}

	return items, nil
}

// GetItemByID returns a single item by ID
func (r *ItemRepository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*models.PlayerItem, error) {
	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, itemID)
	item, err := scanPlayerItemRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found: %s", itemID)
		}
		return nil, err
	}

	return item, nil
}

// GetItemsByIDs returns multiple items by their IDs (for batch operations)
func (r *ItemRepository) GetItemsByIDs(ctx context.Context, itemIDs []uuid.UUID) ([]*models.PlayerItem, error) {
	if len(itemIDs) == 0 {
		return []*models.PlayerItem{}, nil
	}

	query := `
		SELECT id, player_id, item_type, equipment_id, location, location_id,
		       properties, acquired_at, created_at, updated_at
		FROM player_items
		WHERE id = ANY($1)
		ORDER BY item_type, equipment_id
	`

	rows, err := r.db.QueryContext(ctx, query, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query items by IDs: %w", err)
	}
	defer rows.Close()

	var items []*models.PlayerItem
	for rows.Next() {
		item, err := scanPlayerItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items by IDs: %w", err)
	}

	return items, nil
}

// CreateItem creates a new player item
func (r *ItemRepository) CreateItem(ctx context.Context, item *models.PlayerItem) error {
	// Validate before creating
	if err := item.Validate(); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	query := `
		INSERT INTO player_items (player_id, item_type, equipment_id, location, location_id, properties)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, acquired_at, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.PlayerID, item.ItemType, item.EquipmentID,
		item.Location, item.LocationID, item.Properties,
	).Scan(&item.ID, &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// CreateItems creates multiple items in a single transaction (batch creation)
func (r *ItemRepository) CreateItems(ctx context.Context, items []*models.PlayerItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO player_items (player_id, item_type, equipment_id, location, location_id, properties)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, acquired_at, created_at, updated_at
	`

	for _, item := range items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("invalid item in batch: %w", err)
		}

		err := tx.QueryRowContext(ctx, query,
			item.PlayerID, item.ItemType, item.EquipmentID,
			item.Location, item.LocationID, item.Properties,
		).Scan(&item.ID, &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to create item in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch item creation: %w", err)
	}

	return nil
}

// UpdateItemLocation moves an item to a new location
func (r *ItemRepository) UpdateItemLocation(ctx context.Context, itemID uuid.UUID, location models.ItemLocation, locationID *uuid.UUID) error {
	if !location.Valid() {
		return fmt.Errorf("invalid location: %s", location)
	}

	query := `
		UPDATE player_items
		SET location = $1, location_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, location, locationID, itemID)
	if err != nil {
		return fmt.Errorf("failed to update item location: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found: %s", itemID)
	}

	return nil
}

// UpdateItemProperties updates the JSONB properties of an item
func (r *ItemRepository) UpdateItemProperties(ctx context.Context, itemID uuid.UUID, properties *models.ItemProperties) error {
	item := &models.PlayerItem{ID: itemID}
	if err := item.SetProperties(properties); err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	query := `
		UPDATE player_items
		SET properties = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, item.Properties, itemID)
	if err != nil {
		return fmt.Errorf("failed to update item properties: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found: %s", itemID)
	}

	return nil
}

// TransferItem transfers an item to another player (atomic with audit log)
func (r *ItemRepository) TransferItem(ctx context.Context, itemID uuid.UUID, toPlayerID uuid.UUID, transferType string, transferID *uuid.UUID) error {
	if !models.TransferTypeValid(transferType) {
		return fmt.Errorf("invalid transfer type: %s", transferType)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current owner
	var fromPlayerID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT player_id FROM player_items WHERE id = $1", itemID).Scan(&fromPlayerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("item not found: %s", itemID)
		}
		return fmt.Errorf("failed to get item owner: %w", err)
	}

	// Update ownership
	_, err = tx.ExecContext(ctx,
		"UPDATE player_items SET player_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		toPlayerID, itemID)
	if err != nil {
		return fmt.Errorf("failed to transfer item: %w", err)
	}

	// Create audit log entry
	_, err = tx.ExecContext(ctx,
		"INSERT INTO item_transfers (item_id, from_player_id, to_player_id, transfer_type, transfer_id) VALUES ($1, $2, $3, $4, $5)",
		itemID, fromPlayerID, toPlayerID, transferType, transferID,
	)
	if err != nil {
		return fmt.Errorf("failed to log transfer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransferItems transfers multiple items in a single transaction
func (r *ItemRepository) TransferItems(ctx context.Context, itemIDs []uuid.UUID, toPlayerID uuid.UUID, transferType string, transferID *uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	if !models.TransferTypeValid(transferType) {
		return fmt.Errorf("invalid transfer type: %s", transferType)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, itemID := range itemIDs {
		// Get current owner
		var fromPlayerID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT player_id FROM player_items WHERE id = $1", itemID).Scan(&fromPlayerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("item not found: %s", itemID)
			}
			return fmt.Errorf("failed to get item owner: %w", err)
		}

		// Update ownership
		_, err = tx.ExecContext(ctx,
			"UPDATE player_items SET player_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			toPlayerID, itemID)
		if err != nil {
			return fmt.Errorf("failed to transfer item: %w", err)
		}

		// Create audit log entry
		_, err = tx.ExecContext(ctx,
			"INSERT INTO item_transfers (item_id, from_player_id, to_player_id, transfer_type, transfer_id) VALUES ($1, $2, $3, $4, $5)",
			itemID, fromPlayerID, toPlayerID, transferType, transferID,
		)
		if err != nil {
			return fmt.Errorf("failed to log transfer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch transfer: %w", err)
	}

	return nil
}

// DeleteItem removes an item (e.g., consumed, destroyed)
func (r *ItemRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	query := "DELETE FROM player_items WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found: %s", itemID)
	}

	return nil
}

// DeleteItems removes multiple items in a single transaction
func (r *ItemRepository) DeleteItems(ctx context.Context, itemIDs []uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	query := "DELETE FROM player_items WHERE id = ANY($1)"
	result, err := r.db.ExecContext(ctx, query, itemIDs)
	if err != nil {
		return fmt.Errorf("failed to delete items: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != int64(len(itemIDs)) {
		return fmt.Errorf("expected to delete %d items, deleted %d", len(itemIDs), rowsAffected)
	}

	return nil
}

// GetItemTransferHistory returns transfer audit log for an item
func (r *ItemRepository) GetItemTransferHistory(ctx context.Context, itemID uuid.UUID) ([]*models.ItemTransfer, error) {
	query := `
		SELECT id, item_id, from_player_id, to_player_id, transfer_type, transfer_id, transferred_at
		FROM item_transfers
		WHERE item_id = $1
		ORDER BY transferred_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer history: %w", err)
	}
	defer rows.Close()

	var transfers []*models.ItemTransfer
	for rows.Next() {
		var transfer models.ItemTransfer
		err := rows.Scan(
			&transfer.ID, &transfer.ItemID, &transfer.FromPlayerID,
			&transfer.ToPlayerID, &transfer.TransferType, &transfer.TransferID,
			&transfer.TransferredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer: %w", err)
		}
		transfers = append(transfers, &transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfer history: %w", err)
	}

	return transfers, nil
}

// GetPlayerTransfers returns all transfers for a player (sent or received)
func (r *ItemRepository) GetPlayerTransfers(ctx context.Context, playerID uuid.UUID, limit int) ([]*models.ItemTransfer, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, item_id, from_player_id, to_player_id, transfer_type, transfer_id, transferred_at
		FROM item_transfers
		WHERE from_player_id = $1 OR to_player_id = $1
		ORDER BY transferred_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query player transfers: %w", err)
	}
	defer rows.Close()

	var transfers []*models.ItemTransfer
	for rows.Next() {
		var transfer models.ItemTransfer
		err := rows.Scan(
			&transfer.ID, &transfer.ItemID, &transfer.FromPlayerID,
			&transfer.ToPlayerID, &transfer.TransferType, &transfer.TransferID,
			&transfer.TransferredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer: %w", err)
		}
		transfers = append(transfers, &transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player transfers: %w", err)
	}

	return transfers, nil
}

// CountPlayerItems returns the number of items owned by a player
func (r *ItemRepository) CountPlayerItems(ctx context.Context, playerID uuid.UUID) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM player_items WHERE player_id = $1"
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count player items: %w", err)
	}
	return count, nil
}

// CountItemsByType returns the number of items of a specific type owned by a player
func (r *ItemRepository) CountItemsByType(ctx context.Context, playerID uuid.UUID, itemType models.ItemType) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM player_items WHERE player_id = $1 AND item_type = $2"
	err := r.db.QueryRowContext(ctx, query, playerID, itemType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items by type: %w", err)
	}
	return count, nil
}

// Helper functions

// scanPlayerItem scans a row into a PlayerItem struct
func scanPlayerItem(rows *sql.Rows) (*models.PlayerItem, error) {
	var item models.PlayerItem
	err := rows.Scan(
		&item.ID, &item.PlayerID, &item.ItemType, &item.EquipmentID,
		&item.Location, &item.LocationID, &item.Properties,
		&item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	return &item, nil
}

// scanPlayerItemRow scans a QueryRow result into a PlayerItem struct
func scanPlayerItemRow(row *sql.Row) (*models.PlayerItem, error) {
	var item models.PlayerItem
	err := row.Scan(
		&item.ID, &item.PlayerID, &item.ItemType, &item.EquipmentID,
		&item.Location, &item.LocationID, &item.Properties,
		&item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	return &item, nil
}
