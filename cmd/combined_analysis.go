package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// isCombinedTypeAndGenericsMode reports whether both --type-aware and
// --suggest-generics are set. Instead of one flag silently discarding the
// other, the combined mode runs two detection passes over one shared
// go/packages type-check (the expensive part): a type-aware pass and a
// type-erased pass, then merges their match streams. See ADR-0026.
func isCombinedTypeAndGenericsMode(cfg *config.Config) bool {
	return cfg.TypeAware && cfg.SuggestGenerics
}

// executeCombinedAnalysis runs the two-pass combined mode:
//
// Pass 1 (type-aware, types encoded in the hash) produces exact type-matched
// clone groups — the same output as --type-aware alone.
//
// Pass 2 (type-erased hashing) matches structurally identical clones across
// different concrete types. Only families with at least
// syntax.MinDivergentPositions divergent type positions flow through:
// zero-divergence families are exact duplicates of pass-1 groups, and
// one-position divergence is the shallow receiver noise class --type-aware
// exists to eliminate. Surviving families are generics-extraction candidates;
// actionability and line gates apply downstream exactly as in single modes.
//
// Filtering, crawling, and type-checking happen once; only the cheap
// transform + suffix-tree passes run twice.
func executeCombinedAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	filterParam *gogenfilter.Filter,
	filterStats *FilterStats,
	outputFormat config.OutputFormat,
	stderr io.Writer,
) (chan syntax.Match, job.ParseStats, *FilterStats, error) {
	params := buildParams{
		ctx:          ctx,
		paths:        paths,
		cfg:          cfg,
		filterParam:  filterParam,
		filterStats:  filterStats,
		outputFormat: outputFormat,
		stderr:       stderr,
	}

	// Crawl and filter once; both passes replay the same file list.
	filesChan := params.getFilesChan()
	filesChan = progressFilesChan(ctx, filesChan, cfg, outputFormat, stderr)
	allFiles := collectFiles(ctx, filesChan)

	if err := ctx.Err(); err != nil {
		return nil, job.ParseStats{}, nil, err
	}

	typeAwareData, erasedData := loadCombinedTypeData(stderr, allFiles)

	resultTa := runCombinedPass(params, allFiles, typeAwareData, passStatus{
		verbose: "Building suffix tree (type-aware pass)",
		text:    "    📖 Parsing files and building analysis tree...",
	})

	resultSg := runCombinedPass(params, allFiles, erasedData, passStatus{
		verbose: "Building suffix tree (type-erased generics pass)",
		text:    "    🧬 Parsing with type-erased hashing...",
	})

	for _, result := range []treeBuildResult{resultTa, resultSg} {
		if result.err != nil {
			return nil, job.ParseStats{}, nil, duplerrors.Wrap(
				result.err,
				duplerrors.AnalysisError,
				fmt.Sprintf(
					"failed to build suffix tree for paths %v (combined type-aware + suggest-generics mode)",
					paths,
				),
			)
		}
	}

	taChan := spawnCloneDetection(ctx, combinedDetector(cfg, resultTa), cfg.Threshold)
	sgChan := spawnCloneDetection(ctx, combinedDetector(cfg, resultSg), cfg.Threshold)

	merged := mergeCombinedMatches(ctx, taChan, sgChan)

	return merged, resultTa.parseStats, filterStats, nil
}

// combinedDetector builds the detector for one combined-mode pass.
func combinedDetector(cfg *config.Config, result treeBuildResult) *detection.MultiDetector {
	return detection.NewMultiDetector(detection.Config{
		Methods:       cfg.DetectionMethods,
		Verbose:       cfg.Verbose,
		SearchWorkers: cfg.SearchWorkers,
	}, result.data, result.tree)
}

// passStatus carries the per-pass status messages so the combined mode can
// distinguish its two tree builds in the progress output.
type passStatus struct {
	verbose string
	text    string
}

// runCombinedPass runs one parse + tree-build pass over a replay of the
// crawled file list with the given type data disposition.
func runCombinedPass(
	params buildParams,
	allFiles []string,
	typeInfos golang.TypeAwareData,
	status passStatus,
) treeBuildResult {
	printBuildingStatus(params.stderr, params.cfg, params.outputFormat, status.verbose, status.text)

	if params.cfg.Incremental {
		incParser := job.NewIncrementalParser(
			params.cfg.CacheDir,
			params.cfg.ClearCache,
			detectionMode(params.cfg),
			params.cfg.MaxChildrenSerial,
			params.cfg.MaxCacheEntries,
			params.cfg.MemoryCacheEntries,
		)
		incParser.SetTypeAwareData(typeInfos)

		return buildSuffixTreeIncrementalPass(params, incParser, replayFiles(allFiles), true)
	}

	return buildSuffixTreeStandardPass(params, replayFiles(allFiles), typeInfos)
}

// mergeCombinedMatches forwards type-aware matches unchanged, then generics-pass
// matches that qualify structurally as generics candidates (at least
// syntax.MinDivergentPositions divergent type positions). Dropped generics-pass
// families are either exact content duplicates of type-aware groups (zero
// divergence) or shallow receiver noise (one divergence). Presentation gates —
// actionability patterns, line minimums, accept directives — apply downstream,
// identically for both passes.
func mergeCombinedMatches(
	ctx context.Context,
	typeAwareChan, genericsChan <-chan syntax.Match,
) chan syntax.Match {
	merged := make(chan syntax.Match)

	go func() {
		defer close(merged)

		for match := range typeAwareChan {
			if !sendMatch(ctx, merged, match) {
				return
			}
		}

		for match := range genericsChan {
			if !syntax.IsGenericsCandidateStructure(match.Frags) {
				continue
			}

			if !sendMatch(ctx, merged, match) {
				return
			}
		}
	}()

	return merged
}

// sendMatch sends one match honoring context cancellation.
func sendMatch(ctx context.Context, ch chan<- syntax.Match, match syntax.Match) bool {
	select {
	case ch <- match:
		return true
	case <-ctx.Done():
		return false
	}
}

// loadCombinedTypeData loads type information once and derives both hash
// dispositions from the single go/packages load. Falls back to syntax-only
// (nil, nil) on load failure, in which case both passes produce identical
// syntax-only results and the generics-pass filter drops everything.
func loadCombinedTypeData(
	stderr io.Writer,
	allFiles []string,
) (typeAware, erased golang.TypeAwareData) {
	goFiles := make([]string, 0, len(allFiles))

	for _, f := range allFiles {
		if filepath.Ext(f) == ".go" {
			goFiles = append(goFiles, f)
		}
	}

	if len(goFiles) == 0 {
		return nil, nil
	}

	fmt.Fprintf(stderr, "🔍 Combined mode: loading type information for %d Go files...\n", len(goFiles))

	typeData, err := golang.LoadTypeAwareData(goFiles, false)
	if err != nil {
		logger.Default.Error("combined mode type loading failed, falling back to syntax-only", "err", err)

		return nil, nil
	}

	return typeData, typeData.WithEraseHash(true)
}
