// File: internal/tui/combat_escorts_test.go
// Project: Terminal Velocity
// Description: Unit tests for combat escort helpers — damage bonus
//   math, support counting, and strip rendering.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"strings"
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
)

func esc(behavior fleet.EscortBehavior) combatEscort {
	return combatEscort{
		id:       "test",
		pilot:    "Test Pilot",
		shipType: "Corvette",
		behavior: behavior,
		level:    5,
	}
}

func TestComputeEscortDamageBonus(t *testing.T) {
	tests := []struct {
		name    string
		escorts []combatEscort
		want    float64
	}{
		{"no escorts is 1.0x", nil, 1.0},
		{"empty slice is 1.0x", []combatEscort{}, 1.0},
		{"1 aggressive is 1.10x", []combatEscort{esc(fleet.BehaviorAggressive)}, 1.10},
		{"1 defensive is 1.05x", []combatEscort{esc(fleet.BehaviorDefensive)}, 1.05},
		{"1 passive is 1.00x", []combatEscort{esc(fleet.BehaviorPassive)}, 1.00},
		{"1 support is 1.00x (no DPS contribution)", []combatEscort{esc(fleet.BehaviorSupport)}, 1.00},
		{
			name: "3 aggressive is 1.30x",
			escorts: []combatEscort{
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
			},
			want: 1.30,
		},
		{
			name: "2 aggressive + 1 defensive is 1.25x",
			escorts: []combatEscort{
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorDefensive),
			},
			want: 1.25,
		},
		{
			name: "mixed fleet: passive and support don't count",
			escorts: []combatEscort{
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorPassive),
				esc(fleet.BehaviorSupport),
			},
			want: 1.10,
		},
		{
			name: "bonus is capped at 1.5x",
			escorts: []combatEscort{
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
				esc(fleet.BehaviorAggressive),
			},
			want: 1.5,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := computeEscortDamageBonus(tc.escorts)
			// Tolerate small float rounding — our math is all in
			// 0.05 steps so 1e-9 tolerance is generous.
			if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApplyEscortBonus(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage int
		escorts    []combatEscort
		want       int
	}{
		{"zero base stays zero (miss)", 0, []combatEscort{esc(fleet.BehaviorAggressive)}, 0},
		{"negative base is unchanged", -10, []combatEscort{esc(fleet.BehaviorAggressive)}, -10},
		{"no escorts returns base", 50, nil, 50},
		{"empty escorts returns base", 50, []combatEscort{}, 50},
		{"50 base × 1.10 = 55", 50, []combatEscort{esc(fleet.BehaviorAggressive)}, 55},
		{"35 base × 1.15 (agg+def) = 40 (truncated)", 35, []combatEscort{esc(fleet.BehaviorAggressive), esc(fleet.BehaviorDefensive)}, 40},
		// 35 * 1.15 = 40.25 → int = 40 (truncation)
		{"100 base × 1.5 cap = 150", 100, []combatEscort{
			esc(fleet.BehaviorAggressive), esc(fleet.BehaviorAggressive), esc(fleet.BehaviorAggressive),
			esc(fleet.BehaviorAggressive), esc(fleet.BehaviorAggressive), esc(fleet.BehaviorAggressive),
		}, 150},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := applyEscortBonus(tc.baseDamage, tc.escorts)
			if got != tc.want {
				t.Fatalf("applyEscortBonus(%d, ...) = %d, want %d", tc.baseDamage, got, tc.want)
			}
		})
	}
}

func TestCountSupportEscorts(t *testing.T) {
	tests := []struct {
		name    string
		escorts []combatEscort
		want    int
	}{
		{"empty returns 0", nil, 0},
		{"no support escorts returns 0", []combatEscort{
			esc(fleet.BehaviorAggressive), esc(fleet.BehaviorDefensive),
		}, 0},
		{"counts only support", []combatEscort{
			esc(fleet.BehaviorSupport), esc(fleet.BehaviorAggressive), esc(fleet.BehaviorSupport),
		}, 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := countSupportEscorts(tc.escorts)
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestEscortBehaviorLabel(t *testing.T) {
	tests := map[fleet.EscortBehavior]string{
		fleet.BehaviorAggressive:  "AGG",
		fleet.BehaviorDefensive:   "DEF",
		fleet.BehaviorPassive:     "PAS",
		fleet.BehaviorSupport:     "SUP",
		fleet.EscortBehavior("?"): "???",
	}
	for b, want := range tests {
		if got := escortBehaviorLabel(b); got != want {
			t.Errorf("escortBehaviorLabel(%q) = %q, want %q", b, got, want)
		}
	}
}

func TestRenderEscortStripEmpty(t *testing.T) {
	if got := renderEscortStrip(nil, 60); got != "" {
		t.Fatalf("nil escorts should render empty, got %q", got)
	}
	if got := renderEscortStrip([]combatEscort{esc(fleet.BehaviorAggressive)}, 0); got != "" {
		t.Fatalf("zero width should render empty, got %q", got)
	}
}

func TestRenderEscortStripContent(t *testing.T) {
	out := renderEscortStrip([]combatEscort{
		{pilot: "Kira Rehn", shipType: "Interceptor", behavior: fleet.BehaviorAggressive, level: 3},
		{pilot: "Drex", shipType: "Gunship", behavior: fleet.BehaviorSupport, level: 7},
	}, 60)

	if !strings.Contains(out, "FLEET ESCORTS") {
		t.Fatalf("strip missing heading: %q", out)
	}
	if !strings.Contains(out, "Kira Rehn") || !strings.Contains(out, "Drex") {
		t.Fatalf("strip missing pilot names: %q", out)
	}
	if !strings.Contains(out, "Interceptor") || !strings.Contains(out, "Gunship") {
		t.Fatalf("strip missing ship types: %q", out)
	}
	if !strings.Contains(out, "[AGG]") || !strings.Contains(out, "[SUP]") {
		t.Fatalf("strip missing behavior labels: %q", out)
	}
	// Footer should report the aggressive damage bonus (+10%) and
	// that one support escort is active. Literal +10% for one
	// aggressive + 0 for support = +10%.
	if !strings.Contains(out, "+10%") {
		t.Fatalf("strip footer missing +10%% bonus: %q", out)
	}
	if !strings.Contains(out, "Support: 1") {
		t.Fatalf("strip footer missing support count: %q", out)
	}
}

// TestRenderEscortStripPreview dumps sample output so a reviewer can
// eyeball the strip with `go test -v`. Covers three states: full
// health, mid-damage, and with a destroyed escort still visible.
func TestRenderEscortStripPreview(t *testing.T) {
	full := []combatEscort{
		{pilot: "Kira Rehn", shipType: "Interceptor", behavior: fleet.BehaviorAggressive, level: 3, hull: 100, maxHull: 100},
		{pilot: "Maz Oort", shipType: "Corvette", behavior: fleet.BehaviorDefensive, level: 5, hull: 100, maxHull: 100},
		{pilot: "Drex", shipType: "Gunship", behavior: fleet.BehaviorSupport, level: 7, hull: 100, maxHull: 100},
	}
	damaged := []combatEscort{
		{pilot: "Kira Rehn", shipType: "Interceptor", behavior: fleet.BehaviorAggressive, level: 3, hull: 45, maxHull: 100},
		{pilot: "Maz Oort", shipType: "Corvette", behavior: fleet.BehaviorDefensive, level: 5, hull: 100, maxHull: 100},
		{pilot: "Drex", shipType: "Gunship", behavior: fleet.BehaviorSupport, level: 7, hull: 20, maxHull: 100},
	}
	oneDown := []combatEscort{
		{pilot: "Kira Rehn", shipType: "Interceptor", behavior: fleet.BehaviorAggressive, level: 3, hull: 45, maxHull: 100},
		{pilot: "Maz Oort", shipType: "Corvette", behavior: fleet.BehaviorDefensive, level: 5, hull: 0, maxHull: 100, destroyed: true},
		{pilot: "Drex", shipType: "Gunship", behavior: fleet.BehaviorSupport, level: 7, hull: 20, maxHull: 100},
	}

	t.Logf("full health (width=70):\n%s", renderEscortStrip(full, 70))
	t.Logf("damaged (width=70):\n%s", renderEscortStrip(damaged, 70))
	t.Logf("one destroyed — row shown until combat loop cleans it (width=70):\n%s", renderEscortStrip(oneDown, 70))
}

// ============================================================================
// P5B-2 tests: escort AI turns
// ============================================================================

// fixedRNG is a deterministic escortRNG for unit tests — always returns
// the configured float and int, regardless of the input range.
type fixedRNG struct {
	f float64
	i int
}

func (r fixedRNG) Float64() float64 { return r.f }
func (r fixedRNG) Intn(n int) int   { return r.i }

func TestLevelScale(t *testing.T) {
	tests := []struct {
		level int
		want  float64
	}{
		{0, 1.1},  // clamp to 1
		{-5, 1.1}, // clamp to 1
		{1, 1.1},
		{5, 1.5},
		{10, 2.0},
	}
	for _, tc := range tests {
		if got := levelScale(tc.level); got != tc.want {
			t.Errorf("levelScale(%d) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

func TestResolveEscortActionsEmpty(t *testing.T) {
	if got := resolveEscortActions(nil, fixedRNG{f: 0.0}); got != nil {
		t.Fatalf("nil escorts should return nil, got %v", got)
	}
	if got := resolveEscortActions([]combatEscort{}, fixedRNG{f: 0.0}); got != nil {
		t.Fatalf("empty escorts should return nil, got %v", got)
	}
}

func TestResolveEscortActionsAggressiveAlwaysAttacks(t *testing.T) {
	escorts := []combatEscort{
		{pilot: "A", behavior: fleet.BehaviorAggressive, level: 1},
		{pilot: "B", behavior: fleet.BehaviorAggressive, level: 10},
	}
	// Force roll=0.99 to show aggressive ignores RNG.
	actions := resolveEscortActions(escorts, fixedRNG{f: 0.99})

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	for _, a := range actions {
		if a.kind != escortActionAttack {
			t.Fatalf("aggressive escort produced non-attack action: %v", a)
		}
	}
	// level 1 → 12 * 1.1 = 13 (truncated)
	if actions[0].damage != 13 {
		t.Fatalf("level 1 aggressive damage: got %d, want 13", actions[0].damage)
	}
	// level 10 → 12 * 2.0 = 24
	if actions[1].damage != 24 {
		t.Fatalf("level 10 aggressive damage: got %d, want 24", actions[1].damage)
	}
}

func TestResolveEscortActionsDefensiveRollsForAttack(t *testing.T) {
	escort := combatEscort{pilot: "D", behavior: fleet.BehaviorDefensive, level: 10}

	// Roll below threshold → attack.
	acts := resolveEscortActions([]combatEscort{escort}, fixedRNG{f: 0.1})
	if len(acts) != 1 || acts[0].kind != escortActionAttack {
		t.Fatalf("defensive with low roll should attack, got %v", acts)
	}
	// Damage is 0.75 × 12 × 2.0 = 18
	if acts[0].damage != 18 {
		t.Fatalf("defensive lv10 damage: got %d, want 18", acts[0].damage)
	}

	// Roll above threshold → idle.
	acts = resolveEscortActions([]combatEscort{escort}, fixedRNG{f: 0.9})
	if len(acts) != 1 || acts[0].kind != escortActionIdle {
		t.Fatalf("defensive with high roll should idle, got %v", acts)
	}
}

func TestResolveEscortActionsSupportHealsScaledByLevel(t *testing.T) {
	escorts := []combatEscort{
		{pilot: "S1", behavior: fleet.BehaviorSupport, level: 0},
		{pilot: "S5", behavior: fleet.BehaviorSupport, level: 5},
		{pilot: "S10", behavior: fleet.BehaviorSupport, level: 10},
	}
	acts := resolveEscortActions(escorts, fixedRNG{f: 0.0})
	if len(acts) != 3 {
		t.Fatalf("expected 3 heal actions, got %d", len(acts))
	}
	// Level 0 → floor-clamped to 1
	if acts[0].heal != 1 {
		t.Fatalf("level 0 support heal: got %d, want 1", acts[0].heal)
	}
	if acts[1].heal != 5 {
		t.Fatalf("level 5 support heal: got %d, want 5", acts[1].heal)
	}
	if acts[2].heal != 10 {
		t.Fatalf("level 10 support heal: got %d, want 10", acts[2].heal)
	}
	for _, a := range acts {
		if a.kind != escortActionHeal {
			t.Fatalf("support escort produced non-heal action: %v", a)
		}
	}
}

func TestResolveEscortActionsPassiveIsSkipped(t *testing.T) {
	escorts := []combatEscort{
		{pilot: "P", behavior: fleet.BehaviorPassive, level: 5},
		{pilot: "A", behavior: fleet.BehaviorAggressive, level: 5},
	}
	acts := resolveEscortActions(escorts, fixedRNG{f: 0.0})
	if len(acts) != 1 {
		t.Fatalf("passive should be skipped, got %d actions", len(acts))
	}
	if acts[0].pilot != "A" {
		t.Fatalf("expected aggressive escort action, got %q", acts[0].pilot)
	}
}

func TestApplyEscortActionsAttackChipsShieldsThenHull(t *testing.T) {
	player := &combatShip{shields: 100, maxShields: 100, hull: 100}
	enemy := &combatShip{shields: 20, maxShields: 100, hull: 100}

	acts := []escortAction{
		{kind: escortActionAttack, pilot: "A", damage: 50, message: "strafe"},
	}
	logs := applyEscortActions(acts, player, enemy)

	if enemy.shields != 0 {
		t.Errorf("enemy shields should be drained: got %d, want 0", enemy.shields)
	}
	if enemy.hull != 70 {
		t.Errorf("enemy hull should take remainder: got %d, want 70", enemy.hull)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log line, got %d", len(logs))
	}
}

func TestApplyEscortActionsAttackSkipsDeadEnemy(t *testing.T) {
	player := &combatShip{shields: 100, maxShields: 100, hull: 100}
	enemy := &combatShip{shields: 0, maxShields: 100, hull: 0} // already destroyed

	acts := []escortAction{
		{kind: escortActionAttack, pilot: "A", damage: 999, message: "chip"},
	}
	logs := applyEscortActions(acts, player, enemy)

	// Hull stays 0, not negative — no over-kill accumulation.
	if enemy.hull != 0 {
		t.Errorf("dead enemy hull should stay 0, got %d", enemy.hull)
	}
	// Log line still prints — narrative flavor for "I fired on the wreck" is fine.
	if len(logs) != 1 {
		t.Errorf("expected log line even on dead enemy (narrative), got %d", len(logs))
	}
}

func TestApplyEscortActionsHealClampsAtMaxShields(t *testing.T) {
	player := &combatShip{shields: 95, maxShields: 100, hull: 100}
	enemy := &combatShip{shields: 50, maxShields: 100, hull: 100}

	acts := []escortAction{
		{kind: escortActionHeal, pilot: "S", heal: 20, message: "regen"},
	}
	applyEscortActions(acts, player, enemy)

	if player.shields != 100 {
		t.Errorf("player shields should clamp at max 100, got %d", player.shields)
	}
	if enemy.shields != 50 {
		t.Errorf("enemy shields should be untouched by heal, got %d", enemy.shields)
	}
}

func TestApplyEscortActionsIdleEmitsLogNoMutation(t *testing.T) {
	player := &combatShip{shields: 50, maxShields: 100, hull: 100}
	enemy := &combatShip{shields: 50, maxShields: 100, hull: 100}

	acts := []escortAction{
		{kind: escortActionIdle, pilot: "D", message: "holds"},
	}
	logs := applyEscortActions(acts, player, enemy)

	if player.shields != 50 || enemy.shields != 50 || enemy.hull != 100 {
		t.Errorf("idle action should not mutate state; got player.shields=%d enemy.shields=%d enemy.hull=%d",
			player.shields, enemy.shields, enemy.hull)
	}
	if len(logs) != 1 {
		t.Errorf("idle should still log, got %d lines", len(logs))
	}
}

func TestApplyEscortActionsNilGuards(t *testing.T) {
	// Nil actions → nil logs, no panic.
	if logs := applyEscortActions(nil, &combatShip{}, &combatShip{}); logs != nil {
		t.Errorf("nil actions should return nil logs, got %v", logs)
	}
	// Nil ships → nil logs, no panic.
	acts := []escortAction{{kind: escortActionAttack, damage: 50}}
	if logs := applyEscortActions(acts, nil, &combatShip{}); logs != nil {
		t.Errorf("nil player should return nil logs, got %v", logs)
	}
	if logs := applyEscortActions(acts, &combatShip{}, nil); logs != nil {
		t.Errorf("nil enemy should return nil logs, got %v", logs)
	}
}

// ============================================================================
// P5B-3 tests: escort targeting + destruction
// ============================================================================

func escWithHull(behavior fleet.EscortBehavior, hull, maxHull int, destroyed bool) combatEscort {
	return combatEscort{
		id:        "test-" + string(behavior),
		pilot:     "Test Pilot",
		shipType:  "Corvette",
		behavior:  behavior,
		level:     5,
		hull:      hull,
		maxHull:   maxHull,
		destroyed: destroyed,
	}
}

func TestSelectEscortInterceptIndexNoEscorts(t *testing.T) {
	// Roll extremely low — if the function didn't guard against the
	// empty case, this test would panic on Intn(0).
	if got := selectEscortInterceptIndex(nil, fixedRNG{f: 0.0}); got != -1 {
		t.Fatalf("empty escorts should return -1, got %d", got)
	}
	if got := selectEscortInterceptIndex([]combatEscort{}, fixedRNG{f: 0.0}); got != -1 {
		t.Fatalf("empty slice should return -1, got %d", got)
	}
}

func TestSelectEscortInterceptIndexAllDestroyed(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 0, 100, true),
		escWithHull(fleet.BehaviorDefensive, 0, 100, true),
	}
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.0}); got != -1 {
		t.Fatalf("all-destroyed fleet should return -1, got %d", got)
	}
}

func TestSelectEscortInterceptIndexRollAboveThreshold(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 100, 100, false),
	}
	// 1 escort → 15% threshold. Roll 0.20 → player takes the hit.
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.20}); got != -1 {
		t.Fatalf("roll above threshold should return -1, got %d", got)
	}
}

func TestSelectEscortInterceptIndexRollBelowThreshold(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 100, 100, false),
	}
	// 1 escort → 15% threshold. Roll 0.10 → escort intercepts.
	// fixedRNG.Intn always returns 0 → picks first alive escort.
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.10, i: 0}); got != 0 {
		t.Fatalf("roll below threshold should intercept, got %d", got)
	}
}

func TestSelectEscortInterceptIndexCapsAt60Percent(t *testing.T) {
	// 6 escorts would be 90% naïvely, but cap is 60%. Roll 0.61 →
	// player takes hit; roll 0.59 → escort intercepts.
	escorts := make([]combatEscort, 6)
	for i := range escorts {
		escorts[i] = escWithHull(fleet.BehaviorAggressive, 100, 100, false)
	}
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.61}); got != -1 {
		t.Fatalf("roll above 60%% cap should return -1 with 6 escorts, got %d", got)
	}
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.59, i: 2}); got != 2 {
		t.Fatalf("roll below 60%% cap should intercept with 6 escorts, got %d", got)
	}
}

func TestSelectEscortInterceptIndexSkipsDestroyed(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 0, 100, true),
		escWithHull(fleet.BehaviorAggressive, 100, 100, false),
		escWithHull(fleet.BehaviorAggressive, 0, 100, true),
	}
	// fixedRNG.Intn returns 0, but alive list is [1], so it must
	// map back to index 1 in the original slice.
	if got := selectEscortInterceptIndex(escorts, fixedRNG{f: 0.0, i: 0}); got != 1 {
		t.Fatalf("expected interception at original index 1 (only alive), got %d", got)
	}
}

func TestApplyEscortHit(t *testing.T) {
	t.Run("hit reduces hull, not destroyed yet", func(t *testing.T) {
		e := escWithHull(fleet.BehaviorAggressive, 50, 100, false)
		msg, destroyed := applyEscortHit(&e, 30)
		if destroyed {
			t.Error("should not be destroyed yet")
		}
		if e.hull != 20 {
			t.Errorf("hull: got %d, want 20", e.hull)
		}
		if e.destroyed {
			t.Error("destroyed flag should not be set yet")
		}
		if !strings.Contains(msg, "intercepts") {
			t.Errorf("log line should mention intercepts, got %q", msg)
		}
	})

	t.Run("exact-kill hit marks destroyed", func(t *testing.T) {
		e := escWithHull(fleet.BehaviorAggressive, 30, 100, false)
		msg, destroyed := applyEscortHit(&e, 30)
		if !destroyed {
			t.Error("should be destroyed on exact-kill")
		}
		if e.hull != 0 {
			t.Errorf("hull should clamp at 0, got %d", e.hull)
		}
		if !e.destroyed {
			t.Error("destroyed flag should be set")
		}
		if !strings.Contains(msg, "destroyed") {
			t.Errorf("log line should mention destroyed, got %q", msg)
		}
	})

	t.Run("over-kill clamps at 0 and marks destroyed", func(t *testing.T) {
		e := escWithHull(fleet.BehaviorAggressive, 10, 100, false)
		_, destroyed := applyEscortHit(&e, 999)
		if !destroyed {
			t.Error("should be destroyed on over-kill")
		}
		if e.hull != 0 {
			t.Errorf("hull should clamp at 0, got %d", e.hull)
		}
	})

	t.Run("zero-damage hit is a no-op", func(t *testing.T) {
		e := escWithHull(fleet.BehaviorAggressive, 50, 100, false)
		msg, destroyed := applyEscortHit(&e, 0)
		if destroyed {
			t.Error("zero damage should not destroy")
		}
		if e.hull != 50 {
			t.Errorf("hull should be unchanged, got %d", e.hull)
		}
		if msg != "" {
			t.Errorf("zero damage should return empty log, got %q", msg)
		}
	})

	t.Run("already-destroyed escort is a no-op", func(t *testing.T) {
		e := escWithHull(fleet.BehaviorAggressive, 0, 100, true)
		msg, destroyed := applyEscortHit(&e, 50)
		if destroyed {
			t.Error("already-destroyed should not re-destroy")
		}
		if msg != "" {
			t.Errorf("already-destroyed should return empty log, got %q", msg)
		}
	})

	t.Run("nil escort is a no-op", func(t *testing.T) {
		msg, destroyed := applyEscortHit(nil, 50)
		if destroyed {
			t.Error("nil escort should not report destroyed")
		}
		if msg != "" {
			t.Errorf("nil escort should return empty log, got %q", msg)
		}
	})
}

func TestRemoveDestroyedEscorts(t *testing.T) {
	escorts := []combatEscort{
		{pilot: "A", destroyed: false},
		{pilot: "B", destroyed: true},
		{pilot: "C", destroyed: false},
		{pilot: "D", destroyed: true},
	}
	alive := removeDestroyedEscorts(escorts)
	if len(alive) != 2 {
		t.Fatalf("expected 2 alive escorts, got %d", len(alive))
	}
	if alive[0].pilot != "A" || alive[1].pilot != "C" {
		t.Fatalf("expected [A, C], got [%s, %s]", alive[0].pilot, alive[1].pilot)
	}

	// Empty input returns empty (not nil — we always allocate).
	if got := removeDestroyedEscorts(nil); got == nil || len(got) != 0 {
		t.Errorf("nil input should return empty (not nil) slice, got %v", got)
	}
}

func TestComputeEscortDamageBonusSkipsDestroyed(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 100, 100, false),
		escWithHull(fleet.BehaviorAggressive, 0, 100, true), // destroyed — shouldn't count
		escWithHull(fleet.BehaviorDefensive, 100, 100, false),
	}
	got := computeEscortDamageBonus(escorts)
	// Alive: 1 aggressive (+0.10) + 1 defensive (+0.05) = 1.15
	if diff := got - 1.15; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("destroyed escort should not contribute to bonus; got %v, want 1.15", got)
	}
}

func TestCountSupportEscortsSkipsDestroyed(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorSupport, 100, 100, false),
		escWithHull(fleet.BehaviorSupport, 0, 100, true), // destroyed
		escWithHull(fleet.BehaviorSupport, 100, 100, false),
	}
	if got := countSupportEscorts(escorts); got != 2 {
		t.Fatalf("destroyed support should not count; got %d, want 2", got)
	}
}

func TestResolveEscortActionsSkipsDestroyed(t *testing.T) {
	escorts := []combatEscort{
		escWithHull(fleet.BehaviorAggressive, 0, 100, true), // destroyed
		escWithHull(fleet.BehaviorAggressive, 100, 100, false),
	}
	acts := resolveEscortActions(escorts, fixedRNG{f: 0.0})
	if len(acts) != 1 {
		t.Fatalf("destroyed escort should not take a turn; got %d actions", len(acts))
	}
}

func TestApplyShipDamageShieldThenHull(t *testing.T) {
	s := &combatShip{shields: 30, hull: 100}
	applyShipDamage(s, 50)
	if s.shields != 0 {
		t.Errorf("shields should drain first: got %d, want 0", s.shields)
	}
	if s.hull != 80 {
		t.Errorf("hull should take remainder: got %d, want 80", s.hull)
	}

	// Exact-shield kill: no hull damage.
	s = &combatShip{shields: 30, hull: 100}
	applyShipDamage(s, 30)
	if s.shields != 0 || s.hull != 100 {
		t.Errorf("exact-shield kill: got shields=%d hull=%d, want 0/100", s.shields, s.hull)
	}

	// Negative/zero damage is a no-op.
	s = &combatShip{shields: 30, hull: 100}
	applyShipDamage(s, 0)
	applyShipDamage(s, -5)
	if s.shields != 30 || s.hull != 100 {
		t.Errorf("non-positive damage should be no-op, got shields=%d hull=%d", s.shields, s.hull)
	}

	// Over-kill clamps at 0.
	s = &combatShip{shields: 10, hull: 10}
	applyShipDamage(s, 999)
	if s.hull != 0 {
		t.Errorf("over-kill should clamp hull at 0, got %d", s.hull)
	}

	// Nil ship is a no-op, not a panic.
	applyShipDamage(nil, 50)
}
