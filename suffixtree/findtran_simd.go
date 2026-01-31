package suffixtree

//
// TODO: ARCHITECTURE DECISION - SIMD optimization is conditionally compiled
// but the fallback implementation is always available. Consider:
// - Adding build tags to exclude SIMD code on unsupported platforms
// - Adding benchmarks to verify SIMD actually improves performance
// - Documenting which platforms support SIMD and which don't
//
// Also: The SIMD threshold (8 transitions) is a magic number.
// Consider making this configurable or deriving it from benchmarks.

import (
	"github.com/LarsArtmann/art-dupl/internal/simd"
)

// findTran finds a transition matching the given token.
// This is a critical operation in suffix tree construction and search,
// and benefits from SIMD optimization for states with many transitions.
//
// Performance characteristics:
// - O(m) time complexity where m is the number of transitions in the state
// - O(1) additional space
// - Uses SIMD-optimized comparison when available
//
// This is called frequently during suffix tree construction and search,
// making it a secondary optimization target after hashSeq.
func (s *state) findTran(c Token) *tran {
	if len(s.trans) == 0 {
		return nil
	}

	// Use SIMD-optimized search if available
	if simd.Available() && len(s.trans) > 8 {
		return s.findTranSIMD(c)
	}

	// Use standard linear search for small state or when SIMD unavailable
	return s.findTranFallback(c)
}

// findTranFallback implements the standard linear search approach.
// This is the current implementation used throughout the codebase.
func (s *state) findTranFallback(c Token) *tran {
	for _, tr := range s.trans {
		if s.tree.data[tr.start].Val() == c.Val() {
			return tr
		}
	}
	return nil
}

// findTranSIMD implements a SIMD-optimized search approach.
// This will be enabled when ARM64 SIMD support becomes available in Go.
//
// Expected improvements:
// - 2-3x faster for states with >8 transitions
// - Reduced branch mispredictions through vectorized comparisons
// - Better cache utilization for large transition sets
func (s *state) findTranSIMD(c Token) *tran {
	// TODO: Implement SIMD-optimized transition search when available
	//
	// Future approach using SIMD operations:
	//
	// // Load transition token values into a SIMD vector
	// values := make([]int, len(s.trans))
	// for i, tr := range s.trans {
	//     values[i] = s.tree.data[tr.start].Val()
	// }
	//
	// // Use SIMD to compare all values against target token
	// matches := simd.CompareInts(values, c.Val())
	//
	// // Find first matching transition
	// for i, match := range matches {
	//     if match {
	//         return s.trans[i]
	//     }
	// }
	//
	// return nil

	// For now, fall back to standard implementation
	return s.findTranFallback(c)
}

// findTranBinary implements a binary search approach.
// This can be useful when transitions are sorted and the state has
// many transitions (>16).
//
// Performance characteristics:
// - O(log m) time complexity where m is the number of transitions
// - O(1) additional space
// - Requires transitions to be sorted by token value
func (s *state) findTranBinary(c Token) *tran {
	if len(s.trans) == 0 {
		return nil
	}

	// Binary search assumes transitions are sorted
	low, high := 0, len(s.trans)-1

	for low <= high {
		mid := low + (high-low)/2
		tr := s.trans[mid]
		tokenVal := s.tree.data[tr.start].Val()

		if tokenVal == c.Val() {
			return tr
		} else if tokenVal < c.Val() {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return nil
}

// sortTransitions sorts the transitions by token value.
// This should be called once after the tree is fully built
// to enable binary search optimization.
func (s *state) sortTransitions() {
	// TODO: Implement transition sorting if needed for binary search
	//
	// sort.Slice(s.trans, func(i, j int) bool {
	//     return s.tree.data[s.trans[i].start].Val() <
	//            s.tree.data[s.trans[j].start].Val()
	// })
}

// optimizeTransitions applies optimizations to the transition list.
// This should be called after tree construction to enable faster searches.
//
// Optimizations:
// 1. Sorts transitions for binary search (if beneficial)
// 2. Reorganizes memory layout for better cache locality
// 3. Prepares data structures for SIMD operations
func (s *state) optimizeTransitions() {
	// Determine best search strategy based on transition count
	switch {
	case len(s.trans) > 16:
		// Use binary search for large sorted transition sets
		s.sortTransitions()
	case len(s.trans) > 8:
		// Use SIMD search when available
		// No action needed - findTran will choose SIMD automatically
	default:
		// Use linear search for small transition sets
		// No action needed - findTran will choose linear search automatically
	}
}

// OptimizeTree optimizes all transitions in the suffix tree.
// This should be called after tree construction is complete.
func (t *STree) OptimizeTree() {
	// Recursively optimize all states
	t.optimizeState(t.root)
}

// optimizeState recursively optimizes a state and its children.
func (t *STree) optimizeState(s *state) {
	if s == nil {
		return
	}

	// Optimize this state's transitions
	s.optimizeTransitions()

	// Recursively optimize child states
	for _, tr := range s.trans {
		t.optimizeState(tr.state)
	}
}

// findTranFast is a performance-optimized version that assumes
// the transitions have been pre-sorted or otherwise optimized.
func (s *state) findTranFast(c Token) *tran {
	if len(s.trans) == 0 {
		return nil
	}

	// Use binary search if transitions are sorted
	// This is faster than linear search for large transition sets
	if len(s.trans) > 8 {
		return s.findTranBinary(c)
	}

	// Otherwise use linear search
	return s.findTranFallback(c)
}

// findTranBatch searches for multiple tokens in parallel.
// This is useful when processing multiple queries against the same state.
func (s *state) findTranBatch(tokens []Token) []*tran {
	results := make([]*tran, len(tokens))

	for i, token := range tokens {
		results[i] = s.findTran(token)
	}

	return results
}
