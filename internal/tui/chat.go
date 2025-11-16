// File: internal/tui/chat.go
// Project: Terminal Velocity
// Description: Chat screen - Multiplayer communication across multiple channels with commands
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The chat screen provides:
// - 5 distinct chat channels (Global, System, Faction, Direct, Trade)
// - Real-time message broadcasting and receiving
// - Input mode for typing messages (200 char limit)
// - Chat commands (/help, /dm, /clear, /me)
// - Message scrolling (15 lines visible)
// - Direct message conversation management
// - ANSI escape code sanitization for security
// - Channel-specific access control
// - Message history per channel
//
// Channels:
//   - Global: All players server-wide
//   - System: Players in the same star system
//   - Faction: Faction members only (requires faction membership)
//   - Direct: Private 1-on-1 conversations
//   - Trade: Trade-focused channel for all players
//
// Chat Commands:
//   - /help: Display command help
//   - /dm <username> <message>: Send direct message
//   - /clear: Clear current channel history
//   - /me <action>: Send action message (e.g., "*player waves*")
//
// Security:
//   - ANSI escape code stripping to prevent terminal injection
//   - Control character filtering
//   - Message length limits (200 chars)
//   - Muted player filtering
//
// Visual Features:
//   - Sender highlighting (own messages in green)
//   - System messages in gray
//   - Channel tabs with active indicator
//   - Scroll pagination indicator
//   - Input cursor (_) when typing

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

// chatModel contains the state for the chat screen.
// Manages channels, message input, scrolling, and conversation state.
type chatModel struct {
	currentChannel   models.ChatChannel // Currently active channel
	inputBuffer      string             // Current message being typed
	dmRecipient      string             // Target username for direct messages
	scrollOffset     int                // Scroll position in message history
	inputMode        bool               // True when typing a message
	selectedDMChat   int                // Index of selected DM conversation in list
	availableDMChats []string           // List of active DM conversations (usernames)
}

// newChatModel creates and initializes a new chat screen model.
// Starts in Global channel with no active input.
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

// updateChat handles input and state updates for the chat screen.
//
// Key Bindings (Input Mode):
//   - esc: Cancel message input, clear buffer
//   - enter: Send message (or execute command if starts with /)
//   - backspace: Delete character from input buffer
//   - Any printable char: Add to input buffer (with sanitization)
//
// Key Bindings (Normal Mode):
//   - esc/backspace/q: Return to main menu
//   - up/k: Scroll messages up
//   - down/j: Scroll messages down
//   - i/enter: Enter input mode to type message
//   - 1-5: Switch chat channel
//     - 1: Global channel
//     - 2: System channel
//     - 3: Faction channel
//     - 4: Direct messages
//     - 5: Trade channel
//   - c: Clear current channel history
//
// Message Flow:
//   1. User presses 'i' or Enter to start typing
//   2. Type message (up to 200 chars, sanitized)
//   3. Press Enter to send or Esc to cancel
//   4. Command messages (/) handled by handleChatCommand()
//   5. Regular messages sent via sendChatMessage()
//
// Security:
//   - ANSI escape codes stripped via validation.StripANSI()
//   - Control characters filtered (except escape for ANSI)
//   - Buffer limited to 200 characters
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

// viewChat renders the chat screen.
//
// Layout:
//   - Title: Icon + "CHAT - [Channel Name]"
//   - Channel Tabs: 5 channels with active indicator
//   - Separator line
//   - Message Display Area: 15 visible lines with scroll
//   - Separator line
//   - Input Area: Prompt or status message
//   - Footer: Controls help
//
// Visual Features:
//   - Active channel highlighted
//   - Own messages in green (successStyle)
//   - System messages in gray (helpStyle)
//   - Scroll pagination indicator
//   - Input cursor when typing
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

// renderChatMessages renders the message list for the current channel.
// Handles pagination, scrolling, and empty states for each channel type.
// Shows DM conversation list when in Direct channel with no recipient selected.
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

// renderDMConversationsList renders the list of active direct message conversations.
// Shows conversation preview with most recent message excerpt.
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

// sendChatMessage sends a chat message to the current channel.
// Handles channel-specific sending logic and permission checking.
// Commands (starting with /) are routed to handleChatCommand() instead.
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

// handleChatCommand processes chat slash commands.
// Supported commands: /help, /dm, /clear, /me
// Unknown commands show error message in chat.
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
