// File: internal/tutorial/manager.go
// Project: Terminal Velocity
// Description: Tutorial management and progression system
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package tutorial

import (
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Manager handles tutorial progression and state
type Manager struct {
	mu        sync.RWMutex
	tutorials map[string]*models.Tutorial              // tutorialID -> Tutorial
	progress  map[uuid.UUID]*models.TutorialProgress   // playerID -> Progress
	triggers  map[models.TutorialTrigger][]string      // trigger -> tutorialIDs
}

// NewManager creates a new tutorial manager
func NewManager() *Manager {
	m := &Manager{
		tutorials: make(map[string]*models.Tutorial),
		progress:  make(map[uuid.UUID]*models.TutorialProgress),
		triggers:  make(map[models.TutorialTrigger][]string),
	}

	// Initialize default tutorials
	m.initializeDefaultTutorials()

	return m
}

// initializeDefaultTutorials creates the standard tutorial sequence
func (m *Manager) initializeDefaultTutorials() {
	// Basics Tutorial
	basicsTutorial := models.NewTutorial(
		"tutorial_basics",
		"Welcome to Terminal Velocity",
		"Learn the basics of navigating and playing the game",
		models.TutorialBasics,
	)
	basicsTutorial.OrderIndex = 1
	basicsTutorial.AddStep(&models.TutorialStep{
		ID:          "basics_1_welcome",
		Title:       "Welcome, Commander!",
		Description: "Welcome to Terminal Velocity, a multiplayer space trading and combat game.",
		Screen:      "main_menu",
		Objective:   "Read the welcome message and press Enter to continue",
		Hints:       []string{"Press Enter or Space to continue", "You can skip tutorials at any time with 'S'"},
		OrderIndex:  1,
	})
	basicsTutorial.AddStep(&models.TutorialStep{
		ID:          "basics_2_navigation",
		Title:       "Navigation Basics",
		Description: "Use arrow keys or vim keys (hjkl) to navigate menus.",
		Screen:      "main_menu",
		Objective:   "Navigate up and down the menu",
		Hints:       []string{"Use ↑/↓ or j/k to move", "Press Enter to select an item", "Press ESC or q to go back"},
		OrderIndex:  2,
	})
	basicsTutorial.AddStep(&models.TutorialStep{
		ID:          "basics_3_credits",
		Title:       "Your Credits",
		Description: "Credits are the currency in Terminal Velocity. You start with 10,000 credits.",
		Screen:      "main_menu",
		Objective:   "Check your credit balance at the top of the screen",
		Hints:       []string{"Your credits are shown in the header", "Use credits to buy ships, equipment, and cargo"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(basicsTutorial)
	m.AddTrigger(models.TriggerFirstLogin, "tutorial_basics")

	// Trading Tutorial
	tradingTutorial := models.NewTutorial(
		"tutorial_trading",
		"Space Trading 101",
		"Learn how to buy and sell commodities for profit",
		models.TutorialTrading,
	)
	tradingTutorial.OrderIndex = 2
	tradingTutorial.Prerequisites = []string{"tutorial_basics"}
	tradingTutorial.AddStep(&models.TutorialStep{
		ID:          "trading_1_market",
		Title:       "The Trading Market",
		Description: "Each planet has a market where you can buy and sell commodities.",
		Screen:      "trading",
		Objective:   "Select 'Trading' from the main menu",
		Hints:       []string{"Navigate to Trading in the menu", "Press Enter to open the trading screen"},
		OrderIndex:  1,
	})
	tradingTutorial.AddStep(&models.TutorialStep{
		ID:          "trading_2_buy",
		Title:       "Buying Commodities",
		Description: "Buy commodities low and sell them high at other planets for profit.",
		Screen:      "trading",
		Objective:   "Buy at least one unit of any commodity",
		Hints:       []string{"Select a commodity to see its price", "Press 'B' to buy", "Check your cargo space!"},
		OrderIndex:  2,
	})
	tradingTutorial.AddStep(&models.TutorialStep{
		ID:          "trading_3_prices",
		Title:       "Understanding Prices",
		Description: "Commodity prices vary by planet based on supply and demand.",
		Screen:      "trading",
		Objective:   "View the price history of a commodity",
		Hints:       []string{"Prices change over time", "Some planets produce certain goods cheaper", "Look for trade routes!"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(tradingTutorial)
	m.AddTrigger(models.TriggerFirstTrade, "tutorial_trading")

	// Navigation Tutorial
	navigationTutorial := models.NewTutorial(
		"tutorial_navigation",
		"Galactic Navigation",
		"Learn how to jump between star systems",
		models.TutorialBasics,
	)
	navigationTutorial.OrderIndex = 3
	navigationTutorial.Prerequisites = []string{"tutorial_basics"}
	navigationTutorial.AddStep(&models.TutorialStep{
		ID:          "nav_1_starmap",
		Title:       "The Star Map",
		Description: "The navigation screen shows nearby star systems you can jump to.",
		Screen:      "navigation",
		Objective:   "Open the Navigation screen",
		Hints:       []string{"Select Navigation from the main menu", "Each system shows jump cost and distance"},
		OrderIndex:  1,
	})
	navigationTutorial.AddStep(&models.TutorialStep{
		ID:          "nav_2_jump",
		Title:       "Jumping to Systems",
		Description: "Jumping between systems costs fuel. Make sure you have enough!",
		Screen:      "navigation",
		Objective:   "Jump to a neighboring system",
		Hints:       []string{"Select a system and press Enter", "Fuel cost is shown for each jump", "You can't jump if you don't have enough fuel"},
		OrderIndex:  2,
	})
	m.RegisterTutorial(navigationTutorial)
	m.AddTrigger(models.TriggerScreenEnter, "tutorial_navigation")

	// Ship Management Tutorial
	shipTutorial := models.NewTutorial(
		"tutorial_ships",
		"Ship Management",
		"Learn about upgrading and managing your spacecraft",
		models.TutorialShips,
	)
	shipTutorial.OrderIndex = 4
	shipTutorial.Prerequisites = []string{"tutorial_basics", "tutorial_trading"}
	shipTutorial.AddStep(&models.TutorialStep{
		ID:          "ships_1_shipyard",
		Title:       "The Shipyard",
		Description: "Visit the shipyard to buy new ships or upgrade your current one.",
		Screen:      "shipyard",
		Objective:   "Open the Shipyard screen",
		Hints:       []string{"Different ships have different cargo capacities", "Combat ships have more weapon slots", "Larger ships cost more to operate"},
		OrderIndex:  1,
	})
	shipTutorial.AddStep(&models.TutorialStep{
		ID:          "ships_2_outfitter",
		Title:       "Ship Outfitting",
		Description: "The outfitter lets you install weapons, shields, and other equipment.",
		Screen:      "outfitter",
		Objective:   "Open the Outfitter screen",
		Hints:       []string{"Equipment requires outfit space", "Better equipment costs more", "Balance offense and defense!"},
		OrderIndex:  2,
	})
	shipTutorial.AddStep(&models.TutorialStep{
		ID:          "ships_3_cargo",
		Title:       "Managing Cargo",
		Description: "Your cargo hold has limited space. Manage it wisely!",
		Screen:      "cargo",
		Objective:   "Open the Cargo Hold screen",
		Hints:       []string{"You can jettison cargo if needed", "Some missions require specific cargo", "Cargo space affects your ship's value"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(shipTutorial)

	// Combat Tutorial
	combatTutorial := models.NewTutorial(
		"tutorial_combat",
		"Space Combat Basics",
		"Learn how to survive and win in space combat",
		models.TutorialCombat,
	)
	combatTutorial.OrderIndex = 5
	combatTutorial.Prerequisites = []string{"tutorial_basics"}
	combatTutorial.AddStep(&models.TutorialStep{
		ID:          "combat_1_encounters",
		Title:       "Random Encounters",
		Description: "You may encounter pirates, police, or other ships while traveling.",
		Screen:      "encounter",
		Objective:   "Understand encounter types",
		Hints:       []string{"Not all encounters are hostile", "You can try to flee", "Some ships are much stronger than others!"},
		OrderIndex:  1,
	})
	combatTutorial.AddStep(&models.TutorialStep{
		ID:          "combat_2_fighting",
		Title:       "Combat Actions",
		Description: "During combat, you can attack, defend, use abilities, or attempt to flee.",
		Screen:      "combat",
		Objective:   "Survive your first combat encounter",
		Hints:       []string{"Watch your shield and hull integrity", "Some weapons are better against shields", "Fleeing is sometimes the best option"},
		OrderIndex:  2,
	})
	combatTutorial.AddStep(&models.TutorialStep{
		ID:          "combat_3_strategy",
		Title:       "Combat Strategy",
		Description: "Different ships and equipment require different tactics.",
		Screen:      "combat",
		Objective:   "Win a combat encounter",
		Hints:       []string{"Use shields to absorb damage", "Target enemy weapons first", "Don't forget special abilities!"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(combatTutorial)
	m.AddTrigger(models.TriggerFirstCombat, "tutorial_combat")

	// Missions Tutorial
	missionsTutorial := models.NewTutorial(
		"tutorial_missions",
		"Mission System",
		"Learn how to accept and complete missions",
		models.TutorialMissions,
	)
	missionsTutorial.OrderIndex = 6
	missionsTutorial.Prerequisites = []string{"tutorial_basics", "tutorial_navigation"}
	missionsTutorial.AddStep(&models.TutorialStep{
		ID:          "missions_1_board",
		Title:       "Mission Board",
		Description: "The mission board shows available missions from various factions.",
		Screen:      "missions",
		Objective:   "Open the Missions screen",
		Hints:       []string{"Missions offer credits and reputation", "Some missions have time limits", "Read requirements carefully!"},
		OrderIndex:  1,
	})
	missionsTutorial.AddStep(&models.TutorialStep{
		ID:          "missions_2_accept",
		Title:       "Accepting Missions",
		Description: "You can accept multiple missions, but cargo missions require cargo space.",
		Screen:      "missions",
		Objective:   "Accept your first mission",
		Hints:       []string{"Check the destination", "Make sure you have required cargo space", "Some missions are dangerous!"},
		OrderIndex:  2,
	})
	missionsTutorial.AddStep(&models.TutorialStep{
		ID:          "missions_3_complete",
		Title:       "Completing Missions",
		Description: "Complete missions by meeting their objectives and returning to the destination.",
		Screen:      "missions",
		Objective:   "Complete a mission",
		Hints:       []string{"Track active missions in the mission screen", "Rewards are given upon completion", "Reputation affects available missions"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(missionsTutorial)
	m.AddTrigger(models.TriggerFirstMission, "tutorial_missions")

	// Multiplayer Tutorial
	multiplayerTutorial := models.NewTutorial(
		"tutorial_multiplayer",
		"Multiplayer Features",
		"Learn about playing with other commanders",
		models.TutorialMultiplayer,
	)
	multiplayerTutorial.OrderIndex = 7
	multiplayerTutorial.IsOptional = true
	multiplayerTutorial.AddStep(&models.TutorialStep{
		ID:          "multi_1_players",
		Title:       "Other Players",
		Description: "Terminal Velocity is a multiplayer game. See who else is online!",
		Screen:      "players",
		Objective:   "Open the Players screen",
		Hints:       []string{"See what other players are doing", "Check their ships and locations", "Rankings show top players"},
		OrderIndex:  1,
	})
	multiplayerTutorial.AddStep(&models.TutorialStep{
		ID:          "multi_2_chat",
		Title:       "Chat System",
		Description: "Communicate with other players using the chat system.",
		Screen:      "chat",
		Objective:   "Send a message in chat",
		Hints:       []string{"Be respectful to other players", "You can use different chat channels", "Admins monitor chat"},
		OrderIndex:  2,
	})
	multiplayerTutorial.AddStep(&models.TutorialStep{
		ID:          "multi_3_factions",
		Title:       "Factions",
		Description: "Join or create factions to play with other commanders.",
		Screen:      "factions",
		Objective:   "View the factions screen",
		Hints:       []string{"Factions can control territory", "Work together for faction goals", "Faction reputation matters"},
		OrderIndex:  3,
	})
	m.RegisterTutorial(multiplayerTutorial)
}

// RegisterTutorial adds a tutorial to the manager
func (m *Manager) RegisterTutorial(tutorial *models.Tutorial) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tutorials[tutorial.ID] = tutorial
}

// AddTrigger associates a trigger with a tutorial
func (m *Manager) AddTrigger(trigger models.TutorialTrigger, tutorialID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.triggers[trigger] == nil {
		m.triggers[trigger] = make([]string, 0)
	}
	m.triggers[trigger] = append(m.triggers[trigger], tutorialID)
}

// GetPlayerProgress returns the tutorial progress for a player
func (m *Manager) GetPlayerProgress(playerID uuid.UUID) *models.TutorialProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return nil
	}
	return progress
}

// InitializePlayer creates tutorial progress for a new player
func (m *Manager) InitializePlayer(playerID uuid.UUID) *models.TutorialProgress {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress := models.NewTutorialProgress(playerID)
	progress.TotalSteps = m.countTotalSteps()
	m.progress[playerID] = progress

	return progress
}

// countTotalSteps counts all tutorial steps across all tutorials
func (m *Manager) countTotalSteps() int {
	total := 0
	for _, tutorial := range m.tutorials {
		total += len(tutorial.Steps)
	}
	return total
}

// CompleteStep marks a tutorial step as completed
func (m *Manager) CompleteStep(playerID uuid.UUID, stepID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return
	}

	// Find which tutorial this step belongs to
	for _, tutorial := range m.tutorials {
		for _, step := range tutorial.Steps {
			if step.ID == stepID {
				progress.CompleteStep(stepID, tutorial.Category)
				step.Completed = true
				return
			}
		}
	}
}

// SkipStep marks a tutorial step as skipped
func (m *Manager) SkipStep(playerID uuid.UUID, stepID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return
	}

	progress.SkipStep(stepID)
}

// SkipTutorial skips all steps in a tutorial
func (m *Manager) SkipTutorial(playerID uuid.UUID, tutorialID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return
	}

	tutorial, exists := m.tutorials[tutorialID]
	if !exists {
		return
	}

	for _, step := range tutorial.Steps {
		if !progress.IsStepCompleted(step.ID) {
			progress.SkipStep(step.ID)
		}
	}
}

// DisableTutorials disables tutorials for a player
func (m *Manager) DisableTutorials(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return
	}

	progress.TutorialEnabled = false
}

// EnableTutorials enables tutorials for a player
func (m *Manager) EnableTutorials(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return
	}

	progress.TutorialEnabled = true
}

// GetCurrentTutorial returns the current active tutorial for a player
func (m *Manager) GetCurrentTutorial(playerID uuid.UUID) *models.Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progress[playerID]
	if !exists || !progress.TutorialEnabled {
		return nil
	}

	// Find the first incomplete tutorial that has prerequisites met
	for _, tutorial := range m.getSortedTutorials() {
		if !tutorial.IsCompleted(progress) && m.prerequisitesMet(tutorial, progress) {
			return tutorial
		}
	}

	return nil
}

// GetCurrentStep returns the current active step for a player
func (m *Manager) GetCurrentStep(playerID uuid.UUID) *models.TutorialStep {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tutorial := m.GetCurrentTutorial(playerID)
	if tutorial == nil {
		return nil
	}

	progress := m.progress[playerID]
	return tutorial.GetNextIncompleteStep(progress)
}

// GetTutorialForScreen returns active tutorial steps for a specific screen
func (m *Manager) GetTutorialForScreen(playerID uuid.UUID, screen string) *models.TutorialStep {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progress[playerID]
	if !exists || !progress.TutorialEnabled {
		return nil
	}

	step := m.GetCurrentStep(playerID)
	if step != nil && step.Screen == screen {
		return step
	}

	return nil
}

// HandleTrigger processes a tutorial trigger event
func (m *Manager) HandleTrigger(playerID uuid.UUID, trigger models.TutorialTrigger) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, exists := m.progress[playerID]
	if !exists || !progress.TutorialEnabled {
		return
	}

	tutorialIDs, exists := m.triggers[trigger]
	if !exists {
		return
	}

	// Activate triggered tutorials if prerequisites are met
	for _, tutorialID := range tutorialIDs {
		tutorial, exists := m.tutorials[tutorialID]
		if !exists {
			continue
		}

		if !tutorial.IsCompleted(progress) && m.prerequisitesMet(tutorial, progress) {
			progress.CurrentStep = tutorial.Steps[0].ID
		}
	}
}

// prerequisitesMet checks if tutorial prerequisites are satisfied
func (m *Manager) prerequisitesMet(tutorial *models.Tutorial, progress *models.TutorialProgress) bool {
	for _, prereqID := range tutorial.Prerequisites {
		prereqTutorial, exists := m.tutorials[prereqID]
		if !exists {
			continue
		}

		if !prereqTutorial.IsCompleted(progress) {
			return false
		}
	}
	return true
}

// getSortedTutorials returns tutorials sorted by OrderIndex
func (m *Manager) getSortedTutorials() []*models.Tutorial {
	tutorials := make([]*models.Tutorial, 0, len(m.tutorials))
	for _, tutorial := range m.tutorials {
		tutorials = append(tutorials, tutorial)
	}

	// Simple bubble sort by OrderIndex
	for i := 0; i < len(tutorials); i++ {
		for j := i + 1; j < len(tutorials); j++ {
			if tutorials[i].OrderIndex > tutorials[j].OrderIndex {
				tutorials[i], tutorials[j] = tutorials[j], tutorials[i]
			}
		}
	}

	return tutorials
}

// GetAllTutorials returns all registered tutorials
func (m *Manager) GetAllTutorials() []*models.Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getSortedTutorials()
}

// GetTutorial returns a specific tutorial by ID
func (m *Manager) GetTutorial(tutorialID string) *models.Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.tutorials[tutorialID]
}

// GetStats returns tutorial statistics for a player
func (m *Manager) GetStats(playerID uuid.UUID) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progress[playerID]
	if !exists {
		return map[string]interface{}{
			"initialized": false,
		}
	}

	return map[string]interface{}{
		"initialized":          true,
		"enabled":              progress.TutorialEnabled,
		"total_steps":          progress.TotalSteps,
		"completed_steps":      progress.CompletedCount,
		"skipped_steps":        len(progress.SkippedSteps),
		"completion_percent":   progress.GetCompletionPercentage(),
		"categories_started":   len(progress.CategoryProgress),
		"current_tutorial":     m.GetCurrentTutorial(playerID),
	}
}
