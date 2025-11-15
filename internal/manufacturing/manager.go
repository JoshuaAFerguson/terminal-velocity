// File: internal/manufacturing/manager.go
// Project: Terminal Velocity
// Description: Manufacturing system with crafting, tech tree, and player stations
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15

package manufacturing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Manufacturing")

// Manager handles crafting, tech research, and player stations
type Manager struct {
	mu sync.RWMutex

	// Manufacturing data
	blueprints    map[uuid.UUID]*Blueprint
	craftingJobs  map[uuid.UUID]*CraftingJob
	stations      map[uuid.UUID]*PlayerStation
	technologies  map[uuid.UUID]*Technology
	playerTech    map[uuid.UUID]map[string]int // playerID -> techID -> level

	// Configuration
	config ManufacturingConfig

	// Repositories
	playerRepo *database.PlayerRepository

	// Callbacks
	onCraftingComplete  func(job *CraftingJob)
	onTechResearched    func(playerID uuid.UUID, tech *Technology)
	onStationBuilt      func(station *PlayerStation)
	onStationUpgraded   func(station *PlayerStation)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// ManufacturingConfig defines manufacturing parameters
type ManufacturingConfig struct {
	// Crafting settings
	CraftingSpeedModifier    float64       // Global crafting speed multiplier
	CraftingCostModifier     float64       // Global cost multiplier
	MaxConcurrentJobs        int           // Max crafting jobs per player
	CraftingSkillBonusRate   float64       // Skill bonus per level

	// Tech research settings
	ResearchPointsPerDay     int           // Daily research points
	MaxResearchQueue         int           // Max queued research
	TechCostScaling          float64       // Cost increase per level
	TechPrerequisiteStrict   bool          // Require all prerequisites

	// Station settings
	StationBuildCost         int64         // Base cost to build station
	StationUpgradeCost       int64         // Base upgrade cost
	MaxStationsPerPlayer     int           // Max stations per player
	StationProductionBonus   float64       // Production bonus from station
	StationStorageCapacity   int           // Base storage capacity
}

// DefaultManufacturingConfig returns sensible defaults
func DefaultManufacturingConfig() ManufacturingConfig {
	return ManufacturingConfig{
		CraftingSpeedModifier:    1.0,
		CraftingCostModifier:     1.0,
		MaxConcurrentJobs:        3,
		CraftingSkillBonusRate:   0.05,  // 5% per level
		ResearchPointsPerDay:     100,
		MaxResearchQueue:         5,
		TechCostScaling:          1.5,   // 50% increase per level
		TechPrerequisiteStrict:   true,
		StationBuildCost:         1000000,
		StationUpgradeCost:       500000,
		MaxStationsPerPlayer:     3,
		StationProductionBonus:   0.25,  // 25% bonus
		StationStorageCapacity:   10000,
	}
}

// NewManager creates a new manufacturing manager
func NewManager(playerRepo *database.PlayerRepository) *Manager {
	m := &Manager{
		blueprints:   make(map[uuid.UUID]*Blueprint),
		craftingJobs: make(map[uuid.UUID]*CraftingJob),
		stations:     make(map[uuid.UUID]*PlayerStation),
		technologies: make(map[uuid.UUID]*Technology),
		playerTech:   make(map[uuid.UUID]map[string]int),
		config:       DefaultManufacturingConfig(),
		playerRepo:   playerRepo,
		stopChan:     make(chan struct{}),
	}

	// Initialize default blueprints and technologies
	m.initializeBlueprints()
	m.initializeTechnologies()

	return m
}

// Start begins background workers
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.craftingWorker()
	log.Info("Manufacturing manager started")
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Manufacturing manager stopped")
}

// SetCallbacks sets all manufacturing callbacks
func (m *Manager) SetCallbacks(
	onCraftingComplete func(job *CraftingJob),
	onTechResearched func(playerID uuid.UUID, tech *Technology),
	onStationBuilt func(station *PlayerStation),
	onStationUpgraded func(station *PlayerStation),
) {
	m.onCraftingComplete = onCraftingComplete
	m.onTechResearched = onTechResearched
	m.onStationBuilt = onStationBuilt
	m.onStationUpgraded = onStationUpgraded
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// Blueprint defines how to craft an item
type Blueprint struct {
	ID           uuid.UUID
	Name         string
	Description  string
	ItemType     string // "weapon", "outfit", "ship_component", "consumable"
	Tier         int    // 1-5 rarity tier
	CraftingTime time.Duration
	Requirements map[string]int // resource -> quantity
	Produces     map[string]int // item -> quantity
	SkillLevel   int            // Required crafting skill
	TechRequired string         // Required technology ID (optional)
}

// CraftingJob represents an active crafting operation
type CraftingJob struct {
	ID           uuid.UUID
	PlayerID     uuid.UUID
	BlueprintID  uuid.UUID
	Blueprint    *Blueprint
	StartTime    time.Time
	CompletionTime time.Time
	Status       string    // "in_progress", "complete", "failed"
	Quantity     int
	StationID    *uuid.UUID // Optional: crafting at a station
}

// PlayerStation represents a player-owned manufacturing station
type PlayerStation struct {
	ID            uuid.UUID
	OwnerID       uuid.UUID
	Name          string
	SystemID      uuid.UUID
	SystemName    string
	Level         int       // Station upgrade level 1-10
	BuildTime     time.Time
	Facilities    []StationFacility
	Storage       map[string]int // resource -> quantity
	StorageCapacity int
	ProductionBonus float64  // Bonus to crafting speed
	Status        string    // "active", "upgrading", "damaged"
}

// StationFacility represents a facility within a station
type StationFacility string

const (
	FacilityManufacturing StationFacility = "manufacturing" // Craft items
	FacilityResearch      StationFacility = "research"      // Research tech
	FacilityRefinery      StationFacility = "refinery"      // Process raw materials
	FacilityShipyard      StationFacility = "shipyard"      // Build ships
	FacilityWarehouse     StationFacility = "warehouse"     // Extra storage
	FacilityDefense       StationFacility = "defense"       // Station defenses
)

// Technology represents a researchable technology
type Technology struct {
	ID            string
	Name          string
	Description   string
	Category      TechCategory
	MaxLevel      int
	ResearchCost  int       // Base research points needed
	CreditCost    int64     // Credits required
	Prerequisites []string  // Required tech IDs
	Unlocks       []string  // What this tech unlocks
	Benefits      map[string]float64 // Bonuses provided
}

// TechCategory defines technology types
type TechCategory string

const (
	TechCategoryWeapons    TechCategory = "weapons"
	TechCategoryDefense    TechCategory = "defense"
	TechCategoryEngines    TechCategory = "engines"
	TechCategoryEnergy     TechCategory = "energy"
	TechCategoryManufacturing TechCategory = "manufacturing"
	TechCategoryCloaking   TechCategory = "cloaking"
	TechCategoryJumpDrive  TechCategory = "jump_drive"
)

// ============================================================================
// CRAFTING SYSTEM
// ============================================================================

// StartCrafting initiates a crafting job
func (m *Manager) StartCrafting(ctx context.Context, playerID uuid.UUID, blueprintID uuid.UUID, quantity int, stationID *uuid.UUID) (*CraftingJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get blueprint
	blueprint, exists := m.blueprints[blueprintID]
	if !exists {
		return nil, fmt.Errorf("blueprint not found")
	}

	// Check concurrent job limit
	activeJobs := 0
	for _, job := range m.craftingJobs {
		if job.PlayerID == playerID && job.Status == "in_progress" {
			activeJobs++
		}
	}
	if activeJobs >= m.config.MaxConcurrentJobs {
		return nil, fmt.Errorf("maximum concurrent crafting jobs reached (%d)", m.config.MaxConcurrentJobs)
	}

	// Check tech requirements
	if blueprint.TechRequired != "" {
		playerTechs := m.playerTech[playerID]
		if playerTechs == nil || playerTechs[blueprint.TechRequired] == 0 {
			return nil, fmt.Errorf("required technology: %s", blueprint.TechRequired)
		}
	}

	// Check skill level (TODO: Add actual skill system)
	// For now, assume player has sufficient skill

	// Check resources (TODO: Implement resource checking)
	// For now, assume player has resources

	// Calculate crafting time
	baseTime := blueprint.CraftingTime * time.Duration(quantity)
	productionBonus := 1.0

	// Station bonus
	if stationID != nil {
		station, exists := m.stations[*stationID]
		if exists && station.OwnerID == playerID {
			productionBonus += station.ProductionBonus
		}
	}

	craftingTime := time.Duration(float64(baseTime) / productionBonus)

	// Create crafting job
	job := &CraftingJob{
		ID:             uuid.New(),
		PlayerID:       playerID,
		BlueprintID:    blueprintID,
		Blueprint:      blueprint,
		StartTime:      time.Now(),
		CompletionTime: time.Now().Add(craftingTime),
		Status:         "in_progress",
		Quantity:       quantity,
		StationID:      stationID,
	}

	m.craftingJobs[job.ID] = job

	log.Info("Crafting started: player=%s, item=%s, quantity=%d, time=%v",
		playerID, blueprint.Name, quantity, craftingTime)

	return job, nil
}

// CancelCrafting cancels an in-progress crafting job
func (m *Manager) CancelCrafting(ctx context.Context, jobID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.craftingJobs[jobID]
	if !exists {
		return fmt.Errorf("crafting job not found")
	}

	if job.PlayerID != playerID {
		return fmt.Errorf("not your crafting job")
	}

	if job.Status != "in_progress" {
		return fmt.Errorf("job is not in progress")
	}

	job.Status = "failed"

	// Calculate partial resource refund based on time elapsed
	now := time.Now()
	timeElapsed := now.Sub(job.StartTime)
	totalTime := job.CompletionTime.Sub(job.StartTime)

	if totalTime > 0 {
		progressPercent := float64(timeElapsed) / float64(totalTime)
		if progressPercent > 1.0 {
			progressPercent = 1.0
		}

		// Refund percentage is inverse of progress (if 30% done, refund 70%)
		refundPercent := 1.0 - progressPercent

		// Calculate refunded resources
		if job.Blueprint != nil && len(job.Blueprint.Requirements) > 0 {
			log.Info("Crafting cancelled: job=%s, progress=%.1f%%, refunding %.1f%% of resources",
				jobID, progressPercent*100, refundPercent*100)

			// Log each resource that would be refunded
			// NOTE: When resource system is implemented, actually return these to player
			for resource, required := range job.Blueprint.Requirements {
				totalRequired := required * job.Quantity
				refundAmount := int(float64(totalRequired) * refundPercent)
				if refundAmount > 0 {
					log.Debug("Refund: %s x%d (%.1f%% of %d)",
						resource, refundAmount, refundPercent*100, totalRequired)
					// TODO: When inventory system exists, add resources back:
					// m.addResourceToPlayer(ctx, playerID, resource, refundAmount)
				}
			}
		}
	}

	return nil
}

// GetCraftingJobs retrieves player's crafting jobs
func (m *Manager) GetCraftingJobs(playerID uuid.UUID) []*CraftingJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var jobs []*CraftingJob
	for _, job := range m.craftingJobs {
		if job.PlayerID == playerID {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// GetBlueprints retrieves all available blueprints
func (m *Manager) GetBlueprints() []*Blueprint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blueprints := make([]*Blueprint, 0, len(m.blueprints))
	for _, bp := range m.blueprints {
		blueprints = append(blueprints, bp)
	}
	return blueprints
}

// ============================================================================
// TECHNOLOGY SYSTEM
// ============================================================================

// ResearchTechnology initiates technology research
func (m *Manager) ResearchTechnology(ctx context.Context, playerID uuid.UUID, techID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get technology
	tech, exists := m.technologies[uuid.MustParse(techID)]
	if !exists {
		return fmt.Errorf("technology not found")
	}

	// Check prerequisites
	if m.config.TechPrerequisiteStrict {
		for _, prereqID := range tech.Prerequisites {
			playerTechs := m.playerTech[playerID]
			if playerTechs == nil || playerTechs[prereqID] == 0 {
				return fmt.Errorf("missing prerequisite: %s", prereqID)
			}
		}
	}

	// Get current level
	playerTechs := m.playerTech[playerID]
	if playerTechs == nil {
		playerTechs = make(map[string]int)
		m.playerTech[playerID] = playerTechs
	}

	currentLevel := playerTechs[techID]
	if currentLevel >= tech.MaxLevel {
		return fmt.Errorf("technology already at max level")
	}

	// Calculate cost
	costMultiplier := 1.0
	for i := 0; i < currentLevel; i++ {
		costMultiplier *= m.config.TechCostScaling
	}
	// TODO: Implement research resource system
	// researchCost := int(float64(tech.ResearchCost) * costMultiplier)
	creditCost := int64(float64(tech.CreditCost) * costMultiplier)

	// Check credits
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %v", err)
	}
	if player.Credits < creditCost {
		return fmt.Errorf("insufficient credits (need %d)", creditCost)
	}

	// Deduct credits
	player.Credits -= creditCost
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Research technology (instant for now, could be time-based)
	playerTechs[techID]++

	log.Info("Technology researched: player=%s, tech=%s, level=%d",
		playerID, tech.Name, playerTechs[techID])

	if m.onTechResearched != nil {
		go m.onTechResearched(playerID, tech)
	}

	return nil
}

// GetPlayerTechnologies retrieves player's researched technologies
func (m *Manager) GetPlayerTechnologies(playerID uuid.UUID) map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	techs := m.playerTech[playerID]
	if techs == nil {
		return make(map[string]int)
	}

	// Return copy
	result := make(map[string]int)
	for k, v := range techs {
		result[k] = v
	}
	return result
}

// GetAllTechnologies retrieves all available technologies
func (m *Manager) GetAllTechnologies() []*Technology {
	m.mu.RLock()
	defer m.mu.RUnlock()

	techs := make([]*Technology, 0, len(m.technologies))
	for _, tech := range m.technologies {
		techs = append(techs, tech)
	}
	return techs
}

// ============================================================================
// PLAYER STATION SYSTEM
// ============================================================================

// BuildStation constructs a new player station
func (m *Manager) BuildStation(ctx context.Context, playerID uuid.UUID, name string, systemID uuid.UUID, systemName string) (*PlayerStation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check station limit
	stationCount := 0
	for _, station := range m.stations {
		if station.OwnerID == playerID && station.Status != "destroyed" {
			stationCount++
		}
	}
	if stationCount >= m.config.MaxStationsPerPlayer {
		return nil, fmt.Errorf("maximum stations reached (%d)", m.config.MaxStationsPerPlayer)
	}

	// Check credits
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %v", err)
	}
	if player.Credits < m.config.StationBuildCost {
		return nil, fmt.Errorf("insufficient credits (need %d)", m.config.StationBuildCost)
	}

	// Deduct credits
	player.Credits -= m.config.StationBuildCost
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Create station
	station := &PlayerStation{
		ID:              uuid.New(),
		OwnerID:         playerID,
		Name:            name,
		SystemID:        systemID,
		SystemName:      systemName,
		Level:           1,
		BuildTime:       time.Now(),
		Facilities:      []StationFacility{FacilityManufacturing}, // Start with basic manufacturing
		Storage:         make(map[string]int),
		StorageCapacity: m.config.StationStorageCapacity,
		ProductionBonus: m.config.StationProductionBonus,
		Status:          "active",
	}

	m.stations[station.ID] = station

	log.Info("Station built: owner=%s, name=%s, system=%s", playerID, name, systemName)

	if m.onStationBuilt != nil {
		go m.onStationBuilt(station)
	}

	return station, nil
}

// UpgradeStation upgrades a player station
func (m *Manager) UpgradeStation(ctx context.Context, stationID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	station, exists := m.stations[stationID]
	if !exists {
		return fmt.Errorf("station not found")
	}

	if station.OwnerID != playerID {
		return fmt.Errorf("not your station")
	}

	if station.Level >= 10 {
		return fmt.Errorf("station already at max level")
	}

	// Calculate upgrade cost
	upgradeCost := m.config.StationUpgradeCost * int64(station.Level)

	// Check credits
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %v", err)
	}
	if player.Credits < upgradeCost {
		return fmt.Errorf("insufficient credits (need %d)", upgradeCost)
	}

	// Deduct credits
	player.Credits -= upgradeCost
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Upgrade station
	station.Level++
	station.ProductionBonus += 0.05 // +5% per level
	station.StorageCapacity += 1000 // +1000 per level

	log.Info("Station upgraded: station=%s, level=%d", stationID, station.Level)

	if m.onStationUpgraded != nil {
		go m.onStationUpgraded(station)
	}

	return nil
}

// getFacilityCost calculates the cost of adding a facility
func (m *Manager) getFacilityCost(facility StationFacility, stationLevel int) int64 {
	// Base facility cost is 40% of upgrade cost
	baseCost := int64(float64(m.config.StationUpgradeCost) * 0.4)

	// Some facilities are more expensive
	multiplier := 1.0
	switch facility {
	case FacilityShipyard:
		multiplier = 2.0 // Most expensive
	case FacilityResearch:
		multiplier = 1.5
	case FacilityManufacturing:
		multiplier = 1.3
	case FacilityRefinery:
		multiplier = 1.2
	case FacilityDefense:
		multiplier = 1.5
	case FacilityWarehouse:
		multiplier = 0.8 // Cheapest
	}

	// Scale with station level
	levelMultiplier := 1.0 + (float64(stationLevel-1) * 0.2)

	return int64(float64(baseCost) * multiplier * levelMultiplier)
}

// AddFacility adds a facility to a station
func (m *Manager) AddFacility(ctx context.Context, stationID, playerID uuid.UUID, facility StationFacility) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	station, exists := m.stations[stationID]
	if !exists {
		return fmt.Errorf("station not found")
	}

	if station.OwnerID != playerID {
		return fmt.Errorf("not your station")
	}

	// Check if already has facility
	for _, f := range station.Facilities {
		if f == facility {
			return fmt.Errorf("station already has this facility")
		}
	}

	// Calculate cost and check requirements
	cost := m.getFacilityCost(facility, station.Level)

	// Get player to check credits
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has enough credits
	if player.Credits < cost {
		return fmt.Errorf("insufficient credits (need %d, have %d)", cost, player.Credits)
	}

	// Check station status
	if station.Status != "active" {
		return fmt.Errorf("station must be active to add facilities (current status: %s)", station.Status)
	}

	// Check facility requirements based on station level
	requiredLevel := 1
	switch facility {
	case FacilityShipyard:
		requiredLevel = 5 // Requires level 5 station
	case FacilityResearch:
		requiredLevel = 3
	case FacilityDefense:
		requiredLevel = 2
	}

	if station.Level < requiredLevel {
		return fmt.Errorf("station level %d required for %s (current level: %d)",
			requiredLevel, facility, station.Level)
	}

	// Deduct credits
	newCredits := player.Credits - cost
	if err := m.playerRepo.UpdateCredits(ctx, playerID, newCredits); err != nil {
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	// Add facility
	station.Facilities = append(station.Facilities, facility)

	// Apply facility benefits
	switch facility {
	case FacilityWarehouse:
		station.StorageCapacity += m.config.StationStorageCapacity
	case FacilityManufacturing:
		station.ProductionBonus += 0.15 // +15% production speed
	case FacilityResearch:
		station.ProductionBonus += 0.10 // +10% research speed
	}

	log.Info("Facility added: station=%s, facility=%s, cost=%d, level=%d",
		stationID, facility, cost, station.Level)
	return nil
}

// GetPlayerStations retrieves all stations owned by a player
func (m *Manager) GetPlayerStations(playerID uuid.UUID) []*PlayerStation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stations []*PlayerStation
	for _, station := range m.stations {
		if station.OwnerID == playerID {
			stations = append(stations, station)
		}
	}
	return stations
}

// ============================================================================
// BACKGROUND WORKERS
// ============================================================================

// craftingWorker handles crafting job completion
func (m *Manager) craftingWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkCraftingJobs()
		case <-m.stopChan:
			return
		}
	}
}

// checkCraftingJobs checks for completed crafting jobs
func (m *Manager) checkCraftingJobs() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for _, job := range m.craftingJobs {
		if job.Status == "in_progress" && now.After(job.CompletionTime) {
			job.Status = "complete"
			log.Info("Crafting completed: job=%s, item=%s", job.ID, job.Blueprint.Name)

			// TODO: Add crafted items to player inventory

			if m.onCraftingComplete != nil {
				go m.onCraftingComplete(job)
			}
		}
	}
}

// ============================================================================
// INITIALIZATION
// ============================================================================

// initializeBlueprints sets up default blueprints
func (m *Manager) initializeBlueprints() {
	// Example blueprints
	blueprints := []*Blueprint{
		{
			ID:           uuid.New(),
			Name:         "Basic Laser",
			Description:  "A simple laser weapon",
			ItemType:     "weapon",
			Tier:         1,
			CraftingTime: 5 * time.Minute,
			Requirements: map[string]int{"iron": 10, "copper": 5},
			Produces:     map[string]int{"basic_laser": 1},
			SkillLevel:   1,
		},
		{
			ID:           uuid.New(),
			Name:         "Shield Generator",
			Description:  "Basic shield protection",
			ItemType:     "outfit",
			Tier:         1,
			CraftingTime: 10 * time.Minute,
			Requirements: map[string]int{"titanium": 15, "crystals": 3},
			Produces:     map[string]int{"shield_generator": 1},
			SkillLevel:   2,
		},
		{
			ID:           uuid.New(),
			Name:         "Advanced Thruster",
			Description:  "High-performance engine component",
			ItemType:     "ship_component",
			Tier:         2,
			CraftingTime: 30 * time.Minute,
			Requirements: map[string]int{"platinum": 20, "rare_earth": 5},
			Produces:     map[string]int{"advanced_thruster": 1},
			SkillLevel:   3,
			TechRequired: "engines_2",
		},
	}

	for _, bp := range blueprints {
		m.blueprints[bp.ID] = bp
	}
}

// initializeTechnologies sets up technology tree
func (m *Manager) initializeTechnologies() {
	// Example technologies
	technologies := []*Technology{
		{
			ID:           "weapons_1",
			Name:         "Basic Weapons",
			Description:  "Unlock basic weapon crafting",
			Category:     TechCategoryWeapons,
			MaxLevel:     3,
			ResearchCost: 100,
			CreditCost:   10000,
			Prerequisites: []string{},
			Unlocks:      []string{"basic_laser", "basic_missile"},
			Benefits:     map[string]float64{"weapon_damage": 0.10},
		},
		{
			ID:           "weapons_2",
			Name:         "Advanced Weapons",
			Description:  "Unlock advanced weapon systems",
			Category:     TechCategoryWeapons,
			MaxLevel:     3,
			ResearchCost: 200,
			CreditCost:   25000,
			Prerequisites: []string{"weapons_1"},
			Unlocks:      []string{"plasma_cannon", "railgun"},
			Benefits:     map[string]float64{"weapon_damage": 0.20},
		},
		{
			ID:           "engines_1",
			Name:         "Engine Technology",
			Description:  "Improve ship engines",
			Category:     TechCategoryEngines,
			MaxLevel:     5,
			ResearchCost: 150,
			CreditCost:   15000,
			Prerequisites: []string{},
			Unlocks:      []string{"improved_thruster"},
			Benefits:     map[string]float64{"engine_efficiency": 0.15},
		},
		{
			ID:           "engines_2",
			Name:         "Advanced Propulsion",
			Description:  "High-performance engines",
			Category:     TechCategoryEngines,
			MaxLevel:     5,
			ResearchCost: 300,
			CreditCost:   40000,
			Prerequisites: []string{"engines_1"},
			Unlocks:      []string{"advanced_thruster"},
			Benefits:     map[string]float64{"engine_efficiency": 0.25, "max_speed": 0.20},
		},
		{
			ID:           "manufacturing_1",
			Name:         "Industrial Processes",
			Description:  "Improve manufacturing efficiency",
			Category:     TechCategoryManufacturing,
			MaxLevel:     5,
			ResearchCost: 200,
			CreditCost:   20000,
			Prerequisites: []string{},
			Unlocks:      []string{},
			Benefits:     map[string]float64{"crafting_speed": 0.20},
		},
	}

	for _, tech := range technologies {
		m.technologies[uuid.MustParse(tech.ID)] = tech
	}
}

// GetStats returns manufacturing statistics
func (m *Manager) GetStats() ManufacturingStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ManufacturingStats{
		ActiveCraftingJobs: 0,
		PlayerStations:     len(m.stations),
	}

	for _, job := range m.craftingJobs {
		if job.Status == "in_progress" {
			stats.ActiveCraftingJobs++
		}
	}

	return stats
}

// ManufacturingStats contains manufacturing statistics
type ManufacturingStats struct {
	ActiveCraftingJobs int `json:"active_crafting_jobs"`
	PlayerStations     int `json:"player_stations"`
}
