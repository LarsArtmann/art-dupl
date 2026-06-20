package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

func marshalStringID(s, validationMsg string) ([]byte, error) {
	if s == "" {
		return nil, errors.New(validationMsg)
	}

	return json.Marshal(s) //nolint:wrapcheck // Standard JSON marshaling
}

func unmarshalWithValidation[T any](
	data []byte,
	typeName, validationMsg string,
	validator func(T) bool,
	assign func(T),
) error {
	var value T

	err := json.Unmarshal(data, &value)
	if err != nil {
		return fmt.Errorf(
			"unmarshal %s failed (validationMsg=%q, value=%v): %w",
			typeName,
			validationMsg,
			value,
			err,
		)
	}

	if !validator(value) {
		return fmt.Errorf("%s (typeName=%q, value=%v)", validationMsg, typeName, value)
	}

	assign(value)

	return nil
}

func unmarshalStringID(data []byte, typeName, validationMsg string, assign func(string)) error {
	return unmarshalWithValidation(
		data,
		typeName,
		validationMsg,
		func(s string) bool { return s != "" },
		assign,
	)
}
