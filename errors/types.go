// Package errors provides type-safe error handling for dupl
package errors

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// ErrorType categorizes different types of errors
type ErrorType string

const (
	ParseError      ErrorType = "parse"
	ConfigError     ErrorType = "config"
	IOError         ErrorType = "io"
	ValidationError ErrorType = "validation"
	InternalError   ErrorType = "internal"
)

// DuplError is the main error type with rich context
type DuplError struct {
	Type    ErrorType
	Message string
	File    string
	Line    int
	Cause   error
	Stack   string
}

// Error implements the error interface
func (e *DuplError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s error at %s:%d: %s", e.Type, e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Wrapper interface
func (e *DuplError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new parse error with context
func NewParseError(file string, line int, msg string, cause error) *DuplError {
	return &DuplError{
		Type:    ParseError,
		Message: msg,
		File:    file,
		Line:    line,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    ConfigError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewIOError creates a new I/O error
func NewIOError(file, msg string, cause error) *DuplError {
	return &DuplError{
		Type:    IOError,
		Message: msg,
		File:    file,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    ValidationError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewInternalError creates a new internal error
func NewInternalError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    InternalError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// Is checks if error matches a specific type
func Is(err error, errorType ErrorType) bool {
	var duplErr *DuplError
	if errors.As(err, &duplErr) {
		return duplErr.Type == errorType
	}
	return false
}
