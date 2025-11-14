// File: internal/loadouts/manager.go
// Project: Terminal Velocity
// Description: Ship loadout sharing and comparison system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package loadouts

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Loadouts")

// Manager handles loadout sharing and comparison
type Manager struct {
	mu         sync.RWMutex
	loadoutRepo *database.LoadoutRepository

	// Cache of recently accessed loadouts
	cache map[uuid.UUID]*models.SharedLoadout
}

// NewManager creates a new loadout manager
func NewManager(loadoutRepo *database.LoadoutRepository) *Manager {
	return &Manager{
		loadoutRepo: loadoutRepo,
		cache:       make(map[uuid.UUID]*models.SharedLoadout),
	}
}

// Start initializes the loadout manager
func (m *Manager) Start() error {
	log.Info("Loadout manager started")
	return nil
}

// Stop cleans up the loadout manager
func (m *Manager) Stop() {
	log.Info("Loadout manager stopped")
}

// ShareLoadout shares a player's current ship loadout
func (m *Manager) ShareLoadout(ctx context.Context, playerID uuid.UUID, ship *models.Ship, name, description string, isPublic bool) (*models.SharedLoadout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get ship type
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return nil, fmt.Errorf("invalid ship type: %s", ship.TypeID)
	}

	// Create shared loadout
	loadout := &models.SharedLoadout{
		ID:          uuid.New(),
		PlayerID:    playerID,
		ShipTypeID:  ship.TypeID,
		Name:        name,
		Description: description,
		Weapons:     ship.Weapons,
		Outfits:     ship.Outfits,
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Views:       0,
		Favorites:   0,
	}

	// Calculate stats
	loadout.Stats = m.calculateLoadoutStats(shipType, ship.Weapons, ship.Outfits)

	if err := m.loadoutRepo.CreateLoadout(ctx, loadout); err != nil {
		log.Error("Failed to create shared loadout: %v", err)
		return nil, err
	}

	// Add to cache
	m.cache[loadout.ID] = loadout

	log.Info("Loadout shared by player %s: %s", playerID, name)
	return loadout, nil
}

// GetLoadout retrieves a shared loadout by ID
func (m *Manager) GetLoadout(ctx context.Context, loadoutID uuid.UUID, viewerID uuid.UUID) (*models.SharedLoadout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check cache first
	if cached, ok := m.cache[loadoutID]; ok {
		// Increment view count
		m.loadoutRepo.IncrementViews(ctx, loadoutID)
		return cached, nil
	}

	// Load from database
	loadout, err := m.loadoutRepo.GetLoadout(ctx, loadoutID)
	if err != nil {
		return nil, err
	}

	// Check access permissions
	if !loadout.IsPublic && loadout.PlayerID != viewerID {
		return nil, fmt.Errorf("access denied: loadout is private")
	}

	// Increment view count
	m.loadoutRepo.IncrementViews(ctx, loadoutID)
	loadout.Views++

	// Add to cache
	m.cache[loadoutID] = loadout

	return loadout, nil
}

// GetPlayerLoadouts retrieves all loadouts created by a player
func (m *Manager) GetPlayerLoadouts(ctx context.Context, playerID uuid.UUID) ([]*models.SharedLoadout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadouts, err := m.loadoutRepo.GetPlayerLoadouts(ctx, playerID)
	if err != nil {
		log.Error("Failed to get player loadouts: %v", err)
		return nil, err
	}

	return loadouts, nil
}

// GetPublicLoadouts retrieves public loadouts, optionally filtered by ship type
func (m *Manager) GetPublicLoadouts(ctx context.Context, shipTypeID string, limit, offset int) ([]*models.SharedLoadout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadouts, err := m.loadoutRepo.GetPublicLoadouts(ctx, shipTypeID, limit, offset)
	if err != nil {
		log.Error("Failed to get public loadouts: %v", err)
		return nil, err
	}

	return loadouts, nil
}

// GetPopularLoadouts retrieves most viewed/favorited loadouts
func (m *Manager) GetPopularLoadouts(ctx context.Context, limit int) ([]*models.SharedLoadout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadouts, err := m.loadoutRepo.GetPopularLoadouts(ctx, limit)
	if err != nil {
		log.Error("Failed to get popular loadouts: %v", err)
		return nil, err
	}

	return loadouts, nil
}

// UpdateLoadout updates a shared loadout
func (m *Manager) UpdateLoadout(ctx context.Context, loadoutID, playerID uuid.UUID, name, description string, isPublic bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify ownership
	loadout, err := m.loadoutRepo.GetLoadout(ctx, loadoutID)
	if err != nil {
		return err
	}

	if loadout.PlayerID != playerID {
		return fmt.Errorf("access denied: not the loadout owner")
	}

	// Update fields
	loadout.Name = name
	loadout.Description = description
	loadout.IsPublic = isPublic
	loadout.UpdatedAt = time.Now()

	if err := m.loadoutRepo.UpdateLoadout(ctx, loadout); err != nil {
		log.Error("Failed to update loadout: %v", err)
		return err
	}

	// Update cache
	m.cache[loadoutID] = loadout

	log.Info("Loadout updated: %s", loadoutID)
	return nil
}

// DeleteLoadout deletes a shared loadout
func (m *Manager) DeleteLoadout(ctx context.Context, loadoutID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify ownership
	loadout, err := m.loadoutRepo.GetLoadout(ctx, loadoutID)
	if err != nil {
		return err
	}

	if loadout.PlayerID != playerID {
		return fmt.Errorf("access denied: not the loadout owner")
	}

	if err := m.loadoutRepo.DeleteLoadout(ctx, loadoutID); err != nil {
		log.Error("Failed to delete loadout: %v", err)
		return err
	}

	// Remove from cache
	delete(m.cache, loadoutID)

	log.Info("Loadout deleted: %s", loadoutID)
	return nil
}

// FavoriteLoadout adds a loadout to player's favorites
func (m *Manager) FavoriteLoadout(ctx context.Context, loadoutID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.loadoutRepo.AddFavorite(ctx, loadoutID, playerID); err != nil {
		log.Error("Failed to favorite loadout: %v", err)
		return err
	}

	// Increment favorite count
	if loadout, ok := m.cache[loadoutID]; ok {
		loadout.Favorites++
	}

	log.Debug("Player %s favorited loadout %s", playerID, loadoutID)
	return nil
}

// UnfavoriteLoadout removes a loadout from player's favorites
func (m *Manager) UnfavoriteLoadout(ctx context.Context, loadoutID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.loadoutRepo.RemoveFavorite(ctx, loadoutID, playerID); err != nil {
		log.Error("Failed to unfavorite loadout: %v", err)
		return err
	}

	// Decrement favorite count
	if loadout, ok := m.cache[loadoutID]; ok && loadout.Favorites > 0 {
		loadout.Favorites--
	}

	log.Debug("Player %s unfavorited loadout %s", playerID, loadoutID)
	return nil
}

// GetFavorites retrieves a player's favorited loadouts
func (m *Manager) GetFavorites(ctx context.Context, playerID uuid.UUID) ([]*models.SharedLoadout, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadouts, err := m.loadoutRepo.GetFavorites(ctx, playerID)
	if err != nil {
		log.Error("Failed to get favorites: %v", err)
		return nil, err
	}

	return loadouts, nil
}

// IsFavorited checks if a player has favorited a loadout
func (m *Manager) IsFavorited(ctx context.Context, loadoutID, playerID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.loadoutRepo.IsFavorited(ctx, loadoutID, playerID)
}

// CompareLoadouts compares two loadouts and returns a comparison
func (m *Manager) CompareLoadouts(ctx context.Context, loadoutID1, loadoutID2, viewerID uuid.UUID) (*models.LoadoutComparison, error) {
	loadout1, err := m.GetLoadout(ctx, loadoutID1, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadout 1: %w", err)
	}

	loadout2, err := m.GetLoadout(ctx, loadoutID2, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadout 2: %w", err)
	}

	comparison := &models.LoadoutComparison{
		Loadout1: loadout1,
		Loadout2: loadout2,
		Differences: m.findDifferences(loadout1, loadout2),
	}

	return comparison, nil
}

// ApplyLoadout applies a shared loadout to a player's ship
func (m *Manager) ApplyLoadout(ctx context.Context, loadoutID, playerID uuid.UUID, ship *models.Ship) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	loadout, err := m.GetLoadout(ctx, loadoutID, playerID)
	if err != nil {
		return err
	}

	// Verify ship type matches
	if ship.TypeID != loadout.ShipTypeID {
		return fmt.Errorf("loadout is for %s, but you have a %s", loadout.ShipTypeID, ship.TypeID)
	}

	// Apply loadout
	ship.Weapons = make([]string, len(loadout.Weapons))
	copy(ship.Weapons, loadout.Weapons)

	ship.Outfits = make([]string, len(loadout.Outfits))
	copy(ship.Outfits, loadout.Outfits)

	log.Info("Applied loadout %s to player %s's ship", loadoutID, playerID)
	return nil
}

// calculateLoadoutStats calculates statistics for a loadout
func (m *Manager) calculateLoadoutStats(shipType *models.ShipType, weapons, outfits []string) *models.LoadoutStats {
	stats := &models.LoadoutStats{
		TotalDPS:     0,
		TotalArmor:   0,
		TotalShield:  0,
		TotalSpeed:   0,
		TotalCargo:   0,
		EnergyUsage:  0,
		MassUsage:    0,
	}

	// Base stats from ship type
	stats.TotalArmor = shipType.MaxHull
	stats.TotalShield = shipType.MaxShields
	stats.TotalSpeed = shipType.Speed
	stats.TotalCargo = shipType.CargoSpace

	// Add weapon stats
	for _, weaponID := range weapons {
		if weaponID == "" {
			continue
		}
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			stats.TotalDPS += weapon.Damage
			stats.EnergyUsage += weapon.EnergyCost
		}
	}

	// Add outfit stats
	for _, outfitID := range outfits {
		if outfitID == "" {
			continue
		}
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			// Apply outfit bonuses based on type
			switch outfit.Type {
			case "hull_reinforcement":
				stats.TotalArmor += outfit.HullBonus
			case "shield_booster":
				stats.TotalShield += outfit.ShieldBonus
			case "engine_upgrade":
				stats.TotalSpeed += outfit.SpeedBonus
			case "cargo_pod":
				stats.TotalCargo += outfit.CargoBonus
			}
			// Note: Mass tracking not yet implemented in Outfit model
		}
	}

	return stats
}

// findDifferences finds differences between two loadouts
func (m *Manager) findDifferences(loadout1, loadout2 *models.SharedLoadout) []string {
	var diffs []string

	// Compare stats
	if loadout1.Stats.TotalDPS != loadout2.Stats.TotalDPS {
		diffs = append(diffs, fmt.Sprintf("DPS: %d vs %d", loadout1.Stats.TotalDPS, loadout2.Stats.TotalDPS))
	}
	if loadout1.Stats.TotalArmor != loadout2.Stats.TotalArmor {
		diffs = append(diffs, fmt.Sprintf("Armor: %d vs %d", loadout1.Stats.TotalArmor, loadout2.Stats.TotalArmor))
	}
	if loadout1.Stats.TotalShield != loadout2.Stats.TotalShield {
		diffs = append(diffs, fmt.Sprintf("Shield: %d vs %d", loadout1.Stats.TotalShield, loadout2.Stats.TotalShield))
	}
	if loadout1.Stats.TotalSpeed != loadout2.Stats.TotalSpeed {
		diffs = append(diffs, fmt.Sprintf("Speed: %d vs %d", loadout1.Stats.TotalSpeed, loadout2.Stats.TotalSpeed))
	}
	if loadout1.Stats.TotalCargo != loadout2.Stats.TotalCargo {
		diffs = append(diffs, fmt.Sprintf("Cargo: %d vs %d", loadout1.Stats.TotalCargo, loadout2.Stats.TotalCargo))
	}

	// Compare weapons
	weaponDiffs := m.compareArrays("Weapon", loadout1.Weapons, loadout2.Weapons)
	diffs = append(diffs, weaponDiffs...)

	// Compare outfits
	outfitDiffs := m.compareArrays("Outfit", loadout1.Outfits, loadout2.Outfits)
	diffs = append(diffs, outfitDiffs...)

	return diffs
}

// compareArrays compares two string arrays and returns differences
func (m *Manager) compareArrays(prefix string, arr1, arr2 []string) []string {
	var diffs []string

	// Create maps for comparison
	map1 := make(map[string]int)
	map2 := make(map[string]int)

	for _, item := range arr1 {
		if item != "" {
			map1[item]++
		}
	}
	for _, item := range arr2 {
		if item != "" {
			map2[item]++
		}
	}

	// Find items in arr1 but not arr2
	for item, count1 := range map1 {
		count2 := map2[item]
		if count1 > count2 {
			diffs = append(diffs, fmt.Sprintf("%s: Loadout1 has %d more %s", prefix, count1-count2, item))
		}
	}

	// Find items in arr2 but not arr1
	for item, count2 := range map2 {
		count1 := map1[item]
		if count2 > count1 {
			diffs = append(diffs, fmt.Sprintf("%s: Loadout2 has %d more %s", prefix, count2-count1, item))
		}
	}

	return diffs
}

// ExportLoadout exports a loadout as JSON
func (m *Manager) ExportLoadout(ctx context.Context, loadoutID, playerID uuid.UUID) (string, error) {
	loadout, err := m.GetLoadout(ctx, loadoutID, playerID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(loadout, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal loadout: %w", err)
	}

	return string(data), nil
}

// ImportLoadout imports a loadout from JSON
func (m *Manager) ImportLoadout(ctx context.Context, jsonData string, playerID uuid.UUID) (*models.SharedLoadout, error) {
	var loadout models.SharedLoadout
	if err := json.Unmarshal([]byte(jsonData), &loadout); err != nil {
		return nil, fmt.Errorf("failed to unmarshal loadout: %w", err)
	}

	// Reset ID and ownership
	loadout.ID = uuid.New()
	loadout.PlayerID = playerID
	loadout.CreatedAt = time.Now()
	loadout.UpdatedAt = time.Now()
	loadout.Views = 0
	loadout.Favorites = 0

	if err := m.loadoutRepo.CreateLoadout(ctx, &loadout); err != nil {
		return nil, err
	}

	log.Info("Imported loadout for player %s: %s", playerID, loadout.Name)
	return &loadout, nil
}
