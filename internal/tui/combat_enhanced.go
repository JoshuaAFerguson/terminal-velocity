// File: internal/tui/combat_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced active combat screen with tactical display and turn-based combat
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/combat"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type combatEnhancedModel struct {
	// Combat state
	isPlayerTurn bool
	combatPhase  string // "ongoing", "victory", "defeat", "fled", "loot"

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

	// Loot system
	lootDrop       *combat.LootDrop
	showingLoot    bool
	enemyShipType  *models.ShipType // Store for loot generation
	enemyWasHostile bool
	enemyHadBounty bool
	enemyBounty    int64
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

	// Action panel - turn selection or loot display
	actionWidth := 68
	var actionContent strings.Builder

	// Show loot screen if available
	if m.combatEnhanced.showingLoot && m.combatEnhanced.lootDrop != nil {
		loot := m.combatEnhanced.lootDrop
		actionContent.WriteString(" SALVAGE AVAILABLE:                                             \n")
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString(fmt.Sprintf("  Credits: %s                                     \n",
			PadRight(formatCredits(loot.Credits), 40)))

		if len(loot.Cargo) > 0 {
			actionContent.WriteString(fmt.Sprintf("  Cargo Items: %d                                              \n", len(loot.Cargo)))
		}
		if len(loot.Weapons) > 0 {
			actionContent.WriteString(fmt.Sprintf("  Weapons: %d (will be sold)                                   \n", len(loot.Weapons)))
		}
		if len(loot.RareItems) > 0 {
			for _, item := range loot.RareItems {
				actionContent.WriteString(fmt.Sprintf("  RARE: %s (%s)                 \n",
					PadRight(item.Name, 30), item.Rarity))
			}
		}
		actionContent.WriteString("                                                                \n")
		actionContent.WriteString("  [C] Collect Salvage        [L] Leave and Continue            \n")
	} else if m.combatEnhanced.isPlayerTurn {
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

	// Footer - dynamic based on combat state
	var footerText string
	if m.combatEnhanced.showingLoot {
		footerText = "[C]ollect Loot  [L]eave Loot  [ESC] Leave"
	} else {
		footerText = "[1-3] Fire Weapon  [E]vade  [D]efend  [R]etreat  [H]ail  [ESC] Menu"
	}
	footer := DrawFooter(footerText, width)
	sb.WriteString(footer)

	return sb.String()
}

// fireWeaponCmd fires a weapon at the enemy
func (m Model) fireWeaponCmd(weaponIndex int) tea.Cmd {
	return func() tea.Msg {
		if m.currentShip == nil {
			return combatActionMsg{
				actionType: "fire",
				weaponSlot: weaponIndex,
				err:        fmt.Errorf("no ship equipped"),
			}
		}

		// Simple hit calculation (would use combat.Fire() in real scenario)
		// For now, we'll simulate combat since we don't have full Weapon models
		weapons := m.combatEnhanced.playerShip.weapons
		if weaponIndex >= len(weapons) {
			return combatActionMsg{
				actionType: "fire",
				weaponSlot: weaponIndex,
				err:        fmt.Errorf("weapon slot not available"),
			}
		}

		weapon := weapons[weaponIndex]

		// Check ammo for missile weapons
		if weapon.maxAmmo > 0 && weapon.ammo <= 0 {
			return combatActionMsg{
				actionType:  "fire",
				weaponSlot:  weaponIndex,
				logMessage:  fmt.Sprintf("No ammo remaining for %s", weapon.name),
				err:         fmt.Errorf("no ammo"),
			}
		}

		// Check energy
		if m.combatEnhanced.playerShip.energy < weapon.energyCost {
			return combatActionMsg{
				actionType:  "fire",
				weaponSlot:  weaponIndex,
				logMessage:  "Insufficient energy to fire weapon",
				err:         fmt.Errorf("insufficient energy"),
			}
		}

		// Calculate hit (simple random for now)
		hit := rand.Float64() < 0.75 // 75% hit chance

		damage := 0
		logMsg := ""
		if hit {
			damage = weapon.damage
			logMsg = fmt.Sprintf("You fire %s - HIT for %d damage!", weapon.name, damage)

			// Apply damage to enemy (shields first, then hull)
			if m.combatEnhanced.enemyShip.shields > 0 {
				if m.combatEnhanced.enemyShip.shields >= damage {
					m.combatEnhanced.enemyShip.shields -= damage
				} else {
					remainingDamage := damage - m.combatEnhanced.enemyShip.shields
					m.combatEnhanced.enemyShip.shields = 0
					m.combatEnhanced.enemyShip.hull -= remainingDamage
					if m.combatEnhanced.enemyShip.hull < 0 {
						m.combatEnhanced.enemyShip.hull = 0
					}
				}
			} else {
				m.combatEnhanced.enemyShip.hull -= damage
				if m.combatEnhanced.enemyShip.hull < 0 {
					m.combatEnhanced.enemyShip.hull = 0
				}
			}
		} else {
			logMsg = fmt.Sprintf("You fire %s - MISS!", weapon.name)
		}

		// Deduct energy
		m.combatEnhanced.playerShip.energy -= weapon.energyCost

		// Deduct ammo if applicable
		if weapon.maxAmmo > 0 {
			m.combatEnhanced.playerShip.weapons[weaponIndex].ammo--
		}

		// Check if enemy is destroyed
		combatOver := m.combatEnhanced.enemyShip.hull <= 0
		victory := combatOver

		return combatActionMsg{
			actionType:  "fire",
			weaponSlot:  weaponIndex,
			hit:         hit,
			damage:      damage,
			logMessage:  logMsg,
			combatOver:  combatOver,
			victory:     victory,
		}
	}
}

// performEvasionCmd performs evasive maneuvers
func (m Model) performEvasionCmd() tea.Cmd {
	return func() tea.Msg {
		// Evasion gives temporary defense bonus
		// For now, just log the action
		logMsg := "You perform evasive maneuvers! (Next enemy attack has reduced accuracy)"

		return combatActionMsg{
			actionType: "evade",
			logMessage: logMsg,
		}
	}
}

// performDefendCmd boosts shields
func (m Model) performDefendCmd() tea.Cmd {
	return func() tea.Msg {
		// Boost shields by 10%
		shieldBoost := 10
		m.combatEnhanced.playerShip.shields += shieldBoost
		if m.combatEnhanced.playerShip.shields > 100 {
			m.combatEnhanced.playerShip.shields = 100
		}

		logMsg := fmt.Sprintf("You boost your shields! (+%d%% shields)", shieldBoost)

		return combatActionMsg{
			actionType: "defend",
			logMessage: logMsg,
		}
	}
}

// generateCombatLootCmd generates loot from destroyed enemy ship
func (m Model) generateCombatLootCmd() tea.Cmd {
	return func() tea.Msg {
		// Need to create a dummy enemy ship from the combatEnhanced model
		// In a real scenario, this would come from the encounter system
		enemyShip := &models.Ship{
			TypeID:  m.combatEnhanced.enemyShip.shipType,
			Hull:    0,
			Shields: 0,
			Cargo:   []models.CargoItem{},
			Weapons: []string{},
			Outfits: []string{},
		}

		// For demo purposes, use a basic ship type
		// In production, this would be loaded from shipRepo
		enemyShipType := &models.ShipType{
			ID:          m.combatEnhanced.enemyShip.shipType,
			Name:        m.combatEnhanced.enemyShip.name,
			Price:       50000, // Default pirate ship value
			Class:       "fighter",
			CargoSpace:  20,
			WeaponSlots: 2,
		}

		// Use stored values or defaults
		wasHostile := m.combatEnhanced.enemyWasHostile
		if !wasHostile {
			wasHostile = m.combatEnhanced.enemyShip.attitude == "hostile"
		}
		hadBounty := m.combatEnhanced.enemyHadBounty
		bountyAmount := m.combatEnhanced.enemyBounty

		// Generate loot
		loot := combat.GenerateLoot(
			enemyShip,
			enemyShipType,
			wasHostile,
			hadBounty,
			bountyAmount,
		)

		return combatLootGeneratedMsg{
			loot:          loot,
			enemyShipType: enemyShipType,
			err:           nil,
		}
	}
}

// collectCombatLootCmd collects the loot and updates player ship/credits
func (m Model) collectCombatLootCmd() tea.Cmd {
	return func() tea.Msg {
		if m.combatEnhanced.lootDrop == nil {
			return combatLootCollectedMsg{
				success: false,
				message: "No loot to collect",
				err:     fmt.Errorf("no loot available"),
			}
		}

		// Get current ship type for cargo space checking
		// In production, this would be loaded from shipRepo
		if m.currentShip == nil {
			return combatLootCollectedMsg{
				success: false,
				message: "No ship equipped",
				err:     fmt.Errorf("no ship"),
			}
		}

		// Create a dummy ship type for demo
		// In production, load from database
		playerShipType := &models.ShipType{
			ID:         m.currentShip.TypeID,
			CargoSpace: 100, // Default cargo space
		}

		// Check cargo space
		if !combat.CanCarryLoot(m.currentShip, playerShipType, m.combatEnhanced.lootDrop) {
			cargoNeeded := combat.CalculateCargoSpaceRequired(m.combatEnhanced.lootDrop)
			cargoUsed := m.currentShip.GetCargoUsed()
			cargoAvailable := playerShipType.CargoSpace - cargoUsed

			return combatLootCollectedMsg{
				success: false,
				message: fmt.Sprintf("Insufficient cargo space! Need %d tons, have %d tons available",
					cargoNeeded, cargoAvailable),
				err:     fmt.Errorf("insufficient cargo space"),
			}
		}

		// Apply loot to player ship
		success, message := combat.ApplyLoot(m.currentShip, playerShipType, m.combatEnhanced.lootDrop)

		if !success {
			return combatLootCollectedMsg{
				success: false,
				message: message,
				err:     fmt.Errorf("failed to apply loot"),
			}
		}

		// Update player credits
		creditsEarned := m.combatEnhanced.lootDrop.Credits

		// Update in database (async)
		if m.player != nil {
			newBalance := m.player.Credits + creditsEarned
			ctx := context.Background()
			err := m.playerRepo.UpdateCredits(ctx, m.playerID, newBalance)
			if err != nil {
				return combatLootCollectedMsg{
					success:       true,
					creditsEarned: creditsEarned,
					message:       message + "\n(Warning: Failed to save credits to database)",
					err:           err,
				}
			}

			// Update local player state
			m.player.Credits = newBalance
		}

		return combatLootCollectedMsg{
			success:       true,
			creditsEarned: creditsEarned,
			message:       message,
			err:           nil,
		}
	}
}

// processAITurnCmd processes the AI enemy's turn
func (m Model) processAITurnCmd() tea.Cmd {
	return func() tea.Msg {
		// Simple AI: randomly choose an action
		action := rand.Intn(3) // 0=fire, 1=evade, 2=defend

		var logMsg string
		hit := false
		damage := 0
		combatOver := false

		switch action {
		case 0: // Fire weapon
			if len(m.combatEnhanced.enemyShip.weapons) > 0 {
				weapon := m.combatEnhanced.enemyShip.weapons[0]

				// Random hit chance
				hit = rand.Float64() < 0.65 // 65% hit chance for enemy

				if hit {
					damage = weapon.damage
					logMsg = fmt.Sprintf("Enemy fires %s - HIT for %d damage!", weapon.name, damage)

					// Apply damage to player (shields first, then hull)
					if m.combatEnhanced.playerShip.shields > 0 {
						if m.combatEnhanced.playerShip.shields >= damage {
							m.combatEnhanced.playerShip.shields -= damage
						} else {
							remainingDamage := damage - m.combatEnhanced.playerShip.shields
							m.combatEnhanced.playerShip.shields = 0
							m.combatEnhanced.playerShip.hull -= remainingDamage
							if m.combatEnhanced.playerShip.hull < 0 {
								m.combatEnhanced.playerShip.hull = 0
							}
						}
					} else {
						m.combatEnhanced.playerShip.hull -= damage
						if m.combatEnhanced.playerShip.hull < 0 {
							m.combatEnhanced.playerShip.hull = 0
						}
					}

					// Check if player is destroyed
					combatOver = m.combatEnhanced.playerShip.hull <= 0
				} else {
					logMsg = fmt.Sprintf("Enemy fires %s - MISS!", weapon.name)
				}
			}

		case 1: // Evade
			logMsg = "Enemy performs evasive maneuvers!"

		case 2: // Defend
			logMsg = "Enemy boosts their shields!"
			m.combatEnhanced.enemyShip.shields += 10
			if m.combatEnhanced.enemyShip.shields > 100 {
				m.combatEnhanced.enemyShip.shields = 100
			}
		}

		return enemyTurnMsg{
			action:     "attack",
			hit:        hit,
			damage:     damage,
			logMessage: logMsg,
			combatOver: combatOver,
		}
	}
}

func (m Model) updateCombatEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if we're showing loot screen
		if m.combatEnhanced.showingLoot {
			switch msg.String() {
			case "c", "C":
				// Collect loot
				return m, m.collectCombatLootCmd()

			case "l", "L":
				// Leave loot and return to space
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> You leave the salvage behind")
				m.combatEnhanced.showingLoot = false
				m.screen = ScreenSpaceView
				return m, nil

			case "esc":
				// ESC also leaves loot
				m.combatEnhanced.showingLoot = false
				m.screen = ScreenSpaceView
				return m, nil
			}
			return m, nil
		}

		// Normal combat controls
		switch msg.String() {
		case "1":
			// Fire weapon slot 0
			if m.combatEnhanced.isPlayerTurn {
				return m, m.fireWeaponCmd(0)
			}
			return m, nil

		case "2":
			// Fire weapon slot 1
			if m.combatEnhanced.isPlayerTurn {
				return m, m.fireWeaponCmd(1)
			}
			return m, nil

		case "3":
			// Fire weapon slot 2
			if m.combatEnhanced.isPlayerTurn {
				return m, m.fireWeaponCmd(2)
			}
			return m, nil

		case "e", "E":
			// Evasive maneuvers
			if m.combatEnhanced.isPlayerTurn {
				return m, m.performEvasionCmd()
			}
			return m, nil

		case "d", "D":
			// Defend (boost shields)
			if m.combatEnhanced.isPlayerTurn {
				return m, m.performDefendCmd()
			}
			return m, nil

		case "r", "R":
			// Retreat to space view
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

	case combatActionMsg:
		// Handle player combat action result
		if msg.err != nil && msg.logMessage != "" {
			// Non-critical error (like no ammo) - show log message
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog, "> "+msg.logMessage)
		} else if msg.err != nil {
			// Critical error
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
		} else {
			// Success - add to combat log
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog, "> "+msg.logMessage)

			// End player turn
			m.combatEnhanced.isPlayerTurn = false
			m.combatEnhanced.turnNumber++

			// Check if combat is over
			if msg.combatOver {
				if msg.victory {
					// Victory - generate loot
					m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
						"> ENEMY DESTROYED! Victory!")
					m.combatEnhanced.combatPhase = "victory"
					// Generate loot from destroyed enemy
					return m, m.generateCombatLootCmd()
				} else {
					// Defeat
					m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
						"> YOUR SHIP IS DESTROYED! Defeat!")
					m.combatEnhanced.combatPhase = "defeat"
					m.screen = ScreenMainMenu
				}
			} else {
				// Continue combat - trigger AI turn
				return m, m.processAITurnCmd()
			}
		}
		return m, nil

	case enemyTurnMsg:
		// Handle AI enemy turn result
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
		} else {
			// Add to combat log
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog, "> "+msg.logMessage)

			// Check if combat is over (player defeated)
			if msg.combatOver {
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> YOUR SHIP IS DESTROYED! Defeat!")
				m.combatEnhanced.combatPhase = "defeat"
				m.screen = ScreenMainMenu
			} else {
				// Return to player turn
				m.combatEnhanced.isPlayerTurn = true
			}
		}
		return m, nil

	case combatLootGeneratedMsg:
		// Handle loot generation result
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
			m.screen = ScreenSpaceView // Return to space on error
		} else {
			// Store loot and show loot screen
			loot, ok := msg.loot.(*combat.LootDrop)
			if ok && loot != nil {
				m.combatEnhanced.lootDrop = loot
				m.combatEnhanced.showingLoot = true
				m.combatEnhanced.combatPhase = "loot"

				// Add loot message to combat log
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> Salvage available! Press [C] to collect or [L] to leave")
			} else {
				// No loot or error converting
				m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
					"> No salvage recovered from wreckage")
				m.screen = ScreenSpaceView
			}
		}
		return m, nil

	case combatLootCollectedMsg:
		// Handle loot collection result
		if msg.err != nil && !msg.success {
			// Failed to collect loot
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
				"> "+msg.message)
			// Stay on loot screen to try again or leave
		} else {
			// Successfully collected loot
			m.combatEnhanced.combatLog = append(m.combatEnhanced.combatLog,
				"> Loot collected! Earned "+formatCredits(msg.creditsEarned)+" credits")
			m.combatEnhanced.showingLoot = false
			m.screen = ScreenSpaceView
		}
		return m, nil
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
