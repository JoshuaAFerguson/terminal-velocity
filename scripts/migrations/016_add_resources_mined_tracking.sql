-- Migration: 016_add_resources_mined_tracking
-- Description: Add resources_mined field to players table to track mining statistics by resource type
-- Date: 2025-11-15

-- Add resources_mined column to track which resources have been mined
ALTER TABLE players
ADD COLUMN IF NOT EXISTS resources_mined JSONB DEFAULT '{}'::jsonb;

-- Create index on resources_mined for better query performance
CREATE INDEX IF NOT EXISTS idx_players_resources_mined ON players USING gin (resources_mined);

-- Add comment
COMMENT ON COLUMN players.resources_mined IS 'Tracks the quantity of each resource type mined by the player';
