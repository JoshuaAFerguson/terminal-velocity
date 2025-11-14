// File: internal/server/hostkey.go
// Project: Terminal Velocity
// Description: SSH server host key management with persistent storage
// Version: 2.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// loadOrGenerateHostKey loads an existing SSH host key from file or generates a new one
// The key is persisted to disk with secure permissions (0600)
func loadOrGenerateHostKey(keyPath string) (ssh.Signer, error) {
	// Try to load existing key
	if _, err := os.Stat(keyPath); err == nil {
		log.Info("Loading existing SSH host key from %s", keyPath)
		return loadHostKeyFromFile(keyPath)
	}

	// Key doesn't exist, generate a new one
	log.Info("Generating new SSH host key and saving to %s", keyPath)
	return generateAndSaveHostKey(keyPath)
}

// loadHostKeyFromFile loads an SSH private key from a file
func loadHostKeyFromFile(keyPath string) (ssh.Signer, error) {
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key file: %w", err)
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}

	log.Info("SSH host key loaded successfully (type: %s)", signer.PublicKey().Type())
	return signer, nil
}

// generateAndSaveHostKey generates a new ED25519 host key and saves it to disk
func generateAndSaveHostKey(keyPath string) (ssh.Signer, error) {
	// Generate ED25519 key pair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create SSH signer from private key
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	// Marshal the private key to OpenSSH format
	privateKeyBytes := marshalED25519PrivateKey(privateKey)

	// Write private key to file with restrictive permissions (0600)
	if err := os.WriteFile(keyPath, privateKeyBytes, 0600); err != nil {
		return nil, fmt.Errorf("failed to write host key: %w", err)
	}

	log.Info("New SSH host key generated and saved (type: ED25519, fingerprint: %s)",
		ssh.FingerprintSHA256(signer.PublicKey()))

	// Also save the public key for reference
	publicKeyPath := keyPath + ".pub"
	publicKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		log.Warn("Failed to write public key file: %v", err)
		// Non-fatal, continue
	}

	// Log the public key fingerprint for verification
	log.Info("Public key fingerprint (SHA256): %s", ssh.FingerprintSHA256(signer.PublicKey()))
	log.Info("Public key saved to: %s", publicKeyPath)

	return signer, nil
}

// marshalED25519PrivateKey marshals an ED25519 private key to OpenSSH format
func marshalED25519PrivateKey(privateKey ed25519.PrivateKey) []byte {
	// Use ssh.MarshalPrivateKey to generate the OpenSSH format
	pemBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		// Fallback to raw key format if marshaling fails
		log.Warn("Failed to marshal private key to PEM format, using raw format: %v", err)
		return privateKey
	}

	// Encode the PEM block to bytes
	return pem.EncodeToMemory(pemBlock)
}

// generateHostKey generates a temporary ED25519 host key (deprecated)
// This function is kept for backward compatibility but should not be used
// Use loadOrGenerateHostKey instead
func generateHostKey() (ssh.Signer, error) {
	log.Warn("Using deprecated generateHostKey() - use loadOrGenerateHostKey() instead")

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return signer, nil
}
