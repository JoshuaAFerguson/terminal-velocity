// File: internal/tui/trading_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced trading screen with market listings
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type tradingEnhancedModel struct {
	selectedCommodity int
	quantity          int
	commodities       []commodityListing
	mode              string // "view", "buy", "sell"
}

type commodityListing struct {
	name      string
	buyPrice  int
	sellPrice int
	stock     string // "High", "Med", "Low"
	inCargo   int
}

func newTradingEnhancedModel() tradingEnhancedModel {
	// Sample commodities
	commodities := []commodityListing{
		{"Food", 45, 52, "High", 10},
		{"Water", 28, 35, "Med", 0},
		{"Textiles", 110, 125, "Low", 0},
		{"Electronics", 380, 425, "Med", 5},
		{"Computers", 890, 950, "Low", 0},
		{"Weapons", 1200, 1350, "Med", 0},
		{"Medical Sup.", 450, 490, "High", 0},
		{"Luxury Goods", 2100, 2300, "Low", 0},
		{"Industrial", 180, 205, "High", 0},
		{"Minerals", 95, 110, "High", 0},
	}

	return tradingEnhancedModel{
		selectedCommodity: 0,
		quantity:          0,
		commodities:       commodities,
		mode:              "view",
	}
}

func (m Model) viewTradingEnhanced() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Header
	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	header := DrawHeader("COMMODITY EXCHANGE - Earth Station", "", credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Commodity list table
	tableWidth := width - 4
	var tableContent strings.Builder

	// Table header
	tableContent.WriteString(" COMMODITY          BUY PRICE   SELL PRICE   STOCK   YOUR CARGO       \n")
	tableContent.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Initialize commodities if not set
	if m.tradingEnhanced.commodities == nil || len(m.tradingEnhanced.commodities) == 0 {
		m.tradingEnhanced = newTradingEnhancedModel()
	}

	// Commodity rows
	for i := 0; i < 10; i++ {
		prefix := "   "
		if i == m.tradingEnhanced.selectedCommodity {
			prefix = " " + IconArrow + " "
		}

		if i < len(m.tradingEnhanced.commodities) {
			comm := m.tradingEnhanced.commodities[i]
			line := fmt.Sprintf("%s%-18s %5d cr    %5d cr   %-7s %5d tons       ",
				prefix,
				comm.name,
				comm.buyPrice,
				comm.sellPrice,
				comm.stock,
				comm.inCargo,
			)
			tableContent.WriteString(PadRight(line, tableWidth-2) + "\n")
		}
	}

	// Draw table panel
	table := DrawPanel("", tableContent.String(), tableWidth, 14, false)
	tableLines := strings.Split(table, "\n")
	for _, line := range tableLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Selected commodity details
	detailsWidth := width - 4
	var detailsContent strings.Builder

	if m.tradingEnhanced.selectedCommodity < len(m.tradingEnhanced.commodities) {
		comm := m.tradingEnhanced.commodities[m.tradingEnhanced.selectedCommodity]

		detailsContent.WriteString(fmt.Sprintf(" SELECTED: %-58s\n", comm.name))
		detailsContent.WriteString("                                                                      \n")
		detailsContent.WriteString(fmt.Sprintf(" Buy Price:  %d cr/ton             Sell Price: %d cr/ton             \n",
			comm.sellPrice, comm.buyPrice))

		// Get actual cargo space from ship
		cargoSpace := 0
		cargoUsed := 0
		if m.currentShip != nil {
			// Get cargo space from ship type
			shipType := models.GetShipTypeByID(m.currentShip.TypeID)
			if shipType != nil {
				totalSpace := shipType.CargoSpace
				cargoUsed = m.currentShip.GetCargoUsed()
				cargoSpace = totalSpace - cargoUsed
			}
		}
		detailsContent.WriteString(fmt.Sprintf(" In Cargo:   %d tons               Available Space: %d tons          \n",
			comm.inCargo, cargoSpace))
		detailsContent.WriteString("                                                                      \n")
		detailsContent.WriteString(" Quantity: [____] tons                                                \n")
		detailsContent.WriteString("                                                                      \n")
		detailsContent.WriteString(" [ Buy ]  [ Sell ]  [ Max Buy ]  [ Sell All ]                        \n")
	}

	details := DrawPanel("", detailsContent.String(), detailsWidth, 10, false)
	detailLines := strings.Split(details, "\n")
	for _, line := range detailLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Trading tip
	tipWidth := width - 4
	var tipContent strings.Builder
	tipContent.WriteString(" TIP: Food is cheap here! Best profit selling at mining colonies.    ")

	tip := DrawPanel("", tipContent.String(), tipWidth, 3, false)
	tipLines := strings.Split(tip, "\n")
	for _, line := range tipLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[↑↓] Select  [B]uy  [S]ell  [M]ax Buy  [A]ll Sell  [ESC] Back", width)
	sb.WriteString(footer)

	return sb.String()
}

// buyCommodityCmd purchases a commodity from the market
func (m Model) buyCommodityCmd(commodityName string, quantity int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate inputs
		if quantity <= 0 {
			quantity = 1 // Default to 1
		}

		// Check if player has a ship
		if m.currentShip == nil {
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("no ship equipped"),
			}
		}

		// Check if player is on a planet
		if m.player.CurrentPlanet == nil {
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("not landed on a planet"),
			}
		}

		// Get market price for this commodity
		// TODO: Get actual commodity ID from commodityName mapping
		commodityID := strings.ToLower(commodityName)
		marketPrice, err := m.marketRepo.GetMarketPrice(ctx, *m.player.CurrentPlanet, commodityID)
		if err != nil {
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("commodity not available: %w", err),
			}
		}

		// Calculate total cost
		totalCost := marketPrice.BuyPrice * int64(quantity)

		// Check if player has enough credits
		if m.player.Credits < totalCost {
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("insufficient credits (need %d, have %d)", totalCost, m.player.Credits),
			}
		}

		// Check cargo space
		// TODO: Calculate current cargo usage and check against ship capacity
		// For now, assume we have space (we'll implement this fully later)

		// Add cargo to ship
		err = m.shipRepo.AddCargo(ctx, m.currentShip.ID, commodityID, quantity)
		if err != nil {
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("failed to add cargo: %w", err),
			}
		}

		// Deduct credits
		m.player.Credits -= totalCost
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Rollback cargo addition
			_ = m.shipRepo.RemoveCargo(ctx, m.currentShip.ID, commodityID, quantity)
			return transactionCompleteMsg{
				action: "buy",
				err:    fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Update market stock (decrease)
		_ = m.marketRepo.UpdateStock(ctx, *m.player.CurrentPlanet, commodityID, -quantity)

		return transactionCompleteMsg{
			action:      "buy",
			commodityID: commodityName,
			quantity:    quantity,
			newBalance:  m.player.Credits,
			err:         nil,
		}
	}
}

// sellCommodityCmd sells a commodity to the market
func (m Model) sellCommodityCmd(commodityName string, quantity int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate inputs
		if quantity <= 0 {
			quantity = 1 // Default to 1
		}

		// Check if player has a ship
		if m.currentShip == nil {
			return transactionCompleteMsg{
				action: "sell",
				err:    fmt.Errorf("no ship equipped"),
			}
		}

		// Check if player is on a planet
		if m.player.CurrentPlanet == nil {
			return transactionCompleteMsg{
				action: "sell",
				err:    fmt.Errorf("not landed on a planet"),
			}
		}

		// TODO: Check if player actually has this commodity in cargo
		// For now, we'll proceed and let the database operation fail if they don't

		// Get market price for this commodity
		commodityID := strings.ToLower(commodityName)
		marketPrice, err := m.marketRepo.GetMarketPrice(ctx, *m.player.CurrentPlanet, commodityID)
		if err != nil {
			return transactionCompleteMsg{
				action: "sell",
				err:    fmt.Errorf("commodity not available: %w", err),
			}
		}

		// Calculate total earnings
		totalEarnings := marketPrice.SellPrice * int64(quantity)

		// Remove cargo from ship
		err = m.shipRepo.RemoveCargo(ctx, m.currentShip.ID, commodityID, quantity)
		if err != nil {
			return transactionCompleteMsg{
				action: "sell",
				err:    fmt.Errorf("failed to remove cargo (not enough in cargo?): %w", err),
			}
		}

		// Add credits
		m.player.Credits += totalEarnings
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Rollback cargo removal
			_ = m.shipRepo.AddCargo(ctx, m.currentShip.ID, commodityID, quantity)
			return transactionCompleteMsg{
				action: "sell",
				err:    fmt.Errorf("failed to add credits: %w", err),
			}
		}

		// Update market stock (increase)
		_ = m.marketRepo.UpdateStock(ctx, *m.player.CurrentPlanet, commodityID, quantity)

		return transactionCompleteMsg{
			action:      "sell",
			commodityID: commodityName,
			quantity:    quantity,
			newBalance:  m.player.Credits,
			err:         nil,
		}
	}
}

func (m Model) updateTradingEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.tradingEnhanced.selectedCommodity > 0 {
				m.tradingEnhanced.selectedCommodity--
			}
			return m, nil

		case "down", "j":
			if m.tradingEnhanced.selectedCommodity < len(m.tradingEnhanced.commodities)-1 {
				m.tradingEnhanced.selectedCommodity++
			}
			return m, nil

		case "b", "B":
			// Buy commodity
			if m.tradingEnhanced.selectedCommodity < len(m.tradingEnhanced.commodities) {
				commodity := m.tradingEnhanced.commodities[m.tradingEnhanced.selectedCommodity]
				return m, m.buyCommodityCmd(commodity.name, 1)
			}
			return m, nil

		case "s", "S":
			// Sell commodity
			if m.tradingEnhanced.selectedCommodity < len(m.tradingEnhanced.commodities) {
				commodity := m.tradingEnhanced.commodities[m.tradingEnhanced.selectedCommodity]
				return m, m.sellCommodityCmd(commodity.name, 1)
			}
			return m, nil

		case "m", "M":
			// Max buy
			// TODO: Calculate and buy maximum affordable
			return m, nil

		case "a", "A":
			// Sell all
			// TODO: Sell all of selected commodity
			return m, nil

		case "esc":
			// Back to landing
			m.screen = ScreenLanding
			return m, nil
		}

	case transactionCompleteMsg:
		// Handle buy/sell completion
		if msg.err != nil {
			// Show error message
			m.errorMessage = fmt.Sprintf("%s failed: %v", msg.action, msg.err)
			m.showErrorDialog = true
		} else {
			// Show success message
			var actionText string
			if msg.action == "buy" {
				actionText = "Purchased"
			} else {
				actionText = "Sold"
			}
			m.errorMessage = fmt.Sprintf("%s %d %s. Balance: %d credits",
				actionText, msg.quantity, msg.commodityID, msg.newBalance)
			m.showErrorDialog = true
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenTradingEnhanced constant to Screen enum when integrating
