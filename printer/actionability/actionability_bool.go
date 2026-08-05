package actionability

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// isBoolAccumulatorInitializer reports whether every clone is a sequence of two
// or more simple boolean variable initializations such as:
//
//	hasFloatFormat := false
//	hasSeparatorLoop := false
//
// These are common in AST-traversal accumulators, feature-flag parsers, and
// other code that independently tracks a few boolean observations. They are
// not duplicated logic: the variables are local to each detector and carry
// unrelated meaning, so extracting them would couple unrelated code without
// saving meaningful lines. Suppressing them at low thresholds avoids the
// noise that comes from matching the `name := false` shape in isolation.
func isBoolAccumulatorInitializer(nodeSeqs [][]*domain.CloneNode) bool {
	return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
		if len(seq) < 2 {
			return false
		}

		for _, node := range seq {
			if !isBoolVariableInitialization(node) {
				return false
			}
		}

		return true
	})
}

// isBoolVariableInitialization reports whether a single node is a short var
// declaration (or assignment) that assigns a boolean literal to one or more
// identifiers, with no other expression structure. Examples:
//
//	hasX := false
//	hasX, hasY := false, true
//	hasX = true
//
// The check is intentionally structural: it does not try to prove the
// assignment is a *declaration*, because the clone is already matched by the
// suffix tree and the only actionable distinction left is whether the RHS is
// a plain boolean literal. Extracting even a repeated `x = true` pair into a
// helper would be worse than the duplication.
func isBoolVariableInitialization(node *domain.CloneNode) bool {
	if node.BaseType != golang.AssignStmt {
		return false
	}

	if len(node.Children) < 2 {
		return false
	}

	var (
		hasBoolLiteral bool
		hasIdentifier  bool
	)

	for _, child := range node.Children {
		if child.BaseType != golang.Ident {
			return false
		}

		if isBoolLiteralName(child.Name) {
			hasBoolLiteral = true
		} else {
			hasIdentifier = true
		}
	}

	return hasBoolLiteral && hasIdentifier
}

// isBoolLiteralName reports whether a name is one of the Go boolean literals.
func isBoolLiteralName(name string) bool {
	return name == "true" || name == "false"
}
