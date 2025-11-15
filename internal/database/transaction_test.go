// File: internal/database/transaction_test.go
// Project: Terminal Velocity
// Description: Regression tests for transaction atomicity bugs
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package database

import (
	"context"
	"database/sql"
	"testing"
)

// TestTransactionAtomicity tests that transactions are properly atomic
// Regression test for money duplication exploits (61 bugs fixed)
func TestTransactionAtomicity(t *testing.T) {
	db, err := NewDB(DefaultConfig())
	if err != nil {
		t.Skip("Skipping database tests: failed to connect to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("Rollback on error prevents partial updates", func(t *testing.T) {
		// This test ensures that if any step in a multi-step transaction fails,
		// all changes are rolled back (fixing the money duplication bugs)

		err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
			// Step 1: Insert a test record
			_, err := tx.ExecContext(ctx, "CREATE TEMP TABLE test_atomic (id INT, value INT)")
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "INSERT INTO test_atomic (id, value) VALUES (1, 100)")
			if err != nil {
				return err
			}

			// Step 2: Update the record
			_, err = tx.ExecContext(ctx, "UPDATE test_atomic SET value = 200 WHERE id = 1")
			if err != nil {
				return err
			}

			// Step 3: Intentionally cause an error
			_, err = tx.ExecContext(ctx, "INSERT INTO non_existent_table VALUES (1)")
			return err // This should cause a rollback
		})

		if err == nil {
			t.Error("Expected transaction to fail, but it succeeded")
		}

		// Verify that the temp table doesn't exist (transaction was rolled back)
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_atomic").Scan(&count)
		if err == nil {
			t.Error("Table should not exist after rollback, but it does")
		}
	})

	t.Run("Panic recovery with rollback", func(t *testing.T) {
		// Test that panics are recovered and trigger rollback

		defer func() {
			if r := recover(); r != nil {
				t.Error("Panic should have been recovered by WithTransaction")
			}
		}()

		err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "CREATE TEMP TABLE test_panic (id INT)")
			if err != nil {
				return err
			}

			// This would normally panic, but WithTransaction should recover it
			panic("intentional panic for testing")
		})

		if err == nil {
			t.Error("Expected error from panic recovery")
		}
	})

	t.Run("Successful transaction commits all changes", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "CREATE TEMP TABLE test_commit (id INT, value INT)")
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "INSERT INTO test_commit (id, value) VALUES (1, 100)")
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "UPDATE test_commit SET value = 200 WHERE id = 1")
			return err
		})

		if err != nil {
			t.Fatalf("Transaction should have succeeded: %v", err)
		}

		// Verify the final value
		var value int
		err = db.QueryRowContext(ctx, "SELECT value FROM test_commit WHERE id = 1").Scan(&value)
		if err != nil {
			t.Fatalf("Failed to query committed data: %v", err)
		}

		if value != 200 {
			t.Errorf("Expected value 200, got %d", value)
		}
	})
}

// TestConcurrentTransactions tests that concurrent transactions don't interfere
// Regression test for race conditions in transaction handling
func TestConcurrentTransactions(t *testing.T) {
	db, err := NewDB(DefaultConfig())
	if err != nil {
		t.Skip("Skipping database tests: failed to connect to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a test table
	_, err = db.ExecContext(ctx, "CREATE TEMP TABLE test_concurrent (counter INT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO test_concurrent (counter) VALUES (0)")
	if err != nil {
		t.Fatalf("Failed to insert initial value: %v", err)
	}

	// Run multiple concurrent transactions
	const numTransactions = 10
	errChan := make(chan error, numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func() {
			err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
				// Read current value
				var current int
				err := tx.QueryRowContext(ctx, "SELECT counter FROM test_concurrent FOR UPDATE").Scan(&current)
				if err != nil {
					return err
				}

				// Increment
				_, err = tx.ExecContext(ctx, "UPDATE test_concurrent SET counter = $1", current+1)
				return err
			})
			errChan <- err
		}()
	}

	// Wait for all transactions
	for i := 0; i < numTransactions; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Transaction %d failed: %v", i, err)
		}
	}

	// Verify final count
	var final int
	err = db.QueryRowContext(ctx, "SELECT counter FROM test_concurrent").Scan(&final)
	if err != nil {
		t.Fatalf("Failed to query final count: %v", err)
	}

	if final != numTransactions {
		t.Errorf("Expected counter to be %d, got %d (race condition detected)", numTransactions, final)
	}
}
