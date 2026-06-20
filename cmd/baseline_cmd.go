package cmd

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/baseline"
	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/spf13/cobra"
)

const baselinePathFlag = "baseline-path"

// NewBaselineCommand creates the `baseline` subcommand that snapshots the
// currently-detected clones into a baseline file for later CI comparison.
func NewBaselineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baseline [flags] [paths...]",
		Short: "Record accepted clones to a baseline file",
		Long: "Runs clone detection and writes every detected clone-group hash to a\n" +
			"baseline file. Subsequent `art-dupl check` runs report only clones whose\n" +
			"hash is NOT in this baseline (i.e. newly introduced duplication).",
		Args: cobra.ArbitraryArgs,
		RunE: runBaseline,
	}

	addSharedFlags(cmd)
	cmd.Flags().String(baselinePathFlag, baseline.DefaultBaselinePath, "path to write the baseline file")

	return cmd
}

// NewCheckCommand creates the `check` subcommand that reports only new clones
// relative to a baseline, exiting non-zero if any are found. Designed for CI.
func NewCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [flags] [paths...]",
		Short: "Report clones new relative to a baseline (CI mode)",
		Long: "Runs clone detection and compares against the baseline. Only clones with\n" +
			"a hash not present in the baseline are reported. Exits with code 1 if any\n" +
			"new clones are found, 0 otherwise. Run `art-dupl baseline` first to create\n" +
			"the baseline.",
		Args: cobra.ArbitraryArgs,
		RunE: runCheck,
	}

	addSharedFlags(cmd)
	cmd.Flags().String(baselinePathFlag, baseline.DefaultBaselinePath, "path to the baseline file")
	cmd.Flags().Bool("json", false, "output new clones as JSON")

	return cmd
}

func runBaseline(c *cobra.Command, arguments []string) error {
	ctx := c.Context()

	mergedConfig, err := BuildConfigFromFlags(c, arguments)
	if err != nil {
		return err
	}

	ctx, cancel := utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	startProfile := job.StartProfile()

	duplChan, parseStats, filterStats, err := executeAnalysis(
		ctx, mergedConfig, mergedConfig.Paths, config.OutputFormatText,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	_ = job.EndProfile(startProfile)

	groups := printer.BuildCloneGroups(duplChan)
	groups = printer.EliminateOverlaps(groups)
	keys := getSortedKeys(groups, config.SortByHash)

	bf := baseline.NewFile(mergedConfig.Threshold)
	recorder := newBaselineRecorderPrinter(bf)

	err = printCloneGroups(recorder, os.ReadFile, groups, keys, config.SortByHash,
		mergedConfig.Semantic, mergedConfig.EffectiveSuppressTestLow(),
		mergedConfig.EffectiveTestThreshold())
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "recording baseline")
	}

	path, _ := c.Flags().GetString(baselinePathFlag)

	err = bf.Save(path)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.IOError, "saving baseline to "+path)
	}

	fmt.Fprintf(os.Stderr,
		"Recorded %d clone groups (%d files, %d filtered) to %s\n",
		bf.Len(), parseStats.FilesCount, filteredCount(filterStats), path)

	return nil
}

func runCheck(c *cobra.Command, arguments []string) error {
	ctx := c.Context()

	path, _ := c.Flags().GetString(baselinePathFlag)
	if !baseline.Exists(path) {
		return duplerrors.NewValidationError(
			fmt.Sprintf("baseline file %s not found; run `art-dupl baseline` first", path), nil,
		)
	}

	bf, err := baseline.Load(path)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.IOError, "loading baseline "+path)
	}

	mergedConfig, err := BuildConfigFromFlags(c, arguments)
	if err != nil {
		return err
	}

	ctx, cancel := utils.ApplyTimeout(ctx, mergedConfig.Timeout)
	defer cancel()

	duplChan, _, _, err := executeAnalysis(
		ctx, mergedConfig, mergedConfig.Paths, config.OutputFormatText,
	)
	if err != nil {
		return wrapAnalysisError(err, mergedConfig.Paths)
	}

	groups := printer.BuildCloneGroups(duplChan)
	groups = printer.EliminateOverlaps(groups)
	keys := getSortedKeys(groups, config.SortByHash)

	jsonOut, _ := c.Flags().GetBool("json")

	inner := printer.NewText(os.Stdout, os.ReadFile)
	if jsonOut {
		inner = printer.NewJSON(os.Stdout, os.ReadFile)
	}

	filter := newBaselineFilterPrinter(inner, bf)

	err = printHeader(filter, config.SortByHash, mergedConfig.Threshold)
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "check header")
	}

	err = printCloneGroups(filter, os.ReadFile, groups, keys, config.SortByHash,
		mergedConfig.Semantic, mergedConfig.EffectiveSuppressTestLow(),
		mergedConfig.EffectiveTestThreshold())
	if err != nil {
		return duplerrors.Wrap(err, duplerrors.AnalysisError, "check output")
	}

	footerErr := filter.PrintFooter()
	if footerErr != nil {
		return duplerrors.Wrap(footerErr, duplerrors.IOError, "check footer")
	}

	if filter.newCloneCount > 0 {
		fmt.Fprintf(os.Stderr,
			"\n\U0001f534 %d new clone group(s) detected (baseline had %d). "+
				"Run `art-dupl baseline` to update.\n",
			filter.newCloneCount, bf.Len())

		// Returning an error makes Cobra/Fang exit with code 1 for CI gates.
		return duplerrors.NewValidationError(
			fmt.Sprintf("%d new clone group(s) introduced", filter.newCloneCount), nil,
		)
	}

	fmt.Fprintf(os.Stderr, "\n\u2705 No new clones detected (baseline: %d groups).\n", bf.Len())

	return nil
}

// filteredCount returns the total number of filtered files, or 0 if nil.
func filteredCount(fs *FilterStats) int {
	if fs == nil {
		return 0
	}

	return fs.TotalFiltered()
}
