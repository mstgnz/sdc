package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel defines log levels
type LogLevel int

const (
	// Log seviyeleri
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger configurable logger
type Logger struct {
	mu      sync.Mutex
	level   LogLevel
	output  io.Writer
	logger  *log.Logger
	prefix  string
	context map[string]interface{}
}

// Config logger configuration
type Config struct {
	Level   LogLevel
	Output  io.Writer
	Prefix  string
	Context map[string]interface{}
}

// NewLogger creates a new logger
func NewLogger(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	return &Logger{
		level:   config.Level,
		output:  config.Output,
		logger:  log.New(config.Output, config.Prefix, log.LstdFlags),
		prefix:  config.Prefix,
		context: config.Context,
	}
}

// formatMessage formats the log message
func (l *Logger) formatMessage(level LogLevel, message string, fields map[string]interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := getLevelString(level)

	// Combine context and fields
	allFields := make(map[string]interface{})
	for k, v := range l.context {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Convert Fields to string
	fieldsStr := ""
	for k, v := range allFields {
		fieldsStr += fmt.Sprintf(" %s=%v", k, v)
	}

	return fmt.Sprintf("%s [%s] %s%s %s", timestamp, levelStr, l.prefix, message, fieldsStr)
}

// getLevelString returns the log level as a string
func getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Log creates a log record at the specified level
func (l *Logger) Log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	formattedMessage := l.formatMessage(level, message, fields)
	l.logger.Println(formattedMessage)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug creates a log record at debug level
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.Log(DEBUG, message, fields)
}

// Info creates log record at info info level
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.Log(INFO, message, fields)
}

// Warn creates a log record at warn level
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.Log(WARN, message, fields)
}

// Error creates a log record at error level
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.Log(ERROR, message, fields)
}

// Fatal creates a log record at fatal level and exits the application
func (l *Logger) Fatal(message string, fields map[string]interface{}) {
	l.Log(FATAL, message, fields)
}

// WithContext adds new context fields
func (l *Logger) WithContext(context map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:   l.level,
		output:  l.output,
		logger:  l.logger,
		prefix:  l.prefix,
		context: make(map[string]interface{}),
	}

	// Copy current context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context
	for k, v := range context {
		newLogger.context[k] = v
	}

	return newLogger
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput changes the log output
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	l.logger.SetOutput(output)
}
