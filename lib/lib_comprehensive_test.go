package lib

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSendFilesToChannel tests the sendFilesToChannel function.
func TestSendFilesToChannel(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "empty slice",
			files:    []string{},
			expected: []string{},
		},
		{
			name:     "single file",
			files:    []string{"file1.go"},
			expected: []string{"file1.go"},
		},
		{
			name:     "multiple files",
			files:    []string{"file1.go", "file2.go", "file3.go"},
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fchan := make(chan string, 1024)
			sendFilesToChannel(tt.files, fchan)

			var received []string
			for f := range fchan {
				received = append(received, f)
			}

			if len(received) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(received))
			}

			// Order should match
			for i, f := range received {
				if i < len(tt.expected) && f != tt.expected[i] {
					t.Errorf("file[%d] = %q, want %q", i, f, tt.expected[i])
				}
			}
		})
	}
}

// TestSendFilesToChannelClosesChannel verifies the channel is properly closed.
func TestSendFilesToChannelClosesChannel(t *testing.T) {
	fchan := make(chan string, 1024)
	sendFilesToChannel([]string{"a.go", "b.go"}, fchan)

	// Read all values - should complete when channel closes
	count := 0
	for range fchan {
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 files, got %d", count)
	}
}

// TestRunWithEmptyFiles tests Run with empty file list.
func TestRunWithEmptyFiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	issues, err := Run(ctx, []string{}, 15)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected no issues for empty files, got %d", len(issues))
	}
}

// TestRunWithNonExistentFile tests Run with a non-existent file.
func TestRunWithNonExistentFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	issues, err := Run(ctx, []string{"/nonexistent/file.go"}, 15)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Non-existent files should result in no issues
	if len(issues) != 0 {
		t.Errorf("expected no issues for non-existent file, got %d", len(issues))
	}
}

// TestRunWithSimpleFile tests Run with a simple Go file.
func TestRunWithSimpleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "simple.go")

	content := `package simple

func add(a, b int) int {
	return a + b
}
`
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	issues, err := Run(ctx, []string{filePath}, 15)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Single file with no duplicates should result in no issues
	if len(issues) != 0 {
		t.Errorf("expected no issues for single file, got %d", len(issues))
	}
}

// TestRunWithDuplicates tests Run with duplicate code.
func TestRunWithDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with identical functions
	duplicateCode := `package test

func duplicateFunction(x, y int) int {
	result := x + y
	result = result * 2
	result = result + 1
	return result
}
`

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	err := os.WriteFile(file1, []byte(duplicateCode), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file2, []byte(duplicateCode), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	issues, err := Run(ctx, []string{file1, file2}, 5) // Low threshold to catch duplicates
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should detect the duplicate function
	if len(issues) == 0 {
		t.Error("expected issues for duplicate code, got none")
	}
}

// TestRunWithContextCancellation tests Run with context cancellation.
func TestRunWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	issues, err := Run(ctx, []string{}, 15)
	// Should handle cancellation gracefully
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}

	// Empty or nil issues is acceptable
	_ = issues
}

// TestRunIncrementalWithEmptyFiles tests RunIncremental with empty file list.
func TestRunIncrementalWithEmptyFiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	issues, stats, err := RunIncremental(ctx, []string{}, 15, "", true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}

	// Stats should be returned even for empty input
	_ = stats
}

// TestRunIncrementalWithNonExistentFile tests RunIncremental with non-existent file.
func TestRunIncrementalWithNonExistentFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	issues, _, err := RunIncremental(ctx, []string{"/nonexistent/path.go"}, 15, "", true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Non-existent files should result in no issues
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

// TestRunIncrementalWithSimpleFile tests RunIncremental with a simple file.
func TestRunIncrementalWithSimpleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "simple.go")

	content := `package simple

func process(data string) string {
	return data + "processed"
}
`
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cacheDir := filepath.Join(tmpDir, ".cache")

	issues, stats, err := RunIncremental(ctx, []string{filePath}, 15, cacheDir, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Single file with no duplicates
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}

	_ = stats
}

// TestRunIncrementalWithCache tests RunIncremental caching behavior.
func TestRunIncrementalWithCache(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "cached.go")

	content := `package cached

func cachedFunc() int {
	return 42
}
`
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cacheDir := filepath.Join(tmpDir, ".cache")

	// First run - should parse and cache
	issues1, stats1, err := RunIncremental(ctx, []string{filePath}, 15, cacheDir, false)
	if err != nil {
		t.Errorf("first run error: %v", err)
	}

	// Second run - should use cache
	issues2, stats2, err := RunIncremental(ctx, []string{filePath}, 15, cacheDir, false)
	if err != nil {
		t.Errorf("second run error: %v", err)
	}

	// Both runs should produce same results
	if len(issues1) != len(issues2) {
		t.Errorf("issue count mismatch: first=%d, second=%d", len(issues1), len(issues2))
	}

	// Stats should indicate cache usage on second run
	_ = stats1
	_ = stats2
}

// TestRunIncrementalWithClearCache tests RunIncremental with cache clearing.
func TestRunIncrementalWithClearCache(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "clear.go")

	content := `package clear

func clearFunc() bool {
	return true
}
`
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cacheDir := filepath.Join(tmpDir, ".cache")

	// First run - creates cache
	_, _, err = RunIncremental(ctx, []string{filePath}, 15, cacheDir, false)
	if err != nil {
		t.Errorf("first run error: %v", err)
	}

	// Second run with clearCache=true
	_, _, err = RunIncremental(ctx, []string{filePath}, 15, cacheDir, true)
	if err != nil {
		t.Errorf("second run with clearCache error: %v", err)
	}
}

// TestDefaultCacheDir tests that the re-exported constant is accessible.
func TestDefaultCacheDir(t *testing.T) {
	if DefaultCacheDir == "" {
		t.Error("DefaultCacheDir should not be empty")
	}
}
