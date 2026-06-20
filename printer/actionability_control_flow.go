package printer

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// RAII cleanup method names that indicate non-actionable patterns.
const cleanupMethodName = "Unlock"

// isPureDeferPattern reports whether every clone is a DeferStmt
// wrapping a RAII-style call (Unlock, Close, etc.).
//
// Using the Name field on child nodes, we can now distinguish
// `defer mu.Unlock()` from `defer processOrder()` — the former is
// idiomatic RAII cleanup (non-actionable), the latter is real duplication.
func isPureDeferPattern(nodeSeqs [][]*syntax.Node) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*syntax.Node) bool {
		if len(seq) != 1 {
			return false
		}

		if baseTypeOf(seq[0]) != golang.DeferStmt {
			return false
		}

		return isRAIIDeferCall(seq[0])
	})
}

// isRAIIDeferCall checks if a DeferStmt wraps a known RAII cleanup method.
func isRAIIDeferCall(node *syntax.Node) bool {
	for _, child := range node.Children {
		if baseTypeOf(child) == golang.CallExpr {
			for _, arg := range child.Children {
				if baseTypeOf(arg) == golang.SelectorExpr && isCleanupMethod(arg.Name) {
					return true
				}
			}
		}
	}

	return false
}

// isCleanupMethod reports whether a method name is a known RAII cleanup.
func isCleanupMethod(name string) bool {
	switch name {
	case cleanupMethodName,
		"Close", "Done", "Cancel", "Release", "Finish", "Disconnect", "Free",
		"Stop", "Shutdown", "Cleanup", "Reset", "Put", "Drop", "Abort", "Teardown":
		return true
	default:
		return false
	}
}

// isPureErrorPropagation reports whether every clone is an IfStmt
// that only contains error propagation: if err != nil { return err }.
func isPureErrorPropagation(nodeSeqs [][]*syntax.Node) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*syntax.Node) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if baseTypeOf(root) != golang.IfStmt {
			return false
		}

		return isErrorOnlyIf(root)
	})
}

// isErrorOnlyIf checks if an IfStmt is a pure error propagation pattern.
// It must have:
//   - A condition containing a comparison against nil (BinaryExpr with nil)
//   - A body containing only a ReturnStmt (or CallExpr wrapping error)
//   - No Else branch
func isErrorOnlyIf(node *syntax.Node) bool {
	var (
		hasNilCompare bool
		hasReturnErr  bool
		hasElse       bool
	)

	for _, child := range node.Children {
		switch baseTypeOf(child) {
		case golang.BinaryExpr:
			if containsNilIdentifier(child) {
				hasNilCompare = true
			}
		case golang.BlockStmt:
			if isReturnOrWrappedReturn(child) {
				hasReturnErr = true
			}
		case golang.IfStmt:
			return false
		default:
			if baseTypeOf(child) != golang.AssignStmt &&
				baseTypeOf(child) != golang.DeclStmt {
				hasElse = true
			}
		}
	}

	return hasNilCompare && hasReturnErr && !hasElse
}

// containsNilIdentifier checks if a BinaryExpr compares against nil.
// Uses the Name field to verify an identifier named "nil" is present,
// rather than just checking for any Ident node.
func containsNilIdentifier(node *syntax.Node) bool {
	for _, child := range node.Children {
		if baseTypeOf(child) == golang.Ident && child.Name == "nil" {
			return true
		}
	}

	return false
}

// isReturnOrWrappedReturn checks if a BlockStmt only contains a ReturnStmt.
func isReturnOrWrappedReturn(node *syntax.Node) bool {
	if len(node.Children) == 0 {
		return false
	}

	// Allow single return or return with CallExpr (e.g., return fmt.Errorf("...")).
	if len(node.Children) == 1 {
		bt := baseTypeOf(node.Children[0])

		return bt == golang.ReturnStmt || bt == golang.CallExpr
	}

	return false
}
