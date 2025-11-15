# Universe Design - Terminal Velocity

## Overview

The Terminal Velocity universe consists of approximately 100 star systems arranged in a 2D galaxy map. The universe is divided into distinct regions based on distance from Earth (Sol) and controlled by different NPC factions.

## Spatial Layout

```
                    [Auroran Empire - Edge]
                           ⧈ Alien Territory
                              |
                    [Frontier Worlds - Outer]
                    ⚑ Independent Colonies
                         /    |    \
              [Neutral Space - Mid]
         FTG Trading Hubs & Lawless Systems
            /           |           \
    [Core Systems - Inner Ring]
    UEF & ROM Territory (Earth & Mars)
              ⊕ Sol (Earth) ♂
```

### Region Distribution

**Core Systems (15-20 systems)**
- Distance from Sol: 0-30 light years
- Controlled by: United Earth Federation & Republic of Mars
- Characteristics:
  - High tech level (7-9)
  - High population
  - Strong law enforcement
  - Safe for traders
  - Best shipyards and outfitters
  - Low profit margins (stable prices)

**Mid Systems (30-40 systems)**
- Distance from Sol: 30-60 light years
- Controlled by: Free Traders Guild (stations), mixed governance
- Characteristics:
  - Medium tech level (5-7)
  - Trade hubs
  - Moderate law enforcement
  - Good trading opportunities
  - Occasional pirate activity
  - Player faction starter zone

**Outer Systems (30-40 systems)**
- Distance from Sol: 60-100 light years
- Controlled by: Frontier Worlds Alliance, some independent
- Characteristics:
  - Low tech level (3-6)
  - Low population
  - Weak law enforcement
  - High profit margins
  - Heavy pirate presence (Crimson Collective)
  - Player faction main territory

**Edge Systems (5-10 systems)**
- Distance from Sol: 100+ light years
- Controlled by: Auroran Empire (mysterious alien faction)
- Characteristics:
  - Alien tech level (10)
  - Unknown population
  - Strict border control
  - Rare contact
  - Dangerous to approach
  - Special missions/encounters

**Lawless Space (scattered throughout)**
- Asteroid belts, nebulae, dead systems
- No government control
- Crimson Collective bases
- Highest risk, highest reward

## NPC Factions

### 1. United Earth Federation (UEF) ⊕
**Core Territory**: Sol, Alpha Centauri, Tau Ceti, Epsilon Eridani

- **Government**: Federal democracy
- **Starting Reputation**: +10 (Liked)
- **Characteristics**: Bureaucratic, lawful, stable
- **Military**: Strong defensive navy
- **Economy**: Mixed economy, tech exports
- **Player Start**: Yes (most popular starting faction)

**Gameplay Role**:
- Safe starting area for new players
- High-paying but competitive missions
- Best law enforcement (safe trading)
- Lowest crime but also lowest margins
- Main quest line for "lawful good" players

### 2. Republic of Mars (ROM) ♂
**Core Territory**: Sol (Mars), influence in core systems

- **Government**: Republic, industrial focus
- **Starting Reputation**: +5 (Neutral-Friendly)
- **Characteristics**: Industrial, innovative, proud
- **Military**: Elite shipyards, quality over quantity
- **Economy**: Capitalist, ship manufacturing
- **Player Start**: Yes (industrial/combat focus)

**Gameplay Role**:
- Best ships and equipment available
- Higher prices but better quality
- Combat-focused missions
- Ship upgrade hub
- Rivalry/alliance with UEF creates interesting dynamics

### 3. Free Traders Guild (FTG) ¤
**Core Territory**: Sirius, Procyon (stations in many systems)

- **Government**: Guild council (merchant cooperative)
- **Starting Reputation**: 0 (Neutral)
- **Characteristics**: Neutral, mercantile, opportunistic
- **Military**: Defensive only (escorts)
- **Economy**: Pure capitalism, trade focus
- **Player Start**: Yes (trader focus)

**Gameplay Role**:
- Best trading prices and opportunities
- Trading missions and contracts
- Neutral in conflicts (trade with all sides)
- Banking and loans
- Bulk trading opportunities

### 4. Frontier Worlds Alliance (FWA) ⚑
**Core Territory**: Scattered outer systems

- **Government**: Loose confederation
- **Starting Reputation**: 0 (Neutral)
- **Characteristics**: Independent, tough, resourceful
- **Military**: Militia, weak but numerous
- **Economy**: Frontier economy (mining, agriculture)
- **Player Start**: Yes (hard mode)

**Gameplay Role**:
- High-profit trading (need goods)
- Dangerous space (pirate activity)
- Cheap ships but limited selection
- Freedom to operate without oversight
- "Firefly" style gameplay

### 5. Crimson Collective ☠
**Territory**: Hidden bases in lawless space

- **Government**: Pirate confederation
- **Starting Reputation**: -50 (Hostile)
- **Characteristics**: Lawless, dangerous, ruthless
- **Military**: Raiders and fast attack ships
- **Economy**: Black market, piracy
- **Player Start**: No (but can join later)

**Gameplay Role**:
- Main antagonist for early game
- Can join them (pirate gameplay)
- Black market for contraband
- High risk cargo runs through their space
- Bounty hunting targets

### 6. Auroran Empire ⧈
**Territory**: Edge systems (5-10 systems at map border)

- **Government**: Alien empire
- **Starting Reputation**: -10 (Distrusted)
- **Characteristics**: Mysterious, advanced, alien
- **Military**: Superior technology
- **Economy**: Unknown (exotic goods)
- **Player Start**: No

**Gameplay Role**:
- Late-game content
- Mystery and exploration
- Superior alien technology
- Difficult to gain reputation
- Special storylines
- May or may not be hostile (player choice driven)

## Territory Control Mechanics

### NPC Territory
- **Core Systems**: Fully controlled by major factions
  - Strong patrols
  - Law enforcement
  - Faction-specific stations and missions
  - Reputation affects access

- **Border Systems**: Contested or shared
  - Multiple faction presence
  - Political intrigue
  - Faction conflict zones
  - Reputation with multiple factions matters

- **Independent Systems**: Neutral/unaligned
  - Weak governance
  - Opportunity for player factions
  - Trade hubs
  - Mercenary work

### Player Faction Territory
- **Outer Systems Primary**: Player factions mainly operate in outer/mid systems
- **Control Mechanism**:
  - Can claim independent systems
  - Can contest border systems
  - Cannot (normally) take core NPC systems
  - Must maintain presence (station/base)

- **Benefits**:
  - Tax revenue from trade
  - Mission generation
  - Safe haven for members
  - Strategic positioning

- **Costs**:
  - Weekly upkeep
  - Defense requirements
  - Reputation impact with NPC factions

## Galaxy Generation Algorithm

### Phase 1: Create Star Systems
1. Place Sol (Earth) at center (0, 0)
2. Generate 99 other systems with names
3. Distribute in rough spiral/disk pattern
4. Calculate distances from Sol

### Phase 2: Assign Faction Control
1. **Core** (0-30 LY): UEF and ROM
   - Sol to UEF
   - Adjacent systems split between UEF/ROM
   - 15-20 systems total

2. **Mid** (30-60 LY): FTG stations + mixed
   - FTG controls 5-8 trade hub systems
   - Rest are independent or minor faction
   - 30-40 systems total

3. **Outer** (60-100 LY): Frontier Worlds
   - FWA controls 10-15 systems
   - Rest are independent
   - Crimson Collective hidden bases
   - 30-40 systems total

4. **Edge** (100+ LY): Auroran Empire
   - 5-10 systems at map edge
   - Mysterious, closed borders
   - Special encounter zone

### Phase 3: Tech Level Assignment
- Core systems: 7-9
- Mid systems: 5-7
- Outer systems: 3-6
- Edge systems: 10 (alien)
- Random variance ±1

### Phase 4: Generate Jump Routes
- Minimum spanning tree for connectivity
- Additional routes for interesting topology
- Some systems are "chokepoints"
- Limit: max 2-5 connections per system

### Phase 5: Place Planets
- 1-4 planets per system
- At least 1 habitable/station per system
- Services based on tech level and faction
- Trade good specialization

### Phase 6: Special Locations
- Wormholes (shortcuts)
- Asteroid fields (mining)
- Nebulae (hiding spots)
- Derelict stations (exploration)
- Ancient ruins (lore)

## Reputation System

### Reputation Tiers
- **Beloved** (75-100): Hero status, special missions, discounts
- **Respected** (50-74): Trusted ally, good missions
- **Friendly** (25-49): Allied, most missions available
- **Liked** (10-24): Positive relationship, standard access
- **Neutral** (-9 to 9): No special treatment
- **Disliked** (-24 to -10): Watched closely, higher prices
- **Unfriendly** (-49 to -25): Not welcome, few missions
- **Hostile** (-74 to -50): Attacked if caught with contraband
- **Hated** (-100 to -75): Attacked on sight

### Reputation Gains/Losses
- **Gain**: Complete missions, trade legally, defeat enemies
- **Loss**: Smuggle, attack faction ships, side with enemies
- **Allied factions**: +50% reputation gain
- **Enemy factions**: Reputation loss proportional to enemy gain

### Cross-Faction Effects
If player has +50 rep with UEF:
- ROM: +25 (allied)
- Crimson: -25 (enemy)
- FTG: 0 (neutral)
- Frontier: 0 (neutral)
- Auroran: 0 (don't care about human politics)

## Trade Route Design

### Core → Mid Routes (Safe, Low Profit)
- Manufactured goods (electronics, machinery) → Mid
- Raw materials (ore, food) → Core
- Profit margin: 10-20%

### Mid → Outer Routes (Medium Risk, Good Profit)
- Advanced goods → Outer (high demand, low supply)
- Raw materials → Mid (high supply, good demand)
- Profit margin: 30-50%

### Core → Outer Direct (High Risk, High Profit)
- Long route through pirate space
- Huge price differentials
- Profit margin: 60-100%

### Contraband Routes (Extreme Risk, Extreme Profit)
- Narcotics, weapons to restricted space
- Slaves (highly illegal, moral choice)
- Profit margin: 100-300%
- Risk: Ship seizure, reputation loss, combat

## Conflict Zones

### UEF-Crimson Border
- Constant skirmishes
- Bounty missions
- Convoy escorts
- Combat rating opportunities

### Frontier Space
- Pirate raids common
- Protection missions
- Lawless trading
- Player faction wars

### Auroran Border
- Mysterious encounters
- First contact missions
- Exploration
- Unknown dangers

## Dynamic Universe Events

- **Faction Wars**: NPC factions fight, shift borders
- **Pirate Raids**: Crimson Collective attacks increase/decrease
- **Economic Shifts**: Supply/demand changes from events
- **Alien Incursions**: Auroran activity (rare)
- **Political Changes**: Faction relations shift
- **Discoveries**: New systems, wormholes, ruins

## Player Impact

Players and player factions can:
- Shift economic trends through trading
- Claim outer systems for their faction
- Influence faction wars through mission choices
- Discover new routes and locations
- Affect reputation standings of their faction
- Trigger special events through actions

---

This universe design creates:
1. **Clear progression**: Core → Mid → Outer → Edge
2. **Risk vs Reward**: Safety vs Profit
3. **Factional intrigue**: Choose sides or stay neutral
4. **Multiplayer dynamics**: Player factions in outer systems
5. **Endgame content**: Auroran Empire mysteries
6. **Replayability**: Different starting factions, playstyles
