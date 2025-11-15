# Inventory System Technical Specification
**Version:** 1.0.0
**Author:** Joshua Ferguson
**Date:** 2025-11-15
**Status:** Design Phase

## Executive Summary

This specification defines a hybrid inventory system for Terminal Velocity that maintains backward compatibility with the existing commodity cargo system while introducing UUID-based item tracking for equipment, weapons, and unique items. This system unblocks four key features: mail item attachments, marketplace auctions, contracts, and bounty postings.

## Goals & Non-Goals

### Goals
- ✅ Enable UUID-based item references for mail and marketplace
- ✅ Maintain backward compatibility with existing commodity cargo
- ✅ Support equipment/weapon trading between players
- ✅ Provide reusable item picker UI component
- ✅ Zero data loss during migration

### Non-Goals
- ❌ Replace or refactor existing cargo system
- ❌ Add player housing/storage (future feature)
- ❌ Implement item crafting system (future feature)
- ❌ Add item durability/condition tracking (v2)

## Architecture Overview

### Current System (Commodities)
```
Ship Cargo: []CargoItem
  ├─ CommodityID: string ("food", "metals", "electronics")
  ├─ Quantity: int
  └─ BuyPrice: float64

Use Cases: Bulk trading, market buy/sell
```

### New System (Items - Equipment/Weapons)
```
Player Inventory: []PlayerItem
  ├─ ItemID: uuid.UUID (unique instance)
  ├─ EquipmentType: string ("weapon", "outfit", "special")
  ├─ EquipmentID: string ("laser_cannon", "shield_booster")
  ├─ Properties: JSONB (stats, mods, custom data)
  ├─ Location: string ("ship", "station_storage", "mail", "escrow")
  └─ LocationID: uuid.UUID (ship_id, planet_id, mail_id, etc.)

Use Cases: Equipment trading, mail attachments, marketplace
```

### Hybrid Approach
- **Commodities** remain as `[]CargoItem` on ships (string-based, bulk)
- **Items** use new `player_items` table (UUID-based, individual)
- **UI** distinguishes between cargo bay (bulk) and equipment bay (items)
- **Trading** supports both commodity trading and item trading

## Database Schema

### New Tables

#### `player_items`
Primary table for UUID-based inventory items.

```sql
CREATE TABLE player_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Item type and reference
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('weapon', 'outfit', 'special', 'quest')),
    equipment_id VARCHAR(100) NOT NULL, -- References equipment definition (e.g., "laser_cannon")

    -- Current location
    location VARCHAR(50) NOT NULL CHECK (location IN ('ship', 'station_storage', 'mail', 'escrow', 'auction')),
    location_id UUID, -- ship_id, planet_id, mail_id, auction_id, etc.

    -- Item properties (for modifications, upgrades, etc.)
    properties JSONB DEFAULT '{}',

    -- Metadata
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_player_items_player (player_id),
    INDEX idx_player_items_location (location, location_id),
    INDEX idx_player_items_type (item_type, equipment_id)
);

COMMENT ON TABLE player_items IS 'UUID-based inventory for weapons, outfits, and special items';
COMMENT ON COLUMN player_items.equipment_id IS 'References equipment definition from outfitting/data.go';
COMMENT ON COLUMN player_items.properties IS 'JSON properties for mods, upgrades, custom stats';
```

#### `item_transfers`
Audit trail for item movements (trade, mail, auction).

```sql
CREATE TABLE item_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES player_items(id) ON DELETE CASCADE,

    from_player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    to_player_id UUID REFERENCES players(id) ON DELETE SET NULL,

    transfer_type VARCHAR(50) NOT NULL CHECK (transfer_type IN ('trade', 'mail', 'auction', 'contract', 'admin')),
    transfer_id UUID, -- trade_id, mail_id, auction_id, etc.

    -- Metadata
    transferred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_item_transfers_item (item_id),
    INDEX idx_item_transfers_players (from_player_id, to_player_id),
    INDEX idx_item_transfers_type (transfer_type, transfer_id)
);

COMMENT ON TABLE item_transfers IS 'Audit log for all item movements between players';
```

### Schema Migration

**Migration:** `scripts/migrations/019_inventory_system.sql`

```sql
-- Create player_items table
CREATE TABLE player_items (
    -- (schema from above)
);

-- Create item_transfers table
CREATE TABLE item_transfers (
    -- (schema from above)
);

-- Create indexes
CREATE INDEX idx_player_items_player ON player_items(player_id);
CREATE INDEX idx_player_items_location ON player_items(location, location_id);
CREATE INDEX idx_player_items_type ON player_items(item_type, equipment_id);
CREATE INDEX idx_item_transfers_item ON item_transfers(item_id);
CREATE INDEX idx_item_transfers_players ON item_transfers(from_player_id, to_player_id);
CREATE INDEX idx_item_transfers_type ON item_transfers(transfer_type, transfer_id);

-- Migration for existing equipped items (if applicable)
-- NOTE: Current ships.equipment is []string, not actual items
-- This migration creates UUID items for currently equipped gear
INSERT INTO player_items (player_id, item_type, equipment_id, location, location_id)
SELECT
    s.player_id,
    CASE
        WHEN e.equipment LIKE 'weapon_%' THEN 'weapon'
        ELSE 'outfit'
    END as item_type,
    e.equipment as equipment_id,
    'ship' as location,
    s.id as location_id
FROM ships s
CROSS JOIN LATERAL unnest(s.equipment) AS e(equipment)
WHERE array_length(s.equipment, 1) > 0;

-- Record migration timestamp
INSERT INTO schema_migrations (version, applied_at)
VALUES ('019_inventory_system', CURRENT_TIMESTAMP);
```

## Data Models

### Go Structs

#### `internal/models/item.go`
```go
// File: internal/models/item.go
// Project: Terminal Velocity
// Description: Player inventory item models
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package models

import (
    "time"
    "encoding/json"
    "github.com/google/uuid"
)

// ItemType represents the category of inventory item
type ItemType string

const (
    ItemTypeWeapon  ItemType = "weapon"
    ItemTypeOutfit  ItemType = "outfit"
    ItemTypeSpecial ItemType = "special"
    ItemTypeQuest   ItemType = "quest"
)

// ItemLocation represents where an item is currently stored
type ItemLocation string

const (
    LocationShip           ItemLocation = "ship"
    LocationStationStorage ItemLocation = "station_storage"
    LocationMail           ItemLocation = "mail"
    LocationEscrow         ItemLocation = "escrow"
    LocationAuction        ItemLocation = "auction"
)

// PlayerItem represents a single inventory item with UUID
type PlayerItem struct {
    ID         uuid.UUID       `json:"id"`
    PlayerID   uuid.UUID       `json:"player_id"`
    ItemType   ItemType        `json:"item_type"`
    EquipmentID string         `json:"equipment_id"` // References equipment definition
    Location   ItemLocation    `json:"location"`
    LocationID *uuid.UUID      `json:"location_id,omitempty"`
    Properties json.RawMessage `json:"properties"` // JSONB for mods/upgrades
    AcquiredAt time.Time       `json:"acquired_at"`
    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}

// ItemTransfer represents a transfer audit entry
type ItemTransfer struct {
    ID           uuid.UUID  `json:"id"`
    ItemID       uuid.UUID  `json:"item_id"`
    FromPlayerID *uuid.UUID `json:"from_player_id,omitempty"`
    ToPlayerID   *uuid.UUID `json:"to_player_id,omitempty"`
    TransferType string     `json:"transfer_type"` // trade, mail, auction, etc.
    TransferID   *uuid.UUID `json:"transfer_id,omitempty"`
    TransferredAt time.Time `json:"transferred_at"`
}

// ItemProperties represents modifiable item properties
type ItemProperties struct {
    Mods       []string          `json:"mods,omitempty"`       // Applied modifications
    Upgrades   map[string]int    `json:"upgrades,omitempty"`   // Upgrade levels
    CustomData map[string]interface{} `json:"custom,omitempty"` // Extension point
}

// GetProperties unmarshals JSONB properties
func (i *PlayerItem) GetProperties() (*ItemProperties, error) {
    if len(i.Properties) == 0 {
        return &ItemProperties{}, nil
    }

    var props ItemProperties
    if err := json.Unmarshal(i.Properties, &props); err != nil {
        return nil, err
    }
    return &props, nil
}

// SetProperties marshals and sets JSONB properties
func (i *PlayerItem) SetProperties(props *ItemProperties) error {
    data, err := json.Marshal(props)
    if err != nil {
        return err
    }
    i.Properties = data
    return nil
}

// GetEquipmentName returns human-readable equipment name
func (i *PlayerItem) GetEquipmentName() string {
    // TODO: Look up from equipment definitions
    // For now, return ID formatted nicely
    return formatEquipmentID(i.EquipmentID)
}

// formatEquipmentID converts "laser_cannon" to "Laser Cannon"
func formatEquipmentID(id string) string {
    // Simple title case conversion
    // TODO: Use actual equipment definitions
    return id
}
```

## Repository Layer

### `internal/database/item_repository.go`
```go
// File: internal/database/item_repository.go
// Project: Terminal Velocity
// Description: Repository for player inventory items
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package database

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "terminal-velocity/internal/models"
)

// ItemRepository handles database operations for player items
type ItemRepository struct {
    conn *pgxpool.Pool
}

// NewItemRepository creates a new ItemRepository
func NewItemRepository(conn *pgxpool.Pool) *ItemRepository {
    return &ItemRepository{conn: conn}
}

// GetPlayerItems returns all items owned by a player
func (r *ItemRepository) GetPlayerItems(ctx context.Context, playerID uuid.UUID) ([]*models.PlayerItem, error) {
    query := `
        SELECT id, player_id, item_type, equipment_id, location, location_id,
               properties, acquired_at, created_at, updated_at
        FROM player_items
        WHERE player_id = $1
        ORDER BY acquired_at DESC
    `

    rows, err := r.conn.Query(ctx, query, playerID)
    if err != nil {
        return nil, fmt.Errorf("failed to query player items: %w", err)
    }
    defer rows.Close()

    var items []*models.PlayerItem
    for rows.Next() {
        var item models.PlayerItem
        err := rows.Scan(
            &item.ID, &item.PlayerID, &item.ItemType, &item.EquipmentID,
            &item.Location, &item.LocationID, &item.Properties,
            &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan item: %w", err)
        }
        items = append(items, &item)
    }

    return items, rows.Err()
}

// GetItemsByLocation returns items at a specific location
func (r *ItemRepository) GetItemsByLocation(ctx context.Context, playerID uuid.UUID, location models.ItemLocation, locationID uuid.UUID) ([]*models.PlayerItem, error) {
    query := `
        SELECT id, player_id, item_type, equipment_id, location, location_id,
               properties, acquired_at, created_at, updated_at
        FROM player_items
        WHERE player_id = $1 AND location = $2 AND location_id = $3
        ORDER BY item_type, equipment_id
    `

    rows, err := r.conn.Query(ctx, query, playerID, location, locationID)
    if err != nil {
        return nil, fmt.Errorf("failed to query items by location: %w", err)
    }
    defer rows.Close()

    var items []*models.PlayerItem
    for rows.Next() {
        var item models.PlayerItem
        err := rows.Scan(
            &item.ID, &item.PlayerID, &item.ItemType, &item.EquipmentID,
            &item.Location, &item.LocationID, &item.Properties,
            &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan item: %w", err)
        }
        items = append(items, &item)
    }

    return items, rows.Err()
}

// CreateItem creates a new player item
func (r *ItemRepository) CreateItem(ctx context.Context, item *models.PlayerItem) error {
    query := `
        INSERT INTO player_items (player_id, item_type, equipment_id, location, location_id, properties)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, acquired_at, created_at, updated_at
    `

    err := r.conn.QueryRow(ctx, query,
        item.PlayerID, item.ItemType, item.EquipmentID,
        item.Location, item.LocationID, item.Properties,
    ).Scan(&item.ID, &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to create item: %w", err)
    }

    return nil
}

// UpdateItemLocation moves an item to a new location
func (r *ItemRepository) UpdateItemLocation(ctx context.Context, itemID uuid.UUID, location models.ItemLocation, locationID uuid.UUID) error {
    query := `
        UPDATE player_items
        SET location = $1, location_id = $2, updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
    `

    result, err := r.conn.Exec(ctx, query, location, locationID, itemID)
    if err != nil {
        return fmt.Errorf("failed to update item location: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("item not found: %s", itemID)
    }

    return nil
}

// TransferItem transfers an item to another player (atomic with audit log)
func (r *ItemRepository) TransferItem(ctx context.Context, itemID uuid.UUID, toPlayerID uuid.UUID, transferType string, transferID uuid.UUID) error {
    tx, err := r.conn.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // Get current owner
    var fromPlayerID uuid.UUID
    err = tx.QueryRow(ctx, "SELECT player_id FROM player_items WHERE id = $1", itemID).Scan(&fromPlayerID)
    if err != nil {
        return fmt.Errorf("failed to get item owner: %w", err)
    }

    // Update ownership
    _, err = tx.Exec(ctx, "UPDATE player_items SET player_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", toPlayerID, itemID)
    if err != nil {
        return fmt.Errorf("failed to transfer item: %w", err)
    }

    // Create audit log entry
    _, err = tx.Exec(ctx,
        "INSERT INTO item_transfers (item_id, from_player_id, to_player_id, transfer_type, transfer_id) VALUES ($1, $2, $3, $4, $5)",
        itemID, fromPlayerID, toPlayerID, transferType, transferID,
    )
    if err != nil {
        return fmt.Errorf("failed to log transfer: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// DeleteItem removes an item (e.g., consumed, destroyed)
func (r *ItemRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
    query := "DELETE FROM player_items WHERE id = $1"
    result, err := r.conn.Exec(ctx, query, itemID)
    if err != nil {
        return fmt.Errorf("failed to delete item: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("item not found: %s", itemID)
    }

    return nil
}

// GetItemTransferHistory returns transfer audit log for an item
func (r *ItemRepository) GetItemTransferHistory(ctx context.Context, itemID uuid.UUID) ([]*models.ItemTransfer, error) {
    query := `
        SELECT id, item_id, from_player_id, to_player_id, transfer_type, transfer_id, transferred_at
        FROM item_transfers
        WHERE item_id = $1
        ORDER BY transferred_at DESC
    `

    rows, err := r.conn.Query(ctx, query, itemID)
    if err != nil {
        return nil, fmt.Errorf("failed to query transfer history: %w", err)
    }
    defer rows.Close()

    var transfers []*models.ItemTransfer
    for rows.Next() {
        var transfer models.ItemTransfer
        err := rows.Scan(
            &transfer.ID, &transfer.ItemID, &transfer.FromPlayerID,
            &transfer.ToPlayerID, &transfer.TransferType, &transfer.TransferID,
            &transfer.TransferredAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan transfer: %w", err)
        }
        transfers = append(transfers, &transfer)
    }

    return transfers, rows.Err()
}
```

## UI Components

### Item Picker Component

**Location:** `internal/tui/components/item_picker.go`

**Features:**
- Multi-select support (checkbox list)
- Filter by item type (weapons, outfits, all)
- Filter by location (ship, station)
- Search/filter by name
- Display item stats inline
- Quantity selection (for future stacking)

**UI Layout:**
```
┌─ Select Items ────────────────────────────────┐
│ Filter: [All Items ▼] Location: [Ship ▼]     │
│ Search: _____________________                 │
├───────────────────────────────────────────────┤
│ [✓] Laser Cannon MkII         (Weapon)       │
│ [ ] Plasma Torpedo Launcher   (Weapon)       │
│ [✓] Shield Booster +3         (Outfit)       │
│ [ ] Cargo Expansion           (Outfit)       │
│ [ ] Jump Drive Upgrade        (Outfit)       │
│                                               │
│                                               │
├───────────────────────────────────────────────┤
│ Selected: 2 items                             │
│ [Confirm] [Cancel]                            │
└───────────────────────────────────────────────┘
```

**Usage Example:**
```go
// In mail compose screen
picker := NewItemPicker(m.itemRepo, m.player.ID)
picker.SetFilter(ItemTypeAll)
picker.SetLocation(LocationShip)
picker.SetMultiSelect(true)

// On confirmation
selectedItems := picker.GetSelectedItems() // []uuid.UUID
```

### Item List Component

**Location:** `internal/tui/components/item_list.go`

**Features:**
- Read-only item display
- Grouped by type (weapons, outfits)
- Show item stats
- Pagination for long lists

**UI Layout:**
```
┌─ Your Items ──────────────────────────────────┐
│ Weapons (2)                                   │
│   • Laser Cannon MkII    [Dmg: 45] [Rng: 3]  │
│   • Plasma Torpedo       [Dmg: 120] [Rng: 5] │
│                                               │
│ Outfits (3)                                   │
│   • Shield Booster +3    [Shield: +50]       │
│   • Cargo Expansion      [Cargo: +20t]       │
│   • Jump Drive Upgrade   [Jump: +2]          │
│                                               │
│ Page 1/1                                      │
└───────────────────────────────────────────────┘
```

## Integration Points

### 1. Mail System
**File:** `internal/tui/mail.go:676`

**Current Code:**
```go
// TODO: This would come from item picker
attachments := []uuid.UUID{}
```

**Updated Code:**
```go
// Open item picker
if m.showItemPicker {
    // Render item picker
    return m.itemPicker.View()
}

// When user confirms item selection
if itemPickerConfirmed {
    attachments := m.itemPicker.GetSelectedItems()
    m.mailManager.SendMail(ctx, recipientID, subject, body, attachments)
}
```

### 2. Marketplace - Auction Creation
**File:** `internal/tui/marketplace.go:408`

**Current Code:**
```go
itemID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder
startingBid := 1000.0
duration := 24 * time.Hour
```

**Updated Code:**
```go
// Open item picker (single select)
picker := NewItemPicker(m.itemRepo, m.player.ID)
picker.SetMultiSelect(false)
picker.SetLocation(LocationShip) // Can only auction items in inventory

selectedItems := picker.GetSelectedItems()
if len(selectedItems) != 1 {
    return m, showError("Please select one item to auction")
}

itemID := selectedItems[0]
startingBid := m.auctionForm.StartingBid  // From form input
duration := m.auctionForm.Duration         // From form input

m.marketplaceManager.CreateAuction(ctx, m.player.ID, itemID, startingBid, duration)
```

### 3. Marketplace - Contract Creation
**File:** `internal/tui/marketplace.go:443`

**Current Code:**
```go
itemID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
reward := 5000.0
```

**Updated Code:**
```go
// For contracts requesting items (reverse auction)
// Player specifies what item they WANT, not what they HAVE
itemType := m.contractForm.ItemType     // From dropdown
equipmentID := m.contractForm.EquipmentID // From dropdown/search
reward := m.contractForm.Reward          // From form input

m.marketplaceManager.CreateContract(ctx, m.player.ID, itemType, equipmentID, reward)
```

### 4. Marketplace - Bounty Posting
**File:** `internal/tui/marketplace.go:477`

**Current Code:**
```go
targetPlayerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
reward := 10000.0
```

**Updated Code:**
```go
// Player search/select component
targetPlayerID := m.bountyForm.TargetPlayerID // From player search
reward := m.bountyForm.Reward                  // From form input
description := m.bountyForm.Description        // From text input

m.marketplaceManager.PostBounty(ctx, m.player.ID, targetPlayerID, reward, description)
```

## Migration Strategy

### Phase 1: Database & Models (Week 1)
- [ ] Create migration `019_inventory_system.sql`
- [ ] Add `PlayerItem` and `ItemTransfer` models
- [ ] Implement `ItemRepository` with full CRUD
- [ ] Write repository unit tests
- [ ] Migrate existing equipped items to `player_items` table

### Phase 2: UI Components (Week 2)
- [ ] Build `ItemPicker` component with multi-select
- [ ] Build `ItemList` component (read-only)
- [ ] Add filtering and search functionality
- [ ] Test components in isolation
- [ ] Add keyboard navigation (arrows, space, enter)

### Phase 3: Integration - Mail (Week 3)
- [ ] Integrate `ItemPicker` into mail compose screen
- [ ] Update `SendMail` to handle item attachments
- [ ] Update mail view to display attached items
- [ ] Add "Claim Items" button to received mail
- [ ] Test end-to-end mail with attachments

### Phase 4: Integration - Marketplace (Week 4)
- [ ] Build form components (price input, duration picker)
- [ ] Integrate `ItemPicker` into auction creation
- [ ] Update contract creation (item type/ID selection)
- [ ] Update bounty posting (player search)
- [ ] Test all marketplace creation flows

### Phase 5: Testing & Polish (Week 5)
- [ ] Integration tests for item transfers
- [ ] Load testing (1000+ items per player)
- [ ] UI polish (styling, animations)
- [ ] Documentation updates
- [ ] Migration testing (rollback scenarios)

## Testing Strategy

### Unit Tests
- `ItemRepository`: All CRUD operations
- `PlayerItem`: Property marshaling/unmarshaling
- Migration script (up/down)

### Integration Tests
- Item transfer atomicity (transaction rollback)
- Item picker component (selection, filtering)
- Mail with attachments (send, receive, claim)
- Auction creation and bidding

### Manual Testing Checklist
- [ ] Create item from equipment purchase
- [ ] Attach items to mail
- [ ] Receive mail and claim items
- [ ] Create auction with item
- [ ] Bid on and win auction
- [ ] Transfer items via player trade
- [ ] Verify audit log entries
- [ ] Test with 100+ items in inventory
- [ ] Test item picker search/filtering

## Performance Considerations

### Database Indexes
- `idx_player_items_player` - Fast lookup by player
- `idx_player_items_location` - Fast lookup by location
- `idx_player_items_type` - Fast filtering by type
- Composite index on `(player_id, location)` for common queries

### Caching Strategy
- Cache player inventory in memory (invalidate on update)
- Cache equipment definitions (static data)
- Paginate item lists for players with 1000+ items

### Query Optimization
- Use `LIMIT` for item pickers (default 100 items)
- Lazy load item properties (JSONB parsing on demand)
- Batch item transfers (single transaction)

## Security Considerations

### Validation
- ✅ Verify player owns item before transfer
- ✅ Verify destination location is valid
- ✅ Prevent duplicate item transfers (transaction isolation)
- ✅ Validate equipment_id exists in definitions

### Audit Trail
- ✅ Log all transfers in `item_transfers`
- ✅ Immutable audit log (no deletions)
- ✅ Track transfer type and reference ID

### Anti-Cheat
- ✅ Server-side validation of all item operations
- ✅ Cannot create items client-side
- ✅ Cannot modify item properties arbitrarily
- ✅ Rate limit item transfers (prevent spam)

## Future Enhancements (v2)

### Item Stacking
- Group identical items (e.g., 10x Laser Cannon)
- Quantity field on `PlayerItem`
- Split/merge stack operations

### Item Durability
- Add `durability` field (0-100)
- Degrade on use (combat, jumps)
- Repair mechanics

### Player Storage
- Station storage (rent lockers)
- Personal hangar (ship storage)
- Shared faction storage

### Item Crafting
- Consume items to create new items
- Blueprints and recipes
- Skill-based success rates

### Item Modifications
- Apply mods to items (e.g., +10% damage)
- Mod slots (limited per item)
- Removal/replacement mechanics

## Risks & Mitigations

### Risk: Data Migration Fails
**Mitigation:**
- Test migration on copy of production data
- Backup before migration
- Rollback script ready
- Gradual rollout (staging → production)

### Risk: Performance Degradation
**Mitigation:**
- Index all foreign keys
- Pagination for large inventories
- Cache frequently accessed data
- Load testing with 10k+ items

### Risk: UI Complexity
**Mitigation:**
- Start with simple picker (no search)
- Iterate based on user feedback
- Provide keyboard shortcuts
- Clear visual hierarchy

### Risk: Breaking Existing Features
**Mitigation:**
- Keep commodity cargo unchanged
- New tables only (no schema modifications)
- Comprehensive integration tests
- Feature flag for gradual rollout

## Success Metrics

### Technical Metrics
- Migration completes in < 5 minutes for 10k players
- Item picker renders in < 100ms
- Item transfer completes in < 50ms
- Zero data loss during migration
- 100% test coverage for repository

### User Experience Metrics
- Mail attachments sent successfully
- Auction creation flow completed
- Player feedback positive (survey)
- < 5 clicks to attach item to mail

## Conclusion

This hybrid inventory system provides a solid foundation for UUID-based item tracking while maintaining full backward compatibility with the existing commodity cargo system. The phased implementation allows for incremental delivery and testing, reducing risk and enabling early feedback.

**Recommended Timeline:** 5 weeks (4 weeks implementation + 1 week buffer)

**Next Steps:**
1. Review and approve specification
2. Create GitHub issues for each phase
3. Begin Phase 1 (database & models)
4. Schedule weekly progress reviews

---

**Document Status:** Ready for Review
**Estimated Effort:** 4-5 weeks (1 developer)
**Priority:** High (blocks 4 features)
**Complexity:** Medium-High
