// File: internal/database/achievement_repository.go
// Project: Terminal Velocity
// Description: Database repository for player_achievements persistence.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// AchievementRepository persists which achievements each player has unlocked.
// The achievement catalog itself lives in `models.GetAllAchievements()` —
// this repo only records *which IDs* a given player has already hit, so the
// in-memory manager can skip them on subsequent logins and the Pilot Record
// screen can render a stable list across sessions.
type AchievementRepository struct {
	db *DB
}

// NewAchievementRepository wraps a *DB pool.
func NewAchievementRepository(db *DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

// LoadForPlayer returns every achievement the player has unlocked. Empty
// slice (not nil) when the player has none so callers can range-without-
// guard.
func (r *AchievementRepository) LoadForPlayer(ctx context.Context, playerID uuid.UUID) ([]*models.PlayerAchievement, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, player_id, achievement_id, unlocked_at
		 FROM player_achievements
		 WHERE player_id = $1
		 ORDER BY unlocked_at ASC`,
		playerID,
	)
	if err != nil {
		return nil, fmt.Errorf("load player achievements: %w", err)
	}
	defer rows.Close()

	out := make([]*models.PlayerAchievement, 0)
	for rows.Next() {
		var pa models.PlayerAchievement
		var unlocked time.Time
		if err := rows.Scan(&pa.ID, &pa.PlayerID, &pa.AchievementID, &unlocked); err != nil {
			return nil, fmt.Errorf("scan player achievement row: %w", err)
		}
		pa.UnlockedAt = unlocked
		pa.Progress = 100
		out = append(out, &pa)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player achievements: %w", err)
	}
	return out, nil
}

// Unlock records one achievement unlock. Idempotent via the
// (player_id, achievement_id) unique constraint — callers that race (e.g.
// two screens both calling checkAchievements) won't double-insert.
func (r *AchievementRepository) Unlock(ctx context.Context, playerID uuid.UUID, achievementID string) error {
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO player_achievements (player_id, achievement_id, unlocked_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (player_id, achievement_id) DO NOTHING`,
		playerID, achievementID,
	); err != nil {
		return fmt.Errorf("unlock achievement %s for %s: %w", achievementID, playerID, err)
	}
	return nil
}
