// File: internal/tui/utils.go
// Project: Terminal Velocity
// Description: Shared utility functions for text formatting and display in TUI screens
// Version: 1.1.0
// Author: Claude Code
// Created: 2025-11-15
//
// This file provides common utility functions used across multiple TUI screens.
// These functions handle text formatting, duration formatting, and word wrapping.
//
// Usage:
//   - Import in screen files: truncate(longString, 50)
//   - Used for consistent text display across all screens
//   - Functions are pure (no side effects) and thread-safe

package tui

import (
	"fmt"
	"strings"
	"time"
)

// truncate shortens a string to maxLen characters, adding "..." if truncated.
//
// This is used throughout the TUI to ensure text fits within column widths
// in tables and lists. If the string is shorter than maxLen, it is returned
// unchanged.
//
// Parameters:
//   - s: The string to truncate
//   - maxLen: Maximum length of the returned string
//
// Returns:
//   - Original string if len(s) <= maxLen
//   - Truncated string with "..." appended if len(s) > maxLen
//
// Example:
//   truncate("Hello, World!", 8)  // Returns: "Hello..."
//   truncate("Short", 10)          // Returns: "Short"
//
// Note: The "..." counts toward maxLen, so the actual text will be maxLen-3.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a time.Duration into a human-readable string.
//
// This is used to display time-based information like mission durations,
// event countdowns, ban expiration times, etc.
//
// Format rules:
//   - Negative durations: "Expired"
//   - >= 48 hours: "Xd" (days)
//   - < 48 hours but >= 1 hour: "Xh" (hours)
//   - < 1 hour: "Xm" (minutes)
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - Human-readable duration string
//
// Examples:
//   formatDuration(5 * time.Minute)       // Returns: "5m"
//   formatDuration(2 * time.Hour)         // Returns: "2h"
//   formatDuration(3 * 24 * time.Hour)    // Returns: "3d"
//   formatDuration(-1 * time.Hour)        // Returns: "Expired"
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Expired"
	}

	hours := int(d.Hours())
	if hours > 48 {
		days := hours / 24
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	minutes := int(d.Minutes())
	return fmt.Sprintf("%dm", minutes)
}

// wrapText wraps text to a specified width, breaking on word boundaries.
//
// This is used to format long text (quest descriptions, news articles, etc.)
// to fit within terminal width constraints. Text is broken at word boundaries
// to avoid splitting words mid-way.
//
// Parameters:
//   - text: The text to wrap
//   - width: Maximum width of each line
//
// Returns:
//   - Slice of strings, each representing one line
//   - Empty slice if text is empty or contains no words
//
// Algorithm:
//   1. Split text into words (by whitespace)
//   2. Build lines by adding words until width would be exceeded
//   3. Start new line when adding next word would exceed width
//   4. Words longer than width are not split (will exceed width)
//
// Example:
//   text := "This is a long sentence that needs wrapping"
//   wrapped := wrapText(text, 20)
//   // Returns: ["This is a long", "sentence that needs", "wrapping"]
//
// Note: This is a simple word-based wrapper. It doesn't handle:
//   - Hyphenation
//   - Preserving multiple spaces
//   - Special characters or formatting
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
