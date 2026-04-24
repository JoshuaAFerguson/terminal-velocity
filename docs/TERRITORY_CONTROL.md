# Territory Control System

**Feature**: Territory Control
**Phase**: 11
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Territory Control system allows player factions to claim star systems and generate passive income from their controlled territories. This feature adds strategic depth to faction gameplay by creating competition for valuable systems and providing economic benefits to successful factions.

### Key Features

- **System Claiming**: Factions can claim unclaimed star systems
- **Passive Income**: Controlled territories generate credits over time
- **Control Timers**: Prevents rapid claim flipping with cooldown periods
- **Territory Visualization**: See faction-controlled systems at a glance
- **Treasury Integration**: Income flows directly to faction treasury
- **Strategic Value**: Systems have different economic values based on tech level and location

---

## Architecture

### Components

The territory control system consists of the following components:

1. **Territory Manager** (`internal/territory/manager.go`)
   - Manages territory claims and ownership
   - Handles passive income generation
   - Enforces claim rules and timers
   - Thread-safe with `sync.RWMutex`

2. **Territory UI** (`internal/tui/territory.go`)
   - Displays controlled systems
   - Shows available systems for claiming
   - Territory management interface
   - Income visualization

3. **Data Models** (`internal/models/`)
   - `Territory`: System ownership and claim data
   - Integration with `Faction` model
   - Income calculation parameters

### Data Flow

```
Player Action (Claim System)
         ↓
Territory Manager (Validate Claim)
         ↓
Check Faction Resources
         ↓
Update Territory Ownership
         ↓
Start Passive Income Timer
         ↓
Faction Treasury (Add Income)
         ↓
Update UI Display
```

### Thread Safety

The territory manager uses `sync.RWMutex` to ensure thread-safe operations:

- **Read Operations**: Multiple concurrent reads allowed
- **Write Operations**: Exclusive access for claims and updates
- **Background Workers**: Separate goroutine for income generation

---

## Implementation Details

### Territory Manager

The manager handles all territory-related operations:

```go
type Manager struct {
    mu          sync.RWMutex
    territories map[uuid.UUID]*models.Territory // SystemID -> Territory
    claims      map[uuid.UUID]uuid.UUID         // SystemID -> FactionID

    // Passive income
    incomeInterval time.Duration
    incomeRate     float64

    // Control timers
    claimCooldown  time.Duration

    // Repositories
    systemRepo *database.SystemRepository
    factionManager *factions.Manager
}
```

### Territory Claiming

**Requirements**:
- Faction must have sufficient credits for claim cost
- System must not be already claimed
- System must not be in cooldown period
- Player must have faction officer permissions

**Claim Process**:
1. Validate faction ownership
2. Check claim cost vs faction treasury
3. Verify system availability
4. Deduct claim cost from treasury
5. Create territory record
6. Start passive income generation

### Passive Income Generation

**Income Calculation**:
```go
baseIncome := system.TechLevel * incomeRate
territoryIncome := baseIncome * (1.0 + bonusMultiplier)
```

**Income Factors**:
- **Tech Level**: Higher tech systems generate more income
- **System Type**: Industrial systems have income bonuses
- **Faction Bonuses**: Faction perks can increase income
- **Control Duration**: Long-term control may provide bonuses

**Income Distribution**:
- Income generated every 30 minutes (configurable)
- Credits added directly to faction treasury
- Transaction logged for audit trail
- Notifications sent to faction leaders

### Control Timers

**Claim Cooldown**: 24 hours after losing control
- Prevents immediate reclaiming
- Allows time for faction warfare
- Configurable duration

**Grace Period**: 1 hour after claim
- New claims cannot be challenged immediately
- Allows faction to consolidate control

### Territory Values

Systems have different strategic values:

| Tech Level | Base Income/Hour | Strategic Value |
|------------|------------------|-----------------|
| 1-2        | 100 CR           | Low             |
| 3-4        | 250 CR           | Medium          |
| 5-6        | 500 CR           | High            |
| 7+         | 1000 CR          | Critical        |

**Special Bonuses**:
- **Core Systems**: +50% income (near galactic center)
- **Industrial Hubs**: +25% income
- **Trade Routes**: +20% income per connected high-value system

---

## User Interface

### Territory Screen

The territory UI provides comprehensive management tools:

**Main View**:
- List of controlled territories
- Current passive income rate
- Available systems for claiming
- Territory statistics

**Territory List Display**:
```
=== Controlled Territories ===

System               Tech  Income/Hr  Controlled Since
────────────────────────────────────────────────────
Alpha Centauri       7     1,000 CR   2 days ago
Tau Ceti             5     500 CR     5 hours ago
Barnard's Star       3     250 CR     1 day ago

Total Income: 1,750 CR/hour
Total Value: 42,000 CR/day
```

**Available Systems**:
```
=== Available for Claim ===

System               Tech  Claim Cost  Est. Income
──────────────────────────────────────────────────
Proxima Centauri     6     5,000 CR    750 CR/hr
Wolf 359             4     2,500 CR    300 CR/hr
```

### Navigation

- **↑/↓**: Navigate territory list
- **C**: Claim selected system
- **V**: View system details
- **R**: Release territory (return to pool)
- **ESC**: Return to factions screen

---

## Integration with Other Systems

### Faction System Integration

Territory control is tightly integrated with the faction system:

**Faction Requirements**:
- Must be faction leader or officer to claim
- Treasury must have sufficient funds
- Faction must be active (minimum 3 members)

**Faction Benefits**:
- Passive income increases faction power
- Territory count affects faction ranking
- Strategic systems provide military advantages

### Combat Integration

Territories can be contested through PvP combat:

**Territory Wars**:
- Faction vs faction battles
- Winner can claim contested territory
- Defenders have advantage in their territory
- Combat rewards scaled by territory value

### Economy Integration

**Market Impact**:
- Controlling trade hub systems affects market prices
- Territory income contributes to economic activity
- Faction wealth enables larger operations

**Resource Flow**:
```
Territory Income → Faction Treasury → Member Benefits
                                   → Faction Projects
                                   → War Funds
```

---

## Testing

### Unit Tests

Test coverage for territory manager:

```go
func TestTerritory_ClaimSystem(t *testing.T)
func TestTerritory_PassiveIncome(t *testing.T)
func TestTerritory_ClaimCooldown(t *testing.T)
func TestTerritory_TerritoryRelease(t *testing.T)
func TestTerritory_ThreadSafety(t *testing.T)
```

### Integration Tests

Full workflow testing:

1. **Claim Flow**:
   - Create faction
   - Fund treasury
   - Claim system
   - Verify ownership
   - Check income generation

2. **Contest Flow**:
   - Two factions
   - Claim same system
   - Resolve through combat
   - Verify claim transfer

3. **Income Flow**:
   - Claim multiple systems
   - Wait for income cycle
   - Verify treasury increase
   - Check income calculations

### Performance Tests

- **Concurrent Claims**: 100+ simultaneous claim attempts
- **Large Territory Counts**: 1000+ territories managed
- **Income Generation**: Performance under load
- **Memory Usage**: Monitor for leaks in long-running tests

---

## Configuration

### Manager Configuration

Configure territory system behavior:

```go
cfg := &territory.Config{
    // Income settings
    IncomeInterval:    30 * time.Minute,
    BaseIncomeRate:    100.0,

    // Claim settings
    ClaimCooldown:     24 * time.Hour,
    GracePeriod:       1 * time.Hour,

    // Costs
    ClaimCostBase:     1000,
    ClaimCostPerTech:  500,
}
```

### Economic Balancing

**Claim Costs**:
- Base: 1,000 CR
- Per Tech Level: +500 CR
- Formula: `1000 + (techLevel * 500)`

**Income Rates**:
- Configured to provide sustainable faction income
- Balanced against mission rewards
- Prevents territory from being only income source

---

## Troubleshooting

### Common Issues

**Problem**: Cannot claim system
**Solutions**:
- Check faction treasury balance
- Verify faction permissions
- Ensure system not already claimed
- Check if system in cooldown period

**Problem**: No passive income received
**Solutions**:
- Verify territory ownership
- Check income generation timer
- Ensure faction treasury not capped
- Review manager logs for errors

**Problem**: Territory disappeared
**Solutions**:
- Check if faction was disbanded
- Verify no database errors
- Review claim expiration rules
- Check for administrative actions

### Debug Commands

```bash
# Check territory status
curl http://localhost:8080/stats | grep territory

# View faction territories
SELECT * FROM territories WHERE faction_id = '<faction-id>';

# Check income generation
grep "territory income" /var/log/terminal-velocity/server.log
```

---

## Future Enhancements

### Planned Features

1. **Territory Upgrades**
   - Build defensive structures
   - Economic improvements
   - Research facilities
   - Requires investment of credits

2. **Territory Decay**
   - Unmaintained territories lose value
   - Requires periodic investment
   - Abandoned territories return to neutral

3. **Siege Mechanics**
   - Multi-day faction conflicts
   - Defenders get reinforcements
   - Attackers must maintain presence

4. **Strategic Resources**
   - Rare resources in certain systems
   - Required for advanced crafting
   - Trade commodity production

5. **Territory Bonuses**
   - Ship building speed
   - Research bonuses
   - Recruitment advantages
   - Market price modifiers

### Community Requests

- [ ] Territory voting for democratic factions
- [ ] Alliance territory sharing
- [ ] Territory trading between factions
- [ ] Visual territory maps
- [ ] Historical territory ownership tracking

---

## API Reference

### Core Functions

#### ClaimSystem

```go
func (m *Manager) ClaimSystem(
    factionID uuid.UUID,
    systemID uuid.UUID,
    claimedBy uuid.UUID,
) error
```

Claims a system for a faction.

**Parameters**:
- `factionID`: UUID of claiming faction
- `systemID`: UUID of system to claim
- `claimedBy`: UUID of player making claim

**Returns**: Error if claim fails

**Errors**:
- `ErrInsufficientFunds`: Not enough credits in treasury
- `ErrSystemClaimed`: System already owned
- `ErrCooldownActive`: System in cooldown period
- `ErrNoPermission`: Player lacks permissions

#### ReleaseTerritory

```go
func (m *Manager) ReleaseTerritory(
    factionID uuid.UUID,
    systemID uuid.UUID,
) error
```

Releases a territory back to unclaimed status.

#### GetFactionTerritories

```go
func (m *Manager) GetFactionTerritories(
    factionID uuid.UUID,
) []*models.Territory
```

Returns all territories controlled by a faction.

#### GetTerritoryIncome

```go
func (m *Manager) GetTerritoryIncome(
    factionID uuid.UUID,
) (int64, error)
```

Calculates current passive income rate for faction.

#### GenerateIncome

```go
func (m *Manager) GenerateIncome()
```

Background worker that generates and distributes income.

---

## Related Documentation

- [Player Factions](./PLAYER_FACTIONS.md) - Faction system integration
- [PvP Combat](./PVP_COMBAT.md) - Territory warfare mechanics
- [Chat System](./CHAT_SYSTEM.md) - Faction communication
- [Economy System](./ECONOMY.md) - Economic integration

---

## File Locations

**Core Implementation**:
- `internal/territory/manager.go` - Territory manager implementation
- `internal/models/territory.go` - Territory data models

**User Interface**:
- `internal/tui/territory.go` - Territory UI screens

**Database**:
- `scripts/schema.sql` - Territory table schema

**Tests**:
- `internal/territory/manager_test.go` - Unit tests

**Documentation**:
- `docs/TERRITORY_CONTROL.md` - This file
- `CHANGELOG.md` - Version history
- `ROADMAP.md` - Phase 11 details

---

**For questions or issues with the territory system, see the troubleshooting section above or contact the development team.**
