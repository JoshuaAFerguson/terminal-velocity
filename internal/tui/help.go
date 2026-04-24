// File: internal/tui/help.go
// Project: Terminal Velocity
// Description: Help screen - Comprehensive help topics, tutorial, and quick reference
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07
//
// The help screen provides:
// - Interactive help topic browser
// - Tutorial system with step tracking
// - Quick reference card for common commands
// - Scrollable content for long help texts
// - Context-sensitive help topics
// - Key bindings documentation
// - Getting started guides
//
// View Modes:
//   - topics: List of all help topics
//   - content: Viewing a specific help topic
//   - tutorial: Tutorial progress and current step
//   - quick: Quick reference card
//
// Help Topics:
//   - Getting Started
//   - Navigation
//   - Trading
//   - Combat
//   - Ships
//   - Multiplayer
//   - Keyboard Shortcuts
//
// Tutorial System:
//   - 7 categories covering core gameplay
//   - 20+ tutorial steps
//   - Step completion tracking
//   - Progress indicators (✅/⬜)
//   - Objectives and hints
//   - Can skip tutorial anytime
//
// Visual Features:
//   - Topic icons (🚀🗺️💰⚔️👥⌨️)
//   - Scroll indicators for long content
//   - Line numbers for content position
//   - Active selection highlighting

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Help view modes

const (
	helpViewTopics   = "topics"   // List of help topics
	helpViewContent  = "content"  // Viewing a specific topic
	helpViewTutorial = "tutorial" // Tutorial progress
	helpViewQuick    = "quick"    // Quick reference
)

// helpModel contains the state for the help screen.
// Manages topic browsing, content viewing, and scroll position.
type helpModel struct {
	viewMode     string           // Current view: "topics", "content", "tutorial", "quick"
	cursor       int              // Current cursor position in topic list
	topics       []help.HelpTopic // Available help topics
	currentTopic *help.HelpTopic  // Topic being viewed in detail
	scroll       int              // Scroll offset for long content
}

// newHelpModel creates and initializes a new help screen model.
// Loads all help topics and starts in topics view.
func newHelpModel() helpModel {
	return helpModel{
		viewMode: helpViewTopics,
		cursor:   0,
		topics:   help.GetAllTopics(),
		scroll:   0,
	}
}

// updateHelp handles input and state updates for the help screen.
// Routes to mode-specific update handlers.
//
// View Mode Routing:
//   - content: Handled by updateHelpContent()
//   - tutorial: Handled by updateHelpTutorial()
//   - quick: Handled by updateHelpQuick()
//   - topics: Handled by updateHelpTopics()
func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.helpModel.viewMode {
		case helpViewContent:
			return m.updateHelpContent(msg)
		case helpViewTutorial:
			return m.updateHelpTutorial(msg)
		case helpViewQuick:
			return m.updateHelpQuick(msg)
		default:
			return m.updateHelpTopics(msg)
		}
	}

	return m, nil
}

// updateHelpTopics handles input in topics list view.
// Manages navigation and topic/tutorial/quick ref selection.
func (m Model) updateHelpTopics(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.helpModel.cursor > 0 {
			m.helpModel.cursor--
		}

	case "down", "j":
		// +2 for tutorial and quick reference options
		maxItems := len(m.helpModel.topics) + 2
		if m.helpModel.cursor < maxItems-1 {
			m.helpModel.cursor++
		}

	case "enter":
		// Check if tutorial or quick ref selected
		if m.helpModel.cursor == 0 {
			// Tutorial
			m.helpModel.viewMode = helpViewTutorial
			m.helpModel.scroll = 0
		} else if m.helpModel.cursor == 1 {
			// Quick Reference
			m.helpModel.viewMode = helpViewQuick
			m.helpModel.scroll = 0
		} else if m.helpModel.cursor >= 2 && m.helpModel.cursor-2 < len(m.helpModel.topics) {
			// Regular topic - check bounds to prevent negative index
			m.helpModel.currentTopic = &m.helpModel.topics[m.helpModel.cursor-2]
			m.helpModel.viewMode = helpViewContent
			m.helpModel.scroll = 0
		}

	case "q", "esc":
		m.screen = ScreenMainMenu

	}

	return m, nil
}

// updateHelpContent handles input in topic content view.
// Manages scrolling through help text.
func (m Model) updateHelpContent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.helpModel.scroll > 0 {
			m.helpModel.scroll--
		}

	case "down", "j":
		m.helpModel.scroll++

	case "esc", "q", "backspace":
		m.helpModel.viewMode = helpViewTopics
		m.helpModel.currentTopic = nil
		m.helpModel.scroll = 0
	}

	return m, nil
}

// updateHelpTutorial handles input in tutorial view.
// Manages scrolling through tutorial content.
func (m Model) updateHelpTutorial(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.helpModel.scroll > 0 {
			m.helpModel.scroll--
		}

	case "down", "j":
		m.helpModel.scroll++

	case "esc", "q", "backspace":
		m.helpModel.viewMode = helpViewTopics
		m.helpModel.scroll = 0
	}

	return m, nil
}

// updateHelpQuick handles input in quick reference view.
// Manages scrolling through quick reference card.
func (m Model) updateHelpQuick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.helpModel.scroll > 0 {
			m.helpModel.scroll--
		}

	case "down", "j":
		m.helpModel.scroll++

	case "esc", "q", "backspace":
		m.helpModel.viewMode = helpViewTopics
		m.helpModel.scroll = 0
	}

	return m, nil
}

// viewHelp renders the help screen (dispatches to mode-specific views).
func (m Model) viewHelp() string {
	switch m.helpModel.viewMode {
	case helpViewContent:
		return m.viewHelpContent()
	case helpViewTutorial:
		return m.viewHelpTutorial()
	case helpViewQuick:
		return m.viewHelpQuick()
	default:
		return m.viewHelpTopics()
	}
}

// viewHelpTopics renders the help topics list with tutorial and quick ref options.
func (m Model) viewHelpTopics() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)

	var s strings.Builder
	s.WriteString(titleStyle.Render("📖 Help & Documentation"))
	s.WriteString("\n\n")

	s.WriteString("Select a topic to learn more:\n\n")

	// Special options
	specialOptions := []string{
		"🎓 Interactive Tutorial",
		"⚡ Quick Reference Card",
	}

	for i, option := range specialOptions {
		cursor := "  "
		if m.helpModel.cursor == i {
			cursor = "→ "
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, option))
	}

	s.WriteString("\n")

	// Regular topics
	for i, topic := range m.helpModel.topics {
		cursor := "  "
		if m.helpModel.cursor == i+2 {
			cursor = "→ "
		}

		icon := "📄"
		switch topic.ID {
		case "getting_started":
			icon = "🚀"
		case "navigation":
			icon = "🗺️"
		case "trading":
			icon = "💰"
		case "combat":
			icon = "⚔️"
		case "ships":
			icon = "🚀"
		case "multiplayer":
			icon = "👥"
		case "shortcuts":
			icon = "⌨️"
		}

		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, icon, topic.Title))
	}

	s.WriteString("\n")
	s.WriteString("Controls: [↑/↓] Navigate [Enter] Select [Q] Back\n")

	return boxStyle.Render(s.String())
}

// viewHelpContent renders a help topic's content with scroll support.
// Shows key bindings if available for the topic.
func (m Model) viewHelpContent() string {
	if m.helpModel.currentTopic == nil {
		return "No topic selected"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)

	topic := m.helpModel.currentTopic

	var s strings.Builder
	s.WriteString(titleStyle.Render("📖 " + topic.Title))
	s.WriteString("\n\n")

	// Split content into lines and apply scroll
	lines := strings.Split(topic.Content, "\n")
	visibleLines := 20 // Number of lines to show at once

	startLine := m.helpModel.scroll
	if startLine > len(lines)-visibleLines {
		startLine = max(0, len(lines)-visibleLines)
	}
	endLine := min(len(lines), startLine+visibleLines)

	for i := startLine; i < endLine; i++ {
		s.WriteString(lines[i] + "\n")
	}

	// Show scroll indicator if needed
	if len(lines) > visibleLines {
		s.WriteString(fmt.Sprintf("\n[Showing lines %d-%d of %d | Use ↑/↓ to scroll]\n",
			startLine+1, endLine, len(lines)))
	}

	// Show key bindings if available
	if len(topic.KeyBindings) > 0 {
		s.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("Key Bindings:") + "\n")
		for _, kb := range topic.KeyBindings {
			s.WriteString(fmt.Sprintf("  %s - %s\n", kb.Key, kb.Description))
		}
	}

	s.WriteString("\nControls: [↑/↓] Scroll [Q/Esc] Back to Topics\n")

	return boxStyle.Render(s.String())
}

// viewHelpTutorial renders the tutorial system with step tracking.
// Shows overview, current step, and completion status.
func (m Model) viewHelpTutorial() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("🎓 Interactive Tutorial"))
	s.WriteString("\n\n")

	s.WriteString(labelStyle.Render("TUTORIAL OVERVIEW") + "\n\n")

	s.WriteString("The interactive tutorial will guide you through:\n\n")

	steps := help.GetTutorialSteps()
	for i, step := range steps {
		status := "⬜"
		if step.Completed {
			status = "✅"
		}
		s.WriteString(fmt.Sprintf("%s %d. %s\n", status, i+1, step.Title))
	}

	s.WriteString("\n")
	s.WriteString(labelStyle.Render("HOW IT WORKS") + "\n\n")
	s.WriteString("The tutorial tracks your progress automatically as you play.\n")
	s.WriteString("Complete objectives to advance through the steps.\n")
	s.WriteString("You can skip the tutorial anytime from the Settings screen.\n\n")

	s.WriteString(labelStyle.Render("CURRENT STEP") + "\n\n")

	// Show first uncompleted step
	currentStep := 0
	for i, step := range steps {
		if !step.Completed {
			currentStep = i
			break
		}
	}

	if currentStep < len(steps) {
		step := steps[currentStep]
		s.WriteString(fmt.Sprintf("Step %d: %s\n\n", currentStep+1, step.Title))
		s.WriteString(step.Description + "\n\n")
		s.WriteString(labelStyle.Render("Objective: ") + step.Objective + "\n")
		s.WriteString(labelStyle.Render("Hint: ") + step.Hint + "\n")
	} else {
		s.WriteString("✅ Tutorial Complete! You're ready to conquer the universe!\n")
	}

	s.WriteString("\n")
	s.WriteString("Controls: [↑/↓] Scroll [Q/Esc] Back\n")

	return boxStyle.Render(s.String())
}

// viewHelpQuick renders the quick reference card with scrolling.
func (m Model) viewHelpQuick() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)

	var s strings.Builder
	s.WriteString(titleStyle.Render("⚡ Quick Reference"))
	s.WriteString("\n\n")

	// Get quick reference content
	content := help.GetQuickReference()

	// Split and scroll
	lines := strings.Split(content, "\n")
	visibleLines := 25

	startLine := m.helpModel.scroll
	if startLine > len(lines)-visibleLines {
		startLine = max(0, len(lines)-visibleLines)
	}
	endLine := min(len(lines), startLine+visibleLines)

	for i := startLine; i < endLine; i++ {
		s.WriteString(lines[i] + "\n")
	}

	if len(lines) > visibleLines {
		s.WriteString(fmt.Sprintf("\n[Lines %d-%d of %d | ↑/↓ to scroll]\n",
			startLine+1, endLine, len(lines)))
	}

	s.WriteString("\nControls: [↑/↓] Scroll [Q/Esc] Back\n")

	return boxStyle.Render(s.String())
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
