package printer

import (
	"bytes"
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
	minLen := min(len(comparedLines), len(baseLines))

	// Compare line by line up to the shorter length
	for i := range minLen {
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
	// Aggregate statistics across all diffs in this group
	TotalAdded    int
	TotalRemoved  int
	TotalModified int
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
	for idx := 1; idx < len(clones); idx++ {
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
				Filename:  clones[idx].filename,
				LineStart: clones[idx].lineStart,
				LineEnd:   clones[idx].lineEnd,
				Content:   clones[idx].fragment,
			},
			Diff: diff,
		})
	}

	return result
}

// countDiffLineStats counts added, removed, and modified lines in a DiffResult.
func countDiffLineStats(diff DiffResult) (added, removed, modified int) {
	for _, line := range diff.Compared {
		switch line.Type {
		case DiffLineAdded:
			added++
		case DiffLineModified:
			modified++
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
			result.WriteString(htmlEscape(diff.Text))
			result.WriteString(`</span> `)
		case diffmatchpatch.DiffInsert:
			result.WriteString(`<span class="word-added">`)
			result.WriteString(htmlEscape(diff.Text))
			result.WriteString(`</span> `)
		case diffmatchpatch.DiffEqual:
			result.WriteString(htmlEscape(diff.Text))
			result.WriteString(" ")
		}
	}

	return strings.TrimSpace(result.String())
}

// htmlEscape escapes HTML special characters.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
