// File: internal/models/territory.go
// Project: Terminal Velocity
// Description: Territory control models for faction-owned systems
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TerritoryControlLevel represents the strength of control over a system
type TerritoryControlLevel string

const (
	ControlLevelContested TerritoryControlLevel = "contested" // Multiple factions vying for control
	ControlLevelWeak      TerritoryControlLevel = "weak"      // Recently claimed, low control
	ControlLevelStable    TerritoryControlLevel = "stable"    // Established control
	ControlLevelStrong    TerritoryControlLevel = "strong"    // Well-defended, high influence
	ControlLevelDominant  TerritoryControlLevel = "dominant"  // Maximum control and benefits
)

// TerritoryBenefit represents bonuses from controlling territory
type TerritoryBenefit struct {
	TradeBonus      float64 `json:"trade_bonus"`       // % bonus to trade profits
	ProductionBonus float64 `json:"production_bonus"`  // % bonus to production
	DefenseBonus    int     `json:"defense_bonus"`     // Defense rating boost
	IncomeBonus     int64   `json:"income_bonus"`      // Passive credit income per day
}

// Territory represents a faction's claim on a star system
type Territory struct {
	ID        uuid.UUID `json:"id"`         // Unique territory record
	SystemID  uuid.UUID `json:"system_id"`  // Which system is claimed
	SystemName string   `json:"system_name"` // System name for display
	FactionID uuid.UUID `json:"faction_id"` // Owning faction
	FactionTag string   `json:"faction_tag"` // Faction tag for display

	// Control metrics
	ControlLevel  TerritoryControlLevel `json:"control_level"`  // Current control strength
	ControlPoints int                   `json:"control_points"` // Points toward next level
	
	// Timing
	ClaimedAt  time.Time  `json:"claimed_at"`  // When the system was claimed
	LastUpkeep time.Time  `json:"last_upkeep"` // Last upkeep payment
	NextUpkeep time.Time  `json:"next_upkeep"` // When next payment is due

	// Costs and income
	UpkeepCost int64 `json:"upkeep_cost"` // Weekly upkeep cost
	Income     int64 `json:"income"`      // Weekly passive income

	// Infrastructure
	DefenseLevel   int  `json:"defense_level"`    // 0-5, affects defense strength
	DevelopmentLevel int `json:"development_level"` // 0-5, affects benefits
	HasStation     bool `json:"has_station"`      // Has faction station

	// Activity tracking
	MemberActivity   int       `json:"member_activity"`    // Member visits this week
	TradeVolume      int64     `json:"trade_volume"`       // Credits traded this week
	LastConflict     *time.Time `json:"last_conflict,omitempty"` // Last contested/attacked
}

// NewTerritory creates a new territory claim
func NewTerritory(systemID uuid.UUID, systemName string, factionID uuid.UUID, factionTag string) *Territory {
	now := time.Now()
	
	return &Territory{
		ID:               uuid.New(),
		SystemID:         systemID,
		SystemName:       systemName,
		FactionID:        factionID,
		FactionTag:       factionTag,
		ControlLevel:     ControlLevelWeak,
		ControlPoints:    0,
		ClaimedAt:        now,
		LastUpkeep:       now,
		NextUpkeep:       now.Add(7 * 24 * time.Hour), // Weekly upkeep
		UpkeepCost:       1000, // Base cost
		Income:           500,  // Base income
		DefenseLevel:     0,
		DevelopmentLevel: 0,
		HasStation:       false,
		MemberActivity:   0,
		TradeVolume:      0,
	}
}

// GetControlLevelName returns a display name for the control level
func (t TerritoryControlLevel) GetDisplayName() string {
	names := map[TerritoryControlLevel]string{
		ControlLevelContested: "Contested",
		ControlLevelWeak:      "Weak",
		ControlLevelStable:    "Stable",
		ControlLevelStrong:    "Strong",
		ControlLevelDominant:  "Dominant",
	}
	
	if name, exists := names[t]; exists {
		return name
	}
	return string(t)
}

// GetControlLevelColor returns a color indicator for the control level
func (t TerritoryControlLevel) GetColorIndicator() string {
	colors := map[TerritoryControlLevel]string{
		ControlLevelContested: "🔴", // Red
		ControlLevelWeak:      "🟠", // Orange
		ControlLevelStable:    "🟡", // Yellow
		ControlLevelStrong:    "🟢", // Green
		ControlLevelDominant:  "🔵", // Blue
	}
	
	if color, exists := colors[t]; exists {
		return color
	}
	return "⚪"
}

// CalculateUpkeep calculates the weekly upkeep cost
func (t *Territory) CalculateUpkeep() int64 {
	baseCost := int64(1000)
	
	// Cost scales with development
	developmentMultiplier := float64(1 + t.DevelopmentLevel)
	
	// Defense increases cost
	defenseCost := int64(t.DefenseLevel * 500)
	
	// Station adds significant cost
	stationCost := int64(0)
	if t.HasStation {
		stationCost = 5000
	}
	
	total := int64(float64(baseCost) * developmentMultiplier) + defenseCost + stationCost
	t.UpkeepCost = total
	return total
}

// CalculateIncome calculates the weekly passive income
func (t *Territory) CalculateIncome() int64 {
	baseIncome := int64(500)
	
	// Income scales with development
	developmentBonus := int64(t.DevelopmentLevel * 1000)
	
	// Trade volume bonus (1% of weekly trade)
	tradeBonus := t.TradeVolume / 100
	
	// Station provides income
	stationBonus := int64(0)
	if t.HasStation {
		stationBonus = 3000
	}
	
	// Control level multiplier
	controlMultiplier := t.getControlMultiplier()
	
	total := int64(float64(baseIncome+developmentBonus+stationBonus) * controlMultiplier) + tradeBonus
	t.Income = total
	return total
}

// getControlMultiplier returns an income multiplier based on control level
func (t *Territory) getControlMultiplier() float64 {
	multipliers := map[TerritoryControlLevel]float64{
		ControlLevelContested: 0.5,
		ControlLevelWeak:      0.75,
		ControlLevelStable:    1.0,
		ControlLevelStrong:    1.25,
		ControlLevelDominant:  1.5,
	}
	
	if mult, exists := multipliers[t.ControlLevel]; exists {
		return mult
	}
	return 1.0
}

// GetBenefits calculates the current benefits provided by this territory
func (t *Territory) GetBenefits() TerritoryBenefit {
	benefits := TerritoryBenefit{}
	
	// Base benefits from development level
	benefits.TradeBonus = float64(t.DevelopmentLevel) * 0.05      // 5% per level
	benefits.ProductionBonus = float64(t.DevelopmentLevel) * 0.03  // 3% per level
	
	// Defense benefits
	benefits.DefenseBonus = t.DefenseLevel * 10
	
	// Passive income
	benefits.IncomeBonus = t.Income
	
	// Control level enhances benefits
	controlBonus := t.getControlMultiplier()
	benefits.TradeBonus *= controlBonus
	benefits.ProductionBonus *= controlBonus
	
	return benefits
}

// IsUpkeepDue checks if upkeep payment is due
func (t *Territory) IsUpkeepDue() bool {
	return time.Now().After(t.NextUpkeep)
}

// GetUpkeepStatus returns a human-readable upkeep status
func (t *Territory) GetUpkeepStatus() string {
	if t.IsUpkeepDue() {
		return "⚠️  OVERDUE"
	}
	
	timeUntil := time.Until(t.NextUpkeep)
	
	if timeUntil < 24*time.Hour {
		hours := int(timeUntil.Hours())
		return fmt.Sprintf("Due in %d hours", hours)
	}
	
	days := int(timeUntil.Hours() / 24)
	return fmt.Sprintf("Due in %d days", days)
}

// PayUpkeep processes an upkeep payment
func (t *Territory) PayUpkeep() {
	t.LastUpkeep = time.Now()
	t.NextUpkeep = time.Now().Add(7 * 24 * time.Hour)
	
	// Add control points for maintaining territory
	t.AddControlPoints(10)
}

// AddControlPoints adds control points and handles level progression
func (t *Territory) AddControlPoints(points int) {
	t.ControlPoints += points
	
	// Check for level up
	requiredPoints := t.getRequiredControlPoints()
	if t.ControlPoints >= requiredPoints {
		t.levelUpControl()
	}
}

// getRequiredControlPoints returns points needed for next level
func (t *Territory) getRequiredControlPoints() int {
	requirements := map[TerritoryControlLevel]int{
		ControlLevelWeak:      100,
		ControlLevelStable:    250,
		ControlLevelStrong:    500,
		ControlLevelDominant:  1000,
	}
	
	if req, exists := requirements[t.ControlLevel]; exists {
		return req
	}
	return 999999 // Max level
}

// levelUpControl advances to the next control level
func (t *Territory) levelUpControl() {
	progression := map[TerritoryControlLevel]TerritoryControlLevel{
		ControlLevelContested: ControlLevelWeak,
		ControlLevelWeak:      ControlLevelStable,
		ControlLevelStable:    ControlLevelStrong,
		ControlLevelStrong:    ControlLevelDominant,
	}
	
	if nextLevel, exists := progression[t.ControlLevel]; exists {
		t.ControlLevel = nextLevel
		t.ControlPoints = 0 // Reset for next level
	}
}

// UpgradeDefense increases the defense level
func (t *Territory) UpgradeDefense() bool {
	if t.DefenseLevel >= 5 {
		return false
	}
	t.DefenseLevel++
	t.CalculateUpkeep() // Recalculate costs
	return true
}

// UpgradeDevelopment increases the development level
func (t *Territory) UpgradeDevelopment() bool {
	if t.DevelopmentLevel >= 5 {
		return false
	}
	t.DevelopmentLevel++
	t.CalculateUpkeep() // Recalculate costs
	t.CalculateIncome()  // Recalculate income
	return true
}

// BuildStation constructs a faction station
func (t *Territory) BuildStation() bool {
	if t.HasStation {
		return false
	}
	t.HasStation = true
	t.CalculateUpkeep()
	t.CalculateIncome()
	return true
}

// GetControlAge returns how long the territory has been held
func (t *Territory) GetControlAge() string {
	duration := time.Since(t.ClaimedAt)
	
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hours", hours)
	}
	
	days := int(duration.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%d days", days)
	}
	
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%d months", months)
	}
	
	years := months / 12
	return fmt.Sprintf("%d years", years)
}

// GetUpgradeCost returns the cost to upgrade defense or development
func GetDefenseUpgradeCost(currentLevel int) int64 {
	return int64((currentLevel + 1) * 5000)
}

func GetDevelopmentUpgradeCost(currentLevel int) int64 {
	return int64((currentLevel + 1) * 10000)
}

func GetStationBuildCost() int64 {
	return 100000
}
