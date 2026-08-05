package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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

// feedFromStdin reads newline-delimited file paths from rc, applies filters,
// and sends matching paths on the returned channel. When ctx is cancelled, the
// reader is closed to unblock the inherently blocking bufio.Scanner.Scan call.
// The channel is closed when scanning completes (EOF or cancellation).
func feedFromStdin(
	ctx context.Context,
	rc io.ReadCloser,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includes generatorIncludes,
	only config.FileType,
	stderr io.Writer,
) chan string {
	fchan := make(chan string)

	go func() {
		defer close(fchan)

		done := make(chan struct{})

		go func() {
			select {
			case <-ctx.Done():
				_ = rc.Close()
			case <-done:
			}
		}()

		sc := bufio.NewScanner(rc)
		for sc.Scan() {
			f := sc.Text()
			path := strings.TrimPrefix(f, "./")

			if !shouldIncludeFile(filter, path, filterStats, includes) {
				continue
			}

			if !only.Matches(path) {
				continue
			}

			select {
			case fchan <- path:
			case <-ctx.Done():
				return
			}
		}

		close(done)

		if err := sc.Err(); err != nil && ctx.Err() == nil {
			fmt.Fprintf(stderr, "reading stdin: %v\n", err)
		}
	}()

	return fchan
}

// crawlSinglePathWithOpts handles crawling of a single path using CrawlOptions.
func filesFeedWithOptions(
	ctx context.Context,
	paths []string,
	fromStdin bool,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includes generatorIncludes,
	includeVendor, includeNodeModules bool,
	only config.FileType,
	gitignore *GitignoreMatcher,
) chan string {
	if fromStdin {
		return feedFromStdin(ctx, os.Stdin, filter, filterStats, includes, only)
	}

	fileCheck := func(name string) bool {
		if !isSourceFile(name) {
			return false
		}

		return only.Matches(name)
	}

	return crawlPathsWithFileCheck(
		ctx,
		paths, filter, filterStats, includes,
		includeVendor, includeNodeModules, fileCheck,
		gitignore,
	)
}

// crawlPaths walks paths and returns a channel of Go files.
func crawlPaths(
	ctx context.Context,
	paths []string,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includes generatorIncludes,
	includeVendor, includeNodeModules bool,
	gitignore *GitignoreMatcher,
	stderr io.Writer,
) chan string {
	return crawlPathsWithFileCheck(
		ctx,
		paths, filter, filterStats, includes,
		includeVendor, includeNodeModules, isSourceFile,
		gitignore,
		stderr,
	)
}

// crawlPathsAllFiles walks paths and returns a channel of all files (not just source files).
// This is used for hash-based detection which works on any file type.
// The includeNodeModules parameter controls whether to include node_modules directories
// (excluded by default to avoid processing large dependency directories).
// The only parameter filters to specific file extensions (".go" or ".templ").
func crawlPathsAllFiles(
	ctx context.Context,
	paths []string,
	filter *gogenfilter.Filter,
	filterStats *FilterStats,
	includes generatorIncludes,
	includeVendor, includeNodeModules bool,
	only config.FileType,
	gitignore *GitignoreMatcher,
	stderr io.Writer,
) chan string {
	var fileCheck fileCheckFunc
	if only != config.FileTypeAll {
		fileCheck = func(name string) bool {
			return only.Matches(name)
		}
	}

	return crawlPathsWithFileCheck(
		ctx,
		paths, filter, filterStats, includes,
		includeVendor, includeNodeModules, fileCheck,
		gitignore,
		stderr,
	)
}

// fileCheckFunc returns true if a file should be included based on its name.
// A nil function accepts all files.
type fileCheckFunc func(name string) bool

// CrawlOptions contains the common options for crawling operations.
type CrawlOptions struct {
	Ctx             context.Context
	Filter          *gogenfilter.Filter
	FilterStats     *FilterStats
	Includes        generatorIncludes
	IncludeVendor   bool
	IncludeNodeMods bool
	FileCheck       fileCheckFunc
	Gitignore       *GitignoreMatcher
	FChan           chan string
	Stderr          io.Writer
}

func crawlPathsWithFileCheck(
	ctx context.Context,
	paths []string,
	f *gogenfilter.Filter,
	filterStats *FilterStats,
	includes generatorIncludes,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
	gitignore *GitignoreMatcher,
	stderr io.Writer,
) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			if ctx.Err() != nil {
				break
			}

			crawlSinglePathWithOpts(CrawlOptions{
				Ctx:             ctx,
				Filter:          f,
				FilterStats:     filterStats,
				Includes:        includes,
				IncludeVendor:   includeVendor,
				IncludeNodeMods: includeNodeModules,
				FileCheck:       fileCheck,
				Gitignore:       gitignore,
				FChan:           fchan,
				Stderr:          stderr,
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
		fmt.Fprintf(opts.Stderr, "error: cannot stat %s: %v\n", path, err)

		return
	}

	if !info.IsDir() {
		if opts.Gitignore != nil && opts.Gitignore.IsIgnored(path) {
			return
		}

		if shouldIncludeFile(opts.Filter, path, opts.FilterStats, opts.Includes) &&
			passesFileCheck(info.Name(), opts.FileCheck) {
			opts.sendFile(path)
		}

		return
	}

	crawlDirectoryWithOpts(opts, path)
}

// sendFile pushes a file path onto the channel, respecting context cancellation.
func (opts CrawlOptions) sendFile(path string) {
	select {
	case opts.FChan <- path:
	case <-opts.Ctx.Done():
	}
}

// crawlDirectoryWithOpts walks a directory tree using CrawlOptions.
func crawlDirectoryWithOpts(opts CrawlOptions, path string) {
	err := filepath.Walk(path, func(p string, info os.FileInfo, _ error) error {
		return handleWalkEntry(opts, p, info)
	})
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		fmt.Fprintf(opts.Stderr, "error: walking %s: %v\n", path, err)
	}
}

// handleWalkEntry processes a single entry during directory walk.
func handleWalkEntry(opts CrawlOptions, path string, info os.FileInfo) error {
	if opts.Ctx.Err() != nil {
		return opts.Ctx.Err()
	}

	if info == nil {
		return nil
	}

	if shouldSkipPath(path, opts.IncludeVendor, opts.IncludeNodeMods) {
		return nil
	}

	if info.Name() == DSStoreFile {
		return nil
	}

	if opts.Gitignore != nil && opts.Gitignore.IsIgnored(path) {
		return nil
	}

	if !info.IsDir() && passesFileCheck(info.Name(), opts.FileCheck) &&
		shouldIncludeFile(opts.Filter, path, opts.FilterStats, opts.Includes) {
		opts.sendFile(path)
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
