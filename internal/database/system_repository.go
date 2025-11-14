// File: internal/database/system_repository.go
// Project: Terminal Velocity
// Description: Database repository for system_repository
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var (
	ErrSystemNotFound = errors.New("system not found")
	ErrPlanetNotFound = errors.New("planet not found")
)

// SystemRepository handles star system data access
type SystemRepository struct {
	db *DB
}

// NewSystemRepository creates a new system repository
func NewSystemRepository(db *DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// CreateSystem inserts a new star system
func (r *SystemRepository) CreateSystem(ctx context.Context, system *models.StarSystem) error {
	query := `
		INSERT INTO star_systems (id, name, pos_x, pos_y, government_id, tech_level, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		system.ID,
		system.Name,
		system.Position.X,
		system.Position.Y,
		system.GovernmentID,
		system.TechLevel,
		system.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	return nil
}

// GetSystemByID retrieves a system by ID
func (r *SystemRepository) GetSystemByID(ctx context.Context, id uuid.UUID) (*models.StarSystem, error) {
	query := `
		SELECT id, name, pos_x, pos_y, government_id, tech_level, description
		FROM star_systems
		WHERE id = $1
	`

	var system models.StarSystem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&system.ID,
		&system.Name,
		&system.Position.X,
		&system.Position.Y,
		&system.GovernmentID,
		&system.TechLevel,
		&system.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSystemNotFound
		}
		return nil, fmt.Errorf("failed to query system: %w", err)
	}

	// Load jump connections
	system.ConnectedSystems, err = r.getJumpConnections(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load jump connections: %w", err)
	}

	return &system, nil
}

// GetSystemByName retrieves a system by name
func (r *SystemRepository) GetSystemByName(ctx context.Context, name string) (*models.StarSystem, error) {
	query := `
		SELECT id, name, pos_x, pos_y, government_id, tech_level, description
		FROM star_systems
		WHERE name = $1
	`

	var system models.StarSystem
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&system.ID,
		&system.Name,
		&system.Position.X,
		&system.Position.Y,
		&system.GovernmentID,
		&system.TechLevel,
		&system.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSystemNotFound
		}
		return nil, fmt.Errorf("failed to query system: %w", err)
	}

	// Load jump connections
	system.ConnectedSystems, err = r.getJumpConnections(ctx, system.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load jump connections: %w", err)
	}

	return &system, nil
}

// ListSystems returns all systems
func (r *SystemRepository) ListSystems(ctx context.Context) ([]*models.StarSystem, error) {
	query := `
		SELECT id, name, pos_x, pos_y, government_id, tech_level, description
		FROM star_systems
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query systems: %w", err)
	}
	defer rows.Close()

	var systems []*models.StarSystem
	for rows.Next() {
		var system models.StarSystem
		err := rows.Scan(
			&system.ID,
			&system.Name,
			&system.Position.X,
			&system.Position.Y,
			&system.GovernmentID,
			&system.TechLevel,
			&system.Description,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}

		// Load jump connections
		system.ConnectedSystems, err = r.getJumpConnections(ctx, system.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load jump connections: %w", err)
		}

		systems = append(systems, &system)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating systems: %w", err)
	}

	return systems, nil
}

// GetSystemsByGovernment returns all systems controlled by a government
func (r *SystemRepository) GetSystemsByGovernment(ctx context.Context, governmentID string) ([]*models.StarSystem, error) {
	query := `
		SELECT id, name, pos_x, pos_y, government_id, tech_level, description
		FROM star_systems
		WHERE government_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, governmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query systems: %w", err)
	}
	defer rows.Close()

	var systems []*models.StarSystem
	for rows.Next() {
		var system models.StarSystem
		err := rows.Scan(
			&system.ID,
			&system.Name,
			&system.Position.X,
			&system.Position.Y,
			&system.GovernmentID,
			&system.TechLevel,
			&system.Description,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}

		connectedSystems, err := r.getJumpConnections(ctx, system.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get jump connections for system %s: %w", system.ID, err)
		}
		system.ConnectedSystems = connectedSystems
		systems = append(systems, &system)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating systems: %w", err)
	}

	return systems, nil
}

// CreateJumpRoute creates a bidirectional jump connection between two systems
func (r *SystemRepository) CreateJumpRoute(ctx context.Context, systemA, systemB uuid.UUID) error {
	query := `
		INSERT INTO system_connections (system_a, system_b)
		VALUES ($1, $2), ($2, $1)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, systemA, systemB)
	if err != nil {
		return fmt.Errorf("failed to create jump route: %w", err)
	}

	return nil
}

// GetConnections returns all jump connections for a system (public API method)
func (r *SystemRepository) GetConnections(ctx context.Context, systemID uuid.UUID) ([]uuid.UUID, error) {
	return r.getJumpConnections(ctx, systemID)
}

// getJumpConnections returns all jump connections for a system
func (r *SystemRepository) getJumpConnections(ctx context.Context, systemID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT system_b FROM system_connections WHERE system_a = $1
	`

	rows, err := r.db.QueryContext(ctx, query, systemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		connections = append(connections, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}

// CreatePlanet inserts a new planet
func (r *SystemRepository) CreatePlanet(ctx context.Context, planet *models.Planet) error {
	query := `
		INSERT INTO planets (id, system_id, name, description, population, tech_level, services)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		planet.ID,
		planet.SystemID,
		planet.Name,
		planet.Description,
		planet.Population,
		planet.TechLevel,
		planet.Services,
	)

	if err != nil {
		return fmt.Errorf("failed to create planet: %w", err)
	}

	return nil
}

// GetPlanetByID retrieves a planet by ID
func (r *SystemRepository) GetPlanetByID(ctx context.Context, id uuid.UUID) (*models.Planet, error) {
	query := `
		SELECT id, system_id, name, description, population, tech_level, services
		FROM planets
		WHERE id = $1
	`

	var planet models.Planet
	var services []byte // PostgreSQL array

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&planet.ID,
		&planet.SystemID,
		&planet.Name,
		&planet.Description,
		&planet.Population,
		&planet.TechLevel,
		&services,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPlanetNotFound
		}
		return nil, fmt.Errorf("failed to query planet: %w", err)
	}

	// Parse services array (PostgreSQL text array format)
	// For now, we'll leave it as is - services can be handled later

	return &planet, nil
}

// GetPlanetsBySystem returns all planets in a system
func (r *SystemRepository) GetPlanetsBySystem(ctx context.Context, systemID uuid.UUID) ([]*models.Planet, error) {
	query := `
		SELECT id, system_id, name, description, population, tech_level, services
		FROM planets
		WHERE system_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query planets: %w", err)
	}
	defer rows.Close()

	var planets []*models.Planet
	for rows.Next() {
		var planet models.Planet
		var services []byte

		err := rows.Scan(
			&planet.ID,
			&planet.SystemID,
			&planet.Name,
			&planet.Description,
			&planet.Population,
			&planet.TechLevel,
			&services,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan planet: %w", err)
		}

		planets = append(planets, &planet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating planets: %w", err)
	}

	return planets, nil
}

// BulkCreateSystems creates multiple systems in a single transaction
func (r *SystemRepository) BulkCreateSystems(ctx context.Context, systems []*models.StarSystem) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO star_systems (id, name, pos_x, pos_y, government_id, tech_level, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, system := range systems {
			_, err := stmt.ExecContext(ctx,
				system.ID,
				system.Name,
				system.Position.X,
				system.Position.Y,
				system.GovernmentID,
				system.TechLevel,
				system.Description,
			)
			if err != nil {
				return fmt.Errorf("failed to insert system %s: %w", system.Name, err)
			}
		}

		return nil
	})
}

// BulkCreateJumpRoutes creates multiple jump routes in a single transaction
func (r *SystemRepository) BulkCreateJumpRoutes(ctx context.Context, routes [][2]uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO system_connections (system_a, system_b)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, route := range routes {
			// Create bidirectional connection
			if _, err := stmt.ExecContext(ctx, route[0], route[1]); err != nil {
				return fmt.Errorf("failed to insert jump route: %w", err)
			}
			if _, err := stmt.ExecContext(ctx, route[1], route[0]); err != nil {
				return fmt.Errorf("failed to insert jump route: %w", err)
			}
		}

		return nil
	})
}

// BulkCreatePlanets creates multiple planets in a single transaction
func (r *SystemRepository) BulkCreatePlanets(ctx context.Context, planets []*models.Planet) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO planets (id, system_id, name, description, population, tech_level, services)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, planet := range planets {
			_, err := stmt.ExecContext(ctx,
				planet.ID,
				planet.SystemID,
				planet.Name,
				planet.Description,
				planet.Population,
				planet.TechLevel,
				planet.Services,
			)
			if err != nil {
				return fmt.Errorf("failed to insert planet %s: %w", planet.Name, err)
			}
		}

		return nil
	})
}

// CountSystems returns the total number of systems
func (r *SystemRepository) CountSystems(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM star_systems`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count systems: %w", err)
	}
	return count, nil
}
