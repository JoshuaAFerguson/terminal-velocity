-- Migration: 018_add_research_points
-- Description: Add research_points field to players table for technology research system
-- Date: 2025-11-15

-- Add research_points column to track player's available research points
ALTER TABLE players
ADD COLUMN IF NOT EXISTS research_points INTEGER DEFAULT 100 CHECK (research_points >= 0);

-- Add comment
COMMENT ON COLUMN players.research_points IS 'Available research points for unlocking technologies in manufacturing system';
