// Package suffixtree implements suffix tree for clone detection.
// Core types: STree (tree), state (node), tran (edge), Match (duplicate), Token (interface).
// Algorithm: Build incrementally, FindDuplOver() searches for sequences >= threshold.
// Performance: O(n*m) worst case, typically O(n). Memory: O(n*k).
// Optimization: Map-based transition lookup for O(1) findTran performance.
package suffixtree

import (
	"bytes"
	"fmt"
	"math"
	"strings"

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

// STree is a struct representing a suffix tree.
type STree struct {
	data     []Token
	root     *state
	auxState *state // auxiliary state

	// active point
	s          *state
	start, end Pos
}

// New creates new suffix tree.
func New() *STree {
	t := new(STree)
	t.data = make([]Token, 0, 50)
	t.root = newState(t)
	t.auxState = newState(t)
	t.root.linkState = t.auxState
	t.s = t.root

	return t
}

// Update refreshes the suffix tree to by new data.
func (t *STree) Update(data ...Token) {
	t.data = append(t.data, data...)
	for range data {
		t.update()
		t.s, t.start, _ = t.canonize(t.s, t.start, t.end)
		t.end++
	}
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

		r.fork(end)

		if oldr != t.root {
			oldr.linkState = r
		}

		oldr = r
		s, start, _ = t.canonize(s.linkState, start, end-1)
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
//nolint:funcorder
func (t *STree) testAndSplit(s *state, start, end Pos) (exs *state, endPoint bool) {
	c := t.data[t.end]
	if start <= end {
		tr := s.findTran(t.data[start])

		splitPoint := tr.start + end - start + 1
		if t.data[splitPoint].Val() == c.Val() {
			return s, true
		}
		// make the (s, (start, end)) state explicit
		newSt := newState(s.tree)
		newSt.addTran(splitPoint, tr.end, tr.state)
		tr.end = splitPoint - 1
		tr.state = newSt

		return newSt, false
	}

	if s == t.auxState || s.findTran(c) != nil {
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
			tr = s.findTran(t.data[start])
			if tr == nil {
				return nil, 0, errors.NewInternalError(
					fmt.Sprintf("no transition for token '%d' at position %d",
						t.data[start].Val(), start), nil)
			}
		}

		if tr.end-tr.start > end-start {
			break
		}

		start += tr.end - tr.start + 1
		s = tr.state
	}

	if s == nil {
		return nil, 0, errors.NewInternalError("no suffix link resolution found", nil)
	}

	return s, start, nil
}

func (t *STree) At(p Pos) Token {
	// Safe conversion: len(t.data) will not overflow Pos in practice
	// #nosec G115 -- Data size won't exceed MaxInt32
	if p < 0 || p >= Pos(len(t.data)) {
		return nil
	}

	return t.data[p]
}

func (t *STree) String() string {
	buf := new(bytes.Buffer)
	printState(buf, t.root, 0)

	return buf.String()
}

func printState(buf *bytes.Buffer, s *state, ident int) {
	for _, tr := range s.trans {
		fmt.Fprint(buf, strings.Repeat("  ", ident))
		fmt.Fprintf(buf, "* (%d, %d)\n", tr.start, tr.ActEnd())
		printState(buf, tr.state, ident+1)
	}
}

// state is an explicit state of the suffix tree.
type state struct {
	tree      *STree
	trans     map[TokenValue]*tran
	linkState *state
}

func newState(t *STree) *state {
	return &state{
		tree:      t,
		trans:     make(map[TokenValue]*tran),
		linkState: nil,
	}
}

func (s *state) addTran(start, end Pos, r *state) {
	// Key is the token value at the transition's start position
	key := s.tree.data[start].Val()
	s.trans[key] = newTran(start, end, r)
}

// fork creates a new branch from the state s.
func (s *state) fork(i Pos) *state {
	r := newState(s.tree)
	s.addTran(i, infinity, r)

	return r
}

// tran represents a state's transition.
type tran struct {
	start, end Pos
	state      *state
}

func newTran(start, end Pos, s *state) *tran {
	return &tran{start, end, s}
}

//nolint:funcorder
func (t *tran) len() int {
	return int(t.end - t.start + 1)
}

// ActEnd returns actual end position as consistent with
// the actual length of the data in the STree.
func (t *tran) ActEnd() Pos {
	if t.end == infinity {
		// Safe conversion: tree data size won't overflow Pos in practice
		return Pos(len(t.state.tree.data)) - 1 // #nosec G115 -- Data size won't exceed MaxInt32
	}

	return t.end
}
