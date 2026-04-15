package testhelpers

import "testing"

// PanicRecovery returns a defer recover function that reports a panic with the given input.
// Use in fuzz tests to catch panics and report them with the input that caused the panic.
func PanicRecovery(t *testing.T, input string) func() {
	t.Helper()

	return func() {
		if r := recover(); r != nil {
			t.Errorf("Panicked with input %q: %v", input, r)
		}
	}
}
