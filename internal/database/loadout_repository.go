// File: internal/database/loadout_repository.go
// Project: Terminal Velocity
// Description: Database repository for shared loadouts
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// LoadoutRepository handles loadout data access
type LoadoutRepository struct {
	db *DB
}

// NewLoadoutRepository creates a new loadout repository
func NewLoadoutRepository(db *DB) *LoadoutRepository {
	return &LoadoutRepository{db: db}
}

// CreateLoadout inserts a new shared loadout
func (r *LoadoutRepository) CreateLoadout(ctx context.Context, loadout *models.SharedLoadout) error {
	weaponsJSON, err := json.Marshal(loadout.Weapons)
	if err != nil {
		return fmt.Errorf("failed to marshal weapons: %w", err)
	}

	outfitsJSON, err := json.Marshal(loadout.Outfits)
	if err != nil {
		return fmt.Errorf("failed to marshal outfits: %w", err)
	}

	statsJSON, err := json.Marshal(loadout.Stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	query := `
		INSERT INTO shared_loadouts (
			id, player_id, ship_type_id, name, description,
			weapons, outfits, stats, is_public,
			created_at, updated_at, views, favorites
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.ExecContext(ctx, query,
		loadout.ID,
		loadout.PlayerID,
		loadout.ShipTypeID,
		loadout.Name,
		loadout.Description,
		weaponsJSON,
		outfitsJSON,
		statsJSON,
		loadout.IsPublic,
		loadout.CreatedAt,
		loadout.UpdatedAt,
		loadout.Views,
		loadout.Favorites,
	)

	if err != nil {
		return fmt.Errorf("failed to create loadout: %w", err)
	}

	return nil
}

// GetLoadout retrieves a loadout by ID
func (r *LoadoutRepository) GetLoadout(ctx context.Context, loadoutID uuid.UUID) (*models.SharedLoadout, error) {
	query := `
		SELECT id, player_id, ship_type_id, name, description,
		       weapons, outfits, stats, is_public,
		       created_at, updated_at, views, favorites
		FROM shared_loadouts
		WHERE id = $1
	`

	var loadout models.SharedLoadout
	var weaponsJSON, outfitsJSON, statsJSON []byte

	err := r.db.QueryRowContext(ctx, query, loadoutID).Scan(
		&loadout.ID,
		&loadout.PlayerID,
		&loadout.ShipTypeID,
		&loadout.Name,
		&loadout.Description,
		&weaponsJSON,
		&outfitsJSON,
		&statsJSON,
		&loadout.IsPublic,
		&loadout.CreatedAt,
		&loadout.UpdatedAt,
		&loadout.Views,
		&loadout.Favorites,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("loadout not found")
		}
		return nil, fmt.Errorf("failed to query loadout: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(weaponsJSON, &loadout.Weapons); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weapons: %w", err)
	}
	if err := json.Unmarshal(outfitsJSON, &loadout.Outfits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outfits: %w", err)
	}
	if err := json.Unmarshal(statsJSON, &loadout.Stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return &loadout, nil
}

// GetPlayerLoadouts retrieves all loadouts created by a player
func (r *LoadoutRepository) GetPlayerLoadouts(ctx context.Context, playerID uuid.UUID) ([]*models.SharedLoadout, error) {
	query := `
		SELECT id, player_id, ship_type_id, name, description,
		       weapons, outfits, stats, is_public,
		       created_at, updated_at, views, favorites
		FROM shared_loadouts
		WHERE player_id = $1
		ORDER BY updated_at DESC
	`

	return r.queryLoadouts(ctx, query, playerID)
}

// GetPublicLoadouts retrieves public loadouts, optionally filtered by ship type
func (r *LoadoutRepository) GetPublicLoadouts(ctx context.Context, shipTypeID string, limit, offset int) ([]*models.SharedLoadout, error) {
	var query string
	var args []interface{}

	if shipTypeID != "" {
		query = `
			SELECT id, player_id, ship_type_id, name, description,
			       weapons, outfits, stats, is_public,
			       created_at, updated_at, views, favorites
			FROM shared_loadouts
			WHERE is_public = true AND ship_type_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{shipTypeID, limit, offset}
	} else {
		query = `
			SELECT id, player_id, ship_type_id, name, description,
			       weapons, outfits, stats, is_public,
			       created_at, updated_at, views, favorites
			FROM shared_loadouts
			WHERE is_public = true
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.queryLoadouts(ctx, query, args...)
}

// GetPopularLoadouts retrieves most viewed/favorited loadouts
func (r *LoadoutRepository) GetPopularLoadouts(ctx context.Context, limit int) ([]*models.SharedLoadout, error) {
	query := `
		SELECT id, player_id, ship_type_id, name, description,
		       weapons, outfits, stats, is_public,
		       created_at, updated_at, views, favorites
		FROM shared_loadouts
		WHERE is_public = true
		ORDER BY (favorites * 2 + views) DESC
		LIMIT $1
	`

	return r.queryLoadouts(ctx, query, limit)
}

// UpdateLoadout updates a loadout
func (r *LoadoutRepository) UpdateLoadout(ctx context.Context, loadout *models.SharedLoadout) error {
	weaponsJSON, err := json.Marshal(loadout.Weapons)
	if err != nil {
		return fmt.Errorf("failed to marshal weapons: %w", err)
	}

	outfitsJSON, err := json.Marshal(loadout.Outfits)
	if err != nil {
		return fmt.Errorf("failed to marshal outfits: %w", err)
	}

	statsJSON, err := json.Marshal(loadout.Stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	query := `
		UPDATE shared_loadouts
		SET name = $1, description = $2, weapons = $3, outfits = $4,
		    stats = $5, is_public = $6, updated_at = $7
		WHERE id = $8
	`

	_, err = r.db.ExecContext(ctx, query,
		loadout.Name,
		loadout.Description,
		weaponsJSON,
		outfitsJSON,
		statsJSON,
		loadout.IsPublic,
		loadout.UpdatedAt,
		loadout.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update loadout: %w", err)
	}

	return nil
}

// DeleteLoadout deletes a loadout
func (r *LoadoutRepository) DeleteLoadout(ctx context.Context, loadoutID uuid.UUID) error {
	query := `DELETE FROM shared_loadouts WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, loadoutID)
	if err != nil {
		return fmt.Errorf("failed to delete loadout: %w", err)
	}

	return nil
}

// IncrementViews increments the view count for a loadout
func (r *LoadoutRepository) IncrementViews(ctx context.Context, loadoutID uuid.UUID) error {
	query := `UPDATE shared_loadouts SET views = views + 1 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, loadoutID)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	return nil
}

// AddFavorite adds a loadout to a player's favorites
func (r *LoadoutRepository) AddFavorite(ctx context.Context, loadoutID, playerID uuid.UUID) error {
	// Check if already favorited
	exists, err := r.IsFavorited(ctx, loadoutID, playerID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already favorited
	}

	query := `INSERT INTO loadout_favorites (loadout_id, player_id, created_at) VALUES ($1, $2, NOW())`

	_, err = r.db.ExecContext(ctx, query, loadoutID, playerID)
	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	// Increment favorite count
	query = `UPDATE shared_loadouts SET favorites = favorites + 1 WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, loadoutID)
	if err != nil {
		return fmt.Errorf("failed to increment favorites: %w", err)
	}

	return nil
}

// RemoveFavorite removes a loadout from a player's favorites
func (r *LoadoutRepository) RemoveFavorite(ctx context.Context, loadoutID, playerID uuid.UUID) error {
	query := `DELETE FROM loadout_favorites WHERE loadout_id = $1 AND player_id = $2`

	result, err := r.db.ExecContext(ctx, query, loadoutID, playerID)
	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		// Decrement favorite count
		query = `UPDATE shared_loadouts SET favorites = favorites - 1 WHERE id = $1 AND favorites > 0`
		_, err = r.db.ExecContext(ctx, query, loadoutID)
		if err != nil {
			return fmt.Errorf("failed to decrement favorites: %w", err)
		}
	}

	return nil
}

// IsFavorited checks if a player has favorited a loadout
func (r *LoadoutRepository) IsFavorited(ctx context.Context, loadoutID, playerID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM loadout_favorites WHERE loadout_id = $1 AND player_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, loadoutID, playerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite: %w", err)
	}

	return exists, nil
}

// GetFavorites retrieves a player's favorited loadouts
func (r *LoadoutRepository) GetFavorites(ctx context.Context, playerID uuid.UUID) ([]*models.SharedLoadout, error) {
	query := `
		SELECT l.id, l.player_id, l.ship_type_id, l.name, l.description,
		       l.weapons, l.outfits, l.stats, l.is_public,
		       l.created_at, l.updated_at, l.views, l.favorites
		FROM shared_loadouts l
		INNER JOIN loadout_favorites f ON l.id = f.loadout_id
		WHERE f.player_id = $1
		ORDER BY f.created_at DESC
	`

	return r.queryLoadouts(ctx, query, playerID)
}

// queryLoadouts is a helper function to query multiple loadouts
func (r *LoadoutRepository) queryLoadouts(ctx context.Context, query string, args ...interface{}) ([]*models.SharedLoadout, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query loadouts: %w", err)
	}
	defer rows.Close()

	var loadouts []*models.SharedLoadout
	for rows.Next() {
		var loadout models.SharedLoadout
		var weaponsJSON, outfitsJSON, statsJSON []byte

		err := rows.Scan(
			&loadout.ID,
			&loadout.PlayerID,
			&loadout.ShipTypeID,
			&loadout.Name,
			&loadout.Description,
			&weaponsJSON,
			&outfitsJSON,
			&statsJSON,
			&loadout.IsPublic,
			&loadout.CreatedAt,
			&loadout.UpdatedAt,
			&loadout.Views,
			&loadout.Favorites,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan loadout: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(weaponsJSON, &loadout.Weapons); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weapons: %w", err)
		}
		if err := json.Unmarshal(outfitsJSON, &loadout.Outfits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal outfits: %w", err)
		}
		if err := json.Unmarshal(statsJSON, &loadout.Stats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
		}

		loadouts = append(loadouts, &loadout)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating loadouts: %w", err)
	}

	return loadouts, nil
}
