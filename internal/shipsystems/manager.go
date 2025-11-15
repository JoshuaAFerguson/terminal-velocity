// File: internal/shipsystems/manager.go
// Project: Terminal Velocity
// Description: Advanced ship systems including cloaking, jump drives, and wormholes
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package shipsystems

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("ShipSystems")

// Manager handles advanced ship systems
type Manager struct {
	mu sync.RWMutex

	// Active systems
	cloakedShips   map[uuid.UUID]*CloakStatus
	jumpDrives     map[uuid.UUID]*JumpDriveStatus
	wormholes      map[uuid.UUID]*Wormhole
	activeJumps    map[uuid.UUID]*JumpOperation

	// Configuration
	config ShipSystemsConfig

	// Repositories
	systemRepo *database.SystemRepository
	shipRepo   *database.ShipRepository

	// Callbacks
	onCloakActivated   func(shipID uuid.UUID)
	onCloakDeactivated func(shipID uuid.UUID)
	onJumpComplete     func(shipID uuid.UUID, fromSystem, toSystem uuid.UUID)
	onWormholeDiscovered func(wormhole *Wormhole)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// ShipSystemsConfig defines advanced system parameters
type ShipSystemsConfig struct {
	// Cloaking settings
	CloakEnergyDrainRate    float64       // Energy per second while cloaked
	CloakActivationCost     float64       // Initial energy cost
	CloakDetectionChance    float64       // Base chance to be detected
	CloakCooldownDuration   time.Duration // Cooldown after deactivating
	CloakMaxDuration        time.Duration // Maximum cloak duration

	// Jump drive settings
	JumpDriveFuelCost       float64       // Fuel cost per light-year
	JumpDriveChargeDuration time.Duration // Time to charge jump drive
	JumpDriveRange          float64       // Maximum jump range in light-years
	JumpDriveAccuracy       float64       // Arrival accuracy (0.0-1.0)
	JumpDriveCooldown       time.Duration // Cooldown between jumps

	// Wormhole settings
	WormholeDiscoveryChance float64       // Chance to discover while scanning
	WormholeStabilityDecay  float64       // Stability decay per day
	WormholeMinStability    float64       // Minimum stability to use
	WormholeTravelTime      time.Duration // Time to traverse wormhole
	WormholeMaxLifetime     time.Duration // Maximum wormhole lifetime

	// Advanced navigation
	AutopilotEnabled        bool          // Enable autopilot system
	AutopilotFuelEfficiency float64       // Fuel efficiency bonus (0.0-1.0)
	GravityAssistBonus      float64       // Speed bonus near planets
}

// DefaultShipSystemsConfig returns sensible defaults
func DefaultShipSystemsConfig() ShipSystemsConfig {
	return ShipSystemsConfig{
		CloakEnergyDrainRate:    5.0,
		CloakActivationCost:     20.0,
		CloakDetectionChance:    0.05,  // 5% base detection
		CloakCooldownDuration:   30 * time.Second,
		CloakMaxDuration:        5 * time.Minute,
		JumpDriveFuelCost:       10.0,
		JumpDriveChargeDuration: 10 * time.Second,
		JumpDriveRange:          100.0, // 100 light-years
		JumpDriveAccuracy:       0.95,  // 95% accurate
		JumpDriveCooldown:       60 * time.Second,
		WormholeDiscoveryChance: 0.01,  // 1% chance
		WormholeStabilityDecay:  0.10,  // 10% per day
		WormholeMinStability:    0.30,  // 30% minimum
		WormholeTravelTime:      5 * time.Second,
		WormholeMaxLifetime:     30 * 24 * time.Hour, // 30 days
		AutopilotEnabled:        true,
		AutopilotFuelEfficiency: 0.20,  // 20% fuel savings
		GravityAssistBonus:      0.15,  // 15% speed boost
	}
}

// NewManager creates a new ship systems manager
func NewManager(systemRepo *database.SystemRepository, shipRepo *database.ShipRepository) *Manager {
	return &Manager{
		cloakedShips: make(map[uuid.UUID]*CloakStatus),
		jumpDrives:   make(map[uuid.UUID]*JumpDriveStatus),
		wormholes:    make(map[uuid.UUID]*Wormhole),
		activeJumps:  make(map[uuid.UUID]*JumpOperation),
		config:       DefaultShipSystemsConfig(),
		systemRepo:   systemRepo,
		shipRepo:     shipRepo,
		stopChan:     make(chan struct{}),
	}
}

// Start begins background workers
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.maintenanceWorker()
	log.Info("Ship systems manager started")
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Ship systems manager stopped")
}

// SetCallbacks sets all system callbacks
func (m *Manager) SetCallbacks(
	onCloakActivated func(shipID uuid.UUID),
	onCloakDeactivated func(shipID uuid.UUID),
	onJumpComplete func(shipID uuid.UUID, fromSystem, toSystem uuid.UUID),
	onWormholeDiscovered func(wormhole *Wormhole),
) {
	m.onCloakActivated = onCloakActivated
	m.onCloakDeactivated = onCloakDeactivated
	m.onJumpComplete = onJumpComplete
	m.onWormholeDiscovered = onWormholeDiscovered
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// CloakStatus tracks a ship's cloak state
type CloakStatus struct {
	ShipID        uuid.UUID
	Active        bool
	ActivatedAt   time.Time
	Energy        float64 // Remaining cloak energy
	Cooldown      time.Time
	DetectionRisk float64 // Current detection chance
}

// JumpDriveStatus tracks a ship's jump drive state
type JumpDriveStatus struct {
	ShipID       uuid.UUID
	Charged      bool
	ChargingStart time.Time
	Cooldown     time.Time
	Range        float64 // Current max range
	Accuracy     float64 // Current accuracy
}

// JumpOperation represents an in-progress jump
type JumpOperation struct {
	ShipID         uuid.UUID
	FromSystemID   uuid.UUID
	ToSystemID     uuid.UUID
	Distance       float64
	StartTime      time.Time
	EstimatedArrival time.Time
	Status         string // "charging", "jumping", "complete"
}

// Wormhole represents a space-time anomaly
type Wormhole struct {
	ID           uuid.UUID
	FromSystemID uuid.UUID
	ToSystemID   uuid.UUID
	FromName     string
	ToName       string
	Stability    float64   // 0.0-1.0
	DiscoveredAt time.Time
	ExpiresAt    time.Time
	Type         WormholeType
	Status       string    // "stable", "unstable", "collapsed"
}

// WormholeType defines wormhole characteristics
type WormholeType string

const (
	WormholeStable    WormholeType = "stable"    // Reliable, long-lasting
	WormholeUnstable  WormholeType = "unstable"  // Unpredictable exit
	WormholeTemporal  WormholeType = "temporal"  // Time dilation effects
	WormholeQuantum   WormholeType = "quantum"   // Multiple exits possible
)

// ============================================================================
// CLOAKING SYSTEM
// ============================================================================

// ActivateCloak activates a ship's cloaking device
func (m *Manager) ActivateCloak(ctx context.Context, shipID uuid.UUID, ship *models.Ship) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already cloaked
	if status, exists := m.cloakedShips[shipID]; exists && status.Active {
		return fmt.Errorf("cloak already active")
	}

	// Check cooldown
	if status, exists := m.cloakedShips[shipID]; exists {
		if time.Now().Before(status.Cooldown) {
			remaining := time.Until(status.Cooldown)
			return fmt.Errorf("cloak on cooldown (%v remaining)", remaining)
		}
	}

	// Check energy (using shields as energy for now)
	if ship.Shields < m.config.CloakActivationCost {
		return fmt.Errorf("insufficient energy (need %.1f)", m.config.CloakActivationCost)
	}

	// Deduct activation cost
	ship.Shields -= m.config.CloakActivationCost
	if err := m.shipRepo.Update(ctx, ship); err != nil {
		return fmt.Errorf("failed to update ship: %v", err)
	}

	// Activate cloak
	status := &CloakStatus{
		ShipID:        shipID,
		Active:        true,
		ActivatedAt:   time.Now(),
		Energy:        100.0,
		DetectionRisk: m.config.CloakDetectionChance,
	}
	m.cloakedShips[shipID] = status

	log.Info("Cloak activated: ship=%s", shipID)

	if m.onCloakActivated != nil {
		go m.onCloakActivated(shipID)
	}

	// Start drain goroutine
	go m.cloakDrainWorker(ctx, shipID, ship)

	return nil
}

// DeactivateCloak deactivates a ship's cloak
func (m *Manager) DeactivateCloak(ctx context.Context, shipID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.cloakedShips[shipID]
	if !exists || !status.Active {
		return fmt.Errorf("cloak not active")
	}

	status.Active = false
	status.Cooldown = time.Now().Add(m.config.CloakCooldownDuration)

	log.Info("Cloak deactivated: ship=%s", shipID)

	if m.onCloakDeactivated != nil {
		go m.onCloakDeactivated(shipID)
	}

	return nil
}

// IsCloaked checks if a ship is currently cloaked
func (m *Manager) IsCloaked(shipID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.cloakedShips[shipID]
	return exists && status.Active
}

// GetCloakStatus retrieves cloak status
func (m *Manager) GetCloakStatus(shipID uuid.UUID) (*CloakStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.cloakedShips[shipID]
	return status, exists
}

// cloakDrainWorker drains cloak energy over time
func (m *Manager) cloakDrainWorker(ctx context.Context, shipID uuid.UUID, ship *models.Ship) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			status, exists := m.cloakedShips[shipID]
			if !exists || !status.Active {
				m.mu.Unlock()
				return
			}

			// Drain energy
			status.Energy -= m.config.CloakEnergyDrainRate

			// Check max duration
			if time.Since(status.ActivatedAt) >= m.config.CloakMaxDuration {
				status.Active = false
				status.Cooldown = time.Now().Add(m.config.CloakCooldownDuration)
				log.Info("Cloak deactivated (max duration): ship=%s", shipID)
				m.mu.Unlock()
				if m.onCloakDeactivated != nil {
					go m.onCloakDeactivated(shipID)
				}
				return
			}

			// Check energy depletion
			if status.Energy <= 0 {
				status.Active = false
				status.Cooldown = time.Now().Add(m.config.CloakCooldownDuration)
				log.Info("Cloak deactivated (energy depleted): ship=%s", shipID)
				m.mu.Unlock()
				if m.onCloakDeactivated != nil {
					go m.onCloakDeactivated(shipID)
				}
				return
			}

			m.mu.Unlock()

		case <-ctx.Done():
			return
		}
	}
}

// ============================================================================
// JUMP DRIVE SYSTEM
// ============================================================================

// ChargeJumpDrive begins charging a ship's jump drive
func (m *Manager) ChargeJumpDrive(ctx context.Context, shipID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already charged
	if status, exists := m.jumpDrives[shipID]; exists && status.Charged {
		return fmt.Errorf("jump drive already charged")
	}

	// Check cooldown
	if status, exists := m.jumpDrives[shipID]; exists {
		if time.Now().Before(status.Cooldown) {
			remaining := time.Until(status.Cooldown)
			return fmt.Errorf("jump drive on cooldown (%v remaining)", remaining)
		}
	}

	// Create or update status
	status := &JumpDriveStatus{
		ShipID:       shipID,
		Charged:      false,
		ChargingStart: time.Now(),
		Range:        m.config.JumpDriveRange,
		Accuracy:     m.config.JumpDriveAccuracy,
	}
	m.jumpDrives[shipID] = status

	log.Info("Jump drive charging: ship=%s", shipID)

	// Start charging goroutine
	go m.jumpDriveChargeWorker(shipID)

	return nil
}

// jumpDriveChargeWorker handles jump drive charging
func (m *Manager) jumpDriveChargeWorker(shipID uuid.UUID) {
	time.Sleep(m.config.JumpDriveChargeDuration)

	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.jumpDrives[shipID]
	if !exists {
		return
	}

	status.Charged = true
	log.Info("Jump drive charged: ship=%s", shipID)
}

// ExecuteJump performs a jump drive jump
func (m *Manager) ExecuteJump(ctx context.Context, shipID, fromSystemID, toSystemID uuid.UUID, ship *models.Ship, distance float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if jump drive is charged
	status, exists := m.jumpDrives[shipID]
	if !exists || !status.Charged {
		return fmt.Errorf("jump drive not charged")
	}

	// Check range
	if distance > status.Range {
		return fmt.Errorf("distance %.1f exceeds jump range %.1f", distance, status.Range)
	}

	// Calculate fuel cost
	fuelCost := distance * m.config.JumpDriveFuelCost
	if ship.Fuel < fuelCost {
		return fmt.Errorf("insufficient fuel (need %.1f)", fuelCost)
	}

	// Deduct fuel
	ship.Fuel -= fuelCost
	if err := m.shipRepo.Update(ctx, ship); err != nil {
		return fmt.Errorf("failed to update ship: %v", err)
	}

	// Create jump operation
	jumpTime := time.Duration(distance/10.0) * time.Second // 10 LY per second
	operation := &JumpOperation{
		ShipID:           shipID,
		FromSystemID:     fromSystemID,
		ToSystemID:       toSystemID,
		Distance:         distance,
		StartTime:        time.Now(),
		EstimatedArrival: time.Now().Add(jumpTime),
		Status:           "jumping",
	}
	m.activeJumps[shipID] = operation

	// Reset jump drive
	status.Charged = false
	status.Cooldown = time.Now().Add(m.config.JumpDriveCooldown)

	log.Info("Jump initiated: ship=%s, distance=%.1f LY", shipID, distance)

	// Start jump completion goroutine
	go m.jumpCompletionWorker(ctx, operation)

	return nil
}

// jumpCompletionWorker handles jump completion
func (m *Manager) jumpCompletionWorker(ctx context.Context, operation *JumpOperation) {
	time.Sleep(time.Until(operation.EstimatedArrival))

	m.mu.Lock()
	operation.Status = "complete"
	delete(m.activeJumps, operation.ShipID)
	m.mu.Unlock()

	log.Info("Jump completed: ship=%s", operation.ShipID)

	if m.onJumpComplete != nil {
		go m.onJumpComplete(operation.ShipID, operation.FromSystemID, operation.ToSystemID)
	}
}

// GetJumpStatus retrieves jump drive status
func (m *Manager) GetJumpStatus(shipID uuid.UUID) (*JumpDriveStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.jumpDrives[shipID]
	return status, exists
}

// GetActiveJump retrieves active jump operation
func (m *Manager) GetActiveJump(shipID uuid.UUID) (*JumpOperation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	operation, exists := m.activeJumps[shipID]
	return operation, exists
}

// ============================================================================
// WORMHOLE SYSTEM
// ============================================================================

// DiscoverWormhole attempts to discover a wormhole while scanning
func (m *Manager) DiscoverWormhole(ctx context.Context, fromSystemID, toSystemID uuid.UUID, fromName, toName string) (*Wormhole, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check discovery chance
	if rand.Float64() > m.config.WormholeDiscoveryChance {
		return nil, fmt.Errorf("no wormhole discovered")
	}

	// Determine wormhole type
	typeRoll := rand.Float64()
	var wormholeType WormholeType
	if typeRoll < 0.50 {
		wormholeType = WormholeStable
	} else if typeRoll < 0.80 {
		wormholeType = WormholeUnstable
	} else if typeRoll < 0.95 {
		wormholeType = WormholeTemporal
	} else {
		wormholeType = WormholeQuantum
	}

	// Create wormhole
	wormhole := &Wormhole{
		ID:           uuid.New(),
		FromSystemID: fromSystemID,
		ToSystemID:   toSystemID,
		FromName:     fromName,
		ToName:       toName,
		Stability:    0.80 + (rand.Float64() * 0.20), // 80-100% initial stability
		DiscoveredAt: time.Now(),
		ExpiresAt:    time.Now().Add(m.config.WormholeMaxLifetime),
		Type:         wormholeType,
		Status:       "stable",
	}

	m.wormholes[wormhole.ID] = wormhole

	log.Info("Wormhole discovered: type=%s, from=%s, to=%s", wormholeType, fromName, toName)

	if m.onWormholeDiscovered != nil {
		go m.onWormholeDiscovered(wormhole)
	}

	return wormhole, nil
}

// TraverseWormhole travels through a wormhole
func (m *Manager) TraverseWormhole(ctx context.Context, wormholeID, shipID uuid.UUID) error {
	m.mu.RLock()
	wormhole, exists := m.wormholes[wormholeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("wormhole not found")
	}

	if wormhole.Status == "collapsed" {
		return fmt.Errorf("wormhole has collapsed")
	}

	if wormhole.Stability < m.config.WormholeMinStability {
		return fmt.Errorf("wormhole too unstable (%.0f%% stability)", wormhole.Stability*100)
	}

	// Travel time
	time.Sleep(m.config.WormholeTravelTime)

	// Chance of instability causing issues
	if rand.Float64() > wormhole.Stability {
		log.Warn("Wormhole instability during transit: ship=%s", shipID)
		// Could apply damage or wrong exit
	}

	log.Info("Traversed wormhole: ship=%s, wormhole=%s", shipID, wormholeID)
	return nil
}

// GetWormhole retrieves a wormhole
func (m *Manager) GetWormhole(wormholeID uuid.UUID) (*Wormhole, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wormhole, exists := m.wormholes[wormholeID]
	return wormhole, exists
}

// GetWormholesInSystem retrieves all wormholes from a system
func (m *Manager) GetWormholesInSystem(systemID uuid.UUID) []*Wormhole {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wormholes []*Wormhole
	for _, wormhole := range m.wormholes {
		if wormhole.FromSystemID == systemID && wormhole.Status != "collapsed" {
			wormholes = append(wormholes, wormhole)
		}
	}
	return wormholes
}

// ============================================================================
// MAINTENANCE
// ============================================================================

// maintenanceWorker handles daily maintenance
func (m *Manager) maintenanceWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processMaintenance()
		case <-m.stopChan:
			return
		}
	}
}

// processMaintenance processes daily ship systems maintenance
func (m *Manager) processMaintenance() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Wormhole stability decay
	for _, wormhole := range m.wormholes {
		if wormhole.Status == "collapsed" {
			continue
		}

		// Decay stability
		wormhole.Stability -= m.config.WormholeStabilityDecay

		if wormhole.Stability <= 0 || now.After(wormhole.ExpiresAt) {
			wormhole.Status = "collapsed"
			wormhole.Stability = 0
			log.Info("Wormhole collapsed: id=%s", wormhole.ID)
		} else if wormhole.Stability < m.config.WormholeMinStability {
			wormhole.Status = "unstable"
		}
	}
}

// GetStats returns ship systems statistics
func (m *Manager) GetStats() ShipSystemsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ShipSystemsStats{
		ActiveCloaks:    0,
		ChargedJumps:    0,
		ActiveJumps:     len(m.activeJumps),
		StableWormholes: 0,
	}

	for _, status := range m.cloakedShips {
		if status.Active {
			stats.ActiveCloaks++
		}
	}

	for _, status := range m.jumpDrives {
		if status.Charged {
			stats.ChargedJumps++
		}
	}

	for _, wormhole := range m.wormholes {
		if wormhole.Status == "stable" {
			stats.StableWormholes++
		}
	}

	return stats
}

// ShipSystemsStats contains ship systems statistics
type ShipSystemsStats struct {
	ActiveCloaks    int `json:"active_cloaks"`
	ChargedJumps    int `json:"charged_jumps"`
	ActiveJumps     int `json:"active_jumps"`
	StableWormholes int `json:"stable_wormholes"`
}
