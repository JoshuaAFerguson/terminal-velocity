# Player Presence System

**Feature**: Real-Time Player Presence Tracking
**Phase**: 15
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Player Presence system tracks online players in real-time, providing location information and activity status to enable multiplayer interactions. It serves as the foundation for features like player trading, PvP combat, and social gameplay.

### Key Features

- **Real-Time Tracking**: Track online players and their locations
- **Activity Status**: See what other players are doing
- **Location Broadcasting**: Share current system/planet location
- **Offline Timeout**: Automatic cleanup after 5 minutes of inactivity
- **Privacy Controls**: Players can hide their presence
- **Efficient Updates**: Optimized broadcast mechanism
- **Integration Ready**: Powers trading, PvP, and social features

---

## Architecture

### Components

The presence system consists of the following components:

1. **Presence Manager** (`internal/presence/manager.go`)
   - Tracks online players
   - Manages location updates
   - Handles heartbeats and timeouts
   - Thread-safe with `sync.RWMutex`

2. **Integration Points**:
   - Session manager integration
   - Player location updates
   - Chat system notifications
   - Trade/PvP system queries

3. **Data Models** (`internal/models/`)
   - `PlayerPresence`: Online status and location
   - `ActivityStatus`: Current player activity
   - `PresenceUpdate`: Change notifications

### Data Flow

```
Player Login
     ↓
Register Presence
     ↓
Start Heartbeat Timer
     ↓
Location Updates
     ↓
Broadcast to Interested Players
     ↓
[After 5 minutes inactivity]
Mark Offline
     ↓
Cleanup Presence Data
```

### Thread Safety

The presence manager uses `sync.RWMutex` for concurrent operations:

- **Read Operations**: Efficient concurrent reads for presence queries
- **Write Operations**: Exclusive access for updates
- **Heartbeat Processing**: Asynchronous update handling
- **Cleanup Worker**: Background goroutine for timeout processing

---

## Implementation Details

### Presence Manager

The manager tracks all online players:

```go
type Manager struct {
    mu sync.RWMutex

    // Online players
    presence       map[uuid.UUID]*PlayerPresence    // PlayerID -> Presence
    systemPlayers  map[uuid.UUID][]uuid.UUID        // SystemID -> PlayerIDs
    planetPlayers  map[uuid.UUID][]uuid.UUID        // PlanetID -> PlayerIDs

    // Configuration
    heartbeatInterval time.Duration
    offlineTimeout    time.Duration
    cleanupInterval   time.Duration

    // Background workers
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}
```

### PlayerPresence Structure

```go
type PlayerPresence struct {
    PlayerID    uuid.UUID
    Username    string
    Status      PresenceStatus

    // Location
    CurrentSystem *uuid.UUID
    CurrentPlanet *uuid.UUID
    IsDocked      bool

    // Activity
    Activity      ActivityType
    LastSeen      time.Time
    LastHeartbeat time.Time

    // Privacy
    Visible       bool

    // Ship info (if visible)
    ShipType      string
    ShipName      string
}
```

### Presence Registration

**On Player Login**:
```go
func (m *Manager) RegisterPlayer(
    playerID uuid.UUID,
    username string,
    initialLocation *uuid.UUID,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    presence := &PlayerPresence{
        PlayerID:      playerID,
        Username:      username,
        Status:        StatusOnline,
        CurrentSystem: initialLocation,
        LastSeen:      time.Now(),
        LastHeartbeat: time.Now(),
        Visible:       true,
    }

    m.presence[playerID] = presence
    m.updateLocationIndex(presence)

    log.Info("Player %s (%s) is now online", username, playerID)
    return nil
}
```

### Location Updates

**Efficient Location Tracking**:
```go
func (m *Manager) UpdateLocation(
    playerID uuid.UUID,
    systemID *uuid.UUID,
    planetID *uuid.UUID,
    isDocked bool,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    presence, exists := m.presence[playerID]
    if !exists {
        return ErrPlayerNotOnline
    }

    // Remove from old location indices
    m.removeFromLocationIndex(presence)

    // Update location
    presence.CurrentSystem = systemID
    presence.CurrentPlanet = planetID
    presence.IsDocked = isDocked
    presence.LastSeen = time.Now()

    // Add to new location indices
    m.updateLocationIndex(presence)

    // Broadcast to nearby players
    m.broadcastLocationUpdate(presence)

    return nil
}
```

### Heartbeat System

**Heartbeat Processing**:
```go
func (m *Manager) Heartbeat(playerID uuid.UUID) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    presence, exists := m.presence[playerID]
    if !exists {
        return ErrPlayerNotOnline
    }

    presence.LastHeartbeat = time.Now()
    presence.LastSeen = time.Now()

    return nil
}
```

**Timeout Detection**:
```go
func (m *Manager) cleanupWorker() {
    defer m.wg.Done()

    ticker := time.NewTicker(m.cleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-m.ctx.Done():
            return
        case <-ticker.C:
            m.cleanupInactivePlayers()
        }
    }
}

func (m *Manager) cleanupInactivePlayers() {
    m.mu.Lock()
    defer m.mu.Unlock()

    now := time.Now()
    timeout := now.Add(-m.offlineTimeout)

    for playerID, presence := range m.presence {
        if presence.LastHeartbeat.Before(timeout) {
            log.Info("Player %s timed out (last seen: %v)",
                presence.Username, presence.LastSeen)

            m.removePresence(playerID)
        }
    }
}
```

### Location Queries

**Get Players in System**:
```go
func (m *Manager) GetPlayersInSystem(systemID uuid.UUID) []*PlayerPresence {
    m.mu.RLock()
    defer m.mu.RUnlock()

    playerIDs, exists := m.systemPlayers[systemID]
    if !exists {
        return []*PlayerPresence{}
    }

    players := make([]*PlayerPresence, 0, len(playerIDs))
    for _, playerID := range playerIDs {
        if presence, ok := m.presence[playerID]; ok {
            if presence.Visible {
                players = append(players, presence)
            }
        }
    }

    return players
}
```

**Get Nearby Players**:
```go
func (m *Manager) GetNearbyPlayers(
    playerID uuid.UUID,
    maxDistance int,
) []*PlayerPresence {
    m.mu.RLock()
    defer m.mu.RUnlock()

    presence, exists := m.presence[playerID]
    if !exists || presence.CurrentSystem == nil {
        return []*PlayerPresence{}
    }

    // Get all players in same system
    nearby := m.GetPlayersInSystem(*presence.CurrentSystem)

    // If on planet, filter to same planet
    if presence.IsDocked && presence.CurrentPlanet != nil {
        planetPlayers := make([]*PlayerPresence, 0)
        for _, p := range nearby {
            if p.IsDocked && p.CurrentPlanet != nil &&
                *p.CurrentPlanet == *presence.CurrentPlanet {
                planetPlayers = append(planetPlayers, p)
            }
        }
        return planetPlayers
    }

    return nearby
}
```

### Activity Status

**Activity Types**:
```go
const (
    ActivityIdle        = "idle"
    ActivityTrading     = "trading"
    ActivityCombat      = "in_combat"
    ActivityMission     = "on_mission"
    ActivityExploring   = "exploring"
    ActivityDocked      = "docked"
    ActivityInMenu      = "in_menu"
)
```

**Update Activity**:
```go
func (m *Manager) UpdateActivity(
    playerID uuid.UUID,
    activity ActivityType,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    presence, exists := m.presence[playerID]
    if !exists {
        return ErrPlayerNotOnline
    }

    presence.Activity = activity
    presence.LastSeen = time.Now()

    return nil
}
```

---

## User Interface

### Players Screen Integration

The presence system powers the Players screen:

```
=== Online Players (23) ===

System: Alpha Centauri (5 players)
────────────────────────────────────────
  Alice       [Trading]    Corvette
  Bob         [Docked]     Freighter
  Charlie     [Combat]     Destroyer

System: Tau Ceti (3 players)
────────────────────────────────────────
  David       [Exploring]  Scout
  Eve         [Mission]    Frigate

[View Profile] [Trade] [Challenge] [Message]
```

### Presence Indicators

```
Player Status Indicators:
  ● Online (active < 1 min)
  ○ Idle   (active 1-5 min)
  ◌ Away   (visible but inactive)
  [Hidden] (privacy enabled)
```

---

## Integration with Other Systems

### Session Management

Presence integrates with session lifecycle:

**Session Events**:
```go
// On session start
presenceManager.RegisterPlayer(playerID, username, startSystem)

// On session heartbeat
presenceManager.Heartbeat(playerID)

// On session end
presenceManager.UnregisterPlayer(playerID)
```

### Trading System

Trading requires online presence:

```go
func (t *TradeManager) CreateTradeOffer(...) error {
    // Check if recipient is online
    if !presenceManager.IsPlayerOnline(recipientID) {
        return ErrPlayerOffline
    }

    // Create offer...
}
```

### PvP System

PvP challenges use presence data:

```go
func (p *PvPManager) CreateChallenge(...) error {
    // Check presence
    presence := presenceManager.GetPresence(targetID)
    if presence == nil {
        return ErrPlayerOffline
    }

    // Check if in combat
    if presence.Activity == ActivityCombat {
        return ErrPlayerInCombat
    }

    // Create challenge...
}
```

### Chat System

Presence powers chat features:

- Online user lists
- User availability status
- Direct message delivery
- Typing indicators

---

## Testing

### Unit Tests

```go
func TestPresence_Registration(t *testing.T)
func TestPresence_LocationUpdate(t *testing.T)
func TestPresence_Heartbeat(t *testing.T)
func TestPresence_Timeout(t *testing.T)
func TestPresence_LocationQueries(t *testing.T)
func TestPresence_Privacy(t *testing.T)
func TestPresence_Concurrency(t *testing.T)
```

### Integration Tests

1. **Login/Logout Flow**:
   - Register presence
   - Update locations
   - Unregister cleanly

2. **Timeout Flow**:
   - Stop heartbeats
   - Wait for timeout
   - Verify cleanup

3. **Multi-Player**:
   - Multiple players
   - Same system
   - Location queries

---

## Configuration

```go
cfg := &presence.Config{
    // Heartbeat settings
    HeartbeatInterval: 30 * time.Second,
    OfflineTimeout:    5 * time.Minute,

    // Cleanup settings
    CleanupInterval:   1 * time.Minute,

    // Privacy defaults
    DefaultVisible:    true,
    AllowHidden:       true,

    // Performance
    MaxPresenceCache:  10000,
}
```

---

## Troubleshooting

### Common Issues

**Problem**: Player shows offline but is connected
**Solutions**:
- Check heartbeat mechanism
- Verify session integration
- Review timeout configuration
- Check cleanup worker status

**Problem**: Location not updating
**Solutions**:
- Verify UpdateLocation calls
- Check mutex locks
- Review location index
- Test broadcast mechanism

**Problem**: High memory usage
**Solutions**:
- Check cleanup worker
- Verify timeout enforcement
- Review presence cache size
- Monitor goroutine leaks

---

## API Reference

### Core Functions

#### RegisterPlayer

```go
func (m *Manager) RegisterPlayer(
    playerID uuid.UUID,
    username string,
    initialLocation *uuid.UUID,
) error
```

Registers a player as online.

#### UnregisterPlayer

```go
func (m *Manager) UnregisterPlayer(playerID uuid.UUID) error
```

Marks a player as offline.

#### UpdateLocation

```go
func (m *Manager) UpdateLocation(
    playerID uuid.UUID,
    systemID *uuid.UUID,
    planetID *uuid.UUID,
    isDocked bool,
) error
```

Updates player's current location.

#### Heartbeat

```go
func (m *Manager) Heartbeat(playerID uuid.UUID) error
```

Records player activity to prevent timeout.

#### GetPresence

```go
func (m *Manager) GetPresence(
    playerID uuid.UUID,
) (*PlayerPresence, error)
```

Returns current presence data for a player.

#### IsPlayerOnline

```go
func (m *Manager) IsPlayerOnline(playerID uuid.UUID) bool
```

Checks if a player is currently online.

---

## Related Documentation

- [Player Trading](./PLAYER_TRADING.md) - Trading integration
- [PvP Combat](./PVP_COMBAT.md) - Combat integration
- [Chat System](./CHAT_SYSTEM.md) - Chat integration
- [Session Management](./SESSION.md) - Session lifecycle

---

## File Locations

**Core Implementation**:
- `internal/presence/manager.go` - Presence manager

**Data Models**:
- `internal/models/presence.go` - Presence data structures

**Tests**:
- `internal/presence/manager_test.go` - Unit tests

**Documentation**:
- `docs/PLAYER_PRESENCE.md` - This file
- `ROADMAP.md` - Phase 15 details

---

**For questions about the presence system, see the troubleshooting section or contact the development team.**
