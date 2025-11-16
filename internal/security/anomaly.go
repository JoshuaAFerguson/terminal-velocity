// File: internal/security/anomaly.go
// Project: Terminal Velocity
// Description: Login anomaly detection for suspicious activity
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-14

// Package security - Anomaly detection for suspicious login patterns.
//
// This file implements behavioral analysis to detect account compromise, credential
// theft, and unusual access patterns. It learns normal login behavior for each player
// and flags deviations as potential security risks.
//
// Features:
//   - Behavioral Profiling: Tracks normal login patterns per player
//   - Multi-Factor Analysis: Detects anomalies across IP, device, time, location
//   - Risk Scoring: Calculates 0-100 risk score based on detected anomalies
//   - Pattern Learning: Builds profile from last 100 logins per player
//   - Challenge Triggers: Optionally require additional verification for high-risk logins
//
// Detected Anomalies:
//   - new_ip: Login from never-before-seen IP address
//   - new_device: Login from different user agent (new device/browser)
//   - unusual_time: Login at atypical hour compared to normal pattern
//   - rapid_ip_change: IP change within 1 hour (possible session hijack)
//   - impossible_travel: Geo-location change too fast to be legitimate (future)
//
// Risk Scoring:
//   Each anomaly contributes to risk score:
//   - new_ip: +20 points
//   - new_device: +15 points
//   - unusual_time: +10 points
//   - rapid_ip_change: +30 points
//   - impossible_travel: +50 points (future with GeoIP)
//
//   Score Thresholds:
//   - 0-25: Low risk (normal behavior)
//   - 26-50: Medium risk (mildly suspicious)
//   - 51-75: High risk (requires attention)
//   - 76-100: Critical risk (likely compromise)
//
// How It Works:
//
//	1. Track Logins:
//	   - RecordLogin() stores IP, user agent, timestamp
//	   - Maintains last 100 logins per player
//	   - Builds typical behavior profile
//
//	2. Detect Anomalies:
//	   - DetectAnomalies() compares current login to history
//	   - Returns list of detected anomaly types
//	   - Checks: IP, device, time patterns, geo-location
//
//	3. Calculate Risk:
//	   - GetRiskScore() sums anomaly scores
//	   - Returns 0-100 risk value
//	   - Used to trigger challenges or alerts
//
//	4. Take Action:
//	   - Low/Medium: Allow but log
//	   - High: Require email/2FA verification
//	   - Critical: Block and alert admin
//
// Limitations (Current Implementation):
//   - No GeoIP database (impossible travel not implemented)
//   - Simple time-of-day analysis (hour-based)
//   - No VPN/Proxy detection
//   - No device fingerprinting beyond user agent
//
// Future Enhancements:
//   - Integrate MaxMind GeoIP for location tracking
//   - Velocity-based impossible travel detection
//   - Machine learning for better pattern recognition
//   - VPN/Tor exit node detection
//   - Browser fingerprinting
//
// Thread Safety:
// All AnomalyDetector methods are thread-safe using sync.RWMutex.
//
// Usage Example:
//
//	// Initialize anomaly detector
//	detector := security.NewAnomalyDetector()
//
//	// On successful login, record for pattern learning
//	detector.RecordLogin(playerID, ipAddress, userAgent)
//
//	// On login attempt, detect anomalies
//	anomalies := detector.DetectAnomalies(playerID, ipAddress, userAgent)
//	if len(anomalies) > 0 {
//	    log.Warn("Login anomalies detected: %v", anomalies)
//
//	    // Calculate risk score
//	    riskScore := detector.GetRiskScore(playerID, ipAddress, userAgent)
//	    log.Info("Risk score: %d", riskScore)
//
//	    // Take action based on risk
//	    if riskScore >= 50 {
//	        // High risk - require 2FA
//	        log.Warn("High-risk login detected, requiring 2FA")
//	        return errors.New("additional verification required")
//	    } else if riskScore >= 25 {
//	        // Medium risk - log but allow
//	        log.Info("Medium-risk login, logging event")
//	        logSecurityEvent("suspicious_login", playerID, anomalies)
//	    }
//	}
//
//	// Determine if should challenge
//	if detector.ShouldChallenge(playerID, ipAddress, userAgent) {
//	    // Trigger 2FA or email verification
//	    sendVerificationChallenge(playerID)
//	}
//
// Integration with Main Security Manager:
// Manager.ValidateLogin() automatically calls DetectAnomalies() and logs suspicious
// logins. High-risk logins can trigger additional verification challenges.
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package security

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// LoginHistory tracks login patterns for anomaly detection
type LoginHistory struct {
	PlayerID      uuid.UUID
	Timestamps    []time.Time
	IPAddresses   map[string]time.Time // IP -> last seen
	UserAgents    map[string]time.Time // User agent -> last seen
	TypicalHours  map[int]int          // Hour of day -> count
	Countries     map[string]time.Time // Country code -> last seen (future: with GeoIP)
}

// AnomalyDetector detects suspicious login patterns
type AnomalyDetector struct {
	mu       sync.RWMutex
	history  map[uuid.UUID]*LoginHistory // playerID -> history
	maxHistory int
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		history:    make(map[uuid.UUID]*LoginHistory),
		maxHistory: 100, // Keep last 100 logins per player
	}
}

// RecordLogin records a login for pattern analysis
func (ad *AnomalyDetector) RecordLogin(playerID uuid.UUID, ipAddress, userAgent string) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	history, exists := ad.history[playerID]
	if !exists {
		history = &LoginHistory{
			PlayerID:     playerID,
			Timestamps:   make([]time.Time, 0),
			IPAddresses:  make(map[string]time.Time),
			UserAgents:   make(map[string]time.Time),
			TypicalHours: make(map[int]int),
			Countries:    make(map[string]time.Time),
		}
		ad.history[playerID] = history
	}

	now := time.Now()

	// Record timestamp
	history.Timestamps = append(history.Timestamps, now)

	// Trim if too long
	if len(history.Timestamps) > ad.maxHistory {
		history.Timestamps = history.Timestamps[len(history.Timestamps)-ad.maxHistory:]
	}

	// Record IP
	history.IPAddresses[ipAddress] = now

	// Record user agent
	if userAgent != "" {
		history.UserAgents[userAgent] = now
	}

	// Record typical hour
	hour := now.Hour()
	history.TypicalHours[hour]++
}

// DetectAnomalies detects suspicious login patterns
func (ad *AnomalyDetector) DetectAnomalies(playerID uuid.UUID, ipAddress, userAgent string) []string {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	history, exists := ad.history[playerID]
	if !exists {
		// First login - no anomalies
		return nil
	}

	var anomalies []string

	// Check for new IP
	if _, seen := history.IPAddresses[ipAddress]; !seen {
		anomalies = append(anomalies, "new_ip")
	}

	// Check for new user agent
	if userAgent != "" {
		if _, seen := history.UserAgents[userAgent]; !seen {
			anomalies = append(anomalies, "new_device")
		}
	}

	// Check for unusual time
	now := time.Now()
	hour := now.Hour()
	avgLoginCount := 0
	for _, count := range history.TypicalHours {
		avgLoginCount += count
	}
	if len(history.TypicalHours) > 0 {
		avgLoginCount /= len(history.TypicalHours)
	}

	// If this hour has very few logins compared to average, it's unusual
	if history.TypicalHours[hour] < avgLoginCount/2 {
		anomalies = append(anomalies, "unusual_time")
	}

	// Check for rapid geo-location change (if last login was recent from different IP)
	// This is a simplified version - would need GeoIP database for real implementation
	if len(history.Timestamps) > 0 {
		lastLogin := history.Timestamps[len(history.Timestamps)-1]
		timeSinceLastLogin := now.Sub(lastLogin)

		// If last login was less than 1 hour ago from different IP
		// Could indicate account sharing or compromise
		if timeSinceLastLogin < 1*time.Hour {
			lastIP := ad.getLastIP(history)
			if lastIP != "" && lastIP != ipAddress {
				anomalies = append(anomalies, "rapid_ip_change")
			}
		}
	}

	// Check for impossible travel (would need GeoIP)
	// Future: if IP locations are very far apart in short time

	return anomalies
}

// getLastIP gets the last IP used (simple implementation)
func (ad *AnomalyDetector) getLastIP(history *LoginHistory) string {
	var lastIP string
	var lastTime time.Time

	for ip, t := range history.IPAddresses {
		if t.After(lastTime) {
			lastTime = t
			lastIP = ip
		}
	}

	return lastIP
}

// GetRiskScore calculates a risk score for a login (0-100)
func (ad *AnomalyDetector) GetRiskScore(playerID uuid.UUID, ipAddress, userAgent string) int {
	anomalies := ad.DetectAnomalies(playerID, ipAddress, userAgent)

	score := 0

	for _, anomaly := range anomalies {
		switch anomaly {
		case "new_ip":
			score += 20
		case "new_device":
			score += 15
		case "unusual_time":
			score += 10
		case "rapid_ip_change":
			score += 30
		case "impossible_travel":
			score += 50
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

// ShouldChallenge determines if additional verification is needed
func (ad *AnomalyDetector) ShouldChallenge(playerID uuid.UUID, ipAddress, userAgent string) bool {
	score := ad.GetRiskScore(playerID, ipAddress, userAgent)

	// Challenge if risk score is high (>50)
	return score > 50
}

// GetStats returns statistics
func (ad *AnomalyDetector) GetStats() map[string]interface{} {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	totalIPs := 0
	totalLogins := 0

	for _, history := range ad.history {
		totalIPs += len(history.IPAddresses)
		totalLogins += len(history.Timestamps)
	}

	return map[string]interface{}{
		"tracked_players": len(ad.history),
		"total_ips":       totalIPs,
		"total_logins":    totalLogins,
	}
}
