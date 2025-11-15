// File: internal/api/server/converters.go
// Project: Terminal Velocity
// Description: Converters between database models and API types
// Version: 1.4.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package server

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/api"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
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
			X: player.X,
			Y: player.Y,
			Z: 0, // TODO: Add Z coordinate if needed for 3D space
		},
		Credits:       player.Credits,
		Fuel:          0, // Fuel is on Ship, not Player
		CurrentShipID: player.ShipID,
		LastSave:      player.UpdatedAt,
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

	// Load ship type for stats
	shipType := models.GetShipTypeByID(ship.TypeID)

	maxHull := int32(0)
	maxShields := int32(0)
	maxFuel := int32(0)
	cargoSpace := int32(0)
	speed := int32(0)
	purchasePrice := int64(0)
	currentValue := int64(0)

	if shipType != nil {
		maxHull = int32(shipType.MaxHull)
		maxShields = int32(shipType.MaxShields)
		maxFuel = int32(shipType.MaxFuel)
		cargoSpace = int32(shipType.CargoSpace)
		speed = int32(shipType.Speed)
		purchasePrice = shipType.Price
		// Simple depreciation: 80% of purchase price
		currentValue = int64(float64(shipType.Price) * 0.8)
	}

	apiShip := &api.Ship{
		ShipID:        ship.ID,
		ShipType:      ship.TypeID,
		CustomName:    ship.Name,
		Hull:          int32(ship.Hull),
		MaxHull:       maxHull,
		Shields:       int32(ship.Shields),
		MaxShields:    maxShields,
		Fuel:          int32(ship.Fuel),
		MaxFuel:       maxFuel,
		CargoSpace:    cargoSpace,
		CargoUsed:     int32(ship.GetCargoUsed()),
		Speed:         float64(speed),
		Acceleration:  0, // Not currently modeled
		TurnRate:      0, // Not currently modeled
		PurchasePrice: purchasePrice,
		CurrentValue:  currentValue,
		Weapons:       convertWeaponsToAPI(ship.Weapons, ship.WeaponAmmo),
		Outfits:       convertOutfitsToAPI(ship.Outfits),
	}

	return apiShip
}

// convertWeaponsToAPI converts weapon IDs to API Weapon objects
func convertWeaponsToAPI(weaponIDs []string, weaponAmmo map[int]int) []*api.Weapon {
	weapons := make([]*api.Weapon, 0, len(weaponIDs))
	for slotIndex, weaponID := range weaponIDs {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			// Get current ammo for this weapon slot
			currentAmmo := 0
			if weaponAmmo != nil {
				if ammo, ok := weaponAmmo[slotIndex]; ok {
					currentAmmo = ammo
				}
			}

			weapons = append(weapons, &api.Weapon{
				WeaponID:   weapon.ID,
				WeaponType: weapon.Type,
				Damage:     int32(weapon.Damage),
				Range:      int32(weapon.RangeValue),
				Accuracy:   float64(weapon.Accuracy),
				Ammo:       int32(currentAmmo),
				MaxAmmo:    int32(weapon.AmmoCapacity),
				Cooldown:   int32(weapon.Cooldown),
			})
		}
	}
	return weapons
}

// convertOutfitsToAPI converts outfit IDs to API Outfit objects
func convertOutfitsToAPI(outfitIDs []string) []*api.Outfit {
	outfits := make([]*api.Outfit, 0, len(outfitIDs))
	for _, outfitID := range outfitIDs {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			outfits = append(outfits, &api.Outfit{
				OutfitID:    outfit.ID,
				OutfitType:  outfit.Type,
				Name:        outfit.Name,
				Description: outfit.Description,
				Modifiers:   convertOutfitModifiers(outfit),
			})
		}
	}
	return outfits
}

// convertOutfitModifiers converts outfit bonuses to API format
func convertOutfitModifiers(outfit *models.Outfit) map[string]int32 {
	modifiers := make(map[string]int32)
	if outfit.ShieldBonus > 0 {
		modifiers["shield_bonus"] = int32(outfit.ShieldBonus)
	}
	if outfit.HullBonus > 0 {
		modifiers["hull_bonus"] = int32(outfit.HullBonus)
	}
	if outfit.CargoBonus > 0 {
		modifiers["cargo_bonus"] = int32(outfit.CargoBonus)
	}
	if outfit.FuelBonus > 0 {
		modifiers["fuel_bonus"] = int32(outfit.FuelBonus)
	}
	if outfit.SpeedBonus > 0 {
		modifiers["speed_bonus"] = int32(outfit.SpeedBonus)
	}
	return modifiers
}

// convertInventoryToAPI converts ship cargo to API Inventory
func convertInventoryToAPI(ship *models.Ship) *api.Inventory {
	if ship == nil {
		return nil
	}

	// Get ship type for cargo space
	shipType := models.GetShipTypeByID(ship.TypeID)
	totalCargoSpace := int32(0)
	if shipType != nil {
		totalCargoSpace = int32(shipType.CargoSpace)
	}

	inventory := &api.Inventory{
		Cargo:           make(map[string]int32),
		Items:           make([]*api.Item, 0),
		TotalCargoSpace: totalCargoSpace,
		CargoUsed:       int32(ship.GetCargoUsed()),
	}

	// Convert cargo (ship.Cargo is []CargoItem, not map)
	for _, cargoItem := range ship.Cargo {
		inventory.Cargo[cargoItem.CommodityID] = int32(cargoItem.Quantity)
	}

	// Items array empty (item system not yet implemented)

	return inventory
}

// convertPlayerStatsToAPI converts database player stats to API PlayerStats
func convertPlayerStatsToAPI(player *models.Player) *api.PlayerStats {
	if player == nil {
		return nil
	}

	return &api.PlayerStats{
		Level:              int32(player.Level),
		Experience:         player.Experience,
		TotalCreditsEarned: player.TradeProfit, // Using TradeProfit as proxy
		CombatRating:       int32(player.CombatRating),
		TradeRating:        int32(player.TradingRating),
		ExplorationRating:  int32(player.ExplorationRating),
		ShipsDestroyed:     int32(player.TotalKills),
		MissionsCompleted:  int32(player.MissionsCompleted),
		QuestsCompleted:    int32(player.QuestsCompleted),
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
		LegalStatus:       player.LegalStatus,
		Bounty:            player.Bounty,
	}

	// Convert faction reputation (Player.Reputation is map[string]int)
	for faction, rep := range player.Reputation {
		reputation.FactionReputation[faction] = int32(rep)
	}

	return reputation
}

// convertMarketToAPI converts database market data to API Market
func convertMarketToAPI(prices []models.MarketPrice, commodities map[string]*models.Commodity, governmentID string) *api.Market {
	market := &api.Market{
		Commodities: make([]*api.CommodityListing, 0, len(prices)),
		// LastUpdated will be set from latest price update
	}

	for _, price := range prices {
		commodity := commodities[price.CommodityID]
		if commodity == nil {
			continue // Skip if commodity definition not found
		}

		// Check if commodity is illegal in this system's government
		isIllegal := false
		if governmentID != "" {
			for _, illegalGov := range commodity.IllegalIn {
				if illegalGov == governmentID {
					isIllegal = true
					break
				}
			}
		}

		market.Commodities = append(market.Commodities, &api.CommodityListing{
			CommodityID: price.CommodityID,
			Name:        commodity.Name,
			BuyPrice:    int32(price.BuyPrice),
			SellPrice:   int32(price.SellPrice),
			Stock:       int32(price.Stock),
			IsIllegal:   isIllegal,
		})
	}

	return market
}

// convertPlanetToAPI converts database planet to API Planet
func convertPlanetToAPI(planet *models.Planet, governmentID string) *api.Planet {
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
		Government:  governmentID,
		Population:  planet.Population,
	}
}

// convertMissionToAPI converts database mission to API Mission
func convertMissionToAPI(mission *models.Mission, systemRepo *database.SystemRepository, ctx context.Context) *api.Mission {
	if mission == nil {
		return nil
	}

	// Calculate total reputation reward
	var reputationTotal int32
	for _, repChange := range mission.ReputationChange {
		reputationTotal += int32(repChange)
	}

	// Look up origin planet to get system ID
	originSystemID := uuid.Nil
	if mission.OriginPlanet != uuid.Nil {
		if originPlanet, err := systemRepo.GetPlanetByID(ctx, mission.OriginPlanet); err == nil && originPlanet != nil {
			originSystemID = originPlanet.SystemID
		}
	}

	// Look up destination planet to get system ID (if destination is a planet)
	destinationSystemID := uuid.Nil
	if mission.Destination != nil && *mission.Destination != uuid.Nil {
		if destPlanet, err := systemRepo.GetPlanetByID(ctx, *mission.Destination); err == nil && destPlanet != nil {
			destinationSystemID = destPlanet.SystemID
		}
	}

	apiMission := &api.Mission{
		MissionID:           mission.ID,
		Title:               mission.Title,
		Description:         mission.Description,
		RewardCredits:       mission.Reward,
		RewardReputation:    reputationTotal,
		OriginSystemID:      originSystemID,
		DestinationSystemID: destinationSystemID,
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

	// Convert objectives from quest.Objectives array
	for _, obj := range quest.Objectives {
		if obj != nil {
			apiObjective := &api.QuestObjective{
				ObjectiveID:      obj.ID,
				Description:      obj.Description,
				ProgressCurrent:  int32(obj.Current),
				ProgressRequired: int32(obj.Required),
				Completed:        obj.Completed,
			}

			// Convert objective type
			switch obj.Type {
			case models.ObjectiveKill, models.ObjectiveDestroy:
				apiObjective.Type = api.ObjectiveTypeDestroy
			case models.ObjectiveDeliver:
				apiObjective.Type = api.ObjectiveTypeDeliver
			case models.ObjectiveTravel, models.ObjectiveInvestigate:
				apiObjective.Type = api.ObjectiveTypeTravel
			case models.ObjectiveCollect, models.ObjectiveMine:
				apiObjective.Type = api.ObjectiveTypeCollect
			default:
				apiObjective.Type = api.ObjectiveTypeDeliver
			}

			apiQuest.Objectives = append(apiQuest.Objectives, apiObjective)
		}
	}

	// Convert rewards from quest.Rewards (singular struct)
	if quest.Rewards.Credits > 0 {
		apiQuest.Rewards = append(apiQuest.Rewards, &api.QuestReward{
			Type:  api.RewardTypeCredits,
			Value: quest.Rewards.Credits,
		})
	}

	if quest.Rewards.Experience > 0 {
		apiQuest.Rewards = append(apiQuest.Rewards, &api.QuestReward{
			Type:  api.RewardTypeExperience,
			Value: int64(quest.Rewards.Experience),
		})
	}

	// Add reputation rewards
	for _, amount := range quest.Rewards.Reputation {
		if amount != 0 {
			apiQuest.Rewards = append(apiQuest.Rewards, &api.QuestReward{
				Type:  api.RewardTypeReputation,
				Value: int64(amount),
			})
		}
	}

	// Add item rewards
	for itemID, quantity := range quest.Rewards.Items {
		apiQuest.Rewards = append(apiQuest.Rewards, &api.QuestReward{
			Type:   api.RewardTypeItem,
			ItemID: itemID,
			Value:  int64(quantity),
		})
	}

	// Add special rewards
	if quest.Rewards.ShipUnlock != "" {
		apiQuest.Rewards = append(apiQuest.Rewards, &api.QuestReward{
			Type:   api.RewardTypeUnlock,
			ItemID: quest.Rewards.ShipUnlock,
			Value:  1,
		})
	}

	return apiQuest
}
