// File: internal/database/ssh_key_repository.go
// Project: Terminal Velocity
// Description: Repository for SSH public key authentication and management
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// SSH key errors.
var (
	// ErrSSHKeyNotFound indicates the requested SSH key does not exist.
	ErrSSHKeyNotFound = errors.New("SSH key not found")

	// ErrSSHKeyExists indicates this SSH key is already registered.
	ErrSSHKeyExists = errors.New("SSH key already exists")

	// ErrInvalidPublicKey indicates the public key format is invalid.
	ErrInvalidPublicKey = errors.New("invalid public key format")
)

// SSHKeyRepository handles all database operations for SSH public keys.
//
// Manages SSH key-based authentication:
//   - Adding/removing player SSH public keys
//   - Key verification during SSH handshake
//   - SHA256 fingerprint calculation and matching
//   - Key activation/deactivation
//   - Last used timestamp tracking
//
// Security features:
//   - Keys validated on addition (must parse as valid SSH key)
//   - Fingerprints stored as SHA256 hashes
//   - Only active keys can be used for authentication
//   - Duplicate keys prevented by fingerprint uniqueness
//
// Thread-safety:
//   - All methods are thread-safe
type SSHKeyRepository struct {
	db *DB // Database connection pool
}

// NewSSHKeyRepository creates a new SSH key repository
func NewSSHKeyRepository(db *DB) *SSHKeyRepository {
	return &SSHKeyRepository{db: db}
}

// AddKey adds and validates a new SSH public key for a player.
//
// Parses and validates the public key, calculates SHA256 fingerprint,
// and stores it in the database. The key is marked as active by default.
//
// Security:
//   - Key must be valid SSH format (validated by ssh.ParseAuthorizedKey)
//   - Fingerprint uniqueness prevents duplicate keys
//   - Supports RSA, ECDSA, Ed25519 key types
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - playerID: Player UUID
//   - publicKeyStr: SSH public key in authorized_keys format
//
// Returns:
//   - *models.SSHKey: Added key with fingerprint and metadata
//   - error: ErrInvalidPublicKey if invalid format, ErrSSHKeyExists if duplicate
func (r *SSHKeyRepository) AddKey(ctx context.Context, playerID uuid.UUID, publicKeyStr string) (*models.SSHKey, error) {
	// Parse and validate the public key
	publicKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyStr))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
	}

	// Calculate fingerprint (SHA256)
	fingerprint := ssh.FingerprintSHA256(publicKey)

	// Get key type
	keyType := publicKey.Type()

	// Generate UUID for the key
	keyID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO player_ssh_keys (id, player_id, key_type, public_key, fingerprint, comment, added_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, player_id, key_type, public_key, fingerprint, comment, added_at, is_active
	`

	var sshKey models.SSHKey
	err = r.db.QueryRowContext(ctx, query,
		keyID,
		playerID,
		keyType,
		publicKeyStr,
		fingerprint,
		string(comment),
		now,
		true,
	).Scan(
		&sshKey.ID,
		&sshKey.PlayerID,
		&sshKey.KeyType,
		&sshKey.PublicKey,
		&sshKey.Fingerprint,
		&sshKey.Comment,
		&sshKey.AddedAt,
		&sshKey.IsActive,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrSSHKeyExists
		}
		return nil, fmt.Errorf("failed to add SSH key: %w", err)
	}

	return &sshKey, nil
}

// GetKeysByPlayer returns all SSH keys for a player
func (r *SSHKeyRepository) GetKeysByPlayer(ctx context.Context, playerID uuid.UUID) ([]*models.SSHKey, error) {
	query := `
		SELECT id, player_id, key_type, public_key, fingerprint, comment, added_at, last_used, is_active
		FROM player_ssh_keys
		WHERE player_id = $1
		ORDER BY added_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query SSH keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.SSHKey
	for rows.Next() {
		var key models.SSHKey
		var lastUsed sql.NullTime

		err := rows.Scan(
			&key.ID,
			&key.PlayerID,
			&key.KeyType,
			&key.PublicKey,
			&key.Fingerprint,
			&key.Comment,
			&key.AddedAt,
			&lastUsed,
			&key.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan SSH key: %w", err)
		}

		if lastUsed.Valid {
			key.LastUsed = &lastUsed.Time
		}

		keys = append(keys, &key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SSH keys: %w", err)
	}

	return keys, nil
}

// GetActiveKeysByPlayer returns only active SSH keys for a player
func (r *SSHKeyRepository) GetActiveKeysByPlayer(ctx context.Context, playerID uuid.UUID) ([]*models.SSHKey, error) {
	query := `
		SELECT id, player_id, key_type, public_key, fingerprint, comment, added_at, last_used, is_active
		FROM player_ssh_keys
		WHERE player_id = $1 AND is_active = true
		ORDER BY added_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active SSH keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.SSHKey
	for rows.Next() {
		var key models.SSHKey
		var lastUsed sql.NullTime

		err := rows.Scan(
			&key.ID,
			&key.PlayerID,
			&key.KeyType,
			&key.PublicKey,
			&key.Fingerprint,
			&key.Comment,
			&key.AddedAt,
			&lastUsed,
			&key.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan SSH key: %w", err)
		}

		if lastUsed.Valid {
			key.LastUsed = &lastUsed.Time
		}

		keys = append(keys, &key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SSH keys: %w", err)
	}

	return keys, nil
}

// GetKeyByFingerprint returns an SSH key by its fingerprint
func (r *SSHKeyRepository) GetKeyByFingerprint(ctx context.Context, fingerprint string) (*models.SSHKey, error) {
	query := `
		SELECT id, player_id, key_type, public_key, fingerprint, comment, added_at, last_used, is_active
		FROM player_ssh_keys
		WHERE fingerprint = $1
	`

	var key models.SSHKey
	var lastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, query, fingerprint).Scan(
		&key.ID,
		&key.PlayerID,
		&key.KeyType,
		&key.PublicKey,
		&key.Fingerprint,
		&key.Comment,
		&key.AddedAt,
		&lastUsed,
		&key.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSSHKeyNotFound
		}
		return nil, fmt.Errorf("failed to query SSH key: %w", err)
	}

	if lastUsed.Valid {
		key.LastUsed = &lastUsed.Time
	}

	return &key, nil
}

// UpdateLastUsed updates the last_used timestamp for a key
func (r *SSHKeyRepository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	query := `UPDATE player_ssh_keys SET last_used = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), keyID)
	if err != nil {
		return fmt.Errorf("failed to update last_used: %w", err)
	}
	return nil
}

// DeactivateKey deactivates an SSH key
func (r *SSHKeyRepository) DeactivateKey(ctx context.Context, keyID uuid.UUID) error {
	query := `UPDATE player_ssh_keys SET is_active = false WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to deactivate SSH key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSSHKeyNotFound
	}

	return nil
}

// DeleteKey permanently deletes an SSH key
func (r *SSHKeyRepository) DeleteKey(ctx context.Context, keyID uuid.UUID) error {
	query := `DELETE FROM player_ssh_keys WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSSHKeyNotFound
	}

	return nil
}

// VerifyKey checks if a public key is valid for a player
func (r *SSHKeyRepository) VerifyKey(ctx context.Context, playerID uuid.UUID, publicKeyStr string) (bool, *models.SSHKey, error) {
	// Parse the public key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyStr))
	if err != nil {
		return false, nil, ErrInvalidPublicKey
	}

	// Calculate fingerprint
	fingerprint := ssh.FingerprintSHA256(publicKey)

	// Look up the key
	query := `
		SELECT id, player_id, key_type, public_key, fingerprint, comment, added_at, last_used, is_active
		FROM player_ssh_keys
		WHERE player_id = $1 AND fingerprint = $2 AND is_active = true
	`

	var key models.SSHKey
	var lastUsed sql.NullTime

	err = r.db.QueryRowContext(ctx, query, playerID, fingerprint).Scan(
		&key.ID,
		&key.PlayerID,
		&key.KeyType,
		&key.PublicKey,
		&key.Fingerprint,
		&key.Comment,
		&key.AddedAt,
		&lastUsed,
		&key.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to verify SSH key: %w", err)
	}

	if lastUsed.Valid {
		key.LastUsed = &lastUsed.Time
	}

	return true, &key, nil
}

// GetPlayerIDByPublicKey returns the player ID for a given public key
func (r *SSHKeyRepository) GetPlayerIDByPublicKey(ctx context.Context, publicKeyData []byte) (uuid.UUID, error) {
	// Parse the public key to get fingerprint
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyData)
	if err != nil {
		return uuid.Nil, ErrInvalidPublicKey
	}

	fingerprint := ssh.FingerprintSHA256(publicKey)

	query := `
		SELECT player_id FROM player_ssh_keys
		WHERE fingerprint = $1 AND is_active = true
	`

	var playerID uuid.UUID
	err = r.db.QueryRowContext(ctx, query, fingerprint).Scan(&playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, ErrSSHKeyNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to get player ID: %w", err)
	}

	return playerID, nil
}
