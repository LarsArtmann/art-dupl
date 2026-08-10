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
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	CacheVersion = 2

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

	err := os.MkdirAll(filesDir, cacheDirPerms)
	if err != nil {
		// Cache is best-effort — log and continue without it
		fmt.Fprintf(os.Stderr, "warning: failed to create cache directory %s: %v\n", filesDir, err)
	}

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

// withCachePath runs fn while holding the read lock and passing it the
// resolved cache file path for contentHash.
func (fc *FileCache) withCachePath(contentHash string, fn func(cachePath string)) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	fn(fc.cachePath(contentHash))
}

// Get retrieves cached AST nodes for the given content hash.
// Returns the nodes and true if found (cache hit), nil and false otherwise.
func (fc *FileCache) Get(contentHash string) ([]*syntax.Node, bool) {
	var (
		nodes []*syntax.Node
		hit   bool
	)

	fc.withCachePath(contentHash, func(path string) {
		// #nosec G304 -- Path constructed from controlled cache directory and content hash
		data, err := os.ReadFile(path)
		if err != nil {
			atomic.AddInt64(&fc.metadata.MissCount, 1)

			return
		}

		deserialized, derr := fc.deserialize(data)
		if derr != nil {
			fmt.Fprintf(os.Stderr, "warning: removing stale cache entry %s: %v\n", path, derr)

			removeErr := os.Remove(path)
			if removeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove stale cache entry %s: %v\n", path, removeErr)
			}

			atomic.AddInt64(&fc.metadata.MissCount, 1)

			return
		}

		nodes = deserialized
		hit = true

		atomic.AddInt64(&fc.metadata.HitCount, 1)
	})

	if !hit {
		return nil, false
	}

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
		return errors.NewIOError(
			filesDir,
			"failed to create cache directory for "+contentHash,
			err,
		)
	}

	data, err := fc.serialize(nodes)
	if err != nil {
		return fmt.Errorf("serialize nodes for hash %s: %w", contentHash, err)
	}

	cachePath := fc.cachePath(contentHash)

	err = os.WriteFile(cachePath, data, cacheFilePerms)
	if err != nil {
		return errors.NewIOError(cachePath, "failed to write cache file", err)
	}

	fc.metadata.UpdatedAt = time.Now()

	err = fc.saveMetadata()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save cache metadata: %v\n", err)
	}

	return nil
}

// Has checks if a cache entry exists for the given hash.
func (fc *FileCache) Has(contentHash string) bool {
	var exists bool

	fc.withCachePath(contentHash, func(path string) {
		_, err := os.Stat(path)
		exists = err == nil
	})

	return exists
}

// Remove deletes a cache entry.
func (fc *FileCache) Remove(contentHash string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	cachePath := fc.cachePath(contentHash)

	err := os.Remove(cachePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.NewIOError(
			cachePath,
			"failed to remove cache entry for hash "+contentHash,
			err,
		)
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
		Hits:   atomic.LoadInt64(&fc.metadata.HitCount),
		Misses: atomic.LoadInt64(&fc.metadata.MissCount),
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

// Prune removes the oldest cache entries until the cache contains at most
// maxEntries files. Entries are evicted by modification time (oldest first).
// If maxEntries <= 0, no pruning occurs.
func (fc *FileCache) Prune(maxEntries int) (int, error) {
	if maxEntries <= 0 {
		return 0, nil
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	filesDir := filepath.Join(fc.cacheDir, "files")

	entries, err := os.ReadDir(filesDir)
	if err != nil {
		return 0, nil // Directory might not exist yet
	}

	if len(entries) <= maxEntries {
		return 0, nil
	}

	type entryInfo struct {
		name    string
		modTime time.Time
	}

	infos := make([]entryInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		infos = append(infos, entryInfo{name: entry.Name(), modTime: info.ModTime()})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.Before(infos[j].modTime)
	})

	evicted := 0

	for i := range len(infos) - maxEntries {
		path := filepath.Join(filesDir, infos[i].name)
		if err := os.Remove(path); err == nil {
			evicted++
		}
	}

	return evicted, nil
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
		return nil, errors.NewValidationError(
			fmt.Sprintf("cache version mismatch: got %d, want %d", wrapper.Version, CacheVersion),
			nil,
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

	if fc.metadata.Version != 0 && fc.metadata.Version != CacheVersion {
		fmt.Fprintf(os.Stderr,
			"warning: cache version mismatch (metadata: got %d, want %d) — stale entries will be evicted on access\n",
			fc.metadata.Version, CacheVersion)
	}
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

// Key generates a cache key from file content using SHA-256.
func Key(content []byte) string {
	return KeyWithParams(content, "")
}

// KeyWithParams generates a cache key from file content AND a params string.
// The params string encodes detection-affecting configuration (e.g., detection
// mode, maxChildren, type-aware mode) so that ASTs serialized under different
// configurations get separate cache entries. This prevents cross-mode cache
// contamination where a cached semantic-mode AST is incorrectly reused for an
// exact-mode run.
//
// The content length is written as a prefix to prevent hash ambiguity:
// without it, content="ab"+params="cd" would hash the same stream as
// content="a"+params="bcd".
func KeyWithParams(content []byte, params string) string {
	h := sha256.New()
	// Length-prefix prevents concatenation ambiguity between content and params.
	if _, err := h.Write([]byte(strconv.Itoa(len(content)))); err != nil {
		panic(fmt.Sprintf("sha256 Write failed (infallible): %v", err))
	}

	if _, err := h.Write([]byte{':'}); err != nil {
		panic(fmt.Sprintf("sha256 Write failed (infallible): %v", err))
	}

	if _, err := h.Write(content); err != nil {
		panic(fmt.Sprintf("sha256 Write failed (infallible): %v", err))
	}

	if _, err := h.Write([]byte(params)); err != nil {
		panic(fmt.Sprintf("sha256 Write failed (infallible): %v", err))
	}

	return hex.EncodeToString(h.Sum(nil))
}

//nolint:gochecknoinits // Required for gob registration of types used in cache serialization
func init() {
	// Register types for gob encoding (zero-values are intentional for registration)
	gob.Register(&syntax.Node{})
	gob.Register(&cacheEntry{})
}
