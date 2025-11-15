// File: internal/metrics/server.go
// Project: Terminal Velocity
// Description: HTTP server for metrics endpoint
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("Metrics")

// Server provides an HTTP endpoint for Prometheus metrics
type Server struct {
	addr       string
	collector  *MetricsCollector
	httpServer *http.Server
	wg         sync.WaitGroup
}

// NewServer creates a new metrics server
func NewServer(addr string, collector *MetricsCollector) *Server {
	return &Server{
		addr:      addr,
		collector: collector,
	}
}

// Start begins serving metrics on the configured address
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Prometheus-compatible metrics endpoint
	mux.HandleFunc("/metrics", s.handleMetrics)

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Human-readable stats endpoint
	mux.HandleFunc("/stats", s.handleStats)

	// Enhanced stats with latency and errors
	mux.HandleFunc("/stats/enhanced", s.handleEnhancedStats)

	// Performance profiling endpoint
	mux.HandleFunc("/stats/performance", s.handlePerformanceStats)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info("Starting metrics server on %s", s.addr)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Metrics server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the metrics server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	log.Info("Shutting down metrics server")
	err := s.httpServer.Shutdown(ctx)
	s.wg.Wait() // Wait for HTTP server goroutine to finish
	return err
}

// handleMetrics serves Prometheus-formatted metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprint(w, s.collector.PrometheusFormat())
}

// handleHealth serves a comprehensive health check with service status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	snap := s.collector.Snapshot()
	enhanced := GetEnhanced()

	w.Header().Set("Content-Type", "application/json")

	// Calculate error rate
	errorRate := 0.0
	if snap.DatabaseQueries > 0 {
		errorRate = (float64(snap.DatabaseErrors) / float64(snap.DatabaseQueries)) * 100
	}

	// Get latency metrics
	dbP99, _, _ := enhanced.OperationLatency.GetPercentiles("database")

	// Determine overall health status
	status := "healthy"
	statusCode := 200

	// Degraded if:
	// - Error rate > 1%
	// - Database p99 latency > 500ms
	// - Cache hit rate < 50%
	if errorRate > 1 || dbP99 > 500*time.Millisecond || snap.CacheHitRate < 50 {
		status = "degraded"
		statusCode = 200 // Still return 200 for degraded
	}

	// Unhealthy if:
	// - Error rate > 5%
	// - Database p99 latency > 2s
	// - Database errors in last check
	if errorRate > 5 || dbP99 > 2*time.Second {
		status = "unhealthy"
		statusCode = 503 // Service Unavailable
	}

	w.WriteHeader(statusCode)

	// Write JSON health response
	fmt.Fprintf(w, `{"status":"%s","uptime":"%s","active_connections":%d,"active_players":%d,"database_p99_latency":"%s","error_rate_percent":%.2f,"cache_hit_rate_percent":%.2f,"database_errors":%d,"timestamp":"%s"}`,
		status,
		snap.Uptime.Round(time.Second).String(),
		snap.ActiveConnections,
		snap.ActivePlayers,
		dbP99.Round(time.Millisecond).String(),
		errorRate,
		snap.CacheHitRate,
		snap.DatabaseErrors,
		time.Now().Format(time.RFC3339),
	)
}

// handleStats serves human-readable statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	snap := s.collector.Snapshot()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Terminal Velocity - Server Statistics</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #0a0a0a;
            color: #00ff00;
            padding: 20px;
            max-width: 1200px;
            margin: 0 auto;
        }
        h1, h2 {
            color: #00ff00;
            border-bottom: 2px solid #00ff00;
            padding-bottom: 10px;
        }
        .stat-group {
            background-color: #1a1a1a;
            border: 1px solid #00ff00;
            padding: 15px;
            margin: 20px 0;
            border-radius: 5px;
        }
        .stat-row {
            display: flex;
            justify-content: space-between;
            padding: 5px 0;
            border-bottom: 1px solid #333;
        }
        .stat-row:last-child {
            border-bottom: none;
        }
        .stat-label {
            color: #00cc00;
            font-weight: bold;
        }
        .stat-value {
            color: #00ff00;
            text-align: right;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .timestamp {
            text-align: center;
            color: #00cc00;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>═══════════════════════════════════════</h1>
        <h1>TERMINAL VELOCITY - SERVER STATISTICS</h1>
        <h1>═══════════════════════════════════════</h1>
    </div>

    <div class="stat-group">
        <h2>🌐 CONNECTION STATISTICS</h2>
        <div class="stat-row">
            <span class="stat-label">Total Connections:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Active Connections:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Failed Connections:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Average Connection Time:</span>
            <span class="stat-value">%v</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>👥 PLAYER STATISTICS</h2>
        <div class="stat-row">
            <span class="stat-label">Active Players:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Total Logins:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Total Registrations:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Peak Players:</span>
            <span class="stat-value">%d (at %v)</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>🎮 GAME ACTIVITY</h2>
        <div class="stat-row">
            <span class="stat-label">Trades Completed:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Encounters:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Missions Completed:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Quests Completed:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Hyperspace Jumps:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Cargo Transferred:</span>
            <span class="stat-value">%d tons</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>💰 ECONOMY</h2>
        <div class="stat-row">
            <span class="stat-label">Total Credits in Game:</span>
            <span class="stat-value">%d CR</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Total Market Volume:</span>
            <span class="stat-value">%d CR</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Trade Volume (24h):</span>
            <span class="stat-value">%d CR</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>⚙️ SYSTEM PERFORMANCE</h2>
        <div class="stat-row">
            <span class="stat-label">Database Queries:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Database Errors:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Cache Hit Rate:</span>
            <span class="stat-value">%.2f%%</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Server Uptime:</span>
            <span class="stat-value">%v</span>
        </div>
    </div>

    <div class="timestamp">
        Generated at: %s
    </div>
</body>
</html>`

	peakTimeStr := "Never"
	if !snap.PeakTime.IsZero() {
		peakTimeStr = snap.PeakTime.Format("2006-01-02 15:04:05")
	}

	fmt.Fprintf(w, html,
		snap.TotalConnections,
		snap.ActiveConnections,
		snap.FailedConnections,
		snap.AvgConnectionTime.Round(time.Millisecond),
		snap.ActivePlayers,
		snap.TotalLogins,
		snap.TotalRegistrations,
		snap.PeakPlayers, peakTimeStr,
		snap.TradesCompleted,
		snap.CombatEncounters,
		snap.MissionsCompleted,
		snap.QuestsCompleted,
		snap.JumpsExecuted,
		snap.CargoTransferred,
		snap.TotalCreditsInGame,
		snap.TotalMarketVolume,
		snap.TradeVolume24h,
		snap.DatabaseQueries,
		snap.DatabaseErrors,
		snap.CacheHitRate,
		snap.Uptime.Round(time.Second),
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)
}

// handleEnhancedStats serves enhanced statistics with latency and error tracking
func (s *Server) handleEnhancedStats(w http.ResponseWriter, r *http.Request) {
	enhanced := GetEnhanced()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Terminal Velocity - Enhanced Statistics</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #0a0a0a;
            color: #00ff00;
            padding: 20px;
            max-width: 1400px;
            margin: 0 auto;
        }
        h1, h2 {
            color: #00ff00;
            border-bottom: 2px solid #00ff00;
            padding-bottom: 10px;
        }
        .stat-group {
            background-color: #1a1a1a;
            border: 1px solid #00ff00;
            padding: 15px;
            margin: 20px 0;
            border-radius: 5px;
        }
        .stat-row {
            display: flex;
            justify-content: space-between;
            padding: 5px 0;
            border-bottom: 1px solid #333;
        }
        .stat-row:last-child {
            border-bottom: none;
        }
        .stat-label {
            color: #00cc00;
            font-weight: bold;
        }
        .stat-value {
            color: #00ff00;
            text-align: right;
        }
        .error-record {
            background-color: #2a1a1a;
            border-left: 3px solid #ff3333;
            padding: 10px;
            margin: 10px 0;
        }
        .error-time {
            color: #ff9999;
            font-size: 0.9em;
        }
        .error-category {
            color: #ffcc00;
            font-weight: bold;
        }
        .error-message {
            color: #ff6666;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .timestamp {
            text-align: center;
            color: #00cc00;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>═══════════════════════════════════════</h1>
        <h1>TERMINAL VELOCITY - ENHANCED STATISTICS</h1>
        <h1>═══════════════════════════════════════</h1>
    </div>

    <div class="stat-group">
        <h2>⚡ OPERATION LATENCIES</h2>
        <div class="stat-row">
            <span class="stat-label">Database Queries (p50 / p95 / p99):</span>
            <span class="stat-value">%v / %v / %v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Trade Operations (p50 / p95 / p99):</span>
            <span class="stat-value">%v / %v / %v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Operations (p50 / p95 / p99):</span>
            <span class="stat-value">%v / %v / %v</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>❌ ERROR TRACKING</h2>
        <div class="stat-row">
            <span class="stat-label">Database Errors:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Network Errors:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Game Logic Errors:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Validation Errors:</span>
            <span class="stat-value">%d</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>🔴 RECENT ERRORS (Last 10)</h2>
        %s
    </div>

    <div class="timestamp">
        Generated at: %s
    </div>
</body>
</html>`

	// Get latency percentiles
	dbP50, dbP95, dbP99 := enhanced.OperationLatency.GetPercentiles("database")
	tradeP50, tradeP95, tradeP99 := enhanced.OperationLatency.GetPercentiles("trade")
	combatP50, combatP95, combatP99 := enhanced.OperationLatency.GetPercentiles("combat")

	// Get error counts
	dbErrors := enhanced.Errors.GetCount("database")
	networkErrors := enhanced.Errors.GetCount("network")
	gameErrors := enhanced.Errors.GetCount("game_logic")
	validationErrors := enhanced.Errors.GetCount("validation")

	// Format recent errors
	recentErrors := enhanced.Errors.GetRecentErrors(10)
	var errorsHTML string
	if len(recentErrors) == 0 {
		errorsHTML = "<div class=\"error-record\">No recent errors</div>"
	} else {
		for _, err := range recentErrors {
			errorsHTML += fmt.Sprintf(`
        <div class="error-record">
            <div class="error-time">%s</div>
            <div class="error-category">Category: %s</div>
            <div class="error-message">%s</div>
        </div>`,
				err.Timestamp.Format("2006-01-02 15:04:05"),
				err.Category,
				err.Message,
			)
		}
	}

	fmt.Fprintf(w, html,
		dbP50.Round(time.Millisecond), dbP95.Round(time.Millisecond), dbP99.Round(time.Millisecond),
		tradeP50.Round(time.Millisecond), tradeP95.Round(time.Millisecond), tradeP99.Round(time.Millisecond),
		combatP50.Round(time.Millisecond), combatP95.Round(time.Millisecond), combatP99.Round(time.Millisecond),
		dbErrors,
		networkErrors,
		gameErrors,
		validationErrors,
		errorsHTML,
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)
}

// handlePerformanceStats serves detailed performance profiling data
func (s *Server) handlePerformanceStats(w http.ResponseWriter, r *http.Request) {
	enhanced := GetEnhanced()
	snap := s.collector.Snapshot()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Terminal Velocity - Performance Profiling</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #0a0a0a;
            color: #00ff00;
            padding: 20px;
            max-width: 1400px;
            margin: 0 auto;
        }
        h1, h2 {
            color: #00ff00;
            border-bottom: 2px solid #00ff00;
            padding-bottom: 10px;
        }
        .stat-group {
            background-color: #1a1a1a;
            border: 1px solid #00ff00;
            padding: 15px;
            margin: 20px 0;
            border-radius: 5px;
        }
        .stat-row {
            display: flex;
            justify-content: space-between;
            padding: 5px 0;
            border-bottom: 1px solid #333;
        }
        .stat-row:last-child {
            border-bottom: none;
        }
        .stat-label {
            color: #00cc00;
            font-weight: bold;
        }
        .stat-value {
            color: #00ff00;
            text-align: right;
        }
        .good {
            color: #00ff00;
        }
        .warning {
            color: #ffcc00;
        }
        .critical {
            color: #ff3333;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .timestamp {
            text-align: center;
            color: #00cc00;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>═══════════════════════════════════════</h1>
        <h1>TERMINAL VELOCITY - PERFORMANCE PROFILING</h1>
        <h1>═══════════════════════════════════════</h1>
    </div>

    <div class="stat-group">
        <h2>⏱️ OPERATION LATENCY BREAKDOWN</h2>
        <div class="stat-row">
            <span class="stat-label">Database - p50:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Database - p95:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Database - p99:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Trade Operations - p50:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Trade Operations - p95:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Trade Operations - p99:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Operations - p50:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Operations - p95:</span>
            <span class="stat-value %s">%v</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Operations - p99:</span>
            <span class="stat-value %s">%v</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>📊 THROUGHPUT METRICS</h2>
        <div class="stat-row">
            <span class="stat-label">Trades per Minute:</span>
            <span class="stat-value">%.2f</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Combat Encounters per Minute:</span>
            <span class="stat-value">%.2f</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Database Queries per Minute:</span>
            <span class="stat-value">%.2f</span>
        </div>
    </div>

    <div class="stat-group">
        <h2>💾 RESOURCE UTILIZATION</h2>
        <div class="stat-row">
            <span class="stat-label">Active Connections:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Active Players:</span>
            <span class="stat-value">%d</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Cache Hit Rate:</span>
            <span class="stat-value %s">%.2f%%</span>
        </div>
        <div class="stat-row">
            <span class="stat-label">Database Error Rate:</span>
            <span class="stat-value %s">%.2f%%</span>
        </div>
    </div>

    <div class="timestamp">
        Generated at: %s
    </div>
</body>
</html>`

	// Get latency percentiles
	dbP50, dbP95, dbP99 := enhanced.OperationLatency.GetPercentiles("database")
	tradeP50, tradeP95, tradeP99 := enhanced.OperationLatency.GetPercentiles("trade")
	combatP50, combatP95, combatP99 := enhanced.OperationLatency.GetPercentiles("combat")

	// Calculate throughput rates (per minute)
	uptime := snap.Uptime.Minutes()
	if uptime == 0 {
		uptime = 1 // Avoid division by zero
	}
	tradesPerMin := float64(snap.TradesCompleted) / uptime
	combatPerMin := float64(snap.CombatEncounters) / uptime
	queriesPerMin := float64(snap.DatabaseQueries) / uptime

	// Calculate error rate
	errorRate := 0.0
	if snap.DatabaseQueries > 0 {
		errorRate = (float64(snap.DatabaseErrors) / float64(snap.DatabaseQueries)) * 100
	}

	// Color coding for performance metrics
	dbP50Class := latencyClass(dbP50)
	dbP95Class := latencyClass(dbP95)
	dbP99Class := latencyClass(dbP99)
	tradeP50Class := latencyClass(tradeP50)
	tradeP95Class := latencyClass(tradeP95)
	tradeP99Class := latencyClass(tradeP99)
	combatP50Class := latencyClass(combatP50)
	combatP95Class := latencyClass(combatP95)
	combatP99Class := latencyClass(combatP99)

	cacheHitClass := "good"
	if snap.CacheHitRate < 50 {
		cacheHitClass = "warning"
	}
	if snap.CacheHitRate < 25 {
		cacheHitClass = "critical"
	}

	errorRateClass := "good"
	if errorRate > 1 {
		errorRateClass = "warning"
	}
	if errorRate > 5 {
		errorRateClass = "critical"
	}

	fmt.Fprintf(w, html,
		dbP50Class, dbP50.Round(time.Millisecond),
		dbP95Class, dbP95.Round(time.Millisecond),
		dbP99Class, dbP99.Round(time.Millisecond),
		tradeP50Class, tradeP50.Round(time.Millisecond),
		tradeP95Class, tradeP95.Round(time.Millisecond),
		tradeP99Class, tradeP99.Round(time.Millisecond),
		combatP50Class, combatP50.Round(time.Millisecond),
		combatP95Class, combatP95.Round(time.Millisecond),
		combatP99Class, combatP99.Round(time.Millisecond),
		tradesPerMin,
		combatPerMin,
		queriesPerMin,
		snap.ActiveConnections,
		snap.ActivePlayers,
		cacheHitClass, snap.CacheHitRate,
		errorRateClass, errorRate,
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)
}

// latencyClass returns a CSS class based on latency thresholds
func latencyClass(latency time.Duration) string {
	if latency < 50*time.Millisecond {
		return "good"
	} else if latency < 200*time.Millisecond {
		return "warning"
	}
	return "critical"
}
