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

Exit codes:
  0   Success (no errors, clones may or may not have been found)
  1   General error
  2   Configuration or validation error
  3   Internal error
  130 Interrupted (Ctrl+C or timeout)`,
		Args:    cobra.ArbitraryArgs,
		RunE:    runCmd,
		Example: "art-dupl ./src",
	}

	// Add subcommands
	statsCmd := NewStatsCommand()
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(NewBaselineCommand())
	rootCmd.AddCommand(NewCheckCommand())
	rootCmd.AddCommand(NewVersionCommand())

	return rootCmd
}
