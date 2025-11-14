-- Migration: 010_security_v2_tables
-- Description: Add comprehensive security tables for Security V2 features
-- Author: Joshua Ferguson
-- Date: 2025-11-14

-- Account Activity Events
CREATE TABLE IF NOT EXISTS account_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT true,
    details JSONB,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'none',

    -- Indexes for fast queries
    CONSTRAINT account_events_risk_level_check CHECK (risk_level IN ('none', 'low', 'medium', 'high', 'critical'))
);

CREATE INDEX idx_account_events_player_id ON account_events(player_id);
CREATE INDEX idx_account_events_timestamp ON account_events(timestamp DESC);
CREATE INDEX idx_account_events_risk_level ON account_events(risk_level) WHERE risk_level IN ('high', 'critical');
CREATE INDEX idx_account_events_event_type ON account_events(event_type);
CREATE INDEX idx_account_events_ip_address ON account_events(ip_address);

-- Two-Factor Authentication Configuration
CREATE TABLE IF NOT EXISTS player_two_factor (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID UNIQUE NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    secret VARCHAR(255) NOT NULL,
    backup_codes TEXT[], -- Array of backup codes
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used TIMESTAMP,
    recovery_email VARCHAR(255),

    CONSTRAINT player_two_factor_secret_check CHECK (length(secret) >= 16)
);

CREATE INDEX idx_player_two_factor_player_id ON player_two_factor(player_id);
CREATE INDEX idx_player_two_factor_enabled ON player_two_factor(enabled) WHERE enabled = true;

-- Password Reset Tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMP,
    ip_address VARCHAR(45) NOT NULL,

    CONSTRAINT password_reset_tokens_expiry_check CHECK (expires_at > created_at)
);

CREATE INDEX idx_password_reset_tokens_player_id ON password_reset_tokens(player_id);
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token) WHERE used = false;
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- IP Whitelist for Admin Accounts
CREATE TABLE IF NOT EXISTS admin_ip_whitelist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    cidr_mask INTEGER DEFAULT 32, -- For CIDR notation (e.g., /24)
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES players(id),
    is_active BOOLEAN NOT NULL DEFAULT true,

    CONSTRAINT admin_ip_whitelist_unique UNIQUE (admin_id, ip_address),
    CONSTRAINT admin_ip_whitelist_cidr_check CHECK (cidr_mask >= 0 AND cidr_mask <= 32)
);

CREATE INDEX idx_admin_ip_whitelist_admin_id ON admin_ip_whitelist(admin_id);
CREATE INDEX idx_admin_ip_whitelist_active ON admin_ip_whitelist(is_active) WHERE is_active = true;

-- Login History for Anomaly Detection
CREATE TABLE IF NOT EXISTS login_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    country_code CHAR(2), -- ISO 3166-1 alpha-2 (for future GeoIP)
    city VARCHAR(255),
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(255),
    anomalies JSONB, -- Array of detected anomalies
    risk_score INTEGER DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT login_history_risk_score_check CHECK (risk_score >= 0 AND risk_score <= 100)
);

CREATE INDEX idx_login_history_player_id ON login_history(player_id);
CREATE INDEX idx_login_history_timestamp ON login_history(timestamp DESC);
CREATE INDEX idx_login_history_ip_address ON login_history(ip_address);
CREATE INDEX idx_login_history_risk_score ON login_history(risk_score) WHERE risk_score > 50;

-- Player Sessions (for concurrent session tracking)
CREATE TABLE IF NOT EXISTS player_sessions (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,

    CONSTRAINT player_sessions_expiry_check CHECK (expires_at > started_at)
);

CREATE INDEX idx_player_sessions_player_id ON player_sessions(player_id);
CREATE INDEX idx_player_sessions_active ON player_sessions(is_active) WHERE is_active = true;
CREATE INDEX idx_player_sessions_expires_at ON player_sessions(expires_at);

-- Honeypot Attempts Tracking
CREATE TABLE IF NOT EXISTS honeypot_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username_attempted VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    autobanned BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT honeypot_attempts_unique UNIQUE (ip_address, username_attempted, timestamp)
);

CREATE INDEX idx_honeypot_attempts_ip_address ON honeypot_attempts(ip_address);
CREATE INDEX idx_honeypot_attempts_timestamp ON honeypot_attempts(timestamp DESC);
CREATE INDEX idx_honeypot_attempts_autobanned ON honeypot_attempts(autobanned) WHERE autobanned = true;

-- Rate Limiting Tracking (for action-based rate limits)
CREATE TABLE IF NOT EXISTS rate_limit_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    action_count INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMP NOT NULL DEFAULT NOW(),
    window_end TIMESTAMP NOT NULL,
    blocked BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT rate_limit_tracking_window_check CHECK (window_end > window_start),
    CONSTRAINT rate_limit_tracking_unique UNIQUE (player_id, ip_address, action_type, window_start)
);

CREATE INDEX idx_rate_limit_tracking_player_id ON rate_limit_tracking(player_id);
CREATE INDEX idx_rate_limit_tracking_ip_address ON rate_limit_tracking(ip_address);
CREATE INDEX idx_rate_limit_tracking_action_type ON rate_limit_tracking(action_type);
CREATE INDEX idx_rate_limit_tracking_window_end ON rate_limit_tracking(window_end);

-- Security Settings (per-player security preferences)
CREATE TABLE IF NOT EXISTS player_security_settings (
    player_id UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    login_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    new_ip_email_alert BOOLEAN NOT NULL DEFAULT true,
    session_timeout_minutes INTEGER NOT NULL DEFAULT 15,
    require_2fa BOOLEAN NOT NULL DEFAULT false,
    allow_password_reset_email BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT player_security_settings_timeout_check CHECK (session_timeout_minutes >= 5 AND session_timeout_minutes <= 1440)
);

-- Add trigger to update updated_at
CREATE OR REPLACE FUNCTION update_security_settings_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_security_settings_timestamp
BEFORE UPDATE ON player_security_settings
FOR EACH ROW
EXECUTE FUNCTION update_security_settings_timestamp();

-- Trusted Devices (for remember this device functionality)
CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    trusted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,

    CONSTRAINT trusted_devices_unique UNIQUE (player_id, device_fingerprint),
    CONSTRAINT trusted_devices_expiry_check CHECK (expires_at > trusted_at)
);

CREATE INDEX idx_trusted_devices_player_id ON trusted_devices(player_id);
CREATE INDEX idx_trusted_devices_active ON trusted_devices(is_active) WHERE is_active = true;
CREATE INDEX idx_trusted_devices_expires_at ON trusted_devices(expires_at);

-- Comments for documentation
COMMENT ON TABLE account_events IS 'Tracks all security-relevant account activities for audit trail';
COMMENT ON TABLE player_two_factor IS 'Stores two-factor authentication configuration for players';
COMMENT ON TABLE password_reset_tokens IS 'Manages password reset tokens with expiration';
COMMENT ON TABLE admin_ip_whitelist IS 'IP whitelist for admin account access restriction';
COMMENT ON TABLE login_history IS 'Detailed login history for anomaly detection and analysis';
COMMENT ON TABLE player_sessions IS 'Active player sessions for concurrent session management';
COMMENT ON TABLE honeypot_attempts IS 'Tracks attempts to access honeypot accounts';
COMMENT ON TABLE rate_limit_tracking IS 'Tracks action-based rate limiting per player/IP';
COMMENT ON TABLE player_security_settings IS 'Per-player security preferences and settings';
COMMENT ON TABLE trusted_devices IS 'Trusted devices for streamlined authentication';

-- Insert default security settings for existing players
INSERT INTO player_security_settings (player_id)
SELECT id FROM players
ON CONFLICT (player_id) DO NOTHING;
