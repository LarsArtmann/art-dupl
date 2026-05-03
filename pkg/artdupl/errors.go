package artdupl

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/config"
)

// SDK-specific sentinel errors for simple error comparison via errors.Is().
//
// Design Note: These errors are intentionally separate from the internal errors
// package (errors/types.go) which provides rich error types with context.
// SDK users need simple sentinel errors for comparison, while internal code
// uses rich errors for debugging. This dual approach is by design:
//   - SDK errors: Simple sentinel errors for user-facing error comparison
//   - Internal errors: Rich DuplError types with file, line, stack trace
var (
	// ErrNilOptions is returned when options are nil.
	ErrNilOptions = errors.New("options cannot be nil")

	// ErrInvalidThreshold is returned when threshold is invalid.
	ErrInvalidThreshold = config.ErrInvalidThreshold

	// ErrThresholdTooLarge is returned when threshold exceeds maximum.
	ErrThresholdTooLarge = config.ErrThresholdTooLarge

	// ErrNoDetectionMethods is returned when no detection methods are specified.
	ErrNoDetectionMethods = errors.New("at least one detection method must be specified")

	// ErrInvalidMaxFileSize is returned when max file size is invalid.
	ErrInvalidMaxFileSize = errors.New("max file size must be >= 0")

	// ErrInvalidMaxWorkers is returned when max workers is invalid.
	ErrInvalidMaxWorkers = errors.New("max workers must be >= 1")

	// ErrInvalidTimeout is returned when timeout is invalid.
	ErrInvalidTimeout = errors.New("timeout must be >= 0")

	// ErrUnsupportedMethod is returned when detection method is not supported.
	ErrUnsupportedMethod = errors.New("unsupported detection method")

	// ErrNoFilesProvided is returned when no files are provided for analysis.
	ErrNoFilesProvided = errors.New("no files provided for analysis")

	// ErrFileNotFound is returned when file does not exist.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileTooLarge is returned when file exceeds size limit.
	ErrFileTooLarge = errors.New("file size exceeds maximum limit")

	// ErrParsingFailed is returned when file parsing fails.
	ErrParsingFailed = errors.New("failed to parse file")

	// ErrContextCanceled is returned when context is canceled.
	ErrContextCanceled = errors.New("analysis canceled")

	// ErrAnalysisTimeout is returned when analysis times out.
	ErrAnalysisTimeout = errors.New("analysis timed out")

	// ErrNoDuplicatesFound is returned when no duplicates are detected.
	ErrNoDuplicatesFound = errors.New("no duplicates found")

	// ErrResultProcessing is returned when result processing fails.
	ErrResultProcessing = errors.New("error processing results")

	// ErrCloneEndLineBeforeStart is returned when end line is before start line.
	ErrCloneEndLineBeforeStart = errors.New("clone end line is before start line")

	// ErrCloneZeroLength is returned when clone has zero length.
	ErrCloneZeroLength = errors.New("clone has zero length (start >= end)")

	// ErrCloneInvalidPosition is returned when clone has invalid byte positions.
	ErrCloneInvalidPosition = errors.New("clone has invalid byte positions")

	// ErrMemoryLimit is returned when memory limit is exceeded.
	ErrMemoryLimit = errors.New("memory limit exceeded")

	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal error")
)
