// File: internal/metrics/metrics.go
// Project: Terminal Velocity
// Description: Centralized metrics collection and Prometheus-compatible export
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector manages all game metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Connection metrics
	totalConnections    atomic.Int64
	activeConnections   atomic.Int64
	failedConnections   atomic.Int64
	connectionDurations []time.Duration

	// Player metrics
	activePlayers       atomic.Int64
	totalLogins         atomic.Int64
	totalRegistrations  atomic.Int64

	// Game activity metrics
	tradesCompleted     atomic.Int64
	combatEncounters    atomic.Int64
	missionsCompleted   atomic.Int64
	questsCompleted     atomic.Int64
	jumpsExecuted       atomic.Int64
	cargoTransferred    atomic.Int64

	// Economy metrics
	totalCreditsInGame  atomic.Int64
	totalMarketVolume   atomic.Int64
	tradeVolume24h      atomic.Int64

	// System metrics
	databaseQueries     atomic.Int64
	databaseErrors      atomic.Int64
	cacheHits           atomic.Int64
	cacheMisses         atomic.Int64

	// Performance metrics
	averageTickTime     time.Duration
	peakPlayers         int64
	peakTime            time.Time

	// Custom counters
	customCounters map[string]*atomic.Int64
	customGauges   map[string]*atomic.Int64

	// Start time
	startTime time.Time
}

// Global metrics instance
var global *MetricsCollector
var once sync.Once

// Init initializes the global metrics collector
func Init() *MetricsCollector {
	once.Do(func() {
		global = &MetricsCollector{
			customCounters: make(map[string]*atomic.Int64),
			customGauges:   make(map[string]*atomic.Int64),
			startTime:      time.Now(),
		}
	})
	return global
}

// Global returns the global metrics collector
func Global() *MetricsCollector {
	if global == nil {
		return Init()
	}
	return global
}

// Connection metrics
func (m *MetricsCollector) IncrementConnections() {
	m.totalConnections.Add(1)
}

func (m *MetricsCollector) IncrementActiveConnections() {
	current := m.activeConnections.Add(1)
	m.updatePeakPlayers(current)
}

func (m *MetricsCollector) DecrementActiveConnections() {
	m.activeConnections.Add(-1)
}

func (m *MetricsCollector) IncrementFailedConnections() {
	m.failedConnections.Add(1)
}

func (m *MetricsCollector) RecordConnectionDuration(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectionDurations = append(m.connectionDurations, d)
	// Keep only last 1000 durations
	if len(m.connectionDurations) > 1000 {
		m.connectionDurations = m.connectionDurations[len(m.connectionDurations)-1000:]
	}
}

// Player metrics
func (m *MetricsCollector) IncrementActivePlayers() {
	current := m.activePlayers.Add(1)
	m.updatePeakPlayers(current)
}

func (m *MetricsCollector) DecrementActivePlayers() {
	m.activePlayers.Add(-1)
}

func (m *MetricsCollector) IncrementLogins() {
	m.totalLogins.Add(1)
}

func (m *MetricsCollector) IncrementRegistrations() {
	m.totalRegistrations.Add(1)
}

// Game activity metrics
func (m *MetricsCollector) IncrementTrades() {
	m.tradesCompleted.Add(1)
}

func (m *MetricsCollector) IncrementCombat() {
	m.combatEncounters.Add(1)
}

func (m *MetricsCollector) IncrementMissions() {
	m.missionsCompleted.Add(1)
}

func (m *MetricsCollector) IncrementQuests() {
	m.questsCompleted.Add(1)
}

func (m *MetricsCollector) IncrementJumps() {
	m.jumpsExecuted.Add(1)
}

func (m *MetricsCollector) RecordCargoTransfer(tons int64) {
	m.cargoTransferred.Add(tons)
}

// Economy metrics
func (m *MetricsCollector) UpdateTotalCredits(total int64) {
	m.totalCreditsInGame.Store(total)
}

func (m *MetricsCollector) RecordMarketTransaction(volume int64) {
	m.totalMarketVolume.Add(volume)
	m.tradeVolume24h.Add(volume)
}

// System metrics
func (m *MetricsCollector) IncrementDBQueries() {
	m.databaseQueries.Add(1)
}

func (m *MetricsCollector) IncrementDBErrors() {
	m.databaseErrors.Add(1)
}

func (m *MetricsCollector) IncrementCacheHits() {
	m.cacheHits.Add(1)
}

func (m *MetricsCollector) IncrementCacheMisses() {
	m.cacheMisses.Add(1)
}

// Performance metrics
func (m *MetricsCollector) RecordTickTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.averageTickTime = d
}

func (m *MetricsCollector) updatePeakPlayers(current int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if current > m.peakPlayers {
		m.peakPlayers = current
		m.peakTime = time.Now()
	}
}

// Custom metrics
func (m *MetricsCollector) IncrementCounter(name string) {
	m.mu.Lock()
	if _, ok := m.customCounters[name]; !ok {
		m.customCounters[name] = &atomic.Int64{}
	}
	counter := m.customCounters[name]
	m.mu.Unlock()
	counter.Add(1)
}

func (m *MetricsCollector) SetGauge(name string, value int64) {
	m.mu.Lock()
	if _, ok := m.customGauges[name]; !ok {
		m.customGauges[name] = &atomic.Int64{}
	}
	gauge := m.customGauges[name]
	m.mu.Unlock()
	gauge.Store(value)
}

// Snapshot returns a complete snapshot of all metrics
type MetricsSnapshot struct {
	// Connection metrics
	TotalConnections    int64
	ActiveConnections   int64
	FailedConnections   int64
	AvgConnectionTime   time.Duration

	// Player metrics
	ActivePlayers      int64
	TotalLogins        int64
	TotalRegistrations int64

	// Game activity
	TradesCompleted   int64
	CombatEncounters  int64
	MissionsCompleted int64
	QuestsCompleted   int64
	JumpsExecuted     int64
	CargoTransferred  int64

	// Economy
	TotalCreditsInGame int64
	TotalMarketVolume  int64
	TradeVolume24h     int64

	// System
	DatabaseQueries int64
	DatabaseErrors  int64
	CacheHits       int64
	CacheMisses     int64
	CacheHitRate    float64

	// Performance
	AvgTickTime time.Duration
	PeakPlayers int64
	PeakTime    time.Time
	Uptime      time.Duration

	// Custom metrics
	CustomCounters map[string]int64
	CustomGauges   map[string]int64
}

func (m *MetricsCollector) Snapshot() *MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate average connection time
	var avgConnTime time.Duration
	if len(m.connectionDurations) > 0 {
		var total time.Duration
		for _, d := range m.connectionDurations {
			total += d
		}
		avgConnTime = total / time.Duration(len(m.connectionDurations))
	}

	// Calculate cache hit rate
	hits := m.cacheHits.Load()
	misses := m.cacheMisses.Load()
	var hitRate float64
	if total := hits + misses; total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	// Copy custom metrics
	customCounters := make(map[string]int64)
	for k, v := range m.customCounters {
		customCounters[k] = v.Load()
	}
	customGauges := make(map[string]int64)
	for k, v := range m.customGauges {
		customGauges[k] = v.Load()
	}

	return &MetricsSnapshot{
		TotalConnections:    m.totalConnections.Load(),
		ActiveConnections:   m.activeConnections.Load(),
		FailedConnections:   m.failedConnections.Load(),
		AvgConnectionTime:   avgConnTime,
		ActivePlayers:       m.activePlayers.Load(),
		TotalLogins:         m.totalLogins.Load(),
		TotalRegistrations:  m.totalRegistrations.Load(),
		TradesCompleted:     m.tradesCompleted.Load(),
		CombatEncounters:    m.combatEncounters.Load(),
		MissionsCompleted:   m.missionsCompleted.Load(),
		QuestsCompleted:     m.questsCompleted.Load(),
		JumpsExecuted:       m.jumpsExecuted.Load(),
		CargoTransferred:    m.cargoTransferred.Load(),
		TotalCreditsInGame:  m.totalCreditsInGame.Load(),
		TotalMarketVolume:   m.totalMarketVolume.Load(),
		TradeVolume24h:      m.tradeVolume24h.Load(),
		DatabaseQueries:     m.databaseQueries.Load(),
		DatabaseErrors:      m.databaseErrors.Load(),
		CacheHits:           m.cacheHits.Load(),
		CacheMisses:         m.cacheMisses.Load(),
		CacheHitRate:        hitRate,
		AvgTickTime:         m.averageTickTime,
		PeakPlayers:         m.peakPlayers,
		PeakTime:            m.peakTime,
		Uptime:              time.Since(m.startTime),
		CustomCounters:      customCounters,
		CustomGauges:        customGauges,
	}
}

// PrometheusFormat returns metrics in Prometheus exposition format
func (m *MetricsCollector) PrometheusFormat() string {
	snap := m.Snapshot()

	var out string
	out += fmt.Sprintf("# HELP terminal_velocity_connections_total Total number of SSH connections\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_connections_total counter\n")
	out += fmt.Sprintf("terminal_velocity_connections_total %d\n\n", snap.TotalConnections)

	out += fmt.Sprintf("# HELP terminal_velocity_connections_active Currently active SSH connections\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_connections_active gauge\n")
	out += fmt.Sprintf("terminal_velocity_connections_active %d\n\n", snap.ActiveConnections)

	out += fmt.Sprintf("# HELP terminal_velocity_connections_failed Total failed connection attempts\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_connections_failed counter\n")
	out += fmt.Sprintf("terminal_velocity_connections_failed %d\n\n", snap.FailedConnections)

	out += fmt.Sprintf("# HELP terminal_velocity_players_active Currently active players\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_players_active gauge\n")
	out += fmt.Sprintf("terminal_velocity_players_active %d\n\n", snap.ActivePlayers)

	out += fmt.Sprintf("# HELP terminal_velocity_logins_total Total player logins\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_logins_total counter\n")
	out += fmt.Sprintf("terminal_velocity_logins_total %d\n\n", snap.TotalLogins)

	out += fmt.Sprintf("# HELP terminal_velocity_registrations_total Total player registrations\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_registrations_total counter\n")
	out += fmt.Sprintf("terminal_velocity_registrations_total %d\n\n", snap.TotalRegistrations)

	out += fmt.Sprintf("# HELP terminal_velocity_trades_total Total completed trades\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_trades_total counter\n")
	out += fmt.Sprintf("terminal_velocity_trades_total %d\n\n", snap.TradesCompleted)

	out += fmt.Sprintf("# HELP terminal_velocity_combat_total Total combat encounters\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_combat_total counter\n")
	out += fmt.Sprintf("terminal_velocity_combat_total %d\n\n", snap.CombatEncounters)

	out += fmt.Sprintf("# HELP terminal_velocity_missions_completed_total Total completed missions\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_missions_completed_total counter\n")
	out += fmt.Sprintf("terminal_velocity_missions_completed_total %d\n\n", snap.MissionsCompleted)

	out += fmt.Sprintf("# HELP terminal_velocity_quests_completed_total Total completed quests\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_quests_completed_total counter\n")
	out += fmt.Sprintf("terminal_velocity_quests_completed_total %d\n\n", snap.QuestsCompleted)

	out += fmt.Sprintf("# HELP terminal_velocity_jumps_total Total hyperspace jumps\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_jumps_total counter\n")
	out += fmt.Sprintf("terminal_velocity_jumps_total %d\n\n", snap.JumpsExecuted)

	out += fmt.Sprintf("# HELP terminal_velocity_cargo_transferred_tons Total cargo transferred in tons\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_cargo_transferred_tons counter\n")
	out += fmt.Sprintf("terminal_velocity_cargo_transferred_tons %d\n\n", snap.CargoTransferred)

	out += fmt.Sprintf("# HELP terminal_velocity_economy_credits_total Total credits in game economy\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_economy_credits_total gauge\n")
	out += fmt.Sprintf("terminal_velocity_economy_credits_total %d\n\n", snap.TotalCreditsInGame)

	out += fmt.Sprintf("# HELP terminal_velocity_market_volume_total Total market transaction volume\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_market_volume_total counter\n")
	out += fmt.Sprintf("terminal_velocity_market_volume_total %d\n\n", snap.TotalMarketVolume)

	out += fmt.Sprintf("# HELP terminal_velocity_db_queries_total Total database queries\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_db_queries_total counter\n")
	out += fmt.Sprintf("terminal_velocity_db_queries_total %d\n\n", snap.DatabaseQueries)

	out += fmt.Sprintf("# HELP terminal_velocity_db_errors_total Total database errors\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_db_errors_total counter\n")
	out += fmt.Sprintf("terminal_velocity_db_errors_total %d\n\n", snap.DatabaseErrors)

	out += fmt.Sprintf("# HELP terminal_velocity_cache_hits_total Total cache hits\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_cache_hits_total counter\n")
	out += fmt.Sprintf("terminal_velocity_cache_hits_total %d\n\n", snap.CacheHits)

	out += fmt.Sprintf("# HELP terminal_velocity_cache_misses_total Total cache misses\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_cache_misses_total counter\n")
	out += fmt.Sprintf("terminal_velocity_cache_misses_total %d\n\n", snap.CacheMisses)

	out += fmt.Sprintf("# HELP terminal_velocity_cache_hit_rate Cache hit rate percentage\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_cache_hit_rate gauge\n")
	out += fmt.Sprintf("terminal_velocity_cache_hit_rate %.2f\n\n", snap.CacheHitRate)

	out += fmt.Sprintf("# HELP terminal_velocity_peak_players Peak concurrent players\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_peak_players gauge\n")
	out += fmt.Sprintf("terminal_velocity_peak_players %d\n\n", snap.PeakPlayers)

	out += fmt.Sprintf("# HELP terminal_velocity_uptime_seconds Server uptime in seconds\n")
	out += fmt.Sprintf("# TYPE terminal_velocity_uptime_seconds gauge\n")
	out += fmt.Sprintf("terminal_velocity_uptime_seconds %.0f\n\n", snap.Uptime.Seconds())

	// Custom counters
	for name, value := range snap.CustomCounters {
		out += fmt.Sprintf("# HELP terminal_velocity_custom_%s Custom counter\n", name)
		out += fmt.Sprintf("# TYPE terminal_velocity_custom_%s counter\n", name)
		out += fmt.Sprintf("terminal_velocity_custom_%s %d\n\n", name, value)
	}

	// Custom gauges
	for name, value := range snap.CustomGauges {
		out += fmt.Sprintf("# HELP terminal_velocity_custom_%s Custom gauge\n", name)
		out += fmt.Sprintf("# TYPE terminal_velocity_custom_%s gauge\n", name)
		out += fmt.Sprintf("terminal_velocity_custom_%s %d\n\n", name, value)
	}

	return out
}

// Reset24hCounters resets the 24-hour rolling counters
func (m *MetricsCollector) Reset24hCounters() {
	m.tradeVolume24h.Store(0)
}
