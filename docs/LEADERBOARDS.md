# Leaderboards System

**Feature**: Player Leaderboards and Rankings
**Phase**: 14
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Leaderboards system provides competitive rankings across multiple categories, tracking player achievements and performance. Players can compete globally or within their faction for top positions in various gameplay metrics.

### Key Features

- **4 Ranking Categories**: Credits, Combat, Trading, and Exploration
- **Real-Time Updates**: Rankings update automatically as players progress
- **Global and Faction Rankings**: Compete server-wide or within your faction
- **Historical Tracking**: View ranking changes over time
- **Detailed Statistics**: See breakdown of ranking factors
- **Achievement Integration**: Leaderboard positions unlock achievements
- **Weekly/Monthly Resets**: Seasonal competitive periods

---

## Architecture

### Components

The leaderboards system consists of the following components:

1. **Leaderboards Manager** (`internal/leaderboards/manager.go`)
   - Manages all ranking categories
   - Calculates player positions
   - Updates rankings efficiently
   - Thread-safe with `sync.RWMutex`

2. **Leaderboards UI** (`internal/tui/leaderboards.go`)
   - Category selection interface
   - Ranking displays
   - Player statistics view
   - Personal ranking tracker

3. **Data Models** (`internal/models/`)
   - `LeaderboardEntry`: Individual ranking entry
   - `PlayerStats`: Statistics per category
   - `RankingHistory`: Historical positions

### Data Flow

```
Player Action (Complete Trade/Combat/etc)
         ↓
Update Player Statistics
         ↓
Leaderboards Manager (Recalculate Rank)
         ↓
Update Ranking Position
         ↓
Check Achievement Thresholds
         ↓
Broadcast Rank Changes
         ↓
Update UI Display
```

### Thread Safety

The leaderboards manager uses `sync.RWMutex` for concurrent access:

- **Read Operations**: Efficient concurrent reads for rankings
- **Write Operations**: Exclusive access for updates
- **Batch Updates**: Periodic bulk recalculation
- **Cache Layer**: Frequently accessed rankings cached

---

## Implementation Details

### Leaderboards Manager

The manager handles all ranking operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Rankings by category
    creditsBoard     []*LeaderboardEntry
    combatBoard      []*LeaderboardEntry
    tradeBoard       []*LeaderboardEntry
    explorationBoard []*LeaderboardEntry

    // Player positions
    playerRanks map[uuid.UUID]*PlayerRanks

    // Configuration
    updateInterval   time.Duration
    maxEntries       int
    cacheExpiry      time.Duration

    // Repositories
    playerRepo *database.PlayerRepository
}
```

### Ranking Categories

#### 1. Credits Leaderboard

Ranks players by total credits owned.

**Calculation**:
```go
func (m *Manager) CalculateCreditsRank(playerID uuid.UUID) int {
    player := m.playerRepo.GetPlayer(playerID)
    credits := player.Credits

    // Count players with more credits
    rank := 1
    for _, entry := range m.creditsBoard {
        if entry.Value > credits {
            rank++
        }
    }
    return rank
}
```

**Statistics Tracked**:
- Current credits
- Peak credits achieved
- Credits earned (lifetime)
- Credits spent (lifetime)
- Net worth (credits + assets)

#### 2. Combat Leaderboard

Ranks players by combat prowess and victories.

**Combat Score Formula**:
```go
func (m *Manager) CalculateCombatScore(stats *CombatStats) int64 {
    baseScore := stats.Wins * 100
    baseScore -= stats.Losses * 25
    baseScore += stats.PiratesKilled * 10
    baseScore += stats.BossesDefeated * 500
    baseScore += stats.PvPWins * 200

    // Apply difficulty multiplier
    baseScore = int64(float64(baseScore) * stats.AvgDifficulty)

    return baseScore
}
```

**Statistics Tracked**:
- Total victories
- Win/loss ratio
- Pirates destroyed
- Bosses defeated
- PvP ranking
- Average combat difficulty
- Damage dealt
- Survival rate

#### 3. Trading Leaderboard

Ranks players by trading volume and profit.

**Trading Score Formula**:
```go
func (m *Manager) CalculateTradingScore(stats *TradingStats) int64 {
    profitScore := stats.TotalProfit / 100  // Scale down
    volumeScore := stats.TotalVolume / 1000

    // Efficiency bonus
    efficiency := float64(stats.TotalProfit) / float64(stats.TotalVolume)
    efficiencyBonus := int64(efficiency * 10000)

    return profitScore + volumeScore + efficiencyBonus
}
```

**Statistics Tracked**:
- Total trades completed
- Total profit earned
- Trade volume (tons)
- Highest single profit
- Trade routes discovered
- Market manipulation success
- Commodity diversity

#### 4. Exploration Leaderboard

Ranks players by exploration achievements.

**Exploration Score Formula**:
```go
func (m *Manager) CalculateExplorationScore(stats *ExplorationStats) int64 {
    score := stats.SystemsVisited * 10
    score += stats.PlanetsLanded * 5
    score += stats.JumpsExecuted
    score += stats.DistanceTraveled / 1000

    // Discovery bonuses
    score += stats.SystemsDiscovered * 100
    score += stats.AnomaliesFound * 250

    return int64(score)
}
```

**Statistics Tracked**:
- Systems visited (unique)
- Planets landed on
- Total jumps made
- Distance traveled
- First discoveries
- Anomalies found
- Rare locations visited
- Exploration efficiency

### Ranking Updates

**Update Strategy**:

1. **Immediate Updates**: Small changes (single trade, combat)
2. **Batch Updates**: Large recalculations (hourly)
3. **On-Demand**: When player views leaderboards

**Efficient Ranking Algorithm**:
```go
func (m *Manager) UpdatePlayerRank(
    playerID uuid.UUID,
    category LeaderboardCategory,
    newScore int64,
) {
    m.mu.Lock()
    defer m.mu.Unlock()

    board := m.getBoard(category)

    // Binary search for insertion position
    position := sort.Search(len(board), func(i int) bool {
        return board[i].Score < newScore
    })

    // Create new entry
    entry := &LeaderboardEntry{
        PlayerID: playerID,
        Score:    newScore,
        Rank:     position + 1,
    }

    // Insert and adjust ranks
    board = append(board[:position], append([]*LeaderboardEntry{entry}, board[position:]...)...)
    m.recalculateRanks(board)

    // Trim to max entries
    if len(board) > m.maxEntries {
        board = board[:m.maxEntries]
    }

    m.setBoard(category, board)
}
```

### Faction Rankings

Players can view faction-specific leaderboards:

**Faction Filtering**:
```go
func (m *Manager) GetFactionLeaderboard(
    factionID uuid.UUID,
    category LeaderboardCategory,
) []*LeaderboardEntry {
    globalBoard := m.getBoard(category)
    factionBoard := make([]*LeaderboardEntry, 0)

    for _, entry := range globalBoard {
        if entry.FactionID == factionID {
            factionBoard = append(factionBoard, entry)
        }
    }

    // Rerank within faction
    for i, entry := range factionBoard {
        entry.FactionRank = i + 1
    }

    return factionBoard
}
```

---

## User Interface

### Leaderboards Screen

**Main View**:
```
=== Leaderboards ===

Categories:
  > Credits      [Your Rank: #127]
    Combat       [Your Rank: #45]
    Trading      [Your Rank: #89]
    Exploration  [Your Rank: #203]

[Global] [Faction] [Personal Stats]
```

**Credits Leaderboard Display**:
```
═══════════════════════════════════════════════════════
                 CREDITS LEADERBOARD
═══════════════════════════════════════════════════════

Rank  Player               Faction          Credits
────  ──────────────────   ──────────────   ──────────
  1   ★ MasterTrader       Crimson Tide     15,432,890 CR
  2   AlphaCommander       Star Alliance    12,876,543 CR
  3   CreditKing           Independent       9,654,321 CR
...
 127  You                  Phoenix Fleet       543,210 CR
...

Your Stats:
  Current Credits: 543,210 CR
  Peak Credits:    678,500 CR
  Lifetime Earned: 2,345,678 CR

[Previous] [Next] [Details] [ESC: Back]
```

**Combat Leaderboard Display**:
```
═══════════════════════════════════════════════════════
                 COMBAT LEADERBOARD
═══════════════════════════════════════════════════════

Rank  Player            Score      Wins   W/L Ratio
────  ────────────────  ────────   ────   ─────────
  1   ★ Ace Pilot       125,430    1,234  4.2:1
  2   WarMachine        98,765       987  3.8:1
  3   DeathStar         87,234       876  3.5:1
...
  45  You               12,456       124  2.1:1
...

Your Combat Stats:
  Total Wins:      124
  Total Losses:     59
  Pirates Killed:  234
  PvP Wins:         15
  Combat Score: 12,456
```

### Navigation

- **↑/↓**: Navigate leaderboard
- **←/→**: Switch categories
- **Tab**: Toggle Global/Faction
- **Enter**: View player details
- **P**: View personal stats
- **ESC**: Return to main menu

---

## Integration with Other Systems

### Achievement Integration

Leaderboard positions unlock achievements:

```go
achievements := []Achievement{
    {Name: "Top 10%",  Trigger: RankPercent(10)},
    {Name: "Top 100",  Trigger: RankAbsolute(100)},
    {Name: "Top 10",   Trigger: RankAbsolute(10)},
    {Name: "#1 Trader", Trigger: RankFirst(CategoryTrading)},
}
```

### Chat Integration

Rank changes broadcast to chat:

```
[LEADERBOARD] Alice advanced to #5 in Trading!
[LEADERBOARD] Bob reached #1 in Combat!
[LEADERBOARD] New exploration record by Charlie!
```

### Faction Integration

Faction bonuses for top-ranked members:

- Top 10 members get faction perks
- Collective faction score from top members
- Inter-faction competition

---

## Testing

### Unit Tests

```go
func TestLeaderboard_RankCalculation(t *testing.T)
func TestLeaderboard_UpdateRank(t *testing.T)
func TestLeaderboard_FactionFilter(t *testing.T)
func TestLeaderboard_ThreadSafety(t *testing.T)
func TestLeaderboard_Performance(t *testing.T)
```

### Performance Tests

- **Large User Base**: 10,000+ players
- **Concurrent Updates**: 100+ simultaneous updates
- **Query Performance**: Sub-100ms response time
- **Memory Usage**: Efficient caching

---

## Configuration

```go
cfg := &leaderboards.Config{
    // Update settings
    UpdateInterval:    5 * time.Minute,
    MaxEntries:        1000,

    // Caching
    CacheExpiry:       1 * time.Minute,
    CacheSize:         10000,

    // Categories enabled
    EnableCredits:     true,
    EnableCombat:      true,
    EnableTrading:     true,
    EnableExploration: true,

    // Reset periods
    WeeklyReset:       true,
    MonthlyReset:      true,
}
```

---

## API Reference

### Core Functions

#### GetLeaderboard

```go
func (m *Manager) GetLeaderboard(
    category LeaderboardCategory,
    limit int,
    offset int,
) []*LeaderboardEntry
```

Returns leaderboard entries for a category.

#### GetPlayerRank

```go
func (m *Manager) GetPlayerRank(
    playerID uuid.UUID,
    category LeaderboardCategory,
) (*PlayerRank, error)
```

Returns player's current rank in a category.

#### UpdatePlayerStats

```go
func (m *Manager) UpdatePlayerStats(
    playerID uuid.UUID,
    category LeaderboardCategory,
    delta int64,
) error
```

Updates player statistics and recalculates rank.

---

## Related Documentation

- [Achievements System](./ACHIEVEMENTS.md) - Achievement integration
- [PvP Combat](./PVP_COMBAT.md) - Combat rankings
- [Player Factions](./PLAYER_FACTIONS.md) - Faction leaderboards

---

## File Locations

**Core Implementation**:
- `internal/leaderboards/manager.go` - Leaderboards manager
- `internal/models/leaderboard.go` - Data models

**User Interface**:
- `internal/tui/leaderboards.go` - Leaderboards UI

**Documentation**:
- `docs/LEADERBOARDS.md` - This file
- `ROADMAP.md` - Phase 14 details

---

**For questions about the leaderboards system, see the troubleshooting section or contact the development team.**
