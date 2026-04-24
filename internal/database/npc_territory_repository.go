// File: internal/database/npc_territory_repository.go
// Project: Terminal Velocity
// Description: Persistence for NPC faction territorial control.
//   One row per system holding its current owner; written through
//   on every war-resolution flip so restarts preserve the galaxy's
//   political state.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package database

import (
	"context"
	"fmt"
	"time"
)

// NPCTerritoryRow mirrors the npc_territory schema. Kept outside
// the npcterritory package (and equally, kept out of models) so
// neither side needs the other's types — the server-layer adapter
// translates between NPCTerritoryRow and npcterritory.Manager's
// in-memory representation, same pattern as AuctionRow <-> the
// marketplace manager.
type NPCTerritoryRow struct {
	SystemName string
	FactionID  string
	UpdatedAt  time.Time
}

// NPCTerritoryRepository persists NPC faction control of star
// systems. Writes are idempotent (upsert on system_name), so the
// manager's persister hook can fire on every mutation without
// worrying about row existence.
type NPCTerritoryRepository struct {
	db *DB
}

// NewNPCTerritoryRepository wraps a *DB pool.
func NewNPCTerritoryRepository(db *DB) *NPCTerritoryRepository {
	return &NPCTerritoryRepository{db: db}
}

// UpsertOwnership writes or updates a territory row. Called on
// every flip from the manager's persister hook. The updated_at
// column is always refreshed so the restart-recovery path can
// potentially order ops by recency if needed later.
func (r *NPCTerritoryRepository) UpsertOwnership(ctx context.Context, row *NPCTerritoryRow) error {
	if r == nil || r.db == nil {
		return nil
	}
	if row == nil || row.SystemName == "" || row.FactionID == "" {
		return fmt.Errorf("npc_territory: system_name and faction_id are required")
	}
	const q = `
		INSERT INTO npc_territory (system_name, faction_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (system_name) DO UPDATE
		SET faction_id = EXCLUDED.faction_id, updated_at = NOW()
	`
	if _, err := r.db.ExecContext(ctx, q, row.SystemName, row.FactionID); err != nil {
		return fmt.Errorf("upsert npc_territory: %w", err)
	}
	return nil
}

// LoadAll returns every persisted territory row. Called once at
// server startup, AFTER the manager has seeded from the static
// StandardNPCFactions data — any row returned here overrides the
// seed value for its system, which is how war-captured systems
// survive restart.
func (r *NPCTerritoryRepository) LoadAll(ctx context.Context) ([]*NPCTerritoryRow, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	const q = `SELECT system_name, faction_id, updated_at FROM npc_territory`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("load npc_territory: %w", err)
	}
	defer rows.Close()

	var out []*NPCTerritoryRow
	for rows.Next() {
		row := &NPCTerritoryRow{}
		if err := rows.Scan(&row.SystemName, &row.FactionID, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan npc_territory: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate npc_territory: %w", err)
	}
	return out, nil
}
