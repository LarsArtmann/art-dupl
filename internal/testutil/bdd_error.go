package testutil

import "fmt"

// BDDError represents an error that occurred during BDD test setup or execution.
type BDDError struct {
	Operation string
	Cause     error
	Output    string
}

// NewBDDError creates a new BDDError with the given operation and cause.
func NewBDDError(operation string, cause error) *BDDError {
	return &BDDError{ //nolint:exhaustruct
		Operation: operation,
		Cause:     cause,
	}
}

// WithOutput adds output to the BDDError.
func (e *BDDError) WithOutput(output string) *BDDError {
	e.Output = output

	return e
}

// Error implements the error interface.
func (e *BDDError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("BDD error during %s: %v\nOutput: %s", e.Operation, e.Cause, e.Output)
	}

	return fmt.Sprintf("BDD error during %s: %v", e.Operation, e.Cause)
}

// Unwrap returns the underlying cause.
func (e *BDDError) Unwrap() error {
	return e.Cause
}
