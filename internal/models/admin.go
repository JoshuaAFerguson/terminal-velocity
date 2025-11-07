// File: internal/models/admin.go
// Project: Terminal Velocity
// Description: Admin models and permissions
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// AdminRole represents different admin privilege levels
type AdminRole string

const (
	RolePlayer      AdminRole = "player"       // Normal player (no admin)
	RoleModerator   AdminRole = "moderator"    // Can mute/kick players
	RoleAdmin       AdminRole = "admin"        // Can manage server settings
	RoleSuperAdmin  AdminRole = "superadmin"   // Full server control
)

// AdminPermission represents specific admin permissions
type AdminPermission string

const (
	// Player management
	PermKickPlayer      AdminPermission = "kick_player"
	PermBanPlayer       AdminPermission = "ban_player"
	PermMutePlayer      AdminPermission = "mute_player"
	PermViewPlayerData  AdminPermission = "view_player_data"
	PermEditPlayerData  AdminPermission = "edit_player_data"

	// Server management
	PermServerSettings  AdminPermission = "server_settings"
	PermServerShutdown  AdminPermission = "server_shutdown"
	PermServerRestart   AdminPermission = "server_restart"
	PermViewLogs        AdminPermission = "view_logs"
	PermExecuteCommands AdminPermission = "execute_commands"

	// Content management
	PermEditEconomy     AdminPermission = "edit_economy"
	PermSpawnItems      AdminPermission = "spawn_items"
	PermEditSystems     AdminPermission = "edit_systems"
	PermEditFactions    AdminPermission = "edit_factions"

	// Monitoring
	PermViewMetrics     AdminPermission = "view_metrics"
	PermViewSessions    AdminPermission = "view_sessions"
	PermViewDatabase    AdminPermission = "view_database"

	// Communication
	PermBroadcast       AdminPermission = "broadcast"
	PermViewAllChat     AdminPermission = "view_all_chat"
)

// AdminUser represents an admin user
type AdminUser struct {
	ID          uuid.UUID       `json:"id"`
	PlayerID    uuid.UUID       `json:"player_id"`
	Username    string          `json:"username"`
	Role        AdminRole       `json:"role"`
	Permissions []AdminPermission `json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	LastActive  time.Time       `json:"last_active"`
	IsActive    bool            `json:"is_active"`
}

// AdminAction represents an admin action for audit logging
type AdminAction struct {
	ID          uuid.UUID       `json:"id"`
	AdminID     uuid.UUID       `json:"admin_id"`
	AdminName   string          `json:"admin_name"`
	Action      string          `json:"action"`
	TargetID    uuid.UUID       `json:"target_id,omitempty"`
	TargetName  string          `json:"target_name,omitempty"`
	Details     string          `json:"details"`
	Timestamp   time.Time       `json:"timestamp"`
	IPAddress   string          `json:"ip_address"`
	Success     bool            `json:"success"`
	ErrorMsg    string          `json:"error_msg,omitempty"`
}

// PlayerBan represents a banned player
type PlayerBan struct {
	ID          uuid.UUID `json:"id"`
	PlayerID    uuid.UUID `json:"player_id"`
	Username    string    `json:"username"`
	IPAddress   string    `json:"ip_address"`
	Reason      string    `json:"reason"`
	BannedBy    uuid.UUID `json:"banned_by"`
	BannedAt    time.Time `json:"banned_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsPermanent bool      `json:"is_permanent"`
	IsActive    bool      `json:"is_active"`
}

// PlayerMute represents a muted player
type PlayerMute struct {
	ID        uuid.UUID `json:"id"`
	PlayerID  uuid.UUID `json:"player_id"`
	Username  string    `json:"username"`
	Reason    string    `json:"reason"`
	MutedBy   uuid.UUID `json:"muted_by"`
	MutedAt   time.Time `json:"muted_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// ServerMetrics represents server performance metrics
type ServerMetrics struct {
	Timestamp       time.Time `json:"timestamp"`

	// Player metrics
	TotalPlayers    int       `json:"total_players"`
	ActivePlayers   int       `json:"active_players"`
	PeakPlayers     int       `json:"peak_players"`
	NewPlayersToday int       `json:"new_players_today"`

	// Session metrics
	ActiveSessions  int       `json:"active_sessions"`
	AvgSessionTime  int64     `json:"avg_session_time"` // seconds
	TotalCommands   int       `json:"total_commands"`

	// Performance metrics
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     int64     `json:"memory_usage"` // bytes
	GoroutineCount  int       `json:"goroutine_count"`

	// Game metrics
	ActiveTrades    int       `json:"active_trades"`
	ActiveCombats   int       `json:"active_combats"`
	TotalJumps      int       `json:"total_jumps"`
	TotalTransactions int     `json:"total_transactions"`

	// Database metrics
	DBConnections   int       `json:"db_connections"`
	DBLatency       int64     `json:"db_latency"` // milliseconds
	DBErrors        int       `json:"db_errors"`
}

// ServerSettings represents configurable server settings
type ServerSettings struct {
	// General
	ServerName        string  `json:"server_name"`
	MOTD              string  `json:"motd"` // Message of the day
	MaxPlayers        int     `json:"max_players"`
	TickRate          int     `json:"tick_rate"` // Updates per second

	// Economy
	StartingCredits   int64   `json:"starting_credits"`
	EconomyMultiplier float64 `json:"economy_multiplier"`
	TaxRate           float64 `json:"tax_rate"`

	// Difficulty
	CombatDifficulty  float64 `json:"combat_difficulty"`
	PirateFrequency   float64 `json:"pirate_frequency"`
	PriceVolatility   float64 `json:"price_volatility"`

	// Rules
	PvPEnabled        bool    `json:"pvp_enabled"`
	PermadeathMode    bool    `json:"permadeath_mode"`
	FriendlyFire      bool    `json:"friendly_fire"`

	// Limits
	MaxShipsPerPlayer int     `json:"max_ships_per_player"`
	MaxCargoSpace     int     `json:"max_cargo_space"`
	MaxCredits        int64   `json:"max_credits"`

	// Timeouts
	SessionTimeout    int     `json:"session_timeout"` // minutes
	AutosaveInterval  int     `json:"autosave_interval"` // seconds
	CleanupInterval   int     `json:"cleanup_interval"` // minutes

	// Features
	EnableEncounters  bool    `json:"enable_encounters"`
	EnableFactions    bool    `json:"enable_factions"`
	EnableAchievements bool   `json:"enable_achievements"`
	EnableLeaderboards bool   `json:"enable_leaderboards"`
}

// NewAdminUser creates a new admin user
func NewAdminUser(playerID uuid.UUID, username string, role AdminRole, createdBy uuid.UUID) *AdminUser {
	return &AdminUser{
		ID:          uuid.New(),
		PlayerID:    playerID,
		Username:    username,
		Role:        role,
		Permissions: GetDefaultPermissions(role),
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		LastActive:  time.Now(),
		IsActive:    true,
	}
}

// GetDefaultPermissions returns default permissions for a role
func GetDefaultPermissions(role AdminRole) []AdminPermission {
	switch role {
	case RoleModerator:
		return []AdminPermission{
			PermKickPlayer,
			PermMutePlayer,
			PermViewPlayerData,
			PermViewMetrics,
			PermViewSessions,
			PermBroadcast,
			PermViewAllChat,
		}
	case RoleAdmin:
		return []AdminPermission{
			PermKickPlayer,
			PermBanPlayer,
			PermMutePlayer,
			PermViewPlayerData,
			PermEditPlayerData,
			PermServerSettings,
			PermViewLogs,
			PermEditEconomy,
			PermSpawnItems,
			PermViewMetrics,
			PermViewSessions,
			PermViewDatabase,
			PermBroadcast,
			PermViewAllChat,
		}
	case RoleSuperAdmin:
		return []AdminPermission{
			PermKickPlayer,
			PermBanPlayer,
			PermMutePlayer,
			PermViewPlayerData,
			PermEditPlayerData,
			PermServerSettings,
			PermServerShutdown,
			PermServerRestart,
			PermViewLogs,
			PermExecuteCommands,
			PermEditEconomy,
			PermSpawnItems,
			PermEditSystems,
			PermEditFactions,
			PermViewMetrics,
			PermViewSessions,
			PermViewDatabase,
			PermBroadcast,
			PermViewAllChat,
		}
	default:
		return []AdminPermission{}
	}
}

// HasPermission checks if admin has a specific permission
func (a *AdminUser) HasPermission(perm AdminPermission) bool {
	if !a.IsActive {
		return false
	}

	for _, p := range a.Permissions {
		if p == perm {
			return true
		}
	}

	return false
}

// AddPermission adds a permission to the admin
func (a *AdminUser) AddPermission(perm AdminPermission) {
	if !a.HasPermission(perm) {
		a.Permissions = append(a.Permissions, perm)
	}
}

// RemovePermission removes a permission from the admin
func (a *AdminUser) RemovePermission(perm AdminPermission) {
	for i, p := range a.Permissions {
		if p == perm {
			a.Permissions = append(a.Permissions[:i], a.Permissions[i+1:]...)
			return
		}
	}
}

// NewAdminAction creates a new admin action for logging
func NewAdminAction(adminID uuid.UUID, adminName, action string, ipAddress string) *AdminAction {
	return &AdminAction{
		ID:        uuid.New(),
		AdminID:   adminID,
		AdminName: adminName,
		Action:    action,
		Timestamp: time.Now(),
		IPAddress: ipAddress,
		Success:   true,
	}
}

// SetTarget sets the target of the admin action
func (a *AdminAction) SetTarget(targetID uuid.UUID, targetName string) {
	a.TargetID = targetID
	a.TargetName = targetName
}

// SetError marks the action as failed with an error message
func (a *AdminAction) SetError(err error) {
	a.Success = false
	if err != nil {
		a.ErrorMsg = err.Error()
	}
}

// NewPlayerBan creates a new player ban
func NewPlayerBan(playerID uuid.UUID, username, ipAddress, reason string, bannedBy uuid.UUID, duration *time.Duration) *PlayerBan {
	ban := &PlayerBan{
		ID:          uuid.New(),
		PlayerID:    playerID,
		Username:    username,
		IPAddress:   ipAddress,
		Reason:      reason,
		BannedBy:    bannedBy,
		BannedAt:    time.Now(),
		IsPermanent: duration == nil,
		IsActive:    true,
	}

	if duration != nil {
		expiresAt := time.Now().Add(*duration)
		ban.ExpiresAt = &expiresAt
	}

	return ban
}

// IsExpired checks if the ban has expired
func (b *PlayerBan) IsExpired() bool {
	if b.IsPermanent {
		return false
	}
	if b.ExpiresAt == nil {
		return true
	}
	return time.Now().After(*b.ExpiresAt)
}

// NewPlayerMute creates a new player mute
func NewPlayerMute(playerID uuid.UUID, username, reason string, mutedBy uuid.UUID, duration time.Duration) *PlayerMute {
	return &PlayerMute{
		ID:        uuid.New(),
		PlayerID:  playerID,
		Username:  username,
		Reason:    reason,
		MutedBy:   mutedBy,
		MutedAt:   time.Now(),
		ExpiresAt: time.Now().Add(duration),
		IsActive:  true,
	}
}

// IsExpired checks if the mute has expired
func (m *PlayerMute) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

// GetDefaultServerSettings returns default server settings
func GetDefaultServerSettings() *ServerSettings {
	return &ServerSettings{
		ServerName:        "Terminal Velocity Server",
		MOTD:              "Welcome to Terminal Velocity!",
		MaxPlayers:        100,
		TickRate:          20,
		StartingCredits:   10000,
		EconomyMultiplier: 1.0,
		TaxRate:           0.05,
		CombatDifficulty:  1.0,
		PirateFrequency:   0.2,
		PriceVolatility:   0.15,
		PvPEnabled:        true,
		PermadeathMode:    false,
		FriendlyFire:      false,
		MaxShipsPerPlayer: 5,
		MaxCargoSpace:     1000,
		MaxCredits:        1000000000,
		SessionTimeout:    15,
		AutosaveInterval:  30,
		CleanupInterval:   5,
		EnableEncounters:  true,
		EnableFactions:    true,
		EnableAchievements: true,
		EnableLeaderboards: true,
	}
}
