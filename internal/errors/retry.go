// File: internal/errors/retry.go
// Project: Terminal Velocity
// Description: Retry logic with exponential backoff for transient failures
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package errors provides error handling utilities including retry logic with
// exponential backoff and metrics integration.
//
// The retry package helps handle transient failures (network timeouts, temporary
// database issues, etc.) by automatically retrying failed operations with increasing
// delays between attempts. It supports both context cancellation and custom retry
// logic.
//
// Features:
//   - Exponential backoff with configurable parameters
//   - Context-aware retries (respects cancellation/timeouts)
//   - Generic retry functions for operations with/without return values
//   - Built-in transient error detection
//   - Custom retry predicate support
//   - Automatic logging of retry attempts
//
// Usage Example:
//
//	// Retry a database operation
//	err := errors.Retry(ctx, func() error {
//	    return db.Query(...)
//	}, errors.DefaultRetryConfig(), errors.IsTransientError)
//
//	// Retry with custom return value
//	result, err := errors.RetryWithResult(ctx, func() (*Data, error) {
//	    return fetchData()
//	}, errors.RetryConfig{
//	    MaxAttempts: 5,
//	    InitialDelay: 100 * time.Millisecond,
//	    MaxDelay: 10 * time.Second,
//	    Multiplier: 2.0,
//	}, errors.IsTransientError)
package errors

import (
	"context"
	"fmt"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

// log is a component-specific logger for retry operations.
var log = logger.WithComponent("Retry")

// RetryConfig configures the behavior of retry operations.
//
// This struct defines how retry attempts behave including the maximum number
// of retries, delay timing, and exponential backoff parameters.
//
// Fields:
//   - MaxAttempts: Maximum number of times to attempt the operation (including first try)
//                  Must be >= 1. Common values: 3 for quick operations, 5-10 for critical ops
//   - InitialDelay: Delay before the first retry attempt
//                   Common values: 100ms for local ops, 500ms-1s for network ops
//   - MaxDelay: Maximum delay between retry attempts (caps exponential growth)
//               Prevents excessive wait times. Common values: 5-30 seconds
//   - Multiplier: Factor to multiply delay by after each attempt (exponential backoff)
//                 Common values: 2.0 (doubles delay each time)
//
// Example:
//   // Conservative config for database operations
//   config := RetryConfig{
//       MaxAttempts: 3,
//       InitialDelay: 100 * time.Millisecond,
//       MaxDelay: 5 * time.Second,
//       Multiplier: 2.0,
//   }
//   // Delay sequence: 100ms, 200ms, 400ms
//
//   // Aggressive config for critical operations
//   config := RetryConfig{
//       MaxAttempts: 10,
//       InitialDelay: 500 * time.Millisecond,
//       MaxDelay: 30 * time.Second,
//       Multiplier: 2.0,
//   }
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns a sensible default retry configuration.
//
// The default config is suitable for most transient failure scenarios including
// database queries, API calls, and network operations.
//
// Default values:
//   - MaxAttempts: 3 (try once, retry twice)
//   - InitialDelay: 100ms
//   - MaxDelay: 5 seconds
//   - Multiplier: 2.0 (exponential backoff)
//
// Delay sequence: 100ms, 200ms, 400ms
//
// Returns:
//   - RetryConfig: A retry configuration with sensible defaults
//
// Example:
//   config := errors.DefaultRetryConfig()
//   err := errors.Retry(ctx, operation, config, errors.IsTransientError)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// IsRetryable is a function type that determines if an error should be retried.
//
// Implement this function to define custom retry logic based on error type,
// error message, or any other criteria.
//
// Parameters:
//   - error: The error returned by the operation
//
// Returns:
//   - bool: true if the operation should be retried, false to fail immediately
//
// Example:
//   customRetryable := func(err error) bool {
//       // Retry on network errors only
//       var netErr *net.Error
//       return errors.As(err, &netErr)
//   }
type IsRetryable func(error) bool

// RetryableOperation is a function type for operations that can be retried.
//
// This represents any operation that may fail transiently and should be retried.
// The operation should be idempotent (safe to execute multiple times).
//
// Returns:
//   - error: nil on success, non-nil error if the operation failed
//
// Example:
//   operation := func() error {
//       return database.UpdateRecord(id, data)
//   }
//   err := errors.Retry(ctx, operation, config, isRetryable)
type RetryableOperation func() error

// Retry executes an operation with exponential backoff retry logic.
//
// This function attempts the operation up to config.MaxAttempts times, with
// increasing delays between attempts. It respects context cancellation and can
// use custom retry logic via the isRetryable predicate.
//
// The retry algorithm:
//   1. Execute the operation
//   2. If successful, return nil
//   3. If error is not retryable (per isRetryable), return error immediately
//   4. Wait for exponentially increasing delay (InitialDelay * Multiplier^attempt)
//   5. Repeat until MaxAttempts reached or operation succeeds
//
// Parameters:
//   - ctx: Context for cancellation/timeout. If cancelled, retry stops immediately
//   - operation: The operation to retry. Should be idempotent
//   - config: Retry configuration (attempts, delays, backoff multiplier)
//   - isRetryable: Optional predicate to determine if error should be retried.
//                  If nil, all errors are retried
//
// Returns:
//   - error: nil if operation succeeds, context error if cancelled,
//            wrapped error if all attempts fail
//
// Thread Safety:
//   Safe for concurrent use. Each call maintains independent retry state.
//
// Example:
//   config := errors.DefaultRetryConfig()
//   err := errors.Retry(ctx, func() error {
//       return database.Query(ctx, "SELECT ...")
//   }, config, errors.IsTransientError)
//   if err != nil {
//       log.Error("Query failed after retries: %v", err)
//   }
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

// RetryWithResult executes an operation that returns a result with exponential backoff.
//
// This is a generic version of Retry that supports operations returning both a result
// and an error. It uses the same exponential backoff algorithm as Retry.
//
// Type Parameters:
//   - T: The type of value returned by the operation
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - operation: Function returning (result, error). Should be idempotent
//   - config: Retry configuration
//   - isRetryable: Optional predicate to determine if error should be retried
//
// Returns:
//   - T: The result from a successful operation attempt, or zero value if all fail
//   - error: nil if operation succeeds, context error if cancelled,
//            wrapped error if all attempts fail
//
// Thread Safety:
//   Safe for concurrent use. Each call maintains independent retry state.
//
// Example:
//   data, err := errors.RetryWithResult(ctx, func() (*UserData, error) {
//       return api.FetchUser(userID)
//   }, errors.DefaultRetryConfig(), errors.IsTransientError)
//   if err != nil {
//       return nil, fmt.Errorf("failed to fetch user: %w", err)
//   }
//   return data, nil
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

// IsTransientError returns true for common transient error patterns.
//
// This function examines the error message for patterns that indicate a
// temporary failure that may succeed on retry. It's designed to be used as
// the isRetryable predicate for Retry and RetryWithResult.
//
// Detected transient patterns:
//   - "connection refused" - Server temporarily unavailable
//   - "connection reset" - Network interruption
//   - "timeout" - Operation took too long
//   - "temporary failure" - Explicit temporary error
//   - "too many connections" - Resource exhaustion (may recover)
//   - "deadlock" - Database deadlock (retry often succeeds)
//   - "lock timeout" - Database lock timeout
//
// Parameters:
//   - err: The error to check (nil returns false)
//
// Returns:
//   - bool: true if the error appears to be transient and retryable
//
// Note:
//   This is a heuristic based on common error messages. For more precise
//   error classification, implement a custom IsRetryable function that
//   checks error types using errors.As() or errors.Is().
//
// Example:
//   err := errors.Retry(ctx, operation, config, errors.IsTransientError)
//
//   // Custom transient checker
//   customChecker := func(err error) bool {
//       return errors.IsTransientError(err) || isMyCustomError(err)
//   }
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

// contains checks if a string contains a substring.
// Internal helper function for error pattern matching.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		containsHelper(s, substr))
}

// containsHelper performs substring search.
// Internal helper function for error pattern matching.
func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
