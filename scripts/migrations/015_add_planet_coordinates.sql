-- File: scripts/migrations/015_add_planet_coordinates.sql
-- Project: Terminal Velocity
-- Description: Add X/Y coordinates to planets for distance checking
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-11-15

-- Add X and Y coordinates to planets table
ALTER TABLE planets ADD COLUMN IF NOT EXISTS x DOUBLE PRECISION DEFAULT 0;
ALTER TABLE planets ADD COLUMN IF NOT EXISTS y DOUBLE PRECISION DEFAULT 0;

-- Initialize coordinates for existing planets (random positions within system)
-- Planets will be placed at random coordinates between -1000 and 1000
UPDATE planets SET
    x = (RANDOM() * 2000 - 1000),
    y = (RANDOM() * 2000 - 1000)
WHERE x = 0 AND y = 0;

-- Record migration
INSERT INTO schema_migrations (version, name, checksum)
VALUES (15, 'add_planet_coordinates', 'd7b3e9f4c8a5e2b1')
ON CONFLICT (version) DO NOTHING;

COMMENT ON COLUMN planets.x IS 'X coordinate of planet within its star system';
COMMENT ON COLUMN planets.y IS 'Y coordinate of planet within its star system';
