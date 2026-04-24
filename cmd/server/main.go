// File: cmd/server/main.go
// Project: Terminal Velocity
// Description: Main SSH game server entry point
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package main provides the entry point for the Terminal Velocity SSH game server.
//
// Server Overview:
// This binary starts the Terminal Velocity multiplayer space trading game server.
// Players connect via SSH and play through a terminal user interface (TUI).
//
// Execution Flow:
//   1. Parse command-line flags (config file, port, logging)
//   2. Initialize logging system
//   3. Set up signal handling for graceful shutdown
//   4. Create and initialize server
//   5. Start server (blocks until shutdown signal)
//   6. Clean shutdown on SIGINT/SIGTERM
//
// Command-Line Flags:
//   -config <file>     Path to YAML configuration file (default: configs/config.yaml)
//   -port <number>     SSH server port (default: 2222)
//   -log-level <level> Logging verbosity: debug, info, warn, error (default: info)
//   -log-file <path>   Log file path, empty for stdout only (default: stdout)
//   -version           Show version information and exit
//
// Example Usage:
//   # Start with defaults (port 2222, stdout logging)
//   ./server
//
//   # Start with custom config
//   ./server -config /etc/terminal-velocity/config.yaml
//
//   # Start with debug logging to file
//   ./server -log-level debug -log-file /var/log/terminal-velocity.log
//
//   # Check version
//   ./server -version
//
// Configuration:
// Server reads configuration from YAML file with fallback to defaults:
//   - Database connection (host, port, credentials)
//   - Server settings (host, port, max players)
//   - Metrics (enabled, port)
//   - Rate limiting (enabled, thresholds)
//   - Authentication (password, SSH keys, registration)
//
// Logging:
// Structured logging to stdout and/or file with:
//   - Log levels (debug, info, warn, error)
//   - Component tags (server, database, tui, etc.)
//   - Caller information (file:line)
//   - Timestamp
//
// Graceful Shutdown:
// SIGINT (Ctrl+C) or SIGTERM triggers graceful shutdown:
//   1. Stop accepting new connections
//   2. Close active player sessions
//   3. Stop metrics server
//   4. Close database connections
//   5. Flush logs
//   6. Exit cleanly
//
// Version Information:
// Version, commit hash, and build date are set at build time via ldflags:
//   go build -ldflags "-X main.version=v1.0.0 -X main.commit=$(git rev-parse HEAD)"
//
// Exit Codes:
//   0 - Clean shutdown
//   1 - Initialization error (logger, server creation, startup failure)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/server"
)

var (
	// Version information (set during build via ldflags)
	// Example build command:
	//   go build -ldflags "-X main.version=v1.0.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Component logger for main package
	log = logger.WithComponent("main")
)

// main is the entry point for the Terminal Velocity server.
//
// Initialization Steps:
//   1. Parse command-line flags
//   2. Initialize logger (file and/or stdout)
//   3. Check for version flag (exit if set)
//   4. Set up context for cancellation
//   5. Set up signal handler (SIGINT, SIGTERM)
//   6. Create server instance
//   7. Start server (blocks)
//   8. Clean up on shutdown
//
// Signal Handling:
// A goroutine listens for OS signals (SIGINT from Ctrl+C, SIGTERM from kill).
// When received:
//   - Logs shutdown message
//   - Cancels context
//   - Server.Start() returns
//   - Deferred cleanup runs
//   - Process exits
//
// Error Handling:
// Fatal errors (logger init, server creation, start failure) are logged
// to stderr and cause immediate exit with code 1. Non-fatal errors are
// logged but allow server to continue.
//
// Resource Cleanup:
// Deferred functions ensure proper cleanup:
//   - Logger flushed and closed
//   - Context cancelled
//   - Server shutdown (connections, database, metrics)
func main() {
	// Parse command line flags
	var (
		configFile  = flag.String("config", "configs/config.yaml", "Path to configuration file")
		showVersion = flag.Bool("version", false, "Show version information")
		port        = flag.Int("port", 2222, "SSH server port")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		logFile     = flag.String("log-file", "", "Log file path (empty for stdout only)")
	)
	flag.Parse()

	// Initialize logger
	logCfg := logger.Config{
		Level:      *logLevel,
		FilePath:   *logFile,
		ToStdout:   true,
		WithCaller: true,
	}
	if err := logger.Init(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	log.Info("Terminal Velocity starting up")
	log.Debug("Command line args: config=%s, port=%d, log-level=%s", *configFile, *port, *logLevel)

	// Show version if requested
	if *showVersion {
		fmt.Printf("Terminal Velocity %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Shutdown signal received, gracefully shutting down...")
		cancel()
	}()

	// Initialize and start server
	log.Info("Starting Terminal Velocity server v%s (commit: %s, built: %s)", version, commit, date)

	srv, err := server.NewServer(*configFile, *port)
	if err != nil {
		log.Fatal("Failed to create server: %v", err)
	}

	// Run server
	log.Info("Server initialized successfully, starting main loop")
	if err := srv.Start(ctx); err != nil {
		log.Fatal("Server error: %v", err)
	}

	log.Info("Server shutdown complete")
}
