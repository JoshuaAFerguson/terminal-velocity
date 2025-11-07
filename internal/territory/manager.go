// File: internal/territory/manager.go
// Project: Terminal Velocity
// Version: 1.0.0

package territory

import (
	"errors"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var (
	ErrAlreadyClaimed    = errors.New("system already claimed")
	ErrNotClaimed        = errors.New("system not claimed")
	ErrNotOwner          = errors.New("not territory owner")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Manager struct {
	mu          sync.RWMutex
	territories map[uuid.UUID]*models.Territory
	byFaction   map[uuid.UUID][]*models.Territory
}

func NewManager() *Manager {
	return &Manager{
		territories: make(map[uuid.UUID]*models.Territory),
		byFaction:   make(map[uuid.UUID][]*models.Territory),
	}
}

func (m *Manager) ClaimSystem(systemID uuid.UUID, systemName string, factionID uuid.UUID, factionTag string) (*models.Territory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.territories[systemID]; exists {
		return nil, ErrAlreadyClaimed
	}

	territory := models.NewTerritory(systemID, systemName, factionID, factionTag)
	m.territories[systemID] = territory
	m.byFaction[factionID] = append(m.byFaction[factionID], territory)

	return territory, nil
}

func (m *Manager) GetTerritory(systemID uuid.UUID) (*models.Territory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territory, exists := m.territories[systemID]
	if !exists {
		return nil, ErrNotClaimed
	}
	return territory, nil
}

func (m *Manager) GetFactionTerritories(factionID uuid.UUID) []*models.Territory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.byFaction[factionID]
}

func (m *Manager) GetAllTerritories() []*models.Territory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territories := make([]*models.Territory, 0, len(m.territories))
	for _, t := range m.territories {
		territories = append(territories, t)
	}
	return territories
}

func (m *Manager) IsSystemClaimed(systemID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.territories[systemID]
	return exists
}
