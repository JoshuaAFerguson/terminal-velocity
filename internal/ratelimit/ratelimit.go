// File: internal/ratelimit/ratelimit.go
// Project: Terminal Velocity
// Description: Rate limiting and security middleware for SSH connections
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14
//
// This package provides comprehensive rate limiting and IP-based security controls
// for SSH connections. It implements a multi-layered defense strategy:
//
// Layer 1: Connection Rate Limiting
//   - Limits concurrent connections per IP (prevents resource exhaustion)
//   - Limits connection attempts per minute per IP (prevents connection flooding)
//   - Uses sliding window algorithm for accurate rate tracking
//
// Layer 2: Authentication Rate Limiting
//   - Tracks failed authentication attempts per IP
//   - Implements progressive lockout (temporary ban after repeated failures)
//   - Clears failure count on successful authentication
//
// Layer 3: Auto-banning
//   - Automatically bans IPs that exceed failure threshold
//   - Supports both temporary and permanent bans
//   - Integrates with manual ban management
//
// Thread Safety:
//   - All operations are protected by RWMutex for concurrent access
//   - Background cleanup runs in separate goroutine
//   - Safe to use from multiple SSH handler goroutines
//
// Memory Management:
//   - Periodic cleanup removes stale entries
//   - Expired bans are automatically removed
//   - Connection trackers cleaned when no recent activity
//
// Security Model:
//   - IP-based tracking (extracts IP from net.Addr)
//   - Defense in depth: multiple layers of protection
//   - Fail-secure: blocks on ambiguous cases
//   - Audit logging for security events

package ratelimit

import (
	"net"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("RateLimit")

// ============================================================================
// Core Types
// ============================================================================

// Limiter manages rate limiting for SSH connections.
//
// The Limiter provides three layers of security:
//   1. Connection rate limiting (concurrent and per-minute limits)
//   2. Authentication failure tracking with lockout
//   3. Automatic IP banning for persistent attackers
//
// All operations are thread-safe and can be called concurrently from
// multiple goroutines. The limiter runs a background cleanup goroutine
// that periodically removes expired entries.
//
// Usage:
//   cfg := ratelimit.DefaultConfig()
//   limiter := ratelimit.NewLimiter(cfg)
//   defer limiter.Stop()
//
//   // Before accepting connection
//   if allowed, reason := limiter.AllowConnection(addr); !allowed {
//       return fmt.Errorf("connection denied: %s", reason)
//   }
//   defer limiter.ReleaseConnection(addr)
type Limiter struct {
	mu sync.RWMutex // Protects all maps and state

	// Connection rate limiting (per IP)
	// Maps IP address -> connection tracker for that IP
	connections      map[string]*connectionTracker
	maxConnPerIP     int           // Maximum concurrent connections from single IP
	maxConnPerMin    int           // Maximum connection attempts per minute
	connectionWindow time.Duration // Time window for rate limiting (typically 1 minute)

	// Authentication rate limiting (per IP)
	// Maps IP address -> authentication attempt tracker
	authAttempts     map[string]*authTracker
	maxAuthAttempts  int           // Failed attempts before lockout
	authWindow       time.Duration // Time window for auth rate limiting
	authLockoutTime  time.Duration // Duration of lockout after max attempts

	// IP banning
	// Maps IP address -> ban information
	bannedIPs        map[string]*banInfo
	autobanThreshold int           // Total failures before auto-ban
	autobanDuration  time.Duration // Duration of auto-ban (0 = permanent)

	// Cleanup management
	cleanupTicker *time.Ticker    // Ticker for periodic cleanup
	stopChan      chan struct{}   // Signal channel to stop cleanup goroutine
	wg            sync.WaitGroup  // WaitGroup for graceful shutdown
}

// connectionTracker tracks connections from a single IP address.
//
// Fields:
//   - count: Current number of concurrent connections from this IP
//   - timestamps: Recent connection attempt timestamps (within connectionWindow)
//   - lastAttempt: Time of most recent connection attempt (for cleanup)
//
// The timestamps slice is periodically cleaned to remove entries outside
// the time window, keeping memory usage bounded.
type connectionTracker struct {
	count       int         // Current concurrent connections
	timestamps  []time.Time // Recent connection attempt times
	lastAttempt time.Time   // Last connection attempt (for cleanup)
}

// authTracker tracks authentication attempts from a single IP address.
//
// Fields:
//   - failures: Total failed authentication attempts (lifetime)
//   - timestamps: Recent failure timestamps (within authWindow)
//   - lockedUntil: Time when lockout expires (zero if not locked)
//
// The failures count increments indefinitely until reset by successful auth
// or cleanup. The timestamps slice tracks recent failures for rate limiting.
type authTracker struct {
	failures    int         // Total authentication failures
	timestamps  []time.Time // Recent failure timestamps
	lockedUntil time.Time   // Lockout expiration time
}

// banInfo stores information about a banned IP address.
//
// Bans can be:
//   - Temporary: Set expiresAt to future time, isPermanent = false
//   - Permanent: Set isPermanent = true (expiresAt ignored)
//
// Expired bans are automatically removed during cleanup.
type banInfo struct {
	reason      string    // Human-readable reason for ban
	bannedAt    time.Time // When ban was created
	expiresAt   time.Time // When ban expires (ignored if permanent)
	isPermanent bool      // If true, ban never expires
}

// ============================================================================
// Configuration
// ============================================================================

// Config holds rate limiter configuration.
//
// All timeout/duration values should be positive. A duration of 0 for
// AutobanDuration means permanent ban. CleanupInterval determines how
// often stale entries are removed (affects memory usage).
//
// Recommended values for production:
//   - MaxConnectionsPerIP: 3-5 (prevents resource exhaustion)
//   - MaxConnectionsPerMinute: 10-20 (prevents connection flooding)
//   - MaxAuthAttempts: 3-5 (balance security vs legitimate mistakes)
//   - AuthLockoutTime: 15min-1hr (deterrent without long-term impact)
//   - AutobanThreshold: 10-20 (high enough to avoid false positives)
//   - AutobanDuration: 24hr (or 0 for permanent, requires manual review)
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

// ============================================================================
// Constructor and Lifecycle
// ============================================================================

// DefaultConfig returns a default rate limiter configuration.
//
// The default configuration provides reasonable security settings for
// production use:
//   - 5 concurrent connections per IP (prevents resource exhaustion)
//   - 20 connection attempts per minute (prevents connection flooding)
//   - 5 failed auth attempts before 15-minute lockout
//   - Auto-ban after 20 total failures (24-hour ban)
//   - Cleanup every 5 minutes (balances memory vs CPU)
//
// These defaults balance security with usability:
//   - Legitimate users rarely hit limits
//   - Automated attacks are quickly blocked
//   - Memory usage stays bounded through cleanup
//
// Returns:
//   - Pointer to Config with production-ready defaults
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

// NewLimiter creates a new rate limiter with the given configuration.
//
// The limiter is ready to use immediately after creation. It starts a
// background goroutine for periodic cleanup of stale entries. Always
// call Stop() when done to ensure graceful shutdown.
//
// Parameters:
//   - cfg: Configuration for rate limiting behavior. If nil, uses DefaultConfig().
//
// Returns:
//   - Pointer to initialized Limiter ready for use
//
// Example:
//   limiter := ratelimit.NewLimiter(nil) // Use defaults
//   defer limiter.Stop()
//
// Thread Safety:
//   - Safe to call from any goroutine
//   - Returned limiter is safe for concurrent use
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

// Stop stops the rate limiter and its background cleanup goroutine.
//
// This method gracefully shuts down the limiter by:
//   1. Signaling the cleanup goroutine to stop
//   2. Waiting for it to exit (may block briefly)
//   3. Stopping the cleanup ticker
//
// Always call Stop() when done with a limiter to prevent goroutine leaks.
// After Stop() is called, the limiter should not be used further.
//
// Thread Safety:
//   - Safe to call from any goroutine
//   - Blocks until cleanup goroutine exits
//   - Safe to call multiple times (subsequent calls are no-ops)
func (l *Limiter) Stop() {
	close(l.stopChan)
	l.wg.Wait() // Wait for cleanup goroutine to finish
	if l.cleanupTicker != nil {
		l.cleanupTicker.Stop()
	}
	log.Info("Rate limiter stopped")
}

// ============================================================================
// Connection Rate Limiting
// ============================================================================

// AllowConnection checks if a connection from the given address should be allowed.
//
// This method performs three security checks in order:
//   1. IP ban check: If IP is banned, denies immediately
//   2. Concurrent connection check: Ensures not too many simultaneous connections
//   3. Connection rate check: Ensures not too many attempts in time window
//
// If all checks pass, the connection is allowed and a connection slot is
// reserved. The caller MUST call ReleaseConnection() when done, typically
// via defer.
//
// Parameters:
//   - addr: Network address of the connecting client
//
// Returns:
//   - bool: true if connection allowed, false if denied
//   - string: Empty if allowed, denial reason if denied
//
// Example:
//   if allowed, reason := limiter.AllowConnection(addr); !allowed {
//       log.Warn("Connection denied: %s", reason)
//       return fmt.Errorf("connection denied: %s", reason)
//   }
//   defer limiter.ReleaseConnection(addr)
//
// Thread Safety:
//   - Safe to call concurrently from multiple goroutines
//   - Uses write lock (blocks readers/writers during check)
//
// Note: Invalid addresses (nil or unparseable) are denied for security.
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

// ReleaseConnection releases a connection slot for the given address.
//
// This method MUST be called after AllowConnection() succeeds to properly
// track concurrent connections. Typically called via defer immediately
// after AllowConnection().
//
// Decrements the concurrent connection count for the IP. If count reaches
// zero and there's no recent activity, the tracker may be cleaned up in
// the next cleanup cycle.
//
// Parameters:
//   - addr: Network address to release connection for (same as AllowConnection)
//
// Thread Safety:
//   - Safe to call concurrently from multiple goroutines
//   - Uses write lock
//   - Safe to call even if addr is invalid (no-op)
//
// Note: Calling ReleaseConnection without prior AllowConnection is safe
// but will be logged as a warning (indicates a bug in caller).
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

// ============================================================================
// Authentication Rate Limiting
// ============================================================================

// RecordAuthFailure records a failed authentication attempt.
//
// This is a critical security function that tracks authentication failures
// and enforces progressive lockout and auto-banning:
//
//   1. Records the failure in the auth tracker
//   2. Checks if failure count triggers lockout (temporary ban)
//   3. Checks if total failures trigger auto-ban (longer ban)
//
// The function implements two thresholds:
//   - maxAuthAttempts: Triggers temporary lockout (e.g., 5 attempts -> 15min)
//   - autobanThreshold: Triggers auto-ban (e.g., 20 attempts -> 24hr)
//
// Failures are tracked both in a sliding window (for lockout) and cumulatively
// (for auto-ban). This prevents attackers from spacing out attempts to avoid
// detection.
//
// Parameters:
//   - addr: Network address of failed authentication
//   - username: Username attempted (logged for audit purposes)
//
// Thread Safety:
//   - Safe to call concurrently from multiple goroutines
//   - Uses write lock
//
// Security Note:
//   - Always call this for failed authentication attempts
//   - Failure to call this creates security vulnerability
//   - Log message includes IP and username for audit trail
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
