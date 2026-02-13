package artdupl

import (
	"context"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

// DetectionMethod is an alias to config.DetectionMethod for convenience.
// This ensures type consistency across the codebase.
//
// Note: This is now fully unified with config package - no duplicate types.
type DetectionMethod = config.DetectionMethod

const (
	// MethodArtDupl uses suffix tree algorithm on AST tokens.
	MethodArtDupl = config.DetectionMethodArtDupl

	// MethodHash uses rolling hash on file content.
	MethodHash = config.DetectionMethodHash

	// MethodTodos finds TODO comments.
	MethodTodos = config.DetectionMethodTodos

	// MethodLegacy finds legacy code patterns.
	MethodLegacy = config.DetectionMethodLegacy
)

// Detector is the main interface for code duplication detection.
type Detector interface {
	// FindClones performs duplication analysis and returns complete results
	FindClones(ctx context.Context, files []string) (*Result, error)

	// FindClonesStream provides streaming results for large projects
	FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error)

	// Close releases any resources held by the detector
	Close() error
}

// Result contains all detected duplicates with metadata and statistics.
type Result struct {
	CloneGroups []*CloneGroup `json:"clone_groups"`
	Summary     *Summary      `json:"summary"`
	Metadata    *Metadata     `json:"metadata"`
}

// CloneGroup represents a group of identical or similar code fragments.
type CloneGroup struct {
	Hash      string          `json:"hash"`             // Unique identifier for this clone group
	Clones    []*Clone        `json:"clones"`           // All occurrences of this clone
	Size      int             `json:"size"`             // Size in tokens/bytes
	LineCount int             `json:"line_count"`       // Number of lines
	Method    DetectionMethod `json:"detection_method"` // Method that found this group
}

// Clone represents a single occurrence of duplicated code.
type Clone struct {
	Filename  string `json:"filename"`           // File containing this clone
	StartLine int    `json:"start_line"`         // Starting line number
	EndLine   int    `json:"end_line"`           // Ending line number
	StartPos  int    `json:"start_pos"`          // Starting byte position
	EndPos    int    `json:"end_pos"`            // Ending byte position
	Fragment  string `json:"fragment,omitempty"` // Actual code content (optional)
	Size      int    `json:"size"`               // Size in bytes/tokens
}

// Summary provides statistics and metadata about the analysis.
type Summary struct {
	TotalFiles    int               `json:"total_files"`
	TotalClones   int               `json:"total_clones"`
	TotalGroups   int               `json:"total_groups"`
	AnalysisTime  time.Duration     `json:"analysis_time_ms"`
	MethodsUsed   []DetectionMethod `json:"methods_used"`
	LinesAnalyzed int               `json:"lines_analyzed"`
}

// Metadata contains additional information about the analysis.
type Metadata struct {
	Version    string    `json:"version"`     // SDK version
	Timestamp  time.Time `json:"timestamp"`   // When analysis was performed
	ConfigHash string    `json:"config_hash"` // Hash of configuration used
	Toolchain  string    `json:"toolchain"`   // Go version, etc.
}

// Options configures the detector behavior.
type Options struct {
	// Detection settings
	Threshold        int               `json:"threshold"`         // Minimum size to consider as clone
	DetectionMethods []DetectionMethod `json:"detection_methods"` // Methods to use for detection

	// File processing
	IncludeVendor bool     `json:"include_vendor"` // Include vendor directory
	IgnoreFiles   []string `json:"ignore_files"`   // File patterns to ignore
	MaxFileSize   int64    `json:"max_file_size"`  // Maximum file size to process

	// Performance tuning
	MaxWorkers int           `json:"max_workers"` // Maximum concurrent workers
	Timeout    time.Duration `json:"timeout"`     // Maximum analysis time

	// Output customization
	IncludeFragments  bool `json:"include_fragments"`    // Include actual code fragments
	MaxClonesPerGroup int  `json:"max_clones_per_group"` // Limit clones per group

	// Callbacks and customization
	ProgressCallback func(*Progress) error `json:"-"` // Progress reporting callback
	FileReader       FileReaderFunc        `json:"-"` // Custom file reader function
	Logger           Logger                `json:"-"` // Custom logger
}

// Progress reports analysis progress.
type Progress struct {
	Stage       string  `json:"stage"`                  // Current processing stage
	Completed   int     `json:"completed"`              // Number of items completed
	Total       int     `json:"total"`                  // Total number of items
	Percentage  float64 `json:"percentage"`             // Completion percentage
	Message     string  `json:"message"`                // Human-readable progress message
	CurrentFile string  `json:"current_file,omitempty"` // Currently processing file
}

// FileReaderFunc represents a function that can read file contents.
type FileReaderFunc func(filename string) ([]byte, error)

// Logger is an alias to the logger package's Logger interface.
type Logger = logger.Logger

// DefaultOptions returns a configuration with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Threshold:         15,
		DetectionMethods:  []DetectionMethod{MethodArtDupl},
		IncludeVendor:     false,
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		MaxWorkers:        4,
		Timeout:           30 * time.Minute,
		IncludeFragments:  false,
		MaxClonesPerGroup: 50,
		ProgressCallback:  nil,
		FileReader:        readFileDefault,
		Logger:            logger.Default,
	}
}

// readFileDefault is the default file reader using os package.
var readFileDefault = func(filename string) ([]byte, error) { //nolint:gochecknoglobals // Default implementation for config
	// This will be implemented with actual file reading
	return nil, nil
}

// ValidateOptions checks if the provided options are valid.
func ValidateOptions(opts *Options) error {
	if opts == nil {
		return ErrNilOptions
	}

	if opts.Threshold < 1 {
		return ErrInvalidThreshold
	}

	if opts.Threshold > 1000 {
		return ErrThresholdTooLarge
	}

	if len(opts.DetectionMethods) == 0 {
		return ErrNoDetectionMethods
	}

	if opts.MaxFileSize < 0 {
		return ErrInvalidMaxFileSize
	}

	if opts.MaxWorkers < 1 {
		return ErrInvalidMaxWorkers
	}

	if opts.Timeout < 0 {
		return ErrInvalidTimeout
	}

	return config.ValidateDetectionMethods(opts.DetectionMethods) //nolint:wrapcheck // Pass through validation error
}
