package cmd

import (
	"context"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/gogenfilter"
	"github.com/spf13/cobra"
)

// NewStatsCommand creates the stats command.
func NewStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [flags] [paths...]",
		Short: "Show aggregated duplication statistics",
		Long:  "Prints comprehensive duplication statistics: clone counts, token totals, and file impact metrics in text, JSON, or CSV format.",
		Args:  cobra.ArbitraryArgs,
		RunE:  runStats,
	}

	addSharedFlags(cmd)

	// Stats-specific: output format
	cmd.Flags().StringP("format", "o", "text", "output format: text, json, csv (default: text)")

	return cmd
}

// applyFilterStats applies filter statistics to the stats config.
func applyFilterStats(sp printer.StatsPrinter, filterStats gogenfilter.FilterStats) {
	totalFiltered := filterStats.TotalFiltered()
	if totalFiltered <= 0 {
		return
	}

	breakdown := make(map[string]int)

	for _, reason := range gogenfilter.AllFilterReasons() {
		if reason == "not_filtered" {
			continue
		}

		count := filterStats.FilteredBy(reason)
		if count > 0 {
			breakdown[string(reason)] = count
		}
	}

	sp.ApplyStatsConfig(printer.StatsConfig{
		FilesFiltered:   totalFiltered,
		FilterBreakdown: breakdown,
	})
}

// runStats implements the stats command.
//
//nolint:funlen // Stats command requires handling many CLI flags and configuration options
func runStats(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	formatStr, _ := cmd.Flags().GetString("format")

	// Parse and validate format
	format, err := config.ParseOutputFormat(formatStr)
	if err != nil {
		return duplerrors.WrapValidation(err, "invalid format value")
	}

	mergedConfig, err := BuildConfigFromFlags(cmd, args)
	if err != nil {
		return err
	}

	// Add timeout context if specified
	var cancel context.CancelFunc

	ctx, cancel = utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	// Start profiling for timing
	startProfile := job.StartProfile()

	// Run analysis
	duplChan, parseStats, filterStats, err := executeAnalysis(
		ctx,
		mergedConfig,
		mergedConfig.Paths,
		format,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	// End profiling
	endProfile := job.EndProfile(startProfile)
	duration := endProfile.Duration

	// Create stats printer
	p := printer.NewStats(os.Stdout, os.ReadFile, mergedConfig.Threshold)

	// Configure stats printer
	detectionMethodStr := detectionMethodsToString(mergedConfig.DetectionMethods)

	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.ApplyStatsConfig(printer.StatsConfig{
			Format:              format,
			FilesCount:          parseStats.FilesCount,
			DetectionMethods:    detectionMethodStr,
			SemanticDetection:   mergedConfig.Semantic,
			Timestamp:           time.Now().UTC().Format(time.RFC3339),
			AnalysisDuration:    duration,
			TotalEstimatedLines: parseStats.LinesCount,
		})

		applyFilterStats(sp, filterStats)
	}

	// Build groups from matches
	groups := printer.BuildCloneGroups(duplChan)

	// Print stats
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	// Sort keys for deterministic output
	// Using SortByHash for consistent ordering across runs
	printer.SortCloneGroupKeys(keys, config.SortByHash, groups, nil)

	err = p.PrintHeader()
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats header")
	}

	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) > 1 {
			err := p.PrintClones(uniq, config.SortByHash)
			if err != nil {
				return duplerrors.Wrap(
					err,
					duplerrors.AnalysisError,
					"failed to process clones for hash "+k,
				)
			}
		}
	}

	err = p.PrintFooter()
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats footer")
	}

	return nil
}

// parseDuration parses a duration string using time package.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s) //nolint:wrapcheck // Simple pass-through for time.ParseDuration
}
