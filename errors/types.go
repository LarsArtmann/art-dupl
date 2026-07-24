// Package errors provides type-safe error handling for dupl
package errors

import (
	"errors"
	"fmt"
)

// ErrorType categorizes different types of errors.
type ErrorType string

const (
	ConfigError     ErrorType = "config"
	IOError         ErrorType = "io"
	ValidationError ErrorType = "validation"
	InternalError   ErrorType = "internal"
	AnalysisError   ErrorType = "analysis"
	FileError       ErrorType = "file"
	CacheError      ErrorType = "cache"
)

// DuplError is the main error type with rich context.
type DuplError struct {
	Type    ErrorType
	Message string
	File    string
	Cause   error
}

// newError creates a new error with the specified type.
func newError(errorType ErrorType, msg string, cause error) *DuplError {
	return &DuplError{ // File/Line optional, set by specific constructors
		Type:    errorType,
		Message: msg,
		Cause:   cause,
	}
}

// NewConfigError creates a new configuration error.
func NewConfigError(msg string, cause error) *DuplError {
	return newError(ConfigError, msg, cause)
}

// NewIOError creates a new I/O error.
func NewIOError(file, msg string, cause error) *DuplError {
	return &DuplError{
		Type:    IOError,
		Message: msg,
		File:    file,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(msg string, cause error) *DuplError {
	return newError(ValidationError, msg, cause)
}

// NewInternalError creates a new internal error.
func NewInternalError(
	msg string,
	cause error,
) *DuplError {
	return newError(InternalError, msg, cause)
}

// NewFileError creates a new file error with context.
func NewFileError(file, msg string, cause error) *DuplError {
	return &DuplError{
		Type:    FileError,
		Message: msg,
		File:    file,
		Cause:   cause,
	}
}

// Error implements the error interface.
func (e *DuplError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s error at %s: %s", e.Type, e.File, e.Message)
	}

	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Wrapper interface.
func (e *DuplError) Unwrap() error {
	return e.Cause
}

// Is checks if error matches a specific type.
func Is(err error, errorType ErrorType) bool {
	duplErr, ok := errors.AsType[*DuplError](err)
	if ok {
		return duplErr.Type == errorType
	}

	return false
}

// IsDuplError checks if an error is already a DuplError.
func IsDuplError(err error) bool {
	_, ok := errors.AsType[*DuplError](err)

	return ok
}

// Wrap wraps an error with additional context using the specified error type.
// If the cause is already a DuplError, it returns the original error.
func Wrap(err error, errorType ErrorType, msg string) error {
	if err == nil {
		return nil
	}

	if IsDuplError(err) {
		return err
	}

	return &DuplError{ // File/Line optional for wrapped errors
		Type:    errorType,
		Message: msg,
		Cause:   err,
	}
}

// Wrapf wraps an error with formatted context using the specified error type.
func Wrapf(err error, errorType ErrorType, format string, args ...any) error {
	return Wrap(err, errorType, fmt.Sprintf(format, args...))
}

// WrapIO wraps an error as an IOError with file context.
func WrapIO(err error, file, operation string) error {
	if err == nil {
		return nil
	}

	// Don't wrap if already an IOError
	if Is(err, IOError) {
		return err
	}

	return NewIOError(file, operation, err)
}

// wrapWithMessage wraps an error with context using a specific error type and constructor.
// It appends the original error message to the context.
func wrapWithMessage(
	err error,
	errorType ErrorType,
	context string,
	constructor func(string, error) *DuplError,
) error {
	if err == nil {
		return nil
	}

	// Don't wrap if already the same error type
	if Is(err, errorType) {
		return err
	}

	return constructor(context+": "+err.Error(), err)
}

// WrapConfig wraps an error as a ConfigError.
func WrapConfig(err error, context string) error {
	return wrapWithMessage(err, ConfigError, context, NewConfigError)
}

// WrapValidation wraps an error as a ValidationError.
func WrapValidation(err error, context string) error {
	return wrapWithMessage(err, ValidationError, context, NewValidationError)
}

// WrapFile wraps an error as a FileError with operation context.
func WrapFile(err error, file, operation string) error {
	if err == nil {
		return nil
	}

	// Don't wrap if already a FileError
	if Is(err, FileError) {
		return err
	}

	return NewFileError(file, operation, err)
}
