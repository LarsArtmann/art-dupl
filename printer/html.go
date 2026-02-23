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

// htmlTemplate is the modernized HTML template with dark theme.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Code Duplication Report</title>
<style>
:root {
	--bg-primary: #1e1e1e;
	--bg-secondary: #252526;
	--bg-tertiary: #2d2d30;
	--text-primary: #d4d4d4;
	--text-secondary: #9cdcfe;
	--accent: #58a6ff;
	--accent-hover: #79b8ff;
	--border: #3e3e42;
	--success: #4ec9b0;
	--warning: #ce9178;
	--error: #f44747;
	--keyword: #c586c0;
	--string: #ce9178;
	--comment: #6a9955;
	--function: #dcdcaa;
	--number: #b5cea8;
	--type: #4ec9b0;
}
* { box-sizing: border-box; }
body {
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
	background: var(--bg-primary);
	color: var(--text-primary);
	line-height: 1.6;
	margin: 0;
	padding: 0;
}
.container { max-width: 1400px; margin: 0 auto; padding: 20px; }
header {
	background: var(--bg-secondary);
	border-bottom: 1px solid var(--border);
	padding: 20px 0;
	margin-bottom: 30px;
}
header h1 {
	margin: 0;
	color: var(--accent);
	font-size: 1.8rem;
}
.stats-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	gap: 15px;
	margin-bottom: 30px;
}
.stat-card {
	background: var(--bg-secondary);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 20px;
	text-align: center;
}
.stat-card h3 {
	margin: 0 0 10px 0;
	color: var(--accent);
	font-size: 2rem;
}
.stat-card p {
	margin: 0;
	color: var(--text-secondary);
	font-size: 0.9rem;
}
.clone-group {
	background: var(--bg-secondary);
	border: 1px solid var(--border);
	border-radius: 8px;
	margin-bottom: 20px;
	overflow: hidden;
}
.clone-header {
	background: var(--bg-tertiary);
	padding: 15px 20px;
	display: flex;
	justify-content: space-between;
	align-items: center;
	cursor: pointer;
}
.clone-header h3 {
	margin: 0;
	color: var(--text-primary);
	font-size: 1.1rem;
}
.clone-header .badge {
	background: var(--accent);
	color: white;
	padding: 4px 12px;
	border-radius: 12px;
	font-size: 0.85rem;
}
.clone-body { padding: 20px; }
.occurrence {
	margin-bottom: 20px;
	border-left: 3px solid var(--accent);
	padding-left: 15px;
}
.file-link {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 10px;
}
.file-link a {
	color: var(--accent);
	text-decoration: none;
	font-family: 'Fira Code', 'Consolas', monospace;
	font-size: 0.9rem;
}
.file-link a:hover { color: var(--accent-hover); text-decoration: underline; }
.copy-btn {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	color: var(--text-secondary);
	padding: 4px 12px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 0.8rem;
}
.copy-btn:hover { background: var(--border); }
pre {
	background: var(--bg-primary);
	border: 1px solid var(--border);
	border-radius: 6px;
	padding: 15px;
	overflow-x: auto;
	margin: 0;
}
code {
	font-family: 'Fira Code', 'Consolas', monospace;
	font-size: 0.85rem;
	line-height: 1.5;
}
/* Syntax highlighting classes */
.keyword { color: var(--keyword); }
.string { color: var(--string); }
.comment { color: var(--comment); }
.function { color: var(--function); }
.number { color: var(--number); }
.type { color: var(--type); }
footer {
	text-align: center;
	padding: 30px;
	color: var(--text-secondary);
	font-size: 0.85rem;
	border-top: 1px solid var(--border);
	margin-top: 40px;
}
.collapsed .clone-body { display: none; }
</style>
</head>
<body>
<div class="container">
<header>
<h1>🔍 Code Duplication Report</h1>
</header>
<div class="stats-grid">
<div class="stat-card"><h3>%d</h3><p>Threshold (tokens)</p></div>
</div>
`

func (p *htmlprinter) PrintHeader() error {
	_, err := fmt.Fprintf(p.w, htmlTemplate, p.threshold)
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

	// Calculate total tokens for this group
	totalTokens := 0
	for _, dup := range sortedDups {
		if len(dup) > 0 {
			totalTokens += len(dup)
		}
	}

	// Modern clone group container
	if _, err := fmt.Fprintf(p.w, `<div class="clone-group">
<div class="clone-header" onclick="this.parentElement.classList.toggle('collapsed')">
<h3>Clone Group #%d</h3>
<span class="badge">%d occurrences · %d tokens</span>
</div>
<div class="clone-body">
`, p.iota, len(sortedDups), totalTokens); err != nil {
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
	for i, cl := range clones {
		vscodeLink := fmt.Sprintf("vscode://file/%s:%d", cl.filename, cl.lineStart)
		if _, err := fmt.Fprintf(p.w, `<div class="occurrence">
<div class="file-link">
<a href="%s" title="Open in VSCode">%s:%d</a>
<button class="copy-btn" onclick="copyCode('code-%d-%d')">📋 Copy</button>
</div>
<pre><code id="code-%d-%d">%s</code></pre>
</div>
`, vscodeLink, html.EscapeString(cl.filename), cl.lineStart, p.iota, i, p.iota, i,
			html.EscapeString(string(cl.fragment))); err != nil {
			return err //nolint:wrapcheck // fmt errors are clear in context
		}
	}

	// Close clone-body and clone-group
	if _, err := fmt.Fprint(p.w, "</div></div>\n"); err != nil {
		return err //nolint:wrapcheck // fmt errors are clear in context
	}

	return nil
}

func (p *htmlprinter) PrintFooter() error {
	_, err := fmt.Fprint(p.w, `
</div>
<footer>
<p>Generated by art-dupl · Code Duplication Detection Tool</p>
</footer>
<script>
function copyCode(elementId) {
	const code = document.getElementById(elementId).innerText;
	navigator.clipboard.writeText(code).then(() => {
		const btn = event.target;
		const original = btn.innerText;
		btn.innerText = '✓ Copied!';
		setTimeout(() => btn.innerText = original, 2000);
	});
}
</script>
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
