// File: internal/server/server.go
// Project: Terminal Velocity
// Description: SSH server implementation with anonymous login and application-layer authentication
// Version: 2.0.1
// Author: Joshua Ferguson
// Created: 2025-01-07

package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/metrics"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/ratelimit"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var log = logger.WithComponent("server")

// Server represents the SSH game server
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
	metricsServer *metrics.Server
	rateLimiter   *ratelimit.Limiter
}

// Config holds server configuration
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

// NewServer creates a new game server
func NewServer(configFile string, port int) (*Server, error) {
	log.Debug("NewServer called with configFile=%s, port=%d", configFile, port)

	// TODO: Load config from file
	config := &Config{
		Host:        "0.0.0.0",
		Port:        port,
		HostKeyPath: "data/ssh_host_key", // Persistent SSH host key (writable in Docker)
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
		AllowPasswordAuth:  true,  // Allow password auth
		AllowPublicKeyAuth: false, // SSH key auth disabled - password only
		AllowRegistration:  true,  // Allow new user registration
		RequireEmail:       true,  // Require email for new accounts
		RequireEmailVerify: false, // Email verification not yet implemented
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

// initDatabase initializes the database connection
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

	log.Info("Database connected successfully")
	return nil
}

// Start starts the SSH server
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

// acceptConnections handles incoming SSH connections
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

// handleConnection handles a single SSH connection
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

// handleSession handles a single SSH session
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
	model := tui.NewModel(playerID, username, s.playerRepo, s.systemRepo, s.sshKeyRepo, s.shipRepo, s.marketRepo, s.mailRepo, s.socialRepo)

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

// startAnonymousSession starts an anonymous session (login screen)
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

// initSSHConfig initializes SSH server configuration
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

// shutdown gracefully shuts down the server
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

// PlayerSession represents an active player session
type PlayerSession struct {
	Username string
	Channel  ssh.Channel
}

// Close closes the player session
func (ps *PlayerSession) Close() {
	if ps.Channel != nil {
		if err := ps.Channel.Close(); err != nil {
			log.Warn("Failed to close SSH channel for user %s: %v", ps.Username, err)
		}
	}
}
