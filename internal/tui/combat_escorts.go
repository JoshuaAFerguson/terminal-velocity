// File: internal/tui/combat_escorts.go
// Project: Terminal Velocity
// Description: Fleet-escort participation in enhanced combat. P5B-1
//   scope is visual-present + damage-bonus: escorts render in the
//   combat UI and boost player weapon damage, but do not take damage
//   themselves (deferred to P5B-2).
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
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
		m.combatEnhanced.playerEscorts = append(m.combatEnhanced.playerEscorts, combatEscort{
			id:       e.ID.String(),
			pilot:    e.Pilot,
			shipType: e.Ship.Name,
			behavior: e.Behavior,
			level:    e.Level,
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
		if e.behavior == fleet.BehaviorSupport {
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
		line := fmt.Sprintf(" ▲ %s  (%s)  [%s]  Lv%d",
			e.pilot, e.shipType, escortBehaviorLabel(e.behavior), e.level)
		sb.WriteString(PadRight(line, width) + "\n")
	}
	// Footer summary line showing the active bonus.
	bonus := computeEscortDamageBonus(escorts)
	footer := fmt.Sprintf(" ⚔  Damage bonus: +%d%%   🛡  Support: %d",
		int((bonus-1.0)*100), countSupportEscorts(escorts))
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
