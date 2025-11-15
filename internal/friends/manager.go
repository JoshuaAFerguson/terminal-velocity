// File: internal/friends/manager.go
// Project: Terminal Velocity
// Description: Friends system manager
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package friends

import (
	"context"
	"fmt"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Friends")

// Manager handles friend-related operations
type Manager struct {
	mu            sync.RWMutex
	socialRepo    *database.SocialRepository
	presenceCache map[uuid.UUID]bool // Cache of online players
}

// NewManager creates a new friends manager
func NewManager(socialRepo *database.SocialRepository) *Manager {
	return &Manager{
		socialRepo:    socialRepo,
		presenceCache: make(map[uuid.UUID]bool),
	}
}

// ============================================================================
// Friend Request Operations
// ============================================================================

// SendFriendRequest sends a friend request to another player
func (m *Manager) SendFriendRequest(ctx context.Context, senderID uuid.UUID, receiverUsername string, getPlayerByUsername func(string) (*models.Player, error)) error {
	// Look up receiver by username
	receiver, err := getPlayerByUsername(receiverUsername)
	if err != nil {
		return fmt.Errorf("player not found: %s", receiverUsername)
	}

	// Check if trying to friend self
	if senderID == receiver.ID {
		return fmt.Errorf("cannot send friend request to yourself")
	}

	// Check if already friends
	areFriends, err := m.socialRepo.AreFriends(ctx, senderID, receiver.ID)
	if err != nil {
		return fmt.Errorf("failed to check friendship status: %w", err)
	}
	if areFriends {
		return fmt.Errorf("already friends with %s", receiverUsername)
	}

	// Check if blocked
	isBlocked, err := m.socialRepo.IsBlocked(ctx, receiver.ID, senderID)
	if err != nil {
		return fmt.Errorf("failed to check block status: %w", err)
	}
	if isBlocked {
		return fmt.Errorf("cannot send friend request to %s", receiverUsername)
	}

	// Create friend request
	_, err = m.socialRepo.CreateFriendRequest(ctx, senderID, receiver.ID)
	if err != nil {
		return fmt.Errorf("failed to create friend request: %w", err)
	}

	log.Info("Friend request sent: sender=%s, receiver=%s", senderID, receiver.ID)
	return nil
}

// AcceptFriendRequest accepts a friend request
func (m *Manager) AcceptFriendRequest(ctx context.Context, requestID uuid.UUID, playerID uuid.UUID) error {
	// Get request to verify receiver
	req, err := m.socialRepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("friend request not found: %w", err)
	}

	// Verify player is the receiver
	if req.ReceiverID != playerID {
		return fmt.Errorf("not authorized to accept this request")
	}

	// Accept request (creates friendship)
	err = m.socialRepo.AcceptFriendRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to accept friend request: %w", err)
	}

	log.Info("Friend request accepted: request=%s, friends=%s+%s",
		requestID, req.SenderID, req.ReceiverID)
	return nil
}

// DeclineFriendRequest declines a friend request
func (m *Manager) DeclineFriendRequest(ctx context.Context, requestID uuid.UUID, playerID uuid.UUID) error {
	// Get request to verify receiver
	req, err := m.socialRepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("friend request not found: %w", err)
	}

	// Verify player is the receiver
	if req.ReceiverID != playerID {
		return fmt.Errorf("not authorized to decline this request")
	}

	// Decline request
	err = m.socialRepo.DeclineFriendRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to decline friend request: %w", err)
	}

	log.Info("Friend request declined: request=%s", requestID)
	return nil
}

// GetPendingFriendRequests gets all pending friend requests for a player
func (m *Manager) GetPendingFriendRequests(ctx context.Context, playerID uuid.UUID) ([]models.FriendRequest, error) {
	requests, err := m.socialRepo.GetPendingFriendRequests(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend requests: %w", err)
	}
	return requests, nil
}

// ============================================================================
// Friend Management
// ============================================================================

// RemoveFriend removes a friend
func (m *Manager) RemoveFriend(ctx context.Context, playerID, friendID uuid.UUID) error {
	// Verify friendship exists
	areFriends, err := m.socialRepo.AreFriends(ctx, playerID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}
	if !areFriends {
		return fmt.Errorf("not friends")
	}

	// Remove friendship
	err = m.socialRepo.RemoveFriend(ctx, playerID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	log.Info("Friendship removed: %s and %s", playerID, friendID)
	return nil
}

// GetFriends gets all friends for a player with online status
func (m *Manager) GetFriends(ctx context.Context, playerID uuid.UUID) ([]models.Friend, error) {
	friends, err := m.socialRepo.GetFriends(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	// Populate online status from cache
	m.mu.RLock()
	for i := range friends {
		friends[i].IsOnline = m.presenceCache[friends[i].FriendID]
	}
	m.mu.RUnlock()

	return friends, nil
}

// GetOnlineFriends gets only online friends
func (m *Manager) GetOnlineFriends(ctx context.Context, playerID uuid.UUID) ([]models.Friend, error) {
	allFriends, err := m.GetFriends(ctx, playerID)
	if err != nil {
		return nil, err
	}

	var onlineFriends []models.Friend
	for _, friend := range allFriends {
		if friend.IsOnline {
			onlineFriends = append(onlineFriends, friend)
		}
	}

	return onlineFriends, nil
}

// AreFriends checks if two players are friends
func (m *Manager) AreFriends(ctx context.Context, player1ID, player2ID uuid.UUID) (bool, error) {
	areFriends, err := m.socialRepo.AreFriends(ctx, player1ID, player2ID)
	if err != nil {
		return false, fmt.Errorf("failed to check friendship: %w", err)
	}
	return areFriends, nil
}

// ============================================================================
// Block/Ignore Operations
// ============================================================================

// BlockPlayer blocks a player
func (m *Manager) BlockPlayer(ctx context.Context, blockerID uuid.UUID, blockedUsername string, reason string, getPlayerByUsername func(string) (*models.Player, error)) error {
	// Look up blocked player
	blocked, err := getPlayerByUsername(blockedUsername)
	if err != nil {
		return fmt.Errorf("player not found: %s", blockedUsername)
	}

	// Check if trying to block self
	if blockerID == blocked.ID {
		return fmt.Errorf("cannot block yourself")
	}

	// Block player (also removes friendship if exists)
	err = m.socialRepo.BlockPlayer(ctx, blockerID, blocked.ID, reason)
	if err != nil {
		return fmt.Errorf("failed to block player: %w", err)
	}

	log.Info("Player blocked: blocker=%s, blocked=%s, reason=%s",
		blockerID, blocked.ID, reason)
	return nil
}

// UnblockPlayer unblocks a player
func (m *Manager) UnblockPlayer(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	// Verify block exists
	isBlocked, err := m.socialRepo.IsBlocked(ctx, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to check block status: %w", err)
	}
	if !isBlocked {
		return fmt.Errorf("player is not blocked")
	}

	// Unblock player
	err = m.socialRepo.UnblockPlayer(ctx, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to unblock player: %w", err)
	}

	log.Info("Player unblocked: blocker=%s, blocked=%s", blockerID, blockedID)
	return nil
}

// GetBlockedPlayers gets all blocked players
func (m *Manager) GetBlockedPlayers(ctx context.Context, blockerID uuid.UUID) ([]models.Block, error) {
	blocks, err := m.socialRepo.GetBlockedPlayers(ctx, blockerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked players: %w", err)
	}
	return blocks, nil
}

// IsBlocked checks if a player is blocked
func (m *Manager) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	isBlocked, err := m.socialRepo.IsBlocked(ctx, blockerID, blockedID)
	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}
	return isBlocked, nil
}

// CanInteract checks if two players can interact (not blocked in either direction)
func (m *Manager) CanInteract(ctx context.Context, player1ID, player2ID uuid.UUID) (bool, error) {
	// Check if player1 blocked player2
	blocked1, err := m.IsBlocked(ctx, player1ID, player2ID)
	if err != nil {
		return false, err
	}
	if blocked1 {
		return false, nil
	}

	// Check if player2 blocked player1
	blocked2, err := m.IsBlocked(ctx, player2ID, player1ID)
	if err != nil {
		return false, err
	}
	if blocked2 {
		return false, nil
	}

	return true, nil
}

// ============================================================================
// Presence Management
// ============================================================================

// UpdateOnlineStatus updates the online status cache
func (m *Manager) UpdateOnlineStatus(playerID uuid.UUID, isOnline bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isOnline {
		m.presenceCache[playerID] = true
	} else {
		delete(m.presenceCache, playerID)
	}
}

// GetOnlineStatus checks if a player is online
func (m *Manager) GetOnlineStatus(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.presenceCache[playerID]
}

// GetOnlinePlayerCount returns the number of online players
func (m *Manager) GetOnlinePlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.presenceCache)
}

// ============================================================================
// Statistics
// ============================================================================

// GetFriendStats gets friendship statistics for a player
func (m *Manager) GetFriendStats(ctx context.Context, playerID uuid.UUID) (map[string]int, error) {
	friends, err := m.GetFriends(ctx, playerID)
	if err != nil {
		return nil, err
	}

	onlineCount := 0
	for _, friend := range friends {
		if friend.IsOnline {
			onlineCount++
		}
	}

	return map[string]int{
		"total":  len(friends),
		"online": onlineCount,
	}, nil
}
