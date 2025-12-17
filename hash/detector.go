package hash

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// HashDetector implements file-level hash-based duplicate detection
type HashDetector struct {
	*FileDetector // Embeds the file detector
}

// NewHashDetector creates a new hash detector using file-level hashing
func NewHashDetector(threshold int) *HashDetector {
	return &HashDetector{
		FileDetector: NewFileDetector(threshold),
	}
}

// FindDuplOver delegates to FileDetector for exact file duplicate detection
func (h *HashDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	return h.FileDetector.FindDuplOver(data, threshold)
}
