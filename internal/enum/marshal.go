package enum

import (
	"encoding/json/v2"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/LarsArtmann/art-dupl/errors"
)

// ValidatableEnum interface for enums that can validate themselves.
type ValidatableEnum interface {
	IsValid() bool
}

// StringEnum is a generic type for string-based enums.
type StringEnum string

// String returns the string value of the enum.
func (se StringEnum) String() string {
	return string(se)
}

// UnmarshalJSON provides generic JSON unmarshaling for string-based enum types.
// Usage: implement UnmarshalJSON on your enum type like this:
//
//	func (e *MyEnum) UnmarshalJSON(data []byte) error {
//		return enum.UnmarshalJSON(e, data, MyEnumType, MyEnumValue1)
//	}
func UnmarshalJSON[T StringEnum](dest *T, data []byte, typeName string, defaultValue T, validValues ...T) error {
	str := string(data)
	// Remove quotes if present
	str = strings.TrimSpace(str)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Check if empty
	if str == "" {
		*dest = defaultValue
		return nil
	}

	candidate := T(str)
	// Validate against provided values
	if slices.Contains(validValues, candidate) {
		*dest = candidate
		return nil
	}

	// Create error with rich context
	validationErr := fmt.Errorf("invalid %s value %q, must be one of %v", typeName, str, validValues)
	*dest = defaultValue
	return fmt.Errorf("%s unmarshaling failed for %s (dest=%v, defaultValue=%v, validValues=%v): %w", typeName, str, dest, defaultValue, validValues, errors.NewValidationError(validationErr.Error(), validationErr))
}

// MarshalJSON provides generic JSON marshaling for enum types.
// Usage: implement MarshalJSON on your enum type like this:
//
//	func (e MyEnum) MarshalJSON() ([]byte, error) {
//		return enum.MarshalJSON(e, MyEnumValue1, MyEnumValue2)
//	}
func MarshalJSON[T StringEnum](value T, validValues ...T) ([]byte, error) {
	// Validate that current value is in valid list
	if slices.Contains(validValues, value) {
		return json.Marshal(string(value))
	}

	// If we get here, the value is invalid
	validationErr := fmt.Errorf("enum value %q is not in valid list %v", value, validValues)
	return nil, fmt.Errorf("marshaling failed for value %v (validValues=%v): %w", value, validValues, errors.NewValidationError("failed to marshal enum: "+validationErr.Error(), validationErr))
}

// UnmarshalJSONFromStrings unmarshals JSON using a string slice for validation.
func UnmarshalJSONFromStrings[T StringEnum](dest *T, data []byte, typeName string, defaultValue T, validStrings []string) error {
	str := string(data)
	str = strings.TrimSpace(str)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	if str == "" {
		*dest = defaultValue
		return nil
	}

	// Check against valid strings
	if slices.Contains(validStrings, str) {
		*dest = T(str)
		return nil
	}

	validationErr := fmt.Errorf("invalid %s value %q, must be one of %v", typeName, str, validStrings)
	*dest = defaultValue
	return fmt.Errorf("%s unmarshaling failed for %s (dest=%v, defaultValue=%v, validStrings=%v): %w", typeName, str, dest, defaultValue, validStrings, errors.NewValidationError(validationErr.Error(), validationErr))
}

// ParseEnum parses a string into an enum type with validation.
func ParseEnum[T StringEnum](str string, defaultValue T, validValues ...T) T {
	str = strings.TrimSpace(str)
	if str == "" {
		return defaultValue
	}

	candidate := T(str)
	if slices.Contains(validValues, candidate) {
		return candidate
	}
	return defaultValue
}

// EnumToStringSlice converts enum values to string slice.
func EnumToStringSlice[T StringEnum](enums ...T) []string {
	result := make([]string, 0, len(enums))
	for _, e := range enums {
		result = append(result, string(e))
	}
	return result
}

// EnumType defines common interface for all enum-like types with IsValid method.
type EnumType interface {
	IsValid() bool
	String() string
}

// UnmarshalJSONForInterface unmarshals JSON using the EnumType interface.
// This is useful for enums that implement their own IsValid method.
func UnmarshalJSONForInterface[T EnumType](dest *T, data []byte, typeName string) error {
	str := string(data)
	str = strings.TrimSpace(str)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Get the zero value of type T
	var zero T
	typeOfT := reflect.TypeOf(zero)

	// Create candidate by converting string to type T
	candidatePtr := reflect.New(typeOfT.Elem())
	candidatePtr.Elem().SetString(str)
	candidate := candidatePtr.Elem().Interface().(T)

	// Validate using IsValid method
	if !candidate.IsValid() {
		validationErr := fmt.Errorf("invalid %s value %q (must pass validation)", typeName, str)
		*dest = zero
		return errors.NewValidationError(validationErr.Error(), validationErr)
	}

	*dest = candidate
	return nil
}

// MarshalJSONForInterface marshals JSON using the EnumType interface.
// This is useful for enums that implement their own IsValid method.
func MarshalJSONForInterface[T EnumType](value T, typeName string) ([]byte, error) {
	// Validate using IsValid method
	if !value.IsValid() {
		// Return null for invalid values to allow omitempty to work
		// This handles zero values that should be omitted from JSON output
		return []byte("null"), nil
	}

	return json.Marshal(value.String())
}

// ValidateEnum checks if an enum value is in the list of valid values.
func ValidateEnum[T StringEnum](value T, validValues ...T) bool {
	return slices.Contains(validValues, value)
}

// EnumNames returns string names of enum values.
func EnumNames[T StringEnum](enums ...T) []string {
	names := make([]string, 0, len(enums))
	for _, e := range enums {
		names = append(names, string(e))
	}
	return names
}
