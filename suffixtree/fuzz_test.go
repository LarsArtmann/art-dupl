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

	// High-fanout seeds: many distinct first tokens force root (and internal
	// states) past the linearScanMax cutoff into the binary-search branch of
	// findTran, and stress addTran's sorted insertion across the full byte
	// alphabet.
	f.Add(fullAlphabetRamp(64))
	f.Add(fullAlphabetRamp(200))
	f.Add(interleaveRamp(100))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1000 {
			return // keep test fast
		}

		tree := New()

		// Convert bytes to tokens
		tokens := make([]Token, 0, len(data))
		for _, b := range data {
			tokens = append(tokens, simpleToken(TokenValue(b)))
		}

		mustUpdate(tree, tokens...)

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

// FuzzCtxCancelFindDuplOver tests that context cancellation
// properly stops the walk and closes the channel.
func FuzzCtxCancelFindDuplOver(f *testing.F) {
	f.Add([]byte("abcdefghij"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}

		tree := New()

		tokens := make([]Token, 0, len(data))
		for _, b := range data {
			tokens = append(tokens, simpleToken(TokenValue(b+1))) // +1 to avoid zero
		}

		mustUpdate(tree, tokens...)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		ch := tree.FindDuplOver(ctx, 1)

		// Channel must close even with cancelled context
		for range ch {
		}
	})
}

// FuzzFindDuplOverParallel tests that FindDuplOverParallel never panics on
// arbitrary input and always closes the output channel. Mirrors FuzzFindDuplOver
// for the parallel search path.
func FuzzFindDuplOverParallel(f *testing.F) {
	f.Add([]byte("abcabc"))
	f.Add([]byte("aaaa"))
	f.Add([]byte(""))
	f.Add([]byte("x"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1000 {
			return // keep test fast
		}

		tree := New()

		tokens := make([]Token, 0, len(data))
		for _, b := range data {
			tokens = append(tokens, simpleToken(TokenValue(b)))
		}

		mustUpdate(tree, tokens...)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := tree.FindDuplOverParallel(ctx, 2, 4)

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

// FuzzCtxCancelFindDuplOverParallel tests that context cancellation properly
// stops the parallel walk and closes the channel.
func FuzzCtxCancelFindDuplOverParallel(f *testing.F) {
	f.Add([]byte("abcdefghij"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}

		tree := New()

		tokens := make([]Token, 0, len(data))
		for _, b := range data {
			tokens = append(tokens, simpleToken(TokenValue(b+1)))
		}

		mustUpdate(tree, tokens...)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		ch := tree.FindDuplOverParallel(ctx, 1, 4)

		for range ch {
		}
	})
}

// fullAlphabetRamp returns n strictly ascending distinct byte values — the
// maximum-fanout shape: the root ends up with n transitions, far above
// linearScanMax.
func fullAlphabetRamp(n int) []byte {
	out := make([]byte, n)
	for i := range n {
		out[i] = byte(i % 256)
	}

	return out
}

// interleaveRamp alternates ascending ramps with descending probes so addTran
// sees insertions at both ends and the middle of the sorted slice.
func interleaveRamp(n int) []byte {
	out := make([]byte, 0, n)
	for i := range n {
		if i%2 == 0 {
			out = append(out, byte(i%256))
		} else {
			out = append(out, byte(255-(i%256)))
		}
	}

	return out
}

// FuzzTranLookupSemantics builds a tree from arbitrary bytes and verifies
// that findTran agrees with the naive reference scan on every state for
// every distinct token in the stream plus boundary probes. Stronger than a
// no-panic check: a binary-search off-by-one on high-fanout states fails
// here even without crashing.
func FuzzTranLookupSemantics(f *testing.F) {
	f.Add(fullAlphabetRamp(64))
	f.Add(interleaveRamp(100))
	f.Add([]byte("abababab"))
	f.Add([]byte{0, 255, 0, 255, 128})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 500 {
			return
		}

		tree := New()
		mustUpdate(tree, valueTokens(byteValues(data))...)

		probes := map[TokenValue]bool{0: true, 1: true, -1: true}
		for _, b := range data {
			probes[TokenValue(b+1)] = true
		}

		for _, s := range enumerateStates(tree) {
			for probe := range probes {
				if s.findTran(tree.data, probe) != referenceFindTran(s, tree.data, probe) {
					t.Fatalf("findTran(%d) disagrees with reference scan", probe)
				}
			}
		}
	})
}

// byteValues converts raw bytes to the +1 token domain the fuzz targets use
// (0 is reserved as the sentinel separator in the tree).
func byteValues(data []byte) []TokenValue {
	out := make([]TokenValue, 0, len(data))
	for _, b := range data {
		out = append(out, TokenValue(b+1))
	}

	return out
}
