// File: internal/tui/encounter.go
// Project: Terminal Velocity
// Description: Encounter screen - Random encounter resolution interface
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The encounter screen handles random space encounters:
// - Pirates: Hostile ships demanding cargo or credits
// - Traders: Peaceful merchants offering trades
// - Police: Law enforcement scanning for contraband
// - Distress Calls: Ships in need of rescue
// - Derelicts: Abandoned ships with salvage opportunities
// - Faction Patrols: Military ships with reputation-based reactions
//
// Encounter Types and Options:
// - Pirates: Engage (combat), Flee (escape chance), Bribe (if affordable)
// - Traders: Trade (buy goods), Hail (friendly chat), Ignore
// - Police: Cooperate (scan), Flee, Bribe (if criminal/illegal cargo)
// - Distress: Rescue (rewards), Ignore
// - Derelicts: Salvage (loot), Ignore
// - Patrols: Hail (reputation check), Engage (if hostile), Flee
//
// Encounter Resolution:
// - Player chooses from contextual options
// - Outcomes based on player choices, reputation, and random chance
// - Combat: Transitions to combat screen with generated enemy ships
// - Trade: Exchange credits for cargo
// - Flee: Calculated based on ship speed and enemy ships
// - Bribe: Pay credits to avoid conflict
// - Scan: Police check for illegal cargo and criminal status
// - Salvage: Collect credits and cargo from derelicts
//
// Dynamic Outcomes:
// - Reputation affects faction patrol reactions
// - Criminal status triggers police hostility
// - Failed flee attempts lead to combat
// - Successful rescues award credits and reputation
// - Achievement checks for certain encounter resolutions

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/encounters"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// encounterModel contains the state for the encounter screen.
// Manages random encounter display and player choice resolution.
type encounterModel struct {
	encounter *models.Encounter      // Current active encounter
	generator *encounters.Generator  // Encounter generator for ship creation
	cursor    int                    // Current cursor position in options list
	message   string                 // Status or outcome message to display
	resolved  bool                   // True if encounter has been resolved
}

// newEncounterModel creates and initializes a new encounter screen model.
// Initializes the encounter generator for creating enemy ships.
func newEncounterModel() encounterModel {
	return encounterModel{
		cursor:    0,
		generator: encounters.NewGenerator(),
		resolved:  false,
	}
}

// updateEncounter handles input and state updates for the encounter screen.
//
// Key Bindings:
//   - esc: Return to main menu (only after encounter resolved)
//   - up/k: Move cursor up in options list
//   - down/j: Move cursor down in options list
//   - enter/space: Select and execute current option
//
// Encounter Workflow:
//   1. Encounter triggers (random or scripted)
//   2. Encounter screen displays with title, description, ships
//   3. Player views available options based on encounter type
//   4. Player selects option with cursor and enter
//   5. Option validation (credits, cargo space, requirements)
//   6. Execute option effect (combat, trade, rewards, etc.)
//   7. Encounter marked as resolved
//   8. Return to navigation or combat screen
//
// Option Effects:
//   - engage/attack: Start combat with generated enemy ships
//   - flee: Calculate escape chance, start combat if failed
//   - trade: Deduct credits, add cargo, resolve encounter
//   - rescue: Award credits and reputation, resolve encounter
//   - cooperate: Police scan (hostility if criminal)
//   - bribe: Pay credits to avoid conflict, resolve encounter
//   - salvage: Collect rewards from derelict, resolve encounter
//   - hail: Check reputation, hostile if low (<-50)
//   - ignore: Resolve encounter without interaction
//
// Message Handling:
//   - All updates happen synchronously
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

// viewEncounter renders the encounter screen.
//
// Layout:
//   - Title: "=== ENCOUNTER ==="
//   - Encounter Title: Name with [HOSTILE] tag if applicable
//   - Description: Encounter narrative text
//   - Ships Detected: Count and types of encountered ships
//   - Faction Info: Faction name and reputation status (if applicable)
//   - Separator line
//   - Options List: Available player choices
//   - Option Details: Description for each option
//   - Status Message: Outcome or validation message
//   - Footer: Key bindings help
//
// Encounter Title Display:
//   - Hostile encounters: Red title with [HOSTILE] tag
//   - Peaceful encounters: Highlighted title (yellow/blue)
//   - Title reflects encounter type (Pirates, Traders, Police, etc.)
//
// Ships Display:
//   - Ship count shown
//   - Ship types listed (e.g., "Pirate Corvette", "Police Interceptor")
//   - Multiple ship types displayed as list
//
// Faction Display (if applicable):
//   - Faction name shown
//   - Reputation status: Allied (green), Friendly, Neutral, Unfriendly (red), Hostile (red)
//   - Color-coded based on reputation value:
//     * 50+: Allied (green)
//     * 0-49: Friendly
//     * -1 to -49: Unfriendly (red)
//     * -50 or lower: Hostile (red)
//
// Options Display:
//   - Option label on first line (e.g., "Attack", "Flee", "Trade")
//   - Option description indented below
//   - Unaffordable options shown dimmed with "[Cannot afford]" tag
//   - Hostile options (StartsConflict) shown in red
//   - Reward options (GrantsReward) shown in green
//   - Selected option highlighted with cursor (>)
//
// Visual Features:
//   - Hostile indicators in red (error style)
//   - Reward indicators in green (success style)
//   - Status messages highlighted
//   - Resolved encounters show "Press ESC to continue"
//   - Affordability checks prevent invalid selections
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
