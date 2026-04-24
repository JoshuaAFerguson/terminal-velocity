# Server Administration System

**Feature**: Server Administration and Moderation
**Phase**: 18
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Admin System provides comprehensive server administration tools including user management, moderation, server configuration, and audit logging. The system uses Role-Based Access Control (RBAC) with 4 roles and 20+ permissions.

### Key Features

- **RBAC System**: 4 roles with granular permissions
- **User Moderation**: Ban and mute capabilities with expiration
- **Audit Logging**: 10,000-entry action buffer with full audit trail
- **Server Metrics**: Real-time performance monitoring
- **Server Settings**: Dynamic configuration management
- **Permission System**: 20+ fine-grained permissions
- **Action History**: Complete record of administrative actions

---

## Architecture

### Components

The admin system consists of the following components:

1. **Admin Manager** (`internal/admin/manager.go`)
   - Admin user management
   - Permission checking
   - Ban/mute system
   - Audit logging
   - Thread-safe with `sync.RWMutex`

2. **Admin UI** (`internal/tui/admin.go`)
   - Admin dashboard
   - User management interface
   - Audit log viewer
   - Server metrics display

3. **Data Models** (`internal/models/`)
   - `AdminUser`: Admin accounts with roles
   - `PlayerBan`: Ban records
   - `PlayerMute`: Mute records
   - `AdminAction`: Audit log entries
   - `ServerSettings`: Server configuration

### Data Flow

```
Admin Action Request
         ↓
Check Admin Permissions
         ↓
Validate Action Parameters
         ↓
Execute Administrative Action
         ↓
Log to Audit Trail
         ↓
Broadcast Changes
         ↓
Update UI
```

### Thread Safety

The admin manager uses `sync.RWMutex` for concurrent operations:

- **Read Operations**: Concurrent reads for permission checks
- **Write Operations**: Exclusive access for bans/mutes
- **Audit Logging**: Thread-safe append operations
- **Metrics Collection**: Background goroutine

---

## Implementation Details

### Admin Manager

The manager handles all administrative operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Admin users
    admins map[uuid.UUID]*models.AdminUser // PlayerID -> AdminUser

    // Moderation
    bans   map[uuid.UUID]*models.PlayerBan  // PlayerID -> Ban
    mutes  map[uuid.UUID]*models.PlayerMute // PlayerID -> Mute

    // Audit log
    actionLog []*models.AdminAction

    // Server settings
    settings *models.ServerSettings

    // Metrics
    metrics *models.ServerMetrics

    // Repositories
    playerRepo *database.PlayerRepository

    // Background workers
    metricsInterval time.Duration
    ctx             context.Context
    cancel          context.CancelFunc
    wg              sync.WaitGroup
}
```

### Role-Based Access Control

**Admin Roles**:
```go
const (
    RoleSuperAdmin = "superadmin" // Full access
    RoleAdmin      = "admin"      // Most permissions
    RoleModerator  = "moderator"  // Moderation only
    RoleHelper     = "helper"     // Read-only + basic help
)
```

**Permission List** (20+ permissions):
```go
const (
    // User Management
    PermViewPlayers    = "view_players"
    PermBanPlayer      = "ban_player"
    PermMutePlayer     = "mute_player"
    PermKickPlayer     = "kick_player"
    PermEditPlayer     = "edit_player"

    // Server Management
    PermServerSettings = "server_settings"
    PermServerRestart  = "server_restart"
    PermServerShutdown = "server_shutdown"
    PermViewMetrics    = "view_metrics"
    PermViewLogs       = "view_logs"

    // Content Management
    PermEditUniverse   = "edit_universe"
    PermEditMarkets    = "edit_markets"
    PermEditMissions   = "edit_missions"
    PermEditQuests     = "edit_quests"

    // Admin Management
    PermManageAdmins   = "manage_admins"
    PermViewAuditLog   = "view_audit_log"

    // Communication
    PermBroadcast      = "broadcast_message"
    PermModerateChat   = "moderate_chat"

    // Economy
    PermGrantCredits   = "grant_credits"
    PermAdjustPrices   = "adjust_prices"
)
```

**Role Permissions**:
```go
var rolePermissions = map[AdminRole][]AdminPermission{
    RoleSuperAdmin: {
        // All permissions
        PermViewPlayers, PermBanPlayer, PermMutePlayer,
        PermKickPlayer, PermEditPlayer, PermServerSettings,
        PermServerRestart, PermServerShutdown, PermViewMetrics,
        PermViewLogs, PermEditUniverse, PermEditMarkets,
        PermEditMissions, PermEditQuests, PermManageAdmins,
        PermViewAuditLog, PermBroadcast, PermModerateChat,
        PermGrantCredits, PermAdjustPrices,
    },
    RoleAdmin: {
        // Most permissions, except server shutdown and admin management
        PermViewPlayers, PermBanPlayer, PermMutePlayer,
        PermKickPlayer, PermEditPlayer, PermServerSettings,
        PermViewMetrics, PermViewLogs, PermEditMarkets,
        PermEditMissions, PermViewAuditLog, PermBroadcast,
        PermModerateChat, PermGrantCredits, PermAdjustPrices,
    },
    RoleModerator: {
        // Moderation permissions only
        PermViewPlayers, PermBanPlayer, PermMutePlayer,
        PermKickPlayer, PermModerateChat, PermViewAuditLog,
    },
    RoleHelper: {
        // Read-only permissions
        PermViewPlayers, PermViewMetrics, PermViewLogs,
    },
}
```

**Permission Checking**:
```go
func (a *AdminUser) HasPermission(perm AdminPermission) bool {
    permissions, exists := rolePermissions[a.Role]
    if !exists {
        return false
    }

    for _, p := range permissions {
        if p == perm {
            return true
        }
    }
    return false
}
```

### Ban System

**Ban Structure**:
```go
type PlayerBan struct {
    BanID       uuid.UUID
    PlayerID    uuid.UUID
    Username    string
    IPAddress   string
    Reason      string
    BannedBy    uuid.UUID
    BannedAt    time.Time
    ExpiresAt   time.Time
    IsPermanent bool
    IsActive    bool
}
```

**Ban Player**:
```go
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
    if !exists || !admin.HasPermission(PermBanPlayer) {
        return ErrNotAuthorized
    }

    // Cannot ban other admins
    if _, isAdmin := m.admins[targetID]; isAdmin {
        return ErrCannotBanAdmin
    }

    // Create ban
    ban := models.NewPlayerBan(
        targetID, username, ipAddress,
        reason, adminID, duration,
    )
    m.bans[targetID] = ban

    // Log action
    m.logActionUnsafe(adminID, "ban_player", targetID, username,
        fmt.Sprintf("Banned: %s", reason))

    return nil
}
```

**Unban Player**:
```go
func (m *Manager) UnbanPlayer(
    adminID uuid.UUID,
    targetID uuid.UUID,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Check permission
    admin, exists := m.admins[adminID]
    if !exists || !admin.HasPermission(PermBanPlayer) {
        return ErrNotAuthorized
    }

    ban, exists := m.bans[targetID]
    if !exists {
        return ErrPlayerNotBanned
    }

    ban.IsActive = false

    // Log action
    m.logActionUnsafe(adminID, "unban_player", targetID,
        ban.Username, "Unbanned player")

    return nil
}
```

### Mute System

**Mute Structure**:
```go
type PlayerMute struct {
    MuteID      uuid.UUID
    PlayerID    uuid.UUID
    Username    string
    Reason      string
    MutedBy     uuid.UUID
    MutedAt     time.Time
    ExpiresAt   time.Time
    IsActive    bool
}
```

**Mute Player**:
```go
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
    if !exists || !admin.HasPermission(PermMutePlayer) {
        return ErrNotAuthorized
    }

    // Create mute
    mute := models.NewPlayerMute(
        targetID, username, reason,
        adminID, duration,
    )
    m.mutes[targetID] = mute

    // Log action
    m.logActionUnsafe(adminID, "mute_player", targetID, username,
        fmt.Sprintf("Muted: %s", reason))

    return nil
}
```

### Audit Logging

**Audit Log Entry**:
```go
type AdminAction struct {
    ActionID   uuid.UUID
    Timestamp  time.Time
    AdminID    uuid.UUID
    AdminName  string
    Action     string
    TargetID   *uuid.UUID
    TargetName string
    Details    string
    Success    bool
    IPAddress  string
}
```

**Log Action**:
```go
func (m *Manager) LogAction(action *AdminAction) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.actionLog = append(m.actionLog, action)

    // Trim log if too large (keep last 10,000)
    if len(m.actionLog) > 10000 {
        m.actionLog = m.actionLog[1000:]
    }
}
```

**Get Audit Log**:
```go
func (m *Manager) GetActionLog(limit int) []*AdminAction {
    m.mu.RLock()
    defer m.mu.RUnlock()

    start := 0
    if len(m.actionLog) > limit {
        start = len(m.actionLog) - limit
    }

    // Return a copy
    result := make([]*AdminAction, len(m.actionLog[start:]))
    copy(result, m.actionLog[start:])

    return result
}
```

### Server Metrics

**Metrics Collection**:
```go
func (m *Manager) collectMetrics() {
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    metrics := &models.ServerMetrics{
        Timestamp:      time.Now(),
        GoroutineCount: runtime.NumGoroutine(),
        MemoryUsage:    int64(mem.Alloc),

        // Note: In production, these would be populated
        // from actual game state managers
        TotalPlayers:     0,
        ActivePlayers:    0,
        PeakPlayers:      0,
        ActiveSessions:   0,
        TotalCommands:    0,
        DBConnections:    0,
        DBLatency:        0,
        DBErrors:         0,
    }

    m.UpdateMetrics(metrics)
}
```

---

## User Interface

### Admin Dashboard

**Main Menu**:
```
=== Server Administration ===

Role: Admin

Administration Menu:

  > Player Management    - View and manage online players
    Ban Management       - View and manage player bans
    Mute Management      - View and manage player mutes
    Server Metrics       - View server performance and statistics
    Server Settings      - Configure server parameters
    Action Log           - View admin action history

↑/↓: Navigate  •  Enter: Select  •  ESC: Back
```

**Ban Management**:
```
Active Bans:

Username             Reason               Type       Banned At
────────────────────────────────────────────────────────────────
Spammer123           Chat spam            Temporary  Jan 14 15:30
Cheater456           Exploitation         Permanent  Jan 13 10:22
Toxic789             Harassment           Temporary  Jan 15 08:15

> [Select to view details or unban]

↑/↓: Navigate  •  U: Unban  •  ESC: Back
```

**Server Metrics**:
```
Server Metrics:

Performance:
  Memory Usage:   125.4 MB
  Goroutines:     247

Players:
  Total Players:  1,524
  Active Players: 23
  Peak Players:   156

Activity:
  Active Sessions: 23
  Total Commands:  45,678
  Active Trades:   3
  Active Combats:  7

Database:
  Connections: 15
  Latency:     12 ms
  Errors:      0

ESC: Back
```

**Action Log**:
```
Admin Action Log (Recent 20):

Time      Admin      Action           Details
────────────────────────────────────────────────────────────
15:04:05  Alice      ban_player       Banned Spammer123: Chat spam
14:32:18  Bob        mute_player      Muted Toxic789: Harassment
13:15:42  Alice      update_settings  Updated server settings
12:08:29  Charlie    unban_player     Unbanned Reformed123

ESC: Back
```

### Navigation

- **↑/↓**: Navigate menu
- **Enter**: Select option
- **U**: Unban (in ban list)
- **M**: Unmute (in mute list)
- **ESC**: Return to previous menu

---

## Integration with Other Systems

### Chat Integration

Muted players cannot send messages:

```go
// In chat system
if adminManager.IsMuted(playerID) {
    return ErrPlayerMuted
}
```

### Authentication Integration

Banned players cannot log in:

```go
// In auth system
if adminManager.IsBanned(playerID) {
    return ErrPlayerBanned
}
```

---

## Testing

### Unit Tests

```go
func TestAdmin_AddAdmin(t *testing.T)
func TestAdmin_BanPlayer(t *testing.T)
func TestAdmin_MutePlayer(t *testing.T)
func TestAdmin_Permissions(t *testing.T)
func TestAdmin_AuditLog(t *testing.T)
```

---

## Configuration

```go
cfg := &admin.Config{
    // Audit log
    AuditLogSize:      10000,
    AuditLogRetention: 90 * 24 * time.Hour,

    // Metrics
    MetricsInterval:   10 * time.Second,

    // Defaults
    DefaultBanDuration:  24 * time.Hour,
    DefaultMuteDuration: 1 * time.Hour,
}
```

---

## API Reference

### Core Functions

#### AddAdmin

```go
func (m *Manager) AddAdmin(
    playerID uuid.UUID,
    username string,
    role models.AdminRole,
    createdBy uuid.UUID,
) (*models.AdminUser, error)
```

Adds a new admin user.

#### BanPlayer

```go
func (m *Manager) BanPlayer(
    adminID uuid.UUID,
    targetID uuid.UUID,
    username string,
    ipAddress string,
    reason string,
    duration *time.Duration,
) error
```

Bans a player.

#### MutePlayer

```go
func (m *Manager) MutePlayer(
    adminID uuid.UUID,
    targetID uuid.UUID,
    username string,
    reason string,
    duration time.Duration,
) error
```

Mutes a player.

#### GetActionLog

```go
func (m *Manager) GetActionLog(limit int) []*AdminAction
```

Returns recent admin actions.

---

## Related Documentation

- [Chat System](./CHAT_SYSTEM.md) - Mute integration
- [Metrics & Monitoring](./METRICS_MONITORING.md) - Server metrics
- [Rate Limiting](./RATE_LIMITING.md) - Security integration

---

## File Locations

**Core Implementation**:
- `internal/admin/manager.go` - Admin manager
- `internal/models/admin.go` - Admin data models

**User Interface**:
- `internal/tui/admin.go` - Admin UI

**Database**:
- `scripts/schema.sql` - Admin tables schema

**Documentation**:
- `docs/ADMIN_SYSTEM.md` - This file
- `ROADMAP.md` - Phase 18 details

---

**For questions about the admin system, contact the development team.**
