// File: internal/chat/manager.go
// Project: Terminal Velocity
// Description: Chat manager for multiplayer communication and message routing
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package chat provides multiplayer chat and messaging functionality.
//
// This package handles:
// - Message routing across multiple channels (4 channels)
// - Per-player chat history management
// - Direct messaging between players
// - System notifications and broadcasts
// - Combat log messages
// - Global message retention (last 200 messages)
//
// Chat Channels:
// - Global: Server-wide chat visible to all online players
// - System: System-specific chat (only visible in current star system)
// - Faction: Faction-only chat (visible to faction members)
// - Trade: Trade channel for commerce-related messages
// - Direct: Private messages between two players
// - Combat: Combat notifications and logs
//
// Message Flow:
// 1. Message sent via SendXXXMessage method
// 2. Manager creates ChatMessage with metadata (timestamp, channel, sender)
// 3. Message added to recipient histories
// 4. Global messages cached for late-joining players
//
// Thread Safety:
// All Manager methods are thread-safe using sync.RWMutex. Read operations
// use RLock, write operations use Lock.
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package chat

import (
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles chat message routing and history for all players.
// It maintains per-player message histories and caches global messages
// for late-joining players. All operations are thread-safe.
type Manager struct {
	mu        sync.RWMutex                         // Protects all fields
	histories map[uuid.UUID]*models.ChatHistory    // Per-player chat histories indexed by player ID

	// Message retention for late-joining players
	globalHistory    []*models.ChatMessage // Recent global messages (circular buffer)
	maxGlobalHistory int                   // Maximum global messages to retain (default: 200)
}

// NewManager creates a new chat manager.
//
// Returns:
//   - Pointer to new Manager with 200-message global history retention
//
// Thread Safety:
// Safe to call concurrently, though typically called once at server startup.
func NewManager() *Manager {
	return &Manager{
		histories:        make(map[uuid.UUID]*models.ChatHistory),
		globalHistory:    []*models.ChatMessage{},
		maxGlobalHistory: 200, // Keep last 200 global messages
	}
}

// GetOrCreateHistory gets or creates a chat history for a player.
//
// Creates a new history if player doesn't have one yet. Used when
// players connect to initialize their message buffers.
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Pointer to player's ChatHistory (existing or newly created)
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) GetOrCreateHistory(playerID uuid.UUID) *models.ChatHistory {
	m.mu.Lock()
	defer m.mu.Unlock()

	if history, exists := m.histories[playerID]; exists {
		return history
	}

	history := models.NewChatHistory(playerID)
	m.histories[playerID] = history
	return history
}

// SendGlobalMessage sends a message to all online players.
//
// Message is broadcast to all player histories and cached in global
// history for late-joining players. Maintains a circular buffer of
// last 200 global messages.
//
// Parameters:
//   - senderID: Sender's player UUID
//   - sender: Sender's username for display
//   - content: Message text content
//
// Returns:
//   - Pointer to created ChatMessage
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendGlobalMessage(senderID uuid.UUID, sender string, content string) *models.ChatMessage {
	msg := models.NewChatMessage(models.ChatChannelGlobal, senderID, sender, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to global history
	m.globalHistory = append(m.globalHistory, msg)
	if len(m.globalHistory) > m.maxGlobalHistory {
		m.globalHistory = m.globalHistory[len(m.globalHistory)-m.maxGlobalHistory:]
	}

	// Add to all player histories
	for _, history := range m.histories {
		history.AddMessage(msg)
	}

	return msg
}

// SendSystemMessage sends a message to players in a specific star system.
//
// Only sends to specified recipients (typically players in same system).
//
// Parameters:
//   - systemID: Star system UUID
//   - senderID: Sender's player UUID
//   - sender: Sender's username
//   - content: Message text
//   - recipientIDs: List of player UUIDs to receive message
//
// Returns:
//   - Pointer to created ChatMessage
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendSystemMessage(systemID uuid.UUID, senderID uuid.UUID, sender string, content string, recipientIDs []uuid.UUID) *models.ChatMessage {
	msg := models.NewChatMessage(models.ChatChannelSystem, senderID, sender, content)
	msg.SystemID = systemID

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to specified player histories
	for _, recipientID := range recipientIDs {
		if history, exists := m.histories[recipientID]; exists {
			history.AddMessage(msg)
		}
	}

	return msg
}

// SendFactionMessage sends a message to all faction members.
//
// Only sends to specified members (caller determines membership).
//
// Parameters:
//   - factionID: Faction identifier string
//   - senderID: Sender's player UUID
//   - sender: Sender's username
//   - content: Message text
//   - memberIDs: List of faction member UUIDs
//
// Returns:
//   - Pointer to created ChatMessage
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendFactionMessage(factionID string, senderID uuid.UUID, sender string, content string, memberIDs []uuid.UUID) *models.ChatMessage {
	msg := models.NewChatMessage(models.ChatChannelFaction, senderID, sender, content)
	msg.FactionID = factionID

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to faction member histories
	for _, memberID := range memberIDs {
		if history, exists := m.histories[memberID]; exists {
			history.AddMessage(msg)
		}
	}

	return msg
}

// SendDirectMessage sends a private message to a specific player.
//
// Message is added to both sender and recipient histories. Recipient's
// copy has sender field populated to show it's an incoming message.
//
// Parameters:
//   - senderID: Sender's player UUID
//   - sender: Sender's username
//   - recipientID: Recipient's player UUID
//   - recipient: Recipient's username
//   - content: Message text
//
// Returns:
//   - Pointer to created ChatMessage (sender's copy)
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendDirectMessage(senderID uuid.UUID, sender string, recipientID uuid.UUID, recipient string, content string) *models.ChatMessage {
	msg := models.NewDirectMessage(senderID, sender, recipient, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to sender's history
	if history, exists := m.histories[senderID]; exists {
		history.AddMessage(msg)
	}

	// Create a copy for recipient with swapped sender/recipient
	recipientMsg := models.NewDirectMessage(senderID, sender, recipient, content)
	recipientMsg.Recipient = "" // Clear recipient so it shows as "from"

	// Add to recipient's history
	if history, exists := m.histories[recipientID]; exists {
		history.AddMessage(recipientMsg)
	}

	return msg
}

// SendTradeMessage sends a message to the trade channel.
//
// Trade channel is visible to all online players.
//
// Parameters:
//   - senderID: Sender's player UUID
//   - sender: Sender's username
//   - content: Message text
//
// Returns:
//   - Pointer to created ChatMessage
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendTradeMessage(senderID uuid.UUID, sender string, content string) *models.ChatMessage {
	msg := models.NewChatMessage(models.ChatChannelTrade, senderID, sender, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to all player histories
	for _, history := range m.histories {
		history.AddMessage(msg)
	}

	return msg
}

// SendCombatNotification sends a combat notification to specified players.
//
// Creates system message (no player sender) for combat logs.
//
// Parameters:
//   - playerIDs: List of player UUIDs to receive notification
//   - content: Combat notification text
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) SendCombatNotification(playerIDs []uuid.UUID, content string) {
	msg := models.NewSystemMessage(models.ChatChannelCombat, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to specified player combat logs
	for _, playerID := range playerIDs {
		if history, exists := m.histories[playerID]; exists {
			history.AddMessage(msg)
		}
	}
}

// BroadcastSystemMessage broadcasts a system message to all players.
//
// Creates system message (no player sender) for server-wide notifications.
//
// Parameters:
//   - channel: Chat channel to broadcast on
//   - content: System message text
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) BroadcastSystemMessage(channel models.ChatChannel, content string) {
	msg := models.NewSystemMessage(channel, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to all player histories
	for _, history := range m.histories {
		history.AddMessage(msg)
	}
}

// GetMessages retrieves messages for a specific player and channel.
//
// Parameters:
//   - playerID: Player UUID
//   - channel: Chat channel to retrieve
//   - limit: Maximum messages to return (0 for all)
//
// Returns:
//   - Slice of chat messages, or empty slice if player not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetMessages(playerID uuid.UUID, channel models.ChatChannel, limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []*models.ChatMessage{}
	}

	return history.GetMessages(channel, limit)
}

// GetDirectMessages retrieves direct messages between two players.
//
// Parameters:
//   - playerID: Requesting player UUID
//   - otherPlayer: Other player's username
//   - limit: Maximum messages to return (0 for all)
//
// Returns:
//   - Slice of direct messages, or empty slice if player not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetDirectMessages(playerID uuid.UUID, otherPlayer string, limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []*models.ChatMessage{}
	}

	return history.GetDirectMessages(otherPlayer, limit)
}

// GetRecentGlobal returns recent global messages for a newly connected player.
//
// Used to show chat history when players join the server.
//
// Parameters:
//   - limit: Maximum messages to return (0 for all cached messages)
//
// Returns:
//   - Slice of recent global messages (up to maxGlobalHistory)
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetRecentGlobal(limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > 0 && len(m.globalHistory) > limit {
		return m.globalHistory[len(m.globalHistory)-limit:]
	}

	return m.globalHistory
}

// ClearChannel clears all messages from a specific channel for a player.
//
// Parameters:
//   - playerID: Player UUID
//   - channel: Channel to clear
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) ClearChannel(playerID uuid.UUID, channel models.ChatChannel) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if history, exists := m.histories[playerID]; exists {
		history.ClearChannel(channel)
	}
}

// RemovePlayerHistory removes a player's chat history.
//
// Called when player disconnects to free memory. Player will receive
// recent global messages on reconnect via GetRecentGlobal.
//
// Parameters:
//   - playerID: Player UUID
//
// Thread Safety:
// Thread-safe. Acquires write lock.
func (m *Manager) RemovePlayerHistory(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.histories, playerID)
}

// GetActiveDirectChats returns usernames of players a user has direct chats with.
//
// Parameters:
//   - playerID: Player UUID
//
// Returns:
//   - Slice of usernames, or empty slice if player not found
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetActiveDirectChats(playerID uuid.UUID) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []string{}
	}

	return history.GetActiveDirectChats()
}

// GetStats returns statistics about chat activity.
//
// Returns:
//   - ChatStats with active histories and global message counts
//
// Thread Safety:
// Thread-safe. Acquires read lock.
func (m *Manager) GetStats() ChatStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ChatStats{
		ActiveHistories: len(m.histories),
		GlobalMessages:  len(m.globalHistory),
	}

	return stats
}

// ChatStats contains statistics about chat activity.
type ChatStats struct {
	ActiveHistories int `json:"active_histories"` // Number of online players with chat histories
	GlobalMessages  int `json:"global_messages"`  // Number of cached global messages
}
