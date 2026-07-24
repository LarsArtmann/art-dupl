package job

import (
	"runtime"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// testContentSimple is a reusable test Go file content to avoid goconst warnings.
const testContentSimple = `package main

func main() {
	println("hello")
}`

func TestParseParallel(t *testing.T) {
	ctx := t.Context()
	setup := testutil.NewTestFileSetup(t)

	files := map[string]string{
		"file1.go": `package main

func function1() {
	println("test1")
}`,
		"file2.go": `package main

func function2() {
	println("test2")
}`,
		"file3.go": `package main

func function3() {
	println("test3")
}`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	fchan := make(chan string, 4)
	fchan <- setup.GetFilePath("file1.go")

	fchan <- setup.GetFilePath("file2.go")

	fchan <- setup.GetFilePath("file3.go")

	close(fchan)

	schan, statsChan := ParseParallel(ctx, fchan, 2, golang.DetectionModeSemantic, 0, nil)

	count := 0

	for seq := range schan {
		if len(seq) == 0 {
			t.Error("Expected parsed nodes for valid Go file")
		}

		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 sequences, got %d", count)
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.FilesCount, 3, "FilesCount")
}

func TestParseParallelWithDefaultWorkers(t *testing.T) {
	ctx := t.Context()
	setup := testutil.NewTestFileSetup(t)

	err := setup.CreateTestFile("test.go", testContentSimple)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, _ := ParseParallel(ctx, fchan, 0, golang.DetectionModeSemantic, 0, nil)

	waitForParsedNodes(t, schan, "ParseParallel timed out")
}

func TestParseParallelContextCancellation(t *testing.T) {
	cancel, fchan := cancelledContextAndChannel(t.Context())

	schan, _ := ParseParallel(t.Context(), fchan, 2, golang.DetectionModeSemantic, 0, nil)

	// Should complete without deadlock
	for range schan {
	}

	_ = cancel
}

func TestParseParallelErrorHandling(t *testing.T) {
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- "nonexistent_file.go"

	close(fchan)

	schan, _ := ParseParallel(ctx, fchan, 1, golang.DetectionModeSemantic, 0, nil)

	waitForChannelOrTimeout(
		t,
		schan,
		5*time.Second,
		"ParseParallel should handle errors gracefully",
	)
}

func TestNormalizeWorkerCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "negative value uses GOMAXPROCS",
			input:    -1,
			expected: runtime.GOMAXPROCS(0),
		},
		{
			name:     "zero uses GOMAXPROCS",
			input:    0,
			expected: runtime.GOMAXPROCS(0),
		},
		{
			name:     "positive value unchanged",
			input:    4,
			expected: 4,
		},
		{
			name:     "one stays one",
			input:    1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workerCount := normalizeWorkerCount(tt.input)
			if workerCount != tt.expected {
				t.Errorf(
					"normalizeWorkerCount(%d) = %d, want %d",
					tt.input,
					workerCount,
					tt.expected,
				)
			}
		})
	}
}

func TestParseFileByExtensionWithConfig_GoFile(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)

	err := setup.CreateTestFile("test.go", testContentSimple)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ast, lines, err := ParseFileByExtensionWithConfig(setup.GetFilePath("test.go"), golang.DetectionModeSemantic, nil)
	if err != nil {
		t.Fatalf("ParseFileByExtensionWithConfig failed: %v", err)
	}

	if ast == nil {
		t.Error("Expected non-nil AST")
	}

	if lines < 4 {
		t.Errorf("Expected at least 4 lines, got %d", lines)
	}
}

func TestParseFileByExtensionWithConfig_NonexistentFile(t *testing.T) {
	_, _, err := ParseFileByExtensionWithConfig("nonexistent.go", golang.DetectionModeSemantic, nil)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseStatsFromSequential(t *testing.T) {
	ctx := t.Context()
	setup := testutil.NewTestFileSetup(t)

	err := setup.CreateTestFile("test.go", testContentSimple)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, statsChan := Parse(ctx, fchan, golang.DetectionModeSemantic, 0, nil)

	// Drain the sequences channel
	for range schan {
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.FilesCount, 1, "FilesCount")

	if stats.LinesCount < 4 {
		t.Errorf("Expected LinesCount >= 4, got %d", stats.LinesCount)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected int
	}{
		{
			name:     "empty content",
			content:  []byte(""),
			expected: 0,
		},
		{
			name:     "single line no newline",
			content:  []byte("hello"),
			expected: 1,
		},
		{
			name:     "single line with newline",
			content:  []byte("hello\n"),
			expected: 1,
		},
		{
			name:     "multiple lines",
			content:  []byte("line1\nline2\nline3\n"),
			expected: 3,
		},
		{
			name:     "multiple lines no trailing newline",
			content:  []byte("line1\nline2\nline3"),
			expected: 3,
		},
		{
			name:     "many newlines",
			content:  []byte("\n\n\n"),
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := countLines(tt.content)
			testutil.ExpectTrue(t, actual == tt.expected, "countLines")
		})
	}
}
