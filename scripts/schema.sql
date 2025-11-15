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

    -- Progression - Overall
    level INTEGER DEFAULT 1,
    experience BIGINT DEFAULT 0,

    -- Legal status
    legal_status VARCHAR(20) DEFAULT 'citizen',
    bounty BIGINT DEFAULT 0,

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
    PRIMARY KEY (ship_id, slot_index)
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

-- Indexes for performance
CREATE INDEX idx_players_username ON players(username);
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

-- Comments
COMMENT ON TABLE players IS 'Player accounts and game state';
COMMENT ON TABLE player_ssh_keys IS 'SSH public keys for player authentication';
COMMENT ON TABLE star_systems IS 'Star systems in the universe';
COMMENT ON TABLE planets IS 'Planets and stations';
COMMENT ON TABLE ships IS 'Player and NPC ships';
COMMENT ON TABLE player_factions IS 'Player-created factions/guilds';
COMMENT ON TABLE market_prices IS 'Commodity prices at each planet';
COMMENT ON TABLE missions IS 'Available and active missions';
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
