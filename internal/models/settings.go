// File: internal/models/settings.go
// Project: Terminal Velocity
// Description: Player settings and configuration
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"time"

	"github.com/google/uuid"
)

// Settings represents player-specific configuration

type Settings struct {
	ID        uuid.UUID `json:"id"`
	PlayerID  uuid.UUID `json:"player_id"`
	UpdatedAt time.Time `json:"updated_at"`

	// Display settings
	Display DisplaySettings `json:"display"`

	// Audio settings
	Audio AudioSettings `json:"audio"`

	// Gameplay settings
	Gameplay GameplaySettings `json:"gameplay"`

	// Control settings
	Controls ControlSettings `json:"controls"`

	// Privacy settings
	Privacy PrivacySettings `json:"privacy"`

	// Notification settings
	Notifications NotificationSettings `json:"notifications"`
}

// DisplaySettings controls visual presentation
type DisplaySettings struct {
	ColorScheme      string `json:"color_scheme"`       // default, dark, light, high_contrast, colorblind
	ShowAnimations   bool   `json:"show_animations"`    // Enable/disable animations
	CompactMode      bool   `json:"compact_mode"`       // Reduce spacing and padding
	ShowTutorialTips bool   `json:"show_tutorial_tips"` // Display helpful tips
	ShowIcons        bool   `json:"show_icons"`         // Display emoji/unicode icons
	TerminalWidth    int    `json:"terminal_width"`     // Preferred width (0 = auto)
	TerminalHeight   int    `json:"terminal_height"`    // Preferred height (0 = auto)
}

// AudioSettings controls sound effects (future implementation)
type AudioSettings struct {
	Enabled       bool `json:"enabled"`       // Master audio toggle
	SoundEffects  bool `json:"sound_effects"` // UI sounds, combat sounds
	Music         bool `json:"music"`         // Background music
	Notifications bool `json:"notifications"` // Audio notifications
	Volume        int  `json:"volume"`        // 0-100
}

// GameplaySettings controls game behavior
type GameplaySettings struct {
	AutoSave                bool   `json:"auto_save"`           // Auto-save progress
	AutoSaveInterval        int    `json:"auto_save_interval"`  // Minutes between auto-saves
	ConfirmDangerousActions bool   `json:"confirm_dangerous"`   // Confirm risky actions
	ShowDamageNumbers       bool   `json:"show_damage_numbers"` // Display damage in combat
	AutoPilot               bool   `json:"auto_pilot"`          // Auto-navigation hints
	PauseOnEncounter        bool   `json:"pause_on_encounter"`  // Pause when encounter occurs
	FastTravel              bool   `json:"fast_travel"`         // Skip jump animations
	TutorialMode            bool   `json:"tutorial_mode"`       // Enable tutorial hints
	DifficultyLevel         string `json:"difficulty_level"`    // easy, normal, hard, expert
	PermadeathMode          bool   `json:"permadeath_mode"`     // Permadeath enabled
}

// ControlSettings manages keybindings
type ControlSettings struct {
	// Navigation
	MoveUp    string `json:"move_up"`    // Default: "up", "k"
	MoveDown  string `json:"move_down"`  // Default: "down", "j"
	MoveLeft  string `json:"move_left"`  // Default: "left", "h"
	MoveRight string `json:"move_right"` // Default: "right", "l"

	// Actions
	Confirm  string `json:"confirm"`  // Default: "enter", "space"
	Cancel   string `json:"cancel"`   // Default: "esc"
	Back     string `json:"back"`     // Default: "backspace"
	Help     string `json:"help"`     // Default: "?"
	Settings string `json:"settings"` // Default: "s"

	// Shortcuts
	QuickSave       string `json:"quick_save"`       // Default: "F5"
	QuickLoad       string `json:"quick_load"`       // Default: "F9"
	ToggleMap       string `json:"toggle_map"`       // Default: "m"
	ToggleInventory string `json:"toggle_inventory"` // Default: "i"
	ToggleShip      string `json:"toggle_ship"`      // Default: "v"
	ToggleChat      string `json:"toggle_chat"`      // Default: "c"

	// Combat
	Attack     string `json:"attack"`      // Default: "a"
	Defend     string `json:"defend"`      // Default: "d"
	Flee       string `json:"flee"`        // Default: "f"
	UseItem    string `json:"use_item"`    // Default: "u"
	NextTarget string `json:"next_target"` // Default: "tab"
	PrevTarget string `json:"prev_target"` // Default: "shift+tab"

	// Custom keybinding mode
	VimMode   bool `json:"vim_mode"`   // Enable vim-style navigation
	EmacsMode bool `json:"emacs_mode"` // Enable emacs-style navigation
}

// PrivacySettings controls visibility to other players
type PrivacySettings struct {
	ShowOnline         bool        `json:"show_online"`          // Appear online to others
	ShowLocation       bool        `json:"show_location"`        // Show current location
	ShowShip           bool        `json:"show_ship"`            // Show ship info
	AllowTradeRequests bool        `json:"allow_trade_requests"` // Accept trade requests
	AllowPvPChallenges bool        `json:"allow_pvp_challenges"` // Accept PvP challenges
	AllowPartyInvites  bool        `json:"allow_party_invites"`  // Accept party invitations
	BlockList          []uuid.UUID `json:"block_list"`           // Blocked players
	FriendsList        []uuid.UUID `json:"friends_list"`         // Friends list
}

// NotificationSettings controls alerts and messages
type NotificationSettings struct {
	ShowAchievements     bool `json:"show_achievements"`     // Achievement unlock notifications
	ShowLevelUp          bool `json:"show_level_up"`         // Level up notifications
	ShowTradeComplete    bool `json:"show_trade_complete"`   // Trade completion alerts
	ShowCombatLog        bool `json:"show_combat_log"`       // Combat event log
	ShowPlayerJoined     bool `json:"show_player_joined"`    // Player join/leave notifications
	ShowNewsUpdates      bool `json:"show_news_updates"`     // News article notifications
	ShowEncounters       bool `json:"show_encounters"`       // Encounter notifications
	ShowSystemMessages   bool `json:"show_system_messages"`  // System announcements
	ChatNotifications    bool `json:"chat_notifications"`    // Chat message alerts
	NotificationDuration int  `json:"notification_duration"` // Seconds to display (0 = until dismissed)
}

// NewSettings creates default settings for a player
func NewSettings(playerID uuid.UUID) *Settings {
	return &Settings{
		ID:        uuid.New(),
		PlayerID:  playerID,
		UpdatedAt: time.Now(),
		Display: DisplaySettings{
			ColorScheme:      "default",
			ShowAnimations:   true,
			CompactMode:      false,
			ShowTutorialTips: true,
			ShowIcons:        true,
			TerminalWidth:    0,
			TerminalHeight:   0,
		},
		Audio: AudioSettings{
			Enabled:       false, // Disabled by default (not yet implemented)
			SoundEffects:  true,
			Music:         true,
			Notifications: true,
			Volume:        70,
		},
		Gameplay: GameplaySettings{
			AutoSave:                true,
			AutoSaveInterval:        5,
			ConfirmDangerousActions: true,
			ShowDamageNumbers:       true,
			AutoPilot:               false,
			PauseOnEncounter:        true,
			FastTravel:              false,
			TutorialMode:            true,
			DifficultyLevel:         "normal",
			PermadeathMode:          false,
		},
		Controls: ControlSettings{
			// Navigation
			MoveUp:    "up,k",
			MoveDown:  "down,j",
			MoveLeft:  "left,h",
			MoveRight: "right,l",
			// Actions
			Confirm:  "enter,space",
			Cancel:   "esc",
			Back:     "backspace",
			Help:     "?,F1",
			Settings: "s",
			// Shortcuts
			QuickSave:       "F5",
			QuickLoad:       "F9",
			ToggleMap:       "m",
			ToggleInventory: "i",
			ToggleShip:      "v",
			ToggleChat:      "c",
			// Combat
			Attack:     "a",
			Defend:     "d",
			Flee:       "f",
			UseItem:    "u",
			NextTarget: "tab",
			PrevTarget: "shift+tab",
			// Modes
			VimMode:   false,
			EmacsMode: false,
		},
		Privacy: PrivacySettings{
			ShowOnline:         true,
			ShowLocation:       true,
			ShowShip:           true,
			AllowTradeRequests: true,
			AllowPvPChallenges: true,
			AllowPartyInvites:  true,
			BlockList:          []uuid.UUID{},
			FriendsList:        []uuid.UUID{},
		},
		Notifications: NotificationSettings{
			ShowAchievements:     true,
			ShowLevelUp:          true,
			ShowTradeComplete:    true,
			ShowCombatLog:        true,
			ShowPlayerJoined:     false,
			ShowNewsUpdates:      true,
			ShowEncounters:       true,
			ShowSystemMessages:   true,
			ChatNotifications:    true,
			NotificationDuration: 5,
		},
	}
}

// Clone creates a copy of settings
func (s *Settings) Clone() *Settings {
	clone := *s
	clone.ID = uuid.New()
	clone.UpdatedAt = time.Now()

	// Deep copy slices
	clone.Privacy.BlockList = make([]uuid.UUID, len(s.Privacy.BlockList))
	copy(clone.Privacy.BlockList, s.Privacy.BlockList)

	clone.Privacy.FriendsList = make([]uuid.UUID, len(s.Privacy.FriendsList))
	copy(clone.Privacy.FriendsList, s.Privacy.FriendsList)

	return &clone
}

// IsPlayerBlocked checks if a player is blocked
func (s *Settings) IsPlayerBlocked(playerID uuid.UUID) bool {
	for _, id := range s.Privacy.BlockList {
		if id == playerID {
			return true
		}
	}
	return false
}

// IsPlayerFriend checks if a player is a friend
func (s *Settings) IsPlayerFriend(playerID uuid.UUID) bool {
	for _, id := range s.Privacy.FriendsList {
		if id == playerID {
			return true
		}
	}
	return false
}

// BlockPlayer adds a player to block list
func (s *Settings) BlockPlayer(playerID uuid.UUID) {
	if !s.IsPlayerBlocked(playerID) {
		s.Privacy.BlockList = append(s.Privacy.BlockList, playerID)
		s.UpdatedAt = time.Now()
	}
}

// UnblockPlayer removes a player from block list
func (s *Settings) UnblockPlayer(playerID uuid.UUID) {
	for i, id := range s.Privacy.BlockList {
		if id == playerID {
			s.Privacy.BlockList = append(s.Privacy.BlockList[:i], s.Privacy.BlockList[i+1:]...)
			s.UpdatedAt = time.Now()
			return
		}
	}
}

// AddFriend adds a player to friends list
func (s *Settings) AddFriend(playerID uuid.UUID) {
	if !s.IsPlayerFriend(playerID) {
		s.Privacy.FriendsList = append(s.Privacy.FriendsList, playerID)
		s.UpdatedAt = time.Now()
	}
}

// RemoveFriend removes a player from friends list
func (s *Settings) RemoveFriend(playerID uuid.UUID) {
	for i, id := range s.Privacy.FriendsList {
		if id == playerID {
			s.Privacy.FriendsList = append(s.Privacy.FriendsList[:i], s.Privacy.FriendsList[i+1:]...)
			s.UpdatedAt = time.Now()
			return
		}
	}
}

// ApplyColorScheme returns style settings based on color scheme
func (s *Settings) ApplyColorScheme() ColorScheme {
	schemes := map[string]ColorScheme{
		"default":       DefaultColorScheme(),
		"dark":          DarkColorScheme(),
		"light":         LightColorScheme(),
		"high_contrast": HighContrastColorScheme(),
		"colorblind":    ColorblindColorScheme(),
	}

	if scheme, exists := schemes[s.Display.ColorScheme]; exists {
		return scheme
	}

	return DefaultColorScheme()
}

// ColorScheme defines colors for UI elements
type ColorScheme struct {
	Name       string
	Primary    string // Main text color
	Secondary  string // Secondary text
	Accent     string // Highlights, selections
	Success    string // Success messages
	Warning    string // Warnings
	Error      string // Errors
	Background string // Background
	Border     string // Borders and separators
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Name:       "default",
		Primary:    "#FFFFFF",
		Secondary:  "#AAAAAA",
		Accent:     "#00FFFF",
		Success:    "#00FF00",
		Warning:    "#FFFF00",
		Error:      "#FF0000",
		Background: "#000000",
		Border:     "#444444",
	}
}

// DarkColorScheme returns a dark color scheme
func DarkColorScheme() ColorScheme {
	return ColorScheme{
		Name:       "dark",
		Primary:    "#E0E0E0",
		Secondary:  "#888888",
		Accent:     "#0088FF",
		Success:    "#00CC00",
		Warning:    "#FFAA00",
		Error:      "#CC0000",
		Background: "#1A1A1A",
		Border:     "#333333",
	}
}

// LightColorScheme returns a light color scheme
func LightColorScheme() ColorScheme {
	return ColorScheme{
		Name:       "light",
		Primary:    "#000000",
		Secondary:  "#666666",
		Accent:     "#0066CC",
		Success:    "#008800",
		Warning:    "#CC8800",
		Error:      "#CC0000",
		Background: "#FFFFFF",
		Border:     "#CCCCCC",
	}
}

// HighContrastColorScheme returns a high contrast scheme
func HighContrastColorScheme() ColorScheme {
	return ColorScheme{
		Name:       "high_contrast",
		Primary:    "#FFFFFF",
		Secondary:  "#CCCCCC",
		Accent:     "#FFFF00",
		Success:    "#00FF00",
		Warning:    "#FFFF00",
		Error:      "#FF0000",
		Background: "#000000",
		Border:     "#FFFFFF",
	}
}

// ColorblindColorScheme returns a colorblind-friendly scheme
func ColorblindColorScheme() ColorScheme {
	return ColorScheme{
		Name:       "colorblind",
		Primary:    "#FFFFFF",
		Secondary:  "#AAAAAA",
		Accent:     "#0099FF", // Blue instead of cyan
		Success:    "#00BB00", // Brighter green
		Warning:    "#FF9900", // Orange instead of yellow
		Error:      "#FF3333", // Brighter red
		Background: "#000000",
		Border:     "#666666",
	}
}
