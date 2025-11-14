// File: internal/security/honeypot.go
// Project: Terminal Velocity
// Description: Honeypot accounts and attack detection
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-14

package security

import (
	"strings"
	"sync"
	"time"
)

// HoneypotDetector detects attempts to access honeypot accounts
type HoneypotDetector struct {
	mu              sync.RWMutex
	honeypotNames   map[string]bool
	attempts        map[string][]time.Time // IP -> timestamps
	autobanEnabled  bool
	autobanDuration time.Duration
}

// NewHoneypotDetector creates a new honeypot detector
func NewHoneypotDetector(autobanEnabled bool, autobanDuration time.Duration) *HoneypotDetector {
	hd := &HoneypotDetector{
		honeypotNames:   make(map[string]bool),
		attempts:        make(map[string][]time.Time),
		autobanEnabled:  autobanEnabled,
		autobanDuration: autobanDuration,
	}

	// Add default honeypot usernames
	defaultHoneypots := []string{
		"admin", "administrator", "root", "system",
		"sysadmin", "superadmin", "test", "guest",
		"demo", "support", "helpdesk", "moderator",
		"mod", "staff", "owner", "server",
		"bot", "npc", "null", "undefined",
		"admin123", "root123", "test123",
		"administrator1", "admin1", "testuser",
	}

	for _, name := range defaultHoneypots {
		hd.AddHoneypot(name)
	}

	log.Info("Honeypot detector initialized with %d honeypot accounts", len(hd.honeypotNames))

	return hd
}

// AddHoneypot adds a honeypot username
func (hd *HoneypotDetector) AddHoneypot(username string) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.honeypotNames[strings.ToLower(username)] = true
}

// RemoveHoneypot removes a honeypot username
func (hd *HoneypotDetector) RemoveHoneypot(username string) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	delete(hd.honeypotNames, strings.ToLower(username))
}

// IsHoneypot checks if a username is a honeypot
func (hd *HoneypotDetector) IsHoneypot(username string) bool {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	return hd.honeypotNames[strings.ToLower(username)]
}

// RecordAttempt records an attempt to access a honeypot account
func (hd *HoneypotDetector) RecordAttempt(username, ipAddress string) bool {
	if !hd.IsHoneypot(username) {
		return false
	}

	hd.mu.Lock()
	defer hd.mu.Unlock()

	now := time.Now()

	// Record attempt
	if _, exists := hd.attempts[ipAddress]; !exists {
		hd.attempts[ipAddress] = make([]time.Time, 0)
	}

	hd.attempts[ipAddress] = append(hd.attempts[ipAddress], now)

	log.Warn("HONEYPOT TRIGGERED: username=%s, ip=%s, total_attempts=%d",
		username, ipAddress, len(hd.attempts[ipAddress]))

	return true
}

// ShouldAutoban checks if an IP should be auto-banned for honeypot attempts
func (hd *HoneypotDetector) ShouldAutoban(ipAddress string) bool {
	if !hd.autobanEnabled {
		return false
	}

	hd.mu.RLock()
	defer hd.mu.RUnlock()

	attempts, exists := hd.attempts[ipAddress]
	if !exists {
		return false
	}

	// Auto-ban after first honeypot attempt (aggressive policy)
	// Adjust threshold as needed
	return len(attempts) >= 1
}

// GetAttempts returns honeypot attempts for an IP
func (hd *HoneypotDetector) GetAttempts(ipAddress string) []time.Time {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	attempts, exists := hd.attempts[ipAddress]
	if !exists {
		return nil
	}

	// Return copy
	result := make([]time.Time, len(attempts))
	copy(result, attempts)

	return result
}

// GetAllAttempts returns all honeypot attempts
func (hd *HoneypotDetector) GetAllAttempts() map[string][]time.Time {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	// Return copy
	result := make(map[string][]time.Time)
	for ip, attempts := range hd.attempts {
		result[ip] = make([]time.Time, len(attempts))
		copy(result[ip], attempts)
	}

	return result
}

// GetStats returns statistics
func (hd *HoneypotDetector) GetStats() map[string]interface{} {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	totalAttempts := 0
	for _, attempts := range hd.attempts {
		totalAttempts += len(attempts)
	}

	return map[string]interface{}{
		"honeypot_accounts": len(hd.honeypotNames),
		"suspicious_ips":    len(hd.attempts),
		"total_attempts":    totalAttempts,
		"autoban_enabled":   hd.autobanEnabled,
	}
}
