// File: internal/tui/outfitter_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced outfitter UI with loadout management
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

// Views for enhanced outfitter
const (
	outfitterViewBrowser  = "browser"
	outfitterViewSlots    = "slots"
	outfitterViewLoadouts = "loadouts"
	outfitterViewInventory = "inventory"
)

type outfitterEnhancedModel struct {
	viewMode         string
	cursor           int
	category         models.EquipmentCategory
	selectedEquipment *models.Equipment
	selectedSlot     *models.EquipmentSlot
	currentLoadout   *models.ShipLoadout
	loadouts         []*models.ShipLoadout
	inventory        map[string]int
}

func newOutfitterEnhancedModel() outfitterEnhancedModel {
	return outfitterEnhancedModel{
		viewMode:  outfitterViewBrowser,
		cursor:    0,
		category:  models.CategoryWeapon,
		inventory: make(map[string]int),
	}
}

func (m Model) updateOutfitterEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.outfitterEnhanced.viewMode {
		case outfitterViewBrowser:
			return m.updateOutfitterBrowser(msg)
		case outfitterViewSlots:
			return m.updateOutfitterSlots(msg)
		case outfitterViewLoadouts:
			return m.updateOutfitterLoadouts(msg)
		case outfitterViewInventory:
			return m.updateOutfitterInventory(msg)
		}
	}

	return m, nil
}

func (m Model) updateOutfitterBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.screen = ScreenMainMenu
		return m, nil

	case "1":
		m.outfitterEnhanced.viewMode = outfitterViewBrowser
		m.outfitterEnhanced.cursor = 0
		return m, nil

	case "2":
		m.outfitterEnhanced.viewMode = outfitterViewSlots
		m.outfitterEnhanced.cursor = 0
		return m, nil

	case "3":
		m.outfitterEnhanced.viewMode = outfitterViewLoadouts
		m.outfitterEnhanced.cursor = 0
		return m, nil

	case "4":
		m.outfitterEnhanced.viewMode = outfitterViewInventory
		m.outfitterEnhanced.cursor = 0
		return m, nil

	case "tab":
		// Cycle through categories
		categories := []models.EquipmentCategory{
			models.CategoryWeapon,
			models.CategoryDefense,
			models.CategoryPower,
			models.CategoryPropulsion,
			models.CategoryUtility,
		}
		for i, cat := range categories {
			if cat == m.outfitterEnhanced.category {
				m.outfitterEnhanced.category = categories[(i+1)%len(categories)]
				m.outfitterEnhanced.cursor = 0
				break
			}
		}
		return m, nil

	case "up", "k":
		if m.outfitterEnhanced.cursor > 0 {
			m.outfitterEnhanced.cursor--
		}
		return m, nil

	case "down", "j":
		equipment := m.outfittingManager.GetEquipmentByCategory(m.outfitterEnhanced.category)
		if m.outfitterEnhanced.cursor < len(equipment)-1 {
			m.outfitterEnhanced.cursor++
		}
		return m, nil

	case "enter", " ":
		// Purchase equipment
		equipment := m.outfittingManager.GetEquipmentByCategory(m.outfitterEnhanced.category)
		if m.outfitterEnhanced.cursor < len(equipment) {
			selected := equipment[m.outfitterEnhanced.cursor]
			err := m.outfittingManager.PurchaseEquipment(m.playerID, selected.ID, 1, m.player.Credits)
			if err == nil {
				m.player.Credits -= selected.Price
				m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateOutfitterSlots(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.outfitterEnhanced.viewMode = outfitterViewBrowser
		return m, nil

	case "up", "k":
		if m.outfitterEnhanced.cursor > 0 {
			m.outfitterEnhanced.cursor--
		}
		return m, nil

	case "down", "j":
		if m.outfitterEnhanced.currentLoadout != nil {
			if m.outfitterEnhanced.cursor < len(m.outfitterEnhanced.currentLoadout.Slots)-1 {
				m.outfitterEnhanced.cursor++
			}
		}
		return m, nil

	case "enter", " ":
		// Install/uninstall equipment in slot
		if m.outfitterEnhanced.currentLoadout != nil && m.outfitterEnhanced.cursor < len(m.outfitterEnhanced.currentLoadout.Slots) {
			slot := &m.outfitterEnhanced.currentLoadout.Slots[m.outfitterEnhanced.cursor]

			if slot.IsEmpty() {
				// Switch to inventory to select equipment
				m.outfitterEnhanced.selectedSlot = slot
				m.outfitterEnhanced.viewMode = outfitterViewInventory
			} else {
				// Uninstall
				err := m.outfittingManager.UninstallEquipment(m.playerID, m.outfitterEnhanced.currentLoadout.ID, slot.ID)
				if err == nil {
					m.outfitterEnhanced.currentLoadout, _ = m.outfittingManager.GetLoadout(m.outfitterEnhanced.currentLoadout.ID)
					m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
				}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateOutfitterLoadouts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.outfitterEnhanced.viewMode = outfitterViewBrowser
		return m, nil

	case "up", "k":
		if m.outfitterEnhanced.cursor > 0 {
			m.outfitterEnhanced.cursor--
		}
		return m, nil

	case "down", "j":
		if m.outfitterEnhanced.cursor < len(m.outfitterEnhanced.loadouts)-1 {
			m.outfitterEnhanced.cursor++
		}
		return m, nil

	case "enter", " ":
		// Load selected loadout
		if m.outfitterEnhanced.cursor < len(m.outfitterEnhanced.loadouts) {
			m.outfitterEnhanced.currentLoadout = m.outfitterEnhanced.loadouts[m.outfitterEnhanced.cursor]
			m.outfitterEnhanced.viewMode = outfitterViewSlots
		}
		return m, nil

	case "n":
		// Create new loadout
		if m.currentShip != nil {
			shipType := models.GetShipTypeByID(m.currentShip.TypeID)
			if shipType != nil {
				loadout, err := m.outfittingManager.CreateLoadout(m.playerID, shipType, "New Loadout")
				if err == nil {
					m.outfitterEnhanced.loadouts = m.outfittingManager.GetPlayerLoadouts(m.playerID)
					m.outfitterEnhanced.currentLoadout = loadout
					m.outfitterEnhanced.viewMode = outfitterViewSlots
				}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateOutfitterInventory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.outfitterEnhanced.selectedSlot != nil {
			m.outfitterEnhanced.selectedSlot = nil
			m.outfitterEnhanced.viewMode = outfitterViewSlots
		} else {
			m.outfitterEnhanced.viewMode = outfitterViewBrowser
		}
		return m, nil

	case "up", "k":
		if m.outfitterEnhanced.cursor > 0 {
			m.outfitterEnhanced.cursor--
		}
		return m, nil

	case "down", "j":
		if m.outfitterEnhanced.cursor < len(m.outfitterEnhanced.inventory)-1 {
			m.outfitterEnhanced.cursor++
		}
		return m, nil

	case "enter", " ":
		// Install selected equipment or sell
		if m.outfitterEnhanced.selectedSlot != nil {
			// Install in slot
			equipmentIDs := make([]string, 0, len(m.outfitterEnhanced.inventory))
			for id := range m.outfitterEnhanced.inventory {
				equipmentIDs = append(equipmentIDs, id)
			}

			if m.outfitterEnhanced.cursor < len(equipmentIDs) {
				equipmentID := equipmentIDs[m.outfitterEnhanced.cursor]
				err := m.outfittingManager.InstallEquipment(
					m.playerID,
					m.outfitterEnhanced.currentLoadout.ID,
					m.outfitterEnhanced.selectedSlot.ID,
					equipmentID,
				)
				if err == nil {
					m.outfitterEnhanced.currentLoadout, _ = m.outfittingManager.GetLoadout(m.outfitterEnhanced.currentLoadout.ID)
					m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
					m.outfitterEnhanced.selectedSlot = nil
					m.outfitterEnhanced.viewMode = outfitterViewSlots
				}
			}
		}
		return m, nil

	case "s":
		// Sell equipment
		equipmentIDs := make([]string, 0, len(m.outfitterEnhanced.inventory))
		for id := range m.outfitterEnhanced.inventory {
			equipmentIDs = append(equipmentIDs, id)
		}

		if m.outfitterEnhanced.cursor < len(equipmentIDs) {
			equipmentID := equipmentIDs[m.outfitterEnhanced.cursor]
			credits, err := m.outfittingManager.SellEquipment(m.playerID, equipmentID, 1)
			if err == nil {
				m.player.Credits += credits
				m.outfitterEnhanced.inventory = m.outfittingManager.GetPlayerInventory(m.playerID)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewOutfitterEnhanced() string {
	locationName := "Space Station"
	s := renderHeader(m.username, m.player.Credits, locationName)
	s += "\n"

	s += subtitleStyle.Render("=== Advanced Outfitter ===") + "\n\n"

	// Tab navigation
	tabs := []string{
		"Equipment Browser (1)",
		"Ship Slots (2)",
		"Loadouts (3)",
		"Inventory (4)",
	}

	tabDisplay := ""
	currentTab := 0
	switch m.outfitterEnhanced.viewMode {
	case outfitterViewBrowser:
		currentTab = 0
	case outfitterViewSlots:
		currentTab = 1
	case outfitterViewLoadouts:
		currentTab = 2
	case outfitterViewInventory:
		currentTab = 3
	}

	for i, tab := range tabs {
		if i == currentTab {
			tabDisplay += selectedMenuItemStyle.Render(tab)
		} else {
			tabDisplay += helpStyle.Render(tab)
		}
		if i < len(tabs)-1 {
			tabDisplay += "  |  "
		}
	}
	s += tabDisplay + "\n"
	s += strings.Repeat("─", 78) + "\n\n"

	// View-specific content
	switch m.outfitterEnhanced.viewMode {
	case outfitterViewBrowser:
		s += m.viewOutfitterEquipmentBrowser()
	case outfitterViewSlots:
		s += m.viewOutfitterShipSlots()
	case outfitterViewLoadouts:
		s += m.viewOutfitterLoadoutsList()
	case outfitterViewInventory:
		s += m.viewOutfitterPlayerInventory()
	}

	return s
}

func (m Model) viewOutfitterEquipmentBrowser() string {
	s := ""

	// Category selection
	categories := []string{"Weapons", "Defense", "Power", "Propulsion", "Utility"}
	categoryMap := map[string]models.EquipmentCategory{
		"Weapons":    models.CategoryWeapon,
		"Defense":    models.CategoryDefense,
		"Power":      models.CategoryPower,
		"Propulsion": models.CategoryPropulsion,
		"Utility":    models.CategoryUtility,
	}

	s += "Category: "
	for i, cat := range categories {
		if categoryMap[cat] == m.outfitterEnhanced.category {
			s += selectedMenuItemStyle.Render(cat)
		} else {
			s += helpStyle.Render(cat)
		}
		if i < len(categories)-1 {
			s += " | "
		}
	}
	s += "\n\n"

	// Equipment list
	equipment := m.outfittingManager.GetEquipmentByCategory(m.outfitterEnhanced.category)

	if len(equipment) == 0 {
		s += helpStyle.Render("No equipment available in this category") + "\n\n"
		s += renderFooter("Tab: Category  •  ESC: Main Menu")
		return s
	}

	// Table header
	s += fmt.Sprintf("%-25s %-8s %-10s %-12s %s\n", "Name", "Size", "Space", "Price", "Primary Stats")
	s += strings.Repeat("─", 78) + "\n"

	for i, eq := range equipment {
		affordable := eq.Price <= m.player.Credits
		statsStr := m.getEquipmentStatsString(eq)

		line := fmt.Sprintf("%-25s %-8s %-10d %-12s %s",
			eq.Name,
			m.getSlotSizeName(eq.SlotSize),
			eq.OutfitSpace,
			formatCredits(eq.Price),
			statsStr,
		)

		if !affordable {
			line = helpStyle.Render(line)
		}

		if i == m.outfitterEnhanced.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Purchase  •  Tab: Category  •  ESC: Menu")
	return s
}

func (m Model) viewOutfitterShipSlots() string {
	s := ""

	if m.outfitterEnhanced.currentLoadout == nil {
		s += helpStyle.Render("No loadout selected") + "\n\n"
		s += helpStyle.Render("Go to Loadouts (3) to create or select a loadout") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	loadout := m.outfitterEnhanced.currentLoadout
	shipType := models.GetShipTypeByID(loadout.ShipTypeID)

	// Loadout info
	s += fmt.Sprintf("Loadout: %s\n", statsStyle.Render(loadout.Name))
	if shipType != nil {
		s += fmt.Sprintf("Ship Class: %s\n", shipType.Name)
	}
	s += fmt.Sprintf("Outfit Space: %s / %d\n\n",
		statsStyle.Render(fmt.Sprintf("%d", loadout.UsedOutfitSpace)),
		shipType.OutfitSpace)

	// Combined stats
	combined := loadout.GetCombinedStats()
	s += "Total Bonuses: "
	bonuses := []string{}
	if combined.Damage > 0 {
		bonuses = append(bonuses, fmt.Sprintf("+%d dmg", combined.Damage))
	}
	if combined.ShieldHP > 0 {
		bonuses = append(bonuses, fmt.Sprintf("+%d shields", combined.ShieldHP))
	}
	if combined.SpeedBonus > 0 {
		bonuses = append(bonuses, fmt.Sprintf("+%d speed", combined.SpeedBonus))
	}
	if combined.CargoBonus > 0 {
		bonuses = append(bonuses, fmt.Sprintf("+%d cargo", combined.CargoBonus))
	}
	if len(bonuses) > 0 {
		s += statsStyle.Render(strings.Join(bonuses, ", "))
	} else {
		s += helpStyle.Render("None")
	}
	s += "\n\n"

	// Slots list
	s += "Slots:\n"
	s += strings.Repeat("─", 78) + "\n"

	for i, slot := range loadout.Slots {
		slotInfo := fmt.Sprintf("[%s-%s]", m.getSlotTypeName(slot.SlotType), m.getSlotSizeName(slot.SlotSize))

		var line string
		if slot.IsEmpty() {
			line = fmt.Sprintf("%-20s %s", slotInfo, helpStyle.Render("Empty"))
		} else {
			eq := slot.InstalledItem
			line = fmt.Sprintf("%-20s %s (%s)", slotInfo, eq.Name, m.getEquipmentStatsString(eq))
		}

		if i == m.outfitterEnhanced.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Install/Remove  •  ESC: Back")
	return s
}

func (m Model) viewOutfitterLoadoutsList() string {
	s := ""

	s += "Saved Loadouts:\n\n"

	if len(m.outfitterEnhanced.loadouts) == 0 {
		s += helpStyle.Render("No saved loadouts") + "\n\n"
		s += helpStyle.Render("Press 'N' to create a new loadout") + "\n\n"
		s += renderFooter("N: New Loadout  •  ESC: Back")
		return s
	}

	for i, loadout := range m.outfitterEnhanced.loadouts {
		shipType := models.GetShipTypeByID(loadout.ShipTypeID)
		shipName := "Unknown"
		if shipType != nil {
			shipName = shipType.Name
		}

		line := fmt.Sprintf("%s - %s (Cost: %s cr, Space: %d)",
			loadout.Name,
			shipName,
			formatCredits(loadout.TotalCost),
			loadout.UsedOutfitSpace,
		)

		if i == m.outfitterEnhanced.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Load  •  N: New  •  ESC: Back")
	return s
}

func (m Model) viewOutfitterPlayerInventory() string {
	s := ""

	if m.outfitterEnhanced.selectedSlot != nil {
		s += fmt.Sprintf("Installing to: %s slot (size %s)\n\n",
			m.getSlotTypeName(m.outfitterEnhanced.selectedSlot.SlotType),
			m.getSlotSizeName(m.outfitterEnhanced.selectedSlot.SlotSize))
	} else {
		s += "Your Equipment Inventory:\n\n"
	}

	if len(m.outfitterEnhanced.inventory) == 0 {
		s += helpStyle.Render("No equipment in inventory") + "\n\n"
		s += helpStyle.Render("Purchase equipment from the browser to see it here") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	i := 0
	for equipmentID, quantity := range m.outfitterEnhanced.inventory {
		equipment, err := m.outfittingManager.GetEquipment(equipmentID)
		if err != nil {
			continue
		}

		canInstall := true
		if m.outfitterEnhanced.selectedSlot != nil {
			canInstall = m.outfitterEnhanced.selectedSlot.CanInstall(equipment)
		}

		line := fmt.Sprintf("%s x%d - %s (%s)",
			equipment.Name,
			quantity,
			m.getSlotTypeName(equipment.SlotType),
			m.getEquipmentStatsString(equipment))

		if !canInstall {
			line = helpStyle.Render(line)
		}

		if i == m.outfitterEnhanced.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}

		i++
	}

	if m.outfitterEnhanced.selectedSlot != nil {
		s += "\n" + renderFooter("↑/↓: Select  •  Enter: Install  •  ESC: Cancel")
	} else {
		s += "\n" + renderFooter("↑/↓: Select  •  S: Sell (70%)  •  ESC: Back")
	}

	return s
}

// Helper functions

func (m Model) getSlotTypeName(slotType models.EquipmentSlotType) string {
	names := map[models.EquipmentSlotType]string{
		models.SlotWeapon:  "Weapon",
		models.SlotShield:  "Shield",
		models.SlotEngine:  "Engine",
		models.SlotReactor: "Reactor",
		models.SlotUtility: "Utility",
		models.SlotSpecial: "Special",
	}
	if name, exists := names[slotType]; exists {
		return name
	}
	return string(slotType)
}

func (m Model) getSlotSizeName(size int) string {
	sizes := []string{"", "Small", "Medium", "Large", "Capital"}
	if size >= 0 && size < len(sizes) {
		return sizes[size]
	}
	return fmt.Sprintf("Size-%d", size)
}

func (m Model) getEquipmentStatsString(eq *models.Equipment) string {
	stats := eq.Stats
	parts := []string{}

	if stats.Damage > 0 {
		parts = append(parts, fmt.Sprintf("%d dmg", stats.Damage))
	}
	if stats.ShieldHP > 0 {
		parts = append(parts, fmt.Sprintf("%d shields", stats.ShieldHP))
	}
	if stats.EnergyOutput > 0 {
		parts = append(parts, fmt.Sprintf("%d energy", stats.EnergyOutput))
	}
	if stats.SpeedBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d speed", stats.SpeedBonus))
	}
	if stats.CargoBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d cargo", stats.CargoBonus))
	}
	if stats.FuelBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d fuel", stats.FuelBonus))
	}

	if len(parts) == 0 {
		return "No effects"
	}

	return strings.Join(parts, ", ")
}
