// File: internal/tui/pilot_record.go
// Project: Terminal Velocity
// Description: Pilot Record screen — lifetime stats, ratings, and milestones.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// pilotRecordModel keeps minimal state — everything the screen renders comes
// from m.player, m.currentShip, m.currentSystem, so the model itself only
// needs a scroll cursor for when the stat list outgrows the viewport.
type pilotRecordModel struct {
	scroll int
}

func newPilotRecordModel() pilotRecordModel {
	return pilotRecordModel{}
}

func (m Model) updatePilotRecord(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			// Honor the previousScreen stack so this screen is safe to
			// reach from Landing, Space View, or the main menu.
			if m.hasPreviousScreen {
				m.screen = m.previousScreen
				m.hasPreviousScreen = false
			} else {
				m.screen = ScreenMainMenu
			}
			return m, nil
		case "up", "k":
			if m.pilotRecord.scroll > 0 {
				m.pilotRecord.scroll--
			}
			return m, nil
		case "down", "j":
			m.pilotRecord.scroll++
			return m, nil
		}
	}
	return m, nil
}

// viewPilotRecord renders a read-only summary of the player's career so far.
// Every section reads directly from the cached Player struct — the values
// are kept current by the existing RecordJump / RecordTrade / RecordKill
// hooks that already write to both the in-memory player and the DB on the
// corresponding game events.
func (m Model) viewPilotRecord() string {
	var credits int64
	if m.player != nil {
		credits = m.player.Credits
	}
	s := renderHeader(m.username, credits, m.currentLocationLabel())
	s += "\n"
	s += subtitleStyle.Render("=== Pilot Record ===") + "\n\n"

	if m.player == nil {
		s += "No pilot data available.\n"
		s += renderFooter("ESC: Main Menu")
		return s
	}

	p := m.player
	var b strings.Builder

	// Ratings block — the three Escape Velocity-style disciplines.
	b.WriteString(subtitleStyle.Render("Ratings") + "\n")
	b.WriteString(fmt.Sprintf("  Combat:       %4d   %s\n", p.CombatRating, p.GetCombatRankTitle()))
	b.WriteString(fmt.Sprintf("  Trading:      %4d   %s\n", p.TradingRating, p.GetTradingRankTitle()))
	b.WriteString(fmt.Sprintf("  Exploration:  %4d\n", p.ExplorationRating))
	b.WriteString("\n")

	// Licences block — tiered combat ratings unlock progressively heavier
	// weapon purchases at the Outfitter. Mirrors the thresholds set on
	// StandardWeapons.MinCombatRating so what's shown here is what the
	// catalog actually gates on.
	b.WriteString(subtitleStyle.Render("Licences") + "\n")
	combatTiers := []struct {
		cr    int
		label string
	}{
		{10, "Basic Combat Licence      — Beam Laser, Missile Launcher"},
		{25, "Competent Combat Licence  — Plasma Cannon, Plasma Turret"},
		{40, "Skilled Combat Licence    — Heavy Laser"},
		{50, "Seasoned Combat Licence   — Torpedo Launcher"},
		{60, "Expert Combat Licence     — Railgun"},
		{80, "Elite Combat Licence      — Heavy Railgun"},
	}
	any := false
	for _, t := range combatTiers {
		mark := "  "
		if p.CombatRating >= t.cr {
			mark = successStyle.Render("✓ ")
			any = true
		} else {
			mark = helpStyle.Render("  ")
		}
		b.WriteString(fmt.Sprintf("  %s[CR %2d]  %s\n", mark, t.cr, t.label))
	}
	if !any {
		b.WriteString(helpStyle.Render("  (earn combat rating to unlock heavier weapons)") + "\n")
	}
	b.WriteString("\n")

	// Milestones block
	b.WriteString(subtitleStyle.Render("Milestones") + "\n")
	b.WriteString(fmt.Sprintf("  Ships destroyed:    %d\n", p.TotalKills))
	b.WriteString(fmt.Sprintf("  Jumps logged:       %d\n", p.TotalJumps))
	b.WriteString(fmt.Sprintf("  Systems visited:    %d\n", p.SystemsVisited))
	b.WriteString(fmt.Sprintf("  Trades executed:    %d\n", p.TotalTrades))
	b.WriteString(fmt.Sprintf("  Lifetime profit:    %s cr\n", formatCreditsSigned(p.TradeProfit)))
	b.WriteString(fmt.Sprintf("  Best single trade:  %s cr\n", formatCreditsSigned(p.HighestProfit)))
	b.WriteString(fmt.Sprintf("  Missions completed: %d\n", p.MissionsCompleted))
	b.WriteString(fmt.Sprintf("  Missions failed:    %d\n", p.MissionsFailed))
	b.WriteString(fmt.Sprintf("  Quests completed:   %d\n", p.QuestsCompleted))
	b.WriteString("\n")

	// Specialist stats — hide the block entirely when the player hasn't
	// done any of it, so fresh pilots aren't staring at a wall of zeros.
	if p.TotalCaptureAttempts > 0 || p.TotalMiningOps > 0 || p.TotalCrafts > 0 {
		b.WriteString(subtitleStyle.Render("Specialist") + "\n")
		if p.TotalCaptureAttempts > 0 {
			b.WriteString(fmt.Sprintf("  Boarding actions:   %d (%d successful)\n",
				p.TotalCaptureAttempts, p.SuccessfulBoards))
			b.WriteString(fmt.Sprintf("  Captures:           %d\n", p.SuccessfulCaptures))
		}
		if p.TotalMiningOps > 0 {
			b.WriteString(fmt.Sprintf("  Mining ops:         %d (%d t total)\n",
				p.TotalMiningOps, p.TotalYield))
		}
		if p.TotalCrafts > 0 {
			b.WriteString(fmt.Sprintf("  Items crafted:      %d  (skill %d)\n",
				p.TotalCrafts, p.CraftingSkill))
		}
		b.WriteString("\n")
	}

	// Pilot identity block — tenure + criminal status, useful context for
	// faction interactions.
	b.WriteString(subtitleStyle.Render("Record") + "\n")
	if !p.CreatedAt.IsZero() {
		b.WriteString(fmt.Sprintf("  Enlisted:           %s\n", p.CreatedAt.Format("2006-01-02")))
		b.WriteString(fmt.Sprintf("  Years of service:   %s\n", formatDurationYears(time.Since(p.CreatedAt))))
	}
	if p.IsCriminal {
		b.WriteString("  Status:             " + errorStyle.Render("WANTED — criminal flag set") + "\n")
	} else {
		b.WriteString("  Status:             Clean record\n")
	}
	if len(p.Reputation) > 0 {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Faction Standing") + "\n")
		// Sort factions so the list is stable between renders.
		type kv struct {
			faction string
			rep     int
		}
		rows := make([]kv, 0, len(p.Reputation))
		for f, r := range p.Reputation {
			rows = append(rows, kv{f, r})
		}
		// Simple insertion sort — map iteration order is random, and the
		// list is short enough that sort.Slice would be overkill.
		for i := 1; i < len(rows); i++ {
			for j := i; j > 0 && rows[j-1].faction > rows[j].faction; j-- {
				rows[j-1], rows[j] = rows[j], rows[j-1]
			}
		}
		for _, r := range rows {
			b.WriteString(fmt.Sprintf("  %-28s %s\n", r.faction, reputationLabel(r.rep)))
		}
	}

	lines := strings.Split(b.String(), "\n")
	if m.pilotRecord.scroll >= len(lines) {
		m.pilotRecord.scroll = len(lines) - 1
	}
	if m.pilotRecord.scroll > 0 {
		lines = lines[m.pilotRecord.scroll:]
	}
	s += strings.Join(lines, "\n")

	s += renderFooter("↑/↓: Scroll   ESC: Back")
	return s
}

// reputationLabel renders a reputation int as the EV-style ladder so the
// player sees Friendly / Hostile rather than a bare number.
func reputationLabel(rep int) string {
	switch {
	case rep >= 75:
		return fmt.Sprintf("%4d  Allied", rep)
	case rep >= 25:
		return fmt.Sprintf("%4d  Friendly", rep)
	case rep >= -24:
		return fmt.Sprintf("%4d  Neutral", rep)
	case rep >= -74:
		return fmt.Sprintf("%4d  Unfriendly", rep)
	default:
		return fmt.Sprintf("%4d  Hostile", rep)
	}
}

// formatCreditsSigned renders a credit value with a thousands separator and
// an explicit + prefix for gains, so the Pilot Record distinguishes lifetime
// net-positive traders from net-negative ones at a glance.
func formatCreditsSigned(v int64) string {
	if v > 0 {
		return "+" + formatThousands(v)
	}
	return formatThousands(v)
}

// formatDurationYears renders a time.Duration as "1y 42d 3h" with the
// leading zero-value fields trimmed; reasonable for "years of service"
// display rather than precise elapsed-time math.
func formatDurationYears(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	days := int(d / (24 * time.Hour))
	years := days / 365
	days -= years * 365
	hours := int((d % (24 * time.Hour)) / time.Hour)
	switch {
	case years > 0:
		return fmt.Sprintf("%dy %dd", years, days)
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	default:
		return fmt.Sprintf("%dh", hours)
	}
}

