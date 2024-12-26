package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel defines log levels
type LogLevel int

const (
	// Log levels
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// LogFormat defines the output format
type LogFormat int

const (
	// Log formats
	TEXT LogFormat = iota
	JSON
)

// LogOutput defines where logs should be written
type LogOutput struct {
	Writer    io.Writer
	Formatter LogFormatter
}

// LogFormatter interface for custom formatters
type LogFormatter interface {
	Format(entry *LogEntry) ([]byte, error)
}

// TextFormatter formats logs as text
type TextFormatter struct {
	TimeFormat string
}

// JSONFormatter formats logs as JSON
type JSONFormatter struct {
	TimeFormat string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp  time.Time
	Level      LogLevel
	Message    string
	Fields     map[string]interface{}
	Caller     string
	StackTrace string
}

// Logger configurable logger
type Logger struct {
	mu        sync.Mutex
	level     LogLevel
	outputs   []LogOutput
	format    LogFormat
	context   map[string]interface{}
	callDepth int
}

// Config logger configuration
type Config struct {
	Level     LogLevel
	Outputs   []LogOutput
	Format    LogFormat
	Context   map[string]interface{}
	CallDepth int
}

// Format formats the log entry as text
func (f *TextFormatter) Format(entry *LogEntry) ([]byte, error) {
	timeStr := entry.Timestamp.Format(f.TimeFormat)
	levelStr := getLevelString(entry.Level)

	var fieldsStr string
	for k, v := range entry.Fields {
		fieldsStr += fmt.Sprintf(" %s=%v", k, v)
	}

	var callerInfo string
	if entry.Caller != "" {
		callerInfo = fmt.Sprintf(" [%s]", entry.Caller)
	}

	logLine := fmt.Sprintf("%s [%s]%s %s%s\n",
		timeStr, levelStr, callerInfo, entry.Message, fieldsStr)

	if entry.StackTrace != "" {
		logLine += entry.StackTrace + "\n"
	}

	return []byte(logLine), nil
}

// Format formats the log entry as JSON
func (f *JSONFormatter) Format(entry *LogEntry) ([]byte, error) {
	data := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(f.TimeFormat),
		"level":     getLevelString(entry.Level),
		"message":   entry.Message,
	}

	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}

	if len(entry.Fields) > 0 {
		data["fields"] = entry.Fields
	}

	if entry.StackTrace != "" {
		data["stack_trace"] = entry.StackTrace
	}

	return json.Marshal(data)
}

// NewLogger creates a new logger
func NewLogger(config Config) *Logger {
	if len(config.Outputs) == 0 {
		config.Outputs = []LogOutput{{
			Writer:    os.Stdout,
			Formatter: &TextFormatter{TimeFormat: "2006-01-02 15:04:05"},
		}}
	}

	if config.CallDepth == 0 {
		config.CallDepth = 2
	}

	return &Logger{
		level:     config.Level,
		outputs:   config.Outputs,
		format:    config.Format,
		context:   config.Context,
		callDepth: config.CallDepth,
	}
}

// getCaller returns the caller information
func (l *Logger) getCaller() string {
	if pc, file, line, ok := runtime.Caller(l.callDepth); ok {
		return fmt.Sprintf("%s:%d %s",
			filepath.Base(file), line, filepath.Base(runtime.FuncForPC(pc).Name()))
	}
	return ""
}

// getStack returns the stack trace
func getStack(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var trace string
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("\n    %s:%d - %s",
			filepath.Base(frame.File), frame.Line, filepath.Base(frame.Function))
		if !more {
			break
		}
	}
	return trace
}

// log creates a log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, includeStack bool) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Combine context and fields
	allFields := make(map[string]interface{})
	for k, v := range l.context {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    allFields,
		Caller:    l.getCaller(),
	}

	if includeStack {
		entry.StackTrace = getStack(l.callDepth)
	}

	// Write to all outputs
	for _, output := range l.outputs {
		if formatted, err := output.Formatter.Format(entry); err == nil {
			output.Writer.Write(formatted)
		}
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs at debug level
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log(DEBUG, message, fields, false)
}

// Info logs at info level
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(INFO, message, fields, false)
}

// Warn logs at warn level
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(WARN, message, fields, false)
}

// Error logs at error level with stack trace
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log(ERROR, message, fields, true)
}

// Fatal logs at fatal level with stack trace and exits
func (l *Logger) Fatal(message string, fields map[string]interface{}) {
	l.log(FATAL, message, fields, true)
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(context map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:     l.level,
		outputs:   l.outputs,
		format:    l.format,
		callDepth: l.callDepth,
		context:   make(map[string]interface{}),
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

// AddOutput adds a new output
func (l *Logger) AddOutput(output LogOutput) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.outputs = append(l.outputs, output)
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetCallDepth changes the call depth for caller information
func (l *Logger) SetCallDepth(depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callDepth = depth
}

// getLevelString returns the string representation of a log level
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

// Example usage in comments:
/*
	// Create a file output
	file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	// Configure logger with multiple outputs
	logger := NewLogger(Config{
		Level: INFO,
		Outputs: []LogOutput{
			{
				Writer:    os.Stdout,
				Formatter: &TextFormatter{TimeFormat: "2006-01-02 15:04:05"},
			},
			{
				Writer:    file,
				Formatter: &JSONFormatter{TimeFormat: time.RFC3339},
			},
		},
		Context: map[string]interface{}{
			"app": "myapp",
			"env": "production",
		},
	})

	// Log with additional fields
	logger.Info("Server started", map[string]interface{}{
		"port": 8080,
		"mode": "production",
	})

	// Create a logger with additional context
	dbLogger := logger.WithContext(map[string]interface{}{
		"component": "database",
	})

	// Log error with stack trace
	dbLogger.Error("Connection failed", map[string]interface{}{
		"error": err.Error(),
		"host":  "localhost",
	})
*/
