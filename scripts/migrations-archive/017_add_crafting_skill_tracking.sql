-- Migration: 017_add_crafting_skill_tracking
-- Description: Add crafting_skill and total_crafts fields to players table for manufacturing system
-- Date: 2025-11-15

-- Add crafting_skill column to track player's crafting skill level (0-100)
ALTER TABLE players
ADD COLUMN IF NOT EXISTS crafting_skill INTEGER DEFAULT 0 CHECK (crafting_skill >= 0 AND crafting_skill <= 100);

-- Add total_crafts column to track total items crafted
ALTER TABLE players
ADD COLUMN IF NOT EXISTS total_crafts INTEGER DEFAULT 0 CHECK (total_crafts >= 0);

-- Add comments
COMMENT ON COLUMN players.crafting_skill IS 'Player crafting skill level (0-100), affects crafting time and unlocks blueprints';
COMMENT ON COLUMN players.total_crafts IS 'Total number of items successfully crafted by the player';
