package job

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	"github.com/LarsArtmann/art-dupl/syntax/templ"
)

// ParseStats holds statistics from the parsing phase.
type ParseStats struct {
	FilesCount int
	LinesCount int
}

// parseResult holds the result of parsing a single file.
type parseResult struct {
	ast   *syntax.Node
	lines int
	err   error
}

// Parse parses files sequentially (legacy behavior).
func Parse(ctx context.Context, fchan chan string) (chan []*syntax.Node, chan ParseStats) {
	// parse AST
	achan := make(chan *syntax.Node)
	statsChan := make(chan ParseStats, 1)
	go func() {
		fileCount := 0
		lineCount := 0
		for file := range fchan {
			select {
			case <-ctx.Done():
				statsChan <- ParseStats{FilesCount: fileCount, LinesCount: lineCount}
				close(achan)
				return
			default:
			}
			fileCount++

			var ast *syntax.Node
			var lines int
			var err error

			// Dispatch to appropriate parser based on file extension
			switch filepath.Ext(file) {
			case ".templ":
				ast, lines, err = templ.ParseWithLineCount(file)
			default:
				// Default to Go parser for .go files and any other files that reach here
				ast, lines, err = golang.ParseWithLineCount(file)
			}

			if err != nil {
				logger.Default.Error("failed to parse file", "file", file, "err", err)
				continue
			}
			lineCount += lines
			achan <- ast
		}
		statsChan <- ParseStats{FilesCount: fileCount, LinesCount: lineCount}
		close(achan)
	}()

	// serialize
	schan := make(chan []*syntax.Node)
	go func() {
		for ast := range achan {
			select {
			case <-ctx.Done():
				close(schan)
				return
			default:
			}
			seq := syntax.Serialize(ast)
			schan <- seq
		}
		close(schan)
	}()
	return schan, statsChan
}

// ParseParallel parses files concurrently using a worker pool.
// Workers defaults to runtime.GOMAXPROCS(0) if <= 0.
func ParseParallel(ctx context.Context, fchan chan string, workers int) (chan []*syntax.Node, chan ParseStats) {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	if workers < 1 {
		workers = 1
	}

	resultChan := make(chan parseResult, workers*2)
	statsChan := make(chan ParseStats, 1)

	// Start workers
	var wg sync.WaitGroup
	fileQueue := make(chan string, workers*2)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileQueue {
				select {
				case <-ctx.Done():
					return
				default:
				}

				var ast *syntax.Node
				var lines int
				var err error

				switch filepath.Ext(file) {
				case ".templ":
					ast, lines, err = templ.ParseWithLineCount(file)
				default:
					ast, lines, err = golang.ParseWithLineCount(file)
				}

				resultChan <- parseResult{ast: ast, lines: lines, err: err}
			}
		}()
	}

	// Feed files to workers
	go func() {
		for file := range fchan {
			select {
			case <-ctx.Done():
				close(fileQueue)
				return
			case fileQueue <- file:
			}
		}
		close(fileQueue)
	}()

	// Collect results and forward to achan
	achan := make(chan *syntax.Node)
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		fileCount := 0
		lineCount := 0
		for result := range resultChan {
			if result.err != nil {
				logger.Default.Error("failed to parse file", "err", result.err)
				continue
			}
			fileCount++
			lineCount += result.lines
			achan <- result.ast
		}
		statsChan <- ParseStats{FilesCount: fileCount, LinesCount: lineCount}
		close(achan)
	}()

	// serialize
	schan := make(chan []*syntax.Node)
	go func() {
		for ast := range achan {
			select {
			case <-ctx.Done():
				close(schan)
				return
			default:
			}
			seq := syntax.Serialize(ast)
			schan <- seq
		}
		close(schan)
	}()

	return schan, statsChan
}
