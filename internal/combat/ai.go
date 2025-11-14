// File: internal/combat/ai.go
// Project: Terminal Velocity
// Description: Combat system: ai
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package combat

import (
	"math/rand"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// AILevel represents the difficulty/capability of an AI

type AILevel int

const (
	AILevelEasy AILevel = iota
	AILevelMedium
	AILevelHard
	AILevelExpert
	AILevelAce
)

// AIState tracks the runtime state of an AI combatant
type AIState struct {
	Level           AILevel
	Aggression      float64 // 0.0-1.0, how aggressive the AI is
	Accuracy        float64 // 0.0-1.0, accuracy modifier
	ReactionTime    float64 // seconds, delay before reacting
	CurrentTarget   string  // Ship ID of current target
	LastTargetCheck float64 // time since last target evaluation
	IsRetreating    bool
	FormationPos    *Position // Position in formation, if any
	Morale          float64   // 0.0-1.0, affects retreat decision
}

// Position represents a 2D position in space
type Position struct {
	X int
	Y int
}

// AIAction represents an action the AI wants to take
type AIAction struct {
	Type     string    // "fire", "move", "evade", "retreat", "target"
	TargetID string    // Target ship ID
	WeaponID string    // Weapon to fire
	Position *Position // Position to move to
	Priority float64   // Action priority (0.0-1.0)
}

// NewAIState creates a new AI state with default values for the given level
func NewAIState(level AILevel) *AIState {
	state := &AIState{
		Level:           level,
		LastTargetCheck: 0,
		IsRetreating:    false,
		Morale:          1.0,
	}

	// Set level-specific attributes
	switch level {
	case AILevelEasy:
		state.Aggression = 0.3
		state.Accuracy = 0.7
		state.ReactionTime = 2.0
	case AILevelMedium:
		state.Aggression = 0.5
		state.Accuracy = 0.85
		state.ReactionTime = 1.5
	case AILevelHard:
		state.Aggression = 0.7
		state.Accuracy = 0.95
		state.ReactionTime = 1.0
	case AILevelExpert:
		state.Aggression = 0.85
		state.Accuracy = 1.0
		state.ReactionTime = 0.5
	case AILevelAce:
		state.Aggression = 1.0
		state.Accuracy = 1.1 // Can exceed 100% for bonuses
		state.ReactionTime = 0.25
	}

	return state
}

// DecideAction determines what action the AI should take this turn
func DecideAction(
	ai *AIState,
	self *models.Ship,
	selfType *models.ShipType,
	enemies []*models.Ship,
	enemyTypes map[string]*models.ShipType,
	allies []*models.Ship,
	deltaTime float64,
) []AIAction {

	actions := []AIAction{}

	// Update morale based on ship condition
	ai.updateMorale(self, selfType)

	// Check if should retreat
	if ai.shouldRetreat(self, selfType, enemies) {
		ai.IsRetreating = true
		actions = append(actions, AIAction{
			Type:     "retreat",
			Priority: 1.0,
		})
		return actions
	}

	// Target selection
	ai.LastTargetCheck += deltaTime
	if ai.CurrentTarget == "" || ai.LastTargetCheck > 3.0 {
		target := ai.selectTarget(self, enemies, enemyTypes)
		if target != nil {
			ai.CurrentTarget = target.ID.String()
			ai.LastTargetCheck = 0
			actions = append(actions, AIAction{
				Type:     "target",
				TargetID: target.ID.String(),
				Priority: 0.8,
			})
		}
	}

	// Find current target
	var currentTarget *models.Ship
	var currentTargetType *models.ShipType
	for _, enemy := range enemies {
		if enemy.ID.String() == ai.CurrentTarget {
			currentTarget = enemy
			currentTargetType = enemyTypes[enemy.TypeID]
			break
		}
	}

	if currentTarget == nil {
		return actions
	}

	// Weapon usage - try to fire available weapons
	weaponActions := ai.selectWeapons(self, currentTarget, selfType, currentTargetType)
	actions = append(actions, weaponActions...)

	// Movement/evasion
	if ai.shouldEvade(self, selfType, currentTarget) {
		evasionAction := ai.calculateEvasion(self, currentTarget, selfType)
		if evasionAction != nil {
			actions = append(actions, *evasionAction)
		}
	}

	// Formation maintenance (if in formation)
	if ai.FormationPos != nil && len(allies) > 0 {
		formationAction := ai.maintainFormation(self, allies)
		if formationAction != nil {
			actions = append(actions, *formationAction)
		}
	}

	return actions
}

// selectTarget chooses the best target based on AI level and tactics
func (ai *AIState) selectTarget(self *models.Ship, enemies []*models.Ship,
	enemyTypes map[string]*models.ShipType) *models.Ship {

	if len(enemies) == 0 {
		return nil
	}

	var bestTarget *models.Ship
	bestScore := -1.0

	for _, enemy := range enemies {
		if enemy.Hull <= 0 {
			continue // Skip destroyed ships
		}

		enemyType := enemyTypes[enemy.TypeID]
		if enemyType == nil {
			continue
		}

		score := ai.calculateTargetScore(self, enemy, enemyType)
		if score > bestScore {
			bestScore = score
			bestTarget = enemy
		}
	}

	return bestTarget
}

// calculateTargetScore evaluates how good a target is
func (ai *AIState) calculateTargetScore(self *models.Ship, target *models.Ship,
	targetType *models.ShipType) float64 {

	score := 0.0

	// Prefer weakened targets (easier kills)
	hullPercent := float64(target.Hull) / float64(targetType.MaxHull)
	score += (1.0 - hullPercent) * 30.0

	// Prefer targets with low shields
	shieldPercent := 0.0
	if targetType.MaxShields > 0 {
		shieldPercent = float64(target.Shields) / float64(targetType.MaxShields)
	}
	score += (1.0 - shieldPercent) * 20.0

	// Consider threat level (higher threat = higher priority)
	// Bigger ships are more dangerous
	threatLevel := float64(len(target.Weapons)) * 10.0
	threatLevel += float64(targetType.MaxHull) / 100.0
	score += threatLevel * ai.Aggression

	// Distance factor (TODO: when we have positions)
	// Closer targets are preferred
	// distance := calculateDistance(self, target)
	// score += (1000.0 - float64(distance)) / 10.0

	// Random factor to prevent too predictable behavior
	if ai.Level <= AILevelMedium {
		score += rand.Float64() * 15.0
	} else {
		score += rand.Float64() * 5.0
	}

	return score
}

// selectWeapons determines which weapons to fire at the target
func (ai *AIState) selectWeapons(self *models.Ship, target *models.Ship,
	selfType *models.ShipType, targetType *models.ShipType) []AIAction {

	actions := []AIAction{}

	// Distance to target (placeholder - would be calculated from positions)
	distance := 500 // Medium range assumption

	for _, weaponID := range self.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon == nil {
			continue
		}

		// Check if weapon is in range
		if distance > weapon.RangeValue*2 {
			continue // Too far
		}

		// Decide whether to fire based on AI level and weapon suitability
		shouldFire := ai.shouldFireWeapon(weapon, target, targetType, distance)
		if shouldFire {
			priority := ai.calculateWeaponPriority(weapon, target, targetType, distance)
			actions = append(actions, AIAction{
				Type:     "fire",
				WeaponID: weapon.ID,
				TargetID: target.ID.String(),
				Priority: priority,
			})
		}
	}

	return actions
}

// shouldFireWeapon determines if the AI should fire a specific weapon
func (ai *AIState) shouldFireWeapon(weapon *models.Weapon, target *models.Ship,
	targetType *models.ShipType, distance int) bool {

	// Always try to fire if in optimal range
	if distance <= weapon.RangeValue {
		return true
	}

	// At medium range, consider accuracy and AI level
	if distance <= weapon.RangeValue*2 {
		// Higher level AIs are more conservative with accuracy
		if ai.Level >= AILevelHard {
			// Calculate expected hit chance
			hitChance := CalculateHitChance(weapon, targetType, targetType, distance)
			return hitChance > 40.0 // Only fire if >40% hit chance
		}
		return true // Lower level AIs fire more freely
	}

	return false
}

// calculateWeaponPriority determines priority for firing a weapon
func (ai *AIState) calculateWeaponPriority(weapon *models.Weapon, target *models.Ship,
	targetType *models.ShipType, distance int) float64 {

	priority := 0.5

	// Prefer high-damage weapons
	priority += float64(weapon.Damage) / 200.0

	// Prefer weapons in optimal range
	if distance <= weapon.RangeValue {
		priority += 0.3
	}

	// Consider shield penetration vs target's shield status
	if target.Shields > targetType.MaxShields/2 {
		// Target has strong shields, prefer penetrating weapons
		priority += weapon.ShieldPenetration * 0.2
	} else {
		// Target has weak shields, any weapon is good
		priority += 0.1
	}

	// Missile weapons are valuable, use strategically
	if weapon.Type == "missile" && ai.Level >= AILevelMedium {
		// Save missiles for weakened targets
		hullPercent := float64(target.Hull) / float64(targetType.MaxHull)
		if hullPercent < 0.5 {
			priority += 0.2 // Finish them off
		} else {
			priority -= 0.1 // Save ammo
		}
	}

	// Clamp priority
	if priority > 1.0 {
		priority = 1.0
	}
	if priority < 0.0 {
		priority = 0.0
	}

	return priority
}

// shouldEvade determines if the AI should take evasive action
func (ai *AIState) shouldEvade(self *models.Ship, selfType *models.ShipType,
	target *models.Ship) bool {

	// Low health means high evasion priority
	hullPercent := float64(self.Hull) / float64(selfType.MaxHull)
	if hullPercent < 0.3 {
		return true
	}

	// Low shields
	shieldPercent := 0.0
	if selfType.MaxShields > 0 {
		shieldPercent = float64(self.Shields) / float64(selfType.MaxShields)
	}
	if shieldPercent < 0.2 {
		return true
	}

	// Random evasion based on AI level
	// Higher level AIs evade more tactically
	if ai.Level >= AILevelHard {
		return rand.Float64() < 0.3
	} else if ai.Level >= AILevelMedium {
		return rand.Float64() < 0.2
	}

	return rand.Float64() < 0.1
}

// calculateEvasion determines the best evasion maneuver
func (ai *AIState) calculateEvasion(self *models.Ship, target *models.Ship,
	selfType *models.ShipType) *AIAction {

	// Simplified evasion - would be more complex with actual positioning
	// For now, just signal to move away from threat
	return &AIAction{
		Type:     "evade",
		Priority: 0.6,
	}
}

// shouldRetreat determines if the AI should retreat from combat
func (ai *AIState) shouldRetreat(self *models.Ship, selfType *models.ShipType,
	enemies []*models.Ship) bool {

	// Already retreating
	if ai.IsRetreating {
		return true
	}

	// Check morale
	if ai.Morale < 0.3 {
		return true
	}

	// Critical hull damage
	hullPercent := float64(self.Hull) / float64(selfType.MaxHull)
	if hullPercent < 0.2 {
		return true
	}

	// Outnumbered and damaged
	if len(enemies) > 3 && hullPercent < 0.5 {
		return true
	}

	// AI level affects retreat threshold
	switch ai.Level {
	case AILevelEasy:
		// Easy AI retreats readily
		return hullPercent < 0.4 && len(enemies) >= 2
	case AILevelMedium:
		return hullPercent < 0.3
	case AILevelHard:
		return hullPercent < 0.25
	case AILevelExpert, AILevelAce:
		// Expert/Ace fight to the end unless critical
		return hullPercent < 0.15
	}

	return false
}

// updateMorale updates AI morale based on combat situation
func (ai *AIState) updateMorale(self *models.Ship, selfType *models.ShipType) {
	// Hull damage reduces morale
	hullPercent := float64(self.Hull) / float64(selfType.MaxHull)
	targetMorale := hullPercent

	// Gradually adjust morale toward target
	if ai.Morale > targetMorale {
		ai.Morale -= 0.05
		if ai.Morale < targetMorale {
			ai.Morale = targetMorale
		}
	} else if ai.Morale < targetMorale {
		ai.Morale += 0.02 // Morale recovers slowly
		if ai.Morale > targetMorale {
			ai.Morale = targetMorale
		}
	}

	// AI level affects morale stability
	if ai.Level >= AILevelHard {
		// Higher level AIs have more stable morale
		if ai.Morale < 0.3 {
			ai.Morale = 0.3
		}
	}
}

// maintainFormation calculates movement to maintain formation with allies
func (ai *AIState) maintainFormation(self *models.Ship, allies []*models.Ship) *AIAction {
	if ai.FormationPos == nil || len(allies) == 0 {
		return nil
	}

	// Formation leader is the first ally (unused for now, will be used for position calculation)
	// leader := allies[0]

	// Calculate desired position relative to leader
	// This is simplified - real formation would use actual positions
	desiredPos := &Position{
		X: ai.FormationPos.X,
		Y: ai.FormationPos.Y,
	}

	return &AIAction{
		Type:     "move",
		Position: desiredPos,
		Priority: 0.4,
	}
}

// SetFormationPosition sets the AI's position in a formation
func (ai *AIState) SetFormationPosition(x, y int) {
	ai.FormationPos = &Position{X: x, Y: y}
}

// ClearFormation removes the AI from formation
func (ai *AIState) ClearFormation() {
	ai.FormationPos = nil
}

// GetAILevelName returns the human-readable name of an AI level
func GetAILevelName(level AILevel) string {
	switch level {
	case AILevelEasy:
		return "Easy"
	case AILevelMedium:
		return "Medium"
	case AILevelHard:
		return "Hard"
	case AILevelExpert:
		return "Expert"
	case AILevelAce:
		return "Ace"
	default:
		return "Unknown"
	}
}

// ApplyAIAccuracyModifier applies the AI's accuracy modifier to a weapon
func ApplyAIAccuracyModifier(ai *AIState, baseAccuracy float64) float64 {
	return baseAccuracy * ai.Accuracy
}
