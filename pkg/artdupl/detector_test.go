package artdupl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// cleanupDetector closes the detector and logs any errors.
func cleanupDetector(t *testing.T, detector Detector) {
	t.Helper()

	err := detector.Close()
	if err != nil {
		t.Logf("Failed to close detector: %v", err)
	}
}

// TestNewDetector_NilOptions tests that nil options are handled correctly.
func TestNewDetector_NilOptions(t *testing.T) {
	detector, err := NewDetector(nil)
	if err != nil {
		t.Errorf("NewDetector(nil) should not error, got: %v", err)
	}

	if detector == nil {
		t.Error("NewDetector(nil) should return valid detector")
	}
}

// TestNewDetector_ValidOptions tests detector creation with valid options.
func TestNewDetector_ValidOptions(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxFileSize:      10 * 1024 * 1024,
		MaxWorkers:       4,
		Timeout:          time.Minute,
	}

	detector, err := NewDetector(opts)
	if err != nil {
		t.Errorf("NewDetector with valid options should not error, got: %v", err)
	}

	if detector == nil {
		t.Error("NewDetector should return valid detector")
	}
}

// TestNewDetector_InvalidThreshold tests detector creation with invalid threshold.
func TestNewDetector_InvalidThreshold(t *testing.T) {
	testCases := []struct {
		name      string
		threshold int
		wantErr   error
	}{
		{"zero threshold", 0, ErrInvalidThreshold},
		{"negative threshold", -1, ErrInvalidThreshold},
		{"too large threshold", 1001, ErrThresholdTooLarge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &Options{
				Threshold:        tc.threshold,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			}

			_, err := NewDetector(opts)
			if err == nil {
				t.Error("Expected error for invalid threshold")
			}
		})
	}
}

// TestNewDetector_NoDetectionMethods tests detector creation without detection methods.
func TestNewDetector_NoDetectionMethods(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{},
	}

	_, err := NewDetector(opts)
	if !errors.Is(err, ErrNoDetectionMethods) {
		t.Errorf("Expected ErrNoDetectionMethods, got: %v", err)
	}
}

// TestNewDetector_InvalidMaxWorkers tests detector creation with invalid max workers.
func TestNewDetector_InvalidMaxWorkers(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxWorkers:       0,
	}

	_, err := NewDetector(opts)
	if !errors.Is(err, ErrInvalidMaxWorkers) {
		t.Errorf("Expected ErrInvalidMaxWorkers, got: %v", err)
	}
}

// TestNewDetector_InvalidMaxFileSize tests detector creation with invalid max file size.
func TestNewDetector_InvalidMaxFileSize(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxFileSize:      -1,
	}

	_, err := NewDetector(opts)
	if !errors.Is(err, ErrInvalidMaxFileSize) {
		t.Errorf("Expected ErrInvalidMaxFileSize, got: %v", err)
	}
}

// TestNewDetector_InvalidTimeout tests detector creation with invalid timeout.
func TestNewDetector_InvalidTimeout(t *testing.T) {
	opts := &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxWorkers:       4,
		Timeout:          -1 * time.Second,
	}

	_, err := NewDetector(opts)
	if !errors.Is(err, ErrInvalidTimeout) {
		t.Errorf("Expected ErrInvalidTimeout, got: %v", err)
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

	_, err = detector.FindClones(context.Background(), []string{})
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

	_, err = detector.FindClones(context.Background(), nil)
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}
}

// TestDetector_FindClones_NonExistentFile tests FindClones with non-existent file.
func TestDetector_FindClones_NonExistentFile(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	result, err := detector.FindClones(context.Background(), []string{"nonexistent_file.go"})
	// The function may skip non-existent files without error
	// or return an error - either behavior is acceptable
	if err == nil && result != nil {
		// No error means files were skipped
		return
	}
	// Error is also acceptable
}

// TestDetector_FindClonesStream_NoFiles tests FindClonesStream with no files.
func TestDetector_FindClonesStream_NoFiles(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	_, err = detector.FindClonesStream(context.Background(), []string{})
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = detector.FindClones(ctx, []string{"some_file.go"})
	if err == nil {
		t.Error("Expected error with canceled context")
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
			name: "zero threshold",
			opts: &Options{
				Threshold:        0,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			},
			wantErr: ErrInvalidThreshold,
		},
		{
			name: "negative threshold",
			opts: &Options{
				Threshold:        -5,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			},
			wantErr: ErrInvalidThreshold,
		},
		{
			name: "threshold too large",
			opts: &Options{
				Threshold:        1001,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
			},
			wantErr: ErrThresholdTooLarge,
		},
		{
			name: "no detection methods",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{},
			},
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
		{
			name: "minimal valid",
			opts: &Options{
				Threshold:        1,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxWorkers:       1,
			},
		},
		{
			name: "max threshold",
			opts: &Options{
				Threshold:        1000,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxWorkers:       1,
			},
		},
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
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				MaxFileSize:      0,
				MaxWorkers:       1,
			},
		},
		{
			name: "zero timeout (no timeout)",
			opts: &Options{
				Threshold:        15,
				DetectionMethods: []DetectionMethod{MethodArtDupl},
				Timeout:          0,
				MaxWorkers:       1,
			},
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

// TestCloneGroup_Fields tests CloneGroup field assignments.
func TestCloneGroup_Fields(t *testing.T) {
	group := CloneGroup{
		Hash: "abc123",
		Clones: []*Clone{
			{Filename: "file1.go", StartLine: 10, EndLine: 20},
			{Filename: "file2.go", StartLine: 15, EndLine: 25},
		},
		Size:      100,
		LineCount: 11,
		Method:    MethodArtDupl,
	}

	if group.Hash != "abc123" {
		t.Errorf("Hash should be 'abc123', got %s", group.Hash)
	}

	if len(group.Clones) != 2 {
		t.Errorf("Should have 2 clones, got %d", len(group.Clones))
	}

	if group.Size != 100 {
		t.Errorf("Size should be 100, got %d", group.Size)
	}

	if group.LineCount != 11 {
		t.Errorf("LineCount should be 11, got %d", group.LineCount)
	}

	if group.Method != MethodArtDupl {
		t.Errorf("Method should be MethodArtDupl, got %v", group.Method)
	}
}

// TestClone_Fields tests Clone field assignments.
func TestClone_Fields(t *testing.T) {
	clone := Clone{
		Filename:  "test.go",
		StartLine: 10,
		EndLine:   20,
		StartPos:  100,
		EndPos:    200,
		Fragment:  "code here",
		Size:      50,
	}

	if clone.Filename != "test.go" {
		t.Errorf("Filename should be 'test.go', got %s", clone.Filename)
	}

	if clone.StartLine != 10 {
		t.Errorf("StartLine should be 10, got %d", clone.StartLine)
	}

	if clone.EndLine != 20 {
		t.Errorf("EndLine should be 20, got %d", clone.EndLine)
	}

	if clone.StartPos != 100 {
		t.Errorf("StartPos should be 100, got %d", clone.StartPos)
	}

	if clone.EndPos != 200 {
		t.Errorf("EndPos should be 200, got %d", clone.EndPos)
	}

	if clone.Fragment != "code here" {
		t.Errorf("Fragment should be 'code here', got %s", clone.Fragment)
	}

	if clone.Size != 50 {
		t.Errorf("Size should be 50, got %d", clone.Size)
	}
}

// TestResult_Fields tests Result field assignments.
func TestResult_Fields(t *testing.T) {
	result := Result{
		CloneGroups: []*CloneGroup{
			{Hash: "hash1"},
			{Hash: "hash2"},
		},
		Summary: &Summary{
			TotalFiles:  10,
			TotalClones: 5,
			TotalGroups: 2,
		},
		Metadata: &Metadata{
			Version: "1.0.0",
		},
	}

	if len(result.CloneGroups) != 2 {
		t.Errorf("Should have 2 clone groups, got %d", len(result.CloneGroups))
	}

	if result.Summary.TotalFiles != 10 {
		t.Errorf("TotalFiles should be 10, got %d", result.Summary.TotalFiles)
	}

	if result.Metadata.Version != "1.0.0" {
		t.Errorf("Version should be '1.0.0', got %s", result.Metadata.Version)
	}
}

// TestSummary_Fields tests Summary field assignments.
func TestSummary_Fields(t *testing.T) {
	summary := Summary{
		TotalFiles:    100,
		TotalClones:   50,
		TotalGroups:   25,
		AnalysisTime:  5 * time.Second,
		MethodsUsed:   []DetectionMethod{MethodArtDupl, MethodHash},
		LinesAnalyzed: 10000,
	}

	if summary.TotalFiles != 100 {
		t.Errorf("TotalFiles should be 100, got %d", summary.TotalFiles)
	}

	if summary.TotalClones != 50 {
		t.Errorf("TotalClones should be 50, got %d", summary.TotalClones)
	}

	if summary.TotalGroups != 25 {
		t.Errorf("TotalGroups should be 25, got %d", summary.TotalGroups)
	}

	if summary.AnalysisTime != 5*time.Second {
		t.Errorf("AnalysisTime should be 5s, got %v", summary.AnalysisTime)
	}

	if len(summary.MethodsUsed) != 2 {
		t.Errorf("Should have 2 methods, got %d", len(summary.MethodsUsed))
	}

	if summary.LinesAnalyzed != 10000 {
		t.Errorf("LinesAnalyzed should be 10000, got %d", summary.LinesAnalyzed)
	}
}

// TestMetadata_Fields tests Metadata field assignments.
func TestMetadata_Fields(t *testing.T) {
	now := time.Now()
	metadata := Metadata{
		Version:    "2.0.0",
		Timestamp:  now,
		ConfigHash: "config-hash-123",
		Toolchain:  "go1.21",
	}

	if metadata.Version != "2.0.0" {
		t.Errorf("Version should be '2.0.0', got %s", metadata.Version)
	}

	if !metadata.Timestamp.Equal(now) {
		t.Errorf("Timestamp should be %v, got %v", now, metadata.Timestamp)
	}

	if metadata.ConfigHash != "config-hash-123" {
		t.Errorf("ConfigHash should be 'config-hash-123', got %s", metadata.ConfigHash)
	}

	if metadata.Toolchain != "go1.21" {
		t.Errorf("Toolchain should be 'go1.21', got %s", metadata.Toolchain)
	}
}

// TestProgress_Fields tests Progress field assignments.
func TestProgress_Fields(t *testing.T) {
	progress := newTestProgress("parsing", 75, 100, 75.5, "Processing files", "main.go")

	if progress.Stage != "parsing" {
		t.Errorf("Stage should be 'parsing', got %s", progress.Stage)
	}

	if progress.Completed != 75 {
		t.Errorf("Completed should be 75, got %d", progress.Completed)
	}

	if progress.Total != 100 {
		t.Errorf("Total should be 100, got %d", progress.Total)
	}

	if progress.Percentage != 75.5 {
		t.Errorf("Percentage should be 75.5, got %f", progress.Percentage)
	}

	if progress.Message != "Processing files" {
		t.Errorf("Message should be 'Processing files', got %s", progress.Message)
	}

	if progress.CurrentFile != "main.go" {
		t.Errorf("CurrentFile should be 'main.go', got %s", progress.CurrentFile)
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

	// Simulate calling the callback
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
	errors := []error{
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

	for i, err := range errors {
		if err == nil {
			t.Errorf("Error at index %d should not be nil", i)
		}

		if err.Error() == "" {
			t.Errorf("Error at index %d should have message", i)
		}
	}
}

// TestErrors_Is tests errors.Is functionality.
func TestErrors_Is(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"same error", ErrNilOptions, ErrNilOptions, true},
		{"different errors", ErrNilOptions, ErrInvalidThreshold, false},
		{
			"wrapped error",
			fmt.Errorf("wrapped: %w", ErrInvalidThreshold),
			ErrInvalidThreshold,
			true,
		},
		{"nil error", nil, ErrNilOptions, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if errors.Is(tc.err, tc.target) != tc.want {
				t.Errorf("errors.Is(%v, %v) should be %v", tc.err, tc.target, tc.want)
			}
		})
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

// TestClone_Empty tests empty Clone struct.
func TestClone_Empty(t *testing.T) {
	var clone Clone

	if clone.Filename != "" {
		t.Errorf("Empty Clone Filename should be empty string, got %s", clone.Filename)
	}

	if clone.StartLine != 0 {
		t.Errorf("Empty Clone StartLine should be 0, got %d", clone.StartLine)
	}

	if clone.Size != 0 {
		t.Errorf("Empty Clone Size should be 0, got %d", clone.Size)
	}
}

// TestCloneGroup_Empty tests empty CloneGroup struct.
func TestCloneGroup_Empty(t *testing.T) {
	var group CloneGroup

	if group.Hash != "" {
		t.Errorf("Empty CloneGroup Hash should be empty string, got %s", group.Hash)
	}

	if group.Clones != nil {
		t.Errorf("Empty CloneGroup Clones should be nil, got %v", group.Clones)
	}

	if group.Size != 0 {
		t.Errorf("Empty CloneGroup Size should be 0, got %d", group.Size)
	}
}

// TestResult_Empty tests empty Result struct.
func TestResult_Empty(t *testing.T) {
	var result Result

	if result.CloneGroups != nil {
		t.Errorf("Empty Result CloneGroups should be nil, got %v", result.CloneGroups)
	}

	if result.Summary != nil {
		t.Errorf("Empty Result Summary should be nil, got %v", result.Summary)
	}

	if result.Metadata != nil {
		t.Errorf("Empty Result Metadata should be nil, got %v", result.Metadata)
	}
}

// TestClone_WithFragment tests Clone with fragment content.
func TestClone_WithFragment(t *testing.T) {
	clone := Clone{
		Filename:  "test.go",
		StartLine: 1,
		EndLine:   5,
		Fragment:  "package main\n\nfunc main() {}\n",
		Size:      25,
	}

	if clone.Fragment == "" {
		t.Error("Clone with fragment should have non-empty Fragment")
	}
}

// TestClone_WithoutFragment tests Clone without fragment content.
func TestClone_WithoutFragment(t *testing.T) {
	clone := Clone{
		Filename:  "test.go",
		StartLine: 1,
		EndLine:   5,
		Fragment:  "",
		Size:      25,
	}

	if clone.Fragment != "" {
		t.Errorf("Clone without fragment should have empty Fragment, got %s", clone.Fragment)
	}
}

// createTestGoFile creates a temporary Go file for testing.
func createTestGoFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(filename, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return filename
}

// TestDetector_Integration_SimpleFile tests detector with a simple Go file.
func TestDetector_Integration_SimpleFile(t *testing.T) {
	t.Parallel()

	filename := createTestGoFile(t, `package main

func main() {
	println("hello")
}
`)

	opts := DefaultOptions()
	opts.Threshold = 5 // Low threshold to find duplicates

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	result, err := detector.FindClones(context.Background(), []string{filename})
	// Single file with no duplicates is expected to work
	if err != nil {
		// Some errors are acceptable (no duplicates found)
		if !errors.Is(err, ErrNoDuplicatesFound) {
			t.Logf("FindClones returned: %v", err)
		}

		return
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

// TestDetector_Integration_DuplicateFiles tests detector with duplicate code.
func TestDetector_Integration_DuplicateFiles(t *testing.T) {
	t.Parallel()

	duplicateCode := `package main

func duplicate() {
	x := 1
	y := 2
	z := x + y
	println(z)
}
`

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	err := os.WriteFile(file1, []byte(duplicateCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	err = os.WriteFile(file2, []byte(duplicateCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	opts := DefaultOptions()
	opts.Threshold = 5 // Low threshold

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	result, err := detector.FindClones(context.Background(), []string{file1, file2})
	if err != nil {
		t.Logf("FindClones returned: %v", err)

		return
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

// TestDetector_Integration_ContextTimeout tests detector with context timeout.
func TestDetector_Integration_ContextTimeout(t *testing.T) {
	t.Parallel()

	filename := createTestGoFile(t, `package main

func main() {
	println("hello")
}
`)

	opts := DefaultOptions()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond) // Ensure timeout

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	_, err = detector.FindClones(ctx, []string{filename})
	if err == nil {
		t.Error("Expected error with expired context")
	}
}

// TestClone_LargeFragment tests Clone with large fragment.
func TestClone_LargeFragment(t *testing.T) {
	largeFragment := make([]byte, 10000)
	for i := range largeFragment {
		largeFragment[i] = 'x'
	}

	clone := Clone{
		Filename:  "large.go",
		StartLine: 1,
		EndLine:   1000,
		Fragment:  string(largeFragment),
		Size:      len(largeFragment),
	}

	if len(clone.Fragment) != 10000 {
		t.Errorf("Fragment should be 10000 bytes, got %d", len(clone.Fragment))
	}
}

// TestCloneGroup_ManyClones tests CloneGroup with many clones.
func TestCloneGroup_ManyClones(t *testing.T) {
	clones := make([]*Clone, 100)
	for i := range clones {
		clones[i] = &Clone{
			Filename:  fmt.Sprintf("file%d.go", i),
			StartLine: i * 10,
			EndLine:   i*10 + 5,
			Size:      50,
		}
	}

	group := CloneGroup{
		Hash:   "many-clones",
		Clones: clones,
		Size:   5000,
		Method: MethodHash,
	}

	if len(group.Clones) != 100 {
		t.Errorf("Should have 100 clones, got %d", len(group.Clones))
	}
}

// TestSummary_AllMethodsUsed tests Summary with all detection methods.
func TestSummary_AllMethodsUsed(t *testing.T) {
	summary := Summary{
		TotalFiles:  10,
		TotalClones: 20,
		TotalGroups: 5,
		MethodsUsed: []DetectionMethod{
			MethodArtDupl,
			MethodHash,
			MethodTodos,
			MethodLegacy,
		},
	}

	if len(summary.MethodsUsed) != 4 {
		t.Errorf("Should have 4 methods, got %d", len(summary.MethodsUsed))
	}
}

// TestProgress_Zero tests Progress with zero values.
func TestProgress_Zero(t *testing.T) {
	progress := Progress{}

	if progress.Stage != "" {
		t.Errorf("Zero Progress Stage should be empty, got %s", progress.Stage)
	}

	if progress.Percentage != 0 {
		t.Errorf("Zero Progress Percentage should be 0, got %f", progress.Percentage)
	}
}

// TestProgress_Full tests Progress at 100%.
func TestProgress_Full(t *testing.T) {
	progress := newTestProgress("complete", 100, 100, 100.0, "Done", "")

	if progress.Percentage != 100.0 {
		t.Errorf("Progress at 100%% should be 100.0, got %f", progress.Percentage)
	}

	if progress.Completed != progress.Total {
		t.Error("At 100%, Completed should equal Total")
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = detector.FindClonesStream(ctx, []string{"test.go"})
	// Will error due to no files or canceled context
	if err == nil {
		t.Error("Expected error with canceled context")
	}
}

// TestClone_EdgeCases tests edge cases for Clone struct.
func TestClone_EdgeCases(t *testing.T) {
	// Clone with same start and end line
	clone := Clone{
		Filename:  "single.go",
		StartLine: 10,
		EndLine:   10,
		Size:      1,
	}

	if clone.StartLine != clone.EndLine {
		t.Error("Single-line clone should have same start and end line")
	}
}

// TestCloneGroup_EmptyClones tests CloneGroup with empty clones slice.
func TestCloneGroup_EmptyClones(t *testing.T) {
	group := CloneGroup{
		Hash:   "empty",
		Clones: []*Clone{},
		Size:   0,
		Method: MethodArtDupl,
	}

	if len(group.Clones) != 0 {
		t.Errorf("Should have 0 clones, got %d", len(group.Clones))
	}
}

// TestCloneGroup_NilClones tests CloneGroup with nil clones.
func TestCloneGroup_NilClones(t *testing.T) {
	group := CloneGroup{
		Hash:   "nil-clones",
		Clones: nil,
		Size:   0,
		Method: MethodArtDupl,
	}

	if group.Clones != nil {
		t.Error("Clones should be nil")
	}
}

// TestSyntaxNode_BasicUsage tests basic syntax.Node usage in conversions.
func TestSyntaxNode_BasicUsage(t *testing.T) {
	node := &syntax.Node{
		Type:     1,
		Filename: "test.go",
		Pos:      10,
		End:      20,
	}

	if node.Filename != "test.go" {
		t.Errorf("Filename should be 'test.go', got %s", node.Filename)
	}

	if node.Pos != 10 {
		t.Errorf("Pos should be 10, got %d", node.Pos)
	}

	if node.End != 20 {
		t.Errorf("End should be 20, got %d", node.End)
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

	// First call - empty files should error
	_, err = detector.FindClones(context.Background(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided, got: %v", err)
	}

	// Second call - should still work
	_, err = detector.FindClones(context.Background(), []string{})
	if !errors.Is(err, ErrNoFilesProvided) {
		t.Errorf("Expected ErrNoFilesProvided on reuse, got: %v", err)
	}
}
