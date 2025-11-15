// File: internal/tui/marketplace_form_test.go
// Project: Terminal Velocity
// Description: Form logic tests for marketplace
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestContractCreation_FormInitialization tests that contract form initializes correctly
func TestContractCreation_FormInitialization(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract

	// Simulate key press to trigger form initialization
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, _ = model.updateMarketplaceCreateContract(msg)

	// Check form was initialized
	if _, exists := model.marketplace.createForm["contract_type"]; !exists {
		t.Error("Contract type field not initialized")
	}

	if model.marketplace.createForm["contract_type"] != "0" {
		t.Errorf("Expected contract_type '0', got '%s'", model.marketplace.createForm["contract_type"])
	}

	if model.marketplace.createForm["reward"] != "10000" {
		t.Errorf("Expected default reward '10000', got '%s'", model.marketplace.createForm["reward"])
	}

	if model.marketplace.createForm["duration"] != "48" {
		t.Errorf("Expected default duration '48', got '%s'", model.marketplace.createForm["duration"])
	}
}

// TestContractCreation_TabNavigation tests tab navigation through contract form fields
func TestContractCreation_TabNavigation(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{
		"contract_type": "0",
		"title":         "",
		"description":   "",
		"reward":        "10000",
		"target_name":   "",
		"duration":      "48",
	}
	model.marketplace.formField = 0

	// Press Tab 6 times (should cycle through all 6 fields and wrap to 0)
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	for i := 0; i < 6; i++ {
		_, _ = model.updateMarketplaceCreateContract(tabMsg)
	}

	// Should wrap back to field 0
	if model.marketplace.formField != 0 {
		t.Errorf("Expected formField 0 after 6 tabs, got %d", model.marketplace.formField)
	}
}

// TestContractCreation_ArrowKeyTypeSelection tests arrow key contract type selection
func TestContractCreation_ArrowKeyTypeSelection(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{
		"contract_type": "0",
	}
	model.marketplace.formField = 0

	// Press right arrow (should increment type to 1)
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	_, _ = model.updateMarketplaceCreateContract(rightMsg)

	if model.marketplace.createForm["contract_type"] != "1" {
		t.Errorf("Expected contract_type '1' after right arrow, got '%s'", model.marketplace.createForm["contract_type"])
	}

	// Press right 3 more times (should wrap to 0: 1->2->3->0)
	for i := 0; i < 3; i++ {
		_, _ = model.updateMarketplaceCreateContract(rightMsg)
	}

	if model.marketplace.createForm["contract_type"] != "0" {
		t.Errorf("Expected contract_type '0' after wrapping, got '%s'", model.marketplace.createForm["contract_type"])
	}

	// Press left arrow (should go to 3)
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	_, _ = model.updateMarketplaceCreateContract(leftMsg)

	if model.marketplace.createForm["contract_type"] != "3" {
		t.Errorf("Expected contract_type '3' after left arrow, got '%s'", model.marketplace.createForm["contract_type"])
	}
}

// TestContractCreation_TextInput tests text input in form fields
func TestContractCreation_TextInput(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{
		"contract_type": "0",
		"title":         "",
		"description":   "",
		"reward":        "10000",
		"target_name":   "",
		"duration":      "48",
	}
	model.marketplace.formField = 1 // Title field

	// Type "Test" into title field
	chars := []rune{'T', 'e', 's', 't'}
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		_, _ = model.updateMarketplaceCreateContract(msg)
	}

	if model.marketplace.createForm["title"] != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", model.marketplace.createForm["title"])
	}
}

// TestContractCreation_BackspaceInput tests backspace functionality
func TestContractCreation_Backspace(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{
		"contract_type": "0",
		"title":         "TestTitle",
	}
	model.marketplace.formField = 1 // Title field

	// Press backspace 3 times
	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	for i := 0; i < 3; i++ {
		_, _ = model.updateMarketplaceCreateContract(backspaceMsg)
	}

	if model.marketplace.createForm["title"] != "TestTi" {
		t.Errorf("Expected title 'TestTi' after 3 backspaces, got '%s'", model.marketplace.createForm["title"])
	}
}

// TestContractCreation_NumericInput tests that reward and duration only accept numbers
func TestContractCreation_NumericInput(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{
		"contract_type": "0",
		"title":         "",
		"description":   "",
		"reward":        "",
		"target_name":   "",
		"duration":      "",
	}
	model.marketplace.formField = 3 // Reward field

	// Try to type letters (should be ignored)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, _ = model.updateMarketplaceCreateContract(msg)

	if model.marketplace.createForm["reward"] != "" {
		t.Errorf("Expected empty reward after letter input, got '%s'", model.marketplace.createForm["reward"])
	}

	// Type numbers (should be accepted)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	_, _ = model.updateMarketplaceCreateContract(msg)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}
	_, _ = model.updateMarketplaceCreateContract(msg)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}
	_, _ = model.updateMarketplaceCreateContract(msg)

	if model.marketplace.createForm["reward"] != "500" {
		t.Errorf("Expected reward '500', got '%s'", model.marketplace.createForm["reward"])
	}
}

// TestBountyPosting_FormInitialization tests that bounty form initializes correctly
func TestBountyPosting_FormInitialization(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModePostBounty

	// Simulate key press to trigger form initialization
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, _ = model.updateMarketplacePostBounty(msg)

	// Check form was initialized
	if _, exists := model.marketplace.createForm["target_name"]; !exists {
		t.Error("Target name field not initialized")
	}

	if model.marketplace.createForm["amount"] != "5000" {
		t.Errorf("Expected default amount '5000', got '%s'", model.marketplace.createForm["amount"])
	}

	if _, exists := model.marketplace.createForm["reason"]; !exists {
		t.Error("Reason field not initialized")
	}
}

// TestBountyPosting_TabNavigation tests tab navigation through bounty form fields
func TestBountyPosting_TabNavigation(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModePostBounty
	model.marketplace.createForm = map[string]string{
		"target_name": "",
		"amount":      "5000",
		"reason":      "",
	}
	model.marketplace.formField = 0

	// Press Tab 3 times (should cycle through all 3 fields and wrap to 0)
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	for i := 0; i < 3; i++ {
		_, _ = model.updateMarketplacePostBounty(tabMsg)
	}

	// Should wrap back to field 0
	if model.marketplace.formField != 0 {
		t.Errorf("Expected formField 0 after 3 tabs, got %d", model.marketplace.formField)
	}
}

// TestBountyPosting_TextInput tests text input in bounty form
func TestBountyPosting_TextInput(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModePostBounty
	model.marketplace.createForm = map[string]string{
		"target_name": "",
		"amount":      "5000",
		"reason":      "",
	}
	model.marketplace.formField = 0 // Target name field

	// Type "Pirate" into target name field
	chars := []rune{'P', 'i', 'r', 'a', 't', 'e'}
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		_, _ = model.updateMarketplacePostBounty(msg)
	}

	if model.marketplace.createForm["target_name"] != "Pirate" {
		t.Errorf("Expected target_name 'Pirate', got '%s'", model.marketplace.createForm["target_name"])
	}
}

// TestBountyPosting_NumericInput tests that amount only accepts numbers
func TestBountyPosting_NumericInput(t *testing.T) {
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModePostBounty
	model.marketplace.createForm = map[string]string{
		"target_name": "",
		"amount":      "",
		"reason":      "",
	}
	model.marketplace.formField = 1 // Amount field

	// Try to type letters (should be ignored)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, _ = model.updateMarketplacePostBounty(msg)

	if model.marketplace.createForm["amount"] != "" {
		t.Errorf("Expected empty amount after letter input, got '%s'", model.marketplace.createForm["amount"])
	}

	// Type numbers (should be accepted)
	digits := []rune{'1', '0', '0', '0', '0'}
	for _, digit := range digits {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{digit}}
		_, _ = model.updateMarketplacePostBounty(msg)
	}

	if model.marketplace.createForm["amount"] != "10000" {
		t.Errorf("Expected amount '10000', got '%s'", model.marketplace.createForm["amount"])
	}
}

// TestBountyPosting_FeeCalculation tests the 10% fee calculation logic
func TestBountyPosting_FeeCalculation(t *testing.T) {
	testCases := []struct {
		amount      int64
		expectedFee int64
	}{
		{5000, 500},
		{10000, 1000},
		{25000, 2500},
		{100000, 10000},
	}

	for _, tc := range testCases {
		// Calculate fee (same logic as in postBounty)
		fee := int64(float64(tc.amount) * 0.10)

		if fee != tc.expectedFee {
			t.Errorf("For amount %d, expected fee %d, got %d", tc.amount, tc.expectedFee, fee)
		}

		// Total cost should be amount + fee
		totalCost := tc.amount + fee
		expectedTotal := tc.amount + tc.expectedFee

		if totalCost != expectedTotal {
			t.Errorf("For amount %d, expected total %d, got %d", tc.amount, expectedTotal, totalCost)
		}
	}
}

// TestForm_CancelWithEscape tests that ESC cancels and returns to menu
func TestForm_CancelWithEscape(t *testing.T) {
	// Test contract form
	model := &Model{
		marketplace: newMarketplaceState(),
	}

	model.marketplace.mode = marketplaceModeCreateContract
	model.marketplace.createForm = map[string]string{"title": "Test"}

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _ = model.updateMarketplaceCreateContract(escMsg)

	if model.marketplace.mode != marketplaceModeMenu {
		t.Errorf("Expected mode %s after ESC, got %s", marketplaceModeMenu, model.marketplace.mode)
	}

	if len(model.marketplace.createForm) != 0 {
		t.Error("Expected form to be cleared after ESC")
	}

	// Test bounty form
	model.marketplace.mode = marketplaceModePostBounty
	model.marketplace.createForm = map[string]string{"target_name": "Test"}

	_, _ = model.updateMarketplacePostBounty(escMsg)

	if model.marketplace.mode != marketplaceModeMenu {
		t.Errorf("Expected mode %s after ESC, got %s", marketplaceModeMenu, model.marketplace.mode)
	}

	if len(model.marketplace.createForm) != 0 {
		t.Error("Expected form to be cleared after ESC")
	}
}
