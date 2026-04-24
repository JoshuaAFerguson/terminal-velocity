// File: internal/spaceflight/physics.go
// Project: Terminal Velocity
// Description: Real-time space-flight physics for the in-system
//   viewport. Newtonian — thrust accelerates the ship along its
//   facing direction, no auto-deceleration. Stays a pure-Go
//   package with no TUI imports so the math is unit-testable in
//   isolation and we can later run it server-side in a multiplayer
//   tick service without dragging BubbleTea along.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package spaceflight

import "math"

// FlightState is one ship's position + velocity + heading at a point
// in time. World coordinates are floating-point so sub-cell motion
// accumulates smoothly across ticks; the renderer quantizes to
// terminal cells when it draws. Heading is in radians, 0 = +X (east),
// rotating counter-clockwise (math convention, not compass).
//
// Velocity is the only non-input state — heading and position derive
// from cumulative thrust + rotate operations, but velocity persists
// across ticks because there's no friction in space.
type FlightState struct {
	X, Y     float64 // world position
	VX, VY   float64 // velocity components per second
	Heading  float64 // radians; 0 = +X axis, +π/2 = +Y axis (down on screen)
	MaxSpeed float64 // hard cap on |velocity| (prevents accelerating to render-breaking speeds)
}

// NewFlightState returns a ship at the origin, facing "up" on a
// terminal screen. "Up" is -Y in screen coordinates, which maps to
// -π/2 in our math heading. Players expect ↑ to thrust toward the
// top of their screen; this default makes that intuitive.
func NewFlightState() FlightState {
	return FlightState{
		X: 0, Y: 0,
		VX: 0, VY: 0,
		Heading:  -math.Pi / 2, // facing -Y == screen up
		MaxSpeed: 200.0,        // world units per second
	}
}

// ThrustImpulse is one keypress of forward thrust. Applied as a
// velocity delta along the heading direction. Magnitude is tuned for
// arrow-key auto-repeat on a typical terminal (~20-30 events/sec when
// held), giving roughly 3-second 0-to-cruise feel.
const ThrustImpulse = 8.0

// RotateStep is one keypress of rotation. ~10° per press, so
// auto-repeat sweeps the ship around in under a second — fast enough
// to dogfight, slow enough to aim. Stored as radians for consistency
// with Heading.
const RotateStep = math.Pi / 18.0 // 10 degrees

// BrakeImpulse is the magnitude of a "brake" press: applies thrust
// directly opposite to current velocity. Slightly weaker than
// ThrustImpulse because braking is convenient, not free; players
// still need to plan their stops.
const BrakeImpulse = 6.0

// Tick advances the state by `dt` seconds. Pure: returns a new
// FlightState, doesn't mutate the receiver. dt is expected to be
// the wall-clock delta between ticks (typically 50ms = 0.05s for a
// 20Hz loop).
//
// No friction. Velocity persists indefinitely until the player
// thrusts in another direction or brakes. This is the EV feel —
// space has no air, so you drift forever.
func (s FlightState) Tick(dt float64) FlightState {
	s.X += s.VX * dt
	s.Y += s.VY * dt
	return s
}

// Thrust applies one ThrustImpulse along the current heading. Caps
// total speed at MaxSpeed by re-normalizing the velocity vector if
// the resulting magnitude would exceed it — this preserves direction
// (so the ship still goes the way the player aimed) while clamping
// magnitude.
func (s FlightState) Thrust() FlightState {
	s.VX += math.Cos(s.Heading) * ThrustImpulse
	s.VY += math.Sin(s.Heading) * ThrustImpulse
	return s.clampSpeed()
}

// Brake applies thrust opposite to current velocity. If the ship is
// already stopped, no-op (otherwise we'd accelerate from rest in a
// random direction). Magnitude can't overshoot zero — if BrakeImpulse
// would reverse direction, we clamp to exactly zero so a held brake
// reliably comes to a full stop.
func (s FlightState) Brake() FlightState {
	speed := math.Hypot(s.VX, s.VY)
	if speed < 1e-9 {
		return s
	}
	// Direction opposite to current velocity.
	dx := -s.VX / speed
	dy := -s.VY / speed
	delta := BrakeImpulse
	if delta > speed {
		// Would overshoot — clamp to exact stop.
		delta = speed
	}
	s.VX += dx * delta
	s.VY += dy * delta
	return s
}

// RotateLeft rotates heading counter-clockwise by RotateStep.
func (s FlightState) RotateLeft() FlightState {
	s.Heading -= RotateStep
	s.Heading = normalizeAngle(s.Heading)
	return s
}

// RotateRight rotates heading clockwise by RotateStep.
func (s FlightState) RotateRight() FlightState {
	s.Heading += RotateStep
	s.Heading = normalizeAngle(s.Heading)
	return s
}

// Speed returns the current velocity magnitude. Exposed so the HUD
// can display it without re-deriving from VX/VY.
func (s FlightState) Speed() float64 {
	return math.Hypot(s.VX, s.VY)
}

// HeadingDegrees returns Heading converted to compass-style degrees
// (0-360, 0 = north/up, 90 = east/right). Used by HUD labels.
//
// Internal Heading is math-convention radians: 0 = +X (east), +π/2 =
// +Y (south on screen). Compass convention is rotated 90° and flipped:
// 0 = north (-Y), 90 = east (+X). The conversion handles both.
func (s FlightState) HeadingDegrees() int {
	deg := s.Heading*180/math.Pi + 90 // shift so -π/2 (up) becomes 0
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}
	return int(deg)
}

// clampSpeed rescales velocity to MaxSpeed if it exceeds the cap.
// Direction is preserved; only magnitude is clamped.
func (s FlightState) clampSpeed() FlightState {
	speed := math.Hypot(s.VX, s.VY)
	if speed <= s.MaxSpeed {
		return s
	}
	scale := s.MaxSpeed / speed
	s.VX *= scale
	s.VY *= scale
	return s
}

// normalizeAngle wraps an angle into (-π, π]. Keeps Heading from
// drifting to numerically-large values that could lose precision
// after many rotations.
func normalizeAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a <= -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

// HeadingGlyph returns the ASCII/Unicode arrow character that best
// represents the ship's facing in 8 cardinal directions. Used by the
// renderer to draw the player's ship as an oriented arrow at the
// center of the viewport.
//
// Compass convention used here: 0° = north, increasing clockwise.
// We round-to-nearest-octant rather than to-nearest-direction so a
// 22.5° facing renders as N (▲) rather than NE (◥), matching how
// players typically intend small heading adjustments.
func (s FlightState) HeadingGlyph() string {
	// Map heading degrees to one of 8 octants; +22.5° lookup window.
	deg := s.HeadingDegrees()
	octant := ((deg + 22) / 45) % 8
	glyphs := [8]string{
		"▲", // 0   = N
		"◥", // 45  = NE
		"▶", // 90  = E
		"◢", // 135 = SE
		"▼", // 180 = S
		"◣", // 225 = SW
		"◀", // 270 = W
		"◤", // 315 = NW
	}
	return glyphs[octant]
}
