// File: internal/tui/landing.go
// Project: Terminal Velocity
// Description: Planetary landing screen with services menu
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// pct computes a 0-100 integer percentage with a zero-max guard. Used by the
// Ship Status panel so we never flash "NaN%" / "+Inf%" for a fresh ship whose
// type's max hasn't loaded yet.
func pct(cur, max int) int {
	if max <= 0 {
		return 0
	}
	return (cur * 100) / max
}

// takeoffCmd clears the player's docked-planet state both locally and in the
// DB. Returns a dockedMsg with planet=nil so the top-level handler wipes the
// cache consistently with how it cached on dock.
func (m Model) takeoffCmd() tea.Cmd {
	return func() tea.Msg {
		if m.player == nil || m.playerRepo == nil {
			return dockedMsg{planet: nil}
		}
		ctx := context.Background()
		if err := m.playerRepo.UpdateLocation(ctx, m.player.ID, m.player.CurrentSystem, nil); err != nil {
			return dockedMsg{err: fmt.Errorf("persist takeoff: %w", err)}
		}
		return dockedMsg{planet: nil}
	}
}

type landingModel struct {
	selectedService int
	planetName      string
	government      string
	techLevel       int
	population      string
}

func newLandingModel() landingModel {
	return landingModel{
		selectedService: 0,
		planetName:      "Earth Station",
		government:      "United Earth",
		techLevel:       9,
		population:      "8.2B",
	}
}

func (m Model) viewLanding() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Resolve the docked planet. dockCmd caches m.currentPlanet when the
	// space view's "L" key transitions here; the player may also have
	// landed from a cold start with m.player.CurrentPlanet set but the
	// cache still empty — treat a missing cache as "awaiting data" instead
	// of lying about the location.
	planetName := "Awaiting orbital clearance"
	government := ""
	if m.currentPlanet != nil {
		planetName = m.currentPlanet.Name
	}
	systemName := m.currentLocationLabel()
	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}

	// Header. DrawHeader's second slot was being used for government; keep
	// that but route through the real system name when the government is
	// empty so the top bar is never blank.
	headerRight := government
	if headerRight == "" {
		headerRight = systemName
	}
	header := DrawHeader(planetName, headerRight, credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Main content area with ASCII planet art
	planetArtWidth := 65
	planetArtLeft := (width - planetArtWidth) / 2

	// Planet card — the frame is static decoration, but the three data
	// rows (name, system, population-placeholder) read from real state.
	popLine := ""
	if m.currentPlanet != nil && m.currentPlanet.Population > 0 {
		popLine = fmt.Sprintf("Pop: %s", formatThousands(int64(m.currentPlanet.Population)))
	}
	var planetArt strings.Builder
	planetArt.WriteString("                                                               \n")
	planetArt.WriteString(Center(fmt.Sprintf("Welcome to %s, Commander.", planetName), 63) + "\n")
	planetArt.WriteString("                                                               \n")
	planetArt.WriteString("                       _______________                         \n")
	planetArt.WriteString("                      /               \\                        \n")
	planetArt.WriteString("                     /    " + IconPlanet + "  " + PadRight(truncateCells(strings.ToUpper(planetName), 10), 10) + "\\                       \n")
	planetArt.WriteString("                    |  " + PadRight(truncateCells(systemName+" system", 17), 17) + "|                      \n")
	planetArt.WriteString("                     \\    " + PadRight(popLine, 15) + "/                      \n")
	planetArt.WriteString("                      \\_____    _______/                       \n")
	planetArt.WriteString("                        /   \\__/   \\                           \n")
	planetArt.WriteString("                       /  Station   \\                          \n")
	planetArt.WriteString("                       \\____________/                          \n")
	planetArt.WriteString("                                                               \n")

	// Draw planet art (centered)
	artLines := strings.Split(planetArt.String(), "\n")
	for _, line := range artLines {
		if line == "" {
			continue
		}
		sb.WriteString(BoxVertical)
		sb.WriteString(strings.Repeat(" ", planetArtLeft-1))
		sb.WriteString(line)
		sb.WriteString(strings.Repeat(" ", width-planetArtLeft-len(line)-2))
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Services and Ship Status panels (side by side)
	servicesWidth := 30
	statusWidth := 39
	panelHeight := 12

	// Services panel content
	services := []struct {
		key   string
		label string
		price string
	}{
		{"C", "Commodity Exchange", ""},
		{"O", "Outfitters", ""},
		{"S", "Shipyard", ""},
		{"M", "Mission BBS", ""},
		{"Q", "Quest Terminal", ""},
		{"B", "Bar & News", ""},
		{"R", "Refuel", "(1,200 cr)"},
		{"H", "Repairs", "(Free)"},
	}

	var servicesContent strings.Builder
	servicesContent.WriteString("  AVAILABLE SERVICES:       \n")
	servicesContent.WriteString("                            \n")
	for i, svc := range services {
		prefix := "  "
		if i == m.navigation.cursor {
			prefix = IconArrow + " "
		}
		line := fmt.Sprintf("%s[%s] %-18s %s", prefix, svc.key, svc.label, svc.price)
		servicesContent.WriteString(PadRight(line, servicesWidth-2) + "\n")
	}
	servicesContent.WriteString("                            \n")

	// Ship status panel content — drive from m.currentShip + m.currentSystem
	// instead of the old "Corvette Starhawk / Hull 100% / Sol" demo line.
	var statusContent strings.Builder
	statusContent.WriteString("  SHIP STATUS:                   \n")
	statusContent.WriteString("                                 \n")
	shipLine := "  Ship: (no ship)"
	hullShieldsLine := "  Hull: -   Shields: -"
	fuelCargoLine := "  Fuel: -   Cargo: -"
	if m.currentShip != nil {
		name := m.currentShip.Name
		if name == "" {
			name = "Unnamed"
		}
		shipLine = fmt.Sprintf("  Ship: %s", truncateCells(name, 26))

		maxHull, maxShields, maxFuel, cargoSpace := 100, 100, 100, 0
		if st := models.GetShipTypeByID(m.currentShip.TypeID); st != nil {
			maxHull = st.MaxHull
			maxShields = st.MaxShields
			maxFuel = st.MaxFuel
			cargoSpace = st.CargoSpace
		}
		hullShieldsLine = fmt.Sprintf("  Hull: %d%%  Shields: %d%%",
			pct(m.currentShip.Hull, maxHull), pct(m.currentShip.Shields, maxShields))
		cargoUsed := 0
		for _, item := range m.currentShip.Cargo {
			cargoUsed += item.Quantity
		}
		fuelCargoLine = fmt.Sprintf("  Fuel: %d%%   Cargo: %d/%dt",
			pct(m.currentShip.Fuel, maxFuel), cargoUsed, cargoSpace)
	}
	statusContent.WriteString(PadRight(shipLine, 33) + "\n")
	statusContent.WriteString(PadRight(hullShieldsLine, 33) + "\n")
	statusContent.WriteString(PadRight(fuelCargoLine, 33) + "\n")
	statusContent.WriteString("                                 \n")
	statusContent.WriteString(PadRight(fmt.Sprintf("  Current System: %s", systemName), 33) + "\n")
	if m.currentSystem != nil && m.currentSystem.GovernmentID != "" {
		statusContent.WriteString(PadRight(fmt.Sprintf("  Government: %s", m.currentSystem.GovernmentID), 33) + "\n")
	} else {
		statusContent.WriteString("  Government: -                  \n")
	}
	if m.currentSystem != nil {
		statusContent.WriteString(PadRight(fmt.Sprintf("  Tech Level: %d", m.currentSystem.TechLevel), 33) + "\n")
	} else {
		statusContent.WriteString("  Tech Level: -                  \n")
	}
	statusContent.WriteString("                                 \n")

	// Draw panels (simplified - actual implementation would render side-by-side)
	servicesPanel := DrawPanel("", servicesContent.String(), servicesWidth, panelHeight, false)
	statusPanel := DrawPanel("", statusContent.String(), statusWidth, panelHeight, false)

	// Draw both panels (this is simplified)
	servicesLines := strings.Split(servicesPanel, "\n")
	statusLines := strings.Split(statusPanel, "\n")

	for i := 0; i < len(servicesLines) && i < len(statusLines); i++ {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(servicesLines[i])
		sb.WriteString("  ")
		sb.WriteString(statusLines[i])
		sb.WriteString("    ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// News ticker — empty-state until the news stream is wired to the
	// landing screen. Honest "(quiet)" beats a fake pirate-activity alert.
	newsWidth := width - 8
	newsPanel := DrawPanel("", " NEWS: (no recent dispatches)", newsWidth, 3, false)
	newsLines := strings.Split(newsPanel, "\n")
	for _, line := range newsLines {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(line)
		sb.WriteString("    ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[T]akeoff  [Tab] Next Service  [ESC] Exit", width)
	sb.WriteString(footer)

	return sb.String()
}

// refuelShipCmd refuels the player's ship to full
func (m Model) refuelShipCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if player has a ship
		if m.currentShip == nil {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    0,
				err:     fmt.Errorf("no ship equipped"),
			}
		}

		// Get ship type to determine max fuel
		maxFuel := 300 // Default fallback
		shipType := models.GetShipTypeByID(m.currentShip.TypeID)
		if shipType != nil {
			maxFuel = shipType.MaxFuel
		}
		currentFuel := m.currentShip.Fuel

		// Calculate fuel needed
		fuelNeeded := maxFuel - currentFuel
		if fuelNeeded <= 0 {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    0,
				err:     fmt.Errorf("ship is already fully fueled"),
			}
		}

		// Calculate cost (10 credits per unit of fuel)
		costPerUnit := int64(10)
		totalCost := costPerUnit * int64(fuelNeeded)

		// Check if player has enough credits
		if m.player.Credits < totalCost {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("insufficient credits (need %d, have %d)", totalCost, m.player.Credits),
			}
		}

		// Update ship fuel in database
		err := m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, maxFuel)
		if err != nil {
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("failed to refuel ship: %w", err),
			}
		}

		// Deduct credits from player
		m.player.Credits -= totalCost
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Try to rollback fuel update
			_ = m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, currentFuel)
			return serviceCompleteMsg{
				service: "refuel",
				cost:    totalCost,
				err:     fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Update local ship state
		m.currentShip.Fuel = maxFuel

		return serviceCompleteMsg{
			service: "refuel",
			cost:    totalCost,
			err:     nil,
		}
	}
}

// repairShipCmd repairs the player's ship to full hull and shields
func (m Model) repairShipCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if player has a ship
		if m.currentShip == nil {
			return serviceCompleteMsg{
				service: "repair",
				cost:    0,
				err:     fmt.Errorf("no ship equipped"),
			}
		}

		// Get ship type to determine max hull/shields
		maxHull := 100 // Default fallback
		maxShields := 100 // Default fallback
		shipType := models.GetShipTypeByID(m.currentShip.TypeID)
		if shipType != nil {
			maxHull = shipType.MaxHull
			maxShields = shipType.MaxShields
		}
		currentHull := m.currentShip.Hull
		currentShields := m.currentShip.Shields

		// Calculate damage
		hullDamage := maxHull - currentHull
		shieldDamage := maxShields - currentShields
		totalDamage := hullDamage + shieldDamage

		if totalDamage <= 0 {
			return serviceCompleteMsg{
				service: "repair",
				cost:    0,
				err:     fmt.Errorf("ship is already fully repaired"),
			}
		}

		// Calculate cost (50 credits per point of hull damage, 10 per shield)
		hullCostPerPoint := int64(50)
		shieldCostPerPoint := int64(10)
		totalCost := (hullCostPerPoint * int64(hullDamage)) + (shieldCostPerPoint * int64(shieldDamage))

		// Check if player has enough credits
		if m.player.Credits < totalCost {
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("insufficient credits (need %d, have %d)", totalCost, m.player.Credits),
			}
		}

		// Update ship hull and shields in database
		err := m.shipRepo.UpdateHullAndShields(ctx, m.currentShip.ID, maxHull, maxShields)
		if err != nil {
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("failed to repair ship: %w", err),
			}
		}

		// Deduct credits from player
		m.player.Credits -= totalCost
		err = m.playerRepo.UpdateCredits(ctx, m.playerID, m.player.Credits)
		if err != nil {
			// Try to rollback repair
			_ = m.shipRepo.UpdateHullAndShields(ctx, m.currentShip.ID, currentHull, currentShields)
			return serviceCompleteMsg{
				service: "repair",
				cost:    totalCost,
				err:     fmt.Errorf("failed to deduct credits: %w", err),
			}
		}

		// Update local ship state
		m.currentShip.Hull = maxHull
		m.currentShip.Shields = maxShields

		return serviceCompleteMsg{
			service: "repair",
			cost:    totalCost,
			err:     nil,
		}
	}
}

func (m Model) updateLanding(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.navigation.cursor > 0 {
				m.navigation.cursor--
			}
			return m, nil

		case "down", "j":
			// Max 8 services
			if m.navigation.cursor < 7 {
				m.navigation.cursor++
			}
			return m, nil

		case "c", "C":
			// Commodity Exchange
			m.screen = ScreenTradingEnhanced
			return m, nil

		case "o", "O":
			// Outfitters
			m.screen = ScreenOutfitterEnhanced
			return m, nil

		case "s", "S":
			// Shipyard
			m.screen = ScreenShipyardEnhanced
			return m, nil

		case "m", "M":
			// Missions
			m.screen = ScreenMissionBoardEnhanced
			return m, nil

		case "q", "Q":
			// Quest Terminal
			m.screen = ScreenQuestBoardEnhanced
			return m, nil

		case "b", "B":
			// Bar & News
			m.screen = ScreenNews
			return m, nil

		case "r", "R":
			// Refuel
			return m, m.refuelShipCmd()

		case "h", "H":
			// Repairs
			return m, m.repairShipCmd()

		case "t", "T":
			// Takeoff — leave the planet. Clear the local cache and
			// persist current_planet=NULL so reconnecting users don't see
			// the old landing screen.
			m.screen = ScreenSpaceView
			return m, m.takeoffCmd()

		case "esc":
			// Esc and Takeoff do the same thing — we're no longer on the
			// planet, so unset it in the DB too.
			m.screen = ScreenSpaceView
			return m, m.takeoffCmd()

		case "enter":
			// Select current service
			switch m.navigation.cursor {
			case 0: // Commodity Exchange
				m.screen = ScreenTradingEnhanced
			case 1: // Outfitters
				m.screen = ScreenOutfitterEnhanced
			case 2: // Shipyard
				m.screen = ScreenShipyardEnhanced
			case 3: // Missions
				m.screen = ScreenMissionBoardEnhanced
			case 4: // Quest Terminal
				m.screen = ScreenQuestBoardEnhanced
			case 5: // Bar & News
				m.screen = ScreenNews
			case 6: // Refuel
				return m, m.refuelShipCmd()
			case 7: // Repairs
				return m, m.repairShipCmd()
			}
			return m, nil
		}

	case serviceCompleteMsg:
		// Handle refuel/repair completion
		if msg.err != nil {
			// Show error message
			m.errorMessage = fmt.Sprintf("%s failed: %v", msg.service, msg.err)
			m.showErrorDialog = true
		} else {
			// Show success message
			m.errorMessage = fmt.Sprintf("%s completed! Cost: %d credits",
				strings.Title(msg.service), msg.cost)
			m.showErrorDialog = true
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenLanding and ScreenTradingEnhanced constants to Screen enum when integrating
