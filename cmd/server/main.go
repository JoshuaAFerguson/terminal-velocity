// File: cmd/server/main.go
// Project: Terminal Velocity
// Description: Main SSH game server entry point
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

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
	// Version information (set during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"

	log = logger.WithComponent("main")
)

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
