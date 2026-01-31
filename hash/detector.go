// Package hash provides hash-based duplicate detection.
//
// This package implements file-level hash-based detection, which is
// complementary to syntax-level suffix tree detection.
//
// Detection Strategy:
// - Hash entire file content using rolling hash algorithm
// - Compare file hashes to find exact duplicates
// - Faster than syntax detection but less granular
// - Good for detecting file-level code copying
//
// Core Types:
// - HashDetector: Main detector using file-level hashing
// - FileDetector: Low-level file detection algorithm
//
// Comparison to Syntax Detection:
// - Syntax: Finds duplicate fragments within files
// - Hash: Finds exact duplicate files
// - Combined: Use both for comprehensive coverage
//
// Usage:
//	detector := hash.NewHashDetector(threshold)
//	matches := detector.FindDuplOver(syntaxNodes, threshold)
//	for match := range matches {
//	    // Process duplicate matches
//	}
//
// Performance:
// - O(n) file comparison where n = number of files
// - O(1) hash computation per file
// - Very fast for large codebases
//
// Limitations:
// - Only detects exact file duplicates
// - Cannot detect duplicate code fragments within files
// - Use with syntax detection for comprehensive coverage
//
package hash

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// HashDetector implements file-level hash-based duplicate detection.
type HashDetector struct {
	*FileDetector // Embeds the file detector
}

// NewHashDetector creates a new hash detector using file-level hashing.
func NewHashDetector(threshold int) *HashDetector {
	return &HashDetector{
		FileDetector: NewFileDetector(threshold),
	}
}

// FindDuplOver delegates to FileDetector for exact file duplicate detection.
func (h *HashDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	return h.FileDetector.FindDuplOver(data, threshold)
}
