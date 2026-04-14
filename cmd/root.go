package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root Cobra command.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "art-dupl [flags] [paths...]",
		Short:   "Find code clones",
		Long:    "art-dupl finds code clones in Go source files.",
		Args:    cobra.ArbitraryArgs,
		RunE:    runCmd,
		Example: "art-dupl ./src",
	}

	// Add stats subcommand
	statsCmd := NewStatsCommand()
	rootCmd.AddCommand(statsCmd)

	return rootCmd
}
