// File: internal/tui/settings.go
// Project: Terminal Velocity
// Description: Settings screen - Player preferences and configuration across 6 categories
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The settings screen provides:
// - 6 configuration categories (Display, Audio, Gameplay, Controls, Privacy, Notifications)
// - Boolean toggle settings with Enter/Space
// - Multi-option settings with up/down cycling
// - Settings persistence to database (JSON format)
// - Reset to defaults functionality
// - Real-time setting changes
//
// Categories:
//   - Display: Color scheme, animations, compact mode, tutorial tips, icons
//   - Audio: Audio enable, sound effects, music, notifications (not yet implemented)
//   - Gameplay: Auto-save, confirmations, damage numbers, autopilot, encounters, fast travel, tutorial mode, difficulty, permadeath
//   - Controls: Keybindings for movement, actions, combat (view-only, customization coming soon)
//   - Privacy: Online status, location, ship info visibility, trade/PvP/party requests, blocklist, friends list
//   - Notifications: Achievement, level up, trade, combat, player joined, news, encounters, system messages, chat notifications
//
// Display Settings:
//   - Color Scheme: default, dark, light, high_contrast, colorblind
//   - Show Animations: ON/OFF
//   - Compact Mode: ON/OFF
//   - Show Tutorial Tips: ON/OFF
//   - Show Icons: ON/OFF
//
// Gameplay Settings:
//   - Auto-Save: ON/OFF
//   - Confirm Dangerous Actions: ON/OFF
//   - Show Damage Numbers: ON/OFF
//   - Auto-Pilot Hints: ON/OFF
//   - Pause on Encounter: ON/OFF
//   - Fast Travel: ON/OFF
//   - Tutorial Mode: ON/OFF
//   - Difficulty Level: easy, normal, hard, expert
//   - Permadeath Mode: ON/OFF
//
// Privacy Settings:
//   - Show Online Status: ON/OFF
//   - Show Location: ON/OFF
//   - Show Ship Info: ON/OFF
//   - Allow Trade Requests: ON/OFF
//   - Allow PvP Challenges: ON/OFF
//   - Allow Party Invites: ON/OFF
//   - Blocked Players count
//   - Friends list count
//
// Visual Features:
//   - [EDITING] indicator on active setting
//   - ON/OFF display for booleans
//   - Multi-option values shown inline
//   - Active setting highlighted
//   - Help text for categories

package tui

import (
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// Settings views

const (
	settingsViewMain          = "main"
	settingsViewDisplay       = "display"
	settingsViewAudio         = "audio"
	settingsViewGameplay      = "gameplay"
	settingsViewControls      = "controls"
	settingsViewPrivacy       = "privacy"
	settingsViewNotifications = "notifications"
)

// settingsModel contains the state for the settings screen.
// Manages category navigation, setting editing, and persistence.
type settingsModel struct {
	viewMode string           // Current view: "main" or category name
	cursor   int              // Current cursor position in setting list
	editing  bool             // True when editing a setting value
	settings *models.Settings // Player's current settings (loaded from database)
}

// newSettingsModel creates and initializes a new settings screen model.
// Starts in main category selection view.
func newSettingsModel() settingsModel {
	return settingsModel{
		viewMode: settingsViewMain,
		cursor:   0,
		editing:  false,
	}
}

// updateSettings handles input and state updates for the settings screen.
//
// Key Bindings (Category/List Mode):
//   - esc/backspace: Return to main menu (or main from category)
//   - up/k: Move cursor up in setting list
//   - down/j: Move cursor down in setting list
//   - enter/space: Edit selected setting (or select category in main)
//   - r: Reset current category to defaults
//
// Key Bindings (Editing Mode):
//   - esc: Cancel editing
//   - enter/space: Confirm value change
//
// Settings are automatically saved to database when changed.
func (m Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.settingsModel.editing {
			return m.updateSettingsEditing(msg)
		}

		switch msg.String() {
		case "esc", "backspace":
			if m.settingsModel.viewMode == settingsViewMain {
				m.screen = ScreenMainMenu
			} else {
				m.settingsModel.viewMode = settingsViewMain
				m.settingsModel.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.settingsModel.cursor > 0 {
				m.settingsModel.cursor--
			}
			return m, nil

		case "down", "j":
			maxCursor := m.getSettingsMaxCursor()
			if m.settingsModel.cursor < maxCursor {
				m.settingsModel.cursor++
			}
			return m, nil

		case "enter", " ":
			return m.handleSettingsSelect()

		case "r":
			// Reset to defaults
			if m.settingsModel.viewMode != settingsViewMain {
				m.settingsManager.ResetToDefaults(m.playerID)
				m.settingsModel.settings, _ = m.settingsManager.GetSettings(m.playerID)
			}
			return m, nil
		}
	}

	return m, nil
}

// updateSettingsEditing handles input when editing a setting value.
// Toggles boolean values or cycles through multi-option values.
func (m Model) updateSettingsEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle editing mode (toggle boolean values)
	switch msg.String() {
	case "esc":
		m.settingsModel.editing = false
		return m, nil

	case "enter", " ":
		m.toggleSettingValue()
		m.settingsModel.editing = false
		return m, nil
	}

	return m, nil
}

// handleSettingsSelect handles Enter/Space key in main or category view.
// Navigates to category or enters editing mode for a setting.
func (m Model) handleSettingsSelect() (tea.Model, tea.Cmd) {
	if m.settingsModel.viewMode == settingsViewMain {
		// Navigate to category
		categories := []string{"display", "audio", "gameplay", "controls", "privacy", "notifications"}
		if m.settingsModel.cursor < len(categories) {
			m.settingsModel.viewMode = categories[m.settingsModel.cursor]
			m.settingsModel.cursor = 0
		}
	} else {
		// Edit setting
		m.settingsModel.editing = true
	}

	return m, nil
}

// toggleSettingValue toggles the value of the currently selected setting.
// Handles boolean toggles and multi-option cycling based on category.
// Automatically saves updated settings to database.
func (m *Model) toggleSettingValue() {
	if m.settingsModel.settings == nil {
		return
	}

	switch m.settingsModel.viewMode {
	case settingsViewDisplay:
		m.toggleDisplaySetting()
	case settingsViewAudio:
		m.toggleAudioSetting()
	case settingsViewGameplay:
		m.toggleGameplaySetting()
	case settingsViewControls:
		m.toggleControlSetting()
	case settingsViewPrivacy:
		m.togglePrivacySetting()
	case settingsViewNotifications:
		m.toggleNotificationSetting()
	}

	// Save updated settings
	m.settingsManager.UpdateSettings(m.playerID, func(s *models.Settings) {
		*s = *m.settingsModel.settings
	})
}

// toggleDisplaySetting toggles display category settings.
// Handles color scheme cycling and boolean toggles.
func (m *Model) toggleDisplaySetting() {
	switch m.settingsModel.cursor {
	case 0: // Color scheme
		schemes := []string{"default", "dark", "light", "high_contrast", "colorblind"}
		for i, scheme := range schemes {
			if scheme == m.settingsModel.settings.Display.ColorScheme {
				m.settingsModel.settings.Display.ColorScheme = schemes[(i+1)%len(schemes)]
				break
			}
		}
	case 1:
		m.settingsModel.settings.Display.ShowAnimations = !m.settingsModel.settings.Display.ShowAnimations
	case 2:
		m.settingsModel.settings.Display.CompactMode = !m.settingsModel.settings.Display.CompactMode
	case 3:
		m.settingsModel.settings.Display.ShowTutorialTips = !m.settingsModel.settings.Display.ShowTutorialTips
	case 4:
		m.settingsModel.settings.Display.ShowIcons = !m.settingsModel.settings.Display.ShowIcons
	}
}

// toggleAudioSetting toggles audio category settings.
func (m *Model) toggleAudioSetting() {
	switch m.settingsModel.cursor {
	case 0:
		m.settingsModel.settings.Audio.Enabled = !m.settingsModel.settings.Audio.Enabled
	case 1:
		m.settingsModel.settings.Audio.SoundEffects = !m.settingsModel.settings.Audio.SoundEffects
	case 2:
		m.settingsModel.settings.Audio.Music = !m.settingsModel.settings.Audio.Music
	case 3:
		m.settingsModel.settings.Audio.Notifications = !m.settingsModel.settings.Audio.Notifications
	}
}

// toggleGameplaySetting toggles gameplay category settings.
// Handles difficulty cycling and boolean toggles.
func (m *Model) toggleGameplaySetting() {
	switch m.settingsModel.cursor {
	case 0:
		m.settingsModel.settings.Gameplay.AutoSave = !m.settingsModel.settings.Gameplay.AutoSave
	case 1:
		m.settingsModel.settings.Gameplay.ConfirmDangerousActions = !m.settingsModel.settings.Gameplay.ConfirmDangerousActions
	case 2:
		m.settingsModel.settings.Gameplay.ShowDamageNumbers = !m.settingsModel.settings.Gameplay.ShowDamageNumbers
	case 3:
		m.settingsModel.settings.Gameplay.AutoPilot = !m.settingsModel.settings.Gameplay.AutoPilot
	case 4:
		m.settingsModel.settings.Gameplay.PauseOnEncounter = !m.settingsModel.settings.Gameplay.PauseOnEncounter
	case 5:
		m.settingsModel.settings.Gameplay.FastTravel = !m.settingsModel.settings.Gameplay.FastTravel
	case 6:
		m.settingsModel.settings.Gameplay.TutorialMode = !m.settingsModel.settings.Gameplay.TutorialMode
	case 7: // Difficulty
		difficulties := []string{"easy", "normal", "hard", "expert"}
		for i, diff := range difficulties {
			if diff == m.settingsModel.settings.Gameplay.DifficultyLevel {
				m.settingsModel.settings.Gameplay.DifficultyLevel = difficulties[(i+1)%len(difficulties)]
				break
			}
		}
	case 8:
		m.settingsModel.settings.Gameplay.PermadeathMode = !m.settingsModel.settings.Gameplay.PermadeathMode
	}
}

// toggleControlSetting toggles control category settings.
// Currently view-only; keybinding customization coming soon.
func (m *Model) toggleControlSetting() {
	// Controls would require more complex input handling
	// For now, just show them
}

// togglePrivacySetting toggles privacy category settings.
func (m *Model) togglePrivacySetting() {
	switch m.settingsModel.cursor {
	case 0:
		m.settingsModel.settings.Privacy.ShowOnline = !m.settingsModel.settings.Privacy.ShowOnline
	case 1:
		m.settingsModel.settings.Privacy.ShowLocation = !m.settingsModel.settings.Privacy.ShowLocation
	case 2:
		m.settingsModel.settings.Privacy.ShowShip = !m.settingsModel.settings.Privacy.ShowShip
	case 3:
		m.settingsModel.settings.Privacy.AllowTradeRequests = !m.settingsModel.settings.Privacy.AllowTradeRequests
	case 4:
		m.settingsModel.settings.Privacy.AllowPvPChallenges = !m.settingsModel.settings.Privacy.AllowPvPChallenges
	case 5:
		m.settingsModel.settings.Privacy.AllowPartyInvites = !m.settingsModel.settings.Privacy.AllowPartyInvites
	}
}

// toggleNotificationSetting toggles notification category settings.
func (m *Model) toggleNotificationSetting() {
	switch m.settingsModel.cursor {
	case 0:
		m.settingsModel.settings.Notifications.ShowAchievements = !m.settingsModel.settings.Notifications.ShowAchievements
	case 1:
		m.settingsModel.settings.Notifications.ShowLevelUp = !m.settingsModel.settings.Notifications.ShowLevelUp
	case 2:
		m.settingsModel.settings.Notifications.ShowTradeComplete = !m.settingsModel.settings.Notifications.ShowTradeComplete
	case 3:
		m.settingsModel.settings.Notifications.ShowCombatLog = !m.settingsModel.settings.Notifications.ShowCombatLog
	case 4:
		m.settingsModel.settings.Notifications.ShowPlayerJoined = !m.settingsModel.settings.Notifications.ShowPlayerJoined
	case 5:
		m.settingsModel.settings.Notifications.ShowNewsUpdates = !m.settingsModel.settings.Notifications.ShowNewsUpdates
	case 6:
		m.settingsModel.settings.Notifications.ShowEncounters = !m.settingsModel.settings.Notifications.ShowEncounters
	case 7:
		m.settingsModel.settings.Notifications.ShowSystemMessages = !m.settingsModel.settings.Notifications.ShowSystemMessages
	case 8:
		m.settingsModel.settings.Notifications.ChatNotifications = !m.settingsModel.settings.Notifications.ChatNotifications
	}
}

// getSettingsMaxCursor returns the maximum cursor position for current view.
// Different categories have different numbers of settings.
func (m Model) getSettingsMaxCursor() int {
	switch m.settingsModel.viewMode {
	case settingsViewMain:
		return 5 // 6 categories
	case settingsViewDisplay:
		return 4 // 5 settings
	case settingsViewAudio:
		return 3 // 4 settings
	case settingsViewGameplay:
		return 8 // 9 settings
	case settingsViewControls:
		return 0 // View only for now
	case settingsViewPrivacy:
		return 5 // 6 settings
	case settingsViewNotifications:
		return 8 // 9 settings
	}
	return 0
}

// viewSettings renders the settings screen (dispatches to category-specific views).
func (m Model) viewSettings() string {
	s := renderHeader(m.username, m.player.Credits, "Settings")
	s += "\n"

	s += subtitleStyle.Render("=== Settings ===") + "\n\n"

	switch m.settingsModel.viewMode {
	case settingsViewMain:
		s += m.viewSettingsMain()
	case settingsViewDisplay:
		s += m.viewSettingsDisplay()
	case settingsViewAudio:
		s += m.viewSettingsAudio()
	case settingsViewGameplay:
		s += m.viewSettingsGameplay()
	case settingsViewControls:
		s += m.viewSettingsControls()
	case settingsViewPrivacy:
		s += m.viewSettingsPrivacy()
	case settingsViewNotifications:
		s += m.viewSettingsNotifications()
	}

	return s
}

// viewSettingsMain renders the main category selection view.
func (m Model) viewSettingsMain() string {
	s := "Select a category:\n\n"

	categories := []struct {
		name string
		desc string
	}{
		{"Display", "Visual appearance and UI options"},
		{"Audio", "Sound effects and music (not yet implemented)"},
		{"Gameplay", "Game behavior and difficulty"},
		{"Controls", "Keybindings and input settings"},
		{"Privacy", "Visibility and social settings"},
		{"Notifications", "Alert and message preferences"},
	}

	for i, cat := range categories {
		line := fmt.Sprintf("%s - %s", cat.name, helpStyle.Render(cat.desc))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Open  •  ESC: Back")
	return s
}

// viewSettingsDisplay renders the display settings category.
func (m Model) viewSettingsDisplay() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Display Settings:\n\n"

	settings := []struct {
		name  string
		value string
	}{
		{"Color Scheme", m.settingsModel.settings.Display.ColorScheme},
		{"Show Animations", boolToString(m.settingsModel.settings.Display.ShowAnimations)},
		{"Compact Mode", boolToString(m.settingsModel.settings.Display.CompactMode)},
		{"Show Tutorial Tips", boolToString(m.settingsModel.settings.Display.ShowTutorialTips)},
		{"Show Icons", boolToString(m.settingsModel.settings.Display.ShowIcons)},
	}

	for i, setting := range settings {
		line := fmt.Sprintf("%-25s %s", setting.name, statsStyle.Render(setting.value))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line)
			if m.settingsModel.editing {
				s += " " + helpStyle.Render("[EDITING]")
			}
			s += "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back")
	return s
}

// viewSettingsAudio renders the audio settings category.
// Audio not yet implemented, settings shown as disabled.
func (m Model) viewSettingsAudio() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Audio Settings:\n\n"
	s += helpStyle.Render("(Audio not yet implemented)") + "\n\n"

	settings := []struct {
		name  string
		value string
	}{
		{"Audio Enabled", boolToString(m.settingsModel.settings.Audio.Enabled)},
		{"Sound Effects", boolToString(m.settingsModel.settings.Audio.SoundEffects)},
		{"Music", boolToString(m.settingsModel.settings.Audio.Music)},
		{"Notifications", boolToString(m.settingsModel.settings.Audio.Notifications)},
	}

	for i, setting := range settings {
		line := fmt.Sprintf("%-25s %s", setting.name, statsStyle.Render(setting.value))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + helpStyle.Render(line) + "\n"
		}
	}

	s += "\n" + renderFooter("ESC: Back")
	return s
}

// viewSettingsGameplay renders the gameplay settings category.
func (m Model) viewSettingsGameplay() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Gameplay Settings:\n\n"

	settings := []struct {
		name  string
		value string
	}{
		{"Auto-Save", boolToString(m.settingsModel.settings.Gameplay.AutoSave)},
		{"Confirm Dangerous Actions", boolToString(m.settingsModel.settings.Gameplay.ConfirmDangerousActions)},
		{"Show Damage Numbers", boolToString(m.settingsModel.settings.Gameplay.ShowDamageNumbers)},
		{"Auto-Pilot Hints", boolToString(m.settingsModel.settings.Gameplay.AutoPilot)},
		{"Pause on Encounter", boolToString(m.settingsModel.settings.Gameplay.PauseOnEncounter)},
		{"Fast Travel", boolToString(m.settingsModel.settings.Gameplay.FastTravel)},
		{"Tutorial Mode", boolToString(m.settingsModel.settings.Gameplay.TutorialMode)},
		{"Difficulty Level", m.settingsModel.settings.Gameplay.DifficultyLevel},
		{"Permadeath Mode", boolToString(m.settingsModel.settings.Gameplay.PermadeathMode)},
	}

	for i, setting := range settings {
		line := fmt.Sprintf("%-30s %s", setting.name, statsStyle.Render(setting.value))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line)
			if m.settingsModel.editing {
				s += " " + helpStyle.Render("[EDITING]")
			}
			s += "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back")
	return s
}

// viewSettingsControls renders the controls settings category.
// Currently view-only; keybinding customization coming soon.
func (m Model) viewSettingsControls() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Control Settings:\n\n"
	s += helpStyle.Render("(Keybinding customization coming soon)") + "\n\n"

	s += "Navigation:\n"
	s += fmt.Sprintf("  Move Up:    %s\n", m.settingsModel.settings.Controls.MoveUp)
	s += fmt.Sprintf("  Move Down:  %s\n", m.settingsModel.settings.Controls.MoveDown)
	s += fmt.Sprintf("  Move Left:  %s\n", m.settingsModel.settings.Controls.MoveLeft)
	s += fmt.Sprintf("  Move Right: %s\n\n", m.settingsModel.settings.Controls.MoveRight)

	s += "Actions:\n"
	s += fmt.Sprintf("  Confirm:  %s\n", m.settingsModel.settings.Controls.Confirm)
	s += fmt.Sprintf("  Cancel:   %s\n", m.settingsModel.settings.Controls.Cancel)
	s += fmt.Sprintf("  Back:     %s\n", m.settingsModel.settings.Controls.Back)
	s += fmt.Sprintf("  Help:     %s\n\n", m.settingsModel.settings.Controls.Help)

	s += "Combat:\n"
	s += fmt.Sprintf("  Attack:  %s\n", m.settingsModel.settings.Controls.Attack)
	s += fmt.Sprintf("  Defend:  %s\n", m.settingsModel.settings.Controls.Defend)
	s += fmt.Sprintf("  Flee:    %s\n", m.settingsModel.settings.Controls.Flee)

	s += "\n" + renderFooter("ESC: Back")
	return s
}

// viewSettingsPrivacy renders the privacy settings category.
func (m Model) viewSettingsPrivacy() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Privacy Settings:\n\n"

	settings := []struct {
		name  string
		value string
	}{
		{"Show Online Status", boolToString(m.settingsModel.settings.Privacy.ShowOnline)},
		{"Show Location", boolToString(m.settingsModel.settings.Privacy.ShowLocation)},
		{"Show Ship Info", boolToString(m.settingsModel.settings.Privacy.ShowShip)},
		{"Allow Trade Requests", boolToString(m.settingsModel.settings.Privacy.AllowTradeRequests)},
		{"Allow PvP Challenges", boolToString(m.settingsModel.settings.Privacy.AllowPvPChallenges)},
		{"Allow Party Invites", boolToString(m.settingsModel.settings.Privacy.AllowPartyInvites)},
	}

	for i, setting := range settings {
		line := fmt.Sprintf("%-25s %s", setting.name, statsStyle.Render(setting.value))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line)
			if m.settingsModel.editing {
				s += " " + helpStyle.Render("[EDITING]")
			}
			s += "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n"
	s += fmt.Sprintf("Blocked Players: %s\n", statsStyle.Render(fmt.Sprintf("%d", len(m.settingsModel.settings.Privacy.BlockList))))
	s += fmt.Sprintf("Friends: %s\n", statsStyle.Render(fmt.Sprintf("%d", len(m.settingsModel.settings.Privacy.FriendsList))))

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back")
	return s
}

// viewSettingsNotifications renders the notifications settings category.
func (m Model) viewSettingsNotifications() string {
	if m.settingsModel.settings == nil {
		return helpStyle.Render("Settings not loaded") + "\n"
	}

	s := "Notification Settings:\n\n"

	settings := []struct {
		name  string
		value string
	}{
		{"Achievement Unlocks", boolToString(m.settingsModel.settings.Notifications.ShowAchievements)},
		{"Level Up", boolToString(m.settingsModel.settings.Notifications.ShowLevelUp)},
		{"Trade Complete", boolToString(m.settingsModel.settings.Notifications.ShowTradeComplete)},
		{"Combat Log", boolToString(m.settingsModel.settings.Notifications.ShowCombatLog)},
		{"Player Join/Leave", boolToString(m.settingsModel.settings.Notifications.ShowPlayerJoined)},
		{"News Updates", boolToString(m.settingsModel.settings.Notifications.ShowNewsUpdates)},
		{"Random Encounters", boolToString(m.settingsModel.settings.Notifications.ShowEncounters)},
		{"System Messages", boolToString(m.settingsModel.settings.Notifications.ShowSystemMessages)},
		{"Chat Messages", boolToString(m.settingsModel.settings.Notifications.ChatNotifications)},
	}

	for i, setting := range settings {
		line := fmt.Sprintf("%-25s %s", setting.name, statsStyle.Render(setting.value))

		if i == m.settingsModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line)
			if m.settingsModel.editing {
				s += " " + helpStyle.Render("[EDITING]")
			}
			s += "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: Edit  •  R: Reset  •  ESC: Back")
	return s
}

// boolToString converts a boolean to ON/OFF string for display.
func boolToString(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
