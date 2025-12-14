package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
)

const (
	vendorDirPrefix = "vendor" + string(filepath.Separator)
	vendorDirInPath = string(filepath.Separator) + vendorDirPrefix
)

func main() {
	os.Exit(Run())
}

func filesFeed() chan string {
	if *files {
		fchan := make(chan string)
		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				fchan <- strings.TrimPrefix(f, "./")
			}
			close(fchan)
		}()
		return fchan
	}
	return crawlPaths(paths)
}

func crawlPaths(paths []string) chan string {
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				if _, err := fmt.Fprintf(cli.Stderr(), "error: cannot stat %s: %v\n", path, err); err != nil {
					cli.Exit(1)
					return
				}
				cli.Exit(1)
				return
			}
			if !info.IsDir() {
				fchan <- path
				continue
			}
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if !*vendor && (strings.HasPrefix(path, vendorDirPrefix) ||
					strings.Contains(path, vendorDirInPath)) {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					fchan <- path
				}
				return nil
			})
			if err != nil {
				if _, err := fmt.Fprintf(cli.Stderr(), "error: cannot walk %s: %v\n", path, err); err != nil {
					cli.Exit(1)
					return
				}
				cli.Exit(1)
				return
			}
		}
		close(fchan)
	}()
	return fchan
}

func printDupls(p printer.Printer, duplChan <-chan syntax.Match) error {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := p.PrintHeader(); err != nil {
		return err
	}
	for _, k := range keys {
		uniq := util.Unique(groups[k])
		if len(uniq) > 1 {
			// Set hash for JSONPrinter if applicable
			if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
				jsonPrinter.SetHash(k)
			}
			if err := p.PrintClones(uniq); err != nil {
				return err
			}
		}
	}

	// Handle JSON output special case
	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		return jsonPrinter.OutputJSON(*threshold, *sortBy)
	}

	return p.PrintFooter()
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: dupl [flags] [paths]

Paths:
  If given path is a file, dupl will use it regardless of
  file extension. If it is a directory, it will recursively
  search for *.go files in that directory.

  If no path is given, dupl will recursively search for *.go
  files in the current directory.

Flags:
  -config string
    	path to configuration file (JSON format)
  -files
    	read file names from stdin, one per line
  -html
    	output results as HTML with syntax-highlighted code fragments
  -json
    	output structured JSON format with metadata and statistics
  -plumbing
    	output machine-readable plumbing format for script integration
  -t, -threshold int
    	minimum token sequence size to consider as clone (default 15)
  -vendor
    	include vendor directory in analysis
  -v, -verbose
    	enable verbose logging to show processing progress

Output Formats:
  text     - Human-readable clone listing (default)
  html      - HTML report with syntax-highlighted code
  json      - Structured JSON data for automation/CI
  plumbing  - Machine-readable format for scripts

Configuration File:
  Create a JSON file with settings:
  {
    "threshold": 30,
    "outputFormat": "json",
    "paths": ["./src", "./lib"],
    "includeVendor": false,
    "verbose": true
  }
  Use with: dupl -config config.json

Examples:
  # Basic analysis with default threshold
  dupl

  # Higher threshold for larger clones only
  dupl -t 100

  # JSON output for CI/CD integration
  dupl -json -t 20

  # HTML report file
  dupl -html > report.html

  # Use configuration file
  dupl -config dupl.json ./src

  # Analyze test files only
  find . -name '*_test.go' | dupl -files

  # CI/CD: Fail if too many duplicates
  TOTAL_CLONES=$(dupl -json . | jq '.summary.total_clones')
  if [ "$TOTAL_CLONES" -gt 100 ]; then
    echo "Too many code duplicates: $TOTAL_CLONES"
    exit 1
  fi`)
	os.Exit(2)
}
