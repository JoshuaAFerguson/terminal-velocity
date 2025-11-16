// File: internal/database/migrations.go
// Project: Terminal Velocity
// Description: Database schema migrations and version management
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RunMigrations executes SQL migration files from the given path.
//
// This method loads and executes the schema.sql file to initialize
// the database schema. For production use, consider using a proper
// migration tool like golang-migrate or goose.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - migrationsPath: Directory path containing schema.sql
//
// Returns:
//   - error: File read error or SQL execution error
//
// Thread-safety:
//   - Safe to call, but should only be run during server initialization
//   - Running migrations while server is active may cause issues
func (db *DB) RunMigrations(ctx context.Context, migrationsPath string) error {
	// Read the schema file
	schemaFile := filepath.Join(migrationsPath, "schema.sql")
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema
	if _, err := db.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// LoadUniverseData loads universe data from the generator into the database
func (db *DB) LoadUniverseData(ctx context.Context, systemsJSON io.Reader) error {
	// This will be implemented later when we integrate the universe generator
	// For now, it's a placeholder
	return nil
}

// ClearDatabase drops all tables in the database.
//
// WARNING: This is a destructive operation that deletes ALL data.
// Only use for testing or complete database resets.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - error: SQL execution error
//
// Thread-safety:
//   - Safe to call, but WILL destroy all data
//   - Should never be called in production
func (db *DB) ClearDatabase(ctx context.Context) error {
	tables := []string{
		"events",
		"chat_messages",
		"player_missions",
		"missions",
		"market_prices",
		"faction_reputation",
		"faction_officers",
		"faction_members",
		"player_factions",
		"ship_outfits",
		"ship_weapons",
		"ship_cargo",
		"ships",
		"planets",
		"system_connections",
		"star_systems",
		"player_reputation",
		"players",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// GetSchemaVersion returns the current schema version
func (db *DB) GetSchemaVersion(ctx context.Context) (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`

	err := db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}
