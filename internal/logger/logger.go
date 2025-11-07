// File: internal/logger/logger.go
// Project: Terminal Velocity
// Description: Structured logging with configurable levels
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a Level
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Logger is a structured logger with configurable levels
type Logger struct {
	level      Level
	logger     *log.Logger
	mu         sync.Mutex
	file       *os.File
	component  string
	withCaller bool
}

// Config holds logger configuration
type Config struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	ToStdout   bool
	WithCaller bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(cfg)
	})
	return err
}

// New creates a new Logger instance
func New(cfg Config) (*Logger, error) {
	level := ParseLevel(cfg.Level)

	var writers []io.Writer
	var file *os.File
	var err error

	// Set up file output if specified
	if cfg.FilePath != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err = os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, file)
	}

	// Add stdout if configured
	if cfg.ToStdout || cfg.FilePath == "" {
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer
	writer := io.MultiWriter(writers...)

	return &Logger{
		level:      level,
		logger:     log.New(writer, "", 0),
		file:       file,
		withCaller: cfg.WithCaller,
	}, nil
}

// WithComponent returns a new logger with a component name
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:      l.level,
		logger:     l.logger,
		file:       l.file,
		component:  component,
		withCaller: l.withCaller,
	}
}

// Close closes the log file if open
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a log message with the specified level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	var logLine string
	if l.component != "" {
		logLine = fmt.Sprintf("[%s] %s [%s] %s", timestamp, level.String(), l.component, msg)
	} else {
		logLine = fmt.Sprintf("[%s] %s %s", timestamp, level.String(), msg)
	}

	// Add caller information if enabled
	if l.withCaller && level >= LevelError {
		if _, file, line, ok := runtime.Caller(2); ok {
			logLine = fmt.Sprintf("%s (at %s:%d)", logLine, filepath.Base(file), line)
		}
	}

	l.logger.Println(logLine)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Default logger convenience functions

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	}
	os.Exit(1)
}

// WithComponent returns a logger with a component name
func WithComponent(component string) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithComponent(component)
	}
	// Return a basic logger if default is not initialized
	l, _ := New(Config{Level: "info", ToStdout: true})
	return l.WithComponent(component)
}

// SetLevel changes the logging level of the default logger
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}

// Close closes the default logger
func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
