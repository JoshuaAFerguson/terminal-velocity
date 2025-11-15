# Settings System

**Feature**: Player Settings and Preferences
**Phase**: 17
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Settings system provides comprehensive player customization options across display, audio, gameplay, controls, privacy, and notification categories. Settings are persisted to the database and loaded automatically on login.

### Key Features

- **6 Setting Categories**: Display, Audio, Gameplay, Controls, Privacy, Notifications
- **5 Color Schemes**: Default, Dark, Light, High Contrast, Colorblind
- **JSON Persistence**: Settings stored in database as JSON
- **Real-Time Updates**: Changes apply immediately
- **Reset Functionality**: Restore defaults per category or all
- **Privacy Controls**: Fine-grained privacy settings
- **Cross-Session Persistence**: Settings saved across logins

---

## Architecture

### Components

The settings system consists of the following components:

1. **Settings Manager** (`internal/settings/manager.go`)
   - Settings storage and retrieval
   - Default values management
   - Update validation
   - Thread-safe with `sync.RWMutex`

2. **Settings UI** (`internal/tui/settings.go`)
   - Category navigation
   - Setting editors
   - Preview and validation
   - Reset controls

3. **Data Models** (`internal/models/`)
   - `Settings`: Main settings container
   - Category-specific structs
   - Validation rules

### Data Flow

```
Player Changes Setting
         ↓
Validate New Value
         ↓
Update Settings Object
         ↓
Save to Database (JSON)
         ↓
Apply Changes to UI
         ↓
Update Relevant Systems
```

### Thread Safety

The settings manager uses `sync.RWMutex` for concurrent operations:

- **Read Operations**: Concurrent reads for settings access
- **Write Operations**: Exclusive access for updates
- **Persistence**: Atomic save operations

---

## Implementation Details

### Settings Manager

The manager handles all settings operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Player settings
    settings map[uuid.UUID]*models.Settings // PlayerID -> Settings

    // Default settings
    defaults *models.Settings

    // Repositories
    playerRepo *database.PlayerRepository
}
```

### Settings Structure

**Main Settings Container**:
```go
type Settings struct {
    PlayerID      uuid.UUID
    Display       DisplaySettings
    Audio         AudioSettings
    Gameplay      GameplaySettings
    Controls      ControlSettings
    Privacy       PrivacySettings
    Notifications NotificationSettings

    LastUpdated   time.Time
}
```

### Display Settings

**Display Configuration**:
```go
type DisplaySettings struct {
    ColorScheme      string  // default, dark, light, high_contrast, colorblind
    ShowAnimations   bool
    CompactMode      bool
    ShowTutorialTips bool
    ShowIcons        bool
}
```

**Color Schemes**:

1. **Default**: Standard green-on-black terminal
2. **Dark**: Darker background, softer colors
3. **Light**: Light background for daylight use
4. **High Contrast**: Maximum readability
5. **Colorblind**: Colorblind-friendly palette

**Implementation**:
```go
func (m *Manager) ApplyColorScheme(
    playerID uuid.UUID,
    scheme string,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    settings := m.settings[playerID]
    settings.Display.ColorScheme = scheme
    settings.LastUpdated = time.Now()

    // Persist to database
    return m.saveSettings(playerID, settings)
}
```

### Audio Settings

**Audio Configuration**:
```go
type AudioSettings struct {
    Enabled       bool
    SoundEffects  bool
    Music         bool
    Notifications bool
    Volume        int  // 0-100
}
```

**Note**: Audio playback not yet implemented, but settings structure prepared for future use.

### Gameplay Settings

**Gameplay Configuration**:
```go
type GameplaySettings struct {
    AutoSave                bool
    ConfirmDangerousActions bool
    ShowDamageNumbers       bool
    AutoPilot               bool
    PauseOnEncounter        bool
    FastTravel              bool
    TutorialMode            bool
    DifficultyLevel         string // easy, normal, hard, expert
    PermadeathMode          bool
}
```

**Difficulty Levels**:
- **Easy**: Reduced combat difficulty, higher rewards
- **Normal**: Balanced gameplay
- **Hard**: Tougher enemies, normal rewards
- **Expert**: Maximum challenge, bonus rewards

### Control Settings

**Control Configuration**:
```go
type ControlSettings struct {
    // Navigation
    MoveUp    string
    MoveDown  string
    MoveLeft  string
    MoveRight string

    // Actions
    Confirm   string
    Cancel    string
    Back      string
    Help      string

    // Combat
    Attack    string
    Defend    string
    Flee      string
}
```

**Default Key Bindings**:
```go
func getDefaultControls() ControlSettings {
    return ControlSettings{
        MoveUp:    "↑ / k",
        MoveDown:  "↓ / j",
        MoveLeft:  "← / h",
        MoveRight: "→ / l",
        Confirm:   "Enter / Space",
        Cancel:    "ESC",
        Back:      "Backspace / q",
        Help:      "? / F1",
        Attack:    "a",
        Defend:    "d",
        Flee:      "f",
    }
}
```

### Privacy Settings

**Privacy Configuration**:
```go
type PrivacySettings struct {
    ShowOnline           bool
    ShowLocation         bool
    ShowShip             bool
    AllowTradeRequests   bool
    AllowPvPChallenges   bool
    AllowPartyInvites    bool
    BlockList            []uuid.UUID
    FriendsList          []uuid.UUID
}
```

**Privacy Impact**:

| Setting | Affects |
|---------|---------|
| ShowOnline | Player presence system |
| ShowLocation | System/planet visibility |
| ShowShip | Ship details in player list |
| AllowTradeRequests | Trade offer reception |
| AllowPvPChallenges | PvP duel challenges |
| AllowPartyInvites | Party system invitations |

### Notification Settings

**Notification Configuration**:
```go
type NotificationSettings struct {
    ShowAchievements   bool
    ShowLevelUp        bool
    ShowTradeComplete  bool
    ShowCombatLog      bool
    ShowPlayerJoined   bool
    ShowNewsUpdates    bool
    ShowEncounters     bool
    ShowSystemMessages bool
    ChatNotifications  bool
}
```

### Settings Persistence

**Save to Database**:
```go
func (m *Manager) saveSettings(
    playerID uuid.UUID,
    settings *Settings,
) error {
    // Serialize to JSON
    jsonData, err := json.Marshal(settings)
    if err != nil {
        return err
    }

    // Store in database
    return m.playerRepo.UpdateSettings(playerID, jsonData)
}
```

**Load from Database**:
```go
func (m *Manager) GetSettings(
    playerID uuid.UUID,
) (*Settings, error) {
    m.mu.RLock()
    cached := m.settings[playerID]
    m.mu.RUnlock()

    if cached != nil {
        return cached, nil
    }

    // Load from database
    jsonData, err := m.playerRepo.GetSettings(playerID)
    if err != nil {
        // Return defaults if not found
        return m.getDefaultSettings(playerID), nil
    }

    // Deserialize from JSON
    settings := &Settings{}
    if err := json.Unmarshal(jsonData, settings); err != nil {
        return nil, err
    }

    // Cache settings
    m.mu.Lock()
    m.settings[playerID] = settings
    m.mu.Unlock()

    return settings, nil
}
```

### Update Settings

**Generic Update Function**:
```go
func (m *Manager) UpdateSettings(
    playerID uuid.UUID,
    updateFunc func(*Settings),
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    settings := m.settings[playerID]
    if settings == nil {
        settings = m.getDefaultSettings(playerID)
    }

    // Apply update
    updateFunc(settings)
    settings.LastUpdated = time.Now()

    // Save to database
    return m.saveSettings(playerID, settings)
}
```

**Example Usage**:
```go
// Update display settings
m.UpdateSettings(playerID, func(s *Settings) {
    s.Display.ColorScheme = "dark"
    s.Display.CompactMode = true
})

// Update privacy settings
m.UpdateSettings(playerID, func(s *Settings) {
    s.Privacy.ShowOnline = false
    s.Privacy.AllowPvPChallenges = false
})
```

### Reset Functionality

**Reset to Defaults**:
```go
func (m *Manager) ResetToDefaults(playerID uuid.UUID) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    defaults := m.getDefaultSettings(playerID)
    m.settings[playerID] = defaults

    return m.saveSettings(playerID, defaults)
}
```

**Reset Category**:
```go
func (m *Manager) ResetCategory(
    playerID uuid.UUID,
    category string,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    settings := m.settings[playerID]
    defaults := m.getDefaultSettings(playerID)

    switch category {
    case "display":
        settings.Display = defaults.Display
    case "audio":
        settings.Audio = defaults.Audio
    case "gameplay":
        settings.Gameplay = defaults.Gameplay
    case "controls":
        settings.Controls = defaults.Controls
    case "privacy":
        settings.Privacy = defaults.Privacy
    case "notifications":
        settings.Notifications = defaults.Notifications
    }

    settings.LastUpdated = time.Now()
    return m.saveSettings(playerID, settings)
}
```

---

## User Interface

### Settings Screen

**Main Menu**:
```
=== Settings ===

Select a category:

> Display       - Visual appearance and UI options
  Audio         - Sound effects and music (not yet implemented)
  Gameplay      - Game behavior and difficulty
  Controls      - Keybindings and input settings
  Privacy       - Visibility and social settings
  Notifications - Alert and message preferences

↑/↓: Select  •  Enter: Open  •  ESC: Back
```

**Display Settings**:
```
Display Settings:

  > Color Scheme           default
    Show Animations        ON
    Compact Mode           OFF
    Show Tutorial Tips     ON
    Show Icons             ON

Color Schemes: default, dark, light, high_contrast, colorblind

↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back
```

**Gameplay Settings**:
```
Gameplay Settings:

  > Auto-Save                       ON
    Confirm Dangerous Actions       ON
    Show Damage Numbers             ON
    Auto-Pilot Hints                OFF
    Pause on Encounter              OFF
    Fast Travel                     OFF
    Tutorial Mode                   ON
    Difficulty Level                normal
    Permadeath Mode                 OFF

Difficulty Levels: easy, normal, hard, expert

↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back
```

**Privacy Settings**:
```
Privacy Settings:

  > Show Online Status        ON
    Show Location             ON
    Show Ship Info            ON
    Allow Trade Requests      ON
    Allow PvP Challenges      ON
    Allow Party Invites       ON

Blocked Players: 0
Friends: 5

↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back
```

### Navigation

- **↑/↓**: Navigate settings
- **Enter/Space**: Toggle or edit setting
- **R**: Reset current category to defaults
- **ESC**: Return to previous menu

---

## Integration with Other Systems

### UI Integration

Settings affect UI appearance:

```go
// Apply color scheme
switch settings.Display.ColorScheme {
case "dark":
    applyDarkTheme()
case "light":
    applyLightTheme()
// ...
}

// Apply compact mode
if settings.Display.CompactMode {
    setLineSpacing(1)
} else {
    setLineSpacing(2)
}
```

### Privacy Integration

Privacy settings control feature availability:

```go
// In trade system
if !recipientSettings.Privacy.AllowTradeRequests {
    return ErrTradeRequestsDisabled
}

// In PvP system
if !targetSettings.Privacy.AllowPvPChallenges {
    return ErrPvPChallengesDisabled
}

// In presence system
if !playerSettings.Privacy.ShowOnline {
    hidePlayerPresence(playerID)
}
```

### Notification Integration

Notification settings filter messages:

```go
// In notification system
if settings.Notifications.ShowAchievements {
    displayAchievementNotification(achievement)
}

if settings.Notifications.ChatNotifications {
    displayChatMessage(message)
}
```

---

## Testing

### Unit Tests

```go
func TestSettings_Load(t *testing.T)
func TestSettings_Save(t *testing.T)
func TestSettings_Update(t *testing.T)
func TestSettings_Reset(t *testing.T)
func TestSettings_Defaults(t *testing.T)
func TestSettings_Persistence(t *testing.T)
```

---

## Configuration

```go
cfg := &settings.Config{
    // Persistence
    AutoSave:          true,
    SaveInterval:      5 * time.Second,

    // Defaults
    DefaultColorScheme:     "default",
    DefaultDifficulty:      "normal",
    DefaultAutoSave:        true,
    DefaultTutorialMode:    true,

    // Validation
    ValidateOnLoad:         true,
    MigrateOldSettings:     true,
}
```

---

## API Reference

### Core Functions

#### GetSettings

```go
func (m *Manager) GetSettings(
    playerID uuid.UUID,
) (*Settings, error)
```

Retrieves player settings from cache or database.

#### UpdateSettings

```go
func (m *Manager) UpdateSettings(
    playerID uuid.UUID,
    updateFunc func(*Settings),
) error
```

Updates settings using a callback function.

#### ResetToDefaults

```go
func (m *Manager) ResetToDefaults(
    playerID uuid.UUID,
) error
```

Resets all settings to default values.

#### ResetCategory

```go
func (m *Manager) ResetCategory(
    playerID uuid.UUID,
    category string,
) error
```

Resets a specific category to defaults.

---

## Related Documentation

- [Player Presence](./PLAYER_PRESENCE.md) - Privacy settings integration
- [Player Trading](./PLAYER_TRADING.md) - Trade request settings
- [PvP Combat](./PVP_COMBAT.md) - PvP challenge settings
- [Tutorial System](./TUTORIAL_SYSTEM.md) - Tutorial mode setting

---

## File Locations

**Core Implementation**:
- `internal/settings/manager.go` - Settings manager
- `internal/models/settings.go` - Settings data structures

**User Interface**:
- `internal/tui/settings.go` - Settings UI

**Database**:
- `scripts/schema.sql` - Settings storage schema

**Documentation**:
- `docs/SETTINGS_SYSTEM.md` - This file
- `ROADMAP.md` - Phase 17 details

---

**For questions about the settings system, contact the development team.**
