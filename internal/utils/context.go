// Package utils provides utility functions for context management and file processing.
package utils

import (
	"context"
	"fmt"
	"os"
	"time"
)

// ApplyTimeout applies a timeout to the given context if timeout is greater than 0.
// It returns the potentially modified context and a cancel function that should be deferred by the caller.
// If a timeout is applied, it prints a message to stderr indicating the timeout duration.
func ApplyTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	fmt.Fprintf(os.Stderr, "⏱️  Execution timeout: %s\n", timeout)

	return ctx, cancel
}
