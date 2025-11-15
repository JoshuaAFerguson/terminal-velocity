// File: internal/metrics/enhanced_test.go
// Project: Terminal Velocity
// Description: Tests for enhanced metrics system
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package metrics

import (
	"sync"
	"testing"
	"time"
)

// TestLatencyHistogram tests the latency histogram functionality
func TestLatencyHistogram(t *testing.T) {
	t.Run("Record and retrieve latencies", func(t *testing.T) {
		h := NewLatencyHistogram(100)

		// Record some sample latencies
		h.Record("test_op", 10*time.Millisecond)
		h.Record("test_op", 20*time.Millisecond)
		h.Record("test_op", 30*time.Millisecond)
		h.Record("test_op", 40*time.Millisecond)
		h.Record("test_op", 50*time.Millisecond)

		p50, _, _ := h.GetPercentiles("test_op")

		if p50 != 30*time.Millisecond {
			t.Errorf("Expected p50 to be 30ms, got %v", p50)
		}
	})

	t.Run("Empty operation returns zeros", func(t *testing.T) {
		h := NewLatencyHistogram(100)

		p50, p95, p99 := h.GetPercentiles("nonexistent")

		if p50 != 0 || p95 != 0 || p99 != 0 {
			t.Errorf("Expected zeros for non-existent operation, got p50=%v, p95=%v, p99=%v", p50, p95, p99)
		}
	})

	t.Run("Sample limit enforcement", func(t *testing.T) {
		h := NewLatencyHistogram(5) // Limit to 5 samples

		// Record 10 samples
		for i := 0; i < 10; i++ {
			h.Record("test_op", time.Duration(i)*time.Millisecond)
		}

		// Get all operations to verify internal state
		ops := h.GetOperations()
		if len(ops) != 1 {
			t.Fatalf("Expected 1 operation, got %d", len(ops))
		}

		// Verify that only the last 5 samples are kept
		// (we can't directly check the internal state, but percentiles should reflect recent samples)
		p50, _, _ := h.GetPercentiles("test_op")
		if p50 < 5*time.Millisecond {
			t.Errorf("Expected p50 to reflect recent samples (>= 5ms), got %v", p50)
		}
	})

	t.Run("Average calculation", func(t *testing.T) {
		h := NewLatencyHistogram(100)

		h.Record("test_op", 10*time.Millisecond)
		h.Record("test_op", 20*time.Millisecond)
		h.Record("test_op", 30*time.Millisecond)

		avg := h.GetAverage("test_op")
		expected := 20 * time.Millisecond

		if avg != expected {
			t.Errorf("Expected average of 20ms, got %v", avg)
		}
	})

	t.Run("Multiple operations tracked separately", func(t *testing.T) {
		h := NewLatencyHistogram(100)

		h.Record("op_a", 10*time.Millisecond)
		h.Record("op_b", 20*time.Millisecond)
		h.Record("op_a", 15*time.Millisecond)

		avgA := h.GetAverage("op_a")
		avgB := h.GetAverage("op_b")

		if avgA != 12*time.Millisecond+500*time.Microsecond {
			t.Errorf("Expected op_a average of 12.5ms, got %v", avgA)
		}
		if avgB != 20*time.Millisecond {
			t.Errorf("Expected op_b average of 20ms, got %v", avgB)
		}

		ops := h.GetOperations()
		if len(ops) != 2 {
			t.Errorf("Expected 2 operations, got %d", len(ops))
		}
	})
}

// TestLatencyHistogramConcurrency tests thread safety
func TestLatencyHistogramConcurrency(t *testing.T) {
	h := NewLatencyHistogram(1000)

	const numGoroutines = 50
	const recordsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines recording latencies concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < recordsPerGoroutine; j++ {
				h.Record("concurrent_op", time.Duration(j)*time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify we can safely read percentiles after concurrent writes
	p50, p95, p99 := h.GetPercentiles("concurrent_op")

	if p50 == 0 || p95 == 0 || p99 == 0 {
		t.Error("Expected non-zero percentiles after concurrent writes")
	}
}

// TestErrorCounter tests the error counter functionality
func TestErrorCounter(t *testing.T) {
	t.Run("Record and retrieve error counts", func(t *testing.T) {
		ec := NewErrorCounter(100)

		ec.RecordError("database", "connection timeout")
		ec.RecordError("database", "query failed")
		ec.RecordError("network", "connection refused")

		dbCount := ec.GetCount("database")
		netCount := ec.GetCount("network")

		if dbCount != 2 {
			t.Errorf("Expected 2 database errors, got %d", dbCount)
		}
		if netCount != 1 {
			t.Errorf("Expected 1 network error, got %d", netCount)
		}
	})

	t.Run("Non-existent category returns zero", func(t *testing.T) {
		ec := NewErrorCounter(100)

		count := ec.GetCount("nonexistent")

		if count != 0 {
			t.Errorf("Expected 0 for non-existent category, got %d", count)
		}
	})

	t.Run("Recent errors tracking", func(t *testing.T) {
		ec := NewErrorCounter(10) // Keep last 10 errors

		ec.RecordError("test", "error 1")
		ec.RecordError("test", "error 2")
		ec.RecordError("test", "error 3")

		recent := ec.GetRecentErrors(5)

		if len(recent) != 3 {
			t.Errorf("Expected 3 recent errors, got %d", len(recent))
		}

		// Verify errors are in chronological order (oldest to newest)
		if recent[0].Message != "error 1" {
			t.Errorf("Expected first error to be 'error 1', got '%s'", recent[0].Message)
		}
	})

	t.Run("Recent errors limit enforcement", func(t *testing.T) {
		ec := NewErrorCounter(5) // Keep only last 5 errors

		// Record 10 errors
		for i := 0; i < 10; i++ {
			ec.RecordError("test", "error")
		}

		recent := ec.GetRecentErrors(100) // Request more than available

		if len(recent) > 5 {
			t.Errorf("Expected at most 5 recent errors, got %d", len(recent))
		}
	})

	t.Run("All categories retrieval", func(t *testing.T) {
		ec := NewErrorCounter(100)

		ec.RecordError("cat_a", "error 1")
		ec.RecordError("cat_b", "error 2")
		ec.RecordError("cat_a", "error 3")

		all := ec.GetAllCategories()

		if len(all) != 2 {
			t.Errorf("Expected 2 categories, got %d", len(all))
		}
		if all["cat_a"] != 2 {
			t.Errorf("Expected 2 errors in cat_a, got %d", all["cat_a"])
		}
		if all["cat_b"] != 1 {
			t.Errorf("Expected 1 error in cat_b, got %d", all["cat_b"])
		}
	})
}

// TestErrorCounterConcurrency tests thread safety
func TestErrorCounterConcurrency(t *testing.T) {
	ec := NewErrorCounter(1000)

	const numGoroutines = 50
	const errorsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines recording errors concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < errorsPerGoroutine; j++ {
				ec.RecordError("concurrent", "test error")
			}
		}(i)
	}

	wg.Wait()

	// Verify final count is correct
	count := ec.GetCount("concurrent")
	expected := int64(numGoroutines * errorsPerGoroutine)

	if count != expected {
		t.Errorf("Expected %d errors, got %d (race condition detected)", expected, count)
	}
}

// TestRateCounter tests the rate counter functionality
func TestRateCounter(t *testing.T) {
	t.Run("Record and retrieve rates", func(t *testing.T) {
		rc := NewRateCounter(time.Minute)

		// Record some events
		for i := 0; i < 10; i++ {
			rc.Record()
		}

		// Get rate
		rate := rc.GetRate()

		if rate == 0 {
			t.Error("Expected non-zero rate after recording events")
		}
	})

	t.Run("Empty counter returns zero", func(t *testing.T) {
		rc := NewRateCounter(time.Minute)

		rate := rc.GetRate()

		if rate != 0 {
			t.Errorf("Expected 0 for empty counter, got %f", rate)
		}
	})

	t.Run("Rate calculation", func(t *testing.T) {
		rc := NewRateCounter(time.Minute)

		// Record some events
		for i := 0; i < 5; i++ {
			rc.Record()
		}

		// Check rate
		rate := rc.GetRate()
		if rate != 5.0 {
			t.Errorf("Expected rate of 5.0, got %f", rate)
		}
	})
}

// TestRateCounterConcurrency tests thread safety
func TestRateCounterConcurrency(t *testing.T) {
	rc := NewRateCounter(time.Minute)

	const numGoroutines = 50
	const eventsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines recording events concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < eventsPerGoroutine; j++ {
				rc.Record()
			}
		}(i)
	}

	wg.Wait()

	// Verify we can safely read rate after concurrent writes
	rate := rc.GetRate()

	if rate == 0 {
		t.Error("Expected non-zero rate after concurrent writes")
	}
}

// TestEnhancedMetrics tests the global enhanced metrics
func TestEnhancedMetrics(t *testing.T) {
	t.Run("Global instance accessible", func(t *testing.T) {
		enhanced := GetEnhanced()

		if enhanced == nil {
			t.Fatal("Expected non-nil enhanced metrics instance")
		}

		if enhanced.OperationLatency == nil {
			t.Error("Expected non-nil OperationLatency")
		}
		if enhanced.Errors == nil {
			t.Error("Expected non-nil Errors")
		}
		if enhanced.TradeRate == nil {
			t.Error("Expected non-nil TradeRate")
		}
	})

	t.Run("Convenience functions work", func(t *testing.T) {
		// These should not panic
		RecordDatabaseQuery(10 * time.Millisecond)
		RecordTradeOperation(20 * time.Millisecond)
		RecordCombatOperation(30 * time.Millisecond)

		enhanced := GetEnhanced()

		// Verify operations were recorded
		ops := enhanced.OperationLatency.GetOperations()
		if len(ops) == 0 {
			t.Error("Expected operations to be recorded by convenience functions")
		}
	})
}
