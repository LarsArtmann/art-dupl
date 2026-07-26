package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/LarsArtmann/art-dupl/baseline"
	"github.com/LarsArtmann/art-dupl/domain"
)

// ResolvedClone represents a clone group that existed in the baseline
// but is no longer detected in the current scan.
type ResolvedClone struct {
	Hash   string   `json:"hash"`
	Files  []string `json:"files"`
	Tokens int      `json:"tokens"`
}

// DiffReport partitions clone groups relative to a baseline file.
type DiffReport struct {
	New        []domain.ProcessedCloneGroup `json:"new"`
	Suppressed []domain.ProcessedCloneGroup `json:"suppressed"`
	Resolved   []ResolvedClone              `json:"resolved"`
}

// NewDiffReport computes a diff between current clone groups and a baseline.
// Groups present in both the current scan and the baseline are "suppressed"
// (previously accepted). Groups in the current scan but not in the baseline
// are "new". Groups in the baseline but not in the current scan are "resolved".
func NewDiffReport(
	currentGroups []domain.ProcessedCloneGroup,
	bf *baseline.File,
) DiffReport {
	report := DiffReport{}

	currentHashes := make(map[string]bool, len(currentGroups))

	for _, group := range currentGroups {
		currentHashes[group.Hash] = true

		if bf.Has(group.Hash) {
			report.Suppressed = append(report.Suppressed, group)
		} else {
			report.New = append(report.New, group)
		}
	}

	for _, entry := range bf.Entries {
		if !currentHashes[entry.Hash] {
			report.Resolved = append(report.Resolved, ResolvedClone{
				Hash:   entry.Hash,
				Files:  entry.Files,
				Tokens: entry.Tokens,
			})
		}
	}

	return report
}

// HasChanges returns true if any clones are new or resolved.
func (d DiffReport) HasChanges() bool {
	return len(d.New) > 0 || len(d.Resolved) > 0
}

// PrintDiffText writes a human-readable diff summary to the writer.
func PrintDiffText(w io.Writer, report DiffReport) error {
	if err := writeDiffSummary(w, report); err != nil {
		return err
	}

	if err := writeNewClonesSection(w, report.New); err != nil {
		return err
	}

	if err := writeResolvedSection(w, report.Resolved); err != nil {
		return err
	}

	if !report.HasChanges() {
		if _, err := fmt.Fprintln(w,
			"\nNo changes detected. All current clones match the baseline."); err != nil {
			return err
		}
	}

	return nil
}

func writeDiffSummary(w io.Writer, report DiffReport) error {
	lines := []string{
		"Clone Diff Report",
		fmt.Sprintf("  New clones:      %d", len(report.New)),
		fmt.Sprintf("  Suppressed:      %d (previously accepted)", len(report.Suppressed)),
		fmt.Sprintf("  Resolved:        %d", len(report.Resolved)),
	}

	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))

	return err
}

func writeNewClonesSection(w io.Writer, groups []domain.ProcessedCloneGroup) error {
	if len(groups) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(w, "\n--- New Clones ---"); err != nil {
		return err
	}

	for _, group := range groups {
		if err := writeHashEntry(w, group.Hash,
			"  [%s] %d clones, %d tokens\n",
			len(group.Clones), group.TokenCount,
		); err != nil {
			return err
		}

		for _, clone := range group.Clones {
			if _, err := fmt.Fprintf(
				w, "    %s:%d-%d\n",
				clone.Filename, clone.LineStart, clone.LineEnd,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeResolvedSection(w io.Writer, resolved []ResolvedClone) error {
	if len(resolved) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(w, "\n--- Resolved Clones ---"); err != nil {
		return err
	}

	for _, rc := range resolved {
		if err := writeHashEntry(w, rc.Hash,
			"  [%s] %d tokens, was in: %v\n",
			rc.Tokens, rc.Files,
		); err != nil {
			return err
		}
	}

	return nil
}

func writeHashEntry(w io.Writer, hash string, format string, args ...any) error {
	short := truncHash(hash)
	all := append([]any{short}, args...)

	if _, err := fmt.Fprintf(w, format, all...); err != nil {
		return fmt.Errorf("write hash entry: %w", err)
	}

	return nil
}

func truncHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}

	return hash
}
