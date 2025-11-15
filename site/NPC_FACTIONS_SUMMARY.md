# NPC Faction System - Summary

## Six Major Factions

### 1. United Earth Federation (UEF) ⊕
**"The Core Government"**
- **Location**: Core systems (Sol, Alpha Centauri, Tau Ceti, Epsilon Eridani)
- **Distance from Sol**: 0-30 LY
- **Tech Level**: 8
- **Starting Reputation**: +10 (Liked)
- **Player Start**: Yes (most popular)

**Characteristics**:
- Bureaucratic but stable
- Democratic federal government
- Strong defensive navy
- Law and order focus
- Main human government

**Gameplay**:
- Safest starting zone
- Good missions, standard pay
- Excellent facilities
- Low trading margins (stable economy)
- Tutorial faction

---

### 2. Republic of Mars (ROM) ♂
**"The Industrial Powerhouse"**
- **Location**: Sol system (Mars), influence in core
- **Distance from Sol**: 0-30 LY
- **Tech Level**: 9
- **Starting Reputation**: +5 (Neutral-Friendly)
- **Player Start**: Yes (combat/industrial focus)

**Characteristics**:
- Independent from Earth
- Industrial and technological leader
- Best shipyards in human space
- Proud and innovative
- Capitalist economy

**Gameplay**:
- Best ships and equipment
- Higher prices, better quality
- Combat-focused missions
- Ship upgrade destination
- Alliance with UEF (but independent)

---

### 3. Free Traders Guild (FTG) ¤
**"The Merchant Network"**
- **Location**: Sirius, Procyon, trade stations throughout
- **Distance from Sol**: 30-60 LY (mid systems)
- **Tech Level**: 7
- **Starting Reputation**: 0 (Neutral)
- **Player Start**: Yes (trader focus)

**Characteristics**:
- Merchant cooperative, not a government
- Strictly neutral in conflicts
- Controls key trade routes
- Profit and freedom above all
- Safe havens for all traders

**Gameplay**:
- Best trading prices (10% bonus)
- Trading-focused missions
- Bulk cargo opportunities
- Banking and contracts
- "Space truckers" faction

---

### 4. Frontier Worlds Alliance (FWA) ⚑
**"The Independent Colonists"**
- **Location**: Scattered outer systems
- **Distance from Sol**: 60-100 LY
- **Tech Level**: 5
- **Starting Reputation**: 0 (Neutral)
- **Player Start**: Yes (hard mode)

**Characteristics**:
- Rejected core government control
- Rugged individualists
- Loose confederation
- Poor but resourceful
- Freedom above all

**Gameplay**:
- High-profit trading (desperate for goods)
- Dangerous space (pirate raids)
- Limited facilities
- "Firefly/frontier" gameplay
- Challenging start

---

### 5. Crimson Collective ☠
**"The Pirates"**
- **Location**: Hidden bases in lawless space
- **Distance from Sol**: Outer systems and asteroid belts
- **Tech Level**: 6
- **Starting Reputation**: -50 (Hostile)
- **Player Start**: No (can join later)

**Characteristics**:
- Pirate confederation
- Outlaws and smugglers
- Prey on merchant shipping
- Enemy of all legitimate governments
- Black market dealers

**Gameplay**:
- Main early-game antagonist
- Can join them (pirate path)
- Black market trading (20% discount)
- High-risk, high-reward
- Bounty targets

---

### 6. Auroran Empire ⧈
**"The Aliens"**
- **Location**: Edge systems (100+ LY from Sol)
- **Distance from Sol**: 100-120 LY
- **Tech Level**: 10 (superior)
- **Starting Reputation**: -10 (Distrusted)
- **Player Start**: No

**Characteristics**:
- Alien civilization
- First contact in 2245
- Advanced technology
- Mysterious motives
- Strict borders

**Gameplay**:
- Late-game content
- Mystery and exploration
- Superior alien tech (expensive)
- Difficult to gain reputation
- May be hostile or peaceful (player-driven)
- Endgame storylines

---

## Territory Distribution

```
Core (0-30 LY):    UEF 70%, ROM 30% [15-20 systems]
Mid (30-60 LY):    FTG 25%, UEF 25%, Independent 50% [30-40 systems]
Outer (60-100 LY): FWA 40%, Independent 50%, Crimson 10% [30-40 systems]
Edge (100+ LY):    Auroran 100% [5-10 systems]
```

## Faction Relations

### Alliances
- UEF ↔ ROM (strong military/economic alliance)

### Hostilities
- UEF ↔ Crimson (active war on piracy)
- ROM ↔ Crimson (defend trade routes)
- FTG ↔ Crimson (pirates threaten business)
- FWA ↔ Crimson (raid frontier colonies)

### Neutral
- Auroran with everyone (isolationist)
- FTG with legitimate governments (strict neutrality)

## Player Progression Path

### Early Game (Core Systems)
1. Start with UEF, ROM, or FTG
2. Learn trading in safe space
3. Basic missions
4. Upgrade from shuttle
5. Build reputation

### Mid Game (Mid Systems)
1. Trade routes to outer systems
2. Higher-profit missions
3. Combat with pirates
4. Better ships
5. Choose faction allegiance

### Late Game (Outer Systems)
1. Player faction territory
2. High-risk trading
3. Faction wars
4. Capital ships
5. Control systems

### End Game (Edge Systems)
1. Auroran contact
2. Alien technology
3. Mysteries and lore
4. Universe-changing events

## Key Features

**Dynamic Universe**:
- Faction strengths change based on player actions
- Territory can shift
- Wars can break out
- Economic changes from player trading

**Multiple Playstyles**:
- Lawful trader (UEF/FTG)
- Combat pilot (ROM/UEF)
- Frontiersman (FWA)
- Pirate (Crimson)
- Explorer (Auroran mysteries)
- Faction leader (player factions)

**Reputation Matters**:
- Affects prices, missions, access
- Cross-faction effects (allies/enemies)
- Can be beloved or hated
- Long-term consequences

**Player Agency**:
- Choose which factions to support
- Can switch allegiances (with penalties)
- Join pirate faction
- Affect faction wars
- Discover alien mysteries

## Implementation Status

✅ **Complete**:
- 6 NPC factions defined
- Territory distribution designed
- Reputation system designed
- Faction relationships mapped
- Universe generation algorithm
- Galaxy map visualization

🚧 **Phase 1 (Next)**:
- Universe generator implementation
- Database faction tables
- Reputation tracking
- Faction-based pricing

📋 **Future Phases**:
- Faction missions (Phase 5)
- Faction wars (Phase 6)
- Dynamic events (Phase 7)
- Auroran storyline (Phase 8)

## Files Created

- `internal/models/npc_faction.go` - Faction data structures
- `internal/game/universe/generator.go` - Universe generation with factions
- `docs/UNIVERSE_DESIGN.md` - Detailed universe structure
- `docs/FACTION_RELATIONS.md` - Faction politics and conflicts
- `docs/GALAXY_MAP.txt` - ASCII visual galaxy map
- `docs/NPC_FACTIONS_SUMMARY.md` - This file

---

The NPC faction system adds:
- **Structure**: Clear regions and progression
- **Lore**: Rich backstory and politics
- **Choice**: Multiple valid paths
- **Challenge**: Risk vs reward zones
- **Mystery**: Alien faction to discover
- **Endgame**: Meaningful late-game content
