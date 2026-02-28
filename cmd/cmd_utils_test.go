package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

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
			name: "multiple methods",
			methods: config.DetectionMethods{
				config.DetectionMethodArtDupl,
				config.DetectionMethodHash,
			},
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
		createTestCase("duplicate entries same position", [][]*syntax.Node{
			createTestNodes("test.go", 1, 10),
			createTestNodes("test.go", 1, 10),
		}, 1),
		createTestCase("different positions", [][]*syntax.Node{
			createTestNodes("test1.go", 1, 10),
			createTestNodes("test2.go", 1, 10),
		}, 2),
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
	origVersion := Version
	origCommit := Commit
	origDate := Date

	t.Cleanup(func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	})

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

		err := writeFormatFile(
			context.TODO(),
			cfg,
			matches,
			parseStats,
			format,
			filename,
			sortByEnum,
			"art-dupl",
		)
		if err != nil {
			t.Fatalf("writeFormatFile() error = %v", err)
		}
	})
}

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
