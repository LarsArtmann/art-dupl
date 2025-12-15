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
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
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
	sortBy        = flag.String("sort", "size", "sort clone groups by: size, occurrence, hash")
	paths         []string
)

const (
	defaultThreshold = 15
	vendorDirPrefix  = "vendor" + string(filepath.Separator)
	vendorDirInPath  = string(filepath.Separator) + vendorDirPrefix
)

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

	// Handle verbose flag (either -v or -verbose)
	verboseFlag := *verbose || *verboseLong
	if verboseFlag {
		cliConfig.Verbose = verboseFlag
	}

	// Handle threshold flag (either -t or -threshold)
	thresholdFlag := *threshold
	if thresholdFlag != defaultThreshold || *thresholdLong != defaultThreshold {
		if *thresholdLong != defaultThreshold {
			thresholdFlag = *thresholdLong
		}
		cliConfig.Threshold = thresholdFlag
	}

	if *vendor {
		cliConfig.IncludeVendor = *vendor
	}
	if *files {
		cliConfig.FilesFromStdin = *files
	}

	if *html {
		cliConfig.OutputFormat = config.OutputFormatHTML
	} else if *plumbing {
		cliConfig.OutputFormat = config.OutputFormatPlumbing
	} else if *jsonFlag {
		cliConfig.OutputFormat = config.OutputFormatJSON
	}

	if flag.NArg() > 0 {
		cliConfig.Paths = flag.Args()
	}

	// Merge file and CLI configurations
	mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

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
	paths = mergedConfig.Paths
	vendor = &mergedConfig.IncludeVendor
	verbose = &verboseFlag
	threshold = &thresholdFlag
	files = &mergedConfig.FilesFromStdin

	if *verbose {
		log.Println("Building suffix tree")
	}
	schan, filesCountChan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

	// Get file count
	filesCount := <-filesCountChan

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	if *verbose {
		log.Println("Searching for clones")
	}
	mchan := t.FindDuplOver(*threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			match := syntax.FindSyntaxUnits(*data, m, *threshold)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	// Select printer based on output format
	var newPrinter func(io.Writer, printer.ReadFile) printer.Printer
	switch mergedConfig.OutputFormat {
	case config.OutputFormatHTML:
		newPrinter = printer.NewHTML
	case config.OutputFormatPlumbing:
		newPrinter = printer.NewPlumbing
	case config.OutputFormatJSON:
		newPrinter = printer.NewJSON
	default:
		newPrinter = printer.NewText
	}

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

	if err := printDupls(p, duplChan, *sortBy); err != nil {
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
    	sort clone groups by: size, occurrence, hash (default "size")
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

// filesFeed creates a channel of file paths to process
func filesFeed() chan string {
	if *files {
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
func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy string) error {
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
		return jsonPrinter.OutputJSON(*threshold, sortBy)
	}

	return p.PrintFooter()
}

// runCobraCommand implements the command execution with Cobra flags
func runCobraCommand(cmd *cobra.Command, args []string) error {
	// Get flag values from Cobra command
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verboseShort, _ := cmd.Flags().GetBool("verbose") // -v flag
	verboseLong, _ := cmd.Flags().GetBool("verbose") // --verbose flag (hidden)
	thresholdShort, _ := cmd.Flags().GetInt("threshold") // -t flag
	thresholdLong, _ := cmd.Flags().GetInt("threshold") // --threshold flag (hidden)
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	json, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")

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
	}

	// Create CLI config from command line arguments
	cliConfig := &config.Config{}

	// Handle verbose flag (either -v or -verbose)
	verboseFlag := verboseShort || verboseLong
	if verboseFlag {
		cliConfig.Verbose = verboseFlag
	}

	// Handle threshold flag (either -t or -threshold)
	thresholdFlag := thresholdShort
	if thresholdFlag != defaultThreshold || thresholdLong != defaultThreshold {
		if thresholdLong != defaultThreshold {
			thresholdFlag = thresholdLong
		}
		cliConfig.Threshold = thresholdFlag
	}

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

	// Merge file and CLI configurations
	mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

	// Validate merged configuration
	if err = config.ValidateConfig(mergedConfig); err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "configuration error: %v\n", err); err != nil {
			return err
		}
		return err
	}

	// Validate output format conflicts
	if mergedConfig.OutputFormat == "html" && plumbing {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or HTML output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}
	if mergedConfig.OutputFormat == "html" && json {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either HTML or JSON output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}
	if mergedConfig.OutputFormat == "plumbing" && json {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or JSON output\n"); err != nil {
			return err
		}
		return fmt.Errorf("conflicting output formats")
	}

	// Update global variables with merged config (for compatibility with existing code)
	paths = mergedConfig.Paths
	// Note: vendor, verbose, threshold, files global variables are not used in this function
	// as we now use the local variables instead

	if verboseFlag {
		log.Println("Building suffix tree")
	}
	schan, filesCountChan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

	// Get file count
	filesCount := <-filesCountChan

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	if verboseFlag {
		log.Println("Searching for clones")
	}
	mchan := t.FindDuplOver(thresholdFlag)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			match := syntax.FindSyntaxUnits(*data, m, thresholdFlag)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	// Select printer based on output format
	var newPrinter func(io.Writer, printer.ReadFile) printer.Printer
	switch mergedConfig.OutputFormat {
	case config.OutputFormatHTML:
		newPrinter = printer.NewHTML
	case config.OutputFormatPlumbing:
		newPrinter = printer.NewPlumbing
	case config.OutputFormatJSON:
		newPrinter = printer.NewJSON
	default:
		newPrinter = printer.NewText
	}

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

	if err := printDupls(p, duplChan, sortBy); err != nil {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: %v\n", err); err != nil {
			return err
		}
		return err
	}
	return nil
}
