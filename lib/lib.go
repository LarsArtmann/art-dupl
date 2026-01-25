// Package lib Golangci-lint: altered version of main.go
package lib

import (
	"context"
	"os"
	"sort"

	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
)

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
