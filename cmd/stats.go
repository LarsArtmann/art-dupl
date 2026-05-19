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
	"github.com/LarsArtmann/gogenfilter/v3"
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
	cmd.Flags().String("format", "text", "output format: text, json, csv (default: text)")

	return cmd
}

// applyFilterStats applies filter statistics to the stats config.
func applyFilterStats(sp printer.StatsPrinter, filterStats *FilterStats) {
	if filterStats == nil {
		return
	}

	totalFiltered := filterStats.TotalFiltered()
	if totalFiltered <= 0 {
		return
	}

	breakdown := make(map[string]int)

	for _, reason := range gogenfilter.AllFilterReasons() {
		if reason == "not_filtered" {
			continue
		}

		count := filterStats.FilteredBy(string(reason))
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

func runStats(c *cobra.Command, arguments []string) error {
	ctx := c.Context()
	formatStr, _ := c.Flags().GetString("format")

	// Parse and validate format
	format, err := config.ParseOutputFormat(formatStr)
	if err != nil {
		return duplerrors.WrapValidation(err, "invalid format value")
	}

	mergedConfig, err := BuildConfigFromFlags(c, arguments)
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

	// Build groups from matches and print
	groups := printer.BuildCloneGroups(duplChan)
	keys := getSortedKeys(groups, config.SortByHash)

	err = printHeader(p, config.SortByHash, mergedConfig.Threshold)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats header")
	}

	err = printCloneGroups(p, os.ReadFile, groups, keys, config.SortByHash, mergedConfig.Semantic)
	if err != nil {
		return err
	}

	footerErr := p.PrintFooter()
	if footerErr != nil {
		return duplerrors.Wrap(footerErr, duplerrors.AnalysisError, "failed to print stats footer")
	}

	return nil
}

// parseDuration parses a duration string using time package.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s) //nolint:wrapcheck // Simple pass-through for time.ParseDuration
}
