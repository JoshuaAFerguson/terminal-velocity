# Bug & Security Triage — 2026-04-23

**Supersedes:** large portions of `BUG_SECURITY_ANALYSIS.md` (dated 2025-11-15,
Phase 8 / ~37k LoC).
**Current state:** Phase 20+, ~78k LoC Go.
**Method:** fresh `grep -rnE '(TODO|FIXME|XXX|HACK)'` + `panic(` + stale-claim
re-verification.

## TL;DR

The 2025-11-15 document's headline claim — **"~100 incomplete features (TODOs)"** —
is no longer true. The current tree carries **4** TODO markers, all minor
and well-scoped. Most of the flagged gaps (marketplace, fleet, friends,
notifications, arena matchmaking, mining tracking, manufacturing, MST,
config loading, API panic) have been **completed** in the five months
since the doc was written.

## What the old doc got right that's still outstanding

### S-1. bcrypt cost = `DefaultCost` (10) — recommendation was 12
**File:** `internal/database/player_repository.go:49`
**Severity:** LOW (10 is still considered secure in 2026; 12 is a
"defense-in-depth" bump).
**Action:** either bump to `12` in a single line + note the ~4x
CPU-time increase per login, or accept 10 and close the item. No
migration needed — bcrypt-hashed passwords carry their cost inline, so
new logins hash at the new cost while old ones stay valid.

### S-2. ~80 `context.Background()` call sites in the TUI
**Pattern:** `ctx := context.Background()` inside a tea.Cmd, then
passed into a repo method, with no plumbing from the calling screen or
from the SSH session.
**Severity:** LOW-MEDIUM (depending on intent).
**Impact:**
- No per-command cancellation — if a user leaves a screen mid-query,
  the query still runs to completion on the background goroutine.
- No timeout — a slow query blocks its goroutine indefinitely.
- The SSH session's context is never propagated, so when the client
  disconnects, in-flight DB calls don't abort.

**Action:** the right fix is to plumb `ctx` from the SSH session
through the BubbleTea model into every tea.Cmd. That's a cross-cutting
refactor, reasonable for a dedicated sprint. Low urgency, moderate
reward.

### S-3. Managers not consistently `Shutdown()`-called on server exit
**File:** `internal/server/server.go`
**Severity:** LOW (goroutines leak only on shutdown, so it's a cleanup
concern, not a runtime concern).
**Action:** audit which managers have Shutdown() methods, add explicit
`defer m.Shutdown()` in the server lifecycle. Spot check: metrics and
ratelimit are called; others may not be.

## Remaining TODOs in the current tree

These are the only four TODO/FIXME markers in `internal/` and `cmd/`:

| # | Location | Current text | Assessment |
|---|---|---|---|
| T-1 | `internal/combat/ai.go:273` | `// Distance factor (TODO: when we have positions)` | Blocked on positional combat, not a defect. Leave. |
| T-2 | `internal/api/server/converters.go:34` | `// TODO: Add Z coordinate if needed for 3D space` | Speculative future feature. Delete the comment — YAGNI. |
| T-3 | `internal/tui/mail.go:877` | `// No credits attached (TODO: Add UI for credit attachments)` | Real feature gap, but cosmetic (mail UI only, not a security/safety issue). Optional sprint pickup. |
| T-4 | `internal/api/server/server.go:429` | `// TODO: Implement streaming` | Phase 2 gRPC work, tracked in `docs/ARCHITECTURE_REFACTORING.md`. Leave. |

**Recommended action:** delete T-2 (pure speculation), keep T-1/T-3/T-4
as tracked work.

## Things the old doc claimed that are no longer true

Verified as resolved (spot-checked in the current tree):

- **"~100 TODOs across the codebase"** → 4. The managers called out
  (arena, capture, mining, manufacturing) have **zero TODO markers now**.
- **"Panic in `internal/api/client.go:124`"** → file has no `panic(`.
  The documented fix was applied.
- **"Marketplace/fleet/friends/notifications — UI exists, backend missing"**
  → zero TODOs in those files. Backend integration completed.
- **"Universe MST TODO in generator.go:252"** →
  `internal/game/universe/mst.go` implements Kruskal's algorithm;
  generator.go calls `UpdatedGenerateJumpRoutes`.
- **"Hardcoded config in cmd/server/main.go:77"** → file now accepts
  `-config configs/config.yaml` and passes it to `server.NewServer`.
- **"Test file panics"** → still present, still intentional. No-op.
- **"Database transaction recover+re-panic"** → still present, still the
  correct idiom. No-op.

## Real issues flagged in THIS triage (not in the old doc)

### T-5. Tracked binary `genmap` in the repo
**File:** `genmap` (root of tree)
**Severity:** LOW (hygiene)
**Impact:** `make build-tools` modifies the file, leaving a 14 MB
Mach-O diff in every `git status`. It pollutes diffs and would balloon
the history if ever committed.
**Action:** `git rm --cached genmap` and add it to `.gitignore`. The
binary is produced deterministically by `make build-tools`.

### T-6. Outfitting main menu has both "Outfitter" and "Advanced Outfitting"
**File:** `internal/tui/main_menu.go`
**Severity:** LOW (UX / drift).
**Context:** One of the items tracked in
`docs/ENHANCED_VARIANT_CONSOLIDATION.md`. Product call pending.

### T-7. Anonymous SSH layer, authentication in TUI
**Files:** `internal/server/server.go`, `internal/tui/login.go`
**Severity:** MEDIUM (depending on threat model).
**Context:** The SSH server accepts any identity and authenticates in
the TUI. SSH public-key auth is implemented but not wired — the
`startGameSession` / `startRegistrationSession` entrypoints exist but
aren't called from the live login path. This is an architectural
choice, not a defect, but it means:
- SSH-layer rate limiting protects the connection, not the credential
  check.
- `accounts add-key` stores keys that aren't consulted.
- OTP/2FA scaffolding (`pquerna/otp` in `go.mod`) is unused.

**Action:** land SSH public-key auth as the primary path and fall
back to in-TUI login only for first-time or password reset — before
this is exposed beyond trusted users. Not urgent for current
development/demo traffic.

## Prioritized action list

Sorted by reward / effort, smallest and highest-signal first:

1. **[T-2] Delete speculative Z-coordinate TODO** — 1-line change.
2. **[T-5] Stop tracking the `genmap` binary** — 2-line change
   (`.gitignore` + `git rm --cached`).
3. **[S-1] Bump bcrypt cost to 12** — 1-line change, tiny login
   latency penalty. Do only if threat model warrants.
4. **[S-3] Call `Shutdown()` on all managers at server exit** — 30-60
   minute audit, small diff, closes the goroutine-leak concern cleanly.
5. **[T-3] Credit attachments for mail** — real UI feature, needs
   design first.
6. **[S-2] Plumb `ctx` from SSH session through the TUI** — multi-file
   refactor. Defer until a sprint is budgeted.
7. **[T-7] Land SSH public-key auth as the primary path** — threat-
   model-dependent, needs design + migration plan.

## Document hygiene

- Keep `BUG_SECURITY_ANALYSIS.md` as a historical snapshot (Nov 2025,
  Phase 8). Add a banner at the top linking here.
- This file tracks the current state; re-run a fresh triage quarterly
  or at each major phase boundary.
