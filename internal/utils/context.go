package utils

import (
	"context"
	"fmt"
	"os"
	"time"
)

// ApplyTimeout applies a timeout to the given context if timeoutSeconds is greater than 0.
// It returns the potentially modified context and a cancel function that should be deferred by the caller.
// If a timeout is applied, it prints a message to stderr indicating the timeout duration.
func ApplyTimeout(ctx context.Context, timeoutSeconds int) (context.Context, context.CancelFunc) {
	if timeoutSeconds <= 0 {
		return ctx, func() {}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	fmt.Fprintf(os.Stderr, "⏱️  Execution timeout: %ds\n", timeoutSeconds)

	return ctx, cancel
}
