# Terminal Velocity — Game Completeness Plan

**Goal:** a terminal-based, multiplayer Escape Velocity (+ Nova + Override)
that is fun in single-player and adds real multiplayer depth on top.
**Written:** 2026-04-23 after confirming login → launch → land → trade works
end-to-end with seeded market data.
**Owner:** maintainer.

The codebase already ships ~74k lines of Go across 30 gameplay packages —
**most of the engine exists, the gaps are wiring, data seeding, and content.**
This plan maps Escape Velocity pillars to the current code, flags what's
working vs stubbed vs missing, and proposes a five-phase roadmap so the
MVP shipping state is clear.

## EV-Style Gameplay Pillars

Each pillar is the minimum experience a single-player EV fan expects. A
multiplayer Terminal Velocity must also deliver each, then layer social
systems on top.

| # | Pillar | What it means in TV |
|---|---|---|
| P1 | **Persistent universe**           | Procedural star systems, planets, jump routes, governments, news. |
| P2 | **Flying & navigation**           | 2D viewport, target cycling, jump-to-system, random encounters. |
| P3 | **Trading**                       | Per-planet commodity markets with supply/demand, legal/illegal goods. |
| P4 | **Ship ownership & progression**  | Buy/sell/upgrade ships, swap outfits, equip weapons, fuel/repair. |
| P5 | **Combat**                        | Shields/armor/hull, weapons with ammo/cooldowns, AI opponents, loot. |
| P6 | **Missions & storylines**         | Spaceport bar missions (freight, bounty, escort, scan, secret cargo). |
| P7 | **Reputation & factions**         | Per-government reputation, rank, access gates, faction storylines. |
| P8 | **Economy dynamics**              | Price fluctuation, trade routes, smuggling, stock scarcity. |
| P9 | **Hailing / bribery / boarding**  | Interact with ships — hail, bribe patrols, disable + board for capture. |
| P10 | **Meta / progression**           | Combat rating, wealth, pilot licences, achievements, leaderboards. |

### Multiplayer additions (not in original EV)

| # | Feature | Description |
|---|---|---|
| M1 | **Live presence**                | See other players' ships in system; target/hail them. |
| M2 | **Player chat**                  | Global / system / faction / DM channels. |
| M3 | **Player-vs-player combat**      | Duel, bounty contracts, arena. |
| M4 | **Player-run factions**          | Create factions, invite members, ranks, shared loadouts. |
| M5 | **Shared economy**               | Player actions move planet prices; trade routes between worlds. |
| M6 | **Player trade + marketplace**   | P2P offers, auction listings, bounty board. |
| M7 | **Fleet play**                   | Hire escorts, mutual protection, coordinated raids. |
| M8 | **Territory & diplomacy**        | Faction control of systems, treaties, wars. |

## Current State vs Pillars

Read: ✅ working end-to-end · 🟡 partial (package exists, wiring/data gaps) ·
⛔ missing · 🧪 package exists but untested via gameplay.

### Single-player pillars

| Pillar | State | Known gaps |
|---|---|---|
| P1 Universe                    | ✅ | `genmap -save` seeds 100 systems; news/events packages exist but no content generator. |
| P2 Navigation + viewport       | 🟡 | Jump routes and viewport work; **jump execution untested**, encounters on jump not surfaced, no radar-from-real-data yet. |
| P3 Trading                     | ✅ | Market seeds on first visit (this session). Prices static — price model per planet/category not implemented. |
| P4 Ship management             | 🟡 | Own a starter shuttle; **buying new ships untested**, outfit install untested end-to-end, repair/refuel works via landing. |
| P5 Combat                      | 🧪 | `internal/combat` is 1760 LoC; encounter + PvP exist; **never exercised from a live jump**. |
| P6 Missions                    | 🧪 | `missions.GenerateMissions` generates; UI shows empty state; **no bar / no automatic mission generation on land**. |
| P7 Factions + rep              | 🧪 | 369 + 832 LoC across factions + content; **no UI surfaces reputation per government**; governments are assigned to systems but inert. |
| P8 Economy dynamics            | ⛔ | Prices never change; no supply/demand feedback loop; trade routes package exists but inert. |
| P9 Hail / bribe / board        | ⛔ | Space-view has "H" hail + "F" fire keys; hail does nothing today; no bribe or board. |
| P10 Meta progression           | 🟡 | Combat rating tracked, leaderboards show #1; no licence system, no pilot history. |

### Multiplayer pillars

| Pillar | State | Known gaps |
|---|---|---|
| M1 Live presence               | 🟡 | Presence manager exists; space-view queries it; **other players don't actually render in your viewport yet** (need integration test). |
| M2 Player chat                 | 🧪 | Chat screen renders; **messages aren't broadcast across sessions yet** (manager in-memory only). |
| M3 PvP combat                  | 🧪 | PvP screen works; arena package has tournaments; **no end-to-end duel tested**. |
| M4 Player factions             | 🧪 | Create/invite/rank exists in managers; UI stubs tour the feature; end-to-end "create faction → invite → rank member → shared loadout" untested. |
| M5 Shared economy              | ⛔ | Nothing yet. When one player sells, prices stay flat. |
| M6 Marketplace / trade         | 🟡 | Marketplace screen renders; listings exist in DB; **posting + buying untested**. |
| M7 Fleet play                  | 🧪 | Fleet manager + screen exist; hire-escort path untested. |
| M8 Territory / diplomacy       | 🧪 | 88 LoC territory + 832 LoC diplomacy; **zero UI surfacing**. |

## The Real Blocking Issues

1. **The manager pattern isn't persisted.** Most managers (chat, presence,
   factions, missions, fleets) hold state in memory. Single-process state
   dies on restart. Once a second game server exists, they don't share.
   Every manager needs a persistence boundary.

2. **Data seeding is missing in five places.**
   - Markets: done this session.
   - Missions: generated on demand, but never automatically — the bar is empty.
   - News: manager exists, no cron / reaction wiring.
   - Ship/outfit/weapon catalogs: loaded from `models.StandardXxx` (in-memory,
     fine), but no persistence of per-planet/station stock.
   - Faction reputation starter values: players exist with `player_reputation`
     empty for every government.

3. **Several major features have UIs but no wiring.**
   - Outfitter+Shipyard from landing: route works, but haven't verified buy
     flow persists.
   - Missions board: screen exists, no `GenerateOnLand` tea.Cmd.
   - Hail: key binding, no command implementation.

4. **No ticker/tick loop.** Escape Velocity's economy/stock/news/random-event
   loop is tick-driven. TV has no equivalent yet, so the world feels frozen.

5. **Multiplayer presence is not integrated into the viewport.** `presence`
   manager knows who's in system; `spaceViewLoadedMsg` converts presence to
   spaceObjects; the conversion exists but I haven't verified a second-player
   actually renders on the first-player's viewport.

## Five-Phase Roadmap

Each phase has a clear "definition of done" (DoD). Phases are ordered so
each unblocks the next.

### Phase 1 — Ship the core loop end-to-end (this week)

Convert "can launch, can land, can trade" into "can earn credits, upgrade,
travel." Every item ships a tmux regression test + unit tests for repos.

- [ ] **Buy/sell commodities** works and persists (player credits + ship
      cargo + market stock). Tmux: buy 10 tons of food on Castor I, takeoff,
      verify cargo + credits updated.
- [ ] **Jump between systems** works and persists player.current_system.
      Tmux: from Castor → select connected system → Enter → land on a
      different system's planet.
- [ ] **Missions generate on land.** Landing triggers
      `missions.GenerateMissions(ctx, planet, faction, N)` if the planet
      has < 5 active missions. Missions screen shows them; accepting a
      freight run persists to `player_missions`.
- [ ] **Outfits & weapons install.** Outfitter [B]uy writes to ship_outfits
      / ship_weapons. Ship stats recalculate. Combat uses real damage from
      installed weapons (no placeholder).
- [ ] **Jumps trigger encounters.** A jump rolls `encounters.Generate`; a
      hit routes into ScreenEncounter, which can escalate into combat.

**DoD:** a new player can log in, fly to another system, dock, buy/sell
a commodity at a profit, install an outfit, and be challenged by a
pirate encounter — without touching the DB manually.

### Phase 2 — World breathes (2 weeks)

Tick loop + content generation. The universe must change between logins.

- [ ] **Game tick loop.** A goroutine per-system runs at `tick_rate`,
      updating market stocks (slow drift back to equilibrium), news
      articles on significant trades, mission expirations, encounter
      spawns.
- [ ] **Dynamic pricing.** Market price shifts when players buy/sell:
      stock decreases on buy, increases on sell; price uses
      `base * (1 + demand_factor - stock_factor)`. Bounded.
- [ ] **News system.** On player kill / big trade / faction event,
      `news.Manager.AnnounceXxx` writes to news table. Main-menu News
      shows the last 20.
- [ ] **Reputation consequences.** Attacking a government ship hurts rep
      with that gov and helps with its enemies. Cargo restrictions gated
      on rep (blocked at spaceport if too low).
- [ ] **Random events.** System-level random events (pirate raid,
      merchant convoy, bounty opened) from `events.Manager`.

**DoD:** a player who logs in daily sees the universe has changed — new
news, different prices, mission board rotated.

### Phase 3 — Pilot depth (2 weeks)

Make progression feel real.

- [ ] **Licences.** Pilot must earn licences (combat, trade, exploration)
      via achievements. Certain ships/outfits gated behind licences.
- [ ] **Pilot log.** History of jumps, kills, trades — visible in a
      "Pilot Record" screen. Persisted.
- [ ] **Ship customisation.** Rename ships; rename ports in your fleet;
      choose a transponder code.
- [ ] **Storyline hooks.** Faction questlines (multi-step missions with
      branching). Leverages the existing `factioncontent` package.
- [ ] **Achievements surface.** Achievement unlocks notify the player
      in-screen and persist to DB.

**DoD:** a player has a pilot record that grows with play and unlocks
content.

### Phase 4 — Multiplayer wiring (2 weeks)

Everything between players. Requires Phase 1.

- [ ] **Presence in viewport.** Other players in your star system render
      as ships in the 2D viewport with their real ship model and name.
      Targeting another player shows their real hull/shields over the
      wire.
- [ ] **Chat broadcast.** Chat messages go through a message bus
      (initially the DB via LISTEN/NOTIFY; eventually a dedicated
      broker). Multi-session tmux test: two players exchange messages.
- [ ] **PvP duel flow.** Player A hails player B → duel offer → accept →
      arena instance → loser pays ante, winner earns reward.
- [ ] **Bounties.** A criminal-rated player gets a bounty; any player
      who kills them claims it (via `capture` package).
- [ ] **Marketplace orders.** Player posts a buy/sell order for an item
      or ship; another player completes it. Persists to a new
      `marketplace_listings` table (or the existing package's schema).

**DoD:** two concurrent SSH sessions on the same stack can see each
other, fight, trade, and chat.

### Phase 5 — Meta + polish (ongoing)

Things that make the game stick.

- [ ] **Fleet play.** Hire NPC escorts that jump with you, follow in
      combat, and collect a cut.
- [ ] **Faction wars.** Player factions can declare war/peace via the
      diplomacy package. System control flips based on engagement.
- [ ] **Territory.** Capture an uncontrolled station; collect docking fees.
- [ ] **Tutorials.** The existing tutorial package hosts a proper
      first-run walkthrough.
- [ ] **Leaderboards across categories.** Exploration, combat, wealth,
      reputation — surface all via the leaderboards screen.
- [ ] **Newsreel / galactic news broadcast.** In-game feed of recent
      events visible to all players, time-ordered.

## Cross-Cutting Work

Not features — infrastructure every phase above relies on.

- **Ticker service.** `internal/tick` (new) — a goroutine pool that owns
  periodic work for markets, news, missions. Every manager registers
  handlers.
- **Message bus.** Start with Postgres `LISTEN/NOTIFY` for chat + presence;
  upgrade to NATS or Redis streams once load demands.
- **Manager persistence audit.** Every manager with in-memory state gets a
  `Save()` on stop and `Load()` on start. File some as "intentionally
  ephemeral" in doc.
- **Integration test harness.** Expand the tmux scripts to cover per-phase
  DoD scenarios as regression tests.
- **Server lifecycle.** Call `Shutdown()` on every manager during clean
  exit; currently tracked as S-3 in
  `docs/BUG_SECURITY_TRIAGE_2026-04-23.md`.
- **Observability.** The `/metrics` HTTP endpoint exists; add per-phase
  Prometheus counters (jumps, trades, kills, missions accepted).

## Immediate Next Slice (Phase 1, first three items)

Concrete checklist to attempt next session:

1. **Buy flow.** In `trading.go`, `case "b"` already exists; walk it end-
   to-end. Verify `trade.Manager.ExecuteBuy` (or equivalent) debits credits,
   credits cargo, reduces planet stock. Add a tmux capture that buys 10
   food and asserts the numbers.
2. **Jump flow.** Verify `executeJump` in `navigation.go` against a live
   server: from Castor select Omega Orionis, press Enter, confirm system
   changed and fuel dropped.
3. **Mission generation on land.** Hook `missions.Manager.GenerateMissions`
   into `dockCmd` (with a low-count guard). Missions screen already
   exists; this makes it populate.

Picking these three gets P3 (trading), P2 (navigation), P6 (missions)
end-to-end in one session.

## Tracking

- Phase DoD checkboxes live in this file; tick off as they land.
- Each commit referencing this plan should cite the phase/item number,
  e.g. `feat(trading): P1.1 buy flow persists to DB`.
- At the top of each session, scan `git log --grep 'P[1-5]'` to see how
  the phase is progressing.
