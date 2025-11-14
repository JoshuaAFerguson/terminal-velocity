// File: internal/tui/space_view.go
// Project: Terminal Velocity
// Description: Main space view with 2D viewport, HUD, radar, status, and real-time interactions
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type spaceViewModel struct {
	// Space objects visible in current system
	planets  []*models.Planet
	ships    []spaceObject
	player   playerPosition

	// Target selection
	targetIndex int
	hasTarget   bool

	// Chat state
	chatExpanded bool
	chatInput    string
	chatChannel  int // 0: Global, 1: System, 2: Faction, 3: DM
}

type spaceObject struct {
	name     string
	icon     string
	x, y     float64
	distance float64
	hostile  bool
	objType  string // "planet", "ship", "enemy", "player"
}

type playerPosition struct {
	x, y float64
}

func newSpaceViewModel() spaceViewModel {
	return spaceViewModel{
		player:       playerPosition{x: 0, y: 0},
		chatExpanded: false,
		chatChannel:  0,
	}
}

// convertShipsToSpaceObjects converts Ship models to spaceObject for display
func convertShipsToSpaceObjects(ships []*models.Ship, playerPos playerPosition) []spaceObject {
	objects := make([]spaceObject, 0, len(ships))

	for i, ship := range ships {
		// Position ships in a circle around the player for demo
		// In production, use actual ship coordinates
		angle := float64(i) * (360.0 / float64(len(ships))) * (3.14159 / 180.0)
		distance := 50.0 + float64(i)*10.0

		x := playerPos.x + distance*math.Cos(angle)
		y := playerPos.y + distance*math.Sin(angle)

		// Determine if ship is hostile (simplified)
		hostile := false // Would check ship faction/reputation

		objects = append(objects, spaceObject{
			name:     ship.Name,
			icon:     "◊", // Ship icon
			x:        x,
			y:        y,
			distance: distance,
			hostile:  hostile,
			objType:  "player", // All nearby ships are players in this context
		})
	}

	return objects
}

// convertPlanetsToSpaceObjects converts Planet models to spaceObject for display
func convertPlanetsToSpaceObjects(planets []*models.Planet, playerPos playerPosition) []spaceObject {
	objects := make([]spaceObject, 0, len(planets))

	for i, planet := range planets {
		// Position planets in a different pattern from ships
		angle := float64(i) * (360.0 / float64(len(planets))) * (3.14159 / 180.0)
		distance := 150.0 + float64(i)*20.0

		x := playerPos.x + distance*math.Cos(angle)
		y := playerPos.y + distance*math.Sin(angle)

		objects = append(objects, spaceObject{
			name:     planet.Name,
			icon:     "●", // Planet icon
			x:        x,
			y:        y,
			distance: distance,
			hostile:  false,
			objType:  "planet",
		})
	}

	return objects
}

// Command functions for async space view operations

// loadSpaceViewDataCmd loads current system data, planets, and nearby ships
func (m Model) loadSpaceViewDataCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get current system from player's location
		var system *models.StarSystem
		var planets []*models.Planet
		var nearbyShips []*models.Ship
		var err error

		if m.player != nil && m.systemRepo != nil {
			// Load current system
			system, err = m.systemRepo.GetSystemByID(ctx, m.player.CurrentSystem)
			if err != nil {
				return spaceViewLoadedMsg{
					system:      nil,
					planets:     nil,
					nearbyShips: nil,
					playerShip:  m.currentShip,
					err:         fmt.Errorf("failed to load system: %w", err),
				}
			}

			// Load planets in system
			planets, err = m.systemRepo.GetPlanetsBySystem(ctx, m.player.CurrentSystem)
			if err != nil {
				// Log error but don't fail - system may have no planets
				planets = []*models.Planet{}
			}
		}

		// Get nearby ships from presence manager
		if m.player != nil && m.presenceManager != nil {
			playersInSystem := m.presenceManager.GetPlayersInSystem(m.player.CurrentSystem)

			// Convert PlayerPresence to Ship objects for display
			// Note: This is simplified. In production, you'd load full ship data
			// from database using the ship IDs from presence data
			nearbyShips = make([]*models.Ship, 0, len(playersInSystem))

			for _, presence := range playersInSystem {
				// Skip the current player
				if presence.PlayerID == m.playerID {
					continue
				}

				// Create a simplified ship representation
				// In production, load actual ship data from database
				nearbyShips = append(nearbyShips, &models.Ship{
					ID:      uuid.New(), // Would be actual ship ID from database
					OwnerID: presence.PlayerID,
					TypeID:  presence.ShipType,
					Name:    presence.ShipName,
					Hull:    100, // Would be actual hull from database
					Shields: 100, // Would be actual shields from database
				})
			}
		}

		return spaceViewLoadedMsg{
			system:      system,
			planets:     planets,
			nearbyShips: nearbyShips,
			playerShip:  m.currentShip,
			err:         nil,
		}
	}
}

// cycleTargetCmd cycles to the next targetable object
func (m Model) cycleTargetCmd() tea.Cmd {
	return func() tea.Msg {
		// Build list of targetable objects
		targetables := []interface{}{}
		targetTypes := []string{}

		// Add planets
		for _, planet := range m.spaceView.planets {
			targetables = append(targetables, planet)
			targetTypes = append(targetTypes, "planet")
		}

		// Add ships
		for _, ship := range m.spaceView.ships {
			targetables = append(targetables, ship)
			targetTypes = append(targetTypes, ship.objType)
		}

		if len(targetables) == 0 {
			return targetSelectedMsg{
				target:     nil,
				targetType: "",
				err:        fmt.Errorf("no targetable objects in range"),
			}
		}

		// Cycle to next target
		m.spaceView.targetIndex++
		if m.spaceView.targetIndex >= len(targetables) {
			m.spaceView.targetIndex = 0
		}

		return targetSelectedMsg{
			target:     targetables[m.spaceView.targetIndex],
			targetType: targetTypes[m.spaceView.targetIndex],
			err:        nil,
		}
	}
}

// hailTargetCmd initiates communication with target
func (m Model) hailTargetCmd() tea.Cmd {
	return func() tea.Msg {
		if !m.spaceView.hasTarget {
			return errorMsg{
				context: "hail",
				err:     fmt.Errorf("no target selected"),
			}
		}

		// Check if we have a valid target
		if m.spaceView.targetIndex >= len(m.spaceView.ships) {
			return errorMsg{
				context: "hail",
				err:     fmt.Errorf("invalid target"),
			}
		}

		target := m.spaceView.ships[m.spaceView.targetIndex]
		var message string

		// Handle hailing based on target type
		switch target.objType {
		case "planet":
			// Hailing a planet - show planet info
			message = "Hailing " + target.name + "...\n"
			message += "Receiving docking clearance and planet information.\n"
			message += "Use [L] to land on this planet."

		case "player":
			// Hailing another player - could open DM chat
			message = "Hailing " + target.name + "...\n"
			message += "Opening communication channel.\n"
			message += "Use chat (DM channel) to communicate."

		case "ship", "enemy":
			// Hailing an NPC ship
			message = "Hailing " + target.name + "...\n"
			if target.hostile {
				message += target.name + ": \"You're in the wrong sector, " +
					"traveler. Turn back now!\""
			} else {
				message += target.name + ": \"Greetings! Safe travels in this system.\""
			}

		default:
			message = "Hailing " + target.name + "...\n"
			message += "No response."
		}

		return operationCompleteMsg{
			operation: "hail",
			success:   true,
			message:   message,
			err:       nil,
		}
	}
}

// triggerRandomEncounterCmd triggers a random encounter
func (m Model) triggerRandomEncounterCmd() tea.Cmd {
	return func() tea.Msg {
		if m.encounterManager == nil {
			return errorMsg{
				context: "encounter",
				err:     fmt.Errorf("encounter manager not available"),
			}
		}

		// Generate random encounter
		// encounter := m.encounterManager.GenerateEncounter(m.player.CurrentSystemID)
		// if encounter != nil {
		//     return combatInitMsg with encounter data
		// }

		// For now, just return success
		return operationCompleteMsg{
			operation: "encounter",
			success:   true,
			message:   "Encounter generated",
			err:       nil,
		}
	}
}

func (m Model) viewSpaceView() string {
	width := 80
	height := 24
	if m.width > 80 {
		width = m.width
	}
	if m.height > 24 {
		height = m.height
	}

	var sb strings.Builder

	// Calculate shield percentage
	// TODO: Get max values from ShipType when API integration is complete
	maxShields := 100
	shieldPercent := 80
	if m.currentShip != nil {
		if maxShields > 0 {
			shieldPercent = (m.currentShip.Shields * 100) / maxShields
		}
	}

	// Header
	systemName := "Unknown System"
	if m.player != nil {
		// Would load system name from database
		systemName = "Sol System"
	}
	credits := int64(0)
	if m.player != nil {
		credits = m.player.Credits
	}

	header := DrawHeader("TERMINAL VELOCITY v1.0", systemName, credits, shieldPercent, width)
	sb.WriteString(header + "\n")

	// Main content area
	contentHeight := height - 6 // Header + footer + chat
	if m.spaceView.chatExpanded {
		contentHeight -= 8 // More space for expanded chat
	}

	// Left side: Space viewport + target/cargo panels
	viewportWidth := width - 17 // Leave room for right sidebar
	viewportHeight := contentHeight - 8 // Leave room for bottom panels

	// Draw space viewport
	sb.WriteString(m.drawSpaceViewport(viewportWidth, viewportHeight))

	// Right sidebar: Radar + Status
	// TODO: Implement proper side-by-side rendering
	// rightSidebar := m.drawRightSidebar(15, viewportHeight)

	// For now, the sidebar is rendered inline below the viewport
	sb.WriteString("\n")

	// Bottom panels: Target info + Cargo
	sb.WriteString(m.drawBottomPanels(viewportWidth, 6))

	// Chat window
	if m.spaceView.chatExpanded {
		sb.WriteString(m.drawChatExpanded(width))
	} else {
		sb.WriteString(m.drawChatCollapsed(width))
	}

	// Footer
	footer := DrawFooter("[L]and  [J]ump  [T]arget  [F]ire  [H]ail  [M]ap  [C]hat  [I]nfo  [ESC] Menu", width)
	sb.WriteString("\n" + footer)

	return sb.String()
}

func (m Model) drawSpaceViewport(width, height int) string {
	var sb strings.Builder

	// Top border
	sb.WriteString(BoxVertical + "    ")
	sb.WriteString(BoxTopLeftDouble)
	sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-8))
	sb.WriteString(BoxTopRightDouble + "\n")

	// Space content
	for i := 0; i < height; i++ {
		sb.WriteString(BoxVertical + "    ")
		sb.WriteString(BoxVerticalDouble)

		// Draw space objects based on y position
		line := ""
		switch i {
		case 2:
			// Stars scattered
			line = "                          " + IconStar + "                                    "
		case 4:
			// Planet (Earth)
			line = "             " + IconStar + "                    " + IconPlanet + " Earth                      "
		case height / 2:
			// Player ship in center
			line = Center(IconShip, width-8)
			line += "\n" + BoxVertical + "    " + BoxVerticalDouble
			line += Center("You", width-8)
		case height/2 + 3:
			// Enemy ship
			line = "                                             " + IconEnemy + " Pirate          "
		case height - 3:
			// Another planet (Mars)
			line = "           " + IconPlanet + " Mars                                              "
		case 1, 6, height - 2:
			// Stars
			line = "        " + IconStar + "                                                      " + IconStar + "       "
		default:
			line = strings.Repeat(" ", width-8)
		}

		if len(line) < width-8 {
			line = PadRight(line, width-8)
		}
		sb.WriteString(line[:width-8])
		sb.WriteString(BoxVerticalDouble + "\n")
	}

	// Bottom border
	sb.WriteString(BoxVertical + "    ")
	sb.WriteString(BoxBottomLeftDouble)
	sb.WriteString(strings.Repeat(BoxHorizontalDouble, width-8))
	sb.WriteString(BoxBottomRightDouble)

	return sb.String()
}

func (m Model) drawRightSidebar(width, height int) string {
	var sb strings.Builder

	// Radar panel
	radarHeight := 13
	var radarContent strings.Builder
	radarContent.WriteString("   RADAR     \n")
	radarContent.WriteString("             \n")
	radarContent.WriteString("      " + IconStar + "      \n")
	radarContent.WriteString("             \n")
	radarContent.WriteString("   " + IconPlanet + "    " + IconEnemy + "    \n")
	radarContent.WriteString("        " + IconPlayer + "    \n")
	radarContent.WriteString("      " + IconStar + "      \n")
	radarContent.WriteString("             \n")

	radar := DrawPanel("", radarContent.String(), width, radarHeight, false)
	sb.WriteString(radar + "\n")

	// Status panel
	var statusContent strings.Builder
	// TODO: Get max values from ShipType when API integration is complete
	maxHull := 100
	maxFuel := 100
	hullPercent := 100
	fuelPercent := 67
	if m.currentShip != nil {
		if maxHull > 0 {
			hullPercent = (m.currentShip.Hull * 100) / maxHull
		}
		if maxFuel > 0 {
			fuelPercent = (m.currentShip.Fuel * 100) / maxFuel
		}
	}

	statusContent.WriteString("   STATUS    \n")
	statusContent.WriteString("━━━━━━━━━━━━━\n")
	statusContent.WriteString(fmt.Sprintf(" Hull: %s\n", DrawProgressBar(hullPercent, 100, 6)))
	statusContent.WriteString(fmt.Sprintf("       %d%%  \n", hullPercent))
	statusContent.WriteString(fmt.Sprintf(" Fuel: %s\n", DrawProgressBar(fuelPercent, 100, 6)))
	statusContent.WriteString(fmt.Sprintf("       %d%%   \n", fuelPercent))
	statusContent.WriteString(" Speed: 340  \n")

	credits := int64(52400)
	if m.player != nil {
		credits = m.player.Credits
	}
	statusContent.WriteString(" Credits:    \n")
	statusContent.WriteString(fmt.Sprintf("  %s\n", FormatCredits(credits)))

	status := DrawPanel("", statusContent.String(), width, height-radarHeight-1, false)
	sb.WriteString(status)

	return sb.String()
}

func (m Model) drawBottomPanels(width, height int) string {
	var sb strings.Builder

	// Target panel (left)
	targetWidth := 25
	var targetContent strings.Builder
	targetContent.WriteString(" TARGET: Pirate Viper    \n")
	targetContent.WriteString(" Distance: 2,340 km      \n")
	targetContent.WriteString(" Shields: 45%            \n")
	targetContent.WriteString(" Attitude: Hostile       \n")

	// Cargo panel (right)
	cargoWidth := 38
	var cargoContent strings.Builder
	cargoContent.WriteString(" CARGO: 15/50 tons                \n")
	cargoContent.WriteString(" " + IconBullet + " Food (10t)  " + IconBullet + " Electronics (5t) \n")

	sb.WriteString(BoxVertical + "  ")

	// Draw both panels inline (simplified)
	target := DrawPanel("", targetContent.String(), targetWidth, height, false)
	cargo := DrawPanel("", cargoContent.String(), cargoWidth, height, false)

	// This is simplified - actual implementation would render side-by-side
	sb.WriteString(target)
	sb.WriteString("  ")
	sb.WriteString(cargo)

	return sb.String()
}

func (m Model) drawChatCollapsed(width int) string {
	var sb strings.Builder

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxTopRight + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" CHAT [Global] " + IconArrow + "                                               [C] to expand ")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" SpaceCadet: Anyone near Sol system?                                  3m ago ")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxBottomRight + " ")
	sb.WriteString(BoxVertical)

	return sb.String()
}

func (m Model) drawChatExpanded(width int) string {
	var sb strings.Builder

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxTopLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxTopRight + " ")
	sb.WriteString(BoxVertical + "\n")

	// Chat header with channels
	channels := []string{"Global", "System", "Faction", "DM"}
	channelText := " CHAT: "
	for i, ch := range channels {
		if i == m.spaceView.chatChannel {
			channelText += "[" + ch + " " + IconArrow + "] "
		} else {
			channelText += "[" + ch + "] "
		}
	}
	channelText = PadRight(channelText, width-29) + "[C] to collapse "

	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(channelText)
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Separator
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxCrossLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxCross + " ")
	sb.WriteString(BoxVertical + "\n")

	// Chat messages
	messages := []string{
		" [SpaceCadet] Anyone near Sol system?                     3m ago ",
		" [TraderJoe] Yeah I'm docked at Earth. Need anything?     2m ago ",
		" [SpaceCadet] Looking for escort to Alpha Centauri        2m ago ",
		" [PirateKing] I'll escort you... to your doom! Arr!       1m ago ",
		" [TraderJoe] Ignore him. I can escort for 5k credits      1m ago ",
		" [YOU] I'm at Earth too, what's the pirate situation?     now    ",
	}

	for _, msg := range messages {
		sb.WriteString(BoxVertical + " ")
		sb.WriteString(BoxVertical)
		sb.WriteString(PadRight(msg, width-4))
		sb.WriteString(BoxVertical + " ")
		sb.WriteString(BoxVertical + "\n")
	}

	// Empty line
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(strings.Repeat(" ", width-4))
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Message input
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical)
	sb.WriteString(" Message: [" + PadRight(m.spaceView.chatInput+"_", width-16) + "]")
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxVertical + "\n")

	// Bottom border
	sb.WriteString(BoxVertical + " ")
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-4))
	sb.WriteString(BoxBottomRight + " ")
	sb.WriteString(BoxVertical)

	return sb.String()
}

func (m Model) updateSpaceView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c", "C":
			// Toggle chat
			m.spaceView.chatExpanded = !m.spaceView.chatExpanded
			return m, nil

		case "l", "L":
			// Land on planet (if near one)
			m.screen = ScreenLanding
			return m, nil

		case "j", "J":
			// Jump (navigation)
			m.screen = ScreenNavigationEnhanced
			return m, nil

		case "t", "T":
			// Target next object
			return m, m.cycleTargetCmd()

		case "h", "H":
			// Hail target
			return m, m.hailTargetCmd()

		case "f", "F":
			// Fire / Enter combat
			if m.spaceView.hasTarget {
				m.screen = ScreenCombatEnhanced
				return m, nil
			} else {
				return m, func() tea.Msg {
					return errorMsg{
						context: "combat",
						err:     fmt.Errorf("no target selected - press T to target"),
					}
				}
			}

		case "m", "M":
			// System map
			m.screen = ScreenNavigationEnhanced
			return m, nil

		case "i", "I":
			// Player info
			// TODO: Implement ScreenPlayerInfo
			// m.screen = ScreenPlayerInfo
			return m, nil

		case "esc":
			// Menu
			m.screen = ScreenMainMenu
			return m, nil

		default:
			// Handle chat input if expanded
			if m.spaceView.chatExpanded {
				if msg.String() == "enter" {
					// Send chat message
					if m.chatManager != nil && m.spaceView.chatInput != "" {
						username := "Player"
						if m.player != nil {
							username = m.player.Username
						}

						// Send to appropriate channel
						switch m.spaceView.chatChannel {
						case 0: // Global
							m.chatManager.SendGlobalMessage(m.playerID, username, m.spaceView.chatInput)

						case 1: // System
							// Get current system ID and other players in system
							if m.player != nil && m.presenceManager != nil {
								systemID := m.player.CurrentSystem
								playersInSystem := m.presenceManager.GetPlayersInSystem(systemID)

								// Extract player IDs for recipients
								recipientIDs := make([]uuid.UUID, 0, len(playersInSystem))
								for _, presence := range playersInSystem {
									recipientIDs = append(recipientIDs, presence.PlayerID)
								}

								m.chatManager.SendSystemMessage(systemID, m.playerID, username, m.spaceView.chatInput, recipientIDs)
							}

						case 2: // Faction
							// Get faction ID and member IDs
							if m.player != nil && m.player.IsInFaction() && m.factionManager != nil {
								factionID := *m.player.FactionID
								faction, err := m.factionManager.GetFaction(factionID)
								if err == nil && faction != nil {
									// Convert UUID to string for faction ID
									factionIDStr := factionID.String()
									memberIDs := faction.Members

									m.chatManager.SendFactionMessage(factionIDStr, m.playerID, username, m.spaceView.chatInput, memberIDs)
								} else {
									// Player not in faction or faction not found
									// Could show error message here
								}
							} else {
								// Player not in a faction
								// Could show error message here
							}

						case 3: // DM
							// Get recipient from targeted ship
							// Note: This requires the space view to have actual player ship data
							// For now, this is a placeholder that checks if we have a valid target
							if m.spaceView.hasTarget && m.spaceView.targetIndex < len(m.spaceView.ships) {
								targetShip := m.spaceView.ships[m.spaceView.targetIndex]

								// In production, the spaceObject would need a playerID field
								// loaded from presenceManager. For now, we log that DM requires
								// proper target selection with player IDs
								_ = targetShip // Avoid unused variable warning

								// TODO: When space view data loading is implemented, this will use:
								// - targetShip.playerID for recipientID
								// - targetShip.playerName for recipient name
								// m.chatManager.SendDirectMessage(m.playerID, username, recipientID, recipientName, m.spaceView.chatInput)

								// For now, could show a message that DM requires target selection
							}
						}

						m.spaceView.chatInput = ""
					}
					return m, nil
				} else if msg.String() == "backspace" {
					if len(m.spaceView.chatInput) > 0 {
						m.spaceView.chatInput = m.spaceView.chatInput[:len(m.spaceView.chatInput)-1]
					}
					return m, nil
				} else if len(msg.String()) == 1 {
					// Add character to chat input
					m.spaceView.chatInput += msg.String()
					return m, nil
				}
			}
		}

	case spaceViewLoadedMsg:
		// Handle space view data loaded
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
		} else {
			// Update space view data with loaded planets
			m.spaceView.planets = msg.planets

			// Convert ships and planets to spaceObjects for rendering
			shipObjects := convertShipsToSpaceObjects(msg.nearbyShips, m.spaceView.player)
			planetObjects := convertPlanetsToSpaceObjects(msg.planets, m.spaceView.player)

			// Combine all objects (ships + planets)
			allObjects := make([]spaceObject, 0, len(shipObjects)+len(planetObjects))
			allObjects = append(allObjects, planetObjects...)
			allObjects = append(allObjects, shipObjects...)
			m.spaceView.ships = allObjects

			// Reset target if it's now out of range
			if m.spaceView.targetIndex >= len(allObjects) {
				m.spaceView.targetIndex = 0
				m.spaceView.hasTarget = len(allObjects) > 0
			}
		}
		return m, nil

	case targetSelectedMsg:
		// Handle target selection
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
			m.spaceView.hasTarget = false
		} else {
			// Target selected successfully
			m.spaceView.hasTarget = true
			// The target display is automatically updated in the view function
			// which reads from m.spaceView.ships[m.spaceView.targetIndex]
		}
		return m, nil

	case errorMsg:
		// Handle errors
		m.errorMessage = msg.err.Error()
		m.showErrorDialog = true
		return m, nil

	case operationCompleteMsg:
		// Handle operation complete
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.showErrorDialog = true
		} else {
			// Show success message if needed
			// For hail, could trigger dialog or chat
		}
		return m, nil
	}

	return m, nil
}

// Add ScreenSpaceView and ScreenLanding constants to Screen enum when integrating
