// File: internal/models/universe.go
// Project: Terminal Velocity
// Description: Data models for universe
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import "github.com/google/uuid"

// StarSystem represents a star system in the galaxy
type StarSystem struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Position         Position    `json:"position"`
	GovernmentID     string      `json:"government_id"`           // NPC faction controlling system
	ControlledBy     *uuid.UUID  `json:"controlled_by,omitempty"` // Player faction ID if controlled
	TechLevel        int         `json:"tech_level"`              // 1-10
	Description      string      `json:"description"`
	Planets          []Planet    `json:"planets"`
	ConnectedSystems []uuid.UUID `json:"connected_systems"` // Systems reachable from here
}

// Position represents 2D coordinates in galaxy map
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Planet represents a planet or station in a system
type Planet struct {
	ID          uuid.UUID `json:"id"`
	SystemID    uuid.UUID `json:"system_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Services    []string  `json:"services"` // shipyard, outfitter, missions, trading, bar
	Population  int64     `json:"population"`
	TechLevel   int       `json:"tech_level"`
}

// Government represents an NPC faction
type Government struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alignment   string `json:"alignment"` // lawful, neutral, chaotic
	Color       string `json:"color"`     // for UI display

	// AI behavior
	HostileTo  []string `json:"hostile_to"`  // Government IDs
	AlliedWith []string `json:"allied_with"` // Government IDs
}

// Distance calculates distance between two positions
func (p Position) DistanceTo(other Position) float64 {
	dx := float64(p.X - other.X)
	dy := float64(p.Y - other.Y)
	return (dx*dx + dy*dy) // Return squared distance (good enough for comparison)
}

// HasService checks if planet offers a service
func (p *Planet) HasService(service string) bool {
	for _, s := range p.Services {
		if s == service {
			return true
		}
	}
	return false
}

// IsConnectedTo checks if system has a direct jump route to another
func (s *StarSystem) IsConnectedTo(systemID uuid.UUID) bool {
	for _, id := range s.ConnectedSystems {
		if id == systemID {
			return true
		}
	}
	return false
}

// AddConnection adds a bidirectional jump route between systems
func (s *StarSystem) AddConnection(systemID uuid.UUID) {
	if !s.IsConnectedTo(systemID) {
		s.ConnectedSystems = append(s.ConnectedSystems, systemID)
	}
}

// IsControlledByPlayer checks if system is controlled by a player faction
func (s *StarSystem) IsControlledByPlayer() bool {
	return s.ControlledBy != nil
}
