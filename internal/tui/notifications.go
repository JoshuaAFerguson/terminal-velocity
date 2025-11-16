// File: internal/tui/notifications.go
// Project: Terminal Velocity
// Description: Notifications system TUI screen for viewing and managing game notifications
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15
//
// This file implements the notifications screen where players can:
//   - View all notifications (system, quest, mission, achievement, social, etc.)
//   - Filter notifications by type or read status
//   - Mark notifications as read
//   - Dismiss individual notifications
//   - Clear all dismissed notifications
//
// The notifications screen supports multiple modes:
//   - All: Shows all notifications
//   - Unread: Shows only unread notifications
//   - By Type: Filters by notification type (quest, mission, achievement, etc.)
//   - View: Shows detailed view of a single notification
//
// Notification Types:
//   - System: Server announcements, maintenance, etc.
//   - Quest: Quest updates and completions
//   - Mission: Mission assignments and completions
//   - Achievement: Achievement unlocks
//   - Event: Event updates and rewards
//   - Social: Friend requests, messages
//   - Faction: Faction invites, promotions
//   - Trade: Trade offers and completions
//   - Combat: PvP challenges, territory attacks
//
// Thread Safety:
//   - Update functions are called sequentially by BubbleTea
//   - Notifications manager is thread-safe for concurrent access

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// ===== Notifications Screen Modes =====
// Constants defining the different modes/views of the notifications screen

const (
	notificationsModeAll    = "all"     // View all notifications (read and unread)
	notificationsModeUnread = "unread"  // View only unread notifications
	notificationsModeByType = "by_type" // Filter notifications by type
	notificationsModeView   = "view"    // View detailed single notification
)

// notificationsState holds the state for the notifications screen.
//
// This struct preserves state when switching between screens, allowing
// players to return to the notifications screen with their previous
// filter/selection intact.
//
// State Management:
//   - mode: Current view mode (all, unread, by_type, view)
//   - selectedIndex: Index of currently selected notification in list
//   - currentNotif: Notification being viewed in detail (view mode only)
//   - loading: Whether async operation is in progress
//   - error: Error message to display (if any)
type notificationsState struct {
	mode            string                  // Current screen mode
	notifications   []models.Notification   // Currently displayed notifications
	selectedIndex   int                     // Index of selected notification
	currentNotif    *models.Notification    // Notification being viewed (view mode)
	loading         bool                    // Loading state
	error           string                  // Error message (if any)
	unreadCount     int                     // Total unread count (for display)
	selectedType    string                  // Selected type filter (by_type mode)
	availableTypes  []string                // Available notification types for filtering
	typeSelectIndex int                     // Index in type selection list
}

func newNotificationsState() notificationsState {
	return notificationsState{
		mode:          notificationsModeAll,
		selectedIndex: 0,
		availableTypes: []string{
			"system",
			"quest",
			"mission",
			"achievement",
			"event",
			"social",
			"faction",
			"trade",
			"combat",
		},
	}
}

func (m *Model) updateNotifications(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle view mode separately
		if m.notifications.mode == notificationsModeView {
			return m.updateNotificationView(msg)
		}

		// Handle type selection mode
		if m.notifications.mode == notificationsModeByType && m.notifications.selectedType == "" {
			return m.updateNotificationTypeSelect(msg)
		}

		// Common navigation
		switch msg.String() {
		case "q", "esc":
			if m.notifications.mode == notificationsModeByType && m.notifications.selectedType != "" {
				// Return to type selection
				m.notifications.selectedType = ""
				m.notifications.selectedIndex = 0
				return m, nil
			}
			m.screen = ScreenGame
			return m, nil

		case "1":
			// Switch to all notifications
			m.notifications.mode = notificationsModeAll
			m.notifications.selectedIndex = 0
			m.notifications.loading = true
			return m, m.loadAllNotifications()

		case "2":
			// Switch to unread only
			m.notifications.mode = notificationsModeUnread
			m.notifications.selectedIndex = 0
			m.notifications.loading = true
			return m, m.loadUnreadNotifications()

		case "3":
			// Switch to filter by type
			m.notifications.mode = notificationsModeByType
			m.notifications.selectedType = ""
			m.notifications.typeSelectIndex = 0
			m.notifications.selectedIndex = 0
			return m, nil

		case "up", "k":
			if m.notifications.selectedIndex > 0 {
				m.notifications.selectedIndex--
			}

		case "down", "j":
			maxIndex := len(m.notifications.notifications) - 1
			if m.notifications.selectedIndex < maxIndex {
				m.notifications.selectedIndex++
			}

		case "r":
			// Refresh current view
			m.notifications.loading = true
			switch m.notifications.mode {
			case notificationsModeAll:
				return m, m.loadAllNotifications()
			case notificationsModeUnread:
				return m, m.loadUnreadNotifications()
			case notificationsModeByType:
				if m.notifications.selectedType != "" {
					return m, m.loadNotificationsByType(m.notifications.selectedType)
				}
			}

		case "enter":
			// View selected notification
			if len(m.notifications.notifications) > 0 {
				m.notifications.currentNotif = &m.notifications.notifications[m.notifications.selectedIndex]
				m.notifications.mode = notificationsModeView
				// Mark as read when viewed
				return m, m.markNotificationAsRead(m.notifications.currentNotif.ID)
			}

		case "d":
			// Dismiss notification
			if len(m.notifications.notifications) > 0 {
				notifID := m.notifications.notifications[m.notifications.selectedIndex].ID
				return m, m.dismissNotification(notifID)
			}

		case "a":
			// Mark all as read
			return m, m.markAllNotificationsAsRead()

		case "c":
			// Clear dismissed
			return m, m.clearDismissedNotifications()
		}

	case notificationsLoadedMsg:
		m.notifications.loading = false
		m.notifications.notifications = msg.notifications
		m.notifications.unreadCount = msg.unreadCount
		m.notifications.error = msg.err
		if m.notifications.selectedIndex >= len(msg.notifications) {
			m.notifications.selectedIndex = 0
		}

	case notificationActionMsg:
		m.notifications.loading = false
		if msg.err == "" {
			// Success - reload current view
			m.notifications.error = ""
			switch m.notifications.mode {
			case notificationsModeAll:
				return m, m.loadAllNotifications()
			case notificationsModeUnread:
				return m, m.loadUnreadNotifications()
			case notificationsModeByType:
				if m.notifications.selectedType != "" {
					return m, m.loadNotificationsByType(m.notifications.selectedType)
				}
			}
		} else {
			m.notifications.error = msg.err
		}
	}

	return m, nil
}

func (m *Model) updateNotificationTypeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.notifications.mode = notificationsModeAll
		m.notifications.loading = true
		return m, m.loadAllNotifications()

	case "up", "k":
		if m.notifications.typeSelectIndex > 0 {
			m.notifications.typeSelectIndex--
		}

	case "down", "j":
		if m.notifications.typeSelectIndex < len(m.notifications.availableTypes)-1 {
			m.notifications.typeSelectIndex++
		}

	case "enter":
		// Select type and load notifications
		m.notifications.selectedType = m.notifications.availableTypes[m.notifications.typeSelectIndex]
		m.notifications.selectedIndex = 0
		m.notifications.loading = true
		return m, m.loadNotificationsByType(m.notifications.selectedType)
	}

	return m, nil
}

func (m *Model) updateNotificationView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Return to previous view
		m.notifications.mode = notificationsModeAll
		m.notifications.currentNotif = nil
		return m, nil

	case "d":
		// Dismiss and return to list
		if m.notifications.currentNotif != nil {
			notifID := m.notifications.currentNotif.ID
			m.notifications.mode = notificationsModeAll
			m.notifications.currentNotif = nil
			return m, m.dismissNotification(notifID)
		}
	}

	return m, nil
}

func (m *Model) viewNotifications() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("═══ NOTIFICATIONS ═══")

	b.WriteString(title + "\n\n")

	if m.notifications.loading {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("Loading...") + "\n\n")
		return b.String()
	}

	if m.notifications.error != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("Error: "+m.notifications.error) + "\n\n")
	}

	// Mode selector
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render("Filters:") + " ")

	if m.notifications.mode == notificationsModeAll {
		b.WriteString(fmt.Sprintf("[1] All (%d)  ", len(m.notifications.notifications)))
	} else {
		b.WriteString("[1] All  ")
	}

	if m.notifications.mode == notificationsModeUnread {
		b.WriteString(fmt.Sprintf("[2] Unread (%d)  ", m.notifications.unreadCount))
	} else {
		b.WriteString(fmt.Sprintf("[2] Unread (%d)  ", m.notifications.unreadCount))
	}

	if m.notifications.mode == notificationsModeByType {
		b.WriteString("[3] By Type  ")
	} else {
		b.WriteString("[3] By Type  ")
	}

	b.WriteString("\n\n")

	// Render based on mode
	switch m.notifications.mode {
	case notificationsModeAll, notificationsModeUnread, notificationsModeByType:
		if m.notifications.mode == notificationsModeByType && m.notifications.selectedType == "" {
			b.WriteString(m.renderNotificationTypeSelect())
		} else {
			b.WriteString(m.renderNotificationsList())
		}
	case notificationsModeView:
		b.WriteString(m.renderNotificationView())
	}

	return b.String()
}

func (m *Model) renderNotificationsList() string {
	var b strings.Builder

	if len(m.notifications.notifications) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No notifications.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-15s %-50s %-20s\n", "Type", "Message", "Date")))
	b.WriteString(strings.Repeat("─", 85) + "\n")

	// Notifications list (show max 15)
	maxDisplay := 15
	if len(m.notifications.notifications) < maxDisplay {
		maxDisplay = len(m.notifications.notifications)
	}

	for i := 0; i < maxDisplay; i++ {
		notif := m.notifications.notifications[i]

		// Color code by type
		style := lipgloss.NewStyle()
		if !notif.IsRead {
			style = style.Foreground(lipgloss.Color("11")).Bold(true) // Yellow for unread
		} else {
			style = style.Foreground(lipgloss.Color("8")) // Gray for read
		}

		// Special colors for important types
		switch notif.Type {
		case models.NotificationTypeFriendRequest:
			if !notif.IsRead {
				style = style.Foreground(lipgloss.Color("14")) // Cyan
			}
		case models.NotificationTypeMail:
			if !notif.IsRead {
				style = style.Foreground(lipgloss.Color("10")) // Green
			}
		case models.NotificationTypePvPChallenge:
			if !notif.IsRead {
				style = style.Foreground(lipgloss.Color("9")) // Red
			}
		case models.NotificationTypeTerritoryAttack:
			if !notif.IsRead {
				style = style.Foreground(lipgloss.Color("9")) // Red
			}
		case models.NotificationTypeAchievement:
			if !notif.IsRead {
				style = style.Foreground(lipgloss.Color("226")) // Gold
			}
		}

		if i == m.notifications.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.notifications.selectedIndex {
			cursor = "→ "
		}

		// Format type (remove "notification_" prefix if present)
		typeStr := notif.Type
		if strings.HasPrefix(typeStr, "notification_") {
			typeStr = strings.TrimPrefix(typeStr, "notification_")
		}

		timeStr := notif.CreatedAt.Format("2006-01-02 15:04")

		line := fmt.Sprintf("%s%-15s %-50s %-20s",
			cursor,
			truncate(typeStr, 13),
			truncate(notif.Message, 48),
			timeStr,
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Enter] View  [D] Dismiss  [A] Mark All Read  [R] Refresh  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderNotificationTypeSelect() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Render("Select Notification Type:") + "\n\n")

	for i, typeStr := range m.notifications.availableTypes {
		style := lipgloss.NewStyle()
		if i == m.notifications.typeSelectIndex {
			style = style.Background(lipgloss.Color("235")).Foreground(lipgloss.Color("11"))
		}

		cursor := "  "
		if i == m.notifications.typeSelectIndex {
			cursor = "→ "
		}

		// Format type name
		displayName := strings.ReplaceAll(typeStr, "_", " ")
		displayName = strings.Title(displayName)

		b.WriteString(style.Render(fmt.Sprintf("%s%s\n", cursor, displayName)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Enter] Select  [Esc] Back\n"))

	return b.String()
}

func (m *Model) renderNotificationView() string {
	var b strings.Builder

	if m.notifications.currentNotif == nil {
		return "No notification selected"
	}

	notif := m.notifications.currentNotif

	viewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(80)

	var content strings.Builder

	// Type badge
	typeStyle := lipgloss.NewStyle().Bold(true)
	switch notif.Type {
	case models.NotificationTypeFriendRequest:
		typeStyle = typeStyle.Foreground(lipgloss.Color("14"))
	case models.NotificationTypeMail:
		typeStyle = typeStyle.Foreground(lipgloss.Color("10"))
	case models.NotificationTypePvPChallenge:
		typeStyle = typeStyle.Foreground(lipgloss.Color("9"))
	case models.NotificationTypeTerritoryAttack:
		typeStyle = typeStyle.Foreground(lipgloss.Color("9"))
	case models.NotificationTypeAchievement:
		typeStyle = typeStyle.Foreground(lipgloss.Color("226"))
	default:
		typeStyle = typeStyle.Foreground(lipgloss.Color("11"))
	}

	content.WriteString(typeStyle.Render(fmt.Sprintf("[%s]", notif.Type)) + "\n\n")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(notif.Title) + "\n\n")
	content.WriteString(notif.Message + "\n\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(fmt.Sprintf("Received: %s", notif.CreatedAt.Format("2006-01-02 15:04"))) + "\n")

	if notif.ExpiresAt.After(notif.CreatedAt) {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(fmt.Sprintf("Expires: %s", notif.ExpiresAt.Format("2006-01-02 15:04"))) + "\n")
	}

	// Show action data if present
	if len(notif.ActionData) > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Render("Additional Info:") + "\n")
		for key, value := range notif.ActionData {
			content.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	b.WriteString(viewStyle.Render(content.String()))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[D] Dismiss  [Q] Back\n"))

	return b.String()
}

// Commands

func (m *Model) loadAllNotifications() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to load all notifications
		var notifications []models.Notification
		var unreadCount int
		var errStr string

		if m.notificationsManager != nil {
			var err error
			notifications, err = m.notificationsManager.GetNotifications(ctx, m.playerID, 100)
			if err != nil {
				errStr = err.Error()
			}

			// Get unread count
			unreadCount, _ = m.notificationsManager.GetUnreadCount(ctx, m.playerID)
		} else {
			notifications = []models.Notification{}
		}

		return notificationsLoadedMsg{
			notifications: notifications,
			unreadCount:   unreadCount,
			err:           errStr,
		}
	}
}

func (m *Model) loadUnreadNotifications() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to load unread notifications
		var notifications []models.Notification
		var unreadCount int
		var errStr string

		if m.notificationsManager != nil {
			var err error
			notifications, err = m.notificationsManager.GetUnreadNotifications(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}

			unreadCount = len(notifications)
		} else {
			notifications = []models.Notification{}
		}

		return notificationsLoadedMsg{
			notifications: notifications,
			unreadCount:   unreadCount,
			err:           errStr,
		}
	}
}

func (m *Model) loadNotificationsByType(notifType string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to load notifications by type
		var notifications []models.Notification
		var unreadCount int
		var errStr string

		if m.notificationsManager != nil {
			var err error
			notifications, err = m.notificationsManager.GetNotificationsByType(ctx, m.playerID, notifType)
			if err != nil {
				errStr = err.Error()
			}

			// Get unread count
			unreadCount, _ = m.notificationsManager.GetUnreadCount(ctx, m.playerID)
		} else {
			notifications = []models.Notification{}
		}

		return notificationsLoadedMsg{
			notifications: notifications,
			unreadCount:   unreadCount,
			err:           errStr,
		}
	}
}

func (m *Model) markNotificationAsRead(notifID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to mark as read
		var errStr string
		if m.notificationsManager != nil {
			err := m.notificationsManager.MarkAsRead(ctx, notifID, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return notificationActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) dismissNotification(notifID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to dismiss notification
		var errStr string
		if m.notificationsManager != nil {
			err := m.notificationsManager.DismissNotification(ctx, notifID, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return notificationActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) markAllNotificationsAsRead() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to mark all as read
		var errStr string
		if m.notificationsManager != nil {
			err := m.notificationsManager.MarkAllAsRead(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return notificationActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) clearDismissedNotifications() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use notifications manager to dismiss all read notifications
		var errStr string
		if m.notificationsManager != nil {
			err := m.notificationsManager.DismissAllRead(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return notificationActionMsg{
			err: errStr,
		}
	}
}

// ===== Messages =====
// Custom message types for notifications screen async operations

// notificationsLoadedMsg is sent when notifications have been loaded from the manager.
//
// This message is returned by:
//   - loadAllNotifications()
//   - loadUnreadNotifications()
//   - loadNotificationsByType()
type notificationsLoadedMsg struct {
	notifications []models.Notification // Loaded notifications
	unreadCount   int                   // Total unread count
	err           string                // Error message (if loading failed)
}

// notificationActionMsg is sent when a notification action has completed.
//
// Actions include:
//   - Mark as read
//   - Dismiss notification
//   - Mark all as read
//   - Clear dismissed notifications
type notificationActionMsg struct {
	err string // Error message (if action failed)
}
