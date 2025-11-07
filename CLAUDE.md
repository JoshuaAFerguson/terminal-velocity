# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terminal Velocity is a multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH. Players navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions—all within a terminal UI built with BubbleTea.

**Tech Stack**: Go 1.24+, PostgreSQL (pgx/v5), BubbleTea + Lipgloss (TUI), golang.org/x/crypto/ssh

**Current Phase**: Phase 1 - Foundation & Navigation (see ROADMAP.md for full development plan)

## Building and Running

### Development Commands

```bash
# Build and run server
make build          # Build server binary
make run            # Run server in development mode
make build-tools    # Build genmap and accounts utilities

# Database
make setup-db       # Initialize PostgreSQL schema (requires psql client)

# Testing and quality
make test           # Run tests with race detector and coverage
make coverage       # View test coverage in browser
make lint           # Run golangci-lint
make fmt            # Format code with gofmt
make vet            # Run go vet

# Tools
make genmap         # Generate and preview universe (100 systems)
./genmap -systems 50 -stats              # Custom universe generation
./accounts create <username> <email>     # Create player account
./accounts add-key <username> <key-file> # Add SSH key to account
```

### Docker Development

```bash
make docker-compose-up      # Start full stack (PostgreSQL + server)
make docker-compose-down    # Stop stack
make docker-compose-logs    # View logs
make docker-compose-restart # Restart services
```

### Connecting to Server

```bash
ssh -p 2222 username@localhost
```

## Architecture

### Directory Structure

- `cmd/` - Executable entry points
  - `server/` - Main SSH game server
  - `genmap/` - Universe generation CLI tool
  - `accounts/` - Account management tool
- `internal/` - Private application code
  - `server/` - SSH server, authentication, session management
  - `tui/` - BubbleTea UI screens (main menu, navigation, registration, etc.)
  - `database/` - Repository pattern for data access (PlayerRepository, SystemRepository, SSHKeyRepository)
  - `models/` - Core data models (Player, Ship, StarSystem, Planet, etc.)
  - `game/` - Game logic modules
    - `universe/` - Procedural universe generation with MST-based jump routes
    - `combat/`, `trading/`, `faction/`, `mission/`, `ship/` - Future game systems
  - `ui/` - Reusable UI components (future)
- `scripts/` - Database schema and migrations
- `configs/` - YAML configuration files

### Key Architectural Patterns

**Repository Pattern**: All database access goes through repositories in `internal/database/`. Each repository provides typed CRUD operations and encapsulates SQL queries.

**BubbleTea MVC**: The TUI uses BubbleTea's Model-View-Update architecture:
- Each screen has its own model (e.g., `navigationModel`, `mainMenuModel`)
- Update functions handle messages (keyboard input, async results)
- View functions render the current state
- Screens are composed in `internal/tui/model.go` with a `Screen` enum for routing

**SSH Integration**: The server runs BubbleTea programs directly over SSH channels:
- `internal/server/server.go` handles SSH authentication (password + public key)
- Session permissions pass player_id from auth to game session
- Each SSH connection gets its own BubbleTea program instance

### Authentication Flow

1. SSH connection → `handlePasswordAuth` or `handlePublicKeyAuth`
2. Authenticate against database (PlayerRepository or SSHKeyRepository)
3. Return `ssh.Permissions` with player_id in Extensions
4. `startGameSession` extracts player_id, loads player data
5. Initialize TUI model with repositories and player state
6. Run BubbleTea program with SSH channel as I/O

New users can register if `AllowRegistration` is enabled in server config.

### Database Schema

PostgreSQL with UUID primary keys. Key tables:
- `players` - User accounts (nullable password_hash for SSH-only accounts)
- `player_ssh_keys` - SSH public keys with SHA256 fingerprints
- `player_reputation` - Reputation with NPC factions (-100 to +100)
- `star_systems` - Systems with position, tech level, government
- `system_connections` - Jump routes (bidirectional)
- `planets` - Planets with services array
- `ships` - Player ships with hull, shields, fuel, cargo

See `scripts/schema.sql` for full schema.

### Universe Generation

The `internal/game/universe/` package generates procedural universes:
- Systems placed using spiral galaxy distribution
- Jump routes created via Minimum Spanning Tree (Prim's algorithm) + extra connections
- Tech levels distributed radially (high in core, low at edges)
- 6 NPC factions (governments) assigned to systems
- Planets generated per system with randomized services

Use `cmd/genmap/` to preview generated universes before importing to database.

## Common Development Tasks

### Adding a New TUI Screen

1. Create `internal/tui/screenname.go` with:
   - Model struct (e.g., `screenNameModel`)
   - `updateScreenName(msg tea.Msg)` method on Model
   - `viewScreenName()` method on Model
2. Add screen to `Screen` enum in `internal/tui/model.go`
3. Add case to Update() switch for routing
4. Add case to View() switch for rendering
5. Initialize screen state in `newScreenNameModel()` if needed
6. Add menu item in `internal/tui/main_menu.go` if accessible from main menu

### Adding a Database Table

1. Add table to `scripts/schema.sql`
2. Create migration in `scripts/migrations/NNN_description.sql`
3. Add Go model to `internal/models/`
4. Create repository in `internal/database/` with CRUD methods
5. Add repository field to server and TUI models
6. Initialize repository in `NewServer()` and `NewModel()`

### Working with Player State

Player data flows: Database → PlayerRepository → TUI Model → Screen

To update player state:
1. Modify via repository method (e.g., `playerRepo.UpdateLocation()`)
2. Update local `m.player` in TUI model
3. Call `tea.Cmd` to refresh data if needed

### Testing Universe Generation

```bash
# Generate 50-system universe and show statistics
./genmap -systems 50 -stats

# Preview system connectivity
./genmap -systems 100 -preview

# Generate and save to database (future feature)
./genmap -systems 100 -save
```

## Code Conventions

### Error Handling

Return typed errors from repositories:
```go
var ErrPlayerNotFound = errors.New("player not found")
var ErrSSHKeyNotFound = errors.New("ssh key not found")
```

Check for specific errors in callers, return wrapped errors up the stack.

### Database Context

Always pass `context.Context` as first parameter to repository methods. Use `context.Background()` in async tea.Cmd functions.

### UUID Handling

Use `github.com/google/uuid` for all IDs. Handle nullable UUIDs with pointers:
```go
CurrentPlanet *uuid.UUID `json:"current_planet,omitempty"` // nil if in space
```

### BubbleTea Messages

Define custom message types for async operations:
```go
type dataLoadedMsg struct {
    data *Data
    err  error
}
```

Return these from tea.Cmd functions, handle in Update().

### SQL Queries

Use parameterized queries ($1, $2, etc.) to prevent SQL injection. Handle nullable columns with `sql.NullString`, `sql.NullInt64`, etc.

## Development Notes

### Current State (as of Phase 1)

✅ **Implemented**:
- SSH server with password + public key authentication
- User registration flow (interactive TUI)
- Account management CLI tool
- Database layer with repositories
- BubbleTea TUI framework (main menu, registration, navigation screens)
- Universe generation (procedural systems, MST jump routes)
- Navigation system (jump between systems, fuel cost calculation)
- Location persistence

⏳ **Partial**:
- Ship repository (models exist, repository not implemented)
- Fuel consumption (calculated but not persisted to database)
- Player ship loading (currentShip is nil in TUI)

❌ **Not Implemented**:
- Trading system
- Combat system
- Mission system
- Player factions
- Chat/messaging
- Most game screens (Trading, Shipyard, Missions, Settings)

### Known Limitations

- No ship repository yet - ship data not loaded or updated
- Fuel cost calculated but not deducted from ship fuel
- No jump animation or travel time
- No landing/docking UI for planets
- Registration creates account but requires reconnection (no live auth transition)

### Testing

Only 2 test files exist currently. When adding tests:
- Place `*_test.go` files alongside source
- Use `testify/assert` for assertions
- Test critical paths: auth, database operations, universe generation
- Run with race detector: `go test -race`

## Phase 1 Completion Checklist

Refer to GitHub Issues with `phase-1` label and ROADMAP.md Phase 1 section for remaining tasks.
