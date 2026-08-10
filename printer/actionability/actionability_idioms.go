package actionability

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isDeferCallPattern reports whether every clone is a lone DeferStmt wrapping
// any CallExpr. This catches resource-specific teardowns that raii-defer misses
// because the cleanup method name is not in the RAII name list (e.g.,
// defer unsubscribe(), defer cleanupFoo()). These are never actionable: defer
// semantics make extraction dangerous, and the function name carries the
// resource identity.
func isDeferCallPattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		return len(seq) == 1 &&
			seq[0].BaseType == golang.DeferStmt &&
			hasAnyCallExpr(seq[0])
	})
}

// hasAnyCallExpr checks if a DeferStmt contains any CallExpr at its top level.
func hasAnyCallExpr(node *domain.CloneNode) bool {
	return slices.ContainsFunc(node.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.CallExpr
	})
}

// testFrameworkMethodNames are Go testing.T/B methods that are linter-mandated
// or framework-required and never actionable duplication.
var testFrameworkMethodNames = []string{ //nolint:gochecknoglobals // static name set
	"Parallel", "Helper", "Cleanup", "Setenv", "Skip", "FailNow", "Logf", "TempDir",
}

// isTestFrameworkCallPattern reports whether every clone is a lone call to a
// testing framework method (t.Parallel(), b.Helper(), etc.). These are
// linter-mandated boilerplate that appears once per test function and cannot
// be shared or extracted.
func isTestFrameworkCallPattern(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		unwrapped := unwrapExprStmt(seq[0])

		return isTestFrameworkCall(unwrapped)
	})
}

// isTestFrameworkCall checks if a node is a CallExpr to a testing framework
// method (t.Parallel(), b.Helper(), etc.).
func isTestFrameworkCall(node *domain.CloneNode) bool {
	if node.BaseType != golang.CallExpr {
		return false
	}

	return slices.ContainsFunc(node.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.SelectorExpr &&
			slices.Contains(testFrameworkMethodNames, child.Name)
	})
}

// isStateFlagMutation reports whether every clone is a single AssignStmt
// assigning a literal or identifier to a struct field (x.flag = true).
// These are state-machine mutations where the field name carries the meaning;
// extracting a generic would add indirection for a single assignment.
func isStateFlagMutation(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.AssignStmt {
			return false
		}

		return hasSelectorLHS(root)
	})
}

// hasSelectorLHS checks if an AssignStmt has a SelectorExpr (field access)
// in its children, indicating a struct field mutation.
func hasSelectorLHS(node *domain.CloneNode) bool {
	return slices.ContainsFunc(node.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.SelectorExpr
	})
}

// isEmptyDefault reports whether every clone is an IfStmt implementing the
// "empty string default" idiom: if X == "" { X = Y }. This is a common Go
// idiom (now replaceable by cmp.Or in Go 1.22+) but extracting a helper saves
// zero lines for a 1-statement guard.
func isEmptyDefault(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.IfStmt {
			return false
		}

		return isEmptyStringGuard(root)
	})
}

// isEmptyStringGuard checks if an IfStmt is `if X == "" { X = ... }`.
func isEmptyStringGuard(node *domain.CloneNode) bool {
	var (
		hasEmptyCompare bool
		hasAssignBody   bool
	)

	for _, child := range node.Children {
		if child.BaseType == golang.BinaryExpr && containsEmptyStringLit(child) {
			hasEmptyCompare = true
		}

		if child.BaseType == golang.BlockStmt && blockHasAssign(child) {
			hasAssignBody = true
		}
	}

	return hasEmptyCompare && hasAssignBody
}

// containsEmptyStringLit checks if a subtree contains a BasicLit with an
// empty-string Name. In art-dupl's transform, BasicLit.Name is the literal
// value (e.g., `""`, `"hello"`). An empty-string check looks for Name == `""`.
func containsEmptyStringLit(node *domain.CloneNode) bool {
	if node.BaseType == golang.BasicLit && node.Name == `""` {
		return true
	}

	return slices.ContainsFunc(node.Children, containsEmptyStringLit)
}

// blockHasAssign checks if a BlockStmt contains an AssignStmt.
func blockHasAssign(node *domain.CloneNode) bool {
	return slices.ContainsFunc(node.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.AssignStmt
	})
}
