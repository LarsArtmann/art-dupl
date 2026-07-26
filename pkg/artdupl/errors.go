package artdupl

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/domain"
)

// SDK-specific sentinel errors for simple error comparison via errors.Is().
//
// ErrInvalidThreshold and ErrThresholdTooLarge are re-exported from domain
// so that errors.Is() works across config, SDK, and domain boundaries.
// All other sentinels are SDK-specific.
var (
	// ErrNilOptions is returned when options are nil.
	ErrNilOptions = errors.New("options cannot be nil")

	// ErrInvalidThreshold is returned when threshold is invalid.
	// Re-exported from domain for cross-package errors.Is() compatibility.
	ErrInvalidThreshold = domain.ErrInvalidThreshold

	// ErrThresholdTooLarge is returned when threshold exceeds maximum.
	// Re-exported from domain for cross-package errors.Is() compatibility.
	ErrThresholdTooLarge = domain.ErrThresholdTooLarge

	// ErrNoDetectionMethods is returned when no detection methods are specified.
	ErrNoDetectionMethods = errors.New("at least one detection method must be specified")

	// ErrInvalidMaxFileSize is returned when max file size is invalid.
	ErrInvalidMaxFileSize = errors.New("max file size must be >= 0")

	// ErrInvalidMaxWorkers is returned when max workers is invalid.
	ErrInvalidMaxWorkers = errors.New("max workers must be >= 1")

	// ErrInvalidTimeout is returned when timeout is invalid.
	ErrInvalidTimeout = errors.New("timeout must be >= 0")

	// ErrUnsupportedMethod is returned when detection method is not supported.
	// Aliased to domain.ErrInvalidDetectionMethod so errors.Is works across SDK/domain boundaries.
	ErrUnsupportedMethod = domain.ErrInvalidDetectionMethod

	// ErrNoFilesProvided is returned when no files are provided for analysis.
	ErrNoFilesProvided = errors.New("no files provided for analysis")

	// ErrFileNotFound is returned when file does not exist.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileTooLarge is returned when file exceeds size limit.
	ErrFileTooLarge = errors.New("file size exceeds maximum limit")

	// ErrFileIgnored is returned when a file matches an ignore pattern.
	ErrFileIgnored = errors.New("file matched an ignore pattern")

	// ErrAnalysisTimeout is returned when analysis times out.
	ErrAnalysisTimeout = errors.New("analysis timed out")

	// ErrNoDuplicatesFound is returned when no duplicates are detected.
	ErrNoDuplicatesFound = errors.New("no duplicates found")

	// ErrCloneLineEndBeforeStart is returned when end line is before start line.
	// Aliased to domain.ErrLineEndBeforeStart so errors.Is works across SDK/domain boundaries.
	ErrCloneLineEndBeforeStart = domain.ErrLineEndBeforeStart

	// ErrCloneZeroLength is returned when clone has zero length.
	ErrCloneZeroLength = errors.New("clone has zero length (start >= end)")
)
