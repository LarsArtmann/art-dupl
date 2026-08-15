package cache

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestStats_IncludesMemHits(t *testing.T) {
	t.Parallel()

	fc := NewFileCache(t.TempDir())

	if fc.Stats().MemHits != 0 {
		t.Fatalf("expected 0 MemHits on fresh cache, got %d", fc.Stats().MemHits)
	}

	nodes := []*syntax.Node{{Type: 1, Name: "f"}}

	if err := fc.Set("k1", nodes); err != nil {
		t.Fatal(err)
	}

	if _, ok := fc.Get("k1"); !ok {
		t.Fatal("expected hit after Set")
	}

	if _, ok := fc.Get("k1"); !ok {
		t.Fatal("expected second hit")
	}

	stats := fc.Stats()
	if stats.MemHits < 2 {
		t.Errorf("expected MemHits >= 2 (both Gets served from memory), got %d", stats.MemHits)
	}
}

func TestNewFileCacheWithMemoryEntries_EvictsWhenCapacityExceeded(t *testing.T) {
	t.Parallel()

	fc := NewFileCacheWithMemoryEntries(t.TempDir(), 1)

	nodes := []*syntax.Node{{Type: 1, Name: "a"}}

	if err := fc.Set("first", nodes); err != nil {
		t.Fatal(err)
	}

	if err := fc.Set("second", nodes); err != nil {
		t.Fatal(err)
	}

	if _, ok := fc.Get("first"); !ok {
		t.Error("first entry must still be retrievable from disk after LRU eviction")
	}

	if got := fc.Stats().MemHits; got > 1 {
		t.Errorf("capacity-1 LRU should serve at most 1 in-memory hit, got %d", got)
	}
}
