// File: internal/tui/friends.go
// Project: Terminal Velocity
// Description: Friends system TUI screen
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// Friends screen modes
const (
	friendsModeFriends  = "friends"
	friendsModeRequests = "requests"
	friendsModeBlocked  = "blocked"
	friendsModeAdd      = "add"
)

// Friends screen state
type friendsState struct {
	mode          string
	friends       []models.Friend
	requests      []models.FriendRequest
	blocked       []models.Block
	selectedIndex int
	loading       bool
	error         string

	// Add friend state
	usernameInput string
}

func newFriendsState() friendsState {
	return friendsState{
		mode:          friendsModeFriends,
		selectedIndex: 0,
	}
}

func (m *Model) updateFriends(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle add friend mode separately
		if m.friends.mode == friendsModeAdd {
			return m.updateFriendsAdd(msg)
		}

		// Common navigation
		switch msg.String() {
		case "q", "esc":
			m.screen = ScreenGame
			return m, nil

		case "1":
			// Switch to friends list
			m.friends.mode = friendsModeFriends
			m.friends.selectedIndex = 0
			m.friends.loading = true
			return m, m.loadFriendsList()

		case "2":
			// Switch to friend requests
			m.friends.mode = friendsModeRequests
			m.friends.selectedIndex = 0
			m.friends.loading = true
			return m, m.loadFriendRequests()

		case "3":
			// Switch to blocked list
			m.friends.mode = friendsModeBlocked
			m.friends.selectedIndex = 0
			m.friends.loading = true
			return m, m.loadBlockedList()

		case "a":
			// Add friend
			m.friends.mode = friendsModeAdd
			m.friends.usernameInput = ""
			m.friends.error = ""
			return m, nil

		case "up", "k":
			if m.friends.selectedIndex > 0 {
				m.friends.selectedIndex--
			}

		case "down", "j":
			maxIndex := 0
			switch m.friends.mode {
			case friendsModeFriends:
				maxIndex = len(m.friends.friends) - 1
			case friendsModeRequests:
				maxIndex = len(m.friends.requests) - 1
			case friendsModeBlocked:
				maxIndex = len(m.friends.blocked) - 1
			}
			if m.friends.selectedIndex < maxIndex {
				m.friends.selectedIndex++
			}

		case "r":
			// Refresh current view
			m.friends.loading = true
			switch m.friends.mode {
			case friendsModeFriends:
				return m, m.loadFriendsList()
			case friendsModeRequests:
				return m, m.loadFriendRequests()
			case friendsModeBlocked:
				return m, m.loadBlockedList()
			}

		case "enter", "y":
			// Accept friend request
			if m.friends.mode == friendsModeRequests && len(m.friends.requests) > 0 {
				request := m.friends.requests[m.friends.selectedIndex]
				// Only accept if we're the receiver
				if request.ReceiverID == m.playerID && request.Status == models.FriendRequestPending {
					return m, m.acceptFriendRequest(request.ID)
				}
			}

		case "n", "d":
			// Decline/remove based on mode
			switch m.friends.mode {
			case friendsModeFriends:
				// Remove friend
				if len(m.friends.friends) > 0 {
					friend := m.friends.friends[m.friends.selectedIndex]
					return m, m.removeFriend(friend.FriendID)
				}

			case friendsModeRequests:
				// Decline friend request
				if len(m.friends.requests) > 0 {
					request := m.friends.requests[m.friends.selectedIndex]
					if request.ReceiverID == m.playerID && request.Status == models.FriendRequestPending {
						return m, m.declineFriendRequest(request.ID)
					}
				}

			case friendsModeBlocked:
				// Unblock player
				if len(m.friends.blocked) > 0 {
					block := m.friends.blocked[m.friends.selectedIndex]
					return m, m.unblockPlayer(block.BlockedID)
				}
			}
		}

	case friendsLoadedMsg:
		m.friends.loading = false
		m.friends.friends = msg.friends
		m.friends.error = msg.err
		if m.friends.selectedIndex >= len(msg.friends) {
			m.friends.selectedIndex = 0
		}

	case friendRequestsLoadedMsg:
		m.friends.loading = false
		m.friends.requests = msg.requests
		m.friends.error = msg.err
		if m.friends.selectedIndex >= len(msg.requests) {
			m.friends.selectedIndex = 0
		}

	case blockedLoadedMsg:
		m.friends.loading = false
		m.friends.blocked = msg.blocked
		m.friends.error = msg.err
		if m.friends.selectedIndex >= len(msg.blocked) {
			m.friends.selectedIndex = 0
		}

	case friendActionMsg:
		m.friends.loading = false
		if msg.err == "" {
			// Success - reload current view
			m.friends.error = ""
			switch m.friends.mode {
			case friendsModeFriends:
				return m, m.loadFriendsList()
			case friendsModeRequests:
				return m, m.loadFriendRequests()
			case friendsModeBlocked:
				return m, m.loadBlockedList()
			case friendsModeAdd:
				// Return to friends list after successful add
				m.friends.mode = friendsModeFriends
				return m, m.loadFriendsList()
			}
		} else {
			m.friends.error = msg.err
		}
	}

	return m, nil
}

func (m *Model) updateFriendsAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel add
		m.friends.mode = friendsModeFriends
		m.friends.error = ""
		return m, nil

	case "enter":
		// Send friend request
		if m.friends.usernameInput == "" {
			m.friends.error = "Username is required"
			return m, nil
		}

		m.friends.loading = true
		m.friends.error = ""
		return m, m.sendFriendRequest(m.friends.usernameInput)

	case "backspace":
		if len(m.friends.usernameInput) > 0 {
			m.friends.usernameInput = m.friends.usernameInput[:len(m.friends.usernameInput)-1]
		}

	default:
		// Add character to input
		if len(msg.String()) == 1 {
			if len(m.friends.usernameInput) < 32 {
				m.friends.usernameInput += msg.String()
			}
		}
	}

	return m, nil
}

func (m *Model) viewFriends() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("═══ FRIENDS & SOCIAL ═══")

	b.WriteString(title + "\n\n")

	if m.friends.loading {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("Loading...") + "\n\n")
		return b.String()
	}

	if m.friends.error != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("Error: "+m.friends.error) + "\n\n")
	}

	// Mode selector
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render("Tabs:") + " ")

	if m.friends.mode == friendsModeFriends {
		b.WriteString(fmt.Sprintf("[1] Friends (%d)  ", len(m.friends.friends)))
	} else {
		b.WriteString("[1] Friends  ")
	}

	if m.friends.mode == friendsModeRequests {
		b.WriteString(fmt.Sprintf("[2] Requests (%d)  ", len(m.friends.requests)))
	} else {
		b.WriteString("[2] Requests  ")
	}

	if m.friends.mode == friendsModeBlocked {
		b.WriteString(fmt.Sprintf("[3] Blocked (%d)  ", len(m.friends.blocked)))
	} else {
		b.WriteString("[3] Blocked  ")
	}

	b.WriteString("[A] Add Friend\n\n")

	// Render based on mode
	switch m.friends.mode {
	case friendsModeFriends:
		b.WriteString(m.renderFriendsList())
	case friendsModeRequests:
		b.WriteString(m.renderFriendRequests())
	case friendsModeBlocked:
		b.WriteString(m.renderBlockedList())
	case friendsModeAdd:
		b.WriteString(m.renderAddFriend())
	}

	return b.String()
}

func (m *Model) renderFriendsList() string {
	var b strings.Builder

	if len(m.friends.friends) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("You haven't added any friends yet.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[A] Add Friend  [Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-20s %-15s %-30s %-10s\n", "Username", "Status", "Location", "Ship")))
	b.WriteString(strings.Repeat("─", 80) + "\n")

	// Friends list
	for i, friend := range m.friends.friends {
		style := lipgloss.NewStyle()
		if friend.IsOnline {
			style = style.Foreground(lipgloss.Color("10")) // Green for online
		} else {
			style = style.Foreground(lipgloss.Color("8")) // Gray for offline
		}

		if i == m.friends.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.friends.selectedIndex {
			cursor = "→ "
		}

		status := "Offline"
		if friend.IsOnline {
			status = "Online"
		}

		location := friend.Location
		if location == "" {
			location = "Unknown"
		}

		ship := friend.CurrentShip
		if ship == "" {
			ship = "Unknown"
		}

		line := fmt.Sprintf("%s%-20s %-15s %-30s %-10s",
			cursor,
			truncate(friend.FriendName, 18),
			status,
			truncate(location, 28),
			truncate(ship, 8),
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [D] Remove Friend  [R] Refresh  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderFriendRequests() string {
	var b strings.Builder

	if len(m.friends.requests) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No pending friend requests.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[A] Add Friend  [Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-20s %-15s %-20s %-10s\n", "Player", "Direction", "Date", "Status")))
	b.WriteString(strings.Repeat("─", 70) + "\n")

	// Requests list
	for i, request := range m.friends.requests {
		style := lipgloss.NewStyle()
		if request.Status == models.FriendRequestPending {
			if request.ReceiverID == m.playerID {
				style = style.Foreground(lipgloss.Color("11")) // Yellow for received pending
			} else {
				style = style.Foreground(lipgloss.Color("14")) // Cyan for sent pending
			}
		} else {
			style = style.Foreground(lipgloss.Color("8")) // Gray for non-pending
		}

		if i == m.friends.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.friends.selectedIndex {
			cursor = "→ "
		}

		// Determine player name and direction
		playerName := ""
		direction := ""
		if request.ReceiverID == m.playerID {
			playerName = request.SenderName
			direction = "From"
		} else {
			playerName = request.ReceiverName
			direction = "To"
		}

		timeStr := request.CreatedAt.Format("2006-01-02")

		line := fmt.Sprintf("%s%-20s %-15s %-20s %-10s",
			cursor,
			truncate(playerName, 18),
			direction,
			timeStr,
			request.Status,
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Y] Accept  [N] Decline  [R] Refresh  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderBlockedList() string {
	var b strings.Builder

	if len(m.friends.blocked) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No blocked players.\n\n"))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[Q] Back\n"))
		return b.String()
	}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-25s %-35s %-15s\n", "Username", "Reason", "Blocked On")))
	b.WriteString(strings.Repeat("─", 80) + "\n")

	// Blocked list
	for i, block := range m.friends.blocked {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red for blocked

		if i == m.friends.selectedIndex {
			style = style.Background(lipgloss.Color("235"))
		}

		cursor := "  "
		if i == m.friends.selectedIndex {
			cursor = "→ "
		}

		timeStr := block.CreatedAt.Format("2006-01-02")

		line := fmt.Sprintf("%s%-25s %-35s %-15s",
			cursor,
			truncate(block.BlockedName, 23),
			truncate(block.Reason, 33),
			timeStr,
		)

		b.WriteString(style.Render(line) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [D] Unblock  [R] Refresh  [Q] Back\n"))

	return b.String()
}

func (m *Model) renderAddFriend() string {
	var b strings.Builder

	addStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(60)

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Add Friend") + "\n\n")
	content.WriteString("Enter the username of the player you want to add as a friend:\n\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true).
		Render(fmt.Sprintf("Username: %s_\n", m.friends.usernameInput)))

	b.WriteString(addStyle.Render(content.String()))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[Enter] Send Request  [Esc] Cancel\n"))

	return b.String()
}

// Commands

func (m *Model) loadFriendsList() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to load friends list
		var friends []models.Friend
		var errStr string
		if m.friendsManager != nil {
			var err error
			friends, err = m.friendsManager.GetFriends(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		} else {
			friends = []models.Friend{}
		}

		return friendsLoadedMsg{
			friends: friends,
			err:     errStr,
		}
	}
}

func (m *Model) loadFriendRequests() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to load friend requests
		var requests []models.FriendRequest
		var errStr string
		if m.friendsManager != nil {
			var err error
			requests, err = m.friendsManager.GetPendingFriendRequests(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		} else {
			requests = []models.FriendRequest{}
		}

		return friendRequestsLoadedMsg{
			requests: requests,
			err:      errStr,
		}
	}
}

func (m *Model) loadBlockedList() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to load blocked players
		var blocked []models.Block
		var errStr string
		if m.friendsManager != nil {
			var err error
			blocked, err = m.friendsManager.GetBlockedPlayers(ctx, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		} else {
			blocked = []models.Block{}
		}

		return blockedLoadedMsg{
			blocked: blocked,
			err:     errStr,
		}
	}
}

func (m *Model) sendFriendRequest(username string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to send friend request
		var errStr string
		if m.friendsManager != nil && m.playerRepo != nil {
			// Helper function to get player by username
			getPlayerByUsername := func(user string) (*models.Player, error) {
				return m.playerRepo.GetByUsername(ctx, user)
			}

			err := m.friendsManager.SendFriendRequest(ctx, m.playerID, username, getPlayerByUsername)
			if err != nil {
				errStr = err.Error()
			}
		}

		return friendActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) acceptFriendRequest(requestID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to accept friend request
		var errStr string
		if m.friendsManager != nil {
			err := m.friendsManager.AcceptFriendRequest(ctx, requestID, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return friendActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) declineFriendRequest(requestID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to decline friend request
		var errStr string
		if m.friendsManager != nil {
			err := m.friendsManager.DeclineFriendRequest(ctx, requestID, m.playerID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return friendActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) removeFriend(friendID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to remove friend
		var errStr string
		if m.friendsManager != nil {
			err := m.friendsManager.RemoveFriend(ctx, m.playerID, friendID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return friendActionMsg{
			err: errStr,
		}
	}
}

func (m *Model) unblockPlayer(blockedID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use friends manager to unblock player
		var errStr string
		if m.friendsManager != nil {
			err := m.friendsManager.UnblockPlayer(ctx, m.playerID, blockedID)
			if err != nil {
				errStr = err.Error()
			}
		}

		return friendActionMsg{
			err: errStr,
		}
	}
}

// Messages

type friendsLoadedMsg struct {
	friends []models.Friend
	err     string
}

type friendRequestsLoadedMsg struct {
	requests []models.FriendRequest
	err      string
}

type blockedLoadedMsg struct {
	blocked []models.Block
	err     string
}

type friendActionMsg struct {
	err string
}
