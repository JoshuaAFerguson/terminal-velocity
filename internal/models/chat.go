// File: internal/models/chat.go
// Project: Terminal Velocity
// Description: Chat message models and channel types for multiplayer communication
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ChatChannel represents different types of chat channels
type ChatChannel string

const (
	ChatChannelGlobal  ChatChannel = "global"  // All online players
	ChatChannelSystem  ChatChannel = "system"  // Players in same system
	ChatChannelFaction ChatChannel = "faction" // Faction members only
	ChatChannelDirect  ChatChannel = "direct"  // Private 1-on-1 messages
	ChatChannelTrade   ChatChannel = "trade"   // Trade-related messages
	ChatChannelCombat  ChatChannel = "combat"  // Combat notifications
)

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID        uuid.UUID   `json:"id"`         // Unique message identifier
	Channel   ChatChannel `json:"channel"`    // Which channel this message is in
	SenderID  uuid.UUID   `json:"sender_id"`  // Player who sent the message
	Sender    string      `json:"sender"`     // Sender's username
	Recipient string      `json:"recipient,omitempty"` // For direct messages
	Content   string      `json:"content"`    // Message text
	Timestamp time.Time   `json:"timestamp"`  // When the message was sent

	// Context information
	SystemID  uuid.UUID `json:"system_id,omitempty"`  // For system chat
	FactionID string    `json:"faction_id,omitempty"` // For faction chat

	// Message metadata
	IsSystem  bool   `json:"is_system"`  // Is this a system message (not from player)?
	Color     string `json:"color,omitempty"` // Optional color for formatting
}

// NewChatMessage creates a new chat message
func NewChatMessage(channel ChatChannel, senderID uuid.UUID, sender string, content string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		Channel:   channel,
		SenderID:  senderID,
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
		IsSystem:  false,
	}
}

// NewSystemMessage creates a system-generated message
func NewSystemMessage(channel ChatChannel, content string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		Channel:   channel,
		Sender:    "System",
		Content:   content,
		Timestamp: time.Now(),
		IsSystem:  true,
		Color:     "yellow",
	}
}

// NewDirectMessage creates a direct message to a specific player
func NewDirectMessage(senderID uuid.UUID, sender string, recipient string, content string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		Channel:   ChatChannelDirect,
		SenderID:  senderID,
		Sender:    sender,
		Recipient: recipient,
		Content:   content,
		Timestamp: time.Now(),
		IsSystem:  false,
	}
}

// GetTimestampString returns a formatted timestamp for display
func (m *ChatMessage) GetTimestampString() string {
	return m.Timestamp.Format("15:04:05")
}

// GetAgeString returns how long ago the message was sent
func (m *ChatMessage) GetAgeString() string {
	duration := time.Since(m.Timestamp)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	return m.Timestamp.Format("Jan 2")
}

// FormatMessage returns the message formatted for display
func (m *ChatMessage) FormatMessage() string {
	timestamp := m.GetTimestampString()

	if m.IsSystem {
		return fmt.Sprintf("[%s] * %s", timestamp, m.Content)
	}

	switch m.Channel {
	case ChatChannelGlobal:
		return fmt.Sprintf("[%s] [Global] %s: %s", timestamp, m.Sender, m.Content)

	case ChatChannelSystem:
		return fmt.Sprintf("[%s] [System] %s: %s", timestamp, m.Sender, m.Content)

	case ChatChannelFaction:
		return fmt.Sprintf("[%s] [Faction] %s: %s", timestamp, m.Sender, m.Content)

	case ChatChannelDirect:
		if m.Recipient != "" {
			return fmt.Sprintf("[%s] [DM to %s] %s: %s", timestamp, m.Recipient, m.Sender, m.Content)
		}
		return fmt.Sprintf("[%s] [DM from %s] %s: %s", timestamp, m.Sender, m.Sender, m.Content)

	case ChatChannelTrade:
		return fmt.Sprintf("[%s] [Trade] %s: %s", timestamp, m.Sender, m.Content)

	case ChatChannelCombat:
		return fmt.Sprintf("[%s] [Combat] %s", timestamp, m.Content)

	default:
		return fmt.Sprintf("[%s] %s: %s", timestamp, m.Sender, m.Content)
	}
}

// GetChannelDisplayName returns a human-readable channel name
func GetChannelDisplayName(channel ChatChannel) string {
	names := map[ChatChannel]string{
		ChatChannelGlobal:  "Global Chat",
		ChatChannelSystem:  "System Chat",
		ChatChannelFaction: "Faction Chat",
		ChatChannelDirect:  "Direct Messages",
		ChatChannelTrade:   "Trade Channel",
		ChatChannelCombat:  "Combat Log",
	}

	if name, exists := names[channel]; exists {
		return name
	}
	return string(channel)
}

// GetChannelIcon returns an emoji icon for the channel
func GetChannelIcon(channel ChatChannel) string {
	icons := map[ChatChannel]string{
		ChatChannelGlobal:  "🌍",
		ChatChannelSystem:  "📍",
		ChatChannelFaction: "🏛️",
		ChatChannelDirect:  "💬",
		ChatChannelTrade:   "💰",
		ChatChannelCombat:  "⚔️",
	}

	if icon, exists := icons[channel]; exists {
		return icon
	}
	return "💬"
}

// ChatHistory represents a player's chat history
type ChatHistory struct {
	PlayerID     uuid.UUID      `json:"player_id"`
	GlobalChat   []*ChatMessage `json:"global_chat"`
	SystemChat   []*ChatMessage `json:"system_chat"`
	FactionChat  []*ChatMessage `json:"faction_chat"`
	DirectChats  map[string][]*ChatMessage `json:"direct_chats"` // Key: other player's username
	TradeChat    []*ChatMessage `json:"trade_chat"`
	CombatLog    []*ChatMessage `json:"combat_log"`

	// Limits
	MaxMessagesPerChannel int `json:"max_messages_per_channel"`
}

// NewChatHistory creates a new chat history for a player
func NewChatHistory(playerID uuid.UUID) *ChatHistory {
	return &ChatHistory{
		PlayerID:              playerID,
		GlobalChat:            []*ChatMessage{},
		SystemChat:            []*ChatMessage{},
		FactionChat:           []*ChatMessage{},
		DirectChats:           make(map[string][]*ChatMessage),
		TradeChat:             []*ChatMessage{},
		CombatLog:             []*ChatMessage{},
		MaxMessagesPerChannel: 100, // Keep last 100 messages per channel
	}
}

// AddMessage adds a message to the appropriate channel history
func (h *ChatHistory) AddMessage(msg *ChatMessage) {
	switch msg.Channel {
	case ChatChannelGlobal:
		h.GlobalChat = append(h.GlobalChat, msg)
		h.trimChannel(&h.GlobalChat)

	case ChatChannelSystem:
		h.SystemChat = append(h.SystemChat, msg)
		h.trimChannel(&h.SystemChat)

	case ChatChannelFaction:
		h.FactionChat = append(h.FactionChat, msg)
		h.trimChannel(&h.FactionChat)

	case ChatChannelDirect:
		// Determine the other player's username
		otherPlayer := msg.Recipient
		if msg.Recipient == "" {
			otherPlayer = msg.Sender
		}

		if _, exists := h.DirectChats[otherPlayer]; !exists {
			h.DirectChats[otherPlayer] = []*ChatMessage{}
		}

		h.DirectChats[otherPlayer] = append(h.DirectChats[otherPlayer], msg)
		h.trimDirectChannel(otherPlayer)

	case ChatChannelTrade:
		h.TradeChat = append(h.TradeChat, msg)
		h.trimChannel(&h.TradeChat)

	case ChatChannelCombat:
		h.CombatLog = append(h.CombatLog, msg)
		h.trimChannel(&h.CombatLog)
	}
}

// trimChannel keeps only the most recent messages
func (h *ChatHistory) trimChannel(channel *[]*ChatMessage) {
	if len(*channel) > h.MaxMessagesPerChannel {
		*channel = (*channel)[len(*channel)-h.MaxMessagesPerChannel:]
	}
}

// trimDirectChannel trims a specific direct message channel
func (h *ChatHistory) trimDirectChannel(otherPlayer string) {
	if messages, exists := h.DirectChats[otherPlayer]; exists {
		if len(messages) > h.MaxMessagesPerChannel {
			h.DirectChats[otherPlayer] = messages[len(messages)-h.MaxMessagesPerChannel:]
		}
	}
}

// GetMessages returns messages for a specific channel
func (h *ChatHistory) GetMessages(channel ChatChannel, limit int) []*ChatMessage {
	var messages []*ChatMessage

	switch channel {
	case ChatChannelGlobal:
		messages = h.GlobalChat
	case ChatChannelSystem:
		messages = h.SystemChat
	case ChatChannelFaction:
		messages = h.FactionChat
	case ChatChannelTrade:
		messages = h.TradeChat
	case ChatChannelCombat:
		messages = h.CombatLog
	default:
		return []*ChatMessage{}
	}

	// Return last N messages
	if limit > 0 && len(messages) > limit {
		return messages[len(messages)-limit:]
	}

	return messages
}

// GetDirectMessages returns messages for a specific direct message conversation
func (h *ChatHistory) GetDirectMessages(otherPlayer string, limit int) []*ChatMessage {
	messages, exists := h.DirectChats[otherPlayer]
	if !exists {
		return []*ChatMessage{}
	}

	// Return last N messages
	if limit > 0 && len(messages) > limit {
		return messages[len(messages)-limit:]
	}

	return messages
}

// GetUnreadCount returns the number of unread messages in a channel
// (This is a placeholder - actual implementation would need read tracking)
func (h *ChatHistory) GetUnreadCount(channel ChatChannel) int {
	// For now, return 0 - in a full implementation, we'd track read status
	return 0
}

// ClearChannel clears all messages from a specific channel
func (h *ChatHistory) ClearChannel(channel ChatChannel) {
	switch channel {
	case ChatChannelGlobal:
		h.GlobalChat = []*ChatMessage{}
	case ChatChannelSystem:
		h.SystemChat = []*ChatMessage{}
	case ChatChannelFaction:
		h.FactionChat = []*ChatMessage{}
	case ChatChannelTrade:
		h.TradeChat = []*ChatMessage{}
	case ChatChannelCombat:
		h.CombatLog = []*ChatMessage{}
	}
}

// GetActiveDirectChats returns a list of usernames with active DM conversations
func (h *ChatHistory) GetActiveDirectChats() []string {
	chats := []string{}
	for username := range h.DirectChats {
		chats = append(chats, username)
	}
	return chats
}
