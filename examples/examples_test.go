package examples

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

const testFile = "test.go"

// TestExamplesPackage tests examples package functionality.
func TestExamplesPackage(t *testing.T) {
	// Test default options
	opts := artdupl.DefaultOptions()

	// Test threshold
	if opts.Threshold <= 0 {
		t.Error("Threshold should be positive")
	}

	// Test detection methods
	if len(opts.DetectionMethods) == 0 {
		t.Error("Should have detection methods")
	}

	// Test timeout
	if opts.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}

	// Test max workers
	if opts.MaxWorkers <= 0 {
		t.Error("Max workers should be positive")
	}

	// Test max file size
	if opts.MaxFileSize <= 0 {
		t.Error("Max file size should be positive")
	}

	// Test max clones per group
	if opts.MaxClonesPerGroup <= 0 {
		t.Error("Max clones per group should be positive")
	}
}

// TestExamplesTypes tests type definitions.
func TestExamplesTypes(t *testing.T) {
	// Test detection method constants
	methods := []artdupl.DetectionMethod{
		artdupl.MethodArtDupl,
		artdupl.MethodHash,
	}

	for _, method := range methods {
		if string(method) == "" {
			t.Errorf("Detection method should not be empty: %v", method)
		}
	}

	// Test error variables
	errors := []error{
		artdupl.ErrNilOptions,
		artdupl.ErrInvalidThreshold,
		artdupl.ErrNoDetectionMethods,
		artdupl.ErrInvalidMaxFileSize,
		artdupl.ErrInvalidMaxWorkers,
		artdupl.ErrInvalidTimeout,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("Error variable should not be nil")
		}

		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}

	// Test result structure
	result := &artdupl.Result{
		CloneGroups: []*artdupl.CloneGroup{{}},
		Summary:     &artdupl.Summary{},
		Metadata:    &artdupl.Metadata{},
	}

	if result.CloneGroups == nil {
		t.Error("Clone groups should not be nil")
	}

	if result.Summary == nil {
		t.Error("Summary should not be nil")
	}

	if result.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// Test clone group structure
	cloneGroup := &artdupl.CloneGroup{
		Hash:   "test-hash",
		Clones: []*artdupl.Clone{{Filename: testFile}},
		Size:   10,
		Method: artdupl.MethodArtDupl,
	}

	if cloneGroup.Hash == "" {
		t.Error("Clone group should have hash")
	}

	if len(cloneGroup.Clones) == 0 {
		t.Error("Clone group should have clones")
	}

	if cloneGroup.Size <= 0 {
		t.Error("Clone group size should be positive")
	}

	if cloneGroup.Method == "" {
		t.Error("Clone group should have method")
	}

	// Test clone structure
	clone := &artdupl.Clone{
		Filename:  testFile,
		LineStart: 1,
		LineEnd:   10,
		StartPos:  0,
		EndPos:    100,
		Fragment:  "test code",
		Size:      100,
	}

	if clone.Filename == "" {
		t.Error("Clone should have filename")
	}

	if clone.LineStart <= 0 {
		t.Error("Start line should be positive")
	}

	if clone.LineEnd <= clone.LineStart {
		t.Error("End line should be after start line")
	}

	if clone.Size <= 0 {
		t.Error("Clone size should be positive")
	}

	// Test summary structure
	summary := &artdupl.Summary{
		TotalFiles:    10,
		TotalClones:   5,
		TotalGroups:   3,
		AnalysisTime:  1000,
		MethodsUsed:   []artdupl.DetectionMethod{artdupl.MethodArtDupl},
		LinesAnalyzed: 1000,
	}

	if summary.TotalFiles <= 0 {
		t.Error("Total files should be positive")
	}

	if len(summary.MethodsUsed) == 0 {
		t.Error("Methods used should not be empty")
	}

	// Test metadata structure
	metadata := &artdupl.Metadata{
		Version:    "dev",
		ConfigHash: "test-config-hash",
		Toolchain:  "go1.21.0",
	}

	if metadata.Version == "" {
		t.Error("Version should not be empty")
	}

	if metadata.ConfigHash == "" {
		t.Error("Config hash should not be empty")
	}

	if metadata.Toolchain == "" {
		t.Error("Toolchain should not be empty")
	}

	// Test progress structure
	progress := &artdupl.Progress{
		Stage:       "test",
		Completed:   50,
		Total:       100,
		Percentage:  50.0,
		Message:     "test message",
		CurrentFile: testFile,
	}

	if progress.Stage == "" {
		t.Error("Progress stage should not be empty")
	}

	if progress.Percentage < 0 || progress.Percentage > 100 {
		t.Errorf("Progress percentage should be between 0-100, got %f", progress.Percentage)
	}
}

// TestExamplesDetector tests detector functionality.
func TestExamplesDetector(t *testing.T) {
	// Test detector creation with nil options
	detector, err := artdupl.NewDetector(nil)
	if err != nil {
		t.Errorf("Failed to create detector with nil options: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}

	_ = detector.Close()

	// Test detector creation with default options
	opts := artdupl.DefaultOptions()

	detector, err = artdupl.NewDetector(opts)
	if err != nil {
		t.Errorf("Failed to create detector with default options: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}

	_ = detector.Close()

	// Test detector creation with custom options
	customOpts := &artdupl.Options{
		Threshold:         20,
		DetectionMethods:  []artdupl.DetectionMethod{artdupl.MethodArtDupl},
		IncludeVendor:     false,
		MaxFileSize:       1024 * 1024,
		MaxWorkers:        2,
		IncludeFragments:  true,
		MaxClonesPerGroup: 5,
	}

	detector, err = artdupl.NewDetector(customOpts)
	if err != nil {
		t.Errorf("Failed to create detector with custom options: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}

	_ = detector.Close()
}

// TestExamplesInterfaces tests interface implementations.
func TestExamplesInterfaces(t *testing.T) {
	// Test logger interface
	var l artdupl.Logger = &logger.NoOpLogger{}

	// Test that all methods are implemented
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warning message")
	l.Error("error message")

	// Test file reader function
	var fileReader artdupl.FileReaderFunc = func(path string) ([]byte, error) {
		return []byte("test content"), nil
	}

	data, err := fileReader("test.txt")
	if err != nil {
		t.Errorf("File reader failed: %v", err)
	}

	if string(data) != "test content" {
		t.Error("File reader returned wrong content")
	}

	// Test progress callback
	callback := func(progress *artdupl.Progress) error {
		// Test progress structure
		if progress.Stage == "" {
			t.Error("Progress stage should not be empty")
		}

		return nil
	}

	opts := artdupl.DefaultOptions()
	opts.ProgressCallback = callback

	detector, err := artdupl.NewDetector(opts)
	if err != nil {
		t.Errorf("Failed to create detector with progress callback: %v", err)
	}

	if detector == nil {
		t.Error("Detector should not be nil")
	}

	_ = detector.Close()
}
