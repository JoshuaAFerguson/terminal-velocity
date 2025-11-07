// File: cmd/accounts/main.go
// Project: Terminal Velocity
// Description: Account management CLI tool
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"golang.org/x/term"
)

func main() {
	// Subcommands
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
		listCmd.Parse(os.Args[2:])
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

func printUsage() {
	fmt.Println("Terminal Velocity - Account Management")
	fmt.Println("\nUsage:")
	fmt.Println("  accounts create -username <name> [-email <email>] [-password | -ssh-key <file>]")
	fmt.Println("  accounts add-key -username <name> -key <file>")
	fmt.Println("  accounts list [-v]")
	fmt.Println("\nExamples:")
	fmt.Println("  # Create account with password")
	fmt.Println("  accounts create -username alice -email alice@example.com")
	fmt.Println("")
	fmt.Println("  # Create account with SSH key only")
	fmt.Println("  accounts create -username bob -email bob@example.com -ssh-key ~/.ssh/id_ed25519.pub")
	fmt.Println("")
	fmt.Println("  # Add SSH key to existing account")
	fmt.Println("  accounts add-key -username alice -key ~/.ssh/id_rsa.pub")
	fmt.Println("")
	fmt.Println("  # List all accounts")
	fmt.Println("  accounts list -v")
}

func createAccountWithPassword(ctx context.Context, repo *database.PlayerRepository, username, email string) error {
	// Prompt for password
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

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
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

func createAccountWithSSHKey(ctx context.Context, playerRepo *database.PlayerRepository, sshKeyRepo *database.SSHKeyRepository, username, email, keyFile string) error {
	// Read the SSH public key
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

func addSSHKeyToAccount(ctx context.Context, playerRepo *database.PlayerRepository, sshKeyRepo *database.SSHKeyRepository, username, keyFile string) error {
	// Get player by username
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

func listAccounts(ctx context.Context, repo *database.PlayerRepository, verbose bool) error {
	// For a simple list, we'll need to add a method to list all players
	// For now, just print a message
	fmt.Println("Account listing:")
	fmt.Println("(Full listing functionality coming soon)")
	fmt.Println("\nTo view online players, the server must be running")

	// List online players as an example
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
