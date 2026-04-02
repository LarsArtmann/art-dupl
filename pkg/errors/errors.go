// Package errors provides custom error types and error handling utilities.
//
// This package defines application-specific error types that can be used
// throughout the codebase for consistent error handling and reporting.
package errors

// CoreError is the foundation for all custom errors in this project.
type CoreError struct {
	Message string
	Code    string
}

// Error implements the error interface.
func (e *CoreError) Error() string {
	return e.Message
}

// New creates a new CoreError with the given message and code.
func New(message, code string) *CoreError {
	return &CoreError{
		Message: message,
		Code:    code,
	}
}
