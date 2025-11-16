// File: internal/server/server.go
// Project: Terminal Velocity
// Description: SSH server implementation with anonymous login and application-layer authentication
// Version: 2.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package server implements the SSH game server for Terminal Velocity.
//
// Architecture Overview:
// The server uses an anonymous SSH authentication model where all SSH connections
// are accepted without credentials. Authentication is then handled at the application
// layer through an interactive login screen (TUI). This design allows for:
//   - Password-based authentication via login screen
//   - SSH public key authentication (optional)
//   - New user registration flow
//   - Flexible authentication without SSH protocol limitations
//
// Server Lifecycle:
//   1. Initialize configuration from YAML file or defaults
//   2. Connect to PostgreSQL database and initialize repositories
//   3. Load or generate persistent SSH host key
//   4. Start metrics server (if enabled)
//   5. Initialize rate limiter (if enabled)
//   6. Listen for SSH connections on configured port
//   7. Handle each connection in separate goroutine
//   8. Graceful shutdown on context cancellation
//
// Session Flow:
//   1. Client connects via SSH → accepted anonymously
//   2. Rate limiting checks (connection and IP limits)
//   3. SSH handshake with server host key
//   4. Start anonymous session with login screen (TUI)
//   5. User authenticates via application layer
//   6. Load player data and start game session
//   7. Run BubbleTea TUI over SSH channel
//   8. Handle disconnection and cleanup
//
// Security Features:
//   - Rate limiting per IP (connections and authentication attempts)
//   - Auto-banning after repeated failures
//   - Authentication lockout periods
//   - Secure password hashing (bcrypt)
//   - SSH public key fingerprint validation
//   - Connection tracking and metrics
//
// Thread Safety:
// The server manages concurrent connections using goroutines. Each connection
// runs independently with its own SSH channel and TUI instance. Shared resources
// (database, metrics, rate limiter) are protected by internal synchronization.
package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/fleet"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/friends"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/mail"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/marketplace"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/metrics"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/notifications"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/ratelimit"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

var log = logger.WithComponent("server")

// Server represents the SSH game server instance.
//
// The server manages SSH connections, player sessions, and game state.
// It uses anonymous SSH authentication with application-layer login,
// allowing flexible authentication methods (password, SSH keys, registration).
//
// Components:
//   - config: Server configuration from YAML file or defaults
//   - sshConfig: SSH protocol configuration (host key, auth handlers)
//   - listener: TCP listener for incoming SSH connections
//   - sessions: Active player sessions (currently unused, reserved for future)
//   - db: Database connection pool
//   - Repositories: Data access layers for different entities
//   - Managers: Business logic managers for game features
//   - metricsServer: Prometheus metrics HTTP server
//   - rateLimiter: Connection and authentication rate limiting
//
// Lifecycle:
//   - Created with NewServer() - initializes all dependencies
//   - Started with Start(ctx) - begins accepting connections
//   - Runs until context is cancelled
//   - Shutdown with shutdown() - cleans up resources
//
// Thread Safety:
// The Server struct itself is not protected by mutexes because it's typically
// used from a single goroutine. However, it spawns goroutines for each connection,
// and all shared resources (database, managers, metrics) have internal synchronization.
type Server struct {
	config        *Config
	port          int
	sshConfig     *ssh.ServerConfig
	listener      net.Listener
	sessions      map[string]*PlayerSession
	db            *database.DB
	playerRepo    *database.PlayerRepository
	systemRepo    *database.SystemRepository
	sshKeyRepo    *database.SSHKeyRepository
	shipRepo      *database.ShipRepository
	marketRepo    *database.MarketRepository
	mailRepo      *database.MailRepository
	socialRepo    *database.SocialRepository
	itemRepo      *database.ItemRepository
	metricsServer *metrics.Server
	rateLimiter   *ratelimit.Limiter

	// Managers
	fleetManager         *fleet.Manager
	mailManager          *mail.Manager
	notificationsManager *notifications.Manager
	friendsManager       *friends.Manager
	marketplaceManager   *marketplace.Manager
}

// Config holds server configuration loaded from YAML file or defaults.
//
// Configuration Sources:
//   1. Default values (defined in loadConfig)
//   2. YAML configuration file (optional, specified via -config flag)
//   3. Command-line flags (port, log level, etc.)
//
// Configuration files use YAML format with the following structure:
//   host: "0.0.0.0"
//   port: 2222
//   database:
//     host: "localhost"
//     port: 5432
//     user: "terminal_velocity"
//     password: "password"
//   metrics_enabled: true
//   rate_limit_enabled: true
//
// Fields are merged: file config overrides defaults, command-line flags override both.
//
// Security Considerations:
//   - Database password should not be committed to version control
//   - Use environment variables or separate config files for sensitive data
//   - Host key path should have restrictive permissions (0600)
type Config struct {
	Host        string
	Port        int
	DatabaseURL string
	HostKeyPath string
	MaxPlayers  int
	TickRate    int // Game loop ticks per second
	Database    *database.Config

	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    int

	// Rate limiting configuration
	RateLimitEnabled bool
	RateLimit        *ratelimit.Config

	// Authentication settings
	AllowPasswordAuth  bool // Allow password authentication
	AllowPublicKeyAuth bool // Allow SSH public key authentication
	AllowRegistration  bool // Allow new user registration
	RequireEmail       bool // Require email for new accounts
	RequireEmailVerify bool // Require email verification (future)
}

// loadConfig loads configuration from YAML file if it exists, otherwise uses defaults.
//
// Configuration Loading Strategy:
//   1. Start with hardcoded default values
//   2. If configFile is empty or doesn't exist, return defaults
//   3. If configFile exists, read and parse YAML
//   4. Merge file config with defaults (file takes precedence for non-zero values)
//   5. Command-line port overrides file config if specified
//
// Parameters:
//   - configFile: Path to YAML configuration file (empty string for defaults)
//   - port: Port number from command-line flag (used if file doesn't specify)
//
// Returns:
//   - *Config: Merged configuration with all fields populated
//   - error: File read or YAML parse error (missing file is not an error)
//
// Default Configuration:
//   - Host: "0.0.0.0" (listen on all interfaces)
//   - Port: 2222 (standard SSH alternate port)
//   - Database: localhost:5432 with terminal_velocity user
//   - Metrics: enabled on port 8080
//   - Rate limiting: enabled with conservative defaults
//   - Authentication: password auth enabled, registration enabled
//
// Error Handling:
// Missing config file is not an error (logs warning, uses defaults).
// Invalid YAML or unreadable file returns error.
func loadConfig(configFile string, port int) (*Config, error) {
	// Start with defaults
	config := &Config{
		Host:        "0.0.0.0",
		Port:        port,
		HostKeyPath: "data/ssh_host_key",
		MaxPlayers:  100,
		TickRate:    10,
		Database:    database.DefaultConfig(),

		// Metrics configuration
		MetricsEnabled: true,
		MetricsPort:    8080,

		// Rate limiting configuration
		RateLimitEnabled: true,
		RateLimit:        ratelimit.DefaultConfig(),

		// Default authentication settings
		AllowPasswordAuth:  true,
		AllowPublicKeyAuth: false,
		AllowRegistration:  true,
		RequireEmail:       true,
		RequireEmailVerify: false,
	}

	// If no config file specified or file doesn't exist, use defaults
	if configFile == "" {
		log.Info("No config file specified, using defaults")
		return config, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Warn("Config file %s not found, using defaults", configFile)
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Parse YAML
	var fileConfig Config
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// Merge file config with defaults (file config takes precedence for non-zero values)
	if fileConfig.Host != "" {
		config.Host = fileConfig.Host
	}
	if fileConfig.Port != 0 {
		config.Port = fileConfig.Port
	}
	if fileConfig.HostKeyPath != "" {
		config.HostKeyPath = fileConfig.HostKeyPath
	}
	if fileConfig.MaxPlayers != 0 {
		config.MaxPlayers = fileConfig.MaxPlayers
	}
	if fileConfig.TickRate != 0 {
		config.TickRate = fileConfig.TickRate
	}

	// Merge database config
	if fileConfig.Database.Host != "" {
		config.Database.Host = fileConfig.Database.Host
	}
	if fileConfig.Database.Port != 0 {
		config.Database.Port = fileConfig.Database.Port
	}
	if fileConfig.Database.User != "" {
		config.Database.User = fileConfig.Database.User
	}
	if fileConfig.Database.Password != "" {
		config.Database.Password = fileConfig.Database.Password
	}
	if fileConfig.Database.Database != "" {
		config.Database.Database = fileConfig.Database.Database
	}

	// Merge metrics config
	config.MetricsEnabled = fileConfig.MetricsEnabled
	if fileConfig.MetricsPort != 0 {
		config.MetricsPort = fileConfig.MetricsPort
	}

	// Merge rate limit config
	config.RateLimitEnabled = fileConfig.RateLimitEnabled

	// Merge auth settings
	config.AllowPasswordAuth = fileConfig.AllowPasswordAuth
	config.AllowPublicKeyAuth = fileConfig.AllowPublicKeyAuth
	config.AllowRegistration = fileConfig.AllowRegistration
	config.RequireEmail = fileConfig.RequireEmail
	config.RequireEmailVerify = fileConfig.RequireEmailVerify

	log.Info("Loaded configuration from %s", configFile)
	return config, nil
}

// NewServer creates a new game server instance with all dependencies initialized.
//
// Initialization Sequence:
//   1. Load configuration from YAML file or defaults
//   2. Initialize database connection pool
//   3. Create all repository instances (data access layer)
//   4. Create all manager instances (business logic)
//   5. Load or generate SSH host key
//   6. Start metrics server (if enabled)
//   7. Initialize rate limiter (if enabled)
//
// Parameters:
//   - configFile: Path to YAML configuration file (empty for defaults)
//   - port: SSH server port (overrides config file if non-zero)
//
// Returns:
//   - *Server: Fully initialized server ready to Start()
//   - error: Database connection, SSH key, or initialization error
//
// Resource Cleanup:
// If initialization fails partway through, already-created resources are cleaned up:
//   - Database connections closed
//   - Metrics server stopped
//   - Error returned to caller
//
// The server is NOT started by this function. Call Start(ctx) to begin accepting connections.
//
// Error Handling:
// All initialization errors are fatal and prevent server creation.
// Metrics server failures are non-fatal (logs warning, continues without metrics).
func NewServer(configFile string, port int) (*Server, error) {
	log.Debug("NewServer called with configFile=%s, port=%d", configFile, port)

	// Load configuration from file or use defaults
	config, err := loadConfig(configFile, port)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log.Info("Server configuration: host=%s, port=%d, maxPlayers=%d, tickRate=%d, metricsEnabled=%t, rateLimitEnabled=%t",
		config.Host, config.Port, config.MaxPlayers, config.TickRate, config.MetricsEnabled, config.RateLimitEnabled)

	srv := &Server{
		config:   config,
		port:     port,
		sessions: make(map[string]*PlayerSession),
	}

	// Initialize database
	log.Debug("Initializing database connection")
	if err := srv.initDatabase(); err != nil {
		log.Error("Failed to initialize database: %v", err)
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize SSH config
	log.Debug("Initializing SSH configuration")
	if err := srv.initSSHConfig(); err != nil {
		log.Error("Failed to initialize SSH config: %v", err)
		// Clean up database on error
		if closeErr := srv.db.Close(); closeErr != nil {
			log.Warn("Failed to close database during cleanup: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to init SSH config: %w", err)
	}

	// Initialize metrics
	if config.MetricsEnabled {
		log.Debug("Initializing metrics system")
		metricsCollector := metrics.Init()
		metricsAddr := fmt.Sprintf(":%d", config.MetricsPort)
		srv.metricsServer = metrics.NewServer(metricsAddr, metricsCollector)
		if err := srv.metricsServer.Start(); err != nil {
			log.Warn("Failed to start metrics server: %v", err)
			// Non-fatal, continue without metrics
		} else {
			log.Info("Metrics server started on port %d (endpoints: /metrics, /stats, /health)", config.MetricsPort)
		}
	}

	// Initialize rate limiter
	if config.RateLimitEnabled {
		log.Debug("Initializing rate limiter")
		srv.rateLimiter = ratelimit.NewLimiter(config.RateLimit)
		log.Info("Rate limiter enabled: maxConnPerIP=%d, maxAuthAttempts=%d, autoban=%d failures",
			config.RateLimit.MaxConnectionsPerIP, config.RateLimit.MaxAuthAttempts, config.RateLimit.AutobanThreshold)
	}

	log.Info("Server created successfully")
	return srv, nil
}

// initDatabase initializes the database connection pool and all data access components.
//
// Initialization Steps:
//   1. Connect to PostgreSQL using pgx connection pool
//   2. Create repository instances (PlayerRepository, SystemRepository, etc.)
//   3. Create manager instances (FleetManager, MailManager, etc.)
//   4. Start background workers for managers
//
// Repositories (Data Access Layer):
//   - PlayerRepository: Player accounts, authentication, online status
//   - SystemRepository: Star systems, planets, jump routes
//   - SSHKeyRepository: SSH public keys for authentication
//   - ShipRepository: Player ships and loadouts
//   - MarketRepository: Market prices and commodities
//   - MailRepository: Player mail messages
//   - SocialRepository: Friends, notifications, social features
//   - ItemRepository: Items and equipment
//
// Managers (Business Logic):
//   - FleetManager: Fleet operations and coordination
//   - MailManager: Mail delivery and notifications
//   - NotificationsManager: Real-time notifications (starts background worker)
//   - FriendsManager: Friend relationship management
//   - MarketplaceManager: Player marketplace (starts background worker)
//
// Connection Pool:
// Uses pgx/v5 connection pooling with configuration from database.Config.
// Pool size, timeouts, and other settings are managed by the database package.
//
// Error Handling:
// Database connection failures are fatal errors that prevent server startup.
// All cleanup is handled by the caller (NewServer).
func (s *Server) initDatabase() error {
	log.Debug("Connecting to database at %s:%d", s.config.Database.Host, s.config.Database.Port)

	var err error
	s.db, err = database.NewDB(s.config.Database)
	if err != nil {
		log.Error("Database connection failed: %v", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize repositories
	log.Debug("Initializing database repositories")
	s.playerRepo = database.NewPlayerRepository(s.db)
	s.systemRepo = database.NewSystemRepository(s.db)
	s.sshKeyRepo = database.NewSSHKeyRepository(s.db)
	s.shipRepo = database.NewShipRepository(s.db)
	s.marketRepo = database.NewMarketRepository(s.db)
	s.mailRepo = database.NewMailRepository(s.db)
	s.socialRepo = database.NewSocialRepository(s.db)
	s.itemRepo = database.NewItemRepository(s.db)

	// Initialize managers
	log.Debug("Initializing game managers")
	s.fleetManager = fleet.NewManager(s.playerRepo, s.shipRepo)
	s.mailManager = mail.NewManager(s.socialRepo)
	s.notificationsManager = notifications.NewManager(s.socialRepo)
	s.friendsManager = friends.NewManager(s.socialRepo)
	s.marketplaceManager = marketplace.NewManager(s.playerRepo, s.shipRepo)

	// Start background workers for managers
	s.fleetManager.Start()
	s.notificationsManager.Start()
	s.marketplaceManager.Start()

	log.Info("Database connected successfully")
	return nil
}

// Start starts the SSH server and begins accepting connections.
//
// Execution Flow:
//   1. Create TCP listener on configured host:port
//   2. Log server startup information
//   3. Spawn goroutine to accept connections (acceptConnections)
//   4. Block waiting for context cancellation
//   5. Graceful shutdown when context is cancelled
//
// Parameters:
//   - ctx: Context for cancellation (typically from signal handler)
//
// Returns:
//   - error: Listener creation failure or shutdown error
//
// Blocking Behavior:
// This function blocks until the context is cancelled (typically by SIGINT/SIGTERM).
// It does not return until the server is fully shut down.
//
// Graceful Shutdown:
// When context is cancelled:
//   1. Stop accepting new connections
//   2. Close all active sessions
//   3. Stop metrics server
//   4. Stop rate limiter
//   5. Close database connections
//
// Thread Safety:
// Safe to call only once. Calling Start() multiple times will fail.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Debug("Starting SSH server on %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Failed to listen on %s: %v", addr, err)
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	log.Info("SSH server listening on %s", addr)

	// Start accepting connections
	go s.acceptConnections(ctx)

	// Wait for context cancellation
	<-ctx.Done()

	log.Info("Context cancelled, shutting down server...")
	return s.shutdown()
}

// acceptConnections continuously accepts incoming SSH connections until context is cancelled.
//
// Execution Model:
// Runs in a separate goroutine, spawned by Start(). Loops indefinitely accepting
// connections and spawning a goroutine to handle each one.
//
// Flow:
//   1. Check context for cancellation
//   2. Accept next TCP connection (blocks until connection or error)
//   3. Log connection from remote address
//   4. Spawn goroutine to handleConnection
//   5. Repeat
//
// Context Cancellation:
// When context is cancelled, the loop exits cleanly. The listener is closed
// by shutdown(), which causes Accept() to return an error.
//
// Error Handling:
// Connection accept errors are logged as warnings but don't stop the loop.
// This handles transient network issues without bringing down the server.
//
// Thread Safety:
// Safe to call in goroutine. Each accepted connection is handled independently.
func (s *Server) acceptConnections(ctx context.Context) {
	log.Debug("Started accepting connections")
	for {
		select {
		case <-ctx.Done():
			log.Debug("Accept loop terminated due to context cancellation")
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Warn("Failed to accept connection: %v", err)
				continue
			}

			log.Debug("Accepted new connection from %s", conn.RemoteAddr())
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single SSH connection from initial TCP handshake to cleanup.
//
// Connection Lifecycle:
//   1. Track connection metrics (increment connection counter)
//   2. Check rate limits (per-IP connection limits, bans)
//   3. Perform SSH protocol handshake (anonymous auth)
//   4. Track active connection metrics
//   5. Discard out-of-band SSH requests
//   6. Handle SSH channels (session channels only)
//   7. Clean up on disconnection
//
// Rate Limiting (Security Layer 1):
// Before SSH handshake, check:
//   - IP not banned (auto-ban after repeated failures)
//   - Concurrent connections per IP < limit (default: 5)
//   - Connection rate per IP < limit (default: 20/minute)
//
// If rate limit exceeded, connection is rejected with error message.
//
// SSH Handshake (Security Layer 2):
// Uses anonymous authentication (NoClientAuth=true), meaning:
//   - All SSH connections are accepted at protocol level
//   - No SSH username/password validation
//   - No SSH public key validation at this layer
//   - Server presents host key for client verification
//
// Authentication is deferred to application layer (login screen in TUI).
//
// Channel Handling:
// Only "session" channels are accepted. Other channel types (port forwarding,
// X11, etc.) are rejected for security.
//
// Metrics:
//   - Connection attempts (total)
//   - Active connections (incremented/decremented)
//   - Failed connections (rate limited or SSH handshake failed)
//   - Connection duration (tracked on cleanup)
//   - Login events
//
// Thread Safety:
// Each connection runs in its own goroutine. No shared state except:
//   - Metrics (internally synchronized)
//   - Rate limiter (internally synchronized)
//   - Database (connection pool)
//
// Resource Cleanup:
// Ensures connection and SSH channel are closed via defer, even on panic.
// Connection duration is tracked and recorded to metrics.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr()
	log.Debug("handleConnection called for %s", remoteAddr)

	// Track connection attempt
	metrics.Global().IncrementConnections()
	connStart := time.Now()

	// Check rate limit
	if s.rateLimiter != nil {
		if allowed, reason := s.rateLimiter.AllowConnection(remoteAddr); !allowed {
			log.Warn("Connection rejected from %s: %s", remoteAddr, reason)
			metrics.Global().IncrementFailedConnections()
			// Send rejection message and close
			if _, err := conn.Write([]byte("Connection rejected: " + reason + "\r\n")); err != nil {
				log.Warn("Failed to write rejection message to %s: %v", remoteAddr, err)
			}
			return
		}
		defer s.rateLimiter.ReleaseConnection(remoteAddr)
	}

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		log.Warn("SSH handshake failed from %s: %v", remoteAddr, err)
		metrics.Global().IncrementFailedConnections()
		return
	}
	defer func() {
		if err := sshConn.Close(); err != nil {
			log.Warn("Failed to close SSH connection from %s: %v", remoteAddr, err)
		}
		// Track connection duration
		metrics.Global().RecordConnectionDuration(time.Since(connStart))
	}()

	log.Info("SSH connection established: user=%s, addr=%s", sshConn.User(), sshConn.RemoteAddr())
	metrics.Global().IncrementActiveConnections()
	metrics.Global().IncrementLogins()
	defer metrics.Global().DecrementActiveConnections()

	// Discard all global out-of-band requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		channelType := newChannel.ChannelType()
		if channelType != "session" {
			log.Debug("Rejecting unknown channel type %s from %s", channelType, sshConn.User())
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Error("Could not accept channel from %s: %v", sshConn.User(), err)
			continue
		}

		log.Debug("Accepted session channel for %s", sshConn.User())
		// Handle this session
		go s.handleSession(sshConn.User(), sshConn.Permissions, channel, requests)
	}

	log.Debug("SSH connection closed for %s", sshConn.User())
}

// handleSession handles a single SSH session channel (pty and shell requests).
//
// SSH Session Protocol:
// An SSH session consists of a series of requests from the client:
//   1. "pty-req" - Request pseudo-terminal allocation (for TUI)
//   2. "shell" - Request shell execution (where we launch the game)
//   3. Other requests - Window size changes, signals, etc. (rejected)
//
// Flow:
//   1. Loop reading SSH requests on channel
//   2. Accept "pty-req" (needed for BubbleTea TUI)
//   3. Accept "shell" and launch anonymous session (login screen)
//   4. Return after shell session ends
//
// Anonymous Session:
// Since SSH authentication is anonymous, all shells start with the login screen.
// Players authenticate at application layer (via TUI) and then proceed to game.
//
// Parameters:
//   - username: SSH username (from SSH handshake, unused in anonymous mode)
//   - perms: SSH permissions (from auth callback, unused in anonymous mode)
//   - channel: SSH channel for I/O (will be used by BubbleTea)
//   - requests: Channel of SSH requests (pty-req, shell, etc.)
//
// Request Handling:
//   - "pty-req": Accepted (BubbleTea needs a PTY for TUI rendering)
//   - "shell": Accepted and starts anonymous session
//   - All others: Rejected
//
// Session Lifecycle:
// Once "shell" is accepted, control transfers to startAnonymousSession which
// runs the BubbleTea login screen. This function returns when TUI exits.
//
// Thread Safety:
// Each session runs in its own goroutine. SSH channel provides I/O synchronization.
func (s *Server) handleSession(username string, perms *ssh.Permissions, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Handle session requests (pty-req, shell, etc.)
	for req := range requests {
		switch req.Type {
		case "pty-req":
			req.Reply(true, nil)
		case "shell":
			req.Reply(true, nil)
			// Start anonymous session (login screen)
			s.startAnonymousSession(channel)
			return
		default:
			req.Reply(false, nil)
		}
	}
}

// startGameSession starts a game session for a player
func (s *Server) startGameSession(username string, perms *ssh.Permissions, channel ssh.Channel) {
	log.Debug("startGameSession called for user=%s", username)
	ctx := context.Background()

	// Get player ID from permissions
	var playerID uuid.UUID
	if perms != nil && perms.Extensions != nil {
		if playerIDStr, ok := perms.Extensions["player_id"]; ok {
			var err error
			playerID, err = uuid.Parse(playerIDStr)
			if err != nil {
				log.Error("Invalid player ID in permissions for %s: %v", username, err)
				if _, writeErr := channel.Write([]byte("Error: Invalid session. Please reconnect.\r\n")); writeErr != nil {
					log.Warn("Failed to write error message to %s: %v", username, writeErr)
				}
				return
			}
			log.Debug("Player ID from permissions: %s", playerID)
		}
	}

	// If no player ID, this might be a new user registration flow
	if playerID == uuid.Nil {
		log.Debug("No player ID in permissions, checking if player exists: %s", username)
		// Check if player exists
		player, err := s.playerRepo.GetByUsername(ctx, username)
		if err == database.ErrPlayerNotFound && s.config.AllowRegistration {
			log.Info("Starting registration flow for new user: %s", username)
			// Start registration flow
			s.startRegistrationSession(username, channel)
			return
		} else if err != nil {
			log.Error("Error checking for player %s: %v", username, err)
			if _, writeErr := channel.Write([]byte("Error: Authentication failed. Please reconnect.\r\n")); writeErr != nil {
				log.Warn("Failed to write error message to %s: %v", username, writeErr)
			}
			return
		}
		playerID = player.ID
		log.Debug("Found existing player ID: %s", playerID)
	}

	log.Info("Starting game session for user=%s, playerID=%s", username, playerID)

	// Initialize TUI model
	model := tui.NewModel(
		playerID,
		username,
		s.playerRepo,
		s.systemRepo,
		s.sshKeyRepo,
		s.shipRepo,
		s.marketRepo,
		s.mailRepo,
		s.socialRepo,
		s.itemRepo,
		s.fleetManager,
		s.mailManager,
		s.notificationsManager,
		s.friendsManager,
		s.marketplaceManager,
	)

	// Create BubbleTea program with SSH channel as input/output
	p := tea.NewProgram(
		model,
		tea.WithInput(channel),
		tea.WithOutput(channel),
		tea.WithAltScreen(), // Use alternate screen buffer to prevent artifacts
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Error("Error running TUI for %s: %v", username, err)
	}

	log.Info("Game session ended for user=%s, playerID=%s", username, playerID)

	// Mark player as offline
	if err := s.playerRepo.SetOnlineStatus(ctx, playerID, false); err != nil {
		log.Warn("Failed to set offline status for %s: %v", username, err)
	}
}

// startAnonymousSession starts an anonymous session with the login screen TUI.
//
// This is the entry point for all SSH connections. Since SSH authentication is
// anonymous, every connection starts here regardless of credentials.
//
// Login Flow:
//   1. Create TUI model with login screen
//   2. Run BubbleTea program over SSH channel
//   3. User sees login prompts (username/password)
//   4. TUI handles authentication via PlayerRepository
//   5. On successful auth, TUI transitions to game screen
//   6. On failed auth, user can retry or disconnect
//
// TUI Integration:
//   - Input: SSH channel (reads keypresses from client)
//   - Output: SSH channel (sends terminal output to client)
//   - AltScreen: Enabled (uses alternate screen buffer for clean rendering)
//
// The login model (NewLoginModel) is a minimal TUI that only handles authentication.
// After successful login, it's replaced by the full game model.
//
// Parameters:
//   - channel: SSH channel for I/O with the client
//
// Blocking Behavior:
// This function blocks until the TUI exits (user quits or disconnects).
//
// Error Handling:
// TUI errors are logged but not returned. The session simply ends.
func (s *Server) startAnonymousSession(channel ssh.Channel) {
	log.Debug("startAnonymousSession called")

	// Initialize TUI model with login screen
	model := tui.NewLoginModel(s.playerRepo, s.systemRepo, s.sshKeyRepo, s.shipRepo, s.marketRepo, s.mailRepo, s.socialRepo)

	// Create BubbleTea program with SSH channel as input/output
	p := tea.NewProgram(
		model,
		tea.WithInput(channel),
		tea.WithOutput(channel),
		tea.WithAltScreen(), // Use alternate screen buffer to prevent artifacts
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Info("Error running login TUI: %v", err)
	}

	log.Info("Anonymous session ended")
}

// startRegistrationSession starts a registration session for a new player
func (s *Server) startRegistrationSession(username string, channel ssh.Channel) {
	// Initialize TUI model for registration
	model := tui.NewRegistrationModel(username, s.config.RequireEmail, nil, s.playerRepo, s.systemRepo, s.sshKeyRepo, s.shipRepo, s.marketRepo)

	// Create BubbleTea program with SSH channel as input/output
	p := tea.NewProgram(
		model,
		tea.WithInput(channel),
		tea.WithOutput(channel),
		tea.WithAltScreen(), // Use alternate screen buffer to prevent artifacts
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Info("Error running registration TUI for %s: %v", username, err)
	}

	log.Info("Registration session ended for %s", username)
}

// initSSHConfig initializes SSH server configuration with anonymous authentication.
//
// SSH Configuration Strategy:
// This server uses an unconventional SSH setup called "anonymous authentication":
//   - NoClientAuth is enabled (skips SSH-level authentication)
//   - No password callback (would validate SSH passwords)
//   - No public key callback (would validate SSH keys)
//   - Host key is loaded/generated for server identity
//
// Why Anonymous SSH?
// Traditional SSH authentication happens at protocol level before the client
// can interact. This doesn't work well for:
//   - Interactive registration (need to show prompts before credentials exist)
//   - Multiple auth methods (password OR SSH key OR registration)
//   - Application-specific auth logic (reputation checks, bans, etc.)
//
// Instead, we:
//   1. Accept all SSH connections (NoClientAuth=true)
//   2. Present login screen via TUI
//   3. Handle authentication at application layer
//   4. Provide rich feedback and interactivity
//
// Security Implications:
// Anonymous SSH means ANYONE can connect and reach the login screen. Security depends on:
//   - Rate limiting (connections per IP, auth attempts)
//   - Auto-banning after repeated failures
//   - Strong password hashing (bcrypt)
//   - Application-layer authentication validation
//
// Host Key:
// The server's SSH host key is loaded from disk (or generated if missing).
// This key:
//   - Identifies the server to clients
//   - Prevents man-in-the-middle attacks
//   - Is persisted across restarts (same fingerprint)
//   - Uses ED25519 algorithm (modern, secure)
//
// Returns:
//   - error: Host key loading/generation error (fatal)
func (s *Server) initSSHConfig() error {
	s.sshConfig = &ssh.ServerConfig{}

	// Anonymous authentication - accept all connections
	// Authentication is handled at the application layer (login screen)
	s.sshConfig.NoClientAuth = true

	// Load or generate persistent SSH host key
	log.Debug("Loading SSH host key from: %s", s.config.HostKeyPath)
	privateKey, err := loadOrGenerateHostKey(s.config.HostKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load/generate host key: %w", err)
	}

	s.sshConfig.AddHostKey(privateKey)
	log.Info("SSH authentication: anonymous (authentication via login screen)")
	log.Info("SSH host key fingerprint: %s", ssh.FingerprintSHA256(privateKey.PublicKey()))
	return nil
}

// handlePasswordAuth handles password-based authentication
func (s *Server) handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	remoteAddr := conn.RemoteAddr()
	log.Info("Password login attempt: %s from %s", username, remoteAddr)

	ctx := context.Background()

	// Check if authentication is locked for this IP
	if s.rateLimiter != nil {
		if locked, remaining := s.rateLimiter.IsAuthLocked(remoteAddr); locked {
			log.Warn("Auth attempt from locked IP %s (user: %s, remaining: %v)", remoteAddr, username, remaining)
			return nil, fmt.Errorf("too many failed attempts, try again in %v", remaining.Round(time.Second))
		}
	}

	// Try to authenticate
	player, err := s.playerRepo.Authenticate(ctx, username, string(password))
	if err != nil {
		// Record auth failure for rate limiting
		if s.rateLimiter != nil {
			s.rateLimiter.RecordAuthFailure(remoteAddr, username)
		}

		if err == database.ErrInvalidCredentials {
			// Check if user exists
			existingPlayer, checkErr := s.playerRepo.GetByUsername(ctx, username)
			if checkErr == database.ErrPlayerNotFound {
				// User doesn't exist - offer registration if enabled
				if s.config.AllowRegistration {
					return s.handleNewUserRegistration(ctx, conn, string(password))
				}
				log.Info("Failed login - user not found: %s", username)
				return nil, fmt.Errorf("invalid username or password")
			}

			// User exists but SSH-key-only
			if existingPlayer != nil && existingPlayer.PasswordHash == "" {
				log.Info("Failed login - SSH key required for: %s", username)
				return nil, fmt.Errorf("this account requires SSH key authentication")
			}

			log.Info("Failed login - invalid password: %s", username)
			return nil, fmt.Errorf("invalid username or password")
		}
		log.Info("Authentication error for %s: %v", username, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Successful authentication - record success and clear failures
	if s.rateLimiter != nil {
		s.rateLimiter.RecordAuthSuccess(remoteAddr)
	}

	return s.onSuccessfulAuth(ctx, player)
}

// handlePublicKeyAuth handles SSH public key authentication
func (s *Server) handlePublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	remoteAddr := conn.RemoteAddr()
	log.Info("Public key login attempt: %s from %s", username, remoteAddr)

	ctx := context.Background()

	// Check if authentication is locked for this IP
	if s.rateLimiter != nil {
		if locked, remaining := s.rateLimiter.IsAuthLocked(remoteAddr); locked {
			log.Warn("Auth attempt from locked IP %s (user: %s, remaining: %v)", remoteAddr, username, remaining)
			return nil, fmt.Errorf("too many failed attempts, try again in %v", remaining.Round(time.Second))
		}
	}

	// Get the public key in authorized_keys format
	keyData := ssh.MarshalAuthorizedKey(key)

	// Try to find the player by public key
	playerID, err := s.sshKeyRepo.GetPlayerIDByPublicKey(ctx, keyData)
	if err != nil {
		// Record auth failure for rate limiting
		if s.rateLimiter != nil {
			s.rateLimiter.RecordAuthFailure(remoteAddr, username)
		}

		if err == database.ErrSSHKeyNotFound {
			// Key not found - check if user exists
			player, checkErr := s.playerRepo.GetByUsername(ctx, username)
			if checkErr == database.ErrPlayerNotFound {
				// New user with SSH key - offer registration if enabled
				if s.config.AllowRegistration {
					return s.handleNewUserSSHRegistration(ctx, conn, keyData)
				}
				log.Info("Failed SSH key login - user not found: %s", username)
				return nil, fmt.Errorf("public key not authorized")
			}

			// User exists but key not registered
			log.Info("Failed SSH key login - key not registered for: %s (ID: %s)", username, player.ID)
			return nil, fmt.Errorf("public key not authorized for this user")
		}
		log.Info("SSH key authentication error for %s: %v", username, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Get player by ID
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		// Record auth failure for rate limiting
		if s.rateLimiter != nil {
			s.rateLimiter.RecordAuthFailure(remoteAddr, username)
		}
		log.Info("Failed to get player %s: %v", playerID, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Successful authentication - record success and clear failures
	if s.rateLimiter != nil {
		s.rateLimiter.RecordAuthSuccess(remoteAddr)
	}

	// Verify username matches
	if player.Username != username {
		log.Info("SSH key login - username mismatch: %s vs %s", username, player.Username)
		return nil, fmt.Errorf("username does not match public key")
	}

	// Update last used timestamp for the key
	go func() {
		fingerprint := ssh.FingerprintSHA256(key)
		if sshKey, err := s.sshKeyRepo.GetKeyByFingerprint(context.Background(), fingerprint); err == nil {
			s.sshKeyRepo.UpdateLastUsed(context.Background(), sshKey.ID)
		}
	}()

	// Successful authentication
	return s.onSuccessfulAuth(ctx, player)
}

// onSuccessfulAuth handles post-authentication tasks
func (s *Server) onSuccessfulAuth(ctx context.Context, player *models.Player) (*ssh.Permissions, error) {
	// Update last login and set online status
	go func() {
		ctx := context.Background()
		s.playerRepo.UpdateLastLogin(ctx, player.ID)
		s.playerRepo.SetOnlineStatus(ctx, player.ID, true)
	}()

	log.Info("Successful login: %s (ID: %s)", player.Username, player.ID)

	// Return permissions with player ID
	return &ssh.Permissions{
		Extensions: map[string]string{
			"player_id": player.ID.String(),
			"username":  player.Username,
		},
	}, nil
}

// handleNewUserRegistration handles registration for a new user with password
func (s *Server) handleNewUserRegistration(ctx context.Context, conn ssh.ConnMetadata, password string) (*ssh.Permissions, error) {
	username := conn.User()

	// For now, we'll just reject and log
	// In a full implementation, this would:
	// 1. Send a prompt to the user asking for email
	// 2. Wait for email input
	// 3. Validate email
	// 4. Create the account
	// 5. Authenticate the user

	// This requires interactive SSH session handling which we'll implement later
	log.Info("New user registration requested: %s (not yet implemented)", username)
	return nil, fmt.Errorf("account not found. Contact administrator to create an account")
}

// handleNewUserSSHRegistration handles registration for a new user with SSH key
func (s *Server) handleNewUserSSHRegistration(ctx context.Context, conn ssh.ConnMetadata, keyData []byte) (*ssh.Permissions, error) {
	username := conn.User()

	// Similar to password registration, this needs interactive handling
	log.Info("New user SSH key registration requested: %s (not yet implemented)", username)
	return nil, fmt.Errorf("account not found. Contact administrator to create an account")
}

// shutdown gracefully shuts down the server and cleans up all resources.
//
// Shutdown Sequence:
//   1. Close TCP listener (stops accepting new connections)
//   2. Close all active player sessions (disconnects clients)
//   3. Stop metrics server (with 5 second timeout)
//   4. Stop rate limiter (cleanup background workers)
//   5. Close database connection pool
//
// Graceful vs Forced:
// Active sessions are closed immediately (not graceful for clients).
// For truly graceful shutdown, would need to:
//   - Signal sessions to save and exit
//   - Wait for sessions to close voluntarily
//   - Force close after timeout
//
// Current implementation prioritizes server shutdown speed over client experience.
//
// Timeout Handling:
// Metrics server shutdown has 5 second timeout. If it doesn't stop cleanly,
// the error is logged but shutdown continues (non-fatal).
//
// Error Handling:
// Errors during shutdown are logged but don't stop the shutdown process.
// All cleanup steps are attempted regardless of individual failures.
//
// Returns:
//   - error: Always returns nil (errors are logged, not returned)
func (s *Server) shutdown() error {
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all active sessions
	for _, session := range s.sessions {
		session.Close()
	}

	// Shutdown metrics server
	if s.metricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.metricsServer.Stop(ctx); err != nil {
			log.Warn("Error stopping metrics server: %v", err)
		} else {
			log.Info("Metrics server stopped")
		}
	}

	// Shutdown rate limiter
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
		log.Info("Rate limiter stopped")
	}

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Info("Error closing database: %v", err)
		} else {
			log.Info("Database connection closed")
		}
	}

	log.Info("Server shutdown complete")
	return nil
}

// PlayerSession represents an active player session (currently unused).
//
// Design Note:
// This struct is defined but not currently used by the server. It's reserved
// for future functionality where we might need to track active sessions for:
//   - Broadcasting server messages to all players
//   - Graceful shutdown (notify players before disconnecting)
//   - Session management (kick players, view active sessions)
//   - Metrics (track session durations, activity)
//
// Current Implementation:
// Sessions are tracked in Server.sessions map but never populated.
// Connection handling doesn't register sessions in this map.
//
// Future Enhancement:
// To use this, modify handleConnection to:
//   1. Create PlayerSession after successful authentication
//   2. Register in Server.sessions map (with mutex protection)
//   3. Remove from map on disconnection
//   4. Add methods for session management operations
type PlayerSession struct {
	Username string      // Player's username
	Channel  ssh.Channel // SSH channel for communication
}

// Close closes the player session cleanly.
//
// This method closes the SSH channel which:
//   - Sends EOF to client
//   - Closes write side of channel
//   - Triggers client disconnection
//
// The BubbleTea program running on this channel will receive an error and exit.
//
// Error Handling:
// Channel close errors are logged but not returned. Common errors:
//   - Channel already closed
//   - Network connection lost
func (ps *PlayerSession) Close() {
	if ps.Channel != nil {
		if err := ps.Channel.Close(); err != nil {
			log.Warn("Failed to close SSH channel for user %s: %v", ps.Username, err)
		}
	}
}
