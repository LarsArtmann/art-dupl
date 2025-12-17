package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
	"github.com/spf13/cobra"
)

// CLIInterface defines the interface for CLI operations
type CLIInterface interface {
	Exit(code int)
	Stderr() io.Writer
	Stdout() io.Writer
}

// RealCLI is the production implementation of CLIInterface
type RealCLI struct{}

func (r *RealCLI) Exit(code int)     { os.Exit(code) }
func (r *RealCLI) Stderr() io.Writer { return os.Stderr }
func (r *RealCLI) Stdout() io.Writer { return os.Stdout }

// TestCLI is the test implementation of CLIInterface
type TestCLI struct {
	ExitCode  int
	StderrBuf strings.Builder
	StdoutBuf strings.Builder
}

func (t *TestCLI) Exit(code int)     { t.ExitCode = code }
func (t *TestCLI) Stderr() io.Writer { return &t.StderrBuf }
func (t *TestCLI) Stdout() io.Writer { return &t.StdoutBuf }

var cli CLIInterface = &RealCLI{}

var (
	configFile    = flag.String("config", "", "path to configuration file (JSON format)")
	vendor        = flag.Bool("vendor", false, "include vendor directory in analysis")
	verbose       = flag.Bool("v", false, "enable verbose logging to show processing progress")
	verboseLong   = flag.Bool("verbose", false, "enable verbose logging to show processing progress")
	threshold     = flag.Int("t", 15, "minimum token sequence size to consider as clone")
	thresholdLong = flag.Int("threshold", 15, "minimum token sequence size to consider as clone")
	files         = flag.Bool("files", false, "read file names from stdin, one per line")
	html          = flag.Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	jsonFlag      = flag.Bool("json", false, "output structured JSON format with metadata and statistics")
	plumbing      = flag.Bool("plumbing", false, "output machine-readable plumbing format for script integration")
	sortBy        = flag.String("sort", "size", "sort clone groups by: size, occurrence, hash, total-tokens")
)

const (
	defaultThreshold = 15
	vendorDirPrefix  = "vendor" + string(filepath.Separator)
	vendorDirInPath  = string(filepath.Separator) + vendorDirPrefix
)

// createPrinter returns the appropriate printer function based on the output format
func createPrinter(outputFormat config.OutputFormat) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return printer.NewHTML
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	default:
		return printer.NewText
	}
}

// handleConfigError handles configuration validation errors consistently
func handleConfigError(err error) error {
	if err != nil {
		if _, writeErr := fmt.Fprintf(cli.Stderr(), "configuration error: %v\n", err); writeErr != nil {
			return writeErr
		}
		return err
	}
	return nil
}

// validateOutputFormatConflicts checks for conflicting output format combinations
func validateOutputFormatConflicts(outputFormat config.OutputFormat, htmlFlag, jsonFlag, plumbingFlag bool) error {
	if outputFormat == config.OutputFormatHTML && plumbingFlag {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or HTML output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}
	if outputFormat == config.OutputFormatHTML && jsonFlag {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either HTML or JSON output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}
	if outputFormat == config.OutputFormatPlumbing && jsonFlag {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or JSON output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}
	return nil
}

// Run is the main application logic
func Run() int {
	flag.Usage = usage
	flag.Parse()

	// Load configuration from file if specified
	var fileConfig *config.Config
	var err error
	if *configFile != "" {
		fileConfig, err = config.LoadConfig(*configFile)
		if err != nil {
			if _, err := fmt.Fprintf(cli.Stderr(), "error loading config: %v\n", err); err != nil {
				// If we can't even write to stderr, just exit
				cli.Exit(1)
				return 1
			}
			cli.Exit(1)
			return 1
		}
	}

	// Create CLI config from command line arguments
	cliConfig := &config.Config{}

	// Note: We don't set threshold/outputFormat from CLI here initially
	// to allow config file values to take precedence
	// CLI values will override only if explicitly provided

	// Merge file and CLI configurations
	mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

	// Now handle CLI overrides for explicitly set flags
	verboseFlag := *verbose || *verboseLong
	if verboseFlag {
		mergedConfig.Verbose = verboseFlag
	}

	// Handle threshold flag (either -t or -threshold)
	thresholdFlag := *threshold
	if *thresholdLong != defaultThreshold {
		thresholdFlag = *thresholdLong
	}
	// Only override threshold if CLI flags were explicitly used
	if thresholdFlag != defaultThreshold || *thresholdLong != defaultThreshold {
		mergedConfig.Threshold = thresholdFlag
	}

	if *vendor {
		mergedConfig.IncludeVendor = *vendor
	}
	if *files {
		mergedConfig.FilesFromStdin = *files
	}

	if *html {
		mergedConfig.OutputFormat = config.OutputFormatHTML
	} else if *plumbing {
		mergedConfig.OutputFormat = config.OutputFormatPlumbing
	} else if *jsonFlag {
		mergedConfig.OutputFormat = config.OutputFormatJSON
	}

	if flag.NArg() > 0 {
		mergedConfig.Paths = flag.Args()
	}

	// Validate merged configuration
	if err = config.ValidateConfig(mergedConfig); err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "configuration error: %v\n", err); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}

	// Validate output format conflicts
	if mergedConfig.OutputFormat == "html" && *plumbing {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or HTML output\n"); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}
	if mergedConfig.OutputFormat == "html" && *jsonFlag {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either HTML or JSON output\n"); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}
	if mergedConfig.OutputFormat == "plumbing" && *jsonFlag {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or JSON output\n"); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}

	// Update global variables with merged config (for compatibility with existing code)
	vendor = &mergedConfig.IncludeVendor
	threshold = &mergedConfig.Threshold // Also update global threshold
	// For verbose and files, use values from mergedConfig
	// (these will be used by code that expects global variables)
	files = &mergedConfig.FilesFromStdin

	// Execute analysis using common helper
	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "analysis error: %v\n", err); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}

	// Select printer based on output format
	newPrinter := createPrinter(mergedConfig.OutputFormat)

	// Handle output file if specified
	outputWriter := cli.Stdout()
	if mergedConfig.OutputFile != "" {
		file, err := os.Create(mergedConfig.OutputFile)
		if err != nil {
			if _, err := fmt.Fprintf(cli.Stderr(), "error creating output file: %v\n", err); err != nil {
				cli.Exit(1)
				return 1
			}
			cli.Exit(1)
			return 1
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(cli.Stderr(), "warning: failed to close file: %v\n", closeErr)
			}
		}()
		outputWriter = file
	}

	p := newPrinter(outputWriter, os.ReadFile)

	// Set filesCount for JSONPrinter
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, *sortBy, *threshold); err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: %v\n", err); err != nil {
			// If we can't even write to stderr, just exit
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}
	return 0
}

// usage prints the usage information
func usage() {
	fmt.Fprintln(os.Stderr, `Usage: art-dupl [flags] [paths]

Paths:
  If given path is a file, art-dupl will use it regardless of
  file extension. If it is a directory, it will recursively
  search for *.go files in that directory.

  If no path is given, art-dupl will recursively search for *.go
  files in the current directory.

Flags:
  -config string
    	path to configuration file (JSON format)
  -files
    	read file names from stdin, one per line
  -html
    	output results as HTML with syntax-highlighted code fragments
  -json
    	output structured JSON format with metadata and statistics
  -plumbing
    	output machine-readable plumbing format for script integration
  -sort string
    	sort clone groups by: size, occurrence, hash, total-tokens (default "size")
  -t, -threshold int
    	minimum token sequence size to consider as clone (default 15)
  -vendor
    	include vendor directory in analysis
  -v, -verbose
    	enable verbose logging to show processing progress

Output Formats:
  text     - Human-readable clone listing (default)
  html      - HTML report with syntax-highlighted code
  json      - Structured JSON data for automation/CI
  plumbing  - Machine-readable format for scripts

Configuration File:
  Create a JSON file with settings:
  {
    "threshold": 30,
    "outputFormat": "json",
    "paths": ["./src", "./lib"],
    "includeVendor": false,
    "verbose": true
  }
  Use with: art-dupl -config config.json

Examples:
  # Basic analysis with default threshold
  art-dupl

  # Higher threshold for larger clones only
  art-dupl -t 100

  # Sort by different criteria
  art-dupl -sort size .          # Largest clones first (default)
  art-dupl -sort occurrence .     # Most widespread clones first
  art-dupl -sort hash .          # Alphabetical order
  art-dupl -sort total-tokens .  # Total tokens across all files

  # JSON output with sorting
  art-dupl -json -sort size . | jq '.clone_groups[0]'

  # HTML report file with occurrence sorting
  art-dupl -html -sort occurrence . > report.html

  # Plumbing output for CI/CD with hash sorting
  art-dupl -plumbing -sort hash . > duplicates.txt

  # Use configuration file
  art-dupl -config dupl.json ./src

  # Analyze test files only
  find . -name '*_test.go' | art-dupl -files

  # CI/CD: Fail if too many duplicates
  TOTAL_CLONES=$(art-dupl -json . | jq '.summary.total_clones')
  if [ "$TOTAL_CLONES" -gt 100 ]; then
    echo "Too many code duplicates: $TOTAL_CLONES"
    exit 1
  fi`)
	os.Exit(2)
}

// createHashDuplChannel creates a channel for hash-based duplicate detection
// createHashDuplChannel creates a channel for hash-based duplicate detection
func createHashDuplChannel(cfg *config.Config, data []*syntax.Node, t *suffixtree.STree, verbose bool) chan syntax.Match {
	multiDetector := detection.NewMultiDetector(cfg, data, t, verbose)
	duplChan := make(chan syntax.Match)
	// Find duplicates
	go func() {
		defer close(duplChan)
		matches := multiDetector.FindDuplOver(cfg.Threshold)
		// Detection completed
		for match := range matches {
			duplChan <- match
		}
	}()
	return duplChan
}

// createArtDuplChannel creates a channel for art-dupl (suffix tree) based duplicate detection
// createArtDuplChannel creates a channel for art-dupl (suffix tree) based duplicate detection
func createArtDuplChannel(cfg *config.Config, data []*syntax.Node, t *suffixtree.STree) chan syntax.Match {
	mchan := t.FindDuplOver(cfg.Threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		defer close(duplChan)
		for m := range mchan {
			match := syntax.FindSyntaxUnits(data, m, cfg.Threshold)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
	}()
	return duplChan
}

// createDuplChannel creates a channel for duplicate detection based on the method
func createDuplChannel(cfg *config.Config, data []*syntax.Node, t *suffixtree.STree, verbose bool) chan syntax.Match {
	if cfg.DetectionMethods.Contains(config.DetectionMethodHash) {
		return createHashDuplChannel(cfg, data, t, verbose)
	}
	return createArtDuplChannel(cfg, data, t)
}

// createDuplChannelForMethod creates a channel for duplicate detection based on a single method
func createDuplChannelForMethod(method config.DetectionMethod, cfg *config.Config, data []*syntax.Node, t *suffixtree.STree, verbose bool) chan syntax.Match {
	// Create a temporary config with the single detection method
	tempConfig := *cfg
	tempConfig.DetectionMethods = config.DetectionMethods{method}
	return createDuplChannel(&tempConfig, data, t, verbose)
}

// buildSuffixTree builds a suffix tree from provided paths and returns tree, data, and file count
func buildSuffixTree(paths []string, verbose, filesFromStdin bool) (*suffixtree.STree, []*syntax.Node, int, error) {
	if verbose {
		log.Println("Building suffix tree")
	} else {
		// Show basic progress even without verbose
		if _, err := fmt.Fprintf(cli.Stderr(), "    📖 Parsing files and building analysis tree..."); err != nil {
			_ = err
		}
	}

	schan, filesCountChan := job.Parse(filesFeedWithOptions(paths, filesFromStdin))
	t, data, done := job.BuildTree(schan)
	<-done

	// Get file count
	filesCount := <-filesCountChan

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	if verbose {
		log.Println("Searching for clones")
	} else {
		// Show completion
		if _, err := fmt.Fprintf(cli.Stderr(), " ✅\n"); err != nil {
			_ = err
		}
	}

	return t, *data, filesCount, nil
}

// executeAnalysis runs the core duplicate analysis logic
func executeAnalysis(mergedConfig *config.Config, paths []string) (chan syntax.Match, int, error) {
	t, data, filesCount, err := buildSuffixTree(paths, mergedConfig.Verbose, mergedConfig.FilesFromStdin)
	if err != nil {
		return nil, 0, err
	}

	// Use multi-detector if hash detection is enabled
	duplChan := createDuplChannel(mergedConfig, data, t, mergedConfig.Verbose)

	return duplChan, filesCount, nil
}

// filesFeedWithOptions creates a channel of file paths with options
func filesFeedWithOptions(paths []string, fromStdin bool) chan string {
	if fromStdin {
		fchan := make(chan string)
		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				fchan <- strings.TrimPrefix(f, "./")
			}
			close(fchan)
		}()
		return fchan
	}
	return crawlPaths(paths)
}

// crawlPaths walks paths and returns a channel of Go files
func crawlPaths(paths []string) chan string {
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				if _, err := fmt.Fprintf(cli.Stderr(), "error: cannot stat %s: %v\n", path, err); err != nil {
					cli.Exit(1)
					return
				}
				cli.Exit(1)
				return
			}
			if !info.IsDir() {
				fchan <- path
				continue
			}
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if !*vendor && (strings.HasPrefix(path, vendorDirPrefix) ||
					strings.Contains(path, vendorDirInPath)) {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					fchan <- path
				}
				return nil
			})
			if err != nil {
				if _, err := fmt.Fprintf(cli.Stderr(), "error: cannot walk %s: %v\n", path, err); err != nil {
					cli.Exit(1)
					return
				}
				cli.Exit(1)
				return
			}
		}
		close(fchan)
	}()
	return fchan
}

// printDupls prints duplicates using the specified printer
func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy string, threshold int) error {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := p.PrintHeader(); err != nil {
		return err
	}
	for _, k := range keys {
		uniq := util.Unique(groups[k])
		if len(uniq) > 1 {
			// Set hash for JSONPrinter if applicable
			if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
				jsonPrinter.SetHash(k)
			}
			if err := p.PrintClones(uniq, sortBy); err != nil {
				return err
			}
		}
	}

	// Handle JSON output special case
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		return jsonPrinter.OutputJSON(threshold, sortBy)
	}

	return p.PrintFooter()
}

// runCobraCommand implements the command execution with Cobra flags
func runCobraCommand(cmd *cobra.Command, args []string) error {
	// Get flag values from Cobra command
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verbose, _ := cmd.Flags().GetBool("verbose")    // -v flag
	threshold, _ := cmd.Flags().GetInt("threshold") // -t flag
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	json, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")
	detectionMethods, _ := cmd.Flags().GetString("detection-methods")
	allFlag, _ := cmd.Flags().GetBool("all")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	// Load configuration from file if specified
	var fileConfig *config.Config
	var err error
	if configFile != "" {
		fileConfig, err = config.LoadConfig(configFile)
		if err != nil {
			if _, err := fmt.Fprintf(cli.Stderr(), "error loading config: %v\n", err); err != nil {
				return err
			}
			return err
		}
		// Debug output removed for production builds
	}

	// Create CLI config from command line arguments
	cliConfig := &config.Config{}

	// Check if threshold flag was explicitly changed from default
	thresholdValue := threshold

	thresholdExplicitlySet := thresholdValue != defaultThreshold

	// Handle threshold flag (only if explicitly set)
	if thresholdExplicitlySet {
		cliConfig.Threshold = thresholdValue
	}
	// Don't set CLI config.Threshold at all if not explicitly set
	// This allows file config threshold to take precedence

	if vendor {
		cliConfig.IncludeVendor = vendor
	}
	if files {
		cliConfig.FilesFromStdin = files
	}

	if html {
		cliConfig.OutputFormat = config.OutputFormatHTML
	} else if plumbing {
		cliConfig.OutputFormat = config.OutputFormatPlumbing
	} else if json {
		cliConfig.OutputFormat = config.OutputFormatJSON
	}

	if len(args) > 0 {
		cliConfig.Paths = args
	}

	// Handle detection methods
	if detectionMethods != "" {
		methods, err := config.ParseDetectionMethods(detectionMethods)
		if err != nil {
			if _, err := fmt.Fprintf(cli.Stderr(), "detection methods error: %v\n", err); err != nil {
				return err
			}
			return err
		}
		cliConfig.DetectionMethods = methods
	}

	// Handle "all" flag - this overrides other output formats
	if allFlag {
		// "all" mode enables all detection methods and formats
		cliConfig.DetectionMethods = config.DetectionMethods{config.DetectionMethodArtDupl, config.DetectionMethodHash}

		// We need to merge configs first to get the complete configuration
		mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

		// Validate merged configuration
		if err = handleConfigError(config.ValidateConfig(mergedConfig)); err != nil {
			return err
		}

		// Run the all-mode handler
		return runAllMode(outputDir, mergedConfig.Threshold, vendor, verbose, args)
	}

	// Merge file and CLI configurations
	mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

	// Validate merged configuration
	if err = handleConfigError(config.ValidateConfig(mergedConfig)); err != nil {
		return err
	}

	// Validate output format conflicts
	if err = validateOutputFormatConflicts(mergedConfig.OutputFormat, html, json, plumbing); err != nil {
		return err
	}

	// Update global variables with merged config (for compatibility with existing code)
	// Note: vendor, verbose, threshold, files global variables are not used in this function
	// as we now use the local variables instead

	// Execute analysis using common helper
	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "analysis error: %v\n", err); err != nil {
			return err
		}
		return err
	}

	// Select printer based on output format
	newPrinter := createPrinter(mergedConfig.OutputFormat)

	// Handle output file if specified
	outputWriter := cli.Stdout()
	if mergedConfig.OutputFile != "" {
		file, err := os.Create(mergedConfig.OutputFile)
		if err != nil {
			if _, err := fmt.Fprintf(cli.Stderr(), "error creating output file: %v\n", err); err != nil {
				return err
			}
			return err
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(cli.Stderr(), "warning: failed to close file: %v\n", closeErr)
			}
		}()
		outputWriter = file
	}

	p := newPrinter(outputWriter, os.ReadFile)

	// Set filesCount for JSONPrinter
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, sortBy, mergedConfig.Threshold); err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: %v\n", err); err != nil {
			return err
		}
		return err
	}
	return nil
}

// runAllMode generates all output formats for all detection methods
func runAllMode(outputDir string, threshold int, vendor, verbose bool, paths []string) error {
	// Record start time for overall timing
	overallStart := time.Now()

	// Start with clear feedback about what's happening
	if _, err := fmt.Fprintf(cli.Stderr(), "🚀 Starting comprehensive code duplication analysis...\n\n"); err != nil {
		// Continue even if we can't write to stderr
		_ = err // Explicitly ignore error
	}

	// Show what's being analyzed
	if len(paths) == 0 {
		if _, err := fmt.Fprintf(cli.Stderr(), "📂 Analyzing current directory\n"); err != nil {
			_ = err
		}
	} else {
		if _, err := fmt.Fprintf(cli.Stderr(), "📂 Analyzing paths: %v\n", paths); err != nil {
			_ = err
		}
	}

	if _, err := fmt.Fprintf(cli.Stderr(), "📋 Output directory: %s\n", outputDir); err != nil {
		_ = err
	}
	if _, err := fmt.Fprintf(cli.Stderr(), "⚡ Threshold: %d tokens\n", threshold); err != nil {
		_ = err
	}
	if _, err := fmt.Fprintf(cli.Stderr(), "🔧 Include vendor: %t\n\n", vendor); err != nil {
		_ = err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("❌ Failed to create output directory '%s': %v\n💡 Try: Check if you have write permissions or specify a different output directory with --output-dir", outputDir, err)
	}

	// Define detection methods and output formats
	detectionMethods := []config.DetectionMethod{config.DetectionMethodArtDupl, config.DetectionMethodHash}
	outputFormats := []struct {
		name       string
		format     config.OutputFormat
		ext        string
		newPrinter func(io.Writer, printer.ReadFile) printer.Printer
	}{
		{"text", config.OutputFormatText, ".txt", printer.NewText},
		{"html", config.OutputFormatHTML, ".html", printer.NewHTML},
		{"json", config.OutputFormatJSON, ".json", printer.NewJSON},
		{"plumbing", config.OutputFormatPlumbing, ".plumbing", printer.NewPlumbing},
	}

	if _, err := fmt.Fprintf(cli.Stderr(), "🔍 Running analysis with %d detection methods and generating %d output formats each...\n\n", len(detectionMethods), len(outputFormats)); err != nil {
		_ = err
	}

	// Track overall statistics
	totalFilesAnalyzed := 0
	totalClonesFound := 0

	// Run analysis for each detection method
	for i, method := range detectionMethods {
		methodStart := time.Now()
		if _, err := fmt.Fprintf(cli.Stderr(), "[%d/%d] 🔍 Analyzing with %s detection method...\n", i+1, len(detectionMethods), method); err != nil {
			_ = err
		}

		if verbose {
			if _, err := fmt.Fprintf(cli.Stderr(), "    Debug: Running %s detection method...\n", method); err != nil {
				_ = err
			}
		}

		// Configure the analysis with this detection method
		config := &config.Config{
			Threshold:        threshold,
			IncludeVendor:    vendor,
			Verbose:          verbose,
			Paths:            paths,
			DetectionMethods: config.DetectionMethods{method},
		}

		// Run the analysis once per method and generate all formats
		stats, err := runAnalysisForAllFormats(config, outputDir, outputFormats, method, verbose)
		if err != nil {
			return fmt.Errorf("error running %s analysis: %v", method, err)
		}

		// Track overall statistics (use max files analyzed since each method analyzes same files)
		if stats.FilesCount > totalFilesAnalyzed {
			totalFilesAnalyzed = stats.FilesCount
		}
		totalClonesFound += stats.ClonesCount

		if _, err := fmt.Fprintf(cli.Stderr(), "✅ %s analysis completed (%.2fs) - %d files analyzed, %d clone groups found\n\n", method, time.Since(methodStart).Seconds(), stats.FilesCount, stats.ClonesCount); err != nil {
			_ = err
		}
	}

	// Final summary
	if _, err := fmt.Fprintf(cli.Stderr(), "🎉 All analysis complete! (Total time: %.2fs)\n", time.Since(overallStart).Seconds()); err != nil {
		_ = err
	}
	if _, err := fmt.Fprintf(cli.Stderr(), "📊 Overall statistics: %d files analyzed, %d total clone groups found\n", totalFilesAnalyzed, totalClonesFound); err != nil {
		_ = err
	}
	if _, err := fmt.Fprintf(cli.Stderr(), "📁 All reports generated in: %s\n", outputDir); err != nil {
		_ = err
	}

	// Show a summary of clones found for each method to stdout
	// Count clones from the text files we already generated
	for _, method := range detectionMethods {
		// Read the text file to count clones
		filename := filepath.Join(outputDir, fmt.Sprintf("%s.txt", method))
		if data, err := os.ReadFile(filename); err == nil {
			// Count occurrences of "found X clones:" in the file
			content := string(data)
			lines := strings.Split(content, "\n")
			totalClones := 0
			for _, line := range lines {
				if strings.HasPrefix(line, "found ") && strings.Contains(line, " clones:") {
					// Extract number from "found X clones:"
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if count, err := strconv.Atoi(parts[1]); err == nil {
							totalClones += count
						}
					}
				}
			}
			if _, err := fmt.Fprintf(cli.Stdout(), "found %d clones (%s method)\n", totalClones, method); err != nil {
				_ = err
			}
		}
	}

	// List all generated files
	if _, err := fmt.Fprintf(cli.Stderr(), "📋 Generated files:\n"); err != nil {
		_ = err
	}
	for _, method := range detectionMethods {
		for _, fmtInfo := range outputFormats {
			filename := filepath.Join(outputDir, fmt.Sprintf("%s%s", method, fmtInfo.ext))
			if _, err := fmt.Fprintf(cli.Stderr(), "   📄 %s\n", filename); err != nil {
				_ = err
			}
		}
	}

	if _, err := fmt.Fprintf(cli.Stderr(), "\n✨ Done! Use the reports above to review code duplication findings.\n"); err != nil {
		_ = err
	}

	if verbose {
		if _, err := fmt.Fprintf(cli.Stderr(), "\nDebug: All reports generated in: %s\n", outputDir); err != nil {
			_ = err
		}
	}
	return nil
}

// AnalysisStats holds statistics from a single analysis run
type AnalysisStats struct {
	FilesCount  int
	ClonesCount int
}

// runAnalysisForAllFormats runs analysis once and generates all output formats
func runAnalysisForAllFormats(cfg *config.Config, outputDir string, formats []struct {
	name       string
	format     config.OutputFormat
	ext        string
	newPrinter func(io.Writer, printer.ReadFile) printer.Printer
}, method config.DetectionMethod, verbose bool,
) (*AnalysisStats, error) {
	t, data, filesCount, err := buildSuffixTree(cfg.Paths, verbose, cfg.FilesFromStdin)
	if err != nil {
		return nil, fmt.Errorf("failed to build suffix tree: %v", err)
	}

	// Count clones by consuming the channel once
	duplChanForCounting := createDuplChannelForMethod(method, cfg, data, t, false)
	clonesCount := 0
	for range duplChanForCounting {
		clonesCount++
	}

	// Generate all output formats
	// Generate output for each format
	for _, fmtInfo := range formats {
		// Generate output for this format
		filename := filepath.Join(outputDir, fmt.Sprintf("%s%s", method, fmtInfo.ext))

		if !verbose {
			// Show format generation progress
			if _, err := fmt.Fprintf(cli.Stderr(), "    📝 Generating %s format...", fmtInfo.name); err != nil {
				_ = err
			}
		}

		// Create output file
		file, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s file: %v", filename, err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				// Log error but don't fail the operation
				if verbose {
					if _, err := fmt.Fprintf(cli.Stderr(), "Error closing file: %v\n", err); err != nil {
						// Can't write error to stderr
						_ = err // Explicitly ignore error
					}
				}
			}
		}()
		// File created successfully

		p := fmtInfo.newPrinter(file, os.ReadFile)

		// Set filesCount for JSONPrinter
		if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
			jsonPrinter.SetFilesCount(filesCount)
		}

		// We need to recreate the channel for each format since it gets consumed
		duplChanCopy := createDuplChannelForMethod(method, cfg, data, t, false) // don't log again

		if err := printDupls(p, duplChanCopy, "size", cfg.Threshold); err != nil {
			return nil, fmt.Errorf("error writing %s format: %v", fmtInfo.name, err)
		}

		if verbose {
			if _, err := fmt.Fprintf(cli.Stderr(), "Generated %s report: %s\n", fmtInfo.name, filename); err != nil {
				// Continue even if we can't write verbose output
				_ = err // Explicitly ignore error
			}
		} else {
			// Show completion for non-verbose mode
			if _, err := fmt.Fprintf(cli.Stderr(), " ✅\n"); err != nil {
				_ = err
			}
		}
	}

	return &AnalysisStats{
		FilesCount:  filesCount,
		ClonesCount: clonesCount,
	}, nil
}
