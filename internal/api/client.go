// File: internal/api/client.go
// Project: Terminal Velocity
// Description: API client interface for game server communication
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package api

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrGRPCNotImplemented is returned when gRPC mode is requested but not yet implemented
	ErrGRPCNotImplemented = errors.New("gRPC mode not yet implemented (Phase 2+)")
)

// Client provides a unified interface for communicating with the game server.
// This abstraction allows the TUI to work with either:
// - In-process calls (Phase 1: monolithic binary)
// - gRPC calls (Phase 2+: distributed services)
type Client interface {
	AuthClient
	PlayerClient
	GameClient

	// Close releases any resources held by the client
	Close() error
}

// AuthClient handles authentication and session management
type AuthClient interface {
	// Authenticate validates user credentials and returns a session token
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error)

	// AuthenticateSSH validates SSH public key and returns a session token
	AuthenticateSSH(ctx context.Context, req *SSHAuthRequest) (*AuthResponse, error)

	// CreateSession creates a new game session for an authenticated player
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error)

	// ValidateSession checks if a session token is still valid
	ValidateSession(ctx context.Context, req *ValidateSessionRequest) (*Session, error)

	// EndSession terminates an active session
	EndSession(ctx context.Context, req *EndSessionRequest) error

	// RefreshSession extends a session's lifetime
	RefreshSession(ctx context.Context, req *RefreshSessionRequest) (*AuthResponse, error)

	// Register creates a new player account
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
}

// PlayerClient manages player state and real-time updates
type PlayerClient interface {
	// GetPlayerState retrieves complete player state
	GetPlayerState(ctx context.Context, playerID uuid.UUID) (*PlayerState, error)

	// UpdatePlayerLocation updates the player's location
	UpdatePlayerLocation(ctx context.Context, req *LocationUpdate) (*PlayerState, error)

	// GetPlayerShip retrieves the player's current ship details
	GetPlayerShip(ctx context.Context, playerID uuid.UUID) (*Ship, error)

	// GetPlayerInventory retrieves the player's cargo and items
	GetPlayerInventory(ctx context.Context, playerID uuid.UUID) (*Inventory, error)

	// GetPlayerStats retrieves player statistics
	GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error)

	// GetPlayerReputation retrieves faction reputation
	GetPlayerReputation(ctx context.Context, playerID uuid.UUID) (*ReputationInfo, error)

	// StreamPlayerUpdates subscribes to real-time player state changes
	StreamPlayerUpdates(ctx context.Context, playerID uuid.UUID) (PlayerUpdateStream, error)
}

// PlayerUpdateStream represents a stream of player state updates
type PlayerUpdateStream interface {
	// Recv receives the next update from the stream
	Recv() (*PlayerUpdate, error)

	// Close closes the stream
	Close() error
}

// GameClient handles core game actions
type GameClient interface {
	// Navigation
	Jump(ctx context.Context, req *JumpRequest) (*JumpResponse, error)
	Land(ctx context.Context, req *LandRequest) (*LandResponse, error)
	Takeoff(ctx context.Context, req *TakeoffRequest) (*TakeoffResponse, error)

	// Trading
	GetMarket(ctx context.Context, systemID uuid.UUID) (*Market, error)
	BuyCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error)
	SellCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error)

	// Ship Management
	BuyShip(ctx context.Context, req *ShipPurchaseRequest) (*ShipPurchaseResponse, error)
	SellShip(ctx context.Context, req *ShipSaleRequest) (*ShipSaleResponse, error)
	BuyOutfit(ctx context.Context, req *OutfitPurchaseRequest) (*OutfitPurchaseResponse, error)
	SellOutfit(ctx context.Context, req *OutfitSaleRequest) (*OutfitSaleResponse, error)

	// Missions
	GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error)
	AcceptMission(ctx context.Context, req *MissionAcceptRequest) (*Mission, error)
	AbandonMission(ctx context.Context, missionID uuid.UUID) error
	GetActiveMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error)

	// Quests
	GetAvailableQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error)
	AcceptQuest(ctx context.Context, req *QuestAcceptRequest) (*Quest, error)
	GetActiveQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error)
}

// NewClient creates a new API client
// In Phase 1 (monolithic), this returns an in-process client
// In Phase 2+, this can return a gRPC client based on configuration
func NewClient(config *ClientConfig) (Client, error) {
	if config.Mode == ClientModeInProcess {
		return newInProcessClient(config)
	}
	// Phase 2+: return newGRPCClient(config)
	return nil, ErrGRPCNotImplemented
}

// ClientConfig configures the API client
type ClientConfig struct {
	Mode ClientMode

	// For in-process mode
	InProcessServer Server

	// For gRPC mode (Phase 2+)
	ServerAddress string
	TLSConfig     *TLSConfig
}

// ClientMode specifies how the client communicates with the server
type ClientMode string

const (
	// ClientModeInProcess uses direct function calls (Phase 1)
	ClientModeInProcess ClientMode = "in_process"

	// ClientModeGRPC uses gRPC network calls (Phase 2+)
	ClientModeGRPC ClientMode = "grpc"
)

// TLSConfig for secure gRPC connections (Phase 2+)
type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// Server is the interface that game server implementations must satisfy
// This allows the in-process client to call server methods directly
type Server interface {
	AuthService
	PlayerService
	GameService
}

// AuthService interface matches the protobuf service definition
type AuthService interface {
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error)
	AuthenticateSSH(ctx context.Context, req *SSHAuthRequest) (*AuthResponse, error)
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error)
	ValidateSession(ctx context.Context, req *ValidateSessionRequest) (*Session, error)
	EndSession(ctx context.Context, req *EndSessionRequest) error
	RefreshSession(ctx context.Context, req *RefreshSessionRequest) (*AuthResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
}

// PlayerService interface matches the protobuf service definition
type PlayerService interface {
	GetPlayerState(ctx context.Context, playerID uuid.UUID) (*PlayerState, error)
	UpdatePlayerLocation(ctx context.Context, req *LocationUpdate) (*PlayerState, error)
	GetPlayerShip(ctx context.Context, playerID uuid.UUID) (*Ship, error)
	GetPlayerInventory(ctx context.Context, playerID uuid.UUID) (*Inventory, error)
	GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error)
	GetPlayerReputation(ctx context.Context, playerID uuid.UUID) (*ReputationInfo, error)
	StreamPlayerUpdates(ctx context.Context, playerID uuid.UUID) (PlayerUpdateStream, error)
}

// GameService interface matches the protobuf service definition
type GameService interface {
	Jump(ctx context.Context, req *JumpRequest) (*JumpResponse, error)
	Land(ctx context.Context, req *LandRequest) (*LandResponse, error)
	Takeoff(ctx context.Context, req *TakeoffRequest) (*TakeoffResponse, error)

	GetMarket(ctx context.Context, systemID uuid.UUID) (*Market, error)
	BuyCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error)
	SellCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error)

	BuyShip(ctx context.Context, req *ShipPurchaseRequest) (*ShipPurchaseResponse, error)
	SellShip(ctx context.Context, req *ShipSaleRequest) (*ShipSaleResponse, error)
	BuyOutfit(ctx context.Context, req *OutfitPurchaseRequest) (*OutfitPurchaseResponse, error)
	SellOutfit(ctx context.Context, req *OutfitSaleRequest) (*OutfitSaleResponse, error)

	GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error)
	AcceptMission(ctx context.Context, req *MissionAcceptRequest) (*Mission, error)
	AbandonMission(ctx context.Context, missionID uuid.UUID) error
	GetActiveMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error)

	GetAvailableQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error)
	AcceptQuest(ctx context.Context, req *QuestAcceptRequest) (*Quest, error)
	GetActiveQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error)
}

// inProcessClient implements Client by calling server methods directly
type inProcessClient struct {
	server Server
}

// newInProcessClient creates a client that calls server methods directly
func newInProcessClient(config *ClientConfig) (*inProcessClient, error) {
	if config.InProcessServer == nil {
		return nil, ErrNoServerProvided
	}

	return &inProcessClient{
		server: config.InProcessServer,
	}, nil
}

// Close implements Client
func (c *inProcessClient) Close() error {
	// No cleanup needed for in-process client
	return nil
}

// Auth methods - delegate directly to server
func (c *inProcessClient) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error) {
	return c.server.Authenticate(ctx, req)
}

func (c *inProcessClient) AuthenticateSSH(ctx context.Context, req *SSHAuthRequest) (*AuthResponse, error) {
	return c.server.AuthenticateSSH(ctx, req)
}

func (c *inProcessClient) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	return c.server.CreateSession(ctx, req)
}

func (c *inProcessClient) ValidateSession(ctx context.Context, req *ValidateSessionRequest) (*Session, error) {
	return c.server.ValidateSession(ctx, req)
}

func (c *inProcessClient) EndSession(ctx context.Context, req *EndSessionRequest) error {
	return c.server.EndSession(ctx, req)
}

func (c *inProcessClient) RefreshSession(ctx context.Context, req *RefreshSessionRequest) (*AuthResponse, error) {
	return c.server.RefreshSession(ctx, req)
}

func (c *inProcessClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	return c.server.Register(ctx, req)
}

// Player methods - delegate directly to server
func (c *inProcessClient) GetPlayerState(ctx context.Context, playerID uuid.UUID) (*PlayerState, error) {
	return c.server.GetPlayerState(ctx, playerID)
}

func (c *inProcessClient) UpdatePlayerLocation(ctx context.Context, req *LocationUpdate) (*PlayerState, error) {
	return c.server.UpdatePlayerLocation(ctx, req)
}

func (c *inProcessClient) GetPlayerShip(ctx context.Context, playerID uuid.UUID) (*Ship, error) {
	return c.server.GetPlayerShip(ctx, playerID)
}

func (c *inProcessClient) GetPlayerInventory(ctx context.Context, playerID uuid.UUID) (*Inventory, error) {
	return c.server.GetPlayerInventory(ctx, playerID)
}

func (c *inProcessClient) GetPlayerStats(ctx context.Context, playerID uuid.UUID) (*PlayerStats, error) {
	return c.server.GetPlayerStats(ctx, playerID)
}

func (c *inProcessClient) GetPlayerReputation(ctx context.Context, playerID uuid.UUID) (*ReputationInfo, error) {
	return c.server.GetPlayerReputation(ctx, playerID)
}

func (c *inProcessClient) StreamPlayerUpdates(ctx context.Context, playerID uuid.UUID) (PlayerUpdateStream, error) {
	return c.server.StreamPlayerUpdates(ctx, playerID)
}

// Game methods - delegate directly to server
func (c *inProcessClient) Jump(ctx context.Context, req *JumpRequest) (*JumpResponse, error) {
	return c.server.Jump(ctx, req)
}

func (c *inProcessClient) Land(ctx context.Context, req *LandRequest) (*LandResponse, error) {
	return c.server.Land(ctx, req)
}

func (c *inProcessClient) Takeoff(ctx context.Context, req *TakeoffRequest) (*TakeoffResponse, error) {
	return c.server.Takeoff(ctx, req)
}

func (c *inProcessClient) GetMarket(ctx context.Context, systemID uuid.UUID) (*Market, error) {
	return c.server.GetMarket(ctx, systemID)
}

func (c *inProcessClient) BuyCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error) {
	return c.server.BuyCommodity(ctx, req)
}

func (c *inProcessClient) SellCommodity(ctx context.Context, req *TradeRequest) (*TradeResponse, error) {
	return c.server.SellCommodity(ctx, req)
}

func (c *inProcessClient) BuyShip(ctx context.Context, req *ShipPurchaseRequest) (*ShipPurchaseResponse, error) {
	return c.server.BuyShip(ctx, req)
}

func (c *inProcessClient) SellShip(ctx context.Context, req *ShipSaleRequest) (*ShipSaleResponse, error) {
	return c.server.SellShip(ctx, req)
}

func (c *inProcessClient) BuyOutfit(ctx context.Context, req *OutfitPurchaseRequest) (*OutfitPurchaseResponse, error) {
	return c.server.BuyOutfit(ctx, req)
}

func (c *inProcessClient) SellOutfit(ctx context.Context, req *OutfitSaleRequest) (*OutfitSaleResponse, error) {
	return c.server.SellOutfit(ctx, req)
}

func (c *inProcessClient) GetAvailableMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error) {
	return c.server.GetAvailableMissions(ctx, playerID)
}

func (c *inProcessClient) AcceptMission(ctx context.Context, req *MissionAcceptRequest) (*Mission, error) {
	return c.server.AcceptMission(ctx, req)
}

func (c *inProcessClient) AbandonMission(ctx context.Context, missionID uuid.UUID) error {
	return c.server.AbandonMission(ctx, missionID)
}

func (c *inProcessClient) GetActiveMissions(ctx context.Context, playerID uuid.UUID) (*MissionList, error) {
	return c.server.GetActiveMissions(ctx, playerID)
}

func (c *inProcessClient) GetAvailableQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error) {
	return c.server.GetAvailableQuests(ctx, playerID)
}

func (c *inProcessClient) AcceptQuest(ctx context.Context, req *QuestAcceptRequest) (*Quest, error) {
	return c.server.AcceptQuest(ctx, req)
}

func (c *inProcessClient) GetActiveQuests(ctx context.Context, playerID uuid.UUID) (*QuestList, error) {
	return c.server.GetActiveQuests(ctx, playerID)
}

// Compile-time check that inProcessClient implements Client
var _ Client = (*inProcessClient)(nil)
