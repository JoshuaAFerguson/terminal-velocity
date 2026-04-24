# Information & Statistics Screens

This document covers all information display and statistics screens in Terminal Velocity.

## Overview

**Screens**: 5
- News Screen
- Leaderboards Screen
- Achievements Screen
- Notifications Screen
- Player Trade (P2P Trading)

**Purpose**: Display game information, player statistics, achievements, rankings, and facilitate player-to-player trading.

**Source Files**:
- `internal/tui/news.go` - News feed and events
- `internal/tui/leaderboards.go` - Player rankings
- `internal/tui/achievements.go` - Achievement tracking
- `internal/tui/notifications.go` - Alert and notification system
- `internal/tui/trade.go` - Player-to-player trading

---

## News Screen

### Source File
`internal/tui/news.go`

### Purpose
Display dynamically generated news articles based on game events.

### News Types (10+ event types)

1. **Economic** - Market crashes, booms, trade deals
2. **Political** - Government changes, alliances, wars
3. **Military** - Battles, fleet movements, victories
4. **Pirate** - Pirate raids, bounties, defeats
5. **Exploration** - New discoveries, first contacts
6. **Technology** - Breakthroughs, new equipment
7. **Crime** - Notable crimes, arrests
8. **Faction** - Player faction achievements
9. **Player** - Major player accomplishments
10. **System** - System-specific local news

### Features
- Auto-generated based on game events
- Historical archive (last 30 days)
- Filter by news type
- Lore building through news
- Quest hooks in news articles

### Related Systems
- `internal/news/generator.go` - News article generation
- Event system triggers news
- System-specific news generation

---

## Leaderboards Screen

### Source File
`internal/tui/leaderboards.go`

### Purpose
Display player rankings across multiple categories.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ LEADERBOARDS                                                52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ [Credits ▼]  [Combat]  [Trading]  [Exploration]          Season 3    ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ TOP PILOTS BY NET WORTH                                              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃ RANK  PILOT            FACTION              NET WORTH      SHIP      ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  🥇 1  TradeKing        Merchant Guild      45,890,250 cr  Cruiser  ┃  ┃
┃  ┃  🥈 2  PirateLord       Crimson Raiders     42,156,800 cr  Battlesh ┃  ┃
┃  ┃  🥉 3  SpaceBaron       Free Traders        38,920,100 cr  Cruiser  ┃  ┃
┃  ┃     4  CreditCollector  Star Traders Guild  35,445,670 cr  Corvette ┃  ┃
┃  ┃     5  WealthSeeker     Independent         32,108,900 cr  Freighter┃  ┃
┃  ┃    ...                                                               ┃  ┃
┃  ┃   ▶ 47  SpaceCaptain    Star Traders Guild   8,234,500 cr  Corvette ┃  ┃
┃  ┃    ...                                                               ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR RANK: #47 of 247 active pilots                                 ┃  ┃
┃  ┃ Next Rank: Gain 720,000 cr to reach #46                             ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Switch Category  [F]ilter by Faction  [S]eason History  [ESC] Back  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Leaderboard Categories (4 main)

1. **Credits** - Total net worth (ship + cargo + credits)
2. **Combat** - Total kills, combat rating
3. **Trading** - Total trade volume, profit
4. **Exploration** - Systems explored, discoveries

### Features
- Season-based rankings (reset quarterly)
- All-time rankings
- Faction-specific leaderboards
- Your rank highlighted
- Gap to next rank shown
- Historical season data

### Related Systems
- `internal/leaderboards/manager.go` - Ranking calculations
- Real-time updates on significant changes
- Cached rankings (refresh every 5 minutes)

---

## Achievements Screen

### Source File
`internal/tui/achievements.go`

### Purpose
Track and display player achievements and milestones.

### Achievement Categories

- **Combat**: Kill counts, boss defeats
- **Trading**: Trade volume, profit milestones
- **Exploration**: Systems visited, discoveries
- **Social**: Faction membership, friendships
- **Story**: Quest completion, storyline progress
- **Secrets**: Hidden achievements
- **Mastery**: Skill-based achievements

### Features
- Progress bars for incremental achievements
- Locked/unlocked status
- Rarity indicators
- Completion percentage
- Recent unlocks highlighted
- Rewards for achievements (credits, titles, items)

### Achievement Types
- **One-time**: Binary completion
- **Incremental**: Progress toward goal
- **Hidden**: Requirements not shown until unlocked
- **Secret**: Not visible until discovered

---

## Notifications Screen

### Source File
`internal/tui/notifications.go`

### Purpose
Centralized notification system for game events and alerts.

### Notification Types

- **Combat**: Ship damage, combat start/end
- **Trading**: Trade complete, market alerts
- **Missions**: Mission accepted, completed, failed
- **Social**: Friend requests, faction invites, messages
- **System**: Server announcements, maintenance
- **Achievements**: Achievement unlocked
- **Economy**: Price alerts, market opportunities

### Features
- Toast notifications (temporary overlay)
- Notification center (persistent log)
- Priority levels (info, warning, critical)
- Read/unread status
- Filter by type
- Customizable notification settings
- Sound/vibration options

---

## Player Trade Screen (P2P)

### Source File
`internal/tui/trade.go`

### Purpose
Secure player-to-player trading with escrow system.

### Features

**Trade Window**:
- Split-screen: Your offer | Their offer
- Add items from inventory
- Add credits to offer
- Accept/decline trade
- Trade history

**Escrow System**:
- Both players lock in offers
- Both must accept
- Atomic transaction (all or nothing)
- No partial trades
- Prevents scamming

**Security**:
- Trade confirmation required
- Timeout after 5 minutes of inactivity
- Cancel anytime before both accept
- Trade log for disputes

### Trade Flow

1. Player A initiates trade with Player B
2. Both players add items and credits
3. Player A locks offer
4. Player B locks offer
5. Both review final trade
6. Both accept
7. Escrow executes atomic swap
8. Trade complete

---

## Implementation Notes

### News Generation

```go
type NewsGenerator struct {
    Templates map[string][]string
    Events    chan GameEvent
}

func (ng *NewsGenerator) GenerateArticle(event GameEvent) *NewsArticle {
    template := ng.SelectTemplate(event.Type)
    article := ng.PopulateTemplate(template, event.Data)
    return &NewsArticle{
        Title:     article.Title,
        Content:   article.Content,
        Category:  event.Type,
        Timestamp: time.Now(),
        System:    event.System,
    }
}
```

### Leaderboard Calculations

```go
func CalculateNetWorth(player *Player) int {
    shipValue := player.Ship.Value
    cargoValue := CalculateCargoValue(player.Ship.Cargo)
    credits := player.Credits

    return shipValue + cargoValue + credits
}

func UpdateLeaderboards() {
    players := database.GetAllActivePlayers()
    rankings := make(map[string][]*PlayerRank)

    for _, player := range players {
        rankings["credits"] = append(rankings["credits"], &PlayerRank{
            Player: player,
            Score:  CalculateNetWorth(player),
        })
        // ... other categories
    }

    // Sort and persist
    for category, ranks := range rankings {
        sort.Slice(ranks, func(i, j int) bool {
            return ranks[i].Score > ranks[j].Score
        })
        database.UpdateLeaderboard(category, ranks)
    }
}
```

### Achievement Tracking

```go
type Achievement struct {
    ID          string
    Name        string
    Description string
    Category    string
    Hidden      bool
    Progress    int
    Target      int
    Unlocked    bool
}

func (a *Achievement) CheckUnlock(player *Player) bool {
    switch a.ID {
    case "first_kill":
        return player.Stats.Kills >= 1
    case "trade_baron":
        return player.Stats.TradeVolume >= 1000000
    // ... etc
    }
}
```

### P2P Trade Escrow

```go
type TradeEscrow struct {
    PlayerA      *Player
    PlayerB      *Player
    OfferA       *TradeOffer
    OfferB       *TradeOffer
    BothLocked   bool
    BothAccepted bool
}

func (te *TradeEscrow) Execute() error {
    if !te.BothAccepted {
        return errors.New("both players must accept")
    }

    // Begin transaction
    tx := database.BeginTransaction()

    // Validate both players still have offered items
    if !te.ValidateOffers(tx) {
        tx.Rollback()
        return errors.New("invalid offers")
    }

    // Transfer items atomically
    te.TransferItems(tx, te.PlayerA, te.PlayerB, te.OfferA)
    te.TransferItems(tx, te.PlayerB, te.PlayerA, te.OfferB)

    // Commit transaction
    return tx.Commit()
}
```

### Testing
- News generation tests
- Leaderboard calculation tests
- Achievement unlock tests
- Notification delivery tests
- P2P trade escrow tests
- Trade timeout tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
