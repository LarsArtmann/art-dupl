package artdupl

import (
	"time"

	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// convertToCloneGroup converts internal format to SDK CloneGroup format.
func (d *detector) convertToCloneGroup(hash string, frags [][]*syntax.Node, method DetectionMethod) *CloneGroup {
	clones := make([]*Clone, len(frags))
	totalSize := 0
	maxLines := 0

	for i, frag := range frags {
		clone := d.convertFragmentToClone(frag)
		clones[i] = clone
		totalSize += clone.Size
		lines := clone.EndLine - clone.StartLine + 1
		if lines > maxLines {
			maxLines = lines
		}
	}

	// Limit clones per group if specified
	if d.opts.MaxClonesPerGroup > 0 && len(clones) > d.opts.MaxClonesPerGroup {
		clones = clones[:d.opts.MaxClonesPerGroup]
	}

	return &CloneGroup{
		Hash:      hash,
		Clones:    clones,
		Size:      totalSize,
		LineCount: maxLines,
		Method:    method,
	}
}

// convertFragmentToClone converts a syntax fragment to SDK Clone format.
func (d *detector) convertFragmentToClone(frag []*syntax.Node) *Clone {
	if len(frag) == 0 {
		return &Clone{}
	}

	// Get file information from first node
	firstNode := frag[0]
	lastNode := frag[len(frag)-1]

	clone := &Clone{
		Filename:  firstNode.Filename,
		StartLine: int(firstNode.Pos),
		EndLine:   int(lastNode.End),
		StartPos:  int(firstNode.Pos),
		EndPos:    int(lastNode.End),
		Size:      len(frag),
	}

	// Include fragment content if requested
	if d.opts.IncludeFragments {
		clone.Fragment = d.extractFragmentContent(frag)
	}

	return clone
}

// extractFragmentContent extracts the actual source code for a fragment.
func (d *detector) extractFragmentContent(frag []*syntax.Node) string {
	if len(frag) == 0 {
		return ""
	}

	// Read the source file
	content, err := d.opts.FileReader(frag[0].Filename)
	if err != nil {
		d.logger.Warn("Failed to read file %s: %v", frag[0].Filename, err)
		return "[content unavailable]"
	}

	// Extract the relevant lines
	lines := position.SplitLines(content)
	start := frag[0].Pos - 1 // Convert to 0-based
	end := frag[len(frag)-1].End

	if start < 0 || int(end) >= len(lines) {
		return "[content unavailable]"
	}

	var fragmentLines []string
	for i := int(start); i <= int(end) && i < len(lines); i++ {
		fragmentLines = append(fragmentLines, lines[i])
	}

	return position.JoinLines(fragmentLines)
}

// buildResult creates final Result structure.
func (d *detector) buildResult(cloneGroups []*CloneGroup, fileCount int) *Result {
	analysisTime := time.Since(d.started)

	// Calculate summary statistics
	totalClones := 0
	for _, group := range cloneGroups {
		totalClones += len(group.Clones)
	}

	return &Result{
		CloneGroups: cloneGroups,
		Summary: &Summary{
			TotalFiles:    fileCount,
			TotalClones:   totalClones,
			TotalGroups:   len(cloneGroups),
			AnalysisTime:  analysisTime,
			MethodsUsed:   d.opts.DetectionMethods,
			LinesAnalyzed: 0, // Calculate actual lines analyzed
		},
		Metadata: &Metadata{
			Version:    "1.0.0", // Get from build info
			Timestamp:  time.Now(),
			ConfigHash: d.hashConfig(d.opts),
			Toolchain:  "go", // Get actual version
		},
	}
}
