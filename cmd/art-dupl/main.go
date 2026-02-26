package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/cmd"
	"github.com/charmbracelet/fang"
)

//nolint:gocognit // Main function orchestrates CLI setup with enhanced error handling and context-aware suggestions
func main() {
	// Create root command
	rootCmd := cmd.NewRootCommand()

	// Add all flags
	cmd.AddFlags(rootCmd)

	// Enhanced error handler with fang styling and context-aware suggestions
	errorHandler := func(w io.Writer, styles fang.Styles, err error) {
		// Use fang's default error rendering as base
		fang.DefaultErrorHandler(w, styles, err)

		// Add context-aware suggestions based on error type
		errStr := err.Error()
		switch errStr {
		case "flag: help requested":
			if _, writeErr := fmt.Fprintln(w); writeErr == nil {
				if _, writeErr := fmt.Fprintln(
					w,
					styles.Text.Render("💡 Use examples below to get started:"),
				); writeErr == nil {
					examples := []string{
						"  art-dupl                    # Analyze current directory",
						"  art-dupl -t 50 ./src        # Higher threshold for larger clones",
						"  art-dupl --json -t 20 . | jq # JSON output with post-processing",
						"  art-dupl --all --output-dir ./reports # Generate all formats",
					}
					for _, ex := range examples {
						if _, writeErr := fmt.Fprintln(
							w,
							styles.Codeblock.Program.Name.Render(ex),
						); writeErr != nil {
							break
						}
					}
				}
			}
		default:
			// Provide helpful hints for common errors
			if _, writeErr := fmt.Fprintln(w); writeErr == nil {
				if _, writeErr := fmt.Fprintln(
					w,
					styles.Text.Render("Quick Fix: Check file paths and permissions"),
				); writeErr == nil {
					if _, writeErr := fmt.Fprintln(
						w,
						styles.Text.Render("Get Help: art-dupl --help"),
					); writeErr != nil {
						_ = writeErr
					}
				}
			}
		}

		if _, writeErr := fmt.Fprintln(w); writeErr == nil {
			if _, writeErr := fmt.Fprintln(
				w,
				styles.Text.Render(
					"📚 Visit https://github.com/LarsArtmann/art-dupl for documentation",
				),
			); writeErr != nil {
				_ = writeErr
			}
		}
	}

	// Enable fang features for better CLI experience
	options := []fang.Option{
		fang.WithVersion(cmd.GetVersion()),
		fang.WithColorSchemeFunc(fang.DefaultColorScheme),
		fang.WithErrorHandler(errorHandler),
		fang.WithNotifySignal(os.Interrupt), // Handle Ctrl+C gracefully
	}

	err := fang.Execute(context.Background(), rootCmd, options...)
	if err != nil {
		os.Exit(1)
	}
}
