// File: internal/admin/manager.go
// Project: Terminal Velocity
// Description: Server administration and monitoring
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package admin

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles server administration

var log = logger.WithComponent("Admin")

type Manager struct {
	mu sync.RWMutex

	// Admin users
	admins map[uuid.UUID]*models.AdminUser // PlayerID -> AdminUser

	// Moderation
	bans  map[uuid.UUID]*models.PlayerBan  // PlayerID -> Ban
	mutes map[uuid.UUID]*models.PlayerMute // PlayerID -> Mute

	// Audit log
	actionLog []*models.AdminAction

	// Server settings
	settings *models.ServerSettings

	// Metrics
	metrics *models.ServerMetrics

	// Repositories
	playerRepo *database.PlayerRepository

	// Metrics collection
	metricsInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewManager creates a new admin manager
func NewManager(playerRepo *database.PlayerRepository) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		admins:          make(map[uuid.UUID]*models.AdminUser),
		bans:            make(map[uuid.UUID]*models.PlayerBan),
		mutes:           make(map[uuid.UUID]*models.PlayerMute),
		actionLog:       make([]*models.AdminAction, 0),
		settings:        models.GetDefaultServerSettings(),
		metrics:         &models.ServerMetrics{},
		playerRepo:      playerRepo,
		metricsInterval: 10 * time.Second,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start metrics collection
	m.wg.Add(1)
	go m.metricsWorker()

	return m
}

// AddAdmin adds an admin user
func (m *Manager) AddAdmin(playerID uuid.UUID, username string, role models.AdminRole, createdBy uuid.UUID) (*models.AdminUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already admin
	if _, exists := m.admins[playerID]; exists {
		return nil, errors.New("player is already an admin")
	}

	admin := models.NewAdminUser(playerID, username, role, createdBy)
	m.admins[playerID] = admin

	// Log action
	m.logActionUnsafe(createdBy, "add_admin", playerID, username, "Added admin with role: "+string(role))

	return admin, nil
}

// RemoveAdmin removes an admin user
func (m *Manager) RemoveAdmin(adminID uuid.UUID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	admin, exists := m.admins[adminID]
	if !exists || !admin.IsActive {
		return errors.New("not authorized")
	}

	target, exists := m.admins[targetID]
	if !exists {
		return errors.New("admin not found")
	}

	// Can't remove superadmin unless you're a superadmin
	if target.Role == models.RoleSuperAdmin && admin.Role != models.RoleSuperAdmin {
		return errors.New("cannot remove superadmin")
	}

	delete(m.admins, targetID)

	// Log action
	m.logActionUnsafe(adminID, "remove_admin", targetID, target.Username, "Removed admin")

	return nil
}

// IsAdmin checks if a player is an admin
func (m *Manager) IsAdmin(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	admin, exists := m.admins[playerID]
	return exists && admin.IsActive
}

// HasPermission checks if a player has a specific permission
func (m *Manager) HasPermission(playerID uuid.UUID, permission models.AdminPermission) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	admin, exists := m.admins[playerID]
	if !exists || !admin.IsActive {
		return false
	}

	return admin.HasPermission(permission)
}

// BanPlayer bans a player
func (m *Manager) BanPlayer(
	adminID uuid.UUID,
	targetID uuid.UUID,
	username string,
	ipAddress string,
	reason string,
	duration *time.Duration,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check permission
	admin, exists := m.admins[adminID]
	if !exists || !admin.HasPermission(models.PermBanPlayer) {
		return errors.New("not authorized")
	}

	// Can't ban other admins
	if _, isAdmin := m.admins[targetID]; isAdmin {
		return errors.New("cannot ban admin users")
	}

	// Create ban
	ban := models.NewPlayerBan(targetID, username, ipAddress, reason, adminID, duration)
	m.bans[targetID] = ban

	// Log action
	m.logActionUnsafe(adminID, "ban_player", targetID, username, "Banned: "+reason)

	return nil
}

// UnbanPlayer unbans a player
func (m *Manager) UnbanPlayer(adminID uuid.UUID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check permission
	admin, exists := m.admins[adminID]
	if !exists || !admin.HasPermission(models.PermBanPlayer) {
		return errors.New("not authorized")
	}

	ban, exists := m.bans[targetID]
	if !exists {
		return errors.New("player not banned")
	}

	ban.IsActive = false

	// Log action
	m.logActionUnsafe(adminID, "unban_player", targetID, ban.Username, "Unbanned player")

	return nil
}

// IsBanned checks if a player is banned
func (m *Manager) IsBanned(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ban, exists := m.bans[playerID]
	if !exists || !ban.IsActive {
		return false
	}

	// Check if expired
	if ban.IsExpired() {
		return false
	}

	return true
}

// MutePlayer mutes a player
func (m *Manager) MutePlayer(
	adminID uuid.UUID,
	targetID uuid.UUID,
	username string,
	reason string,
	duration time.Duration,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check permission
	admin, exists := m.admins[adminID]
	if !exists || !admin.HasPermission(models.PermMutePlayer) {
		return errors.New("not authorized")
	}

	// Create mute
	mute := models.NewPlayerMute(targetID, username, reason, adminID, duration)
	m.mutes[targetID] = mute

	// Log action
	m.logActionUnsafe(adminID, "mute_player", targetID, username, "Muted: "+reason)

	return nil
}

// UnmutePlayer unmutes a player
func (m *Manager) UnmutePlayer(adminID uuid.UUID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check permission
	admin, exists := m.admins[adminID]
	if !exists || !admin.HasPermission(models.PermMutePlayer) {
		return errors.New("not authorized")
	}

	mute, exists := m.mutes[targetID]
	if !exists {
		return errors.New("player not muted")
	}

	mute.IsActive = false

	// Log action
	m.logActionUnsafe(adminID, "unmute_player", targetID, mute.Username, "Unmuted player")

	return nil
}

// IsMuted checks if a player is muted
func (m *Manager) IsMuted(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mute, exists := m.mutes[playerID]
	if !exists || !mute.IsActive {
		return false
	}

	// Check if expired
	if mute.IsExpired() {
		return false
	}

	return true
}

// UpdateSettings updates server settings
func (m *Manager) UpdateSettings(adminID uuid.UUID, settings *models.ServerSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check permission
	admin, exists := m.admins[adminID]
	if !exists || !admin.HasPermission(models.PermServerSettings) {
		return errors.New("not authorized")
	}

	m.settings = settings

	// Log action
	m.logActionUnsafe(adminID, "update_settings", uuid.Nil, "", "Updated server settings")

	return nil
}

// GetSettings returns current server settings
func (m *Manager) GetSettings() *models.ServerSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	settings := *m.settings
	return &settings
}

// GetMetrics returns current server metrics
func (m *Manager) GetMetrics() *models.ServerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	metrics := *m.metrics
	return &metrics
}

// UpdateMetrics updates server metrics (called by metrics worker)
func (m *Manager) UpdateMetrics(metrics *models.ServerMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = metrics
}

// GetActionLog returns recent admin actions
func (m *Manager) GetActionLog(limit int) []*models.AdminAction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	start := 0
	if len(m.actionLog) > limit {
		start = len(m.actionLog) - limit
	}

	// Return a copy
	result := make([]*models.AdminAction, len(m.actionLog[start:]))
	copy(result, m.actionLog[start:])

	return result
}

// LogAction logs an admin action
func (m *Manager) LogAction(action *models.AdminAction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actionLog = append(m.actionLog, action)

	// Trim log if too large
	if len(m.actionLog) > 10000 {
		m.actionLog = m.actionLog[1000:]
	}
}

// logActionUnsafe logs an action (must be called with lock held)
func (m *Manager) logActionUnsafe(adminID uuid.UUID, action string, targetID uuid.UUID, targetName string, details string) {
	adminName := "system"
	if admin, exists := m.admins[adminID]; exists {
		adminName = admin.Username
	}

	logEntry := models.NewAdminAction(adminID, adminName, action, "")
	if targetID != uuid.Nil {
		logEntry.SetTarget(targetID, targetName)
	}
	logEntry.Details = details

	m.actionLog = append(m.actionLog, logEntry)

	// Trim log if too large
	if len(m.actionLog) > 10000 {
		m.actionLog = m.actionLog[1000:]
	}
}

// metricsWorker collects server metrics periodically
func (m *Manager) metricsWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics gathers current server metrics
func (m *Manager) collectMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	metrics := &models.ServerMetrics{
		Timestamp:      time.Now(),
		GoroutineCount: runtime.NumGoroutine(),
		MemoryUsage:    int64(mem.Alloc),
	}

	// Note: In production, these would be populated from actual game state
	// For now, we just update the structure

	m.UpdateMetrics(metrics)
}

// GetActiveBans returns all active bans
func (m *Manager) GetActiveBans() []*models.PlayerBan {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bans := make([]*models.PlayerBan, 0)
	for _, ban := range m.bans {
		if ban.IsActive && !ban.IsExpired() {
			bans = append(bans, ban)
		}
	}

	return bans
}

// GetActiveMutes returns all active mutes
func (m *Manager) GetActiveMutes() []*models.PlayerMute {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mutes := make([]*models.PlayerMute, 0)
	for _, mute := range m.mutes {
		if mute.IsActive && !mute.IsExpired() {
			mutes = append(mutes, mute)
		}
	}

	return mutes
}

// GetAdmins returns all admin users
func (m *Manager) GetAdmins() []*models.AdminUser {
	m.mu.RLock()
	defer m.mu.RUnlock()

	admins := make([]*models.AdminUser, 0, len(m.admins))
	for _, admin := range m.admins {
		admins = append(admins, admin)
	}

	return admins
}

// Shutdown gracefully shuts down the admin manager
func (m *Manager) Shutdown() {
	m.cancel()
	m.wg.Wait()
}

// GetStats returns admin manager statistics
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeBans := 0
	for _, ban := range m.bans {
		if ban.IsActive && !ban.IsExpired() {
			activeBans++
		}
	}

	activeMutes := 0
	for _, mute := range m.mutes {
		if mute.IsActive && !mute.IsExpired() {
			activeMutes++
		}
	}

	return map[string]int{
		"total_admins":  len(m.admins),
		"active_bans":   activeBans,
		"active_mutes":  activeMutes,
		"total_actions": len(m.actionLog),
	}
}
