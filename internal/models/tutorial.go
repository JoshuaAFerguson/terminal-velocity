// File: internal/models/tutorial.go
// Project: Terminal Velocity
// Description: Tutorial system models and progression
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// TutorialStep represents a single tutorial step

type TutorialStep struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Screen      string   `json:"screen"`    // Which screen this tutorial is for
	Objective   string   `json:"objective"` // What the player needs to do
	Hints       []string `json:"hints"`     // Progressive hints
	Completed   bool     `json:"completed"`
	OrderIndex  int      `json:"order_index"` // Order in the tutorial sequence
}

// TutorialCategory represents a group of related tutorial steps
type TutorialCategory string

const (
	TutorialBasics      TutorialCategory = "basics"      // Basic navigation and UI
	TutorialTrading     TutorialCategory = "trading"     // Trading mechanics
	TutorialCombat      TutorialCategory = "combat"      // Combat basics
	TutorialShips       TutorialCategory = "ships"       // Ship management
	TutorialMissions    TutorialCategory = "missions"    // Mission system
	TutorialMultiplayer TutorialCategory = "multiplayer" // Social features
	TutorialAdvanced    TutorialCategory = "advanced"    // Advanced mechanics
)

// TutorialProgress tracks a player's tutorial progress
type TutorialProgress struct {
	ID               uuid.UUID                `json:"id"`
	PlayerID         uuid.UUID                `json:"player_id"`
	CurrentStep      string                   `json:"current_step"`      // Current step ID
	CompletedSteps   map[string]time.Time     `json:"completed_steps"`   // stepID -> completion time
	SkippedSteps     map[string]time.Time     `json:"skipped_steps"`     // stepID -> skip time
	CategoryProgress map[TutorialCategory]int `json:"category_progress"` // category -> steps completed
	TutorialEnabled  bool                     `json:"tutorial_enabled"`  // Can be disabled by player
	StartedAt        time.Time                `json:"started_at"`
	LastUpdated      time.Time                `json:"last_updated"`
	TotalSteps       int                      `json:"total_steps"`
	CompletedCount   int                      `json:"completed_count"`
}

// TutorialHintLevel represents hint progression
type TutorialHintLevel int

const (
	HintNone   TutorialHintLevel = 0
	HintBasic  TutorialHintLevel = 1
	HintMedium TutorialHintLevel = 2
	HintFull   TutorialHintLevel = 3
)

// NewTutorialProgress creates a new tutorial progress tracker
func NewTutorialProgress(playerID uuid.UUID) *TutorialProgress {
	return &TutorialProgress{
		ID:               uuid.New(),
		PlayerID:         playerID,
		CurrentStep:      "",
		CompletedSteps:   make(map[string]time.Time),
		SkippedSteps:     make(map[string]time.Time),
		CategoryProgress: make(map[TutorialCategory]int),
		TutorialEnabled:  true,
		StartedAt:        time.Now(),
		LastUpdated:      time.Now(),
		TotalSteps:       0,
		CompletedCount:   0,
	}
}

// CompleteStep marks a tutorial step as completed
func (tp *TutorialProgress) CompleteStep(stepID string, category TutorialCategory) {
	if _, exists := tp.CompletedSteps[stepID]; !exists {
		tp.CompletedSteps[stepID] = time.Now()
		tp.CompletedCount++
		tp.CategoryProgress[category]++
		tp.LastUpdated = time.Now()
	}
}

// SkipStep marks a tutorial step as skipped
func (tp *TutorialProgress) SkipStep(stepID string) {
	if _, completed := tp.CompletedSteps[stepID]; !completed {
		if _, skipped := tp.SkippedSteps[stepID]; !skipped {
			tp.SkippedSteps[stepID] = time.Now()
			tp.LastUpdated = time.Now()
		}
	}
}

// IsStepCompleted checks if a step is completed
func (tp *TutorialProgress) IsStepCompleted(stepID string) bool {
	_, exists := tp.CompletedSteps[stepID]
	return exists
}

// IsStepSkipped checks if a step is skipped
func (tp *TutorialProgress) IsStepSkipped(stepID string) bool {
	_, exists := tp.SkippedSteps[stepID]
	return exists
}

// GetCompletionPercentage returns overall completion percentage
func (tp *TutorialProgress) GetCompletionPercentage() float64 {
	if tp.TotalSteps == 0 {
		return 0
	}
	return float64(tp.CompletedCount) / float64(tp.TotalSteps) * 100
}

// GetCategoryProgress returns completion percentage for a category
func (tp *TutorialProgress) GetCategoryProgress(category TutorialCategory, totalInCategory int) float64 {
	if totalInCategory == 0 {
		return 0
	}
	completed := tp.CategoryProgress[category]
	return float64(completed) / float64(totalInCategory) * 100
}

// Tutorial represents a complete tutorial sequence
type Tutorial struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	Category      TutorialCategory `json:"category"`
	Steps         []*TutorialStep  `json:"steps"`
	Prerequisites []string         `json:"prerequisites"` // Tutorial IDs that must be completed first
	IsOptional    bool             `json:"is_optional"`
	OrderIndex    int              `json:"order_index"`
}

// NewTutorial creates a new tutorial
func NewTutorial(id, title, description string, category TutorialCategory) *Tutorial {
	return &Tutorial{
		ID:            id,
		Title:         title,
		Description:   description,
		Category:      category,
		Steps:         make([]*TutorialStep, 0),
		Prerequisites: make([]string, 0),
		IsOptional:    false,
		OrderIndex:    0,
	}
}

// AddStep adds a step to the tutorial
func (t *Tutorial) AddStep(step *TutorialStep) {
	t.Steps = append(t.Steps, step)
}

// GetNextIncompleteStep returns the next step that hasn't been completed
func (t *Tutorial) GetNextIncompleteStep(progress *TutorialProgress) *TutorialStep {
	for _, step := range t.Steps {
		if !progress.IsStepCompleted(step.ID) && !progress.IsStepSkipped(step.ID) {
			return step
		}
	}
	return nil
}

// IsCompleted checks if all steps in the tutorial are completed
func (t *Tutorial) IsCompleted(progress *TutorialProgress) bool {
	for _, step := range t.Steps {
		if !progress.IsStepCompleted(step.ID) {
			return false
		}
	}
	return true
}

// GetProgress returns the number of completed steps and total steps
func (t *Tutorial) GetProgress(progress *TutorialProgress) (completed int, total int) {
	total = len(t.Steps)
	completed = 0
	for _, step := range t.Steps {
		if progress.IsStepCompleted(step.ID) {
			completed++
		}
	}
	return
}

// TutorialTrigger represents a condition that triggers a tutorial
type TutorialTrigger string

const (
	TriggerFirstLogin        TutorialTrigger = "first_login"
	TriggerFirstTrade        TutorialTrigger = "first_trade"
	TriggerFirstCombat       TutorialTrigger = "first_combat"
	TriggerFirstJump         TutorialTrigger = "first_jump"
	TriggerFirstShipPurchase TutorialTrigger = "first_ship_purchase"
	TriggerFirstMission      TutorialTrigger = "first_mission"
	TriggerFirstDeath        TutorialTrigger = "first_death"
	TriggerLowCredits        TutorialTrigger = "low_credits"
	TriggerHighCredits       TutorialTrigger = "high_credits"
	TriggerScreenEnter       TutorialTrigger = "screen_enter"
)

// TutorialEvent represents an event that can trigger tutorials
type TutorialEvent struct {
	Trigger   TutorialTrigger        `json:"trigger"`
	Context   map[string]interface{} `json:"context"` // Additional context data
	Timestamp time.Time              `json:"timestamp"`
}

// NewTutorialEvent creates a new tutorial event
func NewTutorialEvent(trigger TutorialTrigger) *TutorialEvent {
	return &TutorialEvent{
		Trigger:   trigger,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}
