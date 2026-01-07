// Package position provides utilities for working with source code positions.
package position

// ByteRangeToLines converts byte positions to line numbers.
// Returns (startLine, endLine) where both are 1-indexed.
// Handles edge cases where positions are at file boundaries.
func ByteRangeToLines(content []byte, start, end int) (int, int) {
	if len(content) == 0 {
		return 1, 1
	}

	line := 1
	lineStart, lineEnd := 0, 0

	for offset := 0; offset < len(content); offset++ {
		if content[offset] == '\n' {
			line++
		}
		if offset == start {
			lineStart = line
		}
		if offset == end-1 {
			lineEnd = line
			break
		}
	}

	// Default values if positions were not found
	if lineStart == 0 {
		lineStart = 1
	}
	if lineEnd == 0 {
		lineEnd = lineStart
	}

	return lineStart, lineEnd
}
