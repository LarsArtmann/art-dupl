package actionability

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Two data-backed patterns from the accept-directive mining
// (docs/research/2026-10-11_accept-directive-mining.md) and the post-T7
// precision re-measure (docs/research/2026-10-11_suggest-generics-precision-baseline.md
// ledger): append/filter-collector tails (~47% of false generics candidates)
// and bare option-apply loops (~34%) are the two dominant false classes.
// Both are language idioms — the right fix is a stdlib function
// (slices.DeleteFunc, maps.Keys, slices.Collect) or nothing at all, never a
// user-written generic — so suppressing them removes noise from both clone
// output and --suggest-generics candidacy.

// collectorCallAllowList holds the only call targets permitted inside a
// collector tail: builtin collectors/plumbing and common scalar conversions.
// Any other call (strings.Join, processItem, fmt.Fprintf, ...) is real logic
// and disqualifies the pattern.
var collectorCallAllowList = map[string]bool{ //nolint:gochecknoglobals // static set
	"append": true,
	"make":   true,
	"len":    true,
	"cap":    true,
	"min":    true,
	"max":    true,
	"string": true,
	"int":    true,
	"int32":  true,
	"int64":  true,
	"uint":   true,
	"bool":   true,
	"float64": true,
}

// isAppendCollectorTail reports whether every clone consists solely of
// collector statements: append/filter loop tails that accumulate elements
// into a slice (`out = append(out, x)` in a loop, optionally guarded by a
// trivial filter `if`, optionally preceded by a make preallocation and
// followed by a bare `return result`). The Go idiom for all of these is a
// stdlib function or staying as-is; extracting a generic collector saves
// nothing and the shape is too ubiquitous to be actionable duplication.
func isAppendCollectorTail(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) == 0 || len(seq) > 3 {
			return false
		}

		sawAccumulation := false

		for _, root := range seq {
			acc, ok := collectorStatementAccumulates(root)
			if !ok {
				return false
			}

			sawAccumulation = sawAccumulation || acc
		}

		// Without an append/make accumulation somewhere this is just a guard
		// or control-flow fragment, not a collector tail.
		return sawAccumulation
	})
}

// collectorStatementAccumulates classifies one statement root of a
// collector tail and reports whether it performs an accumulation
// (append/make assignment) at itself or any nesting depth.
func collectorStatementAccumulates(n *domain.CloneNode) (accumulates bool, ok bool) {
	switch n.BaseType {
	case golang.RangeStmt, golang.ForStmt:
		body := soleBlockChild(n)
		if body == nil {
			return false, false
		}

		if !blockIsCollectorOnly(n) {
			return false, false
		}

		if headerHasRealLogic(n, body) {
			return false, false
		}

		return blockAccumulates(body), true

	case golang.IfStmt:
		if !ifIsCollectorFilter(n) {
			return false, false
		}

		for _, child := range n.Children {
			if child.BaseType == golang.BlockStmt {
				return blockAccumulates(child), true
			}
		}

		return false, true

	case golang.AssignStmt:
		if isAppendAssign(n) || isMakePrealloc(n) {
			return true, true
		}

		return false, false

	case golang.DeclStmt:
		return false, !subtreeHasRealLogic(n)

	case golang.ReturnStmt:
		return false, returnsIdentsOnly(n)

	case golang.BranchStmt:
		return false, true

	default:
		return false, false
	}
}

// blockAccumulates reports whether a block subtree contains an accumulation
// statement at any depth.
func blockAccumulates(n *domain.CloneNode) bool {
	if n.BaseType == golang.AssignStmt && (isAppendAssign(n) || isMakePrealloc(n)) {
		return true
	}

	for _, child := range n.Children {
		if blockAccumulates(child) {
			return true
		}
	}

	return false
}

// ifIsCollectorFilter matches `if <plumbing cond> { <collector-only body> }`
// with no else branch — the in-loop filter guard of a collector tail.
func ifIsCollectorFilter(n *domain.CloneNode) bool {
	blocks := 0

	for _, child := range n.Children {
		if child.BaseType == golang.BlockStmt {
			blocks++
		}

		if child.BaseType == golang.IfStmt {
			return false
		}
	}

	if blocks != 1 {
		return false
	}

	if !blockIsCollectorOnly(n) {
		return false
	}

	return !headerHasRealLogic(n, nil)
}

// blockIsCollectorOnly reports whether every BlockStmt child of n contains
// only collector statements.
func blockIsCollectorOnly(n *domain.CloneNode) bool {
	for _, child := range n.Children {
		if child.BaseType != golang.BlockStmt {
			continue
		}

		for _, stmt := range child.Children {
			if _, ok := collectorStatementAccumulates(stmt); !ok {
				return false
			}
		}
	}

	return true
}

// headerHasRealLogic checks all children of n except block (the loop header
// or if condition) for real logic.
func headerHasRealLogic(n *domain.CloneNode, block *domain.CloneNode) bool {
	for _, child := range n.Children {
		if block != nil && child == block {
			continue
		}

		if subtreeHasRealLogic(child) {
			return true
		}
	}

	return false
}

// isAppendAssign matches `x = append(x, ...)` / `x := append(x, ...)`
// (first append argument is the same identifier as the assignment target).
func isAppendAssign(n *domain.CloneNode) bool {
	if len(n.Children) != 2 {
		return false
	}

	target, rhs := n.Children[0], n.Children[1]
	if target.BaseType != golang.Ident || rhs.BaseType != golang.CallExpr {
		return false
	}

	call := rhs
	if len(call.Children) == 0 || call.Children[0].BaseType != golang.Ident ||
		call.Children[0].Name != "append" {
		return false
	}

	if len(call.Children) < 2 {
		return false
	}

	firstArg := call.Children[1]
	if firstArg.BaseType != golang.Ident || firstArg.Name != target.Name {
		return false
	}

	for _, arg := range call.Children[2:] {
		if subtreeHasRealLogic(arg) {
			return false
		}
	}

	return true
}

// isMakePrealloc matches `x := make([]T, 0, len(y))`-style preallocation.
func isMakePrealloc(n *domain.CloneNode) bool {
	if len(n.Children) != 2 {
		return false
	}

	target, rhs := n.Children[0], n.Children[1]
	if target.BaseType != golang.Ident || rhs.BaseType != golang.CallExpr {
		return false
	}

	if len(rhs.Children) == 0 || rhs.Children[0].BaseType != golang.Ident ||
		rhs.Children[0].Name != "make" {
		return false
	}

	return !subtreeHasRealLogic(rhs)
}

// returnsIdentsOnly matches `return` / `return x` / `return x, nil` where
// every returned expression is a bare identifier.
func returnsIdentsOnly(n *domain.CloneNode) bool {
	return !slices.ContainsFunc(n.Children, func(child *domain.CloneNode) bool {
		return child.BaseType != golang.Ident
	})
}

// collectorRealLogicTypes are node types that indicate real logic (not
// element plumbing) inside a collector tail.
var collectorRealLogicTypes = map[int32]bool{ //nolint:gochecknoglobals // static set
	golang.CompositeLit:  true,
	golang.FuncLit:       true,
	golang.DeferStmt:     true,
	golang.GoStmt:        true,
	golang.SelectStmt:    true,
	golang.SwitchStmt:    true,
	golang.TypeSwitchStmt: true,
	golang.SendStmt:      true,
	golang.TypeAssertExpr: true,
}

// subtreeHasRealLogic reports whether the subtree contains anything beyond
// value plumbing: a call outside the allow list or a logic-bearing node.
func subtreeHasRealLogic(n *domain.CloneNode) bool {
	if n == nil {
		return false
	}

	if collectorRealLogicTypes[n.BaseType] {
		return true
	}

	if n.BaseType == golang.CallExpr {
		if len(n.Children) == 0 {
			return true
		}

		fun := n.Children[0]
		if fun.BaseType != golang.Ident || !collectorCallAllowList[fun.Name] {
			return true
		}
	}

	for _, child := range n.Children {
		if subtreeHasRealLogic(child) {
			return true
		}
	}

	return false
}

// isOptionApplyLoop reports whether every clone is the bare functional-
// options idiom: `for _, opt := range opts { opt(x) }` optionally followed
// by a bare `return x`. The loop body must be exactly one call of the range
// variable on one plumbing argument. The idiom is how every options-based
// constructor works; a constructor with additional logic (defaults, wiring)
// is NOT this pattern and stays actionable.
func isOptionApplyLoop(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, seqIsOptionApply)
}

func seqIsOptionApply(seq []*domain.CloneNode) bool {
	if len(seq) == 0 || len(seq) > 2 {
		return false
	}

	var loop *domain.CloneNode

	for _, root := range seq {
		switch root.BaseType {
		case golang.RangeStmt:
			if loop != nil {
				return false
			}

			loop = root
		case golang.ReturnStmt:
			if !returnsIdentsOnly(root) {
				return false
			}
		default:
			return false
		}
	}

	if loop == nil {
		return false
	}

	body := soleBlockChild(loop)
	if body == nil || len(body.Children) != 1 {
		return false
	}

	stmt := body.Children[0]
	if stmt.BaseType != golang.ExprStmt {
		return false
	}

	call := unwrapExprStmt(stmt)
	if call == nil || call.BaseType != golang.CallExpr || len(call.Children) != 2 {
		return false
	}

	fun := call.Children[0]
	if fun.BaseType != golang.Ident || !isRangeHeaderIdent(loop, fun.Name) {
		return false
	}

	return !subtreeHasRealLogic(call.Children[1])
}

// isRangeHeaderIdent reports whether name matches an identifier declared in
// the loop header (the range key/value variables), canonical or verbatim.
func isRangeHeaderIdent(loop *domain.CloneNode, name string) bool {
	for _, child := range loop.Children {
		if child.BaseType != golang.Ident {
			continue
		}

		if child.Name == name {
			return true
		}
	}

	return false
}

// soleBlockChild returns the single BlockStmt child of n, or nil.
func soleBlockChild(n *domain.CloneNode) *domain.CloneNode {
	var block *domain.CloneNode

	for _, child := range n.Children {
		if child.BaseType == golang.BlockStmt {
			if block != nil {
				return nil
			}

			block = child
		}
	}

	return block
}
