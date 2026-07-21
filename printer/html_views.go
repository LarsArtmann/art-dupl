package printer

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/a-h/templ"
)

type CloneOccurrenceView struct {
	VSCodeLink templ.SafeURL
	Filename   string
	LineStart  int
	LineEnd    int
	Fragment   string
}

type CloneGroupView struct {
	GroupNum    int
	Category    CloneCategory
	Priority    ClonePriority
	HasTest     bool
	Occurrences int
	TotalTokens int
	BadgesHTML  string
	Suggestion  string
	Clones      []CloneOccurrenceView
}

type DiffView struct {
	GroupNum       int
	Base           *CloneWithContent
	Others         []CloneDiff
	AggregateStats DiffAggregateStats
}

type DiffAggregateStats struct {
	TotalAdded    int
	TotalRemoved  int
	TotalModified int
	HasStats      bool
}

type SummaryView struct {
	TotalClones    int
	TotalTokens    int
	ProdCount      int
	TestCount      int
	CategoryCounts map[CloneCategory]int
	PriorityCounts map[ClonePriority]int
	HasData        bool
}

func toCloneOccurrenceView(cl domain.ProcessedClone) CloneOccurrenceView {
	return CloneOccurrenceView{
		VSCodeLink: templ.SafeURL(fmt.Sprintf("vscode://file/%s:%d", cl.Filename, cl.LineStart)),
		Filename:   cl.Filename,
		LineStart:  cl.LineStart,
		LineEnd:    cl.LineEnd,
		Fragment:   cl.Fragment,
	}
}

func toCloneGroupView(
	groupNum int,
	clones []domain.ProcessedClone,
) CloneGroupView {
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
	}

	maxCount := 0

	for cat, count := range categoryCounts {
		if count > maxCount {
			maxCount = count
			primaryCategory = cat
		}
	}

	suggestion := ""
	if len(clones) > 0 {
		suggestion = clones[0].Classification.Suggestion
	}

	categoryEmoji := primaryCategory.GetCategoryEmoji()
	priorityEmoji := highestPriority.GetPriorityEmoji()

	badgesHTML := fmt.Sprintf(
		`<span class="badge-category">%s %s</span><span class="badge-priority %s">%s %s</span>`,
		categoryEmoji, primaryCategory, highestPriority, priorityEmoji, highestPriority,
	)

	if hasTest {
		badgesHTML += `<span class="badge-test">🧪 test</span>`
	}

	occurrenceData := make([]CloneOccurrenceView, 0, len(clones))
	for _, cl := range clones {
		occurrenceData = append(occurrenceData, toCloneOccurrenceView(cl))
	}

	return CloneGroupView{
		GroupNum:    groupNum,
		Category:    primaryCategory,
		Priority:    highestPriority,
		HasTest:     hasTest,
		Occurrences: occurrences,
		TotalTokens: totalTokens,
		BadgesHTML:  badgesHTML,
		Suggestion:  suggestion,
		Clones:      occurrenceData,
	}
}

func toDiffView(groupNum int, groupDiff CloneGroupDiff) DiffView {
	return DiffView{
		GroupNum: groupNum,
		Base:     groupDiff.Base,
		Others:   groupDiff.Others,
		AggregateStats: DiffAggregateStats{
			TotalAdded:    groupDiff.TotalAdded,
			TotalRemoved:  groupDiff.TotalRemoved,
			TotalModified: groupDiff.TotalModified,
			HasStats:      groupDiff.TotalAdded > 0 || groupDiff.TotalRemoved > 0 || groupDiff.TotalModified > 0,
		},
	}
}

func toSummaryView(stats classificationStats) SummaryView {
	return SummaryView{
		TotalClones:    stats.totalClones,
		TotalTokens:    stats.totalTokens,
		ProdCount:      stats.prodCount,
		TestCount:      stats.testCount,
		CategoryCounts: stats.categoryCounts,
		PriorityCounts: stats.priorityCounts,
		HasData:        stats.totalClones > 0,
	}
}

func buildMetadataBadgesHTML(meta ReportMetadata) string {
	var parts []string

	if meta.Semantic {
		parts = append(parts, `<span class="metadata-badge enabled">🔤 Semantic</span>`)
	} else {
		parts = append(parts, `<span class="metadata-badge disabled">🔤 Structural</span>`)
	}

	if len(meta.DetectionMethods) > 0 {
		parts = append(parts, fmt.Sprintf(`<span class="metadata-badge">🔍 %s</span>`,
			html.EscapeString(strings.Join(meta.DetectionMethods, ", "))))
	}

	if meta.SortBy != "" {
		parts = append(parts, fmt.Sprintf(`<span class="metadata-badge">📊 %s</span>`,
			html.EscapeString(meta.SortBy)))
	}

	if meta.IncludeSQLC {
		parts = append(parts, `<span class="metadata-badge">📦 Include SQLC</span>`)
	}

	if meta.IncludeTempl {
		parts = append(parts, `<span class="metadata-badge">📝 Include Templ</span>`)
	}

	return strings.Join(parts, "")
}

func buildSummaryHTML(data SummaryView) string {
	if !data.HasData {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(`<div class="summary-section">`)
	sb.WriteString(`<h2>📊 Summary</h2>`)

	sb.WriteString(`<div class="summary-grid">`)

	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Total Clones</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(data.TotalClones))
	sb.WriteString(`</span></div>`)

	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Total Tokens</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(data.TotalTokens))
	sb.WriteString(`</span></div>`)

	sb.WriteString(
		`<div class="summary-item"><span class="summary-label">Production</span><span class="summary-value">`,
	)
	sb.WriteString(strconv.Itoa(data.ProdCount))
	sb.WriteString(`</span></div>`)

	sb.WriteString(`<div class="summary-item"><span class="summary-label">Test Code</span><span class="summary-value">`)
	sb.WriteString(strconv.Itoa(data.TestCount))
	sb.WriteString(`</span></div>`)

	sb.WriteString(`</div>`)

	if len(data.CategoryCounts) > 0 {
		sb.WriteString(`<div class="summary-category"><h3>By Category</h3><div class="category-list">`)

		for _, cat := range orderedCategories() {
			if count := data.CategoryCounts[cat]; count > 0 {
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

	if len(data.PriorityCounts) > 0 {
		sb.WriteString(`<div class="summary-category"><h3>By Priority</h3><div class="priority-list">`)

		for _, pri := range orderedPriorities() {
			if count := data.PriorityCounts[pri]; count > 0 {
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

	sb.WriteString(`<div class="filter-buttons"><h3>Filters</h3>`)
	sb.WriteString(`<button class="filter-btn active" data-filter="all" onclick="filterClones('all')">All</button>`)
	sb.WriteString(`<button class="filter-btn" data-filter="prod" onclick="filterClones('prod')">📦 Production</button>`)
	sb.WriteString(`<button class="filter-btn" data-filter="test" onclick="filterClones('test')">🧪 Test</button>`)
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

	sb.WriteString(`<div class="collapse-controls">`)
	sb.WriteString(`<button class="collapse-btn" onclick="collapseAll(true)">⊟ Collapse All</button>`)
	sb.WriteString(`<button class="collapse-btn" onclick="collapseAll(false)">⊞ Expand All</button>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)

	return sb.String()
}

const (
	cssClassAdded    = "added"
	cssClassRemoved  = "removed"
	cssClassModified = "modified"
	cssClassEqual    = "equal"
)

func diffLineTypeClass(lineType DiffLineType) string {
	switch lineType {
	case DiffLineAdded:
		return cssClassAdded
	case DiffLineRemoved:
		return cssClassRemoved
	case DiffLineModified:
		return cssClassModified
	case DiffLineEqual:
		return cssClassEqual
	default:
		return ""
	}
}

func diffStatPrefix(cssClass string) string {
	switch cssClass {
	case cssClassAdded:
		return "+"
	case cssClassRemoved:
		return "-"
	case cssClassModified:
		return "~"
	default:
		return ""
	}
}

var diffStatTypes = []string{ //nolint:gochecknoglobals // static CSS class list
	cssClassAdded,
	cssClassRemoved,
	cssClassModified,
}

func diffStatValue(cssClass string, stats DiffAggregateStats) int {
	switch cssClass {
	case cssClassAdded:
		return stats.TotalAdded
	case cssClassRemoved:
		return stats.TotalRemoved
	case cssClassModified:
		return stats.TotalModified
	default:
		return 0
	}
}

func diffStatLabel(cssClass string) string {
	switch cssClass {
	case cssClassAdded:
		return "Added"
	case cssClassRemoved:
		return "Removed"
	case cssClassModified:
		return "Modified"
	default:
		return ""
	}
}

func diffLineContent(line DiffLine, index int, oppositeLines []DiffLine, isBasePanel bool) string {
	if line.Type == DiffLineModified && index < len(oppositeLines) &&
		oppositeLines[index].Type == DiffLineModified {
		if isBasePanel {
			return WordDiff(line.Content, oppositeLines[index].Content)
		}

		return WordDiff(oppositeLines[index].Content, line.Content)
	}

	return html.EscapeString(line.Content)
}

func safeVSCodeURL(filename string, lineStart int) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("vscode://file/%s:%d", filename, lineStart))
}
