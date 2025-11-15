# Social & Multiplayer Screens

This document covers all social and multiplayer interaction screens in Terminal Velocity.

## Overview

**Screens**: 6
- Chat Screen
- Players Screen
- Friends Screen
- Factions Screen
- Mail Screen
- Territory (integrated with Factions)

**Purpose**: Handle multiplayer interactions including chat, player tracking, friendships, faction management, messaging, and territory control.

**Source Files**:
- `internal/tui/chat.go` - Multi-channel chat system
- `internal/tui/players.go` - Online players list
- `internal/tui/friends.go` - Friend list management
- `internal/tui/factions.go` - Player faction management
- `internal/tui/mail.go` - Asynchronous messaging system

---

## Chat Screen

### Source File
`internal/tui/chat.go`

### Purpose
Real-time multi-channel chat for player communication.

### Channels (4 types)

1. **Global** - All players server-wide
2. **System** - Players in current star system
3. **Faction** - Your faction members only
4. **Direct Message** - Private 1-on-1 chat

### Key Features
- Real-time message delivery
- Channel switching
- Message history (last 100 messages)
- Player mentions (@username)
- Emote support
- Muted player filtering
- Timestamps optional

### Related Systems
- `internal/chat/manager.go` - Chat message routing
- Message moderation for muted players
- Chat log persistence

---

## Players Screen

### Source File
`internal/tui/players.go`

### Purpose
View online players, their locations, and status.

### Features
- List all online players
- Current system/planet location
- Ship type display
- Online status indicator
- Quick actions: Message, Trade, Challenge, Add Friend

### Presence System
- Players marked online when connected
- Location updates every 30 seconds
- Timeout after 5 minutes of inactivity
- System-specific player lists

---

## Friends Screen

### Source File
`internal/tui/friends.go`

### Purpose
Manage friend list, see friend status and locations.

### Features
- Add/remove friends
- Friend request system
- Online/offline status
- Last seen timestamp
- Quick travel to friend's location
- Friend-only chat channel

---

## Factions Screen

### Source File
`internal/tui/factions.go`

### Purpose
Player faction management including members, treasury, territory, and wars.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ FACTION MANAGEMENT                                          52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR FACTION: Star Traders Guild                           [Leader]  ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  "We trade in the stars, and the stars trade with us."              ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Founded: 23 days ago                Members: 47                    ┃  ┃
┃  ┃  Faction Level: 12                   Active Members: 18             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Treasury: 2,450,000 credits         Territory: 3 systems           ┃  ┃
┃  ┃  Passive Income: +12,000 cr/day      Tax Rate: 5%                   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ MEMBERS (Online: 18/47)          ┃  ┃ FACTION STATS                  ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                  ┃  ┃                                ┃  ┃
┃  ┃ ▶ SpaceCaptain    [Leader] ●    ┃  ┃ Total Kills: 2,340             ┃  ┃
┃  ┃   TraderJoe       [Officer] ●   ┃  ┃ Trade Volume: 45M cr           ┃  ┃
┃  ┃   CargoPilot      [Member] ●    ┃  ┃ Systems Explored: 89           ┃  ┃
┃  ┃   SpaceMerc       [Member] ●    ┃  ┃ Faction Wars Won: 12           ┃  ┃
┃  ┃   MiningBoss      [Member] ○    ┃  ┃                                ┃  ┃
┃  ┃   QuickShip       [Recruit] ●   ┃  ┃ Alliances: 2                   ┃  ┃
┃  ┃   ... (41 more)                 ┃  ┃ Enemies: 1                     ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ CONTROLLED TERRITORY                                                 ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃  ⊙ Sirius System         Income: +5,000 cr/day    Defense: ████░░   ┃  ┃
┃  ┃  ⊙ Barnard's Star        Income: +4,200 cr/day    Defense: ██████   ┃  ┃
┃  ┃  ⊙ Wolf 359              Income: +2,800 cr/day    Defense: ███░░░   ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [I]nvite  [K]ick  [P]romote  [D]emote  [T]erritory  [W]ar  [ESC] Back      ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Faction Features

**Hierarchy**:
- Leader (1) - Full permissions
- Officer (unlimited) - Most permissions
- Member (unlimited) - Basic permissions
- Recruit (unlimited) - Limited permissions

**Treasury**:
- Shared faction bank
- Deposits from members
- Withdrawals (leader/officer only)
- Automatic tax collection option

**Territory**:
- Claim systems for passive income
- Defense rating (resist takeover)
- Income per day per system
- Territory wars with other factions

### Related Systems
- `internal/factions/manager.go` - Faction operations
- `internal/territory/` - Territory control
- `internal/pvp/` - Faction wars

---

## Mail Screen

### Source File
`internal/tui/mail.go`

### Purpose
Asynchronous messaging system for player-to-player communication.

### Features
- Inbox/sent mail folders
- Compose new messages
- Reply/forward
- Attachments (credits, items)
- Read receipts
- Message expiration (30 days)
- Bulk delete

### Use Cases
- Offline player communication
- Trade negotiations
- Mission coordination
- Alliance discussions

---

## Implementation Notes

### Thread Safety
All social features use `sync.RWMutex` for concurrent access:
- Chat message broadcasting
- Player presence updates
- Faction operations
- Mail delivery

### Real-Time Updates
- Chat: WebSocket-style push messages
- Presence: 30-second heartbeat
- Faction events: Broadcast to all members

### Data Persistence
- Chat: Last 100 messages per channel
- Presence: Redis-like in-memory cache
- Mail: Full persistence in database
- Factions: Complete state in database

### Testing
- Chat message routing tests
- Presence timeout tests
- Faction permission tests
- Mail delivery tests
- Concurrent access tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
