package config

import "fmt"

// UnmarshalStringToEnum is a generic helper for unmarshaling JSON strings to typed enums
func UnmarshalStringToEnum[T ~string](data []byte, enumType func(string) T, isValid func(T) bool, errorMsg string) (T, error) {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	candidate := enumType(str)
	if !isValid(candidate) {
		var zero T
		return zero, fmt.Errorf(errorMsg, str)
	}
	return candidate, nil
}

// UnmarshalEnumJSON is a generic helper for implementing UnmarshalJSON for enum types
func UnmarshalEnumJSON[T ~string](data []byte, enumType func(string) T, isValid func(T) bool, typeName string) (T, error) {
	return UnmarshalStringToEnum(data, enumType, isValid, fmt.Sprintf("invalid %s: %%s", typeName))
}

// MarshalEnumJSON is a generic helper for implementing MarshalJSON for enum types
func MarshalEnumJSON[T ~string](value T, isValid func(T) bool, typeName string) ([]byte, error) {
	if !isValid(value) {
		return nil, fmt.Errorf("invalid %s: %s", typeName, value)
	}
	return fmt.Appendf(nil, `"%s"`, value), nil
}

// EnumType defines the common interface for all enum-like types
type EnumType[T ~string] interface {
	~string
	String() string
	IsValid() bool
}

// NewEnumUnmarshalJSON creates an UnmarshalJSON function for any enum type
func NewEnumUnmarshalJSON[T ~string](typeName string) func(*T, []byte) error {
	return func(e *T, data []byte) error {
		candidate, err := UnmarshalEnumJSON(data, func(s string) T { return T(s) }, func(t T) bool {
			// Use the interface method to check validity
			var enum T = t
			return any(enum).(interface{ IsValid() bool }).IsValid()
		}, typeName)
		if err != nil {
			return err
		}
		*e = candidate
		return nil
	}
}

// UnmarshalJSONForEnum is a generic function that can be used as UnmarshalJSON method
func UnmarshalJSONForEnum[T ~string](e *T, data []byte, typeName string) error {
	candidate, err := UnmarshalEnumJSON(data, func(s string) T { return T(s) }, func(t T) bool {
		// Use the interface method to check validity
		var enum T = t
		return any(enum).(interface{ IsValid() bool }).IsValid()
	}, typeName)
	if err != nil {
		return err
	}
	*e = candidate
	return nil
}
