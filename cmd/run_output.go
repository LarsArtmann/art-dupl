package cmd

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// printDupls prints duplicates using the specified printer.
func printDupls(
	ctx context.Context,
	p printer.Printer,
	duplChan <-chan syntax.Match,
	sortBy config.SortCriteria,
	threshold int,
	detectionMethod string,
) error {
	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}

	groups := printer.BuildCloneGroups(duplChan)
	keys := getSortedKeys(groups, sortBy)

	err := printHeader(p, sortBy, threshold)
	if err != nil {
		return err
	}

	err = printCloneGroups(p, groups, keys, sortBy)
	if err != nil {
		return err
	}

	err = handleJSONOutput(p, threshold, sortBy, detectionMethod)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}

	return printFooter(p)
}

// getSortedKeys extracts and sorts clone group keys.
func getSortedKeys(groups map[string][][]*syntax.Node, sortBy printer.SortBy) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	uniqueCounts := printer.ComputeUniqueCounts(groups)
	printer.SortCloneGroupKeys(keys, sortBy, groups, uniqueCounts)

	return keys
}

// printHeader prints the header with error wrapping.
func printHeader(p printer.Printer, sortBy printer.SortBy, threshold int) error {
	err := p.PrintHeader()
	if err != nil {
		return errors.Wrap(
			err,
			errors.AnalysisError,
			fmt.Sprintf(
				"failed to print header (sortBy: %s, threshold: %d)",
				sortBy.String(),
				threshold,
			),
		)
	}

	return nil
}

// printCloneGroups iterates over clone groups and prints them.
func printCloneGroups(
	p printer.Printer,
	groups map[string][][]*syntax.Node,
	keys []string,
	sortBy config.SortCriteria,
) error {
	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) <= 1 {
			continue
		}

		// Set hash for printers that support HashSetter interface
		if hs, ok := p.(printer.HashSetter); ok {
			hs.SetHash(k)
		}

		err := p.PrintClones(uniq, sortBy)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError,
				fmt.Sprintf("failed to print clones for hash %s (sortBy: %s)", k, sortBy.String()))
		}
	}

	return nil
}

// handleJSONOutput handles JSON-specific output if the printer is a JSONPrinter.
func handleJSONOutput(
	p printer.Printer,
	threshold int,
	sortBy config.SortCriteria,
	detectionMethod string,
) error {
	jsonPrinter, ok := p.(*printer.JSONPrinter)
	if !ok {
		return nil
	}

	err := jsonPrinter.OutputJSON(threshold, sortBy, detectionMethod)
	if err != nil {
		return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s): %w",
			threshold, sortBy.String(), err)
	}

	return nil
}

// printFooter prints the footer with error wrapping.
func printFooter(p printer.Printer) error {
	err := p.PrintFooter()
	if err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}

	return nil
}
