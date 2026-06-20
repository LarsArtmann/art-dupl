package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// runAllModes runs all detection methods and generates all output formats.
//
//nolint:funlen // Function orchestrates multi-format output generation
func runAllModes(ctx context.Context, cfg *config.Config, sortBy, outputDir string) error {
	// Set detection methods to all available methods
	cfg.DetectionMethods = config.AllDetectionMethods()

	// Set output directory if not specified
	if outputDir == "" {
		outputDir = "reports/art-dupl"
	}

	err := os.MkdirAll(outputDir, 0o750)
	if err != nil {
		return fmt.Errorf(
			"failed to create output directory %q (sortBy=%s): %w",
			outputDir,
			sortBy,
			err,
		)
	}

	fmt.Fprintf(
		os.Stderr,
		"📂 Running all detection methods and generating all output formats in %s...\n",
		outputDir,
	)

	// Run analysis once
	duplChan, parseStats, _, err := executeAnalysis(ctx, cfg, cfg.Paths, cfg.OutputFormat)
	if err != nil {
		return fmt.Errorf(
			"analysis failed for paths %v (sortBy=%s, outputDir=%s): %w",
			cfg.Paths,
			sortBy,
			outputDir,
			err,
		)
	}

	// Check for cancellation after analysis completes
	if ctx.Err() != nil {
		return fmt.Errorf(
			"context cancelled (sortBy=%s, outputDir=%s): %w",
			sortBy,
			outputDir,
			ctx.Err(),
		)
	}

	// Convert channel to slice for reuse
	matches := collectMatches(duplChan)

	// Generate all output formats
	formats := config.AllOutputFormats()
	sortByEnum := config.SortCriteria(sortBy)

	// Convert detection methods to comma-separated string
	detectionMethodStr := detectionMethodsToString(cfg.DetectionMethods)

	for _, format := range formats {
		filename := filepath.Join(outputDir, "report."+string(format))

		err := writeFormatFile(
			ctx,
			cfg,
			matches,
			parseStats,
			format,
			filename,
			sortByEnum,
			detectionMethodStr,
		)
		if err != nil {
			return fmt.Errorf(
				"write format file failed (matches=%v, parseStats=%v, format=%s, filename=%s, sortByEnum=%s, detectionMethodStr=%s): %w",
				matches,
				parseStats,
				format,
				filename,
				sortByEnum,
				detectionMethodStr,
				err,
			)
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

// writeFormatFile writes a single output format to a file.
func writeFormatFile(
	ctx context.Context,
	cfg *config.Config,
	matches []syntax.Match,
	parseStats job.ParseStats,
	format config.OutputFormat,
	filename string,
	sortByEnum config.SortCriteria,
	detectionMethodStr string,
) error {
	// #nosec G304 -- filename is constructed from controlled config output dir and format
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file %q: %w", filename, err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close file %q: %v\n", filename, closeErr)
		}
	}()

	// Build metadata for HTML report
	metadata := newReportMetadata(cfg, sortByEnum.String())

	p := createPrinter(
		format,
		cfg.Threshold,
		cfg.DiffMode,
		metadata,
		GetVersion(),
	)(
		file,
		os.ReadFile,
	)

	setJSONPrinterFilesCount(p, parseStats.FilesCount)

	// Create channel from matches for this printer
	matchChan := make(chan syntax.Match)

	go func() {
		defer close(matchChan)

		for _, match := range matches {
			select {
			case matchChan <- match:
			case <-ctx.Done():
				return
			}
		}
	}()

	err = printDupls(
		ctx,
		p,
		os.ReadFile,
		matchChan,
		sortByEnum,
		cfg.Threshold,
		detectionMethodStr,
		cfg.Semantic,
		cfg.SuppressTestLow,
		cfg.TestThreshold,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to print %s format (matches=%v, parseStats=%v, filename=%s, sortByEnum=%s, detectionMethodStr=%s): %w",
			format,
			matches,
			parseStats,
			filename,
			sortByEnum,
			detectionMethodStr,
			err,
		)
	}

	return nil
}
