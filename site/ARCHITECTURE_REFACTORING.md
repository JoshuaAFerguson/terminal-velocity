# Terminal Velocity: Client-Server Architecture Refactoring

**Document Version**: 1.0.0
**Created**: 2025-01-14
**Status**: Design Proposal
**Target Phase**: Phase 9 (Post-Launch Enhancement)

## Executive Summary

This document outlines the architectural refactoring of Terminal Velocity from a monolithic SSH server to a distributed client-server architecture. This refactoring will enable horizontal scaling, support for multiple client types, and better separation of concerns while maintaining backward compatibility with the existing SSH-based gameplay.

### Goals
- **Scalability**: Scale game logic servers independently from SSH gateways
- **Flexibility**: Support multiple client types (SSH, native terminal, web)
- **Maintainability**: Clean separation between presentation and business logic
- **Performance**: Reduce per-connection overhead on game servers
- **Future-Proofing**: Enable fat client architectures and offline features

---

## Current Architecture (Monolithic)

### Overview
```
┌─────────────────────────────────────────────────────────┐
│                 SSH Server Process                       │
│                   (Port 2222)                            │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  SSH Layer (internal/server)                   │    │
│  │  - Authentication (password + pubkey)          │    │
│  │  - Session management                          │    │
│  │  - Per-connection BubbleTea instance           │    │
│  └──────────────┬─────────────────────────────────┘    │
│                 │                                        │
│  ┌──────────────▼─────────────────────────────────┐    │
│  │  TUI Layer (internal/tui)                      │    │
│  │  - 26 BubbleTea screens                        │    │
│  │  - 28 component files                          │    │
│  │  - Direct manager access                       │    │
│  └──────────────┬─────────────────────────────────┘    │
│                 │                                        │
│  ┌──────────────▼─────────────────────────────────┐    │
│  │  Game Logic Layer                              │    │
│  │  - 16 Manager packages                         │    │
│  │  - Combat, trading, quests, events, etc.       │    │
│  │  - All thread-safe (sync.RWMutex)              │    │
│  └──────────────┬─────────────────────────────────┘    │
│                 │                                        │
│  ┌──────────────▼─────────────────────────────────┐    │
│  │  Database Layer (internal/database)            │    │
│  │  - 7 repositories                              │    │
│  │  - Direct pgx connection pool                  │    │
│  └──────────────┬─────────────────────────────────┘    │
│                 │                                        │
└─────────────────┼─────────────────────────────────────┘
                  │
                  ▼
         ┌─────────────────┐
         │   PostgreSQL    │
         │   (Port 5432)   │
         └─────────────────┘
```

### Problems with Current Architecture

**Scalability Issues**:
- Cannot scale SSH gateways independently from game logic
- Each SSH connection runs full game logic in-process
- Database connection pool shared across all connections
- Limited to vertical scaling (more CPU/RAM on one server)

**Coupling Issues**:
- TUI tightly coupled to game logic managers
- Database access embedded in manager packages
- Difficult to test game logic without SSH layer
- Cannot support alternative clients without code duplication

**Resource Issues**:
- Each connection holds manager references
- Background workers run per-process (not distributed)
- No session affinity or connection pooling to backend
- Database connections scale linearly with SSH connections

**Deployment Issues**:
- Single binary must include all components
- Cannot deploy hotfixes to game logic without SSH restart
- No way to gradually roll out changes
- Difficult to run multiple versions for testing

---

## Proposed Architecture (Client-Server)

### High-Level Overview

```
┌─────────────────────────────┐         ┌──────────────────────────────┐
│   SSH Gateway Servers        │         │   Game Logic Servers         │
│   (Multiple instances)       │◄───────►│   (Multiple instances)       │
│                              │   gRPC  │                              │
│  ┌────────────────────────┐ │         │  ┌────────────────────────┐  │
│  │ SSH Server             │ │         │  │ Game State Manager     │  │
│  │ - Auth gateway         │ │         │  │ - Universe state       │  │
│  │ - Session routing      │ │         │  │ - Player sessions      │  │
│  │ - Connection mgmt      │ │         │  │ - Event coordination   │  │
│  └────────────────────────┘ │         │  └────────────────────────┘  │
│                              │         │                              │
│  ┌────────────────────────┐ │         │  ┌────────────────────────┐  │
│  │ TUI Renderer           │ │         │  │ Game Logic Engines     │  │
│  │ - BubbleTea screens    │ │         │  │ - Combat engine        │  │
│  │ - Input handling       │ │         │  │ - Trading engine       │  │
│  │ - Display formatting   │ │         │  │ - Quest engine         │  │
│  └────────────────────────┘ │         │  │ - Event engine         │  │
│                              │         │  └────────────────────────┘  │
│  ┌────────────────────────┐ │         │                              │
│  │ API Client             │ │         │  ┌────────────────────────┐  │
│  │ - gRPC client          │ │         │  │ Managers               │  │
│  │ - State cache          │ │         │  │ - Chat, presence       │  │
│  │ - Event streaming      │ │         │  │ - Factions, territory  │  │
│  └────────────────────────┘ │         │  │ - Achievements, etc.   │  │
│                              │         │  └────────────────────────┘  │
└──────────────┬───────────────┘         │                              │
               │                         │  ┌────────────────────────┐  │
               │                         │  │ Database Layer         │  │
               │                         │  │ - Repositories         │  │
               │                         │  │ - Connection pool      │  │
               │                         │  └────────┬───────────────┘  │
               │                         └───────────┼───────────────────┘
               │                                     │
               │                                     ▼
               │                         ┌──────────────────────┐
               │                         │   PostgreSQL         │
               └────────────────────────►│   - Game state       │
                    Auth/Session         │   - User accounts    │
                                         └──────────────────────┘
```

### Component Responsibilities

#### SSH Gateway Servers (Frontend)
**Location**: `cmd/gateway/` (new)
**Purpose**: Handle SSH connections and render UI

**Responsibilities**:
- Accept SSH connections on port 2222
- Authenticate users (delegate to game server)
- Maintain active SSH sessions
- Render TUI using BubbleTea
- Route input to game server
- Stream state updates to display
- Handle connection lifecycle

**Does NOT**:
- ❌ Run game logic
- ❌ Access database directly (except auth cache)
- ❌ Manage game state
- ❌ Run background workers

**Scalability**:
- Stateless (session state on game server)
- Horizontal scaling via load balancer
- Low memory footprint (no game state)
- High connection density

#### Game Logic Servers (Backend)
**Location**: `cmd/gameserver/` (new)
**Purpose**: Execute all game logic and manage state

**Responsibilities**:
- Manage player sessions and state
- Execute game logic (combat, trading, quests)
- Run background workers (events, cleanup)
- Coordinate multiplayer interactions
- Persist state to database
- Broadcast state changes
- Handle game commands

**Does NOT**:
- ❌ Accept SSH connections directly
- ❌ Render UI (returns structured data)
- ❌ Handle terminal formatting

**Scalability**:
- Stateful (player sessions)
- Session affinity via consistent hashing
- Can shard by system/region
- Database connection pooling

---

## API Design

### Protocol Choice: gRPC

**Why gRPC**:
- ✅ Bidirectional streaming (server can push updates)
- ✅ Efficient binary protocol (protobuf)
- ✅ Strong typing and code generation
- ✅ Built-in load balancing and retries
- ✅ Better performance than REST for high-frequency calls
- ✅ Native Go support

**Alternative Considered**: WebSocket
- ❌ Less structured than gRPC
- ❌ No built-in code generation
- ❌ More manual state synchronization
- ✅ Easier to debug (text-based)
- ✅ Better browser support (for future web client)

**Decision**: Start with gRPC, potentially add WebSocket gateway later for web clients.

### API Surface

#### Service Definitions

```protobuf
// File: api/proto/game.proto

syntax = "proto3";
package terminalvelocity.v1;

// Authentication & Session Management
service AuthService {
  rpc Authenticate(AuthRequest) returns (AuthResponse);
  rpc CreateSession(CreateSessionRequest) returns (Session);
  rpc EndSession(EndSessionRequest) returns (Empty);
}

// Player State Management
service PlayerService {
  rpc GetPlayerState(PlayerID) returns (PlayerState);
  rpc UpdatePlayerLocation(LocationUpdate) returns (PlayerState);
  rpc GetPlayerShip(PlayerID) returns (Ship);
  rpc GetPlayerInventory(PlayerID) returns (Inventory);

  // Streaming state updates
  rpc StreamPlayerUpdates(PlayerID) returns (stream PlayerUpdate);
}

// Game Actions
service GameService {
  // Navigation
  rpc Jump(JumpRequest) returns (JumpResponse);
  rpc Land(LandRequest) returns (LandResponse);
  rpc Takeoff(TakeoffRequest) returns (TakeoffResponse);

  // Trading
  rpc GetMarket(SystemID) returns (Market);
  rpc BuyCommodity(TradeRequest) returns (TradeResponse);
  rpc SellCommodity(TradeRequest) returns (TradeResponse);

  // Combat
  rpc InitiateCombat(CombatRequest) returns (CombatSession);
  rpc ExecuteCombatAction(CombatAction) returns (CombatResult);
  rpc StreamCombat(CombatSessionID) returns (stream CombatUpdate);

  // Ship Management
  rpc BuyShip(ShipPurchaseRequest) returns (ShipPurchaseResponse);
  rpc SellShip(ShipSaleRequest) returns (ShipSaleResponse);
  rpc BuyOutfit(OutfitPurchaseRequest) returns (OutfitPurchaseResponse);

  // Missions & Quests
  rpc GetAvailableMissions(PlayerID) returns (MissionList);
  rpc AcceptMission(MissionID) returns (Mission);
  rpc GetActiveQuests(PlayerID) returns (QuestList);
}

// Multiplayer Features
service MultiplayerService {
  rpc SendChatMessage(ChatMessage) returns (Empty);
  rpc StreamChat(ChatChannel) returns (stream ChatMessage);

  rpc GetPlayersInSystem(SystemID) returns (PlayerList);
  rpc StreamPresence(SystemID) returns (stream PresenceUpdate);

  rpc InitiatePvP(PvPRequest) returns (PvPSession);
  rpc InitiateTrade(TradeOffer) returns (TradeSession);
}

// Universe & Discovery
service UniverseService {
  rpc GetStarSystem(SystemID) returns (StarSystem);
  rpc GetConnectedSystems(SystemID) returns (SystemList);
  rpc GetPlanet(PlanetID) returns (Planet);
  rpc GetUniverseEvents(Empty) returns (stream UniverseEvent);
}

// Admin
service AdminService {
  rpc GetServerMetrics(Empty) returns (Metrics);
  rpc BanPlayer(BanRequest) returns (Empty);
  rpc BroadcastMessage(Announcement) returns (Empty);
  rpc TriggerEvent(EventTrigger) returns (Event);
}
```

### State Synchronization Strategy

#### 1. Optimistic UI Updates
```go
// Client-side (SSH Gateway)
func (c *Client) BuyCommodity(commodityID string, quantity int) error {
    // Optimistically update local state
    c.localState.UpdateCargo(commodityID, quantity)
    c.localState.UpdateCredits(-estimatedCost)
    c.renderUI()

    // Send request to server
    resp, err := c.gameClient.BuyCommodity(ctx, &pb.TradeRequest{
        CommodityID: commodityID,
        Quantity: quantity,
    })

    if err != nil {
        // Rollback on error
        c.localState.Rollback()
        c.renderUI()
        return err
    }

    // Confirm with authoritative state
    c.localState.Sync(resp.PlayerState)
    c.renderUI()
    return nil
}
```

#### 2. Server-Authoritative State
- Game server is source of truth for all state
- Clients maintain local cache for rendering
- Server streams updates for critical state changes
- Periodic full state sync (every 30s or on screen change)

#### 3. Event Streaming
```go
// Server continuously streams events
stream, err := client.StreamPlayerUpdates(ctx, playerID)
for {
    update, err := stream.Recv()
    if err != nil {
        // Reconnect logic
        continue
    }

    switch update.Type {
    case "credits_changed":
        c.localState.UpdateCredits(update.NewCredits)
    case "location_changed":
        c.localState.UpdateLocation(update.NewLocation)
    case "combat_started":
        c.switchToCombatScreen(update.CombatSession)
    case "message_received":
        c.displayMessage(update.Message)
    }
    c.renderUI()
}
```

---

## Authentication & Authorization Flow

### Current Flow (Monolithic)
```
1. SSH connection → handlePasswordAuth()
2. Authenticate against DB
3. Return ssh.Permissions with player_id
4. startGameSession() extracts player_id
5. Load player data from DB
6. Initialize TUI with full state
```

### Proposed Flow (Client-Server)

```
┌──────────┐         ┌─────────────┐         ┌──────────────┐
│  Player  │         │ SSH Gateway │         │ Game Server  │
└────┬─────┘         └──────┬──────┘         └──────┬───────┘
     │                      │                       │
     │  ssh user@host       │                       │
     ├─────────────────────►│                       │
     │                      │                       │
     │                      │  AuthRequest          │
     │                      ├──────────────────────►│
     │                      │  (username, password) │
     │                      │                       │
     │                      │                       │  Verify
     │                      │                       │  against DB
     │                      │                       │
     │                      │  AuthResponse         │
     │                      │◄──────────────────────┤
     │                      │  (player_id, token)   │
     │                      │                       │
     │                      │  CreateSession        │
     │                      ├──────────────────────►│
     │                      │  (player_id, token)   │
     │                      │                       │
     │                      │                       │  Create
     │                      │                       │  Session
     │                      │                       │  Load State
     │                      │                       │
     │                      │  SessionCreated       │
     │                      │◄──────────────────────┤
     │                      │  (session_id, state)  │
     │                      │                       │
     │  Display Main Menu   │                       │
     │◄─────────────────────┤                       │
     │                      │                       │
     │                      │  StreamPlayerUpdates  │
     │                      ├──────────────────────►│
     │                      │  (session_id)         │
     │                      │                       │
     │                      │◄──────────stream──────┤
     │                      │                       │
```

### Token-Based Session Management

```go
// Session token structure
type SessionToken struct {
    PlayerID    uuid.UUID
    SessionID   uuid.UUID
    IssuedAt    time.Time
    ExpiresAt   time.Time
    Signature   []byte  // HMAC-SHA256
}

// Gateway caches session tokens
type SessionCache struct {
    tokens map[uuid.UUID]*SessionToken
    mu     sync.RWMutex
    ttl    time.Duration
}

// Game server validates tokens
type SessionValidator struct {
    secretKey []byte
    sessions  map[uuid.UUID]*PlayerSession
    mu        sync.RWMutex
}
```

---

## Data Ownership & Boundaries

### SSH Gateway Owns
- **SSH Connection State**: Active connections, terminal size
- **Input Buffering**: Keystroke handling, command parsing
- **Rendering State**: Screen buffers, cursor position
- **Auth Cache**: Short-lived session tokens (5 min TTL)
- **UI State**: Current screen, menu position, form inputs

### Game Server Owns
- **Player State**: Location, credits, stats, reputation
- **Ship State**: Hull, shields, cargo, equipment
- **Universe State**: Markets, NPCs, events
- **Session State**: Active sessions, last action time
- **Game Logic**: All calculations, AI, event processing
- **Persistence**: Database writes, transaction management

### Shared Concerns
- **Authentication**: Gateway checks cache, server validates
- **Session Management**: Gateway tracks connection, server tracks game state
- **Error Handling**: Both log errors, server determines game impact

---

## Migration Strategy

### Phase 1: Internal API Extraction (2-3 weeks)
**Goal**: Create API layer without splitting binaries

**Tasks**:
1. Define protobuf schemas for all game operations
2. Generate Go code from protobuf
3. Create internal API client interface
4. Implement server-side API handlers
5. Refactor TUI to use API client (in-process initially)
6. Maintain backward compatibility

**Deliverable**: Single binary with clean API boundary

**Benefits**:
- No deployment changes
- Easier testing
- Validates API design
- Identifies missing operations

### Phase 2: Service Extraction (2-3 weeks)
**Goal**: Split into separate binaries

**Tasks**:
1. Create `cmd/gateway/` SSH gateway server
2. Create `cmd/gameserver/` game logic server
3. Move managers to gameserver
4. Move TUI to gateway
5. Implement gRPC transport between services
6. Add service discovery / configuration
7. Update docker-compose.yml for multi-service

**Deliverable**: Two binaries (gateway + gameserver)

**Benefits**:
- Independent scaling
- Separate deployment
- Better resource isolation

### Phase 3: State Optimization (1-2 weeks)
**Goal**: Optimize state transfer and caching

**Tasks**:
1. Implement client-side state cache
2. Add delta updates for efficiency
3. Optimize protobuf messages
4. Add compression for large payloads
5. Implement optimistic UI updates
6. Add reconnection logic

**Deliverable**: Optimized client-server communication

**Benefits**:
- Reduced latency
- Better UX
- Resilient to network issues

### Phase 4: Scalability & Deployment (2-3 weeks)
**Goal**: Production-ready distributed deployment

**Tasks**:
1. Add session affinity / sticky sessions
2. Implement health checks
3. Add metrics and monitoring
4. Create Kubernetes manifests
5. Add horizontal pod autoscaling
6. Implement graceful shutdown
7. Add circuit breakers and retries

**Deliverable**: Production-ready distributed system

**Benefits**:
- Horizontal scaling
- High availability
- Production monitoring

---

## Implementation Checklist

### Prerequisites
- [ ] Create `api/proto/` directory
- [ ] Install protobuf compiler and Go plugin
- [ ] Add gRPC dependencies to go.mod
- [ ] Set up code generation in Makefile

### Phase 1: Internal API
- [ ] Define auth.proto service
- [ ] Define player.proto service
- [ ] Define game.proto service
- [ ] Define multiplayer.proto service
- [ ] Define universe.proto service
- [ ] Generate Go code from protos
- [ ] Create internal API client interface
- [ ] Implement server-side handlers
- [ ] Refactor TUI to use API client (in-process)
- [ ] Add integration tests for API
- [ ] Update CLAUDE.md with API patterns

### Phase 2: Service Split
- [ ] Create cmd/gateway/ entry point
- [ ] Create cmd/gameserver/ entry point
- [ ] Move TUI to gateway package
- [ ] Move managers to gameserver package
- [ ] Implement gRPC server in gameserver
- [ ] Implement gRPC client in gateway
- [ ] Add service configuration
- [ ] Update docker-compose.yml
- [ ] Create migration guide
- [ ] Test distributed deployment

### Phase 3: Optimization
- [ ] Implement client-side state cache
- [ ] Add protobuf delta encoding
- [ ] Optimize message sizes
- [ ] Add gRPC compression
- [ ] Implement optimistic updates
- [ ] Add reconnection logic
- [ ] Add connection pooling
- [ ] Performance testing

### Phase 4: Production
- [ ] Add session affinity configuration
- [ ] Implement health check endpoints
- [ ] Add Prometheus metrics
- [ ] Create Kubernetes manifests
- [ ] Set up HPA (Horizontal Pod Autoscaler)
- [ ] Add graceful shutdown handlers
- [ ] Implement circuit breakers
- [ ] Add distributed tracing
- [ ] Load testing
- [ ] Documentation updates

---

## Backward Compatibility

### During Migration
- Keep existing monolithic server buildable
- Support both architectures via build tags
- Provide migration guide for server operators
- No breaking changes to game data

### Long-Term Support
- Maintain monolithic option for simple deployments
- Document both deployment models
- Provide docker-compose for both architectures
- Keep single-binary option for development

---

## Testing Strategy

### Unit Testing
```go
// Test API handlers in isolation
func TestBuyCommodity(t *testing.T) {
    server := setupTestGameServer(t)
    resp, err := server.BuyCommodity(ctx, &pb.TradeRequest{
        PlayerID: testPlayerID,
        CommodityID: "food",
        Quantity: 10,
    })
    assert.NoError(t, err)
    assert.Equal(t, 10, resp.Inventory.Food)
}
```

### Integration Testing
```go
// Test gateway → gameserver communication
func TestEndToEndTrading(t *testing.T) {
    gameServer := startGameServer(t)
    gateway := startGateway(t, gameServer.Address())

    client := connectSSH(t, gateway.Address())

    // Simulate user buying commodity
    client.SendKeys("t", "b", "1", "10", "enter")

    // Verify state on game server
    state := gameServer.GetPlayerState(testPlayerID)
    assert.Equal(t, 10, state.Cargo["food"])
}
```

### Load Testing
- Simulate 1000+ concurrent SSH connections
- Measure latency under load
- Test horizontal scaling
- Validate session affinity

---

## Performance Considerations

### Expected Latency
- **Local RPCs**: < 1ms
- **Network RPCs**: 10-50ms (depending on distance)
- **Database queries**: 5-20ms
- **Total action latency**: 20-100ms (acceptable for turn-based game)

### Throughput Targets
- **SSH Gateways**: 1000 connections per instance
- **Game Servers**: 500 active sessions per instance
- **Database**: 10,000 queries per second
- **Chat messages**: 100 messages per second

### Resource Requirements

**SSH Gateway** (per instance):
- CPU: 2 cores
- RAM: 2 GB
- Network: 100 Mbps
- Storage: Minimal (logs only)

**Game Server** (per instance):
- CPU: 4 cores
- RAM: 8 GB (session state)
- Network: 1 Gbps
- Storage: Minimal (cache only)

---

## Security Considerations

### Authentication
- Gateway validates SSH credentials
- Game server validates session tokens
- Tokens expire after 24 hours
- Refresh tokens on activity

### Authorization
- Game server enforces permissions
- Gateway only routes authenticated requests
- RBAC implemented on game server
- Admin actions require elevated tokens

### Network Security
- mTLS between gateway and game server
- Encrypted protobuf messages
- No sensitive data in gateway logs
- Rate limiting at gateway

---

## Monitoring & Observability

### Metrics to Track
- **Gateway**: Connections, auth failures, request latency
- **Game Server**: Active sessions, RPC latency, DB queries
- **Database**: Connection pool usage, query performance
- **Business**: Active players, trades/hour, combat events

### Logging
- Structured JSON logs
- Request tracing with correlation IDs
- Error aggregation
- Audit logs for admin actions

### Alerting
- Gateway connection failures > 5%
- Game server RPC latency > 200ms
- Database connection pool exhausted
- Active sessions > 80% capacity

---

## Future Enhancements

### After Initial Migration
1. **WebSocket Gateway**: Support web browser clients
2. **GraphQL API**: For web dashboard
3. **Mobile Client**: Native terminal app
4. **Region Sharding**: Split universe by regions
5. **Redis Cache**: Reduce database load
6. **Message Queue**: Async event processing
7. **CDC (Change Data Capture)**: Real-time analytics

### Alternative Clients
- **Web Client**: Browser-based terminal emulator
- **Native Client**: Electron or Qt app
- **Mobile Client**: iOS/Android terminal app
- **Discord Bot**: Basic info queries

---

## Decision Log

### Why gRPC over REST?
- Bidirectional streaming essential for real-time updates
- Better performance for high-frequency operations
- Strong typing prevents API drift
- Built-in code generation

### Why Not Microservices?
- Game logic is tightly coupled (combat needs trading data)
- Overhead of distributed transactions
- Complexity not justified for current scale
- Can split further if needed later

### Why Keep PostgreSQL?
- Already working well
- ACID transactions critical for game state
- Complex queries benefit from SQL
- Can add caching layer later if needed

### Why Not Event Sourcing?
- Added complexity for uncertain benefit
- Game state changes are deterministic
- Snapshots + audit log sufficient
- Can migrate later if needed

---

## Glossary

- **SSH Gateway**: Frontend server handling SSH connections
- **Game Server**: Backend server executing game logic
- **Session**: Authenticated player connection
- **State Cache**: Client-side cache of server state
- **Delta Update**: Incremental state change
- **Session Affinity**: Routing player to same game server
- **Optimistic Update**: UI update before server confirmation

---

## References

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [Escape Velocity Game Design](https://en.wikipedia.org/wiki/Escape_Velocity_(video_game))
- [BubbleTea Documentation](https://github.com/charmbracelet/bubbletea)

---

**Next Steps**:
1. Review this document with team
2. Validate API design with use cases
3. Create Phase 1 implementation plan
4. Update ROADMAP.md with Phase 9
5. Begin protobuf schema definitions
