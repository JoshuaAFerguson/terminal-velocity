// File: internal/models/social.go
// Project: Terminal Velocity
// Description: Social feature models (friends, blocks, mail, notifications)
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Friends System
// ============================================================================

// Friend represents a friendship between two players
type Friend struct {
	ID        uuid.UUID `json:"id"`
	PlayerID  uuid.UUID `json:"player_id"`
	FriendID  uuid.UUID `json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`

	// Populated fields (not in database)
	FriendName   string `json:"friend_name,omitempty"`
	IsOnline     bool   `json:"is_online,omitempty"`
	CurrentShip  string `json:"current_ship,omitempty"`
	Location     string `json:"location,omitempty"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
}

// FriendRequest represents a pending friend request
type FriendRequest struct {
	ID         uuid.UUID `json:"id"`
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Status     string    `json:"status"` // pending, accepted, declined
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Populated fields
	SenderName   string `json:"sender_name,omitempty"`
	ReceiverName string `json:"receiver_name,omitempty"`
}

// FriendRequest status constants
const (
	FriendRequestPending  = "pending"
	FriendRequestAccepted = "accepted"
	FriendRequestDeclined = "declined"
)

// ============================================================================
// Block/Ignore System
// ============================================================================

// Block represents a player blocking another player
type Block struct {
	ID        uuid.UUID `json:"id"`
	BlockerID uuid.UUID `json:"blocker_id"`
	BlockedID uuid.UUID `json:"blocked_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`

	// Populated fields
	BlockedName string `json:"blocked_name,omitempty"`
}

// ============================================================================
// Mail System
// ============================================================================

// Mail represents a persistent message between players
type Mail struct {
	ID         uuid.UUID  `json:"id"`
	SenderID   *uuid.UUID `json:"sender_id"`   // NULL if sender deleted
	SenderName string     `json:"sender_name"` // Preserved even if sender deleted
	ReceiverID uuid.UUID  `json:"receiver_id"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`

	// Attachments
	AttachedCredits int64           `json:"attached_credits"`
	AttachedItems   []uuid.UUID     `json:"attached_items"`

	// Status
	IsRead    bool `json:"is_read"`
	IsDeleted bool `json:"is_deleted"`

	// Timestamps
	SentAt time.Time  `json:"sent_at"`
	ReadAt *time.Time `json:"read_at"`
}

// MailAttachment represents an item attached to mail
type MailAttachment struct {
	ItemType string    `json:"item_type"` // credits, weapon, outfit, commodity
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
	Name     string    `json:"name"`
}

// ============================================================================
// Notification System
// ============================================================================

// Notification represents a player notification
type Notification struct {
	ID       uuid.UUID `json:"id"`
	PlayerID uuid.UUID `json:"player_id"`

	// Content
	Type    string `json:"type"` // friend_request, mail, trade_offer, pvp_challenge, etc.
	Title   string `json:"title"`
	Message string `json:"message"`

	// Related entities (for clickable notifications)
	RelatedPlayerID   *uuid.UUID `json:"related_player_id"`
	RelatedEntityType string     `json:"related_entity_type"` // mail, trade, pvp, territory, faction
	RelatedEntityID   *uuid.UUID `json:"related_entity_id"`

	// Status
	IsRead      bool `json:"is_read"`
	IsDismissed bool `json:"is_dismissed"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`
	ExpiresAt time.Time  `json:"expires_at"`

	// Action data (flexible JSON)
	ActionData map[string]interface{} `json:"action_data"`
}

// Notification type constants
const (
	NotificationTypeFriendRequest   = "friend_request"
	NotificationTypeMail            = "mail"
	NotificationTypeTradeOffer      = "trade_offer"
	NotificationTypePvPChallenge    = "pvp_challenge"
	NotificationTypeTerritoryAttack = "territory_attack"
	NotificationTypeFactionInvite   = "faction_invite"
	NotificationTypeSystemMessage   = "system_message"
	NotificationTypeAchievement     = "achievement"
	NotificationTypeEvent           = "event"
)

// ============================================================================
// Player Profile Extensions
// ============================================================================

// PlayerProfile represents extended player profile information
type PlayerProfile struct {
	// Basic info (from Player model)
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`

	// Profile fields
	Bio            string    `json:"bio"`
	JoinDate       time.Time `json:"join_date"`
	TotalPlaytime  int       `json:"total_playtime"` // in seconds
	ProfilePrivacy string    `json:"profile_privacy"` // public, friends, private

	// Stats (from Player model)
	Credits       int64  `json:"credits"`
	CombatRating  int    `json:"combat_rating"`
	NetWorth      int64  `json:"net_worth"`
	CurrentShip   string `json:"current_ship"`
	CurrentSystem string `json:"current_system"`
	CurrentPlanet string `json:"current_planet"`

	// Faction info
	FactionID   *uuid.UUID `json:"faction_id"`
	FactionName string     `json:"faction_name,omitempty"`
	FactionRank string     `json:"faction_rank,omitempty"`

	// Achievements (count)
	AchievementCount int `json:"achievement_count"`

	// Social
	FriendCount int  `json:"friend_count"`
	IsOnline    bool `json:"is_online"`
	IsFriend    bool `json:"is_friend"`    // From perspective of viewer
	IsBlocked   bool `json:"is_blocked"`   // Viewer blocked this player
	IsBlocking  bool `json:"is_blocking"`  // This player blocked viewer

	// Activity
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	LastActive *time.Time `json:"last_active,omitempty"`
}

// Profile privacy levels
const (
	ProfilePrivacyPublic  = "public"
	ProfilePrivacyFriends = "friends"
	ProfilePrivacyPrivate = "private"
)

// CanView checks if viewer can see this profile based on privacy settings
func (p *PlayerProfile) CanView(viewerID uuid.UUID, areFriends bool) bool {
	// Owner can always view own profile
	if viewerID == p.ID {
		return true
	}

	switch p.ProfilePrivacy {
	case ProfilePrivacyPublic:
		return true
	case ProfilePrivacyFriends:
		return areFriends
	case ProfilePrivacyPrivate:
		return false
	default:
		return true // Default to public
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// FormatPlaytime formats total playtime into human-readable string
func FormatPlaytime(seconds int) string {
	if seconds < 60 {
		return "< 1 minute"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours == 0 {
		return formatMinutes(minutes)
	}

	if hours < 24 {
		return formatHours(hours, minutes)
	}

	days := hours / 24
	remainingHours := hours % 24
	return formatDays(days, remainingHours)
}

func formatMinutes(m int) string {
	if m == 1 {
		return "1 minute"
	}
	return formatInt(m) + " minutes"
}

func formatHours(h, m int) string {
	if h == 1 {
		if m == 0 {
			return "1 hour"
		}
		return "1 hour " + formatMinutes(m)
	}
	if m == 0 {
		return formatInt(h) + " hours"
	}
	return formatInt(h) + " hours " + formatMinutes(m)
}

func formatDays(d, h int) string {
	if d == 1 {
		if h == 0 {
			return "1 day"
		}
		return "1 day " + formatInt(h) + " hours"
	}
	if h == 0 {
		return formatInt(d) + " days"
	}
	return formatInt(d) + " days " + formatInt(h) + " hours"
}

func formatInt(n int) string {
	// Simple integer to string conversion
	return string(rune('0' + n))
}

// GetMailPreview returns a truncated preview of mail body
func (m *Mail) GetMailPreview(maxLength int) string {
	if len(m.Body) <= maxLength {
		return m.Body
	}
	return m.Body[:maxLength] + "..."
}

// HasAttachments checks if mail has any attachments
func (m *Mail) HasAttachments() bool {
	return m.AttachedCredits > 0 || len(m.AttachedItems) > 0
}

// IsExpired checks if notification has expired
func (n *Notification) IsExpired() bool {
	return time.Now().After(n.ExpiresAt)
}

// CanBeActioned checks if notification can be actioned (not read, not dismissed, not expired)
func (n *Notification) CanBeActioned() bool {
	return !n.IsRead && !n.IsDismissed && !n.IsExpired()
}
