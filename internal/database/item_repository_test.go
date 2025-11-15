// File: internal/database/item_repository_test.go
// Project: Terminal Velocity
// Description: Unit tests for item repository CRUD operations
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"terminal-velocity/internal/models"
)

// Test helper to create a test player for item tests
func createTestPlayer(t *testing.T, repo *PlayerRepository, ctx context.Context) *models.Player {
	t.Helper()

	username := "testplayer_" + uuid.New().String()[:8]
	password := "testpass123"

	player, err := repo.Create(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to create test player: %v", err)
	}

	return player
}

func TestItemRepository_CreateItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	// Create a test ship for the player
	shipID := uuid.New()

	// Create test item
	item := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{"mods":["damage_boost"]}`),
	}

	err := itemRepo.CreateItem(ctx, item)
	if err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	// Verify item was created
	if item.ID == uuid.Nil {
		t.Error("Item ID should not be nil")
	}

	if item.AcquiredAt.IsZero() {
		t.Error("AcquiredAt should be set")
	}

	if item.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	// Cleanup
	_ = itemRepo.DeleteItem(ctx, item.ID)
}

func TestItemRepository_GetPlayerItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create multiple test items
	items := []*models.PlayerItem{
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "laser_cannon",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeOutfit,
			EquipmentID: "shield_booster",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
	}

	for _, item := range items {
		if err := itemRepo.CreateItem(ctx, item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}
		defer func(id uuid.UUID) { _ = itemRepo.DeleteItem(ctx, id) }(item.ID)
	}

	// Get all player items
	playerItems, err := itemRepo.GetPlayerItems(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player items: %v", err)
	}

	if len(playerItems) != 2 {
		t.Errorf("Expected 2 items, got %d", len(playerItems))
	}
}

func TestItemRepository_GetItemByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create test item
	item := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "plasma_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{"upgrades":{"damage":2}}`),
	}

	if err := itemRepo.CreateItem(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}
	defer func() { _ = itemRepo.DeleteItem(ctx, item.ID) }()

	// Get item by ID
	retrieved, err := itemRepo.GetItemByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get item by ID: %v", err)
	}

	if retrieved.ID != item.ID {
		t.Errorf("Expected item ID %s, got %s", item.ID, retrieved.ID)
	}

	if retrieved.EquipmentID != "plasma_cannon" {
		t.Errorf("Expected equipment_id 'plasma_cannon', got %s", retrieved.EquipmentID)
	}

	// Test non-existent item
	_, err = itemRepo.GetItemByID(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error when getting non-existent item")
	}
}

func TestItemRepository_GetItemsByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create weapons and outfits
	items := []*models.PlayerItem{
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "laser_cannon",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "missile_launcher",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeOutfit,
			EquipmentID: "shield_booster",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
	}

	for _, item := range items {
		if err := itemRepo.CreateItem(ctx, item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}
		defer func(id uuid.UUID) { _ = itemRepo.DeleteItem(ctx, id) }(item.ID)
	}

	// Get weapons only
	weapons, err := itemRepo.GetItemsByType(ctx, player.ID, models.ItemTypeWeapon)
	if err != nil {
		t.Fatalf("Failed to get weapons: %v", err)
	}

	if len(weapons) != 2 {
		t.Errorf("Expected 2 weapons, got %d", len(weapons))
	}

	// Get outfits only
	outfits, err := itemRepo.GetItemsByType(ctx, player.ID, models.ItemTypeOutfit)
	if err != nil {
		t.Fatalf("Failed to get outfits: %v", err)
	}

	if len(outfits) != 1 {
		t.Errorf("Expected 1 outfit, got %d", len(outfits))
	}
}

func TestItemRepository_GetAvailableItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()
	stationID := uuid.New()

	// Create items in different locations
	items := []*models.PlayerItem{
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "laser_cannon",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeOutfit,
			EquipmentID: "shield_booster",
			Location:    models.LocationStationStorage,
			LocationID:  &stationID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeSpecial,
			EquipmentID: "rare_artifact",
			Location:    models.LocationMail,
			LocationID:  nil, // In mail, doesn't need location_id
			Properties:  json.RawMessage(`{}`),
		},
	}

	for _, item := range items {
		// Temporarily bypass validation for mail items
		if item.Location == models.LocationMail {
			item.Properties = json.RawMessage(`{}`)
			query := `
				INSERT INTO player_items (player_id, item_type, equipment_id, location, location_id, properties)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id, acquired_at, created_at, updated_at
			`
			err := db.QueryRow(ctx, query, item.PlayerID, item.ItemType, item.EquipmentID, item.Location, item.LocationID, item.Properties).
				Scan(&item.ID, &item.AcquiredAt, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				t.Fatalf("Failed to create mail item: %v", err)
			}
		} else {
			if err := itemRepo.CreateItem(ctx, item); err != nil {
				t.Fatalf("Failed to create item: %v", err)
			}
		}
		defer func(id uuid.UUID) { _ = itemRepo.DeleteItem(ctx, id) }(item.ID)
	}

	// Get available items (ship + station storage only)
	available, err := itemRepo.GetAvailableItems(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get available items: %v", err)
	}

	if len(available) != 2 {
		t.Errorf("Expected 2 available items, got %d", len(available))
	}

	// Verify mail item is not in available list
	for _, item := range available {
		if item.Location == models.LocationMail {
			t.Error("Mail items should not be in available items list")
		}
	}
}

func TestItemRepository_UpdateItemLocation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()
	stationID := uuid.New()

	// Create item on ship
	item := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	if err := itemRepo.CreateItem(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}
	defer func() { _ = itemRepo.DeleteItem(ctx, item.ID) }()

	// Move to station storage
	err := itemRepo.UpdateItemLocation(ctx, item.ID, models.LocationStationStorage, &stationID)
	if err != nil {
		t.Fatalf("Failed to update item location: %v", err)
	}

	// Verify location changed
	updated, err := itemRepo.GetItemByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get updated item: %v", err)
	}

	if updated.Location != models.LocationStationStorage {
		t.Errorf("Expected location 'station_storage', got %s", updated.Location)
	}

	if updated.LocationID == nil || *updated.LocationID != stationID {
		t.Error("LocationID should be set to station ID")
	}
}

func TestItemRepository_UpdateItemProperties(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create item
	item := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	if err := itemRepo.CreateItem(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}
	defer func() { _ = itemRepo.DeleteItem(ctx, item.ID) }()

	// Update properties
	props := &models.ItemProperties{
		Mods:     []string{"damage_boost", "accuracy"},
		Upgrades: map[string]int{"damage": 3, "range": 2},
	}

	err := itemRepo.UpdateItemProperties(ctx, item.ID, props)
	if err != nil {
		t.Fatalf("Failed to update item properties: %v", err)
	}

	// Verify properties updated
	updated, err := itemRepo.GetItemByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get updated item: %v", err)
	}

	retrievedProps, err := updated.GetProperties()
	if err != nil {
		t.Fatalf("Failed to get properties: %v", err)
	}

	if len(retrievedProps.Mods) != 2 {
		t.Errorf("Expected 2 mods, got %d", len(retrievedProps.Mods))
	}

	if retrievedProps.Upgrades["damage"] != 3 {
		t.Errorf("Expected damage upgrade 3, got %d", retrievedProps.Upgrades["damage"])
	}
}

func TestItemRepository_TransferItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	// Create two players
	player1 := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player1.ID) }()

	player2 := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player2.ID) }()

	shipID := uuid.New()

	// Create item owned by player1
	item := &models.PlayerItem{
		PlayerID:    player1.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	if err := itemRepo.CreateItem(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}
	defer func() { _ = itemRepo.DeleteItem(ctx, item.ID) }()

	// Transfer to player2
	tradeID := uuid.New()
	err := itemRepo.TransferItem(ctx, item.ID, player2.ID, "trade", &tradeID)
	if err != nil {
		t.Fatalf("Failed to transfer item: %v", err)
	}

	// Verify ownership changed
	transferred, err := itemRepo.GetItemByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get transferred item: %v", err)
	}

	if transferred.PlayerID != player2.ID {
		t.Errorf("Expected new owner %s, got %s", player2.ID, transferred.PlayerID)
	}

	// Verify audit log
	history, err := itemRepo.GetItemTransferHistory(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get transfer history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 transfer record, got %d", len(history))
	}

	if history[0].FromPlayerID == nil || *history[0].FromPlayerID != player1.ID {
		t.Error("Transfer should record from_player_id as player1")
	}

	if history[0].ToPlayerID == nil || *history[0].ToPlayerID != player2.ID {
		t.Error("Transfer should record to_player_id as player2")
	}

	if history[0].TransferType != "trade" {
		t.Errorf("Expected transfer_type 'trade', got %s", history[0].TransferType)
	}
}

func TestItemRepository_DeleteItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create item
	item := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	if err := itemRepo.CreateItem(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	// Delete item
	err := itemRepo.DeleteItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Verify item no longer exists
	_, err = itemRepo.GetItemByID(ctx, item.ID)
	if err == nil {
		t.Error("Expected error when getting deleted item")
	}

	// Test deleting non-existent item
	err = itemRepo.DeleteItem(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error when deleting non-existent item")
	}
}

func TestItemRepository_CountPlayerItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create multiple items
	items := []*models.PlayerItem{
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "laser_cannon",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "missile_launcher",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeOutfit,
			EquipmentID: "shield_booster",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
	}

	for _, item := range items {
		if err := itemRepo.CreateItem(ctx, item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}
		defer func(id uuid.UUID) { _ = itemRepo.DeleteItem(ctx, id) }(item.ID)
	}

	// Count all items
	count, err := itemRepo.CountPlayerItems(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to count items: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 items, got %d", count)
	}

	// Count weapons only
	weaponCount, err := itemRepo.CountItemsByType(ctx, player.ID, models.ItemTypeWeapon)
	if err != nil {
		t.Fatalf("Failed to count weapons: %v", err)
	}

	if weaponCount != 2 {
		t.Errorf("Expected 2 weapons, got %d", weaponCount)
	}

	// Count outfits only
	outfitCount, err := itemRepo.CountItemsByType(ctx, player.ID, models.ItemTypeOutfit)
	if err != nil {
		t.Fatalf("Failed to count outfits: %v", err)
	}

	if outfitCount != 1 {
		t.Errorf("Expected 1 outfit, got %d", outfitCount)
	}
}

func TestItemRepository_BatchOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Create multiple items in batch
	items := []*models.PlayerItem{
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "laser_cannon",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
		{
			PlayerID:    player.ID,
			ItemType:    models.ItemTypeWeapon,
			EquipmentID: "missile_launcher",
			Location:    models.LocationShip,
			LocationID:  &shipID,
			Properties:  json.RawMessage(`{}`),
		},
	}

	err := itemRepo.CreateItems(ctx, items)
	if err != nil {
		t.Fatalf("Failed to create items in batch: %v", err)
	}
	defer func() {
		for _, item := range items {
			_ = itemRepo.DeleteItem(ctx, item.ID)
		}
	}()

	// Verify all items have IDs
	for i, item := range items {
		if item.ID == uuid.Nil {
			t.Errorf("Item %d should have ID after batch create", i)
		}
	}

	// Get items by IDs
	itemIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}

	retrieved, err := itemRepo.GetItemsByIDs(ctx, itemIDs)
	if err != nil {
		t.Fatalf("Failed to get items by IDs: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 items, got %d", len(retrieved))
	}

	// Delete items in batch
	err = itemRepo.DeleteItems(ctx, itemIDs)
	if err != nil {
		t.Fatalf("Failed to delete items in batch: %v", err)
	}

	// Verify items deleted
	count, err := itemRepo.CountPlayerItems(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to count items: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 items after batch delete, got %d", count)
	}
}

func TestItemRepository_Validation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	playerRepo := NewPlayerRepository(db)
	ctx := context.Background()

	player := createTestPlayer(t, playerRepo, ctx)
	defer func() { _ = playerRepo.Delete(ctx, player.ID) }()

	shipID := uuid.New()

	// Test invalid item type
	invalidItem := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemType("invalid"),
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	err := itemRepo.CreateItem(ctx, invalidItem)
	if err == nil {
		t.Error("Expected error when creating item with invalid type")
		_ = itemRepo.DeleteItem(ctx, invalidItem.ID)
	}

	// Test missing location_id for ship location
	missingLocationID := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  nil, // Should be required for ship
		Properties:  json.RawMessage(`{}`),
	}

	err = itemRepo.CreateItem(ctx, missingLocationID)
	if err == nil {
		t.Error("Expected error when creating ship item without location_id")
		_ = itemRepo.DeleteItem(ctx, missingLocationID.ID)
	}

	// Test invalid transfer type
	validItem := &models.PlayerItem{
		PlayerID:    player.ID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
		LocationID:  &shipID,
		Properties:  json.RawMessage(`{}`),
	}

	if err := itemRepo.CreateItem(ctx, validItem); err != nil {
		t.Fatalf("Failed to create valid item: %v", err)
	}
	defer func() { _ = itemRepo.DeleteItem(ctx, validItem.ID) }()

	err = itemRepo.TransferItem(ctx, validItem.ID, player.ID, "invalid_type", nil)
	if err == nil {
		t.Error("Expected error when transferring with invalid transfer type")
	}
}
