// File: internal/chat/manager.go
// Project: Terminal Velocity
// Description: Chat manager for multiplayer communication and message routing
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package chat

import (
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles chat message routing and history for all players
type Manager struct {
	mu       sync.RWMutex
	histories map[uuid.UUID]*models.ChatHistory // Per-player chat histories

	// Message retention
	globalHistory []*models.ChatMessage // Recent global messages
	maxGlobalHistory int
}

// NewManager creates a new chat manager
func NewManager() *Manager {
	return &Manager{
		histories:        make(map[uuid.UUID]*models.ChatHistory),
		globalHistory:    []*models.ChatMessage{},
		maxGlobalHistory: 200, // Keep last 200 global messages
	}
}

// GetOrCreateHistory gets or creates a chat history for a player
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

// SendGlobalMessage sends a message to all online players
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

// SendSystemMessage sends a message to players in a specific system
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

// SendFactionMessage sends a message to all faction members
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

// SendDirectMessage sends a private message to a specific player
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

// SendTradeMessage sends a message to the trade channel
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

// SendCombatNotification sends a combat notification
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

// BroadcastSystemMessage broadcasts a system message to all players
func (m *Manager) BroadcastSystemMessage(channel models.ChatChannel, content string) {
	msg := models.NewSystemMessage(channel, content)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to all player histories
	for _, history := range m.histories {
		history.AddMessage(msg)
	}
}

// GetMessages retrieves messages for a specific player and channel
func (m *Manager) GetMessages(playerID uuid.UUID, channel models.ChatChannel, limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []*models.ChatMessage{}
	}

	return history.GetMessages(channel, limit)
}

// GetDirectMessages retrieves direct messages between two players
func (m *Manager) GetDirectMessages(playerID uuid.UUID, otherPlayer string, limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []*models.ChatMessage{}
	}

	return history.GetDirectMessages(otherPlayer, limit)
}

// GetRecentGlobal returns recent global messages for a newly connected player
func (m *Manager) GetRecentGlobal(limit int) []*models.ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > 0 && len(m.globalHistory) > limit {
		return m.globalHistory[len(m.globalHistory)-limit:]
	}

	return m.globalHistory
}

// ClearChannel clears all messages from a specific channel for a player
func (m *Manager) ClearChannel(playerID uuid.UUID, channel models.ChatChannel) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if history, exists := m.histories[playerID]; exists {
		history.ClearChannel(channel)
	}
}

// RemovePlayerHistory removes a player's chat history (on disconnect)
func (m *Manager) RemovePlayerHistory(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.histories, playerID)
}

// GetActiveDirectChats returns usernames of players a user has direct chats with
func (m *Manager) GetActiveDirectChats(playerID uuid.UUID) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.histories[playerID]
	if !exists {
		return []string{}
	}

	return history.GetActiveDirectChats()
}

// GetStats returns statistics about chat activity
func (m *Manager) GetStats() ChatStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ChatStats{
		ActiveHistories: len(m.histories),
		GlobalMessages:  len(m.globalHistory),
	}

	return stats
}

// ChatStats contains statistics about chat activity
type ChatStats struct {
	ActiveHistories int `json:"active_histories"`
	GlobalMessages  int `json:"global_messages"`
}
