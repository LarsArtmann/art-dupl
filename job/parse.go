package job

import (
	"context"
	"runtime"
	"sync"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// ParseStats holds statistics from the parsing phase.
type ParseStats struct {
	ParseStatsMixin
}

// parseResult holds the result of parsing a single file.
type parseResult struct {
	ast   *syntax.Node
	lines int
	err   error
}

// Parse parses files sequentially (legacy behavior).
// mode controls how identifier names participate in matching.
// maxChildren caps the number of children serialized per node (prevents stack overflow).
func Parse(
	ctx context.Context,
	fchan chan string,
	mode golang.DetectionMode,
	maxChildren int,
) (chan []*syntax.Node, chan ParseStats) {
	// parse AST
	achan := make(chan *syntax.Node)
	statsChan := make(chan ParseStats, 1)

	go func() {
		fileCount := 0
		lineCount := 0

		for file := range fchan {
			select {
			case <-ctx.Done():
				statsChan <- ParseStats{ParseStatsMixin: ParseStatsMixin{FilesCount: fileCount, LinesCount: lineCount}}

				close(achan)

				return
			default:
			}

			fileCount++

			var (
				ast   *syntax.Node
				lines int
				err   error
			)

			// Dispatch to appropriate parser based on file extension

			ast, lines, err = ParseFileByExtensionWithConfig(file, mode)
			if err != nil {
				logger.Default.Error("failed to parse file", "file", file, "err", err)

				continue
			}

			lineCount += lines

			select {
			case achan <- ast:
			case <-ctx.Done():
				statsChan <- ParseStats{ParseStatsMixin: ParseStatsMixin{FilesCount: fileCount, LinesCount: lineCount}}

				close(achan)

				return
			}
		}

		statsChan <- ParseStats{ParseStatsMixin: ParseStatsMixin{FilesCount: fileCount, LinesCount: lineCount}}

		close(achan)
	}()

	// serialize
	schan := make(chan []*syntax.Node)
	go serializeAST(ctx, achan, schan, maxChildren)

	return schan, statsChan
}

// ParseParallel parses files concurrently using a worker pool.
// Workers defaults to runtime.GOMAXPROCS(0) if <= 0.
// mode controls how identifier names participate in matching.
// maxChildren caps the number of children serialized per node (prevents stack overflow).
func ParseParallel(
	ctx context.Context,
	fchan chan string,
	workers int,
	mode golang.DetectionMode,
	maxChildren int,
) (chan []*syntax.Node, chan ParseStats) {
	workers = normalizeWorkerCount(workers)

	resultChan := make(chan parseResult, workers*2)
	statsChan := make(chan ParseStats, 1)

	// Start workers
	var wg sync.WaitGroup

	fileQueue := make(chan string, workers*2)

	startWorkers(ctx, &wg, fileQueue, resultChan, workers, mode)

	// Feed files to workers
	go feedFiles(ctx, fchan, fileQueue)

	// Collect results and forward to achan
	achan := make(chan *syntax.Node)

	go closeResultChan(&wg, resultChan)

	go collectResults(ctx, resultChan, achan, statsChan)

	// serialize
	schan := make(chan []*syntax.Node)
	go serializeAST(ctx, achan, schan, maxChildren)

	return schan, statsChan
}

// normalizeWorkerCount returns a valid worker count.
func normalizeWorkerCount(workers int) int {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	if workers < 1 {
		workers = 1
	}

	return workers
}

// startWorkers starts the worker goroutines.
func startWorkers(
	ctx context.Context,
	wg *sync.WaitGroup,
	fileQueue <-chan string,
	resultChan chan<- parseResult,
	workers int,
	mode golang.DetectionMode,
) {
	for range workers {
		wg.Go(func() {
			for file := range fileQueue {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result := parseFileWithConfig(file, mode)

				select {
				case resultChan <- result:
				case <-ctx.Done():
					return
				}
			}
		})
	}
}

// parseFileWithConfig parses a single file with the given detection mode.
func parseFileWithConfig(file string, mode golang.DetectionMode) parseResult {
	ast, lines, err := ParseFileByExtensionWithConfig(file, mode)

	return parseResult{ast: ast, lines: lines, err: err}
}

// feedFiles feeds files from fchan to the fileQueue.
func feedFiles(ctx context.Context, fchan <-chan string, fileQueue chan<- string) {
	for file := range fchan {
		select {
		case <-ctx.Done():
			close(fileQueue)

			return
		case fileQueue <- file:
		}
	}

	close(fileQueue)
}

// closeResultChan waits for workers to finish and closes the result channel.
func closeResultChan(wg *sync.WaitGroup, resultChan chan parseResult) {
	wg.Wait()
	close(resultChan)
}

// collectResults collects parse results and forwards AST nodes.
func collectResults(
	ctx context.Context,
	resultChan <-chan parseResult,
	achan chan<- *syntax.Node,
	statsChan chan<- ParseStats,
) {
	fileCount := 0
	lineCount := 0

	for result := range resultChan {
		if result.err != nil {
			logger.Default.Error("failed to parse file", "err", result.err)

			continue
		}

		fileCount++

		lineCount += result.lines

		select {
		case achan <- result.ast:
		case <-ctx.Done():
			statsChan <- ParseStats{ParseStatsMixin: ParseStatsMixin{FilesCount: fileCount, LinesCount: lineCount}}

			close(achan)

			return
		}
	}

	statsChan <- ParseStats{ParseStatsMixin: ParseStatsMixin{FilesCount: fileCount, LinesCount: lineCount}}

	close(achan)
}

// serializeAST serializes AST nodes to token sequences.
func serializeAST(ctx context.Context, achan <-chan *syntax.Node, schan chan<- []*syntax.Node, maxChildren int) {
	for ast := range achan {
		select {
		case <-ctx.Done():
			close(schan)

			return
		default:
		}

		seq := syntax.SerializeWithMaxChildren(ast, maxChildren)

		if !sendCtx(ctx, schan, seq) {
			close(schan)

			return
		}
	}

	close(schan)
}
