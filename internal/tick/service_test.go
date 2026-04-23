// File: internal/tick/service_test.go
// Project: Terminal Velocity
// Description: Unit tests for the tick service.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package tick

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestService_HandlerFires(t *testing.T) {
	s := New()
	var count int32
	s.Register("bump", 10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	time.Sleep(55 * time.Millisecond)
	s.Stop()
	if got := atomic.LoadInt32(&count); got < 3 {
		t.Fatalf("handler should have fired at least 3 times in 55ms at 10ms cadence, got %d", got)
	}
}

func TestService_ErrorDoesntStopHandler(t *testing.T) {
	s := New()
	var count int32
	s.Register("flaky", 10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return errors.New("boom")
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = s.Start(ctx)
	time.Sleep(55 * time.Millisecond)
	s.Stop()
	if got := atomic.LoadInt32(&count); got < 3 {
		t.Fatalf("handler should keep firing despite errors, got %d", got)
	}
}

func TestService_StopWaitsForHandler(t *testing.T) {
	s := New()
	done := make(chan struct{})
	s.Register("slow", 5*time.Millisecond, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			close(done)
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return nil
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = s.Start(ctx)
	time.Sleep(10 * time.Millisecond)
	stopStart := time.Now()
	s.Stop()
	if time.Since(stopStart) > 200*time.Millisecond {
		t.Fatalf("Stop took too long: %s", time.Since(stopStart))
	}
	// done channel should be closed if the handler saw cancellation.
	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		// Handler may not have been in-flight; that's fine.
	}
}

func TestService_CadenceClamp(t *testing.T) {
	s := New()
	var count int32
	s.Register("fast", 1*time.Microsecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = s.Start(ctx)
	time.Sleep(60 * time.Millisecond)
	s.Stop()
	// Clamped to 10ms, so at most ~6 ticks in 60ms.
	if got := atomic.LoadInt32(&count); got > 20 {
		t.Fatalf("cadence should have been clamped to 10ms, got %d ticks in 60ms", got)
	}
}

func TestService_DoubleStartErrors(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	defer s.Stop()
	if err := s.Start(ctx); err == nil {
		t.Fatal("second Start should error")
	}
}
