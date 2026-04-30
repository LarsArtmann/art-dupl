package printer

import (
	"fmt"
	"html"
	"strconv"
)

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
