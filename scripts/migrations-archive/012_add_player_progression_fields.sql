-- File: scripts/migrations/012_add_player_progression_fields.sql
-- Project: Terminal Velocity
-- Description: Add player progression and legal status fields
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-11-15

-- Add trading progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS trading_rating INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_trades INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS trade_profit BIGINT DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS highest_profit BIGINT DEFAULT 0;

-- Add exploration progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS exploration_rating INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS systems_visited INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_jumps INTEGER DEFAULT 0;

-- Add mission progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS missions_completed INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS missions_failed INTEGER DEFAULT 0;

-- Add quest progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS quests_completed INTEGER DEFAULT 0;

-- Add overall progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 1;
ALTER TABLE players ADD COLUMN IF NOT EXISTS experience BIGINT DEFAULT 0;

-- Add legal status fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS legal_status VARCHAR(20) DEFAULT 'citizen';
ALTER TABLE players ADD COLUMN IF NOT EXISTS bounty BIGINT DEFAULT 0;

-- Add updated_at timestamp
ALTER TABLE players ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add constraints
ALTER TABLE players ADD CONSTRAINT IF NOT EXISTS level_range CHECK (level BETWEEN 1 AND 100);
ALTER TABLE players ADD CONSTRAINT IF NOT EXISTS bounty_non_negative CHECK (bounty >= 0);

-- Update existing players to have default values
UPDATE players SET
    trading_rating = 0,
    total_trades = 0,
    trade_profit = 0,
    highest_profit = 0,
    exploration_rating = 0,
    systems_visited = 0,
    total_jumps = 0,
    missions_completed = 0,
    missions_failed = 0,
    quests_completed = 0,
    level = 1,
    experience = 0,
    legal_status = 'citizen',
    bounty = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE trading_rating IS NULL;

-- Record migration
INSERT INTO schema_migrations (version, name, checksum)
VALUES (12, 'add_player_progression_fields', 'a8f7d3c9e4b2f1a6')
ON CONFLICT (version) DO NOTHING;

COMMENT ON COLUMN players.trading_rating IS 'Trading skill rating (0-100)';
COMMENT ON COLUMN players.total_trades IS 'Total number of trades completed';
COMMENT ON COLUMN players.trade_profit IS 'Total profit from all trades';
COMMENT ON COLUMN players.highest_profit IS 'Highest single trade profit';
COMMENT ON COLUMN players.exploration_rating IS 'Exploration skill rating (0-100)';
COMMENT ON COLUMN players.systems_visited IS 'Number of unique systems visited';
COMMENT ON COLUMN players.total_jumps IS 'Total hyperspace jumps made';
COMMENT ON COLUMN players.missions_completed IS 'Total missions completed';
COMMENT ON COLUMN players.missions_failed IS 'Total missions failed';
COMMENT ON COLUMN players.quests_completed IS 'Total quests completed';
COMMENT ON COLUMN players.level IS 'Overall player level (1-100)';
COMMENT ON COLUMN players.experience IS 'Experience points for leveling';
COMMENT ON COLUMN players.legal_status IS 'Legal status: citizen, outlaw, pirate, wanted, hostile';
COMMENT ON COLUMN players.bounty IS 'Bounty on player head (credits)';
COMMENT ON COLUMN players.updated_at IS 'Last update timestamp';
