package artdupl

import (
	"context"
	"fmt"
	"os"
	"time"
)

// DetectionMethod represents the method used for duplicate detection.
// This is an independent SDK type — conversion to internal config types
// happens at the SDK boundary (convertOptionsToConfig).
type DetectionMethod string

const (
	// MethodArtDupl uses suffix tree algorithm on AST tokens.
	MethodArtDupl DetectionMethod = "art-dupl"

	// MethodHash uses rolling hash on file content.
	MethodHash DetectionMethod = "hash"
)

// String returns the string representation of the detection method.
func (m DetectionMethod) String() string { return string(m) }

// IsValid returns true if the detection method is one of the defined constants.
func (m DetectionMethod) IsValid() bool {
	switch m {
	case MethodArtDupl, MethodHash:
		return true
	default:
		return false
	}
}

// Detector is the main interface for code duplication detection.
type Detector interface {
	// FindClones performs duplication analysis and returns complete results
	FindClones(ctx context.Context, files []string) (*Result, error)

	// FindClonesStream provides streaming results for large projects
	FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error)

	// FindClonesStreamResult provides streaming results with error propagation.
	// The final StreamResult will have Err set if the pipeline failed.
	FindClonesStreamResult(ctx context.Context, files []string) (<-chan StreamResult, error)

	// Close releases any resources held by the detector
	Close() error
}

// StreamResult wraps a CloneGroup with an optional error for streaming.
// When Err is non-nil, the stream has failed and no more results will follow.
type StreamResult struct {
	Group *CloneGroup
	Err   error
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
	LineStart int    `json:"line_start"`         // Starting line number
	LineEnd   int    `json:"line_end"`           // Ending line number
	StartPos  int    `json:"start_pos"`          // Starting byte position
	EndPos    int    `json:"end_pos"`            // Ending byte position
	Fragment  string `json:"fragment,omitempty"` // Actual code content (optional)
	Size      int    `json:"size"`               // Size in bytes/tokens
}

// IsValid validates the clone data and returns an error if invalid.
// A valid clone must have LineEnd >= LineStart and EndPos > StartPos.
func (c Clone) IsValid() error {
	if c.LineEnd < c.LineStart {
		return ErrCloneLineEndBeforeStart
	}

	if c.StartPos >= c.EndPos {
		return ErrCloneZeroLength
	}

	return nil
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
	Verbose          bool              `json:"verbose"`           // Enable verbose detection logging

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

// Logger is the SDK's own logging interface.
// Any type implementing these methods satisfies the interface — the internal
// pkg/logger.Logger is structurally compatible without an explicit adapter.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// noOpLogger is the default SDK logger when none is provided.
type noOpLogger struct{}

func (noOpLogger) Debug(string, ...any) {}
func (noOpLogger) Info(string, ...any)  {}
func (noOpLogger) Warn(string, ...any)  {}
func (noOpLogger) Error(string, ...any) {}

// DefaultOptions returns a configuration with sensible defaults.
func DefaultOptions() *Options {
	return &Options{ //nolint:exhaustruct
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
		Logger:            noOpLogger{},
	}
}

// readFileDefault is the default file reader using os package.
var readFileDefault = os.ReadFile

// ValidateOptions checks if the provided options are valid.
func ValidateOptions(opts *Options) error {
	if opts == nil {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrNilOptions)
	}

	if opts.Threshold < 1 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrInvalidThreshold)
	}

	if opts.Threshold > 1000 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrThresholdTooLarge)
	}

	if len(opts.DetectionMethods) == 0 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrNoDetectionMethods)
	}

	if opts.MaxFileSize < 0 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrInvalidMaxFileSize)
	}

	if opts.MaxWorkers < 1 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrInvalidMaxWorkers)
	}

	if opts.Timeout < 0 {
		return fmt.Errorf("validate options failed (opts=%v): %w", opts, ErrInvalidTimeout)
	}

	// Validate detection methods using SDK's own type — no config dependency
	for _, method := range opts.DetectionMethods {
		if !method.IsValid() {
			return fmt.Errorf("validate options failed (opts=%v): %w: %s", opts, ErrUnsupportedMethod, method)
		}
	}

	return nil
}
