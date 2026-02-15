package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root Cobra command.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "art-dupl [flags] [paths...]",
		Short: "Find code clones",
		Long: `art-dupl finds code clones in Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones
while ignoring literal values using suffix tree algorithms.

Sorting Options:
  - size: Shows largest clones first (highest token count)
  - occurrence: Shows most widespread clones first (most files)
  - hash: Alphabetical order by hash value

Detection Methods:
  - art-dupl: Original suffix tree algorithm (default)
  - hash: Fast hash-based detection
  - hash,art-dupl: Run both methods for comprehensive analysis

Examples:
  # Basic analysis
  art-dupl ./src

  # Higher threshold to find larger clones
  art-dupl -t 20 ./src

  # JSON output for post-processing
  art-dupl --json -t 20 ./src | jq

  # HTML report with syntax highlighting
  art-dupl --html --vendor ./src > report.html

  # Most widespread clones first
  art-dupl --plumbing --sort occurrence ./src

  # Generate all formats for all detection methods
  art-dupl --all ./src

  # Custom output directory
  art-dupl --all --output-dir ./my-reports ./src

  # Advanced filtering
  art-dupl --filter-generated ./src
  art-dupl --filter-generated --include-sqlc ./src
  art-dupl --include-templ ./src
  art-dupl --include-pattern "vendor/*" ./src

  # Semantic-aware detection (group by identifier names)
  art-dupl --semantic ./src

  # Read file list from stdin
  find . -name "*.go" | art-dupl --files

  # Verbose mode for debugging
  art-dupl -vv ./src`,
		Args: cobra.ArbitraryArgs,
		RunE: runCmd,
	}

	// Add stats subcommand
	statsCmd := NewStatsCommand()
	rootCmd.AddCommand(statsCmd)

	return rootCmd
}
