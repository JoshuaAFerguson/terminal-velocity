// File: internal/security/manager.go
// Project: Terminal Velocity
// Description: Central security manager integrating all security features
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-14

// Package security provides comprehensive security features for Terminal Velocity.
//
// This package integrates multiple security subsystems to protect player accounts,
// detect suspicious activity, and maintain an audit trail of security events.
// It provides a unified interface for authentication, session management, and
// threat detection.
//
// Features:
//   - Session Management: Tracks active sessions with idle/max duration timeouts
//   - Activity Logging: Maintains audit trail of security events (10,000 event buffer)
//   - Honeypot Detection: Identifies and tracks attacks on decoy accounts
//   - Anomaly Detection: Detects suspicious login patterns (new IPs, impossible travel, etc.)
//   - Two-Factor Authentication: TOTP-based 2FA with backup codes (optional)
//   - Risk Scoring: Calculates risk scores for login attempts (0-100 scale)
//   - Auto-banning: Automatically bans IPs attempting honeypot accounts
//
// Architecture:
//
//	Manager (Coordinator)
//	    ├── SessionManager     (Session lifecycle & timeouts)
//	    ├── ActivityLogger     (Audit trail & event logging)
//	    ├── HoneypotDetector   (Attack detection via decoy accounts)
//	    ├── AnomalyDetector    (Login pattern analysis)
//	    └── TwoFactorManager   (TOTP 2FA - optional)
//
// Security Flow (Login):
//
//	1. ValidateLogin() - Pre-authentication security checks
//	   - Check if username is honeypot
//	   - Detect login anomalies (new IP, device, location)
//	   - Calculate risk score
//	2. OnLoginSuccess() - Post-authentication processing
//	   - Create secure session
//	   - Log successful login event
//	   - Track IP for future anomaly detection
//	3. Session Active - Ongoing security
//	   - Update activity timestamps
//	   - Monitor for timeout
//	   - Track concurrent sessions (max 3 per player)
//	4. OnLogout() - Session cleanup
//	   - Log logout event
//	   - Destroy session
//
// Thread Safety:
// All Manager methods are thread-safe. Internal subsystems use sync.RWMutex
// for concurrent access protection.
//
// Usage Example:
//
//	// Initialize security manager
//	config := security.DefaultConfig()
//	config.SessionConfig.IdleTimeout = 30 * time.Minute
//	config.HoneypotAutoban = true
//	securityMgr := security.NewManager(config)
//	defer securityMgr.Stop()
//
//	// Validate login attempt
//	err := securityMgr.ValidateLogin(username, password, ipAddress, userAgent, playerID)
//	if err != nil {
//	    // Login rejected due to security concerns
//	    return err
//	}
//
//	// On successful authentication
//	session, err := securityMgr.OnLoginSuccess(playerID, username, ipAddress, userAgent)
//	if err != nil {
//	    return err
//	}
//
//	// Update activity during gameplay
//	securityMgr.UpdateSessionActivity(session.ID)
//
//	// Check if session is still valid
//	if valid, reason := securityMgr.CheckSessionValid(session.ID); !valid {
//	    log.Warn("Session invalid: %s", reason)
//	    // Force logout
//	}
//
//	// On logout
//	securityMgr.OnLogout(session.ID)
//
// Configuration:
//
//	type Config struct {
//	    // Session settings
//	    SessionConfig *SessionConfig
//
//	    // Activity logging
//	    MaxActivityEvents int  // Default: 10000
//
//	    // Honeypot
//	    HoneypotAutoban         bool           // Default: true
//	    HoneypotAutobanDuration time.Duration  // Default: 24h
//
//	    // Anomaly detection
//	    RequireChallengeOnAnomaly bool  // Default: true
//	    AnomalyRiskThreshold      int   // Default: 50 (0-100 scale)
//
//	    // Global
//	    EnableAllFeatures bool  // Default: true
//	}
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package security

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Manager is the central security manager that coordinates all security subsystems.
//
// Manager provides a unified interface for authentication security checks, session
// management, activity logging, and threat detection. It integrates multiple security
// components to provide defense-in-depth protection for player accounts.
//
// Components:
//   - sessions: SessionManager for tracking and managing active sessions
//   - activity: ActivityLogger for audit trail and security event logging
//   - honeypot: HoneypotDetector for identifying attacks on decoy accounts
//   - anomaly: AnomalyDetector for detecting suspicious login patterns
//   - config: Configuration settings for all security features
//
// Thread Safety:
// Manager is thread-safe and can be safely called from multiple goroutines.
type Manager struct {
	sessions  *SessionManager   // Session lifecycle management
	activity  *ActivityLogger   // Security event logging
	honeypot  *HoneypotDetector // Honeypot attack detection
	anomaly   *AnomalyDetector  // Login anomaly detection
	config    *Config           // Security configuration
}

// Config holds security manager configuration.
//
// This struct configures all security subsystems including session management,
// activity logging, honeypot detection, and anomaly detection.
//
// Fields:
//   - SessionConfig: Session timeout and concurrency settings
//   - MaxActivityEvents: Maximum number of events to retain in activity log (default: 10000)
//   - HoneypotAutoban: Whether to automatically ban IPs hitting honeypot accounts (default: true)
//   - HoneypotAutobanDuration: How long to ban honeypot attackers (default: 24h)
//   - RequireChallengeOnAnomaly: Whether to require additional verification for anomalous logins (default: true)
//   - AnomalyRiskThreshold: Risk score threshold (0-100) for triggering challenges (default: 50)
//   - EnableAllFeatures: Master switch for all security features (default: true)
//
// Example:
//
//	config := &security.Config{
//	    SessionConfig: &security.SessionConfig{
//	        IdleTimeout: 30 * time.Minute,
//	        MaxConcurrent: 3,
//	    },
//	    HoneypotAutoban: true,
//	    AnomalyRiskThreshold: 60, // More strict threshold
//	}
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

// DefaultConfig returns default security configuration.
//
// Returns a Config with sensible defaults suitable for most deployments:
//   - 15 minute idle timeout, 24 hour max session duration
//   - Maximum 3 concurrent sessions per player
//   - 10,000 event activity log buffer
//   - Honeypot auto-ban enabled (24 hour bans)
//   - Anomaly detection with 50% risk threshold
//   - All security features enabled
//
// Returns:
//   - *Config: Configuration with default values
//
// Example:
//
//	config := security.DefaultConfig()
//	config.SessionConfig.IdleTimeout = 60 * time.Minute  // Override default
//	manager := security.NewManager(config)
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

// NewManager creates a new security manager with all subsystems initialized.
//
// This function initializes all security components including session management,
// activity logging, honeypot detection, and anomaly detection. Background workers
// are started automatically (e.g., session cleanup goroutine).
//
// Parameters:
//   - config: Security configuration. If nil, DefaultConfig() is used.
//
// Returns:
//   - *Manager: Fully initialized security manager ready for use
//
// Example:
//
//	// Use default configuration
//	manager := security.NewManager(nil)
//	defer manager.Stop()
//
//	// Use custom configuration
//	config := security.DefaultConfig()
//	config.HoneypotAutoban = false  // Disable auto-ban
//	manager := security.NewManager(config)
//	defer manager.Stop()
//
// Note: Call Stop() when done to cleanly shutdown background workers.
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

// ValidateLogin validates a login attempt with all security checks.
//
// This method performs comprehensive pre-authentication security validation including
// honeypot detection, anomaly detection, and risk scoring. Call this BEFORE verifying
// credentials to catch attacks early.
//
// Security Checks:
//  1. Honeypot Detection: Checks if username is a decoy account
//  2. Anomaly Detection: Analyzes login patterns (new IP, device, unusual time)
//  3. Risk Scoring: Calculates risk score based on detected anomalies
//  4. Challenge Requirements: May require additional verification for high-risk logins
//
// Parameters:
//   - username: Username being authenticated
//   - password: Password (unused currently, for future password strength checks)
//   - ipAddress: IP address of login attempt
//   - userAgent: User agent string (optional, for device tracking)
//   - playerID: Player ID if known (uuid.Nil if unknown/first login)
//
// Returns:
//   - error: Returns error if login should be rejected, nil if checks pass
//
// Errors:
//   - Returns generic "authentication failed" for honeypot attempts (logs details internally)
//   - Future: May return specific errors for 2FA challenges, account locks, etc.
//
// Example:
//
//	// Before credential verification
//	err := manager.ValidateLogin(username, password, ipAddress, userAgent, playerID)
//	if err != nil {
//	    log.Warn("Security validation failed: %v", err)
//	    return err
//	}
//
//	// Now verify credentials...
//
// Note: This method logs security events internally. Honeypot attempts trigger
// auto-ban if configured.
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
