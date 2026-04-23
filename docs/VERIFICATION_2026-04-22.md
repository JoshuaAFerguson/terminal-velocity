# Manual Verification — Rendering & Input Fixes

This session landed four classes of fix. Each has automated tests where possible,
but the interactive bits need a human in the loop. Please run through the list
below on a real terminal and note anything still broken.

## Setup

```bash
docker compose up -d postgres
DB_HOST=localhost DB_PORT=5432 DB_USER=terminal_velocity \
  DB_PASSWORD=tv_dev_password DB_NAME=terminal_velocity \
  ./server -config configs/config.yaml
```

In another terminal:

```bash
ssh -p 2222 tester@localhost     # password: TestPass123!
```

## What changed, and how to exercise it

### 1. PTY size forwarding (root cause of "everything looks squashed")

The server now parses `pty-req` and forwards `window-change` events to BubbleTea
so the TUI adapts to your actual terminal size.

- [ ] Connect at an 80×24 terminal. Layout should be the compact version — ASCII
      banner fits, login form centered in a 50-col inner panel, clean borders.
- [ ] Resize your terminal to 120×40 **while connected**. The UI should redraw
      with the outer box spanning the new width, not stay stuck at 80.
- [ ] Resize to 180×50. Same — box should span the full width.
- [ ] Shrink below 80×24. Screens may truncate (by design — see
      "minimum size" note in CLAUDE.md) but should not corrupt borders or
      leave replacement characters visible.

**Red flag**: if the outer box is fixed at 80 cols regardless of terminal
width, the fix didn't land — re-check `internal/server/server.go:handleSession`
and the `forwardWindowSize` goroutine in `startAnonymousSession`.

### 2. Rune-safe width math (fixes "boxes with `�` in them")

Tab, type, paste, and scroll through fields that mix box-drawing characters
with user input. Look for replacement characters (`�`), wrapped border rows,
or inner panels whose right edge doesn't line up with surrounding content.

- [ ] Landing screen OR separator shows a full `─────── OR ───────────` row.
- [ ] Login form Username / Password rows have matching right-edge borders.
- [ ] Footer `[Tab/↑/↓] Navigate  [Enter] Select  [Ctrl+C] Quit` padding meets
      the right border cleanly.
- [ ] Sign in with `tester` / `TestPass123!` — verify the post-login screen
      (main menu) doesn't show ghost row duplication or misaligned borders.

If a specific screen looks wrong, grab the raw stream with:

```bash
script -q -c "ssh -p 2222 tester@localhost" /tmp/tv.log
# interact, then q to quit; open /tmp/tv.log in a hex viewer
```

### 3. Text input — pastes and multi-byte characters

Input handlers used to accept only `len(msg.String()) == 1` keystrokes, which
silently dropped pastes. Now they use `printableRuneString`, which accepts
multi-rune payloads and strips embedded control bytes.

Screens to exercise (each has its own input field):
- [ ] Login form — paste a password with Cmd-V / Ctrl-Shift-V (on most
      terminals this appears to BubbleTea as a single KeyMsg with all runes).
- [ ] Chat input (press `c` from main menu, then type/paste).
- [ ] Mail compose (recipient, subject, body).
- [ ] Fleet rename (`f` menu → create fleet).
- [ ] Ship management rename.
- [ ] Marketplace create-listing forms (numeric and text fields).
- [ ] Factions create form (name + tag).
- [ ] Friends add-by-username.
- [ ] PvP challenge and trade-create forms.
- [ ] Item picker search field.

For numeric-only fields (marketplace starting_bid, amount, duration, reward),
pasting mixed alpha-numeric should keep only the digits.

### 4. Tooling fixes

- [ ] `make build` works without `protoc` installed. Previously failed on the
      proto step even though the generated code is unused.
- [ ] `make proto` still works when `protoc` + plugins are installed
      (`make install-proto-tools`).
- [ ] `./accounts --help` prints usage without requiring DB credentials.
      (`./accounts`, `./accounts help`, `./accounts -h` likewise.)
- [ ] `./genmap -systems 100 -save ...` completes successfully with the
      default seed (previously hit a planet-name collision after ~50 systems).
- [ ] `configs/config.example.yaml` uses `host:` / `port:` / `user:` / etc.
      rather than an ignored `url:` field.

## Login flow — confirmed end-to-end via tmux

The "Enter submission" question from the first draft turned out to be **three
real product bugs**, now all fixed. Reproduced with `scripts/tmux_login_test.sh`
which drives a real tmux PTY (not Python `pty.fork()` — that harness can't
simulate Enter reliably).

1. **Missing DB columns `crafting_skill`, `total_crafts`, `resources_mined`**.
   Archived migrations 016 and 017 (`scripts/migrations-archive/`) added these
   but the canonical `scripts/schema.sql` was never updated. `PlayerRepository.GetByID`
   queries them, so every post-login player lookup failed with
   `ERROR: column "crafting_skill" does not exist`. Schema now includes them.

2. **Main `Update` at `model.go` intercepted `playerLoadedMsg` before
   `updateLogin` could see it.** The login screen has its own handler that
   transitions to `ScreenMainMenu` on success; the top-level one just stored
   `m.err` and returned. The result: login succeeded at the DB level but the
   TUI never left the login screen. Fixed by letting login/registration
   screens handle the message themselves.

3. **`research_points` was present in the updated schema.sql but missing from
   the Postgres volume** on my dev machine (the container was initialized
   from an earlier schema). Running `scripts/init-server.sh` against a fresh
   volume now applies everything. If you have an old dev volume, you'll need:

   ```sql
   ALTER TABLE players
     ADD COLUMN IF NOT EXISTS crafting_skill INTEGER DEFAULT 0
       CHECK (crafting_skill >= 0 AND crafting_skill <= 100),
     ADD COLUMN IF NOT EXISTS total_crafts INTEGER DEFAULT 0
       CHECK (total_crafts >= 0),
     ADD COLUMN IF NOT EXISTS resources_mined JSONB DEFAULT '{}'::jsonb,
     ADD COLUMN IF NOT EXISTS research_points INTEGER DEFAULT 100
       CHECK (research_points >= 0);
   ```

Confirmation: `./scripts/tmux_login_test.sh smoke` → `PASS — reached main menu`.

## Post-login screens — walked and fixed via `scripts/tmux_menu_walk.sh`

After the login flow worked, I extended the tmux harness (`scripts/tmux_menu_walk.sh`)
to drive each main menu entry, capture each screen, then press Esc to return
to the menu. Each iteration found new real bugs; each fix unblocked the next
screen. Summary of what that uncovered:

### Starter state wasn't being created

`accounts create` (and the registration flow) persisted a bare `players` row
with no `current_system` and no `ship_id`. Every screen that depends on the
player having a ship or a system displayed stubs or crashed:

- Main menu showed `Location: Unknown`, then `Space` unconditionally once
  `CurrentSystem` was non-nil — it never looked up the actual system name.
- Ship Management panicked on `m.currentShip.Name` because `currentShip`
  was nil.

Fixed by adding `PlayerRepository.bootstrapStarterState` — picks a random
system from the generated universe and inserts a Shuttle ship, wiring both
back into the player row. `CreateWithEmail` calls it on every new account.

### `GetByID` and `GetByUsername` didn't return `ship_id`

The login flow calls `playerRepo.GetByID(playerID)` to populate the TUI. The
query had been updated for several new columns over time but was missing
`ship_id`. So `m.player.ShipID` was always `uuid.Nil` and the follow-up
ship lookup in `loadPlayer` never ran. Fixed — both queries now select
`ship_id` and parse it into the model. `loadPlayer` also now loads the
player's current star system into `m.currentSystem` for header rendering.

### Main-menu Escape trap in `outfitter_enhanced`

Entering **Advanced Outfitting** from the main menu and pressing Esc
transitioned to `ScreenSpaceView` — a WIP stub screen — instead of back to
the main menu. Fixed by adding `previousScreen`/`hasPreviousScreen` fields to
the Model: the main menu records itself before navigating, and sub-screens
honor that on Esc. Space-view-based flows still get the old behavior.

### Launch screen advertised keys that weren't wired

`internal/tui/game.go:viewGame` displayed:

```
  n - Navigation
  t - Trading
  s - Shipyard
  m - Missions
  r - Trade Routes
  M - Mail
```

But only `r`, `M`, and `esc` actually had handlers. Typing `n/t/s/m` did
nothing. Wired them all up; they now create the right sub-model (matching
what main-menu Select does) and kick off the correct async loader.

### Missing DB columns

Several archived migrations (`014_add_weapon_ammo_tracking`, `015_add_planet_coordinates`,
`016_add_resources_mined_tracking`, `017_add_crafting_skill_tracking`) had
never been merged into `scripts/schema.sql` *or* applied to running dev
databases. Code queries referencing those columns failed at runtime:

- `ship_weapons.current_ammo` — blocked ship lookup inside combat/landing
- `players.crafting_skill`, `total_crafts`, `resources_mined`, `research_points`
  — blocked post-login `GetByID`

Fixed by adding them to `scripts/schema.sql`. Existing dev volumes also need
the ALTERs documented in the "Login confirmed" section above.

### Polish

- SSH connection close log now degrades to DEBUG when the remote side
  already closed (`net.ErrClosed`). The `"use of closed network connection"`
  warning on every disconnect is gone.
- `scripts/migrate.sh` points at `scripts/migrations-archive/` (the real
  location) instead of the non-existent `scripts/migrations/`.
- `internal/database/item_repository_test.go` had a test-only build bug
  (`db.QueryRow(ctx, ...)` passed context to the non-context method). Fixed.

### Test coverage

- `scripts/tmux_login_test.sh` — end-to-end login smoke test. Exit 0 iff
  the main menu renders. CI-friendly, runs in ~15s.
- `scripts/tmux_menu_walk.sh` — walks all 22 main menu entries and saves a
  capture of each. Used as a regression sentinel; fails loudly on any
  unhandled panic since a panic drops the session to shell.
- All Go unit tests pass: `go test -race ./internal/server/ ./internal/tui/
  ./internal/game/universe/ ./internal/database/`.

- **`_Enhanced` screen variants.** The `Screen` enum has multiple copies:
  `ScreenTrading` and `ScreenTradingEnhanced`, `ScreenShipyard` and
  `ScreenShipyardEnhanced`, etc. I don't know which are the live ones. Worth
  deleting the dead ones in a follow-up.

## Residual defects worth triaging separately

- `Failed to close SSH connection: use of closed network connection` warning
  on every disconnect. Close-order race; cosmetic.
- `scripts/migrations/` referenced by `scripts/migrate.sh` but the directory
  is actually `scripts/migrations-archive/`. Either rename or update the
  script.
- `internal/arena/manager.go` had two field-reference bugs
  (`tournament.MatchType`, `MatchData.Damage`) I fixed in passing. Suggests
  the arena package hasn't been exercised in a while.
- `BUG_SECURITY_ANALYSIS.md` catalogs ~100 TODO items not touched here.

## If something's still broken

Tell me specifically which screen and at what terminal size. The fastest
reproduction is:

```bash
script -q -c "stty cols 120 rows 40; ssh -p 2222 tester@localhost" /tmp/tv.log
```

That gives me a raw byte log I can decode frame-by-frame.
