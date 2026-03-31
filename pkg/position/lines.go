// Package position provides utilities for working with source code positions.
package position

import (
	"sort"
	"strings"
)

// ByteRangeToLines converts byte positions to line numbers.
// Returns (startLine, endLine) where both are 1-indexed.
// Handles edge cases where positions are at file boundaries.
func ByteRangeToLines(content []byte, start, end int) (int, int) {
	if len(content) == 0 {
		return 1, 1
	}

	// Handle same position (single point) - both start and end are the same
	if start == end {
		return offsetToLine(content, start), offsetToLine(content, end)
	}

	line := 1
	lineStart, lineEnd := 0, 0

	for offset := range content {
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

// offsetToLine returns the 1-based line number for a byte offset.
// Follows the same line-counting semantics as ByteRangeToLines:
// a position at a newline character returns the line that the newline terminates.
func offsetToLine(content []byte, offset int) int {
	if offset < 0 || len(content) == 0 {
		return 1
	}

	// Cap offset to content bounds
	if offset >= len(content) {
		offset = len(content) - 1
	}

	line := 1
	for i := 0; i <= offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}

	return line
}

// SplitLines splits content into lines by newline characters.
func SplitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}

	return strings.Split(string(content), "\n")
}

// JoinLines joins lines with newline characters.
func JoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

// LineIndex helps convert byte offsets to line numbers efficiently using binary search.
// It pre-indexes all newline positions for O(log n) line number lookups.
type LineIndex struct {
	newlines []int
}

// NewLineIndex creates an index from file content.
// It records the byte offset of each newline character.
func NewLineIndex(content []byte) *LineIndex {
	newlines := make([]int, 0, len(content)/40) //nolint:mnd // estimate 40 chars per line
	newlines = append(newlines, 0)              // Line 1 starts at 0

	for i, b := range content {
		if b == '\n' {
			newlines = append(newlines, i+1)
		}
	}

	return &LineIndex{newlines: newlines}
}

// Line returns the 1-based line number for a byte offset.
// Uses binary search for O(log n) performance.
// If offset is out of bounds, returns the last line number.
func (li *LineIndex) Line(offset int) int {
	// Find the first newline index that is > offset.
	// The line number is the index of that newline in our array.
	// We want the index i such that newlines[i] <= offset < newlines[i+1]
	idx := sort.Search(len(li.newlines), func(i int) bool {
		return li.newlines[i] > offset
	})

	return idx
}
