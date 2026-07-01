package printer

import (
	"slices"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// assertionMethodNames are test assertion framework methods that indicate
// non-actionable test boilerplate when they appear in chains of 3+.
var assertionMethodNames = map[string]bool{
	"Expect":  true,
	"Assert":  true,
	"Require": true,
	"Should":  true,
	"Must":    true,
	"So":      true,
}

// isAssertionChain reports whether every clone is dominated by test assertion
// calls (Expect/Assert/Require/Should/Must/So). These chains are Ginkgo/testify
// boilerplate that cannot be deduplicated without breaking test readability.
//
// A clone is an assertion chain when ≥3 CallExpr nodes target assertion methods.
func isAssertionChain(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, isAssertionDominatedSeq)
}

// isAssertionDominatedSeq checks if a node sequence has ≥3 assertion calls.
func isAssertionDominatedSeq(seq []*domain.CloneNode) bool {
	assertionCount := 0

	for _, node := range seq {
		if node.BaseType == golang.CallExpr {
			if hasAssertionTarget(node) {
				assertionCount++
			}
		}
	}

	return assertionCount >= 3
}

// hasAssertionTarget checks if a CallExpr targets a known assertion method.
// Looks at the Fun child: for `Expect(x).To(Equal(y))`, the outer CallExpr
// has Fun=SelectorExpr with Sel="To", but the inner CallExpr has Fun=Ident "Expect".
func hasAssertionTarget(callExpr *domain.CloneNode) bool {
	for _, child := range callExpr.Children {
		if child.BaseType == golang.SelectorExpr {
			if assertionMethodNames[child.Name] {
				return true
			}
		}

		if child.BaseType == golang.Ident {
			if assertionMethodNames[child.Name] {
				return true
			}
		}
	}

	return false
}

// isCobraCommandBoilerplate reports whether every clone is dominated by
// cobra.Command or fang.Command struct literals with Run/RunE fields.
// These are CLI boilerplate that cannot be deduplicated.
func isCobraCommandBoilerplate(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		return slices.ContainsFunc(seq, isCommandLiteral)
	})
}

// isCommandLiteral checks if a node is part of a cobra.Command or fang.Command
// literal. Looks for CompositeLit containing a SelectorExpr named "Command"
// with a parent Ident of "cobra" or "fang".
func isCommandLiteral(node *domain.CloneNode) bool {
	if node.BaseType != golang.CompositeLit {
		return false
	}

	for _, child := range node.Children {
		if child.BaseType == golang.SelectorExpr && child.Name == "Command" {
			return true
		}
	}

	return false
}

// isErrorWrappingReturn reports whether every clone is an IfStmt with a
// nil check where the body contains only a return with fmt.Errorf or
// errors.Wrap. This is the standard Go error-wrapping idiom.
func isErrorWrappingReturn(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) != 1 {
			return false
		}

		root := seq[0]
		if root.BaseType != golang.IfStmt {
			return false
		}

		return isErrorWrappingBody(root)
	})
}

// isErrorWrappingBody checks if an IfStmt body contains only error wrapping.
// Must have: nil comparison, body with return calling Errorf/Wrap/Errorw.
func isErrorWrappingBody(node *domain.CloneNode) bool {
	var (
		hasNilCompare bool
		hasErrorWrap  bool
	)

	for _, child := range node.Children {
		switch child.BaseType {
		case golang.BinaryExpr:
			if containsNilIdentifier(child) {
				hasNilCompare = true
			}

		case golang.BlockStmt:
			if hasErrorWrappingCall(child) {
				hasErrorWrap = true
			}

		case golang.IfStmt:
			return false
		}
	}

	return hasNilCompare && hasErrorWrap
}

// hasErrorWrappingCall checks if a BlockStmt contains a return with an
// error-wrapping function call (Errorf, Wrap, Errorw, Wrapf).
func hasErrorWrappingCall(block *domain.CloneNode) bool {
	if len(block.Children) == 0 {
		return false
	}

	for _, child := range block.Children {
		if child.BaseType == golang.ReturnStmt {
			if slices.ContainsFunc(child.Children, isWrappingCall) {
				return true
			}
		}
	}

	return false
}

// isWrappingCall checks if a node is a CallExpr to a known error-wrapping function.
func isWrappingCall(node *domain.CloneNode) bool {
	if node.BaseType != golang.CallExpr {
		return false
	}

	wrappingNames := map[string]bool{
		"Errorf": true, "Wrap": true, "Errorw": true, "Wrapf": true, "Wrapr": true,
	}

	for _, child := range node.Children {
		if child.BaseType == golang.SelectorExpr && wrappingNames[child.Name] {
			return true
		}
	}

	return false
}
