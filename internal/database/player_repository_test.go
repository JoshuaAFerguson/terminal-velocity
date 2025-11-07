// File: internal/database/player_repository_test.go
// Project: Terminal Velocity
// Description: Database repository for player_repository_test
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// These are integration tests that require a running PostgreSQL database
// Skip them if DATABASE_URL is not set

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	cfg := &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "terminal_velocity",
		Password:        "terminal_velocity",
		Database:        "terminal_velocity_test",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := NewDB(cfg)
	if err != nil {
		t.Skipf("Skipping database tests: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping database tests: cannot connect to database: %v", err)
	}

	return db
}

func TestPlayerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPlayerRepository(db)
	ctx := context.Background()

	username := "testuser_" + uuid.New().String()[:8]
	password := "testpassword123"

	player, err := repo.Create(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	if player.ID == uuid.Nil {
		t.Error("Player ID should not be nil")
	}

	if player.Username != username {
		t.Errorf("Expected username %s, got %s", username, player.Username)
	}

	if player.Credits != 10000 {
		t.Errorf("Expected initial credits 10000, got %d", player.Credits)
	}

	// Cleanup
	_ = repo.Delete(ctx, player.ID)
}

func TestPlayerRepository_Authenticate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPlayerRepository(db)
	ctx := context.Background()

	username := "testuser_" + uuid.New().String()[:8]
	password := "testpassword123"

	// Create a player
	created, err := repo.Create(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}
	defer func() { _ = repo.Delete(ctx, created.ID) }()

	// Test successful authentication
	player, err := repo.Authenticate(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	if player.ID != created.ID {
		t.Errorf("Expected player ID %s, got %s", created.ID, player.ID)
	}

	// Test failed authentication (wrong password)
	_, err = repo.Authenticate(ctx, username, "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	// Test failed authentication (wrong username)
	_, err = repo.Authenticate(ctx, "nonexistent", password)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestPlayerRepository_ModifyCredits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPlayerRepository(db)
	ctx := context.Background()

	username := "testuser_" + uuid.New().String()[:8]
	password := "testpassword123"

	player, err := repo.Create(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}
	defer func() { _ = repo.Delete(ctx, player.ID) }()

	// Add credits
	err = repo.ModifyCredits(ctx, player.ID, 5000)
	if err != nil {
		t.Fatalf("Failed to add credits: %v", err)
	}

	// Verify credits
	updated, err := repo.GetByID(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updated.Credits != 15000 {
		t.Errorf("Expected credits 15000, got %d", updated.Credits)
	}

	// Subtract credits
	err = repo.ModifyCredits(ctx, player.ID, -10000)
	if err != nil {
		t.Fatalf("Failed to subtract credits: %v", err)
	}

	updated, err = repo.GetByID(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updated.Credits != 5000 {
		t.Errorf("Expected credits 5000, got %d", updated.Credits)
	}

	// Try to subtract more than available (should fail)
	err = repo.ModifyCredits(ctx, player.ID, -10000)
	if err == nil {
		t.Error("Expected error when subtracting more credits than available")
	}
}

func TestPlayerRepository_UpdateReputation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPlayerRepository(db)
	ctx := context.Background()

	username := "testuser_" + uuid.New().String()[:8]
	password := "testpassword123"

	player, err := repo.Create(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}
	defer func() { _ = repo.Delete(ctx, player.ID) }()

	// Update reputation
	err = repo.UpdateReputation(ctx, player.ID, "united_earth_federation", 10)
	if err != nil {
		t.Fatalf("Failed to update reputation: %v", err)
	}

	// Get player and verify reputation
	updated, err := repo.GetByID(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if rep, ok := updated.Reputation["united_earth_federation"]; !ok || rep != 10 {
		t.Errorf("Expected reputation 10, got %d", rep)
	}

	// Update again (should add to existing)
	err = repo.UpdateReputation(ctx, player.ID, "united_earth_federation", 5)
	if err != nil {
		t.Fatalf("Failed to update reputation: %v", err)
	}

	updated, err = repo.GetByID(ctx, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if rep, ok := updated.Reputation["united_earth_federation"]; !ok || rep != 15 {
		t.Errorf("Expected reputation 15, got %d", rep)
	}
}
