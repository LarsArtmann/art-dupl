package suffixtree

import (
	"context"
	"math"
	"slices"
)

type Match struct {
	Ps  []Pos
	Len Pos
}

type posList struct {
	positions []Pos
}

func newPosList() *posList {
	return &posList{make([]Pos, 0)}
}

func (p *posList) append(p2 *posList) {
	p.positions = append(p.positions, p2.positions...)
}

func (p *posList) add(pos Pos) {
	p.positions = append(p.positions, pos)
}

type contextList struct {
	lists map[TokenValue]*posList
}

func newContextList() *contextList {
	return &contextList{make(map[TokenValue]*posList)}
}

func (c *contextList) getAll() []Pos {
	keys := make([]TokenValue, 0, len(c.lists))
	for k := range c.lists {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	// Calculate total capacity for all positions
	totalCap := 0
	for _, k := range keys {
		totalCap += len(c.lists[k].positions)
	}

	ps := make([]Pos, 0, totalCap)
	for _, k := range keys {
		ps = append(ps, c.lists[k].positions...)
	}

	return ps
}

func (c *contextList) append(c2 *contextList) {
	for lc, pl := range c2.lists {
		if _, ok := c.lists[lc]; ok {
			c.lists[lc].append(pl)
		} else {
			c.lists[lc] = pl
		}
	}
}

// FindDuplOver find pairs of maximal duplicities over a threshold
// length. The context allows callers to cancel the walk early.
func (t *STree) FindDuplOver(ctx context.Context, threshold int) <-chan Match {
	auxTran := newTran(0, 0, t.root)
	ch := make(chan Match)

	go func() {
		walkTrans(ctx, auxTran, 0, threshold, ch)
		close(ch)
	}()

	return ch
}

func walkTrans(
	ctx context.Context,
	parent *tran,
	length, threshold int,
	ch chan<- Match,
) *contextList {
	if ctx.Err() != nil {
		return newContextList()
	}

	s := parent.state

	cl := newContextList()

	if len(s.trans) == 0 {
		pl := newPosList()
		// Safe conversion: ensure start is non-negative
		start := max(
			Pos(0),
			parent.end+1-Pos(length),
		) // #nosec G115 -- Positions are valid in tree context
		pl.add(start)

		ch := TokenValue(0)
		// Bounds check: ensure start-1 is within data slice bounds
		if start > 0 && int(start-1) < len(s.tree.data) {
			ch = s.tree.data[start-1]
		}

		cl.lists[ch] = pl

		return cl
	}

	// Sort transitions by token value for deterministic iteration order
	transKeys := make([]TokenValue, 0, len(s.trans))
	for k := range s.trans {
		transKeys = append(transKeys, k)
	}

	slices.Sort(transKeys)

	for _, k := range transKeys {
		t := s.trans[k]
		ln := length + t.len()

		cl2 := walkTrans(ctx, t, ln, threshold, ch)
		if ln >= threshold {
			cl.append(cl2)
		}
	}

	if length >= threshold && len(cl.lists) > 1 {
		// Safe conversion: ensure length fits in int32
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
