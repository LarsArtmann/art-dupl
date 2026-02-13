package domain

import (
	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/zeebo/xxh3"
)

// NodeToClone converts syntax nodes to domain Clone.
func NodeToClone(node *syntax.Node, filename string, fileContent []byte) Clone {
	lineStart, lineEnd := 1, 1
	if fileContent != nil {
		lineStart, lineEnd = position.ByteRangeToLines(fileContent, int(node.Pos), int(node.End))
	}

	var fragment string
	if fileContent != nil {
		start := node.Pos
		end := node.End
		if start >= 0 && int(end) <= len(fileContent) && start < end {
			fragment = string(fileContent[start:end])
		}
	}

	hashStr := ""
	if fragment != "" {
		// PERFORMANCE: XXH3 is ~20x faster than crypto/sha256 and includes
		// native ARM64 NEON SIMD optimizations. DO NOT replace with cryptographic
		// hash functions (SHA-256, etc.) - this hash is for deduplication only,
		// not security. See: https://github.com/zeebo/xxh3
		//
		//nolint:gosec // G401,G505: Intentionally using non-cryptographic hash for performance
		hashStr = formatDomainHash(xxh3.Hash([]byte(fragment)))
	}

	startLn, _ := NewLineNumber(uint16(lineStart)) //nolint:gosec //G115 lineStart >= 1 guaranteed by initialize default
	endLn, _ := NewLineNumber(uint16(lineEnd))     //nolint:gosec //G115 lineEnd >= 1 guaranteed by initialize default
	startPos := NewBytePosition(uint32(node.Pos))  //nolint:gosec //G115 node.Pos validated >= 0 in fragment extraction
	endPos := NewBytePosition(uint32(node.End))    //nolint:gosec //G115 node.End validated >= 0 in fragment extraction
	complexity := NewComplexityScore(uint16(calculateComplexity(node)))

	clone := Clone{
		StartLine:  startLn,
		EndLine:    endLn,
		StartPos:   startPos,
		EndPos:     endPos,
		Complexity: complexity,
		Status:     FileProcessingStateCompleted,
	}

	clone.SetFilename(filename)
	clone.SetFragment(fragment)
	clone.SetHash(hashStr)

	return clone
}

// CalculateSeverity determines clone severity based on size and complexity.
func CalculateSeverity(size, complexity uint) CloneSeverity {
	if complexity > 50 || size > 200 {
		return CloneSeverityCritical
	}
	if complexity > 20 || size > 100 {
		return CloneSeverityHigh
	}
	if complexity > 10 || size > 50 {
		return CloneSeverityMedium
	}
	return CloneSeverityLow
}

func calculateComplexity(node *syntax.Node) uint {
	complexity := uint(1)

	for _, child := range node.Children {
		complexity += calculateComplexity(child)
	}

	switch node.Type {
	case 0:
		complexity += 2
	default:
		complexity += 1
	}

	return complexity
}

// formatDomainHash converts a uint64 hash to a hex string.
// This is faster than fmt.Sprintf or encoding/hex for fixed-size uint64.
func formatDomainHash(h uint64) string {
	const hexchars = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		buf[i] = hexchars[h&0xf]
		h >>= 4
	}
	return string(buf)
}
