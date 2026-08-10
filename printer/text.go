package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/dustin/go-humanize"
)

// maxPreviewRunes limits the code preview shown after each clone location
// in text output so wide lines do not break the scannable file:line format.
const maxPreviewRunes = 60

type TextPrinter struct {
	ReadFile

	cnt              int
	w                io.Writer
	totalSize        int
	cloneGroups      [][]domain.ProcessedClone
	currentHash      string
	isFileDupe       bool
	richText         bool
	explain          bool
	diffHintFiles    []string
	suppressionStats SuppressionStats
}

func (p *TextPrinter) SetHash(hash string) {
	p.currentHash = hash
}

func (p *TextPrinter) SetFileDuplicate(isDupe bool) {
	p.isFileDupe = isDupe
}

func (p *TextPrinter) SetSuppressionStats(stats SuppressionStats) {
	p.suppressionStats = stats
}

func (p *TextPrinter) SetRichText(enabled bool) {
	p.richText = enabled
}

func (p *TextPrinter) SetExplain(enabled bool) {
	p.explain = enabled
}

func NewText(w io.Writer, fread ReadFile) Printer {
	return &TextPrinter{w: w, ReadFile: fread, cloneGroups: make([][]domain.ProcessedClone, 0)}
}

func (p *TextPrinter) PrintHeader() error { return nil }

func (p *TextPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	p.cnt++

	clones := SortGroupClones(group, sortBy...)

	groupCloneSize := totalFragmentSize(clones)

	isFileDupe := p.isFileDupe

	if err := p.writeGroupHeader(isFileDupe, clones); err != nil {
		return duplerrors.Wrap(err, duplerrors.InternalError, "write group header")
	}

	if len(clones) > 0 && clones[0].Classification.GenericsCandidate {
		if err := p.writeGenericsHint(clones[0].Classification); err != nil {
			return duplerrors.Wrap(err, duplerrors.InternalError, "write generics hint")
		}
	}

	if p.explain && len(clones) > 0 {
		if err := p.writeExplanation(clones[0].Classification, len(clones)); err != nil {
			return duplerrors.Wrap(err, duplerrors.InternalError, "write explanation")
		}
	}

	p.cloneGroups = append(p.cloneGroups, clones)
	p.totalSize += groupCloneSize

	sort.Sort(byNameAndLineProcessed(clones))

	return p.printCloneList(clones)
}

func (p *TextPrinter) writeGroupHeader(isFileDupe bool, clones []domain.ProcessedClone) error {
	if !isFileDupe || p.currentHash == "" {
		return p.writeCloneHeader(clones)
	}

	return p.writeFileDupeHeader(clones)
}

func (p *TextPrinter) writeFileDupeHeader(clones []domain.ProcessedClone) error {
	hashPrefix := p.currentHash
	if len(hashPrefix) > 12 {
		hashPrefix = hashPrefix[:12]
	}

	fileSizeStr := formatBytes(clones[0].FileSize)
	if _, err := fmt.Fprintf(
		p.w,
		"\U0001f4c4 FILE DUPLICATE | \U0001f517 [%s...] | %d files | %s\n\n",
		hashPrefix,
		len(clones),
		fileSizeStr,
	); err != nil {
		return fmt.Errorf(
			"write file duplicate header (hash: %s, files: %d): %w",
			hashPrefix,
			len(clones),
			err,
		)
	}

	for _, cl := range clones {
		p.diffHintFiles = append(p.diffHintFiles, cl.Filename)
	}

	return nil
}

func (p *TextPrinter) writeCloneHeader(clones []domain.ProcessedClone) error {
	if p.richText && len(clones) > 0 {
		return p.writeRichGroupHeader(len(clones), clones[0].Classification)
	}

	if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(clones)); err != nil {
		return fmt.Errorf("write clone count header (%d clones): %w", len(clones), err)
	}

	return nil
}

func (p *TextPrinter) PrintFooter() error {
	s := p.suppressionStats
	if s.DetectedTotal > s.Shown && s.DetectedTotal > 0 {
		if _, err := fmt.Fprintf(p.w, "\nDetected %d clone groups, %d shown", s.DetectedTotal, s.Shown); err != nil {
			return err
		}

		if s.SuppressedActionable > 0 || s.SuppressedOther > 0 || s.SuppressedGenerics > 0 {
			parts := make([]string, 0, 3)
			if s.SuppressedActionable > 0 {
				parts = append(parts, fmt.Sprintf("%d non-actionable", s.SuppressedActionable))
			}

			if s.SuppressedOther > 0 {
				parts = append(parts, fmt.Sprintf("%d filtered", s.SuppressedOther))
			}

			if s.SuppressedGenerics > 0 {
				parts = append(parts, fmt.Sprintf("%d non-generics", s.SuppressedGenerics))
			}

			if _, err := fmt.Fprintf(p.w, " (%s suppressed)", strings.Join(parts, ", ")); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(p.w, "."); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(p.w, "\nFound total %d clone groups.\n", p.cnt); err != nil {
			return err
		}
	}

	if len(p.diffHintFiles) >= 2 {
		file1 := p.diffHintFiles[0]
		file2 := p.diffHintFiles[1]

		if _, err := fmt.Fprintf(p.w, "\n\u2192 diff %s %s\n", file1, file2); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextPrinter) printCloneList(clones []domain.ProcessedClone) error {
	for _, cl := range clones {
		preview := p.previewFirstLine(cl)
		if _, err := fmt.Fprintf(p.w, "  %s:%d-%d%s\n", cl.Filename, cl.LineStart, cl.LineEnd, preview); err != nil {
			return err
		}
	}

	return nil
}

// previewFirstLine returns a one-line source preview prefixed with "  | " for
// inline display in text output, or an empty string when no source is
// available. It prefers the clone's Fragment (already populated by the
// pipeline) and falls back to reading the file via ReadFile when Fragment is
// empty. Long lines are truncated to maxPreviewRunes.
func (p *TextPrinter) previewFirstLine(cl domain.ProcessedClone) string {
	line := firstNonEmptyLine(cl.Fragment)
	if line == "" {
		line = p.previewFromFile(cl)
	}

	if line == "" {
		return ""
	}

	runes := []rune(line)
	if len(runes) > maxPreviewRunes {
		line = string(runes[:maxPreviewRunes-1]) + "…"
	}

	return "  | " + line
}

// firstNonEmptyLine returns the trimmed first non-whitespace line of s, or
// "" if every line is empty.
func firstNonEmptyLine(s string) string {
	for line := range strings.SplitSeq(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

// previewFromFile reads the clone's source file and returns the trimmed line
// at LineStart, or "" on any error (missing file, nil reader, out-of-range).
func (p *TextPrinter) previewFromFile(cl domain.ProcessedClone) string {
	if p.ReadFile == nil {
		return ""
	}

	data, err := p.ReadFile(cl.Filename)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	if cl.LineStart < 1 || cl.LineStart > len(lines) {
		return ""
	}

	return strings.TrimSpace(lines[cl.LineStart-1])
}

// tokenLineSummary formats token and line counts with correct pluralization.
func tokenLineSummary(tokens, lines int) string {
	return fmt.Sprintf("%d %s, %d %s", tokens, plural("token", tokens), lines, plural("line", lines))
}

func plural(word string, n int) string {
	if n == 1 {
		return word
	}

	return word + "s"
}

func (p *TextPrinter) writeRichGroupHeader(count int, cls domain.CloneClassification) error {
	actionabilityBadge := ""
	if cls.Actionability == domain.NonActionable {
		actionabilityBadge = " [non-actionable]"
	}

	if _, err := fmt.Fprintf(
		p.w,
		"found %d clones: [%s] [%s] %s%s (%s) suggestion: %s\n",
		count,
		cls.Priority,
		cls.CloneType,
		cls.Category,
		actionabilityBadge,
		tokenLineSummary(cls.Tokens, cls.Lines),
		cls.Suggestion,
	); err != nil {
		return duplerrors.Wrapf(
			err,
			duplerrors.IOError,
			"write rich group header (count=%d, category=%q, priority=%q)",
			count, cls.Category, cls.Priority,
		)
	}

	return nil
}

func (p *TextPrinter) writeGenericsHint(cls domain.CloneClassification) error {
	hint := cls.GenericsHint
	if hint == "" {
		hint = "same algorithm, different types; consider generics extraction"
	}

	if _, err := fmt.Fprintf(p.w, "  generics: %s\n", hint); err != nil {
		return fmt.Errorf("write generics hint: %w", err)
	}

	return nil
}

func (p *TextPrinter) writeExplanation(cls domain.CloneClassification, cloneCount int) error {
	parts := []string{string(cls.CloneType)}

	switch cls.Actionability {
	case domain.NonActionable:
		reason := cls.NonActionablePattern
		if reason == "" {
			reason = "boilerplate"
		}

		parts = append(parts, "non-actionable ("+reason+")")
	case domain.LowConfidence:
		reason := cls.NonActionablePattern
		if reason == "" {
			reason = "property engine"
		}

		parts = append(parts, "low-confidence ("+reason+")")
	case domain.Actionable:
		parts = append(parts, "actionable")
	default:
		parts = append(parts, "actionable")
	}

	parts = append(parts, string(cls.Category))
	parts = append(parts, tokenLineSummary(cls.Tokens, cls.Lines))

	if cls.Extractability.CanExtract {
		parts = append(parts, fmt.Sprintf("extractable: ~%d lines saved across %d sites",
			cls.Extractability.EstimatedLinesSaved, cloneCount))
	}

	if _, err := fmt.Fprintf(p.w, "  explain: %s\n", strings.Join(parts, " | ")); err != nil {
		return fmt.Errorf("write explanation line: %w", err)
	}

	if cls.Suggestion != "" {
		label := "fix"
		if cls.Actionability == domain.NonActionable {
			label = "why"
		}

		if _, err := fmt.Fprintf(p.w, "  %s: %s\n", label, cls.Suggestion); err != nil {
			return fmt.Errorf("write suggestion line: %w", err)
		}
	}

	return nil
}

func writeCloneLines(w io.Writer, clones []domain.ProcessedClone, format string) error {
	for _, cl := range clones {
		if _, err := fmt.Fprintf(w, format, cl.Filename, cl.LineStart, cl.LineEnd); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextPrinter) OutputText(threshold int, sortBy config.SortCriteria) error {
	sortedCloneGroups := make([][]domain.ProcessedClone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	sortGroupsByCriteria(sortedCloneGroups, sortBy, GroupMetrics[[]domain.ProcessedClone]{
		Size:  totalFragmentSize,
		Count: func(g []domain.ProcessedClone) int { return len(g) },
		SortKey: func(g []domain.ProcessedClone) string {
			if len(g) == 0 {
				return ""
			}

			return g[0].Filename
		},
	})

	err := p.PrintHeader()
	if err != nil {
		return fmt.Errorf(
			"print header (threshold: %d, sortBy: %s): %w",
			threshold,
			sortBy.String(),
			err,
		)
	}

	for _, cloneGroup := range sortedCloneGroups {
		for _, cl := range cloneGroup {
			if len(cl.Fragment) > 0 {
				if _, err := fmt.Fprintf(
					p.w,
					"%s\n%s:%d-%d\n\n",
					cl.Fragment,
					cl.Filename,
					cl.LineStart,
					cl.LineEnd,
				); err != nil {
					return cloneWriteErr(cl, "fragment", err)
				}
			} else {
				if err := writeCloneLines(
					p.w,
					[]domain.ProcessedClone{cl},
					"%s:%d-%d\n",
				); err != nil {
					return cloneWriteErr(cl, "line", err)
				}
			}
		}
	}

	return p.PrintFooter()
}

// cloneWriteErr wraps a write error with a "write clone X filename:line-line" prefix.
func cloneWriteErr(cl domain.ProcessedClone, kind string, err error) error {
	return fmt.Errorf("write clone %s %s:%d-%d: %w", kind, cl.Filename, cl.LineStart, cl.LineEnd, err)
}

// formatBytes returns a human-readable byte-size string using IEC binary units
// (KiB/MiB/GiB) via go-humanize. Bytes are rendered with the same units as
// `du -h` and `ls -lh` so output matches conventional Unix tool conventions.
// Negative values are formatted as their absolute value.
func formatBytes(bytes int) string {
	if bytes < 0 {
		bytes = -bytes
	}

	return humanize.IBytes(uint64(bytes))
}

// totalFragmentSize returns the total byte length of all clone fragments.
// This is used for display ordering by fragment size and is non-mutating.
func totalFragmentSize(clones []domain.ProcessedClone) int {
	total := 0
	for _, cl := range clones {
		total += len(cl.Fragment)
	}

	return total
}

type byNameAndLineProcessed []domain.ProcessedClone

func (c byNameAndLineProcessed) Len() int      { return len(c) }
func (c byNameAndLineProcessed) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byNameAndLineProcessed) Less(i, j int) bool {
	return compareByNameThenPos(c[i].Filename, c[j].Filename, c[i].LineStart, c[j].LineStart)
}
