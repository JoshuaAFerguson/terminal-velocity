// File: internal/encounters/manager.go
// Project: Terminal Velocity
// Version: 1.0.0

package encounters

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager handles random encounter generation and tracking
type Manager struct {
	mu              sync.RWMutex
	templates       []EncounterTemplate
	activeEncounters map[uuid.UUID]*Encounter
	history         map[uuid.UUID]*EncounterHistory
	rand            *rand.Rand
}

// NewManager creates a new encounter manager
func NewManager() *Manager {
	return &Manager{
		templates:       GetAllTemplates(),
		activeEncounters: make(map[uuid.UUID]*Encounter),
		history:         make(map[uuid.UUID]*EncounterHistory),
		rand:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateEncounter attempts to generate a random encounter
func (m *Manager) GenerateEncounter(
	playerID uuid.UUID,
	systemID uuid.UUID,
	systemName string,
	techLevel int,
	govType string,
) *Encounter {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Base 20% chance of any encounter
	if m.rand.Float64() > 0.20 {
		return nil
	}

	// Select a template based on rarity weights
	template := m.selectTemplate(techLevel, govType)
	if template == nil {
		return nil
	}

	// Create encounter from template
	encounter := NewEncounter(*template, systemID, systemName, playerID)

	// Randomize details
	encounter.NPCLevel = m.randomizeLevel(template.MinLevel, template.MaxLevel)
	encounter.Credits = m.randomizeCredits(template.MinCredits, template.MaxCredits)
	encounter.Reputation = m.randomizeReputation(template.ReputationRange)

	if len(template.ShipTypes) > 0 {
		encounter.NPCShipType = template.ShipTypes[m.rand.Intn(len(template.ShipTypes))]
		encounter.NPCName = m.generateNPCName(encounter.Type)
	}

	// Add random cargo
	if len(template.PossibleCargo) > 0 {
		numCargo := m.rand.Intn(3) + 1 // 1-3 types
		for i := 0; i < numCargo; i++ {
			cargo := template.PossibleCargo[m.rand.Intn(len(template.PossibleCargo))]
			quantity := m.rand.Intn(20) + 5 // 5-25 units
			encounter.Cargo[cargo] = quantity
		}
	}

	// Store active encounter
	m.activeEncounters[encounter.ID] = encounter

	// Ensure player has history
	m.ensureHistory(playerID)

	return encounter
}

// selectTemplate selects a random template based on rarity weights
func (m *Manager) selectTemplate(techLevel int, govType string) *EncounterTemplate {
	// Filter templates by tech level and government
	var validTemplates []EncounterTemplate
	for _, template := range m.templates {
		if template.MinTechLevel <= techLevel && techLevel <= template.MaxTechLevel {
			if template.RequiredGovType == "" || template.RequiredGovType == govType {
				validTemplates = append(validTemplates, template)
			}
		}
	}

	if len(validTemplates) == 0 {
		return nil
	}

	// Weighted random selection
	rarityChance := m.rand.Float64()

	var targetRarity EncounterRarity
	switch {
	case rarityChance < 0.01: // 1%
		targetRarity = RarityLegendary
	case rarityChance < 0.06: // 5%
		targetRarity = RarityVeryRare
	case rarityChance < 0.21: // 15%
		targetRarity = RarityRare
	case rarityChance < 0.51: // 30%
		targetRarity = RarityUncommon
	default: // 49%
		targetRarity = RarityCommon
	}

	// Find templates matching target rarity
	var matchingTemplates []EncounterTemplate
	for _, template := range validTemplates {
		if template.Rarity == targetRarity {
			matchingTemplates = append(matchingTemplates, template)
		}
	}

	// Fallback to any valid template if none match
	if len(matchingTemplates) == 0 {
		matchingTemplates = validTemplates
	}

	// Select random template
	selected := matchingTemplates[m.rand.Intn(len(matchingTemplates))]
	return &selected
}

// randomizeLevel generates a random NPC level
func (m *Manager) randomizeLevel(min, max int) int {
	if max <= min {
		return min
	}
	return min + m.rand.Intn(max-min+1)
}

// randomizeCredits generates random credit reward
func (m *Manager) randomizeCredits(min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + int64(m.rand.Int63n(max-min+1))
}

// randomizeReputation generates random reputation change
func (m *Manager) randomizeReputation(repRange [2]int) int {
	min, max := repRange[0], repRange[1]
	if max <= min {
		return min
	}
	return min + m.rand.Intn(max-min+1)
}

// generateNPCName creates a random NPC name
func (m *Manager) generateNPCName(encounterType EncounterType) string {
	prefixes := []string{"Captain", "Commander", "Admiral", "Pilot"}
	names := []string{
		"Blackstar", "Nova", "Crimson", "Shadow", "Storm",
		"Drake", "Vega", "Orion", "Phoenix", "Raven",
		"Ghost", "Hunter", "Reaper", "Viper", "Falcon",
	}

	switch encounterType {
	case EncounterPirate:
		prefix := []string{"Dread", "Black", "Red", "Iron"}[m.rand.Intn(4)]
		name := names[m.rand.Intn(len(names))]
		return prefix + " " + name
	case EncounterTrader:
		return "Merchant " + names[m.rand.Intn(len(names))]
	case EncounterPatrol:
		return "Officer " + names[m.rand.Intn(len(names))]
	default:
		prefix := prefixes[m.rand.Intn(len(prefixes))]
		name := names[m.rand.Intn(len(names))]
		return prefix + " " + name
	}
}

// ResolveEncounter resolves an encounter with an outcome
func (m *Manager) ResolveEncounter(encounterID uuid.UUID, outcome EncounterOutcome) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	encounter, exists := m.activeEncounters[encounterID]
	if !exists {
		return ErrEncounterNotFound
	}

	encounter.Resolve(outcome)

	// Record in history
	m.ensureHistory(encounter.PlayerID)
	m.history[encounter.PlayerID].RecordEncounter(encounter)

	return nil
}

// GetEncounter retrieves an encounter by ID
func (m *Manager) GetEncounter(encounterID uuid.UUID) (*Encounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	encounter, exists := m.activeEncounters[encounterID]
	if !exists {
		return nil, ErrEncounterNotFound
	}

	return encounter, nil
}

// GetActiveEncounters returns all unresolved encounters for a player
func (m *Manager) GetActiveEncounters(playerID uuid.UUID) []*Encounter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*Encounter
	for _, encounter := range m.activeEncounters {
		if encounter.PlayerID == playerID && !encounter.IsResolved() {
			active = append(active, encounter)
		}
	}

	return active
}

// GetHistory returns a player's encounter history
func (m *Manager) GetHistory(playerID uuid.UUID) *EncounterHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if history, exists := m.history[playerID]; exists {
		return history
	}

	return NewEncounterHistory(playerID)
}

// ensureHistory creates history if it doesn't exist
func (m *Manager) ensureHistory(playerID uuid.UUID) {
	if _, exists := m.history[playerID]; !exists {
		m.history[playerID] = NewEncounterHistory(playerID)
	}
}

// CleanupResolvedEncounters removes old resolved encounters
func (m *Manager) CleanupResolvedEncounters() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	cutoff := time.Now().Add(-24 * time.Hour)

	for id, encounter := range m.activeEncounters {
		if encounter.IsResolved() && encounter.ResolvedAt.Before(cutoff) {
			delete(m.activeEncounters, id)
			cleaned++
		}
	}

	return cleaned
}

// GetStats returns overall encounter statistics
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"active_encounters": len(m.activeEncounters),
		"total_players":     len(m.history),
		"templates":         len(m.templates),
	}

	// Count by rarity in active encounters
	rarityCount := make(map[EncounterRarity]int)
	for _, encounter := range m.activeEncounters {
		rarityCount[encounter.Rarity]++
	}

	stats["common"] = rarityCount[RarityCommon]
	stats["uncommon"] = rarityCount[RarityUncommon]
	stats["rare"] = rarityCount[RarityRare]
	stats["very_rare"] = rarityCount[RarityVeryRare]
	stats["legendary"] = rarityCount[RarityLegendary]

	return stats
}

// Custom error
var ErrEncounterNotFound = &encounterError{"encounter not found"}

type encounterError struct {
	msg string
}

func (e *encounterError) Error() string {
	return e.msg
}
