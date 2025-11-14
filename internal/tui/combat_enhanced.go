// File: internal/tui/combat_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced active combat screen with tactical display and turn-based combat
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type combatEnhancedModel struct {
	// Combat state
	isPlayerTurn bool
	combatPhase  string // "ongoing", "victory", "defeat", "fled"

	// Ships
	playerShip combatShip
	enemyShip  combatShip

	// Combat tracking
	distance     int     // kilometers
	closingSpeed int     // km/s (negative = moving away)
	turnNumber   int
	combatLog    []string

	// Player action selection
	selectedAction int
	actionMode     string // "select", "confirm"
}

type combatShip struct {
	name     string
	shipType string
	hull     int // percentage
	maxHull  int
	shields  int // percentage
	maxShields int
	energy   int // percentage
	weapons  []combatWeapon
	attitude string // "hostile", "neutral", "friendly"
}

type combatWeapon struct {
	name       string
	ready      bool
	ammo       int // -1 for unlimited
	maxAmmo    int
	energyCost int
	damage     int
}

func newCombatEnhancedModel() combatEnhancedModel {
	// Sample combat scenario
	playerWeapons := []combatWeapon{
		{name: "Laser Cannon", ready: true, ammo: -1, maxAmmo: -1, energyCost: 10, damage: 45},
		{name: "Pulse Laser", ready: true, ammo: -1, maxAmmo: -1, energyCost: 15, damage: 35},
		{name: "Missiles", ready: true, ammo: 15, maxAmmo: 15, energyCost: 5, damage: 80},
	}

	enemyWeapons := []combatWeapon{
		{name: "Pulse Laser", ready: true, ammo: -1, maxAmmo: -1, energyCost: 15, damage: 30},
		{name: "Light Cannon", ready: true, ammo: -1, maxAmmo: -1, energyCost: 12, damage: 40},
	}

	return combatEnhancedModel{
		isPlayerTurn: true,
		combatPhase:  "ongoing",
		playerShip: combatShip{
			name:     "Your Ship",
			shipType: "Corvette",
			hull:     100,
			maxHull:  100,
			shields:  60,
			maxShields: 100,
			energy:   80,
			weapons:  playerWeapons,
			attitude: "neutral",
		},
		enemyShip: combatShip{
			name:     "Pirate Viper",
			shipType: "Viper",
			hull:     40,
			maxHull:  100,
			shields:  25,
			maxShields: 100,
			energy:   60,
			weapons:  enemyWeapons,
			attitude: "hostile",
		},
		distance:     1850,
		closingSpeed: 120,
		turnNumber:   1,
		combatLog: []string{
			"> Pirate Viper is hailing you: \"Prepare to die!\"",
			"> You fire Laser Cannon - HIT for 45 damage!",
			"> Pirate fires Pulse Laser - MISS!",
			"> Your shields absorb 30 damage from Pulse Laser",
		},
		selectedAction: 0,
		actionMode:     "select",
	}
}

func (m Model) viewCombatEnhanced() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Header with shields
	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	shieldPercent := m.combatEnhanced.playerShip.shields

	header := DrawHeader("COMBAT ENGAGED!", "[Sol System]", credits, shieldPercent, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Initialize if needed
	if len(m.combatEnhanced.combatLog) == 0 {
		m.combatEnhanced = newCombatEnhancedModel()
	}

	// Main combat area - tactical display
	tacticalWidth := 65
	tacticalHeight := 16

	var tacticalContent strings.Builder
	tacticalContent.WriteString(Center("TACTICAL DISPLAY", tacticalWidth-4) + "\n")
	tacticalContent.WriteString("\n")
	tacticalContent.WriteString("\n")

	// Enemy ship at top
	enemyLine := Center("◆", tacticalWidth-4)
	tacticalContent.WriteString(enemyLine + "\n")
	tacticalContent.WriteString(Center(m.combatEnhanced.enemyShip.name, tacticalWidth-4) + "\n")
	tacticalContent.WriteString(Center("[LOCKED]", tacticalWidth-4) + "\n")
	tacticalContent.WriteString(Center("↓", tacticalWidth-4) + "\n")
	tacticalContent.WriteString(Center("~~~~ WEAPONS ~~~~", tacticalWidth-4) + "\n")
	tacticalContent.WriteString(Center("↓", tacticalWidth-4) + "\n")
	tacticalContent.WriteString("\n")

	// Player ship at bottom
	tacticalContent.WriteString(Center("△", tacticalWidth-4) + "\n")
	tacticalContent.WriteString(Center("Your Ship", tacticalWidth-4) + "\n")
	tacticalContent.WriteString("\n")

	// Distance and speed info
	distanceLine := fmt.Sprintf("Distance: %s km     Closing at %d km/s",
		FormatNumber(m.combatEnhanced.distance), m.combatEnhanced.closingSpeed)
	tacticalContent.WriteString(Center(distanceLine, tacticalWidth-4) + "\n")
	tacticalContent.WriteString("\n")

	tacticalPanel := DrawBoxDouble("", tacticalContent.String(), tacticalWidth, tacticalHeight)

	// Right sidebar - ship status
	sidebarWidth := 15
	var shipStatusContent strings.Builder
	shipStatusContent.WriteString(" YOUR SHIP   \n")
	shipStatusContent.WriteString("━━━━━━━━━━━━━\n")
	shipStatusContent.WriteString(fmt.Sprintf(" %s    \n", m.combatEnhanced.playerShip.shipType))
	shipStatusContent.WriteString("             \n")
	shipStatusContent.WriteString(fmt.Sprintf(" Hull: %s\n", DrawProgressBar(m.combatEnhanced.playerShip.hull, 100, 6)))
	shipStatusContent.WriteString(fmt.Sprintf("       %d%%  \n", m.combatEnhanced.playerShip.hull))
	shipStatusContent.WriteString("             \n")
	shipStatusContent.WriteString(" Shields:    \n")
	shipStatusContent.WriteString(fmt.Sprintf(" %s  \n", DrawProgressBar(m.combatEnhanced.playerShip.shields, 100, 10)))
	shipStatusContent.WriteString(fmt.Sprintf("       %d%%   \n", m.combatEnhanced.playerShip.shields))
	shipStatusContent.WriteString("             \n")
	shipStatusContent.WriteString(" Energy:     \n")
	shipStatusContent.WriteString(fmt.Sprintf(" %s  \n", DrawProgressBar(m.combatEnhanced.playerShip.energy, 100, 10)))
	shipStatusContent.WriteString(fmt.Sprintf("       %d%%   \n", m.combatEnhanced.playerShip.energy))

	shipStatus := DrawPanel("", shipStatusContent.String(), sidebarWidth, tacticalHeight, false)

	// Render tactical display + ship status side by side (simplified rendering)
	tacticalLines := strings.Split(tacticalPanel, "\n")
	statusLines := strings.Split(shipStatus, "\n")

	for i := 0; i < len(tacticalLines); i++ {
		sb.WriteString(BoxVertical + "    ")
		if i < len(tacticalLines) {
			sb.WriteString(tacticalLines[i])
		} else {
			sb.WriteString(strings.Repeat(" ", tacticalWidth))
		}
		sb.WriteString("  ")
		if i < len(statusLines) {
			sb.WriteString(statusLines[i])
		}
		sb.WriteString("\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Enemy status panel
	enemyWidth := 68
	var enemyContent strings.Builder
	enemy := m.combatEnhanced.enemyShip
	enemyContent.WriteString(fmt.Sprintf(" ENEMY: %-56s\n", enemy.name))
	hullBar := DrawProgressBar(enemy.hull, 100, 8)
	shieldBar := DrawProgressBar(enemy.shields, 100, 8)
	enemyContent.WriteString(fmt.Sprintf(" Hull: %s %d%%   Shields: %s %d%%   Weapons: Active \n",
		hullBar, enemy.hull, shieldBar, enemy.shields))

	enemyPanel := DrawPanel("", enemyContent.String(), enemyWidth, 3, false)
	enemyLines := strings.Split(enemyPanel, "\n")

	// Weapons panel (right sidebar)
	var weaponsContent strings.Builder
	weaponsContent.WriteString(" WEAPONS     \n")
	weaponsContent.WriteString("━━━━━━━━━━━━━\n")
	for i, weapon := range m.combatEnhanced.playerShip.weapons {
		weaponsContent.WriteString(fmt.Sprintf(" %d. %-7s\n", i+1, weapon.name))
		if weapon.maxAmmo > 0 {
			weaponsContent.WriteString(fmt.Sprintf("    [%d/%d]  \n", weapon.ammo, weapon.maxAmmo))
		} else {
			if weapon.ready {
				weaponsContent.WriteString("    [READY]  \n")
			} else {
				weaponsContent.WriteString("    [RELOAD] \n")
			}
		}
		weaponsContent.WriteString("             \n")
	}

	weaponsPanel := DrawPanel("", weaponsContent.String(), sidebarWidth, 6, false)
	weaponsLines := strings.Split(weaponsPanel, "\n")

	// Render enemy panel + weapons panel
	for i := 0; i < len(enemyLines) || i < len(weaponsLines); i++ {
		sb.WriteString(BoxVertical + "  ")
		if i < len(enemyLines) {
			sb.WriteString(enemyLines[i])
		} else {
			sb.WriteString(strings.Repeat(" ", enemyWidth))
		}
		sb.WriteString("  ")
		if i < len(weaponsLines) {
			sb.WriteString(weaponsLines[i])
		}
		sb.WriteString("\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Combat log panel
	logWidth := 68
	var logContent strings.Builder
	logContent.WriteString(" COMBAT LOG:                                                    \n")
	// Show last 4 log entries
	startIdx := 0
	if len(m.combatEnhanced.combatLog) > 4 {
		startIdx = len(m.combatEnhanced.combatLog) - 4
	}
	for i := startIdx; i < len(m.combatEnhanced.combatLog); i++ {
		logContent.WriteString(PadRight(" "+m.combatEnhanced.combatLog[i], logWidth-2) + "\n")
	}
	// Pad to 4 lines
	for i := len(m.combatEnhanced.combatLog); i < 4; i++ {
		logContent.WriteString(strings.Repeat(" ", logWidth-2) + "\n")
	}

	logPanel := DrawPanel("", logContent.String(), logWidth, 6, false)
	logLines := strings.Split(logPanel, "\n")
	for _, line := range logLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Action panel - turn selection
	actionWidth := 68
	var actionContent strings.Builder
	if m.combatEnhanced.isPlayerTurn {
		actionContent.WriteString(" YOUR TURN - Select Action:                                     \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("  [1] Fire Laser Cannon     [2] Fire Pulse Laser               \n")
		actionContent.WriteString("  [3] Fire Missile          [E] Evasive Maneuvers              \n")
		actionContent.WriteString("  [D] Defend (Boost Shields) [R] Retreat (Flee Combat)         \n")
		actionContent.WriteString("                                                                \n")
	} else {
		actionContent.WriteString(" ENEMY TURN - Enemy is taking action...                         \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("                                                                \n")
	}

	actionPanel := DrawPanel("", actionContent.String(), actionWidth, 7, false)
	actionLines := strings.Split(actionPanel, "\n")
	for _, line := range actionLines {
		sb.WriteString(BoxVertical + "  ")
		sb.WriteString(line)
		sb.WriteString("  ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[1-3] Fire Weapon  [E]vade  [D]efend  [R]etreat  [H]ail  [ESC] Menu", width)
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) updateCombatEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			// Fire Laser Cannon
			if m.combatEnhanced.isPlayerTurn && len(m.combatEnhanced.playerShip.weapons) > 0 {
				// TODO: Implement weapon firing logic via API
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> You fire "+m.combatEnhanced.playerShip.weapons[0].name+" - HIT!")
				m.combatEnhanced.isPlayerTurn = false
			}
			return m, nil

		case "2":
			// Fire Pulse Laser
			if m.combatEnhanced.isPlayerTurn && len(m.combatEnhanced.playerShip.weapons) > 1 {
				// TODO: Implement weapon firing logic via API
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> You fire "+m.combatEnhanced.playerShip.weapons[1].name+" - HIT!")
				m.combatEnhanced.isPlayerTurn = false
			}
			return m, nil

		case "3":
			// Fire Missile
			if m.combatEnhanced.isPlayerTurn && len(m.combatEnhanced.playerShip.weapons) > 2 {
				// TODO: Implement weapon firing logic via API
				weapon := m.combatEnhanced.playerShip.weapons[2]
				if weapon.ammo > 0 {
					m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
						"> You fire "+weapon.name+" - HIT!")
					m.combatEnhanced.playerShip.weapons[2].ammo--
					m.combatEnhanced.isPlayerTurn = false
				} else {
					m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
						"> No missiles remaining!")
				}
			}
			return m, nil

		case "e", "E":
			// Evasive maneuvers
			if m.combatEnhanced.isPlayerTurn {
				// TODO: Implement evasion logic via API
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> You perform evasive maneuvers!")
				m.combatEnhanced.isPlayerTurn = false
			}
			return m, nil

		case "d", "D":
			// Defend (boost shields)
			if m.combatEnhanced.isPlayerTurn {
				// TODO: Implement shield boost logic via API
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> You boost your shields!")
				m.combatEnhanced.isPlayerTurn = false
			}
			return m, nil

		case "r", "R":
			// Retreat
			// TODO: Implement retreat logic via API
			// Check if retreat is successful, return to space view
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
				"> You attempt to flee combat...")
			m.screen = ScreenSpaceView
			return m, nil

		case "h", "H":
			// Hail enemy
			// TODO: Implement communication system
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
				"> You hail the enemy ship...")
			return m, nil

		case "esc":
			// Pause menu / abort combat
			m.screen = ScreenMainMenu
			return m, nil
		}
	}

	return m, nil
}

// Helper to format numbers with commas
func FormatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	// Add commas
	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// Add ScreenCombatEnhanced constant to Screen enum when integrating
