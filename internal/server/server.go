package server

import (
	"context"
	"fmt"
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/s0v3r1gn/terminal-velocity/internal/database"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
	"github.com/s0v3r1gn/terminal-velocity/internal/tui"
	"golang.org/x/crypto/ssh"
)

// Server represents the SSH game server
type Server struct {
	config     *Config
	port       int
	sshConfig  *ssh.ServerConfig
	listener   net.Listener
	sessions   map[string]*PlayerSession
	db         *database.DB
	playerRepo *database.PlayerRepository
	systemRepo *database.SystemRepository
	sshKeyRepo *database.SSHKeyRepository
}

// Config holds server configuration
type Config struct {
	Host           string
	Port           int
	DatabaseURL    string
	HostKeyPath    string
	MaxPlayers     int
	TickRate       int // Game loop ticks per second
	Database       *database.Config

	// Authentication settings
	AllowPasswordAuth     bool // Allow password authentication
	AllowPublicKeyAuth    bool // Allow SSH public key authentication
	AllowRegistration     bool // Allow new user registration
	RequireEmail          bool // Require email for new accounts
	RequireEmailVerify    bool // Require email verification (future)
}

// NewServer creates a new game server
func NewServer(configFile string, port int) (*Server, error) {
	// TODO: Load config from file
	config := &Config{
		Host:       "0.0.0.0",
		Port:       port,
		MaxPlayers: 100,
		TickRate:   10,
		Database:   database.DefaultConfig(),

		// Default authentication settings
		AllowPasswordAuth:  true,  // Allow password auth
		AllowPublicKeyAuth: true,  // Allow SSH key auth
		AllowRegistration:  true,  // Allow new user registration
		RequireEmail:       true,  // Require email for new accounts
		RequireEmailVerify: false, // Email verification not yet implemented
	}

	srv := &Server{
		config:   config,
		port:     port,
		sessions: make(map[string]*PlayerSession),
	}

	// Initialize database
	if err := srv.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize SSH config
	if err := srv.initSSHConfig(); err != nil {
		srv.db.Close() // Clean up database on error
		return nil, fmt.Errorf("failed to init SSH config: %w", err)
	}

	return srv, nil
}

// initDatabase initializes the database connection
func (s *Server) initDatabase() error {
	var err error
	s.db, err = database.NewDB(s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize repositories
	s.playerRepo = database.NewPlayerRepository(s.db)
	s.systemRepo = database.NewSystemRepository(s.db)
	s.sshKeyRepo = database.NewSSHKeyRepository(s.db)

	log.Println("Database connected successfully")
	return nil
}

// Start starts the SSH server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	log.Printf("SSH server listening on %s", addr)

	// Start accepting connections
	go s.acceptConnections(ctx)

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Shutting down server...")
	return s.shutdown()
}

// acceptConnections handles incoming SSH connections
func (s *Server) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}

			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single SSH connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.User())

	// Discard all global out-of-band requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		// Handle this session
		go s.handleSession(sshConn.User(), sshConn.Permissions, channel, requests)
	}
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
			// Start game session
			s.startGameSession(username, perms, channel)
			return
		default:
			req.Reply(false, nil)
		}
	}
}

// startGameSession starts a game session for a player
func (s *Server) startGameSession(username string, perms *ssh.Permissions, channel ssh.Channel) {
	ctx := context.Background()

	// Get player ID from permissions
	var playerID uuid.UUID
	if perms != nil && perms.Extensions != nil {
		if playerIDStr, ok := perms.Extensions["player_id"]; ok {
			var err error
			playerID, err = uuid.Parse(playerIDStr)
			if err != nil {
				log.Printf("Invalid player ID in permissions: %v", err)
				channel.Write([]byte("Error: Invalid session. Please reconnect.\r\n"))
				return
			}
		}
	}

	// If no player ID, this might be a new user registration flow
	if playerID == uuid.Nil {
		// Check if player exists
		player, err := s.playerRepo.GetByUsername(ctx, username)
		if err == database.ErrPlayerNotFound && s.config.AllowRegistration {
			// Start registration flow
			s.startRegistrationSession(username, channel)
			return
		} else if err != nil {
			log.Printf("Error checking for player %s: %v", username, err)
			channel.Write([]byte("Error: Authentication failed. Please reconnect.\r\n"))
			return
		}
		playerID = player.ID
	}

	// Initialize TUI model
	model := tui.NewModel(playerID, username, s.playerRepo, s.systemRepo, s.sshKeyRepo)

	// Create BubbleTea program with SSH channel as input/output
	p := tea.NewProgram(
		model,
		tea.WithInput(channel),
		tea.WithOutput(channel),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Printf("Error running TUI for %s: %v", username, err)
	}

	log.Printf("Game session ended for %s", username)

	// Mark player as offline
	s.playerRepo.SetOnlineStatus(ctx, playerID, false)
}

// startRegistrationSession starts a registration session for a new player
func (s *Server) startRegistrationSession(username string, channel ssh.Channel) {
	// Initialize TUI model for registration
	model := tui.NewRegistrationModel(username, s.config.RequireEmail, nil, s.playerRepo, s.systemRepo, s.sshKeyRepo)

	// Create BubbleTea program with SSH channel as input/output
	p := tea.NewProgram(
		model,
		tea.WithInput(channel),
		tea.WithOutput(channel),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Printf("Error running registration TUI for %s: %v", username, err)
	}

	log.Printf("Registration session ended for %s", username)
}

// initSSHConfig initializes SSH server configuration
func (s *Server) initSSHConfig() error {
	s.sshConfig = &ssh.ServerConfig{}

	// Password authentication callback
	if s.config.AllowPasswordAuth {
		s.sshConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return s.handlePasswordAuth(conn, password)
		}
	}

	// Public key authentication callback
	if s.config.AllowPublicKeyAuth {
		s.sshConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return s.handlePublicKeyAuth(conn, key)
		}
	}

	// If no auth methods enabled, return error
	if !s.config.AllowPasswordAuth && !s.config.AllowPublicKeyAuth {
		return fmt.Errorf("no authentication methods enabled")
	}

	// Generate a temporary host key (TODO: load from file)
	// For development only - in production, use a persistent key
	privateKey, err := generateHostKey()
	if err != nil {
		return fmt.Errorf("failed to generate host key: %w", err)
	}

	s.sshConfig.AddHostKey(privateKey)
	log.Printf("SSH authentication methods: password=%v publickey=%v",
		s.config.AllowPasswordAuth, s.config.AllowPublicKeyAuth)
	return nil
}

// handlePasswordAuth handles password-based authentication
func (s *Server) handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	log.Printf("Password login attempt: %s from %s", username, conn.RemoteAddr())

	ctx := context.Background()

	// Try to authenticate
	player, err := s.playerRepo.Authenticate(ctx, username, string(password))
	if err != nil {
		if err == database.ErrInvalidCredentials {
			// Check if user exists
			existingPlayer, checkErr := s.playerRepo.GetByUsername(ctx, username)
			if checkErr == database.ErrPlayerNotFound {
				// User doesn't exist - offer registration if enabled
				if s.config.AllowRegistration {
					return s.handleNewUserRegistration(ctx, conn, string(password))
				}
				log.Printf("Failed login - user not found: %s", username)
				return nil, fmt.Errorf("invalid username or password")
			}

			// User exists but SSH-key-only
			if existingPlayer != nil && existingPlayer.PasswordHash == "" {
				log.Printf("Failed login - SSH key required for: %s", username)
				return nil, fmt.Errorf("this account requires SSH key authentication")
			}

			log.Printf("Failed login - invalid password: %s", username)
			return nil, fmt.Errorf("invalid username or password")
		}
		log.Printf("Authentication error for %s: %v", username, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Successful authentication
	return s.onSuccessfulAuth(ctx, player)
}

// handlePublicKeyAuth handles SSH public key authentication
func (s *Server) handlePublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	log.Printf("Public key login attempt: %s from %s", username, conn.RemoteAddr())

	ctx := context.Background()

	// Get the public key in authorized_keys format
	keyData := ssh.MarshalAuthorizedKey(key)

	// Try to find the player by public key
	playerID, err := s.sshKeyRepo.GetPlayerIDByPublicKey(ctx, keyData)
	if err != nil {
		if err == database.ErrSSHKeyNotFound {
			// Key not found - check if user exists
			player, checkErr := s.playerRepo.GetByUsername(ctx, username)
			if checkErr == database.ErrPlayerNotFound {
				// New user with SSH key - offer registration if enabled
				if s.config.AllowRegistration {
					return s.handleNewUserSSHRegistration(ctx, conn, keyData)
				}
				log.Printf("Failed SSH key login - user not found: %s", username)
				return nil, fmt.Errorf("public key not authorized")
			}

			// User exists but key not registered
			log.Printf("Failed SSH key login - key not registered for: %s (ID: %s)", username, player.ID)
			return nil, fmt.Errorf("public key not authorized for this user")
		}
		log.Printf("SSH key authentication error for %s: %v", username, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Get player by ID
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Printf("Failed to get player %s: %v", playerID, err)
		return nil, fmt.Errorf("authentication error")
	}

	// Verify username matches
	if player.Username != username {
		log.Printf("SSH key login - username mismatch: %s vs %s", username, player.Username)
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

	log.Printf("Successful login: %s (ID: %s)", player.Username, player.ID)

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
	log.Printf("New user registration requested: %s (not yet implemented)", username)
	return nil, fmt.Errorf("account not found. Contact administrator to create an account")
}

// handleNewUserSSHRegistration handles registration for a new user with SSH key
func (s *Server) handleNewUserSSHRegistration(ctx context.Context, conn ssh.ConnMetadata, keyData []byte) (*ssh.Permissions, error) {
	username := conn.User()

	// Similar to password registration, this needs interactive handling
	log.Printf("New user SSH key registration requested: %s (not yet implemented)", username)
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

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}

	log.Println("Server shutdown complete")
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
		ps.Channel.Close()
	}
}
