// File: internal/tui/chat.go
// Project: Terminal Velocity
// Description: Chat UI for multiplayer communication across multiple channels
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/validation"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type chatModel struct {
	currentChannel   models.ChatChannel
	inputBuffer      string
	dmRecipient      string // For direct messages
	scrollOffset     int
	inputMode        bool     // true when typing a message
	selectedDMChat   int      // Index of selected DM conversation
	availableDMChats []string // List of active DM conversations
}

func newChatModel() chatModel {
	return chatModel{
		currentChannel:   models.ChatChannelGlobal,
		inputBuffer:      "",
		dmRecipient:      "",
		scrollOffset:     0,
		inputMode:        false,
		selectedDMChat:   0,
		availableDMChats: []string{},
	}
}

func (m Model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Input mode - typing a message
		if m.chatModel.inputMode {
			switch msg.String() {
			case "esc":
				// Cancel input
				m.chatModel.inputMode = false
				m.chatModel.inputBuffer = ""
				return m, nil

			case "enter":
				// Send message
				if len(m.chatModel.inputBuffer) > 0 {
					m.sendChatMessage()
					m.chatModel.inputBuffer = ""
				}
				m.chatModel.inputMode = false
				return m, nil

			case "backspace":
				if len(m.chatModel.inputBuffer) > 0 {
					m.chatModel.inputBuffer = m.chatModel.inputBuffer[:len(m.chatModel.inputBuffer)-1]
				}

			default:
				// Add character to buffer with sanitization
				if len(msg.String()) == 1 && len(m.chatModel.inputBuffer) < 200 {
					// Filter out control characters (except escape for ANSI sequences)
					char := msg.String()[0]
					if (char >= 32 && char != 127) || char == 27 {
						m.chatModel.inputBuffer += msg.String()
						// Strip ANSI escape codes to prevent injection attacks
						m.chatModel.inputBuffer = validation.StripANSI(m.chatModel.inputBuffer)
					}
				}
			}

			return m, nil
		}

		// Normal mode - navigation and commands
		switch msg.String() {
		case "esc", "backspace", "q":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.chatModel.scrollOffset > 0 {
				m.chatModel.scrollOffset--
			}

		case "down", "j":
			m.chatModel.scrollOffset++

		case "i", "enter":
			// Enter input mode
			m.chatModel.inputMode = true
			m.chatModel.inputBuffer = ""
			return m, nil

		// Channel shortcuts
		case "1":
			m.chatModel.currentChannel = models.ChatChannelGlobal
			m.chatModel.scrollOffset = 0

		case "2":
			m.chatModel.currentChannel = models.ChatChannelSystem
			m.chatModel.scrollOffset = 0

		case "3":
			m.chatModel.currentChannel = models.ChatChannelFaction
			m.chatModel.scrollOffset = 0

		case "4":
			m.chatModel.currentChannel = models.ChatChannelDirect
			m.chatModel.scrollOffset = 0
			// Load DM conversations
			m.chatModel.availableDMChats = m.chatManager.GetActiveDirectChats(m.playerID)

		case "5":
			m.chatModel.currentChannel = models.ChatChannelTrade
			m.chatModel.scrollOffset = 0

		case "c":
			// Clear current channel
			m.chatManager.ClearChannel(m.playerID, m.chatModel.currentChannel)
			m.chatModel.scrollOffset = 0
		}
	}

	return m, nil
}

func (m Model) viewChat() string {
	icon := models.GetChannelIcon(m.chatModel.currentChannel)
	displayName := models.GetChannelDisplayName(m.chatModel.currentChannel)

	s := titleStyle.Render(icon+" CHAT - "+displayName) + "\n\n"

	// Channel tabs
	tabs := []struct {
		key     string
		label   string
		channel models.ChatChannel
	}{
		{"1", "Global", models.ChatChannelGlobal},
		{"2", "System", models.ChatChannelSystem},
		{"3", "Faction", models.ChatChannelFaction},
		{"4", "DMs", models.ChatChannelDirect},
		{"5", "Trade", models.ChatChannelTrade},
	}

	s += "Channels: "
	for i, tab := range tabs {
		isActive := m.chatModel.currentChannel == tab.channel

		if isActive {
			s += highlightStyle.Render("[" + tab.label + "]")
		} else {
			s += helpStyle.Render(" " + tab.label + " ")
		}

		if i < len(tabs)-1 {
			s += " "
		}
	}
	s += "\n"
	s += strings.Repeat("─", 80) + "\n\n"

	// Message display area
	s += m.renderChatMessages()

	// Input area
	s += "\n" + strings.Repeat("─", 80) + "\n"

	if m.chatModel.inputMode {
		prompt := "Message: "
		if m.chatModel.currentChannel == models.ChatChannelDirect && m.chatModel.dmRecipient != "" {
			prompt = fmt.Sprintf("To %s: ", m.chatModel.dmRecipient)
		}

		s += highlightStyle.Render(prompt) + m.chatModel.inputBuffer + "█\n"
		s += helpStyle.Render("Enter: Send | ESC: Cancel")
	} else {
		s += "Press I or Enter to send a message\n"
		s += renderFooter("I/Enter: Message | 1-5: Channels | C: Clear | ↑/↓: Scroll | ESC: Back")
	}

	return s
}

func (m Model) renderChatMessages() string {
	var messages []*models.ChatMessage

	if m.chatModel.currentChannel == models.ChatChannelDirect {
		// Show DM conversations list or specific conversation
		if m.chatModel.dmRecipient == "" {
			return m.renderDMConversationsList()
		}

		messages = m.chatManager.GetDirectMessages(m.playerID, m.chatModel.dmRecipient, 50)
	} else {
		messages = m.chatManager.GetMessages(m.playerID, m.chatModel.currentChannel, 50)
	}

	if len(messages) == 0 {
		emptyMsg := "No messages yet. Be the first to say something!"
		if m.chatModel.currentChannel == models.ChatChannelSystem {
			emptyMsg = "No system chat messages.\n\nSystem chat is for players in the same system."
		} else if m.chatModel.currentChannel == models.ChatChannelFaction {
			emptyMsg = "No faction messages.\n\nJoin or create a faction to use faction chat."
		} else if m.chatModel.currentChannel == models.ChatChannelTrade {
			emptyMsg = "No trade messages.\n\nUse this channel to advertise trades and negotiate deals."
		}

		return helpStyle.Render(emptyMsg) + "\n"
	}

	// Calculate visible message range
	displayLines := 15
	startIdx := 0
	if len(messages) > displayLines {
		startIdx = len(messages) - displayLines - m.chatModel.scrollOffset
		if startIdx < 0 {
			startIdx = 0
			m.chatModel.scrollOffset = len(messages) - displayLines
		}
	}

	endIdx := startIdx + displayLines
	if endIdx > len(messages) {
		endIdx = len(messages)
	}

	visibleMessages := messages[startIdx:endIdx]

	// Render messages
	var s strings.Builder
	for _, msg := range visibleMessages {
		formatted := msg.FormatMessage()

		// Apply styling based on message type
		if msg.IsSystem {
			formatted = helpStyle.Render(formatted)
		} else if msg.SenderID == m.playerID {
			formatted = successStyle.Render(formatted)
		}

		s.WriteString(formatted + "\n")
	}

	// Scroll indicator
	if len(messages) > displayLines {
		totalPages := (len(messages) + displayLines - 1) / displayLines
		currentPage := (startIdx / displayLines) + 1
		s.WriteString("\n" + helpStyle.Render(fmt.Sprintf("Page %d/%d (↑/↓ to scroll)", currentPage, totalPages)))
	}

	return s.String()
}

func (m Model) renderDMConversationsList() string {
	chats := m.chatModel.availableDMChats

	if len(chats) == 0 {
		return helpStyle.Render("No direct message conversations yet.\n\nUse /dm <username> <message> to start a conversation.")
	}

	var s strings.Builder
	s.WriteString("Direct Message Conversations:\n\n")

	for i, username := range chats {
		cursor := "  "
		if i == m.chatModel.selectedDMChat {
			cursor = "> "
		}

		messages := m.chatManager.GetDirectMessages(m.playerID, username, 1)
		preview := ""
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			preview = lastMsg.Content
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			preview = " - " + preview
		}

		line := fmt.Sprintf("%s%s%s", cursor, username, preview)
		s.WriteString(line + "\n")
	}

	s.WriteString("\n" + helpStyle.Render("Press Enter to view conversation"))

	return s.String()
}

func (m *Model) sendChatMessage() {
	content := strings.TrimSpace(m.chatModel.inputBuffer)
	if content == "" {
		return
	}

	// Check for commands
	if strings.HasPrefix(content, "/") {
		m.handleChatCommand(content)
		return
	}

	// Send based on current channel
	switch m.chatModel.currentChannel {
	case models.ChatChannelGlobal:
		m.chatManager.SendGlobalMessage(m.playerID, m.username, content)

	case models.ChatChannelSystem:
		// Get players in same system
		if m.player != nil {
			systemPlayers := m.presenceManager.GetPlayersInSystem(m.player.CurrentSystem)
			recipientIDs := []uuid.UUID{m.playerID} // Include self
			for _, p := range systemPlayers {
				recipientIDs = append(recipientIDs, p.PlayerID)
			}
			m.chatManager.SendSystemMessage(m.player.CurrentSystem, m.playerID, m.username, content, recipientIDs)
		}

	case models.ChatChannelFaction:
		// Get player's faction
		faction, err := m.factionManager.GetPlayerFaction(m.playerID)
		if err != nil || faction == nil {
			// Player not in a faction
			msg := models.NewSystemMessage(models.ChatChannelFaction, "You must be in a faction to use faction chat. Join a faction first!")
			history := m.chatManager.GetOrCreateHistory(m.playerID)
			history.AddMessage(msg)
			return
		}

		// Send message to all faction members
		m.chatManager.SendFactionMessage(faction.ID.String(), m.playerID, m.username, content, faction.Members)

	case models.ChatChannelDirect:
		if m.chatModel.dmRecipient != "" {
			// Look up recipient by username
			recipient, err := m.playerRepo.GetByUsername(context.Background(), m.chatModel.dmRecipient)
			if err != nil || recipient == nil {
				msg := models.NewSystemMessage(models.ChatChannelDirect, fmt.Sprintf("Player '%s' not found.", m.chatModel.dmRecipient))
				history := m.chatManager.GetOrCreateHistory(m.playerID)
				history.AddMessage(msg)
				return
			}

			// Check if recipient is online (optional - can send to offline players too)
			// For now, allow sending to offline players (mail-like functionality)

			// Send direct message
			m.chatManager.SendDirectMessage(m.playerID, m.username, recipient.ID, recipient.Username, content)
		}

	case models.ChatChannelTrade:
		m.chatManager.SendTradeMessage(m.playerID, m.username, content)
	}
}

func (m *Model) handleChatCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		helpText := `Chat Commands:
/help - Show this help message
/dm <username> <message> - Send a direct message
/clear - Clear current channel
/me <action> - Send an action message`

		msg := models.NewSystemMessage(m.chatModel.currentChannel, helpText)
		history := m.chatManager.GetOrCreateHistory(m.playerID)
		history.AddMessage(msg)

	case "/dm":
		if len(parts) < 3 {
			msg := models.NewSystemMessage(models.ChatChannelDirect, "Usage: /dm <username> <message>")
			history := m.chatManager.GetOrCreateHistory(m.playerID)
			history.AddMessage(msg)
			return
		}

		recipientUsername := parts[1]
		content := strings.Join(parts[2:], " ")

		// Look up recipient by username
		recipientPlayer, err := m.playerRepo.GetByUsername(context.Background(), recipientUsername)
		if err != nil || recipientPlayer == nil {
			msg := models.NewSystemMessage(models.ChatChannelDirect, fmt.Sprintf("Player '%s' not found.", recipientUsername))
			history := m.chatManager.GetOrCreateHistory(m.playerID)
			history.AddMessage(msg)
			return
		}

		// Send direct message
		m.chatManager.SendDirectMessage(m.playerID, m.username, recipientPlayer.ID, recipientPlayer.Username, content)

		// Confirm to sender
		msg := models.NewSystemMessage(models.ChatChannelDirect, fmt.Sprintf("DM sent to %s", recipientUsername))
		history := m.chatManager.GetOrCreateHistory(m.playerID)
		history.AddMessage(msg)

	case "/clear":
		m.chatManager.ClearChannel(m.playerID, m.chatModel.currentChannel)
		m.chatModel.scrollOffset = 0

	case "/me":
		if len(parts) < 2 {
			return
		}

		action := strings.Join(parts[1:], " ")
		content := fmt.Sprintf("* %s %s", m.username, action)

		switch m.chatModel.currentChannel {
		case models.ChatChannelGlobal:
			m.chatManager.SendGlobalMessage(m.playerID, m.username, content)
		case models.ChatChannelTrade:
			m.chatManager.SendTradeMessage(m.playerID, m.username, content)
		}

	default:
		msg := models.NewSystemMessage(m.chatModel.currentChannel, fmt.Sprintf("Unknown command: %s (type /help for commands)", cmd))
		history := m.chatManager.GetOrCreateHistory(m.playerID)
		history.AddMessage(msg)
	}
}
