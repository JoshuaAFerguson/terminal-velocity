// File: internal/diplomacy/manager.go
// Project: Terminal Velocity
// Description: Alliance and diplomacy system for faction relations
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15

package diplomacy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/factions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Diplomacy")

// Manager handles alliance and diplomacy between factions
type Manager struct {
	mu sync.RWMutex

	// Diplomatic relations
	alliances   map[uuid.UUID]*Alliance // alliance_id -> alliance
	wars        map[uuid.UUID]*War      // war_id -> war
	relations   map[string]*Relation    // "faction1_faction2" -> relation
	treaties    map[uuid.UUID]*Treaty   // treaty_id -> treaty

	// Configuration
	config DiplomacyConfig

	// Repositories
	playerRepo *database.PlayerRepository

	// Managers
	factionManager *factions.Manager

	// Callbacks
	onAllianceFormed   func(alliance *Alliance)
	onAllianceBroken   func(alliance *Alliance)
	onWarDeclared      func(war *War)
	onWarEnded         func(war *War)
	onTreatySigned     func(treaty *Treaty)
	onTreatyViolated   func(treaty *Treaty)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// DiplomacyConfig defines diplomacy system parameters
type DiplomacyConfig struct {
	// Alliance settings
	MinFactionsForAlliance int           // Minimum factions to form alliance
	MaxFactionsInAlliance  int           // Maximum factions in one alliance
	AllianceFormationCost  int64         // Cost to form an alliance
	AllianceMaintenanceCost int64        // Daily maintenance per faction

	// War settings
	WarDeclarationCost     int64         // Cost to declare war
	MinWarDuration         time.Duration // Minimum war duration
	TruceProposalCooldown  time.Duration // Cooldown between truce proposals
	WarExhaustion          float64       // Daily war exhaustion increase

	// Treaty settings
	TreatyDuration         time.Duration // Default treaty duration
	TreatyViolationPenalty float64       // Penalty for breaking treaty
	MaxActiveTreaties      int           // Max treaties per faction

	// Relation settings
	RelationDecayRate      float64       // Daily decay toward neutral
	RelationChangeLimit    float64       // Max relation change per action
	HostileThreshold       float64       // Below this = hostile
	FriendlyThreshold      float64       // Above this = friendly
}

// DefaultDiplomacyConfig returns sensible defaults
func DefaultDiplomacyConfig() DiplomacyConfig {
	return DiplomacyConfig{
		MinFactionsForAlliance:  2,
		MaxFactionsInAlliance:   5,
		AllianceFormationCost:   100000,
		AllianceMaintenanceCost: 10000,
		WarDeclarationCost:      50000,
		MinWarDuration:          7 * 24 * time.Hour, // 7 days
		TruceProposalCooldown:   24 * time.Hour,
		WarExhaustion:           0.05, // 5% per day
		TreatyDuration:          30 * 24 * time.Hour, // 30 days
		TreatyViolationPenalty:  0.50, // -50% relations
		MaxActiveTreaties:       10,
		RelationDecayRate:       0.01, // 1% per day toward neutral
		RelationChangeLimit:     0.10, // Max 10% change per action
		HostileThreshold:        -0.30, // Below -30% = hostile
		FriendlyThreshold:       0.30,  // Above 30% = friendly
	}
}

// NewManager creates a new diplomacy manager
func NewManager(playerRepo *database.PlayerRepository, factionManager *factions.Manager) *Manager {
	return &Manager{
		alliances:      make(map[uuid.UUID]*Alliance),
		wars:           make(map[uuid.UUID]*War),
		relations:      make(map[string]*Relation),
		treaties:       make(map[uuid.UUID]*Treaty),
		config:         DefaultDiplomacyConfig(),
		playerRepo:     playerRepo,
		factionManager: factionManager,
		stopChan:       make(chan struct{}),
	}
}

// Start begins background workers for diplomacy
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.maintenanceWorker()
	log.Info("Diplomacy manager started")
}

// Stop gracefully shuts down the diplomacy manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Diplomacy manager stopped")
}

// SetCallbacks sets all diplomacy callbacks
func (m *Manager) SetCallbacks(
	onAllianceFormed func(alliance *Alliance),
	onAllianceBroken func(alliance *Alliance),
	onWarDeclared func(war *War),
	onWarEnded func(war *War),
	onTreatySigned func(treaty *Treaty),
	onTreatyViolated func(treaty *Treaty),
) {
	m.onAllianceFormed = onAllianceFormed
	m.onAllianceBroken = onAllianceBroken
	m.onWarDeclared = onWarDeclared
	m.onWarEnded = onWarEnded
	m.onTreatySigned = onTreatySigned
	m.onTreatyViolated = onTreatyViolated
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// Alliance represents a multi-faction alliance
type Alliance struct {
	ID          uuid.UUID
	Name        string
	Description string
	LeaderFactionID uuid.UUID
	MemberFactions  []uuid.UUID
	CreatedAt   time.Time
	Status      string // "active", "dissolved"
	Treasury    int64  // Shared alliance treasury
	LastMaintenance time.Time
}

// War represents a conflict between factions/alliances
type War struct {
	ID           uuid.UUID
	Name         string
	AggressorFactions []uuid.UUID
	DefenderFactions  []uuid.UUID
	StartTime    time.Time
	EndTime      time.Time
	Status       string // "active", "truce_proposed", "ended"
	WarScore     map[uuid.UUID]int // faction_id -> war score
	Exhaustion   float64 // 0.0-1.0, increases over time
	TruceProposedBy uuid.UUID
	TruceProposedAt time.Time
}

// Treaty represents a diplomatic agreement
type Treaty struct {
	ID          uuid.UUID
	Type        TreatyType
	Faction1    uuid.UUID
	Faction2    uuid.UUID
	Terms       string // Description of treaty terms
	SignedAt    time.Time
	ExpiresAt   time.Time
	Status      string // "active", "violated", "expired"
	ViolatedBy  uuid.UUID
}

// TreatyType defines types of treaties
type TreatyType string

const (
	TreatyNonAggression  TreatyType = "non_aggression"  // Cannot attack each other
	TreatyTradeAgreement TreatyType = "trade_agreement" // Trade bonuses
	TreatyMutualDefense  TreatyType = "mutual_defense"  // Defend if attacked
	TreatyResearchPact   TreatyType = "research_pact"   // Share technology
)

// Relation tracks relationship between two factions
type Relation struct {
	Faction1    uuid.UUID
	Faction2    uuid.UUID
	Value       float64 // -1.0 (hostile) to +1.0 (friendly)
	Status      string  // "hostile", "neutral", "friendly", "allied"
	UpdatedAt   time.Time
}

// ============================================================================
// ALLIANCE SYSTEM
// ============================================================================

// FormAlliance creates a new alliance
func (m *Manager) FormAlliance(ctx context.Context, name, description string, leaderFactionID uuid.UUID, memberFactions []uuid.UUID) (*Alliance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate faction count
	totalFactions := len(memberFactions) + 1 // +1 for leader
	if totalFactions < m.config.MinFactionsForAlliance {
		return nil, fmt.Errorf("need at least %d factions to form alliance", m.config.MinFactionsForAlliance)
	}
	if totalFactions > m.config.MaxFactionsInAlliance {
		return nil, fmt.Errorf("maximum %d factions in alliance", m.config.MaxFactionsInAlliance)
	}

	// Check if any faction is already in an alliance
	allFactions := append([]uuid.UUID{leaderFactionID}, memberFactions...)
	for _, factionID := range allFactions {
		if m.isInAlliance(factionID) {
			return nil, fmt.Errorf("faction %s is already in an alliance", factionID)
		}
	}

	// Create alliance
	alliance := &Alliance{
		ID:              uuid.New(),
		Name:            name,
		Description:     description,
		LeaderFactionID: leaderFactionID,
		MemberFactions:  memberFactions,
		CreatedAt:       time.Now(),
		Status:          "active",
		Treasury:        0,
		LastMaintenance: time.Now(),
	}

	m.alliances[alliance.ID] = alliance

	// Update relations between member factions
	for i, faction1 := range allFactions {
		for j := i + 1; j < len(allFactions); j++ {
			faction2 := allFactions[j]
			m.modifyRelationUnsafe(faction1, faction2, 0.50) // +50% relation
		}
	}

	log.Info("Alliance formed: name=%s, leader=%s, members=%d", name, leaderFactionID, len(memberFactions))

	if m.onAllianceFormed != nil {
		go m.onAllianceFormed(alliance)
	}

	return alliance, nil
}

// DisbandAlliance dissolves an alliance
func (m *Manager) DisbandAlliance(ctx context.Context, allianceID, factionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alliance, exists := m.alliances[allianceID]
	if !exists {
		return fmt.Errorf("alliance not found")
	}

	// Only leader can disband
	if alliance.LeaderFactionID != factionID {
		return fmt.Errorf("only alliance leader can disband")
	}

	if alliance.Status != "active" {
		return fmt.Errorf("alliance is not active")
	}

	// Distribute treasury evenly among all member factions
	if alliance.Treasury > 0 && m.factionManager != nil {
		allFactions := append([]uuid.UUID{alliance.LeaderFactionID}, alliance.MemberFactions...)
		sharePerFaction := alliance.Treasury / int64(len(allFactions))

		// Distribute to each faction's treasury
		for _, factionID := range allFactions {
			// Use a dummy player ID (first member) to satisfy Deposit signature
			// In practice, this is an alliance action, not a player action
			faction, err := m.factionManager.GetFaction(factionID)
			if err != nil || faction == nil {
				log.Warn("Failed to get faction %s for treasury distribution: %v", factionID, err)
				continue
			}

			// Deposit to faction treasury (use leader as depositor for logging)
			if len(faction.Members) > 0 {
				depositErr := m.factionManager.Deposit(factionID, faction.LeaderID, sharePerFaction)
				if depositErr != nil {
					log.Error("Failed to distribute %d credits to faction %s: %v", sharePerFaction, factionID, depositErr)
				} else {
					log.Info("Distributed %d credits to faction %s from disbanded alliance", sharePerFaction, factionID)
				}
			}
		}

		// Clear alliance treasury
		alliance.Treasury = 0
	}

	alliance.Status = "dissolved"

	log.Info("Alliance disbanded: name=%s", alliance.Name)

	if m.onAllianceBroken != nil {
		go m.onAllianceBroken(alliance)
	}

	return nil
}

// LeaveFaction allows a faction to leave an alliance
func (m *Manager) LeaveAlliance(ctx context.Context, allianceID, factionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alliance, exists := m.alliances[allianceID]
	if !exists {
		return fmt.Errorf("alliance not found")
	}

	// Leader cannot leave (must disband instead)
	if alliance.LeaderFactionID == factionID {
		return fmt.Errorf("leader cannot leave alliance (must disband)")
	}

	// Remove from members
	for i, memberID := range alliance.MemberFactions {
		if memberID == factionID {
			alliance.MemberFactions = append(alliance.MemberFactions[:i], alliance.MemberFactions[i+1:]...)
			log.Info("Faction left alliance: faction=%s, alliance=%s", factionID, alliance.Name)
			return nil
		}
	}

	return fmt.Errorf("faction not in alliance")
}

// ============================================================================
// WAR SYSTEM
// ============================================================================

// DeclareWar starts a war between factions/alliances
func (m *Manager) DeclareWar(ctx context.Context, name string, aggressors, defenders []uuid.UUID) (*War, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	if len(aggressors) == 0 || len(defenders) == 0 {
		return nil, fmt.Errorf("need at least one aggressor and defender")
	}

	// Check for existing wars between these factions
	for _, war := range m.wars {
		if war.Status == "active" {
			// Check if any factions overlap
			for _, aggressor := range aggressors {
				for _, existingAggressor := range war.AggressorFactions {
					if aggressor == existingAggressor {
						return nil, fmt.Errorf("faction already at war")
					}
				}
			}
		}
	}

	// Create war
	war := &War{
		ID:                uuid.New(),
		Name:              name,
		AggressorFactions: aggressors,
		DefenderFactions:  defenders,
		StartTime:         time.Now(),
		Status:            "active",
		WarScore:          make(map[uuid.UUID]int),
		Exhaustion:        0.0,
	}

	// Initialize war scores
	for _, factionID := range aggressors {
		war.WarScore[factionID] = 0
	}
	for _, factionID := range defenders {
		war.WarScore[factionID] = 0
	}

	m.wars[war.ID] = war

	// Update relations
	for _, aggressor := range aggressors {
		for _, defender := range defenders {
			m.modifyRelationUnsafe(aggressor, defender, -0.75) // -75% relation
		}
	}

	log.Info("War declared: name=%s, aggressors=%d, defenders=%d", name, len(aggressors), len(defenders))

	if m.onWarDeclared != nil {
		go m.onWarDeclared(war)
	}

	return war, nil
}

// ProposeTruce proposes ending a war
func (m *Manager) ProposeTruce(ctx context.Context, warID, factionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, exists := m.wars[warID]
	if !exists {
		return fmt.Errorf("war not found")
	}

	if war.Status != "active" {
		return fmt.Errorf("war is not active")
	}

	// Check minimum war duration
	if time.Since(war.StartTime) < m.config.MinWarDuration {
		return fmt.Errorf("war must last at least %v before truce", m.config.MinWarDuration)
	}

	// Check cooldown
	if war.TruceProposedAt.Add(m.config.TruceProposalCooldown).After(time.Now()) {
		return fmt.Errorf("must wait %v between truce proposals", m.config.TruceProposalCooldown)
	}

	war.Status = "truce_proposed"
	war.TruceProposedBy = factionID
	war.TruceProposedAt = time.Now()

	log.Info("Truce proposed: war=%s, proposer=%s", war.Name, factionID)
	return nil
}

// AcceptTruce accepts a truce proposal
func (m *Manager) AcceptTruce(ctx context.Context, warID, factionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, exists := m.wars[warID]
	if !exists {
		return fmt.Errorf("war not found")
	}

	if war.Status != "truce_proposed" {
		return fmt.Errorf("no truce proposed")
	}

	// End war
	war.Status = "ended"
	war.EndTime = time.Now()

	// Slight relation improvement
	for _, aggressor := range war.AggressorFactions {
		for _, defender := range war.DefenderFactions {
			m.modifyRelationUnsafe(aggressor, defender, 0.20) // +20% relation
		}
	}

	log.Info("Truce accepted: war=%s, acceptor=%s", war.Name, factionID)

	if m.onWarEnded != nil {
		go m.onWarEnded(war)
	}

	return nil
}

// UpdateWarScore updates a faction's war score
func (m *Manager) UpdateWarScore(warID, factionID uuid.UUID, change int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, exists := m.wars[warID]
	if !exists {
		return fmt.Errorf("war not found")
	}

	if war.Status != "active" {
		return fmt.Errorf("war is not active")
	}

	war.WarScore[factionID] += change
	log.Debug("War score updated: war=%s, faction=%s, change=%d", war.Name, factionID, change)
	return nil
}

// ============================================================================
// TREATY SYSTEM
// ============================================================================

// SignTreaty creates a new treaty between factions
func (m *Manager) SignTreaty(ctx context.Context, treatyType TreatyType, faction1, faction2 uuid.UUID, terms string, duration time.Duration) (*Treaty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check treaty limits
	count1 := m.countActiveTreaties(faction1)
	count2 := m.countActiveTreaties(faction2)

	if count1 >= m.config.MaxActiveTreaties || count2 >= m.config.MaxActiveTreaties {
		return nil, fmt.Errorf("maximum active treaties reached")
	}

	// Check if already at war
	if m.areAtWar(faction1, faction2) {
		return nil, fmt.Errorf("factions are at war")
	}

	// Create treaty
	treaty := &Treaty{
		ID:        uuid.New(),
		Type:      treatyType,
		Faction1:  faction1,
		Faction2:  faction2,
		Terms:     terms,
		SignedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
		Status:    "active",
	}

	m.treaties[treaty.ID] = treaty

	// Improve relations
	m.modifyRelationUnsafe(faction1, faction2, 0.25) // +25% relation

	log.Info("Treaty signed: type=%s, factions=%s,%s", treatyType, faction1, faction2)

	if m.onTreatySigned != nil {
		go m.onTreatySigned(treaty)
	}

	return treaty, nil
}

// ViolateTreaty marks a treaty as violated
func (m *Manager) ViolateTreaty(ctx context.Context, treatyID, violatorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	treaty, exists := m.treaties[treatyID]
	if !exists {
		return fmt.Errorf("treaty not found")
	}

	if treaty.Status != "active" {
		return fmt.Errorf("treaty is not active")
	}

	treaty.Status = "violated"
	treaty.ViolatedBy = violatorID

	// Severe relation penalty
	otherFaction := treaty.Faction1
	if violatorID == treaty.Faction1 {
		otherFaction = treaty.Faction2
	}

	m.modifyRelationUnsafe(violatorID, otherFaction, -m.config.TreatyViolationPenalty)

	log.Info("Treaty violated: treaty=%s, violator=%s", treaty.ID, violatorID)

	if m.onTreatyViolated != nil {
		go m.onTreatyViolated(treaty)
	}

	return nil
}

// ============================================================================
// RELATION SYSTEM
// ============================================================================

// ModifyRelation changes relationship between two factions
func (m *Manager) ModifyRelation(faction1, faction2 uuid.UUID, change float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.modifyRelationUnsafe(faction1, faction2, change)
}

// modifyRelationUnsafe changes relationship (must hold lock)
func (m *Manager) modifyRelationUnsafe(faction1, faction2 uuid.UUID, change float64) {
	// Limit change
	if change > m.config.RelationChangeLimit {
		change = m.config.RelationChangeLimit
	}
	if change < -m.config.RelationChangeLimit {
		change = -m.config.RelationChangeLimit
	}

	key := m.getRelationKey(faction1, faction2)
	relation, exists := m.relations[key]

	if !exists {
		relation = &Relation{
			Faction1:  faction1,
			Faction2:  faction2,
			Value:     0.0,
			Status:    "neutral",
			UpdatedAt: time.Now(),
		}
		m.relations[key] = relation
	}

	relation.Value += change

	// Clamp to -1.0 to +1.0
	if relation.Value > 1.0 {
		relation.Value = 1.0
	}
	if relation.Value < -1.0 {
		relation.Value = -1.0
	}

	// Update status
	if relation.Value <= m.config.HostileThreshold {
		relation.Status = "hostile"
	} else if relation.Value >= m.config.FriendlyThreshold {
		relation.Status = "friendly"
	} else {
		relation.Status = "neutral"
	}

	relation.UpdatedAt = time.Now()
}

// GetRelation retrieves relationship between factions
func (m *Manager) GetRelation(faction1, faction2 uuid.UUID) *Relation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.getRelationKey(faction1, faction2)
	relation, exists := m.relations[key]

	if !exists {
		return &Relation{
			Faction1:  faction1,
			Faction2:  faction2,
			Value:     0.0,
			Status:    "neutral",
			UpdatedAt: time.Now(),
		}
	}

	return relation
}

// getRelationKey creates a consistent key for two factions
func (m *Manager) getRelationKey(faction1, faction2 uuid.UUID) string {
	// Ensure consistent ordering
	if faction1.String() < faction2.String() {
		return fmt.Sprintf("%s_%s", faction1, faction2)
	}
	return fmt.Sprintf("%s_%s", faction2, faction1)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// isInAlliance checks if a faction is in any alliance
func (m *Manager) isInAlliance(factionID uuid.UUID) bool {
	for _, alliance := range m.alliances {
		if alliance.Status != "active" {
			continue
		}
		if alliance.LeaderFactionID == factionID {
			return true
		}
		for _, memberID := range alliance.MemberFactions {
			if memberID == factionID {
				return true
			}
		}
	}
	return false
}

// areAtWar checks if two factions are at war
func (m *Manager) areAtWar(faction1, faction2 uuid.UUID) bool {
	for _, war := range m.wars {
		if war.Status != "active" {
			continue
		}
		for _, aggressor := range war.AggressorFactions {
			for _, defender := range war.DefenderFactions {
				if (aggressor == faction1 && defender == faction2) ||
					(aggressor == faction2 && defender == faction1) {
					return true
				}
			}
		}
	}
	return false
}

// countActiveTreaties counts active treaties for a faction
func (m *Manager) countActiveTreaties(factionID uuid.UUID) int {
	count := 0
	for _, treaty := range m.treaties {
		if treaty.Status == "active" && time.Now().Before(treaty.ExpiresAt) {
			if treaty.Faction1 == factionID || treaty.Faction2 == factionID {
				count++
			}
		}
	}
	return count
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

// processMaintenance processes daily diplomacy maintenance
func (m *Manager) processMaintenance() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Relation decay toward neutral
	for _, relation := range m.relations {
		if relation.Value > 0 {
			relation.Value -= m.config.RelationDecayRate
			if relation.Value < 0 {
				relation.Value = 0
			}
		} else if relation.Value < 0 {
			relation.Value += m.config.RelationDecayRate
			if relation.Value > 0 {
				relation.Value = 0
			}
		}
	}

	// War exhaustion
	for _, war := range m.wars {
		if war.Status == "active" {
			war.Exhaustion += m.config.WarExhaustion
			if war.Exhaustion >= 1.0 {
				war.Exhaustion = 1.0
				// Auto-propose truce at 100% exhaustion
				if war.TruceProposedBy == uuid.Nil {
					war.Status = "truce_proposed"
					war.TruceProposedAt = now
					log.Info("War exhausted, auto-proposing truce: %s", war.Name)
				}
			}
		}
	}

	// Expire treaties
	for _, treaty := range m.treaties {
		if treaty.Status == "active" && now.After(treaty.ExpiresAt) {
			treaty.Status = "expired"
			log.Info("Treaty expired: %s", treaty.ID)
		}
	}
}

// GetStats returns diplomacy statistics
func (m *Manager) GetStats() DiplomacyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := DiplomacyStats{}

	for _, alliance := range m.alliances {
		if alliance.Status == "active" {
			stats.ActiveAlliances++
		}
	}

	for _, war := range m.wars {
		if war.Status == "active" {
			stats.ActiveWars++
		}
	}

	for _, treaty := range m.treaties {
		if treaty.Status == "active" && time.Now().Before(treaty.ExpiresAt) {
			stats.ActiveTreaties++
		}
	}

	return stats
}

// DiplomacyStats contains diplomacy statistics
type DiplomacyStats struct {
	ActiveAlliances int `json:"active_alliances"`
	ActiveWars      int `json:"active_wars"`
	ActiveTreaties  int `json:"active_treaties"`
}
