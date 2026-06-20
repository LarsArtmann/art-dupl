package errors

import (
	"errors"
	"testing"
)

// assertFieldsEqual is a generic equality assertion helper for struct fields.
func assertFieldsEqual[T comparable](t *testing.T, got, expected T, name string) {
	t.Helper()

	if got != expected {
		t.Errorf("%s: expected %v, got %v", name, expected, got)
	}
}

func TestDuplError(t *testing.T) {
	cause := errors.New("cause error")

	unwrapErr := NewIOError("test.go", "test message", cause)

	expected := "io error at test.go:0: test message"
	if unwrapErr.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, unwrapErr.Error())
	}

	assertUnwrapsToCause(t, unwrapErr, cause, "Unwrap should return cause error")

	assertFieldsEqual(t, unwrapErr.Type, IOError, "Type")
	assertFieldsEqual(t, unwrapErr.File, "test.go", "File")
	assertFieldsEqual(t, unwrapErr.Message, "test message", "Message")
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func() *DuplError
		expected ErrorType
	}{
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
		{"FileError", func() *DuplError {
			return NewFileError("", "", nil)
		}, FileError},
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
	ioErr := NewIOError("", "", nil)
	configErr := NewConfigError("", nil)

	if !Is(ioErr, IOError) {
		t.Error("Should identify IOError")
	}

	if !Is(configErr, ConfigError) {
		t.Error("Should identify ConfigError")
	}

	if Is(ioErr, ConfigError) {
		t.Error("Should not match ConfigError")
	}

	if Is(configErr, IOError) {
		t.Error("Should not match IOError")
	}

	standardErr := errors.New("standard error")
	if Is(standardErr, IOError) {
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

		assertUnwrapsToCause(t, wrapped, cause, "Wrapped error should contain cause")

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

		assertUnwrapsToCause(t, wrapped, cause, "Wrapped error should contain cause")
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
				t.Errorf("Expected %s, got %v", tc.expected, tc.wrappedErr)
			}
		})
	}
}
