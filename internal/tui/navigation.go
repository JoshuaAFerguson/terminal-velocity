package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
)

type navigationModel struct {
	cursor          int
	currentSystem   *models.StarSystem
	connectedSystems []*models.StarSystem
	loading         bool
	error           string
}

type systemsLoadedMsg struct {
	current   *models.StarSystem
	connected []*models.StarSystem
	err       error
}

type jumpCompleteMsg struct {
	success bool
	system  *models.StarSystem
	err     error
}

func newNavigationModel() navigationModel {
	return navigationModel{
		cursor:  0,
		loading: true,
	}
}

func (m Model) updateNavigation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.screen = ScreenMainMenu
			return m, nil

		case "up", "k":
			if m.navigation.cursor > 0 {
				m.navigation.cursor--
			}

		case "down", "j":
			if m.navigation.cursor < len(m.navigation.connectedSystems)-1 {
				m.navigation.cursor++
			}

		case "enter", " ":
			// Initiate jump to selected system
			if m.navigation.cursor < len(m.navigation.connectedSystems) {
				targetSystem := m.navigation.connectedSystems[m.navigation.cursor]
				return m, m.initiateJump(targetSystem)
			}
		}

	case systemsLoadedMsg:
		m.navigation.loading = false
		if msg.err != nil {
			m.navigation.error = fmt.Sprintf("Failed to load systems: %v", msg.err)
		} else {
			m.navigation.currentSystem = msg.current
			m.navigation.connectedSystems = msg.connected
			m.navigation.error = ""
		}

	case jumpCompleteMsg:
		if msg.success {
			// Update local state
			m.player.CurrentSystem = msg.system.ID
			m.navigation.currentSystem = msg.system

			// Update ship fuel in local model
			if m.currentShip != nil {
				jumpCost := calculateJumpCost(m.navigation.currentSystem, msg.system)
				m.currentShip.Fuel -= jumpCost
				if m.currentShip.Fuel < 0 {
					m.currentShip.Fuel = 0
				}
			}

			// Reload systems for new location
			return m, m.loadConnectedSystems()
		} else {
			m.navigation.error = fmt.Sprintf("Jump failed: %v", msg.err)
		}
	}

	return m, nil
}

func (m Model) viewNavigation() string {
	// Header with player stats
	s := renderHeader(m.username, m.player.Credits, m.navigation.currentSystem.Name)
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Navigation ===") + "\n\n"

	// Error display
	if m.navigation.error != "" {
		s += errorStyle.Render("⚠ " + m.navigation.error) + "\n\n"
	}

	// Loading state
	if m.navigation.loading {
		s += "Loading systems...\n"
		return s
	}

	// Current system info
	if m.navigation.currentSystem != nil {
		sys := m.navigation.currentSystem
		info := fmt.Sprintf("Current System: %s\n", statsStyle.Render(sys.Name))
		info += fmt.Sprintf("Tech Level: %d  •  Government: %s\n", sys.TechLevel, sys.GovernmentID)
		if len(sys.Planets) > 0 {
			planetNames := make([]string, len(sys.Planets))
			for i, p := range sys.Planets {
				planetNames[i] = p.Name
			}
			info += fmt.Sprintf("Planets: %s\n", strings.Join(planetNames, ", "))
		}
		s += boxStyle.Render(info) + "\n\n"
	}

	// Ship status (fuel)
	if m.currentShip != nil {
		fuelInfo := fmt.Sprintf("Fuel: %s / %d",
			statsStyle.Render(fmt.Sprintf("%d", m.currentShip.Fuel)),
			100) // TODO: Get max fuel from ship type when we have ship types loaded
		s += fuelInfo + "\n\n"
	} else {
		s += helpStyle.Render("No ship available\n\n")
	}

	// Connected systems list
	s += "Available Jump Routes:\n\n"
	if len(m.navigation.connectedSystems) == 0 {
		s += "  No jump routes available from this system.\n"
	} else {
		for i, sys := range m.navigation.connectedSystems {
			jumpCost := calculateJumpCost(m.navigation.currentSystem, sys)
			canAfford := m.currentShip != nil && m.currentShip.Fuel >= jumpCost

			line := fmt.Sprintf("%-20s  Tech: %d  Fuel: %d",
				sys.Name,
				sys.TechLevel,
				jumpCost)

			if !canAfford {
				line += " (Insufficient fuel)"
				line = helpStyle.Render(line)
			}

			if i == m.navigation.cursor {
				s += "> " + selectedMenuItemStyle.Render(line) + "\n"
			} else {
				s += "  " + menuItemStyle.Render(line) + "\n"
			}
		}
	}

	// Help text
	s += renderFooter("↑/↓: Select  •  Enter: Jump  •  ESC: Back to Main Menu")

	return s
}

// loadConnectedSystems loads the current system and all connected systems
func (m Model) loadConnectedSystems() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Load current system
		currentSystem, err := m.systemRepo.GetSystemByID(ctx, m.player.CurrentSystem)
		if err != nil {
			return systemsLoadedMsg{err: err}
		}

		// Load all connected systems
		var connectedSystems []*models.StarSystem
		for _, connectedID := range currentSystem.ConnectedSystems {
			system, err := m.systemRepo.GetSystemByID(ctx, connectedID)
			if err != nil {
				// Skip systems that fail to load, but log the error
				continue
			}
			connectedSystems = append(connectedSystems, system)
		}

		return systemsLoadedMsg{
			current:   currentSystem,
			connected: connectedSystems,
		}
	}
}

// initiateJump starts a jump to a target system
func (m Model) initiateJump(targetSystem *models.StarSystem) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if we have a ship
		if m.currentShip == nil {
			return jumpCompleteMsg{
				success: false,
				err:     fmt.Errorf("no ship available"),
			}
		}

		// Calculate fuel cost
		jumpCost := calculateJumpCost(m.navigation.currentSystem, targetSystem)

		// Check if we have enough fuel
		if m.currentShip.Fuel < jumpCost {
			return jumpCompleteMsg{
				success: false,
				err:     fmt.Errorf("insufficient fuel (need %d, have %d)", jumpCost, m.currentShip.Fuel),
			}
		}

		// Consume fuel
		newFuel := m.currentShip.Fuel - jumpCost
		err := m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, newFuel)
		if err != nil {
			return jumpCompleteMsg{
				success: false,
				err:     fmt.Errorf("failed to update fuel: %w", err),
			}
		}

		// Update player location
		err = m.playerRepo.UpdateLocation(ctx, m.player.ID, targetSystem.ID, nil)
		if err != nil {
			return jumpCompleteMsg{
				success: false,
				err:     err,
			}
		}

		return jumpCompleteMsg{
			success: true,
			system:  targetSystem,
		}
	}
}

// calculateJumpCost calculates the fuel cost for a jump
func calculateJumpCost(from, to *models.StarSystem) int {
	// For now, use a simple distance-based calculation
	// Each unit of distance costs 1 fuel, minimum 5 fuel
	if from == nil || to == nil {
		return 10 // Default cost
	}

	distance := from.Position.DistanceTo(to.Position)
	// Distance is squared, so we take a simplified approach
	// Each 100 units of squared distance = 1 fuel, minimum 5
	cost := int(distance / 100)
	if cost < 5 {
		cost = 5
	}
	if cost > 50 {
		cost = 50 // Cap at 50 fuel
	}
	return cost
}
