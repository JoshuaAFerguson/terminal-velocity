# PvP Combat System

**Feature**: Player vs Player Combat
**Phase**: 13
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The PvP (Player vs Player) Combat system enables consensual and faction-based combat between players. Built on the existing combat engine, it adds multiplayer-specific mechanics like duel challenges, faction wars, and competitive rewards.

### Key Features

- **Consensual Duels**: Challenge other players to one-on-one combat
- **Faction Warfare**: Faction vs faction battles for territory and honor
- **Duel Challenges**: Send and accept combat requests
- **Combat Rewards**: Winners receive credits, reputation, and loot
- **Loss Penalties**: Balanced penalties to prevent griefing
- **Combat History**: Track your PvP record and statistics
- **Ranking System**: Competitive leaderboards for PvP performance
- **Anti-Griefing**: Protection mechanisms for new/weaker players

---

## Architecture

### Components

The PvP system consists of the following components:

1. **PvP Manager** (`internal/pvp/manager.go`)
   - Manages duel challenges and matchmaking
   - Enforces PvP rules and consent
   - Handles combat initiation and resolution
   - Tracks combat statistics
   - Thread-safe with `sync.RWMutex`

2. **PvP UI** (`internal/tui/pvp.go`)
   - Challenge interface
   - Active duels display
   - Combat history viewer
   - Ranking display

3. **Combat Integration** (`internal/combat/`)
   - Reuses existing combat engine
   - PvP-specific modifiers
   - Reward calculation

### Data Flow

```
Player A Challenges Player B
         ↓
Create Duel Challenge
         ↓
Send Notification to Player B
         ↓
Player B Accepts/Declines
         ↓
[If Accepted]
Initialize Combat Instance
         ↓
Run Combat Turns
         ↓
Determine Winner
         ↓
Apply Rewards/Penalties
         ↓
Update PvP Statistics
         ↓
Update Leaderboards
```

### Thread Safety

The PvP manager uses `sync.RWMutex` for concurrent operations:

- **Read Operations**: Multiple concurrent reads for challenge lists
- **Write Operations**: Exclusive access for challenge creation/acceptance
- **Combat State**: Locked during combat resolution
- **Stat Updates**: Atomic updates for rankings

---

## Implementation Details

### PvP Manager

The manager handles all PvP operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Active challenges
    challenges    map[uuid.UUID]*models.DuelChallenge // ChallengeID -> Challenge
    activeDuels   map[uuid.UUID]*models.PvPCombat     // DuelID -> Combat

    // Player statistics
    playerStats   map[uuid.UUID]*models.PvPStats      // PlayerID -> Stats

    // Configuration
    challengeTimeout time.Duration
    cooldownPeriod   time.Duration

    // Managers
    combatManager  *combat.Manager
    factionManager *factions.Manager
    reputationMgr  *reputation.Manager
}
```

### Duel Challenges

**Challenge Types**:

1. **Friendly Duel**: No stakes, practice combat
2. **Ranked Duel**: Affects rankings and statistics
3. **Wager Duel**: Winner takes agreed-upon stakes
4. **Faction Duel**: Represents faction honor

**Challenge Structure**:
```go
type DuelChallenge struct {
    ChallengeID uuid.UUID
    Challenger  uuid.UUID
    Challenged  uuid.UUID

    ChallengeType DuelType
    Stakes        int64    // Credits wagered
    FactionWar    bool     // Faction battle flag

    Status        ChallengeStatus
    CreatedAt     time.Time
    ExpiresAt     time.Time
    AcceptedAt    *time.Time
}
```

### Combat Mechanics

PvP combat uses the same turn-based system as PvE:

**Combat Initialization**:
```go
func (m *Manager) InitiateDuel(
    challengeID uuid.UUID,
) (*models.PvPCombat, error) {
    challenge := m.challenges[challengeID]

    // Create combat instance
    combat := combat.NewCombat(
        challenge.Challenger,
        challenge.Challenged,
        combat.ModePvP,
    )

    // Apply PvP modifiers
    combat.EnablePvPRules()
    combat.SetStakes(challenge.Stakes)

    // Start combat
    return combat, nil
}
```

**PvP-Specific Rules**:
- No fleeing allowed in ranked duels
- Both players start with full shields/hull
- Turn time limits to prevent stalling
- Spectator mode for faction members
- Automatic tie-breaker after 50 turns

### Rewards and Penalties

**Victory Rewards**:
```go
func (m *Manager) CalculateRewards(
    winner uuid.UUID,
    loser uuid.UUID,
    duelType DuelType,
) *PvPRewards {
    baseCredits := 1000
    reputationGain := 10

    if duelType == DuelRanked {
        // Ranked rewards scale with opponent rating
        loserRating := m.playerStats[loser].Rating
        multiplier := 1.0 + (float64(loserRating) / 1000.0)

        baseCredits = int64(float64(baseCredits) * multiplier)
        reputationGain = int(float64(reputationGain) * multiplier)
    }

    return &PvPRewards{
        Credits:    baseCredits,
        Reputation: reputationGain,
        RatingGain: calculateRatingChange(winner, loser),
    }
}
```

**Defeat Penalties**:
- Small credit loss (10% of base reward)
- Reputation impact (faction warfare)
- Rating decrease (ranked duels)
- No ship/equipment loss (prevents griefing)

### Faction Warfare

**Territory Battles**:
```go
type FactionBattle struct {
    AttackingFaction uuid.UUID
    DefendingFaction uuid.UUID
    Territory        uuid.UUID

    Attackers []uuid.UUID
    Defenders []uuid.UUID

    BattleStatus FactionBattleStatus
    Winner       *uuid.UUID
}
```

**Faction War Rules**:
- Requires minimum faction sizes (5+ members)
- Territory battles are multi-player
- Faction-wide rewards/penalties
- Territory control at stake
- Extended battle duration

### Anti-Griefing Measures

**Protection Mechanisms**:

1. **Level Protection**:
   ```go
   maxLevelDifference := 10
   if abs(p1.Level - p2.Level) > maxLevelDifference {
       return ErrLevelGapTooLarge
   }
   ```

2. **New Player Protection**:
   - Players under 7 days old immune to challenges
   - Must opt-in to PvP after tutorial
   - Grace period after first PvP enable

3. **Cooldowns**:
   - 1 hour cooldown between duels with same player
   - 15 minute cooldown after any duel
   - 24 hour cooldown after faction war

4. **Consent Requirement**:
   - All duels require acceptance
   - Can decline without penalty
   - Settings to disable challenges

---

## User Interface

### PvP Screen

**Main View**:
```
=== PvP Combat ===

Your PvP Stats:
  Wins: 15      Losses: 3      Rating: 1,245
  Win Rate: 83.3%
  Rank: #127 of 1,524 players

Incoming Challenges (2)        Sent Challenges (1)
─────────────────────         ──────────────────
> Duel from Alice             Waiting: Bob
  Faction War: Crimson Tide

[View] [Accept] [Decline]

Recent Duels | Rankings | Settings
```

**Challenge Screen**:
```
=== Challenge Player ===

Opponent: Bob
Ship: Corvette (Combat: 450)

Challenge Type:
  ( ) Friendly Duel - No stakes
  (•) Ranked Duel - Affects rating
  ( ) Wager Duel - Credits: [_____]

Your Ship: Destroyer (Combat: 520)
Estimated Win Chance: 65%

[Send Challenge] [Cancel]
```

**Active Duel Display**:
```
═══════════════════════════════════════════
            PvP DUEL IN PROGRESS
═══════════════════════════════════════════

YOU (Destroyer)              ALICE (Corvette)
Hull:    ████████░░ 80%      Hull:    ██████░░░░ 60%
Shields: ██████████ 100%     Shields: ████░░░░░░ 40%

Turn 5 of 50                 Stakes: 5,000 CR

Your Turn:
1. Fire Laser Cannon (95% accuracy)
2. Fire Missile (75% accuracy)
3. Use Shield Boost
4. Scan Enemy

[Select Action]
```

### Navigation

- **↑/↓**: Navigate challenge list
- **Enter**: View/Accept challenge
- **C**: Challenge player
- **D**: Decline challenge
- **H**: View duel history
- **R**: View rankings
- **ESC**: Return to main menu

---

## Integration with Other Systems

### Combat System Integration

PvP reuses the core combat engine:

**Combat Modes**:
```go
const (
    ModePvE     = "pve"     // Player vs NPC
    ModePvP     = "pvp"     // Player vs Player
    ModeFaction = "faction" // Faction warfare
)
```

**Shared Mechanics**:
- Turn-based combat
- Weapon systems
- Shield mechanics
- Damage calculations
- Loot system (modified for PvP)

### Faction Integration

PvP affects faction standings:

**Faction Benefits**:
- Defend faction territories
- Earn faction reputation
- Participate in faction wars
- Unlock faction-specific rewards

**Faction Penalties**:
- Losing territory battles
- Reputation loss with rivals
- Faction-wide cooldowns

### Leaderboard Integration

PvP statistics feed into leaderboards:

**PvP Rankings**:
- Combat rating
- Total wins/losses
- Win streaks
- Faction war performance
- Monthly tournaments

### Settings Integration

Privacy and PvP preferences:

```go
Settings.Privacy:
  AllowPvPChallenges: true/false
  FactionWarOptIn:    true/false
  MaxChallengesPerDay: 10
```

---

## Testing

### Unit Tests

```go
func TestPvP_CreateChallenge(t *testing.T)
func TestPvP_AcceptChallenge(t *testing.T)
func TestPvP_DeclineChallenge(t *testing.T)
func TestPvP_CombatResolution(t *testing.T)
func TestPvP_RewardCalculation(t *testing.T)
func TestPvP_AntiGriefing(t *testing.T)
func TestPvP_FactionWarfare(t *testing.T)
func TestPvP_Cooldowns(t *testing.T)
```

### Integration Tests

1. **Simple Duel**:
   - Challenge player
   - Accept challenge
   - Complete combat
   - Verify rewards

2. **Faction War**:
   - Start faction battle
   - Multiple participants
   - Resolve battle
   - Update territory

3. **Anti-Griefing**:
   - Test level protection
   - Cooldown enforcement
   - New player immunity
   - Consent requirement

---

## Configuration

```go
cfg := &pvp.Config{
    // Challenge settings
    ChallengeTimeout:     15 * time.Minute,
    MaxChallengesPerDay:  20,

    // Combat settings
    CombatTurnTimeout:    30 * time.Second,
    MaxCombatTurns:       50,

    // Cooldowns
    DuelCooldown:         15 * time.Minute,
    SamePlayerCooldown:   1 * time.Hour,
    FactionWarCooldown:   24 * time.Hour,

    // Protection
    NewPlayerProtection:  7 * 24 * time.Hour,
    MaxLevelDifference:   10,

    // Rewards
    BaseRewardCredits:    1000,
    BaseReputationGain:   10,
    RatingK:              32,  // ELO K-factor
}
```

---

## Troubleshooting

### Common Issues

**Problem**: Cannot challenge player
**Solutions**:
- Check if player is online
- Verify player allows PvP
- Check cooldown status
- Ensure level difference within limits

**Problem**: Challenge keeps timing out
**Solutions**:
- Reduce timeout if too long
- Ensure notifications working
- Check player activity status

**Problem**: Combat won't start
**Solutions**:
- Verify both players accepted
- Check for system errors
- Review combat manager status

---

## Future Enhancements

1. **Tournament System**
   - Scheduled tournaments
   - Bracket-based elimination
   - Prize pools
   - Spectator mode

2. **Arena Battles**
   - Dedicated PvP zones
   - Special combat rules
   - Environmental hazards
   - Team battles

3. **Rewards Enhancement**
   - Unique PvP equipment
   - Titles and badges
   - Seasonal rewards
   - Championship trophies

4. **Advanced Mechanics**
   - Team battles (2v2, 3v3)
   - Ship classes for balance
   - Loadout restrictions
   - Combat modifiers

---

## API Reference

### Core Functions

#### CreateDuelChallenge

```go
func (m *Manager) CreateDuelChallenge(
    challengerID uuid.UUID,
    challengedID uuid.UUID,
    duelType DuelType,
    stakes int64,
) (uuid.UUID, error)
```

Creates a new duel challenge.

#### AcceptChallenge

```go
func (m *Manager) AcceptChallenge(
    challengeID uuid.UUID,
    playerID uuid.UUID,
) error
```

Accepts a pending duel challenge.

#### StartDuel

```go
func (m *Manager) StartDuel(
    challengeID uuid.UUID,
) (*models.PvPCombat, error)
```

Initiates combat for an accepted challenge.

#### ResolveDuel

```go
func (m *Manager) ResolveDuel(
    duelID uuid.UUID,
    winnerID uuid.UUID,
) (*PvPRewards, error)
```

Resolves completed duel and distributes rewards.

#### GetPvPStats

```go
func (m *Manager) GetPvPStats(
    playerID uuid.UUID,
) *models.PvPStats
```

Returns PvP statistics for a player.

---

## Related Documentation

- [Combat System](./COMBAT.md) - Core combat mechanics
- [Player Factions](./PLAYER_FACTIONS.md) - Faction warfare
- [Leaderboards](./LEADERBOARDS.md) - Rankings
- [Settings System](./SETTINGS_SYSTEM.md) - PvP preferences

---

## File Locations

**Core Implementation**:
- `internal/pvp/manager.go` - PvP manager
- `internal/models/pvp.go` - PvP data models

**User Interface**:
- `internal/tui/pvp.go` - PvP UI screens

**Tests**:
- `internal/pvp/manager_test.go` - Unit tests

**Documentation**:
- `docs/PVP_COMBAT.md` - This file
- `ROADMAP.md` - Phase 13 details

---

**For questions or issues with the PvP system, see the troubleshooting section above or contact the development team.**
