// File: internal/spaceflight/physics_test.go
// Project: Terminal Velocity
// Description: Tests for the flight-physics math. Pure-state-in,
//   pure-state-out — no time mocking, no random sources, no I/O.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package spaceflight

import (
	"math"
	"testing"
)

const tolerance = 1e-9

func approxEq(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestNewFlightStateDefaults(t *testing.T) {
	s := NewFlightState()
	if s.X != 0 || s.Y != 0 {
		t.Errorf("position should start at origin, got (%v, %v)", s.X, s.Y)
	}
	if s.VX != 0 || s.VY != 0 {
		t.Errorf("velocity should start at zero, got (%v, %v)", s.VX, s.VY)
	}
	// Default heading is "screen up", which is -π/2 in our math frame.
	if !approxEq(s.Heading, -math.Pi/2) {
		t.Errorf("default heading should be -π/2, got %v", s.Heading)
	}
	if s.MaxSpeed <= 0 {
		t.Errorf("MaxSpeed should be positive, got %v", s.MaxSpeed)
	}
}

func TestTickAdvancesPosition(t *testing.T) {
	s := FlightState{X: 0, Y: 0, VX: 10, VY: 5, MaxSpeed: 100}
	advanced := s.Tick(0.5)
	if !approxEq(advanced.X, 5) {
		t.Errorf("X after 0.5s at vx=10: got %v, want 5", advanced.X)
	}
	if !approxEq(advanced.Y, 2.5) {
		t.Errorf("Y after 0.5s at vy=5: got %v, want 2.5", advanced.Y)
	}
	// Tick is pure — original is unchanged.
	if s.X != 0 || s.Y != 0 {
		t.Errorf("original state mutated")
	}
}

func TestTickPreservesVelocityWithoutFriction(t *testing.T) {
	// Newtonian: no thrust, no friction → velocity persists across
	// many ticks. Verifies we haven't accidentally added drag.
	s := FlightState{X: 0, Y: 0, VX: 10, VY: 0, MaxSpeed: 100}
	for i := 0; i < 100; i++ {
		s = s.Tick(0.05)
	}
	if !approxEq(s.VX, 10) {
		t.Errorf("velocity should persist: got vx=%v, want 10", s.VX)
	}
	if !approxEq(s.X, 50) {
		t.Errorf("position after 100 ticks: got %v, want 50", s.X)
	}
}

func TestThrustAddsVelocityAlongHeading(t *testing.T) {
	// Heading = 0 (east, +X). Thrust should add ThrustImpulse to vx,
	// nothing to vy.
	s := FlightState{Heading: 0, MaxSpeed: 1000}
	thrusted := s.Thrust()
	if !approxEq(thrusted.VX, ThrustImpulse) {
		t.Errorf("thrust east: got vx=%v, want %v", thrusted.VX, ThrustImpulse)
	}
	if !approxEq(thrusted.VY, 0) {
		t.Errorf("thrust east: got vy=%v, want 0", thrusted.VY)
	}

	// Heading = π/2 (south, +Y). Thrust should add ThrustImpulse to vy.
	s = FlightState{Heading: math.Pi / 2, MaxSpeed: 1000}
	thrusted = s.Thrust()
	if !approxEq(thrusted.VX, 0) {
		t.Errorf("thrust south: got vx=%v, want 0", thrusted.VX)
	}
	if !approxEq(thrusted.VY, ThrustImpulse) {
		t.Errorf("thrust south: got vy=%v, want %v", thrusted.VY, ThrustImpulse)
	}
}

func TestThrustAccumulates(t *testing.T) {
	// Multiple thrusts in same direction should accumulate velocity
	// linearly until clamped.
	s := FlightState{Heading: 0, MaxSpeed: 1000}
	for i := 0; i < 5; i++ {
		s = s.Thrust()
	}
	want := 5 * ThrustImpulse
	if !approxEq(s.VX, want) {
		t.Errorf("5 thrusts: got vx=%v, want %v", s.VX, want)
	}
}

func TestThrustClampsToMaxSpeed(t *testing.T) {
	// MaxSpeed=10, ThrustImpulse=8, two thrusts → would be 16 but
	// clamps to 10 in the same direction.
	s := FlightState{Heading: 0, MaxSpeed: 10}
	s = s.Thrust().Thrust()
	if !approxEq(s.Speed(), 10) {
		t.Errorf("speed should clamp at MaxSpeed: got %v, want 10", s.Speed())
	}
	// Direction preserved.
	if !approxEq(s.VX, 10) || !approxEq(s.VY, 0) {
		t.Errorf("direction not preserved on clamp: (%v, %v)", s.VX, s.VY)
	}
}

func TestBrakeReducesSpeed(t *testing.T) {
	s := FlightState{VX: 20, VY: 0, MaxSpeed: 1000}
	braked := s.Brake()
	if braked.Speed() >= 20 {
		t.Errorf("brake should reduce speed: got %v", braked.Speed())
	}
	// Brake direction is opposite to velocity, so VX decreases.
	if braked.VX >= 20 {
		t.Errorf("brake: vx should decrease, got %v", braked.VX)
	}
}

func TestBrakeOnZeroVelocityIsNoop(t *testing.T) {
	// Otherwise we'd accelerate from rest in some arbitrary direction.
	s := FlightState{MaxSpeed: 1000}
	braked := s.Brake()
	if !approxEq(braked.VX, 0) || !approxEq(braked.VY, 0) {
		t.Errorf("brake on stationary should be no-op: got (%v, %v)", braked.VX, braked.VY)
	}
}

func TestBrakeClampsToFullStop(t *testing.T) {
	// If BrakeImpulse > current speed, we should land at exactly 0,
	// not overshoot into reverse motion.
	s := FlightState{VX: 1, VY: 0, MaxSpeed: 1000}
	braked := s.Brake()
	if !approxEq(braked.VX, 0) || !approxEq(braked.VY, 0) {
		t.Errorf("brake should clamp at 0: got (%v, %v)", braked.VX, braked.VY)
	}
}

func TestRotateLeftRight(t *testing.T) {
	s := FlightState{Heading: 0}
	left := s.RotateLeft()
	if !approxEq(left.Heading, -RotateStep) {
		t.Errorf("rotate-left: got %v, want %v", left.Heading, -RotateStep)
	}
	right := s.RotateRight()
	if !approxEq(right.Heading, RotateStep) {
		t.Errorf("rotate-right: got %v, want %v", right.Heading, RotateStep)
	}
}

func TestRotateNormalizesAngle(t *testing.T) {
	// Spam rotations to confirm Heading stays bounded in (-π, π].
	s := FlightState{Heading: 0}
	for i := 0; i < 1000; i++ {
		s = s.RotateRight()
	}
	if s.Heading > math.Pi || s.Heading <= -math.Pi {
		t.Errorf("heading not normalized after many rotations: got %v", s.Heading)
	}
}

func TestHeadingDegreesCardinalDirections(t *testing.T) {
	tests := []struct {
		name    string
		heading float64
		want    int
	}{
		{"up (-π/2)", -math.Pi / 2, 0},
		{"right (0)", 0, 90},
		{"down (π/2)", math.Pi / 2, 180},
		{"left (π)", math.Pi, 270},
	}
	for _, tc := range tests {
		s := FlightState{Heading: tc.heading}
		if got := s.HeadingDegrees(); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestHeadingGlyphCardinalDirections(t *testing.T) {
	// Verifies the 8-octant glyph mapping at exact cardinal headings.
	tests := []struct {
		name    string
		heading float64
		want    string
	}{
		{"up", -math.Pi / 2, "▲"},
		{"right", 0, "▶"},
		{"down", math.Pi / 2, "▼"},
		{"left", math.Pi, "◀"},
	}
	for _, tc := range tests {
		s := FlightState{Heading: tc.heading}
		if got := s.HeadingGlyph(); got != tc.want {
			t.Errorf("%s heading: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSpeedReportsMagnitude(t *testing.T) {
	s := FlightState{VX: 3, VY: 4}
	if !approxEq(s.Speed(), 5) {
		t.Errorf("Pythagorean: got %v, want 5", s.Speed())
	}
}

func TestThrustThenTickIntegration(t *testing.T) {
	// Smoke test of the three operations composing as expected:
	// rotate up, thrust, tick — should move the ship in -Y.
	s := NewFlightState() // facing -Y (up)
	s = s.Thrust()        // velocity now in -Y
	if s.VY >= 0 {
		t.Fatalf("thrust while facing up should give negative VY, got %v", s.VY)
	}
	moved := s.Tick(1.0)
	if moved.Y >= 0 {
		t.Errorf("after 1s of upward thrust, Y should be negative, got %v", moved.Y)
	}
}
