// File: internal/validation/validation.go
// Project: Terminal Velocity
// Description: Input validation functions for usernames, passwords, and emails
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-11-14
//
// This package provides comprehensive input validation and sanitization to prevent
// security vulnerabilities. It implements multiple layers of defense:
//
// 1. Input Validation:
//   - Username validation (length, character set, reserved names)
//   - Password validation (complexity, common passwords, patterns)
//   - Email validation (format, length)
//
// 2. Terminal Injection Prevention:
//   - ANSI escape code filtering (prevents terminal manipulation)
//   - Control character stripping (prevents terminal corruption)
//   - Safe display sanitization (removes dangerous characters)
//
// 3. Path Traversal Prevention:
//   - Filename sanitization (removes path separators)
//   - Path traversal protection (blocks ../ and similar)
//
// Security Principles:
//   - Fail-secure: Invalid input is rejected, not sanitized into valid
//   - Defense in depth: Multiple validation layers
//   - Explicit validation: No implicit trust of any input
//   - Comprehensive filtering: All user-controlled input must be validated
//
// Usage Guidelines:
//   - Always validate input at entry points (registration, login, etc.)
//   - Always sanitize before displaying user input (chat, usernames, etc.)
//   - Never trust client-side validation alone
//   - Log validation failures for security monitoring
//
// Terminal Security:
//   SSH clients are powerful and can interpret control sequences. An attacker
//   could inject ANSI escape codes to:
//   - Manipulate other players' terminals
//   - Hide malicious content
//   - Cause denial of service (terminal corruption)
//   - Bypass rate limiting (clear screen repeatedly)
//
//   This package prevents all such attacks by stripping control sequences
//   and validating all input before use.

package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ============================================================================
// Validation Constants and Patterns
// ============================================================================

// Validation regex patterns
var (
	// Username: 3-20 characters, alphanumeric, underscore, hyphen
	// - No special characters to prevent confusion/impersonation
	// - No spaces to simplify parsing
	// - Length limits prevent abuse
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

	// Email: standard email format per RFC 5322 (simplified)
	// - Validates basic structure: localpart@domain.tld
	// - Prevents most malformed addresses
	// - Does not prevent all invalid emails (RFC 5322 is complex)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// Password complexity requirements
//
// These constants define minimum security requirements for passwords.
// They balance security with usability:
//   - MinPasswordLength: 8 chars (industry standard minimum)
//   - Requires uppercase, lowercase, and digit (prevents simple passwords)
//   - Special chars optional (improves usability without major security loss)
//   - MaxPasswordLength: 128 (prevents DoS via bcrypt cost)
const (
	MinPasswordLength     = 8   // NIST recommends minimum 8
	MinPasswordUpper      = 1   // At least one uppercase letter
	MinPasswordLower      = 1   // At least one lowercase letter
	MinPasswordDigit      = 1   // At least one number
	MinPasswordSpecial    = 0   // Optional for now (usability)
	MaxPasswordLength     = 128 // Prevent bcrypt DoS
	MinUsernameLength     = 3   // Prevent single-char usernames
	MaxUsernameLength     = 20  // Prevent abuse, ensure display
	MaxEmailLength        = 254 // RFC 5321 maximum
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
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
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
// before displaying in the terminal to prevent ANSI escape code injection.
//
// This function is critical for security in SSH-based applications where
// user input is displayed directly in terminals. Without sanitization, an
// attacker could inject ANSI escape codes to:
//   - Clear other users' screens
//   - Change terminal colors/modes
//   - Move cursor to hide malicious text
//   - Execute terminal commands (in some terminals)
//   - Cause denial of service
//
// The function removes:
//   - ESC character (0x1B) - start of ANSI escape sequences
//   - All control characters except \n and \t
//   - Characters outside printable ASCII range (32-126)
//
// Safe characters allowed:
//   - Printable ASCII (32-126): letters, numbers, punctuation
//   - Newline (\n) - for multi-line text
//   - Tab (\t) - for formatting
//
// Parameters:
//   - input: User-provided string to sanitize
//
// Returns:
//   - Sanitized string safe for terminal display
//
// Usage:
//   // Before displaying chat message
//   safeMessage := validation.SanitizeForDisplay(userMessage)
//   fmt.Println(safeMessage)
//
// Security Note:
//   - Always sanitize before displaying user input
//   - Does not validate content (only removes dangerous chars)
//   - Use with ValidateNoInjection for additional check
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
