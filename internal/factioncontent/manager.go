// File: internal/factioncontent/manager.go
// Project: Terminal Velocity
// Description: Shared faction content including missions, events, and cooperative gameplay
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package factioncontent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("FactionContent")

// Manager handles shared faction content and cooperative gameplay
type Manager struct {
	mu sync.RWMutex

	// Faction content
	factionMissions map[uuid.UUID]*FactionMission // mission_id -> mission
	factionEvents   map[uuid.UUID]*FactionEvent   // event_id -> event
	objectives      map[uuid.UUID]*SharedObjective // objective_id -> objective
	contributions   map[string]*MemberContribution // "faction_mission_player" -> contribution
	factionRanks    map[string]*FactionRank        // faction_id -> ranks

	// Configuration
	config FactionContentConfig

	// Repositories
	playerRepo *database.PlayerRepository

	// Callbacks
	onMissionComplete   func(mission *FactionMission)
	onEventComplete     func(event *FactionEvent)
	onObjectiveProgress func(objective *SharedObjective, progress float64)
	onRewardDistributed func(factionID uuid.UUID, rewards map[uuid.UUID]int64)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// FactionContentConfig defines faction content parameters
type FactionContentConfig struct {
	// Mission settings
	MaxActiveMissions      int           // Max active faction missions
	MissionDuration        time.Duration // Default mission duration
	MissionRewardMultiplier float64      // Reward multiplier for faction missions
	ContributionTracking   bool          // Track individual contributions

	// Event settings
	EventMinParticipants   int           // Minimum participants for event
	EventDuration          time.Duration // Default event duration
	EventRewardPool        int64         // Base reward pool
	EventScaling           float64       // Reward scaling per participant

	// Objective settings
	ObjectiveTimeout       time.Duration // Time to complete objectives
	SharedProgressBonus    float64       // Bonus for cooperative progress
	ObjectiveRankThreshold int           // Rank required for objectives

	// Rank settings
	RankProgressionRate    float64       // Points needed per rank
	RankDecayRate          float64       // Daily decay for inactive members
	MaxRank                int           // Maximum faction rank
}

// DefaultFactionContentConfig returns sensible defaults
func DefaultFactionContentConfig() FactionContentConfig {
	return FactionContentConfig{
		MaxActiveMissions:      5,
		MissionDuration:        48 * time.Hour,
		MissionRewardMultiplier: 1.5,
		ContributionTracking:   true,
		EventMinParticipants:   5,
		EventDuration:          7 * 24 * time.Hour,
		EventRewardPool:        1000000,
		EventScaling:           0.10,
		ObjectiveTimeout:       24 * time.Hour,
		SharedProgressBonus:    0.25,
		ObjectiveRankThreshold: 2,
		RankProgressionRate:    1.5,
		RankDecayRate:          0.01,
		MaxRank:                10,
	}
}

// NewManager creates a new faction content manager
func NewManager(playerRepo *database.PlayerRepository) *Manager {
	m := &Manager{
		factionMissions: make(map[uuid.UUID]*FactionMission),
		factionEvents:   make(map[uuid.UUID]*FactionEvent),
		objectives:      make(map[uuid.UUID]*SharedObjective),
		contributions:   make(map[string]*MemberContribution),
		factionRanks:    make(map[string]*FactionRank),
		config:          DefaultFactionContentConfig(),
		playerRepo:      playerRepo,
		stopChan:        make(chan struct{}),
	}

	return m
}

// Start begins background workers
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.maintenanceWorker()
	log.Info("Faction content manager started")
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Faction content manager stopped")
}

// SetCallbacks sets all faction content callbacks
func (m *Manager) SetCallbacks(
	onMissionComplete func(mission *FactionMission),
	onEventComplete func(event *FactionEvent),
	onObjectiveProgress func(objective *SharedObjective, progress float64),
	onRewardDistributed func(factionID uuid.UUID, rewards map[uuid.UUID]int64),
) {
	m.onMissionComplete = onMissionComplete
	m.onEventComplete = onEventComplete
	m.onObjectiveProgress = onObjectiveProgress
	m.onRewardDistributed = onRewardDistributed
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// FactionMission represents a mission for faction members
type FactionMission struct {
	ID           uuid.UUID
	FactionID    uuid.UUID
	Name         string
	Description  string
	Type         MissionType
	Objectives   []*SharedObjective
	Rewards      map[string]int64 // reward_type -> amount
	StartTime    time.Time
	EndTime      time.Time
	Status       string // "active", "completed", "failed", "expired"
	Participants []uuid.UUID
	Progress     float64 // 0.0 - 1.0
	MinRank      int     // Minimum faction rank required
}

// MissionType defines types of faction missions
type MissionType string

const (
	MissionTypeCooperative   MissionType = "cooperative"   // Work together
	MissionTypeCompetitive   MissionType = "competitive"   // Compete with other factions
	MissionTypeDefensive     MissionType = "defensive"     // Defend territory
	MissionTypeExpansion     MissionType = "expansion"     // Expand influence
	MissionTypeResource      MissionType = "resource"      // Gather resources
	MissionTypeElimination   MissionType = "elimination"   // Eliminate threats
)

// FactionEvent represents a large-scale faction event
type FactionEvent struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Type          EventType
	FactionID     uuid.UUID // If faction-specific
	Objectives    []*SharedObjective
	RewardPool    int64
	Leaderboard   map[uuid.UUID]int // player_id -> score
	StartTime     time.Time
	EndTime       time.Time
	Status        string // "upcoming", "active", "completed"
	MinParticipants int
	MaxParticipants int
	Participants  []uuid.UUID
}

// EventType defines types of faction events
type EventType string

const (
	EventTypeBossRaid      EventType = "boss_raid"      // Fight powerful boss
	EventTypeTerritoryWar  EventType = "territory_war"  // Control territory
	EventTypeTradeConvoy   EventType = "trade_convoy"   // Protect convoy
	EventTypeResourceRush  EventType = "resource_rush"  // Gather resources
	EventTypeSiegeDefense  EventType = "siege_defense"  // Defend against siege
	EventTypeExpedition    EventType = "expedition"     // Explore new areas
)

// SharedObjective represents a cooperative objective
type SharedObjective struct {
	ID          uuid.UUID
	Name        string
	Description string
	Type        ObjectiveType
	Target      int64   // Target amount
	Current     int64   // Current progress
	Progress    float64 // Percentage 0.0-1.0
	Contributors map[uuid.UUID]int64 // player_id -> contribution
	Status      string  // "active", "completed", "failed"
	CreatedAt   time.Time
	Deadline    time.Time
}

// ObjectiveType defines types of shared objectives
type ObjectiveType string

const (
	ObjectiveKillEnemies     ObjectiveType = "kill_enemies"
	ObjectiveGatherResources ObjectiveType = "gather_resources"
	ObjectiveControlSystems  ObjectiveType = "control_systems"
	ObjectiveTradeVolume     ObjectiveType = "trade_volume"
	ObjectiveDefendBase      ObjectiveType = "defend_base"
	ObjectiveExploreSpace    ObjectiveType = "explore_space"
)

// MemberContribution tracks individual contributions
type MemberContribution struct {
	FactionID   uuid.UUID
	MissionID   uuid.UUID
	PlayerID    uuid.UUID
	PlayerName  string
	Contribution int64
	Rank        int
	LastUpdated time.Time
}

// FactionRank represents a player's rank within faction
type FactionRank struct {
	FactionID   uuid.UUID
	PlayerID    uuid.UUID
	Rank        int       // 1-10
	Points      int64
	Title       string
	Permissions []string  // Permissions granted by rank
	AchievedAt  time.Time
}

// ============================================================================
// FACTION MISSIONS
// ============================================================================

// CreateFactionMission creates a new faction mission
func (m *Manager) CreateFactionMission(ctx context.Context, factionID uuid.UUID, name, description string, missionType MissionType, duration time.Duration, minRank int) (*FactionMission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check mission limit
	activeCount := 0
	for _, mission := range m.factionMissions {
		if mission.FactionID == factionID && mission.Status == "active" {
			activeCount++
		}
	}
	if activeCount >= m.config.MaxActiveMissions {
		return nil, fmt.Errorf("maximum active faction missions reached (%d)", m.config.MaxActiveMissions)
	}

	mission := &FactionMission{
		ID:           uuid.New(),
		FactionID:    factionID,
		Name:         name,
		Description:  description,
		Type:         missionType,
		Objectives:   []*SharedObjective{},
		Rewards:      make(map[string]int64),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(duration),
		Status:       "active",
		Participants: []uuid.UUID{},
		Progress:     0.0,
		MinRank:      minRank,
	}

	m.factionMissions[mission.ID] = mission

	log.Info("Faction mission created: faction=%s, name=%s, type=%s", factionID, name, missionType)
	return mission, nil
}

// JoinFactionMission adds a player to a faction mission
func (m *Manager) JoinFactionMission(ctx context.Context, missionID, playerID uuid.UUID, playerRank int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mission, exists := m.factionMissions[missionID]
	if !exists {
		return fmt.Errorf("mission not found")
	}

	if mission.Status != "active" {
		return fmt.Errorf("mission is not active")
	}

	if playerRank < mission.MinRank {
		return fmt.Errorf("insufficient rank (need rank %d)", mission.MinRank)
	}

	// Check if already participating
	for _, pid := range mission.Participants {
		if pid == playerID {
			return fmt.Errorf("already participating in this mission")
		}
	}

	mission.Participants = append(mission.Participants, playerID)

	log.Info("Player joined faction mission: mission=%s, player=%s", mission.Name, playerID)
	return nil
}

// UpdateMissionProgress updates mission progress
func (m *Manager) UpdateMissionProgress(ctx context.Context, missionID, playerID uuid.UUID, contribution int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mission, exists := m.factionMissions[missionID]
	if !exists {
		return fmt.Errorf("mission not found")
	}

	if mission.Status != "active" {
		return fmt.Errorf("mission is not active")
	}

	// Track contribution
	if m.config.ContributionTracking {
		key := fmt.Sprintf("%s_%s_%s", mission.FactionID, missionID, playerID)
		contrib := m.contributions[key]
		if contrib == nil {
			contrib = &MemberContribution{
				FactionID: mission.FactionID,
				MissionID: missionID,
				PlayerID:  playerID,
			}
			m.contributions[key] = contrib
		}
		contrib.Contribution += contribution
		contrib.LastUpdated = time.Now()
	}

	// Update mission objectives
	for _, objective := range mission.Objectives {
		if objective.Status == "active" {
			objective.Current += contribution
			objective.Progress = float64(objective.Current) / float64(objective.Target)
			if objective.Progress >= 1.0 {
				objective.Progress = 1.0
				objective.Status = "completed"
			}

			// Track contributor
			if objective.Contributors == nil {
				objective.Contributors = make(map[uuid.UUID]int64)
			}
			objective.Contributors[playerID] += contribution

			if m.onObjectiveProgress != nil {
				go m.onObjectiveProgress(objective, objective.Progress)
			}
		}
	}

	// Calculate overall mission progress
	totalObjectives := len(mission.Objectives)
	completedObjectives := 0
	for _, obj := range mission.Objectives {
		if obj.Status == "completed" {
			completedObjectives++
		}
	}

	mission.Progress = float64(completedObjectives) / float64(totalObjectives)

	// Check if mission complete
	if mission.Progress >= 1.0 {
		mission.Status = "completed"
		m.distributeMissionRewards(ctx, mission)

		if m.onMissionComplete != nil {
			go m.onMissionComplete(mission)
		}
	}

	return nil
}

// distributeMissionRewards distributes rewards to participants
func (m *Manager) distributeMissionRewards(ctx context.Context, mission *FactionMission) {
	rewards := make(map[uuid.UUID]int64)

	// Simple equal distribution for now
	if len(mission.Participants) > 0 && mission.Rewards["credits"] > 0 {
		perPlayer := mission.Rewards["credits"] / int64(len(mission.Participants))
		for _, playerID := range mission.Participants {
			rewards[playerID] = perPlayer

			// Award credits
			player, err := m.playerRepo.GetByID(ctx, playerID)
			if err == nil {
				player.Credits += perPlayer
				_ = m.playerRepo.Update(ctx, player)
			}
		}
	}

	log.Info("Mission rewards distributed: mission=%s, participants=%d", mission.Name, len(mission.Participants))

	if m.onRewardDistributed != nil {
		go m.onRewardDistributed(mission.FactionID, rewards)
	}
}

// GetFactionMissions retrieves active missions for a faction
func (m *Manager) GetFactionMissions(factionID uuid.UUID) []*FactionMission {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var missions []*FactionMission
	for _, mission := range m.factionMissions {
		if mission.FactionID == factionID && mission.Status == "active" {
			missions = append(missions, mission)
		}
	}
	return missions
}

// ============================================================================
// FACTION EVENTS
// ============================================================================

// CreateFactionEvent creates a large-scale faction event
func (m *Manager) CreateFactionEvent(ctx context.Context, name, description string, eventType EventType, factionID uuid.UUID, duration time.Duration, rewardPool int64) (*FactionEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	event := &FactionEvent{
		ID:              uuid.New(),
		Name:            name,
		Description:     description,
		Type:            eventType,
		FactionID:       factionID,
		Objectives:      []*SharedObjective{},
		RewardPool:      rewardPool,
		Leaderboard:     make(map[uuid.UUID]int),
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(duration),
		Status:          "active",
		MinParticipants: m.config.EventMinParticipants,
		MaxParticipants: 100, // Default max
		Participants:    []uuid.UUID{},
	}

	m.factionEvents[event.ID] = event

	log.Info("Faction event created: name=%s, type=%s, reward_pool=%d", name, eventType, rewardPool)
	return event, nil
}

// JoinFactionEvent adds a player to a faction event
func (m *Manager) JoinFactionEvent(ctx context.Context, eventID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, exists := m.factionEvents[eventID]
	if !exists {
		return fmt.Errorf("event not found")
	}

	if event.Status != "active" {
		return fmt.Errorf("event is not active")
	}

	if len(event.Participants) >= event.MaxParticipants {
		return fmt.Errorf("event full")
	}

	// Check if already participating
	for _, pid := range event.Participants {
		if pid == playerID {
			return fmt.Errorf("already participating in this event")
		}
	}

	event.Participants = append(event.Participants, playerID)
	event.Leaderboard[playerID] = 0

	log.Info("Player joined faction event: event=%s, player=%s", event.Name, playerID)
	return nil
}

// UpdateEventScore updates a player's event score
func (m *Manager) UpdateEventScore(ctx context.Context, eventID, playerID uuid.UUID, points int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, exists := m.factionEvents[eventID]
	if !exists {
		return fmt.Errorf("event not found")
	}

	if event.Status != "active" {
		return fmt.Errorf("event is not active")
	}

	event.Leaderboard[playerID] += points

	log.Debug("Event score updated: event=%s, player=%s, points=%d", event.Name, playerID, points)
	return nil
}

// CompleteEvent completes an event and distributes rewards
func (m *Manager) CompleteEvent(ctx context.Context, eventID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, exists := m.factionEvents[eventID]
	if !exists {
		return fmt.Errorf("event not found")
	}

	if event.Status != "active" {
		return fmt.Errorf("event is not active")
	}

	event.Status = "completed"

	// Distribute rewards based on leaderboard
	m.distributeEventRewards(ctx, event)

	log.Info("Faction event completed: name=%s, participants=%d", event.Name, len(event.Participants))

	if m.onEventComplete != nil {
		go m.onEventComplete(event)
	}

	return nil
}

// distributeEventRewards distributes event rewards
func (m *Manager) distributeEventRewards(ctx context.Context, event *FactionEvent) {
	if len(event.Leaderboard) == 0 {
		return
	}

	// Sort leaderboard
	type leaderboardEntry struct {
		playerID uuid.UUID
		score    int
	}

	entries := make([]leaderboardEntry, 0, len(event.Leaderboard))
	for playerID, score := range event.Leaderboard {
		entries = append(entries, leaderboardEntry{playerID, score})
	}

	// Simple bubble sort
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].score > entries[i].score {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Award top 3
	rewards := make(map[uuid.UUID]int64)
	rewardMultipliers := []float64{0.50, 0.30, 0.20} // 50%, 30%, 20%

	for i := 0; i < 3 && i < len(entries); i++ {
		reward := int64(float64(event.RewardPool) * rewardMultipliers[i])
		rewards[entries[i].playerID] = reward

		player, err := m.playerRepo.GetByID(ctx, entries[i].playerID)
		if err == nil {
			player.Credits += reward
			_ = m.playerRepo.Update(ctx, player)
		}
	}

	if m.onRewardDistributed != nil {
		go m.onRewardDistributed(event.FactionID, rewards)
	}
}

// GetActiveEvents retrieves all active faction events
func (m *Manager) GetActiveEvents() []*FactionEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var events []*FactionEvent
	for _, event := range m.factionEvents {
		if event.Status == "active" {
			events = append(events, event)
		}
	}
	return events
}

// ============================================================================
// SHARED OBJECTIVES
// ============================================================================

// CreateSharedObjective creates a cooperative objective
func (m *Manager) CreateSharedObjective(ctx context.Context, name, description string, objectiveType ObjectiveType, target int64, deadline time.Duration) (*SharedObjective, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	objective := &SharedObjective{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Type:         objectiveType,
		Target:       target,
		Current:      0,
		Progress:     0.0,
		Contributors: make(map[uuid.UUID]int64),
		Status:       "active",
		CreatedAt:    time.Now(),
		Deadline:     time.Now().Add(deadline),
	}

	m.objectives[objective.ID] = objective

	log.Info("Shared objective created: name=%s, type=%s, target=%d", name, objectiveType, target)
	return objective, nil
}

// UpdateObjectiveProgress updates objective progress
func (m *Manager) UpdateObjectiveProgress(ctx context.Context, objectiveID, playerID uuid.UUID, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	objective, exists := m.objectives[objectiveID]
	if !exists {
		return fmt.Errorf("objective not found")
	}

	if objective.Status != "active" {
		return fmt.Errorf("objective is not active")
	}

	objective.Current += amount
	objective.Progress = float64(objective.Current) / float64(objective.Target)
	if objective.Progress >= 1.0 {
		objective.Progress = 1.0
		objective.Status = "completed"
	}

	objective.Contributors[playerID] += amount

	if m.onObjectiveProgress != nil {
		go m.onObjectiveProgress(objective, objective.Progress)
	}

	return nil
}

// ============================================================================
// FACTION RANKS
// ============================================================================

// UpdateFactionRank updates a player's faction rank
func (m *Manager) UpdateFactionRank(ctx context.Context, factionID, playerID uuid.UUID, points int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s", factionID, playerID)
	rank := m.factionRanks[key]

	if rank == nil {
		rank = &FactionRank{
			FactionID:   factionID,
			PlayerID:    playerID,
			Rank:        1,
			Points:      0,
			Title:       "Recruit",
			Permissions: []string{},
			AchievedAt:  time.Now(),
		}
		m.factionRanks[key] = rank
	}

	rank.Points += points

	// Check for rank up
	requiredPoints := int64(float64(rank.Rank) * m.config.RankProgressionRate * 1000)
	if rank.Points >= requiredPoints && rank.Rank < m.config.MaxRank {
		rank.Rank++
		rank.Title = m.getRankTitle(rank.Rank)
		rank.AchievedAt = time.Now()
		log.Info("Player ranked up: faction=%s, player=%s, rank=%d", factionID, playerID, rank.Rank)
	}

	return nil
}

// getRankTitle returns title for rank
func (m *Manager) getRankTitle(rank int) string {
	titles := []string{
		"Recruit", "Member", "Veteran", "Elite", "Officer",
		"Commander", "General", "Marshal", "Champion", "Legend",
	}
	if rank > 0 && rank <= len(titles) {
		return titles[rank-1]
	}
	return "Unknown"
}

// GetPlayerRank retrieves a player's faction rank
func (m *Manager) GetPlayerRank(factionID, playerID uuid.UUID) *FactionRank {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", factionID, playerID)
	return m.factionRanks[key]
}

// ============================================================================
// BACKGROUND WORKERS
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

// processMaintenance processes daily faction content maintenance
func (m *Manager) processMaintenance() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := context.Background()
	now := time.Now()

	// Expire missions
	for _, mission := range m.factionMissions {
		if mission.Status == "active" && now.After(mission.EndTime) {
			mission.Status = "expired"
			log.Info("Mission expired: mission=%s", mission.Name)
		}
	}

	// Complete events
	for _, event := range m.factionEvents {
		if event.Status == "active" && now.After(event.EndTime) {
			event.Status = "completed"
			m.distributeEventRewards(ctx, event)
			log.Info("Event auto-completed: event=%s", event.Name)
		}
	}

	// Expire objectives
	for _, objective := range m.objectives {
		if objective.Status == "active" && now.After(objective.Deadline) {
			objective.Status = "failed"
			log.Info("Objective failed: objective=%s", objective.Name)
		}
	}

	// Rank decay for inactive members
	for _, rank := range m.factionRanks {
		decay := int64(float64(rank.Points) * m.config.RankDecayRate)
		rank.Points -= decay
		if rank.Points < 0 {
			rank.Points = 0
		}
	}
}

// GetStats returns faction content statistics
func (m *Manager) GetStats() FactionContentStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := FactionContentStats{
		ActiveMissions:   0,
		ActiveEvents:     0,
		ActiveObjectives: 0,
	}

	for _, mission := range m.factionMissions {
		if mission.Status == "active" {
			stats.ActiveMissions++
		}
	}

	for _, event := range m.factionEvents {
		if event.Status == "active" {
			stats.ActiveEvents++
		}
	}

	for _, objective := range m.objectives {
		if objective.Status == "active" {
			stats.ActiveObjectives++
		}
	}

	return stats
}

// FactionContentStats contains faction content statistics
type FactionContentStats struct {
	ActiveMissions   int `json:"active_missions"`
	ActiveEvents     int `json:"active_events"`
	ActiveObjectives int `json:"active_objectives"`
}
