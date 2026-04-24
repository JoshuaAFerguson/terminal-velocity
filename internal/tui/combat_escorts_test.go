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
// eyeball the strip with `go test -v`.
func TestRenderEscortStripPreview(t *testing.T) {
	escorts := []combatEscort{
		{pilot: "Kira Rehn", shipType: "Interceptor", behavior: fleet.BehaviorAggressive, level: 3},
		{pilot: "Maz Oort", shipType: "Corvette", behavior: fleet.BehaviorDefensive, level: 5},
		{pilot: "Drex", shipType: "Gunship", behavior: fleet.BehaviorSupport, level: 7},
	}
	t.Logf("escort strip (width=70):\n%s", renderEscortStrip(escorts, 70))
}
