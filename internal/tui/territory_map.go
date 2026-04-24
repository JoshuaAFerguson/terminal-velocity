// File: internal/tui/territory_map.go
// Project: Terminal Velocity
// Description: Territory map screen — shows which NPC faction owns
//   each tracked star system, grouped by faction with totals.
//   Reads a live snapshot from npcterritory.Manager each render
//   so wars resolved mid-session update the view without a refresh
//   keybind.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type territoryMapModel struct {
	cursor int
}

func newTerritoryMapModel() territoryMapModel {
	return territoryMapModel{}
}

func (m Model) updateTerritoryMap(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc", "q", "backspace":
		m.screen = ScreenMainMenu
		return m, nil
	case "up", "k":
		if m.territoryMap.cursor > 0 {
			m.territoryMap.cursor--
		}
	case "down", "j":
		groups := m.territoryGroupsSorted()
		if m.territoryMap.cursor < len(groups)-1 {
			m.territoryMap.cursor++
		}
	}
	return m, nil
}

// territoryGroup represents one faction's holdings in the
// territory-map view. Sorted alphabetically by faction name when
// surfaced via territoryGroupsSorted so the left rail stays stable
// between renders even as wars flip individual systems.
type territoryGroup struct {
	factionID string
	name      string // full name for the header
	systems   []string
}

// territoryGroupsSorted assembles the grouped ownership snapshot
// from the manager. Unknown / unseeded states produce an empty
// slice; the view treats that as "no data available" and renders
// a placeholder.
func (m Model) territoryGroupsSorted() []territoryGroup {
	if m.npcTerritoryManager == nil {
		return nil
	}
	snap := m.npcTerritoryManager.AllOwnership()
	if len(snap) == 0 {
		return nil
	}
	byFaction := make(map[string][]string)
	for system, ownerID := range snap {
		byFaction[ownerID] = append(byFaction[ownerID], system)
	}
	groups := make([]territoryGroup, 0, len(byFaction))
	for fid, systems := range byFaction {
		sort.Strings(systems)
		name := m.npcTerritoryManager.GetOwnerName(systems[0])
		if name == "" {
			name = fid // fall back to the slug if display name isn't tracked
		}
		groups = append(groups, territoryGroup{
			factionID: fid,
			name:      name,
			systems:   systems,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].name < groups[j].name
	})
	return groups
}

func (m Model) viewTerritoryMap() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("⚑  TERRITORY MAP") + "\n")

	if m.npcTerritoryManager == nil {
		sb.WriteString(helpStyle.Render("Territory data unavailable — npcterritory manager not wired.\n"))
		sb.WriteString("\n" + renderFooter("ESC: Back"))
		return sb.String()
	}

	groups := m.territoryGroupsSorted()
	if len(groups) == 0 {
		sb.WriteString(helpStyle.Render("No territory data. Server may still be seeding.\n"))
		sb.WriteString("\n" + renderFooter("ESC: Back"))
		return sb.String()
	}

	// Summary line: total systems + faction count.
	totalSystems := 0
	for _, g := range groups {
		totalSystems += len(g.systems)
	}
	sb.WriteString(helpStyle.Render(fmt.Sprintf(
		"%d factions control %d tracked systems",
		len(groups), totalSystems,
	)) + "\n")
	sb.WriteString(strings.Repeat("─", width) + "\n\n")

	// Two-pane layout: faction list on left, selected-group
	// systems on right. Narrow fallback stacks vertically.
	listWidth := 30
	if width < 70 {
		listWidth = width - 2
	}
	detailWidth := width - listWidth - 4

	listPane := renderTerritoryList(groups, m.territoryMap.cursor, listWidth)
	var selected *territoryGroup
	if m.territoryMap.cursor < len(groups) {
		selected = &groups[m.territoryMap.cursor]
	}

	if detailWidth >= 20 {
		detailPane := renderTerritoryDetail(selected, detailWidth)
		sb.WriteString(sideBySide(listPane, detailPane, listWidth, detailWidth))
	} else {
		sb.WriteString(listPane)
		sb.WriteString("\n")
		sb.WriteString(renderTerritoryDetail(selected, width-4))
	}

	sb.WriteString("\n" + renderFooter("↑/↓: Navigate   ESC: Back"))
	return sb.String()
}

// renderTerritoryList formats the left-hand faction list. Each line
// shows faction name + system count; the cursor row gets the
// highlight treatment consistent with the rest of the game's menus.
func renderTerritoryList(groups []territoryGroup, cursor, width int) string {
	var sb strings.Builder
	for i, g := range groups {
		line := fmt.Sprintf("%s (%d)", g.name, len(g.systems))
		line = TruncateString(line, width-2)
		if i == cursor {
			sb.WriteString("> " + selectedMenuItemStyle.Render(line))
		} else {
			sb.WriteString("  " + line)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderTerritoryDetail formats the right-hand systems list for
// the selected faction. Uses the same two-column layout as the
// faction-wars detail panel so long holdings list doesn't overflow
// vertically.
func renderTerritoryDetail(g *territoryGroup, width int) string {
	if g == nil {
		return helpStyle.Render("No faction selected.")
	}
	var sb strings.Builder
	sb.WriteString(highlightStyle.Render(g.name) + "\n")
	sb.WriteString(helpStyle.Render(fmt.Sprintf("%d systems controlled", len(g.systems))) + "\n\n")

	if len(g.systems) == 0 {
		sb.WriteString("  (no holdings)\n")
		return sb.String()
	}

	// Two-column layout mirrors the faction-wars zone list so the
	// two screens feel visually consistent.
	for _, line := range twoColumnList(g.systems, width-2) {
		sb.WriteString("  " + line + "\n")
	}
	return sb.String()
}
