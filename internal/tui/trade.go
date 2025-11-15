// File: internal/tui/trade.go
// Project: Terminal Velocity
// Version: 1.0.0

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

// Trade view modes

const (
	tradeViewReceived = "received" // Offers received from others
	tradeViewSent     = "sent"     // Offers sent by player
	tradeViewHistory  = "history"  // Trade history
	tradeViewCreate   = "create"   // Create new offer
	tradeViewDetail   = "detail"   // View offer details
)

type tradeModel struct {
	viewMode      string
	cursor        int
	selectedTrade *models.TradeOffer

	// Create mode fields
	createRecipient        string
	createOfferedCredits   int64
	createRequestedCredits int64
	createMessage          string
	createInputField       int // 0=recipient, 1=offered, 2=requested, 3=message
}

func newTradeModel() tradeModel {
	return tradeModel{
		viewMode:         tradeViewReceived,
		cursor:           0,
		createInputField: 0,
	}
}

func (m Model) updateTrade(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.tradeModel.viewMode {
		case tradeViewCreate:
			return m.updateTradeCreate(msg)
		case tradeViewDetail:
			return m.updateTradeDetail(msg)
		default:
			return m.updateTradeList(msg)
		}
	}

	return m, nil
}

func (m Model) updateTradeList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.tradeModel.cursor > 0 {
			m.tradeModel.cursor--
		}

	case "down", "j":
		var offers []*models.TradeOffer
		switch m.tradeModel.viewMode {
		case tradeViewReceived:
			offers = m.tradeManager.GetPendingOffers(m.playerID)
		case tradeViewSent:
			offers = m.tradeManager.GetSentOffers(m.playerID)
		}

		if m.tradeModel.cursor < len(offers)-1 {
			m.tradeModel.cursor++
		}

	case "1":
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.cursor = 0

	case "2":
		m.tradeModel.viewMode = tradeViewSent
		m.tradeModel.cursor = 0

	case "3":
		m.tradeModel.viewMode = tradeViewHistory
		m.tradeModel.cursor = 0

	case "n":
		// Create new trade
		m.tradeModel.viewMode = tradeViewCreate
		m.tradeModel.createRecipient = ""
		m.tradeModel.createOfferedCredits = 0
		m.tradeModel.createRequestedCredits = 0
		m.tradeModel.createMessage = ""
		m.tradeModel.createInputField = 0

	case "enter":
		// View trade details
		var offers []*models.TradeOffer
		switch m.tradeModel.viewMode {
		case tradeViewReceived:
			offers = m.tradeManager.GetPendingOffers(m.playerID)
		case tradeViewSent:
			offers = m.tradeManager.GetSentOffers(m.playerID)
		}

		if m.tradeModel.cursor < len(offers) {
			m.tradeModel.selectedTrade = offers[m.tradeModel.cursor]
			m.tradeModel.viewMode = tradeViewDetail
		}

	case "q", "esc":
		m.screen = ScreenMainMenu
	}

	return m, nil
}

func (m Model) updateTradeCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.tradeModel.createInputField = (m.tradeModel.createInputField + 1) % 4

	case "esc":
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.cursor = 0

	case "enter":
		// Create the trade offer
		if m.tradeModel.createRecipient == "" {
			// Don't create without a recipient
			return m, nil
		}

		// Find recipient player
		recipient, err := m.playerRepo.GetByUsername(context.Background(), m.tradeModel.createRecipient)
		if err != nil {
			// Invalid recipient, just reset for now
			m.tradeModel.viewMode = tradeViewReceived
			m.tradeModel.cursor = 0
			return m, nil
		}

		// Check both players are in same system and location
		if recipient.CurrentSystem != m.player.CurrentSystem {
			// Players must be in same system
			return m, nil
		}

		// If either is docked, both must be docked at same planet
		if m.player.CurrentPlanet != nil || recipient.CurrentPlanet != nil {
			if m.player.CurrentPlanet == nil || recipient.CurrentPlanet == nil ||
				*m.player.CurrentPlanet != *recipient.CurrentPlanet {
				// Players must be at same location
				return m, nil
			}
		}

		// Create the trade offer
		planetID := uuid.Nil
		if m.player.CurrentPlanet != nil {
			planetID = *m.player.CurrentPlanet
		}

		offer := m.tradeManager.CreateOffer(
			m.playerID,
			m.player.Username,
			recipient.ID,
			recipient.Username,
			m.player.CurrentSystem,
			planetID,
		)

		// Set offered/requested credits
		offer.OfferedCredits = m.tradeModel.createOfferedCredits
		offer.RequestedCredits = m.tradeModel.createRequestedCredits

		// Reset form and return to list
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.cursor = 0
		m.tradeModel.createRecipient = ""
		m.tradeModel.createOfferedCredits = 0
		m.tradeModel.createRequestedCredits = 0
		m.tradeModel.createMessage = ""
		m.tradeModel.createInputField = 0

	case "backspace":
		switch m.tradeModel.createInputField {
		case 0: // Recipient
			if len(m.tradeModel.createRecipient) > 0 {
				m.tradeModel.createRecipient = m.tradeModel.createRecipient[:len(m.tradeModel.createRecipient)-1]
			}
		case 3: // Message
			if len(m.tradeModel.createMessage) > 0 {
				m.tradeModel.createMessage = m.tradeModel.createMessage[:len(m.tradeModel.createMessage)-1]
			}
		}

	case "up":
		switch m.tradeModel.createInputField {
		case 1: // Offered credits
			m.tradeModel.createOfferedCredits += 1000
		case 2: // Requested credits
			m.tradeModel.createRequestedCredits += 1000
		}

	case "down":
		switch m.tradeModel.createInputField {
		case 1: // Offered credits
			if m.tradeModel.createOfferedCredits >= 1000 {
				m.tradeModel.createOfferedCredits -= 1000
			}
		case 2: // Requested credits
			if m.tradeModel.createRequestedCredits >= 1000 {
				m.tradeModel.createRequestedCredits -= 1000
			}
		}

	default:
		// Handle text input
		if len(msg.String()) == 1 {
			switch m.tradeModel.createInputField {
			case 0: // Recipient
				m.tradeModel.createRecipient += msg.String()
			case 3: // Message
				if len(m.tradeModel.createMessage) < 200 {
					m.tradeModel.createMessage += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m Model) updateTradeDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tradeModel.selectedTrade == nil {
		m.tradeModel.viewMode = tradeViewReceived
		return m, nil
	}

	switch msg.String() {
	case "a":
		// Accept trade
		if m.tradeModel.selectedTrade.RecipientID == m.playerID &&
			m.tradeModel.selectedTrade.Status == models.TradeStatusPending {
			err := m.tradeManager.AcceptOffer(m.tradeModel.selectedTrade.ID, m.playerID)
			if err == nil {
				// Trade accepted - the trade manager handles escrow internally
				// Credits and items are deducted when offer is created/accepted
				// and transferred when trade is completed
				_ = m.tradeManager.CompleteTrade(m.tradeModel.selectedTrade.ID)
			}
		}
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.selectedTrade = nil

	case "r":
		// Reject trade
		if m.tradeModel.selectedTrade.RecipientID == m.playerID &&
			m.tradeModel.selectedTrade.Status == models.TradeStatusPending {
			_ = m.tradeManager.RejectOffer(m.tradeModel.selectedTrade.ID, m.playerID)
		}
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.selectedTrade = nil

	case "c":
		// Cancel trade
		if m.tradeModel.selectedTrade.InitiatorID == m.playerID &&
			m.tradeModel.selectedTrade.CanBeCancelled() {
			_ = m.tradeManager.CancelOffer(m.tradeModel.selectedTrade.ID, m.playerID)
		}
		m.tradeModel.viewMode = tradeViewSent
		m.tradeModel.selectedTrade = nil

	case "esc", "q":
		m.tradeModel.viewMode = tradeViewReceived
		m.tradeModel.selectedTrade = nil
	}

	return m, nil
}

func (m Model) viewTrade() string {
	switch m.tradeModel.viewMode {
	case tradeViewCreate:
		return m.viewTradeCreate()
	case tradeViewDetail:
		return m.viewTradeDetail()
	case tradeViewHistory:
		return m.viewTradeHistory()
	default:
		return m.viewTradeList()
	}
}

func (m Model) viewTradeList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(0, 1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))

	activeTabStyle := tabStyle.Copy().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Background(lipgloss.Color("236"))

	// Header
	var s strings.Builder
	s.WriteString(titleStyle.Render("💱 Player Trading"))
	s.WriteString("\n\n")

	// Tabs
	tabs := []string{"Received (1)", "Sent (2)", "History (3)"}
	tabViews := []string{tradeViewReceived, tradeViewSent, tradeViewHistory}

	for i, tab := range tabs {
		if m.tradeModel.viewMode == tabViews[i] {
			s.WriteString(activeTabStyle.Render(tab))
		} else {
			s.WriteString(tabStyle.Render(tab))
		}
	}
	s.WriteString("\n\n")

	// Get offers based on current view
	var offers []*models.TradeOffer
	switch m.tradeModel.viewMode {
	case tradeViewReceived:
		offers = m.tradeManager.GetPendingOffers(m.playerID)
		s.WriteString("Offers received from other players:\n\n")
	case tradeViewSent:
		offers = m.tradeManager.GetSentOffers(m.playerID)
		s.WriteString("Offers you've sent to other players:\n\n")
	}

	// List offers
	if len(offers) == 0 {
		s.WriteString("  No trades available\n")
	} else {
		for i, offer := range offers {
			cursor := "  "
			if i == m.tradeModel.cursor {
				cursor = "→ "
			}

			otherPlayer := offer.RecipientName
			if m.tradeModel.viewMode == tradeViewReceived {
				otherPlayer = offer.InitiatorName
			}

			status := offer.Status.GetIcon()
			fairness := offer.GetFairnessRating()
			timeRemaining := offer.GetTimeRemaining()

			line := fmt.Sprintf("%s%s %s | Offered: %d cr | Requested: %d cr | %s | %s",
				cursor,
				status,
				otherPlayer,
				offer.GetTotalOfferedValue(),
				offer.GetTotalRequestedValue(),
				fairness,
				timeRemaining,
			)

			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString("Controls: [↑/↓] Navigate [Enter] View Details [N] New Trade [Q] Back\n")

	return boxStyle.Render(s.String())
}

func (m Model) viewTradeCreate() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Background(lipgloss.Color("236"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("💱 Create Trade Offer"))
	s.WriteString("\n\n")

	// Recipient field
	recipientLabel := labelStyle.Render("Recipient:")
	recipientValue := m.tradeModel.createRecipient
	if m.tradeModel.createInputField == 0 {
		recipientValue = activeStyle.Render(recipientValue + "_")
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", recipientLabel, recipientValue))

	// Offered credits
	offeredLabel := labelStyle.Render("Offering Credits:")
	offeredValue := fmt.Sprintf("%d cr", m.tradeModel.createOfferedCredits)
	if m.tradeModel.createInputField == 1 {
		offeredValue = activeStyle.Render(offeredValue)
	}
	s.WriteString(fmt.Sprintf("%s %s (↑/↓ to adjust)\n\n", offeredLabel, offeredValue))

	// Requested credits
	requestedLabel := labelStyle.Render("Requesting Credits:")
	requestedValue := fmt.Sprintf("%d cr", m.tradeModel.createRequestedCredits)
	if m.tradeModel.createInputField == 2 {
		requestedValue = activeStyle.Render(requestedValue)
	}
	s.WriteString(fmt.Sprintf("%s %s (↑/↓ to adjust)\n\n", requestedLabel, requestedValue))

	// Message field
	messageLabel := labelStyle.Render("Message:")
	messageValue := m.tradeModel.createMessage
	if m.tradeModel.createInputField == 3 {
		messageValue = activeStyle.Render(messageValue + "_")
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", messageLabel, messageValue))

	s.WriteString("Note: In full implementation, you would also add cargo items\n\n")

	s.WriteString("Controls: [Tab] Next Field [Enter] Send Offer [Esc] Cancel\n")

	return boxStyle.Render(s.String())
}

func (m Model) viewTradeDetail() string {
	if m.tradeModel.selectedTrade == nil {
		return "No trade selected"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	offer := m.tradeModel.selectedTrade

	var s strings.Builder
	s.WriteString(titleStyle.Render("💱 Trade Offer Details"))
	s.WriteString("\n\n")

	// Status
	statusIcon := offer.Status.GetIcon()
	s.WriteString(fmt.Sprintf("%s Status: %s\n\n", statusIcon, offer.Status))

	// Parties
	s.WriteString(labelStyle.Render("From: ") + offer.InitiatorName + "\n")
	s.WriteString(labelStyle.Render("To: ") + offer.RecipientName + "\n\n")

	// Offered section
	s.WriteString(labelStyle.Render("━━━ Offering ━━━") + "\n")
	s.WriteString(fmt.Sprintf("Credits: %d cr\n", offer.OfferedCredits))
	if len(offer.OfferedItems) > 0 {
		s.WriteString("Items:\n")
		for _, item := range offer.OfferedItems {
			s.WriteString(fmt.Sprintf("  • %s x%d @ %d cr/unit\n",
				item.CommodityName, item.Quantity, item.UnitPrice))
		}
	}
	s.WriteString(fmt.Sprintf("Total Value: %d cr\n\n", offer.GetTotalOfferedValue()))

	// Requested section
	s.WriteString(labelStyle.Render("━━━ Requesting ━━━") + "\n")
	s.WriteString(fmt.Sprintf("Credits: %d cr\n", offer.RequestedCredits))
	if len(offer.RequestedItems) > 0 {
		s.WriteString("Items:\n")
		for _, item := range offer.RequestedItems {
			s.WriteString(fmt.Sprintf("  • %s x%d @ %d cr/unit\n",
				item.CommodityName, item.Quantity, item.UnitPrice))
		}
	}
	s.WriteString(fmt.Sprintf("Total Value: %d cr\n\n", offer.GetTotalRequestedValue()))

	// Fairness assessment
	s.WriteString(fmt.Sprintf("Assessment: %s\n\n", offer.GetFairnessRating()))

	// Message
	if offer.Message != "" {
		s.WriteString(labelStyle.Render("Message: ") + offer.Message + "\n\n")
	}

	// Time info
	if offer.Status == models.TradeStatusPending {
		s.WriteString(fmt.Sprintf("Time Remaining: %s\n\n", offer.GetTimeRemaining()))
	}

	// Trader reputation
	var traderID uuid.UUID
	if offer.RecipientID == m.playerID {
		traderID = offer.InitiatorID
	} else {
		traderID = offer.RecipientID
	}
	history := m.tradeManager.GetHistory(traderID)
	s.WriteString(labelStyle.Render("Trader Reputation: ") + history.GetTrustRating() + "\n")
	s.WriteString(fmt.Sprintf("Completed Trades: %d | Success Rate: %.1f%%\n\n",
		history.SuccessfulTrades, history.GetCompletionRate()))

	// Controls based on role and status
	if offer.RecipientID == m.playerID && offer.Status == models.TradeStatusPending {
		s.WriteString("Controls: [A] Accept [R] Reject [Q] Back\n")
	} else if offer.InitiatorID == m.playerID && offer.CanBeCancelled() {
		s.WriteString("Controls: [C] Cancel [Q] Back\n")
	} else {
		s.WriteString("Controls: [Q] Back\n")
	}

	return boxStyle.Render(s.String())
}

func (m Model) viewTradeHistory() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("💱 Your Trading History"))
	s.WriteString("\n\n")

	history := m.tradeManager.GetHistory(m.playerID)

	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Total Trades:"), history.TotalTrades))
	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Successful:"), history.SuccessfulTrades))
	s.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Cancelled:"), history.CancelledTrades))
	s.WriteString(fmt.Sprintf("%s %.1f%%\n\n", labelStyle.Render("Success Rate:"), history.GetCompletionRate()))

	s.WriteString(fmt.Sprintf("%s %d cr\n\n", labelStyle.Render("Total Volume:"), history.TotalVolume))

	s.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Trust Rating:"), history.GetTrustRating()))
	s.WriteString(fmt.Sprintf("%s %d 👍 | %d 👎\n\n", labelStyle.Render("Feedback:"),
		history.PositiveRatings, history.NegativeRatings))

	// Recent trades
	s.WriteString(labelStyle.Render("━━━ Recent Trades ━━━") + "\n\n")

	allOffers := m.tradeManager.GetPlayerOffers(m.playerID)
	recentCount := 0
	for _, offer := range allOffers {
		if offer.Status == models.TradeStatusCompleted && recentCount < 5 {
			otherPlayer := offer.RecipientName
			if offer.RecipientID == m.playerID {
				otherPlayer = offer.InitiatorName
			}

			s.WriteString(fmt.Sprintf("• With %s | %d cr | %s\n",
				otherPlayer,
				offer.GetTotalOfferedValue(),
				offer.UpdatedAt.Format("Jan 02 15:04"),
			))
			recentCount++
		}
	}

	if recentCount == 0 {
		s.WriteString("  No completed trades yet\n")
	}

	s.WriteString("\n")
	s.WriteString("Controls: [1] Received [2] Sent [Q] Back\n")

	return boxStyle.Render(s.String())
}
