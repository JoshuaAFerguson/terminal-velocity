// File: internal/database/connection.go
// Project: Terminal Velocity
// Description: Database repository for connection
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/errors"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/metrics"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// DB wraps the database connection pool

var log = logger.WithComponent("Database")

type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "terminal_velocity",
		Password:        "",
		Database:        "terminal_velocity",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// NewDB creates a new database connection pool
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
			db.Close()
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

// Close closes the database connection pool
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

// Ping checks if the database is reachable
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

// QueryContext wraps sql.DB.QueryContext with metrics tracking
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	metrics.Global().IncrementDBQueries()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return rows, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with metrics tracking
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	metrics.Global().IncrementDBQueries()
	return db.DB.QueryRowContext(ctx, query, args...)
}

// ExecContext wraps sql.DB.ExecContext with metrics tracking
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	metrics.Global().IncrementDBQueries()
	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return result, err
}

// Exec wraps sql.DB.Exec with metrics tracking (for non-context queries)
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	metrics.Global().IncrementDBQueries()
	result, err := db.DB.Exec(query, args...)
	if err != nil {
		metrics.Global().IncrementDBErrors()
	}
	return result, err
}

// WithTransaction executes a function within a database transaction
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
			_ = tx.Rollback() // Ignore rollback error during panic
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
