# Phase 1b: TUI Migration Proof-of-Concept

This document shows a real before/after example of migrating the trading screen to use the new API layer.

## Overview

**File**: `internal/tui/trading.go`
**Handler**: Commodity buying functionality
**Migration Time**: ~30 minutes per screen

## Before: Direct Database Access

The current implementation directly accesses multiple repositories and managers:

```go
// File: internal/tui/trading.go (BEFORE)
func (m Model) executeBuy() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Validate we have all required data
        if m.trading.selectedCommodity == nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("no commodity selected"),
            }
        }

        if m.currentShip == nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("no ship available"),
            }
        }

        if m.trading.currentPlanet == nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("not docked at a planet"),
            }
        }

        // Find market price
        var price *models.MarketPrice
        for _, p := range m.trading.marketPrices {
            if p.CommodityID == m.trading.selectedCommodity.ID {
                price = p
                break
            }
        }

        if price == nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("price data not available"),
            }
        }

        // Calculate total cost
        totalCost := price.SellPrice * int64(m.trading.quantity)

        // Validate credits
        if totalCost > m.player.Credits {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("insufficient credits (need %d)", totalCost),
            }
        }

        // Validate cargo space
        cargoUsed := m.currentShip.GetCargoUsed()
        cargoAvailable := 100 - cargoUsed
        if m.trading.quantity > cargoAvailable {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("insufficient cargo space (have %d)", cargoAvailable),
            }
        }

        // Validate stock availability
        if m.trading.quantity > price.Stock {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("insufficient stock (available: %d)", price.Stock),
            }
        }

        // Execute transaction: deduct credits
        newCredits := m.player.Credits - totalCost
        err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
        if err != nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("failed to update credits: %w", err),
            }
        }

        // Add cargo to ship
        m.currentShip.AddCargo(m.trading.selectedCommodity.ID, m.trading.quantity)
        err = m.shipRepo.Update(ctx, m.currentShip)
        if err != nil {
            // Rollback would be ideal here
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("failed to update ship: %w", err),
            }
        }

        // Update market stock
        err = m.marketRepo.UpdateStock(ctx,
            m.player.CurrentSystemID,
            m.trading.selectedCommodity.ID,
            -m.trading.quantity)
        if err != nil {
            // Stock update failed, but transaction succeeded
            // Continue anyway as this is non-critical
        }

        // Update local player state
        m.player.Credits = newCredits

        // Calculate profit (negative for purchases)
        profit := -totalCost

        return tradeCompleteMsg{
            success: true,
            profit:  profit,
        }
    }
}
```

**Problems with this approach:**
- 80+ lines of validation and business logic in the UI
- Direct database access from TUI
- Manual state synchronization
- Error-prone rollback logic
- Tight coupling between UI and database
- Difficult to test in isolation
- Cannot be reused by other clients

## After: API-Based Implementation

The new implementation delegates to the server API:

```go
// File: internal/tui/trading.go (AFTER)
func (m Model) executeBuy() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Validate we have the commodity selected
        if m.trading.selectedCommodity == nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("no commodity selected"),
            }
        }

        // Construct API request
        req := &api.TradeRequest{
            PlayerID:    m.playerID,
            CommodityID: m.trading.selectedCommodity.ID,
            Quantity:    int32(m.trading.quantity),
        }

        // Call API
        resp, err := m.apiClient.BuyCommodity(ctx, req)
        if err != nil {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf("API error: %w", err),
            }
        }

        // Check response
        if !resp.Success {
            return tradeCompleteMsg{
                success: false,
                err:     fmt.Errorf(resp.Message),
            }
        }

        // Update local state from server response
        m.updatePlayerState(resp.NewState)

        // Calculate profit (negative for purchases)
        profit := -resp.TotalCost

        return tradeCompleteMsg{
            success: true,
            profit:  profit,
        }
    }
}
```

**Benefits:**
- 40 lines instead of 80+ (50% reduction)
- All validation/business logic on server
- Server is authoritative source of truth
- Automatic state synchronization
- Proper transaction handling on server
- Easy to test (mock API client)
- Can be reused by future clients (web, mobile)

## Model Changes Required

### 1. Add API Client to Model

```go
// File: internal/tui/model.go
type Model struct {
    // ... existing fields ...

    // Phase 1: In-process API client
    apiClient api.Client

    // Deprecated: Will be removed in Phase 2
    // playerRepo *database.PlayerRepository
    // shipRepo   *database.ShipRepository
    // marketRepo *database.MarketRepository
}
```

### 2. Initialize API Client

```go
// File: internal/server/server.go
func (s *Server) startGameSession(pty ssh.Session, player *models.Player) error {
    // ... existing code ...

    // Create game server (implements api.Server interface)
    gameServer, err := server.NewGameServer(&server.Config{
        PlayerRepo: s.playerRepo,
        SystemRepo: s.systemRepo,
        ShipRepo:   s.shipRepo,
        MarketRepo: s.marketRepo,
        SSHKeyRepo: s.sshKeyRepo,
    })
    if err != nil {
        return err
    }

    // Create in-process API client
    apiClient := api.NewInProcessClient(gameServer)

    // Initialize TUI with API client
    tuiModel := tui.NewModel(
        player.ID,
        apiClient,
        // ... other params ...
    )

    // ... rest of code ...
}
```

### 3. Add Helper Method for State Updates

```go
// File: internal/tui/model.go
func (m *Model) updatePlayerState(state *api.PlayerState) {
    // Update all player state from API response
    m.player.Credits = state.Credits
    m.player.CurrentSystemID = state.CurrentSystemID
    m.player.CurrentPlanetID = state.CurrentPlanetID
    m.player.X = state.Position.X
    m.player.Y = state.Position.Y

    // Update ship if provided
    if state.Ship != nil {
        m.updateShipFromAPI(state.Ship)
    }

    // Update stats if provided
    if state.Stats != nil {
        m.updateStatsFromAPI(state.Stats)
    }

    // Update reputation if provided
    if state.Reputation != nil {
        m.updateReputationFromAPI(state.Reputation)
    }
}

func (m *Model) updateShipFromAPI(ship *api.Ship) {
    if m.currentShip == nil {
        return
    }

    m.currentShip.Hull = ship.Hull
    m.currentShip.MaxHull = ship.MaxHull
    m.currentShip.Shields = ship.Shields
    m.currentShip.Fuel = ship.Fuel
    m.currentShip.CargoUsed = ship.CargoUsed

    // Update cargo map
    if ship.Inventory != nil {
        m.currentShip.Cargo = make(map[string]int)
        for commodity, quantity := range ship.Inventory.Cargo {
            m.currentShip.Cargo[commodity] = int(quantity)
        }
    }
}
```

## Migration Checklist

For each TUI screen that needs migration:

- [ ] Identify all repository/manager calls
- [ ] Map to appropriate API methods
- [ ] Replace direct calls with API client calls
- [ ] Update state from API responses using helper methods
- [ ] Remove validation logic (now on server)
- [ ] Test with in-process client
- [ ] Verify behavior matches original

## Testing Strategy

### 1. Unit Tests with Mock Client

```go
// File: internal/tui/trading_test.go
func TestExecuteBuy(t *testing.T) {
    // Create mock API client
    mockClient := &mockAPIClient{
        buyCommodityFunc: func(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
            return &api.TradeResponse{
                Success:        true,
                Message:        "Purchase successful",
                QuantityTraded: req.Quantity,
                TotalCost:      1000,
                PricePerUnit:   100,
                NewState: &api.PlayerState{
                    Credits: 9000, // Had 10000, spent 1000
                    // ... rest of state
                },
            }, nil
        },
    }

    // Create model with mock client
    model := Model{
        apiClient: mockClient,
        trading: tradingModel{
            selectedCommodity: &models.Commodity{ID: "food"},
            quantity: 10,
        },
    }

    // Execute buy
    cmd := model.executeBuy()
    msg := cmd()

    // Assert results
    tradeMsg := msg.(tradeCompleteMsg)
    assert.True(t, tradeMsg.success)
    assert.Equal(t, int64(-1000), tradeMsg.profit)
}
```

### 2. Integration Tests

```go
// File: internal/api/server/integration_test.go
func TestBuyCommodityIntegration(t *testing.T) {
    // Set up test database
    db := setupTestDB(t)
    defer db.Close()

    // Create repositories
    playerRepo := database.NewPlayerRepository(db)
    shipRepo := database.NewShipRepository(db)
    marketRepo := database.NewMarketRepository(db)

    // Create test player with 10000 credits
    player := createTestPlayer(t, playerRepo, 10000)

    // Create game server
    gameServer, err := server.NewGameServer(&server.Config{
        PlayerRepo: playerRepo,
        ShipRepo:   shipRepo,
        MarketRepo: marketRepo,
    })
    require.NoError(t, err)

    // Execute buy
    req := &api.TradeRequest{
        PlayerID:    player.ID,
        CommodityID: "food",
        Quantity:    10,
    }

    resp, err := gameServer.BuyCommodity(context.Background(), req)
    require.NoError(t, err)
    assert.True(t, resp.Success)
    assert.Equal(t, int64(9000), resp.NewState.Credits)

    // Verify database state
    updatedPlayer, err := playerRepo.GetByID(context.Background(), player.ID)
    require.NoError(t, err)
    assert.Equal(t, int64(9000), updatedPlayer.Credits)
}
```

## Performance Considerations

### In-Process Client (Phase 1)

- **Latency**: ~0ms (function call overhead only)
- **Memory**: Minimal (shared process memory)
- **Throughput**: Same as current implementation

The in-process client has virtually no performance overhead compared to direct repository access.

### gRPC Client (Phase 2+)

- **Latency**: 1-5ms (local network)
- **Memory**: Slightly higher (protobuf serialization)
- **Throughput**: 10,000+ req/s per connection
- **Benefits**: Horizontal scaling, service isolation

## Next Steps

1. ✅ Implement core server handlers (GetPlayerState, BuyCommodity, Jump)
2. ✅ Create converter utilities
3. 🔄 Create this proof-of-concept documentation
4. ⏳ Add unit tests for server handlers
5. ⏳ Update migration guide with real examples
6. ⏳ Begin migrating TUI screens (trading → navigation → combat)
7. ⏳ Commit Phase 1b implementation

## Timeline Estimate

**Remaining Phase 1 work**: 2-3 weeks
- Server handlers: 1 week (31 methods, ~2-3 per day)
- TUI migration: 1 week (26 screens, ~3-4 per day)
- Testing & polish: 3-5 days
- Documentation: 2-3 days (ongoing)

**Phase 2 (Service Split)**: 2-3 weeks after Phase 1 complete
