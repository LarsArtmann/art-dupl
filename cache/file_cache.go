// Package cache provides file-based caching for parsed AST nodes.
//
// This package enables incremental duplicate detection by caching the parsed
// AST nodes for each file. When a file hasn't changed (same content hash),
// its cached AST can be reused instead of re-parsing.
//
// Cache Structure:
//
//	.cache/art-dupl/
//	├── files/
//	│   ├── <hash1>.gob    # Serialized AST nodes for file with hash1
//	│   ├── <hash2>.gob    # Serialized AST nodes for file with hash2
//	│   └── ...
//	└── metadata.json      # Cache metadata (version, timestamps)
//
// Usage:
//
//	cache := cache.NewFileCache(".cache/art-dupl")
//	nodes, hit := cache.Get(fileHash)
//	if !hit {
//	    nodes = parseFile(path)
//	    cache.Set(fileHash, nodes)
//	}
package cache

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- SHA1 used for cache keys, not security
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

const (
	// DefaultCacheDir is the default directory name for the cache.
	DefaultCacheDir = ".cache/art-dupl"

	// CacheVersion is incremented when cache format changes.
	CacheVersion = 1

	// Directory permissions.
	cacheDirPerms = 0o750
	// File permissions.
	cacheFilePerms = 0o600
)

// FileCache provides caching for parsed AST nodes.
type FileCache struct {
	mu       sync.RWMutex
	cacheDir string
	metadata Metadata
}

// Metadata contains cache metadata.
type Metadata struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	HitCount  int64     `json:"hitCount"`  // Accessed atomically
	MissCount int64     `json:"missCount"` // Accessed atomically
}

// Stats returns cache statistics.
type Stats struct {
	Hits      int64
	Misses    int64
	Size      int // Number of cached entries
	BytesUsed int64
}

// NewFileCache creates a new FileCache.
// If cacheDir is empty, uses DefaultCacheDir.
func NewFileCache(cacheDir string) *FileCache {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir
	}

	fc := &FileCache{
		cacheDir: cacheDir,
		metadata: newMetadata(),
	}

	// Ensure cache directories exist (tests verify this)
	filesDir := filepath.Join(cacheDir, "files")
	_ = os.MkdirAll(filesDir, cacheDirPerms)

	// Load existing metadata if available
	fc.loadMetadata()

	return fc
}

// newMetadata creates a new Metadata with current timestamps.
func newMetadata() Metadata {
	return Metadata{
		Version:   CacheVersion,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Get retrieves cached AST nodes for the given content hash.
// Returns the nodes and true if found (cache hit), nil and false otherwise.
func (fc *FileCache) Get(contentHash string) ([]*syntax.Node, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	cachePath := fc.cachePath(contentHash)

	// #nosec G304 -- Path constructed from controlled cache directory and content hash
	data, err := os.ReadFile(cachePath)
	if err != nil {
		atomic.AddInt64(&fc.metadata.MissCount, 1)

		return nil, false
	}

	nodes, err := fc.deserialize(data)
	if err != nil {
		// Corrupted cache entry, remove it
		_ = os.Remove(cachePath)

		atomic.AddInt64(&fc.metadata.MissCount, 1)

		return nil, false
	}

	atomic.AddInt64(&fc.metadata.HitCount, 1)

	return nodes, true
}

// Set stores AST nodes for the given content hash.
func (fc *FileCache) Set(contentHash string, nodes []*syntax.Node) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Ensure cache directory exists
	filesDir := filepath.Join(fc.cacheDir, "files")

	err := os.MkdirAll(filesDir, cacheDirPerms)
	if err != nil {
		return errors.NewIOError(filesDir, "failed to create cache directory", err)
	}

	data, err := fc.serialize(nodes)
	if err != nil {
		return err
	}

	cachePath := fc.cachePath(contentHash)

	err = os.WriteFile(cachePath, data, cacheFilePerms)
	if err != nil {
		return errors.NewIOError(cachePath, "failed to write cache file", err)
	}

	fc.metadata.UpdatedAt = time.Now()
	_ = fc.saveMetadata()

	return nil
}

// Has checks if a cache entry exists for the given hash.
func (fc *FileCache) Has(contentHash string) bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	cachePath := fc.cachePath(contentHash)
	_, err := os.Stat(cachePath)

	return err == nil
}

// Remove deletes a cache entry.
func (fc *FileCache) Remove(contentHash string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	cachePath := fc.cachePath(contentHash)

	err := os.Remove(cachePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.NewIOError(cachePath, "failed to remove cache entry", err)
	}

	return nil
}

// Clear removes all cache entries.
func (fc *FileCache) Clear() error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	filesDir := filepath.Join(fc.cacheDir, "files")

	err := os.RemoveAll(filesDir)
	if err != nil {
		return errors.NewIOError(filesDir, "failed to clear cache", err)
	}

	// Recreate the files directory so subsequent Set() calls work
	err = os.MkdirAll(filesDir, cacheDirPerms)
	if err != nil {
		return errors.NewIOError(filesDir, "failed to recreate cache directory", err)
	}

	fc.metadata = newMetadata()

	return nil
}

// Stats returns cache statistics.
func (fc *FileCache) Stats() Stats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	stats := Stats{
		Hits:   fc.metadata.HitCount,
		Misses: fc.metadata.MissCount,
	}

	filesDir := filepath.Join(fc.cacheDir, "files")

	entries, err := os.ReadDir(filesDir)
	if err == nil {
		stats.Size = len(entries)
		for _, entry := range entries {
			info, err := entry.Info()
			if err == nil {
				stats.BytesUsed += info.Size()
			}
		}
	}

	return stats
}

// GetStats returns cache hit/miss statistics.
func (fc *FileCache) GetStats() (int64, int64) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return fc.metadata.HitCount, fc.metadata.MissCount
}

// cachePath returns the full path for a cache entry.
func (fc *FileCache) cachePath(contentHash string) string {
	return filepath.Join(fc.cacheDir, "files", contentHash+".gob")
}

// serialize converts nodes to bytes using gob encoding.
func (fc *FileCache) serialize(nodes []*syntax.Node) ([]byte, error) {
	// Create a wrapper that includes version info
	wrapper := &cacheEntry{
		Version: CacheVersion,
		Nodes:   nodes,
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	err := enc.Encode(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize %d nodes: %w", len(nodes), err)
	}

	return buf.Bytes(), nil
}

// deserialize converts bytes back to nodes.
func (fc *FileCache) deserialize(data []byte) ([]*syntax.Node, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var wrapper cacheEntry

	err := dec.Decode(&wrapper)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to deserialize nodes (dataSize=%d bytes): %w",
			len(data),
			err,
		)
	}

	// Check version compatibility
	if wrapper.Version != CacheVersion {
		return nil, fmt.Errorf( //nolint:err113 // Error message needs dynamic values
			"cache version mismatch: got %d, want %d",
			wrapper.Version,
			CacheVersion,
		)
	}

	return wrapper.Nodes, nil
}

// loadMetadata loads cache metadata from disk.
func (fc *FileCache) loadMetadata() {
	metadataPath := filepath.Join(fc.cacheDir, "metadata.json")
	// #nosec G304 -- Path constructed from controlled cache directory
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return
	}

	_ = errors.SafeUnmarshal(data, &fc.metadata, "cache metadata")
}

// saveMetadata saves cache metadata to disk.
func (fc *FileCache) saveMetadata() error {
	metadataPath := filepath.Join(fc.cacheDir, "metadata.json")

	data, err := errors.SafeMarshalIndent(fc.metadata, "", "  ", "cache metadata")
	if err != nil {
		return errors.Wrap(err, errors.CacheError, "failed to marshal cache metadata")
	}

	err = os.WriteFile(metadataPath, data, cacheFilePerms)
	if err != nil {
		return errors.WrapFile(err, metadataPath, "write")
	}

	return nil
}

// cacheEntry is the serialized format for cached AST nodes.
type cacheEntry struct {
	Version int
	Nodes   []*syntax.Node
}

// Key generates a cache key from file content.
// This uses SHA1 for cache keys (fast and collision-resistant enough for this use case).
//
// Deprecated: Use Key instead. CacheKey is kept for backward compatibility.
func CacheKey(content []byte) string { return Key(content) }

// Key generates a cache key from file content.
// This uses SHA1 for cache keys (fast and collision-resistant enough for this use case).
func Key(content []byte) string {
	// #nosec G401 -- SHA1 used for cache keys, not cryptographic security
	h := sha1.Sum(content)

	return hex.EncodeToString(h[:])
}

//nolint:gochecknoinits // Required for gob registration of types used in cache serialization
func init() {
	// Register types for gob encoding (zero-values are intentional for registration)
	gob.Register(&syntax.Node{})
	gob.Register(&cacheEntry{})
}
