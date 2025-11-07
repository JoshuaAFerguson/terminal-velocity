// File: internal/tui/pvp.go
// Project: Terminal Velocity
// Version: 1.0.0

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// PvP view modes
const (
	pvpViewChallenges = "challenges" // Active challenges
	pvpViewBounties   = "bounties"   // Bounty board
	pvpViewStats      = "stats"      // Player stats
	pvpViewCreate     = "create"     // Create new challenge
)

type pvpModel struct {
	viewMode         string
	cursor           int
	selectedChallenge *models.PvPChallenge

	// Create mode fields
	createTarget      string
	createType        string
	createWager       int64
	createMessage     string
	createInputField  int // 0=target, 1=type, 2=wager, 3=message
	challengeTypes    []models.PvPChallengeType
}

func newPvPModel() pvpModel {
	return pvpModel{
		viewMode:      pvpViewChallenges,
		cursor:        0,
		createInputField: 0,
		challengeTypes: []models.PvPChallengeType{
			models.ChallengeDuel,
			models.ChallengeAggression,
			models.ChallengeBountyHunt,
		},
		createType: string(models.ChallengeDuel),
	}
}

func (m Model) updatePvP(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.pvpModel.viewMode {
		case pvpViewCreate:
			return m.updatePvPCreate(msg)
		default:
			return m.updatePvPList(msg)
		}
	}

	return m, nil
}

func (m Model) updatePvPList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.pvpModel.cursor > 0 {
			m.pvpModel.cursor--
		}

	case "down", "j":
		var itemCount int
		switch m.pvpModel.viewMode {
		case pvpViewChallenges:
			challenges := m.pvpManager.GetPlayerChallenges(m.playerID)
			itemCount = len(challenges)
		case pvpViewBounties:
			bounties := m.pvpManager.GetAllActiveBounties()
			itemCount = len(bounties)
		case pvpViewStats:
			leaderboard := m.pvpManager.GetLeaderboard(10)
			itemCount = len(leaderboard)
		}

		if m.pvpModel.cursor < itemCount-1 {
			m.pvpModel.cursor++
		}

	case "1":
		m.pvpModel.viewMode = pvpViewChallenges
		m.pvpModel.cursor = 0

	case "2":
		m.pvpModel.viewMode = pvpViewBounties
		m.pvpModel.cursor = 0

	case "3":
		m.pvpModel.viewMode = pvpViewStats
		m.pvpModel.cursor = 0

	case "n":
		// Create new challenge
		m.pvpModel.viewMode = pvpViewCreate
		m.pvpModel.createTarget = ""
		m.pvpModel.createWager = 0
		m.pvpModel.createMessage = ""
		m.pvpModel.createInputField = 0
		m.pvpModel.createType = string(models.ChallengeDuel)

	case "a":
		// Accept challenge
		if m.pvpModel.viewMode == pvpViewChallenges {
			challenges := m.pvpManager.GetPendingChallenges(m.playerID)
			if m.pvpModel.cursor < len(challenges) {
				challenge := challenges[m.pvpModel.cursor]
				_ = m.pvpManager.AcceptChallenge(challenge.ID, m.playerID)

				// TODO: In full implementation, this would start actual combat
				// For now, simulate instant combat resolution
				m.simulateCombat(challenge.ID)
			}
		} else if m.pvpModel.viewMode == pvpViewBounties {
			// Hunt bounty
			bounties := m.pvpManager.GetAllActiveBounties()
			if m.pvpModel.cursor < len(bounties) {
				bounty := bounties[m.pvpModel.cursor]
				// Create bounty hunt challenge
				_, _ = m.pvpManager.CreateChallenge(
					m.playerID,
					m.username,
					bounty.TargetID,
					bounty.TargetName,
					models.ChallengeBountyHunt,
					m.player.CurrentSystem,
					0,
					"Claiming bounty!",
				)
			}
		}

	case "r":
		// Reject challenge
		if m.pvpModel.viewMode == pvpViewChallenges {
			challenges := m.pvpManager.GetPendingChallenges(m.playerID)
			if m.pvpModel.cursor < len(challenges) {
				challenge := challenges[m.pvpModel.cursor]
				_ = m.pvpManager.DeclineChallenge(challenge.ID, m.playerID)
			}
		}

	case "q", "esc":
		m.screen = ScreenMainMenu
	}

	return m, nil
}

func (m Model) updatePvPCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.pvpModel.createInputField = (m.pvpModel.createInputField + 1) % 4

	case "esc":
		m.pvpModel.viewMode = pvpViewChallenges
		m.pvpModel.cursor = 0

	case "enter":
		// Create the challenge
		if m.pvpModel.createTarget != "" {
			// TODO: In full implementation, validate target exists
			// For now, just create the challenge
			// Find target player ID (would come from presence system)
			targetID := uuid.New() // Placeholder

			_, _ = m.pvpManager.CreateChallenge(
				m.playerID,
				m.username,
				targetID,
				m.pvpModel.createTarget,
				models.PvPChallengeType(m.pvpModel.createType),
				m.player.CurrentSystem,
				m.pvpModel.createWager,
				m.pvpModel.createMessage,
			)
		}
		m.pvpModel.viewMode = pvpViewChallenges
		m.pvpModel.cursor = 0

	case "backspace":
		switch m.pvpModel.createInputField {
		case 0: // Target
			if len(m.pvpModel.createTarget) > 0 {
				m.pvpModel.createTarget = m.pvpModel.createTarget[:len(m.pvpModel.createTarget)-1]
			}
		case 3: // Message
			if len(m.pvpModel.createMessage) > 0 {
				m.pvpModel.createMessage = m.pvpModel.createMessage[:len(m.pvpModel.createMessage)-1]
			}
		}

	case "up":
		switch m.pvpModel.createInputField {
		case 1: // Challenge type
			for i, t := range m.pvpModel.challengeTypes {
				if string(t) == m.pvpModel.createType {
					if i > 0 {
						m.pvpModel.createType = string(m.pvpModel.challengeTypes[i-1])
					}
					break
				}
			}
		case 2: // Wager
			m.pvpModel.createWager += 1000
		}

	case "down":
		switch m.pvpModel.createInputField {
		case 1: // Challenge type
			for i, t := range m.pvpModel.challengeTypes {
				if string(t) == m.pvpModel.createType {
					if i < len(m.pvpModel.challengeTypes)-1 {
						m.pvpModel.createType = string(m.pvpModel.challengeTypes[i+1])
					}
					break
				}
			}
		case 2: // Wager
			if m.pvpModel.createWager >= 1000 {
				m.pvpModel.createWager -= 1000
			}
		}

	default:
		// Handle text input
		if len(msg.String()) == 1 {
			switch m.pvpModel.createInputField {
			case 0: // Target
				m.pvpModel.createTarget += msg.String()
			case 3: // Message
				if len(m.pvpModel.createMessage) < 200 {
					m.pvpModel.createMessage += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m Model) viewPvP() string {
	switch m.pvpModel.viewMode {
	case pvpViewCreate:
		return m.viewPvPCreate()
	case pvpViewBounties:
		return m.viewPvPBounties()
	case pvpViewStats:
		return m.viewPvPStats()
	default:
		return m.viewPvPChallenges()
	}
}

func (m Model) viewPvPChallenges() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))

	activeTabStyle := tabStyle.Copy().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("236"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("⚔️ PvP Combat"))
	s.WriteString("\n\n")

	// Tabs
	tabs := []string{"Challenges (1)", "Bounties (2)", "Stats (3)"}
	tabViews := []string{pvpViewChallenges, pvpViewBounties, pvpViewStats}

	for i, tab := range tabs {
		if m.pvpModel.viewMode == tabViews[i] {
			s.WriteString(activeTabStyle.Render(tab))
		} else {
			s.WriteString(tabStyle.Render(tab))
		}
	}
	s.WriteString("\n\n")

	// Get challenges
	challenges := m.pvpManager.GetPlayerChallenges(m.playerID)
	pending := m.pvpManager.GetPendingChallenges(m.playerID)

	s.WriteString(fmt.Sprintf("Your challenges (%d pending):\n\n", len(pending)))

	if len(challenges) == 0 {
		s.WriteString("  No challenges\n")
	} else {
		for i, challenge := range challenges {
			cursor := "  "
			if i == m.pvpModel.cursor {
				cursor = "→ "
			}

			typeIcon := challenge.Type.GetIcon()
			statusIcon := challenge.Status.GetIcon()

			otherPlayer := challenge.DefenderName
			role := "→"
			if challenge.DefenderID == m.playerID {
				otherPlayer = challenge.ChallengerName
				role = "←"
			}

			timeInfo := ""
			if challenge.Status == models.ChallengePending {
				timeInfo = challenge.GetTimeRemaining()
			}

			line := fmt.Sprintf("%s%s %s %s %s | %s | Wager: %d cr | %s",
				cursor,
				statusIcon,
				typeIcon,
				role,
				otherPlayer,
				challenge.Status,
				challenge.Wager,
				timeInfo,
			)

			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString("Controls: [↑/↓] Navigate [A] Accept [R] Reject [N] New Challenge [Q] Back\n")

	return boxStyle.Render(s.String())
}

func (m Model) viewPvPBounties() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("🎯 Bounty Board"))
	s.WriteString("\n\n")

	// Check if player has a bounty
	if bounty, hasBounty := m.pvpManager.GetBounty(m.playerID); hasBounty {
		s.WriteString(labelStyle.Render("⚠️  YOU ARE WANTED!") + "\n")
		s.WriteString(fmt.Sprintf("Bounty: %d cr | %s\n", bounty.Amount, bounty.GetWantedLevel()))
		s.WriteString(fmt.Sprintf("Reason: %s\n\n", bounty.Reason))
	}

	// List all active bounties
	bounties := m.pvpManager.GetAllActiveBounties()

	s.WriteString(fmt.Sprintf("Active Bounties: %d\n\n", len(bounties)))

	if len(bounties) == 0 {
		s.WriteString("  No active bounties\n")
	} else {
		for i, bounty := range bounties {
			cursor := "  "
			if i == m.pvpModel.cursor {
				cursor = "→ "
			}

			wanted := bounty.GetWantedLevel()

			line := fmt.Sprintf("%s🎯 %s | %d cr | %s",
				cursor,
				bounty.TargetName,
				bounty.Amount,
				wanted,
			)

			s.WriteString(line + "\n")
			s.WriteString(fmt.Sprintf("     Reason: %s | Issued by: %s\n",
				bounty.Reason, bounty.IssuedBy))
		}
	}

	s.WriteString("\n")
	s.WriteString("Controls: [↑/↓] Navigate [A] Hunt [Q] Back\n")

	return boxStyle.Render(s.String())
}

func (m Model) viewPvPStats() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("⚔️ Combat Statistics"))
	s.WriteString("\n\n")

	// Player's stats
	stats := m.pvpManager.GetStats(m.playerID)

	s.WriteString(labelStyle.Render("━━━ Your Stats ━━━") + "\n\n")
	s.WriteString(fmt.Sprintf("%s %s (%d)\n", labelStyle.Render("Combat Rating:"), stats.GetRatingClass(), stats.CombatRating))
	s.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Honor:"), stats.GetHonorRank()))
	s.WriteString(fmt.Sprintf("%s %d-%d-%d\n", labelStyle.Render("Record:"), stats.Wins, stats.Losses, stats.Draws))
	s.WriteString(fmt.Sprintf("%s %.1f%%\n", labelStyle.Render("Win Rate:"), stats.GetWinRate()))
	s.WriteString(fmt.Sprintf("%s %.2f\n\n", labelStyle.Render("K/D Ratio:"), stats.GetKDRatio()))

	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Duels Won:"), stats.DuelsWon))
	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Bounties Hunted:"), stats.BountiesHunted))
	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Win Streak:"), stats.CurrentWinStreak))
	s.WriteString(fmt.Sprintf("%s %d\n\n", labelStyle.Render("Best Streak:"), stats.LongestWinStreak))

	s.WriteString(fmt.Sprintf("%s %d cr\n", labelStyle.Render("Credits Won:"), stats.TotalCreditsWon))
	s.WriteString(fmt.Sprintf("%s %d cr\n\n", labelStyle.Render("Credits Lost:"), stats.TotalCreditsLost))

	// Leaderboard
	s.WriteString(labelStyle.Render("━━━ Top Pilots ━━━") + "\n\n")

	leaderboard := m.pvpManager.GetLeaderboard(10)
	for i, pilotStats := range leaderboard {
		cursor := "  "
		if i == m.pvpModel.cursor {
			cursor = "→ "
		}

		rank := fmt.Sprintf("%d.", i+1)

		// Highlight current player
		name := ""
		if pilotStats.PlayerID == m.playerID {
			name = "【" + m.username + "】"
		} else {
			name = "Player" // TODO: Get actual name from presence
		}

		line := fmt.Sprintf("%s%s %s | %s | %d rating | %d-%d",
			cursor,
			rank,
			name,
			pilotStats.GetRatingClass(),
			pilotStats.CombatRating,
			pilotStats.Wins,
			pilotStats.Losses,
		)

		s.WriteString(line + "\n")
	}

	s.WriteString("\n")
	s.WriteString("Controls: [1] Challenges [2] Bounties [Q] Back\n")

	return boxStyle.Render(s.String())
}

func (m Model) viewPvPCreate() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("236"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("⚔️ Create Combat Challenge"))
	s.WriteString("\n\n")

	// Target field
	targetLabel := labelStyle.Render("Target Player:")
	targetValue := m.pvpModel.createTarget
	if m.pvpModel.createInputField == 0 {
		targetValue = activeStyle.Render(targetValue + "_")
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", targetLabel, targetValue))

	// Challenge type
	typeLabel := labelStyle.Render("Challenge Type:")
	typeValue := fmt.Sprintf("%s %s", models.PvPChallengeType(m.pvpModel.createType).GetIcon(), m.pvpModel.createType)
	if m.pvpModel.createInputField == 1 {
		typeValue = activeStyle.Render(typeValue)
	}
	s.WriteString(fmt.Sprintf("%s %s (↑/↓ to change)\n\n", typeLabel, typeValue))

	// Wager
	wagerLabel := labelStyle.Render("Wager:")
	wagerValue := fmt.Sprintf("%d cr", m.pvpModel.createWager)
	if m.pvpModel.createInputField == 2 {
		wagerValue = activeStyle.Render(wagerValue)
	}
	s.WriteString(fmt.Sprintf("%s %s (↑/↓ to adjust)\n\n", wagerLabel, wagerValue))

	// Message
	messageLabel := labelStyle.Render("Message:")
	messageValue := m.pvpModel.createMessage
	if m.pvpModel.createInputField == 3 {
		messageValue = activeStyle.Render(messageValue + "_")
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", messageLabel, messageValue))

	// Type descriptions
	s.WriteString("Challenge Types:\n")
	s.WriteString("  ⚔️  Duel: Honorable combat, no penalties\n")
	s.WriteString("  💢 Aggression: Unprovoked attack, bounty risk\n")
	s.WriteString("  🎯 Bounty Hunt: Hunt wanted players for reward\n\n")

	s.WriteString("Controls: [Tab] Next Field [Enter] Send Challenge [Esc] Cancel\n")

	return boxStyle.Render(s.String())
}

// simulateCombat simulates combat for demonstration purposes
func (m *Model) simulateCombat(challengeID uuid.UUID) {
	// Simple 50/50 simulation
	winnerID := m.playerID

	// Simulate combat result
	_, _ = m.pvpManager.CompleteCombat(
		challengeID,
		winnerID,
		1000, // Credits transfer
		850,  // Winner damage dealt
		450,  // Loser damage dealt
	)
}
