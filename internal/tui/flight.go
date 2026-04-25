// File: internal/tui/flight.go
// Project: Terminal Velocity
// Description: Real-time flight cockpit screen. Multi-pane layout
//   with viewport in the center, ship status sidebar on the right,
//   target/system info under the viewport, header at the top, and
//   key reminders at the bottom. Replaces the static "space view"
//   radar as the primary 'in your ship' screen.
// Version: 1.3.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"context"
	"fmt"
	"math"
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

// planetEntity is the flight-layer projection of a models.Planet.
// We compute world (X,Y) once on load — either using the persisted
// values when they're non-zero, or hashing the planet UUID to a
// deterministic position when the universe generator left them at
// (0,0). Caching here means the renderer doesn't recompute every
// frame.
//
// The full *models.Planet is stashed via planet so the L-key
// dock path can pass it straight to dockCmd without re-fetching
// from the repo.
type planetEntity struct {
	id          string
	name        string
	x, y        float64 // world coordinates
	techLevel   int
	hasServices bool
	planet      *models.Planet
}

// flightModel owns the per-session flight state. Storing the ship
// here means velocity persists across screen visits — a player who
// pops out to check trade prices doesn't reset their inertia.
//
// targetID is the player-selected destination — an empty string means
// "no target". Pressing L either picks the nearest landable as target
// (if none set, or if the current one is no longer the closest) or
// lands at the current target when within dockableRange. Persisting
// it on the model means navigation guidance survives screen pops.
type flightModel struct {
	ship          spaceflight.FlightState
	active        bool      // true while the tick loop is running
	lastTick      time.Time // wall-clock of last tick (variable-dt physics)
	initialized   bool      // true after first sync from the player's equipped ship
	planets       []planetEntity
	planetsLoaded bool          // true once the system's planets have been fetched
	loadInFlight  bool          // true while a planet load is queued but not yet returned
	loadedSystem  string        // ID of the system whose planets are cached
	dockTarget    *planetEntity // nearest-in-range planet on this frame; nil when nothing dockable
	targetID      string        // selected target's planet ID; "" = none
}

// flightDataLoadedMsg fires once the async planet fetch completes.
// Sent by loadFlightDataCmd after a successful systemRepo lookup.
type flightDataLoadedMsg struct {
	systemID string
	planets  []planetEntity
	err      error
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

// loadFlightDataCmd fetches the planets for the player's current
// system and projects them to flight-world coordinates. Returns
// flightDataLoadedMsg so the model can cache the result. Async
// because the system repo hits the DB; we don't want to stall a
// flight tick on it.
func (m Model) loadFlightDataCmd() tea.Cmd {
	return func() tea.Msg {
		if m.player == nil || m.systemRepo == nil || m.player.CurrentSystem.String() == "00000000-0000-0000-0000-000000000000" {
			return flightDataLoadedMsg{}
		}
		ctx := context.Background()
		raw, err := m.systemRepo.GetPlanetsBySystem(ctx, m.player.CurrentSystem)
		if err != nil {
			return flightDataLoadedMsg{err: err}
		}
		out := make([]planetEntity, 0, len(raw))
		for _, p := range raw {
			x, y := planetPosition(p)
			out = append(out, planetEntity{
				id:          p.ID.String(),
				name:        p.Name,
				x:           x,
				y:           y,
				techLevel:   p.TechLevel,
				hasServices: len(p.Services) > 0,
				planet:      p,
			})
		}
		return flightDataLoadedMsg{
			systemID: m.player.CurrentSystem.String(),
			planets:  out,
		}
	}
}

// planetPosition resolves a Planet's world coordinates. Falls back to
// a UUID-derived deterministic position when the persisted X,Y are
// (0,0) — which is currently every planet in any universe generated
// before the generator was teaching about in-system layout. Hash
// → angle in [0, 2π) and distance in [400, 1800] u from system
// center. Distinct planets land on distinct angles since the hash
// has full UUID entropy; a system with 5 planets gets them spread
// around the center, not bunched.
func planetPosition(p *models.Planet) (x, y float64) {
	if p.X != 0 || p.Y != 0 {
		return p.X, p.Y
	}
	// Hash the UUID bytes into two uint64s — one drives angle, the
	// other drives distance. Splitting like this keeps the angle
	// and distance distributions independent.
	idBytes := p.ID
	var hAngle, hDist uint64
	for i := 0; i < 8; i++ {
		hAngle = hAngle*1099511628211 ^ uint64(idBytes[i])
	}
	for i := 8; i < 16; i++ {
		hDist = hDist*1099511628211 ^ uint64(idBytes[i])
	}
	angle := float64(hAngle%65536) / 65536.0 * 2 * math.Pi
	dist := 400 + float64(hDist%1400)
	x = math.Cos(angle) * dist
	y = math.Sin(angle) * dist
	return x, y
}

// dockableRange is how close (world units) the player must be to a
// planet to dock. Tuned generous because flight controls are still
// somewhat coarse — players should be able to coast in and tag the
// L key without precision pixel-hunting.
const dockableRange = 80.0

// nearestDockable returns the closest planet within dockableRange,
// or nil. Caller uses this for both the HUD prompt ("Press L to
// land at <name>") and the actual L-key handling.
func nearestDockable(ship spaceflight.FlightState, planets []planetEntity) *planetEntity {
	if len(planets) == 0 {
		return nil
	}
	bestDist := dockableRange
	var best *planetEntity
	for i := range planets {
		dx := planets[i].x - ship.X
		dy := planets[i].y - ship.Y
		d := math.Hypot(dx, dy)
		if d <= bestDist {
			bestDist = d
			best = &planets[i]
		}
	}
	return best
}

// nearestPlanet returns the closest planet at any distance. Used when
// the player hits L without a target set — we pick the closest planet
// in the system as their navigation goal regardless of how far it is,
// so they always have a clear "fly toward this" destination.
func nearestPlanet(ship spaceflight.FlightState, planets []planetEntity) *planetEntity {
	if len(planets) == 0 {
		return nil
	}
	bestDist := math.Inf(1)
	var best *planetEntity
	for i := range planets {
		dx := planets[i].x - ship.X
		dy := planets[i].y - ship.Y
		d := math.Hypot(dx, dy)
		if d < bestDist {
			bestDist = d
			best = &planets[i]
		}
	}
	return best
}

// findPlanetByID returns the planet matching id, or nil. Used to
// resolve the persisted targetID on each frame so the renderer/HUD
// have a fresh pointer (the planet slice is rebuilt across loads).
func findPlanetByID(planets []planetEntity, id string) *planetEntity {
	if id == "" {
		return nil
	}
	for i := range planets {
		if planets[i].id == id {
			return &planets[i]
		}
	}
	return nil
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

	// Process the planet-load completion BEFORE any gating that depends
	// on planetsLoaded. If we let the gate run first it would re-issue
	// the loader and silently drop this message — leaving planetsLoaded
	// false forever, which then drops every subsequent tick and key.
	if loaded, ok := msg.(flightDataLoadedMsg); ok {
		if loaded.err == nil {
			m.flight.planets = loaded.planets
		}
		m.flight.planetsLoaded = true
		m.flight.loadInFlight = false
		m.flight.loadedSystem = loaded.systemID
		// Keep the tick chain alive — the load may have arrived on a
		// frame where the active tick was dropped by the (now-removed)
		// gate. Re-arm only if we somehow lost the loop.
		if !m.flight.active {
			m.flight.active = true
			m.flight.lastTick = time.Now()
			return m, flightTick()
		}
		return m, nil
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

	// Load planets for the current system on first entry, and re-load
	// whenever the player has jumped (system changed). loadInFlight
	// dedupes — without it, every message that arrives during the
	// pending load issues another loader, and *also* gets swallowed
	// on the way through (no tick re-arm, no key processing).
	var loader tea.Cmd
	if m.player != nil && !m.flight.loadInFlight {
		curSys := m.player.CurrentSystem.String()
		if !m.flight.planetsLoaded || m.flight.loadedSystem != curSys {
			m.flight.loadInFlight = true
			loader = m.loadFlightDataCmd()
		}
	}

	// Compose any standalone kicker/loader work with whatever the
	// message switch produces, so we never drop a key or tick just
	// because we needed to start a load in the same frame.
	withSetup := func(cmd tea.Cmd) tea.Cmd {
		switch {
		case kicker == nil && loader == nil:
			return cmd
		case cmd == nil:
			return tea.Batch(kicker, loader)
		default:
			return tea.Batch(cmd, kicker, loader)
		}
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
		// Refresh dock target so the HUD shows "Press L to land at
		// X" only when actually in range. Recomputed every tick is
		// fine — N usually <10 planets per system.
		m.flight.dockTarget = nearestDockable(m.flight.ship, m.flight.planets)
		// Re-arm via withSetup so a load-in-flight Cmd produced this
		// frame still gets dispatched alongside the next tick.
		return m, withSetup(flightTick())

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

		case "l", "L":
			// "Do the right thing" L: if there's a planet within
			// dock range, land at it. Otherwise set the navigation
			// target to the nearest planet so the off-screen arrow
			// + HUD guidance kick in. One key, no surprise — players
			// don't have to remember a "target then commit" sequence.
			if d := nearestDockable(m.flight.ship, m.flight.planets); d != nil && d.planet != nil {
				m.flight.active = false // stop the tick; we're leaving the cockpit
				m.flight.targetID = ""  // clear so the next launch starts fresh
				m.previousScreen = ScreenFlight
				m.hasPreviousScreen = true
				m.screen = ScreenLanding
				return m, m.dockCmd(d.planet)
			}
			if next := nearestPlanet(m.flight.ship, m.flight.planets); next != nil {
				m.flight.targetID = next.id
			}
			return m, withSetup(nil)
		}
	}

	return m, withSetup(nil)
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

	// Viewport grid + parallel kind grid. Two slices kept in lockstep:
	// glyphs in `grid`, color tags in `kinds`. Render walks both to
	// emit ANSI-styled rows.
	grid := make([][]rune, playHeight)
	kinds := make([][]cellKind, playHeight)
	for r := range grid {
		grid[r] = make([]rune, viewportWidth)
		kinds[r] = make([]cellKind, viewportWidth)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	drawStarfieldKinded(grid, kinds, m.flight.ship.X, m.flight.ship.Y)
	// Planets first so the ship glyph wins if it overlaps. Pass the
	// targetID so the targeted planet gets the highlight color.
	drawPlanetsKinded(grid, kinds, m.flight.planets, m.flight.ship.X, m.flight.ship.Y, m.flight.targetID)
	// Off-screen edge arrow toward the target — only drawn when the
	// target exists and is outside the viewport.
	if target := findPlanetByID(m.flight.planets, m.flight.targetID); target != nil {
		drawOffScreenArrow(grid, kinds, target, m.flight.ship.X, m.flight.ship.Y)
	}
	drawShipGlyph(grid, kinds, m.flight.ship.HeadingGlyph())

	// Sidebar — ship status panel with bars for hull/shields/fuel.
	var sidebar []string
	if sidebarWidth > 0 {
		sidebar = m.renderFlightSidebar(sidebarWidth, playHeight)
	}

	// Compose viewport + sidebar side-by-side, line by line. Each row
	// is run-length encoded by kind via renderStyledRow so a row of
	// 60 cells with 3 stars and 1 planet only emits ~5 ANSI runs
	// instead of 60.
	for r := 0; r < playHeight; r++ {
		sb.WriteString(BoxVertical)
		sb.WriteString(renderStyledRow(grid[r], kinds[r]))
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

	// HUD line: speed / heading / position with a context-sensitive
	// hint at the end. Hint priority: in-range "land NOW" beats the
	// "fly to TARGET" guidance, since pressing L while in range is
	// what actually completes the action.
	hint := ""
	if m.flight.dockTarget != nil {
		hint = fmt.Sprintf("⬇ L: land at %s", m.flight.dockTarget.name)
	} else if target := findPlanetByID(m.flight.planets, m.flight.targetID); target != nil {
		dx := target.x - m.flight.ship.X
		dy := target.y - m.flight.ship.Y
		dist := math.Hypot(dx, dy)
		// Bearing as compass-degrees-from-ship so the player can rotate
		// to match. atan2(dy,dx) gives 0=east; we shift to 0=up so the
		// ship's heading display and target bearing share a frame.
		bearing := int(math.Mod(math.Atan2(dy, dx)*180/math.Pi+90+360, 360))
		hint = fmt.Sprintf("⊕ %s  %4.0fu  %3d°", target.name, dist, bearing)
	}
	hud := renderHudLine(
		m.flight.ship.Speed(),
		m.flight.ship.HeadingDegrees(),
		m.flight.ship.X, m.flight.ship.Y,
		hint,
	)
	sb.WriteString(BoxVertical)
	sb.WriteString(PadRight(hud, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Help line — muted so the eye lands on the HUD numbers above.
	help := styleHelp.Render(" W/↑ thrust  •  S/↓ brake  •  A/← turn-L  •  D/→ turn-R  •  L target/land  •  J jump  •  ESC menu")
	sb.WriteString(BoxVertical)
	sb.WriteString(PadRight(help, width-2))
	sb.WriteString(BoxVertical + "\n")

	// Chat tail — last 3 global messages so the player sees
	// incoming traffic without leaving the cockpit. Read-only here;
	// the dedicated chat screen still owns input.
	chatLines := m.renderFlightChat(width - 2)
	if len(chatLines) > 0 {
		sb.WriteString(BoxCrossLeft)
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
		sb.WriteString(BoxCross + "\n")
		for _, line := range chatLines {
			sb.WriteString(BoxVertical)
			sb.WriteString(PadRight(line, width-2))
			sb.WriteString(BoxVertical + "\n")
		}
	}

	// Bottom border.
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// renderFlightChat returns 3 most-recent global chat messages
// formatted to fit `innerWidth` cells. Empty slice when chat isn't
// wired (no manager) — caller skips the chat divider entirely so
// fresh-server cockpits don't show an empty box.
//
// Format: " [HH:MM] sender: message" — matches the dedicated chat
// screen's tail style. Truncation is brutal (single-line) since
// vertical real estate is tight; players read full history in
// the dedicated chat screen.
func (m Model) renderFlightChat(innerWidth int) []string {
	if m.chatManager == nil {
		return nil
	}
	const tailRows = 3
	msgs := m.chatManager.GetRecentGlobal(tailRows)
	if len(msgs) == 0 {
		// Show one placeholder line so players know where messages
		// will appear — silence is more confusing than "no traffic".
		return []string{" " + helpStyle.Render("[chat] no recent messages")}
	}
	out := make([]string, 0, tailRows)
	for _, msg := range msgs {
		// Timestamp + sender + content. Strip leading/trailing
		// whitespace defensively in case some sender escaped one.
		ts := msg.Timestamp.Format("15:04")
		line := fmt.Sprintf(" [%s] %s: %s", ts, msg.Sender, strings.TrimSpace(msg.Content))
		out = append(out, TruncateString(line, innerWidth))
	}
	// Pad to tailRows so the panel doesn't jitter in height as the
	// last few minutes of traffic comes and goes.
	for len(out) < tailRows {
		out = append(out, "")
	}
	return out
}

// renderFlightSidebar produces the right-hand status panel: a system
// radar at the top and ship/engine status below. Returns exactly
// `height` rows of `width`-cell strings; caller blits them next to
// the viewport.
//
// Layout: top ~9 rows are radar (title + grid + divider), the rest
// is ship status. On terminals shorter than ~18 rows the radar is
// dropped so the status panel still fits.
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

	hullBar := styledHealthBar(hullPct, 100, barWidth)
	shldBar := styledHealthBar(shieldsPct, 100, barWidth, &styleBarShield)
	fuelBar := styledHealthBar(fuelPct, 100, barWidth, &styleBarFuel)
	lines := []string{
		PadRight(" "+HighlightStyle.Render("SHIP"), width),
		PadRight(" "+TruncateString(shipName, width-2), width),
		PadRight(" "+styleHudLabel.Render(shipType), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(fmt.Sprintf(" Hull %s", hullBar), width),
		PadRight(styleHudValue.Render(fmt.Sprintf("   %3d%%", hullPct)), width),
		PadRight(fmt.Sprintf(" Shld %s", shldBar), width),
		PadRight(styleHudValue.Render(fmt.Sprintf("   %3d%%", shieldsPct)), width),
		PadRight(fmt.Sprintf(" Fuel %s", fuelBar), width),
		PadRight(styleHudValue.Render(fmt.Sprintf("   %3d%%", fuelPct)), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(fmt.Sprintf(" Cargo %d/%d", cargoUsed, cargoMax), width),
		strings.Repeat(BoxHorizontal, width),
		PadRight(" "+HighlightStyle.Render("ENGINES"), width),
		PadRight(fmt.Sprintf(" Max  %3.0f", m.flight.ship.Params.MaxSpeed), width),
		PadRight(fmt.Sprintf(" Acc  %3.1f", m.flight.ship.Params.ThrustImpulse), width),
		PadRight(fmt.Sprintf(" Rot  %3.0f°", m.flight.ship.Params.RotateStep*180/3.14159), width),
	}

	// Reserve top portion for the radar when there's enough vertical
	// room. 9 rows = title + 7-row grid + bottom divider; below that
	// the radar isn't really legible, so drop it on small terminals
	// rather than render a 2-row useless box.
	radarH := 0
	if height >= 18 {
		radarH = 9
	} else if height >= 14 {
		radarH = 7
	}
	statusH := height - radarH

	if len(lines) > statusH {
		lines = lines[:statusH]
	}
	for len(lines) < statusH {
		lines = append(lines, strings.Repeat(" ", width))
	}

	if radarH == 0 {
		return lines
	}
	radar := renderRadar(width, radarH, m.flight.ship, m.flight.planets, m.flight.targetID)
	if len(radar) == 0 {
		return lines
	}
	return append(radar, lines...)
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

// starHash is a small integer hash over (x, y). Output quality
// matters less than determinism — same input → same star. xorshift-
// flavored mix is fine; we don't need cryptographic distribution.
// Used by drawStarfieldKinded in flight_render.go.
func starHash(x, y int) uint32 {
	h := uint32(x*73856093) ^ uint32(y*19349663)
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return h
}
