// File: internal/chat/commands.go
// Project: Terminal Velocity
// Description: Chat command parser and handlers
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package chat

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// CommandHandler handles chat commands and returns response messages
type CommandHandler struct {
	chatManager    *Manager
	presenceGetter func() []models.PlayerPresence // Get online players
	blockChecker   func(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
	blockAdder     func(ctx context.Context, blockerID uuid.UUID, blockedUsername, reason string) error
	blockRemover   func(ctx context.Context, blockerID, blockedID uuid.UUID) error
	playerGetter   func(username string) (*models.Player, error)
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(chatManager *Manager) *CommandHandler {
	return &CommandHandler{
		chatManager: chatManager,
	}
}

// SetPresenceGetter sets the function to get online players
func (h *CommandHandler) SetPresenceGetter(getter func() []models.PlayerPresence) {
	h.presenceGetter = getter
}

// SetBlockChecker sets the function to check if a player is blocked
func (h *CommandHandler) SetBlockChecker(checker func(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)) {
	h.blockChecker = checker
}

// SetBlockAdder sets the function to block a player
func (h *CommandHandler) SetBlockAdder(adder func(ctx context.Context, blockerID uuid.UUID, blockedUsername, reason string) error) {
	h.blockAdder = adder
}

// SetBlockRemover sets the function to unblock a player
func (h *CommandHandler) SetBlockRemover(remover func(ctx context.Context, blockerID, blockedID uuid.UUID) error) {
	h.blockRemover = remover
}

// SetPlayerGetter sets the function to get a player by username
func (h *CommandHandler) SetPlayerGetter(getter func(username string) (*models.Player, error)) {
	h.playerGetter = getter
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	Success      bool
	Message      string
	SystemOutput string // Message to show to command sender only
}

// ParseAndExecute parses a chat message and executes any commands
func (h *CommandHandler) ParseAndExecute(ctx context.Context, playerID uuid.UUID, playerName, message string) *CommandResult {
	// Check if message starts with /
	if !strings.HasPrefix(message, "/") {
		return nil // Not a command
	}

	// Remove leading / and split into parts
	message = strings.TrimPrefix(message, "/")
	parts := strings.Fields(message)

	if len(parts) == 0 {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Invalid command. Type /help for a list of commands.",
		}
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// Route to appropriate handler
	switch command {
	case "w", "whisper", "msg", "tell":
		return h.handleWhisper(ctx, playerID, playerName, args)
	case "who", "online":
		return h.handleWho(ctx, playerID)
	case "roll", "dice":
		return h.handleRoll(ctx, playerID, playerName, args)
	case "me", "emote":
		return h.handleEmote(ctx, playerID, playerName, args)
	case "ignore", "block":
		return h.handleIgnore(ctx, playerID, args)
	case "unignore", "unblock":
		return h.handleUnignore(ctx, playerID, args)
	case "help", "commands":
		return h.handleHelp(ctx)
	default:
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Unknown command: /%s. Type /help for a list of commands.", command),
		}
	}
}

// handleWhisper handles /whisper, /w, /msg, /tell commands
func (h *CommandHandler) handleWhisper(ctx context.Context, senderID uuid.UUID, senderName string, args []string) *CommandResult {
	if len(args) < 2 {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Usage: /whisper <player> <message>",
		}
	}

	recipientName := args[0]
	message := strings.Join(args[1:], " ")

	// Get recipient player
	if h.playerGetter == nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Whisper functionality not available",
		}
	}

	recipient, err := h.playerGetter(recipientName)
	if err != nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Player '%s' not found", recipientName),
		}
	}

	// Check if blocked
	if h.blockChecker != nil {
		blocked, err := h.blockChecker(ctx, recipient.ID, senderID)
		if err == nil && blocked {
			// Don't reveal that sender is blocked
			return &CommandResult{
				Success:      true,
				SystemOutput: fmt.Sprintf("Whispered to %s: %s", recipientName, message),
			}
		}
	}

	// Send direct message
	h.chatManager.SendDirectMessage(senderID, senderName, recipient.ID, recipient.Username, message)

	return &CommandResult{
		Success:      true,
		SystemOutput: fmt.Sprintf("Whispered to %s: %s", recipientName, message),
	}
}

// handleWho handles /who and /online commands
func (h *CommandHandler) handleWho(ctx context.Context, playerID uuid.UUID) *CommandResult {
	if h.presenceGetter == nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Player list not available",
		}
	}

	onlinePlayers := h.presenceGetter()

	if len(onlinePlayers) == 0 {
		return &CommandResult{
			Success:      true,
			SystemOutput: "No players currently online",
		}
	}

	// Build player list
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Online Players (%d):\n", len(onlinePlayers)))
	output.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, player := range onlinePlayers {
		location := "Unknown"
		if player.CurrentSystem != uuid.Nil {
			location = player.CurrentSystem.String()
			if player.CurrentPlanet != nil && *player.CurrentPlanet != uuid.Nil {
				location = fmt.Sprintf("%s (%s)", player.CurrentPlanet.String(), player.CurrentSystem.String())
			}
		}

		ship := player.ShipName
		if ship == "" {
			ship = player.ShipType
			if ship == "" {
				ship = "Unknown"
			}
		}

		output.WriteString(fmt.Sprintf("• %s - %s @ %s\n", player.Username, ship, location))
	}

	return &CommandResult{
		Success:      true,
		SystemOutput: output.String(),
	}
}

// handleRoll handles /roll and /dice commands
func (h *CommandHandler) handleRoll(ctx context.Context, playerID uuid.UUID, playerName string, args []string) *CommandResult {
	// Default to 1d6 if no args
	diceNotation := "1d6"
	if len(args) > 0 {
		diceNotation = args[0]
	}

	// Parse dice notation (e.g., "2d6", "1d20", "3d10+5")
	result, rolls, modifier, err := parseDiceRoll(diceNotation)
	if err != nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Invalid dice notation. Examples: 1d6, 2d10, 1d20+5"),
		}
	}

	// Build result message
	var message string
	if len(rolls) == 1 {
		if modifier != 0 {
			message = fmt.Sprintf("rolled %s and got %d (rolled %d%+d)", diceNotation, result, rolls[0], modifier)
		} else {
			message = fmt.Sprintf("rolled %s and got %d", diceNotation, result)
		}
	} else {
		rollsStr := make([]string, len(rolls))
		for i, r := range rolls {
			rollsStr[i] = strconv.Itoa(r)
		}
		if modifier != 0 {
			message = fmt.Sprintf("rolled %s and got %d ([%s]%+d)", diceNotation, result, strings.Join(rollsStr, "+"), modifier)
		} else {
			message = fmt.Sprintf("rolled %s and got %d ([%s])", diceNotation, result, strings.Join(rollsStr, "+"))
		}
	}

	// Broadcast to global chat as system message
	fullMessage := fmt.Sprintf("* %s %s", playerName, message)
	h.chatManager.BroadcastSystemMessage(models.ChatChannelGlobal, fullMessage)

	return &CommandResult{
		Success: true,
		Message: fullMessage,
	}
}

// handleEmote handles /me and /emote commands
func (h *CommandHandler) handleEmote(ctx context.Context, playerID uuid.UUID, playerName string, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Usage: /me <action>",
		}
	}

	action := strings.Join(args, " ")
	emoteMessage := fmt.Sprintf("* %s %s", playerName, action)

	// Broadcast to global chat
	h.chatManager.BroadcastSystemMessage(models.ChatChannelGlobal, emoteMessage)

	return &CommandResult{
		Success: true,
		Message: emoteMessage,
	}
}

// handleIgnore handles /ignore and /block commands
func (h *CommandHandler) handleIgnore(ctx context.Context, playerID uuid.UUID, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Usage: /ignore <player> [reason]",
		}
	}

	targetUsername := args[0]
	reason := "Ignored via chat command"
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	if h.blockAdder == nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Block functionality not available",
		}
	}

	err := h.blockAdder(ctx, playerID, targetUsername, reason)
	if err != nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Failed to ignore player: %v", err),
		}
	}

	return &CommandResult{
		Success:      true,
		SystemOutput: fmt.Sprintf("You are now ignoring %s", targetUsername),
	}
}

// handleUnignore handles /unignore and /unblock commands
func (h *CommandHandler) handleUnignore(ctx context.Context, playerID uuid.UUID, args []string) *CommandResult {
	if len(args) == 0 {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Usage: /unignore <player>",
		}
	}

	targetUsername := args[0]

	if h.blockRemover == nil || h.playerGetter == nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: "Block functionality not available",
		}
	}

	// Get target player
	target, err := h.playerGetter(targetUsername)
	if err != nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Player '%s' not found", targetUsername),
		}
	}

	err = h.blockRemover(ctx, playerID, target.ID)
	if err != nil {
		return &CommandResult{
			Success:      false,
			SystemOutput: fmt.Sprintf("Failed to unignore player: %v", err),
		}
	}

	return &CommandResult{
		Success:      true,
		SystemOutput: fmt.Sprintf("You are no longer ignoring %s", targetUsername),
	}
}

// handleHelp handles /help and /commands
func (h *CommandHandler) handleHelp(ctx context.Context) *CommandResult {
	helpText := `Available Chat Commands:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Communication:
  /whisper <player> <message>  Send a private message
  /w, /msg, /tell              Aliases for whisper

Player Information:
  /who                         List all online players
  /online                      Alias for who

Fun & Roleplay:
  /roll <dice>                 Roll dice (e.g., 1d6, 2d10+5)
  /dice                        Alias for roll
  /me <action>                 Perform an emote action
  /emote                       Alias for me

Moderation:
  /ignore <player> [reason]    Block a player
  /block                       Alias for ignore
  /unignore <player>           Unblock a player
  /unblock                     Alias for unignore

Other:
  /help                        Show this help message
  /commands                    Alias for help

Examples:
  /w Alice Hey, want to trade?
  /roll 2d6
  /me waves at the space station
  /ignore TrollPlayer spamming
`

	return &CommandResult{
		Success:      true,
		SystemOutput: helpText,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseDiceRoll parses dice notation and returns result, individual rolls, modifier, error
func parseDiceRoll(notation string) (int, []int, int, error) {
	notation = strings.ToLower(notation)
	notation = strings.TrimSpace(notation)

	// Parse modifier (e.g., "2d6+3" or "1d20-2")
	modifier := 0
	if strings.Contains(notation, "+") {
		parts := strings.Split(notation, "+")
		if len(parts) != 2 {
			return 0, nil, 0, fmt.Errorf("invalid modifier")
		}
		notation = parts[0]
		mod, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, nil, 0, fmt.Errorf("invalid modifier: %v", err)
		}
		modifier = mod
	} else if strings.Contains(notation, "-") {
		parts := strings.Split(notation, "-")
		if len(parts) != 2 {
			return 0, nil, 0, fmt.Errorf("invalid modifier")
		}
		notation = parts[0]
		mod, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, nil, 0, fmt.Errorf("invalid modifier: %v", err)
		}
		modifier = -mod
	}

	// Parse dice (e.g., "2d6")
	parts := strings.Split(notation, "d")
	if len(parts) != 2 {
		return 0, nil, 0, fmt.Errorf("invalid dice notation")
	}

	numDice, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, nil, 0, fmt.Errorf("invalid number of dice: %v", err)
	}

	numSides, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, nil, 0, fmt.Errorf("invalid number of sides: %v", err)
	}

	// Validate
	if numDice < 1 || numDice > 100 {
		return 0, nil, 0, fmt.Errorf("number of dice must be between 1 and 100")
	}
	if numSides < 2 || numSides > 1000 {
		return 0, nil, 0, fmt.Errorf("number of sides must be between 2 and 1000")
	}

	// Roll dice
	rand.Seed(time.Now().UnixNano())
	rolls := make([]int, numDice)
	total := 0

	for i := 0; i < numDice; i++ {
		roll := rand.Intn(numSides) + 1
		rolls[i] = roll
		total += roll
	}

	total += modifier

	return total, rolls, modifier, nil
}
