// File: internal/models/mail.go
// Project: Terminal Velocity
// Description: Data models for player mail system
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package models

import (
	"errors"
)

// NOTE: Mail struct is defined in social.go as part of the social features
// This file contains mail-related helper types and error definitions

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
