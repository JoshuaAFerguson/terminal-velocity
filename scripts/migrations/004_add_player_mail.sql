-- Migration: 004_add_player_mail.sql
-- Description: Add player-to-player mail system
-- Date: 2025-01-14

-- Player mail system
CREATE TABLE IF NOT EXISTS player_mail (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_player UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    to_player UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    deleted_by UUID[] DEFAULT '{}',  -- Array of player IDs who deleted this mail
    CONSTRAINT subject_not_empty CHECK (char_length(subject) > 0),
    CONSTRAINT body_not_empty CHECK (char_length(body) > 0)
);

-- Indexes for mail
CREATE INDEX IF NOT EXISTS idx_mail_recipient ON player_mail(to_player, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_mail_sender ON player_mail(from_player, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_mail_unread ON player_mail(to_player, read, sent_at DESC);

-- Comment
COMMENT ON TABLE player_mail IS 'Player-to-player mail messages with soft delete';
