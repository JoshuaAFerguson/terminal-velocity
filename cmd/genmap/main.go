// File: cmd/genmap/main.go
// Project: Terminal Velocity
// Description: Universe generation and database population tool
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/game/universe"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

func main() {
	var (
		numSystems    = flag.Int("systems", 100, "Number of star systems to generate")
		seed          = flag.Int64("seed", 0, "Random seed (0 for random)")
		showStats     = flag.Bool("stats", false, "Show detailed statistics")
		showSystems   = flag.Bool("systems-list", false, "List all systems")
		factionFilter = flag.String("faction", "", "Filter systems by faction")
		save          = flag.Bool("save", false, "Save universe to database")
		dbHost        = flag.String("db-host", "localhost", "Database host")
		dbPort        = flag.Int("db-port", 5432, "Database port")
		dbUser        = flag.String("db-user", "terminal_velocity", "Database user")
		dbPassword    = flag.String("db-password", "", "Database password")
		dbName        = flag.String("db-name", "terminal_velocity", "Database name")
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

	// Save to database if requested
	if *save {
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println("               SAVING TO DATABASE")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		if err := saveToDatabase(univ, *dbHost, *dbPort, *dbUser, *dbPassword, *dbName); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving universe to database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("✓ Universe saved to database successfully!")
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

func saveToDatabase(univ *universe.Universe, host string, port int, user, password, dbName string) error {
	ctx := context.Background()

	// Connect to database
	fmt.Printf("Connecting to database %s@%s:%d/%s...\n", user, host, port, dbName)
	cfg := &database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: dbName,
		SSLMode:  "disable",
	}
	db, err := database.NewDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create repository
	systemRepo := database.NewSystemRepository(db)

	fmt.Println("✓ Connected to database")
	fmt.Println()

	// Check if universe already exists
	existingSystems, err := systemRepo.ListSystems(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing systems: %w", err)
	}

	if len(existingSystems) > 0 {
		fmt.Printf("⚠️  WARNING: Database already contains %d systems.\n", len(existingSystems))
		fmt.Print("Continue and replace all systems? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
		fmt.Println()

		// Clear existing universe data
		fmt.Println("Clearing existing universe data...")
		if err := clearUniverse(ctx, db); err != nil {
			return fmt.Errorf("failed to clear universe: %w", err)
		}
		fmt.Println("✓ Cleared existing data")
		fmt.Println()
	}

	// Insert star systems
	fmt.Printf("Inserting %d star systems...\n", len(univ.Systems))
	systemCount := 0
	for _, system := range univ.Systems {
		if err := systemRepo.CreateSystem(ctx, system); err != nil {
			return fmt.Errorf("failed to insert system %s: %w", system.Name, err)
		}
		systemCount++
		if systemCount%20 == 0 {
			fmt.Printf("  %d/%d systems inserted...\n", systemCount, len(univ.Systems))
		}
	}
	fmt.Printf("✓ Inserted %d systems\n", systemCount)
	fmt.Println()

	// Insert planets
	fmt.Printf("Inserting %d planets...\n", len(univ.Planets))
	planetCount := 0
	for _, planet := range univ.Planets {
		if err := systemRepo.CreatePlanet(ctx, planet); err != nil {
			return fmt.Errorf("failed to insert planet %s: %w", planet.Name, err)
		}
		planetCount++
		if planetCount%50 == 0 {
			fmt.Printf("  %d/%d planets inserted...\n", planetCount, len(univ.Planets))
		}
	}
	fmt.Printf("✓ Inserted %d planets\n", planetCount)
	fmt.Println()

	// Insert system connections
	fmt.Println("Inserting system connections...")
	connectionCount := 0
	for _, system := range univ.Systems {
		for _, connectedID := range system.ConnectedSystems {
			// Only insert one direction to avoid duplicates
			if system.ID.String() < connectedID.String() {
				if err := systemRepo.CreateJumpRoute(ctx, system.ID, connectedID); err != nil {
					return fmt.Errorf("failed to insert connection: %w", err)
				}
				connectionCount++
			}
		}
	}
	fmt.Printf("✓ Inserted %d connections\n", connectionCount)

	return nil
}

func clearUniverse(ctx context.Context, db *database.DB) error {
	// Delete in correct order due to foreign keys
	queries := []string{
		"DELETE FROM system_connections",
		"DELETE FROM planets",
		"DELETE FROM star_systems",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
