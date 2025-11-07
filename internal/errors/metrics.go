// File: internal/errors/metrics.go
// Project: Terminal Velocity
// Description: Error metrics and monitoring
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package errors

import (
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var metricsLog = logger.WithComponent("ErrorMetrics")

// Metrics tracks error statistics
type Metrics struct {
	mu             sync.RWMutex
	TotalErrors    int64
	ErrorsByType   map[string]int64
	ErrorsBySource map[string]int64
	LastError      time.Time
	LastErrorMsg   string
	startTime      time.Time
}

// NewMetrics creates a new error metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		ErrorsByType:   make(map[string]int64),
		ErrorsBySource: make(map[string]int64),
		startTime:      time.Now(),
	}
}

// RecordError records an error occurrence
func (m *Metrics) RecordError(source, errType string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalErrors++
	m.ErrorsByType[errType]++
	m.ErrorsBySource[source]++
	m.LastError = time.Now()
	if err != nil {
		m.LastErrorMsg = err.Error()
	}

	metricsLog.Debug("Error recorded: source=%s, type=%s, total=%d", source, errType, m.TotalErrors)
}

// GetStats returns current error statistics
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorsByType := make(map[string]int64)
	for k, v := range m.ErrorsByType {
		errorsByType[k] = v
	}

	errorsBySource := make(map[string]int64)
	for k, v := range m.ErrorsBySource {
		errorsBySource[k] = v
	}

	return Stats{
		TotalErrors:    m.TotalErrors,
		ErrorsByType:   errorsByType,
		ErrorsBySource: errorsBySource,
		LastError:      m.LastError,
		LastErrorMsg:   m.LastErrorMsg,
		Uptime:         time.Since(m.startTime),
		ErrorRate:      m.calculateErrorRate(),
	}
}

// Reset clears all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalErrors = 0
	m.ErrorsByType = make(map[string]int64)
	m.ErrorsBySource = make(map[string]int64)
	m.LastError = time.Time{}
	m.LastErrorMsg = ""
	m.startTime = time.Now()

	metricsLog.Info("Error metrics reset")
}

// calculateErrorRate returns errors per minute
func (m *Metrics) calculateErrorRate() float64 {
	uptime := time.Since(m.startTime)
	if uptime == 0 {
		return 0
	}
	return float64(m.TotalErrors) / uptime.Minutes()
}

// Stats represents error statistics
type Stats struct {
	TotalErrors    int64
	ErrorsByType   map[string]int64
	ErrorsBySource map[string]int64
	LastError      time.Time
	LastErrorMsg   string
	Uptime         time.Duration
	ErrorRate      float64 // Errors per minute
}

// Global metrics instance
var globalMetrics = NewMetrics()

// RecordGlobalError records an error to global metrics
func RecordGlobalError(source, errType string, err error) {
	globalMetrics.RecordError(source, errType, err)
}

// GetGlobalStats returns global error statistics
func GetGlobalStats() Stats {
	return globalMetrics.GetStats()
}

// ResetGlobalMetrics clears global error metrics
func ResetGlobalMetrics() {
	globalMetrics.Reset()
}

// LogStats logs current error statistics
func (m *Metrics) LogStats() {
	stats := m.GetStats()

	metricsLog.Info("Error Statistics:")
	metricsLog.Info("  Total Errors: %d", stats.TotalErrors)
	metricsLog.Info("  Error Rate: %.2f errors/min", stats.ErrorRate)
	metricsLog.Info("  Uptime: %v", stats.Uptime)

	if len(stats.ErrorsByType) > 0 {
		metricsLog.Info("  Errors by Type:")
		for errType, count := range stats.ErrorsByType {
			metricsLog.Info("    %s: %d", errType, count)
		}
	}

	if len(stats.ErrorsBySource) > 0 {
		metricsLog.Info("  Errors by Source:")
		for source, count := range stats.ErrorsBySource {
			metricsLog.Info("    %s: %d", source, count)
		}
	}

	if !stats.LastError.IsZero() {
		metricsLog.Info("  Last Error: %v (%s)", stats.LastError, stats.LastErrorMsg)
	}
}

// LogGlobalStats logs global error statistics
func LogGlobalStats() {
	globalMetrics.LogStats()
}
