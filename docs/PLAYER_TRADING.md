# Player-to-Player Trading System

**Feature**: Player Trading
**Phase**: 12
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Player-to-Player Trading system enables secure trading of items, cargo, and credits between players. The system uses an escrow mechanism to prevent fraud and ensure both parties receive their agreed-upon items atomically.

### Key Features

- **Secure Escrow System**: Prevents scamming and ensures atomic trades
- **Multi-Item Trading**: Trade multiple items, cargo, and credits simultaneously
- **Trade Offers**: Send, receive, and counter trade offers
- **Trade History**: Complete audit trail of all trades
- **Real-Time Updates**: Instant notification of trade status changes
- **Privacy Controls**: Players can disable trade requests via settings
- **Timeout Protection**: Trades expire if not accepted within time limit

---

## Architecture

### Components

The trading system consists of the following components:

1. **Trade Manager** (`internal/trade/manager.go`)
   - Manages active trade offers
   - Implements escrow mechanics
   - Handles trade acceptance/rejection
   - Thread-safe with `sync.RWMutex`

2. **Trade UI** (`internal/tui/trade.go`)
   - Trade offer creation interface
   - Active trades display
   - Trade history viewer
   - Offer acceptance/rejection

3. **Data Models** (`internal/models/`)
   - `TradeOffer`: Pending trade proposal
   - `TradeItem`: Items included in trade
   - `TradeHistory`: Completed trade records

### Data Flow

```
Player A Initiates Trade
         ↓
Create Trade Offer
         ↓
Lock Offered Items in Escrow
         ↓
Send Notification to Player B
         ↓
Player B Reviews Offer
         ↓
Accept/Reject/Counter
         ↓
[If Accepted]
Atomic Item Transfer
         ↓
Release Escrow
         ↓
Log to Trade History
         ↓
Update Both Players' Inventories
```

### Thread Safety

The trade manager uses `sync.RWMutex` to ensure thread-safe operations:

- **Read Operations**: Multiple concurrent reads for trade lists
- **Write Operations**: Exclusive access for offer creation/acceptance
- **Escrow Management**: Locked during item transfer
- **Atomic Transactions**: Trade completion is all-or-nothing

---

## Implementation Details

### Trade Manager

The manager handles all trading operations:

```go
type Manager struct {
    mu sync.RWMutex

    // Active trades
    activeOffers  map[uuid.UUID]*models.TradeOffer // OfferID -> Offer
    playerOffers  map[uuid.UUID][]uuid.UUID        // PlayerID -> OfferIDs

    // Escrow
    escrowedItems map[uuid.UUID][]models.TradeItem // OfferID -> Items

    // Configuration
    offerTimeout  time.Duration
    maxActiveOffers int

    // Repositories
    playerRepo    *database.PlayerRepository
    inventoryManager *inventory.Manager
}
```

### Trade Offer Creation

**Validation Process**:
1. Check if recipient allows trade requests
2. Verify sender owns all offered items
3. Validate item quantities
4. Check maximum active offers limit
5. Lock offered items in escrow

**Offer Structure**:
```go
type TradeOffer struct {
    OfferID    uuid.UUID
    Sender     uuid.UUID
    Recipient  uuid.UUID

    // Sender's offer
    OfferedCredits int64
    OfferedCargo   []CargoItem
    OfferedItems   []InventoryItem

    // Recipient's request
    RequestedCredits int64
    RequestedCargo   []CargoItem
    RequestedItems   []InventoryItem

    Status     TradeStatus
    CreatedAt  time.Time
    ExpiresAt  time.Time
}
```

### Escrow System

The escrow system prevents fraud by locking items during trade:

**Escrow Process**:
1. **Lock Phase**: When offer created
   - Remove items from player's inventory
   - Store in escrow vault
   - Items inaccessible to both players

2. **Trade Phase**: During acceptance
   - Validate both sides still have required items
   - Prepare atomic transfer

3. **Complete Phase**: On acceptance
   - Transfer items simultaneously
   - Update both inventories atomically
   - Log transaction

4. **Cancel Phase**: On rejection/timeout
   - Return items to original owner
   - Remove from escrow
   - Unlock inventory

**Escrow Protection**:
```go
// Items are locked and cannot be:
- Traded again
- Sold to NPCs
- Jettisoned
- Used in crafting
- Transferred to another player outside trade
```

### Trade Acceptance

**Acceptance Flow**:
```go
func (m *Manager) AcceptTrade(offerID uuid.UUID, recipientID uuid.UUID) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 1. Validate offer exists and is pending
    offer := m.activeOffers[offerID]
    if offer.Status != StatusPending {
        return ErrInvalidTradeStatus
    }

    // 2. Verify recipient has requested items
    if !m.hasRequiredItems(recipientID, offer.RequestedItems) {
        return ErrInsufficientItems
    }

    // 3. Begin atomic transaction
    tx := m.db.BeginTransaction()

    // 4. Transfer items simultaneously
    m.transferItems(offer.Sender, offer.Recipient, offer.OfferedItems, tx)
    m.transferItems(offer.Recipient, offer.Sender, offer.RequestedItems, tx)

    // 5. Transfer credits
    m.transferCredits(offer.Sender, offer.Recipient, offer.OfferedCredits, tx)
    m.transferCredits(offer.Recipient, offer.Sender, offer.RequestedCredits, tx)

    // 6. Commit transaction
    if err := tx.Commit(); err != nil {
        tx.Rollback()
        return err
    }

    // 7. Release escrow and log
    m.releaseEscrow(offerID)
    m.logTrade(offer)

    return nil
}
```

### Trade Status States

```go
const (
    StatusPending   = "pending"   // Awaiting recipient response
    StatusAccepted  = "accepted"  // Completed successfully
    StatusRejected  = "rejected"  // Recipient declined
    StatusCancelled = "cancelled" // Sender cancelled
    StatusExpired   = "expired"   // Timed out
    StatusCountered = "countered" // Recipient made counter-offer
)
```

### Counter-Offers

Players can modify and re-propose trades:

**Counter-Offer Process**:
1. Recipient reviews initial offer
2. Modifies requested/offered items
3. Sends counter-offer back to original sender
4. Original offer expires
5. New offer created with swapped roles
6. Process repeats until accepted or cancelled

---

## User Interface

### Trade Screen

The trade UI provides comprehensive trading tools:

**Main View**:
```
=== Player Trading ===

Active Offers (3)                Incoming Offers (2)
────────────────────              ────────────────────
> Trade with Alice                Trade from Bob
  Trade with Charlie              Trade from David

Trade History | Settings
```

**Trade Offer Creation**:
```
=== Create Trade Offer ===

Recipient: Bob

Your Offer:
  Credits: [ 1000      ] CR
  Cargo:   [+] Add Cargo
    - Gold: 10 tons
    - Platinum: 5 tons
  Items:   [+] Add Item
    - Energy Shield MK2

Request From Bob:
  Credits: [ 5000      ] CR
  Cargo:   [+] Add Cargo
    - Diamonds: 3 tons
  Items:   [+] Add Item
    - Laser Cannon MK3

[Send Offer] [Cancel]
```

**Trade Review**:
```
=== Trade Offer from Alice ===

Alice Offers:              You Receive:
  5,000 CR                   ✓ 5,000 CR
  Gold (20 tons)             ✓ 20t cargo
  Laser Cannon MK2           ✓ Weapon

Alice Requests:            You Give:
  2,000 CR                   ✓ Available
  Diamonds (10 tons)         ✓ In cargo
  Energy Shield MK3          ✓ In inventory

Status: PENDING
Expires: 23h 45m

[Accept] [Reject] [Counter] [Cancel]
```

### Navigation

- **↑/↓**: Navigate offer list
- **Enter**: View/Accept offer
- **N**: Create new trade offer
- **C**: Counter current offer
- **R**: Reject offer
- **X**: Cancel your sent offer
- **H**: View trade history
- **ESC**: Return to main menu

---

## Integration with Other Systems

### Player Presence Integration

Trade requests only work with online players:

**Presence Checks**:
- Recipient must be online
- Both players must be in-game
- Disconnection cancels pending trades
- Automatic timeout if player goes offline

### Inventory Integration

**Inventory Management**:
```
Trade System ←→ Inventory Manager
           ↓
    Cargo System
           ↓
    Equipment System
```

**Validation**:
- Check cargo capacity
- Verify equipment compatibility
- Ensure items are not equipped
- Validate ownership

### Chat Integration

Trade notifications use the chat system:

**Notifications**:
- "Alice sent you a trade offer"
- "Trade with Bob completed"
- "Your trade offer to Charlie expired"
- "David countered your trade offer"

### Settings Integration

Privacy controls for trade requests:

```go
// In Settings
Privacy:
  AllowTradeRequests: true/false
  TradeNotifications: true/false
  BlockList: []PlayerID
```

---

## Testing

### Unit Tests

Test coverage for trade manager:

```go
func TestTrade_CreateOffer(t *testing.T)
func TestTrade_AcceptOffer(t *testing.T)
func TestTrade_RejectOffer(t *testing.T)
func TestTrade_CounterOffer(t *testing.T)
func TestTrade_EscrowProtection(t *testing.T)
func TestTrade_AtomicTransfer(t *testing.T)
func TestTrade_Timeout(t *testing.T)
func TestTrade_ConcurrentTrades(t *testing.T)
```

### Integration Tests

Full workflow testing:

1. **Simple Trade**:
   - Create offer with credits
   - Accept offer
   - Verify credit transfer
   - Check trade history

2. **Complex Trade**:
   - Offer multiple items + cargo + credits
   - Request multiple items
   - Accept and verify all transfers
   - Check escrow release

3. **Fraud Prevention**:
   - Create offer
   - Modify inventory while pending
   - Attempt accept (should fail)
   - Verify escrow integrity

4. **Edge Cases**:
   - Simultaneous trade acceptance
   - Database transaction failures
   - Player disconnection during trade
   - Inventory full scenarios

### Security Tests

- **Double-Spend Prevention**: Cannot use escrowed items
- **Inventory Manipulation**: Cannot modify locked items
- **Race Conditions**: Concurrent trade attempts
- **Rollback Testing**: Transaction failure recovery

---

## Configuration

### Manager Configuration

```go
cfg := &trade.Config{
    // Offer settings
    OfferTimeout:      24 * time.Hour,
    MaxActiveOffers:   10,
    MaxOfferItems:     20,

    // Escrow settings
    EscrowTimeout:     1 * time.Hour,
    ReleaseDelay:      5 * time.Second,

    // Fees (if enabled)
    TradeFeePercent:   0.0, // Currently disabled
    MinTradeFee:       0,
}
```

### Anti-Abuse Measures

**Rate Limiting**:
- Max 10 active offers per player
- Max 50 offers per hour
- Max 5 offers to same player per day

**Spam Prevention**:
- Blocked players cannot trade
- Repeated rejections auto-block
- Admin oversight for suspicious activity

---

## Troubleshooting

### Common Issues

**Problem**: Cannot send trade offer
**Solutions**:
- Check if recipient is online
- Verify recipient allows trade requests
- Ensure you haven't reached max offers
- Check if recipient blocked you

**Problem**: Trade acceptance fails
**Solutions**:
- Verify you still have requested items
- Check cargo space for incoming items
- Ensure credits available
- Review inventory locks

**Problem**: Items stuck in escrow
**Solutions**:
- Wait for trade timeout
- Cancel the pending offer
- Check trade manager status
- Contact admin if persistent

**Problem**: Trade history not showing
**Solutions**:
- Refresh UI
- Check database connectivity
- Verify trade completion
- Review manager logs

### Debug Commands

```bash
# Check active trades
curl http://localhost:8080/stats/trades

# View player trade history
SELECT * FROM trade_history WHERE player_id = '<player-id>';

# Check escrow status
grep "escrow" /var/log/terminal-velocity/server.log
```

---

## Future Enhancements

### Planned Features

1. **Auction House**
   - Public listing system
   - Bidding mechanics
   - Buyout prices
   - Auction history

2. **Trade Contracts**
   - Delivery missions between players
   - Escrow for future trades
   - Conditional trades
   - Multi-party agreements

3. **Trade Reputation**
   - Rating system
   - Trusted trader badges
   - Trade volume tracking
   - Scam reporting

4. **Advanced Features**
   - Trade templates
   - Bulk trading
   - Faction trading
   - Trade routing

5. **Market Integration**
   - Price comparison tools
   - Market value indicators
   - Profit calculators
   - Historical pricing

### Community Requests

- [ ] Trade notifications via external systems
- [ ] Gift mode (one-way trades)
- [ ] Trade bots/automation
- [ ] Virtual showrooms
- [ ] Trade insurance

---

## API Reference

### Core Functions

#### CreateTradeOffer

```go
func (m *Manager) CreateTradeOffer(
    senderID uuid.UUID,
    recipientID uuid.UUID,
    offer *TradeOffer,
) (uuid.UUID, error)
```

Creates a new trade offer.

**Parameters**:
- `senderID`: UUID of player creating offer
- `recipientID`: UUID of recipient player
- `offer`: Trade offer details

**Returns**: Offer UUID and error

**Errors**:
- `ErrPlayerOffline`: Recipient not online
- `ErrTradeDisabled`: Recipient disabled trades
- `ErrInsufficientItems`: Sender missing offered items
- `ErrMaxOffersReached`: Too many active offers

#### AcceptTradeOffer

```go
func (m *Manager) AcceptTradeOffer(
    offerID uuid.UUID,
    recipientID uuid.UUID,
) error
```

Accepts a pending trade offer.

**Returns**: Error if acceptance fails

#### RejectTradeOffer

```go
func (m *Manager) RejectTradeOffer(
    offerID uuid.UUID,
    recipientID uuid.UUID,
) error
```

Rejects a pending trade offer.

#### CounterTradeOffer

```go
func (m *Manager) CounterTradeOffer(
    offerID uuid.UUID,
    recipientID uuid.UUID,
    counterOffer *TradeOffer,
) (uuid.UUID, error)
```

Creates a counter-offer to an existing trade.

#### GetActiveOffers

```go
func (m *Manager) GetActiveOffers(
    playerID uuid.UUID,
) []*TradeOffer
```

Returns all active offers for a player.

#### GetTradeHistory

```go
func (m *Manager) GetTradeHistory(
    playerID uuid.UUID,
    limit int,
) []*TradeHistory
```

Returns recent trade history.

---

## Related Documentation

- [Player Presence](./PLAYER_PRESENCE.md) - Online player tracking
- [Chat System](./CHAT_SYSTEM.md) - Trade notifications
- [Settings System](./SETTINGS_SYSTEM.md) - Privacy controls
- [Inventory System](./INVENTORY.md) - Item management

---

## File Locations

**Core Implementation**:
- `internal/trade/manager.go` - Trade manager implementation
- `internal/models/trade.go` - Trade data models

**User Interface**:
- `internal/tui/trade.go` - Trade UI screens

**Database**:
- `scripts/schema.sql` - Trade table schema

**Tests**:
- `internal/trade/manager_test.go` - Unit tests

**Documentation**:
- `docs/PLAYER_TRADING.md` - This file
- `CHANGELOG.md` - Version history
- `ROADMAP.md` - Phase 12 details

---

**For questions or issues with the trading system, see the troubleshooting section above or contact the development team.**
