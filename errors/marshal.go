package errors

//
// TODO: TYPE SAFETY ENHANCEMENT - Consider adding typed marshaling functions
// that work with specific types instead of `any`. This would provide:
// - Compile-time type safety
// - Better IDE autocomplete
// - Reduced reflection overhead
//
// Example:
//   func SafeMarshalConfig(cfg *config.Config) ([]byte, error)
//   func SafeMarshalClone(c *domain.Clone) ([]byte, error)
//
// Also: Consider using generics for type-safe marshaling:
//   func SafeMarshalTyped[T any](v T, context string) ([]byte, error)

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

// SafeMarshalIndent provides safe indented marshaling with consistent error handling.
func SafeMarshalIndent(v any, prefix, indent, context string) ([]byte, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, HandleMarshalingError("marshal indent", context, err)
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
