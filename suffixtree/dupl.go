package suffixtree

import (
	"context"
	"math"
	"slices"
	"sync"
)

type Match struct {
	Ps  []Pos
	Len Pos
}

// contextList collects positions grouped by the token value that precedes
// them. It is allocated per walkTrans call and pooled to avoid repeated
// map allocations on the hot search path.
type contextList struct {
	lists map[TokenValue][]Pos
}

// maxStackKeys is the threshold for stack-allocated key buffers in
// contextList.getAll map iteration. Maps with <= maxStackKeys entries use a
// fixed-size stack array instead of heap-allocating a slice. contextList
// maps are keyed by preceding token value and typically hold 1-5 entries, so
// 32 covers the common case.
const maxStackKeys = 32

// contextListPool reuses contextList structs across searches to eliminate
// per-call map allocations. Each walkTrans call acquires a contextList and
// releases it after the caller is done (either after cl.append or at the
// top-level goroutine). The map is cleared on release but retains its
// internal hash table capacity, avoiding growth allocations on reuse.
//
//nolint:gochecknoglobals // sync.Pool is inherently global
var contextListPool = sync.Pool{
	New: func() any {
		return &contextList{lists: make(map[TokenValue][]Pos, 4)}
	},
}

func acquireContextList() *contextList {
	//nolint:forcetypeassert // Pool.New always returns *contextList
	return contextListPool.Get().(*contextList)
}

// releaseContextList returns a contextList to the pool. The caller must not
// touch the contextList afterwards — another goroutine may acquire it at any
// moment.
//
// OWNERSHIP CONTRACT: []Pos slices transferred into another contextList via
// append survive this release. append copies the slice HEADERS (pointer,
// length, capacity) into the destination's map; clear(cl.lists) below removes
// only the map entries, never the backing arrays, which stay alive as long as
// the destination contextList references them. Guarded by
// TestContextListPoolSliceSurvival.
func releaseContextList(cl *contextList) {
	clear(cl.lists)
	contextListPool.Put(cl)
}

func (c *contextList) getAll() []Pos {
	var (
		stackBuf [maxStackKeys]TokenValue
		keys     []TokenValue
	)

	if n := len(c.lists); n <= maxStackKeys {
		idx := 0

		for k := range c.lists {
			stackBuf[idx] = k
			idx++
		}

		keys = stackBuf[:idx]
	} else {
		keys = make([]TokenValue, 0, n)
		for k := range c.lists {
			keys = append(keys, k)
		}
	}

	slices.Sort(keys)

	totalCap := 0
	for _, k := range keys {
		totalCap += len(c.lists[k])
	}

	ps := make([]Pos, 0, totalCap)
	for _, k := range keys {
		ps = append(ps, c.lists[k]...)
	}

	return ps
}

func (c *contextList) append(c2 *contextList) {
	for lc, positions := range c2.lists {
		if existing, ok := c.lists[lc]; ok {
			c.lists[lc] = append(existing, positions...)
		} else {
			c.lists[lc] = positions
		}
	}
}

// FindDuplOver find pairs of maximal duplicities over a threshold
// length. The context allows callers to cancel the walk early.
func (t *STree) FindDuplOver(ctx context.Context, threshold int) <-chan Match {
	auxTran := &tran{state: t.root}
	ch := make(chan Match)

	go func() {
		cl := walkTrans(ctx, t.data, auxTran, 0, threshold, ch)
		releaseContextList(cl)
		close(ch)
	}()

	return ch
}

func walkTrans(
	ctx context.Context,
	data []TokenValue,
	parent *tran,
	length, threshold int,
	ch chan<- Match,
) *contextList {
	if ctx.Err() != nil {
		return acquireContextList()
	}

	s := parent.state

	cl := acquireContextList()

	if len(s.trans) == 0 {
		start := max(
			Pos(0),
			parent.end+1-Pos(length),
		) // #nosec G115 -- Positions are valid in tree context

		c := TokenValue(0)
		if start > 0 && int(start-1) < len(data) {
			c = data[start-1]
		}

		cl.lists[c] = []Pos{start}

		return cl
	}

	for i := range s.trans {
		tr := &s.trans[i]
		ln := length + tr.len()

		cl2 := walkTrans(ctx, data, tr, ln, threshold, ch)
		if ln >= threshold {
			cl.append(cl2)
		}

		releaseContextList(cl2)
	}

	if length >= threshold && len(cl.lists) > 1 {
		if length <= math.MaxInt32 {
			m := Match{cl.getAll(), Pos(length)} // #nosec G115 -- Bounds checked above
			select {
			case ch <- m:
			case <-ctx.Done():
			}
		}
	}

	return cl
}
