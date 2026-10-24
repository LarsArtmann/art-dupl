package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/gogenfilter/v3"
)

const testFile = "test.go"

const binaryName = "art-dupl"

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
			expected: binaryName,
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
			actual := detectionMethodsToString(tt.methods)
			testutil.ExpectTrue(t, actual == tt.expected, "detectionMethodsToString")
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
	tests := []uniqueTestCase{
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
					{Filename: testFile, Pos: 1, End: 10},
					{Filename: testFile, Pos: 11, End: 20},
				},
			},
			expected: 1,
		},
		newTableTest("duplicate entries same position", [][]*syntax.Node{
			testutil.CreateSingleNode(testFile, 1, 10),
			testutil.CreateSingleNode(testFile, 1, 10),
		}, 1),
		newTableTest("different positions", [][]*syntax.Node{
			testutil.CreateSingleNode("test1.go", 1, 10),
			testutil.CreateSingleNode("test2.go", 1, 10),
		}, 2),
		{
			name: "empty inner slice",
			input: [][]*syntax.Node{
				{},
				{
					{Filename: testFile, Pos: 1, End: 10},
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
	saveVersionGlobals(t)

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

			mockRead := func(path string) ([]byte, error) { return nil, nil }

			fn := createPrinter(
				tt.format,
				15,
				config.DiffModeDisabled,
				printer.ReportMetadata{},
				"test-version",
			)
			p := fn(&buf, mockRead)

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

		result := testutil.CollectMatches(ch)
		testutil.AssertCount(t, len(result), 0, "CollectMatches()")
	})

	t.Run("single match", func(t *testing.T) {
		ch := make(chan syntax.Match, 1)
		ch <- syntax.Match{
			Hash:  "abc123",
			Frags: [][]*syntax.Node{{{Filename: testFile, Pos: 1, End: 10}}},
		}

		close(ch)

		result := testutil.CollectMatches(ch)
		testutil.AssertCount(t, len(result), 1, "CollectMatches()")
	})

	t.Run("multiple matches", func(t *testing.T) {
		ch := make(chan syntax.Match, 3)

		ch <- testutil.CreateMatch("hash1", "test1.go")

		ch <- testutil.CreateMatch("hash2", "test2.go")

		ch <- testutil.CreateMatch("hash3", "test3.go")

		close(ch)

		result := testutil.CollectMatches(ch)
		testutil.AssertCount(t, len(result), 3, "CollectMatches()")
	})
}

func TestWriteFormatFile(t *testing.T) {
	t.Run("creates file and writes content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := tmpDir + "/test.txt"

		cfg := &config.Config{Threshold: 15}
		matches := []syntax.Match{}
		parseStats := job.ParseStats{
			FilesCount: 10, LinesCount: 100,
		}
		format := config.OutputFormatText
		sortByEnum := config.SortBySize

		err := writeFormatFile(
			t.Context(),
			cfg,
			matches,
			parseStats,
			format,
			filename,
			sortByEnum,
			binaryName,
			buildSuppressionConfig(cfg),
			io.Discard,
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
			path:     testFile,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeFile(nil, tt.path, nil, generatorIncludes{})
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
			actual := passesFileCheck(testFile, tt.fileCheck)
			testutil.ExpectTrue(t, actual == tt.want, "passesFileCheck")
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		includeVendor      bool
		includeNodeModules bool
		includeExamples    bool
		want               bool
	}{
		// Vendor directory tests
		{"vendor prefix excluded", "vendor/github.com/foo/bar.go", false, false, false, true},
		{"vendor in path excluded", "project/vendor/github.com/foo", false, false, false, true},
		{"vendor prefix included", "vendor/github.com/foo/bar.go", true, false, false, false},
		{"regular path no vendor", "project/src/main.go", false, false, false, false},

		// Git directory tests
		{"git prefix excluded", ".git/config", false, false, false, true},
		{"git in path excluded", "project/.git/objects", false, false, false, true},
		{"regular path no git", "project/src/main.go", false, false, false, false},

		// node_modules tests
		{"node_modules prefix excluded", "node_modules/lodash/index.js", false, false, false, true},
		{"node_modules in path excluded", "project/node_modules/react", false, false, false, true},
		{"node_modules included", "node_modules/lodash/index.js", false, true, false, false},

		// Example directory tests
		{"examples prefix excluded", "examples/foo/main.go", false, false, false, true},
		{"examples in path excluded", "project/examples/demo.go", false, false, false, true},
		{"demo prefix excluded", "demo/app.go", false, false, false, true},
		{"demo in path excluded", "pkg/demo/test.go", false, false, false, true},
		{"examples included", "examples/foo/main.go", false, false, true, false},

		// Both-separator cases (Windows invariant): filepath.Walk yields
		// platform-native separators, and shouldSkipPath must normalize them via
		// filepath.ToSlash. FromSlash produces backslash paths on Windows (a no-op
		// on POSIX), mirroring what the walker hands us on each OS.
		{"vendor native separators", filepath.FromSlash("vendor/github.com/foo/bar.go"), false, false, false, true},
		{"vendor in path native separators", filepath.FromSlash("project/vendor/github.com/foo"), false, false, false, true},
		{"git in path native separators", filepath.FromSlash("project/.git/objects"), false, false, false, true},
		{"node_modules in path native separators", filepath.FromSlash("project/node_modules/react"), false, false, false, true},
		{"examples in path native separators", filepath.FromSlash("project/examples/demo.go"), false, false, false, true},
		{"demo in path native separators", filepath.FromSlash("pkg/demo/test.go"), false, false, false, true},
		{
			"literal backslashes are only boundaries on Windows",
			`project\vendor\github.com\foo`,
			false,
			false,
			false,
			runtime.GOOS == "windows",
		}, // ToSlash rewrites separators only on Windows; on POSIX a backslash is a plain filename character, not a dir boundary

		// Edge cases
		{"empty path", "", false, false, false, false},
		{
			"just vendor (exact match)",
			"vendor",
			false,
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
			false,
		}, // Note: exact match doesn't have separator
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipPath(tt.path, tt.includeVendor, tt.includeNodeModules, tt.includeExamples)
			if got != tt.want {
				t.Errorf("shouldSkipPath(%q, vendor=%v, node_modules=%v, examples=%v) = %v, want %v",
					tt.path, tt.includeVendor, tt.includeNodeModules, tt.includeExamples, got, tt.want)
			}
		})
	}
}

func TestShouldPruneDir(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		includeVendor   bool
		includeNodeMods bool
		includeExamples bool
		want            bool
	}{
		{"vendor bare", "vendor", false, false, false, true},
		{"node_modules bare", "node_modules", false, false, false, true},
		{"git bare", ".git", false, false, false, true},
		{"examples bare", "examples", false, false, false, true},
		{"vendor nested", "project/vendor", false, false, false, true},
		{"node_modules nested", "website/node_modules", false, false, false, true},
		{"normal dir", "internal/cmd", false, false, false, false},
		{"root", ".", false, false, false, false},
		{"vendor included", "vendor", true, false, false, false},
		{"node_modules included", "node_modules", false, true, false, false},
		{"examples included", "examples", false, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPruneDir(tt.path, tt.includeVendor, tt.includeNodeMods, tt.includeExamples)
			if got != tt.want {
				t.Errorf("shouldPruneDir(%q, vendor=%v, node_modules=%v, examples=%v) = %v, want %v",
					tt.path, tt.includeVendor, tt.includeNodeMods, tt.includeExamples, got, tt.want)
			}
		})
	}
}

func TestHandleWalkEntryPrunesExcludedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	gitignoreContent := "logs/\nscratch.go\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	matcher := LoadGitignore([]string{tmpDir}, io.Discard)
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil")
	}

	newOpts := func() CrawlOptions {
		f, err := gogenfilter.NewFilter()
		if err != nil {
			t.Fatalf("NewFilter() error: %v", err)
		}

		return CrawlOptions{
			Ctx:             context.Background(),
			Filter:          f,
			IncludeVendor:   false,
			IncludeNodeMods: false,
			FileCheck:       isSourceFile,
			FChan:           make(chan string, 1),
			Gitignore:       matcher,
		}
	}

	tests := []struct {
		name     string
		path     string
		info     os.FileInfo
		wantSkip bool
	}{
		{
			name:     "root-level .git dir is pruned",
			path:     ".git",
			info:     &mockFileInfo{name: ".git", isDir: true},
			wantSkip: true,
		},
		{
			name:     "nested node_modules dir is pruned",
			path:     "website/node_modules",
			info:     &mockFileInfo{name: "node_modules", isDir: true},
			wantSkip: true,
		},
		{
			name:     "file under excluded dir is not a prune point",
			path:     "project/vendor/dep.go",
			info:     &mockFileInfo{name: "dep.go", isDir: false},
			wantSkip: false,
		},
		{
			name:     "gitignore-ignored dir is pruned",
			path:     filepath.Join(tmpDir, "logs"),
			info:     &mockFileInfo{name: "logs", isDir: true},
			wantSkip: true,
		},
		{
			name:     "gitignore-ignored file is not a prune point",
			path:     filepath.Join(tmpDir, "scratch.go"),
			info:     &mockFileInfo{name: "scratch.go", isDir: false},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleWalkEntry(newOpts(), tt.path, tt.info)

			if tt.wantSkip && !errors.Is(err, filepath.SkipDir) {
				t.Errorf("handleWalkEntry(%q) error = %v, want filepath.SkipDir", tt.path, err)
			}

			if !tt.wantSkip && err != nil {
				t.Errorf("handleWalkEntry(%q) error = %v, want nil", tt.path, err)
			}
		})
	}
}

// TestCrawlPrunesGitignoredDirs pins the walk-level contract that an excluded
// directory's entire subtree is skipped: files beneath it never reach the
// crawl channel, including files a negation tries to re-include (git semantics:
// a parent-excluded path cannot be re-included; the provider crawl behaves the
// same way).
func TestCrawlPrunesGitignoredDirs(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		".gitignore":            "node_modules/\nignored-dir/\n!ignored-dir/keep.go\n",
		"keep.go":               "package main\n",
		"node_modules/dep.go":   "package dep\n",
		"node_modules/keep.go":  "package dep\n",
		"ignored-dir/hidden.go": "package hidden\n",
		"ignored-dir/keep.go":   "package hidden\n",
	}

	for rel, content := range files {
		path := filepath.Join(tmpDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("MkdirAll %s: %v", filepath.Dir(path), err)
		}

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create %s: %v", path, err)
		}
	}

	f, err := gogenfilter.NewFilter()
	if err != nil {
		t.Fatalf("NewFilter() error: %v", err)
	}

	matcher := LoadGitignore([]string{tmpDir}, io.Discard)
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil")
	}

	got := collectStrings(
		crawlPaths(
			context.Background(),
			[]string{tmpDir},
			f,
			nil,
			generatorIncludes{},
			false,
			false,
			false,
			matcher,
			io.Discard,
		),
	)

	want := []string{filepath.Join(tmpDir, "keep.go")}

	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("crawl emitted %v, want exactly %v (ignored subtrees must be pruned)", got, want)
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

	t.Run("crawls all fileList with nil check", func(t *testing.T) {
		f, err := gogenfilter.NewFilter()
		if err != nil {
			t.Fatalf("NewFilter() error: %v", err)
		}

		fileList := collectStrings(
			crawlPathsAllFiles(
				context.Background(),
				[]string{tempDir},
				f,
				nil,
				generatorIncludes{},
				true,
				true,
				false,
				"",
				nil,
				io.Discard,
			),
		)

		// Should find all 3 fileList
		if len(fileList) != 3 {
			t.Errorf("Expected 3 fileList, got %d: %v", len(fileList), fileList)
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

	f, err := gogenfilter.NewFilter()
	if err != nil {
		t.Fatalf("NewFilter() error: %v", err)
	}

	fchan := make(chan string, 10)

	crawlSinglePathWithOpts(CrawlOptions{
		Ctx:             context.Background(),
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

	if len(files) > 0 && files[0] != filepath.ToSlash(testFile) {
		t.Errorf("Expected %s, got %s", filepath.ToSlash(testFile), files[0])
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
			f, err := gogenfilter.NewFilter()
			if err != nil {
				t.Fatalf("NewFilter() error: %v", err)
			}

			fchan := make(chan string, 1)

			err = handleWalkEntry(CrawlOptions{
				Ctx:             context.Background(),
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
