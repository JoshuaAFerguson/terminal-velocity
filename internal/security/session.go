// File: internal/security/session.go
// Project: Terminal Velocity
// Description: Session management with timeout, tracking, and concurrent session limiting
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-14

package security

import (
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Security")

// SessionConfig holds session management configuration
type SessionConfig struct {
	IdleTimeout        time.Duration // Time before idle session expires (default: 15 minutes)
	MaxSessionDuration time.Duration // Maximum session length (default: 24 hours)
	MaxConcurrent      int           // Max concurrent sessions per player (default: 3)
	WarnBeforeKick     time.Duration // Warning time before kick (default: 2 minutes)
	CheckInterval      time.Duration // How often to check for expired sessions (default: 1 minute)
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		IdleTimeout:        15 * time.Minute,
		MaxSessionDuration: 24 * time.Hour,
		MaxConcurrent:      3,
		WarnBeforeKick:     2 * time.Minute,
		CheckInterval:      1 * time.Minute,
	}
}

// Session represents an active player session
type Session struct {
	ID           uuid.UUID
	PlayerID     uuid.UUID
	Username     string
	IPAddress    string
	StartTime    time.Time
	LastActivity time.Time
	UserAgent    string // Optional: track client type
	IsActive     bool
}

// SessionManager manages active sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session          // sessionID -> Session
	players  map[uuid.UUID][]uuid.UUID       // playerID -> []sessionIDs
	config   *SessionConfig
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionConfig) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	sm := &SessionManager{
		sessions: make(map[uuid.UUID]*Session),
		players:  make(map[uuid.UUID][]uuid.UUID),
		config:   config,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	sm.wg.Add(1)
	go sm.cleanupLoop()

	log.Info("Session manager initialized: idleTimeout=%v, maxDuration=%v, maxConcurrent=%d",
		config.IdleTimeout, config.MaxSessionDuration, config.MaxConcurrent)

	return sm
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(playerID uuid.UUID, username, ipAddress string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check concurrent session limit
	playerSessions := sm.players[playerID]
	if len(playerSessions) >= sm.config.MaxConcurrent {
		log.Warn("Max concurrent sessions reached for player: %s (limit: %d)", username, sm.config.MaxConcurrent)
		return nil, &SessionError{
			Code:    ErrMaxSessionsReached,
			Message: "Maximum concurrent sessions reached. Please close an existing session.",
		}
	}

	// Create new session
	session := &Session{
		ID:           uuid.New(),
		PlayerID:     playerID,
		Username:     username,
		IPAddress:    ipAddress,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// Store session
	sm.sessions[session.ID] = session
	sm.players[playerID] = append(sm.players[playerID], session.ID)

	log.Info("Session created: sessionID=%s, player=%s, ip=%s, concurrent=%d",
		session.ID, username, ipAddress, len(playerSessions)+1)

	return session, nil
}

// UpdateActivity updates the last activity time for a session
func (sm *SessionManager) UpdateActivity(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return &SessionError{Code: ErrSessionNotFound, Message: "Session not found"}
	}

	session.LastActivity = time.Now()
	return nil
}

// GetSession returns a session by ID
func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, &SessionError{Code: ErrSessionNotFound, Message: "Session not found"}
	}

	return session, nil
}

// GetPlayerSessions returns all sessions for a player
func (sm *SessionManager) GetPlayerSessions(playerID uuid.UUID) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionIDs := sm.players[playerID]
	sessions := make([]*Session, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		if session, ok := sm.sessions[sessionID]; ok {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetActiveSessions returns the count of active sessions for a player
func (sm *SessionManager) GetActiveSessions(playerID uuid.UUID) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.players[playerID])
}

// DestroySession destroys a session
func (sm *SessionManager) DestroySession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return &SessionError{Code: ErrSessionNotFound, Message: "Session not found"}
	}

	// Remove from sessions map
	delete(sm.sessions, sessionID)

	// Remove from player sessions
	playerSessions := sm.players[session.PlayerID]
	for i, sid := range playerSessions {
		if sid == sessionID {
			sm.players[session.PlayerID] = append(playerSessions[:i], playerSessions[i+1:]...)
			break
		}
	}

	// Clean up empty player entry
	if len(sm.players[session.PlayerID]) == 0 {
		delete(sm.players, session.PlayerID)
	}

	log.Info("Session destroyed: sessionID=%s, player=%s, duration=%v",
		sessionID, session.Username, time.Since(session.StartTime))

	return nil
}

// CheckSessionValid checks if a session is still valid
func (sm *SessionManager) CheckSessionValid(sessionID uuid.UUID) (valid bool, reason string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return false, "Session not found"
	}

	if !session.IsActive {
		return false, "Session inactive"
	}

	// Check idle timeout
	idleDuration := time.Since(session.LastActivity)
	if idleDuration > sm.config.IdleTimeout {
		return false, "Session idle timeout"
	}

	// Check max session duration
	sessionDuration := time.Since(session.StartTime)
	if sessionDuration > sm.config.MaxSessionDuration {
		return false, "Maximum session duration exceeded"
	}

	return true, ""
}

// ShouldWarnTimeout checks if session should be warned about timeout
func (sm *SessionManager) ShouldWarnTimeout(sessionID uuid.UUID) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return false
	}

	idleDuration := time.Since(session.LastActivity)
	timeUntilTimeout := sm.config.IdleTimeout - idleDuration

	return timeUntilTimeout > 0 && timeUntilTimeout <= sm.config.WarnBeforeKick
}

// GetTimeUntilTimeout returns time until session times out
func (sm *SessionManager) GetTimeUntilTimeout(sessionID uuid.UUID) time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return 0
	}

	idleDuration := time.Since(session.LastActivity)
	timeUntilTimeout := sm.config.IdleTimeout - idleDuration

	if timeUntilTimeout < 0 {
		return 0
	}

	return timeUntilTimeout
}

// cleanupLoop periodically checks for expired sessions
func (sm *SessionManager) cleanupLoop() {
	defer sm.wg.Done()
	ticker := time.NewTicker(sm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		case <-sm.stopChan:
			return
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var expired []uuid.UUID

	for sessionID, session := range sm.sessions {
		idleDuration := now.Sub(session.LastActivity)
		sessionDuration := now.Sub(session.StartTime)

		// Check if session expired
		if idleDuration > sm.config.IdleTimeout {
			log.Info("Session expired (idle): sessionID=%s, player=%s, idle=%v",
				sessionID, session.Username, idleDuration)
			expired = append(expired, sessionID)
		} else if sessionDuration > sm.config.MaxSessionDuration {
			log.Info("Session expired (max duration): sessionID=%s, player=%s, duration=%v",
				sessionID, session.Username, sessionDuration)
			expired = append(expired, sessionID)
		}
	}

	// Remove expired sessions
	for _, sessionID := range expired {
		session := sm.sessions[sessionID]

		// Remove from sessions map
		delete(sm.sessions, sessionID)

		// Remove from player sessions
		playerSessions := sm.players[session.PlayerID]
		for i, sid := range playerSessions {
			if sid == sessionID {
				sm.players[session.PlayerID] = append(playerSessions[:i], playerSessions[i+1:]...)
				break
			}
		}

		// Clean up empty player entry
		if len(sm.players[session.PlayerID]) == 0 {
			delete(sm.players, session.PlayerID)
		}
	}

	if len(expired) > 0 {
		log.Debug("Cleaned up %d expired sessions", len(expired))
	}
}

// GetStats returns session statistics
func (sm *SessionManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"total_sessions":  len(sm.sessions),
		"unique_players":  len(sm.players),
		"idle_timeout":    sm.config.IdleTimeout.String(),
		"max_duration":    sm.config.MaxSessionDuration.String(),
		"max_concurrent":  sm.config.MaxConcurrent,
	}
}

// Stop stops the session manager
func (sm *SessionManager) Stop() {
	close(sm.stopChan)
	sm.wg.Wait() // Wait for cleanup goroutine to finish
	log.Info("Session manager stopped")
}

// SessionError represents a session-related error
type SessionError struct {
	Code    int
	Message string
}

func (e *SessionError) Error() string {
	return e.Message
}

// Error codes
const (
	ErrSessionNotFound = iota + 1
	ErrMaxSessionsReached
	ErrSessionExpired
	ErrSessionInactive
)
