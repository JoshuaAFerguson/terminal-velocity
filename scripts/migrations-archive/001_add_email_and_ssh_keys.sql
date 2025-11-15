-- Migration: Add email and SSH key authentication support
-- Date: 2025-01-06

-- Add email to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS email VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS idx_players_email ON players(email) WHERE email IS NOT NULL;

-- Make password_hash nullable (users can auth with SSH keys only)
ALTER TABLE players ALTER COLUMN password_hash DROP NOT NULL;

-- Add email verification fields
ALTER TABLE players ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE players ADD COLUMN IF NOT EXISTS email_verification_token VARCHAR(64);

-- SSH public keys table
CREATE TABLE IF NOT EXISTS player_ssh_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    key_type VARCHAR(20) NOT NULL,  -- rsa, ed25519, ecdsa, etc.
    public_key TEXT NOT NULL,        -- The actual public key
    fingerprint VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 fingerprint
    comment VARCHAR(255),            -- Optional comment from key
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT unique_player_key UNIQUE (player_id, fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_ssh_keys_player ON player_ssh_keys(player_id);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON player_ssh_keys(fingerprint);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_active ON player_ssh_keys(player_id, is_active);

COMMENT ON TABLE player_ssh_keys IS 'SSH public keys for player authentication';
COMMENT ON COLUMN player_ssh_keys.fingerprint IS 'SHA256 fingerprint of the public key';
