-- Terminal Velocity Database Schema
-- PostgreSQL

-- Create database (run as superuser)
-- CREATE DATABASE terminal_velocity;
-- CREATE USER terminal_velocity WITH PASSWORD 'your_password';
-- GRANT ALL PRIVILEGES ON DATABASE terminal_velocity TO terminal_velocity;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(32) UNIQUE NOT NULL,
    password_hash VARCHAR(255),  -- Nullable: users can auth with SSH keys only
    email VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Game state
    credits BIGINT DEFAULT 10000,
    current_system UUID,
    current_planet UUID,
    ship_id UUID,

    -- Position
    x DOUBLE PRECISION DEFAULT 0,
    y DOUBLE PRECISION DEFAULT 0,

    -- Progression - Combat
    combat_rating INTEGER DEFAULT 0,
    total_kills INTEGER DEFAULT 0,
    play_time BIGINT DEFAULT 0,

    -- Progression - Trading
    trading_rating INTEGER DEFAULT 0,
    total_trades INTEGER DEFAULT 0,
    trade_profit BIGINT DEFAULT 0,
    highest_profit BIGINT DEFAULT 0,

    -- Progression - Exploration
    exploration_rating INTEGER DEFAULT 0,
    systems_visited INTEGER DEFAULT 0,
    total_jumps INTEGER DEFAULT 0,

    -- Progression - Missions
    missions_completed INTEGER DEFAULT 0,
    missions_failed INTEGER DEFAULT 0,

    -- Progression - Quests
    quests_completed INTEGER DEFAULT 0,

    -- Progression - Capture
    total_capture_attempts INTEGER DEFAULT 0,
    successful_boards INTEGER DEFAULT 0,
    successful_captures INTEGER DEFAULT 0,

    -- Progression - Mining
    total_mining_ops INTEGER DEFAULT 0,
    total_yield BIGINT DEFAULT 0,

    -- Progression - Crafting
    crafting_skill_metalwork INTEGER DEFAULT 0,
    crafting_skill_electronics INTEGER DEFAULT 0,
    crafting_skill_weapons INTEGER DEFAULT 0,
    crafting_skill_propulsion INTEGER DEFAULT 0,

    -- Progression - Research
    research_points INTEGER DEFAULT 0,

    -- Progression - Overall
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,

    -- Legal status
    legal_status VARCHAR(20) DEFAULT 'citizen',
    bounty BIGINT DEFAULT 0,

    -- Social / Profile
    bio TEXT DEFAULT '',
    join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_playtime INTEGER DEFAULT 0,  -- in seconds
    profile_privacy VARCHAR(20) DEFAULT 'public',  -- public, friends, private

    -- Status
    is_online BOOLEAN DEFAULT FALSE,
    is_criminal BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Metadata
    CONSTRAINT credits_non_negative CHECK (credits >= 0),
    CONSTRAINT level_range CHECK (level BETWEEN 1 AND 100),
    CONSTRAINT bounty_non_negative CHECK (bounty >= 0)
);

-- SSH public keys for authentication
CREATE TABLE IF NOT EXISTS player_ssh_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    key_type VARCHAR(20) NOT NULL,  -- rsa, ed25519, ecdsa, etc.
    public_key TEXT NOT NULL,        -- The actual public key
    fingerprint VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 fingerprint
    comment VARCHAR(255),            -- Optional comment from key
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT unique_player_key UNIQUE (player_id, fingerprint)
);

-- Player reputation with NPC factions
CREATE TABLE IF NOT EXISTS player_reputation (
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    faction_id VARCHAR(50) NOT NULL,
    reputation INTEGER DEFAULT 0,
    PRIMARY KEY (player_id, faction_id),
    CONSTRAINT reputation_range CHECK (reputation BETWEEN -100 AND 100)
);

-- Star systems
CREATE TABLE IF NOT EXISTS star_systems (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    pos_x INTEGER NOT NULL,
    pos_y INTEGER NOT NULL,
    government_id VARCHAR(50) NOT NULL,
    controlled_by_faction UUID,
    tech_level INTEGER DEFAULT 5,
    description TEXT,
    CONSTRAINT tech_level_range CHECK (tech_level BETWEEN 1 AND 10)
);

-- System connections (jump routes)
CREATE TABLE IF NOT EXISTS system_connections (
    system_a UUID REFERENCES star_systems(id) ON DELETE CASCADE,
    system_b UUID REFERENCES star_systems(id) ON DELETE CASCADE,
    PRIMARY KEY (system_a, system_b)
);

-- Planets
CREATE TABLE IF NOT EXISTS planets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    system_id UUID REFERENCES star_systems(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    x DOUBLE PRECISION DEFAULT 0,  -- X coordinate within system
    y DOUBLE PRECISION DEFAULT 0,  -- Y coordinate within system
    services TEXT[] DEFAULT '{}',  -- Array of service types
    population BIGINT DEFAULT 0,
    tech_level INTEGER DEFAULT 5,
    UNIQUE (system_id, name),
    CONSTRAINT tech_level_range CHECK (tech_level BETWEEN 1 AND 10)
);

-- Ships
CREATE TABLE IF NOT EXISTS ships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID REFERENCES players(id) ON DELETE CASCADE,
    type_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,

    -- Status
    hull INTEGER NOT NULL,
    shields INTEGER NOT NULL,
    fuel INTEGER NOT NULL,
    crew INTEGER NOT NULL,

    CONSTRAINT hull_positive CHECK (hull >= 0),
    CONSTRAINT shields_non_negative CHECK (shields >= 0),
    CONSTRAINT fuel_non_negative CHECK (fuel >= 0),
    CONSTRAINT crew_positive CHECK (crew > 0)
);

-- Ship cargo
CREATE TABLE IF NOT EXISTS ship_cargo (
    ship_id UUID REFERENCES ships(id) ON DELETE CASCADE,
    commodity_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (ship_id, commodity_id),
    CONSTRAINT quantity_positive CHECK (quantity > 0)
);

-- Ship weapons
CREATE TABLE IF NOT EXISTS ship_weapons (
    ship_id UUID REFERENCES ships(id) ON DELETE CASCADE,
    weapon_id VARCHAR(50) NOT NULL,
    slot_index INTEGER NOT NULL,
    current_ammo INTEGER DEFAULT 0,
    PRIMARY KEY (ship_id, slot_index),
    CONSTRAINT ammo_non_negative CHECK (current_ammo >= 0)
);

-- Ship outfits
CREATE TABLE IF NOT EXISTS ship_outfits (
    ship_id UUID REFERENCES ships(id) ON DELETE CASCADE,
    outfit_id VARCHAR(50) NOT NULL
);

-- Player factions
CREATE TABLE IF NOT EXISTS player_factions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    tag VARCHAR(4) UNIQUE NOT NULL,
    founder_id UUID REFERENCES players(id),
    leader_id UUID REFERENCES players(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Resources
    treasury BIGINT DEFAULT 0,

    -- Territory
    home_system UUID REFERENCES star_systems(id),

    -- Progression
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,

    -- Properties
    alignment VARCHAR(20) NOT NULL,
    is_recruiting BOOLEAN DEFAULT FALSE,
    tax_rate DECIMAL(4,3) DEFAULT 0.05,
    member_limit INTEGER DEFAULT 10,

    -- Settings
    settings JSONB DEFAULT '{}',

    CONSTRAINT treasury_non_negative CHECK (treasury >= 0),
    CONSTRAINT level_range CHECK (level BETWEEN 1 AND 10),
    CONSTRAINT tax_rate_range CHECK (tax_rate BETWEEN 0 AND 1)
);

-- Faction members
CREATE TABLE IF NOT EXISTS faction_members (
    faction_id UUID REFERENCES player_factions(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    rank VARCHAR(20) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    contribution BIGINT DEFAULT 0,
    PRIMARY KEY (faction_id, player_id)
);

-- Faction officers
CREATE TABLE IF NOT EXISTS faction_officers (
    faction_id UUID REFERENCES player_factions(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    PRIMARY KEY (faction_id, player_id)
);

-- Faction reputation with NPC governments
CREATE TABLE IF NOT EXISTS faction_reputation (
    faction_id UUID REFERENCES player_factions(id) ON DELETE CASCADE,
    government_id VARCHAR(50) NOT NULL,
    reputation INTEGER DEFAULT 0,
    PRIMARY KEY (faction_id, government_id),
    CONSTRAINT reputation_range CHECK (reputation BETWEEN -100 AND 100)
);

-- Market prices
CREATE TABLE IF NOT EXISTS market_prices (
    planet_id UUID REFERENCES planets(id) ON DELETE CASCADE,
    commodity_id VARCHAR(50) NOT NULL,
    buy_price BIGINT NOT NULL,
    sell_price BIGINT NOT NULL,
    stock INTEGER DEFAULT 0,
    demand INTEGER DEFAULT 0,
    last_update BIGINT NOT NULL,
    PRIMARY KEY (planet_id, commodity_id),
    CONSTRAINT prices_positive CHECK (buy_price >= 0 AND sell_price >= 0)
);

-- Missions
CREATE TABLE IF NOT EXISTS missions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,

    -- Giver
    giver_id VARCHAR(100) NOT NULL,
    origin_planet UUID REFERENCES planets(id),

    -- Objectives
    destination UUID,
    target VARCHAR(50),
    quantity INTEGER DEFAULT 0,

    -- Rewards
    reward BIGINT NOT NULL,
    reputation_changes JSONB DEFAULT '{}',

    -- Timing
    deadline TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- State
    status VARCHAR(20) DEFAULT 'available',
    progress INTEGER DEFAULT 0,

    -- Requirements
    min_combat_rating INTEGER DEFAULT 0,
    required_rep JSONB DEFAULT '{}'
);

-- Player missions (active missions)
CREATE TABLE IF NOT EXISTS player_missions (
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    accepted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    progress INTEGER DEFAULT 0,
    PRIMARY KEY (player_id, mission_id)
);

-- Chat messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    sender_id UUID REFERENCES players(id) ON DELETE SET NULL,
    channel VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Events log (for universe events, combat, etc.)
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

-- Player mail system
CREATE TABLE IF NOT EXISTS player_mail (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_player UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    to_player UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    deleted_by UUID[] DEFAULT '{}',  -- Array of player IDs who deleted this mail
    CONSTRAINT subject_not_empty CHECK (char_length(subject) > 0),
    CONSTRAINT body_not_empty CHECK (char_length(body) > 0)
);

-- Shared loadouts system
CREATE TABLE IF NOT EXISTS shared_loadouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ship_type_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    weapons JSONB NOT NULL DEFAULT '[]',
    outfits JSONB NOT NULL DEFAULT '[]',
    stats JSONB NOT NULL DEFAULT '{}',
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    views INTEGER DEFAULT 0,
    favorites INTEGER DEFAULT 0,
    CONSTRAINT name_not_empty CHECK (char_length(name) > 0)
);

-- Loadout favorites (many-to-many)
CREATE TABLE IF NOT EXISTS loadout_favorites (
    loadout_id UUID NOT NULL REFERENCES shared_loadouts(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (loadout_id, player_id)
);

-- ============================================================================
-- Schema Migrations Tracking
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64)
);

CREATE INDEX idx_migrations_version ON schema_migrations(version);

-- ============================================================================
-- Social Features
-- ============================================================================

-- Friend relationships (bidirectional)
CREATE TABLE IF NOT EXISTS player_friends (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_friendship UNIQUE (player_id, friend_id),
    CONSTRAINT no_self_friendship CHECK (player_id != friend_id)
);

-- Friend requests
CREATE TABLE IF NOT EXISTS friend_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_friend_request UNIQUE (sender_id, receiver_id),
    CONSTRAINT no_self_request CHECK (sender_id != receiver_id)
);

-- Blocked players
CREATE TABLE IF NOT EXISTS player_blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blocker_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    reason VARCHAR(100) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_block UNIQUE (blocker_id, blocked_id),
    CONSTRAINT no_self_block CHECK (blocker_id != blocked_id)
);

-- Player notifications
CREATE TABLE IF NOT EXISTS player_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    related_player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    related_entity_type VARCHAR(50),
    related_entity_id UUID,
    is_read BOOLEAN DEFAULT FALSE,
    is_dismissed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),
    action_data JSONB DEFAULT '{}'
);

-- Indexes for social features
CREATE INDEX idx_player_friends_player ON player_friends(player_id);
CREATE INDEX idx_player_friends_friend ON player_friends(friend_id);
CREATE INDEX idx_friend_requests_sender ON friend_requests(sender_id);
CREATE INDEX idx_friend_requests_receiver ON friend_requests(receiver_id);
CREATE INDEX idx_friend_requests_status ON friend_requests(status);
CREATE INDEX idx_player_blocks_blocker ON player_blocks(blocker_id);
CREATE INDEX idx_player_blocks_blocked ON player_blocks(blocked_id);
CREATE INDEX idx_player_notifications_player ON player_notifications(player_id) WHERE is_dismissed = FALSE;
CREATE INDEX idx_player_notifications_unread ON player_notifications(player_id, is_read) WHERE is_dismissed = FALSE;
CREATE INDEX idx_player_notifications_type ON player_notifications(type);
CREATE INDEX idx_player_notifications_created ON player_notifications(created_at DESC);
CREATE INDEX idx_player_notifications_expires ON player_notifications(expires_at) WHERE is_dismissed = FALSE;

-- ============================================================================
-- Security V2 Features
-- ============================================================================

-- Account activity events
CREATE TABLE IF NOT EXISTS account_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    details JSONB,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'none',

    CONSTRAINT account_events_risk_level_check CHECK (risk_level IN ('none', 'low', 'medium', 'high', 'critical'))
);

CREATE INDEX idx_account_events_player_id ON account_events(player_id);
CREATE INDEX idx_account_events_timestamp ON account_events(timestamp DESC);
CREATE INDEX idx_account_events_risk_level ON account_events(risk_level) WHERE risk_level IN ('high', 'critical');
CREATE INDEX idx_account_events_event_type ON account_events(event_type);
CREATE INDEX idx_account_events_ip_address ON account_events(ip_address);

-- Two-factor authentication
CREATE TABLE IF NOT EXISTS player_two_factor (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID UNIQUE NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    secret VARCHAR(255) NOT NULL,
    backup_codes TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    recovery_email VARCHAR(255),

    CONSTRAINT player_two_factor_secret_check CHECK (length(secret) >= 16)
);

CREATE INDEX idx_player_two_factor_player_id ON player_two_factor(player_id);
CREATE INDEX idx_player_two_factor_enabled ON player_two_factor(enabled) WHERE enabled = TRUE;

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    ip_address VARCHAR(45) NOT NULL,

    CONSTRAINT password_reset_tokens_expiry_check CHECK (expires_at > created_at)
);

CREATE INDEX idx_password_reset_tokens_player_id ON password_reset_tokens(player_id);
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token) WHERE used = FALSE;
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- Admin IP whitelist
CREATE TABLE IF NOT EXISTS admin_ip_whitelist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    cidr_mask INTEGER DEFAULT 32,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES players(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT admin_ip_whitelist_unique UNIQUE (admin_id, ip_address),
    CONSTRAINT admin_ip_whitelist_cidr_check CHECK (cidr_mask >= 0 AND cidr_mask <= 32)
);

CREATE INDEX idx_admin_ip_whitelist_admin_id ON admin_ip_whitelist(admin_id);
CREATE INDEX idx_admin_ip_whitelist_active ON admin_ip_whitelist(is_active) WHERE is_active = TRUE;

-- Login history
CREATE TABLE IF NOT EXISTS login_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    country_code CHAR(2),
    city VARCHAR(255),
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(255),
    anomalies JSONB,
    risk_score INTEGER DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT login_history_risk_score_check CHECK (risk_score >= 0 AND risk_score <= 100)
);

CREATE INDEX idx_login_history_player_id ON login_history(player_id);
CREATE INDEX idx_login_history_timestamp ON login_history(timestamp DESC);
CREATE INDEX idx_login_history_ip_address ON login_history(ip_address);
CREATE INDEX idx_login_history_risk_score ON login_history(risk_score) WHERE risk_score > 50;

-- Player sessions
CREATE TABLE IF NOT EXISTS player_sessions (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT player_sessions_expiry_check CHECK (expires_at > started_at)
);

CREATE INDEX idx_player_sessions_player_id ON player_sessions(player_id);
CREATE INDEX idx_player_sessions_active ON player_sessions(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_player_sessions_expires_at ON player_sessions(expires_at);

-- Honeypot attempts tracking
CREATE TABLE IF NOT EXISTS honeypot_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username_attempted VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    autobanned BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT honeypot_attempts_unique UNIQUE (ip_address, username_attempted, timestamp)
);

CREATE INDEX idx_honeypot_attempts_ip_address ON honeypot_attempts(ip_address);
CREATE INDEX idx_honeypot_attempts_timestamp ON honeypot_attempts(timestamp DESC);
CREATE INDEX idx_honeypot_attempts_autobanned ON honeypot_attempts(autobanned) WHERE autobanned = TRUE;

-- Rate limiting tracking
CREATE TABLE IF NOT EXISTS rate_limit_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    action_count INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    window_end TIMESTAMP NOT NULL,
    blocked BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT rate_limit_tracking_window_check CHECK (window_end > window_start),
    CONSTRAINT rate_limit_tracking_unique UNIQUE (player_id, ip_address, action_type, window_start)
);

CREATE INDEX idx_rate_limit_tracking_player_id ON rate_limit_tracking(player_id);
CREATE INDEX idx_rate_limit_tracking_ip_address ON rate_limit_tracking(ip_address);
CREATE INDEX idx_rate_limit_tracking_action_type ON rate_limit_tracking(action_type);
CREATE INDEX idx_rate_limit_tracking_window_end ON rate_limit_tracking(window_end);

-- Player security settings
CREATE TABLE IF NOT EXISTS player_security_settings (
    player_id UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    login_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    new_ip_email_alert BOOLEAN NOT NULL DEFAULT TRUE,
    session_timeout_minutes INTEGER NOT NULL DEFAULT 15,
    require_2fa BOOLEAN NOT NULL DEFAULT FALSE,
    allow_password_reset_email BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT player_security_settings_timeout_check CHECK (session_timeout_minutes >= 5 AND session_timeout_minutes <= 1440)
);

-- Trusted devices
CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    trusted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT trusted_devices_unique UNIQUE (player_id, device_fingerprint),
    CONSTRAINT trusted_devices_expiry_check CHECK (expires_at > trusted_at)
);

CREATE INDEX idx_trusted_devices_player_id ON trusted_devices(player_id);
CREATE INDEX idx_trusted_devices_active ON trusted_devices(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_trusted_devices_expires_at ON trusted_devices(expires_at);

-- Indexes for performance
CREATE INDEX idx_players_username ON players(username);
CREATE INDEX idx_players_username_lower ON players (LOWER(username));
CREATE INDEX idx_players_email ON players(email) WHERE email IS NOT NULL;
CREATE INDEX idx_players_online ON players(is_online);
CREATE INDEX idx_ssh_keys_player ON player_ssh_keys(player_id);
CREATE INDEX idx_ssh_keys_fingerprint ON player_ssh_keys(fingerprint);
CREATE INDEX idx_ssh_keys_active ON player_ssh_keys(player_id, is_active);
CREATE INDEX idx_systems_position ON star_systems(pos_x, pos_y);
CREATE INDEX idx_planets_system ON planets(system_id);
CREATE INDEX idx_ships_owner ON ships(owner_id);
CREATE INDEX idx_faction_members_player ON faction_members(player_id);
CREATE INDEX idx_chat_channel ON chat_messages(channel, created_at DESC);
CREATE INDEX idx_events_type ON events(type, created_at DESC);
CREATE INDEX idx_missions_status ON missions(status);
CREATE INDEX idx_admin_users_player ON admin_users(player_id);
CREATE INDEX idx_admin_users_active ON admin_users(is_active);
CREATE INDEX idx_player_bans_player ON player_bans(player_id);
CREATE INDEX idx_player_bans_active ON player_bans(is_active, expires_at);
CREATE INDEX idx_player_mutes_player ON player_mutes(player_id);
CREATE INDEX idx_player_mutes_active ON player_mutes(is_active, expires_at);
CREATE INDEX idx_admin_actions_admin ON admin_actions(admin_id, timestamp DESC);
CREATE INDEX idx_admin_actions_timestamp ON admin_actions(timestamp DESC);
CREATE INDEX idx_mail_recipient ON player_mail(to_player, sent_at DESC);
CREATE INDEX idx_mail_sender ON player_mail(from_player, sent_at DESC);
CREATE INDEX idx_mail_unread ON player_mail(to_player, read, sent_at DESC);
CREATE INDEX idx_loadouts_player ON shared_loadouts(player_id, updated_at DESC);
CREATE INDEX idx_loadouts_public ON shared_loadouts(is_public, created_at DESC) WHERE is_public = true;
CREATE INDEX idx_loadouts_ship_type ON shared_loadouts(ship_type_id, is_public, created_at DESC) WHERE is_public = true;

-- Performance optimization indexes (added 2025-11-15)
-- Market prices indexes (heavily queried for trading)
CREATE INDEX idx_market_planet ON market_prices(planet_id);
CREATE INDEX idx_market_planet_commodity ON market_prices(planet_id, commodity_id);
CREATE INDEX idx_market_updated ON market_prices(last_update DESC);

-- Ship cargo indexes (frequently accessed during trading/combat)
CREATE INDEX idx_ship_cargo_ship ON ship_cargo(ship_id);
CREATE INDEX idx_ship_cargo_composite ON ship_cargo(ship_id, commodity_id);

-- Player location indexes (for presence/multiplayer features)
CREATE INDEX idx_players_current_system ON players(current_system) WHERE current_system IS NOT NULL;
CREATE INDEX idx_players_current_planet ON players(current_planet) WHERE current_planet IS NOT NULL;
CREATE INDEX idx_players_ship ON players(ship_id) WHERE ship_id IS NOT NULL;

-- Ship weapons and outfits indexes
CREATE INDEX idx_ship_weapons_ship ON ship_weapons(ship_id);
CREATE INDEX idx_ship_outfits_ship ON ship_outfits(ship_id);

-- System connections indexes (for navigation pathfinding)
CREATE INDEX idx_system_connections_a ON system_connections(system_a);
CREATE INDEX idx_system_connections_b ON system_connections(system_b);

-- Faction members indexes (for faction queries)
CREATE INDEX idx_faction_members_faction ON faction_members(faction_id);

-- Player reputation indexes (for NPC interactions)
CREATE INDEX idx_player_reputation_player ON player_reputation(player_id);

-- Composite indexes for common join patterns
CREATE INDEX idx_ships_owner_type ON ships(owner_id, type_id);
CREATE INDEX idx_planets_system_tech ON planets(system_id, tech_level);
CREATE INDEX idx_loadouts_popular ON shared_loadouts((favorites * 2 + views) DESC, is_public) WHERE is_public = true;
CREATE INDEX idx_loadout_favorites_player ON loadout_favorites(player_id, created_at DESC);

-- Update foreign key for controlled systems
ALTER TABLE star_systems
ADD CONSTRAINT fk_controlled_by_faction
FOREIGN KEY (controlled_by_faction)
REFERENCES player_factions(id) ON DELETE SET NULL;

-- Add foreign key for player's faction
ALTER TABLE players
ADD COLUMN faction_id UUID REFERENCES player_factions(id) ON DELETE SET NULL,
ADD COLUMN faction_rank VARCHAR(20);

-- Player Items (UUID-based inventory for equipment/weapons)
CREATE TABLE IF NOT EXISTS player_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Item type and reference
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('weapon', 'outfit', 'special', 'quest')),
    equipment_id VARCHAR(100) NOT NULL, -- References equipment definition (e.g., "laser_cannon")

    -- Current location
    location VARCHAR(50) NOT NULL CHECK (location IN ('ship', 'station_storage', 'mail', 'escrow', 'auction')),
    location_id UUID, -- ship_id, planet_id, mail_id, auction_id, etc.

    -- Item properties (for modifications, upgrades, etc.)
    properties JSONB DEFAULT '{}',

    -- Metadata
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Item Transfers (audit trail for item movements)
CREATE TABLE IF NOT EXISTS item_transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL REFERENCES player_items(id) ON DELETE CASCADE,

    from_player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    to_player_id UUID REFERENCES players(id) ON DELETE SET NULL,

    transfer_type VARCHAR(50) NOT NULL CHECK (transfer_type IN ('trade', 'mail', 'auction', 'contract', 'admin')),
    transfer_id UUID, -- trade_id, mail_id, auction_id, etc.

    -- Metadata
    transferred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for player_items
CREATE INDEX idx_player_items_player ON player_items(player_id);
CREATE INDEX idx_player_items_location ON player_items(location, location_id);
CREATE INDEX idx_player_items_type ON player_items(item_type, equipment_id);

-- Indexes for item_transfers
CREATE INDEX idx_item_transfers_item ON item_transfers(item_id);
CREATE INDEX idx_item_transfers_players ON item_transfers(from_player_id, to_player_id);
CREATE INDEX idx_item_transfers_type ON item_transfers(transfer_type, transfer_id);

-- Comments
COMMENT ON TABLE players IS 'Player accounts and game state';
COMMENT ON TABLE player_ssh_keys IS 'SSH public keys for player authentication';
COMMENT ON TABLE player_reputation IS 'Player reputation with NPC factions';
COMMENT ON TABLE star_systems IS 'Star systems in the universe';
COMMENT ON TABLE system_connections IS 'Jump routes between star systems';
COMMENT ON TABLE planets IS 'Planets and stations';
COMMENT ON TABLE ships IS 'Player and NPC ships';
COMMENT ON TABLE ship_cargo IS 'Ship cargo inventory (commodity-based)';
COMMENT ON TABLE ship_weapons IS 'Ship equipped weapons';
COMMENT ON TABLE ship_outfits IS 'Ship equipped outfits';
COMMENT ON TABLE player_factions IS 'Player-created factions/guilds';
COMMENT ON TABLE faction_members IS 'Faction membership tracking';
COMMENT ON TABLE faction_officers IS 'Faction officers and ranks';
COMMENT ON TABLE faction_reputation IS 'Faction reputation with other factions';
COMMENT ON TABLE market_prices IS 'Commodity prices at each planet';
COMMENT ON TABLE missions IS 'Available and active missions';
COMMENT ON TABLE player_missions IS 'Player active missions tracking';
COMMENT ON TABLE chat_messages IS 'In-game chat history';
COMMENT ON TABLE events IS 'Game events log for analytics';
COMMENT ON TABLE admin_users IS 'Server administrators and moderators';
COMMENT ON TABLE player_bans IS 'Banned players with expiration tracking';
COMMENT ON TABLE player_mutes IS 'Muted players with expiration tracking';
COMMENT ON TABLE admin_actions IS 'Audit log of all admin actions';
COMMENT ON TABLE server_settings IS 'Server configuration (single row)';
COMMENT ON TABLE player_mail IS 'Player-to-player mail messages with soft delete';
COMMENT ON TABLE shared_loadouts IS 'Shared ship loadout configurations with stats tracking';
COMMENT ON TABLE loadout_favorites IS 'Player favorites for shared loadouts';
COMMENT ON TABLE schema_migrations IS 'Tracks applied database migrations';
COMMENT ON TABLE player_friends IS 'Friend relationships between players';
COMMENT ON TABLE friend_requests IS 'Pending friend requests';
COMMENT ON TABLE player_blocks IS 'Blocked players list';
COMMENT ON TABLE player_notifications IS 'Player notifications system';
COMMENT ON TABLE account_events IS 'Security-relevant account activities for audit trail';
COMMENT ON TABLE player_two_factor IS 'Two-factor authentication configuration';
COMMENT ON TABLE password_reset_tokens IS 'Password reset tokens with expiration';
COMMENT ON TABLE admin_ip_whitelist IS 'IP whitelist for admin account access restriction';
COMMENT ON TABLE login_history IS 'Detailed login history for anomaly detection';
COMMENT ON TABLE player_sessions IS 'Active player sessions for concurrent session management';
COMMENT ON TABLE honeypot_attempts IS 'Honeypot account access attempts tracking';
COMMENT ON TABLE rate_limit_tracking IS 'Action-based rate limiting per player/IP';
COMMENT ON TABLE player_security_settings IS 'Per-player security preferences';
COMMENT ON TABLE trusted_devices IS 'Trusted devices for streamlined authentication';
COMMENT ON TABLE player_items IS 'UUID-based inventory for weapons, outfits, and special items';
COMMENT ON TABLE item_transfers IS 'Audit log of all item movements between players';
