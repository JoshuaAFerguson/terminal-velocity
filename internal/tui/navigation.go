// File: internal/tui/navigation.go
// Project: Terminal Velocity
// Description: Terminal UI component for navigation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/encounters"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

type navigationModel struct {
	cursor           int
	currentSystem    *models.StarSystem
	connectedSystems []*models.StarSystem
	loading          bool
	error            string
	jumping          bool
	jumpTarget       *models.StarSystem
	jumpProgress     int
	jumpTotal        int
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

type jumpInitiatedMsg struct {
	targetSystem *models.StarSystem
	travelTime   int // seconds
}

type jumpProgressMsg struct {
	elapsed int
	total   int
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
			// Don't allow jumping while already jumping
			if m.navigation.jumping {
				return m, nil
			}

			// Initiate jump to selected system
			if m.navigation.cursor < len(m.navigation.connectedSystems) {
				targetSystem := m.navigation.connectedSystems[m.navigation.cursor]

				// Validate jump before starting sequence
				if m.currentShip == nil {
					m.navigation.error = "No ship available"
					return m, nil
				}

				jumpCost := calculateJumpCost(m.navigation.currentSystem, targetSystem)
				if m.currentShip.Fuel < jumpCost {
					m.navigation.error = fmt.Sprintf("Insufficient fuel (need %d, have %d)", jumpCost, m.currentShip.Fuel)
					return m, nil
				}

				// Start jump sequence
				m.navigation.jumping = true
				m.navigation.jumpTarget = targetSystem
				m.navigation.error = ""

				// Calculate travel time (1-5 seconds based on distance)
				travelTime := calculateTravelTime(m.navigation.currentSystem, targetSystem)
				m.navigation.jumpProgress = 0
				m.navigation.jumpTotal = travelTime

				return m, tea.Batch(
					m.tickJumpProgress(),
					m.executeJump(targetSystem),
				)
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

	case jumpProgressMsg:
		if m.navigation.jumping {
			m.navigation.jumpProgress = msg.elapsed
			if msg.elapsed < msg.total {
				return m, m.tickJumpProgress()
			}
		}

	case jumpCompleteMsg:
		m.navigation.jumping = false
		m.navigation.jumpProgress = 0
		m.navigation.jumpTotal = 0

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

			// Record jump for exploration tracking
			if m.player != nil {
				m.player.RecordJump()
				m.checkAchievements()
			}

			// Check for random encounter
			generator := encounters.NewGenerator()
			dangerLevel := 5 // Default danger level, could be from system data
			if msg.system != nil {
				// Use system's actual danger level if available
				// For now using default 5
			}

			if generator.ShouldGenerateEncounter(dangerLevel, m.player) {
				// Generate encounter
				encounter := generator.GenerateEncounter(msg.system.ID, dangerLevel, m.player)
				m.encounterModel.encounter = encounter
				m.encounterModel.resolved = false
				m.encounterModel.message = ""
				m.encounterModel.cursor = 0

				// Switch to encounter screen
				m.screen = ScreenEncounter
				return m, nil
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
	systemName := "Unknown"
	if m.navigation.currentSystem != nil {
		systemName = m.navigation.currentSystem.Name
	}
	s := renderHeader(m.username, m.player.Credits, systemName)
	s += "\n"

	// Title
	s += subtitleStyle.Render("=== Navigation ===") + "\n\n"

	// Jump sequence in progress
	if m.navigation.jumping {
		return s + m.renderJumpSequence()
	}

	// Error display
	if m.navigation.error != "" {
		s += errorStyle.Render("⚠ "+m.navigation.error) + "\n\n"
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

// renderJumpSequence renders the jump animation
func (m Model) renderJumpSequence() string {
	if m.navigation.jumpTarget == nil {
		return "Jumping...\n"
	}

	progress := float64(m.navigation.jumpProgress) / float64(m.navigation.jumpTotal)
	barWidth := 40
	filled := int(progress * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	s := fmt.Sprintf("Jumping to %s...\n\n", statsStyle.Render(m.navigation.jumpTarget.Name))
	s += fmt.Sprintf("%s %d%%\n\n", bar, int(progress*100))
	s += "Engaging hyperdrive...\n"

	if progress > 0.3 {
		s += "Entering hyperspace corridor...\n"
	}
	if progress > 0.6 {
		s += "Approaching destination system...\n"
	}
	if progress > 0.9 {
		s += "Preparing to exit hyperspace...\n"
	}

	return s
}

// tickJumpProgress creates a ticker for jump progress animation
func (m Model) tickJumpProgress() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return jumpProgressMsg{
			elapsed: m.navigation.jumpProgress + 1,
			total:   m.navigation.jumpTotal,
		}
	})
}

// executeJump starts a jump to a target system (actual game logic)
func (m Model) executeJump(targetSystem *models.StarSystem) tea.Cmd {
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

// calculateTravelTime calculates the travel time for a jump (in ticks of 100ms)
func calculateTravelTime(from, to *models.StarSystem) int {
	// Travel time is 1-5 seconds (10-50 ticks)
	// Based on distance
	if from == nil || to == nil {
		return 30 // 3 seconds default
	}

	distance := from.Position.DistanceTo(to.Position)
	// Longer distances take more time
	ticks := 10 + int(distance/200)
	if ticks < 10 {
		ticks = 10 // Minimum 1 second
	}
	if ticks > 50 {
		ticks = 50 // Maximum 5 seconds
	}
	return ticks
}
