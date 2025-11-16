// File: cmd/accounts/main.go
// Project: Terminal Velocity
// Description: Account management CLI tool
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package main provides the account management CLI tool for Terminal Velocity.
//
// Tool Overview:
// This utility manages player accounts for Terminal Velocity server administrators.
// It provides commands for creating accounts, managing SSH keys, and listing players.
//
// Subcommands:
//   create      Create a new player account (with password or SSH key)
//   add-key     Add SSH public key to existing account
//   list        List player accounts (currently shows online players only)
//
// Command-Line Usage:
//   accounts create -username <name> [-email <email>] [-password | -ssh-key <file>]
//   accounts add-key -username <name> -key <file>
//   accounts list [-v]
//
// Example Usage:
//   # Create account with password (prompts for password)
//   ./accounts create -username alice -email alice@example.com
//
//   # Create account with SSH key only (no password)
//   ./accounts create -username bob -email bob@example.com -ssh-key ~/.ssh/id_ed25519.pub
//
//   # Add SSH key to existing account
//   ./accounts add-key -username alice -key ~/.ssh/id_rsa.pub
//
//   # List online players
//   ./accounts list
//
//   # List online players with details (verbose)
//   ./accounts list -v
//
// Authentication Methods:
//   1. Password-based: User provides password (prompted securely, not echoed)
//   2. SSH key only: User uploads public key, authenticates via SSH protocol
//   3. Hybrid: User has both password and SSH key (can use either)
//
// Password Security:
//   - Passwords read from terminal with no echo (golang.org/x/term)
//   - Confirmation prompt ensures no typos
//   - Password complexity validation (length, character types)
//   - Hashed with bcrypt before storage (secure, slow, salted)
//
// SSH Key Management:
//   - Supports standard SSH key formats (RSA, ECDSA, ED25519)
//   - Fingerprint calculated (SHA256) for identification
//   - Multiple keys per account supported
//   - Key type and comment extracted and stored
//
// Database Connection:
// Uses default database configuration from internal/database/config.go:
//   - Host: localhost
//   - Port: 5432
//   - User: terminal_velocity
//   - Database: terminal_velocity
//
// To use different database, modify database.DefaultConfig() in code.
//
// Exit Codes:
//   0 - Success
//   1 - Argument parsing error, database error, or operation failure
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/validation"
	"golang.org/x/term"
)

// main is the entry point for the account management tool.
//
// Subcommand Dispatch:
// The tool uses flag.NewFlagSet for subcommand parsing:
//   1. Check for subcommand in os.Args[1]
//   2. Parse subcommand-specific flags
//   3. Connect to database
//   4. Execute subcommand function
//   5. Handle errors and exit
//
// Database Connection:
// Connects using DefaultConfig() which reads from environment or uses defaults.
// Connection is shared across all subcommands and closed on exit.
//
// Error Handling:
// All errors are fatal and exit with code 1.
// Error messages printed to stderr.
func main() {
	// Subcommands using FlagSet for per-command flags
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createUsername := createCmd.String("username", "", "Username for the new account")
	createEmail := createCmd.String("email", "", "Email address (optional)")
	createPassword := createCmd.Bool("password", true, "Create account with password")
	createSSHKey := createCmd.String("ssh-key", "", "Path to SSH public key file")

	addKeyCmd := flag.NewFlagSet("add-key", flag.ExitOnError)
	addKeyUsername := addKeyCmd.String("username", "", "Username to add key to")
	addKeyFile := addKeyCmd.String("key", "", "Path to SSH public key file")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listVerbose := listCmd.Bool("v", false, "Verbose output")

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Database configuration
	dbConfig := database.DefaultConfig()

	// Connect to database
	db, err := database.NewDB(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	playerRepo := database.NewPlayerRepository(db)
	sshKeyRepo := database.NewSSHKeyRepository(db)

	ctx := context.Background()

	switch os.Args[1] {
	case "create":
		if err := createCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse flags: %v\n", err)
			os.Exit(1)
		}
		if *createUsername == "" {
			fmt.Fprintln(os.Stderr, "Error: -username is required")
			createCmd.Usage()
			os.Exit(1)
		}

		if *createSSHKey != "" {
			// Create account with SSH key
			err := createAccountWithSSHKey(ctx, playerRepo, sshKeyRepo, *createUsername, *createEmail, *createSSHKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create account: %v\n", err)
				os.Exit(1)
			}
		} else if *createPassword {
			// Create account with password
			err := createAccountWithPassword(ctx, playerRepo, *createUsername, *createEmail)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create account: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Error: must specify either -password or -ssh-key")
			os.Exit(1)
		}

	case "add-key":
		if err := addKeyCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse flags: %v\n", err)
			os.Exit(1)
		}
		if *addKeyUsername == "" || *addKeyFile == "" {
			fmt.Fprintln(os.Stderr, "Error: -username and -key are required")
			addKeyCmd.Usage()
			os.Exit(1)
		}

		err := addSSHKeyToAccount(ctx, playerRepo, sshKeyRepo, *addKeyUsername, *addKeyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add SSH key: %v\n", err)
			os.Exit(1)
		}

	case "list":
		if err := listCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse flags: %v\n", err)
			os.Exit(1)
		}
		err := listAccounts(ctx, playerRepo, *listVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list accounts: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays help information for the account management tool.
//
// This function is called when:
//   - No subcommand is provided
//   - Unknown subcommand is provided
//   - User needs guidance on command syntax
//
// Output includes:
//   - Tool name and purpose
//   - Subcommand syntax
//   - Detailed examples for each subcommand
func printUsage() {
	fmt.Println("Terminal Velocity - Account Management")
	fmt.Println("\nUsage:")
	fmt.Println("  accounts create -username <name> [-email <email>] [-password | -ssh-key <file>]")
	fmt.Println("  accounts add-key -username <name> -key <file>")
	fmt.Println("  accounts list [-v]")
	fmt.Println("\nExamples:")
	fmt.Println("  # Create account with password (prompts securely)")
	fmt.Println("  accounts create -username alice -email alice@example.com")
	fmt.Println("")
	fmt.Println("  # Create account with SSH key only (no password)")
	fmt.Println("  accounts create -username bob -email bob@example.com -ssh-key ~/.ssh/id_ed25519.pub")
	fmt.Println("")
	fmt.Println("  # Add SSH key to existing account")
	fmt.Println("  accounts add-key -username alice -key ~/.ssh/id_rsa.pub")
	fmt.Println("")
	fmt.Println("  # List all online accounts with details")
	fmt.Println("  accounts list -v")
}

// createAccountWithPassword creates a new player account with password authentication.
//
// Process:
//   1. Prompt user for password (terminal echo disabled for security)
//   2. Prompt for password confirmation
//   3. Verify passwords match
//   4. Validate password complexity (via validation.ValidatePassword)
//   5. Create account in database (password is bcrypt hashed by repository)
//   6. Display success message with account details
//
// Parameters:
//   - ctx: Context for database operations
//   - repo: PlayerRepository for account creation
//   - username: Desired username (must be unique)
//   - email: Email address (optional, can be empty)
//
// Returns:
//   - error: Password mismatch, validation failure, or database error
//
// Password Security:
//   - Terminal echo disabled during entry (golang.org/x/term.ReadPassword)
//   - Confirmation prompt prevents typos
//   - Complexity validation (minimum length, character requirements)
//   - Bcrypt hashing in database layer (cost factor 10)
//
// Validation:
// Uses internal/validation package for:
//   - Minimum length (8 characters)
//   - Character variety requirements
//   - Common password blacklist
func createAccountWithPassword(ctx context.Context, repo *database.PlayerRepository, username, email string) error {
	// Prompt for password (terminal echo disabled)
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	fmt.Print("Confirm password: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println()

	if string(password) != string(confirmPassword) {
		return fmt.Errorf("passwords do not match")
	}

	// Validate password complexity (same validation as TUI registration)
	if err := validation.ValidatePassword(string(password)); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Create the account
	player, err := repo.CreateWithEmail(ctx, username, string(password), email)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Account created successfully!\n")
	fmt.Printf("  Username: %s\n", player.Username)
	fmt.Printf("  ID: %s\n", player.ID)
	if email != "" {
		fmt.Printf("  Email: %s\n", email)
	}
	fmt.Printf("  Credits: %d\n", player.Credits)

	return nil
}

// createAccountWithSSHKey creates a new player account with SSH key authentication only.
//
// This creates an account that can ONLY authenticate via SSH public key (no password).
// Useful for:
//   - Users who prefer key-based authentication
//   - Automated/scripted access
//   - Enhanced security (no password to leak)
//
// Process:
//   1. Read SSH public key from file
//   2. Create account without password (password_hash = NULL)
//   3. Add SSH public key to account
//   4. Display success message with key fingerprint
//
// Parameters:
//   - ctx: Context for database operations
//   - playerRepo: Repository for player account creation
//   - sshKeyRepo: Repository for SSH key management
//   - username: Desired username (must be unique)
//   - email: Email address (optional, can be empty)
//   - keyFile: Path to SSH public key file
//
// Returns:
//   - error: File read error, account creation error, or key add error
//
// SSH Key Format:
// Expects standard SSH public key format (authorized_keys format):
//   ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFoo... user@host
//   ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB... user@host
//
// Supported key types: RSA, ECDSA, ED25519
//
// Error Handling:
// If account creation succeeds but key add fails:
//   - Account exists without password (locked)
//   - User must use 'add-key' command to add key
//   - Warning printed to stderr
func createAccountWithSSHKey(ctx context.Context, playerRepo *database.PlayerRepository, sshKeyRepo *database.SSHKeyRepository, username, email, keyFile string) error {
	// Read the SSH public key from file
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read SSH key file: %w", err)
	}

	// Create account without password
	player, err := playerRepo.CreateWithSSHKey(ctx, username, email)
	if err != nil {
		return err
	}

	// Add the SSH key
	sshKey, err := sshKeyRepo.AddKey(ctx, player.ID, string(keyData))
	if err != nil {
		// Account was created, but key add failed
		fmt.Printf("⚠ Account created but failed to add SSH key: %v\n", err)
		fmt.Printf("  Use 'accounts add-key' to add it later\n")
		return nil
	}

	fmt.Printf("✓ Account created successfully with SSH key!\n")
	fmt.Printf("  Username: %s\n", player.Username)
	fmt.Printf("  ID: %s\n", player.ID)
	if email != "" {
		fmt.Printf("  Email: %s\n", email)
	}
	fmt.Printf("  Credits: %d\n", player.Credits)
	fmt.Printf("  SSH Key: %s\n", sshKey.Fingerprint)
	fmt.Printf("  Key Type: %s\n", sshKey.KeyType)

	return nil
}

// addSSHKeyToAccount adds an SSH public key to an existing player account.
//
// This allows:
//   - Adding SSH key authentication to password-only accounts
//   - Registering multiple SSH keys for same account (home, work, etc.)
//   - Migrating from password to SSH key authentication
//
// Process:
//   1. Look up player by username
//   2. Read SSH public key from file
//   3. Add key to player's account
//   4. Display success message with fingerprint
//
// Parameters:
//   - ctx: Context for database operations
//   - playerRepo: Repository for player lookup
//   - sshKeyRepo: Repository for SSH key management
//   - username: Existing player username
//   - keyFile: Path to SSH public key file
//
// Returns:
//   - error: Player not found, file read error, or key add error
//
// Multiple Keys:
// Players can have multiple SSH keys registered. Useful for:
//   - Different computers (home, work, laptop)
//   - Key rotation (add new, test, remove old)
//   - Shared accounts (not recommended, but supported)
//
// Key Deduplication:
// If the same key is added twice, database constraint prevents duplication.
// Error is returned with "duplicate key" message.
func addSSHKeyToAccount(ctx context.Context, playerRepo *database.PlayerRepository, sshKeyRepo *database.SSHKeyRepository, username, keyFile string) error {
	// Get player by username (verifies account exists)
	player, err := playerRepo.GetByUsername(ctx, username)
	if err != nil {
		return err
	}

	// Read the SSH public key
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read SSH key file: %w", err)
	}

	// Add the SSH key
	sshKey, err := sshKeyRepo.AddKey(ctx, player.ID, string(keyData))
	if err != nil {
		return err
	}

	fmt.Printf("✓ SSH key added successfully!\n")
	fmt.Printf("  Username: %s\n", player.Username)
	fmt.Printf("  Key Fingerprint: %s\n", sshKey.Fingerprint)
	fmt.Printf("  Key Type: %s\n", sshKey.KeyType)
	if sshKey.Comment != "" {
		fmt.Printf("  Comment: %s\n", sshKey.Comment)
	}

	return nil
}

// listAccounts displays player accounts (currently online players only).
//
// Current Implementation:
// This function currently only lists ONLINE players (those connected to server).
// Full player listing (all accounts) is planned for future version.
//
// Parameters:
//   - ctx: Context for database operations
//   - repo: PlayerRepository for querying players
//   - verbose: Show detailed player information (credits, combat rating)
//
// Returns:
//   - error: Database query error
//
// Output Format:
// Normal mode:
//   Online players (2):
//     • alice (ID: 12345678-1234-1234-1234-123456789012)
//     • bob (ID: 12345678-1234-1234-1234-123456789012)
//
// Verbose mode (-v flag):
//   Online players (2):
//     • alice (ID: 12345678-1234-1234-1234-123456789012)
//       Credits: 10000
//       Combat Rating: 150
//     • bob (ID: 12345678-1234-1234-1234-123456789012)
//       Credits: 5000
//       Combat Rating: 75
//
// Future Enhancement:
// To list ALL accounts (not just online), need to add repository method:
//   - PlayerRepository.ListAllPlayers(ctx) ([]Player, error)
func listAccounts(ctx context.Context, repo *database.PlayerRepository, verbose bool) error {
	// Note: Full account listing not yet implemented
	// Current implementation shows online players only
	fmt.Println("Account listing:")
	fmt.Println("(Full listing functionality coming soon)")
	fmt.Println("\nTo view online players, the server must be running")

	// List online players as an example of what's available
	players, err := repo.ListOnlinePlayers(ctx)
	if err != nil {
		return err
	}

	if len(players) == 0 {
		fmt.Println("\nNo players currently online")
		return nil
	}

	fmt.Printf("\nOnline players (%d):\n", len(players))
	for _, player := range players {
		fmt.Printf("  • %s (ID: %s)\n", player.Username, player.ID)
		if verbose {
			fmt.Printf("    Credits: %d\n", player.Credits)
			fmt.Printf("    Combat Rating: %d\n", player.CombatRating)
		}
	}

	return nil
}
