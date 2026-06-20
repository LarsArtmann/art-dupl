package detection

import (
	"context"
	"slices"

	"github.com/LarsArtmann/art-dupl/hash"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// suffixTreeAdapter wraps the suffix tree algorithm as a MethodDetector.
type suffixTreeAdapter struct {
	tree *suffixtree.STree
	data []*syntax.Node
}

func (a *suffixTreeAdapter) FindDuplOver(
	ctx context.Context,
	threshold int,
) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		suffixMatches := a.tree.FindDuplOver(ctx, threshold)

		for match := range suffixMatches {
			if ctx.Err() != nil {
				return
			}

			syntaxMatch := syntax.FindSyntaxUnits(a.data, match, threshold)
			if hasNonEmptyFrag(syntaxMatch.Frags) {
				select {
				case resultChan <- syntaxMatch:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resultChan
}

// hashAdapter wraps the hash-based detector as a MethodDetector.
type hashAdapter struct {
	data []*syntax.Node
}

func (a *hashAdapter) FindDuplOver(
	ctx context.Context,
	threshold int,
) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		hashDetector := hash.NewFileDetector(threshold)
		source := hashDetector.FindDuplOver(ctx, a.data, threshold)

		for match := range source {
			if ctx.Err() != nil {
				return
			}

			if hasNonEmptyFrag(match.Frags) {
				select {
				case resultChan <- match:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resultChan
}

// buildCloneDetectors creates MethodDetector instances based on the configured methods.
// Returns the list of active clone detectors in execution order.
func (md *MultiDetector) buildCloneDetectors() []MethodDetector {
	var detectors []MethodDetector

	if len(md.cfg.Methods) == 0 || slices.Contains(md.cfg.Methods, MethodArtDupl) {
		detectors = append(detectors, &suffixTreeAdapter{tree: md.tree, data: md.data})
	}

	if slices.Contains(md.cfg.Methods, MethodHash) {
		detectors = append(detectors, &hashAdapter{data: md.data})
	}

	return detectors
}
