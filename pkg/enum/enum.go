// Package enum provides generic helpers for string-based enum types,
// eliminating boilerplate IsValid/MarshalJSON/UnmarshalJSON/Parse methods.
//
// Usage:
//
//	type Color string
//
//	const (
//	    Red   Color = "red"
//	    Green Color = "green"
//	    Blue  Color = "blue"
//	)
//
//	func (c Color) IsValid() bool {
//	    switch c {
//	    case Red, Green, Blue:
//	        return true
//	    default:
//	        return false
//	    }
//	}
//
//	func (c Color) MarshalJSON() ([]byte, error) {
//	    return enum.MarshalJSON(c, c.IsValid, ErrInvalidColor)
//	}
package enum

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON marshals a string-based enum value to a JSON string.
// Returns the provided error if the value is invalid.
func MarshalJSON[T ~string](val T, isValid func(T) bool, invalidErr error) ([]byte, error) {
	if !isValid(val) {
		return nil, fmt.Errorf("%w: %q", invalidErr, val)
	}

	data, err := json.Marshal(string(val))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enum value: %w", err)
	}

	return data, nil
}

// UnmarshalJSON unmarshals JSON data into a string-based enum value.
// Returns the provided error if the parsed value is invalid.
func UnmarshalJSON[T ~string](
	data []byte,
	isValid func(T) bool,
	invalidErr error,
) (T, error) {
	var str string

	err := json.Unmarshal(data, &str)
	if err != nil {
		var zero T

		return zero, fmt.Errorf("%w: %q", invalidErr, strings.TrimSpace(string(data)))
	}

	return validate(T(str), str, isValid, invalidErr)
}

// UnmarshalJSONInto is the pointer-receiver UnmarshalJSON companion to UnmarshalJSON.
// It handles the boilerplate that every string-based enum type repeats:
//
//	func (t *T) UnmarshalJSON(data []byte) error {
//	    parsed, err := enum.UnmarshalJSON(data, T.IsValid, ErrInvalidT)
//	    if err != nil {
//	        return err
//	    }
//	    *t = parsed
//	    return nil
//	}
//
// The wrapper avoids that ceremony at every enum declaration site.
func UnmarshalJSONInto[T ~string](
	dst *T,
	data []byte,
	isValid func(T) bool,
	invalidErr error,
) error {
	parsed, err := UnmarshalJSON(data, isValid, invalidErr)
	if err != nil {
		return err
	}

	*dst = parsed

	return nil
}

// Parse converts a raw string to an enum value with validation.
// Returns the provided error if the string is not a valid enum value.
func Parse[T ~string](s string, isValid func(T) bool, invalidErr error) (T, error) {
	return validate(T(s), s, isValid, invalidErr)
}

// validate returns val when it satisfies isValid, otherwise the zero value of T
// and invalidErr wrapped around displayValue. Shared by UnmarshalJSON and Parse
// so the "check + wrap" form lives in one place.
func validate[T ~string](val T, displayValue string, isValid func(T) bool, invalidErr error) (T, error) {
	if isValid(val) {
		return val, nil
	}

	var zero T

	return zero, fmt.Errorf("%w: %q", invalidErr, displayValue)
}
