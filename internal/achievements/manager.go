// File: internal/achievements/manager.go
// Project: Terminal Velocity
// Description: Achievement tracking system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package achievements provides achievement tracking and management.
//
// This package handles:
// - Achievement unlock detection
// - Progress tracking toward achievements
// - Achievement point calculation
// - Filtering by category and status
//
// Version: 1.0.0
// Last Updated: 2025-01-07
package achievements

import (
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles achievement tracking for a player

type Manager struct {
	unlocked map[string]*models.PlayerAchievement // Map of achievement ID to unlock data
}

// NewManager creates a new achievement manager
//
// Returns:
//   - Pointer to new Manager
func NewManager() *Manager {
	return &Manager{
		unlocked: make(map[string]*models.PlayerAchievement),
	}
}

// LoadUnlocked loads a player's unlocked achievements
//
// Parameters:
//   - achievements: Slice of player achievement records
func (m *Manager) LoadUnlocked(achievements []*models.PlayerAchievement) {
	m.unlocked = make(map[string]*models.PlayerAchievement)
	for _, pa := range achievements {
		m.unlocked[pa.AchievementID] = pa
	}
}

// CheckNewUnlocks checks if the player has unlocked any new achievements
//
// Parameters:
//   - player: Player to check
//
// Returns:
//   - Slice of newly unlocked achievements
func (m *Manager) CheckNewUnlocks(player *models.Player) []*models.Achievement {
	newUnlocks := []*models.Achievement{}
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		// Skip if already unlocked
		if _, exists := m.unlocked[achievement.ID]; exists {
			continue
		}

		// Check if criteria met
		if achievement.IsUnlocked(player) {
			// Record unlock
			unlock := &models.PlayerAchievement{
				ID:            uuid.New(),
				PlayerID:      player.ID,
				AchievementID: achievement.ID,
				UnlockedAt:    time.Now(),
				Progress:      100,
			}
			m.unlocked[achievement.ID] = unlock
			newUnlocks = append(newUnlocks, achievement)
		}
	}

	return newUnlocks
}

// IsUnlocked checks if a specific achievement is unlocked
//
// Parameters:
//   - achievementID: ID of achievement to check
//
// Returns:
//   - true if unlocked
func (m *Manager) IsUnlocked(achievementID string) bool {
	_, exists := m.unlocked[achievementID]
	return exists
}

// GetUnlockedAchievements returns all unlocked achievements
//
// Returns:
//   - Slice of unlocked achievements with unlock data
func (m *Manager) GetUnlockedAchievements() []*models.Achievement {
	unlocked := []*models.Achievement{}
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		if m.IsUnlocked(achievement.ID) {
			unlocked = append(unlocked, achievement)
		}
	}

	return unlocked
}

// GetLockedAchievements returns all locked achievements
//
// Parameters:
//   - includeHidden: If false, excludes hidden achievements
//
// Returns:
//   - Slice of locked achievements
func (m *Manager) GetLockedAchievements(includeHidden bool) []*models.Achievement {
	locked := []*models.Achievement{}
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		if !m.IsUnlocked(achievement.ID) {
			// Skip hidden achievements if requested
			if !includeHidden && achievement.Hidden {
				continue
			}
			locked = append(locked, achievement)
		}
	}

	return locked
}

// GetAchievementsByCategory returns achievements filtered by category
//
// Parameters:
//   - category: Category to filter by
//   - unlockedOnly: If true, only returns unlocked achievements
//
// Returns:
//   - Slice of achievements in category
func (m *Manager) GetAchievementsByCategory(category models.AchievementCategory, unlockedOnly bool) []*models.Achievement {
	filtered := []*models.Achievement{}
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		if achievement.Category != category {
			continue
		}

		if unlockedOnly && !m.IsUnlocked(achievement.ID) {
			continue
		}

		filtered = append(filtered, achievement)
	}

	return filtered
}

// GetTotalPoints calculates total achievement points earned
//
// Returns:
//   - Total points from unlocked achievements
func (m *Manager) GetTotalPoints() int {
	total := 0
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		if m.IsUnlocked(achievement.ID) {
			total += achievement.Points
		}
	}

	return total
}

// GetMaxPoints calculates maximum possible achievement points
//
// Returns:
//   - Total points from all achievements
func GetMaxPoints() int {
	total := 0
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		total += achievement.Points
	}

	return total
}

// GetUnlockCount returns number of unlocked achievements
//
// Returns:
//   - Count of unlocked achievements
func (m *Manager) GetUnlockCount() int {
	return len(m.unlocked)
}

// GetTotalCount returns total number of achievements
//
// Returns:
//   - Total achievement count
func GetTotalCount() int {
	return len(models.GetAllAchievements())
}

// GetProgress calculates progress toward a specific achievement
//
// Parameters:
//   - achievementID: ID of achievement to check
//   - player: Player to check progress for
//
// Returns:
//   - Progress percentage (0-100), or 100 if unlocked
func (m *Manager) GetProgress(achievementID string, player *models.Player) int {
	// If unlocked, return 100
	if m.IsUnlocked(achievementID) {
		return 100
	}

	// Find achievement
	allAchievements := models.GetAllAchievements()
	for _, achievement := range allAchievements {
		if achievement.ID == achievementID {
			return achievement.CalculateProgress(player)
		}
	}

	return 0
}

// GetUnlockedPlayerAchievements returns the raw player achievement records
// for database persistence
//
// Returns:
//   - Slice of PlayerAchievement records
func (m *Manager) GetUnlockedPlayerAchievements() []*models.PlayerAchievement {
	achievements := make([]*models.PlayerAchievement, 0, len(m.unlocked))
	for _, pa := range m.unlocked {
		achievements = append(achievements, pa)
	}
	return achievements
}

// GetRecentUnlocks returns the N most recently unlocked achievements
//
// Parameters:
//   - count: Number of recent unlocks to return
//
// Returns:
//   - Slice of recently unlocked achievements, sorted by unlock time
func (m *Manager) GetRecentUnlocks(count int) []*models.Achievement {
	// Get all unlocked with timestamps
	type unlockTime struct {
		achievement *models.Achievement
		unlockedAt  time.Time
	}

	unlocks := []unlockTime{}
	allAchievements := models.GetAllAchievements()

	for _, achievement := range allAchievements {
		if pa, exists := m.unlocked[achievement.ID]; exists {
			unlocks = append(unlocks, unlockTime{
				achievement: achievement,
				unlockedAt:  pa.UnlockedAt,
			})
		}
	}

	// Sort by unlock time (most recent first) - bubble sort for simplicity
	for i := 0; i < len(unlocks)-1; i++ {
		for j := 0; j < len(unlocks)-i-1; j++ {
			if unlocks[j].unlockedAt.Before(unlocks[j+1].unlockedAt) {
				unlocks[j], unlocks[j+1] = unlocks[j+1], unlocks[j]
			}
		}
	}

	// Return up to count achievements
	result := []*models.Achievement{}
	for i := 0; i < len(unlocks) && i < count; i++ {
		result = append(result, unlocks[i].achievement)
	}

	return result
}
