// File: internal/database/mail_repository.go
// Project: Terminal Velocity
// Description: Database repository for player mail
// Version: 1.0.0
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

// MailRepository handles mail data access
type MailRepository struct {
	db *DB
}

// NewMailRepository creates a new mail repository
func NewMailRepository(db *DB) *MailRepository {
	return &MailRepository{db: db}
}

// CreateMail inserts a new mail message
func (r *MailRepository) CreateMail(ctx context.Context, mail *models.Mail) error {
	query := `
		INSERT INTO player_mail (id, from_player, to_player, subject, body, sent_at, read, deleted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		mail.ID,
		mail.From,
		mail.To,
		mail.Subject,
		mail.Body,
		mail.SentAt,
		mail.Read,
		mail.DeletedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create mail: %w", err)
	}

	return nil
}

// GetMail retrieves a specific mail message
func (r *MailRepository) GetMail(ctx context.Context, mailID uuid.UUID) (*models.Mail, error) {
	query := `
		SELECT id, from_player, to_player, subject, body, sent_at, read, read_at, deleted_by
		FROM player_mail
		WHERE id = $1
	`

	var mail models.Mail
	var readAt sql.NullTime
	var deletedBy []uuid.UUID

	err := r.db.QueryRowContext(ctx, query, mailID).Scan(
		&mail.ID,
		&mail.From,
		&mail.To,
		&mail.Subject,
		&mail.Body,
		&mail.SentAt,
		&mail.Read,
		&readAt,
		&deletedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrMailNotFound
		}
		return nil, fmt.Errorf("failed to query mail: %w", err)
	}

	if readAt.Valid {
		mail.ReadAt = &readAt.Time
	}
	mail.DeletedBy = deletedBy

	return &mail, nil
}

// GetInbox retrieves inbox messages for a player
func (r *MailRepository) GetInbox(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*models.Mail, error) {
	query := `
		SELECT id, from_player, to_player, subject, body, sent_at, read, read_at, deleted_by
		FROM player_mail
		WHERE to_player = $1 AND NOT ($1 = ANY(deleted_by))
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
		var deletedBy []uuid.UUID

		err := rows.Scan(
			&mail.ID,
			&mail.From,
			&mail.To,
			&mail.Subject,
			&mail.Body,
			&mail.SentAt,
			&mail.Read,
			&readAt,
			&deletedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mail: %w", err)
		}

		if readAt.Valid {
			mail.ReadAt = &readAt.Time
		}
		mail.DeletedBy = deletedBy

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
		SELECT id, from_player, to_player, subject, body, sent_at, read, read_at, deleted_by
		FROM player_mail
		WHERE from_player = $1 AND NOT ($1 = ANY(deleted_by))
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
		var deletedBy []uuid.UUID

		err := rows.Scan(
			&mail.ID,
			&mail.From,
			&mail.To,
			&mail.Subject,
			&mail.Body,
			&mail.SentAt,
			&mail.Read,
			&readAt,
			&deletedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan mail: %w", err)
		}

		if readAt.Valid {
			mail.ReadAt = &readAt.Time
		}
		mail.DeletedBy = deletedBy

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
		SET read = true, read_at = $1
		WHERE id = $2 AND read = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), mailID)
	if err != nil {
		return fmt.Errorf("failed to mark mail as read: %w", err)
	}

	return nil
}

// MarkAsDeleted adds a player to the deleted list
func (r *MailRepository) MarkAsDeleted(ctx context.Context, mailID, playerID uuid.UUID) error {
	query := `
		UPDATE player_mail
		SET deleted_by = array_append(deleted_by, $1)
		WHERE id = $2 AND NOT ($1 = ANY(deleted_by))
	`

	_, err := r.db.ExecContext(ctx, query, playerID, mailID)
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
		WHERE to_player = $1 AND read = false AND NOT ($1 = ANY(deleted_by))
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
			COUNT(*) FILTER (WHERE to_player = $1 AND NOT ($1 = ANY(deleted_by))) as received,
			COUNT(*) FILTER (WHERE from_player = $1 AND NOT ($1 = ANY(deleted_by))) as sent,
			COUNT(*) FILTER (WHERE to_player = $1 AND read = false AND NOT ($1 = ANY(deleted_by))) as unread
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

// HardDeleteMail permanently deletes mail deleted by both parties
func (r *MailRepository) HardDeleteMail(ctx context.Context) (int, error) {
	query := `
		DELETE FROM player_mail
		WHERE array_length(deleted_by, 1) = 2
		OR (from_player = ANY(deleted_by) AND to_player = ANY(deleted_by))
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
