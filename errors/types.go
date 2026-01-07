// Package errors provides type-safe error handling for dupl
package errors

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// ErrorType categorizes different types of errors.
type ErrorType string

const (
	ParseError      ErrorType = "parse"
	ConfigError     ErrorType = "config"
	IOError         ErrorType = "io"
	ValidationError ErrorType = "validation"
	InternalError   ErrorType = "internal"
)

// DuplError is the main error type with rich context.
type DuplError struct {
	Type    ErrorType
	Message string
	File    string
	Line    int
	Cause   error
	Stack   string
}

// EnumValidationError provides domain-specific error context for enum validation failures.
type EnumValidationError struct {
	DuplError

	EnumValue string
	EnumType  string
}

// NewEnumValidationError creates a new enum validation error with rich context.
func NewEnumValidationError(enumType, enumValue string, cause error) *EnumValidationError {
	return &EnumValidationError{
		DuplError: DuplError{
			Type:    ValidationError,
			Message: fmt.Sprintf("enum validation error: type=%s, value=%q", enumType, enumValue),
			Cause:   cause,
			Stack:   string(debug.Stack()),
		},
		EnumValue: enumValue,
		EnumType:  enumType,
	}
}

// Error implements the error interface for EnumValidationError.
func (e *EnumValidationError) Error() string {
	return fmt.Sprintf("enum validation failed: type=%s, value=%q (file: %s:%d)",
		e.EnumType, e.EnumValue, e.File, e.Line)
}

// NewParseError creates a new parse error with context.
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

// NewConfigError creates a new configuration error.
func NewConfigError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    ConfigError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewIOError creates a new I/O error.
func NewIOError(file, msg string, cause error) *DuplError {
	return &DuplError{
		Type:    IOError,
		Message: msg,
		File:    file,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    ValidationError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(msg string, cause error) *DuplError {
	return &DuplError{
		Type:    InternalError,
		Message: msg,
		Cause:   cause,
		Stack:   string(debug.Stack()),
	}
}

// Error implements the error interface.
func (e *DuplError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s error at %s:%d: %s", e.Type, e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Wrapper interface.
func (e *DuplError) Unwrap() error {
	return e.Cause
}

// Is checks if error matches a specific type.
func Is(err error, errorType ErrorType) bool {
	var duplErr *DuplError
	if errors.As(err, &duplErr) {
		return duplErr.Type == errorType
	}
	return false
}
