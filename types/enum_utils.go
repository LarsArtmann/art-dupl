package types

import (
	"fmt"
)

// ValidatableEnum interface for enums that can validate themselves.
type ValidatableEnum interface {
	IsValid() bool
}

// UnmarshalEnumJSON provides generic unmarshaling for enum types.
func UnmarshalEnumJSON[T ValidatableEnum](data []byte, constructor func(string) T, typeName string) (*T, error) {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	enum := constructor(str)
	if !enum.IsValid() {
		return nil, fmt.Errorf("enum validation failed for %s: invalid value %q (data: %q)", typeName, str, string(data))
	}
	return &enum, nil
}

// MarshalEnumJSON provides generic marshaling for enum types.
func MarshalEnumJSON[T ValidatableEnum](enum T, typeName string) ([]byte, error) {
	if !enum.IsValid() {
		return nil, fmt.Errorf("failed to marshal %s: invalid value %v (value must pass validation)", typeName, enum)
	}
	return []byte(`"` + fmt.Sprintf("%v", enum) + `"`), nil
}

// StringEnum is a generic type for string-based enums.
type StringEnum string

func (se StringEnum) String() string {
	return string(se)
}
