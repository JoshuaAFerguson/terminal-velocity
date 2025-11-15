// File: internal/api/server/server.go
// Project: Terminal Velocity
// Description: In-process API server implementation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/missions"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/quests"
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
	// NOTE: These are in-memory managers designed for single-user TUI sessions.
	// In Phase 2, mission/quest state should be persisted to database.
	// For Phase 1, we use these managers to integrate with existing systems.
	missionMgr *missions.Manager
	questMgr   *quests.Manager

	// Session management
	sessions *SessionManager

	// Database connection for transactions
	db *database.DB
}

// NewGameServer creates a new in-process game server
func NewGameServer(config *Config) (*GameServer, error) {
	server := &GameServer{
		playerRepo: config.PlayerRepo,
		systemRepo: config.SystemRepo,
		shipRepo:   config.ShipRepo,
		marketRepo: config.MarketRepo,
		sshKeyRepo: config.SSHKeyRepo,
		missionMgr: config.MissionMgr,
		questMgr:   config.QuestMgr,
		sessions:   NewSessionManager(),
		db:         config.DB,
	}

	return server, nil
}

// Config for the game server
type Config struct {
	// Database connection
	DB *database.DB

	// Database repositories
	PlayerRepo *database.PlayerRepository
	SystemRepo *database.SystemRepository
	ShipRepo   *database.ShipRepository
	MarketRepo *database.MarketRepository
	SSHKeyRepo *database.SSHKeyRepository

	// Manager packages
	// NOTE: In Phase 2, these should be replaced with database-backed state
	MissionMgr *missions.Manager
	QuestMgr   *quests.Manager
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

	// Get system data for government ID
	system, err := s.systemRepo.GetSystemByID(ctx, systemID)
	if err != nil {
		return nil, err
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

	// Convert to API market with government ID for illegal commodity checking
	// Note: This aggregates prices from all planets in the system
	// In a real implementation, you might want to show the best prices or average prices
	market := convertMarketToAPI(prices, commodities, system.GovernmentID)
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

	// Check cargo space
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "invalid ship type",
		}, nil
	}

	if !ship.CanAddCargo(int(req.Quantity), shipType) {
		cargoUsed := ship.GetCargoUsed()
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient cargo space (used: %d/%d, need: %d more)",
				cargoUsed, shipType.CargoSpace, req.Quantity),
		}, nil
	}

	// Perform the trade atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Deduct credits from player
		_, err := tx.ExecContext(ctx,
			"UPDATE players SET credits = credits - $1 WHERE id = $2",
			totalCost, req.PlayerID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// 2. Add cargo to ship
		ship.AddCargo(req.CommodityID, int(req.Quantity))
		cargoJSON, err := json.Marshal(ship.Cargo)
		if err != nil {
			return fmt.Errorf("failed to serialize cargo: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"UPDATE ships SET cargo = $1 WHERE id = $2",
			cargoJSON, player.ShipID)
		if err != nil {
			return fmt.Errorf("failed to update ship cargo: %w", err)
		}

		// 3. Update market stock
		_, err = tx.ExecContext(ctx,
			"UPDATE market_prices SET stock = stock - $1 WHERE planet_id = $2 AND commodity_id = $3",
			req.Quantity, *player.CurrentPlanet, req.CommodityID)
		if err != nil {
			return fmt.Errorf("failed to update market stock: %w", err)
		}

		// Update local state
		player.AddCredits(-totalCost)

		return nil
	})

	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("trade failed: %v", err),
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

	// Perform the trade atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Add credits to player
		_, err := tx.ExecContext(ctx,
			"UPDATE players SET credits = credits + $1 WHERE id = $2",
			totalPayment, req.PlayerID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// 2. Remove cargo from ship
		if !ship.RemoveCargo(req.CommodityID, int(req.Quantity)) {
			return fmt.Errorf("failed to remove cargo from ship")
		}
		cargoJSON, err := json.Marshal(ship.Cargo)
		if err != nil {
			return fmt.Errorf("failed to serialize cargo: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"UPDATE ships SET cargo = $1 WHERE id = $2",
			cargoJSON, player.ShipID)
		if err != nil {
			return fmt.Errorf("failed to update ship cargo: %w", err)
		}

		// 3. Update market stock (increase stock, decrease demand)
		_, err = tx.ExecContext(ctx,
			"UPDATE market_prices SET stock = stock + $1 WHERE planet_id = $2 AND commodity_id = $3",
			req.Quantity, *player.CurrentPlanet, req.CommodityID)
		if err != nil {
			return fmt.Errorf("failed to update market stock: %w", err)
		}

		// Update local state
		player.AddCredits(totalPayment)

		return nil
	})

	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: fmt.Sprintf("trade failed: %v", err),
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
func (s *GameServer) BuyShip(ctx context.Context, req *api.ShipPurchaseRequest) (*api.ShipPurchaseResponse, error) {
	if req.PlayerID == uuid.Nil || req.ShipType == "" {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "invalid request: missing player or ship type",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Must be docked to buy ships
	if !player.IsDocked() {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "must be docked at a planet to purchase ships",
		}, nil
	}

	// Get ship type definition
	shipType := models.GetShipTypeByID(req.ShipType)
	if shipType == nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("ship type not found: %s", req.ShipType),
		}, nil
	}

	// Calculate cost with optional trade-in
	totalCost := shipType.Price
	tradeInValue := int64(0)

	if req.TradeInShipID != nil {
		// Cannot trade in current ship
		if *req.TradeInShipID == player.ShipID {
			return &api.ShipPurchaseResponse{
				Success: false,
				Message: "cannot trade in your current ship",
			}, nil
		}

		// Load trade-in ship
		tradeInShip, err := s.shipRepo.GetByID(ctx, *req.TradeInShipID)
		if err != nil {
			return &api.ShipPurchaseResponse{
				Success: false,
				Message: "trade-in ship not found",
			}, nil
		}

		// Verify ownership
		if tradeInShip.OwnerID != player.ID {
			return &api.ShipPurchaseResponse{
				Success: false,
				Message: "you don't own that ship",
			}, nil
		}

		// Get trade-in ship type for pricing
		tradeInShipType := models.GetShipTypeByID(tradeInShip.TypeID)
		if tradeInShipType != nil {
			// Trade-in value is 70% of base price (accounting for depreciation)
			tradeInValue = int64(float64(tradeInShipType.Price) * 0.7)
			totalCost -= tradeInValue
		}
	}

	// Check credits
	if !player.CanAfford(totalCost) {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient credits (need: %d, have: %d)", totalCost, player.Credits),
		}, nil
	}

	// Create new ship
	newShip := &models.Ship{
		ID:      uuid.New(),
		OwnerID: player.ID,
		TypeID:  shipType.ID,
		Name:    shipType.Name, // Default name is ship type
		Hull:    shipType.MaxHull,
		Shields: shipType.MaxShields,
		Fuel:    shipType.MaxFuel,
		Cargo:   make([]models.CargoItem, 0),
		Crew:    1, // Default crew
		Weapons: make([]string, 0),
		Outfits: make([]string, 0),
	}

	// Perform ship purchase atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Create the ship in database
		cargoJSON, _ := json.Marshal(newShip.Cargo)
		weaponsJSON, _ := json.Marshal(newShip.Weapons)
		outfitsJSON, _ := json.Marshal(newShip.Outfits)

		_, err := tx.ExecContext(ctx,
			`INSERT INTO ships (id, owner_id, type_id, name, hull, shields, fuel, cargo, crew, weapons, outfits)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			newShip.ID, newShip.OwnerID, newShip.TypeID, newShip.Name, newShip.Hull,
			newShip.Shields, newShip.Fuel, cargoJSON, newShip.Crew, weaponsJSON, outfitsJSON)
		if err != nil {
			return fmt.Errorf("failed to create ship: %w", err)
		}

		// 2. Deduct credits from player
		_, err = tx.ExecContext(ctx,
			"UPDATE players SET credits = credits - $1 WHERE id = $2",
			totalCost, player.ID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// 3. Delete trade-in ship if provided
		if req.TradeInShipID != nil {
			_, err = tx.ExecContext(ctx,
				"DELETE FROM ships WHERE id = $1",
				*req.TradeInShipID)
			if err != nil {
				return fmt.Errorf("failed to delete trade-in ship: %w", err)
			}
		}

		// Update local state
		player.AddCredits(-totalCost)

		return nil
	})

	if err != nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("purchase failed: %v", err),
		}, nil
	}

	// Return success
	return &api.ShipPurchaseResponse{
		Success:      true,
		Message:      "ship purchased successfully",
		NewShip:      convertShipToAPI(newShip),
		TotalCost:    totalCost,
		TradeInValue: tradeInValue,
		NewState:     convertPlayerToAPI(player, newShip),
	}, nil
}


// SellShip sells a ship
func (s *GameServer) SellShip(ctx context.Context, req *api.ShipSaleRequest) (*api.ShipSaleResponse, error) {
	if req.PlayerID == uuid.Nil || req.ShipID == uuid.Nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "invalid request: missing player or ship ID",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Cannot sell current ship
	if req.ShipID == player.ShipID {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "cannot sell your current ship - switch ships first",
		}, nil
	}

	// Must be docked to sell ships
	if !player.IsDocked() {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "must be docked at a planet to sell ships",
		}, nil
	}

	// Load ship
	ship, err := s.shipRepo.GetByID(ctx, req.ShipID)
	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Verify ownership
	if ship.OwnerID != player.ID {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "you don't own that ship",
		}, nil
	}

	// Get ship type for pricing
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "ship type not found",
		}, nil
	}

	// Calculate sale value (60% of base price with depreciation)
	saleValue := int64(float64(shipType.Price) * 0.6)

	// Perform ship sale atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Delete ship
		_, err := tx.ExecContext(ctx,
			"DELETE FROM ships WHERE id = $1",
			req.ShipID)
		if err != nil {
			return fmt.Errorf("failed to delete ship: %w", err)
		}

		// 2. Add credits to player
		_, err = tx.ExecContext(ctx,
			"UPDATE players SET credits = credits + $1 WHERE id = $2",
			saleValue, player.ID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// Update local state
		player.AddCredits(saleValue)

		return nil
	})

	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: fmt.Sprintf("sale failed: %v", err),
		}, nil
	}

	// Load current ship for state
	currentShip, _ := s.shipRepo.GetByID(ctx, player.ShipID)

	// Return success
	return &api.ShipSaleResponse{
		Success:   true,
		Message:   "ship sold successfully",
		SaleValue: saleValue,
		NewState:  convertPlayerToAPI(player, currentShip),
	}, nil
}


// BuyOutfit purchases ship equipment
func (s *GameServer) BuyOutfit(ctx context.Context, req *api.OutfitPurchaseRequest) (*api.OutfitPurchaseResponse, error) {
	if req.PlayerID == uuid.Nil || req.OutfitID == "" || req.Quantity <= 0 {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "invalid request: missing player, outfit ID, or quantity",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Must be docked to buy outfits
	if !player.IsDocked() {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "must be docked at a planet to purchase outfits",
		}, nil
	}

	// Get outfit definition
	outfit := models.GetOutfitByID(req.OutfitID)
	if outfit == nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("outfit not found: %s", req.OutfitID),
		}, nil
	}

	// Calculate total cost
	totalCost := outfit.Price * int64(req.Quantity)

	// Check credits
	if !player.CanAfford(totalCost) {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient credits (need: %d, have: %d)", totalCost, player.Credits),
		}, nil
	}

	// Load current ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Get ship type for outfit space calculations
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "invalid ship type",
		}, nil
	}

	// Check outfit space constraints
	totalOutfitSpaceNeeded := outfit.OutfitSpace * int(req.Quantity)
	spaceAvailable := ship.GetOutfitSpaceAvailable(shipType)

	if spaceAvailable < totalOutfitSpaceNeeded {
		spaceUsed := ship.GetOutfitSpaceUsed()
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient outfit space (used: %d/%d, need: %d more)",
				spaceUsed, shipType.OutfitSpace, totalOutfitSpaceNeeded),
		}, nil
	}

	// Add outfits to ship
	for i := 0; i < int(req.Quantity); i++ {
		ship.Outfits = append(ship.Outfits, outfit.ID)
	}

	// Perform outfit purchase atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Update ship with new outfits
		outfitsJSON, err := json.Marshal(ship.Outfits)
		if err != nil {
			return fmt.Errorf("failed to serialize outfits: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"UPDATE ships SET outfits = $1 WHERE id = $2",
			outfitsJSON, player.ShipID)
		if err != nil {
			return fmt.Errorf("failed to update ship: %w", err)
		}

		// 2. Deduct credits from player
		_, err = tx.ExecContext(ctx,
			"UPDATE players SET credits = credits - $1 WHERE id = $2",
			totalCost, player.ID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// Update local state
		player.AddCredits(-totalCost)

		return nil
	})

	if err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: fmt.Sprintf("purchase failed: %v", err),
		}, nil
	}

	// Return success
	return &api.OutfitPurchaseResponse{
		Success:     true,
		Message:     "outfit(s) purchased successfully",
		Outfits:     make([]*api.Outfit, 0), // TODO: Convert outfit to API format
		TotalCost:   totalCost,
		UpdatedShip: convertShipToAPI(ship),
	}, nil
}


// SellOutfit sells ship equipment
func (s *GameServer) SellOutfit(ctx context.Context, req *api.OutfitSaleRequest) (*api.OutfitSaleResponse, error) {
	if req.PlayerID == uuid.Nil || req.OutfitID == "" || req.Quantity <= 0 {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "invalid request: missing player, outfit ID, or quantity",
		}, nil
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "player not found",
		}, nil
	}

	// Must be docked to sell outfits
	if !player.IsDocked() {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "must be docked at a planet to sell outfits",
		}, nil
	}

	// Get outfit definition
	outfit := models.GetOutfitByID(req.OutfitID)
	if outfit == nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: fmt.Sprintf("outfit not found: %s", req.OutfitID),
		}, nil
	}

	// Load current ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "ship not found",
		}, nil
	}

	// Count how many of this outfit the ship has
	count := 0
	for _, outfitID := range ship.Outfits {
		if outfitID == req.OutfitID {
			count++
		}
	}

	// Check if player has enough outfits
	if count < int(req.Quantity) {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: fmt.Sprintf("insufficient outfits (have: %d, trying to sell: %d)", count, req.Quantity),
		}, nil
	}

	// Remove outfits from ship
	removed := 0
	newOutfits := make([]string, 0, len(ship.Outfits))
	for _, outfitID := range ship.Outfits {
		if outfitID == req.OutfitID && removed < int(req.Quantity) {
			removed++
			continue // Skip this one (remove it)
		}
		newOutfits = append(newOutfits, outfitID)
	}
	ship.Outfits = newOutfits

	// Calculate sale value (60% of purchase price)
	saleValue := int64(float64(outfit.Price) * 0.6 * float64(req.Quantity))

	// Perform outfit sale atomically within a transaction
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Update ship (remove outfits)
		outfitsJSON, err := json.Marshal(ship.Outfits)
		if err != nil {
			return fmt.Errorf("failed to serialize outfits: %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"UPDATE ships SET outfits = $1 WHERE id = $2",
			outfitsJSON, player.ShipID)
		if err != nil {
			return fmt.Errorf("failed to update ship: %w", err)
		}

		// 2. Add credits to player
		_, err = tx.ExecContext(ctx,
			"UPDATE players SET credits = credits + $1 WHERE id = $2",
			saleValue, player.ID)
		if err != nil {
			return fmt.Errorf("failed to update player credits: %w", err)
		}

		// Update local state
		player.AddCredits(saleValue)

		return nil
	})

	if err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: fmt.Sprintf("sale failed: %v", err),
		}, nil
	}

	// Return success
	return &api.OutfitSaleResponse{
		Success:     true,
		Message:     "outfit(s) sold successfully",
		SaleValue:   saleValue,
		UpdatedShip: convertShipToAPI(ship),
	}, nil
}


// GetAvailableMissions retrieves missions available to player
func (s *GameServer) GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*api.MissionList, error) {
	if playerID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player to check current location
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Must be docked to access missions
	if !player.IsDocked() {
		return &api.MissionList{
			Missions:   make([]*api.Mission, 0),
			TotalCount: 0,
		}, nil
	}

	// Get available missions from manager
	// NOTE: In Phase 2, this should query from database per-player
	// Current implementation uses in-memory manager shared across all players
	availableMissions := s.missionMgr.GetAvailableMissions()

	// Convert to API format
	apiMissions := make([]*api.Mission, 0, len(availableMissions))
	for _, mission := range availableMissions {
		apiMissions = append(apiMissions, convertMissionToAPI(mission))
	}

	return &api.MissionList{
		Missions:   apiMissions,
		TotalCount: int32(len(apiMissions)),
	}, nil
}

// AcceptMission accepts a mission
func (s *GameServer) AcceptMission(ctx context.Context, req *api.MissionAcceptRequest) (*api.Mission, error) {
	if req.PlayerID == uuid.Nil || req.MissionID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// Must be docked to accept missions
	if !player.IsDocked() {
		return nil, fmt.Errorf("must be docked to accept missions")
	}

	// Load player's ship
	ship, err := s.shipRepo.GetByID(ctx, player.ShipID)
	if err != nil {
		return nil, err
	}

	// Get ship type for cargo capacity checks
	shipType := models.GetShipTypeByID(ship.TypeID)
	if shipType == nil {
		return nil, fmt.Errorf("ship type not found")
	}

	// Accept mission through manager
	// NOTE: Manager modifies ship cargo for delivery missions
	err = s.missionMgr.AcceptMission(req.MissionID, player, ship, shipType)
	if err != nil {
		return nil, fmt.Errorf("failed to accept mission: %w", err)
	}

	// Persist ship changes (delivery missions add cargo)
	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return nil, fmt.Errorf("failed to update ship: %w", err)
	}

	// Get the accepted mission
	mission := s.missionMgr.GetMissionByID(req.MissionID)
	if mission == nil {
		return nil, api.ErrNotFound
	}

	return convertMissionToAPI(mission), nil
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
	if playerID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player
	_, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Get active missions from manager
	// NOTE: Current manager doesn't track per-player missions.
	// In Phase 2, need database table for player_missions
	activeMissions := s.missionMgr.GetActiveMissions()

	// Convert to API format
	apiMissions := make([]*api.Mission, 0, len(activeMissions))
	for _, mission := range activeMissions {
		apiMissions = append(apiMissions, convertMissionToAPI(mission))
	}

	return &api.MissionList{
		Missions:   apiMissions,
		TotalCount: int32(len(apiMissions)),
	}, nil
}

// GetAvailableQuests retrieves quests available to player
func (s *GameServer) GetAvailableQuests(ctx context.Context, playerID uuid.UUID) (*api.QuestList, error) {
	if playerID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player
	_, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Get available quests from manager
	// Manager filters by prerequisites and player progress
	availableQuests := s.questMgr.GetAvailableQuests(playerID)

	// Convert to API format
	apiQuests := make([]*api.Quest, 0, len(availableQuests))
	for _, quest := range availableQuests {
		apiQuests = append(apiQuests, convertQuestToAPI(quest))
	}

	return &api.QuestList{
		Quests:     apiQuests,
		TotalCount: int32(len(apiQuests)),
	}, nil
}

// AcceptQuest accepts a quest
func (s *GameServer) AcceptQuest(ctx context.Context, req *api.QuestAcceptRequest) (*api.Quest, error) {
	if req.PlayerID == uuid.Nil || req.QuestID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player
	_, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// Convert QuestID from UUID to string (Quest model uses string IDs)
	questID := req.QuestID.String()

	// Start quest through manager
	// Manager handles prerequisites validation and objective initialization
	playerQuest, err := s.questMgr.StartQuest(req.PlayerID, questID)
	if err != nil {
		return nil, fmt.Errorf("failed to start quest: %w", err)
	}

	// Get the quest definition for full info
	quest := s.questMgr.GetQuest(questID)
	if quest == nil {
		return nil, api.ErrNotFound
	}

	// NOTE: playerQuest has progress info, but API Quest doesn't include progress
	// In Phase 2, consider adding progress fields to API Quest
	_ = playerQuest

	return convertQuestToAPI(quest), nil
}

// GetActiveQuests retrieves player's active quests
func (s *GameServer) GetActiveQuests(ctx context.Context, playerID uuid.UUID) (*api.QuestList, error) {
	if playerID == uuid.Nil {
		return nil, api.ErrInvalidRequest
	}

	// Load player
	_, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Get active quests from manager
	// Returns PlayerQuest objects with progress info
	activePlayerQuests := s.questMgr.GetActiveQuests(playerID)

	// Convert to API format
	// NOTE: Need to fetch Quest definitions to get full info
	apiQuests := make([]*api.Quest, 0, len(activePlayerQuests))
	for _, playerQuest := range activePlayerQuests {
		// Get quest definition
		quest := s.questMgr.GetQuest(playerQuest.QuestID)
		if quest != nil {
			apiQuest := convertQuestToAPI(quest)
			// TODO: In Phase 2, add progress info from playerQuest to apiQuest
			_ = playerQuest // Has ObjectiveProgress, CurrentObjective, etc.
			apiQuests = append(apiQuests, apiQuest)
		}
	}

	return &api.QuestList{
		Quests:     apiQuests,
		TotalCount: int32(len(apiQuests)),
	}, nil
}
