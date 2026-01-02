package artdupl

import "errors"

// SDK-specific error types.
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

// validateDetectionMethods checks if all detection methods are supported.
func validateDetectionMethods(methods []DetectionMethod) error {
	if len(methods) == 0 {
		return ErrNoDetectionMethods
	}

	for _, method := range methods {
		switch method {
		case MethodArtDupl, MethodHash, MethodAll:
			// Valid methods
		default:
			return ErrUnsupportedMethod
		}
	}

	return nil
}
