package artdupl

import (
	"runtime/debug"
	"time"

	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// convertToCloneGroup converts internal format to SDK CloneGroup format.
func (d *detector) convertToCloneGroup(
	hash string,
	frags [][]*syntax.Node,
	method DetectionMethod,
) *CloneGroup {
	// First pass: convert fragments and filter out invalid/nil clones
	validClones := make([]*Clone, 0, len(frags))
	totalSize := 0
	maxLines := 0

	for _, frag := range frags {
		clone := d.convertFragmentToClone(frag)
		if clone == nil {
			continue
		}

		err := clone.IsValid()
		if err != nil {
			d.logger.Warn("Skipping invalid clone: %v", err)

			continue
		}

		validClones = append(validClones, clone)
		totalSize += clone.Size

		lines := clone.LineEnd - clone.LineStart + 1
		if lines > maxLines {
			maxLines = lines
		}
	}

	// Skip groups with no valid clones
	if len(validClones) == 0 {
		return nil
	}

	// Limit clones per group if specified
	if d.opts.MaxClonesPerGroup > 0 && len(validClones) > d.opts.MaxClonesPerGroup {
		validClones = validClones[:d.opts.MaxClonesPerGroup]
	}

	return &CloneGroup{
		Hash:      hash,
		Clones:    validClones,
		Size:      totalSize,
		LineCount: maxLines,
		Method:    method,
	}
}

// convertFragmentToClone converts a syntax fragment to SDK Clone format.
// Returns nil for empty fragments (not valid clones).
func (d *detector) convertFragmentToClone(frag []*syntax.Node) *Clone {
	if len(frag) == 0 {
		return nil
	}

	firstNode := frag[0]
	lastNode := frag[len(frag)-1]

	startPos := int(firstNode.Pos)
	endPos := int(lastNode.End)

	startLine, endLine := startPos, endPos

	if d.opts.FileReader != nil {
		content, err := d.opts.FileReader(firstNode.Filename)
		if err == nil && len(content) > 0 {
			startLine, endLine = position.ByteRangeToLines(content, startPos, endPos)
		}
	}

	clone := &Clone{ //nolint:exhaustruct
		Filename:  firstNode.Filename,
		LineStart: startLine,
		LineEnd:   endLine,
		StartPos:  startPos,
		EndPos:    endPos,
		Size:      len(frag),
	}

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

	content, err := d.opts.FileReader(frag[0].Filename)
	if err != nil {
		d.logger.Warn("Failed to read file %s: %v", frag[0].Filename, err)

		return ""
	}

	start := frag[0].Pos
	end := frag[len(frag)-1].End

	if start < 0 || int(end) > len(content) || start >= end {
		return ""
	}

	return string(content[start:end])
}

// buildResult creates final Result structure.
func (d *detector) buildResult(cloneGroups []*CloneGroup, fileCount int) *Result {
	analysisTime := time.Since(d.started)

	// Filter out nil clone groups
	validGroups := make([]*CloneGroup, 0, len(cloneGroups))
	for _, group := range cloneGroups {
		if group != nil {
			validGroups = append(validGroups, group)
		}
	}

	// Calculate summary statistics
	totalClones := 0
	for _, group := range validGroups {
		totalClones += len(group.Clones)
	}

	return &Result{
		CloneGroups: validGroups,
		Summary: &Summary{
			TotalFiles:    fileCount,
			TotalClones:   totalClones,
			TotalGroups:   len(validGroups),
			AnalysisTime:  analysisTime,
			MethodsUsed:   d.opts.DetectionMethods,
			LinesAnalyzed: 0, // Calculate actual lines analyzed
		},
		Metadata: &Metadata{
			Version:    sdkVersion(),
			Timestamp:  time.Now(),
			ConfigHash: d.configDebugString(d.opts),
			Toolchain:  goVersion(),
		},
	}
}

func sdkVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}

func goVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "go"
	}

	return info.GoVersion
}
