package domain_test

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
)

// ExampleDetectionState demonstrates state constants.
func ExampleDetectionState() {
	states := []domain.DetectionState{
		domain.DetectionStateIdle,
		domain.DetectionStateRunning,
		domain.DetectionStateCompleted,
		domain.DetectionStateFailed,
	}

	for _, state := range states {
		fmt.Printf("State: %s (valid: %v)\n", state, state.IsValid())
	}

	// Output:
	// State: idle (valid: true)
	// State: running (valid: true)
	// State: completed (valid: true)
	// State: failed (valid: true)
}

// ExampleAnalysisMode demonstrates analysis modes.
func ExampleAnalysisMode() {
	modes := []domain.AnalysisMode{
		domain.AnalysisModeQuick,
		domain.AnalysisModeDeep,
		domain.AnalysisModeFull,
	}

	for _, mode := range modes {
		fmt.Printf("Mode: %s (valid: %v)\n", mode, mode.IsValid())
	}

	// Output:
	// Mode: quick (valid: true)
	// Mode: deep (valid: true)
	// Mode: full (valid: true)
}

// ExampleLineNumber demonstrates line number validation.
func ExampleLineNumber() {
	// Valid line number
	ln, err := domain.NewLineNumber(42)
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		return
	}

	fmt.Printf("Line number: %d\n", ln)

	// Invalid: zero line number
	_, err = domain.NewLineNumber(0)
	if err != nil {
		fmt.Printf("Zero line error: %v\n", err)
	}

	// Output:
	// Line number: 42
	// Zero line error: validation error: line number cannot be 0
}

// ExampleBytePosition demonstrates byte position operations.
func ExampleBytePosition() {
	pos := domain.NewBytePosition(1024)
	fmt.Printf("Byte position: %d\n", pos)

	// Output: Byte position: 1024
}

// ExampleFileProcessingState demonstrates file processing states.
func ExampleFileProcessingState() {
	states := []domain.FileProcessingState{
		domain.FileProcessingStatePending,
		domain.FileProcessingStateProcessing,
		domain.FileProcessingStateCompleted,
		domain.FileProcessingStateFailed,
	}

	for _, state := range states {
		fmt.Printf("State: %s (valid: %v)\n", state, state.IsValid())
	}

	// Output:
	// State: pending (valid: true)
	// State: processing (valid: true)
	// State: completed (valid: true)
	// State: failed (valid: true)
}
