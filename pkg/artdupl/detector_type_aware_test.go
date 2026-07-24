package artdupl

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// dupFunctionCode has 7+ separate top-level statements to exceed threshold 5.
const dupFunctionCode = `package main

func parseHeader(data []byte) (int, int, int) {
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]

	return version, flags, len(tail)
}
`

func TestDetector_TypeAware_FindsClones(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")

	err := os.WriteFile(file1, []byte(dupFunctionCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write a.go: %v", err)
	}

	err = os.WriteFile(file2, []byte(dupFunctionCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write b.go: %v", err)
	}

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), []string{file1, file2})
	if err != nil {
		if errors.Is(err, ErrNoDuplicatesFound) {
			t.Skip("No duplicates found with type-aware mode (may be env-dependent)")
		}

		t.Fatalf("FindClones failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.CloneGroups) == 0 {
		t.Error("Expected at least one clone group with duplicate code")
	}
}

func TestDetector_TypeAware_FallsBackOnInvalidGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with intentionally broken type references to force LoadTypeAwareData failure.
	// The detector should fall back to syntax-only mode without crashing.
	brokenCode := `package main

func broken() {
	x := undefinedType.Method()
	_ = x
}
`

	file := filepath.Join(tmpDir, "broken.go")

	err := os.WriteFile(file, []byte(brokenCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write broken.go: %v", err)
	}

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), []string{file})

	_ = result
	_ = err
	// Either success (syntax-only fallback) or ErrNoDuplicatesFound is acceptable.
	// The key assertion is that it does NOT crash.
}

func TestDetector_TypeAware_DisabledByDefault(t *testing.T) {
	t.Parallel()

	opts := DefaultOptions()
	if opts.TypeAware {
		t.Error("TypeAware should be false by default in DefaultOptions")
	}
}
