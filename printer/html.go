package printer

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"sync"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type htmlprinter struct {
	iota     int
	w        io.Writer
	dupMutex sync.Mutex
	dupls    [][][]*syntax.Node
	ReadFile
}

func NewHTML(w io.Writer, fread ReadFile) Printer {
	return &htmlprinter{w: w, ReadFile: fread, dupls: make([][][]*syntax.Node, 0)}
}

func (p *htmlprinter) PrintHeader() error {
	_, err := fmt.Fprint(p.w, `<!DOCTYPE html>
<meta charset="utf-8"/>
<title>Duplicates</title>
<style>
	pre {
		background-color: #FFD;
		border: 1px solid #E2E2E2;
		padding: 1ex;
	}
</style>
`)
	return err
}

func (p *htmlprinter) PrintClones(dups [][]*syntax.Node) error {
	p.iota++

	// Store clones for later output with sorting
	p.dupMutex.Lock()
	p.dupls = append(p.dupls, dups)
	p.dupMutex.Unlock()

	if _, err := fmt.Fprintf(p.w, "<h1>#%d found %d clones</h1>\n", p.iota, len(dups)); err != nil {
		return err
	}

	clones := make([]clone, len(dups))
	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return fmt.Errorf("internal error: zero length duplicate found")
		}
		nstart := dup[0]
		nend := dup[cnt-1]

		file, err := p.ReadFile(nstart.Filename)
		if err != nil {
			return err
		}

		lineStart, _ := blockLines(file, nstart.Pos, nend.End)
		cl := clone{filename: nstart.Filename, lineStart: lineStart}
		start := findLineBeg(file, nstart.Pos)
		content := append(toWhitespace(file[start:nstart.Pos]), file[nstart.Pos:nend.End]...)
		cl.fragment = deindent(content)
		clones[i] = cl
	}

	sort.Sort(byNameAndLine(clones))
	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "<h2>%s:%d</h2>\n<pre>%s</pre>\n", cl.filename, cl.lineStart,
			html.EscapeString(string(cl.fragment))); err != nil {
			return err
		}
	}
	return nil
}

func (*htmlprinter) PrintFooter() error { return nil }

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
			for j := 0; j < min; j++ {
				if block[i+j+1] != '\t' {
					continue Loop
				}
			}
			block = append(block[:i+1], block[i+1+min:]...)
		}
	}
	return block
}

// OutputHTML generates HTML output with sorting
func (p *htmlprinter) OutputHTML(threshold int, sortBy string) error {
	// Store clones for sorting
	var allDups [][]*syntax.Node
	p.dupMutex.Lock()
	for i := 0; i < len(p.dupls); i++ {
		allDups = append(allDups, p.dupls[i])
	}
	p.dupMutex.Unlock()

	// Sort clones by size (largest first)
	sort.Slice(allDups, func(i, j int) bool {
		return len(allDups[i]) > len(allDups[j])
	})

	// Clear previous output
	p.iota = 0

	// Print sorted clones
	for _, dup := range allDups {
		if err := p.PrintClones([][]*syntax.Node{dup}); err != nil {
			return err
		}
	}

	return nil
}
