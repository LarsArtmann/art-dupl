package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/filter"
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

// statError prints an error message for file stat failures and exits.
func statError(path string, err error) {
	fmt.Fprintf(os.Stderr, "error: cannot stat %s: %v\n", path, err)
	os.Exit(1)
}

// filesFeedWithOptions creates a channel of file paths with options.
func filesFeedWithOptions(
	paths []string,
	fromStdin bool,
	filter *filter.Filter,
	includeVendor bool,
	only string,
) chan string {
	if fromStdin {
		fchan := make(chan string)

		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				path := strings.TrimPrefix(f, "./")
				// Apply filter if enabled
				if !shouldIncludeFile(filter, path) {
					continue
				}
				// Apply "only" filter
				if only != "" && !matchesOnlyFilter(path, only) {
					continue
				}

				fchan <- path
			}

			close(fchan)
		}()

		return fchan
	}

	return crawlPathsWithOnly(paths, filter, includeVendor, only)
}

// crawlPaths walks paths and returns a channel of Go files.
func crawlPaths(paths []string, filter *filter.Filter, includeVendor bool) chan string {
	// For Go source file crawling, we don't have a node_modules exclusion config
	// so we pass true to maintain backward compatibility
	return crawlPathsWithFileCheck(paths, filter, includeVendor, true, isSourceFile)
}

// crawlPathsAllFiles walks paths and returns a channel of all files (not just source files).
// This is used for hash-based detection which works on any file type.
// The includeNodeModules parameter controls whether to include node_modules directories
// (excluded by default to avoid processing large dependency directories).
func crawlPathsAllFiles(
	paths []string,
	filter *filter.Filter,
	includeVendor, includeNodeModules bool,
) chan string {
	return crawlPathsWithFileCheck(paths, filter, includeVendor, includeNodeModules, nil)
}

// fileCheckFunc returns true if a file should be included based on its name.
// A nil function accepts all files.
type fileCheckFunc func(name string) bool

// crawlPathsWithFileCheck walks paths and returns a channel of files that pass the file check.
// The includeNodeModules parameter controls whether node_modules directories are included.
func crawlPathsWithFileCheck(
	paths []string,
	filter *filter.Filter,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			crawlSinglePath(path, filter, includeVendor, includeNodeModules, fileCheck, fchan)
		}

		close(fchan)
	}()

	return fchan
}

// crawlSinglePath handles crawling of a single path (file or directory).
func crawlSinglePath(
	path string,
	filter *filter.Filter,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
	fchan chan string,
) {
	info, err := os.Lstat(path)
	if err != nil {
		statError(path, err)
	}

	if !info.IsDir() {
		if shouldIncludeFile(filter, path) {
			fchan <- path
		}

		return
	}

	crawlDirectory(path, filter, includeVendor, includeNodeModules, fileCheck, fchan)
}

// crawlDirectory walks a directory tree and sends matching files to the channel.
func crawlDirectory(
	path string,
	filter *filter.Filter,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
	fchan chan string,
) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		return handleWalkEntry(
			path,
			info,
			filter,
			includeVendor,
			includeNodeModules,
			fileCheck,
			fchan,
		)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
		os.Exit(1)
	}
}

// handleWalkEntry processes a single entry during directory walk.
func handleWalkEntry(
	path string,
	info os.FileInfo,
	filter *filter.Filter,
	includeVendor bool,
	includeNodeModules bool,
	fileCheck fileCheckFunc,
	fchan chan string,
) error {
	// info can be nil if there was an error accessing the file/directory
	if info == nil {
		return nil
	}

	if shouldSkipPath(path, includeVendor, includeNodeModules) {
		return nil
	}

	if info.Name() == DSStoreFile {
		return nil
	}

	if !info.IsDir() && passesFileCheck(info.Name(), fileCheck) && shouldIncludeFile(filter, path) {
		fchan <- path
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

// matchesOnlyFilter returns true if the file matches the "only" filter criteria.
func matchesOnlyFilter(path, only string) bool {
	switch only {
	case "go":
		return strings.HasSuffix(path, ".go")
	case "templ":
		return strings.HasSuffix(path, ".templ")
	default:
		return true
	}
}

// crawlPathsWithOnly walks paths and returns a channel of Go files with "only" filter.
func crawlPathsWithOnly(paths []string, filter *filter.Filter, includeVendor bool, only string) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			crawlSinglePathWithOnly(path, filter, includeVendor, only, fchan)
		}

		close(fchan)
	}()

	return fchan
}

// crawlSinglePathWithOnly handles crawling of a single path with "only" filter.
func crawlSinglePathWithOnly(
	path string,
	filter *filter.Filter,
	includeVendor bool,
	only string,
	fchan chan string,
) {
	info, err := os.Lstat(path)
	if err != nil {
		statError(path, err)
	}

	if !info.IsDir() {
		if shouldIncludeFile(filter, path) && matchesOnlyFilter(path, only) {
			fchan <- path
		}

		return
	}

	crawlDirectoryWithOnly(path, filter, includeVendor, only, fchan)
}

// crawlDirectoryWithOnly walks a directory tree with "only" filter.
func crawlDirectoryWithOnly(
	path string,
	filter *filter.Filter,
	includeVendor bool,
	only string,
	fchan chan string,
) {
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, _ error) error {
		return handleWalkEntryWithOnly(
			walkPath,
			info,
			filter,
			includeVendor,
			only,
			fchan,
		)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
		os.Exit(1)
	}
}

// handleWalkEntryWithOnly processes a single entry during directory walk with "only" filter.
func handleWalkEntryWithOnly(
	path string,
	info os.FileInfo,
	filter *filter.Filter,
	includeVendor bool,
	only string,
	fchan chan string,
) error {
	// info can be nil if there was an error accessing the file/directory
	if info == nil {
		return nil
	}

	if shouldSkipPath(path, includeVendor, false) {
		return nil
	}

	if info.Name() == DSStoreFile {
		return nil
	}

	if !info.IsDir() && isSourceFile(info.Name()) && shouldIncludeFile(filter, path) && matchesOnlyFilter(path, only) {
		fchan <- path
	}

	return nil
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
