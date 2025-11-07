// File: internal/logger/logger_test.go
// Project: Terminal Velocity
// Description: Tests for structured logging
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package logger

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"unknown", LevelInfo}, // default
	}

	for _, tt := range tests {
		result := ParseLevel(tt.input)
		if result != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, result, tt.expected)
		}
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  LevelInfo,
		logger: log.New(&buf, "", 0),
	}

	// Debug should not be logged
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Errorf("Debug message was logged when level is Info")
	}

	// Info should be logged
	logger.Info("info message")
	if !strings.Contains(buf.String(), "INFO") || !strings.Contains(buf.String(), "info message") {
		t.Errorf("Info message not logged correctly: %q", buf.String())
	}

	buf.Reset()

	// Warn should be logged
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "WARN") || !strings.Contains(buf.String(), "warn message") {
		t.Errorf("Warn message not logged correctly: %q", buf.String())
	}

	buf.Reset()

	// Error should be logged
	logger.Error("error message")
	if !strings.Contains(buf.String(), "ERROR") || !strings.Contains(buf.String(), "error message") {
		t.Errorf("Error message not logged correctly: %q", buf.String())
	}
}

func TestLoggerWithComponent(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  LevelInfo,
		logger: log.New(&buf, "", 0),
	}

	componentLogger := logger.WithComponent("TestComponent")
	componentLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[TestComponent]") {
		t.Errorf("Component name not included in log: %q", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Message not included in log: %q", output)
	}
}

func TestLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  LevelInfo,
		logger: log.New(&buf, "", 0),
	}

	// Debug should not be logged at Info level
	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Errorf("Debug message logged at Info level")
	}

	// Change to Debug level
	logger.SetLevel(LevelDebug)

	// Now debug should be logged
	logger.Debug("should appear")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Errorf("Debug message not logged after changing level: %q", buf.String())
	}
}

func TestNewLogger(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:    "debug",
		FilePath: logFile,
		ToStdout: false,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log a message
	logger.Info("test message")

	// Check that file was created and contains the message
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "test message") {
		t.Errorf("Log file does not contain expected message: %q", output)
	}
}

func TestLoggerMultiWriter(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:    "info",
		FilePath: logFile,
		ToStdout: false, // Don't redirect stdout in tests
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message")
	logger.Close()

	// Check file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	fileOutput := string(data)
	if !strings.Contains(fileOutput, "test message") {
		t.Errorf("Log file does not contain message: %q", fileOutput)
	}
	if !strings.Contains(fileOutput, "INFO") {
		t.Errorf("Log file does not contain log level: %q", fileOutput)
	}
}
