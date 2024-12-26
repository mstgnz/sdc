package err

import (
	"fmt"
)

// ErrorType defines error types
type ErrorType string

const (
	// Error types
	ErrTypeConnection    ErrorType = "ConnectionError"
	ErrTypeQuery         ErrorType = "QueryError"
	ErrTypeTransaction   ErrorType = "TransactionError"
	ErrTypeMigration     ErrorType = "MigrationError"
	ErrTypeSchema        ErrorType = "SchemaError"
	ErrTypeValidation    ErrorType = "ValidationError"
	ErrTypeConversion    ErrorType = "ConversionError"
	ErrTypeConfiguration ErrorType = "ConfigurationError"
)

// DatabaseError represents database errors
type DatabaseError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error returns the error message
func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new database error
func NewDatabaseError(errType ErrorType, message string, err error) *DatabaseError {
	return &DatabaseError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// IsConnectionError checks if it is a connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*DatabaseError); ok {
		return e.Type == ErrTypeConnection
	}
	return false
}

// IsQueryError checks if it is a query error
func IsQueryError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*DatabaseError); ok {
		return e.Type == ErrTypeQuery
	}
	return false
}

// IsTransactionError checks if it is a transaction error
func IsTransactionError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*DatabaseError); ok {
		return e.Type == ErrTypeTransaction
	}
	return false
}

// IsMigrationError checks if it is a migration error
func IsMigrationError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*DatabaseError); ok {
		return e.Type == ErrTypeMigration
	}
	return false
}
