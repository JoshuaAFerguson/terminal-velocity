// File: internal/database/mail_repository.go
// Project: Terminal Velocity
// Description: Repository for player-to-player mail messaging system
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// MailRepository handles all database operations for player mail.
//
// Manages the in-game mail system:
//   - Creating and sending mail messages
//   - Inbox and sent message retrieval
//   - Read/unread status tracking
//   - Message deletion (soft delete)
//   - Unread count queries
//   - Mail statistics
//
// Thread-safety:
//   - All methods are thread-safe
type MailRepository struct {
	db *DB // Database connection pool
}

// NewMailRepository creates a new mail repository
func NewMailRepository(db *DB) *MailRepository {
	return &MailRepository{db: db}
}

// CreateMail inserts a new mail message
func (r *MailRepository) CreateMail(ctx context.Context, mail *models.Mail) error {
	query := `
		INSERT INTO player_mail (id, sender_id, sender_name, receiver_id, subject, body, sent_at, is_read, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		mail.ID,
		mail.SenderID,
		mail.SenderName,
		mail.ReceiverID,
		mail.Subject,
		mail.Body,
		mail.SentAt,
		mail.IsRead,
		mail.IsDeleted,
	)

	if err != nil {
		return fmt.Errorf("failed to create mail: %w", err)
	}

	return nil
}

// GetMail retrieves a specific mail message
func (r *MailRepository) GetMail(ctx context.Context, mailID uuid.UUID) (*models.Mail, error) {
	query := `
		SELECT id, sender_id, sender_name, receiver_id, subject, body, sent_at, is_read, read_at, is_deleted
		FROM player_mail
		WHERE id = $1
	`

	var mail models.Mail
	var readAt sql.NullTime
	var senderID sql.NullString

	err := r.db.QueryRowContext(ctx, query, mailID).Scan(
		&mail.ID,
		&senderID,
		&mail.SenderName,
		&mail.ReceiverID,
		&mail.Subject,
		&mail.Body,
		&mail.SentAt,
		&mail.IsRead,
		&readAt,
		&mail.IsDeleted,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrMailNotFound
		}
		return nil, fmt.Errorf("failed to query mail: %w", err)
	}

	if senderID.Valid {
		id, err := uuid.Parse(senderID.String)
		if err == nil {
			mail.SenderID = &id
		}
	}

	if readAt.Valid {
		mail.ReadAt = &readAt.Time
	}

	return &mail, nil
}

// GetInbox retrieves inbox messages for a player
func (r *MailRepository) GetInbox(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*models.Mail, error) {
	query := `
		SELECT id, sender_id, sender_name, receiver_id, subject, body, sent_at, is_read, read_at, is_deleted
		FROM player_mail
		WHERE receiver_id = $1 AND is_deleted = false
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbox: %w", err)
	}
	defer rows.Close()

	var messages []*models.Mail
	for rows.Next() {
		var mail models.Mail
		var readAt sql.NullTime
		var senderID sql.NullString

		err := rows.Scan(
			&mail.ID,
			&senderID,
			&mail.SenderName,
			&mail.ReceiverID,
			&mail.Subject,
			&mail.Body,
			&mail.SentAt,
			&mail.IsRead,
			&readAt,
			&mail.IsDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mail: %w", err)
		}

		if senderID.Valid {
			id, err := uuid.Parse(senderID.String)
			if err == nil {
				mail.SenderID = &id
			}
		}

		if readAt.Valid {
			mail.ReadAt = &readAt.Time
		}

		messages = append(messages, &mail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mail: %w", err)
	}

	return messages, nil
}

// GetSent retrieves sent messages for a player
func (r *MailRepository) GetSent(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*models.Mail, error) {
	query := `
		SELECT id, sender_id, sender_name, receiver_id, subject, body, sent_at, is_read, read_at, is_deleted
		FROM player_mail
		WHERE sender_id = $1 AND is_deleted = false
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, playerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query sent messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Mail
	for rows.Next() {
		var mail models.Mail
		var readAt sql.NullTime
		var senderID sql.NullString

		err := rows.Scan(
			&mail.ID,
			&senderID,
			&mail.SenderName,
			&mail.ReceiverID,
			&mail.Subject,
			&mail.Body,
			&mail.SentAt,
			&mail.IsRead,
			&readAt,
			&mail.IsDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mail: %w", err)
		}

		if senderID.Valid {
			id, err := uuid.Parse(senderID.String)
			if err == nil {
				mail.SenderID = &id
			}
		}

		if readAt.Valid {
			mail.ReadAt = &readAt.Time
		}

		messages = append(messages, &mail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mail: %w", err)
	}

	return messages, nil
}

// MarkAsRead marks a mail message as read
func (r *MailRepository) MarkAsRead(ctx context.Context, mailID uuid.UUID) error {
	query := `
		UPDATE player_mail
		SET is_read = true, read_at = $1
		WHERE id = $2 AND is_read = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), mailID)
	if err != nil {
		return fmt.Errorf("failed to mark mail as read: %w", err)
	}

	return nil
}

// MarkAsDeleted marks a mail as deleted
func (r *MailRepository) MarkAsDeleted(ctx context.Context, mailID, playerID uuid.UUID) error {
	query := `
		UPDATE player_mail
		SET is_deleted = true
		WHERE id = $1 AND (sender_id = $2 OR receiver_id = $2)
	`

	_, err := r.db.ExecContext(ctx, query, mailID, playerID)
	if err != nil {
		return fmt.Errorf("failed to mark mail as deleted: %w", err)
	}

	return nil
}

// GetUnreadCount returns the number of unread messages for a player
func (r *MailRepository) GetUnreadCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM player_mail
		WHERE receiver_id = $1 AND is_read = false AND is_deleted = false
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread mail: %w", err)
	}

	return count, nil
}

// GetMailStats returns mail statistics for a player
func (r *MailRepository) GetMailStats(ctx context.Context, playerID uuid.UUID) (*models.MailStats, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE receiver_id = $1 AND is_deleted = false) as received,
			COUNT(*) FILTER (WHERE sender_id = $1 AND is_deleted = false) as sent,
			COUNT(*) FILTER (WHERE receiver_id = $1 AND is_read = false AND is_deleted = false) as unread
		FROM player_mail
	`

	var stats models.MailStats
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(
		&stats.TotalReceived,
		&stats.TotalSent,
		&stats.Unread,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get mail stats: %w", err)
	}

	return &stats, nil
}

// DeleteOldMail permanently deletes mail older than the specified time
func (r *MailRepository) DeleteOldMail(ctx context.Context, before time.Time) (int, error) {
	query := `
		DELETE FROM player_mail
		WHERE sent_at < $1
	`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old mail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// HardDeleteMail permanently deletes mail marked as deleted
func (r *MailRepository) HardDeleteMail(ctx context.Context) (int, error) {
	query := `
		DELETE FROM player_mail
		WHERE is_deleted = true
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to hard delete mail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
