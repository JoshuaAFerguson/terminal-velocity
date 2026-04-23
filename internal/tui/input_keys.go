// File: internal/tui/input_keys.go
// Project: Terminal Velocity
// Description: Shared helpers for extracting user-typed text from tea.KeyMsg
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-22

package tui

import (
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// printableRuneString returns the printable-rune payload of a KeyMsg, or ("", false)
// if the key is a named control key (arrow, function, modifier-combo, etc.).
//
// Why this exists: BubbleTea delivers three shapes of text input as
// tea.KeyMsg:
//
//  1. A single pressed rune — msg.Runes = []rune{'a'}, msg.String() == "a".
//  2. A paste (bracketed paste or just a fast stream of bytes) — msg.Runes
//     contains multiple runes, msg.String() concatenates them.
//  3. A named control key — msg.Runes is empty, msg.String() is "up",
//     "ctrl+a", "f1", "enter", "tab", etc.
//
// The previous "only accept length-1 strings" filter rejected case 2 (silently
// dropping pastes) and misclassified anything tied to case 3.
//
// We use msg.Runes as the source of truth, guarding against control runes via
// unicode.IsPrint so that stray non-printables from a broken paste buffer
// don't contaminate the target field.
// keepDigits returns the ASCII digits from s in order. Used to filter pastes
// into numeric-only input fields without silently dropping the whole paste.
func keepDigits(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			out = append(out, c)
		}
	}
	return string(out)
}

// clampRunes truncates s to at most n runes. Use to respect a field's
// character budget when processing a paste.
func clampRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func printableRuneString(msg tea.KeyMsg) (string, bool) {
	if len(msg.Runes) == 0 {
		return "", false
	}
	out := make([]rune, 0, len(msg.Runes))
	for _, r := range msg.Runes {
		if !unicode.IsPrint(r) {
			// Discard embedded control characters; BubbleTea's paste stream
			// occasionally carries \r or \n inside a single KeyMsg.
			continue
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return "", false
	}
	return string(out), true
}
