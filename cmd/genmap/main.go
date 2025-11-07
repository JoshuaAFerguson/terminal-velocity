package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/s0v3r1gn/terminal-velocity/internal/game/universe"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
)

func main() {
	var (
		numSystems    = flag.Int("systems", 100, "Number of star systems to generate")
		seed          = flag.Int64("seed", 0, "Random seed (0 for random)")
		showStats     = flag.Bool("stats", false, "Show detailed statistics")
		showSystems   = flag.Bool("systems-list", false, "List all systems")
		factionFilter = flag.String("faction", "", "Filter systems by faction")
	)
	flag.Parse()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("       TERMINAL VELOCITY - UNIVERSE GENERATOR")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Create generator
	config := universe.DefaultConfig()
	config.NumSystems = *numSystems
	config.Seed = *seed

	fmt.Printf("Generating universe with %d systems", *numSystems)
	if *seed != 0 {
		fmt.Printf(" (seed: %d)", *seed)
	}
	fmt.Println()
	fmt.Println()

	gen := universe.NewGenerator(config)
	univ, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating universe: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Universe generated successfully!")
	fmt.Println()

	// Show statistics
	showUniverseStats(univ)

	if *showStats {
		fmt.Println()
		showDetailedStats(univ)
	}

	if *showSystems {
		fmt.Println()
		showSystemsList(univ, *factionFilter)
	}
}

func showUniverseStats(univ *universe.Universe) {
	factionCounts := make(map[string]int)
	techLevelCounts := make(map[int]int)
	totalPlanets := 0
	totalConnections := 0

	for _, system := range univ.Systems {
		factionCounts[system.GovernmentID]++
		techLevelCounts[system.TechLevel]++
		totalPlanets += len(system.Planets)
		totalConnections += len(system.ConnectedSystems)
	}

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                   UNIVERSE STATISTICS                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("  Systems:        %d\n", len(univ.Systems))
	fmt.Printf("  Planets:        %d (avg: %.1f per system)\n",
		totalPlanets, float64(totalPlanets)/float64(len(univ.Systems)))
	fmt.Printf("  Jump Routes:    %d (avg: %.1f per system)\n",
		totalConnections/2, float64(totalConnections)/float64(len(univ.Systems)))
	fmt.Println()

	fmt.Println("  FACTION DISTRIBUTION:")
	fmt.Println("  ────────────────────────────────────────────────────────")

	factions := []string{
		"united_earth_federation",
		"republic_of_mars",
		"free_traders_guild",
		"frontier_worlds",
		"crimson_collective",
		"auroran_empire",
		"independent",
	}

	factionSymbols := map[string]string{
		"united_earth_federation": "⊕",
		"republic_of_mars":        "♂",
		"free_traders_guild":      "¤",
		"frontier_worlds":         "⚑",
		"crimson_collective":      "☠",
		"auroran_empire":          "⧈",
		"independent":             "·",
	}

	factionNames := map[string]string{
		"united_earth_federation": "United Earth Federation",
		"republic_of_mars":        "Republic of Mars",
		"free_traders_guild":      "Free Traders Guild",
		"frontier_worlds":         "Frontier Worlds Alliance",
		"crimson_collective":      "Crimson Collective",
		"auroran_empire":          "Auroran Empire",
		"independent":             "Independent",
	}

	for _, fid := range factions {
		if count, ok := factionCounts[fid]; ok && count > 0 {
			symbol := factionSymbols[fid]
			name := factionNames[fid]
			percentage := float64(count) / float64(len(univ.Systems)) * 100
			fmt.Printf("  %s %-25s  %3d systems (%5.1f%%)\n",
				symbol, name, count, percentage)
		}
	}

	fmt.Println()
	fmt.Println("  TECHNOLOGY LEVELS:")
	fmt.Println("  ────────────────────────────────────────────────────────")

	for tech := 1; tech <= 10; tech++ {
		if count, ok := techLevelCounts[tech]; ok && count > 0 {
			bar := ""
			barLength := count * 40 / len(univ.Systems)
			for i := 0; i < barLength; i++ {
				bar += "█"
			}
			fmt.Printf("  Level %2d: %-40s %3d systems\n", tech, bar, count)
		}
	}
}

func showDetailedStats(univ *universe.Universe) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  DETAILED STATISTICS                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Service availability
	serviceCounts := make(map[string]int)
	for _, planet := range univ.Planets {
		for _, service := range planet.Services {
			serviceCounts[service]++
		}
	}

	fmt.Println("  PLANETARY SERVICES:")
	fmt.Println("  ────────────────────────────────────────────────────────")

	services := []string{"trading", "bar", "missions", "outfitter", "shipyard"}
	for _, svc := range services {
		if count, ok := serviceCounts[svc]; ok {
			fmt.Printf("  %-12s  %3d planets\n", svc, count)
		}
	}

	// Find interesting systems
	fmt.Println()
	fmt.Println("  NOTABLE SYSTEMS:")
	fmt.Println("  ────────────────────────────────────────────────────────")

	// Most connected system
	var mostConnected *models.StarSystem
	maxConnections := 0
	for _, system := range univ.Systems {
		if len(system.ConnectedSystems) > maxConnections {
			maxConnections = len(system.ConnectedSystems)
			mostConnected = system
		}
	}

	if mostConnected != nil {
		fmt.Printf("  Most Connected:  %s (%d routes)\n",
			mostConnected.Name, len(mostConnected.ConnectedSystems))
	}

	// Highest tech
	var highestTech *models.StarSystem
	maxTech := 0
	for _, system := range univ.Systems {
		if system.TechLevel > maxTech {
			maxTech = system.TechLevel
			highestTech = system
		}
	}

	if highestTech != nil {
		fmt.Printf("  Highest Tech:    %s (level %d)\n",
			highestTech.Name, highestTech.TechLevel)
	}

	// Most planets
	var mostPlanets *models.StarSystem
	maxPlanets := 0
	for _, system := range univ.Systems {
		if len(system.Planets) > maxPlanets {
			maxPlanets = len(system.Planets)
			mostPlanets = system
		}
	}

	if mostPlanets != nil {
		fmt.Printf("  Most Planets:    %s (%d planets)\n",
			mostPlanets.Name, len(mostPlanets.Planets))
	}
}

func showSystemsList(univ *universe.Universe, factionFilter string) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SYSTEMS LIST                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Convert map to slice for sorting
	systems := make([]*models.StarSystem, 0, len(univ.Systems))
	for _, system := range univ.Systems {
		if factionFilter == "" || system.GovernmentID == factionFilter {
			systems = append(systems, system)
		}
	}

	// Sort by name
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].Name < systems[j].Name
	})

	factionSymbols := map[string]string{
		"united_earth_federation": "⊕",
		"republic_of_mars":        "♂",
		"free_traders_guild":      "¤",
		"frontier_worlds":         "⚑",
		"crimson_collective":      "☠",
		"auroran_empire":          "⧈",
		"independent":             "·",
	}

	fmt.Printf("%-20s  %s  Tech  Planets  Routes  Position\n", "System", "Gov")
	fmt.Println("─────────────────────────────────────────────────────────────────────────")

	for _, system := range systems {
		symbol := factionSymbols[system.GovernmentID]
		fmt.Printf("%-20s   %s    %2d      %d       %d    (%4d, %4d)\n",
			system.Name,
			symbol,
			system.TechLevel,
			len(system.Planets),
			len(system.ConnectedSystems),
			system.Position.X,
			system.Position.Y,
		)
	}

	fmt.Println()
	fmt.Printf("Total: %d systems", len(systems))
	if factionFilter != "" {
		fmt.Printf(" (filtered by: %s)", factionFilter)
	}
	fmt.Println()
}
