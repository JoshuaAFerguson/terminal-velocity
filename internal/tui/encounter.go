// File: internal/tui/encounter.go
// Project: Terminal Velocity
// Description: Terminal UI component for encounter
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/encounters"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type encounterModel struct {
	encounter *models.Encounter
	generator *encounters.Generator
	cursor    int
	message   string
	resolved  bool
}

func newEncounterModel() encounterModel {
	return encounterModel{
		cursor:    0,
		generator: encounters.NewGenerator(),
		resolved:  false,
	}
}

func (m Model) updateEncounter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Can't escape active encounter
			if m.encounterModel.resolved {
				m.screen = ScreenMainMenu
				return m, nil
			}
			return m, nil

		case "up", "k":
			if m.encounterModel.cursor > 0 {
				m.encounterModel.cursor--
			}

		case "down", "j":
			if m.encounterModel.encounter != nil {
				options := m.encounterModel.encounter.GetOptions(m.player)
				if m.encounterModel.cursor < len(options)-1 {
					m.encounterModel.cursor++
				}
			}

		case "enter", " ":
			return m.handleEncounterOption()
		}
	}

	return m, nil
}

func (m Model) handleEncounterOption() (tea.Model, tea.Cmd) {
	if m.encounterModel.encounter == nil || m.player == nil {
		return m, nil
	}

	options := m.encounterModel.encounter.GetOptions(m.player)
	if m.encounterModel.cursor >= len(options) {
		return m, nil
	}

	selectedOption := options[m.encounterModel.cursor]

	// Check if player can afford option
	if !m.encounterModel.encounter.CanAffordOption(selectedOption, m.player) {
		m.encounterModel.message = "You don't meet the requirements for this option"
		return m, nil
	}

	// Handle option effects
	switch selectedOption.ID {
	case "engage", "attack":
		// Start combat
		m.encounterModel.encounter.Resolve()
		m.encounterModel.resolved = true

		// Initialize combat with encounter ships
		m.combat = newCombatModel()
		m.combat.playerShip = m.currentShip
		m.combat.playerType = models.GetShipTypeByID(m.currentShip.TypeID)

		// Generate encounter ships
		m.combat.enemyShips = m.encounterModel.generator.GenerateEncounterShips(m.encounterModel.encounter)
		m.combat.enemyTypes = make(map[string]*models.ShipType)
		for _, ship := range m.combat.enemyShips {
			m.combat.enemyTypes[ship.TypeID] = models.GetShipTypeByID(ship.TypeID)
		}

		m.screen = ScreenCombat
		return m, nil

	case "flee":
		// Attempt to flee
		success := m.encounterModel.generator.CalculateFleeSuccess(
			m.currentShip,
			models.GetShipTypeByID(m.currentShip.TypeID),
			m.encounterModel.generator.GenerateEncounterShips(m.encounterModel.encounter),
		)

		if success {
			m.encounterModel.message = "You successfully escaped!"
			m.encounterModel.encounter.Flee()
			m.encounterModel.resolved = true
		} else {
			// Failed flee -> combat
			m.encounterModel.message = "Escape failed! They've caught up to you!"
			m.encounterModel.encounter.Resolve()
			m.encounterModel.resolved = true

			// Initialize combat
			m.combat = newCombatModel()
			m.combat.playerShip = m.currentShip
			m.combat.playerType = models.GetShipTypeByID(m.currentShip.TypeID)
			m.combat.enemyShips = m.encounterModel.generator.GenerateEncounterShips(m.encounterModel.encounter)
			m.combat.enemyTypes = make(map[string]*models.ShipType)
			for _, ship := range m.combat.enemyShips {
				m.combat.enemyTypes[ship.TypeID] = models.GetShipTypeByID(ship.TypeID)
			}

			m.screen = ScreenCombat
			return m, nil
		}

	case "trade":
		// Buy goods from trader
		if m.player.CanAfford(selectedOption.CostCredits) {
			m.player.AddCredits(-selectedOption.CostCredits)

			// Add cargo
			if m.encounterModel.encounter.CargoReward != "" {
				m.currentShip.AddCargo(m.encounterModel.encounter.CargoReward, m.encounterModel.encounter.CargoQuantity)
			}

			m.encounterModel.message = fmt.Sprintf("Trade complete! Acquired %d tons of %s",
				m.encounterModel.encounter.CargoQuantity, m.encounterModel.encounter.CargoReward)
			m.encounterModel.encounter.Resolve()
			m.encounterModel.resolved = true

			// Check achievements
			m.checkAchievements()
		} else {
			m.encounterModel.message = "Not enough credits for this trade"
		}

	case "rescue":
		// Help distressed ship
		m.player.AddCredits(m.encounterModel.encounter.CreditReward)
		if m.encounterModel.encounter.FactionID != "" {
			m.player.ModifyReputation(m.encounterModel.encounter.FactionID, m.encounterModel.encounter.ReputationGain)
		}

		m.encounterModel.message = fmt.Sprintf("Rescue complete! Earned %d credits and +%d reputation",
			m.encounterModel.encounter.CreditReward, m.encounterModel.encounter.ReputationGain)
		m.encounterModel.encounter.Resolve()
		m.encounterModel.resolved = true

		// Check achievements
		m.checkAchievements()

	case "cooperate":
		// Police scan
		if m.player.IsCriminal {
			m.encounterModel.message = "The police detected your criminal status and are moving to arrest you!"
			m.encounterModel.encounter.Hostile = true

			// Start combat
			m.combat = newCombatModel()
			m.combat.playerShip = m.currentShip
			m.combat.playerType = models.GetShipTypeByID(m.currentShip.TypeID)
			m.combat.enemyShips = m.encounterModel.generator.GenerateEncounterShips(m.encounterModel.encounter)
			m.combat.enemyTypes = make(map[string]*models.ShipType)
			for _, ship := range m.combat.enemyShips {
				m.combat.enemyTypes[ship.TypeID] = models.GetShipTypeByID(ship.TypeID)
			}

			m.encounterModel.encounter.Resolve()
			m.encounterModel.resolved = true
			m.screen = ScreenCombat
			return m, nil
		} else {
			m.encounterModel.message = "Scan complete. You're clear to proceed."
			m.encounterModel.encounter.Resolve()
			m.encounterModel.resolved = true
		}

	case "bribe":
		// Bribe police
		if m.player.CanAfford(selectedOption.CostCredits) {
			m.player.AddCredits(-selectedOption.CostCredits)
			m.encounterModel.message = "The officers accepted your bribe and let you go"
			m.encounterModel.encounter.Resolve()
			m.encounterModel.resolved = true
		} else {
			m.encounterModel.message = "Not enough credits to bribe the officers"
		}

	case "salvage":
		// Salvage derelict
		reward := m.encounterModel.encounter.CreditReward
		if m.encounterModel.encounter.CargoQuantity > 0 {
			m.currentShip.AddCargo("ore", m.encounterModel.encounter.CargoQuantity)
		}

		m.player.AddCredits(reward)
		m.encounterModel.message = fmt.Sprintf("Salvage complete! Found %d credits worth of equipment and %d tons of cargo",
			reward, m.encounterModel.encounter.CargoQuantity)
		m.encounterModel.encounter.Resolve()
		m.encounterModel.resolved = true

		// Check achievements
		m.checkAchievements()

	case "hail":
		// Hail faction patrol
		if m.encounterModel.encounter.FactionID != "" {
			rep := m.player.GetReputation(m.encounterModel.encounter.FactionID)
			if rep >= 50 {
				m.encounterModel.message = "The patrol greets you warmly. They recognize you as an ally."
			} else if rep >= 0 {
				m.encounterModel.message = "The patrol acknowledges your hail and lets you pass."
			} else if rep >= -50 {
				m.encounterModel.message = "The patrol responds coldly but allows you to leave."
			} else {
				m.encounterModel.message = "The patrol is hostile! They're moving to attack!"
				m.encounterModel.encounter.Hostile = true

				// Start combat
				m.combat = newCombatModel()
				m.combat.playerShip = m.currentShip
				m.combat.playerType = models.GetShipTypeByID(m.currentShip.TypeID)
				m.combat.enemyShips = m.encounterModel.generator.GenerateEncounterShips(m.encounterModel.encounter)
				m.combat.enemyTypes = make(map[string]*models.ShipType)
				for _, ship := range m.combat.enemyShips {
					m.combat.enemyTypes[ship.TypeID] = models.GetShipTypeByID(ship.TypeID)
				}

				m.screen = ScreenCombat
				m.encounterModel.encounter.Resolve()
				m.encounterModel.resolved = true
				return m, nil
			}
		}
		m.encounterModel.encounter.Resolve()
		m.encounterModel.resolved = true

	case "ignore":
		// Ignore encounter
		m.encounterModel.message = "You continue on your way"
		m.encounterModel.encounter.Ignore()
		m.encounterModel.resolved = true
	}

	return m, nil
}

func (m Model) viewEncounter() string {
	if m.encounterModel.encounter == nil {
		return "No active encounter\n"
	}

	s := titleStyle.Render("=== ENCOUNTER ===") + "\n\n"

	// Encounter title and description
	titleStr := m.encounterModel.encounter.Title
	if m.encounterModel.encounter.Hostile {
		titleStr = errorStyle.Render(titleStr + " [HOSTILE]")
	} else {
		titleStr = highlightStyle.Render(titleStr)
	}
	s += titleStr + "\n\n"

	s += m.encounterModel.encounter.Description + "\n\n"

	// Show ships involved
	if m.encounterModel.encounter.ShipCount > 0 {
		s += fmt.Sprintf("Ships detected: %d\n", m.encounterModel.encounter.ShipCount)
		for _, shipType := range m.encounterModel.encounter.ShipTypes {
			s += "  - " + shipType + "\n"
		}
		s += "\n"
	}

	// Show faction if relevant
	if m.encounterModel.encounter.FactionID != "" {
		rep := m.player.GetReputation(m.encounterModel.encounter.FactionID)
		repStr := "Neutral"
		if rep >= 50 {
			repStr = successStyle.Render("Allied")
		} else if rep >= 0 {
			repStr = "Friendly"
		} else if rep >= -50 {
			repStr = errorStyle.Render("Unfriendly")
		} else {
			repStr = errorStyle.Render("Hostile")
		}
		s += fmt.Sprintf("Faction: %s (Reputation: %s)\n\n", m.encounterModel.encounter.FactionID, repStr)
	}

	s += strings.Repeat("─", 60) + "\n\n"

	// Show options
	s += subtitleStyle.Render("What do you do?") + "\n\n"

	options := m.encounterModel.encounter.GetOptions(m.player)
	for i, option := range options {
		cursor := "  "
		if i == m.encounterModel.cursor {
			cursor = "> "
		}

		// Check if player can afford
		canAfford := m.encounterModel.encounter.CanAffordOption(option, m.player)
		optionText := option.Label
		if !canAfford {
			optionText = helpStyle.Render(option.Label + " [Cannot afford]")
		} else if option.StartsConflict {
			optionText = errorStyle.Render(option.Label)
		} else if option.GrantsReward {
			optionText = successStyle.Render(option.Label)
		}

		s += cursor + optionText + "\n"
		s += "     " + helpStyle.Render(option.Description) + "\n\n"
	}

	// Show message if any
	if m.encounterModel.message != "" {
		s += "\n" + highlightStyle.Render(m.encounterModel.message) + "\n"
	}

	// Footer
	if m.encounterModel.resolved {
		s += "\n" + renderFooter("Press ESC to continue")
	} else {
		s += "\n" + renderFooter("↑/↓: Select | Enter: Confirm | ESC: Back")
	}

	return s
}
