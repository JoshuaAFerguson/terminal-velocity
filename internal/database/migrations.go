// File: internal/database/migrations.go
// Project: Terminal Velocity
// Description: Database repository for migrations
// Version: 1.0.0
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

// RunMigrations executes SQL migration files

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

// ClearDatabase drops all tables (use with caution!)
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
// This is a placeholder for future migration versioning
func (db *DB) GetSchemaVersion(ctx context.Context) (int, error) {
	// TODO: Implement schema versioning table
	return 1, nil
}
