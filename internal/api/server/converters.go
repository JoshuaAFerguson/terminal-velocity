// File: internal/api/server/converters.go
// Project: Terminal Velocity
// Description: Converters between database models and API types
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
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
		CurrentSystemID: player.CurrentSystemID,
		CurrentPlanetID: player.CurrentPlanetID,
		Position: api.Coordinates{
			X: player.X,
			Y: player.Y,
			Z: 0, // 2D positioning for now
		},
		Credits:       player.Credits,
		Fuel:          player.Fuel,
		CurrentShipID: player.CurrentShipID,
		LastSave:      player.UpdatedAt,
	}

	// Convert ship if provided
	if ship != nil {
		state.Ship = convertShipToAPI(ship)
		state.Inventory = convertInventoryToAPI(ship)
	}

	// Determine player status
	if player.CurrentPlanetID != nil {
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
		ShipType:      ship.ShipType,
		CustomName:    ship.Name,
		Hull:          ship.Hull,
		MaxHull:       ship.MaxHull,
		Shields:       ship.Shields,
		MaxShields:    ship.MaxShields,
		Fuel:          ship.Fuel,
		MaxFuel:       ship.MaxFuel,
		CargoSpace:    ship.CargoSpace,
		CargoUsed:     ship.CargoUsed,
		Speed:         ship.Speed,
		Acceleration:  ship.Acceleration,
		TurnRate:      ship.TurnRate,
		PurchasePrice: ship.PurchasePrice,
		CurrentValue:  ship.CurrentValue,
		Weapons:       make([]*api.Weapon, 0),
		Outfits:       make([]*api.Outfit, 0),
	}

	// Convert weapons
	for _, weapon := range ship.Weapons {
		apiShip.Weapons = append(apiShip.Weapons, &api.Weapon{
			WeaponID:   weapon.ID,
			WeaponType: weapon.WeaponType,
			Damage:     weapon.Damage,
			Range:      weapon.Range,
			Accuracy:   weapon.Accuracy,
			Ammo:       weapon.Ammo,
			MaxAmmo:    weapon.MaxAmmo,
			Cooldown:   weapon.Cooldown,
		})
	}

	// Convert outfits
	for _, outfit := range ship.Outfits {
		apiOutfit := &api.Outfit{
			OutfitID:    outfit.ID,
			OutfitType:  outfit.OutfitType,
			Name:        outfit.Name,
			Description: outfit.Description,
			Modifiers:   make(map[string]int32),
		}

		// Convert modifiers
		for k, v := range outfit.Modifiers {
			apiOutfit.Modifiers[k] = int32(v)
		}

		apiShip.Outfits = append(apiShip.Outfits, apiOutfit)
	}

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
		TotalCargoSpace: ship.CargoSpace,
		CargoUsed:       ship.CargoUsed,
	}

	// Convert cargo
	for commodity, quantity := range ship.Cargo {
		inventory.Cargo[commodity] = int32(quantity)
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
		Level:              player.Level,
		Experience:         player.Experience,
		TotalCreditsEarned: player.TotalCreditsEarned,
		CombatRating:       player.CombatRating,
		TradeRating:        player.TradeRating,
		ExplorationRating:  player.ExplorationRating,
		ShipsDestroyed:     player.ShipsDestroyed,
		MissionsCompleted:  player.MissionsCompleted,
		QuestsCompleted:    player.QuestsCompleted,
		SystemsVisited:     player.SystemsVisited,
		JumpsMade:          player.JumpsMade,
		AccountCreated:     player.CreatedAt,
		PlaytimeSeconds:    player.PlaytimeSeconds,
	}
}

// convertReputationToAPI converts player reputation data to API ReputationInfo
func convertReputationToAPI(player *models.Player) *api.ReputationInfo {
	if player == nil {
		return nil
	}

	reputation := &api.ReputationInfo{
		FactionReputation: make(map[string]int32),
		LegalStatus:       player.LegalStatus,
		Bounty:            player.Bounty,
	}

	// Convert faction reputation
	for faction, rep := range player.FactionReputation {
		reputation.FactionReputation[faction] = int32(rep)
	}

	return reputation
}

// convertMarketToAPI converts database market data to API Market
func convertMarketToAPI(systemID string, commodities []models.CommodityListing, lastUpdated string) *api.Market {
	market := &api.Market{
		Commodities: make([]*api.CommodityListing, 0, len(commodities)),
		// LastUpdated will be set from database timestamp
	}

	for _, commodity := range commodities {
		market.Commodities = append(market.Commodities, &api.CommodityListing{
			CommodityID: commodity.CommodityID,
			Name:        commodity.Name,
			BuyPrice:    commodity.BuyPrice,
			SellPrice:   commodity.SellPrice,
			Stock:       commodity.Stock,
			IsIllegal:   commodity.IsIllegal,
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
		TechLevel:   planet.TechLevel,
		Government:  planet.Government,
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
		RewardCredits:       mission.RewardCredits,
		RewardReputation:    mission.RewardReputation,
		OriginSystemID:      mission.OriginSystemID,
		DestinationSystemID: mission.DestinationSystemID,
		Deadline:            mission.Deadline,
		ProgressCurrent:     mission.ProgressCurrent,
		ProgressRequired:    mission.ProgressRequired,
	}

	// Convert mission type
	switch mission.Type {
	case "delivery":
		apiMission.Type = api.MissionTypeDelivery
	case "combat":
		apiMission.Type = api.MissionTypeCombat
	case "bounty":
		apiMission.Type = api.MissionTypeBounty
	case "trading":
		apiMission.Type = api.MissionTypeTrading
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
	}

	return apiMission
}

// convertQuestToAPI converts database quest to API Quest
func convertQuestToAPI(quest *models.Quest) *api.Quest {
	if quest == nil {
		return nil
	}

	apiQuest := &api.Quest{
		QuestID:          quest.ID,
		Title:            quest.Title,
		Description:      quest.Description,
		Objectives:       make([]*api.QuestObjective, 0),
		Rewards:          make([]*api.QuestReward, 0),
		IsMainQuest:      quest.IsMainQuest,
		RecommendedLevel: quest.RecommendedLevel,
	}

	// Convert quest type
	switch quest.Type {
	case "main":
		apiQuest.Type = api.QuestTypeMain
	case "side":
		apiQuest.Type = api.QuestTypeSide
	case "faction":
		apiQuest.Type = api.QuestTypeFaction
	case "daily":
		apiQuest.Type = api.QuestTypeDaily
	case "chain":
		apiQuest.Type = api.QuestTypeChain
	case "hidden":
		apiQuest.Type = api.QuestTypeHidden
	case "event":
		apiQuest.Type = api.QuestTypeEvent
	}

	// Convert quest status
	switch quest.Status {
	case "locked":
		apiQuest.Status = api.QuestStatusLocked
	case "available":
		apiQuest.Status = api.QuestStatusAvailable
	case "active":
		apiQuest.Status = api.QuestStatusActive
	case "completed":
		apiQuest.Status = api.QuestStatusCompleted
	case "failed":
		apiQuest.Status = api.QuestStatusFailed
	}

	// Convert objectives
	for _, objective := range quest.Objectives {
		apiObjective := &api.QuestObjective{
			ObjectiveID:      objective.ID,
			Description:      objective.Description,
			ProgressCurrent:  objective.ProgressCurrent,
			ProgressRequired: objective.ProgressRequired,
			Completed:        objective.Completed,
		}

		// Convert objective type
		switch objective.Type {
		case "deliver":
			apiObjective.Type = api.ObjectiveTypeDeliver
		case "destroy":
			apiObjective.Type = api.ObjectiveTypeDestroy
		case "travel":
			apiObjective.Type = api.ObjectiveTypeTravel
		case "collect":
			apiObjective.Type = api.ObjectiveTypeCollect
		// Add more as needed
		}

		apiQuest.Objectives = append(apiQuest.Objectives, apiObjective)
	}

	// Convert rewards
	for _, reward := range quest.Rewards {
		apiReward := &api.QuestReward{
			Value:  reward.Value,
			ItemID: reward.ItemID,
		}

		// Convert reward type
		switch reward.Type {
		case "credits":
			apiReward.Type = api.RewardTypeCredits
		case "experience":
			apiReward.Type = api.RewardTypeExperience
		case "item":
			apiReward.Type = api.RewardTypeItem
		case "reputation":
			apiReward.Type = api.RewardTypeReputation
		case "ship":
			apiReward.Type = api.RewardTypeShip
		case "unlock":
			apiReward.Type = api.RewardTypeUnlock
		}

		apiQuest.Rewards = append(apiQuest.Rewards, apiReward)
	}

	return apiQuest
}
