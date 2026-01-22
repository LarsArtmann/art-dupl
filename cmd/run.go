package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/cli"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/spf13/cobra"
)

// runCmd implements Cobra command execution.
func runCmd(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verbose, _ := cmd.Flags().GetBool("verbose")
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")
	detectionMethods, _ := cmd.Flags().GetString("detection-methods")

	// Validate sorting criteria
	if _, err := printer.ParseSortBy(sortBy); err != nil {
		return fmt.Errorf("invalid --sort value %q: %w", sortBy, err)
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

	var fileConfig *config.Config
	var err error
	if configFile != "" {
		fileConfig, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("error loading config from file %q: %w", configFile, err)
		}
	}

	appConfig := &config.Config{}

	if verbose {
		appConfig.Verbose = true
	}

	// Parse and set detection methods
	parsedMethods, err := config.ParseDetectionMethods(detectionMethods)
	if err != nil {
		return fmt.Errorf("invalid detection methods %q: %w", detectionMethods, err)
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
			return fmt.Errorf("invalid timeout format %q (use '30m', '1h', etc.): %w", timeoutStr, err)
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
		return fmt.Errorf("configuration validation failed (paths: %v): %w", mergedConfig.Paths, err)
	}

	if allFlag {
		return runAllModes(mergedConfig, sortBy, outputDir)
	}

	// Add timeout context if specified
	ctx := context.Background()
	if mergedConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mergedConfig.Timeout)*time.Second)
		defer cancel()
		fmt.Fprintf(os.Stderr, "⏱️  Execution timeout: %ds\n", mergedConfig.Timeout)
		// Note: Timeout context is configured but not yet passed through the analysis pipeline
		// This would require refactoring executeAnalysis to accept and propagate context
	}
	// For now, timeout config is stored but not fully implemented in analysis
	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		return fmt.Errorf("analysis failed for paths %v: %w", mergedConfig.Paths, err)
	}

	p := createPrinter(mergedConfig.OutputFormat, mergedConfig.Threshold)(os.Stdout, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
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

	if err := printDupls(p, duplChan, printer.SortBy(sortBy), mergedConfig.Threshold, mergedConfig.OutputFormat, detectionMethodStr); err != nil {
		return fmt.Errorf("failed to print duplicates (sortBy: %s, threshold: %d): %w", sortBy, mergedConfig.Threshold, err)
	}

	return nil
}

// createPrinter returns the appropriate printer based on output format.
func createPrinter(outputFormat config.OutputFormat, threshold int) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewHTML(w, fread, threshold)
		}
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatSimpleJSON:
		return printer.NewJSON
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}

// buildSuffixTree builds a suffix tree from provided paths.
func buildSuffixTree(paths []string, verbose, filesFromStdin bool, filterParam *filter.Filter, includeVendor bool) (*suffixtree.STree, []*syntax.Node, int, error) {
	if verbose {
		fmt.Fprintln(os.Stderr, "Building suffix tree")
	} else {
		fmt.Fprint(os.Stderr, "    📖 Parsing files and building analysis tree...")
	}

	schan, filesCountChan := job.Parse(filesFeedWithOptions(paths, filesFromStdin, filterParam, includeVendor))
	t, data, done := job.BuildTree(schan)
	<-done

	filesCount := <-filesCountChan
	t.Update(&syntax.Node{Type: -1})

	if verbose {
		fmt.Fprintln(os.Stderr, "Searching for clones")
	} else {
		fmt.Fprintln(os.Stderr, " ✅")
	}

	return t, *data, filesCount, nil
}

// filesFeedWithOptions creates a channel of file paths with options.
func filesFeedWithOptions(paths []string, fromStdin bool, filter *filter.Filter, includeVendor bool) chan string {
	if fromStdin {
		fchan := make(chan string)
		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				path := strings.TrimPrefix(f, "./")
				// Apply filter if enabled
				if filter != nil && filter.ShouldFilter(path) {
					continue
				}
				fchan <- path
			}
			close(fchan)
		}()
		return fchan
	}
	return crawlPaths(paths, filter, includeVendor)
}

// crawlPaths walks paths and returns a channel of Go files.
func crawlPaths(paths []string, filter *filter.Filter, includeVendor bool) chan string {
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot stat %s: %v\n", path, err)
				os.Exit(1)
			}
			if !info.IsDir() {
				// Apply filter to single file
				if filter != nil && filter.ShouldFilter(path) {
					continue
				}
				fchan <- path
				continue
			}
			err = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
				// Check vendor flag
				if !includeVendor && (strings.HasPrefix(path, cli.VendorDirPrefix) ||
					strings.Contains(path, cli.VendorDirInPath)) {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					// Apply filter to file
					if filter != nil && filter.ShouldFilter(path) {
						return nil
					}
					fchan <- path
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
				os.Exit(1)
			}
		}
		close(fchan)
	}()
	return fchan
}

// executeAnalysis runs the core duplicate analysis logic.
func executeAnalysis(cfg *config.Config, paths []string) (chan syntax.Match, int, error) {
	var startProfile job.ProfileResult
	if cfg.Profile {
		startProfile = job.StartProfile()
		fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")
	}

	// Create filter based on config
	var filterParam *filter.Filter
	var filterOptions []filter.FilterOption

	// ALWAYS filter templ files by default (unless --include-templ is set)
	if !cfg.IncludeTempl {
		filterOptions = append(filterOptions, filter.FilterTempl)
	}

	// Auto-detect sqlc.yaml files and enable sqlc filtering if found
	// This provides "out of the box" support for sqlc generated code
	sqlcOutputDirs, err := filter.GetSQLOutputDirs(paths)
	if err != nil && cfg.Verbose {
		fmt.Fprintf(os.Stderr, "warning: failed to detect sqlc config: %v\n", err)
	}

	// Enable sqlc filtering if sqlc.yaml is detected AND --include-sqlc is not set
	if len(sqlcOutputDirs) > 0 && !cfg.IncludeSQLC {
		filterOptions = append(filterOptions, filter.FilterSQLC)
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-detected sqlc.yaml, filtering sqlc generated code\n")
			for _, dir := range sqlcOutputDirs {
				fmt.Fprintf(os.Stderr, "   - %s\n", dir)
			}
		}
	}

	// If --filter-generated is set, also filter sqlc files (unless --include-sqlc is set)
	if cfg.FilterGenerated && !cfg.IncludeSQLC {
		// sqlc filtering may already be enabled by auto-detection, avoid duplicate
		alreadyFilteringSQLC := slices.Contains(filterOptions, filter.FilterSQLC)
		if !alreadyFilteringSQLC {
			filterOptions = append(filterOptions, filter.FilterSQLC)
		}
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Extended auto-generated code filtering enabled (sqlc)\n")
		}
	}

	// Create the filter if there are any options or include/exclude patterns
	if len(filterOptions) > 0 || len(cfg.IncludePatterns) > 0 || len(cfg.ExcludePatterns) > 0 {
		filterParam = filter.NewFilter(true, filterOptions)
		filterParam.WithIncludePatterns(cfg.IncludePatterns)
		filterParam.WithExcludePatterns(cfg.ExcludePatterns)

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (templ files filtered by default)\n")
		}
	}

	t, data, filesCount, err := buildSuffixTree(paths, cfg.Verbose, cfg.FilesFromStdin, filterParam, cfg.IncludeVendor)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build suffix tree for paths %v: %w", paths, err)
	}

	multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
	duplChan := make(chan syntax.Match)

	go func() {
		defer close(duplChan)
		matches := multiDetector.FindDuplOver(cfg.Threshold)
		for match := range matches {
			duplChan <- match
		}
	}()

	if cfg.Profile {
		endProfile := job.EndProfile(startProfile)
		job.PrintProfileResult(endProfile)
	}

	return duplChan, filesCount, nil
}

// printDupls prints duplicates using the specified printer.
func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy printer.SortBy, threshold int, outputFormat config.OutputFormat, detectionMethod string) error {
	// Build groups from matches
	groups := printer.BuildCloneGroups(duplChan)

	// Get sorted keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	// Pre-compute unique counts for sorting
	uniqueCounts := printer.ComputeUniqueCounts(groups)

	// Sort clone groups based on sortBy criteria
	printer.SortCloneGroupKeys(keys, sortBy, groups, uniqueCounts)

	if err := p.PrintHeader(); err != nil {
		return fmt.Errorf("failed to print header (sortBy: %s, threshold: %d): %w", sortBy.String(), threshold, err)
	}

	for _, k := range keys {
		uniq := utils.Unique(groups[k])
		if len(uniq) > 1 {
			if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
				jsonPrinter.SetHash(k)
			}
			if err := p.PrintClones(uniq, sortBy); err != nil {
				return fmt.Errorf("failed to print clones for hash %s (sortBy: %s): %w", k, sortBy.String(), err)
			}
		}
	}

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		if err := jsonPrinter.OutputJSON(threshold, sortBy, detectionMethod); err != nil {
			return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s): %w", threshold, sortBy.String(), err)
		}
	}

	if err := p.PrintFooter(); err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}
	return nil
}

// runAllModes runs all detection methods and generates all output formats.
func runAllModes(cfg *config.Config, sortBy, outputDir string) error {
	// Set detection methods to all available methods
	cfg.DetectionMethods = config.AllDetectionMethods()

	// Set output directory if not specified
	if outputDir == "" {
		outputDir = "reports/art-dupl"
	}

	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
	}

	fmt.Fprintf(os.Stderr, "📂 Running all detection methods and generating all output formats in %s...\n", outputDir)

	// Run analysis once
	duplChan, filesCount, err := executeAnalysis(cfg, cfg.Paths)
	if err != nil {
		return fmt.Errorf("analysis failed for paths %v: %w", cfg.Paths, err)
	}

	// Convert channel to slice for reuse
	matches := collectMatches(duplChan)

	// Generate all output formats
	formats := config.AllOutputFormats()
	sortByEnum := printer.SortBy(sortBy)

	// Convert detection methods to comma-separated string
	detectionMethodStr := ""
	if len(cfg.DetectionMethods) > 0 {
		methods := make([]string, len(cfg.DetectionMethods))
		for i, dm := range cfg.DetectionMethods {
			methods[i] = dm.String()
		}
		detectionMethodStr = strings.Join(methods, ",")
	}

	for _, format := range formats {
		filename := filepath.Join(outputDir, "report."+string(format))
		//nolint:gosec //G304 filename is constructed from controlled config output dir and format
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", filename, err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close file %q: %v\n", filename, err)
			}
		}()

		p := createPrinter(format, cfg.Threshold)(file, os.ReadFile)

		if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
			jsonPrinter.SetFilesCount(filesCount)
		}

		// Create channel from matches for this printer
		matchChan := make(chan syntax.Match)
		go func() {
			defer close(matchChan)
			for _, match := range matches {
				matchChan <- match
			}
		}()

		if err := printDupls(p, matchChan, sortByEnum, cfg.Threshold, cfg.OutputFormat, detectionMethodStr); err != nil {
			return fmt.Errorf("failed to print %s format: %w", format, err)
		}

		fmt.Fprintf(os.Stderr, "  ✅ Generated %s\n", filename)
	}

	fmt.Fprintf(os.Stderr, "\n✨ All formats generated successfully!\n")
	return nil
}

// collectMatches collects all matches from a channel into a slice.
func collectMatches(matchChan <-chan syntax.Match) []syntax.Match {
	var matches []syntax.Match
	for match := range matchChan {
		matches = append(matches, match)
	}
	return matches
}
