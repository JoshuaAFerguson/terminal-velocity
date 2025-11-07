// File: internal/settings/manager.go
// Project: Terminal Velocity
// Description: Settings management and persistence
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles player settings
type Manager struct {
	mu           sync.RWMutex
	settings     map[uuid.UUID]*models.Settings // PlayerID -> Settings
	configDir    string                         // Directory for config files
	autosave     bool                           // Auto-save on change
}

// NewManager creates a new settings manager
func NewManager(configDir string) *Manager {
	// Create config directory if it doesn't exist
	if configDir != "" {
		os.MkdirAll(configDir, 0755)
	}

	return &Manager{
		settings:  make(map[uuid.UUID]*models.Settings),
		configDir: configDir,
		autosave:  true,
	}
}

// GetSettings retrieves settings for a player
func (m *Manager) GetSettings(playerID uuid.UUID) (*models.Settings, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if settings, exists := m.settings[playerID]; exists {
		return settings, nil
	}

	return nil, errors.New("settings not found")
}

// LoadSettings loads settings from disk for a player
func (m *Manager) LoadSettings(playerID uuid.UUID) (*models.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already loaded
	if settings, exists := m.settings[playerID]; exists {
		return settings, nil
	}

	// Try to load from file
	if m.configDir != "" {
		filename := filepath.Join(m.configDir, playerID.String()+".json")
		data, err := os.ReadFile(filename)
		if err == nil {
			var settings models.Settings
			if err := json.Unmarshal(data, &settings); err == nil {
				m.settings[playerID] = &settings
				return &settings, nil
			}
		}
	}

	// Create default settings if not found
	settings := models.NewSettings(playerID)
	m.settings[playerID] = settings

	// Save default settings
	if m.autosave {
		m.saveSettingsUnsafe(playerID)
	}

	return settings, nil
}

// SaveSettings saves settings to disk
func (m *Manager) SaveSettings(playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveSettingsUnsafe(playerID)
}

// saveSettingsUnsafe saves without locking (must be called with lock held)
func (m *Manager) saveSettingsUnsafe(playerID uuid.UUID) error {
	settings, exists := m.settings[playerID]
	if !exists {
		return errors.New("settings not found")
	}

	if m.configDir == "" {
		return nil // No persistence configured
	}

	filename := filepath.Join(m.configDir, playerID.String()+".json")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// UpdateSettings updates settings and optionally saves
func (m *Manager) UpdateSettings(playerID uuid.UUID, updater func(*models.Settings)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return errors.New("settings not found")
	}

	// Apply updates
	updater(settings)

	// Auto-save if enabled
	if m.autosave {
		return m.saveSettingsUnsafe(playerID)
	}

	return nil
}

// UpdateDisplaySettings updates display settings
func (m *Manager) UpdateDisplaySettings(playerID uuid.UUID, display models.DisplaySettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Display = display
	})
}

// UpdateAudioSettings updates audio settings
func (m *Manager) UpdateAudioSettings(playerID uuid.UUID, audio models.AudioSettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Audio = audio
	})
}

// UpdateGameplaySettings updates gameplay settings
func (m *Manager) UpdateGameplaySettings(playerID uuid.UUID, gameplay models.GameplaySettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Gameplay = gameplay
	})
}

// UpdateControlSettings updates control settings
func (m *Manager) UpdateControlSettings(playerID uuid.UUID, controls models.ControlSettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Controls = controls
	})
}

// UpdatePrivacySettings updates privacy settings
func (m *Manager) UpdatePrivacySettings(playerID uuid.UUID, privacy models.PrivacySettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Privacy = privacy
	})
}

// UpdateNotificationSettings updates notification settings
func (m *Manager) UpdateNotificationSettings(playerID uuid.UUID, notifications models.NotificationSettings) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.Notifications = notifications
	})
}

// ResetToDefaults resets settings to default values
func (m *Manager) ResetToDefaults(playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	settings := models.NewSettings(playerID)
	m.settings[playerID] = settings

	if m.autosave {
		return m.saveSettingsUnsafe(playerID)
	}

	return nil
}

// ExportSettings exports settings to JSON string
func (m *Manager) ExportSettings(playerID uuid.UUID) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return "", errors.New("settings not found")
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportSettings imports settings from JSON string
func (m *Manager) ImportSettings(playerID uuid.UUID, jsonData string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var settings models.Settings
	if err := json.Unmarshal([]byte(jsonData), &settings); err != nil {
		return err
	}

	// Override IDs to match player
	settings.PlayerID = playerID

	m.settings[playerID] = &settings

	if m.autosave {
		return m.saveSettingsUnsafe(playerID)
	}

	return nil
}

// GetColorScheme gets the active color scheme for a player
func (m *Manager) GetColorScheme(playerID uuid.UUID) models.ColorScheme {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return models.DefaultColorScheme()
	}

	return settings.ApplyColorScheme()
}

// BlockPlayer adds a player to block list
func (m *Manager) BlockPlayer(playerID uuid.UUID, targetID uuid.UUID) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.BlockPlayer(targetID)
	})
}

// UnblockPlayer removes a player from block list
func (m *Manager) UnblockPlayer(playerID uuid.UUID, targetID uuid.UUID) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.UnblockPlayer(targetID)
	})
}

// AddFriend adds a player to friends list
func (m *Manager) AddFriend(playerID uuid.UUID, targetID uuid.UUID) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.AddFriend(targetID)
	})
}

// RemoveFriend removes a player from friends list
func (m *Manager) RemoveFriend(playerID uuid.UUID, targetID uuid.UUID) error {
	return m.UpdateSettings(playerID, func(s *models.Settings) {
		s.RemoveFriend(targetID)
	})
}

// IsPlayerBlocked checks if a player is blocked
func (m *Manager) IsPlayerBlocked(playerID uuid.UUID, targetID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return false
	}

	return settings.IsPlayerBlocked(targetID)
}

// IsPlayerFriend checks if a player is a friend
func (m *Manager) IsPlayerFriend(playerID uuid.UUID, targetID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return false
	}

	return settings.IsPlayerFriend(targetID)
}

// CanReceiveTradeRequest checks if player accepts trade requests
func (m *Manager) CanReceiveTradeRequest(playerID uuid.UUID, fromPlayerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return true // Default to accepting
	}

	// Check if sender is blocked
	if settings.IsPlayerBlocked(fromPlayerID) {
		return false
	}

	return settings.Privacy.AllowTradeRequests
}

// CanReceivePvPChallenge checks if player accepts PvP challenges
func (m *Manager) CanReceivePvPChallenge(playerID uuid.UUID, fromPlayerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return true // Default to accepting
	}

	// Check if sender is blocked
	if settings.IsPlayerBlocked(fromPlayerID) {
		return false
	}

	return settings.Privacy.AllowPvPChallenges
}

// CanReceivePartyInvite checks if player accepts party invites
func (m *Manager) CanReceivePartyInvite(playerID uuid.UUID, fromPlayerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, exists := m.settings[playerID]
	if !exists {
		return true // Default to accepting
	}

	// Check if sender is blocked
	if settings.IsPlayerBlocked(fromPlayerID) {
		return false
	}

	return settings.Privacy.AllowPartyInvites
}

// GetStats returns settings manager statistics
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total_players": len(m.settings),
	}

	// Count color schemes
	colorSchemes := make(map[string]int)
	difficultyLevels := make(map[string]int)
	for _, settings := range m.settings {
		colorSchemes[settings.Display.ColorScheme]++
		difficultyLevels[settings.Gameplay.DifficultyLevel]++
	}

	stats["scheme_default"] = colorSchemes["default"]
	stats["scheme_dark"] = colorSchemes["dark"]
	stats["scheme_light"] = colorSchemes["light"]
	stats["diff_easy"] = difficultyLevels["easy"]
	stats["diff_normal"] = difficultyLevels["normal"]
	stats["diff_hard"] = difficultyLevels["hard"]
	stats["diff_expert"] = difficultyLevels["expert"]

	return stats
}

// SetAutosave enables or disables auto-saving
func (m *Manager) SetAutosave(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.autosave = enabled
}

// SaveAll saves all loaded settings to disk
func (m *Manager) SaveAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for playerID := range m.settings {
		if err := m.saveSettingsUnsafe(playerID); err != nil {
			return err
		}
	}

	return nil
}
