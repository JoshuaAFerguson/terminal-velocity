// File: internal/quests/manager.go
// Project: Terminal Velocity
// Description: Quest and storyline management system
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package quests provides quest and storyline management for the game.
//
// This package handles:
// - Quest template registration and storage (7 quest types)
// - Player quest progression tracking (12 objective types)
// - Storyline management with branching narratives
// - Quest prerequisite validation
// - Objective progress updates and completion
// - Quest rewards (credits, experience, reputation, items, system unlocks)
//
// Quest Types:
// - Main: Primary storyline quests with significant rewards
// - Side: Optional quests with moderate rewards, often repeatable
// - Faction: Faction-specific quests that affect reputation
// - Daily: Repeatable daily quests for consistent rewards
// - Hidden: Secret quests discovered through exploration or actions
// - Event: Time-limited event quests
// - Tutorial: Guided quests for new players
//
// Objective Types:
// - Collect: Gather specific items or commodities
// - Deliver: Deliver items to a specific location
// - Kill: Defeat enemy ships or NPCs
// - Travel: Visit specific systems or locations
// - Scan: Scan anomalies or objects
// - Mine: Mine resources from asteroids
// - Investigate: Explore locations or examine objects
// - Escort: Protect ships or convoys
// - Trade: Complete trading objectives
// - Craft: Craft specific items
// - Hack: Hack terminals or systems
// - Reputation: Achieve specific reputation levels
//
// Thread Safety:
// All Manager methods are thread-safe using sync.RWMutex. Read operations
// use RLock, write operations use Lock.
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package quests

import (
	"errors"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Quests")

// Manager handles quest progression and storylines for all players.
// It maintains quest templates, storylines, and per-player quest state.
// All operations are thread-safe.
type Manager struct {
	mu           sync.RWMutex                        // Protects all fields
	quests       map[string]*models.Quest            // All quest templates indexed by quest ID
	storylines   map[string]*models.Storyline        // All storylines indexed by storyline ID
	playerQuests map[uuid.UUID][]*models.PlayerQuest // Player quest instances indexed by player ID
}

// NewManager creates a new quest manager with default quest content.
//
// The manager is initialized with:
// - Main storyline "The Void Threat" with 3 progressive quests
// - Side quest "Merchant's Request" (repeatable)
// - Faction quest "Federation Patrol"
// - Hidden quest "The Ancient Artifact"
// - Daily quest "Resource Gathering" (repeatable)
//
// Returns:
//   - Pointer to new Manager with quests and storylines loaded
//
// Thread Safety:
// Safe to call concurrently, though typically called once at server startup.
func NewManager() *Manager {
	m := &Manager{
		quests:       make(map[string]*models.Quest),
		storylines:   make(map[string]*models.Storyline),
		playerQuests: make(map[uuid.UUID][]*models.PlayerQuest),
	}

	// Initialize default quests and storylines
	m.initializeDefaultQuests()

	return m
}

// initializeDefaultQuests creates the default quest content.
//
// This method populates the manager with starter quests including:
// - Main storyline quests (First Steps → Distress Signal → Void Anomaly)
// - Side quests (merchant trading)
// - Faction quests (Federation patrols)
// - Hidden quests (ancient artifacts)
// - Daily quests (resource gathering)
//
// Called automatically during NewManager initialization.
func (m *Manager) initializeDefaultQuests() {
	// Main Storyline: The Void Threat
	mainStory := models.NewStoryline(
		"main_void_threat",
		"The Void Threat",
		"An ancient threat emerges from the depths of space, threatening all known systems.",
		true,
	)

	// Quest 1: First Steps
	quest1 := models.NewQuest(
		"main_01_first_steps",
		"First Steps",
		"Commander, welcome to the fleet. Complete your initial training and prove you're ready for active duty.",
		models.QuestTypeMain,
	)
	quest1.Level = 1
	quest1.Giver = "Admiral Voss"
	quest1.AddObjective(&models.QuestObjective{
		ID:          "obj_buy_cargo",
		Type:        models.ObjectiveCollect,
		Description: "Purchase 10 units of food from the market",
		Target:      "food",
		Required:    10,
		Current:     0,
	})
	quest1.AddObjective(&models.QuestObjective{
		ID:          "obj_deliver_cargo",
		Type:        models.ObjectiveDeliver,
		Description: "Deliver the food to New Haven station",
		Target:      "station_new_haven",
		Required:    1,
		Current:     0,
	})
	quest1.Rewards = models.QuestReward{
		Credits:    5000,
		Experience: 100,
		Reputation: map[string]int{"federation": 10},
	}
	quest1.NextQuests = []string{"main_02_distress_signal"}
	m.RegisterQuest(quest1)
	mainStory.AddQuest(quest1.ID)

	// Quest 2: Distress Signal
	quest2 := models.NewQuest(
		"main_02_distress_signal",
		"Distress Signal",
		"A distress signal from a remote outpost hints at something sinister.",
		models.QuestTypeMain,
	)
	quest2.Level = 3
	quest2.Giver = "Admiral Voss"
	quest2.Prerequisites = []string{"main_01_first_steps"}
	quest2.AddObjective(&models.QuestObjective{
		ID:          "obj_travel_outpost",
		Type:        models.ObjectiveTravel,
		Description: "Travel to Frontier Outpost in the Epsilon system",
		Target:      "system_epsilon",
		Required:    1,
		Current:     0,
	})
	quest2.AddObjective(&models.QuestObjective{
		ID:          "obj_investigate",
		Type:        models.ObjectiveInvestigate,
		Description: "Investigate the abandoned outpost",
		Target:      "outpost_frontier",
		Required:    1,
		Current:     0,
	})
	quest2.AddObjective(&models.QuestObjective{
		ID:          "obj_defeat_pirates",
		Type:        models.ObjectiveKill,
		Description: "Defeat the pirate ambush (optional)",
		Target:      "pirate_raider",
		Required:    3,
		Current:     0,
		Optional:    true,
	})
	quest2.Rewards = models.QuestReward{
		Credits:    10000,
		Experience: 250,
		Items:      map[string]int{"scan_data": 1},
		Reputation: map[string]int{"federation": 20},
	}
	quest2.NextQuests = []string{"main_03_void_anomaly"}
	m.RegisterQuest(quest2)
	mainStory.AddQuest(quest2.ID)

	// Quest 3: The Void Anomaly
	quest3 := models.NewQuest(
		"main_03_void_anomaly",
		"The Void Anomaly",
		"The scan data reveals a mysterious anomaly that defies all known physics.",
		models.QuestTypeMain,
	)
	quest3.Level = 5
	quest3.Giver = "Dr. Elena Kira"
	quest3.Prerequisites = []string{"main_02_distress_signal"}
	quest3.AddObjective(&models.QuestObjective{
		ID:          "obj_deliver_data",
		Type:        models.ObjectiveDeliver,
		Description: "Deliver scan data to Dr. Kira at Research Station Alpha",
		Target:      "station_research_alpha",
		Required:    1,
		Current:     0,
	})
	quest3.AddObjective(&models.QuestObjective{
		ID:          "obj_scan_anomaly",
		Type:        models.ObjectiveScan,
		Description: "Scan the void anomaly (dangerous)",
		Target:      "void_anomaly_01",
		Required:    1,
		Current:     0,
	})
	quest3.AddObjective(&models.QuestObjective{
		ID:          "obj_collect_samples",
		Type:        models.ObjectiveCollect,
		Description: "Collect 5 void energy samples",
		Target:      "void_energy",
		Required:    5,
		Current:     0,
	})
	quest3.Rewards = models.QuestReward{
		Credits:      25000,
		Experience:   500,
		Reputation:   map[string]int{"scientists": 50, "federation": 30},
		SystemUnlock: "system_void_sector",
	}
	m.RegisterQuest(quest3)
	mainStory.AddQuest(quest3.ID)

	m.RegisterStoryline(mainStory)

	// Side Quest: Merchant's Request
	sideQuest1 := models.NewQuest(
		"side_merchant_request",
		"Merchant's Request",
		"A local merchant needs rare goods transported across the sector.",
		models.QuestTypeSide,
	)
	sideQuest1.Level = 2
	sideQuest1.Giver = "Merchant Talis"
	sideQuest1.AddObjective(&models.QuestObjective{
		ID:          "obj_buy_luxuries",
		Type:        models.ObjectiveCollect,
		Description: "Purchase 20 units of luxury goods",
		Target:      "luxury_goods",
		Required:    20,
		Current:     0,
	})
	sideQuest1.AddObjective(&models.QuestObjective{
		ID:          "obj_deliver_luxuries",
		Type:        models.ObjectiveDeliver,
		Description: "Deliver luxury goods to Paradise Station",
		Target:      "station_paradise",
		Required:    1,
		Current:     0,
	})
	sideQuest1.Rewards = models.QuestReward{
		Credits:    15000,
		Experience: 150,
		Reputation: map[string]int{"merchants_guild": 25},
	}
	sideQuest1.Repeatable = true
	m.RegisterQuest(sideQuest1)

	// Faction Quest: Federation Patrol
	factionQuest1 := models.NewQuest(
		"faction_fed_patrol",
		"Federation Patrol",
		"Join a Federation patrol to secure the trade routes.",
		models.QuestTypeFaction,
	)
	factionQuest1.Level = 4
	factionQuest1.Giver = "Commander Hayes"
	factionQuest1.AddObjective(&models.QuestObjective{
		ID:          "obj_patrol_route",
		Type:        models.ObjectiveTravel,
		Description: "Patrol the Alpha trade route (3 systems)",
		Target:      "trade_route_alpha",
		Required:    3,
		Current:     0,
	})
	factionQuest1.AddObjective(&models.QuestObjective{
		ID:          "obj_defeat_hostiles",
		Type:        models.ObjectiveKill,
		Description: "Defeat any hostile ships encountered",
		Target:      "hostile_any",
		Required:    5,
		Current:     0,
	})
	factionQuest1.Rewards = models.QuestReward{
		Credits:    20000,
		Experience: 300,
		Reputation: map[string]int{"federation": 50},
		Items:      map[string]int{"patrol_commendation": 1},
	}
	m.RegisterQuest(factionQuest1)

	// Hidden Quest: Ancient Artifact
	hiddenQuest := models.NewQuest(
		"hidden_ancient_artifact",
		"The Ancient Artifact",
		"You've discovered coordinates to an ancient alien artifact.",
		models.QuestTypeHidden,
	)
	hiddenQuest.Level = 7
	hiddenQuest.Giver = "Unknown"
	hiddenQuest.AddObjective(&models.QuestObjective{
		ID:          "obj_find_coordinates",
		Type:        models.ObjectiveInvestigate,
		Description: "Decode the ancient coordinates",
		Target:      "coordinates_ancient",
		Required:    1,
		Current:     0,
		Hidden:      true,
	})
	hiddenQuest.AddObjective(&models.QuestObjective{
		ID:          "obj_travel_ruins",
		Type:        models.ObjectiveTravel,
		Description: "Travel to the ancient ruins",
		Target:      "system_ancient_ruins",
		Required:    1,
		Current:     0,
	})
	hiddenQuest.AddObjective(&models.QuestObjective{
		ID:          "obj_retrieve_artifact",
		Type:        models.ObjectiveCollect,
		Description: "Retrieve the ancient artifact",
		Target:      "artifact_ancient",
		Required:    1,
		Current:     0,
	})
	hiddenQuest.Rewards = models.QuestReward{
		Credits:    50000,
		Experience: 1000,
		Items:      map[string]int{"artifact_ancient": 1},
		Special:    "Unique ship upgrade unlocked",
	}
	m.RegisterQuest(hiddenQuest)

	// Daily Quest: Resource Gathering
	dailyQuest := models.NewQuest(
		"daily_resource_gathering",
		"Daily Resource Run",
		"Gather resources for the station's daily operations.",
		models.QuestTypeDaily,
	)
	dailyQuest.Level = 1
	dailyQuest.Giver = "Station Manager"
	dailyQuest.Repeatable = true
	dailyQuest.AddObjective(&models.QuestObjective{
		ID:          "obj_mine_ore",
		Type:        models.ObjectiveMine,
		Description: "Mine 50 units of ore",
		Target:      "ore_common",
		Required:    50,
		Current:     0,
	})
	dailyQuest.Rewards = models.QuestReward{
		Credits:    5000,
		Experience: 50,
	}
	m.RegisterQuest(dailyQuest)
}

// RegisterQuest adds a quest template to the manager.
//
// Parameters:
//   - quest: Quest template to register
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) RegisterQuest(quest *models.Quest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.quests[quest.ID] = quest
}

// RegisterStoryline adds a storyline to the manager.
//
// Parameters:
//   - storyline: Storyline to register
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) RegisterStoryline(storyline *models.Storyline) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.storylines[storyline.ID] = storyline
}

// GetQuest returns a quest template by ID.
//
// Parameters:
//   - questID: Quest identifier
//
// Returns:
//   - Quest template, or nil if not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetQuest(questID string) *models.Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.quests[questID]
}

// GetStoryline returns a storyline by ID.
//
// Parameters:
//   - storylineID: Storyline identifier
//
// Returns:
//   - Storyline, or nil if not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetStoryline(storylineID string) *models.Storyline {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.storylines[storylineID]
}

// GetAllQuests returns all available quest templates.
//
// Returns:
//   - Slice of all quest templates
//
// Thread Safety:
// Thread-safe. Acquires read lock. Returns a new slice.
func (m *Manager) GetAllQuests() []*models.Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quests := make([]*models.Quest, 0, len(m.quests))
	for _, quest := range m.quests {
		quests = append(quests, quest)
	}
	return quests
}

// GetPlayerQuests returns all quest instances for a player.
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Slice of all player's quests (active, completed, failed, abandoned)
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetPlayerQuests(playerID uuid.UUID) []*models.PlayerQuest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.playerQuests[playerID]
}

// GetActiveQuests returns all active quests for a player.
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Slice of active player quests
//
// Thread Safety:
// Thread-safe. Acquires read lock. Returns a new slice.
func (m *Manager) GetActiveQuests(playerID uuid.UUID) []*models.PlayerQuest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*models.PlayerQuest, 0)
	for _, pq := range m.playerQuests[playerID] {
		if pq.Status == models.QuestStatusActive {
			active = append(active, pq)
		}
	}
	return active
}

// GetCompletedQuests returns all completed quests for a player.
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Slice of completed player quests
//
// Thread Safety:
// Thread-safe. Acquires read lock. Returns a new slice.
func (m *Manager) GetCompletedQuests(playerID uuid.UUID) []*models.PlayerQuest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	completed := make([]*models.PlayerQuest, 0)
	for _, pq := range m.playerQuests[playerID] {
		if pq.Status == models.QuestStatusCompleted {
			completed = append(completed, pq)
		}
	}
	return completed
}

// CanStartQuest checks if a player can start a quest.
//
// Validates:
// - Quest exists
// - Not already active
// - Not already completed (unless repeatable)
// - All prerequisites completed
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//
// Returns:
//   - true if player can start the quest
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) CanStartQuest(playerID uuid.UUID, questID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quest := m.quests[questID]
	if quest == nil {
		return false
	}

	// Check if already active or completed
	for _, pq := range m.playerQuests[playerID] {
		if pq.QuestID == questID {
			if pq.Status == models.QuestStatusActive {
				return false
			}
			if pq.Status == models.QuestStatusCompleted && !quest.Repeatable {
				return false
			}
		}
	}

	// Check prerequisites
	for _, prereqID := range quest.Prerequisites {
		completed := false
		for _, pq := range m.playerQuests[playerID] {
			if pq.QuestID == prereqID && pq.Status == models.QuestStatusCompleted {
				completed = true
				break
			}
		}
		if !completed {
			return false
		}
	}

	return true
}

// StartQuest starts a quest for a player.
//
// Creates a new PlayerQuest instance, initializes objectives,
// sets time limits (if applicable), and adds to player's quest list.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier to start
//
// Returns:
//   - Pointer to new PlayerQuest instance
//   - Error if quest not found or prerequisites not met
//
// Errors:
//   - "quest not found": Quest ID doesn't exist
//   - "cannot start quest": Prerequisites not met or already active
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) StartQuest(playerID uuid.UUID, questID string) (*models.PlayerQuest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	quest := m.quests[questID]
	if quest == nil {
		return nil, errors.New("quest not found")
	}

	// Check if can start
	if !m.canStartQuestUnsafe(playerID, questID) {
		return nil, errors.New("cannot start quest: prerequisites not met or already active")
	}

	// Create player quest
	pq := models.NewPlayerQuest(playerID, questID)

	// Set time limit if quest has one
	if quest.TimeLimit != nil {
		expiresAt := time.Now().Add(*quest.TimeLimit)
		pq.ExpiresAt = &expiresAt
	}

	// Initialize objectives
	for _, obj := range quest.Objectives {
		pq.Objectives[obj.ID] = 0
	}

	// Add to player quests
	m.playerQuests[playerID] = append(m.playerQuests[playerID], pq)

	return pq, nil
}

// canStartQuestUnsafe checks if can start quest (internal helper).
//
// This method must only be called while holding the write lock.
// It performs the same checks as CanStartQuest but without acquiring the lock.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//
// Returns:
//   - true if player can start the quest
//
// Thread Safety:
// NOT thread-safe. Must be called with m.mu lock held.
func (m *Manager) canStartQuestUnsafe(playerID uuid.UUID, questID string) bool {
	quest := m.quests[questID]
	if quest == nil {
		return false
	}

	// Check if already active or completed
	for _, pq := range m.playerQuests[playerID] {
		if pq.QuestID == questID {
			if pq.Status == models.QuestStatusActive {
				return false
			}
			if pq.Status == models.QuestStatusCompleted && !quest.Repeatable {
				return false
			}
		}
	}

	// Check prerequisites
	for _, prereqID := range quest.Prerequisites {
		completed := false
		for _, pq := range m.playerQuests[playerID] {
			if pq.QuestID == prereqID && pq.Status == models.QuestStatusCompleted {
				completed = true
				break
			}
		}
		if !completed {
			return false
		}
	}

	return true
}

// UpdateObjective updates progress on a quest objective.
//
// Increments objective progress and marks objective complete if
// required amount is reached. Automatically updates quest state.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//   - objectiveID: Objective identifier within quest
//   - amount: Amount to add to objective progress
//
// Returns:
//   - Error if quest not found, not active, or objective doesn't exist
//
// Errors:
//   - "quest not found": No active quest with this ID for player
//   - "quest is not active": Quest exists but is not in active state
//   - "quest definition not found": Quest template missing
//   - "objective not found": Objective ID doesn't exist in quest
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) UpdateObjective(playerID uuid.UUID, questID, objectiveID string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pq := m.findPlayerQuestUnsafe(playerID, questID)
	if pq == nil {
		return errors.New("quest not found")
	}

	if pq.Status != models.QuestStatusActive {
		return errors.New("quest is not active")
	}

	quest := m.quests[questID]
	if quest == nil {
		return errors.New("quest definition not found")
	}

	// Find objective
	var objective *models.QuestObjective
	for _, obj := range quest.Objectives {
		if obj.ID == objectiveID {
			objective = obj
			break
		}
	}

	if objective == nil {
		return errors.New("objective not found")
	}

	// Update progress
	pq.UpdateObjective(objectiveID, amount)

	// Check if objective is complete
	if pq.Objectives[objectiveID] >= objective.Required {
		pq.CompleteObjective(objectiveID)
	}

	return nil
}

// CompleteQuest completes a quest and grants rewards.
//
// Validates all required objectives are complete, then marks
// quest as completed. Caller is responsible for granting rewards
// (credits, experience, items, etc.) from quest.Rewards.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//
// Returns:
//   - Error if quest not found, not active, or objectives incomplete
//
// Errors:
//   - "quest not found": No active quest with this ID
//   - "quest is not active": Quest exists but not active
//   - "quest definition not found": Quest template missing
//   - "quest objectives not complete": Required objectives not finished
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) CompleteQuest(playerID uuid.UUID, questID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pq := m.findPlayerQuestUnsafe(playerID, questID)
	if pq == nil {
		return errors.New("quest not found")
	}

	if pq.Status != models.QuestStatusActive {
		return errors.New("quest is not active")
	}

	quest := m.quests[questID]
	if quest == nil {
		return errors.New("quest definition not found")
	}

	// Check if can complete
	if !pq.CanComplete(quest) {
		return errors.New("quest objectives not complete")
	}

	// Mark as complete
	pq.Complete()

	return nil
}

// AbandonQuest abandons an active quest.
//
// Marks the quest as abandoned, removing it from active quest list.
// No rewards are granted. Quest may be restarted if still available.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//
// Returns:
//   - Error if quest not found or not active
//
// Errors:
//   - "quest not found": No active quest with this ID
//   - "quest is not active": Quest exists but not active
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) AbandonQuest(playerID uuid.UUID, questID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pq := m.findPlayerQuestUnsafe(playerID, questID)
	if pq == nil {
		return errors.New("quest not found")
	}

	if pq.Status != models.QuestStatusActive {
		return errors.New("quest is not active")
	}

	pq.Abandon()
	return nil
}

// findPlayerQuestUnsafe finds an active player quest (internal helper).
//
// This method must only be called while holding the write lock.
//
// Parameters:
//   - playerID: Player UUID
//   - questID: Quest identifier
//
// Returns:
//   - PlayerQuest if found and active, nil otherwise
//
// Thread Safety:
// NOT thread-safe. Must be called with m.mu lock held.
func (m *Manager) findPlayerQuestUnsafe(playerID uuid.UUID, questID string) *models.PlayerQuest {
	for _, pq := range m.playerQuests[playerID] {
		if pq.QuestID == questID && pq.Status == models.QuestStatusActive {
			return pq
		}
	}
	return nil
}

// GetAvailableQuests returns quests available for a player to start.
//
// Returns quests where all prerequisites are met and quest is not
// already active or completed (unless repeatable).
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Slice of available quest templates
//
// Thread Safety:
// Thread-safe. Acquires read lock. Returns a new slice.
func (m *Manager) GetAvailableQuests(playerID uuid.UUID) []*models.Quest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := make([]*models.Quest, 0)
	for _, quest := range m.quests {
		if m.canStartQuestUnsafe(playerID, quest.ID) {
			available = append(available, quest)
		}
	}
	return available
}

// GetStorylineProgress returns completion progress for a storyline.
//
// Calculates percentage of storyline quests completed.
//
// Parameters:
//   - playerID: Player UUID
//   - storylineID: Storyline identifier
//
// Returns:
//   - Progress as float64 (0.0 to 100.0), or 0.0 if storyline not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetStorylineProgress(playerID uuid.UUID, storylineID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storyline := m.storylines[storylineID]
	if storyline == nil {
		return 0.0
	}

	completedMap := make(map[string]bool)
	for _, pq := range m.playerQuests[playerID] {
		if pq.Status == models.QuestStatusCompleted {
			completedMap[pq.QuestID] = true
		}
	}

	return storyline.GetProgress(completedMap)
}

// GetStats returns quest statistics for a player.
//
// Returns counts of quests by status (active, completed, failed, abandoned).
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Map with keys: "active", "completed", "failed", "abandoned", "total"
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetStats(playerID uuid.UUID) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := 0
	completed := 0
	failed := 0
	abandoned := 0

	for _, pq := range m.playerQuests[playerID] {
		switch pq.Status {
		case models.QuestStatusActive:
			active++
		case models.QuestStatusCompleted:
			completed++
		case models.QuestStatusFailed:
			failed++
		case models.QuestStatusAbandoned:
			abandoned++
		}
	}

	return map[string]interface{}{
		"active":    active,
		"completed": completed,
		"failed":    failed,
		"abandoned": abandoned,
		"total":     len(m.playerQuests[playerID]),
	}
}
