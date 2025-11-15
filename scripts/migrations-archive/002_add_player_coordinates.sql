-- Migration: Add position coordinates to players table
-- Date: 2025-01-14
-- Description: Adds X and Y coordinate fields for player position tracking within star systems

-- Add position coordinates to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS x DOUBLE PRECISION DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS y DOUBLE PRECISION DEFAULT 0;

-- Create index for spatial queries (useful for finding nearby players)
CREATE INDEX IF NOT EXISTS idx_players_position ON players(current_system, x, y);

COMMENT ON COLUMN players.x IS 'X coordinate of player position within current system';
COMMENT ON COLUMN players.y IS 'Y coordinate of player position within current system';
