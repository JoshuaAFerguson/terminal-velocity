# Rate Limiting and Security System

**Feature**: Connection Rate Limiting and Security
**Phase**: 20
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Rate Limiting system provides comprehensive security controls for SSH connections, including connection rate limiting, authentication attempt tracking, and automatic IP banning. The system prevents brute force attacks and resource exhaustion.

### Key Features

- **Connection Rate Limiting**: 5 concurrent per IP, 20/minute per IP
- **Authentication Rate Limiting**: 5 attempts before 15-minute lockout
- **Auto-Ban System**: 20 failures triggers 24-hour ban
- **IP Tracking**: Per-IP connection and authentication tracking
- **Automatic Cleanup**: Periodic cleanup of old entries
- **Brute Force Protection**: Prevents password guessing attacks
- **Manual Ban/Unban**: Admin controls for IP management

---

## Architecture

### Components

1. **Rate Limiter** (`internal/ratelimit/ratelimit.go`)
   - Connection tracking per IP
   - Authentication attempt tracking
   - Ban management
   - Thread-safe with `sync.RWMutex`

2. **Security Integration**:
   - SSH server integration
   - Admin system integration
   - Metrics reporting

### Data Flow

```
Connection Attempt
         ↓
Check IP Ban Status
         ↓
Check Connection Limit
         ↓
[If Allowed]
Track Connection
         ↓
Authentication Attempt
         ↓
Track Auth Attempt
         ↓
[If Failed]
Increment Failure Count
         ↓
Check Auto-Ban Threshold
         ↓
[If Reached]
Ban IP Automatically
```

---

## Implementation Details

### Rate Limiter

```go
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

    // Cleanup
    cleanupTicker *time.Ticker
    stopChan      chan struct{}
    wg            sync.WaitGroup
}
```

### Connection Tracking

**Connection Tracker**:
```go
type connectionTracker struct {
    count       int           // Current concurrent connections
    timestamps  []time.Time   // Recent connection attempts
    lastAttempt time.Time
}
```

**Allow Connection**:
```go
func (l *Limiter) AllowConnection(addr net.Addr) (bool, string) {
    ip := extractIP(addr)

    l.mu.Lock()
    defer l.mu.Unlock()

    // Check if IP is banned
    if ban, ok := l.bannedIPs[ip]; ok {
        if ban.isPermanent || time.Now().Before(ban.expiresAt) {
            return false, "IP address is banned: " + ban.reason
        }
        delete(l.bannedIPs, ip) // Ban expired
    }

    tracker, ok := l.connections[ip]
    if !ok {
        tracker = &connectionTracker{
            timestamps: make([]time.Time, 0),
        }
        l.connections[ip] = tracker
    }

    // Check concurrent connections
    if tracker.count >= l.maxConnPerIP {
        return false, "too many concurrent connections"
    }

    // Check connection rate
    now := time.Now()
    cutoff := now.Add(-l.connectionWindow)

    // Clean old timestamps
    newTimestamps := make([]time.Time, 0)
    for _, ts := range tracker.timestamps {
        if ts.After(cutoff) {
            newTimestamps = append(newTimestamps, ts)
        }
    }
    tracker.timestamps = newTimestamps

    if len(tracker.timestamps) >= l.maxConnPerMin {
        return false, "connection rate limit exceeded"
    }

    // Allow connection
    tracker.count++
    tracker.timestamps = append(tracker.timestamps, now)
    tracker.lastAttempt = now

    return true, ""
}
```

**Release Connection**:
```go
func (l *Limiter) ReleaseConnection(addr net.Addr) {
    ip := extractIP(addr)

    l.mu.Lock()
    defer l.mu.Unlock()

    if tracker, ok := l.connections[ip]; ok {
        if tracker.count > 0 {
            tracker.count--
        }
    }
}
```

### Authentication Tracking

**Auth Tracker**:
```go
type authTracker struct {
    failures    int           // Total failures
    timestamps  []time.Time   // Recent failure timestamps
    lockedUntil time.Time     // Lockout expiration
}
```

**Record Auth Failure**:
```go
func (l *Limiter) RecordAuthFailure(addr net.Addr, username string) {
    ip := extractIP(addr)

    l.mu.Lock()
    defer l.mu.Unlock()

    tracker, ok := l.authAttempts[ip]
    if !ok {
        tracker = &authTracker{
            timestamps: make([]time.Time, 0),
        }
        l.authAttempts[ip] = tracker
    }

    now := time.Now()

    // Clean old timestamps (within auth window)
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

    // Check for lockout
    if len(tracker.timestamps) >= l.maxAuthAttempts {
        tracker.lockedUntil = now.Add(l.authLockoutTime)
        log.Warn("Auth lockout for %s until %v (failures: %d)",
            ip, tracker.lockedUntil, tracker.failures)
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
        log.Warn("Auto-banned %s for %v (total failures: %d)",
            ip, l.autobanDuration, tracker.failures)
    }
}
```

**Record Auth Success**:
```go
func (l *Limiter) RecordAuthSuccess(addr net.Addr) {
    ip := extractIP(addr)

    l.mu.Lock()
    defer l.mu.Unlock()

    // Clear auth tracker on success
    if tracker, ok := l.authAttempts[ip]; ok {
        tracker.failures = 0
        tracker.timestamps = make([]time.Time, 0)
        tracker.lockedUntil = time.Time{}
    }
}
```

**Check Auth Lock**:
```go
func (l *Limiter) IsAuthLocked(addr net.Addr) (bool, time.Duration) {
    ip := extractIP(addr)

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
```

### IP Ban Management

**Ban Info**:
```go
type banInfo struct {
    reason      string
    bannedAt    time.Time
    expiresAt   time.Time
    isPermanent bool
}
```

**Manual Ban**:
```go
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
```

**Unban**:
```go
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
```

### Automatic Cleanup

**Cleanup Worker**:
```go
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

func (l *Limiter) cleanup() {
    l.mu.Lock()
    defer l.mu.Unlock()

    now := time.Now()
    cleaned := 0

    // Clean connection trackers
    for ip, tracker := range l.connections {
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
        }
    }

    // Clean old auth trackers
    for ip, tracker := range l.authAttempts {
        if len(tracker.timestamps) == 0 && now.After(tracker.lockedUntil) {
            delete(l.authAttempts, ip)
            cleaned++
        }
    }

    if cleaned > 0 {
        log.Debug("Cleanup: removed %d old entries", cleaned)
    }
}
```

---

## Configuration

**Default Configuration**:
```go
type Config struct {
    // Connection limits
    MaxConnectionsPerIP     int           // Default: 5
    MaxConnectionsPerMinute int           // Default: 20
    ConnectionWindow        time.Duration // Default: 1 minute

    // Authentication limits
    MaxAuthAttempts int           // Default: 5
    AuthWindow      time.Duration // Default: 5 minutes
    AuthLockoutTime time.Duration // Default: 15 minutes

    // Auto-banning
    AutobanThreshold int           // Default: 20
    AutobanDuration  time.Duration // Default: 24 hours

    // Cleanup
    CleanupInterval time.Duration // Default: 5 minutes
}
```

**Example Configuration**:
```go
cfg := &ratelimit.Config{
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
```

---

## SSH Server Integration

**Connection Handling**:
```go
// In SSH server
func (s *Server) handleConnection(conn net.Conn) {
    // Check rate limit
    allowed, reason := s.rateLimiter.AllowConnection(conn.RemoteAddr())
    if !allowed {
        log.Warn("Connection rejected: %s (%s)", conn.RemoteAddr(), reason)
        conn.Close()
        return
    }

    defer s.rateLimiter.ReleaseConnection(conn.RemoteAddr())

    // Handle SSH connection...
}
```

**Authentication Handling**:
```go
func handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
    // Check if IP is locked
    if locked, remaining := rateLimiter.IsAuthLocked(conn.RemoteAddr()); locked {
        log.Warn("Auth locked for %s (remaining: %v)", conn.RemoteAddr(), remaining)
        return nil, fmt.Errorf("too many failed attempts, try again in %v", remaining)
    }

    // Attempt authentication
    player, err := playerRepo.GetPlayerByUsername(conn.User())
    if err != nil || !player.CheckPassword(string(password)) {
        rateLimiter.RecordAuthFailure(conn.RemoteAddr(), conn.User())
        return nil, errors.New("authentication failed")
    }

    // Success - clear failures
    rateLimiter.RecordAuthSuccess(conn.RemoteAddr())
    return &ssh.Permissions{...}, nil
}
```

---

## Monitoring

**Statistics**:
```go
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
        "tracked_ips":       len(l.connections),
        "total_connections": totalConnections,
        "locked_ips":        lockedIPs,
        "banned_ips":        totalBanned,
        "permanent_bans":    permanentBans,
        "auth_trackers":     len(l.authAttempts),
    }
}
```

---

## Best Practices

### Security Recommendations

1. **Adjust Thresholds**:
   - Lower limits for production
   - Higher limits for development
   - Monitor false positives

2. **Monitor Metrics**:
   - Track ban rates
   - Review auth failures
   - Alert on spikes

3. **Regular Review**:
   - Review banned IPs
   - Unban legitimate users
   - Adjust thresholds based on patterns

4. **Complementary Measures**:
   - Use fail2ban for system-level protection
   - Implement firewall rules
   - Use SSH key authentication when possible

---

## API Reference

### Core Functions

#### NewLimiter

```go
func NewLimiter(cfg *Config) *Limiter
```

Creates a new rate limiter.

#### AllowConnection

```go
func (l *Limiter) AllowConnection(
    addr net.Addr,
) (bool, string)
```

Checks if a connection should be allowed.

#### ReleaseConnection

```go
func (l *Limiter) ReleaseConnection(addr net.Addr)
```

Releases a connection slot.

#### RecordAuthFailure

```go
func (l *Limiter) RecordAuthFailure(
    addr net.Addr,
    username string,
)
```

Records a failed authentication attempt.

#### RecordAuthSuccess

```go
func (l *Limiter) RecordAuthSuccess(addr net.Addr)
```

Records a successful authentication.

#### BanIP

```go
func (l *Limiter) BanIP(
    ip string,
    reason string,
    duration time.Duration,
)
```

Manually bans an IP address.

#### UnbanIP

```go
func (l *Limiter) UnbanIP(ip string) bool
```

Unbans an IP address.

---

## Related Documentation

- [Admin System](./ADMIN_SYSTEM.md) - Manual ban management
- [Metrics & Monitoring](./METRICS_MONITORING.md) - Rate limit metrics

---

## File Locations

**Core Implementation**:
- `internal/ratelimit/ratelimit.go` - Rate limiter

**Integration**:
- `internal/server/server.go` - SSH server integration

**Documentation**:
- `docs/RATE_LIMITING.md` - This file
- `ROADMAP.md` - Phase 20 details

---

**For questions about the rate limiting system, contact the development team.**
