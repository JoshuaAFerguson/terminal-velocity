// File: internal/tui/admin.go
// Project: Terminal Velocity
// Description: Admin UI and dashboard
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// Admin views

const (
	adminViewMain      = "main"
	adminViewPlayers   = "players"
	adminViewBans      = "bans"
	adminViewMutes     = "mutes"
	adminViewMetrics   = "metrics"
	adminViewSettings  = "settings"
	adminViewActionLog = "actionlog"
)

type adminModel struct {
	viewMode string
	cursor   int
	isAdmin  bool
	role     models.AdminRole
}

func newAdminModel() adminModel {
	return adminModel{
		viewMode: adminViewMain,
		cursor:   0,
		isAdmin:  false,
	}
}

func (m Model) updateAdmin(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if player is admin
	if !m.adminModel.isAdmin {
		m.screen = ScreenMainMenu
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.adminModel.viewMode == adminViewMain {
				m.screen = ScreenMainMenu
			} else {
				m.adminModel.viewMode = adminViewMain
				m.adminModel.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.adminModel.cursor > 0 {
				m.adminModel.cursor--
			}
			return m, nil

		case "down", "j":
			maxCursor := m.getAdminMaxCursor()
			if m.adminModel.cursor < maxCursor {
				m.adminModel.cursor++
			}
			return m, nil

		case "enter", " ":
			return m.handleAdminSelect()
		}
	}

	return m, nil
}

func (m Model) handleAdminSelect() (tea.Model, tea.Cmd) {
	if m.adminModel.viewMode == adminViewMain {
		// Navigate to sub-view
		views := []string{
			adminViewPlayers,
			adminViewBans,
			adminViewMutes,
			adminViewMetrics,
			adminViewSettings,
			adminViewActionLog,
		}
		if m.adminModel.cursor < len(views) {
			m.adminModel.viewMode = views[m.adminModel.cursor]
			m.adminModel.cursor = 0
		}
	}

	return m, nil
}

func (m Model) getAdminMaxCursor() int {
	switch m.adminModel.viewMode {
	case adminViewMain:
		return 5 // 6 menu items
	case adminViewPlayers:
		return 0 // View only for now
	case adminViewBans:
		if m.adminManager != nil {
			return len(m.adminManager.GetActiveBans()) - 1
		}
		return 0
	case adminViewMutes:
		if m.adminManager != nil {
			return len(m.adminManager.GetActiveMutes()) - 1
		}
		return 0
	case adminViewMetrics:
		return 0 // View only
	case adminViewSettings:
		return 0 // View only for now
	case adminViewActionLog:
		return 0 // View only
	}
	return 0
}

func (m Model) viewAdmin() string {
	if !m.adminModel.isAdmin {
		return errorView("Access Denied - Admin privileges required")
	}

	s := renderHeader(m.username, m.player.Credits, "Admin Panel")
	s += "\n"

	s += subtitleStyle.Render("=== Server Administration ===") + "\n"
	s += helpStyle.Render("Role: "+string(m.adminModel.role)) + "\n\n"

	switch m.adminModel.viewMode {
	case adminViewMain:
		s += m.viewAdminMain()
	case adminViewPlayers:
		s += m.viewAdminPlayers()
	case adminViewBans:
		s += m.viewAdminBans()
	case adminViewMutes:
		s += m.viewAdminMutes()
	case adminViewMetrics:
		s += m.viewAdminMetrics()
	case adminViewSettings:
		s += m.viewAdminSettings()
	case adminViewActionLog:
		s += m.viewAdminActionLog()
	}

	return s
}

func (m Model) viewAdminMain() string {
	s := "Administration Menu:\n\n"

	menu := []struct {
		name string
		desc string
	}{
		{"Player Management", "View and manage online players"},
		{"Ban Management", "View and manage player bans"},
		{"Mute Management", "View and manage player mutes"},
		{"Server Metrics", "View server performance and statistics"},
		{"Server Settings", "Configure server parameters"},
		{"Action Log", "View admin action history"},
	}

	for i, item := range menu {
		line := fmt.Sprintf("%s - %s", item.name, helpStyle.Render(item.desc))

		if i == m.adminModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Select  •  ESC: Back")
	return s
}

func (m Model) viewAdminPlayers() string {
	s := "Online Players:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	// Note: In production, would fetch from session manager
	s += helpStyle.Render("Feature coming soon - integrate with session manager") + "\n"

	s += "\n" + renderFooter("ESC: Back")
	return s
}

func (m Model) viewAdminBans() string {
	s := "Active Bans:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	bans := m.adminManager.GetActiveBans()

	if len(bans) == 0 {
		s += helpStyle.Render("No active bans") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	s += fmt.Sprintf("%-20s %-20s %-15s %s\n", "Username", "Reason", "Type", "Banned At")
	s += "─────────────────────────────────────────────────────────────────────────\n"

	for i, ban := range bans {
		banType := "Temporary"
		if ban.IsPermanent {
			banType = "Permanent"
		}

		line := fmt.Sprintf("%-20s %-20s %-15s %s",
			truncate(ban.Username, 20),
			truncate(ban.Reason, 20),
			banType,
			ban.BannedAt.Format("Jan 02 15:04"))

		if i == m.adminModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  U: Unban  •  ESC: Back")
	return s
}

func (m Model) viewAdminMutes() string {
	s := "Active Mutes:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	mutes := m.adminManager.GetActiveMutes()

	if len(mutes) == 0 {
		s += helpStyle.Render("No active mutes") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	s += fmt.Sprintf("%-20s %-25s %-20s\n", "Username", "Reason", "Expires")
	s += "─────────────────────────────────────────────────────────────────────\n"

	for i, mute := range mutes {
		expires := "Never"
		remaining := time.Until(mute.ExpiresAt)
		if remaining > 0 {
			if remaining < time.Hour {
				expires = fmt.Sprintf("%dm", int(remaining.Minutes()))
			} else {
				expires = fmt.Sprintf("%dh", int(remaining.Hours()))
			}
		}

		line := fmt.Sprintf("%-20s %-25s %-20s",
			truncate(mute.Username, 20),
			truncate(mute.Reason, 25),
			expires)

		if i == m.adminModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  U: Unmute  •  ESC: Back")
	return s
}

func (m Model) viewAdminMetrics() string {
	s := "Server Metrics:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	metrics := m.adminManager.GetMetrics()

	s += "Performance:\n"
	s += fmt.Sprintf("  Memory Usage:   %s\n", formatBytes(metrics.MemoryUsage))
	s += fmt.Sprintf("  Goroutines:     %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.GoroutineCount)))
	s += "\n"

	s += "Players:\n"
	s += fmt.Sprintf("  Total Players:  %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.TotalPlayers)))
	s += fmt.Sprintf("  Active Players: %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.ActivePlayers)))
	s += fmt.Sprintf("  Peak Players:   %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.PeakPlayers)))
	s += "\n"

	s += "Activity:\n"
	s += fmt.Sprintf("  Active Sessions: %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.ActiveSessions)))
	s += fmt.Sprintf("  Total Commands:  %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.TotalCommands)))
	s += fmt.Sprintf("  Active Trades:   %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.ActiveTrades)))
	s += fmt.Sprintf("  Active Combats:  %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.ActiveCombats)))
	s += "\n"

	s += "Database:\n"
	s += fmt.Sprintf("  Connections: %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.DBConnections)))
	s += fmt.Sprintf("  Latency:     %s ms\n", statsStyle.Render(fmt.Sprintf("%d", metrics.DBLatency)))
	s += fmt.Sprintf("  Errors:      %s\n", statsStyle.Render(fmt.Sprintf("%d", metrics.DBErrors)))

	s += "\n" + renderFooter("ESC: Back")
	return s
}

func (m Model) viewAdminSettings() string {
	s := "Server Settings:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	settings := m.adminManager.GetSettings()

	s += "General:\n"
	s += fmt.Sprintf("  Server Name:    %s\n", settings.ServerName)
	s += fmt.Sprintf("  Max Players:    %s\n", statsStyle.Render(fmt.Sprintf("%d", settings.MaxPlayers)))
	s += fmt.Sprintf("  Tick Rate:      %s/s\n", statsStyle.Render(fmt.Sprintf("%d", settings.TickRate)))
	s += "\n"

	s += "Economy:\n"
	s += fmt.Sprintf("  Starting Credits: %s cr\n", statsStyle.Render(fmt.Sprintf("%d", settings.StartingCredits)))
	s += fmt.Sprintf("  Economy Mult:     %s\n", statsStyle.Render(fmt.Sprintf("%.2f", settings.EconomyMultiplier)))
	s += fmt.Sprintf("  Tax Rate:         %s%%\n", statsStyle.Render(fmt.Sprintf("%.1f", settings.TaxRate*100)))
	s += "\n"

	s += "Gameplay:\n"
	s += fmt.Sprintf("  PvP Enabled:      %s\n", boolToString(settings.PvPEnabled))
	s += fmt.Sprintf("  Permadeath:       %s\n", boolToString(settings.PermadeathMode))
	s += fmt.Sprintf("  Combat Difficulty: %s\n", statsStyle.Render(fmt.Sprintf("%.2f", settings.CombatDifficulty)))
	s += fmt.Sprintf("  Pirate Frequency:  %s%%\n", statsStyle.Render(fmt.Sprintf("%.0f", settings.PirateFrequency*100)))

	s += "\n" + helpStyle.Render("(Settings editing coming soon)") + "\n"
	s += "\n" + renderFooter("ESC: Back")
	return s
}

func (m Model) viewAdminActionLog() string {
	s := "Admin Action Log (Recent 20):\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	actions := m.adminManager.GetActionLog(20)

	if len(actions) == 0 {
		s += helpStyle.Render("No actions logged") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	s += fmt.Sprintf("%-15s %-15s %-20s %s\n", "Time", "Admin", "Action", "Details")
	s += "─────────────────────────────────────────────────────────────────────────\n"

	for _, action := range actions {
		timeStr := action.Timestamp.Format("15:04:05")
		line := fmt.Sprintf("%-15s %-15s %-20s %s",
			timeStr,
			truncate(action.AdminName, 15),
			truncate(action.Action, 20),
			truncate(action.Details, 30))

		if action.Success {
			s += "  " + line + "\n"
		} else {
			s += "  " + errorStyle.Render(line) + "\n"
		}
	}

	s += "\n" + renderFooter("ESC: Back")
	return s
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
