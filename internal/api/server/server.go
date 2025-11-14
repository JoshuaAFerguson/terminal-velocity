// File: internal/api/server/server.go
// Project: Terminal Velocity
// Description: In-process API server implementation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// GameServer implements api.Server interface
// In Phase 1, this server runs in-process with the SSH gateway
// In Phase 2+, this becomes a standalone gRPC server
type GameServer struct {
	// Database repositories
	playerRepo *database.PlayerRepository
	systemRepo *database.SystemRepository
	shipRepo   *database.ShipRepository
	marketRepo *database.MarketRepository
	sshKeyRepo *database.SSHKeyRepository

	// Manager packages (existing game logic)
	// These will be gradually refactored as we implement API handlers
	// For now, we'll keep them as placeholders

	// Session management
	sessions *SessionManager
}

// NewGameServer creates a new in-process game server
func NewGameServer(config *Config) (*GameServer, error) {
	server := &GameServer{
		playerRepo: config.PlayerRepo,
		systemRepo: config.SystemRepo,
		shipRepo:   config.ShipRepo,
		marketRepo: config.MarketRepo,
		sshKeyRepo: config.SSHKeyRepo,
		sessions:   NewSessionManager(),
	}

	return server, nil
}

// Config for the game server
type Config struct {
	// Database repositories
	PlayerRepo *database.PlayerRepository
	SystemRepo *database.SystemRepository
	ShipRepo   *database.ShipRepository
	MarketRepo *database.MarketRepository
	SSHKeyRepo *database.SSHKeyRepository

	// TODO: Add manager configuration as we implement handlers
}

// Compile-time check that GameServer implements api.Server
var _ api.Server = (*GameServer)(nil)

// ============================================================================
// AuthService Implementation
// ============================================================================

// Authenticate validates username/password and returns session token
func (s *GameServer) Authenticate(ctx context.Context, req *api.AuthRequest) (*api.AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, api.ErrInvalidRequest
	}

	player, err := s.playerRepo.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		if err == database.ErrInvalidCredentials {
			return nil, api.ErrUnauthorized
		}
		return nil, err
	}

	session, err := s.sessions.CreateSession(player.ID)
	if err != nil {
		return nil, err
	}

	_ = s.playerRepo.UpdateLastLogin(ctx, player.ID)

	return &api.AuthResponse{
		PlayerID:  player.ID,
		Token:     session.SessionID.String(),
		IssuedAt:  session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
		PlayerInfo: &api.PlayerInfo{
			PlayerID:  player.ID,
			Username:  player.Username,
			Email:     player.Email,
			CreatedAt: player.CreatedAt,
			LastLogin: time.Now(),
			IsAdmin:   false,
			Role:      "player",
		},
	}, nil
}

// AuthenticateSSH validates SSH public key and returns session token
func (s *GameServer) AuthenticateSSH(ctx context.Context, req *api.SSHAuthRequest) (*api.AuthResponse, error) {
	if req.Username == "" || len(req.PublicKey) == 0 {
		return nil, api.ErrInvalidRequest
	}

	playerID, err := s.sshKeyRepo.GetPlayerIDByPublicKey(ctx, req.PublicKey)
	if err != nil {
		if err == database.ErrSSHKeyNotFound {
			return nil, api.ErrUnauthorized
		}
		return nil, err
	}

	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, api.ErrUnauthorized
	}

	if player.Username != req.Username {
		return nil, api.ErrUnauthorized
	}

	session, err := s.sessions.CreateSession(player.ID)
	if err != nil {
		return nil, err
	}

	_ = s.playerRepo.UpdateLastLogin(ctx, player.ID)

	return &api.AuthResponse{
		PlayerID:  player.ID,
		Token:     session.SessionID.String(),
		IssuedAt:  session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
		PlayerInfo: &api.PlayerInfo{
			PlayerID:  player.ID,
			Username:  player.Username,
			Email:     player.Email,
			CreatedAt: player.CreatedAt,
			LastLogin: time.Now(),
			IsAdmin:   false,
			Role:      "player",
		},
	}, nil
}

// CreateSession creates a new game session for an authenticated player
func (s *GameServer) CreateSession(ctx context.Context, req *api.CreateSessionRequest) (*api.Session, error) {
	session, err := s.sessions.CreateSession(req.PlayerID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession checks if a session is still valid
func (s *GameServer) ValidateSession(ctx context.Context, req *api.ValidateSessionRequest) (*api.Session, error) {
	if req.SessionID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	session, err := s.sessions.GetSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	if req.Token != session.SessionID.String() {
		return nil, api.ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) || session.State != api.SessionStateActive {
		return nil, api.ErrUnauthorized
	}

	return session, nil
}

// EndSession terminates an active session
func (s *GameServer) EndSession(ctx context.Context, req *api.EndSessionRequest) error {
	return s.sessions.EndSession(req.SessionID)
}

// RefreshSession extends a session's lifetime
func (s *GameServer) RefreshSession(ctx context.Context, req *api.RefreshSessionRequest) (*api.AuthResponse, error) {
	if req.SessionID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	session, err := s.sessions.GetSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	if req.Token != session.SessionID.String() || time.Now().After(session.ExpiresAt) {
		return nil, api.ErrUnauthorized
	}

	refreshedSession, err := s.sessions.RefreshSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player, err := s.playerRepo.GetByID(ctx, refreshedSession.PlayerID)
	if err != nil {
		return nil, err
	}

	return &api.AuthResponse{
		PlayerID:  player.ID,
		Token:     refreshedSession.SessionID.String(),
		IssuedAt:  refreshedSession.CreatedAt,
		ExpiresAt: refreshedSession.ExpiresAt,
		PlayerInfo: &api.PlayerInfo{
			PlayerID:  player.ID,
			Username:  player.Username,
			Email:     player.Email,
			CreatedAt: player.CreatedAt,
			LastLogin: player.LastLogin,
			IsAdmin:   false,
			Role:      "player",
		},
	}, nil
}

// Register creates a new player account
func (s *GameServer) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	if req.Username == "" || (req.Password == "" && len(req.SSHPublicKey) == 0) {
		return nil, api.ErrInvalidRequest
	}

	_, err := s.playerRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return &api.RegisterResponse{
			PlayerID: uuid.Nil,
			Username: req.Username,
			Message:  "username already exists",
		}, nil
	}

	var player *models.Player
	if req.Password != "" {
		player, err = s.playerRepo.CreateWithEmail(ctx, req.Username, req.Password, req.Email)
	} else {
		player, err = s.playerRepo.CreateWithEmail(ctx, req.Username, "", req.Email)
	}

	if err != nil {
		if err == database.ErrUsernameExists {
			return &api.RegisterResponse{
				PlayerID: uuid.Nil,
				Username: req.Username,
				Message:  "username already exists",
			}, nil
		}
		return nil, err
	}

	if len(req.SSHPublicKey) > 0 {
		_, _ = s.sshKeyRepo.AddKey(ctx, player.ID, string(req.SSHPublicKey))
	}

	return &api.RegisterResponse{
		PlayerID: player.ID,
		Username: player.Username,
		Message:  "registration successful",
	}, nil
}

// ============================================================================
// PlayerService Implementation
// ============================================================================

// GetPlayerState retrieves complete player state
func (s *GameServer) GetPlayerState(ctx context.Context, playerID uuid.UUID) (*api.PlayerState, error) {
	// Load player from database
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Load player's current ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return nil, err
	}

	// Convert to API types
	state := convertPlayerToAPI(player, ship)
	state.Stats = convertPlayerStatsToAPI(player)
	state.Reputation = convertReputationToAPI(player)

	return state, nil
}

// UpdatePlayerLocation updates player's location and coordinates
func (s *GameServer) UpdatePlayerLocation(ctx context.Context, req *api.LocationUpdate) (*api.PlayerState, error) {
	if req.PlayerID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player to verify they exist
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// Update system and planet if provided
	if req.SystemID != uuid.Nil {
		err = s.playerRepo.UpdateLocation(ctx, req.PlayerID, req.SystemID, req.PlanetID)
		if err != nil {
			return nil, err
		}
		player.CurrentSystem = req.SystemID
		player.CurrentPlanet = req.PlanetID
	}

	// Update position coordinates
	err = s.playerRepo.UpdatePosition(ctx, req.PlayerID, req.Position.X, req.Position.Y)
	if err != nil {
		return nil, err
	}
	player.X = req.Position.X
	player.Y = req.Position.Y

	// Load ship to return complete state
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return nil, err
	}

	// Return updated player state
	return convertPlayerToAPI(player, ship), nil
}

// GetPlayerShip retrieves player's current ship
func (s *GameServer) GetPlayerShip(ctx context.Context, playerID uuid.UUID) (*api.Ship, error) {
	// Load player to get current ship ID
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Load the ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return nil, err
	}

	// Convert to API ship
	return convertShipToAPI(ship), nil
}

// GetPlayerInventory retrieves player's cargo and items
func (s *GameServer) GetPlayerInventory(ctx context.Context, playerID uuid.UUID) (*api.Inventory, error) {
	// Load player to get current ship ID
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Load the ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return nil, err
	}

	// Convert to API inventory
	return convertInventoryToAPI(ship), nil
}

// GetPlayerStats retrieves player statistics
func (s *GameServer) GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*api.PlayerStats, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Convert to API stats
	return convertPlayerStatsToAPI(player), nil
}

// GetPlayerReputation retrieves faction reputation
func (s *GameServer) GetPlayerReputation(ctx context.Context, playerID uuid.UUID) (*api.ReputationInfo, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Convert to API reputation
	return convertReputationToAPI(player), nil
}

// StreamPlayerUpdates subscribes to real-time player state changes
func (s *GameServer) StreamPlayerUpdates(ctx context.Context, playerID uuid.UUID) (api.PlayerUpdateStream, error) {
	// TODO: Implement streaming
	// Will use channels to push updates to client
	return nil, api.ErrNotFound
}

// ============================================================================
// GameService Implementation
// ============================================================================

// Jump performs a hyperspace jump to another system
func (s *GameServer) Jump(ctx context.Context, req *api.JumpRequest) (*api.JumpResponse, error) {
	if req.PlayerID == uuid.Nil || req.TargetSystemID == uuid.Nil {
		return &api.JumpResponse{
			Success: false,
			Message: "invalid request: missing player or target system ID",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Check if player is docked (can't jump while docked)
	if player.IsDocked() {
		return &api.JumpResponse{
			Success: false,
			Message: "cannot jump while docked - takeoff first",
		}, nil
	}

	// Verify jump route exists
	connections, err := s.systemRepo.GetConnections(ctx, player.CurrentSystem)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "failed to load jump routes",
		}, nil
	}

	// Check if target system is connected
	canJump := false
	for _, connectedSystemID := range connections {
		if connectedSystemID == req.TargetSystemID {
			canJump = true
			break
		}
	}

	if !canJump {
		return &api.JumpResponse{
			Success: false,
			Message: "no jump route to target system",
		}, nil
	}

	// Load ship to check fuel
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Check fuel (1 fuel per jump)
	fuelCost := 1
	if ship.Fuel < fuelCost {
		return &api.JumpResponse{
			Success: false,
			Message: "insufficient fuel",
		}, nil
	}

	// Perform the jump
	// Update player location to new system, clear planet, reset position to (0,0)
	err = s.playerRepo.UpdateLocation(ctx, req.PlayerID, req.TargetSystemID, nil)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "failed to update location",
		}, nil
	}

	err = s.playerRepo.UpdatePosition(ctx, req.PlayerID, 0, 0)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "failed to update position",
		}, nil
	}

	// Consume fuel
	ship.Fuel -= fuelCost
	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "failed to update ship fuel",
		}, nil
	}

	// Record jump in player stats (update TotalJumps)
	player.RecordJump()
	err = s.playerRepo.Update(ctx, player)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "failed to update player stats",
		}, nil
	}

	// Update local player state
	player.CurrentSystem = req.TargetSystemID
	player.CurrentPlanet = nil
	player.X = 0
	player.Y = 0

	// Return success with new state
	return &api.JumpResponse{
		Success:      true,
		Message:      "jump successful",
		NewState:     convertPlayerToAPI(player, ship),
		FuelConsumed: int32(fuelCost),
	}, nil
}

// Land lands on a planet
func (s *GameServer) Land(ctx context.Context, req *api.LandRequest) (*api.LandResponse, error) {
	if req.PlayerID == uuid.Nil || req.PlanetID == uuid.Nil {
		return &api.LandResponse{
			Success: false,
			Message: "invalid request: missing player or planet ID",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Check if already docked
	if player.IsDocked() {
		return &api.LandResponse{
			Success: false,
			Message: "already docked - takeoff first",
		}, nil
	}

	// Load planet to verify it exists and is in current system
	planet, err := s.systemRepo.GetPlanetByID(ctx, req.PlanetID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "planet not found",
		}, nil
	}

	// Verify planet is in player's current system
	if planet.SystemID != player.CurrentSystem {
		return &api.LandResponse{
			Success: false,
			Message: "planet is not in current system",
		}, nil
	}

	// TODO: Check distance to planet (requires planet X/Y coordinates)
	// For now, we'll allow landing from anywhere in the system

	// Update player location to be docked at planet
	err = s.playerRepo.UpdateLocation(ctx, req.PlayerID, player.CurrentSystem, &req.PlanetID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "failed to update location",
		}, nil
	}

	// Update local player state
	player.CurrentPlanet = &req.PlanetID

	// Load ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Return success with updated state
	return &api.LandResponse{
		Success:  true,
		Message:  "landed successfully",
		Planet:   convertPlanetToAPI(planet),
		NewState: convertPlayerToAPI(player, ship),
	}, nil
}

// Takeoff takes off from a planet
func (s *GameServer) Takeoff(ctx context.Context, req *api.TakeoffRequest) (*api.TakeoffResponse, error) {
	if req.PlayerID == uuid.Nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "invalid request: missing player ID",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Check if player is docked
	if !player.IsDocked() {
		return &api.TakeoffResponse{
			Success: false,
			Message: "not docked at a planet",
		}, nil
	}

	// Update player location to clear planet (takeoff into space)
	err = s.playerRepo.UpdateLocation(ctx, req.PlayerID, player.CurrentSystem, nil)
	if err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "failed to update location",
		}, nil
	}

	// Update local player state
	player.CurrentPlanet = nil

	// Load ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Return success with updated state
	return &api.TakeoffResponse{
		Success:  true,
		Message:  "takeoff successful",
		NewState: convertPlayerToAPI(player, ship),
	}, nil
}

// GetMarket retrieves market data for a system
func (s *GameServer) GetMarket(ctx context.Context, systemID uuid.UUID) (*api.Market, error) {
	if systemID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Get all market prices for planets in this system
	prices, err := s.marketRepo.GetCommoditiesBySystemID(ctx, systemID)
	if err != nil {
		return nil, err
	}

	// Build commodity map from standard commodities
	commodities := make(map[string]*models.Commodity)
	for i := range models.StandardCommodities {
		commodity := &models.StandardCommodities[i]
		commodities[commodity.ID] = commodity
	}

	// Convert to API market
	// Note: This aggregates prices from all planets in the system
	// In a real implementation, you might want to show the best prices or average prices
	market := convertMarketToAPI(prices, commodities)
	market.SystemID = systemID

	return market, nil
}

// BuyCommodity purchases a commodity from the market
func (s *GameServer) BuyCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
	if req.PlayerID == uuid.Nil || req.CommodityID == "" || req.Quantity <= 0 {
		return &api.TradeResponse{
			Success: false,
			Message: "invalid request: missing player, commodity, or quantity",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Must be docked to trade
	if !player.IsDocked() {
		return &api.TradeResponse{
			Success: false,
			Message: "must be docked at a planet to trade",
		}, nil
	}

	// Load ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Get market price
	price, err := s.marketRepo.GetMarketPrice(ctx, *player.CurrentPlanet, req.CommodityID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "commodity not available at this planet",
		}, nil
	}

	// Check stock availability
	if price.Stock < int(req.Quantity) {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient stock (available: %d)", price.Stock),
		}, nil
	}

	// Calculate cost (player buys at sell price)
	totalCost := price.SellPrice * int64(req.Quantity)

	// Check credits
	if !player.CanAfford(totalCost) {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient credits (need: %d, have: %d)", totalCost, player.Credits),
		}, nil
	}

	// TODO: Check cargo space (requires ShipType lookup)
	// For now, just add to cargo

	// Perform the trade
	// 1. Deduct credits
	player.AddCredits(-totalCost)
	err = s.playerRepo.Update(ctx, player)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update player credits",
		}, nil
	}

	// 2. Add cargo to ship
	ship.AddCargo(req.CommodityID, int(req.Quantity))
	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update ship cargo",
		}, nil
	}

	// 3. Update market stock
	err = s.marketRepo.UpdateStock(ctx, *player.CurrentPlanet, req.CommodityID, -int(req.Quantity))
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update market stock",
		}, nil
	}

	// Return success
	return &api.TradeResponse{
		Success:        true,
		Message:        "purchase successful",
		QuantityTraded: req.Quantity,
		TotalCost:      totalCost,
		PricePerUnit:   int32(price.SellPrice),
		NewState:       convertPlayerToAPI(player, ship),
	}, nil
}


// SellCommodity sells a commodity to the market
func (s *GameServer) SellCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
	if req.PlayerID == uuid.Nil || req.CommodityID == "" || req.Quantity <= 0 {
		return &api.TradeResponse{
			Success: false,
			Message: "invalid request: missing player, commodity, or quantity",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Must be docked to trade
	if !player.IsDocked() {
		return &api.TradeResponse{
			Success: false,
			Message: "must be docked at a planet to trade",
		}, nil
	}

	// Load ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Check if player has the commodity
	currentQuantity := ship.GetCommodityQuantity(req.CommodityID)
	if currentQuantity < int(req.Quantity) {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient cargo (have: %d, trying to sell: %d)", currentQuantity, req.Quantity),
		}, nil
	}

	// Get market price
	price, err := s.marketRepo.GetMarketPrice(ctx, *player.CurrentPlanet, req.CommodityID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "commodity not traded at this planet",
		}, nil
	}

	// Check demand (can't sell more than demand)
	if price.Demand < int(req.Quantity) {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient demand (available: %d)", price.Demand),
		}, nil
	}

	// Calculate payment (player sells at buy price)
	totalPayment := price.BuyPrice * int64(req.Quantity)

	// Perform the trade
	// 1. Add credits
	player.AddCredits(totalPayment)
	err = s.playerRepo.Update(ctx, player)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update player credits",
		}, nil
	}

	// 2. Remove cargo from ship
	success := ship.RemoveCargo(req.CommodityID, int(req.Quantity))
	if !success {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to remove cargo from ship",
		}, nil
	}

	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update ship cargo",
		}, nil
	}

	// 3. Update market stock (increase stock, decrease demand)
	err = s.marketRepo.UpdateStock(ctx, *player.CurrentPlanet, req.CommodityID, int(req.Quantity))
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "failed to update market stock",
		}, nil
	}

	// Return success
	return &api.TradeResponse{
		Success:        true,
		Message:        "sale successful",
		QuantityTraded: req.Quantity,
		TotalCost:      totalPayment,
		PricePerUnit:   int32(price.BuyPrice),
		NewState:       convertPlayerToAPI(player, ship),
	}, nil
}


// BuyShip purchases a new ship
// TODO: Reimplement with correct models
func (s *GameServer) BuyShip(ctx context.Context, req *api.ShipPurchaseRequest) (*api.ShipPurchaseResponse, error) {
	return &api.ShipPurchaseResponse{
		Success: false,
		Message: "BuyShip not yet implemented - requires model updates",
	}, nil
}


// SellShip sells a ship
// TODO: Reimplement with correct models
func (s *GameServer) SellShip(ctx context.Context, req *api.ShipSaleRequest) (*api.ShipSaleResponse, error) {
	return &api.ShipSaleResponse{
		Success: false,
		Message: "SellShip not yet implemented - requires model updates",
	}, nil
}


// BuyOutfit purchases ship equipment
// TODO: Reimplement with correct models
func (s *GameServer) BuyOutfit(ctx context.Context, req *api.OutfitPurchaseRequest) (*api.OutfitPurchaseResponse, error) {
	return &api.OutfitPurchaseResponse{
		Success: false,
		Message: "BuyOutfit not yet implemented - requires model updates",
	}, nil
}


// SellOutfit sells ship equipment
// TODO: Reimplement with correct models
func (s *GameServer) SellOutfit(ctx context.Context, req *api.OutfitSaleRequest) (*api.OutfitSaleResponse, error) {
	return &api.OutfitSaleResponse{
		Success: false,
		Message: "SellOutfit not yet implemented - requires model updates",
	}, nil
}


// GetAvailableMissions retrieves missions available to player
func (s *GameServer) GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*api.MissionList, error) {
	// Load player to check current location and stats
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	_ = player

	// TODO: Integrate with missions manager
	// For now, return empty list as missions system needs manager integration
	// When implemented, this should:
	// 1. Query available missions at current planet/system
	// 2. Filter by player level/reputation
	// 3. Convert to API format using convertMissionToAPI

	missions := &api.MissionList{
		Missions:   make([]*api.Mission, 0),
		TotalCount: 0,
	}

	// Placeholder: In production, query from missions manager
	// missions := s.missionsManager.GetAvailableMissions(player.CurrentSystem, player.Level, player.FactionReputation)

	return missions, nil
}

// AcceptMission accepts a mission
func (s *GameServer) AcceptMission(ctx context.Context, req *api.MissionAcceptRequest) (*api.Mission, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// TODO: Integrate with missions manager
	// For now, return error as missions system needs manager integration
	// When implemented, this should:
	// 1. Verify mission exists and is available
	// 2. Check if player meets requirements
	// 3. Add mission to player's active missions
	// 4. Update mission status to "active"
	// 5. Return updated mission

	// Placeholder: In production, use missions manager
	// mission, err := s.missionsManager.AcceptMission(req.PlayerID, req.MissionID)
	// if err != nil {
	//     return nil, err
	// }
	// return convertMissionToAPI(mission), nil

	_ = player // Use player to avoid unused variable error
	return nil, api.ErrNotFound
}

// AbandonMission abandons an active mission
func (s *GameServer) AbandonMission(ctx context.Context, missionID uuid.UUID) error {
	// TODO: Integrate with missions manager
	// For now, return error as missions system needs manager integration
	// When implemented, this should:
	// 1. Verify mission exists and is active
	// 2. Update mission status to "failed" or remove it
	// 3. Apply any reputation penalties
	// 4. Clean up mission-related state

	// Placeholder: In production, use missions manager
	// err := s.missionsManager.AbandonMission(missionID)
	// return err

	return api.ErrNotFound
}

// GetActiveMissions retrieves player's active missions
func (s *GameServer) GetActiveMissions(ctx context.Context, playerID uuid.UUID) (*api.MissionList, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// TODO: Integrate with missions manager
	// For now, return empty list as missions system needs manager integration
	// When implemented, this should:
	// 1. Query player's active missions
	// 2. Include progress information
	// 3. Convert to API format using convertMissionToAPI

	missions := &api.MissionList{
		Missions:   make([]*api.Mission, 0),
		TotalCount: 0,
	}

	// Placeholder: In production, query from missions manager
	// activeMissions := s.missionsManager.GetActiveMissions(player.ID)

	_ = player // Use player to avoid unused variable error
	return missions, nil
}

// GetAvailableQuests retrieves quests available to player
func (s *GameServer) GetAvailableQuests(ctx context.Context, playerID uuid.UUID) (*api.QuestList, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// TODO: Integrate with quests manager
	// For now, return empty list as quests system needs manager integration
	// When implemented, this should:
	// 1. Query available quests at current location
	// 2. Filter by player level and quest prerequisites
	// 3. Exclude locked/completed quests
	// 4. Convert to API format using convertQuestToAPI

	quests := &api.QuestList{
		Quests:     make([]*api.Quest, 0),
		TotalCount: 0,
	}

	// Placeholder: In production, query from quests manager
	// availableQuests := s.questsManager.GetAvailableQuests(player.ID, player.CurrentSystem, player.Level)

	_ = player // Use player to avoid unused variable error
	return quests, nil
}

// AcceptQuest accepts a quest
func (s *GameServer) AcceptQuest(ctx context.Context, req *api.QuestAcceptRequest) (*api.Quest, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// TODO: Integrate with quests manager
	// For now, return error as quests system needs manager integration
	// When implemented, this should:
	// 1. Verify quest exists and is available
	// 2. Check if player meets prerequisites
	// 3. Add quest to player's active quests
	// 4. Initialize quest objectives
	// 5. Update quest status to "active"
	// 6. Return updated quest with objectives

	// Placeholder: In production, use quests manager
	// quest, err := s.questsManager.AcceptQuest(req.PlayerID, req.QuestID)
	// if err != nil {
	//     return nil, err
	// }
	// return convertQuestToAPI(quest), nil

	_ = player // Use player to avoid unused variable error
	return nil, api.ErrNotFound
}

// GetActiveQuests retrieves player's active quests
func (s *GameServer) GetActiveQuests(ctx context.Context, playerID uuid.UUID) (*api.QuestList, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// TODO: Integrate with quests manager
	// For now, return empty list as quests system needs manager integration
	// When implemented, this should:
	// 1. Query player's active quests
	// 2. Include all objectives and progress
	// 3. Include rewards information
	// 4. Convert to API format using convertQuestToAPI

	quests := &api.QuestList{
		Quests:     make([]*api.Quest, 0),
		TotalCount: 0,
	}

	// Placeholder: In production, query from quests manager
	// activeQuests := s.questsManager.GetActiveQuests(player.ID)

	_ = player // Use player to avoid unused variable error
	return quests, nil
}
