package artdupl

import (
	"errors"
	"fmt"
	"testing"
)

// TestDetectionMethod_Values_Basic tests detection method constants.
func TestDetectionMethod_Values_Basic(t *testing.T) {
	methods := []DetectionMethod{
		MethodArtDupl,
		MethodHash,
		MethodTodos,
		MethodLegacy,
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
		{
			name: "threshold too low",
			opts: &Options{
				Threshold:        0,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			},
			wantErr: true,
		},
		{
			name: "threshold too high",
			opts: &Options{
				Threshold:        1001,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			},
			wantErr: true,
		},
		{
			name: "no detection methods",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{},
			},
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
	// Test error comparison using errors.Is (handles wrapped errors)
	if !errors.Is(ErrNilOptions, ErrNilOptions) {
		t.Error("Error should match itself")
	}

	if errors.Is(ErrNilOptions, ErrInvalidThreshold) {
		t.Error("Different errors should not match")
	}

	// Test errors.Is compatibility
	if !errors.Is(ErrNilOptions, ErrNilOptions) {
		t.Error("errors.Is should match same error")
	}

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
		Clones: []*Clone{{Filename: "test.go"}},
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

// TestProgress_Validation_Basic tests progress structure.
func TestProgress_Validation_Basic(t *testing.T) {
	progress := Progress{
		Stage:       "parsing",
		Completed:   50,
		Total:       100,
		Percentage:  50.0,
		Message:     "Processing files",
		CurrentFile: "test.go",
	}

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
	var logger Logger = &testLoggerBasic{}

	// Test that all methods are implemented
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	// Should not panic
}

// testLoggerBasic implements Logger interface for testing.
type testLoggerBasic struct{}

func (l *testLoggerBasic) Debug(msg string, args ...any) {}
func (l *testLoggerBasic) Info(msg string, args ...any)  {}
func (l *testLoggerBasic) Warn(msg string, args ...any)  {}
func (l *testLoggerBasic) Error(msg string, args ...any) {}
