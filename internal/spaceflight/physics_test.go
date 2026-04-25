// File: internal/spaceflight/physics_test.go
// Project: Terminal Velocity
// Description: Tests for the flight-physics math. Pure-state-in,
//   pure-state-out — no time mocking, no random sources, no I/O.
// Version: 1.1.0
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

// testParams is a deterministic FlightParams fixture for tests that
// need predictable thrust/rotate magnitudes. Values are chosen so
// that a single thrust gives integer-friendly velocity deltas.
var testParams = FlightParams{
	MaxSpeed:      1000,
	ThrustImpulse: 8,
	RotateStep:    math.Pi / 18, // 10°
	BrakeImpulse:  6,
}

func TestNewFlightStateDefaults(t *testing.T) {
	s := NewFlightState(DefaultFlightParams())
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
	if s.Params.MaxSpeed <= 0 {
		t.Errorf("MaxSpeed should be positive, got %v", s.Params.MaxSpeed)
	}
}

func TestTickAdvancesPosition(t *testing.T) {
	s := FlightState{X: 0, Y: 0, VX: 10, VY: 5, Params: testParams}
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
	s := FlightState{X: 0, Y: 0, VX: 10, VY: 0, Params: testParams}
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
	s := FlightState{Heading: 0, Params: testParams}
	thrusted := s.Thrust()
	if !approxEq(thrusted.VX, testParams.ThrustImpulse) {
		t.Errorf("thrust east: got vx=%v, want %v", thrusted.VX, testParams.ThrustImpulse)
	}
	if !approxEq(thrusted.VY, 0) {
		t.Errorf("thrust east: got vy=%v, want 0", thrusted.VY)
	}

	// Heading = π/2 (south, +Y). Thrust should add ThrustImpulse to vy.
	s = FlightState{Heading: math.Pi / 2, Params: testParams}
	thrusted = s.Thrust()
	if !approxEq(thrusted.VX, 0) {
		t.Errorf("thrust south: got vx=%v, want 0", thrusted.VX)
	}
	if !approxEq(thrusted.VY, testParams.ThrustImpulse) {
		t.Errorf("thrust south: got vy=%v, want %v", thrusted.VY, testParams.ThrustImpulse)
	}
}

func TestThrustAccumulates(t *testing.T) {
	// Multiple thrusts in same direction should accumulate velocity
	// linearly until clamped.
	s := FlightState{Heading: 0, Params: testParams}
	for i := 0; i < 5; i++ {
		s = s.Thrust()
	}
	want := 5 * testParams.ThrustImpulse
	if !approxEq(s.VX, want) {
		t.Errorf("5 thrusts: got vx=%v, want %v", s.VX, want)
	}
}

func TestThrustClampsToMaxSpeed(t *testing.T) {
	// MaxSpeed=10, ThrustImpulse=8, two thrusts → would be 16 but
	// clamps to 10 in the same direction.
	tight := FlightParams{MaxSpeed: 10, ThrustImpulse: 8, RotateStep: 0.1, BrakeImpulse: 6}
	s := FlightState{Heading: 0, Params: tight}
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
	s := FlightState{VX: 20, VY: 0, Params: testParams}
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
	s := FlightState{Params: testParams}
	braked := s.Brake()
	if !approxEq(braked.VX, 0) || !approxEq(braked.VY, 0) {
		t.Errorf("brake on stationary should be no-op: got (%v, %v)", braked.VX, braked.VY)
	}
}

func TestBrakeClampsToFullStop(t *testing.T) {
	// If BrakeImpulse > current speed, we should land at exactly 0,
	// not overshoot into reverse motion.
	s := FlightState{VX: 1, VY: 0, Params: testParams}
	braked := s.Brake()
	if !approxEq(braked.VX, 0) || !approxEq(braked.VY, 0) {
		t.Errorf("brake should clamp at 0: got (%v, %v)", braked.VX, braked.VY)
	}
}

func TestRotateLeftRight(t *testing.T) {
	s := FlightState{Heading: 0, Params: testParams}
	left := s.RotateLeft()
	if !approxEq(left.Heading, -testParams.RotateStep) {
		t.Errorf("rotate-left: got %v, want %v", left.Heading, -testParams.RotateStep)
	}
	right := s.RotateRight()
	if !approxEq(right.Heading, testParams.RotateStep) {
		t.Errorf("rotate-right: got %v, want %v", right.Heading, testParams.RotateStep)
	}
}

func TestRotateNormalizesAngle(t *testing.T) {
	// Spam rotations to confirm Heading stays bounded in (-π, π].
	s := FlightState{Heading: 0, Params: testParams}
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
	s := NewFlightState(testParams) // facing -Y (up)
	s = s.Thrust()                  // velocity now in -Y
	if s.VY >= 0 {
		t.Fatalf("thrust while facing up should give negative VY, got %v", s.VY)
	}
	moved := s.Tick(1.0)
	if moved.Y >= 0 {
		t.Errorf("after 1s of upward thrust, Y should be negative, got %v", moved.Y)
	}
}

// TestFlightParamsFromShipStats ensures the ship → flight-params
// mapping monotonically rewards higher Speed/Maneuverability.
// Specific output values are tunables, so we don't lock them down
// — we lock down the ordering.
func TestFlightParamsFromShipStats(t *testing.T) {
	slow := FlightParamsFromShipStats(2, 4)   // Shuttle-grade
	fast := FlightParamsFromShipStats(10, 12) // Interceptor-grade

	if slow.MaxSpeed >= fast.MaxSpeed {
		t.Errorf("higher Speed should yield higher MaxSpeed: slow=%v fast=%v",
			slow.MaxSpeed, fast.MaxSpeed)
	}
	if slow.ThrustImpulse >= fast.ThrustImpulse {
		t.Errorf("higher Speed should yield higher ThrustImpulse: slow=%v fast=%v",
			slow.ThrustImpulse, fast.ThrustImpulse)
	}
	if slow.RotateStep >= fast.RotateStep {
		t.Errorf("higher Maneuverability should yield bigger RotateStep: slow=%v fast=%v",
			slow.RotateStep, fast.RotateStep)
	}
	// BrakeImpulse should track ThrustImpulse but be slightly lower.
	if fast.BrakeImpulse >= fast.ThrustImpulse {
		t.Errorf("BrakeImpulse should be < ThrustImpulse: brake=%v thrust=%v",
			fast.BrakeImpulse, fast.ThrustImpulse)
	}
}

// TestFlightParamsFromShipStatsClampsLowInputs ensures zero or
// negative ship stats don't produce zero or negative physics
// constants — would otherwise lock the player on zero MaxSpeed
// or render Thrust a no-op.
func TestFlightParamsFromShipStatsClampsLowInputs(t *testing.T) {
	for _, in := range []int{0, -1, -100} {
		p := FlightParamsFromShipStats(in, in)
		if p.MaxSpeed <= 0 || p.ThrustImpulse <= 0 || p.RotateStep <= 0 {
			t.Errorf("input %d produced non-positive params: %+v", in, p)
		}
	}
}
