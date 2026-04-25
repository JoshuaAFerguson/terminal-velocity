// File: internal/tui/flight.go
// Project: Terminal Velocity
// Description: Real-time flight cockpit screen. Multi-pane layout
//   with viewport in the center, ship status sidebar on the right,
//   target/system info under the viewport, header at the top, and
//   key reminders at the bottom. Replaces the static "space view"
//   radar as the primary 'in your ship' screen.
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/spaceflight"
	tea "github.com/charmbracelet/bubbletea"
)

// flightTickInterval is how often we advance physics and re-render.
// 16ms ≈ 60fps. Higher than 20fps because key-repeat from a terminal
// emits events at ~30Hz, and a slower physics tick made each rotate
// or thrust press feel laggy. At 80x24 cells differential redraw is
// ~1KB/frame, which is fine over the WebSocket or SSH.
const flightTickInterval = 16 * time.Millisecond

// flightTickMsg fires every flightTickInterval while ScreenFlight is
// active. Drives the physics step + re-render. Self-terminates: the
// top-level Update routes flight ticks away when the player exits
// the screen, breaking the re-arm loop.
type flightTickMsg struct{}

func flightTick() tea.Cmd {
	return tea.Tick(flightTickInterval, func(time.Time) tea.Msg { return flightTickMsg{} })
}

// flightModel owns the per-session flight state. Storing the ship
// here means velocity persists across screen visits — a player who
// pops out to check trade prices doesn't reset their inertia.
type flightModel struct {
	ship       spaceflight.FlightState
	active     bool      // true while the tick loop is running
	lastTick   time.Time // wall-clock of last tick (variable-dt physics)
	initialized bool     // true after first sync from the player's equipped ship
}

func newFlightModel() flightModel {
	return flightModel{
		ship: spaceflight.NewFlightState(spaceflight.DefaultFlightParams()),
	}
}

// flightParamsForCurrentShip resolves the FlightParams for whatever
// ship the player has equipped. Falls back to DefaultFlightParams
// when no ship is loaded yet (login bootstrap, or pre-ship registration).
//
// Pulls Speed + Maneuverability off the ShipType, not the Ship —
// per-ship damage state doesn't change agility. Outfit-driven
// modifiers (engine upgrades) are P1.3+; this slice just reflects
// the base hull.
func (m Model) flightParamsForCurrentShip() spaceflight.FlightParams {
	if m.currentShip == nil {
		return spaceflight.DefaultFlightParams()
	}
	st := models.GetShipTypeByID(m.currentShip.TypeID)
	if st == nil {
		return spaceflight.DefaultFlightParams()
	}
	return spaceflight.FlightParamsFromShipStats(st.Speed, st.Maneuverability)
}

// updateFlight handles input + tick events. Tick advances physics by
// the wall-clock delta since the last tick (variable-dt) so a 200ms
// hiccup doesn't compound into "ship suddenly across the system" —
// only one frame's worth of motion is missed.
func (m Model) updateFlight(msg tea.Msg) (tea.Model, tea.Cmd) {
	// On first entry, sync the ship's flight params from whatever
	// hull the player is flying. Subsequent visits preserve velocity
	// (that's the whole point of m.flight surviving screen changes).
	if !m.flight.initialized {
		params := m.flightParamsForCurrentShip()
		m.flight.ship.Params = params
		m.flight.initialized = true
	}

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

// viewFlight composes the full cockpit: header, viewport+sidebar,
// info panel under the viewport, and a footer with key reminders.
//
// Layout (80×24 minimum, scales up on bigger terminals):
//
//	┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
//	┃ Pilot: alice  Credits: 5,200 cr  Location: Sol             ┃   header
//	┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
//	┃                                          ┃ ┏━━━━━━━━━━━━━┓ ┃
//	┃             [ STARFIELD VIEWPORT ]       ┃ ┃ SHIP STATUS ┃ ┃
//	┃                    ▲                     ┃ ┃ Hull   100% ┃ ┃
//	┃                                          ┃ ┃ Shields 80% ┃ ┃   sidebar
//	┃                                          ┃ ┃ Fuel    65% ┃ ┃
//	┃                                          ┃ ┃ Cargo  0/30 ┃ ┃
//	┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━┻━━━━━━━━━━━━━━━┫
//	┃ SPD 12.3  HDG 270°  POS (123, -45)                          ┃   HUD
//	┃ W/↑ thrust  S/↓ brake  A/← rot-L  D/→ rot-R  J jump  ESC    ┃   help
//	┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
//
// Width scales with terminal; sidebar is fixed at 18 cells, viewport
// takes the remainder. On narrow terminals (<60 cols) the sidebar
// hides and the viewport gets the full width.
func (m Model) viewFlight() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}
	height := 24
	if m.height > 24 {
		height = m.height
	}

	// Reserved rows: 4 for header (top + title + divider), 3 for
	// HUD/help footer (divider + 2 lines + bottom). Whatever's left
	// becomes the playfield.
	playHeight := height - 4 - 3
	if playHeight < 8 {
		playHeight = 8
	}

	// Sidebar layout: 18 cells of ship status on the right when
	// terminal is wide enough; collapses on narrow.
	sidebarWidth := 18
	if width < 60 {
		sidebarWidth = 0
	}
	viewportWidth := width - sidebarWidth - 2 // 2 = 1 outer L/R border each
	if sidebarWidth > 0 {
		viewportWidth -= 1 // inner divider between viewport and sidebar
	}

	var sb strings.Builder

	// Header — same bordered style the rest of the game uses so the
	// flight screen feels of-a-piece.
	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}
	header := DrawHeader("FLIGHT", m.currentLocationLabel(), credits, m.shieldPercent(), width)
	sb.WriteString(header + "\n")

	// Viewport grid. Built as a 2D rune slice for random-access
	// drawing of stars + ship.
	grid := make([][]rune, playHeight)
	for r := range grid {
		grid[r] = make([]rune, viewportWidth)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	drawStarfield(grid, m.flight.ship.X, m.flight.ship.Y)
	cx, cy := viewportWidth/2, playHeight/2
	if cy < len(grid) && cx < len(grid[cy]) {
		shipRunes := []rune(m.flight.ship.HeadingGlyph())
		if len(shipRunes) > 0 {
			grid[cy][cx] = shipRunes[0]
		}
	}

	// Sidebar — ship status panel with bars for hull/shields/fuel.
	var sidebar []string
	if sidebarWidth > 0 {
		sidebar = m.renderFlightSidebar(sidebarWidth, playHeight)
	}

	// Compose viewport + sidebar side-by-side, line by line.
	for r := 0; r < playHeight; r++ {
		sb.WriteString(BoxVertical)
		sb.WriteString(string(grid[r]))
		if sidebarWidth > 0 {
			sb.WriteString(BoxVertical)
			if r < len(sidebar) {
				sb.WriteString(sidebar[r])
			} else {
				sb.WriteString(strings.Repeat(" ", sidebarWidth))
			}
		}
		sb.WriteString(BoxVertical + "\n")
	}

	// Divider between viewport block and HUD footer.
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// HUD line: speed / heading / position. Fixed-width fields so
	// the line doesn't jitter as values change magnitude.
	hud := fmt.Sprintf(
		" SPD %5.1f   HDG %3d°   POS (%6.0f, %6.0f)",
		m.flight.ship.Speed(),
		m.flight.ship.HeadingDegrees(),
		m.flight.ship.X, m.flight.ship.Y,
	)
	sb.WriteString(BoxVertical)
	sb.WriteString(PadRight(hud, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Help line.
	help := " W/↑ thrust  •  S/↓ brake  •  A/← turn-L  •  D/→ turn-R  •  J jump  •  ESC menu"
	sb.WriteString(BoxVertical)
	sb.WriteString(PadRight(help, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Bottom border.
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// renderFlightSidebar produces the right-hand status panel: hull,
// shields, fuel, energy bars + ship name + cargo summary. Returns a
// slice of pre-padded lines exactly `width` cells wide and at most
// `height` rows long; caller blits them next to the viewport.
//
// Bars are drawn with DrawProgressBar from ui_components; the panel
// stays compact so on default 80-col terminals there's still 60+
// cells of viewport.
func (m Model) renderFlightSidebar(width, height int) []string {
	hull, maxHull := 100, 100
	shields, maxShields := 80, 100
	fuel, maxFuel := 100, 100
	cargoUsed, cargoMax := 0, 30
	shipName := "Unknown"
	shipType := ""

	if m.currentShip != nil {
		hull = m.currentShip.Hull
		shields = m.currentShip.Shields
		fuel = m.currentShip.Fuel
		shipName = m.currentShip.Name
		if st := models.GetShipTypeByID(m.currentShip.TypeID); st != nil {
			maxHull = st.MaxHull
			maxShields = st.MaxShields
			maxFuel = st.MaxFuel
			cargoMax = st.CargoSpace
			shipType = st.Name
		}
	}
	if maxHull < 1 {
		maxHull = 1
	}
	if maxShields < 1 {
		maxShields = 1
	}
	if maxFuel < 1 {
		maxFuel = 1
	}

	// Bar width: 8 cells fits "[████ ░░░]" inside the 18-cell panel
	// after the label and percentage suffix.
	barWidth := width - 9
	if barWidth < 4 {
		barWidth = 4
	}

	hullPct := pctClamped(hull, maxHull)
	shieldsPct := pctClamped(shields, maxShields)
	fuelPct := pctClamped(fuel, maxFuel)

	lines := []string{
		PadRight(" SHIP", width),
		PadRight(" "+TruncateString(shipName, width-2), width),
		PadRight(" "+helpStyle.Render(shipType), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(fmt.Sprintf(" Hull %s", DrawProgressBar(hullPct, 100, barWidth)), width),
		PadRight(fmt.Sprintf("   %3d%%", hullPct), width),
		PadRight(fmt.Sprintf(" Shld %s", DrawProgressBar(shieldsPct, 100, barWidth)), width),
		PadRight(fmt.Sprintf("   %3d%%", shieldsPct), width),
		PadRight(fmt.Sprintf(" Fuel %s", DrawProgressBar(fuelPct, 100, barWidth)), width),
		PadRight(fmt.Sprintf("   %3d%%", fuelPct), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(fmt.Sprintf(" Cargo %d/%d", cargoUsed, cargoMax), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(" ENGINES", width),
		PadRight(fmt.Sprintf(" Max  %3.0f", m.flight.ship.Params.MaxSpeed), width),
		PadRight(fmt.Sprintf(" Acc  %3.1f", m.flight.ship.Params.ThrustImpulse), width),
		PadRight(fmt.Sprintf(" Rot  %3.0f°", m.flight.ship.Params.RotateStep*180/3.14159), width),
	}

	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}

// shieldPercent returns the player ship's shield % for the header,
// falling back to 80 (cosmetic default) when no ship is loaded yet.
func (m Model) shieldPercent() int {
	if m.currentShip == nil {
		return 80
	}
	if st := models.GetShipTypeByID(m.currentShip.TypeID); st != nil && st.MaxShields > 0 {
		return pctClamped(m.currentShip.Shields, st.MaxShields)
	}
	return 80
}

// pctClamped is integer percent of cur out of max, clamped to [0,100].
// Saves callers from worrying about division-by-zero or out-of-range
// inputs from corrupted ship data.
func pctClamped(cur, max int) int {
	if max <= 0 {
		return 0
	}
	p := (cur * 100) / max
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

// drawStarfield places stars at deterministic positions based on
// world coordinates. The same world cell always shows the same
// pattern, so stars don't shimmer as the ship flies past — they
// scroll by, which is what gives the parallax sensation.
//
// Density is sparse (~3% of cells) so the field feels deep without
// overwhelming actual game entities (planets, ships) we'll add in
// later phases.
func drawStarfield(grid [][]rune, shipX, shipY float64) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	height := len(grid)
	width := len(grid[0])
	worldOriginX := int(shipX) - width/2
	worldOriginY := int(shipY) - height/2

	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			wx := worldOriginX + c
			wy := worldOriginY + r
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
