package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger

	// LogFilePath is where logs are written
	LogFilePath string
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Init initializes the logging system
func Init(level LogLevel, verbose bool) error {
	// Create logger
	Logger = log.New(os.Stderr)

	// Set level
	switch level {
	case LevelDebug:
		Logger.SetLevel(log.DebugLevel)
	case LevelInfo:
		Logger.SetLevel(log.InfoLevel)
	case LevelWarn:
		Logger.SetLevel(log.WarnLevel)
	case LevelError:
		Logger.SetLevel(log.ErrorLevel)
	default:
		Logger.SetLevel(log.InfoLevel)
	}

	// If verbose, lower to debug
	if verbose {
		Logger.SetLevel(log.DebugLevel)
	}

	// Setup log file
	if err := setupLogFile(); err != nil {
		return fmt.Errorf("failed to setup log file: %w", err)
	}

	return nil
}

// setupLogFile creates the log file and configures file logging
func setupLogFile() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create .merlin directory if it doesn't exist
	merlinDir := filepath.Join(homeDir, ".merlin")
	if err := os.MkdirAll(merlinDir, 0755); err != nil {
		return err
	}

	// Create log file path
	LogFilePath = filepath.Join(merlinDir, "merlin.log")

	// Open log file (append mode)
	logFile, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Create a separate logger for file output
	fileLogger := log.NewWithOptions(logFile, log.Options{
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		ReportCaller:    false,
	})

	// Set same level as main logger
	fileLogger.SetLevel(Logger.GetLevel())

	// Store reference (we'll use this for file-only logging)
	Logger.SetOutput(os.Stderr) // Keep stderr output

	return nil
}

// Debug logs a debug message
func Debug(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Debug(msg, keyvals...)
	}
}

// Info logs an info message
func Info(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Info(msg, keyvals...)
	}
}

// Warn logs a warning message
func Warn(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Warn(msg, keyvals...)
	}
}

// Error logs an error message
func Error(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Error(msg, keyvals...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Fatal(msg, keyvals...)
	} else {
		fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
		os.Exit(1)
	}
}

// GetLogFilePath returns the path to the log file
func GetLogFilePath() string {
	return LogFilePath
}
