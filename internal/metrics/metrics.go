// File: internal/metrics/metrics.go
// Project: Terminal Velocity
// Description: Centralized metrics collection and Prometheus-compatible export
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

// Package metrics provides centralized observability and monitoring infrastructure for Terminal Velocity.
//
// This package implements a thread-safe metrics collection system that tracks server health,
// player activity, game events, and system performance. It provides Prometheus-compatible
// export format for integration with monitoring tools like Prometheus and Grafana.
//
// Features:
//   - Connection metrics (total, active, failed, duration)
//   - Player metrics (active, logins, registrations, peak)
//   - Game activity tracking (trades, combat, missions, quests, jumps, cargo)
//   - Economy monitoring (total credits, market volume, 24h trade volume)
//   - System health (database queries/errors, cache hit rate, uptime)
//   - Custom counters and gauges for application-specific metrics
//   - Thread-safe atomic operations for high-concurrency scenarios
//   - Prometheus exposition format export
//   - Snapshot capability for point-in-time metrics
//
// Architecture:
//   The package uses a singleton pattern with a global MetricsCollector instance.
//   All counters use atomic.Int64 for lock-free increments, while gauges and
//   aggregations use sync.RWMutex for thread-safe access.
//
// Usage Example:
//
//	// Initialize metrics during server startup
//	m := metrics.Init()
//
//	// Track player activity
//	m.IncrementLogins()
//	m.IncrementActivePlayers()
//	defer m.DecrementActivePlayers()
//
//	// Track game events
//	m.IncrementTrades()
//	m.RecordCargoTransfer(50)
//	m.RecordMarketTransaction(10000)
//
//	// Export for Prometheus
//	prometheusData := m.PrometheusFormat()
//
//	// Get snapshot for display
//	snapshot := m.Snapshot()
//	fmt.Printf("Active players: %d\n", snapshot.ActivePlayers)
//	fmt.Printf("Cache hit rate: %.2f%%\n", snapshot.CacheHitRate)
package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector is the central hub for collecting and exporting server metrics.
//
// The collector maintains thread-safe counters and gauges for all aspects of server
// operation including connections, players, game activity, economy, and system health.
// It uses atomic operations for high-frequency counters to minimize lock contention.
//
// Thread Safety:
//   Most counters use atomic.Int64 for lock-free operations. The sync.RWMutex protects
//   slices, maps, and aggregate calculations. All methods are safe for concurrent use.
//
// Metric Categories:
//
// Connection Metrics:
//   - totalConnections: Cumulative SSH connections since server start
//   - activeConnections: Current concurrent connections (gauge)
//   - failedConnections: Cumulative failed authentication attempts
//   - connectionDurations: Last 1000 connection durations for average calculation
//
// Player Metrics:
//   - activePlayers: Current logged-in players (gauge)
//   - totalLogins: Cumulative successful logins
//   - totalRegistrations: Cumulative new player registrations
//   - peakPlayers: Highest concurrent players ever reached
//   - peakTime: Timestamp when peak was reached
//
// Game Activity Metrics:
//   - tradesCompleted: Total market trades executed
//   - combatEncounters: Total combat encounters started
//   - missionsCompleted: Total missions finished
//   - questsCompleted: Total quests finished
//   - jumpsExecuted: Total hyperspace jumps
//   - cargoTransferred: Total cargo moved in tons
//
// Economy Metrics:
//   - totalCreditsInGame: Current total credits across all players (gauge)
//   - totalMarketVolume: Cumulative value of all market transactions
//   - tradeVolume24h: Rolling 24-hour trade volume (reset daily)
//
// System Health Metrics:
//   - databaseQueries: Total database queries executed
//   - databaseErrors: Total database errors encountered
//   - cacheHits: Cache lookups that found data
//   - cacheMisses: Cache lookups that missed
//
// Performance Metrics:
//   - averageTickTime: Average game tick processing time
//   - startTime: Server start timestamp for uptime calculation
//
// Custom Metrics:
//   - customCounters: Application-defined incrementing counters
//   - customGauges: Application-defined settable values
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

var (
	// global is the singleton MetricsCollector instance used by the server.
	// It is initialized once via Init() using sync.Once for thread-safety.
	global *MetricsCollector

	// once ensures global is initialized exactly once even with concurrent Init() calls.
	once sync.Once
)

// Init initializes and returns the global singleton MetricsCollector.
//
// This function should be called once during server startup. It creates a new
// MetricsCollector with initialized maps and sets the server start time. Subsequent
// calls return the existing instance without re-initialization.
//
// Returns:
//   - *MetricsCollector: The global metrics collector instance
//
// Thread Safety:
//   Safe for concurrent calls. Only the first call initializes the collector.
//
// Example:
//   m := metrics.Init()
//   m.IncrementLogins()
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

// Global returns the global MetricsCollector instance.
//
// If Init() has not been called yet, this function calls it automatically to
// ensure a valid collector is always returned. This allows metrics to be
// recorded without explicit initialization.
//
// Returns:
//   - *MetricsCollector: The global metrics collector instance
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   metrics.Global().IncrementConnections()
//   metrics.Global().RecordMarketTransaction(1000)
func Global() *MetricsCollector {
	if global == nil {
		return Init()
	}
	return global
}

// ============================================================================
// Connection Metrics
//
// These methods track SSH connection activity including total connections,
// active connections, failures, and connection duration statistics.
// All methods are thread-safe via atomic operations.
// ============================================================================

// IncrementConnections increments the total connection counter.
//
// This should be called when a new SSH connection is established, regardless
// of whether authentication succeeds. It provides a cumulative count of all
// connection attempts since server start.
//
// Thread Safety:
//   Safe for concurrent use via atomic operations.
//
// Example:
//   // In SSH server handler
//   metrics.Global().IncrementConnections()
func (m *MetricsCollector) IncrementConnections() {
	m.totalConnections.Add(1)
}

// IncrementActiveConnections increments the active connections gauge and updates peak.
//
// This should be called when a player successfully authenticates and begins a session.
// It automatically checks if this is a new peak concurrent player count.
//
// Thread Safety:
//   Safe for concurrent use via atomic operations and mutex-protected peak update.
//
// Example:
//   metrics.Global().IncrementActiveConnections()
//   defer metrics.Global().DecrementActiveConnections()
func (m *MetricsCollector) IncrementActiveConnections() {
	current := m.activeConnections.Add(1)
	m.updatePeakPlayers(current)
}

// DecrementActiveConnections decrements the active connections gauge.
//
// This should be called when a player disconnects or their session ends.
//
// Thread Safety:
//   Safe for concurrent use via atomic operations.
func (m *MetricsCollector) DecrementActiveConnections() {
	m.activeConnections.Add(-1)
}

// IncrementFailedConnections increments the failed connection counter.
//
// This should be called when authentication fails or a connection is rejected.
// Used for security monitoring and rate limit tracking.
//
// Thread Safety:
//   Safe for concurrent use via atomic operations.
func (m *MetricsCollector) IncrementFailedConnections() {
	m.failedConnections.Add(1)
}

// RecordConnectionDuration records the duration of a connection for average calculation.
//
// The collector maintains a rolling window of the last 1000 connection durations
// to calculate average connection time. This helps identify if players are having
// short-lived connections (potential issues) or long sessions (engaged players).
//
// Parameters:
//   - d: Duration of the connection that just ended
//
// Thread Safety:
//   Safe for concurrent use via sync.RWMutex.
//
// Example:
//   start := time.Now()
//   // ... handle connection ...
//   metrics.Global().RecordConnectionDuration(time.Since(start))
func (m *MetricsCollector) RecordConnectionDuration(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectionDurations = append(m.connectionDurations, d)
	// Keep only last 1000 durations
	if len(m.connectionDurations) > 1000 {
		m.connectionDurations = m.connectionDurations[len(m.connectionDurations)-1000:]
	}
}

// ============================================================================
// Player Metrics
//
// These methods track player activity including active players, logins,
// and registrations. Active players is a gauge that goes up/down, while
// logins and registrations are cumulative counters.
// ============================================================================

// IncrementActivePlayers increments active players gauge and updates peak.
// Call when a player successfully logs in. Thread-safe.
func (m *MetricsCollector) IncrementActivePlayers() {
	current := m.activePlayers.Add(1)
	m.updatePeakPlayers(current)
}

// DecrementActivePlayers decrements active players gauge.
// Call when a player logs out or disconnects. Thread-safe.
func (m *MetricsCollector) DecrementActivePlayers() {
	m.activePlayers.Add(-1)
}

// IncrementLogins increments total login counter.
// Call on successful authentication. Thread-safe.
func (m *MetricsCollector) IncrementLogins() {
	m.totalLogins.Add(1)
}

// IncrementRegistrations increments new player registration counter.
// Call when a new account is created. Thread-safe.
func (m *MetricsCollector) IncrementRegistrations() {
	m.totalRegistrations.Add(1)
}

// ============================================================================
// Game Activity Metrics
//
// These methods track in-game actions and events. All are cumulative counters
// that increment with each event occurrence. Thread-safe via atomic operations.
// ============================================================================

// IncrementTrades increments completed trades counter. Thread-safe.
func (m *MetricsCollector) IncrementTrades() {
	m.tradesCompleted.Add(1)
}

// IncrementCombat increments combat encounters counter. Thread-safe.
func (m *MetricsCollector) IncrementCombat() {
	m.combatEncounters.Add(1)
}

// IncrementMissions increments completed missions counter. Thread-safe.
func (m *MetricsCollector) IncrementMissions() {
	m.missionsCompleted.Add(1)
}

// IncrementQuests increments completed quests counter. Thread-safe.
func (m *MetricsCollector) IncrementQuests() {
	m.questsCompleted.Add(1)
}

// IncrementJumps increments hyperspace jumps counter. Thread-safe.
func (m *MetricsCollector) IncrementJumps() {
	m.jumpsExecuted.Add(1)
}

// RecordCargoTransfer adds to total cargo transferred counter.
// Parameters:
//   - tons: Amount of cargo moved in tons
// Thread-safe via atomic operations.
func (m *MetricsCollector) RecordCargoTransfer(tons int64) {
	m.cargoTransferred.Add(tons)
}

// ============================================================================
// Economy Metrics
//
// These methods track the game's economic activity including total credits,
// market volume, and rolling 24-hour trade volume.
// ============================================================================

// UpdateTotalCredits sets the current total credits gauge.
//
// This is a gauge (not a counter) that should be updated periodically to reflect
// the current total credits across all players. Useful for monitoring economy inflation.
//
// Parameters:
//   - total: Current sum of all player credits in the game
//
// Thread Safety:
//   Safe for concurrent use via atomic store operation.
func (m *MetricsCollector) UpdateTotalCredits(total int64) {
	m.totalCreditsInGame.Store(total)
}

// RecordMarketTransaction records a market transaction in both total and 24h volume.
//
// This method updates both the cumulative market volume (all-time) and the
// rolling 24-hour volume. The 24h volume is reset daily via Reset24hCounters().
//
// Parameters:
//   - volume: Value of the transaction in credits
//
// Thread Safety:
//   Safe for concurrent use via atomic operations.
//
// Example:
//   // Player buys 10 tons of Food at 100 cr/ton
//   metrics.Global().RecordMarketTransaction(1000)
func (m *MetricsCollector) RecordMarketTransaction(volume int64) {
	m.totalMarketVolume.Add(volume)
	m.tradeVolume24h.Add(volume)
}

// ============================================================================
// System Health Metrics
//
// These methods track system-level health indicators including database
// performance and cache effectiveness.
// ============================================================================

// IncrementDBQueries increments database query counter. Thread-safe.
func (m *MetricsCollector) IncrementDBQueries() {
	m.databaseQueries.Add(1)
}

// IncrementDBErrors increments database error counter. Thread-safe.
func (m *MetricsCollector) IncrementDBErrors() {
	m.databaseErrors.Add(1)
}

// IncrementCacheHits increments cache hit counter. Thread-safe.
func (m *MetricsCollector) IncrementCacheHits() {
	m.cacheHits.Add(1)
}

// IncrementCacheMisses increments cache miss counter. Thread-safe.
func (m *MetricsCollector) IncrementCacheMisses() {
	m.cacheMisses.Add(1)
}

// ============================================================================
// Performance Metrics
// ============================================================================

// RecordTickTime records the average game tick processing time.
// Thread-safe via sync.RWMutex.
func (m *MetricsCollector) RecordTickTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.averageTickTime = d
}

// updatePeakPlayers is an internal method that updates peak player count if current exceeds it.
// Called automatically by IncrementActiveConnections() and IncrementActivePlayers().
// Thread-safe via sync.RWMutex.
func (m *MetricsCollector) updatePeakPlayers(current int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if current > m.peakPlayers {
		m.peakPlayers = current
		m.peakTime = time.Now()
	}
}

// ============================================================================
// Custom Metrics
//
// These methods allow application code to define custom counters and gauges
// that don't fit into the predefined metric categories.
// ============================================================================

// IncrementCounter increments a named custom counter.
//
// If the counter doesn't exist, it is created automatically. This allows for
// dynamic metrics without requiring pre-definition.
//
// Parameters:
//   - name: Name of the counter (e.g., "special_events", "rare_drops")
//
// Thread Safety:
//   Safe for concurrent use. Map modifications protected by sync.RWMutex,
//   counter increments use atomic operations.
//
// Example:
//   metrics.Global().IncrementCounter("faction_wars_started")
//   metrics.Global().IncrementCounter("rare_item_drops")
func (m *MetricsCollector) IncrementCounter(name string) {
	m.mu.Lock()
	if _, ok := m.customCounters[name]; !ok {
		m.customCounters[name] = &atomic.Int64{}
	}
	counter := m.customCounters[name]
	m.mu.Unlock()
	counter.Add(1)
}

// SetGauge sets a named custom gauge to a specific value.
//
// Gauges differ from counters in that they can be set to arbitrary values
// rather than just incremented. Useful for tracking current levels or states.
//
// Parameters:
//   - name: Name of the gauge (e.g., "queue_size", "active_events")
//   - value: Value to set the gauge to
//
// Thread Safety:
//   Safe for concurrent use. Map modifications protected by sync.RWMutex,
//   gauge stores use atomic operations.
//
// Example:
//   metrics.Global().SetGauge("active_events", int64(len(events)))
//   metrics.Global().SetGauge("pending_trades", int64(tradeQueue.Len()))
func (m *MetricsCollector) SetGauge(name string, value int64) {
	m.mu.Lock()
	if _, ok := m.customGauges[name]; !ok {
		m.customGauges[name] = &atomic.Int64{}
	}
	gauge := m.customGauges[name]
	m.mu.Unlock()
	gauge.Store(value)
}

// ============================================================================
// Snapshot and Export
// ============================================================================

// MetricsSnapshot is a point-in-time snapshot of all metrics values.
//
// This struct contains all metric values at the moment Snapshot() was called.
// It's useful for generating reports, displaying stats pages, or exporting
// to external monitoring systems.
//
// All values are simple types (int64, float64, time.Duration, etc.) so the
// snapshot can be safely passed between goroutines and serialized without
// worrying about concurrent modifications to the underlying collector.
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

// Snapshot creates and returns a point-in-time snapshot of all metrics.
//
// This method safely captures all current metric values including counters, gauges,
// calculated values (like cache hit rate), and custom metrics. The snapshot is
// safe to use concurrently and can be passed to other goroutines or serialized.
//
// The method performs several calculations:
//   - Average connection time from the last 1000 connections
//   - Cache hit rate as a percentage (hits / (hits + misses) * 100)
//   - Uptime since server start
//   - Copies of all custom counter and gauge values
//
// Returns:
//   - *MetricsSnapshot: A complete snapshot of all metrics at this moment
//
// Thread Safety:
//   Safe for concurrent use. Uses sync.RWMutex and atomic loads to ensure
//   consistent reads without blocking metric updates.
//
// Example:
//   snapshot := metrics.Global().Snapshot()
//   fmt.Printf("Active players: %d\n", snapshot.ActivePlayers)
//   fmt.Printf("Peak players: %d (at %s)\n", snapshot.PeakPlayers, snapshot.PeakTime)
//   fmt.Printf("Cache hit rate: %.2f%%\n", snapshot.CacheHitRate)
//   fmt.Printf("Uptime: %s\n", snapshot.Uptime)
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

// PrometheusFormat returns all metrics in Prometheus exposition format.
//
// This method generates a string containing all metrics in the standard Prometheus
// text-based exposition format, which can be scraped by Prometheus or compatible
// monitoring systems. Each metric includes HELP and TYPE metadata followed by
// the metric value.
//
// The output includes:
//   - All predefined counters and gauges with "terminal_velocity_" prefix
//   - Custom counters with "terminal_velocity_custom_" prefix
//   - Custom gauges with "terminal_velocity_custom_" prefix
//   - Calculated metrics (cache hit rate, uptime)
//
// Returns:
//   - string: Prometheus-formatted metrics suitable for /metrics endpoint
//
// Thread Safety:
//   Safe for concurrent use. Creates a snapshot first to ensure consistency.
//
// Format Example:
//   # HELP terminal_velocity_connections_total Total number of SSH connections
//   # TYPE terminal_velocity_connections_total counter
//   terminal_velocity_connections_total 1234
//
//   # HELP terminal_velocity_players_active Currently active players
//   # TYPE terminal_velocity_players_active gauge
//   terminal_velocity_players_active 42
//
// Usage:
//   http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
//       w.Header().Set("Content-Type", "text/plain; version=0.0.4")
//       fmt.Fprint(w, metrics.Global().PrometheusFormat())
//   })
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

// Reset24hCounters resets all 24-hour rolling window counters to zero.
//
// This method should be called on a daily schedule (e.g., via cron job or ticker)
// to reset counters that track rolling 24-hour metrics. Currently resets:
//   - tradeVolume24h: Total trade volume in the last 24 hours
//
// Thread Safety:
//   Safe for concurrent use via atomic store operation.
//
// Example:
//   // Reset daily at midnight
//   ticker := time.NewTicker(24 * time.Hour)
//   go func() {
//       for range ticker.C {
//           metrics.Global().Reset24hCounters()
//           logger.Info("24-hour metrics counters reset")
//       }
//   }()
func (m *MetricsCollector) Reset24hCounters() {
	m.tradeVolume24h.Store(0)
}
