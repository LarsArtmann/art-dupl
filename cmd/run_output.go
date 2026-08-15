package cmd

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/printer/actionability"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SuppressionConfig groups the filter thresholds used to suppress clone groups.
// Passing it as a single value avoids 3-parameter function signatures that are
// prone to argument-swap bugs.
type SuppressionConfig struct {
	SuppressTestLow  bool
	TestThreshold    int
	MinLines         int
	MinTokens        int
	// GenericsMinLines gates generics-extraction candidacy (classification-time
	// precision filter, unlike the fields above which suppress whole groups).
	// 0 disables the gate.
	GenericsMinLines int
	AcceptDirectives *AcceptedSet
	NoActionability  bool
	ShowSuppressed   bool
	DisabledPatterns map[actionability.PatternLabel]bool
}

// buildSuppressionConfig constructs a complete SuppressionConfig from the
// effective configuration. Every subcommand that calls printCloneGroups or
// printDupls MUST use this helper instead of building a struct literal —
// a truncated literal silently disables accept directives, actionability
// filtering, and disabled-pattern suppression.
func buildSuppressionConfig(cfg *config.Config) SuppressionConfig {
	return SuppressionConfig{
		SuppressTestLow:  cfg.EffectiveSuppressTestLow(),
		TestThreshold:    cfg.EffectiveTestThreshold(),
		MinLines:         cfg.MinLines,
		MinTokens:        cfg.MinTokens,
		GenericsMinLines: cfg.SuggestGenericsMinLines,
		AcceptDirectives: newAcceptSet(cfg),
		NoActionability:  cfg.NoActionability,
		ShowSuppressed:   cfg.ShowSuppressed,
		DisabledPatterns: buildDisabledPatternSet(cfg.DisabledPatterns),
	}
}

func printDupls(
	ctx context.Context,
	p printer.Printer,
	fread printer.ReadFile,
	duplChan <-chan syntax.Match,
	sortBy config.SortCriteria,
	threshold int,
	detectionMethod string,
	semantic bool,
	suppression SuppressionConfig,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	groups := printer.BuildCloneGroups(duplChan)
	groups = printer.EliminateOverlaps(groups)
	keys := getSortedKeys(groups, sortBy)

	err := printHeader(p, sortBy, threshold)
	if err != nil {
		return fmt.Errorf("print header (threshold: %d, sortBy: %s, detection: %s): %w",
			threshold, sortBy.String(), detectionMethod, err)
	}

	err = printCloneGroups(p, fread, groups, keys, sortBy, semantic, suppression)
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
	suppression SuppressionConfig,
) error {
	var stats printer.SuppressionStats

	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) <= 1 {
			continue
		}

		stats.DetectedTotal++

		if semantic && !suppression.NoActionability {
			result := actionability.EvaluateActionabilityWithDisabled(
				printer.ToCloneNodeSeqs(uniq), suppression.DisabledPatterns,
			)
			if result == domain.NonActionable {
				stats.SuppressedActionable++

				if !suppression.ShowSuppressed {
					continue
				}
			}
		}

		if hs, ok := p.(printer.HashSetter); ok {
			hs.SetHash(k)
		}

		clones, err := printer.ProcessClones(fread, uniq, printer.WithGenericsMinLines(suppression.GenericsMinLines))
		if err != nil {
			return errors.Wrapf(err, errors.AnalysisError,
				"failed to process clones for hash %s", k)
		}

		group := domain.NewProcessedCloneGroup(k, clones)

		if shouldSuppressGroup(group, suppression) {
			stats.SuppressedOther++

			if !suppression.ShowSuppressed {
				continue
			}
		}

		stats.Shown++

		err = p.PrintClones(group, sortBy)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError,
				fmt.Sprintf("failed to print clones for hash %s (sortBy: %s)", k, sortBy.String()))
		}
	}

	if ss, ok := p.(printer.SuppressionStatsSetter); ok {
		ss.SetSuppressionStats(stats)
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
// suppress-test-low, test-threshold, and min-lines settings.
func shouldSuppressGroup(
	group domain.ProcessedCloneGroup,
	suppression SuppressionConfig,
) bool {
	if len(group.Clones) == 0 {
		return false
	}

	if suppression.AcceptDirectives != nil && suppression.AcceptDirectives.IsAccepted(group) {
		return true
	}

	cls := group.Clones[0].Classification

	if suppression.SuppressTestLow && cls.IsTest && cls.Priority == domain.PriorityLow {
		return true
	}

	if suppression.TestThreshold > 0 && cls.IsTest && group.TokenCount < suppression.TestThreshold {
		return true
	}

	if suppression.MinLines > 0 && minCloneLineCount(group) < suppression.MinLines {
		return true
	}

	if suppression.MinTokens > 0 && minCloneTokenCount(group) < suppression.MinTokens {
		return true
	}

	return false
}

// minCloneTokenCount returns the smallest TokenCount across all clones in a group.
// Used by min-tokens filtering: if ANY clone has fewer tokens than the threshold,
// the entire group is suppressed.
func minCloneTokenCount(group domain.ProcessedCloneGroup) int {
	if len(group.Clones) == 0 {
		return 0
	}

	result := group.Clones[0].TokenCount
	for _, c := range group.Clones[1:] {
		if c.TokenCount < result {
			result = c.TokenCount
		}
	}

	return result
}

// minCloneLineCount returns the smallest LineCount across all clones in a group.
// Used by min-lines filtering: if ANY clone is shorter than the threshold, the
// entire group is suppressed.
func minCloneLineCount(group domain.ProcessedCloneGroup) int {
	if len(group.Clones) == 0 {
		return 0
	}

	result := group.Clones[0].LineCount()
	for _, c := range group.Clones[1:] {
		if lc := c.LineCount(); lc < result {
			result = lc
		}
	}

	return result
}
