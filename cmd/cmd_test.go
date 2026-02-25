package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/spf13/cobra"
)

const testDuplicateCode = `package test

func DuplicateFunction() int {
	x := 1
	y := 2
	z := x + y
	return z * 2
}
`

// runOutputFormatTest runs a test for a specific output format flag.
func runOutputFormatTest(t *testing.T, formatFlag string) {
	t.Helper()

	tmpDir := t.TempDir()

	duplicateCode := testDuplicateCode
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

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

	duplicateCode := testDuplicateCode
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

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

func TestDetectionMethodsToString(t *testing.T) {
	tests := []struct {
		name     string
		methods  config.DetectionMethods
		expected string
	}{
		{
			name:     "empty methods",
			methods:  config.DetectionMethods{},
			expected: "",
		},
		{
			name:     "single method",
			methods:  config.DetectionMethods{config.DetectionMethodArtDupl},
			expected: "art-dupl",
		},
		{
			name:     "multiple methods",
			methods:  config.DetectionMethods{config.DetectionMethodArtDupl, config.DetectionMethodHash},
			expected: "art-dupl,hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectionMethodsToString(tt.methods)
			if result != tt.expected {
				t.Errorf("detectionMethodsToString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "go file",
			filename: "main.go",
			expected: true,
		},
		{
			name:     "templ file",
			filename: "view.templ",
			expected: true,
		},
		{
			name:     "test file",
			filename: "main_test.go",
			expected: true,
		},
		{
			name:     "txt file",
			filename: "readme.txt",
			expected: false,
		},
		{
			name:     "md file",
			filename: "README.md",
			expected: false,
		},
		{
			name:     "json file",
			filename: "config.json",
			expected: false,
		},
		{
			name:     "empty string",
			filename: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSourceFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isSourceFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]*syntax.Node
		expected int
	}{
		{
			name:     "empty input",
			input:    nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			input:    [][]*syntax.Node{},
			expected: 0,
		},
		{
			name: "single duplicate group",
			input: [][]*syntax.Node{
				{
					{Filename: "test.go", Pos: 1, End: 10},
					{Filename: "test.go", Pos: 11, End: 20},
				},
			},
			expected: 1,
		},
		{
			name: "duplicate entries same position",
			input: [][]*syntax.Node{
				{
					{Filename: "test.go", Pos: 1, End: 10},
				},
				{
					{Filename: "test.go", Pos: 1, End: 10},
				},
			},
			expected: 1,
		},
		{
			name: "different positions",
			input: [][]*syntax.Node{
				{
					{Filename: "test1.go", Pos: 1, End: 10},
				},
				{
					{Filename: "test2.go", Pos: 1, End: 10},
				},
			},
			expected: 2,
		},
		{
			name: "empty inner slice",
			input: [][]*syntax.Node{
				{},
				{
					{Filename: "test.go", Pos: 1, End: 10},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := syntax.Unique(tt.input)
			if len(result) != tt.expected {
				t.Errorf("unique() returned %d groups, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestVersionFunctions(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origDate := Date
	defer func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	}()

	t.Run("GetVersion with dev", func(t *testing.T) {
		Version = "dev"
		Commit = "unknown"
		result := GetVersion()
		if result != "dev" {
			t.Errorf("GetVersion() = %q, want %q", result, "dev")
		}
	})

	t.Run("GetVersion with commit", func(t *testing.T) {
		Version = "1.0.0"
		Commit = "abcd1234efgh5678"
		result := GetVersion()
		expected := "1.0.0-abcd123"
		if result != expected {
			t.Errorf("GetVersion() = %q, want %q", result, expected)
		}
	})

	t.Run("GetCommit", func(t *testing.T) {
		Commit = "test123"
		result := GetCommit()
		if result != "test123" {
			t.Errorf("GetCommit() = %q, want %q", result, "test123")
		}
	})

	t.Run("GetBuildDate", func(t *testing.T) {
		Date = "2024-01-01"
		result := GetBuildDate()
		if result != "2024-01-01" {
			t.Errorf("GetBuildDate() = %q, want %q", result, "2024-01-01")
		}
	})
}

func TestCreatePrinter(t *testing.T) {
	tests := []struct {
		name   string
		format config.OutputFormat
	}{
		{
			name:   "text format",
			format: config.OutputFormatText,
		},
		{
			name:   "html format",
			format: config.OutputFormatHTML,
		},
		{
			name:   "json format",
			format: config.OutputFormatJSON,
		},
		{
			name:   "plumbing format",
			format: config.OutputFormatPlumbing,
		},
		{
			name:   "unknown format defaults to text",
			format: config.OutputFormat("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			readFile := func(filename string) ([]byte, error) { return nil, nil }

			fn := createPrinter(tt.format, 15)
			p := fn(&buf, readFile)

			if p == nil {
				t.Error("createPrinter() returned nil printer")
			}
		})
	}
}

func TestCollectMatches(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		ch := make(chan syntax.Match)
		close(ch)
		result := collectMatches(ch)
		if len(result) != 0 {
			t.Errorf("collectMatches() returned %d matches, want 0", len(result))
		}
	})

	t.Run("single match", func(t *testing.T) {
		ch := make(chan syntax.Match, 1)
		ch <- syntax.Match{
			Hash:  "abc123",
			Frags: [][]*syntax.Node{{{Filename: "test.go", Pos: 1, End: 10}}},
		}
		close(ch)
		result := collectMatches(ch)
		if len(result) != 1 {
			t.Errorf("collectMatches() returned %d matches, want 1", len(result))
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		ch := make(chan syntax.Match, 3)
		ch <- syntax.Match{Hash: "hash1", Frags: [][]*syntax.Node{{{Filename: "test1.go"}}}}
		ch <- syntax.Match{Hash: "hash2", Frags: [][]*syntax.Node{{{Filename: "test2.go"}}}}
		ch <- syntax.Match{Hash: "hash3", Frags: [][]*syntax.Node{{{Filename: "test3.go"}}}}
		close(ch)
		result := collectMatches(ch)
		if len(result) != 3 {
			t.Errorf("collectMatches() returned %d matches, want 3", len(result))
		}
	})
}

func TestWriteFormatFile(t *testing.T) {
	t.Run("creates file and writes content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := tmpDir + "/test.txt"

		cfg := &config.Config{Threshold: 15}
		matches := []syntax.Match{}
		parseStats := job.ParseStats{FilesCount: 10, LinesCount: 100}
		format := config.OutputFormatText
		sortByEnum := printer.SortBySize

		err := writeFormatFile(context.TODO(), cfg, matches, parseStats, format, filename, sortByEnum, "art-dupl")
		if err != nil {
			t.Fatalf("writeFormatFile() error = %v", err)
		}
	})
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

func TestShouldIncludeFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "nil filter includes all",
			path:     "test.go",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeFile(nil, tt.path)
			if result != tt.expected {
				t.Errorf("shouldIncludeFile(nil, %q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "valid minutes",
			input:    "30m",
			hasError: false,
		},
		{
			name:     "valid hours",
			input:    "1h",
			hasError: false,
		},
		{
			name:     "valid combined",
			input:    "1h30m",
			hasError: false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDuration(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("parseDuration(%q) error = %v, hasError %v", tt.input, err, tt.hasError)
			}
		})
	}
}

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

	// Check that stats subcommand exists
	if cmd.Commands() == nil {
		t.Error("Expected commands to be registered")
	}

	// Check that RunE is set
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

	// Check that RunE is set
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

func TestAddFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFlags(cmd)

	// Test that flags were added
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

	// Test default threshold value
	threshold, _ := cmd.Flags().GetInt("threshold")
	if threshold != 15 {
		t.Errorf("threshold default = %d, want 15", threshold)
	}
}

func TestPrintVersion(t *testing.T) {
	// Save and restore original values
	origVersion := Version
	origCommit := Commit
	origDate := Date
	defer func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	}()

	Version = "1.0.0"
	Commit = "abc123"
	Date = "2024-01-01"

	// Just verify it doesn't panic
	PrintVersion()
}

func TestCrawlPaths(t *testing.T) {
	t.Run("empty paths", func(t *testing.T) {
		result := crawlPaths([]string{}, nil, false)
		if result == nil {
			t.Fatal("crawlPaths() returned nil")
		}
		// Drain the channel
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
		if err := os.WriteFile(tmpFile, []byte("package main"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := crawlPaths([]string{tmpFile}, nil, false)
		if result == nil {
			t.Fatal("crawlPaths() returned nil")
		}
		// Drain the channel
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
		ch := filesFeedWithOptions([]string{}, false, (*filter.Filter)(nil), false)
		if ch == nil {
			t.Fatal("filesFeedWithOptions() returned nil")
		}
		// Drain the channel
		for range ch {
		}
	})

	t.Run("with filter", func(t *testing.T) {
		f := filter.NewFilter(false, nil)
		ch := filesFeedWithOptions([]string{}, false, f, false)
		if ch == nil {
			t.Fatal("filesFeedWithOptions() returned nil")
		}
		// Drain the channel
		for range ch {
		}
	})
}

func TestPrintDupls(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		mock := &mockPrinter{}
		ch := make(chan syntax.Match)
		close(ch)

		err := printDupls(mock, ch, printer.SortBySize, 15, "art-dupl")
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
				{{Filename: "test1.go", Pos: 1, End: 10}},
				{{Filename: "test2.go", Pos: 1, End: 10}},
			},
		}
		close(ch)

		err := printDupls(mock, ch, printer.SortBySize, 15, "art-dupl")
		if err != nil {
			t.Errorf("printDupls() error = %v", err)
		}
		if !mock.clonesCalled {
			t.Error("Expected PrintClones to be called")
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

// --- Integration Tests for Command Handlers ---

func TestRunCmd_Integration(t *testing.T) {
	t.Run("basic execution with duplicate code", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files with intentional duplicates
		duplicateCode := testDuplicateCode
		file1 := filepath.Join(tmpDir, "file1.go")
		file2 := filepath.Join(tmpDir, "file2.go")
		if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})

	t.Run("invalid sort option returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--sort", "invalid", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runCmd() expected error for invalid sort option")
		}
	})

	t.Run("invalid detection methods returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--detection-methods", "invalid-method", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runCmd() expected error for invalid detection methods")
		}
	})

	t.Run("json output format", func(t *testing.T) {
		runOutputFormatTest(t, "--json")
	})

	t.Run("plumbing output format", func(t *testing.T) {
		runOutputFormatTest(t, "--plumbing")
	})

	t.Run("html output format", func(t *testing.T) {
		runOutputFormatTest(t, "--html")
	})

	t.Run("verbose output", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--verbose", "--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})

	t.Run("with vendor flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		vendorDir := filepath.Join(tmpDir, "vendor")
		if err := os.MkdirAll(vendorDir, 0o750); err != nil {
			t.Fatalf("Failed to create vendor dir: %v", err)
		}
		file1 := filepath.Join(vendorDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--vendor", "--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})
}

func TestRunStats_Integration(t *testing.T) {
	t.Run("basic stats execution", func(t *testing.T) {
		tmpDir := t.TempDir()

		duplicateCode := testDuplicateCode
		file1 := filepath.Join(tmpDir, "file1.go")
		file2 := filepath.Join(tmpDir, "file2.go")
		if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{"--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runStats() error = %v", err)
		}
	})

	t.Run("stats with json format", func(t *testing.T) {
		runStatsFormatTest(t, "json")
	})

	t.Run("stats with csv format", func(t *testing.T) {
		runStatsFormatTest(t, "csv")
	})

	t.Run("stats invalid format returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{"--format", "invalid", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runStats() expected error for invalid format")
		}
	})

	t.Run("stats with verbose", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{"--verbose", "--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runStats() error = %v", err)
		}
	})
}

func TestRunAllModes_Integration(t *testing.T) {
	t.Run("generates all output formats", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "reports")

		duplicateCode := testDuplicateCode
		file1 := filepath.Join(tmpDir, "file1.go")
		file2 := filepath.Join(tmpDir, "file2.go")
		if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--all", "--output-dir", outputDir, "--threshold", "10", tmpDir})
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runAllModes() error = %v", err)
		}

		// Check that report files were created for all output formats
		formats := []string{"text", "html", "json", "plumbing", "simple-json"}
		for _, format := range formats {
			reportFile := filepath.Join(outputDir, "report."+format)
			if _, statErr := os.Stat(reportFile); os.IsNotExist(statErr) {
				t.Errorf("Expected report file %s to be created", reportFile)
			}
		}
	})
}

func TestExecuteAnalysis_Integration(t *testing.T) {
	t.Run("basic analysis", func(t *testing.T) {
		tmpDir := t.TempDir()

		duplicateCode := testDuplicateCode
		file1 := filepath.Join(tmpDir, "file1.go")
		file2 := filepath.Join(tmpDir, "file2.go")
		if err := os.WriteFile(file1, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(file2, []byte(duplicateCode), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &config.Config{
			Threshold:        10,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
			IncludeTempl:     true,
			IncludeSQLC:      true,
		}

		ctx := context.Background()
		duplChan, parseStats, filterStats, err := executeAnalysis(ctx, cfg, []string{tmpDir}, config.OutputFormatText)
		if err != nil {
			t.Fatalf("executeAnalysis() error = %v", err)
		}

		// Drain the channel
		matchCount := 0
		for range duplChan {
			matchCount++
		}

		if parseStats.FilesCount < 2 {
			t.Errorf("Expected at least 2 files parsed, got %d", parseStats.FilesCount)
		}

		_ = filterStats // FilterStats may be empty if no filtering
		_ = matchCount
	})

	t.Run("analysis with profile enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &config.Config{
			Threshold:        10,
			Profile:          true,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
			IncludeTempl:     true,
			IncludeSQLC:      true,
		}

		ctx := context.Background()
		duplChan, _, _, err := executeAnalysis(ctx, cfg, []string{tmpDir}, config.OutputFormatText)
		if err != nil {
			t.Fatalf("executeAnalysis() error = %v", err)
		}

		// Drain the channel
		for range duplChan {
		}
	})
}

func TestBuildSuffixTree_Integration(t *testing.T) {
	t.Run("builds tree from files", func(t *testing.T) {
		tmpDir := t.TempDir()

		code := `package test

func Example() int {
	return 42
}
`
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte(code), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &config.Config{
			Threshold:        10,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
		}

		ctx := context.Background()
		tree, data, parseStats, err := buildSuffixTree(ctx, []string{tmpDir}, cfg, nil, config.OutputFormatText)
		if err != nil {
			t.Fatalf("buildSuffixTree() error = %v", err)
		}

		if tree == nil {
			t.Error("buildSuffixTree() returned nil tree")
		}

		if len(data) == 0 {
			t.Error("buildSuffixTree() returned empty data")
		}

		if parseStats.FilesCount < 1 {
			t.Errorf("Expected at least 1 file parsed, got %d", parseStats.FilesCount)
		}
	})

	t.Run("verbose output", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte("package main\n"), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &config.Config{
			Verbose:          true,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
		}

		ctx := context.Background()
		tree, _, _, err := buildSuffixTree(ctx, []string{tmpDir}, cfg, nil, config.OutputFormatText)
		if err != nil {
			t.Fatalf("buildSuffixTree() error = %v", err)
		}

		if tree == nil {
			t.Error("buildSuffixTree() returned nil tree")
		}
	})
}
