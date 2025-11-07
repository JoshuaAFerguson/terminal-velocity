package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/game/trading"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type tradingModel struct {
	cursor            int
	mode              string // "market", "buy", "sell", "confirm"
	selectedCommodity *models.Commodity
	quantity          int
	marketPrices      []*models.MarketPrice
	commodities       []models.Commodity
	currentPlanet     *models.Planet
	loading           bool
	error             string
	pricingEngine     *trading.PricingEngine
}

type marketLoadedMsg struct {
	prices      []*models.MarketPrice
	commodities []models.Commodity
	planet      *models.Planet
	err         error
}

type tradeCompleteMsg struct {
	success bool
	profit  int64
	err     error
}

func newTradingModel() tradingModel {
	return tradingModel{
		cursor:        0,
		mode:          "market",
		quantity:      1,
		loading:       true,
		pricingEngine: trading.NewPricingEngine(),
	}
}

func (m Model) updateTrading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.trading.mode == "market" {
				// Go back to main menu
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel current operation
			m.trading.mode = "market"
			m.trading.quantity = 1
			m.trading.error = ""
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.trading.cursor > 0 {
				m.trading.cursor--
			}

		case "down", "j":
			maxCursor := len(m.trading.commodities) - 1
			if m.trading.cursor < maxCursor {
				m.trading.cursor++
			}

		case "b":
			// Enter buy mode
			if m.trading.mode == "market" && m.trading.cursor < len(m.trading.commodities) {
				m.trading.selectedCommodity = &m.trading.commodities[m.trading.cursor]
				m.trading.mode = "buy"
				m.trading.quantity = 1
				m.trading.error = ""
			}

		case "s":
			// Enter sell mode
			if m.trading.mode == "market" && m.trading.cursor < len(m.trading.commodities) {
				m.trading.selectedCommodity = &m.trading.commodities[m.trading.cursor]
				m.trading.mode = "sell"
				m.trading.quantity = 1
				m.trading.error = ""
			}

		case "+", "=":
			// Increase quantity
			if m.trading.mode == "buy" || m.trading.mode == "sell" {
				m.trading.quantity += 1
			}

		case "-", "_":
			// Decrease quantity
			if m.trading.mode == "buy" || m.trading.mode == "sell" {
				if m.trading.quantity > 1 {
					m.trading.quantity -= 1
				}
			}

		case "enter", " ":
			// Confirm trade
			if m.trading.mode == "buy" {
				return m, m.executeBuy()
			} else if m.trading.mode == "sell" {
				return m, m.executeSell()
			}
		}

	case marketLoadedMsg:
		m.trading.loading = false
		if msg.err != nil {
			m.trading.error = fmt.Sprintf("Failed to load market: %v", msg.err)
		} else {
			m.trading.marketPrices = msg.prices
			m.trading.commodities = msg.commodities
			m.trading.currentPlanet = msg.planet
			m.trading.error = ""
		}

	case tradeCompleteMsg:
		if msg.success {
			// Trade successful - reload market
			m.trading.mode = "market"
			m.trading.quantity = 1
			m.trading.selectedCommodity = nil
			// Show profit/loss message
			if msg.profit > 0 {
				m.trading.error = fmt.Sprintf("Sale complete! Profit: %d cr", msg.profit)
			} else if msg.profit < 0 {
				m.trading.error = fmt.Sprintf("Purchase complete! Cost: %d cr", -msg.profit)
			}
			// Reload market prices
			return m, m.loadTradingMarket()
		} else {
			m.trading.error = fmt.Sprintf("Trade failed: %v", msg.err)
			m.trading.mode = "market"
		}
	}

	return m, nil
}

func (m Model) viewTrading() string {
	// Header with player stats
	locationName := "Space"
	if m.trading.currentPlanet != nil {
		locationName = m.trading.currentPlanet.Name
	}
	s := renderHeader(m.username, m.player.Credits, locationName)
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Commodity Market ===") + "\n\n"

	// Error display
	if m.trading.error != "" {
		s += errorStyle.Render("⚠ "+m.trading.error) + "\n\n"
	}

	// Loading state
	if m.trading.loading {
		s += "Loading market data...\n"
		return s
	}

	// Mode-specific view
	switch m.trading.mode {
	case "market":
		s += m.viewMarket()
	case "buy":
		s += m.viewBuyInterface()
	case "sell":
		s += m.viewSellInterface()
	default:
		s += "Unknown mode\n"
	}

	return s
}

func (m Model) viewMarket() string {
	s := ""

	// Planet info
	if m.trading.currentPlanet != nil {
		info := fmt.Sprintf("Planet: %s (Tech Level: %d)\n",
			m.trading.currentPlanet.Name,
			m.trading.currentPlanet.TechLevel)
		s += info + "\n"
	}

	// Cargo space info
	if m.currentShip != nil {
		// TODO: Get actual cargo space from ship type
		cargoUsed := m.currentShip.GetCargoUsed()
		s += fmt.Sprintf("Cargo: %s / 100\n\n", statsStyle.Render(fmt.Sprintf("%d", cargoUsed)))
	}

	// Market table header
	s += "Commodity                 Category        Buy      Sell     Stock   Demand\n"
	s += strings.Repeat("─", 78) + "\n"

	// List commodities
	for i, commodity := range m.trading.commodities {
		// Find market price
		var price *models.MarketPrice
		for _, p := range m.trading.marketPrices {
			if p.CommodityID == commodity.ID {
				price = p
				break
			}
		}

		if price == nil {
			continue // Skip if no price data
		}

		line := fmt.Sprintf("%-25s %-15s %-8d %-8d %-7d %-7d",
			commodity.Name,
			commodity.Category,
			price.BuyPrice,
			price.SellPrice,
			price.Stock,
			price.Demand,
		)

		// Highlight illegal goods
		if commodity.IsIllegal(m.trading.currentPlanet.Name) {
			line += " [ILLEGAL]"
			line = errorStyle.Render(line)
		}

		if i == m.trading.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	// Help text
	s += renderFooter("↑/↓: Select  •  B: Buy  •  S: Sell  •  ESC: Main Menu")

	return s
}

func (m Model) viewBuyInterface() string {
	if m.trading.selectedCommodity == nil {
		return "No commodity selected\n"
	}

	s := ""

	// Find market price
	var price *models.MarketPrice
	for _, p := range m.trading.marketPrices {
		if p.CommodityID == m.trading.selectedCommodity.ID {
			price = p
			break
		}
	}

	if price == nil {
		return "Price data not available\n"
	}

	// Commodity info
	s += fmt.Sprintf("Buying: %s\n", statsStyle.Render(m.trading.selectedCommodity.Name))
	s += fmt.Sprintf("Description: %s\n\n", m.trading.selectedCommodity.Description)

	// Price info
	s += fmt.Sprintf("Price per unit: %s cr\n", statsStyle.Render(fmt.Sprintf("%d", price.SellPrice)))
	s += fmt.Sprintf("Quantity: %s\n", statsStyle.Render(fmt.Sprintf("%d", m.trading.quantity)))

	totalCost := price.SellPrice * int64(m.trading.quantity)
	s += fmt.Sprintf("Total cost: %s cr\n\n", statsStyle.Render(fmt.Sprintf("%d", totalCost)))

	// Available stock
	s += fmt.Sprintf("Available: %d units\n", price.Stock)

	// Can afford?
	if totalCost > m.player.Credits {
		s += errorStyle.Render(fmt.Sprintf("Insufficient credits (need %d more)\n", totalCost-m.player.Credits))
	}

	// Cargo space
	if m.currentShip != nil {
		cargoUsed := m.currentShip.GetCargoUsed()
		cargoAvailable := 100 - cargoUsed // TODO: Get from ship type
		if m.trading.quantity > cargoAvailable {
			s += errorStyle.Render(fmt.Sprintf("Insufficient cargo space (have %d)\n", cargoAvailable))
		}
	}

	// Help text
	s += "\n" + helpStyle.Render("+/-: Adjust quantity  •  Enter: Confirm  •  ESC: Cancel")

	return s
}

func (m Model) viewSellInterface() string {
	if m.trading.selectedCommodity == nil {
		return "No commodity selected\n"
	}

	s := ""

	// Find market price
	var price *models.MarketPrice
	for _, p := range m.trading.marketPrices {
		if p.CommodityID == m.trading.selectedCommodity.ID {
			price = p
			break
		}
	}

	if price == nil {
		return "Price data not available\n"
	}

	// Commodity info
	s += fmt.Sprintf("Selling: %s\n", statsStyle.Render(m.trading.selectedCommodity.Name))
	s += fmt.Sprintf("Description: %s\n\n", m.trading.selectedCommodity.Description)

	// Price info
	s += fmt.Sprintf("Price per unit: %s cr\n", statsStyle.Render(fmt.Sprintf("%d", price.BuyPrice)))
	s += fmt.Sprintf("Quantity: %s\n", statsStyle.Render(fmt.Sprintf("%d", m.trading.quantity)))

	totalRevenue := price.BuyPrice * int64(m.trading.quantity)
	s += fmt.Sprintf("Total revenue: %s cr\n\n", statsStyle.Render(fmt.Sprintf("%d", totalRevenue)))

	// Cargo check
	if m.currentShip != nil {
		inCargo := m.currentShip.GetCommodityQuantity(m.trading.selectedCommodity.ID)
		s += fmt.Sprintf("In cargo: %d units\n", inCargo)

		if m.trading.quantity > inCargo {
			s += errorStyle.Render(fmt.Sprintf("Insufficient cargo (have %d)\n", inCargo))
		}
	}

	// Help text
	s += "\n" + helpStyle.Render("+/-: Adjust quantity  •  Enter: Confirm  •  ESC: Cancel")

	return s
}

// loadTradingMarket loads market data for the current location
func (m Model) loadTradingMarket() tea.Cmd {
	return func() tea.Msg {
		// TODO: Determine current planet from player location
		// For now, return placeholder
		// In real implementation, check if player is docked at a planet
		// and load that planet's market data

		// Placeholder: return all commodities
		commodities := models.StandardCommodities

		// Sort by category
		sort.Slice(commodities, func(i, j int) bool {
			if commodities[i].Category == commodities[j].Category {
				return commodities[i].Name < commodities[j].Name
			}
			return commodities[i].Category < commodities[j].Category
		})

		return marketLoadedMsg{
			commodities: commodities,
			err:         nil,
		}
	}
}

// executeBuy executes a buy transaction
func (m Model) executeBuy() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate we have all required data
		if m.trading.selectedCommodity == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("no commodity selected"),
			}
		}

		if m.currentShip == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("no ship available"),
			}
		}

		if m.trading.currentPlanet == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("not docked at a planet"),
			}
		}

		// Find market price
		var price *models.MarketPrice
		for _, p := range m.trading.marketPrices {
			if p.CommodityID == m.trading.selectedCommodity.ID {
				price = p
				break
			}
		}

		if price == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("price data not available"),
			}
		}

		// Calculate total cost
		totalCost := price.SellPrice * int64(m.trading.quantity)

		// Validate credits
		if totalCost > m.player.Credits {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("insufficient credits (need %d)", totalCost),
			}
		}

		// Validate cargo space
		cargoUsed := m.currentShip.GetCargoUsed()
		cargoAvailable := 100 - cargoUsed // TODO: Get from ship type
		if m.trading.quantity > cargoAvailable {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("insufficient cargo space (have %d)", cargoAvailable),
			}
		}

		// Validate stock availability
		if m.trading.quantity > price.Stock {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("insufficient stock (available: %d)", price.Stock),
			}
		}

		// Execute transaction: deduct credits
		newCredits := m.player.Credits - totalCost
		err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
		if err != nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("failed to update credits: %w", err),
			}
		}

		// Add cargo to ship
		err = m.shipRepo.AddCargo(ctx, m.currentShip.ID, m.trading.selectedCommodity.ID, m.trading.quantity)
		if err != nil {
			// Rollback: restore credits
			m.playerRepo.UpdateCredits(ctx, m.player.ID, m.player.Credits)
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("failed to add cargo: %w", err),
			}
		}

		// Update market: decrease stock, adjust prices
		price.Stock -= m.trading.quantity
		price.Demand += m.trading.quantity / 2
		price.BuyPrice, price.SellPrice = m.trading.pricingEngine.CalculateMarketPrice(
			m.trading.selectedCommodity,
			m.trading.currentPlanet,
			price.Stock,
			price.Demand,
		)
		price.LastUpdate = time.Now().Unix()

		err = m.marketRepo.UpdateMarketPrice(ctx, price)
		if err != nil {
			// Continue anyway - market update failure shouldn't block the trade
			// TODO: Log this error
		}

		// Update local player state
		m.player.Credits = newCredits

		return tradeCompleteMsg{
			success: true,
			profit:  -totalCost, // Negative because we spent money
			err:     nil,
		}
	}
}

// executeSell executes a sell transaction
func (m Model) executeSell() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate we have all required data
		if m.trading.selectedCommodity == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("no commodity selected"),
			}
		}

		if m.currentShip == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("no ship available"),
			}
		}

		if m.trading.currentPlanet == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("not docked at a planet"),
			}
		}

		// Find market price
		var price *models.MarketPrice
		for _, p := range m.trading.marketPrices {
			if p.CommodityID == m.trading.selectedCommodity.ID {
				price = p
				break
			}
		}

		if price == nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("price data not available"),
			}
		}

		// Validate cargo
		inCargo := m.currentShip.GetCommodityQuantity(m.trading.selectedCommodity.ID)
		if m.trading.quantity > inCargo {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("insufficient cargo (have %d)", inCargo),
			}
		}

		// Calculate total revenue
		totalRevenue := price.BuyPrice * int64(m.trading.quantity)

		// Execute transaction: remove cargo from ship
		err := m.shipRepo.RemoveCargo(ctx, m.currentShip.ID, m.trading.selectedCommodity.ID, m.trading.quantity)
		if err != nil {
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("failed to remove cargo: %w", err),
			}
		}

		// Add credits
		newCredits := m.player.Credits + totalRevenue
		err = m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
		if err != nil {
			// Rollback: restore cargo
			m.shipRepo.AddCargo(ctx, m.currentShip.ID, m.trading.selectedCommodity.ID, m.trading.quantity)
			return tradeCompleteMsg{
				success: false,
				err:     fmt.Errorf("failed to update credits: %w", err),
			}
		}

		// Update market: increase stock, adjust prices
		price.Stock += m.trading.quantity
		price.Demand -= m.trading.quantity / 3
		if price.Demand < 10 {
			price.Demand = 10
		}
		price.BuyPrice, price.SellPrice = m.trading.pricingEngine.CalculateMarketPrice(
			m.trading.selectedCommodity,
			m.trading.currentPlanet,
			price.Stock,
			price.Demand,
		)
		price.LastUpdate = time.Now().Unix()

		err = m.marketRepo.UpdateMarketPrice(ctx, price)
		if err != nil {
			// Continue anyway - market update failure shouldn't block the trade
			// TODO: Log this error
		}

		// Update local player state
		m.player.Credits = newCredits

		return tradeCompleteMsg{
			success: true,
			profit:  totalRevenue, // Positive because we gained money
			err:     nil,
		}
	}
}
