package domain

import (
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/art-dupl/errors"
)

// marshalStringID is a helper function for marshaling string-based ID types.
// It handles the common pattern of validating non-empty strings and marshaling to JSON.
func marshalStringID(s, validationMsg string) ([]byte, error) {
	if s == "" {
		return nil, errors.NewValidationError(validationMsg, nil)
	}

	return json.Marshal(s) //nolint:wrapcheck // Standard JSON marshaling
}

// unmarshalWithValidation is a generic helper for unmarshaling JSON with custom validation.
// It unmarshals data to type T, validates it using the provided validator function,
// and assigns the result if validation passes.
func unmarshalWithValidation[T any](
	data []byte,
	typeName, validationMsg string,
	validator func(T) bool,
	assign func(T),
) error {
	var value T

	err := json.Unmarshal(data, &value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}

	if !validator(value) {
		return errors.NewValidationError(validationMsg, nil)
	}

	assign(value)

	return nil
}

// unmarshalStringID is a helper for unmarshaling string IDs that must not be empty.
func unmarshalStringID(data []byte, typeName, validationMsg string, assign func(string)) error {
	return unmarshalWithValidation(
		data,
		typeName,
		validationMsg,
		func(s string) bool { return s != "" },
		assign,
	)
}

// unmarshalUintNonZero is a helper for unmarshaling uint-based types that must not be zero.
func unmarshalUintNonZero(data []byte, typeName, validationMsg string, assign func(uint)) error {
	return unmarshalWithValidation(
		data,
		typeName,
		validationMsg,
		func(n uint) bool { return n != 0 },
		assign,
	)
}

// unmarshalUint is a helper function for unmarshaling uint-based types
// that allow zero values. It handles the common pattern of unmarshaling JSON to uint.
func unmarshalUint(data []byte, typeName string, assign func(uint)) error {
	return unmarshalWithValidation(data, typeName, "", func(uint) bool { return true }, assign)
}

// marshalUint is a helper function for marshaling uint-based types.
// It handles the common pattern of marshaling uint-wrapped types to JSON.
func marshalUint(n uint) ([]byte, error) {
	return json.Marshal(n) //nolint:wrapcheck // Standard JSON marshaling
}

// unmarshalUintGeneric is a generic helper function for unmarshaling unsigned integer types.
// It handles the common pattern of unmarshaling JSON to unsigned integers.
func unmarshalUintGeneric[T uint16 | uint32](data []byte, typeName string, assign func(T)) error {
	var n T

	err := json.Unmarshal(data, &n)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}

	assign(n)

	return nil
}

// marshalInt32 is a helper function for marshaling int32-based types.
// It handles the common pattern of marshaling int32-wrapped types to JSON.
func marshalInt32(n int32) ([]byte, error) {
	return json.Marshal(n) //nolint:wrapcheck // Standard JSON marshaling
}

// unmarshalInt32 is a helper function for unmarshaling int32-based types.
// It handles the common pattern of unmarshaling JSON to int32.
func unmarshalInt32(data []byte, typeName string, assign func(int32)) error {
	var n int32

	err := json.Unmarshal(data, &n)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", typeName, err)
	}

	assign(n)

	return nil
}
