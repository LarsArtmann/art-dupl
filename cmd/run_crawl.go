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
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				statError(path, err)
			}

			if !info.IsDir() {
				// Apply filter to single file
				if !shouldIncludeFile(filter, path) {
					continue
				}

				fchan <- path

				continue
			}

			err = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
				// Skip vendor and git directories
				if shouldSkipPath(path, includeVendor) {
					return nil
				}

				// Skip .DS_Store files
				if info.Name() == cli.DSStoreFile {
					return nil
				}

				if !info.IsDir() && isSourceFile(info.Name()) {
					// Apply filter to file
					if !shouldIncludeFile(filter, path) {
						return nil
					}

					fchan <- path
				}

				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
				os.Exit(1)
			}
		}

		close(fchan)
	}()

	return fchan
}

// isSourceFile returns true if the filename has a supported source file extension.
func isSourceFile(name string) bool {
	return strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".templ")
}

// shouldSkipPath returns true if the path should be skipped due to being a vendor or git directory.
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

	return false
}

// crawlPathsAllFiles walks paths and returns a channel of all files (not just source files).
// This is used for hash-based detection which works on any file type.
func crawlPathsAllFiles(paths []string, filter *filter.Filter, includeVendor bool) chan string {
	fchan := make(chan string)

	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				statError(path, err)
			}

			if !info.IsDir() {
				// Apply filter to single file
				if !shouldIncludeFile(filter, path) {
					continue
				}

				fchan <- path

				continue
			}

			err = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
				// Skip vendor and git directories
				if shouldSkipPath(path, includeVendor) {
					return nil
				}

				// Skip .DS_Store files
				if info.Name() == cli.DSStoreFile {
					return nil
				}

				if !info.IsDir() {
					// Apply filter to file
					if !shouldIncludeFile(filter, path) {
						return nil
					}

					fchan <- path
				}

				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
				os.Exit(1)
			}
		}

		close(fchan)
	}()

	return fchan
}
