package printer

// Golangci-lint: altered version of plumbing.go

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type Clone clone

func (c Clone) Filename() string {
	return c.filename
}

func (c Clone) LineStart() int {
	return c.lineStart
}

func (c Clone) LineEnd() int {
	return c.lineEnd
}

func (c Clone) Fragment() []byte {
	return c.fragment
}

type Issue struct {
	From, To Clone
}

type Issuer struct {
	ReadFile
}

func NewIssuer(fread ReadFile) *Issuer {
	return &Issuer{fread}
}

func (p *Issuer) MakeIssues(dups [][]*syntax.Node) ([]Issue, error) {
	clones, err := prepareClonesInfo(p.ReadFile, dups)
	if err != nil {
		return nil, err
	}

	sort.Sort(byNameAndLine(clones))

	var issues []Issue

	// Pair each clone with the first (reference) clone.
	// This avoids bidirectional pairs (A→B and B→A) which represent the same issue.
	// For n clones, we generate n-1 pairs: clone[0]→clone[1], clone[0]→clone[2], etc.
	for i := 1; i < len(clones); i++ {
		issues = append(issues, Issue{
			From: Clone(clones[0]),
			To:   Clone(clones[i]),
		})
	}

	return issues, nil
}
