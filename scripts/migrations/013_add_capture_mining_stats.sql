-- File: scripts/migrations/013_add_capture_mining_stats.sql
-- Project: Terminal Velocity
-- Description: Add capture and mining statistics tracking
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-11-15

-- Add capture progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_capture_attempts INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS successful_boards INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS successful_captures INTEGER DEFAULT 0;

-- Add mining progression fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_mining_ops INTEGER DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_yield BIGINT DEFAULT 0;

-- Update existing players to have default values
UPDATE players SET
    total_capture_attempts = 0,
    successful_boards = 0,
    successful_captures = 0,
    total_mining_ops = 0,
    total_yield = 0
WHERE total_capture_attempts IS NULL;

-- Record migration
INSERT INTO schema_migrations (version, name, checksum)
VALUES (13, 'add_capture_mining_stats', 'c3f5e8a2d9b7f4c1')
ON CONFLICT (version) DO NOTHING;

COMMENT ON COLUMN players.total_capture_attempts IS 'Total ship boarding attempts';
COMMENT ON COLUMN players.successful_boards IS 'Successful ship boardings';
COMMENT ON COLUMN players.successful_captures IS 'Ships successfully captured';
COMMENT ON COLUMN players.total_mining_ops IS 'Total mining operations';
COMMENT ON COLUMN players.total_yield IS 'Total resources mined across all operations';
