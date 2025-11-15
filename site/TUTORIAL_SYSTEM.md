# Tutorial System

**Feature**: Tutorial and Onboarding System
**Phase**: 19
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Tutorial System provides comprehensive onboarding for new players with 7 tutorial categories and 20+ steps. The system features context-sensitive help, progress tracking, and an interactive overlay that guides players through game mechanics.

### Key Features

- **7 Tutorial Categories**: Basics, Trading, Navigation, Ships, Combat, Missions, Multiplayer
- **20+ Tutorial Steps**: Comprehensive coverage of all game features
- **Context-Sensitive Help**: Tutorials appear when relevant
- **Progress Tracking**: Track completion across categories
- **Interactive Overlay**: Non-intrusive tutorial display
- **Skip Functionality**: Players can skip or disable tutorials
- **Hint System**: Progressive hints for each step
- **Tutorial Triggers**: Auto-start based on player actions

---

## Architecture

### Components

1. **Tutorial Manager** (`internal/tutorial/manager.go`)
   - Tutorial registration and tracking
   - Progress management
   - Trigger handling
   - Thread-safe with `sync.RWMutex`

2. **Tutorial UI** (`internal/tui/tutorial.go`)
   - Overlay rendering
   - Step progression
   - Hint display
   - Tutorial list view

3. **Data Models** (`internal/models/`)
   - `Tutorial`: Tutorial definition
   - `TutorialStep`: Individual tutorial steps
   - `TutorialProgress`: Player progress tracking

### Data Flow

```
Trigger Event Occurs
         ↓
Check Tutorial Progress
         ↓
[If Not Completed]
Display Tutorial Overlay
         ↓
Player Completes Objective
         ↓
Mark Step Complete
         ↓
Update Progress
         ↓
Advance to Next Step
```

---

## Implementation Details

### Tutorial Manager

```go
type Manager struct {
    mu        sync.RWMutex
    tutorials map[string]*models.Tutorial            // tutorialID -> Tutorial
    progress  map[uuid.UUID]*models.TutorialProgress // playerID -> Progress
    triggers  map[models.TutorialTrigger][]string    // trigger -> tutorialIDs
}
```

### Tutorial Categories

**7 Categories**:

1. **Basics** (3 steps)
   - Welcome and navigation
   - Credits and currency
   - Menu navigation

2. **Trading** (3 steps)
   - Market basics
   - Buying commodities
   - Understanding prices

3. **Navigation** (2 steps)
   - Star map usage
   - Jumping between systems

4. **Ships** (3 steps)
   - Shipyard overview
   - Ship outfitting
   - Cargo management

5. **Combat** (3 steps)
   - Random encounters
   - Combat actions
   - Combat strategy

6. **Missions** (3 steps)
   - Mission board
   - Accepting missions
   - Completing missions

7. **Multiplayer** (3 steps)
   - Player presence
   - Chat system
   - Factions

### Tutorial Structure

```go
type Tutorial struct {
    ID           string
    Title        string
    Description  string
    Category     TutorialCategory
    Steps        []*TutorialStep
    Prerequisites []string  // Required tutorial IDs
    IsOptional   bool
    OrderIndex   int
}

type TutorialStep struct {
    ID          string
    Title       string
    Description string
    Screen      string    // Which screen triggers this
    Objective   string
    Hints       []string  // Progressive hints
    OrderIndex  int
    Completed   bool
}
```

### Tutorial Initialization

**Default Tutorials**:
```go
func (m *Manager) initializeDefaultTutorials() {
    // Basics Tutorial
    basicsTutorial := models.NewTutorial(
        "tutorial_basics",
        "Welcome to Terminal Velocity",
        "Learn the basics of navigating and playing the game",
        models.TutorialBasics,
    )

    basicsTutorial.AddStep(&models.TutorialStep{
        ID:          "basics_1_welcome",
        Title:       "Welcome, Commander!",
        Description: "Welcome to Terminal Velocity...",
        Screen:      "main_menu",
        Objective:   "Read the welcome message and press Enter to continue",
        Hints:       []string{
            "Press Enter or Space to continue",
            "You can skip tutorials at any time with 'S'",
        },
        OrderIndex:  1,
    })

    m.RegisterTutorial(basicsTutorial)
    m.AddTrigger(models.TriggerFirstLogin, "tutorial_basics")
}
```

### Trigger System

**Trigger Types**:
```go
const (
    TriggerFirstLogin     = "first_login"
    TriggerScreenEnter    = "screen_enter"
    TriggerFirstTrade     = "first_trade"
    TriggerFirstCombat    = "first_combat"
    TriggerFirstMission   = "first_mission"
    TriggerFirstJump      = "first_jump"
)
```

**Handle Trigger**:
```go
func (m *Manager) HandleTrigger(
    playerID uuid.UUID,
    trigger models.TutorialTrigger,
) {
    m.mu.Lock()
    defer m.mu.Unlock()

    progress, exists := m.progress[playerID]
    if !exists || !progress.TutorialEnabled {
        return
    }

    tutorialIDs, exists := m.triggers[trigger]
    if !exists {
        return
    }

    // Activate triggered tutorials if prerequisites are met
    for _, tutorialID := range tutorialIDs {
        tutorial, exists := m.tutorials[tutorialID]
        if !exists {
            continue
        }

        if !tutorial.IsCompleted(progress) &&
            m.prerequisitesMet(tutorial, progress) {
            progress.CurrentStep = tutorial.Steps[0].ID
        }
    }
}
```

### Progress Tracking

**Progress Structure**:
```go
type TutorialProgress struct {
    PlayerID         uuid.UUID
    TutorialEnabled  bool
    CurrentStep      string
    CompletedSteps   []string
    SkippedSteps     []string
    CategoryProgress map[TutorialCategory]int // Category -> completed count
    TotalSteps       int
    CompletedCount   int
    StartedAt        time.Time
    LastUpdated      time.Time
}
```

**Complete Step**:
```go
func (m *Manager) CompleteStep(playerID uuid.UUID, stepID string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    progress, exists := m.progress[playerID]
    if !exists {
        return
    }

    // Find which tutorial this step belongs to
    for _, tutorial := range m.tutorials {
        for _, step := range tutorial.Steps {
            if step.ID == stepID {
                progress.CompleteStep(stepID, tutorial.Category)
                step.Completed = true
                return
            }
        }
    }
}
```

---

## User Interface

### Tutorial Overlay

**Overlay Display**:
```
┌──────────────────────────────────────────────────────┐
│                   📚 TUTORIAL                        │
├──────────────────────────────────────────────────────┤
│ Trading 101: Buying Commodities                      │
│                                                       │
│ Each planet has a market where you can buy and sell  │
│ commodities. Buy low and sell high for profit!       │
│                                                       │
│ Objective:                                            │
│ Buy at least one unit of any commodity                │
│                                                       │
│ Hint: Press 'B' to buy after selecting a commodity   │
├──────────────────────────────────────────────────────┤
│ H: Hint  •  S: Skip  •  T: Hide  •  D: Disable       │
└──────────────────────────────────────────────────────┘
```

**Tutorial List**:
```
=== Available Tutorials ===

○ Welcome to Terminal Velocity (0/3 steps - 0%)
◐ Space Trading 101 (2/3 steps - 67%)
✓ Galactic Navigation (3/3 steps - 100%)
○ Ship Management (0/3 steps - 0%)
○ Space Combat Basics (0/3 steps - 0%)
○ Mission System (0/3 steps - 0%)
○ Multiplayer Features (0/3 steps - 0%)

Overall Progress: 5/20 steps (25%)

↑/↓: Navigate  •  Enter: View  •  ESC: Back
```

### Navigation

- **H**: Show next hint
- **S**: Skip current step
- **T**: Toggle overlay visibility
- **D**: Disable all tutorials
- **Enter**: Complete step / Continue
- **ESC**: Close tutorial screen

---

## Integration with Other Systems

### Settings Integration

Tutorial mode controlled by settings:

```go
// In settings
Settings.Gameplay.TutorialMode: bool
Settings.Display.ShowTutorialTips: bool
```

### Screen Integration

Tutorials appear on relevant screens:

```go
// In each screen's view
func (m Model) viewTrading() string {
    content := renderTradingScreen()

    // Add tutorial overlay if active
    return m.renderTutorialOverlay(content)
}
```

---

## API Reference

### Core Functions

#### InitializePlayer

```go
func (m *Manager) InitializePlayer(
    playerID uuid.UUID,
) *models.TutorialProgress
```

Creates tutorial progress for a new player.

#### CompleteStep

```go
func (m *Manager) CompleteStep(
    playerID uuid.UUID,
    stepID string,
)
```

Marks a tutorial step as completed.

#### HandleTrigger

```go
func (m *Manager) HandleTrigger(
    playerID uuid.UUID,
    trigger models.TutorialTrigger,
)
```

Processes a tutorial trigger event.

#### GetCurrentStep

```go
func (m *Manager) GetCurrentStep(
    playerID uuid.UUID,
) *models.TutorialStep
```

Returns the current active step for a player.

---

## Related Documentation

- [Settings System](./SETTINGS_SYSTEM.md) - Tutorial mode setting
- [UI Screens](./UI.md) - Screen integration

---

## File Locations

**Core Implementation**:
- `internal/tutorial/manager.go` - Tutorial manager
- `internal/models/tutorial.go` - Tutorial models

**User Interface**:
- `internal/tui/tutorial.go` - Tutorial UI

**Documentation**:
- `docs/TUTORIAL_SYSTEM.md` - This file
- `ROADMAP.md` - Phase 19 details

---

**For questions about the tutorial system, contact the development team.**
