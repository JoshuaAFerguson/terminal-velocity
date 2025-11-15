// File: internal/fleet/manager.go
// Project: Terminal Velocity
// Description: Fleet management system for multi-ship ownership and escorts
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package fleet

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

var log = logger.WithComponent("Fleet")

// Manager handles fleet operations including escorts and ship storage
type Manager struct {
	mu sync.RWMutex

	// Active fleets
	fleets map[uuid.UUID]*Fleet // player_id -> fleet

	// Configuration
	config FleetConfig

	// Repositories
	playerRepo *database.PlayerRepository
	shipRepo   *database.ShipRepository

	// Callbacks
	onEscortDestroyed func(playerID uuid.UUID, escort *Escort)
	onFleetCommand    func(fleet *Fleet, command string)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// FleetConfig defines fleet system parameters
type FleetConfig struct {
	// Fleet limits
	MaxShipsPerPlayer    int // Maximum ships a player can own
	MaxEscorts           int // Maximum active escorts
	MaxStoredShips       int // Ships that can be stored at planets

	// Escort settings
	EscortHireCost       int64         // Cost to hire an escort
	EscortMaintenanceCost int64        // Cost per day to maintain escort
	EscortLoyaltyDecay   float64       // Loyalty decay per day (0.0-1.0)
	MinLoyaltyThreshold  float64       // Minimum loyalty before desertion (0.0-1.0)

	// Formation settings
	FormationDistance    float64 // Distance escorts maintain from flagship
	FormationTightness   float64 // How tightly escorts follow (0.0-1.0)

	// AI settings
	EscortResponseTime   time.Duration // How quickly escorts react to commands
	EscortTargetSwitch   float64       // Chance to switch targets (0.0-1.0)
	AutoDefendPlayer     bool          // Escorts auto-defend player
}

// DefaultFleetConfig returns sensible defaults
func DefaultFleetConfig() FleetConfig {
	return FleetConfig{
		MaxShipsPerPlayer:    5,
		MaxEscorts:           2,
		MaxStoredShips:       10,
		EscortHireCost:       50000,
		EscortMaintenanceCost: 5000,
		EscortLoyaltyDecay:   0.05,  // 5% per day
		MinLoyaltyThreshold:  0.30,  // 30% loyalty minimum
		FormationDistance:    100.0,
		FormationTightness:   0.75,
		EscortResponseTime:   2 * time.Second,
		EscortTargetSwitch:   0.20,  // 20% chance
		AutoDefendPlayer:     true,
	}
}

// NewManager creates a new fleet manager
func NewManager(playerRepo *database.PlayerRepository, shipRepo *database.ShipRepository) *Manager {
	return &Manager{
		fleets:     make(map[uuid.UUID]*Fleet),
		config:     DefaultFleetConfig(),
		playerRepo: playerRepo,
		shipRepo:   shipRepo,
		stopChan:   make(chan struct{}),
	}
}

// Start begins background workers for fleet management
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.maintenanceWorker()
	log.Info("Fleet manager started")
}

// Stop gracefully shuts down the fleet manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Fleet manager stopped")
}

// SetEscortDestroyedCallback sets callback for escort destruction
func (m *Manager) SetEscortDestroyedCallback(callback func(playerID uuid.UUID, escort *Escort)) {
	m.onEscortDestroyed = callback
}

// SetFleetCommandCallback sets callback for fleet commands
func (m *Manager) SetFleetCommandCallback(callback func(fleet *Fleet, command string)) {
	m.onFleetCommand = callback
}

// ============================================================================
// FLEET STRUCTURE
// ============================================================================

// Fleet represents a player's complete fleet
type Fleet struct {
	PlayerID     uuid.UUID
	FlagshipID   uuid.UUID // Currently active ship
	OwnedShips   []*models.Ship
	StoredShips  []*StoredShip
	Escorts      []*Escort
	Formation    FormationType
	AutoDefend   bool
	LastMaintenance time.Time
}

// StoredShip represents a ship stored at a planet
type StoredShip struct {
	Ship      *models.Ship
	LocationID uuid.UUID // Planet ID where stored
	Location   string    // Planet name
	StoredAt   time.Time
	StorageFee int64     // Fee paid for storage
}

// Escort represents an AI-controlled wingman ship
type Escort struct {
	ID          uuid.UUID
	Ship        *models.Ship
	OwnerID     uuid.UUID
	Pilot       string    // NPC pilot name
	Loyalty     float64   // 0.0 - 1.0
	HiredAt     time.Time
	Level       int       // Skill level 1-10
	Behavior    EscortBehavior
	CurrentTarget uuid.UUID
	Status      string    // "active", "defending", "attacking", "idle"
}

// EscortBehavior defines how an escort acts
type EscortBehavior string

const (
	BehaviorDefensive  EscortBehavior = "defensive"  // Only engage when attacked
	BehaviorAggressive EscortBehavior = "aggressive" // Engage all hostiles
	BehaviorPassive    EscortBehavior = "passive"    // Never engage
	BehaviorSupport    EscortBehavior = "support"    // Heal/buff player
)

// FormationType defines fleet formation
type FormationType string

const (
	FormationLine     FormationType = "line"     // Single file line
	FormationWedge    FormationType = "wedge"    // V formation
	FormationBox      FormationType = "box"      // Square formation
	FormationCircular FormationType = "circular" // Circle around flagship
)

// ============================================================================
// FLEET MANAGEMENT
// ============================================================================

// GetOrCreateFleet gets or creates a fleet for a player
func (m *Manager) GetOrCreateFleet(playerID uuid.UUID) *Fleet {
	m.mu.Lock()
	defer m.mu.Unlock()

	if fleet, exists := m.fleets[playerID]; exists {
		return fleet
	}

	fleet := &Fleet{
		PlayerID:     playerID,
		OwnedShips:   []*models.Ship{},
		StoredShips:  []*StoredShip{},
		Escorts:      []*Escort{},
		Formation:    FormationLine,
		AutoDefend:   m.config.AutoDefendPlayer,
		LastMaintenance: time.Now(),
	}

	m.fleets[playerID] = fleet
	return fleet
}

// AddShip adds a ship to the player's fleet
func (m *Manager) AddShip(ctx context.Context, playerID uuid.UUID, ship *models.Ship) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet := m.fleets[playerID]
	if fleet == nil {
		fleet = &Fleet{
			PlayerID:   playerID,
			OwnedShips: []*models.Ship{},
		}
		m.fleets[playerID] = fleet
	}

	// Check ship limit
	if len(fleet.OwnedShips) >= m.config.MaxShipsPerPlayer {
		return fmt.Errorf("maximum ships owned (%d)", m.config.MaxShipsPerPlayer)
	}

	// Add ship
	fleet.OwnedShips = append(fleet.OwnedShips, ship)

	// Set as flagship if first ship
	if len(fleet.OwnedShips) == 1 {
		fleet.FlagshipID = ship.ID
	}

	log.Info("Ship added to fleet: player=%s, ship=%s", playerID, ship.ID)
	return nil
}

// SwitchFlagship switches the active ship
func (m *Manager) SwitchFlagship(ctx context.Context, playerID, shipID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	// Check if ship is owned
	shipFound := false
	for _, ship := range fleet.OwnedShips {
		if ship.ID == shipID {
			shipFound = true
			break
		}
	}

	if !shipFound {
		return fmt.Errorf("ship not found in fleet")
	}

	// Switch flagship
	fleet.FlagshipID = shipID
	log.Info("Flagship switched: player=%s, new_flagship=%s", playerID, shipID)
	return nil
}

// StoreShip stores a ship at the current planet
func (m *Manager) StoreShip(ctx context.Context, playerID, shipID, planetID uuid.UUID, planetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	// Cannot store flagship
	if shipID == fleet.FlagshipID {
		return fmt.Errorf("cannot store active flagship")
	}

	// Check storage limit
	if len(fleet.StoredShips) >= m.config.MaxStoredShips {
		return fmt.Errorf("maximum stored ships reached (%d)", m.config.MaxStoredShips)
	}

	// Find and remove ship from owned ships
	shipIndex := -1
	var ship *models.Ship
	for i, s := range fleet.OwnedShips {
		if s.ID == shipID {
			ship = s
			shipIndex = i
			break
		}
	}

	if shipIndex == -1 {
		return fmt.Errorf("ship not found in fleet")
	}

	// Remove from owned ships
	fleet.OwnedShips = append(fleet.OwnedShips[:shipIndex], fleet.OwnedShips[shipIndex+1:]...)

	// Add to stored ships
	storedShip := &StoredShip{
		Ship:       ship,
		LocationID: planetID,
		Location:   planetName,
		StoredAt:   time.Now(),
		StorageFee: 0, // Could charge storage fee
	}
	fleet.StoredShips = append(fleet.StoredShips, storedShip)

	log.Info("Ship stored: player=%s, ship=%s, planet=%s", playerID, shipID, planetName)
	return nil
}

// RetrieveShip retrieves a stored ship (must be at same planet)
func (m *Manager) RetrieveShip(ctx context.Context, playerID, shipID, currentPlanetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	// Find stored ship
	shipIndex := -1
	var storedShip *StoredShip
	for i, s := range fleet.StoredShips {
		if s.Ship.ID == shipID {
			storedShip = s
			shipIndex = i
			break
		}
	}

	if shipIndex == -1 {
		return fmt.Errorf("stored ship not found")
	}

	// Check if at correct planet
	if storedShip.LocationID != currentPlanetID {
		return fmt.Errorf("ship is stored at %s, you are not there", storedShip.Location)
	}

	// Remove from stored ships
	fleet.StoredShips = append(fleet.StoredShips[:shipIndex], fleet.StoredShips[shipIndex+1:]...)

	// Add to owned ships
	fleet.OwnedShips = append(fleet.OwnedShips, storedShip.Ship)

	log.Info("Ship retrieved: player=%s, ship=%s", playerID, shipID)
	return nil
}

// ============================================================================
// ESCORT SYSTEM
// ============================================================================

// HireEscort hires an NPC pilot to fly an escort ship
func (m *Manager) HireEscort(ctx context.Context, playerID uuid.UUID, ship *models.Ship, pilotName string, behavior EscortBehavior) (*Escort, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return nil, fmt.Errorf("fleet not found")
	}

	// Check escort limit
	if len(fleet.Escorts) >= m.config.MaxEscorts {
		return nil, fmt.Errorf("maximum escorts reached (%d)", m.config.MaxEscorts)
	}

	// Check hiring cost
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %v", err)
	}
	if player.Credits < m.config.EscortHireCost {
		return nil, fmt.Errorf("insufficient credits (need %d)", m.config.EscortHireCost)
	}

	// Deduct hiring cost
	player.Credits -= m.config.EscortHireCost
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return nil, fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Create escort
	escort := &Escort{
		ID:       uuid.New(),
		Ship:     ship,
		OwnerID:  playerID,
		Pilot:    pilotName,
		Loyalty:  1.0, // Start at max loyalty
		HiredAt:  time.Now(),
		Level:    rand.Intn(5) + 1, // Random level 1-5
		Behavior: behavior,
		Status:   "idle",
	}

	fleet.Escorts = append(fleet.Escorts, escort)

	log.Info("Escort hired: player=%s, pilot=%s, ship=%s", playerID, pilotName, ship.ID)
	return escort, nil
}

// DismissEscort dismisses an escort (no refund)
func (m *Manager) DismissEscort(ctx context.Context, playerID, escortID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	// Find and remove escort
	escortIndex := -1
	for i, escort := range fleet.Escorts {
		if escort.ID == escortID {
			escortIndex = i
			break
		}
	}

	if escortIndex == -1 {
		return fmt.Errorf("escort not found")
	}

	fleet.Escorts = append(fleet.Escorts[:escortIndex], fleet.Escorts[escortIndex+1:]...)

	log.Info("Escort dismissed: player=%s, escort=%s", playerID, escortID)
	return nil
}

// SetEscortBehavior changes an escort's behavior
func (m *Manager) SetEscortBehavior(playerID, escortID uuid.UUID, behavior EscortBehavior) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	for _, escort := range fleet.Escorts {
		if escort.ID == escortID {
			escort.Behavior = behavior
			log.Info("Escort behavior changed: escort=%s, behavior=%s", escortID, behavior)
			return nil
		}
	}

	return fmt.Errorf("escort not found")
}

// SetFormation changes the fleet formation
func (m *Manager) SetFormation(playerID uuid.UUID, formation FormationType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	fleet.Formation = formation
	log.Info("Formation changed: player=%s, formation=%s", playerID, formation)
	return nil
}

// CommandEscorts issues a command to all escorts
func (m *Manager) CommandEscorts(playerID uuid.UUID, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	// Process command
	switch command {
	case "attack":
		for _, escort := range fleet.Escorts {
			if escort.Behavior != BehaviorPassive {
				escort.Status = "attacking"
			}
		}
	case "defend":
		for _, escort := range fleet.Escorts {
			escort.Status = "defending"
		}
	case "hold":
		for _, escort := range fleet.Escorts {
			escort.Status = "idle"
		}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	log.Info("Fleet command issued: player=%s, command=%s", playerID, command)

	if m.onFleetCommand != nil {
		go m.onFleetCommand(fleet, command)
	}

	return nil
}

// ============================================================================
// ESCORT AI
// ============================================================================

// UpdateEscortAI updates escort AI behavior (called during combat)
func (m *Manager) UpdateEscortAI(ctx context.Context, playerID uuid.UUID, availableTargets []uuid.UUID) {
	m.mu.RLock()
	fleet, exists := m.fleets[playerID]
	m.mu.RUnlock()

	if !exists || len(fleet.Escorts) == 0 {
		return
	}

	for _, escort := range fleet.Escorts {
		m.updateSingleEscort(escort, availableTargets)
	}
}

// updateSingleEscort updates a single escort's AI
func (m *Manager) updateSingleEscort(escort *Escort, availableTargets []uuid.UUID) {
	// Passive escorts don't engage
	if escort.Behavior == BehaviorPassive {
		escort.Status = "idle"
		return
	}

	// If no current target or should switch
	if escort.CurrentTarget == uuid.Nil || rand.Float64() < m.config.EscortTargetSwitch {
		// Pick random target
		if len(availableTargets) > 0 {
			escort.CurrentTarget = availableTargets[rand.Intn(len(availableTargets))]
			escort.Status = "attacking"
		} else {
			escort.CurrentTarget = uuid.Nil
			escort.Status = "idle"
		}
	}

	// Defensive escorts only engage when player is attacked
	if escort.Behavior == BehaviorDefensive && escort.Status != "defending" {
		if len(availableTargets) > 0 && m.config.AutoDefendPlayer {
			escort.Status = "defending"
			escort.CurrentTarget = availableTargets[0] // Attack first threat
		}
	}
}

// GetEscortAction determines what action an escort should take
func (m *Manager) GetEscortAction(escort *Escort) string {
	switch escort.Behavior {
	case BehaviorAggressive:
		if escort.CurrentTarget != uuid.Nil {
			return "attack"
		}
		return "hold"
	case BehaviorDefensive:
		if escort.Status == "defending" {
			return "attack"
		}
		return "defend"
	case BehaviorSupport:
		return "support" // Could heal/buff player
	case BehaviorPassive:
		return "hold"
	default:
		return "hold"
	}
}

// ============================================================================
// MAINTENANCE
// ============================================================================

// maintenanceWorker handles daily maintenance costs and loyalty decay
func (m *Manager) maintenanceWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(24 * time.Hour) // Daily maintenance
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

// processMaintenance processes daily maintenance for all fleets
func (m *Manager) processMaintenance() {
	ctx := context.Background()

	m.mu.Lock()
	defer m.mu.Unlock()

	for playerID, fleet := range m.fleets {
		// Skip if no escorts
		if len(fleet.Escorts) == 0 {
			continue
		}

		// Calculate maintenance cost
		maintenanceCost := int64(len(fleet.Escorts)) * m.config.EscortMaintenanceCost

		// Charge player
		player, err := m.playerRepo.GetByID(ctx, playerID)
		if err != nil {
			continue
		}

		if player.Credits >= maintenanceCost {
			// Can afford maintenance
			player.Credits -= maintenanceCost
			_ = m.playerRepo.Update(ctx, player)
		} else {
			// Cannot afford - decay loyalty faster
			for _, escort := range fleet.Escorts {
				escort.Loyalty -= m.config.EscortLoyaltyDecay * 2.0 // Double decay
			}
		}

		// Normal loyalty decay
		for i := len(fleet.Escorts) - 1; i >= 0; i-- {
			escort := fleet.Escorts[i]
			escort.Loyalty -= m.config.EscortLoyaltyDecay

			// Check for desertion
			if escort.Loyalty < m.config.MinLoyaltyThreshold {
				// Escort deserts
				fleet.Escorts = append(fleet.Escorts[:i], fleet.Escorts[i+1:]...)
				log.Info("Escort deserted: player=%s, pilot=%s (low loyalty)", playerID, escort.Pilot)

				if m.onEscortDestroyed != nil {
					go m.onEscortDestroyed(playerID, escort)
				}
			}
		}

		fleet.LastMaintenance = time.Now()
	}
}

// PayMaintenanceEarly allows player to pay maintenance early and boost loyalty
func (m *Manager) PayMaintenanceEarly(ctx context.Context, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return fmt.Errorf("fleet not found")
	}

	if len(fleet.Escorts) == 0 {
		return fmt.Errorf("no escorts to maintain")
	}

	// Calculate cost
	maintenanceCost := int64(len(fleet.Escorts)) * m.config.EscortMaintenanceCost

	// Check player credits
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %v", err)
	}

	if player.Credits < maintenanceCost {
		return fmt.Errorf("insufficient credits (need %d)", maintenanceCost)
	}

	// Deduct cost
	player.Credits -= maintenanceCost
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Boost loyalty for all escorts
	for _, escort := range fleet.Escorts {
		escort.Loyalty += 0.10 // +10% loyalty
		if escort.Loyalty > 1.0 {
			escort.Loyalty = 1.0
		}
	}

	log.Info("Early maintenance paid: player=%s, cost=%d", playerID, maintenanceCost)
	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// GetFleet retrieves a player's fleet
func (m *Manager) GetFleet(playerID uuid.UUID) (*Fleet, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fleet, exists := m.fleets[playerID]
	return fleet, exists
}

// GetFlagship gets the player's active flagship
func (m *Manager) GetFlagship(playerID uuid.UUID) (*models.Ship, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fleet, exists := m.fleets[playerID]
	if !exists {
		return nil, fmt.Errorf("fleet not found")
	}

	for _, ship := range fleet.OwnedShips {
		if ship.ID == fleet.FlagshipID {
			return ship, nil
		}
	}

	return nil, fmt.Errorf("flagship not found")
}

// GetStats returns fleet statistics
func (m *Manager) GetStats() FleetStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := FleetStats{}

	for _, fleet := range m.fleets {
		stats.TotalFleets++
		stats.TotalShips += len(fleet.OwnedShips)
		stats.TotalStoredShips += len(fleet.StoredShips)
		stats.TotalEscorts += len(fleet.Escorts)
	}

	return stats
}

// FleetStats contains fleet statistics
type FleetStats struct {
	TotalFleets      int `json:"total_fleets"`
	TotalShips       int `json:"total_ships"`
	TotalStoredShips int `json:"total_stored_ships"`
	TotalEscorts     int `json:"total_escorts"`
}
