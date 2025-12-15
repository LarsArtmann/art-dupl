package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dupl [flags] [paths...]",
		Short: "Find code clones",
		Long: `dupl finds code clones in Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones
while ignoring literal values using suffix tree algorithms.`,
		RunE: runCmd,
		Args: cobra.ArbitraryArgs, // Allow any number of positional arguments
	}

	// Add all flags
	rootCmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	rootCmd.Flags().BoolP("vendor", "", false, "include vendor directory in analysis")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging to show processing progress")
	rootCmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone")
	rootCmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	rootCmd.Flags().BoolP("html", "", false, "output results as HTML with syntax-highlighted code fragments")
	rootCmd.Flags().BoolP("json", "j", false, "output structured JSON format with metadata and statistics")
	rootCmd.Flags().BoolP("plumbing", "p", false, "output machine-readable plumbing format for script integration")
	rootCmd.Flags().StringP("sort", "s", "size", "sort clone groups by: size, occurrence, hash")

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

// runCmd implements the Cobra command execution
func runCmd(cmd *cobra.Command, args []string) error {
	return runCobraCommand(cmd, args)
}
