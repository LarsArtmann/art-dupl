package printer

import (
	"bytes"
)

// DiffLine represents a single line in a diff view.
type DiffLine struct {
	Content    string
	Type       DiffLineType
	LineNumber int // Line number in the original file
}

// DiffLineType indicates the type of diff line.
type DiffLineType int

const (
	DiffLineEqual DiffLineType = iota
	DiffLineAdded
	DiffLineRemoved
	DiffLineModified
)

// DiffResult represents the diff between two code fragments.
type DiffResult struct {
	Base     []DiffLine
	Compared []DiffLine
	HasDiff  bool
}

// LineDiff performs a line-by-line diff between two code fragments.
// It uses a simple but effective algorithm optimized for code comparison:
// 1. Split into lines
// 2. Find matching lines
// 3. Mark additions/removals/modifications
//
// Size efficiency: O(n) memory, minimal allocations.
func LineDiff(base, compared []byte) DiffResult {
	baseLines := splitLines(base)
	comparedLines := splitLines(compared)

	result := DiffResult{
		Base:     make([]DiffLine, len(baseLines)),
		Compared: make([]DiffLine, len(comparedLines)),
	}

	// Initialize line numbers and content for base
	for i, line := range baseLines {
		result.Base[i] = DiffLine{
			Content:    string(line),
			Type:       DiffLineEqual,
			LineNumber: i + 1,
		}
	}

	// Initialize line numbers and content for compared
	for i, line := range comparedLines {
		result.Compared[i] = DiffLine{
			Content:    string(line),
			Type:       DiffLineEqual,
			LineNumber: i + 1,
		}
	}

	// Simple line-by-line comparison for same-length sequences
	if len(baseLines) == len(comparedLines) {
		result.HasDiff = diffSameLength(result.Base, result.Compared)
	} else {
		result.HasDiff = diffDifferentLength(baseLines, comparedLines, result.Base, result.Compared)
	}

	return result
}

// splitLines splits bytes into lines, preserving line endings.
func splitLines(data []byte) [][]byte {
	if len(data) == 0 {
		return [][]byte{{}}
	}

	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, data[start:i+1])
			start = i + 1
		}
	}

	// Handle last line if it doesn't end with newline
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}

// diffSameLength compares lines when both fragments have the same number of lines.
func diffSameLength(base, compared []DiffLine) bool {
	hasDiff := false

	for i := range base {
		baseTrimmed := bytes.TrimSpace([]byte(base[i].Content))
		comparedTrimmed := bytes.TrimSpace([]byte(compared[i].Content))

		if !bytes.Equal(baseTrimmed, comparedTrimmed) {
			base[i].Type = DiffLineModified
			compared[i].Type = DiffLineModified
			hasDiff = true
		}
	}

	return hasDiff
}

// diffDifferentLength handles fragments with different line counts.
// Uses a simple LCS (Longest Common Subsequence) approach for small files,
// falls back to line-by-line for performance with large files.
func diffDifferentLength(baseLines, comparedLines [][]byte, base, compared []DiffLine) bool {
	// For performance, use simple heuristic for large files (>100 lines)
	if len(baseLines) > 100 || len(comparedLines) > 100 {
		return diffLargeFiles(baseLines, comparedLines, base, compared)
	}

	return diffLCS(baseLines, comparedLines, base, compared)
}

// diffLargeFiles uses a faster heuristic for large files.
func diffLargeFiles(baseLines, comparedLines [][]byte, base, compared []DiffLine) bool {
	hasDiff := false
	minLen := len(baseLines)
	if len(comparedLines) < minLen {
		minLen = len(comparedLines)
	}

	// Compare line by line up to the shorter length
	for i := 0; i < minLen; i++ {
		baseTrimmed := bytes.TrimSpace(baseLines[i])
		comparedTrimmed := bytes.TrimSpace(comparedLines[i])

		if !bytes.Equal(baseTrimmed, comparedTrimmed) {
			base[i].Type = DiffLineModified
			compared[i].Type = DiffLineModified
			hasDiff = true
		}
	}

	// Mark extra lines as added/removed
	if len(baseLines) > len(comparedLines) {
		for i := len(comparedLines); i < len(baseLines); i++ {
			base[i].Type = DiffLineRemoved
			hasDiff = true
		}
	} else if len(comparedLines) > len(baseLines) {
		for i := len(baseLines); i < len(comparedLines); i++ {
			compared[i].Type = DiffLineAdded
			hasDiff = true
		}
	}

	return hasDiff
}

// diffLCS performs LCS-based diff for smaller files.
// Time: O(n*m), Space: O(n*m) - acceptable for files < 100 lines.
func diffLCS(baseLines, comparedLines [][]byte, base, compared []DiffLine) bool {
	m, n := len(baseLines), len(comparedLines)

	// Build LCS matrix
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if bytes.Equal(bytes.TrimSpace(baseLines[i-1]), bytes.TrimSpace(comparedLines[j-1])) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to mark differences
	i, j := m, n
	hasDiff := false

	for i > 0 && j > 0 {
		if bytes.Equal(bytes.TrimSpace(baseLines[i-1]), bytes.TrimSpace(comparedLines[j-1])) {
			// Lines match
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			// Line removed from base
			base[i-1].Type = DiffLineRemoved
			hasDiff = true
			i--
		} else {
			// Line added in compared
			compared[j-1].Type = DiffLineAdded
			hasDiff = true
			j--
		}
	}

	// Mark remaining lines
	for i > 0 {
		base[i-1].Type = DiffLineRemoved
		hasDiff = true
		i--
	}

	for j > 0 {
		compared[j-1].Type = DiffLineAdded
		hasDiff = true
		j--
	}

	return hasDiff
}

// ComputeCloneGroupDiff computes diffs between all clones in a group.
// Returns the base clone (first) and diffs for all other clones.
type CloneGroupDiff struct {
	Base       *CloneWithContent
	Others     []CloneDiff
	HasAnyDiff bool
}

// CloneWithContent represents a clone with its file content.
type CloneWithContent struct {
	Filename  string
	LineStart int
	LineEnd   int
	Content   []byte
}

// CloneDiff represents a clone with its diff against the base.
type CloneDiff struct {
	CloneWithContent
	Diff DiffResult
}

// ComputeCloneGroupDiff computes diffs for a group of clones.
func ComputeCloneGroupDiff(clones []clone) CloneGroupDiff {
	if len(clones) == 0 {
		return CloneGroupDiff{}
	}

	result := CloneGroupDiff{
		Base: &CloneWithContent{
			Filename:  clones[0].filename,
			LineStart: clones[0].lineStart,
			LineEnd:   clones[0].lineEnd,
			Content:   clones[0].fragment,
		},
		Others: make([]CloneDiff, 0, len(clones)-1),
	}

	// Compute diff for each other clone against base
	for i := 1; i < len(clones); i++ {
		diff := LineDiff(clones[0].fragment, clones[i].fragment)
		if diff.HasDiff {
			result.HasAnyDiff = true
		}

		result.Others = append(result.Others, CloneDiff{
			CloneWithContent: CloneWithContent{
				Filename:  clones[i].filename,
				LineStart: clones[i].lineStart,
				LineEnd:   clones[i].lineEnd,
				Content:   clones[i].fragment,
			},
			Diff: diff,
		})
	}

	return result
}
