// File: internal/models/faction_war.go
// Project: Terminal Velocity
// Description: FactionWar model describing an active or resolved
//   conflict between two NPC factions. See docs/FACTION_RELATIONS.md
//   §"Faction War Mechanics" for the design spec this implements.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package models

import (
	"time"

	"github.com/google/uuid"
)

// FactionWarStatus tracks where a war is in its lifecycle. Wars
// progress Active → (Resolved | Ceased); once terminal, status never
// flips back — resolving the same war twice is a no-op in the
// manager.
type FactionWarStatus string

const (
	// FactionWarActive: war is declared and ongoing. War zones are
	// hot, reputation deltas amplified, war missions available.
	FactionWarActive FactionWarStatus = "active"

	// FactionWarResolved: one side has won; WinnerFactionID is set.
	FactionWarResolved FactionWarStatus = "resolved"

	// FactionWarCeased: mutual cease-fire with no winner. Used when
	// both sides negotiate, or an admin intervenes.
	FactionWarCeased FactionWarStatus = "ceased"
)

// FactionWar records a conflict between two NPC factions. The
// AggressorID / DefenderID reference NPCFaction.ID (string slug, not
// UUID — NPC factions have stable slugs like "united_earth_federation").
//
// WarZoneSystems is snapshotted at declaration time so that later
// border shifts don't retroactively change what was a hot zone; if
// new systems should be added to the war (e.g. a frontier that
// flipped), that's an admin/resolver decision captured as a new war
// record rather than a mutation.
type FactionWar struct {
	ID            uuid.UUID `json:"id"`
	AggressorID   string    `json:"aggressor_id"`
	AggressorName string    `json:"aggressor_name"`
	DefenderID    string    `json:"defender_id"`
	DefenderName  string    `json:"defender_name"`

	Status          FactionWarStatus `json:"status"`
	DeclaredAt      time.Time        `json:"declared_at"`
	ResolvedAt      *time.Time       `json:"resolved_at,omitempty"`
	WinnerFactionID string           `json:"winner_faction_id,omitempty"` // empty on ceasefire

	// WarZoneSystems are the system names flagged as hot at
	// declaration. Snapshot copy (not a reference to the faction
	// struct) so future territory shifts don't rewrite history.
	WarZoneSystems []string `json:"war_zone_systems"`

	// CasusBelli is free-text, displayed in news articles and the
	// war-history UI. Players and admins read this; it's not
	// parsed by any other system.
	CasusBelli string `json:"casus_belli"`
}

// IsActive returns true if this war is still in its Active phase.
// Terminal states (Resolved, Ceased) return false — a resolved war
// doesn't create war zones anymore.
func (w *FactionWar) IsActive() bool {
	return w != nil && w.Status == FactionWarActive
}

// Duration returns how long the war has been (or was) running.
// For active wars, measured from declaration to now; for
// resolved/ceased, the declared-to-resolved span.
func (w *FactionWar) Duration(now time.Time) time.Duration {
	if w == nil {
		return 0
	}
	end := now
	if w.ResolvedAt != nil {
		end = *w.ResolvedAt
	}
	return end.Sub(w.DeclaredAt)
}
