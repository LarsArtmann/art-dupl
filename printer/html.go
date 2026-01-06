package printer

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"sync"

	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type htmlprinter struct {
	ReadFile

	iota     int
	w        io.Writer
	dupMutex sync.Mutex
	dupls    [][][]*syntax.Node
}

//nolint:ireturn // Printer interface is appropriate return type for factory function
func NewHTML(w io.Writer, fread ReadFile) Printer {
	return &htmlprinter{w: w, ReadFile: fread, dupls: make([][][]*syntax.Node, 0)}
}

func (p *htmlprinter) PrintHeader() error {
	_, err := fmt.Fprint(p.w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8"/>
<title>Duplicates</title>
<style>
	pre {
		background-color: #FFD;
		border: 1px solid #E2E2E2;
		padding: 1ex;
	}
</style>
</head>
<body>
`)
	return err //nolint:wrapcheck // fmt errors are clear in context
}

func (p *htmlprinter) PrintClones(dups [][]*syntax.Node, sortBy ...string) error { //nolint:cyclop // HTML generation with multiple formatting paths
	p.iota++

	// Extract sortBy parameter, default to "size"
	sortCriteria := "size"
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	// Apply sorting to the clone groups before processing
	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	// Store clones for later output with sorting
	p.dupMutex.Lock()
	p.dupls = append(p.dupls, sortedDups)
	p.dupMutex.Unlock()

	if _, err := fmt.Fprintf(p.w, "<h1>#%d found %d clones</h1>\n", p.iota, len(sortedDups)); err != nil {
		return err //nolint:wrapcheck // fmt errors are clear in context
	}

	clones := make([]clone, len(sortedDups))
	for i, dup := range sortedDups {
		cnt := len(dup)
		if cnt == 0 {
			return errors.NewInternalError("zero length duplicate found", nil)
		}
		nstart := dup[0]
		nend := dup[cnt-1]

		// Use unified file processor
		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			return err
		}

		cl := clone{filename: fileInfo.Filename, lineStart: fileInfo.LineStart}
		start := findLineBeg(fileInfo.Content, nstart.Pos)
		var content []byte

		// Ensure all indices are within file bounds
		fileLen := len(fileInfo.Content)
		if start > fileLen {
			start = fileLen
		}
		startPos := min(nstart.Pos, fileLen)
		endPos := min(nend.End, fileLen)

		// Only extract content if we have valid bounds
		if startPos < endPos {
			if start < startPos {
				content = append(toWhitespace(fileInfo.Content[start:startPos]), fileInfo.Content[startPos:endPos]...)
			} else {
				content = fileInfo.Content[startPos:endPos]
			}
		}
		cl.fragment = deindent(content)
		clones[i] = cl
	}

	sort.Sort(byNameAndLine(clones))
	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "<h2>%s:%d</h2>\n<pre>%s</pre>\n", cl.filename, cl.lineStart,
			html.EscapeString(string(cl.fragment))); err != nil {
			return err //nolint:wrapcheck // fmt errors are clear in context
		}
	}
	return nil
}

func (p *htmlprinter) PrintFooter() error {
	_, err := fmt.Fprint(p.w, `
</body>
</html>
`)
	return err //nolint:wrapcheck // fmt errors are clear in context
}

func findLineBeg(file []byte, index int) int {
	for i := index; i >= 0; i-- {
		if file[i] == '\n' {
			return i + 1
		}
	}
	return 0
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

func deindent(block []byte) []byte { //nolint:cyclop // String manipulation with multiple parsing paths
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

// OutputHTML generates HTML output with sorting.
func (p *htmlprinter) OutputHTML(threshold int, sortBy string) error {
	// Store clones for sorting - flatten the 3D structure to 2D
	var allDups [][]*syntax.Node
	p.dupMutex.Lock()
	for i := range len(p.dupls) {
		// p.dupls[i] is [][]*syntax.Node, add each clone group to allDups
		for j := range len(p.dupls[i]) {
			allDups = append(allDups, p.dupls[i][j])
		}
	}
	p.dupMutex.Unlock()

	// Apply sorting based on the specified criteria
	// Apply sorting based on specified criteria
	allDups = SortNodesByCriteria(allDups, sortBy)

	// Clear previous output
	p.iota = 0

	// Print sorted clones
	for _, dup := range allDups {
		if err := p.PrintClones([][]*syntax.Node{dup}, sortBy); err != nil {
			return err
		}
	}

	return nil
}
