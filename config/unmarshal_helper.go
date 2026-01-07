package config

import (
	"fmt"
	"reflect"

	"github.com/LarsArtmann/art-dupl/errors"
)

// UnmarshalStringToEnum is a generic helper for unmarshaling JSON strings to typed enums.
func UnmarshalStringToEnum[T ~string](data []byte, enumType func(string) T, isValid func(T) bool, errorMsg string) (T, error) { //nolint:ireturn // Generic type T is appropriate for utility function
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	candidate := enumType(str)
	if !isValid(candidate) {
		var zero T
		// Use errors package for rich error context with stack trace
		validationErr := fmt.Errorf("value %q did not pass validation for type %s (original data: %q, parsed string: %q, candidate: %q)", str, reflect.TypeOf(zero).Name(), data, str, candidate)
		return zero, errors.NewValidationError(validationErr.Error(), validationErr)
	}
	return candidate, nil
}

// UnmarshalEnumJSON is a generic helper for implementing UnmarshalJSON for enum types.
func UnmarshalEnumJSON[T ~string](data []byte, enumType func(string) T, isValid func(T) bool, typeName string) (T, error) { //nolint:ireturn // Generic type T is appropriate for utility function
	return UnmarshalStringToEnum(data, enumType, isValid, "invalid %s: %"+typeName)
}

// MarshalEnumJSON is a generic helper for implementing MarshalJSON for enum types.
func MarshalEnumJSON[T ~string](value T, isValid func(T) bool, typeName string) ([]byte, error) {
	if !isValid(value) {
		// Use errors package for rich error context with stack trace
		validationErr := fmt.Errorf("value=%q, isValid=%v, typeName=%s", value, isValid(value), typeName)
		return nil, errors.NewValidationError("failed to marshal enum: "+validationErr.Error(), validationErr)
	}
	return fmt.Appendf(nil, `"%s"`, value), nil
}

// EnumType defines the common interface for all enum-like types.
type EnumType[T ~string] interface {
	~string
	String() string
	IsValid() bool
}

// NewEnumUnmarshalJSON creates an UnmarshalJSON function for any enum type.
func NewEnumUnmarshalJSON[T ~string](typeName string) func(*T, []byte) error {
	return func(e *T, data []byte) error {
		candidate, err := UnmarshalEnumJSON(data, func(s string) T { return T(s) }, func(t T) bool {
			// Use the interface method to check validity
			enum := t
			return any(enum).(interface{ IsValid() bool }).IsValid()
		}, typeName)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %s from data %q: %w", typeName, string(data), err)
		}
		*e = candidate
		return nil
	}
}

// UnmarshalJSONForEnum is a generic function that can be used as UnmarshalJSON method.
func UnmarshalJSONForEnum[T ~string](e *T, data []byte, typeName string) error {
	candidate, err := UnmarshalEnumJSON(data, func(s string) T { return T(s) }, func(t T) bool {
		// Use the interface method to check validity
		enum := t
		return any(enum).(interface{ IsValid() bool }).IsValid()
	}, typeName)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s from data %q: %w", typeName, string(data), err)
	}
	*e = candidate
	return nil
}
