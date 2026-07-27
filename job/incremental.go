package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"golang.org/x/sync/singleflight"
)

// errUnexpectedSingleflightType is returned when a singleflight result does not
// have the expected []*syntax.Node type. This should be unreachable since the
// callback is internal, but the type assertion is guarded defensively.
var errUnexpectedSingleflightType = errors.New("unexpected singleflight result type")

// IncrementalParser parses files with caching support.
// It uses content hashing to detect unchanged files and loads cached AST nodes.
type IncrementalParser struct {
	cache           *cache.FileCache
	clearCache      bool
	mode            golang.DetectionMode
	maxChildren     int
	maxCacheEntries int
	group           singleflight.Group
	typeInfos       golang.TypeAwareData
}

// NewIncrementalParser creates a new IncrementalParser.
func NewIncrementalParser(
	cacheDir string,
	clearCache bool,
	mode golang.DetectionMode,
	maxChildren int,
	maxCacheEntries int,
) *IncrementalParser {
	logger.Default.Info(
		"creating incremental parser",
		"cacheDir",
		cacheDir,
		"clearCache",
		clearCache,
		"mode",
		mode,
	)

	return &IncrementalParser{
		cache:           cache.NewFileCache(cacheDir),
		clearCache:      clearCache,
		mode:            mode,
		maxChildren:     maxChildren,
		maxCacheEntries: maxCacheEntries,
	}
}

// SetTypeAwareData configures the incremental parser to use pre-loaded type
// information for type-aware detection. When set, parseFile passes the
// pre-loaded AST + TypeInfo to the transformer instead of nil.
func (ip *IncrementalParser) SetTypeAwareData(td golang.TypeAwareData) {
	ip.typeInfos = td
}

// ParseStatsMixin provides common fields for parsing statistics.
type ParseStatsMixin struct {
	FilesCount int
	LinesCount int
}

// IncrementalStats holds statistics from incremental parsing.
type IncrementalStats struct {
	ParseStatsMixin

	CacheHits   int
	CacheMisses int
}

// ParseIncremental parses files sequentially with caching.
// For each file, it checks the cache first using content hash.
// If found in cache, uses cached AST nodes. Otherwise parses and caches the result.
func (ip *IncrementalParser) ParseIncremental(
	ctx context.Context,
	fchan chan string,
) (chan []*syntax.Node, chan IncrementalStats) {
	if ip.clearCache {
		err := ip.cache.Clear()
		if err != nil {
			logger.Default.Error("failed to clear cache", "err", err)
		}
	}

	schan := make(chan []*syntax.Node)
	statsChan := make(chan IncrementalStats, 1)

	go func() {
		stats := IncrementalStats{} // zero-value counters, mutated incrementally

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

			if !sendCtx(ctx, schan, nodes) {
				statsChan <- stats

				close(schan)

				return
			}
		}

		statsChan <- stats

		close(schan)
	}()

	return schan, statsChan
}

// incrementalResult holds the result of incrementally parsing a single file.
type incrementalResult struct {
	nodes     []*syntax.Node
	lines     int
	fromCache bool
}

// ParseIncrementalParallel parses files concurrently using a worker pool.
// Workers defaults to runtime.GOMAXPROCS(0) if <= 0.
// Uses singleflight to deduplicate concurrent cache misses for files with
// identical content (common in monorepos with vendored or copied code).
func (ip *IncrementalParser) ParseIncrementalParallel(
	ctx context.Context,
	fchan chan string,
	workers int,
) (chan []*syntax.Node, chan IncrementalStats) {
	if ip.clearCache {
		err := ip.cache.Clear()
		if err != nil {
			logger.Default.Error("failed to clear cache", "err", err)
		}
	}

	workers = normalizeWorkerCount(workers)

	resultChan := make(chan incrementalResult, workers*2)
	statsChan := make(chan IncrementalStats, 1)

	var wg sync.WaitGroup

	fileQueue := make(chan string, workers*2)

	startIncrementalWorkers(ctx, &wg, ip, fileQueue, resultChan, workers)

	go feedFiles(ctx, fchan, fileQueue)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	schan := make(chan []*syntax.Node)

	go collectIncrementalResults(ctx, resultChan, schan, statsChan)

	return schan, statsChan
}

// startIncrementalWorkers launches a pool of worker goroutines that pull file
// paths from fileQueue, parse them via the cache-backed parseFile, and forward
// results to resultChan. Context cancellation aborts a worker between files.
func startIncrementalWorkers(
	ctx context.Context,
	wg *sync.WaitGroup,
	ip *IncrementalParser,
	fileQueue <-chan string,
	resultChan chan<- incrementalResult,
	workers int,
) {
	for range workers {
		wg.Go(func() {
			for file := range fileQueue {
				select {
				case <-ctx.Done():
					return
				default:
				}

				nodes, lines, fromCache := ip.parseFile(file)

				if !sendCtx(ctx, resultChan, incrementalResult{
					nodes:     nodes,
					lines:     lines,
					fromCache: fromCache,
				}) {
					return
				}
			}
		})
	}
}

// collectIncrementalResults aggregates per-file results into stats, forwarding
// serialized node sequences to schan. On context cancellation it emits the
// partial stats and closes schan so downstream consumers can terminate.
func collectIncrementalResults(
	ctx context.Context,
	resultChan <-chan incrementalResult,
	schan chan<- []*syntax.Node,
	statsChan chan<- IncrementalStats,
) {
	stats := IncrementalStats{} // zero-value counters, mutated incrementally

	for result := range resultChan {
		stats.FilesCount++
		stats.LinesCount += result.lines

		if result.fromCache {
			stats.CacheHits++
		} else {
			stats.CacheMisses++
		}

		if !sendCtx(ctx, schan, result.nodes) {
			statsChan <- stats

			close(schan)

			return
		}
	}

	statsChan <- stats

	close(schan)
}

// GetCacheStats returns cache statistics.
func (ip *IncrementalParser) GetCacheStats() cache.Stats {
	return ip.cache.Stats()
}

// parseFile parses a single file, using cache if available.
// On a cache miss, singleflight deduplicates concurrent parses of files with
// identical content hashes — only one goroutine parses, the rest clone the result.
// Returns the serialized nodes, line count, and whether it was a cache hit.
func (ip *IncrementalParser) parseFile(file string) ([]*syntax.Node, int, bool) {
	// #nosec G304 -- File path comes from controlled source directory walk
	content, err := os.ReadFile(file)
	if err != nil {
		return ip.handleFileError(file, err, "read")
	}

	contentHash := cache.Key(content)
	lines := countLines(content)

	// Fast path: cache hit (thread-safe via RWMutex)
	if cachedNodes, hit := ip.cache.Get(contentHash); hit {
		return ip.cloneWithFilename(cachedNodes, file), lines, true
	}

	// Slow path: cache miss — singleflight deduplicates concurrent parses
	// of files with identical content (e.g. vendored copies, generated stubs).
	v, err, _ := ip.group.Do(contentHash, func() (any, error) {
		// Double-check: another worker may have populated the cache while we waited.
		if cachedNodes, hit := ip.cache.Get(contentHash); hit {
			return cachedNodes, nil
		}

		ast, _, parseErr := ParseFileByExtensionWithConfig(file, ip.mode, ip.typeInfos.LookupPreloaded(file))
		if parseErr != nil {
			return nil, fmt.Errorf("parse file: %w", parseErr)
		}

		nodes := syntax.SerializeWithMaxChildren(ast, ip.maxChildren)

		// Store a deep-cloned copy in cache — the returned slice and cached
		// slice must be independent for concurrent access.
		cachedNodes := deepCloneNodes(nodes)

		cacheErr := ip.cache.Set(contentHash, cachedNodes)
		if cacheErr != nil {
			logger.Default.Error("failed to cache file", "file", file, "err", cacheErr)
		} else {
			logger.Default.Info("cached file AST", "file", file, "hash", contentHash)
		}

		if ip.maxCacheEntries > 0 {
			if evicted, pruneErr := ip.cache.Prune(ip.maxCacheEntries); pruneErr == nil && evicted > 0 {
				logger.Default.Info("pruned cache entries", "evicted", evicted, "max", ip.maxCacheEntries)
			}
		}

		return nodes, nil
	})
	if err != nil {
		return ip.handleFileError(file, err, "parse")
	}

	// The singleflight result is shared — deep-clone before mutating Filename.
	parsedNodes, ok := v.([]*syntax.Node)
	if !ok {
		return ip.handleFileError(file, fmt.Errorf("%w: %T", errUnexpectedSingleflightType, v), "parse")
	}

	nodes := ip.cloneWithFilename(parsedNodes, file)

	return nodes, lines, false
}

// cloneWithFilename deep-clones nodes and sets the filename for each.
// This is necessary because cached/shared nodes have a different (or empty)
// filename than the file currently being processed.
func (ip *IncrementalParser) cloneWithFilename(nodes []*syntax.Node, filename string) []*syntax.Node {
	result := make([]*syntax.Node, len(nodes))
	interned := syntax.InternFilename(filename)

	for i, node := range nodes {
		cloned := node.Clone()
		cloned.Filename = interned
		result[i] = cloned
	}

	return result
}

// deepCloneNodes creates independent copies of all nodes (no shared pointers).
func deepCloneNodes(nodes []*syntax.Node) []*syntax.Node {
	result := make([]*syntax.Node, len(nodes))
	for i, node := range nodes {
		result[i] = node.Clone()
	}

	return result
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
