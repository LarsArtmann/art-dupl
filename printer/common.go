package printer

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type clone struct {
	filename  string
	lineStart int
	lineEnd   int
	fragment  []byte
	size      int // Size field for sorting
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
			content = append(toWhitespace(fileInfo.Content[start:startPos]), fileInfo.Content[startPos:endPos]...)
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
	const maxVal = 99
	min := maxVal
	re := regexp.MustCompile(`(^|\n)(\t*)\S`)
	for _, line := range re.FindAllSubmatch(block, -1) {
		indent := line[2]
		if len(indent) < min {
			min = len(indent)
		}
	}
	if min == 0 || min == maxVal {
		return block
	}
	block = block[min:]
Loop:
	for i := 0; i < len(block); i++ {
		if block[i] == '\n' && i != len(block)-1 {
			for j := 0; j < min && i+j+1 < len(block); j++ {
				if block[i+j+1] != '\t' {
					continue Loop
				}
			}
			if i+min+1 <= len(block) {
				block = append(block[:i+1], block[i+1+min:]...)
			}
		}
	}
	return block
}

// formatCloneLine formats a single clone line with the given format string.
// formatStr should be a fmt.Sprintf format string with %s, %d, %d placeholders.
func formatCloneLine(filename string, lineStart, lineEnd int, formatStr string) string {
	return fmt.Sprintf(formatStr, filename, lineStart, lineEnd)
}
