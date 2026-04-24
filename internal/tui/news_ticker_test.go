// File: internal/tui/news_ticker_test.go
// Project: Terminal Velocity
// Description: Unit tests for news ticker pure helpers.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tui

import (
	"strings"
	"testing"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

func article(headline string, priority models.NewsPriority) *models.NewsArticle {
	return &models.NewsArticle{
		Headline: headline,
		Priority: priority,
	}
}

func TestBuildTickerFeed(t *testing.T) {
	t.Run("empty input returns empty string", func(t *testing.T) {
		if got := buildTickerFeed(nil); got != "" {
			t.Fatalf("nil input: got %q, want empty", got)
		}
		if got := buildTickerFeed([]*models.NewsArticle{}); got != "" {
			t.Fatalf("empty slice: got %q, want empty", got)
		}
	})

	t.Run("nil articles and empty headlines are skipped", func(t *testing.T) {
		got := buildTickerFeed([]*models.NewsArticle{
			nil,
			article("", models.NewsPriorityLow),
			article("   ", models.NewsPriorityLow),
			article("real news", models.NewsPriorityLow),
		})
		if !strings.Contains(got, "real news") {
			t.Fatalf("expected feed to contain 'real news', got %q", got)
		}
		// Should not contain any of the empty strings — but separators
		// are inherent to the feed, so we check the full shape.
		expected := "real news" + tickerSeparator
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("critical comes first, then high, then rest", func(t *testing.T) {
		got := buildTickerFeed([]*models.NewsArticle{
			article("low1", models.NewsPriorityLow),
			article("critical1", models.NewsPriorityCritical),
			article("high1", models.NewsPriorityHigh),
			article("medium1", models.NewsPriorityMedium),
			article("critical2", models.NewsPriorityCritical),
		})

		// Ordering: critical1, critical2, high1, low1, medium1
		idxCritical1 := strings.Index(got, "critical1")
		idxCritical2 := strings.Index(got, "critical2")
		idxHigh1 := strings.Index(got, "high1")
		idxLow1 := strings.Index(got, "low1")
		idxMedium1 := strings.Index(got, "medium1")

		if idxCritical1 < 0 || idxHigh1 < 0 || idxLow1 < 0 {
			t.Fatalf("missing entries in feed: %q", got)
		}
		if idxCritical1 >= idxHigh1 {
			t.Fatalf("critical should precede high, got critical1=%d high1=%d", idxCritical1, idxHigh1)
		}
		if idxCritical2 >= idxHigh1 {
			t.Fatalf("critical2 should precede high1, got critical2=%d high1=%d", idxCritical2, idxHigh1)
		}
		if idxHigh1 >= idxLow1 {
			t.Fatalf("high should precede low, got high1=%d low1=%d", idxHigh1, idxLow1)
		}
		// Input ordering of same-priority items is preserved: low1
		// came before medium1 in input, so stays before in output.
		if idxLow1 >= idxMedium1 {
			t.Fatalf("input order should be preserved within same priority: got low1=%d medium1=%d", idxLow1, idxMedium1)
		}
	})

	t.Run("critical headlines are prefixed with [BREAKING]", func(t *testing.T) {
		got := buildTickerFeed([]*models.NewsArticle{
			article("war declared", models.NewsPriorityCritical),
		})
		if !strings.Contains(got, "[BREAKING] war declared") {
			t.Fatalf("expected [BREAKING] prefix, got %q", got)
		}
	})

	t.Run("high headlines are prefixed with [!]", func(t *testing.T) {
		got := buildTickerFeed([]*models.NewsArticle{
			article("pirate sighted", models.NewsPriorityHigh),
		})
		if !strings.Contains(got, "[!] pirate sighted") {
			t.Fatalf("expected [!] prefix, got %q", got)
		}
	})

	t.Run("feed ends with trailing separator for clean circular wrap", func(t *testing.T) {
		got := buildTickerFeed([]*models.NewsArticle{
			article("headline", models.NewsPriorityLow),
		})
		if !strings.HasSuffix(got, tickerSeparator) {
			t.Fatalf("feed should end with %q separator for circular wrap, got %q", tickerSeparator, got)
		}
	})
}

func TestWindowTickerFeed(t *testing.T) {
	t.Run("empty feed returns blank window of requested width", func(t *testing.T) {
		got := windowTickerFeed("", 0, 10)
		if got != strings.Repeat(" ", 10) {
			t.Fatalf("empty feed: expected 10 spaces, got %q", got)
		}
	})

	t.Run("zero width returns empty", func(t *testing.T) {
		if got := windowTickerFeed("hello", 0, 0); got != "" {
			t.Fatalf("width 0: expected empty, got %q", got)
		}
	})

	t.Run("negative width returns empty", func(t *testing.T) {
		if got := windowTickerFeed("hello", 0, -5); got != "" {
			t.Fatalf("negative width: expected empty, got %q", got)
		}
	})

	t.Run("window at offset 0 reads head", func(t *testing.T) {
		got := windowTickerFeed("abcdefgh", 0, 4)
		if got != "abcd" {
			t.Fatalf("offset 0 width 4: expected 'abcd', got %q", got)
		}
	})

	t.Run("window at offset reads from that position", func(t *testing.T) {
		got := windowTickerFeed("abcdefgh", 3, 4)
		if got != "defg" {
			t.Fatalf("offset 3 width 4: expected 'defg', got %q", got)
		}
	})

	t.Run("window past end wraps circularly", func(t *testing.T) {
		got := windowTickerFeed("abcdefgh", 6, 4)
		// 6,7 = g,h ; then wrap to 0,1 = a,b → "ghab"
		if got != "ghab" {
			t.Fatalf("offset 6 width 4 on 8-char feed: expected 'ghab', got %q", got)
		}
	})

	t.Run("offset > feed length is normalized via modulo", func(t *testing.T) {
		got := windowTickerFeed("abcdefgh", 10, 4) // 10 mod 8 = 2 → "cdef"
		if got != "cdef" {
			t.Fatalf("offset 10 on 8-char feed: expected 'cdef', got %q", got)
		}
	})

	t.Run("negative offset is normalized to positive", func(t *testing.T) {
		got := windowTickerFeed("abcdefgh", -1, 4) // -1 mod 8 = 7 → "habc"
		if got != "habc" {
			t.Fatalf("offset -1 on 8-char feed: expected 'habc', got %q", got)
		}
	})

	t.Run("width wider than feed wraps multiple times", func(t *testing.T) {
		got := windowTickerFeed("abc", 0, 8) // abcabcab
		if got != "abcabcab" {
			t.Fatalf("width 8 on 3-char feed: expected 'abcabcab', got %q", got)
		}
	})

	t.Run("unicode characters in feed are preserved", func(t *testing.T) {
		// Separator contains ◆, make sure rune-based indexing keeps it whole.
		got := windowTickerFeed("hi ◆ ho", 0, 7)
		if got != "hi ◆ ho" {
			t.Fatalf("unicode feed: expected 'hi ◆ ho', got %q", got)
		}
	})
}

func TestBuildAndWindowIntegration(t *testing.T) {
	// Feed-then-window round trip: the headline must appear somewhere
	// in a window that's wider than the full feed.
	feed := buildTickerFeed([]*models.NewsArticle{
		article("pirates raid Tau", models.NewsPriorityCritical),
	})
	window := windowTickerFeed(feed, 0, 200)
	if !strings.Contains(window, "[BREAKING] pirates raid Tau") {
		t.Fatalf("integration: window missing breaking prefix + headline, got %q", window)
	}
}

// TestTickerScrollAdvancement proves that successive offsets produce a
// shifted window — i.e. the scroll actually scrolls. Catches regressions
// where someone caches the window output or never updates the offset.
func TestTickerScrollAdvancement(t *testing.T) {
	feed := buildTickerFeed([]*models.NewsArticle{
		article("headline one", models.NewsPriorityMedium),
		article("headline two", models.NewsPriorityMedium),
	})

	w0 := windowTickerFeed(feed, 0, 20)
	w1 := windowTickerFeed(feed, 1, 20)
	w2 := windowTickerFeed(feed, 2, 20)
	if w0 == w1 || w1 == w2 {
		t.Fatalf("consecutive offsets should shift the window; got w0=%q w1=%q w2=%q", w0, w1, w2)
	}
}

// TestTickerPreview dumps a few sample frames so a reviewer running
// `go test -v` can eyeball how the ticker looks.
func TestTickerPreview(t *testing.T) {
	feed := buildTickerFeed([]*models.NewsArticle{
		article("Pirate fleet sighted in Outer Reach", models.NewsPriorityCritical),
		article("Silicon futures up 22% in Sol core", models.NewsPriorityMedium),
		article("Tau Station extends docking hours", models.NewsPriorityLow),
	})
	for _, off := range []int{0, 10, 30, 60} {
		t.Logf("offset=%d  |%s|", off, windowTickerFeed(feed, off, 76))
	}
}
