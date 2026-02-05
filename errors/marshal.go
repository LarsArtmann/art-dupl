package errors

// SafeMarshal provides safe marshaling with consistent error handling.
// This is a generic JSON marshaling utility that handles common error cases
// with descriptive error messages. For type-specific marshaling functions,
// consider using the typed functions in their respective packages.

import (
	"encoding/json"
	"fmt"
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
func HandleMarshalingError(operation, context string, err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "json: unsupported value":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("unsupported value type: %w", err)}
	case "json: unsupported type":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("unsupported type: %w", err)}
	case "json: invalid UTF-8":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("invalid UTF-8 encoding: %w", err)}
	default:
		return &MarshalError{Operation: operation, Context: context, Cause: err}
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
		return nil, NewConfigError("failed to marshal value", err)
	}
	return data, nil
}

// SafeMarshalIndent provides safe indented marshaling with consistent error handling.
func SafeMarshalIndent(v any, prefix, indent, context string) ([]byte, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, HandleMarshalingError("marshal indent", context, err)
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
		return nil, NewConfigError("failed to marshal value with indent", err)
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
