// File: internal/mining/manager.go
// Project: Terminal Velocity
// Description: Mining and salvage operations manager
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15

package mining

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Mining")

// Manager handles mining and salvage operations
type Manager struct {
	mu sync.RWMutex

	// Active operations
	activeOperations map[uuid.UUID]*MiningOperation

	// Configuration
	config MiningConfig

	// Repositories
	systemRepo *database.SystemRepository
	shipRepo   *database.ShipRepository
	playerRepo *database.PlayerRepository
}

// MiningConfig defines mining system parameters
type MiningConfig struct {
	// Mining parameters
	BaseMiningYield        float64       // Base yield per mining cycle
	MiningCycleDuration    time.Duration // How long each mining cycle takes
	MiningLaserBonus       float64       // Bonus per mining laser level
	CargoScannerBonus      float64       // Bonus from cargo scanner
	AsteroidDepletionRate  float64       // How quickly asteroids deplete
	MaxConcurrentOperations int           // Max mining ops per player

	// Salvage parameters
	BaseSalvageYield      float64       // Base salvage yield
	SalvageCycleDuration  time.Duration // Salvage cycle time
	SalvageScannerBonus   float64       // Bonus from salvage scanner
	DerelictRareItemChance float64       // Chance of rare items
	DebrisFieldDensity    float64       // Average debris field density

	// Resource spawn
	AsteroidSpawnChance  float64 // Chance per system visit
	DerelictSpawnChance  float64 // Chance per system visit
	DebrisFieldChance    float64 // Chance per system visit
	RareResourceModifier float64 // Multiplier for rare resources
}

// DefaultMiningConfig returns sensible defaults
func DefaultMiningConfig() MiningConfig {
	return MiningConfig{
		BaseMiningYield:         10.0,
		MiningCycleDuration:     15 * time.Second,
		MiningLaserBonus:        0.25,  // +25% per level
		CargoScannerBonus:       0.15,  // +15% from scanner
		AsteroidDepletionRate:   0.10,  // -10% per cycle
		MaxConcurrentOperations: 1,
		BaseSalvageYield:        8.0,
		SalvageCycleDuration:    20 * time.Second,
		SalvageScannerBonus:     0.20,  // +20% from scanner
		DerelictRareItemChance:  0.05,  // 5% chance
		DebrisFieldDensity:      0.50,  // 50% average
		AsteroidSpawnChance:     0.30,  // 30% chance
		DerelictSpawnChance:     0.10,  // 10% chance
		DebrisFieldChance:       0.15,  // 15% chance
		RareResourceModifier:    2.5,   // 2.5x value
	}
}

// MiningOperation represents an active mining or salvage operation
type MiningOperation struct {
	ID           uuid.UUID
	PlayerID     uuid.UUID
	ShipID       uuid.UUID
	Type         string // "mining", "salvage"
	Target       *MiningTarget
	StartTime    time.Time
	CyclesLeft   int
	CurrentYield float64
	Status       string // "active", "completed", "cancelled"
	Resources    map[string]int // Resource type -> quantity
}

// MiningTarget represents what is being mined or salvaged
type MiningTarget struct {
	Type        string // "asteroid", "derelict", "debris_field"
	ID          uuid.UUID
	Name        string
	Resources   map[string]float64 // Resource type -> remaining amount
	Rarity      string             // "common", "uncommon", "rare", "legendary"
	Coordinates string
	SystemID    uuid.UUID
}

// ResourceType represents a mineable or salvageable resource
type ResourceType string

const (
	ResourceIron       ResourceType = "iron"
	ResourceCopper     ResourceType = "copper"
	ResourceTitanium   ResourceType = "titanium"
	ResourcePlatinum   ResourceType = "platinum"
	ResourceGold       ResourceType = "gold"
	ResourceCrystals   ResourceType = "crystals"
	ResourceRareEarth  ResourceType = "rare_earth"
	ResourceDeuterium  ResourceType = "deuterium"
	ResourceScrap      ResourceType = "scrap_metal"
	ResourceComponents ResourceType = "components"
	ResourceWeapons    ResourceType = "salvaged_weapons"
	ResourceOutfits    ResourceType = "salvaged_outfits"
)

// NewManager creates a new mining and salvage manager
func NewManager(systemRepo *database.SystemRepository, shipRepo *database.ShipRepository, playerRepo *database.PlayerRepository) *Manager {
	return &Manager{
		activeOperations: make(map[uuid.UUID]*MiningOperation),
		config:           DefaultMiningConfig(),
		systemRepo:       systemRepo,
		shipRepo:         shipRepo,
		playerRepo:       playerRepo,
	}
}

// ScanForResources scans a system for mining and salvage opportunities
func (m *Manager) ScanForResources(ctx context.Context, systemID uuid.UUID, hasScannerBonus bool) ([]MiningTarget, error) {
	targets := []MiningTarget{}

	// Generate asteroids
	if rand.Float64() < m.config.AsteroidSpawnChance {
		numAsteroids := rand.Intn(3) + 1 // 1-3 asteroids
		for i := 0; i < numAsteroids; i++ {
			asteroid := m.generateAsteroid(systemID)
			targets = append(targets, asteroid)
		}
	}

	// Generate derelicts
	if rand.Float64() < m.config.DerelictSpawnChance {
		derelict := m.generateDerelict(systemID)
		targets = append(targets, derelict)
	}

	// Generate debris fields
	if rand.Float64() < m.config.DebrisFieldChance {
		debris := m.generateDebrisField(systemID)
		targets = append(targets, debris)
	}

	// Scanner bonus reveals more targets
	if hasScannerBonus && len(targets) > 0 {
		// 50% chance to reveal one more target
		if rand.Float64() < 0.5 {
			extra := m.generateAsteroid(systemID)
			targets = append(targets, extra)
		}
	}

	log.Debug("Scanned system %s: found %d mining targets", systemID, len(targets))
	return targets, nil
}

// generateAsteroid creates a random asteroid target
func (m *Manager) generateAsteroid(systemID uuid.UUID) MiningTarget {
	rarities := []string{"common", "common", "common", "uncommon", "uncommon", "rare"}
	rarity := rarities[rand.Intn(len(rarities))]

	resources := make(map[string]float64)

	// Common asteroids have basic resources
	if rarity == "common" {
		resources[string(ResourceIron)] = float64(rand.Intn(50) + 50)       // 50-100
		resources[string(ResourceCopper)] = float64(rand.Intn(30) + 20)     // 20-50
	} else if rarity == "uncommon" {
		resources[string(ResourceTitanium)] = float64(rand.Intn(40) + 30)   // 30-70
		resources[string(ResourceGold)] = float64(rand.Intn(20) + 10)       // 10-30
	} else {
		// Rare asteroids
		resources[string(ResourcePlatinum)] = float64(rand.Intn(30) + 20)   // 20-50
		resources[string(ResourceCrystals)] = float64(rand.Intn(25) + 15)   // 15-40
		resources[string(ResourceRareEarth)] = float64(rand.Intn(15) + 10)  // 10-25
	}

	return MiningTarget{
		Type:        "asteroid",
		ID:          uuid.New(),
		Name:        fmt.Sprintf("%s Asteroid %d", capitalize(rarity), rand.Intn(9999)),
		Resources:   resources,
		Rarity:      rarity,
		Coordinates: fmt.Sprintf("%d,%d,%d", rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)),
		SystemID:    systemID,
	}
}

// generateDerelict creates a random derelict ship target
func (m *Manager) generateDerelict(systemID uuid.UUID) MiningTarget {
	rarities := []string{"common", "common", "uncommon", "rare"}
	rarity := rarities[rand.Intn(len(rarities))]

	resources := make(map[string]float64)
	resources[string(ResourceScrap)] = float64(rand.Intn(100) + 50) // 50-150
	resources[string(ResourceComponents)] = float64(rand.Intn(40) + 20) // 20-60

	// Rare derelicts may have weapons/outfits
	if rarity == "uncommon" {
		resources[string(ResourceWeapons)] = float64(rand.Intn(3) + 1) // 1-3
	} else if rarity == "rare" {
		resources[string(ResourceWeapons)] = float64(rand.Intn(5) + 2) // 2-6
		resources[string(ResourceOutfits)] = float64(rand.Intn(4) + 1) // 1-4
	}

	shipTypes := []string{"Shuttle", "Fighter", "Freighter", "Corvette", "Destroyer"}
	shipType := shipTypes[rand.Intn(len(shipTypes))]

	return MiningTarget{
		Type:        "derelict",
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Derelict %s", shipType),
		Resources:   resources,
		Rarity:      rarity,
		Coordinates: fmt.Sprintf("%d,%d,%d", rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)),
		SystemID:    systemID,
	}
}

// generateDebrisField creates a random debris field
func (m *Manager) generateDebrisField(systemID uuid.UUID) MiningTarget {
	resources := make(map[string]float64)
	resources[string(ResourceScrap)] = float64(rand.Intn(200) + 100) // 100-300
	resources[string(ResourceIron)] = float64(rand.Intn(50) + 25)     // 25-75
	resources[string(ResourceComponents)] = float64(rand.Intn(30) + 10) // 10-40

	return MiningTarget{
		Type:        "debris_field",
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Debris Field %d", rand.Intn(9999)),
		Resources:   resources,
		Rarity:      "common",
		Coordinates: fmt.Sprintf("%d,%d,%d", rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)),
		SystemID:    systemID,
	}
}

// StartMining initiates a mining operation
func (m *Manager) StartMining(ctx context.Context, playerID, shipID uuid.UUID, target *MiningTarget, miningLaserLevel int, hasCargoScanner bool) (*MiningOperation, error) {
	// Check for existing operation
	m.mu.RLock()
	if existing, exists := m.activeOperations[shipID]; exists {
		m.mu.RUnlock()
		return nil, fmt.Errorf("already mining (started %v ago)", time.Since(existing.StartTime))
	}
	m.mu.RUnlock()

	// Check if target has resources left
	totalResources := 0.0
	for _, amount := range target.Resources {
		totalResources += amount
	}
	if totalResources <= 0 {
		return nil, fmt.Errorf("target is depleted")
	}

	// Calculate number of cycles based on resources
	cycles := int(totalResources / m.config.BaseMiningYield)
	if cycles < 1 {
		cycles = 1
	}
	if cycles > 10 {
		cycles = 10 // Cap at 10 cycles
	}

	// Create mining operation
	operation := &MiningOperation{
		ID:           uuid.New(),
		PlayerID:     playerID,
		ShipID:       shipID,
		Type:         "mining",
		Target:       target,
		StartTime:    time.Now(),
		CyclesLeft:   cycles,
		CurrentYield: 0,
		Status:       "active",
		Resources:    make(map[string]int),
	}

	m.mu.Lock()
	m.activeOperations[shipID] = operation
	m.mu.Unlock()

	log.Info("Mining started: player=%s, target=%s, cycles=%d",
		playerID, target.Name, cycles)

	// Start mining cycles
	go m.runMiningCycles(ctx, operation, miningLaserLevel, hasCargoScanner)

	return operation, nil
}

// runMiningCycles executes mining cycles
func (m *Manager) runMiningCycles(ctx context.Context, operation *MiningOperation, miningLaserLevel int, hasCargoScanner bool) {
	for operation.CyclesLeft > 0 && operation.Status == "active" {
		// Wait for cycle duration
		time.Sleep(m.config.MiningCycleDuration)

		// Calculate yield for this cycle
		baseYield := m.config.BaseMiningYield
		bonusYield := baseYield * float64(miningLaserLevel) * m.config.MiningLaserBonus
		if hasCargoScanner {
			bonusYield += baseYield * m.config.CargoScannerBonus
		}
		totalYield := baseYield + bonusYield

		// Extract resources from target
		m.extractResources(operation, totalYield)

		operation.CyclesLeft--
		log.Debug("Mining cycle completed: cycles_left=%d, yield=%.1f",
			operation.CyclesLeft, totalYield)
	}

	// Mark as completed
	operation.Status = "completed"
	log.Info("Mining completed: player=%s, total_yield=%.1f",
		operation.PlayerID, operation.CurrentYield)

	// Update player stats for completed mining operation
	if m.playerRepo != nil {
		if player, err := m.playerRepo.GetByID(ctx, operation.PlayerID); err == nil {
			player.RecordMiningOperation(int64(operation.CurrentYield))
			if err := m.playerRepo.Update(ctx, player); err != nil {
				log.Error("Failed to update player stats for mining operation: %v", err)
			}
		}
	}

	// Clean up after 60 seconds
	time.Sleep(60 * time.Second)
	m.mu.Lock()
	delete(m.activeOperations, operation.ShipID)
	m.mu.Unlock()
}

// extractResources extracts resources from the target
func (m *Manager) extractResources(operation *MiningOperation, yieldAmount float64) {
	target := operation.Target

	// Calculate what percentage of total resources this yield represents
	totalRemaining := 0.0
	for _, amount := range target.Resources {
		totalRemaining += amount
	}

	if totalRemaining <= 0 {
		return
	}

	yieldPercentage := yieldAmount / totalRemaining
	if yieldPercentage > 1.0 {
		yieldPercentage = 1.0
	}

	// Extract proportional amounts from each resource type
	for resourceType, remainingAmount := range target.Resources {
		if remainingAmount > 0 {
			extracted := remainingAmount * yieldPercentage
			operation.Resources[resourceType] += int(extracted)
			target.Resources[resourceType] -= extracted
			operation.CurrentYield += extracted
		}
	}
}

// CancelOperation cancels an active mining operation
func (m *Manager) CancelOperation(shipID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	operation, exists := m.activeOperations[shipID]
	if !exists {
		return fmt.Errorf("no active operation")
	}

	if operation.Status != "active" {
		return fmt.Errorf("operation already completed")
	}

	operation.Status = "cancelled"
	log.Info("Mining operation cancelled: player=%s", operation.PlayerID)

	return nil
}

// GetActiveOperation retrieves an active operation
func (m *Manager) GetActiveOperation(shipID uuid.UUID) (*MiningOperation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	operation, exists := m.activeOperations[shipID]
	return operation, exists
}

// GetActiveOperationCount returns the number of active operations
func (m *Manager) GetActiveOperationCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.activeOperations)
}

// Helper functions

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Convert first character to uppercase using strings package
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// MiningStats contains statistics about mining operations
type MiningStats struct {
	ActiveOperations   int
	TotalOperations    int
	TotalYield         float64
	MostCommonResource string
}

// GetStats returns mining statistics for a specific player
func (m *Manager) GetStats(ctx context.Context, playerID uuid.UUID) MiningStats {
	m.mu.RLock()
	activeOperations := len(m.activeOperations)
	m.mu.RUnlock()

	// Get player from database to retrieve stats
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Error("Failed to get player stats: %v", err)
		return MiningStats{
			ActiveOperations:   activeOperations,
			TotalOperations:    0,
			TotalYield:         0,
			MostCommonResource: string(ResourceIron),
		}
	}

	return MiningStats{
		ActiveOperations:   activeOperations,
		TotalOperations:    player.TotalMiningOps,
		TotalYield:         float64(player.TotalYield),
		MostCommonResource: string(ResourceIron), // TODO: Track most common resource
	}
}
