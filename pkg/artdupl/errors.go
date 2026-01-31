package artdupl

import "errors"

// SDK-specific error types.
//
// TODO: SPLIT-BRAIN ALERT! These errors duplicate errors package functionality.
// The errors package (errors/types.go) already provides rich error types with
// context (DuplError, EnumValidationError, etc.).
//
// We have TWO error handling strategies:
// 1. errors package: Rich errors with type, message, file, line, cause, stack
// 2. artdupl/errors.go: Simple errors.New() errors
//
// This creates inconsistency in error handling across the codebase.
// Options:
// - Option A: Use errors package everywhere, remove these
// - Option B: Keep these as SDK-specific but wrap with errors package
// - Option C: Define in domain package as domain errors
//
// Recommendation: Option A - consolidate on errors package for consistency.
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

	// System errors.
	ErrMemoryLimit = errors.New("memory limit exceeded")
	ErrInternal    = errors.New("internal error")
)
