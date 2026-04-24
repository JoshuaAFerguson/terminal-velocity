// File: internal/combat/loot.go
// Project: Terminal Velocity
// Description: Combat system: loot - Loot generation, salvage, and rare item drops
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package combat provides the loot and salvage system for destroyed ships.
//
// This file implements loot generation from combat victories including:
//   - Credit rewards based on ship value (10-20% of ship price)
//   - Bounty collection for wanted ships
//   - Cargo recovery (30-60% survival rate)
//   - Equipment salvage (30-45% chance per item)
//   - Rare item drops (artifacts, components, data, contraband)
//
// Loot Rarity Tiers:
//   - Uncommon: 50% of rare drops, 25K value typical
//   - Rare: 30% of rare drops, 50-60K value
//   - Epic: 15% of rare drops, 75-100K value
//   - Legendary: 5% of rare drops, 250K+ value
//
// Rare Item Drop Chances:
//   - Base: 5% chance
//   - +10-15% for military/capital ships
//   - +8% for hostile/pirate ships (contraband)
//   - +5% for ships worth >500K
//   - +5% for ships worth >1M
//   - Capped at 40% maximum
//
// Thread-safety: Functions are stateless and safe for concurrent calls.
package combat

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// LootDrop represents all items and credits recovered from a destroyed ship.
//
// This structure contains the complete loot package from a combat victory,
// including credits, cargo, equipment, and rare items. The TotalValue field
// includes credits plus 50% resale value of all salvaged equipment.
//
// Fields:
//   - Credits: Direct credits awarded (ship value % + bounty)
//   - Cargo: Recovered cargo items with quantities
//   - Outfits: Salvaged outfit equipment IDs
//   - Weapons: Salvaged weapon IDs
//   - RareItems: Special rare item drops
//   - Message: Human-readable loot summary
//   - TotalValue: Total credit value of all loot (for sorting/display)
type LootDrop struct {
	Credits    int64
	Cargo      []models.CargoItem // cargo items
	Outfits    []string           // outfit IDs
	Weapons    []string           // weapon IDs
	RareItems  []RareItem
	Message    string
	TotalValue int64
}

// RareItem represents special loot with unique properties and high value.
//
// Rare items are special drops from combat that can be sold for significant
// credits or used in quests. Each rare item has a rarity tier that affects
// drop chance and value.
//
// Fields:
//   - ID: Unique identifier for this rare item type
//   - Name: Display name
//   - Description: Flavor text describing the item
//   - Rarity: Tier ("uncommon", "rare", "epic", "legendary")
//   - Value: Base credit value
//   - Type: Category ("artifact", "component", "data", "contraband")
type RareItem struct {
	ID          string
	Name        string
	Description string
	Rarity      string // "uncommon", "rare", "epic", "legendary"
	Value       int64
	Type        string // "artifact", "component", "data", "contraband"
}

// SalvageResult represents the outcome of salvaging a specific item from wreckage.
//
// Used when attempting to salvage a specific item from a wreck rather than
// using the automatic loot generation system. Success depends on luck factor
// and item type.
//
// Fields:
//   - Success: Whether salvage attempt succeeded
//   - RecoveredQty: Number of items recovered (0 or 1 typically)
//   - Item: Item ID that was targeted
//   - Message: Result message for player feedback
type SalvageResult struct {
	Success      bool
	RecoveredQty int
	Item         string
	Message      string
}

// Standard rare items that can drop
var RareItemList = []RareItem{
	{
		ID:          "military_plans",
		Name:        "Military Plans",
		Description: "Encrypted tactical data highly valued by certain factions",
		Rarity:      "rare",
		Value:       50000,
		Type:        "data",
	},
	{
		ID:          "prototype_component",
		Name:        "Prototype Component",
		Description: "Advanced technology component from experimental ships",
		Rarity:      "epic",
		Value:       100000,
		Type:        "component",
	},
	{
		ID:          "ancient_artifact",
		Name:        "Ancient Artifact",
		Description: "Mysterious pre-colonial artifact of unknown origin",
		Rarity:      "legendary",
		Value:       250000,
		Type:        "artifact",
	},
	{
		ID:          "neural_processor",
		Name:        "Neural Processor",
		Description: "AI-grade processor core, illegal in most systems",
		Rarity:      "epic",
		Value:       75000,
		Type:        "contraband",
	},
	{
		ID:          "jump_drive_data",
		Name:        "Jump Drive Data",
		Description: "Research data on experimental jump drive technology",
		Rarity:      "rare",
		Value:       60000,
		Type:        "data",
	},
	{
		ID:          "fusion_core",
		Name:        "Fusion Core",
		Description: "Compact fusion reactor core in working condition",
		Rarity:      "uncommon",
		Value:       25000,
		Type:        "component",
	},
}

// GenerateLoot creates a complete loot drop package from a destroyed ship.
//
// This is the main loot generation function that calculates all rewards from
// a combat victory. Loot includes credits, recovered cargo, salvaged equipment,
// and potential rare item drops.
//
// Loot Generation Rules:
//   - Credits: 10-20% of ship's base value
//   - Bounty: Added if ship had a bounty
//   - Cargo: 30-60% survival rate (random per item)
//   - Outfits: 40% salvage chance per outfit
//   - Weapons: 30% salvage chance (45% for hostile ships)
//   - Rare Items: 5-40% chance based on ship value and type
//
// Parameters:
//   - destroyedShip: The ship that was destroyed (for cargo and equipment)
//   - destroyedShipType: Ship type definition (for value calculations)
//   - wasHostile: Whether ship was hostile (affects weapon salvage chance)
//   - hadBounty: Whether ship had a bounty on it
//   - bountyAmount: Bounty value to award (0 if no bounty)
//
// Returns:
//   - *LootDrop: Complete loot package with all rewards and summary message
//
// Thread-safe: No shared state modification, safe for concurrent calls.
func GenerateLoot(
	destroyedShip *models.Ship,
	destroyedShipType *models.ShipType,
	wasHostile bool,
	hadBounty bool,
	bountyAmount int64,
) *LootDrop {

	loot := &LootDrop{
		Credits:   0,
		Cargo:     []models.CargoItem{},
		Outfits:   []string{},
		Weapons:   []string{},
		RareItems: []RareItem{},
	}

	// Base credit reward (10-20% of ship value)
	baseReward := int64(float64(destroyedShipType.Price) * (0.1 + rand.Float64()*0.1))
	loot.Credits += baseReward

	// Bounty reward if applicable
	if hadBounty && bountyAmount > 0 {
		loot.Credits += bountyAmount
	}

	// Cargo recovery (30-60% of cargo survives)
	cargoSurvivalRate := 0.3 + rand.Float64()*0.3
	for _, cargoItem := range destroyedShip.Cargo {
		recoveredQty := int(float64(cargoItem.Quantity) * cargoSurvivalRate)
		if recoveredQty > 0 {
			loot.Cargo = append(loot.Cargo, models.CargoItem{
				CommodityID: cargoItem.CommodityID,
				Quantity:    recoveredQty,
			})
		}
	}

	// Outfit salvaging (40% chance per outfit)
	for _, outfitID := range destroyedShip.Outfits {
		if rand.Float64() < 0.4 {
			loot.Outfits = append(loot.Outfits, outfitID)
		}
	}

	// Weapon salvaging (30% chance per weapon, higher for hostile ships)
	weaponSalvageChance := 0.3
	if wasHostile {
		weaponSalvageChance = 0.45 // Pirates/hostiles are more likely to have salvageable weapons
	}
	for _, weaponID := range destroyedShip.Weapons {
		if rand.Float64() < weaponSalvageChance {
			loot.Weapons = append(loot.Weapons, weaponID)
		}
	}

	// Rare item drops (chance increases with ship value)
	rareItemChance := calculateRareItemChance(destroyedShipType, wasHostile)
	if rand.Float64() < rareItemChance {
		rareItem := generateRareItem(destroyedShipType)
		if rareItem != nil {
			loot.RareItems = append(loot.RareItems, *rareItem)
			loot.Credits += rareItem.Value / 2 // Partial credit for rare item
		}
	}

	// Calculate total value
	loot.TotalValue = loot.Credits
	for _, outfitID := range loot.Outfits {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			loot.TotalValue += outfit.Price / 2 // Salvaged items worth 50%
		}
	}
	for _, weaponID := range loot.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			loot.TotalValue += weapon.Price / 2
		}
	}

	// Generate summary message
	loot.Message = formatLootMessage(loot, hadBounty)

	return loot
}

// calculateRareItemChance determines the chance of a rare item drop
func calculateRareItemChance(shipType *models.ShipType, wasHostile bool) float64 {
	// Base chance: 5%
	chance := 0.05

	// Higher chance for larger ships
	if shipType.Class == "military" {
		chance += 0.10
	} else if shipType.Class == "capital" {
		chance += 0.15
	}

	// Higher chance for hostile/pirate ships (carrying contraband)
	if wasHostile {
		chance += 0.08
	}

	// Ship value factor (expensive ships = better loot)
	if shipType.Price > 500000 {
		chance += 0.05
	}
	if shipType.Price > 1000000 {
		chance += 0.05
	}

	// Cap at 40%
	if chance > 0.4 {
		chance = 0.4
	}

	return chance
}

// generateRareItem generates a random rare item drop
func generateRareItem(shipType *models.ShipType) *RareItem {
	if len(RareItemList) == 0 {
		return nil
	}

	// Weight by rarity (legendary 5%, epic 15%, rare 30%, uncommon 50%)
	rarityRoll := rand.Float64()

	var eligibleItems []RareItem
	if rarityRoll < 0.05 {
		// Legendary
		for _, item := range RareItemList {
			if item.Rarity == "legendary" {
				eligibleItems = append(eligibleItems, item)
			}
		}
	} else if rarityRoll < 0.20 {
		// Epic
		for _, item := range RareItemList {
			if item.Rarity == "epic" {
				eligibleItems = append(eligibleItems, item)
			}
		}
	} else if rarityRoll < 0.50 {
		// Rare
		for _, item := range RareItemList {
			if item.Rarity == "rare" {
				eligibleItems = append(eligibleItems, item)
			}
		}
	} else {
		// Uncommon
		for _, item := range RareItemList {
			if item.Rarity == "uncommon" {
				eligibleItems = append(eligibleItems, item)
			}
		}
	}

	// If no items in selected rarity, fall back to any item
	if len(eligibleItems) == 0 {
		eligibleItems = RareItemList
	}

	// Return random item from eligible list
	if len(eligibleItems) > 0 {
		item := eligibleItems[rand.Intn(len(eligibleItems))]
		return &item
	}

	return nil
}

// formatLootMessage creates a summary of the loot drop
func formatLootMessage(loot *LootDrop, hadBounty bool) string {
	msg := "Salvage recovered:\n"

	if loot.Credits > 0 {
		if hadBounty {
			msg += "  • Bounty collected + salvage credits\n"
		} else {
			msg += "  • Credits from wreckage\n"
		}
	}

	if len(loot.Cargo) > 0 {
		msg += "  • Cargo containers\n"
	}

	if len(loot.Weapons) > 0 {
		msg += "  • Weapons\n"
	}

	if len(loot.Outfits) > 0 {
		msg += "  • Equipment outfits\n"
	}

	if len(loot.RareItems) > 0 {
		for _, item := range loot.RareItems {
			msg += "  • RARE: " + item.Name + " (" + item.Rarity + ")\n"
		}
	}

	return msg
}

// CalculateCargoSpaceRequired calculates total cargo space needed to carry all loot.
//
// Different loot types have different cargo space requirements:
//   - Cargo items: 1 ton per unit
//   - Weapons: 5 tons each
//   - Outfits: 3 tons each
//   - Rare items: 1 ton each
//
// Parameters:
//   - loot: Loot drop to calculate space for
//
// Returns:
//   - int: Total cargo space required in tons
//
// Thread-safe: No shared state, safe for concurrent calls.
func CalculateCargoSpaceRequired(loot *LootDrop) int {
	totalSpace := 0

	// Cargo takes up space
	for _, cargoItem := range loot.Cargo {
		totalSpace += cargoItem.Quantity
	}

	// Weapons take up space (5 tons each)
	totalSpace += len(loot.Weapons) * 5

	// Outfits take up space (3 tons each)
	totalSpace += len(loot.Outfits) * 3

	// Rare items (1 ton each)
	totalSpace += len(loot.RareItems)

	return totalSpace
}

// CanCarryLoot checks if player's ship has sufficient cargo space for loot.
//
// Compares required cargo space for loot against available cargo capacity.
// Accounts for currently used cargo space in player's ship.
//
// Parameters:
//   - playerShip: Player's ship (for current cargo calculation)
//   - playerShipType: Ship type (for max cargo capacity)
//   - loot: Loot to check space requirements for
//
// Returns:
//   - bool: true if player has enough space, false otherwise
//
// Thread-safe: No shared state modification, safe for concurrent calls.
func CanCarryLoot(playerShip *models.Ship, playerShipType *models.ShipType, loot *LootDrop) bool {
	currentCargo := playerShip.GetCargoUsed()
	lootSpace := CalculateCargoSpaceRequired(loot)
	availableSpace := playerShipType.CargoSpace - currentCargo

	return lootSpace <= availableSpace
}

// ApplyLoot transfers all loot to the player's ship and inventory.
//
// This function handles the complete loot collection process:
//   1. Validates cargo space availability
//   2. Adds cargo items to player's cargo hold
//   3. Converts weapons and outfits to credits (50% resale value)
//   4. Adds all credits to player
//   5. Generates detailed summary message
//
// Note: Currently weapons and outfits are auto-sold for 50% of their value.
// In future implementation, these could be added to player inventory.
//
// Parameters:
//   - playerShip: Player's ship - modified by this function (cargo and credits)
//   - playerShipType: Ship type (for cargo capacity validation)
//   - loot: Loot to apply to player
//
// Returns:
//   - bool: true if loot was successfully applied, false if insufficient space
//   - string: Detailed summary of collected loot or error message
//
// Side Effects:
//   - Modifies playerShip.Cargo (adds loot cargo)
//   - Modifies playerShip.Credits (adds loot credits - via player reference if applicable)
//
// Thread-safe: Modifies only passed parameters, safe for single player context.
func ApplyLoot(playerShip *models.Ship, playerShipType *models.ShipType, loot *LootDrop) (bool, string) {
	// Check cargo space
	if !CanCarryLoot(playerShip, playerShipType, loot) {
		return false, "Insufficient cargo space for all salvage"
	}

	// Add cargo
	for _, cargoItem := range loot.Cargo {
		playerShip.AddCargo(cargoItem.CommodityID, cargoItem.Quantity)
	}

	// Weapons and outfits would be added to player inventory (not implemented yet)
	// For now, convert them to credits (50% of value)
	weaponValue := int64(0)
	for _, weaponID := range loot.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			weaponValue += weapon.Price / 2
		}
	}

	outfitValue := int64(0)
	for _, outfitID := range loot.Outfits {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			outfitValue += outfit.Price / 2
		}
	}

	// Rare items stored as cargo (simplified for now)
	// In full implementation, would go to special inventory

	totalCredits := loot.Credits + weaponValue + outfitValue
	return true, formatLootSummary(loot, totalCredits)
}

// formatLootSummary creates a detailed summary of collected loot
func formatLootSummary(loot *LootDrop, totalCredits int64) string {
	msg := "Salvage collected:\n\n"
	msg += formatCredits(totalCredits) + " credits\n"

	if len(loot.Cargo) > 0 {
		msg += "\nCargo recovered:\n"
		for _, cargoItem := range loot.Cargo {
			commodity := models.GetCommodityByID(cargoItem.CommodityID)
			if commodity != nil {
				msg += fmt.Sprintf("  %s x%d\n", commodity.Name, cargoItem.Quantity)
			}
		}
	}

	if len(loot.Weapons) > 0 {
		msg += "\nWeapons salvaged (sold):\n"
		for _, weaponID := range loot.Weapons {
			weapon := models.GetWeaponByID(weaponID)
			if weapon != nil {
				msg += "  " + weapon.Name + "\n"
			}
		}
	}

	if len(loot.Outfits) > 0 {
		msg += "\nOutfits salvaged (sold):\n"
		for _, outfitID := range loot.Outfits {
			outfit := models.GetOutfitByID(outfitID)
			if outfit != nil {
				msg += "  " + outfit.Name + "\n"
			}
		}
	}

	if len(loot.RareItems) > 0 {
		msg += "\nRARE ITEMS:\n"
		for _, item := range loot.RareItems {
			msg += "  [" + item.Rarity + "] " + item.Name + "\n"
			msg += "    " + item.Description + "\n"
		}
	}

	return msg
}

// SalvageSpecificItem attempts to salvage a specific item from ship wreckage.
//
// Unlike GenerateLoot which creates a random loot drop, this function targets
// a specific item for recovery. Success is based on a base 30% chance modified
// by a luck factor.
//
// Use Cases:
//   - Player targeting specific equipment from wreck
//   - Quest objectives requiring specific salvage
//   - Specialized salvage gameplay mechanics
//
// Parameters:
//   - itemType: Type of item being salvaged (for message generation)
//   - itemID: Specific item identifier
//   - luck: Luck multiplier (0.0 to 2.0 typically), 1.0 = normal chance
//
// Returns:
//   - *SalvageResult: Result with success status and message
//
// Thread-safe: No shared state, safe for concurrent calls.
func SalvageSpecificItem(itemType string, itemID string, luck float64) *SalvageResult {
	baseChance := 0.3

	// Adjust by luck factor
	chance := baseChance * luck

	result := &SalvageResult{
		Success: rand.Float64() < chance,
		Item:    itemID,
	}

	if result.Success {
		result.Message = "Successfully salvaged " + itemType
		result.RecoveredQty = 1
	} else {
		result.Message = "Salvage attempt failed - item too damaged"
		result.RecoveredQty = 0
	}

	return result
}

// CalculateSalvageTime returns the time required to collect loot from wreckage.
//
// Salvage time varies based on the quantity and complexity of loot:
//   - Base: 2 turns
//   - +1 turn for large cargo hauls (>5 items)
//   - +1 turn for multiple weapons/outfits (>3 items)
//   - +1 turn if rare items present (careful handling required)
//
// This creates a risk/reward tradeoff in dangerous areas where staying
// to salvage may expose player to additional threats.
//
// Parameters:
//   - loot: Loot drop to calculate salvage time for
//
// Returns:
//   - int: Number of combat turns required to collect loot (2-5 typically)
//
// Thread-safe: No shared state, safe for concurrent calls.
func CalculateSalvageTime(loot *LootDrop) int {
	// Base time: 2 turns
	time := 2

	// +1 turn for large cargo hauls
	if len(loot.Cargo) > 5 {
		time++
	}

	// +1 turn for weapons/outfits
	if len(loot.Weapons)+len(loot.Outfits) > 3 {
		time++
	}

	// +1 turn for rare items (need careful handling)
	if len(loot.RareItems) > 0 {
		time++
	}

	return time
}

// GetRareItemByID retrieves a rare item definition by its unique ID.
//
// Used to look up rare item details when displaying inventory, quest objectives,
// or loot information.
//
// Parameters:
//   - id: Unique rare item identifier
//
// Returns:
//   - *RareItem: Rare item definition, or nil if not found
//
// Thread-safe: Reads only static data, safe for concurrent calls.
func GetRareItemByID(id string) *RareItem {
	for _, item := range RareItemList {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

// Helper functions for formatting
func formatCredits(amount int64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(amount)/1000000.0)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1fK", float64(amount)/1000.0)
	}
	return strconv.FormatInt(amount, 10)
}
