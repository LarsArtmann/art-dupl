package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/spf13/cobra"
)

// NewStatsCommand creates the stats command.
func NewStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [flags] [paths...]",
		Short: "Show aggregated duplication statistics",
		Long: `stats displays aggregated statistics about code duplication in a project.

The stats command provides a comprehensive overview of duplication levels,
making it easy to compare different projects or track duplication over time.

Statistics include:
- Files scanned and clone groups found
- Total lines of code and duplicate lines
- Total duplicate tokens and complexity score
- Average clone size and impact score
- Clone size distribution
- Top files with most duplicates

Examples:
  art-dupl stats                    # Show stats for current directory (text format)
  art-dupl stats -f json .          # Show stats in JSON format
  art-dupl stats ./src ./lib        # Show stats for specific paths
  art-dupl stats -t 20 .            # Show stats with higher threshold
  art-dupl stats -f csv -t 50 .     # Show stats in CSV format`,
		Args: cobra.ArbitraryArgs,
		RunE: runStats,
	}

	// Add flags to stats command
	cmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	cmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	cmd.Flags().CountP("verbose", "v", "enable verbose logging (repeat for more verbosity)")
	cmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone (default: 15)")
	cmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	cmd.Flags().StringP("detection-methods", "m", "art-dupl", "detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)")
	cmd.Flags().Bool("profile", false, "enable performance profiling")
	cmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = cmd.Flags().MarkHidden("profile")
	_ = cmd.Flags().MarkHidden("timeout")
	cmd.Flags().Bool("filter-generated", false, "enable extended filtering of sqlc.dev generated code (templ files are always filtered by default)")
	cmd.Flags().Bool("include-sqlc", false, "include sqlc.dev generated files (requires --filter-generated)")
	cmd.Flags().Bool("include-templ", false, "include templ.guide generated files (templ files are filtered by default unless this flag is set)")
	cmd.Flags().StringArray("include-pattern", []string{}, "file patterns to always include (takes precedence over filter)")
	cmd.Flags().StringArray("exclude-pattern", []string{}, "additional file patterns to exclude")
	cmd.Flags().String("format", "text", "output format: text, json (default: text)")

	return cmd
}

// runStats implements the stats command.
func runStats(cmd *cobra.Command, args []string) error {
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

	// Parse and validate format
	format, err := printer.ParseFormat(formatStr)
	if err != nil {
		return duplerrors.WrapValidation(err, "invalid format value")
	}

	var fileConfig *config.Config
	if configFile != "" {
		fileConfig, err = config.LoadConfig(configFile)
		if err != nil {
			return duplerrors.WrapConfig(err, fmt.Sprintf("loading config from file %q", configFile))
		}
	}

	appConfig := &config.Config{}

	if verbose {
		appConfig.Verbose = true
	}

	// Parse and set detection methods
	parsedMethods, err := config.ParseDetectionMethods(detectionMethods)
	if err != nil {
		return duplerrors.WrapValidation(err, fmt.Sprintf("invalid detection methods %q", detectionMethods))
	}
	appConfig.DetectionMethods = parsedMethods

	if threshold != 15 {
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
			return duplerrors.WrapValidation(err, fmt.Sprintf("invalid timeout format %q (use '30m', '1h', etc.)", timeoutStr))
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

	if len(args) > 0 {
		appConfig.Paths = args
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if err = config.ValidateConfig(mergedConfig); err != nil {
		return duplerrors.WrapValidation(err, fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths))
	}

	// Run analysis
	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, fmt.Sprintf("analysis failed for paths %v", mergedConfig.Paths))
	}

	// Create stats printer
	p := printer.NewStats(os.Stdout, os.ReadFile, mergedConfig.Threshold)

	// Set file count
	if sp, ok := p.(interface{ SetFilesCount(int) }); ok {
		sp.SetFilesCount(filesCount)
	}

	// Set format
	if sp, ok := p.(interface{ SetFormat(printer.Format) }); ok {
		sp.SetFormat(format)
	}

	// Convert detection methods to comma-separated string
	detectionMethodStr := ""
	if len(mergedConfig.DetectionMethods) > 0 {
		methods := make([]string, len(mergedConfig.DetectionMethods))
		for i, dm := range mergedConfig.DetectionMethods {
			methods[i] = dm.String()
		}
		detectionMethodStr = strings.Join(methods, ",")
	}

	// Set detection methods
	if sp, ok := p.(interface{ SetDetectionMethods(string) }); ok {
		sp.SetDetectionMethods(detectionMethodStr)
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

	if err := p.PrintHeader(); err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats header")
	}

	for _, k := range keys {
		uniq := unique(groups[k])
		if len(uniq) > 1 {
			if err := p.PrintClones(uniq, printer.SortByHash); err != nil {
				return duplerrors.Wrap(err, duplerrors.AnalysisError, fmt.Sprintf("failed to process clones for hash %s", k))
			}
		}
	}

	if err := p.PrintFooter(); err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "failed to print stats footer")
	}

	return nil
}

// unique returns unique nodes by filename and position.
func unique(dups [][]*syntax.Node) [][]*syntax.Node {
	seen := make(map[string]bool)
	result := make([][]*syntax.Node, 0, len(dups))

	for _, dup := range dups {
		if len(dup) == 0 {
			continue
		}

		key := fmt.Sprintf("%s:%d-%d", dup[0].Filename, dup[0].Pos, dup[len(dup)-1].End)
		if !seen[key] {
			seen[key] = true
			result = append(result, dup)
		}
	}

	return result
}

// parseDuration parses a duration string using time package.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
