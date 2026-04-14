package testutil

import (
	"fmt"
	"strings"
	"testing"
)

// AssertLen asserts that the length of a slice matches the expected value.
// If the assertion fails, it reports the actual length.
func AssertLen[T any](t *testing.T, got []T, want int, what string) {
	t.Helper()

	if len(got) != want {
		t.Errorf("%s: expected length %d, got %d", what, want, len(got))
	}
}

// AssertCount asserts that the count of items matches the expected value.
// The what parameter describes what is being counted (e.g., "Clones", "results").
func AssertCount(t *testing.T, got int, want int, what string) {
	t.Helper()

	if got != want {
		t.Errorf("%s: expected %d, got %d", what, want, got)
	}
}

// AssertCountf asserts that the count of items matches the expected value with a formatted message.
func AssertCountf(t *testing.T, got int, want int, format string, args ...any) {
	t.Helper()

	if got != want {
		t.Errorf(format, append(args, got, want)...)
	}
}

// AssertNotNil asserts that a value is not nil.
func AssertNotNil(t *testing.T, got any, what string) {
	t.Helper()

	if got == nil {
		t.Errorf("%s: expected non-nil, got nil", what)
	}
}

// AssertNil asserts that a value is nil.
func AssertNil(t *testing.T, got any, what string) {
	t.Helper()

	if got != nil {
		t.Errorf("%s: expected nil, got %v", what, got)
	}
}

// AssertEqualFn asserts that got equals want using a custom comparison function.
// The compareFn should return true if values are equal.
func AssertEqualFn[T any](t *testing.T, got, want T, compareFn func(a, b T) bool, what string) {
	t.Helper()

	if !compareFn(got, want) {
		t.Errorf("%s: expected %v, got %v", what, want, got)
	}
}

// ExpectTrue asserts that the condition is true.
func ExpectTrue(t *testing.T, cond bool, what string) {
	t.Helper()

	if !cond {
		t.Errorf("%s: expected true, got false", what)
	}
}

// ExpectFalse asserts that the condition is false.
func ExpectFalse(t *testing.T, cond bool, what string) {
	t.Helper()

	if cond {
		t.Errorf("%s: expected false, got true", what)
	}
}

// AssertError asserts that an error occurred.
func AssertError(t *testing.T, err error, what string) {
	t.Helper()

	if err == nil {
		t.Errorf("%s: expected error, got nil", what)
	}
}

// AssertNoError asserts that no error occurred.
func AssertNoError(t *testing.T, err error, what string) {
	t.Helper()

	if err != nil {
		t.Errorf("%s: unexpected error: %v", what, err)
	}
}

// AssertStringContains asserts that a string contains a substring.
func AssertStringContains(t *testing.T, got, substr, what string) {
	t.Helper()

	if !strings.Contains(got, substr) {
		t.Errorf("%s: expected to contain %q, got %q", what, substr, got)
	}
}

// AssertStringPrefix asserts that a string starts with a prefix.
func AssertStringPrefix(t *testing.T, got, prefix, what string) {
	t.Helper()

	if !strings.HasPrefix(got, prefix) {
		t.Errorf("%s: expected prefix %q, got %q", what, prefix, got)
	}
}

// AssertIntInRange asserts that got is within min and max (inclusive).
func AssertIntInRange(t *testing.T, got, min, max int, what string) {
	t.Helper()

	if got < min || got > max {
		t.Errorf("%s: expected %d to be in range [%d, %d]", what, got, min, max)
	}
}

// FormatGotWant formats got and want for error messages.
func FormatGotWant(got, want any) string {
	return fmt.Sprintf("got %v, want %v", got, want)
}
