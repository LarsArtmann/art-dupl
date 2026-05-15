package cmd

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func printDupls(
	ctx context.Context,
	p printer.Printer,
	fread printer.ReadFile,
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

	err = printCloneGroups(p, fread, groups, keys, sortBy)
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

func getSortedKeys(groups map[string][][]*syntax.Node, sortBy config.SortCriteria) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	uniqueCounts := printer.ComputeUniqueCounts(groups)
	printer.SortCloneGroupKeys(keys, sortBy, groups, uniqueCounts)

	return keys
}

func printHeader(p printer.Printer, sortBy config.SortCriteria, threshold int) error {
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

func printCloneGroups(
	p printer.Printer,
	fread printer.ReadFile,
	groups map[string][][]*syntax.Node,
	keys []string,
	sortBy config.SortCriteria,
) error {
	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) <= 1 {
			continue
		}

		if hs, ok := p.(printer.HashSetter); ok {
			hs.SetHash(k)
		}

		clones, err := printer.ProcessClones(fread, uniq)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError,
				fmt.Sprintf("failed to process clones for hash %s", k))
		}

		err = p.PrintClones(domain.ProcessedCloneGroup{
			Hash:   k,
			Size:   totalSize(clones),
			Clones: clones,
		}, sortBy)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError,
				fmt.Sprintf("failed to print clones for hash %s (sortBy: %s)", k, sortBy.String()))
		}
	}

	return nil
}

func totalSize(clones []domain.ProcessedClone) int {
	total := 0
	for _, c := range clones {
		total += c.Size
	}

	return total
}

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

func printFooter(p printer.Printer) error {
	err := p.PrintFooter()
	if err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}

	return nil
}
