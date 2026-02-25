package cmd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// printDupls prints duplicates using the specified printer.
func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy printer.SortBy, threshold int, detectionMethod string) error {
	// Build groups from matches
	groups := printer.BuildCloneGroups(duplChan)

	// Get sorted keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	// Pre-compute unique counts for sorting
	uniqueCounts := printer.ComputeUniqueCounts(groups)

	// Sort clone groups based on sortBy criteria
	printer.SortCloneGroupKeys(keys, sortBy, groups, uniqueCounts)

	if err := p.PrintHeader(); err != nil {
		return errors.Wrap(err, errors.AnalysisError, fmt.Sprintf("failed to print header (sortBy: %s, threshold: %d)", sortBy.String(), threshold))
	}

	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) > 1 {
			if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
				jsonPrinter.SetHash(k)
			}
			if err := p.PrintClones(uniq, sortBy); err != nil {
				return errors.Wrap(err, errors.AnalysisError, fmt.Sprintf("failed to print clones for hash %s (sortBy: %s)", k, sortBy.String()))
			}
		}
	}

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		if err := jsonPrinter.OutputJSON(threshold, sortBy, detectionMethod); err != nil {
			return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s): %w", threshold, sortBy.String(), err)
		}
	}

	if err := p.PrintFooter(); err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}
	return nil
}
