// File: internal/tui/main_menu.go
// Project: Terminal Velocity
// Description: Terminal UI component for main_menu
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type mainMenuModel struct {
	cursor int
	items  []menuItem
}

type menuItem struct {
	label  string
	screen Screen
	action func(*Model) tea.Cmd
}

func newMainMenuModel() mainMenuModel {
	return mainMenuModel{
		cursor: 0,
		items: []menuItem{
			// Launch drops the player into the 2D space view — the actual
			// gameplay loop (radar, HUD, target cycling, land on planets).
			// The old text-hotkey hub at ScreenGame is still reachable via
			// the Mail / Trade Routes entries (which route back to it).
			{label: "Launch", screen: ScreenSpaceView},
			{label: "Navigation", screen: ScreenNavigation},
			{label: "Trading", screen: ScreenTrading},
			{label: "Cargo Hold", screen: ScreenCargo},
			{label: "Shipyard", screen: ScreenShipyard},
			{label: "Outfitter", screen: ScreenOutfitter},
			{label: "Advanced Outfitting", screen: ScreenOutfitterEnhanced},
			{label: "Ship Management", screen: ScreenShipManagement},
			{label: "Pilot Record", screen: ScreenPilotRecord},
			{label: "Missions", screen: ScreenMissions},
			{label: "Quests", screen: ScreenQuests},
			{label: "Achievements", screen: ScreenAchievements},
			{label: "Leaderboards", screen: ScreenLeaderboards},
			{label: "Players", screen: ScreenPlayers},
			{label: "Chat", screen: ScreenChat},
			{label: "Factions", screen: ScreenFactions},
			{label: "Trade", screen: ScreenTrade},
			{label: "Mail", screen: ScreenMail},
			{label: "Trade Routes", screen: ScreenTradeRoutes},
			{label: "Notifications", screen: ScreenNotifications},
			{label: "PvP Combat", screen: ScreenPvP},
			{label: "News", screen: ScreenNews},
			{label: "Help", screen: ScreenHelp},
			{label: "Settings", screen: ScreenSettings},
			{label: "Tutorials", screen: ScreenTutorial},
			{label: "Admin Panel", screen: ScreenAdmin},
			{label: "Quit", action: func(m *Model) tea.Cmd { return tea.Quit }},
		},
	}
}

func (m Model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Quit from main menu
			return m, tea.Quit
		case "up", "k":
			if m.mainMenu.cursor > 0 {
				m.mainMenu.cursor--
			}
		case "down", "j":
			if m.mainMenu.cursor < len(m.mainMenu.items)-1 {
				m.mainMenu.cursor++
			}
		case "enter", " ":
			selected := m.mainMenu.items[m.mainMenu.cursor]
			if selected.action != nil {
				return m, selected.action(&m)
			}
			// Record that we came from the main menu so screens reachable
			// from multiple entry points (outfitter_enhanced, etc.) can
			// return here on Esc instead of defaulting to space view.
			m.previousScreen = ScreenMainMenu
			m.hasPreviousScreen = true
			m.screen = selected.screen

			// Initialize screen-specific data
			if selected.screen == ScreenNavigation {
				m.navigation = newNavigationModel()
				return m, m.loadConnectedSystems()
			}
			if selected.screen == ScreenTrading {
				m.trading = newTradingModel()
				return m, m.loadTradingMarket()
			}
			if selected.screen == ScreenCargo {
				m.cargo = newCargoModel()
				return m, nil
			}
			if selected.screen == ScreenShipyard {
				m.shipyard = newShipyardModel()
				return m, m.loadShipyard()
			}
			if selected.screen == ScreenOutfitter {
				m.outfitter = newOutfitterModel()
				return m, m.loadOutfitter()
			}
			if selected.screen == ScreenOutfitterEnhanced {
				m.outfitterEnhanced = newOutfitterEnhancedModel()
				// Load player inventory and loadouts
				m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
				m.outfitterEnhanced.loadouts = m.outfittingManager.GetPlayerLoadouts(m.playerID)
				return m, nil
			}
			if selected.screen == ScreenShipManagement {
				m.shipManagement = newShipManagementModel()
				return m, m.loadOwnedShips()
			}
			if selected.screen == ScreenLeaderboards {
				m.leaderboardsModel = newLeaderboardsModel()
				return m, m.refreshLeaderboards()
			}
			if selected.screen == ScreenSettings {
				m.settingsModel = newSettingsModel()
				// Load player settings
				if playerSettings, err := m.settingsManager.LoadSettings(m.playerID); err == nil {
					m.settingsModel.settings = playerSettings
				}
				return m, nil
			}
			if selected.screen == ScreenAdmin {
				m.adminModel = newAdminModel()
				// Check if player is admin
				m.adminModel.isAdmin = m.adminManager.IsAdmin(m.playerID)
				if m.adminModel.isAdmin {
					// Get admin role from manager
					// For now, default to moderator
					m.adminModel.role = "moderator"
				}
				return m, nil
			}
			if selected.screen == ScreenTutorial {
				m.tutorialModel = newTutorialModel()
				m.tutorialModel.viewMode = tutorialViewList
				m.tutorialModel.allTutorials = m.tutorialManager.GetAllTutorials()
				return m, nil
			}
			if selected.screen == ScreenQuests {
				m.questsModel = newQuestsModel()
				m.questsModel.viewMode = questViewActive
				m.questsModel.activeQuests = m.questManager.GetActiveQuests(m.playerID)
				m.questsModel.availableQuests = m.questManager.GetAvailableQuests(m.playerID)
				m.questsModel.completedQuests = m.questManager.GetCompletedQuests(m.playerID)
				return m, nil
			}
			if selected.screen == ScreenSpaceView {
				// Fresh viewport state + kick off the async load so the
				// player sees their real system/planets/nearby ships on
				// the first frame instead of empty panels. Also arm the
				// 2-second poll so other players jumping in/out of the
				// system become visible without user input.
				m.spaceView = newSpaceViewModel()
				return m, tea.Batch(m.loadSpaceViewDataCmd(), spaceViewPollTick())
			}
			if selected.screen == ScreenMail {
				m.mail.mode = mailModeInbox
				m.mail.selectedIndex = 0
				m.mail.loading = true
				return m, m.loadInbox()
			}
			if selected.screen == ScreenChat {
				// Kick the poll tick so messages fanned into our history
				// by other sessions show up without requiring the user
				// to press a key.
				return m, chatPollTick()
			}
			if selected.screen == ScreenPvP {
				// Poll so incoming challenges + accept/decline state
				// updates surface without keystrokes. Same pattern as
				// chat and space view.
				return m, pvpPollTick()
			}

			return m, nil
		}
	}

	return m, nil
}

func (m Model) viewMainMenu() string {
	// The login screen uses a full-width heavy box with centered content;
	// keep the main menu in the same visual language so the login->menu
	// transition doesn't feel like two different apps.
	width := 80
	if m.width > 80 {
		width = m.width
	}

	// Resolve the player's current location to a readable label. The TUI
	// caches the last-loaded system on the model, so hitting the DB on
	// every render isn't necessary.
	systemName := "Unknown"
	if m.player != nil {
		if m.currentSystem != nil {
			systemName = m.currentSystem.Name
		} else if m.player.CurrentSystem != uuid.Nil {
			systemName = "In transit"
		}
	}

	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxTopRight + "\n")

	// Title rows
	writeFramedLine(&sb, Center("TERMINAL VELOCITY", width-2))
	writeFramedLine(&sb, Center("= MAIN MENU =", width-2))

	// Divider under the title
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Player stats row
	statsLine := fmt.Sprintf(" Pilot: %s   Credits: %s cr   Location: %s",
		m.username, formatThousands(credits), systemName)
	writeFramedLine(&sb, PadRight(statsLine, width-2))

	// Divider before the menu list
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Menu items in two columns. Paginate rows = ceil(n/2).
	items := m.mainMenu.items
	rows := (len(items) + 1) / 2
	colWidth := (width - 4) / 2 // 4 = 2 outer borders + 1 gutter char + 1 padding
	for row := 0; row < rows; row++ {
		leftIdx := row
		rightIdx := row + rows
		left := renderMenuItem(items, leftIdx, m.mainMenu.cursor, colWidth)
		right := ""
		if rightIdx < len(items) {
			right = renderMenuItem(items, rightIdx, m.mainMenu.cursor, colWidth)
		} else {
			right = strings.Repeat(" ", colWidth)
		}
		writeFramedLine(&sb, " "+left+" "+right)
	}

	// Empty spacer row + footer divider
	writeFramedLine(&sb, strings.Repeat(" ", width-2))
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxCross + "\n")

	// Help text
	writeFramedLine(&sb, Center("↑/↓ or j/k: Navigate   Enter: Select   q: Quit", width-2))

	// Bottom border
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

// formatThousands formats an int64 with thousands separators ("12,345").
// Inline rather than pulling in x/text to keep the menu render cheap.
func formatThousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	}
	if len(s) <= 3 {
		return sign + s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return sign + strings.Join(parts, ",")
}
