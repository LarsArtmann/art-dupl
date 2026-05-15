package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/LarsArtmann/art-dupl/cmd"
	"github.com/charmbracelet/fang"
)

const exitCodeInterrupt = 130

func main() {
	// Create root command
	rootCmd := cmd.NewRootCommand()

	// Add all flags
	cmd.AddFlags(rootCmd)

	// Enhanced error handler with fang styling and context-aware suggestions
	errorHandler := func(w io.Writer, styles fang.Styles, err error) {
		// Special handling for context cancellation (Ctrl+C)
		if errors.Is(err, context.Canceled) {
			cancelMsg := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				Render("⏹ CANCELED")

			_, _ = fmt.Fprintln(w, cancelMsg)

			return
		}

		renderHelpLink := func(message string) {
			_, writeErr := fmt.Fprintln(w)
			if writeErr == nil {
				_, writeErr = fmt.Fprintln(
					w,
					styles.Text.Render(message),
				)
				if writeErr != nil {
					_ = writeErr
				}
			}
		}

		// Use fang's default error rendering as base
		fang.DefaultErrorHandler(w, styles, err)

		// Add context-aware suggestions based on error type
		errStr := err.Error()
		switch errStr {
		case "flag: help requested":
			_, writeErr := fmt.Fprintln(w)
			if writeErr == nil {
				_, writeErr = fmt.Fprintln(
					w,
					styles.Text.Render("💡 Use examples below to get started:"),
				)
				if writeErr == nil {
					examples := []string{
						"  art-dupl                    # Analyze current directory",
						"  art-dupl -t 50 ./src        # Higher threshold for larger clones",
						"  art-dupl --json -t 20 . | jq # JSON output with post-processing",
						"  art-dupl --all --output-dir ./reports # Generate all formats",
					}
					for _, ex := range examples {
						_, writeErr = fmt.Fprintln(
							w,
							styles.Codeblock.Program.Name.Render(ex),
						)
						if writeErr != nil {
							break
						}
					}
				}
			}
		default:
			// Provide helpful hints for common errors
			renderHelpLink("Quick Fix: Check file paths and permissions")
			renderHelpLink("Get Help: art-dupl --help")
		}

		renderHelpLink("📚 Visit https://github.com/LarsArtmann/art-dupl for documentation")
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
		// Use exit code 130 for SIGINT cancellation (128 + signal 2)
		if errors.Is(err, context.Canceled) {
			os.Exit(exitCodeInterrupt)
		}

		os.Exit(1)
	}
}
