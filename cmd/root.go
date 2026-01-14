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

Examples:
  art-dupl ./src                          # Default analysis
  art-dupl -t 20 ./src                    # Higher threshold
  art-dupl --json -t 20 ./src             # JSON output with threshold
  art-dupl --html --vendor ./src           # HTML with vendor included
  art-dupl --plumbing --sort occurrence ./src # Most widespread clones first
  art-dupl --all ./src                     # Generate all formats for all detection methods
  art-dupl --all --output-dir ./my-reports ./src  # Custom output directory
  art-dupl --filter-generated ./src         # Filter out auto-generated code (sqlc, templ)
  art-dupl --filter-generated --include-sqlc ./src  # Filter but keep sqlc files
  art-dupl --filter-generated --include-pattern "vendor/*" ./src  # Include vendor directory`,
		Args: cobra.ArbitraryArgs,
		RunE: runCmd,
	}

	return rootCmd
}
