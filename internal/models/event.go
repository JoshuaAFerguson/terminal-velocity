// File: internal/models/event.go
// Project: Terminal Velocity
// Description: Dynamic events and server event models
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents different types of server events
type EventType string

const (
	EventTypeTrading     EventType = "trading"      // Trading competitions
	EventTypeCombat      EventType = "combat"       // Combat challenges
	EventTypeRacing      EventType = "racing"       // Racing events
	EventTypeScavenging  EventType = "scavenging"   // Resource gathering
	EventTypeInvasion    EventType = "invasion"     // System invasions
	EventTypeFestival    EventType = "festival"     // Peaceful festivals
	EventTypeTournament  EventType = "tournament"   // PvP tournaments
	EventTypeExpedition  EventType = "expedition"   // Exploration events
	EventTypeBoss        EventType = "boss"         // Boss encounters
	EventTypeCommunity   EventType = "community"    // Community goals
)

// EventStatus represents the current status of an event
type EventStatus string

const (
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusActive    EventStatus = "active"
	EventStatusEnding    EventStatus = "ending"
	EventStatusEnded     EventStatus = "ended"
	EventStatusCancelled EventStatus = "cancelled"
)

// EventReward represents rewards for event participation
type EventReward struct {
	Credits      int64          `json:"credits"`
	Items        map[string]int `json:"items"`
	Reputation   map[string]int `json:"reputation"`
	Experience   int            `json:"experience"`
	Title        string         `json:"title"`         // Special title awarded
	Badge        string         `json:"badge"`         // Achievement badge
	Exclusive    string         `json:"exclusive"`     // Exclusive item/ship
	LeaderboardTop int          `json:"leaderboard_top"` // Top N get rewards
}

// Event represents a server-wide event
type Event struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Type        EventType     `json:"type"`
	Status      EventStatus   `json:"status"`

	// Timing
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`

	// Participation
	MinLevel        int       `json:"min_level"`
	MaxParticipants int       `json:"max_participants"`
	CurrentCount    int       `json:"current_count"`
	RequiredPlayers int       `json:"required_players"` // Min to start

	// Objectives
	Objectives      []EventObjective `json:"objectives"`
	CommunityGoal   int64            `json:"community_goal"`   // Total goal for all players
	CommunityProgress int64          `json:"community_progress"`

	// Rewards
	Rewards         EventReward      `json:"rewards"`
	ProgressRewards map[int]EventReward `json:"progress_rewards"` // % -> rewards

	// Location
	SystemID        *uuid.UUID       `json:"system_id"`
	SystemName      string           `json:"system_name"`

	// Modifiers
	CreditsMultiplier float64        `json:"credits_multiplier"`
	ExperienceMultiplier float64     `json:"experience_multiplier"`
	DropRateMultiplier float64       `json:"drop_rate_multiplier"`
}

// EventObjective represents a specific event objective
type EventObjective struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Target      string `json:"target"`
	Required    int64  `json:"required"`
	Individual  bool   `json:"individual"` // Individual vs community goal
}

// EventParticipation tracks a player's participation in an event
type EventParticipation struct {
	ID              uuid.UUID          `json:"id"`
	PlayerID        uuid.UUID          `json:"player_id"`
	EventID         string             `json:"event_id"`
	JoinedAt        time.Time          `json:"joined_at"`
	Progress        map[string]int64   `json:"progress"` // objectiveID -> progress
	Score           int64              `json:"score"`
	Rank            int                `json:"rank"`
	RewardsClaimed  bool               `json:"rewards_claimed"`
	CompletedAt     *time.Time         `json:"completed_at"`
}

// EventLeaderboard represents leaderboard standings for an event
type EventLeaderboard struct {
	EventID   string                 `json:"event_id"`
	UpdatedAt time.Time              `json:"updated_at"`
	Entries   []EventLeaderboardEntry `json:"entries"`
}

// EventLeaderboardEntry represents a single leaderboard entry
type EventLeaderboardEntry struct {
	Rank       int       `json:"rank"`
	PlayerID   uuid.UUID `json:"player_id"`
	Username   string    `json:"username"`
	Score      int64     `json:"score"`
	Completed  bool      `json:"completed"`
}

// EventNotification represents a notification about an event
type EventNotification struct {
	ID        uuid.UUID            `json:"id"`
	PlayerID  uuid.UUID            `json:"player_id"`
	EventID   string               `json:"event_id"`
	Type      EventNotificationType `json:"type"`
	Message   string               `json:"message"`
	Timestamp time.Time            `json:"timestamp"`
	Read      bool                 `json:"read"`
}

// EventNotificationType represents types of event notifications
type EventNotificationType string

const (
	NotificationEventStarting EventNotificationType = "event_starting"
	NotificationEventActive   EventNotificationType = "event_active"
	NotificationEventEnding   EventNotificationType = "event_ending"
	NotificationEventComplete EventNotificationType = "event_complete"
	NotificationRewardReady   EventNotificationType = "reward_ready"
	NotificationRankChange    EventNotificationType = "rank_change"
)

// NewEvent creates a new event
func NewEvent(id, title, description string, eventType EventType, duration time.Duration) *Event {
	return &Event{
		ID:                   id,
		Title:                title,
		Description:          description,
		Type:                 eventType,
		Status:               EventStatusScheduled,
		Duration:             duration,
		MinLevel:             1,
		MaxParticipants:      0, // 0 = unlimited
		RequiredPlayers:      1,
		Objectives:           make([]EventObjective, 0),
		ProgressRewards:      make(map[int]EventReward),
		CreditsMultiplier:    1.0,
		ExperienceMultiplier: 1.0,
		DropRateMultiplier:   1.0,
	}
}

// AddObjective adds an objective to the event
func (e *Event) AddObjective(objective EventObjective) {
	e.Objectives = append(e.Objectives, objective)
}

// Start starts the event
func (e *Event) Start() {
	e.Status = EventStatusActive
	e.StartTime = time.Now()
	e.EndTime = e.StartTime.Add(e.Duration)
}

// End ends the event
func (e *Event) End() {
	e.Status = EventStatusEnded
}

// IsActive checks if the event is currently active
func (e *Event) IsActive() bool {
	return e.Status == EventStatusActive && time.Now().Before(e.EndTime)
}

// TimeRemaining returns the time remaining in the event
func (e *Event) TimeRemaining() time.Duration {
	if !e.IsActive() {
		return 0
	}
	return time.Until(e.EndTime)
}

// GetProgressPercent returns the community goal progress percentage
func (e *Event) GetProgressPercent() float64 {
	if e.CommunityGoal == 0 {
		return 0
	}
	return float64(e.CommunityProgress) / float64(e.CommunityGoal) * 100
}

// CanJoin checks if a player can join the event
func (e *Event) CanJoin(playerLevel int) bool {
	if e.Status != EventStatusActive && e.Status != EventStatusScheduled {
		return false
	}
	if playerLevel < e.MinLevel {
		return false
	}
	if e.MaxParticipants > 0 && e.CurrentCount >= e.MaxParticipants {
		return false
	}
	return true
}

// NewEventParticipation creates a new event participation record
func NewEventParticipation(playerID uuid.UUID, eventID string) *EventParticipation {
	return &EventParticipation{
		ID:             uuid.New(),
		PlayerID:       playerID,
		EventID:        eventID,
		JoinedAt:       time.Now(),
		Progress:       make(map[string]int64),
		Score:          0,
		Rank:           0,
		RewardsClaimed: false,
	}
}

// UpdateProgress updates progress for an objective
func (ep *EventParticipation) UpdateProgress(objectiveID string, amount int64) {
	current := ep.Progress[objectiveID]
	ep.Progress[objectiveID] = current + amount
	ep.Score += amount
}

// IsComplete checks if all objectives are complete
func (ep *EventParticipation) IsComplete(event *Event) bool {
	for _, obj := range event.Objectives {
		if !obj.Individual {
			continue // Skip community objectives
		}
		if ep.Progress[obj.ID] < obj.Required {
			return false
		}
	}
	return true
}

// Complete marks the participation as completed
func (ep *EventParticipation) Complete() {
	now := time.Now()
	ep.CompletedAt = &now
}

// NewEventLeaderboard creates a new event leaderboard
func NewEventLeaderboard(eventID string) *EventLeaderboard {
	return &EventLeaderboard{
		EventID:   eventID,
		UpdatedAt: time.Now(),
		Entries:   make([]EventLeaderboardEntry, 0),
	}
}

// AddEntry adds an entry to the leaderboard
func (el *EventLeaderboard) AddEntry(entry EventLeaderboardEntry) {
	el.Entries = append(el.Entries, entry)
	el.SortEntries()
	el.UpdateRanks()
}

// SortEntries sorts entries by score (descending)
func (el *EventLeaderboard) SortEntries() {
	// Simple bubble sort
	for i := 0; i < len(el.Entries); i++ {
		for j := i + 1; j < len(el.Entries); j++ {
			if el.Entries[i].Score < el.Entries[j].Score {
				el.Entries[i], el.Entries[j] = el.Entries[j], el.Entries[i]
			}
		}
	}
}

// UpdateRanks updates rank numbers for all entries
func (el *EventLeaderboard) UpdateRanks() {
	for i := range el.Entries {
		el.Entries[i].Rank = i + 1
	}
	el.UpdatedAt = time.Now()
}

// GetPlayerRank returns a player's rank
func (el *EventLeaderboard) GetPlayerRank(playerID uuid.UUID) int {
	for _, entry := range el.Entries {
		if entry.PlayerID == playerID {
			return entry.Rank
		}
	}
	return 0
}

// GetTopN returns the top N entries
func (el *EventLeaderboard) GetTopN(n int) []EventLeaderboardEntry {
	if n > len(el.Entries) {
		n = len(el.Entries)
	}
	return el.Entries[:n]
}

// NewEventNotification creates a new event notification
func NewEventNotification(playerID uuid.UUID, eventID string, notifType EventNotificationType, message string) *EventNotification {
	return &EventNotification{
		ID:        uuid.New(),
		PlayerID:  playerID,
		EventID:   eventID,
		Type:      notifType,
		Message:   message,
		Timestamp: time.Now(),
		Read:      false,
	}
}
