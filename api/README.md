# Terminal Velocity API

This directory contains the Protocol Buffer (protobuf) definitions for Terminal Velocity's client-server API.

## Overview

As part of Phase 9 architecture refactoring, Terminal Velocity is transitioning from a monolithic architecture to a client-server model using gRPC. This API layer will enable:

- Horizontal scaling of SSH gateways and game servers independently
- Support for multiple client types (SSH, native terminal, web)
- Clean separation between presentation (TUI) and business logic (game state)

See [docs/ARCHITECTURE_REFACTORING.md](../docs/ARCHITECTURE_REFACTORING.md) for the complete architectural design.

## Directory Structure

```
api/
├── proto/              # Protocol Buffer definitions (.proto files)
│   ├── common.proto    # Common types (UUID, Timestamp, Coordinates, etc.)
│   ├── auth.proto      # Authentication & session management
│   ├── player.proto    # Player state management
│   └── game.proto      # Core game actions (trading, navigation, etc.)
└── gen/
    └── go/
        └── v1/         # Generated Go code (created by `make proto`)
```

## Prerequisites

### 1. Install Protocol Buffers Compiler (protoc)

#### macOS
```bash
brew install protobuf
```

#### Ubuntu/Debian
```bash
apt-get update && apt-get install -y protobuf-compiler
```

#### From source (all platforms)
```bash
# Download latest release from https://github.com/protocolbuffers/protobuf/releases
wget https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
unzip protoc-25.1-linux-x86_64.zip -d /usr/local
```

Verify installation:
```bash
protoc --version
# Should output: libprotoc 3.x.x or higher
```

### 2. Install Go Plugins for Protobuf

```bash
make install-proto-tools
```

This installs:
- `protoc-gen-go` - Generates Go structs from protobuf messages
- `protoc-gen-go-grpc` - Generates Go gRPC service code

## Generating Code

### Generate Go Code from Protobuf

```bash
make proto
```

This command:
1. Creates the output directory `api/gen/go/v1/`
2. Generates Go structs for all message types
3. Generates Go gRPC service interfaces and client/server code

Generated files will appear in `api/gen/go/v1/`:
- `common.pb.go` - Common types
- `auth.pb.go` - Auth message types
- `auth_grpc.pb.go` - Auth gRPC service code
- `player.pb.go` - Player message types
- `player_grpc.pb.go` - Player gRPC service code
- `game.pb.go` - Game message types
- `game_grpc.pb.go` - Game gRPC service code

### Clean Generated Code

```bash
make proto-clean
```

## Services Overview

### AuthService (`auth.proto`)
Handles player authentication and session management.

**Key RPCs**:
- `Authenticate` - Password-based login
- `AuthenticateSSH` - SSH public key authentication
- `CreateSession` - Start new game session
- `ValidateSession` - Check session validity
- `EndSession` - Terminate session
- `Register` - Create new player account

### PlayerService (`player.proto`)
Manages player state and real-time updates.

**Key RPCs**:
- `GetPlayerState` - Full player state (location, credits, ship, inventory)
- `UpdatePlayerLocation` - Move player
- `GetPlayerShip` - Ship details and equipment
- `GetPlayerInventory` - Cargo and items
- `GetPlayerStats` - Statistics and ratings
- `GetPlayerReputation` - Faction reputation
- `StreamPlayerUpdates` - Real-time state changes (streaming)

### GameService (`game.proto`)
Core game actions and commands.

**Key RPCs**:
- **Navigation**: `Jump`, `Land`, `Takeoff`
- **Trading**: `GetMarket`, `BuyCommodity`, `SellCommodity`
- **Ship Management**: `BuyShip`, `SellShip`, `BuyOutfit`, `SellOutfit`
- **Missions**: `GetAvailableMissions`, `AcceptMission`, `AbandonMission`
- **Quests**: `GetAvailableQuests`, `AcceptQuest`, `GetActiveQuests`

## Development Workflow

### 1. Modify Protobuf Definitions

Edit `.proto` files in `api/proto/`:
```bash
vim api/proto/player.proto
```

### 2. Regenerate Code

```bash
make proto
```

### 3. Use Generated Code

Import in your Go code:
```go
import pb "github.com/JoshuaAFerguson/terminal-velocity/api/gen/go/v1"

// Example: Create a player state
state := &pb.PlayerState{
    PlayerId: &pb.UUID{Value: playerID.String()},
    Username: "player1",
    Credits: 10000,
}
```

### 4. Implement Services

Server-side:
```go
type gameServer struct {
    pb.UnimplementedGameServiceServer
}

func (s *gameServer) Jump(ctx context.Context, req *pb.JumpRequest) (*pb.JumpResponse, error) {
    // Implementation
}
```

Client-side:
```go
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := pb.NewGameServiceClient(conn)

resp, err := client.Jump(ctx, &pb.JumpRequest{
    PlayerId: playerID,
    TargetSystemId: systemID,
})
```

## API Design Principles

### Server-Authoritative
- Game server is source of truth for all state
- Clients cache state locally for rendering
- Server validates all actions before applying
- Prevents cheating and ensures consistency

### Optimistic Updates
- Clients can update UI optimistically before server confirmation
- Rollback on server rejection
- Provides responsive UX despite network latency

### Streaming for Real-Time Updates
- Use bidirectional streaming for frequently changing state
- Example: `StreamPlayerUpdates` pushes location, credits, ship changes
- Reduces polling overhead
- Enables real-time multiplayer interactions

### Backwards Compatibility
- Use optional fields for extensibility
- Add new fields without breaking existing clients
- Version services (v1, v2) when making breaking changes

## Testing

### Unit Testing Protocol Buffers

```go
func TestPlayerStateMarshaling(t *testing.T) {
    state := &pb.PlayerState{
        PlayerId: &pb.UUID{Value: "test-id"},
        Credits: 1000,
    }

    // Marshal to bytes
    data, err := proto.Marshal(state)
    require.NoError(t, err)

    // Unmarshal
    var decoded pb.PlayerState
    err = proto.Unmarshal(data, &decoded)
    require.NoError(t, err)

    assert.Equal(t, state.Credits, decoded.Credits)
}
```

### Testing gRPC Services

Use `google.golang.org/grpc/test/bufconn` for in-memory gRPC testing:

```go
func setupTestServer(t *testing.T) *grpc.ClientConn {
    lis := bufconn.Listen(1024 * 1024)
    s := grpc.NewServer()
    pb.RegisterGameServiceServer(s, &gameServer{})

    go s.Serve(lis)

    conn, _ := grpc.DialContext(ctx, "",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithInsecure(),
    )

    return conn
}
```

## Migration Status

### Phase 1: Internal API Extraction (Current)
- ✅ Protobuf schemas defined
- ✅ Makefile configured for code generation
- ⏳ API client interface (pending)
- ⏳ In-process server implementation (pending)
- ⏳ TUI refactoring to use API (pending)

### Phase 2: Service Split (Future)
- Split into `cmd/gateway/` and `cmd/gameserver/`
- Implement gRPC transport between services
- Service discovery and configuration

### Phase 3: Optimization (Future)
- Client-side state caching
- Delta updates
- Connection pooling
- Optimistic UI updates

### Phase 4: Production (Future)
- Kubernetes deployment
- Horizontal pod autoscaling
- Monitoring and observability
- Load testing

## Resources

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [Terminal Velocity Architecture Design](../docs/ARCHITECTURE_REFACTORING.md)

## Contributing

When adding new RPCs or modifying the API:

1. **Design First**: Discuss API changes in GitHub issues
2. **Update Protobuf**: Modify `.proto` files
3. **Regenerate Code**: Run `make proto`
4. **Update Documentation**: Document new services/messages in this README
5. **Test**: Add unit tests for new message types
6. **Commit**: Commit both `.proto` files and generated code

## Troubleshooting

### `protoc: command not found`
Install the Protocol Buffers compiler (see Prerequisites above).

### `protoc-gen-go: program not found`
Run `make install-proto-tools` to install Go plugins.

### Import errors in generated code
Ensure your `go_package` option matches your module path:
```protobuf
option go_package = "github.com/JoshuaAFerguson/terminal-velocity/api/gen/go/v1";
```

### Generated files not found
Run `make proto` to generate code from `.proto` files.

---

**Status**: Phase 1 implementation in progress
**Last Updated**: 2025-01-14
**Version**: 1.0.0
