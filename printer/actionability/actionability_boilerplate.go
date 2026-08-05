package actionability

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isAssignWithErrorCheck reports whether every clone is a 2-statement sequence
// of an assignment followed by an error-check IfStmt. This is the most common
// Go boilerplate idiom:
//
//	err := someFunc()
//	if err != nil {
//	    return ...
//	}
//
// Every Go program contains dozens of these. They are not actionable
// duplication — extracting them would obscure the control flow without
// reducing complexity.
func isAssignWithErrorCheck(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 2 {
			return false
		}

		if seq[0].BaseType != golang.AssignStmt {
			return false
		}

		if seq[1].BaseType != golang.IfStmt {
			return false
		}

		return isErrorOnlyIf(seq[1])
	})
}

// isAssignWithBoolGuard reports whether every clone is a 2-statement sequence
// of an assignment that produces a bool ok value followed by a guard IfStmt:
//
//	X, ok := helper()
//	if !ok { return }
//
// This is the Go comma-ok idiom boilerplate. The helper call IS the extraction;
// the ok guard is irreducible control flow. Extracting the guard into a helper
// would require passing in the return type, obscuring the intent.
func isAssignWithBoolGuard(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 2 {
			return false
		}

		if seq[0].BaseType != golang.AssignStmt {
			return false
		}

		if seq[1].BaseType != golang.IfStmt {
			return false
		}

		return isBoolGuardIf(seq[0], seq[1])
	})
}

// isBoolGuardIf checks if an IfStmt is a bool-ok guard: condition is !ok
// (UnaryExpr wrapping an Ident assigned in the preceding AssignStmt),
// body is return-only, and no else branch.
func isBoolGuardIf(assign *domain.CloneNode, ifStmt *domain.CloneNode) bool {
	okName := findOkVarName(assign)
	if okName == "" {
		return false
	}

	var (
		hasGuard bool
		hasBody  bool
		hasElse  bool
	)

	for _, child := range ifStmt.Children {
		switch child.BaseType {
		case golang.UnaryExpr:
			if isNotOkExpr(child, okName) {
				hasGuard = true
			}

		case golang.BlockStmt:
			if isReturnOnlyBody(child) {
				hasBody = true
			}

		case golang.IfStmt:
			return false

		case golang.BinaryExpr, golang.Ident, golang.CallExpr,
			golang.SelectorExpr, golang.BasicLit,
			golang.ParenExpr, golang.IndexExpr:
			// Other condition shapes — not a simple !ok guard

		default:
			if child.BaseType == golang.AssignStmt || child.BaseType == golang.DeclStmt {
				continue
			}

			hasElse = true
		}
	}

	return hasGuard && hasBody && !hasElse
}

// isBoolGuardVarName reports whether name is a conventional comma-ok / bool
// result variable in Go's comma-ok idiom and existence-check patterns.
func isBoolGuardVarName(name string) bool {
	switch name {
	case "ok", "found", "exists", "success", "present":
		return true
	default:
		return false
	}
}

// findOkVarName returns the name of the bool-guard variable in an AssignStmt,
// or empty string if none found. Looks for Ident children with a conventional
// bool-guard name (ok, found, exists, success, present).
func findOkVarName(assign *domain.CloneNode) string {
	for _, child := range assign.Children {
		if child.BaseType == golang.Ident && isBoolGuardVarName(child.Name) {
			return child.Name
		}
	}

	return ""
}

// isNotOkExpr checks if a UnaryExpr is !ok (negation of the ok identifier).
func isNotOkExpr(node *domain.CloneNode, okName string) bool {
	for _, child := range node.Children {
		if child.BaseType == golang.Ident && child.Name == okName {
			return true
		}
	}

	return false
}

// isSingleCallExpression reports whether every clone is exactly one CallExpr
// node, either as a bare CallExpr (non-statement match) or as an ExprStmt
// wrapping a single CallExpr (statement-level match). Single function calls
// with different arguments are not actionable duplication — they are just
// using the same API with different data.
// Examples: errors.New("foo"), fmt.Println("bar"), http.Get(url), t.Parallel().
func isSingleCallExpression(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		return isLoneCallExpr(seq[0])
	})
}

// isLoneCallExpr reports whether a single CloneNode is a bare CallExpr or an
// ExprStmt wrapping exactly one CallExpr. In Go's AST, a standalone function
// call as a statement (e.g. t.Parallel()) is wrapped in an ExprStmt. When
// statement-level tokenization is active, the matched node has BaseType
// ExprStmt, not CallExpr, so we must look inside the ExprStmt to find the call.
func isLoneCallExpr(n *domain.CloneNode) bool {
	if n.BaseType == golang.CallExpr {
		return true
	}

	if n.BaseType == golang.ExprStmt &&
		len(n.Children) == 1 &&
		n.Children[0].BaseType == golang.CallExpr {
		return true
	}

	return false
}

// isSingleSimpleStatement reports whether every clone is exactly one terminal
// statement that cannot represent actionable duplication. These are statements
// with no body/block to extract: return, assignment, increment/decrement,
// branch (break/continue), channel send, and local var/const declarations.
//
// A single return, assignment, or var declaration can never be meaningfully
// extracted into a reusable unit — they are language primitives, not domain
// logic. Even a complex expression inside (e.g. return f(a,b,c)) is the
// expression that could be extracted, not the return itself.
//
// Type declarations (DeclStmt wrapping TypeSpec) are NOT filtered here because
// duplicated struct/interface definitions can represent real actionable cloning.
func isSingleSimpleStatement(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		return isTerminalStatement(seq[0])
	})
}

// isTerminalStatement classifies a single CloneNode as a terminal statement
// (no body/block, cannot be extracted as a unit).
func isTerminalStatement(n *domain.CloneNode) bool {
	switch n.BaseType {
	case golang.ReturnStmt, golang.AssignStmt, golang.IncDecStmt,
		golang.BranchStmt, golang.SendStmt:
		return true

	case golang.DeclStmt:
		return !subtreeContainsTypeSpec(n)

	default:
		return false
	}
}

// subtreeContainsTypeSpec reports whether any descendant node in the subtree
// is a TypeSpec. Used to exclude type declarations (which can be actionable
// duplication) from the terminal-statement filter.
func subtreeContainsTypeSpec(n *domain.CloneNode) bool {
	if n.BaseType == golang.TypeSpec {
		return true
	}

	return slices.ContainsFunc(n.Children, subtreeContainsTypeSpec)
}

// isSingleDeclaration reports whether every clone is exactly one package-level
// declaration node (a ValueSpec or a TypeSpec) fingerprinted as a single
// statement token by the statement-level tokenizer. These are individual
// const/var/type declarations — the atomic units of Go's package-level syntax.
//
// A single declaration can never be meaningfully extracted into a reusable
// unit. The common shapes are all idiomatic Go that exists to expose a symbol
// under a local name:
//
//	const Foo = pkg.Foo          // re-export
//	type Mode = domain.Mode      // type alias
//	BadNode = iota               // first value of an iota enum
//
// Type DEFINITIONS (type Foo struct{...}, type Bar interface{...}) are NOT
// suppressed here: a duplicated composite type body can represent real
// actionable cloning. Only references to named types (aliases and the
// type-from-named form type X Y) are suppressed, detected by the absence of
// any composite type body in the subtree (see subtreeHasCompositeType).
func isSingleDeclaration(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		return isAtomicDeclaration(seq[0])
	})
}

// isAtomicDeclaration classifies a single CloneNode as an atomic package-level
// declaration that cannot represent actionable duplication.
func isAtomicDeclaration(n *domain.CloneNode) bool {
	switch n.BaseType {
	case golang.ValueSpec:
		return true

	case golang.TypeSpec:
		return !subtreeHasCompositeType(n)

	default:
		return false
	}
}

// subtreeHasCompositeType reports whether any node in the subtree is a
// composite type DEFINITION (struct, interface, func, array, map, chan).
// Presence of such a node means the TypeSpec carries real structure that may
// represent actionable duplication, so isAtomicDeclaration must NOT suppress it.
func subtreeHasCompositeType(n *domain.CloneNode) bool {
	switch n.BaseType {
	case golang.StructType, golang.InterfaceType, golang.FuncType,
		golang.ArrayType, golang.MapType, golang.ChanType:
		return true
	}

	return slices.ContainsFunc(n.Children, subtreeHasCompositeType)
}

// subtreeHasType recursively checks whether any node in the subtree has the
// given base type. Used by isTypeAliasBlock to detect SelectorExpr references
// to external package types (type X = pkg.Y aliases).
func subtreeHasType(n *domain.CloneNode, typ int32) bool {
	if n.BaseType == typ {
		return true
	}

	return slices.ContainsFunc(n.Children, func(c *domain.CloneNode) bool {
		return subtreeHasType(c, typ)
	})
}

// isInterfaceAssertion reports whether every clone is a single ValueSpec with a
// blank-identifier name. This is the compile-time interface-assertion idiom:
//
//	var _ Interface = (*Type)(nil)
//
// These declarations exist solely to make the compiler verify that *Type
// satisfies Interface. They are duplicated across types by design (each type
// needs its own assertion) and are never actionable duplication.
func isInterfaceAssertion(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		n := seq[0]

		if n.BaseType != golang.ValueSpec {
			return false
		}

		// The transformer adds spec names as Ident children. The blank
		// identifier _ is the first name for var _ I = (*T)(nil).
		if len(n.Children) == 0 {
			return false
		}

		first := n.Children[0]

		return first.BaseType == golang.Ident && first.Name == "_"
	})
}

// isTypeAliasBlock reports whether every clone is a multi-node block where all
// nodes are package-type aliases (type X = pkg.Y). These are re-export shims,
// not actionable duplication:
//
//	type (
//	    ToolCall   = protocoltypes.ToolCall
//	    ToolResult = protocoltypes.ToolResult
//	)
//
// The heuristic: every node is a non-composite TypeSpec that contains a
// SelectorExpr child (indicating a reference to an external package type).
// Single-node aliases are already handled by single-declaration; this pattern
// catches 2+ alias blocks that slip past it.
func isTypeAliasBlock(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) < 2 {
			return false
		}

		for _, n := range seq {
			if !isPackageTypeAlias(n) {
				return false
			}
		}

		return true
	})
}

// isPackageTypeAlias reports whether a CloneNode is a TypeSpec aliasing an
// external package type (type X = pkg.Y). The node must be a non-composite
// TypeSpec containing a SelectorExpr in its subtree.
func isPackageTypeAlias(n *domain.CloneNode) bool {
	if n.BaseType != golang.TypeSpec {
		return false
	}

	// Composite types (struct, interface, etc.) carry real structure.
	if subtreeHasCompositeType(n) {
		return false
	}

	// A SelectorExpr child (pkg.Y) indicates a reference to an external
	// package type, which is the signature of a re-export alias.
	return subtreeHasType(n, golang.SelectorExpr)
}

// isTestHelperDelegate reports whether every clone is a 2-statement test helper
// body: t.Helper() as the first statement, followed by a single delegate call
// to a shared assertion function. This is irreducible Go boilerplate:
//
//	t.Helper()                    // marks the CALLING function, cannot be factored out
//	failIfNilf(t, got, "...", x)  // shared logic already extracted
//
// t.Helper() cannot be moved into the delegate because it marks the caller as
// the helper — moving it would mark the delegate instead. The shared assertion
// logic is already extracted. These are the last remaining duplication: the
// t.Helper() call itself, which is structurally identical across all helpers.
//
// The heuristic is intentionally narrow: exactly 2 statements, first is a
// .Helper() method call, second is any call expression. Longer bodies with
// real logic are NOT matched (they may contain actionable duplication).
func isTestHelperDelegate(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 2 {
			return false
		}

		return isHelperCallStmt(seq[0]) && isLoneCallExpr(seq[1])
	})
}

// isHelperCallStmt reports whether a node is an ExprStmt wrapping a CallExpr
// whose function is a SelectorExpr with method name "Helper". This matches
// t.Helper(), b.Helper(), tb.Helper() — the standard testing.T/TB/B idiom.
func isHelperCallStmt(n *domain.CloneNode) bool {
	if n.BaseType != golang.ExprStmt {
		return false
	}

	call := firstChild(n, golang.CallExpr)
	if call == nil {
		return false
	}

	sel := firstChild(call, golang.SelectorExpr)

	return sel != nil && sel.Name == "Helper"
}

// firstChild returns the first direct child of n with the given BaseType,
// or nil if none matches.
func firstChild(n *domain.CloneNode, typ int32) *domain.CloneNode {
	for _, c := range n.Children {
		if c.BaseType == typ {
			return c
		}
	}

	return nil
}
