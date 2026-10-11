package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// treeBuildResult holds the result of a suffix tree build.
type treeBuildResult struct {
	tree         *suffixtree.STree
	data         []*syntax.Node
	parseStats   job.ParseStats
	genericDecls map[string]int
	err          error
}

// buildParams holds parameters for suffix tree building.
type buildParams struct {
	ctx          context.Context
	paths        []string
	cfg          *config.Config
	filterParam  *gogenfilter.Filter
	filterStats  *FilterStats
	outputFormat config.OutputFormat
	stderr       io.Writer
}

// getFilesChan creates a channel of file paths based on the build parameters.
func (p buildParams) getFilesChan() chan string {
	var gitignore *GitignoreMatcher
	if !p.cfg.IncludeIgnored {
		gitignore = LoadGitignore(p.paths, p.stderr)
	}

	return filesFeedWithOptions(
		p.ctx,
		p.paths,
		p.cfg.FilesFromStdin,
		p.filterParam,
		p.filterStats,
		newGeneratorIncludes(p.cfg),
		p.cfg.IncludeVendor,
		p.cfg.IncludeNodeModules,
		p.cfg.IncludeExamples,
		p.cfg.Only,
		gitignore,
		p.stderr,
	)
}

// printSearchStatus outputs the status message after tree building completes.
func printSearchStatus(w io.Writer, cfg *config.Config, outputFormat config.OutputFormat) {
	if cfg.Quiet {
		return
	}

	if cfg.Verbose {
		_, _ = fmt.Fprintln(w, "Searching for clones")
	} else if outputFormat == config.OutputFormatText {
		_, _ = fmt.Fprintln(w, " ✅")
	}
}

// printBuildingStatus outputs the status message before tree building starts.
func printBuildingStatus(
	w io.Writer,
	cfg *config.Config,
	outputFormat config.OutputFormat,
	verboseMsg string,
	textMsg string,
) {
	if cfg.Quiet {
		return
	}

	if cfg.Verbose {
		_, _ = fmt.Fprintln(w, verboseMsg)
	} else if outputFormat == config.OutputFormatText {
		_, _ = fmt.Fprintln(w, textMsg)
	}
}

// finalizeTreeBuild consumes the parsed node stream into a suffix tree, reads
// the parse statistics via readStats, applies the terminator node, prints the
// search status, and returns the assembled result. readStats is invoked only
// after BuildTree has drained schan, so the parser's stats channel is ready.
func finalizeTreeBuild(
	params buildParams,
	schan chan []*syntax.Node,
	readStats func() job.ParseStats,
) treeBuildResult {
	tree, data, done := job.BuildTree(params.ctx, schan)
	if err := <-done; err != nil {
		return treeBuildResult{tree: tree, data: *data}
	}

	parseStats := readStats()

	if err := tree.Update(&syntax.Node{Type: -1}); err != nil {
		logger.Default.Error("suffix tree terminator update failed", "err", err)
	}

	printSearchStatus(params.stderr, params.cfg, params.outputFormat)

	return treeBuildResult{tree: tree, data: *data, parseStats: parseStats}
}

// buildSuffixTree builds a suffix tree from provided paths.
func buildSuffixTree(params buildParams) treeBuildResult {
	printBuildingStatus(
		params.stderr,
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
			params.stderr,
			"🔍 Incremental mode enabled, cache dir: %s\n",
			params.cfg.CacheDir,
		)
	}

	incParser := job.NewIncrementalParser(
		params.cfg.CacheDir,
		params.cfg.ClearCache,
		detectionMode(params.cfg),
		params.cfg.MaxChildrenSerial,
		params.cfg.MaxCacheEntries,
		params.cfg.MemoryCacheEntries,
	)

	filesChan := params.getFilesChan()
	filesChan = progressFilesChan(params.ctx, filesChan, params.cfg, params.outputFormat, params.stderr)

	// Load type-aware data if enabled (supports --type-aware + --incremental)
	var typeInfos golang.TypeAwareData

	if params.cfg.TypeAware || params.cfg.SuggestGenerics {
		typeInfos, filesChan = loadTypeAwareData(params.ctx, filesChan, params.stderr, params.cfg.SuggestGenerics)
		incParser.SetTypeAwareData(typeInfos)
	}

	return withGenericDecls(
		buildSuffixTreeIncrementalPass(params, incParser, filesChan, true),
		golang.CollectGenericDecls(typeInfos),
	)
}

// withGenericDecls attaches a generic-declaration index to a tree build
// result. The index is computed once at type-data load time (empty when the
// run did not load type-aware data) and consumed by the print phase's
// reuse-existing-generic verdict.
func withGenericDecls(result treeBuildResult, decls map[string]int) treeBuildResult {
	result.genericDecls = decls

	return result
}

// buildSuffixTreeIncrementalPass parses filesChan through the given incremental
// parser and finalizes the tree build. Shared by single-mode and combined-mode
// orchestration; printCacheStats controls the per-pass cache report so the
// combined mode reports once (cumulatively, after the last pass).
func buildSuffixTreeIncrementalPass(
	params buildParams,
	incParser *job.IncrementalParser,
	filesChan chan string,
	printCacheStatsReport bool,
) treeBuildResult {
	var (
		schan        chan []*syntax.Node
		incStatsChan chan job.IncrementalStats
	)

	if params.cfg.Workers != 1 {
		schan, incStatsChan = incParser.ParseIncrementalParallel(params.ctx, filesChan, params.cfg.Workers)
	} else {
		schan, incStatsChan = incParser.ParseIncremental(params.ctx, filesChan)
	}

	return finalizeTreeBuild(params, schan, func() job.ParseStats {
		incStats := <-incStatsChan

		cacheStats := incParser.GetCacheStats()

		if printCacheStatsReport {
			printCacheStats(params.stderr, params.cfg, cacheStats)
		}

		return job.ParseStats{
			FilesCount: incStats.FilesCount,
			LinesCount: incStats.LinesCount,
			Cache: &job.RunCacheStats{
				Hits:       incStats.CacheHits,
				Misses:     incStats.CacheMisses,
				MemoryHits: cacheStats.MemHits,
				Entries:    cacheStats.Size,
			},
		}
	})
}

// buildSuffixTreeStandard builds a suffix tree using standard parsing without cache.
func buildSuffixTreeStandard(params buildParams) treeBuildResult {
	filesChan := params.getFilesChan()
	filesChan = progressFilesChan(params.ctx, filesChan, params.cfg, params.outputFormat, params.stderr)

	var typeInfos golang.TypeAwareData
	if params.cfg.TypeAware || params.cfg.SuggestGenerics {
		typeInfos, filesChan = loadTypeAwareData(params.ctx, filesChan, params.stderr, params.cfg.SuggestGenerics)
	}

	return withGenericDecls(
		buildSuffixTreeStandardPass(params, filesChan, typeInfos),
		golang.CollectGenericDecls(typeInfos),
	)
}

// buildSuffixTreeStandardPass parses filesChan into a suffix tree with the
// given type-aware data. Shared by single-mode and combined-mode orchestration.
func buildSuffixTreeStandardPass(
	params buildParams,
	filesChan chan string,
	typeInfos golang.TypeAwareData,
) treeBuildResult {
	var (
		schan     chan []*syntax.Node
		statsChan chan job.ParseStats
	)

	if params.cfg.Workers != 1 {
		schan, statsChan = job.ParseParallel(
			params.ctx,
			filesChan,
			params.cfg.Workers,
			detectionMode(params.cfg),
			params.cfg.MaxChildrenSerial,
			typeInfos,
		)
	} else {
		schan, statsChan = job.Parse(
			params.ctx,
			filesChan,
			detectionMode(params.cfg),
			params.cfg.MaxChildrenSerial,
			typeInfos,
		)
	}

	return finalizeTreeBuild(params, schan, func() job.ParseStats {
		return <-statsChan
	})
}

// validatePaths checks that at least one given path exists on the filesystem.
// Returns a validation error when positional paths accompany --files (the file
// list must come from stdin alone).
// If some paths are invalid, they're silently skipped during crawl (with stderr warning).
// Only returns an error when NONE of the specified paths exist.
func validatePaths(paths []string, filesFromStdin bool) error {
	if filesFromStdin {
		if len(paths) > 0 {
			return duplerrors.NewValidationError(
				"--files reads paths from stdin; positional arguments cannot be combined with it: "+strings.Join(paths, ", "),
				nil,
			)
		}

		return nil
	}

	if len(paths) == 0 {
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
func verboseFprintf(w io.Writer, cfg *config.Config, msg string) {
	if cfg.Verbose {
		_, _ = fmt.Fprintf(w, "%s\n", msg)
	}
}

// setupFilter creates a filter based on config settings.
func setupFilter(stderr io.Writer, cfg *config.Config) (*gogenfilter.Filter, error) {
	type filterEntry struct {
		enabled bool
		option  gogenfilter.FilterOption
		logName string
	}

	genericLogName := "generic — any \"Code generated by\" comment"

	entries := []filterEntry{
		{!cfg.IncludeSQLC, gogenfilter.FilterSQLC, generatedCategorySQLC},
		{!cfg.IncludeTempl, gogenfilter.FilterTempl, generatedCategoryTempl + ", default"},
		{!cfg.IncludeProtobuf, gogenfilter.FilterProtobuf, generatedCategoryProtobuf},
		{!cfg.IncludeMockgen, gogenfilter.FilterMockgen, generatedCategoryMockgen},
		{!cfg.IncludeStringer, gogenfilter.FilterStringer, generatedCategoryStringer},
		{!cfg.IncludeGeneric, gogenfilter.FilterGeneric, genericLogName},
	}

	var filterOptions []gogenfilter.FilterOption

	for _, e := range entries {
		if e.enabled {
			filterOptions = append(filterOptions, e.option)
			verboseFprintf(stderr, cfg, fmt.Sprintf(
				"Auto-generated code filtering enabled (%s)", e.logName,
			))
		}
	}

	excludePatterns := buildExcludePatterns(stderr, cfg)

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

		verboseFprintf(stderr, cfg, "Auto-generated code filtering enabled")

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
func buildExcludePatterns(stderr io.Writer, cfg *config.Config) []string {
	patterns := make([]string, 0, len(cfg.ExcludePatterns)+len(cfg.IgnoreFiles)+1)
	patterns = append(patterns, cfg.ExcludePatterns...)
	patterns = append(patterns, cfg.IgnoreFiles...)

	if cfg.IgnoreTests {
		patterns = append(patterns, "*_test.go")

		verboseFprintf(stderr, cfg, "Test file exclusion enabled (--ignore-tests)")
	}

	return patterns
}

// startProfiling begins profiling if enabled in config, returning the profile result.
func startProfiling(w io.Writer, cfg *config.Config) job.ProfileResult {
	if !cfg.Profile {
		return job.ProfileResult{}
	}

	profile := job.StartProfile()
	_, _ = fmt.Fprintln(w, "📊 Performance profiling enabled")

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
	stderr io.Writer,
) (chan syntax.Match, job.ParseStats, *FilterStats, map[string]int, error) {
	err := validatePaths(paths, cfg.FilesFromStdin)
	if err != nil {
		return nil, job.ParseStats{}, nil, nil, duplerrors.WrapValidation(err, "path validation failed")
	}

	startProfile := startProfiling(stderr, cfg)

	filterParam, err := setupFilter(stderr, cfg)
	if err != nil {
		return nil, job.ParseStats{}, nil, nil, duplerrors.Wrap(
			err,
			duplerrors.AnalysisError,
			fmt.Sprintf("failed to setup filter (outputFormat: %s)", outputFormat),
		)
	}

	var filterStats *FilterStats = newTrackedFilterStats(filterParam, cfg)

	if cfg.DetectionMethods.IsHashOnly() {
		ch, ps, fs, err := executeHashOnlyAnalysis(ctx, cfg, paths, filterParam, filterStats, outputFormat, stderr)

		return ch, ps, fs, nil, err
	}

	if isCombinedTypeAndGenericsMode(cfg) {
		return executeCombinedAnalysis(ctx, cfg, paths, filterParam, filterStats, outputFormat, stderr)
	}

	ingestStart := time.Now()

	result := buildSuffixTree(buildParams{
		ctx:          ctx,
		paths:        paths,
		cfg:          cfg,
		filterParam:  filterParam,
		filterStats:  filterStats,
		outputFormat: outputFormat,
		stderr:       stderr,
	})
	job.RecordStage(ctx, job.PhaseIngest, time.Since(ingestStart))

	if result.err != nil {
		return nil, job.ParseStats{}, nil, nil, duplerrors.Wrap(
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
		Methods:       cfg.DetectionMethods,
		Verbose:       cfg.Verbose,
		SearchWorkers: cfg.SearchWorkers,
	}, result.data, result.tree)

	duplChan := spawnCloneDetection(ctx, multiDetector, cfg.Threshold)

	endProfiling(cfg, startProfile)

	return duplChan, result.parseStats, filterStats, result.genericDecls, nil
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

		searchStart := time.Now()

		matches := detector.FindDuplOver(ctx, threshold)
		for match := range matches {
			select {
			case duplChan <- match:
			case <-ctx.Done():
				return
			}
		}

		job.RecordStage(ctx, job.PhaseSearch, time.Since(searchStart))
	}()

	return duplChan
}

// printCacheStats reports cache effectiveness to stderr in verbose mode.
// Without this, hits/misses/mem-hits are tracked but never surfaced.
func printCacheStats(stderr io.Writer, cfg *config.Config, stats cache.Stats) {
	if !cfg.Verbose {
		return
	}

	_, _ = fmt.Fprintf(
		stderr,
		"📦 Cache: %d hits (%d in memory), %d misses, %d entries on disk\n",
		stats.Hits, stats.MemHits, stats.Misses, stats.Size,
	)
}
