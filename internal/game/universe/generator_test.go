// File: internal/game/universe/generator_test.go
// Project: Terminal Velocity
// Description: Procedural universe generation: generator_test
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package universe

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.NumSystems != 100 {
		t.Errorf("Expected 100 systems, got %d", config.NumSystems)
	}

	if config.CoreRadius != 30.0 {
		t.Errorf("Expected CoreRadius 30.0, got %f", config.CoreRadius)
	}

	if config.MinConnections < 1 || config.MaxConnections > 10 {
		t.Error("Invalid connection limits")
	}
}

func TestGeneratorCreation(t *testing.T) {
	config := DefaultConfig()
	gen := NewGenerator(config)

	if gen == nil {
		t.Fatal("Generator should not be nil")
	}

	if gen.nameGen == nil {
		t.Error("Name generator should be initialized")
	}

	if gen.rand == nil {
		t.Error("Random generator should be initialized")
	}
}

func TestUniverseGeneration(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 10 // Smaller for faster tests
	config.Seed = 12345    // Deterministic

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if universe == nil {
		t.Fatal("Universe should not be nil")
	}

	// Check system count
	if len(universe.Systems) != config.NumSystems {
		t.Errorf("Expected %d systems, got %d", config.NumSystems, len(universe.Systems))
	}

	// Check that Sol exists
	solFound := false
	for _, system := range universe.Systems {
		if system.Name == "Sol" {
			solFound = true
			if system.Position.X != 0 || system.Position.Y != 0 {
				t.Error("Sol should be at position (0, 0)")
			}
			break
		}
	}

	if !solFound {
		t.Error("Sol system not found")
	}

	// Check that all systems have planets
	for _, system := range universe.Systems {
		if len(system.Planets) == 0 {
			t.Errorf("System %s has no planets", system.Name)
		}
	}

	// Check that all planets exist in universe
	for _, system := range universe.Systems {
		for _, planet := range system.Planets {
			if _, ok := universe.Planets[planet.ID]; !ok {
				t.Errorf("Planet %s not found in universe map", planet.Name)
			}
		}
	}
}

func TestFactionAssignment(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 50
	config.Seed = 54321

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	factionCounts := make(map[string]int)

	for _, system := range universe.Systems {
		if system.GovernmentID == "" {
			t.Errorf("System %s has no government assigned", system.Name)
		}
		factionCounts[system.GovernmentID]++
	}

	// Check that Sol is UEF
	for _, system := range universe.Systems {
		if system.Name == "Sol" {
			if system.GovernmentID != "united_earth_federation" {
				t.Error("Sol should be controlled by UEF")
			}
		}
	}

	// Check that we have multiple factions
	if len(factionCounts) < 3 {
		t.Errorf("Expected at least 3 factions, got %d", len(factionCounts))
	}

	t.Logf("Faction distribution: %+v", factionCounts)
}

func TestTechLevelAssignment(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 30
	config.Seed = 99999

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for _, system := range universe.Systems {
		if system.TechLevel < 1 || system.TechLevel > 10 {
			t.Errorf("System %s has invalid tech level: %d", system.Name, system.TechLevel)
		}
	}

	// Check that Sol has high tech
	for _, system := range universe.Systems {
		if system.Name == "Sol" {
			if system.TechLevel < 7 {
				t.Errorf("Sol should have high tech level, got %d", system.TechLevel)
			}
		}
	}
}

func TestJumpRoutes(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 20
	config.Seed = 11111

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Check that all systems have connections
	for _, system := range universe.Systems {
		if len(system.ConnectedSystems) < config.MinConnections {
			t.Errorf("System %s has too few connections: %d (min: %d)",
				system.Name, len(system.ConnectedSystems), config.MinConnections)
		}

		if len(system.ConnectedSystems) > config.MaxConnections {
			t.Errorf("System %s has too many connections: %d (max: %d)",
				system.Name, len(system.ConnectedSystems), config.MaxConnections)
		}
	}

	// Check connectivity (BFS from Sol should reach all systems)
	solID := getSystemIDByName(universe, "Sol")
	if solID == nil {
		t.Fatal("Sol not found")
	}

	visited := make(map[string]bool)
	queue := []string{*solID}
	visited[*solID] = true

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		for _, system := range universe.Systems {
			if system.ID.String() == currentID {
				for _, connID := range system.ConnectedSystems {
					connIDStr := connID.String()
					if !visited[connIDStr] {
						visited[connIDStr] = true
						queue = append(queue, connIDStr)
					}
				}
				break
			}
		}
	}

	if len(visited) != len(universe.Systems) {
		t.Errorf("Not all systems are reachable from Sol. Visited: %d/%d",
			len(visited), len(universe.Systems))
	}
}

func TestPlanetGeneration(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 15
	config.Seed = 22222

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	totalPlanets := 0

	for _, system := range universe.Systems {
		numPlanets := len(system.Planets)

		if numPlanets < 1 || numPlanets > 4 {
			t.Errorf("System %s has invalid number of planets: %d", system.Name, numPlanets)
		}

		totalPlanets += numPlanets

		// Check each planet
		for _, planet := range system.Planets {
			if planet.Name == "" {
				t.Error("Planet has no name")
			}

			if planet.TechLevel != system.TechLevel {
				t.Errorf("Planet %s tech level (%d) doesn't match system (%d)",
					planet.Name, planet.TechLevel, system.TechLevel)
			}

			if len(planet.Services) == 0 {
				t.Errorf("Planet %s has no services", planet.Name)
			}

			// All planets should have trading
			hasTrading := false
			for _, service := range planet.Services {
				if service == "trading" {
					hasTrading = true
					break
				}
			}
			if !hasTrading {
				t.Errorf("Planet %s missing trading service", planet.Name)
			}
		}
	}

	t.Logf("Generated %d planets across %d systems (avg: %.1f per system)",
		totalPlanets, len(universe.Systems), float64(totalPlanets)/float64(len(universe.Systems)))
}

func TestNameUniqueness(t *testing.T) {
	config := DefaultConfig()
	config.NumSystems = 100
	config.Seed = 33333

	gen := NewGenerator(config)
	universe, err := gen.Generate()

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	systemNames := make(map[string]bool)

	for _, system := range universe.Systems {
		if systemNames[system.Name] {
			t.Errorf("Duplicate system name: %s", system.Name)
		}
		systemNames[system.Name] = true
	}
}

// Helper function
func getSystemIDByName(universe *Universe, name string) *string {
	for _, system := range universe.Systems {
		if system.Name == name {
			id := system.ID.String()
			return &id
		}
	}
	return nil
}

func BenchmarkUniverseGeneration(b *testing.B) {
	config := DefaultConfig()

	for i := 0; i < b.N; i++ {
		gen := NewGenerator(config)
		_, err := gen.Generate()
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

func BenchmarkSmallUniverse(b *testing.B) {
	config := DefaultConfig()
	config.NumSystems = 20

	for i := 0; i < b.N; i++ {
		gen := NewGenerator(config)
		_, err := gen.Generate()
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}
