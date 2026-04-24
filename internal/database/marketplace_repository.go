// File: internal/database/marketplace_repository.go
// Project: Terminal Velocity
// Description: Persistence for player auctions in the marketplace.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// AuctionRow mirrors the marketplace_auctions schema without importing the
// marketplace package (which would create a cycle). The marketplace
// package's Manager translates to/from this struct when reading/writing.
//
// Bids are not persisted in the beta schema — they live in-memory on the
// Auction struct's BidHistory slice. When a bid is placed the manager
// calls UpdateAuctionBid so the current_bid + high_bidder fields stay in
// sync for restart recovery.
type AuctionRow struct {
	ID             uuid.UUID
	SellerID       uuid.UUID
	SellerName     string
	AuctionType    string
	ItemID         *uuid.UUID
	ItemName       string
	Quantity       int
	Description    string
	StartingBid    int64
	BuyoutPrice    int64
	CurrentBid     int64
	HighBidder     *uuid.UUID
	HighBidderName string
	StartTime      int64 // Unix seconds
	EndTime        int64 // Unix seconds
	Status         string
}

// MarketplaceRepository persists auctions so a server restart doesn't drop
// every active listing mid-sale. The marketplace.Manager wraps this to
// bridge its in-memory state to the DB.
type MarketplaceRepository struct {
	db *DB
}

// NewMarketplaceRepository wraps a *DB pool.
func NewMarketplaceRepository(db *DB) *MarketplaceRepository {
	return &MarketplaceRepository{db: db}
}

// UpsertAuction writes a full auction row. Idempotent via the primary key
// (id). Used both on CreateAuction (insert) and on state transitions
// (bid, buyout, cancel, expire) so the DB tracks the manager's view.
func (r *MarketplaceRepository) UpsertAuction(ctx context.Context, a *AuctionRow) error {
	query := `
		INSERT INTO marketplace_auctions
			(id, seller_id, seller_name, auction_type, item_id, item_name,
			 quantity, description, starting_bid, buyout_price, current_bid,
			 high_bidder, high_bidder_name, start_time, end_time, status)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			 TO_TIMESTAMP($14), TO_TIMESTAMP($15), $16)
		ON CONFLICT (id) DO UPDATE SET
			current_bid       = EXCLUDED.current_bid,
			high_bidder       = EXCLUDED.high_bidder,
			high_bidder_name  = EXCLUDED.high_bidder_name,
			end_time          = EXCLUDED.end_time,
			status            = EXCLUDED.status
	`
	if _, err := r.db.ExecContext(ctx, query,
		a.ID, a.SellerID, a.SellerName, a.AuctionType,
		a.ItemID, a.ItemName, a.Quantity, a.Description,
		a.StartingBid, a.BuyoutPrice, a.CurrentBid,
		a.HighBidder, a.HighBidderName,
		a.StartTime, a.EndTime, a.Status,
	); err != nil {
		return fmt.Errorf("upsert auction %s: %w", a.ID, err)
	}
	return nil
}

// LoadActiveAuctions returns every auction whose status is "active". Called
// on server startup to restore the in-memory manager's list from the last
// run. Expired auctions (end_time < now) should be swept by a separate
// tick handler; this method just honors the status field.
func (r *MarketplaceRepository) LoadActiveAuctions(ctx context.Context) ([]*AuctionRow, error) {
	query := `
		SELECT id, seller_id, seller_name, auction_type, item_id, item_name,
		       quantity, description, starting_bid, buyout_price, current_bid,
		       high_bidder, high_bidder_name,
		       EXTRACT(EPOCH FROM start_time)::bigint,
		       EXTRACT(EPOCH FROM end_time)::bigint,
		       status
		FROM marketplace_auctions
		WHERE status = 'active'
		ORDER BY end_time ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("load active auctions: %w", err)
	}
	defer rows.Close()

	out := make([]*AuctionRow, 0)
	for rows.Next() {
		var a AuctionRow
		var itemID, highBidder sql.NullString
		var highBidderName sql.NullString
		var description sql.NullString
		if err := rows.Scan(
			&a.ID, &a.SellerID, &a.SellerName, &a.AuctionType,
			&itemID, &a.ItemName, &a.Quantity, &description,
			&a.StartingBid, &a.BuyoutPrice, &a.CurrentBid,
			&highBidder, &highBidderName,
			&a.StartTime, &a.EndTime, &a.Status,
		); err != nil {
			return nil, fmt.Errorf("scan auction row: %w", err)
		}
		if itemID.Valid {
			id, err := uuid.Parse(itemID.String)
			if err == nil {
				a.ItemID = &id
			}
		}
		if highBidder.Valid {
			id, err := uuid.Parse(highBidder.String)
			if err == nil {
				a.HighBidder = &id
			}
		}
		if highBidderName.Valid {
			a.HighBidderName = highBidderName.String
		}
		if description.Valid {
			a.Description = description.String
		}
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auctions: %w", err)
	}
	return out, nil
}
