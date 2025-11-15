// File: internal/tui/input_validation_test.go
// Project: Terminal Velocity
// Description: Regression tests for input validation fixes
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRegistrationInputLengthLimits tests that registration input is properly limited
// Regression test for memory exhaustion bugs (30+ input validation fixes)
func TestRegistrationInputLengthLimits(t *testing.T) {
	tests := []struct {
		name          string
		step          int
		input         string
		expectedLen   int
		maxLen        int
	}{
		{
			name:        "Email length limit (RFC 5321 max)",
			step:        1, // Email step
			input:       strings.Repeat("a", 300),
			expectedLen: 254,
			maxLen:      254,
		},
		{
			name:        "Password length limit",
			step:        2, // Password step
			input:       strings.Repeat("b", 200),
			expectedLen: 128,
			maxLen:      128,
		},
		{
			name:        "Confirm password length limit",
			step:        3, // Confirm password step
			input:       strings.Repeat("c", 200),
			expectedLen: 128,
			maxLen:      128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				registration: newRegistrationModel(false, nil),
			}
			m.registration.step = tt.step

			// Simulate typing each character
			for _, char := range tt.input {
				m.handleRegistrationInput(string(char))
			}

			var actual string
			switch tt.step {
			case 1:
				actual = m.registration.email
			case 2:
				actual = m.registration.password
			case 3:
				actual = m.registration.confirmPass
			}

			if len(actual) > tt.maxLen {
				t.Errorf("Input exceeded max length: got %d, want max %d", len(actual), tt.maxLen)
			}

			if len(actual) != tt.expectedLen {
				t.Errorf("Expected length %d, got %d", tt.expectedLen, len(actual))
			}
		})
	}
}

// TestRegistrationControlCharacterFiltering tests that control characters are filtered
// Regression test for ANSI escape code injection
func TestRegistrationControlCharacterFiltering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Filter null byte",
			input:    "test\x00user",
			expected: "testuser",
		},
		{
			name:     "Filter backspace",
			input:    "test\buser",
			expected: "testuser",
		},
		{
			name:     "Filter escape sequences",
			input:    "test\x1b[31mred\x1b[0m",
			expected: "testred",
		},
		{
			name:     "Filter DEL character",
			input:    "test\x7fuser",
			expected: "testuser",
		},
		{
			name:     "Allow normal characters",
			input:    "testuser123@example.com",
			expected: "testuser123@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				registration: newRegistrationModel(false, nil),
			}
			m.registration.step = 1 // Email step

			// Simulate typing each character
			for _, char := range tt.input {
				m.handleRegistrationInput(string(char))
			}

			if m.registration.email != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, m.registration.email)
			}
		})
	}
}

// TestChatInputSanitization tests that chat input is sanitized
// Regression test for chat input validation fixes
func TestChatInputSanitization(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLen   int
		shouldContain bool
	}{
		{
			name:          "Normal message",
			input:         "Hello, world!",
			expectedLen:   13,
			shouldContain: true,
		},
		{
			name:          "Message with control characters",
			input:         "Hello\x00\x1b[31m world",
			expectedLen:   11, // Control chars filtered
			shouldContain: true,
		},
		{
			name:        "Message at length limit",
			input:       strings.Repeat("a", 200),
			expectedLen: 200,
			shouldContain: true,
		},
		{
			name:        "Message exceeding length limit",
			input:       strings.Repeat("b", 250),
			expectedLen: 200, // Capped at limit
			shouldContain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				chatModel: newChatModel(),
			}
			m.chatModel.inputMode = true

			// Simulate typing each character
			for _, char := range tt.input {
				updatedModel, _ := m.updateChat(keyMsg(string(char)))
				m = updatedModel.(Model)
			}

			if len(m.chatModel.inputBuffer) > 200 {
				t.Errorf("Chat buffer exceeded limit: got %d chars", len(m.chatModel.inputBuffer))
			}

			if len(m.chatModel.inputBuffer) != tt.expectedLen {
				t.Errorf("Expected buffer length %d, got %d", tt.expectedLen, len(m.chatModel.inputBuffer))
			}
		})
	}
}

// TestHelpScreenArrayBounds tests that help screen cursor doesn't cause negative index
// Regression test for array bounds bug in help.go
func TestHelpScreenArrayBounds(t *testing.T) {
	m := Model{
		helpModel: newHelpModel(),
	}

	// Set cursor to a position that would cause negative index if not fixed
	m.helpModel.cursor = 0

	// Try to select an item (should not panic)
	msg := keyMsg("enter")
	_, cmd := m.updateHelp(msg)

	if cmd != nil {
		// Command was created, which is fine
	}

	// Verify we didn't access negative index
	if m.helpModel.currentTopic != nil {
		t.Error("Should not have selected a topic with cursor at 0")
	}

	// Test with cursor = 1
	m.helpModel.cursor = 1
	_, _ = m.updateHelp(msg)

	// Should still not select a topic (these are special options)
	if m.helpModel.currentTopic != nil {
		t.Error("Should not have selected a topic with cursor at 1")
	}

	// Test with valid cursor for topic selection
	m.helpModel.cursor = 2
	if len(m.helpModel.topics) > 0 {
		_, _ = m.updateHelp(msg)
		// This should work without panicking
	}
}

// Helper function to create a key message
func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
