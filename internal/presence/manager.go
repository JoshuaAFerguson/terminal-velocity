// File: internal/presence/manager.go
// Project: Terminal Velocity
// Description: Manages player presence tracking for multiplayer interactions
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package presence

import (
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles tracking of online players and their activities
//
// This is the central system for multiplayer presence. It maintains a thread-safe
// map of all online players and provides methods to query presence information.
type Manager struct {
	mu       sync.RWMutex
	players  map[uuid.UUID]*models.PlayerPresence // Active player presences

	// Configuration
	afkThreshold    time.Duration // How long before a player is marked AFK
	offlineTimeout  time.Duration // How long before an inactive player is removed
}

// NewManager creates a new presence manager
func NewManager() *Manager {
	return &Manager{
		players:        make(map[uuid.UUID]*models.PlayerPresence),
		afkThreshold:   5 * time.Minute,
		offlineTimeout: 15 * time.Minute,
	}
}

// Connect registers a player as online
func (m *Manager) Connect(player *models.Player, ship *models.Ship) {
	m.mu.Lock()
	defer m.mu.Unlock()

	presence := models.NewPlayerPresence(player, ship)
	m.players[player.ID] = presence
}

// Disconnect marks a player as offline and removes them from the active list
func (m *Manager) Disconnect(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.players, playerID)
}

// UpdateActivity updates a player's current activity
func (m *Manager) UpdateActivity(playerID uuid.UUID, activity models.ActivityType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if presence, exists := m.players[playerID]; exists {
		presence.UpdateActivity(activity)
	}
}

// UpdateLocation updates a player's current system and planet
func (m *Manager) UpdateLocation(playerID uuid.UUID, systemID uuid.UUID, planetID *uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if presence, exists := m.players[playerID]; exists {
		presence.UpdateLocation(systemID, planetID)
	}
}

// UpdateShip updates a player's ship information
func (m *Manager) UpdateShip(playerID uuid.UUID, ship *models.Ship) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if presence, exists := m.players[playerID]; exists {
		if ship != nil {
			presence.ShipName = ship.Name
			presence.ShipType = ship.TypeID
		}
	}
}

// Heartbeat should be called periodically to update idle times and clean up stale presence
func (m *Manager) Heartbeat(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if presence, exists := m.players[playerID]; exists {
		presence.LastSeen = time.Now()
		presence.IdleDuration = 0
	}
}

// GetPresence returns a player's presence information
func (m *Manager) GetPresence(playerID uuid.UUID) *models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.players[playerID]
}

// GetAllOnline returns a list of all online players
func (m *Manager) GetAllOnline() []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*models.PlayerPresence, 0, len(m.players))
	for _, presence := range m.players {
		result = append(result, presence)
	}

	return result
}

// GetOnlineCount returns the number of online players
func (m *Manager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.players)
}

// GetPlayersInSystem returns all players in a specific system
func (m *Manager) GetPlayersInSystem(systemID uuid.UUID) []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []*models.PlayerPresence{}
	for _, presence := range m.players {
		if presence.IsInSameSystem(systemID) {
			result = append(result, presence)
		}
	}

	return result
}

// GetPlayersAtPlanet returns all players at a specific planet
func (m *Manager) GetPlayersAtPlanet(planetID uuid.UUID) []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []*models.PlayerPresence{}
	for _, presence := range m.players {
		if presence.IsAtSamePlanet(&planetID) {
			result = append(result, presence)
		}
	}

	return result
}

// GetPlayersByActivity returns all players with a specific activity
func (m *Manager) GetPlayersByActivity(activity models.ActivityType) []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []*models.PlayerPresence{}
	for _, presence := range m.players {
		if presence.CurrentActivity == string(activity) {
			result = append(result, presence)
		}
	}

	return result
}

// GetPlayersInCombat returns all players currently in combat
func (m *Manager) GetPlayersInCombat() []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []*models.PlayerPresence{}
	for _, presence := range m.players {
		if presence.InCombat {
			result = append(result, presence)
		}
	}

	return result
}

// GetNearbyPlayers returns players in the same system who can interact
//
// This excludes AFK players and filters by same system location
func (m *Manager) GetNearbyPlayers(playerID uuid.UUID) []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	presence, exists := m.players[playerID]
	if !exists {
		return []*models.PlayerPresence{}
	}

	result := []*models.PlayerPresence{}
	for id, other := range m.players {
		// Skip self
		if id == playerID {
			continue
		}

		// Check if in same system and can interact
		if other.IsInSameSystem(presence.CurrentSystem) && other.CanInteract() {
			result = append(result, other)
		}
	}

	return result
}

// IsPlayerOnline checks if a specific player is currently online
func (m *Manager) IsPlayerOnline(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.players[playerID]
	return exists
}

// CleanupStale removes players who haven't been seen for a while
//
// This should be called periodically (e.g., every minute) to clean up
// disconnected players whose sessions weren't properly closed.
func (m *Manager) CleanupStale() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	stale := []uuid.UUID{}

	for id, presence := range m.players {
		if now.Sub(presence.LastSeen) > m.offlineTimeout {
			stale = append(stale, id)
		}
	}

	for _, id := range stale {
		delete(m.players, id)
	}
}

// UpdateIdleTimes updates idle duration for all players
//
// This should be called periodically to keep idle times accurate
func (m *Manager) UpdateIdleTimes() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, presence := range m.players {
		presence.UpdateIdleTime()
	}
}

// GetSystemPlayerCount returns the number of players in a specific system
func (m *Manager) GetSystemPlayerCount(systemID uuid.UUID) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, presence := range m.players {
		if presence.IsInSameSystem(systemID) {
			count++
		}
	}

	return count
}

// GetRecentlyActive returns players active within the last N minutes
func (m *Manager) GetRecentlyActive(minutes int) []*models.PlayerPresence {
	m.mu.RLock()
	defer m.mu.RUnlock()

	threshold := time.Duration(minutes) * time.Minute
	result := []*models.PlayerPresence{}

	for _, presence := range m.players {
		if time.Since(presence.LastSeen) < threshold {
			result = append(result, presence)
		}
	}

	return result
}

// GetStats returns statistics about online players
func (m *Manager) GetStats() PresenceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := PresenceStats{
		TotalOnline: len(m.players),
		InCombat:    0,
		Trading:     0,
		Docked:      0,
		InSpace:     0,
		Afk:         0,
	}

	for _, presence := range m.players {
		if presence.InCombat {
			stats.InCombat++
		}

		if presence.IsAfk() {
			stats.Afk++
		}

		if presence.Docked {
			stats.Docked++
		} else {
			stats.InSpace++
		}

		if presence.CurrentActivity == string(models.ActivityTrading) {
			stats.Trading++
		}
	}

	return stats
}

// PresenceStats contains statistics about online player activity
type PresenceStats struct {
	TotalOnline int `json:"total_online"`
	InCombat    int `json:"in_combat"`
	Trading     int `json:"trading"`
	Docked      int `json:"docked"`
	InSpace     int `json:"in_space"`
	Afk         int `json:"afk"`
}
