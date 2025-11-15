// File: internal/tui/marketplace.go
// Project: Terminal Velocity
// Description: Marketplace TUI screen for auctions, contracts, and bounties
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/marketplace"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// Marketplace screen modes
const (
	marketplaceModeMenu      = "menu"
	marketplaceModeAuctions  = "auctions"
	marketplaceModeContracts = "contracts"
	marketplaceModeBounties  = "bounties"
	marketplaceModeCreateAuction  = "create_auction"
	marketplaceModeCreateContract = "create_contract"
	marketplaceModePostBounty     = "post_bounty"
	marketplaceModeViewAuction    = "view_auction"
	marketplaceModeViewContract   = "view_contract"
	marketplaceModeViewBounty     = "view_bounty"
)

type marketplaceState struct {
	mode           string
	menuIndex      int
	selectedIndex  int
	viewOffset     int

	// Auction data
	auctions       []*marketplace.Auction
	selectedAuction *marketplace.Auction
	bidAmount      int64

	// Contract data
	contracts      []*marketplace.Contract
	selectedContract *marketplace.Contract

	// Bounty data
	bounties       []*marketplace.Bounty
	selectedBounty *marketplace.Bounty

	// Creation forms
	createForm     map[string]string
	formField      int

	// Item picker for auction creation
	itemPicker      *ItemPickerModel
	showItemPicker  bool
	selectedItemID  uuid.UUID // Selected item for auction

	loading        bool
	error          string
	message        string
}

func newMarketplaceState() marketplaceState {
	return marketplaceState{
		mode:       marketplaceModeMenu,
		menuIndex:  0,
		createForm: make(map[string]string),
	}
}

var (
	marketplaceMenuItems = []string{
		"Browse Auctions",
		"My Auctions",
		"Create Auction",
		"Browse Contracts",
		"My Contracts",
		"Post Contract",
		"View Bounties",
		"Post Bounty",
		"Back to Main Menu",
	}
)

// updateMarketplace handles all marketplace screen updates
func (m *Model) updateMarketplace(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.marketplace.mode {
		case marketplaceModeMenu:
			return m.updateMarketplaceMenu(msg)
		case marketplaceModeAuctions:
			return m.updateMarketplaceAuctions(msg)
		case marketplaceModeContracts:
			return m.updateMarketplaceContracts(msg)
		case marketplaceModeBounties:
			return m.updateMarketplaceBounties(msg)
		case marketplaceModeViewAuction:
			return m.updateMarketplaceViewAuction(msg)
		case marketplaceModeViewContract:
			return m.updateMarketplaceViewContract(msg)
		case marketplaceModeViewBounty:
			return m.updateMarketplaceViewBounty(msg)
		case marketplaceModeCreateAuction:
			return m.updateMarketplaceCreateAuction(msg)
		case marketplaceModeCreateContract:
			return m.updateMarketplaceCreateContract(msg)
		case marketplaceModePostBounty:
			return m.updateMarketplacePostBounty(msg)
		}

	case marketplaceAuctionCreatedMsg:
		m.marketplace.loading = false
		if msg.err == "" {
			m.marketplace.message = "Auction created successfully!"
			m.marketplace.mode = marketplaceModeMenu
			m.marketplace.error = ""
		} else {
			m.marketplace.error = msg.err
		}
		return m, nil
	}

	return m, nil
}

// updateMarketplaceMenu handles main marketplace menu
func (m *Model) updateMarketplaceMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.marketplace.menuIndex > 0 {
			m.marketplace.menuIndex--
		}
	case "down", "j":
		if m.marketplace.menuIndex < len(marketplaceMenuItems)-1 {
			m.marketplace.menuIndex++
		}
	case "enter":
		switch m.marketplace.menuIndex {
		case 0: // Browse Auctions
			m.marketplace.mode = marketplaceModeAuctions
			m.marketplace.selectedIndex = 0
			// Load auctions from marketplace manager
			if m.marketplaceManager != nil {
				m.marketplace.auctions = m.marketplaceManager.GetActiveAuctions()
			}
		case 1: // My Auctions
			m.marketplace.mode = marketplaceModeAuctions
			m.marketplace.selectedIndex = 0
			// Load player's auctions (filter by seller)
			if m.marketplaceManager != nil {
				allAuctions := m.marketplaceManager.GetActiveAuctions()
				playerAuctions := []*marketplace.Auction{}
				for _, auction := range allAuctions {
					if auction.SellerID == m.playerID {
						playerAuctions = append(playerAuctions, auction)
					}
				}
				m.marketplace.auctions = playerAuctions
			}
		case 2: // Create Auction
			m.marketplace.mode = marketplaceModeCreateAuction
			m.marketplace.createForm = make(map[string]string)
			m.marketplace.formField = 0
			m.marketplace.selectedItemID = uuid.Nil
			m.marketplace.showItemPicker = false
			// Initialize item picker
			m.marketplace.itemPicker = NewItemPicker(m.itemRepo, m.playerID)
			m.marketplace.itemPicker.SetMode(ItemPickerModeSingle)
			m.marketplace.itemPicker.SetFilter(FilterAvailable) // Only available items
			m.marketplace.itemPicker.SetTitle("Select Item to Auction")
			return m, m.marketplace.itemPicker.LoadItems()
		case 3: // Browse Contracts
			m.marketplace.mode = marketplaceModeContracts
			m.marketplace.selectedIndex = 0
			// Load contracts from marketplace manager
			if m.marketplaceManager != nil {
				m.marketplace.contracts = m.marketplaceManager.GetOpenContracts()
			}
		case 4: // My Contracts
			m.marketplace.mode = marketplaceModeContracts
			m.marketplace.selectedIndex = 0
			// Load player's contracts (filter by poster or claimer)
			if m.marketplaceManager != nil {
				allContracts := m.marketplaceManager.GetOpenContracts()
				playerContracts := []*marketplace.Contract{}
				for _, contract := range allContracts {
					if contract.PosterID == m.playerID || contract.ClaimedBy == m.playerID {
						playerContracts = append(playerContracts, contract)
					}
				}
				m.marketplace.contracts = playerContracts
			}
		case 5: // Post Contract
			m.marketplace.mode = marketplaceModeCreateContract
			m.marketplace.createForm = make(map[string]string)
			m.marketplace.formField = 0
		case 6: // View Bounties
			m.marketplace.mode = marketplaceModeBounties
			m.marketplace.selectedIndex = 0
			// Load bounties from marketplace manager
			if m.marketplaceManager != nil {
				m.marketplace.bounties = m.marketplaceManager.GetActiveBounties()
			}
		case 7: // Post Bounty
			m.marketplace.mode = marketplaceModePostBounty
			m.marketplace.createForm = make(map[string]string)
			m.marketplace.formField = 0
		case 8: // Back
			m.screen = ScreenMainMenu
		}
	case "q", "esc":
		m.screen = ScreenMainMenu
	}

	return m, nil
}

// updateMarketplaceAuctions handles auction browsing
func (m *Model) updateMarketplaceAuctions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.marketplace.selectedIndex > 0 {
			m.marketplace.selectedIndex--
			if m.marketplace.selectedIndex < m.marketplace.viewOffset {
				m.marketplace.viewOffset = m.marketplace.selectedIndex
			}
		}
	case "down", "j":
		if m.marketplace.selectedIndex < len(m.marketplace.auctions)-1 {
			m.marketplace.selectedIndex++
			if m.marketplace.selectedIndex >= m.marketplace.viewOffset+10 {
				m.marketplace.viewOffset++
			}
		}
	case "enter":
		if len(m.marketplace.auctions) > 0 && m.marketplace.selectedIndex < len(m.marketplace.auctions) {
			m.marketplace.selectedAuction = m.marketplace.auctions[m.marketplace.selectedIndex]
			m.marketplace.mode = marketplaceModeViewAuction
			m.marketplace.bidAmount = 0
		}
	case "b":
		m.marketplace.mode = marketplaceModeMenu
	case "q", "esc":
		m.marketplace.mode = marketplaceModeMenu
	}

	return m, nil
}

// updateMarketplaceViewAuction handles viewing a specific auction
func (m *Model) updateMarketplaceViewAuction(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	auction := m.marketplace.selectedAuction
	if auction == nil {
		m.marketplace.mode = marketplaceModeAuctions
		return m, nil
	}

	switch msg.String() {
	case "b":
		// Place bid - Implement bid logic
		if auction.Status == "active" && time.Now().Before(auction.EndTime) && m.marketplaceManager != nil {
			// Calculate minimum bid
			minBid := auction.StartingBid
			if auction.CurrentBid > 0 {
				minBid = int64(float64(auction.CurrentBid) * 1.05) // 5% increment
			}
			err := m.marketplaceManager.PlaceBid(context.Background(), auction.ID, m.playerID, m.username, minBid)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to place bid: %v", err)
			} else {
				m.marketplace.bidAmount = minBid
				m.marketplace.message = fmt.Sprintf("Bid placed: %d credits", minBid)
			}
		}
	case "o":
		// Buyout - Implement buyout logic
		if auction.BuyoutPrice > 0 && auction.Status == "active" && m.marketplaceManager != nil {
			err := m.marketplaceManager.Buyout(context.Background(), auction.ID, m.playerID)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to buyout: %v", err)
			} else {
				m.marketplace.message = fmt.Sprintf("Bought out for %d credits!", auction.BuyoutPrice)
				m.marketplace.mode = marketplaceModeAuctions
			}
		}
	case "c":
		// Cancel auction - Implement cancel logic
		if auction.SellerID == m.playerID && auction.Status == "active" && auction.CurrentBid == 0 && m.marketplaceManager != nil {
			err := m.marketplaceManager.CancelAuction(context.Background(), auction.ID, m.playerID)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to cancel: %v", err)
			} else {
				m.marketplace.message = "Auction cancelled"
				m.marketplace.mode = marketplaceModeAuctions
			}
		}
	case "q", "esc":
		m.marketplace.mode = marketplaceModeAuctions
		m.marketplace.selectedAuction = nil
	}

	return m, nil
}

// updateMarketplaceContracts handles contract browsing
func (m *Model) updateMarketplaceContracts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.marketplace.selectedIndex > 0 {
			m.marketplace.selectedIndex--
			if m.marketplace.selectedIndex < m.marketplace.viewOffset {
				m.marketplace.viewOffset = m.marketplace.selectedIndex
			}
		}
	case "down", "j":
		if m.marketplace.selectedIndex < len(m.marketplace.contracts)-1 {
			m.marketplace.selectedIndex++
			if m.marketplace.selectedIndex >= m.marketplace.viewOffset+10 {
				m.marketplace.viewOffset++
			}
		}
	case "enter":
		if len(m.marketplace.contracts) > 0 && m.marketplace.selectedIndex < len(m.marketplace.contracts) {
			m.marketplace.selectedContract = m.marketplace.contracts[m.marketplace.selectedIndex]
			m.marketplace.mode = marketplaceModeViewContract
		}
	case "b":
		m.marketplace.mode = marketplaceModeMenu
	case "q", "esc":
		m.marketplace.mode = marketplaceModeMenu
	}

	return m, nil
}

// updateMarketplaceViewContract handles viewing a specific contract
func (m *Model) updateMarketplaceViewContract(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	contract := m.marketplace.selectedContract
	if contract == nil {
		m.marketplace.mode = marketplaceModeContracts
		return m, nil
	}

	switch msg.String() {
	case "c":
		// Claim contract - Implement claim logic
		if contract.Status == "open" && contract.PosterID != m.playerID && m.marketplaceManager != nil {
			err := m.marketplaceManager.ClaimContract(context.Background(), contract.ID, m.playerID, m.username)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to claim: %v", err)
			} else {
				m.marketplace.message = "Contract claimed!"
			}
		}
	case "enter":
		// Complete contract - Implement completion logic
		if contract.Status == "claimed" && contract.ClaimedBy == m.playerID && m.marketplaceManager != nil {
			err := m.marketplaceManager.CompleteContract(context.Background(), contract.ID, m.playerID)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to complete: %v", err)
			} else {
				m.marketplace.message = "Contract completed!"
				m.marketplace.mode = marketplaceModeContracts
			}
		}
	case "q", "esc":
		m.marketplace.mode = marketplaceModeContracts
		m.marketplace.selectedContract = nil
	}

	return m, nil
}

// updateMarketplaceBounties handles bounty browsing
func (m *Model) updateMarketplaceBounties(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.marketplace.selectedIndex > 0 {
			m.marketplace.selectedIndex--
			if m.marketplace.selectedIndex < m.marketplace.viewOffset {
				m.marketplace.viewOffset = m.marketplace.selectedIndex
			}
		}
	case "down", "j":
		if m.marketplace.selectedIndex < len(m.marketplace.bounties)-1 {
			m.marketplace.selectedIndex++
			if m.marketplace.selectedIndex >= m.marketplace.viewOffset+10 {
				m.marketplace.viewOffset++
			}
		}
	case "enter":
		if len(m.marketplace.bounties) > 0 && m.marketplace.selectedIndex < len(m.marketplace.bounties) {
			m.marketplace.selectedBounty = m.marketplace.bounties[m.marketplace.selectedIndex]
			m.marketplace.mode = marketplaceModeViewBounty
		}
	case "b":
		m.marketplace.mode = marketplaceModeMenu
	case "q", "esc":
		m.marketplace.mode = marketplaceModeMenu
	}

	return m, nil
}

// updateMarketplaceViewBounty handles viewing a specific bounty
func (m *Model) updateMarketplaceViewBounty(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.marketplace.mode = marketplaceModeBounties
		m.marketplace.selectedBounty = nil
	}

	return m, nil
}

// updateMarketplaceCreateAuction handles auction creation form
func (m *Model) updateMarketplaceCreateAuction(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If item picker is showing, route input to it
	if m.marketplace.showItemPicker {
		switch msg.String() {
		case "esc":
			// Cancel item selection
			m.marketplace.showItemPicker = false
			return m, nil

		case "enter":
			// Confirm item selection
			selected := m.marketplace.itemPicker.GetSelectedItems()
			if len(selected) == 0 {
				m.marketplace.error = "Please select an item to auction"
				return m, nil
			}
			m.marketplace.selectedItemID = selected[0]
			m.marketplace.showItemPicker = false
			m.marketplace.error = ""
			// Initialize form fields with defaults
			m.marketplace.createForm["starting_bid"] = "1000"
			m.marketplace.createForm["buyout_price"] = "5000"
			m.marketplace.createForm["duration"] = "24" // hours
			m.marketplace.createForm["description"] = ""
			return m, nil

		default:
			// Route to item picker
			return m, m.marketplace.itemPicker.Update(msg)
		}
	}

	// If no item selected yet, show picker
	if m.marketplace.selectedItemID == uuid.Nil {
		m.marketplace.showItemPicker = true
		return m, nil
	}

	// Handle auction form navigation
	switch msg.String() {
	case "esc", "q":
		// Cancel and return to menu
		m.marketplace.mode = marketplaceModeMenu
		return m, nil

	case "tab":
		// Cycle through form fields (0=starting_bid, 1=buyout_price, 2=duration, 3=description)
		m.marketplace.formField = (m.marketplace.formField + 1) % 4

	case "ctrl+s":
		// Submit auction
		return m, m.createAuction()

	case "backspace":
		// Delete character from current field
		switch m.marketplace.formField {
		case 0: // starting_bid
			if len(m.marketplace.createForm["starting_bid"]) > 0 {
				m.marketplace.createForm["starting_bid"] = m.marketplace.createForm["starting_bid"][:len(m.marketplace.createForm["starting_bid"])-1]
			}
		case 1: // buyout_price
			if len(m.marketplace.createForm["buyout_price"]) > 0 {
				m.marketplace.createForm["buyout_price"] = m.marketplace.createForm["buyout_price"][:len(m.marketplace.createForm["buyout_price"])-1]
			}
		case 2: // duration
			if len(m.marketplace.createForm["duration"]) > 0 {
				m.marketplace.createForm["duration"] = m.marketplace.createForm["duration"][:len(m.marketplace.createForm["duration"])-1]
			}
		case 3: // description
			if len(m.marketplace.createForm["description"]) > 0 {
				m.marketplace.createForm["description"] = m.marketplace.createForm["description"][:len(m.marketplace.createForm["description"])-1]
			}
		}

	default:
		// Add character to current field
		if len(msg.String()) == 1 {
			switch m.marketplace.formField {
			case 0: // starting_bid (numbers only)
				if msg.String() >= "0" && msg.String() <= "9" {
					m.marketplace.createForm["starting_bid"] += msg.String()
				}
			case 1: // buyout_price (numbers only)
				if msg.String() >= "0" && msg.String() <= "9" {
					m.marketplace.createForm["buyout_price"] += msg.String()
				}
			case 2: // duration (numbers only)
				if msg.String() >= "0" && msg.String() <= "9" {
					m.marketplace.createForm["duration"] += msg.String()
				}
			case 3: // description
				if len(m.marketplace.createForm["description"]) < 500 {
					m.marketplace.createForm["description"] += msg.String()
				}
			}
		}
	}

	return m, nil
}

// updateMarketplaceCreateContract handles contract creation form
func (m *Model) updateMarketplaceCreateContract(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.marketplace.mode = marketplaceModeMenu
	case "enter":
		// Create contract with marketplace manager
		if m.marketplaceManager != nil {
			// Basic implementation - in a real implementation, this would use form fields
			targetID := uuid.New() // Placeholder - would come from form
			_, err := m.marketplaceManager.CreateContract(
				context.Background(),
				m.playerID,
				m.username,
				marketplace.ContractTypeCourier,
				"Contract Title", // Title
				"Contract description", // Description
				10000, // Reward
				targetID,
				"Target", // Target name
				48*time.Hour, // Duration
			)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to create contract: %v", err)
			} else {
				m.marketplace.message = "Contract posted!"
			}
		}
		m.marketplace.mode = marketplaceModeMenu
	}

	return m, nil
}

// updateMarketplacePostBounty handles bounty posting form
func (m *Model) updateMarketplacePostBounty(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.marketplace.mode = marketplaceModeMenu
	case "enter":
		// Post bounty with marketplace manager
		if m.marketplaceManager != nil {
			// Basic implementation - in a real implementation, this would use form fields
			targetID := uuid.New() // Placeholder - would come from form
			_, err := m.marketplaceManager.PostBounty(
				context.Background(),
				m.playerID,
				m.username,
				targetID,
				"Target Name", // Target name
				5000, // Bounty amount
				"Bounty reason", // Reason
			)
			if err != nil {
				m.marketplace.error = fmt.Sprintf("Failed to post bounty: %v", err)
			} else {
				m.marketplace.message = "Bounty posted!"
			}
		}
		m.marketplace.mode = marketplaceModeMenu
	}

	return m, nil
}

// viewMarketplace renders the marketplace screen
func (m *Model) viewMarketplace() string {
	var b strings.Builder

	// Header
	header := titleStyle.Render("╔═══════════════════════════════════════════════════════════════════════╗")
	title := titleStyle.Render("║                        PLAYER MARKETPLACE                             ║")
	divider := titleStyle.Render("╠═══════════════════════════════════════════════════════════════════════╣")

	b.WriteString(header + "\n")
	b.WriteString(title + "\n")
	b.WriteString(divider + "\n")

	// Player info
	info := fmt.Sprintf("║ Credits: %s%-47d%s ║\n",
		creditStyle.Render(""), m.player.Credits, resetStyle.Render(""))
	b.WriteString(titleStyle.Render(info))
	b.WriteString(titleStyle.Render("╠═══════════════════════════════════════════════════════════════════════╣") + "\n")

	// Content based on mode
	switch m.marketplace.mode {
	case marketplaceModeMenu:
		b.WriteString(m.viewMarketplaceMenu())
	case marketplaceModeAuctions:
		b.WriteString(m.viewMarketplaceAuctions())
	case marketplaceModeViewAuction:
		b.WriteString(m.viewMarketplaceAuctionDetails())
	case marketplaceModeContracts:
		b.WriteString(m.viewMarketplaceContracts())
	case marketplaceModeViewContract:
		b.WriteString(m.viewMarketplaceContractDetails())
	case marketplaceModeBounties:
		b.WriteString(m.viewMarketplaceBounties())
	case marketplaceModeViewBounty:
		b.WriteString(m.viewMarketplaceBountyDetails())
	case marketplaceModeCreateAuction:
		b.WriteString(m.viewMarketplaceCreateAuction())
	case marketplaceModeCreateContract:
		b.WriteString(m.viewMarketplaceCreateContract())
	case marketplaceModePostBounty:
		b.WriteString(m.viewMarketplacePostBounty())
	}

	// Footer
	b.WriteString(titleStyle.Render("╚═══════════════════════════════════════════════════════════════════════╝") + "\n")

	// Help text
	switch m.marketplace.mode {
	case marketplaceModeMenu:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: Select | Q: Back\n"))
	case marketplaceModeAuctions:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: View Auction | B: Back | Q: Menu\n"))
	case marketplaceModeViewAuction:
		b.WriteString(helpStyle.Render("B: Bid | O: Buyout | C: Cancel (own) | Q: Back\n"))
	case marketplaceModeContracts:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: View Contract | B: Back | Q: Menu\n"))
	case marketplaceModeViewContract:
		b.WriteString(helpStyle.Render("C: Claim | Enter: Complete (if claimed) | Q: Back\n"))
	case marketplaceModeBounties:
		b.WriteString(helpStyle.Render("↑/↓: Navigate | Enter: View Bounty | B: Back | Q: Menu\n"))
	default:
		b.WriteString(helpStyle.Render("Q: Cancel\n"))
	}

	// Messages
	if m.marketplace.message != "" {
		b.WriteString("\n" + successStyle.Render(m.marketplace.message) + "\n")
	}
	if m.marketplace.error != "" {
		b.WriteString("\n" + errorStyle.Render(m.marketplace.error) + "\n")
	}

	return b.String()
}

// viewMarketplaceMenu renders the main menu
func (m *Model) viewMarketplaceMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	for i, item := range marketplaceMenuItems {
		if i == m.marketplace.menuIndex {
			line := fmt.Sprintf("║   ► %-65s ║", item)
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			line := fmt.Sprintf("║     %-65s ║", item)
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewMarketplaceAuctions renders auction list
func (m *Model) viewMarketplaceAuctions() string {
	var b strings.Builder

	if len(m.marketplace.auctions) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No auctions available                              ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-30s %-15s %-15s %-10s ║", "Item", "Current Bid", "Buyout", "Time Left")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	// Show 10 items at a time
	end := m.marketplace.viewOffset + 10
	if end > len(m.marketplace.auctions) {
		end = len(m.marketplace.auctions)
	}

	for i := m.marketplace.viewOffset; i < end; i++ {
		auction := m.marketplace.auctions[i]

		timeLeft := time.Until(auction.EndTime).Round(time.Minute)
		timeStr := formatDuration(timeLeft)

		currentBid := "No bids"
		if auction.CurrentBid > 0 {
			currentBid = fmt.Sprintf("%d cr", auction.CurrentBid)
		}

		buyout := "None"
		if auction.BuyoutPrice > 0 {
			buyout = fmt.Sprintf("%d cr", auction.BuyoutPrice)
		}

		line := fmt.Sprintf("║ %-30s %-15s %-15s %-10s ║",
			truncate(auction.ItemName, 30),
			currentBid,
			buyout,
			timeStr)

		if i == m.marketplace.selectedIndex {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	// Pagination info
	if len(m.marketplace.auctions) > 10 {
		pageInfo := fmt.Sprintf("║ Showing %d-%d of %d%53s ║",
			m.marketplace.viewOffset+1,
			end,
			len(m.marketplace.auctions),
			"")
		b.WriteString(titleStyle.Render(pageInfo) + "\n")
	}

	return b.String()
}

// viewMarketplaceAuctionDetails renders detailed auction view
func (m *Model) viewMarketplaceAuctionDetails() string {
	var b strings.Builder

	auction := m.marketplace.selectedAuction
	if auction == nil {
		return ""
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Item: %-63s ║", auction.ItemName)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Type: %-63s ║", auction.Type)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Seller: %-61s ║", auction.SellerName)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	if auction.Description != "" {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Description: %-58s ║", truncate(auction.Description, 58))) + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Starting Bid: %-55d cr ║", auction.StartingBid)) + "\n")

	if auction.CurrentBid > 0 {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Current Bid: %-55d cr ║", auction.CurrentBid)) + "\n")
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ High Bidder: %-56s ║", auction.HighBidderName)) + "\n")
	} else {
		b.WriteString(titleStyle.Render("║ Current Bid: No bids yet                                             ║") + "\n")
	}

	if auction.BuyoutPrice > 0 {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Buyout Price: %-54d cr ║", auction.BuyoutPrice)) + "\n")
	}

	timeLeft := time.Until(auction.EndTime)
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Time Remaining: %-53s ║", formatDuration(timeLeft))) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Status: %-61s ║", auction.Status)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	// Bid history
	if len(auction.BidHistory) > 0 {
		b.WriteString(titleStyle.Render("║ Bid History:                                                          ║") + "\n")
		// Show last 5 bids
		start := 0
		if len(auction.BidHistory) > 5 {
			start = len(auction.BidHistory) - 5
		}
		for i := start; i < len(auction.BidHistory); i++ {
			bid := auction.BidHistory[i]
			bidLine := fmt.Sprintf("║   %s - %d cr", bid.BidderName, bid.Amount)
			b.WriteString(titleStyle.Render(fmt.Sprintf("%-73s ║\n", bidLine)))
		}
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewMarketplaceContracts renders contract list
func (m *Model) viewMarketplaceContracts() string {
	var b strings.Builder

	if len(m.marketplace.contracts) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No contracts available                             ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-30s %-15s %-15s %-10s ║", "Title", "Type", "Reward", "Status")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	// Show 10 items at a time
	end := m.marketplace.viewOffset + 10
	if end > len(m.marketplace.contracts) {
		end = len(m.marketplace.contracts)
	}

	for i := m.marketplace.viewOffset; i < end; i++ {
		contract := m.marketplace.contracts[i]

		line := fmt.Sprintf("║ %-30s %-15s %11d cr %-10s ║",
			truncate(contract.Title, 30),
			contract.Type,
			contract.Reward,
			contract.Status)

		if i == m.marketplace.selectedIndex {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

// viewMarketplaceContractDetails renders detailed contract view
func (m *Model) viewMarketplaceContractDetails() string {
	var b strings.Builder

	contract := m.marketplace.selectedContract
	if contract == nil {
		return ""
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Title: %-63s ║", contract.Title)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Type: %-64s ║", contract.Type)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Poster: %-62s ║", contract.PosterName)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	if contract.Description != "" {
		descLines := wrapText(contract.Description, 65)
		for _, line := range descLines {
			b.WriteString(titleStyle.Render(fmt.Sprintf("║ %-69s ║", line)) + "\n")
		}
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Reward: %-60d cr ║", contract.Reward)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Target: %-62s ║", contract.TargetName)) + "\n")

	timeLeft := time.Until(contract.ExpiryTime)
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Expires In: %-58s ║", formatDuration(timeLeft))) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Status: %-62s ║", contract.Status)) + "\n")

	if contract.Status == "claimed" {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Claimed By: %-58s ║", contract.ClaimedName)) + "\n")
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewMarketplaceBounties renders bounty list
func (m *Model) viewMarketplaceBounties() string {
	var b strings.Builder

	if len(m.marketplace.bounties) == 0 {
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		b.WriteString(titleStyle.Render("║                    No active bounties                                 ║") + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
		return b.String()
	}

	// Table header
	headerLine := fmt.Sprintf("║ %-25s %-20s %-15s %-10s ║", "Target", "Posted By", "Amount", "Expires")
	b.WriteString(titleStyle.Render(headerLine) + "\n")
	b.WriteString(titleStyle.Render("╟───────────────────────────────────────────────────────────────────────╢") + "\n")

	// Show 10 items at a time
	end := m.marketplace.viewOffset + 10
	if end > len(m.marketplace.bounties) {
		end = len(m.marketplace.bounties)
	}

	for i := m.marketplace.viewOffset; i < end; i++ {
		bounty := m.marketplace.bounties[i]

		timeLeft := time.Until(bounty.ExpiryTime).Round(time.Hour)
		timeStr := formatDuration(timeLeft)

		line := fmt.Sprintf("║ %-25s %-20s %11d cr %-10s ║",
			truncate(bounty.TargetName, 25),
			truncate(bounty.PosterName, 20),
			bounty.Amount,
			timeStr)

		if i == m.marketplace.selectedIndex {
			b.WriteString(warningStyle.Render(line) + "\n") // Red for bounties
		} else {
			b.WriteString(titleStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

// viewMarketplaceBountyDetails renders detailed bounty view
func (m *Model) viewMarketplaceBountyDetails() string {
	var b strings.Builder

	bounty := m.marketplace.selectedBounty
	if bounty == nil {
		return ""
	}

	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(warningStyle.Render(fmt.Sprintf("║ WANTED: %-62s ║", bounty.TargetName)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Posted By: %-59s ║", bounty.PosterName)) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Reward: %-60d cr ║", bounty.Amount)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	if bounty.Reason != "" {
		b.WriteString(titleStyle.Render(fmt.Sprintf("║ Reason: %-62s ║", truncate(bounty.Reason, 62))) + "\n")
		b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	}

	timeLeft := time.Until(bounty.ExpiryTime)
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Expires In: %-58s ║", formatDuration(timeLeft))) + "\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("║ Status: %-62s ║", bounty.Status)) + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// Commands

// createAuction creates a new auction with the marketplace manager
func (m *Model) createAuction() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Parse form values
		startingBid := int64(0)
		if bid := m.marketplace.createForm["starting_bid"]; bid != "" {
			fmt.Sscanf(bid, "%d", &startingBid)
		}

		buyoutPrice := int64(0)
		if buyout := m.marketplace.createForm["buyout_price"]; buyout != "" {
			fmt.Sscanf(buyout, "%d", &buyoutPrice)
		}

		durationHours := 24
		if dur := m.marketplace.createForm["duration"]; dur != "" {
			fmt.Sscanf(dur, "%d", &durationHours)
		}

		description := m.marketplace.createForm["description"]
		if description == "" {
			description = "No description provided"
		}

		// Get item details
		item, err := m.itemRepo.GetItemByID(ctx, m.marketplace.selectedItemID)
		if err != nil {
			return marketplaceAuctionCreatedMsg{err: fmt.Sprintf("Failed to get item: %v", err)}
		}

		// Validate starting bid
		if startingBid < 100 {
			return marketplaceAuctionCreatedMsg{err: "Starting bid must be at least 100 credits"}
		}

		// Validate buyout price
		if buyoutPrice > 0 && buyoutPrice < startingBid {
			return marketplaceAuctionCreatedMsg{err: "Buyout price must be higher than starting bid"}
		}

		// Validate duration
		if durationHours < 1 || durationHours > 168 {
			return marketplaceAuctionCreatedMsg{err: "Duration must be between 1 and 168 hours (1 week)"}
		}

		// Determine auction type from item type
		var auctionType marketplace.AuctionType
		switch item.ItemType {
		case "weapon":
			auctionType = marketplace.AuctionTypeOutfit // Weapons count as outfits
		case "outfit":
			auctionType = marketplace.AuctionTypeOutfit
		case "special", "quest":
			auctionType = marketplace.AuctionTypeSpecial
		default:
			auctionType = marketplace.AuctionTypeSpecial
		}

		// Create auction
		_, err = m.marketplaceManager.CreateAuction(
			ctx,
			m.playerID,
			m.username,
			auctionType,
			m.marketplace.selectedItemID,
			item.GetDisplayName(),
			1, // Quantity (inventory items are unique)
			description,
			startingBid,
			time.Duration(durationHours)*time.Hour,
			buyoutPrice,
		)

		if err != nil {
			return marketplaceAuctionCreatedMsg{err: fmt.Sprintf("Failed to create auction: %v", err)}
		}

		return marketplaceAuctionCreatedMsg{err: ""}
	}
}

type marketplaceAuctionCreatedMsg struct {
	err string
}

// View functions

// viewMarketplaceCreateAuction renders auction creation form
func (m *Model) viewMarketplaceCreateAuction() string {
	var b strings.Builder

	// If item picker is showing, render it
	if m.marketplace.showItemPicker {
		b.WriteString(titleStyle.Render("CREATE AUCTION - Select Item") + "\n\n")
		b.WriteString(m.marketplace.itemPicker.View())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[Space] Toggle  [Enter] Confirm  [Esc] Cancel") + "\n")
		return b.String()
	}

	// If no item selected, should show picker (this shouldn't happen but handle it)
	if m.marketplace.selectedItemID == uuid.Nil {
		b.WriteString(errorStyle.Render("No item selected") + "\n")
		b.WriteString(helpStyle.Render("Press any key to select an item") + "\n")
		return b.String()
	}

	// Show auction form
	b.WriteString(titleStyle.Render("CREATE AUCTION") + "\n\n")

	// Get selected item details
	ctx := context.Background()
	item, err := m.itemRepo.GetItemByID(ctx, m.marketplace.selectedItemID)
	itemName := "Unknown Item"
	if err == nil {
		itemName = item.GetDisplayName()
	}

	b.WriteString(subtitleStyle.Render(fmt.Sprintf("Item: %s", itemName)) + "\n\n")

	// Form fields with highlighting
	startingBidStyle := lipgloss.NewStyle()
	if m.marketplace.formField == 0 {
		startingBidStyle = startingBidStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	b.WriteString(startingBidStyle.Render(fmt.Sprintf("Starting Bid: %s_ credits", m.marketplace.createForm["starting_bid"])) + "\n")

	buyoutStyle := lipgloss.NewStyle()
	if m.marketplace.formField == 1 {
		buyoutStyle = buyoutStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	b.WriteString(buyoutStyle.Render(fmt.Sprintf("Buyout Price: %s_ credits (0 for none)", m.marketplace.createForm["buyout_price"])) + "\n")

	durationStyle := lipgloss.NewStyle()
	if m.marketplace.formField == 2 {
		durationStyle = durationStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	b.WriteString(durationStyle.Render(fmt.Sprintf("Duration: %s_ hours (1-168)", m.marketplace.createForm["duration"])) + "\n\n")

	descStyle := lipgloss.NewStyle()
	if m.marketplace.formField == 3 {
		descStyle = descStyle.Foreground(lipgloss.Color("11")).Bold(true)
	}
	b.WriteString(descStyle.Render("Description:\n"))
	descLines := wrapText(m.marketplace.createForm["description"]+"_", 70)
	for _, line := range descLines {
		b.WriteString(descStyle.Render(line) + "\n")
	}

	b.WriteString("\n")

	// Show error if any
	if m.marketplace.error != "" {
		b.WriteString(errorStyle.Render(m.marketplace.error) + "\n\n")
	}

	// Help text
	b.WriteString(helpStyle.Render("[Tab] Next Field  [Ctrl+S] Create Auction  [Esc] Cancel") + "\n")

	return b.String()
}

// viewMarketplaceCreateContract renders contract creation form
func (m *Model) viewMarketplaceCreateContract() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                   POST CONTRACT                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ Contract creation form (Press Enter to create with defaults)          ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

// viewMarketplacePostBounty renders bounty posting form
func (m *Model) viewMarketplacePostBounty() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("║                   POST BOUNTY                                         ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")
	b.WriteString(titleStyle.Render("║ Bounty posting form (Press Enter to post with defaults)               ║") + "\n")
	b.WriteString(titleStyle.Render("║                                                                       ║") + "\n")

	return b.String()
}

var (
	creditStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
)
