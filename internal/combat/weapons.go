// File: internal/combat/weapons.go
// Project: Terminal Velocity
// Description: Combat system: weapons - Weapon firing, hit chance calculation, damage application
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package combat provides the weapons system for the Terminal Velocity space combat game.
//
// This file implements weapon firing mechanics including:
//   - Hit chance calculation with distance, accuracy, and evasion factors
//   - Damage calculation with critical hits and shield penetration
//   - Weapon cooldown management
//   - Ammo tracking for missile weapons
//   - Energy cost tracking for energy weapons
//
// The weapons system supports multiple weapon types:
//   - Laser: Energy weapons with fast firing, no ammo, moderate energy cost
//   - Missile: Explosive weapons with high damage, limited ammo, good shield penetration
//   - Plasma: Balanced weapons with good damage and shield penetration
//   - Railgun: Kinetic weapons with very high damage and excellent shield penetration
//
// Damage is applied in two phases:
//   1. Shield damage (reduced by shield penetration stat)
//   2. Hull damage (direct damage bypassing shields based on penetration stat)
//
// Thread-safety: Functions in this file are stateless and safe for concurrent calls.
// WeaponState structs are not thread-safe and should be managed per-combat instance.
package combat

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// WeaponState tracks the runtime state of a weapon during combat.
//
// This structure maintains per-weapon state including cooldowns and ammo counts.
// Each weapon on a ship should have its own WeaponState instance during combat.
//
// Fields:
//   - WeaponID: Unique identifier of the weapon definition
//   - CurrentAmmo: Remaining ammo for missile weapons (unlimited for energy weapons)
//   - CooldownRemaining: Seconds until weapon can fire again (0 = ready)
//   - LastFiredTurn: Turn number when weapon was last fired (for turn-based combat)
type WeaponState struct {
	WeaponID          string
	CurrentAmmo       int     // current ammo (for missiles)
	CooldownRemaining float64 // seconds until weapon can fire again
	LastFiredTurn     int     // turn number when weapon was last fired
}

// FireResult contains the complete result of a weapon firing attempt.
//
// This structure provides detailed information about the outcome of a weapon shot,
// including hit/miss status, damage dealt, and appropriate combat log messages.
//
// Fields:
//   - Hit: Whether the shot hit the target
//   - Damage: Total damage dealt (base damage, includes critical hit multiplier)
//   - ShieldDamage: Damage absorbed by shields
//   - HullDamage: Damage applied to hull (penetrating or shield overflow)
//   - CriticalHit: Whether this was a critical hit (1.5x damage)
//   - AmmoRemaining: Remaining ammo after firing (for missile weapons)
//   - Message: Human-readable combat log message
type FireResult struct {
	Hit           bool
	Damage        int
	ShieldDamage  int
	HullDamage    int
	CriticalHit   bool
	AmmoRemaining int
	Message       string
}

// CanFire checks if a weapon is ready to fire based on cooldown and ammo availability.
//
// This function validates whether a weapon can be fired in the current turn by checking:
//   - Cooldown status (must be 0 to fire)
//   - Ammo availability (for missile weapons)
//   - Energy availability (for energy weapons, currently unlimited)
//
// Parameters:
//   - weapon: The weapon definition containing type, cooldown, and ammo capacity
//   - state: Current weapon state tracking cooldown and ammo
//   - ship: The ship firing the weapon (for future energy checks)
//   - shipType: Ship type definition (for future energy pool system)
//
// Returns:
//   - bool: true if weapon can fire, false otherwise
//   - string: Empty if can fire, error message explaining why not otherwise
//
// Thread-safe: No shared state modification, safe for concurrent calls.
func CanFire(weapon *models.Weapon, state *WeaponState, ship *models.Ship, shipType *models.ShipType) (bool, string) {
	// Check cooldown
	if state.CooldownRemaining > 0 {
		return false, fmt.Sprintf("Weapon cooling down (%.1fs remaining)", state.CooldownRemaining)
	}

	// Check ammo for missile weapons
	if weapon.Type == "missile" && weapon.AmmoCapacity > 0 {
		if state.CurrentAmmo <= 0 {
			return false, "Out of ammo"
		}
	}

	// Check energy for energy weapons
	if weapon.EnergyCost > 0 {
		// For now, assume ships have unlimited energy
		// This could be expanded later with an energy pool system
	}

	return true, ""
}

// Fire executes a complete weapon firing attempt including hit calculation and damage application.
//
// This is the main function for weapon combat, handling the full firing sequence:
//   1. Validate weapon can fire (cooldown and ammo)
//   2. Calculate hit chance based on accuracy, distance, and ship stats
//   3. Roll for hit/miss
//   4. If hit, calculate damage with critical hit chance (10%)
//   5. Apply damage to target (shields first, then hull)
//   6. Update weapon state (cooldown and ammo consumption)
//   7. Generate combat log message
//
// Damage Mechanics:
//   - Base damage from weapon definition
//   - Critical hits deal 1.5x damage (10% chance)
//   - Shield penetration determines direct hull damage vs shield damage
//   - Shields absorb damage first, overflow goes to hull
//   - All damage is applied directly to target ship state
//
// Hit Chance Factors:
//   - Weapon base accuracy
//   - Distance penalty (1% per 100 units beyond optimal range)
//   - Target evasion bonus (based on maneuverability)
//   - Attacker accuracy bonus (based on precision)
//   - Clamped between 5% and 95%
//
// Parameters:
//   - weapon: Weapon definition (damage, accuracy, range, type, etc.)
//   - state: Weapon state (cooldown, ammo) - modified by this function
//   - attacker: Ship firing the weapon
//   - target: Ship being fired upon - modified by this function (damage applied)
//   - attackerType: Ship type for attacker (for accuracy calculations)
//   - targetType: Ship type for target (for evasion and max HP)
//   - distance: Distance between ships in game units
//
// Returns:
//   - FireResult with complete combat outcome including damage breakdown
//
// Side Effects:
//   - Modifies target.Shields and target.Hull (applies damage)
//   - Modifies state.CooldownRemaining and state.CurrentAmmo
//   - Modifies state.LastFiredTurn
//
// Thread-safe: Modifies only passed parameters, safe for single combat instance.
func Fire(weapon *models.Weapon, state *WeaponState, attacker *models.Ship, target *models.Ship,
	attackerType *models.ShipType, targetType *models.ShipType, distance int) FireResult {

	result := FireResult{}

	// Check if can fire
	canFire, msg := CanFire(weapon, state, attacker, attackerType)
	if !canFire {
		result.Message = msg
		return result
	}

	// Calculate hit chance
	hitChance := CalculateHitChance(weapon, attackerType, targetType, distance)
	roll := rand.Float64() * 100

	if roll > hitChance {
		// Miss
		result.Hit = false
		result.Message = fmt.Sprintf("%s missed! (%.1f%% chance, rolled %.1f)", weapon.Name, hitChance, roll)
	} else {
		// Hit!
		result.Hit = true

		// Calculate damage
		baseDamage := weapon.Damage

		// Critical hit check (10% chance)
		if rand.Float64() < 0.1 {
			result.CriticalHit = true
			baseDamage = int(float64(baseDamage) * 1.5)
		}

		// Apply damage to shields first, then hull
		shieldPenetration := weapon.ShieldPenetration
		directDamage := int(float64(baseDamage) * shieldPenetration)
		shieldDamage := baseDamage - directDamage

		// Apply to target shields
		if target.Shields > 0 {
			if shieldDamage >= target.Shields {
				// Shield broken, remaining damage to hull
				overflow := shieldDamage - target.Shields
				result.ShieldDamage = target.Shields
				result.HullDamage = overflow + directDamage
				target.Shields = 0
			} else {
				// Shields absorb damage
				result.ShieldDamage = shieldDamage
				result.HullDamage = directDamage
				target.Shields -= shieldDamage
			}
		} else {
			// No shields, all damage to hull
			result.HullDamage = baseDamage
		}

		// Apply hull damage
		if result.HullDamage > 0 {
			target.Hull -= result.HullDamage
			if target.Hull < 0 {
				target.Hull = 0
			}
		}

		result.Damage = baseDamage

		// Build message
		if result.CriticalHit {
			result.Message = fmt.Sprintf("CRITICAL HIT! %s dealt %d damage", weapon.Name, baseDamage)
		} else {
			result.Message = fmt.Sprintf("%s hit for %d damage", weapon.Name, baseDamage)
		}

		if result.ShieldDamage > 0 && result.HullDamage > 0 {
			result.Message += fmt.Sprintf(" (%d to shields, %d to hull)", result.ShieldDamage, result.HullDamage)
		} else if result.ShieldDamage > 0 {
			result.Message += fmt.Sprintf(" (shields absorbed %d)", result.ShieldDamage)
		} else {
			result.Message += fmt.Sprintf(" (hull damage: %d)", result.HullDamage)
		}
	}

	// Update weapon state
	state.CooldownRemaining = weapon.Cooldown
	state.LastFiredTurn++

	// Consume ammo if missile
	if weapon.Type == "missile" && weapon.AmmoCapacity > 0 {
		state.CurrentAmmo -= weapon.AmmoConsumption
		if state.CurrentAmmo < 0 {
			state.CurrentAmmo = 0
		}
		result.AmmoRemaining = state.CurrentAmmo
	}

	return result
}

// CalculateHitChance calculates the probability of a weapon shot hitting its target.
//
// Hit chance is determined by multiple factors:
//   - Base weapon accuracy (75-95% typically)
//   - Distance penalty: 1% reduction per 100 units beyond optimal range
//   - Target evasion: Higher maneuverability makes ships harder to hit
//   - Attacker precision: Attacker maneuverability improves accuracy slightly
//
// Formula:
//
//	hitChance = baseAccuracy - rangePenalty - evasionBonus + attackerBonus
//	where:
//	  rangePenalty = (distance - weaponRange) / 100 (if beyond optimal range)
//	  evasionBonus = targetManeuverability * 2.0
//	  attackerBonus = attackerManeuverability * 0.5
//
// The final hit chance is clamped between 5% (always a small chance to hit) and
// 95% (always a small chance to miss for balance).
//
// Parameters:
//   - weapon: Weapon being fired (for base accuracy and optimal range)
//   - attackerType: Ship type of attacker (for maneuverability bonus)
//   - targetType: Ship type of target (for evasion calculation)
//   - distance: Distance between ships in game units
//
// Returns:
//   - float64: Hit chance percentage (5.0 to 95.0)
//
// Thread-safe: No shared state, safe for concurrent calls.
func CalculateHitChance(weapon *models.Weapon, attackerType *models.ShipType,
	targetType *models.ShipType, distance int) float64 {

	// Base accuracy from weapon
	baseAccuracy := float64(weapon.Accuracy)

	// Distance penalty
	rangePenalty := 0.0
	if distance > weapon.RangeValue {
		// Out of optimal range
		excessDistance := distance - weapon.RangeValue
		rangePenalty = float64(excessDistance) / 100.0 // 1% per 100 units over range
	}

	// Target evasion bonus (based on maneuverability)
	evasionBonus := float64(targetType.Maneuverability) * 2.0

	// Attacker accuracy bonus (based on speed/precision)
	attackerBonus := float64(attackerType.Maneuverability) * 0.5

	// Final hit chance calculation
	hitChance := baseAccuracy - rangePenalty - evasionBonus + attackerBonus

	// Clamp between 5% and 95%
	if hitChance < 5.0 {
		hitChance = 5.0
	}
	if hitChance > 95.0 {
		hitChance = 95.0
	}

	return hitChance
}

// UpdateCooldowns decrements all weapon cooldowns based on elapsed time.
//
// This function should be called each turn or frame to advance weapon cooldown timers.
// Cooldowns are reduced by deltaTime and clamped at 0 (ready to fire).
//
// Parameters:
//   - states: Slice of all weapon states to update
//   - deltaTime: Time elapsed since last update (in seconds)
//
// Side Effects:
//   - Modifies CooldownRemaining for each weapon state
//
// Thread-safe: Modifies only passed parameters, safe for single combat instance.
func UpdateCooldowns(states []*WeaponState, deltaTime float64) {
	for _, state := range states {
		if state.CooldownRemaining > 0 {
			state.CooldownRemaining -= deltaTime
			if state.CooldownRemaining < 0 {
				state.CooldownRemaining = 0
			}
		}
	}
}

// InitializeWeaponState creates a new weapon state with default values for combat.
//
// This function should be called when initializing a ship for combat to set up
// weapon state tracking. All weapons start with full ammo and zero cooldown (ready to fire).
//
// Parameters:
//   - weapon: Weapon definition to create state for
//
// Returns:
//   - *WeaponState: New weapon state with full ammo and no cooldown
//
// Thread-safe: Creates new state, no shared state modification.
func InitializeWeaponState(weapon *models.Weapon) *WeaponState {
	return &WeaponState{
		WeaponID:          weapon.ID,
		CurrentAmmo:       weapon.AmmoCapacity,
		CooldownRemaining: 0,
		LastFiredTurn:     0,
	}
}

// ReloadAmmo replenishes ammunition for missile weapons at stations or planets.
//
// This function handles ammunition reloading, respecting the weapon's ammo capacity.
// Only works for missile-type weapons; returns 0 for energy weapons.
//
// Parameters:
//   - weapon: Weapon definition (must be missile type)
//   - state: Weapon state to reload - modified by this function
//   - amount: Maximum amount of ammo to reload
//
// Returns:
//   - int: Actual amount of ammo reloaded (clamped to capacity)
//
// Side Effects:
//   - Modifies state.CurrentAmmo (adds reloaded ammo)
//
// Thread-safe: Modifies only passed parameters.
func ReloadAmmo(weapon *models.Weapon, state *WeaponState, amount int) int {
	if weapon.Type != "missile" {
		return 0
	}

	maxReload := weapon.AmmoCapacity - state.CurrentAmmo
	if amount > maxReload {
		amount = maxReload
	}

	state.CurrentAmmo += amount
	return amount
}

// GetDPS calculates the damage per second rating for a weapon.
//
// DPS is a useful metric for comparing weapon effectiveness. Calculated as
// base damage divided by cooldown time.
//
// Parameters:
//   - weapon: Weapon to calculate DPS for
//
// Returns:
//   - float64: Damage per second (0 if cooldown is 0)
//
// Thread-safe: No shared state, safe for concurrent calls.
func GetDPS(weapon *models.Weapon) float64 {
	if weapon.Cooldown == 0 {
		return 0
	}
	return float64(weapon.Damage) / weapon.Cooldown
}

// GetEffectiveRange calculates the maximum effective range for a weapon.
//
// Effective range is defined as the distance at which accuracy drops to a
// specified minimum threshold. Beyond this range, the weapon becomes unreliable.
//
// Calculation: Each 100 units beyond optimal range reduces accuracy by 1%.
//
// Parameters:
//   - weapon: Weapon to calculate effective range for
//   - minAccuracy: Minimum acceptable accuracy percentage (e.g., 50.0 for 50%)
//
// Returns:
//   - int: Maximum distance at which weapon maintains minimum accuracy
//
// Thread-safe: No shared state, safe for concurrent calls.
func GetEffectiveRange(weapon *models.Weapon, minAccuracy float64) int {
	// Calculate distance where accuracy drops to minAccuracy
	accuracyDrop := float64(weapon.Accuracy) - minAccuracy
	if accuracyDrop <= 0 {
		return weapon.RangeValue
	}

	// Each 100 units over range reduces accuracy by 1%
	extraRange := int(accuracyDrop * 100.0)
	return weapon.RangeValue + extraRange
}

// CalculateDistance computes the Euclidean distance between two 2D points.
//
// Uses the Pythagorean theorem: distance = sqrt((x2-x1)² + (y2-y1)²)
//
// Parameters:
//   - x1, y1: Coordinates of first point
//   - x2, y2: Coordinates of second point
//
// Returns:
//   - int: Distance in game units (rounded)
//
// Thread-safe: No shared state, safe for concurrent calls.
func CalculateDistance(x1, y1, x2, y2 int) int {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return int(math.Sqrt(dx*dx + dy*dy))
}

// GetWeaponTypeInfo returns human-readable descriptions of weapon types.
//
// Provides flavor text and tactical information about different weapon categories
// for display in UI, help systems, and tooltips.
//
// Supported Types:
//   - laser: Fast firing energy weapon
//   - missile: High damage explosive with limited ammo
//   - plasma: Balanced damage and penetration
//   - railgun: Extreme damage kinetic weapon
//
// Parameters:
//   - weaponType: Type identifier from weapon definition
//
// Returns:
//   - string: Human-readable description of weapon type and characteristics
//
// Thread-safe: No shared state, safe for concurrent calls.
func GetWeaponTypeInfo(weaponType string) string {
	switch weaponType {
	case "laser":
		return "Energy weapon - Fast firing, no ammo, moderate energy cost"
	case "missile":
		return "Explosive weapon - High damage, limited ammo, good shield penetration"
	case "plasma":
		return "Balanced weapon - Good damage and shield penetration, moderate energy cost"
	case "railgun":
		return "Kinetic weapon - Very high damage, excellent shield penetration, high energy cost"
	default:
		return "Unknown weapon type"
	}
}
