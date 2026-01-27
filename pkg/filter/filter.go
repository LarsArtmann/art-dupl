package filter

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FilterOption represents a type of generated code to filter.
type FilterOption string

const (
	// FilterSQLC filters sqlc.dev generated files.
	FilterSQLC FilterOption = "sqlc"

	// FilterTempl filters templ.guide generated files.
	FilterTempl FilterOption = "templ"

	// FilterAll filters all auto-generated code.
	FilterAll FilterOption = "all"
)

// FilterReason represents the reason a file was filtered.
type FilterReason string

const (
	ReasonSQLC         FilterReason = "sqlc"
	ReasonTempl        FilterReason = "templ"
	ReasonIncludePattern FilterReason = "include-pattern"
	ReasonExcludePattern FilterReason = "exclude-pattern"
	ReasonNotFiltered  FilterReason = "not-filtered"
)

// Metrics tracks filter statistics for analysis and debugging.
type Metrics struct {
	mu sync.RWMutex
	
	// TotalFilesChecked is the total number of files the filter evaluated
	TotalFilesChecked int
	
	// FilteredByReason tracks how many files were filtered for each reason
	FilteredByReason map[FilterReason]int
	
	// FilteredFiles maps reason to list of files filtered for that reason
	FilteredFiles map[FilterReason][]string
}

// NewMetrics creates a new filter metrics tracker.
func NewMetrics() *Metrics {
	return &Metrics{
		FilteredByReason: make(map[FilterReason]int),
		FilteredFiles:    make(map[FilterReason][]string),
	}
}

// Record records that a file was filtered for a given reason.
// OPTIMIZATION: Handle nil metrics for better performance.
func (m *Metrics) Record(filePath string, reason FilterReason) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalFilesChecked++

	if reason != ReasonNotFiltered {
		m.FilteredByReason[reason]++
		m.FilteredFiles[reason] = append(m.FilteredFiles[reason], filePath)
	}
}

// GetStats returns the current filter statistics.
func (m *Metrics) GetStats() FilterStats {
	if m == nil {
		return FilterStats{}
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Create copies of the maps to avoid concurrent access issues
	filteredByReason := make(map[FilterReason]int)
	for k, v := range m.FilteredByReason {
		filteredByReason[k] = v
	}
	
	return FilterStats{
		TotalFilesChecked: m.TotalFilesChecked,
		FilteredByReason:  filteredByReason,
	}
}

// FilterStats represents a snapshot of filter statistics.
type FilterStats struct {
	TotalFilesChecked int
	FilteredByReason  map[FilterReason]int
}

// TotalFiltered returns the total number of filtered files.
func (fs FilterStats) TotalFiltered() int {
	total := 0
	for reason, count := range fs.FilteredByReason {
		if reason != ReasonNotFiltered {
			total += count
		}
	}
	return total
}


// Filter provides smart filtering of auto-generated Go code.
type Filter struct {
	options         map[FilterOption]bool
	enabled         bool
	includePatterns []string
	excludePatterns []string
	metrics         *Metrics
}

// NewFilter creates a new filter with specified options.
// OPTIMIZATION: Only create metrics if filtering is actually enabled.
func NewFilter(enabled bool, options []FilterOption) *Filter {
	f := &Filter{
		enabled:         enabled,
		options:         make(map[FilterOption]bool),
		includePatterns: make([]string, 0),
		excludePatterns: make([]string, 0),
		// OPTIMIZATION: Only create metrics if filter is enabled
		metrics: nil,
	}

	for _, opt := range options {
		if opt == FilterAll {
			f.options[FilterSQLC] = true
			f.options[FilterTempl] = true
		} else {
			f.options[opt] = true
		}
	}

	// OPTIMIZATION: Create metrics only if filter is enabled
	if enabled {
		f.metrics = NewMetrics()
	}

	return f
}

// WithIncludePatterns adds custom include patterns to the filter.
func (f *Filter) WithIncludePatterns(patterns []string) *Filter {
	f.includePatterns = append(f.includePatterns, patterns...)
	return f
}

// WithExcludePatterns adds custom exclude patterns to the filter.
func (f *Filter) WithExcludePatterns(patterns []string) *Filter {
	f.excludePatterns = append(f.excludePatterns, patterns...)
	return f
}

// ShouldFilter determines if a file should be filtered out (excluded from analysis).
// Returns true if the file should be filtered, false otherwise.
func (f *Filter) ShouldFilter(filePath string) bool {
	if !f.enabled {
		if f.metrics != nil {
			f.metrics.Record(filePath, ReasonNotFiltered)
		}
		return false
	}

	// If include patterns are specified, use strict include-only logic
	// Only files matching include patterns are analyzed, all others are filtered out
	if len(f.includePatterns) > 0 {
		for _, pattern := range f.includePatterns {
			if matchPattern(filePath, pattern) {
				if f.metrics != nil {
					f.metrics.Record(filePath, ReasonNotFiltered)
				}
				return false // Don't filter if it matches include pattern
			}
		}
		if f.metrics != nil {
			f.metrics.Record(filePath, ReasonIncludePattern)
		}
		return true // Filter out if it doesn't match ANY include pattern
	}

	// No include patterns specified, use exclude logic
	// Check exclude patterns
	for _, pattern := range f.excludePatterns {
		if matchPattern(filePath, pattern) {
			if f.metrics != nil {
				f.metrics.Record(filePath, ReasonExcludePattern)
			}
			return true // Filter if it matches exclude pattern
		}
	}

	// Check auto-generated code patterns
	if f.isAutoGenerated(filePath) {
		reason := f.getAutoGeneratedReason(filePath)
		if f.metrics != nil {
			f.metrics.Record(filePath, reason)
		}
		return true
	}

	if f.metrics != nil {
		f.metrics.Record(filePath, ReasonNotFiltered)
	}
	return false
}

// isAutoGenerated checks if a file is auto-generated based on content and patterns.
func (f *Filter) isAutoGenerated(filePath string) bool {
	return f.getAutoGeneratedReason(filePath) != ReasonNotFiltered
}

// getAutoGeneratedReason determines if and why a file is auto-generated.
// OPTIMIZATION: Do filename-based check FIRST (no I/O), only read file if necessary.
func (f *Filter) getAutoGeneratedReason(filePath string) FilterReason {
	// OPTIMIZATION: Try filename-based detection first (no I/O)
	// Only read file content if filename check is inconclusive
	filenameBasedReason := f.getFilenameBasedReason(filePath)
	if filenameBasedReason != ReasonNotFiltered {
		return filenameBasedReason
	}

	// Filename check was inconclusive, read file for content-based detection
	// OPTIMIZATION: Only read if filename suggests it might be generated
	content, err := os.ReadFile(filePath) //nolint:gosec //G304 filePath is from controlled source (path crawling)
	if err != nil {
		// If we can't read the file, fall back to filename-based detection only
		return filenameBasedReason
	}

	contentStr := string(content)

	// Check sqlc patterns
	if f.options[FilterSQLC] && isSQLCGenerated(filePath, contentStr) {
		return ReasonSQLC
	}

	// Check templ patterns
	if f.options[FilterTempl] && isTemplGenerated(filePath, contentStr) {
		return ReasonTempl
	}

	return ReasonNotFiltered
}

// getFilenameBasedReason determines the filter reason based on filename only.
func (f *Filter) getFilenameBasedReason(filePath string) FilterReason {
	filename := filepath.Base(filePath)

	// sqlc patterns
	if f.options[FilterSQLC] {
		sqlcPatterns := []string{
			"models.go",
			"querier.go",
			"query.sql.go",
			"batch.go",
		}
		for _, pattern := range sqlcPatterns {
			if strings.Contains(filename, pattern) {
				return ReasonSQLC
			}
		}
		
		// Also check for *.sql.go pattern (SQLC generates files like articles.sql.go, users.sql.go, etc.)
		if strings.HasSuffix(filename, ".sql.go") {
			return ReasonSQLC
		}
	}

	// templ patterns
	if f.options[FilterTempl] && strings.HasSuffix(filename, "_templ.go") {
		return ReasonTempl
	}

	return ReasonNotFiltered
}

// isGeneratedByFilename performs filename-based detection only (legacy, kept for compatibility).
func (f *Filter) isGeneratedByFilename(filePath string) bool {
	return f.getFilenameBasedReason(filePath) != ReasonNotFiltered
}

// GetMetrics returns the filter metrics tracker. Returns nil if metrics are disabled.
func (f *Filter) GetMetrics() *Metrics {
	return f.metrics
}

// GetStats returns a snapshot of filter statistics.
func (f *Filter) GetStats() FilterStats {
	if f.metrics == nil {
		return FilterStats{}
	}
	return f.metrics.GetStats()
}

// matchPattern checks if a path matches a pattern (supports * and ** wildcards).
func matchPattern(path, pattern string) bool {
	// If pattern contains a path separator, match against path components
	// This allows patterns like "pkg1/*" to match "/tmp/test/pkg1/file.go"
	if strings.Contains(pattern, string(filepath.Separator)) || strings.Contains(pattern, "/") {
		// Convert pattern to a simple path matching string
		// Replace "*" with empty string to get path prefix to match
		// e.g., "pkg1/*" -> "pkg1/"
		prefix := strings.Split(pattern, "*")[0]

		// Ensure prefix ends with separator for proper matching
		if !strings.HasSuffix(prefix, string(filepath.Separator)) && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		// Check if path contains prefix
		return strings.Contains(path, prefix)
	}

	// Pattern has no path separator, use filepath.Match for filename matching
	// This handles patterns like "*.go" correctly
	filename := filepath.Base(path)
	matched, err := filepath.Match(pattern, filename)
	if err != nil {
		// Invalid pattern, don't match
		return false
	}
	return matched
}

// isSQLCGenerated checks if a file is generated by sqlc.dev.
func isSQLCGenerated(filePath, content string) bool {
	filename := filepath.Base(filePath)

	// Check for sqlc file patterns
	sqlcFilePatterns := []string{
		"models.go",
		"querier.go",
		"query.sql.go",
		"batch.go",
	}
	isSQLCFile := false
	for _, pattern := range sqlcFilePatterns {
		if strings.Contains(filename, pattern) {
			isSQLCFile = true
			break
		}
	}
	
	// Also check for *.sql.go pattern (SQLC generates files like articles.sql.go, users.sql.go, etc.)
	if !isSQLCFile && strings.HasSuffix(filename, ".sql.go") {
		isSQLCFile = true
	}

	if !isSQLCFile {
		return false
	}

	// Check for sqlc comment header
	if strings.Contains(content, "// Code generated by sqlc. DO NOT EDIT.") {
		return true
	}

	// Check for sqlc version comment
	if strings.Contains(content, "sqlc ") && strings.Contains(content, "versions:") {
		return true
	}

	// Check for sqlc-specific code patterns
	sqlcCodePatterns := []string{
		"sqlc.Arg",
		"sqlc.NamedArg",
		"sqlc.Literal",
		"sqlc.SliceArg",
		"sqlc.Narg",
		".query(ctx", // Common sqlc method pattern
	}
	for _, pattern := range sqlcCodePatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// isTemplGenerated checks if a file is generated by templ.guide.
func isTemplGenerated(filePath, content string) bool {
	filename := filepath.Base(filePath)

	// Check for templ file pattern (*_templ.go)
	if !strings.HasSuffix(filename, "_templ.go") {
		return false
	}

	// Check for templ-specific code patterns
	if strings.Contains(content, "templ.Component") {
		return true
	}

	// Check for templ render function signature
	if strings.Contains(content, "Render(ctx context.Context, w io.Writer) error") {
		return true
	}

	return false
}
