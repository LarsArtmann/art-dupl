package job

import "context"

// sendCtx sends v on ch, or returns false if ctx is canceled before the send
// completes. This centralizes the context-aware send pattern used throughout
// the pipeline goroutines, preventing the check-then-send race that would
// occur with select-default + bare send.
func sendCtx[T any](ctx context.Context, ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	case <-ctx.Done():
		return false
	}
}
