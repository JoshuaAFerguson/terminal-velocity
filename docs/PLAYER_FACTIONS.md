# Player Factions System Documentation

**Version**: 1.0.0
**Phase**: 10 - Player Factions
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

## Overview

The Player Factions system enables players to form persistent organizations with shared resources, hierarchy, and collective goals. Factions provide social structure, collaborative gameplay, and competitive elements through territory control and faction wars.

### Key Features

- **Faction Creation**: Players can found new factions with custom names and tags
- **Hierarchy System**: 3-tier rank structure (Leader, Officer, Member)
- **Shared Treasury**: Collective funds for faction operations
- **Member Management**: Recruit, promote, demote, and kick members
- **Level Progression**: Factions gain experience and level up
- **Faction Chat**: Private communication channel for members
- **Alignment System**: 5 faction types (Trader, Mercenary, Explorer, Pirate, Corporate)
- **Territory Control**: Claim and control star systems (Phase 11 integration)
- **Faction Wars**: PvP combat between rival factions (Phase 13 integration)

## Architecture

### Components

```
Player Factions System
├── Manager (internal/factions/manager.go)
│   ├── Faction lifecycle management
│   ├── Member tracking and permissions
│   ├── Treasury operations
│   └── Thread-safe operations (RWMutex)
│
├── UI (internal/tui/factions.go)
│   ├── Faction list browser
│   ├── Faction creation interface
│   ├── Member management UI
│   └── Treasury interface
│
└── Models (internal/models/faction.go)
    ├── PlayerFaction
    ├── FactionMember
    └── Alignment constants
```

### Data Structure

```go
type PlayerFaction struct {
    ID           uuid.UUID
    Name         string        // Faction name (max 30 chars)
    Tag          string        // Faction tag (3-5 chars, uppercase)
    Alignment    string        // trader, mercenary, explorer, pirate, corporate
    LeaderID     uuid.UUID     // Founder/current leader
    Officers     []uuid.UUID   // Officer IDs
    Members      []uuid.UUID   // All member IDs (includes leader & officers)
    Treasury     int64         // Shared credits
    Level        int           // Faction level
    Experience   int64         // Faction XP
    MemberLimit  int           // Max members (default: 50)
    IsRecruiting bool          // Public recruiting status
    CreatedAt    time.Time     // Foundation date
}
```

### Manager State

```go
type Manager struct {
    mu       sync.RWMutex
    factions map[uuid.UUID]*PlayerFaction  // All factions by ID
    names    map[string]uuid.UUID          // Name → ID (enforce uniqueness)
    tags     map[string]uuid.UUID          // Tag → ID (enforce uniqueness)
    members  map[uuid.UUID]uuid.UUID       // Player ID → Faction ID
}
```

## Faction Lifecycle

### 1. Creation

**Requirements**:
- Unique faction name (max 30 characters)
- Unique faction tag (3-5 uppercase letters)
- Player not already in a faction
- Creation cost: 10,000 CR (configurable)
- Valid alignment selection

**Process**:
```go
// Player initiates creation
faction, err := factionManager.CreateFaction(name, tag, founderID, alignment)

// System validates
1. Check name uniqueness
2. Check tag uniqueness
3. Verify player not in faction
4. Deduct creation cost from player
5. Create faction with founder as leader
6. Add founder to faction members
7. Initialize treasury at 0 CR
```

**Result**:
- New faction created with founder as leader
- Founder automatically joined as first member
- Faction appears in faction list
- Faction chat channel becomes available

### 2. Joining

**Methods**:
- **Invitation**: Leader/officers invite player
- **Application**: Player requests to join recruiting faction
- **Direct Join**: Join open recruiting factions

**Requirements**:
- Player not in another faction
- Faction has available slots (< memberLimit)
- Faction is recruiting (or invitation received)

**Process**:
```go
err := factionManager.JoinFaction(factionID, playerID)

// System validates
1. Check player not already in faction
2. Verify faction exists and has space
3. Add player to faction members
4. Update player-faction mapping
5. Grant access to faction chat
```

### 3. Leaving

**Methods**:
- **Voluntary**: Player leaves faction
- **Kicked**: Leader/officer removes player
- **Disbanded**: Faction dissolves (all members leave)

**Requirements**:
- Leader cannot leave (must transfer leadership first)
- Officers can only be kicked by leader
- Members can be kicked by officers or leader

**Process**:
```go
// Voluntary leave
err := factionManager.LeaveFaction(playerID)

// Kick member
err := factionManager.KickMember(factionID, kickerID, targetID)

// System performs
1. Verify permissions
2. Remove player from faction
3. Update member mappings
4. Remove faction chat access
5. Return player to independent status
```

### 4. Disbanding

**Requirements**:
- Only leader can disband
- All members must be kicked/leave first (or auto-removed)
- Treasury distributed or forfeited

**Process**:
```go
err := factionManager.DisbandFaction(factionID, leaderID)

// System performs
1. Verify leader permissions
2. Remove all members
3. Clear territory claims (Phase 11)
4. Delete faction from manager
5. Clean up all references
```

## Hierarchy & Permissions

### Rank Structure

| Rank | Icon | Permissions | Promotion |
|------|------|-------------|-----------|
| **Leader** | 👑 | Full control (all actions) | Founded faction or promoted by previous leader |
| **Officer** | ⭐ | Manage members, moderate chat, treasury access | Promoted by leader |
| **Member** | 👤 | Chat, contribute treasury, participate | Joined faction |

### Permission Matrix

| Action | Member | Officer | Leader |
|--------|--------|---------|--------|
| **View faction info** | ✅ | ✅ | ✅ |
| **Use faction chat** | ✅ | ✅ | ✅ |
| **Deposit treasury** | ✅ | ✅ | ✅ |
| **Withdraw treasury** | ❌ | ✅ | ✅ |
| **Invite players** | ❌ | ✅ | ✅ |
| **Kick members** | ❌ | ✅ (members only) | ✅ (all) |
| **Kick officers** | ❌ | ❌ | ✅ |
| **Promote to officer** | ❌ | ❌ | ✅ |
| **Demote officers** | ❌ | ❌ | ✅ |
| **Transfer leadership** | ❌ | ❌ | ✅ |
| **Disband faction** | ❌ | ❌ | ✅ |
| **Edit faction settings** | ❌ | ❌ | ✅ |
| **Claim territory** | ❌ | ✅ | ✅ |
| **Declare war** | ❌ | ❌ | ✅ |

### Rank Management

**Promoting to Officer**:
```go
// Only leader can promote
err := factionManager.PromoteMember(factionID, leaderID, targetID)

// Effects
- Player gains officer permissions
- Added to faction.Officers list
- Can now manage members and treasury
- Displays with ⭐ icon
```

**Demoting from Officer**:
```go
// Only leader can demote
err := factionManager.DemoteMember(factionID, leaderID, targetID)

// Effects
- Player loses officer permissions
- Removed from faction.Officers list
- Retains member status
- Displays with 👤 icon
```

**Transferring Leadership**:
```go
// Leader transfers to another member
err := factionManager.TransferLeadership(factionID, currentLeaderID, newLeaderID)

// Effects
- New leader gains full control
- Previous leader becomes officer (or member)
- Leadership icon (👑) updates
- Required before leader can leave
```

## Treasury System

### Purpose

Shared resource pool for faction operations:
- Territory claim costs
- Faction upgrades
- War reparations
- Bounties and rewards
- Collective investments

### Operations

**Deposit** (All members):
```go
// Any member can contribute
err := factionManager.Deposit(factionID, playerID, amount)

// Process
1. Verify player is member
2. Deduct credits from player
3. Add credits to faction treasury
4. Log transaction
```

**Withdraw** (Officers & Leader only):
```go
// Officers and leader can withdraw
err := factionManager.Withdraw(factionID, playerID, amount)

// Process
1. Verify officer/leader rank
2. Check sufficient treasury funds
3. Deduct from faction treasury
4. Add credits to player
5. Log transaction
```

**Treasury Limits**:
- No maximum balance
- Withdrawals limited by current treasury
- All transactions logged (future: audit trail)

**Use Cases**:
```
Examples:
- Member deposits 50,000 CR to help buy territory
- Officer withdraws 10,000 CR for faction supplies
- Leader withdraws 100,000 CR for war costs
- Collective saves for faction flagship (future)
```

## Alignment System

### Faction Types

| Alignment | Description | Playstyle | Benefits (Future) |
|-----------|-------------|-----------|-------------------|
| **Trader** | Merchant guilds, trade consortiums | Economy-focused, peaceful trade | +10% trade profits, better market access |
| **Mercenary** | Hired guns, bounty hunters | Combat-for-hire, neutral stance | +10% bounty rewards, combat bonuses |
| **Explorer** | Scientists, cartographers | Discovery, exploration | +15% exploration XP, rare finds |
| **Pirate** | Raiders, outlaws | Aggressive, loot-focused | +20% loot drops, stealth bonuses |
| **Corporate** | Megacorporations, business empires | Strategic, territorial | +10% territory income, influence bonuses |

### Alignment Selection

**During Creation**:
```
Available: trader, mercenary, explorer, pirate, corporate
Selection: Choose one (permanent for now)
Impact: Determines faction identity and future bonuses
```

**Future Enhancements**:
- Alignment-specific missions
- Faction wars based on alignment conflicts
- Reputation modifiers with NPC factions
- Exclusive ships/equipment per alignment

## Faction Progression

### Level System

**Experience Gain**:
- Member activities contribute faction XP
- Territory control generates XP
- Missions/quests completed as faction
- Faction wars and combat
- Trade volume (for Trader alignment)
- Exploration (for Explorer alignment)

**Leveling Benefits**:

| Level | Member Limit | Treasury Bonus | Perks |
|-------|--------------|----------------|-------|
| 1 | 50 | - | Basic faction features |
| 2 | 75 | - | Enhanced chat features |
| 3 | 100 | +5% deposits | Additional officer slots |
| 4 | 150 | +10% deposits | Territory control unlocked |
| 5 | 200 | +15% deposits | Faction missions unlocked |
| 10 | 500 | +25% deposits | Faction flagship (future) |

**XP Formula** (example):
```
Level 2: 10,000 XP
Level 3: 25,000 XP
Level 4: 50,000 XP
Level 5: 100,000 XP
(exponential growth)
```

### Member Limit

**Default**: 50 members
**Growth**: Increases with faction level
**Calculation**:
```go
memberLimit = 50 + (level * 25)
// Level 1: 50 members
// Level 5: 175 members
// Level 10: 300 members
```

**Reaching Limit**:
- Cannot join if faction is full
- Leader must kick members or level up
- Factions can set recruiting=false to control growth

## User Interface

### Faction List Screen

```
┌─────────────────────────────────────────────────────────────────┐
│                    🏛️  FACTIONS                                 │
│                                                                  │
│ Total Factions: 15 | Total Members: 327 | Recruiting: 8        │
│─────────────────────────────────────────────────────────────────│
│                                                                  │
│ Your Faction: Galactic Traders [GTR]                            │
│                                                                  │
│ Available Factions:                                              │
│                                                                  │
│ > Galactic Traders [GTR] - 45 members | Level 5 [Recruiting]   │
│   Star Corsairs [CRSR] - 38 members | Level 4                  │
│   Deep Space Explorers [DSE] - 22 members | Level 3 [Recruiting]│
│   Corporate Raiders [CORP] - 51 members | Level 6              │
│   ...                                                            │
│                                                                  │
│ C: Create | V: View My Faction | ESC: Back                      │
└─────────────────────────────────────────────────────────────────┘
```

### Create Faction Screen

```
┌─────────────────────────────────────────────────────────────────┐
│                    🏛️  CREATE FACTION                           │
│                                                                  │
│ Name: Galactic Traders█                                          │
│                                                                  │
│ Tag (3-5 chars): GTR                                             │
│                                                                  │
│ Alignment: trader                                                │
│                                                                  │
│ Available: trader, mercenary, explorer, pirate, corporate       │
│                                                                  │
│ Cost: 10,000 CR                                                  │
│                                                                  │
│ Tab: Next Field | Enter: Create | ESC: Cancel                   │
└─────────────────────────────────────────────────────────────────┘
```

### My Faction Screen

```
┌─────────────────────────────────────────────────────────────────┐
│                🏛️  Galactic Traders [GTR]                       │
│                                                                  │
│ Leader: PlayerName (you)                                         │
│ Founded: 2025-01-10                                              │
│ Members: 45/175                                                  │
│ Level: 5 (XP: 87,542)                                            │
│ Treasury: 2,450,000 CR                                           │
│ Alignment: trader                                                │
│─────────────────────────────────────────────────────────────────│
│                                                                  │
│ Members:                                                         │
│   👑 Leader: PlayerName (you)                                    │
│   ⭐ Officer: TraderJoe                                          │
│   ⭐ Officer: MerchantMary                                       │
│   👤 Member: NewPlayer                                           │
│   👤 Member: Pilot42                                             │
│   ... (40 more members)                                          │
│                                                                  │
│ D: Deposit | W: Withdraw | M: Manage Members | ESC: Back        │
└─────────────────────────────────────────────────────────────────┘
```

### Member Management Screen

```
┌─────────────────────────────────────────────────────────────────┐
│                🏛️  Manage Members                               │
│                                                                  │
│ > TraderJoe (Officer) - Online                                  │
│   MerchantMary (Officer) - Offline                              │
│   NewPlayer (Member) - Online                                   │
│   Pilot42 (Member) - Offline                                    │
│                                                                  │
│ Selected: TraderJoe                                              │
│ Actions:                                                         │
│   P: Promote to Officer / Demote to Member                       │
│   K: Kick from Faction                                           │
│   T: Transfer Leadership                                         │
│   V: View Player Profile                                         │
│                                                                  │
│ ESC: Back                                                        │
└─────────────────────────────────────────────────────────────────┘
```

## Integration with Other Systems

### Chat System (Phase 9)

**Faction Chat Channel**:
```go
// Send faction message
chatManager.SendFactionMessage(factionID, senderID, sender, content, memberIDs)

// Routing
- Message sent to all faction members
- Only visible to faction members
- Integrated with chat UI (Channel 3)
```

**Usage**:
```
Press '3' in Chat screen to access Faction Chat
Type message → sent to all faction members
Private strategy discussions
Coordination and planning
```

### Territory Control (Phase 11)

**System Claiming**:
```go
// Faction claims system
territoryManager.ClaimSystem(systemID, factionID, playerID)

// Requirements
- Player must be officer or leader
- Faction must have sufficient treasury
- System must be unclaimed or conquerable
```

**Passive Income**:
```
Controlled territories generate credits
Income deposited to faction treasury
Distributed based on territory value
Provides ongoing faction funding
```

### PvP Combat (Phase 13)

**Faction Wars**:
```go
// Declare war on another faction
pvpManager.DeclareWar(attackerFactionID, defenderFactionID)

// Effects
- Faction members can attack each other
- Territory can be contested
- War spoils to faction treasury
- XP rewards for faction
```

**Alliance System** (Future):
```go
// Ally with another faction
factionManager.CreateAlliance(faction1ID, faction2ID)

// Benefits
- Shared faction chat
- No friendly fire in combat
- Joint territory control
- Combined treasury (optional)
```

### Leaderboards (Phase 14)

**Faction Rankings**:
```
- Top Factions by Members
- Top Factions by Treasury
- Top Factions by Level/XP
- Top Factions by Territory Control
```

### Player Presence (Phase 15)

**Online Members**:
```go
// Get online faction members
onlineMembers := presenceManager.GetOnlinePlayers(faction.Members)

// Display
- Show online/offline status in member list
- Real-time presence updates
- Coordinate activities with online members
```

## Implementation Details

### Thread Safety

**RWMutex Protection**:
```go
// All manager operations are thread-safe
func (m *Manager) GetFaction(factionID uuid.UUID) (*PlayerFaction, error) {
    m.mu.RLock()         // Read lock for queries
    defer m.mu.RUnlock()
    // ...
}

func (m *Manager) CreateFaction(...) (*PlayerFaction, error) {
    m.mu.Lock()          // Write lock for modifications
    defer m.mu.Unlock()
    // ...
}
```

**Concurrent Access**:
- Multiple players can view factions simultaneously
- Writes (create, join, leave) are serialized
- No race conditions in member management

### Error Handling

**Defined Errors**:
```go
var (
    ErrFactionNotFound   = errors.New("faction not found")
    ErrNotMember         = errors.New("player is not a member")
    ErrInsufficientRank  = errors.New("insufficient rank for this action")
    ErrFactionFull       = errors.New("faction has reached member limit")
    ErrAlreadyMember     = errors.New("player is already a member")
    ErrInsufficientFunds = errors.New("insufficient faction treasury funds")
    ErrNameTaken         = errors.New("faction name already taken")
    ErrTagTaken          = errors.New("faction tag already taken")
)
```

**Validation**:
- Name/tag uniqueness enforced at manager level
- Rank permissions checked before operations
- Membership conflicts prevented
- Treasury operations validated

### Data Persistence

**Current State** (Phase 10):
- In-memory storage only
- Factions persist during server uptime
- Lost on server restart

**Future** (Phase 21+):
```sql
-- Database schema (planned)
CREATE TABLE factions (
    id UUID PRIMARY KEY,
    name VARCHAR(30) UNIQUE NOT NULL,
    tag VARCHAR(5) UNIQUE NOT NULL,
    alignment VARCHAR(20) NOT NULL,
    leader_id UUID NOT NULL REFERENCES players(id),
    treasury BIGINT DEFAULT 0,
    level INT DEFAULT 1,
    experience BIGINT DEFAULT 0,
    member_limit INT DEFAULT 50,
    is_recruiting BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE faction_members (
    faction_id UUID REFERENCES factions(id),
    player_id UUID REFERENCES players(id),
    rank VARCHAR(20) NOT NULL, -- 'leader', 'officer', 'member'
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (faction_id, player_id)
);

CREATE TABLE faction_treasury_log (
    id UUID PRIMARY KEY,
    faction_id UUID REFERENCES factions(id),
    player_id UUID REFERENCES players(id),
    amount BIGINT NOT NULL,
    transaction_type VARCHAR(20) NOT NULL, -- 'deposit', 'withdraw'
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Testing

### Manual Testing Checklist

**Faction Creation**:
- [ ] Create faction with valid name/tag
- [ ] Try duplicate name (should fail)
- [ ] Try duplicate tag (should fail)
- [ ] Verify creation cost deducted
- [ ] Verify founder becomes leader
- [ ] Check faction appears in list

**Joining/Leaving**:
- [ ] Join faction as new player
- [ ] Try joining while in another faction (should fail)
- [ ] Join faction at member limit (should fail)
- [ ] Leave faction as member
- [ ] Try leaving as leader (should fail without transfer)

**Member Management**:
- [ ] Promote member to officer (as leader)
- [ ] Demote officer to member (as leader)
- [ ] Try promoting as non-leader (should fail)
- [ ] Kick member (as officer)
- [ ] Kick officer (as leader)
- [ ] Try kicking leader (should fail)

**Treasury**:
- [ ] Deposit credits as member
- [ ] Withdraw credits as officer
- [ ] Try withdrawing as member (should fail)
- [ ] Try withdrawing more than treasury (should fail)
- [ ] Verify transactions update balances correctly

**Permissions**:
- [ ] Test all member actions (chat, view, deposit)
- [ ] Test officer actions (kick, invite, withdraw)
- [ ] Test leader actions (promote, demote, disband)
- [ ] Verify rank checks enforce permissions

**UI/UX**:
- [ ] Navigate faction list
- [ ] View faction details
- [ ] Create faction flow
- [ ] Member list displays correctly
- [ ] Online/offline status (with presence system)

**Integration**:
- [ ] Faction chat channel works
- [ ] Territory claim requires faction (Phase 11)
- [ ] PvP faction war works (Phase 13)
- [ ] Faction leaderboards (Phase 14)

### Automated Tests

Location: `internal/factions/manager_test.go`

```bash
# Run faction tests
go test ./internal/factions -v

# Test with race detector
go test -race ./internal/factions
```

**Test Coverage**:
```go
// Test cases
- TestCreateFaction
- TestJoinFaction
- TestLeaveFaction
- TestKickMember
- TestPromoteMember
- TestDemoteMember
- TestDeposit
- TestWithdraw
- TestPermissions
- TestConcurrentAccess
```

## Performance Considerations

### Memory Usage

**Per-Faction Overhead**:
```
PlayerFaction struct: ~500 bytes base
+ 24 bytes per member (UUID)
+ string allocations (name, tag)

Example:
50-member faction: ~1.7 KB
100-member faction: ~3 KB
```

**Server-wide**:
```
100 factions × 50 members avg = ~170 KB
1000 factions × 100 members avg = ~3 MB
```

### Lookup Performance

**O(1) Operations**:
- GetFaction by ID: `factions[id]`
- GetFactionByName: `names[name]` → `factions[id]`
- GetFactionByTag: `tags[tag]` → `factions[id]`
- GetPlayerFaction: `members[playerID]` → `factions[id]`

**O(n) Operations**:
- GetAllFactions: Iterate all factions
- GetStats: Aggregate all factions

### Optimization Strategies

**Caching**:
- Faction list cache (refresh on changes)
- Stats cache (refresh every 30s)
- Member counts cached in faction struct

**Database** (Future):
- Index on name, tag for uniqueness checks
- Index on leader_id for leader queries
- Composite index on (faction_id, player_id) for membership

## Configuration

### Default Settings

```go
// In internal/models/faction.go
const (
    DefaultMemberLimit   = 50
    DefaultTreasuryStart = 0
    DefaultFactionLevel  = 1
    DefaultFactionXP     = 0
    FactionCreationCost  = 10000  // Credits
)
```

### Customization

**Adjusting Member Limits**:
```go
// Modify in models.NewPlayerFaction()
faction.MemberLimit = 100  // Increase default limit
```

**Adjusting Creation Cost**:
```go
// In TUI faction creation flow
const factionCreationCost = 50000  // Increase to 50,000 CR
```

**Adding New Alignments**:
```go
// In internal/models/faction.go
const (
    AlignmentTrader    = "trader"
    AlignmentMercenary = "mercenary"
    AlignmentExplorer  = "explorer"
    AlignmentPirate    = "pirate"
    AlignmentCorporate = "corporate"
    AlignmentMilitary  = "military"      // New alignment
    AlignmentScientist = "scientist"     // New alignment
)
```

## Troubleshooting

### Common Issues

**Issue**: Cannot create faction (name/tag taken)

**Solutions**:
- Try different name or tag
- Check existing factions list
- Use `/factions list` to see all names/tags
- Tags are case-insensitive and unique

---

**Issue**: Cannot leave faction as leader

**Solutions**:
- Transfer leadership to another member first
- Use "Transfer Leadership" option in member management
- Promote an officer, then transfer to them
- Alternatively, disband the faction

---

**Issue**: Cannot kick an officer

**Solutions**:
- Only leader can kick officers
- Demote officer to member first, then kick (as officer)
- Or kick directly if you're the leader

---

**Issue**: Treasury withdrawal failed

**Solutions**:
- Verify you're an officer or leader
- Check treasury has sufficient funds
- Ensure amount is positive integer
- Check for permission errors

---

**Issue**: Faction disappeared after server restart

**Solutions**:
- Factions are in-memory only (Phase 10)
- Database persistence planned for Phase 21+
- For now, factions only persist during server uptime
- Use save/load system when implemented

## Future Enhancements

### Planned Features (Phase 21+)

**Database Persistence**:
- PostgreSQL storage for factions
- Member history tracking
- Treasury transaction logs
- Audit trail for all operations

**Advanced Permissions**:
- Custom rank creation beyond 3 tiers
- Fine-grained permission system
- Role templates (Treasurer, Recruiter, etc.)

**Faction Diplomacy**:
- Alliance system
- Trade agreements between factions
- Faction reputation system
- Diplomatic victories in wars

**Enhanced Progression**:
- Faction missions and quests
- Collective achievements
- Faction-specific storylines
- Unique rewards for high-level factions

**Faction Bases**:
- Player-owned stations
- Faction headquarters
- Defensive structures
- Economic infrastructure

**Advanced Treasury**:
- Budget system (allocate funds)
- Automated payments (salaries)
- Investment system (generate returns)
- Faction marketplace

**Communication**:
- Faction message board
- Event calendar
- Voice chat integration (external)
- In-game mail system

## API Reference

### Faction Manager Methods

```go
// Faction lifecycle
CreateFaction(name, tag string, founderID uuid.UUID, alignment string) (*PlayerFaction, error)
DisbandFaction(factionID, leaderID uuid.UUID) error

// Faction queries
GetFaction(factionID uuid.UUID) (*PlayerFaction, error)
GetFactionByName(name string) (*PlayerFaction, error)
GetFactionByTag(tag string) (*PlayerFaction, error)
GetPlayerFaction(playerID uuid.UUID) (*PlayerFaction, error)
GetAllFactions() []*PlayerFaction

// Member management
JoinFaction(factionID, playerID uuid.UUID) error
LeaveFaction(playerID uuid.UUID) error
KickMember(factionID, kickerID, targetID uuid.UUID) error

// Rank management
PromoteMember(factionID, promoterID, targetID uuid.UUID) error
DemoteMember(factionID, demoterID, targetID uuid.UUID) error
TransferLeadership(factionID, currentLeaderID, newLeaderID uuid.UUID) error

// Treasury operations
Deposit(factionID, playerID uuid.UUID, amount int64) error
Withdraw(factionID, playerID uuid.UUID, amount int64) error

// Statistics
GetStats() FactionStats
```

### PlayerFaction Methods

```go
// Member checks
IsMember(playerID uuid.UUID) bool
IsOfficer(playerID uuid.UUID) bool
IsLeader(playerID uuid.UUID) bool
AddMember(playerID uuid.UUID) bool
RemoveMember(playerID uuid.UUID)

// Rank management
PromoteToOfficer(playerID uuid.UUID) bool
DemoteFromOfficer(playerID uuid.UUID) bool

// Treasury
Deposit(amount int64)
Withdraw(amount int64) bool

// Utility
GetFullName() string  // Returns "Name [TAG]"
```

## Related Documentation

- [CHAT_SYSTEM.md](CHAT_SYSTEM.md) - Faction chat integration
- [TERRITORY_CONTROL.md](TERRITORY_CONTROL.md) - System claiming and control
- [PVP_COMBAT.md](PVP_COMBAT.md) - Faction wars and combat
- [LEADERBOARDS.md](LEADERBOARDS.md) - Faction rankings
- [PLAYER_PRESENCE.md](PLAYER_PRESENCE.md) - Online member tracking
- [ROADMAP.md](../ROADMAP.md) - Phase 10 implementation details
- [FEATURES.md](../FEATURES.md) - Complete feature catalog

## File Locations

### Core Implementation
- `internal/factions/manager.go` - Faction manager (400+ lines)
- `internal/models/faction.go` - Faction data models
- `internal/tui/factions.go` - Faction UI (252 lines)

### Related Files
- `internal/chat/manager.go` - Faction chat integration
- `internal/territory/manager.go` - Territory control (Phase 11)
- `internal/pvp/manager.go` - Faction wars (Phase 13)

---

**Document Version**: 1.0.0
**Last Updated**: 2025-01-15
**Maintainer**: Joshua Ferguson
