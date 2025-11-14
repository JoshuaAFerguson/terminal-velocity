// File: internal/models/loadout.go
// Project: Terminal Velocity
// Description: Data models for ship loadout sharing
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package models

import (
	"time"

	"github.com/google/uuid"
)

// SharedLoadout represents a shared ship configuration
type SharedLoadout struct {
	ID          uuid.UUID      `json:"id"`
	PlayerID    uuid.UUID      `json:"player_id"`
	ShipTypeID  string         `json:"ship_type_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Weapons     []string       `json:"weapons"`
	Outfits     []string       `json:"outfits"`
	Stats       *LoadoutStats  `json:"stats"`
	IsPublic    bool           `json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Views       int            `json:"views"`
	Favorites   int            `json:"favorites"`
}

// LoadoutStats contains calculated statistics for a loadout
type LoadoutStats struct {
	TotalDPS     int `json:"total_dps"`
	TotalArmor   int `json:"total_armor"`
	TotalShield  int `json:"total_shield"`
	TotalSpeed   int `json:"total_speed"`
	TotalCargo   int `json:"total_cargo"`
	EnergyUsage  int `json:"energy_usage"`
	MassUsage    int `json:"mass_usage"`
}

// LoadoutComparison represents a comparison between two loadouts
type LoadoutComparison struct {
	Loadout1    *SharedLoadout `json:"loadout1"`
	Loadout2    *SharedLoadout `json:"loadout2"`
	Differences []string       `json:"differences"`
}

// LoadoutFilter represents search filters for loadouts
type LoadoutFilter struct {
	ShipTypeID string
	PlayerID   uuid.UUID
	IsPublic   bool
	SortBy     string // views, favorites, created_at
	Limit      int
	Offset     int
}
