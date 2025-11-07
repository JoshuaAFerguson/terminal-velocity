package universe

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/google/uuid"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
)

// Generator handles universe creation
type Generator struct {
	config  GeneratorConfig
	rand    *rand.Rand
	nameGen *NameGenerator
}

// GeneratorConfig configures universe generation
type GeneratorConfig struct {
	NumSystems     int
	CoreRadius     float64 // Distance from Sol for core systems
	MidRadius      float64 // Distance from Sol for mid systems
	OuterRadius    float64 // Distance from Sol for outer systems
	EdgeRadius     float64 // Distance from Sol for edge systems
	MinConnections int     // Min jump routes per system
	MaxConnections int     // Max jump routes per system
	Seed           int64   // Random seed
}

// DefaultConfig returns default generator configuration
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		NumSystems:     100,
		CoreRadius:     30.0,
		MidRadius:      60.0,
		OuterRadius:    100.0,
		EdgeRadius:     120.0,
		MinConnections: 2,
		MaxConnections: 5,
		Seed:           0, // 0 = random seed
	}
}

// NewGenerator creates a new universe generator
func NewGenerator(config GeneratorConfig) *Generator {
	seed := config.Seed
	if seed == 0 {
		seed = rand.Int63()
	}

	r := rand.New(rand.NewSource(seed))

	return &Generator{
		config:  config,
		rand:    r,
		nameGen: NewNameGenerator(r),
	}
}

// Generate creates a complete universe
func (g *Generator) Generate() (*Universe, error) {
	universe := &Universe{
		Systems: make(map[uuid.UUID]*models.StarSystem),
		Planets: make(map[uuid.UUID]*models.Planet),
	}

	// Phase 1: Create star systems
	systems := g.generateSystems()

	// Phase 2: Assign faction control
	g.assignFactionControl(systems)

	// Phase 3: Assign tech levels
	g.assignTechLevels(systems)

	// Phase 3.5: Generate descriptions (after faction and tech assignment)
	g.generateDescriptions(systems)

	// Phase 4: Generate jump routes (using MST + extra connections)
	g.UpdatedGenerateJumpRoutes(systems)

	// Phase 5: Generate planets
	planets := g.generatePlanets(systems)

	// Convert to maps
	for i := range systems {
		universe.Systems[systems[i].ID] = &systems[i]
	}
	for i := range planets {
		universe.Planets[planets[i].ID] = &planets[i]
	}

	return universe, nil
}

// generateSystems creates all star systems
func (g *Generator) generateSystems() []models.StarSystem {
	systems := make([]models.StarSystem, g.config.NumSystems)

	// System 0 is always Sol (Earth)
	systems[0] = models.StarSystem{
		ID:               uuid.New(),
		Name:             "Sol",
		Position:         models.Position{X: 0, Y: 0},
		Description:      "Birthplace of humanity. Home to Earth, Mars, and the United Earth Federation capital.",
		ConnectedSystems: []uuid.UUID{},
	}

	// Generate other systems in a disk pattern
	for i := 1; i < g.config.NumSystems; i++ {
		systems[i] = g.generateSystem(i)
	}

	return systems
}

// generateSystem creates a single star system
func (g *Generator) generateSystem(index int) models.StarSystem {
	// Generate position in a disk/spiral pattern
	angle := g.rand.Float64() * 2 * math.Pi

	// Distance distribution: more systems in mid/outer regions
	distance := g.generateDistance()

	x := int(distance * math.Cos(angle))
	y := int(distance * math.Sin(angle))

	name := g.nameGen.GenerateSystemName()

	return models.StarSystem{
		ID:               uuid.New(),
		Name:             name,
		Position:         models.Position{X: x, Y: y},
		Description:      "", // Will be set after faction assignment
		ConnectedSystems: []uuid.UUID{},
	}
}

// generateDistance generates a distance from Sol with appropriate distribution
func (g *Generator) generateDistance() float64 {
	// Weighted random: more systems in outer/mid regions
	roll := g.rand.Float64()

	switch {
	case roll < 0.15: // 15% in core
		return g.rand.Float64() * g.config.CoreRadius
	case roll < 0.50: // 35% in mid
		return g.config.CoreRadius + g.rand.Float64()*(g.config.MidRadius-g.config.CoreRadius)
	case roll < 0.90: // 40% in outer
		return g.config.MidRadius + g.rand.Float64()*(g.config.OuterRadius-g.config.MidRadius)
	default: // 10% at edge
		return g.config.OuterRadius + g.rand.Float64()*(g.config.EdgeRadius-g.config.OuterRadius)
	}
}

// assignFactionControl assigns NPC factions to systems based on distance
func (g *Generator) assignFactionControl(systems []models.StarSystem) {
	for i := range systems {
		distance := g.getDistanceFromSol(systems[i].Position)

		// Special case: Sol is always UEF
		if systems[i].Name == "Sol" {
			systems[i].GovernmentID = "united_earth_federation"
			continue
		}

		// Assign based on distance
		switch {
		case distance < g.config.CoreRadius:
			// Core systems: UEF or ROM
			if g.rand.Float64() < 0.7 {
				systems[i].GovernmentID = "united_earth_federation"
			} else {
				systems[i].GovernmentID = "republic_of_mars"
			}

		case distance < g.config.MidRadius:
			// Mid systems: FTG hubs or independent
			roll := g.rand.Float64()
			if roll < 0.25 {
				systems[i].GovernmentID = "free_traders_guild"
			} else if roll < 0.5 {
				systems[i].GovernmentID = "united_earth_federation" // UEF influence
			} else {
				systems[i].GovernmentID = "independent"
			}

		case distance < g.config.OuterRadius:
			// Outer systems: Frontier Worlds or independent
			if g.rand.Float64() < 0.4 {
				systems[i].GovernmentID = "frontier_worlds"
			} else {
				systems[i].GovernmentID = "independent"
			}

		default:
			// Edge systems: Auroran Empire
			systems[i].GovernmentID = "auroran_empire"
		}
	}
}

// assignTechLevels assigns technology levels based on faction and distance
func (g *Generator) assignTechLevels(systems []models.StarSystem) {
	for i := range systems {
		distance := g.getDistanceFromSol(systems[i].Position)

		var baseTech int
		switch {
		case distance < g.config.CoreRadius:
			baseTech = 8
		case distance < g.config.MidRadius:
			baseTech = 6
		case distance < g.config.OuterRadius:
			baseTech = 4
		default: // Edge (Auroran)
			baseTech = 10
		}

		// Add variance ±1
		variance := g.rand.Intn(3) - 1 // -1, 0, or 1
		techLevel := baseTech + variance

		// Clamp to 1-10
		if techLevel < 1 {
			techLevel = 1
		} else if techLevel > 10 {
			techLevel = 10
		}

		systems[i].TechLevel = techLevel
	}
}

// generateJumpRoutes creates hyperspace routes between systems
func (g *Generator) generateJumpRoutes(systems []models.StarSystem) {
	// Use a simplified minimum spanning tree approach
	// Then add additional routes for interesting topology

	// TODO: Implement proper MST algorithm
	// For now, connect nearby systems

	for i := range systems {
		// Find 2-5 nearest neighbors
		numConnections := g.config.MinConnections + g.rand.Intn(g.config.MaxConnections-g.config.MinConnections+1)

		nearest := g.findNearestSystems(&systems[i], systems, numConnections)

		for _, neighborID := range nearest {
			systems[i].AddConnection(neighborID)
		}
	}
}

// findNearestSystems finds the N nearest systems
func (g *Generator) findNearestSystems(system *models.StarSystem, allSystems []models.StarSystem, n int) []uuid.UUID {
	type distPair struct {
		id       uuid.UUID
		distance float64
	}

	distances := make([]distPair, 0, len(allSystems))

	for i := range allSystems {
		if allSystems[i].ID == system.ID {
			continue
		}

		dist := system.Position.DistanceTo(allSystems[i].Position)
		distances = append(distances, distPair{allSystems[i].ID, dist})
	}

	// Sort by distance (simple bubble sort for small N)
	for i := 0; i < len(distances)-1; i++ {
		for j := 0; j < len(distances)-i-1; j++ {
			if distances[j].distance > distances[j+1].distance {
				distances[j], distances[j+1] = distances[j+1], distances[j]
			}
		}
	}

	// Return N nearest
	result := make([]uuid.UUID, 0, n)
	for i := 0; i < n && i < len(distances); i++ {
		result = append(result, distances[i].id)
	}

	return result
}

// generatePlanets creates planets for each system
func (g *Generator) generatePlanets(systems []models.StarSystem) []models.Planet {
	planets := make([]models.Planet, 0)

	for i := range systems {
		numPlanets := 1 + g.rand.Intn(3) // 1-4 planets per system

		for j := 0; j < numPlanets; j++ {
			planet := g.generatePlanet(&systems[i], j)
			planets = append(planets, planet)
			systems[i].Planets = append(systems[i].Planets, planet)
		}
	}

	return planets
}

// generatePlanet creates a single planet
func (g *Generator) generatePlanet(system *models.StarSystem, index int) models.Planet {
	planetName := GeneratePlanetName(system.Name, index, g.rand)
	isStation := planetName != fmt.Sprintf("%s %c", system.Name, rune('A'+index))

	services := g.generateServices(system.TechLevel)

	return models.Planet{
		ID:          uuid.New(),
		SystemID:    system.ID,
		Name:        planetName,
		Description: GeneratePlanetDescription(g.rand, isStation),
		Services:    services,
		Population:  g.generatePopulation(system.TechLevel),
		TechLevel:   system.TechLevel,
	}
}

// generateServices determines what services a planet offers
func (g *Generator) generateServices(techLevel int) []string {
	services := []string{"trading"} // All planets have basic trading

	// Tech-based services
	if techLevel >= 3 {
		services = append(services, "bar")
	}
	if techLevel >= 4 {
		services = append(services, "missions")
	}
	if techLevel >= 5 {
		services = append(services, "outfitter")
	}
	if techLevel >= 6 {
		services = append(services, "shipyard")
	}

	return services
}

// generatePopulation generates a random population based on tech level
func (g *Generator) generatePopulation(techLevel int) int64 {
	base := int64(techLevel * techLevel * 1000000) // Higher tech = higher pop
	variance := g.rand.Int63n(base / 2)
	return base + variance
}

// getDistanceFromSol calculates distance from Sol
func (g *Generator) getDistanceFromSol(pos models.Position) float64 {
	return math.Sqrt(float64(pos.X*pos.X + pos.Y*pos.Y))
}

// generateDescriptions sets system descriptions based on faction and location
func (g *Generator) generateDescriptions(systems []models.StarSystem) {
	for i := range systems {
		distance := g.getDistanceFromSol(systems[i].Position)
		systems[i].Description = GenerateDescription(systems[i].GovernmentID, distance)
	}
}

// Universe holds the generated universe
type Universe struct {
	Systems map[uuid.UUID]*models.StarSystem
	Planets map[uuid.UUID]*models.Planet
}
