package errors

import (
	"errors"
	"testing"
)

func TestNewEnumValidationError(t *testing.T) {
	// Test that EnumValidationError is created correctly
	enumType := "DetectionMethod"
	enumValue := "invalid_method"
	cause := errors.New("invalid detection method")

	err := NewEnumValidationError(enumType, enumValue, cause)

	if err == nil {
		t.Fatal("NewEnumValidationError should return error, got nil")
	}

	// Check DuplError fields
	if err.Type != ValidationError {
		t.Errorf("Expected Type=ValidationError, got %s", err.Type)
	}
	if err.Cause != cause {
		t.Errorf("Expected Cause to match input cause")
	}
	if err.Stack == "" {
		t.Errorf("Expected non-empty Stack field")
	}

	// Check EnumValidationError specific fields
	if err.EnumType != enumType {
		t.Errorf("Expected EnumType=%s, got %s", enumType, err.EnumType)
	}
	if err.EnumValue != enumValue {
		t.Errorf("Expected EnumValue=%s, got %s", enumValue, err.EnumValue)
	}
}

func TestEnumValidationError_Error(t *testing.T) {
	// Test that Error() method returns formatted message
	enumType := "OutputFormat"
	enumValue := "invalid_format"
	cause := errors.New("invalid format")

	err := NewEnumValidationError(enumType, enumValue, cause)

	msg := err.Error()

	// Check that message contains expected fields
	if len(msg) == 0 {
		t.Fatal("Error() should return non-empty message")
	}

	// Check for enum type in message
	expectedType := "type=" + enumType
	if !containsString(msg, expectedType) {
		t.Errorf("Expected message to contain %q, got %q", expectedType, msg)
	}
}

func TestEnumValidationError_Unwrap(t *testing.T) {
	// Test that EnumValidationError unwraps to cause
	enumType := "DetectionMethod"
	enumValue := "invalid_method"
	cause := errors.New("original error")

	err := NewEnumValidationError(enumType, enumValue, cause)

	// Unwrap should return original cause
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Expected Unwrap to return cause, got %v", unwrapped)
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findString(s, substr) >= 0
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
