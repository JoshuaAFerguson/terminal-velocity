package missions

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles mission lifecycle and generation
type Manager struct {
	availableMissions []*models.Mission
	activeMissions    []*models.Mission
	completedMissions []*models.Mission
}

// NewManager creates a new mission manager
func NewManager() *Manager {
	return &Manager{
		availableMissions: []*models.Mission{},
		activeMissions:    []*models.Mission{},
		completedMissions: []*models.Mission{},
	}
}

// GenerateMissions creates random missions for a planet/station
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

// generateRandomMission creates a single random mission
func (m *Manager) generateRandomMission(planetID uuid.UUID, factionID string) *models.Mission {
	// Randomly select mission type
	missionTypes := []string{
		models.MissionTypeDelivery,
		models.MissionTypeCombat,
		models.MissionTypeBounty,
		models.MissionTypeTrading,
	}
	missionType := missionTypes[rand.Intn(len(missionTypes))]

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

		// Load mission cargo
		playerShip.AddCargo(mission.Cargo.CommodityID, mission.Cargo.Quantity)
	}

	// Move to active
	mission.Status = models.MissionStatusActive
	mission.AcceptedAt = time.Now()

	// Remove from available
	m.availableMissions = append(m.availableMissions[:missionIndex], m.availableMissions[missionIndex+1:]...)

	// Add to active
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

// FailMission marks a mission as failed
func (m *Manager) FailMission(missionID uuid.UUID, reason string) error {
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
			err := m.FailMission(mission.ID, "deadline expired")
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
			// Would check combat stats (not implemented yet)
		case models.MissionTypeBounty:
			// Would check if target killed (not implemented yet)
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

// ApplyMissionRewards applies credits and reputation from completed mission
func ApplyMissionRewards(player *models.Player, playerShip *models.Ship, mission *models.Mission) string {
	// Apply credit reward
	player.AddCredits(mission.Reward)

	// Apply reputation changes
	for factionID, repChange := range mission.ReputationChange {
		player.ModifyReputation(factionID, repChange)
	}

	// Remove mission cargo if delivery mission
	if mission.Type == models.MissionTypeDelivery && mission.Cargo != nil {
		playerShip.RemoveCargo(mission.Cargo.CommodityID, mission.Cargo.Quantity)
	}

	// Format reward message
	msg := fmt.Sprintf("Received %d credits", mission.Reward)
	if len(mission.ReputationChange) > 0 {
		msg += " and reputation bonuses"
	}

	return msg
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
