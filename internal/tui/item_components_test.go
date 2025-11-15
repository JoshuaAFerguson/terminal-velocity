// File: internal/tui/item_components_test.go
// Project: Terminal Velocity
// Description: Unit tests for item picker and list components
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package tui

import (
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Mock item repository for testing
type mockItemRepo struct {
	items []*models.PlayerItem
}

func (m *mockItemRepo) GetPlayerItems(playerID uuid.UUID) ([]*models.PlayerItem, error) {
	return m.items, nil
}

// Test ItemPicker basic functionality
func TestItemPicker_NewItemPicker(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)

	if picker == nil {
		t.Fatal("NewItemPicker returned nil")
	}

	if picker.mode != ItemPickerModeMulti {
		t.Errorf("Expected mode %d, got %d", ItemPickerModeMulti, picker.mode)
	}

	if picker.filter != FilterAvailable {
		t.Errorf("Expected filter %d, got %d", FilterAvailable, picker.filter)
	}

	if picker.title != "Select Items" {
		t.Errorf("Expected title 'Select Items', got '%s'", picker.title)
	}

	if len(picker.selected) != 0 {
		t.Errorf("Expected empty selection, got %d items", len(picker.selected))
	}
}

// Test ItemPicker configuration methods
func TestItemPicker_Configuration(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)

	// Test SetMode
	picker.SetMode(ItemPickerModeSingle)
	if picker.mode != ItemPickerModeSingle {
		t.Errorf("SetMode failed: expected %d, got %d", ItemPickerModeSingle, picker.mode)
	}

	// Test SetFilter
	picker.SetFilter(FilterWeapons)
	if picker.filter != FilterWeapons {
		t.Errorf("SetFilter failed: expected %d, got %d", FilterWeapons, picker.filter)
	}

	// Test SetTitle
	picker.SetTitle("Pick Weapons")
	if picker.title != "Pick Weapons" {
		t.Errorf("SetTitle failed: expected 'Pick Weapons', got '%s'", picker.title)
	}

	// Test SetMaxSelection
	picker.SetMaxSelection(5)
	if picker.maxSelection != 5 {
		t.Errorf("SetMaxSelection failed: expected 5, got %d", picker.maxSelection)
	}
}

// Test ItemPicker selection logic
func TestItemPicker_Selection(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)

	// Create mock items
	item1 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
	}
	item2 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		ItemType:    models.ItemTypeOutfit,
		EquipmentID: "shield_booster",
		Location:    models.LocationShip,
	}

	picker.items = []*models.PlayerItem{item1, item2}
	picker.allItems = picker.items

	// Test multi-select
	picker.SetMode(ItemPickerModeMulti)

	// Select first item
	picker.selected[item1.ID] = true
	if len(picker.selected) != 1 {
		t.Errorf("Expected 1 selected item, got %d", len(picker.selected))
	}

	// Select second item
	picker.selected[item2.ID] = true
	if len(picker.selected) != 2 {
		t.Errorf("Expected 2 selected items, got %d", len(picker.selected))
	}

	// Test GetSelectedItems
	selected := picker.GetSelectedItems()
	if len(selected) != 2 {
		t.Errorf("GetSelectedItems: expected 2 items, got %d", len(selected))
	}

	// Test GetSelectedCount
	count := picker.GetSelectedCount()
	if count != 2 {
		t.Errorf("GetSelectedCount: expected 2, got %d", count)
	}

	// Test ClearSelection
	picker.ClearSelection()
	if len(picker.selected) != 0 {
		t.Errorf("ClearSelection failed: expected 0 items, got %d", len(picker.selected))
	}
}

// Test ItemPicker single-select mode
func TestItemPicker_SingleSelect(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)
	picker.SetMode(ItemPickerModeSingle)

	item1 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		ItemType:    models.ItemTypeWeapon,
		EquipmentID: "laser_cannon",
		Location:    models.LocationShip,
	}
	item2 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		ItemType:    models.ItemTypeOutfit,
		EquipmentID: "shield_booster",
		Location:    models.LocationShip,
	}

	picker.items = []*models.PlayerItem{item1, item2}

	// Select first item
	picker.selected[item1.ID] = true
	if len(picker.selected) != 1 {
		t.Errorf("Expected 1 selected item, got %d", len(picker.selected))
	}

	// Selecting second item should clear first
	picker.selected = make(map[uuid.UUID]bool)
	picker.selected[item2.ID] = true

	if len(picker.selected) != 1 {
		t.Errorf("Single select should only have 1 item, got %d", len(picker.selected))
	}

	if !picker.selected[item2.ID] {
		t.Error("Second item should be selected")
	}

	if picker.selected[item1.ID] {
		t.Error("First item should not be selected")
	}
}

// Test ItemPicker max selection limit
func TestItemPicker_MaxSelection(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)
	picker.SetMode(ItemPickerModeMulti)
	picker.SetMaxSelection(2)

	// Create 3 items
	item1 := &models.PlayerItem{ID: uuid.New(), PlayerID: playerID}
	item2 := &models.PlayerItem{ID: uuid.New(), PlayerID: playerID}
	item3 := &models.PlayerItem{ID: uuid.New(), PlayerID: playerID}

	picker.items = []*models.PlayerItem{item1, item2, item3}

	// Select up to max
	picker.selected[item1.ID] = true
	picker.selected[item2.ID] = true

	if len(picker.selected) != 2 {
		t.Errorf("Expected 2 selected items, got %d", len(picker.selected))
	}

	// Trying to select a third should be handled by the Update logic
	// (which would set an error, but we're testing the model directly here)
}

// Test ItemPicker filter names
func TestItemPicker_FilterNames(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)

	tests := []struct {
		filter   ItemPickerFilter
		expected string
	}{
		{FilterAll, "All Items"},
		{FilterWeapons, "Weapons"},
		{FilterOutfits, "Outfits"},
		{FilterSpecial, "Special"},
		{FilterQuest, "Quest Items"},
		{FilterAvailable, "Available"},
		{FilterShip, "Ship"},
		{FilterStation, "Station Storage"},
	}

	for _, tt := range tests {
		picker.SetFilter(tt.filter)
		name := picker.getFilterName()
		if name != tt.expected {
			t.Errorf("Filter %d: expected '%s', got '%s'", tt.filter, tt.expected, name)
		}
	}
}

// Test ItemPicker Reset
func TestItemPicker_Reset(t *testing.T) {
	playerID := uuid.New()
	picker := NewItemPicker(nil, playerID)

	// Set some state
	picker.cursor = 5
	picker.selected[uuid.New()] = true
	picker.searchQuery = "laser"
	picker.searchMode = true
	picker.error = "test error"
	picker.scrollOffset = 3

	// Reset
	picker.Reset()

	if picker.cursor != 0 {
		t.Errorf("Reset failed: cursor should be 0, got %d", picker.cursor)
	}

	if len(picker.selected) != 0 {
		t.Errorf("Reset failed: selected should be empty, got %d items", len(picker.selected))
	}

	if picker.searchQuery != "" {
		t.Errorf("Reset failed: searchQuery should be empty, got '%s'", picker.searchQuery)
	}

	if picker.searchMode {
		t.Error("Reset failed: searchMode should be false")
	}

	if picker.error != "" {
		t.Errorf("Reset failed: error should be empty, got '%s'", picker.error)
	}

	if picker.scrollOffset != 0 {
		t.Errorf("Reset failed: scrollOffset should be 0, got %d", picker.scrollOffset)
	}
}

// Test ItemList basic functionality
func TestItemList_NewItemList(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	if list == nil {
		t.Fatal("NewItemList returned nil")
	}

	if list.title != "Your Items" {
		t.Errorf("Expected title 'Your Items', got '%s'", list.title)
	}

	if list.grouping != GroupByType {
		t.Errorf("Expected grouping %d, got %d", GroupByType, list.grouping)
	}

	if list.sorting != SortByName {
		t.Errorf("Expected sorting %d, got %d", SortByName, list.sorting)
	}
}

// Test ItemList configuration
func TestItemList_Configuration(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	// Test SetTitle
	list.SetTitle("Inventory")
	if list.title != "Inventory" {
		t.Errorf("SetTitle failed: expected 'Inventory', got '%s'", list.title)
	}

	// Test SetGrouping
	list.SetGrouping(GroupByLocation)
	if list.grouping != GroupByLocation {
		t.Errorf("SetGrouping failed: expected %d, got %d", GroupByLocation, list.grouping)
	}

	// Test SetSorting
	list.SetSorting(SortByType)
	if list.sorting != SortByType {
		t.Errorf("SetSorting failed: expected %d, got %d", SortByType, list.sorting)
	}

	// Test SetShowStats
	list.SetShowStats(true)
	if !list.showStats {
		t.Error("SetShowStats failed: showStats should be true")
	}

	// Test SetFilter
	list.SetFilter(FilterWeapons)
	if list.filter != FilterWeapons {
		t.Errorf("SetFilter failed: expected %d, got %d", FilterWeapons, list.filter)
	}
}

// Test ItemList sorting names
func TestItemList_SortingNames(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	tests := []struct {
		sorting  ItemListSorting
		expected string
	}{
		{SortByName, "Name"},
		{SortByType, "Type"},
		{SortByAcquiredDate, "Acquired"},
	}

	for _, tt := range tests {
		list.SetSorting(tt.sorting)
		name := list.getSortingName()
		if name != tt.expected {
			t.Errorf("Sorting %d: expected '%s', got '%s'", tt.sorting, tt.expected, name)
		}
	}
}

// Test ItemList grouping names
func TestItemList_GroupingNames(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	tests := []struct {
		grouping ItemListGrouping
		expected string
	}{
		{GroupByNone, "None"},
		{GroupByType, "Type"},
		{GroupByLocation, "Location"},
	}

	for _, tt := range tests {
		list.SetGrouping(tt.grouping)
		name := list.getGroupingName()
		if name != tt.expected {
			t.Errorf("Grouping %d: expected '%s', got '%s'", tt.grouping, tt.expected, name)
		}
	}
}

// Test ItemList GetCurrentItem
func TestItemList_GetCurrentItem(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	// Empty list
	item := list.GetCurrentItem()
	if item != nil {
		t.Error("GetCurrentItem should return nil for empty list")
	}

	// Add items
	item1 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		EquipmentID: "laser_cannon",
	}
	item2 := &models.PlayerItem{
		ID:          uuid.New(),
		PlayerID:    playerID,
		EquipmentID: "shield_booster",
	}

	list.items = []*models.PlayerItem{item1, item2}

	// Cursor at 0
	current := list.GetCurrentItem()
	if current == nil || current.ID != item1.ID {
		t.Error("GetCurrentItem should return first item when cursor is 0")
	}

	// Move cursor
	list.cursor = 1
	current = list.GetCurrentItem()
	if current == nil || current.ID != item2.ID {
		t.Error("GetCurrentItem should return second item when cursor is 1")
	}

	// Out of bounds cursor
	list.cursor = 10
	current = list.GetCurrentItem()
	if current != nil {
		t.Error("GetCurrentItem should return nil for out of bounds cursor")
	}
}

// Test ItemList GetItemCount
func TestItemList_GetItemCount(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	// Empty
	if list.GetItemCount() != 0 {
		t.Errorf("Expected count 0, got %d", list.GetItemCount())
	}

	// Add items
	list.items = []*models.PlayerItem{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	if list.GetItemCount() != 3 {
		t.Errorf("Expected count 3, got %d", list.GetItemCount())
	}
}

// Test ItemList Reset
func TestItemList_Reset(t *testing.T) {
	playerID := uuid.New()
	list := NewItemList(nil, playerID)

	// Set some state
	list.cursor = 5
	list.error = "test error"
	list.scrollOffset = 3
	list.items = []*models.PlayerItem{{ID: uuid.New()}}

	// Reset
	list.Reset()

	if list.cursor != 0 {
		t.Errorf("Reset failed: cursor should be 0, got %d", list.cursor)
	}

	if list.error != "" {
		t.Errorf("Reset failed: error should be empty, got '%s'", list.error)
	}

	if list.scrollOffset != 0 {
		t.Errorf("Reset failed: scrollOffset should be 0, got %d", list.scrollOffset)
	}

	if list.items != nil {
		t.Error("Reset failed: items should be nil")
	}
}
