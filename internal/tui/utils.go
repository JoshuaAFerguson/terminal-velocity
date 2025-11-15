// File: internal/tui/utils.go
// Project: Terminal Velocity
// Description: Shared utility functions for TUI screens
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"fmt"
	"strings"
	"time"
)

// truncate shortens a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a duration into a human-readable string
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

// wrapText wraps text to a specified width, breaking on word boundaries
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
