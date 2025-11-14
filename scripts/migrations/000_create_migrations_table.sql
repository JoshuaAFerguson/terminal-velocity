-- File: scripts/migrations/000_create_migrations_table.sql
-- Project: Terminal Velocity
-- Description: Create migrations tracking table
-- Version: 1.0.0
-- Author: Joshua Ferguson
-- Created: 2025-01-14

-- Migrations tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version INTEGER UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64)
);

CREATE INDEX idx_migrations_version ON schema_migrations(version);

COMMENT ON TABLE schema_migrations IS 'Tracks applied database migrations';
