package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/baseline"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// runDiffReport runs clone detection, then compares the results against a
// baseline file and outputs a diff report showing new, suppressed, and
// resolved clone groups.
func runDiffReport(
	ctx context.Context,
	mergedConfig *config.Config,
	baselinePath string,
	useJSON bool,
) error {
	if !baseline.Exists(baselinePath) {
		return duplerrors.NewValidationError(
			fmt.Sprintf("baseline file %s not found; run `art-dupl baseline` first", baselinePath), nil,
		)
	}

	bf, err := baseline.Load(baselinePath)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.IOError, "loading baseline "+baselinePath)
	}

	duplChan, _, _, err := executeAnalysis(
		ctx, mergedConfig, mergedConfig.Paths, config.OutputFormatText,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	groups := printer.BuildCloneGroups(duplChan)
	groups = printer.EliminateOverlaps(groups)
	keys := getSortedKeys(groups, config.SortByHash)

	suppression := SuppressionConfig{
		SuppressTestLow:  mergedConfig.EffectiveSuppressTestLow(),
		TestThreshold:    mergedConfig.EffectiveTestThreshold(),
		MinLines:         mergedConfig.MinLines,
		AcceptDirectives: newAcceptSet(mergedConfig),
		NoActionability:  mergedConfig.NoActionability,
		DisabledPatterns: buildDisabledPatternSet(mergedConfig.DisabledPatterns),
	}

	currentGroups := collectCurrentGroups(keys, groups, mergedConfig, suppression)

	report := printer.NewDiffReport(currentGroups, bf)

	if useJSON {
		return outputDiffJSON(report)
	}

	if err := printer.PrintDiffText(os.Stdout, report); err != nil {
		return fmt.Errorf("write diff report: %w", err)
	}

	return nil
}

func collectCurrentGroups(
	keys []string,
	groups map[string][][]*syntax.Node,
	mergedConfig *config.Config,
	suppression SuppressionConfig,
) []domain.ProcessedCloneGroup {
	var currentGroups []domain.ProcessedCloneGroup

	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) <= 1 {
			continue
		}

		if mergedConfig.DetectionMode.IsSemantic() && !suppression.NoActionability {
			result := printer.EvaluateActionabilityWithDisabled(
				printer.ToCloneNodeSeqs(uniq), suppression.DisabledPatterns,
			)
			if result == domain.NonActionable {
				continue
			}
		}

		clones, err := printer.ProcessClones(os.ReadFile, uniq)
		if err != nil {
			continue
		}

		group := domain.NewProcessedCloneGroup(k, clones)

		if shouldSuppressGroup(group, suppression) {
			continue
		}

		currentGroups = append(currentGroups, group)
	}

	return currentGroups
}

func outputDiffJSON(report printer.DiffReport) error {
	//nolint:musttag // DiffReport has json tags; inner types are domain's responsibility
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal diff report: %w", err)
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return fmt.Errorf("write diff report: %w", err)
	}

	return nil
}
