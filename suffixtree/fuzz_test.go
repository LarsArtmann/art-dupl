package suffixtree

import (
	"context"
	"testing"
)

// FuzzFindDuplOver tests that FindDuplOver never panics on arbitrary input
// and always closes the output channel.
func FuzzFindDuplOver(f *testing.F) {
	// Seed corpus: simple sequences that exercise common paths
	f.Add([]byte("abcabc"))
	f.Add([]byte("aaaa"))
	f.Add([]byte(""))
	f.Add([]byte("x"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1000 {
			return // keep test fast
		}

		tree := New()

		// Convert bytes to tokens
		tokens := make([]Token, len(data))
		for i, b := range data {
			tokens[i] = simpleToken(TokenValue(b))
		}

		tree.Update(tokens...)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := tree.FindDuplOver(ctx, 2)

		// Must drain and close without panic
		count := 0
		for range ch {
			count++
			if count > 10000 {
				cancel()

				break
			}
		}
	})
}

// simpleToken implements Token for fuzz testing.
type simpleToken TokenValue

func (s simpleToken) Val() TokenValue { return TokenValue(s) }

// FuzzFindDuplOverCancellation tests that context cancellation
// properly stops the walk and closes the channel.
func FuzzFindDuplOverCancellation(f *testing.F) {
	f.Add([]byte("abcdefghij"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}

		tree := New()

		tokens := make([]Token, len(data))
		for i, b := range data {
			tokens[i] = simpleToken(TokenValue(b + 1)) // +1 to avoid zero
		}

		tree.Update(tokens...)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		ch := tree.FindDuplOver(ctx, 1)

		// Channel must close even with cancelled context
		for range ch {
		}
	})
}
