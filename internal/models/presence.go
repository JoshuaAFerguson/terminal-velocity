// File: internal/models/presence.go
// Project: Terminal Velocity
// Description: Player presence tracking for multiplayer visibility and interaction
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlayerPresence represents a player's online status and current location
//
// This is used to track which players are online, where they are in the universe,
// and when they were last active. This enables multiplayer features like seeing
// other players in the same system, chat, and player-to-player interactions.

type PlayerPresence struct {
	PlayerID      uuid.UUID  `json:"player_id"`                // Player unique identifier
	Username      string     `json:"username"`                 // Player display name
	CurrentSystem uuid.UUID  `json:"current_system"`           // System player is currently in
	CurrentPlanet *uuid.UUID `json:"current_planet,omitempty"` // Planet if landed, nil if in space
	ShipName      string     `json:"ship_name"`                // Name of player's current ship
	ShipType      string     `json:"ship_type"`                // Type of ship (Fighter, Freighter, etc.)
	CombatRating  int        `json:"combat_rating"`            // Player's combat rating

	// Online status
	Online       bool          `json:"online"`        // Is the player currently online?
	LastSeen     time.Time     `json:"last_seen"`     // Last activity timestamp
	ConnectedAt  time.Time     `json:"connected_at"`  // When the player logged in
	IdleDuration time.Duration `json:"idle_duration"` // How long since last action

	// Activity status
	CurrentActivity string `json:"current_activity"` // What the player is doing
	InCombat        bool   `json:"in_combat"`        // Is player in combat?
	Docked          bool   `json:"docked"`           // Is player docked at a planet?

	// Faction and reputation
	FactionID  string `json:"faction_id,omitempty"` // Player faction (if any)
	IsCriminal bool   `json:"is_criminal"`          // Is player wanted/criminal?
}

// ActivityType represents different player activities
type ActivityType string

const (
	ActivityIdle       ActivityType = "idle"       // Not doing anything specific
	ActivityTrading    ActivityType = "trading"    // At a commodity market
	ActivityCombat     ActivityType = "combat"     // In active combat
	ActivityNavigation ActivityType = "navigating" // Planning routes
	ActivityShipyard   ActivityType = "shipyard"   // Browsing ships
	ActivityOutfitter  ActivityType = "outfitter"  // Browsing equipment
	ActivityMissions   ActivityType = "missions"   // Viewing mission board
	ActivityJumping    ActivityType = "jumping"    // In hyperspace
	ActivityDocked     ActivityType = "docked"     // Docked at planet
	ActivityInSpace    ActivityType = "in_space"   // Flying in system
)

// NewPlayerPresence creates a new presence record for a player
func NewPlayerPresence(player *Player, ship *Ship) *PlayerPresence {
	shipName := "Unknown"
	shipType := "Unknown"
	if ship != nil {
		shipName = ship.Name
		shipType = ship.TypeID
	}

	return &PlayerPresence{
		PlayerID:        player.ID,
		Username:        player.Username,
		CurrentSystem:   player.CurrentSystem,
		CurrentPlanet:   player.CurrentPlanet,
		ShipName:        shipName,
		ShipType:        shipType,
		CombatRating:    player.CombatRating,
		Online:          true,
		LastSeen:        time.Now(),
		ConnectedAt:     time.Now(),
		IdleDuration:    0,
		CurrentActivity: string(ActivityIdle),
		InCombat:        false,
		Docked:          player.CurrentPlanet != nil,
		IsCriminal:      player.IsCriminal,
	}
}

// UpdateActivity updates the player's current activity
func (p *PlayerPresence) UpdateActivity(activity ActivityType) {
	p.CurrentActivity = string(activity)
	p.LastSeen = time.Now()
	p.IdleDuration = 0

	// Update status flags based on activity
	p.InCombat = (activity == ActivityCombat)
	p.Docked = (activity == ActivityDocked)
}

// UpdateLocation updates the player's current system and planet
func (p *PlayerPresence) UpdateLocation(systemID uuid.UUID, planetID *uuid.UUID) {
	p.CurrentSystem = systemID
	p.CurrentPlanet = planetID
	p.LastSeen = time.Now()
	p.IdleDuration = 0

	if planetID != nil {
		p.Docked = true
		p.CurrentActivity = string(ActivityDocked)
	} else {
		p.Docked = false
		p.CurrentActivity = string(ActivityInSpace)
	}
}

// UpdateIdleTime calculates how long the player has been idle
func (p *PlayerPresence) UpdateIdleTime() {
	p.IdleDuration = time.Since(p.LastSeen)
}

// IsIdle returns true if the player has been idle for more than the given duration
func (p *PlayerPresence) IsIdle(threshold time.Duration) bool {
	return p.IdleDuration > threshold
}

// IsAfk returns true if the player is away from keyboard (idle > 5 minutes)
func (p *PlayerPresence) IsAfk() bool {
	return p.IsIdle(5 * time.Minute)
}

// GetStatusString returns a human-readable status string
func (p *PlayerPresence) GetStatusString() string {
	if p.InCombat {
		return "⚔️  In Combat"
	}

	if p.IsAfk() {
		return "💤 AFK"
	}

	if p.Docked {
		return "🛬 Docked"
	}

	switch ActivityType(p.CurrentActivity) {
	case ActivityTrading:
		return "💰 Trading"
	case ActivityNavigation:
		return "🗺️  Navigating"
	case ActivityShipyard:
		return "🚢 Shipyard"
	case ActivityOutfitter:
		return "⚙️  Outfitter"
	case ActivityMissions:
		return "📋 Missions"
	case ActivityJumping:
		return "⚡ Jumping"
	case ActivityInSpace:
		return "🚀 In Space"
	default:
		return "🌟 Online"
	}
}

// GetLastSeenString returns a human-readable last seen time
func (p *PlayerPresence) GetLastSeenString() string {
	duration := time.Since(p.LastSeen)

	if duration < time.Minute {
		return "Just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// GetOnlineDuration returns how long the player has been online
func (p *PlayerPresence) GetOnlineDuration() time.Duration {
	return time.Since(p.ConnectedAt)
}

// GetOnlineDurationString returns a human-readable online duration
func (p *PlayerPresence) GetOnlineDurationString() string {
	duration := p.GetOnlineDuration()

	if duration < time.Minute {
		return "< 1 min"
	}

	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours < 24 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	days := hours / 24
	hours = hours % 24

	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// CanInteract returns true if another player can interact with this player
//
// Players can interact if they're in the same system and not AFK
func (p *PlayerPresence) CanInteract() bool {
	return p.Online && !p.IsAfk()
}

// IsInSameSystem checks if this player is in the same system as another
func (p *PlayerPresence) IsInSameSystem(otherSystemID uuid.UUID) bool {
	return p.CurrentSystem == otherSystemID
}

// IsAtSamePlanet checks if this player is at the same planet as another
func (p *PlayerPresence) IsAtSamePlanet(otherPlanetID *uuid.UUID) bool {
	if p.CurrentPlanet == nil || otherPlanetID == nil {
		return false
	}
	return *p.CurrentPlanet == *otherPlanetID
}
