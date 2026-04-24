# Chat System Documentation

**Version**: 1.0.0
**Phase**: 9 - Multiplayer Chat System
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

## Overview

The Chat System provides real-time multiplayer communication across multiple channels, enabling players to interact globally, locally, and privately. It supports five distinct chat channels with message routing, history management, and command processing.

### Key Features

- **5 Chat Channels**: Global, System (local), Faction, Direct Messages, Trade
- **Real-time Messaging**: Instant message delivery to online players
- **Message History**: Persistent per-player chat histories (200 global messages, unlimited per-player)
- **Direct Messaging**: Private player-to-player communication
- **Chat Commands**: Built-in commands for enhanced functionality
- **Message Sanitization**: ANSI escape code stripping prevents injection attacks
- **Thread-safe**: Concurrent access protection with RWMutex
- **Channel Management**: Clear, scroll, and navigate chat channels

## Architecture

### Components

```
Chat System
├── Manager (internal/chat/manager.go)
│   ├── Message routing and distribution
│   ├── Per-player history management
│   ├── Global message buffer (200 messages)
│   └── Thread-safe operations
│
├── UI (internal/tui/chat.go)
│   ├── Channel navigation (tabs)
│   ├── Message input with sanitization
│   ├── Scrollable message display
│   └── Command processing
│
└── Models (internal/models/chat.go)
    ├── ChatMessage
    ├── ChatHistory
    └── ChatChannel enum
```

### Data Flow

```
Player Input → TUI Chat Screen → Chat Manager → Routing Logic
                                      ↓
                        ┌─────────────┼─────────────┐
                        ↓             ↓             ↓
                   Global Chat   System Chat   Faction Chat
                        ↓             ↓             ↓
                All Players    System Players  Faction Members
                        ↓             ↓             ↓
                Player Histories → Message Display → TUI
```

### Message Routing

| Channel | Scope | Routing Logic |
|---------|-------|---------------|
| **Global** | All players | Broadcast to all online player histories |
| **System** | Current system | Send to players in same star system |
| **Faction** | Faction members | Send to all members of player's faction |
| **Direct** | 1-to-1 | Send to specific recipient's history |
| **Trade** | All players | Broadcast to all online player histories |
| **Combat** | Combatants | Internal combat notification system |

## Chat Channels

### 1. Global Chat (`ChatChannelGlobal`)

**Purpose**: Server-wide communication for all players

**Features**:
- Visible to all online players
- 200-message global history buffer
- Most recent messages loaded on connect
- Ideal for announcements, questions, community interaction

**Usage**:
```
[Global] Username: Hello everyone!
[Global] OtherPlayer: Welcome to Terminal Velocity!
```

### 2. System Chat (`ChatChannelSystem`)

**Purpose**: Local communication within current star system

**Features**:
- Only visible to players in same system
- Automatically filtered by system location
- Useful for coordinating local activities
- Empty if no other players in system

**Routing**:
```go
// Players in same system receive message
systemPlayers := presenceManager.GetPlayersInSystem(player.CurrentSystem)
chatManager.SendSystemMessage(systemID, senderID, sender, content, recipientIDs)
```

**Usage**:
```
[System: Sol] Trader1: Anyone selling fuel here?
[System: Sol] Trader2: Docked at Earth, I have extra
```

### 3. Faction Chat (`ChatChannelFaction`)

**Purpose**: Private communication for faction members

**Features**:
- Restricted to faction members only
- Requires player to be in a faction
- Separate history per faction
- Useful for strategy, coordination, announcements

**Access Control**:
- Player must be member of faction
- Automatically filtered by faction membership
- Error message if not in faction

**Usage**:
```
[Faction: Galactic Traders] Leader: Meeting at the station tonight!
[Faction: Galactic Traders] Member: I'll be there
```

### 4. Direct Messages (`ChatChannelDirect`)

**Purpose**: Private 1-to-1 player communication

**Features**:
- Private conversations between two players
- Multiple concurrent DM conversations
- Persistent message history per conversation
- Works even if recipient is offline (mail-like)
- Conversation list with last message preview

**Commands**:
- `/dm <username> <message>` - Send direct message
- Switch to DM channel and select conversation
- Type message directly in DM mode

**UI Flow**:
```
1. Press '4' to switch to DM channel
2. View list of active DM conversations
3. Select conversation or use /dm command
4. Type and send messages
```

**Usage**:
```
[DM] To PlayerName: Want to trade some cargo?
[DM] From PlayerName: Sure, meet me at Earth
```

### 5. Trade Chat (`ChatChannelTrade`)

**Purpose**: Advertising trades, buying, selling commodities

**Features**:
- Broadcast to all online players
- Dedicated channel for commerce
- Reduces noise in global chat
- Supports trade advertisements and negotiations

**Usage**:
```
[Trade] Merchant: WTS 50 tons Luxury Goods at Sol
[Trade] Buyer: WTB Fuel, any amount, good price
```

### 6. Combat Channel (`ChatChannelCombat`) - Internal

**Purpose**: Combat notifications and logs (not user-accessible)

**Features**:
- Automatic combat notifications
- Damage reports, hits, misses
- Combat results and loot
- Displayed in combat UI only

## Chat Commands

### Built-in Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `/help` | `/help` | Display help message with command list |
| `/dm` | `/dm <username> <message>` | Send direct message to player |
| `/clear` | `/clear` | Clear current channel messages |
| `/me` | `/me <action>` | Send action message (e.g., "* Username waves") |

### Command Examples

```
/help
Output: Chat Commands:
        /help - Show this help message
        /dm <username> <message> - Send a direct message
        /clear - Clear current channel
        /me <action> - Send an action message

/dm Trader Hello, interested in your cargo?
Output: [DM] To Trader: Hello, interested in your cargo?
        DM sent to Trader

/me waves at everyone
Output: [Global] * Username waves at everyone

/clear
Output: (Channel messages cleared)
```

## User Interface

### Channel Navigation

```
┌─────────────────────────────────────────────────────────────────┐
│                    🗨️  CHAT - Global                            │
│                                                                  │
│ Channels: [Global]  System  Faction  DMs  Trade                │
│─────────────────────────────────────────────────────────────────│
│                                                                  │
│ [Global] Player1: Welcome to the server!                        │
│ [Global] Player2: Thanks! This game is awesome                  │
│ [Global] Player3: Anyone want to trade?                         │
│ ...                                                              │
│                                                                  │
│─────────────────────────────────────────────────────────────────│
│ Press I or Enter to send a message                              │
│ I/Enter: Message | 1-5: Channels | C: Clear | ↑/↓: Scroll | ESC │
└─────────────────────────────────────────────────────────────────┘
```

### Input Mode

```
┌─────────────────────────────────────────────────────────────────┐
│                    🗨️  CHAT - Global                            │
│                                                                  │
│ (Channel tabs and messages...)                                  │
│                                                                  │
│─────────────────────────────────────────────────────────────────│
│ Message: Hello everyone!█                                        │
│ Enter: Send | ESC: Cancel                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Keyboard Controls

**Navigation Mode** (default):
- `1` - Switch to Global channel
- `2` - Switch to System channel
- `3` - Switch to Faction channel
- `4` - Switch to Direct Messages
- `5` - Switch to Trade channel
- `↑/k` - Scroll up through messages
- `↓/j` - Scroll down through messages
- `I` or `Enter` - Start typing message (enter input mode)
- `C` - Clear current channel
- `ESC/Q` - Return to main menu

**Input Mode** (typing message):
- `Enter` - Send message
- `ESC` - Cancel message, return to navigation mode
- `Backspace` - Delete last character
- `Any character` - Add to message (max 200 characters)

## Implementation Details

### Chat Manager (`internal/chat/manager.go`)

**Core Structure**:
```go
type Manager struct {
    mu               sync.RWMutex
    histories        map[uuid.UUID]*models.ChatHistory  // Per-player histories
    globalHistory    []*models.ChatMessage              // Global message buffer
    maxGlobalHistory int                                // Max global messages (200)
}
```

**Key Methods**:

```go
// Global message broadcast
SendGlobalMessage(senderID uuid.UUID, sender string, content string) *ChatMessage

// System-local messaging
SendSystemMessage(systemID, senderID uuid.UUID, sender, content string, recipientIDs []uuid.UUID) *ChatMessage

// Faction messaging
SendFactionMessage(factionID string, senderID uuid.UUID, sender, content string, memberIDs []uuid.UUID) *ChatMessage

// Direct messaging
SendDirectMessage(senderID uuid.UUID, sender string, recipientID uuid.UUID, recipient, content string) *ChatMessage

// Retrieve messages
GetMessages(playerID uuid.UUID, channel ChatChannel, limit int) []*ChatMessage
GetDirectMessages(playerID uuid.UUID, otherPlayer string, limit int) []*ChatMessage

// History management
GetOrCreateHistory(playerID uuid.UUID) *ChatHistory
RemovePlayerHistory(playerID uuid.UUID)  // On disconnect
```

### Message Models (`internal/models/chat.go`)

**ChatMessage**:
```go
type ChatMessage struct {
    ID        uuid.UUID     `json:"id"`
    Channel   ChatChannel   `json:"channel"`
    SenderID  uuid.UUID     `json:"sender_id"`
    Sender    string        `json:"sender"`
    Content   string        `json:"content"`
    Recipient string        `json:"recipient,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
    IsSystem  bool          `json:"is_system"`
    SystemID  uuid.UUID     `json:"system_id,omitempty"`
    FactionID string        `json:"faction_id,omitempty"`
}
```

**ChatHistory**:
```go
type ChatHistory struct {
    PlayerID       uuid.UUID
    messagesByChannel map[ChatChannel][]*ChatMessage
    directMessages map[string][]*ChatMessage  // Key: other player username
    mu             sync.RWMutex
}

// Methods
AddMessage(msg *ChatMessage)
GetMessages(channel ChatChannel, limit int) []*ChatMessage
GetDirectMessages(otherPlayer string, limit int) []*ChatMessage
GetActiveDirectChats() []string
ClearChannel(channel ChatChannel)
```

### Security Features

**ANSI Escape Code Stripping**:
```go
// Prevents terminal injection attacks
m.chatModel.inputBuffer = validation.StripANSI(m.chatModel.inputBuffer)
```

**Input Validation**:
- Maximum message length: 200 characters
- Control character filtering (except ESC for ANSI)
- Command validation and sanitization
- Username validation for DMs

**Thread Safety**:
- All manager operations protected by RWMutex
- Read locks for retrieving messages
- Write locks for sending/modifying messages
- No race conditions in concurrent access

## Integration with Other Systems

### Player Presence System
```go
// Get players in same system for system chat
systemPlayers := presenceManager.GetPlayersInSystem(player.CurrentSystem)
```

### Faction System
```go
// Get player's faction for faction chat
faction, err := factionManager.GetPlayerFaction(playerID)
if err == nil && faction != nil {
    chatManager.SendFactionMessage(faction.ID.String(), playerID, username, content, faction.Members)
}
```

### Player Repository
```go
// Look up recipient for direct messages
recipient, err := playerRepo.GetByUsername(context.Background(), username)
```

### Combat System
```go
// Send combat notifications
chatManager.SendCombatNotification([]uuid.UUID{player1ID, player2ID}, "Combat begins!")
```

## Testing

### Manual Testing Checklist

**Global Chat**:
- [ ] Send global message from Player 1
- [ ] Verify Player 2 receives message
- [ ] Check 200-message history limit
- [ ] Verify new player sees recent global history

**System Chat**:
- [ ] Player 1 and 2 in same system, send system message
- [ ] Verify both receive message
- [ ] Player 3 in different system, verify does NOT receive
- [ ] Check empty system chat message displays

**Faction Chat**:
- [ ] Create faction with 2+ members
- [ ] Send faction message from member
- [ ] Verify all faction members receive
- [ ] Non-member tries to send, verify error message
- [ ] Player not in faction, verify error message

**Direct Messages**:
- [ ] Send `/dm Username Message` from Player 1
- [ ] Verify Player 2 receives DM
- [ ] Check DM conversation list updates
- [ ] Send DM to offline player (mail functionality)
- [ ] Try DM to non-existent username, verify error

**Trade Chat**:
- [ ] Send trade advertisement
- [ ] Verify all players receive
- [ ] Check message format and channel indicator

**Chat Commands**:
- [ ] Test `/help` command
- [ ] Test `/dm` command with valid/invalid usernames
- [ ] Test `/clear` command
- [ ] Test `/me` action command
- [ ] Test unknown command, verify error message

**UI/UX**:
- [ ] Navigate channels with 1-5 keys
- [ ] Scroll messages with ↑/↓
- [ ] Enter input mode with I/Enter
- [ ] Send message with Enter in input mode
- [ ] Cancel message with ESC
- [ ] Test message character limit (200)
- [ ] Test backspace in input mode

**Security**:
- [ ] Try sending ANSI escape codes, verify stripped
- [ ] Test control character filtering
- [ ] Test message length limit enforcement
- [ ] Verify XSS prevention (terminal context)

### Automated Tests

Location: `internal/tui/chat_test.go`, `internal/chat/manager_test.go`

```bash
# Run chat system tests
go test ./internal/chat -v
go test ./internal/tui -run TestChat -v

# Run with race detector
go test -race ./internal/chat
```

## Performance Considerations

### Message History Limits

- **Global History**: 200 messages (trimmed on overflow)
- **Per-Player History**: Unlimited (cleaned on disconnect)
- **Message Retention**: In-memory only (not persisted to database)

### Memory Usage

- Per-player overhead: ~1-5 KB (empty history)
- Per-message overhead: ~200-500 bytes
- 100 players with 200 messages each: ~4-10 MB

### Optimization Strategies

**Memory Management**:
- History cleanup on player disconnect
- Global history ring buffer (200 messages)
- No database persistence (reduces I/O)

**Thread Safety**:
- RWMutex allows concurrent reads
- Write operations minimally locked
- Per-channel message storage reduces contention

**Message Distribution**:
- O(n) broadcast for global/trade channels (n = active players)
- O(1) direct messages
- O(m) faction messages (m = faction members)

## Configuration

### Manager Configuration

```go
// Default configuration in NewManager()
maxGlobalHistory: 200  // Keep last 200 global messages
```

### Customization

To change global message history limit:

```go
// In internal/chat/manager.go
func NewManager() *Manager {
    return &Manager{
        histories:        make(map[uuid.UUID]*models.ChatHistory),
        globalHistory:    []*models.ChatMessage{},
        maxGlobalHistory: 500, // Change to desired limit
    }
}
```

To add message persistence:

```go
// Add repository to manager
type Manager struct {
    // ... existing fields
    messageRepo *database.MessageRepository
}

// Save messages to database
func (m *Manager) SendGlobalMessage(...) *ChatMessage {
    msg := models.NewChatMessage(...)
    m.messageRepo.Create(context.Background(), msg) // Persist
    // ... existing broadcast logic
}
```

## Troubleshooting

### Common Issues

**Issue**: Messages not appearing in chat

**Causes**:
- Player histories not initialized
- Channel mismatch (viewing wrong channel)
- Player not in faction (for faction chat)
- Not in same system (for system chat)

**Solutions**:
- Verify `chatManager.GetOrCreateHistory()` called on login
- Check current channel matches message channel
- Verify faction membership for faction chat
- Ensure players in same system for system chat

---

**Issue**: Direct messages not working

**Causes**:
- Recipient username incorrect
- Recipient player doesn't exist
- Player history not found

**Solutions**:
- Verify username spelling (case-sensitive)
- Check player exists: `playerRepo.GetByUsername()`
- Ensure both sender and recipient histories initialized

---

**Issue**: Chat history cleared unexpectedly

**Causes**:
- `/clear` command used
- Player disconnected and reconnected
- Server restart (in-memory only)

**Solutions**:
- Chat history is in-memory, cleared on disconnect
- For persistence, implement database storage
- Global history preserved across individual disconnects

---

**Issue**: Performance degradation with many messages

**Causes**:
- Excessive message history growth
- Too many concurrent players
- Inefficient message rendering

**Solutions**:
- Reduce `maxGlobalHistory` limit
- Implement message pagination
- Add database persistence with lazy loading
- Profile with `go tool pprof`

---

**Issue**: ANSI escape codes in messages

**Causes**:
- Sanitization not working
- Validation bypass

**Solutions**:
- Verify `validation.StripANSI()` called on input
- Check input character filtering
- Review `internal/validation/sanitization.go`

## Future Enhancements

### Planned Features

**Message Persistence** (Phase 21+):
- Database storage for message history
- Retrieve messages on login
- Searchable message archive
- Long-term conversation history

**Enhanced Moderation** (Phase 21+):
- Mute/ban from chat channels
- Message filtering (profanity, spam)
- Report system for abuse
- Admin message deletion

**Rich Messaging** (Phase 22+):
- Emojis and Unicode support
- Color codes for emphasis
- Mentions (@username)
- Links to in-game entities (systems, players, factions)

**Chat Notifications** (Phase 22+):
- Unread message counters
- Visual/audio notifications
- Priority channels (DMs, faction)
- Notification settings per channel

**Advanced Features** (Phase 23+):
- Message reactions (emoji reactions)
- Message editing/deletion
- Voice chat integration (external)
- Chat logs export

## API Reference

### Chat Manager Methods

```go
// Message sending
SendGlobalMessage(senderID uuid.UUID, sender, content string) *ChatMessage
SendSystemMessage(systemID, senderID uuid.UUID, sender, content string, recipientIDs []uuid.UUID) *ChatMessage
SendFactionMessage(factionID string, senderID uuid.UUID, sender, content string, memberIDs []uuid.UUID) *ChatMessage
SendDirectMessage(senderID uuid.UUID, sender string, recipientID uuid.UUID, recipient, content string) *ChatMessage
SendTradeMessage(senderID uuid.UUID, sender, content string) *ChatMessage
SendCombatNotification(playerIDs []uuid.UUID, content string)
BroadcastSystemMessage(channel ChatChannel, content string)

// Message retrieval
GetMessages(playerID uuid.UUID, channel ChatChannel, limit int) []*ChatMessage
GetDirectMessages(playerID uuid.UUID, otherPlayer string, limit int) []*ChatMessage
GetRecentGlobal(limit int) []*ChatMessage

// History management
GetOrCreateHistory(playerID uuid.UUID) *ChatHistory
RemovePlayerHistory(playerID uuid.UUID)
GetActiveDirectChats(playerID uuid.UUID) []string
ClearChannel(playerID uuid.UUID, channel ChatChannel)

// Statistics
GetStats() ChatStats
```

### ChatHistory Methods

```go
AddMessage(msg *ChatMessage)
GetMessages(channel ChatChannel, limit int) []*ChatMessage
GetDirectMessages(otherPlayer string, limit int) []*ChatMessage
GetActiveDirectChats() []string
ClearChannel(channel ChatChannel)
```

## Related Documentation

- [PLAYER_FACTIONS.md](PLAYER_FACTIONS.md) - Faction chat integration
- [PLAYER_PRESENCE.md](PLAYER_PRESENCE.md) - System chat player filtering
- [ADMIN_SYSTEM.md](ADMIN_SYSTEM.md) - Chat moderation and administration
- [ROADMAP.md](../ROADMAP.md) - Phase 9 implementation details
- [FEATURES.md](../FEATURES.md) - Complete feature catalog

## File Locations

### Core Implementation
- `internal/chat/manager.go` - Chat manager and message routing (262 lines)
- `internal/models/chat.go` - Chat message and history models
- `internal/tui/chat.go` - Chat UI screen (440 lines)

### Related Files
- `internal/validation/sanitization.go` - Message sanitization
- `internal/presence/manager.go` - Player presence for system chat
- `internal/factions/manager.go` - Faction membership for faction chat

---

**Document Version**: 1.0.0
**Last Updated**: 2025-01-15
**Maintainer**: Joshua Ferguson
