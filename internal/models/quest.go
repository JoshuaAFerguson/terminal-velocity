// File: internal/models/quest.go
// Project: Terminal Velocity
// Description: Quest and storyline system - hand-crafted narrative content
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// Quests are hand-crafted story missions that provide narrative structure
// and progression. Unlike missions (procedurally generated), quests are
// pre-designed with specific storylines, characters, and branching paths.
//
// Quest System Features:
//   - Branching narratives with player choices
//   - Multiple objective types (12 types)
//   - Dialogue and character interactions
//   - Prerequisites and quest chains
//   - Repeatable and one-time quests
//   - Hidden/secret quests
//   - Special rewards (ship unlocks, system access, etc.)
//
// Quest Types (7):
//   - Main: Primary storyline quests (linear progression)
//   - Side: Optional content (exploration, character stories)
//   - Faction: Faction-specific storylines
//   - Daily: Repeatable quests that reset daily
//   - Chain: Multi-part quest series
//   - Hidden: Secret quests unlocked by exploration/actions
//   - Event: Limited-time special event quests
//
// Objective Types (12):
//   - Deliver: Bring items to location
//   - Destroy: Eliminate specific targets
//   - Travel: Visit specific location
//   - Collect: Gather items from world
//   - Escort: Protect NPC during journey
//   - Defend: Protect location from enemies
//   - Investigate: Scan or examine objects
//   - Talk: Interact with specific NPCs
//   - Scan: Scan objects in space
//   - Mine: Extract resources
//   - Trade: Complete trades with NPCs
//   - Kill: Defeat specific enemies
//
// Quest Progression:
//   - Prerequisites: Quests that must be completed first
//   - Branching: Player choices affect outcomes
//   - Hidden objectives: Revealed during quest
//   - Optional objectives: Bonus content
//   - Multiple endings: Different conclusion paths
//
// Rewards:
//   - Credits, items, reputation
//   - Experience points
//   - Ship unlocks (new ships become available)
//   - System unlocks (new areas accessible)
//   - Special unique rewards
//
// Storylines:
//   - Connected quests form storylines
//   - Main storyline provides core narrative
//   - Side storylines add depth and content
//   - Faction storylines for each major faction

package models

import (
	"time"

	"github.com/google/uuid"
)

// QuestStatus represents the current status of a quest.
//
// Status transitions:
//   - NotStarted -> Active (when accepted)
//   - Active -> Completed (objectives met)
//   - Active -> Failed (failure condition met)
//   - Active -> Abandoned (player cancels)
type QuestStatus string

const (
	QuestStatusNotStarted QuestStatus = "not_started"
	QuestStatusActive     QuestStatus = "active"
	QuestStatusCompleted  QuestStatus = "completed"
	QuestStatusFailed     QuestStatus = "failed"
	QuestStatusAbandoned  QuestStatus = "abandoned"
)

// QuestType represents different categories of quests
type QuestType string

const (
	QuestTypeMain    QuestType = "main"    // Main storyline quests
	QuestTypeSide    QuestType = "side"    // Side quests
	QuestTypeFaction QuestType = "faction" // Faction-specific quests
	QuestTypeDaily   QuestType = "daily"   // Daily repeatable quests
	QuestTypeChain   QuestType = "chain"   // Part of a quest chain
	QuestTypeHidden  QuestType = "hidden"  // Hidden/secret quests
	QuestTypeEvent   QuestType = "event"   // Special event quests
)

// ObjectiveType represents different types of quest objectives
type ObjectiveType string

const (
	ObjectiveDeliver     ObjectiveType = "deliver"     // Deliver items
	ObjectiveDestroy     ObjectiveType = "destroy"     // Destroy targets
	ObjectiveTravel      ObjectiveType = "travel"      // Travel to location
	ObjectiveCollect     ObjectiveType = "collect"     // Collect items
	ObjectiveEscort      ObjectiveType = "escort"      // Escort NPC
	ObjectiveDefend      ObjectiveType = "defend"      // Defend location
	ObjectiveInvestigate ObjectiveType = "investigate" // Investigate location
	ObjectiveTalk        ObjectiveType = "talk"        // Talk to NPC
	ObjectiveScan        ObjectiveType = "scan"        // Scan objects
	ObjectiveMine        ObjectiveType = "mine"        // Mine resources
	ObjectiveTrade       ObjectiveType = "trade"       // Complete trades
	ObjectiveKill        ObjectiveType = "kill"        // Kill specific targets
)

// QuestObjective represents a single objective within a quest
type QuestObjective struct {
	ID          string        `json:"id"`
	Type        ObjectiveType `json:"type"`
	Description string        `json:"description"`
	Target      string        `json:"target"`   // Target ID (item, NPC, location, etc.)
	Required    int           `json:"required"` // Required amount
	Current     int           `json:"current"`  // Current progress
	Optional    bool          `json:"optional"` // Is this objective optional?
	Hidden      bool          `json:"hidden"`   // Hidden until revealed
	Completed   bool          `json:"completed"`
}

// QuestReward represents rewards given upon quest completion
type QuestReward struct {
	Credits      int64          `json:"credits"`
	Items        map[string]int `json:"items"`      // itemID -> quantity
	Reputation   map[string]int `json:"reputation"` // factionID -> amount
	Experience   int            `json:"experience"`
	ShipUnlock   string         `json:"ship_unlock"`   // Unlock a ship type
	SystemUnlock string         `json:"system_unlock"` // Unlock a star system
	Special      string         `json:"special"`       // Special reward description
}

// QuestChoice represents a choice the player can make
type QuestChoice struct {
	ID           string                 `json:"id"`
	Text         string                 `json:"text"`
	Description  string                 `json:"description"`
	Requirements map[string]interface{} `json:"requirements"`   // Requirements to select
	Consequences string                 `json:"consequences"`   // Description of consequences
	LeadsToQuest string                 `json:"leads_to_quest"` // Quest ID this choice leads to
}

// QuestDialogue represents dialogue in a quest
type QuestDialogue struct {
	Speaker string        `json:"speaker"`
	Text    string        `json:"text"`
	Choices []QuestChoice `json:"choices"`
}

// Quest represents a quest or storyline mission
type Quest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        QuestType `json:"type"`
	Level       int       `json:"level"` // Recommended level

	// Quest flow
	Prerequisites []string          `json:"prerequisites"` // Quest IDs required
	Objectives    []*QuestObjective `json:"objectives"`
	Rewards       QuestReward       `json:"rewards"`

	// Dialogue and story
	StartDialogue    []QuestDialogue `json:"start_dialogue"`
	CompleteDialogue []QuestDialogue `json:"complete_dialogue"`

	// Metadata
	Giver      string         `json:"giver"`      // NPC who gives quest
	Location   uuid.UUID      `json:"location"`   // System where quest starts
	TimeLimit  *time.Duration `json:"time_limit"` // Optional time limit
	Repeatable bool           `json:"repeatable"`

	// Branching
	NextQuests      []string `json:"next_quests"`      // Quests unlocked on completion
	FailureQuests   []string `json:"failure_quests"`   // Quests unlocked on failure
	AlternateEnding string   `json:"alternate_ending"` // Alt ending quest ID
}

// PlayerQuest represents a player's progress on a quest
type PlayerQuest struct {
	ID       uuid.UUID   `json:"id"`
	PlayerID uuid.UUID   `json:"player_id"`
	QuestID  string      `json:"quest_id"`
	Status   QuestStatus `json:"status"`

	// Progress tracking
	Objectives          map[string]int `json:"objectives"` // objectiveID -> current count
	CompletedObjectives []string       `json:"completed_objectives"`

	// Choices made
	ChoicesMade []string `json:"choices_made"` // Choice IDs selected

	// Timing
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ExpiresAt   *time.Time `json:"expires_at"`

	// Metadata
	CurrentStage  int    `json:"current_stage"`
	FailureReason string `json:"failure_reason"`
}

// Storyline represents a series of connected quests
type Storyline struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Quests      []string `json:"quests"`     // Quest IDs in order
	MainStory   bool     `json:"main_story"` // Is this the main storyline?
	OrderIndex  int      `json:"order_index"`
}

// NewQuest creates a new quest
func NewQuest(id, title, description string, questType QuestType) *Quest {
	return &Quest{
		ID:               id,
		Title:            title,
		Description:      description,
		Type:             questType,
		Level:            1,
		Prerequisites:    make([]string, 0),
		Objectives:       make([]*QuestObjective, 0),
		Rewards:          QuestReward{Items: make(map[string]int), Reputation: make(map[string]int)},
		StartDialogue:    make([]QuestDialogue, 0),
		CompleteDialogue: make([]QuestDialogue, 0),
		Repeatable:       false,
		NextQuests:       make([]string, 0),
		FailureQuests:    make([]string, 0),
	}
}

// AddObjective adds an objective to the quest
func (q *Quest) AddObjective(objective *QuestObjective) {
	q.Objectives = append(q.Objectives, objective)
}

// AddChoice adds a choice to quest dialogue
func (q *Quest) AddStartChoice(speaker, text string, choices []QuestChoice) {
	q.StartDialogue = append(q.StartDialogue, QuestDialogue{
		Speaker: speaker,
		Text:    text,
		Choices: choices,
	})
}

// NewPlayerQuest creates a new player quest instance
func NewPlayerQuest(playerID uuid.UUID, questID string) *PlayerQuest {
	return &PlayerQuest{
		ID:                  uuid.New(),
		PlayerID:            playerID,
		QuestID:             questID,
		Status:              QuestStatusActive,
		Objectives:          make(map[string]int),
		CompletedObjectives: make([]string, 0),
		ChoicesMade:         make([]string, 0),
		StartedAt:           time.Now(),
		CurrentStage:        0,
	}
}

// UpdateObjective updates progress on an objective
func (pq *PlayerQuest) UpdateObjective(objectiveID string, amount int) {
	current := pq.Objectives[objectiveID]
	pq.Objectives[objectiveID] = current + amount
}

// CompleteObjective marks an objective as completed
func (pq *PlayerQuest) CompleteObjective(objectiveID string) {
	for _, completed := range pq.CompletedObjectives {
		if completed == objectiveID {
			return
		}
	}
	pq.CompletedObjectives = append(pq.CompletedObjectives, objectiveID)
}

// MakeChoice records a choice made by the player
func (pq *PlayerQuest) MakeChoice(choiceID string) {
	pq.ChoicesMade = append(pq.ChoicesMade, choiceID)
}

// IsObjectiveComplete checks if an objective is completed
func (pq *PlayerQuest) IsObjectiveComplete(objectiveID string) bool {
	for _, completed := range pq.CompletedObjectives {
		if completed == objectiveID {
			return true
		}
	}
	return false
}

// GetProgress returns the completion progress (0.0 to 1.0)
func (pq *PlayerQuest) GetProgress(quest *Quest) float64 {
	if len(quest.Objectives) == 0 {
		return 0.0
	}

	completed := 0
	for _, obj := range quest.Objectives {
		if obj.Optional {
			continue
		}
		if pq.IsObjectiveComplete(obj.ID) {
			completed++
		}
	}

	required := 0
	for _, obj := range quest.Objectives {
		if !obj.Optional {
			required++
		}
	}

	if required == 0 {
		return 1.0
	}

	return float64(completed) / float64(required)
}

// CanComplete checks if all required objectives are complete
func (pq *PlayerQuest) CanComplete(quest *Quest) bool {
	for _, obj := range quest.Objectives {
		if obj.Optional || obj.Hidden {
			continue
		}
		if !pq.IsObjectiveComplete(obj.ID) {
			return false
		}
	}
	return true
}

// Complete marks the quest as completed
func (pq *PlayerQuest) Complete() {
	pq.Status = QuestStatusCompleted
	now := time.Now()
	pq.CompletedAt = &now
}

// Fail marks the quest as failed
func (pq *PlayerQuest) Fail(reason string) {
	pq.Status = QuestStatusFailed
	pq.FailureReason = reason
	now := time.Now()
	pq.CompletedAt = &now
}

// Abandon marks the quest as abandoned
func (pq *PlayerQuest) Abandon() {
	pq.Status = QuestStatusAbandoned
	now := time.Now()
	pq.CompletedAt = &now
}

// IsExpired checks if the quest has expired
func (pq *PlayerQuest) IsExpired() bool {
	if pq.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*pq.ExpiresAt)
}

// NewStoryline creates a new storyline
func NewStoryline(id, title, description string, mainStory bool) *Storyline {
	return &Storyline{
		ID:          id,
		Title:       title,
		Description: description,
		Quests:      make([]string, 0),
		MainStory:   mainStory,
		OrderIndex:  0,
	}
}

// AddQuest adds a quest to the storyline
func (s *Storyline) AddQuest(questID string) {
	s.Quests = append(s.Quests, questID)
}

// GetProgress returns storyline completion progress
func (s *Storyline) GetProgress(completedQuests map[string]bool) float64 {
	if len(s.Quests) == 0 {
		return 0.0
	}

	completed := 0
	for _, questID := range s.Quests {
		if completedQuests[questID] {
			completed++
		}
	}

	return float64(completed) / float64(len(s.Quests))
}
