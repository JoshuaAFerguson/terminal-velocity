// File: internal/database/social_repository.go
// Project: Terminal Velocity
// Description: Repository for social features including friends, blocks, mail,
//              notifications, and player profiles
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// SocialRepository handles all database operations for social features.
//
// Manages multiplayer social interactions:
//   - Friend system (requests, acceptance, removal)
//   - Block/ignore system
//   - Mail with attachments (credits and items)
//   - Notifications system
//   - Player profiles and playtime
//
// Features:
//   - Bidirectional friendships (both directions stored)
//   - Friend requests with pending/accepted/declined states
//   - Block prevents mail and interactions
//   - Mail supports credit/item attachments with escrow
//   - Notifications with expiration and action data
//
// Thread-safety:
//   - All methods are thread-safe
//   - Uses transactions for multi-step operations
type SocialRepository struct {
	db *DB // Database connection pool
}

// NewSocialRepository creates a new social repository
func NewSocialRepository(db *DB) *SocialRepository {
	return &SocialRepository{db: db}
}

// ============================================================================
// Friends System
// ============================================================================

// AddFriend adds a friendship (both directions)
func (r *SocialRepository) AddFriend(ctx context.Context, playerID, friendID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Add friendship in both directions
		_, err := tx.ExecContext(ctx, `
			INSERT INTO player_friends (player_id, friend_id)
			VALUES ($1, $2), ($2, $1)
			ON CONFLICT (player_id, friend_id) DO NOTHING
		`, playerID, friendID)
		return err
	})
}

// RemoveFriend removes a friendship (both directions)
func (r *SocialRepository) RemoveFriend(ctx context.Context, playerID, friendID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM player_friends
		WHERE (player_id = $1 AND friend_id = $2)
		   OR (player_id = $2 AND friend_id = $1)
	`, playerID, friendID)
	return err
}

// GetFriends gets all friends for a player
func (r *SocialRepository) GetFriends(ctx context.Context, playerID uuid.UUID) ([]models.Friend, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.id, f.player_id, f.friend_id, f.created_at,
		       p.username, p.current_ship
		FROM player_friends f
		JOIN players p ON p.id = f.friend_id
		WHERE f.player_id = $1
		ORDER BY p.username ASC
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []models.Friend
	for rows.Next() {
		var f models.Friend
		var shipID sql.NullString
		err := rows.Scan(&f.ID, &f.PlayerID, &f.FriendID, &f.CreatedAt,
			&f.FriendName, &shipID)
		if err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}

	return friends, rows.Err()
}

// AreFriends checks if two players are friends
func (r *SocialRepository) AreFriends(ctx context.Context, player1ID, player2ID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT are_friends($1, $2)
	`, player1ID, player2ID).Scan(&exists)
	return exists, err
}

// CreateFriendRequest creates a new friend request
func (r *SocialRepository) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*models.FriendRequest, error) {
	var req models.FriendRequest
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO friend_requests (sender_id, receiver_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, sender_id, receiver_id, status, created_at, updated_at
	`, senderID, receiverID, models.FriendRequestPending).Scan(
		&req.ID, &req.SenderID, &req.ReceiverID, &req.Status,
		&req.CreatedAt, &req.UpdatedAt,
	)
	return &req, err
}

// GetFriendRequest gets a friend request by ID
func (r *SocialRepository) GetFriendRequest(ctx context.Context, requestID uuid.UUID) (*models.FriendRequest, error) {
	var req models.FriendRequest
	err := r.db.QueryRowContext(ctx, `
		SELECT fr.id, fr.sender_id, fr.receiver_id, fr.status, fr.created_at, fr.updated_at,
		       sender.username, receiver.username
		FROM friend_requests fr
		JOIN players sender ON sender.id = fr.sender_id
		JOIN players receiver ON receiver.id = fr.receiver_id
		WHERE fr.id = $1
	`, requestID).Scan(
		&req.ID, &req.SenderID, &req.ReceiverID, &req.Status,
		&req.CreatedAt, &req.UpdatedAt, &req.SenderName, &req.ReceiverName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &req, err
}

// GetPendingFriendRequests gets all pending friend requests for a player
func (r *SocialRepository) GetPendingFriendRequests(ctx context.Context, playerID uuid.UUID) ([]models.FriendRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT fr.id, fr.sender_id, fr.receiver_id, fr.status, fr.created_at, fr.updated_at,
		       sender.username, receiver.username
		FROM friend_requests fr
		JOIN players sender ON sender.id = fr.sender_id
		JOIN players receiver ON receiver.id = fr.receiver_id
		WHERE fr.receiver_id = $1 AND fr.status = $2
		ORDER BY fr.created_at DESC
	`, playerID, models.FriendRequestPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.FriendRequest
	for rows.Next() {
		var req models.FriendRequest
		err := rows.Scan(&req.ID, &req.SenderID, &req.ReceiverID, &req.Status,
			&req.CreatedAt, &req.UpdatedAt, &req.SenderName, &req.ReceiverName)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, rows.Err()
}

// AcceptFriendRequest accepts a friend request and creates friendship
func (r *SocialRepository) AcceptFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Get request details
		var senderID, receiverID uuid.UUID
		err := tx.QueryRowContext(ctx, `
			SELECT sender_id, receiver_id FROM friend_requests
			WHERE id = $1 AND status = $2
		`, requestID, models.FriendRequestPending).Scan(&senderID, &receiverID)
		if err != nil {
			return err
		}

		// Update request status
		_, err = tx.ExecContext(ctx, `
			UPDATE friend_requests
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`, models.FriendRequestAccepted, requestID)
		if err != nil {
			return err
		}

		// Create friendship (both directions)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO player_friends (player_id, friend_id)
			VALUES ($1, $2), ($2, $1)
			ON CONFLICT (player_id, friend_id) DO NOTHING
		`, senderID, receiverID)
		return err
	})
}

// DeclineFriendRequest declines a friend request
func (r *SocialRepository) DeclineFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE friend_requests
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status = $3
	`, models.FriendRequestDeclined, requestID, models.FriendRequestPending)
	return err
}

// ============================================================================
// Block/Ignore System
// ============================================================================

// BlockPlayer adds a block
func (r *SocialRepository) BlockPlayer(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Add block
		_, err := tx.ExecContext(ctx, `
			INSERT INTO player_blocks (blocker_id, blocked_id, reason)
			VALUES ($1, $2, $3)
			ON CONFLICT (blocker_id, blocked_id) DO NOTHING
		`, blockerID, blockedID, reason)
		if err != nil {
			return err
		}

		// Remove friendship if exists
		_, err = tx.ExecContext(ctx, `
			DELETE FROM player_friends
			WHERE (player_id = $1 AND friend_id = $2)
			   OR (player_id = $2 AND friend_id = $1)
		`, blockerID, blockedID)
		return err
	})
}

// UnblockPlayer removes a block
func (r *SocialRepository) UnblockPlayer(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM player_blocks
		WHERE blocker_id = $1 AND blocked_id = $2
	`, blockerID, blockedID)
	return err
}

// GetBlockedPlayers gets all blocked players for a blocker
func (r *SocialRepository) GetBlockedPlayers(ctx context.Context, blockerID uuid.UUID) ([]models.Block, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.blocker_id, b.blocked_id, b.reason, b.created_at,
		       p.username
		FROM player_blocks b
		JOIN players p ON p.id = b.blocked_id
		WHERE b.blocker_id = $1
		ORDER BY b.created_at DESC
	`, blockerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []models.Block
	for rows.Next() {
		var b models.Block
		err := rows.Scan(&b.ID, &b.BlockerID, &b.BlockedID, &b.Reason, &b.CreatedAt, &b.BlockedName)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}

	return blocks, rows.Err()
}

// IsBlocked checks if player is blocked
func (r *SocialRepository) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT is_blocked($1, $2)
	`, blockerID, blockedID).Scan(&exists)
	return exists, err
}

// ============================================================================
// Mail System
// ============================================================================

// SendMail sends mail to a player
func (r *SocialRepository) SendMail(ctx context.Context, mail *models.Mail) error {
	itemsJSON, err := json.Marshal(mail.AttachedItems)
	if err != nil {
		return err
	}

	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Deduct attached credits from sender if any
		if mail.AttachedCredits > 0 && mail.SenderID != nil {
			_, err := tx.ExecContext(ctx, `
				UPDATE players SET credits = credits - $1
				WHERE id = $2 AND credits >= $1
			`, mail.AttachedCredits, *mail.SenderID)
			if err != nil {
				return fmt.Errorf("insufficient credits for attachment: %w", err)
			}
		}

		// Insert mail
		err = tx.QueryRowContext(ctx, `
			INSERT INTO player_mail (sender_id, sender_name, receiver_id, subject, body, attached_credits, attached_items)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, sent_at
		`, mail.SenderID, mail.SenderName, mail.ReceiverID, mail.Subject, mail.Body,
			mail.AttachedCredits, itemsJSON).Scan(&mail.ID, &mail.SentAt)
		return err
	})
}

// GetMail gets mail by ID
func (r *SocialRepository) GetMail(ctx context.Context, mailID uuid.UUID) (*models.Mail, error) {
	var mail models.Mail
	var itemsJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, sender_id, sender_name, receiver_id, subject, body,
		       attached_credits, attached_items, is_read, is_deleted, sent_at, read_at
		FROM player_mail
		WHERE id = $1 AND is_deleted = FALSE
	`, mailID).Scan(
		&mail.ID, &mail.SenderID, &mail.SenderName, &mail.ReceiverID,
		&mail.Subject, &mail.Body, &mail.AttachedCredits, &itemsJSON,
		&mail.IsRead, &mail.IsDeleted, &mail.SentAt, &mail.ReadAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &mail.AttachedItems); err != nil {
		return nil, err
	}

	return &mail, nil
}

// GetPlayerMail gets all mail for a player (inbox)
func (r *SocialRepository) GetPlayerMail(ctx context.Context, playerID uuid.UUID, limit int) ([]models.Mail, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, sender_id, sender_name, receiver_id, subject, body,
		       attached_credits, attached_items, is_read, is_deleted, sent_at, read_at
		FROM player_mail
		WHERE receiver_id = $1 AND is_deleted = FALSE
		ORDER BY sent_at DESC
		LIMIT $2
	`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []models.Mail
	for rows.Next() {
		var mail models.Mail
		var itemsJSON []byte
		err := rows.Scan(
			&mail.ID, &mail.SenderID, &mail.SenderName, &mail.ReceiverID,
			&mail.Subject, &mail.Body, &mail.AttachedCredits, &itemsJSON,
			&mail.IsRead, &mail.IsDeleted, &mail.SentAt, &mail.ReadAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(itemsJSON, &mail.AttachedItems); err != nil {
			return nil, err
		}

		mails = append(mails, mail)
	}

	return mails, rows.Err()
}

// MarkMailAsRead marks mail as read
func (r *SocialRepository) MarkMailAsRead(ctx context.Context, mailID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE player_mail
		SET is_read = TRUE, read_at = NOW()
		WHERE id = $1 AND is_read = FALSE
	`, mailID)
	return err
}

// DeleteMail soft deletes mail
func (r *SocialRepository) DeleteMail(ctx context.Context, mailID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE player_mail
		SET is_deleted = TRUE
		WHERE id = $1
	`, mailID)
	return err
}

// ClaimMailAttachments claims credits and items from mail
func (r *SocialRepository) ClaimMailAttachments(ctx context.Context, mailID, playerID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Get mail details
		var credits int64
		var receiverID uuid.UUID
		err := tx.QueryRowContext(ctx, `
			SELECT receiver_id, attached_credits
			FROM player_mail
			WHERE id = $1 AND is_deleted = FALSE
		`, mailID).Scan(&receiverID, &credits)
		if err != nil {
			return err
		}

		// Verify player owns mail
		if receiverID != playerID {
			return fmt.Errorf("not authorized to claim attachments")
		}

		// Add credits to player
		if credits > 0 {
			_, err = tx.ExecContext(ctx, `
				UPDATE players SET credits = credits + $1 WHERE id = $2
			`, credits, playerID)
			if err != nil {
				return err
			}
		}

		// Clear attachments from mail
		_, err = tx.ExecContext(ctx, `
			UPDATE player_mail
			SET attached_credits = 0, attached_items = '[]'
			WHERE id = $1
		`, mailID)
		return err
	})
}

// GetUnreadMailCount gets count of unread mail
func (r *SocialRepository) GetUnreadMailCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT get_unread_mail_count($1)
	`, playerID).Scan(&count)
	return count, err
}

// ============================================================================
// Notification System
// ============================================================================

// CreateNotification creates a new notification
func (r *SocialRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	actionDataJSON, err := json.Marshal(notification.ActionData)
	if err != nil {
		return err
	}

	err = r.db.QueryRowContext(ctx, `
		INSERT INTO player_notifications (player_id, type, title, message, related_player_id, related_entity_type, related_entity_id, action_data, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`, notification.PlayerID, notification.Type, notification.Title, notification.Message,
		notification.RelatedPlayerID, notification.RelatedEntityType, notification.RelatedEntityID,
		actionDataJSON, notification.ExpiresAt).Scan(&notification.ID, &notification.CreatedAt)
	return err
}

// GetPlayerNotifications gets notifications for a player
func (r *SocialRepository) GetPlayerNotifications(ctx context.Context, playerID uuid.UUID, limit int) ([]models.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, player_id, type, title, message, related_player_id, related_entity_type, related_entity_id,
		       is_read, is_dismissed, created_at, read_at, expires_at, action_data
		FROM player_notifications
		WHERE player_id = $1 AND is_dismissed = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT $2
	`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var actionDataJSON []byte
		err := rows.Scan(
			&n.ID, &n.PlayerID, &n.Type, &n.Title, &n.Message,
			&n.RelatedPlayerID, &n.RelatedEntityType, &n.RelatedEntityID,
			&n.IsRead, &n.IsDismissed, &n.CreatedAt, &n.ReadAt, &n.ExpiresAt,
			&actionDataJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(actionDataJSON, &n.ActionData); err != nil {
			n.ActionData = make(map[string]interface{})
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

// MarkNotificationAsRead marks notification as read
func (r *SocialRepository) MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE player_notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE id = $1 AND is_read = FALSE
	`, notificationID)
	return err
}

// DismissNotification dismisses a notification
func (r *SocialRepository) DismissNotification(ctx context.Context, notificationID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE player_notifications
		SET is_dismissed = TRUE
		WHERE id = $1
	`, notificationID)
	return err
}

// GetUnreadNotificationCount gets count of unread notifications
func (r *SocialRepository) GetUnreadNotificationCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT get_unread_notification_count($1)
	`, playerID).Scan(&count)
	return count, err
}

// ============================================================================
// Player Profile
// ============================================================================

// UpdatePlayerProfile updates player profile fields
func (r *SocialRepository) UpdatePlayerProfile(ctx context.Context, playerID uuid.UUID, bio string, privacy string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE players
		SET bio = $1, profile_privacy = $2
		WHERE id = $3
	`, bio, privacy, playerID)
	return err
}

// UpdatePlaytime adds time to player's total playtime
func (r *SocialRepository) UpdatePlaytime(ctx context.Context, playerID uuid.UUID, seconds int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE players
		SET total_playtime = total_playtime + $1
		WHERE id = $2
	`, seconds, playerID)
	return err
}

// ErrNotFound is returned when a record is not found
var ErrNotFound = fmt.Errorf("record not found")
