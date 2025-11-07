// File: internal/leaderboards/manager.go
// Project: Terminal Velocity
// Description: Leaderboard management and ranking system for competitive player statistics
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package leaderboards

import (
	"sort"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles leaderboard generation and ranking
//
// The manager maintains snapshots of leaderboards across all categories
// and can generate rankings from player data.

var log = logger.WithComponent("Leaderboards")

type Manager struct {
	snapshots  map[models.LeaderboardCategory]*models.LeaderboardSnapshot
	lastUpdate time.Time
}

// NewManager creates a new leaderboard manager
func NewManager() *Manager {
	return &Manager{
		snapshots:  make(map[models.LeaderboardCategory]*models.LeaderboardSnapshot),
		lastUpdate: time.Now(),
	}
}

// UpdateLeaderboard generates a new leaderboard snapshot for a specific category
//
// This method takes all players, calculates their scores for the given category,
// sorts them by score, and assigns ranks.
func (m *Manager) UpdateLeaderboard(category models.LeaderboardCategory, players []*models.Player) {
	// Create entries for all players
	entries := make([]*models.LeaderboardEntry, 0, len(players))

	for _, player := range players {
		score := models.GetScoreForCategory(player, category)

		// Skip players with zero score (haven't participated in this category)
		if score == 0 && category != models.LeaderboardWealth {
			continue
		}

		entry := models.NewLeaderboardEntry(player.ID, player.Username, category, score)
		entry.PopulateDetails(player)
		entries = append(entries, entry)
	}

	// Sort entries by score (descending)
	sort.Slice(entries, func(i, j int) bool {
		// If scores are equal, sort by player name for consistency
		if entries[i].Score == entries[j].Score {
			return entries[i].PlayerName < entries[j].PlayerName
		}
		return entries[i].Score > entries[j].Score
	})

	// Assign ranks
	for i, entry := range entries {
		entry.Rank = i + 1
	}

	// Create snapshot
	snapshot := &models.LeaderboardSnapshot{
		Category:     category,
		Entries:      entries,
		UpdatedAt:    time.Now(),
		TotalPlayers: len(entries),
	}

	m.snapshots[category] = snapshot
	m.lastUpdate = time.Now()
}

// UpdateAllLeaderboards generates snapshots for all leaderboard categories
func (m *Manager) UpdateAllLeaderboards(players []*models.Player) {
	categories := []models.LeaderboardCategory{
		models.LeaderboardCombat,
		models.LeaderboardTrading,
		models.LeaderboardExploration,
		models.LeaderboardWealth,
		models.LeaderboardReputation,
		models.LeaderboardMissions,
		models.LeaderboardOverall,
	}

	for _, category := range categories {
		m.UpdateLeaderboard(category, players)
	}
}

// GetLeaderboard returns the current snapshot for a specific category
//
// Returns nil if no snapshot exists for this category yet.
func (m *Manager) GetLeaderboard(category models.LeaderboardCategory) *models.LeaderboardSnapshot {
	return m.snapshots[category]
}

// GetTopEntries returns the top N entries for a specific category
func (m *Manager) GetTopEntries(category models.LeaderboardCategory, limit int) []*models.LeaderboardEntry {
	snapshot := m.GetLeaderboard(category)
	if snapshot == nil {
		return []*models.LeaderboardEntry{}
	}

	if limit > len(snapshot.Entries) {
		limit = len(snapshot.Entries)
	}

	return snapshot.Entries[:limit]
}

// GetPlayerRank returns a player's current rank in a specific category
//
// Returns 0 if the player is not ranked in this category.
func (m *Manager) GetPlayerRank(playerID uuid.UUID, category models.LeaderboardCategory) int {
	snapshot := m.GetLeaderboard(category)
	if snapshot == nil {
		return 0
	}

	for _, entry := range snapshot.Entries {
		if entry.PlayerID == playerID {
			return entry.Rank
		}
	}

	return 0
}

// GetPlayerEntry returns a player's full leaderboard entry for a category
//
// Returns nil if the player is not ranked in this category.
func (m *Manager) GetPlayerEntry(playerID uuid.UUID, category models.LeaderboardCategory) *models.LeaderboardEntry {
	snapshot := m.GetLeaderboard(category)
	if snapshot == nil {
		return nil
	}

	for _, entry := range snapshot.Entries {
		if entry.PlayerID == playerID {
			return entry
		}
	}

	return nil
}

// GetPlayerRankings returns a player's rankings across all categories
//
// This is useful for displaying a player's overall standing in the game.
func (m *Manager) GetPlayerRankings(playerID uuid.UUID) map[models.LeaderboardCategory]int {
	rankings := make(map[models.LeaderboardCategory]int)

	categories := []models.LeaderboardCategory{
		models.LeaderboardCombat,
		models.LeaderboardTrading,
		models.LeaderboardExploration,
		models.LeaderboardWealth,
		models.LeaderboardReputation,
		models.LeaderboardMissions,
		models.LeaderboardOverall,
	}

	for _, category := range categories {
		rankings[category] = m.GetPlayerRank(playerID, category)
	}

	return rankings
}

// GetLastUpdateTime returns when the leaderboards were last updated
func (m *Manager) GetLastUpdateTime() time.Time {
	return m.lastUpdate
}

// NeedsUpdate returns true if the leaderboards should be refreshed
//
// Leaderboards should be updated periodically to reflect recent player actions.
// Default update interval is 5 minutes.
func (m *Manager) NeedsUpdate(updateInterval time.Duration) bool {
	return time.Since(m.lastUpdate) > updateInterval
}

// GetLeaderboardsAroundPlayer returns leaderboard entries centered around a specific player
//
// This is useful for showing a player's position relative to nearby competitors.
// Returns entries from (player_rank - before) to (player_rank + after).
func (m *Manager) GetLeaderboardsAroundPlayer(
	playerID uuid.UUID,
	category models.LeaderboardCategory,
	before, after int,
) []*models.LeaderboardEntry {
	snapshot := m.GetLeaderboard(category)
	if snapshot == nil {
		return []*models.LeaderboardEntry{}
	}

	// Find player's position
	playerIndex := -1
	for i, entry := range snapshot.Entries {
		if entry.PlayerID == playerID {
			playerIndex = i
			break
		}
	}

	// Player not found in leaderboard
	if playerIndex == -1 {
		return []*models.LeaderboardEntry{}
	}

	// Calculate range
	start := playerIndex - before
	if start < 0 {
		start = 0
	}

	end := playerIndex + after + 1
	if end > len(snapshot.Entries) {
		end = len(snapshot.Entries)
	}

	return snapshot.Entries[start:end]
}

// GetCategoryCount returns the number of active leaderboard categories
func (m *Manager) GetCategoryCount() int {
	return len(m.snapshots)
}

// HasCategory returns true if a leaderboard snapshot exists for the given category
func (m *Manager) HasCategory(category models.LeaderboardCategory) bool {
	_, exists := m.snapshots[category]
	return exists
}

// GetAllCategories returns all available leaderboard categories
func GetAllCategories() []models.LeaderboardCategory {
	return []models.LeaderboardCategory{
		models.LeaderboardCombat,
		models.LeaderboardTrading,
		models.LeaderboardExploration,
		models.LeaderboardWealth,
		models.LeaderboardReputation,
		models.LeaderboardMissions,
		models.LeaderboardOverall,
	}
}
