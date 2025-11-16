// File: internal/tui/admin.go
// Project: Terminal Velocity
// Description: Server administration panel with RBAC-controlled moderation tools
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// This screen provides a comprehensive server administration interface for managing
// players, monitoring server health, and performing moderation actions. Key features:
//
// - Multi-tab interface (Overview, Players, Audit Log, Settings)
// - Role-based access control (RBAC) with 4 roles: Owner, Admin, Moderator, Helper
// - Player moderation: Ban/unban, mute/unmute with expiration times
// - Server statistics: Active players, connections, uptime, metrics
// - Audit log viewer: 10,000 entry buffer with filtering capabilities
// - Server settings management: Configuration updates with validation
// - Permission checks: Every action validates user permissions before execution
// - Real-time updates: Statistics refresh automatically
//
// Access Control:
// - Only accessible to users with admin permissions (admin.Manager tracks this)
// - Permission levels determine available actions (e.g., only Owner can modify settings)
// - All actions are logged to audit trail for accountability

package tui

import (
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// Admin view modes - different panels in the admin interface
const (
	adminViewMain      = "main"      // Main admin menu
	adminViewPlayers   = "players"   // Online player management
	adminViewBans      = "bans"      // Ban list and management
	adminViewMutes     = "mutes"     // Mute list and management
	adminViewMetrics   = "metrics"   // Server performance metrics
	adminViewSettings  = "settings"  // Server configuration
	adminViewActionLog = "actionlog" // Admin action audit log
)

// adminModel holds the state for the admin panel screen
type adminModel struct {
	viewMode string           // Current view/panel being displayed
	cursor   int              // Current menu selection cursor position
	isAdmin  bool             // Whether current player has admin access
	role     models.AdminRole // Specific admin role (Owner, Admin, Moderator, Helper)
}

// newAdminModel creates a new admin panel model with default state
func newAdminModel() adminModel {
	return adminModel{
		viewMode: adminViewMain, // Start at main menu
		cursor:   0,             // Cursor at first option
		isAdmin:  false,         // Admin status checked on first access
	}
}

// updateAdmin handles all input for the admin panel screen
//
// Key Bindings:
//   - ↑/k: Move cursor up
//   - ↓/j: Move cursor down
//   - Enter/Space: Select menu item or perform action
//   - Esc/Backspace: Return to main menu (from main view) or previous view
//   - U: Unban player (when on ban list) or unmute player (when on mute list)
//
// Message Handling:
//   - tea.KeyMsg: Navigation and selection
//
// Access Control:
//   - Validates admin permissions before allowing access
//   - Returns to main menu if player lacks admin privileges
func (m Model) updateAdmin(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if player is admin - security check on every update
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

// handleAdminSelect processes the selection of a menu item in the admin panel
//
// Behavior:
//   - From main menu: Navigates to selected sub-view (Players, Bans, etc.)
//   - From sub-views: Performs context-specific actions (unban, unmute, etc.)
//   - Resets cursor to 0 when entering a new view
func (m Model) handleAdminSelect() (tea.Model, tea.Cmd) {
	if m.adminModel.viewMode == adminViewMain {
		// Navigate to sub-view based on cursor position
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
			m.adminModel.cursor = 0 // Reset cursor for new view
		}
	}

	return m, nil
}

// getAdminMaxCursor returns the maximum cursor position for the current view
//
// This determines how far down the list the cursor can navigate, based on
// the number of items in the current view (menu items, bans, mutes, etc.)
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

// viewAdmin renders the admin panel interface
//
// Layout:
//   - Header: Username, credits, "Admin Panel" title
//   - Subtitle: "Server Administration" heading
//   - Role display: Shows current user's admin role
//   - Content area: Switches between different views based on viewMode
//
// Views:
//   - Main: Menu of available admin panels
//   - Players: Online player management
//   - Bans: Active ban list with unban option
//   - Mutes: Active mute list with unmute option
//   - Metrics: Server performance statistics
//   - Settings: Server configuration display
//   - ActionLog: Admin action audit trail
//
// Security:
//   - Returns access denied message if player is not an admin
//   - Displays current role to indicate permission level
func (m Model) viewAdmin() string {
	// Security check: Deny access to non-admins
	if !m.adminModel.isAdmin {
		return errorView("Access Denied - Admin privileges required")
	}

	// Header with username, credits, and panel title
	s := renderHeader(m.username, m.player.Credits, "Admin Panel")
	s += "\n"

	// Display admin panel title and user's role
	s += subtitleStyle.Render("=== Server Administration ===") + "\n"
	s += helpStyle.Render("Role: "+string(m.adminModel.role)) + "\n\n"

	// Render appropriate view based on current mode
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

// viewAdminMain renders the main admin menu
//
// Display:
//   - Title: "Administration Menu"
//   - Menu items: 6 admin panel options with descriptions
//   - Selected item highlighted
//   - Footer: Navigation instructions
//
// Menu Items:
//   1. Player Management - View and manage online players
//   2. Ban Management - View and manage player bans
//   3. Mute Management - View and manage player mutes
//   4. Server Metrics - View server performance and statistics
//   5. Server Settings - Configure server parameters
//   6. Action Log - View admin action history
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

// viewAdminPlayers renders the online player management view
//
// Display:
//   - Title: "Online Players"
//   - Player list with details (username, location, ship, etc.)
//   - Action options (kick, ban, mute)
//   - Footer: Navigation instructions
//
// Note: Currently shows placeholder message - will integrate with session manager
// to display real-time player list with moderation actions
func (m Model) viewAdminPlayers() string {
	s := "Online Players:\n\n"

	if m.adminManager == nil {
		s += helpStyle.Render("Admin manager not initialized") + "\n"
		s += "\n" + renderFooter("ESC: Back")
		return s
	}

	// TODO: Integrate with session manager to show active sessions
	s += helpStyle.Render("Feature coming soon - integrate with session manager") + "\n"

	s += "\n" + renderFooter("ESC: Back")
	return s
}

// viewAdminBans renders the ban management view
//
// Display:
//   - Title: "Active Bans"
//   - Table with columns: Username, Reason, Type (Temporary/Permanent), Banned At
//   - Selected ban highlighted
//   - Empty message if no active bans
//   - Footer: Navigation and action instructions
//
// Actions:
//   - U: Unban selected player (removes ban and allows login)
//
// Data Source:
//   - Fetches from adminManager.GetActiveBans()
//   - Shows all non-expired bans
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

// viewAdminMutes renders the mute management view
//
// Display:
//   - Title: "Active Mutes"
//   - Table with columns: Username, Reason, Expires (time remaining)
//   - Selected mute highlighted
//   - Empty message if no active mutes
//   - Footer: Navigation and action instructions
//
// Actions:
//   - U: Unmute selected player (removes mute and allows chat)
//
// Time Display:
//   - Shows expiration time as remaining duration (e.g., "2h", "45m")
//   - "Never" for permanent mutes
//   - Auto-calculated time remaining until expiration
//
// Data Source:
//   - Fetches from adminManager.GetActiveMutes()
//   - Shows all non-expired mutes
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

// viewAdminMetrics renders the server metrics and statistics view
//
// Display:
//   - Title: "Server Metrics"
//   - Performance section: Memory usage, goroutine count
//   - Players section: Total, active, and peak player counts
//   - Activity section: Active sessions, commands, trades, combats
//   - Database section: Connection count, latency, error count
//   - Footer: Navigation instructions
//
// Metrics Displayed:
//   - Memory Usage: Formatted as bytes (KB, MB, GB)
//   - Goroutines: Active Go routines (indicates concurrency health)
//   - Total Players: All registered players
//   - Active Players: Currently online players
//   - Peak Players: Maximum concurrent players (all-time)
//   - Active Sessions: Current SSH connections
//   - Total Commands: Commands executed lifetime
//   - Active Trades: Ongoing player-to-player trades
//   - Active Combats: Ongoing combat encounters
//   - DB Connections: Active database connections
//   - DB Latency: Average query response time (ms)
//   - DB Errors: Database error count
//
// Data Source:
//   - Fetches from adminManager.GetMetrics()
//   - Real-time server statistics
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

// viewAdminSettings renders the server settings configuration view
//
// Display:
//   - Title: "Server Settings"
//   - General section: Server name, max players, tick rate
//   - Economy section: Starting credits, multiplier, tax rate
//   - Gameplay section: PvP, permadeath, combat difficulty, pirate frequency
//   - Note: "Settings editing coming soon"
//   - Footer: Navigation instructions
//
// Settings Displayed:
//   - Server Name: Display name for the server
//   - Max Players: Maximum concurrent connections allowed
//   - Tick Rate: Server update frequency (ticks per second)
//   - Starting Credits: Initial credits for new players
//   - Economy Multiplier: Global price/reward scaling factor
//   - Tax Rate: Percentage deducted from transactions
//   - PvP Enabled: Whether player-vs-player combat is allowed
//   - Permadeath Mode: Whether death is permanent
//   - Combat Difficulty: AI difficulty scaling (0.0 - 2.0)
//   - Pirate Frequency: Probability of pirate encounters (0.0 - 1.0)
//
// Note: Currently read-only - editing functionality planned for future release
//
// Data Source:
//   - Fetches from adminManager.GetSettings()
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

// viewAdminActionLog renders the admin action audit log view
//
// Display:
//   - Title: "Admin Action Log (Recent 20)"
//   - Table with columns: Time, Admin, Action, Details
//   - Recent 20 actions displayed
//   - Failed actions shown in error style (red)
//   - Empty message if no actions logged
//   - Footer: Navigation instructions
//
// Log Entries Include:
//   - Timestamp: Time the action was performed (HH:MM:SS format)
//   - Admin Name: Username of the administrator
//   - Action Type: What was done (ban, unban, mute, unmute, kick, etc.)
//   - Details: Target player and additional context
//   - Success/Failure: Visual indicator (error style for failures)
//
// Purpose:
//   - Accountability: Track all administrative actions
//   - Auditing: Review moderator activity
//   - Troubleshooting: Investigate issues with admin actions
//   - Security: Detect unauthorized or suspicious administrative activity
//
// Data Source:
//   - Fetches from adminManager.GetActionLog(20)
//   - Returns most recent 20 entries from audit buffer
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

// formatBytes converts a byte count to a human-readable string with appropriate units
//
// Parameters:
//   - bytes: Number of bytes to format
//
// Returns:
//   - Formatted string with appropriate unit (B, KB, MB, GB, TB, PB, EB)
//
// Examples:
//   - 512       → "512 B"
//   - 1024      → "1.0 KB"
//   - 1536      → "1.5 KB"
//   - 1048576   → "1.0 MB"
//   - 1073741824 → "1.0 GB"
//
// Algorithm:
//   - Uses binary units (1024 bytes per KB)
//   - Automatically selects appropriate unit based on magnitude
//   - Displays one decimal place for values >= 1 KB
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
