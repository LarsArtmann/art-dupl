package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "art-dupl [flags] [paths...]",
		Short: "Find code clones",
		Long: `art-dupl finds code clones in Go source files.

It analyzes abstract syntax trees (ASTs) to find structural code clones
while ignoring literal values using suffix tree algorithms.

Examples:
  art-dupl ./src                          # Default analysis
  art-dupl -t 20 ./src                    # Higher threshold
  art-dupl --json -t 20 ./src             # JSON output with threshold
  art-dupl --html --vendor ./src           # HTML with vendor included
  art-dupl --plumbing --sort occurrence ./src # Plumbing sorted by occurrence`,
		Args: cobra.ArbitraryArgs, // Allow any number of positional arguments
		RunE:  runCmd,
	}

	// Add all flags to root command
	rootCmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	rootCmd.Flags().BoolP("vendor", "", false, "include vendor directory in analysis")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging to show processing progress")
	rootCmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone")
	rootCmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	rootCmd.Flags().BoolP("html", "", false, "output results as HTML with syntax-highlighted code fragments")
	rootCmd.Flags().BoolP("json", "j", false, "output structured JSON format with metadata and statistics")
	rootCmd.Flags().BoolP("plumbing", "p", false, "output machine-readable plumbing format for script integration")
	rootCmd.Flags().StringP("sort", "s", "size", "sort clone groups by: size, occurrence, hash")
	rootCmd.Flags().StringP("detection-methods", "m", "art-dupl", "detection methods to use: hash, art-dupl, hash,art-dupl")

	// Custom error handler for better user experience
	errorHandler := func(w io.Writer, styles fang.Styles, err error) {
		fmt.Fprintf(w, "❌ Error: %v\n", err)
		fmt.Fprintf(w, "💡 Run '%s --help' for usage information\n", rootCmd.Name())
	}

	// Enable fang features for better CLI experience
	options := []fang.Option{
		fang.WithVersion(GetVersion()),
		fang.WithTheme(fang.DefaultTheme(true)), // Auto-detect dark/light theme
		fang.WithColorSchemeFunc(fang.AnsiColorScheme), // Better color support
		fang.WithErrorHandler(errorHandler), // Better error messages
	}

	if err := fang.Execute(context.Background(), rootCmd, options...); err != nil {
		os.Exit(1)
	}
}

// runCmd implements the Cobra command execution
func runCmd(cmd *cobra.Command, args []string) error {
	return runCobraCommand(cmd, args)
}