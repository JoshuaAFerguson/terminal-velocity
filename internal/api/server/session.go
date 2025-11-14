// File: internal/api/server/session.go
// Project: Terminal Velocity
// Description: Session management for game server
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/google/uuid"
)

// SessionManager manages active game sessions
type SessionManager struct {
	sessions map[uuid.UUID]*api.Session
	mu       sync.RWMutex

	// Configuration
	sessionTTL time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:   make(map[uuid.UUID]*api.Session),
		sessionTTL: 24 * time.Hour, // Default 24 hour sessions
	}
}

// CreateSession creates a new game session
func (m *SessionManager) CreateSession(playerID uuid.UUID) (*api.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	session := &api.Session{
		SessionID:    uuid.New(),
		PlayerID:     playerID,
		CreatedAt:    now,
		LastActivity: now,
		ExpiresAt:    now.Add(m.sessionTTL),
		State:        api.SessionStateActive,
	}

	m.sessions[session.SessionID] = session

	return session, nil
}

// GetSession retrieves a session by ID
func (m *SessionManager) GetSession(sessionID uuid.UUID) (*api.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, api.ErrNotFound
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		return nil, api.ErrUnauthorized
	}

	return session, nil
}

// UpdateActivity updates the last activity time for a session
func (m *SessionManager) UpdateActivity(sessionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return api.ErrNotFound
	}

	session.LastActivity = time.Now()
	return nil
}

// RefreshSession extends a session's expiration time
func (m *SessionManager) RefreshSession(sessionID uuid.UUID) (*api.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, api.ErrNotFound
	}

	// Check if session is already expired
	if time.Now().After(session.ExpiresAt) {
		return nil, api.ErrUnauthorized
	}

	// Extend expiration by session TTL from now
	now := time.Now()
	session.LastActivity = now
	session.ExpiresAt = now.Add(m.sessionTTL)

	return session, nil
}

// EndSession terminates a session
func (m *SessionManager) EndSession(sessionID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return api.ErrNotFound
	}

	session.State = api.SessionStateTerminated
	delete(m.sessions, sessionID)

	return nil
}

// CleanupExpiredSessions removes expired sessions
// This should be called periodically (e.g., every 5 minutes)
func (m *SessionManager) CleanupExpiredSessions() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for sessionID, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			session.State = api.SessionStateExpired
			delete(m.sessions, sessionID)
			cleaned++
		}
	}

	return cleaned
}

// GetActiveSessions returns the number of active sessions
func (m *SessionManager) GetActiveSessions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessions)
}

// GetSessionsByPlayer returns all sessions for a specific player
func (m *SessionManager) GetSessionsByPlayer(playerID uuid.UUID) []*api.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*api.Session
	for _, session := range m.sessions {
		if session.PlayerID == playerID {
			sessions = append(sessions, session)
		}
	}

	return sessions
}
