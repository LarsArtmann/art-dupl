package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/golangci/dupl/config"
	"github.com/golangci/dupl/job"
	"github.com/golangci/dupl/printer"
	"github.com/golangci/dupl/syntax"
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
			fmt.Fprintf(cli.Stderr(), "error loading config: %v\n", err)
			cli.Exit(1)
			return 1
		}
	}
	
	// Create CLI config from command line arguments
	cliConfig := &config.Config{}
	if *vendor {
		cliConfig.IncludeVendor = *vendor
	}
	if *verbose {
		cliConfig.Verbose = *verbose
	}
	if *threshold != defaultThreshold {
		cliConfig.Threshold = *threshold
	}
	if *files {
		cliConfig.FilesFromStdin = *files
	}
	
	if *html {
		cliConfig.OutputFormat = "html"
	} else if *plumbing {
		cliConfig.OutputFormat = "plumbing"
	} else if *json {
		cliConfig.OutputFormat = "json"
	}
	
	if flag.NArg() > 0 {
		cliConfig.Paths = flag.Args()
	}
	
	// Merge file and CLI configurations
	mergedConfig := config.MergeConfigs(fileConfig, cliConfig)
	
	// Validate merged configuration
	if err = config.ValidateConfig(mergedConfig); err != nil {
		fmt.Fprintf(cli.Stderr(), "configuration error: %v\n", err)
		cli.Exit(1)
		return 1
	}
	
	// Validate output format conflicts
	if mergedConfig.OutputFormat == "html" && *plumbing {
		fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or HTML output\n")
		cli.Exit(1)
		return 1
	}
	if mergedConfig.OutputFormat == "html" && *json {
		fmt.Fprintf(cli.Stderr(), "error: you can have either HTML or JSON output\n")
		cli.Exit(1)
		return 1
	}
	if mergedConfig.OutputFormat == "plumbing" && *json {
		fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or JSON output\n")
		cli.Exit(1)
		return 1
	}
	
	// Update global variables with merged config (for compatibility with existing code)
	paths = mergedConfig.Paths
	vendor = &mergedConfig.IncludeVendor
	verbose = &mergedConfig.Verbose
	threshold = &mergedConfig.Threshold
	files = &mergedConfig.FilesFromStdin

	if *verbose {
		log.Println("Building suffix tree")
	}
	schan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

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
	case "html":
		newPrinter = printer.NewHTML
	case "plumbing":
		newPrinter = printer.NewPlumbing
	case "json":
		newPrinter = printer.NewJSON
	default:
		newPrinter = printer.NewText
	}
	
	// Handle output file if specified
	var outputWriter io.Writer = cli.Stdout()
	if mergedConfig.OutputFile != "" {
		file, err := os.Create(mergedConfig.OutputFile)
		if err != nil {
			fmt.Fprintf(cli.Stderr(), "error creating output file: %v\n", err)
			cli.Exit(1)
			return 1
		}
		defer file.Close()
		outputWriter = file
	}
	
	p := newPrinter(outputWriter, os.ReadFile)
	if err := printDupls(p, duplChan); err != nil {
		fmt.Fprintf(cli.Stderr(), "error: %v\n", err)
		cli.Exit(1)
		return 1
	}
	return 0
}
