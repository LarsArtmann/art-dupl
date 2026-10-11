package actionability

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// The nil-guard-after-helper pattern (plan T13) comes from accept-directive
// mining (docs/research/2026-10-11_accept-directive-mining.md): the
// "canonical IndexIdent+exprName nil-guard prelude" and "nil-params guard"
// accept reasons — consumers repeatedly hand-accept fragments of the shape
//
//	x := helper(args)
//	if x == nil {
//		return false
//	}
//
// The prelude is canonical Go early-return plumbing: the helper differs per
// site, so there is no shared algorithm to extract — only the guard shape.
// A group is NonActionable only when EVERY clone is such a prelude.
//
// FP GUARD (deliberate): guard bodies must be trivial returns only (bare
// return / return idents such as nil or false). A guard that closes
// resources (`return tx.Rollback()`), mutates state, or calls anything is
// REAL logic and must NOT match — those guards are load-bearing.
// Comma-ok guards are separately declined (over-suppression risk).
func isNilGuardAfterHelper(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, seqIsNilGuardAfterHelper)
}

// seqIsNilGuardAfterHelper accepts sequences of at most three statements:
// helper-result assignments (`x := H(y)`), nil guards, and a trailing bare
// return — with at least one nil guard present.
func seqIsNilGuardAfterHelper(seq []*domain.CloneNode) bool {
	if len(seq) == 0 || len(seq) > 3 {
		return false
	}

	sawGuard := false

	for _, root := range seq {
		switch root.BaseType {
		case golang.IfStmt:
			if !ifIsNilGuardWithTrivialBody(root) {
				return false
			}

			sawGuard = true
		case golang.AssignStmt:
			if !assignIsHelperResult(root) {
				return false
			}
		case golang.ReturnStmt:
			if !returnsIdentsOnly(root) {
				return false
			}
		default:
			return false
		}
	}

	return sawGuard
}

// ifIsNilGuardWithTrivialBody matches `if X == nil { <trivial returns> }`
// (either comparison order, == or !=): a guard-clause body whose statements
// are all bare/identifier returns, and whose condition tests an identifier
// against the predeclared nil. `nil` survives alpha-normalization verbatim
// (universe-scope identifiers are never canonicalized), so a Name match is
// reliable on real serialized clones.
func ifIsNilGuardWithTrivialBody(n *domain.CloneNode) bool {
	if !isGuardClauseBody(n) {
		return false
	}

	for _, child := range n.Children {
		if child.BaseType != golang.BinaryExpr {
			continue
		}

		if !binaryExprComparesAgainstNil(child) {
			continue
		}

		return blockReturnsOnlyIdents(n)
	}

	return false
}

// binaryExprComparesAgainstNil reports whether a BinaryExpr has an operand
// that is the predeclared nil identifier. Both operand orders are accepted;
// the operator itself is not inspected (== and != against nil are both nil
// tests in practice).
func binaryExprComparesAgainstNil(n *domain.CloneNode) bool {
	return slices.ContainsFunc(n.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.Ident && child.Name == "nil"
	})
}

// blockReturnsOnlyIdents requires every BlockStmt child of the guard to
// return only identifiers — the trivial-body FP guard. `return tx.Rollback()`
// (call), `return fmt.Errorf(...)` (call), and any assignment fail here and
// keep the group actionable.
func blockReturnsOnlyIdents(n *domain.CloneNode) bool {
	for _, child := range n.Children {
		if child.BaseType != golang.BlockStmt {
			continue
		}

		if !slices.ContainsFunc(child.Children, returnsIdentsOnly) {
			return false
		}
	}

	return true
}

// assignIsHelperResult matches `x := H(y)` / `x := H(y, z)`: exactly one
// assignment target, the value a single call whose arguments carry no
// complexity (no closures, composite literals, defers, goroutines, channel
// or type-assertion operations). The callee may be any identifier or
// selector — the helper is exactly what differs across the clone sites.
func assignIsHelperResult(n *domain.CloneNode) bool {
	sides := callAndValueSides(n)
	if sides == nil {
		return false
	}

	call, target := sides[0], sides[1]

	if target.BaseType != golang.Ident {
		return false
	}

	return !subtreeHasComplexity(call)
}

// callAndValueSides returns the call and non-call children of an assignment
// in serialized order, or nil when the assignment does not have exactly one
// call side and one value side. AssignStmt children serialize as
// [rhs..., lhs...] — sides are identified by TYPE, never position (see the
// child-order contract in actionability_collector.go).
func callAndValueSides(n *domain.CloneNode) []*domain.CloneNode {
	if len(n.Children) != 2 {
		return nil
	}

	first, second := n.Children[0], n.Children[1]

	if first.BaseType == golang.CallExpr && second.BaseType != golang.CallExpr {
		return []*domain.CloneNode{first, second}
	}

	if second.BaseType == golang.CallExpr && first.BaseType != golang.CallExpr {
		return []*domain.CloneNode{second, first}
	}

	return nil
}
