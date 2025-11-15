// File: internal/marketplace/manager.go
// Project: Terminal Velocity
// Description: Player marketplace manager for auctions, contracts, and bounties
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package marketplace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Marketplace")

// Manager handles all marketplace operations including auctions, contracts, and bounties
type Manager struct {
	mu sync.RWMutex

	// Active listings
	auctions  map[uuid.UUID]*Auction
	contracts map[uuid.UUID]*Contract
	bounties  map[uuid.UUID]*Bounty

	// Configuration
	config MarketplaceConfig

	// Repositories
	playerRepo *database.PlayerRepository
	shipRepo   *database.ShipRepository

	// Callbacks
	onAuctionComplete func(auction *Auction)
	onContractClaimed func(contract *Contract)
	onBountyClaimed   func(bounty *Bounty)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// MarketplaceConfig defines marketplace parameters
type MarketplaceConfig struct {
	// Auction settings
	MinAuctionDuration    time.Duration // Minimum auction duration
	MaxAuctionDuration    time.Duration // Maximum auction duration
	AuctionFeePercent     float64       // Fee taken from final sale (0.0 - 1.0)
	MinimumBidIncrement   float64       // Minimum bid increment percentage
	BuyoutPremium         float64       // Premium for instant buyout (e.g., 1.5 = 150% of starting bid)

	// Contract settings
	ContractPostCost      int64         // Cost to post a contract
	ContractExpiryTime    time.Duration // How long until contracts expire
	MaxActiveContracts    int           // Max contracts per player
	ContractFailurePenalty float64      // Credits penalty for failing contract (0.0 - 1.0)

	// Bounty settings
	MinBountyAmount       int64         // Minimum bounty amount
	BountyPostFee         float64       // Fee to post bounty (0.0 - 1.0 of bounty amount)
	BountyExpiryTime      time.Duration // How long until bounties expire
	MaxBountiesPerPlayer  int           // Max bounties one player can have on their head
	BountyClaimWindow     time.Duration // Window to claim bounty after kill
}

// DefaultMarketplaceConfig returns sensible defaults
func DefaultMarketplaceConfig() MarketplaceConfig {
	return MarketplaceConfig{
		MinAuctionDuration:     1 * time.Hour,
		MaxAuctionDuration:     7 * 24 * time.Hour, // 7 days
		AuctionFeePercent:      0.05,               // 5% fee
		MinimumBidIncrement:    0.05,               // 5% increment
		BuyoutPremium:          1.5,                // 150% for instant buyout
		ContractPostCost:       1000,
		ContractExpiryTime:     48 * time.Hour,
		MaxActiveContracts:     3,
		ContractFailurePenalty: 0.10, // 10% penalty
		MinBountyAmount:        5000,
		BountyPostFee:          0.10, // 10% fee
		BountyExpiryTime:       72 * time.Hour,
		MaxBountiesPerPlayer:   5,
		BountyClaimWindow:      5 * time.Minute,
	}
}

// NewManager creates a new marketplace manager
func NewManager(playerRepo *database.PlayerRepository, shipRepo *database.ShipRepository) *Manager {
	return &Manager{
		auctions:  make(map[uuid.UUID]*Auction),
		contracts: make(map[uuid.UUID]*Contract),
		bounties:  make(map[uuid.UUID]*Bounty),
		config:    DefaultMarketplaceConfig(),
		playerRepo: playerRepo,
		shipRepo:   shipRepo,
		stopChan:  make(chan struct{}),
	}
}

// Start begins background workers for marketplace
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.expiryWorker()
	log.Info("Marketplace manager started")
}

// Stop gracefully shuts down the marketplace
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Marketplace manager stopped")
}

// SetAuctionCompleteCallback sets callback for auction completion
func (m *Manager) SetAuctionCompleteCallback(callback func(auction *Auction)) {
	m.onAuctionComplete = callback
}

// SetContractClaimedCallback sets callback for contract claims
func (m *Manager) SetContractClaimedCallback(callback func(contract *Contract)) {
	m.onContractClaimed = callback
}

// SetBountyClaimedCallback sets callback for bounty claims
func (m *Manager) SetBountyClaimedCallback(callback func(bounty *Bounty)) {
	m.onBountyClaimed = callback
}

// ============================================================================
// AUCTION SYSTEM
// ============================================================================

// AuctionType represents the type of item being auctioned
type AuctionType string

const (
	AuctionTypeShip      AuctionType = "ship"
	AuctionTypeOutfit    AuctionType = "outfit"
	AuctionTypeCommodity AuctionType = "commodity"
	AuctionTypeSpecial   AuctionType = "special" // Rare items, blueprints, etc.
)

// Auction represents an active auction listing
type Auction struct {
	ID          uuid.UUID
	SellerID    uuid.UUID
	SellerName  string
	Type        AuctionType
	ItemID      uuid.UUID // Ship ID, outfit ID, etc.
	ItemName    string
	Quantity    int       // For commodities/stackable items
	Description string
	StartingBid int64
	BuyoutPrice int64     // Instant purchase price (optional)
	CurrentBid  int64
	HighBidder  uuid.UUID // Player who has current high bid
	HighBidderName string
	StartTime   time.Time
	EndTime     time.Time
	Status      string    // "active", "sold", "expired", "cancelled"
	BidHistory  []Bid
}

// Bid represents a bid on an auction
type Bid struct {
	BidderID   uuid.UUID
	BidderName string
	Amount     int64
	Timestamp  time.Time
}

// CreateAuction creates a new auction listing
func (m *Manager) CreateAuction(ctx context.Context, sellerID uuid.UUID, sellerName string, auctionType AuctionType, itemID uuid.UUID, itemName string, quantity int, description string, startingBid int64, duration time.Duration, buyoutPrice int64) (*Auction, error) {
	// Validate duration
	if duration < m.config.MinAuctionDuration || duration > m.config.MaxAuctionDuration {
		return nil, fmt.Errorf("auction duration must be between %v and %v", m.config.MinAuctionDuration, m.config.MaxAuctionDuration)
	}

	// Validate prices
	if startingBid <= 0 {
		return nil, fmt.Errorf("starting bid must be positive")
	}
	if buyoutPrice > 0 && buyoutPrice <= startingBid {
		return nil, fmt.Errorf("buyout price must be higher than starting bid")
	}

	auction := &Auction{
		ID:          uuid.New(),
		SellerID:    sellerID,
		SellerName:  sellerName,
		Type:        auctionType,
		ItemID:      itemID,
		ItemName:    itemName,
		Quantity:    quantity,
		Description: description,
		StartingBid: startingBid,
		BuyoutPrice: buyoutPrice,
		CurrentBid:  0,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(duration),
		Status:      "active",
		BidHistory:  []Bid{},
	}

	m.mu.Lock()
	m.auctions[auction.ID] = auction
	m.mu.Unlock()

	log.Info("Auction created: seller=%s, item=%s, duration=%v", sellerName, itemName, duration)
	return auction, nil
}

// PlaceBid places a bid on an auction
func (m *Manager) PlaceBid(ctx context.Context, auctionID uuid.UUID, bidderID uuid.UUID, bidderName string, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auction, exists := m.auctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	if time.Now().After(auction.EndTime) {
		return fmt.Errorf("auction has ended")
	}

	if bidderID == auction.SellerID {
		return fmt.Errorf("cannot bid on your own auction")
	}

	// Calculate minimum bid
	minBid := auction.StartingBid
	if auction.CurrentBid > 0 {
		minBid = int64(float64(auction.CurrentBid) * (1.0 + m.config.MinimumBidIncrement))
	}

	if amount < minBid {
		return fmt.Errorf("bid must be at least %d credits", minBid)
	}

	// Check if bidder has sufficient credits
	bidder, err := m.playerRepo.GetByID(ctx, bidderID)
	if err != nil {
		return fmt.Errorf("failed to get bidder: %v", err)
	}
	if bidder.Credits < amount {
		return fmt.Errorf("insufficient credits")
	}

	// Refund previous high bidder
	if auction.HighBidder != uuid.Nil {
		previousBidder, err := m.playerRepo.GetByID(ctx, auction.HighBidder)
		if err == nil {
			previousBidder.Credits += auction.CurrentBid
			_ = m.playerRepo.Update(ctx, previousBidder)
		}
	}

	// Deduct credits from new bidder
	bidder.Credits -= amount
	if err := m.playerRepo.Update(ctx, bidder); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Update auction
	auction.CurrentBid = amount
	auction.HighBidder = bidderID
	auction.HighBidderName = bidderName
	auction.BidHistory = append(auction.BidHistory, Bid{
		BidderID:   bidderID,
		BidderName: bidderName,
		Amount:     amount,
		Timestamp:  time.Now(),
	})

	log.Info("Bid placed: auction=%s, bidder=%s, amount=%d", auction.ItemName, bidderName, amount)
	return nil
}

// Buyout instantly purchases an auction at buyout price
func (m *Manager) Buyout(ctx context.Context, auctionID uuid.UUID, buyerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auction, exists := m.auctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	if auction.BuyoutPrice <= 0 {
		return fmt.Errorf("auction does not have buyout option")
	}

	// Check buyer has sufficient credits
	buyer, err := m.playerRepo.GetByID(ctx, buyerID)
	if err != nil {
		return fmt.Errorf("failed to get buyer: %v", err)
	}
	if buyer.Credits < auction.BuyoutPrice {
		return fmt.Errorf("insufficient credits for buyout")
	}

	// Refund previous high bidder if any
	if auction.HighBidder != uuid.Nil {
		previousBidder, err := m.playerRepo.GetByID(ctx, auction.HighBidder)
		if err == nil {
			previousBidder.Credits += auction.CurrentBid
			_ = m.playerRepo.Update(ctx, previousBidder)
		}
	}

	// Deduct buyout price from buyer
	buyer.Credits -= auction.BuyoutPrice
	if err := m.playerRepo.Update(ctx, buyer); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Pay seller (minus fee)
	seller, err := m.playerRepo.GetByID(ctx, auction.SellerID)
	if err == nil {
		fee := int64(float64(auction.BuyoutPrice) * m.config.AuctionFeePercent)
		seller.Credits += (auction.BuyoutPrice - fee)
		_ = m.playerRepo.Update(ctx, seller)
	}

	// Complete auction
	auction.Status = "sold"
	auction.CurrentBid = auction.BuyoutPrice
	auction.HighBidder = buyerID

	log.Info("Auction buyout: auction=%s, buyer=%s, price=%d", auction.ItemName, buyer.Username, auction.BuyoutPrice)

	// Trigger callback
	if m.onAuctionComplete != nil {
		go m.onAuctionComplete(auction)
	}

	return nil
}

// CancelAuction cancels an active auction (only if no bids)
func (m *Manager) CancelAuction(ctx context.Context, auctionID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auction, exists := m.auctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.SellerID != playerID {
		return fmt.Errorf("only the seller can cancel this auction")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	if auction.CurrentBid > 0 {
		return fmt.Errorf("cannot cancel auction with active bids")
	}

	auction.Status = "cancelled"
	log.Info("Auction cancelled: auction=%s, seller=%s", auction.ItemName, auction.SellerName)
	return nil
}

// GetActiveAuctions returns all active auctions
func (m *Manager) GetActiveAuctions() []*Auction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := []*Auction{}
	for _, auction := range m.auctions {
		if auction.Status == "active" && time.Now().Before(auction.EndTime) {
			active = append(active, auction)
		}
	}
	return active
}

// GetAuction retrieves a specific auction
func (m *Manager) GetAuction(auctionID uuid.UUID) (*Auction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	auction, exists := m.auctions[auctionID]
	return auction, exists
}

// ============================================================================
// CONTRACT SYSTEM
// ============================================================================

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeCourier      ContractType = "courier"      // Deliver cargo
	ContractTypeAssassination ContractType = "assassination" // Kill a target
	ContractTypeEscort       ContractType = "escort"       // Escort a ship
	ContractTypeBountyHunt   ContractType = "bounty_hunt"  // Hunt a bounty target
)

// Contract represents a player-posted contract
type Contract struct {
	ID          uuid.UUID
	PosterID    uuid.UUID
	PosterName  string
	Type        ContractType
	Title       string
	Description string
	Reward      int64
	Deposit     int64 // Amount poster must deposit
	TargetID    uuid.UUID // Target player/system/location
	TargetName  string
	ClaimedBy   uuid.UUID
	ClaimedName string
	PostTime    time.Time
	ExpiryTime  time.Time
	ClaimTime   time.Time
	CompleteTime time.Time
	Status      string // "open", "claimed", "completed", "failed", "expired"
}

// CreateContract posts a new contract
func (m *Manager) CreateContract(ctx context.Context, posterID uuid.UUID, posterName string, contractType ContractType, title string, description string, reward int64, targetID uuid.UUID, targetName string, duration time.Duration) (*Contract, error) {
	// Check active contract limit
	m.mu.RLock()
	activeCount := 0
	for _, c := range m.contracts {
		if c.PosterID == posterID && (c.Status == "open" || c.Status == "claimed") {
			activeCount++
		}
	}
	m.mu.RUnlock()

	if activeCount >= m.config.MaxActiveContracts {
		return nil, fmt.Errorf("maximum active contracts reached (%d)", m.config.MaxActiveContracts)
	}

	// Validate reward
	if reward <= 0 {
		return nil, fmt.Errorf("reward must be positive")
	}

	// Calculate deposit (reward + posting cost)
	deposit := reward + m.config.ContractPostCost

	// Check poster has sufficient credits
	poster, err := m.playerRepo.GetByID(ctx, posterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get poster: %v", err)
	}
	if poster.Credits < deposit {
		return nil, fmt.Errorf("insufficient credits (need %d)", deposit)
	}

	// Deduct deposit
	poster.Credits -= deposit
	if err := m.playerRepo.Update(ctx, poster); err != nil {
		return nil, fmt.Errorf("failed to deduct deposit: %v", err)
	}

	contract := &Contract{
		ID:          uuid.New(),
		PosterID:    posterID,
		PosterName:  posterName,
		Type:        contractType,
		Title:       title,
		Description: description,
		Reward:      reward,
		Deposit:     deposit,
		TargetID:    targetID,
		TargetName:  targetName,
		PostTime:    time.Now(),
		ExpiryTime:  time.Now().Add(duration),
		Status:      "open",
	}

	m.mu.Lock()
	m.contracts[contract.ID] = contract
	m.mu.Unlock()

	log.Info("Contract created: poster=%s, type=%s, reward=%d", posterName, contractType, reward)
	return contract, nil
}

// ClaimContract claims an open contract
func (m *Manager) ClaimContract(ctx context.Context, contractID uuid.UUID, claimerID uuid.UUID, claimerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	contract, exists := m.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract not found")
	}

	if contract.Status != "open" {
		return fmt.Errorf("contract is not available")
	}

	if time.Now().After(contract.ExpiryTime) {
		return fmt.Errorf("contract has expired")
	}

	if claimerID == contract.PosterID {
		return fmt.Errorf("cannot claim your own contract")
	}

	contract.Status = "claimed"
	contract.ClaimedBy = claimerID
	contract.ClaimedName = claimerName
	contract.ClaimTime = time.Now()

	log.Info("Contract claimed: contract=%s, claimer=%s", contract.Title, claimerName)

	if m.onContractClaimed != nil {
		go m.onContractClaimed(contract)
	}

	return nil
}

// CompleteContract marks a contract as completed
func (m *Manager) CompleteContract(ctx context.Context, contractID uuid.UUID, completerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	contract, exists := m.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract not found")
	}

	if contract.Status != "claimed" {
		return fmt.Errorf("contract is not claimed")
	}

	if contract.ClaimedBy != completerID {
		return fmt.Errorf("only the claimer can complete this contract")
	}

	// Pay reward to completer
	completer, err := m.playerRepo.GetByID(ctx, completerID)
	if err != nil {
		return fmt.Errorf("failed to get completer: %v", err)
	}
	completer.Credits += contract.Reward
	if err := m.playerRepo.Update(ctx, completer); err != nil {
		return fmt.Errorf("failed to pay reward: %v", err)
	}

	contract.Status = "completed"
	contract.CompleteTime = time.Now()

	log.Info("Contract completed: contract=%s, completer=%s, reward=%d", contract.Title, completer.Username, contract.Reward)
	return nil
}

// FailContract marks a contract as failed
func (m *Manager) FailContract(ctx context.Context, contractID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	contract, exists := m.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract not found")
	}

	if contract.Status != "claimed" {
		return fmt.Errorf("contract is not claimed")
	}

	// Apply penalty to claimer
	if contract.ClaimedBy != uuid.Nil {
		claimer, err := m.playerRepo.GetByID(ctx, contract.ClaimedBy)
		if err == nil {
			penalty := int64(float64(contract.Reward) * m.config.ContractFailurePenalty)
			if claimer.Credits >= penalty {
				claimer.Credits -= penalty
				_ = m.playerRepo.Update(ctx, claimer)
			}
		}
	}

	// Refund poster
	poster, err := m.playerRepo.GetByID(ctx, contract.PosterID)
	if err == nil {
		poster.Credits += contract.Deposit
		_ = m.playerRepo.Update(ctx, poster)
	}

	contract.Status = "failed"
	log.Info("Contract failed: contract=%s, claimer=%s", contract.Title, contract.ClaimedName)
	return nil
}

// GetOpenContracts returns all open contracts
func (m *Manager) GetOpenContracts() []*Contract {
	m.mu.RLock()
	defer m.mu.RUnlock()

	open := []*Contract{}
	for _, contract := range m.contracts {
		if contract.Status == "open" && time.Now().Before(contract.ExpiryTime) {
			open = append(open, contract)
		}
	}
	return open
}

// ============================================================================
// BOUNTY SYSTEM
// ============================================================================

// Bounty represents a bounty on a player's head
type Bounty struct {
	ID          uuid.UUID
	PosterID    uuid.UUID
	PosterName  string
	TargetID    uuid.UUID
	TargetName  string
	Amount      int64
	Reason      string
	PostTime    time.Time
	ExpiryTime  time.Time
	ClaimedBy   uuid.UUID
	ClaimedName string
	ClaimTime   time.Time
	Status      string // "active", "claimed", "expired"
}

// PostBounty posts a bounty on a player
func (m *Manager) PostBounty(ctx context.Context, posterID uuid.UUID, posterName string, targetID uuid.UUID, targetName string, amount int64, reason string) (*Bounty, error) {
	// Validate amount
	if amount < m.config.MinBountyAmount {
		return nil, fmt.Errorf("bounty must be at least %d credits", m.config.MinBountyAmount)
	}

	// Check target bounty limit
	m.mu.RLock()
	targetBountyCount := 0
	for _, b := range m.bounties {
		if b.TargetID == targetID && b.Status == "active" {
			targetBountyCount++
		}
	}
	m.mu.RUnlock()

	if targetBountyCount >= m.config.MaxBountiesPerPlayer {
		return nil, fmt.Errorf("target already has maximum bounties")
	}

	// Calculate total cost (amount + fee)
	fee := int64(float64(amount) * m.config.BountyPostFee)
	totalCost := amount + fee

	// Check poster has sufficient credits
	poster, err := m.playerRepo.GetByID(ctx, posterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get poster: %v", err)
	}
	if poster.Credits < totalCost {
		return nil, fmt.Errorf("insufficient credits (need %d)", totalCost)
	}

	// Deduct total cost
	poster.Credits -= totalCost
	if err := m.playerRepo.Update(ctx, poster); err != nil {
		return nil, fmt.Errorf("failed to deduct credits: %v", err)
	}

	bounty := &Bounty{
		ID:         uuid.New(),
		PosterID:   posterID,
		PosterName: posterName,
		TargetID:   targetID,
		TargetName: targetName,
		Amount:     amount,
		Reason:     reason,
		PostTime:   time.Now(),
		ExpiryTime: time.Now().Add(m.config.BountyExpiryTime),
		Status:     "active",
	}

	m.mu.Lock()
	m.bounties[bounty.ID] = bounty
	m.mu.Unlock()

	log.Info("Bounty posted: poster=%s, target=%s, amount=%d", posterName, targetName, amount)
	return bounty, nil
}

// ClaimBounty claims a bounty after killing the target
func (m *Manager) ClaimBounty(ctx context.Context, targetID uuid.UUID, killerID uuid.UUID, killerName string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalPayout := int64(0)

	// Find all active bounties on target
	for _, bounty := range m.bounties {
		if bounty.TargetID == targetID && bounty.Status == "active" {
			// Pay bounty to killer
			killer, err := m.playerRepo.GetByID(ctx, killerID)
			if err == nil {
				killer.Credits += bounty.Amount
				_ = m.playerRepo.Update(ctx, killer)
				totalPayout += bounty.Amount
			}

			bounty.Status = "claimed"
			bounty.ClaimedBy = killerID
			bounty.ClaimedName = killerName
			bounty.ClaimTime = time.Now()

			log.Info("Bounty claimed: target=%s, killer=%s, amount=%d", bounty.TargetName, killerName, bounty.Amount)

			if m.onBountyClaimed != nil {
				go m.onBountyClaimed(bounty)
			}
		}
	}

	if totalPayout == 0 {
		return 0, fmt.Errorf("no active bounties on target")
	}

	return totalPayout, nil
}

// GetActiveBounties returns all active bounties
func (m *Manager) GetActiveBounties() []*Bounty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := []*Bounty{}
	for _, bounty := range m.bounties {
		if bounty.Status == "active" && time.Now().Before(bounty.ExpiryTime) {
			active = append(active, bounty)
		}
	}
	return active
}

// GetPlayerBounties returns all bounties on a specific player
func (m *Manager) GetPlayerBounties(playerID uuid.UUID) []*Bounty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	playerBounties := []*Bounty{}
	for _, bounty := range m.bounties {
		if bounty.TargetID == playerID && bounty.Status == "active" && time.Now().Before(bounty.ExpiryTime) {
			playerBounties = append(playerBounties, bounty)
		}
	}
	return playerBounties
}

// ============================================================================
// BACKGROUND WORKERS
// ============================================================================

// expiryWorker handles expiration of auctions, contracts, and bounties
func (m *Manager) expiryWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processExpiries()
		case <-m.stopChan:
			return
		}
	}
}

// processExpiries checks and processes all expiries
func (m *Manager) processExpiries() {
	ctx := context.Background()
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Expire auctions
	for _, auction := range m.auctions {
		if auction.Status == "active" && now.After(auction.EndTime) {
			if auction.HighBidder != uuid.Nil {
				// Auction sold - pay seller and complete
				seller, err := m.playerRepo.GetByID(ctx, auction.SellerID)
				if err == nil {
					fee := int64(float64(auction.CurrentBid) * m.config.AuctionFeePercent)
					seller.Credits += (auction.CurrentBid - fee)
					_ = m.playerRepo.Update(ctx, seller)
				}
				auction.Status = "sold"
				log.Info("Auction completed: item=%s, winner=%s, price=%d", auction.ItemName, auction.HighBidderName, auction.CurrentBid)

				if m.onAuctionComplete != nil {
					go m.onAuctionComplete(auction)
				}
			} else {
				// No bids - expire
				auction.Status = "expired"
				log.Info("Auction expired: item=%s (no bids)", auction.ItemName)
			}
		}
	}

	// Expire contracts
	for _, contract := range m.contracts {
		if contract.Status == "open" && now.After(contract.ExpiryTime) {
			// Refund poster
			poster, err := m.playerRepo.GetByID(ctx, contract.PosterID)
			if err == nil {
				poster.Credits += contract.Deposit
				_ = m.playerRepo.Update(ctx, poster)
			}
			contract.Status = "expired"
			log.Info("Contract expired: title=%s", contract.Title)
		}
	}

	// Expire bounties
	for _, bounty := range m.bounties {
		if bounty.Status == "active" && now.After(bounty.ExpiryTime) {
			// Refund poster
			poster, err := m.playerRepo.GetByID(ctx, bounty.PosterID)
			if err == nil {
				poster.Credits += bounty.Amount
				_ = m.playerRepo.Update(ctx, poster)
			}
			bounty.Status = "expired"
			log.Info("Bounty expired: target=%s", bounty.TargetName)
		}
	}
}

// GetStats returns marketplace statistics
func (m *Manager) GetStats() MarketplaceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := MarketplaceStats{}

	// Count auctions
	for _, auction := range m.auctions {
		if auction.Status == "active" {
			stats.ActiveAuctions++
		}
	}

	// Count contracts
	for _, contract := range m.contracts {
		if contract.Status == "open" {
			stats.OpenContracts++
		} else if contract.Status == "claimed" {
			stats.ClaimedContracts++
		}
	}

	// Count bounties
	for _, bounty := range m.bounties {
		if bounty.Status == "active" {
			stats.ActiveBounties++
			stats.TotalBountyPool += bounty.Amount
		}
	}

	return stats
}

// MarketplaceStats contains marketplace statistics
type MarketplaceStats struct {
	ActiveAuctions    int   `json:"active_auctions"`
	OpenContracts     int   `json:"open_contracts"`
	ClaimedContracts  int   `json:"claimed_contracts"`
	ActiveBounties    int   `json:"active_bounties"`
	TotalBountyPool   int64 `json:"total_bounty_pool"`
}
