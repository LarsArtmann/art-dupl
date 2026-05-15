package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type Issue struct {
	From, To domain.ProcessedClone
}

type Issuer struct {
	ReadFile
}

func NewIssuer(fread ReadFile) *Issuer {
	return &Issuer{fread}
}

func (p *Issuer) MakeIssues(dups [][]*syntax.Node) ([]Issue, error) {
	clones, err := ProcessClones(p.ReadFile, dups)
	if err != nil {
		return nil, err
	}

	sort.Sort(byNameAndLineProcessed(clones))

	var issues []Issue

	for i := 1; i < len(clones); i++ {
		issues = append(issues, Issue{
			From: clones[0],
			To:   clones[i],
		})
	}

	return issues, nil
}
