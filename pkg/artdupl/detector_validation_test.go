package artdupl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
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

	testutil.AssertFieldValue(t, opts.Threshold, 15, "Threshold")

	if len(opts.DetectionMethods) != 1 {
		t.Errorf("Default should have 1 detection method, got %d", len(opts.DetectionMethods))
	}

	if opts.DetectionMethods[0] != MethodArtDupl {
		t.Errorf("Default method should be MethodArtDupl, got %v", opts.DetectionMethods[0])
	}

	testutil.AssertFieldValue(t, opts.MaxFileSize, 10*1024*1024, "MaxFileSize")

	testutil.AssertFieldValue(t, opts.MaxWorkers, 4, "MaxWorkers")

	testutil.AssertFieldValue(t, opts.Timeout, 30*time.Minute, "Timeout")

	testutil.AssertFieldValue(t, opts.MaxClonesPerGroup, 50, "MaxClonesPerGroup")
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

	testutil.AssertFieldValue(t, cfg.Threshold, 25, "Threshold")

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

// TestProgressCallback tests progress progressHandler functionality.
func TestProgressCallback(t *testing.T) {
	var receivedProgress *Progress

	progressHandler := func(p *Progress) error {
		receivedProgress = p

		return nil
	}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		ProgressCallback: progressHandler,
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

	err = progressHandler(testProgress)
	if err != nil {
		t.Errorf("Callback should not error, got: %v", err)
	}

	if receivedProgress != testProgress {
		t.Error("Callback should receive progress")
	}
}

// TestOptions_WithFileReader tests custom file reader option.
func TestOptions_WithFileReader(t *testing.T) {
	testReader := func(fname string) ([]byte, error) {
		return []byte("custom content"), nil
	}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		FileReader:       testReader,
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
	log := &logger.NoOpLogger{}

	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		Logger:           log,
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

// TestFileReaderFunc tests the FileReaderFunc type.
func TestFileReaderFunc(t *testing.T) {
	var reader FileReaderFunc = func(_ string) ([]byte, error) {
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

// mustCreateDetector creates a detector with the given options or fails the test.
func mustCreateDetector(t *testing.T, opts *Options) Detector {
	t.Helper()

	detector, actualErr := NewDetector(opts)
	if actualErr != nil {
		t.Fatalf("Failed to create detector: %v", actualErr)
	}

	return detector
}

// TestDetector_FindClones_NoFiles tests FindClones with no files.
func TestDetector_FindClones_NoFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
	}{
		{"empty files", []string{}},
		{"nil files", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mustCreateDetector(t, DefaultOptions())

			_, returnedErr := detector.FindClones(t.Context(), tt.files)
			if !errors.Is(returnedErr, ErrNoFilesProvided) {
				t.Errorf("Expected ErrNoFilesProvided, got: %v", returnedErr)
			}
		})
	}
}

// TestDetector_FindClones_ContextCanceled tests FindClones with canceled context.
func TestDetector_FindClones_ContextCanceled(t *testing.T) {
	detector := mustCreateDetector(t, DefaultOptions())

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := detector.FindClones(ctx, []string{"some_file.go"})
	if err == nil {
		t.Error("Expected error with canceled context")
	}
}

// TestDetector_FindClonesStream_NoFiles tests FindClonesStream with no files.
func TestDetector_FindClonesStream_NoFiles(t *testing.T) {
	opts := DefaultOptions()

	detector, actualErr := NewDetector(opts)
	if actualErr != nil {
		t.Fatalf("Failed to create detector: %v", actualErr)
	}

	_, actualErr = detector.FindClonesStream(t.Context(), []string{})
	if !errors.Is(actualErr, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", actualErr)
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

	detector, returnedErr := NewDetector(opts)
	if returnedErr != nil {
		t.Fatalf("Failed to create detector: %v", returnedErr)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	_, returnedErr = detector.FindClones(t.Context(), []string{})
	if !errors.Is(returnedErr, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", returnedErr)
	}

	_, returnedErr = detector.FindClones(t.Context(), []string{})
	if !errors.Is(returnedErr, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided on reuse, got: %v", returnedErr)
	}
}
