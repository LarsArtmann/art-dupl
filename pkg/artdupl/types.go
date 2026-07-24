package artdupl

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

// DetectionMethod is an alias for domain.DetectionMethod, the canonical
// definition shared across config, SDK, and detection packages.
type DetectionMethod = domain.DetectionMethod

const (
	// MethodArtDupl uses suffix tree algorithm on AST tokens.
	MethodArtDupl = domain.MethodArtDupl

	// MethodHash uses rolling hash on file content.
	MethodHash = domain.MethodHash

	// DefaultThreshold is the default minimum number of duplicated statements to report.
	// Mirrors config.DefaultThreshold (the SDK cannot import config/ due to arch-lint).
	// A threshold of 5 filters trivial patterns (single-call noise, error-check
	// boilerplate) while catching meaningful duplication.
	DefaultThreshold = 5
)

// Detector is the main interface for code duplication detection.
type Detector interface {
	// FindClones performs duplication analysis and returns complete results
	FindClones(ctx context.Context, files []string) (*Result, error)

	// FindClonesStreamResult provides streaming results with error propagation.
	// The channel emits StreamResult values. A final StreamResult with Err != nil
	// indicates pipeline failure. The channel is always closed after all results.
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
	domain.CloneRef

	StartPos int `json:"start_pos"` // Starting byte position
	EndPos   int `json:"end_pos"`   // Ending byte position
	Size     int `json:"size"`      // Size in bytes/tokens
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
	AnalysisTime  time.Duration     `json:"-"`
	MethodsUsed   []DetectionMethod `json:"methods_used"`
	LinesAnalyzed int               `json:"lines_analyzed"`
}

// MarshalJSON emits AnalysisTime as milliseconds (matching the JSON contract)
// rather than the default time.Duration serialization (nanoseconds).
func (s Summary) MarshalJSON() ([]byte, error) {
	type alias Summary

	data, err := json.Marshal(struct {
		alias

		AnalysisTimeMS int64 `json:"analysis_time_ms"`
	}{
		alias:          alias(s),
		AnalysisTimeMS: s.AnalysisTime.Milliseconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal summary: %w", err)
	}

	return data, nil
}

// UnmarshalJSON reads AnalysisTime from milliseconds back into time.Duration.
func (s *Summary) UnmarshalJSON(data []byte) error {
	type alias Summary

	aux := struct {
		alias

		AnalysisTimeMS int64 `json:"analysis_time_ms"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unmarshal summary: %w", err)
	}

	*s = Summary(aux.alias)
	s.AnalysisTime = time.Duration(aux.AnalysisTimeMS) * time.Millisecond

	return nil
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
	TypeAware        bool              `json:"type_aware"`        // Enable go/types-based detection (10-100x slower)

	// File processing — the SDK takes explicit file lists, so callers
	// filter vendor/test directories themselves before calling FindClones.
	IgnoreFiles []string `json:"ignore_files"`  // File patterns to ignore (filepath.Match)
	MaxFileSize int64    `json:"max_file_size"` // Maximum file size to process

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

// FileReaderFunc is an alias for domain.FileReaderFunc, the canonical
// file-reader function type shared across the codebase.
type FileReaderFunc = domain.FileReaderFunc

// Logger is an alias for logger.Logger, the canonical logging interface.
// This eliminates the split-brain between the SDK's own Logger interface
// and pkg/logger's implementations.
type Logger = logger.Logger

// DefaultOptions returns a configuration with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Threshold:         DefaultThreshold,
		DetectionMethods:  []DetectionMethod{MethodArtDupl},
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		MaxWorkers:        4,
		Timeout:           30 * time.Minute,
		IncludeFragments:  false,
		MaxClonesPerGroup: 50,
		ProgressCallback:  nil,
		FileReader:        readFileDefault,
		Logger:            &logger.NoOpLogger{},
	}
}

// readFileDefault is the default file reader using os package.
var readFileDefault = os.ReadFile //nolint:gochecknoglobals // overridable default file reader

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
