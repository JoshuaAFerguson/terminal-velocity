package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/combat"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type combatModel struct {
	// Mode: "tactical", "weapons", "target_select"
	mode string

	// Combat state
	playerShip   *models.Ship
	playerType   *models.ShipType
	enemyShips   []*models.Ship
	enemyTypes   map[string]*models.ShipType
	allyShips    []*models.Ship
	weaponStates []*combat.WeaponState

	// UI state
	selectedTarget int // Index in enemyShips
	selectedWeapon int // Index in playerShip.Weapons
	cursor         int // General cursor position
	radarZoom      int // 1-5, zoom level for radar
	combatLog      []string
	maxLogLines    int

	// Radar/Scanner
	radarSize    int
	radarCenterX int
	radarCenterY int

	// Turn tracking
	turnNumber int
	playerTurn bool

	loading bool
	error   string
}

func newCombatModel() combatModel {
	return combatModel{
		mode:         "tactical",
		cursor:       0,
		radarZoom:    2,
		combatLog:    []string{},
		maxLogLines:  10,
		radarSize:    20,
		radarCenterX: 10,
		radarCenterY: 10,
		turnNumber:   1,
		playerTurn:   true,
		loading:      false,
	}
}

func (m Model) updateCombat(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.combat.mode == "tactical" {
				// Return to main menu
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel current mode
			m.combat.mode = "tactical"
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.combat.cursor > 0 {
				m.combat.cursor--
			}

		case "down", "j":
			// Max cursor depends on mode
			maxCursor := m.getCombatMaxCursor()
			if m.combat.cursor < maxCursor {
				m.combat.cursor++
			}

		case "enter", " ":
			return m.handleCombatAction()

		case "t": // Target selection
			if m.combat.mode == "tactical" && len(m.combat.enemyShips) > 0 {
				m.combat.mode = "target_select"
				m.combat.cursor = m.combat.selectedTarget
			}

		case "w": // Weapons
			if m.combat.mode == "tactical" && m.combat.playerShip != nil {
				m.combat.mode = "weapons"
				m.combat.cursor = m.combat.selectedWeapon
			}

		case "f": // Fire selected weapon
			if m.combat.mode == "tactical" && m.combat.playerTurn {
				return m.executeFireWeapon()
			}

		case "e": // End turn
			if m.combat.playerTurn {
				return m.executeEndTurn()
			}

		case "+", "=": // Zoom in radar
			if m.combat.radarZoom < 5 {
				m.combat.radarZoom++
			}

		case "-", "_": // Zoom out radar
			if m.combat.radarZoom > 1 {
				m.combat.radarZoom--
			}
		}
	}

	return m, nil
}

func (m Model) getCombatMaxCursor() int {
	switch m.combat.mode {
	case "target_select":
		return len(m.combat.enemyShips) - 1
	case "weapons":
		if m.combat.playerShip != nil {
			return len(m.combat.playerShip.Weapons) - 1
		}
	}
	return 0
}

func (m Model) handleCombatAction() (tea.Model, tea.Cmd) {
	switch m.combat.mode {
	case "target_select":
		// Select target
		if m.combat.cursor < len(m.combat.enemyShips) {
			m.combat.selectedTarget = m.combat.cursor
			m.addCombatLog(fmt.Sprintf("Target: %s",
				m.combat.enemyShips[m.combat.selectedTarget].Name))
		}
		m.combat.mode = "tactical"

	case "weapons":
		// Select weapon
		if m.combat.playerShip != nil && m.combat.cursor < len(m.combat.playerShip.Weapons) {
			m.combat.selectedWeapon = m.combat.cursor
			weaponID := m.combat.playerShip.Weapons[m.combat.selectedWeapon]
			weapon := models.GetWeaponByID(weaponID)
			if weapon != nil {
				m.addCombatLog(fmt.Sprintf("Selected: %s", weapon.Name))
			}
		}
		m.combat.mode = "tactical"
	}

	return m, nil
}

func (m Model) executeFireWeapon() (tea.Model, tea.Cmd) {
	if m.combat.playerShip == nil || len(m.combat.enemyShips) == 0 {
		return m, nil
	}

	// Get selected weapon and target
	if m.combat.selectedWeapon >= len(m.combat.playerShip.Weapons) {
		m.addCombatLog("Error: No weapon selected")
		return m, nil
	}

	if m.combat.selectedTarget >= len(m.combat.enemyShips) {
		m.addCombatLog("Error: No target selected")
		return m, nil
	}

	weaponID := m.combat.playerShip.Weapons[m.combat.selectedWeapon]
	weapon := models.GetWeaponByID(weaponID)
	if weapon == nil {
		m.addCombatLog("Error: Invalid weapon")
		return m, nil
	}

	target := m.combat.enemyShips[m.combat.selectedTarget]
	targetType := m.combat.enemyTypes[target.TypeID]

	// Get weapon state
	var weaponState *combat.WeaponState
	for _, ws := range m.combat.weaponStates {
		if ws.WeaponID == weapon.ID {
			weaponState = ws
			break
		}
	}
	if weaponState == nil {
		weaponState = combat.InitializeWeaponState(weapon)
		m.combat.weaponStates = append(m.combat.weaponStates, weaponState)
	}

	// Fire weapon (distance placeholder)
	distance := 500
	result := combat.Fire(weapon, weaponState, m.combat.playerShip, target,
		m.combat.playerType, targetType, distance)

	// Add to combat log
	m.addCombatLog(result.Message)

	// Check if target destroyed
	if target.Hull <= 0 {
		m.addCombatLog(fmt.Sprintf("%s DESTROYED!", target.Name))

		// Record kill for player progression
		if m.player != nil {
			m.player.RecordKill()

			// Check for achievement unlocks
			m.checkAchievements()

			// Show rating update if it changed
			newRating := m.player.CombatRating
			if newRating%10 == 0 { // Show message every 10 points
				m.addCombatLog(fmt.Sprintf("Combat Rating: %d (%s)", newRating, m.player.GetCombatRankTitle()))
			}

			// Show achievement notification if any
			if notification := m.getAchievementNotification(); notification != "" {
				m.addCombatLog(notification)
				m.clearAchievementNotification()
			}
		}

		// Remove from enemy list
		m.combat.enemyShips = append(m.combat.enemyShips[:m.combat.selectedTarget],
			m.combat.enemyShips[m.combat.selectedTarget+1:]...)
		if m.combat.selectedTarget >= len(m.combat.enemyShips) && m.combat.selectedTarget > 0 {
			m.combat.selectedTarget--
		}

		// Check if combat over
		if len(m.combat.enemyShips) == 0 {
			m.addCombatLog("VICTORY! All enemies destroyed!")
			m.combat.playerTurn = false
		}
	}

	return m, nil
}

func (m Model) executeEndTurn() (tea.Model, tea.Cmd) {
	m.combat.playerTurn = false
	m.combat.turnNumber++
	m.addCombatLog(fmt.Sprintf("--- Turn %d ---", m.combat.turnNumber))

	// TODO: Execute enemy AI turns
	// For now, just end turn immediately
	m.addCombatLog("Enemy turn...")

	// Regenerate shields
	if m.combat.playerShip != nil && m.combat.playerType != nil {
		if m.combat.playerShip.Shields < m.combat.playerType.MaxShields {
			regen := m.combat.playerType.ShieldRegen
			m.combat.playerShip.Shields += regen
			if m.combat.playerShip.Shields > m.combat.playerType.MaxShields {
				m.combat.playerShip.Shields = m.combat.playerType.MaxShields
			}
			m.addCombatLog(fmt.Sprintf("Shields recharged +%d", regen))
		}
	}

	// Update weapon cooldowns
	combat.UpdateCooldowns(m.combat.weaponStates, 1.0)

	// Start player's turn again
	m.combat.playerTurn = true

	return m, nil
}

func (m *Model) addCombatLog(message string) {
	m.combat.combatLog = append(m.combat.combatLog, message)
	// Keep only last N lines
	if len(m.combat.combatLog) > m.combat.maxLogLines {
		m.combat.combatLog = m.combat.combatLog[len(m.combat.combatLog)-m.combat.maxLogLines:]
	}
}

func (m Model) viewCombat() string {
	s := renderHeader(m.username, m.player.Credits, "Combat")
	s += "\n"

	// Title
	s += titleStyle.Render(fmt.Sprintf("=== Combat - Turn %d ===", m.combat.turnNumber)) + "\n\n"

	// Error display
	if m.combat.error != "" {
		s += errorStyle.Render(m.combat.error) + "\n\n"
	}

	// Loading state
	if m.combat.loading {
		s += "Loading combat...\n"
		return s
	}

	// Mode-specific view
	switch m.combat.mode {
	case "tactical":
		s += m.viewTacticalDisplay()
	case "target_select":
		s += m.viewTargetSelection()
	case "weapons":
		s += m.viewWeaponSelection()
	default:
		s += "Unknown combat mode\n"
	}

	return s
}

func (m Model) viewTacticalDisplay() string {
	s := ""

	// Player ship status (left side)
	s += m.renderShipStatus(m.combat.playerShip, m.combat.playerType, "YOUR SHIP")
	s += "\n"

	// Target ship status (right side, if selected)
	if m.combat.selectedTarget < len(m.combat.enemyShips) {
		target := m.combat.enemyShips[m.combat.selectedTarget]
		targetType := m.combat.enemyTypes[target.TypeID]
		s += m.renderShipStatus(target, targetType, "TARGET")
		s += "\n"
	}

	// Radar view
	s += m.renderRadar()
	s += "\n"

	// Combat log
	s += m.renderCombatLog()
	s += "\n"

	// Weapon status
	if m.combat.playerShip != nil && len(m.combat.playerShip.Weapons) > 0 {
		s += m.renderWeaponStatus()
		s += "\n"
	}

	// Controls
	helpText := "T: Target  •  W: Weapons  •  F: Fire  •  E: End Turn  •  +/-: Radar Zoom  •  ESC: Main Menu"
	if !m.combat.playerTurn {
		helpText = "Enemy turn in progress..."
	}
	s += renderFooter(helpText)

	return s
}

func (m Model) renderShipStatus(ship *models.Ship, shipType *models.ShipType, label string) string {
	if ship == nil || shipType == nil {
		return ""
	}

	s := subtitleStyle.Render(label+": "+ship.Name) + "\n"

	// Hull bar
	hullPercent := 0
	if shipType.MaxHull > 0 {
		hullPercent = (ship.Hull * 100) / shipType.MaxHull
	}
	hullBar := m.renderStatusBar(ship.Hull, shipType.MaxHull, 20, "█", "░")
	hullColor := statsStyle
	if hullPercent < 30 {
		hullColor = errorStyle
	}
	s += fmt.Sprintf("  Hull:    %s %s\n",
		hullColor.Render(hullBar),
		hullColor.Render(fmt.Sprintf("%d/%d (%d%%)", ship.Hull, shipType.MaxHull, hullPercent)))

	// Shield bar
	shieldPercent := 0
	if shipType.MaxShields > 0 {
		shieldPercent = (ship.Shields * 100) / shipType.MaxShields
	}
	shieldBar := m.renderStatusBar(ship.Shields, shipType.MaxShields, 20, "█", "░")
	s += fmt.Sprintf("  Shields: %s %s\n",
		statsStyle.Render(shieldBar),
		statsStyle.Render(fmt.Sprintf("%d/%d (%d%%)", ship.Shields, shipType.MaxShields, shieldPercent)))

	return s
}

func (m Model) renderStatusBar(current, max, width int, filled, empty string) string {
	if max == 0 {
		return strings.Repeat(empty, width)
	}

	filledWidth := (current * width) / max
	if filledWidth > width {
		filledWidth = width
	}
	if filledWidth < 0 {
		filledWidth = 0
	}

	emptyWidth := width - filledWidth
	return strings.Repeat(filled, filledWidth) + strings.Repeat(empty, emptyWidth)
}

func (m Model) renderRadar() string {
	s := subtitleStyle.Render("Tactical Radar") + " "
	s += helpStyle.Render(fmt.Sprintf("(Zoom: %dx)", m.combat.radarZoom))
	s += "\n"

	// Simple ASCII radar
	size := m.combat.radarSize
	radar := make([][]rune, size)
	for i := range radar {
		radar[i] = make([]rune, size)
		for j := range radar[i] {
			radar[i][j] = '·'
		}
	}

	// Place player at center
	centerX := size / 2
	centerY := size / 2
	radar[centerY][centerX] = 'P'

	// Place enemies (simplified positions)
	for i, enemy := range m.combat.enemyShips {
		if enemy.Hull > 0 {
			// Simplified: arrange enemies in a grid pattern
			// angle := float64(i) * 6.28 / float64(len(m.combat.enemyShips)) // For future circular placement
			distance := 5 // Fixed distance for now
			x := centerX + int(float64(distance)*1.5)
			y := centerY + int(float64(distance)*0.7)

			if x >= 0 && x < size && y >= 0 && y < size {
				if i == m.combat.selectedTarget {
					radar[y][x] = 'T' // Target
				} else {
					radar[y][x] = 'E' // Enemy
				}
			}
		}
	}

	// Render radar
	s += "  " + strings.Repeat("─", size) + "\n"
	for _, row := range radar {
		s += "  " + string(row) + "\n"
	}
	s += "  " + strings.Repeat("─", size) + "\n"
	s += "  P=You  E=Enemy  T=Target\n"

	return s
}

func (m Model) renderCombatLog() string {
	s := subtitleStyle.Render("Combat Log:") + "\n"

	if len(m.combat.combatLog) == 0 {
		s += helpStyle.Render("  No messages yet\n")
	} else {
		for _, msg := range m.combat.combatLog {
			s += "  " + msg + "\n"
		}
	}

	return s
}

func (m Model) renderWeaponStatus() string {
	if m.combat.playerShip == nil {
		return ""
	}

	s := subtitleStyle.Render("Weapons:") + "\n"

	for i, weaponID := range m.combat.playerShip.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon == nil {
			continue
		}

		// Find weapon state
		var state *combat.WeaponState
		for _, ws := range m.combat.weaponStates {
			if ws.WeaponID == weapon.ID {
				state = ws
				break
			}
		}

		prefix := "  "
		if i == m.combat.selectedWeapon {
			prefix = "> "
		}

		status := "Ready"
		if state != nil && state.CooldownRemaining > 0 {
			status = fmt.Sprintf("Cooldown: %.1fs", state.CooldownRemaining)
		}

		ammoInfo := ""
		if weapon.AmmoCapacity > 0 {
			ammo := weapon.AmmoCapacity
			if state != nil {
				ammo = state.CurrentAmmo
			}
			ammoInfo = fmt.Sprintf(" [%d/%d ammo]", ammo, weapon.AmmoCapacity)
		}

		line := fmt.Sprintf("%s - Dmg:%d Range:%s%s - %s",
			weapon.Name, weapon.Damage, weapon.Range, ammoInfo, status)

		if i == m.combat.selectedWeapon {
			s += prefix + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += prefix + line + "\n"
		}
	}

	return s
}

func (m Model) viewTargetSelection() string {
	s := subtitleStyle.Render("=== Select Target ===") + "\n\n"

	if len(m.combat.enemyShips) == 0 {
		s += helpStyle.Render("No enemies remaining\n\n")
		s += renderFooter("ESC: Back")
		return s
	}

	s += "Target                     Hull         Shields\n"
	s += strings.Repeat("─", 60) + "\n"

	for i, enemy := range m.combat.enemyShips {
		if enemy.Hull <= 0 {
			continue
		}

		enemyType := m.combat.enemyTypes[enemy.TypeID]
		if enemyType == nil {
			continue
		}

		hullPercent := (enemy.Hull * 100) / enemyType.MaxHull
		shieldPercent := 0
		if enemyType.MaxShields > 0 {
			shieldPercent = (enemy.Shields * 100) / enemyType.MaxShields
		}

		line := fmt.Sprintf("%-25s %-12s %-12s",
			enemy.Name,
			fmt.Sprintf("%d/%d (%d%%)", enemy.Hull, enemyType.MaxHull, hullPercent),
			fmt.Sprintf("%d/%d (%d%%)", enemy.Shields, enemyType.MaxShields, shieldPercent))

		if i == m.combat.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Confirm  •  ESC: Cancel")

	return s
}

func (m Model) viewWeaponSelection() string {
	s := subtitleStyle.Render("=== Select Weapon ===") + "\n\n"

	if m.combat.playerShip == nil || len(m.combat.playerShip.Weapons) == 0 {
		s += helpStyle.Render("No weapons installed\n\n")
		s += renderFooter("ESC: Back")
		return s
	}

	s += "Weapon                Damage  Range   Type        Status\n"
	s += strings.Repeat("─", 70) + "\n"

	for i, weaponID := range m.combat.playerShip.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon == nil {
			continue
		}

		// Find weapon state
		var state *combat.WeaponState
		for _, ws := range m.combat.weaponStates {
			if ws.WeaponID == weapon.ID {
				state = ws
				break
			}
		}

		status := "Ready"
		if state != nil && state.CooldownRemaining > 0 {
			status = fmt.Sprintf("%.1fs", state.CooldownRemaining)
		}

		ammoInfo := ""
		if weapon.AmmoCapacity > 0 {
			ammo := weapon.AmmoCapacity
			if state != nil {
				ammo = state.CurrentAmmo
			}
			ammoInfo = fmt.Sprintf(" (%d)", ammo)
		}

		line := fmt.Sprintf("%-21s %-7d %-7s %-11s %s%s",
			weapon.Name, weapon.Damage, weapon.Range,
			weapon.Type, status, ammoInfo)

		if i == m.combat.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Confirm  •  ESC: Cancel")

	return s
}
