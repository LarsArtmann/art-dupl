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
//
// CHILD-ORDER CONTRACT: AssignStmt nodes serialize their children as
// [rhs..., lhs...] (syntax/golang/transform.go), so every assignment matcher
// below identifies sides by node TYPE, never by position. The 2026-10-11
// first version read Children[0] as the target, which silently disabled the
// whole append path on real serialized clones (hand-built test trees shared
// the wrong order, so the unit tests passed anyway) — caught while
// investigating the surviving go-paperless deep-copy family.

// collectorCallAllowList holds the only call targets permitted inside a
// collector tail: builtin collectors/plumbing and common scalar conversions.
// Any other call (strings.Join, processItem, fmt.Fprintf, ...) is real logic
// and disqualifies the pattern.
var collectorCallAllowList = map[string]bool{ //nolint:gochecknoglobals // static set
	"append":  true,
	"make":    true,
	"len":     true,
	"cap":     true,
	"min":     true,
	"max":     true,
	"string":  true,
	"int":     true,
	"int32":   true,
	"int64":   true,
	"uint":    true,
	"bool":    true,
	"float64": true,
}

// clonePackageAllowList holds the package qualifiers whose Clone calls are
// copy plumbing (`slices.Clone`, `maps.Clone`) in deep-copy collector loops.
var clonePackageAllowList = map[string]bool{ //nolint:gochecknoglobals // static set
	"slices": true,
	"maps":   true,
}

// isAppendCollectorTail reports whether every clone consists solely of
// collector statements: append/filter/deep-copy loop tails that accumulate
// elements into a slice (`out = append(out, x)` in a loop, optionally guarded
// by a trivial filter `if`, optionally preceded by a make preallocation and
// followed by a bare `return result`). Deep-copy loops (`copied := *src;
// copied.Field = slices.Clone(src.Field); out = append(out, copied)`) are
// covered: the per-type field lists differ across clones, so generics cannot
// unify them. The Go idiom for all of these is a stdlib function or staying
// as-is; the shape is too ubiquitous to be actionable duplication.
func isAppendCollectorTail(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, seqIsCollectorTail)
}

func seqIsCollectorTail(seq []*domain.CloneNode) bool {
	if len(seq) == 0 || len(seq) > 3 {
		return false
	}

	sawAccumulation := false

	for _, root := range seq {
		accumulates, ok := collectorStatementAccumulates(root)
		if !ok {
			return false
		}

		sawAccumulation = sawAccumulation || accumulates
	}

	// Without an append/make accumulation somewhere this is just a guard
	// or control-flow fragment, not a collector tail.
	return sawAccumulation
}

// collectorStatementAccumulates classifies one statement root of a
// collector tail and reports whether it performs an accumulation
// (append/make assignment) at itself or any nesting depth.
func collectorStatementAccumulates(n *domain.CloneNode) (bool, bool) {
	switch n.BaseType {
	case golang.RangeStmt, golang.ForStmt:
		return loopAccumulates(n)

	case golang.IfStmt:
		return ifAccumulates(n)

	case golang.AssignStmt:
		return collectorAssign(n)

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

// loopAccumulates validates a collector loop: exactly one body block, a
// header free of real logic, and a body consisting solely of collector
// statements.
func loopAccumulates(n *domain.CloneNode) (bool, bool) {
	body := soleBlockChild(n)
	if body == nil {
		return false, false
	}

	if !blockIsCollectorOnly(n) || headerHasRealLogic(n, body) {
		return false, false
	}

	return blockAccumulates(body), true
}

// ifAccumulates validates an in-loop filter guard: collector-only branches
// with a plumbing condition.
func ifAccumulates(n *domain.CloneNode) (bool, bool) {
	if !ifIsCollectorFilter(n) {
		return false, false
	}

	for _, child := range n.Children {
		if child.BaseType == golang.BlockStmt {
			return blockAccumulates(child), true
		}
	}

	return false, true
}

// blockAccumulates reports whether a block subtree contains an accumulation
// statement at any depth.
func blockAccumulates(n *domain.CloneNode) bool {
	if n.BaseType == golang.AssignStmt {
		if accumulates, ok := collectorAssign(n); ok && accumulates {
			return true
		}
	}

	return slices.ContainsFunc(n.Children, blockAccumulates)
}

// collectorAssign classifies an assignment permitted inside a collector
// tail, identifying the call side by node TYPE (never position — see the
// child-order contract above). Permitted forms: append-accumulation,
// make-preallocation, deep-copy plumbing (`copied := src`, `copied := *src`,
// `copied.Field = src.Field`, `copied.Field = slices.Clone(src.Field)`),
// and scalar conversions onto an identifier or field target.
func collectorAssign(n *domain.CloneNode) (bool, bool) {
	if len(n.Children) != 2 {
		return false, false
	}

	rhs, lhs := n.Children[0], n.Children[1]

	if rhs.BaseType == golang.CallExpr && lhs.BaseType != golang.CallExpr {
		return callIsCollectorPlumbing(rhs, lhs)
	}

	// Tolerate call-last order for robustness; Go assignment targets are
	// never call expressions, so the sides stay unambiguous either way.
	if lhs.BaseType == golang.CallExpr && rhs.BaseType != golang.CallExpr {
		return callIsCollectorPlumbing(lhs, rhs)
	}

	// Plain copy: `copied := src`, `copied.Field = src.Field`.
	if lhs.BaseType != golang.Ident && lhs.BaseType != golang.SelectorExpr {
		return false, false
	}

	return false, !copyPlumbingSubtree(rhs)
}

// callIsCollectorPlumbing classifies a call-RHS assignment against its
// target: append/make accumulate, clone/conversion calls are copy plumbing.
func callIsCollectorPlumbing(call, target *domain.CloneNode) (bool, bool) {
	if len(call.Children) == 0 {
		return false, false
	}

	fun := call.Children[0]

	if fun.BaseType == golang.Ident {
		switch fun.Name {
		case "append":
			return appendAccumulates(call, target)
		case "make":
			return true, target.BaseType == golang.Ident && !subtreeHasRealLogic(call)
		default:
			return false, collectorCallAllowList[fun.Name] && !subtreeHasRealLogic(call)
		}
	}

	if isCloneCall(fun) {
		return false, !copyPlumbingSubtree(call)
	}

	return false, false
}

// appendAccumulates matches `x = append(x, ...)` — the first append argument
// is the same identifier as the assignment target.
func appendAccumulates(call, target *domain.CloneNode) (bool, bool) {
	if target.BaseType != golang.Ident || len(call.Children) < 2 {
		return false, false
	}

	firstArg := call.Children[1]
	if firstArg.BaseType != golang.Ident || firstArg.Name != target.Name {
		return false, false
	}

	return true, !slices.ContainsFunc(call.Children[2:], subtreeHasRealLogic)
}

// isCloneCall matches the `pkg.Clone` selector form with pkg in
// {slices, maps}. The SelectorExpr node carries the field name in Name and
// its package identifier as the first child.
func isCloneCall(fun *domain.CloneNode) bool {
	return fun.BaseType == golang.SelectorExpr && fun.Name == "Clone" &&
		len(fun.Children) > 0 && fun.Children[0].BaseType == golang.Ident &&
		clonePackageAllowList[fun.Children[0].Name]
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

	if blocks != 1 || !blockIsCollectorOnly(n) {
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

		if slices.ContainsFunc(child.Children, func(stmt *domain.CloneNode) bool {
			_, ok := collectorStatementAccumulates(stmt)

			return !ok
		}) {
			return false
		}
	}

	return true
}

// headerHasRealLogic checks all children of n except block (the loop header
// or if condition) for real logic.
func headerHasRealLogic(n *domain.CloneNode, block *domain.CloneNode) bool {
	return slices.ContainsFunc(n.Children, func(child *domain.CloneNode) bool {
		return child != block && subtreeHasRealLogic(child)
	})
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
	golang.CompositeLit:   true,
	golang.FuncLit:        true,
	golang.DeferStmt:      true,
	golang.GoStmt:         true,
	golang.SelectStmt:     true,
	golang.SwitchStmt:     true,
	golang.TypeSwitchStmt: true,
	golang.SendStmt:       true,
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

	return slices.ContainsFunc(n.Children, subtreeHasRealLogic)
}

// copyPlumbingSubtree is subtreeHasRealLogic plus the `slices.Clone` /
// `maps.Clone` exception for deep-copy loops.
func copyPlumbingSubtree(n *domain.CloneNode) bool {
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
			if !isCloneCall(fun) {
				return true
			}
		}
	}

	return slices.ContainsFunc(n.Children, copyPlumbingSubtree)
}

// subtreeHasComplexity reports whether the subtree contains any
// logic-bearing node type (closures, composite literals, defers, goroutines,
// selects, switches, channel sends, type asserts) — WITHOUT the call-callee
// restriction of subtreeHasRealLogic. Used where a call of any callee is
// acceptable but its arguments must be plain values (helper results,
// constructor seeding).
func subtreeHasComplexity(n *domain.CloneNode) bool {
	if n == nil {
		return false
	}

	if collectorRealLogicTypes[n.BaseType] {
		return true
	}

	return slices.ContainsFunc(n.Children, subtreeHasComplexity)
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
	if len(seq) == 0 || len(seq) > 3 {
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
		case golang.IfStmt:
			// Constructor lookup guard: `if !ok { return }` before the apply
			// loop (gap family 67761296 in the survivor labeling).
			if !ifIsBareReturnGuard(root) {
				return false
			}
		case golang.AssignStmt:
			// Target seeding: `cfg := defaultConfig()` or `cfg := Config{...}`
			// before the apply loop (gap families 2384b605 / 87e6d65d / 754d60f8).
			if !assignSeedsOptionTarget(root) {
				return false
			}
		default:
			return false
		}
	}

	if loop == nil {
		return false
	}

	return rangeLoopAppliesOption(loop)
}

// rangeLoopAppliesOption validates the option-apply loop body: exactly one
// statement, a call of a range-header identifier on one plumbing argument.
func rangeLoopAppliesOption(loop *domain.CloneNode) bool {
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
	return slices.ContainsFunc(loop.Children, func(child *domain.CloneNode) bool {
		return child.BaseType == golang.Ident && child.Name == name
	})
}

// ifIsBareReturnGuard matches a pre-loop lookup guard: guard-clause shape
// (`if !ok { ... }`, no else) whose body is only bare/identifier returns.
func ifIsBareReturnGuard(root *domain.CloneNode) bool {
	return isGuardClauseBody(root) && blockReturnsOnlyIdents(root)
}

// assignSeedsOptionTarget matches the constructor seeding that precedes an
// option-apply loop: `cfg := defaultConfig()` (call of any callee with
// plain-value arguments) or `cfg := Config{seed: true}` (composite literal
// with plain-value elements), onto a single identifier target. The seeded
// values are idiom plumbing — where they carry closures or nested literals
// the constructor has real defaulting logic and stays actionable.
func assignSeedsOptionTarget(n *domain.CloneNode) bool {
	if len(n.Children) != 2 {
		return false
	}

	value, target := n.Children[0], n.Children[1]
	if target.BaseType == golang.CallExpr && value.BaseType != golang.CallExpr {
		value, target = target, value
	}

	if target.BaseType != golang.Ident {
		return false
	}

	switch value.BaseType {
	case golang.CallExpr:
		return !subtreeHasComplexity(value)
	case golang.CompositeLit:
		return !slices.ContainsFunc(value.Children, subtreeHasComplexity)
	default:
		return false
	}
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
