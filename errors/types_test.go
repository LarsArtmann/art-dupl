package errors

import (
	"errors"
	"testing"
)

func TestDuplError(t *testing.T) {
	cause := errors.New("cause error")

	// Test error creation
	err := NewParseError("test.go", 42, "test message", cause)

	// Test error message
	expected := "parse error at test.go:42: test message"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test unwrap
	if !errors.Is(err, cause) {
		t.Error("Unwrap should return cause error")
	}

	// Test fields
	if err.Type != ParseError {
		t.Errorf("Expected ParseError type, got %s", err.Type)
	}
	if err.File != "test.go" {
		t.Errorf("Expected 'test.go', got '%s'", err.File)
	}
	if err.Line != 42 {
		t.Errorf("Expected 42, got %d", err.Line)
	}
	if err.Message != "test message" {
		t.Errorf("Expected 'test message', got '%s'", err.Message)
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func() *DuplError
		expected ErrorType
	}{
		{"ParseError", func() *DuplError {
			return NewParseError("", 0, "", nil)
		}, ParseError},
		{"ConfigError", func() *DuplError {
			return NewConfigError("", nil)
		}, ConfigError},
		{"IOError", func() *DuplError {
			return NewIOError("", "", nil)
		}, IOError},
		{"ValidationError", func() *DuplError {
			return NewValidationError("", nil)
		}, ValidationError},
		{"InternalError", func() *DuplError {
			return NewInternalError("", nil)
		}, InternalError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.errFunc()
			if err.Type != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, err.Type)
			}
		})
	}
}

func TestIs(t *testing.T) {
	parseErr := NewParseError("", 0, "", nil)
	configErr := NewConfigError("", nil)

	// Test positive cases
	if !Is(parseErr, ParseError) {
		t.Error("Should identify ParseError")
	}
	if !Is(configErr, ConfigError) {
		t.Error("Should identify ConfigError")
	}

	// Test negative cases
	if Is(parseErr, ConfigError) {
		t.Error("Should not match ConfigError")
	}
	if Is(configErr, ParseError) {
		t.Error("Should not match ParseError")
	}

	// Test with standard error
	standardErr := errors.New("standard error")
	if Is(standardErr, ParseError) {
		t.Error("Should not match standard error")
	}
}
