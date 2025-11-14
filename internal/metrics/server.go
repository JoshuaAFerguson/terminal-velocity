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
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("Metrics")

// Server provides an HTTP endpoint for Prometheus metrics
type Server struct {
	addr       string
	collector  *MetricsCollector
	httpServer *http.Server
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

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info("Starting metrics server on %s", s.addr)
	go func() {
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
	return s.httpServer.Shutdown(ctx)
}

// handleMetrics serves Prometheus-formatted metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprint(w, s.collector.PrometheusFormat())
}

// handleHealth serves a simple health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","uptime":"%v"}`, time.Since(s.collector.startTime))
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
