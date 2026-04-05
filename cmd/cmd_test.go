package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/spf13/cobra"
)

// createTestNodes creates a test node slice with the specified filename and position.
func createTestNodes(filename string, pos, end int32) []*syntax.Node {
	return []*syntax.Node{
		{Filename: filename, Pos: pos, End: end},
	}
}

func testFilesFeedWithExtension(t *testing.T, tmpDir, ext string, expectedCount int) {
	t.Helper()

	ch := filesFeedWithOptions([]string{tmpDir}, false, nil, false, ext)

	found := make([]string, 0, expectedCount)
	for f := range ch {
		found = append(found, filepath.Base(f))
	}

	if len(found) != expectedCount {
		t.Errorf("Expected %d %s files, got %d: %v", expectedCount, ext, len(found), found)
	}

	for _, f := range found {
		if !strings.HasSuffix(f, "."+ext) {
			t.Errorf("Expected only .%s files, found: %s", ext, f)
		}
	}
}

const testDuplicateCode = `package test

func DuplicateFunction() int {
	x := 1
	y := 2
	z := x + y
	return z * 2
}
`

// createDuplicateTestFiles creates two test files with duplicate code in the given directory.
func createDuplicateTestFiles(t *testing.T, tmpDir string) {
	t.Helper()

	duplicateCode := testDuplicateCode
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	err := os.WriteFile(file1, []byte(duplicateCode), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(file2, []byte(duplicateCode), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

// runOutputFormatTest runs a test for a specific output format flag.
func runOutputFormatTest(t *testing.T, formatFlag string) {
	t.Helper()

	tmpDir := t.TempDir()

	createDuplicateTestFiles(t, tmpDir)

	cmd := NewRootCommand()
	AddFlags(cmd)
	cmd.SetArgs([]string{formatFlag, "--threshold", "10", tmpDir})

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("runCmd() error = %v", err)
	}
}

// runStatsFormatTest runs a stats test for a specific format.
func runStatsFormatTest(t *testing.T, format string) {
	t.Helper()

	tmpDir := t.TempDir()

	createDuplicateTestFiles(t, tmpDir)

	cmd := NewStatsCommand()
	cmd.SetArgs([]string{"--format", format, "--threshold", "10", tmpDir})

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("runStats() error = %v", err)
	}
}

// createTestCase creates a test case for the unique function.
func createTestCase(name string, input [][]*syntax.Node, expected int) struct {
	name     string
	input    [][]*syntax.Node
	expected int
} {
	return struct {
		name     string
		input    [][]*syntax.Node
		expected int
	}{
		name:     name,
		input:    input,
		expected: expected,
	}
}

// Printer interface mock for testing.
type mockPrinter struct {
	headerCalled bool
	footerCalled bool
	clonesCalled bool
}

func (m *mockPrinter) PrintHeader() error {
	m.headerCalled = true

	return nil
}

func (m *mockPrinter) PrintFooter() error {
	m.footerCalled = true

	return nil
}

func (m *mockPrinter) PrintClones(_ [][]*syntax.Node, _ ...printer.SortBy) error {
	m.clonesCalled = true

	return nil
}

var _ printer.Printer = (*mockPrinter)(nil)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	if cmd.Use != "art-dupl [flags] [paths...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "art-dupl [flags] [paths...]")
	}

	if cmd.Short != "Find code clones" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Find code clones")
	}

	if cmd.Commands() == nil {
		t.Error("Expected commands to be registered")
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

func TestNewStatsCommand(t *testing.T) {
	cmd := NewStatsCommand()

	if cmd == nil {
		t.Fatal("NewStatsCommand() returned nil")
	}

	if cmd.Use != "stats [flags] [paths...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "stats [flags] [paths...]")
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

func TestAddFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlags(cmd)

	flags := []string{
		"config",
		"vendor",
		"verbose",
		"threshold",
		"files",
		"html",
		"json",
		"plumbing",
		"sort",
		"detection-methods",
		"all",
		"output-dir",
		"filter-generated",
		"include-sqlc",
		"include-templ",
		"include-pattern",
		"exclude-pattern",
		"incremental",
		"since",
		"cache-dir",
		"clear-cache",
	}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be added", flag)
		}
	}

	threshold, _ := cmd.Flags().GetInt("threshold")
	if threshold != 15 {
		t.Errorf("threshold default = %d, want 15", threshold)
	}
}

func TestPrintVersion(t *testing.T) {
	origVersion := Version
	origCommit := Commit
	origDate := Date

	t.Cleanup(func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	})

	Version = "1.0.0"
	Commit = "abc123"
	Date = "2024-01-01"

	PrintVersion()
}

func TestCrawlPaths(t *testing.T) {
	t.Run("empty paths", func(t *testing.T) {
		result := crawlPaths([]string{}, nil, false)
		if result == nil {
			t.Fatal("crawlPaths() returned nil")
		}

		count := 0
		for range result {
			count++
		}

		if count != 0 {
			t.Errorf("crawlPaths() = %d paths, want 0", count)
		}
	})

	t.Run("single file path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tmpFile := filepath.Join(tmpDir, "test.go")

		err := os.WriteFile(tmpFile, []byte("package main"), 0o600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := crawlPaths([]string{tmpFile}, nil, false)
		if result == nil {
			t.Fatal("crawlPaths() returned nil")
		}

		count := 0
		for range result {
			count++
		}

		if count != 1 {
			t.Errorf("crawlPaths() = %d paths, want 1", count)
		}
	})
}

func TestFilesFeedWithOptions(t *testing.T) {
	t.Run("empty options", func(t *testing.T) {
		ch := filesFeedWithOptions([]string{}, false, (*filter.Filter)(nil), false, "")
		if ch == nil {
			t.Fatal("filesFeedWithOptions() returned nil")
		}

		for range ch {
			// Drain channel
		}
	})

	t.Run("with filter", func(t *testing.T) {
		f := filter.NewFilter(false, nil)

		ch := filesFeedWithOptions([]string{}, false, f, false, "")
		if ch == nil {
			t.Fatal("filesFeedWithOptions() returned nil")
		}

		for range ch {
			// Drain channel
		}
	})
}

func TestFilesFeedWithOptions_OnlyFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{"test1.go", "test2.go", "test1.templ", "test2.templ"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)

		err := os.WriteFile(path, []byte("content"), 0o600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("only go files", func(t *testing.T) {
		testFilesFeedWithExtension(t, tmpDir, "go", 2)
	})

	t.Run("only templ files", func(t *testing.T) {
		testFilesFeedWithExtension(t, tmpDir, "templ", 2)
	})

	t.Run("all files with empty filter", func(t *testing.T) {
		ch := filesFeedWithOptions([]string{tmpDir}, false, nil, false, "")

		found := make([]string, 0, 4)
		for f := range ch {
			found = append(found, filepath.Base(f))
		}

		if len(found) != 4 {
			t.Errorf("Expected 4 files, got %d: %v", len(found), found)
		}
	})
}

func TestPrintDupls(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		mock := &mockPrinter{}
		ch := make(chan syntax.Match)
		close(ch)

		err := printDupls(t.Context(), mock, ch, printer.SortBySize, 15, "art-dupl")
		if err != nil {
			t.Errorf("printDupls() error = %v", err)
		}

		if !mock.headerCalled {
			t.Error("Expected PrintHeader to be called")
		}

		if !mock.footerCalled {
			t.Error("Expected PrintFooter to be called")
		}
	})

	t.Run("with matches", func(t *testing.T) {
		mock := &mockPrinter{}

		ch := make(chan syntax.Match, 1)
		ch <- syntax.Match{
			Hash: "abc123",
			Frags: [][]*syntax.Node{
				createTestNodes("test1.go", 1, 10),
				createTestNodes("test2.go", 1, 10),
			},
		}

		close(ch)

		err := printDupls(t.Context(), mock, ch, printer.SortBySize, 15, "art-dupl")
		if err != nil {
			t.Errorf("printDupls() error = %v", err)
		}

		if !mock.clonesCalled {
			t.Error("Expected PrintClones to be called")
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		mock := &mockPrinter{}

		ch := make(chan syntax.Match, 1)
		ch <- syntax.Match{
			Hash: "abc123",
			Frags: [][]*syntax.Node{
				createTestNodes("test1.go", 1, 10),
				createTestNodes("test2.go", 1, 10),
			},
		}

		close(ch)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := printDupls(ctx, mock, ch, printer.SortBySize, 15, "art-dupl")
		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}

		if mock.footerCalled {
			t.Error("Expected PrintFooter NOT to be called on cancellation")
		}
	})
}

func TestSetupFilter(t *testing.T) {
	t.Run("empty config returns filter", func(t *testing.T) {
		cfg := &config.Config{}

		f := setupFilter(cfg)
		if f == nil {
			t.Error("setupFilter() returned nil")
		}
	})

	t.Run("with filter generated", func(t *testing.T) {
		cfg := &config.Config{FilterGenerated: true}

		f := setupFilter(cfg)
		if f == nil {
			t.Error("setupFilter() returned nil")
		}
	})

	t.Run("with include sqlc", func(t *testing.T) {
		cfg := &config.Config{IncludeSQLC: true}

		f := setupFilter(cfg)
		if f == nil {
			t.Error("setupFilter() returned nil")
		}
	})

	t.Run("with include templ", func(t *testing.T) {
		cfg := &config.Config{IncludeTempl: true}

		f := setupFilter(cfg)
		if f == nil {
			t.Error("setupFilter() returned nil")
		}
	})
}
