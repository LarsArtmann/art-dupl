package printer

import (
	"fmt"
	"html"
	"io"
	"sort"
	"sync"

	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type htmlprinter struct {
	ReadFile

	iota      int
	w         io.Writer
	threshold int
	dupMutex  sync.Mutex
	dupls     [][][]*syntax.Node
}

func NewHTML(w io.Writer, fread ReadFile, threshold ...int) Printer {
	thresh := 15
	if len(threshold) > 0 {
		thresh = threshold[0]
	}
	return &htmlprinter{w: w, ReadFile: fread, threshold: thresh, dupls: make([][][]*syntax.Node, 0)}
}

func (p *htmlprinter) PrintHeader() error {
	_, err := fmt.Fprintf(p.w, `<!DOCTYPE html>
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
	.meta {
		background-color: #f0f0f0;
		padding: 1ex;
		margin-bottom: 1em;
		border: 1px solid #ccc;
	}
</style>
</head>
<body>
<div class="meta"><strong>Threshold:</strong> %d tokens</div>
`, p.threshold)
	return err //nolint:wrapcheck // fmt errors are clear in context
}

func (p *htmlprinter) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	p.iota++

	// Extract sortBy parameter, default to SortBySize
	sortCriteria := SortBySize
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
		cl.fragment = extractContent(fileInfo, nstart, nend)
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

// OutputHTML generates HTML output with sorting.
func (p *htmlprinter) OutputHTML(threshold int, sortBy SortBy) error {
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
