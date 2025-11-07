// File: internal/events/manager.go
// Project: Terminal Velocity
// Description: Dynamic event management and scheduling
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles server events and scheduling

var log = logger.WithComponent("Events")

type Manager struct {
	mu             sync.RWMutex
	events         map[string]*models.Event
	participations map[uuid.UUID][]*models.EventParticipation // playerID -> participations
	leaderboards   map[string]*models.EventLeaderboard
	notifications  map[uuid.UUID][]*models.EventNotification

	// Background worker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewManager creates a new event manager
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		events:         make(map[string]*models.Event),
		participations: make(map[uuid.UUID][]*models.EventParticipation),
		leaderboards:   make(map[string]*models.EventLeaderboard),
		notifications:  make(map[uuid.UUID][]*models.EventNotification),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize default events
	m.initializeDefaultEvents()

	// Start event scheduler
	m.wg.Add(1)
	go m.eventScheduler()

	return m
}

// initializeDefaultEvents creates sample events
func (m *Manager) initializeDefaultEvents() {
	// Trading Competition
	tradingComp := models.NewEvent(
		"trading_comp_01",
		"Trade Route Challenge",
		"Complete as many trades as possible for the highest profit!",
		models.EventTypeTrading,
		2*time.Hour,
	)
	tradingComp.MinLevel = 1
	tradingComp.AddObjective(models.EventObjective{
		ID:          "obj_trades",
		Description: "Complete profitable trades",
		Target:      "trade_profit",
		Required:    100000,
		Individual:  true,
	})
	tradingComp.Rewards = models.EventReward{
		Credits:    50000,
		Experience: 1000,
		Badge:      "Master Trader",
	}
	tradingComp.CreditsMultiplier = 1.5
	m.RegisterEvent(tradingComp)

	// Combat Tournament
	combatTournament := models.NewEvent(
		"combat_tournament_01",
		"Galactic Combat Tournament",
		"Prove your combat skills against the best pilots!",
		models.EventTypeTournament,
		3*time.Hour,
	)
	combatTournament.MinLevel = 5
	combatTournament.MaxParticipants = 32
	combatTournament.AddObjective(models.EventObjective{
		ID:          "obj_wins",
		Description: "Win combat encounters",
		Target:      "combat_wins",
		Required:    10,
		Individual:  true,
	})
	combatTournament.Rewards = models.EventReward{
		Credits:    100000,
		Experience: 2000,
		Title:      "Combat Champion",
		Exclusive:  "champion_ship_skin",
	}
	combatTournament.ProgressRewards = map[int]models.EventReward{
		50: {Credits: 25000},
		75: {Credits: 50000},
	}
	m.RegisterEvent(combatTournament)

	// Community Expedition
	expedition := models.NewEvent(
		"expedition_01",
		"Deep Space Expedition",
		"Join forces to explore the uncharted void sector!",
		models.EventTypeExpedition,
		24*time.Hour,
	)
	expedition.MinLevel = 3
	expedition.RequiredPlayers = 10
	expedition.CommunityGoal = 1000000
	expedition.AddObjective(models.EventObjective{
		ID:          "obj_explore",
		Description: "Explore systems (community goal)",
		Target:      "systems_explored",
		Required:    1000,
		Individual:  false,
	})
	expedition.Rewards = models.EventReward{
		Credits:    75000,
		Experience: 1500,
		Exclusive:  "void_sector_access",
	}
	m.RegisterEvent(expedition)

	// Boss Encounter
	boss := models.NewEvent(
		"boss_void_leviathan",
		"Void Leviathan Appears!",
		"A massive alien vessel threatens the sector. All pilots respond!",
		models.EventTypeBoss,
		1*time.Hour,
	)
	boss.MinLevel = 7
	boss.CommunityGoal = 5000000 // Total damage needed
	boss.AddObjective(models.EventObjective{
		ID:          "obj_damage",
		Description: "Deal damage to the Void Leviathan",
		Target:      "boss_damage",
		Required:    5000000,
		Individual:  false,
	})
	boss.Rewards = models.EventReward{
		Credits:    200000,
		Items:      map[string]int{"void_crystal": 5},
		Experience: 5000,
		Badge:      "Leviathan Slayer",
	}
	boss.DropRateMultiplier = 2.0
	m.RegisterEvent(boss)

	// Festival Event
	festival := models.NewEvent(
		"festival_harvest",
		"Harvest Festival",
		"Celebrate the harvest season with bonuses to trading and gathering!",
		models.EventTypeFestival,
		12*time.Hour,
	)
	festival.MinLevel = 1
	festival.CreditsMultiplier = 2.0
	festival.ExperienceMultiplier = 1.5
	festival.Rewards = models.EventReward{
		Credits: 25000,
		Title:   "Festival Goer",
	}
	m.RegisterEvent(festival)
}

// RegisterEvent registers an event
func (m *Manager) RegisterEvent(event *models.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.ID] = event
	m.leaderboards[event.ID] = models.NewEventLeaderboard(event.ID)
}

// GetEvent returns an event by ID
func (m *Manager) GetEvent(eventID string) *models.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events[eventID]
}

// GetActiveEvents returns all active events
func (m *Manager) GetActiveEvents() []*models.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*models.Event, 0)
	for _, event := range m.events {
		if event.IsActive() {
			active = append(active, event)
		}
	}
	return active
}

// GetScheduledEvents returns upcoming events
func (m *Manager) GetScheduledEvents() []*models.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scheduled := make([]*models.Event, 0)
	for _, event := range m.events {
		if event.Status == models.EventStatusScheduled {
			scheduled = append(scheduled, event)
		}
	}
	return scheduled
}

// JoinEvent allows a player to join an event
func (m *Manager) JoinEvent(playerID uuid.UUID, eventID string, playerLevel int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event := m.events[eventID]
	if event == nil {
		return fmt.Errorf("event not found")
	}

	if !event.CanJoin(playerLevel) {
		return fmt.Errorf("cannot join event")
	}

	// Check if already participating
	for _, p := range m.participations[playerID] {
		if p.EventID == eventID {
			return fmt.Errorf("already participating")
		}
	}

	participation := models.NewEventParticipation(playerID, eventID)
	m.participations[playerID] = append(m.participations[playerID], participation)
	event.CurrentCount++

	return nil
}

// UpdateProgress updates a player's event progress
func (m *Manager) UpdateProgress(playerID uuid.UUID, eventID, objectiveID string, amount int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var participation *models.EventParticipation
	for _, p := range m.participations[playerID] {
		if p.EventID == eventID {
			participation = p
			break
		}
	}

	if participation == nil {
		return
	}

	event := m.events[eventID]
	if event == nil {
		return
	}

	participation.UpdateProgress(objectiveID, amount)

	// Update community progress
	for _, obj := range event.Objectives {
		if obj.ID == objectiveID && !obj.Individual {
			event.CommunityProgress += amount
		}
	}

	// Update leaderboard
	m.updateLeaderboardUnsafe(eventID, playerID, participation)
}

// updateLeaderboardUnsafe updates leaderboard (must hold lock)
func (m *Manager) updateLeaderboardUnsafe(eventID string, playerID uuid.UUID, participation *models.EventParticipation) {
	lb := m.leaderboards[eventID]
	if lb == nil {
		return
	}

	// Find or create entry
	found := false
	for i, entry := range lb.Entries {
		if entry.PlayerID == playerID {
			lb.Entries[i].Score = participation.Score
			lb.Entries[i].Completed = participation.CompletedAt != nil
			found = true
			break
		}
	}

	if !found {
		lb.AddEntry(models.EventLeaderboardEntry{
			PlayerID:  playerID,
			Username:  "Player", // Would get from player data
			Score:     participation.Score,
			Completed: participation.CompletedAt != nil,
		})
	} else {
		lb.SortEntries()
		lb.UpdateRanks()
	}
}

// GetLeaderboard returns an event's leaderboard
func (m *Manager) GetLeaderboard(eventID string) *models.EventLeaderboard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.leaderboards[eventID]
}

// GetParticipation returns a player's participation for an event
func (m *Manager) GetParticipation(playerID uuid.UUID, eventID string) *models.EventParticipation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.participations[playerID] {
		if p.EventID == eventID {
			return p
		}
	}
	return nil
}

// GetPlayerEvents returns all events a player is participating in
func (m *Manager) GetPlayerEvents(playerID uuid.UUID) []*models.EventParticipation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.participations[playerID]
}

// NotifyPlayer sends a notification to a player
func (m *Manager) NotifyPlayer(playerID uuid.UUID, eventID string, notifType models.EventNotificationType, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	notif := models.NewEventNotification(playerID, eventID, notifType, message)
	m.notifications[playerID] = append(m.notifications[playerID], notif)

	// Trim old notifications
	if len(m.notifications[playerID]) > 50 {
		m.notifications[playerID] = m.notifications[playerID][len(m.notifications[playerID])-50:]
	}
}

// GetNotifications returns unread notifications for a player
func (m *Manager) GetNotifications(playerID uuid.UUID) []*models.EventNotification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	unread := make([]*models.EventNotification, 0)
	for _, n := range m.notifications[playerID] {
		if !n.Read {
			unread = append(unread, n)
		}
	}
	return unread
}

// eventScheduler periodically checks and updates event statuses
func (m *Manager) eventScheduler() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateEvents()
		}
	}
}

// updateEvents checks event timers and transitions states
func (m *Manager) updateEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, event := range m.events {
		switch event.Status {
		case models.EventStatusActive:
			// Check if ending soon (5 minutes remaining)
			if event.TimeRemaining() <= 5*time.Minute && event.Status != models.EventStatusEnding {
				event.Status = models.EventStatusEnding
				// Notify participants
				for playerID, participations := range m.participations {
					for _, p := range participations {
						if p.EventID == event.ID {
							m.notifyPlayerUnsafe(playerID, event.ID, models.NotificationEventEnding, "Event ending soon!")
						}
					}
				}
			}

			// Check if ended
			if now.After(event.EndTime) {
				event.End()
				// Notify participants
				for playerID, participations := range m.participations {
					for _, p := range participations {
						if p.EventID == event.ID {
							m.notifyPlayerUnsafe(playerID, event.ID, models.NotificationEventComplete, "Event complete! Check your rewards.")
						}
					}
				}
			}
		}
	}
}

// notifyPlayerUnsafe sends notification (must hold lock)
func (m *Manager) notifyPlayerUnsafe(playerID uuid.UUID, eventID string, notifType models.EventNotificationType, message string) {
	notif := models.NewEventNotification(playerID, eventID, notifType, message)
	m.notifications[playerID] = append(m.notifications[playerID], notif)
}

// Shutdown gracefully shuts down the event manager
func (m *Manager) Shutdown() {
	m.cancel()
	m.wg.Wait()
}

// GetStats returns event system statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeCount := 0
	scheduledCount := 0
	endedCount := 0

	for _, event := range m.events {
		switch event.Status {
		case models.EventStatusActive:
			activeCount++
		case models.EventStatusScheduled:
			scheduledCount++
		case models.EventStatusEnded:
			endedCount++
		}
	}

	totalParticipations := 0
	for _, participations := range m.participations {
		totalParticipations += len(participations)
	}

	return map[string]interface{}{
		"total_events":         len(m.events),
		"active_events":        activeCount,
		"scheduled_events":     scheduledCount,
		"ended_events":         endedCount,
		"total_participations": totalParticipations,
	}
}
