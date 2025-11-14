# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terminal Velocity is a multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH. Players navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions—all within a terminal UI built with BubbleTea.

**Tech Stack**: Go 1.24+, PostgreSQL (pgx/v5), BubbleTea + Lipgloss (TUI), golang.org/x/crypto/ssh

**Current Phase**: Phases 0-7 Complete - Feature Complete (see ROADMAP.md for full development history)

**Codebase Stats**:
- ~37,000 lines of Go code
- 26 distinct UI screens
- 28 TUI component files
- 7 database repositories
- 101 Go files across all packages

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
make lint           # Run golangci-lint (see .golangci.yml for config)
make fmt            # Format code with gofmt
make vet            # Run go vet

# Development helpers
make dev-setup      # Complete development environment setup
make watch          # Watch for changes and auto-rebuild (requires entr)
make release        # Build cross-platform release binaries

# Tools
make genmap         # Generate and preview universe (100 systems)
./genmap -systems 50 -stats              # Custom universe generation
./accounts create <username> <email>     # Create player account
./accounts add-key <username> <key-file> # Add SSH key to account
```

### Docker Development

```bash
make docker compose-up      # Start full stack (PostgreSQL + server)
make docker compose-down    # Stop stack
make docker compose-logs    # View logs
make docker compose-restart # Restart services
make docker-clean           # Remove all Docker artifacts and volumes
```

### Connecting to Server

```bash
ssh -p 2222 username@localhost
```

## Architecture

### Directory Structure

- `cmd/` - Executable entry points (3 commands)
  - `server/` - Main SSH game server
  - `genmap/` - Universe generation CLI tool
  - `accounts/` - Account management tool
- `internal/` - Private application code (28 packages)
  - `server/` - SSH server, authentication, session management
  - `tui/` - BubbleTea UI screens (26 screens, 28 component files)
    - Screens: MainMenu, Game, Navigation, Trading, Cargo, Shipyard, Outfitter, ShipManagement, Combat, Missions, Achievements, Encounter, News, Leaderboards, Players, Chat, Factions, Trade, PvP, Help, OutfitterEnhanced, Settings, Admin, Tutorial, Quests, Registration
  - `database/` - Repository pattern for data access (7 repositories)
    - PlayerRepository, SystemRepository, SSHKeyRepository, ShipRepository, MarketRepository
    - connection.go - Database connection pooling (pgx)
    - migrations.go - Schema migration support
  - `models/` - Core data models (Player, Ship, StarSystem, Planet, Quest, Event, etc.)
  - `game/` - Core game logic packages
    - `universe/` - Procedural universe generation with MST-based jump routes
    - `trading/` - Trading calculations and economy logic
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
  - `encounters/` - Random encounter system (templates, types, generation)
  - `outfitting/` - Equipment & loadout management (6 slot types, 16 items)
  - `settings/` - Player settings (6 categories, JSON persistence)
  - `help/` - Help content and context system
  - `tutorial/` - Tutorial & onboarding (7 categories, 20+ steps)
  - `admin/` - Server administration (RBAC, 20+ permissions)
  - `session/` - Session management & auto-save (30s autosave)
  - `errors/` - Error handling, metrics, and retry logic
  - `logger/` - Centralized logging infrastructure
- `scripts/` - Database schema and migrations
  - `schema.sql` - Main database schema
  - `migrations/` - Database migration scripts
- `configs/` - YAML configuration files
  - `config.example.yaml` - Example server configuration
- `docs/` - Additional documentation
- `.github/` - GitHub Actions CI/CD workflows

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

### File Headers

All Go source files include a standard header comment block:
```go
// File: internal/package/filename.go
// Project: Terminal Velocity
// Description: Brief description of file purpose
// Version: X.Y.Z
// Author: Joshua Ferguson
// Created: YYYY-MM-DD
```

When modifying a file, increment the version number:
- **Patch** (X.Y.Z+1): Bug fixes, minor changes
- **Minor** (X.Y+1.0): New features, significant changes
- **Major** (X+1.0.0): Breaking changes, major refactoring

### Linting Configuration

The project uses `golangci-lint` with configuration in `.golangci.yml`:

**Enabled linters**:
- errcheck, gosimple, govet, ineffassign, staticcheck
- typecheck, gofmt, goimports, misspell, dupl

**Disabled linters** (with reasoning):
- unparam, unused, gosec, gocritic, unconvert, gocyclo
- TUI files exempt from errcheck (UI flow doesn't require strict error checking)
- Game mechanics use math/rand appropriately (gosec G404 excluded)

**Settings**:
- Cyclomatic complexity: 40 (relaxed for complex game logic)
- Duplication threshold: 150 lines
- Test files exempt from dupl, gosec, gocyclo

Run `make lint` before committing. Files must pass lint checks.

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
- 26 BubbleTea UI screens across 28 component files
- 7 database repositories with connection pooling
- Thread-safe concurrency (sync.RWMutex throughout)
- Background workers (event scheduling, metrics, cleanup)
- 100+ star systems with MST jump routes
- ~37,000 lines of Go code
- Comprehensive error handling with retry logic
- Centralized logging infrastructure
- File header comments with version tracking

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

**Current Test Coverage**:
- 5 test files in codebase
- Focus on critical database operations
- Player repository tests
- Error handling and retry logic tests
- Metrics testing

**When adding tests**:
- Place `*_test.go` files alongside source
- Use `testify/assert` for assertions (if available) or standard library testing
- Test critical paths: auth, database operations, game systems
- Run with race detector: `go test -race`
- Focus on integration testing across multiple systems
- **Note**: Testing coverage is a priority for Phase 8

## Key Dependencies

**Core Libraries**:
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling
- `github.com/jackc/pgx/v5` v5.7.6 - PostgreSQL driver with connection pooling
- `github.com/google/uuid` v1.6.0 - UUID generation
- `golang.org/x/crypto` v0.37.0 - SSH server and password hashing
- `golang.org/x/term` v0.31.0 - Terminal utilities

**Development Requirements**:
- Go 1.24+
- PostgreSQL 12+
- Docker & Docker Compose (for containerized development)
- golangci-lint (for linting)
- entr (optional, for `make watch`)

## Common Patterns & Gotchas

### Thread Safety
- All manager packages (chat, events, factions, etc.) use `sync.RWMutex`
- Always lock before reading/writing shared state
- Use `RLock()`/`RUnlock()` for read operations
- Use `Lock()`/`Unlock()` for write operations

### BubbleTea Message Flow
- Never block in `Update()` - return tea.Cmd for async operations
- Use custom message types for async results
- Handle context cancellation in long-running operations
- Screen transitions happen via `m.screen = ScreenName`

### Database Operations
- Always pass `context.Context` as first parameter
- Use `context.Background()` in tea.Cmd functions
- Handle `pgx` specific errors (not database/sql)
- UUID fields may be nil (use pointers for optional IDs)
- Connection pool managed in `database.Connection`

### Error Handling
- Use `internal/errors` package for retry logic and metrics
- Return wrapped errors for better stack traces
- Define package-level errors (e.g., `ErrPlayerNotFound`)
- Log errors using `internal/logger` package

### Common Pitfalls
- Don't forget to initialize sub-models in TUI Model
- Repository methods must handle nil/empty results gracefully
- BubbleTea programs need proper cleanup on exit
- SSH authentication requires both password and public key support
- Market prices use float64 - be careful with precision
- Ship cargo capacity is in tons, not individual units

## Development Best Practices

**Before committing**:
- Lint files: `make lint` - files must pass lint checks
- Run tests: `make test`
- Format code: `make fmt`
- Update CHANGELOG.md after changes are made
- Check if ROADMAP.md needs updates before commit
- Check if README.md needs updates before commits
- Increment version numbers in changelog accordingly
- Always add comment headers to code files
- Thoroughly comment the code
- Increment file version numbers in header after changes

**Git Workflow**:
- Feature branches for all changes
- Descriptive commit messages
- Reference issue numbers in commits when applicable
- Keep commits focused and atomic

**Code Quality**:
- Follow Go best practices and idioms
- Keep functions under 50 lines when possible
- Document all exported functions and types
- Use meaningful variable names
- Avoid global state when possible

The github repository is located at https://github.com/JoshuaAFerguson/terminal-velocity/

## Current Focus (Phase 8)

Phases 0-7 are complete! Current priorities:
1. **Integration Testing**: Ensure all 29+ systems work together seamlessly
2. **Balance Tuning**: Economy, combat, and progression adjustments
3. **Performance Optimization**: Database indexing, caching, load testing
4. **Community Testing**: Gather feedback from players
5. **Launch Preparation**: Deployment, monitoring, community management

## Troubleshooting

### Common Issues

**Build Errors**:
- Ensure Go 1.24+ is installed: `go version`
- Run `go mod download` to fetch dependencies
- Check that all imports are correct

**Database Connection**:
- Verify PostgreSQL is running: `docker compose ps`
- Check connection string in config.yaml
- Ensure database schema is initialized: `make setup-db`
- Check pgx connection pool settings

**Linting Failures**:
- Run `make fmt` to auto-format code
- Check `.golangci.yml` for exemptions
- TUI files are exempt from errcheck
- Some test files have relaxed requirements

**SSH Connection Issues**:
- Verify server is running on port 2222
- Check SSH host key generation in configs/
- Ensure user exists in database
- For public key auth, verify fingerprint matches

**Docker Issues**:
- Run `make docker-clean` to reset everything
- Check `.env` file exists and has DB_PASSWORD set
- Ensure ports 2222 and 5432 are available
- View logs: `make docker compose-logs`

## Quick Reference

### Key Files to Know
- `internal/tui/model.go` - Main TUI model and screen routing
- `internal/server/server.go` - SSH server and authentication
- `internal/database/connection.go` - Database connection setup
- `scripts/schema.sql` - Complete database schema
- `Makefile` - All build and dev commands
- `.golangci.yml` - Linting configuration

### Important Constants
- Default SSH port: 2222
- Auto-save interval: 30 seconds
- Maximum active missions: 5
- Player online status timeout: 5 minutes
- Combat AI difficulty levels: 5 (Easy to Ace)
- Ship types: 11 (Shuttle to Battleship)
- Commodities: 15
- Weapon types: 9

### Manager Packages (All Thread-Safe)
Each manager package follows similar patterns:
- `New()` - Constructor
- `Start()` - Initialize background workers
- `Stop()` - Cleanup
- `sync.RWMutex` - Thread safety
- Background goroutines for periodic tasks

Managers: achievements, admin, chat, encounters, events, factions, leaderboards, missions, news, presence, pvp, quests, session, territory, trade, tutorial

### Screen Navigation Pattern
```go
// In TUI Update()
case key.Matches(msg, m.keys.someKey):
    m.screen = ScreenDesired
    return m, nil

// In TUI View()
switch m.screen {
case ScreenDesired:
    return m.viewDesiredScreen()
}
```

### Adding a New Feature Checklist
1. ✅ Create package in `internal/` if needed
2. ✅ Define models in `internal/models/`
3. ✅ Add repository if database access needed
4. ✅ Create manager if background tasks needed
5. ✅ Add TUI screen in `internal/tui/`
6. ✅ Update Screen enum and routing
7. ✅ Add database tables/migrations if needed
8. ✅ Write tests
9. ✅ Update documentation (README, CHANGELOG, CLAUDE.md)
10. ✅ Run `make lint` and `make test`

---

**Last Updated**: 2025-01-14
**Document Version**: 2.0.0