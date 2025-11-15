// File: internal/capture/manager.go
// Project: Terminal Velocity
// Description: Ship capture and boarding system manager
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package capture

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

var log = logger.WithComponent("Capture")

// Manager handles ship capture and boarding operations
type Manager struct {
	mu       sync.RWMutex
	shipRepo *database.ShipRepository

	// Active boarding attempts
	activeBoardings map[uuid.UUID]*BoardingAttempt

	// Configuration
	config CaptureConfig
}

// CaptureConfig defines capture system parameters
type CaptureConfig struct {
	// Disable thresholds
	DisableHullThreshold   float64 // Target must be below this % hull
	DisableShieldThreshold float64 // Target must be below this % shields

	// Boarding chances
	BaseBoardingChance float64 // Base chance to successfully board
	CrewBonus          float64 // Bonus per crew member
	MarineBonus        float64 // Bonus for specialized marines
	DefenseBonus       float64 // Defense bonus per defender
	ShipSizeModifier   float64 // Modifier based on ship size difference

	// Capture chances
	BaseCaptureChance   float64 // Base chance to capture after boarding
	DamageModifier      float64 // Modifier based on hull damage
	CrewRatioModifier   float64 // Modifier based on crew vs defenders
	LoyaltyResistance   float64 // Enemy crew loyalty resistance
	EliteCrewBonus      float64 // Bonus for elite boarding crews
	BribeEffectiveness  float64 // Effectiveness of bribing crew

	// Time limits
	BoardingDuration time.Duration // How long a boarding attempt takes
	CooldownDuration time.Duration // Cooldown between attempts
}

// DefaultCaptureConfig returns sensible default configuration
func DefaultCaptureConfig() CaptureConfig {
	return CaptureConfig{
		DisableHullThreshold:   0.25,  // Must be below 25% hull
		DisableShieldThreshold: 0.10,  // Must be below 10% shields
		BaseBoardingChance:     0.40,  // 40% base chance
		CrewBonus:              0.05,  // +5% per crew member
		MarineBonus:            0.10,  // +10% per marine
		DefenseBonus:           0.07,  // +7% per defender
		ShipSizeModifier:       0.15,  // ±15% per size difference
		BaseCaptureChance:      0.50,  // 50% base chance
		DamageModifier:         0.20,  // +20% per 10% damage
		CrewRatioModifier:      0.25,  // +25% per 2:1 crew ratio
		LoyaltyResistance:      0.15,  // -15% from crew loyalty
		EliteCrewBonus:         0.20,  // +20% for elite crews
		BribeEffectiveness:     0.10,  // +10% if bribed
		BoardingDuration:       30 * time.Second,
		CooldownDuration:       60 * time.Second,
	}
}

// BoardingAttempt tracks an active boarding attempt
type BoardingAttempt struct {
	AttackerID   uuid.UUID
	DefenderID   uuid.UUID
	AttackerShip *models.Ship
	DefenderShip *models.Ship
	StartTime    time.Time
	Status       string // "in_progress", "success", "failed"
	Outcome      *BoardingOutcome
}

// BoardingOutcome represents the result of a boarding attempt
type BoardingOutcome struct {
	Success        bool
	CaptureSuccess bool
	AttackerLosses int // Crew lost
	DefenderLosses int // Crew lost
	ShipDamage     int // Additional damage to target ship
	Message        string
}

// NewManager creates a new capture manager
func NewManager(shipRepo *database.ShipRepository) *Manager {
	return &Manager{
		shipRepo:        shipRepo,
		activeBoardings: make(map[uuid.UUID]*BoardingAttempt),
		config:          DefaultCaptureConfig(),
	}
}

// CanDisable checks if a ship can be disabled (pre-boarding check)
func (m *Manager) CanDisable(target *models.Ship) (bool, string) {
	// Get ship type for max values
	shipType := models.GetShipTypeByID(target.TypeID)
	if shipType == nil {
		return false, "Invalid ship type"
	}

	// Check shields
	shieldPercent := float64(target.Shields) / float64(shipType.MaxShields)
	if shieldPercent > m.config.DisableShieldThreshold {
		return false, fmt.Sprintf("Target shields too high (%.0f%% > %.0f%%)",
			shieldPercent*100, m.config.DisableShieldThreshold*100)
	}

	// Check hull
	hullPercent := float64(target.Hull) / float64(shipType.MaxHull)
	if hullPercent > m.config.DisableHullThreshold {
		return false, fmt.Sprintf("Target hull too high (%.0f%% > %.0f%%)",
			hullPercent*100, m.config.DisableHullThreshold*100)
	}

	return true, "Target can be disabled"
}

// AttemptBoarding initiates a boarding attempt
func (m *Manager) AttemptBoarding(ctx context.Context, attackerShip, defenderShip *models.Ship, attackerCrew, defenderCrew int) (*BoardingAttempt, error) {
	// Check if target can be disabled
	canDisable, reason := m.CanDisable(defenderShip)
	if !canDisable {
		return nil, fmt.Errorf("cannot board: %s", reason)
	}

	// Check for existing boarding attempt
	m.mu.RLock()
	if existing, exists := m.activeBoardings[attackerShip.ID]; exists {
		m.mu.RUnlock()
		return nil, fmt.Errorf("already boarding a ship (started %v ago)", time.Since(existing.StartTime))
	}
	m.mu.RUnlock()

	// Create boarding attempt
	attempt := &BoardingAttempt{
		AttackerID:   attackerShip.OwnerID,
		DefenderID:   defenderShip.OwnerID,
		AttackerShip: attackerShip,
		DefenderShip: defenderShip,
		StartTime:    time.Now(),
		Status:       "in_progress",
	}

	m.mu.Lock()
	m.activeBoardings[attackerShip.ID] = attempt
	m.mu.Unlock()

	log.Info("Boarding initiated: attacker=%s, defender=%s, crew=%d vs %d",
		attackerShip.ID, defenderShip.ID, attackerCrew, defenderCrew)

	// Resolve boarding after duration
	go m.resolveBoardingAfterDelay(ctx, attempt, attackerCrew, defenderCrew)

	return attempt, nil
}

// resolveBoardingAfterDelay resolves a boarding attempt after the configured duration
func (m *Manager) resolveBoardingAfterDelay(ctx context.Context, attempt *BoardingAttempt, attackerCrew, defenderCrew int) {
	// Wait for boarding duration
	time.Sleep(m.config.BoardingDuration)

	// Resolve boarding
	outcome := m.resolveBoarding(attempt, attackerCrew, defenderCrew)
	attempt.Outcome = outcome

	if outcome.Success {
		attempt.Status = "success"
		log.Info("Boarding successful: attacker=%s, losses=%d/%d",
			attempt.AttackerID, outcome.AttackerLosses, attackerCrew)

		// Attempt to capture ship
		if outcome.CaptureSuccess {
			m.captureShip(ctx, attempt)
		}
	} else {
		attempt.Status = "failed"
		log.Info("Boarding failed: attacker=%s, losses=%d/%d",
			attempt.AttackerID, outcome.AttackerLosses, attackerCrew)
	}

	// Clean up after cooldown
	time.Sleep(m.config.CooldownDuration)
	m.mu.Lock()
	delete(m.activeBoardings, attempt.AttackerShip.ID)
	m.mu.Unlock()
}

// resolveBoarding calculates the outcome of a boarding attempt
func (m *Manager) resolveBoarding(attempt *BoardingAttempt, attackerCrew, defenderCrew int) *BoardingOutcome {
	outcome := &BoardingOutcome{}

	// Calculate boarding success chance
	chance := m.config.BaseBoardingChance

	// Crew bonuses
	chance += float64(attackerCrew) * m.config.CrewBonus
	chance -= float64(defenderCrew) * m.config.DefenseBonus

	// Ship size modifier (larger ships easier to board)
	attackerType := models.GetShipTypeByID(attempt.AttackerShip.TypeID)
	defenderType := models.GetShipTypeByID(attempt.DefenderShip.TypeID)
	if attackerType != nil && defenderType != nil {
		sizeDiff := defenderType.CargoSpace - attackerType.CargoSpace
		chance += float64(sizeDiff) / 100 * m.config.ShipSizeModifier
	}

	// Clamp chance between 5% and 95%
	if chance < 0.05 {
		chance = 0.05
	}
	if chance > 0.95 {
		chance = 0.95
	}

	// Roll for success
	roll := rand.Float64()
	outcome.Success = roll < chance

	// Calculate casualties
	if outcome.Success {
		// Successful boarding - lighter losses
		outcome.AttackerLosses = int(float64(attackerCrew) * 0.15) // 15% losses
		outcome.DefenderLosses = int(float64(defenderCrew) * 0.40) // 40% losses
		outcome.Message = fmt.Sprintf("Boarding successful! Lost %d crew, eliminated %d defenders",
			outcome.AttackerLosses, outcome.DefenderLosses)

		// Attempt capture
		outcome.CaptureSuccess = m.calculateCaptureSuccess(attempt, attackerCrew, defenderCrew)
		if outcome.CaptureSuccess {
			outcome.Message = "Ship captured! The vessel is now yours."
		}
	} else {
		// Failed boarding - heavier losses
		outcome.AttackerLosses = int(float64(attackerCrew) * 0.30) // 30% losses
		outcome.DefenderLosses = int(float64(defenderCrew) * 0.20) // 20% losses
		outcome.ShipDamage = 50 // Additional damage from failed boarding
		outcome.Message = fmt.Sprintf("Boarding failed! Lost %d crew, defenders lost %d",
			outcome.AttackerLosses, outcome.DefenderLosses)
	}

	return outcome
}

// calculateCaptureSuccess determines if the ship is captured after successful boarding
func (m *Manager) calculateCaptureSuccess(attempt *BoardingAttempt, attackerCrew, defenderCrew int) bool {
	chance := m.config.BaseCaptureChance

	// Damage modifier - more damage = easier to capture
	defenderType := models.GetShipTypeByID(attempt.DefenderShip.TypeID)
	if defenderType != nil {
		hullPercent := float64(attempt.DefenderShip.Hull) / float64(defenderType.MaxHull)
		damagePercent := 1.0 - hullPercent
		chance += damagePercent * m.config.DamageModifier
	}

	// Crew ratio modifier
	crewRatio := float64(attackerCrew) / float64(defenderCrew)
	if crewRatio > 2.0 {
		chance += m.config.CrewRatioModifier
	}

	// Defender loyalty resistance
	chance -= m.config.LoyaltyResistance

	// Clamp between 10% and 90%
	if chance < 0.10 {
		chance = 0.10
	}
	if chance > 0.90 {
		chance = 0.90
	}

	roll := rand.Float64()
	return roll < chance
}

// captureShip transfers ownership of a ship to the attacker
func (m *Manager) captureShip(ctx context.Context, attempt *BoardingAttempt) error {
	// Transfer ship ownership
	attempt.DefenderShip.OwnerID = attempt.AttackerID

	// Set hull to 50% after capture
	defenderType := models.GetShipTypeByID(attempt.DefenderShip.TypeID)
	if defenderType != nil {
		attempt.DefenderShip.Hull = int(float64(defenderType.MaxHull) * 0.5)
	}

	// Update ship in database
	err := m.shipRepo.Update(ctx, attempt.DefenderShip)
	if err != nil {
		log.Error("Failed to capture ship: %v", err)
		return fmt.Errorf("failed to capture ship: %w", err)
	}

	log.Info("Ship captured: ship=%s, new_owner=%s",
		attempt.DefenderShip.ID, attempt.AttackerID)

	return nil
}

// GetActiveBoardingAttempt retrieves an active boarding attempt
func (m *Manager) GetActiveBoardingAttempt(shipID uuid.UUID) (*BoardingAttempt, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attempt, exists := m.activeBoardings[shipID]
	return attempt, exists
}

// CancelBoardingAttempt cancels an active boarding attempt
func (m *Manager) CancelBoardingAttempt(shipID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	attempt, exists := m.activeBoardings[shipID]
	if !exists {
		return fmt.Errorf("no active boarding attempt")
	}

	if attempt.Status != "in_progress" {
		return fmt.Errorf("boarding attempt already resolved")
	}

	delete(m.activeBoardings, shipID)
	log.Info("Boarding attempt cancelled: attacker=%s", shipID)

	return nil
}

// CalculateBoardingChance estimates the chance of a successful boarding
func (m *Manager) CalculateBoardingChance(attackerShip, defenderShip *models.Ship, attackerCrew, defenderCrew int) float64 {
	chance := m.config.BaseBoardingChance

	// Crew bonuses
	chance += float64(attackerCrew) * m.config.CrewBonus
	chance -= float64(defenderCrew) * m.config.DefenseBonus

	// Ship size modifier
	attackerType := models.GetShipTypeByID(attackerShip.TypeID)
	defenderType := models.GetShipTypeByID(defenderShip.TypeID)
	if attackerType != nil && defenderType != nil {
		sizeDiff := defenderType.CargoSpace - attackerType.CargoSpace
		chance += float64(sizeDiff) / 100 * m.config.ShipSizeModifier
	}

	// Clamp
	if chance < 0.05 {
		chance = 0.05
	}
	if chance > 0.95 {
		chance = 0.95
	}

	return chance
}

// CalculateCaptureChance estimates the chance of capturing a ship after successful boarding
func (m *Manager) CalculateCaptureChance(defenderShip *models.Ship, attackerCrew, defenderCrew int) float64 {
	chance := m.config.BaseCaptureChance

	// Damage modifier
	defenderType := models.GetShipTypeByID(defenderShip.TypeID)
	if defenderType != nil {
		hullPercent := float64(defenderShip.Hull) / float64(defenderType.MaxHull)
		damagePercent := 1.0 - hullPercent
		chance += damagePercent * m.config.DamageModifier
	}

	// Crew ratio modifier
	crewRatio := float64(attackerCrew) / float64(defenderCrew)
	if crewRatio > 2.0 {
		chance += m.config.CrewRatioModifier
	}

	// Loyalty resistance
	chance -= m.config.LoyaltyResistance

	// Clamp
	if chance < 0.10 {
		chance = 0.10
	}
	if chance > 0.90 {
		chance = 0.90
	}

	return chance
}

// GetActiveBoardingCount returns the number of active boarding attempts
func (m *Manager) GetActiveBoardingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.activeBoardings)
}

// Statistics returns capture system statistics
type CaptureStats struct {
	ActiveBoardings int
	TotalAttempts   int
	SuccessfulBoards int
	SuccessfulCaptures int
}

// GetStats returns capture statistics (placeholder for now)
func (m *Manager) GetStats() CaptureStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return CaptureStats{
		ActiveBoardings: len(m.activeBoardings),
		// TODO: Track these in database
		TotalAttempts:      0,
		SuccessfulBoards:   0,
		SuccessfulCaptures: 0,
	}
}
