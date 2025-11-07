// File: internal/models/outfitting.go
// Project: Terminal Velocity
// Description: Ship customization and outfitting system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// EquipmentSlotType represents different types of equipment slots
type EquipmentSlotType string

const (
	SlotWeapon    EquipmentSlotType = "weapon"     // Gun/missile hardpoints
	SlotShield    EquipmentSlotType = "shield"     // Shield generator
	SlotEngine    EquipmentSlotType = "engine"     // Propulsion system
	SlotReactor   EquipmentSlotType = "reactor"    // Power generation
	SlotUtility   EquipmentSlotType = "utility"    // Cargo pods, fuel tanks, etc.
	SlotSpecial   EquipmentSlotType = "special"    // Cloaking, special weapons
)

// EquipmentSlot represents a single equipment slot on a ship
type EquipmentSlot struct {
	ID            uuid.UUID         `json:"id"`
	SlotType      EquipmentSlotType `json:"slot_type"`
	SlotSize      int               `json:"slot_size"`      // 1=small, 2=medium, 3=large, 4=capital
	InstalledItem *Equipment        `json:"installed_item,omitempty"`
	Locked        bool              `json:"locked"`         // Cannot be modified
}

// Equipment represents any installable ship equipment
type Equipment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    EquipmentCategory `json:"category"`
	SlotType    EquipmentSlotType `json:"slot_type"`
	SlotSize    int               `json:"slot_size"`

	// Requirements
	MinTechLevel    int   `json:"min_tech_level"`
	RequiredLicense string `json:"required_license,omitempty"`

	// Cost
	Price       int64 `json:"price"`
	OutfitSpace int   `json:"outfit_space"`

	// Stats (varies by type)
	Stats EquipmentStats `json:"stats"`

	// Availability
	Rarity      string `json:"rarity"` // common, uncommon, rare, military, experimental
	Faction     string `json:"faction,omitempty"` // Faction-specific equipment
}

// EquipmentCategory defines broad equipment types
type EquipmentCategory string

const (
	CategoryWeapon    EquipmentCategory = "weapon"
	CategoryDefense   EquipmentCategory = "defense"
	CategoryPower     EquipmentCategory = "power"
	CategoryPropulsion EquipmentCategory = "propulsion"
	CategoryUtility   EquipmentCategory = "utility"
	CategorySpecial   EquipmentCategory = "special"
)

// EquipmentStats holds all possible equipment statistics
type EquipmentStats struct {
	// Weapon stats
	Damage            int     `json:"damage,omitempty"`
	Range             int     `json:"range,omitempty"`
	Accuracy          int     `json:"accuracy,omitempty"`
	Cooldown          float64 `json:"cooldown,omitempty"`
	EnergyCost        int     `json:"energy_cost,omitempty"`
	AmmoCapacity      int     `json:"ammo_capacity,omitempty"`
	ShieldPenetration float64 `json:"shield_penetration,omitempty"`

	// Defense stats
	ShieldHP      int     `json:"shield_hp,omitempty"`
	ShieldRegen   int     `json:"shield_regen,omitempty"`
	HullBonus     int     `json:"hull_bonus,omitempty"`
	ArmorRating   int     `json:"armor_rating,omitempty"`

	// Power stats
	EnergyOutput  int `json:"energy_output,omitempty"`
	EnergyStorage int `json:"energy_storage,omitempty"`

	// Propulsion stats
	SpeedBonus    int `json:"speed_bonus,omitempty"`
	TurnRate      int `json:"turn_rate,omitempty"`
	AfterburnerBoost int `json:"afterburner_boost,omitempty"`

	// Utility stats
	CargoBonus    int `json:"cargo_bonus,omitempty"`
	FuelBonus     int `json:"fuel_bonus,omitempty"`
	ScannerRange  int `json:"scanner_range,omitempty"`
	JumpRange     int `json:"jump_range,omitempty"`

	// Special stats
	CloakingPower int     `json:"cloaking_power,omitempty"`
	ECMStrength   int     `json:"ecm_strength,omitempty"` // Electronic countermeasures
	RepairRate    int     `json:"repair_rate,omitempty"`
}

// ShipLoadout represents a saved ship configuration
type ShipLoadout struct {
	ID          uuid.UUID          `json:"id"`
	PlayerID    uuid.UUID          `json:"player_id"`
	ShipTypeID  string             `json:"ship_type_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`

	// Equipment configuration
	Slots       []EquipmentSlot    `json:"slots"`

	// Metadata
	TotalCost   int64              `json:"total_cost"`   // Sum of all equipment prices
	UsedOutfitSpace int            `json:"used_outfit_space"`
}

// OutfitPurchase represents a transaction for buying equipment
type OutfitPurchase struct {
	ID          uuid.UUID `json:"id"`
	PlayerID    uuid.UUID `json:"player_id"`
	EquipmentID string    `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	TotalCost   int64     `json:"total_cost"`
	StationID   uuid.UUID `json:"station_id"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// OutfitInstallation represents installing equipment on a ship
type OutfitInstallation struct {
	ID          uuid.UUID `json:"id"`
	ShipID      uuid.UUID `json:"ship_id"`
	SlotID      uuid.UUID `json:"slot_id"`
	EquipmentID string    `json:"equipment_id"`
	InstalledAt time.Time `json:"installed_at"`
	InstalledBy uuid.UUID `json:"installed_by"` // Player or NPC mechanic
}

// NewEquipmentSlot creates a new equipment slot
func NewEquipmentSlot(slotType EquipmentSlotType, slotSize int) EquipmentSlot {
	return EquipmentSlot{
		ID:       uuid.New(),
		SlotType: slotType,
		SlotSize: slotSize,
		Locked:   false,
	}
}

// CanInstall checks if equipment can be installed in this slot
func (slot *EquipmentSlot) CanInstall(equipment *Equipment) bool {
	if slot.Locked {
		return false
	}
	if slot.SlotType != equipment.SlotType {
		return false
	}
	if slot.SlotSize < equipment.SlotSize {
		return false
	}
	return true
}

// Install installs equipment in this slot
func (slot *EquipmentSlot) Install(equipment *Equipment) bool {
	if !slot.CanInstall(equipment) {
		return false
	}
	slot.InstalledItem = equipment
	return true
}

// Uninstall removes equipment from this slot
func (slot *EquipmentSlot) Uninstall() *Equipment {
	if slot.Locked {
		return nil
	}
	removed := slot.InstalledItem
	slot.InstalledItem = nil
	return removed
}

// IsEmpty checks if slot has no equipment
func (slot *EquipmentSlot) IsEmpty() bool {
	return slot.InstalledItem == nil
}

// NewShipLoadout creates a new ship loadout
func NewShipLoadout(playerID uuid.UUID, shipTypeID string, name string) *ShipLoadout {
	return &ShipLoadout{
		ID:         uuid.New(),
		PlayerID:   playerID,
		ShipTypeID: shipTypeID,
		Name:       name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Slots:      []EquipmentSlot{},
		TotalCost:  0,
		UsedOutfitSpace: 0,
	}
}

// AddSlot adds an equipment slot to the loadout
func (loadout *ShipLoadout) AddSlot(slot EquipmentSlot) {
	loadout.Slots = append(loadout.Slots, slot)
	loadout.UpdatedAt = time.Now()
}

// GetSlotByID finds a slot by ID
func (loadout *ShipLoadout) GetSlotByID(slotID uuid.UUID) *EquipmentSlot {
	for i := range loadout.Slots {
		if loadout.Slots[i].ID == slotID {
			return &loadout.Slots[i]
		}
	}
	return nil
}

// GetSlotsByType returns all slots of a specific type
func (loadout *ShipLoadout) GetSlotsByType(slotType EquipmentSlotType) []EquipmentSlot {
	var slots []EquipmentSlot
	for _, slot := range loadout.Slots {
		if slot.SlotType == slotType {
			slots = append(slots, slot)
		}
	}
	return slots
}

// GetInstalledEquipment returns all installed equipment
func (loadout *ShipLoadout) GetInstalledEquipment() []*Equipment {
	var equipment []*Equipment
	for _, slot := range loadout.Slots {
		if !slot.IsEmpty() {
			equipment = append(equipment, slot.InstalledItem)
		}
	}
	return equipment
}

// CalculateTotalCost calculates total cost of all equipment
func (loadout *ShipLoadout) CalculateTotalCost() int64 {
	total := int64(0)
	for _, slot := range loadout.Slots {
		if !slot.IsEmpty() {
			total += slot.InstalledItem.Price
		}
	}
	loadout.TotalCost = total
	return total
}

// CalculateOutfitSpace calculates total outfit space used
func (loadout *ShipLoadout) CalculateOutfitSpace() int {
	total := 0
	for _, slot := range loadout.Slots {
		if !slot.IsEmpty() {
			total += slot.InstalledItem.OutfitSpace
		}
	}
	loadout.UsedOutfitSpace = total
	return total
}

// GetCombinedStats calculates combined stats from all equipment
func (loadout *ShipLoadout) GetCombinedStats() EquipmentStats {
	combined := EquipmentStats{}

	for _, slot := range loadout.Slots {
		if slot.IsEmpty() {
			continue
		}

		stats := slot.InstalledItem.Stats

		// Sum all applicable stats
		combined.Damage += stats.Damage
		combined.ShieldHP += stats.ShieldHP
		combined.ShieldRegen += stats.ShieldRegen
		combined.HullBonus += stats.HullBonus
		combined.ArmorRating += stats.ArmorRating
		combined.EnergyOutput += stats.EnergyOutput
		combined.EnergyStorage += stats.EnergyStorage
		combined.SpeedBonus += stats.SpeedBonus
		combined.TurnRate += stats.TurnRate
		combined.AfterburnerBoost += stats.AfterburnerBoost
		combined.CargoBonus += stats.CargoBonus
		combined.FuelBonus += stats.FuelBonus
		combined.ScannerRange += stats.ScannerRange
		combined.JumpRange += stats.JumpRange
		combined.CloakingPower += stats.CloakingPower
		combined.ECMStrength += stats.ECMStrength
		combined.RepairRate += stats.RepairRate
	}

	return combined
}

// Clone creates a copy of this loadout
func (loadout *ShipLoadout) Clone(newName string) *ShipLoadout {
	clone := &ShipLoadout{
		ID:         uuid.New(),
		PlayerID:   loadout.PlayerID,
		ShipTypeID: loadout.ShipTypeID,
		Name:       newName,
		Description: "Clone of " + loadout.Name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Slots:      make([]EquipmentSlot, len(loadout.Slots)),
		TotalCost:  loadout.TotalCost,
		UsedOutfitSpace: loadout.UsedOutfitSpace,
	}

	// Deep copy slots
	for i, slot := range loadout.Slots {
		clone.Slots[i] = EquipmentSlot{
			ID:       uuid.New(),
			SlotType: slot.SlotType,
			SlotSize: slot.SlotSize,
			Locked:   slot.Locked,
		}
		if !slot.IsEmpty() {
			// Copy equipment reference (not deep copy of equipment itself)
			clone.Slots[i].InstalledItem = slot.InstalledItem
		}
	}

	return clone
}

// Validate checks if loadout is valid for a ship type
func (loadout *ShipLoadout) Validate(shipType *ShipType) []string {
	errors := []string{}

	// Check outfit space
	usedSpace := loadout.CalculateOutfitSpace()
	if usedSpace > shipType.OutfitSpace {
		errors = append(errors, "Outfit space exceeded")
	}

	// Check weapon slots
	weaponSlots := loadout.GetSlotsByType(SlotWeapon)
	if len(weaponSlots) > shipType.WeaponSlots {
		errors = append(errors, "Too many weapon slots")
	}

	// Check for required systems
	hasReactor := false
	hasEngine := false
	for _, slot := range loadout.Slots {
		if !slot.IsEmpty() {
			if slot.SlotType == SlotReactor {
				hasReactor = true
			}
			if slot.SlotType == SlotEngine {
				hasEngine = true
			}
		}
	}

	if !hasReactor {
		errors = append(errors, "No reactor installed")
	}
	if !hasEngine {
		errors = append(errors, "No engine installed")
	}

	return errors
}
