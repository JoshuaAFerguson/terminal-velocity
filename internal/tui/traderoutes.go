// File: internal/tui/traderoutes.go
// Project: Terminal Velocity
// Description: Trade routes and navigation planning screen
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/traderoutes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// Trade routes screen state
type tradeRoutesState struct {
	routes         []*traderoutes.TradeRoute
	selectedIndex  int
	loading        bool
	error          string
	mode           string // "best", "from_here", "plan"
	navPath        *traderoutes.NavigationPath
	targetSystemID uuid.UUID
}

func (m *Model) updateTradeRoutes(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.screen = ScreenGame
			return m, nil

		case "up", "k":
			if m.tradeRoutes.selectedIndex > 0 {
				m.tradeRoutes.selectedIndex--
			}

		case "down", "j":
			if m.tradeRoutes.selectedIndex < len(m.tradeRoutes.routes)-1 {
				m.tradeRoutes.selectedIndex++
			}

		case "1":
			// Find best routes globally
			m.tradeRoutes.mode = "best"
			m.tradeRoutes.loading = true
			return m, m.loadBestRoutes()

		case "2":
			// Find routes from current system
			m.tradeRoutes.mode = "from_here"
			m.tradeRoutes.loading = true
			return m, m.loadRoutesFromHere()

		case "3":
			// Plan navigation to selected route destination
			if len(m.tradeRoutes.routes) > 0 && m.tradeRoutes.selectedIndex < len(m.tradeRoutes.routes) {
				route := m.tradeRoutes.routes[m.tradeRoutes.selectedIndex]
				m.tradeRoutes.mode = "plan"
				m.tradeRoutes.loading = true
				m.tradeRoutes.targetSystemID = route.ToSystem.ID
				return m, m.planNavigation(route.ToSystem.ID)
			}

		case "enter":
			// Set destination to selected route's target system
			if len(m.tradeRoutes.routes) > 0 && m.tradeRoutes.selectedIndex < len(m.tradeRoutes.routes) {
				route := m.tradeRoutes.routes[m.tradeRoutes.selectedIndex]
				m.targetSystemID = route.ToSystem.ID
				m.statusMessage = fmt.Sprintf("Destination set: %s", route.ToSystem.Name)
				m.screen = ScreenNavigation
				return m, nil
			}
		}

	case tradeRoutesLoadedMsg:
		m.tradeRoutes.loading = false
		m.tradeRoutes.routes = msg.routes
		m.tradeRoutes.selectedIndex = 0
		m.tradeRoutes.error = msg.err

	case navigationPlanLoadedMsg:
		m.tradeRoutes.loading = false
		m.tradeRoutes.navPath = msg.path
		m.tradeRoutes.error = msg.err
	}

	return m, nil
}

func (m *Model) viewTradeRoutes() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("═══ TRADE ROUTES & NAVIGATION ═══")

	b.WriteString(title + "\n\n")

	if m.tradeRoutes.loading {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("Loading...") + "\n\n")
		return b.String()
	}

	if m.tradeRoutes.error != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("Error: "+m.tradeRoutes.error) + "\n\n")
	}

	// Mode selector
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render("Mode:") + " ")

	if m.tradeRoutes.mode == "best" {
		b.WriteString("[1] Best Routes Globally  ")
	} else {
		b.WriteString("[1] Best Routes  ")
	}

	if m.tradeRoutes.mode == "from_here" {
		b.WriteString("[2] From Current System  ")
	} else {
		b.WriteString("[2] From Here  ")
	}

	if m.tradeRoutes.mode == "plan" {
		b.WriteString("[3] Navigation Plan")
	} else {
		b.WriteString("[3] Plan Route")
	}

	b.WriteString("\n\n")

	// Display navigation plan if in plan mode
	if m.tradeRoutes.mode == "plan" && m.tradeRoutes.navPath != nil {
		b.WriteString(m.renderNavigationPlan())
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("[Q] Back") + "\n")
		return b.String()
	}

	// Display routes
	if len(m.tradeRoutes.routes) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("No profitable routes found. Press [1] or [2] to search.") + "\n\n")
	} else {
		// Header
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		b.WriteString(headerStyle.Render(
			fmt.Sprintf("%-20s %-20s %-12s %8s %8s %8s %5s %6s\n",
				"From", "To", "Commodity", "Buy", "Sell", "Profit", "Jumps", "ROI")))

		b.WriteString(strings.Repeat("─", 100) + "\n")

		// Routes (show max 15)
		maxDisplay := 15
		if len(m.tradeRoutes.routes) < maxDisplay {
			maxDisplay = len(m.tradeRoutes.routes)
		}

		for i := 0; i < maxDisplay; i++ {
			route := m.tradeRoutes.routes[i]

			style := lipgloss.NewStyle()
			if i == m.tradeRoutes.selectedIndex {
				style = style.Foreground(lipgloss.Color("11")).Bold(true)
			} else {
				style = style.Foreground(lipgloss.Color("7"))
			}

			cursor := "  "
			if i == m.tradeRoutes.selectedIndex {
				cursor = "→ "
			}

			line := fmt.Sprintf("%s%-20s %-20s %-12s %8.0f %8.0f %8.0f %5d %5.1f%%",
				cursor,
				truncate(route.FromSystem.Name, 18),
				truncate(route.ToSystem.Name, 18),
				truncate(route.Commodity, 10),
				route.BuyPrice,
				route.SellPrice,
				route.ProfitPerUnit,
				len(route.JumpPath)-1,
				route.ROI,
			)

			b.WriteString(style.Render(line) + "\n")
		}

		b.WriteString("\n")

		// Selected route details
		if m.tradeRoutes.selectedIndex < len(m.tradeRoutes.routes) {
			route := m.tradeRoutes.routes[m.tradeRoutes.selectedIndex]
			b.WriteString(m.renderRouteDetails(route))
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("[↑/↓] Navigate  [Enter] Set Destination  [3] Plan Route  [Q] Back") + "\n")

	return b.String()
}

func (m *Model) renderRouteDetails(route *traderoutes.TradeRoute) string {
	var b strings.Builder

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(80)

	var details strings.Builder
	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Route Details") + "\n\n")

	details.WriteString(fmt.Sprintf("From:          %s\n", route.FromSystem.Name))
	details.WriteString(fmt.Sprintf("To:            %s\n", route.ToSystem.Name))
	details.WriteString(fmt.Sprintf("Commodity:     %s\n", route.Commodity))
	details.WriteString(fmt.Sprintf("Buy Price:     %.0f CR\n", route.BuyPrice))
	details.WriteString(fmt.Sprintf("Sell Price:    %.0f CR\n", route.SellPrice))
	details.WriteString(fmt.Sprintf("Profit/Unit:   %.0f CR\n", route.ProfitPerUnit))
	details.WriteString(fmt.Sprintf("Total Profit:  %d CR (full cargo)\n", route.TotalProfit))
	details.WriteString(fmt.Sprintf("ROI:           %.1f%%\n", route.ROI))
	details.WriteString(fmt.Sprintf("Distance:      %d units\n", route.Distance))
	details.WriteString(fmt.Sprintf("Jumps:         %d\n", len(route.JumpPath)-1))
	details.WriteString(fmt.Sprintf("Profit/Jump:   %.0f CR\n", route.ProfitPerJump))

	b.WriteString(detailStyle.Render(details.String()))
	return b.String()
}

func (m *Model) renderNavigationPlan() string {
	var b strings.Builder

	if m.tradeRoutes.navPath == nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("No navigation plan available")
	}

	nav := m.tradeRoutes.navPath

	planStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(80)

	var plan strings.Builder
	plan.WriteString(lipgloss.NewStyle().Bold(true).Render("Navigation Plan") + "\n\n")

	plan.WriteString(fmt.Sprintf("Total Jumps:    %d\n", nav.TotalJumps))
	plan.WriteString(fmt.Sprintf("Distance:       %d units\n", nav.TotalDistance))
	plan.WriteString(fmt.Sprintf("Fuel Required:  %d\n", nav.FuelRequired))
	plan.WriteString("\n")

	plan.WriteString(lipgloss.NewStyle().Bold(true).Render("Waypoints:") + "\n")
	for i, waypoint := range nav.Waypoints {
		if i == 0 {
			plan.WriteString(fmt.Sprintf("  %d. %s (Start)\n", i+1, waypoint))
		} else if i == len(nav.Waypoints)-1 {
			plan.WriteString(fmt.Sprintf("  %d. %s (Destination)\n", i+1, waypoint))
		} else {
			plan.WriteString(fmt.Sprintf("  %d. %s\n", i+1, waypoint))
		}
	}

	b.WriteString(planStyle.Render(plan.String()))
	return b.String()
}

// Commands

func (m *Model) loadBestRoutes() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		calculator := traderoutes.NewCalculator(m.systemRepo, m.marketRepo)

		opts := traderoutes.DefaultRouteOptions()
		opts.CargoCapacity = m.currentShip.ShipType.CargoSpace

		routes, err := calculator.FindBestRoutes(ctx, opts)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		return tradeRoutesLoadedMsg{
			routes: routes,
			err:    errStr,
		}
	}
}

func (m *Model) loadRoutesFromHere() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		calculator := traderoutes.NewCalculator(m.systemRepo, m.marketRepo)

		opts := traderoutes.DefaultRouteOptions()
		opts.CargoCapacity = m.currentShip.ShipType.CargoSpace

		routes, err := calculator.FindRoutesFromSystem(ctx, m.player.CurrentSystemID, opts)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		return tradeRoutesLoadedMsg{
			routes: routes,
			err:    errStr,
		}
	}
}

func (m *Model) planNavigation(targetID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		calculator := traderoutes.NewCalculator(m.systemRepo, m.marketRepo)

		path, err := calculator.PlanRoute(ctx, m.player.CurrentSystemID, targetID)

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		return navigationPlanLoadedMsg{
			path: path,
			err:  errStr,
		}
	}
}

// Messages

type tradeRoutesLoadedMsg struct {
	routes []*traderoutes.TradeRoute
	err    string
}

type navigationPlanLoadedMsg struct {
	path *traderoutes.NavigationPath
	err  string
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}
