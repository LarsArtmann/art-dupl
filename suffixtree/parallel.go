package suffixtree

import (
	"context"
	"runtime"
	"sync"
)

// FindDuplOverParallel searches for duplicate sequences >= threshold length
// using multiple goroutines. It dispatches each root-level subtree to a
// separate worker goroutine, limited by workers. Because each suffix starts
// with exactly one first token, root-level subtrees are disjoint — parallel
// traversal produces the identical match set as the sequential FindDuplOver.
//
// The root node itself never emits a match (length=0 < threshold for any
// practical threshold), so there is no need to merge context lists across
// workers. Each worker independently emits its subtree's maximal repeats.
//
// workers <= 0 defaults to runtime.NumCPU().
func (t *STree) FindDuplOverParallel(ctx context.Context, threshold, workers int) <-chan Match {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	ch := make(chan Match)

	go func() {
		defer close(ch)

		t.parallelWalkRoot(ctx, threshold, workers, ch)
	}()

	return ch
}

// parallelWalkRoot dispatches each root transition's subtree to a worker
// goroutine. The semaphore limits concurrency to workers. Context cancellation
// stops dispatching new subtrees; already-running workers check ctx in
// walkTrans and return early. The root's transition slice is already sorted
// by key, so dispatch order is deterministic without an extra allocation.
func (t *STree) parallelWalkRoot(
	ctx context.Context,
	threshold, workers int,
	ch chan<- Match,
) {
	if len(t.root.trans) == 0 {
		return
	}

	var wg sync.WaitGroup

	sem := make(chan struct{}, workers)

	for i := range t.root.trans {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Wait()

			return
		}

		tr := &t.root.trans[i]

		wg.Add(1)

		go func(tr *tran) {
			defer wg.Done()
			defer func() { <-sem }()

			cl := walkTrans(ctx, t.data, tr, tr.len(), threshold, ch)
			releaseContextList(cl)
		}(tr)
	}

	wg.Wait()
}
