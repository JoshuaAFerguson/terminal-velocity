# Internal API Package

**Package**: `internal/api`
**Purpose**: Abstraction layer between TUI (frontend) and game logic (backend)
**Phase**: Phase 1 - Internal API Extraction

## Overview

This package provides a clean API interface for the TUI to communicate with game logic. In Phase 1, this is an in-process call. In Phase 2+, this becomes gRPC over the network.

## Architecture

```
┌─────────────────────┐
│   TUI (frontend)    │
│  internal/tui/      │
└──────────┬──────────┘
           │ uses
           ▼
┌─────────────────────┐
│   API Client        │
│  internal/api/      │  ← This package
│  - client.go        │
│  - types.go         │
└──────────┬──────────┘
           │ calls
           ▼
┌─────────────────────┐
│   API Server        │
│  internal/api/      │
│  server/            │
│  - server.go        │
│  - session.go       │
└──────────┬──────────┘
           │ uses
           ▼
┌─────────────────────┐
│  Game Logic         │
│  - Database repos   │
│  - Manager packages │
└─────────────────────┘
```

## Files

### `client.go`
Defines the `Client` interface that TUI uses for all game operations.

**Key Interfaces**:
- `Client` - Unified interface combining Auth, Player, and Game clients
- `AuthClient` - Authentication and session management
- `PlayerClient` - Player state and real-time updates
- `GameClient` - Core game actions (trading, navigation, combat)

**Implementation**:
- `inProcessClient` - Phase 1: Direct function calls to server
- `grpcClient` - Phase 2+: gRPC network calls (not yet implemented)

**Usage**:
```go
import "github.com/JoshuaAFerguson/terminal-velocity/internal/api"

// Create client
client, err := api.NewClient(&api.ClientConfig{
    Mode: api.ClientModeInProcess,
    InProcessServer: gameServer,
})

// Use client
resp, err := client.BuyCommodity(ctx, &api.TradeRequest{
    PlayerID: playerID,
    CommodityID: "food",
    Quantity: 10,
})
```

### `types.go`
Defines all request and response types used by the API.

**Type Categories**:
- **Auth Types**: AuthRequest, AuthResponse, Session, etc.
- **Player Types**: PlayerState, Ship, Inventory, Stats, Reputation
- **Game Types**: JumpRequest, TradeRequest, Mission, Quest, etc.

**Note**: In Phase 2+, these will be replaced by generated protobuf types from `api/proto/*.proto`.

### `server/server.go`
Implements the `api.Server` interface that handles all API requests.

**Services**:
- `AuthService` - 7 methods for authentication
- `PlayerService` - 7 methods for player state
- `GameService` - 17 methods for game actions

**Current Status**: Skeleton implementation with TODOs. Each handler needs to be implemented by wrapping existing game logic.

### `server/session.go`
Manages active game sessions.

**Features**:
- Session creation and validation
- Session expiration (24 hour TTL)
- Activity tracking
- Automatic cleanup of expired sessions

## Usage Examples

### Initialize Client (in TUI)

```go
// internal/tui/model.go

func NewModel(playerID uuid.UUID, repos Repositories) *Model {
    // Create server
    gameServer, err := server.NewGameServer(&server.Config{
        PlayerRepo: repos.PlayerRepo,
        SystemRepo: repos.SystemRepo,
        ShipRepo:   repos.ShipRepo,
        MarketRepo: repos.MarketRepo,
        SSHKeyRepo: repos.SSHKeyRepo,
    })
    if err != nil {
        panic(err)
    }

    // Create client
    apiClient, err := api.NewClient(&api.ClientConfig{
        Mode:            api.ClientModeInProcess,
        InProcessServer: gameServer,
    })
    if err != nil {
        panic(err)
    }

    return &Model{
        apiClient: apiClient,
        playerID:  playerID,
    }
}
```

### Call API Methods

```go
// Buy commodity
resp, err := m.apiClient.BuyCommodity(ctx, &api.TradeRequest{
    PlayerID:    m.playerID,
    CommodityID: "food",
    Quantity:    10,
})
if err != nil {
    return err
}
if !resp.Success {
    return errors.New(resp.Message)
}

// Update local state
m.playerState = resp.NewState

// Get player state
state, err := m.apiClient.GetPlayerState(ctx, m.playerID)
if err != nil {
    return err
}

// Jump to system
resp, err := m.apiClient.Jump(ctx, &api.JumpRequest{
    PlayerID:       m.playerID,
    TargetSystemID: targetSystemID,
})
```

### Stream Real-Time Updates

```go
// Subscribe to player updates
stream, err := m.apiClient.StreamPlayerUpdates(ctx, m.playerID)
if err != nil {
    return err
}

go func() {
    for {
        update, err := stream.Recv()
        if err != nil {
            if err == io.EOF {
                return
            }
            log.Printf("stream error: %v", err)
            return
        }

        // Handle update
        switch update.Type {
        case api.UpdateTypeCredits:
            m.updateCredits(update.CreditsUpdate)
        case api.UpdateTypeLocation:
            m.updateLocation(update.LocationUpdate)
        }
    }
}()
```

## Implementing Server Handlers

When implementing a server handler, follow this pattern:

```go
// internal/api/server/server.go

func (s *GameServer) BuyCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
    // 1. Validate request
    if req.PlayerID == uuid.Nil {
        return nil, api.ErrInvalidRequest
    }
    if req.Quantity <= 0 {
        return nil, api.ErrInvalidRequest
    }

    // 2. Load necessary data
    player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
    if err != nil {
        return nil, err
    }

    ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
    if err != nil {
        return nil, err
    }

    market, err := s.marketRepo.GetBySystemID(ctx, player.CurrentSystemID)
    if err != nil {
        return nil, err
    }

    // 3. Validate business rules
    commodity := market.FindCommodity(req.CommodityID)
    if commodity == nil {
        return &api.TradeResponse{
            Success: false,
            Message: "commodity not available in this system",
        }, nil
    }

    totalCost := int64(commodity.BuyPrice) * int64(req.Quantity)
    if player.Credits < totalCost {
        return &api.TradeResponse{
            Success: false,
            Message: "insufficient credits",
        }, nil
    }

    if ship.CargoUsed+req.Quantity > ship.CargoSpace {
        return &api.TradeResponse{
            Success: false,
            Message: "insufficient cargo space",
        }, nil
    }

    // 4. Perform transaction (in database transaction if needed)
    if ship.Cargo == nil {
        ship.Cargo = make(map[string]int32)
    }
    ship.Cargo[req.CommodityID] += req.Quantity
    ship.CargoUsed += req.Quantity
    player.Credits -= totalCost

    // 5. Persist changes
    if err := s.shipRepo.Update(ctx, ship); err != nil {
        return nil, err
    }
    if err := s.playerRepo.Update(ctx, player); err != nil {
        return nil, err
    }

    // Update market
    market.UpdateStock(req.CommodityID, -int(req.Quantity))
    if err := s.marketRepo.Update(ctx, market); err != nil {
        return nil, err
    }

    // 6. Build response with complete state
    newState := s.buildPlayerState(player, ship)

    return &api.TradeResponse{
        Success:        true,
        Message:        "purchase successful",
        QuantityTraded: req.Quantity,
        TotalCost:      totalCost,
        PricePerUnit:   commodity.BuyPrice,
        NewState:       newState,
    }, nil
}

// Helper to build complete player state
func (s *GameServer) buildPlayerState(player *models.Player, ship *models.Ship) *api.PlayerState {
    return &api.PlayerState{
        PlayerID:        player.ID,
        Username:        player.Username,
        CurrentSystemID: player.CurrentSystemID,
        CurrentPlanetID: player.CurrentPlanetID,
        Credits:         player.Credits,
        CurrentShipID:   ship.ID,
        Ship:            convertShipToAPI(ship),
        // ... etc
    }
}
```

## Error Handling

### Common Errors

```go
// Defined in types.go
var (
    ErrNoServerProvided = errors.New("no server provided for in-process client")
    ErrInvalidRequest   = errors.New("invalid request")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrNotFound         = errors.New("not found")
    ErrForbidden        = errors.New("forbidden")
)
```

### Response Pattern

API responses use a two-level error system:

1. **Transport errors** - Returned as `error` (network, server down, etc.)
2. **Business logic errors** - Returned in response with `Success: false`

```go
resp, err := client.SomeAction(ctx, req)
if err != nil {
    // Transport/server error
    return fmt.Errorf("API call failed: %w", err)
}

if !resp.Success {
    // Business logic error (insufficient credits, etc.)
    return errors.New(resp.Message)
}

// Success
```

## Testing

### Mock Client for Unit Tests

```go
type MockAPIClient struct {
    BuyCommodityFunc func(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error)
    // ... other methods
}

func (m *MockAPIClient) BuyCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
    if m.BuyCommodityFunc != nil {
        return m.BuyCommodityFunc(ctx, req)
    }
    return nil, errors.New("not implemented")
}

// Use in tests
func TestBuyAction(t *testing.T) {
    mockClient := &MockAPIClient{
        BuyCommodityFunc: func(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
            return &api.TradeResponse{
                Success: true,
                NewState: &api.PlayerState{Credits: 9000},
            }, nil
        },
    }

    model := &Model{apiClient: mockClient}
    // ... test model methods
}
```

### Integration Tests

```go
func TestServerIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Create repositories
    playerRepo := database.NewPlayerRepository(db)
    // ... other repos

    // Create server
    server, err := server.NewGameServer(&server.Config{
        PlayerRepo: playerRepo,
        // ... other repos
    })
    require.NoError(t, err)

    // Test API calls
    resp, err := server.BuyCommodity(context.Background(), &api.TradeRequest{
        PlayerID:    testPlayerID,
        CommodityID: "food",
        Quantity:    10,
    })
    require.NoError(t, err)
    assert.True(t, resp.Success)
}
```

## Migration Status

### ✅ Complete
- Client interface defined
- Type system established
- In-process client implementation
- Server skeleton with session management
- Documentation

### 🚧 In Progress
- Server handler implementations (all TODOs)
- TUI refactoring to use API

### ⏳ Not Started
- Streaming implementation (PlayerUpdateStream)
- gRPC client (Phase 2)
- Performance optimizations
- Caching layer

## Next Steps

1. **Implement Server Handlers** - Replace TODO implementations
2. **Migrate TUI Screens** - One screen at a time (see `docs/API_MIGRATION_GUIDE.md`)
3. **Add Tests** - Unit tests for each handler
4. **Optimize** - Add caching where appropriate
5. **Document** - Add examples for complex operations

## Related Documentation

- `../../docs/ARCHITECTURE_REFACTORING.md` - Overall architecture design
- `../../docs/API_MIGRATION_GUIDE.md` - Step-by-step TUI migration guide
- `../../api/README.md` - Protocol Buffers API documentation

---

**Last Updated**: 2025-01-14
**Version**: 1.0.0
