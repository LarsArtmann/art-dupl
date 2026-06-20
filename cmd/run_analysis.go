package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// treeBuildResult holds the result of a suffix tree build.
type treeBuildResult struct {
	tree       *suffixtree.STree
	data       []*syntax.Node
	parseStats job.ParseStats
	err        error
}

// buildParams holds parameters for suffix tree building.
type buildParams struct {
	ctx          context.Context
	paths        []string
	cfg          *config.Config
	filterParam  *gogenfilter.Filter
	filterStats  *FilterStats
	outputFormat config.OutputFormat
}

// getFilesChan creates a channel of file paths based on the build parameters.
func (p buildParams) getFilesChan() chan string {
	return filesFeedWithOptions(
		p.paths,
		p.cfg.FilesFromStdin,
		p.filterParam,
		p.filterStats,
		newGeneratorIncludes(p.cfg),
		p.cfg.IncludeVendor,
		p.cfg.IncludeNodeModules,
		p.cfg.Only,
	)
}

// printSearchStatus outputs the status message after tree building completes.
func printSearchStatus(cfg *config.Config, outputFormat config.OutputFormat) {
	if cfg.Verbose {
		_, _ = fmt.Fprintln(os.Stderr, "Searching for clones")
	} else if outputFormat == config.OutputFormatText {
		_, _ = fmt.Fprintln(os.Stderr, " ✅")
	}
}

// printBuildingStatus outputs the status message before tree building starts.
func printBuildingStatus(
	cfg *config.Config,
	outputFormat config.OutputFormat,
	verboseMsg string,
	textMsg string,
) {
	if cfg.Verbose {
		_, _ = fmt.Fprintln(os.Stderr, verboseMsg)
	} else if outputFormat == config.OutputFormatText {
		_, _ = fmt.Fprintln(os.Stderr, textMsg)
	}
}

// buildSuffixTree builds a suffix tree from provided paths.
func buildSuffixTree(params buildParams) treeBuildResult {
	printBuildingStatus(
		params.cfg,
		params.outputFormat,
		"Building suffix tree",
		"    📖 Parsing files and building analysis tree...",
	)

	if params.cfg.Incremental {
		return buildSuffixTreeIncremental(params)
	}

	return buildSuffixTreeStandard(params)
}

// buildSuffixTreeIncremental builds a suffix tree using incremental parsing with cache.
func buildSuffixTreeIncremental(params buildParams) treeBuildResult {
	if params.cfg.Verbose {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"🔍 Incremental mode enabled, cache dir: %s\n",
			params.cfg.CacheDir,
		)
	}

	incParser := job.NewIncrementalParser(
		params.cfg.CacheDir,
		params.cfg.ClearCache,
		detectionMode(params.cfg),
	)

	filesChan := params.getFilesChan()
	schan, incStatsChan := incParser.ParseIncremental(params.ctx, filesChan)
	tree, data, done := job.BuildTree(params.ctx, schan)
	<-done

	incStats := <-incStatsChan
	parseStats := job.ParseStats{
		ParseStatsMixin: job.ParseStatsMixin{
			FilesCount: incStats.FilesCount,
			LinesCount: incStats.LinesCount,
		},
	}

	tree.Update(&syntax.Node{Type: -1})
	printSearchStatus(params.cfg, params.outputFormat)

	return treeBuildResult{tree: tree, data: *data, parseStats: parseStats}
}

// buildSuffixTreeStandard builds a suffix tree using standard parsing without cache.
func buildSuffixTreeStandard(params buildParams) treeBuildResult {
	filesChan := params.getFilesChan()

	var (
		schan     chan []*syntax.Node
		statsChan chan job.ParseStats
	)

	if params.cfg.Workers > 1 {
		schan, statsChan = job.ParseParallel(
			params.ctx,
			filesChan,
			params.cfg.Workers,
			detectionMode(params.cfg),
		)
	} else {
		schan, statsChan = job.Parse(params.ctx, filesChan, detectionMode(params.cfg))
	}

	tree, data, done := job.BuildTree(params.ctx, schan)
	<-done

	parseStats := <-statsChan

	tree.Update(&syntax.Node{Type: -1})
	printSearchStatus(params.cfg, params.outputFormat)

	return treeBuildResult{tree: tree, data: *data, parseStats: parseStats}
}

// validatePaths checks that at least one given path exists on the filesystem.
// Skips validation when reading paths from stdin.
// If some paths are invalid, they're silently skipped during crawl (with stderr warning).
// Only returns an error when NONE of the specified paths exist.
func validatePaths(paths []string, filesFromStdin bool) error {
	if filesFromStdin || len(paths) == 0 {
		return nil
	}

	var invalid []string

	for _, path := range paths {
		_, err := os.Stat(path)
		if err != nil {
			invalid = append(invalid, path)
		}
	}

	if len(invalid) == len(paths) {
		return duplerrors.NewValidationError(
			"none of the specified paths exist: "+strings.Join(invalid, ", "),
			nil,
		)
	}

	return nil
}

// verboseFprintf prints a message to stderr if verbose mode is enabled.
func verboseFprintf(cfg *config.Config, msg string) {
	if cfg.Verbose {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}

// setupFilter creates a filter based on config settings.
func setupFilter(cfg *config.Config) (*gogenfilter.Filter, error) {
	type filterEntry struct {
		enabled bool
		option  gogenfilter.FilterOption
		logName string
	}

	genericLogName := "generic — any \"Code generated by\" comment"

	entries := []filterEntry{
		{!cfg.IncludeSQLC, gogenfilter.FilterSQLC, "sqlc"},
		{!cfg.IncludeTempl, gogenfilter.FilterTempl, "templ, default"},
		{!cfg.IncludeProtobuf, gogenfilter.FilterProtobuf, "protobuf"},
		{!cfg.IncludeMockgen, gogenfilter.FilterMockgen, "mockgen"},
		{!cfg.IncludeStringer, gogenfilter.FilterStringer, "stringer"},
		{!cfg.IncludeGeneric, gogenfilter.FilterGeneric, genericLogName},
	}

	var filterOptions []gogenfilter.FilterOption

	for _, e := range entries {
		if e.enabled {
			filterOptions = append(filterOptions, e.option)
			verboseFprintf(cfg, fmt.Sprintf(
				"Auto-generated code filtering enabled (%s)", e.logName,
			))
		}
	}

	excludePatterns := buildExcludePatterns(cfg)

	if len(filterOptions) > 0 || len(excludePatterns) > 0 || len(cfg.IncludePatterns) > 0 {
		configs := []gogenfilter.FilterConfig{
			gogenfilter.WithIncludePatterns(cfg.IncludePatterns...),
			gogenfilter.WithExcludePatterns(excludePatterns...),
		}

		if len(filterOptions) > 0 {
			filterConfig, err := gogenfilter.WithFilterOptions(filterOptions...)
			if err != nil {
				return nil, fmt.Errorf("failed to create filter options: %w", err)
			}

			configs = append(configs, filterConfig)
		}

		verboseFprintf(cfg, "Auto-generated code filtering enabled")

		fltr, err := gogenfilter.NewFilter(configs...)
		if err != nil {
			return nil, fmt.Errorf("failed to create filter: %w", err)
		}

		return fltr, nil
	}

	fltr, err := gogenfilter.NewFilter()
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}

	return fltr, nil
}

// buildExcludePatterns collects all file exclusion patterns from config,
// including IgnoreFiles, ExcludePatterns, and *_test.go when IgnoreTests is set.
func buildExcludePatterns(cfg *config.Config) []string {
	patterns := make([]string, 0, len(cfg.ExcludePatterns)+len(cfg.IgnoreFiles)+1)
	patterns = append(patterns, cfg.ExcludePatterns...)
	patterns = append(patterns, cfg.IgnoreFiles...)

	if cfg.IgnoreTests {
		patterns = append(patterns, "*_test.go")

		verboseFprintf(cfg, "Test file exclusion enabled (--ignore-tests)")
	}

	return patterns
}

// startProfiling begins profiling if enabled in config, returning the profile result.
func startProfiling(cfg *config.Config) job.ProfileResult {
	if !cfg.Profile {
		return job.ProfileResult{}
	}

	profile := job.StartProfile()
	_, _ = fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")

	return profile
}

// endProfiling ends profiling and prints results if enabled.
func endProfiling(cfg *config.Config, startProfile job.ProfileResult) {
	if !cfg.Profile {
		return
	}

	endProfile := job.EndProfile(startProfile)
	job.PrintProfileResult(endProfile)
}

// executeAnalysis runs the core duplicate analysis logic.
func executeAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	outputFormat config.OutputFormat,
) (chan syntax.Match, job.ParseStats, *FilterStats, error) {
	err := validatePaths(paths, cfg.FilesFromStdin)
	if err != nil {
		return nil, job.ParseStats{}, nil, duplerrors.WrapValidation(err, "path validation failed")
	}

	startProfile := startProfiling(cfg)

	filterParam, err := setupFilter(cfg)
	if err != nil {
		return nil, job.ParseStats{}, nil, duplerrors.Wrap(
			err,
			duplerrors.AnalysisError,
			fmt.Sprintf("failed to setup filter (outputFormat: %s)", outputFormat),
		)
	}

	var filterStats *FilterStats
	if filterParam != nil {
		filterStats = NewFilterStats(filterParam.FilterReasons())
	}

	if cfg.DetectionMethods.IsHashOnly() {
		ch, ps, fs, err := executeHashOnlyAnalysis(ctx, cfg, paths, filterParam, filterStats, outputFormat)

		return ch, ps, fs, err
	}

	result := buildSuffixTree(buildParams{
		ctx:          ctx,
		paths:        paths,
		cfg:          cfg,
		filterParam:  filterParam,
		filterStats:  filterStats,
		outputFormat: outputFormat,
	})
	if result.err != nil {
		return nil, job.ParseStats{}, nil, duplerrors.Wrap(
			result.err,
			duplerrors.AnalysisError,
			fmt.Sprintf(
				"failed to build suffix tree for paths %v (outputFormat: %s)",
				paths,
				outputFormat,
			),
		)
	}

	multiDetector := detection.NewMultiDetector(detection.Config{
		Methods: cfg.DetectionMethods.Strings(),
		Verbose: cfg.Verbose,
	}, result.data, result.tree)

	duplChan := spawnCloneDetection(ctx, multiDetector, cfg.Threshold)

	endProfiling(cfg, startProfile)

	return duplChan, result.parseStats, filterStats, nil
}

// spawnCloneDetection starts a goroutine that drains clone matches from the detector.
func spawnCloneDetection(
	ctx context.Context,
	detector *detection.MultiDetector,
	threshold int,
) chan syntax.Match {
	duplChan := make(chan syntax.Match)

	go func() {
		defer close(duplChan)

		matches := detector.FindDuplOver(ctx, threshold)
		for match := range matches {
			select {
			case duplChan <- match:
			case <-ctx.Done():
				return
			}
		}
	}()

	return duplChan
}
