// File: internal/mail/manager.go
// Project: Terminal Velocity
// Description: Mail system manager
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package mail

import (
	"context"
	"fmt"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Mail")

// Manager handles mail-related operations
type Manager struct {
	mu         sync.RWMutex
	socialRepo *database.SocialRepository

	// Callbacks for notifications
	onNewMail func(receiverID uuid.UUID, mail *models.Mail)
}

// NewManager creates a new mail manager
func NewManager(socialRepo *database.SocialRepository) *Manager {
	return &Manager{
		socialRepo: socialRepo,
	}
}

// SetNewMailCallback sets the callback for new mail notifications
func (m *Manager) SetNewMailCallback(callback func(receiverID uuid.UUID, mail *models.Mail)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onNewMail = callback
}

// ============================================================================
// Mail Operations
// ============================================================================

// SendMail sends mail to another player
func (m *Manager) SendMail(ctx context.Context, senderID *uuid.UUID, senderName, receiverUsername string, subject, body string, attachedCredits int64, attachedItems []uuid.UUID, getPlayerByUsername func(string) (*models.Player, error), checkBlocked func(uuid.UUID, uuid.UUID) (bool, error)) error {
	// Validate subject and body
	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if len(subject) > 100 {
		return fmt.Errorf("subject too long (max 100 characters)")
	}
	if body == "" {
		return fmt.Errorf("body cannot be empty")
	}
	if len(body) > 10000 {
		return fmt.Errorf("body too long (max 10000 characters)")
	}

	// Look up receiver
	receiver, err := getPlayerByUsername(receiverUsername)
	if err != nil {
		return fmt.Errorf("player not found: %s", receiverUsername)
	}

	// Check if sender is blocked by receiver (if sender is a player)
	if senderID != nil {
		isBlocked, err := checkBlocked(receiver.ID, *senderID)
		if err != nil {
			return fmt.Errorf("failed to check block status: %w", err)
		}
		if isBlocked {
			// Don't reveal that sender is blocked
			return fmt.Errorf("failed to send mail")
		}
	}

	// Validate credits
	if attachedCredits < 0 {
		return fmt.Errorf("cannot attach negative credits")
	}

	// Validate attached items
	if attachedItems == nil {
		attachedItems = []uuid.UUID{}
	}
	if len(attachedItems) > 10 {
		return fmt.Errorf("cannot attach more than 10 items")
	}

	// Create mail
	mail := &models.Mail{
		SenderID:        senderID,
		SenderName:      senderName,
		ReceiverID:      receiver.ID,
		Subject:         subject,
		Body:            body,
		AttachedCredits: attachedCredits,
		AttachedItems:   attachedItems,
	}

	// Send mail (deducts credits if attached)
	err = m.socialRepo.SendMail(ctx, mail)
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	// Trigger notification callback
	m.mu.RLock()
	if m.onNewMail != nil {
		m.onNewMail(receiver.ID, mail)
	}
	m.mu.RUnlock()

	log.Info("Mail sent: from=%s, to=%s, subject=%s, credits=%d, items=%d",
		senderName, receiverUsername, subject, attachedCredits, len(attachedItems))
	return nil
}

// SendSystemMail sends mail from the system (no sender)
func (m *Manager) SendSystemMail(ctx context.Context, receiverID uuid.UUID, subject, body string) error {
	mail := &models.Mail{
		SenderID:        nil,
		SenderName:      "System",
		ReceiverID:      receiverID,
		Subject:         subject,
		Body:            body,
		AttachedCredits: 0,
		AttachedItems:   []uuid.UUID{},
	}

	err := m.socialRepo.SendMail(ctx, mail)
	if err != nil {
		return fmt.Errorf("failed to send system mail: %w", err)
	}

	// Trigger notification
	m.mu.RLock()
	if m.onNewMail != nil {
		m.onNewMail(receiverID, mail)
	}
	m.mu.RUnlock()

	log.Info("System mail sent: to=%s, subject=%s", receiverID, subject)
	return nil
}

// GetMail gets a specific mail by ID
func (m *Manager) GetMail(ctx context.Context, mailID, playerID uuid.UUID) (*models.Mail, error) {
	mail, err := m.socialRepo.GetMail(ctx, mailID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mail: %w", err)
	}

	// Verify player owns this mail
	if mail.ReceiverID != playerID {
		return nil, fmt.Errorf("not authorized to view this mail")
	}

	return mail, nil
}

// GetInbox gets player's inbox (all received mail)
func (m *Manager) GetInbox(ctx context.Context, playerID uuid.UUID, limit int) ([]models.Mail, error) {
	if limit <= 0 || limit > 100 {
		limit = 100 // Default to 100, max 100
	}

	mails, err := m.socialRepo.GetPlayerMail(ctx, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	return mails, nil
}

// GetUnreadMail gets only unread mail
func (m *Manager) GetUnreadMail(ctx context.Context, playerID uuid.UUID) ([]models.Mail, error) {
	allMail, err := m.GetInbox(ctx, playerID, 100)
	if err != nil {
		return nil, err
	}

	var unreadMail []models.Mail
	for _, mail := range allMail {
		if !mail.IsRead {
			unreadMail = append(unreadMail, mail)
		}
	}

	return unreadMail, nil
}

// MarkAsRead marks mail as read
func (m *Manager) MarkAsRead(ctx context.Context, mailID, playerID uuid.UUID) error {
	// Verify ownership
	mail, err := m.GetMail(ctx, mailID, playerID)
	if err != nil {
		return err
	}

	if mail.IsRead {
		return nil // Already read
	}

	err = m.socialRepo.MarkMailAsRead(ctx, mailID)
	if err != nil {
		return fmt.Errorf("failed to mark mail as read: %w", err)
	}

	log.Debug("Mail marked as read: mail=%s, player=%s", mailID, playerID)
	return nil
}

// DeleteMail soft deletes mail
func (m *Manager) DeleteMail(ctx context.Context, mailID, playerID uuid.UUID) error {
	// Verify ownership
	_, err := m.GetMail(ctx, mailID, playerID)
	if err != nil {
		return err
	}

	err = m.socialRepo.DeleteMail(ctx, mailID)
	if err != nil {
		return fmt.Errorf("failed to delete mail: %w", err)
	}

	log.Info("Mail deleted: mail=%s, player=%s", mailID, playerID)
	return nil
}

// ClaimAttachments claims credits and items from mail
func (m *Manager) ClaimAttachments(ctx context.Context, mailID, playerID uuid.UUID) (int64, error) {
	// Get mail to check attachments
	mail, err := m.GetMail(ctx, mailID, playerID)
	if err != nil {
		return 0, err
	}

	if !mail.HasAttachments() {
		return 0, fmt.Errorf("no attachments to claim")
	}

	credits := mail.AttachedCredits

	// Claim attachments (adds credits to player, clears attachments from mail)
	err = m.socialRepo.ClaimMailAttachments(ctx, mailID, playerID)
	if err != nil {
		return 0, fmt.Errorf("failed to claim attachments: %w", err)
	}

	log.Info("Mail attachments claimed: mail=%s, player=%s, credits=%d",
		mailID, playerID, credits)
	return credits, nil
}

// GetUnreadCount gets count of unread mail
func (m *Manager) GetUnreadCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	count, err := m.socialRepo.GetUnreadMailCount(ctx, playerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// MarkAllAsRead marks all mail as read
func (m *Manager) MarkAllAsRead(ctx context.Context, playerID uuid.UUID) error {
	mails, err := m.GetUnreadMail(ctx, playerID)
	if err != nil {
		return err
	}

	for _, mail := range mails {
		err := m.socialRepo.MarkMailAsRead(ctx, mail.ID)
		if err != nil {
			log.Warn("Failed to mark mail as read: mail=%s, error=%v", mail.ID, err)
		}
	}

	log.Info("All mail marked as read: player=%s, count=%d", playerID, len(mails))
	return nil
}

// DeleteAllRead deletes all read mail
func (m *Manager) DeleteAllRead(ctx context.Context, playerID uuid.UUID) error {
	allMail, err := m.GetInbox(ctx, playerID, 100)
	if err != nil {
		return err
	}

	deleteCount := 0
	for _, mail := range allMail {
		if mail.IsRead && !mail.HasAttachments() {
			err := m.socialRepo.DeleteMail(ctx, mail.ID)
			if err != nil {
				log.Warn("Failed to delete mail: mail=%s, error=%v", mail.ID, err)
			} else {
				deleteCount++
			}
		}
	}

	log.Info("Read mail deleted: player=%s, count=%d", playerID, deleteCount)
	return nil
}

// ============================================================================
// Templates
// ============================================================================

// SendWelcomeMail sends welcome mail to new players
func (m *Manager) SendWelcomeMail(ctx context.Context, playerID uuid.UUID) error {
	subject := "Welcome to Terminal Velocity!"
	body := `Welcome, Commander!

You've successfully registered and are now part of the Terminal Velocity universe.

Here are some tips to get started:
- Visit the Trading screen to buy and sell commodities
- Upgrade your ship at the Shipyard
- Complete missions for credits and reputation
- Join a faction to participate in territorial control
- Check the Help screen for detailed guides

The galaxy awaits. Good luck, Commander!

- Terminal Velocity System`

	return m.SendSystemMail(ctx, playerID, subject, body)
}

// SendFriendAcceptedMail sends notification that friend request was accepted
func (m *Manager) SendFriendAcceptedMail(ctx context.Context, playerID uuid.UUID, friendName string) error {
	subject := fmt.Sprintf("%s accepted your friend request", friendName)
	body := fmt.Sprintf(`Good news, Commander!

%s has accepted your friend request. You are now friends and can:
- See each other's online status
- Send direct messages
- View each other's profiles
- Trade with each other

Stay connected with your friends across the galaxy!

- Terminal Velocity System`, friendName)

	return m.SendSystemMail(ctx, playerID, subject, body)
}

// ============================================================================
// Statistics
// ============================================================================

// GetMailStats gets mail statistics for a player
func (m *Manager) GetMailStats(ctx context.Context, playerID uuid.UUID) (map[string]int, error) {
	allMail, err := m.GetInbox(ctx, playerID, 100)
	if err != nil {
		return nil, err
	}

	unreadCount := 0
	withAttachments := 0
	for _, mail := range allMail {
		if !mail.IsRead {
			unreadCount++
		}
		if mail.HasAttachments() {
			withAttachments++
		}
	}

	return map[string]int{
		"total":             len(allMail),
		"unread":            unreadCount,
		"with_attachments": withAttachments,
	}, nil
}
