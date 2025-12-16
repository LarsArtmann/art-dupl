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
