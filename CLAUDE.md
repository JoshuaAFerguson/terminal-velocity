# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terminal Velocity is a multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH. Players navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions — all within a terminal UI built with BubbleTea.

**Tech stack**: Go 1.24+, PostgreSQL (pgx/v5), BubbleTea + Lipgloss (TUI), `golang.org/x/crypto/ssh`, gRPC/protobuf (in progress).

**Codebase size** (re-measure these rather than quoting them; they drift fast):
- ~78k lines of Go (`find . -name '*.go' -not -path './api/gen/*' | xargs wc -l`)
- 48 packages under `internal/`
- 53 files under `internal/tui/` covering ~41 screens (`Screen` enum in `internal/tui/model.go`)
- 9 repositories under `internal/database/` (named `*_repository.go`)
- 4 binaries under `cmd/`: `server`, `genmap`, `accounts`, `loadtest`

For feature-level status and phase history see `ROADMAP.md` and `FEATURES.md`. For known security/bug findings see `BUG_SECURITY_ANALYSIS.md`.

## Building and Running

### Common development commands

```bash
# Build
make build              # Build server binary (runs `make proto` first)
make build-tools        # Build genmap and accounts utilities
make release            # Cross-compile linux/darwin × amd64/arm64 to build/
make install-deps       # go mod download + go mod tidy

# Run
make run                # Run server in development mode
make watch              # Auto-rebuild on *.go change (requires entr)

# Database
make setup-db           # Apply scripts/schema.sql (requires psql)
./scripts/migrate.sh status|up|reset   # Migration runner

# Quality
make test               # go test -race -coverprofile=coverage.out ./...
make coverage           # Open HTML coverage report
make lint               # golangci-lint (config in .golangci.yml)
make fmt                # gofmt -s -w
make vet                # go vet

# Protobuf (gRPC API work, in progress)
make install-proto-tools  # protoc-gen-go, protoc-gen-go-grpc
make proto                # Regenerate api/gen/go/v1/**
make proto-clean          # Remove generated protobuf code

# Docker (note: Makefile target names contain a space, e.g. `make docker compose-up`)
make docker-build
make docker compose-up / docker compose-down / docker compose-logs / docker compose-restart
make docker-clean         # Also removes volumes

# Tools
./genmap -systems 100 -stats            # Preview a universe
./genmap -systems 100 -save             # Generate and persist to DB
./accounts create <username> <email>
./accounts add-key <username> <key-file>
```

Run a single test: `go test -run TestName ./internal/<pkg>/...` (or add `-race -v`).

### First-time server setup

```bash
make build-tools
./scripts/init-server.sh          # Creates DB, applies schema, generates 100-system universe
./accounts create <username> <email>
make run
ssh -p 2222 <username>@localhost
```

For manual setup see `README.md` and `QUICKSTART.md`.

### Observability

Metrics HTTP server defaults to `:8080`:
- `/metrics` — Prometheus-compatible
- `/stats` — human-readable HTML
- `/health` — liveness probe

Rate-limit/auto-ban defaults (configurable in `internal/ratelimit/ratelimit.go`): 5 connections/IP, 20 conn/min/IP, 5 failed auth → 15 min lockout, 20 failures → 24 h ban.

Backups: `./scripts/backup.sh` (gzip + retention), `./scripts/restore.sh`. Example cron in `scripts/crontab.example`.

## Architecture

### Layout

- `cmd/` — entry points: `server/`, `genmap/`, `accounts/`, `loadtest/`
- `api/` — gRPC/protobuf surface (in-progress refactor, see below)
  - `proto/` — `auth.proto`, `common.proto`, `game.proto`, `player.proto`
  - `gen/go/v1/` — generated Go code (produced by `make proto`, not checked in)
- `internal/` — 48 packages, grouped here for orientation:
  - **Infra**: `server` (SSH), `api` (gRPC client + server), `database`, `models`, `session`, `metrics`, `ratelimit`, `security`, `errors`, `logger`, `validation`, `notifications`, `qol`
  - **Presentation**: `tui` (41-screen BubbleTea model + sub-components)
  - **Core gameplay**: `game` (universe, trading), `combat`, `outfitting`, `loadouts`, `shipsystems`, `missions`, `quests`, `events`, `encounters`, `achievements`, `news`, `leaderboards`, `traderoutes`, `mining`, `manufacturing`, `marketplace`, `capture`, `arena`, `launch`
  - **Social / multiplayer**: `chat`, `mail`, `friends`, `presence`, `pvp`, `trade`, `factions`, `factioncontent`, `fleet`, `territory`, `diplomacy`
  - **Meta**: `admin`, `help`, `tutorial`, `settings`
- `scripts/` — `schema.sql`, `migrations/`, `init-server.sh`, `migrate.sh`, `backup.sh`, `restore.sh`, `crontab.example`
- `configs/config.example.yaml` — reference server config (copy to `config.yaml`)
- `docs/` — supplemental docs, including `TESTING_CHECKLIST.md` and `ARCHITECTURE_REFACTORING.md`

### Key patterns

**Repository pattern** — all SQL lives in `internal/database/*_repository.go`. Repositories expose typed CRUD and accept `context.Context` as the first arg. pgx-native errors (not `database/sql`). Nullable IDs use `*uuid.UUID`.

**BubbleTea MVC** — each screen has a model, an `update*` method, and a `view*` method on the top-level `Model`. The `Screen` enum in `internal/tui/model.go` drives routing in `Update()` and `View()`. Never block in `Update()`; return `tea.Cmd` for async work. Define custom message types for async results and handle them in `Update()`.

**SSH integration** — `internal/server/server.go` handles password + public-key auth. On successful auth the server stores `player_id` in `ssh.Permissions.Extensions`, then `startGameSession` loads the player and runs a BubbleTea program with the SSH channel as I/O.

**Thread safety** — manager packages use `sync.RWMutex` throughout. Use `RLock/RUnlock` for reads and `Lock/Unlock` for writes. Typical manager shape: `New()`, `Start()` (kicks off background goroutines), `Stop()` (cleanup).

### Current state: monolithic, with an in-progress client-server split

Today, one binary runs SSH server + TUI + game logic + DB access. `api/proto/` and `internal/api/` (with `client.go`, `server/server.go`, `server/session.go`, `server/converters.go`, `types.go`) exist as the first landing of a gRPC-based client-server refactor that will eventually let SSH gateways scale independently from game-logic servers.

What's landed: proto schema, generated-code pipeline (`make proto`), initial client/server skeletons. What's still ahead: wiring the TUI to the API client, moving managers behind the server, session-affinity gateway. Design doc: `docs/ARCHITECTURE_REFACTORING.md`.

When touching this area, prefer the gRPC boundary over adding direct cross-layer calls.

### Authentication flow

The live flow is **anonymous at the SSH layer, authenticated in the TUI**. Any SSH identity is accepted; `accounts add-key` stores keys in the DB but they aren't consulted by the SSH handshake (current state — `startGameSession` and `startRegistrationSession` in `server.go` are defined but not called from the login path).

1. SSH `newChannel.Accept()` → `handleSession` (server.go).
2. pty-req payload is parsed for window size (`parsePTYReq` in `pty.go`) and stored.
3. On `shell` request, `startAnonymousSession(channel, requests, initialSize)` runs the login BubbleTea program with `tea.NewProgram(... WithAltScreen())`.
4. A goroutine forwards the initial `tea.WindowSizeMsg` and subsequent `window-change` requests into the program so the TUI resizes live.
5. The user types username + password in the form; `playerRepo.Authenticate` checks credentials.
6. On success, the same Model stays in-process and transitions its `Screen` enum forward — we do not spawn a new BubbleTea program per screen.

Registration is gated by `AllowRegistration` in server config. `pquerna/otp` is in `go.mod` as scaffolding for future TOTP/2FA but is not wired to the login screen today.

### TUI rendering rules

- **Cell width, not byte length.** The helpers in `ui_components.go` — `PadRight`, `PadLeft`, `Center`, `TruncateString` — measure and slice by terminal cells via `lipgloss.Width` and rune iteration. Never use `len()` on a string that contains box-drawing chars, styled text, or anything beyond ASCII — it returns byte length, and layouts built on it will wrap or split mid-codepoint.
- **Printable-rune text input.** Screens that accept typed text in a field use `printableRuneString(msg tea.KeyMsg) (string, bool)` from `input_keys.go`, not `msg.String()`. That accepts pastes (multi-rune KeyMsgs), strips embedded control bytes, and rejects named keys ("up", "f1", "ctrl+x", etc.). Don't reintroduce `if len(msg.String()) == 1 { ... }` — it silently drops pastes.
- **Terminal size from SSH.** `NewLoginModel` / `NewModel` default to 80×24 as a seed; the real size arrives via `tea.WindowSizeMsg` from the goroutine in `startAnonymousSession`. Don't hardcode widths in `view*` functions — read `m.width` / `m.height`.

### Database

PostgreSQL with UUID PKs. See `scripts/schema.sql` for the full schema. Notable tables: `players`, `player_ssh_keys`, `player_reputation`, `star_systems`, `system_connections`, `planets`, `ships`, plus admin/audit tables (`admin_users`, `player_bans`, `player_mutes`, `admin_actions`, `server_settings`) and `schema_migrations`.

Use parameterized queries (`$1`, `$2`) — never string-concat. Nullable columns: `sql.NullString`, etc. Connection pooling is managed in `internal/database/connection.go`.

### Universe generation

`internal/game/universe/` places systems in a spiral galaxy distribution, builds jump routes via Prim's MST plus a few extra connections for loops, and assigns tech levels radially (high in the core) with 6 NPC governments. `cmd/genmap/` previews or persists universes.

## Conventions

### File headers

Every Go source file carries a header comment; bump the version on edit:

```go
// File: internal/package/filename.go
// Project: Terminal Velocity
// Description: Brief description of file purpose
// Version: X.Y.Z
// Author: Joshua Ferguson
// Created: YYYY-MM-DD
```

- Patch (`Z+1`): bug fixes, minor changes
- Minor (`Y+1.0`): new features, significant changes
- Major (`X+1.0.0`): breaking changes, major refactors

This overrides the default "don't write comments" posture — it's a project-wide convention.

### Errors

- Package-level sentinel errors, e.g. `var ErrPlayerNotFound = errors.New("player not found")`.
- Wrap with `%w` when returning up the stack.
- Use `internal/errors` for retry/backoff helpers and metrics.
- Log via `internal/logger`.

### BubbleTea messages

```go
type dataLoadedMsg struct {
    data *Data
    err  error
}
```

Return these from `tea.Cmd` functions (use `context.Background()` inside commands) and route them in `Update()`.

### Linting

`make lint` must pass. Highlights from `.golangci.yml`:

- Enabled: errcheck, gosimple, govet, ineffassign, staticcheck, typecheck, gofmt, goimports, misspell, dupl.
- TUI files exempt from errcheck (UI-flow noise). Game code is allowed to use `math/rand` (gosec G404 excluded). Test files exempt from dupl/gosec/gocyclo.
- Cyclomatic complexity budget: 40. Dup threshold: 150 lines.

## Common tasks

### Add a new TUI screen

1. Add `internal/tui/<name>.go` with a model struct, `update<Name>(msg tea.Msg)` and `view<Name>()` methods on `Model`.
2. Add a constant to the `Screen` enum in `internal/tui/model.go`.
3. Add cases to `Update()` and `View()` switches.
4. Initialize state in `newModel` / main-menu entry if user-reachable.

### Add a new database table

1. Append to `scripts/schema.sql` and add a migration under `scripts/migrations/NNN_description.sql`.
2. Add a model in `internal/models/`.
3. Add a `*_repository.go` with CRUD methods (context-first, parameterized SQL).
4. Wire the repo through the server and TUI model constructors.

### Update player state

Flow: DB → repository → TUI model → screen. Mutate via repo method (e.g. `playerRepo.UpdateLocation`), then update local `m.player`, then return a `tea.Cmd` if you need to refresh derived data.

## Gotchas

- Market prices are `float64` — be careful with precision; cargo capacity is measured in tons, not units.
- Repository methods must handle nil/empty results gracefully.
- SSH auth needs both password and public-key paths wired; don't break one while fixing the other.
- Don't forget to initialize sub-models when adding to the root `Model` — uninitialized screens manifest as blank panels.
- BubbleTea programs need clean shutdown on SSH channel close.

## Key files

- `internal/tui/model.go` — root TUI model, `Screen` enum, routing
- `internal/server/server.go` — SSH server + auth
- `internal/database/connection.go` — pgx connection pool
- `api/proto/` — gRPC schema (in-progress refactor)
- `scripts/schema.sql` — canonical DB schema
- `Makefile` — all build/dev commands
- `.golangci.yml` — lint config
- `docs/ARCHITECTURE_REFACTORING.md` — client-server split design
- `docs/TESTING_CHECKLIST.md` — integration/regression checklist (moved out of CLAUDE.md)
- `BUG_SECURITY_ANALYSIS.md` — known issues and triage

## Key dependencies

From `go.mod` (verify versions before quoting):

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — terminal styling
- `github.com/jackc/pgx/v5` — PostgreSQL driver with pooling
- `github.com/google/uuid` — UUID
- `golang.org/x/crypto` — SSH server + password hashing
- `golang.org/x/term` — terminal utilities
- `github.com/pquerna/otp` — TOTP (2FA scaffolding)
- `gopkg.in/yaml.v3` — config parsing

Tooling: Go 1.24+, PostgreSQL 12+, `golangci-lint`, optional `entr` (for `make watch`), `protoc` + Go plugins (for `make proto`).

## Before committing

- `make fmt && make lint && make test` must all pass.
- Bump the `Version:` header of every edited Go file.
- Update `CHANGELOG.md`; check whether `README.md` or `ROADMAP.md` need updates too.

Repository: https://github.com/JoshuaAFerguson/terminal-velocity/
