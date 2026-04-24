// File: internal/server/marketplace_adapter.go
// Project: Terminal Velocity
// Description: Server-side bridge from marketplace.AuctionPersister to
//   database.MarketplaceRepository. Lives here (not in internal/database)
//   to avoid a marketplace <-> database import cycle.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package server

import (
	"context"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/marketplace"
	"github.com/google/uuid"
)

// auctionPersister satisfies marketplace.AuctionPersister using a
// database.MarketplaceRepository underneath.
type auctionPersister struct {
	repo *database.MarketplaceRepository
}

func newAuctionPersister(repo *database.MarketplaceRepository) *auctionPersister {
	return &auctionPersister{repo: repo}
}

// SaveAuction converts a marketplace.Auction into the repo's AuctionRow
// shape and upserts it. Errors propagate to the marketplace.Manager,
// which currently swallows them because the in-memory state is the
// authority — the next mutation retries.
func (a *auctionPersister) SaveAuction(auction *marketplace.Auction) error {
	if auction == nil {
		return nil
	}
	row := &database.AuctionRow{
		ID:             auction.ID,
		SellerID:       auction.SellerID,
		SellerName:     auction.SellerName,
		AuctionType:    string(auction.Type),
		ItemName:       auction.ItemName,
		Quantity:       auction.Quantity,
		Description:    auction.Description,
		StartingBid:    auction.StartingBid,
		BuyoutPrice:    auction.BuyoutPrice,
		CurrentBid:     auction.CurrentBid,
		HighBidderName: auction.HighBidderName,
		StartTime:      auction.StartTime.Unix(),
		EndTime:        auction.EndTime.Unix(),
		Status:         auction.Status,
	}
	// Copy optional-looking fields. uuid.Nil is treated as "unset" so
	// the DB column stays NULL — matters for the FK on high_bidder
	// when no one has bid yet and for seller-only items with no
	// item_id (e.g. stackable commodities).
	if auction.ItemID != uuid.Nil {
		id := auction.ItemID
		row.ItemID = &id
	}
	if auction.HighBidder != uuid.Nil {
		id := auction.HighBidder
		row.HighBidder = &id
	}
	return a.repo.UpsertAuction(context.Background(), row)
}
