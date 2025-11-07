package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0v3r1gn/terminal-velocity/internal/server"
)

var (
	// Version information (set during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	var (
		configFile  = flag.String("config", "configs/config.yaml", "Path to configuration file")
		showVersion = flag.Bool("version", false, "Show version information")
		port        = flag.Int("port", 2222, "SSH server port")
	)
	flag.Parse()

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
		log.Println("Shutdown signal received, gracefully shutting down...")
		cancel()
	}()

	// Initialize and start server
	log.Printf("Starting Terminal Velocity server v%s...\n", version)

	srv, err := server.NewServer(*configFile, *port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Run server
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}
