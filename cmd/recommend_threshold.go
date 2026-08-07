package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/config"
)

// Threshold recommendation boundaries based on codebase size.
const (
	smallCodebaseFiles  = 100  // <100 files → threshold 3
	mediumCodebaseFiles = 1000 // 100-1000 → threshold 5
	largeCodebaseFiles  = 5000 // 1000-5000 → threshold 7

	smallCodebaseThreshold  = 3
	mediumCodebaseThreshold = 5
	largeCodebaseThreshold  = 7
	hugeCodebaseThreshold   = 10
)

// RecommendThreshold suggests a threshold based on codebase size.
// Heuristic: <100 files -> 3, 100-1000 -> 5, 1000-5000 -> 7, >5000 -> 10.
func RecommendThreshold(fileCount int) int {
	switch {
	case fileCount < smallCodebaseFiles:
		return smallCodebaseThreshold
	case fileCount < mediumCodebaseFiles:
		return mediumCodebaseThreshold
	case fileCount < largeCodebaseFiles:
		return largeCodebaseThreshold
	default:
		return hugeCodebaseThreshold
	}
}

// runRecommendThreshold counts Go files in the configured paths and prints
// a threshold recommendation. Exits 0 without running detection.
func runRecommendThreshold(ctx context.Context, cfg *config.Config) error {
	fileCount := 0

	for _, path := range cfg.Paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if isGoFile(path) {
				fileCount++
			}

			continue
		}

		_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil //nolint:nilerr // skip unreadable entries, don't abort the walk
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if !d.IsDir() && isGoFile(p) {
				fileCount++
			}

			return nil
		})

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	recommended := RecommendThreshold(fileCount)

	// Deep audit always runs at threshold 1 to surface every potential clone;
	// the CI gate uses the size-based recommendation to filter noise.
	auditThreshold := 1

	output := fmt.Sprintf(
		"Codebase: %d Go files\n\n"+
			"CI gate:       art-dupl -t %d .\n"+
			"Deep audit:    art-dupl -t %d --explain .\n\n"+
			"  CI gate filters noise (test boilerplate, guard clauses, etc.)\n"+
			"  Deep audit finds everything; pair with --explain to triage manually.\n",
		fileCount, recommended, auditThreshold,
	)

	if _, err := os.Stdout.WriteString(output); err != nil {
		return fmt.Errorf("write threshold recommendation: %w", err)
	}

	return nil
}

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}
