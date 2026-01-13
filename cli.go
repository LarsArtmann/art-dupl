package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/cli"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
	"github.com/spf13/cobra"
)

func Run() int { //nolint:cyclop,funlen // Main CLI entry point with error handling paths
	flag.Usage = func() { fmt.Fprintln(os.Stderr, `Usage: art-dupl [flags] [paths]`) }
	cliCfg := cli.NewCLIConfig()

	flag.Parse()

	var fileConfig *config.Config
	var err error
	if *cliCfg.ConfigFile != "" {
		fileConfig, err = config.LoadConfig(*cliCfg.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
			return 1
		}
	}
	appConfig := &config.Config{}

	if cliCfg.IsVerbose() {
		appConfig.Verbose = true
	}

	threshold := cliCfg.GetThreshold()
	if threshold != cli.DefaultThreshold {
		appConfig.Threshold = threshold
	}

	if *cliCfg.Vendor {
		appConfig.IncludeVendor = *cliCfg.Vendor
	}
	if *cliCfg.Files {
		appConfig.FilesFromStdin = *cliCfg.Files
	}

	switch {
	case *cliCfg.HTML:
		appConfig.OutputFormat = config.OutputFormatHTML
	case *cliCfg.Plumbing:
		appConfig.OutputFormat = config.OutputFormatPlumbing
	case *cliCfg.JSONFlag:
		appConfig.OutputFormat = config.OutputFormatJSON
	}

	if flag.NArg() > 0 {
		appConfig.Paths = flag.Args()
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if err = config.ValidateConfig(mergedConfig); err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
		return 1
	}

	// Validate mutually exclusive output format flags
	if cli.ExitIfBothSet(cliCfg.HTML, cliCfg.Plumbing, "plumbing", "HTML") != 0 {
		return 1
	}
	if cli.ExitIfBothSet(cliCfg.HTML, cliCfg.JSONFlag, "HTML", "JSON") != 0 {
		return 1
	}
	if cli.ExitIfBothSet(cliCfg.Plumbing, cliCfg.JSONFlag, "plumbing", "JSON") != 0 {
		return 1
	}

	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analysis error: %v\n", err)
		os.Exit(1)
		return 1
	}

	outputWriter := os.Stdout
	var outputFile *os.File
	if mergedConfig.OutputFile != "" {
		file, err := os.Create(mergedConfig.OutputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
			return 1
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", closeErr)
			}
		}()
		outputFile = file
		outputWriter = file
	}

	p := createPrinter(mergedConfig.OutputFormat)(outputWriter, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, printer.SortBy(*cliCfg.SortBy), mergedConfig.Threshold); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if outputFile != nil {
			_ = outputFile.Close()
		}
		os.Exit(1) //nolint:gocritic // Defer already executed before this point
		return 1
	}
	return 0
}

func createPrinter(outputFormat config.OutputFormat) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return printer.NewHTML
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}

func buildSuffixTree(paths []string, verbose, filesFromStdin bool, filterParam *filter.Filter) (*suffixtree.STree, []*syntax.Node, int, error) {
	if verbose {
		log.Println("Building suffix tree")
	} else {
		fmt.Fprintf(os.Stderr, "    📖 Parsing files and building analysis tree...")
	}

	schan, filesCountChan := job.Parse(filesFeedWithOptions(paths, filesFromStdin, filterParam))
	t, data, done := job.BuildTree(schan)
	<-done

	filesCount := <-filesCountChan
	t.Update(&syntax.Node{Type: -1})

	if verbose {
		log.Println("Searching for clones")
	} else {
		fmt.Fprintf(os.Stderr, " ✅\n")
	}

	return t, *data, filesCount, nil
}

func filesFeedWithOptions(paths []string, fromStdin bool, filter *filter.Filter) chan string {
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
	return crawlPaths(paths, filter)
}

func crawlPaths(paths []string, filter *filter.Filter) chan string { //nolint:cyclop // Path crawling with multiple error handling paths
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
				// Check vendor flag using flag lookup to avoid redefinition
				vendorFlag := flag.Lookup("vendor")
				includeVendor := vendorFlag == nil || vendorFlag.Value.(flag.Getter).Get().(bool)
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

func executeAnalysis(cfg *config.Config, paths []string) (chan syntax.Match, int, error) {
	var startProfile job.ProfileResult
	if cfg.Profile {
		startProfile = job.StartProfile()
		fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")
	}

	// Create filter based on config
	var filterParam *filter.Filter
	if cfg.FilterGenerated {
		// Determine which types to filter out
		var filterOptions []filter.FilterOption
		if !cfg.IncludeSQLC {
			filterOptions = append(filterOptions, filter.FilterSQLC)
		}
		if !cfg.IncludeTempl {
			filterOptions = append(filterOptions, filter.FilterTempl)
		}

		filterParam = filter.NewFilter(true, filterOptions)
		filterParam.WithIncludePatterns(cfg.IncludePatterns)
		filterParam.WithExcludePatterns(cfg.ExcludePatterns)

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled\n")
		}
	}

	t, data, filesCount, err := buildSuffixTree(paths, cfg.Verbose, cfg.FilesFromStdin, filterParam)
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

// buildCloneGroups builds a map of hash to clone groups from matches.
func buildCloneGroups(duplChan <-chan syntax.Match) map[string][][]*syntax.Node {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	return groups
}

// computeUniqueCounts calculates unique file counts for each clone group.
func computeUniqueCounts(groups map[string][][]*syntax.Node) map[string]int {
	uniqueCounts := make(map[string]int)
	for k, v := range groups {
		uniqueCounts[k] = len(util.Unique(v))
	}
	return uniqueCounts
}

// sortCloneGroupKeys sorts clone group hashes based on specified criteria.
func sortCloneGroupKeys(keys []string, sortBy printer.SortBy, groups map[string][][]*syntax.Node, uniqueCounts map[string]int) {
	switch sortBy {
	case printer.SortByOccurrence:
		// Sort by number of unique files in each clone group (most files first, descending)
		sort.Slice(keys, func(i, j int) bool {
			return uniqueCounts[keys[i]] > uniqueCounts[keys[j]]
		})
	case printer.SortByHash:
		// Sort alphabetically by hash (ascending)
		sort.Strings(keys)
	case printer.SortBySize:
		// Sort by size of first clone in each group (largest first, descending)
		sort.Slice(keys, func(i, j int) bool {
			sizeI := 0
			if len(groups[keys[i]]) > 0 && len(groups[keys[i]][0]) > 0 {
				sizeI = groups[keys[i]][0][0].Owns
			}
			sizeJ := 0
			if len(groups[keys[j]]) > 0 && len(groups[keys[j]][0]) > 0 {
				sizeJ = groups[keys[j]][0][0].Owns
			}
			return sizeI > sizeJ
		})
	default:
		// For unrecognized criteria, sort alphabetically by hash
		sort.Strings(keys)
	}
}

func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy printer.SortBy, threshold int) error { //nolint:cyclop // Output formatting with multiple conditional paths
	// Build groups from matches
	groups := buildCloneGroups(duplChan)

	// Get sorted keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}

	// Pre-compute unique counts for sorting
	uniqueCounts := computeUniqueCounts(groups)

	// Sort clone groups based on sortBy criteria
	sortCloneGroupKeys(keys, sortBy, groups, uniqueCounts)

	if err := p.PrintHeader(); err != nil {
		return fmt.Errorf("failed to print header (sortBy: %s, threshold: %d): %w", sortBy.String(), threshold, err)
	}

	for _, k := range keys {
		uniq := util.Unique(groups[k])
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
		if err := jsonPrinter.OutputJSON(threshold, sortBy); err != nil {
			return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s): %w", threshold, sortBy.String(), err)
		}
	}

	if err := p.PrintFooter(); err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}
	return nil
}

func runCobraCommand(cmd *cobra.Command, args []string) error { //nolint:cyclop,funlen // Cobra command with multiple flag handling paths
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verbose, _ := cmd.Flags().GetBool("verbose")
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")

	// Validate sorting criteria
	if _, err := printer.ParseSortBy(sortBy); err != nil {
		return fmt.Errorf("invalid --sort value %q: %w", sortBy, err)
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	_, _ = cmd.Flags().GetString("output-dir")
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
		return errors.New("all mode not yet implemented")
	}

	// Add timeout context if specified
	ctx := context.Background()
	if mergedConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mergedConfig.Timeout)*time.Second)
		defer cancel()
		fmt.Fprintf(os.Stderr, "⏱️  Execution timeout: %ds\n", mergedConfig.Timeout)
		_ = ctx // TODO: Pass context to executeAnalysis for proper timeout handling
	}
	// For now, timeout config is stored but not fully implemented in analysis
	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		return fmt.Errorf("analysis failed for paths %v: %w", mergedConfig.Paths, err)
	}

	p := createPrinter(mergedConfig.OutputFormat)(os.Stdout, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, printer.SortBy(sortBy), mergedConfig.Threshold); err != nil {
		return fmt.Errorf("failed to print duplicates (sortBy: %s, threshold: %d): %w", sortBy, mergedConfig.Threshold, err)
	}

	return nil
}
