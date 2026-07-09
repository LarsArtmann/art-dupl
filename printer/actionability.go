package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// everySequenceMatch returns true if pred returns true for every sequence.
// Used by actionability checks that require ALL clones to match a pattern.
func everySequenceMatch(
	nodeSeqs [][]*domain.CloneNode,
	pred func(seq []*domain.CloneNode) bool,
) bool {
	for _, seq := range nodeSeqs {
		if !pred(seq) {
			return false
		}
	}

	return true
}

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// A group is actionable when it contains real logic that can be extracted,
// composed, or otherwise refactored. Non-actionable patterns are standard
// Go idioms that cannot be eliminated without breaking semantics.
//
// To be non-actionable, EVERY clone in the group must match the same
// boilerplate pattern. If any clone differs, the group is actionable.
func EvaluateActionability(nodeSeqs [][]*domain.CloneNode) domain.CloneActionability {
	_, a := evaluateActionabilityDetailed(nodeSeqs)

	return a
}

// PatternLabel identifies which non-actionable pattern was detected.
// Empty string means actionable.
type PatternLabel string

const (
	PatternNone             PatternLabel = ""
	PatternTestData         PatternLabel = "testdata-pair"
	PatternTableDrivenTest  PatternLabel = "table-driven-test"
	PatternTestScaffolding  PatternLabel = "test-scaffolding"
	PatternDataDominated    PatternLabel = "data-dominated"
	PatternSignatureOnly    PatternLabel = "signature-only"
	PatternRAIIDefer        PatternLabel = "raii-defer"
	PatternErrorPropagation PatternLabel = "error-propagation"
	PatternErrorWrapping    PatternLabel = "error-wrapping"
	PatternAssertionChain   PatternLabel = "assertion-chain"
	PatternCobraBoilerplate PatternLabel = "cobra-boilerplate"
	PatternInterfaceImpl    PatternLabel = "interface-implementation"
	PatternDescribeTable    PatternLabel = "describe-table"
	PatternBuilderCallback  PatternLabel = "builder-callback"
	PatternAssignErrorCheck PatternLabel = "assign-error-check"
	PatternSingleCallExpr   PatternLabel = "single-call-expression"
)

// EvaluateActionabilityWithLabel returns both the actionability and the
// pattern label that caused it. This allows downstream code to adjust
// category/suggestion based on which specific pattern was detected.
func EvaluateActionabilityWithLabel(nodeSeqs [][]*domain.CloneNode) (PatternLabel, domain.CloneActionability) {
	return evaluateActionabilityDetailed(nodeSeqs)
}

func evaluateActionabilityDetailed(nodeSeqs [][]*domain.CloneNode) (PatternLabel, domain.CloneActionability) {
	if len(nodeSeqs) == 0 {
		return PatternNone, domain.Actionable
	}

	if isSignatureOnlyMatch(nodeSeqs) {
		return PatternSignatureOnly, domain.NonActionable
	}

	if isInterfaceImplementation(nodeSeqs) {
		return PatternInterfaceImpl, domain.NonActionable
	}

	if isPureDeferPattern(nodeSeqs) {
		return PatternRAIIDefer, domain.NonActionable
	}

	if isPureErrorPropagation(nodeSeqs) {
		return PatternErrorPropagation, domain.NonActionable
	}

	if isAssignWithErrorCheck(nodeSeqs) {
		return PatternAssignErrorCheck, domain.NonActionable
	}

	if isSingleCallExpression(nodeSeqs) {
		return PatternSingleCallExpr, domain.NonActionable
	}

	if isErrorWrappingReturn(nodeSeqs) {
		return PatternErrorWrapping, domain.NonActionable
	}

	if isAssertionChain(nodeSeqs) {
		return PatternAssertionChain, domain.NonActionable
	}

	if isCobraCommandBoilerplate(nodeSeqs) {
		return PatternCobraBoilerplate, domain.NonActionable
	}

	if isTestDataFilePair(nodeSeqs) {
		return PatternTestData, domain.NonActionable
	}

	if isTableDrivenTestBody(nodeSeqs) {
		return PatternTableDrivenTest, domain.NonActionable
	}

	if isTestScaffolding(nodeSeqs) {
		return PatternTestScaffolding, domain.NonActionable
	}

	if isDataDominated(nodeSeqs) {
		return PatternDataDominated, domain.NonActionable
	}

	if isDescribeTablePattern(nodeSeqs) {
		return PatternDescribeTable, domain.NonActionable
	}

	if isBuilderCallbackPattern(nodeSeqs) {
		return PatternBuilderCallback, domain.NonActionable
	}

	return PatternNone, domain.Actionable
}

// isSignatureOnlyMatch reports whether every clone is a single FuncDecl node
// WITHOUT a substantial body (>3 child nodes). Interface method signatures,
// empty stubs, and forwarding methods fit this pattern.
//
// A FuncDecl with a real body (BlockStmt with >0 children) is actionable —
// identical bodies can be extracted to shared functions. But if two FuncDecls
// match and the body is empty (or has only boilerplate), deduplication is
// impossible without changing the interface.
func isSignatureOnlyMatch(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.FuncDecl {
			return false
		}

		return !hasRealBody(root)
	})
}

// hasRealBody checks if a FuncDecl contains a body with meaningful logic
// beyond the signature itself.
func hasRealBody(node *domain.CloneNode) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.BlockStmt && len(child.Children) > 0 {
			return true
		}
	}

	return false
}

// isInterfaceImplementation detects when all clone fragments are implementations
// of the same interface method. The signal is: multiple fragments from different
// files that all start with a FuncType node, indicating identical method signatures.
//
// In Go, when 3+ files contain a method with the same FuncType signature and
// matching body prefix, they almost certainly satisfy a common interface. The
// suffix tree already guarantees the token sequences match, so we only need to
// verify the structural pattern and file diversity.
//
// This works because:
//   - The suffix tree matches on Type values, so identical FuncType subtrees
//     (same parameter types, same return types) produce the same token sequence.
//   - 3+ different files with matching FuncType + body prefix is almost never
//     coincidence — it means they all satisfy the same interface contract.
//   - Semantic mode makes this even more precise by encoding identifier/operator
//     names into the Type field.
func isInterfaceImplementation(nodeSeqs [][]*domain.CloneNode) bool {
	if len(nodeSeqs) < 3 {
		return false
	}

	files := make(map[string]bool)

	for _, seq := range nodeSeqs {
		if len(seq) == 0 {
			return false
		}

		root := seq[0]
		// The fragment must be rooted at a FuncType — this is the signature
		// node shared across interface implementations.
		if root.BaseType != golang.FuncType {
			return false
		}

		if root.Filename != "" {
			files[root.Filename] = true
		}
	}

	return len(files) >= 3
}
