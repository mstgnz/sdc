package err

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType defines error types
type ErrorType string

const (
	// Error types with detailed descriptions
	ErrTypeConnection    ErrorType = "ConnectionError"    // Database connection related errors
	ErrTypeQuery         ErrorType = "QueryError"         // SQL query execution errors
	ErrTypeTransaction   ErrorType = "TransactionError"   // Transaction related errors
	ErrTypeMigration     ErrorType = "MigrationError"     // Database migration errors
	ErrTypeSchema        ErrorType = "SchemaError"        // Database schema related errors
	ErrTypeValidation    ErrorType = "ValidationError"    // Data validation errors
	ErrTypeConversion    ErrorType = "ConversionError"    // Data type conversion errors
	ErrTypeConfiguration ErrorType = "ConfigurationError" // Configuration related errors
	ErrTypeParser        ErrorType = "ParserError"        // SQL parsing errors
	ErrTypeIO            ErrorType = "IOError"            // Input/Output related errors
)

// ErrorSeverity defines the severity level of errors
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// DatabaseError represents enhanced database errors
type DatabaseError struct {
	Type     ErrorType
	Message  string
	Err      error
	Severity ErrorSeverity
	Stack    string
	Context  map[string]interface{}
}

// Error returns the formatted error message
func (e *DatabaseError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Type, e.Message))

	if e.Err != nil {
		sb.WriteString(fmt.Sprintf(" - %v", e.Err))
	}

	if len(e.Context) > 0 {
		sb.WriteString("\nContext:")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("\n  %s: %v", k, v))
		}
	}

	return sb.String()
}

// Unwrap returns the underlying error
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// WithContext adds context information to the error
func (e *DatabaseError) WithContext(key string, value interface{}) *DatabaseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSeverity sets the error severity
func (e *DatabaseError) WithSeverity(severity ErrorSeverity) *DatabaseError {
	e.Severity = severity
	return e
}

// captureStack captures the current stack trace
func captureStack(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	sb.WriteString("\nStack Trace:")

	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			sb.WriteString(fmt.Sprintf("\n  %s:%d - %s", frame.File, frame.Line, frame.Function))
		}
		if !more {
			break
		}
	}

	return sb.String()
}

// NewDatabaseError creates a new enhanced database error
func NewDatabaseError(errType ErrorType, message string, err error) *DatabaseError {
	return &DatabaseError{
		Type:     errType,
		Message:  message,
		Err:      err,
		Severity: SeverityMedium,
		Stack:    captureStack(2),
		Context:  make(map[string]interface{}),
	}
}

// Error type checking functions
func IsErrorType(err error, errType ErrorType) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*DatabaseError); ok {
		return e.Type == errType
	}
	return false
}

// Convenience functions for error type checking
func IsConnectionError(err error) bool    { return IsErrorType(err, ErrTypeConnection) }
func IsQueryError(err error) bool         { return IsErrorType(err, ErrTypeQuery) }
func IsTransactionError(err error) bool   { return IsErrorType(err, ErrTypeTransaction) }
func IsMigrationError(err error) bool     { return IsErrorType(err, ErrTypeMigration) }
func IsSchemaError(err error) bool        { return IsErrorType(err, ErrTypeSchema) }
func IsValidationError(err error) bool    { return IsErrorType(err, ErrTypeValidation) }
func IsConversionError(err error) bool    { return IsErrorType(err, ErrTypeConversion) }
func IsConfigurationError(err error) bool { return IsErrorType(err, ErrTypeConfiguration) }
func IsParserError(err error) bool        { return IsErrorType(err, ErrTypeParser) }
func IsIOError(err error) bool            { return IsErrorType(err, ErrTypeIO) }

// IsCriticalError checks if the error is critical
func IsCriticalError(err error) bool {
	if e, ok := err.(*DatabaseError); ok {
		return e.Severity == SeverityCritical
	}
	return false
}

// Example usage in comments:
/*
	// Creating a new error with context and severity
	err := NewDatabaseError(ErrTypeQuery, "Failed to execute query", originalErr).
		WithContext("query", "SELECT * FROM users").
		WithContext("params", []string{"id", "name"}).
		WithSeverity(SeverityHigh)

	// Checking error type and handling
	if IsQueryError(err) {
		// Handle query error
	}

	// Checking severity
	if IsCriticalError(err) {
		// Handle critical error
	}
*/
