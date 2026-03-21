package job

import (
	"context"
	"os"

	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// IncrementalParser parses files with caching support.
// It uses content hashing to detect unchanged files and loads cached AST nodes.
type IncrementalParser struct {
	cache      *cache.FileCache
	clearCache bool
	semantic   bool
}

// NewIncrementalParser creates a new IncrementalParser.
func NewIncrementalParser(cacheDir string, clearCache bool, semantic bool) *IncrementalParser {
	logger.Default.Info(
		"creating incremental parser",
		"cacheDir",
		cacheDir,
		"clearCache",
		clearCache,
		"semantic",
		semantic,
	)

	return &IncrementalParser{
		cache:      cache.NewFileCache(cacheDir),
		clearCache: clearCache,
		semantic:   semantic,
	}
}

// IncrementalStats holds statistics from incremental parsing.
type IncrementalStats struct {
	FilesCount  int
	LinesCount  int
	CacheHits   int
	CacheMisses int
}

// ParseIncremental parses files with caching.
// For each file, it checks the cache first using content hash.
// If found in cache, uses cached AST nodes. Otherwise parses and caches the result.
func (ip *IncrementalParser) ParseIncremental(
	ctx context.Context,
	fchan chan string,
) (chan []*syntax.Node, chan IncrementalStats) {
	// Clear cache if requested
	if ip.clearCache {
		err := ip.cache.Clear()
		if err != nil {
			logger.Default.Error("failed to clear cache", "err", err)
		}
	}

	schan := make(chan []*syntax.Node)
	statsChan := make(chan IncrementalStats, 1)

	go func() {
		stats := IncrementalStats{} //nolint:exhaustruct

		for file := range fchan {
			select {
			case <-ctx.Done():
				statsChan <- stats

				close(schan)

				return
			default:
			}

			stats.FilesCount++
			nodes, lines, fromCache := ip.parseFile(file)
			stats.LinesCount += lines

			if fromCache {
				stats.CacheHits++
			} else {
				stats.CacheMisses++
			}

			schan <- nodes
		}

		statsChan <- stats

		close(schan)
	}()

	return schan, statsChan
}

// GetCacheStats returns cache statistics.
func (ip *IncrementalParser) GetCacheStats() cache.Stats {
	return ip.cache.Stats()
}

// parseFile parses a single file, using cache if available.
// Returns the serialized nodes, line count, and whether it was a cache hit.
func (ip *IncrementalParser) parseFile(file string) ([]*syntax.Node, int, bool) {
	// Read file content for hashing
	// #nosec G304 -- File path comes from controlled source directory walk in ParseIncremental
	content, err := os.ReadFile(file)
	if err != nil {
		return ip.handleFileError(file, err, "read")
	}

	// Compute content hash
	contentHash := cache.Key(content)

	// Check cache first
	if cachedNodes, hit := ip.cache.Get(contentHash); hit {
		// Cache hit - use cached nodes but update filename to current file
		// This is critical because cached nodes retain the filename of the first file
		// that was cached with this content hash
		for _, node := range cachedNodes {
			node.Filename = file
		}

		lines := countLines(content)

		return cachedNodes, lines, true
	}

	// Cache miss - parse the file
	var (
		ast   *syntax.Node
		lines int
	)

	ast, lines, err = ParseFileByExtensionWithConfig(file, ip.semantic)
	if err != nil {
		return ip.handleFileError(file, err, "parse")
	}

	// Serialize AST to nodes
	nodes := syntax.Serialize(ast)

	// Cache the result
	cacheErr := ip.cache.Set(contentHash, nodes)
	if cacheErr != nil {
		logger.Default.Error("failed to cache file", "file", file, "err", cacheErr)
	} else {
		logger.Default.Info("cached file AST", "file", file, "hash", contentHash)
	}

	return nodes, lines, false
}

// handleFileError logs the error and returns zero values.
func (ip *IncrementalParser) handleFileError(
	file string,
	err error,
	operation string,
) ([]*syntax.Node, int, bool) {
	logger.Default.Error(
		"failed to "+operation+" file",
		"file", file,
		"operation", operation,
		"err", err,
	)

	return nil, 0, false
}

// countLines counts the number of lines in content.
func countLines(content []byte) int {
	count := 0

	for _, b := range content {
		if b == '\n' {
			count++
		}
	}
	// Count last line if it doesn't end with newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		count++
	}

	return count
}
