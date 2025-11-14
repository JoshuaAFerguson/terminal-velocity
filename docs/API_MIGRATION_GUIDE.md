# API Migration Guide

**Version**: 1.1.0
**Phase**: Phase 1 - Internal API Extraction
**Status**: Implementation In Progress
**Last Updated**: 2025-01-14

## Overview

This guide explains how to migrate existing TUI code from direct manager/repository access to using the new API client interface. This is the core work of Phase 1.

**📚 See Also**: [`PHASE_1B_EXAMPLE.md`](./PHASE_1B_EXAMPLE.md) for a complete real-world migration example with before/after code.

## Goals

- **Decouple** TUI from game logic implementation
- **Standardize** all game operations through a single API
- **Enable** future distributed architecture (Phase 2+)
- **Maintain** backward compatibility during migration

## Current Progress

### ✅ Completed
- API interface definitions (`internal/api/client.go`)
- API type system (`internal/api/types.go`)
- Server infrastructure (`internal/api/server/server.go`)
- Session management (`internal/api/server/session.go`)
- Data converters (`internal/api/server/converters.go`)
- Example handlers: `GetPlayerState`, `BuyCommodity`, `Jump`
- Proof-of-concept migration example

### 🔄 In Progress
- Implementing remaining 28 server handlers
- TUI screen migrations
- Unit tests for handlers

### ⏳ Planned
- Integration testing
- Performance benchmarking
- Documentation completion

## Architecture Comparison

### Before (Current Monolithic)

```go
// internal/tui/model.go
type Model struct {
    // Direct access to repositories
    playerRepo *database.PlayerRepository
    systemRepo *database.SystemRepository
    shipRepo   *database.ShipRepository
    marketRepo *database.MarketRepository

    // Direct access to managers
    missionsManager *missions.Manager
    questsManager   *quests.Manager
    chatManager     *chat.Manager
    // ... etc
}

// TUI directly calls repositories/managers
func (m *Model) buyShip(shipType string) error {
    // Direct database calls
    player, err := m.playerRepo.GetByID(ctx, m.playerID)
    if err != nil {
        return err
    }

    // Direct game logic
    if player.Credits < shipCost {
        return errors.New("insufficient credits")
    }

    // Multiple repository calls
    newShip, err := m.shipRepo.Create(ctx, ...)
    player.Credits -= shipCost
    err = m.playerRepo.Update(ctx, player)

    return nil
}
```

### After (API-Based)

```go
// internal/tui/model.go
type Model struct {
    // Single API client
    apiClient api.Client

    // TUI state (for rendering)
    playerState *api.PlayerState
    currentShip *api.Ship
}

// TUI calls API client
func (m *Model) buyShip(shipType string) error {
    req := &api.ShipPurchaseRequest{
        PlayerID: m.playerID,
        ShipType: shipType,
    }

    resp, err := m.apiClient.BuyShip(context.Background(), req)
    if err != nil {
        return err
    }

    if !resp.Success {
        return errors.New(resp.Message)
    }

    // Update local state from response
    m.playerState = resp.NewState
    m.currentShip = resp.NewShip

    return nil
}
```

## Migration Steps

### Step 1: Initialize API Client

**File**: `internal/tui/model.go`

```go
// Add API client to Model
type Model struct {
    // OLD: Direct access
    // playerRepo *database.PlayerRepository
    // systemRepo *database.SystemRepository
    // ...

    // NEW: API client
    apiClient api.Client

    // Cache of state for rendering
    playerState *api.PlayerState
    currentShip *api.Ship
    inventory   *api.Inventory
}

// Update NewModel to create API client
func NewModel(playerID uuid.UUID, repos Repositories) *Model {
    // Create in-process server
    gameServer, err := server.NewGameServer(&server.Config{
        PlayerRepo: repos.PlayerRepo,
        SystemRepo: repos.SystemRepo,
        ShipRepo:   repos.ShipRepo,
        MarketRepo: repos.MarketRepo,
        SSHKeyRepo: repos.SSHKeyRepo,
    })
    if err != nil {
        panic(err) // Handle properly in production
    }

    // Create in-process API client
    apiClient, err := api.NewClient(&api.ClientConfig{
        Mode:            api.ClientModeInProcess,
        InProcessServer: gameServer,
    })
    if err != nil {
        panic(err)
    }

    m := &Model{
        apiClient: apiClient,
        playerID:  playerID,
    }

    // Load initial state
    m.refreshPlayerState()

    return m
}

// Helper to refresh player state from API
func (m *Model) refreshPlayerState() error {
    ctx := context.Background()
    state, err := m.apiClient.GetPlayerState(ctx, m.playerID)
    if err != nil {
        return err
    }

    m.playerState = state
    m.currentShip = state.Ship
    m.inventory = state.Inventory

    return nil
}
```

### Step 2: Migrate Trading Screen

**File**: `internal/tui/trading.go`

#### Before:
```go
func (m *Model) buyCommodity(commodityID string, quantity int) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Get player
        player, err := m.playerRepo.GetByID(ctx, m.playerID)
        if err != nil {
            return errorMsg{err}
        }

        // Get ship
        ship, err := m.shipRepo.GetByID(ctx, player.CurrentShipID)
        if err != nil {
            return errorMsg{err}
        }

        // Get market
        market, err := m.marketRepo.GetBySystemID(ctx, player.CurrentSystemID)
        if err != nil {
            return errorMsg{err}
        }

        // Calculate cost
        commodity := market.FindCommodity(commodityID)
        totalCost := commodity.BuyPrice * quantity

        // Validate credits
        if player.Credits < totalCost {
            return errorMsg{errors.New("insufficient credits")}
        }

        // Validate cargo space
        if ship.CargoUsed+quantity > ship.CargoSpace {
            return errorMsg{errors.New("insufficient cargo space")}
        }

        // Update cargo
        ship.Cargo[commodityID] += quantity
        ship.CargoUsed += quantity

        // Update credits
        player.Credits -= totalCost

        // Save to database
        if err := m.shipRepo.Update(ctx, ship); err != nil {
            return errorMsg{err}
        }
        if err := m.playerRepo.Update(ctx, player); err != nil {
            return errorMsg{err}
        }

        // Update market stock
        market.UpdateStock(commodityID, -quantity)
        if err := m.marketRepo.Update(ctx, market); err != nil {
            return errorMsg{err}
        }

        return tradeCompleteMsg{commodity: commodityID, quantity: quantity}
    }
}
```

#### After:
```go
func (m *Model) buyCommodity(commodityID string, quantity int) tea.Cmd {
    return func() tea.Msg {
        req := &api.TradeRequest{
            PlayerID:    m.playerID,
            CommodityID: commodityID,
            Quantity:    quantity,
        }

        resp, err := m.apiClient.BuyCommodity(context.Background(), req)
        if err != nil {
            return errorMsg{err}
        }

        if !resp.Success {
            return errorMsg{errors.New(resp.Message)}
        }

        // Update local state from API response
        m.playerState = resp.NewState
        m.currentShip = resp.NewState.Ship
        m.inventory = resp.NewState.Inventory

        return tradeCompleteMsg{
            commodity: commodityID,
            quantity:  resp.QuantityTraded,
            cost:      resp.TotalCost,
        }
    }
}
```

**Benefits**:
- ✅ Reduced from ~40 lines to ~15 lines
- ✅ No database transaction management in TUI
- ✅ No multi-step updates (atomic on server)
- ✅ Server validates everything
- ✅ Clear error messages from server
- ✅ State always consistent (from server)

### Step 3: Migrate Navigation Screen

**File**: `internal/tui/navigation.go`

#### Before:
```go
func (m *Model) jumpToSystem(targetSystemID uuid.UUID) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Validate connection exists
        connections, err := m.systemRepo.GetConnections(ctx, m.player.CurrentSystemID)
        if err != nil {
            return errorMsg{err}
        }

        canJump := false
        for _, conn := range connections {
            if conn == targetSystemID {
                canJump = true
                break
            }
        }

        if !canJump {
            return errorMsg{errors.New("no jump route to that system")}
        }

        // Check fuel
        fuelRequired := calculateFuelCost(m.player.CurrentSystemID, targetSystemID)
        if m.player.Fuel < fuelRequired {
            return errorMsg{errors.New("insufficient fuel")}
        }

        // Update location
        m.player.CurrentSystemID = targetSystemID
        m.player.CurrentPlanetID = nil
        m.player.Fuel -= fuelRequired

        // Save
        if err := m.playerRepo.Update(ctx, m.player); err != nil {
            return errorMsg{err}
        }

        return jumpCompleteMsg{systemID: targetSystemID}
    }
}
```

#### After:
```go
func (m *Model) jumpToSystem(targetSystemID uuid.UUID) tea.Cmd {
    return func() tea.Msg {
        req := &api.JumpRequest{
            PlayerID:       m.playerID,
            TargetSystemID: targetSystemID,
        }

        resp, err := m.apiClient.Jump(context.Background(), req)
        if err != nil {
            return errorMsg{err}
        }

        if !resp.Success {
            return errorMsg{errors.New(resp.Message)}
        }

        // Update local state
        m.playerState = resp.NewState

        return jumpCompleteMsg{
            systemID:     targetSystemID,
            fuelConsumed: resp.FuelConsumed,
        }
    }
}
```

### Step 4: Handle Optimistic Updates

For actions that should feel instant, we can optimistically update the UI before the server responds:

```go
func (m *Model) buyCommodityOptimistic(commodityID string, quantity int) tea.Cmd {
    // Get estimated cost
    commodity := m.findCommodityInMarket(commodityID)
    estimatedCost := commodity.BuyPrice * quantity

    // Optimistically update UI immediately
    m.optimisticUpdate = &OptimisticUpdate{
        Type:      "buy_commodity",
        CommodityID: commodityID,
        Quantity:  quantity,
        EstimatedCost: estimatedCost,
    }

    // Apply optimistic changes to local state
    if m.inventory.Cargo == nil {
        m.inventory.Cargo = make(map[string]int32)
    }
    m.inventory.Cargo[commodityID] += int32(quantity)
    m.playerState.Credits -= estimatedCost

    // Send actual request
    return func() tea.Msg {
        req := &api.TradeRequest{
            PlayerID:    m.playerID,
            CommodityID: commodityID,
            Quantity:    int32(quantity),
        }

        resp, err := m.apiClient.BuyCommodity(context.Background(), req)
        if err != nil {
            // Rollback optimistic update
            m.rollbackOptimisticUpdate()
            return errorMsg{err}
        }

        if !resp.Success {
            m.rollbackOptimisticUpdate()
            return errorMsg{errors.New(resp.Message)}
        }

        // Confirm with server state (may differ from optimistic)
        m.optimisticUpdate = nil
        m.playerState = resp.NewState
        m.inventory = resp.NewState.Inventory

        return tradeCompleteMsg{
            commodity: commodityID,
            quantity:  resp.QuantityTraded,
            cost:      resp.TotalCost,
        }
    }
}

func (m *Model) rollbackOptimisticUpdate() {
    if m.optimisticUpdate == nil {
        return
    }

    switch m.optimisticUpdate.Type {
    case "buy_commodity":
        m.inventory.Cargo[m.optimisticUpdate.CommodityID] -= int32(m.optimisticUpdate.Quantity)
        m.playerState.Credits += m.optimisticUpdate.EstimatedCost
    }

    m.optimisticUpdate = nil
}
```

### Step 5: Stream Real-Time Updates

For multiplayer features, subscribe to real-time updates:

```go
func (m *Model) subscribeToPlayerUpdates() tea.Cmd {
    return func() tea.Msg {
        stream, err := m.apiClient.StreamPlayerUpdates(context.Background(), m.playerID)
        if err != nil {
            return errorMsg{err}
        }

        // Start goroutine to receive updates
        go func() {
            for {
                update, err := stream.Recv()
                if err != nil {
                    if err == io.EOF {
                        return
                    }
                    // Send error to TUI
                    m.program.Send(errorMsg{err})
                    return
                }

                // Send update to TUI
                m.program.Send(playerUpdateMsg{update})
            }
        }()

        return subscribeCompleteMsg{}
    }
}

// Handle updates in Update()
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case playerUpdateMsg:
        m.handlePlayerUpdate(msg.update)
        return m, nil
    }
    // ... rest of Update
}

func (m *Model) handlePlayerUpdate(update *api.PlayerUpdate) {
    switch update.Type {
    case api.UpdateTypeCredits:
        m.playerState.Credits = update.CreditsUpdate.NewCredits
    case api.UpdateTypeLocation:
        m.playerState.CurrentSystemID = update.LocationUpdate.SystemID
        m.playerState.CurrentPlanetID = update.LocationUpdate.PlanetID
    case api.UpdateTypeShip:
        m.currentShip = update.ShipUpdate.Ship
    }
}
```

## Migration Checklist

For each TUI screen:

### Before Starting
- [ ] Read through existing screen code
- [ ] Identify all repository/manager calls
- [ ] List required API endpoints
- [ ] Check if endpoints exist in `api/client.go`

### During Migration
- [ ] Replace repository fields with API client
- [ ] Convert database calls to API calls
- [ ] Update state from API responses (not local updates)
- [ ] Handle errors from API (don't duplicate validation)
- [ ] Remove database transaction logic
- [ ] Test the screen thoroughly

### After Migration
- [ ] Verify no direct repository/manager access remains
- [ ] Check error handling is correct
- [ ] Ensure UI updates from server state
- [ ] Test edge cases (low credits, full cargo, etc.)
- [ ] Update tests to use API mocks

## Screen Migration Priority

Suggested order (easiest to hardest):

1. **Trading Screen** - Simple CRUD, good starter
2. **Navigation Screen** - Straightforward state updates
3. **Shipyard Screen** - Moderate complexity
4. **Mission Screen** - List operations
5. **Combat Screen** - Complex state, tackle last

## Common Patterns

### Pattern 1: Simple Query
```go
// Before: m.playerRepo.GetByID(ctx, playerID)
// After:
state, err := m.apiClient.GetPlayerState(ctx, playerID)
```

### Pattern 2: State Mutation
```go
// Before: Multiple repo.Update() calls
// After:
resp, err := m.apiClient.BuyShip(ctx, &api.ShipPurchaseRequest{...})
m.playerState = resp.NewState  // Single state update
```

### Pattern 3: List Operations
```go
// Before: m.missionsManager.GetAvailable(playerID)
// After:
missions, err := m.apiClient.GetAvailableMissions(ctx, playerID)
```

### Pattern 4: Error Handling
```go
// Server returns success boolean + message
resp, err := m.apiClient.SomethingAction(ctx, req)
if err != nil {
    return errorMsg{err}  // Network/server error
}
if !resp.Success {
    return errorMsg{errors.New(resp.Message)}  // Business logic error
}
```

## Testing

### Unit Tests
```go
func TestBuyCommodity(t *testing.T) {
    // Create mock API client
    mockClient := &MockAPIClient{
        BuyCommodityFunc: func(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
            return &api.TradeResponse{
                Success:        true,
                QuantityTraded: req.Quantity,
                TotalCost:      1000,
                NewState: &api.PlayerState{
                    Credits: 9000,  // 10000 - 1000
                },
            }, nil
        },
    }

    model := &Model{
        apiClient: mockClient,
        playerState: &api.PlayerState{
            Credits: 10000,
        },
    }

    // Execute action
    cmd := model.buyCommodity("food", 10)
    msg := cmd()

    // Verify
    assert.IsType(t, tradeCompleteMsg{}, msg)
    assert.Equal(t, int64(9000), model.playerState.Credits)
}
```

### Integration Tests
```go
func TestBuyCommodityIntegration(t *testing.T) {
    // Create real server with test database
    server, cleanup := setupTestServer(t)
    defer cleanup()

    client, _ := api.NewClient(&api.ClientConfig{
        Mode:            api.ClientModeInProcess,
        InProcessServer: server,
    })

    // Execute real API call
    resp, err := client.BuyCommodity(context.Background(), &api.TradeRequest{
        PlayerID:    testPlayerID,
        CommodityID: "food",
        Quantity:    10,
    })

    assert.NoError(t, err)
    assert.True(t, resp.Success)
}
```

## Troubleshooting

### Issue: "API endpoint not implemented"
**Solution**: The server handler returns `api.ErrNotFound`. Implement the handler in `internal/api/server/server.go`.

### Issue: "State not updating in UI"
**Solution**: Ensure you're updating local state from API responses:
```go
resp, _ := m.apiClient.SomeAction(...)
m.playerState = resp.NewState  // Don't forget this!
```

### Issue: "Too many API calls"
**Solution**: Batch related queries or cache state locally:
```go
// Instead of calling GetPlayerState() on every render
// Call once and cache the result
```

### Issue: "Optimistic updates not rolling back"
**Solution**: Always save old state before optimistic update:
```go
m.stateSnapshot = m.playerState.Clone()
// ... optimistic update ...
// On error:
m.playerState = m.stateSnapshot
```

## Benefits Summary

After migration, you'll have:

- ✅ **Cleaner TUI code** - No database logic in UI layer
- ✅ **Consistent state** - Server is source of truth
- ✅ **Better errors** - Server validates and returns clear messages
- ✅ **Easier testing** - Mock API client instead of repositories
- ✅ **Future-proof** - Can switch to gRPC in Phase 2 without TUI changes
- ✅ **Type safety** - Compile-time checks for API calls

## Real Implementation Examples

The following handlers have been fully implemented as reference examples:

### 1. GetPlayerState - Data Retrieval Pattern

**File**: `internal/api/server/server.go:128-147`

Shows the pattern for retrieving and aggregating player state:
```go
func (s *GameServer) GetPlayerState(ctx context.Context, playerID uuid.UUID) (*api.PlayerState, error) {
    // Load data from repositories
    player, err := s.playerRepo.GetByID(ctx, playerID)
    if err != nil {
        return nil, err
    }

    ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
    if err != nil {
        return nil, err
    }

    // Convert to API types using converters
    state := convertPlayerToAPI(player, ship)
    state.Stats = convertPlayerStatsToAPI(player)
    state.Reputation = convertReputationToAPI(player)

    return state, nil
}
```

**Key Points**:
- Load all required data from repositories
- Use converter functions to transform models → API types
- Return aggregate state in single response
- Handle errors early with direct returns

### 2. BuyCommodity - State Mutation Pattern

**File**: `internal/api/server/server.go:223-348`

Shows the pattern for validating and mutating game state:
```go
func (s *GameServer) BuyCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
    // 1. Load required data
    player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
    ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
    commodities, err := s.marketRepo.GetCommoditiesBySystemID(ctx, player.CurrentSystemID)

    // 2. Validate preconditions
    if player.CurrentPlanetID == nil {
        return &api.TradeResponse{Success: false, Message: "You must be docked"}, nil
    }

    // 3. Validate business rules
    if totalCost > player.Credits {
        return &api.TradeResponse{Success: false, Message: "Insufficient credits"}, nil
    }
    if cargoAvailable < req.Quantity {
        return &api.TradeResponse{Success: false, Message: "Insufficient cargo space"}, nil
    }

    // 4. Mutate state
    ship.Cargo[req.CommodityID] += int(req.Quantity)
    ship.CargoUsed += req.Quantity
    player.Credits -= totalCost

    // 5. Persist changes
    s.shipRepo.Update(ctx, ship)
    s.playerRepo.Update(ctx, player)
    s.marketRepo.UpdateStock(ctx, player.CurrentSystemID, req.CommodityID, -int(req.Quantity))

    // 6. Return success with updated state
    return &api.TradeResponse{
        Success: true,
        NewState: convertPlayerToAPI(player, ship),
        QuantityTraded: req.Quantity,
        TotalCost: totalCost,
    }, nil
}
```

**Key Points**:
- Always return `Success` field for client validation
- Include descriptive `Message` for failures
- Return updated `NewState` so client can sync
- Use transactions where available (TODO: add transaction support)
- Handle partial failures gracefully

### 3. Jump - Navigation Pattern

**File**: `internal/api/server/server.go:191-297`

Shows the pattern for complex multi-step operations:
```go
func (s *GameServer) Jump(ctx context.Context, req *api.JumpRequest) (*api.JumpResponse, error) {
    // 1. Load current state
    player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
    ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)

    // 2. Validate preconditions
    if player.CurrentPlanetID != nil {
        return &api.JumpResponse{Success: false, Message: "You must take off first"}, nil
    }

    // 3. Validate game rules (connectivity)
    connections, err := s.systemRepo.GetConnections(ctx, player.CurrentSystemID)
    isConnected := false
    for _, connectedSystemID := range connections {
        if connectedSystemID == req.TargetSystemID {
            isConnected = true
            break
        }
    }
    if !isConnected {
        return &api.JumpResponse{Success: false, Message: "No jump route"}, nil
    }

    // 4. Validate resources
    if ship.Fuel < fuelCostPerJump {
        return &api.JumpResponse{Success: false, Message: "Insufficient fuel"}, nil
    }

    // 5. Execute operation
    player.CurrentSystemID = req.TargetSystemID
    player.X = 0
    player.Y = 0
    ship.Fuel -= fuelCostPerJump
    player.JumpsMade++

    // 6. Persist and return
    s.playerRepo.Update(ctx, player)
    s.shipRepo.Update(ctx, ship)

    return &api.JumpResponse{
        Success: true,
        NewState: convertPlayerToAPI(player, ship),
        FuelConsumed: fuelCostPerJump,
    }, nil
}
```

**Key Points**:
- Validate preconditions before any mutations
- Check game rules (connections, resources)
- Update all affected state atomically
- Return specific operation results (`FuelConsumed`)
- Always include updated state in response

### 4. Data Converters

**File**: `internal/api/server/converters.go`

All converters follow this pattern:
```go
func convertPlayerToAPI(player *models.Player, ship *models.Ship) *api.PlayerState {
    if player == nil {
        return nil
    }

    state := &api.PlayerState{
        PlayerID:        player.ID,
        Username:        player.Username,
        CurrentSystemID: player.CurrentSystemID,
        Credits:         player.Credits,
        // ... map all fields ...
    }

    // Convert nested objects
    if ship != nil {
        state.Ship = convertShipToAPI(ship)
        state.Inventory = convertInventoryToAPI(ship)
    }

    // Determine derived fields
    if player.CurrentPlanetID != nil {
        state.Status = api.PlayerStatusDocked
    } else {
        state.Status = api.PlayerStatusInSpace
    }

    return state
}
```

**Key Points**:
- Always nil-check input
- Map all relevant fields
- Convert nested objects recursively
- Calculate derived/computed fields
- Use enums from API types, not database strings

## Proof-of-Concept Migration

See **[`PHASE_1B_EXAMPLE.md`](./PHASE_1B_EXAMPLE.md)** for:
- Complete before/after comparison
- Trading screen migration (80+ lines → 40 lines)
- Model changes required
- Testing strategies
- Performance analysis

## Next Steps

After migrating a screen:

1. Remove unused repository references from Model
2. Update screen tests to use API mocks
3. Document any new API endpoints needed
4. Consider implementing missing server handlers
5. Move on to next screen

---

**Need Help?**
- See `internal/api/client.go` for available API methods
- See `internal/api/server/server.go` for server implementation
- See `internal/api/server/converters.go` for conversion utilities
- See `internal/api/types.go` for request/response types
- See `docs/ARCHITECTURE_REFACTORING.md` for overall design
- See `docs/PHASE_1B_EXAMPLE.md` for real migration example

**Last Updated**: 2025-01-14
**Document Version**: 1.1.0
