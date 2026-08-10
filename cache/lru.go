package cache

import (
	"container/list"
	"sync"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// defaultMemoryEntries is the maximum number of AST node slices held in the
// in-memory LRU layer. This is intentionally larger than a typical hot-set of
// frequently-changed files, so that repeated runs (the common workflow) hit
// memory instead of re-deserializing gob files from disk.
const defaultMemoryEntries = 512

// lru is an in-memory least-recently-used cache for deserialized AST node
// slices. It sits on top of FileCache's on-disk store and eliminates redundant
// gob deserialization on hot paths — the biggest perf win for cache-heavy
// workflows where the same files are processed repeatedly.
//
// The stored []*syntax.Node is the canonical copy; Get returns a deep clone so
// callers can freely mutate the result without corrupting the cache.
//
// Concurrency: all methods are goroutine-safe via a single sync.Mutex.
type lru struct {
	mu       sync.Mutex
	max      int
	entries  map[string]*list.Element
	order    *list.List
	memHits  int64
}

type lruEntry struct {
	key  string
	data []*syntax.Node
}

func newLRU(max int) *lru {
	if max <= 0 {
		max = defaultMemoryEntries
	}

	return &lru{
		max:     max,
		entries: make(map[string]*list.Element, max),
		order:   list.New(),
	}
}

// get returns a deep clone of the cached nodes for key, or nil if absent.
// A hit moves the entry to the front (most-recently-used).
func (l *lru) get(key string) []*syntax.Node {
	l.mu.Lock()
	defer l.mu.Unlock()

	elem, ok := l.entries[key]
	if !ok {
		return nil
	}

	l.order.MoveToFront(elem)
	l.memHits++

	return cloneNodes(elem.Value.(*lruEntry).data)
}

// put stores nodes as the canonical copy for key. If the cache is full, the
// least-recently-used entry is evicted. The caller must not mutate nodes
// after calling put.
func (l *lru) put(key string, nodes []*syntax.Node) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, ok := l.entries[key]; ok {
		elem.Value.(*lruEntry).data = nodes
		l.order.MoveToFront(elem)

		return
	}

	elem := l.order.PushFront(&lruEntry{key: key, data: nodes})
	l.entries[key] = elem

	if l.order.Len() > l.max {
		oldest := l.order.Back()
		if oldest != nil {
			l.order.Remove(oldest)
			delete(l.entries, oldest.Value.(*lruEntry).key)
		}
	}
}

// remove deletes a single entry from the LRU, if present.
func (l *lru) remove(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, ok := l.entries[key]; ok {
		l.order.Remove(elem)
		delete(l.entries, key)
	}
}

// clear removes all entries from the LRU.
func (l *lru) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make(map[string]*list.Element, l.max)
	l.order = list.New()
}

// cloneNodes creates independent copies of all nodes (no shared pointers).
func cloneNodes(nodes []*syntax.Node) []*syntax.Node {
	result := make([]*syntax.Node, len(nodes))
	for i, node := range nodes {
		result[i] = node.Clone()
	}

	return result
}
