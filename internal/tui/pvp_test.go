// File: internal/tui/pvp_test.go
// Project: Terminal Velocity
// Description: Unit tests for PvP challenger-side auto-transition helper.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tui

import (
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/pvp"
	"github.com/google/uuid"
)

func TestFindChallengerActiveChallenge(t *testing.T) {
	me := uuid.New()
	other := uuid.New()
	third := uuid.New()

	mk := func(challenger, defender uuid.UUID, status models.PvPChallengeStatus) *models.PvPChallenge {
		return &models.PvPChallenge{
			ID:           uuid.New(),
			ChallengerID: challenger,
			DefenderID:   defender,
			DefenderName: "target",
			Status:       status,
		}
	}

	tests := []struct {
		name       string
		challenges []*models.PvPChallenge
		wantMatch  bool
	}{
		{
			name:       "empty list returns nil",
			challenges: nil,
			wantMatch:  false,
		},
		{
			name: "pending outbound does not transition",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengePending),
			},
			wantMatch: false,
		},
		{
			name: "active outbound triggers transition",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengeActive),
			},
			wantMatch: true,
		},
		{
			name: "active inbound (I am defender) ignored",
			challenges: []*models.PvPChallenge{
				mk(other, me, models.ChallengeActive),
			},
			wantMatch: false,
		},
		{
			name: "completed outbound ignored (already fought)",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengeComplete),
			},
			wantMatch: false,
		},
		{
			name: "declined outbound ignored",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengeDeclined),
			},
			wantMatch: false,
		},
		{
			name: "expired outbound ignored",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengeExpired),
			},
			wantMatch: false,
		},
		{
			name: "nil entry skipped, active outbound after it still matches",
			challenges: []*models.PvPChallenge{
				nil,
				mk(me, third, models.ChallengeActive),
			},
			wantMatch: true,
		},
		{
			name: "multiple active outbound returns first",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengeActive),
				mk(me, third, models.ChallengeActive),
			},
			wantMatch: true,
		},
		{
			name: "mix: pending outbound before active outbound — still returns the active one",
			challenges: []*models.PvPChallenge{
				mk(me, other, models.ChallengePending),
				mk(me, third, models.ChallengeActive),
			},
			wantMatch: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := findChallengerActiveChallenge(tc.challenges, me)
			if tc.wantMatch && got == nil {
				t.Fatalf("expected a matching challenge, got nil")
			}
			if !tc.wantMatch && got != nil {
				t.Fatalf("expected no match, got challenge id=%s status=%s", got.ID, got.Status)
			}
			if tc.wantMatch && got != nil {
				if got.ChallengerID != me {
					t.Fatalf("matched challenge has wrong challenger: got %s want %s", got.ChallengerID, me)
				}
				if got.Status != models.ChallengeActive {
					t.Fatalf("matched challenge has wrong status: got %s want %s", got.Status, models.ChallengeActive)
				}
			}
		})
	}
}

// TestChallengerAutoTransitionFlow verifies the helper works against the
// real pvp.Manager accept path, which flips Pending → Accepted → Active
// atomically. The helper MUST see Active (not Accepted) because Start()
// runs under the same lock as Accept().
func TestChallengerAutoTransitionFlow(t *testing.T) {
	mgr := pvp.NewManager()
	challenger := uuid.New()
	defender := uuid.New()
	system := uuid.New()

	ch, err := mgr.CreateChallenge(
		challenger, "alice",
		defender, "bob",
		models.ChallengeDuel,
		system, 0, "en garde",
	)
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	// Before acceptance, challenger should not auto-transition.
	if got := findChallengerActiveChallenge(mgr.GetPlayerChallenges(challenger), challenger); got != nil {
		t.Fatalf("pre-accept: expected no active challenge for challenger, got %s", got.Status)
	}

	if err := mgr.AcceptChallenge(ch.ID, defender); err != nil {
		t.Fatalf("AcceptChallenge: %v", err)
	}

	// After acceptance, challenger should auto-transition.
	got := findChallengerActiveChallenge(mgr.GetPlayerChallenges(challenger), challenger)
	if got == nil {
		t.Fatalf("post-accept: expected active challenge for challenger, got nil")
	}
	if got.ID != ch.ID {
		t.Fatalf("post-accept: matched wrong challenge id: got %s want %s", got.ID, ch.ID)
	}
	if got.Status != models.ChallengeActive {
		t.Fatalf("post-accept: expected status Active, got %s", got.Status)
	}

	// Defender side MUST NOT auto-transition via this helper — they
	// already transitioned synchronously via the 'a' keypress path.
	if got := findChallengerActiveChallenge(mgr.GetPlayerChallenges(defender), defender); got != nil {
		t.Fatalf("defender should not match challenger helper, got %s", got.Status)
	}
}
