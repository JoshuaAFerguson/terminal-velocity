// File: internal/metrics/enhanced.go
// Project: Terminal Velocity
// Description: Enhanced metrics with histograms and detailed tracking
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// LatencyHistogram tracks latency distribution for operations
type LatencyHistogram struct {
	mu      sync.RWMutex
	buckets map[string][]time.Duration // Operation name -> durations
	limits  int                         // Max samples per bucket
}

// NewLatencyHistogram creates a new histogram
func NewLatencyHistogram(sampleLimit int) *LatencyHistogram {
	return &LatencyHistogram{
		buckets: make(map[string][]time.Duration),
		limits:  sampleLimit,
	}
}

// Record adds a latency sample for an operation
func (h *LatencyHistogram) Record(operation string, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.buckets[operation] == nil {
		h.buckets[operation] = make([]time.Duration, 0, h.limits)
	}

	h.buckets[operation] = append(h.buckets[operation], duration)

	// Keep only recent samples
	if len(h.buckets[operation]) > h.limits {
		h.buckets[operation] = h.buckets[operation][len(h.buckets[operation])-h.limits:]
	}
}

// GetPercentiles returns p50, p95, p99 for an operation
func (h *LatencyHistogram) GetPercentiles(operation string) (p50, p95, p99 time.Duration) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	durations := h.buckets[operation]
	if len(durations) == 0 {
		return 0, 0, 0
	}

	// Make a copy and sort
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	p50 = sorted[len(sorted)*50/100]
	p95 = sorted[len(sorted)*95/100]
	p99 = sorted[len(sorted)*99/100]

	return p50, p95, p99
}

// GetAverage returns the average latency for an operation
func (h *LatencyHistogram) GetAverage(operation string) time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	durations := h.buckets[operation]
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// GetOperations returns all tracked operations
func (h *LatencyHistogram) GetOperations() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ops := make([]string, 0, len(h.buckets))
	for op := range h.buckets {
		ops = append(ops, op)
	}
	return ops
}

// ErrorCounter tracks errors by category
type ErrorCounter struct {
	mu         sync.RWMutex
	categories map[string]*atomic.Int64
	recent     []ErrorRecord
	maxRecent  int
}

// ErrorRecord tracks a single error occurrence
type ErrorRecord struct {
	Category  string
	Message   string
	Timestamp time.Time
}

// NewErrorCounter creates a new error counter
func NewErrorCounter(maxRecent int) *ErrorCounter {
	return &ErrorCounter{
		categories: make(map[string]*atomic.Int64),
		recent:     make([]ErrorRecord, 0, maxRecent),
		maxRecent:  maxRecent,
	}
}

// RecordError records an error by category
func (ec *ErrorCounter) RecordError(category, message string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Increment category counter
	if ec.categories[category] == nil {
		ec.categories[category] = &atomic.Int64{}
	}
	ec.categories[category].Add(1)

	// Add to recent errors
	ec.recent = append(ec.recent, ErrorRecord{
		Category:  category,
		Message:   message,
		Timestamp: time.Now(),
	})

	// Trim if needed
	if len(ec.recent) > ec.maxRecent {
		ec.recent = ec.recent[len(ec.recent)-ec.maxRecent:]
	}
}

// GetCount returns the count for a category
func (ec *ErrorCounter) GetCount(category string) int64 {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if counter := ec.categories[category]; counter != nil {
		return counter.Load()
	}
	return 0
}

// GetAllCategories returns all error categories and their counts
func (ec *ErrorCounter) GetAllCategories() map[string]int64 {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	result := make(map[string]int64)
	for category, counter := range ec.categories {
		result[category] = counter.Load()
	}
	return result
}

// GetRecentErrors returns the most recent errors
func (ec *ErrorCounter) GetRecentErrors(limit int) []ErrorRecord {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if limit > len(ec.recent) {
		limit = len(ec.recent)
	}

	// Return last N errors
	start := len(ec.recent) - limit
	if start < 0 {
		start = 0
	}

	result := make([]ErrorRecord, limit)
	copy(result, ec.recent[start:])
	return result
}

// EnhancedMetrics adds advanced metric tracking
type EnhancedMetrics struct {
	// Latency tracking
	OperationLatency *LatencyHistogram

	// Error tracking
	Errors *ErrorCounter

	// Activity rates (per minute)
	TradeRate    *RateCounter
	CombatRate   *RateCounter
	LoginRate    *RateCounter
	CommandRate  *RateCounter

	// Resource usage
	ActiveSessions   atomic.Int64
	MessageQueueSize atomic.Int64
	CacheSize        atomic.Int64
}

// RateCounter tracks events per time window
type RateCounter struct {
	mu      sync.RWMutex
	events  []time.Time
	window  time.Duration
}

// NewRateCounter creates a new rate counter
func NewRateCounter(window time.Duration) *RateCounter {
	return &RateCounter{
		events: make([]time.Time, 0),
		window: window,
	}
}

// Record records an event
func (rc *RateCounter) Record() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	rc.events = append(rc.events, now)

	// Remove events outside the window
	cutoff := now.Add(-rc.window)
	for i, t := range rc.events {
		if t.After(cutoff) {
			rc.events = rc.events[i:]
			return
		}
	}
	rc.events = rc.events[:0]
}

// GetRate returns events per minute
func (rc *RateCounter) GetRate() float64 {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	// Count events in the last minute
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	count := 0
	for _, t := range rc.events {
		if t.After(cutoff) {
			count++
		}
	}

	return float64(count)
}

// Global enhanced metrics
var enhanced *EnhancedMetrics
var enhancedOnce sync.Once

// InitEnhanced initializes enhanced metrics
func InitEnhanced() *EnhancedMetrics {
	enhancedOnce.Do(func() {
		enhanced = &EnhancedMetrics{
			OperationLatency: NewLatencyHistogram(1000),
			Errors:           NewErrorCounter(100),
			TradeRate:        NewRateCounter(time.Hour),
			CombatRate:       NewRateCounter(time.Hour),
			LoginRate:        NewRateCounter(time.Hour),
			CommandRate:      NewRateCounter(time.Hour),
		}
	})
	return enhanced
}

// GetEnhanced returns the enhanced metrics instance
func GetEnhanced() *EnhancedMetrics {
	if enhanced == nil {
		return InitEnhanced()
	}
	return enhanced
}

// Operation timing helper
type Timer struct {
	start     time.Time
	operation string
}

// StartTimer starts timing an operation
func StartTimer(operation string) *Timer {
	return &Timer{
		start:     time.Now(),
		operation: operation,
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() {
	duration := time.Since(t.start)
	GetEnhanced().OperationLatency.Record(t.operation, duration)
}

// Convenience functions for common operations
func RecordDatabaseQuery(duration time.Duration) {
	GetEnhanced().OperationLatency.Record("database_query", duration)
}

func RecordTradeOperation(duration time.Duration) {
	GetEnhanced().OperationLatency.Record("trade", duration)
	GetEnhanced().TradeRate.Record()
}

func RecordCombatOperation(duration time.Duration) {
	GetEnhanced().OperationLatency.Record("combat", duration)
	GetEnhanced().CombatRate.Record()
}

func RecordLoginOperation(duration time.Duration) {
	GetEnhanced().OperationLatency.Record("login", duration)
	GetEnhanced().LoginRate.Record()
}

func RecordError(category, message string) {
	GetEnhanced().Errors.RecordError(category, message)
}

// GetOperationStats returns formatted statistics for an operation
type OperationStats struct {
	Average time.Duration
	P50     time.Duration
	P95     time.Duration
	P99     time.Duration
}

func GetOperationStats(operation string) OperationStats {
	h := GetEnhanced().OperationLatency
	avg := h.GetAverage(operation)
	p50, p95, p99 := h.GetPercentiles(operation)

	return OperationStats{
		Average: avg,
		P50:     p50,
		P95:     p95,
		P99:     p99,
	}
}
