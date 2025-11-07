// File: internal/errors/retry.go
// Project: Terminal Velocity
// Description: Retry logic for transient failures
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package errors

import (
	"context"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("Retry")

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// IsRetryable determines if an error should be retried
type IsRetryable func(error) bool

// RetryableOperation is a function that can be retried
type RetryableOperation func() error

// Retry executes an operation with exponential backoff
func Retry(ctx context.Context, operation RetryableOperation, config RetryConfig, isRetryable IsRetryable) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Warn("Retry cancelled due to context: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		// Execute operation
		err := operation()
		if err == nil {
			if attempt > 1 {
				log.Info("Operation succeeded after %d attempts", attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if isRetryable != nil && !isRetryable(err) {
			log.Debug("Error is not retryable: %v", err)
			return err
		}

		// Don't sleep on last attempt
		if attempt == config.MaxAttempts {
			break
		}

		log.Warn("Operation failed (attempt %d/%d), retrying after %v: %v",
			attempt, config.MaxAttempts, delay, err)

		// Wait with backoff
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	log.Error("Operation failed after %d attempts: %v", config.MaxAttempts, lastErr)
	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult executes an operation that returns a result with exponential backoff
func RetryWithResult[T any](ctx context.Context, operation func() (T, error), config RetryConfig, isRetryable IsRetryable) (T, error) {
	var result T
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Warn("Retry cancelled due to context: %v", ctx.Err())
			return result, ctx.Err()
		default:
		}

		// Execute operation
		r, err := operation()
		if err == nil {
			if attempt > 1 {
				log.Info("Operation succeeded after %d attempts", attempt)
			}
			return r, nil
		}

		lastErr = err

		// Check if error is retryable
		if isRetryable != nil && !isRetryable(err) {
			log.Debug("Error is not retryable: %v", err)
			return result, err
		}

		// Don't sleep on last attempt
		if attempt == config.MaxAttempts {
			break
		}

		log.Warn("Operation failed (attempt %d/%d), retrying after %v: %v",
			attempt, config.MaxAttempts, delay, err)

		// Wait with backoff
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return result, ctx.Err()
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	log.Error("Operation failed after %d attempts: %v", config.MaxAttempts, lastErr)
	return result, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// IsTransientError returns true for common transient errors
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Common transient error patterns
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"too many connections",
		"deadlock",
		"lock timeout",
	}

	for _, pattern := range transientPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
