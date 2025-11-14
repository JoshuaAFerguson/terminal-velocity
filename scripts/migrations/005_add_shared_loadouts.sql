-- Migration: 005_add_shared_loadouts.sql
-- Description: Add ship loadout sharing and comparison system
-- Date: 2025-01-14

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

-- Indexes for loadouts
CREATE INDEX IF NOT EXISTS idx_loadouts_player ON shared_loadouts(player_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_loadouts_public ON shared_loadouts(is_public, created_at DESC) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_loadouts_ship_type ON shared_loadouts(ship_type_id, is_public, created_at DESC) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_loadouts_popular ON shared_loadouts((favorites * 2 + views) DESC, is_public) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_loadout_favorites_player ON loadout_favorites(player_id, created_at DESC);

-- Comments
COMMENT ON TABLE shared_loadouts IS 'Shared ship loadout configurations with stats tracking';
COMMENT ON TABLE loadout_favorites IS 'Player favorites for shared loadouts';
