// File: internal/missions/manager.go
// Project: Terminal Velocity
// Description: Mission system manager
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package missions provides mission lifecycle management and generation.
//
// This package handles all aspects of the mission system including:
// - Mission generation (delivery, combat, bounty, trading)
// - Mission lifecycle (available -> active -> completed/failed)
// - Mission requirements validation
// - Mission progress tracking
// - Reward application
// - Player progression tracking (completions/failures)
//
// Version: 1.2.0
// Last Updated: 2025-01-07
package missions

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles mission lifecycle and generation.
// It maintains separate lists for available, active, and completed missions.

var log = logger.WithComponent("Missions")

type Manager struct {
	availableMissions []*models.Mission    // Missions available for acceptance
	activeMissions    []*models.Mission    // Currently active missions
	completedMissions []*models.Mission    // Completed missions (for history)
	bountyTargets     map[string]uuid.UUID // Maps target name to mission ID for kill tracking
}

// NewManager creates a new mission manager with empty mission lists.
func NewManager() *Manager {
	return &Manager{
		availableMissions: []*models.Mission{},
		activeMissions:    []*models.Mission{},
		completedMissions: []*models.Mission{},
		bountyTargets:     make(map[string]uuid.UUID),
	}
}

// GenerateMissions creates random missions for a planet/station.
// It generates 'count' number of missions and adds them to the available missions list.
//
// Parameters:
//   - ctx: Context for cancellation
//   - planetID: ID of the planet where missions are offered
//   - factionID: ID of the faction offering the missions
//   - count: Number of missions to generate
//
// Returns:
//   - Slice of generated missions
func (m *Manager) GenerateMissions(ctx context.Context, planetID uuid.UUID, factionID string, count int) []*models.Mission {
	missions := []*models.Mission{}

	for i := 0; i < count; i++ {
		mission := m.generateRandomMission(planetID, factionID)
		if mission != nil {
			missions = append(missions, mission)
			m.availableMissions = append(m.availableMissions, mission)
		}
	}

	return missions
}

// generateRandomMission creates a single random mission of a random type.
// Mission types are weighted equally (delivery, combat, bounty, trading).
//
// Parameters:
//   - planetID: Origin planet for the mission
//   - factionID: Faction offering the mission
//
// Returns:
//   - Generated mission or nil if generation failed
func (m *Manager) generateRandomMission(planetID uuid.UUID, factionID string) *models.Mission {
	// Randomly select mission type with equal weighting
	missionTypes := []string{
		models.MissionTypeDelivery,
		models.MissionTypeCombat,
		models.MissionTypeBounty,
		models.MissionTypeTrading,
	}
	missionType := missionTypes[rand.Intn(len(missionTypes))]

	// Generate mission based on selected type
	switch missionType {
	case models.MissionTypeDelivery:
		return generateDeliveryMission(planetID, factionID)
	case models.MissionTypeCombat:
		return generateCombatMission(planetID, factionID)
	case models.MissionTypeBounty:
		return generateBountyMission(planetID, factionID)
	case models.MissionTypeTrading:
		return generateTradingMission(planetID, factionID)
	}

	return nil
}

// AcceptMission moves a mission from available to active
func (m *Manager) AcceptMission(missionID uuid.UUID, player *models.Player, playerShip *models.Ship, playerShipType *models.ShipType) error {
	// Find mission in available list
	missionIndex := -1
	var mission *models.Mission

	for i, m := range m.availableMissions {
		if m.ID == missionID {
			missionIndex = i
			mission = m
			break
		}
	}

	if mission == nil {
		return fmt.Errorf("mission not found")
	}

	// Check if player can accept
	if !mission.CanAccept(player) {
		return fmt.Errorf("player does not meet mission requirements")
	}

	// Check active mission limit (5 active missions max)
	if len(m.activeMissions) >= 5 {
		return fmt.Errorf("too many active missions (max 5)")
	}

	// For delivery missions, check cargo space and load cargo
	if mission.Type == models.MissionTypeDelivery && mission.Cargo != nil {
		if !playerShip.CanAddCargo(mission.Cargo.Quantity, playerShipType) {
			return fmt.Errorf("insufficient cargo space (need %d tons)", mission.Cargo.Quantity)
		}

		// Load mission cargo into player's ship
		playerShip.AddCargo(mission.Cargo.CommodityID, mission.Cargo.Quantity)
	}

	// For bounty missions, register the target for kill tracking
	if mission.Type == models.MissionTypeBounty && mission.Target != nil {
		m.bountyTargets[*mission.Target] = mission.ID
	}

	// Move mission to active status
	mission.Status = models.MissionStatusActive
	mission.AcceptedAt = time.Now()

	// Remove from available missions list
	m.availableMissions = append(m.availableMissions[:missionIndex], m.availableMissions[missionIndex+1:]...)

	// Add to active missions list
	m.activeMissions = append(m.activeMissions, mission)

	return nil
}

// DeclineMission removes a mission from available list
func (m *Manager) DeclineMission(missionID uuid.UUID) error {
	// Find and remove mission
	for i, mission := range m.availableMissions {
		if mission.ID == missionID {
			m.availableMissions = append(m.availableMissions[:i], m.availableMissions[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("mission not found")
}

// CompleteMission marks a mission as completed and moves it to completed list
func (m *Manager) CompleteMission(missionID uuid.UUID) (*models.Mission, error) {
	// Find mission in active list
	missionIndex := -1
	var mission *models.Mission

	for i, m := range m.activeMissions {
		if m.ID == missionID {
			missionIndex = i
			mission = m
			break
		}
	}

	if mission == nil {
		return nil, fmt.Errorf("active mission not found")
	}

	// Mark as completed
	mission.Complete()

	// Remove from active
	m.activeMissions = append(m.activeMissions[:missionIndex], m.activeMissions[missionIndex+1:]...)

	// Add to completed
	m.completedMissions = append(m.completedMissions, mission)

	return mission, nil
}

// FailMission marks a mission as failed and records it in player progression.
//
// Parameters:
//   - missionID: UUID of the mission to fail
//   - reason: Reason for failure (e.g., "deadline expired", "abandoned")
//   - player: Optional player for progression tracking (can be nil)
//
// Returns:
//   - Error if mission not found
func (m *Manager) FailMission(missionID uuid.UUID, reason string, player *models.Player) error {
	// Find mission in active list
	missionIndex := -1
	var mission *models.Mission

	for i, m := range m.activeMissions {
		if m.ID == missionID {
			missionIndex = i
			mission = m
			break
		}
	}

	if mission == nil {
		return fmt.Errorf("active mission not found")
	}

	// Mark as failed
	mission.Fail()

	// Record mission failure for player progression
	if player != nil {
		player.RecordMissionFailure()
	}

	// Remove from active
	m.activeMissions = append(m.activeMissions[:missionIndex], m.activeMissions[missionIndex+1:]...)

	// Could optionally add to a failed missions list
	// For now, we just remove it

	return nil
}

// UpdateMissions checks all active missions for expiration and completion
func (m *Manager) UpdateMissions(player *models.Player) []string {
	messages := []string{}

	// Check for expired missions
	for i := len(m.activeMissions) - 1; i >= 0; i-- {
		mission := m.activeMissions[i]

		if mission.IsExpired() && mission.Status != models.MissionStatusCompleted {
			err := m.FailMission(mission.ID, "deadline expired", player)
			if err == nil {
				messages = append(messages, fmt.Sprintf("Mission '%s' failed: deadline expired", mission.Title))
			}
		}
	}

	return messages
}

// GetAvailableMissions returns all available missions
func (m *Manager) GetAvailableMissions() []*models.Mission {
	return m.availableMissions
}

// GetActiveMissions returns all active missions
func (m *Manager) GetActiveMissions() []*models.Mission {
	return m.activeMissions
}

// GetMissionByID finds a mission by ID in any list
func (m *Manager) GetMissionByID(missionID uuid.UUID) *models.Mission {
	// Check available
	for _, mission := range m.availableMissions {
		if mission.ID == missionID {
			return mission
		}
	}

	// Check active
	for _, mission := range m.activeMissions {
		if mission.ID == missionID {
			return mission
		}
	}

	// Check completed
	for _, mission := range m.completedMissions {
		if mission.ID == missionID {
			return mission
		}
	}

	return nil
}

// CheckMissionProgress checks if mission objectives have been met
func (m *Manager) CheckMissionProgress(player *models.Player, playerShip *models.Ship) []string {
	messages := []string{}

	for _, mission := range m.activeMissions {
		// Check based on mission type
		switch mission.Type {
		case models.MissionTypeDelivery:
			// Check if player is at destination with cargo
			if mission.Destination != nil && player.CurrentPlanet != nil {
				if *mission.Destination == *player.CurrentPlanet {
					// Check if player has the cargo
					if mission.Cargo != nil {
						cargoQty := playerShip.GetCommodityQuantity(mission.Cargo.CommodityID)
						if cargoQty >= mission.Cargo.Quantity {
							// Mission can be completed
							mission.Progress = mission.Quantity
						}
					}
				}
			}
		case models.MissionTypeCombat:
			// Combat missions are updated via RecordEnemyKill() when enemies are destroyed
			// Progress is tracked automatically in that method
		case models.MissionTypeBounty:
			// Bounty missions are updated via RecordEnemyKill() when the target is destroyed
			// Progress is tracked automatically in that method
		}

		// Auto-complete if progress meets quantity
		if mission.IsCompleted() && mission.Status != models.MissionStatusCompleted {
			completedMission, err := m.CompleteMission(mission.ID)
			if err == nil {
				messages = append(messages, fmt.Sprintf("Mission '%s' completed!", mission.Title))

				// Apply rewards
				rewardMsg := ApplyMissionRewards(player, playerShip, completedMission)
				if rewardMsg != "" {
					messages = append(messages, rewardMsg)
				}
			}
		}
	}

	return messages
}

// ApplyMissionRewards applies credits and reputation from completed mission.
// Handles reward distribution and cleanup for all mission types.
//
// Parameters:
//   - player: Player receiving the rewards
//   - playerShip: Player's ship (for cargo removal)
//   - mission: Completed mission
//
// Returns:
//   - Formatted message describing rewards received
func ApplyMissionRewards(player *models.Player, playerShip *models.Ship, mission *models.Mission) string {
	// Apply credit reward to player account
	player.AddCredits(mission.Reward)

	// Apply reputation changes with all affected factions
	for factionID, repChange := range mission.ReputationChange {
		player.ModifyReputation(factionID, repChange)
	}

	// Remove mission cargo if delivery mission (cargo has been delivered)
	if mission.Type == models.MissionTypeDelivery && mission.Cargo != nil {
		playerShip.RemoveCargo(mission.Cargo.CommodityID, mission.Cargo.Quantity)
	}

	// Record mission completion for player progression
	if player != nil {
		player.RecordMissionCompletion()
	}

	// Format reward message for player feedback
	msg := fmt.Sprintf("Received %d credits", mission.Reward)
	if len(mission.ReputationChange) > 0 {
		msg += " and reputation bonuses"
	}

	return msg
}

// RegisterBountyKill records a ship kill for bounty mission tracking.
// Checks if the killed ship matches any active bounty target.
//
// Parameters:
//   - targetName: Name of the killed ship/target
//   - player: Player who made the kill
//   - playerShip: Player's ship
//
// Returns:
//   - Slice of messages about completed bounty missions
func (m *Manager) RegisterBountyKill(targetName string, player *models.Player, playerShip *models.Ship) []string {
	messages := []string{}

	// Check if this target is part of an active bounty mission
	if missionID, exists := m.bountyTargets[targetName]; exists {
		// Find the mission
		for _, mission := range m.activeMissions {
			if mission.ID == missionID && mission.Type == models.MissionTypeBounty {
				// Increment mission progress (kill count)
				mission.Progress++

				// Check if bounty mission is complete
				if mission.Progress >= mission.Quantity {
					// Complete the mission
					completedMission, err := m.CompleteMission(mission.ID)
					if err == nil {
						messages = append(messages, fmt.Sprintf("Bounty completed: %s", mission.Title))

						// Apply rewards
						rewardMsg := ApplyMissionRewards(player, playerShip, completedMission)
						if rewardMsg != "" {
							messages = append(messages, rewardMsg)
						}

						// Remove from bounty targets
						delete(m.bountyTargets, targetName)
					}
				} else {
					// Partial progress message
					messages = append(messages, fmt.Sprintf("Bounty progress: %d/%d targets eliminated",
						mission.Progress, mission.Quantity))
				}
				break
			}
		}
	}

	return messages
}

// GetBountyTargets returns a list of all active bounty target names.
// Useful for UI display of active bounties.
//
// Returns:
//   - Slice of target names currently under bounty
func (m *Manager) GetBountyTargets() []string {
	targets := make([]string, 0, len(m.bountyTargets))
	for targetName := range m.bountyTargets {
		targets = append(targets, targetName)
	}
	return targets
}

// IsBountyTarget checks if a given target name is part of an active bounty.
//
// Parameters:
//   - targetName: Name to check
//
// Returns:
//   - true if target is part of an active bounty mission
func (m *Manager) IsBountyTarget(targetName string) bool {
	_, exists := m.bountyTargets[targetName]
	return exists
}

// Mission generation helpers

func generateDeliveryMission(originPlanet uuid.UUID, factionID string) *models.Mission {
	// Generate random destination (would query database in real implementation)
	destination := uuid.New()

	// Random commodity
	commodities := []string{"food", "medicine", "electronics", "luxury_goods"}
	commodity := commodities[rand.Intn(len(commodities))]

	// Random quantity (10-100 tons)
	quantity := 10 + rand.Intn(91)

	// Reward based on quantity and distance (simplified)
	reward := int64(quantity * (100 + rand.Intn(200)))

	// Deadline: 24-72 hours
	deadline := time.Now().Add(time.Duration(24+rand.Intn(49)) * time.Hour)

	mission := models.NewDeliveryMission(factionID, originPlanet, destination, commodity, quantity, reward, deadline)
	mission.Title = "Cargo Delivery: " + commodity
	mission.Description = fmt.Sprintf("Deliver %d tons of %s to the destination planet", quantity, commodity)

	// Add reputation reward
	mission.ReputationChange[factionID] = 5 + rand.Intn(11) // 5-15 rep

	return mission
}

func generateCombatMission(originPlanet uuid.UUID, factionID string) *models.Mission {
	// Enemy types
	enemies := []string{"pirate", "rogue_fighter", "rebel_ship", "hostile_patrol"}
	enemy := enemies[rand.Intn(len(enemies))]

	// Kill count (1-5)
	kills := 1 + rand.Intn(5)

	// Reward based on difficulty
	reward := int64(kills * (5000 + rand.Intn(10000)))

	// Min combat rating (5-50)
	minCombatRating := 5 + rand.Intn(46)

	mission := models.NewCombatMission(factionID, originPlanet, enemy, kills, reward, minCombatRating)
	mission.Title = "Combat Patrol: Eliminate " + enemy
	mission.Description = fmt.Sprintf("Destroy %d %s ships in this sector", kills, enemy)

	// Add reputation reward
	mission.ReputationChange[factionID] = 10 + rand.Intn(21) // 10-30 rep

	// Require minimum positive reputation
	mission.RequiredRep[factionID] = 0

	return mission
}

func generateBountyMission(originPlanet uuid.UUID, factionID string) *models.Mission {
	// Bounty targets
	targets := []string{"Pirate Captain", "Rogue Commander", "Fugitive Criminal", "Cartel Boss"}
	target := targets[rand.Intn(len(targets))]

	// Bounty reward (10K-100K)
	reward := int64(10000 + rand.Intn(90000))

	mission := &models.Mission{
		ID:               uuid.New(),
		Type:             models.MissionTypeBounty,
		Title:            "Bounty: " + target,
		Description:      fmt.Sprintf("Hunt down and eliminate the notorious %s", target),
		GiverID:          factionID,
		OriginPlanet:     originPlanet,
		Target:           &target,
		Quantity:         1, // Kill one target
		Reward:           reward,
		Deadline:         time.Now().Add(72 * time.Hour), // 3 days
		Status:           models.MissionStatusAvailable,
		Progress:         0,
		MinCombatRating:  20 + rand.Intn(31),                            // 20-50
		ReputationChange: map[string]int{factionID: 20 + rand.Intn(31)}, // 20-50 rep
		RequiredRep:      map[string]int{factionID: 25},                 // Need decent rep
	}

	return mission
}

func generateTradingMission(originPlanet uuid.UUID, factionID string) *models.Mission {
	// Trading goods
	goods := []string{"rare_metals", "gems", "art", "antiques"}
	good := goods[rand.Intn(len(goods))]

	// Quantity (5-50 tons)
	quantity := 5 + rand.Intn(46)

	// High reward for trading missions
	reward := int64(quantity * (500 + rand.Intn(1000)))

	mission := &models.Mission{
		ID:               uuid.New(),
		Type:             models.MissionTypeTrading,
		Title:            "Trading Contract: " + good,
		Description:      fmt.Sprintf("Purchase and deliver %d tons of %s for profit", quantity, good),
		GiverID:          factionID,
		OriginPlanet:     originPlanet,
		Target:           &good,
		Quantity:         quantity,
		Reward:           reward,
		Deadline:         time.Now().Add(48 * time.Hour), // 2 days
		Status:           models.MissionStatusAvailable,
		Progress:         0,
		ReputationChange: map[string]int{factionID: 5 + rand.Intn(11)}, // 5-15 rep
		RequiredRep:      map[string]int{},
	}

	return mission
}

// RecordEnemyKill updates progress for active combat and bounty missions when an enemy is destroyed.
// Should be called by the combat system after each enemy kill.
//
// Parameters:
//   - enemyType: The type of enemy destroyed (e.g., "Pirate Fighter", "Bounty Hunter")
//   - enemyName: The name of the specific enemy (for bounty missions, empty string for generic enemies)
//
// Returns:
//   - Array of messages about mission progress updates
func (m *Manager) RecordEnemyKill(enemyType string, enemyName string) []string {
	messages := []string{}

	// Check all active missions
	for _, mission := range m.activeMissions {
		// Skip if mission is already completed or failed
		if mission.Status != models.MissionStatusActive {
			continue
		}

		// Update combat missions
		if mission.Type == models.MissionTypeCombat {
			// Check if enemy type matches mission target
			if mission.Target != nil && *mission.Target == enemyType {
				mission.Progress++
				if mission.Progress >= mission.Quantity {
					messages = append(messages, fmt.Sprintf("Combat mission '%s' objective complete! (%d/%d)",
						mission.Title, mission.Progress, mission.Quantity))
				} else {
					messages = append(messages, fmt.Sprintf("Mission progress: %d/%d %s destroyed",
						mission.Progress, mission.Quantity, enemyType))
				}
			}
		}

		// Update bounty missions
		if mission.Type == models.MissionTypeBounty {
			// Check if this is the specific target for the bounty
			if mission.Target != nil && enemyName != "" && *mission.Target == enemyName {
				mission.Progress = 1 // Bounty missions are typically single-target
				messages = append(messages, fmt.Sprintf("Bounty target '%s' eliminated! Mission '%s' complete!",
					enemyName, mission.Title))
			}
		}
	}

	return messages
}
