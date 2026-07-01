package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isDataDominated reports whether every clone is dominated by data nodes
// (BasicLit and KeyValueExpr) rather than logic. When ≥60% of all nodes
// are data, the clone represents struct initialization, config fixtures,
// or test data arrays — not duplicated business logic.
func isDataDominated(nodeSeqs [][]*domain.CloneNode) bool {
	if len(nodeSeqs) == 0 {
		return false
	}

	return everySequenceMatch(nodeSeqs, isSequenceDataDominated)
}

const dataDominanceRatio = 0.6

// isSequenceDataDominated checks if a single clone sequence is dominated
// by data nodes (BasicLit, KeyValueExpr) rather than logic nodes.
func isSequenceDataDominated(seq []*domain.CloneNode) bool {
	total := 0
	data := 0

	for _, node := range seq {
		countDataNodes(node, &total, &data)
	}

	if total == 0 {
		return false
	}

	return float64(data)/float64(total) >= dataDominanceRatio
}

// countDataNodes walks a node tree counting all nodes and data-type nodes.
// Data nodes are BasicLit (string/number literals) and KeyValueExpr
// (struct field initializers). A high ratio of data nodes indicates
// struct initialization or config fixtures, not duplicated logic.
func countDataNodes(node *domain.CloneNode, total, data *int) {
	*total++

	bt := node.BaseType
	if bt == golang.BasicLit || bt == golang.KeyValueExpr {
		*data++
	}

	for _, child := range node.Children {
		countDataNodes(child, total, data)
	}
}

// isDescribeTablePattern detects Ginkgo DescribeTable/DescribeTableEntry
// patterns. These are framework-generated repetitive structures where each
// Entry call shares the same body shape with different data — structurally
// duplicated but intentionally so.
func isDescribeTablePattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		return containsCallTo(seq, "DescribeTable", "Entry", "FDescribeTable", "PDescribeTable")
	})
}

// isBuilderCallbackPattern detects builder/callback patterns where the
// dominant structure is a chain of method calls on different receiver types.
// This indicates intentional API design (builder pattern, fluent interface)
// rather than logic duplication.
func isBuilderCallbackPattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, isChainOfCallsWithDifferentReceivers)
}

// containsCallTo checks if any node in the sequence is a CallExpr that
// calls one of the named functions (via Ident or SelectorExpr Name field).
func containsCallTo(seq []*domain.CloneNode, names ...string) bool {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	for _, node := range seq {
		if node.BaseType != golang.CallExpr {
			continue
		}

		if node.Name != "" && nameSet[node.Name] {
			return true
		}

		for _, child := range node.Children {
			if child.Name != "" && nameSet[child.Name] {
				return true
			}
		}
	}

	return false
}

// isChainOfCallsWithDifferentReceivers checks if a sequence is dominated by
// CallExpr nodes where each call targets a different receiver — indicating
// builder pattern chains (a.WithX().WithY().Build()) rather than logic.
func isChainOfCallsWithDifferentReceivers(seq []*domain.CloneNode) bool {
	if len(seq) < 3 {
		return false
	}

	callCount := 0
	receiverNames := make(map[string]bool)

	for _, node := range seq {
		if node.BaseType != golang.CallExpr {
			continue
		}

		callCount++

		for _, child := range node.Children {
			if child.BaseType == golang.SelectorExpr && child.Name != "" {
				receiverNames[child.Name] = true
			}
		}
	}

	return callCount >= 3 && len(receiverNames) >= 2
}
