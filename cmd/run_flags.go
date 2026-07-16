package cmd

import (
	"context"
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

// parseOutputFormat reads the output-format flags from the cobra command and
// returns the corresponding config.OutputFormat. Returns config.OutputFormatText
// when no format flag is set.
func parseOutputFormat(cmd *cobra.Command) config.OutputFormat {
	if v, _ := cmd.Flags().GetBool("html"); v {
		return config.OutputFormatHTML
	}

	if v, _ := cmd.Flags().GetBool("plumbing"); v {
		return config.OutputFormatPlumbing
	}

	if v, _ := cmd.Flags().GetBool("sarif"); v {
		return config.OutputFormatSARIF
	}

	if v, _ := cmd.Flags().GetBool("simple-json"); v {
		return config.OutputFormatSimpleJSON
	}

	if v, _ := cmd.Flags().GetBool("json"); v {
		return config.OutputFormatJSON
	}

	return config.OutputFormatText
}

// runCmd implements Cobra command execution.
func runCmd(cmd *cobra.Command, args []string) error {
	if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
		_ = os.Setenv("NO_COLOR", "1")
	}

	sortBy, _ := cmd.Flags().GetString("sort")

	_, err := config.ParseSortCriteria(sortBy)
	if err != nil {
		return duplerrors.WrapValidation(err, fmt.Sprintf("invalid --sort value %q", sortBy))
	}

	mergedConfig, err := BuildConfigFromFlags(cmd, args)
	if err != nil {
		return err
	}

	mergedConfig.OutputFormat = parseOutputFormat(cmd)

	err = config.ValidateConfig(mergedConfig)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
	}

	ctx := cmd.Context()
	ctx, cancel := utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	allFlag, _ := cmd.Flags().GetBool("all")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	if allFlag {
		return runAllModes(ctx, mergedConfig, sortBy, outputDir)
	}

	if dumpTokens, _ := cmd.Flags().GetBool("dump-tokens"); dumpTokens {
		return dumpTokensOutput(ctx, mergedConfig, os.Stdout)
	}

	return runStandardAnalysis(ctx, mergedConfig, sortBy)
}

// runStandardAnalysis runs the default clone-detection pipeline and prints results.
func runStandardAnalysis(ctx context.Context, mergedConfig *config.Config, sortBy string) error {
	duplChan, parseStats, _, err := executeAnalysis(
		ctx,
		mergedConfig,
		mergedConfig.Paths,
		mergedConfig.OutputFormat,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	metadata := newReportMetadata(mergedConfig, sortBy)

	p := createPrinter(
		mergedConfig.OutputFormat,
		mergedConfig.Threshold,
		mergedConfig.DiffMode,
		metadata,
		GetVersion(),
	)(os.Stdout, os.ReadFile)

	setJSONPrinterFilesCount(p, parseStats.FilesCount)

	if mergedConfig.RichText {
		if rts, ok := p.(printer.RichTextSetter); ok {
			rts.SetRichText(true)
		}
	}

	suppression := SuppressionConfig{
		SuppressTestLow: mergedConfig.EffectiveSuppressTestLow(),
		TestThreshold:   mergedConfig.EffectiveTestThreshold(),
		MinLines:        mergedConfig.MinLines,
	}

	err = printDupls(
		ctx,
		p,
		os.ReadFile,
		duplChan,
		config.SortCriteria(sortBy),
		mergedConfig.Threshold,
		detectionMethodsToString(mergedConfig.DetectionMethods),
		mergedConfig.DetectionMode.IsSemantic(),
		suppression,
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
