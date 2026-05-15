package printer

import (
	"bytes"
	"html"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
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

// initDiffLines converts raw lines into DiffLine structs with line numbers.
func initDiffLines(lines [][]byte) []DiffLine {
	result := make([]DiffLine, len(lines))
	for i, line := range lines {
		result[i] = DiffLine{
			Content:    string(line),
			Type:       DiffLineEqual,
			LineNumber: i + 1,
		}
	}

	return result
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
		Base:     initDiffLines(baseLines),
		Compared: initDiffLines(comparedLines),
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

	for i := range data {
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

// markLinesModified marks lines as modified if their trimmed content differs.
func markLinesModified(baseTrimmed, comparedTrimmed []byte, baseLine, comparedLine *DiffLine) bool {
	if !bytes.Equal(baseTrimmed, comparedTrimmed) {
		baseLine.Type = DiffLineModified
		comparedLine.Type = DiffLineModified

		return true
	}

	return false
}

// diffSameLength compares lines when both fragments have the same number of lines.
func diffSameLength(base, compared []DiffLine) bool {
	hasDiff := false

	for i := range base {
		if markLinesModified(
			bytes.TrimSpace([]byte(base[i].Content)),
			bytes.TrimSpace([]byte(compared[i].Content)),
			&base[i],
			&compared[i],
		) {
			hasDiff = true
		}
	}

	return hasDiff
}

// diffDifferentLength handles fragments with different line counts.
// Uses a simple LCS (Longest Common Subsequence) approach for small files,
// falls back to line-by-line for performance with large files.
func diffDifferentLength(srcRows, dstRows [][]byte, left, right []DiffLine) bool {
	// For performance, use simple heuristic for large files (>100 lines)
	if len(srcRows) > 100 || len(dstRows) > 100 {
		return diffLargeFiles(srcRows, dstRows, left, right)
	}

	return diffLCS(srcRows, dstRows, left, right)
}

// diffLargeFiles uses a faster heuristic for large files.
func diffLargeFiles(srcRows, dstRows [][]byte, left, right []DiffLine) bool {
	hasDiff := false
	minLen := min(len(dstRows), len(srcRows))

	// Compare line by line up to the shorter length
	for i := range minLen {
		if markLinesModified(
			bytes.TrimSpace(srcRows[i]),
			bytes.TrimSpace(dstRows[i]),
			&left[i],
			&right[i],
		) {
			hasDiff = true
		}
	}

	// Mark extra lines as added/removed
	if len(srcRows) > len(dstRows) {
		for i := len(dstRows); i < len(srcRows); i++ {
			left[i].Type = DiffLineRemoved
			hasDiff = true
		}
	} else if len(dstRows) > len(srcRows) {
		for i := len(srcRows); i < len(dstRows); i++ {
			right[i].Type = DiffLineAdded
			hasDiff = true
		}
	}

	return hasDiff
}

// linesMatch compares two lines after trimming whitespace.
func linesMatch(line1, line2 []byte) bool {
	return bytes.Equal(bytes.TrimSpace(line1), bytes.TrimSpace(line2))
}

// diffLCS performs LCS-based diff for smaller files.
// Time: O(n*m), Space: O(n*m) - acceptable for files < 100 lines.
func diffLCS(srcRows, dstRows [][]byte, left, right []DiffLine) bool {
	m, n := len(srcRows), len(dstRows)

	// Build LCS matrix
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if linesMatch(srcRows[i-1], dstRows[j-1]) {
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
		switch {
		case linesMatch(srcRows[i-1], dstRows[j-1]):
			// Lines match
			i--
			j--
		case dp[i-1][j] >= dp[i][j-1]:
			// Line removed from base
			left[i-1].Type = DiffLineRemoved
			hasDiff = true
			i--
		default:
			// Line added in compared
			right[j-1].Type = DiffLineAdded
			hasDiff = true
			j--
		}
	}

	// Mark remaining lines
	for i > 0 {
		left[i-1].Type = DiffLineRemoved
		hasDiff = true
		i--
	}

	for j > 0 {
		right[j-1].Type = DiffLineAdded
		hasDiff = true
		j--
	}

	return hasDiff
}

// CloneGroupDiff computes diffs between all clones in a group.
// Returns the base clone (first) and diffs for all other clones.
type CloneGroupDiff struct {
	Base       *CloneWithContent
	Others     []CloneDiff
	HasAnyDiff bool
	// Aggregate statistics across all diffs in this group
	TotalAdded    int
	TotalRemoved  int
	TotalModified int
}

// CloneWithContentMixin provides common location fields for clone structures.
type CloneWithContentMixin struct {
	Filename  string
	LineStart int
	LineEnd   int
}

// CloneWithContent represents a clone with its file content.
type CloneWithContent struct {
	CloneWithContentMixin

	Content []byte
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
			CloneWithContentMixin: CloneWithContentMixin{
				Filename:  clones[0].filename,
				LineStart: clones[0].lineStart,
				LineEnd:   clones[0].lineEnd,
			},
			Content: clones[0].fragment,
		},
		Others: make([]CloneDiff, 0, len(clones)-1),
	}

	// Compute diff for each other clone against base
	for idx := 1; idx < len(clones); idx++ {
		//nolint:gosec // G602: Bounds checked above (len(clones) > 0, idx < len(clones))
		diff := LineDiff(clones[0].fragment, clones[idx].fragment)
		if diff.HasDiff {
			result.HasAnyDiff = true
		}

		// Aggregate diff stats
		added, removed, modified := countDiffLineStats(diff)
		result.TotalAdded += added
		result.TotalRemoved += removed
		result.TotalModified += modified

		result.Others = append(result.Others, CloneDiff{
			CloneWithContent: CloneWithContent{
				CloneWithContentMixin: CloneWithContentMixin{
					Filename:  clones[idx].filename,
					LineStart: clones[idx].lineStart,
					LineEnd:   clones[idx].lineEnd,
				},
				Content: clones[idx].fragment,
			},
			Diff: diff,
		})
	}

	return result
}

// countDiffLineStats counts added, removed, and modified lines in a DiffResult.
//
//nolint:nonamedreturns // Named returns are appropriate for counting functions
func countDiffLineStats(diff DiffResult) (added, removed, modified int) {
	for _, line := range diff.Compared {
		switch line.Type {
		case DiffLineAdded:
			added++
		case DiffLineModified:
			modified++
		case DiffLineEqual, DiffLineRemoved:
			// No action needed
		}
	}

	for _, line := range diff.Base {
		if line.Type == DiffLineRemoved {
			removed++
		}
	}

	return added, removed, modified
}

// WordDiff performs a word-level diff between two strings using go-diff.
// Returns HTML-formatted string with highlighted word changes.
func WordDiff(base, compared string) string {
	dmp := diffmatchpatch.New()

	// Split texts into words for word-level diffing
	words1 := strings.Fields(base)
	words2 := strings.Fields(compared)

	// Convert words to unique characters (line-mode optimization)
	chars1, chars2, wordArray := dmp.DiffLinesToChars(
		strings.Join(words1, "\n"),
		strings.Join(words2, "\n"),
	)

	// Diff the character representations
	diffs := dmp.DiffMain(chars1, chars2, false)

	// Convert back to words
	diffs = dmp.DiffCharsToLines(diffs, wordArray)

	// Cleanup for better readability
	diffs = dmp.DiffCleanupSemantic(diffs)

	// Build HTML output
	var result strings.Builder

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			result.WriteString(`<span class="word-removed">`)
			result.WriteString(html.EscapeString(diff.Text))
			result.WriteString(`</span> `)
		case diffmatchpatch.DiffInsert:
			result.WriteString(`<span class="word-added">`)
			result.WriteString(html.EscapeString(diff.Text))
			result.WriteString(`</span> `)
		case diffmatchpatch.DiffEqual:
			result.WriteString(html.EscapeString(diff.Text))
			result.WriteString(" ")
		}
	}

	return strings.TrimSpace(result.String())
}
