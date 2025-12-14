package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
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

// Command line flags
var (
	configFile    = flag.String("config", "", "path to configuration file (JSON format)")
	vendor        = flag.Bool("vendor", false, "include vendor directory in analysis")
	verbose       = flag.Bool("v", false, "enable verbose logging to show processing progress")
	verboseLong   = flag.Bool("verbose", false, "enable verbose logging to show processing progress")
	threshold     = flag.Int("t", 15, "minimum token sequence size to consider as clone")
	thresholdLong = flag.Int("threshold", 15, "minimum token sequence size to consider as clone")
	files         = flag.Bool("files", false, "read file names from stdin, one per line")
	html          = flag.Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	json          = flag.Bool("json", false, "output structured JSON format with metadata and statistics")
	plumbing      = flag.Bool("plumbing", false, "output machine-readable plumbing format for script integration")
	sortBy        = flag.String("sort", "size", "sort clone groups by: size, occurrence, hash")
	paths         []string
)

const defaultThreshold = 15

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
	} else if *json {
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
	if mergedConfig.OutputFormat == "html" && *json {
		if _, err := fmt.Fprintf(cli.Stderr(), "error: you can have either HTML or JSON output\n"); err != nil {
			cli.Exit(1)
			return 1
		}
		cli.Exit(1)
		return 1
	}
	if mergedConfig.OutputFormat == "plumbing" && *json {
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

	if err := printDupls(p, duplChan); err != nil {
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
