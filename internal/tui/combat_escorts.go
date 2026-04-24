// File: internal/tui/combat_escorts.go
// Project: Terminal Velocity
// Description: Fleet-escort participation in enhanced combat. P5B-1
//   added visual-present + damage-bonus (rendered strip + weapon
//   multiplier). P5B-2 adds per-turn escort actions: aggressive
//   escorts fire on the enemy, support escorts regen player shields,
//   defensive escorts probabilistically attack. Escorts still don't
//   take damage — that's P5B-3.
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
)

// combatEscort is the combat-layer projection of a fleet.Escort.
// Kept deliberately lightweight — we only snapshot the fields that
// combat_enhanced actually reads so that mutations on the fleet
// manager during combat can't silently corrupt the combat view.
type combatEscort struct {
	id       string // escort UUID as string for display/tests
	pilot    string
	shipType string
	behavior fleet.EscortBehavior
	level    int

	// P5B-3: per-escort hull tracking. maxHull is snapshotted at
	// combat entry from fleet.Escort.Ship.MaxHull (or defaults to
	// 100 when ship data is missing). destroyed short-circuits all
	// combat paths — resolver skips them, targeting skips them,
	// render strip shows them grayed.
	hull      int
	maxHull   int
	destroyed bool
}

// initializeCombatEscorts populates combatEnhancedModel.playerEscorts
// from the player's active fleet. No-op when the manager is nil (CLI
// tests / login state), which means solo-pilot combat still works
// without a fleet configured.
func (m *Model) initializeCombatEscorts() {
	m.combatEnhanced.playerEscorts = nil
	if m.fleetManager == nil {
		return
	}
	flt, ok := m.fleetManager.GetFleet(m.playerID)
	if !ok || flt == nil {
		return
	}
	for _, e := range flt.Escorts {
		if e == nil || e.Ship == nil {
			continue
		}
		// Defaults protect against ship records with zeroed hull
		// (fleet-manager TODO items): a freshly-hired escort with
		// unset MaxHull shouldn't start combat pre-destroyed.
		maxHull := e.Ship.Hull
		if maxHull <= 0 {
			maxHull = 100
		}
		m.combatEnhanced.playerEscorts = append(m.combatEnhanced.playerEscorts, combatEscort{
			id:       e.ID.String(),
			pilot:    e.Pilot,
			shipType: e.Ship.Name,
			behavior: e.Behavior,
			level:    e.Level,
			hull:     maxHull,
			maxHull:  maxHull,
		})
	}
}

// computeEscortDamageBonus returns a damage multiplier for the
// player's weapons based on their active escort loadout. The rules
// are deliberately simple for P5B-1:
//   - Aggressive:  +10% per escort
//   - Defensive:   +5%  per escort (they suppress the enemy too)
//   - Support:      0%  (contributes via shield regen, not DPS)
//   - Passive:      0%
//
// Bonus is capped at 1.5× so a fleet of 10 aggressive escorts
// doesn't trivialize every encounter. The level field is reserved
// for future tuning (P5B-2 will scale bonuses by level).
func computeEscortDamageBonus(escorts []combatEscort) float64 {
	bonus := 1.0
	for _, e := range escorts {
		if e.destroyed {
			continue
		}
		switch e.behavior {
		case fleet.BehaviorAggressive:
			bonus += 0.10
		case fleet.BehaviorDefensive:
			bonus += 0.05
		}
	}
	if bonus > 1.5 {
		bonus = 1.5
	}
	return bonus
}

// applyEscortBonus scales a base damage value by the escort fleet
// damage multiplier, rounding down. Pulled out as a pure function so
// fireWeaponCmd can call it once per shot and tests can assert the
// rounding behavior. Always returns >= baseDamage when escorts exist
// and the multiplier is > 1.0, except when baseDamage is zero (a
// miss, ammo-out branch etc.).
func applyEscortBonus(baseDamage int, escorts []combatEscort) int {
	if baseDamage <= 0 || len(escorts) == 0 {
		return baseDamage
	}
	bonus := computeEscortDamageBonus(escorts)
	return int(float64(baseDamage) * bonus)
}

// countSupportEscorts returns how many escorts are on Support
// behavior — used by the combat loop to apply small per-turn shield
// regen. Separate from the damage path so Support escorts never
// contribute to the offense multiplier.
func countSupportEscorts(escorts []combatEscort) int {
	n := 0
	for _, e := range escorts {
		if !e.destroyed && e.behavior == fleet.BehaviorSupport {
			n++
		}
	}
	return n
}

// renderEscortStrip returns the multi-line escort panel shown below
// the enemy status row in the combat view. Returns "" when no
// escorts are present so the view collapses the row cleanly instead
// of drawing an empty box.
//
// Format (one line per escort):
//
//	▲ <pilot>  (<ship>)  [<behavior>]  Lv<level>
//
// Truncated to `width` cells via PadRight so the right border aligns.
func renderEscortStrip(escorts []combatEscort, width int) string {
	if len(escorts) == 0 || width <= 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(" FLEET ESCORTS\n")
	sb.WriteString(" " + strings.Repeat("─", width-2) + "\n")
	for _, e := range escorts {
		var line string
		if e.destroyed {
			// Destroyed escorts stay in the strip until the combat
			// loop removes them at the end of the enemy turn; shown
			// with a wreck marker so the player sees the loss
			// without the UI flashing the row out.
			line = fmt.Sprintf(" ✗ %s  (%s)  [DESTROYED]", e.pilot, e.shipType)
		} else {
			hullPct := 0
			if e.maxHull > 0 {
				hullPct = (e.hull * 100) / e.maxHull
			}
			line = fmt.Sprintf(" ▲ %s  (%s)  [%s]  Lv%d  H:%d%%",
				e.pilot, e.shipType, escortBehaviorLabel(e.behavior), e.level, hullPct)
		}
		sb.WriteString(PadRight(line, width) + "\n")
	}
	// Footer: computeEscortDamageBonus and countSupportEscorts both
	// skip destroyed escorts internally, so the numbers shown here
	// reflect the live effective fleet, not the starting roster.
	bonus := computeEscortDamageBonus(escorts)
	alive := 0
	for _, e := range escorts {
		if !e.destroyed {
			alive++
		}
	}
	footer := fmt.Sprintf(" ⚔  Damage bonus: +%d%%   🛡  Support: %d   🚀  Alive: %d",
		int((bonus-1.0)*100), countSupportEscorts(escorts), alive)
	sb.WriteString(PadRight(footer, width) + "\n")
	return sb.String()
}

// escortBehaviorLabel maps the internal constant to a short UI
// label. Keeps the combat UI decoupled from any future renaming in
// the fleet package.
func escortBehaviorLabel(b fleet.EscortBehavior) string {
	switch b {
	case fleet.BehaviorAggressive:
		return "AGG"
	case fleet.BehaviorDefensive:
		return "DEF"
	case fleet.BehaviorPassive:
		return "PAS"
	case fleet.BehaviorSupport:
		return "SUP"
	default:
		return "???"
	}
}

// ============================================================================
// P5B-2: escort AI turns
// ============================================================================

// escortActionKind tags what an escort did this turn. Drives both the
// log message and the mutation that combat_enhanced applies.
type escortActionKind string

const (
	escortActionAttack escortActionKind = "attack"
	escortActionHeal   escortActionKind = "heal"
	escortActionIdle   escortActionKind = "idle"
)

// escortAction is one escort's outcome for a single combat turn. All
// numeric fields are already fully resolved (level-scaled, RNG-rolled)
// so applyEscortActions is pure state mutation — no further rolls.
type escortAction struct {
	kind    escortActionKind
	pilot   string
	damage  int // for attack
	heal    int // for heal
	message string
}

// escortRNG is the seam that lets tests inject a deterministic RNG.
// math/rand.Rand satisfies this via its Float64 and Intn methods.
type escortRNG interface {
	Float64() float64
	Intn(n int) int
}

// baseEscortAttackDamage is the damage a level-1 aggressive escort
// deals before level scaling. Level scaling is 1 + level/10, so a
// level-10 pilot deals 2× base. Kept tiny on purpose — escorts should
// support the player, not replace them.
const baseEscortAttackDamage = 12

// defensiveAttackChance is the probability a defensive escort engages
// on any given turn. They're primarily reactive/protective in lore,
// so they fire ~40% of turns and otherwise brace.
const defensiveAttackChance = 0.40

// supportShieldRegenPerLevel is how many shield points a support
// escort restores per level-tier. A level-5 support escort regens
// 5 × 1.0 = 5 shield per turn; level-10 regens 10. Capped later by
// the player's max shields at application time.
const supportShieldRegenPerLevel = 1.0

// levelScale returns the damage/heal multiplier for a given escort
// level. Level-0 (shouldn't occur) is treated as level-1 so callers
// never see a 0× scale.
func levelScale(level int) float64 {
	if level < 1 {
		level = 1
	}
	return 1.0 + float64(level)/10.0
}

// resolveEscortActions rolls one action per escort for the current
// turn. Pure given the RNG — the same RNG seed + escort slice always
// produces the same actions. combat_enhanced calls this once per
// player turn and then applies the returned actions via
// applyEscortActions.
//
// Behavior-by-behavior:
//   - Aggressive: always attacks. damage = baseEscortAttackDamage * levelScale
//   - Defensive:  40% chance to attack; otherwise idle ("bracing")
//   - Support:    regens player shields; no attack
//   - Passive:    idle every turn
//   - Unknown:    idle (safe fallback for future behaviors)
func resolveEscortActions(escorts []combatEscort, rng escortRNG) []escortAction {
	if len(escorts) == 0 {
		return nil
	}
	actions := make([]escortAction, 0, len(escorts))
	for _, e := range escorts {
		// Destroyed escorts skip their turn — still in the slice
		// for render purposes (grayed out) until the combat loop
		// removes them, but they don't roll.
		if e.destroyed {
			continue
		}
		switch e.behavior {
		case fleet.BehaviorAggressive:
			dmg := int(float64(baseEscortAttackDamage) * levelScale(e.level))
			actions = append(actions, escortAction{
				kind:    escortActionAttack,
				pilot:   e.pilot,
				damage:  dmg,
				message: fmt.Sprintf("%s (AGG) strafes the enemy for %d damage!", e.pilot, dmg),
			})
		case fleet.BehaviorDefensive:
			if rng.Float64() < defensiveAttackChance {
				dmg := int(float64(baseEscortAttackDamage) * levelScale(e.level) * 0.75)
				actions = append(actions, escortAction{
					kind:    escortActionAttack,
					pilot:   e.pilot,
					damage:  dmg,
					message: fmt.Sprintf("%s (DEF) breaks formation to attack for %d damage!", e.pilot, dmg),
				})
			} else {
				actions = append(actions, escortAction{
					kind:    escortActionIdle,
					pilot:   e.pilot,
					message: fmt.Sprintf("%s (DEF) holds the line.", e.pilot),
				})
			}
		case fleet.BehaviorSupport:
			heal := int(float64(e.level) * supportShieldRegenPerLevel)
			if heal < 1 {
				heal = 1
			}
			actions = append(actions, escortAction{
				kind:    escortActionHeal,
				pilot:   e.pilot,
				heal:    heal,
				message: fmt.Sprintf("%s (SUP) diverts shield power to you: +%d shields.", e.pilot, heal),
			})
		default:
			// Passive / unknown — no-op, no log line (escort strip
			// already shows they're on passive).
		}
	}
	return actions
}

// applyEscortActions folds the resolved actions into mutable combat
// state: attacks chip the enemy's shields/hull (shields first, same
// as player fire), heals raise the player's shields up to the cap.
// Returns the log messages in order so the caller can append them to
// the combat log. Pure-ish: no RNG, just arithmetic on the structs.
//
// Accepts pointers so the caller (combat_enhanced's combatActionMsg
// handler) can mutate the live combat state without assigning back.
func applyEscortActions(actions []escortAction, playerShip, enemyShip *combatShip) []string {
	if len(actions) == 0 || playerShip == nil || enemyShip == nil {
		return nil
	}
	logs := make([]string, 0, len(actions))
	for _, a := range actions {
		switch a.kind {
		case escortActionAttack:
			// Skip chipping a corpse — if the player's own fire
			// already killed the enemy this turn, the escort fire
			// is just narrative, not stat damage.
			if enemyShip.hull > 0 {
				applyShipDamage(enemyShip, a.damage)
			}
			logs = append(logs, a.message)
		case escortActionHeal:
			playerShip.shields += a.heal
			if playerShip.shields > playerShip.maxShields {
				playerShip.shields = playerShip.maxShields
			}
			logs = append(logs, a.message)
		case escortActionIdle:
			logs = append(logs, a.message)
		}
	}
	return logs
}

// applyShipDamage applies damage to a combatShip, shields first
// then hull, floor-clamped at 0. Shared with the player-fire path
// (see fireWeaponCmd) so both routes obey identical rules.
func applyShipDamage(ship *combatShip, damage int) {
	if damage <= 0 || ship == nil {
		return
	}
	if ship.shields > 0 {
		if ship.shields >= damage {
			ship.shields -= damage
			return
		}
		damage -= ship.shields
		ship.shields = 0
	}
	ship.hull -= damage
	if ship.hull < 0 {
		ship.hull = 0
	}
}

// globalRandRNG is a zero-alloc adapter from math/rand's package
// functions to the escortRNG interface. We don't want to allocate a
// fresh rand.Rand per combat turn — the global source is already
// seeded by the runtime (Go 1.20+), and combat rolls don't need
// cryptographic quality.
type globalRandRNG struct{}

func (globalRandRNG) Float64() float64 { return rand.Float64() }
func (globalRandRNG) Intn(n int) int   { return rand.Intn(n) }

// defaultEscortRNG returns the production RNG used by combat_enhanced.
// Tests inject a deterministic fixedRNG via resolveEscortActions directly.
func defaultEscortRNG() escortRNG { return globalRandRNG{} }

// ============================================================================
// P5B-3: enemy targeting escorts + destruction
// ============================================================================

// escortInterceptChancePerAlive is the per-escort probability that an
// incoming enemy shot is redirected onto that escort. With 1 alive
// escort the chance is 15%, 2 is 30%, 3 is 45%, 4+ is clamped at
// escortInterceptCap. Tuned so a small fleet feels protective without
// trivializing enemy threat.
const escortInterceptChancePerAlive = 0.15

// escortInterceptCap is the ceiling on combined intercept probability.
// Without this, 6 escorts would redirect 90% of hits, which makes the
// player ship functionally invulnerable.
const escortInterceptCap = 0.60

// selectEscortInterceptIndex decides whether an incoming enemy hit is
// intercepted by an escort, and if so, which escort takes it. Returns
// -1 to indicate the player ship takes the hit. Pure given the RNG.
//
// The intercept roll uses Float64() once; the target selection uses
// Intn() once. Tests can assert both legs via the fixedRNG seam.
func selectEscortInterceptIndex(escorts []combatEscort, rng escortRNG) int {
	aliveIndices := make([]int, 0, len(escorts))
	for i, e := range escorts {
		if !e.destroyed {
			aliveIndices = append(aliveIndices, i)
		}
	}
	if len(aliveIndices) == 0 {
		return -1
	}
	chance := float64(len(aliveIndices)) * escortInterceptChancePerAlive
	if chance > escortInterceptCap {
		chance = escortInterceptCap
	}
	if rng.Float64() >= chance {
		return -1
	}
	// Pick a random alive escort to take the hit.
	return aliveIndices[rng.Intn(len(aliveIndices))]
}

// applyEscortHit applies `damage` to the escort at `index`, clamping
// hull at 0 and flipping the destroyed flag when that happens.
// Returns (logMessage, becameDestroyed). Callers are expected to pass
// the address of the slice element, not a copy.
func applyEscortHit(e *combatEscort, damage int) (string, bool) {
	if e == nil || damage <= 0 || e.destroyed {
		return "", false
	}
	e.hull -= damage
	if e.hull <= 0 {
		e.hull = 0
		e.destroyed = true
		return fmt.Sprintf("%s's ship has been destroyed!", e.pilot), true
	}
	return fmt.Sprintf("%s intercepts the hit — %d hull damage (%d/%d remaining).",
		e.pilot, damage, e.hull, e.maxHull), false
}

// removeDestroyedEscorts returns a new slice containing only escorts
// that are still alive. Used after an enemy turn to clear dead
// escorts out of the combat view. Kept pure so tests can assert
// without needing the full Model.
func removeDestroyedEscorts(escorts []combatEscort) []combatEscort {
	alive := make([]combatEscort, 0, len(escorts))
	for _, e := range escorts {
		if !e.destroyed {
			alive = append(alive, e)
		}
	}
	return alive
}
