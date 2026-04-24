// File: internal/tui/flight.go
// Project: Terminal Velocity
// Description: Real-time flight cockpit screen. Replaces the old
//   static "space view" radar with an actual EV-style 2D viewport
//   the player flies their ship through. Phase 1.1 of the
//   real-time-flight redesign — solo-only, no NPCs, no projectiles
//   yet (those land in 1.3 and 2.1 respectively).
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/spaceflight"
	tea "github.com/charmbracelet/bubbletea"
)

// flightTickInterval is how often we advance physics and re-render.
// 50ms = 20fps. Higher rates (e.g. 30 or 60fps) feel smoother but
// burn more bandwidth over SSH/WebSocket — at 80x24 cells, even a
// fully-redrawn frame is only ~2KB so 20fps is conservative-friendly
// on slow links and tight enough that movement feels live.
const flightTickInterval = 50 * time.Millisecond

// flightTickMsg fires every flightTickInterval while ScreenFlight is
// active. Drives the physics step + re-render. Self-terminates: the
// top-level Update routes flight ticks away when the player exits
// the screen, breaking the re-arm loop the same way the news ticker
// pattern works.
type flightTickMsg struct{}

func flightTick() tea.Cmd {
	return tea.Tick(flightTickInterval, func(time.Time) tea.Msg { return flightTickMsg{} })
}

// flightModel owns the per-session flight state. The ship is stored
// here, but the same struct will eventually carry NPC ships and
// projectiles in this system as well. Server-authoritative state
// for multiplayer (P3.1) will live elsewhere; this is the local
// view's source of truth for now.
type flightModel struct {
	ship    spaceflight.FlightState
	active  bool      // true while the tick loop is running
	lastTick time.Time // wall-clock of last tick (for variable-dt physics)
}

func newFlightModel() flightModel {
	return flightModel{
		ship: spaceflight.NewFlightState(),
	}
}

// updateFlight handles input + tick events. Tick advances physics by
// the wall-clock delta since the last tick (variable-dt) so a 200ms
// hiccup doesn't compound into "ship suddenly across the system" —
// only one frame's worth of motion is missed.
func (m Model) updateFlight(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Self-start the tick loop on first entry. Subsequent entries
	// (after returning from a sub-screen, etc.) re-arm here too —
	// the top-level dispatcher drops stale flightTickMsg arrivals
	// when the screen isn't ScreenFlight, so the loop naturally
	// terminates on exit and restarts on re-entry without any
	// global goroutines.
	var kicker tea.Cmd
	if !m.flight.active {
		m.flight.active = true
		m.flight.lastTick = time.Now()
		kicker = flightTick()
	}

	switch msg := msg.(type) {
	case flightTickMsg:
		now := time.Now()
		dt := now.Sub(m.flight.lastTick).Seconds()
		// Cap dt to avoid huge jumps after a pause/resume — better
		// to under-simulate one frame than to teleport the ship.
		if dt > 0.25 {
			dt = 0.25
		}
		m.flight.ship = m.flight.ship.Tick(dt)
		m.flight.lastTick = now
		// Re-arm. The top-level Update drops flightTickMsg when the
		// active screen isn't ScreenFlight, which kills the loop
		// when the player navigates away.
		return m, flightTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			// Leave flight. lastTick stays as-is so velocity is
			// preserved for next entry — the ship doesn't reset
			// to origin every time the player opens a menu.
			m.flight.active = false
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "w", "W":
			m.flight.ship = m.flight.ship.Thrust()
		case "down", "s", "S":
			m.flight.ship = m.flight.ship.Brake()
		case "left", "a", "A":
			m.flight.ship = m.flight.ship.RotateLeft()
		case "right", "d", "D":
			m.flight.ship = m.flight.ship.RotateRight()

		case "j", "J":
			// Shortcut to the navigation/jump screen, matches the
			// old space view's binding so existing players' muscle
			// memory keeps working.
			m.previousScreen = ScreenFlight
			m.hasPreviousScreen = true
			m.screen = ScreenNavigation
			return m, nil
		}
	}

	if kicker != nil {
		return m, kicker
	}
	return m, nil
}

// viewFlight draws the cockpit: a scrolling viewport with the ship
// in the center, a parallax star field, and a HUD with speed +
// heading + key reminders.
//
// World→screen mapping: viewport center = ship position. World
// coordinates are floating-point; we round-to-nearest-cell when
// drawing. Stars are deterministic (hash of world coords + a small
// offset) so the field is consistent as the ship flies through it.
func (m Model) viewFlight() string {
	width, height := m.flightViewportSize()

	// Build the viewport as a 2D rune grid for easy random-access
	// writes. strings.Builder would force linear assembly which
	// makes "place X at row R, col C" awkward.
	grid := make([][]rune, height)
	for r := range grid {
		grid[r] = make([]rune, width)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}

	drawStarfield(grid, m.flight.ship.X, m.flight.ship.Y)

	// Player ship at center.
	cx, cy := width/2, height/2
	if cy < len(grid) && cx < len(grid[cy]) {
		// Ship glyph is potentially multi-byte; we store the rune
		// so PadRight + cellWidth handle it correctly on render.
		// Heading→glyph mapping lives on the physics struct.
		shipGlyphRunes := []rune(m.flight.ship.HeadingGlyph())
		if len(shipGlyphRunes) > 0 {
			grid[cy][cx] = shipGlyphRunes[0]
		}
	}

	// Compose grid into the final string.
	var sb strings.Builder
	for _, row := range grid {
		sb.WriteString(string(row))
		sb.WriteByte('\n')
	}

	hud := fmt.Sprintf(
		" SPD %5.1f  HDG %3d°  POS (%6.0f, %6.0f) ",
		m.flight.ship.Speed(),
		m.flight.ship.HeadingDegrees(),
		m.flight.ship.X, m.flight.ship.Y,
	)
	help := " W/↑ thrust  •  S/↓ brake  •  A/← turn left  •  D/→ turn right  •  J jump  •  ESC menu "

	// HUD row + help row sit below the viewport.
	sb.WriteString(highlightStyle.Render(PadRight(hud, width)))
	sb.WriteByte('\n')
	sb.WriteString(helpStyle.Render(PadRight(help, width)))

	return sb.String()
}

// flightViewportSize returns the cell dimensions of the play area.
// Two rows are reserved for the HUD/help footer below the grid; the
// rest goes to the playfield. Falls back to a reasonable 80x22 when
// terminal size hasn't been reported yet.
func (m Model) flightViewportSize() (width, height int) {
	width = 80
	height = 22
	if m.width > width {
		width = m.width
	}
	if m.height > height+2 {
		height = m.height - 2
	}
	return width, height
}

// drawStarfield places stars at deterministic positions based on
// world coordinates. The same world cell always shows the same
// pattern, so stars don't shimmer as the ship flies past — they
// scroll by, which is what gives the parallax sensation.
//
// Density is sparse-ish (~3% of cells get a star) so the field
// feels deep without overwhelming actual game entities (planets,
// ships) we'll add in later phases.
func drawStarfield(grid [][]rune, shipX, shipY float64) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	height := len(grid)
	width := len(grid[0])
	// Top-left world coordinate of the viewport.
	worldOriginX := int(shipX) - width/2
	worldOriginY := int(shipY) - height/2

	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			wx := worldOriginX + c
			wy := worldOriginY + r
			// Cheap hash of world coords. The mod ratio sets star
			// density (~3% of cells); the second hash level picks
			// which glyph (faint dot vs brighter star) so the field
			// has visual texture instead of one repeating char.
			h := starHash(wx, wy)
			if h%32 == 0 {
				if h%128 == 0 {
					grid[r][c] = '*'
				} else {
					grid[r][c] = '·'
				}
			}
		}
	}
}

// starHash is a small integer hash over (x, y). Output quality
// matters less than determinism — same input → same star. xorshift-
// flavored mix is fine; we don't need cryptographic distribution.
func starHash(x, y int) uint32 {
	h := uint32(x*73856093) ^ uint32(y*19349663)
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return h
}
