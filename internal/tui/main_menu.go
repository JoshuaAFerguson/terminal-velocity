// File: internal/tui/main_menu.go
// Project: Terminal Velocity
// Description: Main menu screen — top-level navigation hub. Two-tier
//   layout (categories at top, items inside) so the player isn't
//   staring at a 30-item flat list, plus dock-aware filtering so
//   station services only appear when docked at a planet.
// Version: 2.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// menuCategory groups related items so the top-level menu shows ~10
// entries instead of 30+. Empty string is "top-level" — for items
// like Launch / Take Off / Quit that don't belong inside a category.
type menuCategory string

const (
	catNone     menuCategory = ""
	catStation  menuCategory = "Station"  // visible only when docked
	catPilot    menuCategory = "Pilot"
	catProgress menuCategory = "Progress"
	catSocial   menuCategory = "Social"
	catPolitics menuCategory = "Politics"
	catMarkets  menuCategory = "Markets"
	catHelp     menuCategory = "Help & Settings"
)

// menuVisibility narrows when a top-level entry shows up. Most items
// are visibleAlways; Launch swaps to Take Off when docked, and
// dock-only entries (Take Off, the Station category) only show up
// when the player has CurrentPlanet set.
type menuVisibility int

const (
	showAlways    menuVisibility = iota // visible regardless of dock state
	showWhenInFlight                    // visible only when not docked
	showWhenDocked                      // visible only when docked
)

// orderedCategories is the canonical render order for the top-level
// list. Driven by a slice (not iteration over a map) so categories
// land in predictable positions regardless of Go map traversal.
var orderedCategories = []menuCategory{
	catStation, // first when present so it's the obvious choice on land
	catPilot,
	catProgress,
	catSocial,
	catPolitics,
	catMarkets,
	catHelp,
}

// mainMenuModel holds the menu's render state. openCategory != ""
// means the user has drilled into a category and we're showing its
// items + a "Back" entry; "" means we're at the top level showing
// categories + action shortcuts.
type mainMenuModel struct {
	cursor        int
	openCategory  menuCategory
	items         []menuItem // single source of truth — view filters
}

// menuItem is one row in the menu tree. category="" means it's a
// top-level action (Launch, Quit). Otherwise it lives under a
// category and only shows up when that category is open.
type menuItem struct {
	label      string
	screen     Screen
	action     func(*Model) tea.Cmd
	category   menuCategory
	visibility menuVisibility
	adminOnly  bool // hide unless the player is in adminManager
}

// newMainMenuModel returns the menu's items as a single flat list
// with category tags. The view filters this slice on render — there
// is no pre-baked "categories" data, just a derivation from the
// items themselves.
func newMainMenuModel() mainMenuModel {
	return mainMenuModel{
		items: []menuItem{
			// === Top-level action shortcuts ===
			//
			// Launch + Take Off are mutually exclusive based on dock
			// state. Take Off chains takeoffCmd → ScreenFlight so the
			// dock state clears in the DB before the player drops
			// into the cockpit; without that, the next CurrentPlanet
			// read would think they're still on a planet.
			{label: "▸ Launch", screen: ScreenFlight, visibility: showWhenInFlight},
			{
				label:      "▸ Take Off",
				visibility: showWhenDocked,
				action: func(m *Model) tea.Cmd {
					// Clear m.flight.active so the cockpit starts a
					// fresh tick loop on entry rather than racing
					// the takeoff DB write.
					m.flight.active = false
					m.screen = ScreenFlight
					return m.takeoffCmd()
				},
			},

			// === Station category — docked-only services ===
			{label: "Trading", screen: ScreenTrading, category: catStation, visibility: showWhenDocked},
			{label: "Shipyard", screen: ScreenShipyard, category: catStation, visibility: showWhenDocked},
			{label: "Outfitter", screen: ScreenOutfitter, category: catStation, visibility: showWhenDocked},
			{label: "Advanced Outfitting", screen: ScreenOutfitterEnhanced, category: catStation, visibility: showWhenDocked},
			{label: "Mission Board", screen: ScreenMissions, category: catStation, visibility: showWhenDocked},

			// === Pilot category — your ship + your record ===
			{label: "Cargo Hold", screen: ScreenCargo, category: catPilot},
			{label: "Ship Management", screen: ScreenShipManagement, category: catPilot},
			{label: "Pilot Record", screen: ScreenPilotRecord, category: catPilot},
			{label: "Navigation", screen: ScreenNavigation, category: catPilot},

			// === Progress category — long-running goals ===
			{label: "Quests", screen: ScreenQuests, category: catProgress},
			{label: "Achievements", screen: ScreenAchievements, category: catProgress},
			{label: "Leaderboards", screen: ScreenLeaderboards, category: catProgress},
			{label: "News", screen: ScreenNews, category: catProgress},

			// === Social category — communication + relationships ===
			{label: "Chat", screen: ScreenChat, category: catSocial},
			{label: "Mail", screen: ScreenMail, category: catSocial},
			{label: "Players", screen: ScreenPlayers, category: catSocial},
			{label: "Trade (Player-to-Player)", screen: ScreenTrade, category: catSocial},

			// === Politics category — galactic-scale faction systems ===
			{label: "Factions", screen: ScreenFactions, category: catPolitics},
			{label: "Faction Wars", screen: ScreenFactionWars, category: catPolitics},
			{label: "Territory Map", screen: ScreenTerritoryMap, category: catPolitics},
			{label: "PvP Combat", screen: ScreenPvP, category: catPolitics},

			// === Markets category — commerce + bounties ===
			{label: "Marketplace", screen: ScreenMarketplace, category: catMarkets},
			{label: "Trade Routes", screen: ScreenTradeRoutes, category: catMarkets},
			{label: "Notifications", screen: ScreenNotifications, category: catMarkets},

			// === Help & Settings ===
			{label: "Help", screen: ScreenHelp, category: catHelp},
			{label: "Tutorials", screen: ScreenTutorial, category: catHelp},
			{label: "Settings", screen: ScreenSettings, category: catHelp},
			{label: "Admin Panel", screen: ScreenAdmin, category: catHelp, adminOnly: true},

			// === Top-level: legacy + quit ===
			//
			// Space View kept around so players who need the radar/
			// HUD overview can still get there until P1.4 folds those
			// affordances into the flight cockpit.
			{label: "  Space View (legacy)", screen: ScreenSpaceView},
			{label: "  Quit", action: func(m *Model) tea.Cmd { return tea.Quit }},
		},
	}
}

// isDocked reports whether the player is currently landed at a
// planet/station. Drives the showWhenDocked / showWhenInFlight
// filter and the Launch ⇄ Take Off label swap.
func (m Model) isDocked() bool {
	return m.player != nil && m.player.CurrentPlanet != nil
}

// playerIsAdmin checks the admin manager for the current player. A
// missing manager (test setups) reports false so the admin panel
// stays hidden by default.
func (m Model) playerIsAdmin() bool {
	if m.adminManager == nil || m.player == nil {
		return false
	}
	return m.adminManager.IsAdmin(m.playerID)
}

// visibleTopLevelItems filters the menu items down to what should
// appear when openCategory == "". Includes:
//   - All non-categorized items whose visibility matches dock state
//   - One synthetic entry per category that has at least one
//     visible item under it (so empty categories don't show up)
//
// The synthetic entries use a sentinel screen value (ScreenMainMenu)
// and an action that opens the category. updateMainMenu detects them
// by checking action != nil and category != "".
func (m Model) visibleTopLevelItems() []menuItem {
	docked := m.isDocked()
	out := make([]menuItem, 0, len(m.mainMenu.items))

	// Top-level (non-categorized) action items first.
	for _, it := range m.mainMenu.items {
		if it.category != catNone {
			continue
		}
		if !visibilityMatches(it.visibility, docked) {
			continue
		}
		out = append(out, it)
	}

	// Insert category headers for every category that has at least
	// one visible item under it, in `orderedCategories` order.
	admin := m.playerIsAdmin()
	for _, cat := range orderedCategories {
		if !categoryHasVisibleItems(m.mainMenu.items, cat, docked, admin) {
			continue
		}
		thisCat := cat // closure capture
		out = append(out, menuItem{
			label:    "▸ " + string(cat),
			category: thisCat,
			action: func(m *Model) tea.Cmd {
				// Clicking a category opens it (handled below in
				// updateMainMenuDispatch by detecting action !=
				// nil + category != "").
				m.mainMenu.openCategory = thisCat
				m.mainMenu.cursor = 0
				return nil
			},
		})
	}

	return out
}

// visibleCategoryItems returns items inside the given category that
// pass dock + admin filtering, plus a "← Back" action at the top so
// the player can return to the top-level list without ESC.
func (m Model) visibleCategoryItems(cat menuCategory) []menuItem {
	docked := m.isDocked()
	admin := m.playerIsAdmin()

	out := []menuItem{
		{
			label: "← Back",
			action: func(m *Model) tea.Cmd {
				m.mainMenu.openCategory = catNone
				m.mainMenu.cursor = 0
				return nil
			},
		},
	}
	for _, it := range m.mainMenu.items {
		if it.category != cat {
			continue
		}
		if !visibilityMatches(it.visibility, docked) {
			continue
		}
		if it.adminOnly && !admin {
			continue
		}
		out = append(out, it)
	}
	return out
}

func visibilityMatches(v menuVisibility, docked bool) bool {
	switch v {
	case showWhenDocked:
		return docked
	case showWhenInFlight:
		return !docked
	default:
		return true
	}
}

func categoryHasVisibleItems(items []menuItem, cat menuCategory, docked, admin bool) bool {
	for _, it := range items {
		if it.category != cat {
			continue
		}
		if !visibilityMatches(it.visibility, docked) {
			continue
		}
		if it.adminOnly && !admin {
			continue
		}
		return true
	}
	return false
}

// currentMenuView returns the slice the cursor is operating on right
// now — top-level when openCategory is empty, sub-menu otherwise.
// Centralizes the "what's on screen?" question so update + view
// don't drift.
func (m Model) currentMenuView() []menuItem {
	if m.mainMenu.openCategory == catNone {
		return m.visibleTopLevelItems()
	}
	return m.visibleCategoryItems(m.mainMenu.openCategory)
}

func (m Model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Self-start the news ticker when re-entering the main menu.
	// The top-level Update routes off-screen newsTickerMsg into
	// stopNewsTicker, which clears the active flag; the next non-
	// tick message here kicks it back on. The kicker runs in
	// parallel with whatever the user's message would normally
	// produce, so we batch the two commands.
	var kickerCmd tea.Cmd
	if _, isTick := msg.(newsTickerMsg); !isTick {
		var updated Model
		updated, kickerCmd = m.ensureNewsTickerTick()
		m = updated
	}

	model, cmd := m.updateMainMenuDispatch(msg)
	switch {
	case kickerCmd == nil:
		return model, cmd
	case cmd == nil:
		return model, kickerCmd
	default:
		return model, tea.Batch(cmd, kickerCmd)
	}
}

func (m Model) updateMainMenuDispatch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case newsTickerMsg:
		// Ticker advances only while the main menu is on-screen.
		// When the user navigates away, the next tea.Tick is
		// scheduled in-flight but will arrive at a different
		// screen's updater, which drops it silently — so the loop
		// self-terminates.
		newModel, cmd := m.updateNewsTicker()
		return newModel, cmd

	case tea.KeyMsg:
		view := m.currentMenuView()

		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "esc", "backspace":
			// ESC at top level → no-op (q quits).
			// ESC inside a category → back to top level.
			if m.mainMenu.openCategory != catNone {
				m.mainMenu.openCategory = catNone
				m.mainMenu.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.mainMenu.cursor > 0 {
				m.mainMenu.cursor--
			}

		case "down", "j":
			if m.mainMenu.cursor < len(view)-1 {
				m.mainMenu.cursor++
			}

		case "enter", " ":
			if m.mainMenu.cursor < 0 || m.mainMenu.cursor >= len(view) {
				return m, nil
			}
			selected := view[m.mainMenu.cursor]

			// Action items (including category openers, the Back
			// entry, Quit, and Take Off) run their action and
			// stop. They handle their own state mutation.
			if selected.action != nil {
				return m, selected.action(&m)
			}

			// Otherwise: navigate to the item's screen, with
			// per-screen initialization for screens that need
			// data fetched on first display.
			m.previousScreen = ScreenMainMenu
			m.hasPreviousScreen = true
			m.screen = selected.screen
			return m, m.initScreenForMenuSelection(selected.screen)
		}
	}

	return m, nil
}

// initScreenForMenuSelection returns a tea.Cmd that primes the
// destination screen (loads data, kicks off polls). Keeps the menu
// dispatcher's switch statement contained — every per-screen
// initialization rule lives here, not splattered across the dispatch
// case.
//
// Returns nil when no init is needed (most screens render purely
// from in-memory state).
func (m *Model) initScreenForMenuSelection(screen Screen) tea.Cmd {
	switch screen {
	case ScreenNavigation:
		m.navigation = newNavigationModel()
		return m.loadConnectedSystems()
	case ScreenTrading:
		m.trading = newTradingModel()
		return m.loadTradingMarket()
	case ScreenCargo:
		m.cargo = newCargoModel()
	case ScreenShipyard:
		m.shipyard = newShipyardModel()
		return m.loadShipyard()
	case ScreenOutfitter:
		m.outfitter = newOutfitterModel()
		return m.loadOutfitter()
	case ScreenOutfitterEnhanced:
		m.outfitterEnhanced = newOutfitterEnhancedModel()
		m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
		m.outfitterEnhanced.loadouts = m.outfittingManager.GetPlayerLoadouts(m.playerID)
	case ScreenShipManagement:
		m.shipManagement = newShipManagementModel()
		return m.loadOwnedShips()
	case ScreenLeaderboards:
		m.leaderboardsModel = newLeaderboardsModel()
		return m.refreshLeaderboards()
	case ScreenSettings:
		m.settingsModel = newSettingsModel()
		if m.settingsManager != nil {
			if ps, err := m.settingsManager.LoadSettings(m.playerID); err == nil {
				m.settingsModel.settings = ps
			}
		}
	case ScreenAdmin:
		m.adminModel = newAdminModel()
		if m.adminManager != nil {
			m.adminModel.isAdmin = m.adminManager.IsAdmin(m.playerID)
			if m.adminModel.isAdmin {
				m.adminModel.role = "moderator"
			}
		}
	case ScreenTutorial:
		m.tutorialModel = newTutorialModel()
		m.tutorialModel.viewMode = tutorialViewList
		if m.tutorialManager != nil {
			m.tutorialModel.allTutorials = m.tutorialManager.GetAllTutorials()
		}
	case ScreenQuests:
		m.questsModel = newQuestsModel()
		m.questsModel.viewMode = questViewActive
		if m.questManager != nil {
			m.questsModel.activeQuests = m.questManager.GetActiveQuests(m.playerID)
			m.questsModel.availableQuests = m.questManager.GetAvailableQuests(m.playerID)
			m.questsModel.completedQuests = m.questManager.GetCompletedQuests(m.playerID)
		}
	case ScreenSpaceView:
		m.spaceView = newSpaceViewModel()
		return tea.Batch(m.loadSpaceViewDataCmd(), spaceViewPollTick())
	case ScreenMail:
		m.mail.mode = mailModeInbox
		m.mail.selectedIndex = 0
		m.mail.loading = true
		return m.loadInbox()
	case ScreenChat:
		return chatPollTick()
	case ScreenPvP:
		return pvpPollTick()
	case ScreenMarketplace:
		return marketplacePollTick()
	}
	return nil
}

func (m Model) viewMainMenu() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	// Resolve the player's current location to a readable label.
	systemName := "Unknown"
	if m.player != nil {
		if m.currentSystem != nil {
			systemName = m.currentSystem.Name
		} else if m.player.CurrentSystem != uuid.Nil {
			systemName = "In transit"
		}
	}
	if m.isDocked() && m.currentPlanet != nil {
		systemName = m.currentPlanet.Name + ", " + systemName
	}

	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}

	var sb strings.Builder

	// Top border + titles.
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxTopRight + "\n")
	writeFramedLine(&sb, Center("TERMINAL VELOCITY", width-2))

	// Subtitle adapts to dock state + open category — gives the
	// player a clear "where am I?" cue at all times.
	subtitle := "= MAIN MENU ="
	if m.mainMenu.openCategory != catNone {
		subtitle = "= " + strings.ToUpper(string(m.mainMenu.openCategory)) + " ="
	} else if m.isDocked() {
		subtitle = "= STATION =" // visually distinct when landed
	}
	writeFramedLine(&sb, Center(subtitle, width-2))

	// Divider under the title.
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Player stats row.
	statsLine := fmt.Sprintf(" Pilot: %s   Credits: %s cr   Location: %s",
		m.username, formatThousands(credits), systemName)
	writeFramedLine(&sb, PadRight(statsLine, width-2))

	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Menu list — one item per row when in a sub-menu (so items
	// have room to read), two columns at top level when there are
	// >6 entries (older long-list density when collapsed view).
	view := m.currentMenuView()
	colWidth := width - 4
	twoCol := m.mainMenu.openCategory == catNone && len(view) > 8
	if twoCol {
		colWidth = (width - 4) / 2
		rows := (len(view) + 1) / 2
		for row := 0; row < rows; row++ {
			leftIdx := row
			rightIdx := row + rows
			left := renderMenuItem(view, leftIdx, m.mainMenu.cursor, colWidth)
			right := strings.Repeat(" ", colWidth)
			if rightIdx < len(view) {
				right = renderMenuItem(view, rightIdx, m.mainMenu.cursor, colWidth)
			}
			writeFramedLine(&sb, " "+left+" "+right)
		}
	} else {
		for i := range view {
			line := renderMenuItem(view, i, m.mainMenu.cursor, colWidth)
			writeFramedLine(&sb, " "+line+" ")
		}
	}

	// Spacer + footer divider.
	writeFramedLine(&sb, strings.Repeat(" ", width-2))
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Newsreel ticker — same as before, suppressed when news
	// manager has no content.
	ticker := m.renderNewsTicker(width - 4)
	if ticker != "" {
		writeFramedLine(&sb, " "+PadRight(ticker, width-3))
		sb.WriteString(BoxCrossLeft)
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
		sb.WriteString(BoxCross + "\n")
	}

	// Help text adapts to context.
	help := "↑/↓: Navigate   Enter: Select   q: Quit"
	if m.mainMenu.openCategory != catNone {
		help = "↑/↓: Navigate   Enter: Select   ESC: Back   q: Quit"
	}
	writeFramedLine(&sb, Center(help, width-2))

	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// writeFramedLine writes a single content row bracketed by the outer box
// borders. Content must already be padded to width-2 cells (cell width, not
// byte length) — callers typically produce it via Center or PadRight from
// ui_components.go.
func writeFramedLine(sb *strings.Builder, content string) {
	sb.WriteString(BoxVertical)
	sb.WriteString(content)
	sb.WriteString(BoxVertical + "\n")
}

// formatThousands turns 1234567 into "1,234,567". Used by anywhere
// the player sees a credit balance — the comma separators read
// faster than raw digit runs.
func formatThousands(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := fmt.Sprintf("%d", n)
	out := ""
	for i, r := range digits {
		// Insert a comma every 3 digits from the right; leftmost
		// group can be 1-3 digits depending on total length.
		if i > 0 && (len(digits)-i)%3 == 0 {
			out += ","
		}
		out += string(r)
	}
	if neg {
		out = "-" + out
	}
	return out
}

// renderMenuItem renders one menu item at the given cursor position, padded
// to columnWidth cells. The cursor row uses the same selected-item treatment
// as other menus; non-cursor rows use a two-space left gutter.
func renderMenuItem(items []menuItem, idx, cursor, columnWidth int) string {
	if idx >= len(items) {
		return strings.Repeat(" ", columnWidth)
	}
	label := items[idx].label
	if idx == cursor {
		// Selected: marker + cyan/bold label. Measure the raw label for
		// padding since ANSI escapes don't contribute to cell width.
		rendered := selectedMenuItemStyle.Render("> " + label)
		padSize := columnWidth - cellWidth("> "+label)
		if padSize < 0 {
			padSize = 0
		}
		return rendered + strings.Repeat(" ", padSize)
	}
	return PadRight("  "+label, columnWidth)
}
