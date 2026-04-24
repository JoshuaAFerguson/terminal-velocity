// File: internal/tui/news_ticker.go
// Project: Terminal Velocity
// Description: Scrolling newsreel ticker for the main menu — pulls
//   headlines from the news.Manager and advances a one-cell-per-tick
//   rolling window across a joined ASCII feed. The tick lifecycle is
//   tied to the main-menu screen so idle screens don't re-render.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tui

import (
	"strings"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// newsTickerMsg drives the ticker scroll. Re-armed by the main-menu
// update handler and discarded silently on any other screen so that
// leaving the main menu naturally drains the tick loop within one
// interval. No background goroutine, no shutdown coordination — the
// tea.Tick command model handles it for us.
type newsTickerMsg struct{}

// tickerInterval controls scroll speed. 150ms gives a readable pace
// (about 6.6 cells/sec) that matches real-world news tickers without
// hammering the renderer. Changing this is a UX knob, not a perf one.
const tickerInterval = 150 * time.Millisecond

// tickerFeedRebuildEvery forces a feed rebuild from the news.Manager
// every N ticks so new/expired articles eventually make it into the
// scroll without stalling. 40 × 150ms = 6 seconds — short enough for
// breaking news to surface, long enough that the manager isn't queried
// on every frame.
const tickerFeedRebuildEvery = 40

// tickerSeparator joins individual headlines. Trailing copy is
// appended to the feed in buildTickerFeed so the circular wrap never
// visually smashes the last headline into the first.
const tickerSeparator = "  ◆  "

type newsTickerState struct {
	offset  int
	feed    string
	ticks   int
	lastLen int  // cached feed rune count so windowTickerFeed doesn't recompute
	active  bool // true while a tea.Tick is in-flight; see ensureNewsTickerTick
}

func newNewsTickerState() newsTickerState {
	return newsTickerState{}
}

func newsTickerTick() tea.Cmd {
	return tea.Tick(tickerInterval, func(time.Time) tea.Msg {
		return newsTickerMsg{}
	})
}

// buildTickerFeed joins article headlines into a single scrolling
// line. Articles with NewsPriorityCritical are prefixed "[BREAKING] "
// and pulled to the front; NewsPriorityHigh is prefixed "[!] " and
// placed after the critical block; the rest run in input order.
//
// A trailing separator is appended so the circular window in
// windowTickerFeed rejoins the head without visually smashing the
// last headline into the first.
func buildTickerFeed(articles []*models.NewsArticle) string {
	if len(articles) == 0 {
		return ""
	}

	var critical, high, rest []string
	for _, a := range articles {
		if a == nil || strings.TrimSpace(a.Headline) == "" {
			continue
		}
		switch a.Priority {
		case models.NewsPriorityCritical:
			critical = append(critical, "[BREAKING] "+a.Headline)
		case models.NewsPriorityHigh:
			high = append(high, "[!] "+a.Headline)
		default:
			rest = append(rest, a.Headline)
		}
	}

	ordered := make([]string, 0, len(critical)+len(high)+len(rest))
	ordered = append(ordered, critical...)
	ordered = append(ordered, high...)
	ordered = append(ordered, rest...)
	if len(ordered) == 0 {
		return ""
	}
	return strings.Join(ordered, tickerSeparator) + tickerSeparator
}

// windowTickerFeed extracts a width-rune circular window from feed
// starting at offset. Offset is normalized modulo the feed's rune
// length; Go's `%` can return negative values for negative dividends,
// so we add-and-re-modulo to force a non-negative result.
//
// Width is in runes, not cells. The ticker feed is built from ASCII
// headlines + the ◆ separator (2 cells), so rune-based indexing is
// close enough visually — emoji that sneak in via a headline body
// will cause one cell of shimmer as they enter/leave the window,
// which is the same behavior every other terminal ticker has.
func windowTickerFeed(feed string, offset, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(feed)
	n := len(runes)
	if n == 0 {
		return strings.Repeat(" ", width)
	}
	offset = ((offset % n) + n) % n
	out := make([]rune, width)
	for i := 0; i < width; i++ {
		out[i] = runes[(offset+i)%n]
	}
	return string(out)
}

// updateNewsTicker handles a ticker tick. Rebuilds the feed every
// `tickerFeedRebuildEvery` ticks so new articles surface without
// polling on every frame. Returns the re-armed tick command so the
// scroll continues — callers are expected to invoke this only while
// the main menu is active.
func (m Model) updateNewsTicker() (Model, tea.Cmd) {
	st := &m.newsTicker
	if st.ticks%tickerFeedRebuildEvery == 0 {
		var articles []*models.NewsArticle
		if m.newsManager != nil {
			articles = m.newsManager.GetRecentArticles(20, "")
		}
		st.feed = buildTickerFeed(articles)
		st.lastLen = len([]rune(st.feed))
	}
	if st.lastLen > 0 {
		st.offset = (st.offset + 1) % st.lastLen
	}
	st.ticks++
	st.active = true
	return m, newsTickerTick()
}

// ensureNewsTickerTick kicks off a fresh tick if none is in-flight.
// Called from the main-menu updater on non-tick messages so the ticker
// self-starts on every (re-)entry to the screen without having to
// instrument the ~40 places that set `m.screen = ScreenMainMenu`.
// When the user navigates away, the next tick arrives at a non-main-
// menu updater, drops, and the handler at that entry point marks
// active=false for us — but we also proactively reset it here if a
// stale active=true ever survives (e.g. from a crash-recovery path).
func (m Model) ensureNewsTickerTick() (Model, tea.Cmd) {
	if m.newsTicker.active {
		return m, nil
	}
	m.newsTicker.active = true
	return m, newsTickerTick()
}

// stopNewsTicker drops the active flag so the next main-menu entry
// restarts a tick. Called from the top-level Update when a
// newsTickerMsg arrives on a non-main-menu screen.
func (m Model) stopNewsTicker() Model {
	m.newsTicker.active = false
	return m
}

// renderNewsTicker returns the one-line rendered ticker strip, or an
// empty string if the news manager has no content. The caller is
// responsible for framing/padding (the main menu wraps this in its
// box borders via writeFramedLine).
//
// width is the interior cell budget. We deduct a leading "▸ " icon
// (2 cells) from the scroll window so it visually reads as a
// ticker with a fixed indicator rather than text that happens to
// scroll. Returns "" (not padding) when there is no news so the
// caller can decide whether to suppress the entire row.
func (m Model) renderNewsTicker(width int) string {
	if width < 10 {
		return ""
	}
	feed := m.newsTicker.feed
	if strings.TrimSpace(feed) == "" {
		return ""
	}
	prefix := tickerPrefixStyle.Render("▸ ")
	// 2 cells consumed by "▸ " prefix.
	window := windowTickerFeed(feed, m.newsTicker.offset, width-2)
	return prefix + tickerBodyStyle.Render(window)
}

// Local styles without MarginTop so .Render() doesn't inject a
// leading newline into the single-line ticker strip. See the same
// fix on the leaderboard tab bar.
var (
	tickerPrefixStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")). // Orange
				Bold(true)

	tickerBodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray
)
