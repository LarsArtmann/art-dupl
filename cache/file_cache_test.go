package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// testNodes creates sample nodes for testing.
func testNodes() []*syntax.Node {
	return []*syntax.Node{
		{
			Type:     1,
			Pos:      0,
			End:      100,
			Owns:     50,
			Filename: "test.go",
			Children: []*syntax.Node{
				{Type: 2, Pos: 10, End: 50, Owns: 20, Filename: "test.go"},
			},
		},
	}
}

// TestNewFileCache tests cache creation with default and custom directories.
func TestNewFileCache(t *testing.T) {
	t.Run("default directory", func(t *testing.T) {
		fc := NewFileCache("")
		if fc == nil {
			t.Fatal("Expected non-nil FileCache")
		}

		testutil.AssertFieldValue(t, fc.cacheDir, DefaultCacheDir, "cacheDir")

		testutil.AssertFieldValue(t, fc.metadata.Version, CacheVersion, "Version")
	})

	t.Run("custom directory", func(t *testing.T) {
		customDir := t.TempDir()

		fc := NewFileCache(customDir)
		if fc == nil {
			t.Fatal("Expected non-nil FileCache")
		}

		if fc.cacheDir != customDir {
			t.Errorf("Expected cacheDir=%q, got %q", customDir, fc.cacheDir)
		}

		// Verify cache directory structure was created
		filesDir := filepath.Join(customDir, "files")
		if _, err := os.Stat(filesDir); os.IsNotExist(err) {
			t.Errorf("Expected files directory to be created at %q", filesDir)
		}
	})
}

// TestFileCache_Get_Set tests cache hit and miss behavior.
func TestFileCache_Get_Set(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)

	t.Run("cache miss", func(t *testing.T) {
		nodes, hit := fc.Get("nonexistent-hash")
		if hit {
			t.Error("Expected cache miss, got hit")
		}

		if nodes != nil {
			t.Error("Expected nil nodes for cache miss")
		}
	})

	t.Run("cache set and get", func(t *testing.T) {
		originalNodes := testNodes()
		contentHash := Key([]byte("test content"))

		// Set cache entry
		err := fc.Set(contentHash, originalNodes)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get cache entry
		retrievedNodes, hit := fc.Get(contentHash)
		if !hit {
			t.Error("Expected cache hit, got miss")
		}

		if retrievedNodes == nil {
			t.Fatal("Expected non-nil nodes")
		}

		// Verify content matches
		if len(retrievedNodes) != len(originalNodes) {
			t.Errorf("Expected %d nodes, got %d", len(originalNodes), len(retrievedNodes))
		}

		if retrievedNodes[0].Type != originalNodes[0].Type {
			t.Errorf("Expected Type=%d, got %d", originalNodes[0].Type, retrievedNodes[0].Type)
		}

		if retrievedNodes[0].Filename != originalNodes[0].Filename {
			t.Errorf(
				"Expected Filename=%q, got %q",
				originalNodes[0].Filename,
				retrievedNodes[0].Filename,
			)
		}
	})

	t.Run("multiple entries", func(t *testing.T) {
		hash1 := Key([]byte("content1"))
		hash2 := Key([]byte("content2"))

		primaryNodes := []*syntax.Node{testutil.CreateNodeWithPos(1, "file1.go", 0, 10)}
		secondaryNodes := []*syntax.Node{testutil.CreateNodeWithPos(2, "file2.go", 5, 15)}

		err := fc.Set(hash1, primaryNodes)
		if err != nil {
			t.Fatalf("Set hash1 failed: %v", err)
		}

		err = fc.Set(hash2, secondaryNodes)
		if err != nil {
			t.Fatalf("Set hash2 failed: %v", err)
		}

		retrieved1, hit1 := fc.Get(hash1)
		retrieved2, hit2 := fc.Get(hash2)

		if !hit1 || !hit2 {
			t.Error("Expected both entries to be cache hits")
		}

		if retrieved1[0].Type != primaryNodes[0].Type {
			t.Errorf("hash1: Expected Type=%d, got %d", primaryNodes[0].Type, retrieved1[0].Type)
		}

		if retrieved2[0].Type != secondaryNodes[0].Type {
			t.Errorf("hash2: Expected Type=%d, got %d", secondaryNodes[0].Type, retrieved2[0].Type)
		}
	})
}

// TestFileCache_Has tests existence checking.
func TestFileCache_Has(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)
	contentHash := Key([]byte("test content"))

	t.Run("not exists initially", func(t *testing.T) {
		if fc.Has(contentHash) {
			t.Error("Expected Has to return false for non-existent entry")
		}
	})

	t.Run("exists after set", func(t *testing.T) {
		nodes := testNodes()

		err := fc.Set(contentHash, nodes)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		if !fc.Has(contentHash) {
			t.Error("Expected Has to return true for existing entry")
		}
	})

	t.Run("not exists after remove", func(t *testing.T) {
		err := fc.Remove(contentHash)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if fc.Has(contentHash) {
			t.Error("Expected Has to return false after removal")
		}
	})
}

// TestFileCache_Remove tests entry removal.
func TestFileCache_Remove(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)
	contentHash := Key([]byte("test content"))

	t.Run("remove non-existent entry", func(t *testing.T) {
		err := fc.Remove("nonexistent-hash")
		if err != nil {
			t.Errorf("Expected no error for removing non-existent entry, got: %v", err)
		}
	})

	t.Run("remove existing entry", func(t *testing.T) {
		nodes := testNodes()

		err := fc.Set(contentHash, nodes)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		if !fc.Has(contentHash) {
			t.Fatal("Entry should exist before removal")
		}

		err = fc.Remove(contentHash)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if fc.Has(contentHash) {
			t.Error("Entry should not exist after removal")
		}
	})
}

// TestFileCache_Clear tests cache clearing.
func TestFileCache_Clear(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)

	// Add multiple entries
	hashes := []string{
		Key([]byte("content1")),
		Key([]byte("content2")),
		Key([]byte("content3")),
	}

	nodes := testNodes()
	for _, hash := range hashes {
		err := fc.Set(hash, nodes)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Verify entries exist
	stats := fc.Stats()
	testutil.AssertFieldValue(t, stats.Size, 3, "entries before clear")

	// Clear cache
	err := fc.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all entries removed
	for _, hash := range hashes {
		if fc.Has(hash) {
			t.Errorf("Entry %q should not exist after clear", hash)
		}
	}

	// Verify stats reset
	stats = fc.Stats()
	testutil.AssertFieldValue(t, stats.Size, 0, "entries after clear")
}

// assertCacheStats asserts that cache stats match expected values.
func assertCacheStats(t *testing.T, stats Stats, expectedHits, expectedMisses, expectedSize int64) {
	t.Helper()

	if stats.Hits != expectedHits {
		t.Errorf("Expected %d hits, got %d", expectedHits, stats.Hits)
	}

	if stats.Misses != expectedMisses {
		t.Errorf("Expected %d misses, got %d", expectedMisses, stats.Misses)
	}

	if stats.Size != int(expectedSize) {
		t.Errorf("Expected %d size, got %d", expectedSize, stats.Size)
	}
}

// TestFileCache_Stats tests statistics tracking.
func TestFileCache_Stats(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)

	t.Run("empty cache stats", func(t *testing.T) {
		stats := fc.Stats()
		assertCacheStats(t, stats, 0, 0, 0)
	})

	t.Run("stats after operations", func(t *testing.T) {
		contentHash := Key([]byte("test content"))
		nodes := testNodes()

		// Cache miss
		fc.Get("nonexistent")

		// Set entry
		err := fc.Set(contentHash, nodes)
		if err != nil {
			t.Fatalf("Failed to set cache entry: %v", err)
		}

		// Cache hit
		fc.Get(contentHash)

		stats := fc.Stats()
		assertCacheStats(t, stats, 1, 1, 1)

		if stats.BytesUsed <= 0 {
			t.Errorf("Expected positive BytesUsed, got %d", stats.BytesUsed)
		}
	})
}

// TestKey tests hash generation.
func TestKey(t *testing.T) {
	t.Run("consistent hashing", func(t *testing.T) {
		content := []byte("test content")
		hash1 := Key(content)
		hash2 := Key(content)

		if hash1 != hash2 {
			t.Errorf("Expected same hash for same content, got %q and %q", hash1, hash2)
		}
	})

	t.Run("different content different hash", func(t *testing.T) {
		hash1 := Key([]byte("content1"))
		hash2 := Key([]byte("content2"))

		if hash1 == hash2 {
			t.Errorf("Expected different hashes for different content")
		}
	})

	t.Run("hash length", func(t *testing.T) {
		hash := Key([]byte("test"))
		// SHA-256 produces 32 bytes, hex encoded to 64 characters
		if len(hash) != 64 {
			t.Errorf("Expected hash length 64, got %d", len(hash))
		}
	})

	t.Run("empty content", func(t *testing.T) {
		hash := Key([]byte{})
		if hash == "" {
			t.Error("Expected non-empty hash for empty content")
		}
	})
}

// TestFileCache_Concurrency tests concurrent access safety.
func TestFileCache_Concurrency(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)

	const numOps = 100

	done := make(chan bool, numOps*3)

	// Concurrent writes
	for i := range numOps {
		go func(i int) {
			hash := Key([]byte{byte(i)})

			nodes := []*syntax.Node{{Type: int32(i), Filename: "test.go"}}

			err := fc.Set(hash, nodes)
			if err != nil {
				t.Errorf("Failed to set cache entry: %v", err)
			}

			done <- true
		}(i)
	}

	// Concurrent reads
	for i := range numOps {
		go func(i int) {
			hash := Key([]byte{byte(i)})
			fc.Get(hash)

			done <- true
		}(i)
	}

	// Concurrent Has checks
	for i := range numOps {
		go func(i int) {
			hash := Key([]byte{byte(i)})
			fc.Has(hash)

			done <- true
		}(i)
	}

	// Wait for all operations
	for range numOps * 3 {
		<-done
	}
}

// TestFileCache_EmptyNodes tests handling of empty node slices.
func TestFileCache_EmptyNodes(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)
	contentHash := Key([]byte("test content"))

	t.Run("set and get empty slice", func(t *testing.T) {
		err := fc.Set(contentHash, []*syntax.Node{})
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		nodes, hit := fc.Get(contentHash)
		if !hit {
			t.Fatal("Expected cache hit")
		}

		if len(nodes) != 0 {
			t.Errorf("Expected 0 nodes, got %d", len(nodes))
		}
	})
}

// TestFileCache_NestedNodes tests handling of nested node structures.
func TestFileCache_NestedNodes(t *testing.T) {
	const parentFile = "parent.go"

	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)
	contentHash := Key([]byte("nested test"))

	originalNodes := []*syntax.Node{
		{
			Type:     1,
			Pos:      0,
			End:      100,
			Owns:     50,
			Filename: parentFile,
			Children: []*syntax.Node{
				{
					Type:     2,
					Pos:      10,
					End:      50,
					Owns:     20,
					Filename: parentFile,
					Children: []*syntax.Node{
						{Type: 3, Pos: 15, End: 25, Filename: parentFile},
						{Type: 4, Pos: 30, End: 45, Filename: parentFile},
					},
				},
			},
		},
	}

	err := fc.Set(contentHash, originalNodes)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrievedNodes, hit := fc.Get(contentHash)
	if !hit {
		t.Fatal("Expected cache hit")
	}

	// Verify nested structure
	if len(retrievedNodes[0].Children) == 0 {
		t.Fatal("Expected children to be preserved")
	}

	if len(retrievedNodes[0].Children[0].Children) != 2 {
		t.Errorf("Expected 2 grandchildren, got %d", len(retrievedNodes[0].Children[0].Children))
	}

	if retrievedNodes[0].Children[0].Children[0].Type != 3 {
		t.Errorf(
			"Expected grandchild Type=3, got %d",
			retrievedNodes[0].Children[0].Children[0].Type,
		)
	}
}

// TestFileCache_Persistence tests that cache persists across instances.
func TestFileCache_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	contentHash := Key([]byte("persistent test"))
	nodes := testNodes()

	// Create first cache instance and set data
	fc1 := NewFileCache(tempDir)

	err := fc1.Set(contentHash, nodes)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Create second cache instance with same directory
	fc2 := NewFileCache(tempDir)

	// Verify data persists
	retrievedNodes, hit := fc2.Get(contentHash)
	if !hit {
		t.Fatal("Expected cache hit in second instance")
	}

	if len(retrievedNodes) != len(nodes) {
		t.Errorf("Expected %d nodes, got %d", len(nodes), len(retrievedNodes))
	}
}

// TestFileCache_Prune tests the Prune function for cache eviction.
func TestFileCache_Prune(t *testing.T) {
	t.Run("zero_max_no_op", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())

		hash := Key([]byte("a"))
		if err := fc.Set(hash, testNodes()); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		evicted, err := fc.Prune(0)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted != 0 {
			t.Errorf("Expected 0 evicted, got %d", evicted)
		}

		if _, hit := fc.Get(hash); !hit {
			t.Error("Entry should still exist after Prune(0)")
		}
	})

	t.Run("fewer_than_max_no_eviction", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		for _, content := range []string{"a", "b", "c"} {
			if err := fc.Set(Key([]byte(content)), testNodes()); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		evicted, err := fc.Prune(10)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted != 0 {
			t.Errorf("Expected 0 evicted, got %d", evicted)
		}
	})

	t.Run("evicts_oldest", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		// Insert 5 entries, each with a small delay to get distinct mtimes.
		hashes := make([]string, 5)
		for i, content := range []string{"a", "b", "c", "d", "e"} {
			hashes[i] = Key([]byte(content))
			if err := fc.Set(hashes[i], testNodes()); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
			// Touch the file to ensure distinct mtimes (some filesystems have low resolution).
			path := filepath.Join(fc.cacheDir, "files", hashes[i]+".gob")

			modTime := time.Now().Add(time.Duration(i) * 100 * time.Millisecond)
			if err := os.Chtimes(path, modTime, modTime); err != nil {
				t.Fatalf("Chtimes failed: %v", err)
			}
		}

		// Prune to 3 entries — should evict the 2 oldest (a, b).
		evicted, err := fc.Prune(3)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted != 2 {
			t.Fatalf("Expected 2 evicted, got %d", evicted)
		}

		// Oldest entries should be gone.
		if _, hit := fc.Get(hashes[0]); hit {
			t.Error("Oldest entry 'a' should have been evicted")
		}

		if _, hit := fc.Get(hashes[1]); hit {
			t.Error("Second-oldest entry 'b' should have been evicted")
		}

		// Newer entries should remain.
		for i := 2; i < 5; i++ {
			if _, hit := fc.Get(hashes[i]); !hit {
				t.Errorf("Entry %q should still exist after prune", string(rune('a'+i)))
			}
		}
	})

	t.Run("empty_cache_no_error", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())

		evicted, err := fc.Prune(5)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted != 0 {
			t.Errorf("Expected 0 evicted on empty cache, got %d", evicted)
		}
	})
}

// TestKeyWithParams tests that the params-aware cache key produces different
// keys for the same content with different params, preventing cross-mode cache contamination.
func TestKeyWithParams(t *testing.T) {
	content := []byte("package main")

	t.Run("different_params_produce_different_keys", func(t *testing.T) {
		keyA := KeyWithParams(content, "semantic:5:")

		keyB := KeyWithParams(content, "exact:5:")
		if keyA == keyB {
			t.Fatal("Expected different keys for different params")
		}
	})

	t.Run("same_params_produce_same_key", func(t *testing.T) {
		keyA := KeyWithParams(content, "semantic:5:ta")

		keyB := KeyWithParams(content, "semantic:5:ta")
		if keyA != keyB {
			t.Fatal("Expected identical keys for same params")
		}
	})

	t.Run("empty_params_equals_Key", func(t *testing.T) {
		// KeyWithParams with empty params should equal Key — empty string writes
		// 0 extra bytes into the hash stream, so the input is identical.
		keyNormal := Key(content)

		keyWithEmptyParams := KeyWithParams(content, "")
		if keyNormal != keyWithEmptyParams {
			t.Fatalf("Expected identical keys for empty params, got %q vs %q", keyNormal, keyWithEmptyParams)
		}
	})

	t.Run("returns_64_char_hex", func(t *testing.T) {
		key := KeyWithParams(content, "semantic:5:ta")
		if len(key) != 64 {
			t.Errorf("Expected 64-char hex hash, got %d chars", len(key))
		}
	})

	t.Run("length_prefix_prevents_concatenation_collision", func(t *testing.T) {
		// Without a length prefix, content="ab"+params="cd" hashes the same
		// stream as content="a"+params="bcd". The length prefix prevents this.
		key1 := KeyWithParams([]byte("ab"), "cd")

		key2 := KeyWithParams([]byte("a"), "bcd")
		if key1 == key2 {
			t.Fatal("Expected different keys for different content/params splits")
		}
	})
}

// TestFileCache_ErrorPaths tests Get behavior with corrupt, stale, and
// malformed cache files. In all cases Get must return a miss and remove the
// offending file so it doesn't waste disk or cause repeated decode failures.
func TestFileCache_ErrorPaths(t *testing.T) {
	writeCacheFile := func(t *testing.T, fc *FileCache, hash string, data []byte) string {
		t.Helper()

		path := filepath.Join(fc.cacheDir, "files", hash+".gob")
		if err := os.WriteFile(path, data, cacheFilePerms); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		return path
	}

	assertFileRemoved := func(t *testing.T, path string) {
		t.Helper()

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected file %q to be removed after Get, but it still exists", path)
		}
	}

	encodeEntry := func(t *testing.T, entry *cacheEntry) []byte {
		t.Helper()

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(entry); err != nil {
			t.Fatalf("Gob encode failed: %v", err)
		}

		return buf.Bytes()
	}

	t.Run("corrupt_gob_removed_on_get", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		hash := Key([]byte("corrupt"))
		path := writeCacheFile(t, fc, hash, []byte("not valid gob data at all"))

		nodes, hit := fc.Get(hash)
		if hit {
			t.Fatal("Expected cache miss for corrupt data")
		}

		if nodes != nil {
			t.Fatal("Expected nil nodes for corrupt data")
		}

		assertFileRemoved(t, path)
	})

	t.Run("truncated_gob_removed_on_get", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		hash := Key([]byte("truncated"))
		full := encodeEntry(t, &cacheEntry{Version: CacheVersion, Nodes: testNodes()})
		path := writeCacheFile(t, fc, hash, full[:len(full)/2])

		_, hit := fc.Get(hash)
		if hit {
			t.Fatal("Expected cache miss for truncated gob")
		}

		assertFileRemoved(t, path)
	})

	t.Run("version_mismatch_removed_on_get", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		hash := Key([]byte("stale-version"))
		data := encodeEntry(t, &cacheEntry{Version: CacheVersion + 999, Nodes: testNodes()})
		path := writeCacheFile(t, fc, hash, data)

		_, hit := fc.Get(hash)
		if hit {
			t.Fatal("Expected cache miss for version mismatch")
		}

		assertFileRemoved(t, path)
	})

	t.Run("empty_file_removed_on_get", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		hash := Key([]byte("empty-file"))
		path := writeCacheFile(t, fc, hash, []byte{})

		_, hit := fc.Get(hash)
		if hit {
			t.Fatal("Expected cache miss for empty file")
		}

		assertFileRemoved(t, path)
	})

	t.Run("miss_does_not_affect_subsequent_set", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())
		hash := Key([]byte("recover"))
		path := writeCacheFile(t, fc, hash, []byte("garbage"))

		// Corrupt entry is auto-removed on Get miss
		_, hit := fc.Get(hash)
		if hit {
			t.Fatal("Expected miss for garbage")
		}

		assertFileRemoved(t, path)

		// After removal, Set should succeed and Get should hit
		if err := fc.Set(hash, testNodes()); err != nil {
			t.Fatalf("Set after corrupt-removal failed: %v", err)
		}

		retrieved, hit := fc.Get(hash)
		if !hit {
			t.Fatal("Expected hit after Set recovery")
		}

		if len(retrieved) != 1 {
			t.Errorf("Expected 1 node, got %d", len(retrieved))
		}
	})
}

// TestFileCache_ConcurrentPruneAndSet verifies that Prune and Set can run
// concurrently without data races or panics. Run with -race to verify.
func TestFileCache_ConcurrentPruneAndSet(t *testing.T) {
	fc := NewFileCache(t.TempDir())
	nodes := testNodes()

	// Pre-populate with some entries so Prune has work to do.
	for i := range 20 {
		if err := fc.Set(Key(fmt.Appendf(nil, "seed-%d", i)), nodes); err != nil {
			t.Fatalf("Seed Set failed: %v", err)
		}
	}

	var wg sync.WaitGroup

	done := make(chan struct{})

	// Writer goroutine: continuously Set new entries.
	wg.Go(func() {
		i := 0

		for {
			select {
			case <-done:
				return
			default:
			}

			hash := Key(fmt.Appendf(nil, "concurrent-%d", i))
			_ = fc.Set(hash, nodes)
			fc.Get(hash)

			i++
		}
	})

	// Pruner goroutine: continuously prune to a small max.
	wg.Go(func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			_, _ = fc.Prune(10)
		}
	})

	// Stats reader goroutine: concurrently read stats.
	wg.Go(func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			_ = fc.Stats()
		}
	})

	// Let them run for a short burst.
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}

// TestFileCache_LRUHitAfterSet verifies that a Get immediately following a Set
// hits the in-memory LRU without needing to read from disk.
func TestFileCache_LRUHitAfterSet(t *testing.T) {
	fc := NewFileCache(t.TempDir())
	hash := Key([]byte("lru-test"))
	original := testNodes()

	if err := fc.Set(hash, original); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Remove the on-disk file so only the LRU has the data.
	diskPath := filepath.Join(fc.cacheDir, "files", hash+".gob")
	if err := os.Remove(diskPath); err != nil {
		t.Fatalf("Failed to remove disk cache file: %v", err)
	}

	// Get should still succeed via the in-memory LRU.
	nodes, hit := fc.Get(hash)
	if !hit {
		t.Fatal("Expected cache hit from LRU after disk file removed")
	}

	if len(nodes) != len(original) {
		t.Errorf("Expected %d nodes, got %d", len(original), len(nodes))
	}
}

// TestFileCache_LRUReturnsDeepClone verifies that Get returns independent
// copies so callers cannot corrupt the cached canonical nodes.
func TestFileCache_LRUReturnsDeepClone(t *testing.T) {
	fc := NewFileCache(t.TempDir())
	hash := Key([]byte("clone-test"))

	if err := fc.Set(hash, testNodes()); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	first, _ := fc.Get(hash)
	second, _ := fc.Get(hash)

	// Mutate the first result.
	if len(first) > 0 {
		first[0].Type = 999
	}

	// Second result must be unaffected.
	if len(second) > 0 && second[0].Type == 999 {
		t.Error("Get returned a shared reference — mutating one result corrupted another")
	}
}

// TestFileCache_LHysteresisPruning verifies that Prune only triggers above 110%
// of maxEntries and evicts down to 90% of maxEntries.
func TestFileCache_HysteresisPruning(t *testing.T) {
	t.Run("no_prune_below_high_water", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())

		// maxEntries=10, highWater=11. Insert 11 entries (exactly at high water).
		for i := range 11 {
			if err := fc.Set(Key([]byte{byte(i)}), testNodes()); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		}

		evicted, err := fc.Prune(10)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted != 0 {
			t.Errorf("Expected 0 evicted at high water mark (11 <= 11), got %d", evicted)
		}
	})

	t.Run("prune_above_high_water_to_low_water", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())

		// maxEntries=10, highWater=11, lowWater=9.
		// Insert 15 entries (> 11) to trigger pruning.
		hashes := make([]string, 15)
		for i := range 15 {
			hashes[i] = Key([]byte{byte(i)})

			if err := fc.Set(hashes[i], testNodes()); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			// Ensure distinct mtimes for deterministic eviction order.
			path := filepath.Join(fc.cacheDir, "files", hashes[i]+".gob")
			modTime := time.Now().Add(time.Duration(i) * 100 * time.Millisecond)

			if err := os.Chtimes(path, modTime, modTime); err != nil {
				t.Fatalf("Chtimes failed: %v", err)
			}
		}

		evicted, err := fc.Prune(10)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		// Should evict 15 - 9 = 6 entries (down to low water = 9).
		if evicted != 6 {
			t.Errorf("Expected 6 evicted (15 → 9), got %d", evicted)
		}

		// Verify remaining count.
		stats := fc.Stats()
		if stats.Size != 9 {
			t.Errorf("Expected 9 entries after prune, got %d", stats.Size)
		}
	})

	t.Run("prune_evicts_from_lru", func(t *testing.T) {
		fc := NewFileCache(t.TempDir())

		// maxEntries=2, highWater=2, lowWater=1.
		// Insert 3 entries to trigger pruning (> 2).
		hashes := []string{
			Key([]byte("x")),
			Key([]byte("y")),
			Key([]byte("z")),
		}
		for i, h := range hashes {
			if err := fc.Set(h, testNodes()); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			path := filepath.Join(fc.cacheDir, "files", h+".gob")
			modTime := time.Now().Add(time.Duration(i) * 100 * time.Millisecond)

			if err := os.Chtimes(path, modTime, modTime); err != nil {
				t.Fatalf("Chtimes failed: %v", err)
			}
		}

		evicted, err := fc.Prune(2)
		if err != nil {
			t.Fatalf("Prune error: %v", err)
		}

		if evicted == 0 {
			t.Fatal("Expected some evictions")
		}

		// Evicted entries should not be in the LRU. Remove the disk file for
		// an evicted entry and verify Get returns a miss (not an LRU hit).
		stats := fc.Stats()
		remaining := stats.Size

		// Remove all disk files and check that Get only succeeds for entries
		// still in the LRU (if any). Evicted entries should be gone from both.
		filesDir := filepath.Join(fc.cacheDir, "files")
		entries, _ := os.ReadDir(filesDir)

		for _, entry := range entries {
			_ = os.Remove(filepath.Join(filesDir, entry.Name()))
		}

		hits := 0

		for _, h := range hashes {
			if _, hit := fc.Get(h); hit {
				hits++
			}
		}

		// Only remaining disk entries could have been in the LRU (from the Set
		// that wrote them). After disk removal, only LRU entries survive.
		// Evicted entries should NOT be in the LRU.
		if hits > remaining {
			t.Errorf("Expected at most %d LRU hits (remaining disk entries), got %d", remaining, hits)
		}
	})
}
