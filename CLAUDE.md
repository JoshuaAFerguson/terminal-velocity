# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terminal Velocity is a multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH. Players navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions—all within a terminal UI built with BubbleTea.

**Tech Stack**: Go 1.24+, PostgreSQL (pgx/v5), BubbleTea + Lipgloss (TUI), golang.org/x/crypto/ssh

**Current Phase**: Phases 0-8 Complete - Feature Complete with Enhanced UI Integration (see ROADMAP.md for full development history)

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
./genmap -systems 100 -save              # Generate and save to database
./accounts create <username> <email>     # Create player account
./accounts add-key <username> <key-file> # Add SSH key to account
```

### First-Time Server Setup

**Quick Start** (recommended):
```bash
# 1. Build tools
make build-tools

# 2. Run initialization script (sets up database + universe)
./scripts/init-server.sh

# 3. The script will:
#    - Create database and user
#    - Initialize schema
#    - Generate and populate universe (100 systems)
#    - Display connection instructions

# 4. Create your player account
./accounts create <username> <email>

# 5. Start the server
make run
```

**Manual Setup**:
```bash
# 1. Create database
psql -U postgres -c "CREATE DATABASE terminal_velocity;"
psql -U postgres -c "CREATE USER terminal_velocity WITH PASSWORD 'your_password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE terminal_velocity TO terminal_velocity;"

# 2. Initialize schema
psql -U terminal_velocity -d terminal_velocity -f scripts/schema.sql

# 3. Generate and save universe
./genmap -systems 100 -save \
  -db-host localhost \
  -db-port 5432 \
  -db-user terminal_velocity \
  -db-password your_password \
  -db-name terminal_velocity

# 4. Create player account
./accounts create <username> <email>

# 5. Start server
./server -config configs/config.yaml
```

**Database Migrations**:
```bash
# Check migration status
./scripts/migrate.sh status

# Apply pending migrations
./scripts/migrate.sh up

# Reset all migrations (DANGEROUS)
./scripts/migrate.sh reset
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

### Production Infrastructure

**Observability & Monitoring**:
```bash
# Metrics server runs on port 8080 by default
curl http://localhost:8080/metrics  # Prometheus-compatible metrics
curl http://localhost:8080/stats    # Human-readable HTML stats page
curl http://localhost:8080/health   # Health check endpoint
```

Metrics tracked:
- Connection metrics (total, active, failed, duration)
- Player metrics (active, logins, registrations, peak)
- Game activity (trades, combat, missions, quests, jumps, cargo)
- Economy (total credits, market volume, trade volume 24h)
- System performance (database queries/errors, cache hit rate, uptime)

**Automated Backups**:
```bash
# Manual backup with defaults
./scripts/backup.sh

# Custom backup configuration
./scripts/backup.sh -d /var/backups -r 30 -c 50

# List available backups
./scripts/restore.sh --list

# Restore from backup
./scripts/restore.sh /path/to/backup.sql.gz

# Automated backups via cron (see scripts/crontab.example)
0 2 * * * /path/to/terminal-velocity/scripts/backup.sh  # Daily at 2 AM
```

Backup features:
- Compression with gzip
- Retention policies (days and count limits)
- Automatic cleanup of old backups
- Safe restore with confirmation prompts
- Progress tracking for large databases

**Rate Limiting & Security**:
- Connection rate limiting (5 concurrent per IP, 20/min per IP)
- Authentication rate limiting (5 attempts before 15min lockout)
- Automatic IP banning (20 failures = 24h ban)
- Brute force protection
- Per-IP tracking with automatic cleanup

Configuration in `internal/ratelimit/ratelimit.go`:
```go
cfg := &ratelimit.Config{
    MaxConnectionsPerIP:     5,
    MaxConnectionsPerMinute: 20,
    MaxAuthAttempts:         5,
    AuthLockoutTime:         15 * time.Minute,
    AutobanThreshold:        20,
    AutobanDuration:         24 * time.Hour,
}
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
  - `metrics/` - Observability & monitoring (Prometheus metrics, HTTP server)
  - `ratelimit/` - Rate limiting & security (connection/auth limiting, auto-banning)
- `scripts/` - Database schema, migrations, and operations
  - `schema.sql` - Main database schema
  - `migrations/` - Database migration scripts
  - `backup.sh` - Automated backup with retention policies
  - `restore.sh` - Database restore from backup
  - `init-server.sh` - Server initialization script
  - `migrate.sh` - Migration runner (up/down/status)
  - `crontab.example` - Example cron jobs for automation
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

### Current Architecture: Monolithic Design

**Note**: The current architecture is monolithic, with SSH server, TUI, and game logic tightly coupled in a single binary. See [Planned Refactoring](#planned-architecture-refactoring) below for future architectural direction.

**Current Structure**:
```
SSH Server (Port 2222)
  ├─ Authentication & Session Management
  ├─ TUI Layer (BubbleTea)
  │   └─ Direct access to managers
  ├─ Game Logic Layer (Managers)
  │   └─ Direct database access
  └─ Database Layer (Repositories)
      └─ PostgreSQL connection pool
```

**Limitations**:
- Cannot scale SSH gateways independently from game logic
- TUI tightly coupled to game logic managers
- Each SSH connection runs full game logic in-process
- Limited to vertical scaling
- Difficult to support alternative clients

### Planned Architecture Refactoring

**Status**: Design phase (Phase 9, post-launch)
**Document**: See [docs/ARCHITECTURE_REFACTORING.md](../docs/ARCHITECTURE_REFACTORING.md) for complete design

**Goal**: Split into client-server architecture to enable:
- Horizontal scaling (scale gateways vs game servers independently)
- Multiple client types (SSH, native terminal, web)
- Better separation of concerns (presentation vs business logic)
- Future fat-client support

**Proposed Architecture**:
```
SSH Gateway Servers          Game Logic Servers
(Frontend - Stateless)  ←───→ (Backend - Stateful)
                        gRPC
├─ SSH Server                  ├─ Game State Manager
├─ TUI Rendering               ├─ Game Logic Engines
└─ API Client                  │   ├─ Combat
                               │   ├─ Trading
                               │   └─ Quests
                               ├─ All Managers
                               └─ Database Layer
                                   └─ PostgreSQL
```

**Migration Strategy**:
1. **Phase 1**: Extract internal API (keep single binary)
2. **Phase 2**: Split into separate services (gateway + gameserver)
3. **Phase 3**: Optimize state synchronization
4. **Phase 4**: Production scalability (K8s, monitoring)

**API Protocol**: gRPC with protobuf
- Bidirectional streaming for real-time updates
- Server-authoritative state management
- Optimistic UI updates with rollback
- Session affinity via consistent hashing

**When Implementing**:
- Start with protobuf schema definitions in `api/proto/`
- Create API client interface for TUI to consume
- Game server implements all business logic
- Gateway only handles SSH, rendering, and API calls
- Maintain backward compatibility during migration

See the full design document for complete details on:
- API surface design
- State synchronization strategy
- Authentication flow
- Data ownership boundaries
- Performance considerations
- Migration checklist

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
- `admin_users` - Server administrators with RBAC (NEW)
- `player_bans` - Banned players with expiration tracking (NEW)
- `player_mutes` - Muted players with expiration tracking (NEW)
- `admin_actions` - Audit log of all admin actions (NEW)
- `server_settings` - Server configuration persistence (NEW)
- `schema_migrations` - Migration tracking (NEW)

See `scripts/schema.sql` for full schema.

### Universe Generation

The `internal/game/universe/` package generates procedural universes:
- Systems placed using spiral galaxy distribution
- Jump routes created via Minimum Spanning Tree (Prim's algorithm) + extra connections
- Tech levels distributed radially (high in core, low at edges)
- 6 NPC factions (governments) assigned to systems
- Planets generated per system with randomized services

Use `cmd/genmap/` to preview and generate universes:
```bash
# Preview universe statistics
./genmap -systems 1000 -stats

# Generate and save to database
./genmap -systems 1000 -save \
  -db-host localhost \
  -db-port 5432 \
  -db-user terminal_velocity \
  -db-password your_password
```

**NEW**: The `-save` flag populates the database directly with generated universe data.

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

### Current State (Phases 0-8 Complete)

✅ **Fully Implemented** (29+ interconnected systems with enhanced UI integration):

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

Future enhancements (Phase 9+):
- Final integration testing across all systems in live environment
- Balance tuning based on community playtesting (economy, combat, progression)
- Performance optimization (database indexing, caching, load testing with 100+ players)
- Community beta testing program and feedback gathering
- Launch preparation (deployment infrastructure, monitoring, community management)
- Post-launch content (more quests, events, ships, storylines)
- Advanced features (player stations, mining, manufacturing, modding support)

### Testing

**Current Test Coverage**:
- ✅ **56 TUI tests passing** (17 integration tests + 39 unit tests)
- ✅ Screen navigation tests (all screens)
- ✅ Combat system tests (weapon firing, AI turns)
- ✅ Space view targeting tests (cycling, wrapping)
- ✅ Async message flow tests
- ✅ Error handling tests
- ✅ State synchronization tests
- Database repository tests (player, ship, market)
- Error handling and retry logic tests
- Metrics testing

**When adding tests**:
- Place `*_test.go` files alongside source
- Use `testify/assert` for assertions (if available) or standard library testing
- Test critical paths: auth, database operations, game systems
- Run with race detector: `go test -race`
- Focus on integration testing across multiple systems
- **Note**: All TUI integration tests are passing as of Phase 8

## Comprehensive Testing Checklist

This section provides a complete testing guide for Phase 9 integration testing. Use this checklist to ensure all features work correctly in a live environment.

### Pre-Testing Setup

**Environment Preparation**:
- [ ] Fresh database with `make setup-db`
- [ ] Universe generated with `./genmap -systems 100 -save`
- [ ] At least 3 test accounts created with different roles (admin, regular, new player)
- [ ] SSH keys added for public key authentication testing
- [ ] Server running with metrics enabled (`make run`)
- [ ] Metrics endpoint accessible (`curl http://localhost:8080/health`)
- [ ] Log files being written to configured location
- [ ] Database backup tested (`./scripts/backup.sh`)

**Test Accounts to Create**:
- [ ] `testadmin` - Admin user with full permissions
- [ ] `testplayer1` - Regular player for gameplay testing
- [ ] `testplayer2` - Second player for multiplayer features
- [ ] `testnewbie` - Fresh account for tutorial/onboarding testing

### 1. Authentication & Account Management

**Password Authentication**:
- [ ] Connect via SSH with valid password
- [ ] Reject invalid password
- [ ] Test rate limiting (5 failed attempts triggers 15min lockout)
- [ ] Test auto-ban (20 failures = 24h ban)
- [ ] Verify lockout message displays correctly
- [ ] Test successful login after lockout expires

**SSH Key Authentication**:
- [ ] Add SSH public key via `./accounts add-key`
- [ ] Connect using SSH key (no password prompt)
- [ ] Test with invalid key (should reject)
- [ ] Test with multiple keys for same account
- [ ] Verify fingerprint matching works correctly

**Account Registration** (if enabled):
- [ ] Access registration screen from main menu
- [ ] Create new account with valid username/email/password
- [ ] Reject invalid usernames (special chars, too short/long)
- [ ] Reject weak passwords
- [ ] Reject duplicate usernames
- [ ] Reject duplicate emails
- [ ] Verify new player starts at default location with starter ship
- [ ] Check tutorial triggers for new players

**Connection Rate Limiting**:
- [ ] Test 5 concurrent connections per IP limit
- [ ] Test 20 connections per minute per IP limit
- [ ] Verify error messages for rate limit violations
- [ ] Check metrics track connection attempts correctly

### 2. User Interface Testing (26 Screens)

**Main Menu Screen**:
- [ ] All menu options visible and correctly labeled
- [ ] Navigation with up/down arrow keys works
- [ ] Selection with Enter key works
- [ ] Quit option (Ctrl+C or Q) exits gracefully
- [ ] Help text displays correctly

**Game/Navigation Screen**:
- [ ] Current system displays with correct info
- [ ] Current planet displays (if docked)
- [ ] Ship status shows (fuel, hull, shields, cargo)
- [ ] Jump to connected systems works
- [ ] Land on planets works
- [ ] Takeoff from planets works
- [ ] Fuel consumption calculated correctly
- [ ] Cannot jump without sufficient fuel
- [ ] Jump route visualization works
- [ ] System info panel updates correctly

**Trading Screen**:
- [ ] Market prices display for all commodities (15 items)
- [ ] Buy commodities (deducts credits, adds cargo)
- [ ] Sell commodities (adds credits, removes cargo)
- [ ] Cargo capacity limits enforced
- [ ] Insufficient credits prevents purchase
- [ ] Prices update based on supply/demand
- [ ] Tech level affects available commodities
- [ ] Transaction history updates

**Cargo Screen**:
- [ ] All cargo items listed with quantities
- [ ] Total cargo weight displayed
- [ ] Cargo capacity shown (current/max)
- [ ] Jettison cargo works
- [ ] Empty cargo bay displays message
- [ ] Cargo sorting works (if implemented)

**Shipyard Screen**:
- [ ] Available ships listed with stats (11 ship types)
- [ ] Ship prices displayed
- [ ] Cannot buy ship without sufficient credits
- [ ] Ship purchase transfers cargo correctly
- [ ] Ship purchase preserves credits correctly
- [ ] Current ship highlighted/indicated
- [ ] Ship stats comparison works
- [ ] Trade-in value calculated correctly

**Outfitter Screen**:
- [ ] All equipment categories visible (6 slot types)
- [ ] Available items listed (16 equipment items)
- [ ] Equipment prices displayed
- [ ] Install equipment (deducts credits, adds to ship)
- [ ] Uninstall equipment (refund credits, removes from ship)
- [ ] Slot limits enforced (e.g., max weapons)
- [ ] Equipment requirements checked (tech level, ship size)
- [ ] Ship stats update when equipment changes
- [ ] Cannot install without sufficient credits

**OutfitterEnhanced Screen**:
- [ ] Equipment browser with filtering works
- [ ] Equipment details panel shows stats
- [ ] Install/uninstall from enhanced UI
- [ ] Visual indicators for installed equipment
- [ ] Slot capacity visualization works
- [ ] Equipment comparison works

**Loadout System**:
- [ ] Save current loadout with custom name
- [ ] Load saved loadout (equipment changes to match)
- [ ] Clone loadout to new name
- [ ] Delete saved loadout
- [ ] List all saved loadouts
- [ ] Loadout validation (checks slot limits, requirements)

**Ship Management Screen**:
- [ ] Ship status overview displays
- [ ] Repair hull option works (deducts credits)
- [ ] Recharge shields option works (if applicable)
- [ ] Refuel option works (deducts credits)
- [ ] Cannot repair/refuel without credits
- [ ] Service costs calculated correctly
- [ ] Ship stats update after services

**Combat Screen**:
- [ ] Combat UI initializes when encountering enemy
- [ ] Player and enemy stats displayed
- [ ] Weapon selection works
- [ ] Fire weapon (damage calculation, accuracy)
- [ ] Enemy AI takes turns (5 difficulty levels)
- [ ] Combat log shows actions
- [ ] Combat ends on victory (loot awarded)
- [ ] Combat ends on defeat (respawn logic)
- [ ] Flee option works (escape chance)
- [ ] Shield damage vs hull damage works correctly
- [ ] Different weapon types work (energy, projectile, missile)

**Missions Screen**:
- [ ] Available missions listed (4 mission types)
- [ ] Accept mission (max 5 active)
- [ ] Cannot accept more than 5 missions
- [ ] Mission details displayed
- [ ] Mission progress tracking works
- [ ] Complete mission (rewards awarded)
- [ ] Abandon mission works
- [ ] Mission objectives update correctly
- [ ] Different mission types work (cargo, combat, explore, bounty)

**Quests Screen**:
- [ ] Available quests listed (7 quest types)
- [ ] Quest details and objectives displayed (12 objective types)
- [ ] Accept quest
- [ ] Quest progress tracking works
- [ ] Branching narrative choices work
- [ ] Complete quest (rewards, story progression)
- [ ] Quest chain progression works
- [ ] Quest markers/hints display correctly

**Achievements Screen**:
- [ ] All achievements listed
- [ ] Locked vs unlocked indicated
- [ ] Progress bars for incremental achievements
- [ ] Achievement descriptions shown
- [ ] Recent achievements highlighted
- [ ] Achievement categories organized
- [ ] Unlock notification triggers correctly

**Events Screen**:
- [ ] Active events listed (10 event types)
- [ ] Event details and objectives shown
- [ ] Event progress tracking works
- [ ] Event leaderboards displayed
- [ ] Participate in event
- [ ] Event rewards distributed correctly
- [ ] Event timer/countdown works
- [ ] Event completion notification

**Encounter Screen**:
- [ ] Random encounters trigger (pirates, traders, police, distress)
- [ ] Encounter dialog displays
- [ ] Encounter choices presented
- [ ] Choice selection works
- [ ] Encounter outcomes apply correctly
- [ ] Loot from encounters awarded
- [ ] Reputation changes from encounters

**News Screen**:
- [ ] News articles listed (10+ event types)
- [ ] News sorted by date (newest first)
- [ ] News details readable
- [ ] News pagination works (if many articles)
- [ ] News updates dynamically from game events
- [ ] Different news types display correctly

**Leaderboards Screen**:
- [ ] All categories accessible (4: credits, combat, trade, exploration)
- [ ] Top players listed with rankings
- [ ] Player's own rank displayed
- [ ] Stats shown for each category
- [ ] Leaderboard updates correctly
- [ ] Ties handled appropriately

**Players Screen**:
- [ ] Online players listed
- [ ] Player locations shown
- [ ] Player status (docked/in space)
- [ ] Player details viewable
- [ ] View player profile works
- [ ] Presence updates in real-time (5min timeout)

**Chat Screen**:
- [ ] All channels accessible (global, system, faction, DM)
- [ ] Send message to channel
- [ ] Receive messages in real-time
- [ ] Channel switching works
- [ ] Direct message to specific player
- [ ] Chat history scrolls correctly
- [ ] Muted players cannot send messages
- [ ] Chat formatting works

**Factions Screen**:
- [ ] Create new faction (deducts creation cost)
- [ ] Join existing faction
- [ ] Leave faction
- [ ] View faction members
- [ ] View faction treasury
- [ ] Deposit to faction treasury (if leader/officer)
- [ ] Withdraw from treasury (if authorized)
- [ ] Promote/demote members (if leader)
- [ ] Kick members (if leader)
- [ ] Faction ranks display correctly

**Territory Screen**:
- [ ] View controlled systems
- [ ] Claim system (if faction has resources)
- [ ] Cannot claim already-claimed system
- [ ] Passive income from territories
- [ ] Territory control timer works
- [ ] Territory visualization works

**Trade Screen** (Player-to-Player):
- [ ] Initiate trade with online player
- [ ] Offer items and credits
- [ ] Accept/reject trade offer
- [ ] Escrow system prevents cheating
- [ ] Trade completion transfers items correctly
- [ ] Trade cancellation returns items
- [ ] Cannot trade with offline players

**PvP Screen**:
- [ ] Challenge player to duel
- [ ] Accept/decline duel challenge
- [ ] Consensual duels work correctly
- [ ] Faction war combat works
- [ ] PvP combat follows same rules as PvE
- [ ] PvP rewards awarded correctly
- [ ] PvP losses penalized appropriately

**Help Screen**:
- [ ] Help topics listed (context-sensitive)
- [ ] Help content displays correctly
- [ ] Help navigation works
- [ ] Search help (if implemented)
- [ ] Help content accurate and helpful

**Settings Screen**:
- [ ] All setting categories accessible (6 categories)
- [ ] Color scheme selection (5 schemes)
- [ ] Color preview works
- [ ] Save settings persists to database
- [ ] Load settings on login
- [ ] Default settings reset works
- [ ] Settings validation works

**Admin Screen**:
- [ ] Only accessible to admin users
- [ ] View server stats
- [ ] Ban player (with expiration)
- [ ] Unban player
- [ ] Mute player (with expiration)
- [ ] Unmute player
- [ ] View audit log (10,000 entries)
- [ ] Server settings modification
- [ ] Permission checks enforce RBAC (4 roles, 20+ permissions)

**Tutorial Screen**:
- [ ] Tutorial triggers for new players
- [ ] All categories accessible (7 categories)
- [ ] Tutorial steps display correctly (20+ steps)
- [ ] Step progression works
- [ ] Skip tutorial option works
- [ ] Tutorial completion tracked
- [ ] Tutorial help context-sensitive

### 3. Core Gameplay Systems

**Ship Systems**:
- [ ] Hull damage tracked correctly
- [ ] Shield recharge works
- [ ] Fuel consumption accurate
- [ ] Cargo capacity enforced
- [ ] Ship upgrades apply correctly
- [ ] Ship destruction/respawn works

**Economy & Trading**:
- [ ] Supply/demand affects prices
- [ ] Tech level affects availability
- [ ] Trade profit calculations correct
- [ ] Market refreshes periodically
- [ ] Illegal commodities tracked (if implemented)
- [ ] Trade volume affects economy

**Reputation System**:
- [ ] Reputation with 6 NPC factions tracked
- [ ] Reputation range: -100 to +100
- [ ] Reputation affects prices (if implemented)
- [ ] Reputation affects encounters
- [ ] Bounty system works
- [ ] Reputation changes from actions

**Loot & Salvage**:
- [ ] Loot drops from combat (4 rarity tiers)
- [ ] Rare items drop correctly
- [ ] Loot awarded to cargo
- [ ] Salvage mechanics work
- [ ] Loot rarity affects value

**Universe & Navigation**:
- [ ] 100+ systems generated correctly
- [ ] Jump routes MST-based
- [ ] Tech levels distributed radially
- [ ] Government assignments work
- [ ] System information accurate
- [ ] Jump visualization works

### 4. Content Systems

**Quest System**:
- [ ] All 7 quest types available
- [ ] All 12 objective types work
- [ ] Branching narratives function
- [ ] Quest rewards awarded
- [ ] Quest completion tracked
- [ ] Quest chains progress correctly

**Mission System**:
- [ ] All 4 mission types available (cargo, combat, explore, bounty)
- [ ] Mission generation works
- [ ] Mission progress tracked
- [ ] Mission completion rewards
- [ ] Mission time limits enforced
- [ ] Mission failures handled

**Dynamic Events**:
- [ ] All 10 event types trigger
- [ ] Event scheduling works
- [ ] Event participation works
- [ ] Event leaderboards update
- [ ] Event rewards distributed
- [ ] Events end correctly

**Random Encounters**:
- [ ] Pirates encounter works
- [ ] Traders encounter works
- [ ] Police encounter works
- [ ] Distress call works
- [ ] Encounter frequency appropriate
- [ ] Encounter outcomes balanced

**Achievements**:
- [ ] Achievement tracking works
- [ ] Achievement unlock triggers
- [ ] Achievement notifications display
- [ ] Achievement progress saved
- [ ] All achievement types work

**News System**:
- [ ] News generated from events (10+ types)
- [ ] News articles created dynamically
- [ ] News displayed chronologically
- [ ] News updates regularly
- [ ] News content accurate

### 5. Multiplayer Features

**Player Presence**:
- [ ] Online players tracked
- [ ] Player locations updated
- [ ] Offline timeout works (5 minutes)
- [ ] Presence broadcasts to others
- [ ] Presence updates efficient

**Chat System**:
- [ ] Global chat broadcasts to all
- [ ] System chat limited to current system
- [ ] Faction chat limited to faction members
- [ ] Direct messages work
- [ ] Chat history persists
- [ ] Muted players blocked

**Faction System**:
- [ ] Faction creation works
- [ ] Faction joining works
- [ ] Faction treasury works
- [ ] Member management works
- [ ] Rank system works
- [ ] Faction permissions enforced

**Territory Control**:
- [ ] System claiming works
- [ ] Passive income generated
- [ ] Territory conflicts work
- [ ] Territory visualization accurate
- [ ] Territory control timer works

**Player Trading**:
- [ ] Trade initiation works
- [ ] Trade offers display correctly
- [ ] Escrow system prevents exploits
- [ ] Trade completion atomic
- [ ] Trade cancellation safe

**PvP Combat**:
- [ ] Duel challenges work
- [ ] Consensual combat only
- [ ] Faction wars work
- [ ] PvP rewards distributed
- [ ] PvP balance fair

**Leaderboards**:
- [ ] Credits leaderboard accurate
- [ ] Combat leaderboard accurate
- [ ] Trade leaderboard accurate
- [ ] Exploration leaderboard accurate
- [ ] Rankings update correctly

### 6. Infrastructure & Administration

**Server Administration**:
- [ ] RBAC enforced (4 roles: owner, admin, moderator, helper)
- [ ] All 20+ permissions checked
- [ ] Ban system works (with expiration)
- [ ] Mute system works (with expiration)
- [ ] Audit log records actions (10,000 buffer)
- [ ] Admin commands work
- [ ] Permission violations blocked

**Session Management**:
- [ ] Auto-save works (30 second interval)
- [ ] Session persistence across reconnects
- [ ] Graceful disconnection handling
- [ ] Session cleanup on logout
- [ ] Concurrent session handling

**Metrics & Monitoring**:
- [ ] Prometheus metrics endpoint works (`/metrics`)
- [ ] HTML stats page works (`/stats`)
- [ ] Health check endpoint works (`/health`)
- [ ] Connection metrics accurate
- [ ] Player metrics accurate
- [ ] Game activity metrics accurate
- [ ] Economy metrics accurate
- [ ] System performance metrics accurate

**Error Handling**:
- [ ] Database errors handled gracefully
- [ ] Network errors handled gracefully
- [ ] Invalid input rejected safely
- [ ] Error messages user-friendly
- [ ] Errors logged appropriately
- [ ] Retry logic works (exponential backoff)

**Logging**:
- [ ] Log levels work (debug, info, warn, error)
- [ ] Log files created correctly
- [ ] Log rotation works (if implemented)
- [ ] Sensitive data not logged (passwords)
- [ ] Logs parseable and useful

### 7. Database Operations

**Connection Pooling**:
- [ ] Connection pool initializes correctly
- [ ] Pool size configurable
- [ ] Connections reused efficiently
- [ ] Pool cleanup on shutdown
- [ ] Connection errors handled

**Repositories**:
- [ ] PlayerRepository CRUD works
- [ ] SystemRepository CRUD works
- [ ] SSHKeyRepository CRUD works
- [ ] ShipRepository CRUD works
- [ ] MarketRepository CRUD works
- [ ] All other repositories work
- [ ] Transaction handling works

**Migrations**:
- [ ] Migration tracking works (`schema_migrations` table)
- [ ] Migration up/down works
- [ ] Migration status accurate
- [ ] Migration failures rollback
- [ ] Migration scripts correct

**Backups**:
- [ ] Manual backup works (`./scripts/backup.sh`)
- [ ] Backup compression works (gzip)
- [ ] Backup retention enforced
- [ ] Old backup cleanup works
- [ ] Restore works (`./scripts/restore.sh`)
- [ ] Restore prompts for confirmation
- [ ] Large database backups work

### 8. Performance Testing

**Load Testing**:
- [ ] Test with 10 concurrent players
- [ ] Test with 50 concurrent players
- [ ] Test with 100+ concurrent players
- [ ] Response times acceptable under load
- [ ] No race conditions detected (`go test -race`)
- [ ] Memory usage stable over time
- [ ] CPU usage acceptable under load

**Database Performance**:
- [ ] Query performance acceptable
- [ ] Indexes used correctly
- [ ] Connection pool performs well
- [ ] No connection pool exhaustion
- [ ] Database locks minimal
- [ ] Transaction deadlocks handled

**Background Workers**:
- [ ] Event scheduler performs efficiently
- [ ] Metrics collection efficient
- [ ] Cleanup tasks run on schedule
- [ ] Auto-save doesn't block gameplay
- [ ] Worker goroutines don't leak

### 9. Security Testing

**Input Validation**:
- [ ] SQL injection prevented (parameterized queries)
- [ ] Command injection prevented
- [ ] XSS not applicable (terminal UI)
- [ ] Buffer overflow prevented
- [ ] Path traversal prevented
- [ ] Invalid UUIDs rejected

**Authentication Security**:
- [ ] Passwords hashed correctly (bcrypt/scrypt)
- [ ] SSH keys validated correctly
- [ ] Session tokens secure
- [ ] Rate limiting effective
- [ ] Auto-ban prevents brute force
- [ ] Timing attacks mitigated

**Authorization**:
- [ ] RBAC enforced everywhere
- [ ] Permission checks cannot be bypassed
- [ ] Horizontal privilege escalation prevented
- [ ] Vertical privilege escalation prevented
- [ ] Resource access controlled

**Data Protection**:
- [ ] Passwords never logged
- [ ] Sensitive data encrypted at rest (if applicable)
- [ ] Sensitive data encrypted in transit (SSH)
- [ ] Database credentials secured
- [ ] API tokens secured (if applicable)

### 10. Edge Cases & Error Conditions

**Boundary Testing**:
- [ ] Cargo at max capacity
- [ ] Credits at 0
- [ ] Fuel at 0 (cannot jump)
- [ ] Hull at 0 (ship destroyed)
- [ ] Reputation at -100 and +100
- [ ] Maximum missions (5) active
- [ ] Empty markets
- [ ] Single-player universe

**Concurrent Operations**:
- [ ] Multiple players trading same item
- [ ] Simultaneous faction treasury access
- [ ] Concurrent territory claims
- [ ] Race conditions in PvP
- [ ] Concurrent database writes

**Network Conditions**:
- [ ] Handle disconnections gracefully
- [ ] Reconnect preserves state
- [ ] Timeout handling works
- [ ] Partial message handling
- [ ] SSH connection drops handled

**Database Failures**:
- [ ] Connection loss handled
- [ ] Query timeout handled
- [ ] Transaction failures rollback
- [ ] Database unavailable handled
- [ ] Retry logic works

**Resource Exhaustion**:
- [ ] Memory limits respected
- [ ] File descriptor limits respected
- [ ] Database connection limits enforced
- [ ] Goroutine limits (no leaks)
- [ ] Disk space handling

### Testing Workflow

**Daily Testing Routine**:
1. Start fresh server instance
2. Test 3-5 major features from checklist
3. Document any bugs found
4. Test bug fixes immediately
5. Run automated tests: `make test`
6. Check metrics dashboard
7. Review logs for errors

**Weekly Testing Routine**:
1. Full database backup/restore test
2. Load testing with multiple clients
3. Review all open bugs
4. Test recent code changes thoroughly
5. Cross-feature integration tests
6. Performance profiling
7. Security audit of new features

**Pre-Release Testing**:
1. Complete entire checklist
2. Multi-hour stress test
3. Security penetration testing
4. Fresh player experience (tutorial → endgame)
5. All multiplayer features with real players
6. Backup/restore full cycle
7. Migration testing (fresh install vs upgrade)
8. Documentation accuracy review

### Bug Reporting Template

When you find a bug, document it with:

```
**Bug Title**: Brief description

**Severity**: Critical / High / Medium / Low

**Category**: Authentication / UI / Database / Gameplay / etc.

**Steps to Reproduce**:
1. Step one
2. Step two
3. Step three

**Expected Behavior**:
What should happen

**Actual Behavior**:
What actually happens

**Error Messages**:
Any error messages or logs

**Environment**:
- Go version
- PostgreSQL version
- OS
- Server commit hash

**Possible Fix** (optional):
Ideas for fixing the issue
```

### Performance Benchmarking

**Baseline Metrics to Record**:
- Average player login time: _______
- Average market price calculation: _______
- Average combat turn processing: _______
- Average database query time: _______
- Memory usage (idle): _______
- Memory usage (10 players): _______
- Memory usage (50 players): _______
- CPU usage (idle): _______
- CPU usage (10 players): _______
- CPU usage (50 players): _______

**Performance Goals**:
- Player login: < 500ms
- Market operations: < 100ms
- Combat turns: < 200ms
- Database queries: < 50ms
- UI responsiveness: < 16ms (60 FPS feel)
- Memory per player: < 10MB
- Support 100+ concurrent players

### Final Integration Testing Report

After completing the checklist, create a report:

**Features Tested**: ___ / Total
**Bugs Found**: ___
**Critical Bugs**: ___
**High Priority Bugs**: ___
**Medium Priority Bugs**: ___
**Low Priority Bugs**: ___
**Performance Issues**: ___
**Security Issues**: ___

**Overall Assessment**: Ready for Beta / Needs Work / Not Ready

**Notes**:
- Major issues found
- Features needing refinement
- Performance bottlenecks identified
- Security concerns
- Recommendations for beta testing

---

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

## Current Status (Phase 8 Complete)

Phases 0-8 are complete! All 56 TUI tests passing. Next priorities (Phase 9):
1. **Final Integration Testing**: Test all 29+ systems in live environment
2. **Community Beta Testing**: Recruit beta testers and gather comprehensive feedback
3. **Balance Tuning**: Fine-tune economy, combat, and progression based on playtesting
4. **Performance Optimization**: Database optimization, caching, load testing with 100+ players
5. **Launch Preparation**: Deployment infrastructure, monitoring, community management tools

## Future Direction (Phase 9+)

**Phase 9: Architecture Refactoring** (Post-Launch)
- Client-server split using gRPC
- Horizontal scalability
- Support for multiple client types
- See [docs/ARCHITECTURE_REFACTORING.md](../docs/ARCHITECTURE_REFACTORING.md) for complete design

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

**Last Updated**: 2025-11-14
**Document Version**: 2.4.0
**Project Version**: 0.8.0 (Phase 8 Complete, Production-Ready Infrastructure)