// File: internal/tui/faction_wars.go
// Project: Terminal Velocity
// Description: Faction wars screen — list of active + recent wars
//   with a detail panel showing belligerents, zones, duration, and
//   casus belli. Reads from the server-wide factionwar.Manager.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// factionWarsModel is the screen-local state. The war list itself is
// owned by factionWarManager and queried each frame — this struct
// only tracks cursor / scroll / view toggles.
type factionWarsModel struct {
	cursor      int
	showHistory bool // false = active only, true = include resolved/ceased
}

func newFactionWarsModel() factionWarsModel {
	return factionWarsModel{cursor: 0, showHistory: false}
}

// visibleWars returns the list the screen should render given the
// current toggle. Pure over the manager read; the cursor clamps
// against this slice length in updateFactionWars.
func (m Model) visibleWars() []*models.FactionWar {
	if m.factionWarManager == nil {
		return nil
	}
	if m.factionWarsModel.showHistory {
		return m.factionWarManager.GetAllWars()
	}
	return m.factionWarManager.GetActiveWars()
}

func (m Model) updateFactionWars(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc", "q", "backspace":
		m.screen = ScreenMainMenu
		return m, nil

	case "up", "k":
		if m.factionWarsModel.cursor > 0 {
			m.factionWarsModel.cursor--
		}
	case "down", "j":
		wars := m.visibleWars()
		if m.factionWarsModel.cursor < len(wars)-1 {
			m.factionWarsModel.cursor++
		}

	case "h":
		// Toggle active-only vs. full history. Cursor clamps back
		// into range next render — worst case it lands on a
		// reasonable neighbor rather than a blank row.
		m.factionWarsModel.showHistory = !m.factionWarsModel.showHistory
		wars := m.visibleWars()
		if m.factionWarsModel.cursor >= len(wars) {
			m.factionWarsModel.cursor = maxZero(len(wars) - 1)
		}
	}

	return m, nil
}

func maxZero(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func (m Model) viewFactionWars() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("⚔  FACTION WARS") + "\n")

	if m.factionWarManager == nil {
		sb.WriteString(helpStyle.Render("Faction war system is not available.\n"))
		sb.WriteString("\n" + renderFooter("ESC: Back"))
		return sb.String()
	}

	wars := m.visibleWars()

	// Summary line: active count + toggle state.
	active := len(m.factionWarManager.GetActiveWars())
	total := len(m.factionWarManager.GetAllWars())
	mode := "Active only"
	if m.factionWarsModel.showHistory {
		mode = fmt.Sprintf("All (%d active, %d total)", active, total)
	} else {
		mode = fmt.Sprintf("Active only (%d)", active)
	}
	sb.WriteString(helpStyle.Render("Viewing: "+mode) + "\n")
	sb.WriteString(strings.Repeat("─", width) + "\n\n")

	if len(wars) == 0 {
		if m.factionWarsModel.showHistory {
			sb.WriteString(helpStyle.Render("No wars have ever been declared.\n"))
		} else {
			sb.WriteString(helpStyle.Render("The galaxy is at peace. No active wars.\n"))
			sb.WriteString(helpStyle.Render("Press H to toggle history view.\n"))
		}
		sb.WriteString("\n" + renderFooter("H: Toggle history   ESC: Back"))
		return sb.String()
	}

	// Two-column layout: list on the left (40% width), detail on
	// the right. Fall back to stacked layout on narrow terminals.
	listWidth := (width - 4) / 2
	if listWidth < 30 {
		listWidth = width - 4
	}
	detailWidth := width - listWidth - 4

	listPane := renderWarList(wars, m.factionWarsModel.cursor, listWidth)

	var selected *models.FactionWar
	if m.factionWarsModel.cursor < len(wars) {
		selected = wars[m.factionWarsModel.cursor]
	}

	if detailWidth >= 30 {
		detailPane := renderWarDetail(selected, time.Now(), detailWidth)
		sb.WriteString(sideBySide(listPane, detailPane, listWidth, detailWidth))
	} else {
		sb.WriteString(listPane)
		sb.WriteString("\n")
		sb.WriteString(renderWarDetail(selected, time.Now(), width-4))
	}

	sb.WriteString("\n" + renderFooter("↑/↓: Navigate   H: Toggle history   ESC: Back"))
	return sb.String()
}

// renderWarList formats the left-hand war list. Each entry is one
// line, with the active-cursor row highlighted via the same
// selectedMenuItemStyle the rest of the game uses.
//
// Active wars get a ⚔ marker, resolved wars get ✓, ceased wars get
// ⏸ — scannable at a glance without reading the status label.
func renderWarList(wars []*models.FactionWar, cursor, width int) string {
	var sb strings.Builder
	for i, w := range wars {
		marker := warStatusMarker(w.Status)
		line := fmt.Sprintf("%s %s vs %s", marker, shortName(w.AggressorName), shortName(w.DefenderName))
		line = TruncateString(line, width-3)

		if i == cursor {
			sb.WriteString("> " + selectedMenuItemStyle.Render(line))
		} else {
			sb.WriteString("  " + line)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderWarDetail formats the right-hand detail panel for the
// selected war. nil-safe — returns a placeholder when nothing is
// selected (empty list + out-of-range cursor).
func renderWarDetail(w *models.FactionWar, now time.Time, width int) string {
	if w == nil {
		return helpStyle.Render("No war selected.")
	}
	var sb strings.Builder

	sb.WriteString(highlightStyle.Render(fmt.Sprintf("%s vs %s", w.AggressorName, w.DefenderName)) + "\n")

	statusLabel := fmt.Sprintf("Status: %s", warStatusLabel(w))
	sb.WriteString(statusLabel + "\n")

	sb.WriteString(fmt.Sprintf("Declared: %s\n", w.DeclaredAt.Format("2006-01-02 15:04 UTC")))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", formatWarDuration(w.Duration(now))))

	if w.Status != models.FactionWarActive && w.ResolvedAt != nil {
		sb.WriteString(fmt.Sprintf("Ended:    %s\n", w.ResolvedAt.Format("2006-01-02 15:04 UTC")))
	}
	if w.Status == models.FactionWarResolved {
		winner := w.AggressorName
		if w.WinnerFactionID == w.DefenderID {
			winner = w.DefenderName
		}
		sb.WriteString(fmt.Sprintf("Winner:   %s\n", winner))
	}
	sb.WriteString("\n")

	if w.CasusBelli != "" {
		sb.WriteString(helpStyle.Render("Casus belli:") + "\n")
		for _, line := range wrapText(w.CasusBelli, width-2) {
			sb.WriteString("  " + line + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(helpStyle.Render(fmt.Sprintf("War zones (%d systems):", len(w.WarZoneSystems))) + "\n")
	if len(w.WarZoneSystems) == 0 {
		sb.WriteString("  (none recorded)\n")
	} else {
		// Two-column zones list to save vertical space; keeps the
		// detail panel roughly symmetric with the left list.
		for _, line := range twoColumnList(w.WarZoneSystems, width-2) {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}

// warStatusMarker returns a single-cell glyph identifying war status.
// Kept as a function (not a map) so the exhaustiveness check catches
// new FactionWarStatus values at review time.
func warStatusMarker(s models.FactionWarStatus) string {
	switch s {
	case models.FactionWarActive:
		return "⚔"
	case models.FactionWarResolved:
		return "✓"
	case models.FactionWarCeased:
		return "⏸"
	default:
		return "?"
	}
}

func warStatusLabel(w *models.FactionWar) string {
	switch w.Status {
	case models.FactionWarActive:
		return "ACTIVE"
	case models.FactionWarResolved:
		return "RESOLVED"
	case models.FactionWarCeased:
		return "CEASEFIRE"
	default:
		return string(w.Status)
	}
}

// formatWarDuration turns a duration into a compact "3d 4h" format.
// Zero returns "just now"; sub-minute returns "<1m". Keeps the detail
// panel stable-width regardless of how long the war runs.
func formatWarDuration(d time.Duration) string {
	if d <= 0 {
		return "just now"
	}
	if d < time.Minute {
		return "<1m"
	}
	days := int(d / (24 * time.Hour))
	hours := int((d % (24 * time.Hour)) / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

// shortName returns a compact display name for a faction. The full
// faction name is stored on the war record, but for list rows we
// want something that reliably fits inside a 40-cell column.
func shortName(fullName string) string {
	if len(fullName) <= 20 {
		return fullName
	}
	return fullName[:17] + "..."
}

// twoColumnList arranges the input strings into two columns sized to
// fit `width` cells. Used by the detail panel for the war-zone list
// so 10+ zones don't blow out the screen vertically.
func twoColumnList(items []string, width int) []string {
	if len(items) == 0 {
		return nil
	}
	col := width / 2
	if col < 8 {
		// Narrow panel — single column.
		out := make([]string, len(items))
		for i, it := range items {
			out[i] = TruncateString(it, width)
		}
		return out
	}

	rows := (len(items) + 1) / 2
	out := make([]string, rows)
	for i := 0; i < rows; i++ {
		left := items[i]
		right := ""
		if j := i + rows; j < len(items) {
			right = items[j]
		}
		out[i] = PadRight(TruncateString(left, col-1), col) + TruncateString(right, col)
	}
	return out
}

// sideBySide joins two multi-line strings into a single side-by-side
// layout. left is PadRight-ed to leftWidth; right is joined with a
// two-space gutter between.
func sideBySide(left, right string, leftWidth, rightWidth int) string {
	leftLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right, "\n"), "\n")
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		l := ""
		r := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		sb.WriteString(PadRight(l, leftWidth))
		sb.WriteString("  ")
		sb.WriteString(TruncateString(r, rightWidth))
		sb.WriteString("\n")
	}
	return sb.String()
}

// ============================================================================
// Space-view war-zone banner (P5C-2 gameplay integration)
// ============================================================================

// warZoneBanner returns a single-line red banner shown at the top of
// space_view when the player's current system is covered by one or
// more active wars. Returns "" when the system is peaceful so the
// caller can suppress the row.
func (m Model) warZoneBanner() string {
	if m.factionWarManager == nil || m.player == nil || m.currentSystem == nil {
		return ""
	}
	wars := m.factionWarManager.WarsInSystem(m.currentSystem.Name)
	if len(wars) == 0 {
		return ""
	}
	// Pair-label first war explicitly; note count if more.
	first := wars[0]
	label := fmt.Sprintf("%s vs %s", shortName(first.AggressorName), shortName(first.DefenderName))
	if len(wars) > 1 {
		label = fmt.Sprintf("%s (+%d more)", label, len(wars)-1)
	}
	return warBannerStyle.Render(fmt.Sprintf("⚠  WAR ZONE: %s", label))
}

// warBannerStyle is the style for the space-view contested banner.
// No MarginTop for the same reason documented on the leaderboard
// tab bar / newsreel ticker — lipgloss would otherwise inject a
// blank line ahead of the banner row.
var warBannerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")). // Red
	Bold(true)

// ============================================================================
// P5D-1 territory banner (space view)
// ============================================================================

// territoryOwnerBanner returns a single-line strip showing the NPC
// faction currently controlling the player's system. Returns "" if
// ownership isn't tracked (peaceful / unaligned system) or the
// manager isn't wired. Color is a muted cyan to differentiate from
// the red war banner — both can appear on the same screen.
func (m Model) territoryOwnerBanner() string {
	if m.npcTerritoryManager == nil || m.currentSystem == nil {
		return ""
	}
	name := m.npcTerritoryManager.GetOwnerShortName(m.currentSystem.Name)
	if name == "" {
		return ""
	}
	return territoryBannerStyle.Render(fmt.Sprintf("⚑  %s territory", name))
}

var territoryBannerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("44")). // Teal
	Bold(true)
