# System & Core Screens

This document covers system-level and core UI screens in Terminal Velocity.

## Overview

**Screens**: 7
- Main Menu Screen
- Game View (Wrapper)
- Settings Screen
- Help Screen
- Tutorial Screen
- Admin Screen
- (Space View & Landing covered in 02-NAVIGATION.md)

**Purpose**: Handle system-level operations including main menu, settings configuration, help documentation, tutorial system, and server administration.

**Source Files**:
- `internal/tui/main_menu.go` - Main menu interface
- `internal/tui/game.go` - Game view wrapper
- `internal/tui/settings.go` - Player settings
- `internal/tui/help.go` - Help system
- `internal/tui/tutorial.go` - Tutorial and onboarding
- `internal/tui/admin.go` - Server administration

---

## Main Menu Screen

### Source File
`internal/tui/main_menu.go`

### Purpose
Primary navigation hub accessed via ESC key from any screen.

### Menu Options

- **Resume** - Return to game
- **Ship Status** - View ship info
- **Character** - Player stats and info
- **Inventory** - Equipment and items
- **Map** - Galactic map
- **Missions** - Active missions
- **Quests** - Story quests
- **Social** - Chat, players, factions
- **Settings** - Game configuration
- **Help** - Documentation
- **Logout** - Exit to login screen
- **Quit** - Close game

### Key Features
- Quick navigation to any screen
- Current location displayed
- Credits and ship status shown
- Notification indicators
- Keyboard shortcuts for common actions

---

## Settings Screen

### Source File
`internal/tui/settings.go`

### Purpose
Configure player preferences and game settings.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ SETTINGS                                                    52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ [Display ▼]  [Gameplay]  [Audio]  [Controls]  [Privacy]  [Account]  ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ DISPLAY SETTINGS                                                     ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Color Scheme:                                                       ┃  ┃
┃  ┃  ◉ Classic Green       ○ Blue Plasma      ○ Amber Terminal          ┃  ┃
┃  ┃  ○ White on Black      ○ Cyberpunk Neon                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  UI Scale:  [────●──────] 100%                                       ┃  ┃
┃  ┃  Animation Speed:  [───────●───] Normal                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ☑ Show FPS Counter                                                 ┃  ┃
┃  ┃  ☑ Show Combat Damage Numbers                                       ┃  ┃
┃  ┃  ☑ Animate Space Objects                                            ┃  ┃
┃  ┃  ☐ Reduce Visual Effects (Performance Mode)                         ┃  ┃
┃  ┃  ☑ Show Player Names in Space                                       ┃  ┃
┃  ┃  ☐ Show Chat Timestamps                                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  HUD Opacity:  [──────────●] 100%                                    ┃  ┃
┃  ┃  Chat Window Position:  ◉ Bottom  ○ Top  ○ Floating                ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃                   [ Save Settings ]  [ Reset to Default ]            ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Switch Category  [Space] Toggle Option  [S]ave  [R]eset  [ESC] Back  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Setting Categories (6 total)

1. **Display**
   - Color schemes (5 options)
   - UI scale
   - Animation speed
   - Visual effects toggles
   - HUD customization

2. **Gameplay**
   - Auto-save settings
   - Auto-pilot options
   - Combat assistance
   - Trade notifications
   - Mission reminders

3. **Audio**
   - Sound effects volume
   - Music volume
   - Voice volume
   - Mute all

4. **Controls**
   - Key bindings
   - Mouse sensitivity
   - Gamepad support
   - Custom shortcuts

5. **Privacy**
   - Online status visibility
   - Location sharing
   - Trade request permissions
   - Block list

6. **Account**
   - Change password
   - Email preferences
   - Two-factor authentication
   - Account deletion

### Color Schemes (5 options)

1. **Classic Green** - Retro terminal green on black
2. **Blue Plasma** - Electric blue theme
3. **Amber Terminal** - Warm amber/orange
4. **White on Black** - High contrast monochrome
5. **Cyberpunk Neon** - Vibrant neon colors

### Settings Persistence
- Stored in database per player
- JSON format for flexibility
- Cloud sync across sessions
- Export/import settings

---

## Help Screen

### Source File
`internal/tui/help.go`

### Purpose
Context-sensitive help and documentation system.

### Help Categories

- **Getting Started** - Basic controls and concepts
- **Navigation** - Flying and jumping
- **Trading** - Economy and markets
- **Combat** - Fighting tactics
- **Ships** - Ship types and upgrades
- **Missions** - Mission system
- **Quests** - Story quests
- **Multiplayer** - Social features
- **Advanced** - Pro tips and tricks

### Features
- Search help topics
- Context-sensitive (shows relevant help for current screen)
- Hyperlinked topics
- Example screenshots (ASCII)
- Keyboard shortcut reference
- FAQ section

---

## Tutorial Screen

### Source File
`internal/tui/tutorial.go`

### Purpose
Interactive onboarding system for new players.

### Tutorial Structure

**7 Categories, 20+ Steps**:

1. **Basic Controls** (3 steps)
   - Movement and navigation
   - Menus and screens
   - Keyboard shortcuts

2. **Space Flight** (3 steps)
   - Takeoff and landing
   - Hyperjumps
   - Docking

3. **Trading** (4 steps)
   - Buying commodities
   - Selling for profit
   - Reading markets
   - Trade routes

4. **Combat** (4 steps)
   - Weapons and shields
   - Targeting enemies
   - Combat tactics
   - Fleeing/retreating

5. **Ships** (3 steps)
   - Ship upgrades
   - Outfitting
   - Ship management

6. **Missions** (2 steps)
   - Accepting missions
   - Completing objectives

7. **Multiplayer** (2 steps)
   - Chat and social
   - Factions and PvP

### Features
- Step-by-step guidance
- Interactive prompts
- Skip tutorial option
- Resume at any step
- Rewards for completion
- Hints and tips

---

## Admin Screen

### Source File
`internal/tui/admin.go`

### Purpose
Server administration interface with RBAC (Role-Based Access Control).

### Admin Features

**User Management**:
- View all players
- Ban/unban players (with expiration)
- Mute/unmute players (with expiration)
- Kick online players
- View player details
- Reset passwords

**Server Control**:
- Server statistics
- Player count and activity
- Database health
- Cache statistics
- Background workers status

**Moderation**:
- Chat logs
- Report review
- Abuse detection
- IP ban management

**Content Management**:
- Create news articles
- Spawn events
- Adjust economy (prices, stock)
- Create quests

**Audit Log**:
- All admin actions logged (10,000 entry buffer)
- Who did what, when
- IP tracking
- Action rollback

### RBAC Roles (4 levels)

1. **Owner** - Full server access
2. **Admin** - Most administrative functions
3. **Moderator** - User management, moderation
4. **Helper** - View-only, chat moderation

### Permissions (20+ granular)

- PLAYER_BAN
- PLAYER_MUTE
- PLAYER_KICK
- VIEW_STATS
- EDIT_ECONOMY
- CREATE_EVENTS
- VIEW_LOGS
- DELETE_MESSAGES
- etc.

### Security
- All actions require authentication
- Permission checks on every action
- Audit trail for accountability
- IP restrictions (optional)
- Two-factor authentication for admins

---

## Game View Wrapper

### Source File
`internal/tui/game.go`

### Purpose
High-level game state container and screen router.

### Responsibilities
- Track current game state (in space, docked, combat, menu)
- Route to appropriate screen based on state
- Handle screen transitions
- Manage autosave timer (30 seconds)
- Coordinate with session manager
- Handle disconnect/reconnect

### Game States
- **InSpace** - Flying in space
- **Docked** - At planet/station
- **InCombat** - Active combat
- **InMenu** - Main menu
- **Trading** - In trading screen
- **InDialogue** - NPC conversation
- **InTransit** - Hyperjump animation

---

## Implementation Notes

### Settings Storage

```go
type PlayerSettings struct {
    Display   *DisplaySettings   `json:"display"`
    Gameplay  *GameplaySettings  `json:"gameplay"`
    Audio     *AudioSettings     `json:"audio"`
    Controls  *ControlSettings   `json:"controls"`
    Privacy   *PrivacySettings   `json:"privacy"`
}

func (ps *PlayerSettings) Save(playerID string) error {
    json, err := json.Marshal(ps)
    if err != nil {
        return err
    }

    return database.UpdatePlayerSettings(playerID, json)
}
```

### Tutorial Progress

```go
type TutorialProgress struct {
    CurrentCategory int
    CurrentStep     int
    CompletedSteps  map[string]bool
    Skipped         bool
}

func (tp *TutorialProgress) MarkComplete(stepID string) {
    tp.CompletedSteps[stepID] = true
    tp.AdvanceStep()

    if tp.AllComplete() {
        AwardTutorialRewards()
    }
}
```

### Admin Audit Log

```go
type AuditEntry struct {
    Timestamp   time.Time
    AdminID     string
    AdminName   string
    Action      string
    TargetID    string
    TargetName  string
    Details     string
    IPAddress   string
}

func LogAdminAction(admin *Player, action string, target *Player, details string) {
    entry := &AuditEntry{
        Timestamp:  time.Now(),
        AdminID:    admin.ID,
        AdminName:  admin.Username,
        Action:     action,
        TargetID:   target.ID,
        TargetName: target.Username,
        Details:    details,
        IPAddress:  admin.IPAddress,
    }

    auditLog.Append(entry)

    if auditLog.Size() > 10000 {
        auditLog.Persist()
        auditLog.Clear()
    }
}
```

### Testing
- Settings save/load tests
- Tutorial progress tests
- Admin permission tests
- Audit log tests
- Screen routing tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
