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
	"github.com/LarsArtmann/art-dupl/domain"
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
	dupls     [][]domain.ProcessedClone
	diffMode  config.DiffMode
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
		dupls:     make([][]domain.ProcessedClone, 0),
		diffMode:  diffMode,
		stats: classificationStats{
			categoryCounts: make(map[CloneCategory]int),
			priorityCounts: make(map[ClonePriority]int),
		},
		metadata: metadata,
	}
}

// htmlTemplate is the modernized HTML template with dark theme.

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

	if p.metadata.IncludeSQLC {
		parts = append(parts, `<span class="metadata-badge">📦 Include SQLC</span>`)
	}

	if p.metadata.IncludeTempl {
		parts = append(parts, `<span class="metadata-badge">📝 Include Templ</span>`)
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

func (p *htmlprinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	p.iota++

	clones := group.Clones
	SortProcessedClonesByCriteria(clones, ExtractSortCriteria(sortBy...))

	p.dupMutex.Lock()
	p.dupls = append(p.dupls, clones)
	p.dupMutex.Unlock()

	if err := p.writeCloneGroupHeader(clones); err != nil {
		return fmt.Errorf(
			"print clone group #%d (hash: %s, clones: %d): %w",
			p.iota,
			group.Hash,
			len(clones),
			err,
		)
	}

	sort.Sort(byNameAndLineProcessed(clones))

	if p.diffMode.IsEnabled() && len(clones) > 1 {
		err := p.writeDiffView(clones)
		if err != nil {
			return fmt.Errorf(
				"diff view for clone group #%d (hash: %s): %w",
				p.iota,
				group.Hash,
				err,
			)
		}
	} else {
		err := p.writeCloneOccurrences(clones)
		if err != nil {
			return fmt.Errorf(
				"clone occurrences for clone group #%d (hash: %s): %w",
				p.iota,
				group.Hash,
				err,
			)
		}
	}

	return p.writeCloneGroupFooter()
}

func (p *htmlprinter) writeCloneGroupHeader(clones []domain.ProcessedClone) error {
	// Calculate aggregate metrics from clones
	occurrences := len(clones)
	totalTokens := 0
	hasTest := false
	highestPriority := domain.PriorityLow
	primaryCategory := domain.CategoryUnknown
	categoryCounts := make(map[CloneCategory]int)

	for _, cl := range clones {
		totalTokens += cl.Classification.Tokens
		if cl.Classification.IsTest {
			hasTest = true
		}

		categoryCounts[cl.Classification.Category]++
		if priorityHigher(cl.Classification.Priority, highestPriority) {
			highestPriority = cl.Classification.Priority
		}

		p.stats.categoryCounts[cl.Classification.Category]++

		p.stats.priorityCounts[cl.Classification.Priority]++
		if cl.Classification.IsTest {
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
		suggestion = clones[0].Classification.Suggestion
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
		domain.PriorityCritical: 4,
		domain.PriorityHigh:     3,
		domain.PriorityMedium:   2,
		domain.PriorityLow:      1,
	}

	return priorityOrder[p1] > priorityOrder[p2]
}

func suggestionHTML(suggestion string) string {
	if suggestion == "" {
		return ""
	}

	return fmt.Sprintf(`<div class="suggestion">💡 %s</div>`, html.EscapeString(suggestion))
}

func (p *htmlprinter) writeCloneOccurrences(clones []domain.ProcessedClone) error {
	for i, cl := range clones {
		vscodeLink := fmt.Sprintf("vscode://file/%s:%d", cl.Filename, cl.LineStart)

		_, err := fmt.Fprintf(p.w, `<div class="occurrence">
<div class="file-link">
<a href="%s" title="Open in VSCode">%s:%d</a>
<button class="copy-btn" onclick="copyCode('code-%d-%d')">📋 Copy</button>
</div>
<pre><code id="code-%d-%d">%s</code></pre>
</div>
`, vscodeLink, html.EscapeString(cl.Filename), cl.LineStart, p.iota, i, p.iota, i,
			html.EscapeString(string(cl.Fragment)))
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
