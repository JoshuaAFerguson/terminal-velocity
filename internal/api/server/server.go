// File: internal/api/server/server.go
// Project: Terminal Velocity
// Description: In-process API server implementation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"context"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
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
	// TODO: Implement authentication logic
	// This will use existing password validation from internal/server
	return nil, api.ErrNotFound
}

// AuthenticateSSH validates SSH public key and returns session token
func (s *GameServer) AuthenticateSSH(ctx context.Context, req *api.SSHAuthRequest) (*api.AuthResponse, error) {
	// TODO: Implement SSH authentication
	// This will use s.sshKeyRepo to validate public key
	return nil, api.ErrNotFound
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
	session, err := s.sessions.GetSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	// TODO: Validate token
	return session, nil
}

// EndSession terminates an active session
func (s *GameServer) EndSession(ctx context.Context, req *api.EndSessionRequest) error {
	return s.sessions.EndSession(req.SessionID)
}

// RefreshSession extends a session's lifetime
func (s *GameServer) RefreshSession(ctx context.Context, req *api.RefreshSessionRequest) (*api.AuthResponse, error) {
	// TODO: Implement session refresh
	return nil, api.ErrNotFound
}

// Register creates a new player account
func (s *GameServer) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	// TODO: Implement registration
	// This will use s.playerRepo to create new player
	return nil, api.ErrNotFound
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
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return nil, err
	}

	// Convert to API types
	state := convertPlayerToAPI(player, ship)
	state.Stats = convertPlayerStatsToAPI(player)
	state.Reputation = convertReputationToAPI(player)

	return state, nil
}

// UpdatePlayerLocation updates player's location
func (s *GameServer) UpdatePlayerLocation(ctx context.Context, req *api.LocationUpdate) (*api.PlayerState, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return nil, err
	}

	// Update location
	player.CurrentSystemID = req.SystemID
	player.CurrentPlanetID = req.PlanetID
	player.X = req.Position.X
	player.Y = req.Position.Y

	// Persist changes
	if err := s.playerRepo.Update(ctx, player); err != nil {
		return nil, err
	}

	// Load ship for complete state
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return nil, err
	}

	// Get updated player state
	state := convertPlayerToAPI(player, ship)
	state.Stats = convertPlayerStatsToAPI(player)
	state.Reputation = convertReputationToAPI(player)

	return state, nil
}

// GetPlayerShip retrieves player's current ship
func (s *GameServer) GetPlayerShip(ctx context.Context, playerID uuid.UUID) (*api.Ship, error) {
	// Load player to get current ship ID
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Load the ship
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
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
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
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
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Load player's ship
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// Validate player is in space
	if player.CurrentPlanetID != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "You must take off before jumping to another system",
		}, nil
	}

	// Validate not jumping to same system
	if player.CurrentSystemID == req.TargetSystemID {
		return &api.JumpResponse{
			Success: false,
			Message: "Already in target system",
		}, nil
	}

	// Verify systems are connected
	connections, err := s.systemRepo.GetConnections(ctx, player.CurrentSystemID)
	if err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "Failed to load jump routes",
		}, nil
	}

	isConnected := false
	for _, connectedSystemID := range connections {
		if connectedSystemID == req.TargetSystemID {
			isConnected = true
			break
		}
	}

	if !isConnected {
		return &api.JumpResponse{
			Success: false,
			Message: "No jump route to target system",
		}, nil
	}

	// Calculate fuel cost (simplified - could be distance-based)
	const fuelCostPerJump = 10
	if ship.Fuel < fuelCostPerJump {
		return &api.JumpResponse{
			Success: false,
			Message: "Insufficient fuel for jump",
		}, nil
	}

	// Perform jump
	player.CurrentSystemID = req.TargetSystemID
	player.X = 0 // Reset to system center
	player.Y = 0
	ship.Fuel -= fuelCostPerJump

	// Update player stats
	player.JumpsMade++

	// Check if this is a new system
	// TODO: Track visited systems and increment SystemsVisited if new

	// Update database
	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "Failed to update player location",
		}, nil
	}

	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return &api.JumpResponse{
			Success: false,
			Message: "Failed to update ship fuel",
		}, nil
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, ship)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.JumpResponse{
		Success:      true,
		Message:      "Jump successful",
		NewState:     newState,
		FuelConsumed: fuelCostPerJump,
	}, nil
}

// Land lands on a planet
func (s *GameServer) Land(ctx context.Context, req *api.LandRequest) (*api.LandResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is in space
	if player.CurrentPlanetID != nil {
		return &api.LandResponse{
			Success: false,
			Message: "Already docked at a planet",
		}, nil
	}

	// Load planet to verify it exists and is in current system
	planet, err := s.systemRepo.GetPlanetByID(ctx, req.PlanetID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "Planet not found",
		}, nil
	}

	// Validate planet is in player's current system
	if planet.SystemID != player.CurrentSystemID {
		return &api.LandResponse{
			Success: false,
			Message: "Planet not in current system",
		}, nil
	}

	// Update player location
	player.CurrentPlanetID = &req.PlanetID

	// Persist changes
	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "Failed to update player location",
		}, nil
	}

	// Load ship for complete state
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.LandResponse{
			Success: false,
			Message: "Failed to load ship data",
		}, nil
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, ship)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.LandResponse{
		Success:  true,
		Message:  "Landing successful",
		Planet:   convertPlanetToAPI(planet),
		NewState: newState,
	}, nil
}

// Takeoff takes off from a planet
func (s *GameServer) Takeoff(ctx context.Context, req *api.TakeoffRequest) (*api.TakeoffResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "Already in space",
		}, nil
	}

	// Clear planet location (now in space)
	player.CurrentPlanetID = nil

	// Persist changes
	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "Failed to update player location",
		}, nil
	}

	// Load ship for complete state
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.TakeoffResponse{
			Success: false,
			Message: "Failed to load ship data",
		}, nil
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, ship)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.TakeoffResponse{
		Success:  true,
		Message:  "Takeoff successful",
		NewState: newState,
	}, nil
}

// GetMarket retrieves market data for a system
func (s *GameServer) GetMarket(ctx context.Context, systemID uuid.UUID) (*api.Market, error) {
	// Get all commodities for the system
	commodities, err := s.marketRepo.GetCommoditiesBySystemID(ctx, systemID)
	if err != nil {
		return nil, err
	}

	// Convert to API market format
	// TODO: Get actual last updated timestamp from database
	market := convertMarketToAPI(systemID.String(), commodities, "")
	market.SystemID = systemID

	return market, nil
}

// BuyCommodity purchases a commodity from the market
func (s *GameServer) BuyCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
	// Load player and ship
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "You must be docked at a planet to trade",
		}, nil
	}

	// Get market data for current system
	commodities, err := s.marketRepo.GetCommoditiesBySystemID(ctx, player.CurrentSystemID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Market data unavailable",
		}, nil
	}

	// Find the commodity
	var commodity *models.CommodityListing
	for i := range commodities {
		if commodities[i].CommodityID == req.CommodityID {
			commodity = &commodities[i]
			break
		}
	}

	if commodity == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Commodity not available at this market",
		}, nil
	}

	// Validate stock
	if commodity.Stock < req.Quantity {
		return &api.TradeResponse{
			Success: false,
			Message: "Insufficient stock available",
		}, nil
	}

	// Calculate total cost
	totalCost := int64(commodity.BuyPrice) * int64(req.Quantity)

	// Validate credits
	if player.Credits < totalCost {
		return &api.TradeResponse{
			Success: false,
			Message: "Insufficient credits",
		}, nil
	}

	// Validate cargo space
	cargoAvailable := ship.CargoSpace - ship.CargoUsed
	if cargoAvailable < req.Quantity {
		return &api.TradeResponse{
			Success: false,
			Message: "Insufficient cargo space",
		}, nil
	}

	// Update ship cargo
	if ship.Cargo == nil {
		ship.Cargo = make(map[string]int)
	}
	ship.Cargo[req.CommodityID] += int(req.Quantity)
	ship.CargoUsed += req.Quantity

	// Update player credits
	player.Credits -= totalCost

	// Update database
	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Failed to update ship cargo",
		}, nil
	}

	if err := s.playerRepo.Update(ctx, player); err != nil {
		// Rollback ship update would be ideal here
		// For now, log the error
		return &api.TradeResponse{
			Success: false,
			Message: "Failed to update player credits",
		}, nil
	}

	// Update market stock
	if err := s.marketRepo.UpdateStock(ctx, player.CurrentSystemID, req.CommodityID, -int(req.Quantity)); err != nil {
		// Stock update failed, but transaction succeeded
		// Continue anyway as this is non-critical
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, ship)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.TradeResponse{
		Success:        true,
		Message:        "Purchase successful",
		QuantityTraded: req.Quantity,
		TotalCost:      totalCost,
		PricePerUnit:   commodity.BuyPrice,
		NewState:       newState,
	}, nil
}

// SellCommodity sells a commodity to the market
func (s *GameServer) SellCommodity(ctx context.Context, req *api.TradeRequest) (*api.TradeResponse, error) {
	// Load player and ship
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "You must be docked at a planet to trade",
		}, nil
	}

	// Get market data for current system
	commodities, err := s.marketRepo.GetCommoditiesBySystemID(ctx, player.CurrentSystemID)
	if err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Market data unavailable",
		}, nil
	}

	// Find the commodity
	var commodity *models.CommodityListing
	for i := range commodities {
		if commodities[i].CommodityID == req.CommodityID {
			commodity = &commodities[i]
			break
		}
	}

	if commodity == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Commodity not available at this market",
		}, nil
	}

	// Validate player has the commodity
	if ship.Cargo == nil {
		return &api.TradeResponse{
			Success: false,
			Message: "You don't have any cargo",
		}, nil
	}

	cargoQuantity, hasCargo := ship.Cargo[req.CommodityID]
	if !hasCargo || cargoQuantity < int(req.Quantity) {
		return &api.TradeResponse{
			Success: false,
			Message: "You don't have enough of this commodity",
		}, nil
	}

	// Calculate total sale value
	totalValue := int64(commodity.SellPrice) * int64(req.Quantity)

	// Update ship cargo
	ship.Cargo[req.CommodityID] -= int(req.Quantity)
	if ship.Cargo[req.CommodityID] == 0 {
		delete(ship.Cargo, req.CommodityID)
	}
	ship.CargoUsed -= req.Quantity

	// Update player credits
	player.Credits += totalValue

	// Update database
	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Failed to update ship cargo",
		}, nil
	}

	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.TradeResponse{
			Success: false,
			Message: "Failed to update player credits",
		}, nil
	}

	// Update market stock (selling increases market stock)
	if err := s.marketRepo.UpdateStock(ctx, player.CurrentSystemID, req.CommodityID, int(req.Quantity)); err != nil {
		// Stock update failed, but transaction succeeded
		// Continue anyway as this is non-critical
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, ship)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.TradeResponse{
		Success:        true,
		Message:        "Sale successful",
		QuantityTraded: req.Quantity,
		TotalCost:      totalValue,
		PricePerUnit:   commodity.SellPrice,
		NewState:       newState,
	}, nil
}

// BuyShip purchases a new ship
func (s *GameServer) BuyShip(ctx context.Context, req *api.ShipPurchaseRequest) (*api.ShipPurchaseResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is docked (must be at shipyard)
	if player.CurrentPlanetID == nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "You must be docked at a planet to purchase a ship",
		}, nil
	}

	// TODO: Get ship type data from database to determine cost
	// For now, use placeholder costs based on ship type
	var shipCost int64
	switch req.ShipType {
	case "shuttle":
		shipCost = 10000
	case "fighter":
		shipCost = 50000
	case "freighter":
		shipCost = 100000
	case "corvette":
		shipCost = 200000
	case "frigate":
		shipCost = 500000
	default:
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "Unknown ship type",
		}, nil
	}

	// Handle trade-in
	var tradeInValue int64
	var oldShip *models.Ship
	if req.TradeInShipID != nil {
		oldShip, err = s.shipRepo.GetByID(ctx, *req.TradeInShipID)
		if err != nil {
			return &api.ShipPurchaseResponse{
				Success: false,
				Message: "Trade-in ship not found",
			}, nil
		}

		// Verify player owns the trade-in ship
		if oldShip.PlayerID != player.ID {
			return &api.ShipPurchaseResponse{
				Success: false,
				Message: "You don't own the trade-in ship",
			}, nil
		}

		// Trade-in value is 70% of current value
		tradeInValue = int64(float64(oldShip.CurrentValue) * 0.7)
	}

	// Calculate total cost
	totalCost := shipCost - tradeInValue

	// Validate credits
	if player.Credits < totalCost {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "Insufficient credits",
		}, nil
	}

	// Create new ship
	// TODO: Use ship template from database for proper stats
	newShip := &models.Ship{
		PlayerID:      player.ID,
		ShipType:      req.ShipType,
		Name:          req.ShipType, // Default name, player can rename later
		Hull:          100,
		MaxHull:       100,
		Shields:       50,
		MaxShields:    50,
		Fuel:          100,
		MaxFuel:       100,
		CargoSpace:    50,
		CargoUsed:     0,
		Cargo:         make(map[string]int),
		Speed:         10.0,
		Acceleration:  5.0,
		TurnRate:      5.0,
		PurchasePrice: shipCost,
		CurrentValue:  shipCost,
		Weapons:       make([]*models.Weapon, 0),
		Outfits:       make([]*models.Outfit, 0),
	}

	// Save new ship
	if err := s.shipRepo.Create(ctx, newShip); err != nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "Failed to create new ship",
		}, nil
	}

	// Update player credits and current ship
	player.Credits -= totalCost
	player.CurrentShipID = newShip.ID

	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.ShipPurchaseResponse{
			Success: false,
			Message: "Failed to update player",
		}, nil
	}

	// Delete old ship if traded in
	if oldShip != nil {
		if err := s.shipRepo.Delete(ctx, oldShip.ID); err != nil {
			// Non-fatal, ship purchase succeeded
			// Just log the error in production
		}
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, newShip)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.ShipPurchaseResponse{
		Success:      true,
		Message:      "Ship purchase successful",
		NewShip:      convertShipToAPI(newShip),
		TotalCost:    totalCost,
		TradeInValue: tradeInValue,
		NewState:     newState,
	}, nil
}

// SellShip sells a ship
func (s *GameServer) SellShip(ctx context.Context, req *api.ShipSaleRequest) (*api.ShipSaleResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "You must be docked at a planet to sell a ship",
		}, nil
	}

	// Load the ship to sell
	ship, err := s.shipRepo.GetByID(ctx, req.ShipID)
	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// Verify player owns the ship
	if ship.PlayerID != player.ID {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "You don't own this ship",
		}, nil
	}

	// Prevent selling current ship (must switch ships first)
	if ship.ID == player.CurrentShipID {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Cannot sell your current ship. Switch ships first.",
		}, nil
	}

	// Calculate sale value (75% of current value)
	saleValue := int64(float64(ship.CurrentValue) * 0.75)

	// Delete the ship
	if err := s.shipRepo.Delete(ctx, ship.ID); err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Failed to delete ship",
		}, nil
	}

	// Add credits to player
	player.Credits += saleValue

	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Failed to update player credits",
		}, nil
	}

	// Load current ship for state
	currentShip, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.ShipSaleResponse{
			Success: false,
			Message: "Failed to load current ship",
		}, nil
	}

	// Get updated player state
	newState := convertPlayerToAPI(player, currentShip)
	newState.Stats = convertPlayerStatsToAPI(player)
	newState.Reputation = convertReputationToAPI(player)

	return &api.ShipSaleResponse{
		Success:   true,
		Message:   "Ship sold successfully",
		SaleValue: saleValue,
		NewState:  newState,
	}, nil
}

// BuyOutfit purchases ship equipment
func (s *GameServer) BuyOutfit(ctx context.Context, req *api.OutfitPurchaseRequest) (*api.OutfitPurchaseResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "You must be docked at a planet to purchase outfits",
		}, nil
	}

	// Load current ship
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// TODO: Get outfit data from database
	// For now, use placeholder costs and stats
	var outfitCost int64
	var outfitName string
	var outfitDesc string
	modifiers := make(map[string]int)

	switch req.OutfitID {
	case "shield_booster":
		outfitCost = 5000
		outfitName = "Shield Booster"
		outfitDesc = "Increases shield capacity"
		modifiers["shields"] = 25
	case "cargo_expansion":
		outfitCost = 3000
		outfitName = "Cargo Expansion"
		outfitDesc = "Increases cargo capacity"
		modifiers["cargo"] = 20
	case "fuel_tank":
		outfitCost = 2000
		outfitName = "Fuel Tank"
		outfitDesc = "Increases fuel capacity"
		modifiers["fuel"] = 50
	default:
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Unknown outfit type",
		}, nil
	}

	// Calculate total cost for quantity
	totalCost := outfitCost * int64(req.Quantity)

	// Validate credits
	if player.Credits < totalCost {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Insufficient credits",
		}, nil
	}

	// Add outfits to ship
	newOutfits := make([]*api.Outfit, 0, int(req.Quantity))
	for i := int32(0); i < req.Quantity; i++ {
		outfit := &models.Outfit{
			OutfitType:  req.OutfitID,
			Name:        outfitName,
			Description: outfitDesc,
			Modifiers:   modifiers,
		}
		ship.Outfits = append(ship.Outfits, outfit)

		// Convert for response
		apiOutfit := &api.Outfit{
			OutfitID:    req.OutfitID,
			OutfitType:  req.OutfitID,
			Name:        outfitName,
			Description: outfitDesc,
			Modifiers:   make(map[string]int32),
		}
		for k, v := range modifiers {
			apiOutfit.Modifiers[k] = int32(v)
		}
		newOutfits = append(newOutfits, apiOutfit)
	}

	// Update ship in database
	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Failed to update ship",
		}, nil
	}

	// Deduct credits
	player.Credits -= totalCost

	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.OutfitPurchaseResponse{
			Success: false,
			Message: "Failed to update player credits",
		}, nil
	}

	return &api.OutfitPurchaseResponse{
		Success:     true,
		Message:     "Outfit purchase successful",
		Outfits:     newOutfits,
		TotalCost:   totalCost,
		UpdatedShip: convertShipToAPI(ship),
	}, nil
}

// SellOutfit sells ship equipment
func (s *GameServer) SellOutfit(ctx context.Context, req *api.OutfitSaleRequest) (*api.OutfitSaleResponse, error) {
	// Load player
	player, err := s.playerRepo.GetByID(ctx, req.PlayerID)
	if err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "Player not found",
		}, nil
	}

	// Validate player is docked
	if player.CurrentPlanetID == nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "You must be docked at a planet to sell outfits",
		}, nil
	}

	// Load current ship
	ship, err := s.shipRepo.GetByID(ctx, player.CurrentShipID)
	if err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "Ship not found",
		}, nil
	}

	// Find and remove outfits of the specified type
	removedCount := int32(0)
	var totalValue int64
	newOutfits := make([]*models.Outfit, 0)

	// TODO: Get actual outfit values from database
	var outfitValue int64
	switch req.OutfitID {
	case "shield_booster":
		outfitValue = 5000
	case "cargo_expansion":
		outfitValue = 3000
	case "fuel_tank":
		outfitValue = 2000
	default:
		outfitValue = 1000 // Default value for unknown outfits
	}

	// Remove specified quantity of outfits
	for _, outfit := range ship.Outfits {
		if outfit.OutfitType == req.OutfitID && removedCount < req.Quantity {
			// Skip this outfit (remove it)
			removedCount++
			// Sell for 60% of purchase price
			totalValue += int64(float64(outfitValue) * 0.6)
		} else {
			// Keep this outfit
			newOutfits = append(newOutfits, outfit)
		}
	}

	// Validate we had enough outfits to sell
	if removedCount < req.Quantity {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "You don't have enough of that outfit equipped",
		}, nil
	}

	// Update ship outfits
	ship.Outfits = newOutfits

	// Update ship in database
	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "Failed to update ship",
		}, nil
	}

	// Add credits to player
	player.Credits += totalValue

	if err := s.playerRepo.Update(ctx, player); err != nil {
		return &api.OutfitSaleResponse{
			Success: false,
			Message: "Failed to update player credits",
		}, nil
	}

	return &api.OutfitSaleResponse{
		Success:     true,
		Message:     "Outfit sold successfully",
		SaleValue:   totalValue,
		UpdatedShip: convertShipToAPI(ship),
	}, nil
}

// GetAvailableMissions retrieves missions available to player
func (s *GameServer) GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*api.MissionList, error) {
	// Load player to check current location and stats
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

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
	// missions := s.missionsManager.GetAvailableMissions(player.CurrentSystemID, player.Level, player.FactionReputation)

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
	// availableQuests := s.questsManager.GetAvailableQuests(player.ID, player.CurrentSystemID, player.Level)

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
