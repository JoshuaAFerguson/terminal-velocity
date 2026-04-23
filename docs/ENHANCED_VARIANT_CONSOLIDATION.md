# `_Enhanced` Screen Variant Consolidation

**Status:** analysis + migration recipe. Requires product decisions before execution.
**Last updated:** 2026-04-23
**Owner:** TBD

## Summary

The TUI carries five parallel screen pairs — a non-enhanced and a `*_enhanced`
variant for the same conceptual screen. Both sides are reachable in production
and both have real users (main menu, landing, space view), so features and
bug fixes drift between them. Consolidation removes that drift class but
requires a product call per pair: which variant stays, and where does any
unique behavior land.

## The pairs

| Conceptual screen | Non-enhanced file | Enhanced file | Non-enhanced real? | Enhanced real? |
|---|---|---|---|---|
| Combat    | `combat.go` (802 L)    | `combat_enhanced.go` (1031 L) | Yes — `playerRepo.UpdateCredits` | Partial — "in production, this would be loaded from shipRepo" |
| Navigation| `navigation.go` (452 L)| `navigation_enhanced.go` (452 L) | Yes — `systemRepo.GetSystemByID` | **No** — hardcoded "Alpha Centauri", "Proxima Centauri", "Sol" sample data |
| Outfitter | `outfitter.go` (743 L) | `outfitter_enhanced.go` (860 L) | Yes — `playerRepo.UpdateCredits` | Yes — `outfittingManager` (manager abstraction over repos) |
| Shipyard  | `shipyard.go` (874 L)  | `shipyard_enhanced.go` (515 L) | Yes — `systemRepo.GetPlanetByID` | Yes — `shipRepo.Create` |
| Trading   | `trading.go` (689 L)   | `trading_enhanced.go` (691 L) | Yes — `systemRepo.GetPlanetByID` | Yes — `shipRepo.AddCargo` |

## Caller map

Who routes to the non-enhanced vs enhanced entry for each screen:

| Screen | Non-enhanced callers | Enhanced callers |
|---|---|---|
| Navigation | main_menu, traderoutes, tutorial, game | space_view (`j` jump, `m` map) |
| Trading    | main_menu, tutorial, game | landing (×2) |
| Shipyard   | main_menu, tutorial, game | landing (×2) |
| Outfitter  | main_menu ("Outfitter"), tutorial, game | main_menu ("Advanced Outfitting"), landing (×2), space_view (`o`) |
| Combat     | encounter (×4), tutorial, game | pvp, space_view (`f` fire) |

Note the Outfitter row: **main menu has both entries**. That's a user-visible
duplication, not just code drift.

## Why this matters

Concrete drift hazards already observable in the tree:
- A `printableRuneString` fix landed in all input screens in the `fix(tui):
  cell-aware width math and paste-safe text input` commit; any input in a
  future _enhanced variant would have been silently missed.
- `combat_enhanced.go` carries an explicit `// In production, this would be
  loaded from shipRepo` TODO. The non-enhanced path already loads it.
- `navigation_enhanced.go` has a hardcoded in-file fallback:
  `m.navigationEnhanced = newNavigationEnhancedModel()` runs when
  `len(m.navigationEnhanced.systems) == 0`, meaning the screen lies to the
  user about what's reachable.

## Per-pair recommendation

### Navigation — **clear win, low risk**
`navigation_enhanced.go` is a stub with hardcoded sample systems and no DB
integration. `navigation.go` is the real screen (connected via
`systemRepo.GetSystemByID` and the encounters package).

Recommended action:
1. Update `space_view.go` — change both `ScreenNavigationEnhanced`
   transitions (`j` jump, `m` map) to `ScreenNavigation`.
2. Delete `internal/tui/navigation_enhanced.go`.
3. Remove `ScreenNavigationEnhanced` from the `Screen` enum in
   `internal/tui/model.go`. *Note:* `Screen` is an `iota` enum — verify
   no code path encodes specific integer values for persistence or IPC
   before removing (`grep -rn 'Screen(' internal/`).
4. Delete `navigationEnhanced` field from `Model`, drop
   `newNavigationEnhancedModel()` calls, drop the
   `updateNavigationEnhanced` / `viewNavigationEnhanced` cases.
5. Delete `navigationEnhancedDataMsg` from `messages.go`.
6. Rewrite `navigation_test.go`: drop the 7 test cases scoped to
   `ScreenNavigationEnhanced`; the three that exercise `ScreenSpaceView →
   NavigationEnhanced` should now assert `ScreenSpaceView → Navigation`.

Risk: one product trade-off. Space-view `m` was intended to be a
**visual star map**, not the jump interface. Merging means losing the
placeholder map entirely until a real map screen is built. That's
probably fine — the current map is hardcoded lies — but flag it.

### Combat — **mostly mechanical, one product call**
`combat_enhanced.go` is the newer turn-based tactical screen (PvP
support, loot, distance/closing speed). `combat.go` is the simpler older
combat. PvP only goes through enhanced; encounters only go through
plain.

Recommended action:
- **Keep enhanced**, migrate encounter and tutorial callers over.
- Reconcile: the enhanced version needs to also handle the single-player
  encounter flow that today calls non-enhanced. Verify
  `combat_enhanced.go` accepts an encounter-started entry point.
- Then same 6-step removal recipe as Navigation, targeting `combat.go`.

Risk: encounter-start flow regression. Needs explicit test coverage
before deletion (tmux capture pre/post).

### Outfitter — **largest change, needs product input first**
Both sides have real code and **both are in the main menu**. The
enhanced path adds loadouts and an inventory browser. The non-enhanced
path is simpler.

Blocker: is "Advanced Outfitting" a separate feature (keep both) or a
redundant entry (pick one)?

If pick-one:
- Recommended winner = `outfitter_enhanced.go` (loadout management is a
  real feature).
- Migrate `outfitter.go` callers (main menu "Outfitter", tutorial, game
  hub) to `ScreenOutfitterEnhanced`.
- Remove `outfitter.go` and `ScreenOutfitter`.
- Rename: drop the `Enhanced` suffix once there's only one.

### Shipyard — **needs feature diff**
Both sides have repo code. Sizes differ (874 L vs 515 L) so they've
diverged substantially. Before migrating, run a feature diff:
- Ship browsing, purchasing, selling, trade-in?
- Which side handles planet constraints, reputation gating, tech-level
  checks?

Recommended action: produce a capability table (rows = features, cols =
variants, cells = implemented Y/N/partial), then pick the side with
fewer partials and port the missing features across before deletion.

### Trading — **same shape as Shipyard**
Near-identical line counts (689 vs 691) but different repo methods
(`systemRepo.GetPlanetByID` vs `shipRepo.AddCargo`) suggest they solve
slightly different sub-problems. Same capability-diff recipe applies.

## Generic migration recipe

For any pair where a winner is chosen:

1. **Call-site redirect.** `grep -rn 'Screen<Variant>Enhanced' internal/`
   and rewrite each to the winner. Mirror-edit `views.go` routing.
2. **Field removal.** Delete the loser's sub-model field from `Model`
   and any `newXxxModel()` initializer call in `NewModel` /
   `NewLoginModel` / `newModel` (there are multiple constructors that
   must stay in sync).
3. **Enum removal.** Drop the loser's `Screen` constant. Because
   `Screen` is iota-based, check for any caller that compares against
   the underlying int (rare but possible in persistence/IPC).
4. **Switch-case cleanup.** Drop the loser's `case` clauses from the
   `Update()` and `View()` method switches in `model.go`.
5. **Messages/commands.** Delete any `Xxx<Variant>EnhancedMsg`,
   `<variant>EnhancedDataMsg`, and `tea.Cmd` helpers scoped to the loser.
6. **Tests.** Rewrite or delete `<variant>_test.go` cases.
7. **Comments.** Each enhanced file ends with a stale `// Add
   Screen<Variant>Enhanced constant to Screen enum when integrating`
   comment — remove if keeping enhanced; it's obsolete regardless.
8. **File deletion.** `git rm internal/tui/<loser>.go`.
9. **Rename.** If keeping enhanced, rename files and symbols to drop
   the `Enhanced` suffix. Two-step: (a) delete loser, land, green; (b)
   rename winner, land, green. Don't combine — rename diffs drown the
   semantic delete.
10. **Verify.** `go test -race ./internal/tui/` +
    `scripts/tmux_menu_walk.sh` to confirm no screens panic.

## Deferred decisions that must be answered first

1. Keep or remove the "Advanced Outfitting" main-menu entry
   (Outfitter pair)?
2. Is there an intended "system map" screen distinct from "navigation"
   that justified navigation_enhanced's map-view framing?
3. For Shipyard / Trading: is the landing-page variant meant to be the
   "docked-at-planet" flavour while the main-menu variant is the "in
   space" flavour? If so, the names should say that, not
   `_Enhanced`.
4. Are there persistence points (database, session replay, saved
   URLs) that encode `Screen` integer values? Affects the enum-removal
   mechanics.

## Tracking

When a pair is consolidated, leave a one-line record here:

- [ ] Navigation — merged to `navigation.go`, enhanced removed (date, PR)
- [ ] Combat — merged to `combat_enhanced.go`, encounter/tutorial migrated
- [ ] Outfitter — **blocked on product decision**
- [ ] Shipyard — **blocked on feature diff**
- [ ] Trading — **blocked on feature diff**
