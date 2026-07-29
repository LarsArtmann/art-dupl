package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/printer/stats"
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
	cmd.Flags().String("output-file", "", "write stats output to file instead of stdout")

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

	sp.SetFilterStats(totalFiltered, breakdown)

	sourceBreakdown := filterStats.SourceBreakdown()
	if len(sourceBreakdown) > 0 {
		sp.SetFilterSourceStats(sourceBreakdown)
	}
}

// runStats implements the stats command.
//

func runStats(c *cobra.Command, arguments []string) error {
	ctx := c.Context()
	formatStr, _ := c.Flags().GetString("format")
	outputFile, _ := c.Flags().GetString("output-file")

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
	writer, cleanup, err := createOutputWriter(outputFile)
	if err != nil {
		return err
	}

	defer cleanup()

	p := stats.NewStats(writer, os.ReadFile, mergedConfig.Threshold)

	configureStatsPrinter(p, format, mergedConfig, parseStats, filterStats, duration)

	// Build groups from matches and print
	groups := printer.BuildCloneGroups(duplChan)
	groups = printer.EliminateOverlaps(groups)
	keys := getSortedKeys(groups, config.SortByHash)

	err = printHeader(p, config.SortByHash, mergedConfig.Threshold)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats header")
	}

	err = printCloneGroups(p, os.ReadFile, groups, keys, config.SortByHash, mergedConfig.DetectionMode.IsSemantic(),
		buildSuppressionConfig(mergedConfig))
	if err != nil {
		return err
	}

	footerErr := p.PrintFooter()
	if footerErr != nil {
		return duplerrors.Wrap(footerErr, duplerrors.AnalysisError, "failed to print stats footer")
	}

	return nil
}

// configureStatsPrinter applies stats configuration and filter stats to the printer.
func configureStatsPrinter(
	p printer.Printer,
	format config.OutputFormat,
	mergedConfig *config.Config,
	parseStats job.ParseStats,
	filterStats *FilterStats,
	duration time.Duration,
) {
	detectionMethodStr := detectionMethodsToString(mergedConfig.DetectionMethods)

	sp, ok := p.(printer.StatsPrinter)
	if !ok {
		return
	}

	sp.ApplyStatsConfig(printer.StatsConfig{
		Format:              format,
		FilesCount:          parseStats.FilesCount,
		DetectionMethods:    detectionMethodStr,
		SemanticDetection:   mergedConfig.DetectionMode.IsSemantic(),
		Timestamp:           time.Now().UTC().Format(time.RFC3339),
		AnalysisDuration:    duration,
		TotalEstimatedLines: parseStats.LinesCount,
	})

	applyFilterStats(sp, filterStats)
}

// parseDuration parses a duration string using time package.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s) //nolint:wrapcheck // Simple pass-through for time.ParseDuration
}

// createOutputWriter returns an io.Writer for stats output.
// If filename is empty, it returns os.Stdout with a no-op cleanup.
// Otherwise it creates the file and returns a cleanup function that closes it.
func createOutputWriter(filename string) (io.Writer, func(), error) {
	if filename == "" {
		return os.Stdout, func() {}, nil
	}

	// #nosec G304 -- user-controlled output path
	f, err := os.Create(filename)
	if err != nil {
		return nil, nil, duplerrors.Wrap(err, duplerrors.IOError, "failed to create output file")
	}

	cleanup := func() {
		closeErr := f.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close file %q: %v\n", filename, closeErr)
		}
	}

	return f, cleanup, nil
}
