# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terminal Velocity is a multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH. Players navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions—all within a terminal UI built with BubbleTea.

**Tech Stack**: Go 1.23+, PostgreSQL (pgx/v5), BubbleTea + Lipgloss (TUI), golang.org/x/crypto/ssh

**Current Phase**: Phases 0-7 Complete - Feature Complete (see ROADMAP.md for full development history)

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
  - `tui/` - BubbleTea UI screens (30+ screens for all features)
  - `database/` - Repository pattern for data access (20+ repositories)
  - `models/` - Core data models (Player, Ship, StarSystem, Planet, Quest, Event, etc.)
  - `combat/` - Turn-based combat system with AI (5 difficulty levels)
  - `missions/` - Mission lifecycle management (4 mission types)
  - `quests/` - Quest & storyline system (7 types, 12 objectives)
  - `events/` - Dynamic events manager (10 event types)
  - `achievements/` - Achievement tracking system
  - `news/` - News generation system (10+ event types)
  - `leaderboards/` - Player rankings (4 categories)
  - `chat/` - Multiplayer chat (4 channels)
  - `factions/` - Player faction system with treasury
  - `territory/` - Territory control & passive income
  - `trade/` - Player-to-player trading with escrow
  - `pvp/` - PvP combat system
  - `presence/` - Player presence tracking
  - `encounters/` - Random encounter system
  - `outfitting/` - Equipment & loadout management (6 slot types, 16 items)
  - `settings/` - Player settings (6 categories, JSON persistence)
  - `tutorial/` - Tutorial & onboarding (7 categories, 20+ steps)
  - `admin/` - Server administration (RBAC, 20+ permissions)
  - `session/` - Session & auto-save (30s autosave)
  - `universe/` - Procedural universe generation with MST-based jump routes
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

### Current State (Phases 0-7 Complete)

✅ **Fully Implemented** (29+ interconnected systems):

**Core Gameplay**:
- SSH server with password + public key authentication
- Dynamic trading economy (15 commodities, supply/demand)
- Ship progression (11 ship types: Shuttle → Battleship)
- Turn-based combat with tactical AI (5 difficulty levels)
- Advanced ship outfitting (6 slot types, 16 equipment items)
- Loadout system (save/load/clone configurations)
- Reputation system (6 NPC factions, bounty tracking)
- Loot & salvage (4 rarity tiers, rare item drops)

**Content Systems**:
- Quest & storyline system (7 quest types, 12 objective types, branching narratives)
- Mission system (4 mission types with progress tracking)
- Dynamic events (10 event types, leaderboards, progress rewards)
- Achievements system (milestone tracking)
- Random encounters (pirates, traders, police, distress calls)
- News system (10+ event types, dynamic generation)

**Multiplayer Features**:
- Chat system (4 channels: global, system, faction, DM)
- Player presence tracking (real-time locations)
- Player factions (treasury, member management, ranks)
- Territory control (system claiming, passive income)
- Player-to-player trading (escrow system)
- PvP combat (consensual duels, faction wars)
- Leaderboards (4 categories: credits, combat, trade, exploration)

**Infrastructure & Polish**:
- Server administration (RBAC with 4 roles, 20+ permissions)
- Moderation tools (ban/mute with expiration)
- Audit logging (10,000 entry buffer)
- Session management (auto-save every 30 seconds)
- Server-authoritative architecture
- Interactive tutorial (7 categories, 20+ steps)
- Settings system (6 categories, 5 color schemes, JSON persistence)
- Server metrics & monitoring

**Technical**:
- 30+ BubbleTea UI screens
- 20+ database repositories
- Thread-safe concurrency (sync.RWMutex throughout)
- Background workers (event scheduling, metrics, cleanup)
- 100+ star systems with MST jump routes

### Known Limitations & Future Enhancements

Current limitations:
- No two-factor authentication
- No password reset functionality
- No web dashboard (SSH only)
- No modding support yet

Future enhancements (Phase 8+):
- Integration testing across all systems
- Balance tuning (economy, combat, progression)
- Performance optimization (indexing, caching, load testing)
- Community testing and feedback gathering
- Additional content (more quests, events, ships)
- Advanced features (player stations, mining, manufacturing)

### Testing

When adding tests:
- Place `*_test.go` files alongside source
- Use `testify/assert` for assertions
- Test critical paths: auth, database operations, game systems
- Run with race detector: `go test -race`
- Focus on integration testing across multiple systems

## Development Best Practices

Before committing:
- Lint files: `make lint` - files must pass lint checks
- Run tests: `make test`
- Update CHANGELOG.md after changes are made
- Check if ROADMAP.md needs updates before commit
- Check if README.md needs updates before commits
- Increment version numbers in changelog accordingly
- Always add comment headers to code files
- Thoroughly comment the code
- Increment file version numbers in header after changes

The github repository is located at https://github.com/JoshuaAFerguson/terminal-velocity/

## Current Focus (Phase 8)

Phases 0-7 are complete! Current priorities:
1. **Integration Testing**: Ensure all 29+ systems work together seamlessly
2. **Balance Tuning**: Economy, combat, and progression adjustments
3. **Performance Optimization**: Database indexing, caching, load testing
4. **Community Testing**: Gather feedback from players
5. **Launch Preparation**: Deployment, monitoring, community management