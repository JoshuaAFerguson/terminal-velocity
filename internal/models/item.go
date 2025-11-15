// File: internal/models/item.go
// Project: Terminal Velocity
// Description: Player inventory item models for UUID-based equipment tracking
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// Valid returns true if the ItemType is valid
func (it ItemType) Valid() bool {
	switch it {
	case ItemTypeWeapon, ItemTypeOutfit, ItemTypeSpecial, ItemTypeQuest:
		return true
	default:
		return false
	}
}

// String returns the string representation of ItemType
func (it ItemType) String() string {
	return string(it)
}

// ItemLocation represents where an item is currently stored
type ItemLocation string

const (
	LocationShip           ItemLocation = "ship"
	LocationStationStorage ItemLocation = "station_storage"
	LocationMail           ItemLocation = "mail"
	LocationEscrow         ItemLocation = "escrow"
	LocationAuction        ItemLocation = "auction"
)

// Valid returns true if the ItemLocation is valid
func (il ItemLocation) Valid() bool {
	switch il {
	case LocationShip, LocationStationStorage, LocationMail, LocationEscrow, LocationAuction:
		return true
	default:
		return false
	}
}

// String returns the string representation of ItemLocation
func (il ItemLocation) String() string {
	return string(il)
}

// PlayerItem represents a single inventory item with UUID
type PlayerItem struct {
	ID          uuid.UUID       `json:"id"`
	PlayerID    uuid.UUID       `json:"player_id"`
	ItemType    ItemType        `json:"item_type"`
	EquipmentID string          `json:"equipment_id"` // References equipment definition
	Location    ItemLocation    `json:"location"`
	LocationID  *uuid.UUID      `json:"location_id,omitempty"`
	Properties  json.RawMessage `json:"properties"` // JSONB for mods/upgrades
	AcquiredAt  time.Time       `json:"acquired_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ItemTransfer represents a transfer audit entry
type ItemTransfer struct {
	ID            uuid.UUID  `json:"id"`
	ItemID        uuid.UUID  `json:"item_id"`
	FromPlayerID  *uuid.UUID `json:"from_player_id,omitempty"`
	ToPlayerID    *uuid.UUID `json:"to_player_id,omitempty"`
	TransferType  string     `json:"transfer_type"` // trade, mail, auction, etc.
	TransferID    *uuid.UUID `json:"transfer_id,omitempty"`
	TransferredAt time.Time  `json:"transferred_at"`
}

// ItemProperties represents modifiable item properties
type ItemProperties struct {
	Mods       []string               `json:"mods,omitempty"`       // Applied modifications
	Upgrades   map[string]int         `json:"upgrades,omitempty"`   // Upgrade levels (e.g., "damage": 3)
	CustomData map[string]interface{} `json:"custom,omitempty"`     // Extension point for future features
	Quantity   int                    `json:"quantity,omitempty"`   // For stackable items (future)
	Durability int                    `json:"durability,omitempty"` // Condition 0-100 (future)
}

// GetProperties unmarshals JSONB properties
func (i *PlayerItem) GetProperties() (*ItemProperties, error) {
	if len(i.Properties) == 0 {
		return &ItemProperties{}, nil
	}

	var props ItemProperties
	if err := json.Unmarshal(i.Properties, &props); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item properties: %w", err)
	}
	return &props, nil
}

// SetProperties marshals and sets JSONB properties
func (i *PlayerItem) SetProperties(props *ItemProperties) error {
	data, err := json.Marshal(props)
	if err != nil {
		return fmt.Errorf("failed to marshal item properties: %w", err)
	}
	i.Properties = data
	return nil
}

// GetEquipmentName returns human-readable equipment name
// Converts "laser_cannon_mk2" to "Laser Cannon Mk2"
func (i *PlayerItem) GetEquipmentName() string {
	return formatEquipmentID(i.EquipmentID)
}

// GetDisplayName returns a formatted display name with item type prefix
func (i *PlayerItem) GetDisplayName() string {
	name := i.GetEquipmentName()

	// Add type prefix for clarity
	switch i.ItemType {
	case ItemTypeWeapon:
		return fmt.Sprintf("[W] %s", name)
	case ItemTypeOutfit:
		return fmt.Sprintf("[O] %s", name)
	case ItemTypeSpecial:
		return fmt.Sprintf("[S] %s", name)
	case ItemTypeQuest:
		return fmt.Sprintf("[Q] %s", name)
	default:
		return name
	}
}

// GetLocationName returns human-readable location name
func (i *PlayerItem) GetLocationName() string {
	switch i.Location {
	case LocationShip:
		return "Ship"
	case LocationStationStorage:
		return "Station Storage"
	case LocationMail:
		return "In Mail"
	case LocationEscrow:
		return "In Trade"
	case LocationAuction:
		return "On Auction"
	default:
		return string(i.Location)
	}
}

// IsAvailable returns true if the item can be used/traded
// Items in mail, escrow, or auction are not available
func (i *PlayerItem) IsAvailable() bool {
	return i.Location == LocationShip || i.Location == LocationStationStorage
}

// CanAttachToMail returns true if the item can be attached to mail
// Only available items can be attached
func (i *PlayerItem) CanAttachToMail() bool {
	return i.IsAvailable()
}

// CanAuction returns true if the item can be put up for auction
// Only available items can be auctioned
func (i *PlayerItem) CanAuction() bool {
	return i.IsAvailable()
}

// Validate validates the PlayerItem fields
func (i *PlayerItem) Validate() error {
	if i.PlayerID == uuid.Nil {
		return fmt.Errorf("player_id cannot be nil")
	}

	if !i.ItemType.Valid() {
		return fmt.Errorf("invalid item_type: %s", i.ItemType)
	}

	if i.EquipmentID == "" {
		return fmt.Errorf("equipment_id cannot be empty")
	}

	if !i.Location.Valid() {
		return fmt.Errorf("invalid location: %s", i.Location)
	}

	// Some locations require a location_id
	switch i.Location {
	case LocationShip, LocationStationStorage:
		if i.LocationID == nil {
			return fmt.Errorf("location_id required for location: %s", i.Location)
		}
	}

	return nil
}

// formatEquipmentID converts "laser_cannon_mk2" to "Laser Cannon Mk2"
func formatEquipmentID(id string) string {
	// Replace underscores with spaces
	name := strings.ReplaceAll(id, "_", " ")

	// Title case each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			// Special case for "mk" (mark) - keep uppercase
			if strings.ToLower(word) == "mk" || strings.HasPrefix(strings.ToLower(word), "mk") {
				words[i] = "Mk" + word[2:]
			} else {
				words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
			}
		}
	}

	return strings.Join(words, " ")
}

// TransferTypeValid returns true if the transfer type is valid
func TransferTypeValid(transferType string) bool {
	switch transferType {
	case "trade", "mail", "auction", "contract", "admin":
		return true
	default:
		return false
	}
}

// GetTransferTypeName returns a human-readable transfer type name
func GetTransferTypeName(transferType string) string {
	switch transferType {
	case "trade":
		return "Player Trade"
	case "mail":
		return "Mail Attachment"
	case "auction":
		return "Auction Purchase"
	case "contract":
		return "Contract Fulfillment"
	case "admin":
		return "Admin Transfer"
	default:
		return transferType
	}
}
