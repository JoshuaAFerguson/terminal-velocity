// File: internal/tui/trading_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced trading screen with market listings
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

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
	if m.trading.commodities == nil || len(m.trading.commodities) == 0 {
		tm := newTradingEnhancedModel()
		m.trading.commodities = tm.commodities
		m.trading.selectedCommodity = 0
	}

	// Commodity rows
	for i := 0; i < 10; i++ {
		prefix := "   "
		if i == m.trading.selectedCommodity {
			prefix = " " + IconArrow + " "
		}

		if i < len(m.trading.commodities) {
			comm := m.trading.commodities[i]
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

	if m.trading.selectedCommodity < len(m.trading.commodities) {
		comm := m.trading.commodities[m.trading.selectedCommodity]

		detailsContent.WriteString(fmt.Sprintf(" SELECTED: %-58s\n", comm.name))
		detailsContent.WriteString("                                                                      \n")
		detailsContent.WriteString(fmt.Sprintf(" Buy Price:  %d cr/ton             Sell Price: %d cr/ton             \n",
			comm.sellPrice, comm.buyPrice))

		cargoSpace := 35 // TODO: Get from ship
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

func (m Model) updateTradingEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.trading.selectedCommodity > 0 {
				m.trading.selectedCommodity--
			}
			return m, nil

		case "down", "j":
			if m.trading.selectedCommodity < len(m.trading.commodities)-1 {
				m.trading.selectedCommodity++
			}
			return m, nil

		case "b", "B":
			// Buy commodity
			// TODO: Implement buy logic via API
			return m, nil

		case "s", "S":
			// Sell commodity
			// TODO: Implement sell logic via API
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
	}

	return m, nil
}

// Add ScreenTradingEnhanced constant to Screen enum when integrating
