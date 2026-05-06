package config

import "slices"

// DetectionMethods is a slice of DetectionMethod for type safety.
type DetectionMethods []DetectionMethod

// IsDefault checks if methods have default value.
func (dm DetectionMethods) IsDefault() bool {
	return len(dm) == 1 && dm[0] == DetectionMethodArtDupl
}

// Contains checks if method is in methods list.
func (dm DetectionMethods) Contains(method DetectionMethod) bool {
	return slices.Contains(dm, method)
}

// IsEmpty checks if methods list is empty.
func (dm DetectionMethods) IsEmpty() bool {
	return len(dm) == 0
}

// IsHashOnly checks if only hash detection method is selected.
// This is useful for optimizing the analysis pipeline - when only hash detection
// is used, we can skip AST parsing and work directly with file paths.
func (dm DetectionMethods) IsHashOnly() bool {
	return len(dm) == 1 && dm[0] == DetectionMethodHash
}

// Config represents the dupl configuration with strong typing.
//
// Note: Config uses primitive types (int, string, bool) for JSON marshaling
// compatibility. Typed enums (FileType, DetectionMethod, etc.) provide validation.
type Config struct {
	// Threshold sets the minimum token sequence size to consider as duplicate
	Threshold int `json:"threshold,omitempty"`

	// IncludeVendor includes vendor directory in analysis
	IncludeVendor bool `json:"includeVendor,omitempty"`

	// IncludeNodeModules includes node_modules directory in hash-based analysis.
	// By default, node_modules is excluded in hash detection mode to avoid
	// processing large dependency directories. Set to true to include them.
	IncludeNodeModules bool `json:"includeNodeModules,omitempty"`

	// FilesFromStdin reads file paths from stdin when true
	FilesFromStdin bool `json:"filesFromStdin,omitempty"`

	// OutputFormat sets the output format with type safety
	OutputFormat OutputFormat `json:"outputFormat,omitempty"`

	// Verbose enables verbose output
	Verbose bool `json:"verbose,omitempty"`

	// Paths specifies the paths to analyze
	Paths []string `json:"paths,omitempty"`

	// IgnoreFiles specifies file patterns to ignore
	IgnoreFiles []string `json:"ignoreFiles,omitempty"`

	// MaxChildrenSerial sets the maximum children serial for large slices
	MaxChildrenSerial int `json:"maxChildrenSerial,omitempty"`

	// OutputFile specifies the output file (if not stdout)
	OutputFile string `json:"outputFile,omitempty"`

	// SortBy specifies sorting criteria for clone groups
	SortBy SortCriteria `json:"sortBy,omitempty"`

	// DetectionMethods specifies which detection methods to use
	DetectionMethods DetectionMethods `json:"detectionMethods,omitempty"`

	// Profile enables performance profiling output
	Profile bool `json:"profile,omitempty"`

	// Timeout specifies maximum execution time in seconds (0 = no timeout)
	Timeout int `json:"timeout,omitempty"`

	// IncludeSQLC includes sqlc.dev generated files in analysis (default: false, filtered)
	IncludeSQLC bool `json:"includeSQLC,omitempty"`

	// IncludeTempl includes templ-generated *_templ.go files in analysis (default: false, filtered).
	// .templ source files are always included regardless of this setting.
	IncludeTempl bool `json:"includeTempl,omitempty"`

	// IncludeProtobuf includes protobuf generated files (.pb.go, _grpc.pb.go)
	IncludeProtobuf bool `json:"includeProtobuf,omitempty"`

	// IncludeMockgen includes mockgen generated files
	IncludeMockgen bool `json:"includeMockgen,omitempty"`

	// IncludeStringer includes stringer generated files
	IncludeStringer bool `json:"includeStringer,omitempty"`

	// Only restricts analysis to a specific file type (FileTypeGo or FileTypeTempl).
	// FileTypeAll (empty string) means analyze all file types.
	Only FileType `json:"only,omitempty"`

	// IncludePatterns specifies file patterns to always include (takes precedence over filter)
	IncludePatterns []string `json:"includePatterns,omitempty"`

	// ExcludePatterns specifies additional file patterns to exclude
	ExcludePatterns []string `json:"excludePatterns,omitempty"`

	// Incremental enables incremental analysis (only analyze changed files)
	Incremental bool `json:"incremental,omitempty"`

	// Since specifies the git reference for incremental analysis
	// Can be a commit hash, branch name, tag, or relative reference (e.g., "HEAD~1")
	// If empty and Incremental is true, uses HEAD (all uncommitted changes)
	Since string `json:"since,omitempty"`

	// CacheDir specifies the cache directory for AST caching
	// If empty, uses default .cache/art-dupl
	CacheDir string `json:"cacheDir,omitempty"`

	// ClearCache clears the cache before running (useful for forced full rebuild)
	ClearCache bool `json:"clearCache,omitempty"`

	// Semantic enables semantic-aware duplicate detection.
	// When true, identifier names are included in the type hash, reducing false positives
	// from structurally similar but semantically different code.
	//
	// Default is false (structural-only matching). Enable with --semantic flag when
	// you want to reduce false positives from similar-looking but different code.
	//
	// Example: Expect(x).To(Equal(y)) vs Expect(z).To(Equal(w))
	// - Semantic=true: Only matches if method names are the same (e.g., Equal vs Equal)
	// - Semantic=false (default): Matches based on structure only (more potential matches)
	Semantic bool `json:"semantic,omitempty"`

	// Workers specifies the number of concurrent workers for file parsing.
	// 0 or negative means use runtime.GOMAXPROCS(0).
	// 1 means sequential processing (same as Parse()).
	Workers int `json:"workers,omitempty"`

	// DiffMode enables diff visualization for HTML output.
	// When enabled, duplicate occurrences are shown with visual diff highlighting.
	DiffMode DiffMode `json:"diffMode,omitempty"`
}

// DefaultThreshold is the default minimum token sequence size for clone detection.
const DefaultThreshold = 15

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{ //nolint:exhaustruct
		Threshold:          15,
		IncludeVendor:      false,
		IncludeNodeModules: false,
		FilesFromStdin:     false,
		OutputFormat:       OutputFormatText,
		Verbose:            false,
		Paths:              []string{"."},
		IgnoreFiles:        []string{},
		MaxChildrenSerial:  10000,
		OutputFile:         "",
		SortBy:             SortBySize,
		DetectionMethods:   DetectionMethods{DetectionMethodArtDupl},
		Profile:            false,
		Timeout:            0,
		IncludeSQLC:        false,
		IncludeTempl:       false,
		IncludeProtobuf:    false,
		IncludeMockgen:     false,
		IncludeStringer:    false,
		Only:               "",
		IncludePatterns:    []string{},
		ExcludePatterns:    []string{},
		Incremental:        false,
		Since:              "",
		CacheDir:           "",
		ClearCache:         false,
		Semantic:           false,
		DiffMode:           DiffModeDisabled,
	}
}
