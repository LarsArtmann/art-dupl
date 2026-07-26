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
	if _, err := fmt.Fprintf(w, "Clone Diff Report\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "  New clones:      %d\n", len(report.New)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w,
		"  Suppressed:      %d (previously accepted)\n", len(report.Suppressed)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "  Resolved:        %d\n", len(report.Resolved)); err != nil {
		return err
	}

	if len(report.New) > 0 {
		if _, err := fmt.Fprintln(w, "\n--- New Clones ---"); err != nil {
			return err
		}

		for _, group := range report.New {
			shortHash := group.Hash
			if len(shortHash) > 12 {
				shortHash = shortHash[:12]
			}

			if _, err := fmt.Fprintf(
				w,
				"  [%s] %d clones, %d tokens\n",
				shortHash, len(group.Clones), group.TokenCount,
			); err != nil {
				return err
			}

			for _, clone := range group.Clones {
				if _, err := fmt.Fprintf(
					w,
					"    %s:%d-%d\n",
					clone.Filename, clone.LineStart, clone.LineEnd,
				); err != nil {
					return err
				}
			}
		}
	}

	if len(report.Resolved) > 0 {
		if _, err := fmt.Fprintln(w, "\n--- Resolved Clones ---"); err != nil {
			return err
		}

		for _, rc := range report.Resolved {
			shortHash := rc.Hash
			if len(shortHash) > 12 {
				shortHash = shortHash[:12]
			}

			if _, err := fmt.Fprintf(
				w,
				"  [%s] %d tokens, was in: %v\n",
				shortHash, rc.Tokens, rc.Files,
			); err != nil {
				return err
			}
		}
	}

	if !report.HasChanges() {
		if _, err := fmt.Fprintln(w,
			"\nNo changes detected. All current clones match the baseline."); err != nil {
			return err
		}
	}

	return nil
}
