# Navigation & Space Screens

This document covers all navigation and space-related UI screens in Terminal Velocity.

## Overview

**Screens**: 5
- Space View (Main Flight)
- Navigation/Map Screen
- Navigation Enhanced
- Landing Screen
- Game View (General)

**Purpose**: Core gameplay screens for flying, navigating between systems, landing on planets, and spatial awareness.

**Source Files**:
- `internal/tui/space_view.go` - Main space flight view
- `internal/tui/navigation.go` - Navigation and jump map
- `internal/tui/navigation_enhanced.go` - Enhanced navigation features
- `internal/tui/landing.go` - Planetary landing services
- `internal/tui/game.go` - General game view wrapper

---

## Space View (Main Flight)

### Source File
`internal/tui/space_view.go`

### Purpose
Primary in-flight screen showing real-time space environment, nearby objects, targets, and ship status. This is the "home" screen for active gameplay.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ TERMINAL VELOCITY v1.0          [Sol System]          Shields: ████████░░ 80%┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                                                               ║        ┃
┃    ║                          *                                    ║  ┏━━━━━━━━━━━━━┓
┃    ║                                                               ║  ┃   RADAR     ┃
┃    ║             *                    ⊕ Earth                      ║  ┃             ┃
┃    ║                                                               ║  ┃      *      ┃
┃    ║                                                               ║  ┃             ┃
┃    ║        *                                                      ║  ┃   ⊕    ◆    ┃
┃    ║                                 △                             ║  ┃        ▲    ┃
┃    ║                                You                            ║  ┃      *      ┃
┃    ║                                                               ║  ┃             ┃
┃    ║                                             ◆ Pirate          ║  ┗━━━━━━━━━━━━━┛
┃    ║           ⊕ Mars                                              ║
┃    ║                                                               ║  ┏━━━━━━━━━━━━━┓
┃    ║  *                                                            ║  ┃   STATUS    ┃
┃    ║                                                               ║  ┣━━━━━━━━━━━━━┫
┃    ║                                                       *       ║  ┃ Hull: ██████┃
┃    ║                                                               ║  ┃       100%  ┃
┃    ╚═══════════════════════════════════════════════════════════════╝  ┃ Fuel: ████░░┃
┃                                                                        ┃       67%   ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ Speed: 340  ┃
┃  ┃ TARGET: Pirate Viper    ┃  ┃ CARGO: 15/50 tons                ┃  ┃ Credits:    ┃
┃  ┃ Distance: 2,340 km      ┃  ┃ ▪ Food (10t)  ▪ Electronics (5t) ┃  ┃  52,400 cr  ┃
┃  ┃ Shields: 45%            ┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━┛
┃  ┃ Attitude: Hostile       ┃                                                        ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━┛                                                        ┃
┃ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ┃
┃ ┃ CHAT [Global] ▼                                               [C] to expand ┃ ┃
┃ ┃ SpaceCadet: Anyone near Sol system?                                  3m ago ┃ ┃
┃ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [L]and  [J]ump  [T]arget  [F]ire  [H]ail  [M]ap  [C]hat  [I]nfo  [ESC] Menu ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Main Viewport**: Top-down 2D space view showing your ship, planets, enemies, stars
- **Radar**: Miniature overview of relative positions
- **Status Panel**: Hull, fuel, speed, credits
- **Target Info Panel**: Selected target details and attitude
- **Cargo Summary**: Quick cargo overview
- **Chat Bar**: Minimized chat with expand option
- **Command Bar**: Quick-access flight commands

### Visual Legend
- `△` = Your ship
- `⊕` = Planet/Station
- `◆` = Hostile ship
- `◇` = Neutral ship
- `◈` = Friendly ship
- `*` = Stars/background
- `▲` = Your ship on radar

### Key Bindings
- `L` - Land on nearest planet (when in range)
- `J` - Open jump/navigation map
- `T` - Cycle targets (Tab also works)
- `F` - Fire weapons at target
- `H` - Hail target ship
- `M` - Open system map
- `C` - Expand chat interface
- `I` - View system/target info
- `ESC` - Open main menu

### State Management

**Model Structure** (`spaceViewModel`):
```go
type spaceViewModel struct {
    playerShip    *models.Ship
    nearbyObjects []SpaceObject  // Planets, ships, etc.
    selectedTarget *SpaceObject
    targetIndex   int
    chatExpanded  bool
    width         int
    height        int
}

type SpaceObject struct {
    Type      string  // "planet", "ship", "station"
    Name      string
    Position  Position
    Distance  float64
    Attitude  string  // "hostile", "neutral", "friendly"
    ShipData  *models.Ship
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `targetCycledMsg` - Target selection changed
- `objectsUpdatedMsg` - Nearby objects refreshed
- `combatInitiatedMsg` - Combat encounter started

### Data Flow
1. Load current system data from `SystemRepository`
2. Query nearby ships/planets
3. Update object positions (if real-time movement implemented)
4. Handle player input (targeting, firing, landing)
5. Transition to combat if hostile engagement
6. Transition to landing screen when docking

### Related Screens
- **Navigation Map** - Press `J` or `M`
- **Landing Screen** - Press `L` near planet
- **Combat Screen** - Auto-transition on hostile engagement
- **Chat Screen** - Press `C` to expand
- **Main Menu** - Press `ESC`

---

## Navigation Map

### Source Files
- `internal/tui/navigation.go` - Standard navigation
- `internal/tui/navigation_enhanced.go` - Enhanced features

### Purpose
System map showing jump routes, nearby systems, and hyperjump interface.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ NAVIGATION MAP                    [Sol System]               52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                    NEARBY SYSTEMS                             ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║         Barnard's Star                   Proxima Centauri    ║        ┃
┃    ║              ◉ (8.2 ly)                      ◉ (4.5 ly)       ║        ┃
┃    ║                  ╲                            ╱               ║        ┃
┃    ║                   ╲                          ╱                ║        ┃
┃    ║                    ╲                        ╱                 ║        ┃
┃    ║                     ╲                      ╱                  ║        ┃
┃    ║                      ╲                    ╱                   ║        ┃
┃    ║                       ╲                  ╱                    ║        ┃
┃    ║                        ╲                ╱                     ║        ┃
┃    ║                         ╲              ╱                      ║        ┃
┃    ║                       ⊕══════⊙════════◉                       ║        ┃
┃    ║                      Earth  SOL  Alpha Centauri               ║        ┃
┃    ║                      (YOU)   ▲    (3.2 ly)                    ║        ┃
┃    ║                                                               ║        ┃
┃    ║              Sirius                                           ║        ┃
┃    ║               ◉ (6.8 ly)                                      ║        ┃
┃    ║                                                               ║        ┃
┃    ╚═══════════════════════════════════════════════════════════════╝        ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ JUMP DESTINATIONS:             ┃  ┃ SELECTED: Alpha Centauri        ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃ ▶ Alpha Centauri   (3.2 ly)   ┃  ┃ Distance: 3.2 light years       ┃  ┃
┃  ┃   Proxima Centauri (4.5 ly)   ┃  ┃ Fuel Required: 32 units         ┃  ┃
┃  ┃   Sirius           (6.8 ly)   ┃  ┃ Your Fuel: 201/300 units        ┃  ┃
┃  ┃   Barnard's Star   (8.2 ly)   ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ Government: Confederation       ┃  ┃
┃  ┃                                ┃  ┃ Tech Level: 8                   ┃  ┃
┃  ┃                                ┃  ┃ Population: 2.4 billion         ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ Services:                       ┃  ┃
┃  ┃                                ┃  ┃ ✓ Shipyard  ✓ Outfitter         ┃  ┃
┃  ┃                                ┃  ┃ ✓ Missions  ✓ Refuel            ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ [ Engage Hyperdrive ]           ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select System  [Enter] Jump  [I]nfo  [ESC] Back to Space              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **System Map**: Visual representation of connected systems
- **Jump Route Lines**: Show hyperspace lanes (MST-based)
- **Destination List**: Scrollable list of reachable systems
- **System Info Panel**: Details about selected destination
- **Fuel Calculator**: Shows required fuel vs available
- **Jump Button**: Confirms hyperjump

### Visual Legend
- `⊙` = Current system (you are here)
- `◉` = Reachable system
- `⊕` = Planet in current system
- `══`, `╱`, `╲` = Jump routes
- `▲` = Current position marker

### Key Bindings
- `↑`/`↓` or `J`/`K` - Select destination system
- `Enter` - Engage hyperdrive to selected system
- `I` - View detailed system information
- `ESC` - Return to space view

### State Management

**Model Structure** (`navigationModel`):
```go
type navigationModel struct {
    currentSystem    *models.StarSystem
    connectedSystems []*models.StarSystem
    selectedIndex    int
    playerShip       *models.Ship
    width            int
    height           int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `systemsLoadedMsg` - Connected systems data loaded
- `jumpInitiatedMsg` - Hyperjump started
- `jumpFailedMsg` - Insufficient fuel or other error

### Data Flow
1. Load current system from database
2. Query connected systems via `system_connections` table
3. Calculate fuel requirements (distance * 10 units/ly)
4. On jump: validate fuel, deduct fuel, update player location
5. Transition to new system's space view
6. Trigger random encounter check

### Jump Mechanics
- **Fuel Cost**: 10 units per light-year
- **Jump Time**: Instant (could add animation)
- **Encounter Chance**: 10% per jump (pirates, traders, events)
- **Fuel Check**: Must have sufficient fuel before jump
- **Unreachable Systems**: Grayed out if too far or disconnected

### Navigation Enhanced Features
Additional features in `navigation_enhanced.go`:
- Multi-hop route planning (A* pathfinding)
- Trade route suggestions
- Danger rating display (pirate activity)
- Faction territory coloring
- Bookmark favorite systems

### Related Screens
- **Space View** - Return with `ESC`
- **System Info** - Press `I` on selected system
- **Trade Routes** - Advanced trade planning
- **Combat/Encounter** - May trigger on jump

---

## Landing Screen

### Source File
`internal/tui/landing.go`

### Purpose
Planetary/station services menu when docked. Gateway to trading, shipyard, outfitter, missions, and other planet-based activities.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ EARTH STATION                    United Earth                  52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                                                               ║        ┃
┃    ║            Welcome to Earth Station, Commander.               ║        ┃
┃    ║                                                               ║        ┃
┃    ║         [ASCII art of planet/station could go here]           ║        ┃
┃    ║                       _______________                         ║        ┃
┃    ║                      /               \                        ║        ┃
┃    ║                     /    ⊕  EARTH     \                       ║        ┃
┃    ║                    |  (Terran Alliance)|                      ║        ┃
┃    ║                     \     Pop: 8.2B    /                      ║        ┃
┃    ║                      \_____    _______/                       ║        ┃
┃    ║                        /   \__/   \                           ║        ┃
┃    ║                       /  Station   \                          ║        ┃
┃    ║                       \____________/                          ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ╚═══════════════════════════════════════════════════════════════╝        ┃
┃                                                                              ┃
┃    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓    ┃
┃    ┃  AVAILABLE SERVICES:       ┃  ┃  SHIP STATUS:                   ┃    ┃
┃    ┃                            ┃  ┃                                 ┃    ┃
┃    ┃  [C] Commodity Exchange    ┃  ┃  Ship: Corvette "Starhawk"      ┃    ┃
┃    ┃  [O] Outfitters            ┃  ┃  Hull: 100%  Shields: 80%       ┃    ┃
┃    ┃  [S] Shipyard              ┃  ┃  Fuel: 67%   Cargo: 15/50t      ┃    ┃
┃    ┃  [M] Mission BBS           ┃  ┃                                 ┃    ┃
┃    ┃  [B] Bar & News            ┃  ┃  Current System: Sol            ┃    ┃
┃    ┃  [R] Refuel (1,200 cr)     ┃  ┃  Government: United Earth       ┃    ┃
┃    ┃  [H] Repairs (Free)        ┃  ┃  Tech Level: 9                  ┃    ┃
┃    ┃                            ┃  ┃                                 ┃    ┃
┃    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛    ┃
┃                                                                              ┃
┃    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓    ┃
┃    ┃ NEWS: Pirate activity reported in nearby systems...             ┃    ┃
┃    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛    ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [T]akeoff  [Tab] Next Service  [ESC] Exit                                   ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Welcome Banner**: Planet/station art and name
- **Services Menu**: Available facilities based on tech level
- **Ship Status Panel**: Current ship condition
- **News Ticker**: Recent system events
- **Service Icons**: Visual indicators for each facility

### Key Bindings
- `C` - Commodity Exchange (trading screen)
- `O` - Outfitters (equipment shop)
- `S` - Shipyard (buy/sell ships)
- `M` - Mission BBS (accept missions)
- `B` - Bar & News (news feed, rumors)
- `R` - Refuel ship
- `H` - Repair hull and recharge shields
- `T` - Takeoff (return to space view)
- `Tab` - Navigate between services
- `ESC` - Quick takeoff

### State Management

**Model Structure** (`landingModel`):
```go
type landingModel struct {
    planet         *models.Planet
    system         *models.StarSystem
    availableServices []string
    selectedService int
    playerShip     *models.Ship
    newsItems      []*models.NewsItem
    width          int
    height         int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `servicesLoadedMsg` - Planet services loaded
- `refuelCompleteMsg` - Refueling completed
- `repairsCompleteMsg` - Repairs completed

### Data Flow
1. Load planet data from database
2. Determine available services (based on tech level, government)
3. Display service menu
4. On service selection, transition to appropriate screen
5. Return to landing screen when service complete
6. On takeoff, transition back to space view

### Service Availability
Services determined by planet attributes:
- **Tech Level**: Higher tech = more services
- **Government**: Some governments restrict certain services
- **Planet Type**: Stations vs planets have different offerings

Example service requirements:
- Commodity Exchange: TechLevel >= 1 (all planets)
- Shipyard: TechLevel >= 5
- Outfitter: TechLevel >= 3
- Missions: Always available
- Refuel/Repair: Always available

### Refuel/Repair Costs
- **Fuel**: 10 cr per unit
- **Hull Repair**: 50 cr per point
- **Shield Recharge**: Free (automatic when docked)

### Related Screens
- **Trading Screen** - Press `C`
- **Shipyard Screen** - Press `S`
- **Outfitter Screen** - Press `O`
- **Missions Screen** - Press `M`
- **News Screen** - Press `B`
- **Space View** - Press `T` to takeoff

---

## Game View Wrapper

### Source File
`internal/tui/game.go`

### Purpose
General wrapper/container for game screens. Handles high-level state management and screen transitions.

### State Management

**Model Structure** (`gameViewModel`):
```go
type gameViewModel struct {
    currentLocation string  // "space", "landed", "combat", etc.
    activeScreen    Screen
    player          *models.Player
    ship            *models.Ship
    system          *models.StarSystem
    width           int
    height          int
}
```

### Responsibilities
- Track overall game state (in space, docked, in combat, etc.)
- Manage transitions between major game modes
- Handle autosave triggers
- Coordinate with session manager
- Dispatch to appropriate sub-screens

---

## Implementation Notes

### Spatial Calculations
Position and distance calculations in `internal/game/universe/`:
- Systems have X,Y,Z coordinates
- Distance formula: `sqrt((x2-x1)² + (y2-y1)² + (z2-z1)²)`
- Light-year conversion for display
- Fuel costs based on distance

### Jump Route Generation
MST-based connectivity in `internal/game/universe/generator.go`:
- Prim's algorithm for minimum spanning tree
- Additional random connections for variety
- Ensures all systems reachable
- Stored in `system_connections` table

### Real-Time Updates
Space view considerations:
- Ships move (if real-time movement implemented)
- Chat messages appear
- Other players visible (presence system)
- Background star twinkling (visual effect)

### Performance
- Cache nearby objects to reduce database queries
- Update positions every tick (if movement enabled)
- Lazy-load system data as needed
- Efficient rendering with Lipgloss

### Testing
Test files:
- `internal/tui/navigation_test.go` - Navigation tests
- `internal/tui/integration_test.go` - Full navigation flow tests
- Unit tests for jump calculations
- Fuel consumption edge cases

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
