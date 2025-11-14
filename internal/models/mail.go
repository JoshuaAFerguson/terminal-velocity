// File: internal/models/mail.go
// Project: Terminal Velocity
// Description: Data models for player mail system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Mail represents a message between players
type Mail struct {
	ID      uuid.UUID `json:"id"`
	From    uuid.UUID `json:"from"`     // Sender player ID
	To      uuid.UUID `json:"to"`       // Recipient player ID
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	SentAt  time.Time `json:"sent_at"`
	Read    bool      `json:"read"`
	ReadAt  *time.Time `json:"read_at,omitempty"`

	// Soft delete - track which players have deleted this mail
	DeletedBy []uuid.UUID `json:"deleted_by,omitempty"`
}

// IsDeletedBy checks if mail has been deleted by a specific player
func (m *Mail) IsDeletedBy(playerID uuid.UUID) bool {
	for _, id := range m.DeletedBy {
		if id == playerID {
			return true
		}
	}
	return false
}

// MarkAsDeleted adds a player to the deleted list
func (m *Mail) MarkAsDeleted(playerID uuid.UUID) {
	if !m.IsDeletedBy(playerID) {
		m.DeletedBy = append(m.DeletedBy, playerID)
	}
}

// CanBeHardDeleted returns true if both sender and recipient have deleted the mail
func (m *Mail) CanBeHardDeleted() bool {
	return m.IsDeletedBy(m.From) && m.IsDeletedBy(m.To)
}

// MailStats represents mail statistics for a player
type MailStats struct {
	TotalReceived int `json:"total_received"`
	TotalSent     int `json:"total_sent"`
	Unread        int `json:"unread"`
}

// Common mail errors
var (
	ErrMailNotFound    = errors.New("mail not found")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrInvalidRecipient = errors.New("invalid recipient")
	ErrSubjectRequired = errors.New("subject is required")
	ErrBodyRequired    = errors.New("body is required")
)
