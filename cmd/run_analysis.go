package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// buildSuffixTree builds a suffix tree from provided paths.
func buildSuffixTree(ctx context.Context, paths []string, cfg *config.Config, filterParam *filter.Filter, outputFormat config.OutputFormat) (*suffixtree.STree, []*syntax.Node, job.ParseStats, error) {
	if cfg.Verbose {
		fmt.Fprintln(os.Stderr, "Building suffix tree")
	} else if outputFormat == config.OutputFormatText {
		fmt.Fprint(os.Stderr, "    📖 Parsing files and building analysis tree...")
	}

	var schan chan []*syntax.Node
	var parseStats job.ParseStats

	// Debug output for incremental mode detection
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "🔍 buildSuffixTree: incremental=%v, cacheDir=%q\n", cfg.Incremental, cfg.CacheDir)
	}

	if cfg.Incremental {
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Incremental mode enabled, cache dir: %s\n", cfg.CacheDir)
		}
		incParser := job.NewIncrementalParser(cfg.CacheDir, cfg.ClearCache)
		var incStatsChan chan job.IncrementalStats
		schan, incStatsChan = incParser.ParseIncremental(ctx, filesFeedWithOptions(paths, cfg.FilesFromStdin, filterParam, cfg.IncludeVendor))
		t, data, done := job.BuildTree(ctx, schan)
		<-done
		incStats := <-incStatsChan
		parseStats = job.ParseStats{FilesCount: incStats.FilesCount, LinesCount: incStats.LinesCount}
		t.Update(&syntax.Node{Type: -1})

		if cfg.Verbose {
			fmt.Fprintln(os.Stderr, "Searching for clones")
		} else if outputFormat == config.OutputFormatText {
			fmt.Fprintln(os.Stderr, " ✅")
		}

		return t, *data, parseStats, nil
	}

	// Standard parsing without cache
	var statsChan chan job.ParseStats
	schan, statsChan = job.Parse(ctx, filesFeedWithOptions(paths, cfg.FilesFromStdin, filterParam, cfg.IncludeVendor))
	t, data, done := job.BuildTree(ctx, schan)
	<-done
	parseStats = <-statsChan
	t.Update(&syntax.Node{Type: -1})

	if cfg.Verbose {
		fmt.Fprintln(os.Stderr, "Searching for clones")
	} else if outputFormat == config.OutputFormatText {
		fmt.Fprintln(os.Stderr, " ✅")
	}

	return t, *data, parseStats, nil
}

// setupFilter creates a filter based on config settings.
func setupFilter(cfg *config.Config) *filter.Filter {
	var filterOptions []filter.FilterOption

	// Filter sqlc files by default (filename-based detection is very fast)
	// User can opt-out with --include-sqlc
	if !cfg.IncludeSQLC {
		filterOptions = append(filterOptions, filter.FilterSQLC)
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (sqlc)\n")
		}
	}

	// Filter templ files by default (filename-based detection is very fast)
	// User can opt-out with --include-templ
	if !cfg.IncludeTempl {
		filterOptions = append(filterOptions, filter.FilterTempl)
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (templ)\n")
		}
	}

	// Create the filter if there are any options or include/exclude patterns
	if len(filterOptions) > 0 || len(cfg.IncludePatterns) > 0 || len(cfg.ExcludePatterns) > 0 || len(cfg.IgnoreFiles) > 0 {
		filterParam := filter.NewFilter(true, filterOptions)
		filterParam.WithIncludePatterns(cfg.IncludePatterns)
		// IgnoreFiles are treated as exclude patterns
		filterParam.WithExcludePatterns(append(cfg.ExcludePatterns, cfg.IgnoreFiles...))

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "🔍 Auto-generated code filtering enabled (templ files filtered by default)\n")
		}
		return filterParam
	}
	return nil
}

// executeAnalysis runs the core duplicate analysis logic.
func executeAnalysis(ctx context.Context, cfg *config.Config, paths []string, outputFormat config.OutputFormat) (chan syntax.Match, job.ParseStats, filter.FilterStats, error) {
	var startProfile job.ProfileResult
	if cfg.Profile {
		startProfile = job.StartProfile()
		fmt.Fprintln(os.Stderr, "📊 Performance profiling enabled")
	}

	// Create filter based on config
	filterParam := setupFilter(cfg)

	t, data, parseStats, err := buildSuffixTree(ctx, paths, cfg, filterParam, outputFormat)
	if err != nil {
		return nil, job.ParseStats{}, filter.FilterStats{}, duplerrors.Wrap(err, duplerrors.AnalysisError, fmt.Sprintf("failed to build suffix tree for paths %v", paths))
	}

	// Get filter statistics if filter was enabled
	var filterStats filter.FilterStats
	if filterParam != nil {
		filterStats = filterParam.GetStats()
	}

	multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
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

	return duplChan, parseStats, filterStats, nil
}

// createPrinter returns the appropriate printer based on output format.
func createPrinter(outputFormat config.OutputFormat, threshold int) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return func(w io.Writer, fread printer.ReadFile) printer.Printer {
			return printer.NewHTML(w, fread, threshold)
		}
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatSimpleJSON:
		return printer.NewJSON
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}
