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
	semantic bool,
	suppressTestLow bool,
	testThreshold int,
) error {
	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}

	groups := printer.BuildCloneGroups(duplChan)
	keys := getSortedKeys(groups, sortBy)

	err := printHeader(p, sortBy, threshold)
	if err != nil {
		return fmt.Errorf("print header (threshold: %d, sortBy: %s, detection: %s): %w",
			threshold, sortBy.String(), detectionMethod, err)
	}

	err = printCloneGroups(p, fread, groups, keys, sortBy, semantic, suppressTestLow, testThreshold)
	if err != nil {
		return fmt.Errorf("print clone groups (fread: %v, sortBy: %s): %w", fread, sortBy.String(), err)
	}

	err = handleJSONOutput(p, threshold, sortBy, detectionMethod)
	if err != nil {
		return fmt.Errorf("json output (threshold: %d, sortBy: %s, detection: %s): %w",
			threshold, sortBy.String(), detectionMethod, err)
	}

	if ctx.Err() != nil {
		return fmt.Errorf(
			"context cancelled after output (threshold: %d, sortBy: %s): %w",
			threshold,
			sortBy.String(),
			ctx.Err(),
		)
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
	semantic bool,
	suppressTestLow bool,
	testThreshold int,
) error {
	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) <= 1 {
			continue
		}

		if semantic {
			if printer.EvaluateActionability(uniq) == domain.NonActionable {
				continue
			}
		}

		if hs, ok := p.(printer.HashSetter); ok {
			hs.SetHash(k)
		}

		clones, err := printer.ProcessClones(fread, uniq)
		if err != nil {
			return errors.Wrapf(err, errors.AnalysisError,
				"failed to process clones for hash %s", k)
		}

		group := domain.ProcessedCloneGroup{
			Hash:   k,
			Clones: clones,
		}
		group.TokenCount = group.TotalTokenCount()

		if shouldSuppressGroup(group, suppressTestLow, testThreshold) {
			continue
		}

		err = p.PrintClones(group, sortBy)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError,
				fmt.Sprintf("failed to print clones for hash %s (sortBy: %s)", k, sortBy.String()))
		}
	}

	return nil
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
		return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s, detection: %s): %w",
			threshold, sortBy.String(), detectionMethod, err)
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

// shouldSuppressGroup checks if a clone group should be filtered out based on
// suppress-test-low and test-threshold settings.
func shouldSuppressGroup(
	group domain.ProcessedCloneGroup,
	suppressTestLow bool,
	testThreshold int,
) bool {
	if len(group.Clones) == 0 {
		return false
	}

	cls := group.Clones[0].Classification

	if suppressTestLow && cls.IsTest && cls.Priority == domain.PriorityLow {
		return true
	}

	if testThreshold > 0 && cls.IsTest && group.TokenCount < testThreshold {
		return true
	}

	return false
}
