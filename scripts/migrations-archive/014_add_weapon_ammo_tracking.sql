-- File: scripts/migrations/014_add_weapon_ammo_tracking.sql
-- Project: Terminal Velocity
-- Description: Add ammo tracking for ship weapons
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-11-15

-- Add current_ammo column to ship_weapons table
ALTER TABLE ship_weapons ADD COLUMN IF NOT EXISTS current_ammo INTEGER DEFAULT 0;

-- Add constraint to ensure ammo is non-negative
ALTER TABLE ship_weapons ADD CONSTRAINT IF NOT EXISTS ammo_non_negative CHECK (current_ammo >= 0);

-- Initialize ammo for existing weapons (set to 0, will be loaded on next restock)
UPDATE ship_weapons SET current_ammo = 0 WHERE current_ammo IS NULL;

-- Record migration
INSERT INTO schema_migrations (version, name, checksum)
VALUES (14, 'add_weapon_ammo_tracking', 'f9e4a7c2b8d5f1a3')
ON CONFLICT (version) DO NOTHING;

COMMENT ON COLUMN ship_weapons.current_ammo IS 'Current ammunition count for this weapon (0 for energy weapons)';
