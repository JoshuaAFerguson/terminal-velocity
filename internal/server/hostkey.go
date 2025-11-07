// File: internal/server/hostkey.go
// Project: Terminal Velocity
// Description: SSH server implementation and hostkey
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// generateHostKey generates a temporary ED25519 host key
// In production, this should be loaded from a file
func generateHostKey() (ssh.Signer, error) {
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
