// File: internal/tui/mail.go
// Project: Terminal Velocity
// Description: Player mail system TUI screen
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

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

// Mail screen modes
const (
	mailModeInbox   = "inbox"
	mailModeSent    = "sent"
	mailModeCompose = "compose"
	mailModeRead    = "read"
)

// Mail screen state
type mailState struct {
	mode          string
	inbox         []*models.Mail
	sent          []*models.Mail
	selectedIndex int
	currentMail   *models.Mail
	unreadCount   int
	loading       bool
	error         string

	// Compose state
	recipientInput string
	subjectInput   string
	bodyInput      string
	composeField   int // 0=recipient, 1=subject, 2=body
}

func (m *Model) updateMail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle compose mode separately
		if m.mail.mode == mailModeCompose {
			return m.updateMailCompose(msg)
		}

		// Handle read mode separately
		if m.mail.mode == mailModeRead {
			return m.updateMailRead(msg)
		}

		// Inbox/Sent mode navigation
		switch msg.String() {
		case "q", "esc":
			m.screen = ScreenGame
			return m, nil

		case "1":
			// Switch to inbox
			m.mail.mode = mailModeInbox
			m.mail.selectedIndex = 0
			m.mail.loading = true
			return m, m.loadInbox()

		case "2":
			// Switch to sent
			m.mail.mode = mailModeSent
			m.mail.selectedIndex = 0
			m.mail.loading = true
			return m, m.loadSent()

		case "c":
			// Compose new mail
			m.mail.mode = mailModeCompose
			m.mail.recipientInput = ""
			m.mail.subjectInput = ""
			m.mail.bodyInput = ""
			m.mail.composeField = 0
			return m, nil

		case "up", "k":
			if m.mail.mode == mailModeInbox && len(m.mail.inbox) > 0 {
				if m.mail.selectedIndex > 0 {
					m.mail.selectedIndex--
				}
			} else if m.mail.mode == mailModeSent && len(m.mail.sent) > 0 {
				if m.mail.selectedIndex > 0 {
					m.mail.selectedIndex--
				}
			}

		case "down", "j":
			if m.mail.mode == mailModeInbox && len(m.mail.inbox) > 0 {
				if m.mail.selectedIndex < len(m.mail.inbox)-1 {
					m.mail.selectedIndex++
				}
			} else if m.mail.mode == mailModeSent && len(m.mail.sent) > 0 {
				if m.mail.selectedIndex < len(m.mail.sent)-1 {
					m.mail.selectedIndex++
				}
			}

		case "enter":
			// Open selected mail
			if m.mail.mode == mailModeInbox && len(m.mail.inbox) > 0 {
				m.mail.currentMail = m.mail.inbox[m.mail.selectedIndex]
				m.mail.mode = mailModeRead
				return m, m.markMailAsRead(m.mail.currentMail.ID)
			} else if m.mail.mode == mailModeSent && len(m.mail.sent) > 0 {
				m.mail.currentMail = m.mail.sent[m.mail.selectedIndex]
				m.mail.mode = mailModeRead
				return m, nil
			}

		case "d":
			// Delete selected mail
			if m.mail.mode == mailModeInbox && len(m.mail.inbox) > 0 {
				mailID := m.mail.inbox[m.mail.selectedIndex].ID
				return m, m.deleteMail(mailID, true) // true = reload inbox
			} else if m.mail.mode == mailModeSent && len(m.mail.sent) > 0 {
				mailID := m.mail.sent[m.mail.selectedIndex].ID
				return m, m.deleteMail(mailID, false) // false = reload sent
			}

		case "r":
			// Refresh current view
			if m.mail.mode == mailModeInbox {
				m.mail.loading = true
				return m, m.loadInbox()
			} else if m.mail.mode == mailModeSent {
				m.mail.loading = true
				return m, m.loadSent()
			}
		}

	case mailLoadedMsg:
		m.mail.loading = false
		if msg.inbox {
			m.mail.inbox = msg.messages
			m.mail.unreadCount = msg.unreadCount
		} else {
			m.mail.sent = msg.messages
		}
		m.mail.error = msg.err
		if m.mail.selectedIndex >= len(msg.messages) {
			m.mail.selectedIndex = 0
		}

	case mailSentMsg:
		m.mail.loading = false
		if msg.err == "" {
			// Success - return to inbox
			m.mail.mode = mailModeInbox
			m.mail.loading = true
			return m, m.loadInbox()
		}
		m.mail.error = msg.err

	case mailDeletedMsg:
		m.mail.loading = false
		if msg.err == "" {
			// Reload appropriate view
			if msg.reloadInbox {
				return m, m.loadInbox()
			}
			return m, m.loadSent()
		}
		m.mail.error = msg.err
	}

	return m, nil
}

func (m *Model) updateMailCompose(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel compose
		m.mail.mode = mailModeInbox
		return m, nil

	case "tab":
		// Cycle through fields
		m.mail.composeField = (m.mail.composeField + 1) % 3

	case "ctrl+s":
		// Send mail
		if m.mail.recipientInput == "" {
			m.mail.error = "Recipient is required"
			return m, nil
		}
		if m.mail.subjectInput == "" {
			m.mail.error = "Subject is required"
			return m, nil
		}
		if m.mail.bodyInput == "" {
			m.mail.error = "Body is required"
			return m, nil
		}

		m.mail.loading = true
		m.mail.error = ""
		return m, m.sendMail(m.mail.recipientInput, m.mail.subjectInput, m.mail.bodyInput)

	case "backspace":
		switch m.mail.composeField {
		case 0: // recipient
			if len(m.mail.recipientInput) > 0 {
				m.mail.recipientInput = m.mail.recipientInput[:len(m.mail.recipientInput)-1]
			}
		case 1: // subject
			if len(m.mail.subjectInput) > 0 {
				m.mail.subjectInput = m.mail.subjectInput[:len(m.mail.subjectInput)-1]
			}
		case 2: // body
			if len(m.mail.bodyInput) > 0 {
				m.mail.bodyInput = m.mail.bodyInput[:len(m.mail.bodyInput)-1]
			}
		}

	default:
		// Add character to current field
		if len(msg.String()) == 1 {
			switch m.mail.composeField {
			case 0: // recipient
				if len(m.mail.recipientInput) < 32 {
					m.mail.recipientInput += msg.String()
				}
			case 1: // subject
				if len(m.mail.subjectInput) < 200 {
					m.mail.subjectInput += msg.String()
				}
			case 2: // body
				if len(m.mail.bodyInput) < 5000 {
					m.mail.bodyInput += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m *Model) updateMailRead(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Return to previous view
		if m.mail.currentMail.ReceiverID == m.playerID {
			m.mail.mode = mailModeInbox
		} else {
			m.mail.mode = mailModeSent
		}
		m.mail.currentMail = nil
		return m, nil

	case "d":
		// Delete current mail
		mailID := m.mail.currentMail.ID
		isInbox := m.mail.currentMail.ReceiverID == m.playerID
		m.mail.mode = mailModeInbox
		m.mail.currentMail = nil
		return m, m.deleteMail(mailID, isInbox)

	case "r":
		// Reply (only if in inbox)
		if m.mail.currentMail.ReceiverID == m.playerID {
			// Get sender username
			ctx := context.Background()
			sender, err := m.playerRepo.GetByID(ctx, *m.mail.currentMail.SenderID)
			if err == nil {
				m.mail.mode = mailModeCompose
				m.mail.recipientInput = sender.Username
				m.mail.subjectInput = "Re: " + m.mail.currentMail.Subject
				m.mail.bodyInput = ""
				m.mail.composeField = 2 // Start in body
			}
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) viewMail() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("═══ PLAYER MAIL ═══")

	b.WriteString(title + "\n\n")

	if m.mail.loading {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("Loading...") + "\n\n")
		return b.String()
	}

	if m.mail.error != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("Error: "+m.mail.error) + "\n\n")
	}

	// Mode selector
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render("Folders:") + " ")

	if m.mail.mode == mailModeInbox {
		b.WriteString(fmt.Sprintf("[1] Inbox (%d unread)  ", m.mail.unreadCount))
	} else {
		b.WriteString("[1] Inbox  ")
	}

	if m.mail.mode == mailModeSent {
		b.WriteString("[2] Sent  ")
	} else {
		b.WriteString("[2] Sent  ")
	}

	b.WriteString("[C] Compose\n\n")

	// Render based on mode
	switch m.mail.mode {
	case mailModeInbox:
		b.WriteString(m.renderInbox())
	case mailModeSent:
		b.WriteString(m.renderSent())
	case mailModeCompose:
		b.WriteString(m.renderCompose())
	case mailModeRead:
		b.WriteString(m.renderRead())
	}

	return b.String()
}

func (m *Model) renderInbox() string {
	var b strings.Builder

	if len(m.mail.inbox) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No messages in inbox.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[C] Compose  [Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-20s %-40s %-20s\n", "From", "Subject", "Date")))
	b.WriteString(strings.Repeat("─", 80) + "\n")

	// Messages (show max 15)
	maxDisplay := 15
	if len(m.mail.inbox) < maxDisplay {
		maxDisplay = len(m.mail.inbox)
	}

	for i := 0; i < maxDisplay; i++ {
		msg := m.mail.inbox[i]

		style := lipgloss.NewStyle()
		if !msg.IsRead {
			style = style.Foreground(lipgloss.Color("11")).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("7"))
		}

		if i == m.mail.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.mail.selectedIndex {
			cursor = "→ "
		}

		// Get sender username
		ctx := context.Background()
		sender, err := m.playerRepo.GetByID(ctx, *msg.SenderID)
		senderName := "Unknown"
		if err == nil {
			senderName = sender.Username
		}

		timeStr := msg.SentAt.Format("2006-01-02 15:04")
		line := fmt.Sprintf("%s%-20s %-40s %-20s",
			cursor,
			truncate(senderName, 18),
			truncate(msg.Subject, 38),
			timeStr,
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Enter] Read  [D] Delete  [R] Refresh  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderSent() string {
	var b strings.Builder

	if len(m.mail.sent) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No sent messages.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[C] Compose  [Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-20s %-40s %-20s\n", "To", "Subject", "Date")))
	b.WriteString(strings.Repeat("─", 80) + "\n")

	// Messages (show max 15)
	maxDisplay := 15
	if len(m.mail.sent) < maxDisplay {
		maxDisplay = len(m.mail.sent)
	}

	for i := 0; i < maxDisplay; i++ {
		msg := m.mail.sent[i]

		style := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		if i == m.mail.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.mail.selectedIndex {
			cursor = "→ "
		}

		// Get recipient username
		ctx := context.Background()
		recipient, err := m.playerRepo.GetByID(ctx, msg.ReceiverID)
		recipientName := "Unknown"
		if err == nil {
			recipientName = recipient.Username
		}

		timeStr := msg.SentAt.Format("2006-01-02 15:04")
		line := fmt.Sprintf("%s%-20s %-40s %-20s",
			cursor,
			truncate(recipientName, 18),
			truncate(msg.Subject, 38),
			timeStr,
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Enter] Read  [D] Delete  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderCompose() string {
	var b strings.Builder

	composeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(80)

	var compose strings.Builder
	compose.WriteString(lipgloss.NewStyle().Bold(true).Render("Compose New Message") + "\n\n")

	// Recipient field
	recipientStyle := lipgloss.NewStyle()
	if m.mail.composeField == 0 {
		recipientStyle = recipientStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	compose.WriteString(recipientStyle.Render(fmt.Sprintf("To:      %s_\n", m.mail.recipientInput)))

	// Subject field
	subjectStyle := lipgloss.NewStyle()
	if m.mail.composeField == 1 {
		subjectStyle = subjectStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	compose.WriteString(subjectStyle.Render(fmt.Sprintf("Subject: %s_\n", m.mail.subjectInput)))

	compose.WriteString("\n")

	// Body field
	bodyStyle := lipgloss.NewStyle()
	if m.mail.composeField == 2 {
		bodyStyle = bodyStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	compose.WriteString(bodyStyle.Render("Message:\n"))
	bodyLines := wrapText(m.mail.bodyInput+"_", 76)
	for _, line := range bodyLines {
		compose.WriteString(bodyStyle.Render(line) + "\n")
	}

	b.WriteString(composeStyle.Render(compose.String()))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[Tab] Next Field  [Ctrl+S] Send  [Esc] Cancel\n"))

	return b.String()
}

func (m *Model) renderRead() string {
	var b strings.Builder

	if m.mail.currentMail == nil {
		return "No message selected"
	}

	msg := m.mail.currentMail

	readStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(80)

	var content strings.Builder

	// Get sender/recipient username
	ctx := context.Background()
	var otherPlayer *models.Player
	var err error
	if msg.ReceiverID == m.playerID {
		otherPlayer, err = m.playerRepo.GetByID(ctx, *msg.SenderID)
	} else {
		otherPlayer, err = m.playerRepo.GetByID(ctx, msg.ReceiverID)
	}
	otherName := "Unknown"
	if err == nil {
		otherName = otherPlayer.Username
	}

	if msg.ReceiverID == m.playerID {
		content.WriteString(fmt.Sprintf("From: %s\n", otherName))
	} else {
		content.WriteString(fmt.Sprintf("To: %s\n", otherName))
	}

	content.WriteString(fmt.Sprintf("Date: %s\n", msg.SentAt.Format("2006-01-02 15:04")))
	content.WriteString(fmt.Sprintf("Subject: %s\n", msg.Subject))
	content.WriteString("\n")

	// Wrap body text
	bodyLines := wrapText(msg.Body, 76)
	for _, line := range bodyLines {
		content.WriteString(line + "\n")
	}

	b.WriteString(readStyle.Render(content.String()))
	b.WriteString("\n\n")

	if msg.ReceiverID == m.playerID {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[R] Reply  [D] Delete  [Q] Back\n"))
	} else {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[D] Delete  [Q] Back\n"))
	}

	return b.String()
}

// Commands

func (m *Model) loadInbox() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		messages, err := m.mailManager.GetInbox(ctx, m.playerID, 50)
		unreadCount, _ := m.mailManager.GetUnreadCount(ctx, m.playerID)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		// Convert []models.Mail to []*models.Mail
		messagePtrs := make([]*models.Mail, len(messages))
		for i := range messages {
			messagePtrs[i] = &messages[i]
		}

		return mailLoadedMsg{
			messages:    messagePtrs,
			inbox:       true,
			unreadCount: unreadCount,
			err:         errStr,
		}
	}
}

func (m *Model) loadSent() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		messages, err := m.mailRepo.GetSent(ctx, m.playerID, 50, 0)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		return mailLoadedMsg{
			messages: messages,
			inbox:    false,
			err:      errStr,
		}
	}
}

func (m *Model) sendMail(recipient, subject, body string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Wrapper for player lookup function to match expected signature
		getPlayerByUsername := func(username string) (*models.Player, error) {
			return m.playerRepo.GetByUsername(ctx, username)
		}

		// Check if sender is blocked by receiver using friends manager
		checkBlocked := func(receiverID, senderID uuid.UUID) (bool, error) {
			if m.friendsManager == nil {
				// Friends manager not available, allow mail
				return false, nil
			}

			// Check if senderID is blocked by receiverID
			blocked, err := m.friendsManager.IsBlocked(ctx, receiverID, senderID)
			if err != nil {
				// On error, default to not blocked to avoid blocking legitimate mail
				return false, nil
			}
			return blocked, nil
		}

		err := m.mailManager.SendMail(
			ctx,
			&m.playerID,
			m.player.Username,
			recipient,
			subject,
			body,
			0,           // No credits attached
			[]uuid.UUID{}, // No items attached (TODO: Add UI for item attachments)
			getPlayerByUsername,
			checkBlocked,
		)
		if err != nil {
			return mailSentMsg{err: err.Error()}
		}

		return mailSentMsg{err: ""}
	}
}

func (m *Model) markMailAsRead(mailID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		m.mailManager.MarkAsRead(ctx, mailID, m.playerID)
		return nil
	}
}

func (m *Model) deleteMail(mailID uuid.UUID, reloadInbox bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		err := m.mailManager.DeleteMail(ctx, mailID, m.playerID)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		return mailDeletedMsg{
			reloadInbox: reloadInbox,
			err:         errStr,
		}
	}
}

// Messages

type mailLoadedMsg struct {
	messages    []*models.Mail
	inbox       bool
	unreadCount int
	err         string
}

type mailSentMsg struct {
	err string
}

type mailDeletedMsg struct {
	reloadInbox bool
	err         string
}
