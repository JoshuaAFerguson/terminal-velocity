// File: internal/combat/weapons.go
// Project: Terminal Velocity
// Description: Combat system: weapons
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package combat

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// WeaponState tracks the runtime state of a weapon
type WeaponState struct {
	WeaponID          string
	CurrentAmmo       int     // current ammo (for missiles)
	CooldownRemaining float64 // seconds until weapon can fire again
	LastFiredTurn     int     // turn number when weapon was last fired
}

// FireResult contains the result of a weapon firing
type FireResult struct {
	Hit           bool
	Damage        int
	ShieldDamage  int
	HullDamage    int
	CriticalHit   bool
	AmmoRemaining int
	Message       string
}

// CanFire checks if a weapon can fire (cooldown and ammo check)
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

// Fire executes a weapon firing attempt
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

// CalculateHitChance calculates the chance to hit based on weapon accuracy,
// attacker/target stats, and distance
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

// UpdateCooldowns updates all weapon cooldowns based on time elapsed
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

// InitializeWeaponState creates initial state for a weapon
func InitializeWeaponState(weapon *models.Weapon) *WeaponState {
	return &WeaponState{
		WeaponID:          weapon.ID,
		CurrentAmmo:       weapon.AmmoCapacity,
		CooldownRemaining: 0,
		LastFiredTurn:     0,
	}
}

// ReloadAmmo reloads ammo for missile weapons (for use at stations/planets)
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

// GetDPS calculates damage per second for a weapon
func GetDPS(weapon *models.Weapon) float64 {
	if weapon.Cooldown == 0 {
		return 0
	}
	return float64(weapon.Damage) / weapon.Cooldown
}

// GetEffectiveRange returns the effective range of a weapon considering accuracy drop
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

// CalculateDistance calculates distance between two points (simple 2D for now)
func CalculateDistance(x1, y1, x2, y2 int) int {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return int(math.Sqrt(dx*dx + dy*dy))
}

// GetWeaponTypeInfo returns descriptive information about weapon types
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
