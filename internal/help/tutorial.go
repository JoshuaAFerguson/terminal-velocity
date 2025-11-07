// File: internal/help/tutorial.go
// Project: Terminal Velocity
// Description: Interactive tutorial system for new players
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package help

// TutorialStep represents a single step in the tutorial
type TutorialStep struct {
	ID          string
	Title       string
	Description string
	Objective   string
	Hint        string
	Completed   bool
}

// TutorialProgress tracks a player's tutorial progress
type TutorialProgress struct {
	CurrentStep int
	Steps       []TutorialStep
	Completed   bool
}

// NewTutorialProgress creates a new tutorial progression
func NewTutorialProgress() *TutorialProgress {
	return &TutorialProgress{
		CurrentStep: 0,
		Steps:       GetTutorialSteps(),
		Completed:   false,
	}
}

// GetTutorialSteps returns all tutorial steps
func GetTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			ID:          "welcome",
			Title:       "Welcome to Terminal Velocity!",
			Description: "You've successfully connected to the game via SSH. You're starting with a basic Shuttle and 10,000 credits. Your goal is to build wealth, upgrade your ship, and become a legendary pilot.",
			Objective:   "Navigate to the Trading screen from the Main Menu",
			Hint:        "Use ↑/↓ or J/K to navigate, Enter to select",
			Completed:   false,
		},
		{
			ID:          "trading_basics",
			Title:       "Trading 101",
			Description: "Trading is the primary way to earn credits. Buy commodities where they're cheap, sell where they're expensive. Each system has different prices based on supply, demand, and tech level.",
			Objective:   "Buy any commodity (start with Food or Water)",
			Hint:        "Select a commodity and press 'B' to buy. Check your credits!",
			Completed:   false,
		},
		{
			ID:          "check_cargo",
			Title:       "Cargo Management",
			Description: "Your cargo hold shows what you're carrying. The Shuttle has 50 tons capacity. You can jettison cargo you don't want to make room for more profitable goods.",
			Objective:   "Visit the Cargo Hold screen to see your purchases",
			Hint:        "Navigate to 'Cargo Hold' from the Main Menu",
			Completed:   false,
		},
		{
			ID:          "navigation",
			Title:       "Navigation & Travel",
			Description: "Time to find a better market! The Navigation screen shows connected star systems. Each jump costs fuel based on distance. Look for systems with different government or tech levels for better prices.",
			Objective:   "Jump to a different star system",
			Hint:        "Navigate to 'Navigation', select a system, press Enter",
			Completed:   false,
		},
		{
			ID:          "selling",
			Title:       "Sell for Profit",
			Description: "Now you're in a different system with different prices. Check if your cargo is worth more here than what you paid. Selling for profit is how you build wealth!",
			Objective:   "Sell your cargo for more than you paid",
			Hint:        "Go to Trading, select your commodity, press 'S' to sell",
			Completed:   false,
		},
		{
			ID:          "trade_route",
			Title:       "Profitable Trade Routes",
			Description: "You've completed your first trade! Real profit comes from finding reliable routes. Look for: Low-tech → High-tech (Machinery), High-tech → Low-tech (Luxury Goods). Tech level matters!",
			Objective:   "Make 5,000 credits profit from trading",
			Hint:        "Repeat buy-jump-sell until you have 15,000+ credits",
			Completed:   false,
		},
		{
			ID:          "shipyard_visit",
			Title:       "Ship Upgrades",
			Description: "With 50,000+ credits, you can afford your first upgrade! The Light Freighter has 3x the cargo capacity (150 tons) and can equip weapons. More cargo = more profit per trip.",
			Objective:   "Visit the Shipyard (you don't have to buy yet)",
			Hint:        "Navigate to 'Shipyard' to see available ships",
			Completed:   false,
		},
		{
			ID:          "outfitter_basics",
			Title:       "Weapons & Equipment",
			Description: "The Outfitter sells weapons and equipment. Weapons let you defend yourself and accept combat missions. Equipment includes shields (protection), engines (speed), and cargo expansions.",
			Objective:   "Visit the Outfitter to see available equipment",
			Hint:        "Navigate to 'Outfitter' from Main Menu",
			Completed:   false,
		},
		{
			ID:          "missions_intro",
			Title:       "Mission Board",
			Description: "Missions provide guaranteed income and reputation. Types: Cargo Delivery (easy), Bounty Hunt (combat), Patrol (combat), Assassination (hard combat). Start with cargo missions!",
			Objective:   "View the Mission Board",
			Hint:        "Navigate to 'Missions' from Main Menu",
			Completed:   false,
		},
		{
			ID:          "multiplayer_intro",
			Title:       "Multiplayer Features",
			Description: "Terminal Velocity is multiplayer! See online players, chat with them, trade, form factions, and engage in PvP. Check the Players screen to see who's online right now.",
			Objective:   "Visit the Players screen",
			Hint:        "Navigate to 'Players' from Main Menu",
			Completed:   false,
		},
		{
			ID:          "chat_system",
			Title:       "Communication",
			Description: "The Chat system has multiple channels: Global (all players), System (same location), Faction (your faction), and Direct Messages. Say hi to fellow pilots!",
			Objective:   "Open the Chat screen",
			Hint:        "Navigate to 'Chat' from Main Menu. Press 'I' or Enter to send messages",
			Completed:   false,
		},
		{
			ID:          "achievements",
			Title:       "Track Your Progress",
			Description: "Achievements track your accomplishments. Leaderboards show top pilots. News shows universe events. These screens help you set goals and see how you compare to others.",
			Objective:   "Check your Achievements",
			Hint:        "Navigate to 'Achievements' from Main Menu",
			Completed:   false,
		},
		{
			ID:          "tutorial_complete",
			Title:       "Tutorial Complete!",
			Description: "Congratulations! You now know the basics of Terminal Velocity. Your journey has just begun. Build your trading empire, join a faction, claim territory, or become a feared pirate. The universe is yours!",
			Objective:   "Continue playing and exploring!",
			Hint:        "Check the Help screen anytime for detailed guides",
			Completed:   false,
		},
	}
}

// GetCurrentStep returns the current tutorial step
func (t *TutorialProgress) GetCurrentStep() *TutorialStep {
	if t.CurrentStep >= len(t.Steps) {
		return nil
	}
	return &t.Steps[t.CurrentStep]
}

// CompleteStep marks the current step as complete and advances
func (t *TutorialProgress) CompleteStep() bool {
	if t.CurrentStep >= len(t.Steps) {
		return false
	}

	t.Steps[t.CurrentStep].Completed = true
	t.CurrentStep++

	// Check if tutorial is complete
	if t.CurrentStep >= len(t.Steps) {
		t.Completed = true
	}

	return true
}

// GetProgress returns the completion percentage
func (t *TutorialProgress) GetProgress() float64 {
	if len(t.Steps) == 0 {
		return 0.0
	}
	return (float64(t.CurrentStep) / float64(len(t.Steps))) * 100.0
}

// IsComplete checks if the tutorial is finished
func (t *TutorialProgress) IsComplete() bool {
	return t.Completed
}

// Reset resets the tutorial progress
func (t *TutorialProgress) Reset() {
	t.CurrentStep = 0
	t.Completed = false
	for i := range t.Steps {
		t.Steps[i].Completed = false
	}
}

// SkipTutorial marks the tutorial as complete without doing steps
func (t *TutorialProgress) SkipTutorial() {
	t.CurrentStep = len(t.Steps)
	t.Completed = true
	for i := range t.Steps {
		t.Steps[i].Completed = true
	}
}

// CheckObjective checks if a given action completes the current step's objective
func (t *TutorialProgress) CheckObjective(action string) bool {
	step := t.GetCurrentStep()
	if step == nil {
		return false
	}

	// Match action to step ID
	actionMatches := map[string]string{
		"view_trading":      "welcome",
		"buy_commodity":     "trading_basics",
		"view_cargo":        "check_cargo",
		"jump_system":       "navigation",
		"sell_commodity":    "selling",
		"profit_5000":       "trade_route",
		"view_shipyard":     "shipyard_visit",
		"view_outfitter":    "outfitter_basics",
		"view_missions":     "missions_intro",
		"view_players":      "multiplayer_intro",
		"view_chat":         "chat_system",
		"view_achievements": "achievements",
	}

	if expectedAction, exists := actionMatches[step.ID]; exists {
		if action == expectedAction || action == step.ID {
			return t.CompleteStep()
		}
	}

	return false
}
