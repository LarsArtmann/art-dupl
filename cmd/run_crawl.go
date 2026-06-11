package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// Directory exclusion constants.
const (
	// VendorDirPrefix is the vendor directory prefix for exclusion.
	VendorDirPrefix = "vendor" + string(filepath.Separator)
	// VendorDirInPath is the vendor directory marker when it appears in a path.
	VendorDirInPath = string(filepath.Separator) + VendorDirPrefix
	// DSStoreFile is the macOS Finder metadata file that should always be excluded.
	DSStoreFile = ".DS_Store"
	// GitDirPrefix is the Git directory prefix for exclusion.
	GitDirPrefix = ".git" + string(filepath.Separator)
	// GitDirInPath is the Git directory marker when it appears in a path.
	GitDirInPath = string(filepath.Separator) + GitDirPrefix
	// NodeModulesDirPrefix is the node_modules directory prefix for exclusion.
	NodeModulesDirPrefix = "node_modules" + string(filepath.Separator)
	// NodeModulesDirInPath is the node_modules directory marker when it appears in a path.
	NodeModulesDirInPath = string(filepath.Separator) + NodeModulesDirPrefix
)

// crawlSinglePathWithOpts handles crawling of a single path using CrawlOptions.
func filesFeedWithOptions(
	paths []string,
	fromStdin bool,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includeVendor, includeNodeModules bool,
	only config.FileType,
) chan string {
	if fromStdin {
		fchan := make(chan string)

		go func() {
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				f := sc.Text()
				path := strings.TrimPrefix(f, "./")

				if !shouldIncludeFile(filter, path, filterStats) {
					continue
				}

				if !only.Matches(path) {
					continue
				}

				fchan <- path
			}

			err := sc.Err()
			if err != nil {
				fmt.Fprintf(os.Stderr, "reading stdin: %v\n", err)
			}

			close(fchan)
		}()

		return fchan
	}

	fileCheck := func(name string) bool {
		if !isSourceFile(name) {
			return false
		}

		return only.Matches(name)
	}

	return crawlPathsWithFileCheck(
		paths, filter, filterStats,
		includeVendor, includeNodeModules, fileCheck,
	)
}

// crawlPaths walks paths and returns a channel of Go files.
func crawlPaths(
	paths []string,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includeVendor, includeNodeModules bool,
) chan string {
	return crawlPathsWithFileCheck(
		paths, filter, filterStats,
		includeVendor, includeNodeModules, isSourceFile,
	)
}

// crawlPathsAllFiles walks paths and returns a channel of all files (not just source files).
// This is used for hash-based detection which works on any file type.
// The includeNodeModules parameter controls whether to include node_modules directories
// (excluded by default to avoid processing large dependency directories).
// The only parameter filters to specific file extensions (".go" or ".templ").
func crawlPathsAllFiles(
	paths []string,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includeVendor, includeNodeModules bool,
	only config.FileType,
) chan string {
	var fileCheck fileCheckFunc
	if only != config.FileTypeAll {
		fileCheck = func(name string) bool {
			return only.Matches(name)
		}
	}

	return crawlPathsWithFileCheck(
		paths, filter, filterStats,
		includeVendor, includeNodeModules, fileCheck,
	)
}

// fileCheckFunc returns true if a file should be included based on its name.
// A nil function accepts all files.
type fileCheckFunc func(name string) bool

// CrawlOptions contains the common options for crawling operations.
type CrawlOptions struct {
	Filter          *gogenfilter.Filter
	FilterStats     *FilterStats
	IncludeVendor   bool
	IncludeNodeMods bool
	FileCheck       fileCheckFunc
	FChan           chan string
}

func crawlPathsWithFileCheck(
	paths []string,
	f *gogenfilter.Filter,
	filterStats *FilterStats,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			crawlSinglePathWithOpts(CrawlOptions{
				Filter:          f,
				FilterStats:     filterStats,
				IncludeVendor:   includeVendor,
				IncludeNodeMods: includeNodeModules,
				FileCheck:       fileCheck,
				FChan:           fchan,
			}, path)
		}

		close(fchan)
	}()

	return fchan
}

// crawlSinglePathWithOpts handles crawling of a single path using CrawlOptions.
func crawlSinglePathWithOpts(opts CrawlOptions, path string) {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot stat %s: %v\n", path, err)

		return
	}

	if !info.IsDir() {
		if shouldIncludeFile(opts.Filter, path, opts.FilterStats) &&
			passesFileCheck(info.Name(), opts.FileCheck) {
			opts.FChan <- path
		}

		return
	}

	crawlDirectoryWithOpts(opts, path)
}

// crawlDirectoryWithOpts walks a directory tree using CrawlOptions.
func crawlDirectoryWithOpts(opts CrawlOptions, path string) {
	err := filepath.Walk(path, func(p string, info os.FileInfo, _ error) error {
		return handleWalkEntry(opts, p, info)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: walking %s: %v\n", path, err)
	}
}

// handleWalkEntry processes a single entry during directory walk.
func handleWalkEntry(opts CrawlOptions, path string, info os.FileInfo) error {
	if info == nil {
		return nil
	}

	if shouldSkipPath(path, opts.IncludeVendor, opts.IncludeNodeMods) {
		return nil
	}

	if info.Name() == DSStoreFile {
		return nil
	}

	if !info.IsDir() && passesFileCheck(info.Name(), opts.FileCheck) &&
		shouldIncludeFile(opts.Filter, path, opts.FilterStats) {
		opts.FChan <- path
	}

	return nil
}

// passesFileCheck returns true if the file passes the optional file check.
func passesFileCheck(name string, fileCheck fileCheckFunc) bool {
	return fileCheck == nil || fileCheck(name)
}

// isSourceFile returns true if the filename has a supported source file extension.
func isSourceFile(name string) bool {
	return strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".templ")
}

// shouldSkipPath returns true if the path should be skipped due to being a vendor, git, or node_modules directory.
func shouldSkipPath(path string, includeVendor, includeNodeModules bool) bool {
	if !includeVendor && (strings.HasPrefix(path, VendorDirPrefix) ||
		strings.Contains(path, VendorDirInPath)) {
		return true
	}

	// Skip .git directories
	if strings.HasPrefix(path, GitDirPrefix) ||
		strings.Contains(path, GitDirInPath) {
		return true
	}

	// Skip node_modules directories (configurable, excluded by default in hash detection)
	// This prevents processing large dependency directories in hash detection mode
	if !includeNodeModules && (strings.HasPrefix(path, NodeModulesDirPrefix) ||
		strings.Contains(path, NodeModulesDirInPath)) {
		return true
	}

	return false
}
