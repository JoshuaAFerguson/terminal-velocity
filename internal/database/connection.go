// File: internal/database/connection.go
// Project: Terminal Velocity
// Description: Database connection pool management with retry logic, metrics tracking,
//              and transaction support. Provides thread-safe database operations for
//              the Terminal Velocity game server.
// Version: 1.2.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package database provides PostgreSQL database access for Terminal Velocity.
//
// This package implements the repository pattern for all database operations,
// providing:
//   - Connection pool management with configurable limits
//   - Automatic retry logic for transient errors
//   - Transaction support with automatic rollback
//   - Metrics tracking for all database operations
//   - Thread-safe concurrent access
//
// All database queries use parameterized statements to prevent SQL injection.
// The package uses pgx/v5 as the PostgreSQL driver for optimal performance.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/errors"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/metrics"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// log is the logger instance for database operations.
var log = logger.WithComponent("Database")

// DB wraps the database connection pool and provides enhanced functionality.
//
// This struct embeds *sql.DB and adds metrics tracking to all query operations.
// It is safe for concurrent use by multiple goroutines as the underlying
// sql.DB connection pool is thread-safe.
//
// All methods on DB automatically:
//   - Track metrics for monitoring
//   - Log errors for debugging
//   - Use the connection pool efficiently
type DB struct {
	*sql.DB // Embedded connection pool from database/sql
}

// Config holds database configuration parameters.
//
// Configuration can be loaded from environment variables or provided directly.
// Environment variables take precedence over defaults when using DefaultConfig().
//
// Environment variables:
//   - DB_HOST: Database server hostname (default: localhost)
//   - DB_PORT: Database server port (default: 5432)
//   - DB_USER: Database username (default: terminal_velocity)
//   - DB_PASSWORD: Database password (required for security)
//   - DB_NAME: Database name (default: terminal_velocity)
//   - DB_SSLMODE: SSL mode (default: disable)
//   - DB_MAX_OPEN_CONNS: Maximum open connections (default: 25)
//   - DB_MAX_IDLE_CONNS: Maximum idle connections (default: 5)
type Config struct {
	Host     string // Database server hostname or IP address
	Port     int    // Database server port (typically 5432)
	User     string // Database username for authentication
	Password string // Database password (should never be logged)
	Database string // Database name to connect to
	SSLMode  string // SSL mode: disable, require, verify-ca, or verify-full

	// Connection pool settings control resource usage and performance
	MaxOpenConns    int           // Maximum number of open connections to the database
	MaxIdleConns    int           // Maximum number of idle connections in the pool
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection (prevents stale connections)
	ConnMaxIdleTime time.Duration // Maximum time a connection can be idle before closing
}

// DefaultConfig returns a default database configuration.
//
// This function creates a Config struct with sensible defaults, overriding them
// with environment variables when present. This allows for flexible deployment:
//   - Development: Use defaults (no env vars needed)
//   - Production: Override via environment variables
//   - Docker: Configure via docker-compose.yml or Kubernetes secrets
//
// Returns:
//   - *Config: Configuration with defaults or environment variable overrides
//
// Security notes:
//   - Warns if DB_PASSWORD is not set (insecure default)
//   - Never logs password values
//   - Supports SSL modes for encrypted connections
func DefaultConfig() *Config {
	cfg := &Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "terminal_velocity"),
		Password:        getEnv("DB_PASSWORD", ""),
		Database:        getEnv("DB_NAME", "terminal_velocity"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	// Warn if using default password
	if cfg.Password == "" {
		log.Warn("Database password not set! Set DB_PASSWORD environment variable for security")
	}

	// Log which values came from environment variables
	if os.Getenv("DB_HOST") != "" {
		log.Debug("Using DB_HOST from environment: %s", cfg.Host)
	}
	if os.Getenv("DB_PASSWORD") != "" {
		log.Debug("Using DB_PASSWORD from environment (value hidden)")
	}

	return cfg
}

// getEnv retrieves an environment variable or returns a default value.
//
// Parameters:
//   - key: Environment variable name to lookup
//   - defaultValue: Value to return if environment variable is not set or empty
//
// Returns:
//   - string: Environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
//
// If the environment variable exists but cannot be parsed as an integer,
// a warning is logged and the default value is returned.
//
// Parameters:
//   - key: Environment variable name to lookup
//   - defaultValue: Value to return if variable is not set or invalid
//
// Returns:
//   - int: Parsed integer value or default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		log.Warn("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// NewDB creates a new database connection pool with automatic retry logic.
//
// This function establishes a connection to PostgreSQL and configures the
// connection pool for optimal performance. It uses retry logic to handle
// transient connection failures common in containerized environments.
//
// Connection process:
//  1. Build PostgreSQL DSN from config
//  2. Open connection with pgx driver
//  3. Configure connection pool limits
//  4. Ping database to verify connectivity
//  5. Retry on failure with exponential backoff
//
// Parameters:
//   - cfg: Database configuration (nil uses DefaultConfig)
//
// Returns:
//   - *DB: Wrapped connection pool ready for use
//   - error: Connection error (after all retries exhausted)
//
// Thread-safety:
//   - The returned *DB is safe for concurrent use by multiple goroutines
//   - Connection pool handles concurrency automatically
//
// Example:
//
//	db, err := NewDB(nil) // Use default config
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func NewDB(cfg *Config) (*DB, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log.Info("Connecting to database: host=%s, port=%d, database=%s", cfg.Host, cfg.Port, cfg.Database)

	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	var db *sql.DB
	var dbWrapper *DB

	// Use retry logic for establishing connection
	retryConfig := errors.DefaultRetryConfig()
	ctx := context.Background()

	err := errors.Retry(ctx, func() error {
		// Open database connection
		var err error
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			errors.RecordGlobalError("database", "connection_open", err)
			log.Error("Failed to open database connection: error=%v", err)
			return err
		}

		// Configure connection pool
		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

		// Test the connection with ping
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(pingCtx); err != nil {
			errors.RecordGlobalError("database", "connection_ping", err)
			log.Error("Failed to ping database: error=%v", err)
			// Clean up database connection on ping failure
			if closeErr := db.Close(); closeErr != nil {
				log.Warn("Failed to close database during cleanup: error=%v", closeErr)
			}
			return err
		}

		return nil
	}, retryConfig, errors.IsTransientError)

	if err != nil {
		log.Error("Failed to establish database connection after retries: error=%v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Debug("Connection pool configured: max_open=%d, max_idle=%d, max_lifetime=%v",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)

	dbWrapper = &DB{DB: db}
	log.Info("Database connection established successfully")
	return dbWrapper, nil
}

// Close gracefully closes the database connection pool.
//
// This method should be called when shutting down the server to ensure
// all database connections are properly closed. It waits for active
// queries to complete before closing.
//
// Returns:
//   - error: Error if closure fails (rare, but should be logged)
//
// Thread-safety:
//   - Safe to call concurrently, but should only be called once during shutdown
//   - Subsequent database operations will fail after Close() is called
func (db *DB) Close() error {
	log.Info("Closing database connection")
	err := db.DB.Close()
	if err != nil {
		log.Error("Error closing database connection: error=%v", err)
		return err
	}
	log.Info("Database connection closed successfully")
	return nil
}

// Ping verifies database connectivity and records metrics.
//
// This method sends a lightweight query to the database to verify the
// connection is alive and responsive. Use this for health checks.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - error: Error if database is unreachable or unresponsive
//
// Thread-safety:
//   - Safe for concurrent use
func (db *DB) Ping(ctx context.Context) error {
	err := db.PingContext(ctx)
	if err != nil {
		errors.RecordGlobalError("database", "ping_failed", err)
		log.Error("Database ping failed: error=%v", err)
		return err
	}
	log.Debug("Database ping successful")
	return nil
}

// QueryContext executes a query that returns multiple rows with metrics tracking.
//
// This method wraps sql.DB.QueryContext to automatically track metrics for
// monitoring and debugging. Use this for SELECT queries that return 0 or more rows.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - query: SQL query with parameter placeholders ($1, $2, etc.)
//   - args: Values to bind to query parameters (prevents SQL injection)
//
// Returns:
//   - *sql.Rows: Result set (must be closed by caller)
//   - error: Query execution error
//
// Thread-safety:
//   - Safe for concurrent use
//
// Example:
//
//	rows, err := db.QueryContext(ctx, "SELECT id, name FROM players WHERE credits > $1", 1000)
//	if err != nil { return err }
//	defer rows.Close()
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	metrics.Global().IncrementDBQueries()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return rows, err
}

// QueryRowContext executes a query that returns at most one row with metrics tracking.
//
// This method wraps sql.DB.QueryRowContext to automatically track metrics.
// Use this for SELECT queries expected to return a single row or nothing.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - query: SQL query with parameter placeholders ($1, $2, etc.)
//   - args: Values to bind to query parameters (prevents SQL injection)
//
// Returns:
//   - *sql.Row: Single row result (call .Scan() to retrieve values)
//
// Thread-safety:
//   - Safe for concurrent use
//
// Example:
//
//	var name string
//	err := db.QueryRowContext(ctx, "SELECT name FROM players WHERE id = $1", playerID).Scan(&name)
//	if err == sql.ErrNoRows { /* not found */ }
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	metrics.Global().IncrementDBQueries()
	return db.DB.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a query that doesn't return rows with metrics tracking.
//
// This method wraps sql.DB.ExecContext to automatically track metrics.
// Use this for INSERT, UPDATE, DELETE, and DDL queries.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - query: SQL query with parameter placeholders ($1, $2, etc.)
//   - args: Values to bind to query parameters (prevents SQL injection)
//
// Returns:
//   - sql.Result: Contains RowsAffected() and LastInsertId()
//   - error: Query execution error
//
// Thread-safety:
//   - Safe for concurrent use
//
// Example:
//
//	result, err := db.ExecContext(ctx, "UPDATE players SET credits = $1 WHERE id = $2", 1000, playerID)
//	if err != nil { return err }
//	rows, _ := result.RowsAffected()
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	metrics.Global().IncrementDBQueries()
	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return result, err
}

// Exec executes a query without context with metrics tracking.
//
// This method is provided for compatibility with legacy code that doesn't
// use contexts. New code should use ExecContext instead.
//
// Parameters:
//   - query: SQL query with parameter placeholders ($1, $2, etc.)
//   - args: Values to bind to query parameters (prevents SQL injection)
//
// Returns:
//   - sql.Result: Contains RowsAffected() and LastInsertId()
//   - error: Query execution error
//
// Thread-safety:
//   - Safe for concurrent use
//
// Deprecated: Use ExecContext for better timeout and cancellation control.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	metrics.Global().IncrementDBQueries()
	result, err := db.DB.Exec(query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return result, err
}

// WithTransaction executes a function within a database transaction with automatic rollback on error.
//
// This function provides ACID guarantees for multi-step database operations by wrapping
// them in a transaction. It handles:
//   - Automatic commit on success
//   - Automatic rollback on error
//   - Panic recovery with rollback
//   - Error metrics tracking
//   - Comprehensive logging
//
// Usage:
//
//	err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
//	    // Step 1: Deduct credits
//	    _, err := tx.ExecContext(ctx, "UPDATE players SET credits = credits - $1 WHERE id = $2", cost, playerID)
//	    if err != nil {
//	        return err  // Will trigger rollback
//	    }
//
//	    // Step 2: Add item to inventory
//	    _, err = tx.ExecContext(ctx, "INSERT INTO inventory (player_id, item_id) VALUES ($1, $2)", playerID, itemID)
//	    if err != nil {
//	        return err  // Will trigger rollback
//	    }
//
//	    return nil  // Will commit
//	})
//
// Critical for preventing exploits:
//   - Money duplication bugs: Ensures all credit transfers are atomic
//   - Inventory inconsistencies: Ensures item creation and payment happen together
//   - Race conditions: Database-level locking prevents concurrent modification
//
// Thread-safe: Each transaction gets its own connection from the pool.
func (db *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	log.Debug("Beginning database transaction")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		errors.RecordGlobalError("database", "transaction_begin", err)
		log.Error("Failed to begin transaction: error=%v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			errors.RecordGlobalError("database", "transaction_panic", fmt.Errorf("panic: %v", p))
			log.Error("PANIC in transaction, rolling back: panic=%v", p)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Rollback failed during panic: rollback_error=%v, panic=%v", rbErr, p)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		errors.RecordGlobalError("database", "transaction_error", err)
		log.Warn("Transaction failed, rolling back: error=%v", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			errors.RecordGlobalError("database", "transaction_rollback", rbErr)
			log.Error("Rollback failed: rollback_error=%v, original_error=%v", rbErr, err)
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		errors.RecordGlobalError("database", "transaction_commit", err)
		log.Error("Failed to commit transaction: error=%v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Debug("Transaction committed successfully")
	return nil
}
