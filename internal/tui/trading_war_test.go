// File: internal/tui/trading_war_test.go
// Project: Terminal Velocity
// Description: Tests for war-economy price overlay in the trading
//   market-loaded handler.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-24

package tui

import (
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/factionwar"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// warFixtureModel returns a Model with just enough wiring to exercise
// applyWarEconomyToPrices: a real factionwar manager, a current
// system, and nothing else that would require DB handles.
func warFixtureModel(systemName string) (Model, *factionwar.Manager) {
	mgr := factionwar.NewManager(nil) // no news bus — keeps the test self-contained
	sys := &models.StarSystem{
		ID:   uuid.New(),
		Name: systemName,
	}
	return Model{
		factionWarManager: mgr,
		currentSystem:     sys,
	}, mgr
}

func TestApplyWarEconomyToPricesNoManager(t *testing.T) {
	// Sanity: without a manager, prices pass through untouched.
	prices := []*models.MarketPrice{
		{CommodityID: "weapons", BuyPrice: 100, SellPrice: 150},
	}
	commodities := []models.Commodity{
		{ID: "weapons", Category: "weapons"},
	}
	m := Model{currentSystem: &models.StarSystem{Name: "Sol"}}
	got := m.applyWarEconomyToPrices(prices, commodities)

	// Same slice identity expected (early-return path).
	if &got[0] == nil || got[0].BuyPrice != 100 || got[0].SellPrice != 150 {
		t.Errorf("nil manager: prices mutated, got %+v", got[0])
	}
}

func TestApplyWarEconomyToPricesPeacetime(t *testing.T) {
	// With a manager but no active wars, prices should pass through.
	m, _ := warFixtureModel("Sol")
	prices := []*models.MarketPrice{
		{CommodityID: "weapons", BuyPrice: 100, SellPrice: 150},
	}
	commodities := []models.Commodity{
		{ID: "weapons", Category: "weapons"},
	}
	got := m.applyWarEconomyToPrices(prices, commodities)
	if got[0].BuyPrice != 100 || got[0].SellPrice != 150 {
		t.Errorf("peacetime prices changed: got %+v, want unchanged", got[0])
	}
}

func TestApplyWarEconomyToPricesAmplifiesWarMaterial(t *testing.T) {
	m, mgr := warFixtureModel("Sol")
	// Declare a war that covers Sol.
	a := &models.NPCFaction{ID: "a", Name: "A", CoreSystems: []string{"Sol"}}
	b := &models.NPCFaction{ID: "b", Name: "B", CoreSystems: []string{"Wolf 359"}}
	_, err := mgr.DeclareWar(a, b, "")
	if err != nil {
		t.Fatalf("DeclareWar: %v", err)
	}

	prices := []*models.MarketPrice{
		{CommodityID: "weapons", BuyPrice: 100, SellPrice: 150},
		{CommodityID: "medical", BuyPrice: 50, SellPrice: 80},
		{CommodityID: "food", BuyPrice: 30, SellPrice: 50}, // not a war material
	}
	commodities := []models.Commodity{
		{ID: "weapons", Category: "weapons"},
		{ID: "medical", Category: "medical"},
		{ID: "food", Category: "food"},
	}
	got := m.applyWarEconomyToPrices(prices, commodities)

	// 1.4 multiplier:
	//   100 × 1.4 = 140
	//   150 × 1.4 = 210
	//   50 × 1.4 = 70
	//   80 × 1.4 = 112
	if got[0].BuyPrice != 140 || got[0].SellPrice != 210 {
		t.Errorf("weapons not amplified: got %+v, want buy 140, sell 210", got[0])
	}
	if got[1].BuyPrice != 70 || got[1].SellPrice != 112 {
		t.Errorf("medical not amplified: got %+v, want buy 70, sell 112", got[1])
	}
	// Food is not a war material — unchanged at its original 30/50.
	if got[2].BuyPrice != 30 || got[2].SellPrice != 50 {
		t.Errorf("food incorrectly amplified: got %+v, want buy 30, sell 50", got[2])
	}
}

func TestApplyWarEconomyToPricesDoesNotMutateInput(t *testing.T) {
	m, mgr := warFixtureModel("Sol")
	a := &models.NPCFaction{ID: "a", Name: "A", CoreSystems: []string{"Sol"}}
	b := &models.NPCFaction{ID: "b", Name: "B", CoreSystems: []string{"Wolf 359"}}
	_, _ = mgr.DeclareWar(a, b, "")

	orig := &models.MarketPrice{CommodityID: "weapons", BuyPrice: 100, SellPrice: 150}
	prices := []*models.MarketPrice{orig}
	commodities := []models.Commodity{{ID: "weapons", Category: "weapons"}}

	_ = m.applyWarEconomyToPrices(prices, commodities)

	// Caller's underlying struct must still show the original prices
	// even after we've amplified a copy. Database state and any
	// cached views of the DB rows stay consistent.
	if orig.BuyPrice != 100 || orig.SellPrice != 150 {
		t.Errorf("input mutated: got %+v, want buy 100 sell 150", orig)
	}
}

func TestApplyWarEconomyToPricesUnknownCommodityPassesThrough(t *testing.T) {
	// If a market row references a commodity that isn't in the
	// loaded commodities list, we can't classify it, so leave it
	// alone rather than risking an incorrect categorization.
	m, mgr := warFixtureModel("Sol")
	a := &models.NPCFaction{ID: "a", Name: "A", CoreSystems: []string{"Sol"}}
	b := &models.NPCFaction{ID: "b", Name: "B", CoreSystems: []string{"Wolf 359"}}
	_, _ = mgr.DeclareWar(a, b, "")

	prices := []*models.MarketPrice{
		{CommodityID: "mystery", BuyPrice: 100, SellPrice: 200},
	}
	got := m.applyWarEconomyToPrices(prices, nil)
	if got[0].BuyPrice != 100 || got[0].SellPrice != 200 {
		t.Errorf("unknown commodity should pass through, got %+v", got[0])
	}
}
