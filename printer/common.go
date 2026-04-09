package printer

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type clone struct {
	filename       string
	lineStart      int
	lineEnd        int
	fragment       []byte
	size           int                 // Fragment/token size for sorting
	fileSize       int                 // Full file size in bytes (for file duplicates)
	classification CloneClassification // Classification metadata for actionable reports
}

type byNameAndLine []clone

func (c byNameAndLine) Len() int { return len(c) }

func (c byNameAndLine) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c byNameAndLine) Less(i, j int) bool {
	if c[i].filename == c[j].filename {
		return c[i].lineStart < c[j].lineStart
	}

	return c[i].filename < c[j].filename
}

func findLineBeg(file []byte, index int) int {
	for i := index; i >= 0; i-- {
		if file[i] == '\n' {
			return i + 1
		}
	}

	return 0
}

func extractContent(fileInfo *FileInfo, nstart, nend *syntax.Node) []byte {
	var content []byte

	start := findLineBeg(fileInfo.Content, int(nstart.Pos))

	fileLen := len(fileInfo.Content)
	if start > fileLen {
		start = fileLen
	}

	startPos := min(int(nstart.Pos), fileLen)
	endPos := min(int(nend.End), fileLen)

	if startPos < endPos {
		if start < startPos {
			content = append(
				toWhitespace(fileInfo.Content[start:startPos]),
				fileInfo.Content[startPos:endPos]...)
		} else {
			content = fileInfo.Content[startPos:endPos]
		}
	}

	return deindent(content)
}

func toWhitespace(str []byte) []byte {
	var out []byte

	for _, c := range bytes.Runes(str) {
		if c == '\t' {
			out = append(out, '\t')
		} else {
			out = append(out, ' ')
		}
	}

	return out
}

func deindent(block []byte) []byte {
	min := findMinIndent(block)
	if min == 0 {
		return block
	}

	return stripIndent(block, min)
}

// findMinIndent finds the minimum tab indentation in the block.
func findMinIndent(block []byte) int {
	const maxVal = 99

	min := maxVal

	re := regexp.MustCompile(`(^|\n)(\t*)\S`)
	for _, line := range re.FindAllSubmatch(block, -1) {
		indent := line[2]
		if len(indent) < min {
			min = len(indent)
		}
	}

	if min == maxVal {
		return 0
	}

	return min
}

// stripIndent removes the given number of tabs from each line.
func stripIndent(block []byte, indentCount int) []byte {
	if indentCount <= 0 || len(block) == 0 {
		return block
	}

	if indentCount >= len(block) {
		return block
	}

	block = block[indentCount:]

	for i := 0; i < len(block); i++ {
		if block[i] != '\n' || i == len(block)-1 {
			continue
		}
		// Bounds check before slicing
		if i+1+indentCount > len(block) {
			continue
		}

		if canStripTabs(block, i+1, indentCount) {
			block = append(block[:i+1], block[i+1+indentCount:]...)
		}
	}

	return block
}

// canStripTabs checks if there are enough tabs to strip at the given position.
func canStripTabs(block []byte, start, count int) bool {
	end := min(count, len(block)-start)
	for j := range end {
		if block[start+j] != '\t' {
			return false
		}
	}

	return true
}

// formatCloneLine formats a single clone line with the given format string.
// formatStr should be a fmt.Sprintf format string with %s, %d, %d placeholders.
func formatCloneLine(filename string, lineStart, lineEnd int, formatStr string) string {
	return fmt.Sprintf(formatStr, filename, lineStart, lineEnd)
}

// writeCloneLines writes formatted clone lines to the writer.
// formatStr uses %s for filename, %d for lineStart, %d for lineEnd.
func writeCloneLines(w io.Writer, clones []clone, formatStr string) error {
	for _, cl := range clones {
		if _, err := fmt.Fprintln(
			w,
			formatCloneLine(cl.filename, cl.lineStart, cl.lineEnd, formatStr),
		); err != nil {
			return err
		}
	}

	return nil
}

// writeFormattedOutput writes data to the writer with proper error wrapping.
func writeFormattedOutput(w io.Writer, data []byte, formatName string) error {
	if _, err := w.Write(data); err != nil {
		return duplerrors.WrapIO(err, formatName, "write")
	}
	return nil
}
