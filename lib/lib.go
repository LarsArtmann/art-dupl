// Package lib Golangci-lint: altered version of main.go
package lib

import (
	"context"
	"os"
	"sort"

	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// DefaultCacheDir re-exports the default cache directory for consumers.
const DefaultCacheDir = cache.DefaultCacheDir

// IncrementalStats holds statistics from incremental parsing.
type IncrementalStats struct {
	FilesCount  int
	LinesCount  int
	CacheHits   int
	CacheMisses int
}

func Run(ctx context.Context, files []string, threshold int) ([]printer.Issue, error) {
	fchan := make(chan string, 1024)
	go func() {
		for _, f := range files {
			fchan <- f
		}
		close(fchan)
	}()
	schan, _ := job.Parse(ctx, fchan)
	t, data, done := job.BuildTree(ctx, schan)
	<-done

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	mchan := t.FindDuplOver(threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			match := syntax.FindSyntaxUnits(*data, m, threshold)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	return makeIssues(duplChan)
}

// RunIncremental runs duplicate detection with AST caching for incremental performance.
// On subsequent runs, unchanged files are loaded from cache instead of being reparsed.
// cacheDir specifies where to store cached AST nodes (empty for default: .cache/art-dupl/files/).
// clearCache forces a cache reset before running.
func RunIncremental(ctx context.Context, files []string, threshold int, cacheDir string, clearCache bool) ([]printer.Issue, IncrementalStats, error) {
	fchan := make(chan string, 1024)
	go func() {
		for _, f := range files {
			fchan <- f
		}
		close(fchan)
	}()

	incParser := job.NewIncrementalParser(cacheDir, clearCache)
	schan, statsChan := incParser.ParseIncremental(ctx, fchan)

	t, data, done := job.BuildTree(ctx, schan)
	<-done

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	mchan := t.FindDuplOver(threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			match := syntax.FindSyntaxUnits(*data, m, threshold)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	issues, err := makeIssues(duplChan)
	if err != nil {
		return nil, IncrementalStats{}, err
	}

	stats := <-statsChan
	return issues, IncrementalStats{
		FilesCount:  stats.FilesCount,
		LinesCount:  stats.LinesCount,
		CacheHits:   stats.CacheHits,
		CacheMisses: stats.CacheMisses,
	}, nil
}

func makeIssues(duplChan <-chan syntax.Match) ([]printer.Issue, error) {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	p := printer.NewIssuer(os.ReadFile)

	var issues []printer.Issue
	for _, k := range keys {
		uniq := utils.Unique(groups[k])
		if len(uniq) > 1 {
			i, err := p.MakeIssues(uniq)
			if err != nil {
				return nil, err //nolint:wrapcheck // Printer errors are already clear
			}
			issues = append(issues, i...)
		}
	}

	return issues, nil
}
