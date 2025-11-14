-- File: scripts/migrations/001_add_admin_tables.sql
-- Project: Terminal Velocity
-- Description: Add admin, bans, mutes, and audit log tables
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-01-14

-- Admin users
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(32) NOT NULL,
    role VARCHAR(20) NOT NULL,
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES players(id) ON DELETE SET NULL,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(player_id)
);

-- Player bans
CREATE TABLE IF NOT EXISTS player_bans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(32) NOT NULL,
    ip_address VARCHAR(45),
    reason TEXT NOT NULL,
    banned_by UUID NOT NULL REFERENCES players(id),
    banned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_permanent BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE
);

-- Player mutes
CREATE TABLE IF NOT EXISTS player_mutes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(32) NOT NULL,
    reason TEXT NOT NULL,
    muted_by UUID NOT NULL REFERENCES players(id),
    muted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- Admin actions (audit log)
CREATE TABLE IF NOT EXISTS admin_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES players(id) ON DELETE SET NULL,
    admin_name VARCHAR(32) NOT NULL,
    action VARCHAR(50) NOT NULL,
    target_id UUID,
    target_name VARCHAR(100),
    details TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    success BOOLEAN DEFAULT TRUE,
    error_msg TEXT
);

-- Server settings (single row configuration)
CREATE TABLE IF NOT EXISTS server_settings (
    id INTEGER PRIMARY KEY DEFAULT 1,
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES players(id),
    CONSTRAINT only_one_settings_row CHECK (id = 1)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_admin_users_player ON admin_users(player_id);
CREATE INDEX IF NOT EXISTS idx_admin_users_active ON admin_users(is_active);
CREATE INDEX IF NOT EXISTS idx_player_bans_player ON player_bans(player_id);
CREATE INDEX IF NOT EXISTS idx_player_bans_active ON player_bans(is_active, expires_at);
CREATE INDEX IF NOT EXISTS idx_player_mutes_player ON player_mutes(player_id);
CREATE INDEX IF NOT EXISTS idx_player_mutes_active ON player_mutes(is_active, expires_at);
CREATE INDEX IF NOT EXISTS idx_admin_actions_admin ON admin_actions(admin_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_admin_actions_timestamp ON admin_actions(timestamp DESC);

-- Comments
COMMENT ON TABLE admin_users IS 'Server administrators and moderators';
COMMENT ON TABLE player_bans IS 'Banned players with expiration tracking';
COMMENT ON TABLE player_mutes IS 'Muted players with expiration tracking';
COMMENT ON TABLE admin_actions IS 'Audit log of all admin actions';
COMMENT ON TABLE server_settings IS 'Server configuration (single row)';
