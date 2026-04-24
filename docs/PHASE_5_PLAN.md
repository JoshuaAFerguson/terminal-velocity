# Phase 5 Plan — Meta Depth & Polish

**Status:** planning. No code yet.
**Written:** 2026-04-23 after Phases 1–4 landed.
**Supersedes:** the "Phase 5 — Meta + polish" section of
`docs/GAME_COMPLETENESS_PLAN.md`, which called out the items without
scoping them. This doc breaks each one into achievable slices and
surfaces design questions before anyone writes code.

## Why This Phase Needs Planning

Phase 1–4 each had a single coherent goal ("core loop works", "world
breathes", "multiplayer visible"). Phase 5 is a grab-bag of "things
that make the game stick": fleets, faction wars, territory, tutorials,
storylines, leaderboards, a galactic newsreel, plus polish carry-overs
from Phase 4 (DB-persist marketplace, LISTEN/NOTIFY chat, duel-accept
routing).

Trying to execute these in parallel burns context and produces
half-done slices. Executing them in the wrong order wastes work —
e.g., faction wars that don't persist to a DB schema need rewriting
once faction membership grows beyond one server process.

Goal of this doc: decide the order, surface the design questions, and
commit only *after* an answer.

## Pillar-by-Pillar

### 5A. Infrastructure Hardening (polish carry-over from Phase 4)

**Scope:** the known gaps between "works in one session" and "works
across restarts / multiple gateways".

- Server-own the remaining per-session managers (`fleet.Manager`,
  `territory.Manager`, `tutorial.Manager`) — same recipe as
  news/chat/pvp/presence. Right now each session spins its own
  instance so fleet assignments disappear on disconnect.
- **DB-persist marketplace listings.** Currently in-memory — restart
  kills all auctions/contracts. Needs new `marketplace_listings`,
  `marketplace_bids`, `marketplace_contracts` tables and matching
  repository methods.
- **Challenger-side duel auto-transition.** When target accepts, the
  challenger's session doesn't know. Fix options:
  1. Poll: challenger's pvpPollTick reads `GetChallenge(id)` and if
     status flipped to Accepted, initialize combat_enhanced.
  2. Channel: server-side pub-sub via pvp.Manager callbacks.
  Option 1 fits the current polling architecture.
- **Chat over DB LISTEN/NOTIFY.** Polling costs a tick every second
  per connected session. A LISTEN/NOTIFY channel on a `chat_messages`
  table lets clients re-render on actual events. Also survives
  restart.

**Dependencies:** none — isolated polish.
**Effort:** ~2 days across all four items.
**Output:** Phase 4 promotes from "demoable" to "production".

### 5B. Fleet Play

**Scope:** a player can hire, command, and lose NPC escorts.

**Existing scaffolding:** `internal/fleet/manager.go` has 20 methods:
GetOrCreateFleet, AddShip, SwitchFlagship, StoreShip, and presumably
hire/dismiss/upkeep. 764 LoC — most mechanics are there.

**TUI gap:** `internal/tui/fleet.go` exists but is mostly stubbed
(per the BUG_SECURITY_ANALYSIS triage — "TODO: Use friends manager
once integrated" pattern, repeated throughout fleet.go).

**Design questions to resolve first:**
1. What's the hiring source? A "Hire Escort" option at the Shipyard?
   A dedicated tab at the spaceport bar?
2. Does the escort follow the player through hyperspace jumps?
   Immediate answer: yes, or it's not really an escort. Mechanical
   answer: the jump cost equation has to account for them.
3. How does an escort fight? Auto-combat during PvE encounters?
   Needs a simple AI loop in `encounters.Generator` or `combat.AI`.
4. Upkeep: flat credits/day? Per-ship based on class? How does the
   player notice their treasury draining?

**Effort:** ~3 days. Most of the work is TUI wiring + design, not
new engine logic.

### 5C. Faction Wars + Diplomacy

**Scope:** player-founded factions can declare war, negotiate peace,
and contest star systems.

**Existing scaffolding:** `internal/diplomacy/manager.go` (832 LoC,
15 methods) has FormAlliance, DisbandAlliance, DeclareWar,
ProposeTruce. Totally unwired to any TUI — no `diplomacyManager`
field on Model.

**Design questions:**
1. **What's a "faction war"?** Mechanically:
   - Visible hostile status between two factions?
   - Auto-aggression from NPC ships tagged with the enemy faction?
   - Reputation damage on sight?
   - System capture mechanics (see 5D)?
2. **Who can declare?** Only faction leaders/officers?
   (`faction_officers` table exists.) What about the player's
   personal reputation vs. NPC governments?
3. **How does peace work?** Propose → vote among officers → accept?
   Credit ante as a peace offering?

**Effort:** ~5 days for a lean MVP. Real faction PvP at scale is
weeks.

**Dependency:** requires shared `faction.Manager` + 5A's
persistence work (war state must survive restart).

### 5D. Territory Capture

**Scope:** a faction can claim a star system or station; players
docked there pay a tax/fee; controlling faction earns revenue.

**Existing scaffolding:** `internal/territory/manager.go` is only
**88 LoC / 5 methods** — mostly a skeleton. This one needs real
engine work, not just wiring.

**Design questions (a lot):**
1. Granularity: system-level or per-planet? EV used per-planet
   allegiance.
2. What triggers a capture? Combat wins? Paying an ante? Both?
3. Uncontrolled systems — default NPC governments keep them?
4. Revenue: percentage of every trade at the docked planet? Fixed
   docking fee?
5. Defence: can other factions challenge control by winning X
   combats in system over Y days?
6. Persistence: new `territory_control` table
   (system_id, faction_id, captured_at, revenue_ytd)?

**Effort:** ~7 days including the engine work. Highest-complexity
Phase 5 item.

**Dependency:** 5C (faction wars) should land first so "control"
has a consumer beyond vanity.

### 5E. Storylines (P3.4 carry-over)

**Scope:** multi-step faction questlines. "Earn the trust of the
Rigel Outer Marches", branching on player choice.

**Existing scaffolding:** `internal/factioncontent/manager.go`
(832 LoC, 17 methods) already has CreateFactionMission,
UpdateMissionProgress, and event callbacks. Well-structured enough
that the content authoring is the bottleneck, not the engine.

**Design questions:**
1. **Who writes the content?** The engine supports N questlines,
   but we ship with zero. Do we author 1–2 as reference?
2. **Unlock gate?** Reputation tier with the faction? Pilot licence
   (Phase 3)?
3. **UI surface.** New "Faction Contacts" tab at the spaceport?
   Or just exposed through the existing mission board filtered by
   faction?

**Effort:** ~2 days for wiring + 1 questline. More for richer
content.

**Dependency:** depends on reputation persisting (already done in
Phase 2).

### 5F. Tutorials

**Scope:** a first-run walkthrough that teaches the core loop
(land → trade → jump → combat) and doesn't get in the way of
returning players.

**Existing scaffolding:** `internal/tutorial/manager.go` (597 LoC,
16 methods) has RegisterTutorial, AddTrigger, InitializePlayer,
CompleteStep, trigger-based step advancement. Model calls
`InitializePlayer` + `HandleTrigger(FirstLogin)` at login already.

**What's missing:** the *content* — what steps do we ship? And
a UI overlay that renders step hints over the active screen.

**Design questions:**
1. Step format — toast at screen bottom? Modal dialog? Sidebar
   callout?
2. Skip path — global "I know what I'm doing" vs. per-step skip.
3. Tutorial screens: one per core loop beat? Or one overlay that
   re-uses existing screens with arrows?

**Effort:** ~2 days wiring + design + 5–7 tutorial steps.

### 5G. Leaderboards Across Categories

**Scope:** expose the existing per-category leaderboards (combat,
trade, exploration, wealth, reputation, missions) in the
Leaderboards screen rather than the current single Overall ranking.

**Existing scaffolding:** `internal/leaderboards/manager.go`
(260 LoC, 12 methods) already computes rankings per category. The
TUI screen renders one "Overall Rankings" view.

**Effort:** ~half a day. Mostly adding tabs + per-tab render
branches.

### 5H. Galactic Newsreel

**Scope:** a time-ordered feed of significant events (kills, big
trades, system captures, faction wars, pirate raids) visible to all
players. Already have NEWS; the newsreel is the *live feed* vs. the
per-session article list.

**Existing scaffolding:** news.Manager is server-wide and already
writes articles on trades, combat, achievements, system events,
faction changes. What's missing is a dedicated ticker UI (top-of-
screen crawl? chat-like scrolling list?).

**Effort:** ~1 day. Presentation layer on existing data.

## Proposed Execution Order

Numbered by recommended order, not priority:

1. **5A (infra polish)** — unblocks everything else. Fleet/territory
   need server-wide managers; marketplace needs DB.
2. **5F (tutorials)** — small, self-contained, highest onboarding ROI.
3. **5G (leaderboard tabs)** — half a day, easy win.
4. **5H (newsreel ticker)** — one day, uses existing data.
5. **5B (fleet play)** — needs 5A done.
6. **5E (storylines)** — content-heavy but independent once wired.
7. **5C (faction wars)** — needs 5A persistence.
8. **5D (territory)** — capstone, depends on 5C.

This ordering front-loads the small wins (5G, 5H, 5F) so the game
gains onboarding + polish early, while the biggest items (5B, 5C,
5D) come after the infra hardening that makes them scalable.

## Immediate Next-Slice Candidates

Pick one to start:

- **A. 5A-1 — server-own the remaining managers.** 3 packages
  (fleet, territory, tutorial). Follow the recipe used for
  news/chat/pvp/presence. Maybe 2 hours.
- **B. 5A-2 — marketplace DB persistence.** New table, new repo,
  wire existing manager's CreateAuction/Bid/Complete to persist.
  Maybe 4 hours.
- **C. 5F — tutorials.** Write the trigger content + overlay
  renderer. Maybe 1 day.
- **D. 5G — leaderboard category tabs.** Half a day, trivial
  risk.
- **E. 5H — newsreel ticker.** One day, touches the main menu
  layout (room for a permanent headline strip?).

My recommendation: **A** first, so subsequent work inherits
server-wide managers by default. B right after because it unblocks
marketplace durability. Then **G** as a palate cleanser.

## Decisions Made (2026-04-23)

Answered the four open questions and rebased the plan accordingly:

1. **Multi-region expected.** Chat / presence / PvP / marketplace
   must work across multiple server instances. `LISTEN/NOTIFY`
   alone isn't sufficient long-term — plan lands a broker (NATS or
   Redis Streams) during 5A. For beta we can stay single-region
   and swap the event transport later, but every new persistence
   schema introduced in 5A should be shaped so a broker swap
   doesn't require another migration.
2. **Faction war design doc exists** at `docs/FACTION_RELATIONS.md`
   §"Faction War Mechanics". War declaration → border war zones →
   NPC fleet engagements → amplified reputation swings → war-
   materials price spikes → territorial resolution. 5C scope
   follows this spec rather than inventing one.
3. **Tutorials are a beta blocker.** 5F moves ahead of 5A in the
   execution order — the overlay + step content must ship before
   inviting new players.
4. **Ship with 15 storylines.** 5E grows from "1–2 as reference" to
   "author 15 questlines". Content dominates engine; plan as a
   dedicated milestone with a shared structure template and per-
   faction authoring.

### Updated Execution Order

1. **5F tutorials** — beta blocker; small engine work, most of it
   is content + overlay UX.
2. **5A infra hardening** — server-own remaining managers,
   DB-persist marketplace, challenger-side duel routing. Broker
   deferred until beta has real cross-instance traffic.
3. **5G leaderboard categories** — half-day palate cleanser.
4. **5H newsreel ticker** — one day, uses existing data.
5. **5B fleet play** — needs 5A.
6. **5E storylines × 15** — largest content block. Could run in
   parallel with 5B for content authoring.
7. **5C faction wars** — implement per `docs/FACTION_RELATIONS.md`.
8. **5D territory capture** — capstone; hardest item.
