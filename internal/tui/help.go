// File: internal/tui/help.go
// Project: Terminal Velocity
// Version: 1.0.0

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/help"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Help view modes
const (
	helpViewTopics   = "topics"   // List of help topics
	helpViewContent  = "content"  // Viewing a specific topic
	helpViewTutorial = "tutorial" // Tutorial progress
	helpViewQuick    = "quick"    // Quick reference
)

type helpModel struct {
	viewMode     string
	cursor       int
	topics       []help.HelpTopic
	currentTopic *help.HelpTopic
	scroll       int // For scrolling long content
}

func newHelpModel() helpModel {
	return helpModel{
		viewMode: helpViewTopics,
		cursor:   0,
		topics:   help.GetAllTopics(),
		scroll:   0,
	}
}

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
		} else if m.helpModel.cursor-2 < len(m.helpModel.topics) {
			// Regular topic
			m.helpModel.currentTopic = &m.helpModel.topics[m.helpModel.cursor-2]
			m.helpModel.viewMode = helpViewContent
			m.helpModel.scroll = 0
		}

	case "q", "esc":
		m.screen = ScreenMainMenu

	}

	return m, nil
}

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

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
