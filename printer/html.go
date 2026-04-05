package printer

import (
	"fmt"
	"html"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/LarsArtmann/art-dupl/config"
	errors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// classificationStats tracks counts for HTML summary.
type classificationStats struct {
	categoryCounts map[CloneCategory]int
	priorityCounts map[ClonePriority]int
	testCount      int
	prodCount      int
	totalClones    int
	totalTokens    int
}

// ReportMetadata contains CLI settings used for the report.
type ReportMetadata struct {
	Semantic         bool
	DetectionMethods []string
	SortBy           string
	FilterGenerated  bool
	IncludeSQLC      bool
	IncludeTempl     bool
}

// htmlprinter generates HTML output for code duplication reports.
type htmlprinter struct {
	ReadFile

	iota      int
	w         io.Writer
	threshold int
	dupMutex  sync.Mutex
	dupls     [][][]*syntax.Node
	diffMode  config.DiffMode // Enable diff visualization mode
	stats     classificationStats
	metadata  ReportMetadata
}

// NewHTML creates a new HTML printer.
// Supports optional threshold parameter (default: 15).
func NewHTML(w io.Writer, fread ReadFile, threshold ...int) Printer {
	return NewHTMLWithOptions(w, fread, config.DiffModeDisabled, ReportMetadata{}, threshold...)
}

// NewHTMLWithOptions creates a new HTML printer with full options.
// diffMode enables visual diff highlighting between duplicate occurrences.
// metadata contains CLI settings to display in the report.
func NewHTMLWithOptions(
	w io.Writer,
	fread ReadFile,
	diffMode config.DiffMode,
	metadata ReportMetadata,
	threshold ...int,
) Printer {
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
		stats: classificationStats{
			categoryCounts: make(map[CloneCategory]int),
			priorityCounts: make(map[ClonePriority]int),
		},
		metadata: metadata,
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
/* Classification badges */
.badge-group {
	display: flex;
	gap: 8px;
	flex-wrap: wrap;
	align-items: center;
}
.badge-category, .badge-priority, .badge-test {
	padding: 3px 10px;
	border-radius: 12px;
	font-size: 0.8rem;
	font-weight: 500;
	display: inline-flex;
	align-items: center;
	gap: 4px;
}
.badge-category { background: var(--bg-tertiary); color: var(--text-secondary); border: 1px solid var(--border); }
.badge-priority.critical { background: rgba(244, 71, 71, 0.2); color: var(--error); border: 1px solid var(--error); }
.badge-priority.high { background: rgba(206, 145, 120, 0.2); color: var(--warning); border: 1px solid var(--warning); }
.badge-priority.medium { background: rgba(88, 166, 255, 0.2); color: var(--accent); border: 1px solid var(--accent); }
.badge-priority.low { background: rgba(78, 201, 176, 0.2); color: var(--success); border: 1px solid var(--success); }
.badge-test { background: rgba(197, 134, 192, 0.2); color: var(--keyword); border: 1px solid var(--keyword); }
.suggestion {
	background: var(--bg-tertiary);
	border-left: 3px solid var(--accent);
	padding: 8px 12px;
	margin: 10px 0;
	font-size: 0.85rem;
	color: var(--text-secondary);
	font-style: italic;
}
/* Summary section for filter buttons */
.summary-section {
	background: var(--bg-secondary);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 20px;
	margin-bottom: 30px;
}
.summary-section h3 {
	margin: 0 0 20px 0;
	color: var(--accent);
	font-size: 1.2rem;
}
.summary-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
	gap: 15px;
}
.summary-item {
	background: var(--bg-tertiary);
	border-radius: 6px;
	padding: 12px;
	text-align: center;
}
.summary-emoji {
	font-size: 1.5rem;
	display: block;
}
.summary-label {
	font-size: 0.8rem;
	color: var(--text-secondary);
	display: block;
}
.summary-count {
	font-size: 1.4rem;
	color: var(--text-primary);
	font-weight: bold;
}
.filter-buttons {
	display: flex;
	gap: 10px;
	flex-wrap: wrap;
	margin-bottom: 15px;
}
.filter-btn {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	color: var(--text-secondary);
	padding: 6px 12px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 0.85rem;
	transition: all 0.2s;
}
.filter-btn:hover {
	background: var(--border);
	color: var(--text-primary);
}
.filter-btn.active {
	background: var(--accent);
	color: white;
	border-color: var(--accent);
}
.summary-value {
	font-size: 1.4rem;
	color: var(--text-primary);
	font-weight: bold;
}
.summary-category {
	margin-top: 15px;
}
.summary-category h3 {
	margin: 0 0 10px 0;
	font-size: 1rem;
	color: var(--accent);
}
.category-list, .priority-list {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
}
.category-tag, .priority-tag {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	border-radius: 4px;
	padding: 4px 10px;
	font-size: 0.8rem;
	color: var(--text-secondary);
}
.priority-tag.priority-critical { border-color: var(--error); color: var(--error); }
.priority-tag.priority-high { border-color: var(--warning); color: var(--warning); }
.priority-tag.priority-medium { border-color: var(--accent); color: var(--accent); }
.priority-tag.priority-low { border-color: var(--success); color: var(--success); }
.filter-separator {
	color: var(--border);
	margin: 0 5px;
}
/* Metadata section */
.metadata-section {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
	margin-top: 10px;
}
.metadata-badge {
	background: var(--bg-tertiary);
	border: 1px solid var(--border);
	border-radius: 4px;
	padding: 4px 10px;
	font-size: 0.75rem;
	color: var(--text-secondary);
}
.metadata-badge.enabled {
	background: rgba(88, 166, 255, 0.15);
	border-color: var(--accent);
	color: var(--accent);
}
.metadata-badge.disabled {
	opacity: 0.6;
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
<div class="metadata-section" id="report-metadata"></div>
`

func (p *htmlprinter) PrintHeader() error {
	if _, err := fmt.Fprintf(p.w, htmlTemplate, p.threshold); err != nil {
		return err
	}

	return p.writeMetadata()
}

func (p *htmlprinter) writeMetadata() error {
	var parts []string

	if p.metadata.Semantic {
		parts = append(parts, `<span class="metadata-badge enabled">🔤 Semantic</span>`)
	} else {
		parts = append(parts, `<span class="metadata-badge disabled">🔤 Structural</span>`)
	}

	if len(p.metadata.DetectionMethods) > 0 {
		parts = append(parts, fmt.Sprintf(`<span class="metadata-badge">🔍 %s</span>`,
			html.EscapeString(strings.Join(p.metadata.DetectionMethods, ", "))))
	}

	if p.metadata.SortBy != "" {
		parts = append(parts, fmt.Sprintf(`<span class="metadata-badge">📊 %s</span>`,
			html.EscapeString(p.metadata.SortBy)))
	}

	if p.metadata.FilterGenerated {
		parts = append(parts, `<span class="metadata-badge">🚫 Filter Generated</span>`)
	}

	if p.metadata.IncludeSQLC {
		parts = append(parts, `<span class="metadata-badge">📦 Include SQLC</span>`)
	}

	if p.metadata.IncludeTempl {
		parts = append(parts, `<span class="metadata-badge">📦 Include Templ</span>`)
	}

	if len(parts) == 0 {
		return nil
	}

	metaStr := strings.Join(parts, "")
	if _, err := fmt.Fprintf(
		p.w,
		`<script>document.getElementById('report-metadata').innerHTML = %s;</script>`,
		strconv.Quote(metaStr),
	); err != nil {
		return err
	}

	return nil
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

	// Build clones first so we can extract classification data for the header
	clones, err := p.buildClones(sortedDups)
	if err != nil {
		return err
	}

	if err := p.writeCloneGroupHeader(clones); err != nil {
		return err
	}

	sort.Sort(byNameAndLine(clones))

	if p.diffMode.IsEnabled() && len(clones) > 1 {
		err := p.writeDiffView(clones)
		if err != nil {
			return err
		}
	} else {
		err := p.writeCloneOccurrences(clones)
		if err != nil {
			return err
		}
	}

	return p.writeCloneGroupFooter()
}

func (p *htmlprinter) writeCloneGroupHeader(clones []clone) error {
	// Calculate aggregate metrics from clones
	occurrences := len(clones)
	totalTokens := 0
	hasTest := false
	highestPriority := PriorityLow
	primaryCategory := CategoryUnknown
	categoryCounts := make(map[CloneCategory]int)

	for _, cl := range clones {
		totalTokens += cl.classification.Tokens
		if cl.classification.IsTest {
			hasTest = true
		}

		categoryCounts[cl.classification.Category]++
		// Track highest priority
		if priorityHigher(cl.classification.Priority, highestPriority) {
			highestPriority = cl.classification.Priority
		}

		// Update global stats
		p.stats.categoryCounts[cl.classification.Category]++

		p.stats.priorityCounts[cl.classification.Priority]++
		if cl.classification.IsTest {
			p.stats.testCount++
		} else {
			p.stats.prodCount++
		}
	}

	p.stats.totalClones += occurrences
	p.stats.totalTokens += totalTokens

	// Find most common category
	maxCount := 0
	for cat, count := range categoryCounts {
		if count > maxCount {
			maxCount = count
			primaryCategory = cat
		}
	}

	// Get first clone's suggestion (they're usually similar)
	suggestion := ""
	if len(clones) > 0 {
		suggestion = clones[0].classification.Suggestion
	}

	// Build badges HTML
	categoryEmoji := primaryCategory.GetCategoryEmoji()
	priorityEmoji := highestPriority.GetPriorityEmoji()

	badgesHTML := fmt.Sprintf(
		`<span class="badge-category">%s %s</span><span class="badge-priority %s">%s %s</span>`,
		categoryEmoji, primaryCategory, highestPriority, priorityEmoji, highestPriority,
	)

	if hasTest {
		badgesHTML += `<span class="badge-test">🧪 test</span>`
	}

	_, err := fmt.Fprintf(
		p.w,
		`<div class="clone-group" data-category="%s" data-priority="%s" data-test="%t">
<div class="clone-header" onclick="this.parentElement.classList.toggle('collapsed')">
<h3>Clone Group #%d</h3>
<div class="badge-group">
%s
<span class="badge">%d occurrences · %d tokens</span>
</div>
</div>
<div class="clone-body">
%s
`,
		primaryCategory,
		highestPriority,
		hasTest,
		p.iota,
		badgesHTML,
		occurrences,
		totalTokens,
		suggestionHTML(suggestion),
	)

	return err
}

// priorityHigher returns true if p1 is higher priority than p2.
func priorityHigher(p1, p2 ClonePriority) bool {
	priorityOrder := map[ClonePriority]int{
		PriorityCritical: 4,
		PriorityHigh:     3,
		PriorityMedium:   2,
		PriorityLow:      1,
	}

	return priorityOrder[p1] > priorityOrder[p2]
}

func suggestionHTML(suggestion string) string {
	if suggestion == "" {
		return ""
	}

	return fmt.Sprintf(`<div class="suggestion">💡 %s</div>`, html.EscapeString(suggestion))
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
				fmt.Sprintf(
					"failed to process clone in group #%d (index=%d, file=%s)",
					p.iota,
					i,
					nstart.Filename,
				),
			)
		}

		// Calculate clone metrics for classification
		tokens := cnt
		lines := fileInfo.LineEnd - fileInfo.LineStart + 1
		nodeType := nstart.Type

		// Classify the clone for actionable reporting
		classification := ClassifyClone(fileInfo.Filename, nodeType, tokens, lines)

		clones[i] = clone{
			filename:       fileInfo.Filename,
			lineStart:      fileInfo.LineStart,
			lineEnd:        fileInfo.LineEnd,
			fragment:       extractContent(fileInfo, nstart, nend),
			size:           cnt,
			classification: classification,
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
			return err
		}
	}

	return nil
}

func (p *htmlprinter) writeCloneGroupFooter() error {
	_, err := fmt.Fprint(p.w, "</div></div>\n")

	return err
}

// writeDiffView renders the diff visualization for clone groups.
//
//nolint:gocyclo,cyclop // High cyclomatic complexity is inherent to HTML generation with multiple cases
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
	baseVSCode := fmt.Sprintf(
		"vscode://file/%s:%d",
		groupDiff.Base.Filename,
		groupDiff.Base.LineStart,
	)

	_, err := fmt.Fprintf(p.w, `
<div class="diff-mode">
<div class="diff-base-header">
<span class="label">📋 BASE REFERENCE</span>
<div class="file-info"><a href="%s">%s:%d</a></div>
`, baseVSCode, html.EscapeString(groupDiff.Base.Filename), groupDiff.Base.LineStart)
	if err != nil {
		return err
	}

	// Write aggregate diff stats if available
	//nolint:nestif // Complex but necessary HTML generation with conditional blocks
	if hasAggregateStats {
		_, err = fmt.Fprint(p.w, `<div class="diff-aggregate-stats">`)
		if err != nil {
			return err
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
			return err
		}
	}

	// Write view toggle buttons
	if err := p.writeDiffViewToggle(); err != nil {
		return err
	}

	// Write comparison selector for multiple clones
	if len(groupDiff.Others) > 1 {
		err := p.writeDiffSelector(groupDiff)
		if err != nil {
			return err
		}
	}

	// Write diff panels for each comparison
	for idx, other := range groupDiff.Others {
		err := p.writeDiffComparison(
			groupDiff.Base,
			other,
			idx,
			len(groupDiff.Others),
		)
		if err != nil {
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

	return err
}

// writeDiffViewToggle writes the view mode toggle buttons (side-by-side vs inline).
func (p *htmlprinter) writeDiffViewToggle() error {
	_, err := fmt.Fprintf(p.w, `
<div class="diff-view-toggle">
<button id="diff-toggle-%d-side" class="active" data-mode="side" onclick="toggleDiffView(%d, 'side')">◫ Side by Side</button>
<button id="diff-toggle-%d-inline" data-mode="inline" onclick="toggleDiffView(%d, 'inline')">▣ Inline</button>
</div>
`, p.iota, p.iota, p.iota, p.iota)

	return err
}

// writeDiffSelector writes the dropdown for selecting which clone to compare.
func (p *htmlprinter) writeDiffSelector(groupDiff CloneGroupDiff) error {
	_, err := fmt.Fprint(p.w, `
<div class="diff-selector">
<select id="diff-select-`+strconv.Itoa(p.iota)+`" onchange="showDiff(this.value, `+strconv.Itoa(p.iota)+`)">
<option value="" disabled selected>Compare with...</option>
`)
	if err != nil {
		return err
	}

	for idx, other := range groupDiff.Others {
		_, err := fmt.Fprintf(p.w, `<option value="%d">%s:%d</option>
`, idx, html.EscapeString(other.Filename), other.LineStart)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(p.w, `</select>
</div>
`)

	return err
}

// writeDiffComparison writes a side-by-side diff comparison.
func (p *htmlprinter) writeDiffComparison(
	base *CloneWithContent,
	other CloneDiff,
	index, total int,
) error {
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
		return err
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
		return err
	}

	// Write base and compared panels with word-level highlighting
	if err := p.writeDiffPanelsWithWordDiff(base, other); err != nil {
		return err
	}

	_, err = fmt.Fprint(p.w, `</div>
</div>
`)

	return err
}

// writeDiffPanelsWithWordDiff renders both base and compared panels with word-level highlighting.
func (p *htmlprinter) writeDiffPanelsWithWordDiff(_ *CloneWithContent, other CloneDiff) error {
	baseLines := other.Diff.Base
	comparedLines := other.Diff.Compared

	// Start base panel
	_, err := fmt.Fprint(p.w, `
<div class="diff-panel base">
<div class="diff-panel-title">Base Reference</div>
<pre><code>
`)
	if err != nil {
		return err
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
		return err
	}

	// Render compared panel
	if err := p.renderDiffLines(comparedLines, baseLines, false); err != nil {
		return err
	}

	_, err = fmt.Fprint(p.w, `</code></pre>
</div>
`)

	return err
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
		case DiffLineEqual:
			typeClass = "equal"
		}

		lineNum := fmt.Sprintf(`<span class="diff-line-num">%d</span>`, line.LineNumber)

		var content string

		if line.Type == DiffLineModified && i < len(oppositeLines) &&
			oppositeLines[i].Type == DiffLineModified {
			// For modified lines, show word-level diff
			if isBasePanel {
				content = WordDiff(line.Content, oppositeLines[i].Content)
			} else {
				content = WordDiff(oppositeLines[i].Content, line.Content)
			}
		} else {
			content = html.EscapeString(line.Content)
		}

		_, err := fmt.Fprintf(
			p.w,
			`<div class="diff-line %s">%s<span class="diff-line-content">%s</span></div>
`,
			typeClass,
			lineNum,
			content,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// countDiffStats counts the number of added, removed, and modified lines.
//
//nolint:nonamedreturns // Named returns are appropriate for counting functions
func countDiffStats(diff DiffResult) (added, removed, modified int) {
	for _, line := range diff.Compared {
		switch line.Type {
		case DiffLineAdded:
			added++
		case DiffLineRemoved:
			removed++
		case DiffLineModified:
			modified++
		case DiffLineEqual:
			// No action needed
		}
	}

	return added, removed, modified
}

// buildSummarySection generates the HTML summary section with category/priority distribution
// and filter buttons.
//

func (p *htmlprinter) buildSummarySection() string {
	if p.stats.totalClones == 0 {
		return ""
	}

	var sb strings.Builder

	// Summary section header
	sb.WriteString(`<div class="summary-section">`)
	sb.WriteString(`<h2>📊 Summary</h2>`)

	// Overview stats
	sb.WriteString(`<div class="summary-grid">`)
	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Total Clones</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(p.stats.totalClones))
	sb.WriteString(`</span></div>`)
	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Total Tokens</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(p.stats.totalTokens))
	sb.WriteString(`</span></div>`)
	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Production</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(p.stats.prodCount))
	sb.WriteString(`</span></div>`)
	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Test Code</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(p.stats.testCount))
	sb.WriteString(`</span></div>`)
	sb.WriteString(`</div>`)

	// Category breakdown
	if len(p.stats.categoryCounts) > 0 {
		sb.WriteString(
			`<div class="summary-category"><h3>By Category</h3><div class="category-list">`,
		)

		for _, cat := range orderedCategories() {
			if count := p.stats.categoryCounts[cat]; count > 0 {
				sb.WriteString(`<span class="category-tag">`)
				sb.WriteString(cat.GetCategoryEmoji())
				sb.WriteString(" ")
				sb.WriteString(string(cat))
				sb.WriteString(": ")
				sb.WriteString(strconv.Itoa(count))
				sb.WriteString(`</span>`)
			}
		}

		sb.WriteString(`</div></div>`)
	}

	// Priority breakdown
	if len(p.stats.priorityCounts) > 0 {
		sb.WriteString(
			`<div class="summary-category"><h3>By Priority</h3><div class="priority-list">`,
		)

		for _, pri := range orderedPriorities() {
			if count := p.stats.priorityCounts[pri]; count > 0 {
				sb.WriteString(`<span class="priority-tag priority-`)
				sb.WriteString(string(pri))
				sb.WriteString(`">`)
				sb.WriteString(pri.GetPriorityEmoji())
				sb.WriteString(" ")
				sb.WriteString(string(pri))
				sb.WriteString(": ")
				sb.WriteString(strconv.Itoa(count))
				sb.WriteString(`</span>`)
			}
		}

		sb.WriteString(`</div></div>`)
	}

	// Filter buttons
	sb.WriteString(`<div class="filter-buttons"><h3>Filters</h3>`)
	sb.WriteString(
		`<button class="filter-btn active" data-filter="all" onclick="filterClones('all')">All</button>`,
	)
	sb.WriteString(
		`<button class="filter-btn" data-filter="prod" onclick="filterClones('prod')">📦 Production</button>`,
	)
	sb.WriteString(
		`<button class="filter-btn" data-filter="test" onclick="filterClones('test')">🧪 Test</button>`,
	)
	sb.WriteString(`<span class="filter-separator">|</span>`)

	for _, cat := range orderedCategories() {
		sb.WriteString(`<button class="filter-btn" data-filter="cat-`)
		sb.WriteString(string(cat))
		sb.WriteString(`" onclick="filterClones('cat-`)
		sb.WriteString(string(cat))
		sb.WriteString(`')">`)
		sb.WriteString(cat.GetCategoryEmoji())
		sb.WriteString(`</button>`)
	}

	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)

	return sb.String()
}

// orderedCategories returns categories in a stable order for display.
func orderedCategories() []CloneCategory {
	return []CloneCategory{
		CategoryFunction,
		CategoryMethod,
		CategoryHandler,
		CategoryStruct,
		CategoryInterface,
		CategoryLoop,
		CategoryConditional,
		CategoryAssignment,
		CategoryExpression,
		CategoryTest,
		CategoryUnknown,
	}
}

// orderedPriorities returns priorities in a stable order for display.
func orderedPriorities() []ClonePriority {
	return []ClonePriority{
		PriorityCritical,
		PriorityHigh,
		PriorityMedium,
		PriorityLow,
	}
}

func (p *htmlprinter) PrintFooter() error {
	summary := p.buildSummarySection()

	_, err := fmt.Fprint(p.w, summary+`
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

// Filter clones by test/prod or category
function filterClones(filter) {
	// Update active button state
	document.querySelectorAll('.filter-btn').forEach(function(btn) {
		btn.classList.remove('active');
		if (btn.dataset.filter === filter) {
			btn.classList.add('active');
		}
	});

	// Filter clone groups
	document.querySelectorAll('.clone-group').forEach(function(group) {
		if (filter === 'all') {
			group.style.display = '';
		} else if (filter === 'prod') {
			group.style.display = group.dataset.test === 'false' ? '' : 'none';
		} else if (filter === 'test') {
			group.style.display = group.dataset.test === 'true' ? '' : 'none';
		} else if (filter.startsWith('cat-')) {
			var cat = filter.substring(4);
			group.style.display = group.dataset.category === cat ? '' : 'none';
		}
	});
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

	// Move summary section to top (after header)
	var summary = document.querySelector('.summary-section');
	var statsGrid = document.querySelector('.stats-grid');
	if (summary && statsGrid) {
		statsGrid.parentNode.insertBefore(summary, statsGrid.nextSibling);
	}
});
</script>
</body>
</html>
`)

	return err
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
