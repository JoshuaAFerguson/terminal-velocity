// File: internal/models/trade.go
// Project: Terminal Velocity
// Description: Player-to-player trading models with escrow system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TradeStatus represents the current state of a trade offer

type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"   // Awaiting response
	TradeStatusAccepted  TradeStatus = "accepted"  // Both parties agreed, in escrow
	TradeStatusCompleted TradeStatus = "completed" // Successfully executed
	TradeStatusRejected  TradeStatus = "rejected"  // Declined by recipient
	TradeStatusCancelled TradeStatus = "cancelled" // Cancelled by initiator
	TradeStatusExpired   TradeStatus = "expired"   // Offer timed out
)

// TradeItem represents a single item in a trade offer
type TradeItem struct {
	CommodityName string `json:"commodity_name"`
	Quantity      int    `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"` // For display purposes
}

// TradeOffer represents a trade proposal between two players
type TradeOffer struct {
	ID            uuid.UUID `json:"id"`
	InitiatorID   uuid.UUID `json:"initiator_id"`
	InitiatorName string    `json:"initiator_name"`
	RecipientID   uuid.UUID `json:"recipient_id"`
	RecipientName string    `json:"recipient_name"`

	// What the initiator is offering
	OfferedCredits int64       `json:"offered_credits"`
	OfferedItems   []TradeItem `json:"offered_items"`

	// What the initiator wants in return
	RequestedCredits int64       `json:"requested_credits"`
	RequestedItems   []TradeItem `json:"requested_items"`

	// Status and timing
	Status    TradeStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Location requirement
	SystemID uuid.UUID `json:"system_id"` // Must be in same system
	PlanetID uuid.UUID `json:"planet_id"` // Must be docked at same planet

	// Optional message
	Message string `json:"message,omitempty"`

	// Contract mode
	IsContract    bool   `json:"is_contract"` // If true, can't be cancelled after accept
	ContractTerms string `json:"contract_terms,omitempty"`
}

// TradeEscrow represents items/credits held during a trade
type TradeEscrow struct {
	TradeID uuid.UUID `json:"trade_id"`

	// Locked funds from initiator
	InitiatorCredits int64          `json:"initiator_credits"`
	InitiatorItems   map[string]int `json:"initiator_items"` // commodity -> quantity

	// Locked funds from recipient
	RecipientCredits int64          `json:"recipient_credits"`
	RecipientItems   map[string]int `json:"recipient_items"`

	LockedAt time.Time `json:"locked_at"`
}

// TradeHistory tracks completed trades for reputation
type TradeHistory struct {
	PlayerID         uuid.UUID `json:"player_id"`
	TotalTrades      int       `json:"total_trades"`
	SuccessfulTrades int       `json:"successful_trades"`
	CancelledTrades  int       `json:"cancelled_trades"`
	TotalVolume      int64     `json:"total_volume"` // Total credits traded
	LastTradeAt      time.Time `json:"last_trade_at"`

	// Reputation
	PositiveRatings int     `json:"positive_ratings"`
	NegativeRatings int     `json:"negative_ratings"`
	TrustScore      float64 `json:"trust_score"` // 0.0 - 1.0
}

// NewTradeOffer creates a new trade offer
func NewTradeOffer(
	initiatorID uuid.UUID,
	initiatorName string,
	recipientID uuid.UUID,
	recipientName string,
	systemID uuid.UUID,
	planetID uuid.UUID,
) *TradeOffer {
	now := time.Now()

	return &TradeOffer{
		ID:             uuid.New(),
		InitiatorID:    initiatorID,
		InitiatorName:  initiatorName,
		RecipientID:    recipientID,
		RecipientName:  recipientName,
		Status:         TradeStatusPending,
		CreatedAt:      now,
		ExpiresAt:      now.Add(1 * time.Hour), // 1 hour default expiry
		UpdatedAt:      now,
		SystemID:       systemID,
		PlanetID:       planetID,
		OfferedItems:   []TradeItem{},
		RequestedItems: []TradeItem{},
		IsContract:     false,
	}
}

// AddOfferedCredits adds credits to the offer
func (t *TradeOffer) AddOfferedCredits(amount int64) {
	t.OfferedCredits += amount
	t.UpdatedAt = time.Now()
}

// AddOfferedItem adds an item to the offer
func (t *TradeOffer) AddOfferedItem(commodity string, quantity int, unitPrice int64) {
	t.OfferedItems = append(t.OfferedItems, TradeItem{
		CommodityName: commodity,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
	})
	t.UpdatedAt = time.Now()
}

// AddRequestedCredits adds credits to the request
func (t *TradeOffer) AddRequestedCredits(amount int64) {
	t.RequestedCredits += amount
	t.UpdatedAt = time.Now()
}

// AddRequestedItem adds an item to the request
func (t *TradeOffer) AddRequestedItem(commodity string, quantity int, unitPrice int64) {
	t.RequestedItems = append(t.RequestedItems, TradeItem{
		CommodityName: commodity,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
	})
	t.UpdatedAt = time.Now()
}

// IsExpired checks if the offer has expired
func (t *TradeOffer) IsExpired() bool {
	return time.Now().After(t.ExpiresAt) && t.Status == TradeStatusPending
}

// CanBeCancelled checks if the offer can be cancelled
func (t *TradeOffer) CanBeCancelled() bool {
	// Contracts can't be cancelled after acceptance
	if t.IsContract && t.Status == TradeStatusAccepted {
		return false
	}
	return t.Status == TradeStatusPending || t.Status == TradeStatusAccepted
}

// Accept marks the trade as accepted
func (t *TradeOffer) Accept() {
	t.Status = TradeStatusAccepted
	t.UpdatedAt = time.Now()
}

// Reject marks the trade as rejected
func (t *TradeOffer) Reject() {
	t.Status = TradeStatusRejected
	t.UpdatedAt = time.Now()
}

// Cancel marks the trade as cancelled
func (t *TradeOffer) Cancel() {
	if t.CanBeCancelled() {
		t.Status = TradeStatusCancelled
		t.UpdatedAt = time.Now()
	}
}

// Complete marks the trade as completed
func (t *TradeOffer) Complete() {
	t.Status = TradeStatusCompleted
	t.UpdatedAt = time.Now()
}

// MarkExpired marks the trade as expired
func (t *TradeOffer) MarkExpired() {
	t.Status = TradeStatusExpired
	t.UpdatedAt = time.Now()
}

// GetTotalOfferedValue calculates total value of offered items + credits
func (t *TradeOffer) GetTotalOfferedValue() int64 {
	total := t.OfferedCredits
	for _, item := range t.OfferedItems {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return total
}

// GetTotalRequestedValue calculates total value of requested items + credits
func (t *TradeOffer) GetTotalRequestedValue() int64 {
	total := t.RequestedCredits
	for _, item := range t.RequestedItems {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return total
}

// GetValueRatio returns the ratio of offered to requested value
func (t *TradeOffer) GetValueRatio() float64 {
	offered := float64(t.GetTotalOfferedValue())
	requested := float64(t.GetTotalRequestedValue())

	if requested == 0 {
		if offered > 0 {
			return 999.0 // Giving for free
		}
		return 1.0 // Both zero
	}

	return offered / requested
}

// GetFairnessRating returns a human-readable fairness assessment
func (t *TradeOffer) GetFairnessRating() string {
	ratio := t.GetValueRatio()

	switch {
	case ratio >= 0.95 && ratio <= 1.05:
		return "⚖️  Fair Trade"
	case ratio > 1.05 && ratio <= 1.2:
		return "💰 Good Deal"
	case ratio > 1.2:
		return "🎁 Very Generous"
	case ratio >= 0.8 && ratio < 0.95:
		return "⚠️  Slightly Unfavorable"
	case ratio < 0.8:
		return "❌ Poor Deal"
	default:
		return "❓ Unknown"
	}
}

// GetStatusIcon returns an icon for the trade status
func (t TradeStatus) GetIcon() string {
	icons := map[TradeStatus]string{
		TradeStatusPending:   "⏳",
		TradeStatusAccepted:  "✅",
		TradeStatusCompleted: "✔️",
		TradeStatusRejected:  "❌",
		TradeStatusCancelled: "🚫",
		TradeStatusExpired:   "⌛",
	}

	if icon, exists := icons[t]; exists {
		return icon
	}
	return "❓"
}

// GetStatusColor returns a color for the trade status
func (t TradeStatus) GetColor() string {
	colors := map[TradeStatus]string{
		TradeStatusPending:   "yellow",
		TradeStatusAccepted:  "green",
		TradeStatusCompleted: "blue",
		TradeStatusRejected:  "red",
		TradeStatusCancelled: "gray",
		TradeStatusExpired:   "gray",
	}

	if color, exists := colors[t]; exists {
		return color
	}
	return "white"
}

// GetTimeRemaining returns a human-readable time remaining string
func (t *TradeOffer) GetTimeRemaining() string {
	if t.Status != TradeStatusPending {
		return "-"
	}

	duration := time.Until(t.ExpiresAt)
	if duration < 0 {
		return "Expired"
	}

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// NewTradeEscrow creates a new escrow for a trade
func NewTradeEscrow(tradeID uuid.UUID) *TradeEscrow {
	return &TradeEscrow{
		TradeID:        tradeID,
		InitiatorItems: make(map[string]int),
		RecipientItems: make(map[string]int),
		LockedAt:       time.Now(),
	}
}

// LockInitiatorAssets locks the initiator's offered assets
func (e *TradeEscrow) LockInitiatorAssets(credits int64, items map[string]int) {
	e.InitiatorCredits = credits
	e.InitiatorItems = items
}

// LockRecipientAssets locks the recipient's requested assets
func (e *TradeEscrow) LockRecipientAssets(credits int64, items map[string]int) {
	e.RecipientCredits = credits
	e.RecipientItems = items
}

// NewTradeHistory creates a new trade history for a player
func NewTradeHistory(playerID uuid.UUID) *TradeHistory {
	return &TradeHistory{
		PlayerID:         playerID,
		TotalTrades:      0,
		SuccessfulTrades: 0,
		CancelledTrades:  0,
		TotalVolume:      0,
		TrustScore:       1.0, // Start with perfect trust
		PositiveRatings:  0,
		NegativeRatings:  0,
	}
}

// RecordTrade records a completed trade
func (h *TradeHistory) RecordTrade(value int64, successful bool) {
	h.TotalTrades++
	if successful {
		h.SuccessfulTrades++
		h.TotalVolume += value
	} else {
		h.CancelledTrades++
	}
	h.LastTradeAt = time.Now()
	h.UpdateTrustScore()
}

// AddRating adds a rating from a trade partner
func (h *TradeHistory) AddRating(positive bool) {
	if positive {
		h.PositiveRatings++
	} else {
		h.NegativeRatings++
	}
	h.UpdateTrustScore()
}

// UpdateTrustScore recalculates the trust score
func (h *TradeHistory) UpdateTrustScore() {
	// Base score from trade completion rate
	completionRate := 0.0
	if h.TotalTrades > 0 {
		completionRate = float64(h.SuccessfulTrades) / float64(h.TotalTrades)
	} else {
		completionRate = 1.0 // Benefit of the doubt for new traders
	}

	// Rating score
	totalRatings := h.PositiveRatings + h.NegativeRatings
	ratingScore := 1.0
	if totalRatings > 0 {
		ratingScore = float64(h.PositiveRatings) / float64(totalRatings)
	}

	// Weighted average (completion 60%, ratings 40%)
	h.TrustScore = (completionRate * 0.6) + (ratingScore * 0.4)
}

// GetTrustRating returns a human-readable trust rating
func (h *TradeHistory) GetTrustRating() string {
	switch {
	case h.TrustScore >= 0.95:
		return "⭐⭐⭐⭐⭐ Excellent"
	case h.TrustScore >= 0.85:
		return "⭐⭐⭐⭐ Very Good"
	case h.TrustScore >= 0.70:
		return "⭐⭐⭐ Good"
	case h.TrustScore >= 0.50:
		return "⭐⭐ Fair"
	case h.TrustScore >= 0.30:
		return "⭐ Poor"
	default:
		return "❌ Untrusted"
	}
}

// GetCompletionRate returns the percentage of successful trades
func (h *TradeHistory) GetCompletionRate() float64 {
	if h.TotalTrades == 0 {
		return 100.0
	}
	return (float64(h.SuccessfulTrades) / float64(h.TotalTrades)) * 100.0
}
