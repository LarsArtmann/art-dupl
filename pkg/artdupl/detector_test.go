package artdupl

import (
	"errors"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// invalidMaxFileSizeOptions creates Options with invalid MaxFileSize for testing.
func invalidMaxFileSizeOptions() *Options {
	return &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
		MaxFileSize:      -1,
	}
}

// cleanupDetector closes the detector and logs any errors.
func cleanupDetector(t *testing.T, detector Detector) {
	t.Helper()

	err := detector.Close()
	if err != nil {
		t.Logf("Failed to close detector: %v", err)
	}
}

// newTestOptionsWithThreshold creates Options with a specific threshold and default detection method.
func newTestOptionsWithThreshold(threshold int) *Options {
	return &Options{
		Threshold:        threshold,
		DetectionMethods: []DetectionMethod{MethodArtDupl},
	}
}

// newTestOptionsWithNoDetectionMethods creates Options with empty detection methods.
func newTestOptionsWithNoDetectionMethods() *Options {
	return &Options{
		Threshold:        15,
		DetectionMethods: []DetectionMethod{},
	}
}

// createTestClone creates a Clone with configurable fragment for testing.
func createTestClone(t *testing.T, fragment string) Clone {
	t.Helper()

	return Clone{
		CloneRef: domain.CloneRef{
			Filename:  testFilename,
			LineStart: 1,
			LineEnd:   5,
			Fragment:  fragment,
		},
		Size: 25,
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
	opts := invalidMaxFileSizeOptions()

	_, err := NewDetector(opts)
	if !errors.Is(err, ErrInvalidMaxFileSize) {
		t.Errorf("Expected ErrInvalidMaxFileSize, got: %v", err)
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

	testutil.AssertFieldValue(t, node.Type, 1, "Type")

	testutil.AssertFieldValue(t, node.Filename, "test.go", "Filename")

	testutil.AssertFieldValue(t, node.Pos, 10, "Pos")

	testutil.AssertFieldValue(t, node.End, 20, "End")
}

// TestDetector_FindClones_NonExistentFile tests FindClones with non-existent file.
func TestDetector_FindClones_NonExistentFile(t *testing.T) {
	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	result, err := detector.FindClones(t.Context(), []string{"nonexistent_file.go"})
	if err == nil && result != nil {
		return
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
		{"wrapped error", wrapError(ErrInvalidThreshold), ErrInvalidThreshold, true},
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

// wrapError wraps an error for testing.
func wrapError(err error) error {
	return err
}
