package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type outfitterModel struct {
	cursor          int
	tab             string // "weapons", "outfits", "installed"
	mode            string // "browse", "confirm_install", "confirm_remove"
	selectedWeapon  *models.Weapon
	selectedOutfit  *models.Outfit
	selectedInstall string // ID of item to remove
	availableItems  interface{}
	loading         bool
	error           string
}

type outfitterLoadedMsg struct {
	weapons []models.Weapon
	outfits []models.Outfit
	err     error
}

type equipmentChangedMsg struct {
	success bool
	err     error
}

func newOutfitterModel() outfitterModel {
	return outfitterModel{
		cursor:  0,
		tab:     "weapons",
		mode:    "browse",
		loading: true,
	}
}

func (m Model) updateOutfitter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.outfitter.mode == "browse" {
				m.screen = ScreenMainMenu
				return m, nil
			}
			// Cancel operation
			m.outfitter.mode = "browse"
			m.outfitter.error = ""
			return m, nil

		case "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "tab":
			// Switch tabs
			if m.outfitter.mode == "browse" {
				switch m.outfitter.tab {
				case "weapons":
					m.outfitter.tab = "outfits"
				case "outfits":
					m.outfitter.tab = "installed"
				case "installed":
					m.outfitter.tab = "weapons"
				}
				m.outfitter.cursor = 0
			}

		case "up", "k":
			if m.outfitter.cursor > 0 {
				m.outfitter.cursor--
			}

		case "down", "j":
			maxCursor := m.getMaxCursor()
			if m.outfitter.cursor < maxCursor {
				m.outfitter.cursor++
			}

		case "enter", " ":
			if m.outfitter.mode == "browse" {
				if m.outfitter.tab == "installed" {
					// Remove installed item
					return m, m.confirmRemoveEquipment()
				} else {
					// Install new item
					return m, m.confirmInstallEquipment()
				}
			} else if m.outfitter.mode == "confirm_install" {
				return m, m.executeInstall()
			} else if m.outfitter.mode == "confirm_remove" {
				return m, m.executeRemove()
			}
		}

	case outfitterLoadedMsg:
		m.outfitter.loading = false
		if msg.err != nil {
			m.outfitter.error = fmt.Sprintf("Failed to load outfitter: %v", msg.err)
		} else {
			m.outfitter.error = ""
		}

	case equipmentChangedMsg:
		if msg.success {
			m.outfitter.mode = "browse"
			m.outfitter.error = "Equipment updated successfully!"
			m.outfitter.selectedWeapon = nil
			m.outfitter.selectedOutfit = nil
			m.outfitter.selectedInstall = ""
			// Reload player/ship data
			return m, m.loadPlayer()
		} else {
			m.outfitter.error = fmt.Sprintf("Failed: %v", msg.err)
			m.outfitter.mode = "browse"
		}
	}

	return m, nil
}

func (m Model) viewOutfitter() string {
	locationName := "Space"
	s := renderHeader(m.username, m.player.Credits, locationName)
	s += "\n"

	s += subtitleStyle.Render("=== Outfitter ===") + "\n\n"

	if m.outfitter.error != "" {
		s += helpStyle.Render(m.outfitter.error) + "\n\n"
	}

	if m.outfitter.loading {
		s += "Loading outfitter data...\n"
		return s
	}

	if m.currentShip == nil {
		s += errorStyle.Render("No ship available") + "\n"
		return s
	}

	// Ship info and stats
	shipType := models.GetShipTypeByID(m.currentShip.TypeID)
	if shipType == nil {
		s += errorStyle.Render("Unknown ship type") + "\n"
		return s
	}

	s += m.viewShipStats(shipType)
	s += "\n"

	// Tab indicator
	tabs := []string{"Weapons", "Outfits", "Installed"}
	tabDisplay := ""
	for i, tab := range tabs {
		tabName := strings.ToLower(tab)
		if tabName == m.outfitter.tab {
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

	// Mode-specific view
	switch m.outfitter.mode {
	case "browse":
		s += m.viewOutfitterBrowse()
	case "confirm_install":
		s += m.viewInstallConfirmation()
	case "confirm_remove":
		s += m.viewRemoveConfirmation()
	}

	return s
}

func (m Model) viewShipStats(shipType *models.ShipType) string {
	s := fmt.Sprintf("Ship: %s (%s)\n", statsStyle.Render(m.currentShip.Name), shipType.Name)

	// Calculate bonuses from outfits
	shieldBonus, hullBonus, cargoBonus, fuelBonus, speedBonus := models.CalculateShipBonuses(m.currentShip.Outfits)

	// Weapon slots
	usedWeaponSlots := len(m.currentShip.Weapons)
	s += fmt.Sprintf("Weapon Slots: %s / %d\n",
		statsStyle.Render(fmt.Sprintf("%d", usedWeaponSlots)),
		shipType.WeaponSlots)

	// Outfit space
	usedOutfitSpace := m.calculateUsedOutfitSpace()
	totalOutfitSpace := shipType.OutfitSpace
	s += fmt.Sprintf("Outfit Space: %s / %d\n",
		statsStyle.Render(fmt.Sprintf("%d", usedOutfitSpace)),
		totalOutfitSpace)

	// Show bonuses if any
	if shieldBonus > 0 || hullBonus > 0 || cargoBonus > 0 || fuelBonus > 0 || speedBonus > 0 {
		s += "\nActive Bonuses: "
		bonuses := []string{}
		if shieldBonus > 0 {
			bonuses = append(bonuses, fmt.Sprintf("+%d shields", shieldBonus))
		}
		if hullBonus > 0 {
			bonuses = append(bonuses, fmt.Sprintf("+%d hull", hullBonus))
		}
		if cargoBonus > 0 {
			bonuses = append(bonuses, fmt.Sprintf("+%d cargo", cargoBonus))
		}
		if fuelBonus > 0 {
			bonuses = append(bonuses, fmt.Sprintf("+%d fuel", fuelBonus))
		}
		if speedBonus > 0 {
			bonuses = append(bonuses, fmt.Sprintf("+%d speed", speedBonus))
		}
		s += statsStyle.Render(strings.Join(bonuses, ", "))
	}

	return s
}

func (m Model) viewOutfitterBrowse() string {
	s := ""

	if m.outfitter.tab == "weapons" {
		s += m.viewWeaponsList()
	} else if m.outfitter.tab == "outfits" {
		s += m.viewOutfitsList()
	} else if m.outfitter.tab == "installed" {
		s += m.viewInstalledList()
	}

	return s
}

func (m Model) viewWeaponsList() string {
	s := ""
	weapons := models.StandardWeapons

	// Sort by price
	sort.Slice(weapons, func(i, j int) bool {
		return weapons[i].Price < weapons[j].Price
	})

	if len(weapons) == 0 {
		s += helpStyle.Render("No weapons available") + "\n\n"
		s += renderFooter("Tab: Switch  •  ESC: Main Menu")
		return s
	}

	// Table header
	s += "Weapon                    Type      Damage  Range    Space   Price\n"
	s += strings.Repeat("─", 78) + "\n"

	for i, weapon := range weapons {
		affordable := weapon.Price <= m.player.Credits
		canFit := m.canFitWeapon(&weapon)

		line := fmt.Sprintf("%-25s %-9s %-7d %-8s %-7d %d cr",
			weapon.Name,
			weapon.Type,
			weapon.Damage,
			weapon.Range,
			weapon.OutfitSpace,
			weapon.Price,
		)

		if !affordable || !canFit {
			line = helpStyle.Render(line)
		}

		if i == m.outfitter.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Install  •  Tab: Switch  •  ESC: Menu")
	return s
}

func (m Model) viewOutfitsList() string {
	s := ""
	outfits := models.StandardOutfits

	// Sort by type then price
	sort.Slice(outfits, func(i, j int) bool {
		if outfits[i].Type == outfits[j].Type {
			return outfits[i].Price < outfits[j].Price
		}
		return outfits[i].Type < outfits[j].Type
	})

	if len(outfits) == 0 {
		s += helpStyle.Render("No outfits available") + "\n\n"
		s += renderFooter("Tab: Switch  •  ESC: Main Menu")
		return s
	}

	// Table header
	s += "Outfit                    Type             Space   Price      Effect\n"
	s += strings.Repeat("─", 78) + "\n"

	for i, outfit := range outfits {
		affordable := outfit.Price <= m.player.Credits
		canFit := m.canFitOutfit(&outfit)

		effect := m.getOutfitEffectString(&outfit)
		line := fmt.Sprintf("%-25s %-16s %-7d %-10d %s",
			outfit.Name,
			outfit.Type,
			outfit.OutfitSpace,
			outfit.Price,
			effect,
		)

		if !affordable || !canFit {
			line = helpStyle.Render(line)
		}

		if i == m.outfitter.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Install  •  Tab: Switch  •  ESC: Menu")
	return s
}

func (m Model) viewInstalledList() string {
	s := ""

	if len(m.currentShip.Weapons) == 0 && len(m.currentShip.Outfits) == 0 {
		s += helpStyle.Render("No equipment installed") + "\n\n"
		s += renderFooter("Tab: Switch  •  ESC: Main Menu")
		return s
	}

	// List installed weapons
	if len(m.currentShip.Weapons) > 0 {
		s += "Installed Weapons:\n"
		for i, weaponID := range m.currentShip.Weapons {
			weapon := models.GetWeaponByID(weaponID)
			if weapon == nil {
				continue
			}

			line := fmt.Sprintf("  %s (Dmg: %d, Space: %d)",
				weapon.Name, weapon.Damage, weapon.OutfitSpace)

			if i == m.outfitter.cursor && m.outfitter.cursor < len(m.currentShip.Weapons) {
				s += "> " + selectedMenuItemStyle.Render(line) + "\n"
			} else {
				s += "  " + line + "\n"
			}
		}
		s += "\n"
	}

	// List installed outfits
	if len(m.currentShip.Outfits) > 0 {
		s += "Installed Outfits:\n"
		outfitStartIdx := len(m.currentShip.Weapons)
		for i, outfitID := range m.currentShip.Outfits {
			outfit := models.GetOutfitByID(outfitID)
			if outfit == nil {
				continue
			}

			effect := m.getOutfitEffectString(outfit)
			line := fmt.Sprintf("  %s (%s, Space: %d)",
				outfit.Name, effect, outfit.OutfitSpace)

			cursorIdx := outfitStartIdx + i
			if m.outfitter.cursor == cursorIdx {
				s += "> " + selectedMenuItemStyle.Render(line) + "\n"
			} else {
				s += "  " + line + "\n"
			}
		}
	}

	s += "\n" + renderFooter("↑/↓: Select  •  Enter: Remove  •  Tab: Switch  •  ESC: Menu")
	return s
}

func (m Model) viewInstallConfirmation() string {
	s := ""

	if m.outfitter.selectedWeapon != nil {
		weapon := m.outfitter.selectedWeapon
		s += errorStyle.Render("=== Install Weapon ===") + "\n\n"
		s += fmt.Sprintf("Weapon: %s\n", weapon.Name)
		s += fmt.Sprintf("Damage: %d  Range: %s  Accuracy: %d%%\n", weapon.Damage, weapon.Range, weapon.Accuracy)
		s += fmt.Sprintf("Outfit Space: %d\n", weapon.OutfitSpace)
		s += fmt.Sprintf("Price: %s cr\n\n", statsStyle.Render(fmt.Sprintf("%d", weapon.Price)))

		if weapon.Price > m.player.Credits {
			s += errorStyle.Render("⚠ Insufficient credits!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else if !m.canFitWeapon(weapon) {
			s += errorStyle.Render("⚠ No weapon slots or outfit space available!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else {
			s += helpStyle.Render("Enter: Confirm Installation  •  ESC: Cancel")
		}
	} else if m.outfitter.selectedOutfit != nil {
		outfit := m.outfitter.selectedOutfit
		s += errorStyle.Render("=== Install Outfit ===") + "\n\n"
		s += fmt.Sprintf("Outfit: %s\n", outfit.Name)
		s += fmt.Sprintf("Description: %s\n", outfit.Description)
		s += fmt.Sprintf("Effect: %s\n", m.getOutfitEffectString(outfit))
		s += fmt.Sprintf("Outfit Space: %d\n", outfit.OutfitSpace)
		s += fmt.Sprintf("Price: %s cr\n\n", statsStyle.Render(fmt.Sprintf("%d", outfit.Price)))

		if outfit.Price > m.player.Credits {
			s += errorStyle.Render("⚠ Insufficient credits!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else if !m.canFitOutfit(outfit) {
			s += errorStyle.Render("⚠ Insufficient outfit space!\n\n")
			s += helpStyle.Render("ESC: Cancel")
		} else {
			s += helpStyle.Render("Enter: Confirm Installation  •  ESC: Cancel")
		}
	}

	return s
}

func (m Model) viewRemoveConfirmation() string {
	s := errorStyle.Render("=== Remove Equipment ===") + "\n\n"

	if m.outfitter.cursor < len(m.currentShip.Weapons) {
		weaponID := m.currentShip.Weapons[m.outfitter.cursor]
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			s += fmt.Sprintf("Remove: %s\n", weapon.Name)
			s += fmt.Sprintf("Refund: %s cr (50%% of value)\n\n",
				statsStyle.Render(fmt.Sprintf("%d", weapon.Price/2)))
		}
	} else {
		outfitIdx := m.outfitter.cursor - len(m.currentShip.Weapons)
		if outfitIdx < len(m.currentShip.Outfits) {
			outfitID := m.currentShip.Outfits[outfitIdx]
			outfit := models.GetOutfitByID(outfitID)
			if outfit != nil {
				s += fmt.Sprintf("Remove: %s\n", outfit.Name)
				s += fmt.Sprintf("Refund: %s cr (50%% of value)\n\n",
					statsStyle.Render(fmt.Sprintf("%d", outfit.Price/2)))
			}
		}
	}

	s += helpStyle.Render("Enter: Confirm Removal  •  ESC: Cancel")
	return s
}

// Helper functions

func (m Model) getMaxCursor() int {
	if m.outfitter.tab == "weapons" {
		return len(models.StandardWeapons) - 1
	} else if m.outfitter.tab == "outfits" {
		return len(models.StandardOutfits) - 1
	} else if m.outfitter.tab == "installed" {
		return len(m.currentShip.Weapons) + len(m.currentShip.Outfits) - 1
	}
	return 0
}

func (m Model) calculateUsedOutfitSpace() int {
	total := 0
	for _, weaponID := range m.currentShip.Weapons {
		weapon := models.GetWeaponByID(weaponID)
		if weapon != nil {
			total += weapon.OutfitSpace
		}
	}
	for _, outfitID := range m.currentShip.Outfits {
		outfit := models.GetOutfitByID(outfitID)
		if outfit != nil {
			total += outfit.OutfitSpace
		}
	}
	return total
}

func (m Model) canFitWeapon(weapon *models.Weapon) bool {
	if m.currentShip == nil {
		return false
	}
	shipType := models.GetShipTypeByID(m.currentShip.TypeID)
	if shipType == nil {
		return false
	}

	// Check weapon slots
	if len(m.currentShip.Weapons) >= shipType.WeaponSlots {
		return false
	}

	// Check outfit space
	usedSpace := m.calculateUsedOutfitSpace()
	return (usedSpace + weapon.OutfitSpace) <= shipType.OutfitSpace
}

func (m Model) canFitOutfit(outfit *models.Outfit) bool {
	if m.currentShip == nil {
		return false
	}
	shipType := models.GetShipTypeByID(m.currentShip.TypeID)
	if shipType == nil {
		return false
	}

	// Check outfit space
	usedSpace := m.calculateUsedOutfitSpace()
	return (usedSpace + outfit.OutfitSpace) <= shipType.OutfitSpace
}

func (m Model) getOutfitEffectString(outfit *models.Outfit) string {
	effects := []string{}
	if outfit.ShieldBonus > 0 {
		effects = append(effects, fmt.Sprintf("+%d shields", outfit.ShieldBonus))
	}
	if outfit.HullBonus > 0 {
		effects = append(effects, fmt.Sprintf("+%d hull", outfit.HullBonus))
	}
	if outfit.CargoBonus > 0 {
		effects = append(effects, fmt.Sprintf("+%d cargo", outfit.CargoBonus))
	}
	if outfit.FuelBonus > 0 {
		effects = append(effects, fmt.Sprintf("+%d fuel", outfit.FuelBonus))
	}
	if outfit.SpeedBonus > 0 {
		effects = append(effects, fmt.Sprintf("+%d speed", outfit.SpeedBonus))
	}
	if len(effects) == 0 {
		return "No effect"
	}
	return strings.Join(effects, ", ")
}

// Commands

func (m Model) confirmInstallEquipment() tea.Cmd {
	return func() tea.Msg {
		if m.outfitter.tab == "weapons" {
			weapons := models.StandardWeapons
			if m.outfitter.cursor < len(weapons) {
				m.outfitter.selectedWeapon = &weapons[m.outfitter.cursor]
				m.outfitter.mode = "confirm_install"
			}
		} else if m.outfitter.tab == "outfits" {
			outfits := models.StandardOutfits
			if m.outfitter.cursor < len(outfits) {
				m.outfitter.selectedOutfit = &outfits[m.outfitter.cursor]
				m.outfitter.mode = "confirm_install"
			}
		}
		return nil
	}
}

func (m Model) confirmRemoveEquipment() tea.Cmd {
	return func() tea.Msg {
		m.outfitter.mode = "confirm_remove"
		return nil
	}
}

func (m Model) executeInstall() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Install weapon
		if m.outfitter.selectedWeapon != nil {
			weapon := m.outfitter.selectedWeapon

			// Validate
			if weapon.Price > m.player.Credits {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("insufficient credits")}
			}
			if !m.canFitWeapon(weapon) {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("no space available")}
			}

			// Deduct credits
			newCredits := m.player.Credits - weapon.Price
			err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
			if err != nil {
				return equipmentChangedMsg{success: false, err: err}
			}

			// Add weapon to ship (database operation)
			m.currentShip.Weapons = append(m.currentShip.Weapons, weapon.ID)
			err = m.shipRepo.Update(ctx, m.currentShip)
			if err != nil {
				// Rollback credits
				m.playerRepo.UpdateCredits(ctx, m.player.ID, m.player.Credits)
				return equipmentChangedMsg{success: false, err: err}
			}

			return equipmentChangedMsg{success: true}
		}

		// Install outfit
		if m.outfitter.selectedOutfit != nil {
			outfit := m.outfitter.selectedOutfit

			// Validate
			if outfit.Price > m.player.Credits {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("insufficient credits")}
			}
			if !m.canFitOutfit(outfit) {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("no space available")}
			}

			// Deduct credits
			newCredits := m.player.Credits - outfit.Price
			err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
			if err != nil {
				return equipmentChangedMsg{success: false, err: err}
			}

			// Add outfit to ship
			m.currentShip.Outfits = append(m.currentShip.Outfits, outfit.ID)
			err = m.shipRepo.Update(ctx, m.currentShip)
			if err != nil {
				// Rollback credits
				m.playerRepo.UpdateCredits(ctx, m.player.ID, m.player.Credits)
				return equipmentChangedMsg{success: false, err: err}
			}

			return equipmentChangedMsg{success: true}
		}

		return equipmentChangedMsg{success: false, err: fmt.Errorf("nothing selected")}
	}
}

func (m Model) executeRemove() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Remove weapon
		if m.outfitter.cursor < len(m.currentShip.Weapons) {
			weaponID := m.currentShip.Weapons[m.outfitter.cursor]
			weapon := models.GetWeaponByID(weaponID)
			if weapon == nil {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("weapon not found")}
			}

			// Refund 50%
			refund := weapon.Price / 2
			newCredits := m.player.Credits + refund
			err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
			if err != nil {
				return equipmentChangedMsg{success: false, err: err}
			}

			// Remove from ship
			m.currentShip.Weapons = append(m.currentShip.Weapons[:m.outfitter.cursor],
				m.currentShip.Weapons[m.outfitter.cursor+1:]...)
			err = m.shipRepo.Update(ctx, m.currentShip)
			if err != nil {
				// Rollback credits
				m.playerRepo.UpdateCredits(ctx, m.player.ID, m.player.Credits)
				return equipmentChangedMsg{success: false, err: err}
			}

			return equipmentChangedMsg{success: true}
		}

		// Remove outfit
		outfitIdx := m.outfitter.cursor - len(m.currentShip.Weapons)
		if outfitIdx >= 0 && outfitIdx < len(m.currentShip.Outfits) {
			outfitID := m.currentShip.Outfits[outfitIdx]
			outfit := models.GetOutfitByID(outfitID)
			if outfit == nil {
				return equipmentChangedMsg{success: false, err: fmt.Errorf("outfit not found")}
			}

			// Refund 50%
			refund := outfit.Price / 2
			newCredits := m.player.Credits + refund
			err := m.playerRepo.UpdateCredits(ctx, m.player.ID, newCredits)
			if err != nil {
				return equipmentChangedMsg{success: false, err: err}
			}

			// Remove from ship
			m.currentShip.Outfits = append(m.currentShip.Outfits[:outfitIdx],
				m.currentShip.Outfits[outfitIdx+1:]...)
			err = m.shipRepo.Update(ctx, m.currentShip)
			if err != nil {
				// Rollback credits
				m.playerRepo.UpdateCredits(ctx, m.player.ID, m.player.Credits)
				return equipmentChangedMsg{success: false, err: err}
			}

			return equipmentChangedMsg{success: true}
		}

		return equipmentChangedMsg{success: false, err: fmt.Errorf("invalid selection")}
	}
}

// loadOutfitter loads available equipment
func (m Model) loadOutfitter() tea.Cmd {
	return func() tea.Msg {
		// For now, just return success
		// In future, filter by tech level/location
		return outfitterLoadedMsg{
			weapons: models.StandardWeapons,
			outfits: models.StandardOutfits,
			err:     nil,
		}
	}
}
