package provider

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/gogenfilter/v3"
)

// excludedDirs are directory names never analyzed, mirroring the CLI's
// default exclusions (vendor/node_modules/examples/demo/demos); dot-
// directories (.git, .direnv, ...) are skipped wholesale.
var excludedDirs = map[string]struct{}{ //nolint:gochecknoglobals // fixed exclusion set, read-only
	"vendor":       {},
	"node_modules": {},
	"examples":     {},
	"demo":         {},
	"demos":        {},
}

// collectSourceFiles walks root and returns every analyzable source file:
// .go and .templ files outside excluded directories, with generated Go files
// removed via gogenfilter (all categories, content-based detection included).
// A file whose generated-filter check fails I/O-wise (e.g. vanished mid-walk)
// is skipped rather than failing the whole run.
func collectSourceFiles(root string) ([]string, error) {
	filterConfig, err := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
	if err != nil {
		return nil, fmt.Errorf("configure generated-code filter: %w", err)
	}

	filter, err := gogenfilter.NewFilter(filterConfig)
	if err != nil {
		return nil, fmt.Errorf("create generated-code filter: %w", err)
	}

	files := make([]string, 0, 64)

	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if _, excluded := excludedDirs[entry.Name()]; excluded || strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}

			return nil
		}

		if !isSourceFile(path) {
			return nil
		}

		if filepath.Ext(path) == ".templ" {
			files = append(files, path)

			return nil
		}

		filtered, err := filter.Filter(path)
		if err != nil || filtered {
			return nil
		}

		files = append(files, path)

		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", root, walkErr)
	}

	return files, nil
}

// isSourceFile reports whether path has a supported source extension.
// Templ's generated output (*_templ.go) is handled by gogenfilter's filename
// phase, not here — one authority for generated-file detection.
func isSourceFile(path string) bool {
	switch filepath.Ext(path) {
	case ".go", ".templ":
		return true
	default:
		return false
	}
}
