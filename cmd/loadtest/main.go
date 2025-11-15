// File: cmd/loadtest/main.go
// Project: Terminal Velocity
// Description: Load testing tool for inventory system with 1000+ items
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var (
	dbHost     = flag.String("db-host", "localhost", "Database host")
	dbPort     = flag.Int("db-port", 5432, "Database port")
	dbUser     = flag.String("db-user", "terminal_velocity", "Database user")
	dbPassword = flag.String("db-password", "", "Database password")
	dbName     = flag.String("db-name", "terminal_velocity", "Database name")
	itemCount  = flag.Int("items", 1000, "Number of items to create")
	playerName = flag.String("player", "loadtest", "Player username to create items for")
	cleanup    = flag.Bool("cleanup", false, "Clean up test items after load test")
)

type LoadTestResult struct {
	ItemCount        int
	InsertTime       time.Duration
	AvgInsertTime    time.Duration
	QueryTime        time.Duration
	QueryCount       int
	PaginationTime   time.Duration
	ItemsPerSecond   float64
	QueriesPerSecond float64
	Success          bool
	Errors           []string
}

func main() {
	flag.Parse()

	if *dbPassword == "" {
		fmt.Println("Error: Database password required")
		flag.Usage()
		os.Exit(1)
	}

	// Connect to database
	cfg := &database.Config{
		Host:     *dbHost,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPassword,
		Database: *dbName,
		SSLMode:  "disable",
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	playerRepo := database.NewPlayerRepository(db)
	itemRepo := database.NewItemRepository(db)

	ctx := context.Background()

	// Get or create test player
	player, err := playerRepo.GetByUsername(ctx, *playerName)
	if err != nil {
		log.Printf("Player %s not found, creating...", *playerName)
		player, err = playerRepo.Create(ctx, *playerName, "loadtest123")
		if err != nil {
			log.Fatalf("Failed to create player: %v", err)
		}
	}

	fmt.Printf("=== Terminal Velocity Inventory Load Test ===\n\n")
	fmt.Printf("Player: %s (ID: %s)\n", player.Username, player.ID)
	fmt.Printf("Target Items: %d\n\n", *itemCount)

	result := &LoadTestResult{
		ItemCount: *itemCount,
		Errors:    []string{},
	}

	// Clean up first if requested
	if *cleanup {
		fmt.Println("Cleaning up existing test items...")
		if err := cleanupItems(ctx, itemRepo, player.ID); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
		fmt.Println("Cleanup complete!\n")
		return
	}

	// Phase 1: Insert items
	fmt.Printf("Phase 1: Inserting %d items...\n", *itemCount)
	insertStart := time.Now()

	itemTypes := []models.ItemType{
		models.ItemTypeWeapon,
		models.ItemTypeOutfit,
		models.ItemTypeSpecial,
	}

	locations := []models.ItemLocation{
		models.LocationShip,
		models.LocationStationStorage,
	}

	equipmentIDs := []string{
		"laser_cannon_mk1", "laser_cannon_mk2", "pulse_laser",
		"shield_generator_mk1", "shield_generator_mk2",
		"cargo_expansion", "fuel_tank_expansion",
		"targeting_computer", "advanced_radar",
	}

	// Ship and station IDs for location_id
	shipID := player.ShipID
	stationID := uuid.New() // Mock station

	for i := 0; i < *itemCount; i++ {
		itemType := itemTypes[rand.Intn(len(itemTypes))]
		location := locations[rand.Intn(len(locations))]
		equipmentID := equipmentIDs[rand.Intn(len(equipmentIDs))]

		var locationID *uuid.UUID
		if location == models.LocationShip {
			locationID = &shipID
		} else {
			locationID = &stationID
		}

		properties := map[string]interface{}{
			"damage":       rand.Intn(100) + 10,
			"accuracy":     rand.Float64(),
			"weight":       rand.Float64() * 100,
			"tech_level":   rand.Intn(10) + 1,
			"test_item":    true,
			"batch_number": i,
		}

		propJSON, _ := json.Marshal(properties)

		item := &models.PlayerItem{
			ID:          uuid.New(),
			PlayerID:    player.ID,
			ItemType:    itemType,
			EquipmentID: equipmentID,
			Location:    location,
			LocationID:  locationID,
			Properties:  propJSON,
			AcquiredAt:  time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := itemRepo.CreateItem(ctx, item); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Insert %d failed: %v", i, err))
			if len(result.Errors) > 10 {
				break
			}
		}

		// Progress indicator
		if (i+1)%100 == 0 {
			fmt.Printf("  Inserted %d/%d items...\n", i+1, *itemCount)
		}
	}

	result.InsertTime = time.Since(insertStart)
	result.AvgInsertTime = result.InsertTime / time.Duration(*itemCount)
	result.ItemsPerSecond = float64(*itemCount) / result.InsertTime.Seconds()

	fmt.Printf("✓ Insert complete: %v (avg: %v per item, %.2f items/sec)\n\n",
		result.InsertTime, result.AvgInsertTime, result.ItemsPerSecond)

	// Phase 2: Query performance
	fmt.Println("Phase 2: Testing query performance...")
	queryStart := time.Now()
	result.QueryCount = 10

	for i := 0; i < result.QueryCount; i++ {
		items, err := itemRepo.GetPlayerItems(ctx, player.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Query %d failed: %v", i, err))
		}
		fmt.Printf("  Query %d: Retrieved %d items\n", i+1, len(items))
	}

	result.QueryTime = time.Since(queryStart)
	result.QueriesPerSecond = float64(result.QueryCount) / result.QueryTime.Seconds()

	fmt.Printf("✓ Query complete: %v for %d queries (avg: %v per query, %.2f queries/sec)\n\n",
		result.QueryTime, result.QueryCount, result.QueryTime/time.Duration(result.QueryCount), result.QueriesPerSecond)

	// Phase 3: Item filtering performance
	fmt.Println("Phase 3: Testing item filtering...")
	filterStart := time.Now()

	// Test filtering by type
	for _, itemType := range itemTypes {
		items, err := itemRepo.GetItemsByType(ctx, player.ID, itemType)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Filter by type %s failed: %v", itemType, err))
		}
		fmt.Printf("  Filtered by type %s: %d items\n", itemType, len(items))
	}

	// Test filtering by location
	for _, location := range locations {
		var locationID uuid.UUID
		if location == models.LocationShip {
			locationID = shipID
		} else {
			locationID = stationID
		}
		items, err := itemRepo.GetItemsByLocation(ctx, player.ID, location, locationID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Filter by location %s failed: %v", location, err))
		}
		fmt.Printf("  Filtered by location %s: %d items\n", location, len(items))
	}

	result.PaginationTime = time.Since(filterStart)

	fmt.Printf("✓ Filtering complete: %v\n\n", result.PaginationTime)

	// Summary
	result.Success = len(result.Errors) == 0

	fmt.Println("=== Load Test Results ===")
	fmt.Printf("Item Count: %d\n", result.ItemCount)
	fmt.Printf("Insert Time: %v (%.2f items/sec)\n", result.InsertTime, result.ItemsPerSecond)
	fmt.Printf("Query Time: %v for %d queries (%.2f queries/sec)\n", result.QueryTime, result.QueryCount, result.QueriesPerSecond)
	fmt.Printf("Filter Time: %v\n", result.PaginationTime)
	fmt.Printf("Total Time: %v\n", result.InsertTime+result.QueryTime+result.PaginationTime)

	if result.Success {
		fmt.Println("\n✓ All tests PASSED")
	} else {
		fmt.Printf("\n✗ %d errors occurred:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Println("\nTo clean up test items, run:")
	fmt.Printf("  %s -cleanup -player=%s -db-password=%s\n", os.Args[0], *playerName, *dbPassword)
}

func cleanupItems(ctx context.Context, itemRepo *database.ItemRepository, playerID uuid.UUID) error {
	// Get all items for player
	items, err := itemRepo.GetPlayerItems(ctx, playerID)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d items to delete\n", len(items))

	// Delete items with test_item property
	deleted := 0
	for _, item := range items {
		var props map[string]interface{}
		if err := json.Unmarshal(item.Properties, &props); err == nil {
			if isTest, ok := props["test_item"].(bool); ok && isTest {
				if err := itemRepo.DeleteItem(ctx, item.ID); err != nil {
					return fmt.Errorf("failed to delete item %s: %v", item.ID, err)
				}
				deleted++
				if deleted%100 == 0 {
					fmt.Printf("  Deleted %d items...\n", deleted)
				}
			}
		}
	}

	fmt.Printf("Deleted %d test items\n", deleted)
	return nil
}
