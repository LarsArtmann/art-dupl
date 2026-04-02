package artdupl

import "errors"

// SDK-specific sentinel errors for simple error comparison via errors.Is().
//
// Design Note: These errors are intentionally separate from the internal errors
// package (errors/types.go) which provides rich error types with context.
// SDK users need simple sentinel errors for comparison, while internal code
// uses rich errors for debugging. This dual approach is by design:
//   - SDK errors: Simple sentinel errors for user-facing error comparison
//   - Internal errors: Rich DuplError types with file, line, stack trace
var (
	// Configuration errors.
	ErrNilOptions         = errors.New("options cannot be nil")
	ErrInvalidThreshold   = errors.New("threshold must be >= 1")
	ErrThresholdTooLarge  = errors.New("threshold too large (max 1000)")
	ErrNoDetectionMethods = errors.New("at least one detection method must be specified")
	ErrInvalidMaxFileSize = errors.New("max file size must be >= 0")
	ErrInvalidMaxWorkers  = errors.New("max workers must be >= 1")
	ErrInvalidTimeout     = errors.New("timeout must be >= 0")
	ErrUnsupportedMethod  = errors.New("unsupported detection method")

	// Analysis errors.
	ErrNoFilesProvided = errors.New("no files provided for analysis")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileTooLarge    = errors.New("file size exceeds maximum limit")
	ErrParsingFailed   = errors.New("failed to parse file")
	ErrContextCanceled = errors.New("analysis canceled")
	ErrAnalysisTimeout = errors.New("analysis timed out")

	// Result errors.
	ErrNoDuplicatesFound = errors.New("no duplicates found")
	ErrResultProcessing  = errors.New("error processing results")

	// Clone validation errors.
	ErrCloneEndLineBeforeStart = errors.New("clone end line is before start line")
	ErrCloneZeroLength         = errors.New("clone has zero length (start >= end)")
	ErrCloneInvalidPosition    = errors.New("clone has invalid byte positions")

	// System errors.
	ErrMemoryLimit = errors.New("memory limit exceeded")
	ErrInternal    = errors.New("internal error")
)
