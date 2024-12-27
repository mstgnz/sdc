package monitoring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// LogFormat represents the output format of log messages
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// LogConfig holds configuration for the logger
type LogConfig struct {
	Level      LogLevel
	Format     LogFormat
	OutputPath string
	ErrorPath  string
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// Logger handles logging operations
type Logger struct {
	config LogConfig
	output io.Writer
	error  io.Writer
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config LogConfig) (*Logger, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(config.OutputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Configure main log output
	output := &lumberjack.Logger{
		Filename:   config.OutputPath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Configure error log output
	errorOutput := &lumberjack.Logger{
		Filename:   config.ErrorPath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	return &Logger{
		config: config,
		output: output,
		error:  errorOutput,
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if l.config.Level <= DebugLevel {
		l.log(DebugLevel, msg, fields)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if l.config.Level <= InfoLevel {
		l.log(InfoLevel, msg, fields)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if l.config.Level <= WarnLevel {
		l.log(WarnLevel, msg, fields)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	if l.config.Level <= ErrorLevel {
		l.log(ErrorLevel, msg, fields)
	}
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)

	var output io.Writer
	if level == ErrorLevel {
		output = l.error
	} else {
		output = l.output
	}

	if l.config.Format == JSONFormat {
		l.logJSON(output, level, timestamp, msg, fields)
	} else {
		l.logText(output, level, timestamp, msg, fields)
	}
}

// logJSON writes a log message in JSON format
func (l *Logger) logJSON(w io.Writer, level LogLevel, timestamp, msg string, fields map[string]interface{}) {
	// Add basic fields
	logEntry := map[string]interface{}{
		"timestamp": timestamp,
		"level":     level.String(),
		"message":   msg,
	}

	// Add custom fields
	for k, v := range fields {
		logEntry[k] = v
	}

	// Write JSON to output
	fmt.Fprintf(w, "%v\n", logEntry)
}

// logText writes a log message in text format
func (l *Logger) logText(w io.Writer, level LogLevel, timestamp, msg string, fields map[string]interface{}) {
	// Write basic log entry
	fmt.Fprintf(w, "%s [%s] %s", timestamp, level.String(), msg)

	// Add fields if present
	if len(fields) > 0 {
		fmt.Fprintf(w, " fields=%v", fields)
	}

	fmt.Fprintln(w)
}

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}
