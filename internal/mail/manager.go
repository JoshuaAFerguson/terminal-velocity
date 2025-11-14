// File: internal/mail/manager.go
// Project: Terminal Velocity
// Description: Player mail system for asynchronous messaging
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package mail

import (
	"context"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Mail")

// Manager handles the player mail system
type Manager struct {
	mu       sync.RWMutex
	mailRepo *database.MailRepository

	// Cache for quick unread counts
	unreadCounts map[uuid.UUID]int
}

// NewManager creates a new mail manager
func NewManager(mailRepo *database.MailRepository) *Manager {
	return &Manager{
		mailRepo:     mailRepo,
		unreadCounts: make(map[uuid.UUID]int),
	}
}

// Start initializes the mail manager
func (m *Manager) Start() error {
	log.Info("Mail manager started")
	return nil
}

// Stop cleans up the mail manager
func (m *Manager) Stop() {
	log.Info("Mail manager stopped")
}

// SendMail sends a mail message from one player to another
func (m *Manager) SendMail(ctx context.Context, from, to uuid.UUID, subject, body string) (*models.Mail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mail := &models.Mail{
		ID:        uuid.New(),
		From:      from,
		To:        to,
		Subject:   subject,
		Body:      body,
		SentAt:    time.Now(),
		Read:      false,
		DeletedBy: []uuid.UUID{},
	}

	if err := m.mailRepo.CreateMail(ctx, mail); err != nil {
		log.Error("Failed to create mail: %v", err)
		return nil, err
	}

	// Increment unread count
	m.unreadCounts[to]++

	log.Info("Mail sent from %s to %s: %s", from, to, subject)
	return mail, nil
}

// GetInbox retrieves inbox messages for a player
func (m *Manager) GetInbox(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*models.Mail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages, err := m.mailRepo.GetInbox(ctx, playerID, limit, offset)
	if err != nil {
		log.Error("Failed to get inbox for %s: %v", playerID, err)
		return nil, err
	}

	return messages, nil
}

// GetSent retrieves sent messages for a player
func (m *Manager) GetSent(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*models.Mail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages, err := m.mailRepo.GetSent(ctx, playerID, limit, offset)
	if err != nil {
		log.Error("Failed to get sent messages for %s: %v", playerID, err)
		return nil, err
	}

	return messages, nil
}

// GetMail retrieves a specific mail message
func (m *Manager) GetMail(ctx context.Context, mailID, playerID uuid.UUID) (*models.Mail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mail, err := m.mailRepo.GetMail(ctx, mailID)
	if err != nil {
		return nil, err
	}

	// Verify player has access (sender or recipient)
	if mail.From != playerID && mail.To != playerID {
		log.Warn("Player %s attempted to access mail %s without permission", playerID, mailID)
		return nil, models.ErrUnauthorized
	}

	return mail, nil
}

// MarkAsRead marks a mail message as read
func (m *Manager) MarkAsRead(ctx context.Context, mailID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mail, err := m.mailRepo.GetMail(ctx, mailID)
	if err != nil {
		return err
	}

	// Verify player is recipient
	if mail.To != playerID {
		log.Warn("Player %s attempted to mark mail %s as read (not recipient)", playerID, mailID)
		return models.ErrUnauthorized
	}

	// Already read?
	if mail.Read {
		return nil
	}

	if err := m.mailRepo.MarkAsRead(ctx, mailID); err != nil {
		log.Error("Failed to mark mail %s as read: %v", mailID, err)
		return err
	}

	// Decrement unread count
	if m.unreadCounts[playerID] > 0 {
		m.unreadCounts[playerID]--
	}

	log.Debug("Mail %s marked as read by %s", mailID, playerID)
	return nil
}

// DeleteMail marks a mail message as deleted for a player
func (m *Manager) DeleteMail(ctx context.Context, mailID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mail, err := m.mailRepo.GetMail(ctx, mailID)
	if err != nil {
		return err
	}

	// Verify player has access
	if mail.From != playerID && mail.To != playerID {
		log.Warn("Player %s attempted to delete mail %s without permission", playerID, mailID)
		return models.ErrUnauthorized
	}

	// Add player to deleted list
	if err := m.mailRepo.MarkAsDeleted(ctx, mailID, playerID); err != nil {
		log.Error("Failed to mark mail %s as deleted: %v", mailID, err)
		return err
	}

	// Update unread count if recipient is deleting unread mail
	if mail.To == playerID && !mail.Read {
		if m.unreadCounts[playerID] > 0 {
			m.unreadCounts[playerID]--
		}
	}

	log.Debug("Mail %s deleted by %s", mailID, playerID)
	return nil
}

// GetUnreadCount returns the number of unread messages for a player
func (m *Manager) GetUnreadCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	m.mu.RLock()

	// Check cache first
	if count, ok := m.unreadCounts[playerID]; ok {
		m.mu.RUnlock()
		return count, nil
	}
	m.mu.RUnlock()

	// Cache miss - query database
	count, err := m.mailRepo.GetUnreadCount(ctx, playerID)
	if err != nil {
		log.Error("Failed to get unread count for %s: %v", playerID, err)
		return 0, err
	}

	// Update cache
	m.mu.Lock()
	m.unreadCounts[playerID] = count
	m.mu.Unlock()

	return count, nil
}

// RefreshUnreadCount refreshes the unread count cache for a player
func (m *Manager) RefreshUnreadCount(ctx context.Context, playerID uuid.UUID) error {
	count, err := m.mailRepo.GetUnreadCount(ctx, playerID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.unreadCounts[playerID] = count
	m.mu.Unlock()

	return nil
}

// CleanupOldMail deletes mail messages older than the specified duration
func (m *Manager) CleanupOldMail(ctx context.Context, olderThan time.Duration) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	deleted, err := m.mailRepo.DeleteOldMail(ctx, cutoff)
	if err != nil {
		log.Error("Failed to cleanup old mail: %v", err)
		return 0, err
	}

	if deleted > 0 {
		log.Info("Cleaned up %d old mail messages", deleted)

		// Clear cache to force refresh
		m.unreadCounts = make(map[uuid.UUID]int)
	}

	return deleted, nil
}
