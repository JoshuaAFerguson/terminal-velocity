// File: internal/tui/trading.go
// Project: Terminal Velocity
// Description: Trading screen - Commodity market and dynamic economy interface
// Version: 1.3.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The trading screen provides access to planetary commodity markets:
// - View market prices for all 15 commodities
// - Buy commodities with credit and cargo space checks
// - Sell commodities from ship's cargo hold
// - Real-time price adjustments based on supply and demand
// - Tech level filtering (higher tech planets offer more commodities)
// - Illegal goods detection and warnings
// - Trade profit/loss tracking for player progression
// - Achievement notifications for trading milestones
//
// Trading Mechanics:
// - Prices fluctuate based on stock and demand
// - Player trades affect market conditions
// - Cargo space limits enforced by ship type
// - Transaction rollback on database errors
// - 50% price adjustment on supply/demand changes

package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/game/trading"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// tradingModel contains the state for the trading screen.
// Manages market display, buy/sell operations, and pricing calculations.
type tradingModel struct {
	cursor            int                     // Current cursor position in commodity list
	mode              string                  // Current mode: "market", "buy", "sell", "confirm"
	selectedCommodity *models.Commodity       // Commodity selected for trade operation
	quantity          int                     // Quantity to buy/sell (adjustable with +/-)
	marketPrices      []*models.MarketPrice   // Current market prices for all commodities
	commodities       []models.Commodity      // List of all available commodities
	currentPlanet     *models.Planet          // Current planet (market location)
	loading           bool                    // True while loading market data
	error             string                  // Error or status message to display
	pricingEngine     *trading.PricingEngine  // Engine for dynamic price calculations
}

// marketLoadedMsg is sent when market data has been loaded from database.
// Contains commodity list, prices, planet data, and any loading error.
type marketLoadedMsg struct {
	prices      []*models.MarketPrice   // Market prices for current planet
	commodities []models.Commodity      // All commodity definitions
	planet      *models.Planet          // Current planet data
	err         error                   // Error if loading failed
}

// tradeCompleteMsg is sent when a buy/sell transaction completes.
// Contains success status, profit/loss amount, and any transaction error.
type tradeCompleteMsg struct {
	success bool   // True if trade succeeded
	profit  int64  // Profit (positive for sell) or cost (negative for buy)
	err     error  // Error if trade failed
}

// newTradingModel creates and initializes a new trading screen model.
// Sets loading flag to true to trigger market data load on screen entry.
// Initializes pricing engine for dynamic market calculations.
func newTradingModel() tradingModel {
	return tradingModel{
		cursor:        0,
		mode:          "market",
		quantity:      1,
		loading:       true,
		pricingEngine: trading.NewPricingEngine(),
	}
}

// updateTrading handles input and state updates for the trading screen.
//
// Key Bindings (Market Mode):
//   - esc: Cancel operation or return to main menu
//   - backspace: Quick return to main menu
//   - up/k: Move cursor up in commodity list
//   - down/j: Move cursor down in commodity list
//   - b: Enter buy mode for selected commodity
//   - s: Enter sell mode for selected commodity
//
// Key Bindings (Buy/Sell Mode):
//   - esc: Cancel and return to market mode
//   - +/=: Increase quantity
//   - -/_: Decrease quantity
//   - enter/space: Confirm transaction
//
// Trading Workflow:
//   1. Market loads with prices and commodities
//   2. Player selects commodity and mode (buy/sell)
//   3. Player adjusts quantity with +/-
//   4. Player confirms with enter
//   5. Transaction executes with validation (credits, cargo, stock)
//   6. Market prices update based on new supply/demand
//   7. Player stats update (credits, trading rating, achievements)
//   8. Return to market mode with success message
//
// Message Handling:
//   - marketLoadedMsg: Display market with prices and commodities
//   - tradeCompleteMsg: Update player state, show profit/loss, check achievements
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

			// Record trade for player progression
			if m.player != nil {
				m.player.RecordTrade(msg.profit)

				// Check for achievement unlocks
				m.checkAchievements()

				// Show rank update on milestone achievements
				tradingRating := m.player.TradingRating
				if tradingRating > 0 && tradingRating%10 == 0 { // Every 10 points
					rankTitle := m.player.GetTradingRankTitle()
					m.trading.error = fmt.Sprintf("Trading Rank: %s (Rating: %d)", rankTitle, tradingRating)
				}

				// Show achievement notification if any
				if notification := m.getAchievementNotification(); notification != "" {
					if m.trading.error == "" {
						m.trading.error = notification
					} else {
						m.trading.error += "\n" + notification
					}
					m.clearAchievementNotification()
				}
			}

			// Show profit/loss message
			if msg.profit > 0 {
				if m.trading.error == "" { // Only if not showing rank update
					m.trading.error = fmt.Sprintf("Sale complete! Profit: %d cr", msg.profit)
				}
			} else if msg.profit < 0 {
				if m.trading.error == "" {
					m.trading.error = fmt.Sprintf("Purchase complete! Cost: %d cr", -msg.profit)
				}
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

// viewTrading renders the trading screen.
//
// Layout (Market Mode):
//   - Header: Player stats (name, credits, planet location)
//   - Title: "=== Commodity Market ==="
//   - Planet Info: Planet name and tech level
//   - Cargo Info: Current/max cargo space
//   - Market Table: Commodities with buy/sell prices, stock, demand
//   - Footer: Key bindings help
//
// Layout (Buy/Sell Mode):
//   - Commodity Info: Name, description
//   - Price Info: Unit price and total cost/revenue
//   - Validation: Credit/cargo/stock checks with error messages
//   - Footer: Quantity adjustment and confirmation help
//
// Visual Features:
//   - Illegal goods highlighted in red with [ILLEGAL] tag
//   - Insufficient resources shown in dimmed text with warnings
//   - Selected item highlighted with cursor
//   - Sorted by category, then name
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
		// Get cargo space from ship type
		cargoSpace := 100 // Default fallback
		shipType := models.GetShipTypeByID(m.currentShip.TypeID)
		if shipType != nil {
			cargoSpace = shipType.CargoSpace
		}
		cargoUsed := m.currentShip.GetCargoUsed()
		s += fmt.Sprintf("Cargo: %s / %d\n\n", statsStyle.Render(fmt.Sprintf("%d", cargoUsed)), cargoSpace)
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
		// Get cargo space from ship type
		cargoSpace := 100 // Default fallback
		shipType := models.GetShipTypeByID(m.currentShip.TypeID)
		if shipType != nil {
			cargoSpace = shipType.CargoSpace
		}
		cargoUsed := m.currentShip.GetCargoUsed()
		cargoAvailable := cargoSpace - cargoUsed
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
		ctx := context.Background()

		// Determine current planet from player location
		if m.player.CurrentPlanet == nil {
			return marketLoadedMsg{
				err: fmt.Errorf("not docked at a planet - cannot access market"),
			}
		}

		// Load planet data
		planet, err := m.systemRepo.GetPlanetByID(ctx, *m.player.CurrentPlanet)
		if err != nil {
			return marketLoadedMsg{
				err: fmt.Errorf("failed to load planet data: %w", err),
			}
		}

		// Load market prices for this planet
		prices, err := m.marketRepo.GetMarketPricesForPlanet(ctx, planet.ID)
		if err != nil {
			return marketLoadedMsg{
				err: fmt.Errorf("failed to load market prices: %w", err),
			}
		}

		// Get all commodities
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
			prices:      prices,
			planet:      planet,
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
		cargoSpace := 100 // Default fallback
		shipType := models.GetShipTypeByID(m.currentShip.TypeID)
		if shipType != nil {
			cargoSpace = shipType.CargoSpace
		}
		cargoUsed := m.currentShip.GetCargoUsed()
		cargoAvailable := cargoSpace - cargoUsed
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
			logger.Warn("Failed to update market price after buy: commodityID=%s, error=%v", m.trading.selectedCommodity.ID, err)
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
			logger.Warn("Failed to update market price after sell: commodityID=%s, error=%v", m.trading.selectedCommodity.ID, err)
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
