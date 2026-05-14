package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Static errors for config validation.
var (
	ErrInvalidDetectionMethod = errors.New("invalid detection method")
	ErrInvalidType            = errors.New("invalid type")
	ErrInvalidThreshold       = errors.New("threshold must be >= 1")
	ErrThresholdTooLarge      = errors.New("threshold too large (max 1000)")
)

// isValidStringType validates a string type against a set of valid values.
func isValidStringType[T ~string](val T, validValues map[T]bool) bool {
	return validValues[val]
}

// isValidable is a constraint for types with an IsValid() method.
type isValidable interface {
	IsValid() bool
}

// isValidMethod returns a function that calls IsValid() on a value.
// This eliminates duplicate anonymous functions like func(s T) bool { return s.IsValid() }.
func isValidMethod[T isValidable]() func(T) bool {
	return func(v T) bool {
		return v.IsValid()
	}
}

// unmarshalStringType unmarshals JSON into a string type with validation.
func unmarshalStringType[T ~string](
	data []byte,
	isValid func(T) bool,
	defaultVal T,
	typeName string,
) (T, error) {
	var str string

	err := json.Unmarshal(data, &str)
	if err != nil {
		return defaultVal, fmt.Errorf(
			"unmarshaling %s from data %q (defaultValue=%s): %w",
			typeName,
			string(data),
			defaultVal,
			err,
		)
	}

	typed := T(str)
	if !isValid(typed) {
		return defaultVal, fmt.Errorf(
			"%w %s (value=%q, defaultVal=%v): %s",
			ErrInvalidType,
			typeName,
			str,
			defaultVal,
			str,
		)
	}

	return typed, nil
}

// unmarshalStringTypeToPointer unmarshals JSON into a pointer to a string type with validation.
func unmarshalStringTypeToPointer[T ~string](
	data []byte,
	isValid func(T) bool,
	defaultVal T,
	typeName string,
	target *T,
) error {
	val, err := unmarshalStringType(data, isValid, defaultVal, typeName)
	if err != nil {
		return fmt.Errorf(
			"unmarshal to %s failed (target=%v, defaultVal=%v): %w",
			typeName,
			target,
			defaultVal,
			err,
		)
	}

	*target = val

	return nil
}

// marshalStringType marshals a string type to JSON with validation.
// For invalid values (typically zero values with omitempty), returns null to allow omission.
func marshalStringType[T ~string](val T, isValid func(T) bool, typeName string) ([]byte, error) {
	if !isValid(val) {
		return []byte("null"), nil
	}

	data, err := json.Marshal(string(val))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s (val=%v): %w", typeName, val, err)
	}

	return data, nil
}
