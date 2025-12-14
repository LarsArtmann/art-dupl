package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

const (
	vendorDirPrefix = "vendor" + string(filepath.Separator)
	vendorDirInPath = string(filepath.Separator) + vendorDirPrefix
	defaultThreshold = 15
)

// CLIInterface defines interface for CLI operations
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

// Global variables for compatibility with existing code
var (
	paths    []string
	vendor   *bool
	verbose  *bool
	threshold *int
	files    *bool
	sortBy   *string
)

func main() {
	// Create root command
	rootCmd := createRootCommand()

	// Execute with fang for enhanced CLI experience
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

// createRootCommand creates the main cobra command with all flags and configuration
func createRootCommand() *cobra.Command {
	var (
		configFile    string
		vendorFlag    bool
		verboseFlag   bool
		verboseLong   bool
		thresholdFlag  int
		thresholdLong int
		filesFlag     bool
		html          bool
		json          bool
		plumbing      bool
		sortByFlag    string
	)

	cmd := &cobra.Command{
		Use:   "dupl [flags] [paths]",
		Short: "Find code clones in Go source files",
		Long: `Dupl is a code duplication detection tool for Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones while
ignoring literal values. The tool uses suffix tree algorithms to efficiently
identify duplicate code patterns.

If no path is given, dupl will recursively search for *.go files in current directory.`,
		Example: `  # Basic analysis with default threshold
  dupl

  # Higher threshold for larger clones only
  dupl -t 100

  # JSON output for CI/CD integration
  dupl -json -t 20

  # HTML report file
  dupl -html > report.html

  # Use configuration file
  dupl -config dupl.json ./src

  # Analyze test files only
  find . -name '*_test.go' | dupl -files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set global variables for compatibility with existing functions
			paths = args
			vendor = &vendorFlag
			verbose = &verboseFlag
			threshold = &thresholdFlag
			files = &filesFlag
			sortBy = &sortByFlag

			// Create CLI config
			cliConfig := &config.Config{}

			// Handle verbose flag (either -v or -verbose)
			verboseFlag := verboseFlag || verboseLong
			if verboseFlag {
				cliConfig.Verbose = verboseFlag
			}

			// Handle threshold flag (either -t or -threshold)
			if thresholdFlag != defaultThreshold || thresholdLong != defaultThreshold {
				if thresholdLong != defaultThreshold {
					thresholdFlag = thresholdLong
				}
				cliConfig.Threshold = thresholdFlag
			}

			if vendorFlag {
				cliConfig.IncludeVendor = vendorFlag
			}
			if filesFlag {
				cliConfig.FilesFromStdin = filesFlag
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

			// Load configuration from file if specified
			var fileConfig *config.Config
			var err error
			if configFile != "" {
				fileConfig, err = config.LoadConfig(configFile)
				if err != nil {
					return fmt.Errorf("error loading config: %w", err)
				}
			}

			// Merge file and CLI configurations
			mergedConfig := config.MergeConfigs(fileConfig, cliConfig)

			// Validate merged configuration
			if err = config.ValidateConfig(mergedConfig); err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}

			// Validate output format conflicts
			if mergedConfig.OutputFormat == "html" && plumbing {
				return fmt.Errorf("error: you can have either plumbing or HTML output")
			}
			if mergedConfig.OutputFormat == "html" && json {
				return fmt.Errorf("error: you can have either HTML or JSON output")
			}
			if mergedConfig.OutputFormat == "plumbing" && json {
				return fmt.Errorf("error: you can have either plumbing or JSON output")
			}

			// Execute the main analysis logic
			return runDuplAnalysis(mergedConfig, thresholdFlag, verboseFlag)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&configFile, "config", "", "path to configuration file (JSON format)")
	cmd.Flags().BoolVar(&vendorFlag, "vendor", false, "include vendor directory in analysis")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose logging to show processing progress")
	cmd.Flags().IntVarP(&thresholdFlag, "threshold", "t", defaultThreshold, "minimum token sequence size to consider as clone")
	cmd.Flags().BoolVar(&filesFlag, "files", false, "read file names from stdin, one per line")
	cmd.Flags().BoolVar(&html, "html", false, "output results as HTML with syntax-highlighted code fragments")
	cmd.Flags().BoolVar(&json, "json", false, "output structured JSON format with metadata and statistics")
	cmd.Flags().BoolVar(&plumbing, "plumbing", false, "output machine-readable plumbing format for script integration")
	cmd.Flags().StringVar(&sortByFlag, "sort", "size", "sort clone groups by: size, occurrence, hash")

	return cmd
}

// runDuplAnalysis contains the main analysis logic
func runDuplAnalysis(mergedConfig *config.Config, thresholdFlag int, verboseFlag bool) error {
	// Create CLI interface for compatibility
	cli := &RealCLI{}

	if verboseFlag {
		fmt.Fprintf(cli.Stderr(), "Building suffix tree\n")
	}
	schan, filesCountChan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

	// Get file count
	filesCount := <-filesCountChan

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	if verboseFlag {
		fmt.Fprintf(cli.Stderr(), "Searching for clones\n")
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
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(cli.Stderr(), "warning: failed to close file: %v\n", closeErr)
			}
		}()
		outputWriter = file
	}

	p := newPrinter(outputWriter, os.ReadFile)

	// Set filesCount for JSONPrinter if applicable
	if jsonSetter, ok := p.(interface{ SetFilesCount(int) }); ok {
		jsonSetter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

// filesFeed creates a channel for file paths
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

// crawlPaths recursively walks paths and returns Go files
func crawlPaths(paths []string) chan string {
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot stat %s: %v\n", path, err)
				continue
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
				fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
				continue
			}
		}
		close(fchan)
	}()
	return fchan
}

// printDupls prints the duplicate code matches
func printDupls(p printer.Printer, duplChan <-chan syntax.Match) error {
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
			if jsonSetter, ok := p.(interface{ SetHash(string) }); ok {
				jsonSetter.SetHash(k)
			}
			if err := p.PrintClones(uniq); err != nil {
				return err
			}
		}
	}

	return p.PrintFooter()
}