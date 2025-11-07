package combat

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// LootDrop represents items dropped from a destroyed ship
type LootDrop struct {
	Credits    int64
	Cargo      []models.CargoItem // cargo items
	Outfits    []string           // outfit IDs
	Weapons    []string           // weapon IDs
	RareItems  []RareItem
	Message    string
	TotalValue int64
}

// RareItem represents special loot with unique properties
type RareItem struct {
	ID          string
	Name        string
	Description string
	Rarity      string // "uncommon", "rare", "epic", "legendary"
	Value       int64
	Type        string // "artifact", "component", "data", "contraband"
}

// SalvageResult represents the outcome of salvaging a destroyed ship
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

// GenerateLoot creates a loot drop from a destroyed ship
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

// CalculateCargoSpaceRequired calculates how much cargo space is needed for loot
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

// CanCarryLoot checks if player has enough cargo space
func CanCarryLoot(playerShip *models.Ship, playerShipType *models.ShipType, loot *LootDrop) bool {
	currentCargo := playerShip.GetCargoUsed()
	lootSpace := CalculateCargoSpaceRequired(loot)
	availableSpace := playerShipType.CargoSpace - currentCargo

	return lootSpace <= availableSpace
}

// ApplyLoot adds loot to player's ship and credits
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

// SalvageSpecificItem attempts to salvage a specific item from wreckage
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

// CalculateSalvageTime returns the time (in turns) required to salvage
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

// GetRareItemByID retrieves a rare item by its ID
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
