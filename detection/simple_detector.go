package detection

//
// SIMPLE DETECTOR INTERFACE
//
// This is a minimal, non-breaking interface for detection methods.
// It provides type safety where possible without requiring changes
// to existing detectors.
//
// DESIGN:
// - Minimal interface for easy adoption
// - Non-breaking (existing detectors don't need to implement it)
// - Incremental adoption (new detectors can implement it)
// - Type-safe where possible (domain.Threshold)
//
// FUTURE:
// - Expand interface as more detectors are added
// - Add common methods (Name(), Close(), etc.) as needed
// - Consider adapter pattern for existing detectors
//
// USAGE:
// For new detectors:
//
//	type MyDetector struct {
//	    detection.SimpleDetector
//	}
//
//	func (d *MyDetector) FindDuplOver(threshold int) <-chan syntax.Match {
//	    // Implementation
//	}
//
// For existing detectors:
// - Continue as-is (no changes required)
// - Optionally implement interface if needed for type safety
//
// LIMITATIONS:
// - Domain types not yet used (int threshold)
// - No Name() or Close() methods yet
// - Not all detectors implement this interface yet
//
// RECOMMENDATION:
// Use this interface for NEW detectors only.
// Don't refactor existing detectors to implement this yet
// Expand interface incrementally as needed.

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SimpleDetector is a minimal interface for clone detection.
//
// This interface matches the existing signature used by MultiDetector
// and HashDetector, allowing them to be used interchangeably
// without requiring changes to existing code.
//
// Methods:
// - FindDuplOver(threshold int) <-chan syntax.Match
//   Finds all clones/sequences with size >= threshold
//   Returns channel for streaming results
//
// Usage:
//	// Can use any detector implementing this interface
//	var detector SimpleDetector
//	if useMultiDetector {
//	    detector = NewMultiDetector(...)
//	} else if useHashDetector {
//	    detector = NewHashDetector(...)
//	}
//
//	matches := detector.FindDuplOver(threshold)
//	for match := range matches {
//	    // Process matches
//	}
type SimpleDetector interface {
	// FindDuplOver finds all clones/sequences with size >= threshold.
	//
	// Parameters:
	// - threshold: Minimum size in tokens to consider as duplicate
	//
	// Returns:
	// - <-chan syntax.Match: Channel for streaming results
	//
	// Behavior:
	// - Channel is closed when all matches are sent
	// - Results are sent as they are found (non-blocking)
	// - Caller should range over channel to receive all matches
	FindDuplOver(threshold int) <-chan syntax.Match
}
