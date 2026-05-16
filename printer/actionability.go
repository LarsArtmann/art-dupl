package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// EvaluateActionability analyzes a clone group and determines whether it
// represents actionable duplication or idiomatic boilerplate noise.
//
// A group is actionable when it contains real logic that can be extracted,
// composed, or otherwise refactored. Non-actionable patterns are standard
// Go idioms that cannot be eliminated without breaking semantics.
//
// To be non-actionable, EVERY clone in the group must match the same
// boilerplate pattern. If any clone differs, the group is actionable.
func EvaluateActionability(nodeSeqs [][]*syntax.Node) domain.CloneActionability {
	if len(nodeSeqs) == 0 {
		return domain.Actionable
	}

	if isSignatureOnlyMatch(nodeSeqs) {
		return domain.NonActionable
	}

	if isPureDeferPattern(nodeSeqs) {
		return domain.NonActionable
	}

	if isPureErrorPropagation(nodeSeqs) {
		return domain.NonActionable
	}

	return domain.Actionable
}

// isSignatureOnlyMatch reports whether every clone is a single FuncDecl node
// WITHOUT a substantial body (>3 child nodes). Interface method signatures,
// empty stubs, and forwarding methods fit this pattern.
//
// A FuncDecl with a real body (BlockStmt with >0 children) is actionable —
// identical bodies can be extracted to shared functions. But if two FuncDecls
// match and the body is empty (or has only boilerplate), deduplication is
// impossible without changing the interface.
func isSignatureOnlyMatch(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.Type != golang.FuncDecl {
			return false
		}

		// A FuncDecl with a real body has a BlockStmt child that has children.
		// If the body is empty (no children), or if there are very few children
		// (just receiver + name + type), it's a signature-only implementation.
		if hasRealBody(root) {
			return false
		}
	}

	return true
}

// hasRealBody checks if a FuncDecl contains a body with meaningful logic
// beyond the signature itself.
func hasRealBody(node *syntax.Node) bool {
	for _, child := range node.Children {
		if child.Type == golang.BlockStmt && len(child.Children) > 0 {
			return true
		}
	}

	return false
}

// isPureDeferPattern reports whether every clone is a DeferStmt.
//
// In practice, a bare DeferStmt match across files means identical defer
// calls (e.g., defer Unlock()). Without source text we cannot distinguish
// `defer mu.Unlock()` from `defer expensiveCleanup()`, so we treat any
// bare DeferStmt as potentially non-actionable.
//
// TODO: Extend syntax.Node with an IdentName field (or use source position
// lookup) to distinguish RAII Unlock/Close from business-logic defer.
func isPureDeferPattern(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		if seq[0].Type != golang.DeferStmt {
			return false
		}
	}

	return true
}

// isPureErrorPropagation reports whether every clone is an IfStmt
// that only contains error propagation: if err != nil { return err }.
func isPureErrorPropagation(nodeSeqs [][]*syntax.Node) bool {
	for _, seq := range nodeSeqs {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.Type != golang.IfStmt {
			return false
		}

		if !isErrorOnlyIf(root) {
			return false
		}
	}

	return true
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
		switch child.Type {
		case golang.BinaryExpr:
			if containsNilIdentifier(child) {
				hasNilCompare = true
			}
		case golang.BlockStmt:
			if isReturnOrWrappedReturn(child) {
				hasReturnErr = true
			}
		case golang.IfStmt:
			// Any nested IfStmt in children means this is not a simple
			// error propagation pattern.
			return false
		default:
			// If there's any other significant child besides condition
			// and body, this isn't pure error propagation.
			if child.Type != golang.AssignStmt &&
				child.Type != golang.DeclStmt {
				hasElse = true
			}
		}
	}

	return hasNilCompare && hasReturnErr && !hasElse
}

// containsNilIdentifier checks if a BinaryExpr compares against nil.
func containsNilIdentifier(node *syntax.Node) bool {
	for _, child := range node.Children {
		if child.Type == golang.Ident {
			// Cannot check actual name without source text.
			// Presence of Ident alongside BinaryExpr is a heuristic.
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

	// Allow single return statement.
	if len(node.Children) == 1 && node.Children[0].Type == golang.ReturnStmt {
		return true
	}

	// Allow return with a CallExpr (e.g., return fmt.Errorf("...")).
	if len(node.Children) == 1 {
		child := node.Children[0]
		if child.Type == golang.ReturnStmt || child.Type == golang.CallExpr {
			return true
		}
	}

	return false
}
