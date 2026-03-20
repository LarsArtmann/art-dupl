package printer

import (
	"fmt"
	"html"
	"io"
	"sort"
	"strconv"
	"sync"

	"github.com/LarsArtmann/art-dupl/config"
	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// htmlprinter generates HTML output for code duplication reports.
type htmlprinter struct {
	ReadFile

	iota      int
	w         io.Writer
	threshold int
	dupMutex  sync.Mutex
	dupls     [][][]*syntax.Node
	diffMode  config.DiffMode // Enable diff visualization mode
}

// NewHTML creates a new HTML printer.
// Supports optional threshold parameter (default: 15).
func NewHTML(w io.Writer, fread ReadFile, threshold ...int) Printer {
	return NewHTMLWithOptions(w, fread, config.DiffModeDisabled, threshold...)
}

// NewHTMLWithOptions creates a new HTML printer with full options.
// diffMode enables visual diff highlighting between duplicate occurrences.
func NewHTMLWithOptions(w io.Writer, fread ReadFile, diffMode config.DiffMode, threshold ...int) Printer {
	thresh := 15
	if len(threshold) > 0 {
		thresh = threshold[0]
	}

	return &htmlprinter{
		w:         w,
		ReadFile:  fread,
		threshold: thresh,
		dupls:     make([][][]*syntax.Node, 0),
		diffMode:  diffMode,
	}
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
/* Diff mode styles */
.diff-mode { margin-top: 10px; }
.diff-base-header {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	border-radius: 6px;
	padding: 10px 15px;
	margin-bottom: 15px;
}
.diff-base-header .label {
	color: var(--success);
	font-weight: bold;
	font-size: 0.85rem;
}
.diff-base-header .file-info {
	color: var(--text-secondary);
	font-family: 'Fira Code', 'Consolas', monospace;
	font-size: 0.9rem;
}
.diff-selector {
	margin-bottom: 15px;
}
.diff-selector select {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	color: var(--text-primary);
	padding: 8px 12px;
	border-radius: 4px;
	font-size: 0.9rem;
	cursor: pointer;
	width: 100%%;
	max-width: 400px;
}
.diff-selector select:hover { border-color: var(--accent); }
.diff-selector select:focus { outline: none; border-color: var(--accent); }
.diff-comparison {
	display: none;
	border: 1px solid var(--border);
	border-radius: 6px;
	overflow: hidden;
}
.diff-comparison.active { display: block; }
.diff-header {
	background: var(--bg-tertiary);
	padding: 10px 15px;
	display: flex;
	justify-content: space-between;
	align-items: center;
	border-bottom: 1px solid var(--border);
}
.diff-header .file-info {
	color: var(--text-secondary);
	font-family: 'Fira Code', 'Consolas', monospace;
	font-size: 0.9rem;
}
.diff-stats {
	font-size: 0.8rem;
	color: var(--text-secondary);
}
.diff-stats .added { color: var(--success); margin-right: 10px; }
.diff-stats .removed { color: var(--error); margin-right: 10px; }
.diff-stats .modified { color: var(--warning); }
.diff-aggregate-stats {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	border-radius: 6px;
	padding: 10px 15px;
	margin: 10px 0;
	display: flex;
	gap: 15px;
	font-size: 0.9rem;
	font-weight: 500;
}
.diff-aggregate-stats .added { color: var(--success); }
.diff-aggregate-stats .removed { color: var(--error); }
.diff-aggregate-stats .modified { color: var(--warning); }
.diff-content {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 0;
}
.diff-content.single {
	grid-template-columns: 1fr;
}
.diff-panel {
	background: var(--bg-primary);
	overflow-x: auto;
}
.diff-panel.base { border-right: 1px solid var(--border); }
.diff-panel-title {
	background: var(--bg-tertiary);
	padding: 8px 15px;
	font-size: 0.8rem;
	color: var(--text-secondary);
	border-bottom: 1px solid var(--border);
}
.diff-panel pre {
	border: none;
	border-radius: 0;
	margin: 0;
	background: transparent;
}
.diff-panel code {
	display: block;
}
.diff-line {
	display: flex;
	padding: 0;
	min-height: 1.5em;
}
.diff-line:hover { background: rgba(88, 166, 255, 0.1); }
.diff-line-num {
	color: var(--text-secondary);
	padding: 0 10px;
	min-width: 40px;
	text-align: right;
	user-select: none;
	font-size: 0.8rem;
	opacity: 0.5;
}
.diff-line-content {
	flex: 1;
	white-space: pre;
	padding: 0 10px;
}
.diff-line.added { background: rgba(78, 201, 176, 0.15); }
.diff-line.added .diff-line-num { color: var(--success); opacity: 1; }
.diff-line.removed { background: rgba(244, 71, 71, 0.15); }
.diff-line.removed .diff-line-num { color: var(--error); opacity: 1; }
.diff-line.modified { background: rgba(206, 145, 120, 0.15); }
.diff-line.modified .diff-line-num { color: var(--warning); opacity: 1; }
/* Word-level diff highlighting */
.word-added {
	background: rgba(78, 201, 176, 0.4);
	border-radius: 2px;
	padding: 1px 2px;
	font-weight: 500;
}
.word-removed {
	background: rgba(244, 71, 71, 0.4);
	border-radius: 2px;
	padding: 1px 2px;
	font-weight: 500;
	text-decoration: line-through;
}
/* Inline diff view */
.diff-content.inline {
	display: block;
}
.diff-content.inline .diff-panel {
	border-right: none;
	border-bottom: 1px solid var(--border);
}
.diff-content.inline .diff-panel:last-child {
	border-bottom: none;
}
/* Diff view toggle buttons */
.diff-view-toggle {
	display: flex;
	gap: 10px;
	margin-bottom: 15px;
}
.diff-view-toggle button {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	color: var(--text-secondary);
	padding: 6px 12px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 0.85rem;
	transition: all 0.2s;
}
.diff-view-toggle button:hover {
	background: var(--border);
	color: var(--text-primary);
}
.diff-view-toggle button.active {
	background: var(--accent);
	color: white;
	border-color: var(--accent);
}
.diff-legend {
	display: flex;
	gap: 20px;
	padding: 10px 15px;
	background: var(--bg-tertiary);
	border-top: 1px solid var(--border);
	font-size: 0.8rem;
}
.diff-legend-item { display: flex; align-items: center; gap: 6px; }
.diff-legend-color {
	width: 12px;
	height: 12px;
	border-radius: 2px;
}
.diff-legend-color.added { background: rgba(78, 201, 176, 0.3); border: 1px solid var(--success); }
.diff-legend-color.removed { background: rgba(244, 71, 71, 0.3); border: 1px solid var(--error); }
.diff-legend-color.modified { background: rgba(206, 145, 120, 0.3); border: 1px solid var(--warning); }
@media (max-width: 900px) {
	.diff-content { grid-template-columns: 1fr; }
	.diff-panel.base { border-right: none; border-bottom: 1px solid var(--border); }
}
/* Media query percentage escaped for Go template */
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

	sortCriteria := SortBySize
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	p.dupMutex.Lock()
	p.dupls = append(p.dupls, sortedDups)
	p.dupMutex.Unlock()

	totalTokens := calculateTotalTokens(sortedDups)

	if err := p.writeCloneGroupHeader(len(sortedDups), totalTokens); err != nil {
		return err
	}

	clones, err := p.buildClones(sortedDups)
	if err != nil {
		return err
	}

	sort.Sort(byNameAndLine(clones))

	if p.diffMode.IsEnabled() && len(clones) > 1 {
		if err := p.writeDiffView(clones); err != nil {
			return err
		}
	} else {
		if err := p.writeCloneOccurrences(clones); err != nil {
			return err
		}
	}

	return p.writeCloneGroupFooter()
}

func calculateTotalTokens(dups [][]*syntax.Node) int {
	total := 0
	for _, dup := range dups {
		total += len(dup)
	}

	return total
}

func (p *htmlprinter) writeCloneGroupHeader(occurrences, tokens int) error {
	_, err := fmt.Fprintf(p.w, `<div class="clone-group">
<div class="clone-header" onclick="this.parentElement.classList.toggle('collapsed')">
<h3>Clone Group #%d</h3>
<span class="badge">%d occurrences · %d tokens</span>
</div>
<div class="clone-body">
`, p.iota, occurrences, tokens)

	return err //nolint:wrapcheck
}

func (p *htmlprinter) buildClones(dups [][]*syntax.Node) ([]clone, error) {
	clones := make([]clone, len(dups))
	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return nil, errors.NewInternalError(
				fmt.Sprintf("zero length duplicate found in clone group #%d (index=%d)", p.iota, i),
				nil,
			)
		}

		nstart := dup[0]
		nend := dup[cnt-1]

		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			return nil, errors.Wrap(
				err,
				errors.AnalysisError,
				fmt.Sprintf("failed to process clone in group #%d (index=%d, file=%s)", p.iota, i, nstart.Filename),
			)
		}

		clones[i] = clone{
			filename:  fileInfo.Filename,
			lineStart: fileInfo.LineStart,
			fragment:  extractContent(fileInfo, nstart, nend),
		}
	}

	return clones, nil
}

func (p *htmlprinter) writeCloneOccurrences(clones []clone) error {
	for i, cl := range clones {
		vscodeLink := fmt.Sprintf("vscode://file/%s:%d", cl.filename, cl.lineStart)
		_, err := fmt.Fprintf(p.w, `<div class="occurrence">
<div class="file-link">
<a href="%s" title="Open in VSCode">%s:%d</a>
<button class="copy-btn" onclick="copyCode('code-%d-%d')">📋 Copy</button>
</div>
<pre><code id="code-%d-%d">%s</code></pre>
</div>
`, vscodeLink, html.EscapeString(cl.filename), cl.lineStart, p.iota, i, p.iota, i,
			html.EscapeString(string(cl.fragment)))
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

func (p *htmlprinter) writeCloneGroupFooter() error {
	_, err := fmt.Fprint(p.w, "</div></div>\n")

	return err //nolint:wrapcheck
}

// writeDiffView renders the diff visualization for clone groups.
func (p *htmlprinter) writeDiffView(clones []clone) error {
	groupDiff := ComputeCloneGroupDiff(clones)

	if len(groupDiff.Others) == 0 {
		// Single clone, no diff possible
		return p.writeCloneOccurrences(clones)
	}

	// Calculate aggregate stats
	totalAdded, totalRemoved, totalModified := groupDiff.TotalAdded, groupDiff.TotalRemoved, groupDiff.TotalModified
	hasAggregateStats := totalAdded > 0 || totalRemoved > 0 || totalModified > 0

	// Write base clone header
	baseVSCode := fmt.Sprintf("vscode://file/%s:%d", groupDiff.Base.Filename, groupDiff.Base.LineStart)
	_, err := fmt.Fprintf(p.w, `
<div class="diff-mode">
<div class="diff-base-header">
<span class="label">📋 BASE REFERENCE</span>
<div class="file-info"><a href="%s">%s:%d</a></div>
`, baseVSCode, html.EscapeString(groupDiff.Base.Filename), groupDiff.Base.LineStart)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Write aggregate diff stats if available
	if hasAggregateStats {
		_, err = fmt.Fprint(p.w, `<div class="diff-aggregate-stats">`)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if totalAdded > 0 {
			_, _ = fmt.Fprintf(p.w, `<span class="added">+%d added</span>`, totalAdded)
		}
		if totalRemoved > 0 {
			_, _ = fmt.Fprintf(p.w, `<span class="removed">-%d removed</span>`, totalRemoved)
		}
		if totalModified > 0 {
			_, _ = fmt.Fprintf(p.w, `<span class="modified">~%d modified</span>`, totalModified)
		}

		_, err = fmt.Fprint(p.w, `</div>`)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	// Write view toggle buttons
	if err := p.writeDiffViewToggle(); err != nil {
		return err
	}

	// Write comparison selector for multiple clones
	if len(groupDiff.Others) > 1 {
		if err := p.writeDiffSelector(groupDiff); err != nil {
			return err
		}
	}

	// Write diff panels for each comparison
	for idx, other := range groupDiff.Others {
		if err := p.writeDiffComparison(groupDiff.Base, other, idx, len(groupDiff.Others)); err != nil {
			return err
		}
	}

	// Write legend
	_, err = fmt.Fprint(p.w, `
<div class="diff-legend">
<div class="diff-legend-item"><div class="diff-legend-color added"></div>Added</div>
<div class="diff-legend-item"><div class="diff-legend-color removed"></div>Removed</div>
<div class="diff-legend-item"><div class="diff-legend-color modified"></div>Modified</div>
</div>
</div>
`)

	return err //nolint:wrapcheck
}

// writeDiffViewToggle writes the view mode toggle buttons (side-by-side vs inline).
func (p *htmlprinter) writeDiffViewToggle() error {
	_, err := fmt.Fprintf(p.w, `
<div class="diff-view-toggle">
<button id="diff-toggle-%d-side" class="active" data-mode="side" onclick="toggleDiffView(%d, 'side')">◫ Side by Side</button>
<button id="diff-toggle-%d-inline" data-mode="inline" onclick="toggleDiffView(%d, 'inline')">▣ Inline</button>
</div>
`, p.iota, p.iota, p.iota, p.iota)

	return err //nolint:wrapcheck
}

// writeDiffSelector writes the dropdown for selecting which clone to compare.
func (p *htmlprinter) writeDiffSelector(groupDiff CloneGroupDiff) error {
	_, err := fmt.Fprint(p.w, `
<div class="diff-selector">
<select id="diff-select-`+strconv.Itoa(p.iota)+`" onchange="showDiff(this.value, `+strconv.Itoa(p.iota)+`)">
<option value="" disabled selected>Compare with...</option>
`)
	if err != nil {
		return err //nolint:wrapcheck
	}

	for idx, other := range groupDiff.Others {
		_, err := fmt.Fprintf(p.w, `<option value="%d">%s:%d</option>
`, idx, html.EscapeString(other.Filename), other.LineStart)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	_, err = fmt.Fprint(p.w, `</select>
</div>
`)

	return err //nolint:wrapcheck
}

// writeDiffComparison writes a side-by-side diff comparison.
func (p *htmlprinter) writeDiffComparison(base *CloneWithContent, other CloneDiff, index, total int) error {
	activeClass := "active"
	if total > 1 && index > 0 {
		activeClass = ""
	}

	otherVSCode := fmt.Sprintf("vscode://file/%s:%d", other.Filename, other.LineStart)

	// Calculate diff stats
	added, removed, modified := countDiffStats(other.Diff)

	_, err := fmt.Fprintf(p.w, `
<div class="diff-comparison %s" id="diff-compare-%d-%d">
<div class="diff-header">
<span class="file-info"><a href="%s">%s:%d</a></span>
<span class="diff-stats">
`, activeClass, p.iota, index, otherVSCode, html.EscapeString(other.Filename), other.LineStart)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Write stats
	if added > 0 {
		_, _ = fmt.Fprintf(p.w, `<span class="added">+%d</span>`, added)
	}
	if removed > 0 {
		_, _ = fmt.Fprintf(p.w, `<span class="removed">-%d</span>`, removed)
	}
	if modified > 0 {
		_, _ = fmt.Fprintf(p.w, `<span class="modified">~%d</span>`, modified)
	}

	_, err = fmt.Fprint(p.w, `</span>
</div>
<div class="diff-content" data-diff-index="`+strconv.Itoa(index)+`">
`)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Write base and compared panels with word-level highlighting
	if err := p.writeDiffPanelsWithWordDiff(base, other); err != nil {
		return err
	}

	_, err = fmt.Fprint(p.w, `</div>
</div>
`)

	return err //nolint:wrapcheck
}

// writeDiffPanel writes a single diff panel (base or compared).
func (p *htmlprinter) writeDiffPanel(panelType string, content []byte, lines []DiffLine, showLineNumbers bool) error {
	_, err := fmt.Fprintf(p.w, `
<div class="diff-panel %s">
<div class="diff-panel-title">%s</div>
<pre><code>
`, panelType, map[string]string{"base": "Base Reference", "compared": "Compared"}[panelType])
	if err != nil {
		return err //nolint:wrapcheck
	}

	// If no diff computed (same number of lines), show plain content
	if len(lines) == 0 {
		_, err = fmt.Fprintf(p.w, `%s`, html.EscapeString(string(content)))
	} else {
		for _, line := range lines {
			typeClass := ""
			switch line.Type {
			case DiffLineAdded:
				typeClass = "added"
			case DiffLineRemoved:
				typeClass = "removed"
			case DiffLineModified:
				typeClass = "modified"
			}

			lineNum := ""
			if showLineNumbers {
				lineNum = fmt.Sprintf(`<span class="diff-line-num">%d</span>`, line.LineNumber)
			}

			_, err := fmt.Fprintf(p.w, `<div class="diff-line %s">%s<span class="diff-line-content">%s</span></div>
`, typeClass, lineNum, html.EscapeString(line.Content))
			if err != nil {
				return err //nolint:wrapcheck
			}
		}
	}

	_, err = fmt.Fprint(p.w, `</code></pre>
</div>
`)

	return err //nolint:wrapcheck
}

// writeDiffPanelsWithWordDiff renders both base and compared panels with word-level highlighting.
func (p *htmlprinter) writeDiffPanelsWithWordDiff(base *CloneWithContent, other CloneDiff) error {
	baseLines := other.Diff.Base
	comparedLines := other.Diff.Compared

	// Start base panel
	_, err := fmt.Fprint(p.w, `
<div class="diff-panel base">
<div class="diff-panel-title">Base Reference</div>
<pre><code>
`)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Render base panel
	if err := p.renderDiffLines(baseLines, comparedLines, true); err != nil {
		return err
	}

	// Close base panel, start compared panel
	_, err = fmt.Fprint(p.w, `</code></pre>
</div>
<div class="diff-panel compared">
<div class="diff-panel-title">Compared</div>
<pre><code>
`)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Render compared panel
	if err := p.renderDiffLines(comparedLines, baseLines, false); err != nil {
		return err
	}

	_, err = fmt.Fprint(p.w, `</code></pre>
</div>
`)

	return err //nolint:wrapcheck
}

// renderDiffLines renders diff lines with optional word-level highlighting for modified lines.
func (p *htmlprinter) renderDiffLines(lines, oppositeLines []DiffLine, isBasePanel bool) error {
	for i, line := range lines {
		typeClass := ""
		switch line.Type {
		case DiffLineAdded:
			typeClass = "added"
		case DiffLineRemoved:
			typeClass = "removed"
		case DiffLineModified:
			typeClass = "modified"
		}

		lineNum := fmt.Sprintf(`<span class="diff-line-num">%d</span>`, line.LineNumber)

		var content string
		if line.Type == DiffLineModified && i < len(oppositeLines) && oppositeLines[i].Type == DiffLineModified {
			// For modified lines, show word-level diff
			if isBasePanel {
				content = WordDiff(line.Content, oppositeLines[i].Content)
			} else {
				content = WordDiff(oppositeLines[i].Content, line.Content)
			}
		} else {
			content = html.EscapeString(line.Content)
		}

		_, err := fmt.Fprintf(p.w, `<div class="diff-line %s">%s<span class="diff-line-content">%s</span></div>
`, typeClass, lineNum, content)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

// countDiffStats counts the number of added, removed, and modified lines.
func countDiffStats(diff DiffResult) (added, removed, modified int) {
	for _, line := range diff.Compared {
		switch line.Type {
		case DiffLineAdded:
			added++
		case DiffLineRemoved:
			removed++
		case DiffLineModified:
			modified++
		}
	}

	return added, removed, modified
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
function showDiff(index, groupId) {
	// Hide all diff comparisons for this group
	const comparisons = document.querySelectorAll('[id^="diff-compare-' + groupId + '-"]');
	comparisons.forEach(function(comp) {
		comp.classList.remove('active');
	});
	// Show selected comparison
	const selected = document.getElementById('diff-compare-' + groupId + '-' + index);
	if (selected) {
		selected.classList.add('active');
	}
}

// Diff view mode toggle: side-by-side vs inline
function toggleDiffView(groupId, mode) {
	const comparisons = document.querySelectorAll('[id^="diff-compare-' + groupId + '-"]');
	comparisons.forEach(function(comp) {
		const content = comp.querySelector('.diff-content');
		if (content) {
			if (mode === 'inline') {
				content.classList.add('inline');
			} else {
				content.classList.remove('inline');
			}
		}
	});

	// Update toggle button states
	const buttons = document.querySelectorAll('[id^="diff-toggle-' + groupId + '-"]');
	buttons.forEach(function(btn) {
		btn.classList.remove('active');
		if (btn.dataset.mode === mode) {
			btn.classList.add('active');
		}
	});

	// Save preference
	try {
		localStorage.setItem('artdupl-diff-mode', mode);
	} catch (e) {
		// Ignore localStorage errors
	}
}

// Initialize diff view mode from saved preference
document.addEventListener('DOMContentLoaded', function() {
	try {
		const savedMode = localStorage.getItem('artdupl-diff-mode');
		if (savedMode === 'inline') {
			// Apply inline mode to all diff comparisons
			const allContents = document.querySelectorAll('.diff-content');
			allContents.forEach(function(content) {
				content.classList.add('inline');
			});
			// Update all toggle buttons
			const allToggles = document.querySelectorAll('.diff-view-toggle button');
			allToggles.forEach(function(btn) {
				btn.classList.remove('active');
				if (btn.dataset.mode === 'inline') {
					btn.classList.add('active');
				}
			});
		}
	} catch (e) {
		// Ignore localStorage errors
	}
});
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
		err := p.PrintClones([][]*syntax.Node{dup}, sortBy)
		if err != nil {
			return err
		}
	}

	return nil
}
