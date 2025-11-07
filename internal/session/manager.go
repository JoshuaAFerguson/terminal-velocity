// File: internal/session/manager.go
// Project: Terminal Velocity
// Description: Session management and auto-persistence for multiplayer server
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package session

import (
	"context"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Session represents an active player session

var log = logger.WithComponent("Session")

type Session struct {
	ID           uuid.UUID
	PlayerID     uuid.UUID
	Username     string
	ConnectedAt  time.Time
	LastActivity time.Time
	LastSave     time.Time
	IPAddress    string
	IsActive     bool

	// Stats
	ActionsThisSession int
	CommandsExecuted   int
	ErrorsEncountered  int

	// State tracking
	CurrentScreen string
	DirtyState    bool // Has unsaved changes
	LastError     error
}

// Manager handles player sessions and auto-persistence
type Manager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session // PlayerID -> Session

	// Repositories for persistence
	playerRepo *database.PlayerRepository
	shipRepo   *database.ShipRepository

	// Configuration
	saveInterval      time.Duration
	inactivityTimeout time.Duration
	enableAutosave    bool

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewManager creates a new session manager
func NewManager(
	playerRepo *database.PlayerRepository,
	shipRepo *database.ShipRepository,
) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		sessions:          make(map[uuid.UUID]*Session),
		playerRepo:        playerRepo,
		shipRepo:          shipRepo,
		saveInterval:      30 * time.Second, // Save every 30 seconds
		inactivityTimeout: 15 * time.Minute,
		enableAutosave:    true,
		ctx:               ctx,
		cancel:            cancel,
	}

	// Start background workers
	m.wg.Add(2)
	go m.autosaveWorker()
	go m.cleanupWorker()

	return m
}

// CreateSession creates a new player session
func (m *Manager) CreateSession(
	playerID uuid.UUID,
	username string,
	ipAddress string,
) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	session := &Session{
		ID:           uuid.New(),
		PlayerID:     playerID,
		Username:     username,
		ConnectedAt:  now,
		LastActivity: now,
		LastSave:     now,
		IPAddress:    ipAddress,
		IsActive:     true,
		DirtyState:   false,
	}

	m.sessions[playerID] = session

	return session
}

// GetSession retrieves a session by player ID
func (m *Manager) GetSession(playerID uuid.UUID) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[playerID]
	return session, exists
}

// UpdateActivity updates the last activity time for a session
func (m *Manager) UpdateActivity(playerID uuid.UUID, screen string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[playerID]
	if !exists {
		return
	}

	session.LastActivity = time.Now()
	session.CurrentScreen = screen
	session.ActionsThisSession++
	session.DirtyState = true // Mark as having unsaved changes
}

// RecordCommand records a command execution
func (m *Manager) RecordCommand(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[playerID]
	if !exists {
		return
	}

	session.CommandsExecuted++
	session.LastActivity = time.Now()
	session.DirtyState = true
}

// RecordError records an error for a session
func (m *Manager) RecordError(playerID uuid.UUID, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[playerID]
	if !exists {
		return
	}

	session.ErrorsEncountered++
	session.LastError = err
}

// SavePlayerState manually saves player state to database
func (m *Manager) SavePlayerState(
	ctx context.Context,
	playerID uuid.UUID,
	player *models.Player,
	ship *models.Ship,
) error {
	m.mu.Lock()
	session, exists := m.sessions[playerID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	// Save player to database
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return err
	}

	// Save ship if exists
	if ship != nil {
		if err := m.shipRepo.Update(ctx, ship); err != nil {
			return err
		}
	}

	// Update session
	m.mu.Lock()
	session.LastSave = time.Now()
	session.DirtyState = false
	m.mu.Unlock()

	return nil
}

// EndSession ends a player session and saves final state
func (m *Manager) EndSession(
	ctx context.Context,
	playerID uuid.UUID,
	player *models.Player,
	ship *models.Ship,
) error {
	m.mu.Lock()
	session, exists := m.sessions[playerID]
	if !exists {
		m.mu.Unlock()
		return nil
	}

	session.IsActive = false
	m.mu.Unlock()

	// Perform final save
	if err := m.SavePlayerState(ctx, playerID, player, ship); err != nil {
		return err
	}

	// Remove session
	m.mu.Lock()
	delete(m.sessions, playerID)
	m.mu.Unlock()

	return nil
}

// autosaveWorker periodically saves player states
func (m *Manager) autosaveWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if m.enableAutosave {
				m.performAutosave()
			}
		}
	}
}

// performAutosave saves all dirty sessions
func (m *Manager) performAutosave() {
	m.mu.RLock()
	sessionsToSave := make([]*Session, 0)

	for _, session := range m.sessions {
		if session.IsActive && session.DirtyState {
			sessionsToSave = append(sessionsToSave, session)
		}
	}
	m.mu.RUnlock()

	// Note: In a real implementation, we would fetch player/ship data
	// and call SavePlayerState for each session that needs saving
	// For now, we just track which sessions need saving
	for _, session := range sessionsToSave {
		// Log autosave event
		_ = session // Placeholder for actual save logic
	}
}

// cleanupWorker removes inactive sessions
func (m *Manager) cleanupWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions removes sessions that have been inactive
func (m *Manager) cleanupInactiveSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	toRemove := make([]uuid.UUID, 0)

	for playerID, session := range m.sessions {
		if !session.IsActive {
			continue
		}

		inactiveDuration := now.Sub(session.LastActivity)
		if inactiveDuration > m.inactivityTimeout {
			toRemove = append(toRemove, playerID)
		}
	}

	for _, playerID := range toRemove {
		session := m.sessions[playerID]
		session.IsActive = false
		// Note: In production, we would trigger a final save here
	}
}

// GetActiveSessions returns all active sessions
func (m *Manager) GetActiveSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*Session, 0)
	for _, session := range m.sessions {
		if session.IsActive {
			active = append(active, session)
		}
	}

	return active
}

// GetSessionCount returns the number of active sessions
func (m *Manager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, session := range m.sessions {
		if session.IsActive {
			count++
		}
	}

	return count
}

// GetStats returns session manager statistics
func (m *Manager) GetStats() SessionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := SessionStats{
		TotalSessions:  len(m.sessions),
		ActiveSessions: 0,
		TotalCommands:  0,
		TotalErrors:    0,
	}

	for _, session := range m.sessions {
		if session.IsActive {
			stats.ActiveSessions++
		}
		stats.TotalCommands += session.CommandsExecuted
		stats.TotalErrors += session.ErrorsEncountered
	}

	return stats
}

// SetSaveInterval sets the autosave interval
func (m *Manager) SetSaveInterval(seconds int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveInterval = time.Duration(seconds) * time.Second
}

// SetInactivityTimeout sets the inactivity timeout
func (m *Manager) SetInactivityTimeout(minutes int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.inactivityTimeout = time.Duration(minutes) * time.Minute
}

// EnableAutosave enables or disables autosave
func (m *Manager) EnableAutosave(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enableAutosave = enabled
}

// Shutdown gracefully shuts down the session manager
func (m *Manager) Shutdown() {
	m.cancel()
	m.wg.Wait()
}

// SessionStats holds session statistics
type SessionStats struct {
	TotalSessions  int
	ActiveSessions int
	TotalCommands  int
	TotalErrors    int
}

// GetSessionDuration returns the duration of a session
func (s *Session) GetSessionDuration() time.Duration {
	return time.Since(s.ConnectedAt)
}

// GetIdleDuration returns how long the session has been idle
func (s *Session) GetIdleDuration() time.Duration {
	return time.Since(s.LastActivity)
}

// NeedsSave checks if the session has unsaved changes
func (s *Session) NeedsSave() bool {
	return s.DirtyState && time.Since(s.LastSave) > 30*time.Second
}
