// File: internal/ratelimit/ratelimit.go
// Project: Terminal Velocity
// Description: Rate limiting and security middleware for SSH connections
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package ratelimit

import (
	"net"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("RateLimit")

// Limiter manages rate limiting for SSH connections
type Limiter struct {
	mu sync.RWMutex

	// Connection rate limiting (per IP)
	connections      map[string]*connectionTracker
	maxConnPerIP     int
	maxConnPerMin    int
	connectionWindow time.Duration

	// Authentication rate limiting (per IP)
	authAttempts     map[string]*authTracker
	maxAuthAttempts  int
	authWindow       time.Duration
	authLockoutTime  time.Duration

	// IP banning
	bannedIPs        map[string]*banInfo
	autobanThreshold int
	autobanDuration  time.Duration

	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// connectionTracker tracks connections from a single IP
type connectionTracker struct {
	count       int
	timestamps  []time.Time
	lastAttempt time.Time
}

// authTracker tracks authentication attempts from a single IP
type authTracker struct {
	failures    int
	timestamps  []time.Time
	lockedUntil time.Time
}

// banInfo stores information about a banned IP
type banInfo struct {
	reason     string
	bannedAt   time.Time
	expiresAt  time.Time
	isPermanent bool
}

// Config holds rate limiter configuration
type Config struct {
	// Connection limits
	MaxConnectionsPerIP     int           // Max concurrent connections per IP
	MaxConnectionsPerMinute int           // Max connection attempts per minute per IP
	ConnectionWindow        time.Duration // Time window for connection rate limiting

	// Authentication limits
	MaxAuthAttempts int           // Max failed auth attempts before lockout
	AuthWindow      time.Duration // Time window for auth rate limiting
	AuthLockoutTime time.Duration // How long to lock out after max attempts

	// Auto-banning
	AutobanThreshold int           // Failed attempts before auto-ban
	AutobanDuration  time.Duration // How long to auto-ban (0 = permanent)

	// Cleanup
	CleanupInterval time.Duration // How often to clean up old entries
}

// DefaultConfig returns a default rate limiter configuration
func DefaultConfig() *Config {
	return &Config{
		MaxConnectionsPerIP:     5,
		MaxConnectionsPerMinute: 20,
		ConnectionWindow:        1 * time.Minute,
		MaxAuthAttempts:         5,
		AuthWindow:              5 * time.Minute,
		AuthLockoutTime:         15 * time.Minute,
		AutobanThreshold:        20,
		AutobanDuration:         24 * time.Hour,
		CleanupInterval:         5 * time.Minute,
	}
}

// NewLimiter creates a new rate limiter
func NewLimiter(cfg *Config) *Limiter {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	l := &Limiter{
		connections:      make(map[string]*connectionTracker),
		authAttempts:     make(map[string]*authTracker),
		bannedIPs:        make(map[string]*banInfo),
		maxConnPerIP:     cfg.MaxConnectionsPerIP,
		maxConnPerMin:    cfg.MaxConnectionsPerMinute,
		connectionWindow: cfg.ConnectionWindow,
		maxAuthAttempts:  cfg.MaxAuthAttempts,
		authWindow:       cfg.AuthWindow,
		authLockoutTime:  cfg.AuthLockoutTime,
		autobanThreshold: cfg.AutobanThreshold,
		autobanDuration:  cfg.AutobanDuration,
		stopChan:         make(chan struct{}),
	}

	// Start cleanup goroutine
	l.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
	l.wg.Add(1)
	go l.cleanupLoop()

	log.Info("Rate limiter initialized: maxConnPerIP=%d, maxAuthAttempts=%d, autobanThreshold=%d",
		cfg.MaxConnectionsPerIP, cfg.MaxAuthAttempts, cfg.AutobanThreshold)

	return l
}

// Stop stops the rate limiter cleanup goroutine
func (l *Limiter) Stop() {
	close(l.stopChan)
	l.wg.Wait() // Wait for cleanup goroutine to finish
	if l.cleanupTicker != nil {
		l.cleanupTicker.Stop()
	}
	log.Info("Rate limiter stopped")
}

// AllowConnection checks if a connection from the given address should be allowed
func (l *Limiter) AllowConnection(addr net.Addr) (bool, string) {
	ip := extractIP(addr)
	if ip == "" {
		return false, "invalid address"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if IP is banned
	if ban, ok := l.bannedIPs[ip]; ok {
		if ban.isPermanent || time.Now().Before(ban.expiresAt) {
			log.Warn("Blocked banned IP: %s (reason: %s)", ip, ban.reason)
			return false, "IP address is banned: " + ban.reason
		}
		// Ban expired, remove it
		delete(l.bannedIPs, ip)
	}

	// Get or create connection tracker
	tracker, ok := l.connections[ip]
	if !ok {
		tracker = &connectionTracker{
			timestamps: make([]time.Time, 0),
		}
		l.connections[ip] = tracker
	}

	now := time.Now()
	tracker.lastAttempt = now

	// Check concurrent connections
	if tracker.count >= l.maxConnPerIP {
		log.Warn("Connection rate limit exceeded for %s (concurrent: %d/%d)", ip, tracker.count, l.maxConnPerIP)
		return false, "too many concurrent connections"
	}

	// Clean old timestamps
	cutoff := now.Add(-l.connectionWindow)
	newTimestamps := make([]time.Time, 0)
	for _, ts := range tracker.timestamps {
		if ts.After(cutoff) {
			newTimestamps = append(newTimestamps, ts)
		}
	}
	tracker.timestamps = newTimestamps

	// Check connection rate
	if len(tracker.timestamps) >= l.maxConnPerMin {
		log.Warn("Connection rate limit exceeded for %s (rate: %d/%d per minute)", ip, len(tracker.timestamps), l.maxConnPerMin)
		return false, "connection rate limit exceeded"
	}

	// Allow connection
	tracker.count++
	tracker.timestamps = append(tracker.timestamps, now)
	log.Debug("Connection allowed for %s (concurrent: %d, recent: %d)", ip, tracker.count, len(tracker.timestamps))
	return true, ""
}

// ReleaseConnection releases a connection slot for the given address
func (l *Limiter) ReleaseConnection(addr net.Addr) {
	ip := extractIP(addr)
	if ip == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if tracker, ok := l.connections[ip]; ok {
		if tracker.count > 0 {
			tracker.count--
		}
		log.Debug("Connection released for %s (concurrent: %d)", ip, tracker.count)
	}
}

// RecordAuthFailure records a failed authentication attempt
func (l *Limiter) RecordAuthFailure(addr net.Addr, username string) {
	ip := extractIP(addr)
	if ip == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get or create auth tracker
	tracker, ok := l.authAttempts[ip]
	if !ok {
		tracker = &authTracker{
			timestamps: make([]time.Time, 0),
		}
		l.authAttempts[ip] = tracker
	}

	now := time.Now()

	// Clean old timestamps
	cutoff := now.Add(-l.authWindow)
	newTimestamps := make([]time.Time, 0)
	for _, ts := range tracker.timestamps {
		if ts.After(cutoff) {
			newTimestamps = append(newTimestamps, ts)
		}
	}
	tracker.timestamps = newTimestamps

	// Record failure
	tracker.failures++
	tracker.timestamps = append(tracker.timestamps, now)

	log.Warn("Auth failure for %s (user: %s, failures: %d)", ip, username, tracker.failures)

	// Check for lockout
	if len(tracker.timestamps) >= l.maxAuthAttempts {
		tracker.lockedUntil = now.Add(l.authLockoutTime)
		log.Warn("Auth lockout for %s until %v (failures: %d)", ip, tracker.lockedUntil, tracker.failures)
	}

	// Check for auto-ban
	if tracker.failures >= l.autobanThreshold {
		reason := "too many failed authentication attempts"
		l.bannedIPs[ip] = &banInfo{
			reason:      reason,
			bannedAt:    now,
			expiresAt:   now.Add(l.autobanDuration),
			isPermanent: l.autobanDuration == 0,
		}
		log.Warn("Auto-banned %s for %v (reason: %s, total failures: %d)", ip, l.autobanDuration, reason, tracker.failures)
	}
}

// RecordAuthSuccess records a successful authentication
func (l *Limiter) RecordAuthSuccess(addr net.Addr) {
	ip := extractIP(addr)
	if ip == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Clear auth tracker on success
	if tracker, ok := l.authAttempts[ip]; ok {
		tracker.failures = 0
		tracker.timestamps = make([]time.Time, 0)
		tracker.lockedUntil = time.Time{}
		log.Debug("Auth success for %s, cleared failure count", ip)
	}
}

// IsAuthLocked checks if authentication is locked for the given address
func (l *Limiter) IsAuthLocked(addr net.Addr) (bool, time.Duration) {
	ip := extractIP(addr)
	if ip == "" {
		return false, 0
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if tracker, ok := l.authAttempts[ip]; ok {
		if time.Now().Before(tracker.lockedUntil) {
			remaining := time.Until(tracker.lockedUntil)
			return true, remaining
		}
	}

	return false, 0
}

// BanIP manually bans an IP address
func (l *Limiter) BanIP(ip string, reason string, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.bannedIPs[ip] = &banInfo{
		reason:      reason,
		bannedAt:    now,
		expiresAt:   now.Add(duration),
		isPermanent: duration == 0,
	}

	if duration == 0 {
		log.Info("Permanently banned %s (reason: %s)", ip, reason)
	} else {
		log.Info("Banned %s for %v (reason: %s)", ip, duration, reason)
	}
}

// UnbanIP manually unbans an IP address
func (l *Limiter) UnbanIP(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.bannedIPs[ip]; ok {
		delete(l.bannedIPs, ip)
		log.Info("Unbanned %s", ip)
		return true
	}

	return false
}

// IsBanned checks if an IP is currently banned
func (l *Limiter) IsBanned(ip string) (bool, string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if ban, ok := l.bannedIPs[ip]; ok {
		if ban.isPermanent || time.Now().Before(ban.expiresAt) {
			return true, ban.reason
		}
	}

	return false, ""
}

// GetStats returns statistics about the rate limiter
func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	totalConnections := 0
	for _, tracker := range l.connections {
		totalConnections += tracker.count
	}

	totalBanned := len(l.bannedIPs)
	permanentBans := 0
	for _, ban := range l.bannedIPs {
		if ban.isPermanent {
			permanentBans++
		}
	}

	lockedIPs := 0
	now := time.Now()
	for _, tracker := range l.authAttempts {
		if now.Before(tracker.lockedUntil) {
			lockedIPs++
		}
	}

	return map[string]interface{}{
		"tracked_ips":        len(l.connections),
		"total_connections":  totalConnections,
		"locked_ips":         lockedIPs,
		"banned_ips":         totalBanned,
		"permanent_bans":     permanentBans,
		"auth_trackers":      len(l.authAttempts),
	}
}

// cleanupLoop periodically cleans up old entries
func (l *Limiter) cleanupLoop() {
	defer l.wg.Done()

	for {
		select {
		case <-l.cleanupTicker.C:
			l.cleanup()
		case <-l.stopChan:
			return
		}
	}
}

// cleanup removes old entries from the rate limiter
func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cleaned := 0

	// Clean connection trackers
	for ip, tracker := range l.connections {
		// Remove if no concurrent connections and no recent activity
		if tracker.count == 0 && now.Sub(tracker.lastAttempt) > l.connectionWindow {
			delete(l.connections, ip)
			cleaned++
		}
	}

	// Clean expired bans
	for ip, ban := range l.bannedIPs {
		if !ban.isPermanent && now.After(ban.expiresAt) {
			delete(l.bannedIPs, ip)
			cleaned++
			log.Debug("Removed expired ban for %s", ip)
		}
	}

	// Clean old auth trackers
	for ip, tracker := range l.authAttempts {
		// Remove if no recent failures and not locked
		if len(tracker.timestamps) == 0 && now.After(tracker.lockedUntil) {
			delete(l.authAttempts, ip)
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Debug("Cleanup: removed %d old entries", cleaned)
	}
}

// extractIP extracts the IP address from a net.Addr
func extractIP(addr net.Addr) string {
	if addr == nil {
		return ""
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		// If no port, assume it's just an IP
		return addr.String()
	}

	return host
}
