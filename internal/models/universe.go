// File: internal/models/universe.go
// Project: Terminal Velocity
// Description: Universe, star system, and planet models
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import "github.com/google/uuid"

// StarSystem represents a star system in the galaxy with its planets and connections.
//
// Star systems are the primary locations in Terminal Velocity's universe. The game
// typically contains 100+ systems generated procedurally using:
//   - Spiral galaxy distribution for realistic positioning
//   - Minimum Spanning Tree (Prim's algorithm) for jump route connections
//   - Radial tech level distribution (high tech in core, low at edges)
//   - Random assignment of 6 NPC government factions
//   - Procedural planet generation with randomized services
//
// Systems can be controlled by either NPC governments (GovernmentID) or player
// factions (ControlledBy). Player factions can claim systems for territory control
// and passive income generation.
//
// See internal/game/universe/ for the procedural generation code.
type StarSystem struct {
	// ID is the unique identifier for this star system
	ID uuid.UUID `json:"id"`

	// Name is the system name (e.g., "Sol", "Alpha Centauri", "Betelgeuse")
	// Generated procedurally or from predefined list
	Name string `json:"name"`

	// Position is the 2D coordinates in the galaxy map
	// Used for visualization and jump distance calculations
	Position Position `json:"position"`

	// GovernmentID identifies the NPC faction controlling this system
	// Valid values: united_earth_federation, auroran_empire, free_worlds_alliance,
	//               rigel_outer_marches, pacifist_union, crimson_syndicate
	// Affects market prices, police presence, and mission availability
	GovernmentID string `json:"government_id"`

	// ControlledBy is the player faction UUID that has claimed this system
	// nil if system is still under NPC government control
	// Player factions earn passive income from controlled systems
	// Omitted from JSON if nil
	ControlledBy *uuid.UUID `json:"controlled_by,omitempty"`

	// TechLevel represents technological advancement of the system
	// Range: 1 (primitive) to 10 (ultra high-tech)
	// Affects:
	//   - Available commodities for trading (higher tech = more advanced goods)
	//   - Commodity prices (tech level differences create profit opportunities)
	//   - Ship and equipment availability at shipyards
	TechLevel int `json:"tech_level"`

	// Description provides flavor text about the system
	// Generated procedurally based on tech level and government
	Description string `json:"description"`

	// Planets is the list of planets and stations in this system
	// Typically 1-5 planets per system
	// Players can land on planets to access services (trading, shipyard, missions, etc.)
	Planets []Planet `json:"planets"`

	// ConnectedSystems is the list of system UUIDs reachable via hyperspace jump
	// Jump routes are bidirectional (if A connects to B, B connects to A)
	// Generated using MST algorithm to ensure all systems are reachable
	// Additional connections added for gameplay variety
	ConnectedSystems []uuid.UUID `json:"connected_systems"`
}

// Position represents 2D coordinates in the galaxy map.
//
// Positions are used for:
//   - Star system placement in the galaxy visualization
//   - Jump distance calculations between systems
//   - Procedural generation using spiral galaxy distribution
type Position struct {
	// X is the horizontal coordinate in the galaxy
	// Typically ranges from -1000 to +1000 in a 100-system universe
	X int `json:"x"`

	// Y is the vertical coordinate in the galaxy
	// Typically ranges from -1000 to +1000 in a 100-system universe
	Y int `json:"y"`
}

// Planet represents a planet or station in a star system.
//
// Planets are where players access services like trading markets, shipyards,
// outfitters, and mission boards. Each planet offers a subset of available
// services based on its population and tech level.
//
// Available Services:
//   - trading: Commodity market for buying/selling goods
//   - shipyard: Ship purchase and sales
//   - outfitter: Equipment purchase and installation
//   - missions: Mission board for accepting jobs
//   - bar: Information, rumors, and special encounters
//
// Planets inherit their parent system's tech level but can have variations.
// Higher tech levels provide access to more advanced commodities and equipment.
type Planet struct {
	// ID is the unique identifier for this planet
	ID uuid.UUID `json:"id"`

	// SystemID is the UUID of the parent star system
	// Used to locate which system this planet belongs to
	SystemID uuid.UUID `json:"system_id"`

	// Name is the planet name (e.g., "Earth", "Mars", "Betelgeuse VII")
	// Generated procedurally with system name + roman numeral/suffix
	Name string `json:"name"`

	// Description provides flavor text about the planet
	// Generated based on population, tech level, and government
	Description string `json:"description"`

	// X is the X coordinate within the parent system
	// Used for within-system visualization (future feature)
	X float64 `json:"x"`

	// Y is the Y coordinate within the parent system
	// Used for within-system visualization (future feature)
	Y float64 `json:"y"`

	// Services is the list of available services on this planet
	// Valid values: "trading", "shipyard", "outfitter", "missions", "bar"
	// Not all planets have all services - depends on population/tech
	Services []string `json:"services"`

	// Population is the planet's population count
	// Range: 1,000 to 10,000,000,000
	// Higher population planets tend to have more services
	Population int64 `json:"population"`

	// TechLevel represents the planet's tech advancement
	// Range: 1-10
	// Usually matches or is close to the parent system's tech level
	// Affects available commodities and equipment
	TechLevel int `json:"tech_level"`
}

// Government represents an NPC faction that controls star systems.
//
// There are 6 NPC governments in Terminal Velocity, each with distinct
// characteristics, alignments, and diplomatic relationships:
//
//   - United Earth Federation: Lawful, democratic, allied with Free Worlds
//   - Auroran Empire: Lawful, militaristic, allied with Rigel
//   - Free Worlds Alliance: Neutral, independent, allied with Federation
//   - Rigel Outer Marches: Neutral, frontier, allied with Empire
//   - Pacifist Union: Lawful, peaceful, hostile to none
//   - Crimson Syndicate: Chaotic, pirates, hostile to all lawful governments
//
// Government affects:
//   - Commodity legality (contraband varies by government)
//   - Police patrol frequency and behavior
//   - Mission types and availability
//   - Player reputation consequences
//   - Market prices and availability
type Government struct {
	// ID is the unique identifier for this government
	// Valid values: united_earth_federation, auroran_empire, free_worlds_alliance,
	//               rigel_outer_marches, pacifist_union, crimson_syndicate
	ID string `json:"id"`

	// Name is the display name of the government
	// e.g., "United Earth Federation", "Crimson Syndicate"
	Name string `json:"name"`

	// Description provides background about the government
	// Includes history, political structure, and characteristics
	Description string `json:"description"`

	// Alignment indicates the government's moral/legal stance
	// Valid values: "lawful", "neutral", "chaotic"
	// Affects police behavior, contraband enforcement, and diplomacy
	Alignment string `json:"alignment"`

	// Color is the hex color code for UI display
	// Used in galaxy maps, system info, and faction UI elements
	Color string `json:"color"`

	// HostileTo lists government IDs that this government is hostile toward
	// Players with high reputation in hostile governments may be attacked
	// Affects encounter generation and police behavior
	HostileTo []string `json:"hostile_to"`

	// AlliedWith lists government IDs that this government is allied with
	// Players gain indirect reputation with allies when helping this government
	// Affects trade prices and mission rewards
	AlliedWith []string `json:"allied_with"`
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

// DistanceFrom calculates distance from given coordinates to this planet
func (p *Planet) DistanceFrom(x, y float64) float64 {
	dx := p.X - x
	dy := p.Y - y
	return dx*dx + dy*dy // Return squared distance (good enough for comparison)
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
