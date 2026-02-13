package filter

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/quick"
)

// Helper function for contains check.
func contains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}

// runTestCases is a generic helper for running table-driven tests.
func runTestCases[T any, R comparable](t *testing.T, tests []T, nameFunc func(T) string, testFunc, expectFunc func(T) R) {
	t.Helper()
	for _, tt := range tests {
		t.Run(nameFunc(tt), func(t *testing.T) {
			t.Helper()
			result := testFunc(tt)
			if result != expectFunc(tt) {
				t.Errorf("Expected %v, got %v", expectFunc(tt), result)
			}
		})
	}
}

// Test case types for explicit typing.
type fileContentTest struct {
	name     string
	filePath string
	content  string
	expected bool
}

func (t fileContentTest) GetName() string   { return t.name }
func (t fileContentTest) GetExpected() bool { return t.expected }

// runFileContentTestCases is a helper for running file content test cases.
func runFileContentTestCases(t *testing.T, tests []fileContentTest, detectionFunc func(string, string) bool) {
	t.Helper()
	runTestCases(t, tests,
		fileContentTest.GetName,
		func(tt fileContentTest) bool { return detectionFunc(tt.filePath, tt.content) },
		fileContentTest.GetExpected,
	)
}

// runSimpleTestCases is a generic helper for running simple test cases with a single function.
func runSimpleTestCases[T any, R comparable](t *testing.T, tests []T, testFunc func(T) R, nameFunc func(T) string, expectFunc func(T) R) {
	t.Helper()
	runTestCases(t, tests, nameFunc, testFunc, expectFunc)
}

// runGenericTestCases is a generic helper for running test cases that use GetName and GetExpected methods.
type testCaseInterface[T any, R comparable] interface {
	GetName() string
	GetExpected() R
}

func runGenericTestCases[T testCaseInterface[T, R], R comparable](t *testing.T, tests []T, testFunc func(T) R) {
	t.Helper()
	runSimpleTestCases(t, tests,
		testFunc,
		func(tt T) string { return tt.GetName() },
		func(tt T) R { return tt.GetExpected() },
	)
}

// runMatchPatternTestCases is a helper for running match pattern test cases.
func runMatchPatternTestCases(t *testing.T, tests []matchPatternTest) {
	t.Helper()
	runGenericTestCases(t, tests,
		func(tt matchPatternTest) bool { return matchPattern(tt.path, tt.pattern) },
	)
}

type containsTest struct {
	name   string
	slice  []string
	item   string
	expect bool
}

func (t containsTest) GetName() string   { return t.name }
func (t containsTest) GetExpected() bool { return t.expect }

// runContainsTestCases is a helper for running contains test cases.
func runContainsTestCases(t *testing.T, tests []containsTest) {
	t.Helper()
	runGenericTestCases(t, tests,
		func(tt containsTest) bool { return contains(tt.slice, tt.item) },
	)
}

type matchPatternTest struct {
	name     string
	path     string
	pattern  string
	expected bool
}

func (t matchPatternTest) GetName() string   { return t.name }
func (t matchPatternTest) GetExpected() bool { return t.expected }

func TestNewFilter(t *testing.T) {
	t.Parallel()

	t.Run("creates disabled filter", func(t *testing.T) {
		f := NewFilter(false, nil)
		if f.enabled {
			t.Error("Expected disabled filter")
		}
		if len(f.options) != 0 {
			t.Errorf("Expected empty options, got %v", f.options)
		}
	})

	t.Run("creates enabled filter with options", func(t *testing.T) {
		f := NewFilter(true, []FilterOption{FilterSQLC, FilterTempl})
		if !f.enabled {
			t.Error("Expected enabled filter")
		}
		if !f.options[FilterSQLC] {
			t.Error("Expected SQLC option enabled")
		}
		if !f.options[FilterTempl] {
			t.Error("Expected Templ option enabled")
		}
	})

	t.Run("creates enabled filter with FilterAll", func(t *testing.T) {
		f := NewFilter(true, []FilterOption{FilterAll})
		if !f.enabled {
			t.Error("Expected enabled filter")
		}
		if !f.options[FilterSQLC] {
			t.Error("Expected SQLC option enabled for FilterAll")
		}
		if !f.options[FilterTempl] {
			t.Error("Expected Templ option enabled for FilterAll")
		}
		if !f.options[FilterGoEnum] {
			t.Error("Expected GoEnum option enabled for FilterAll")
		}
	})
}

// patternSetter is a function type for setting patterns on a Filter.
type patternSetter func(*Filter, []string)

// Common test pattern constants to avoid duplication.
var (
	testIncludePatterns = []string{"vendor/*", "generated/keep.go"}
	testExcludePatterns = []string{"test/*", "*.pb.go"}
)

// withPatterns runs a test for pattern setting methods.
func withPatterns(t *testing.T, patternType string, setPatterns patternSetter, getPatterns func(*Filter) []string, patterns []string) {
	t.Helper()
	t.Parallel()
	f := NewFilter(true, []FilterOption{FilterAll})
	setPatterns(f, patterns)
	testPatternSlices(t, patternType, getPatterns(f), patterns)
}

// patternTestCase represents a test case for pattern setting methods.
type patternTestCase struct {
	name        string
	patternType string
	setPatterns func(*Filter, []string)
	getPatterns func(*Filter) []string
	patterns    []string
}

func (t patternTestCase) GetName() string { return t.name }

func TestPatternSetting(t *testing.T) {
	t.Parallel()

	tests := []patternTestCase{
		{
			name:        "WithIncludePatterns",
			patternType: "Include",
			setPatterns: func(f *Filter, p []string) { f.WithIncludePatterns(p) },
			getPatterns: func(f *Filter) []string { return f.includePatterns },
			patterns:    testIncludePatterns,
		},
		{
			name:        "WithExcludePatterns",
			patternType: "Exclude",
			setPatterns: func(f *Filter, p []string) { f.WithExcludePatterns(p) },
			getPatterns: func(f *Filter) []string { return f.excludePatterns },
			patterns:    testExcludePatterns,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPatterns(t, tt.patternType, tt.setPatterns, tt.getPatterns, tt.patterns)
		})
	}
}

// testPatternSlices is a helper for testing pattern slices.
func testPatternSlices(t *testing.T, patternType string, patterns, wantPatterns []string) {
	t.Helper()
	if len(patterns) != len(wantPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(wantPatterns), len(patterns))
	}
	for _, pattern := range wantPatterns {
		if !contains(patterns, pattern) {
			t.Errorf("Expected %s in %s patterns", pattern, patternType)
		}
	}
}

func TestShouldFilter(t *testing.T) {
	t.Parallel()

	t.Run("disabled filter never filters", func(t *testing.T) {
		f := NewFilter(false, []FilterOption{FilterAll})
		if f.ShouldFilter("any/file.go") {
			t.Error("Disabled filter should not filter")
		}
	})
}

func TestMatchPattern(t *testing.T) {
	t.Parallel()

	tests := []matchPatternTest{
		{
			name:     "exact match",
			path:     "file.go",
			pattern:  "file.go",
			expected: true,
		},
		{
			name:     "wildcard match",
			path:     "test.go",
			pattern:  "*.go",
			expected: true,
		},
		{
			name:     "wildcard no match",
			path:     "test.txt",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "directory pattern",
			path:     "vendor/file.go",
			pattern:  "vendor/*",
			expected: true,
		},
		{
			name:     "complex pattern",
			path:     "generated/sqlc/models.go",
			pattern:  "generated/sqlc/*.go",
			expected: true,
		},
	}

	runMatchPatternTestCases(t, tests)
}

// Property-based tests for filter logic

func TestFilterIdempotentProperty(t *testing.T) {
	t.Parallel()

	// Property: Filter should be idempotent (applying twice gives same result)
	f := func(filePath string) bool {
		if filePath == "" || len(filePath) < 1 {
			return true // Empty path, skip
		}

		filter1 := NewFilter(true, nil)
		filter2 := NewFilter(true, nil)

		result1 := filter1.ShouldFilter(filePath)
		result2 := filter2.ShouldFilter(filePath)

		return result1 == result2
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Idempotent property failed: %v", err)
	}
}

func TestDisabledFilterProperty(t *testing.T) {
	t.Parallel()

	// Property: Disabled filter never filters
	f := func(filePath string) bool {
		if filePath == "" {
			return true // Empty path, skip
		}

		filter := NewFilter(false, nil)
		return !filter.ShouldFilter(filePath)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Disabled filter property failed: %v", err)
	}
}

func TestIncludePatternProperty(t *testing.T) {
	t.Parallel()

	// Property: Files matching include pattern are not filtered
	// This property only applies when the pattern actually matches the path
	f := func(includePattern, filePath string) bool {
		// Skip invalid inputs
		if includePattern == "" || filePath == "" {
			return true // Empty strings, skip
		}

		filter := NewFilter(true, nil)
		filter.WithIncludePatterns([]string{includePattern})

		// Only test the case where pattern matches path
		if !matchPattern(filePath, includePattern) {
			return true // Pattern doesn't match, skip this test case
		}

		// If pattern matches, filter should NOT filter (return false)
		if filter.ShouldFilter(filePath) {
			return false // Failed: pattern matched but file would be filtered
		}

		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Include pattern property failed: %v", err)
	}
}

func TestExcludePatternProperty(t *testing.T) {
	t.Parallel()

	// Property: Files matching exclude pattern are filtered
	f := func(excludePattern, filePath string) bool {
		if excludePattern == "" || filePath == "" {
			return true // Empty strings, skip
		}

		filter := NewFilter(true, nil)
		filter.WithExcludePatterns([]string{excludePattern})

		// If pattern should match, filter should return true
		shouldFilter := matchPattern(filePath, excludePattern)
		isFiltered := filter.ShouldFilter(filePath)

		// The two should match
		return shouldFilter == isFiltered
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Exclude pattern property failed: %v", err)
	}
}

// Integration test helpers

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return filePath
}

// assertFilterBehavior creates a temp file and asserts the filter behavior.
func assertFilterBehavior(t *testing.T, name, content string, opts []FilterOption, shouldFilter bool) {
	t.Helper()
	tmpFile := createTempFile(t, name, content)
	f := NewFilter(true, opts)
	got := f.ShouldFilter(tmpFile)
	if got != shouldFilter {
		t.Errorf("ShouldFilter() = %v, want %v", got, shouldFilter)
	}
}

func TestShouldFilterIntegration(t *testing.T) {
	t.Parallel()

	t.Run("filters sqlc file", func(t *testing.T) {
		assertFilterBehavior(t, "models.go", "// Code generated by sqlc. DO NOT EDIT.\npackage db\ntype User struct {}\n",
			[]FilterOption{FilterSQLC}, true)
	})

	t.Run("filters templ file", func(t *testing.T) {
		assertFilterBehavior(t, "header_templ.go", "package components\nimport \"github.com/a-h/templ\"\nfunc header() templ.Component { return nil }\n",
			[]FilterOption{FilterTempl}, true)
	})

	t.Run("does not filter regular file", func(t *testing.T) {
		assertFilterBehavior(t, "main.go", "package main\nfunc main() {}\n",
			[]FilterOption{FilterAll}, false)
	})

	t.Run("filters go-enum file", func(t *testing.T) {
		assertFilterBehavior(t, "status_enum.go", "// Code generated by go-enum DO NOT EDIT.\npackage enums\nconst StatusPending Status = iota\n",
			[]FilterOption{FilterGoEnum}, true)
	})
}

func TestIsSQLCGenerated(t *testing.T) {
	t.Parallel()

	tests := []fileContentTest{
		{
			name:     "sqlc models with comment",
			filePath: "db/models.go",
			content: `// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: schema.sql

package db

type User struct {
	ID int
}
`,
			expected: true,
		},
		{
			name:     "sqlc query with version comment",
			filePath: "db/query.sql.go",
			content: `package db

// versions:
//   sqlc v1.25.0

const getAuthors = ` + "`-- name: GetAuthors :many\nSELECT * FROM authors`" + `
`,
			expected: true,
		},
		{
			name:     "sqlc articles.sql.go with comment",
			filePath: "internal/storage/queries/articles.sql.go",
			content: `// Code generated by sqlc. DO NOT EDIT.
package queries

type Article struct {
	ID    int64
	Title string
}
`,
			expected: true,
		},
		{
			name:     "sqlc users.sql.go with comment",
			filePath: "internal/storage/queries/users.sql.go",
			content: `// Code generated by sqlc. DO NOT EDIT.
package queries

type User struct {
	ID   int64
	Name string
}
`,
			expected: true,
		},
		{
			name:     "not sqlc - regular Go file",
			filePath: "models/user.go",
			content: `package models

type User struct {
	ID string
}
`,
			expected: false,
		},
	}

	runFileContentTestCases(t, tests, isSQLCGenerated)
}

func TestIsTemplGenerated(t *testing.T) {
	t.Parallel()

	tests := []fileContentTest{
		{
			name:     "templ generated file",
			filePath: "components/header_templ.go",
			content: `// Code generated by templ DO NOT EDIT

package components

import (
	"context"
	"io"
	"github.com/a-h/templ"
)

func header(name string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		// implementation
		return nil
	})
}
`,
			expected: true,
		},
		{
			name:     "not templ - regular file",
			filePath: "components/helper.go",
			content: `package components

func Helper() string {
	return "helper"
}
`,
			expected: false,
		},
	}

	runFileContentTestCases(t, tests, isTemplGenerated)
}

func TestIsGoEnumGenerated(t *testing.T) {
	t.Parallel()

	tests := []fileContentTest{
		{
			name:     "go-enum generated file with comment",
			filePath: "enums/status_enum.go",
			content: `// Code generated by go-enum DO NOT EDIT.
// Version: v0.5.0
// Revision: abc123

package enums

type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusCompleted
)

func (s Status) String() string { return "" }
`,
			expected: true,
		},
		{
			name:     "go-enum generated file with version info",
			filePath: "types/priority_enum.go",
			content: `// Code generated by go-enum
// Build Date: 2024-01-15

package types

type Priority int

const PriorityLow Priority = iota
`,
			expected: true,
		},
		{
			name:     "not go-enum - regular file",
			filePath: "models/user.go",
			content: `package models

type User struct {
	ID   int
	Name string
}
`,
			expected: false,
		},
		{
			name:     "wrong filename - not _enum.go suffix",
			filePath: "enums/status.go",
			content: `// Code generated by go-enum DO NOT EDIT.
package enums

type Status int
`,
			expected: false,
		},
	}

	runFileContentTestCases(t, tests, isGoEnumGenerated)
}

func TestStringContains(t *testing.T) {
	t.Parallel()

	// Test our contains helper function
	tests := []containsTest{
		{
			name:   "contains item",
			slice:  []string{"a", "b", "c"},
			item:   "b",
			expect: true,
		},
		{
			name:   "does not contain item",
			slice:  []string{"a", "b", "c"},
			item:   "d",
			expect: false,
		},
		{
			name:   "empty slice",
			slice:  []string{},
			item:   "a",
			expect: false,
		},
	}

	runContainsTestCases(t, tests)
}

func TestPatternMatching(t *testing.T) {
	t.Parallel()

	t.Run("filename matching works", func(t *testing.T) {
		paths := []string{"file.go", "test.txt", "main.go"}
		pattern := "*.go"

		for _, path := range paths {
			match := matchPattern(path, pattern)
			shouldBeMatch := strings.HasSuffix(path, ".go")
			if match != shouldBeMatch {
				t.Errorf("Pattern %q with path %q: got %v, want %v", pattern, path, match, shouldBeMatch)
			}
		}
	})

	t.Run("directory pattern works", func(t *testing.T) {
		paths := []string{
			"vendor/file.go",
			"src/main.go",
			"vendor/subdir/other.go",
		}
		pattern := "vendor/*"

		for _, path := range paths {
			match := matchPattern(path, pattern)
			shouldBeMatch := strings.HasPrefix(path, "vendor/") || strings.Contains(path, "/vendor/")
			if match != shouldBeMatch {
				t.Errorf("Pattern %q with path %q: got %v, want %v", pattern, path, match, shouldBeMatch)
			}
		}
	})
}

func TestFilterMetrics(t *testing.T) {
	t.Parallel()

	t.Run("tracks filtered files by reason", func(t *testing.T) {
		metrics := NewMetrics()

		// Record some filtered files
		metrics.Record("db/models.go", ReasonSQLC)
		metrics.Record("db/query.sql.go", ReasonSQLC)
		metrics.Record("components/header_templ.go", ReasonTempl)
		metrics.Record("vendor/lib.go", ReasonExcludePattern)

		stats := metrics.GetStats()

		if stats.TotalFilesChecked != 4 {
			t.Errorf("Expected TotalFilesChecked=4, got %d", stats.TotalFilesChecked)
		}
		if stats.FilteredByReason[ReasonSQLC] != 2 {
			t.Errorf("Expected SQLC count=2, got %d", stats.FilteredByReason[ReasonSQLC])
		}
		if stats.FilteredByReason[ReasonTempl] != 1 {
			t.Errorf("Expected Templ count=1, got %d", stats.FilteredByReason[ReasonTempl])
		}
		if stats.FilteredByReason[ReasonExcludePattern] != 1 {
			t.Errorf("Expected ExcludePattern count=1, got %d", stats.FilteredByReason[ReasonExcludePattern])
		}
		if stats.TotalFiltered() != 4 {
			t.Errorf("Expected TotalFiltered=4, got %d", stats.TotalFiltered())
		}
	})

	t.Run("tracks not filtered files", func(t *testing.T) {
		metrics := NewMetrics()

		// Record both filtered and not filtered
		metrics.Record("db/models.go", ReasonSQLC)
		metrics.Record("main.go", ReasonNotFiltered)
		metrics.Record("service/user.go", ReasonNotFiltered)

		stats := metrics.GetStats()

		if stats.TotalFilesChecked != 3 {
			t.Errorf("Expected TotalFilesChecked=3, got %d", stats.TotalFilesChecked)
		}
		if stats.FilteredByReason[ReasonSQLC] != 1 {
			t.Errorf("Expected SQLC count=1, got %d", stats.FilteredByReason[ReasonSQLC])
		}
		if stats.TotalFiltered() != 1 {
			t.Errorf("Expected TotalFiltered=1, got %d", stats.TotalFiltered())
		}
	})

	t.Run("nil metrics handler", func(t *testing.T) {
		var metrics *Metrics

		// Should not panic
		metrics.Record("test.go", ReasonSQLC)
		stats := metrics.GetStats()

		if stats.TotalFilesChecked != 0 {
			t.Errorf("Expected empty stats, got %+v", stats)
		}
	})
}

func TestFilterWithMetrics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"db/models.go": `// Code generated by sqlc. DO NOT EDIT.
package db
type User struct{ ID int64 }`,
		"main.go": `package main
func main() {}`,
		"components/header_templ.go": `// Code generated by templ DO NOT EDIT

package components

import (
    "context"
    "io"
    "github.com/a-h/templ"
)

func Header(title string) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, err := io.WriteString(w, "<header>"+title+"</header>")
        return err
    })
}`,
		"enums/status_enum.go": `// Code generated by go-enum DO NOT EDIT.
package enums
type Status int
const StatusPending Status = iota`,
	}

	for name, content := range files {
		dir := filepath.Join(tmpDir, filepath.Dir(name))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Create and use filter
	fltr := NewFilter(true, []FilterOption{FilterAll})

	// Filter the files
	for name := range files {
		filePath := filepath.Join(tmpDir, name)
		_ = fltr.ShouldFilter(filePath)
	}

	// Check metrics
	stats := fltr.GetStats()

	if stats.TotalFilesChecked != 4 {
		t.Errorf("Expected TotalFilesChecked=4, got %d", stats.TotalFilesChecked)
	}

	// Verify SQLC file was filtered
	if stats.FilteredByReason[ReasonSQLC] != 1 {
		t.Errorf("Expected 1 SQLC file filtered, got %d", stats.FilteredByReason[ReasonSQLC])
	}

	// Verify Templ file was filtered
	if stats.FilteredByReason[ReasonTempl] != 1 {
		t.Errorf("Expected 1 Templ file filtered, got %d", stats.FilteredByReason[ReasonTempl])
	}

	// Verify GoEnum file was filtered
	if stats.FilteredByReason[ReasonGoEnum] != 1 {
		t.Errorf("Expected 1 GoEnum file filtered, got %d", stats.FilteredByReason[ReasonGoEnum])
	}

	// Verify at least 3 files were filtered total (SQLC + Templ + GoEnum)
	totalFiltered := stats.TotalFiltered()
	if totalFiltered < 3 {
		t.Errorf("Expected at least 3 files filtered, got %d", totalFiltered)
	}

	// Verify the main.go file was processed (metrics recorded for it)
	if stats.FilteredByReason[ReasonNotFiltered] > 1 {
		t.Errorf("Unexpected number of not-filtered files: %d", stats.FilteredByReason[ReasonNotFiltered])
	}
}
