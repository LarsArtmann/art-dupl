package testutil

import "testing"

// AssertLen asserts that the length of a slice matches the expected value.
// If the assertion fails, it reports the actual length.
func AssertLen[T any](t *testing.T, got []T, want int, what string) {
	t.Helper()
	if len(got) != want {
		t.Errorf("%s: expected length %d, got %d", what, want, len(got))
	}
}

// AssertNotNil asserts that a value is not nil.
func AssertNotNil(t *testing.T, got interface{}, what string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected non-nil, got nil", what)
	}
}
