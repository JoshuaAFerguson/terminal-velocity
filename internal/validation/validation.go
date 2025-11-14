// File: internal/validation/validation.go
// Project: Terminal Velocity
// Description: Input validation functions for usernames, passwords, and emails
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-14

package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Validation regex patterns
var (
	// Username: 3-20 characters, alphanumeric, underscore, hyphen
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

	// Email: standard email format
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// Password complexity requirements
const (
	MinPasswordLength     = 8
	MinPasswordUpper      = 1
	MinPasswordLower      = 1
	MinPasswordDigit      = 1
	MinPasswordSpecial    = 0 // Optional for now
	MaxPasswordLength     = 128
	MinUsernameLength     = 3
	MaxUsernameLength     = 20
	MaxEmailLength        = 254
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateUsername validates a username against security requirements
func ValidateUsername(username string) error {
	if username == "" {
		return &ValidationError{Field: "username", Message: "username is required"}
	}

	if len(username) < MinUsernameLength {
		return &ValidationError{
			Field:   "username",
			Message: fmt.Sprintf("username must be at least %d characters", MinUsernameLength),
		}
	}

	if len(username) > MaxUsernameLength {
		return &ValidationError{
			Field:   "username",
			Message: fmt.Sprintf("username must be no more than %d characters", MaxUsernameLength),
		}
	}

	if !usernameRegex.MatchString(username) {
		return &ValidationError{
			Field:   "username",
			Message: "username can only contain letters, numbers, underscore, and hyphen",
		}
	}

	// Check for reserved/prohibited usernames
	if isReservedUsername(username) {
		return &ValidationError{
			Field:   "username",
			Message: "this username is reserved and cannot be used",
		}
	}

	return nil
}

// ValidatePassword validates a password against security requirements
func ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < MinPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters", MinPasswordLength),
		}
	}

	if len(password) > MaxPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be no more than %d characters", MaxPasswordLength),
		}
	}

	// Count character types
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Check complexity requirements
	var requirements []string

	if !hasUpper {
		requirements = append(requirements, "at least one uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "at least one lowercase letter")
	}
	if !hasDigit {
		requirements = append(requirements, "at least one number")
	}

	if len(requirements) > 0 {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must contain %s", strings.Join(requirements, ", ")),
		}
	}

	// Check for common weak passwords
	if isCommonPassword(password) {
		return &ValidationError{
			Field:   "password",
			Message: "this password is too common, please choose a stronger one",
		}
	}

	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	if len(email) > MaxEmailLength {
		return &ValidationError{
			Field:   "email",
			Message: fmt.Sprintf("email must be no more than %d characters", MaxEmailLength),
		}
	}

	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	return nil
}

// ValidateEmailOptional validates an email address if provided
func ValidateEmailOptional(email string) error {
	if email == "" {
		return nil // Email is optional
	}
	return ValidateEmail(email)
}

// isReservedUsername checks if a username is reserved
func isReservedUsername(username string) bool {
	// Convert to lowercase for comparison
	lower := strings.ToLower(username)

	reserved := []string{
		"admin", "administrator", "root", "system", "moderator",
		"mod", "superadmin", "sysadmin", "support", "help",
		"server", "bot", "npc", "null", "undefined",
		"anonymous", "guest", "user", "player", "test",
		"official", "staff", "team", "owner",
	}

	for _, r := range reserved {
		if lower == r {
			return true
		}
	}

	return false
}

// isCommonPassword checks if a password is in the list of common passwords
func isCommonPassword(password string) bool {
	// Convert to lowercase for comparison
	lower := strings.ToLower(password)

	// Common weak passwords
	common := []string{
		"password", "12345678", "qwerty", "abc123", "letmein",
		"welcome", "monkey", "password1", "123456789", "1234567890",
		"iloveyou", "princess", "rockyou", "654321", "sunshine",
		"admin123", "test1234", "password123", "welcome123",
	}

	for _, c := range common {
		if lower == c {
			return true
		}
	}

	// Check for patterns like "aaaaaaaa" or "12345678"
	if hasRepeatingChars(password, 4) {
		return true
	}

	if hasSequentialChars(password, 5) {
		return true
	}

	return false
}

// hasRepeatingChars checks if password has repeating characters
func hasRepeatingChars(s string, minRepeat int) bool {
	if len(s) < minRepeat {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= minRepeat {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// hasSequentialChars checks if password has sequential characters
func hasSequentialChars(s string, minSeq int) bool {
	if len(s) < minSeq {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		// Check if characters are sequential (e.g., "abc", "123")
		if s[i] == s[i-1]+1 || s[i] == s[i-1]-1 {
			count++
			if count >= minSeq {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// GetPasswordStrength returns a strength score and description for a password
func GetPasswordStrength(password string) (score int, description string) {
	if len(password) == 0 {
		return 0, "No password"
	}

	score = 0

	// Length score (0-40 points)
	if len(password) >= 8 {
		score += 10
	}
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}
	if len(password) >= 20 {
		score += 10
	}

	// Complexity score (0-40 points)
	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasDigit {
		score += 10
	}
	if hasSpecial {
		score += 10
	}

	// Deductions (0-20 points)
	if hasRepeatingChars(password, 3) {
		score -= 10
	}
	if hasSequentialChars(password, 4) {
		score -= 10
	}
	if isCommonPassword(password) {
		score -= 20
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// Determine description
	switch {
	case score < 30:
		description = "Weak"
	case score < 50:
		description = "Fair"
	case score < 70:
		description = "Good"
	case score < 90:
		description = "Strong"
	default:
		description = "Excellent"
	}

	return score, description
}

// ============================================================================
// Terminal Injection Prevention
// ============================================================================

// SanitizeForDisplay removes potentially dangerous characters from user input
// before displaying in the terminal to prevent ANSI escape code injection
func SanitizeForDisplay(input string) string {
	if input == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(input))

	for _, r := range input {
		// Allow printable ASCII and common safe characters
		if r >= 32 && r <= 126 {
			// Skip ANSI escape sequence start
			if r == 27 { // ESC character
				continue
			}
			result.WriteRune(r)
		} else if r == '\n' || r == '\t' {
			// Allow newlines and tabs (but be cautious where used)
			result.WriteRune(r)
		}
		// All other control characters are stripped
	}

	return result.String()
}

// StripANSI removes all ANSI escape codes from a string
func StripANSI(input string) string {
	// Regex to match ANSI escape codes
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(input, "")
}

// SanitizeUsername sanitizes a username for safe display
// This is more aggressive than validation and ensures safe display
func SanitizeUsername(username string) string {
	// First validate structure
	if err := ValidateUsername(username); err != nil {
		return "[invalid]"
	}

	// Then sanitize for display
	return SanitizeForDisplay(username)
}

// SanitizeChatMessage sanitizes a chat message for safe display
// Allows more characters than username but still prevents injection
func SanitizeChatMessage(message string) string {
	if len(message) == 0 {
		return ""
	}

	// Limit message length
	const maxMessageLength = 500
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength]
	}

	// Strip ANSI codes first
	message = StripANSI(message)

	// Remove control characters except newline and tab
	var result strings.Builder
	result.Grow(len(message))

	for _, r := range message {
		if r >= 32 && r <= 126 {
			result.WriteRune(r)
		} else if r == '\n' {
			// Allow newlines but limit consecutive ones
			result.WriteRune(r)
		} else if r == '\t' {
			// Convert tabs to spaces
			result.WriteRune(' ')
		}
		// Skip other control characters
	}

	return result.String()
}

// IsTerminalInjection checks if input contains potential terminal injection
func IsTerminalInjection(input string) bool {
	// Check for ANSI escape codes
	if strings.Contains(input, "\x1b[") {
		return true
	}

	// Check for ESC character
	if strings.Contains(input, "\x1b") {
		return true
	}

	// Check for other control sequences
	controlSequences := []string{
		"\x00", // NULL
		"\x07", // BEL (bell)
		"\x08", // Backspace
		"\x0c", // Form feed
		"\x1b]", // OSC (Operating System Command)
		"\x9b", // CSI (Control Sequence Introducer)
	}

	for _, seq := range controlSequences {
		if strings.Contains(input, seq) {
			return true
		}
	}

	return false
}

// SanitizeFilename sanitizes a filename to prevent path traversal
func SanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	// Remove control characters
	filename = SanitizeForDisplay(filename)

	// Limit length
	const maxFilenameLength = 255
	if len(filename) > maxFilenameLength {
		filename = filename[:maxFilenameLength]
	}

	return filename
}

// ValidateNoInjection validates that input doesn't contain injection attempts
func ValidateNoInjection(field, value string) error {
	if IsTerminalInjection(value) {
		return &ValidationError{
			Field:   field,
			Message: "input contains potentially dangerous characters",
		}
	}

	return nil
}
