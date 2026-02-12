package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// runAllModes runs all detection methods and generates all output formats.
func runAllModes(ctx context.Context, cfg *config.Config, sortBy, outputDir string) error {
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
	duplChan, parseStats, _, err := executeAnalysis(ctx, cfg, cfg.Paths, cfg.OutputFormat)
	if err != nil {
		return fmt.Errorf("analysis failed for paths %v: %w", cfg.Paths, err)
	}

	// Convert channel to slice for reuse
	matches := collectMatches(duplChan)

	// Generate all output formats
	formats := config.AllOutputFormats()
	sortByEnum := printer.SortBy(sortBy)

	// Convert detection methods to comma-separated string
	detectionMethodStr := detectionMethodsToString(cfg.DetectionMethods)

	for _, format := range formats {
		filename := filepath.Join(outputDir, "report."+string(format))
		//nolint:gosec //G304 filename is constructed from controlled config output dir and format
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", filename, err)
		}
		// TODO: DEFER ANTI-PATTERN - defer in loop. Consider refactoring to check error immediately
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close file %q: %v\n", filename, err)
			}
		}()

		p := createPrinter(format, cfg.Threshold)(file, os.ReadFile)

		if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
			jsonPrinter.SetFilesCount(parseStats.FilesCount)
		}

		// Create channel from matches for this printer
		matchChan := make(chan syntax.Match)
		go func() {
			defer close(matchChan)
			for _, match := range matches {
				matchChan <- match
			}
		}()

		if err := printDupls(p, matchChan, sortByEnum, cfg.Threshold, detectionMethodStr); err != nil {
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
