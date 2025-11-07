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
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Game state
    credits BIGINT DEFAULT 10000,
    current_system UUID,
    current_planet UUID,
    ship_id UUID,

    -- Progression
    combat_rating INTEGER DEFAULT 0,
    total_kills INTEGER DEFAULT 0,
    play_time BIGINT DEFAULT 0,

    -- Status
    is_online BOOLEAN DEFAULT FALSE,
    is_criminal BOOLEAN DEFAULT FALSE,

    -- Metadata
    CONSTRAINT credits_non_negative CHECK (credits >= 0)
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

-- Indexes for performance
CREATE INDEX idx_players_username ON players(username);
CREATE INDEX idx_players_online ON players(is_online);
CREATE INDEX idx_systems_position ON star_systems(pos_x, pos_y);
CREATE INDEX idx_planets_system ON planets(system_id);
CREATE INDEX idx_ships_owner ON ships(owner_id);
CREATE INDEX idx_faction_members_player ON faction_members(player_id);
CREATE INDEX idx_chat_channel ON chat_messages(channel, created_at DESC);
CREATE INDEX idx_events_type ON events(type, created_at DESC);
CREATE INDEX idx_missions_status ON missions(status);

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
COMMENT ON TABLE star_systems IS 'Star systems in the universe';
COMMENT ON TABLE planets IS 'Planets and stations';
COMMENT ON TABLE ships IS 'Player and NPC ships';
COMMENT ON TABLE player_factions IS 'Player-created factions/guilds';
COMMENT ON TABLE market_prices IS 'Commodity prices at each planet';
COMMENT ON TABLE missions IS 'Available and active missions';
COMMENT ON TABLE chat_messages IS 'In-game chat history';
COMMENT ON TABLE events IS 'Game events log for analytics';
