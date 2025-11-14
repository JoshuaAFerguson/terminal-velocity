// File: internal/api/server/converters.go
// Project: Terminal Velocity
// Description: Converters between database models and API types
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"time"

	"github.com/google/uuid"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// convertPlayerToAPI converts a database Player model to API PlayerState
func convertPlayerToAPI(player *models.Player, ship *models.Ship) *api.PlayerState {
	if player == nil {
		return nil
	}

	state := &api.PlayerState{
		PlayerID:        player.ID,
		Username:        player.Username,
		CurrentSystemID: player.CurrentSystem,
		CurrentPlanetID: player.CurrentPlanet,
		Position: api.Coordinates{
			X: 0, // TODO: Add coordinates to Player model
			Y: 0,
			Z: 0,
		},
		Credits:       player.Credits,
		Fuel:          0, // Fuel is on Ship, not Player
		CurrentShipID: player.ShipID,
		LastSave:      time.Now(), // TODO: Add UpdatedAt to Player model
	}

	// Convert ship if provided
	if ship != nil {
		state.Ship = convertShipToAPI(ship)
		state.Inventory = convertInventoryToAPI(ship)
		state.Fuel = int32(ship.Fuel) // Get fuel from ship
	}

	// Determine player status
	if player.CurrentPlanet != nil {
		state.Status = api.PlayerStatusDocked
	} else {
		state.Status = api.PlayerStatusInSpace
	}

	return state
}

// convertShipToAPI converts a database Ship model to API Ship
func convertShipToAPI(ship *models.Ship) *api.Ship {
	if ship == nil {
		return nil
	}

	apiShip := &api.Ship{
		ShipID:        ship.ID,
		ShipType:      ship.TypeID,
		CustomName:    ship.Name,
		Hull:          int32(ship.Hull),
		MaxHull:       0, // TODO: Get from ShipType
		Shields:       int32(ship.Shields),
		MaxShields:    0, // TODO: Get from ShipType
		Fuel:          int32(ship.Fuel),
		MaxFuel:       0,  // TODO: Get from ShipType
		CargoSpace:    0,  // TODO: Get from ShipType
		CargoUsed:     int32(ship.GetCargoUsed()),
		Speed:         0,  // TODO: Get from ShipType
		Acceleration:  0,  // Not in model
		TurnRate:      0,  // Not in model
		PurchasePrice: 0,  // TODO: Get from ShipType
		CurrentValue:  0,  // TODO: Calculate depreciation
		Weapons:       make([]*api.Weapon, 0),
		Outfits:       make([]*api.Outfit, 0),
	}

	// Ship.Weapons and Ship.Outfits are just string IDs, not full objects
	// TODO: Load actual weapon and outfit data from repositories if needed
	// For now, just return the basic ship info

	return apiShip
}

// convertInventoryToAPI converts ship cargo to API Inventory
func convertInventoryToAPI(ship *models.Ship) *api.Inventory {
	if ship == nil {
		return nil
	}

	inventory := &api.Inventory{
		Cargo:           make(map[string]int32),
		Items:           make([]*api.Item, 0),
		TotalCargoSpace: 0, // TODO: Get from ShipType
		CargoUsed:       int32(ship.GetCargoUsed()),
	}

	// Convert cargo (ship.Cargo is []CargoItem, not map)
	for _, cargoItem := range ship.Cargo {
		inventory.Cargo[cargoItem.CommodityID] = int32(cargoItem.Quantity)
	}

	// TODO: Convert items when item system is implemented
	// For now, items array is empty

	return inventory
}

// convertPlayerStatsToAPI converts database player stats to API PlayerStats
func convertPlayerStatsToAPI(player *models.Player) *api.PlayerStats {
	if player == nil {
		return nil
	}

	return &api.PlayerStats{
		Level:              0, // TODO: Add Level to Player model
		Experience:         0, // TODO: Add Experience to Player model
		TotalCreditsEarned: player.TradeProfit, // Using TradeProfit as proxy
		CombatRating:       int32(player.CombatRating),
		TradeRating:        int32(player.TradingRating),
		ExplorationRating:  int32(player.ExplorationRating),
		ShipsDestroyed:     int32(player.TotalKills),
		MissionsCompleted:  int32(player.MissionsCompleted),
		QuestsCompleted:    0, // TODO: Add QuestsCompleted to Player model
		SystemsVisited:     int32(player.SystemsVisited),
		JumpsMade:          int32(player.TotalJumps),
		AccountCreated:     player.CreatedAt,
		PlaytimeSeconds:    player.PlayTime,
	}
}

// convertReputationToAPI converts player reputation data to API ReputationInfo
func convertReputationToAPI(player *models.Player) *api.ReputationInfo {
	if player == nil {
		return nil
	}

	reputation := &api.ReputationInfo{
		FactionReputation: make(map[string]int32),
		LegalStatus:       "citizen", // TODO: Add LegalStatus to Player model
		Bounty:            0,          // TODO: Add Bounty to Player model
	}

	// Convert faction reputation (Player.Reputation is map[string]int)
	for faction, rep := range player.Reputation {
		reputation.FactionReputation[faction] = int32(rep)
	}

	// Criminal status
	if player.IsCriminal {
		reputation.LegalStatus = "criminal"
	}

	return reputation
}

// convertMarketToAPI converts database market data to API Market
func convertMarketToAPI(prices []models.MarketPrice, commodities map[string]*models.Commodity) *api.Market {
	market := &api.Market{
		Commodities: make([]*api.CommodityListing, 0, len(prices)),
		// LastUpdated will be set from latest price update
	}

	for _, price := range prices {
		commodity := commodities[price.CommodityID]
		if commodity == nil {
			continue // Skip if commodity definition not found
		}

		market.Commodities = append(market.Commodities, &api.CommodityListing{
			CommodityID: price.CommodityID,
			Name:        commodity.Name,
			BuyPrice:    int32(price.BuyPrice),
			SellPrice:   int32(price.SellPrice),
			Stock:       int32(price.Stock),
			IsIllegal:   false, // TODO: Check if commodity is illegal in this system
		})
	}

	return market
}

// convertPlanetToAPI converts database planet to API Planet
func convertPlanetToAPI(planet *models.Planet) *api.Planet {
	if planet == nil {
		return nil
	}

	return &api.Planet{
		PlanetID:    planet.ID,
		Name:        planet.Name,
		Description: planet.Description,
		SystemID:    planet.SystemID,
		Services:    planet.Services,
		TechLevel:   int32(planet.TechLevel),
		Government:  "", // TODO: Get government from StarSystem
		Population:  planet.Population,
	}
}

// convertMissionToAPI converts database mission to API Mission
func convertMissionToAPI(mission *models.Mission) *api.Mission {
	if mission == nil {
		return nil
	}

	apiMission := &api.Mission{
		MissionID:           mission.ID,
		Title:               mission.Title,
		Description:         mission.Description,
		RewardCredits:       mission.Reward,
		RewardReputation:    0, // TODO: Sum reputation changes
		OriginSystemID:      uuid.Nil, // TODO: Get system from OriginPlanet
		DestinationSystemID: uuid.Nil, // TODO: Get from Destination
		Deadline:            mission.Deadline,
		ProgressCurrent:     int32(mission.Progress),
		ProgressRequired:    int32(mission.Quantity),
	}

	// Convert mission type
	switch mission.Type {
	case "delivery":
		apiMission.Type = api.MissionTypeDelivery
	case "combat":
		apiMission.Type = api.MissionTypeCombat
	case "bounty":
		apiMission.Type = api.MissionTypeBounty
	default:
		apiMission.Type = api.MissionTypeDelivery
	}

	// Convert mission status
	switch mission.Status {
	case "available":
		apiMission.Status = api.MissionStatusAvailable
	case "active":
		apiMission.Status = api.MissionStatusActive
	case "completed":
		apiMission.Status = api.MissionStatusCompleted
	case "failed":
		apiMission.Status = api.MissionStatusFailed
	default:
		apiMission.Status = api.MissionStatusAvailable
	}

	return apiMission
}

// convertQuestToAPI converts database quest to API Quest
func convertQuestToAPI(quest *models.Quest) *api.Quest {
	if quest == nil {
		return nil
	}

	// Quest.ID is string, but api.Quest.QuestID is uuid.UUID
	questUUID, _ := uuid.Parse(quest.ID)

	apiQuest := &api.Quest{
		QuestID:          questUUID,
		Title:            quest.Title,
		Description:      quest.Description,
		Objectives:       make([]*api.QuestObjective, 0),
		Rewards:          make([]*api.QuestReward, 0),
		IsMainQuest:      quest.Type == models.QuestTypeMain, // Derive from Type
		RecommendedLevel: int32(quest.Level),
	}

	// Convert quest type (quest.Type is QuestType enum, not string)
	switch quest.Type {
	case models.QuestTypeMain:
		apiQuest.Type = api.QuestTypeMain
	case models.QuestTypeSide:
		apiQuest.Type = api.QuestTypeSide
	case models.QuestTypeFaction:
		apiQuest.Type = api.QuestTypeFaction
	case models.QuestTypeDaily:
		apiQuest.Type = api.QuestTypeDaily
	default:
		apiQuest.Type = api.QuestTypeSide
	}

	// Quest model doesn't have Status field - it's in PlayerQuest
	// For now, assume all quests are available
	apiQuest.Status = api.QuestStatusAvailable

	// TODO: Convert objectives - Quest.Objectives structure is complex
	// For now, return empty array as these are placeholder converters

	// TODO: Convert rewards - Quest.Rewards is QuestReward (singular struct), not array
	// apiQuest.Rewards should include credits from quest.Rewards.Credits, etc.

	return apiQuest
}
