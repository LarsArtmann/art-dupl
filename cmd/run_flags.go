package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/printer/actionability"
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
	//art-dupl:accept standard context cleanup idiom
	defer cancel()

	return dispatchAnalysis(ctx, cmd, mergedConfig, sortBy)
}

// dispatchAnalysis routes to the appropriate analysis mode based on CLI flags.
// Handles --all (batch generation), --dump-tokens (token inspection), and the
// default standard-analysis path.
func dispatchAnalysis(ctx context.Context, cmd *cobra.Command, mergedConfig *config.Config, sortBy string) error {
	allFlag, _ := cmd.Flags().GetBool("all")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	if allFlag {
		return runAllModes(ctx, mergedConfig, sortBy, outputDir, cmd.ErrOrStderr())
	}

	if dumpTokens, _ := cmd.Flags().GetBool("dump-tokens"); dumpTokens {
		return dumpTokensOutput(ctx, mergedConfig, os.Stdout, cmd.ErrOrStderr())
	}

	if listPatterns, _ := cmd.Flags().GetBool("list-patterns"); listPatterns {
		actionability.ListActionabilityPatterns(os.Stdout)

		return nil
	}

	if recommend, _ := cmd.Flags().GetBool("recommend-threshold"); recommend {
		return runRecommendThreshold(ctx, mergedConfig)
	}

	if diffReportPath, _ := cmd.Flags().GetString("diff-report"); diffReportPath != "" {
		useJSON, _ := cmd.Flags().GetBool("json")

		return runDiffReport(ctx, mergedConfig, diffReportPath, useJSON, cmd.ErrOrStderr())
	}

	return runStandardAnalysis(ctx, cmd, mergedConfig, sortBy)
}

// runStandardAnalysis runs the default clone-detection pipeline and prints results.
func runStandardAnalysis(ctx context.Context, cmd *cobra.Command, mergedConfig *config.Config, sortBy string) error {
	duplChan, parseStats, _, err := executeAnalysis(
		ctx,
		mergedConfig,
		mergedConfig.Paths,
		mergedConfig.OutputFormat,
		cmd.ErrOrStderr(),
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	metadata := newReportMetadata(mergedConfig, sortBy)

	out, cleanup, err := openHTMLOutput(cmd)
	if err != nil {
		return err
	}

	if cleanup != nil {
		defer cleanup()
	}

	p := createPrinter(
		mergedConfig.OutputFormat,
		mergedConfig.Threshold,
		mergedConfig.DiffMode,
		metadata,
		GetVersion(),
	)(out, os.ReadFile)

	setJSONPrinterFilesCount(p, parseStats.FilesCount)

	if mergedConfig.RichText {
		if rts, ok := p.(printer.RichTextSetter); ok {
			rts.SetRichText(true)
		}
	}

	if mergedConfig.Explain {
		if es, ok := p.(printer.ExplainSetter); ok {
			es.SetExplain(true)
		}
	}

	suppression := buildSuppressionConfig(mergedConfig)

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

// buildDisabledPatternSet converts a list of pattern label strings to a set
// for O(1) lookup during actionability evaluation.
func buildDisabledPatternSet(labels []string) map[actionability.PatternLabel]bool {
	if len(labels) == 0 {
		return nil
	}

	set := make(map[actionability.PatternLabel]bool, len(labels))
	for _, label := range labels {
		set[actionability.PatternLabel(label)] = true
	}

	return set
}

// openHTMLOutput returns the output writer for the printer. When --html-out is
// set, it creates the file and returns a cleanup function to close it.
func openHTMLOutput(cmd *cobra.Command) (io.Writer, func(), error) {
	htmlOut, _ := cmd.Flags().GetString("html-out")
	if htmlOut == "" {
		return os.Stdout, nil, nil
	}

	f, err := os.Create(htmlOut)
	if err != nil {
		return nil, nil, duplerrors.Wrap(err, duplerrors.IOError, "creating HTML output file "+htmlOut)
	}

	return f, func() { _ = f.Close() }, nil
}

// setJSONPrinterFilesCount sets the files count on a JSON printer if the printer is a JSON printer.
func setJSONPrinterFilesCount(p printer.Printer, filesCount int) {
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}
}
