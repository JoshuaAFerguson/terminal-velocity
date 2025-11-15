// File: internal/models/chat_test.go
// Project: Terminal Velocity
// Description: Regression tests for chat concurrency fixes
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package models

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

// TestChatHistoryConcurrency tests that ChatHistory is thread-safe
// Regression test for race condition in ChatHistory (15 concurrency fixes)
func TestChatHistoryConcurrency(t *testing.T) {
	playerID := uuid.New()
	history := NewChatHistory(playerID)

	const numGoroutines = 100
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines adding messages concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				msg := &ChatMessage{
					ID:        uuid.New(),
					SenderID:  uuid.New(),
					Channel:   ChatChannelGlobal,
					Content:   "Test message",
					Timestamp: 0,
				}
				history.AddMessage(msg)
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were added (no race conditions)
	if len(history.GlobalChat) != numGoroutines*messagesPerGoroutine {
		t.Errorf("Expected %d messages, got %d (race condition detected)",
			numGoroutines*messagesPerGoroutine, len(history.GlobalChat))
	}
}

// TestChatHistoryGetMessagesConcurrency tests concurrent reads
func TestChatHistoryGetMessagesConcurrency(t *testing.T) {
	playerID := uuid.New()
	history := NewChatHistory(playerID)

	// Add some initial messages
	for i := 0; i < 100; i++ {
		msg := &ChatMessage{
			ID:        uuid.New(),
			SenderID:  uuid.New(),
			Channel:   ChatChannelGlobal,
			Content:   "Test message",
			Timestamp: int64(i),
		}
		history.AddMessage(msg)
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines reading messages concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			// Get messages from different channels
			history.GetMessages(ChatChannelGlobal, 10)
			history.GetMessages(ChatChannelSystem, 10)
			history.GetMessages(ChatChannelFaction, 10)
			history.GetMessages(ChatChannelDirect, 10)
		}()
	}

	wg.Wait()

	// If we got here without panicking, the test passed
}

// TestChatHistoryConcurrentReadWrite tests concurrent reads and writes
func TestChatHistoryConcurrentReadWrite(t *testing.T) {
	playerID := uuid.New()
	history := NewChatHistory(playerID)

	const numReaders = 25
	const numWriters = 25

	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Spawn reader goroutines
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				history.GetMessages(ChatChannelGlobal, 10)
			}
		}()
	}

	// Spawn writer goroutines
	for i := 0; i < numWriters; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				msg := &ChatMessage{
					ID:        uuid.New(),
					SenderID:  uuid.New(),
					Channel:   ChatChannelGlobal,
					Content:   "Test message",
					Timestamp: int64(j),
				}
				history.AddMessage(msg)
			}
		}()
	}

	wg.Wait()

	// Verify we can still read messages after concurrent access
	messages := history.GetMessages(ChatChannelGlobal, 100)
	if len(messages) == 0 {
		t.Error("Expected messages after concurrent writes, got 0")
	}
}

// TestChatHistoryClearChannel tests that clearing is thread-safe
func TestChatHistoryClearChannel(t *testing.T) {
	playerID := uuid.New()
	history := NewChatHistory(playerID)

	// Add messages to multiple channels
	for i := 0; i < 50; i++ {
		for _, channel := range []ChatChannel{
			ChatChannelGlobal,
			ChatChannelSystem,
			ChatChannelFaction,
		} {
			msg := &ChatMessage{
				ID:        uuid.New(),
				SenderID:  uuid.New(),
				Channel:   channel,
				Content:   "Test message",
				Timestamp: int64(i),
			}
			history.AddMessage(msg)
		}
	}

	// Clear global channel while adding to other channels
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		history.ClearChannel(ChatChannelGlobal)
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			msg := &ChatMessage{
				ID:        uuid.New(),
				SenderID:  uuid.New(),
				Channel:   ChatChannelSystem,
				Content:   "Test message",
				Timestamp: int64(i),
			}
			history.AddMessage(msg)
		}
	}()

	wg.Wait()

	// Verify global channel was cleared
	globalMessages := history.GetMessages(ChatChannelGlobal, 100)
	if len(globalMessages) != 0 {
		t.Errorf("Expected 0 global messages after clear, got %d", len(globalMessages))
	}

	// Verify system channel still has messages
	systemMessages := history.GetMessages(ChatChannelSystem, 100)
	if len(systemMessages) == 0 {
		t.Error("Expected system messages to remain after clearing global")
	}
}
