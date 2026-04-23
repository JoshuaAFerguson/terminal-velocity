// File: internal/tick/service.go
// Project: Terminal Velocity
// Description: Server-wide tick service for background universe simulation.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

// Package tick owns the long-running background work that makes the world
// feel alive between player actions: market stocks drift back toward
// equilibrium, news articles trickle in, encounters spawn in unoccupied
// systems, missions expire.
//
// Design:
//   - One Service per server, owned by Server.
//   - Handlers register with a cadence and a fn. Each runs on its own
//     ticker goroutine so a slow handler can't stall a fast one.
//   - Stop() cancels a parent context and waits for all handlers to return
//     before unblocking — so server shutdown is clean.
//   - Handlers receive the parent context on every call and must respect
//     cancellation promptly.
package tick

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
)

var log = logger.WithComponent("Tick")

// Handler is a function the tick service calls periodically. It receives
// the service's parent context; returning is the only way to defer work
// until the next tick.
type Handler func(ctx context.Context) error

// handlerSpec pairs a handler with its cadence and a human-readable name
// for logs.
type handlerSpec struct {
	name    string
	cadence time.Duration
	fn      Handler
}

// Service runs registered Handlers on independent goroutines.
type Service struct {
	mu       sync.Mutex
	handlers []handlerSpec
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	started  bool
}

// New returns a ready-to-configure Service. Call Register before Start.
func New() *Service {
	return &Service{}
}

// Register adds a handler to run every `cadence` until the service is
// stopped. The name is used in log lines and should be unique-enough for
// greppability. Registering after Start is ignored with a warning rather
// than a panic so late-loading plugins can't crash the server.
func (s *Service) Register(name string, cadence time.Duration, fn Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		log.Warn("Register(%q) ignored — tick service already started", name)
		return
	}
	if cadence < 10*time.Millisecond {
		log.Warn("Register(%q) cadence %s clamped to 10ms minimum", name, cadence)
		cadence = 10 * time.Millisecond
	}
	s.handlers = append(s.handlers, handlerSpec{name: name, cadence: cadence, fn: fn})
}

// Start kicks off a goroutine per registered handler. Safe to call only
// once per Service.
func (s *Service) Start(parent context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return fmt.Errorf("tick service already started")
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	for _, spec := range s.handlers {
		s.wg.Add(1)
		go s.run(ctx, spec)
	}
	s.started = true
	log.Info("Tick service started with %d handler(s)", len(s.handlers))
	return nil
}

// Stop cancels the parent context and waits for every handler goroutine to
// return. Safe to call multiple times.
func (s *Service) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	s.wg.Wait()
	log.Info("Tick service stopped")
}

// run owns one handler's goroutine lifecycle. Each handler sees its own
// ticker so a slow handler can't back up fast ones.
func (s *Service) run(ctx context.Context, spec handlerSpec) {
	defer s.wg.Done()
	ticker := time.NewTicker(spec.cadence)
	defer ticker.Stop()
	log.Debug("Tick handler %q running every %s", spec.name, spec.cadence)
	for {
		select {
		case <-ctx.Done():
			log.Debug("Tick handler %q exiting: %v", spec.name, ctx.Err())
			return
		case <-ticker.C:
			if err := spec.fn(ctx); err != nil {
				log.Error("Tick handler %q failed: %v", spec.name, err)
				// Keep going — one bad tick shouldn't kill the handler.
			}
		}
	}
}
