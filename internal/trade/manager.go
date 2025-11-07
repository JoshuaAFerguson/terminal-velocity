// File: internal/trade/manager.go
// Project: Terminal Velocity
// Version: 1.0.0

package trade

import (
	"errors"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var (
	ErrTradeNotFound      = errors.New("trade offer not found")
	ErrNotAuthorized      = errors.New("not authorized for this trade")
	ErrTradeExpired       = errors.New("trade offer has expired")
	ErrInvalidStatus      = errors.New("trade has invalid status for this operation")
	ErrInsufficientFunds  = errors.New("insufficient funds for trade")
	ErrInsufficientCargo  = errors.New("insufficient cargo for trade")
	ErrLocationMismatch   = errors.New("players must be in same location")
	ErrCannotCancel       = errors.New("cannot cancel this trade")
)

// Manager handles all trade operations
type Manager struct {
	mu       sync.RWMutex
	offers   map[uuid.UUID]*models.TradeOffer          // Trade ID -> Offer
	byPlayer map[uuid.UUID][]*models.TradeOffer        // Player ID -> Offers (sent or received)
	escrow   map[uuid.UUID]*models.TradeEscrow         // Trade ID -> Escrow
	history  map[uuid.UUID]*models.TradeHistory        // Player ID -> History
}

// NewManager creates a new trade manager
func NewManager() *Manager {
	return &Manager{
		offers:   make(map[uuid.UUID]*models.TradeOffer),
		byPlayer: make(map[uuid.UUID][]*models.TradeOffer),
		escrow:   make(map[uuid.UUID]*models.TradeEscrow),
		history:  make(map[uuid.UUID]*models.TradeHistory),
	}
}

// CreateOffer creates a new trade offer
func (m *Manager) CreateOffer(
	initiatorID uuid.UUID,
	initiatorName string,
	recipientID uuid.UUID,
	recipientName string,
	systemID uuid.UUID,
	planetID uuid.UUID,
) *models.TradeOffer {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer := models.NewTradeOffer(
		initiatorID,
		initiatorName,
		recipientID,
		recipientName,
		systemID,
		planetID,
	)

	m.offers[offer.ID] = offer
	m.byPlayer[initiatorID] = append(m.byPlayer[initiatorID], offer)
	m.byPlayer[recipientID] = append(m.byPlayer[recipientID], offer)

	return offer
}

// GetOffer retrieves a trade offer by ID
func (m *Manager) GetOffer(tradeID uuid.UUID) (*models.TradeOffer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offer, exists := m.offers[tradeID]
	if !exists {
		return nil, ErrTradeNotFound
	}

	// Auto-expire if needed
	if offer.IsExpired() {
		offer.MarkExpired()
	}

	return offer, nil
}

// GetPlayerOffers returns all offers for a player (sent or received)
func (m *Manager) GetPlayerOffers(playerID uuid.UUID) []*models.TradeOffer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offers := m.byPlayer[playerID]

	// Auto-expire stale offers
	for _, offer := range offers {
		if offer.IsExpired() {
			offer.MarkExpired()
		}
	}

	return offers
}

// GetPendingOffers returns all pending offers for a player
func (m *Manager) GetPendingOffers(playerID uuid.UUID) []*models.TradeOffer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pending := []*models.TradeOffer{}
	for _, offer := range m.byPlayer[playerID] {
		if offer.IsExpired() {
			offer.MarkExpired()
		}
		if offer.Status == models.TradeStatusPending && offer.RecipientID == playerID {
			pending = append(pending, offer)
		}
	}

	return pending
}

// GetSentOffers returns all offers sent by a player
func (m *Manager) GetSentOffers(playerID uuid.UUID) []*models.TradeOffer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sent := []*models.TradeOffer{}
	for _, offer := range m.byPlayer[playerID] {
		if offer.IsExpired() {
			offer.MarkExpired()
		}
		if offer.InitiatorID == playerID {
			sent = append(sent, offer)
		}
	}

	return sent
}

// AcceptOffer accepts a trade offer and creates escrow
func (m *Manager) AcceptOffer(tradeID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, exists := m.offers[tradeID]
	if !exists {
		return ErrTradeNotFound
	}

	// Only recipient can accept
	if offer.RecipientID != playerID {
		return ErrNotAuthorized
	}

	// Check status
	if offer.Status != models.TradeStatusPending {
		return ErrInvalidStatus
	}

	// Check expiry
	if offer.IsExpired() {
		offer.MarkExpired()
		return ErrTradeExpired
	}

	// Accept the offer
	offer.Accept()

	// Create escrow
	escrow := models.NewTradeEscrow(tradeID)

	// Lock initiator's assets
	initiatorItems := make(map[string]int)
	for _, item := range offer.OfferedItems {
		initiatorItems[item.CommodityName] = item.Quantity
	}
	escrow.LockInitiatorAssets(offer.OfferedCredits, initiatorItems)

	// Lock recipient's assets
	recipientItems := make(map[string]int)
	for _, item := range offer.RequestedItems {
		recipientItems[item.CommodityName] = item.Quantity
	}
	escrow.LockRecipientAssets(offer.RequestedCredits, recipientItems)

	m.escrow[tradeID] = escrow

	return nil
}

// RejectOffer rejects a trade offer
func (m *Manager) RejectOffer(tradeID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, exists := m.offers[tradeID]
	if !exists {
		return ErrTradeNotFound
	}

	// Only recipient can reject
	if offer.RecipientID != playerID {
		return ErrNotAuthorized
	}

	// Check status
	if offer.Status != models.TradeStatusPending {
		return ErrInvalidStatus
	}

	offer.Reject()

	// Update history
	m.ensureHistory(offer.InitiatorID)
	m.history[offer.InitiatorID].RecordTrade(offer.GetTotalOfferedValue(), false)

	return nil
}

// CancelOffer cancels a trade offer
func (m *Manager) CancelOffer(tradeID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, exists := m.offers[tradeID]
	if !exists {
		return ErrTradeNotFound
	}

	// Only initiator can cancel
	if offer.InitiatorID != playerID {
		return ErrNotAuthorized
	}

	// Check if can be cancelled
	if !offer.CanBeCancelled() {
		return ErrCannotCancel
	}

	offer.Cancel()

	// Release escrow if exists
	if escrow, exists := m.escrow[tradeID]; exists {
		delete(m.escrow, tradeID)
		_ = escrow // Escrow released
	}

	// Update history
	m.ensureHistory(playerID)
	m.history[playerID].RecordTrade(offer.GetTotalOfferedValue(), false)

	return nil
}

// CompleteTrade completes an accepted trade
func (m *Manager) CompleteTrade(tradeID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer, exists := m.offers[tradeID]
	if !exists {
		return ErrTradeNotFound
	}

	// Must be accepted
	if offer.Status != models.TradeStatusAccepted {
		return ErrInvalidStatus
	}

	// Complete the trade
	offer.Complete()

	// Release escrow
	delete(m.escrow, tradeID)

	// Update histories
	m.ensureHistory(offer.InitiatorID)
	m.ensureHistory(offer.RecipientID)

	tradeValue := offer.GetTotalOfferedValue()
	m.history[offer.InitiatorID].RecordTrade(tradeValue, true)
	m.history[offer.RecipientID].RecordTrade(tradeValue, true)

	return nil
}

// AddRating adds a rating for a player after a trade
func (m *Manager) AddRating(playerID uuid.UUID, positive bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureHistory(playerID)
	m.history[playerID].AddRating(positive)
}

// GetHistory returns a player's trade history
func (m *Manager) GetHistory(playerID uuid.UUID) *models.TradeHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if history, exists := m.history[playerID]; exists {
		return history
	}

	return models.NewTradeHistory(playerID)
}

// GetEscrow returns the escrow for a trade
func (m *Manager) GetEscrow(tradeID uuid.UUID) (*models.TradeEscrow, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	escrow, exists := m.escrow[tradeID]
	return escrow, exists
}

// CleanupExpiredOffers removes expired offers
func (m *Manager) CleanupExpiredOffers() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	for id, offer := range m.offers {
		if offer.IsExpired() {
			offer.MarkExpired()
			cleaned++
		}

		// Remove very old completed/rejected/expired offers (older than 24 hours)
		if offer.Status != models.TradeStatusPending &&
		   offer.Status != models.TradeStatusAccepted &&
		   offer.UpdatedAt.Add(24 * 60 * 60 * 1000000000).Before(offer.UpdatedAt) {
			delete(m.offers, id)

			// Remove from player lists
			m.removeFromPlayerList(offer.InitiatorID, id)
			m.removeFromPlayerList(offer.RecipientID, id)
		}
	}

	return cleaned
}

// ensureHistory creates a history record if it doesn't exist
func (m *Manager) ensureHistory(playerID uuid.UUID) {
	if _, exists := m.history[playerID]; !exists {
		m.history[playerID] = models.NewTradeHistory(playerID)
	}
}

// removeFromPlayerList removes an offer from a player's list
func (m *Manager) removeFromPlayerList(playerID uuid.UUID, tradeID uuid.UUID) {
	offers := m.byPlayer[playerID]
	for i, offer := range offers {
		if offer.ID == tradeID {
			m.byPlayer[playerID] = append(offers[:i], offers[i+1:]...)
			break
		}
	}
}

// GetAllActiveOffers returns all active offers (for admin/debug)
func (m *Manager) GetAllActiveOffers() []*models.TradeOffer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := []*models.TradeOffer{}
	for _, offer := range m.offers {
		if offer.Status == models.TradeStatusPending || offer.Status == models.TradeStatusAccepted {
			active = append(active, offer)
		}
	}

	return active
}

// GetStats returns overall trading statistics
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total_offers":     len(m.offers),
		"pending":          0,
		"accepted":         0,
		"completed":        0,
		"active_escrows":   len(m.escrow),
	}

	for _, offer := range m.offers {
		switch offer.Status {
		case models.TradeStatusPending:
			stats["pending"]++
		case models.TradeStatusAccepted:
			stats["accepted"]++
		case models.TradeStatusCompleted:
			stats["completed"]++
		}
	}

	return stats
}
