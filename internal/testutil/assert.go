package testutil

import (
	"encoding/json"
	"errors"
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
func AssertCount(t *testing.T, got, want int, what string) {
	t.Helper()

	if got != want {
		t.Errorf("%s: expected %d, got %d", what, want, got)
	}
}

// AssertCountf asserts that the count of items matches the expected value with a formatted message.
func AssertCountf(t *testing.T, got, want int, format string, args ...any) {
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

// AssertErrorIs asserts that err wraps the expected error.
func AssertErrorIs(t *testing.T, err, want error, what string) {
	t.Helper()

	if !errors.Is(err, want) {
		t.Errorf("%s: error %v does not wrap %v", what, err, want)
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
func AssertIntInRange(t *testing.T, got, minVal, maxVal int, what string) {
	t.Helper()

	if got < minVal || got > maxVal {
		t.Errorf("%s: expected %d to be in range [%d, %d]", what, got, minVal, maxVal)
	}
}

// FormatGotWant formats got and want for error messages.
func FormatGotWant(got, want any) string {
	return fmt.Sprintf("got %v, want %v", got, want)
}

// AssertIsValid asserts that the IsValid() method returns the expected error state.
// The typeName parameter is used in the error message (e.g., "Analysis", "Clone").
func AssertIsValid[T any](
	t *testing.T,
	typeName string,
	obj T,
	wantErr bool,
	getErr func(T) error,
) {
	t.Helper()

	err := getErr(obj)
	if (err != nil) != wantErr {
		t.Errorf("%s.IsValid() error = %v, wantErr %v", typeName, err, wantErr)
	}
}

// AssertMarshalJSONError checks that a JSON marshal operation returns the expected error state.
// The methodName parameter is used in the error message (e.g., "MarshalJSON").
func AssertMarshalJSONError[T any](
	t *testing.T,
	methodName string,
	obj T,
	wantError bool,
	marshalFn func(T) ([]byte, error),
) {
	t.Helper()

	got, err := marshalFn(obj)
	if (err != nil) != wantError {
		t.Errorf("%s() error = %v, wantError %v", methodName, err, wantError)

		return
	}

	if !wantError && len(got) == 0 {
		t.Errorf("%s() returned empty bytes", methodName)
	}
}

// AssertUnmarshalJSONError checks that a JSON unmarshal operation returns the expected error state.
// The methodName parameter is used in the error message (e.g., "UnmarshalJSON").
func AssertUnmarshalJSONError[T any](
	t *testing.T,
	methodName string,
	data []byte,
	wantError bool,
	unmarshalFn func([]byte) error,
) {
	t.Helper()

	err := unmarshalFn(data)
	if (err != nil) != wantError {
		t.Errorf("%s() error = %v, wantErr %v", methodName, err, wantError)
	}
}

// AssertErrorMatches checks that an error matches expected state.
// The wantErr parameter indicates whether an error is expected.
// The msg parameter is used in the error message (e.g., "handleWalkEntry()").
func AssertErrorMatches(t *testing.T, err error, wantErr bool, msg string) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("%s error = %v, wantErr %v", msg, err, wantErr)
	}
}

// AssertJSONRoundTrip verifies that a value can be marshaled to JSON and unmarshaled back.
// The obj parameter is the original object to roundtrip.
// The t.Helper() should be called before this function.
// Returns the unmarshaled result for further assertions.
func AssertJSONRoundTrip[T any](t *testing.T, obj T) T {
	t.Helper()

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result T

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	return result
}

// AssertUnmarshalError verifies that unmarshaling from JSON produces the expected error state.
// The methodName is used in error messages (e.g., "StringID.UnmarshalJSON").
// The data parameter is the JSON bytes to unmarshal.
func AssertUnmarshalError(
	t *testing.T,
	methodName string,
	data []byte,
	wantErr bool,
	unmarshalFn func([]byte) error,
) {
	t.Helper()

	err := unmarshalFn(data)
	if (err != nil) != wantErr {
		t.Errorf("%s() error = %v, wantErr %v", methodName, err, wantErr)
	}
}

// AssertConfigField asserts that a config field matches the expected value.
// The fieldName is used in the error message (e.g., "Threshold").
// The actual parameter should be the field value, and expected is the expected value.
func AssertConfigField[T comparable](t *testing.T, fieldName string, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("Expected %s %v, got %v", fieldName, expected, actual)
	}
}

// AssertConfigFieldFunc asserts that a config field matches the expected value using a custom message.
// The msg parameter is prepended to the error message.
// The actual parameter should be the field value, and expected is the expected value.
func AssertConfigFieldFunc[T comparable](t *testing.T, msg string, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertExpectedGot asserts that got matches the expected value with "Expected X, got Y" format.
// This helper eliminates AST clone patterns from inline assertions like:
//
//	if got != want { t.Errorf("Expected X, got Y", X, got) }
func AssertExpectedGot[T any](t *testing.T, got T, wantDescription string, wantValue T) {
	t.Helper()

	if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", wantValue) {
		t.Errorf("Expected %s, got %v", wantDescription, got)
	}
}

// AssertFieldValue asserts that a field matches the expected value.
// The fieldName is used in the error message as a description.
// This helper eliminates AST clone patterns from inline field assertions.
func AssertFieldValue[T comparable](t *testing.T, got, want T, fieldName string) {
	t.Helper()

	if got != want {
		t.Errorf("Expected %s %v, got %v", fieldName, want, got)
	}
}
