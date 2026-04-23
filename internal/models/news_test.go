// File: internal/models/news_test.go
// Project: Terminal Velocity
// Description: Unit tests for the news article generators.
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-23

package models

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateSystemEventNews_ReferencesSystemName(t *testing.T) {
	id := uuid.New()
	const name = "Testopia"
	for i := 0; i < 40; i++ {
		article := GenerateSystemEventNews(id, name)
		if article == nil {
			t.Fatalf("iteration %d: nil article", i)
		}
		if !strings.Contains(article.Headline, name) {
			t.Errorf("iteration %d: headline %q missing system name %q", i, article.Headline, name)
		}
		if !strings.Contains(article.Body, name) {
			t.Errorf("iteration %d: body missing system name %q (body=%q)", i, name, article.Body)
		}
		if article.SystemID == nil || *article.SystemID != id {
			t.Errorf("iteration %d: SystemID not persisted (got %v)", i, article.SystemID)
		}
	}
}

func TestGenerateFactionNews_IndicesAlignedAcrossBranches(t *testing.T) {
	// Regression guard for the headlines/bodies length mismatch that
	// panicked GenerateInitialNews when called at server start. Runs the
	// generator many times so any off-by-one on either branch would
	// eventually hit an out-of-range index.
	for i := 0; i < 200; i++ {
		if a := GenerateFactionNews("Alice", "Bob", i%2 == 0); a == nil {
			t.Fatalf("iteration %d: nil article", i)
		}
	}
}
