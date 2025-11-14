# API Integration Plan for Enhanced TUI Screens

**Created**: 2025-01-14
**Status**: Planning Phase
**Purpose**: Define strategy for connecting enhanced TUI screens to backend game logic

---

## Overview

The enhanced TUI screens currently use **sample/mock data** for display. This document outlines the integration strategy to connect these screens to the actual game backend through the existing repository and manager layers.

### Current Architecture

```
TUI Screens (Presentation)
     ↓
Model (State Management)
     ↓
??? (Need to add API layer)
     ↓
Repositories (Data Access)
     ↓
Database (PostgreSQL)
```

### Target Architecture

```
TUI Screens (Presentation)
     ↓
Model (State + Commands)
     ↓
BubbleTea Commands (Async Operations)
     ↓
Repositories + Managers (Business Logic)
     ↓
Database (PostgreSQL)
```

---

## Integration Patterns

### Pattern 1: BubbleTea Command Pattern

All async operations should use BubbleTea's `tea.Cmd` pattern:

```go
// Define message types for async results
type dataLoadedMsg struct {
    data *Data
    err  error
}

// Create command function
func loadDataCmd() tea.Msg {
    data, err := repository.LoadData(context.Background())
    return dataLoadedMsg{data: data, err: err}
}

// In Update(), return command
case "enter":
    return m, loadDataCmd

// Handle result in Update()
case dataLoadedMsg:
    if msg.err != nil {
        m.error = msg.err
        return m, nil
    }
    m.data = msg.data
    return m, nil
```

### Pattern 2: Repository Access

Repositories are already available in `Model`:
- `playerRepo *database.PlayerRepository`
- `systemRepo *database.SystemRepository`
- `sshKeyRepo *database.SSHKeyRepository`

**Need to add**:
- `shipRepo *database.ShipRepository`
- `marketRepo *database.MarketRepository`
- Access to managers (combat, missions, quests, etc.)

### Pattern 3: Error Handling

All API calls should handle errors gracefully:

```go
case apiResultMsg:
    if msg.err != nil {
        m.error = fmt.Sprintf("Operation failed: %v", msg.err)
        m.showErrorDialog = true
        return m, nil
    }
    // Success path
```

---

## Screen-by-Screen Integration

### 1. Trading Enhanced ✅ Ready

**Current State**: Mock data (15 commodities)

**Required Integration**:
- Load market data from `MarketRepository`
- Implement buy/sell transactions
- Update player credits and cargo
- Refresh market prices after transactions

**API Calls Needed**:
```go
// Load market data
func (m Model) loadMarketDataCmd() tea.Msg {
    ctx := context.Background()

    // Get current planet's market
    planetID := m.player.CurrentPlanet
    commodities, err := m.marketRepo.GetMarketPrices(ctx, planetID)
    if err != nil {
        return marketLoadedMsg{err: err}
    }

    // Get player's cargo
    cargo, err := m.playerRepo.GetCargo(ctx, m.playerID)
    if err != nil {
        return marketLoadedMsg{err: err}
    }

    return marketLoadedMsg{
        commodities: commodities,
        cargo: cargo,
        err: nil,
    }
}

// Buy commodity
func (m Model) buyCommodityCmd(commodityID string, quantity int) tea.Msg {
    ctx := context.Background()
    err := m.marketRepo.BuyCommodity(ctx, m.playerID, commodityID, quantity)
    return transactionCompleteMsg{err: err, action: "buy"}
}

// Sell commodity
func (m Model) sellCommodityCmd(commodityID string, quantity int) tea.Msg {
    ctx := context.Background()
    err := m.marketRepo.SellCommodity(ctx, m.playerID, commodityID, quantity)
    return transactionCompleteMsg{err: err, action: "sell"}
}
```

**Message Types**:
```go
type marketLoadedMsg struct {
    commodities []models.Commodity
    cargo       []models.CargoItem
    err         error
}

type transactionCompleteMsg struct {
    action string // "buy" or "sell"
    err    error
}
```

**TODO Comments**: Lines 202-203, 207-208 in trading_enhanced.go

---

### 2. Shipyard Enhanced ✅ Ready

**Current State**: Mock data (7 ships)

**Required Integration**:
- Load available ships from database
- Get current ship details
- Implement ship purchase/trade-in
- Calculate trade-in values

**API Calls Needed**:
```go
// Load shipyard inventory
func (m Model) loadShipyardCmd() tea.Msg {
    ctx := context.Background()

    // Get available ships at this planet
    planetID := m.player.CurrentPlanet
    ships, err := m.shipRepo.GetShipsForSale(ctx, planetID)
    if err != nil {
        return shipyardLoadedMsg{err: err}
    }

    // Get player's current ship
    currentShip, err := m.shipRepo.GetShip(ctx, m.currentShip.ID)
    if err != nil {
        return shipyardLoadedMsg{err: err}
    }

    return shipyardLoadedMsg{
        ships: ships,
        currentShip: currentShip,
        err: nil,
    }
}

// Purchase/trade ship
func (m Model) purchaseShipCmd(shipTypeID string) tea.Msg {
    ctx := context.Background()

    // Calculate trade-in value
    tradeInValue := calculateTradeInValue(m.currentShip)

    // Execute purchase
    newShip, err := m.shipRepo.PurchaseShip(ctx, m.playerID, shipTypeID, tradeInValue)
    return shipPurchasedMsg{ship: newShip, err: err}
}
```

**Message Types**:
```go
type shipyardLoadedMsg struct {
    ships       []models.ShipType
    currentShip *models.Ship
    err         error
}

type shipPurchasedMsg struct {
    ship *models.Ship
    err  error
}
```

**TODO Comments**: Lines 162-163 in shipyard_enhanced.go

---

### 3. Mission Board Enhanced ✅ Ready

**Current State**: Mock data (5 missions)

**Required Integration**:
- Load available missions from MissionManager
- Implement mission accept/decline
- Track active missions
- Update mission progress

**API Calls Needed**:
```go
// Load missions
func (m Model) loadMissionsCmd() tea.Msg {
    ctx := context.Background()

    // Get available missions at current location
    missions, err := m.missionManager.GetAvailableMissions(ctx, m.player.CurrentPlanet)
    if err != nil {
        return missionsLoadedMsg{err: err}
    }

    // Get player's active missions
    activeMissions, err := m.missionManager.GetActiveMissions(ctx, m.playerID)
    if err != nil {
        return missionsLoadedMsg{err: err}
    }

    return missionsLoadedMsg{
        available: missions,
        active: activeMissions,
        err: nil,
    }
}

// Accept mission
func (m Model) acceptMissionCmd(missionID uuid.UUID) tea.Msg {
    ctx := context.Background()
    err := m.missionManager.AcceptMission(ctx, m.playerID, missionID)
    return missionActionMsg{action: "accept", err: err}
}

// Decline mission
func (m Model) declineMissionCmd(missionID uuid.UUID) tea.Msg {
    ctx := context.Background()
    err := m.missionManager.DeclineMission(ctx, missionID)
    return missionActionMsg{action: "decline", err: err}
}
```

**Message Types**:
```go
type missionsLoadedMsg struct {
    available []models.Mission
    active    []models.Mission
    err       error
}

type missionActionMsg struct {
    action string // "accept" or "decline"
    err    error
}
```

**TODO Comments**: Lines 260-261, 264-266 in mission_board_enhanced.go

---

### 4. Quest Board Enhanced ✅ Ready

**Current State**: Mock data (2 active, 3 available quests)

**Required Integration**:
- Load quests from QuestManager
- Track quest progress
- Implement quest accept/abandon
- Update objectives

**API Calls Needed**:
```go
// Load quests
func (m Model) loadQuestsCmd() tea.Msg {
    ctx := context.Background()

    quests, err := m.questManager.GetPlayerQuests(ctx, m.playerID)
    if err != nil {
        return questsLoadedMsg{err: err}
    }

    available, err := m.questManager.GetAvailableQuests(ctx, m.playerID)
    if err != nil {
        return questsLoadedMsg{err: err}
    }

    return questsLoadedMsg{
        active: quests,
        available: available,
        err: nil,
    }
}

// Abandon quest
func (m Model) abandonQuestCmd(questID uuid.UUID) tea.Msg {
    ctx := context.Background()
    err := m.questManager.AbandonQuest(ctx, m.playerID, questID)
    return questActionMsg{action: "abandon", err: err}
}
```

**Message Types**:
```go
type questsLoadedMsg struct {
    active    []models.Quest
    available []models.Quest
    err       error
}

type questActionMsg struct {
    action string
    err    error
}
```

**TODO Comments**: Lines 334-336 in quest_board_enhanced.go

---

### 5. Navigation Enhanced ✅ Ready

**Current State**: Mock data (4 systems)

**Required Integration**:
- Load nearby systems from SystemRepository
- Calculate jump routes
- Implement hyperdrive jump
- Validate fuel requirements

**API Calls Needed**:
```go
// Load nearby systems
func (m Model) loadNearbySystems Cmd() tea.Msg {
    ctx := context.Background()

    // Get current system
    currentSystem, err := m.systemRepo.GetSystem(ctx, m.player.CurrentSystem)
    if err != nil {
        return systemsLoadedMsg{err: err}
    }

    // Get connected systems (via jump routes)
    connections, err := m.systemRepo.GetConnectedSystems(ctx, m.player.CurrentSystem)
    if err != nil {
        return systemsLoadedMsg{err: err}
    }

    return systemsLoadedMsg{
        current: currentSystem,
        nearby: connections,
        err: nil,
    }
}

// Execute jump
func (m Model) executeJumpCmd(targetSystemID uuid.UUID) tea.Msg {
    ctx := context.Background()

    // Check fuel requirements
    fuelNeeded := calculateFuelNeeded(m.player.CurrentSystem, targetSystemID)
    if m.currentShip.Fuel < fuelNeeded {
        return jumpCompleteMsg{err: errors.New("insufficient fuel")}
    }

    // Execute jump (updates player location, consumes fuel)
    err := m.playerRepo.UpdateLocation(ctx, m.playerID, targetSystemID)
    if err != nil {
        return jumpCompleteMsg{err: err}
    }

    // Update ship fuel
    err = m.shipRepo.ConsumeFuel(ctx, m.currentShip.ID, fuelNeeded)
    return jumpCompleteMsg{err: err}
}
```

**Message Types**:
```go
type systemsLoadedMsg struct {
    current *models.StarSystem
    nearby  []models.StarSystem
    err     error
}

type jumpCompleteMsg struct {
    err error
}
```

**TODO Comments**: Lines 293-296, 299-302 in navigation_enhanced.go

---

### 6. Combat Enhanced ⚠️ Complex

**Current State**: Mock combat scenario

**Required Integration**:
- Initialize combat from encounter
- Implement weapon firing via CombatManager
- Handle turn-based combat flow
- Process combat results (damage, victory, defeat)
- Award loot/experience

**API Calls Needed**:
```go
// Initialize combat encounter
func (m Model) initCombatCmd(enemyShipID uuid.UUID) tea.Msg {
    ctx := context.Background()

    // Create combat session
    combat, err := m.combatManager.StartCombat(ctx, m.playerID, enemyShipID)
    if err != nil {
        return combatInitMsg{err: err}
    }

    return combatInitMsg{
        combat: combat,
        err: nil,
    }
}

// Fire weapon
func (m Model) fireWeaponCmd(weaponSlot int) tea.Msg {
    ctx := context.Background()

    result, err := m.combatManager.FireWeapon(ctx, m.playerID, weaponSlot)
    if err != nil {
        return combatActionMsg{err: err}
    }

    return combatActionMsg{
        result: result,
        err: nil,
    }
}

// Execute enemy turn
func (m Model) enemyTurnCmd() tea.Msg {
    ctx := context.Background()

    result, err := m.combatManager.ExecuteEnemyTurn(ctx)
    return combatActionMsg{result: result, err: err}
}

// End combat
func (m Model) endCombatCmd(victory bool) tea.Msg {
    ctx := context.Background()

    rewards, err := m.combatManager.EndCombat(ctx, m.playerID, victory)
    return combatEndMsg{victory: victory, rewards: rewards, err: err}
}
```

**Message Types**:
```go
type combatInitMsg struct {
    combat *models.Combat
    err    error
}

type combatActionMsg struct {
    result *models.CombatResult
    err    error
}

type combatEndMsg struct {
    victory bool
    rewards *models.CombatRewards
    err     error
}
```

**TODO Comments**: Lines 393-413 in combat_enhanced.go

---

### 7. Outfitter Enhanced ⚠️ Complex

**Current State**: Mock equipment data

**Required Integration**:
- Load equipment from OutfittingManager
- Implement equipment purchase
- Handle loadout save/load
- Validate equipment compatibility

**API Calls Needed**:
```go
// Load equipment
func (m Model) loadEquipmentCmd() tea.Msg {
    ctx := context.Background()

    equipment, err := m.outfittingManager.GetAvailableEquipment(ctx, m.player.CurrentPlanet)
    if err != nil {
        return equipmentLoadedMsg{err: err}
    }

    installed, err := m.outfittingManager.GetInstalledEquipment(ctx, m.currentShip.ID)
    if err != nil {
        return equipmentLoadedMsg{err: err}
    }

    return equipmentLoadedMsg{
        available: equipment,
        installed: installed,
        err: nil,
    }
}

// Purchase and install equipment
func (m Model) installEquipmentCmd(equipmentID uuid.UUID, slotIndex int) tea.Msg {
    ctx := context.Background()

    err := m.outfittingManager.InstallEquipment(ctx, m.currentShip.ID, equipmentID, slotIndex)
    return equipmentActionMsg{action: "install", err: err}
}

// Save loadout
func (m Model) saveLoadoutCmd(name string) tea.Msg {
    ctx := context.Background()

    err := m.outfittingManager.SaveLoadout(ctx, m.currentShip.ID, name)
    return loadoutActionMsg{action: "save", err: err}
}
```

**Message Types**:
```go
type equipmentLoadedMsg struct {
    available []models.Equipment
    installed []models.Equipment
    err       error
}

type equipmentActionMsg struct {
    action string
    err    error
}

type loadoutActionMsg struct {
    action string
    err    error
}
```

**TODO Comments**: Multiple in outfitter_enhanced.go

---

### 8. Space View ⚠️ Data Loading

**Current State**: Static viewport

**Required Integration**:
- Load current system objects (planets, ships)
- Load player ship stats
- Refresh shield/hull/fuel displays
- Load nearby targets

**API Calls Needed**:
```go
// Load space view data
func (m Model) loadSpaceViewCmd() tea.Msg {
    ctx := context.Background()

    // Get current system
    system, err := m.systemRepo.GetSystem(ctx, m.player.CurrentSystem)
    if err != nil {
        return spaceViewLoadedMsg{err: err}
    }

    // Get planets in system
    planets, err := m.systemRepo.GetPlanets(ctx, m.player.CurrentSystem)
    if err != nil {
        return spaceViewLoadedMsg{err: err}
    }

    // Get nearby ships (for encounter/combat)
    ships, err := m.encounterManager.GetNearbyShips(ctx, m.player.CurrentSystem)
    if err != nil {
        return spaceViewLoadedMsg{err: err}
    }

    // Get player ship current stats
    ship, err := m.shipRepo.GetShip(ctx, m.currentShip.ID)
    if err != nil {
        return spaceViewLoadedMsg{err: err}
    }

    return spaceViewLoadedMsg{
        system: system,
        planets: planets,
        ships: ships,
        playerShip: ship,
        err: nil,
    }
}
```

**Message Types**:
```go
type spaceViewLoadedMsg struct {
    system     *models.StarSystem
    planets    []*models.Planet
    ships      []models.Ship
    playerShip *models.Ship
    err        error
}
```

**TODO Comments**: Lines 68-75, 203-216 in space_view.go

---

### 9. Landing ✅ Simple

**Current State**: Static services menu

**Required Integration**:
- Load planet information
- Implement refuel service
- Implement repair service
- Check service availability

**API Calls Needed**:
```go
// Load planet data
func (m Model) loadPlanetCmd() tea.Msg {
    ctx := context.Background()

    planet, err := m.systemRepo.GetPlanet(ctx, m.player.CurrentPlanet)
    if err != nil {
        return planetLoadedMsg{err: err}
    }

    return planetLoadedMsg{planet: planet, err: nil}
}

// Refuel ship
func (m Model) refuelShipCmd() tea.Msg {
    ctx := context.Background()

    cost := calculateRefuelCost(m.currentShip)
    if m.player.Credits < cost {
        return serviceCompleteMsg{err: errors.New("insufficient credits")}
    }

    // Refuel ship
    err := m.shipRepo.Refuel(ctx, m.currentShip.ID)
    if err != nil {
        return serviceCompleteMsg{err: err}
    }

    // Deduct credits
    err = m.playerRepo.DeductCredits(ctx, m.playerID, cost)
    return serviceCompleteMsg{service: "refuel", err: err}
}

// Repair ship
func (m Model) repairShipCmd() tea.Msg {
    ctx := context.Background()

    cost := calculateRepairCost(m.currentShip)
    if m.player.Credits < cost {
        return serviceCompleteMsg{err: errors.New("insufficient credits")}
    }

    // Repair ship
    err := m.shipRepo.Repair(ctx, m.currentShip.ID)
    if err != nil {
        return serviceCompleteMsg{err: err}
    }

    // Deduct credits
    err = m.playerRepo.DeductCredits(ctx, m.playerID, cost)
    return serviceCompleteMsg{service: "repair", err: err}
}
```

**Message Types**:
```go
type planetLoadedMsg struct {
    planet *models.Planet
    err    error
}

type serviceCompleteMsg struct {
    service string
    err     error
}
```

**TODO Comments**: Lines 230-232, 235-237 in landing.go

---

## Implementation Priority

### Phase 1: Foundation (Week 1)
1. ✅ Add missing repositories to Model
2. ✅ Add missing managers to Model
3. ✅ Create message type definitions file
4. ✅ Create helper functions for common patterns

### Phase 2: Simple Screens (Week 1-2)
1. **Landing** - Refuel/Repair services
2. **Trading** - Buy/Sell commodities
3. **Navigation** - Load systems, execute jumps

### Phase 3: Medium Complexity (Week 2-3)
4. **Shipyard** - Load ships, purchase/trade
5. **Mission Board** - Load/accept/decline missions
6. **Quest Board** - Load/abandon quests

### Phase 4: Complex Screens (Week 3-4)
7. **Combat** - Full combat system integration
8. **Outfitter** - Equipment management
9. **Space View** - Dynamic object loading

### Phase 5: Polish & Error Handling (Week 4)
10. Add loading indicators
11. Add error dialogs
12. Add confirmation dialogs
13. Add success/failure feedback
14. Handle edge cases

---

## Required Model Changes

### Add Repositories

```go
// In model.go, add to Model struct:
shipRepo   *database.ShipRepository
marketRepo *database.MarketRepository

// In NewModel(), initialize:
shipRepo:   database.NewShipRepository(db),
marketRepo: database.NewMarketRepository(db),
```

### Add Manager Access

```go
// Many managers already exist:
- combatManager      *combat.Manager
- missionManager     *missions.Manager
- questManager       *quests.Manager
- encounterManager   *encounters.Manager
- outfittingManager  *outfitting.Manager

// Just need to ensure they're accessible in enhanced screens
```

### Add Message Types File

Create `internal/tui/messages.go`:
```go
package tui

import (
    "github.com/JoshuaAFerguson/terminal-velocity/internal/models"
    "github.com/google/uuid"
)

// Market messages
type marketLoadedMsg struct {
    commodities []models.Commodity
    cargo       []models.CargoItem
    err         error
}

type transactionCompleteMsg struct {
    action string
    err    error
}

// ... (all other message types)
```

---

## Error Handling Strategy

### Error Display

Add to Model:
```go
type Model struct {
    // ... existing fields

    // Error handling
    error           string
    showErrorDialog bool
}
```

### Error UI Component

```go
func (m Model) renderErrorDialog() string {
    if !m.showErrorDialog || m.error == "" {
        return ""
    }

    var sb strings.Builder
    sb.WriteString("┏━━━━━━━━━━━━━━━━━━━━━━━━┓\n")
    sb.WriteString("┃ ERROR                  ┃\n")
    sb.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━┫\n")
    sb.WriteString(fmt.Sprintf("┃ %s ┃\n", PadRight(m.error, 22)))
    sb.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━┫\n")
    sb.WriteString("┃ [Press any key]        ┃\n")
    sb.WriteString("┗━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
    return sb.String()
}
```

---

## Testing Strategy

### Unit Tests

Test each command function independently:
```go
func TestLoadMarketDataCmd(t *testing.T) {
    // Mock repository
    // Call command
    // Verify result message
}
```

### Integration Tests

Test full flow from key press to data display:
```go
func TestTradingBuyFlow(t *testing.T) {
    // Initialize model with test data
    // Simulate buy action
    // Verify credits deducted
    // Verify cargo updated
    // Verify UI reflects changes
}
```

### Manual Testing Checklist

- [ ] Load data on screen entry
- [ ] Handle empty/null data
- [ ] Handle API errors gracefully
- [ ] Show loading indicators
- [ ] Validate user input
- [ ] Confirm destructive actions
- [ ] Update UI after successful operations
- [ ] Prevent duplicate operations

---

## Performance Considerations

### Caching Strategy

- Cache market data for 1 minute
- Cache system data until jump
- Invalidate cache on transactions
- Pre-load next screens in background

### Optimization

- Batch related API calls
- Use goroutines for independent operations
- Implement pagination for large lists
- Lazy-load details on demand

---

## Next Steps

1. ✅ Review this plan with team
2. ⏳ Create `messages.go` with all message type definitions
3. ⏳ Add missing repositories to Model
4. ⏳ Implement Phase 1 (Foundation)
5. ⏳ Begin Phase 2 (Simple Screens)
6. ⏳ Create integration test suite

---

**Last Updated**: 2025-01-14
**Estimated Timeline**: 4 weeks for full integration
**Risk Level**: Medium (complexity in combat and outfitter)
