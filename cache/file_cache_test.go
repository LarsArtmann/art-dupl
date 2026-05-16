package cache

import (
	"os"
	"path/filepath"
	"testing"

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

// TestFileCache_GetStats tests the GetStats method.
func TestFileCache_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)

	t.Run("initial stats", func(t *testing.T) {
		hits, misses := fc.GetStats()
		if hits != 0 || misses != 0 {
			t.Errorf("Expected (0, 0), got (%d, %d)", hits, misses)
		}
	})

	t.Run("after operations", func(t *testing.T) {
		contentHash := Key([]byte("test content"))
		nodes := testNodes()

		fc.Get("nonexistent") // miss

		err := fc.Set(contentHash, nodes)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		fc.Get(contentHash) // hit

		hits, misses := fc.GetStats()
		if hits != 1 {
			t.Errorf("Expected 1 hit, got %d", hits)
		}

		if misses != 1 {
			t.Errorf("Expected 1 miss, got %d", misses)
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
		// SHA1 produces 20 bytes, hex encoded to 40 characters
		if len(hash) != 40 {
			t.Errorf("Expected hash length 40, got %d", len(hash))
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
	tempDir := t.TempDir()
	fc := NewFileCache(tempDir)
	contentHash := Key([]byte("nested test"))

	originalNodes := []*syntax.Node{
		{
			Type:     1,
			Pos:      0,
			End:      100,
			Owns:     50,
			Filename: "parent.go",
			Children: []*syntax.Node{
				{
					Type:     2,
					Pos:      10,
					End:      50,
					Owns:     20,
					Filename: "parent.go",
					Children: []*syntax.Node{
						{Type: 3, Pos: 15, End: 25, Filename: "parent.go"},
						{Type: 4, Pos: 30, End: 45, Filename: "parent.go"},
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
