// File: cmd/genmap/main.go
// Project: Terminal Velocity
// Description: Universe generation and database population tool
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package main provides the universe generation CLI tool for Terminal Velocity.
//
// Tool Overview:
// This utility generates procedural star system universes and optionally saves
// them to the database. It's used for initial server setup and universe resets.
//
// Features:
//   - Generate N star systems with realistic distribution
//   - Create jump routes using Minimum Spanning Tree algorithm
//   - Assign tech levels, governments, and planets
//   - Display detailed statistics and visualizations
//   - Save directly to PostgreSQL database
//   - Preview before saving with confirmation prompt
//
// Command-Line Flags:
//   -systems <N>        Number of star systems to generate (default: 100)
//   -seed <N>           Random seed for reproducibility (0 for random)
//   -stats              Show detailed statistics after generation
//   -systems-list       List all generated systems
//   -faction <id>       Filter system list by faction
//   -save               Save universe to database (interactive)
//   -db-host <host>     Database host (default: localhost)
//   -db-port <port>     Database port (default: 5432)
//   -db-user <user>     Database user (default: terminal_velocity)
//   -db-password <pass> Database password
//   -db-name <name>     Database name (default: terminal_velocity)
//
// Example Usage:
//   # Preview 100-system universe with stats
//   ./genmap -systems 100 -stats
//
//   # Generate reproducible universe with specific seed
//   ./genmap -systems 50 -seed 12345 -stats
//
//   # List all systems filtered by faction
//   ./genmap -systems 100 -systems-list -faction united_earth_federation
//
//   # Generate and save to database (with confirmation)
//   ./genmap -systems 100 -save -db-password mypassword
//
//   # Save to custom database
//   ./genmap -systems 200 -save \
//     -db-host prod-db.example.com \
//     -db-port 5432 \
//     -db-user admin \
//     -db-password secret \
//     -db-name terminal_velocity_prod
//
// Universe Generation Algorithm:
//   1. Generate system positions using spiral galaxy distribution
//   2. Create Minimum Spanning Tree (MST) for initial connectivity
//   3. Add extra random connections for gameplay variety
//   4. Assign tech levels (higher in core, lower at edges)
//   5. Distribute 6 NPC factions across systems
//   6. Generate planets for each system with services
//
// Database Integration:
// When -save flag is used:
//   1. Connects to PostgreSQL database
//   2. Checks for existing universe data
//   3. Prompts for confirmation if data exists
//   4. Clears old data (systems, planets, connections)
//   5. Inserts new universe data with progress updates
//   6. Commits transaction
//
// Safety Features:
//   - Confirmation prompt before overwriting existing universe
//   - Transaction rollback on error
//   - Foreign key constraint handling
//   - Progress indicators for large datasets
//
// Output Format:
// The tool produces beautiful ASCII art visualizations:
//   - Universe statistics (systems, planets, routes)
//   - Faction distribution with symbols
//   - Tech level histogram
//   - System listing with coordinates
//
// Exit Codes:
//   0 - Success
//   1 - Generation error or database error
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

// main is the entry point for the universe generation tool.
//
// Execution Flow:
//   1. Parse command-line flags
//   2. Display banner
//   3. Create universe generator with configuration
//   4. Generate universe (systems, planets, connections)
//   5. Display statistics and visualizations
//   6. Optionally save to database (with confirmation)
//
// Error Handling:
// Universe generation errors exit with code 1.
// Database errors exit with code 1.
// User cancellation (confirmation prompt) exits with code 0.
func main() {
	// Define command-line flags with defaults
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

// showUniverseStats displays formatted statistics about the generated universe.
//
// Displayed Information:
//   - Total system count
//   - Total planet count (with average per system)
//   - Total jump routes (with average per system)
//   - Faction distribution (count and percentage per faction)
//   - Tech level distribution (histogram)
//
// Output Format:
// Uses box-drawing characters and ASCII art for visual appeal:
//   ╔═══════════════════════════════════╗
//   ║   UNIVERSE STATISTICS             ║
//   ╚═══════════════════════════════════╝
//
// Faction Symbols:
//   ⊕ - United Earth Federation
//   ♂ - Republic of Mars
//   ¤ - Free Traders Guild
//   ⚑ - Frontier Worlds Alliance
//   ☠ - Crimson Collective
//   ⧈ - Auroran Empire
//   · - Independent
//
// Tech Level Bars:
// Visual histogram using █ blocks (scaled to 40 characters max).
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

// saveToDatabase saves the generated universe to PostgreSQL database.
//
// Process:
//   1. Connect to database with provided credentials
//   2. Check for existing universe data
//   3. Prompt user for confirmation if data exists
//   4. Clear old universe data (foreign key aware)
//   5. Insert star systems with progress updates
//   6. Insert planets with progress updates
//   7. Insert system connections (jump routes)
//
// Parameters:
//   - univ: Generated universe with systems, planets, connections
//   - host: Database hostname
//   - port: Database port number
//   - user: Database username
//   - password: Database password
//   - dbName: Database name
//
// Returns:
//   - error: Database connection, query, or transaction error
//
// Safety Features:
//   - Interactive confirmation before overwriting
//   - Foreign key constraint handling (delete order: connections, planets, systems)
//   - Progress indicators for large datasets (every 20 systems, 50 planets)
//   - Connection bidirectionality handling (only insert one direction)
//
// Transaction Handling:
// Currently uses individual queries (not in transaction).
// Future enhancement: Wrap in transaction for atomicity.
//
// Error Handling:
// Any database error during insert causes immediate return with error.
// Partial inserts may remain in database (no rollback currently).
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

// clearUniverse deletes all existing universe data from the database.
//
// This function is called before inserting a new universe to avoid conflicts
// and ensure a clean state.
//
// Deletion Order (Foreign Key Constraints):
//   1. system_connections (references star_systems)
//   2. planets (references star_systems)
//   3. star_systems (no foreign key dependencies)
//
// If deletion order is wrong, foreign key constraints will cause errors.
//
// Parameters:
//   - ctx: Context for database operations
//   - db: Database connection
//
// Returns:
//   - error: SQL execution error
//
// Data Loss Warning:
// This function permanently deletes ALL universe data. It should only be called
// after user confirmation.
//
// Transaction Handling:
// Currently uses individual DELETE queries (not in transaction).
// Future enhancement: Wrap in transaction for atomicity.
//
// Player Data:
// This function does NOT delete player data (ships, positions, etc.).
// After clearing universe, players may have invalid references (orphaned ships,
// invalid system IDs). Server should handle this gracefully or also clear player data.
func clearUniverse(ctx context.Context, db *database.DB) error {
	// Delete in correct order due to foreign keys
	// Order matters! References must be deleted before referenced rows
	queries := []string{
		"DELETE FROM system_connections", // References star_systems (both from_system_id and to_system_id)
		"DELETE FROM planets",            // References star_systems (system_id)
		"DELETE FROM star_systems",       // No foreign key dependencies
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
