package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/gogenfilter"
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
	outputFormat config.OutputFormat
}

// getFilesChan creates a channel of file paths based on the build parameters.
func (p buildParams) getFilesChan() chan string {
	return filesFeedWithOptions(
		p.paths,
		p.cfg.FilesFromStdin,
		p.filterParam,
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
		params.cfg.Semantic,
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
			params.cfg.Semantic,
		)
	} else {
		schan, statsChan = job.Parse(params.ctx, filesChan, params.cfg.Semantic)
	}

	tree, data, done := job.BuildTree(params.ctx, schan)
	<-done

	parseStats := <-statsChan

	tree.Update(&syntax.Node{Type: -1})
	printSearchStatus(params.cfg, params.outputFormat)

	return treeBuildResult{tree: tree, data: *data, parseStats: parseStats}
}

// verboseFprintf prints a message to stderr if verbose mode is enabled.
func verboseFprintf(cfg *config.Config, msg string) {
	if cfg.Verbose {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}

// setupFilter creates a filter based on config settings.
func setupFilter(cfg *config.Config) (*gogenfilter.Filter, error) {
	var filterOptions []gogenfilter.FilterOption

	if !cfg.IncludeSQLC {
		filterOptions = append(filterOptions, gogenfilter.FilterSQLC)

		verboseFprintf(cfg, "Auto-generated code filtering enabled (sqlc)")
	}

	if !cfg.IncludeTempl {
		filterOptions = append(filterOptions, gogenfilter.FilterTempl)

		verboseFprintf(cfg, "Auto-generated code filtering enabled (templ, default)")
	}

	if !cfg.IncludeProtobuf {
		filterOptions = append(filterOptions, gogenfilter.FilterProtobuf)

		verboseFprintf(cfg, "Auto-generated code filtering enabled (protobuf)")
	}

	if !cfg.IncludeMockgen {
		filterOptions = append(filterOptions, gogenfilter.FilterMockgen)

		verboseFprintf(cfg, "Auto-generated code filtering enabled (mockgen)")
	}

	if !cfg.IncludeStringer {
		filterOptions = append(filterOptions, gogenfilter.FilterStringer)

		verboseFprintf(cfg, "Auto-generated code filtering enabled (stringer)")
	}

	if len(filterOptions) > 0 || len(cfg.IncludePatterns) > 0 || len(cfg.ExcludePatterns) > 0 ||
		len(cfg.IgnoreFiles) > 0 {
		configs := []gogenfilter.FilterConfig{
			gogenfilter.WithIncludePatterns(cfg.IncludePatterns...),
			gogenfilter.WithExcludePatterns(append(cfg.ExcludePatterns, cfg.IgnoreFiles...)...),
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

// executeAnalysis runs the core duplicate analysis logic.
func executeAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	outputFormat config.OutputFormat,
) (chan syntax.Match, job.ParseStats, gogenfilter.FilterStats, error) {
	var startProfile job.ProfileResult
	if cfg.Profile {
		startProfile = job.StartProfile()

		_, _ = fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")
	}

	filterParam, err := setupFilter(cfg)
	if err != nil {
		return nil, job.ParseStats{}, gogenfilter.FilterStats{}, duplerrors.Wrap(
			err,
			duplerrors.AnalysisError,
			"failed to setup filter",
		)
	}

	var filterStats gogenfilter.FilterStats
	if filterParam != nil {
		filterStats = filterParam.GetStats()
	}

	if cfg.DetectionMethods.IsHashOnly() {
		return executeHashOnlyAnalysis(ctx, cfg, paths, filterParam, outputFormat)
	}

	result := buildSuffixTree(buildParams{
		ctx:          ctx,
		paths:        paths,
		cfg:          cfg,
		filterParam:  filterParam,
		outputFormat: outputFormat,
	})
	if result.err != nil {
		return nil, job.ParseStats{}, gogenfilter.FilterStats{}, duplerrors.Wrap(
			result.err,
			duplerrors.AnalysisError,
			fmt.Sprintf("failed to build suffix tree for paths %v", paths),
		)
	}

	multiDetector := detection.NewMultiDetector(config.DetectionConfig{
		Methods: cfg.DetectionMethods,
		Verbose: cfg.Verbose,
	}, result.data, result.tree)
	duplChan := make(chan syntax.Match)

	go func() {
		defer close(duplChan)

		matches := multiDetector.FindDuplOver(cfg.Threshold)
		for match := range matches {
			select {
			case <-ctx.Done():
				return
			default:
			}

			duplChan <- match
		}
	}()

	if cfg.Profile {
		endProfile := job.EndProfile(startProfile)
		job.PrintProfileResult(endProfile)
	}

	return duplChan, result.parseStats, filterStats, nil
}
