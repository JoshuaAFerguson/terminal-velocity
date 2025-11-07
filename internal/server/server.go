package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/s0v3r1gn/terminal-velocity/internal/database"
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
		go s.handleSession(sshConn.User(), channel, requests)
	}
}

// handleSession handles a single SSH session
func (s *Server) handleSession(username string, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Handle session requests (pty-req, shell, etc.)
	for req := range requests {
		switch req.Type {
		case "pty-req":
			req.Reply(true, nil)
		case "shell":
			req.Reply(true, nil)
			// Start game session
			s.startGameSession(username, channel)
			return
		default:
			req.Reply(false, nil)
		}
	}
}

// startGameSession starts a game session for a player
func (s *Server) startGameSession(username string, channel ssh.Channel) {
	// TODO: Implement game session using bubbletea
	welcome := fmt.Sprintf("\r\n=== Welcome to Terminal Velocity ===\r\n\r\nHello, %s!\r\n\r\nThis is a placeholder. The game is under development.\r\n\r\nPress Ctrl+D to disconnect.\r\n", username)
	channel.Write([]byte(welcome))

	// Wait for input (simple for now)
	buf := make([]byte, 1024)
	for {
		n, err := channel.Read(buf)
		if err != nil {
			break
		}
		// Echo back for now
		channel.Write(buf[:n])
	}
}

// initSSHConfig initializes SSH server configuration
func (s *Server) initSSHConfig() error {
	s.sshConfig = &ssh.ServerConfig{
		// Authenticate users against the database
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			username := conn.User()
			log.Printf("Login attempt: %s from %s", username, conn.RemoteAddr())

			// Authenticate against database
			ctx := context.Background()
			player, err := s.playerRepo.Authenticate(ctx, username, string(password))
			if err != nil {
				if err == database.ErrInvalidCredentials {
					log.Printf("Failed login attempt: %s", username)
					return nil, fmt.Errorf("invalid credentials")
				}
				log.Printf("Authentication error for %s: %v", username, err)
				return nil, fmt.Errorf("authentication error")
			}

			// Update last login and set online status
			go func() {
				ctx := context.Background()
				s.playerRepo.UpdateLastLogin(ctx, player.ID)
				s.playerRepo.SetOnlineStatus(ctx, player.ID, true)
			}()

			log.Printf("Successful login: %s (ID: %s)", username, player.ID)

			// Return permissions with player ID
			return &ssh.Permissions{
				Extensions: map[string]string{
					"player_id": player.ID.String(),
					"username":  username,
				},
			}, nil
		},
	}

	// Generate a temporary host key (TODO: load from file)
	// For development only - in production, use a persistent key
	privateKey, err := generateHostKey()
	if err != nil {
		return fmt.Errorf("failed to generate host key: %w", err)
	}

	s.sshConfig.AddHostKey(privateKey)
	return nil
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
