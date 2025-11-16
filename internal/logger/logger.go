// File: internal/logger/logger.go
// Project: Terminal Velocity
// Description: Structured logging with configurable levels and file rotation support
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package logger provides centralized structured logging infrastructure for Terminal Velocity.
//
// This package implements a thread-safe logging system with multiple severity levels,
// component-based logging, and flexible output options. It supports both file and
// stdout output with optional caller information for debugging.
//
// Features:
//   - Five log levels: Debug, Info, Warn, Error, Fatal
//   - Thread-safe concurrent logging via sync.Mutex
//   - Multi-writer output (file + stdout simultaneously)
//   - Component-based logging for better organization
//   - Optional caller information (file:line) for errors
//   - Configurable log level filtering
//   - Singleton default logger for package-level functions
//
// Usage Example:
//
//	// Initialize default logger
//	err := logger.Init(logger.Config{
//	    Level: "info",
//	    FilePath: "/var/log/terminal-velocity/app.log",
//	    ToStdout: true,
//	    WithCaller: true,
//	})
//
//	// Use package-level functions
//	logger.Info("Server starting on port %d", 2222)
//	logger.Error("Failed to connect: %v", err)
//
//	// Create component-specific logger
//	authLog := logger.WithComponent("auth")
//	authLog.Debug("Processing login for user: %s", username)
//
//	// Create custom logger instance
//	customLogger, err := logger.New(logger.Config{
//	    Level: "debug",
//	    FilePath: "/tmp/debug.log",
//	})
//	customLogger.Debug("Custom logger message")
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

// Level represents the severity level of a log message.
//
// Log levels are ordered from least to most severe:
//   LevelDebug < LevelInfo < LevelWarn < LevelError < LevelFatal
//
// When a logger is configured with a specific level, only messages
// at that level or higher will be logged. For example, a logger set
// to LevelInfo will log Info, Warn, Error, and Fatal messages, but
// will suppress Debug messages.
type Level int

const (
	// LevelDebug is the most verbose level, used for detailed diagnostic information.
	// Typically disabled in production to reduce log volume.
	LevelDebug Level = iota

	// LevelInfo is used for general informational messages about normal operations.
	// This is the default level for most production deployments.
	LevelInfo

	// LevelWarn indicates potentially harmful situations or unexpected conditions
	// that don't prevent the application from continuing.
	LevelWarn

	// LevelError indicates error conditions that should be investigated.
	// The application can continue but some functionality may be impaired.
	// When WithCaller is enabled, caller information (file:line) is included.
	LevelError

	// LevelFatal indicates severe errors that cause the application to terminate.
	// Logging at this level triggers os.Exit(1) after writing the message.
	// When WithCaller is enabled, caller information (file:line) is included.
	LevelFatal
)

// String returns the uppercase string representation of the log level.
//
// This is used when formatting log messages to include the severity level.
//
// Returns:
//   - "DEBUG", "INFO", "WARN", "ERROR", or "FATAL" for valid levels
//   - "UNKNOWN" for invalid/unrecognized levels
//
// Example output in log:
//   [2025-01-07 15:04:05.000] INFO Server started successfully
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

// ParseLevel converts a string to a Level constant.
//
// The function performs case-insensitive matching and accepts common
// variations of level names. If the input doesn't match any known level,
// it returns LevelInfo as a safe default.
//
// Parameters:
//   - s: The level name to parse (case-insensitive)
//
// Returns:
//   - The corresponding Level constant
//   - LevelInfo if the string doesn't match any known level
//
// Supported inputs:
//   - "debug" -> LevelDebug
//   - "info" -> LevelInfo
//   - "warn" or "warning" -> LevelWarn
//   - "error" -> LevelError
//   - "fatal" -> LevelFatal
//   - anything else -> LevelInfo (default)
//
// Example:
//   level := logger.ParseLevel("error")  // Returns LevelError
//   level := logger.ParseLevel("DEBUG")  // Returns LevelDebug (case-insensitive)
//   level := logger.ParseLevel("invalid") // Returns LevelInfo (safe default)
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

// Logger is a thread-safe structured logger with configurable severity levels.
//
// Logger provides formatted, timestamped logging to multiple outputs (file and/or stdout)
// with optional component tagging and caller information. It supports dynamic level
// filtering to control log verbosity at runtime.
//
// Thread Safety:
//   All methods are safe for concurrent use. The logger uses a sync.Mutex to protect
//   shared state during log writes and level changes.
//
// Fields:
//   - level: Current minimum severity level for logged messages
//   - logger: Underlying Go standard library logger for I/O
//   - mu: Mutex protecting concurrent access to logger state
//   - file: Optional log file handle (nil if logging only to stdout)
//   - component: Optional component name for tagged logging (e.g., "auth", "database")
//   - withCaller: Whether to include caller file:line for Error/Fatal levels
//
// Log Format:
//   Without component: [2025-01-07 15:04:05.000] INFO message here
//   With component:    [2025-01-07 15:04:05.000] INFO [auth] message here
//   With caller:       [2025-01-07 15:04:05.000] ERROR message here (at server.go:123)
type Logger struct {
	level      Level
	logger     *log.Logger
	mu         sync.Mutex
	file       *os.File
	component  string
	withCaller bool
}

// Config holds configuration options for creating a new Logger.
//
// Fields:
//   - Level: Minimum log level as a string ("debug", "info", "warn", "error", "fatal")
//            Defaults to "info" if invalid or empty
//   - FilePath: Path to log file. If empty, logs only to stdout (unless ToStdout is false)
//               Parent directories are created automatically if they don't exist
//   - MaxSizeMB: Maximum log file size in megabytes before rotation (currently unused)
//                Reserved for future log rotation implementation
//   - MaxBackups: Maximum number of old log files to retain (currently unused)
//                 Reserved for future log rotation implementation
//   - ToStdout: Whether to write logs to stdout in addition to (or instead of) file
//               If true and FilePath is set, logs go to both destinations
//               If true and FilePath is empty, logs go only to stdout
//               If false and FilePath is empty, no logs are written (not recommended)
//   - WithCaller: Whether to include caller information (file:line) for Error and Fatal logs
//                 Useful for debugging but adds slight performance overhead
//
// Example:
//   cfg := logger.Config{
//       Level: "info",
//       FilePath: "/var/log/app.log",
//       ToStdout: true,
//       WithCaller: true,
//   }
type Config struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	ToStdout   bool
	WithCaller bool
}

var (
	// defaultLogger is the singleton logger instance used by package-level functions.
	// It is initialized once via Init() using sync.Once to ensure thread-safety.
	// If not initialized, package-level functions will create a fallback logger.
	defaultLogger *Logger

	// once ensures that defaultLogger is initialized exactly once, even if Init()
	// is called multiple times from different goroutines.
	once sync.Once
)

// Init initializes the default singleton logger instance.
//
// This function should be called once during application startup to configure
// the logger used by package-level functions (Debug, Info, Warn, Error, Fatal).
// Subsequent calls to Init are ignored due to sync.Once protection.
//
// Parameters:
//   - cfg: Logger configuration specifying level, output destinations, and options
//
// Returns:
//   - error: Any error encountered while creating the logger (e.g., file open failure)
//            nil if initialization succeeds or if logger was already initialized
//
// Thread Safety:
//   Safe for concurrent calls. Only the first call will initialize the logger.
//
// Example:
//   err := logger.Init(logger.Config{
//       Level: "info",
//       FilePath: "/var/log/app.log",
//       ToStdout: true,
//       WithCaller: true,
//   })
//   if err != nil {
//       log.Fatal("Failed to initialize logger:", err)
//   }
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(cfg)
	})
	return err
}

// New creates a new Logger instance with the specified configuration.
//
// This function creates an independent logger that can be used alongside or
// instead of the default logger. It's useful for creating specialized loggers
// with different configurations (e.g., separate debug logger, per-module loggers).
//
// The function:
//   1. Parses the log level from the config
//   2. Creates the log file (if FilePath is specified) and its parent directories
//   3. Sets up multi-writer for simultaneous file and stdout output
//   4. Returns a configured Logger instance
//
// Parameters:
//   - cfg: Logger configuration specifying level, output destinations, and options
//
// Returns:
//   - *Logger: A new logger instance configured according to cfg
//   - error: Non-nil if log directory creation or file opening fails
//
// Thread Safety:
//   The returned Logger is safe for concurrent use via internal mutex.
//
// Error Conditions:
//   - Directory creation failure: Returns error if parent directories can't be created
//   - File open failure: Returns error if log file can't be opened/created
//
// Example:
//   // Create a debug logger for development
//   debugLogger, err := logger.New(logger.Config{
//       Level: "debug",
//       FilePath: "/tmp/debug.log",
//       ToStdout: false,
//   })
//   if err != nil {
//       return fmt.Errorf("failed to create debug logger: %w", err)
//   }
//   defer debugLogger.Close()
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

// WithComponent returns a new logger instance that tags all messages with a component name.
//
// This is useful for organizing logs from different parts of the application without
// creating entirely separate loggers. The component name appears in brackets in the log output.
//
// The returned logger shares the same underlying writer, log level, and configuration
// as the parent logger, but adds the component tag to all messages.
//
// Parameters:
//   - component: The component name to tag messages with (e.g., "auth", "database", "combat")
//
// Returns:
//   - *Logger: A new logger instance that includes the component tag in all messages
//
// Thread Safety:
//   Safe to call concurrently. The returned logger has its own state but shares I/O.
//
// Log Format Change:
//   Without component: [2025-01-07 15:04:05.000] INFO message here
//   With component:    [2025-01-07 15:04:05.000] INFO [auth] message here
//
// Example:
//   authLog := mainLogger.WithComponent("auth")
//   authLog.Info("User %s authenticated", username)
//   // Output: [2025-01-07 15:04:05.000] INFO [auth] User alice authenticated
//
//   dbLog := mainLogger.WithComponent("database")
//   dbLog.Debug("Query executed in %dms", duration)
//   // Output: [2025-01-07 15:04:05.000] DEBUG [database] Query executed in 42ms
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:      l.level,
		logger:     l.logger,
		file:       l.file,
		component:  component,
		withCaller: l.withCaller,
	}
}

// Close closes the log file handle if one was opened during logger creation.
//
// This should be called when the logger is no longer needed to release file
// resources. It's safe to call even if the logger doesn't have an open file
// (e.g., stdout-only loggers).
//
// Note: After calling Close, the logger should not be used for further logging
// as the file handle is no longer valid.
//
// Returns:
//   - error: Any error from closing the file, or nil if no file was open
//
// Thread Safety:
//   Safe to call concurrently, but behavior is undefined if logging continues
//   after Close is called.
//
// Example:
//   logger, err := logger.New(cfg)
//   if err != nil {
//       return err
//   }
//   defer logger.Close()  // Ensure file is closed on function exit
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log is the internal method that writes a formatted log message at the specified level.
//
// This method is called by all public logging methods (Debug, Info, Warn, Error, Fatal).
// It handles level filtering, timestamp generation, message formatting, component tagging,
// optional caller information, and thread-safe output.
//
// The method performs the following steps:
//   1. Filters out messages below the current log level
//   2. Acquires mutex lock for thread-safe access
//   3. Generates timestamp in millisecond precision
//   4. Formats the message with provided arguments
//   5. Adds component tag if set
//   6. Adds caller information (file:line) for Error/Fatal if WithCaller is enabled
//   7. Writes the formatted log line to configured outputs
//
// Parameters:
//   - level: The severity level of this message
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Thread-safe via sync.Mutex. Multiple goroutines can log concurrently without
//   interleaving output.
//
// Performance:
//   Messages below the current log level return immediately without formatting.
//   Caller information adds ~2 stack frame lookups for Error/Fatal levels.
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

// Debug logs a debug message at the Debug level.
//
// Debug messages are used for detailed diagnostic information useful during
// development and troubleshooting. They are typically disabled in production
// to reduce log volume.
//
// The message is only written if the logger's current level is LevelDebug.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Debug("Processing request from %s with ID %d", addr, requestID)
//   // Output: [2025-01-07 15:04:05.000] DEBUG Processing request from 192.168.1.1 with ID 42
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an informational message at the Info level.
//
// Info messages are used for general informational events that highlight the
// progress of the application at a coarse-grained level. This is the default
// level for most production deployments.
//
// The message is only written if the logger's level is Info or lower (Debug/Info).
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Info("Server started on port %d", port)
//   // Output: [2025-01-07 15:04:05.000] INFO Server started on port 2222
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message at the Warn level.
//
// Warn messages indicate potentially harmful situations or unexpected conditions
// that don't prevent the application from continuing. Examples include deprecated
// API usage, poor use of API, or recoverable errors.
//
// The message is only written if the logger's level is Warn or lower.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Warn("Database query took %dms, exceeding recommended threshold", duration)
//   // Output: [2025-01-07 15:04:05.000] WARN Database query took 5000ms, exceeding recommended threshold
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message at the Error level.
//
// Error messages indicate error conditions that should be investigated. The
// application can continue but some functionality may be impaired. If WithCaller
// is enabled in the logger config, caller information (file:line) is automatically
// included.
//
// The message is only written if the logger's level is Error or lower.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Error("Failed to connect to database: %v", err)
//   // Output: [2025-01-07 15:04:05.000] ERROR Failed to connect to database: connection refused (at server.go:123)
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal error message and terminates the application with exit code 1.
//
// Fatal messages indicate severe errors that prevent the application from continuing.
// After logging the message, this method calls os.Exit(1) to terminate the process
// immediately. Deferred functions will NOT run.
//
// If WithCaller is enabled, caller information (file:line) is included.
//
// This method is always written regardless of logger level.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use, though application terminates immediately after.
//
// Important:
//   Deferred functions will NOT execute after Fatal is called. Use Error instead
//   if you need to perform cleanup before exiting.
//
// Example:
//   logger.Fatal("Cannot start server: %v", err)
//   // Output: [2025-01-07 15:04:05.000] FATAL Cannot start server: permission denied (at main.go:45)
//   // Application exits with code 1
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// SetLevel dynamically changes the minimum logging level.
//
// This allows runtime adjustment of log verbosity without restarting the application.
// For example, you could temporarily enable Debug logging to troubleshoot an issue,
// then return to Info level for normal operation.
//
// Messages below the new level will be suppressed starting immediately.
//
// Parameters:
//   - level: The new minimum level (LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal)
//
// Thread Safety:
//   Safe for concurrent use. Protected by mutex.
//
// Example:
//   logger.SetLevel(logger.LevelDebug)  // Enable debug logging
//   // ... perform debugging operations ...
//   logger.SetLevel(logger.LevelInfo)   // Return to normal verbosity
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current minimum logging level.
//
// This can be used to check the current verbosity setting, for example to
// conditionally perform expensive operations only when debug logging is enabled.
//
// Returns:
//   - Level: The current minimum log level
//
// Thread Safety:
//   Safe for concurrent use. Protected by mutex.
//
// Example:
//   if logger.GetLevel() <= logger.LevelDebug {
//       // Only build expensive debug message if it will be logged
//       details := buildDetailedDebugInfo()
//       logger.Debug("Details: %s", details)
//   }
func (l *Logger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// ============================================================================
// Package-Level Convenience Functions
//
// The following functions provide a convenient interface to the default logger
// singleton. They should be used for most application logging after calling
// Init() during startup.
//
// If the default logger is not initialized (Init was not called), the Debug,
// Info, Warn, Error functions silently do nothing. The Fatal function will
// still exit the application.
// ============================================================================

// Debug logs a debug message using the default logger.
//
// This is a convenience function equivalent to calling defaultLogger.Debug().
// If Init() has not been called, this function does nothing.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "debug"})
//   logger.Debug("Processing request ID: %d", reqID)
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger.
//
// This is a convenience function equivalent to calling defaultLogger.Info().
// If Init() has not been called, this function does nothing.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "info"})
//   logger.Info("Server started on port %d", port)
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger.
//
// This is a convenience function equivalent to calling defaultLogger.Warn().
// If Init() has not been called, this function does nothing.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "warn"})
//   logger.Warn("Deprecated API usage detected: %s", apiName)
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger.
//
// This is a convenience function equivalent to calling defaultLogger.Error().
// If Init() has not been called, this function does nothing.
//
// If WithCaller was enabled during Init(), caller information is included.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "error", WithCaller: true})
//   logger.Error("Database connection failed: %v", err)
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

// Fatal logs a fatal message and exits using the default logger.
//
// This is a convenience function equivalent to calling defaultLogger.Fatal().
// Unlike other convenience functions, this ALWAYS exits the application with
// code 1, even if Init() was not called.
//
// If the default logger is initialized, the message is logged before exiting.
// If not initialized, the application still exits but the message is not logged.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments to format into the message
//
// Thread Safety:
//   Safe for concurrent use, though application terminates immediately.
//
// Important:
//   Deferred functions will NOT execute. Use Error + explicit exit if cleanup is needed.
//
// Example:
//   logger.Init(logger.Config{Level: "info"})
//   logger.Fatal("Critical failure: %v", err)
//   // Application exits with code 1
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	}
	os.Exit(1)
}

// WithComponent returns a component-tagged logger based on the default logger.
//
// This function creates a new logger that tags all messages with the specified
// component name. If the default logger is initialized, it returns a logger
// sharing the default logger's configuration. If not initialized, it creates
// a minimal fallback logger to stdout.
//
// The fallback behavior ensures that component loggers can be created even if
// Init() hasn't been called, though this is not recommended for production use.
//
// Parameters:
//   - component: The component name to tag messages with
//
// Returns:
//   - *Logger: A logger instance that tags all messages with the component name
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "info"})
//   authLog := logger.WithComponent("auth")
//   authLog.Info("User authenticated: %s", username)
//   // Output: [2025-01-07 15:04:05.000] INFO [auth] User authenticated: alice
//
//   dbLog := logger.WithComponent("database")
//   dbLog.Debug("Query executed successfully")
//   // Output: [2025-01-07 15:04:05.000] DEBUG [database] Query executed successfully
func WithComponent(component string) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithComponent(component)
	}
	// Return a basic logger if default is not initialized
	l, err := New(Config{Level: "info", ToStdout: true})
	if err != nil {
		// Fallback to a minimal logger
		return &Logger{
			logger:     log.New(os.Stdout, "", log.LstdFlags),
			level:      LevelInfo,
			component:  component,
			withCaller: false,
		}
	}
	return l.WithComponent(component)
}

// SetLevel changes the logging level of the default logger.
//
// This allows dynamic runtime adjustment of log verbosity for the entire
// application. If Init() has not been called, this function does nothing.
//
// Parameters:
//   - level: The new minimum log level
//
// Thread Safety:
//   Safe for concurrent use.
//
// Example:
//   logger.Init(logger.Config{Level: "info"})
//   // ... normal operation ...
//   logger.SetLevel(logger.LevelDebug)  // Enable verbose logging for debugging
//   // ... debug problematic code ...
//   logger.SetLevel(logger.LevelInfo)   // Return to normal verbosity
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}

// Close closes the default logger's file handle if one was opened.
//
// This should be called during application shutdown to ensure log files are
// properly flushed and closed. If Init() was not called or the logger doesn't
// have a file, this function does nothing.
//
// Returns:
//   - error: Any error from closing the file, or nil
//
// Thread Safety:
//   Safe for concurrent use, though logging after Close has undefined behavior.
//
// Example:
//   logger.Init(logger.Config{FilePath: "/var/log/app.log"})
//   defer logger.Close()
//   // ... application code ...
func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
