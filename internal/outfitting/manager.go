// File: internal/outfitting/manager.go
// Project: Terminal Velocity
// Description: Ship outfitting and equipment management
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package outfitting

import (
	"errors"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles ship outfitting and equipment
type Manager struct {
	mu        sync.RWMutex
	equipment map[string]*models.Equipment // Equipment catalog
	loadouts  map[uuid.UUID]*models.ShipLoadout // Player loadouts
	inventory map[uuid.UUID]map[string]int // Player equipment inventory (playerID -> equipmentID -> quantity)
}

// NewManager creates a new outfitting manager
func NewManager() *Manager {
	m := &Manager{
		equipment: make(map[string]*models.Equipment),
		loadouts:  make(map[uuid.UUID]*models.ShipLoadout),
		inventory: make(map[uuid.UUID]map[string]int),
	}

	// Load equipment catalog
	m.loadEquipmentCatalog()

	return m
}

// loadEquipmentCatalog populates the equipment catalog
func (m *Manager) loadEquipmentCatalog() {
	// Weapons
	m.addEquipment(createLaserCannon())
	m.addEquipment(createPlasmaTurret())
	m.addEquipment(createRailgun())
	m.addEquipment(createMissileLauncher())

	// Shields
	m.addEquipment(createBasicShield())
	m.addEquipment(createAdvancedShield())
	m.addEquipment(createMilitaryShield())

	// Engines
	m.addEquipment(createBasicEngine())
	m.addEquipment(createAfterburnerEngine())
	m.addEquipment(createMilitaryEngine())

	// Reactors
	m.addEquipment(createBasicReactor())
	m.addEquipment(createFusionReactor())
	m.addEquipment(createAntimatterReactor())

	// Utilities
	m.addEquipment(createCargoPod())
	m.addEquipment(createFuelTank())
	m.addEquipment(createScanner())
	m.addEquipment(createRepairDrone())
}

// addEquipment adds equipment to the catalog
func (m *Manager) addEquipment(equipment *models.Equipment) {
	m.equipment[equipment.ID] = equipment
}

// GetEquipment retrieves equipment by ID
func (m *Manager) GetEquipment(equipmentID string) (*models.Equipment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	equipment, exists := m.equipment[equipmentID]
	if !exists {
		return nil, errors.New("equipment not found")
	}

	return equipment, nil
}

// GetEquipmentByCategory returns all equipment in a category
func (m *Manager) GetEquipmentByCategory(category models.EquipmentCategory) []*models.Equipment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*models.Equipment
	for _, eq := range m.equipment {
		if eq.Category == category {
			result = append(result, eq)
		}
	}

	return result
}

// GetEquipmentBySlotType returns all equipment for a slot type
func (m *Manager) GetEquipmentBySlotType(slotType models.EquipmentSlotType) []*models.Equipment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*models.Equipment
	for _, eq := range m.equipment {
		if eq.SlotType == slotType {
			result = append(result, eq)
		}
	}

	return result
}

// GetAllEquipment returns all available equipment
func (m *Manager) GetAllEquipment() []*models.Equipment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*models.Equipment
	for _, eq := range m.equipment {
		result = append(result, eq)
	}

	return result
}

// PurchaseEquipment adds equipment to player's inventory
func (m *Manager) PurchaseEquipment(playerID uuid.UUID, equipmentID string, quantity int, playerCredits int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	equipment, exists := m.equipment[equipmentID]
	if !exists {
		return errors.New("equipment not found")
	}

	totalCost := equipment.Price * int64(quantity)
	if playerCredits < totalCost {
		return errors.New("insufficient credits")
	}

	// Add to inventory
	if _, exists := m.inventory[playerID]; !exists {
		m.inventory[playerID] = make(map[string]int)
	}

	m.inventory[playerID][equipmentID] += quantity

	return nil
}

// SellEquipment removes equipment from player's inventory
func (m *Manager) SellEquipment(playerID uuid.UUID, equipmentID string, quantity int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	equipment, exists := m.equipment[equipmentID]
	if !exists {
		return 0, errors.New("equipment not found")
	}

	playerInv, exists := m.inventory[playerID]
	if !exists {
		return 0, errors.New("no inventory for player")
	}

	if playerInv[equipmentID] < quantity {
		return 0, errors.New("insufficient quantity")
	}

	// Remove from inventory
	playerInv[equipmentID] -= quantity
	if playerInv[equipmentID] == 0 {
		delete(playerInv, equipmentID)
	}

	// Sell at 70% of purchase price
	sellPrice := int64(float64(equipment.Price) * 0.7 * float64(quantity))

	return sellPrice, nil
}

// GetPlayerInventory returns player's equipment inventory
func (m *Manager) GetPlayerInventory(playerID uuid.UUID) map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if inv, exists := m.inventory[playerID]; exists {
		// Return a copy
		result := make(map[string]int)
		for k, v := range inv {
			result[k] = v
		}
		return result
	}

	return make(map[string]int)
}

// CreateLoadout creates a new ship loadout
func (m *Manager) CreateLoadout(playerID uuid.UUID, shipType *models.ShipType, name string) (*models.ShipLoadout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	loadout := models.NewShipLoadout(playerID, shipType.ID, name)

	// Create default slots based on ship type
	// Weapon slots
	for i := 0; i < shipType.WeaponSlots; i++ {
		slot := models.NewEquipmentSlot(models.SlotWeapon, 2) // Medium slots
		loadout.AddSlot(slot)
	}

	// Required slots
	loadout.AddSlot(models.NewEquipmentSlot(models.SlotReactor, 2))
	loadout.AddSlot(models.NewEquipmentSlot(models.SlotEngine, 2))
	loadout.AddSlot(models.NewEquipmentSlot(models.SlotShield, 2))

	// Utility slots (2-4 depending on ship class)
	utilitySlots := 2
	switch shipType.Class {
	case "freighter":
		utilitySlots = 4
	case "corvette", "destroyer":
		utilitySlots = 3
	}

	for i := 0; i < utilitySlots; i++ {
		slot := models.NewEquipmentSlot(models.SlotUtility, 1)
		loadout.AddSlot(slot)
	}

	m.loadouts[loadout.ID] = loadout

	return loadout, nil
}

// GetLoadout retrieves a loadout by ID
func (m *Manager) GetLoadout(loadoutID uuid.UUID) (*models.ShipLoadout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadout, exists := m.loadouts[loadoutID]
	if !exists {
		return nil, errors.New("loadout not found")
	}

	return loadout, nil
}

// GetPlayerLoadouts returns all loadouts for a player
func (m *Manager) GetPlayerLoadouts(playerID uuid.UUID) []*models.ShipLoadout {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*models.ShipLoadout
	for _, loadout := range m.loadouts {
		if loadout.PlayerID == playerID {
			result = append(result, loadout)
		}
	}

	return result
}

// SaveLoadout updates an existing loadout
func (m *Manager) SaveLoadout(loadout *models.ShipLoadout) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.loadouts[loadout.ID]; !exists {
		return errors.New("loadout not found")
	}

	m.loadouts[loadout.ID] = loadout

	return nil
}

// DeleteLoadout removes a loadout
func (m *Manager) DeleteLoadout(loadoutID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.loadouts[loadoutID]; !exists {
		return errors.New("loadout not found")
	}

	delete(m.loadouts, loadoutID)

	return nil
}

// InstallEquipment installs equipment in a loadout slot
func (m *Manager) InstallEquipment(
	playerID uuid.UUID,
	loadoutID uuid.UUID,
	slotID uuid.UUID,
	equipmentID string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get loadout
	loadout, exists := m.loadouts[loadoutID]
	if !exists {
		return errors.New("loadout not found")
	}

	if loadout.PlayerID != playerID {
		return errors.New("not your loadout")
	}

	// Get equipment
	equipment, exists := m.equipment[equipmentID]
	if !exists {
		return errors.New("equipment not found")
	}

	// Check player inventory
	playerInv, exists := m.inventory[playerID]
	if !exists || playerInv[equipmentID] < 1 {
		return errors.New("equipment not in inventory")
	}

	// Get slot
	slot := loadout.GetSlotByID(slotID)
	if slot == nil {
		return errors.New("slot not found")
	}

	// Check if equipment can be installed
	if !slot.CanInstall(equipment) {
		return errors.New("equipment cannot be installed in this slot")
	}

	// If slot already has equipment, return it to inventory
	if !slot.IsEmpty() {
		oldEquipment := slot.Uninstall()
		playerInv[oldEquipment.ID]++
	}

	// Install new equipment
	slot.Install(equipment)

	// Remove from inventory
	playerInv[equipmentID]--
	if playerInv[equipmentID] == 0 {
		delete(playerInv, equipmentID)
	}

	// Update loadout stats
	loadout.CalculateTotalCost()
	loadout.CalculateOutfitSpace()

	return nil
}

// UninstallEquipment removes equipment from a loadout slot
func (m *Manager) UninstallEquipment(
	playerID uuid.UUID,
	loadoutID uuid.UUID,
	slotID uuid.UUID,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get loadout
	loadout, exists := m.loadouts[loadoutID]
	if !exists {
		return errors.New("loadout not found")
	}

	if loadout.PlayerID != playerID {
		return errors.New("not your loadout")
	}

	// Get slot
	slot := loadout.GetSlotByID(slotID)
	if slot == nil {
		return errors.New("slot not found")
	}

	if slot.IsEmpty() {
		return errors.New("slot is already empty")
	}

	// Uninstall equipment
	equipment := slot.Uninstall()
	if equipment == nil {
		return errors.New("cannot uninstall from locked slot")
	}

	// Add back to inventory
	if _, exists := m.inventory[playerID]; !exists {
		m.inventory[playerID] = make(map[string]int)
	}
	m.inventory[playerID][equipment.ID]++

	// Update loadout stats
	loadout.CalculateTotalCost()
	loadout.CalculateOutfitSpace()

	return nil
}

// GetStats returns equipment manager statistics
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"equipment_types": len(m.equipment),
		"total_loadouts":  len(m.loadouts),
		"players_with_inventory": len(m.inventory),
	}

	// Count by category
	categoryCount := make(map[models.EquipmentCategory]int)
	for _, eq := range m.equipment {
		categoryCount[eq.Category]++
	}

	stats["weapons"] = categoryCount[models.CategoryWeapon]
	stats["defense"] = categoryCount[models.CategoryDefense]
	stats["power"] = categoryCount[models.CategoryPower]
	stats["propulsion"] = categoryCount[models.CategoryPropulsion]
	stats["utility"] = categoryCount[models.CategoryUtility]
	stats["special"] = categoryCount[models.CategorySpecial]

	return stats
}
