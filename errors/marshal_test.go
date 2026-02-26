package errors

import (
	"errors"
	"testing"
)

func TestMarshalError(t *testing.T) {
	cause := errors.New("json syntax error")
	err := &MarshalError{
		Operation: "marshal",
		Context:   "test data",
		Cause:     cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message should not be empty")
	}

	if !errors.Is(err, cause) {
		t.Error("Should unwrap to cause")
	}
}

func TestHandleMarshalingError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := HandleMarshalingError("marshal", "test", nil)
		if result != nil {
			t.Error("Should return nil for nil error")
		}
	})

	t.Run("handles unsupported value", func(t *testing.T) {
		err := errors.New("json: unsupported value")

		result := HandleMarshalingError("marshal", "test", err)
		if result == nil {
			t.Fatal("Should return error")
		}

		var marshalErr *MarshalError
		if !errors.As(result, &marshalErr) {
			t.Error("Should return MarshalError")
		}
	})

	tests := []struct {
		name string
		err  string
	}{
		{"unsupported type", "json: unsupported type"},
		{"invalid UTF-8", "json: invalid UTF-8"},
		{"generic error", "generic json error"},
	}
	for _, tt := range tests {
		t.Run("handles "+tt.name, func(t *testing.T) {
			err := errors.New(tt.err)

			result := HandleMarshalingError("marshal", "test", err)
			if result == nil {
				t.Fatal("Should return error")
			}
		})
	}
}

func TestSafeMarshal(t *testing.T) {
	t.Run("successful marshal", func(t *testing.T) {
		data := map[string]string{"key": "value"}

		result, err := SafeMarshal(data, "test data")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}

		if len(result) == 0 {
			t.Error("Should return data")
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		// Channel cannot be marshaled
		data := make(chan int)

		_, err := SafeMarshal(data, "test data")
		if err == nil {
			t.Error("Should error for unmarshalable data")
		}
	})
}

func TestSafeMarshalNilSafe(t *testing.T) {
	t.Run("successful marshal", func(t *testing.T) {
		data := map[string]string{"key": "value"}

		result, err := SafeMarshalNilSafe(data, "nil error")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}

		if len(result) == 0 {
			t.Error("Should return data")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := SafeMarshalNilSafe(nil, "input is nil")
		if err == nil {
			t.Error("Should error for nil input")
		}

		if !Is(err, ValidationError) {
			t.Error("Should be ValidationError")
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		data := make(chan int)

		_, err := SafeMarshalNilSafe(data, "nil error")
		if err == nil {
			t.Error("Should error for unmarshalable data")
		}
	})
}

func TestSafeMarshalIndent(t *testing.T) {
	t.Run("successful indent marshal", func(t *testing.T) {
		data := map[string]string{"key": "value"}

		result, err := SafeMarshalIndent(data, "", "  ", "test data")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}

		if len(result) == 0 {
			t.Error("Should return data")
		}
		// Check it's indented
		if string(result)[0] != '{' {
			t.Error("Should start with brace")
		}
	})

	t.Run("indent marshal error", func(t *testing.T) {
		data := make(chan int)

		_, err := SafeMarshalIndent(data, "", "  ", "test data")
		if err == nil {
			t.Error("Should error for unmarshalable data")
		}
	})
}

func TestSafeMarshalIndentNilSafe(t *testing.T) {
	t.Run("successful indent marshal", func(t *testing.T) {
		data := map[string]string{"key": "value"}

		result, err := SafeMarshalIndentNilSafe(data, "", "  ", "nil error")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}

		if len(result) == 0 {
			t.Error("Should return data")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := SafeMarshalIndentNilSafe(nil, "", "  ", "input is nil")
		if err == nil {
			t.Error("Should error for nil input")
		}
	})

	t.Run("indent marshal error", func(t *testing.T) {
		data := make(chan int)

		_, err := SafeMarshalIndentNilSafe(data, "", "  ", "nil error")
		if err == nil {
			t.Error("Should error for unmarshalable data")
		}
	})
}

func TestSafeUnmarshal(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		data := []byte(`{"key":"value"}`)

		var result map[string]string

		err := SafeUnmarshal(data, &result, "test data")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}

		if result["key"] != "value" {
			t.Error("Should unmarshal correctly")
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		data := []byte(`{invalid json`)

		var result map[string]string

		err := SafeUnmarshal(data, &result, "test data")
		if err == nil {
			t.Error("Should error for invalid JSON")
		}

		var marshalErr *MarshalError
		if !errors.As(err, &marshalErr) {
			t.Error("Should return MarshalError")
		}
	})
}
