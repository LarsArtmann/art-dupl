package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/cli"
	"github.com/LarsArtmann/art-dupl/pkg/filter"
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

				fchan <- path
			}

			close(fchan)
		}()

		return fchan
	}

	return crawlPaths(paths, filter, includeVendor)
}

// crawlPaths walks paths and returns a channel of Go files.
func crawlPaths(paths []string, filter *filter.Filter, includeVendor bool) chan string {
	return crawlPathsWithFileCheck(paths, filter, includeVendor, isSourceFile)
}

// crawlPathsAllFiles walks paths and returns a channel of all files (not just source files).
// This is used for hash-based detection which works on any file type.
func crawlPathsAllFiles(paths []string, filter *filter.Filter, includeVendor bool) chan string {
	return crawlPathsWithFileCheck(paths, filter, includeVendor, nil)
}

// fileCheckFunc returns true if a file should be included based on its name.
// A nil function accepts all files.
type fileCheckFunc func(name string) bool

// crawlPathsWithFileCheck walks paths and returns a channel of files that pass the file check.
func crawlPathsWithFileCheck(
	paths []string,
	filter *filter.Filter,
	includeVendor bool,
	fileCheck fileCheckFunc,
) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			crawlSinglePath(path, filter, includeVendor, fileCheck, fchan)
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

	crawlDirectory(path, filter, includeVendor, fileCheck, fchan)
}

// crawlDirectory walks a directory tree and sends matching files to the channel.
func crawlDirectory(
	path string,
	filter *filter.Filter,
	includeVendor bool,
	fileCheck fileCheckFunc,
	fchan chan string,
) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		return handleWalkEntry(path, info, filter, includeVendor, fileCheck, fchan)
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
	fileCheck fileCheckFunc,
	fchan chan string,
) error {
	// info can be nil if there was an error accessing the file/directory
	if info == nil {
		return nil
	}

	if shouldSkipPath(path, includeVendor) {
		return nil
	}

	if info.Name() == cli.DSStoreFile {
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

// shouldSkipPath returns true if the path should be skipped due to being a vendor, git, or node_modules directory.
func shouldSkipPath(path string, includeVendor bool) bool {
	if !includeVendor && (strings.HasPrefix(path, cli.VendorDirPrefix) ||
		strings.Contains(path, cli.VendorDirInPath)) {
		return true
	}

	// Skip .git directories
	if strings.HasPrefix(path, cli.GitDirPrefix) ||
		strings.Contains(path, cli.GitDirInPath) {
		return true
	}

	// Skip node_modules directories (always excluded, not configurable like vendor)
	// This prevents processing large dependency directories in hash detection mode
	if strings.HasPrefix(path, cli.NodeModulesDirPrefix) ||
		strings.Contains(path, cli.NodeModulesDirInPath) {
		return true
	}

	return false
}
