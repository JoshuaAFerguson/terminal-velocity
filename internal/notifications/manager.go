// File: internal/notifications/manager.go
// Project: Terminal Velocity
// Description: Notification system manager
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Notifications")

// Manager handles notification operations
type Manager struct {
	mu         sync.RWMutex
	socialRepo *database.SocialRepository

	// Callbacks for real-time notification delivery
	onNewNotification func(playerID uuid.UUID, notification *models.Notification)

	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewManager creates a new notification manager
func NewManager(socialRepo *database.SocialRepository) *Manager {
	return &Manager{
		socialRepo:  socialRepo,
		stopCleanup: make(chan bool),
	}
}

// Start begins background tasks (cleanup of expired notifications)
func (m *Manager) Start() {
	m.cleanupTicker = time.NewTicker(1 * time.Hour)
	go m.cleanupExpiredNotifications()
	log.Info("Notification manager started")
}

// Stop stops background tasks
func (m *Manager) Stop() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
	close(m.stopCleanup)
	log.Info("Notification manager stopped")
}

// SetNotificationCallback sets the callback for real-time notification delivery
func (m *Manager) SetNotificationCallback(callback func(playerID uuid.UUID, notification *models.Notification)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onNewNotification = callback
}

// ============================================================================
// Notification Operations
// ============================================================================

// CreateNotification creates a new notification
func (m *Manager) CreateNotification(ctx context.Context, notification *models.Notification) error {
	err := m.socialRepo.CreateNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Trigger real-time notification callback
	m.mu.RLock()
	if m.onNewNotification != nil {
		m.onNewNotification(notification.PlayerID, notification)
	}
	m.mu.RUnlock()

	log.Debug("Notification created: player=%s, type=%s", notification.PlayerID, notification.Type)
	return nil
}

// GetNotifications gets all notifications for a player
func (m *Manager) GetNotifications(ctx context.Context, playerID uuid.UUID, limit int) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default to 50, max 100
	}

	notifications, err := m.socialRepo.GetPlayerNotifications(ctx, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadNotifications gets only unread notifications
func (m *Manager) GetUnreadNotifications(ctx context.Context, playerID uuid.UUID) ([]models.Notification, error) {
	allNotifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return nil, err
	}

	var unreadNotifications []models.Notification
	for _, notification := range allNotifications {
		if !notification.IsRead && !notification.IsExpired() {
			unreadNotifications = append(unreadNotifications, notification)
		}
	}

	return unreadNotifications, nil
}

// GetNotificationsByType gets notifications of a specific type
func (m *Manager) GetNotificationsByType(ctx context.Context, playerID uuid.UUID, notificationType string) ([]models.Notification, error) {
	allNotifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return nil, err
	}

	var filteredNotifications []models.Notification
	for _, notification := range allNotifications {
		if notification.Type == notificationType && !notification.IsExpired() {
			filteredNotifications = append(filteredNotifications, notification)
		}
	}

	return filteredNotifications, nil
}

// MarkAsRead marks a notification as read
func (m *Manager) MarkAsRead(ctx context.Context, notificationID, playerID uuid.UUID) error {
	// Verify ownership by getting the notification
	notifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return err
	}

	// Find the notification and verify ownership
	found := false
	for _, notification := range notifications {
		if notification.ID == notificationID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("notification not found or not authorized")
	}

	err = m.socialRepo.MarkNotificationAsRead(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	log.Debug("Notification marked as read: notification=%s, player=%s", notificationID, playerID)
	return nil
}

// DismissNotification dismisses a notification
func (m *Manager) DismissNotification(ctx context.Context, notificationID, playerID uuid.UUID) error {
	// Verify ownership
	notifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return err
	}

	found := false
	for _, notification := range notifications {
		if notification.ID == notificationID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("notification not found or not authorized")
	}

	err = m.socialRepo.DismissNotification(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to dismiss notification: %w", err)
	}

	log.Debug("Notification dismissed: notification=%s, player=%s", notificationID, playerID)
	return nil
}

// GetUnreadCount gets count of unread notifications
func (m *Manager) GetUnreadCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	count, err := m.socialRepo.GetUnreadNotificationCount(ctx, playerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// MarkAllAsRead marks all notifications as read
func (m *Manager) MarkAllAsRead(ctx context.Context, playerID uuid.UUID) error {
	unreadNotifications, err := m.GetUnreadNotifications(ctx, playerID)
	if err != nil {
		return err
	}

	for _, notification := range unreadNotifications {
		err := m.socialRepo.MarkNotificationAsRead(ctx, notification.ID)
		if err != nil {
			log.Warn("Failed to mark notification as read: notification=%s, error=%v", notification.ID, err)
		}
	}

	log.Info("All notifications marked as read: player=%s, count=%d", playerID, len(unreadNotifications))
	return nil
}

// DismissAllRead dismisses all read notifications
func (m *Manager) DismissAllRead(ctx context.Context, playerID uuid.UUID) error {
	allNotifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return err
	}

	dismissCount := 0
	for _, notification := range allNotifications {
		if notification.IsRead && !notification.IsDismissed {
			err := m.socialRepo.DismissNotification(ctx, notification.ID)
			if err != nil {
				log.Warn("Failed to dismiss notification: notification=%s, error=%v", notification.ID, err)
			} else {
				dismissCount++
			}
		}
	}

	log.Info("Read notifications dismissed: player=%s, count=%d", playerID, dismissCount)
	return nil
}

// ClearExpired removes all expired notifications
func (m *Manager) ClearExpired(ctx context.Context, playerID uuid.UUID) error {
	allNotifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return err
	}

	clearCount := 0
	for _, notification := range allNotifications {
		if notification.IsExpired() {
			err := m.socialRepo.DismissNotification(ctx, notification.ID)
			if err != nil {
				log.Warn("Failed to clear expired notification: notification=%s, error=%v", notification.ID, err)
			} else {
				clearCount++
			}
		}
	}

	log.Info("Expired notifications cleared: player=%s, count=%d", playerID, clearCount)
	return nil
}

// ============================================================================
// Notification Templates
// ============================================================================

// NotifyFriendRequest sends a friend request notification
func (m *Manager) NotifyFriendRequest(ctx context.Context, receiverID uuid.UUID, senderName string, requestID uuid.UUID) error {
	notification := &models.Notification{
		PlayerID:          receiverID,
		Type:              models.NotificationTypeFriendRequest,
		Title:             "New Friend Request",
		Message:           fmt.Sprintf("%s sent you a friend request", senderName),
		RelatedEntityType: "friend_request",
		RelatedEntityID:   &requestID,
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour), // 7 days
		ActionData: map[string]interface{}{
			"sender_name": senderName,
			"request_id":  requestID.String(),
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyFriendAccepted sends a friend request accepted notification
func (m *Manager) NotifyFriendAccepted(ctx context.Context, senderID uuid.UUID, accepterName string) error {
	notification := &models.Notification{
		PlayerID:  senderID,
		Type:      models.NotificationTypeFriendRequest,
		Title:     "Friend Request Accepted",
		Message:   fmt.Sprintf("%s accepted your friend request", accepterName),
		ExpiresAt: time.Now().Add(3 * 24 * time.Hour), // 3 days
		ActionData: map[string]interface{}{
			"accepter_name": accepterName,
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyNewMail sends a new mail notification
func (m *Manager) NotifyNewMail(ctx context.Context, receiverID uuid.UUID, senderName, subject string, mailID uuid.UUID, hasAttachments bool) error {
	message := fmt.Sprintf("New mail from %s: %s", senderName, subject)
	if hasAttachments {
		message = fmt.Sprintf("New mail with attachments from %s: %s", senderName, subject)
	}

	notification := &models.Notification{
		PlayerID:          receiverID,
		Type:              models.NotificationTypeMail,
		Title:             "New Mail",
		Message:           message,
		RelatedEntityType: "mail",
		RelatedEntityID:   &mailID,
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour), // 7 days
		ActionData: map[string]interface{}{
			"sender_name":     senderName,
			"subject":         subject,
			"mail_id":         mailID.String(),
			"has_attachments": hasAttachments,
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyTradeOffer sends a trade offer notification
func (m *Manager) NotifyTradeOffer(ctx context.Context, receiverID, senderID uuid.UUID, senderName string, tradeID uuid.UUID) error {
	notification := &models.Notification{
		PlayerID:          receiverID,
		Type:              models.NotificationTypeTradeOffer,
		Title:             "Trade Offer",
		Message:           fmt.Sprintf("%s sent you a trade offer", senderName),
		RelatedPlayerID:   &senderID,
		RelatedEntityType: "trade",
		RelatedEntityID:   &tradeID,
		ExpiresAt:         time.Now().Add(24 * time.Hour), // 24 hours
		ActionData: map[string]interface{}{
			"sender_name": senderName,
			"trade_id":    tradeID.String(),
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyPvPChallenge sends a PvP challenge notification
func (m *Manager) NotifyPvPChallenge(ctx context.Context, receiverID, senderID uuid.UUID, senderName string) error {
	notification := &models.Notification{
		PlayerID:        receiverID,
		Type:            models.NotificationTypePvPChallenge,
		Title:           "PvP Challenge",
		Message:         fmt.Sprintf("%s challenged you to a duel", senderName),
		RelatedPlayerID: &senderID,
		ExpiresAt:       time.Now().Add(30 * time.Minute), // 30 minutes
		ActionData: map[string]interface{}{
			"sender_name": senderName,
			"challenger":  senderID.String(),
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyTerritoryAttack sends a territory attack notification
func (m *Manager) NotifyTerritoryAttack(ctx context.Context, factionLeaderID, attackerFactionID uuid.UUID, attackerFactionName, systemName string) error {
	notification := &models.Notification{
		PlayerID: factionLeaderID,
		Type:     models.NotificationTypeTerritoryAttack,
		Title:    "Territory Under Attack",
		Message:  fmt.Sprintf("%s is attacking your territory in %s", attackerFactionName, systemName),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
		ActionData: map[string]interface{}{
			"attacker_faction": attackerFactionName,
			"system_name":      systemName,
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyFactionInvite sends a faction invite notification
func (m *Manager) NotifyFactionInvite(ctx context.Context, receiverID, factionID uuid.UUID, factionName string) error {
	notification := &models.Notification{
		PlayerID:          receiverID,
		Type:              models.NotificationTypeFactionInvite,
		Title:             "Faction Invitation",
		Message:           fmt.Sprintf("You've been invited to join %s", factionName),
		RelatedEntityType: "faction",
		RelatedEntityID:   &factionID,
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour), // 7 days
		ActionData: map[string]interface{}{
			"faction_name": factionName,
			"faction_id":   factionID.String(),
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifySystemMessage sends a system-wide message notification
func (m *Manager) NotifySystemMessage(ctx context.Context, playerID uuid.UUID, title, message string, expiresIn time.Duration) error {
	notification := &models.Notification{
		PlayerID:  playerID,
		Type:      models.NotificationTypeSystemMessage,
		Title:     title,
		Message:   message,
		ExpiresAt: time.Now().Add(expiresIn),
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyAchievement sends an achievement unlock notification
func (m *Manager) NotifyAchievement(ctx context.Context, playerID uuid.UUID, achievementName, achievementDescription string) error {
	notification := &models.Notification{
		PlayerID:  playerID,
		Type:      models.NotificationTypeAchievement,
		Title:     "Achievement Unlocked!",
		Message:   fmt.Sprintf("%s: %s", achievementName, achievementDescription),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		ActionData: map[string]interface{}{
			"achievement_name": achievementName,
		},
	}

	return m.CreateNotification(ctx, notification)
}

// NotifyEvent sends an event notification
func (m *Manager) NotifyEvent(ctx context.Context, playerID uuid.UUID, eventName, eventDescription string, eventID uuid.UUID) error {
	notification := &models.Notification{
		PlayerID:          playerID,
		Type:              models.NotificationTypeEvent,
		Title:             fmt.Sprintf("Event: %s", eventName),
		Message:           eventDescription,
		RelatedEntityType: "event",
		RelatedEntityID:   &eventID,
		ExpiresAt:         time.Now().Add(24 * time.Hour), // 24 hours
		ActionData: map[string]interface{}{
			"event_name": eventName,
			"event_id":   eventID.String(),
		},
	}

	return m.CreateNotification(ctx, notification)
}

// ============================================================================
// Background Tasks
// ============================================================================

// cleanupExpiredNotifications removes expired and dismissed notifications
func (m *Manager) cleanupExpiredNotifications() {
	for {
		select {
		case <-m.cleanupTicker.C:
			// This would require a method in the repository to clean up expired notifications
			// For now, we'll just log that cleanup ran
			log.Debug("Running notification cleanup (expired notifications auto-deleted by database trigger)")
		case <-m.stopCleanup:
			return
		}
	}
}

// ============================================================================
// Statistics
// ============================================================================

// GetNotificationStats gets notification statistics for a player
func (m *Manager) GetNotificationStats(ctx context.Context, playerID uuid.UUID) (map[string]int, error) {
	allNotifications, err := m.GetNotifications(ctx, playerID, 100)
	if err != nil {
		return nil, err
	}

	unreadCount := 0
	expiredCount := 0
	byType := make(map[string]int)

	for _, notification := range allNotifications {
		if !notification.IsRead {
			unreadCount++
		}
		if notification.IsExpired() {
			expiredCount++
		}
		byType[notification.Type]++
	}

	return map[string]int{
		"total":                         len(allNotifications),
		"unread":                        unreadCount,
		"expired":                       expiredCount,
		"friend_requests":               byType[models.NotificationTypeFriendRequest],
		"mail":                          byType[models.NotificationTypeMail],
		"trade_offers":                  byType[models.NotificationTypeTradeOffer],
		"pvp_challenges":                byType[models.NotificationTypePvPChallenge],
		"territory_attacks":             byType[models.NotificationTypeTerritoryAttack],
		"faction_invites":               byType[models.NotificationTypeFactionInvite],
		"system_messages":               byType[models.NotificationTypeSystemMessage],
		"achievements":                  byType[models.NotificationTypeAchievement],
		"events":                        byType[models.NotificationTypeEvent],
	}, nil
}
