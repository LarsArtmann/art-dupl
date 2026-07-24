package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// loadTypeAwareData drains the files channel, collects .go file paths, and
// loads type-checking data via go/packages. Returns the type-aware data map and
// a new channel that replays all collected files (both .go and .templ) for the
// subsequent parsing phase.
//
// This is a blocking operation — the entire file set must be known before
// go/packages can resolve imports and type-check.
func loadTypeAwareData(
	ctx context.Context,
	filesChan chan string,
) (golang.TypeAwareData, chan string) {
	var allFiles []string

	for file := range filesChan {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		allFiles = append(allFiles, file)
	}

	goFiles := make([]string, 0, len(allFiles))

	for _, f := range allFiles {
		if filepath.Ext(f) == ".go" {
			goFiles = append(goFiles, f)
		}
	}

	if len(goFiles) == 0 {
		newChan := make(chan string, len(allFiles))

		for _, f := range allFiles {
			newChan <- f
		}

		close(newChan)

		return nil, newChan
	}

	fmt.Fprintf(os.Stderr, "🔍 Type-aware mode: loading type information for %d Go files...\n", len(goFiles))

	typeData, err := golang.LoadTypeAwareData(goFiles)
	if err != nil {
		logger.Default.Error("type-aware mode failed, falling back to syntax-only", "err", err)
		typeData = nil
	}

	newChan := make(chan string, len(allFiles))

	for _, f := range allFiles {
		newChan <- f
	}

	close(newChan)

	return typeData, newChan
}
