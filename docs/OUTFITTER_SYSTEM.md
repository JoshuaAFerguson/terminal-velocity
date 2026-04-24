# Ship Outfitting System

**Feature**: Equipment and Loadout Management
**Phase**: 16
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Ship Outfitting system provides comprehensive equipment management, allowing players to customize their ships with weapons, shields, engines, and other equipment. The loadout system enables saving and loading ship configurations for quick switching between different setups.

### Key Features

- **6 Equipment Slot Types**: Weapons, shields, engines, sensors, cargo bays, special
- **16 Equipment Items**: Wide variety of equipment across all categories
- **Loadout System**: Save, load, and clone ship configurations
- **Enhanced UI**: Dedicated outfitter screen with filtering and comparison
- **Slot Limits**: Type-specific slot restrictions based on ship class
- **Requirement Checking**: Tech level and ship size validation
- **Real-Time Stats**: Live ship statistics as equipment changes
- **Equipment Comparison**: Side-by-side item comparison

---

## Architecture

### Components

The outfitting system consists of the following components:

1. **Outfitting Manager** (`internal/outfitting/manager.go`)
   - Equipment installation/removal
   - Loadout management
   - Requirement validation
   - Thread-safe with `sync.RWMutex`

2. **Enhanced Outfitter UI** (`internal/tui/outfitter_enhanced.go`)
   - Equipment browser with filtering
   - Equipment details panel
   - Loadout management interface
   - Ship stats visualization

3. **Data Models** (`internal/models/`)
   - `Equipment`: Equipment items
   - `EquipmentSlot`: Ship equipment slots
   - `Loadout`: Saved ship configurations

### Data Flow

```
Player Selects Equipment
         ↓
Validate Requirements
         ↓
Check Slot Availability
         ↓
Deduct Credits
         ↓
Install Equipment
         ↓
Update Ship Stats
         ↓
Refresh UI Display
```

### Thread Safety

The outfitting manager uses `sync.RWMutex` for concurrent operations:

- **Read Operations**: Concurrent reads for equipment browsing
- **Write Operations**: Exclusive access for installations
- **Loadout Management**: Atomic save/load operations

---

## Implementation Details

### Outfitting Manager

The manager handles all equipment operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Available equipment
    equipment map[string]*Equipment // ItemID -> Equipment

    // Player loadouts
    loadouts  map[uuid.UUID]map[string]*Loadout // PlayerID -> LoadoutName -> Loadout

    // Repositories
    shipRepo     *database.ShipRepository
    inventoryMgr *inventory.Manager
}
```

### Equipment Types and Slots

**Slot Types**:
```go
const (
    SlotWeapon  = "weapon"
    SlotShield  = "shield"
    SlotEngine  = "engine"
    SlotSensor  = "sensor"
    SlotCargo   = "cargo"
    SlotSpecial = "special"
)
```

**Equipment Categories**:

1. **Weapons** (9 types):
   - Laser Cannon (Basic, MK2, MK3)
   - Missile Launcher (Basic, Advanced, Heavy)
   - Plasma Gun, Ion Cannon, Railgun

2. **Shields** (3 types):
   - Energy Shield (Basic, MK2, MK3)

3. **Engines** (2 types):
   - Hyperspace Drive, Advanced Drive

4. **Sensors** (1 type):
   - Long Range Scanner

5. **Cargo** (1 type):
   - Cargo Expansion

**Equipment Structure**:
```go
type Equipment struct {
    ItemID      string
    Name        string
    Type        EquipmentType
    Slot        SlotType
    TechLevel   int
    MinShipSize int

    // Stats
    Damage      int
    Accuracy    int
    Range       int
    ShieldBoost int
    SpeedBoost  int
    CargoBoost  int

    // Cost
    Price       int64
    Mass        int

    Description string
}
```

### Equipment Installation

**Installation Process**:
```go
func (m *Manager) InstallEquipment(
    playerID uuid.UUID,
    shipID uuid.UUID,
    equipmentID string,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 1. Get ship and equipment
    ship := m.shipRepo.GetShip(shipID)
    equipment := m.equipment[equipmentID]

    // 2. Validate requirements
    if err := m.validateRequirements(ship, equipment); err != nil {
        return err
    }

    // 3. Check slot availability
    if err := m.checkSlotAvailability(ship, equipment.Slot); err != nil {
        return err
    }

    // 4. Deduct credits
    player := m.playerRepo.GetPlayer(playerID)
    if player.Credits < equipment.Price {
        return ErrInsufficientCredits
    }
    player.Credits -= equipment.Price

    // 5. Install equipment
    slot := m.findFreeSlot(ship, equipment.Slot)
    slot.EquipmentID = equipmentID
    slot.Installed = true

    // 6. Update ship stats
    m.recalculateShipStats(ship)

    // 7. Save changes
    m.shipRepo.UpdateShip(ship)
    m.playerRepo.UpdatePlayer(player)

    return nil
}
```

**Requirement Validation**:
```go
func (m *Manager) validateRequirements(
    ship *Ship,
    equipment *Equipment,
) error {
    // Check tech level
    if ship.TechLevel < equipment.TechLevel {
        return ErrTechLevelTooLow
    }

    // Check ship size
    if ship.Size < equipment.MinShipSize {
        return ErrShipTooSmall
    }

    // Check mass capacity
    currentMass := m.calculateTotalMass(ship)
    if currentMass + equipment.Mass > ship.MaxMass {
        return ErrExceedsMassLimit
    }

    return nil
}
```

### Loadout System

**Loadout Structure**:
```go
type Loadout struct {
    LoadoutID   uuid.UUID
    PlayerID    uuid.UUID
    ShipID      uuid.UUID
    Name        string

    // Equipment configuration
    Weapons     []string // EquipmentIDs
    Shields     []string
    Engines     []string
    Sensors     []string
    Cargo       []string
    Special     []string

    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Save Loadout**:
```go
func (m *Manager) SaveLoadout(
    playerID uuid.UUID,
    shipID uuid.UUID,
    name string,
) (*Loadout, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Get current ship configuration
    ship := m.shipRepo.GetShip(shipID)

    // Create loadout from current equipment
    loadout := &Loadout{
        LoadoutID: uuid.New(),
        PlayerID:  playerID,
        ShipID:    shipID,
        Name:      name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Extract equipment by slot type
    for _, slot := range ship.EquipmentSlots {
        if !slot.Installed {
            continue
        }

        switch slot.SlotType {
        case SlotWeapon:
            loadout.Weapons = append(loadout.Weapons, slot.EquipmentID)
        case SlotShield:
            loadout.Shields = append(loadout.Shields, slot.EquipmentID)
        // ... other slot types
        }
    }

    // Store loadout
    if m.loadouts[playerID] == nil {
        m.loadouts[playerID] = make(map[string]*Loadout)
    }
    m.loadouts[playerID][name] = loadout

    return loadout, nil
}
```

**Load Loadout**:
```go
func (m *Manager) LoadLoadout(
    playerID uuid.UUID,
    shipID uuid.UUID,
    loadoutName string,
) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Get loadout
    loadout := m.loadouts[playerID][loadoutName]
    if loadout == nil {
        return ErrLoadoutNotFound
    }

    ship := m.shipRepo.GetShip(shipID)

    // Uninstall all current equipment
    for _, slot := range ship.EquipmentSlots {
        slot.Installed = false
        slot.EquipmentID = ""
    }

    // Install loadout equipment
    allEquipment := []string{}
    allEquipment = append(allEquipment, loadout.Weapons...)
    allEquipment = append(allEquipment, loadout.Shields...)
    // ... other slot types

    for _, equipID := range allEquipment {
        equipment := m.equipment[equipID]
        slot := m.findFreeSlot(ship, equipment.Slot)
        slot.EquipmentID = equipID
        slot.Installed = true
    }

    // Recalculate ship stats
    m.recalculateShipStats(ship)
    m.shipRepo.UpdateShip(ship)

    return nil
}
```

**Clone Loadout**:
```go
func (m *Manager) CloneLoadout(
    playerID uuid.UUID,
    sourceLoadout string,
    newName string,
) (*Loadout, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    source := m.loadouts[playerID][sourceLoadout]
    if source == nil {
        return nil, ErrLoadoutNotFound
    }

    clone := &Loadout{
        LoadoutID: uuid.New(),
        PlayerID:  playerID,
        ShipID:    source.ShipID,
        Name:      newName,
        Weapons:   append([]string{}, source.Weapons...),
        Shields:   append([]string{}, source.Shields...),
        // ... copy other slots
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    m.loadouts[playerID][newName] = clone
    return clone, nil
}
```

### Ship Stats Calculation

**Recalculate Stats**:
```go
func (m *Manager) recalculateShipStats(ship *Ship) {
    // Reset to base stats
    ship.TotalDamage = ship.BaseDamage
    ship.TotalShields = ship.BaseShields
    ship.Speed = ship.BaseSpeed
    ship.CargoCapacity = ship.BaseCargoCapacity
    ship.SensorRange = ship.BaseSensorRange

    // Apply equipment bonuses
    for _, slot := range ship.EquipmentSlots {
        if !slot.Installed {
            continue
        }

        equipment := m.equipment[slot.EquipmentID]
        ship.TotalDamage += equipment.Damage
        ship.TotalShields += equipment.ShieldBoost
        ship.Speed += equipment.SpeedBoost
        ship.CargoCapacity += equipment.CargoBoost
        ship.SensorRange += equipment.Range
    }
}
```

---

## User Interface

### Enhanced Outfitter Screen

**Main View**:
```
═══════════════════════════════════════════════════════
              SHIP OUTFITTER - ENHANCED
═══════════════════════════════════════════════════════

Ship: Destroyer                   Credits: 125,000 CR

┌─ Equipment Browser ─────┐  ┌─ Ship Stats ──────────┐
│ [Weapons▼] [All Techs▼] │  │ Hull:      1000/1000  │
│                          │  │ Shields:   500/500    │
│ > Laser Cannon MK3       │  │ Speed:     75         │
│   Missile Launcher       │  │ Cargo:     200 tons   │
│   Plasma Gun             │  │ Sensors:   100 AU     │
│   Ion Cannon             │  │                       │
│   Railgun                │  │ Damage Output: 450    │
└──────────────────────────┘  └───────────────────────┘

┌─ Equipment Details ──────────────────────────────────┐
│ LASER CANNON MK3                                     │
│ ──────────────────────────────────────────────────   │
│ Type:     Weapon              Tech Level: 5          │
│ Damage:   120                 Accuracy:   95%        │
│ Range:    Medium              Price: 25,000 CR       │
│                                                       │
│ A powerful energy weapon with high accuracy and      │
│ sustained fire capability.                           │
│                                                       │
│ Requirements:                                        │
│   ✓ Tech Level 5              ✓ Ship Size: Medium+  │
│   ✓ Available Slots: 2/4                            │
└──────────────────────────────────────────────────────┘

[I]nstall [U]ninstall [L]oadouts [ESC]Back
```

**Loadout Management**:
```
=== Ship Loadouts ===

Saved Loadouts (3):
  > Combat Heavy
    Trade Runner
    Explorer

Current Loadout: Combat Heavy
  Weapons:  Laser Cannon MK3 x2, Missile Launcher
  Shields:  Energy Shield MK2
  Engines:  Advanced Drive
  Sensors:  Long Range Scanner
  Cargo:    -
  Special:  -

[L]oad [S]ave New [C]lone [D]elete [R]ename [ESC]Back
```

### Navigation

- **↑/↓**: Navigate equipment list
- **Tab**: Switch between panels
- **F**: Filter by type/tech level
- **I**: Install selected equipment
- **U**: Uninstall equipment
- **L**: Open loadout manager
- **C**: Compare equipment
- **ESC**: Return to main menu

---

## Integration with Other Systems

### Combat Integration

Equipment directly affects combat performance:

```go
// In combat system
weaponDamage := ship.TotalDamage
hitChance := calculateAccuracy(ship.Sensors, target.Speed)
shieldAbsorption := ship.TotalShields
```

### Trading Integration

Cargo expansion affects trading capacity:

```go
maxCargo := ship.BaseCargoCapacity
for _, expansion := range ship.CargoExpansions {
    maxCargo += expansion.CargoBoost
}
```

---

## Testing

### Unit Tests

```go
func TestOutfitting_Install(t *testing.T)
func TestOutfitting_Uninstall(t *testing.T)
func TestOutfitting_Requirements(t *testing.T)
func TestOutfitting_Loadout_Save(t *testing.T)
func TestOutfitting_Loadout_Load(t *testing.T)
func TestOutfitting_StatsCalculation(t *testing.T)
```

---

## Configuration

```go
cfg := &outfitting.Config{
    // Slot limits by ship class
    ShuttleWeaponSlots:    2,
    FrigateWeaponSlots:    4,
    DestroyerWeaponSlots:  6,

    // Equipment restrictions
    EnforceTechLevel:      true,
    EnforceShipSize:       true,
    EnforceMassLimits:     true,

    // Loadout settings
    MaxLoadoutsPerPlayer:  10,
    LoadoutNameMaxLength:  32,
}
```

---

## API Reference

### Core Functions

#### InstallEquipment

```go
func (m *Manager) InstallEquipment(
    playerID uuid.UUID,
    shipID uuid.UUID,
    equipmentID string,
) error
```

Installs equipment on a ship.

#### UninstallEquipment

```go
func (m *Manager) UninstallEquipment(
    shipID uuid.UUID,
    slotIndex int,
) error
```

Removes equipment from a ship.

#### SaveLoadout

```go
func (m *Manager) SaveLoadout(
    playerID uuid.UUID,
    shipID uuid.UUID,
    name string,
) (*Loadout, error)
```

Saves current ship configuration as a loadout.

#### LoadLoadout

```go
func (m *Manager) LoadLoadout(
    playerID uuid.UUID,
    shipID uuid.UUID,
    loadoutName string,
) error
```

Applies a saved loadout to a ship.

---

## Related Documentation

- [Combat System](./COMBAT.md) - Equipment effects in combat
- [Ship Management](./SHIPS.md) - Ship stats and classes
- [Trading System](./TRADING.md) - Cargo equipment

---

## File Locations

**Core Implementation**:
- `internal/outfitting/manager.go` - Outfitting manager
- `internal/models/equipment.go` - Equipment data models

**User Interface**:
- `internal/tui/outfitter_enhanced.go` - Enhanced UI
- `internal/tui/outfitter.go` - Basic outfitter

**Documentation**:
- `docs/OUTFITTER_SYSTEM.md` - This file
- `ROADMAP.md` - Phase 16 details

---

**For questions about the outfitting system, contact the development team.**
