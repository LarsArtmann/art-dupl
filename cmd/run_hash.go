package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/hash"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// executeHashOnlyAnalysis runs hash-based duplicate detection without AST parsing.
// This is an optimized streaming path that hashes files one at a time via io.Copy
// through xxh3.Hasher — file content is never held in memory.
// Only (hash, filename, size) is retained per file, yielding O(1) memory per file
// regardless of file size.
func executeHashOnlyAnalysis(
	ctx context.Context,
	cfg *config.Config,
	paths []string,
	filterParam *gogenfilter.Filter,
	filterStats *FilterStats,
	outputFormat config.OutputFormat,
	stderr io.Writer,
) (chan syntax.Match, job.ParseStats, *FilterStats, error) {
	printBuildingStatus(
		stderr,
		cfg,
		outputFormat,
		"Running hash-only duplicate detection",
		"    📖 Hashing files for duplicate detection...",
	)

	if cfg.Verbose {
		_, _ = fmt.Fprintln(
			stderr,
			"🔍 Excluding node_modules/ directory (use --include-node-modules to include)",
		)
	}

	var gitignore *GitignoreMatcher
	if !cfg.IncludeIgnored {
		gitignore = LoadGitignore(paths, stderr)
	}

	filesChan := crawlPathsAllFiles(
		ctx,
		paths,
		filterParam,
		filterStats,
		newGeneratorIncludes(cfg),
		cfg.IncludeVendor,
		cfg.IncludeNodeModules,
		cfg.Only,
		gitignore,
		stderr,
	)

	filesChan = progressFilesChan(ctx, filesChan, cfg, outputFormat, stderr)

	files, err := collectFilesFromChannel(ctx, filesChan)
	if err != nil {
		return nil, job.ParseStats{}, nil, fmt.Errorf(
			"hash-only analysis (outputFormat: %s): %w", outputFormat, err,
		)
	}

	printFileCollectionStatus(stderr, cfg, outputFormat, len(files))

	fileDuplicates := hash.FindFileDuplicates(ctx, files, cfg.Threshold)

	duplChan := convertFileDuplicatesToMatches(ctx, fileDuplicates)

	return duplChan, job.ParseStats{
		ParseStatsMixin: job.ParseStatsMixin{FilesCount: len(files), LinesCount: 0},
	}, filterStats, nil
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
			select {
			case duplChan <- match:
			case <-ctx.Done():
				return
			}
		}
	}()

	return duplChan
}

// createFragmentsFromFileHashes converts file hashes to syntax.Node fragments.
func createFragmentsFromFileHashes(files []hash.FileHash) [][]*syntax.Node {
	fragments := make([][]*syntax.Node, 0, len(files))
	for _, fileHash := range files {
		node := syntax.NewSyntheticFileNode(fileHash.Filename, fileHash.Size)
		fragments = append(fragments, []*syntax.Node{node})
	}

	return fragments
}

// printFileCollectionStatus outputs status after file collection.
func printFileCollectionStatus(
	w io.Writer,
	cfg *config.Config,
	outputFormat config.OutputFormat,
	fileCount int,
) {
	if cfg.Verbose {
		fmt.Fprintf(w, "Found %d files to hash\n", fileCount)
	} else if outputFormat == config.OutputFormatText {
		fmt.Fprintln(w, " ✅")
	}
}
