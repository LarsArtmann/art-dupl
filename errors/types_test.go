package errors

import (
	"errors"
	"testing"
)

// assertEqual is a generic equality assertion helper.
func assertEqual[T comparable](t *testing.T, got, expected T, name string) {
	t.Helper()

	if got != expected {
		t.Errorf("%s: expected %v, got %v", name, expected, got)
	}
}

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
	assertEqual(t, err.Type, ParseError, "Type")
	assertEqual(t, err.File, "test.go", "File")
	assertEqual(t, err.Line, 42, "Line")
	assertEqual(t, err.Message, "test message", "Message")
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
		{"DetectionError", func() *DuplError {
			return NewDetectionError("", nil)
		}, DetectionError},
		{"AnalysisError", func() *DuplError {
			return NewAnalysisError("", nil)
		}, AnalysisError},
		{"FileError", func() *DuplError {
			return NewFileError("", "", nil)
		}, FileError},
		{"TimeoutError", func() *DuplError {
			return NewTimeoutError("", nil)
		}, TimeoutError},
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

func TestWrap(t *testing.T) {
	cause := errors.New("cause error")

	t.Run("Wrap creates new error", func(t *testing.T) {
		wrapped := Wrap(cause, InternalError, "wrapping message")
		if wrapped == nil {
			t.Fatal("Wrapped error should not be nil")
		}

		if !errors.Is(wrapped, cause) {
			t.Error("Wrapped error should contain cause")
		}

		if !Is(wrapped, InternalError) {
			t.Error("Wrapped error should be InternalError type")
		}
	})

	t.Run("Wrap nil returns nil", func(t *testing.T) {
		wrapped := Wrap(nil, InternalError, "message")
		if wrapped != nil {
			t.Error("Wrapping nil should return nil")
		}
	})

	t.Run("Wrap does not double-wrap DuplError", func(t *testing.T) {
		duplErr := NewInternalError("original", nil)

		wrapped := Wrap(duplErr, ConfigError, "wrapping")
		if !errors.Is(wrapped, duplErr) {
			t.Error("Should not re-wrap DuplError")
		}
	})
}

func TestWrapf(t *testing.T) {
	cause := errors.New("cause error")

	t.Run("Wrapf with format", func(t *testing.T) {
		wrapped := Wrapf(cause, InternalError, "failed to process %d items", 42)
		if wrapped == nil {
			t.Fatal("Wrapped error should not be nil")
		}

		if !errors.Is(wrapped, cause) {
			t.Error("Wrapped error should contain cause")
		}
	})
}

func TestWrapIO(t *testing.T) {
	cause := errors.New("read error")

	t.Run("WrapIO creates IOError", func(t *testing.T) {
		wrapped := WrapIO(cause, "test.go", "reading file")
		if !Is(wrapped, IOError) {
			t.Error("Should be IOError type")
		}
	})

	t.Run("WrapIO does not double-wrap", func(t *testing.T) {
		ioErr := NewIOError("test.go", "original", nil)

		wrapped := WrapIO(ioErr, "test.go", "new operation")
		if !errors.Is(wrapped, ioErr) {
			t.Error("Should not re-wrap IOError")
		}
	})
}

func TestWrapFunctions(t *testing.T) {
	tests := []struct {
		name       string
		wrappedErr error
		expected   ErrorType
	}{
		{
			name:       "WrapConfig creates ConfigError",
			wrappedErr: WrapConfig(errors.New("parse error"), "loading config"),
			expected:   ConfigError,
		},
		{
			name:       "WrapValidation creates ValidationError",
			wrappedErr: WrapValidation(errors.New("invalid value"), "field validation"),
			expected:   ValidationError,
		},
		{
			name:       "WrapFile creates FileError",
			wrappedErr: WrapFile(errors.New("not found"), "test.go", "stat"),
			expected:   FileError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !Is(tc.wrappedErr, tc.expected) {
				t.Errorf("Expected %s type", tc.expected)
			}
		})
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ParseError, "parse"},
		{ConfigError, "config"},
		{IOError, "io"},
		{ValidationError, "validation"},
		{InternalError, "internal"},
		{DetectionError, "detection"},
		{AnalysisError, "analysis"},
		{FileError, "file"},
		{TimeoutError, "timeout"},
	}

	for _, tc := range tests {
		t.Run(string(tc.errorType), func(t *testing.T) {
			if tc.errorType.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.errorType.String())
			}
		})
	}
}
