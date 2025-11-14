// File: internal/api/types.go
// Project: Terminal Velocity
// Description: API types for client-server communication
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package api

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Common errors
var (
	ErrNoServerProvided = errors.New("no server provided for in-process client")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
)

// Auth Types

type AuthRequest struct {
	Username   string
	Password   string
	ClientInfo string
}

type SSHAuthRequest struct {
	Username              string
	PublicKeyFingerprint  string
	PublicKey             []byte
}

type AuthResponse struct {
	PlayerID   uuid.UUID
	Token      string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	PlayerInfo *PlayerInfo
}

type PlayerInfo struct {
	PlayerID  uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
	LastLogin time.Time
	IsAdmin   bool
	Role      string
}

type CreateSessionRequest struct {
	PlayerID uuid.UUID
	Token    string
}

type Session struct {
	SessionID    uuid.UUID
	PlayerID     uuid.UUID
	CreatedAt    time.Time
	LastActivity time.Time
	ExpiresAt    time.Time
	State        SessionState
}

type SessionState string

const (
	SessionStateActive     SessionState = "active"
	SessionStateIdle       SessionState = "idle"
	SessionStateExpired    SessionState = "expired"
	SessionStateTerminated SessionState = "terminated"
)

type ValidateSessionRequest struct {
	SessionID uuid.UUID
	Token     string
}

type EndSessionRequest struct {
	SessionID uuid.UUID
}

type RefreshSessionRequest struct {
	SessionID uuid.UUID
	Token     string
}

type RegisterRequest struct {
	Username     string
	Password     string
	Email        string
	SSHPublicKey []byte
}

type RegisterResponse struct {
	PlayerID uuid.UUID
	Username string
	Message  string
}

// Player Types

type PlayerState struct {
	PlayerID        uuid.UUID
	Username        string
	CurrentSystemID uuid.UUID
	CurrentPlanetID *uuid.UUID
	Position        Coordinates
	Credits         int64
	Fuel            int32
	CurrentShipID   uuid.UUID
	Ship            *Ship
	Inventory       *Inventory
	Stats           *PlayerStats
	Reputation      *ReputationInfo
	Status          PlayerStatus
	LastSave        time.Time
}

type Coordinates struct {
	X float64
	Y float64
	Z float64
}

type Ship struct {
	ShipID         uuid.UUID
	ShipType       string
	CustomName     string
	Hull           int32
	MaxHull        int32
	Shields        int32
	MaxShields     int32
	Fuel           int32
	MaxFuel        int32
	CargoSpace     int32
	CargoUsed      int32
	Weapons        []*Weapon
	Outfits        []*Outfit
	Speed          float64
	Acceleration   float64
	TurnRate       float64
	PurchasePrice  int64
	CurrentValue   int64
}

type Weapon struct {
	WeaponID   string
	WeaponType string
	Damage     int32
	Range      int32
	Accuracy   float64
	Ammo       int32
	MaxAmmo    int32
	Cooldown   int32
}

type Outfit struct {
	OutfitID    string
	OutfitType  string
	Name        string
	Description string
	Modifiers   map[string]int32
}

type Inventory struct {
	Cargo           map[string]int32
	Items           []*Item
	TotalCargoSpace int32
	CargoUsed       int32
}

type Item struct {
	ItemID      string
	Name        string
	Description string
	Rarity      Rarity
	IsQuestItem bool
}

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityLegendary Rarity = "legendary"
)

type PlayerStats struct {
	Level               int32
	Experience          int64
	TotalCreditsEarned  int64
	CombatRating        int32
	TradeRating         int32
	ExplorationRating   int32
	ShipsDestroyed      int32
	MissionsCompleted   int32
	QuestsCompleted     int32
	SystemsVisited      int32
	JumpsMade           int32
	AccountCreated      time.Time
	PlaytimeSeconds     int64
}

type ReputationInfo struct {
	FactionReputation map[string]int32
	LegalStatus       string
	Bounty            int64
}

type PlayerStatus string

const (
	PlayerStatusDocked  PlayerStatus = "docked"
	PlayerStatusInSpace PlayerStatus = "in_space"
	PlayerStatusCombat  PlayerStatus = "in_combat"
	PlayerStatusJumping PlayerStatus = "jumping"
	PlayerStatusTrading PlayerStatus = "trading"
)

type LocationUpdate struct {
	PlayerID  uuid.UUID
	SystemID  uuid.UUID
	PlanetID  *uuid.UUID
	Position  Coordinates
}

type PlayerUpdate struct {
	PlayerID  uuid.UUID
	Type      UpdateType
	Timestamp time.Time
	// Union of update types - only one will be set
	CreditsUpdate    *CreditsUpdate
	LocationUpdate   *LocationUpdate
	ShipUpdate       *ShipUpdate
	InventoryUpdate  *InventoryUpdate
	StatusUpdate     *StatusUpdate
	ReputationUpdate *ReputationUpdate
}

type UpdateType string

const (
	UpdateTypeCredits    UpdateType = "credits"
	UpdateTypeLocation   UpdateType = "location"
	UpdateTypeShip       UpdateType = "ship"
	UpdateTypeInventory  UpdateType = "inventory"
	UpdateTypeStatus     UpdateType = "status"
	UpdateTypeReputation UpdateType = "reputation"
)

type CreditsUpdate struct {
	OldCredits int64
	NewCredits int64
	Delta      int64
	Reason     string
}

type ShipUpdate struct {
	Ship *Ship
}

type InventoryUpdate struct {
	Inventory *Inventory
}

type StatusUpdate struct {
	OldStatus PlayerStatus
	NewStatus PlayerStatus
}

type ReputationUpdate struct {
	FactionID     string
	OldReputation int32
	NewReputation int32
	Delta         int32
	Reason        string
}

// Game Types

type JumpRequest struct {
	PlayerID       uuid.UUID
	TargetSystemID uuid.UUID
}

type JumpResponse struct {
	Success       bool
	Message       string
	NewState      *PlayerState
	FuelConsumed  int32
}

type LandRequest struct {
	PlayerID uuid.UUID
	PlanetID uuid.UUID
}

type LandResponse struct {
	Success  bool
	Message  string
	Planet   *Planet
	NewState *PlayerState
}

type TakeoffRequest struct {
	PlayerID uuid.UUID
}

type TakeoffResponse struct {
	Success  bool
	Message  string
	NewState *PlayerState
}

type Planet struct {
	PlanetID    uuid.UUID
	Name        string
	Description string
	SystemID    uuid.UUID
	Services    []string
	TechLevel   int32
	Government  string
	Population  int64
}

type Market struct {
	SystemID    uuid.UUID
	Commodities []*CommodityListing
	LastUpdated time.Time
}

type CommodityListing struct {
	CommodityID string
	Name        string
	BuyPrice    int32
	SellPrice   int32
	Stock       int32
	IsIllegal   bool
}

type TradeRequest struct {
	PlayerID    uuid.UUID
	CommodityID string
	Quantity    int32
}

type TradeResponse struct {
	Success        bool
	Message        string
	QuantityTraded int32
	TotalCost      int64
	PricePerUnit   int32
	NewState       *PlayerState
}

type ShipPurchaseRequest struct {
	PlayerID      uuid.UUID
	ShipType      string
	TradeInShipID *uuid.UUID
}

type ShipPurchaseResponse struct {
	Success      bool
	Message      string
	NewShip      *Ship
	TotalCost    int64
	TradeInValue int64
	NewState     *PlayerState
}

type ShipSaleRequest struct {
	PlayerID uuid.UUID
	ShipID   uuid.UUID
}

type ShipSaleResponse struct {
	Success   bool
	Message   string
	SaleValue int64
	NewState  *PlayerState
}

type OutfitPurchaseRequest struct {
	PlayerID uuid.UUID
	OutfitID string
	Quantity int32
}

type OutfitPurchaseResponse struct {
	Success     bool
	Message     string
	Outfits     []*Outfit
	TotalCost   int64
	UpdatedShip *Ship
}

type OutfitSaleRequest struct {
	PlayerID uuid.UUID
	OutfitID string
	Quantity int32
}

type OutfitSaleResponse struct {
	Success     bool
	Message     string
	SaleValue   int64
	UpdatedShip *Ship
}

type MissionList struct {
	Missions   []*Mission
	TotalCount int32
}

type Mission struct {
	MissionID           uuid.UUID
	Title               string
	Description         string
	Type                MissionType
	RewardCredits       int64
	RewardReputation    int32
	OriginSystemID      uuid.UUID
	DestinationSystemID uuid.UUID
	Deadline            time.Time
	Status              MissionStatus
	ProgressCurrent     int32
	ProgressRequired    int32
}

type MissionType string

const (
	MissionTypeDelivery MissionType = "delivery"
	MissionTypeCombat   MissionType = "combat"
	MissionTypeBounty   MissionType = "bounty"
	MissionTypeTrading  MissionType = "trading"
)

type MissionStatus string

const (
	MissionStatusAvailable MissionStatus = "available"
	MissionStatusActive    MissionStatus = "active"
	MissionStatusCompleted MissionStatus = "completed"
	MissionStatusFailed    MissionStatus = "failed"
)

type MissionAcceptRequest struct {
	PlayerID  uuid.UUID
	MissionID uuid.UUID
}

type QuestList struct {
	Quests     []*Quest
	TotalCount int32
}

type Quest struct {
	QuestID          uuid.UUID
	Title            string
	Description      string
	Type             QuestType
	Objectives       []*QuestObjective
	Rewards          []*QuestReward
	Status           QuestStatus
	IsMainQuest      bool
	RecommendedLevel int32
}

type QuestType string

const (
	QuestTypeMain    QuestType = "main"
	QuestTypeSide    QuestType = "side"
	QuestTypeFaction QuestType = "faction"
	QuestTypeDaily   QuestType = "daily"
	QuestTypeChain   QuestType = "chain"
	QuestTypeHidden  QuestType = "hidden"
	QuestTypeEvent   QuestType = "event"
)

type QuestStatus string

const (
	QuestStatusLocked    QuestStatus = "locked"
	QuestStatusAvailable QuestStatus = "available"
	QuestStatusActive    QuestStatus = "active"
	QuestStatusCompleted QuestStatus = "completed"
	QuestStatusFailed    QuestStatus = "failed"
)

type QuestObjective struct {
	ObjectiveID      string
	Description      string
	Type             ObjectiveType
	ProgressCurrent  int32
	ProgressRequired int32
	Completed        bool
}

type ObjectiveType string

const (
	ObjectiveTypeDeliver     ObjectiveType = "deliver"
	ObjectiveTypeDestroy     ObjectiveType = "destroy"
	ObjectiveTypeTravel      ObjectiveType = "travel"
	ObjectiveTypeCollect     ObjectiveType = "collect"
	ObjectiveTypeEscort      ObjectiveType = "escort"
	ObjectiveTypeDefend      ObjectiveType = "defend"
	ObjectiveTypeInvestigate ObjectiveType = "investigate"
	ObjectiveTypeTalk        ObjectiveType = "talk"
	ObjectiveTypeScan        ObjectiveType = "scan"
	ObjectiveTypeDiscover    ObjectiveType = "discover"
	ObjectiveTypeHack        ObjectiveType = "hack"
	ObjectiveTypeSurvive     ObjectiveType = "survive"
)

type QuestReward struct {
	Type   RewardType
	Value  int64
	ItemID string
}

type RewardType string

const (
	RewardTypeCredits    RewardType = "credits"
	RewardTypeExperience RewardType = "experience"
	RewardTypeItem       RewardType = "item"
	RewardTypeReputation RewardType = "reputation"
	RewardTypeShip       RewardType = "ship"
	RewardTypeUnlock     RewardType = "unlock"
)

type QuestAcceptRequest struct {
	PlayerID uuid.UUID
	QuestID  uuid.UUID
}
