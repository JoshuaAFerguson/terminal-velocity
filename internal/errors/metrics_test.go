// File: internal/errors/metrics_test.go
// Project: Terminal Velocity
// Description: Tests for error metrics
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package errors

import (
	"errors"
	"testing"
	"time"
)

func TestMetrics_RecordError(t *testing.T) {
	m := NewMetrics()

	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")
	err3 := errors.New("test error 3")

	m.RecordError("database", "connection", err1)
	m.RecordError("database", "query", err2)
	m.RecordError("server", "connection", err3)

	stats := m.GetStats()

	if stats.TotalErrors != 3 {
		t.Errorf("Expected TotalErrors=3, got %d", stats.TotalErrors)
	}

	if stats.ErrorsByType["connection"] != 2 {
		t.Errorf("Expected ErrorsByType[connection]=2, got %d", stats.ErrorsByType["connection"])
	}

	if stats.ErrorsByType["query"] != 1 {
		t.Errorf("Expected ErrorsByType[query]=1, got %d", stats.ErrorsByType["query"])
	}

	if stats.ErrorsBySource["database"] != 2 {
		t.Errorf("Expected ErrorsBySource[database]=2, got %d", stats.ErrorsBySource["database"])
	}

	if stats.ErrorsBySource["server"] != 1 {
		t.Errorf("Expected ErrorsBySource[server]=1, got %d", stats.ErrorsBySource["server"])
	}

	// Last error should be the most recent one
	if stats.LastErrorMsg != "test error 3" {
		t.Errorf("Expected LastErrorMsg='test error 3', got '%s'", stats.LastErrorMsg)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics()

	err1 := errors.New("test error")
	m.RecordError("database", "connection", err1)

	stats := m.GetStats()
	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors=1, got %d", stats.TotalErrors)
	}

	m.Reset()

	stats = m.GetStats()
	if stats.TotalErrors != 0 {
		t.Errorf("Expected TotalErrors=0 after reset, got %d", stats.TotalErrors)
	}

	if len(stats.ErrorsByType) != 0 {
		t.Errorf("Expected ErrorsByType to be empty after reset, got %d items", len(stats.ErrorsByType))
	}

	if len(stats.ErrorsBySource) != 0 {
		t.Errorf("Expected ErrorsBySource to be empty after reset, got %d items", len(stats.ErrorsBySource))
	}

	if !stats.LastError.IsZero() {
		t.Errorf("Expected LastError to be zero after reset")
	}

	if stats.LastErrorMsg != "" {
		t.Errorf("Expected LastErrorMsg to be empty after reset, got '%s'", stats.LastErrorMsg)
	}
}

func TestMetrics_ErrorRate(t *testing.T) {
	m := NewMetrics()

	// Record some errors
	for i := 0; i < 10; i++ {
		m.RecordError("test", "test", errors.New("test"))
	}

	// Wait a bit to get a measurable rate
	time.Sleep(100 * time.Millisecond)

	stats := m.GetStats()

	// Error rate should be > 0
	if stats.ErrorRate <= 0 {
		t.Errorf("Expected ErrorRate > 0, got %f", stats.ErrorRate)
	}

	// Should have recorded 10 errors
	if stats.TotalErrors != 10 {
		t.Errorf("Expected TotalErrors=10, got %d", stats.TotalErrors)
	}
}

func TestGlobalMetrics(t *testing.T) {
	// Reset global metrics first
	ResetGlobalMetrics()

	err1 := errors.New("test error")
	RecordGlobalError("database", "connection", err1)
	RecordGlobalError("server", "authentication", err1)

	stats := GetGlobalStats()

	if stats.TotalErrors != 2 {
		t.Errorf("Expected TotalErrors=2, got %d", stats.TotalErrors)
	}

	if stats.ErrorsBySource["database"] != 1 {
		t.Errorf("Expected ErrorsBySource[database]=1, got %d", stats.ErrorsBySource["database"])
	}

	if stats.ErrorsBySource["server"] != 1 {
		t.Errorf("Expected ErrorsBySource[server]=1, got %d", stats.ErrorsBySource["server"])
	}

	// Clean up
	ResetGlobalMetrics()
}

func TestMetrics_ConcurrentRecording(t *testing.T) {
	m := NewMetrics()

	// Record errors concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				m.RecordError("test", "concurrent", errors.New("test"))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := m.GetStats()

	// Should have recorded 1000 errors (10 goroutines * 100 errors each)
	if stats.TotalErrors != 1000 {
		t.Errorf("Expected TotalErrors=1000, got %d", stats.TotalErrors)
	}
}

func TestMetrics_NilError(t *testing.T) {
	m := NewMetrics()

	// Recording with nil error should work but not set LastErrorMsg
	m.RecordError("test", "nil_error", nil)

	stats := m.GetStats()

	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors=1, got %d", stats.TotalErrors)
	}

	if stats.LastErrorMsg != "" {
		t.Errorf("Expected LastErrorMsg to be empty for nil error, got '%s'", stats.LastErrorMsg)
	}
}
