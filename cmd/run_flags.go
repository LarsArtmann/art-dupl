package cmd

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/spf13/cobra"
)

// wrapAnalysisError wraps an analysis error with context about the paths being analyzed.
func wrapAnalysisError(err error, paths []string) error {
	return duplerrors.Wrap(
		err,
		duplerrors.AnalysisError,
		fmt.Sprintf("analysis failed for paths %v", paths),
	)
}

// setDetectionMethods parses and sets detection methods from flag value.
func setDetectionMethods(appConfig *config.Config, detectionMethods string) error {
	parsedMethods, err := config.ParseDetectionMethods(detectionMethods)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("invalid detection methods %q", detectionMethods),
		)
	}

	appConfig.DetectionMethods = parsedMethods

	return nil
}

// runCmd implements Cobra command execution.
//
//nolint:funlen // Command execution requires handling many CLI flags and configuration options
func runCmd(cmd *cobra.Command, args []string) error {
	html, _ := cmd.Flags().GetBool("html")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sarif, _ := cmd.Flags().GetBool("sarif")
	simpleJSON, _ := cmd.Flags().GetBool("simple-json")
	sortBy, _ := cmd.Flags().GetString("sort")

	// Validate sorting criteria
	_, err := config.ParseSortCriteria(sortBy)
	if err != nil {
		return duplerrors.WrapValidation(err, fmt.Sprintf("invalid --sort value %q", sortBy))
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	mergedConfig, err := BuildConfigFromFlags(cmd, args)
	if err != nil {
		return err
	}

	// Set output format (specific to run command)
	switch {
	case html:
		mergedConfig.OutputFormat = config.OutputFormatHTML
	case plumbing:
		mergedConfig.OutputFormat = config.OutputFormatPlumbing
	case sarif:
		mergedConfig.OutputFormat = config.OutputFormatSARIF
	case simpleJSON:
		mergedConfig.OutputFormat = config.OutputFormatSimpleJSON
	case jsonFlag:
		mergedConfig.OutputFormat = config.OutputFormatJSON
	}

	// Re-validate after adding output format
	err = config.ValidateConfig(mergedConfig)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
	}

	// Get context from Cobra (includes Fang's signal handling)
	ctx := cmd.Context()

	if allFlag {
		return runAllModes(ctx, mergedConfig, sortBy, outputDir)
	}
	// Add timeout context if specified
	ctx, cancel := utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	duplChan, parseStats, _, err := executeAnalysis(
		ctx,
		mergedConfig,
		mergedConfig.Paths,
		mergedConfig.OutputFormat,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	// Check for cancellation after analysis completes
	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}

	// Build metadata for HTML report
	metadata := newReportMetadata(mergedConfig, sortBy)

	p := createPrinter(
		mergedConfig.OutputFormat,
		mergedConfig.Threshold,
		mergedConfig.DiffMode,
		metadata,
		GetVersion(),
	)(
		os.Stdout,
		os.ReadFile,
	)

	setJSONPrinterFilesCount(p, parseStats.FilesCount)

	if mergedConfig.RichText {
		if rts, ok := p.(printer.RichTextSetter); ok {
			rts.SetRichText(true)
		}
	}

	// Convert detection methods to comma-separated string
	detectionMethodStr := detectionMethodsToString(mergedConfig.DetectionMethods)

	err = printDupls(
		ctx,
		p,
		os.ReadFile,
		duplChan,
		config.SortCriteria(sortBy),
		mergedConfig.Threshold,
		detectionMethodStr,
		mergedConfig.Semantic,
	)
	if err != nil {
		return duplerrors.Wrap(
			err,
			duplerrors.AnalysisError,
			fmt.Sprintf(
				"failed to print duplicates (sortBy: %s, threshold: %d)",
				sortBy,
				mergedConfig.Threshold,
			),
		)
	}

	return nil
}

// setJSONPrinterFilesCount sets the files count on a JSON printer if the printer is a JSON printer.
func setJSONPrinterFilesCount(p printer.Printer, filesCount int) {
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}
}
