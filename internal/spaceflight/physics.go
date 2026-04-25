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

// FlightParams bundles the per-ship physics constants. Pulled out of
// FlightState so we can keep ship-derived tuning alongside the
// continuously-changing position/velocity, and so callers compose a
// single struct from ship + outfits and hand the whole thing to
// NewFlightState.
//
// All values are floats in world-unit-per-second terms. Tick rate is
// the renderer's concern; physics is rate-independent (Tick takes a
// dt argument).
type FlightParams struct {
	// MaxSpeed is the hard cap on |velocity|. Slow freighters cap
	// around 90 u/s; agile interceptors approach 360.
	MaxSpeed float64

	// ThrustImpulse is the velocity delta added to vx/vy each time
	// the player presses thrust. With terminal key-repeat at
	// ~30Hz and a 60fps physics tick, this needs to land somewhere
	// in the 1-3 range to feel responsive without overshooting on
	// every press.
	ThrustImpulse float64

	// RotateStep is the heading delta per rotate press, in radians.
	// Typical: π/12 (15°) for nimble ships, π/24 (7.5°) for slow.
	RotateStep float64

	// BrakeImpulse is the thrust opposite to current velocity when
	// the player brakes. Roughly 75% of ThrustImpulse so braking
	// feels useful without being instantly free.
	BrakeImpulse float64
}

// DefaultFlightParams returns starter-ship-grade flight tuning. Used
// when no ship is equipped (login screen, pre-ship registration) —
// keeps the cockpit demo-able even before a real Ship is loaded.
func DefaultFlightParams() FlightParams {
	return FlightParams{
		MaxSpeed:      150.0,
		ThrustImpulse: 1.5,
		RotateStep:    math.Pi / 18.0, // 10°
		BrakeImpulse:  1.2,
	}
}

// FlightParamsFromShipStats derives flight characteristics from a
// ship's combat-stat fields. Inputs match models.ShipType ranges:
//
//	speed:          1-10  (existing combat initiative scale)
//	maneuverability: 2-12 (existing combat evasion scale)
//
// The mapping is intentionally generous — players should feel a
// clear difference flying a Shuttle (Speed=2) vs. an Interceptor
// (Speed=10). Tuning room left for engine outfits to push values
// further once equipment integration lands in P1.3+.
func FlightParamsFromShipStats(speed, maneuverability int) FlightParams {
	if speed < 1 {
		speed = 1
	}
	if maneuverability < 1 {
		maneuverability = 1
	}
	// Speed → MaxSpeed: 90 (Speed=1) to 360 (Speed=10).
	maxSpeed := 60.0 + float64(speed)*30.0
	// Speed → ThrustImpulse: 1.0 (slow) to 2.5 (peppy). Acceleration
	// matters more than top speed for combat feel, so we keep this
	// range tight.
	thrust := 0.8 + float64(speed)*0.17
	// Maneuverability → RotateStep: 5° (capital ships) to 20°
	// (interceptors).
	rotateDeg := 4.0 + float64(maneuverability)*1.3
	rotate := rotateDeg * math.Pi / 180.0
	return FlightParams{
		MaxSpeed:      maxSpeed,
		ThrustImpulse: thrust,
		RotateStep:    rotate,
		BrakeImpulse:  thrust * 0.8,
	}
}

// FlightState is one ship's position + velocity + heading + tuning
// at a point in time. World coordinates are floating-point so
// sub-cell motion accumulates smoothly across ticks; the renderer
// quantizes to terminal cells when it draws. Heading is in radians,
// 0 = +X (east), rotating counter-clockwise (math convention, not
// compass).
//
// Velocity persists across ticks (no friction). FlightParams are
// embedded by-value rather than via pointer so a server-authoritative
// snapshot can be passed to clients without the params field
// becoming a shared aliased reference.
type FlightState struct {
	X, Y    float64 // world position
	VX, VY  float64 // velocity components per second
	Heading float64 // radians; 0 = +X axis, +π/2 = +Y axis (down on screen)
	Params  FlightParams
}

// NewFlightState returns a ship at the origin, facing "up" on a
// terminal screen, with the given physics parameters. "Up" is -Y in
// screen coordinates, which maps to -π/2 in our math heading.
// Players expect ↑ to thrust toward the top of their screen; this
// default makes that intuitive.
func NewFlightState(params FlightParams) FlightState {
	return FlightState{
		X: 0, Y: 0,
		VX: 0, VY: 0,
		Heading: -math.Pi / 2, // facing -Y == screen up
		Params:  params,
	}
}

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
	s.VX += math.Cos(s.Heading) * s.Params.ThrustImpulse
	s.VY += math.Sin(s.Heading) * s.Params.ThrustImpulse
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
	delta := s.Params.BrakeImpulse
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
	s.Heading -= s.Params.RotateStep
	s.Heading = normalizeAngle(s.Heading)
	return s
}

// RotateRight rotates heading clockwise by RotateStep.
func (s FlightState) RotateRight() FlightState {
	s.Heading += s.Params.RotateStep
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
	if speed <= s.Params.MaxSpeed {
		return s
	}
	scale := s.Params.MaxSpeed / speed
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
