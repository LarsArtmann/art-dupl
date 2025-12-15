package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
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

// Global CLI interface for dependency injection
var cli CLIInterface = &RealCLI{}

// Global CLI configuration
var cliConfig = NewCLIConfig()

// initFlags initializes all CLI flags
func initFlags() {
	// Flag setup is now handled by CLIConfig
}

func errorHandler(err error, format string, args ...interface{}) {
	cli.Stderr().WriteString(fmt.Sprintf("🚨 ERROR: %v\n", err))
	cli.Stderr().WriteString("💡 Suggestion:\n")
	if strings.Contains(err.Error(), "no such file") {
		cli.Stderr().WriteString("  • Check if file/directory exists\n")
		cli.Stderr().WriteString("  • Verify file path is correct\n")
	}
	if strings.Contains(err.Error(), "permission denied") {
		cli.Stderr().WriteString("  • Check file permissions\n")
		cli.Stderr().WriteString("  • Ensure you have read access\n")
	}
	if strings.Contains(err.Error(), "config") {
		cli.Stderr().WriteString("  • Use --help to see configuration options\n")
		cli.Stderr().WriteString("  • Check config file format is valid JSON\n")
	}
	cli.Stderr().WriteString(fmt.Sprintf(format, args...))
	cli.Exit(1)
}

// Run executes the CLI with provided interface and arguments
func Run(cliInterface CLIInterface, args []string) {
	cli = cliInterface
	
	// Initialize flags
	initFlags()
	
	// Parse arguments after flag parsing
	if len(args) > 0 && args[0] != "--help" && args[0] != "-h" {
		cliConfig.Paths = args
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "art-dupl",
		Short: "Find duplicated code fragments",
		Long: `art-dupl finds duplicated code fragments in Go source code.

It provides multiple output formats and detection methods for comprehensive
code duplication analysis.`,
		Run: func(cmd *cobra.Command, args []string) {
			runCobraCommand(cmd, args)
		},
	}

	// Add configuration flags
	cliConfig.AddFlagsToCommand(*rootCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		errorHandler(err, "")
	}
}

func runCobraCommand(cmd *cobra.Command, args []string) {
	// Update paths from Cobra args
	if len(args) > 0 {
		cliConfig.Paths = args
	}

	// Create configuration
	cfg, err := createConfiguration()
	if err != nil {
		errorHandler(err, "")
		return
	}

	// Handle different modes
	if *cliConfig.Files {
		handleFilesMode(cfg)
		return
	}

	// Regular analysis mode
	if len(cliConfig.Paths) == 0 {
		cliConfig.Paths = []string{"."}
	}

	runAnalysis(cfg)
}

func createConfiguration() (*config.Config, error) {
	cfg := config.New()

	// Load configuration file if specified
	if *cliConfig.ConfigFile != "" {
		if err := cfg.LoadFile(*cliConfig.ConfigFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %v", err)
		}
	}

	// Override with CLI flags
	if cliConfig.GetThreshold() != 15 {
		cfg.Threshold = cliConfig.GetThreshold()
	}
	if *cliConfig.Vendor {
		cfg.Vendor = true
	}
	if cliConfig.IsVerbose() {
		cfg.Verbose = true
	}
	if *cliConfig.SortBy != "size" {
		cfg.SortBy = *cliConfig.SortBy
	}
	if len(cliConfig.GetOutputFormats()) > 0 && cliConfig.GetOutputFormats()[0] != "text" {
		cfg.OutputFormats = cliConfig.GetOutputFormats()
	}

	return cfg, nil
}

func handleFilesMode(cfg *config.Config) {
	var paths []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			paths = append(paths, path)
		}
	}

	if err := scanner.Err(); err != nil {
		errorHandler(fmt.Errorf("failed to read from stdin: %v", err), "")
		return
	}

	if len(paths) == 0 {
		errorHandler(fmt.Errorf("no file paths provided via stdin"), "")
		return
	}

	cliConfig.Paths = paths
	runAnalysis(cfg)
}

func runAnalysis(cfg *config.Config) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		errorHandler(err, "")
		return
	}

	// Determine output mode
	outputFormats := cliConfig.GetOutputFormats()
	outputDir := ""
	
	// Check if all formats should be generated
	if len(outputFormats) > 1 {
		outputDir = "--output-dir"
		outputFormats = []string{"all"}
	}

	// Run analysis
	for _, format := range outputFormats {
		if outputDir != "" {
			// Run multi-format analysis
			runAllFormats(cfg, outputDir)
			return
		} else {
			// Run single format analysis
			runSingleFormat(cfg, format)
		}
	}
}

func runSingleFormat(cfg *config.Config, format string) {
	filesFeed := job.FilesFromPaths(cliConfig.Paths, cfg.Vendor)
	schan, filesCountChan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

	filesCount := <-filesCountChan

	// Get matches based on detection method
	var duplChan chan syntax.Match
	if cfg.DetectionMethod == config.DetectionMethodHash {
		multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
		duplChan = make(chan syntax.Match)
		go func() {
			defer close(duplChan)
			matches := multiDetector.FindDuplOver(cfg.Threshold)
			for match := range matches {
				duplChan <- match
			}
		}()
	} else {
		duplChan = t.FindDuplOver(cfg.Threshold)
	}

	// Create appropriate printer and run analysis
	var p printer.Interface
	switch format {
	case "json":
		p = printer.NewJSON(cli.Stdout(), os.ReadFile)
	case "html":
		p = printer.NewHTML(cli.Stdout(), os.ReadFile)
	case "plumbing":
		p = printer.NewPlumbing(cli.Stdout(), os.ReadFile)
	default:
		p = printer.NewText(cli.Stdout(), os.ReadFile)
	}

	if err := p.PrintDupls(duplChan, filesCount, cfg.SortBy, cfg.DetectionMethod); err != nil {
		errorHandler(err, "")
		return
	}
}

func runAllFormats(cfg *config.Config, outputDir string) {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errorHandler(fmt.Errorf("failed to create output directory: %v", err), "")
		return
	}

	if cfg.Verbose {
		cli.Stderr().WriteString(fmt.Sprintf("📁 Output directory: %s\n", outputDir))
	}

	filesFeed := job.FilesFromPaths(cliConfig.Paths, cfg.Vendor)
	
	if cfg.Verbose {
		cli.Stderr().WriteString("🔍 Parsing files...\n")
	}

	schan, filesCountChan := job.Parse(filesFeed())
	
	if cfg.Verbose {
		cli.Stderr().WriteString("🌳 Building suffix tree...\n")
	}

	t, data, done := job.BuildTree(schan)
	<-done

	// Get file count
	filesCount := <-filesCountChan

	if cfg.Verbose {
		cli.Stderr().WriteString(fmt.Sprintf("📄 Processed %d files\n", filesCount))
	}

	// Define formats to generate
	formats := []struct {
		name     string
		ext      string
		newPrinter func(io.Writer, func(string) ([]byte, error)) printer.Interface
	}{
		{"text", ".txt", func(w io.Writer, rf func(string) ([]byte, error)) printer.Interface { return printer.NewText(w, rf) }},
		{"html", ".html", func(w io.Writer, rf func(string) ([]byte, error)) printer.Interface { return printer.NewHTML(w, rf) }},
		{"json", ".json", func(w io.Writer, rf func(string) ([]byte, error)) printer.Interface { return printer.NewJSON(w, rf) }},
		{"plumbing", ".txt", func(w io.Writer, rf func(string) ([]byte, error)) printer.Interface { return printer.NewPlumbing(w, rf) }},
	}

	// Run analysis for each detection method
	methods := []config.DetectionMethod{cfg.DetectionMethod}
	
	for _, method := range methods {
		if cfg.Verbose {
			cli.Stderr().WriteString(fmt.Sprintf("🔎 Running analysis with %s method...\n", method))
		}

		// Get matches based on detection method
		var duplChan chan syntax.Match
		if method == config.DetectionMethodHash {
			multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
			duplChan = make(chan syntax.Match)
			go func() {
				defer close(duplChan)
				matches := multiDetector.FindDuplOver(cfg.Threshold)
				for match := range matches {
					duplChan <- match
				}
			}()
		} else {
			duplChan = t.FindDuplOver(cfg.Threshold)
		}

		// Generate all output formats
		for _, fmtInfo := range formats {
			filename := filepath.Join(outputDir, fmt.Sprintf("%s%s", method, fmtInfo.ext))
			file, err := os.Create(filename)
			if err != nil {
				errorHandler(fmt.Errorf("failed to create %s file: %v", filename, err), "")
				return
			}
			defer file.Close()

			p := fmtInfo.newPrinter(file, os.ReadFile)
			if err := p.PrintDupls(duplChan, filesCount, cfg.SortBy, method); err != nil {
				errorHandler(err, "")
				return
			}

			if cfg.Verbose {
				cli.Stderr().WriteString(fmt.Sprintf("✅ Generated %s\n", filename))
			}
		}
	}

	if cfg.Verbose {
		cli.Stderr().WriteString("🎉 Analysis complete!\n")
	}
}

// filterFiles filters files based on vendor and other criteria
func filterFiles(files []string, includeVendor bool) []string {
	var result []string
	for _, file := range files {
		// Skip vendor directory unless explicitly included
		if !includeVendor && strings.Contains(file, VendorDirPrefix) {
			continue
		}
		result = append(result, file)
	}
	return result
}

// expandPaths expands input paths to include all Go files
func expandPaths(paths []string) []string {
	var allFiles []string
	for _, path := range paths {
		if util.IsDir(path) {
			files := util.GoFilesInDir(path)
			sort.Strings(files)
			allFiles = append(allFiles, files...)
		} else {
			allFiles = append(allFiles, path)
		}
	}
	return allFiles
}

func main() {
	Run(&RealCLI{}, os.Args[1:])
}