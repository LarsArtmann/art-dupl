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
