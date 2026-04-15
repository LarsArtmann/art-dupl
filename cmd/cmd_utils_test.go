package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/gogenfilter"
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

			fn := createPrinter(tt.format, 15, config.DiffModeDisabled, printer.ReportMetadata{})
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
		testutil.AssertCount(t, len(result), 0, "collectMatches()")
	})

	t.Run("single match", func(t *testing.T) {
		ch := make(chan syntax.Match, 1)
		ch <- syntax.Match{
			Hash:  "abc123",
			Frags: [][]*syntax.Node{{{Filename: "test.go", Pos: 1, End: 10}}},
		}

		close(ch)

		result := collectMatches(ch)
		testutil.AssertCount(t, len(result), 1, "collectMatches()")
	})

	t.Run("multiple matches", func(t *testing.T) {
		ch := make(chan syntax.Match, 3)

		ch <- testutil.CreateMatch("hash1", "test1.go")

		ch <- testutil.CreateMatch("hash2", "test2.go")

		ch <- testutil.CreateMatch("hash3", "test3.go")

		close(ch)

		result := collectMatches(ch)
		testutil.AssertCount(t, len(result), 3, "collectMatches()")
	})
}

func TestWriteFormatFile(t *testing.T) {
	t.Run("creates file and writes content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := tmpDir + "/test.txt"

		cfg := &config.Config{Threshold: 15}
		matches := []syntax.Match{}
		parseStats := job.ParseStats{
			ParseStatsMixin: job.ParseStatsMixin{FilesCount: 10, LinesCount: 100},
		}
		format := config.OutputFormatText
		sortByEnum := printer.SortBySize

		err := writeFormatFile(
			t.Context(),
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

func TestPassesFileCheck(t *testing.T) {
	tests := []struct {
		name      string
		fileCheck fileCheckFunc
		want      bool
	}{
		{
			name:      "nil check accepts all",
			fileCheck: nil,
			want:      true,
		},
		{
			name:      "check that returns true",
			fileCheck: func(_ string) bool { return true },
			want:      true,
		},
		{
			name:      "check that returns false",
			fileCheck: func(_ string) bool { return false },
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passesFileCheck("test.go", tt.fileCheck)
			if got != tt.want {
				t.Errorf("passesFileCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		includeVendor      bool
		includeNodeModules bool
		want               bool
	}{
		// Vendor directory tests
		{"vendor prefix excluded", "vendor/github.com/foo/bar.go", false, false, true},
		{"vendor in path excluded", "project/vendor/github.com/foo", false, false, true},
		{"vendor prefix included", "vendor/github.com/foo/bar.go", true, false, false},
		{"regular path no vendor", "project/src/main.go", false, false, false},

		// Git directory tests
		{"git prefix excluded", ".git/config", false, false, true},
		{"git in path excluded", "project/.git/objects", false, false, true},
		{"regular path no git", "project/src/main.go", false, false, false},

		// node_modules tests
		{"node_modules prefix excluded", "node_modules/lodash/index.js", false, false, true},
		{"node_modules in path excluded", "project/node_modules/react", false, false, true},
		{"node_modules included", "node_modules/lodash/index.js", false, true, false},

		// Edge cases
		{"empty path", "", false, false, false},
		{
			"just vendor (exact match)",
			"vendor",
			false,
			false,
			false,
		}, // Note: exact "vendor" doesn't have separator
		{
			"just node_modules (exact match)",
			"node_modules",
			false,
			false,
			false,
		}, // Note: exact match doesn't have separator
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipPath(tt.path, tt.includeVendor, tt.includeNodeModules)
			if got != tt.want {
				t.Errorf("shouldSkipPath(%q, vendor=%v, node_modules=%v) = %v, want %v",
					tt.path, tt.includeVendor, tt.includeNodeModules, got, tt.want)
			}
		})
	}
}

func TestCrawlPathsAllFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files including non-source
	testFiles := []string{
		"main.go",
		"readme.md",
		"app.js",
	}

	for _, f := range testFiles {
		path := tempDir + "/" + f

		err := os.WriteFile(path, []byte("test"), 0o644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	t.Run("crawls all files with nil check", func(t *testing.T) {
		f := gogenfilter.NewFilter(false, nil)
		files := collectStrings(crawlPathsAllFiles([]string{tempDir}, f, true, true, ""))

		// Should find all 3 files
		if len(files) != 3 {
			t.Errorf("Expected 3 files, got %d: %v", len(files), files)
		}
	})
}

func TestCrawlSinglePath_File(t *testing.T) {
	tempDir := t.TempDir()

	testFile := tempDir + "/test.go"

	err := os.WriteFile(testFile, []byte("package main"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	f := gogenfilter.NewFilter(false, nil)
	fchan := make(chan string, 10)

	crawlSinglePathWithOpts(CrawlOptions{
		Filter:          f,
		IncludeVendor:   false,
		IncludeNodeMods: false,
		FileCheck:       isSourceFile,
		FChan:           fchan,
	}, testFile)
	close(fchan)

	files := collectStrings(fchan)
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if len(files) > 0 && files[0] != testFile {
		t.Errorf("Expected %s, got %s", testFile, files[0])
	}
}

func TestHandleWalkEntry(t *testing.T) {
	tempDir := t.TempDir()

	testFile := tempDir + "/test.go"
	testutil.WriteTestFile(t, testFile, "package main")

	info, err := os.Lstat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		info      os.FileInfo
		fileCheck fileCheckFunc
		wantErr   bool
		wantSent  bool
	}{
		{
			name:      "valid Go file",
			path:      testFile,
			info:      info,
			fileCheck: isSourceFile,
			wantErr:   false,
			wantSent:  true,
		},
		{
			name:      "nil info skips",
			path:      testFile,
			info:      nil,
			fileCheck: isSourceFile,
			wantErr:   false,
			wantSent:  false,
		},
		{
			name:      "DS_Store file excluded",
			path:      tempDir + "/.DS_Store",
			info:      &mockFileInfo{name: ".DS_Store", isDir: false},
			fileCheck: isSourceFile,
			wantErr:   false,
			wantSent:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := gogenfilter.NewFilter(false, nil)
			fchan := make(chan string, 1)

			err := handleWalkEntry(CrawlOptions{
				Filter:          f,
				IncludeVendor:   false,
				IncludeNodeMods: false,
				FileCheck:       tt.fileCheck,
				FChan:           fchan,
			}, tt.path, tt.info)
			close(fchan)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleWalkEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			files := collectStrings(fchan)
			if tt.wantSent && len(files) == 0 {
				t.Error("Expected file to be sent to channel")
			}

			if !tt.wantSent && len(files) > 0 {
				t.Error("Expected no file to be sent to channel")
			}
		})
	}
}

// collectStrings collects strings from channel.
func collectStrings(ch <-chan string) []string {
	var result []string
	for s := range ch {
		result = append(result, s)
	}

	return result
}

// mockFileInfo is a mock implementation of os.FileInfo for testing.
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }
