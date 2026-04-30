package domain

import (
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/art-dupl/errors"
)

func marshalStringID(s, validationMsg string) ([]byte, error) {
	if s == "" {
		return nil, errors.NewValidationError(validationMsg, nil)
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
		return errors.NewValidationError(
			fmt.Sprintf("%s (typeName=%q, value=%v)", validationMsg, typeName, value),
			nil,
		)
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
