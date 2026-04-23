// File: internal/tui/input_keys_test.go
// Project: Terminal Velocity
// Description: Tests for printableRuneString
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-22

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPrintableRuneStringSingleLetter(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	s, ok := printableRuneString(msg)
	if !ok || s != "a" {
		t.Errorf("got (%q, %v), want (%q, %v)", s, ok, "a", true)
	}
}

func TestPrintableRuneStringPaste(t *testing.T) {
	// Bracketed paste arrives as a single KeyMsg with multi-rune payload.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("TestPass123!")}
	s, ok := printableRuneString(msg)
	if !ok || s != "TestPass123!" {
		t.Errorf("paste lost: got (%q, %v)", s, ok)
	}
}

func TestPrintableRuneStringRejectsNamedKeys(t *testing.T) {
	// Named keys have empty Runes. String() returns the name ("enter", "up", ...).
	for _, k := range []tea.KeyType{tea.KeyEnter, tea.KeyTab, tea.KeyUp, tea.KeyEsc, tea.KeyF1, tea.KeyBackspace} {
		msg := tea.KeyMsg{Type: k}
		if s, ok := printableRuneString(msg); ok {
			t.Errorf("named key %v produced text %q; should have been rejected", k, s)
		}
	}
}

func TestPrintableRuneStringStripsEmbeddedControl(t *testing.T) {
	// Paste buffers sometimes contain embedded \r / \n. We strip those to keep
	// multi-line pastes from landing in a single-line text field.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello\rworld")}
	s, ok := printableRuneString(msg)
	if !ok {
		t.Fatalf("expected ok")
	}
	if s != "helloworld" {
		t.Errorf("got %q, want %q", s, "helloworld")
	}
}

func TestPrintableRuneStringRejectsAllControl(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\r', '\n', 0x1b}}
	if _, ok := printableRuneString(msg); ok {
		t.Errorf("should reject all-control payload")
	}
}

func TestPrintableRuneStringUnicode(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("héllo")}
	s, ok := printableRuneString(msg)
	if !ok || s != "héllo" {
		t.Errorf("got (%q, %v), want (%q, %v)", s, ok, "héllo", true)
	}
}
