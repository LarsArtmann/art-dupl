package filter

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/quick"
)

// Helper function for contains check
func contains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}

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
	})
}

func TestWithIncludePatterns(t *testing.T) {
	t.Parallel()

	f := NewFilter(true, []FilterOption{FilterAll})
	f.WithIncludePatterns([]string{"vendor/*", "generated/keep.go"})

	if len(f.includePatterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(f.includePatterns))
	}
	if !contains(f.includePatterns, "vendor/*") {
		t.Error("Expected vendor/* in include patterns")
	}
	if !contains(f.includePatterns, "generated/keep.go") {
		t.Error("Expected generated/keep.go in include patterns")
	}
}

func TestWithExcludePatterns(t *testing.T) {
	t.Parallel()

	f := NewFilter(true, []FilterOption{FilterAll})
	f.WithExcludePatterns([]string{"test/*", "*.pb.go"})

	if len(f.excludePatterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(f.excludePatterns))
	}
	if !contains(f.excludePatterns, "test/*") {
		t.Error("Expected test/* in exclude patterns")
	}
	if !contains(f.excludePatterns, "*.pb.go") {
		t.Error("Expected *.pb.go in exclude patterns")
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

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("For pattern %q and path %q: got %v, want %v", tt.pattern, tt.path, result, tt.expected)
			}
		})
	}
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
	f := func(includePattern, filePath string) bool {
		// Skip invalid inputs
		if includePattern == "" || filePath == "" {
			return true // Empty strings, skip
		}

		filter := NewFilter(true, nil)
		filter.WithIncludePatterns([]string{includePattern})

		// If pattern should match, filter should return false
		shouldNotFilter := matchPattern(filePath, includePattern)
		shouldBeFiltered := !filter.ShouldFilter(filePath)

		// The two should match
		return shouldNotFilter == shouldBeFiltered
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

func TestShouldFilterIntegration(t *testing.T) {
	t.Parallel()

	t.Run("filters sqlc file", func(t *testing.T) {
		sqlcFile := createTempFile(t, "models.go", `// Code generated by sqlc. DO NOT EDIT.
package db
type User struct {}
`)

		f := NewFilter(true, []FilterOption{FilterSQLC})
		if !f.ShouldFilter(sqlcFile) {
			t.Error("Should filter sqlc file")
		}
	})

	t.Run("filters templ file", func(t *testing.T) {
		templFile := createTempFile(t, "header_templ.go", `package components
import "github.com/a-h/templ"
func header() templ.Component { return nil }
`)

		f := NewFilter(true, []FilterOption{FilterTempl})
		if !f.ShouldFilter(templFile) {
			t.Error("Should filter templ file")
		}
	})

	t.Run("does not filter regular file", func(t *testing.T) {
		regularFile := createTempFile(t, "main.go", `package main
func main() {}
`)

		f := NewFilter(true, []FilterOption{FilterAll})
		if f.ShouldFilter(regularFile) {
			t.Error("Should not filter regular file")
		}
	})
}

func TestIsSQLCGenerated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		content  string
		expected bool
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSQLCGenerated(tt.filePath, tt.content)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsTemplGenerated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		content  string
		expected bool
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemplGenerated(tt.filePath, tt.content)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStringContains(t *testing.T) {
	t.Parallel()

	// Test our contains helper function
	tests := []struct {
		name   string
		slice  []string
		item   string
		expect bool
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expect {
				t.Errorf("Expected %v, got %v", tt.expect, result)
			}
		})
	}
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
