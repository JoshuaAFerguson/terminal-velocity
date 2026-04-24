// File: internal/combat/reputation_test.go
// Project: Terminal Velocity
// Description: Tests for the reputation math — specifically the
//   P5C-3 war-zone amplifier. Existing Calculate/Apply helpers are
//   not directly unit-tested here (they're covered indirectly
//   through the amplifier's interaction with ReputationChange).
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package combat

import "testing"

func TestAmplifyReputationChangesNoOps(t *testing.T) {
	tests := []struct {
		name    string
		changes []ReputationChange
		mult    float64
	}{
		{"nil slice is unchanged", nil, 1.5},
		{"empty slice is unchanged", []ReputationChange{}, 1.5},
		{"multiplier 1.0 returns original", []ReputationChange{{FactionID: "x", Amount: 5}}, 1.0},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// No-op paths return the input slice identity — checked
			// via reflect-style direct compare on length + content.
			got := AmplifyReputationChanges(tc.changes, tc.mult)
			if len(got) != len(tc.changes) {
				t.Fatalf("length changed: got %d, want %d", len(got), len(tc.changes))
			}
			for i := range got {
				if got[i] != tc.changes[i] {
					t.Errorf("entry %d mutated: got %+v, want %+v", i, got[i], tc.changes[i])
				}
			}
		})
	}
}

func TestAmplifyReputationChangesScalesPositive(t *testing.T) {
	in := []ReputationChange{
		{FactionID: "uef", Amount: 5, Reason: "kill hostile"},
		{FactionID: "mars", Amount: 10, Reason: "defend ally"},
	}
	got := AmplifyReputationChanges(in, 1.5)
	if got[0].Amount != 7 { // 5 × 1.5 = 7.5 → truncate 7
		t.Errorf("5 × 1.5: got %d, want 7", got[0].Amount)
	}
	if got[1].Amount != 15 { // 10 × 1.5 = 15.0
		t.Errorf("10 × 1.5: got %d, want 15", got[1].Amount)
	}
	// FactionID + Reason flow through unchanged.
	if got[0].FactionID != "uef" || got[0].Reason != "kill hostile" {
		t.Errorf("metadata mangled: %+v", got[0])
	}
}

func TestAmplifyReputationChangesScalesNegative(t *testing.T) {
	in := []ReputationChange{
		{FactionID: "crimson", Amount: -5, Reason: "kill ally"},
		{FactionID: "pirate", Amount: -10, Reason: "pirate action"},
	}
	got := AmplifyReputationChanges(in, 1.5)
	// -5 × 1.5 = -7.5 → truncate toward zero via |abs|, then restore
	// sign → -7 (not -8). The comment on the function documents this.
	if got[0].Amount != -7 {
		t.Errorf("-5 × 1.5: got %d, want -7", got[0].Amount)
	}
	if got[1].Amount != -15 {
		t.Errorf("-10 × 1.5: got %d, want -15", got[1].Amount)
	}
}

func TestAmplifyReputationChangesClampsNegativeMultiplier(t *testing.T) {
	in := []ReputationChange{{FactionID: "x", Amount: 10}}
	// Negative multiplier would flip sign — clamped to 0.
	got := AmplifyReputationChanges(in, -2.0)
	if got[0].Amount != 0 {
		t.Errorf("negative multiplier clamped: got %d, want 0", got[0].Amount)
	}
}

func TestAmplifyReputationChangesDoesNotMutateInput(t *testing.T) {
	in := []ReputationChange{{FactionID: "x", Amount: 10}}
	_ = AmplifyReputationChanges(in, 1.5)
	if in[0].Amount != 10 {
		t.Errorf("input mutated: got %d, want 10 (unchanged)", in[0].Amount)
	}
}

func TestAmplifyReputationChangesZeroAmount(t *testing.T) {
	// 0 × anything = 0. Edge case: |0| path still returns 0, not
	// accidentally -0 or anything weird.
	in := []ReputationChange{{FactionID: "x", Amount: 0}}
	got := AmplifyReputationChanges(in, 1.5)
	if got[0].Amount != 0 {
		t.Errorf("0 × 1.5: got %d, want 0", got[0].Amount)
	}
}

func TestAmplifyThenApplyChain(t *testing.T) {
	// Integration-style: amplify + apply should produce the same
	// result as calling apply on manually-scaled inputs. Smoke-tests
	// that the two helpers compose correctly.
	base := []ReputationChange{
		{FactionID: "uef", Amount: 5},
		{FactionID: "uef", Amount: -2},
	}
	amplified := AmplifyReputationChanges(base, 2.0)
	rep := ApplyReputationChanges(nil, amplified)
	// +10 -4 = +6
	if got := rep["uef"]; got != 6 {
		t.Errorf("chained result: got %d, want 6", got)
	}
}
