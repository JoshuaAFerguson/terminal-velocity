-- Migration: Social Features (Phase 9)
-- Version: 011
-- Description: Add friends, blocks, profiles, mail, and notifications
-- Author: Claude Code
-- Created: 2025-11-15

-- ============================================================================
-- Player Profile Extensions
-- ============================================================================

-- Add profile fields to players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT '';
ALTER TABLE players ADD COLUMN IF NOT EXISTS join_date TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE players ADD COLUMN IF NOT EXISTS total_playtime INTEGER DEFAULT 0; -- in seconds
ALTER TABLE players ADD COLUMN IF NOT EXISTS profile_privacy VARCHAR(20) DEFAULT 'public'; -- public, friends, private

-- Add index for profile lookups
CREATE INDEX IF NOT EXISTS idx_players_username_lower ON players (LOWER(username));

-- ============================================================================
-- Friends System
-- ============================================================================

-- Friend relationships (bidirectional - both players must be friends)
CREATE TABLE IF NOT EXISTS player_friends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Ensure unique friendship
    CONSTRAINT unique_friendship UNIQUE (player_id, friend_id),
    -- Prevent self-friendship
    CONSTRAINT no_self_friendship CHECK (player_id != friend_id)
);

-- Friend requests (pending friendships)
CREATE TABLE IF NOT EXISTS friend_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, declined
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Ensure unique request
    CONSTRAINT unique_friend_request UNIQUE (sender_id, receiver_id),
    -- Prevent self-request
    CONSTRAINT no_self_request CHECK (sender_id != receiver_id)
);

-- Indexes for friends
CREATE INDEX IF NOT EXISTS idx_player_friends_player ON player_friends(player_id);
CREATE INDEX IF NOT EXISTS idx_player_friends_friend ON player_friends(friend_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_sender ON friend_requests(sender_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver ON friend_requests(receiver_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_status ON friend_requests(status);

-- ============================================================================
-- Block/Ignore System
-- ============================================================================

-- Blocked players (unidirectional - blocker blocks blockee)
CREATE TABLE IF NOT EXISTS player_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    reason VARCHAR(100) DEFAULT '', -- optional reason
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Ensure unique block
    CONSTRAINT unique_block UNIQUE (blocker_id, blocked_id),
    -- Prevent self-block
    CONSTRAINT no_self_block CHECK (blocker_id != blocked_id)
);

-- Indexes for blocks
CREATE INDEX IF NOT EXISTS idx_player_blocks_blocker ON player_blocks(blocker_id);
CREATE INDEX IF NOT EXISTS idx_player_blocks_blocked ON player_blocks(blocked_id);

-- ============================================================================
-- Mail System
-- ============================================================================

-- Persistent player mail
CREATE TABLE IF NOT EXISTS player_mail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID REFERENCES players(id) ON DELETE SET NULL, -- NULL if sender deleted
    sender_name VARCHAR(50) NOT NULL, -- Store name in case sender deleted
    receiver_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    subject VARCHAR(100) NOT NULL,
    body TEXT NOT NULL,

    -- Attachments (optional)
    attached_credits BIGINT DEFAULT 0,
    attached_items JSONB DEFAULT '[]', -- Array of item IDs

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE, -- Soft delete

    -- Timestamps
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    read_at TIMESTAMPTZ,

    -- Prevent spam
    CONSTRAINT valid_credits CHECK (attached_credits >= 0)
);

-- Indexes for mail
CREATE INDEX IF NOT EXISTS idx_player_mail_receiver ON player_mail(receiver_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_player_mail_unread ON player_mail(receiver_id, is_read) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_player_mail_sent_at ON player_mail(sent_at DESC);

-- ============================================================================
-- Notification System
-- ============================================================================

-- Player notifications
CREATE TABLE IF NOT EXISTS player_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Notification details
    type VARCHAR(50) NOT NULL, -- friend_request, mail, trade_offer, pvp_challenge, territory_attack, faction_invite
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,

    -- Related entities (optional, for clickable notifications)
    related_player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    related_entity_type VARCHAR(50), -- mail, trade, pvp, territory, faction
    related_entity_id UUID,

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    is_dismissed BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    read_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '7 days'), -- Auto-cleanup after 7 days

    -- Action data (JSON for flexibility)
    action_data JSONB DEFAULT '{}'
);

-- Indexes for notifications
CREATE INDEX IF NOT EXISTS idx_player_notifications_player ON player_notifications(player_id) WHERE is_dismissed = FALSE;
CREATE INDEX IF NOT EXISTS idx_player_notifications_unread ON player_notifications(player_id, is_read) WHERE is_dismissed = FALSE;
CREATE INDEX IF NOT EXISTS idx_player_notifications_type ON player_notifications(type);
CREATE INDEX IF NOT EXISTS idx_player_notifications_created ON player_notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_player_notifications_expires ON player_notifications(expires_at) WHERE is_dismissed = FALSE;

-- ============================================================================
-- Helper Functions
-- ============================================================================

-- Function to check if players are friends
CREATE OR REPLACE FUNCTION are_friends(player1_id UUID, player2_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM player_friends
        WHERE (player_id = player1_id AND friend_id = player2_id)
           OR (player_id = player2_id AND friend_id = player1_id)
    );
END;
$$ LANGUAGE plpgsql;

-- Function to check if player is blocked
CREATE OR REPLACE FUNCTION is_blocked(blocker_id UUID, blocked_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM player_blocks
        WHERE blocker_id = $1 AND blocked_id = $2
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get unread notification count
CREATE OR REPLACE FUNCTION get_unread_notification_count(p_player_id UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)::INTEGER
        FROM player_notifications
        WHERE player_id = p_player_id
          AND is_read = FALSE
          AND is_dismissed = FALSE
          AND expires_at > NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get unread mail count
CREATE OR REPLACE FUNCTION get_unread_mail_count(p_player_id UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)::INTEGER
        FROM player_mail
        WHERE receiver_id = p_player_id
          AND is_read = FALSE
          AND is_deleted = FALSE
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Cleanup Triggers
-- ============================================================================

-- Auto-cleanup expired notifications
CREATE OR REPLACE FUNCTION cleanup_expired_notifications()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM player_notifications
    WHERE expires_at < NOW() AND is_dismissed = TRUE;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for notification cleanup (runs daily)
DROP TRIGGER IF EXISTS trigger_cleanup_notifications ON player_notifications;
CREATE TRIGGER trigger_cleanup_notifications
    AFTER INSERT ON player_notifications
    EXECUTE FUNCTION cleanup_expired_notifications();

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE player_friends IS 'Bidirectional friend relationships between players';
COMMENT ON TABLE friend_requests IS 'Pending friend requests with acceptance status';
COMMENT ON TABLE player_blocks IS 'Unidirectional block list - blocker cannot be contacted by blocked';
COMMENT ON TABLE player_mail IS 'Persistent mail system with optional credit/item attachments';
COMMENT ON TABLE player_notifications IS 'Real-time notification system for game events';

COMMENT ON COLUMN players.bio IS 'Player profile biography (max 200 characters enforced in app)';
COMMENT ON COLUMN players.profile_privacy IS 'Profile visibility: public, friends, private';
COMMENT ON COLUMN player_mail.attached_credits IS 'Credits sent as mail attachment';
COMMENT ON COLUMN player_mail.attached_items IS 'JSON array of item IDs attached to mail';
COMMENT ON COLUMN player_notifications.action_data IS 'JSON data for notification actions';

-- ============================================================================
-- Grants (if using role-based access)
-- ============================================================================

-- GRANT SELECT, INSERT, UPDATE, DELETE ON player_friends TO terminal_velocity;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON friend_requests TO terminal_velocity;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON player_blocks TO terminal_velocity;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON player_mail TO terminal_velocity;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON player_notifications TO terminal_velocity;
