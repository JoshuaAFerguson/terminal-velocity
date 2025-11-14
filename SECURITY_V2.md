# Security V2 - Comprehensive Security Enhancement

**Version**: 2.0.0
**Date**: 2025-11-14
**Author**: Joshua Ferguson

This document describes the Security V2 implementation for Terminal Velocity, which adds comprehensive security features including session management, activity logging, anomaly detection, two-factor authentication, and automated security scanning.

---

## Table of Contents

1. [Overview](#overview)
2. [New Features](#new-features)
3. [Architecture](#architecture)
4. [Database Schema](#database-schema)
5. [Integration Guide](#integration-guide)
6. [Configuration](#configuration)
7. [API Documentation](#api-documentation)
8. [Migration Guide](#migration-guide)
9. [Testing](#testing)
10. [Security Considerations](#security-considerations)

---

## Overview

Security V2 builds upon the foundation established in Security V1 (persistent SSH host keys, password complexity, terminal injection prevention) by adding enterprise-grade security features:

- **Session Management**: Idle timeout, max duration, concurrent session limits
- **Activity Logging**: Comprehensive audit trail with risk-based classification
- **Honeypot Detection**: Trap accounts to detect automated attacks
- **Anomaly Detection**: Machine learning-inspired login anomaly detection
- **Two-Factor Authentication**: TOTP-based 2FA with backup codes
- **Automated Scanning**: CI/CD security pipeline with multiple tools
- **Terminal Injection Prevention**: Protection against ANSI escape code attacks

**Security Posture Improvement**:
- Before Security V2: 8.5/10
- After Security V2: **9.5/10** (estimated)

---

## New Features

### 1. Session Management

**Purpose**: Prevent session hijacking, resource exhaustion, and enforce re-authentication.

**Features**:
- Idle timeout: 15 minutes (configurable)
- Maximum session duration: 24 hours (configurable)
- Concurrent session limit: 3 per account (configurable)
- Warning before timeout: 2 minutes
- Automatic session cleanup
- Session tracking by player, IP, and user agent

**Security Benefits**:
- Mitigates session hijacking by limiting session lifetime
- Prevents resource exhaustion from abandoned sessions
- Enforces re-authentication periodically
- Detects concurrent session abuse

**Implementation**: `internal/security/session.go`

### 2. Account Activity Logging

**Purpose**: Provide comprehensive audit trail for security forensics and compliance.

**Features**:
- Event types: login_success, login_failure, password_changed, suspicious_login, new_ip_login, 2fa_enabled, 2fa_disabled, account_locked
- Risk levels: none, low, medium, high, critical
- Detailed metadata: IP address, user agent, timestamp, success/failure, custom details
- High-risk event querying
- Per-player activity history

**Security Benefits**:
- Detects compromised accounts
- Provides forensic evidence
- Enables security analytics
- Supports compliance requirements (audit logs)

**Implementation**: `internal/security/activity.go`

### 3. Honeypot Detection

**Purpose**: Detect and deter automated attacks by using trap accounts.

**Features**:
- Reserved honeypot usernames: admin, administrator, root, system, sysadmin, test, guest, operator, superuser, support
- Auto-ban on first attempt (configurable)
- Ban duration: 24 hours (configurable)
- Attempt tracking and statistics

**Security Benefits**:
- Early warning for automated attacks
- Aggressive response to malicious actors
- Low false positive rate (legitimate users won't try these names)

**Implementation**: `internal/security/honeypot.go`

### 4. Login Anomaly Detection

**Purpose**: Identify suspicious login attempts based on behavioral patterns.

**Anomaly Types**:
- **new_ip**: Login from previously unseen IP address
- **new_device**: Login from new user agent/device fingerprint
- **unusual_time**: Login at unusual hour (e.g., 3 AM for daytime player)
- **rapid_ip_change**: Multiple IP addresses in short time (impossible travel)
- **impossible_travel**: Geographic impossibility (requires GeoIP - future)

**Risk Scoring**:
- Each anomaly type contributes to risk score (0-100)
- Threshold configurable (default: 50)
- High-risk logins can trigger additional verification (2FA, email)

**Security Benefits**:
- Detects account takeovers
- Identifies credential stuffing attacks
- Flags compromised accounts

**Implementation**: `internal/security/anomaly.go`

### 5. Two-Factor Authentication (TOTP)

**Purpose**: Protect accounts against password compromise.

**Features**:
- Time-based One-Time Password (RFC 6238)
- QR code generation for easy setup
- 6-digit codes with 30-second validity window
- Backup codes for account recovery (10 codes)
- Per-player 2FA enable/disable
- Recovery email configuration

**Security Benefits**:
- Industry-standard 2FA
- Protects against password leaks
- Compatible with Google Authenticator, Authy, etc.

**Dependencies**:
- `github.com/pquerna/otp` - TOTP library
- `github.com/boombuler/barcode` - QR code generation

**Implementation**: `internal/security/totp.go`

### 6. Terminal Injection Prevention

**Purpose**: Prevent malicious ANSI escape codes from manipulating user terminals.

**Features**:
- ANSI escape code stripping
- Control character filtering
- Username sanitization
- Chat message sanitization
- Filename sanitization
- Injection detection

**Attack Vectors Prevented**:
- Terminal cursor manipulation
- Color/formatting abuse
- Terminal command injection
- Screen clearing/scrolling attacks
- Bell/beep spam

**Implementation**: `internal/validation/validation.go` (extended)

### 7. Automated Security Scanning

**Purpose**: Continuous security monitoring in CI/CD pipeline.

**Tools Integrated**:
- **Gosec**: Go security scanner (AST-based vulnerability detection)
- **govulncheck**: Official Go vulnerability database checker
- **Nancy**: Dependency vulnerability scanner
- **TruffleHog**: Secret detection (API keys, passwords, tokens)
- **Gitleaks**: Git secret scanner
- **golangci-lint**: Code quality and security linting
- **Trivy**: Container and filesystem vulnerability scanner

**Triggers**:
- Every push to repository
- Every pull request
- Weekly scheduled scan
- Manual workflow dispatch

**Implementation**: `.github/workflows/security.yml`

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Security Manager                        │
│                 (internal/security/manager.go)               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Session Mgr   │  │ Activity Log │  │  Honeypot    │     │
│  │ - Timeout     │  │ - Audit Trail│  │  - Detection │     │
│  │ - Concurrent  │  │ - Risk Levels│  │  - Auto-ban  │     │
│  └───────────────┘  └──────────────┘  └──────────────┘     │
│                                                               │
│  ┌───────────────┐  ┌──────────────┐                        │
│  │ Anomaly Det.  │  │  2FA (TOTP)  │                        │
│  │ - Behavior    │  │ - QR Codes   │                        │
│  │ - Risk Score  │  │ - Backup     │                        │
│  └───────────────┘  └──────────────┘                        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Database Layer                           │
│                                                               │
│  • account_events          • player_sessions                 │
│  • player_two_factor       • honeypot_attempts               │
│  • login_history           • rate_limit_tracking             │
│  • password_reset_tokens   • player_security_settings        │
│  • admin_ip_whitelist      • trusted_devices                 │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Security Manager

The `security.Manager` is the central coordinator for all security features. It integrates:

- **SessionManager**: Manages active sessions with timeout and limits
- **ActivityLogger**: Logs security-relevant events
- **HoneypotDetector**: Detects honeypot login attempts
- **AnomalyDetector**: Identifies suspicious login patterns

**Workflow**:

1. **Login Attempt** → `ValidateLogin()`
   - Check honeypot
   - Detect anomalies
   - Calculate risk score
   - Decide if additional verification needed

2. **Login Success** → `OnLoginSuccess()`
   - Check if new IP
   - Log successful login
   - Record in anomaly detector
   - Create session
   - Return session token

3. **Login Failure** → `OnLoginFailure()`
   - Log failure with reason
   - Check honeypot
   - Increment failure count (rate limiting)

4. **Session Activity** → `UpdateSessionActivity()`
   - Update last activity timestamp
   - Reset idle timeout

5. **Logout** → `OnLogout()`
   - Log logout event
   - Destroy session

---

## Database Schema

Security V2 adds 10 new tables to the database. Run the migration script to create them:

```bash
psql -U terminal_velocity -d terminal_velocity -f scripts/migrations/010_security_v2_tables.sql
```

### Tables Overview

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `account_events` | Activity audit trail | player_id, event_type, risk_level, ip_address |
| `player_two_factor` | 2FA configuration | player_id, secret, backup_codes, enabled |
| `password_reset_tokens` | Password reset | player_id, token, expires_at, used |
| `admin_ip_whitelist` | IP restrictions for admins | admin_id, ip_address, cidr_mask |
| `login_history` | Detailed login tracking | player_id, ip_address, anomalies, risk_score |
| `player_sessions` | Active sessions | id (session_id), player_id, expires_at, is_active |
| `honeypot_attempts` | Honeypot attack log | username_attempted, ip_address, autobanned |
| `rate_limit_tracking` | Action-based rate limits | player_id, ip_address, action_type, action_count |
| `player_security_settings` | Per-player preferences | player_id, login_notifications_enabled, session_timeout_minutes |
| `trusted_devices` | Device fingerprinting | player_id, device_fingerprint, trusted_at, expires_at |

### Indexes

All tables include optimized indexes for common queries:

- `account_events`: Indexed on player_id, timestamp, risk_level, event_type, ip_address
- `player_sessions`: Indexed on player_id, is_active, expires_at
- `login_history`: Indexed on player_id, timestamp, ip_address, risk_score
- `honeypot_attempts`: Indexed on ip_address, timestamp, autobanned
- `rate_limit_tracking`: Indexed on player_id, ip_address, action_type, window_end

### Constraints

- Check constraints ensure data integrity (e.g., risk_score 0-100)
- Foreign keys cascade deletes (player deletion removes all related data)
- Unique constraints prevent duplicates (e.g., one 2FA config per player)

---

## Integration Guide

### Basic Integration

**Step 1**: Import the security package

```go
import "github.com/JoshuaAFerguson/terminal-velocity/internal/security"
```

**Step 2**: Create security manager in server initialization

```go
// In cmd/server/main.go or internal/server/server.go

securityConfig := security.DefaultConfig()
// Customize if needed:
// securityConfig.SessionConfig.IdleTimeout = 20 * time.Minute
// securityConfig.HoneypotAutoban = true

securityManager := security.NewManager(securityConfig)
defer securityManager.Stop()
```

**Step 3**: Integrate into authentication flow

```go
func (s *Server) handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
    username := conn.User()
    ipAddress := conn.RemoteAddr().String()
    userAgent := "" // Extract from SSH client version if available

    // Get player from database
    player, err := s.playerRepo.GetByUsername(ctx, username)
    if err != nil {
        // Pre-validation (before we know player_id)
        s.securityManager.OnLoginFailure(username, ipAddress, "player_not_found")
        return nil, fmt.Errorf("authentication failed")
    }

    // Validate login with security manager
    if err := s.securityManager.ValidateLogin(username, string(password), ipAddress, userAgent, player.ID); err != nil {
        // Honeypot or anomaly detected
        return nil, err
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), password); err != nil {
        s.securityManager.OnLoginFailure(username, ipAddress, "invalid_password")
        return nil, fmt.Errorf("authentication failed")
    }

    // Success - create session
    session, err := s.securityManager.OnLoginSuccess(player.ID, username, ipAddress, userAgent)
    if err != nil {
        return nil, err
    }

    // Store session ID in SSH permissions for later use
    return &ssh.Permissions{
        Extensions: map[string]string{
            "player_id":  player.ID.String(),
            "session_id": session.ID.String(),
        },
    }, nil
}
```

**Step 4**: Update session activity during gameplay

```go
// In game loop or command handlers
func (s *Server) handleGameActivity(sessionID uuid.UUID) {
    if err := s.securityManager.UpdateSessionActivity(sessionID); err != nil {
        log.Warn("Failed to update session activity: %v", err)
    }
}
```

**Step 5**: Handle logout

```go
func (s *Server) handleLogout(sessionID uuid.UUID) error {
    return s.securityManager.OnLogout(sessionID)
}
```

### Two-Factor Authentication Integration

**Setup Flow** (when player enables 2FA):

```go
import "github.com/JoshuaAFerguson/terminal-velocity/internal/security"

func (s *Server) enable2FA(playerID uuid.UUID, username string) error {
    tfm := security.NewTwoFactorManager()

    // Generate secret
    secret, err := tfm.GenerateSecret(username)
    if err != nil {
        return err
    }

    // Generate QR code for display
    var qrBuf bytes.Buffer
    if err := tfm.GenerateQRCode(username, secret, &qrBuf); err != nil {
        return err
    }

    // Display QR code to user (convert to ASCII art or save as image)
    // ...

    // Generate backup codes
    backupCodes, err := tfm.GenerateBackupCodes(10)
    if err != nil {
        return err
    }

    // Display backup codes to user (IMPORTANT: one-time display)
    // ...

    // Save to database
    _, err = s.db.Exec(ctx, `
        INSERT INTO player_two_factor (player_id, enabled, secret, backup_codes)
        VALUES ($1, true, $2, $3)
        ON CONFLICT (player_id) DO UPDATE
        SET enabled = true, secret = $2, backup_codes = $3
    `, playerID, secret, pq.Array(backupCodes))

    return err
}
```

**Verification Flow** (during login):

```go
func (s *Server) verify2FACode(playerID uuid.UUID, code string) (bool, error) {
    tfm := security.NewTwoFactorManager()

    // Get 2FA config from database
    var secret string
    var backupCodes []string
    err := s.db.QueryRow(ctx, `
        SELECT secret, backup_codes
        FROM player_two_factor
        WHERE player_id = $1 AND enabled = true
    `, playerID).Scan(&secret, pq.Array(&backupCodes))

    if err != nil {
        return false, err
    }

    // Try TOTP code first
    if tfm.VerifyCode(secret, code) {
        return true, nil
    }

    // Try backup codes
    for i, backupCode := range backupCodes {
        if backupCode == code {
            // Remove used backup code
            backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
            _, err := s.db.Exec(ctx, `
                UPDATE player_two_factor
                SET backup_codes = $1, last_used = NOW()
                WHERE player_id = $2
            `, pq.Array(backupCodes), playerID)
            return true, err
        }
    }

    return false, nil
}
```

### Terminal Injection Prevention

**Sanitize all user input before display**:

```go
import "github.com/JoshuaAFerguson/terminal-velocity/internal/validation"

// Sanitize username before displaying
func displayUsername(username string) string {
    return validation.SanitizeUsername(username)
}

// Sanitize chat messages
func displayChatMessage(message string) string {
    return validation.SanitizeChatMessage(message)
}

// Check for injection attempts
func validateUserInput(field, value string) error {
    return validation.ValidateNoInjection(field, value)
}
```

---

## Configuration

### Security Manager Configuration

```go
type Config struct {
    // Session management
    SessionConfig *SessionConfig

    // Activity logging
    MaxActivityEvents int // Max events to keep in memory before flushing

    // Honeypot
    HoneypotAutoban         bool
    HoneypotAutobanDuration time.Duration

    // Anomaly detection
    RequireChallengeOnAnomaly bool
    AnomalyRiskThreshold      int // 0-100

    // General
    EnableAllFeatures bool
}
```

**Default Configuration**:

```go
config := &security.Config{
    SessionConfig:             security.DefaultSessionConfig(),
    MaxActivityEvents:         10000,
    HoneypotAutoban:           true,
    HoneypotAutobanDuration:   24 * time.Hour,
    RequireChallengeOnAnomaly: true,
    AnomalyRiskThreshold:      50,
    EnableAllFeatures:         true,
}
```

### Session Configuration

```go
type SessionConfig struct {
    IdleTimeout        time.Duration // Time before idle session expires (default: 15 minutes)
    MaxSessionDuration time.Duration // Maximum session length (default: 24 hours)
    MaxConcurrent      int           // Max concurrent sessions per player (default: 3)
    WarnBeforeKick     time.Duration // Warning time before kick (default: 2 minutes)
    CheckInterval      time.Duration // How often to check for expired sessions (default: 1 minute)
}
```

**Customization Example**:

```go
sessionConfig := &security.SessionConfig{
    IdleTimeout:        20 * time.Minute,  // 20 minutes idle
    MaxSessionDuration: 48 * time.Hour,    // 2 days max
    MaxConcurrent:      5,                 // Allow 5 concurrent sessions
    WarnBeforeKick:     5 * time.Minute,   // Warn 5 minutes before kick
    CheckInterval:      30 * time.Second,  // Check every 30 seconds
}
```

### Per-Player Security Settings

Players can customize their security preferences via the `player_security_settings` table:

```sql
CREATE TABLE player_security_settings (
    player_id UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    login_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    new_ip_email_alert BOOLEAN NOT NULL DEFAULT true,
    session_timeout_minutes INTEGER NOT NULL DEFAULT 15,
    require_2fa BOOLEAN NOT NULL DEFAULT false,
    allow_password_reset_email BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Accessing Settings**:

```go
func (s *Server) getPlayerSecuritySettings(playerID uuid.UUID) (*SecuritySettings, error) {
    var settings SecuritySettings
    err := s.db.QueryRow(ctx, `
        SELECT login_notifications_enabled, new_ip_email_alert, session_timeout_minutes, require_2fa
        FROM player_security_settings
        WHERE player_id = $1
    `, playerID).Scan(&settings.LoginNotifications, &settings.NewIPAlert, &settings.SessionTimeout, &settings.Require2FA)
    return &settings, err
}
```

---

## API Documentation

### Security Manager API

#### `NewManager(config *Config) *Manager`

Creates a new security manager with the given configuration.

**Parameters**:
- `config`: Security configuration (nil uses defaults)

**Returns**: Initialized security manager

**Example**:
```go
manager := security.NewManager(nil) // Use defaults
defer manager.Stop()
```

---

#### `ValidateLogin(username, password, ipAddress, userAgent string, playerID uuid.UUID) error`

Validates a login attempt with all security checks (honeypot, anomaly detection, rate limiting).

**Parameters**:
- `username`: Username attempting to log in
- `password`: Password (used for future password strength checks)
- `ipAddress`: Source IP address
- `userAgent`: User agent string (optional, can be empty)
- `playerID`: Player UUID (use uuid.Nil if player not found yet)

**Returns**: Error if login should be blocked, nil if allowed

**Example**:
```go
err := manager.ValidateLogin("alice", "password123", "192.168.1.1", "SSH-2.0", playerID)
if err != nil {
    // Login blocked (honeypot, high risk, etc.)
    return err
}
```

---

#### `OnLoginSuccess(playerID uuid.UUID, username, ipAddress, userAgent string) (*Session, error)`

Handles successful login, creates session, logs activity.

**Parameters**:
- `playerID`: Player UUID
- `username`: Username
- `ipAddress`: Source IP
- `userAgent`: User agent string

**Returns**: Created session, error if session creation fails

**Example**:
```go
session, err := manager.OnLoginSuccess(playerID, "alice", "192.168.1.1", "SSH-2.0")
if err != nil {
    return err
}
// Store session.ID for future requests
```

---

#### `OnLoginFailure(username, ipAddress, reason string)`

Logs failed login attempt.

**Parameters**:
- `username`: Username that failed
- `ipAddress`: Source IP
- `reason`: Failure reason (e.g., "invalid_password", "player_not_found")

**Example**:
```go
manager.OnLoginFailure("alice", "192.168.1.1", "invalid_password")
```

---

#### `OnLogout(sessionID uuid.UUID) error`

Handles logout, destroys session, logs activity.

**Parameters**:
- `sessionID`: Session ID to destroy

**Returns**: Error if session not found

**Example**:
```go
err := manager.OnLogout(sessionID)
```

---

#### `UpdateSessionActivity(sessionID uuid.UUID) error`

Updates last activity timestamp for session (resets idle timeout).

**Parameters**:
- `sessionID`: Session ID

**Returns**: Error if session not found

**Example**:
```go
// Call on every user action
err := manager.UpdateSessionActivity(sessionID)
```

---

#### `GetSession(sessionID uuid.UUID) (*Session, error)`

Retrieves session by ID.

**Returns**: Session object, error if not found

---

#### `GetPlayerSessions(playerID uuid.UUID) []*Session`

Returns all active sessions for a player.

---

#### `GetActiveSessions(playerID uuid.UUID) int`

Returns count of active sessions for a player.

---

#### `CheckSessionValid(sessionID uuid.UUID) (valid bool, reason string)`

Checks if session is still valid (not expired, not idle).

**Returns**:
- `valid`: true if session is valid
- `reason`: Reason if invalid ("Session not found", "Session idle timeout", etc.)

---

#### `GetTimeUntilTimeout(sessionID uuid.UUID) time.Duration`

Returns time remaining before session times out.

---

#### `GetPlayerActivity(playerID uuid.UUID, limit int) []*ActivityEvent`

Returns recent activity events for a player.

---

#### `GetHighRiskEvents(since time.Time) []*ActivityEvent`

Returns all high-risk events since a given time.

---

#### `IsHoneypot(username string) bool`

Checks if username is a honeypot account.

---

#### `GetStats() map[string]interface{}`

Returns comprehensive security statistics (sessions, activity, honeypot, anomaly).

---

### Two-Factor Authentication API

#### `NewTwoFactorManager() *TwoFactorManager`

Creates a new 2FA manager.

---

#### `GenerateSecret(username string) (string, error)`

Generates a new TOTP secret for a user.

**Returns**: Base32-encoded secret

---

#### `GenerateQRCode(username, secret string, output io.Writer) error`

Generates a QR code for easy 2FA setup.

**Parameters**:
- `username`: Account username
- `secret`: TOTP secret from GenerateSecret()
- `output`: Writer for PNG image data

**Example**:
```go
var buf bytes.Buffer
err := tfm.GenerateQRCode("alice", secret, &buf)
// buf now contains PNG image
```

---

#### `VerifyCode(secret, code string) bool`

Verifies a 6-digit TOTP code.

**Parameters**:
- `secret`: User's TOTP secret
- `code`: 6-digit code to verify

**Returns**: true if code is valid

---

#### `GenerateBackupCodes(count int) ([]string, error)`

Generates backup codes for account recovery.

**Parameters**:
- `count`: Number of codes to generate (recommended: 10)

**Returns**: Array of backup codes

---

## Migration Guide

### From Security V1 to V2

**Prerequisites**:
- Security V1 must be deployed (persistent SSH host keys, password validation)
- Database backup recommended before migration

**Step 1**: Apply database migrations

```bash
# Run migration script
psql -U terminal_velocity -d terminal_velocity -f scripts/migrations/010_security_v2_tables.sql

# Verify tables created
psql -U terminal_velocity -d terminal_velocity -c "\dt" | grep -E "account_events|player_two_factor|player_sessions"
```

**Step 2**: Install new dependencies

```bash
go get github.com/pquerna/otp@v1.4.0
go get github.com/boombuler/barcode@v1.0.1
go mod tidy
```

**Step 3**: Update server code

```go
import "github.com/JoshuaAFerguson/terminal-velocity/internal/security"

// In server initialization
securityManager := security.NewManager(nil)
defer securityManager.Stop()

// Update authentication flow (see Integration Guide)
```

**Step 4**: Initialize default security settings for existing players

```sql
-- Run after migration
INSERT INTO player_security_settings (player_id)
SELECT id FROM players
ON CONFLICT (player_id) DO NOTHING;
```

**Step 5**: Test thoroughly

- [ ] Login/logout works correctly
- [ ] Sessions expire after idle timeout
- [ ] Activity events logged
- [ ] Honeypot detection works
- [ ] 2FA setup and verification works

**Step 6**: Monitor security logs

```sql
-- Check recent activity
SELECT * FROM account_events ORDER BY timestamp DESC LIMIT 100;

-- Check high-risk events
SELECT * FROM account_events WHERE risk_level IN ('high', 'critical') ORDER BY timestamp DESC;

-- Check honeypot attempts
SELECT * FROM honeypot_attempts ORDER BY timestamp DESC;
```

### Rolling Back

If issues occur, you can roll back:

```sql
-- Drop Security V2 tables (WARNING: data loss)
DROP TABLE IF EXISTS trusted_devices CASCADE;
DROP TABLE IF EXISTS player_security_settings CASCADE;
DROP TABLE IF EXISTS rate_limit_tracking CASCADE;
DROP TABLE IF EXISTS honeypot_attempts CASCADE;
DROP TABLE IF EXISTS player_sessions CASCADE;
DROP TABLE IF EXISTS login_history CASCADE;
DROP TABLE IF EXISTS admin_ip_whitelist CASCADE;
DROP TABLE IF EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS player_two_factor CASCADE;
DROP TABLE IF EXISTS account_events CASCADE;

-- Remove dependencies
go get -u github.com/pquerna/otp@none
go get -u github.com/boombuler/barcode@none
go mod tidy
```

---

## Testing

### Unit Testing

**Session Management**:

```go
func TestSessionTimeout(t *testing.T) {
    config := &security.SessionConfig{
        IdleTimeout:   1 * time.Second,
        CheckInterval: 100 * time.Millisecond,
    }
    sm := security.NewSessionManager(config)
    defer sm.Stop()

    // Create session
    session, err := sm.CreateSession(playerID, "testuser", "192.168.1.1")
    require.NoError(t, err)

    // Wait for timeout
    time.Sleep(2 * time.Second)

    // Session should be invalid
    valid, reason := sm.CheckSessionValid(session.ID)
    assert.False(t, valid)
    assert.Equal(t, "Session idle timeout", reason)
}
```

**Activity Logging**:

```go
func TestActivityLogging(t *testing.T) {
    logger := security.NewActivityLogger(1000)

    // Log event
    logger.LogLoginSuccess(playerID, "testuser", "192.168.1.1", false)

    // Retrieve events
    events := logger.GetPlayerEvents(playerID, 10)
    assert.Len(t, events, 1)
    assert.Equal(t, security.ActivityLoginSuccess, events[0].EventType)
}
```

**Honeypot Detection**:

```go
func TestHoneypotDetection(t *testing.T) {
    detector := security.NewHoneypotDetector(true, 24*time.Hour)

    // Check honeypot username
    assert.True(t, detector.IsHoneypot("admin"))
    assert.False(t, detector.IsHoneypot("alice"))

    // Record attempt
    detector.RecordAttempt("admin", "192.168.1.1")

    // Should auto-ban
    assert.True(t, detector.ShouldAutoban("192.168.1.1"))
}
```

**Two-Factor Authentication**:

```go
func TestTOTPVerification(t *testing.T) {
    tfm := security.NewTwoFactorManager()

    // Generate secret
    secret, err := tfm.GenerateSecret("testuser")
    require.NoError(t, err)

    // Generate valid code
    code, err := totp.GenerateCode(secret, time.Now())
    require.NoError(t, err)

    // Verify code
    assert.True(t, tfm.VerifyCode(secret, code))

    // Invalid code
    assert.False(t, tfm.VerifyCode(secret, "000000"))
}
```

### Integration Testing

**Full Authentication Flow**:

```go
func TestSecureAuthenticationFlow(t *testing.T) {
    // Setup
    manager := security.NewManager(nil)
    defer manager.Stop()

    // Valid login
    err := manager.ValidateLogin("alice", "password", "192.168.1.1", "SSH-2.0", playerID)
    assert.NoError(t, err)

    // Create session
    session, err := manager.OnLoginSuccess(playerID, "alice", "192.168.1.1", "SSH-2.0")
    assert.NoError(t, err)
    assert.NotNil(t, session)

    // Session valid
    valid, _ := manager.CheckSessionValid(session.ID)
    assert.True(t, valid)

    // Update activity
    err = manager.UpdateSessionActivity(session.ID)
    assert.NoError(t, err)

    // Logout
    err = manager.OnLogout(session.ID)
    assert.NoError(t, err)

    // Session no longer valid
    valid, _ = manager.CheckSessionValid(session.ID)
    assert.False(t, valid)
}
```

**Honeypot Auto-Ban**:

```go
func TestHoneypotAutoBan(t *testing.T) {
    manager := security.NewManager(&security.Config{
        HoneypotAutoban: true,
        HoneypotAutobanDuration: 1 * time.Hour,
    })
    defer manager.Stop()

    // Attempt login to honeypot
    err := manager.ValidateLogin("admin", "password", "192.168.1.100", "", uuid.Nil)
    assert.Error(t, err)

    // Should be banned now
    err = manager.ValidateLogin("alice", "password", "192.168.1.100", "", playerID)
    assert.Error(t, err) // Banned IP
}
```

### Manual Testing Checklist

- [ ] Login creates session successfully
- [ ] Session expires after idle timeout (15 min)
- [ ] Session expires after max duration (24 hours)
- [ ] Concurrent session limit enforced (3 sessions)
- [ ] Activity logged for login/logout
- [ ] Honeypot username triggers auto-ban
- [ ] Anomaly detection flags new IP
- [ ] 2FA setup generates valid QR code
- [ ] 2FA verification accepts valid codes
- [ ] 2FA backup codes work
- [ ] Terminal injection prevention blocks malicious input
- [ ] Security scanning workflow runs on push

---

## Security Considerations

### Session Security

**Session IDs**:
- Use UUID v4 (cryptographically random)
- Never expose session IDs in logs
- Transmit only over encrypted channels (SSH)
- Invalidate on logout

**Session Storage**:
- Store sessions in memory (fast access)
- Persist critical data to database for recovery
- Clean up expired sessions regularly

**Timeout Tuning**:
- Balance security vs usability
- Shorter timeout = more secure but inconvenient
- Longer timeout = convenient but higher risk
- Default: 15 min idle, 24 hour max

### Activity Logging

**Privacy**:
- Log only security-relevant data
- Do NOT log passwords or sensitive PII
- IP addresses are logged (may be PII in some jurisdictions)
- Consider GDPR/privacy regulations

**Data Retention**:
- Define retention policy (e.g., 90 days)
- Purge old logs regularly
- Archive critical events before deletion

**Performance**:
- Use async logging to avoid blocking
- Batch database writes
- Monitor log volume

### Honeypot Detection

**Honeypot Username Selection**:
- Use obviously administrative names (admin, root)
- Avoid names real users might choose
- Document honeypot accounts

**False Positives**:
- Very low risk (legitimate users won't try admin/root)
- Consider warning before auto-ban
- Provide unban mechanism for mistakes

**Effectiveness**:
- Honeypots detect automated attacks, not targeted attacks
- Complement with other security measures
- Monitor honeypot attempts for threat intelligence

### Anomaly Detection

**Machine Learning Considerations**:
- Current implementation is rule-based (not ML)
- Future: Train models on normal behavior
- Risk of false positives (new device, travel)

**Risk Score Tuning**:
- Start with conservative threshold (50)
- Adjust based on false positive rate
- Different thresholds for different security levels

**Privacy**:
- IP geolocation may be considered PII
- User agent tracking may raise privacy concerns
- Provide opt-out for privacy-conscious users

### Two-Factor Authentication

**Secret Storage**:
- **CRITICAL**: Secrets must be encrypted at rest
- Use application-level encryption (e.g., AES-256-GCM)
- Never log secrets
- Rotate encryption keys periodically

**Backup Codes**:
- One-time use only
- Store hashed (bcrypt), not plaintext
- Display to user only once during setup
- Provide regeneration mechanism

**Recovery**:
- Backup codes are primary recovery method
- Consider email recovery as fallback
- Require strong verification for recovery
- Log all 2FA recovery events

**QR Code Security**:
- QR codes contain secret - handle carefully
- Display only to authenticated user
- Do not log or store QR code images
- Clear from memory after display

### Automated Security Scanning

**Secret Detection**:
- Never commit API keys, passwords, tokens
- Use environment variables for secrets
- Rotate secrets if accidentally committed
- Configure .gitignore properly

**Dependency Vulnerabilities**:
- Monitor security advisories
- Update dependencies promptly
- Use dependabot or renovate for automation
- Test before deploying updates

**False Positives**:
- Security scanners may flag test data
- Review and whitelist false positives
- Document exceptions

### Terminal Injection Prevention

**Defense in Depth**:
- Sanitize all user input
- Escape special characters
- Strip ANSI codes
- Validate before storage AND before display

**Bypass Attempts**:
- Attackers may try Unicode tricks
- Double encoding
- Null byte injection
- Test thoroughly with malicious payloads

### Rate Limiting

**Integration with Security V2**:
- Session creation counts toward rate limit
- Activity logging tracks rate limit violations
- Honeypot attempts trigger aggressive rate limiting
- Future: Dynamic rate limits based on risk score

**Distributed Rate Limiting**:
- Current: Per-server rate limiting
- Future: Shared state for multi-server deployments (Redis)

---

## Performance Considerations

### Memory Usage

**Session Storage**:
- Each session: ~200 bytes
- 1000 concurrent sessions: ~200 KB
- Negligible impact

**Activity Logging**:
- Each event: ~500 bytes
- 10,000 events in memory: ~5 MB
- Periodic flush to database recommended

**Anomaly Detection**:
- Stores recent login history per player
- Memory usage scales with player count
- Implement TTL for old data

### Database Performance

**Indexes**:
- All tables have optimized indexes
- Query performance: <10ms for most operations
- Monitor slow queries and add indexes as needed

**Connection Pooling**:
- Reuse database connections
- Configure pool size based on load
- Monitor connection exhaustion

**Batch Operations**:
- Batch insert activity events
- Bulk cleanup of expired sessions
- Use transactions for consistency

### CPU Usage

**Cryptography**:
- TOTP verification: <1ms
- bcrypt password hashing: 50-100ms (intentionally slow)
- Session validation: <1ms

**Background Workers**:
- Session cleanup: Runs every 1 minute
- Activity log flushing: Runs every 5 minutes
- Minimal CPU impact

---

## Troubleshooting

### Common Issues

**Issue**: Sessions expire too quickly

**Solution**: Adjust idle timeout in configuration:
```go
config.SessionConfig.IdleTimeout = 30 * time.Minute
```

---

**Issue**: Too many concurrent sessions error

**Solution**: Increase limit or clean up old sessions:
```go
config.SessionConfig.MaxConcurrent = 5
```

Or manually destroy old sessions:
```sql
DELETE FROM player_sessions WHERE is_active = false;
```

---

**Issue**: Honeypot auto-ban blocking legitimate users

**Solution**: Disable auto-ban or increase threshold:
```go
config.HoneypotAutoban = false
```

Or unban specific IP:
```sql
-- Remove from ban list (implementation dependent)
```

---

**Issue**: 2FA QR code not scanning

**Solution**:
- Ensure QR code is large enough (200x200 pixels minimum)
- Check secret is valid Base32
- Test with multiple authenticator apps
- Provide manual entry option

---

**Issue**: Activity logs growing too large

**Solution**: Implement retention policy:
```sql
DELETE FROM account_events WHERE timestamp < NOW() - INTERVAL '90 days';
```

---

**Issue**: Anomaly detection too sensitive (false positives)

**Solution**: Increase risk threshold:
```go
config.AnomalyRiskThreshold = 70 // Less sensitive
```

---

## Future Enhancements

### Planned Features

1. **Password Reset Flow**
   - Email-based password reset
   - Uses `password_reset_tokens` table
   - Time-limited tokens (1 hour expiration)

2. **IP Whitelisting for Admins**
   - Uses `admin_ip_whitelist` table
   - CIDR notation support
   - Automatic lockout from unauthorized IPs

3. **Trusted Devices**
   - Uses `trusted_devices` table
   - Device fingerprinting
   - "Remember this device for 30 days"

4. **GeoIP Integration**
   - Populate `country_code` and `city` in `login_history`
   - Impossible travel detection
   - Location-based alerts

5. **Machine Learning Anomaly Detection**
   - Train on normal behavior patterns
   - Adaptive risk scoring
   - Behavioral biometrics

6. **Security Dashboard**
   - Real-time security metrics
   - Anomaly visualization
   - Incident response tools

7. **Compliance Reporting**
   - GDPR compliance tools
   - Audit log export
   - Data retention management

---

## Support and Contributing

### Reporting Security Issues

**DO NOT** open public GitHub issues for security vulnerabilities.

Contact: joshua.ferguson@example.com (replace with actual security contact)

Provide:
- Description of vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Contributing

Contributions welcome! Areas for improvement:

- Additional anomaly detection rules
- Better risk scoring algorithms
- Performance optimizations
- Additional authenticator support (U2F, WebAuthn)
- Security dashboard UI
- Compliance tools

See CONTRIBUTING.md for guidelines.

---

## Changelog

### Version 2.0.0 (2025-11-14)

**Added**:
- Session management with idle timeout and concurrent limits
- Comprehensive activity logging with risk classification
- Honeypot account detection with auto-ban
- Login anomaly detection with risk scoring
- Two-factor authentication (TOTP) with QR codes
- Terminal injection prevention
- Automated security scanning CI/CD pipeline
- 10 new database tables for security features
- Player security settings customization
- Trusted devices foundation

**Changed**:
- Enhanced authentication flow with security checks
- Improved validation module with injection prevention

**Security**:
- Mitigates session hijacking
- Detects automated attacks
- Prevents terminal manipulation
- Continuous vulnerability monitoring

---

## License

MIT License - See LICENSE file for details.

---

## Acknowledgments

- **OWASP**: Security best practices and guidelines
- **RFC 6238**: TOTP specification
- **Gosec**: Go security scanner
- **Go Team**: govulncheck and security infrastructure

---

**Document Version**: 2.0.0
**Last Updated**: 2025-11-14
**Author**: Joshua Ferguson
