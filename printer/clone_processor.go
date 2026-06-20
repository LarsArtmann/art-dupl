package printer

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// ErrZeroLengthDuplicate indicates a duplicate group with no nodes was encountered.
var ErrZeroLengthDuplicate = errors.New("zero length duplicate found")

// ProcessClones converts raw syntax.Node groups into ProcessedClone slices.
// This is the single point where [][]*syntax.Node is decoded into domain types,
// eliminating the need for each printer to understand AST internals.
func ProcessClones(fread ReadFile, dups [][]*syntax.Node) ([]domain.ProcessedClone, error) {
	if len(dups) == 0 {
		return nil, nil
	}

	clones := make([]domain.ProcessedClone, len(dups))

	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return nil, fmt.Errorf("%w at index %d", ErrZeroLengthDuplicate, i)
		}

		nstart := dup[0]
		nend := dup[cnt-1]

		fileInfo, err := ProcessNodeRange(fread, nstart, nend)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to process node range for file %s (index %d): %w",
				nstart.Filename,
				i,
				err,
			)
		}

		fragment := extractContent(fileInfo, nstart, nend)
		tokens := cnt
		lines := fileInfo.LineEnd - fileInfo.LineStart + 1

		clones[i] = domain.ProcessedClone{
			Filename:   fileInfo.Filename,
			LineStart:  fileInfo.LineStart,
			LineEnd:    fileInfo.LineEnd,
			Fragment:   fragment,
			TokenCount: tokens,
			FileSize:   len(fileInfo.Content),
			Classification: ClassifyClone(domain.ClassificationInput{
				Filename: fileInfo.Filename,
				NodeType: golang.DecodeBaseType(nstart.Type),
				Tokens:   tokens,
				Lines:    lines,
			}),
		}
	}

	// Populate group-level Actionability and CloneType after all per-instance
	// classifications are computed. Both are group properties: every instance
	// in a clone group shares the same actionability verdict and clone type.
	label, actionability := EvaluateActionabilityWithLabel(dups)
	cloneType := classifyCloneType(dups)

	for i := range clones {
		clones[i].Classification.Actionability = actionability
		clones[i].Classification.CloneType = cloneType
		clones[i].Classification = applyPatternLabel(
			clones[i].Classification, label,
		)
	}

	return clones, nil
}

// classifyCloneType determines the Bellon clone type (1/2/3) for a group of
// fragments by comparing the identifier Name fields across corresponding nodes.
//
// The suffix tree matches on Node.Type only; Name is never part of the matching
// criterion. Therefore any Name divergence across fragments signals a renamed
// (parameterized) clone:
//
//   - Type 1 (exact): every node position has identical Name across ALL fragments.
//   - Type 2 (parameterized): structure matches (guaranteed by the suffix tree)
//     but at least one identifier name differs between fragments.
//   - Type 3 (near-miss): fragments differ in length. The suffix tree normally
//     guarantees equal-length fragments, so this is a defensive fallback.
//
// In the current semantic mode (exact-name hashing) matched nodes always share
// identical Names, so clones are Type 1. After alpha-normalization (Type 2
// detection) renamed identifiers will match structurally while their original
// Names diverge, producing Type 2 classifications.
func classifyCloneType(dups [][]*syntax.Node) domain.CloneType {
	if len(dups) < 2 {
		return domain.CloneType1
	}

	// Flatten each fragment into its complete pre-order node sequence. The
	// fragment slice contains only the top-level "syntax unit" roots; the
	// renamed identifiers live in their descendants, so we must walk the full
	// subtree to detect Name divergence.
	seqs := make([][]*syntax.Node, len(dups))
	for i, dup := range dups {
		for _, node := range dup {
			seqs[i] = append(seqs[i], flattenSubtree(node)...)
		}
	}

	first := seqs[0]

	for _, other := range seqs[1:] {
		if len(other) != len(first) {
			return domain.CloneType3
		}

		for i := range first {
			if first[i].Name != other[i].Name {
				return domain.CloneType2
			}
		}
	}

	return domain.CloneType1
}

// flattenSubtree collects a node and all its descendants in pre-order, matching
// the serialization order used by the suffix tree so that corresponding
// positions across clone fragments align.
func flattenSubtree(n *syntax.Node) []*syntax.Node {
	var out []*syntax.Node
	flattenInto(n, &out)

	return out
}

func flattenInto(n *syntax.Node, out *[]*syntax.Node) {
	*out = append(*out, n)

	for _, child := range n.Children {
		flattenInto(child, out)
	}
}

// NodesToGroup converts raw syntax.Node groups into a ProcessedCloneGroup.
func NodesToGroup(
	fread ReadFile,
	hash string,
	dups [][]*syntax.Node,
) (domain.ProcessedCloneGroup, error) {
	clones, err := ProcessClones(fread, dups)
	if err != nil {
		return domain.ProcessedCloneGroup{}, fmt.Errorf("process clones for hash %s: %w", hash, err)
	}

	size := 0
	for _, c := range clones {
		size += c.TokenCount
	}

	return domain.ProcessedCloneGroup{
		Hash:       hash,
		TokenCount: size,
		Clones:     clones,
	}, nil
}
