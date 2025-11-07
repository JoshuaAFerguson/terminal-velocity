// File: internal/tui/tutorial.go
// Project: Terminal Velocity
// Description: Tutorial UI and overlay system
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package tui

import (
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/charmbracelet/bubbletea"
)

// Tutorial view modes
const (
	tutorialViewOverlay = "overlay" // Shows as overlay on current screen
	tutorialViewFull    = "full"    // Full screen tutorial view
	tutorialViewList    = "list"    // List all tutorials
)

type tutorialModel struct {
	viewMode     string
	cursor       int
	hintLevel    models.TutorialHintLevel
	showOverlay  bool
	currentStep  *models.TutorialStep
	allTutorials []*models.Tutorial
}

func newTutorialModel() tutorialModel {
	return tutorialModel{
		viewMode:     tutorialViewOverlay,
		cursor:       0,
		hintLevel:    models.HintNone,
		showOverlay:  true,
		currentStep:  nil,
		allTutorials: []*models.Tutorial{},
	}
}

func (m Model) updateTutorial(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.tutorialModel.viewMode == tutorialViewList {
				m.screen = ScreenMainMenu
			}
			return m, nil

		case "up", "k":
			if m.tutorialModel.cursor > 0 {
				m.tutorialModel.cursor--
			}
			return m, nil

		case "down", "j":
			maxCursor := m.getTutorialMaxCursor()
			if m.tutorialModel.cursor < maxCursor {
				m.tutorialModel.cursor++
			}
			return m, nil

		case "h":
			// Cycle through hint levels
			if m.tutorialModel.currentStep != nil {
				m.tutorialModel.hintLevel++
				if int(m.tutorialModel.hintLevel) >= len(m.tutorialModel.currentStep.Hints) {
					m.tutorialModel.hintLevel = models.HintNone
				}
			}
			return m, nil

		case "s":
			// Skip current step
			if m.tutorialModel.currentStep != nil && m.tutorialManager != nil {
				m.tutorialManager.SkipStep(m.playerID, m.tutorialModel.currentStep.ID)
				m.tutorialModel.currentStep = m.tutorialManager.GetCurrentStep(m.playerID)
			}
			return m, nil

		case "enter", " ":
			// Complete current step or select tutorial
			if m.tutorialModel.viewMode == tutorialViewList {
				// View selected tutorial details
				return m, nil
			}
			if m.tutorialModel.currentStep != nil && m.tutorialManager != nil {
				m.tutorialManager.CompleteStep(m.playerID, m.tutorialModel.currentStep.ID)
				m.tutorialModel.currentStep = m.tutorialManager.GetCurrentStep(m.playerID)
				m.tutorialModel.hintLevel = models.HintNone
			}
			return m, nil

		case "d":
			// Disable tutorials
			if m.tutorialManager != nil {
				m.tutorialManager.DisableTutorials(m.playerID)
				m.tutorialModel.showOverlay = false
			}
			return m, nil

		case "t":
			// Toggle tutorial overlay
			m.tutorialModel.showOverlay = !m.tutorialModel.showOverlay
			return m, nil
		}
	}

	return m, nil
}

func (m Model) getTutorialMaxCursor() int {
	switch m.tutorialModel.viewMode {
	case tutorialViewList:
		return len(m.tutorialModel.allTutorials) - 1
	default:
		return 0
	}
}

func (m Model) viewTutorial() string {
	if m.tutorialManager == nil {
		return errorView("Tutorial system not initialized")
	}

	s := renderHeader(m.username, m.player.Credits, "Tutorials")
	s += "\n"

	switch m.tutorialModel.viewMode {
	case tutorialViewList:
		s += m.viewTutorialList()
	case tutorialViewFull:
		s += m.viewTutorialFull()
	default:
		s += helpStyle.Render("Unknown tutorial view mode") + "\n"
	}

	return s
}

func (m Model) viewTutorialList() string {
	s := subtitleStyle.Render("=== Available Tutorials ===") + "\n\n"

	progress := m.tutorialManager.GetPlayerProgress(m.playerID)
	if progress == nil {
		s += helpStyle.Render("No tutorial progress found") + "\n"
		return s
	}

	tutorials := m.tutorialManager.GetAllTutorials()
	if len(tutorials) == 0 {
		s += helpStyle.Render("No tutorials available") + "\n"
		return s
	}

	for i, tutorial := range tutorials {
		completed, total := tutorial.GetProgress(progress)
		percent := 0.0
		if total > 0 {
			percent = float64(completed) / float64(total) * 100
		}

		statusIcon := "○"
		if completed == total {
			statusIcon = "✓"
		} else if completed > 0 {
			statusIcon = "◐"
		}

		line := fmt.Sprintf("%s %s (%d/%d steps - %.0f%%)",
			statusIcon,
			tutorial.Title,
			completed,
			total,
			percent,
		)

		if i == m.tutorialModel.cursor {
			s += "> " + selectedMenuItemStyle.Render(line) + "\n"
			s += "  " + helpStyle.Render(tutorial.Description) + "\n"
		} else {
			s += "  " + line + "\n"
		}
	}

	s += "\n" + renderFooter("↑/↓: Navigate  •  Enter: View  •  ESC: Back")
	return s
}

func (m Model) viewTutorialFull() string {
	if m.tutorialModel.currentStep == nil {
		s := subtitleStyle.Render("All Tutorials Complete!") + "\n\n"
		s += helpStyle.Render("You've completed all available tutorials.") + "\n"
		s += helpStyle.Render("You can review them anytime from the tutorial menu.") + "\n\n"
		s += renderFooter("ESC: Back")
		return s
	}

	step := m.tutorialModel.currentStep
	s := subtitleStyle.Render("=== "+step.Title+" ===") + "\n\n"

	s += step.Description + "\n\n"

	s += highlightStyle.Render("Objective: ") + step.Objective + "\n\n"

	// Show hints based on hint level
	if m.tutorialModel.hintLevel > models.HintNone && len(step.Hints) > 0 {
		hintIndex := int(m.tutorialModel.hintLevel) - 1
		if hintIndex < len(step.Hints) {
			s += highlightStyle.Render("Hint: ") + step.Hints[hintIndex] + "\n\n"
		}
	}

	// Progress indicator
	progress := m.tutorialManager.GetPlayerProgress(m.playerID)
	if progress != nil {
		s += fmt.Sprintf("Overall Progress: %d/%d steps (%.0f%%)\n\n",
			progress.CompletedCount,
			progress.TotalSteps,
			progress.GetCompletionPercentage(),
		)
	}

	s += renderFooter("H: Show Hint  •  Enter: Complete  •  S: Skip  •  D: Disable  •  ESC: Back")
	return s
}

// renderTutorialOverlay renders a tutorial overlay on top of the current screen
func (m Model) renderTutorialOverlay(screenContent string) string {
	if !m.tutorialModel.showOverlay || m.tutorialManager == nil {
		return screenContent
	}

	progress := m.tutorialManager.GetPlayerProgress(m.playerID)
	if progress == nil || !progress.TutorialEnabled {
		return screenContent
	}

	// Get current step for the active screen
	screenName := m.getScreenName()
	step := m.tutorialManager.GetTutorialForScreen(m.playerID, screenName)
	if step == nil {
		return screenContent
	}

	// Build overlay box
	width := 60
	lines := []string{
		"",
		"┌" + strings.Repeat("─", width-2) + "┐",
		"│ " + centerText("📚 TUTORIAL", width-4) + " │",
		"├" + strings.Repeat("─", width-2) + "┤",
	}

	// Add title
	lines = append(lines, "│ "+highlightStyle.Render(step.Title)+strings.Repeat(" ", width-4-len(step.Title))+" │")
	lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")

	// Add description (wrap text)
	descLines := wrapText(step.Description, width-4)
	for _, line := range descLines {
		lines = append(lines, "│ "+line+strings.Repeat(" ", width-4-len(line))+" │")
	}
	lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")

	// Add objective
	lines = append(lines, "│ "+highlightStyle.Render("Objective:")+strings.Repeat(" ", width-14)+" │")
	objLines := wrapText(step.Objective, width-4)
	for _, line := range objLines {
		lines = append(lines, "│ "+line+strings.Repeat(" ", width-4-len(line))+" │")
	}

	// Add hint if requested
	if m.tutorialModel.hintLevel > models.HintNone && len(step.Hints) > 0 {
		hintIndex := int(m.tutorialModel.hintLevel) - 1
		if hintIndex < len(step.Hints) {
			lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
			lines = append(lines, "│ "+helpStyle.Render("Hint:")+strings.Repeat(" ", width-10)+" │")
			hintLines := wrapText(step.Hints[hintIndex], width-4)
			for _, line := range hintLines {
				lines = append(lines, "│ "+line+strings.Repeat(" ", width-4-len(line))+" │")
			}
		}
	}

	// Add controls
	lines = append(lines, "├"+strings.Repeat("─", width-2)+"┤")
	lines = append(lines, "│ "+helpStyle.Render("H: Hint  •  S: Skip  •  T: Hide  •  D: Disable")+strings.Repeat(" ", width-48)+" │")
	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")
	lines = append(lines, "")

	overlay := strings.Join(lines, "\n")

	// Append overlay to screen content
	return screenContent + "\n\n" + overlay
}

// getScreenName returns the name of the current screen for tutorial matching
func (m Model) getScreenName() string {
	switch m.screen {
	case ScreenMainMenu:
		return "main_menu"
	case ScreenGame:
		return "game"
	case ScreenNavigation:
		return "navigation"
	case ScreenTrading:
		return "trading"
	case ScreenCargo:
		return "cargo"
	case ScreenShipyard:
		return "shipyard"
	case ScreenOutfitter:
		return "outfitter"
	case ScreenShipManagement:
		return "ship_management"
	case ScreenCombat:
		return "combat"
	case ScreenMissions:
		return "missions"
	case ScreenAchievements:
		return "achievements"
	case ScreenEncounter:
		return "encounter"
	case ScreenNews:
		return "news"
	case ScreenLeaderboards:
		return "leaderboards"
	case ScreenPlayers:
		return "players"
	case ScreenChat:
		return "chat"
	case ScreenFactions:
		return "factions"
	case ScreenTrade:
		return "trade"
	case ScreenPvP:
		return "pvp"
	case ScreenHelp:
		return "help"
	case ScreenSettings:
		return "settings"
	case ScreenAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// wrapText wraps text to a specified width
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// centerText centers text within a given width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-padding-len(text))
}
