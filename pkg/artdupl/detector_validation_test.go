package artdupl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
)

func validOpts(threshold int) *Options {
	return &Options{
		Threshold:        threshold,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxWorkers:       1,
	}
}

// TestDefaultOptions_Values tests that DefaultOptions returns sensible values.
func TestDefaultOptions_Values(t *testing.T) {
	opts := DefaultOptions()

	if opts.Threshold != 15 {
		t.Errorf("Default threshold should be 15, got %d", opts.Threshold)
	}

	if len(opts.DetectionMethods) != 1 {
		t.Errorf("Default should have 1 detection method, got %d", len(opts.DetectionMethods))
	}

	if opts.DetectionMethods[0] != MethodArtDupl {
		t.Errorf("Default method should be MethodArtDupl, got %v", opts.DetectionMethods[0])
	}

	if opts.MaxFileSize != 10*1024*1024 {
		t.Errorf("Default max file size should be 10MB, got %d", opts.MaxFileSize)
	}

	if opts.MaxWorkers != 4 {
		t.Errorf("Default max workers should be 4, got %d", opts.MaxWorkers)
	}

	if opts.Timeout != 30*time.Minute {
		t.Errorf("Default timeout should be 30m, got %v", opts.Timeout)
	}

	if opts.MaxClonesPerGroup != 50 {
		t.Errorf("Default max clones per group should be 50, got %d", opts.MaxClonesPerGroup)
	}
}

// TestValidateOptions_AllErrors tests all validation errors.
func TestValidateOptions_AllErrors(t *testing.T) {
	testCases := []struct {
		name    string
		opts    *Options
		wantErr error
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: ErrNilOptions,
		},
		{
			name:    "zero threshold",
			opts:    newTestOptionsWithThreshold(0),
			wantErr: ErrInvalidThreshold,
		},
		{
			name:    "negative threshold",
			opts:    newTestOptionsWithThreshold(-5),
			wantErr: ErrInvalidThreshold,
		},
		{
			name:    "threshold too large",
			opts:    newTestOptionsWithThreshold(1001),
			wantErr: ErrThresholdTooLarge,
		},
		{
			name:    "no detection methods",
			opts:    newTestOptionsWithNoDetectionMethods(),
			wantErr: ErrNoDetectionMethods,
		},
		{
			name: "nil detection methods",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: nil,
			},
			wantErr: ErrNoDetectionMethods,
		},
		{
			name: "negative max file size",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxFileSize:      -1,
			},
			wantErr: ErrInvalidMaxFileSize,
		},
		{
			name: "zero max workers",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxWorkers:       0,
			},
			wantErr: ErrInvalidMaxWorkers,
		},
		{
			name: "negative max workers",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxWorkers:       -1,
			},
			wantErr: ErrInvalidMaxWorkers,
		},
		{
			name: "negative timeout",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxWorkers:       1,
				Timeout:          -1 * time.Second,
			},
			wantErr: ErrInvalidTimeout,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOptions(tc.opts)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Expected %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

// TestValidateOptions_ValidCases tests valid options configurations.
func TestValidateOptions_ValidCases(t *testing.T) {
	testCases := []struct {
		name string
		opts *Options
	}{
		{name: "min threshold", opts: validOpts(1)},
		{name: "max threshold", opts: validOpts(1000)},
		{
			name: "multiple detection methods",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl, MethodHash},
				MaxWorkers:       1,
			},
		},
		{
			name: "all fields set",
			opts: &Options{
				Threshold:         20,
				DetectionMethods:  []DetectionMethod{MethodArtDupl},
				IncludeVendor:     true,
				IgnoreFiles:       []string{"*_test.go"},
				MaxFileSize:       5 * 1024 * 1024,
				MaxWorkers:        8,
				Timeout:           time.Hour,
				IncludeFragments:  true,
				MaxClonesPerGroup: 100,
			},
		},
		{
			name: "zero max file size (unlimited)",
			opts: func() *Options {
				o := validOpts(15)
				o.MaxFileSize = 0

				return o
			}(),
		},
		{
			name: "zero timeout (no timeout)",
			opts: func() *Options {
				o := validOpts(15)
				o.Timeout = 0

				return o
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOptions(tc.opts)
			if err != nil {
				t.Errorf("Valid options should not error, got: %v", err)
			}
		})
	}
}

// TestValidateOptions_InvalidDetectionMethod tests validation with invalid detection method.
func TestValidateOptions_InvalidDetectionMethod(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{"invalid-method"},
		MaxWorkers:       4,
	}

	err := ValidateOptions(opts)
	if err == nil {
		t.Error("Invalid detection method should cause validation error")
	}
}

// TestDetectionMethod_Constants tests detection method constants.
func TestDetectionMethod_Constants(t *testing.T) {
	methods := []struct {
		name   string
		method DetectionMethod
	}{
		{"MethodArtDupl", MethodArtDupl},
		{"MethodHash", MethodHash},
		{"MethodTodos", MethodTodos},
		{"MethodLegacy", MethodLegacy},
	}

	for _, tc := range methods {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.method) == "" {
				t.Errorf("%s should not be empty string", tc.name)
			}
		})
	}
}

// TestDetectionMethod_Equality tests detection method equality.
func TestDetectionMethod_Equality(t *testing.T) {
	if MethodArtDupl != config.DetectionMethodArtDupl {
		t.Error("MethodArtDupl should equal config.DetectionMethodArtDupl")
	}

	if MethodHash != config.DetectionMethodHash {
		t.Error("MethodHash should equal config.DetectionMethodHash")
	}

	if MethodTodos != config.DetectionMethodTodos {
		t.Error("MethodTodos should equal config.DetectionMethodTodos")
	}

	if MethodLegacy != config.DetectionMethodLegacy {
		t.Error("MethodLegacy should equal config.DetectionMethodLegacy")
	}
}

// TestConvertOptionsToConfig tests options to config conversion.
func TestConvertOptionsToConfig(t *testing.T) {
	opts := &Options{
		Threshold:        25,
		DetectionMethods: []DetectionMethod{MethodArtDupl, MethodHash},
		IncludeVendor:    true,
		IgnoreFiles:      []string{"*_test.go", "vendor/*"},
	}

	cfg := convertOptionsToConfig(opts)

	if cfg.Threshold != 25 {
		t.Errorf("Threshold should be 25, got %d", cfg.Threshold)
	}

	if len(cfg.DetectionMethods) != 2 {
		t.Errorf("Should have 2 detection methods, got %d", len(cfg.DetectionMethods))
	}

	if !cfg.IncludeVendor {
		t.Error("IncludeVendor should be true")
	}

	if len(cfg.IgnoreFiles) != 2 {
		t.Errorf("Should have 2 ignore files, got %d", len(cfg.IgnoreFiles))
	}
}

// TestErrors_AllErrors tests that all exported errors exist.
func TestErrors_AllErrors(t *testing.T) {
	allErrors := []error{
		ErrNilOptions,
		ErrInvalidThreshold,
		ErrThresholdTooLarge,
		ErrNoDetectionMethods,
		ErrInvalidMaxFileSize,
		ErrInvalidMaxWorkers,
		ErrInvalidTimeout,
		ErrUnsupportedMethod,
		ErrNoFilesProvided,
		ErrFileNotFound,
		ErrFileTooLarge,
		ErrParsingFailed,
		ErrContextCanceled,
		ErrAnalysisTimeout,
		ErrNoDuplicatesFound,
		ErrResultProcessing,
		ErrMemoryLimit,
		ErrInternal,
	}

	for i, err := range allErrors {
		if err == nil {
			t.Errorf("Error at index %d should not be nil", i)
		}

		if err.Error() == "" {
			t.Errorf("Error at index %d should have message", i)
		}
	}
}

// TestProgressCallback tests progress callback functionality.
func TestProgressCallback(t *testing.T) {
	var receivedProgress *Progress

	callback := func(p *Progress) error {
		receivedProgress = p

		return nil
	}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		ProgressCallback: callback,
		MaxWorkers:       4,
		Timeout:          time.Minute,
	}

	err := ValidateOptions(opts)
	if err != nil {
		t.Fatalf("Valid options should not error: %v", err)
	}

	testProgress := &Progress{
		Stage:      "test",
		Completed:  50,
		Total:      100,
		Percentage: 50.0,
		Message:    "Testing",
	}

	err = callback(testProgress)
	if err != nil {
		t.Errorf("Callback should not error, got: %v", err)
	}

	if receivedProgress != testProgress {
		t.Error("Callback should receive progress")
	}
}

// TestOptions_WithFileReader tests custom file reader option.
func TestOptions_WithFileReader(t *testing.T) {
	customReader := func(filename string) ([]byte, error) {
		return []byte("custom content"), nil
	}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		FileReader:       customReader,
		MaxWorkers:       4,
	}

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}
}

// TestOptions_WithLogger tests custom logger option.
func TestOptions_WithLogger(t *testing.T) {
	logger := &testLogger{}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		Logger:           logger,
		MaxWorkers:       4,
	}

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}
}

// testLogger implements Logger interface for testing.
type testLogger struct{}

func (l *testLogger) Debug(msg string, args ...any) {}
func (l *testLogger) Info(msg string, args ...any)  {}
func (l *testLogger) Warn(msg string, args ...any)  {}
func (l *testLogger) Error(msg string, args ...any) {}

// TestFileReaderFunc tests the FileReaderFunc type.
func TestFileReaderFunc(t *testing.T) {
	var reader FileReaderFunc = func(filename string) ([]byte, error) {
		return []byte("test content"), nil
	}

	content, err := reader("test.go")
	if err != nil {
		t.Errorf("FileReaderFunc should not error, got: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Content should be 'test content', got %s", string(content))
	}
}

// TestDetector_Close tests that Close returns nil.
func TestDetector_Close(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	err = detector.Close()
	if err != nil {
		t.Errorf("Close() should not error, got: %v", err)
	}
}

// TestDetector_FindClones_NoFiles tests FindClones with no files.
func TestDetector_FindClones_NoFiles(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	_, err = detector.FindClones(t.Context(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}
}

// TestDetector_FindClones_NilFiles tests FindClones with nil files slice.
func TestDetector_FindClones_NilFiles(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	_, err = detector.FindClones(t.Context(), nil)
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}
}

// TestDetector_FindClones_ContextCanceled tests FindClones with canceled context.
func TestDetector_FindClones_ContextCanceled(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = detector.FindClones(ctx, []string{"some_file.go"})
	if err == nil {
		t.Error("Expected error with canceled context")
	}
}

// TestDetector_FindClonesStream_NoFiles tests FindClonesStream with no files.
func TestDetector_FindClonesStream_NoFiles(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	_, err = detector.FindClonesStream(t.Context(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}
}

// TestDetector_FindClonesStream_Cancellation tests stream detection with context cancellation.
func TestDetector_FindClonesStream_Cancellation(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = detector.FindClonesStream(ctx, []string{"test.go"})
	if err == nil {
		t.Error("Expected error with canceled context")
	}
}

// TestDetector_Reuse tests that a detector can be reused.
func TestDetector_Reuse(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	_, err = detector.FindClones(t.Context(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}

	_, err = detector.FindClones(t.Context(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided on reuse, got: %v", err)
	}
}
