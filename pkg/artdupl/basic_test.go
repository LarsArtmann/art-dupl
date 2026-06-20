package artdupl

import (
	"errors"
	"fmt"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

const testFilename = "test.go"

// newTestConfig creates a detectorConfig for testing.
func newTestConfig() *detectorConfig {
	return &detectorConfig{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		Semantic:         true,
	}
}

// newTestProgress creates a Progress struct with the given parameters.
func newTestProgress(
	stage string,
	completed, total int,
	percentage float64,
	message, currentFile string,
) Progress {
	return Progress{
		Stage:       stage,
		Completed:   completed,
		Total:       total,
		Percentage:  percentage,
		Message:     message,
		CurrentFile: currentFile,
	}
}

// TestDetectionMethod_Values_Basic tests detection method constants.
func TestDetectionMethod_Values_Basic(t *testing.T) {
	methods := []DetectionMethod{
		MethodArtDupl,
		MethodHash,
	}

	for _, method := range methods {
		if string(method) == "" {
			t.Error("Detection method should not be empty string")
		}
	}
}

// TestDefaultOptions_Basic tests default options creation.
func TestDefaultOptions_Basic(t *testing.T) {
	opts := DefaultOptions()

	// Test that default threshold is reasonable
	if opts.Threshold <= 0 || opts.Threshold > 1000 {
		t.Errorf("Default threshold should be reasonable, got %d", opts.Threshold)
	}

	// Test that timeout is set
	if opts.Timeout <= 0 {
		t.Error("Default timeout should be positive")
	}
}

// TestValidateOptions_Valid_Basic tests valid options.
func TestValidateOptions_Valid_Basic(t *testing.T) {
	opts := DefaultOptions()

	err := ValidateOptions(opts)
	if err != nil {
		t.Errorf("Valid options should pass validation: %v", err)
	}
}

// newInvalidThresholdTestCase creates a test case for invalid threshold validation.
func newInvalidThresholdTestCase(name string, threshold int) struct {
	name    string
	opts    *Options
	wantErr bool
} {
	return struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		name: name,
		opts: &Options{
			Threshold:        threshold,
			DetectionMethods: []DetectionMethod{MethodArtDupl},
		},
		wantErr: true,
	}
}

// TestValidateOptions_Invalid_Basic tests invalid options.
func TestValidateOptions_Invalid_Basic(t *testing.T) {
	testCases := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
		},
		newInvalidThresholdTestCase("threshold too low", 0),
		newInvalidThresholdTestCase("threshold too high", 1001),
		{
			name:    "no detection methods",
			opts:    newTestOptionsWithNoDetectionMethods(),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOptions(tc.opts)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestErrorComparison_Basic tests error comparison behavior.
func TestErrorComparison_Basic(t *testing.T) {
	testutil.AssertErrorIs(t, ErrNilOptions, ErrNilOptions, "Error should match itself")

	if errors.Is(ErrNilOptions, ErrInvalidThreshold) {
		t.Error("Different errors should not match")
	}

	testutil.AssertErrorIs(t, ErrNilOptions, ErrNilOptions, "errors.Is should match same")

	if errors.Is(ErrNilOptions, ErrInvalidThreshold) {
		t.Error("errors.Is should not match different errors")
	}
}

// TestErrorWrapping_Basic tests error wrapping functionality.
func TestErrorWrapping_Basic(t *testing.T) {
	// Test wrapping SDK errors
	wrappedErr := fmt.Errorf("validation failed: %w", ErrInvalidThreshold)

	if !errors.Is(wrappedErr, ErrInvalidThreshold) {
		t.Error("Wrapped error should be recognizable with errors.Is")
	}

	// Test double wrapping
	doubleWrapped := fmt.Errorf("setup failed: %w", wrappedErr)

	if !errors.Is(doubleWrapped, ErrInvalidThreshold) {
		t.Error("Double wrapped error should be recognizable")
	}
}

// TestCloneGroup_Validation_Basic tests clone group structure.
func TestCloneGroup_Validation_Basic(t *testing.T) {
	group := CloneGroup{
		Hash:   "test-hash",
		Clones: []*Clone{{Filename: testFilename}},
		Size:   10,
		Method: MethodArtDupl,
	}

	if group.Hash == "" {
		t.Error("Clone group should have hash")
	}

	if len(group.Clones) == 0 {
		t.Error("Clone group should have at least one clone")
	}

	if group.Size <= 0 {
		t.Error("Clone group size should be positive")
	}

	if group.Method == "" {
		t.Error("Clone group should have detection method")
	}
}

// TestClone_IsValid tests Clone validation.
func TestClone_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		clone   Clone
		wantErr error
	}{
		{
			name:    "valid clone with positions",
			clone:   Clone{StartLine: 1, EndLine: 5, StartPos: 10, EndPos: 50},
			wantErr: nil,
		},
		{
			clone:   Clone{StartLine: 5, EndLine: 5, StartPos: 10, EndPos: 20},
			wantErr: nil,
			name:    "valid testClone single line",
		},
		{
			name:    "invalid clone without positions (zero length)",
			clone:   Clone{StartLine: 1, EndLine: 10},
			wantErr: ErrCloneZeroLength,
		},
		{
			name:    "invalid end line before start",
			clone:   Clone{StartLine: 10, EndLine: 5},
			wantErr: ErrCloneEndLineBeforeStart,
		},
		{
			clone:   Clone{StartLine: 1, EndLine: 1, StartPos: 50, EndPos: 50},
			wantErr: ErrCloneZeroLength,
			name:    "invalid zero length positions",
		},
		{
			name:    "invalid negative length positions",
			wantErr: ErrCloneZeroLength,
			clone:   Clone{StartLine: 1, EndLine: 1, StartPos: 50, EndPos: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValidErr := tt.clone.IsValid()

			expectedErr := tt.wantErr
			if !errors.Is(isValidErr, expectedErr) {
				t.Errorf("Clone.IsValid() = %v, want %v", isValidErr, expectedErr)
			}
		})
	}
}

// TestProgress_Validation_Basic tests progress structure.
func TestProgress_Validation_Basic(t *testing.T) {
	progress := newTestProgress("parsing", 50, 100, 50.0, "Processing files", "test.go")

	if progress.Stage == "" {
		t.Error("Stage should not be empty")
	}

	if progress.Completed < 0 {
		t.Error("Completed should be non-negative")
	}

	if progress.Total <= 0 {
		t.Error("Total should be positive")
	}

	if progress.Percentage < 0 || progress.Percentage > 100 {
		t.Error("Percentage should be between 0-100")
	}

	if progress.Message == "" {
		t.Error("Message should not be empty")
	}
}

// TestLoggerInterface_Basic tests logger interface compliance.
func TestLoggerInterface_Basic(t *testing.T) {
	var l Logger = &logger.NoOpLogger{}

	// Test that all methods are implemented
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warning message")
	l.Error("error message")

	// Should not panic
}
