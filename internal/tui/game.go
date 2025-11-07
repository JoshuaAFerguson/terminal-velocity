package tui

import (
	"github.com/charmbracelet/bubbletea"
)

type gameViewModel struct {
	// Game state will go here
}

func (m Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.screen = ScreenMainMenu
			return m, nil
		}
	}

	return m, nil
}

func (m Model) viewGame() string {
	s := renderHeader(m.username, m.player.Credits, "Space")
	s += "\n"

	content := `You are floating in space.

Your ship's systems hum quietly as you gaze out at the stars.

Commands:
  n - Navigation
  t - Trading
  s - Shipyard
  m - Missions

Press ESC to return to main menu.`

	s += boxStyle.Render(content)

	s += renderFooter("ESC: Main Menu  •  n/t/s/m: Quick Access")

	return s
}
