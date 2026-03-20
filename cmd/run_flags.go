package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/spf13/cobra"
)

// wrapAnalysisError wraps an analysis error with context about the paths being analyzed.
func wrapAnalysisError(err error, paths []string) error {
	return duplerrors.Wrap(
		err,
		duplerrors.AnalysisError,
		fmt.Sprintf("analysis failed for paths %v", paths),
	)
}

// runCmd implements Cobra command execution.
//
//nolint:gocyclo,cyclop,funlen,gocognit // Command execution requires handling many CLI flags and configuration options
func runCmd(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verboseCount, _ := cmd.Flags().GetCount("verbose")
	verbose := verboseCount > 0
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")
	detectionMethods, _ := cmd.Flags().GetString("detection-methods")

	// Validate sorting criteria
	if _, err := printer.ParseSortBy(sortBy); err != nil {
		return duplerrors.WrapValidation(err, fmt.Sprintf("invalid --sort value %q", sortBy))
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	profile, _ := cmd.Flags().GetBool("profile")
	timeoutStr, _ := cmd.Flags().GetString("timeout")
	filterGenerated, _ := cmd.Flags().GetBool("filter-generated")
	includeSQLC, _ := cmd.Flags().GetBool("include-sqlc")
	includeTempl, _ := cmd.Flags().GetBool("include-templ")
	includePatterns, _ := cmd.Flags().GetStringArray("include-pattern")
	excludePatterns, _ := cmd.Flags().GetStringArray("exclude-pattern")

	// Incremental analysis flags
	incremental, _ := cmd.Flags().GetBool("incremental")
	since, _ := cmd.Flags().GetString("since")
	cacheDir, _ := cmd.Flags().GetString("cache-dir")
	clearCache, _ := cmd.Flags().GetBool("clear-cache")

	// Semantic detection flags
	semantic, _ := cmd.Flags().GetBool("semantic")
	structural, _ := cmd.Flags().GetBool("structural")
	diffModeStr, _ := cmd.Flags().GetString("diff")

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

	// Concurrent processing flag
	workers, _ := cmd.Flags().GetInt("workers")

	fileConfig, err := config.LoadOptionalConfig(configFile)
	if err != nil {
		return duplerrors.WrapConfig(err, fmt.Sprintf("loading config from file %q", configFile))
	}

	appConfig := &config.Config{}

	if verbose {
		appConfig.Verbose = true
	}

	// Parse and set detection methods
	parsedMethods, err := config.ParseDetectionMethods(detectionMethods)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("invalid detection methods %q", detectionMethods),
		)
	}

	appConfig.DetectionMethods = parsedMethods

	if threshold != 15 {
		appConfig.Threshold = threshold
	}

	if vendor {
		appConfig.IncludeVendor = vendor
	}

	includeNodeModules, _ := cmd.Flags().GetBool("include-node-modules")
	if includeNodeModules {
		appConfig.IncludeNodeModules = includeNodeModules
	}

	if files {
		appConfig.FilesFromStdin = files
	}

	switch {
	case html:
		appConfig.OutputFormat = config.OutputFormatHTML
	case plumbing:
		appConfig.OutputFormat = config.OutputFormatPlumbing
	case jsonFlag:
		appConfig.OutputFormat = config.OutputFormatJSON
	}

	if profile {
		appConfig.Profile = true
	}

	// Parse timeout duration (e.g., "30m", "1h", "2h30m")
	if timeoutStr != "30m" && timeoutStr != "" {
		duration, err := time.ParseDuration(timeoutStr)
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

	// Set incremental analysis configuration
	if incremental {
		appConfig.Incremental = true
	}

	if since != "" {
		appConfig.Since = since
	}

	if cacheDir != "" {
		appConfig.CacheDir = cacheDir
	}

	if clearCache {
		appConfig.ClearCache = true
	}

	if semantic {
		appConfig.Semantic = true
	}
	// Note: structural handling moved to after merge to properly override the true default
	if workers != 0 {
		appConfig.Workers = workers
	}

	if diffModeStr != "" {
		parsedDiffMode, err := config.ParseDiffMode(diffModeStr)
		if err != nil {
			return duplerrors.WrapValidation(err, fmt.Sprintf("invalid --diff value %q", diffModeStr))
		}
		appConfig.DiffMode = parsedDiffMode
	}

	if len(args) > 0 {
		appConfig.Paths = args
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if err = config.ValidateConfig(mergedConfig); err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
	}

	// Get context from Cobra (includes Fang's signal handling)
	ctx := cmd.Context()

	if allFlag {
		return runAllModes(ctx, mergedConfig, sortBy, outputDir)
	}
	// Add timeout context if specified
	ctx, cancel := utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	duplChan, parseStats, _, err := executeAnalysis(
		ctx,
		mergedConfig,
		mergedConfig.Paths,
		mergedConfig.OutputFormat,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	// Check for cancellation after analysis completes
	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}

	p := createPrinter(mergedConfig.OutputFormat, mergedConfig.Threshold, mergedConfig.DiffMode)(os.Stdout, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(parseStats.FilesCount)
	}

	// Convert detection methods to comma-separated string
	detectionMethodStr := detectionMethodsToString(mergedConfig.DetectionMethods)

	if err := printDupls(
		ctx,
		p,
		duplChan,
		printer.SortBy(sortBy),
		mergedConfig.Threshold,
		detectionMethodStr,
	); err != nil {
		return duplerrors.Wrap(
			err,
			duplerrors.AnalysisError,
			fmt.Sprintf(
				"failed to print duplicates (sortBy: %s, threshold: %d)",
				sortBy,
				mergedConfig.Threshold,
			),
		)
	}

	return nil
}
