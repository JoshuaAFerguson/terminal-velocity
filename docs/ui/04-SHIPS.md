# Ship Management Screens

This document covers all ship-related UI screens in Terminal Velocity.

## Overview

**Screens**: 6
- Shipyard Screen
- Shipyard Enhanced Screen
- Outfitter Screen
- Outfitter Enhanced Screen
- Ship Management Screen
- Fleet Screen

**Purpose**: Handle all ship-related activities including purchasing ships, installing equipment, managing loadouts, and fleet operations.

**Source Files**:
- `internal/tui/shipyard.go` - Ship purchasing interface
- `internal/tui/shipyard_enhanced.go` - Advanced ship comparison
- `internal/tui/outfitter.go` - Equipment installation
- `internal/tui/outfitter_enhanced.go` - Advanced outfitting with loadouts
- `internal/tui/ship_management.go` - Ship services and status
- `internal/tui/fleet.go` - Multi-ship fleet management

---

## Shipyard Screen

### Source File
`internal/tui/shipyard.go`

### Purpose
Ship purchasing interface showing available ships, stats, and trade-in values.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ SHIPYARD - Earth Station                                     52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE SHIPS:           ┃  ┃ SHIP: LIGHTNING FIGHTER              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                            ┃  ┃                                      ┃  ┃
┃  ┃   Shuttle        12,000 cr ┃  ┃         ___                          ┃  ┃
┃  ┃ ▶ Lightning      45,000 cr ┃  ┃        /   \___                      ┃  ┃
┃  ┃   Courier        75,000 cr ┃  ┃       |  △  ___>                     ┃  ┃
┃  ┃   Corvette      180,000 cr ┃  ┃        \___/                         ┃  ┃
┃  ┃   Destroyer     450,000 cr ┃  ┃                                      ┃  ┃
┃  ┃   Freighter     220,000 cr ┃  ┃  Class: Light Fighter                ┃  ┃
┃  ┃   Cruiser       780,000 cr ┃  ┃  Price: 45,000 cr                    ┃  ┃
┃  ┃   Battleship  1,500,000 cr ┃  ┃                                      ┃  ┃
┃  ┃                            ┃  ┃  Hull: ████░░ 80                     ┃  ┃
┃  ┃                            ┃  ┃  Shields: ███░░░ 60                  ┃  ┃
┃  ┃                            ┃  ┃  Speed: ████████ 450                 ┃  ┃
┃  ┃                            ┃  ┃  Accel: ███████░ 380                 ┃  ┃
┃  ┃                            ┃  ┃  Maneuver: ████████ 420              ┃  ┃
┃  ┃                            ┃  ┃                                      ┃  ┃
┃  ┃                            ┃  ┃  Cargo: 15 tons                      ┃  ┃
┃  ┃                            ┃  ┃  Fuel: 300 units                     ┃  ┃
┃  ┃                            ┃  ┃  Weapon Slots: 2                     ┃  ┃
┃  ┃                            ┃  ┃  Outfit Slots: 3                     ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR SHIP: Corvette "Starhawk"                     Trade-in: 126,000 ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Purchase Lightning for 45,000 cr?                                   ┃  ┃
┃  ┃ With trade-in credit: You will GAIN 81,000 cr                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ [ Purchase ] [ Trade-In Purchase ] [ Cancel ]                       ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Ship  [Enter] Details  [P]urchase  [T]rade-In  [ESC] Back      ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Ship List**: All ships available at current station
- **Ship Detail Panel**: Stats and ASCII art for selected ship
- **Transaction Panel**: Purchase options and trade-in calculations
- **Stat Bars**: Visual representation of ship capabilities

### Ship Types (11 total)

1. **Shuttle** - Starter ship, minimal combat, small cargo
2. **Lightning** - Light fighter, fast and agile
3. **Courier** - Fast transport, good cargo, light defenses
4. **Corvette** - Balanced combat ship
5. **Destroyer** - Heavy combat, military-grade
6. **Freighter** - Maximum cargo, slow, minimal weapons
7. **Clipper** - Fast trader, good cargo and speed
8. **Gunship** - Heavy weapons platform
9. **Cruiser** - Large combat ship, expensive
10. **Carrier** - Fleet command ship, fighter bay
11. **Battleship** - Ultimate combat vessel

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate ship list
- `Enter` - View detailed ship information
- `P` - Purchase ship outright (requires full price)
- `T` - Trade-in purchase (current ship credit applied)
- `C` - Compare with current ship (side-by-side)
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`shipyardModel`):
```go
type shipyardModel struct {
    availableShips  []*models.ShipType
    selectedIndex   int
    playerShip      *models.Ship
    playerCredits   int
    tradeInValue    int
    purchaseMode    string  // "outright", "tradein"
    width           int
    height          int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `shipsLoadedMsg` - Available ships loaded (filtered by tech level)
- `purchaseCompleteMsg` - Ship purchased successfully
- `purchaseErrorMsg` - Insufficient funds or error
- `tradeInCalculatedMsg` - Trade-in value computed

### Data Flow
1. Load available ships (filtered by station tech level)
2. Calculate trade-in value of current ship (70% of base price)
3. Display ship list with prices
4. User selects ship to view details
5. User chooses purchase method
6. Validate transaction (credits, tech requirements)
7. Transfer cargo from old ship to new ship
8. Update player ship in database
9. Deduct credits or apply trade-in

### Ship Availability

**Tech Level Requirements**:
- Tech 1-3: Shuttle only
- Tech 4-5: Shuttle, Lightning, Courier
- Tech 6-7: All except Battleship
- Tech 8-9: All ships available

**Government Restrictions**:
- Pirate systems: No military ships (Destroyer, Battleship)
- Corporate systems: Premium prices (+20%)
- Military systems: Military ships available, discount on combat vessels

### Purchase Mechanics

**Outright Purchase**:
- Pay full price in credits
- Keep current ship (added to fleet if fleet system implemented)
- No price reduction

**Trade-In Purchase**:
- Current ship sold for 70% of base price
- Trade-in value deducted from purchase price
- Cargo transferred automatically (if space available)
- Equipped items transferred (if slots available)
- Excess cargo/items must be sold or jettisoned

**Trade-In Value Calculation**:
```go
basePrice := currentShip.BasePrice
condition := float64(currentShip.Hull) / float64(currentShip.MaxHull)
tradeInValue := int(float64(basePrice) * 0.7 * condition)
```

### Ship Statistics

**Combat Stats**:
- **Hull**: Total hit points before destruction
- **Shields**: Regenerating protection layer
- **Armor**: Damage reduction percentage
- **Speed**: Maximum velocity
- **Acceleration**: How quickly ship reaches max speed
- **Maneuverability**: Turn rate and agility

**Capacity Stats**:
- **Cargo**: Tons of cargo space
- **Fuel**: Maximum fuel capacity (affects jump range)
- **Weapon Slots**: Number of weapons that can be installed
- **Outfit Slots**: Number of equipment items
- **Crew**: Crew capacity (affects some abilities)

### Related Screens
- **Ship Management** - Services for current ship
- **Outfitter** - Install equipment on ship
- **Shipyard Enhanced** - Advanced comparison features
- **Fleet** - Manage multiple owned ships

---

## Outfitter Screen

### Source File
`internal/tui/outfitter.go`

### Purpose
Equipment installation interface for weapons, defenses, and ship systems.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ OUTFITTER - Earth Station                                    52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ WEAPONS                        ┃  ┃ OUTFIT: SHIELD BOOSTER MK2      ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃   Laser Cannon      8,500 cr   ┃  ┃                                 ┃  ┃
┃  ┃   Pulse Laser      12,000 cr   ┃  ┃  "Advanced shield regeneration  ┃  ┃
┃  ┃   Blaze Cannon     18,000 cr   ┃  ┃   matrix increases your shield  ┃  ┃
┃  ┃   Proton Turret    45,000 cr   ┃  ┃   strength by 50 points."       ┃  ┃
┃  ┃   Missile Launcher 25,000 cr   ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃  Price: 35,000 cr               ┃  ┃
┃  ┃ DEFENSE                        ┃  ┃  Mass: 5 tons                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃                                 ┃  ┃
┃  ┃   Shield Booster     15,000 cr ┃  ┃  Effect: +50 shields            ┃  ┃
┃  ┃ ▶ Shield Boost MK2   35,000 cr ┃  ┃          +10 regen/sec          ┃  ┃
┃  ┃   Armor Plating      22,000 cr ┃  ┃                                 ┃  ┃
┃  ┃   Point Defense       8,500 cr ┃  ┃  Slots Used: 1                  ┃  ┃
┃  ┃                                ┃  ┃  Tech Level Required: 7         ┃  ┃
┃  ┃ SYSTEMS                        ┃  ┃                                 ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃  [ Purchase (Qty: 1) ]          ┃  ┃
┃  ┃   Cargo Pod         10,000 cr  ┃  ┃  [ Install & Purchase ]         ┃  ┃
┃  ┃   Fuel Tank          7,500 cr  ┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃  ┃   Engine Upgrade    50,000 cr  ┃                                       ┃
┃  ┃   Scanner Array     12,000 cr  ┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃ YOUR SHIP: Corvette            ┃  ┃
┃                                       ┃                                 ┃  ┃
┃                                       ┃ Outfit Slots: 4/6 used          ┃  ┃
┃                                       ┃ Weapon Slots: 2/4 used          ┃  ┃
┃                                       ┃                                 ┃  ┃
┃                                       ┃ Installed:                      ┃  ┃
┃                                       ┃  ▪ Laser Cannon (2x)            ┃  ┃
┃                                       ┃  ▪ Shield Booster               ┃  ┃
┃                                       ┃  ▪ Cargo Pod                    ┃  ┃
┃                                       ┃  ▪ Fuel Tank                    ┃  ┃
┃                                       ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Tab] Category  [B]uy  [S]ell  [I]nstall  [U]ninstall  [ESC]  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Equipment Categories**: Weapons, Defense, Systems (6 types total)
- **Item List**: Available equipment in selected category
- **Item Details**: Stats and effects of selected item
- **Ship Status**: Current equipment and slot usage
- **Action Buttons**: Buy, sell, install, uninstall

### Equipment Categories (6 types)

1. **Weapons**: Laser cannons, missiles, turrets (9 weapon types)
2. **Defense**: Shields, armor, point defense
3. **Systems**: Engines, scanners, jammers
4. **Cargo**: Cargo pods, specialized holds
5. **Fuel**: Fuel tanks, ram scoops
6. **Special**: Cloaking, tractor beams, unique items

### Equipment Items (16 total)

**Weapons** (9 types):
- Laser Cannon - Basic energy weapon
- Pulse Laser - Improved laser with higher damage
- Blaze Cannon - Rapid-fire energy weapon
- Proton Turret - Heavy energy turret
- Missile Launcher - Homing missiles
- Torpedo Launcher - Heavy missiles
- Mass Driver - Projectile weapon
- Railgun - High-velocity kinetic weapon
- Plasma Cannon - Advanced energy weapon

**Defense** (3 types):
- Shield Booster - Increases shield capacity
- Shield Booster MK2 - Advanced shield enhancement
- Armor Plating - Damage reduction

**Systems** (4 types):
- Cargo Pod - +10 tons cargo
- Fuel Tank - +100 fuel capacity
- Engine Upgrade - +20% speed
- Scanner Array - Extended sensor range

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate equipment list
- `Tab` - Switch between equipment categories
- `B` - Buy selected equipment (add to inventory)
- `S` - Sell equipment from inventory
- `I` - Install equipment from inventory to ship
- `U` - Uninstall equipment from ship to inventory
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`outfitterModel`):
```go
type outfitterModel struct {
    categories       []string
    selectedCategory int
    equipment        []*models.Equipment
    selectedIndex    int
    playerShip       *models.Ship
    installedItems   []*models.Equipment
    inventory        []*models.Equipment  // Owned but not installed
    playerCredits    int
    width            int
    height           int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `equipmentLoadedMsg` - Available equipment loaded
- `purchaseCompleteMsg` - Equipment purchased
- `installCompleteMsg` - Equipment installed on ship
- `uninstallCompleteMsg` - Equipment removed from ship
- `sellCompleteMsg` - Equipment sold

### Data Flow
1. Load available equipment (filtered by tech level)
2. Display equipment by category
3. User selects item and action
4. Validate action (slots, credits, requirements)
5. Execute transaction
6. Update ship stats if equipment affects performance
7. Refresh display

### Equipment System

**Slot Types**:
- **Weapon Slots**: Fixed by ship type, cannot be increased
- **Outfit Slots**: For non-weapon equipment
- **Turret Slots**: For turret weapons (360° rotation)
- **Fighter Bay Slots**: For carried fighters (Carrier only)
- **Special Slots**: For unique items

**Installation Requirements**:
- Sufficient slot space
- Tech level requirement met
- Some items require specific ship classes
- Mass limit (total outfit mass < ship capacity)

**Effects on Ship**:
- **Weapons**: Increase firepower in combat
- **Shields**: Increase shield capacity and regen
- **Armor**: Reduce incoming damage
- **Cargo Pods**: Increase cargo capacity
- **Fuel Tanks**: Increase maximum fuel (longer jumps)
- **Engines**: Increase speed and maneuverability
- **Scanners**: Detect cloaked ships, show more info

### Equipment Pricing

**Purchase**: Full retail price
**Sell**: 80% of purchase price (better than ship trade-in)
**Installation**: Free at stations
**Uninstallation**: Free (item goes to inventory)

### Tech Level Availability

- Tech 1-3: Basic weapons, small cargo pods
- Tech 4-5: Improved weapons, shield boosters
- Tech 6-7: Advanced weapons, armor plating
- Tech 8-9: Military-grade weapons, special systems

### Related Screens
- **Outfitter Enhanced** - Advanced loadout management
- **Ship Management** - View total ship stats
- **Shipyard** - Ship slot capacities

---

## Ship Management Screen

### Source File
`internal/tui/ship_management.go`

### Purpose
Ship services interface for repairs, refueling, and status overview.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ SHIP SERVICES - Earth Station                               52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR SHIP: Corvette "Starhawk"                                       ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃          ___                                                         ┃  ┃
┃  ┃         /   \___                                                     ┃  ┃
┃  ┃        |  ▲  ====>                                                   ┃  ┃
┃  ┃         \___/                                                        ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Class: Light Corvette              Value: 180,000 cr               ┃  ┃
┃  ┃  Mass: 150 tons                     Age: 47 days                    ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ SHIP STATUS                                                          ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Hull Integrity:     [████████░░] 85/100  (Good)                    ┃  ┃
┃  ┃  Shield Charge:      [██████████] 100/100 (Full)                    ┃  ┃
┃  ┃  Fuel:               [████░░░░░░] 201/300 (67%)                     ┃  ┃
┃  ┃  Cargo:              [███░░░░░░░] 15/50 tons (30%)                  ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Engine Status:      Operational                                    ┃  ┃
┃  ┃  Weapons Status:     All Systems Nominal                            ┃  ┃
┃  ┃  Life Support:       Optimal                                        ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE SERVICES                                                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ▶ [R] Repair Hull (15 points)                         750 cr       ┃  ┃
┃  ┃    [F] Refuel (99 units to full)                       990 cr       ┃  ┃
┃  ┃    [S] Recharge Shields                                FREE         ┃  ┃
┃  ┃    [C] Complete Service (All)                        1,740 cr       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃    [N] Rename Ship                                     100 cr       ┃  ┃
┃  ┃    [P] Paint Job / Cosmetics                         1,000 cr       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ──────────────────────────────────────────────────────────────      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Current Repairs: None in progress                                  ┃  ┃
┃  ┃  Estimated Time: Instant (docked)                                   ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Service  [Enter] Confirm  [A]uto-Repair  [ESC] Back            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Ship Overview**: Name, class, ASCII art, value
- **Status Panel**: Hull, shields, fuel, cargo levels
- **Services Menu**: Available maintenance and customization
- **Cost Calculator**: Service costs displayed

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate services
- `R` - Repair hull damage
- `F` - Refuel ship to maximum
- `S` - Recharge shields (free when docked)
- `C` - Complete service (all repairs and refuel)
- `N` - Rename ship
- `P` - Apply cosmetic customization
- `A` - Auto-repair (enable/disable auto-service on dock)
- `Enter` - Confirm selected service
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`shipManagementModel`):
```go
type shipManagementModel struct {
    ship            *models.Ship
    selectedService int
    serviceCosts    map[string]int
    autoRepair      bool
    playerCredits   int
    width           int
    height          int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `serviceCompleteMsg` - Service performed
- `serviceErrorMsg` - Insufficient funds
- `shipRenamedMsg` - Ship name changed

### Data Flow
1. Load current ship status
2. Calculate required services (hull damage, fuel needed)
3. Display service menu with costs
4. User selects service
5. Validate credits
6. Perform service (update ship stats)
7. Deduct credits
8. Refresh display

### Service Types

**Repairs**:
- **Cost**: 50 cr per hull point
- **Time**: Instant when docked
- **Effect**: Restores hull to 100%

**Refueling**:
- **Cost**: 10 cr per fuel unit
- **Time**: Instant
- **Effect**: Refills fuel tank to capacity

**Shield Recharge**:
- **Cost**: Free when docked
- **Time**: Instant
- **Effect**: Shields to 100%

**Complete Service**:
- All of the above in one action
- Small discount (5%) vs individual services

**Cosmetic Services**:
- **Rename Ship**: 100 cr
- **Paint Job**: 1,000 cr (visual customization)
- **Ship Insignia**: 500 cr (faction/personal symbol)

### Auto-Repair Feature

When enabled:
- Automatically repairs hull on every landing
- Automatically refuels to maximum
- Deducts credits automatically
- Can be enabled/disabled in settings
- Warning if insufficient credits

### Related Screens
- **Shipyard** - Purchase new ships
- **Outfitter** - Install equipment
- **Landing** - Access ship services

---

## Outfitter Enhanced & Fleet Screens

Due to length constraints, here are summaries:

### Outfitter Enhanced Screen
(`internal/tui/outfitter_enhanced.go`)

**Features**:
- Loadout saving/loading system
- Side-by-side equipment comparison
- DPS calculator for weapons
- Mass and power budget visualization
- Recommended builds for ship class

**Loadouts**:
- Save current equipment configuration
- Load saved configurations
- Clone loadouts with modifications
- Share loadouts with faction

### Fleet Screen
(`internal/tui/fleet.go`)

**Features**:
- Manage multiple owned ships
- Switch active ship
- View fleet total value
- Assign ships to hangar locations
- Fleet-wide statistics

---

## Implementation Notes

### Database Integration
- `database.ShipRepository` - Ship ownership and stats
- `database.EquipmentRepository` - Equipment inventory
- `database.LoadoutRepository` - Saved configurations

### Equipment Effects

Equipment modifies ship stats in real-time:
```go
func CalculateShipStats(ship *Ship, equipment []*Equipment) *ShipStats {
    stats := ship.BaseStats
    for _, item := range equipment {
        stats.Shields += item.ShieldBonus
        stats.Hull += item.HullBonus
        stats.Speed *= item.SpeedMultiplier
        // ... etc
    }
    return stats
}
```

### Slot Validation

```go
func CanInstallEquipment(ship *Ship, equipment *Equipment) error {
    switch equipment.SlotType {
    case SlotTypeWeapon:
        if ship.UsedWeaponSlots >= ship.MaxWeaponSlots {
            return ErrNoWeaponSlots
        }
    case SlotTypeOutfit:
        if ship.UsedOutfitSlots >= ship.MaxOutfitSlots {
            return ErrNoOutfitSlots
        }
    }

    if equipment.RequiredTechLevel > ship.CurrentSystem.TechLevel {
        return ErrTechLevelTooLow
    }

    return nil
}
```

### Testing

- Ship purchase validation tests
- Equipment installation slot tests
- Trade-in value calculation tests
- Loadout save/load tests
- Fleet management tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
