# Combat & Encounter Screens

This document covers all combat-related UI screens in Terminal Velocity.

## Overview

**Screens**: 4
- Combat Screen
- Combat Enhanced Screen
- PvP Screen
- Encounter Screen

**Purpose**: Handle turn-based combat, PvP duels, and random encounter interactions.

**Source Files**:
- `internal/tui/combat.go` - Basic turn-based combat interface
- `internal/tui/combat_enhanced.go` - Advanced combat with tactical options
- `internal/tui/pvp.go` - Player vs Player combat
- `internal/tui/encounter.go` - Random encounter decision screen

---

## Combat Screen

### Source File
`internal/tui/combat.go`

### Purpose
Turn-based combat interface for engaging hostile ships, pirates, and enemies.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ COMBAT ENGAGED!                  [Sol System]          Shields: ██████░░░░ 60%┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                    TACTICAL DISPLAY                           ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║                          ◆                                    ║  ┏━━━━━━━━━━━━━┓
┃    ║                      Pirate Viper                             ║  ┃ YOUR SHIP   ┃
┃    ║                       [LOCKED]                                ║  ┣━━━━━━━━━━━━━┫
┃    ║                          ↓                                    ║  ┃ Corvette    ┃
┃    ║                   ~~~~ WEAPONS ~~~~                           ║  ┃             ┃
┃    ║                          ↓                                    ║  ┃ Hull: ██████┃
┃    ║                                                               ║  ┃       100%  ┃
┃    ║                          △                                    ║  ┃             ┃
┃    ║                      Your Ship                                ║  ┃ Shields:    ┃
┃    ║                                                               ║  ┃ ██████░░░░  ┃
┃    ║                                                               ║  ┃       60%   ┃
┃    ║        Distance: 1,850 km     Closing at 120 km/s            ║  ┃             ┃
┃    ║                                                               ║  ┃ Energy:     ┃
┃    ╚═══════════════════════════════════════════════════════════════╝  ┃ ████████░░  ┃
┃                                                                        ┃       80%   ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┗━━━━━━━━━━━━━┛
┃  ┃ ENEMY: Pirate Viper                                           ┃
┃  ┃ Hull: ████░░░░ 40%   Shields: ██░░░░░░ 25%   Weapons: Active ┃  ┏━━━━━━━━━━━━━┓
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃ WEAPONS     ┃
┃                                                                        ┣━━━━━━━━━━━━━┫
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ 1. Laser    ┃
┃  ┃ COMBAT LOG:                                                    ┃  ┃    Cannon   ┃
┃  ┃ > Pirate Viper is hailing you: "Prepare to die!"              ┃  ┃    [READY]  ┃
┃  ┃ > You fire Laser Cannon - HIT for 45 damage!                  ┃  ┃             ┃
┃  ┃ > Pirate fires Pulse Laser - MISS!                            ┃  ┃ 2. Pulse    ┃
┃  ┃ > Your shields absorb 30 damage from Pulse Laser              ┃  ┃    Laser    ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃    [READY]  ┃
┃                                                                        ┃             ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ 3. Missiles ┃
┃  ┃ YOUR TURN - Select Action:                                     ┃  ┃    [15/15]  ┃
┃  ┃                                                                ┃  ┗━━━━━━━━━━━━━┛
┃  ┃  [1] Fire Laser Cannon     [2] Fire Pulse Laser               ┃
┃  ┃  [3] Fire Missile          [E] Evasive Maneuvers              ┃
┃  ┃  [D] Defend (Boost Shields) [R] Retreat (Flee Combat)         ┃
┃  ┃                                                                ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [1-3] Fire Weapon  [E]vade  [D]efend  [R]etreat  [H]ail  [ESC] Menu        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Tactical Display**: Visual representation of combat positions
- **Ship Status Panels**: Your ship and enemy ship stats
- **Combat Log**: Scrolling history of combat actions
- **Weapons List**: Available weapons and ammo counts
- **Action Menu**: Available combat actions for current turn

### Combat Actions

**Offensive**:
- **Fire Weapon** - Attack with selected weapon
- **Missile** - Fire homing missile (limited ammo)

**Defensive**:
- **Evasive Maneuvers** - Increase dodge chance for 1 turn
- **Defend** - Boost shields, reduce incoming damage

**Special**:
- **Hail** - Attempt communication (bribe, surrender, etc.)
- **Retreat** - Attempt to flee combat (chance-based)

### Key Bindings
- `1-9` - Fire weapon from weapon list
- `E` - Evasive maneuvers
- `D` - Defensive stance (shield boost)
- `R` - Attempt to retreat/flee
- `H` - Hail enemy ship
- `S` - Special abilities (if equipped)
- `ESC` - Open combat menu (surrender, etc.)

### State Management

**Model Structure** (`combatModel`):
```go
type combatModel struct {
    playerShip    *models.Ship
    enemyShip     *models.Ship
    combatLog     []string
    turn          int
    playerTurn    bool
    weaponCooldowns map[string]int
    combatState   CombatState
    width         int
    height        int
}

type CombatState string

const (
    CombatStateOngoing  CombatState = "ongoing"
    CombatStateVictory  CombatState = "victory"
    CombatStateDefeat   CombatState = "defeat"
    CombatStateFled     CombatState = "fled"
)
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `combatInitiatedMsg` - Combat started
- `playerTurnMsg` - Player's turn to act
- `enemyTurnMsg` - Enemy AI taking turn
- `damageDealtMsg` - Damage applied
- `combatEndMsg` - Combat resolved
- `lootAwardedMsg` - Loot given on victory

### Data Flow
1. Combat initiated (random encounter, mission, player action)
2. Load combatants (player ship, enemy ship)
3. Roll initiative (who goes first)
4. **Player Turn**:
   - Display action menu
   - Player selects action
   - Calculate hit/damage
   - Apply effects
   - Update combat log
5. **Enemy Turn**:
   - AI selects action (5 difficulty levels)
   - Calculate hit/damage
   - Apply effects
   - Update combat log
6. Check combat end conditions
7. Award loot/experience on victory
8. Return to space view

### Combat Mechanics

**Turn Order**:
- Initiative based on ship speed/maneuverability
- Faster ships go first
- Turn order locked for duration of combat

**Hit Calculation**:
```go
func CalculateHit(attacker, defender *Ship, weapon *Weapon) bool {
    baseAccuracy := weapon.Accuracy
    attackerBonus := attacker.Stats.Targeting
    defenderBonus := defender.Stats.Evasion

    if defender.IsEvading {
        defenderBonus *= 2
    }

    hitChance := baseAccuracy + attackerBonus - defenderBonus
    roll := rand.Intn(100)

    return roll < hitChance
}
```

**Damage Calculation**:
```go
func CalculateDamage(weapon *Weapon, shieldsUp bool, armor int) int {
    baseDamage := weapon.Damage

    if shieldsUp {
        // Shields absorb damage first
        // Energy weapons better vs shields
        // Projectile weapons better vs hull
    }

    reducedDamage := baseDamage * (100 - armor) / 100
    return max(1, reducedDamage)  // Min 1 damage
}
```

**Weapon Types**:
- **Energy Weapons**: Laser cannons, plasma - effective vs shields
- **Projectile Weapons**: Railguns, mass drivers - effective vs hull
- **Missiles**: Homing, high damage, limited ammo

**AI Difficulty Levels** (5 levels):
1. **Easy**: Random actions, poor targeting
2. **Medium**: Basic tactics, 60% accuracy
3. **Hard**: Good tactics, weapon selection
4. **Very Hard**: Advanced tactics, focus fire
5. **Ace**: Perfect tactics, optimal weapon use

**Combat End Conditions**:
- **Victory**: Enemy hull reduced to 0
- **Defeat**: Player hull reduced to 0
- **Flee**: Successfully retreat (chance-based)
- **Surrender**: Player/enemy gives up

### Loot System

On victory:
- Credits (based on enemy ship value)
- Cargo (if cargo ship)
- Equipment (random drops, 4 rarity tiers)
- Reputation (with opposing faction)

**Rarity Tiers**:
- Common (70% chance)
- Uncommon (20% chance)
- Rare (8% chance)
- Legendary (2% chance)

### Related Screens
- **Combat Enhanced** - Advanced tactical options
- **PvP** - Player combat
- **Encounter** - Pre-combat negotiation
- **Space View** - Return after combat

---

## Encounter Screen

### Source File
`internal/tui/encounter.go`

### Purpose
Random encounter interaction screen for pirates, traders, police, distress calls.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ENCOUNTER                        [Sol System]                52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                                                               ║        ┃
┃    ║                          ◆                                    ║        ┃
┃    ║                     Pirate Raider                             ║        ┃
┃    ║                   "Red Skull" Faction                         ║        ┃
┃    ║                                                               ║        ┃
┃    ║                         ════>                                 ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║                          △                                    ║        ┃
┃    ║                       Your Ship                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║        Distance: 5,000 km     Approaching                    ║        ┃
┃    ║                                                               ║        ┃
┃    ╚═══════════════════════════════════════════════════════════════╝        ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ INCOMING TRANSMISSION                                                ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [Pirate Captain Redeye]                                            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  "Well, well... what do we have here? A nice fat trader cruising    ┃  ┃
┃  ┃   through our territory. Hand over your cargo and we'll let you     ┃  ┃
┃  ┃   live. Resist and... well, let's just say we could use the scrap   ┃  ┃
┃  ┃   metal from your ship."                                            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [Pirate Raider is demanding: 15 tons of cargo or 5,000 credits]   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR RESPONSE:                                                       ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ▶ [P]ay Bribe (5,000 cr)         - Avoid combat, lose credits     ┃  ┃
┃  ┃    [J]ettison Cargo (15 tons)     - Avoid combat, lose cargo       ┃  ┃
┃  ┃    [F]ight                        - Enter combat                   ┃  ┃
┃  ┃    [N]egotiate (Barter skill)     - Try to reduce demand           ┃  ┃
┃  ┃    [R]un                           - Attempt to flee (30% chance)  ┃  ┃
┃  ┃    [B]luff (Intimidate)            - Scare them off (risky)        ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Reputation with Red Skull: -45 (Hostile)                           ┃  ┃
┃  ┃  Your Combat Rating: Good (75/100)                                  ┃  ┃
┃  ┃  Enemy Combat Rating: Medium (55/100)                               ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Option  [Enter] Confirm  [I]nfo  [ESC] Cancel (Fight)          ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Visual Display**: Ships in encounter
- **Transmission**: NPC dialogue and demands
- **Response Options**: Player choices with outcomes
- **Stat Comparison**: Combat ratings, reputation

### Encounter Types

1. **Pirates** - Demand cargo/credits, hostile
2. **Traders** - Offer to trade, buy/sell
3. **Police** - Scan for illegal cargo
4. **Distress Call** - Ship in trouble, help requested
5. **Derelict** - Abandoned ship, salvage opportunity
6. **Patrol** - Faction patrol, reputation check

### Response Options

**Peaceful**:
- Pay bribe/ransom
- Jettison demanded cargo
- Trade/barter
- Assist (distress calls)

**Hostile**:
- Fight immediately
- Refuse and fight

**Evasive**:
- Flee (chance-based)
- Bluff/intimidate
- Negotiate (reduce demands)

### Key Bindings
- `↑`/`↓` or `J`/`K` - Select response option
- `Enter` - Confirm selection
- `I` - View detailed information about encounter
- `F` - Quick fight (bypass dialogue)
- `R` - Quick run attempt
- `ESC` - Cancel (defaults to fight)

### State Management

**Model Structure** (`encounterModel`):
```go
type encounterModel struct {
    encounterType   string
    npcShip         *models.Ship
    playerShip      *models.Ship
    dialogue        string
    demands         *Demands
    options         []*ResponseOption
    selectedIndex   int
    playerSkills    *PlayerSkills  // For skill checks
    width           int
    height          int
}

type Demands struct {
    Credits  int
    Cargo    int
    Item     string
}

type ResponseOption struct {
    Text         string
    Outcome      string  // Description of what happens
    SkillCheck   string  // "barter", "intimidate", etc.
    SuccessRate  int     // Percentage chance
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `encounterGeneratedMsg` - Encounter initialized
- `responseSelectedMsg` - Player chose option
- `encounterResolvedMsg` - Outcome determined
- `combatInitiatedMsg` - Transition to combat
- `encounterEndMsg` - Return to space

### Data Flow
1. Random encounter triggers (10% per jump)
2. Generate encounter based on system danger level
3. Display encounter screen with NPC dialogue
4. Present response options
5. Player selects response
6. Resolve outcome:
   - Peaceful: Exchange cargo/credits, end encounter
   - Combat: Transition to combat screen
   - Flee: Roll chance, success = escape, fail = combat
   - Skill check: Roll vs skill, adjust outcome
7. Update reputation if applicable
8. Return to space view

### Skill Checks

Some options require skill checks:

```go
func CheckSkill(player *Player, skillType string, difficulty int) bool {
    skill := player.Skills[skillType]
    roll := rand.Intn(100)

    bonus := 0
    if player.Reputation[NPC.Faction] > 50 {
        bonus = 10  // Reputation helps
    }

    return (skill + bonus) > (difficulty + roll)
}
```

**Skills**:
- **Barter**: Reduce demands, better trade deals
- **Intimidate**: Scare off weaker enemies
- **Persuasion**: Convince NPCs, avoid conflict
- **Engineering**: Repair derelicts, salvage better

### Encounter Generation

Encounters based on system properties:

```go
func GenerateEncounter(system *StarSystem) *Encounter {
    dangerLevel := system.DangerRating

    encounterRoll := rand.Intn(100)

    if encounterRoll < dangerLevel {
        // Hostile encounter (pirates)
        return NewPirateEncounter(dangerLevel)
    } else if encounterRoll < 50 {
        // Neutral encounter (trader)
        return NewTraderEncounter()
    } else if encounterRoll < 70 {
        // Police patrol
        return NewPoliceEncounter(system.Government)
    } else {
        // Distress or derelict
        return NewDistressEncounter()
    }
}
```

### Related Screens
- **Combat** - If fighting chosen
- **Trading** - If trader encounter
- **Space View** - Return after encounter

---

## PvP & Combat Enhanced

### PvP Screen
(`internal/tui/pvp.go`)

**Features**:
- Consensual duel system
- Faction war combat
- Combat rewards (credits, reputation)
- Spectator mode for others
- Combat history/statistics

### Combat Enhanced Screen
(`internal/tui/combat_enhanced.go`)

**Features**:
- Advanced tactical options
- Formation combat (if multi-ship)
- Targeting specific subsystems
- Special abilities (cloaking, EMP, etc.)
- Detailed damage model
- Combat replay/analysis

---

## Implementation Notes

### Combat System Architecture

```go
type CombatEngine struct {
    Combatants []*Combatant
    Turn       int
    Log        *CombatLog
}

type Combatant struct {
    Ship       *models.Ship
    AI         *CombatAI  // nil for player
    Status     *CombatStatus
}

type CombatStatus struct {
    Hull       int
    Shields    int
    Energy     int
    IsEvading  bool
    IsDefending bool
    StatusEffects []StatusEffect
}
```

### AI System

```go
type CombatAI struct {
    Difficulty AILevel
    Strategy   AIStrategy
}

type AILevel int

const (
    AILevelEasy AILevel = iota
    AILevelMedium
    AILevelHard
    AILevelVeryHard
    AILevelAce
)

func (ai *CombatAI) SelectAction(combat *CombatEngine) Action {
    switch ai.Difficulty {
    case AILevelEasy:
        return ai.RandomAction()
    case AILevelMedium:
        return ai.BasicTactics(combat)
    case AILevelHard:
        return ai.AdvancedTactics(combat)
    // ...etc
    }
}
```

### Damage Model

```go
func ApplyDamage(target *Ship, damage int, damageType DamageType) {
    if target.Shields > 0 {
        // Shields absorb first
        shieldDamage := damage
        if damageType == DamageTypeKinetic {
            shieldDamage = int(float64(damage) * 0.5)  // 50% vs shields
        }

        target.Shields -= shieldDamage
        if target.Shields < 0 {
            // Overflow to hull
            target.Hull += target.Shields
            target.Shields = 0
        }
    } else {
        // Direct hull damage
        armorReduction := target.Armor
        reducedDamage := damage * (100 - armorReduction) / 100
        target.Hull -= max(1, reducedDamage)
    }

    if target.Hull <= 0 {
        target.Destroy()
    }
}
```

### Testing

- Combat turn resolution tests
- AI behavior tests (all difficulty levels)
- Damage calculation tests
- Hit/miss calculation tests
- Weapon cooldown tests
- Loot generation tests
- Encounter generation tests
- Skill check tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
