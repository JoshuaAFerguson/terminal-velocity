// File: internal/security/activity.go
// Project: Terminal Velocity
// Description: Account activity logging for security events and audit trail
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-14

// Package security - Activity logging for security audit trails and forensics.
//
// This file implements comprehensive logging of security-related events including
// authentication attempts, account changes, suspicious activity, and admin actions.
// It maintains an in-memory circular buffer for fast access and provides export
// capabilities for long-term archival.
//
// Features:
//   - Comprehensive Event Logging: 20+ event types tracked
//   - Risk-Level Classification: Automatic risk assessment (none/low/medium/high/critical)
//   - Circular Buffer: 10,000 event in-memory cache (configurable)
//   - Player History: Quick access to per-player event history
//   - IP Tracking: Tracks known IPs per player for new-IP detection
//   - High-Risk Filtering: Query events by risk level and timeframe
//   - JSON Export: Export events for archival or external analysis
//   - Automatic Logging: Events automatically logged to system logger by risk level
//
// Event Types (20+):
//   Authentication:
//   - login_success, login_failure, logout, session_expired, session_kicked
//
//   Account Management:
//   - password_changed, email_changed, 2fa_enabled, 2fa_disabled
//
//   Security Events:
//   - suspicious_login, new_ip_login, new_location_login
//   - account_locked, account_unlocked
//   - password_reset_requested, password_reset_completed
//
//   Admin Actions:
//   - admin_action, permission_changed
//
//   Game Security (optional):
//   - rapid_actions, anomalous_trading
//
// Risk Levels:
//   - none: Normal activity (login_success, logout)
//   - low: Minor events (session_expired, email_changed)
//   - medium: Moderate concern (login_failure, password_changed, 2fa_disabled)
//   - high: Significant concern (new_ip_login, password_reset_requested)
//   - critical: Immediate attention (account_locked, suspicious_login)
//
// Event Structure:
//
//	type ActivityEvent struct {
//	    ID          uuid.UUID              // Unique event ID
//	    PlayerID    uuid.UUID              // Player involved (Nil for unknown)
//	    Username    string                 // Username attempted/used
//	    EventType   ActivityType           // Type of event
//	    IPAddress   string                 // Source IP address
//	    UserAgent   string                 // User agent (optional)
//	    Timestamp   time.Time              // When event occurred
//	    Success     bool                   // Whether action succeeded
//	    Details     map[string]interface{} // Additional context
//	    RiskLevel   RiskLevel              // Auto-calculated risk
//	}
//
// Automatic Features:
//   - Risk auto-calculation based on event type
//   - Automatic logging to system logger (level based on risk)
//   - IP tracking for new-IP detection
//   - Circular buffer management (auto-trim oldest events)
//
// Thread Safety:
// All ActivityLogger methods are thread-safe using sync.RWMutex.
//
// Usage Example:
//
//	// Initialize activity logger
//	logger := security.NewActivityLogger(10000) // 10k event buffer
//
//	// Log successful login
//	logger.LogLoginSuccess(playerID, username, ipAddress, isNewIP)
//
//	// Log failed login
//	logger.LogLoginFailure(username, ipAddress, "invalid_password")
//
//	// Log custom event
//	logger.Log(playerID, username, security.ActivityPasswordChanged, ipAddress, true, nil)
//
//	// Log event with details
//	logger.Log(playerID, username, security.ActivitySuspiciousLogin, ipAddress, true,
//	    map[string]interface{}{
//	        "anomalies": []string{"new_ip", "unusual_time"},
//	        "risk_score": 65,
//	    })
//
//	// Query player history
//	events := logger.GetPlayerEvents(playerID, 50) // Last 50 events
//	for _, event := range events {
//	    fmt.Printf("%s: %s from %s (risk: %s)\n",
//	        event.Timestamp, event.EventType, event.IPAddress, event.RiskLevel)
//	}
//
//	// Get high-risk events from last 24 hours
//	since := time.Now().Add(-24 * time.Hour)
//	highRisk := logger.GetHighRiskEvents(since)
//	if len(highRisk) > 0 {
//	    log.Warn("Found %d high-risk events in last 24 hours", len(highRisk))
//	}
//
//	// Export to JSON for archival
//	recentEvents := logger.GetRecentEvents(1000)
//	jsonData, err := logger.ExportJSON(recentEvents)
//	if err == nil {
//	    saveToFile("security_audit.json", jsonData)
//	}
//
//	// Get statistics
//	stats := logger.GetStats()
//	fmt.Printf("Total events: %d\n", stats["total_events"])
//	fmt.Printf("Tracked players: %d\n", stats["tracked_players"])
//
// Integration with Main Security Manager:
// Manager uses ActivityLogger to log all security events automatically.
// Events are logged at appropriate system log levels based on risk.
//
// Best Practices:
//   - Export events daily for long-term archival
//   - Monitor high-risk events regularly
//   - Correlate events across players to detect attack patterns
//   - Use details field for event-specific context
//   - Set buffer size based on activity level (1000-50000 events)
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package security

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ActivityType represents types of account activities
type ActivityType string

const (
	// Authentication events
	ActivityLoginSuccess    ActivityType = "login_success"
	ActivityLoginFailure    ActivityType = "login_failure"
	ActivityLogout          ActivityType = "logout"
	ActivitySessionExpired  ActivityType = "session_expired"
	ActivitySessionKicked   ActivityType = "session_kicked"

	// Account management events
	ActivityPasswordChanged   ActivityType = "password_changed"
	ActivityEmailChanged      ActivityType = "email_changed"
	ActivityTwoFactorEnabled  ActivityType = "2fa_enabled"
	ActivityTwoFactorDisabled ActivityType = "2fa_disabled"

	// Security events
	ActivitySuspiciousLogin    ActivityType = "suspicious_login"
	ActivityNewIPLogin         ActivityType = "new_ip_login"
	ActivityNewLocationLogin   ActivityType = "new_location_login"
	ActivityAccountLocked      ActivityType = "account_locked"
	ActivityAccountUnlocked    ActivityType = "account_unlocked"
	ActivityPasswordResetRequested ActivityType = "password_reset_requested"
	ActivityPasswordResetCompleted ActivityType = "password_reset_completed"

	// Admin actions
	ActivityAdminAction       ActivityType = "admin_action"
	ActivityPermissionChanged ActivityType = "permission_changed"

	// Game events (optional - for suspicious activity)
	ActivityRapidActions      ActivityType = "rapid_actions"
	ActivityAnomalousTrading  ActivityType = "anomalous_trading"
)

// Activity Event represents a logged security event
type ActivityEvent struct {
	ID          uuid.UUID                `json:"id"`
	PlayerID    uuid.UUID                `json:"player_id"`
	Username    string                   `json:"username"`
	EventType   ActivityType             `json:"event_type"`
	IPAddress   string                   `json:"ip_address"`
	UserAgent   string                   `json:"user_agent,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
	Success     bool                     `json:"success"`
	Details     map[string]interface{}   `json:"details,omitempty"`
	RiskLevel   RiskLevel                `json:"risk_level"`
}

// RiskLevel represents the risk level of an activity
type RiskLevel string

const (
	RiskNone     RiskLevel = "none"      // Normal activity
	RiskLow      RiskLevel = "low"       // Slightly unusual
	RiskMedium   RiskLevel = "medium"    // Moderately suspicious
	RiskHigh     RiskLevel = "high"      // Very suspicious
	RiskCritical RiskLevel = "critical"  // Immediate attention needed
)

// ActivityLogger logs account activities
type ActivityLogger struct {
	mu         sync.RWMutex
	events     []*ActivityEvent
	maxEvents  int
	playerIPs  map[uuid.UUID]map[string]time.Time // playerID -> IP -> last seen
}

// NewActivityLogger creates a new activity logger
func NewActivityLogger(maxEvents int) *ActivityLogger {
	if maxEvents <= 0 {
		maxEvents = 10000
	}

	return &ActivityLogger{
		events:    make([]*ActivityEvent, 0, maxEvents),
		maxEvents: maxEvents,
		playerIPs: make(map[uuid.UUID]map[string]time.Time),
	}
}

// LogEvent logs an activity event
func (al *ActivityLogger) LogEvent(event *ActivityEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Set ID and timestamp if not already set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Auto-detect risk level if not set
	if event.RiskLevel == "" {
		event.RiskLevel = al.calculateRiskLevel(event)
	}

	// Add event
	al.events = append(al.events, event)

	// Trim if too large
	if len(al.events) > al.maxEvents {
		// Remove oldest 1000 events
		al.events = al.events[1000:]
	}

	// Track IP addresses for anomaly detection
	if event.EventType == ActivityLoginSuccess {
		al.trackIP(event.PlayerID, event.IPAddress)
	}

	// Log to system log based on risk level
	al.logToSystem(event)
}

// Log logs a simple activity event
func (al *ActivityLogger) Log(playerID uuid.UUID, username string, eventType ActivityType, ipAddress string, success bool, details map[string]interface{}) {
	event := &ActivityEvent{
		ID:        uuid.UUID(uuid.New()),
		PlayerID:  playerID,
		Username:  username,
		EventType: eventType,
		IPAddress: ipAddress,
		Timestamp: time.Now(),
		Success:   success,
		Details:   details,
	}

	al.LogEvent(event)
}

// LogLoginSuccess logs a successful login
func (al *ActivityLogger) LogLoginSuccess(playerID uuid.UUID, username, ipAddress string, isNewIP bool) {
	details := map[string]interface{}{
		"new_ip": isNewIP,
	}

	eventType := ActivityLoginSuccess
	if isNewIP {
		eventType = ActivityNewIPLogin
	}

	al.Log(playerID, username, eventType, ipAddress, true, details)
}

// LogLoginFailure logs a failed login attempt
func (al *ActivityLogger) LogLoginFailure(username, ipAddress, reason string) {
	details := map[string]interface{}{
		"reason": reason,
	}

	al.Log(uuid.Nil, username, ActivityLoginFailure, ipAddress, false, details)
}

// IsNewIP checks if an IP is new for a player
func (al *ActivityLogger) IsNewIP(playerID uuid.UUID, ipAddress string) bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	playerIPMap, exists := al.playerIPs[playerID]
	if !exists {
		return true
	}

	_, seen := playerIPMap[ipAddress]
	return !seen
}

// trackIP tracks an IP address for a player
func (al *ActivityLogger) trackIP(playerID uuid.UUID, ipAddress string) {
	if _, exists := al.playerIPs[playerID]; !exists {
		al.playerIPs[playerID] = make(map[string]time.Time)
	}

	al.playerIPs[playerID][ipAddress] = time.Now()
}

// GetPlayerEvents returns recent events for a player
func (al *ActivityLogger) GetPlayerEvents(playerID uuid.UUID, limit int) []*ActivityEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var playerEvents []*ActivityEvent

	// Iterate in reverse to get most recent first
	for i := len(al.events) - 1; i >= 0 && len(playerEvents) < limit; i-- {
		if al.events[i].PlayerID == playerID {
			playerEvents = append(playerEvents, al.events[i])
		}
	}

	return playerEvents
}

// GetRecentEvents returns the most recent events across all players
func (al *ActivityLogger) GetRecentEvents(limit int) []*ActivityEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	start := len(al.events) - limit
	if start < 0 {
		start = 0
	}

	// Return copy
	result := make([]*ActivityEvent, len(al.events[start:]))
	copy(result, al.events[start:])

	return result
}

// GetHighRiskEvents returns events with high or critical risk level
func (al *ActivityLogger) GetHighRiskEvents(since time.Time) []*ActivityEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var highRisk []*ActivityEvent

	for _, event := range al.events {
		if event.Timestamp.After(since) &&
			(event.RiskLevel == RiskHigh || event.RiskLevel == RiskCritical) {
			highRisk = append(highRisk, event)
		}
	}

	return highRisk
}

// calculateRiskLevel automatically determines risk level for an event
func (al *ActivityLogger) calculateRiskLevel(event *ActivityEvent) RiskLevel {
	switch event.EventType {
	// Critical risk events
	case ActivityAccountLocked, ActivitySuspiciousLogin:
		return RiskCritical

	// High risk events
	case ActivityNewIPLogin, ActivityNewLocationLogin, ActivityPasswordResetRequested:
		return RiskHigh

	// Medium risk events
	case ActivityLoginFailure, ActivityPasswordChanged, ActivityTwoFactorDisabled:
		return RiskMedium

	// Low risk events
	case ActivitySessionExpired, ActivityEmailChanged:
		return RiskLow

	// Normal events
	case ActivityLoginSuccess, ActivityLogout:
		return RiskNone

	default:
		return RiskNone
	}
}

// logToSystem logs event to system logger based on risk
func (al *ActivityLogger) logToSystem(event *ActivityEvent) {
	switch event.RiskLevel {
	case RiskCritical:
		log.Error("CRITICAL SECURITY EVENT: type=%s, player=%s, ip=%s, details=%v",
			event.EventType, event.Username, event.IPAddress, event.Details)

	case RiskHigh:
		log.Warn("HIGH RISK EVENT: type=%s, player=%s, ip=%s, details=%v",
			event.EventType, event.Username, event.IPAddress, event.Details)

	case RiskMedium:
		log.Warn("MEDIUM RISK EVENT: type=%s, player=%s, ip=%s",
			event.EventType, event.Username, event.IPAddress)

	case RiskLow:
		log.Info("Security event: type=%s, player=%s, ip=%s",
			event.EventType, event.Username, event.IPAddress)

	case RiskNone:
		log.Debug("Activity: type=%s, player=%s", event.EventType, event.Username)
	}
}

// ExportJSON exports events to JSON format
func (al *ActivityLogger) ExportJSON(events []*ActivityEvent) ([]byte, error) {
	return json.MarshalIndent(events, "", "  ")
}

// GetStats returns statistics about logged activities
func (al *ActivityLogger) GetStats() map[string]interface{} {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Count by type
	typeCount := make(map[ActivityType]int)
	riskCount := make(map[RiskLevel]int)

	for _, event := range al.events {
		typeCount[event.EventType]++
		riskCount[event.RiskLevel]++
	}

	return map[string]interface{}{
		"total_events":    len(al.events),
		"tracked_players": len(al.playerIPs),
		"by_type":         typeCount,
		"by_risk":         riskCount,
	}
}
