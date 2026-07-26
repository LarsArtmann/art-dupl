package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/config"
)

// RecommendThreshold suggests a threshold based on codebase size.
// Heuristic: <100 files → 3, 100–1000 → 5, 1000–5000 → 7, >5000 → 10.
func RecommendThreshold(fileCount int) int {
	switch {
	case fileCount < 100:
		return 3
	case fileCount < 1000:
		return 5
	case fileCount < 5000:
		return 7
	default:
		return 10
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

		err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if !d.IsDir() && isGoFile(p) {
				fileCount++
			}

			return nil
		})
		if err != nil && ctx.Err() == nil {
			continue
		}
	}

	recommended := RecommendThreshold(fileCount)

	fmt.Fprintf(os.Stdout,
		"Codebase: %d Go files\n", fileCount)
	fmt.Fprintf(os.Stdout,
		"Recommended threshold: %d\n", recommended)
	fmt.Fprintf(os.Stdout,
		"  art-dupl -t %d .\n", recommended)

	if fileCount < 100 {
		fmt.Fprintln(os.Stdout,
			"  (small codebase: lower threshold catches more clones)")
	} else if fileCount >= 5000 {
		fmt.Fprintln(os.Stdout,
			"  (large codebase: higher threshold reduces noise)")
	}

	return nil
}

func isGoFile(path string) bool {
	ext := filepath.Ext(path)

	return ext == ".go"
}
