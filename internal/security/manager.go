// File: internal/security/manager.go
// Project: Terminal Velocity
// Description: Central security manager integrating all security features
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-14

package security

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Manager is the central security manager
type Manager struct {
	sessions  *SessionManager
	activity  *ActivityLogger
	honeypot  *HoneypotDetector
	anomaly   *AnomalyDetector
	config    *Config
}

// Config holds security manager configuration
type Config struct {
	// Session management
	SessionConfig *SessionConfig

	// Activity logging
	MaxActivityEvents int

	// Honeypot
	HoneypotAutoban         bool
	HoneypotAutobanDuration time.Duration

	// Anomaly detection
	RequireChallengeOnAnomaly bool
	AnomalyRiskThreshold      int // 0-100

	// General
	EnableAllFeatures bool
}

// DefaultConfig returns default security configuration
func DefaultConfig() *Config {
	return &Config{
		SessionConfig:             DefaultSessionConfig(),
		MaxActivityEvents:         10000,
		HoneypotAutoban:           true,
		HoneypotAutobanDuration:   24 * time.Hour,
		RequireChallengeOnAnomaly: true,
		AnomalyRiskThreshold:      50,
		EnableAllFeatures:         true,
	}
}

// NewManager creates a new security manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	manager := &Manager{
		sessions: NewSessionManager(config.SessionConfig),
		activity: NewActivityLogger(config.MaxActivityEvents),
		honeypot: NewHoneypotDetector(config.HoneypotAutoban, config.HoneypotAutobanDuration),
		anomaly:  NewAnomalyDetector(),
		config:   config,
	}

	log.Info("Security manager initialized with all features enabled")

	return manager
}

// ValidateLogin validates a login attempt with all security checks
func (m *Manager) ValidateLogin(username, password, ipAddress, userAgent string, playerID uuid.UUID) error {
	// Check honeypot
	if m.honeypot.IsHoneypot(username) {
		m.honeypot.RecordAttempt(username, ipAddress)
		m.activity.LogLoginFailure(username, ipAddress, "honeypot_account")

		// Check if should autoban
		if m.honeypot.ShouldAutoban(ipAddress) {
			log.Error("AUTOBAN triggered for honeypot attempt: ip=%s, username=%s", ipAddress, username)
			return fmt.Errorf("authentication failed")
		}

		return fmt.Errorf("authentication failed")
	}

	// If playerID is known, check for anomalies
	if playerID != uuid.Nil {
		anomalies := m.anomaly.DetectAnomalies(playerID, ipAddress, userAgent)
		if len(anomalies) > 0 {
			log.Warn("Login anomalies detected: player=%s, ip=%s, anomalies=%v",
				username, ipAddress, anomalies)

			// Log suspicious login
			m.activity.Log(playerID, username, ActivitySuspiciousLogin, ipAddress, true, map[string]interface{}{
				"anomalies": anomalies,
			})

			// Check risk score
			riskScore := m.anomaly.GetRiskScore(playerID, ipAddress, userAgent)
			if riskScore >= m.config.AnomalyRiskThreshold && m.config.RequireChallengeOnAnomaly {
				// In a full implementation, this would trigger 2FA or email verification
				log.Warn("High risk login detected (score: %d): player=%s, ip=%s", riskScore, username, ipAddress)
			}
		}
	}

	return nil
}

// OnLoginSuccess handles successful login
func (m *Manager) OnLoginSuccess(playerID uuid.UUID, username, ipAddress, userAgent string) (*Session, error) {
	// Check if this is a new IP
	isNewIP := m.activity.IsNewIP(playerID, ipAddress)

	// Log successful login
	m.activity.LogLoginSuccess(playerID, username, ipAddress, isNewIP)

	// Record in anomaly detector
	m.anomaly.RecordLogin(playerID, ipAddress, userAgent)

	// Create session
	session, err := m.sessions.CreateSession(playerID, username, ipAddress)
	if err != nil {
		return nil, err
	}

	log.Info("Login successful: player=%s, ip=%s, sessionID=%s, newIP=%v",
		username, ipAddress, session.ID, isNewIP)

	return session, nil
}

// OnLoginFailure handles failed login
func (m *Manager) OnLoginFailure(username, ipAddress, reason string) {
	m.activity.LogLoginFailure(username, ipAddress, reason)

	// Check if honeypot
	if m.honeypot.IsHoneypot(username) {
		m.honeypot.RecordAttempt(username, ipAddress)
	}
}

// OnLogout handles logout
func (m *Manager) OnLogout(sessionID uuid.UUID) error {
	session, err := m.sessions.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Log logout
	m.activity.Log(session.PlayerID, session.Username, ActivityLogout, session.IPAddress, true, nil)

	// Destroy session
	return m.sessions.DestroySession(sessionID)
}

// UpdateSessionActivity updates session activity timestamp
func (m *Manager) UpdateSessionActivity(sessionID uuid.UUID) error {
	return m.sessions.UpdateActivity(sessionID)
}

// GetSession returns a session by ID
func (m *Manager) GetSession(sessionID uuid.UUID) (*Session, error) {
	return m.sessions.GetSession(sessionID)
}

// GetPlayerSessions returns all sessions for a player
func (m *Manager) GetPlayerSessions(playerID uuid.UUID) []*Session {
	return m.sessions.GetPlayerSessions(playerID)
}

// GetActiveSessions returns count of active sessions for a player
func (m *Manager) GetActiveSessions(playerID uuid.UUID) int {
	return m.sessions.GetActiveSessions(playerID)
}

// CheckSessionValid checks if a session is still valid
func (m *Manager) CheckSessionValid(sessionID uuid.UUID) (valid bool, reason string) {
	return m.sessions.CheckSessionValid(sessionID)
}

// GetTimeUntilTimeout returns time until session times out
func (m *Manager) GetTimeUntilTimeout(sessionID uuid.UUID) time.Duration {
	return m.sessions.GetTimeUntilTimeout(sessionID)
}

// GetPlayerActivity returns recent activity for a player
func (m *Manager) GetPlayerActivity(playerID uuid.UUID, limit int) []*ActivityEvent {
	return m.activity.GetPlayerEvents(playerID, limit)
}

// GetHighRiskEvents returns high risk security events
func (m *Manager) GetHighRiskEvents(since time.Time) []*ActivityEvent {
	return m.activity.GetHighRiskEvents(since)
}

// IsHoneypot checks if a username is a honeypot
func (m *Manager) IsHoneypot(username string) bool {
	return m.honeypot.IsHoneypot(username)
}

// GetStats returns comprehensive security statistics
func (m *Manager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"sessions":  m.sessions.GetStats(),
		"activity":  m.activity.GetStats(),
		"honeypot":  m.honeypot.GetStats(),
		"anomaly":   m.anomaly.GetStats(),
	}
}

// Stop stops all security subsystems
func (m *Manager) Stop() {
	m.sessions.Stop()
	log.Info("Security manager stopped")
}
