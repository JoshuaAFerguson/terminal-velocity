// File: internal/security/totp.go
// Project: Terminal Velocity
// Description: Two-Factor Authentication (TOTP) support
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-14

package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"image/png"
	"io"

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
