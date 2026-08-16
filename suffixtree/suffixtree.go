// Package suffixtree implements suffix tree for clone detection.
// Core types: STree (tree), state (node), tran (edge), Match (duplicate), Token (interface).
// Algorithm: Build incrementally, FindDuplOver() searches for sequences >= threshold.
// Performance: O(n*m) worst case, typically O(n). Memory: O(n*k).
// Optimization: Transitions are stored as a sorted []tran slice per state —
// linear scan for the common case (<= linearScanMax transitions), binary
// search for high-fanout states like the root. No per-state map allocation.
package suffixtree

import (
	"fmt"
	"math"
	"slices"

	"github.com/LarsArtmann/art-dupl/errors"
)

const infinity = math.MaxInt32

// Pos denotes position in data slice.
type Pos int32

// TokenValue represents a unique token identifier in the suffix tree.
// Uses int32 to keep memory pressure low while supporting large token sets.
type TokenValue int32

// Token represents a token in the suffix tree sequence.
// Implementations must provide a TokenValue that uniquely identifies the token type.
type Token interface {
	Val() TokenValue
}

// stateBlockSize is the number of states per arena block.
// Each block is a contiguous []state slice for cache-friendly traversal.
// 512 * 32 bytes = 16 KB per block. Benchmarked against 256/1024/4096:
// smaller blocks cap worst-case waste (one partially-filled block) for small
// trees without measurably hurting large-tree construction. Production builds
// ONE tree per analysis run (job.BuildTree), so the waste bound is per run,
// not per file.
const stateBlockSize = 512

// STree is a struct representing a suffix tree.
//
// The data slice stores TokenValue (int32) rather than Token (interface)
// to minimize memory pressure on large codebases. Each position in data
// corresponds 1:1 to a position in the caller's []*syntax.Node slice, so
// callers can index back into their own typed slice using suffix tree
// positions.
type STree struct {
	data     []TokenValue
	root     *state
	auxState *state // auxiliary state

	// active point
	s          *state
	start, end Pos

	// stateArena provides contiguous allocation for state structs,
	// improving cache locality during tree construction and search.
	// States within a block are adjacent in memory, reducing cache
	// misses when following suffix links or transitions.
	stateArena stateArena
}

// stateArena allocates state structs in contiguous blocks for cache locality.
// Go's GC is non-moving, so pointers into block backing arrays are stable
// for the lifetime of the arena (i.e., the STree).
type stateArena struct {
	blocks [][]state
	idx    int
}

func (a *stateArena) alloc() *state {
	if a.idx >= stateBlockSize || len(a.blocks) == 0 {
		a.blocks = append(a.blocks, make([]state, stateBlockSize))
		a.idx = 0
	}

	block := &a.blocks[len(a.blocks)-1]
	s := &(*block)[a.idx]
	a.idx++

	return s
}

// New creates new suffix tree.
func New() *STree {
	t := new(STree)
	t.data = make([]TokenValue, 0, 50)
	t.root = t.newState()
	t.auxState = t.newState()
	t.root.linkState = t.auxState
	t.s = t.root

	return t
}

// newState allocates a new state from the arena. The transition slice is
// allocated lazily by addTran on the first transition: leaf states never
// receive transitions, so roughly half of all states skip the allocation
// entirely.
func (t *STree) newState() *state {
	return t.stateArena.alloc()
}

// Update refreshes the suffix tree by new data.
// Tokens are converted to TokenValue immediately and stored compactly;
// the original Token objects are not retained, reducing memory by 75%
// compared to storing interface values.
func (t *STree) Update(data ...Token) error {
	for _, tok := range data {
		t.data = append(t.data, tok.Val())
	}

	for range data {
		t.update()

		s, start, err := t.canonize(t.s, t.start, t.end)
		if err != nil {
			return fmt.Errorf("suffixtree canonize failed during Update: %w", err)
		}

		t.s, t.start = s, start
		t.end++
	}

	return nil
}

// update transforms suffix tree T(n) to T(n+1).
//
//nolint:funcorder
func (t *STree) update() {
	oldr := t.root

	// (s, (start, end)) is the canonical reference pair for the active point
	s := t.s
	start, end := t.start, t.end

	var r *state

	for {
		var endPoint bool

		r, endPoint = t.testAndSplit(s, start, end-1)
		if endPoint {
			break
		}

		t.fork(r, end)

		if oldr != t.root {
			oldr.linkState = r
		}

		oldr = r

		var canonizeErr error

		s, start, canonizeErr = t.canonize(s.linkState, start, end-1)
		if canonizeErr != nil {
			panic(fmt.Sprintf("suffixtree canonize failed during update: %v", canonizeErr))
		}
	}

	if oldr != t.root {
		oldr.linkState = r
	}

	// update active point
	t.s = s
	t.start = start
}

// testAndSplit tests whether a state with canonical ref. pair
// (s, (start, end)) is the end point, that is, a state that have
// a c-transition. If not, then state (exs, (start, end)) is made
// explicit (if not already so).
//
//nolint:nonamedreturns
func (t *STree) testAndSplit(s *state, start, end Pos) (exs *state, endPoint bool) {
	c := t.data[t.end]
	if start <= end {
		tr := s.findTran(t.data, t.data[start])

		splitPoint := tr.start + end - start + 1
		if t.data[splitPoint] == c {
			return s, true
		}
		// make the (s, (start, end)) state explicit
		newSt := t.newState()
		t.addTran(newSt, splitPoint, tr.end, tr.state)
		tr.end = splitPoint - 1
		tr.state = newSt

		return newSt, false
	}

	if s == t.auxState || s.findTran(t.data, c) != nil {
		return s, true
	}

	return s, false
}

// canonize returns updated state and start position for ref. pair
// (s, (start, end)) of state r so the new ref. pair is canonical,
// that is, referenced from the closest explicit ancestor of r.
//
//nolint:funcorder
func (t *STree) canonize(s *state, start, end Pos) (*state, Pos, error) {
	if s == t.auxState {
		s, start = t.root, start+1
	}

	if start > end {
		return s, start, nil
	}

	var tr *tran

	for {
		if start <= end {
			tr = s.findTran(t.data, t.data[start])
			if tr == nil {
				return nil, 0, errors.NewInternalError(
					fmt.Sprintf("no transition for token '%d' at position %d",
						t.data[start], start), nil,
				)
			}
		}

		if tr.end-tr.start > end-start {
			break
		}

		start += tr.end - tr.start + 1
		s = tr.state
	}

	if s == nil {
		return nil, 0, errors.NewInternalError(
			fmt.Sprintf("no suffix link resolution found: start=%d, end=%d", start, end), nil,
		)
	}

	return s, start, nil
}

// At returns the TokenValue at position p, or 0 if p is out of range.
func (t *STree) At(p Pos) TokenValue {
	if p < 0 || p >= Pos(len(t.data)) {
		return 0
	}

	return t.data[p]
}

// state is an explicit state of the suffix tree.
//
// Field ordering: trans (most accessed during search and construction) first,
// linkState second. The struct is 32 bytes — a []tran slice header (24 bytes)
// plus one pointer. Transitions are stored BY VALUE in a slice sorted by key
// (the token at data[tr.start]), so most states need exactly one allocation
// (the slice backing array) instead of a map hash table plus one *tran per
// edge. Leaf states never receive a transition and keep a nil slice — zero
// allocations. Previously included a tree *STree back-pointer, removed in
// favor of passing data []TokenValue as a parameter.
type state struct {
	trans     []tran
	linkState *state
}

// addTran adds a transition to state s, keeping the slice sorted by key
// (the token at data[start]). The slice is allocated lazily on the first
// transition so leaf states never allocate.
func (t *STree) addTran(s *state, start, end Pos, r *state) {
	key := t.data[start]

	low, high := 0, len(s.trans)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if t.data[s.trans[mid].start] < key {
			low = mid + 1
		} else {
			high = mid
		}
	}

	s.trans = slices.Insert(s.trans, low, tran{start: start, end: end, state: r})
}

// fork creates a new branch from the state s.
func (t *STree) fork(s *state, i Pos) *state {
	r := t.newState()
	t.addTran(s, i, infinity, r)

	return r
}

// tran represents a state's transition.
type tran struct {
	start, end Pos
	state      *state
}

//nolint:funcorder
func (t *tran) len() int {
	return int(t.end - t.start + 1)
}

// DataLen returns the number of tokens stored in the tree.
func (t *STree) DataLen() int {
	return len(t.data)
}

// actEnd returns actual end position as consistent with
// the actual length of the data in the STree.
func (t *tran) actEnd(data []TokenValue) Pos {
	if t.end == infinity {
		return Pos(len(data)) - 1 // #nosec G115 -- Data size won't exceed MaxInt32
	}

	return t.end
}
