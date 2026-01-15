package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/cmd"
	"github.com/charmbracelet/fang"
)

func main() {
	// Create root command
	rootCmd := cmd.NewRootCommand()

	// Add all flags
	cmd.AddFlags(rootCmd)

	// Enhanced error handler with context-aware suggestions
	errorHandler := func(w io.Writer, _ fang.Styles, err error) {
		if _, err := fmt.Fprintf(w, "\n❌ ERROR: %v\n\n", err); err != nil {
			// Can't write error, continue anyway
			_ = err // Explicitly ignore error
		}

		// Provide context-aware suggestions based on error type
		switch {
		case fmt.Sprint(err) == "flag: help requested":
			if _, err := fmt.Fprintf(w, "💡 Use examples below to get started:\n\n"); err != nil {
				_ = err
			}
			if _, err := fmt.Fprintf(w, "  art-dupl                    # Analyze current directory\n"); err != nil {
				_ = err
			}
			if _, err := fmt.Fprintf(w, "  art-dupl -t 50 ./src        # Higher threshold for larger clones\n"); err != nil {
				_ = err
			}
			if _, err := fmt.Fprintf(w, "  art-dupl --json -t 20 . | jq # JSON output with post-processing\n"); err != nil {
				_ = err
			}
			if _, err := fmt.Fprintf(w, "  art-dupl --all --output-dir ./reports # Generate all formats\n"); err != nil {
				_ = err
			}
		default:
			if _, err := fmt.Fprintf(w, "Quick Fix: Check file paths and permissions\n"); err != nil {
				_ = err
			}
			if _, err := fmt.Fprintf(w, "Get Help: art-dupl --help\n"); err != nil {
				_ = err
			}
		}
		if _, err := fmt.Fprintf(w, "\n📚 Visit https://github.com/LarsArtmann/art-dupl for documentation\n"); err != nil {
			_ = err
		}
	}

	// Enable fang features for better CLI experience
	options := []fang.Option{
		fang.WithVersion(cmd.GetVersion()),
		fang.WithColorSchemeFunc(fang.DefaultColorScheme),
		fang.WithErrorHandler(errorHandler),
	}

	if err := fang.Execute(context.Background(), rootCmd, options...); err != nil {
		os.Exit(1)
	}
}
