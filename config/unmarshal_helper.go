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

// EnumUnmarshaler is a generic type that provides UnmarshalJSON for enum types
type EnumUnmarshaler[T ~string] struct {
	enumType func(string) T
	isValid  func(T) bool
	typeName string
}

// NewEnumUnmarshaler creates a new EnumUnmarshaler for the given enum type
func NewEnumUnmarshaler[T ~string](enumType func(string) T, isValid func(T) bool, typeName string) *EnumUnmarshaler[T] {
	return &EnumUnmarshaler[T]{
		enumType: enumType,
		isValid:  isValid,
		typeName: typeName,
	}
}

// UnmarshalJSON implements json.Unmarshaler for the enum type
func (e *EnumUnmarshaler[T]) UnmarshalJSON(data []byte, ptr *T) error {
	candidate, err := UnmarshalEnumJSON(data, e.enumType, e.isValid, e.typeName)
	if err != nil {
		return err
	}
	*ptr = candidate
	return nil
}
