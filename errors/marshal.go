package errors

// SafeMarshal provides safe marshaling with consistent error handling.
// This is a generic JSON marshaling utility that handles common error cases
// with descriptive error messages. For type-specific marshaling functions,
// consider using the typed functions in their respective packages.

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Pre-defined error formats for static error wrapping.
var (
	ErrUnsupportedValueType = errors.New("unsupported value type")
	ErrUnsupportedType      = errors.New("unsupported type")
	ErrInvalidUTF8          = errors.New("invalid UTF-8 encoding")
)

// MarshalError is a specialized error for JSON marshaling failures.
type MarshalError struct {
	Operation string
	Context   string
	Cause     error
}

func (e *MarshalError) Error() string {
	return fmt.Sprintf("JSON %s error for %s: %v", e.Operation, e.Context, e.Cause)
}

func (e *MarshalError) Unwrap() error {
	return e.Cause
}

// HandleMarshalingError provides unified JSON marshaling error handling.
// It inspects the concrete encoding/json error type (not its message string,
// which is unstable) so callers can match the typed sentinels via errors.Is.
func HandleMarshalingError(operation, context string, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.As(err, new(*json.UnsupportedValueError)):
		return newMarshalError(operation, context, err, ErrUnsupportedValueType)
	case errors.As(err, new(*json.UnsupportedTypeError)):
		return newMarshalError(operation, context, err, ErrUnsupportedType)
	default:
		return &MarshalError{Operation: operation, Context: context, Cause: err}
	}
}

// newMarshalError creates a new MarshalError with a wrapped static error.
func newMarshalError(operation, context string, err, staticErr error) *MarshalError {
	return &MarshalError{
		Operation: operation,
		Context:   context,
		Cause:     fmt.Errorf("%w: %w", staticErr, err),
	}
}

// SafeMarshal provides safe marshaling with consistent error handling.
func SafeMarshal(v any, context string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, HandleMarshalingError("marshal", context, err)
	}

	return data, nil
}

// SafeMarshalNilSafe provides safe marshaling with nil check.
// Returns a validation error if the input is nil.
func SafeMarshalNilSafe(v any, nilErrorMessage string) ([]byte, error) {
	if v == nil {
		return nil, NewValidationError(nilErrorMessage, nil)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, HandleMarshalingError("marshal", nilErrorMessage, err)
	}

	return data, nil
}

// SafeMarshalIndent provides safe indented marshaling with consistent error handling.
func SafeMarshalIndent(v any, prefix, indent, context string) ([]byte, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, fmt.Errorf("marshal indent (prefix: %q, context: %s): %w",
			prefix, context, HandleMarshalingError("marshal indent", context, err))
	}

	return data, nil
}

// SafeMarshalIndentNilSafe provides safe indented marshaling with nil check.
// Returns a validation error if the input is nil.
func SafeMarshalIndentNilSafe(v any, prefix, indent, nilErrorMessage string) ([]byte, error) {
	if v == nil {
		return nil, NewValidationError(nilErrorMessage, nil)
	}

	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, HandleMarshalingError(
			"marshal indent",
			fmt.Sprintf("%s (prefix: %q)", nilErrorMessage, prefix),
			err,
		)
	}

	return data, nil
}

// SafeUnmarshal provides safe unmarshaling with consistent error handling.
func SafeUnmarshal(data []byte, v any, context string) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return HandleMarshalingError("unmarshal", context, err)
	}

	return nil
}
