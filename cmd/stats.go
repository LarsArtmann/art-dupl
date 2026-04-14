package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/cli"
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

	// Add flags to stats command
	cmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	cmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	cmd.Flags().CountP("verbose", "v", "enable verbose logging (repeat for more verbosity)")
	cmd.Flags().
		IntP("threshold", "t", cli.DefaultThreshold, "minimum token sequence size to consider as clone (default: 15)")
	cmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	cmd.Flags().
		StringP("detection-methods", "m", "art-dupl", "detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)")
	cmd.Flags().Bool("profile", false, "enable performance profiling")
	cmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = cmd.Flags().MarkHidden("profile")
	_ = cmd.Flags().MarkHidden("timeout")
	cmd.Flags().
		Bool("filter-generated", false, "enable filtering of sqlc.dev generated code (auto-detects sqlc.yaml in parent directories)")
	cmd.Flags().
		Bool("include-sqlc", false, "include sqlc.dev generated files (override auto-detection)")
	cmd.Flags().
		Bool("include-templ", false, "include templ.guide generated files (filtered by default)")
	cmd.Flags().
		StringArray("include-pattern", []string{}, "file patterns to always include (takes precedence over filter)")
	cmd.Flags().StringArray("exclude-pattern", []string{}, "additional file patterns to exclude")
	cmd.Flags().StringP("format", "o", "text", "output format: text, json, csv (default: text)")

	// Add semantic-aware detection flags
	cmd.Flags().
		Bool("semantic", false, "explicitly enable semantic-aware detection (already the default; use only to override config file)")
	cmd.Flags().
		Bool("structural", false, "disable semantic detection and use structural-only matching (may increase false positives) [opt-out from default]")

	return cmd
}

// runStats implements the stats command.
//
//nolint:gocognit,gocyclo,cyclop,funlen,maintidx // Stats command requires handling many CLI flags and configuration options
func runStats(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verbose, _ := cmd.Flags().GetBool("verbose")
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	detectionMethods, _ := cmd.Flags().GetString("detection-methods")
	profile, _ := cmd.Flags().GetBool("profile")
	timeoutStr, _ := cmd.Flags().GetString("timeout")
	filterGenerated, _ := cmd.Flags().GetBool("filter-generated")
	includeSQLC, _ := cmd.Flags().GetBool("include-sqlc")
	includeTempl, _ := cmd.Flags().GetBool("include-templ")
	includePatterns, _ := cmd.Flags().GetStringArray("include-pattern")
	excludePatterns, _ := cmd.Flags().GetStringArray("exclude-pattern")
	formatStr, _ := cmd.Flags().GetString("format")
	semantic, _ := cmd.Flags().GetBool("semantic")
	structural, _ := cmd.Flags().GetBool("structural")

	// Validate conflicting flags - only an error if both are explicitly set
	if semantic && structural {
		return duplerrors.NewValidationError(
			"cannot use both --semantic and --structural flags; these are mutually exclusive",
			nil,
		)
	}

	// Warn about --structural flag (opt-out from recommended default)
	if structural {
		fmt.Fprintf(
			os.Stderr,
			"Note: --structural flag disables semantic detection. This may increase false positives from similar-looking but semantically different code.\n",
		)
	}

	// Parse and validate format
	format, err := printer.ParseFormat(formatStr)
	if err != nil {
		return duplerrors.WrapValidation(err, "invalid format value")
	}

	fileConfig, err := config.LoadOptionalConfig(configFile)
	if err != nil {
		return duplerrors.WrapConfig(err, fmt.Sprintf("loading config from file %q", configFile))
	}

	appConfig := &config.Config{}

	if verbose {
		appConfig.Verbose = true
	}

	if err := setDetectionMethods(appConfig, detectionMethods); err != nil {
		return err
	}

	if threshold != cli.DefaultThreshold {
		appConfig.Threshold = threshold
	}

	if vendor {
		appConfig.IncludeVendor = vendor
	}

	if files {
		appConfig.FilesFromStdin = files
	}

	if profile {
		appConfig.Profile = true
	}

	// Parse timeout duration
	if timeoutStr != "30m" && timeoutStr != "" {
		duration, err := parseDuration(timeoutStr)
		if err != nil {
			return duplerrors.WrapValidation(
				err,
				fmt.Sprintf("invalid timeout format %q (use '30m', '1h', etc.)", timeoutStr),
			)
		}

		appConfig.Timeout = int(duration.Seconds())
	}

	// Set filter configuration
	if filterGenerated {
		appConfig.FilterGenerated = true
	}

	if includeSQLC {
		appConfig.IncludeSQLC = true
	}

	if includeTempl {
		appConfig.IncludeTempl = true
	}

	if len(includePatterns) > 0 {
		appConfig.IncludePatterns = includePatterns
	}

	if len(excludePatterns) > 0 {
		appConfig.ExcludePatterns = excludePatterns
	}

	if semantic {
		appConfig.Semantic = true
	}
	// Note: structural handling moved to after merge to properly override the true default

	if len(args) > 0 {
		appConfig.Paths = args
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	// Handle --structural flag to explicitly disable semantic detection (opt-out from default)
	if structural {
		mergedConfig.Semantic = false
	}

	err = config.ValidateConfig(mergedConfig)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
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

	// Set file count
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetFilesCount(parseStats.FilesCount)
	}

	// Set format
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetFormat(format)
	}

	// Convert detection methods to comma-separated string
	detectionMethodStr := detectionMethodsToString(mergedConfig.DetectionMethods)

	// Set detection methods
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetDetectionMethods(detectionMethodStr)
	}

	// Set semantic detection status
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetSemanticDetection(mergedConfig.Semantic)
	}

	// Set analysis timestamp and duration
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetTimestamp(time.Now().UTC().Format(time.RFC3339))
	}

	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetAnalysisDuration(duration)
	}

	// Set actual total lines from parsing
	if sp, ok := p.(printer.StatsPrinter); ok {
		sp.SetTotalEstimatedLines(parseStats.LinesCount)
	}

	// Set filter statistics
	if sp, ok := p.(printer.StatsPrinter); ok {
		totalFiltered := filterStats.TotalFiltered()
		if totalFiltered > 0 {
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

			sp.SetFilterStats(totalFiltered, breakdown)
		}
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
	printer.SortCloneGroupKeys(keys, printer.SortByHash, groups, nil)

	err = p.PrintHeader()
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats header")
	}

	for _, k := range keys {
		uniq := syntax.Unique(groups[k])
		if len(uniq) > 1 {
			err := p.PrintClones(uniq, printer.SortByHash)
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
