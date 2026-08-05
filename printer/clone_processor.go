package printer

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer/actionability"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// ErrZeroLengthDuplicate indicates a duplicate group with no nodes was encountered.
var ErrZeroLengthDuplicate = errors.New("zero length duplicate found")

type (
	CloneCategory       = domain.CloneCategory
	ClonePriority       = domain.ClonePriority
	CloneClassification = domain.CloneClassification
)

// ToCloneNodeSeqs converts raw syntax.Node sequences into domain.CloneNode
// sequences. This is the exported bridge that decouples the actionability
// evaluation layer from syntax.Node internals. External callers (e.g. cmd/)
// use this before calling EvaluateActionability.
func ToCloneNodeSeqs(dups [][]*syntax.Node) [][]*domain.CloneNode {
	return toCloneNodeSeqs(dups)
}

// toCloneNodeSeqs converts raw syntax.Node sequences into domain.CloneNode
// sequences. This is the bridge that decouples the actionability evaluation
// layer from syntax.Node internals.
func toCloneNodeSeqs(dups [][]*syntax.Node) [][]*domain.CloneNode {
	result := make([][]*domain.CloneNode, len(dups))
	for i, seq := range dups {
		result[i] = make([]*domain.CloneNode, len(seq))
		for j, n := range seq {
			result[i][j] = syntaxToCloneNode(n)
		}
	}

	return result
}

// syntaxToCloneNode recursively converts a syntax.Node subtree into an
// immutable domain.CloneNode, decoding the base AST type from the semantic
// encoding.
func syntaxToCloneNode(n *syntax.Node) *domain.CloneNode {
	cn := &domain.CloneNode{
		BaseType:             golang.DecodeBaseType(n.Type),
		Name:                 n.Name,
		Filename:             n.Filename,
		VarType:              n.VarType,
		EnclosingReturnArity: n.EnclosingReturnArity,
		InterfaceMethod:      n.InterfaceMethod,
		IsAlias:              n.IsAlias,
	}
	if len(n.Children) > 0 {
		cn.Children = make([]*domain.CloneNode, len(n.Children))
		for i, c := range n.Children {
			cn.Children[i] = syntaxToCloneNode(c)
		}
	}

	return cn
}

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
			CloneRef: domain.CloneRef{
				Filename:  fileInfo.Filename,
				LineStart: fileInfo.LineStart,
				LineEnd:   fileInfo.LineEnd,
				Fragment:  string(fragment),
			},
			StartPos:   nstart.Pos,
			EndPos:     nend.End,
			TokenCount: tokens,
			FileSize:   len(fileInfo.Content),
			Classification: actionability.ClassifyClone(domain.ClassificationInput{
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
	label, verdict := actionability.EvaluateActionabilityWithLabel(toCloneNodeSeqs(dups))
	cloneType := classifyCloneType(dups)

	for i := range clones {
		clones[i].Classification.Actionability = verdict
		clones[i].Classification.CloneType = cloneType
		clones[i].Classification.Extractability = domain.AssessExtractability(
			clones[i].LineCount(), len(clones),
			clones[i].Classification.Category.IsCompleteUnit(),
		)
		clones[i].Classification = actionability.ApplyPatternLabel(
			clones[i].Classification, label,
		)

		if label != actionability.PatternNone {
			clones[i].Classification.NonActionablePattern = string(label)
		}
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
// In exact mode (verbatim name hashing) matched nodes always share identical
// Names, so clones are Type 1. In semantic mode (alpha-normalization, the
// default), renamed identifiers match structurally while their original Names
// diverge, producing Type 2 classifications. Structural mode ignores names
// entirely, also yielding Type 2 when names happen to differ.
func classifyCloneType(dups [][]*syntax.Node) domain.CloneType {
	if len(dups) < 2 {
		return domain.CloneType1
	}

	// Walk node.Children directly to collect Names in pre-order. We deliberately
	// avoid syntax.Serialize here because it destructively mutates n.Type for
	// statement nodes via fingerprintSubtree — calling it during classification
	// corrupts the tree for downstream consumers.
	nameSeqs := make([][]string, len(dups))
	for i, dup := range dups {
		nameSeqs[i] = collectNamesPreOrder(dup)
	}

	first := nameSeqs[0]

	for _, other := range nameSeqs[1:] {
		if len(other) != len(first) {
			return domain.CloneType3
		}

		for i := range first {
			if first[i] != other[i] {
				return domain.CloneType2
			}
		}
	}

	return domain.CloneType1
}

// collectNamesPreOrder traverses the given nodes and all their descendants in
// pre-order, collecting the Name field of each node. Unlike syntax.Serialize,
// this does NOT mutate any node fields.
func collectNamesPreOrder(nodes []*syntax.Node) []string {
	var names []string

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		names = append(names, n.Name)
		for _, child := range n.Children {
			walk(child)
		}
	}

	for _, node := range nodes {
		walk(node)
	}

	return names
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

	return domain.NewProcessedCloneGroup(hash, clones), nil
}
