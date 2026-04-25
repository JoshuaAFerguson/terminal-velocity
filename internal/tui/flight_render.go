// File: internal/tui/flight_render.go
// Project: Terminal Velocity
// Description: Color + drawing primitives for the flight cockpit
//   viewport. The viewport is a rune grid plus a parallel "kind"
//   grid that tags each cell with its category (star/planet/ship/
//   target). Rendering walks each row, groups same-kind runs, and
//   styles the runs with lipgloss — keeps ANSI overhead modest while
//   giving us per-glyph color without a 2-D Style array.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-25

package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/spaceflight"
	"github.com/charmbracelet/lipgloss"
)

// cellKind tags grid cells for styled rendering. One byte per cell
// keeps the parallel kind grid cheap (60×16 viewport ≈ 1KB).
type cellKind uint8

const (
	kEmpty cellKind = iota
	kStarDim
	kStarBright
	kPlanetLow
	kPlanetMid
	kPlanetHigh
	kPlanetLabel
	kShip
	kTarget
	kTargetLabel
	kArrow
)

// Cockpit color palette. ANSI 256-color codes — reuse the codebase's
// established "10/11/14/9/8" convention so the cockpit feels of-a-piece
// with the rest of the TUI rather than its own theme.
var (
	styleStarDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleStarBright = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	stylePlanetLow  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red — primitive worlds
	stylePlanetMid  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow — mid-tech
	stylePlanetHigh = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan — high-tech
	stylePlanetLbl  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	styleShipGlyph  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	styleTargetMark = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	styleTargetLbl  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	styleArrowMark  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))

	styleHudLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleHudValue = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleHudHint  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	styleHelp     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	styleBarFull   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleBarMid    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleBarLow    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleBarShield = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	styleBarFuel   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleBarEmpty  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// styleForKind returns the lipgloss style for a kind, or nil for
// kEmpty (nil means "no styling, write raw"). Kept as a switch so
// the compiler builds a jump table instead of a map lookup per cell.
func styleForKind(k cellKind) *lipgloss.Style {
	switch k {
	case kStarDim:
		return &styleStarDim
	case kStarBright:
		return &styleStarBright
	case kPlanetLow:
		return &stylePlanetLow
	case kPlanetMid:
		return &stylePlanetMid
	case kPlanetHigh:
		return &stylePlanetHigh
	case kPlanetLabel:
		return &stylePlanetLbl
	case kShip:
		return &styleShipGlyph
	case kTarget:
		return &styleTargetMark
	case kTargetLabel:
		return &styleTargetLbl
	case kArrow:
		return &styleArrowMark
	}
	return nil
}

// planetKind maps tech level → color tier. Buckets chosen so a typical
// 1-10 distribution gives a readable spread at a glance: red worlds are
// the frontier, cyan worlds are the polished metropolises.
func planetKind(techLevel int) cellKind {
	switch {
	case techLevel >= 7:
		return kPlanetHigh
	case techLevel >= 4:
		return kPlanetMid
	default:
		return kPlanetLow
	}
}

// renderStyledRow emits a single row as ANSI-styled text. Groups
// consecutive cells with the same kind into a single Render call so a
// run of 12 stars costs one ANSI prefix/suffix instead of 12 — keeps
// per-frame bytes manageable on the WebSocket path.
func renderStyledRow(runes []rune, kinds []cellKind) string {
	if len(runes) == 0 {
		return ""
	}
	var sb strings.Builder
	n := len(runes)
	i := 0
	for i < n {
		j := i + 1
		for j < n && kinds[j] == kinds[i] {
			j++
		}
		run := string(runes[i:j])
		if s := styleForKind(kinds[i]); s != nil {
			sb.WriteString(s.Render(run))
		} else {
			sb.WriteString(run)
		}
		i = j
	}
	return sb.String()
}

// drawStarfieldKinded fills empty cells with stars, tagging them as
// dim or bright in the kind grid. Identical placement to the older
// drawStarfield; the kind grid is the only addition.
func drawStarfieldKinded(grid [][]rune, kinds [][]cellKind, shipX, shipY float64) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	height := len(grid)
	width := len(grid[0])
	worldOriginX := int(shipX) - width/2
	worldOriginY := int(shipY) - height/2

	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			if grid[r][c] != ' ' {
				continue
			}
			h := starHash(worldOriginX+c, worldOriginY+r)
			if h%32 != 0 {
				continue
			}
			if h%128 == 0 {
				grid[r][c] = '*'
				kinds[r][c] = kStarBright
			} else {
				grid[r][c] = '·'
				kinds[r][c] = kStarDim
			}
		}
	}
}

// drawPlanetsKinded plots planets at world positions and writes their
// labels to the right of each glyph. Highlights the planet that matches
// targetID with the kTarget kind so it pops in a different color than
// the other worlds. Planets outside the viewport are skipped here —
// the off-screen guidance arrow is drawn separately.
func drawPlanetsKinded(grid [][]rune, kinds [][]cellKind, planets []planetEntity, shipX, shipY float64, targetID string) {
	if len(grid) == 0 || len(grid[0]) == 0 || len(planets) == 0 {
		return
	}
	height := len(grid)
	width := len(grid[0])
	cx, cy := width/2, height/2

	for _, p := range planets {
		sx := cx + int(p.x-shipX)
		sy := cy + int(p.y-shipY)
		if sx < 0 || sx >= width || sy < 0 || sy >= height {
			continue
		}
		grid[sy][sx] = '●'
		isTarget := p.id == targetID
		if isTarget {
			kinds[sy][sx] = kTarget
		} else {
			kinds[sy][sx] = planetKind(p.techLevel)
		}

		// Label right of the glyph with a 1-cell gap. Truncate when
		// it would clip; clean elision reads better than a half label.
		labelStart := sx + 2
		if labelStart >= width {
			continue
		}
		labelMax := width - labelStart
		name := p.name
		if len(name) > labelMax {
			name = name[:labelMax]
		}
		labelKind := kPlanetLabel
		if isTarget {
			labelKind = kTargetLabel
		}
		for i, r := range name {
			if labelStart+i >= width {
				break
			}
			grid[sy][labelStart+i] = r
			kinds[sy][labelStart+i] = labelKind
		}
	}
}

// drawShipGlyph stamps the ship in the center of the viewport with the
// correct heading rune. Tagged kShip so it gets the bold cyan style
// and stands out clearly against the starfield.
func drawShipGlyph(grid [][]rune, kinds [][]cellKind, headingGlyph string) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	cx, cy := len(grid[0])/2, len(grid)/2
	runes := []rune(headingGlyph)
	if len(runes) == 0 {
		return
	}
	grid[cy][cx] = runes[0]
	kinds[cy][cx] = kShip
}

// drawOffScreenArrow places a directional indicator on the viewport
// edge pointing toward the target when the target is off-screen.
// Picks the rune (▲▶▼◀ / corners) by quadrant of the ship→target
// vector, and clamps the cell to the closest edge so it's actually
// visible. No-op when the target is on-screen (drawPlanetsKinded
// already highlighted it there).
func drawOffScreenArrow(grid [][]rune, kinds [][]cellKind, target *planetEntity, shipX, shipY float64) {
	if target == nil || len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	height := len(grid)
	width := len(grid[0])
	cx, cy := width/2, height/2
	sx := cx + int(target.x-shipX)
	sy := cy + int(target.y-shipY)
	if sx >= 0 && sx < width && sy >= 0 && sy < height {
		return // on-screen, no edge arrow needed
	}
	dx := target.x - shipX
	dy := target.y - shipY
	if dx == 0 && dy == 0 {
		return
	}

	// Choose a glyph by 8-way quadrant of the bearing. Using octants
	// rather than the four cardinals because a target at "up and right"
	// reads more clearly with ↗ than with either ▲ or ▶ alone.
	angle := math.Atan2(dy, dx) // -π..π, +X = 0
	// Normalize to 0..2π then split into 8 octants of π/4 each, with
	// boundary rotated by π/8 so cardinals land in the middle of their
	// octant rather than on the seam.
	a := angle
	if a < 0 {
		a += 2 * math.Pi
	}
	octant := int(math.Floor((a+math.Pi/8)/(math.Pi/4))) % 8
	glyphs := []rune{'▶', '◢', '▼', '◣', '◀', '◤', '▲', '◥'}
	glyph := glyphs[octant]

	// Place the arrow on the edge cell along the ship→target ray.
	// Walk from the center toward the target until we step off-grid;
	// last in-bounds cell is where the arrow lands.
	steps := math.Max(math.Abs(dx), math.Abs(dy))
	if steps == 0 {
		return
	}
	stepX := dx / steps
	stepY := dy / steps
	lastX, lastY := cx, cy
	for t := 1.0; t <= steps; t++ {
		nx := cx + int(stepX*t)
		ny := cy + int(stepY*t)
		if nx < 0 || nx >= width || ny < 0 || ny >= height {
			break
		}
		lastX, lastY = nx, ny
	}
	grid[lastY][lastX] = glyph
	kinds[lastY][lastX] = kArrow
}

// styledHealthBar renders a width-cell progress bar where the filled
// portion is colored by remaining ratio (green→yellow→red) and the
// empty portion is muted. Mirrors DrawProgressBar's character set so
// terminals without color support still see a meaningful bar.
func styledHealthBar(cur, max, width int, palette ...*lipgloss.Style) string {
	if max <= 0 || width < 1 {
		return strings.Repeat(" ", width)
	}
	pct := float64(cur) / float64(max)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filledW := int(float64(width) * pct)
	emptyW := width - filledW

	// Default tri-color palette for hull-style bars; callers pass an
	// override (e.g. shield uses solid blue) when they want a fixed hue.
	full := &styleBarFull
	switch {
	case len(palette) >= 1 && palette[0] != nil:
		full = palette[0]
	case pct < 0.33:
		full = &styleBarLow
	case pct < 0.66:
		full = &styleBarMid
	}

	var sb strings.Builder
	if filledW > 0 {
		sb.WriteString(full.Render(strings.Repeat(ProgressFull, filledW)))
	}
	if emptyW > 0 {
		sb.WriteString(styleBarEmpty.Render(strings.Repeat(ProgressEmpty, emptyW)))
	}
	return sb.String()
}

// renderRadar produces a compact system-overview radar sized for the
// sidebar's top portion. The radar centers on the system origin (0,0)
// and scales to fit every known planet plus the ship inside the panel,
// so a glance answers "where am I in the system?". Returns exactly
// `height` rows of `width`-cell strings — caller splices them above
// the ship-status block in the sidebar.
//
// Layout per row:
//   row 0:           " RADAR "                          (title)
//   row 1..height-2: grid (planets + ship + target)
//   row height-1:    horizontal divider
//
// Planet glyphs reuse the cellKind palette so the radar agrees with
// the main viewport visually — a high-tech world is cyan in both
// places, the target is orange in both, etc.
func renderRadar(width, height int, ship spaceflight.FlightState, planets []planetEntity, targetID string) []string {
	if width < 6 || height < 4 {
		return nil
	}
	gridH := height - 2
	gridW := width

	// Compute the system extent so the radar fits everything. We keep
	// the ship in the calculation so a player who has flown well past
	// the planet ring still appears on the panel rather than vanishing
	// off the edge. 1.1× padding stops dots from clinging to borders.
	maxExtent := 800.0
	for _, p := range planets {
		if v := math.Abs(p.x); v > maxExtent {
			maxExtent = v
		}
		if v := math.Abs(p.y); v > maxExtent {
			maxExtent = v
		}
	}
	if v := math.Abs(ship.X); v > maxExtent {
		maxExtent = v
	}
	if v := math.Abs(ship.Y); v > maxExtent {
		maxExtent = v
	}
	maxExtent *= 1.1

	grid := make([][]rune, gridH)
	kinds := make([][]cellKind, gridH)
	for r := range grid {
		grid[r] = make([]rune, gridW)
		kinds[r] = make([]cellKind, gridW)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}

	cx, cy := gridW/2, gridH/2

	// Origin tick — a faint plus marks (0,0) so the player has a
	// stable visual anchor as the ship moves around the panel.
	if cx >= 0 && cx < gridW && cy >= 0 && cy < gridH {
		grid[cy][cx] = '+'
		kinds[cy][cx] = kStarDim
	}

	// World → radar cell. Cells are anisotropic in pixel-space (chars
	// are roughly twice as tall as wide), but we scale X and Y the
	// same here so distances/bearings on the radar match the player's
	// mental model from the main viewport.
	plot := func(wx, wy float64) (int, int, bool) {
		rx := cx + int(wx/maxExtent*float64(cx))
		ry := cy + int(wy/maxExtent*float64(cy))
		if rx < 0 || rx >= gridW || ry < 0 || ry >= gridH {
			return 0, 0, false
		}
		return rx, ry, true
	}

	// Plot planets first; ship on top so it always wins overlap.
	for _, p := range planets {
		rx, ry, ok := plot(p.x, p.y)
		if !ok {
			continue
		}
		if p.id == targetID {
			grid[ry][rx] = '◉'
			kinds[ry][rx] = kTarget
		} else {
			grid[ry][rx] = '·'
			kinds[ry][rx] = planetKind(p.techLevel)
		}
	}

	if rx, ry, ok := plot(ship.X, ship.Y); ok {
		runes := []rune(ship.HeadingGlyph())
		if len(runes) > 0 {
			grid[ry][rx] = runes[0]
			kinds[ry][rx] = kShip
		}
	}

	out := make([]string, 0, height)
	out = append(out, padStyledRight(" "+HighlightStyle.Render("RADAR"), width))
	for r := 0; r < gridH; r++ {
		out = append(out, padStyledRight(renderStyledRow(grid[r], kinds[r]), width))
	}
	out = append(out, strings.Repeat(BoxHorizontal, width))
	return out
}

// padStyledRight pads a styled string to `width` cells. lipgloss.Width
// already strips ANSI escapes when measuring, so this is identical in
// shape to PadRight but kept local since the radar grid is the only
// caller that mixes per-cell styled runs into a single line.
func padStyledRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// renderHudLine composes the bottom-of-viewport HUD with colored
// label/value pairs. Callers append the dock/target hint themselves
// since its presence depends on flight state.
func renderHudLine(speed float64, headingDeg int, x, y float64, hint string) string {
	parts := []string{
		styleHudLabel.Render(" SPD "), styleHudValue.Render(fmt.Sprintf("%5.1f", speed)),
		styleHudLabel.Render("  HDG "), styleHudValue.Render(fmt.Sprintf("%3d°", headingDeg)),
		styleHudLabel.Render("  POS "), styleHudValue.Render(fmt.Sprintf("(%6.0f,%6.0f)", x, y)),
	}
	line := strings.Join(parts, "")
	if hint != "" {
		line += "  " + styleHudHint.Render(hint)
	}
	return line
}
