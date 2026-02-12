package job

import (
	"context"
	"path/filepath"

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
