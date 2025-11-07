// File: internal/factions/manager.go
// Project: Terminal Velocity
// Description: Faction management system for player organizations
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package factions

import (
	"errors"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Factions")

var (
	ErrFactionNotFound   = errors.New("faction not found")
	ErrNotMember         = errors.New("player is not a member")
	ErrInsufficientRank  = errors.New("insufficient rank for this action")
	ErrFactionFull       = errors.New("faction has reached member limit")
	ErrAlreadyMember     = errors.New("player is already a member")
	ErrInsufficientFunds = errors.New("insufficient faction treasury funds")
	ErrNameTaken         = errors.New("faction name already taken")
	ErrTagTaken          = errors.New("faction tag already taken")
)

// Manager handles faction operations and state
type Manager struct {
	mu       sync.RWMutex
	factions map[uuid.UUID]*models.PlayerFaction // All factions by ID
	names    map[string]uuid.UUID                // Name -> ID mapping
	tags     map[string]uuid.UUID                // Tag -> ID mapping
	members  map[uuid.UUID]uuid.UUID             // Player ID -> Faction ID
}

// NewManager creates a new faction manager
func NewManager() *Manager {
	return &Manager{
		factions: make(map[uuid.UUID]*models.PlayerFaction),
		names:    make(map[string]uuid.UUID),
		tags:     make(map[string]uuid.UUID),
		members:  make(map[uuid.UUID]uuid.UUID),
	}
}

// CreateFaction creates a new faction
func (m *Manager) CreateFaction(name, tag string, founderID uuid.UUID, alignment string) (*models.PlayerFaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if name is taken
	if _, exists := m.names[name]; exists {
		return nil, ErrNameTaken
	}

	// Check if tag is taken
	if _, exists := m.tags[tag]; exists {
		return nil, ErrTagTaken
	}

	// Check if founder is already in a faction
	if _, exists := m.members[founderID]; exists {
		return nil, ErrAlreadyMember
	}

	faction := models.NewPlayerFaction(name, tag, founderID, alignment)

	m.factions[faction.ID] = faction
	m.names[name] = faction.ID
	m.tags[tag] = faction.ID
	m.members[founderID] = faction.ID

	return faction, nil
}

// GetFaction retrieves a faction by ID
func (m *Manager) GetFaction(factionID uuid.UUID) (*models.PlayerFaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return nil, ErrFactionNotFound
	}

	return faction, nil
}

// GetFactionByName retrieves a faction by name
func (m *Manager) GetFactionByName(name string) (*models.PlayerFaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factionID, exists := m.names[name]
	if !exists {
		return nil, ErrFactionNotFound
	}

	return m.factions[factionID], nil
}

// GetFactionByTag retrieves a faction by tag
func (m *Manager) GetFactionByTag(tag string) (*models.PlayerFaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factionID, exists := m.tags[tag]
	if !exists {
		return nil, ErrFactionNotFound
	}

	return m.factions[factionID], nil
}

// GetPlayerFaction retrieves the faction a player belongs to
func (m *Manager) GetPlayerFaction(playerID uuid.UUID) (*models.PlayerFaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factionID, exists := m.members[playerID]
	if !exists {
		return nil, ErrNotMember
	}

	return m.factions[factionID], nil
}

// JoinFaction adds a player to a faction
func (m *Manager) JoinFaction(factionID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already in a faction
	if _, exists := m.members[playerID]; exists {
		return ErrAlreadyMember
	}

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	if !faction.AddMember(playerID) {
		return ErrFactionFull
	}

	m.members[playerID] = factionID
	return nil
}

// LeaveFaction removes a player from their faction
func (m *Manager) LeaveFaction(playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	factionID, exists := m.members[playerID]
	if !exists {
		return ErrNotMember
	}

	faction := m.factions[factionID]

	// Can't leave if you're the leader
	if faction.IsLeader(playerID) {
		return errors.New("leader must transfer leadership before leaving")
	}

	faction.RemoveMember(playerID)
	delete(m.members, playerID)

	return nil
}

// KickMember removes a member from the faction
func (m *Manager) KickMember(factionID, kickerID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Must be officer or leader
	if !faction.IsOfficer(kickerID) {
		return ErrInsufficientRank
	}

	// Can't kick leader
	if faction.IsLeader(targetID) {
		return errors.New("cannot kick faction leader")
	}

	// Can't kick higher or equal rank unless you're leader
	if faction.IsOfficer(targetID) && !faction.IsLeader(kickerID) {
		return ErrInsufficientRank
	}

	faction.RemoveMember(targetID)
	delete(m.members, targetID)

	return nil
}

// PromoteMember promotes a member to officer
func (m *Manager) PromoteMember(factionID, promoterID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Only leader can promote
	if !faction.IsLeader(promoterID) {
		return ErrInsufficientRank
	}

	if !faction.PromoteToOfficer(targetID) {
		return errors.New("cannot promote player")
	}

	return nil
}

// DemoteMember demotes an officer to member
func (m *Manager) DemoteMember(factionID, demoterID, targetID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Only leader can demote
	if !faction.IsLeader(demoterID) {
		return ErrInsufficientRank
	}

	if !faction.DemoteFromOfficer(targetID) {
		return errors.New("cannot demote player")
	}

	return nil
}

// Deposit adds credits to faction treasury
func (m *Manager) Deposit(factionID, playerID uuid.UUID, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Must be a member
	if !faction.IsMember(playerID) {
		return ErrNotMember
	}

	faction.Deposit(amount)
	return nil
}

// Withdraw removes credits from faction treasury
func (m *Manager) Withdraw(factionID, playerID uuid.UUID, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Must be officer or leader
	if !faction.IsOfficer(playerID) {
		return ErrInsufficientRank
	}

	if !faction.Withdraw(amount) {
		return ErrInsufficientFunds
	}

	return nil
}

// GetAllFactions returns all factions
func (m *Manager) GetAllFactions() []*models.PlayerFaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factions := make([]*models.PlayerFaction, 0, len(m.factions))
	for _, faction := range m.factions {
		factions = append(factions, faction)
	}

	return factions
}

// GetRecruitingFactions returns factions that are recruiting
func (m *Manager) GetRecruitingFactions() []*models.PlayerFaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factions := []*models.PlayerFaction{}
	for _, faction := range m.factions {
		if faction.IsRecruiting && faction.CanRecruit() {
			factions = append(factions, faction)
		}
	}

	return factions
}

// UpdateSettings updates faction settings
func (m *Manager) UpdateSettings(factionID, playerID uuid.UUID, settings models.FactionSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	faction, exists := m.factions[factionID]
	if !exists {
		return ErrFactionNotFound
	}

	// Only leader can update settings
	if !faction.IsLeader(playerID) {
		return ErrInsufficientRank
	}

	faction.Settings = settings
	return nil
}

// GetStats returns faction statistics
func (m *Manager) GetStats() FactionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := FactionStats{
		TotalFactions:      len(m.factions),
		TotalMembers:       len(m.members),
		RecruitingFactions: 0,
	}

	for _, faction := range m.factions {
		if faction.IsRecruiting {
			stats.RecruitingFactions++
		}
	}

	return stats
}

// FactionStats contains statistics about faction system
type FactionStats struct {
	TotalFactions      int `json:"total_factions"`
	TotalMembers       int `json:"total_members"`
	RecruitingFactions int `json:"recruiting_factions"`
}
