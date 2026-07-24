package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
)

const progressInterval = 5 * time.Second

// progressFilesChan wraps a file channel with periodic progress reporting.
// Prints the file count to stderr every progressInterval and the final count
// when the channel closes. Returns the input channel unchanged when progress
// is suppressed (Quiet, non-text output, or ARTDUPL_NO_PROGRESS=1).
func progressFilesChan(
	ctx context.Context,
	ch chan string,
	cfg *config.Config,
	outputFormat config.OutputFormat,
	progressOut io.Writer,
) chan string {
	if !shouldShowProgress(cfg, outputFormat) {
		return ch
	}

	out := make(chan string)

	go func() {
		defer close(out)

		count := 0

		ticker := time.NewTicker(progressInterval)
		defer ticker.Stop()

		for {
			select {
			case path, ok := <-ch:
				if !ok {
					if count > 0 {
						_, _ = fmt.Fprintf(progressOut, "    %d files discovered\n", count)
					}

					return
				}

				count++

				select {
				case out <- path:
				case <-ctx.Done():
					return
				}
			case <-ticker.C:
				if count > 0 {
					_, _ = fmt.Fprintf(progressOut, "    %d files so far...\n", count)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func shouldShowProgress(cfg *config.Config, outputFormat config.OutputFormat) bool {
	if cfg.Quiet {
		return false
	}

	if os.Getenv("ARTDUPL_NO_PROGRESS") == "1" {
		return false
	}

	return outputFormat == config.OutputFormatText || cfg.Verbose
}
