package detection

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// MethodDetector is the interface that all detection algorithms implement.
// Each detector produces a stream of syntax.Match results for duplicate code.
type MethodDetector interface {
	// FindDuplOver runs detection with the given threshold and returns
	// matches via a channel for streaming consumption.
	FindDuplOver(threshold int) <-chan syntax.Match
}
