// File: internal/security/totp.go
// Project: Terminal Velocity
// Description: Two-Factor Authentication (TOTP) support
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-14

// Package security - TOTP (Time-based One-Time Password) two-factor authentication.
//
// This file implements TOTP-based two-factor authentication following RFC 6238.
// It provides secret generation, QR code creation, code verification, and backup
// code management for account recovery.
//
// Features:
//   - TOTP Secret Generation: Creates cryptographically secure 32-byte secrets
//   - QR Code Generation: Generates scannable QR codes for authenticator apps
//   - Code Verification: Validates 6-digit time-based codes with time window tolerance
//   - Backup Codes: Generates one-time-use recovery codes (10 codes by default)
//   - Standards Compliant: Follows RFC 6238 TOTP specification
//   - Compatible with: Google Authenticator, Authy, 1Password, Microsoft Authenticator
//
// TOTP Flow:
//
//	1. Setup (One-time):
//	   - Generate secret: GenerateSecret()
//	   - Display QR code: GenerateQRCode() or manual entry URL: GetTOTPURL()
//	   - User scans QR code with authenticator app
//	   - Generate backup codes: GenerateBackupCodes()
//	   - Store secret and backup codes in database
//
//	2. Login (Every time):
//	   - User provides username/password
//	   - Prompt for 6-digit TOTP code
//	   - Verify code: VerifyCode()
//	   - If verification fails, allow backup code: VerifyBackupCode()
//	   - Mark backup code as used (one-time use)
//
// Security Considerations:
//   - Secrets are 32 bytes (256 bits) for strong security
//   - Backup codes are base32-encoded for readability (no ambiguous characters)
//   - Backup codes should be stored hashed (not implemented here, do in database layer)
//   - QR codes should only be displayed over secure connections
//   - Time sync is critical: server time must be accurate (use NTP)
//   - Code verification has ~30 second window (1 period before/after)
//
// Database Schema (example):
//
//	CREATE TABLE player_2fa (
//	    player_id UUID PRIMARY KEY,
//	    enabled BOOLEAN NOT NULL DEFAULT false,
//	    secret TEXT NOT NULL,
//	    backup_codes TEXT[], -- Array of hashed backup codes
//	    created_at TIMESTAMP NOT NULL,
//	    last_used TIMESTAMP
//	);
//
// Usage Example:
//
//	// Setup 2FA for a player
//	tfm := security.NewTwoFactorManager("Terminal Velocity")
//
//	// Generate secret
//	secret, err := tfm.GenerateSecret(username)
//	if err != nil {
//	    return err
//	}
//
//	// Generate QR code for display
//	var qrBuf bytes.Buffer
//	err = tfm.GenerateQRCode(username, secret, &qrBuf)
//	if err != nil {
//	    return err
//	}
//	// Display qrBuf.Bytes() as PNG image
//
//	// Or provide manual entry URL
//	url := tfm.GetTOTPURL(username, secret)
//	fmt.Println("Manual entry:", url)
//
//	// Generate backup codes
//	backupCodes, err := tfm.GenerateBackupCodes(10)
//	if err != nil {
//	    return err
//	}
//
//	// Store in database (hash backup codes!)
//	SaveToDatabase(playerID, secret, hashBackupCodes(backupCodes))
//
//	// Verify code during login
//	userCode := promptUserForCode()
//	if tfm.VerifyCode(secret, userCode) {
//	    log.Info("2FA verification successful")
//	} else {
//	    // Try backup code
//	    valid, remaining := tfm.VerifyBackupCode(backupCodes, userCode)
//	    if valid {
//	        log.Info("Backup code accepted")
//	        UpdateBackupCodes(playerID, remaining) // Update database
//	    } else {
//	        return errors.New("invalid code")
//	    }
//	}
//
// Version: 1.1.0
// Last Updated: 2025-11-16
package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"image/png"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TwoFactorConfig holds 2FA configuration for a player
type TwoFactorConfig struct {
	PlayerID     uuid.UUID
	Enabled      bool
	Secret       string
	BackupCodes  []string
	CreatedAt    time.Time
	LastUsed     *time.Time
}

// TwoFactorManager manages two-factor authentication
type TwoFactorManager struct {
	issuer string // Application name for TOTP
}

// NewTwoFactorManager creates a new 2FA manager
func NewTwoFactorManager(issuer string) *TwoFactorManager {
	if issuer == "" {
		issuer = "Terminal Velocity"
	}

	return &TwoFactorManager{
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (tfm *TwoFactorManager) GenerateSecret(username string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tfm.issuer,
		AccountName: username,
		SecretSize:  32,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key.Secret(), nil
}

// GenerateQRCode generates a QR code for TOTP setup
func (tfm *TwoFactorManager) GenerateQRCode(username, secret string, output io.Writer) error {
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		tfm.issuer, username, secret, tfm.issuer))
	if err != nil {
		return fmt.Errorf("failed to create TOTP key: %w", err)
	}

	// Generate QR code
	img, err := key.Image(200, 200)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Encode as PNG
	if err := png.Encode(output, img); err != nil {
		return fmt.Errorf("failed to encode QR code: %w", err)
	}

	return nil
}

// VerifyCode verifies a TOTP code
func (tfm *TwoFactorManager) VerifyCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates backup codes for account recovery
func (tfm *TwoFactorManager) GenerateBackupCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code, err := generateRandomCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}

	return codes, nil
}

// VerifyBackupCode verifies a backup code (should be one-time use)
func (tfm *TwoFactorManager) VerifyBackupCode(backupCodes []string, code string) (bool, []string) {
	for i, backupCode := range backupCodes {
		if backupCode == code {
			// Remove used code
			remaining := append(backupCodes[:i], backupCodes[i+1:]...)
			return true, remaining
		}
	}

	return false, backupCodes
}

// generateRandomCode generates a random alphanumeric code
func generateRandomCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode as base32 for readability (no ambiguous characters)
	code := base32.StdEncoding.EncodeToString(bytes)

	// Trim to desired length and remove padding
	code = code[:length]

	return code, nil
}

// GetTOTPURL returns the TOTP URL for manual entry
func (tfm *TwoFactorManager) GetTOTPURL(username, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		tfm.issuer, username, secret, tfm.issuer)
}
