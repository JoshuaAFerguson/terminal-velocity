// File: internal/tui/navigation_enhanced.go
// Project: Terminal Velocity
// Description: Enhanced navigation screen with visual star map
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type navigationEnhancedModel struct {
	selectedSystem int
	systems        []systemDestination
	currentSystem  string
}

type systemDestination struct {
	name         string
	distance     float64 // light years
	fuelRequired int
	government   string
	techLevel    int
	population   string
	services     []string
	x            int // for map display
	y            int
}

func newNavigationEnhancedModel() navigationEnhancedModel {
	// Sample nearby systems (would come from database in real implementation)
	systems := []systemDestination{
		{
			name: "Alpha Centauri", distance: 3.2, fuelRequired: 32,
			government: "Confederation", techLevel: 8, population: "2.4 billion",
			services: []string{"Shipyard", "Outfitter", "Missions", "Refuel"},
			x:        45, y: 8,
		},
		{
			name: "Proxima Centauri", distance: 4.5, fuelRequired: 45,
			government: "Independent", techLevel: 7, population: "800 million",
			services: []string{"Outfitter", "Missions", "Refuel"},
			x:        50, y: 3,
		},
		{
			name: "Sirius", distance: 6.8, fuelRequired: 68,
			government: "United Earth", techLevel: 9, population: "1.2 billion",
			services: []string{"Shipyard", "Outfitter", "Missions", "Refuel"},
			x:        20, y: 13,
		},
		{
			name: "Barnard's Star", distance: 8.2, fuelRequired: 82,
			government: "Corporate", techLevel: 6, population: "450 million",
			services: []string{"Missions", "Refuel"},
			x:        15, y: 3,
		},
	}

	return navigationEnhancedModel{
		selectedSystem: 0,
		systems:        systems,
		currentSystem:  "Sol",
	}
}

func (m Model) viewNavigationEnhanced() string {
	width := 80
	if m.width > 80 {
		width = m.width
	}

	var sb strings.Builder

	// Header
	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	header := DrawHeader("NAVIGATION MAP", "["+m.navigationEnhanced.currentSystem+" System]", credits, -1, width)
	sb.WriteString(header + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Initialize if needed
	if len(m.navigationEnhanced.systems) == 0 {
		m.navigationEnhanced = newNavigationEnhancedModel()
	}

	// Star map visualization
	mapWidth := width - 8
	mapHeight := 16
	var mapContent strings.Builder

	mapContent.WriteString(Center("NEARBY SYSTEMS", mapWidth-4) + "\n")
	mapContent.WriteString(strings.Repeat(" ", mapWidth-4) + "\n")

	// Create empty map grid
	mapGrid := make([][]rune, mapHeight)
	for i := range mapGrid {
		mapGrid[i] = make([]rune, mapWidth-4)
		for j := range mapGrid[i] {
			mapGrid[i][j] = ' '
		}
	}

	// Place Sol (current system) in center
	solX := (mapWidth - 4) / 2
	solY := mapHeight / 2
	if solY >= 0 && solY < len(mapGrid) && solX >= 0 && solX < len(mapGrid[solY]) {
		mapGrid[solY][solX] = '⊙'
	}

	// Place connected systems and draw connection lines
	for i, sys := range m.navigationEnhanced.systems {
		// Adjust coordinates to fit in grid
		sysX := sys.x
		sysY := sys.y

		if sysY >= 0 && sysY < len(mapGrid) && sysX >= 0 && sysX < len(mapGrid[sysY]) {
			if i == m.navigationEnhanced.selectedSystem {
				mapGrid[sysY][sysX] = '◉' // Selected system
			} else {
				mapGrid[sysY][sysX] = '◉'
			}
		}

		// Draw simple connection line (simplified - just marks path)
		// In a full implementation, would use proper line drawing algorithm
	}

	// Add labels
	// Sol label
	if solY+1 >= 0 && solY+1 < len(mapGrid) {
		label := "SOL ▲"
		startX := solX - len(label)/2
		if startX >= 0 {
			for i, ch := range label {
				if startX+i < len(mapGrid[solY+1]) {
					mapGrid[solY+1][startX+i] = ch
				}
			}
		}
	}

	// System labels
	for i, sys := range m.navigationEnhanced.systems {
		labelY := sys.y + 1
		if labelY >= 0 && labelY < len(mapGrid) {
			label := sys.name
			if i == m.navigationEnhanced.selectedSystem {
				label = label + " *"
			}
			startX := sys.x - len(label)/2
			if startX >= 0 {
				for j, ch := range label {
					if startX+j < len(mapGrid[labelY]) && startX+j >= 0 {
						if mapGrid[labelY][startX+j] == ' ' {
							mapGrid[labelY][startX+j] = ch
						}
					}
				}
			}
		}
	}

	// Render map grid
	for _, row := range mapGrid {
		mapContent.WriteString(string(row) + "\n")
	}

	mapPanel := DrawBoxDouble("", mapContent.String(), mapWidth, mapHeight+2)
	mapLines := strings.Split(mapPanel, "\n")
	for _, line := range mapLines {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(line)
		sb.WriteString("        ")
		sb.WriteString(BoxVertical + "\n")
	}

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Two-column layout: Jump destinations (left) + System details (right)
	leftWidth := 32
	rightWidth := width - leftWidth - 6

	// Jump destinations list
	var destContent strings.Builder
	destContent.WriteString(" JUMP DESTINATIONS:             \n")
	destContent.WriteString("                                \n")

	for i, sys := range m.navigationEnhanced.systems {
		prefix := "   "
		if i == m.navigationEnhanced.selectedSystem {
			prefix = " " + IconArrow + " "
		}
		line := fmt.Sprintf("%s%-18s (%.1f ly)", prefix, sys.name, sys.distance)
		destContent.WriteString(PadRight(line, leftWidth-2) + "\n")
	}

	// Pad to height
	for i := len(m.navigationEnhanced.systems); i < 6; i++ {
		destContent.WriteString(strings.Repeat(" ", leftWidth-2) + "\n")
	}

	destPanel := DrawPanel("", destContent.String(), leftWidth, 10, false)

	// System details
	var detailsContent strings.Builder
	if m.navigationEnhanced.selectedSystem < len(m.navigationEnhanced.systems) {
		sys := m.navigationEnhanced.systems[m.navigationEnhanced.selectedSystem]

		detailsContent.WriteString(fmt.Sprintf(" SELECTED: %-26s\n", sys.name))
		detailsContent.WriteString("                                 \n")
		detailsContent.WriteString(fmt.Sprintf(" Distance: %.1f light years%-9s\n", sys.distance, ""))
		detailsContent.WriteString(fmt.Sprintf(" Fuel Required: %d units%-11s\n", sys.fuelRequired, ""))

		currentFuel := 201 // TODO: Get from ship
		maxFuel := 300
		detailsContent.WriteString(fmt.Sprintf(" Your Fuel: %d/%d units%-9s\n", currentFuel, maxFuel, ""))
		detailsContent.WriteString("                                 \n")
		detailsContent.WriteString(fmt.Sprintf(" Government: %-20s\n", sys.government))
		detailsContent.WriteString(fmt.Sprintf(" Tech Level: %-20d\n", sys.techLevel))
		detailsContent.WriteString(fmt.Sprintf(" Population: %-20s\n", sys.population))
		detailsContent.WriteString("                                 \n")
		detailsContent.WriteString(" Services:                       \n")

		// Format services with checkmarks
		servicesLine := " "
		for i, service := range sys.services {
			if i > 0 && i%2 == 0 {
				detailsContent.WriteString(PadRight(servicesLine, rightWidth-2) + "\n")
				servicesLine = " "
			}
			servicesLine += IconCheck + " " + service + "  "
		}
		if servicesLine != " " {
			detailsContent.WriteString(PadRight(servicesLine, rightWidth-2) + "\n")
		}
		detailsContent.WriteString("                                 \n")
		detailsContent.WriteString(" [ Engage Hyperdrive ]           \n")
	}

	detailsPanel := DrawPanel("", detailsContent.String(), rightWidth, 10, false)

	// Combine panels (simplified rendering)
	sb.WriteString(BoxVertical + "  ")
	sb.WriteString(destPanel)
	sb.WriteString("  ")
	sb.WriteString(BoxVertical + "\n")
	sb.WriteString(BoxVertical + "  ")
	sb.WriteString(detailsPanel)
	sb.WriteString("  ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-2))
	sb.WriteString(BoxVertical + "\n")

	// Footer
	footer := DrawFooter("[↑↓] Select System  [Enter] Jump  [I]nfo  [ESC] Back to Space", width)
	sb.WriteString(footer)

	return sb.String()
}

// executeJumpCmd executes a hyperdrive jump to the selected system
func (m Model) executeJumpCmd(systemName string, fuelRequired int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Validate player has a ship
		if m.currentShip == nil {
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("no ship equipped"),
			}
		}

		// Check current fuel
		if m.currentShip.Fuel < fuelRequired {
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("insufficient fuel (need %d, have %d)", fuelRequired, m.currentShip.Fuel),
			}
		}

		// TODO: Get actual system ID from system name
		// For now, we'll need to look up the system in the database
		// This is a simplified implementation - in production we'd have the system ID already

		// Get connected system IDs
		connectedIDs, err := m.systemRepo.GetConnections(ctx, m.player.CurrentSystem)
		if err != nil {
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("failed to get connected systems: %w", err),
			}
		}

		// Find the target system by name
		var targetSystemID uuid.UUID
		var found bool
		for _, sysID := range connectedIDs {
			sys, err := m.systemRepo.GetSystemByID(ctx, sysID)
			if err != nil {
				continue // Skip systems we can't load
			}
			if sys.Name == systemName {
				targetSystemID = sys.ID
				found = true
				break
			}
		}

		if !found {
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("system %s not found or not connected", systemName),
			}
		}

		// Execute the jump - update player location
		err = m.playerRepo.UpdateLocation(ctx, m.playerID, targetSystemID, nil)
		if err != nil {
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("failed to update location: %w", err),
			}
		}

		// Update ship fuel
		newFuel := m.currentShip.Fuel - fuelRequired
		err = m.shipRepo.UpdateFuel(ctx, m.currentShip.ID, newFuel)
		if err != nil {
			// Try to rollback location update
			_ = m.playerRepo.UpdateLocation(ctx, m.playerID, m.player.CurrentSystem, m.player.CurrentPlanet)
			return jumpExecutedMsg{
				destination: nil,
				fuelUsed:    0,
				err:         fmt.Errorf("failed to update fuel: %w", err),
			}
		}

		// Update local state
		m.player.CurrentSystem = targetSystemID
		m.player.CurrentPlanet = nil // In space after jump
		m.currentShip.Fuel = newFuel

		// Get the destination system for the response
		destSystem, err := m.systemRepo.GetSystemByID(ctx, targetSystemID)
		if err != nil {
			// Jump succeeded but couldn't get system details
			destSystem = nil
		}

		return jumpExecutedMsg{
			destination: destSystem,
			fuelUsed:    fuelRequired,
			err:         nil,
		}
	}
}

func (m Model) updateNavigationEnhanced(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.navigationEnhanced.selectedSystem > 0 {
				m.navigationEnhanced.selectedSystem--
			}
			return m, nil

		case "down", "j":
			if m.navigationEnhanced.selectedSystem < len(m.navigationEnhanced.systems)-1 {
				m.navigationEnhanced.selectedSystem++
			}
			return m, nil

		case "enter":
			// Jump to selected system
			if m.navigationEnhanced.selectedSystem < len(m.navigationEnhanced.systems) {
				system := m.navigationEnhanced.systems[m.navigationEnhanced.selectedSystem]
				return m, m.executeJumpCmd(system.name, system.fuelRequired)
			}
			return m, nil

		case "i", "I":
			// Show detailed system info
			// TODO: Implement detailed info screen
			return m, nil

		case "esc":
			// Back to space view
			m.screen = ScreenSpaceView
			return m, nil
		}

	case jumpExecutedMsg:
		// Handle jump completion
		if msg.err != nil {
			// Show error message
			m.errorMessage = fmt.Sprintf("Jump failed: %v", msg.err)
			m.showErrorDialog = true
		} else {
			// Jump successful - show success message and return to space view
			var destName string
			if msg.destination != nil {
				destName = msg.destination.Name
			} else {
				destName = "Unknown System"
			}
			m.errorMessage = fmt.Sprintf("Jumped to %s. Fuel used: %d units", destName, msg.fuelUsed)
			m.showErrorDialog = true
			// Return to space view after successful jump
			m.screen = ScreenSpaceView
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenNavigationEnhanced constant to Screen enum when integrating
