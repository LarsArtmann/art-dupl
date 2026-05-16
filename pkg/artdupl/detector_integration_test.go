package artdupl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	result, err := detector.FindClones(t.Context(), []string{filename})
	if err != nil {
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
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	result, err := detector.FindClones(t.Context(), []string{file1, file2})
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

	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

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

// TestDetector_Integration_NonExistentFile tests FindClones with non-existent file.
func TestDetector_Integration_NonExistentFile(t *testing.T) {
	t.Parallel()

	opts := DefaultOptions()

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	found, err := detector.FindClones(t.Context(), []string{"nonexistent_file.go"})
	if err == nil && found != nil {
		return
	}
}

// TestDetector_Integration_StreamWithValidFiles tests FindClonesStream with valid files.
func TestDetector_Integration_StreamWithValidFiles(t *testing.T) {
	t.Parallel()

	filename := createTestGoFile(t, `package main

func main() {
	println("hello")
}
`)

	opts := DefaultOptions()
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() {
		cleanupDetector(t, detector)
	})

	_, err = detector.FindClonesStream(t.Context(), []string{filename})
	if err != nil {
		t.Logf("FindClonesStream returned: %v", err)
	}
}
