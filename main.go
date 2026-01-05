package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() { //nolint:cyclop,funlen // Main entry point with complex error handling and CLI setup
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
  art-dupl --plumbing --sort occurrence ./src # Plumbing sorted by occurrence
  art-dupl --all ./src                     # Generate all formats for all detection methods
  art-dupl --all --output-dir ./my-reports ./src  # Custom output directory`,
		Args: cobra.ArbitraryArgs, // Allow any number of positional arguments
		RunE: runCmd,
	}

	// Add all flags to root command with better descriptions
	rootCmd.Flags().StringP("config", "c", "", "path to configuration file (JSON format)")
	rootCmd.Flags().Bool("vendor", false, "include vendor directory in analysis")
	rootCmd.Flags().CountP("verbose", "v", "enable verbose logging (repeat for more verbosity)")
	rootCmd.Flags().IntP("threshold", "t", 15, "minimum token sequence size to consider as clone (default: 15)")
	rootCmd.Flags().BoolP("files", "f", false, "read file names from stdin, one per line")
	rootCmd.Flags().Bool("html", false, "output results as HTML with syntax-highlighted code fragments")
	rootCmd.Flags().BoolP("json", "j", false, "output structured JSON format with metadata and statistics")
	rootCmd.Flags().BoolP("plumbing", "p", false, "output machine-readable plumbing format for script integration")
	rootCmd.Flags().StringP("sort", "s", "size", "sort clone groups by: size, occurrence, hash (default: size)")
	rootCmd.Flags().StringP("detection-methods", "m", "art-dupl", "detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)")
	rootCmd.Flags().BoolP("all", "a", false, "generate all output formats for all detection methods")
	rootCmd.Flags().StringP("output-dir", "o", "reports/art-dupl", "output directory for generated files (used with --all)")

	// Add hidden flags for advanced features
	rootCmd.Flags().Bool("profile", false, "enable performance profiling")
	rootCmd.Flags().String("timeout", "30m", "maximum execution time (default: 30m)")
	_ = rootCmd.Flags().MarkHidden("profile")
	_ = rootCmd.Flags().MarkHidden("timeout")

	// Enhanced error handler with context-aware suggestions
	errorHandler := func(w io.Writer, _ fang.Styles, err error) {
		if _, err := fmt.Fprintf(w, "\n❌ ERROR: %v\n\n", err); err != nil {
			// Can't write the error, continue anyway
			_ = err // Explicitly ignore error
		}

		// Provide context-aware suggestions based on error type
		switch {
		case fmt.Sprint(err) == "flag: help requested":
			if _, err := fmt.Fprintf(w, "💡 Use examples below to get started:\n\n"); err != nil {
				// Can't write help, continue anyway
				_ = err // Explicitly ignore error
			}
			if _, err := fmt.Fprintf(w, "  art-dupl                    # Analyze current directory\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
			if _, err := fmt.Fprintf(w, "  art-dupl -t 50 ./src        # Higher threshold for larger clones\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
			if _, err := fmt.Fprintf(w, "  art-dupl --json -t 20 . | jq # JSON output with post-processing\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
			if _, err := fmt.Fprintf(w, "  art-dupl --all --output-dir ./reports # Generate all formats\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
		default:
			if _, err := fmt.Fprintf(w, "Quick Fix: Check file paths and permissions\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
			if _, err := fmt.Fprintf(w, "Get Help: art-dupl --help\n"); err != nil {
				// Can't write help examples
				_ = err // Explicitly ignore error
			}
		}
		if _, err := fmt.Fprintf(w, "\n📚 Visit https://github.com/LarsArtmann/art-dupl for documentation\n"); err != nil {
			// Can't write final message
			_ = err // Explicitly ignore error
		}
	}

	// Enable fang features for better CLI experience
	options := []fang.Option{
		fang.WithVersion(GetVersion()),
		fang.WithColorSchemeFunc(fang.DefaultColorScheme), // Auto-detect dark/light theme
		fang.WithErrorHandler(errorHandler),
	}

	if err := fang.Execute(context.Background(), rootCmd, options...); err != nil {
		os.Exit(1)
	}
}

// runCmd implements the Cobra command execution.
func runCmd(cmd *cobra.Command, args []string) error {
	return runCobraCommand(cmd, args)
}
