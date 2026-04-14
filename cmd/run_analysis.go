package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/hash"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

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
	verboseMsg, textMsg string,
) {
	if cfg.Verbose {
		_, _ = fmt.Fprintln(os.Stderr, verboseMsg)
	} else if outputFormat == config.OutputFormatText {
		_, _ = fmt.Fprint(os.Stderr, textMsg)
	}
}

// buildParams holds common parameters for building a suffix tree.
type buildParams struct {
	ctx          context.Context
	paths        []string
	cfg          *config.Config
	filterParam  *filter.Filter
	outputFormat config.OutputFormat
}

// treeBuildResult holds the result of building a suffix tree.
type treeBuildResult struct {
	tree       *suffixtree.STree
	data       []*syntax.Node
	parseStats job.ParseStats
	err        error
}

// getFilesChan creates a channel of file paths based on the build parameters.
func (p buildParams) getFilesChan() chan string {
	return filesFeedWithOptions(
		p.paths,
		p.cfg.FilesFromStdin,
		p.filterParam,
		p.cfg.IncludeVendor,
		p.cfg.Only,
	)
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

// setupFilter creates a filter based on config settings.
func setupFilter(cfg *config.Config) *filter.Filter {
	var filterOptions []filter.FilterOption

	// Filter sqlc files by default (filename-based detection is very fast)
	// User can opt-out with --include-sqlc
	if !cfg.IncludeSQLC {
		filterOptions = append(filterOptions, filter.FilterSQLC)

		if cfg.Verbose {
			_, _ = fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (sqlc)\n")
		}
	}

	// Filter templ files by default (filename-based detection is very fast)
	// User can opt-out with --include-templ
	if !cfg.IncludeTempl {
		filterOptions = append(filterOptions, filter.FilterTempl)

		if cfg.Verbose {
			_, _ = fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (templ)\n")
		}
	}

	// Create the filter if there are any options or include/exclude patterns
	if len(filterOptions) > 0 || len(cfg.IncludePatterns) > 0 || len(cfg.ExcludePatterns) > 0 ||
		len(cfg.IgnoreFiles) > 0 {
		filterParam := filter.NewFilter(true, filterOptions)
		filterParam.WithIncludePatterns(cfg.IncludePatterns)
		// IgnoreFiles are treated as exclude patterns
		filterParam.WithExcludePatterns(append(cfg.ExcludePatterns, cfg.IgnoreFiles...))

		if cfg.Verbose {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"🔍 Auto-generated code filtering enabled (templ files filtered by default)\n",
			)
		}

		return filterParam
	}

	return nil
}

// executeAnalysis runs the core duplicate analysis logic.
func executeAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	outputFormat config.OutputFormat,
) (chan syntax.Match, job.ParseStats, filter.FilterStats, error) {
	var startProfile job.ProfileResult
	if cfg.Profile {
		startProfile = job.StartProfile()

		_, _ = fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")
	}

	// Create filter based on config
	filterParam := setupFilter(cfg)

	// Get filter statistics if filter was enabled
	var filterStats filter.FilterStats
	if filterParam != nil {
		filterStats = filterParam.GetStats()
	}

	// For hash-only detection, skip AST parsing and work directly with file paths
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
		return nil, job.ParseStats{}, filter.FilterStats{}, duplerrors.Wrap(
			result.err,
			duplerrors.AnalysisError,
			fmt.Sprintf("failed to build suffix tree for paths %v", paths),
		)
	}

	multiDetector := detection.NewMultiDetector(cfg, result.data, result.tree, cfg.Verbose)
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

// collectFilesFromChannel collects file paths from a channel into a slice,
// respecting context cancellation.
func collectFilesFromChannel(ctx context.Context, filesChan <-chan string) ([]string, error) {
	var files []string

	for file := range filesChan {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
			files = append(files, file)
		}
	}

	return files, nil
}

// convertFileDuplicatesToMatches converts hash.FileDuplicate slices to syntax.Match channel.
func convertFileDuplicatesToMatches(
	ctx context.Context,
	fileDuplicates []hash.FileDuplicate,
) chan syntax.Match {
	duplChan := make(chan syntax.Match)

	go func() {
		defer close(duplChan)

		for _, fileDup := range fileDuplicates {
			select {
			case <-ctx.Done():
				return
			default:
			}

			match := syntax.Match{
				Hash:  fileDup.Hash,
				Frags: createFragmentsFromFileHashes(fileDup.Files),
			}
			duplChan <- match
		}
	}()

	return duplChan
}

// createFragmentsFromFileHashes converts file hashes to syntax.Node fragments.
func createFragmentsFromFileHashes(files []hash.FileHash) [][]*syntax.Node {
	fragments := make([][]*syntax.Node, len(files))
	for i, fileHash := range files {
		node := syntax.NewSyntheticFileNode(fileHash.Filename, fileHash.Size)
		fragments[i] = []*syntax.Node{node}
	}

	return fragments
}

// executeHashOnlyAnalysis runs hash-based duplicate detection without AST parsing.
// This is an optimized streaming path that hashes files one at a time via io.Copy
// through xxh3.Hasher — file content is never held in memory.
// Only (hash, filename, size) is retained per file, yielding O(1) memory per file
// regardless of file size.
func executeHashOnlyAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	filterParam *filter.Filter,
	outputFormat config.OutputFormat,
) (chan syntax.Match, job.ParseStats, filter.FilterStats, error) {
	printBuildingStatus(
		cfg,
		outputFormat,
		"Running hash-only duplicate detection",
		"    📖 Hashing files for duplicate detection...",
	)

	if cfg.Verbose {
		_, _ = fmt.Fprintln(
			os.Stderr,
			"🔍 Excluding node_modules/ directory (use --include-node-modules to include)",
		)
	}

	filesChan := crawlPathsAllFiles(
		paths,
		filterParam,
		cfg.IncludeVendor,
		cfg.IncludeNodeModules,
		cfg.Only,
	)

	// Collect file paths (just strings — negligible memory) so we can report
	// file count before streaming results downstream.
	files, err := collectFilesFromChannel(ctx, filesChan)
	if err != nil {
		return nil, job.ParseStats{}, filter.FilterStats{}, err
	}

	printFileCollectionStatus(cfg, outputFormat, len(files))

	// Stream hash detection: files are hashed one at a time via io.Copy(xxh3, file).
	// No file content retained in memory — only (hash, filename, size) per file.
	fileDuplicates := hash.FindFileDuplicates(files, cfg.Threshold)

	duplChan := convertFileDuplicatesToMatches(ctx, fileDuplicates)

	var filterStats filter.FilterStats
	if filterParam != nil {
		filterStats = filterParam.GetStats()
	}

	return duplChan, job.ParseStats{
		ParseStatsMixin: job.ParseStatsMixin{FilesCount: len(files), LinesCount: 0},
	}, filterStats, nil
}

// printFileCollectionStatus outputs status after file collection.
func printFileCollectionStatus(
	cfg *config.Config,
	outputFormat config.OutputFormat,
	fileCount int,
) {
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d files to hash\n", fileCount)
	} else if outputFormat == config.OutputFormatText {
		fmt.Fprintln(os.Stderr, " ✅")
	}
}

// createPrinter returns the appropriate printer based on output format.
func createPrinter(
	outputFormat config.OutputFormat,
	threshold int,
	diffMode config.DiffMode,
	metadata printer.ReportMetadata,
) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewHTMLWithOptions(w, fread, diffMode, metadata, threshold)
		}
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatSimpleJSON:
		return printer.NewJSON
	case config.OutputFormatSARIF:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewSARIF(w, fread, threshold)
		}
	case config.OutputFormatCSV:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewStats(w, fread, threshold)
		}
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}
