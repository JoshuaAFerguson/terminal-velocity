package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/s0v3r1gn/terminal-velocity/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPlayerNotFound      = errors.New("player not found")
	ErrUsernameExists      = errors.New("username already exists")
	ErrInvalidCredentials  = errors.New("invalid username or password")
)

// PlayerRepository handles player data access
type PlayerRepository struct {
	db *DB
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(db *DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

// Create creates a new player account
func (r *PlayerRepository) Create(ctx context.Context, username, password string) (*models.Player, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate a new UUID
	playerID := uuid.New()

	// Insert the player
	query := `
		INSERT INTO players (id, username, password_hash, created_at, last_login)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, credits, combat_rating, created_at
	`

	now := time.Now()
	var player models.Player

	err = r.db.QueryRowContext(ctx, query, playerID, username, string(hashedPassword), now, now).Scan(
		&player.ID,
		&player.Username,
		&player.Credits,
		&player.CombatRating,
		&player.CreatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	player.Reputation = make(map[string]int)
	return &player, nil
}

// Authenticate verifies a player's credentials and returns the player if valid
func (r *PlayerRepository) Authenticate(ctx context.Context, username, password string) (*models.Player, error) {
	query := `
		SELECT id, username, password_hash, credits, current_system, combat_rating,
		       total_kills, is_online, is_criminal, faction_id, faction_rank, created_at
		FROM players
		WHERE username = $1
	`

	var player models.Player
	var passwordHash string
	var currentSystem, factionID sql.NullString
	var factionRank sql.NullString

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&player.ID,
		&player.Username,
		&passwordHash,
		&player.Credits,
		&currentSystem,
		&player.CombatRating,
		&player.TotalKills,
		&player.IsOnline,
		&player.IsCriminal,
		&factionID,
		&factionRank,
		&player.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to query player: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Handle nullable fields
	if currentSystem.Valid {
		sysID, err := uuid.Parse(currentSystem.String)
		if err == nil {
			player.CurrentSystem = sysID
		}
	}

	if factionID.Valid {
		facID, err := uuid.Parse(factionID.String)
		if err == nil {
			player.FactionID = &facID
		}
	}

	// Load reputation
	player.Reputation, err = r.loadReputation(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reputation: %w", err)
	}

	return &player, nil
}

// GetByID retrieves a player by ID
func (r *PlayerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	query := `
		SELECT id, username, credits, current_system, combat_rating,
		       total_kills, is_online, is_criminal, faction_id, faction_rank, created_at
		FROM players
		WHERE id = $1
	`

	var player models.Player
	var currentSystem, factionID sql.NullString
	var factionRank sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&player.ID,
		&player.Username,
		&player.Credits,
		&currentSystem,
		&player.CombatRating,
		&player.TotalKills,
		&player.IsOnline,
		&player.IsCriminal,
		&factionID,
		&factionRank,
		&player.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to query player: %w", err)
	}

	// Handle nullable fields
	if currentSystem.Valid {
		sysID, err := uuid.Parse(currentSystem.String)
		if err == nil {
			player.CurrentSystem = sysID
		}
	}

	if factionID.Valid {
		facID, err := uuid.Parse(factionID.String)
		if err == nil {
			player.FactionID = &facID
		}
	}

	// Load reputation
	player.Reputation, err = r.loadReputation(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reputation: %w", err)
	}

	return &player, nil
}

// GetByUsername retrieves a player by username
func (r *PlayerRepository) GetByUsername(ctx context.Context, username string) (*models.Player, error) {
	query := `
		SELECT id, username, credits, current_system, combat_rating,
		       total_kills, is_online, is_criminal, faction_id, faction_rank, created_at
		FROM players
		WHERE username = $1
	`

	var player models.Player
	var currentSystem, factionID sql.NullString
	var factionRank sql.NullString

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&player.ID,
		&player.Username,
		&player.Credits,
		&currentSystem,
		&player.CombatRating,
		&player.TotalKills,
		&player.IsOnline,
		&player.IsCriminal,
		&factionID,
		&factionRank,
		&player.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to query player: %w", err)
	}

	// Handle nullable fields
	if currentSystem.Valid {
		sysID, err := uuid.Parse(currentSystem.String)
		if err == nil {
			player.CurrentSystem = sysID
		}
	}

	if factionID.Valid {
		facID, err := uuid.Parse(factionID.String)
		if err == nil {
			player.FactionID = &facID
		}
	}

	// Load reputation
	player.Reputation, err = r.loadReputation(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reputation: %w", err)
	}

	return &player, nil
}

// Update updates a player's data
func (r *PlayerRepository) Update(ctx context.Context, player *models.Player) error {
	query := `
		UPDATE players
		SET credits = $1, current_system = $2, combat_rating = $3,
		    total_kills = $4, is_online = $5, is_criminal = $6,
		    faction_id = $7, faction_rank = $8
		WHERE id = $9
	`

	var currentSystem, factionID interface{}
	if player.CurrentSystem != uuid.Nil {
		currentSystem = player.CurrentSystem
	}
	if player.FactionID != nil {
		factionID = *player.FactionID
	}

	result, err := r.db.ExecContext(ctx, query,
		player.Credits,
		currentSystem,
		player.CombatRating,
		player.TotalKills,
		player.IsOnline,
		player.IsCriminal,
		factionID,
		nil, // faction_rank
		player.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPlayerNotFound
	}

	return nil
}

// UpdateLastLogin updates the player's last login timestamp
func (r *PlayerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE players SET last_login = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// SetOnlineStatus sets a player's online status
func (r *PlayerRepository) SetOnlineStatus(ctx context.Context, id uuid.UUID, online bool) error {
	query := `UPDATE players SET is_online = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, online, id)
	if err != nil {
		return fmt.Errorf("failed to set online status: %w", err)
	}
	return nil
}

// ModifyCredits adds or subtracts credits from a player
func (r *PlayerRepository) ModifyCredits(ctx context.Context, id uuid.UUID, amount int64) error {
	query := `
		UPDATE players
		SET credits = credits + $1
		WHERE id = $2 AND credits + $1 >= 0
	`

	result, err := r.db.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to modify credits: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("insufficient credits or player not found")
	}

	return nil
}

// UpdateReputation updates a player's reputation with a faction
func (r *PlayerRepository) UpdateReputation(ctx context.Context, playerID uuid.UUID, factionID string, change int) error {
	query := `
		INSERT INTO player_reputation (player_id, faction_id, reputation)
		VALUES ($1, $2, $3)
		ON CONFLICT (player_id, faction_id)
		DO UPDATE SET reputation = GREATEST(-100, LEAST(100, player_reputation.reputation + $3))
	`

	_, err := r.db.ExecContext(ctx, query, playerID, factionID, change)
	if err != nil {
		return fmt.Errorf("failed to update reputation: %w", err)
	}

	return nil
}

// loadReputation loads a player's reputation with all factions
func (r *PlayerRepository) loadReputation(ctx context.Context, playerID uuid.UUID) (map[string]int, error) {
	query := `SELECT faction_id, reputation FROM player_reputation WHERE player_id = $1`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reputation := make(map[string]int)
	for rows.Next() {
		var factionID string
		var rep int
		if err := rows.Scan(&factionID, &rep); err != nil {
			return nil, err
		}
		reputation[factionID] = rep
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reputation, nil
}

// Delete deletes a player (use with caution!)
func (r *PlayerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM players WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPlayerNotFound
	}

	return nil
}

// ListOnlinePlayers returns all currently online players
func (r *PlayerRepository) ListOnlinePlayers(ctx context.Context) ([]*models.Player, error) {
	query := `
		SELECT id, username, credits, current_system, combat_rating,
		       total_kills, faction_id, created_at
		FROM players
		WHERE is_online = true
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query online players: %w", err)
	}
	defer rows.Close()

	var players []*models.Player
	for rows.Next() {
		var player models.Player
		var currentSystem, factionID sql.NullString

		err := rows.Scan(
			&player.ID,
			&player.Username,
			&player.Credits,
			&currentSystem,
			&player.CombatRating,
			&player.TotalKills,
			&factionID,
			&player.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}

		// Handle nullable fields
		if currentSystem.Valid {
			sysID, err := uuid.Parse(currentSystem.String)
			if err == nil {
				player.CurrentSystem = sysID
			}
		}

		if factionID.Valid {
			facID, err := uuid.Parse(factionID.String)
			if err == nil {
				player.FactionID = &facID
			}
		}

		player.IsOnline = true
		players = append(players, &player)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	return players, nil
}

// isDuplicateKeyError checks if an error is a duplicate key violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code is 23505
	return err.Error() == "pq: duplicate key value violates unique constraint \"players_username_key\"" ||
		err.Error() == "ERROR: duplicate key value violates unique constraint \"players_username_key\" (SQLSTATE 23505)"
}
